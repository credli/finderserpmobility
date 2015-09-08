package main

import (
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/gorilla/context"
)

const lenStatic = len("/static/")

func InitAppHandlers(r *mux.Router, oAuth *OAuthHandler) {
	r.HandleFunc("/", handleIndex).Methods("GET")
	r.HandleFunc("/sales/pending/{partnerId}", oAuth.MiddlewareFunc(handlePendingSalesOrders)).Methods("GET")
	r.HandleFunc("/sales/approve/{salesOrderId}", oAuth.MiddlewareFunc(handleApproveSalesOrder)).Methods("POST")
	r.HandleFunc("/sales/reject/{salesOrderId}", oAuth.MiddlewareFunc(handleRejectSalesOrder)).Methods("POST")
	r.HandleFunc("/products", handleProducts).Methods("GET")
}

func handleStatic(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Path[lenStatic:]
	serveFileStatic(w, r, file)
}

func serveFileStatic(w http.ResponseWriter, r *http.Request, filename string) {
	filePath := filepath.Join("static", filename)
	if !pathExists(filePath) {
		http.Error(w, deferror.Get(E_FILE_NOT_EXISTS, filePath), http.StatusNotFound)
		return
	}
	http.ServeFile(w, r, filePath)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index.html", nil)
}

func handleMe(w http.ResponseWriter, r *http.Request) {
	data := context.Get(r, USERDATA)
	user, ok := data.(*User)
	if ok == false {
		http.Error(w, deferror.Get(E_INVALID_CONTEXT), http.StatusInternalServerError)
		return
	}
	userInfo, err := userRepo.GetUser(user.UserId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("%s\n", err)
		return
	}
	b, err := json.Marshal(userInfo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("%s\n", err)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(b)
}

func handlePendingSalesOrders(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	vars := mux.Vars(r)
	includeItems := r.URL.Query().Get("includeItems")
	inclItm := false
	if includeItems != "" {
		inclItm, _ = strconv.ParseBool(includeItems)
	}
	partnerId := vars["partnerId"] //"DD9B4E55-958B-42C6-866F-8C18EDDCE076"
	var salesOrders []*SalesOrder
	err := salesOrderRepo.GetPendingSalesOrders(&salesOrders, partnerId, inclItm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("%s\n", err)
		return
	}
	b, err := json.Marshal(salesOrders)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("%s\n", err)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(b)
}

func handleApproveSalesOrder(w http.ResponseWriter, r *http.Request) {
	data := context.Get(r, USERDATA)
	user, ok := data.(*User)
	if ok == false {
		http.Error(w, deferror.Get(E_INVALID_CONTEXT), http.StatusInternalServerError)
		return
	}
	r.ParseForm()
	vars := mux.Vars(r)
	salesOrderId := vars["salesOrderId"]
	generateDeliveryRequest := r.Form.Get("generateDeliveryRequest")

	if salesOrderId == "" {
		http.Error(w, deferror.Get(E_MISSING_VALUE, "salesOrderId"), http.StatusBadRequest)
		return
	}
	generateDeliveryRequestBool, err := strconv.ParseBool(generateDeliveryRequest)
	if err != nil {
		http.Error(w, deferror.Get(E_UNRECOGNIZED_VALUE, generateDeliveryRequest, "generateDeliveryRequest"), http.StatusBadRequest)
		log.Printf("ERROR: %s\n", err)
		return
	}
	privileged, err := userRepo.UserHasAdminPrivileges(user.UserId)
	if err != nil {
		http.Error(w, deferror.Get(E_INVALID_PRIVILEGES), http.StatusUnauthorized)
		log.Printf("ERROR: %s\n", err)
		return
	}
	if privileged == false {
		http.Error(w, deferror.Get(E_NO_PERMISSION), http.StatusUnauthorized)
		return
	}
	result, desc, err := salesOrderRepo.ApproveSalesOrder(salesOrderId, generateDeliveryRequestBool, user.UserId)
	writeResult(w, result, desc, err)
}

func handleRejectSalesOrder(w http.ResponseWriter, r *http.Request) {
	data := context.Get(r, USERDATA)
	user, ok := data.(*User)
	if ok == false {
		http.Error(w, deferror.Get(E_INVALID_CONTEXT), http.StatusInternalServerError)
		return
	}
	r.ParseForm()
	vars := mux.Vars(r)
	salesOrderId := vars["salesOrderId"]
	reason := r.Form.Get("reason")

	if salesOrderId == "" {
		http.Error(w, deferror.Get(E_MISSING_VALUE, "salesOrderId"), http.StatusBadRequest)
		return
	}
	if reason == "" {
		http.Error(w, deferror.Get(E_MISSING_VALUE, "reason"), http.StatusBadRequest)
		return
	}
	privileged, err := userRepo.UserHasAdminPrivileges(user.UserId)
	if err != nil {
		http.Error(w, deferror.Get(E_INVALID_PRIVILEGES), http.StatusUnauthorized)
		log.Printf("ERROR: %s\n", err)
		return
	}
	if privileged == false {
		http.Error(w, deferror.Get(E_NO_PERMISSION), http.StatusUnauthorized)
		return
	}
	result, desc, err := salesOrderRepo.RejectSalesOrder(salesOrderId, reason, user.UserId)
	writeResult(w, result, desc, err)
}

func fillProducts(p *[]*Product) error {
	var err error
	*p, err = productsRepo.GetProducts()
	if err != nil {
		return err
	}
	return nil
}

func handleProducts(w http.ResponseWriter, r *http.Request) {
	_ = "breakpoint"
	var products []*Product

	if cacheErr := ReadFromCache(CacheProductsKeyInfo, &products); cacheErr != nil {
		//the memcache error is something other than 'no server' or 'server unavailable', then its worth reporting
		if cacheErr != ErrNoServers && cacheErr != ErrServerError && cacheErr != ErrCacheMiss && cacheErr.Error() != "EOF" {
			http.Error(w, cacheErr.Error(), http.StatusInternalServerError)
			log.Printf("%s\n", cacheErr)
			return
		}

		//fetch products from db
		//products, err := productsRepo.GetProducts()
		err := fillProducts(&products)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Printf("%s\n", err)
			return
		}

		//if we miss the cache, then update it for later requires
		if cacheErr == ErrCacheMiss {
			err = WriteToCache(CacheProductsKeyInfo, products)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				log.Printf("%s\n", err)
				return
			}
		}
	}

	b, err := json.Marshal(products)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("%s\n", err)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(b)
}

func writeResult(w http.ResponseWriter, result, description string, err error) {
	if err != nil {
		http.Error(w, deferror.Get(E_UNKOWN_ERROR), http.StatusInternalServerError)
		log.Printf("ERROR: %s\n", err)
		return
	}
	if result == "ERROR" {
		w.WriteHeader(http.StatusNotAcceptable)
	}
	output := struct {
		Result      string `json:"result"`
		Description string `json:"description"`
	}{
		Result:      result,
		Description: description,
	}
	b, err := json.Marshal(output)
	if err != nil {
		http.Error(w, deferror.Get(E_PARSER_ERROR), http.StatusInternalServerError)
		log.Printf("ERROR: %s\n", err)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(b)
}
