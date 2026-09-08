package handlers

import (
	"encoding/json"
	"net/http"

	"taller_vehiculos/db"
)

func RevisionesCDA(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		rows, err := db.DB.Query(
			`SELECT r.revision_id, r.vehiculo_id, r.empleado_id, r.fecha_revision,
			 r.resultado, r.observaciones, r.fecha_vencimiento,
			 v.placa, CONCAT(v.marca,' ',v.modelo) AS vehiculo_desc,
			 CONCAT(e.nombre,' ',e.apellido) AS inspector,
			 CONCAT(c.nombre,' ',c.apellido) AS cliente
			 FROM revisiones_cda r
			 JOIN vehiculos  v ON v.vehiculo_id = r.vehiculo_id
			 JOIN clientes   c ON c.cliente_id  = v.cliente_id
			 JOIN empleados  e ON e.empleado_id = r.empleado_id
			 ORDER BY r.fecha_revision DESC`)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()

		var lista []map[string]interface{}
		for rows.Next() {
			var rid, vid, eid int
			var fecha, resultado, placa, vehiculo, inspector, cliente string
			var obs, vencimiento *string
			rows.Scan(&rid, &vid, &eid, &fecha, &resultado, &obs, &vencimiento,
				&placa, &vehiculo, &inspector, &cliente)
			lista = append(lista, map[string]interface{}{
				"revisionId": rid, "vehiculoId": vid, "empleadoId": eid,
				"fechaRevision": fecha, "resultado": resultado, "observaciones": obs,
				"fechaVencimiento": vencimiento, "placa": placa,
				"vehiculoDesc": vehiculo, "inspector": inspector, "cliente": cliente,
			})
		}
		if lista == nil {
			lista = []map[string]interface{}{}
		}
		json.NewEncoder(w).Encode(lista)

	case http.MethodPost:
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		_, err := db.DB.Exec(`CALL sp_registrar_revision_cda(?,?,?,?,@rid,@msg)`,
			body["vehiculoId"], body["empleadoId"], body["resultado"], body["observaciones"])
		if err != nil {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		var revisionId int
		var mensaje string
		db.DB.QueryRow("SELECT @rid, @msg").Scan(&revisionId, &mensaje)

		if revisionId == -1 {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(map[string]string{"error": mensaje})
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"revisionId": revisionId, "mensaje": mensaje})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
