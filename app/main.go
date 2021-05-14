package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/tpanum/hjem"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	dbFile := flag.String("db-file", "hjem.db", "file for the database. default: hjem.db.")
	port := flag.Int("port", 8080, "port to use for the webserver. default: 8080")
	flag.Parse()

	db, err := gorm.Open(sqlite.Open(*dbFile), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	s := hjem.NewServer(db)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), s.Routes()); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
