package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"taller_vehiculos/db"
)

type detalleCDAInput struct {
	ItemID      int    `json:"itemId"`
	Resultado   string `json:"resultado"`
	Observacion string `json:"observacion"`
}

func ItemsInspeccion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.DB.Query(`SELECT item_id,nombre,grupo,es_critico
		FROM items_inspeccion ORDER BY grupo,nombre`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := []map[string]interface{}{}
	for rows.Next() {
		var id int
		var nombre, grupo string
		var critico bool
		if err := rows.Scan(&id, &nombre, &grupo, &critico); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out = append(out, map[string]interface{}{
			"itemId": id, "nombre": nombre, "grupo": grupo, "esCritico": critico,
		})
	}
	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(out)
}

func DetalleRevisionCDA(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 4 || parts[0] != "api" || parts[1] != "revisionescda" || parts[3] != "detalle" {
		http.Error(w, `{"error":"Ruta de detalle CDA invalida"}`, http.StatusBadRequest)
		return
	}
	revisionID, err := strconv.Atoi(parts[2])
	if err != nil || revisionID <= 0 {
		http.Error(w, `{"error":"revisionId invalido"}`, http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		rows, err := db.DB.Query(`SELECT d.item_id,i.nombre,i.grupo,i.es_critico,d.resultado,d.observacion
			FROM detalle_revision_cda d
			JOIN items_inspeccion i ON i.item_id=d.item_id
			WHERE d.revision_id=?
			ORDER BY i.grupo,i.nombre`, revisionID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		out := []map[string]interface{}{}
		for rows.Next() {
			var id int
			var nombre, grupo, resultado string
			var critico bool
			var observacion *string
			if err := rows.Scan(&id, &nombre, &grupo, &critico, &resultado, &observacion); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			out = append(out, map[string]interface{}{
				"itemId": id, "nombre": nombre, "grupo": grupo, "esCritico": critico,
				"resultado": resultado, "observacion": observacion,
			})
		}
		if err := rows.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(out)

	case http.MethodPost:
		var detalles []detalleCDAInput
		if err := json.NewDecoder(r.Body).Decode(&detalles); err != nil {
			http.Error(w, `{"error":"Solicitud de checklist invalida"}`, http.StatusBadRequest)
			return
		}
		if len(detalles) == 0 {
			http.Error(w, `{"error":"El checklist no contiene elementos"}`, http.StatusBadRequest)
			return
		}

		tx, err := db.DB.Begin()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		var existe int
		if err := tx.QueryRow(`SELECT COUNT(*) FROM revisiones_cda WHERE revision_id=?`, revisionID).Scan(&existe); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if existe == 0 {
			http.Error(w, `{"error":"Revision no encontrada"}`, http.StatusNotFound)
			return
		}

		seen := make(map[int]bool, len(detalles))
		for _, d := range detalles {
			if d.ItemID <= 0 || seen[d.ItemID] ||
				(d.Resultado != "CONFORME" && d.Resultado != "NO_CONFORME") {
				http.Error(w, `{"error":"Detalle de inspeccion invalido o repetido"}`, http.StatusBadRequest)
				return
			}
			seen[d.ItemID] = true

			_, err = tx.Exec(`INSERT INTO detalle_revision_cda
				(revision_id,item_id,resultado,observacion) VALUES(?,?,?,?)
				ON DUPLICATE KEY UPDATE resultado=VALUES(resultado), observacion=VALUES(observacion)`,
				revisionID, d.ItemID, d.Resultado, d.Observacion)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
		if err := tx.Commit(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"mensaje": "Checklist guardado correctamente"})

	default:
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}
