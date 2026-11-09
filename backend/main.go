package main

import (
	"log"
	"net/http"
	"os"

	"taller_vehiculos/db"
	"taller_vehiculos/handlers"
	"taller_vehiculos/mongodb"
)

func main() {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		log.Fatal("Define MYSQL_DSN antes de iniciar el servidor")
	}
	db.Connect(dsn)

	if err := mongodb.Connect(); err != nil {
		log.Fatal("Error conectando a MongoDB: ", err)
	}
	log.Println("Conectado a MongoDB Atlas correctamente")

	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.Dir("./public")))
	mux.HandleFunc("/api/clientes", handlers.Clientes)
	mux.HandleFunc("/api/clientes/", handlers.ClienteById)
	mux.HandleFunc("/api/vehiculos", handlers.Vehiculos)
	mux.HandleFunc("/api/historial", handlers.Historial)
	mux.HandleFunc("/api/historial/", handlers.Historial)
	mux.HandleFunc("/api/empleados", handlers.Empleados)
	mux.HandleFunc("/api/servicios", handlers.Servicios)
	mux.HandleFunc("/api/repuestos", handlers.Repuestos)
	mux.HandleFunc("/api/repuestos/bajo-stock", handlers.RepuestosBajoStock)
	mux.HandleFunc("/api/ordenes", handlers.Ordenes)
	mux.HandleFunc("/api/ordenes/", handlers.OrdenDetalle)
	mux.HandleFunc("/api/revisionescda", handlers.RevisionesCDA)

	log.Println("Servidor corriendo en http://localhost:8080")

	err := http.ListenAndServe(":8080", corsMiddleware(mux))

	if err != nil && err != http.ErrServerClosed {
		log.Println("Error en el servidor:", err)
	}

	if err := mongodb.Disconnect(); err != nil {
		log.Println("Error cerrando MongoDB:", err)
	}
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
