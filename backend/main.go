package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"taller_vehiculos/internal/auth"
	"taller_vehiculos/internal/config"
	"taller_vehiculos/internal/db"
	"taller_vehiculos/internal/handlers"
	"taller_vehiculos/internal/mongodb"
)

func main() {
	// Carga backend/.env o el .env de la raíz del proyecto, si existe.
	envFile, err := config.LoadDotEnv(".env", "../.env")
	if err != nil {
		log.Fatal("Error leyendo el archivo .env: ", err)
	}
	if envFile != "" {
		log.Println("Configuración cargada desde", envFile)
	}

	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		log.Fatal("Define MYSQL_DSN antes de iniciar el servidor")
	}
	db.Connect(dsn)

	if err := auth.Configure(os.Getenv("SESSION_SECRET")); err != nil {
		log.Fatal(err)
	}
	if err := auth.EnsureBootstrapAdmin(); err != nil {
		log.Fatal("Configuracion inicial de autenticacion: ", err)
	}

	if err := mongodb.Connect(); err != nil {
		log.Fatal("Error conectando a MongoDB: ", err)
	}
	log.Println("Conectado a MongoDB Atlas correctamente")

	port := getEnv("PORT", "8080")
	frontendDir := getEnv("FRONTEND_DIR", "../frontend")

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(frontendDir)))

	// Autenticacion y administracion de usuarios.
	mux.HandleFunc("/api/auth/login", auth.Login)
	mux.HandleFunc("/api/auth/logout", auth.Logout)
	mux.HandleFunc("/api/auth/me", auth.Me)
	mux.HandleFunc("/api/usuarios", auth.Users)
	mux.HandleFunc("/api/usuarios/", auth.UserByID)

	// API operativa.
	mux.HandleFunc("/api/clientes", handlers.Clientes)
	mux.HandleFunc("/api/clientes/", handlers.ClienteById)
	mux.HandleFunc("/api/vehiculos", handlers.Vehiculos)
	mux.HandleFunc("/api/historial", handlers.Historial)
	mux.HandleFunc("/api/historial/", handlers.Historial)
	mux.HandleFunc("/api/empleados", handlers.Empleados)
	mux.HandleFunc("/api/servicios", handlers.Servicios)
	mux.HandleFunc("/api/repuestos", handlers.Repuestos)
	mux.HandleFunc("/api/repuestos/bajo-stock", handlers.RepuestosBajoStock)
	mux.HandleFunc("/api/repuestos/", handlers.RepuestoProveedores)
	mux.HandleFunc("/api/proveedores", handlers.Proveedores)
	mux.HandleFunc("/api/ordenes", handlers.Ordenes)
	mux.HandleFunc("/api/ordenes/", handlers.OrdenDetalle)
	mux.HandleFunc("/api/revisionescda", handlers.RevisionesCDA)
	mux.HandleFunc("/api/revisionescda/", func(w http.ResponseWriter, r *http.Request) {
		suffix := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/revisionescda/"), "/")
		parts := strings.Split(suffix, "/")
		if suffix != "" && len(parts) == 1 {
			handlers.RevisionCDAByID(w, r)
			return
		}
		if len(parts) == 2 && parts[0] != "" && parts[1] == "detalle" {
			handlers.DetalleRevisionCDA(w, r)
			return
		}
		http.Error(w, `{"error":"Ruta CDA no encontrada"}`, http.StatusNotFound)
	})
	mux.HandleFunc("/api/items-inspeccion", handlers.ItemsInspeccion)

	log.Printf("Servidor corriendo en http://localhost:%s", port)
	err = http.ListenAndServe(":"+port, corsMiddleware(auth.Middleware(mux)))
	if err != nil && err != http.ErrServerClosed {
		log.Println("Error en el servidor:", err)
	}
	if err := mongodb.Disconnect(); err != nil {
		log.Println("Error cerrando MongoDB:", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8080")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
