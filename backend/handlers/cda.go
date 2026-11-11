package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

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
			 JOIN vehiculos v ON v.vehiculo_id = r.vehiculo_id
			 JOIN clientes c ON c.cliente_id = v.cliente_id
			 JOIN empleados e ON e.empleado_id = r.empleado_id
			 ORDER BY r.fecha_revision DESC`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		lista := []map[string]interface{}{}
		for rows.Next() {
			var rid, vid, eid int
			var fecha, resultado, placa, vehiculo, inspector, cliente string
			var obs, vencimiento *string
			if err := rows.Scan(&rid, &vid, &eid, &fecha, &resultado, &obs, &vencimiento,
				&placa, &vehiculo, &inspector, &cliente); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			lista = append(lista, map[string]interface{}{
				"revisionId": rid, "vehiculoId": vid, "empleadoId": eid,
				"fechaRevision": fecha, "resultado": resultado, "observaciones": obs,
				"fechaVencimiento": vencimiento, "placa": placa,
				"vehiculoDesc": vehiculo, "inspector": inspector, "cliente": cliente,
			})
		}
		if err := rows.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(lista)

	case http.MethodPost:
		var body struct {
			VehiculoID int    `json:"vehiculoId"`
			EmpleadoID int    `json:"empleadoId"`
			Resultado  string `json:"resultado"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, `{"error":"Solicitud JSON invalida"}`, http.StatusBadRequest)
			return
		}
		if body.VehiculoID <= 0 || body.EmpleadoID <= 0 || !resultadoCDAValido(body.Resultado) {
			http.Error(w, `{"error":"Datos de revision CDA invalidos"}`, http.StatusBadRequest)
			return
		}

		// El formulario ya no captura observaciones generales. Se conserva la columna
		// por compatibilidad con el procedimiento almacenado y se envía una cadena vacía.
		_, err := db.DB.Exec(`CALL sp_registrar_revision_cda(?,?,?,?,@rid,@msg)`,
			body.VehiculoID, body.EmpleadoID, body.Resultado, "")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		var revisionID int
		var mensaje string
		if err := db.DB.QueryRow("SELECT @rid, @msg").Scan(&revisionID, &mensaje); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if revisionID == -1 {
			http.Error(w, mensaje, http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"revisionId": revisionID, "mensaje": mensaje})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func resultadoCDAValido(resultado string) bool {
	return resultado == "APROBADO" ||
		resultado == "REPROBADO" ||
		resultado == "APROBADO_CON_OBSERVACIONES"
}

func RevisionCDAByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 3 || parts[0] != "api" || parts[1] != "revisionescda" {
		http.Error(w, `{"error":"Ruta CDA invalida"}`, http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(parts[2])
	if err != nil || id <= 0 {
		http.Error(w, `{"error":"revisionId invalido"}`, http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodPut:
		var body struct {
			VehiculoID int               `json:"vehiculoId"`
			EmpleadoID int               `json:"empleadoId"`
			Resultado  string            `json:"resultado"`
			Detalles   []detalleCDAInput `json:"detalles"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, `{"error":"Solicitud JSON invalida"}`, http.StatusBadRequest)
			return
		}
		if body.VehiculoID <= 0 || body.EmpleadoID <= 0 || !resultadoCDAValido(body.Resultado) {
			http.Error(w, `{"error":"Datos de revision CDA invalidos"}`, http.StatusBadRequest)
			return
		}
		if len(body.Detalles) == 0 {
			http.Error(w, `{"error":"Debes enviar los elementos del checklist"}`, http.StatusBadRequest)
			return
		}
		seen := make(map[int]bool, len(body.Detalles))
		for _, d := range body.Detalles {
			if d.ItemID <= 0 || seen[d.ItemID] ||
				(d.Resultado != "CONFORME" && d.Resultado != "NO_CONFORME") {
				http.Error(w, `{"error":"Hay elementos del checklist invalidos o repetidos"}`, http.StatusBadRequest)
				return
			}
			seen[d.ItemID] = true
		}

		tx, err := db.DB.Begin()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		res, err := tx.Exec(
			`UPDATE revisiones_cda
			 SET vehiculo_id=?, empleado_id=?, resultado=?,
			     fecha_vencimiento=CASE WHEN ?='REPROBADO' THEN NULL
			     ELSE DATE_ADD(fecha_revision, INTERVAL 1 YEAR) END
			 WHERE revision_id=?`,
			body.VehiculoID, body.EmpleadoID, body.Resultado, body.Resultado, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		affected, err := res.RowsAffected()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if affected == 0 {
			var exists int
			if err := tx.QueryRow(`SELECT COUNT(*) FROM revisiones_cda WHERE revision_id=?`, id).Scan(&exists); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if exists == 0 {
				http.Error(w, `{"error":"Revision no encontrada"}`, http.StatusNotFound)
				return
			}
		}

		// Reemplazo transaccional: la revisión y sus 12 resultados se guardan juntos.
		if _, err := tx.Exec(`DELETE FROM detalle_revision_cda WHERE revision_id=?`, id); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		for _, d := range body.Detalles {
			if _, err := tx.Exec(
				`INSERT INTO detalle_revision_cda(revision_id,item_id,resultado,observacion)
				 VALUES(?,?,?,?)`,
				id, d.ItemID, d.Resultado, d.Observacion); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
		if err := tx.Commit(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"mensaje": "Revision CDA y checklist actualizados correctamente"})

	case http.MethodDelete:
		tx, err := db.DB.Begin()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		if _, err = tx.Exec(`DELETE FROM detalle_revision_cda WHERE revision_id=?`, id); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		res, err := tx.Exec(`DELETE FROM revisiones_cda WHERE revision_id=?`, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		n, err := res.RowsAffected()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if n == 0 {
			http.Error(w, `{"error":"Revision no encontrada"}`, http.StatusNotFound)
			return
		}
		if err := tx.Commit(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"mensaje": "Revision CDA y su checklist eliminados correctamente"})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
