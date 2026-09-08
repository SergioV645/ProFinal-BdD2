package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"taller_vehiculos/db"
)

func Ordenes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		estado := r.URL.Query().Get("estado")
		query := `SELECT orden_id, fecha_ingreso, estado, cliente, placa,
			vehiculo, responsable, total_facturado, estado_pago, factura_id
			FROM v_ordenes_detalladas`
		args := []interface{}{}
		if estado != "" {
			query += " WHERE estado = ?"
			args = append(args, estado)
		}
		query += " ORDER BY fecha_ingreso DESC"

		rows, err := db.DB.Query(query, args...)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()

		var lista []map[string]interface{}
		for rows.Next() {
			var oid int
			var fechaIngreso, estadoO, cliente, placa, vehiculo, responsable string
			var totalFact *float64
			var estadoPago *string
			var facturaId *int
			rows.Scan(&oid, &fechaIngreso, &estadoO, &cliente, &placa, &vehiculo,
				&responsable, &totalFact, &estadoPago, &facturaId)
			lista = append(lista, map[string]interface{}{
				"ordenId": oid, "fechaIngreso": fechaIngreso, "estado": estadoO,
				"cliente": cliente, "placa": placa, "vehiculoDesc": vehiculo,
				"responsable": responsable, "totalFacturado": totalFact,
				"estadoPago": estadoPago, "facturaId": facturaId,
			})
		}
		if lista == nil {
			lista = []map[string]interface{}{}
		}
		json.NewEncoder(w).Encode(lista)

	case http.MethodPost:
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		_, err := db.DB.Exec(`CALL sp_registrar_orden_trabajo(?,?,?,?,?,@oid,@msg)`,
			body["vehiculoId"], body["empleadoId"], body["fechaEstimada"],
			body["kmIngreso"], body["diagnostico"])
		if err != nil {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		var ordenId int
		var mensaje string
		db.DB.QueryRow("SELECT @oid, @msg").Scan(&ordenId, &mensaje)

		if ordenId == -1 {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(map[string]string{"error": mensaje})
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"ordenId": ordenId, "mensaje": mensaje})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func OrdenDetalle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parsear la URL: /api/ordenes/{id} o /api/ordenes/{id}/servicios etc.
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/ordenes/"), "/")
	id := parts[0]
	sub := ""
	if len(parts) > 1 {
		sub = parts[1]
	}

	switch {
	case sub == "servicios" && r.Method == http.MethodPost:
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		_, err := db.DB.Exec(`CALL sp_agregar_servicio_orden(?,?,?,?,@did,@msg)`,
			id, body["servicioId"], body["cantidad"], body["manoObra"])
		if err != nil {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		var detalleId int
		var mensaje string
		db.DB.QueryRow("SELECT @did, @msg").Scan(&detalleId, &mensaje)
		if detalleId == -1 {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(map[string]string{"error": mensaje})
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"detalleId": detalleId, "mensaje": mensaje})

	case sub == "repuestos" && r.Method == http.MethodPost:
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		_, err := db.DB.Exec(`CALL sp_agregar_repuesto_orden(?,?,?,@did,@msg)`,
			id, body["repuestoId"], body["cantidad"])
		if err != nil {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		var detalleId int
		var mensaje string
		db.DB.QueryRow("SELECT @did, @msg").Scan(&detalleId, &mensaje)
		if detalleId == -1 {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(map[string]string{"error": mensaje})
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"detalleId": detalleId, "mensaje": mensaje})

	case sub == "cerrar" && r.Method == http.MethodPost:
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		_, err := db.DB.Exec(`CALL sp_cerrar_orden_y_facturar(?,?,?,@fid,@total,@msg)`,
			id, body["observaciones"], body["metodoPago"])
		if err != nil {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		var facturaId int
		var total float64
		var mensaje string
		db.DB.QueryRow("SELECT @fid, @total, @msg").Scan(&facturaId, &total, &mensaje)
		if facturaId == -1 {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(map[string]string{"error": mensaje})
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"facturaId": facturaId, "total": total, "mensaje": mensaje})

	case sub == "" && r.Method == http.MethodGet:
		orden := map[string]interface{}{}
		row := db.DB.QueryRow(`SELECT orden_id, fecha_ingreso, estado, cliente, placa,
			vehiculo, responsable, total_facturado, estado_pago, factura_id, diagnostico, km_ingreso
			FROM v_ordenes_detalladas WHERE orden_id = ?`, id)

		var oid, km int
		var fechaIngreso, estadoO, cliente, placa, vehiculo, responsable string
		var diagnostico *string
		var totalFact *float64
		var estadoPago *string
		var facturaId *int
		if err := row.Scan(&oid, &fechaIngreso, &estadoO, &cliente, &placa, &vehiculo,
			&responsable, &totalFact, &estadoPago, &facturaId, &diagnostico, &km); err != nil {
			w.WriteHeader(404)
			return
		}
		orden = map[string]interface{}{
			"ordenId": oid, "fechaIngreso": fechaIngreso, "estado": estadoO,
			"cliente": cliente, "placa": placa, "vehiculoDesc": vehiculo,
			"responsable": responsable, "totalFacturado": totalFact,
			"estadoPago": estadoPago, "facturaId": facturaId,
			"diagnostico": diagnostico, "kmIngreso": km,
		}

		// Servicios de la orden
		sRows, _ := db.DB.Query(
			`SELECT ds.detalle_id, s.nombre, ds.cantidad, ds.precio_unitario, ds.mano_obra, ds.subtotal
			 FROM detalle_servicios ds JOIN servicios s ON s.servicio_id = ds.servicio_id
			 WHERE ds.orden_id = ?`, id)
		defer sRows.Close()
		var servicios []map[string]interface{}
		for sRows.Next() {
			var did, cant int
			var nombre string
			var pu, mo, sub float64
			sRows.Scan(&did, &nombre, &cant, &pu, &mo, &sub)
			servicios = append(servicios, map[string]interface{}{
				"detalleId": did, "servicio": nombre, "cantidad": cant,
				"precioUnitario": pu, "manoObra": mo, "subtotal": sub,
			})
		}

		// Repuestos de la orden
		rRows, _ := db.DB.Query(
			`SELECT dr.detalle_id, r.nombre, r.referencia, dr.cantidad, dr.precio_unitario, dr.subtotal
			 FROM detalle_repuestos dr JOIN repuestos r ON r.repuesto_id = dr.repuesto_id
			 WHERE dr.orden_id = ?`, id)
		defer rRows.Close()
		var repuestos []map[string]interface{}
		for rRows.Next() {
			var did, cant int
			var nombre, ref string
			var pu, sub float64
			rRows.Scan(&did, &nombre, &ref, &cant, &pu, &sub)
			repuestos = append(repuestos, map[string]interface{}{
				"detalleId": did, "repuesto": nombre, "referencia": ref,
				"cantidad": cant, "precioUnitario": pu, "subtotal": sub,
			})
		}

		if servicios == nil {
			servicios = []map[string]interface{}{}
		}
		if repuestos == nil {
			repuestos = []map[string]interface{}{}
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"orden": orden, "servicios": servicios, "repuestos": repuestos,
		})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
