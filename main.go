package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"runtime"

	"github.com/RangelReale/osin"
	_ "github.com/alexbrainman/odbc"
	c "github.com/credli/finderserpmobility/config"
	"github.com/gorilla/mux"
)

var (
	salesOrderRepo *SalesOrderRepository
	userRepo       *UserRepository
	productsRepo   *ProductsRepository

	config    = c.NewConfig()
	templates = template.Must(template.ParseGlob("tmpl/*.html"))
)

type Repository interface{}

func main() {
	//enable parallelism
	runtime.GOMAXPROCS(2)

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
	productsRepo = NewProductsRepository(db)
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
	//API routes
	InitAppHandlers(mainRouter, oAuth)
	mainRouter.HandleFunc("/me", oAuth.MiddlewareFunc(handleMe)).Methods("GET")

	//oauth2 routes
	oauthSub := mainRouter.PathPrefix("/oauth2").Subrouter()
	oauthSub.HandleFunc("/authorize", oAuth.AuthorizeClient)
	oauthSub.HandleFunc("/token", oAuth.GenerateToken)
	oauthSub.HandleFunc("/info", oAuth.HandleInfo)

	//static routes
	http.HandleFunc("/static/", handleStatic)
	http.Handle("/", mainRouter)

	//listen and serve (default port is 5001)...
	fmt.Printf("Listening on %s\n", config.HostAddr)
	http.ListenAndServe(config.HostAddr, nil)
}
