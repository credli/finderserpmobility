package main

import (
	"net/http"
)

func InitRoutes() {
	r := mux.NewRouter()
	r.HandleFunc("/auth/login", handleLogin).Methods("POST")
	r.HandleFunc("/sales/pending/{partnerId}", handlePendingSalesOrders).Methods("GET")
	r.HandleFunc("/sales/approve/{salesOrderId}", handleApproveSalesOrder).Methods("POST")
	http.Handle(r)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {

}

func handlePendingSalesOrders(w http.ResponseWriter, r *http.Request) {

}

func handleApproveSalesOrder(w http.ResponseWriter, r *http.Request) {

}
