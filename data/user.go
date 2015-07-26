package data

type User struct {
	UserId   string
	UserName string
	Password string
	Email    string
}

type UserRepository struct {
	Repository
	db *Database
}

func NewUserRepository(db *Database) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (s *SalesOrderRepository) GetDb() *sql.DB {
	return s.db.db
}

func (s *SalesOrderRepository) GetUserByUserName(string username) *User {

}
