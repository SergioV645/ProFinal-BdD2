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
- **Revisiones CDA**: registro de revisiones, resultado y fecha de vencimiento, con alerta de vehículos con CDA vencida.

---

## Stack técnico

| Capa          | Tecnología                                  |
|---------------|---------------------------------------------|
| Backend       | Go 1.22+ (`net/http`, biblioteca estándar)  |
| Base de datos | MySQL 8+                                    |
| Driver        | `github.com/go-sql-driver/mysql`            |
| Frontend      | HTML, CSS y JavaScript (vanilla)            |

---

## Estructura del proyecto

```
.
├── backend/                  # API REST en Go
│   ├── main.go               # Punto de entrada, configuración y rutas
│   ├── go.mod / go.sum
│   └── internal/
│       ├── db/               # Conexión a MySQL
│       └── handlers/         # Lógica de cada recurso de la API
├── frontend/                 # Interfaz web (la sirve el backend)
│   ├── index.html
│   ├── css/
│   └── js/
├── database/                 # Scripts SQL, se ejecutan en orden
│   ├── 01_esquema.sql        # Base de datos y tablas
│   ├── 02_datos_prueba.sql   # Datos de ejemplo
│   └── 03_logica_negocio.sql # Funciones, procedimientos, triggers y vistas
├── docs/
│   └── Proyecto_Final_v20261.pdf  # Enunciado / documentación del proyecto
├── .env.example              # Variables de entorno de referencia
└── README.md
```

---

## Base de datos

El esquema `taller_vehiculos` contiene las tablas:

`cargos`, `categorias_servicio`, `clientes`, `vehiculos`, `empleados`, `proveedores`, `repuestos`, `servicios`, `ordenes_trabajo`, `detalle_servicios`, `detalle_repuestos`, `facturas`, `revisiones_cda`.

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

- [Go](https://go.dev/dl/) 1.22 o superior
- MySQL 8 o superior en ejecución

### 1. Clonar el repositorio

```bash
git clone https://github.com/SergioV645/ProFinal-BdD2.git
cd ProFinal-BdD2
```

### 2. Crear la base de datos

Ejecuta los scripts en orden (también puedes abrirlos en MySQL Workbench):

```bash
mysql -u root -p < database/01_esquema.sql
mysql -u root -p < database/02_datos_prueba.sql
mysql -u root -p < database/03_logica_negocio.sql
```

### 3. Configurar la conexión

El backend lee la configuración de variables de entorno (ver `.env.example`):

| Variable       | Obligatoria | Valor por defecto | Descripción |
|----------------|-------------|-------------------|-------------|
| `MYSQL_DSN`    | Sí          | —                 | Cadena de conexión a MySQL |
| `PORT`         | No          | `8080`            | Puerto del servidor |
| `FRONTEND_DIR` | No          | `../frontend`     | Carpeta del frontend |

### 4. Ejecutar el backend

Linux / macOS:

```bash
cd backend
export MYSQL_DSN="root:tu_clave@tcp(localhost:3306)/taller_vehiculos?parseTime=true"
go run .
```

Windows (PowerShell):

```powershell
cd backend
$env:MYSQL_DSN="root:tu_clave@tcp(localhost:3306)/taller_vehiculos?parseTime=true"
go run .
```

Abre **http://localhost:8080** en el navegador. El backend sirve la API y el frontend.

---

## API REST

| Método   | Endpoint                          | Descripción |
|----------|-----------------------------------|-------------|
| GET/POST | `/api/clientes`                   | Listar / crear clientes (con vehículo opcional) |
| GET      | `/api/clientes/{id}`              | Detalle de un cliente y sus vehículos |
| GET/POST | `/api/vehiculos`                  | Listar / crear vehículos (filtro `?placa=`) |
| GET      | `/api/empleados`                  | Listar empleados |
| GET      | `/api/servicios`                  | Listar servicios |
| GET/POST | `/api/repuestos`                  | Listar / crear repuestos |
| GET      | `/api/repuestos/bajo-stock`       | Repuestos por debajo del stock mínimo |
| GET/POST | `/api/ordenes`                    | Listar / crear órdenes (filtro `?estado=`) |
| GET      | `/api/ordenes/{id}`               | Detalle de una orden con servicios y repuestos |
| POST     | `/api/ordenes/{id}/servicios`     | Agregar un servicio a la orden |
| POST     | `/api/ordenes/{id}/repuestos`     | Agregar un repuesto a la orden |
| POST     | `/api/ordenes/{id}/cerrar`        | Cerrar la orden y facturar |
| GET/POST | `/api/revisionescda`              | Listar / registrar revisiones CDA |

---

## Autores

- **Federico Ospina**
- **Natalia Rodríguez**
- **Sergio Valderrama**

---

Proyecto académico desarrollado con fines educativos.
