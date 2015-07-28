package main

import (
	"database/sql"
	"github.com/pborman/uuid"
	"github.com/shopspring/decimal"
	"log"
	"time"
)

type Repository interface {
}

type SalesOrderRepository struct {
	Repository
	db *sql.DB
}

func NewSalesOrderRepository(db *sql.DB) *SalesOrderRepository {
	return &SalesOrderRepository{
		db: db,
	}
}

func (s *SalesOrderRepository) GetPendingSalesOrders(partnerId string, includeItems bool) ([]*SalesOrder, error) {
	salesOrders := make([]*SalesOrder, 0)
	rows, err := s.db.Query("exec Mobile_GetPendingSalesOrders @PartnerID = ?;", partnerId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id           uuid.UUID
			seqNumber    int
			addedBy      string
			addedDate    time.Time
			customerName string
		)
		rows.Scan(&id, &seqNumber, &addedBy, &addedDate, &customerName)
		salesOrder := NewSalesOrder(id, seqNumber, addedBy, addedDate, customerName)

		if includeItems == true {
			rows2, err := s.db.Query("exec Mobile_GetSalesOrderItems @SalesOrderID = ?;", salesOrder.ID.String())
			if err != nil {
				return nil, err
			}
			defer rows2.Close()

			for rows2.Next() {
				var (
					soId               uuid.UUID
					salesoid           uuid.UUID
					productName        string
					pricePerKg         decimal.Decimal
					discountPercentage float64
					unitName           string
					qtyInKg            int
					deliveryDeadline   time.Time
				)
				rows2.Scan(&soId, &salesoid, &productName, &pricePerKg, &discountPercentage, &unitName, &qtyInKg, &deliveryDeadline)
				salesOrderItem := NewSalesOrderItem(soId, productName, pricePerKg, discountPercentage, unitName, qtyInKg, deliveryDeadline)
				salesOrder.AddItem(salesOrderItem)
			}
		}

		salesOrders = append(salesOrders, salesOrder)
	}

	return salesOrders, nil
}

func (s *SalesOrderRepository) ApproveSalesOrder(salesOrderId string, generateDeliveryRequest bool, userId string) error {
	stmt, err := s.db.Prepare("exec Mobile_ApproveSalesOrder @SalesOrderID = ?, @GenerateDeliveryRequest = ?, @UserID = ?;")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(salesOrderId, generateDeliveryRequest, userId)
	return err
}

type SalesOrder struct {
	ID           uuid.UUID         `json:"id"`
	SeqNumber    int               `json:"seqNumber"`
	AddedBy      string            `json:"addedBy"`
	AddedDate    time.Time         `json:"addedDate"`
	CustomerName string            `json:"customerName"`
	Items        []*SalesOrderItem `json:"items"`
	GrandTotal   decimal.Decimal   `json:"grandTotal"`
}

func NewSalesOrder(id uuid.UUID, seq int, addedBy string, addedDate time.Time, customerName string) *SalesOrder {
	leid := toLittleEndian(id)
	return &SalesOrder{
		ID:           leid,
		SeqNumber:    seq,
		AddedBy:      addedBy,
		AddedDate:    addedDate,
		CustomerName: customerName,
		Items:        make([]*SalesOrderItem, 0),
	}
}

func (s *SalesOrder) AddItem(item *SalesOrderItem) {
	s.Items = append(s.Items, item)
	s.CalculateGrandTotal()
}

func (s *SalesOrder) CalculateGrandTotal() decimal.Decimal {
	total := decimal.Zero
	for _, item := range s.Items {
		item.CalculateLineTotal()
		total = total.Add(item.LineTotal)
	}
	s.GrandTotal = total
	return s.GrandTotal
}

type SalesOrderItem struct {
	ID               uuid.UUID       `json:"id"`
	ProductName      string          `json:"productName"`
	PricePerKG       decimal.Decimal `json:"pricePerKG"`
	DiscountPercent  float64         `json:"discountPercent"`
	UnitName         string          `json:"unitName"`
	QtyInKG          int             `json:"qtyInKG"`
	DeliveryDeadline time.Time       `json:"deliveryDeadline"`
	LineTotal        decimal.Decimal `json:"lineTotal"`
}

func NewSalesOrderItem(id uuid.UUID, productName string, price decimal.Decimal, discountPercent float64, unitName string, qty int, deliveryDeadline time.Time) *SalesOrderItem {
	leid := toLittleEndian(id)
	return &SalesOrderItem{
		ID:               leid,
		ProductName:      productName,
		PricePerKG:       price,
		DiscountPercent:  discountPercent,
		UnitName:         unitName,
		QtyInKG:          qty,
		DeliveryDeadline: deliveryDeadline,
	}
}

func (s *SalesOrderItem) CalculateLineTotal() decimal.Decimal {
	qty := decimal.NewFromFloat(float64(s.QtyInKG))
	total := s.PricePerKG.Mul(qty)
	discount := decimal.NewFromFloat(s.DiscountPercent)
	percentageValue := total.Mul(discount).Div(decimal.NewFromFloat(100))

	if total.IntPart() > 0 && discount.IntPart() > 0 && total.Sub(percentageValue).IntPart() > 0 {
		total.Sub(percentageValue)
	}
	s.LineTotal = total
	return total
}

type User struct {
	UserId     uuid.UUID `json:"userId"`
	UserName   string    `json:"username"`
	Password   string    `json:"-"`
	PartnerID  uuid.UUID `json:"partnerId"`
	LoggedInAt time.Time `json:"loggedInAt"`
}

func NewUser(id uuid.UUID, username string, password string, partnerID uuid.UUID) *User {
	leid := toLittleEndian(id)
	lepartnerID := toLittleEndian(partnerID)
	return &User{
		UserId:     leid,
		UserName:   username,
		Password:   password,
		PartnerID:  lepartnerID,
		LoggedInAt: time.Now(),
	}
}

type UserRepository struct {
	Repository
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (u *UserRepository) GetUser(userId uuid.UUID) (*User, error) {
	var (
		UserId    uuid.UUID
		UserName  string
		Password  string
		PartnerID uuid.UUID
	)
	row := u.db.QueryRow(`SELECT a.UserId, a.UserName, b.Password, d.ID AS PartnerID FROM aspnet_Users AS a
		INNER JOIN aspnet_Membership AS b ON a.UserId = b.UserId
		INNER JOIN PartnerUsers AS c ON c.UserID = a.UserId
		INNER JOIN Partners AS d ON d.ID = c.PartnerID
		WHERE a.UserId = ?`, userId.String())
	err := row.Scan(&UserId, &UserName, &Password, &PartnerID)
	if err != nil {
		log.Printf("Error in GetUser: %s", err)
		return nil, err
	}
	if UserName == "" {
		return nil, nil
	}
	return NewUser(UserId, UserName, Password, PartnerID), nil
}

func (u *UserRepository) Login(name string, pass string) (*User, error) {
	var (
		UserId    uuid.UUID
		UserName  string
		Password  string
		PartnerID uuid.UUID
	)
	row := u.db.QueryRow(`SELECT a.UserId, a.UserName, b.Password, d.ID AS PartnerID FROM aspnet_Users AS a
		INNER JOIN aspnet_Membership AS b ON a.UserId = b.UserId
		INNER JOIN PartnerUsers AS c ON c.UserID = a.UserId
		INNER JOIN Partners AS d ON d.ID = c.PartnerID
		WHERE a.UserName = ? AND b.Password = ?`, name, pass)
	err := row.Scan(&UserId, &UserName, &Password, &PartnerID)
	if err != nil {
		return nil, err
	}
	if UserName == "" {
		return nil, nil
	}
	return NewUser(UserId, UserName, Password, PartnerID), nil
}

func (u *UserRepository) UserHasAdminPrivileges(userId uuid.UUID) (bool, error) {
	var ()
	rows, err := u.db.Query(`SELECT c.RoleName FROM aspnet_Users AS a
		INNER JOIN aspnet_UsersInRoles AS b ON a.UserId = b.UserId
		INNER JOIN aspnet_Roles AS c ON c.RoleId = b.RoleId
		WHERE a.UserId = ? AND c.RoleName = ? OR c.RoleName = ? OR c.RoleName = ?`, userId.String(), "Administrators", "SalesManager", "MarketingManager")
	if err != nil {
		return false, err
	}
	defer rows.Close()

	result := false
	for rows.Next() {
		result = true
		break
	}

	return result, nil
}
