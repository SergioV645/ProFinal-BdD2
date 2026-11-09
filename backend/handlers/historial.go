package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"taller_vehiculos/db"
	"taller_vehiculos/mongodb"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type TecnicoHistorial struct {
	EmpleadoID   int    `json:"empleadoId,omitempty" bson:"empleado_id,omitempty"`
	Nombre       string `json:"nombre" bson:"nombre"`
	Especialidad string `json:"especialidad,omitempty" bson:"especialidad,omitempty"`
}

type HistorialVehicular struct {
	ID              bson.ObjectID    `json:"id" bson:"_id,omitempty"`
	VehiculoID      int              `json:"vehiculoId" bson:"vehiculo_id"`
	Placa           string           `json:"placa,omitempty" bson:"placa,omitempty"`
	OrdenID         *int             `json:"ordenId,omitempty" bson:"orden_id,omitempty"`
	RevisionCDAID   *int             `json:"revisionCdaId,omitempty" bson:"revision_cda_id,omitempty"`
	TipoEvento      string           `json:"tipoEvento" bson:"tipo_evento"`
	FechaEvento     time.Time        `json:"fechaEvento" bson:"fecha_evento"`
	Tecnico         TecnicoHistorial `json:"tecnico" bson:"tecnico"`
	Detalle         map[string]any   `json:"detalle" bson:"detalle"`
	Observaciones   string           `json:"observaciones,omitempty" bson:"observaciones,omitempty"`
	Recomendaciones string           `json:"recomendaciones,omitempty" bson:"recomendaciones,omitempty"`
	CreadoEn        time.Time        `json:"creadoEn" bson:"creado_en"`
	ActualizadoEn   time.Time        `json:"actualizadoEn" bson:"actualizado_en"`
}

func Historial(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if mongodb.HistorialCollection == nil {
		http.Error(w, `{"error":"MongoDB no está conectado"}`, http.StatusServiceUnavailable)
		return
	}

	path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/historial"), "/")
	parts := []string{}
	if path != "" {
		parts = strings.Split(path, "/")
	}

	if len(parts) == 2 && parts[0] == "vehiculo" {
		if r.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}
		id, err := strconv.Atoi(parts[1])
		if err != nil || id <= 0 {
			writeJSONError(w, http.StatusBadRequest, "vehiculoId inválido")
			return
		}
		listarHistorialVehiculo(w, r, id)
		return
	}
	if len(parts) > 1 {
		http.NotFound(w, r)
		return
	}
	if len(parts) == 1 {
		if parts[0] == "estadisticas" {
			if r.Method != http.MethodGet {
				methodNotAllowed(w)
				return
			}
			obtenerEstadisticasHistorial(w, r)
			return
		}

		id, err := bson.ObjectIDFromHex(parts[0])
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "Identificador de historial inválido")
			return
		}

		switch r.Method {
		case http.MethodGet:
			obtenerHistorial(w, r, id)
		case http.MethodPut:
			actualizarHistorial(w, r, id)
		case http.MethodDelete:
			eliminarHistorial(w, r, id)
		default:
			methodNotAllowed(w)
		}
		return
	}

	switch r.Method {
	case http.MethodGet:
		listarHistorial(w, r)
	case http.MethodPost:
		crearHistorial(w, r)
	default:
		methodNotAllowed(w)
	}
}

func crearHistorial(w http.ResponseWriter, r *http.Request) {
	var item HistorialVehicular
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		writeJSONError(w, 400, "JSON inválido")
		return
	}
	if err := validarHistorial(&item); err != nil {
		writeJSONError(w, 400, err.Error())
		return
	}
	var placa string
	err := db.DB.QueryRow("SELECT placa FROM vehiculos WHERE vehiculo_id = ?", item.VehiculoID).Scan(&placa)
	if errors.Is(err, sql.ErrNoRows) {
		writeJSONError(w, 404, "Vehículo no encontrado")
		return
	}
	if err != nil {
		writeJSONError(w, 500, "No fue posible validar el vehículo")
		return
	}
	item.ID = bson.NewObjectID()
	item.Placa = placa
	item.CreadoEn = time.Now().UTC()
	item.ActualizadoEn = item.CreadoEn
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	if _, err := mongodb.HistorialCollection.InsertOne(ctx, item); err != nil {
		writeJSONError(w, 500, "No fue posible guardar el historial")
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

func listarHistorial(w http.ResponseWriter, r *http.Request) {
	filter := bson.M{}
	if raw := r.URL.Query().Get("vehiculoId"); raw != "" {
		id, err := strconv.Atoi(raw)
		if err != nil || id <= 0 {
			writeJSONError(w, 400, "vehiculoId inválido")
			return
		}
		filter["vehiculo_id"] = id
	}
	if tipo := strings.TrimSpace(r.URL.Query().Get("tipoEvento")); tipo != "" {
		filter["tipo_evento"] = strings.ToUpper(tipo)
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	findOpts := options.Find().SetSort(bson.D{{Key: "fecha_evento", Value: -1}})
	cursor, err := mongodb.HistorialCollection.Find(ctx, filter, findOpts)
	if err != nil {
		writeJSONError(w, 500, "No fue posible consultar el historial")
		return
	}
	defer cursor.Close(ctx)
	items := []HistorialVehicular{}
	if err := cursor.All(ctx, &items); err != nil {
		writeJSONError(w, 500, "No fue posible leer los registros")
		return
	}
	json.NewEncoder(w).Encode(items)
}

func listarHistorialVehiculo(w http.ResponseWriter, r *http.Request, id int) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	cursor, err := mongodb.HistorialCollection.Find(ctx, bson.M{"vehiculo_id": id}, options.Find().SetSort(bson.D{{Key: "fecha_evento", Value: -1}}))
	if err != nil {
		writeJSONError(w, 500, "No fue posible consultar el historial")
		return
	}
	defer cursor.Close(ctx)
	items := []HistorialVehicular{}
	if err := cursor.All(ctx, &items); err != nil {
		writeJSONError(w, 500, "No fue posible leer los registros")
		return
	}
	json.NewEncoder(w).Encode(items)
}

func obtenerHistorial(w http.ResponseWriter, r *http.Request, id bson.ObjectID) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	var item HistorialVehicular
	if err := mongodb.HistorialCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&item); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			writeJSONError(w, 404, "Registro no encontrado")
		} else {
			writeJSONError(w, 500, "No fue posible consultar el registro")
		}
		return
	}
	json.NewEncoder(w).Encode(item)
}

func actualizarHistorial(w http.ResponseWriter, r *http.Request, id bson.ObjectID) {
	var item HistorialVehicular
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		writeJSONError(w, 400, "JSON inválido")
		return
	}
	if err := validarHistorial(&item); err != nil {
		writeJSONError(w, 400, err.Error())
		return
	}
	var placa string
	if err := db.DB.QueryRow("SELECT placa FROM vehiculos WHERE vehiculo_id = ?", item.VehiculoID).Scan(&placa); err != nil {
		writeJSONError(w, 404, "Vehículo no encontrado")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	item.ID = id
	item.Placa = placa
	item.ActualizadoEn = time.Now().UTC()
	result, err := mongodb.HistorialCollection.ReplaceOne(ctx, bson.M{"_id": id}, item)
	if err != nil {
		writeJSONError(w, 500, "No fue posible actualizar el registro")
		return
	}
	if result.MatchedCount == 0 {
		writeJSONError(w, 404, "Registro no encontrado")
		return
	}
	json.NewEncoder(w).Encode(item)
}

func eliminarHistorial(w http.ResponseWriter, r *http.Request, id bson.ObjectID) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	result, err := mongodb.HistorialCollection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		writeJSONError(w, 500, "No fue posible eliminar el registro")
		return
	}
	if result.DeletedCount == 0 {
		writeJSONError(w, 404, "Registro no encontrado")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func validarHistorial(item *HistorialVehicular) error {
	if item.VehiculoID <= 0 {
		return errors.New("vehiculoId es obligatorio y debe ser mayor que cero")
	}
	item.TipoEvento = strings.ToUpper(strings.TrimSpace(item.TipoEvento))
	switch item.TipoEvento {
	case "DIAGNOSTICO", "MANTENIMIENTO", "REPARACION", "INSPECCION_CDA":
	default:
		return errors.New("tipoEvento debe ser DIAGNOSTICO, MANTENIMIENTO, REPARACION o INSPECCION_CDA")
	}
	if item.FechaEvento.IsZero() {
		return errors.New("fechaEvento es obligatoria (formato AAAA-MM-DD)")
	}
	if strings.TrimSpace(item.Tecnico.Nombre) == "" {
		return errors.New("el nombre del técnico es obligatorio")
	}
	if item.Detalle == nil {
		item.Detalle = map[string]any{}
	}
	return nil
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
func methodNotAllowed(w http.ResponseWriter) {
	writeJSONError(w, http.StatusMethodNotAllowed, "Método no permitido")
}

type EventoPorTipo struct {
	TipoEvento string `json:"tipoEvento" bson:"_id"`
	Cantidad   int64  `json:"cantidad" bson:"cantidad"`
}

type EstadisticasHistorial struct {
	TotalEventos       int64           `json:"totalEventos"`
	VehiculosAtendidos int64           `json:"vehiculosAtendidos"`
	PorTipo            []EventoPorTipo `json:"porTipo"`
}

func obtenerEstadisticasHistorial(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	coleccion := mongodb.HistorialCollection

	total, err := coleccion.CountDocuments(ctx, bson.M{})
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError,
			"No fue posible contar los eventos")
		return
	}

	// Contar vehículos diferentes que tienen historial mediante $group.
	pipelineVehiculos := bson.A{
		bson.D{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$vehiculo_id"},
			}},
		},
	}

	cursorVehiculos, err := coleccion.Aggregate(ctx, pipelineVehiculos)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError,
			"No fue posible consultar los vehículos atendidos")
		return
	}

	var vehiculos []bson.M
	if err := cursorVehiculos.All(ctx, &vehiculos); err != nil {
		_ = cursorVehiculos.Close(ctx)
		writeJSONError(w, http.StatusInternalServerError,
			"No fue posible leer los vehículos atendidos")
		return
	}
	if err := cursorVehiculos.Close(ctx); err != nil {
		writeJSONError(w, http.StatusInternalServerError,
			"No fue posible cerrar la consulta de vehículos")
		return
	}

	pipeline := bson.A{
		bson.D{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$tipo_evento"},
				{Key: "cantidad", Value: bson.D{
					{Key: "$sum", Value: 1},
				}},
			}},
		},
		bson.D{
			{Key: "$sort", Value: bson.D{
				{Key: "cantidad", Value: -1},
			}},
		},
	}

	cursor, err := coleccion.Aggregate(ctx, pipeline)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError,
			"No fue posible generar las estadísticas por tipo")
		return
	}
	defer cursor.Close(ctx)

	porTipo := []EventoPorTipo{}
	if err := cursor.All(ctx, &porTipo); err != nil {
		writeJSONError(w, http.StatusInternalServerError,
			"No fue posible leer las estadísticas")
		return
	}

	resultado := EstadisticasHistorial{
		TotalEventos:       total,
		VehiculosAtendidos: int64(len(vehiculos)),
		PorTipo:            porTipo,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(resultado)
}
