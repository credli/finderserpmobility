package data

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

func (s *SalesOrderRepository) GetPendingSalesOrders(partnerId string) ([]*SalesOrder, error) {
	db, err := s.db.GetDatabase()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	salesOrders := make([]*SalesOrder, 0)
	rows, err := s.db.Query("exec Mobile_GetPendingSalesOrders @PartnerID = ?", partnerId)
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
		salesOrder := NewSalesOrder(id, seqNumber, addedBy, addedDate, customerName)
		salesOrders = append(salesOrders, salesOrder)
	}
	return salesOrders, nil
}

func (s *SalesOrderRepository) ApproveSalesOrder(salesOrderId string) error {

}
