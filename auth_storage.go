package main

import (
	"database/sql"
	"github.com/RangelReale/osin"
)

type AuthStorage struct {
	osin.Storage
	clients   map[string]osin.Client
	authorize map[string]*osin.AuthorizeData
	access    map[string]*osin.AccessData
	refresh   map[string]string
	db        *sql.DB
}

func NewAuthStorage() *AuthStorage {
	r = &AuthStorage{
		clients:   make(map[string]osin.Client),
		authorize: make(map[string]*osin.AuthorizeData),
		access:    make(map[string]*osin.AccessData),
		refresh:   make(map[string]string),
	}

	r.db, err := sql.Open(driverName, dataSourceName)

	r.clients["finderserpapp"] = &osin.DefaultClient{
		Id:          "finderserpapp",
		Secret:      "MyInternet!@#",
		RedirectUri: "",
	}

	return r
}

func (s *AuthStorage) Clone() osin.Storage {
	return s
}

func (s *AuthStorage) Close() {
	s.db.Close()
}
