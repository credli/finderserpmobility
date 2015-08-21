package main

import (
	"database/sql"
	"fmt"
	//"sync"
	// "errors"
	"time"
)

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
	fmt.Printf("Started:%v\n", time.Now())

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
		salesOrders = append(salesOrders, salesOrder)
	}

	_ = "breakpoint"
	if includeItems == true {
		itemsChan := make(chan []*SalesOrderItem)
		errChan := make(chan error)

		ordersCount := len(salesOrders)

		for i := 0; i < ordersCount; i++ {
			go func(salesOrderId string) {
				db, err := sql.Open("odbc", config.DbConnectionString)
				if err != nil {
					errChan <- err
					return
				}
				defer db.Close()

				rows, err := db.Query("exec Mobile_GetSalesOrderItems @SalesOrderID = ?;", salesOrderId)
				if err != nil {
					errChan <- err
					return
				}
				defer rows.Close()

				var salesOrderItems = make([]*SalesOrderItem, 0)
				for rows.Next() {
					var (
						soId               string
						salesoid           string
						productName        string
						pricePerKg         float64
						discountPercentage float64
						unitName           string
						qtyInKg            int
						deliveryDeadline   time.Time
					)
					rows.Scan(&soId, &salesoid, &productName, &pricePerKg, &discountPercentage, &unitName, &qtyInKg, &deliveryDeadline)
					salesOrderItem, err := NewSalesOrderItem(soId, salesoid, productName, pricePerKg, discountPercentage, unitName, qtyInKg, deliveryDeadline)
					if err != nil {
						errChan <- err
						return
					}
					salesOrderItems = append(salesOrderItems, salesOrderItem)
				}
				itemsChan <- salesOrderItems
			}(salesOrders[i].ID)
		}

		for i := 0; i < ordersCount; i++ {
			select {
			case items := <-itemsChan:
				if len(items) > 0 {
					soId := items[0].SalesOrderID //just pick the first item, info is the same in all other items
					for _, so := range salesOrders {
						if so.ID == soId {
							for _, soi := range items {
								so.AddItem(soi)
							}
							break
						}
					}
				}
			case err := <-errChan:
				return nil, err
			}
		}
	}
	fmt.Printf("Ended:%v\n", time.Now())

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
	GrandTotal   float64           `json:"grandTotal"`
}

func NewSalesOrder(id string, seq int, addedBy string, addedDate time.Time, customerName string) (*SalesOrder, error) {
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
	item.CalculateLineTotal()
	s.Items = append(s.Items, item)
	s.CalculateGrandTotal()
}

func (s *SalesOrder) CalculateGrandTotal() float64 {
	tally := 0.0
	for _, item := range s.Items {
		tally += item.LineTotal
	}
	s.GrandTotal = tally
	return s.GrandTotal
}

type SalesOrderItem struct {
	ID               string    `json:"id"`
	SalesOrderID     string    `json:"salesOrderId"`
	ProductName      string    `json:"productName"`
	PricePerKG       float64   `json:"pricePerKG"`
	DiscountPercent  float64   `json:"discountPercent"`
	UnitName         string    `json:"unitName"`
	QtyInKG          int       `json:"qtyInKG"`
	DeliveryDeadline time.Time `json:"deliveryDeadline"`
	LineTotal        float64   `json:"lineTotal"`
}

func NewSalesOrderItem(id string, soid string, productName string, price float64, discountPercent float64, unitName string, qty int, deliveryDeadline time.Time) (*SalesOrderItem, error) {
	return &SalesOrderItem{
		ID:               id,
		SalesOrderID:     soid,
		ProductName:      productName,
		PricePerKG:       price,
		DiscountPercent:  discountPercent,
		UnitName:         unitName,
		QtyInKG:          qty,
		DeliveryDeadline: deliveryDeadline,
	}, nil
}

func (s *SalesOrderItem) CalculateLineTotal() float64 {
	total := s.PricePerKG * float64(s.QtyInKG)
	discountPercentValue := total * s.DiscountPercent / 100
	if total > 0 && s.DiscountPercent > 0 && (total-discountPercentValue) > 0 {
		total = total - discountPercentValue
	}
	s.LineTotal = total
	return total
}
