package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"taller_vehiculos/db"
)

func Repuestos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		rows, err := db.DB.Query(
			`SELECT r.repuesto_id, r.proveedor_id, r.referencia, r.nombre, r.descripcion,
			 r.precio_compra, r.precio_venta, r.stock_actual, r.stock_minimo, p.nombre
			 FROM repuestos r JOIN proveedores p ON p.proveedor_id = r.proveedor_id
			 ORDER BY r.nombre`)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()

		var lista []map[string]interface{}
		for rows.Next() {
			var rid, provid, stock, stockMin int
			var ref, nombre, proveedor string
			var descripcion *string
			var pcompra, pventa float64
			rows.Scan(&rid, &provid, &ref, &nombre, &descripcion, &pcompra, &pventa, &stock, &stockMin, &proveedor)
			lista = append(lista, map[string]interface{}{
				"repuestoId": rid, "proveedorId": provid, "referencia": ref,
				"nombre": nombre, "descripcion": descripcion,
				"precioCompra": pcompra, "precioVenta": pventa,
				"stockActual": stock, "stockMinimo": stockMin, "nombreProveedor": proveedor,
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
			`INSERT INTO repuestos (proveedor_id,referencia,nombre,descripcion,precio_compra,precio_venta,stock_actual,stock_minimo)
			 VALUES (?,?,?,?,?,?,?,?)`,
			body["proveedorId"], body["referencia"], body["nombre"], body["descripcion"],
			body["precioCompra"], body["precioVenta"], body["stockActual"], body["stockMinimo"])
		if err != nil {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		id, _ := res.LastInsertId()
		json.NewEncoder(w).Encode(map[string]interface{}{"repuestoId": id})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func RepuestosBajoStock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// Evitar que /api/repuestos/bajo-stock sea capturado por el handler de /api/repuestos
	if !strings.HasSuffix(r.URL.Path, "bajo-stock") {
		Repuestos(w, r)
		return
	}
	rows, err := db.DB.Query("SELECT * FROM v_repuestos_bajo_stock")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	var lista []map[string]interface{}
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		rows.Scan(ptrs...)
		row := map[string]interface{}{}
		for i, col := range cols {
			b, ok := vals[i].([]byte)
			if ok {
				row[col] = string(b)
			} else {
				row[col] = vals[i]
			}
		}
		lista = append(lista, row)
	}
	if lista == nil {
		lista = []map[string]interface{}{}
	}
	json.NewEncoder(w).Encode(lista)
}
