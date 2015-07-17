package main

import (
	"github.com/shopspring/decimal"
	"time"
)

type SalesOrder struct {
	ID           string            `json: "id"`
	SeqNumber    int               `json: "seqNumber"`
	AddedBy      string            `json: "addedBy"`
	AddedDate    time.Time         `json: "addedDate"`
	CustomerName string            `json: "customerName"`
	Items        []*SalesOrderItem `json: "items"`
	GrandTotal   decimal.Decimal   `json: "grandTotal"`
}

func NewSalesOrder(id string, seq int, addedBy string, addedDate time.Time, customerName string) *SalesOrder {
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
}

func (s *SalesOrder) CalculateGrandTotal() decimal.Decimal {
	total := decimal.NewFromFloat(0)
	for i := range s.Items {
		total.Add(s.Items[i].LineTotal)
	}
	s.GrandTotal = total
	return s.GrandTotal
}

type SalesOrderItem struct {
	ID               string          `json: "id"`
	ProductName      string          `json: "productName"`
	PricePerKG       decimal.Decimal `json: "pricePerKG"`
	DiscountPercent  float64         `json: "discountPercent"`
	UnitName         string          `json: "unitName"`
	QtyInKG          int             `json: "qtyInKG"`
	DeliveryDeadline time.Time       `json: "deliveryDeadline"`
	LineTotal        decimal.Decimal `json: "lineTotal"`
}

func NewSalesOrderItem(id, productName string, price string, discountPercent float64, unitName string, qty int, deliveryDeadline time.Time) (*SalesOrderItem, error) {
	decimalPrice, err := decimal.NewFromString(price)
	if err != nil {
		return nil, err
	}
	return &SalesOrderItem{
		ID:               id,
		ProductName:      productName,
		PricePerKG:       decimalPrice,
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
