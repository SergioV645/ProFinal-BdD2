package handlers

import (
	"encoding/json"
	"net/http"

	"taller_vehiculos/db"
)

func Vehiculos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		placa := r.URL.Query().Get("placa")
		query := `SELECT v.vehiculo_id, v.cliente_id, v.placa, v.marca, v.modelo,
			v.anio, v.color, v.cilindraje, v.km_actual,
			CONCAT(c.nombre,' ',c.apellido) AS nombre_cliente, c.telefono
			FROM vehiculos v JOIN clientes c ON c.cliente_id = v.cliente_id`
		args := []interface{}{}
		if placa != "" {
			query += " WHERE v.placa LIKE ?"
			args = append(args, "%"+placa+"%")
		}
		query += " ORDER BY v.placa"

		rows, err := db.DB.Query(query, args...)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()

		var lista []map[string]interface{}
		for rows.Next() {
			var vid, cid, anio, km int
			var placa, marca, modelo string
			var color, nombreCliente, telefono *string
			var cil *float64
			rows.Scan(&vid, &cid, &placa, &marca, &modelo, &anio, &color, &cil, &km, &nombreCliente, &telefono)
			lista = append(lista, map[string]interface{}{
				"vehiculoId": vid, "clienteId": cid, "placa": placa,
				"marca": marca, "modelo": modelo, "anio": anio,
				"color": color, "cilindraje": cil, "kmActual": km,
				"nombreCliente": nombreCliente, "telefonoCliente": telefono,
			})
		}
		if lista == nil {
			lista = []map[string]interface{}{}
		}
		json.NewEncoder(w).Encode(lista)

	case http.MethodPost:
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		res, err := db.DB.Exec(
			`INSERT INTO vehiculos (cliente_id,placa,marca,modelo,anio,color,cilindraje,km_actual)
			 VALUES (?,?,?,?,?,?,?,?)`,
			body["clienteId"], body["placa"], body["marca"], body["modelo"],
			body["anio"], body["color"], body["cilindraje"], body["kmActual"])
		if err != nil {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		id, _ := res.LastInsertId()
		json.NewEncoder(w).Encode(map[string]interface{}{"vehiculoId": id})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
