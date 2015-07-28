package main

import (
	"database/sql"
	"github.com/RangelReale/osin"
	"github.com/pborman/uuid"
	"log"
	"time"
)

type AuthStorage struct {
	osin.Storage
	db *sql.DB
}

func NewAuthStorage(db *sql.DB) *AuthStorage {
	return &AuthStorage{
		db: db,
	}
}

func (s *AuthStorage) Clone() osin.Storage {
	return s
}

func (s *AuthStorage) Close() {
}

func (s *AuthStorage) GetClient(id string) (osin.Client, error) {
	var (
		clientID    string
		secret      string
		redirectUri string
	)
	row := s.db.QueryRow("SELECT * FROM clients WHERE id = ?", id)

	err := row.Scan(&clientID, &secret, &redirectUri)
	if err != nil {
		return nil, err
	}

	return &osin.DefaultClient{
		Id:          clientID,
		Secret:      secret,
		RedirectUri: redirectUri,
	}, nil
}

func (s *AuthStorage) SetClient(client osin.Client) error {
	stmt, err := s.db.Prepare("INSERT INTO clients(id, secret, redirect_uri) VALUES (?, ?, ?)")
	_, err = stmt.Exec(client.GetId(), client.GetSecret(), client.GetRedirectUri())
	return err
}

func (s *AuthStorage) RemoveClient(id string) error {
	stmt, err := s.db.Prepare("DELETE FROM clients WHERE id = ?")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(id)
	return err
}

func (s *AuthStorage) SaveAuthorize(auth *osin.AuthorizeData) error {
	log.Printf("UserData: %s\n", auth.UserData)
	var userId *uuid.UUID
	if user, ok := auth.UserData.(*User); ok == true {
		userId = &user.UserId
	}
	log.Printf("userId at this point is %s", userId)
	stmt, err := s.db.Prepare(`
		INSERT INTO authorize_data(code, expires_in, scope, redirect_uri, state, created_at, client_id, user_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(auth.Code, auth.ExpiresIn, auth.Scope, auth.RedirectUri, auth.State, auth.CreatedAt, auth.Client.GetId(), userId.String())
	return err
}

func (s *AuthStorage) LoadAuthorize(code string) (*osin.AuthorizeData, error) {
	var (
		authCode    string
		expiresIn   int32
		scope       string
		redirectUri string
		state       string
		createdAt   time.Time
		clientID    string
		userID      *uuid.UUID
	)

	row := s.db.QueryRow("SELECT * FROM authorize_data WHERE code = ?", code)
	err := row.Scan(&authCode, &expiresIn, &scope, &redirectUri, &state, &createdAt, &clientID, &userID)
	if err != nil {
		return nil, err
	}

	client, err := s.GetClient(clientID)
	if err != nil {
		return nil, err
	}

	var user *User
	if userID != nil {
		uid := toLittleEndian(*userID)
		user, _ = userRepo.GetUser(uid)
		log.Printf("Yo! user: %s\n%s\n%s\n", user, userID, uid)
	}

	authData := &osin.AuthorizeData{
		Code:        authCode,
		ExpiresIn:   expiresIn,
		Scope:       scope,
		RedirectUri: redirectUri,
		State:       state,
		CreatedAt:   createdAt,
		Client:      client,
		UserData:    user,
	}

	return authData, nil
}

func (s *AuthStorage) RemoveAuthorize(code string) error {
	stmt, err := s.db.Prepare("DELETE FROM authorize_data WHERE code = ?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(code)
	return err
}

func (s *AuthStorage) SaveAccess(access *osin.AccessData) error {
	log.Printf("UserData: %s\n", access.UserData)
	var userId *uuid.UUID
	if user, ok := access.UserData.(*User); ok == true {
		userId = &user.UserId
	}
	log.Printf("userId at this point is %s", userId)
	stmt, err := s.db.Prepare(`
		INSERT INTO access_data(access_token, refresh_token, expires_in,
			scope, redirect_uri, created_at, authorize_data_code, prev_access_data_token, client_id, user_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`)
	if err != nil {
		return err
	}
	prevAccessDataToken := ""
	if access.AccessData != nil {
		prevAccessDataToken = access.AccessData.AccessToken
	}

	authDataCode := ""
	if access.AuthorizeData != nil {
		authDataCode = access.AuthorizeData.Code
	}

	_, err = stmt.Exec(access.AccessToken, access.RefreshToken, access.ExpiresIn, access.Scope,
		access.RedirectUri, access.CreatedAt, authDataCode, prevAccessDataToken, access.Client.GetId(), userId.String())
	return err
}

func (s *AuthStorage) loadAccess(token string, isRefresh ...bool) (*osin.AccessData, string, string, string, error) {
	var (
		accessToken         string
		refreshToken        string
		expiresIn           int32
		scope               string
		redirectUri         string
		createdAt           time.Time
		authorizeDataCode   string
		prevAccessDataToken string
		clientID            string
		userID              *uuid.UUID
	)

	var rows *sql.Rows
	var err error
	if len(isRefresh) > 0 && isRefresh[0] == true {
		rows, err = s.db.Query("SELECT * FROM access_data WHERE refresh_token = ?", token)
	} else {
		rows, err = s.db.Query("SELECT * FROM access_data WHERE access_token = ?", token)
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&accessToken, &refreshToken, &expiresIn, &scope, &redirectUri, &createdAt,
			&authorizeDataCode, &prevAccessDataToken, &clientID, &userID)
		if err != nil {
			return nil, "", "", "", err
		}
		break
	}

	var user *User
	if userID != nil {
		uid := toLittleEndian(*userID)
		user, _ = userRepo.GetUser(uid)
	}

	return &osin.AccessData{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		Scope:        scope,
		RedirectUri:  redirectUri,
		CreatedAt:    createdAt,
		UserData:     user,
	}, authorizeDataCode, prevAccessDataToken, clientID, err
}

func (s *AuthStorage) LoadAccess(token string) (*osin.AccessData, error) {
	accessData, _, prevAccessDataToken, clientID, err := s.loadAccess(token)
	//load previous access data if the token is not empty
	var prevAccessData *osin.AccessData
	if prevAccessDataToken != "" {
		prevAccessData, _, _, _, err = s.loadAccess(prevAccessDataToken)
		if err != nil {
			return nil, err
		}
	}
	//load client data
	client, err := s.GetClient(clientID)
	if err != nil {
		return nil, err
	}
	//load authorize data, cannot find it since generating the token deletes it for some reason
	// authData, err := s.LoadAuthorize(authDataCode)

	accessData.Client = client
	//accessData.AuthorizeData = authData
	accessData.AccessData = prevAccessData
	return accessData, err
}

func (s *AuthStorage) RemoveAccess(token string) error {
	stmt, err := s.db.Prepare("DELETE FROM access_data WHERE access_token = ?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(token)
	return err
}

func (s *AuthStorage) LoadRefresh(token string) (*osin.AccessData, error) {
	accessData, authDataCode, prevAccessDataToken, clientID, err := s.loadAccess(token, true)
	//load previous access data if token is not empty
	var prevAccessData *osin.AccessData
	if prevAccessDataToken != "" {
		prevAccessData, _, _, _, err = s.loadAccess(prevAccessDataToken)
		if err != nil {
			return nil, err
		}
	}
	//load client data
	client, err := s.GetClient(clientID)
	if err != nil {
		return nil, err
	}
	//load authorize data
	authData, err := s.LoadAuthorize(authDataCode)

	accessData.Client = client
	accessData.AuthorizeData = authData
	accessData.AccessData = prevAccessData
	return accessData, err
}

func (s *AuthStorage) RemoveRefresh(token string) error {
	stmt, err := s.db.Prepare("DELETE FROM access_data WHERE refresh_token = ?")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(token)
	return err
}
