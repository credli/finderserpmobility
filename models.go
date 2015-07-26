package main

import (
	"database/sql"
	"github.com/pborman/uuid"
	"github.com/shopspring/decimal"
	"log"
	"time"
)

type Repository interface {
	GetDb() *sql.DB
}

type Database struct {
	db               *sql.DB
	driverName       string
	connectionString string
}

func NewDatabase(driverName string, connectionString string) *Database {
	if driverName == "" || connectionString == "" {
		log.Panicln("Both driver name and connection string are required")
	}
	db, err := sql.Open(driverName, connectionString)
	if err != nil {
		log.Panicf("%s\n", err.Error())
	}
	return &Database{
		db: db,
	}
}

func (d *Database) Open() (*sql.DB, error) {
	return sql.Open(d.driverName, d.connectionString)
}

func (d *Database) GetDriverName() string {
	return d.driverName
}

func (d *Database) GetConnectionString() string {
	return d.connectionString
}

func (d *Database) GetDatabase() *sql.DB {
	return d.db
}

func (d *Database) Close() {
	d.db.Close()
}

type SalesOrderRepository struct {
	Repository
	db *Database
}

func NewSalesOrderRepository(db *Database) *SalesOrderRepository {
	return &SalesOrderRepository{
		db: db,
	}
}

func (s *SalesOrderRepository) GetDb() *sql.DB {
	return s.db.db
}

func (s *SalesOrderRepository) GetPendingSalesOrders(partnerId string, includeItems bool) ([]*SalesOrder, error) {
	db := s.db.GetDatabase()
	if db == nil {
		panic("Database object was returned empty")
	}
	defer db.Close()

	salesOrders := make([]*SalesOrder, 0)
	rows, err := s.db.GetDatabase().Query("exec Mobile_GetPendingSalesOrders @PartnerID = ?;", partnerId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		var seqNumber int
		var addedBy string
		var addedDate time.Time
		var customerName string
		rows.Scan(&id, &seqNumber, &addedBy, &addedDate, &customerName)
		id = toLittleEndian(id)
		salesOrder := NewSalesOrder(id, seqNumber, addedBy, addedDate, customerName)

		if includeItems == true {
			rows2, err := s.db.GetDatabase().Query("exec Mobile_GetSalesOrderItems @SalesOrderID = ?;", salesOrder.ID.String())
			if err != nil {
				return nil, err
			}
			for rows2.Next() {
				var soId uuid.UUID
				var salesoid uuid.UUID
				var productName string
				var pricePerKg decimal.Decimal
				var discountPercentage float64
				var unitName string
				var qtyInKg int
				var deliveryDeadline time.Time
				rows2.Scan(&soId, &salesoid, &productName, &pricePerKg, &discountPercentage, &unitName, &qtyInKg, &deliveryDeadline)
				soId = toLittleEndian(soId)
				salesoid = toLittleEndian(salesoid)
				salesOrderItem := NewSalesOrderItem(soId, productName, pricePerKg, discountPercentage, unitName, qtyInKg, deliveryDeadline)
				salesOrder.AddItem(salesOrderItem)
			}
			rows2.Close()
		} else {
			log.Println("Not including items...")
		}

		salesOrders = append(salesOrders, salesOrder)
	}

	return salesOrders, nil
}

// func (s *SalesOrderRepository) ApproveSalesOrder(salesOrderId string) error {

// }

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
	return &SalesOrder{
		ID:           id,
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
	return &SalesOrderItem{
		ID:               id,
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

func toLittleEndian(largeEndian uuid.UUID) uuid.UUID {
	littleEndian := uuid.NewUUID()
	for i := 8; i < 16; i++ {
		littleEndian[i] = largeEndian[i]
	}
	littleEndian[3] = largeEndian[0]
	littleEndian[2] = largeEndian[1]
	littleEndian[1] = largeEndian[2]
	littleEndian[0] = largeEndian[3]
	littleEndian[5] = largeEndian[4]
	littleEndian[4] = largeEndian[5]
	littleEndian[7] = largeEndian[6]
	littleEndian[6] = largeEndian[7]
	return littleEndian
}

func toLargeEndian(littleEndian uuid.UUID) uuid.UUID {
	largeEndian := uuid.NewUUID()
	for i := 8; i < 16; i++ {
		largeEndian[i] = littleEndian[i]
	}
	largeEndian[0] = littleEndian[3]
	largeEndian[1] = littleEndian[2]
	largeEndian[2] = littleEndian[1]
	largeEndian[3] = littleEndian[0]
	largeEndian[4] = littleEndian[5]
	largeEndian[5] = littleEndian[4]
	largeEndian[6] = littleEndian[7]
	largeEndian[7] = littleEndian[6]
	return largeEndian
}
