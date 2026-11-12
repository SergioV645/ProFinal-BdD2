package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"taller_vehiculos/db"

	"golang.org/x/crypto/bcrypt"
)

const cookieName = "aurum_session"

type User struct {
	ID         int    `json:"userId"`
	Username   string `json:"username"`
	Role       string `json:"role"`
	EmployeeID *int   `json:"employeeId,omitempty"`
	Employee   string `json:"employee,omitempty"`
}

type claims struct {
	UserID int    `json:"uid"`
	Expiry int64  `json:"exp"`
	Nonce  string `json:"nonce"`
}

type contextKey string

const userKey contextKey = "aurum-user"

var signingSecret []byte

func Configure(secret string) error {
	if len(secret) < 32 {
		return errors.New("SESSION_SECRET debe tener al menos 32 caracteres")
	}
	signingSecret = []byte(secret)
	return nil
}

func EnsureBootstrapAdmin() error {
	_, err := db.DB.Exec(`INSERT IGNORE INTO roles (nombre, descripcion) VALUES
		('ADMINISTRADOR','Acceso completo al sistema'),
		('EMPLEADO','Acceso operativo sin permisos de eliminacion ni administracion')`)
	if err != nil {
		return fmt.Errorf("no se pudieron verificar los roles: %w", err)
	}
	var count int
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM usuarios`).Scan(&count); err != nil {
		return fmt.Errorf("no se pudo consultar usuarios: %w", err)
	}
	if count > 0 {
		return nil
	}
	username := strings.TrimSpace(os.Getenv("ADMIN_USERNAME"))
	password := os.Getenv("ADMIN_PASSWORD")
	if username == "" || len(password) < 10 {
		return errors.New("no hay usuarios. Define ADMIN_USERNAME y ADMIN_PASSWORD (minimo 10 caracteres) para crear el administrador inicial")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = db.DB.Exec(`INSERT INTO usuarios (username,password_hash,rol_id,empleado_id,activo)
		SELECT ?, ?, rol_id, NULL, 1 FROM roles WHERE nombre='ADMINISTRADOR'`,
		username, string(hash))
	if err != nil {
		return fmt.Errorf("no se pudo crear el administrador inicial: %w", err)
	}
	return nil
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "Solicitud invalida")
		return
	}
	var user User
	var passwordHash string
	err := db.DB.QueryRow(`SELECT u.usuario_id,u.username,u.password_hash,ro.nombre,u.empleado_id,
		COALESCE(CONCAT(e.nombre,' ',e.apellido),'')
		FROM usuarios u JOIN roles ro ON ro.rol_id=u.rol_id
		LEFT JOIN empleados e ON e.empleado_id=u.empleado_id
		WHERE u.username=? AND u.activo=1 AND (u.empleado_id IS NULL OR e.activo=1)`, strings.TrimSpace(input.Username)).
		Scan(&user.ID, &user.Username, &passwordHash, &user.Role, &user.EmployeeID, &user.Employee)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(input.Password)) != nil {
		writeError(w, http.StatusUnauthorized, "Usuario o contraseña incorrectos")
		return
	}
	token, err := createToken(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "No se pudo iniciar la sesion")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: cookieName, Value: token, Path: "/", HttpOnly: true,
		Secure: os.Getenv("COOKIE_SECURE") == "true", SameSite: http.SameSiteStrictMode,
		Expires: time.Now().Add(8 * time.Hour), MaxAge: 8 * 60 * 60,
	})
	writeJSON(w, http.StatusOK, map[string]interface{}{"user": user})
}

func Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: cookieName, Value: "", Path: "/", HttpOnly: true,
		Secure: os.Getenv("COOKIE_SECURE") == "true", SameSite: http.SameSiteStrictMode,
		MaxAge: -1, Expires: time.Unix(0, 0),
	})
	writeJSON(w, http.StatusOK, map[string]string{"mensaje": "Sesion cerrada"})
}

func Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	user, ok := CurrentUser(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Sesion no valida")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"user": user})
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}
		if r.URL.Path == "/api/auth/login" || r.URL.Path == "/api/auth/logout" {
			next.ServeHTTP(w, r)
			return
		}
		cookie, err := r.Cookie(cookieName)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Debes iniciar sesion")
			return
		}
		userID, err := verifyToken(cookie.Value)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Sesion vencida o invalida")
			return
		}
		user, err := loadUser(userID)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "Usuario inactivo o inexistente")
			return
		}
		if !allowed(user, r) {
			writeError(w, http.StatusForbidden, "No tienes permisos para realizar esta operacion")
			return
		}
		ctx := context.WithValue(r.Context(), userKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func CurrentUser(r *http.Request) (User, bool) {
	user, ok := r.Context().Value(userKey).(User)
	return user, ok
}

func loadUser(id int) (User, error) {
	var u User
	err := db.DB.QueryRow(`SELECT u.usuario_id,u.username,ro.nombre,u.empleado_id,
		COALESCE(CONCAT(e.nombre,' ',e.apellido),'')
		FROM usuarios u JOIN roles ro ON ro.rol_id=u.rol_id
		LEFT JOIN empleados e ON e.empleado_id=u.empleado_id
		WHERE u.usuario_id=? AND u.activo=1 AND (u.empleado_id IS NULL OR e.activo=1)`, id).
		Scan(&u.ID, &u.Username, &u.Role, &u.EmployeeID, &u.Employee)
	return u, err
}

func allowed(user User, r *http.Request) bool {
	if user.Role == "ADMINISTRADOR" {
		return true
	}
	if user.Role != "EMPLEADO" {
		return false
	}
	path, method := r.URL.Path, r.Method
	if strings.HasPrefix(path, "/api/usuarios") || strings.HasPrefix(path, "/api/roles") {
		return false
	}
	if method == http.MethodGet || method == http.MethodHead {
		return true
	}
	if method == http.MethodDelete {
		return false
	}
	if method != http.MethodPost && method != http.MethodPut && method != http.MethodPatch {
		return false
	}
	// El empleado puede crear y actualizar informacion operativa.
	// Proveedores y sus asociaciones son de solo consulta para este rol.
	if path == "/api/proveedores" || strings.Contains(path, "/proveedores") {
		return false
	}
	allowedPrefixes := []string{
		"/api/clientes", "/api/vehiculos", "/api/ordenes",
		"/api/repuestos", "/api/revisionescda", "/api/historial",
	}
	for _, prefix := range allowedPrefixes {
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			return true
		}
	}
	return false
}

func Users(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	if r.Method == http.MethodGet {
		rows, err := db.DB.Query(`SELECT u.usuario_id,u.username,ro.nombre,u.empleado_id,
			COALESCE(CONCAT(e.nombre,' ',e.apellido),''),u.activo,u.creado_en
			FROM usuarios u JOIN roles ro ON ro.rol_id=u.rol_id
			LEFT JOIN empleados e ON e.empleado_id=u.empleado_id
			ORDER BY u.usuario_id`)
		if err != nil {
			writeError(w, 500, err.Error())
			return
		}
		defer rows.Close()
		out := []map[string]interface{}{}
		for rows.Next() {
			var id int
			var username, role, created string
			var employeeID sql.NullInt64
			var employee string
			var active bool
			if err := rows.Scan(&id, &username, &role, &employeeID, &employee, &active, &created); err != nil {
				writeError(w, 500, err.Error())
				return
			}
			var eid interface{}
			if employeeID.Valid {
				eid = employeeID.Int64
			}
			out = append(out, map[string]interface{}{"userId": id, "username": username, "role": role, "employeeId": eid, "employee": employee, "active": active, "createdAt": created})
		}
		writeJSON(w, 200, out)
		return
	}
	var input struct {
		Username   string `json:"username"`
		Password   string `json:"password"`
		EmployeeID int    `json:"employeeId"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&input); err != nil {
		writeError(w, 400, "Solicitud invalida")
		return
	}
	if strings.TrimSpace(input.Username) == "" || len(input.Password) < 10 || input.EmployeeID <= 0 {
		writeError(w, 400, "Usuario, contraseña de al menos 10 caracteres y empleado son obligatorios")
		return
	}
	var exists int
	if err := db.DB.QueryRow(`SELECT COUNT(*) FROM empleados WHERE empleado_id=? AND activo=1`, input.EmployeeID).Scan(&exists); err != nil || exists == 0 {
		writeError(w, 400, "El empleado seleccionado no existe o esta inactivo")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, 500, "No se pudo proteger la contraseña")
		return
	}
	res, err := db.DB.Exec(`INSERT INTO usuarios(username,password_hash,rol_id,empleado_id,activo)
		SELECT ?,?,rol_id,?,1 FROM roles WHERE nombre='EMPLEADO'`, strings.TrimSpace(input.Username), string(hash), input.EmployeeID)
	if err != nil {
		writeError(w, 400, "No se pudo crear el usuario. Verifica que el nombre de usuario y el empleado no esten asignados.")
		return
	}
	id, _ := res.LastInsertId()
	writeJSON(w, http.StatusCreated, map[string]interface{}{"userId": id, "mensaje": "Usuario empleado creado"})
}

func UserByID(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 3 {
		writeError(w, 404, "Ruta de usuario no encontrada")
		return
	}
	id, err := strconv.Atoi(parts[2])
	if err != nil || id <= 0 {
		writeError(w, 400, "ID de usuario invalido")
		return
	}
	if r.Method != http.MethodPut {
		methodNotAllowed(w)
		return
	}
	var input struct {
		Active     *bool   `json:"active"`
		Password   string  `json:"password"`
		Username   *string `json:"username"`
		EmployeeID *int    `json:"employeeId"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&input); err != nil {
		writeError(w, 400, "Solicitud invalida")
		return
	}
	if input.Active == nil && input.Password == "" && input.Username == nil && input.EmployeeID == nil {
		writeError(w, 400, "No hay cambios para aplicar")
		return
	}
	if input.Username != nil {
		username := strings.TrimSpace(*input.Username)
		if username == "" {
			writeError(w, 400, "El nombre de usuario es obligatorio")
			return
		}
		var duplicate int
		if err := db.DB.QueryRow(`SELECT COUNT(*) FROM usuarios WHERE username=? AND usuario_id<>?`, username, id).Scan(&duplicate); err != nil {
			writeError(w, 500, "No se pudo validar el nombre de usuario")
			return
		}
		if duplicate > 0 {
			writeError(w, 400, "Ese nombre de usuario ya esta en uso")
			return
		}
		if _, err := db.DB.Exec(`UPDATE usuarios SET username=? WHERE usuario_id=?`, username, id); err != nil {
			writeError(w, 400, "No se pudo actualizar el nombre de usuario")
			return
		}
	}
	if input.EmployeeID != nil {
		var role string
		if err := db.DB.QueryRow(`SELECT ro.nombre FROM usuarios u JOIN roles ro ON ro.rol_id=u.rol_id WHERE u.usuario_id=?`, id).Scan(&role); err != nil {
			writeError(w, 404, "Usuario no encontrado")
			return
		}
		if role != "EMPLEADO" {
			writeError(w, 400, "Solo se puede cambiar el empleado asociado a una cuenta de empleado")
			return
		}
		if *input.EmployeeID <= 0 {
			writeError(w, 400, "Debes seleccionar un empleado")
			return
		}
		var exists int
		if err := db.DB.QueryRow(`SELECT COUNT(*) FROM empleados WHERE empleado_id=? AND activo=1`, *input.EmployeeID).Scan(&exists); err != nil || exists == 0 {
			writeError(w, 400, "El empleado seleccionado no existe o esta inactivo")
			return
		}
		if _, err := db.DB.Exec(`UPDATE usuarios SET empleado_id=? WHERE usuario_id=?`, *input.EmployeeID, id); err != nil {
			writeError(w, 400, "No se pudo actualizar el empleado asociado. Verifica que no tenga otra cuenta.")
			return
		}
	}
	if input.Password != "" && len(input.Password) < 10 {
		writeError(w, 400, "La contraseña debe tener al menos 10 caracteres")
		return
	}
	if input.Active != nil {
		if !*input.Active {
			current, _ := CurrentUser(r)
			if current.ID == id {
				writeError(w, http.StatusBadRequest, "No puedes desactivar tu propia cuenta")
				return
			}
			var role string
			if err := db.DB.QueryRow(`SELECT ro.nombre FROM usuarios u JOIN roles ro ON ro.rol_id=u.rol_id WHERE u.usuario_id=?`, id).Scan(&role); err == nil && role == "ADMINISTRADOR" {
				var activeAdmins int
				if err := db.DB.QueryRow(`SELECT COUNT(*) FROM usuarios u JOIN roles ro ON ro.rol_id=u.rol_id WHERE ro.nombre='ADMINISTRADOR' AND u.activo=1`).Scan(&activeAdmins); err != nil {
					writeError(w, 500, err.Error())
					return
				}
				if activeAdmins <= 1 {
					writeError(w, http.StatusBadRequest, "No puedes desactivar el ultimo administrador activo")
					return
				}
			}
			_, err = db.DB.Exec(`UPDATE usuarios SET activo=0 WHERE usuario_id=?`, id)
		} else {
			_, err = db.DB.Exec(`UPDATE usuarios SET activo=1 WHERE usuario_id=?`, id)
		}
		if err != nil {
			writeError(w, 400, err.Error())
			return
		}
	}
	if input.Password != "" {
		hash, e := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if e != nil {
			writeError(w, 500, "No se pudo proteger la contraseña")
			return
		}
		if _, err = db.DB.Exec(`UPDATE usuarios SET password_hash=? WHERE usuario_id=?`, string(hash), id); err != nil {
			writeError(w, 400, err.Error())
			return
		}
	}
	writeJSON(w, 200, map[string]string{"mensaje": "Usuario actualizado"})
}

func createToken(userID int) (string, error) {
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	payload, err := json.Marshal(claims{UserID: userID, Expiry: time.Now().Add(8 * time.Hour).Unix(), Nonce: base64.RawURLEncoding.EncodeToString(nonce)})
	if err != nil {
		return "", err
	}
	body := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, signingSecret)
	mac.Write([]byte(body))
	return body + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

func verifyToken(token string) (int, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return 0, errors.New("token invalido")
	}
	mac := hmac.New(sha256.New, signingSecret)
	mac.Write([]byte(parts[0]))
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !hmac.Equal(signature, mac.Sum(nil)) {
		return 0, errors.New("firma invalida")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return 0, err
	}
	var c claims
	if err = json.Unmarshal(payload, &c); err != nil {
		return 0, err
	}
	if c.UserID <= 0 || c.Expiry < time.Now().Unix() {
		return 0, errors.New("token vencido")
	}
	return c.UserID, nil
}

func writeJSON(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
func methodNotAllowed(w http.ResponseWriter) {
	writeError(w, http.StatusMethodNotAllowed, "Metodo no permitido")
}

var _ = sql.ErrNoRows
