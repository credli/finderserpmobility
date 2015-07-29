package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/RangelReale/osin"
	c "github.com/credli/finderserpmobility/config"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
)

var (
	salesOrderRepo *SalesOrderRepository
	userRepo       *UserRepository

	config    = c.NewConfig()
	templates = template.Must(template.ParseGlob("tmpl/*.html"))
)

func main() {
	db, err := sql.Open(config.DbDriverName, config.DbConnectionString)
	if err != nil {
		log.Panicln(err)
	}
	err = db.Ping()
	if err != nil {
		log.Panicln(err)
	}
	defer db.Close()

	//repositories
	salesOrderRepo = NewSalesOrderRepository(db)
	userRepo = NewUserRepository(db)
	oAuth := NewOAuthHandler(db)

	if _, err := oAuth.Storage.GetClient("finderserpmobility"); err != nil {
		if err = oAuth.Storage.SetClient(&osin.DefaultClient{
			Id:          "finderserpmobility",
			Secret:      "smellycat",
			RedirectUri: "http://localhost:5001/",
		}); err != nil {
			panic(err)
		}
	}

	mainRouter := mux.NewRouter()

	//oauth2 routes
	oauthSub := mainRouter.PathPrefix("/oauth2").Subrouter()
	oauthSub.HandleFunc("/authorize", oAuth.AuthorizeClient)
	oauthSub.HandleFunc("/token", oAuth.GenerateToken)
	oauthSub.HandleFunc("/info", oAuth.HandleInfo)
	//API routes
	mainRouter.HandleFunc("/", handleIndex).Methods("GET")
	mainRouter.HandleFunc("/sales/pending/{partnerId}", oAuth.MiddlewareFunc(handlePendingSalesOrders)).Methods("GET")
	mainRouter.HandleFunc("/sales/approve/{salesOrderId}", oAuth.MiddlewareFunc(handleApproveSalesOrder)).Methods("POST")
	//static routes
	http.HandleFunc("/static/", handleStatic)
	http.Handle("/", mainRouter)

	//listen and serve (default port is 5001)...
	fmt.Printf("Listening on %s\n", config.HostAddr)
	http.ListenAndServe(config.HostAddr, nil)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index.html", nil)
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
	salesOrders, err := salesOrderRepo.GetPendingSalesOrders(partnerId, inclItm)
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
	w.Header().Set("Content-Type", "application/json; encoding=utf8")
	w.Write(b)
}

func handleApproveSalesOrder(w http.ResponseWriter, r *http.Request) {
	data := context.Get(r, USERDATA)
	user, ok := data.(*User)
	if ok == false {
		http.Error(w, "Context data is invalid", http.StatusInternalServerError)
		return
	}
	r.ParseForm()
	vars := mux.Vars(r)
	salesOrderId := vars["salesOrderId"]
	log.Printf("salesOrderId: %s", salesOrderId)
	generateDeliveryRequest := r.Form.Get("generateDeliveryRequest")

	if salesOrderId == "" {
		http.Error(w, "salesOrderId is not specified", http.StatusBadRequest)
		return
	}
	generateDeliveryRequestBool, err := strconv.ParseBool(generateDeliveryRequest)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not recognize %s as boolean value in generateDeliveryRequest", generateDeliveryRequest), http.StatusBadRequest)
		log.Printf("ERROR: %s\n", err)
		return
	}
	privileged, err := userRepo.UserHasAdminPrivileges(user.UserId)
	if err != nil {
		http.Error(w, "Could not determine user privileges", http.StatusUnauthorized)
		log.Printf("ERROR: %s\n", err)
		return
	}
	if privileged == false {
		http.Error(w, "You do not have permission to approve sales orders", http.StatusUnauthorized)
		return
	}
	result, err := salesOrderRepo.ApproveSalesOrder(salesOrderId, generateDeliveryRequestBool, user.UserId.String())
	log.Printf("Result: %s\n", result)
	if err != nil {
		http.Error(w, "Could not approve sales order", http.StatusInternalServerError)
		log.Printf("ERROR: %s\n", err)
		return
	}
	output := struct {
		Result string `json:"result"`
	}{
		Result: result,
	}
	b, err := json.Marshal(output)
	if err != nil {
		http.Error(w, "Could not parse json result output", http.StatusInternalServerError)
		log.Printf("ERROR: %s\n", err)
		return
	}
	w.Header().Set("Content-Type", "application/json; encoding=utf8")
	w.Write(b)
}

const lenStatic = len("/static/")

func handleStatic(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Path[lenStatic:]
	serveFileStatic(w, r, file)
}

func serveFileStatic(w http.ResponseWriter, r *http.Request, filename string) {
	filePath := filepath.Join("static", filename)
	if !pathExists(filePath) {
		http.Error(w, fmt.Sprintf("File %s does not exist", filePath), http.StatusNotFound)
		return
	}
	http.ServeFile(w, r, filePath)
}
