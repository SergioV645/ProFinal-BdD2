package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"taller_vehiculos/db"
)

type Empleado struct {
	EmpleadoID  int       `json:"empleadoId"`
	CargoID     int       `json:"cargoId"`
	Cedula      string    `json:"cedula"`
	Nombre      string    `json:"nombre"`
	Apellido    string    `json:"apellido"`
	Telefono    *string   `json:"telefono"`
	Salario     float64   `json:"salario"`
	FechaIngreso time.Time `json:"fechaIngreso"`
	Activo      bool      `json:"activo"`
	NombreCargo string    `json:"nombreCargo"`
}

func Empleados(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rows, err := db.DB.Query(
		`SELECT e.empleado_id, e.cargo_id, e.cedula, e.nombre, e.apellido,
		 e.telefono, e.salario, e.fecha_ingreso, e.activo, c.nombre
		 FROM empleados e JOIN cargos c ON c.cargo_id = e.cargo_id
		 WHERE e.activo = 1 ORDER BY e.apellido, e.nombre`)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	var lista []Empleado
	for rows.Next() {
		var e Empleado
		rows.Scan(&e.EmpleadoID, &e.CargoID, &e.Cedula, &e.Nombre, &e.Apellido,
			&e.Telefono, &e.Salario, &e.FechaIngreso, &e.Activo, &e.NombreCargo)
		lista = append(lista, e)
	}
	if lista == nil {
		lista = []Empleado{}
	}
	json.NewEncoder(w).Encode(lista)
}
