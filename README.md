# Aurum Motors — Sistema de Gestión para Taller & CDA

Sistema web para la administración integral de un taller de vehículos y su Centro de Diagnóstico Automotor (CDA): clientes, vehículos, órdenes de trabajo, repuestos, servicios y revisiones técnico-mecánicas, todo desde un solo dashboard.

Proyecto final del curso de Bases de Datos II.

---

## Funcionalidades

- **Dashboard** con indicadores generales del taller.
- **Clientes**: registro y consulta de clientes.
- **Vehículos**: alta de vehículos asociados a cada cliente, búsqueda por placa.
- **Órdenes de trabajo**: creación, seguimiento por estado, asignación de servicios y repuestos, cierre y facturación.
- **Repuestos**: inventario con control de stock mínimo y alertas de bajo stock.
- **Servicios**: catálogo de servicios agrupados por categoría.
- **Revisiones CDA**: registro de revisiones técnico-mecánicas, resultado y fecha de vencimiento, con alerta de vehículos con CDA vencida.

---

## Stack técnico

| Capa           | Tecnología                                   |
|----------------|-----------------------------------------------|
| Backend        | Go     
| Base de datos  | MySQL                                          |
| Driver DB      | `go-sql-driver/mysql`                          |
| Frontend       | HTML, CSS, JavaScript (vanilla) + Bootstrap    |

---

## Estructura del proyecto

```
Proyecto Final/
├── backend/
│   ├── main.go              # Punto de entrada y rutas HTTP
│   ├── db/
│   │   └── db.go            # Conexión a MySQL
│   ├── handlers/            # Lógica de cada endpoint
│   └── public/               # Frontend estático
├── sql/
│   ├── 01_creacion.sql       
│   ├── 02_insercion.sql      # Datos de prueba
│   ├── 03_logica_negocio.sql 
│   └── 04_fix_vistas.sql     
└── Proyecto_Final_v20261.pdf # Documentación / informe del proyecto
```

---
## Base de datos
El esquema `taller_vehiculos` incluye las siguientes tablas principales:

`cargos`, `categorias_servicio`, `clientes`, `vehiculos`, `empleados`, `proveedores`, `repuestos`, `servicios`, `ordenes_trabajo`, `detalle_servicios`, `detalle_repuestos`, `facturas`, `revisiones_cda`.

### Lógica de negocio
-- Por avanzar**
**Funciones**
- `fn_calcular_total_orden` — calcula el total (servicios + repuestos) de una orden.
- `fn_obtener_stock_disponible` — retorna el stock actual de un repuesto.
- `fn_dias_ultima_revision_cda` — días transcurridos desde la última revisión CDA de un vehículo.

**Procedimientos almacenados**
- `sp_registrar_cliente_vehiculo`
- `sp_registrar_orden_trabajo`
- `sp_agregar_servicio_orden`
- `sp_agregar_repuesto_orden`
- `sp_cerrar_orden_y_facturar`
- `sp_registrar_revision_cda`
- `sp_reporte_ordenes_por_periodo`

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

### Requisitos previos
- [Go](https://go.dev/dl/) 1.22 o superior
- MySQL 8+ en ejecución local

### 1. Clonar el repositorio

```bash
git clone https://github.com/SergioV645/ProFinal-BdD2.git
cd ProFinal-BdD2/ProyectoFinal
```

### 2. Crear la base de datos

Ejecuta los scripts SQL en orden desde tu cliente MySQL preferido (Workbench, DBeaver, CLI, etc.):

```bash
mysql -u root -p < sql/01_creacion.sql
mysql -u root -p < sql/02_insercion.sql
mysql -u root -p < sql/03_logica_negocio.sql
# Si el script 03 falla al crear las vistas, ejecuta también:
mysql -u root -p < sql/04_fix_vistas.sql
```

### 3. Configurar la conexión

En `backend/main.go`, ajusta el DSN con tus credenciales de MySQL:

```go
dsn := "root:0000@tcp(localhost:3306)/taller_vehiculos?parseTime=true"
```

### 4. Instalar dependencias y ejecutar el backend

```bash
cd backend
go mod tidy
go run main.go
```

El servidor quedará disponible en **http://localhost:8080**, sirviendo también el frontend estático desde `backend/public`.

---
## Endpoints principales de la API

| Método | Endpoint                          | Descripción                              |
|--------|-------------------------------------|-------------------------------------------|
| GET/POST | `/api/clientes`                   | Listar / crear clientes                   |
| GET    | `/api/clientes/{id}`               | Detalle de un cliente                     |
| GET/POST | `/api/vehiculos`                  | Listar / crear vehículos (filtro `?placa=`) |
| GET/POST | `/api/empleados`                  | Listar / crear empleados                  |
| GET/POST | `/api/servicios`                  | Listar / crear servicios                  |
| GET/POST | `/api/repuestos`                  | Listar / crear repuestos                  |
| GET    | `/api/repuestos/bajo-stock`        | Repuestos por debajo del stock mínimo     |
| GET/POST | `/api/ordenes`                    | Listar / crear órdenes (filtro `?estado=`) |
| GET    | `/api/ordenes/{id}`                | Detalle de una orden                      |
| GET/POST | `/api/revisionescda`              | Listar / registrar revisiones CDA         |

---
## Autores

**Federico Ospina**
**Natalia Rodriguez**
**Sergio Valderrama**

---
Proyecto académico desarrollado con fines educativos.
