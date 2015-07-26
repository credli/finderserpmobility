package main

import (
	"encoding/json"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

func InitRoutes() {
	r := mux.NewRouter()
	r.HandleFunc("/sales/pending/{partnerId}", handlePendingSalesOrders).Methods("GET")
	r.HandleFunc("/sales/approve/{salesOrderId}", handleApproveSalesOrder).Methods("POST")
	http.Handle("/", r)
}

func handlePendingSalesOrders(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	vars := mux.Vars(r)
	includeItems := r.URL.Query().Get("includeItems")
	inclItm := false
	if includeItems != "" {
		inclItm, _ = strconv.ParseBool(includeItems)
		log.Printf("includeItems: %s inclItm: %d\n", includeItems, inclItm)
	}
	partnerId := vars["partnerId"] //"DD9B4E55-958B-42C6-866F-8C18EDDCE076"
	var salesOrders []*SalesOrder
	db := NewDatabase(config.DbDriverName, config.DbConnectionString)
	salesOrderRepo := NewSalesOrderRepository(db)
	log.Printf("inclItm:%s\n", inclItm)
	salesOrders, err := salesOrderRepo.GetPendingSalesOrders(partnerId, inclItm)
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
