package main

import (
	"log"
	"net/http"
	"os"

	"taller_vehiculos/db"
	"taller_vehiculos/handlers"
)

func main() {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		log.Fatal("Define MYSQL_DSN antes de iniciar el servidor")
	}
	db.Connect(dsn)

	mux := http.NewServeMux()

	// Frontend estatico
	mux.Handle("/", http.FileServer(http.Dir("./public")))

	// Clientes
	mux.HandleFunc("/api/clientes", handlers.Clientes)
	mux.HandleFunc("/api/clientes/", handlers.ClienteById)

	// Vehiculos
	mux.HandleFunc("/api/vehiculos", handlers.Vehiculos)

	// Empleados
	mux.HandleFunc("/api/empleados", handlers.Empleados)

	// Servicios
	mux.HandleFunc("/api/servicios", handlers.Servicios)

	// Repuestos
	mux.HandleFunc("/api/repuestos", handlers.Repuestos)
	mux.HandleFunc("/api/repuestos/bajo-stock", handlers.RepuestosBajoStock)

	// Ordenes
	mux.HandleFunc("/api/ordenes", handlers.Ordenes)
	mux.HandleFunc("/api/ordenes/", handlers.OrdenDetalle)

	// CDA
	mux.HandleFunc("/api/revisionescda", handlers.RevisionesCDA)

	log.Println("Servidor corriendo en http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", corsMiddleware(mux)))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
