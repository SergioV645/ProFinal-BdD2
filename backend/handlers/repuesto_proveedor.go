package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"taller_vehiculos/db"
)

type proveedorRepuestoInput struct {
	ProveedorID  int     `json:"proveedorId"`
	PrecioCompra float64 `json:"precioCompra"`
	DiasEntrega  int     `json:"diasEntrega"`
}

func RepuestoProveedores(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 4 {
		http.Error(w, `{"error":"Ruta invalida"}`, http.StatusBadRequest)
		return
	}
	repuestoID, err := strconv.Atoi(parts[2])
	if err != nil || repuestoID <= 0 {
		http.Error(w, `{"error":"ID de repuesto invalido"}`, 400)
		return
	}

	switch r.Method {
	case http.MethodGet:
		rows, err := db.DB.Query(`SELECT rp.repuesto_id, rp.proveedor_id, p.nombre, rp.precio_compra, rp.dias_entrega
			FROM repuesto_proveedor rp JOIN proveedores p ON p.proveedor_id=rp.proveedor_id
			WHERE rp.repuesto_id=? ORDER BY rp.precio_compra`, repuestoID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()
		out := []map[string]interface{}{}
		for rows.Next() {
			var rid, pid, dias int
			var nombre string
			var precio float64
			if err := rows.Scan(&rid, &pid, &nombre, &precio, &dias); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			out = append(out, map[string]interface{}{"repuestoId": rid, "proveedorId": pid, "proveedor": nombre, "precioCompra": precio, "diasEntrega": dias})
		}
		json.NewEncoder(w).Encode(out)
	case http.MethodPost:
		var in proveedorRepuestoInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		if in.ProveedorID <= 0 || in.PrecioCompra <= 0 || in.DiasEntrega < 0 {
			http.Error(w, `{"error":"Proveedor, precio y dias de entrega invalidos"}`, 400)
			return
		}
		_, err := db.DB.Exec(`INSERT INTO repuesto_proveedor(repuesto_id,proveedor_id,precio_compra,dias_entrega) VALUES(?,?,?,?) ON DUPLICATE KEY UPDATE precio_compra=VALUES(precio_compra), dias_entrega=VALUES(dias_entrega)`, repuestoID, in.ProveedorID, in.PrecioCompra, in.DiasEntrega)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"mensaje": "Proveedor asociado correctamente"})
	case http.MethodDelete:
		pid, err := strconv.Atoi(r.URL.Query().Get("proveedorId"))
		if err != nil {
			http.Error(w, `{"error":"proveedorId requerido"}`, 400)
			return
		}
		_, err = db.DB.Exec(`DELETE FROM repuesto_proveedor WHERE repuesto_id=? AND proveedor_id=?`, repuestoID, pid)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"mensaje": "Proveedor eliminado"})
	default:
		http.Error(w, "Metodo no permitido", 405)
	}
}

func Proveedores(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		http.Error(w, "Metodo no permitido", 405)
		return
	}
	rows, err := db.DB.Query(`SELECT proveedor_id,nombre FROM proveedores ORDER BY nombre`)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()
	out := []map[string]interface{}{}
	for rows.Next() {
		var id int
		var nombre string
		if err := rows.Scan(&id, &nombre); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		out = append(out, map[string]interface{}{"proveedorId": id, "nombre": nombre})
	}
	json.NewEncoder(w).Encode(out)
}
