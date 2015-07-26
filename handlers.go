package main

import (
	"encoding/json"
	"github.com/RangelReale/osin"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

var (
	authConfig *osin.ServerConfig
	authServer *osin.Server
)

func InitRoutes() {
	authConfig = osin.NewServerConfig()
	authConfig.AllowedAuthorizeTypes = osin.AllowedAuthorizeType{osin.TOKEN}
	authConfig.AllowedAccessTypes = osin.AllowedAccessType{osin.AUTHORIZATION_CODE, osin.REFRESH_TOKEN}
	authConfig.AllowGetAccessRequest = true
	authServer = osin.NewServer(authConfig, &AuthStorage{})

	r := mux.NewRouter()
	r.HandleFunc("/auth/authorize", handleLogin)
	r.HandleFunc("/auth/token", handleToken)
	r.HandleFunc(("/auth/refresh"), handleRefresh)
	r.HandleFunc("/sales/pending/{partnerId}", handlePendingSalesOrders).Methods("GET")
	r.HandleFunc("/sales/approve/{salesOrderId}", handleApproveSalesOrder).Methods("POST")

	http.Handle("/", r)
}

func handleAuthorize(w http.ResponseWriter, r *http.Request) {
	resp := authServer.NewResponse()
	defer resp.Close()
	if ar := authServer.HandleAuthorizeRequest(resp, r); ar != nil {
		ar.Authorized = true
		authServer.FinishAuthorizeRequest(resp, r, ar)
	}
	osin.OutputJSON(resp, w, r)
}

func handleloginPage(ar *osin.AuthorizeRequest, w http.ResponseWriter, r *http.Request) bool {
	r.ParseForm()
	if r.Method == "POST" {

	}
}

func handleToken(w http.ResponseWriter, r *http.Request) {
	resp := authServer.NewResponse()
	defer resp.Close()
	if ar := authServer.HandleAccessRequest(resp, r); ar != nil {
		ar.Authorized = true
		authServer.FinishAccessRequest(resp, r, ar)
	}
	osin.OutputJSON(resp, w, r)
}

func handleRefresh(w http.ResponseWriter, r *http.Request) {
	resp := authServer.NewResponse()
	defer resp.Close()
	if ar := authServer.HandleAccessRequest(resp, r); ar != nil {
		ar.GenerateRefresh = true
		authServer.FinishAccessRequest(resp, r, ar)
	}
	osin.OutputJSON(resp, w, r)
}

func handlePendingSalesOrders(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	vars := mux.Vars(r)
	partnerId := vars["partnerId"] //"DD9B4E55-958B-42C6-866F-8C18EDDCE076"
	var salesOrders []*SalesOrder
	salesOrders, err := GetPendingSalesOrders(partnerId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("%s\n", err)
		return
	}
	log.Printf("%d rows returned", len(salesOrders))
	w.Header().Set("Content-Type", "application/json; encoding=utf8")
	b, err := json.Marshal(salesOrders)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("%s\n", err)
		return
	}
	w.Write(b)
}

func handleApproveSalesOrder(w http.ResponseWriter, r *http.Request) {

}
