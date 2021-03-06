package main

import (
	"database/sql"
	"github.com/RangelReale/osin"
	//"log"
	"net/http"
	"net/url"
)

type LoginPage struct {
	ResponseType osin.AuthorizeRequestType
	ClientId     string
	State        string
	RedirectUri  string
	Username     string
	LoginError   bool
}

type OAuthHandler struct {
	config  *osin.ServerConfig
	server  *osin.Server
	Storage *AuthStorage
	db      *sql.DB
}

func NewOAuthHandler(db *sql.DB) *OAuthHandler {
	config := osin.NewServerConfig()
	config.AllowedAuthorizeTypes = osin.AllowedAuthorizeType{osin.CODE, osin.TOKEN}
	config.AllowedAccessTypes = osin.AllowedAccessType{osin.AUTHORIZATION_CODE, osin.REFRESH_TOKEN}
	storage := NewAuthStorage(db)
	server := osin.NewServer(config, storage)
	return &OAuthHandler{config, server, storage, db}
}

func (o *OAuthHandler) AuthorizeClient(w http.ResponseWriter, r *http.Request) {
	server := o.server
	resp := server.NewResponse()
	defer resp.Close()

	if ar := server.HandleAuthorizeRequest(resp, r); ar != nil {
		if !o.handleLoginPage(ar, w, r) {
			return
		}
		ar.Authorized = true
		server.FinishAuthorizeRequest(resp, r, ar)
	}
	if resp.IsError || resp.InternalError != nil {
		//log.Printf("ERROR: %s\n", resp.InternalError)
		resp.StatusCode = http.StatusBadRequest
	}
	osin.OutputJSON(resp, w, r)
}

func (o *OAuthHandler) handleLoginPage(ar *osin.AuthorizeRequest, w http.ResponseWriter, r *http.Request) bool {
	_ = "breakpoint"
	r.ParseForm()
	username := ""
	password := ""
	loginError := false
	if r.Method == "POST" {
		username = r.Form.Get("username")
		password = r.Form.Get("password")
		user, _ := userRepo.Login(username, password)
		ar.UserData = user
		loginError = (user == nil)
		if user != nil || loginError == false {
			return true
		}
		//return (user != nil)
	}
	page := &LoginPage{
		ResponseType: ar.Type,
		ClientId:     ar.Client.GetId(),
		State:        ar.State,
		RedirectUri:  url.QueryEscape(ar.RedirectUri),
		Username:     username,
		LoginError:   loginError,
	}
	renderLoginPage(w, page)
	return false
}

func renderLoginPage(w http.ResponseWriter, page *LoginPage) {
	err := templates.ExecuteTemplate(w, "login.html", page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (o *OAuthHandler) GenerateToken(w http.ResponseWriter, r *http.Request) {
	server := o.server
	resp := server.NewResponse()
	defer resp.Close()
	if ar := server.HandleAccessRequest(resp, r); ar != nil {
		switch ar.Type {
		case osin.AUTHORIZATION_CODE:
			ar.Authorized = true
		case osin.REFRESH_TOKEN:
			ar.Authorized = true
		case osin.CLIENT_CREDENTIALS:
			ar.Authorized = true
		}
		server.FinishAccessRequest(resp, r, ar)
	}
	if resp.IsError || resp.InternalError != nil {
		//log.Printf("ERROR: %s\n", resp.InternalError)
		resp.StatusCode = http.StatusBadRequest
	}
	osin.OutputJSON(resp, w, r)
}

func (o *OAuthHandler) HandleInfo(w http.ResponseWriter, r *http.Request) {
	server := o.server
	resp := server.NewResponse()
	defer resp.Close()
	if ir := server.HandleInfoRequest(resp, r); ir != nil {
		server.FinishInfoRequest(resp, r, ir)
	}
	if resp.IsError || resp.InternalError != nil {
		//log.Printf("ERROR: %s\n", resp.InternalError) //unnecessary to report on bad token
		resp.StatusCode = http.StatusBadRequest
	}
	osin.OutputJSON(resp, w, r)
}
