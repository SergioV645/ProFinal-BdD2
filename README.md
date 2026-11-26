# Aurum Motors — Sistema de Gestión para Taller y CDA

Sistema web para administrar un taller de vehículos y su Centro de Diagnóstico Automotor (CDA): clientes, vehículos, órdenes de trabajo, repuestos, servicios y revisiones técnico-mecánicas desde un solo panel.

Proyecto final del curso **Bases de Datos II**.

---

## Tabla de contenido

- [Funcionalidades](#funcionalidades)
- [Stack técnico](#stack-técnico)
- [Estructura del proyecto](#estructura-del-proyecto)
- [Base de datos](#base-de-datos)
- [Instalación y ejecución](#instalación-y-ejecución)
- [API REST](#api-rest)
- [Autores](#autores)

---

## Funcionalidades

- **Dashboard** con indicadores generales del taller.
- **Clientes**: registro y consulta, con opción de registrar cliente y vehículo en un solo paso.
- **Vehículos**: alta de vehículos asociados a un cliente y búsqueda por placa.
- **Órdenes de trabajo**: creación, seguimiento por estado, asignación de servicios y repuestos, cierre y facturación (IVA incluido).
- **Repuestos**: inventario con control de stock mínimo y alertas de bajo stock.
- **Servicios**: catálogo agrupado por categoría.
- **Revisiones CDA**: registro, edición y eliminación de revisiones con checklist de inspección, resultado y fecha de vencimiento.
- **Proveedores**: un repuesto puede asociarse a varios proveedores.
- **Historial técnico (MongoDB)**: eventos por vehículo con detalle flexible, CRUD completo y estadísticas por tipo de evento.
- **Usuarios y roles**: inicio de sesión con roles `ADMINISTRADOR` y `EMPLEADO`; contraseñas cifradas con bcrypt.

---

## Stack técnico

| Capa          | Tecnología                                  |
|---------------|---------------------------------------------|
| Backend       | Go 1.25+ (`net/http`, biblioteca estándar)  |
| Base relacional | MySQL 8+                                  |
| Base documental | MongoDB Atlas                             |
| Drivers       | `go-sql-driver/mysql`, `mongo-driver/v2`    |
| Seguridad     | `golang.org/x/crypto/bcrypt`, cookies HttpOnly |
| Frontend      | HTML, CSS y JavaScript (vanilla)            |

---

## Estructura del proyecto

```
.
├── backend/                  # API REST en Go
│   ├── main.go               # Punto de entrada, configuración y rutas
│   ├── go.mod / go.sum
│   └── internal/
│       ├── auth/             # Inicio de sesión, sesiones y roles
│       ├── db/               # Conexión a MySQL
│       ├── mongodb/          # Conexión a MongoDB Atlas
│       └── handlers/         # Lógica de cada recurso de la API
├── frontend/                 # Interfaz web (la sirve el backend)
│   ├── index.html
│   ├── css/
│   └── js/
├── database/                 # Scripts SQL, se ejecutan en orden
│   ├── 01_esquema.sql        # Base de datos y tablas
│   ├── 02_datos_prueba.sql   # Datos de ejemplo
│   ├── 03_logica_negocio.sql # Funciones, procedimientos, triggers y vistas
│   └── 04_roles_autenticacion.sql # Roles y usuarios
├── docs/
│   └── Proyecto_Final_v20261.pdf  # Enunciado / documentación del proyecto
├── .env.example              # Variables de entorno de referencia
└── README.md
```

---

## Base de datos

El esquema `taller_vehiculos` contiene las tablas:

`cargos`, `categorias_servicio`, `clientes`, `vehiculos`, `empleados`, `proveedores`, `repuestos`, `repuesto_proveedor`, `servicios`, `ordenes_trabajo`, `detalle_servicios`, `detalle_repuestos`, `facturas`, `revisiones_cda`, `items_inspeccion`, `detalle_revision_cda`, `roles`, `usuarios`.

En MongoDB se usa la base `aurum_nosql` con la colección `historial_vehicular`: cada documento es un evento técnico de un vehículo, con el técnico embebido y un campo `detalle` de estructura libre.

### Lógica de negocio

**Funciones**

| Nombre | Descripción |
|--------|-------------|
| `fn_calcular_total_orden` | Total (servicios + repuestos) de una orden. |
| `fn_obtener_stock_disponible` | Stock actual de un repuesto. |
| `fn_dias_ultima_revision_cda` | Días desde la última revisión CDA de un vehículo. |

**Procedimientos almacenados**

| Nombre | Descripción |
|--------|-------------|
| `sp_registrar_cliente_vehiculo` | Registra un cliente y su vehículo en una transacción. |
| `sp_registrar_orden_trabajo` | Abre una orden de trabajo. |
| `sp_agregar_servicio_orden` | Agrega un servicio a una orden. |
| `sp_agregar_repuesto_orden` | Agrega un repuesto validando el stock. |
| `sp_cerrar_orden_y_facturar` | Cierra la orden y genera la factura. |
| `sp_registrar_revision_cda` | Registra una revisión técnico-mecánica. |
| `sp_reporte_ordenes_por_periodo` | Reporte de órdenes en un rango de fechas. |

**Triggers**

- `trg_before_detalle_repuesto_insert`
- `trg_after_detalle_repuesto_insert`
- `trg_after_detalle_repuesto_delete`
- `trg_after_orden_update_completado`

**Vistas**

- `v_ordenes_detalladas`
- `v_repuestos_bajo_stock`
- `v_vehiculos_con_cda_vencida`

---

## Instalación y ejecución

### Requisitos

- [Go](https://go.dev/dl/) 1.25 o superior
- MySQL 8 o superior en ejecución

### 1. Clonar el repositorio

```bash
git clone https://github.com/SergioV645/ProFinal-BdD2.git
cd ProFinal-BdD2
```

### 2. Crear la base de datos

Ejecuta los scripts en orden:

```bash
mysql -u root -p < database/01_esquema.sql
mysql -u root -p < database/02_datos_prueba.sql
mysql -u root -p < database/03_logica_negocio.sql
mysql -u root -p < database/04_roles_autenticacion.sql
```

### 3. Configurar MongoDB Atlas

1. En Atlas, en **Network Access**, autoriza la IP pública desde la que se conecta el servidor.
2. En **Database Access**, crea un usuario con rol `readWrite` solo sobre la base `aurum_nosql`.
3. En **Connect > Drivers > Go**, copia la URI de conexión.

### 4. Configurar las variables de entorno

La configuración sensible (claves de MySQL y MongoDB, secreto de sesión y administrador inicial) **nunca se escribe en el código ni en este README**. Se guarda en un archivo `.env` local, que está en `.gitignore` y por eso no se sube a GitHub.

**1. Crea tu `.env` a partir de la plantilla** (en la raíz del proyecto):

Linux / macOS:

```bash
cp .env.example .env
chmod 600 .env   # solo tu usuario puede leerlo
```

Windows (PowerShell):

```powershell
Copy-Item .env.example .env
```

**2. Genera un `SESSION_SECRET` aleatorio** (mínimo 32 caracteres; el servidor no arranca con uno más corto) y pégalo en el `.env`:

Linux / macOS:

```bash
openssl rand -base64 48
```

Windows (PowerShell):

```powershell
[Convert]::ToBase64String((1..48 | ForEach-Object { Get-Random -Maximum 256 }))
```

**3. Completa el resto de valores en el `.env`** con tus credenciales reales. Usa para MySQL un usuario propio de la aplicación (`taller_app`), no `root`.

| Variable         | Obligatoria | Valor por defecto | Descripción |
|------------------|-------------|-------------------|-------------|
| `MYSQL_DSN`      | Sí          | —                 | Cadena de conexión a MySQL |
| `MONGODB_URI`    | Sí          | —                 | URI de MongoDB Atlas |
| `SESSION_SECRET` | Sí          | —                 | Secreto aleatorio para firmar las sesiones (32 caracteres o más) |
| `ADMIN_USERNAME` | Primera vez | —                 | Usuario del administrador inicial |
| `ADMIN_PASSWORD` | Primera vez | —                 | Contraseña del administrador inicial (10 caracteres o más) |
| `COOKIE_SECURE`  | No          | `false`           | `true` si se despliega con HTTPS |
| `PORT`           | No          | `8080`            | Puerto del servidor |
| `FRONTEND_DIR`   | No          | `../frontend`     | Carpeta del frontend |

`ADMIN_USERNAME` y `ADMIN_PASSWORD` solo se usan si la tabla `usuarios` está vacía, para crear el primer administrador (la contraseña se guarda cifrada con bcrypt). Después de ese primer arranque, bórralas del `.env` y cambia la contraseña desde la aplicación.

> **Buenas prácticas**
> - No escribas claves directamente en la terminal (`export CLAVE=...`): quedan guardadas en el historial de comandos.
> - Nunca subas el `.env` ni compartas su contenido por chat o correo.
> - En un servidor, define las variables en el propio entorno del sistema o del servicio; el backend les da prioridad sobre el `.env`.

### 5. Ejecutar el backend

```bash
cd backend
go run .
```

Al iniciar, el backend lee automáticamente el `.env` (busca en `backend/` y en la raíz del proyecto). Si falta una variable obligatoria, se detiene con un mensaje que indica cuál.

Abre **http://localhost:8080** e inicia sesión con el administrador. El backend sirve la API y el frontend.

---

## API REST

Todas las rutas, excepto el inicio de sesión, requieren una sesión activa. El rol `EMPLEADO` no puede eliminar registros ni administrar usuarios.

| Método | Endpoint | Descripción |
|--------|----------|-------------|
| POST | `/api/auth/login`, `/api/auth/logout` | Iniciar y cerrar sesión |
| GET | `/api/auth/me` | Usuario de la sesión actual |
| GET/POST | `/api/usuarios` | Listar / crear usuarios (solo administrador) |
| PUT | `/api/usuarios/{id}` | Editar un usuario (rol, estado o contraseña) |
| GET/POST | `/api/clientes` | Listar / crear clientes (con vehículo opcional) |
| GET | `/api/clientes/{id}`, `/api/clientes/{id}/vehiculos` | Un cliente y sus vehículos |
| GET/POST | `/api/vehiculos` | Listar / crear vehículos (filtro `?placa=`) |
| GET | `/api/empleados`, `/api/servicios`, `/api/proveedores` | Catálogos |
| GET/POST | `/api/repuestos` | Listar / crear repuestos |
| GET | `/api/repuestos/bajo-stock` | Repuestos por debajo del stock mínimo |
| GET/POST/DELETE | `/api/repuestos/{id}/proveedores` | Proveedores asociados a un repuesto |
| GET/POST | `/api/ordenes` | Listar / crear órdenes (filtro `?estado=`) |
| GET | `/api/ordenes/{id}` | Detalle de una orden con servicios y repuestos |
| POST | `/api/ordenes/{id}/servicios`, `/repuestos`, `/cerrar` | Agregar ítems, cerrar y facturar |
| GET/POST | `/api/revisionescda` | Listar / registrar revisiones CDA |
| GET/PUT/DELETE | `/api/revisionescda/{id}` y `/{id}/detalle` | Revisión CDA y su checklist |
| GET | `/api/items-inspeccion` | Ítems del checklist CDA |
| GET/POST | `/api/historial` | Listar / crear eventos del historial (MongoDB) |
| GET/PUT/DELETE | `/api/historial/{id}` | Consultar, editar o eliminar un evento |
| GET | `/api/historial/vehiculo/{id}` | Historial de un vehículo |
| GET | `/api/historial/estadisticas` | Total de eventos, vehículos atendidos y eventos por tipo |

---

## Autores

- **Federico Ospina**
- **Natalia Rodríguez**
- **Sergio Valderrama**

---

Proyecto académico desarrollado con fines educativos.
