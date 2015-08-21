package main

import (
	"database/sql"
	"log"
	"time"
)

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
