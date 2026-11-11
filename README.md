# Aurum Motors — Sistema de Gestión para Taller y CDA

Sistema web académico para la administración de un taller de vehículos y su Centro de Diagnóstico Automotor (CDA). Incluye clientes, vehículos, empleados, órdenes de trabajo, servicios, repuestos, proveedores, facturación, revisiones CDA e historial técnico vehicular.

## Tecnologías

- **Backend:** Go
- **Frontend:** HTML, CSS, JavaScript y Bootstrap
- **Base de datos relacional:** MySQL 8.0+
- **Base de datos NoSQL:** MongoDB Atlas
- **Driver MySQL:** `github.com/go-sql-driver/mysql`
- **Driver MongoDB:** MongoDB Go Driver

## Estructura del proyecto

```text
ProFinal-BdD2/
├── backend/
│   ├── main.go
│   ├── db/
│   ├── handlers/
│   │   ├── repuesto_proveedor.go
│   │   └── cda_detalle.go
│   ├── mongodb/
│   └── public/
│       ├── index.html
│       ├── css/
│       └── js/
│           └── app.js
├── sql/
│   ├── 01_creacion.sql
│   ├── 02_insercion.sql
│   ├── 03_logica_negocio.sql
│   ├── 04_fix_vistas.sql
│   └── 05_tablas_faltantes.sql
└── README.md
```

## Funcionalidades

- Dashboard con indicadores generales.
- Gestión de clientes, vehículos y empleados.
- Gestión de órdenes de trabajo, servicios y repuestos.
- Control de inventario y alertas de bajo stock.
- Gestión de proveedores asociados a repuestos, con precio de compra y días de entrega.
- Revisiones CDA con catálogo de ítems de inspección y checklist por revisión.
- Registro y consulta del historial técnico vehicular en MongoDB Atlas.
- Selección de técnico desde los empleados activos registrados en MySQL.

## Base de datos MySQL

El esquema se llama `taller_vehiculos`.

### Tablas principales

`cargos`, `categorias_servicio`, `clientes`, `vehiculos`, `empleados`, `proveedores`, `repuestos`, `servicios`, `ordenes_trabajo`, `detalle_servicios`, `detalle_repuestos`, `facturas` y `revisiones_cda`.

### Ampliación del modelo — Script 05

El archivo `sql/05_tablas_faltantes.sql` agrega:

- `repuesto_proveedor`: relación muchos-a-muchos entre repuestos y proveedores; registra precio de compra y días de entrega. Se conserva `repuestos.proveedor_id` como proveedor principal para mantener compatibilidad con los datos y operaciones anteriores.
- `items_inspeccion`: catálogo de los puntos que se verifican durante una revisión CDA.
- `detalle_revision_cda`: resultado (`CONFORME` o `NO_CONFORME`) y observación de cada ítem por revisión.

El script usa `CREATE TABLE IF NOT EXISTS` e inserciones idempotentes para reducir errores por ejecuciones repetidas. Debe ejecutarse después de los scripts 01 a 04.

## Instalación de la base de datos

Requisitos: MySQL 8.0+ y MySQL Workbench.

### Primera instalación

Ejecutar los scripts en este orden:

1. `sql/01_creacion.sql`
2. `sql/02_insercion.sql`
3. `sql/03_logica_negocio.sql`
4. `sql/04_fix_vistas.sql`
5. `sql/05_tablas_faltantes.sql`

Desde MySQL Workbench, abrir cada archivo y ejecutarlo en el orden indicado.

También se pueden ejecutar desde una terminal con el cliente MySQL:

```bash
mysql -u root -p < sql/01_creacion.sql
mysql -u root -p < sql/02_insercion.sql
mysql -u root -p < sql/03_logica_negocio.sql
mysql -u root -p < sql/04_fix_vistas.sql
mysql -u root -p < sql/05_tablas_faltantes.sql
```

### Si la base de datos ya existe

No volver a ejecutar `01_creacion.sql` ni `02_insercion.sql` sobre una base que ya tiene datos: el primero crea el esquema y las tablas, y el segundo inserta datos de prueba que podrían duplicarse o fallar por restricciones únicas.

Para actualizar una instalación existente que ya tiene los scripts 01 a 04, ejecutar únicamente:

```bash
mysql -u root -p < sql/05_tablas_faltantes.sql
```

### Verificación del Script 05

```sql
USE taller_vehiculos;

SHOW TABLES;

SELECT COUNT(*) AS total_repuesto_proveedor
FROM repuesto_proveedor;

SELECT COUNT(*) AS total_items_inspeccion
FROM items_inspeccion;

SELECT COUNT(*) AS total_detalle_revision_cda
FROM detalle_revision_cda;
```

## Configuración y ejecución del backend

Requisitos: Go instalado y variables de entorno configuradas.

El backend lee las credenciales desde variables de entorno. No guardar contraseñas ni URI privadas directamente en el código o en Git.

En PowerShell, configurar la conexión MySQL:

```powershell
$env:MYSQL_DSN = "USUARIO:CLAVE@tcp(localhost:3306)/taller_vehiculos?parseTime=true"
```

Configurar la conexión de MongoDB Atlas:

```powershell
$env:MONGODB_URI = "TU_URI_DE_MONGODB_ATLAS"
```

Desde la carpeta `backend`:

```powershell
go mod tidy
go run .
```

El servidor queda disponible en:

`http://localhost:8080`

El frontend estático se sirve desde `backend/public`.

## Endpoints principales

| Método | Endpoint | Descripción |
|---|---|---|
| GET/POST | `/api/clientes` | Consultar y registrar clientes |
| GET | `/api/clientes/{id}` | Consultar cliente |
| GET/POST | `/api/vehiculos` | Consultar y registrar vehículos |
| GET | `/api/empleados` | Consultar empleados activos |
| GET/POST | `/api/servicios` | Consultar y registrar servicios |
| GET/POST | `/api/repuestos` | Consultar y registrar repuestos |
| GET | `/api/repuestos/bajo-stock` | Consultar repuestos con bajo stock |
| GET | `/api/proveedores` | Consultar proveedores disponibles |
| GET | `/api/repuestos/{id}/proveedores` | Consultar proveedores de un repuesto |
| POST | `/api/repuestos/{id}/proveedores` | Asociar o actualizar proveedor de un repuesto |
| DELETE | `/api/repuestos/{id}/proveedores?proveedorId={id}` | Eliminar asociación proveedor-repuesto |
| GET/POST | `/api/ordenes` | Consultar y registrar órdenes |
| GET | `/api/ordenes/{id}` | Consultar detalle de orden |
| GET/POST | `/api/revisionescda` | Consultar y registrar revisiones CDA |
| GET | `/api/items-inspeccion` | Consultar catálogo de ítems CDA |
| GET | `/api/revisionescda/{id}/detalle` | Consultar checklist de una revisión |
| POST | `/api/revisionescda/{id}/detalle` | Guardar o actualizar resultados del checklist |
| GET/POST | `/api/historial` | Consultar y registrar historial técnico en MongoDB |
| GET | `/api/historial/vehiculo/{id}` | Consultar historial por vehículo |
| GET/PUT/DELETE | `/api/historial/{id}` | Consultar, actualizar o eliminar evento del historial |

### Ejemplo: asociar proveedor a repuesto

`POST /api/repuestos/1/proveedores`

```json
{
  "proveedorId": 2,
  "precioCompra": 125000,
  "diasEntrega": 2
}
```

### Ejemplo: guardar checklist CDA

`POST /api/revisionescda/1/detalle`

```json
[
  {
    "itemId": 1,
    "resultado": "CONFORME",
    "observacion": ""
  },
  {
    "itemId": 2,
    "resultado": "NO_CONFORME",
    "observacion": "Eficacia de frenado por debajo del valor esperado"
  }
]
```

El flujo del formulario CDA registra primero la revisión general y después guarda su checklist. Si falla el segundo paso, la revisión general puede permanecer registrada; el detalle debe reintentarse.

## Autores

- Federico Ospina
- Liseth Natalia Vergel Rodriguez
- Sergio Valderrama

Proyecto académico desarrollado con fines educativos.
