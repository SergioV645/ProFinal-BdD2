package handlers

import (
	"encoding/json"
	"net/http"

	"taller_vehiculos/db"
)

func Servicios(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rows, err := db.DB.Query(
		`SELECT s.servicio_id, s.categoria_id, s.nombre, s.descripcion, s.precio_base, cs.nombre
		 FROM servicios s JOIN categorias_servicio cs ON cs.categoria_id = s.categoria_id
		 ORDER BY cs.nombre, s.nombre`)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	var lista []map[string]interface{}
	for rows.Next() {
		var sid, catid int
		var nombre, descripcion, categoria string
		var precio float64
		rows.Scan(&sid, &catid, &nombre, &descripcion, &precio, &categoria)
		lista = append(lista, map[string]interface{}{
			"servicioId": sid, "categoriaId": catid, "nombre": nombre,
			"descripcion": descripcion, "precioBase": precio, "nombreCategoria": categoria,
		})
	}
	if lista == nil {
		lista = []map[string]interface{}{}
	}
	json.NewEncoder(w).Encode(lista)
}
