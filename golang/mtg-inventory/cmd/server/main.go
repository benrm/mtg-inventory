/*
Executable server is the HTTP server that serves out the Handler in the main
package. This right now is just for testing; the finished server will have
authentication.
*/
package main

import (
	"database/sql"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"

	inventory "github.com/benrm/mtg-inventory/golang/mtg-inventory"
	intsql "github.com/benrm/mtg-inventory/golang/mtg-inventory/backends/sql"
	_ "github.com/go-sql-driver/mysql"
)

var (
	bindAddr = flag.String("bind", "127.0.0.1:8080", "The address to bind to")
)

func main() {
	flag.Parse()

	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		log.Fatalf("MYSQL_DSN is not set")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error opening database connection: %s", err.Error())
	}

	handler := inventory.NewHandler(intsql.NewBackend(db), nil)

	server := &http.Server{
		Addr:    *bindAddr,
		Handler: handler,
	}

	err = server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Unexpected error on server close: %s", err.Error())
	}
}
