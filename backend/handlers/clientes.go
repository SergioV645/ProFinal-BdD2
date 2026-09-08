package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"taller_vehiculos/db"
)

type Cliente struct {
	ClienteID     int       `json:"clienteId"`
	Cedula        string    `json:"cedula"`
	Nombre        string    `json:"nombre"`
	Apellido      string    `json:"apellido"`
	Telefono      *string   `json:"telefono"`
	Email         *string   `json:"email"`
	Direccion     *string   `json:"direccion"`
	FechaRegistro time.Time `json:"fechaRegistro"`
}

func Clientes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		rows, err := db.DB.Query(
			`SELECT cliente_id, cedula, nombre, apellido, telefono, email, direccion, fecha_registro
			 FROM clientes ORDER BY apellido, nombre`)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()

		var lista []Cliente
		for rows.Next() {
			var c Cliente
			if err := rows.Scan(&c.ClienteID, &c.Cedula, &c.Nombre, &c.Apellido,
				&c.Telefono, &c.Email, &c.Direccion, &c.FechaRegistro); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			lista = append(lista, c)
		}
		if lista == nil {
			lista = []Cliente{}
		}
		json.NewEncoder(w).Encode(lista)

	case http.MethodPost:
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		placa, _ := body["placa"].(string)
		marca, _ := body["marca"].(string)

		if placa != "" && marca != "" {
			// Usar stored procedure sp_registrar_cliente_vehiculo
			var clienteID, vehiculoID int
			var mensaje string

			anio := 2020
			if a, ok := body["anio"].(float64); ok {
				anio = int(a)
			}
			km := 0
			if k, ok := body["kmActual"].(float64); ok {
				km = int(k)
			}

			_, err := db.DB.Exec(`CALL sp_registrar_cliente_vehiculo(?,?,?,?,?,?,?,?,?,?,?,?,?,@cid,@vid,@msg)`,
				body["cedula"], body["nombre"], body["apellido"],
				body["telefono"], body["email"], body["direccion"],
				placa, marca, body["modelo"], anio,
				body["color"], body["cilindraje"], km)

			if err != nil {
				json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}

			db.DB.QueryRow("SELECT @cid, @vid, @msg").Scan(&clienteID, &vehiculoID, &mensaje)

			if clienteID == -1 {
				w.WriteHeader(400)
				json.NewEncoder(w).Encode(map[string]string{"error": mensaje})
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{
				"clienteId": clienteID, "vehiculoId": vehiculoID, "mensaje": mensaje,
			})
		} else {
			var id int64
			err := db.DB.QueryRow(
				`INSERT INTO clientes (cedula,nombre,apellido,telefono,email,direccion)
				 VALUES (?,?,?,?,?,?); SELECT LAST_INSERT_ID()`,
				body["cedula"], body["nombre"], body["apellido"],
				body["telefono"], body["email"], body["direccion"],
			).Scan(&id)
			if err != nil {
				// fallback: exec + last insert id
				res, err2 := db.DB.Exec(
					`INSERT INTO clientes (cedula,nombre,apellido,telefono,email,direccion) VALUES (?,?,?,?,?,?)`,
					body["cedula"], body["nombre"], body["apellido"],
					body["telefono"], body["email"], body["direccion"])
				if err2 != nil {
					w.WriteHeader(400)
					json.NewEncoder(w).Encode(map[string]string{"error": err2.Error()})
					return
				}
				id, _ = res.LastInsertId()
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"clienteId": id})
		}

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func ClienteById(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// Extraer ID de la URL: /api/clientes/5/vehiculos o /api/clientes/5
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/clientes/"), "/")
	id := parts[0]

	if len(parts) > 1 && parts[1] == "vehiculos" {
		rows, err := db.DB.Query(
			`SELECT vehiculo_id, cliente_id, placa, marca, modelo, anio, color, cilindraje, km_actual
			 FROM vehiculos WHERE cliente_id = ?`, id)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()
		var lista []map[string]interface{}
		for rows.Next() {
			var vid, cid, anio, km int
			var placa, marca, modelo string
			var color *string
			var cil *float64
			rows.Scan(&vid, &cid, &placa, &marca, &modelo, &anio, &color, &cil, &km)
			lista = append(lista, map[string]interface{}{
				"vehiculoId": vid, "clienteId": cid, "placa": placa,
				"marca": marca, "modelo": modelo, "anio": anio,
				"color": color, "cilindraje": cil, "kmActual": km,
			})
		}
		if lista == nil {
			lista = []map[string]interface{}{}
		}
		json.NewEncoder(w).Encode(lista)
		return
	}

	var c Cliente
	err := db.DB.QueryRow(
		`SELECT cliente_id, cedula, nombre, apellido, telefono, email, direccion, fecha_registro
		 FROM clientes WHERE cliente_id = ?`, id).
		Scan(&c.ClienteID, &c.Cedula, &c.Nombre, &c.Apellido,
			&c.Telefono, &c.Email, &c.Direccion, &c.FechaRegistro)
	if err == sql.ErrNoRows {
		w.WriteHeader(404)
		return
	}
	json.NewEncoder(w).Encode(c)
}
