package db

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func Connect(dsn string) {
	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Error abriendo conexion:", err)
	}
	if err = DB.Ping(); err != nil {
		log.Fatal("Error conectando a MySQL:", err)
	}
	log.Println("Conectado a MySQL correctamente")
}
