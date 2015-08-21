package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/RangelReale/osin"
	_ "github.com/alexbrainman/odbc"
	c "github.com/credli/finderserpmobility/config"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	//"runtime"
	"strconv"
)

var (
	salesOrderRepo *SalesOrderRepository
	userRepo       *UserRepository

	config    = c.NewConfig()
	templates = template.Must(template.ParseGlob("tmpl/*.html"))
)

type Repository interface{}

func main() {
	//enable parallelism
	//runtime.GOMAXPROCS(4)

	db, err := sql.Open("odbc", config.DbConnectionString)
	//db.SetMaxIdleConns(0)
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

	if _, err := oAuth.Storage.GetClient("finderserpmobilityapp"); err != nil {
		if err = oAuth.Storage.SetClient(&osin.DefaultClient{
			Id:          "finderserpmobilityapp",
			Secret:      "spookyturtle",
			RedirectUri: "finderserpmobility://",
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
	mainRouter.HandleFunc("/me", oAuth.MiddlewareFunc(handleMe)).Methods("GET")
	mainRouter.HandleFunc("/sales/pending/{partnerId}", oAuth.MiddlewareFunc(handlePendingSalesOrders)).Methods("GET")
	mainRouter.HandleFunc("/sales/approve/{salesOrderId}", oAuth.MiddlewareFunc(handleApproveSalesOrder)).Methods("POST")
	mainRouter.HandleFunc("/sales/reject/{salesOrderId}", oAuth.MiddlewareFunc(handleRejectSalesOrder)).Methods("POST")
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

const lenStatic = len("/static/")

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
