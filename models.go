package main

import (
	"database/sql"
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
			id           string
			seqNumber    int
			addedBy      string
			addedDate    time.Time
			customerName string
		)
		rows.Scan(&id, &seqNumber, &addedBy, &addedDate, &customerName)
		salesOrder, err := NewSalesOrder(id, seqNumber, addedBy, addedDate, customerName)
		if err != nil {
			return nil, err
		}

		if includeItems == true {
			rows2, err := s.db.Query("exec Mobile_GetSalesOrderItems @SalesOrderID = ?;", salesOrder.ID)
			if err != nil {
				return nil, err
			}
			defer rows2.Close()

			for rows2.Next() {
				var (
					soId               string
					salesoid           string
					productName        string
					pricePerKg         decimal.Decimal
					discountPercentage float64
					unitName           string
					qtyInKg            int
					deliveryDeadline   time.Time
				)
				rows2.Scan(&soId, &salesoid, &productName, &pricePerKg, &discountPercentage, &unitName, &qtyInKg, &deliveryDeadline)
				salesOrderItem, err := NewSalesOrderItem(soId, productName, pricePerKg, discountPercentage, unitName, qtyInKg, deliveryDeadline)
				if err != nil {
					return nil, err
				}
				salesOrder.AddItem(salesOrderItem)
			}
		}

		salesOrders = append(salesOrders, salesOrder)
	}

	return salesOrders, nil
}

func (s *SalesOrderRepository) ApproveSalesOrder(salesOrderId string, generateDeliveryRequest bool, userId string) (string, string, error) {
	var (
		result      string
		description string
	)
	row := s.db.QueryRow("exec Mobile_ApproveSalesOrder @SalesOrderID = ?, @GenerateDeliveryRequest = ?, @UserID = ?;", salesOrderId, generateDeliveryRequest, userId)
	err := row.Scan(&result, &description)
	if err != nil {
		return "ERROR", err.Error(), err
	}
	if description == "" && err != nil && result == "OK" {
		description = "Sales order was successfully approved"
	}
	return result, description, err
}

func (s *SalesOrderRepository) RejectSalesOrder(salesOrderId string, reason string, userId string) (string, string, error) {
	var (
		result      string
		description string
	)
	row := s.db.QueryRow("exec Mobile_RejectSalesOrder @SalesOrderID = ?, @Reason = ?, @UserID = ?;", salesOrderId, reason, userId)
	err := row.Scan(&result, &description)
	if err != nil {
		return "ERROR", err.Error(), err
	}
	if description == "" && err != nil && result == "OK" {
		description = "Sales order was rejected"
	}
	return result, description, err
}

type SalesOrder struct {
	ID           string            `json:"id"`
	SeqNumber    int               `json:"seqNumber"`
	AddedBy      string            `json:"addedBy"`
	AddedDate    time.Time         `json:"addedDate"`
	CustomerName string            `json:"customerName"`
	Items        []*SalesOrderItem `json:"items"`
	GrandTotal   decimal.Decimal   `json:"grandTotal"`
}

func NewSalesOrder(id string, seq int, addedBy string, addedDate time.Time, customerName string) (*SalesOrder, error) {
	// leid, err := toLittleEndian(id)
	// if err != nil {
	// 	return nil, err
	// }
	return &SalesOrder{
		ID:           id,
		SeqNumber:    seq,
		AddedBy:      addedBy,
		AddedDate:    addedDate,
		CustomerName: customerName,
		Items:        make([]*SalesOrderItem, 0),
	}, nil
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
	ID               string          `json:"id"`
	ProductName      string          `json:"productName"`
	PricePerKG       decimal.Decimal `json:"pricePerKG"`
	DiscountPercent  float64         `json:"discountPercent"`
	UnitName         string          `json:"unitName"`
	QtyInKG          int             `json:"qtyInKG"`
	DeliveryDeadline time.Time       `json:"deliveryDeadline"`
	LineTotal        decimal.Decimal `json:"lineTotal"`
}

func NewSalesOrderItem(id string, productName string, price decimal.Decimal, discountPercent float64, unitName string, qty int, deliveryDeadline time.Time) (*SalesOrderItem, error) {
	return &SalesOrderItem{
		ID:               id,
		ProductName:      productName,
		PricePerKG:       price,
		DiscountPercent:  discountPercent,
		UnitName:         unitName,
		QtyInKG:          qty,
		DeliveryDeadline: deliveryDeadline,
	}, nil
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
	UserId     string    `json:"userId"`
	UserName   string    `json:"username"`
	Password   string    `json:"-"`
	PartnerID  string    `json:"partnerId"`
	LoggedInAt time.Time `json:"loggedInAt"`
	Email      string    `json:"email"`
}

func NewUser(id string, username string, password string, partnerID string, email string) (*User, error) {
	return &User{
		UserId:     id,
		UserName:   username,
		Password:   password,
		PartnerID:  partnerID,
		LoggedInAt: time.Now(),
		Email:      email,
	}, nil
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

func (u *UserRepository) GetUser(userId string) (*User, error) {
	var (
		UserId    string
		UserName  string
		Password  string
		PartnerID string
		Email     string
	)
	row := u.db.QueryRow(`
		SELECT a.UserId, a.UserName, b.Password, d.ID AS PartnerID, b.LoweredEmail as Email FROM aspnet_Users AS a
		INNER JOIN aspnet_Membership AS b ON a.UserId = b.UserId
		INNER JOIN PartnerUsers AS c ON c.UserID = a.UserId
		INNER JOIN Partners AS d ON d.ID = c.PartnerID
		WHERE a.UserId = ?`, userId)
	err := row.Scan(&UserId, &UserName, &Password, &PartnerID, &Email)
	if err != nil {
		log.Printf("Error in GetUser: %s", err)
		return nil, err
	}
	if UserName == "" {
		return nil, nil
	}
	newUser, err := NewUser(UserId, UserName, Password, PartnerID, Email)
	if err != nil {
		return nil, err
	}
	return newUser, nil
}

func (u *UserRepository) Login(name string, pass string) (*User, error) {
	var (
		UserId    string
		UserName  string
		Password  string
		PartnerID string
		Email     string
	)
	row := u.db.QueryRow(`
		SELECT a.UserId, a.UserName, b.Password, d.ID AS PartnerID, b.LoweredEmail as Email FROM aspnet_Users AS a
		INNER JOIN aspnet_Membership AS b ON a.UserId = b.UserId
		INNER JOIN PartnerUsers AS c ON c.UserID = a.UserId
		INNER JOIN Partners AS d ON d.ID = c.PartnerID
		WHERE a.UserName = ? AND b.Password = ?`, name, pass)
	err := row.Scan(&UserId, &UserName, &Password, &PartnerID, &Email)
	if err != nil {
		return nil, err
	}
	if UserName == "" {
		return nil, nil
	}
	newUser, err := NewUser(UserId, UserName, Password, PartnerID, Email)
	if err != nil {
		return nil, err
	}
	return newUser, nil
}

func (u *UserRepository) UserHasAdminPrivileges(userId string) (bool, error) {
	var ()
	rows, err := u.db.Query(`
		SELECT c.RoleName FROM aspnet_Users AS a
		INNER JOIN aspnet_UsersInRoles AS b ON a.UserId = b.UserId
		INNER JOIN aspnet_Roles AS c ON c.RoleId = b.RoleId
		WHERE a.UserId = ? AND c.RoleName = ? OR c.RoleName = ? OR c.RoleName = ?`, userId, "Administrators", "SalesManager", "MarketingManager")
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
