# Aurum Motors — Sistema de gestión para Taller y CDA

Aplicación web académica para administrar clientes, vehículos, órdenes de trabajo, repuestos, proveedores, revisiones CDA e historial técnico vehicular. Incluye autenticación con dos roles: **ADMINISTRADOR** y **EMPLEADO**.

## Tecnologías

- Backend: Go
- Frontend: HTML, CSS, JavaScript y Bootstrap
- Base relacional: MySQL 8.0+
- Historial técnico: MongoDB Atlas
- Contraseñas: bcrypt
- Sesión: cookie HttpOnly firmada, con expiración de 8 horas

## Roles y permisos

| Funcionalidad | Administrador | Empleado |
|---|---|---|
| Dashboard | Consultar | Consultar |
| Clientes, vehículos y órdenes | Crear, consultar, editar y eliminar | Crear, consultar y editar* |
| Repuestos | Crear, consultar, editar y eliminar | Crear, consultar y editar* |
| Proveedores | Administrar asociaciones | Solo consultar |
| Revisiones y checklist CDA | Crear, consultar, editar y eliminar | Crear, consultar y editar |
| Historial técnico MongoDB | Crear, consultar, editar y eliminar | Crear, consultar y editar |
| Usuarios | Crear, consultar, activar/desactivar y cambiar contraseñas | Sin acceso |
| Roles y permisos | Gestionados por el sistema | Sin acceso |

\* La autorización del backend permite operaciones POST/PUT operativas; cada módulo debe tener implementado su handler y formulario de edición para ofrecer esa acción en la interfaz. Las rutas que todavía no implementen PUT responderán `405 Method Not Allowed`.

**Seguridad:** los permisos se verifican en el backend. Ocultar botones en el frontend es solo una mejora de interfaz y no reemplaza la autorización del servidor. El empleado no puede ejecutar peticiones DELETE ni modificar usuarios o asociaciones de proveedores.

## Estructura

```text
ProFinal-BdD2/
├── backend/
│   ├── auth/                 # Login, sesiones, roles y autorización
│   ├── db/
│   ├── handlers/
│   ├── mongodb/
│   └── public/
│       ├── index.html
│       ├── css/style.css
│       └── js/app.js
├── sql/
│   ├── 01_creacion.sql
│   ├── 02_insercion.sql
│   ├── 03_logica_negocio.sql
│   ├── 04_fix_vistas.sql
│   └── 05_roles_autenticacion.sql
└── README.md
```

## Instalación desde cero

Requisitos: MySQL 8.0+, Go 1.25+, MySQL Workbench y una URI de MongoDB Atlas.

Ejecuta en MySQL Workbench, en este orden:

1. `sql/01_creacion.sql`
2. `sql/02_insercion.sql`
3. `sql/03_logica_negocio.sql`
4. `sql/04_fix_vistas.sql`

El esquema de instalación inicial ya incluye las tablas de proveedores múltiples, checklist CDA y autenticación (`roles` y `usuarios`). El script 02 inserta los dos roles, pero **no crea una contraseña predeterminada**. El primer administrador se crea de forma segura al iniciar el backend mediante variables de entorno.

No ejecutes los scripts 01 y 02 sobre una base existente con información: 01 crea el esquema y 02 inserta datos de prueba.

## Actualizar una base de datos existente

Si ya tienes `taller_vehiculos` instalada con los módulos actuales y no tiene autenticación, ejecuta únicamente:

```sql
SOURCE ruta/a/sql/06_roles_autenticacion.sql;
```

O abre `sql/05_roles_autenticacion.sql` en MySQL Workbench y ejecútalo completo.

La migración:
- Crea `roles` y `usuarios` si no existen.
- Inserta `ADMINISTRADOR` y `EMPLEADO` si faltan.
- No elimina ni modifica clientes, vehículos, órdenes, repuestos, revisiones ni historial.
- No crea contraseñas ni usuarios predeterminados.

Si la tabla `usuarios` ya existe y tiene una estructura diferente, detén la instalación y revisa el esquema antes de continuar.

## Configurar variables de entorno

El backend requiere cuatro variables. No guardes credenciales reales en Git.

### PowerShell — primera ejecución sin usuarios

```powershell
$env:MYSQL_DSN = "USUARIO:CLAVE@tcp(localhost:3306)/taller_vehiculos?parseTime=true"
$env:MONGODB_URI = "TU_URI_DE_MONGODB_ATLAS"
$env:SESSION_SECRET = "CAMBIA_ESTO_POR_UN_SECRETO_ALEATORIO_DE_32_CARACTERES_O_MAS"
$env:ADMIN_USERNAME = "admin"
$env:ADMIN_PASSWORD = "DefineUnaClaveLargaYUnica"
```

`ADMIN_PASSWORD` debe tener al menos 10 caracteres. `SESSION_SECRET` debe tener al menos 32 caracteres. Usa valores propios y privados.

Desde la carpeta `backend`:

```powershell
go mod tidy
go run .
```

En el primer arranque, si `usuarios` está vacía, se crea el administrador inicial con bcrypt. En arranques posteriores, si ya existe al menos un usuario, las variables `ADMIN_USERNAME` y `ADMIN_PASSWORD` no se vuelven a utilizar para crear cuentas.

Después de crear el administrador, puedes limpiar las variables de bootstrap de la sesión actual:

```powershell
Remove-Item Env:ADMIN_USERNAME
Remove-Item Env:ADMIN_PASSWORD
```

Conserva `MYSQL_DSN`, `MONGODB_URI` y `SESSION_SECRET` para cada arranque. Si cambias `SESSION_SECRET`, todas las sesiones abiertas deberán iniciar sesión nuevamente.

La aplicación queda en `http://localhost:8080`.

## Crear cuentas de empleados

1. Inicia sesión con el administrador inicial.
2. Abre **Usuarios** en el menú lateral.
3. Selecciona un empleado activo existente.
4. Define nombre de usuario y contraseña (mínimo 10 caracteres).
5. Pulsa **Crear usuario**.

Cada empleado solo puede tener una cuenta. Las cuentas nuevas reciben automáticamente el rol `EMPLEADO`. El rol del administrador inicial es `ADMINISTRADOR`.

Desde el módulo Usuarios, el administrador puede activar/desactivar cuentas y cambiar contraseñas. No se permite administrar roles arbitrarios: el sistema mantiene únicamente los dos roles definidos.

## Endpoints de autenticación y usuarios

| Método | Endpoint | Acceso | Descripción |
|---|---|---|---|
| POST | `/api/auth/login` | Público | Iniciar sesión |
| POST | `/api/auth/logout` | Público | Cerrar sesión |
| GET | `/api/auth/me` | Autenticado | Consultar usuario actual |
| GET/POST | `/api/usuarios` | Administrador | Listar/crear usuarios empleados |
| PUT | `/api/usuarios/{id}` | Administrador | Activar, desactivar o cambiar contraseña |

El login establece una cookie `HttpOnly`, `SameSite=Strict`, con duración de 8 horas. Para despliegues HTTPS, define `COOKIE_SECURE=true`.

## Endpoints funcionales principales

| Método | Endpoint | Descripción |
|---|---|---|
| GET/POST | `/api/clientes` | Consultar y registrar clientes |
| GET | `/api/clientes/{id}` | Consultar cliente |
| GET/POST | `/api/vehiculos` | Consultar y registrar vehículos |
| GET | `/api/empleados` | Consultar empleados activos |
| GET | `/api/servicios` | Consultar servicios |
| GET/POST | `/api/repuestos` | Consultar y registrar repuestos |
| GET | `/api/repuestos/bajo-stock` | Consultar bajo stock |
| GET | `/api/proveedores` | Consultar proveedores |
| GET/POST/DELETE | `/api/repuestos/{id}/proveedores` | Consultar, asociar o quitar proveedor (solo administrador puede escribir) |
| GET/POST | `/api/ordenes` | Consultar y registrar órdenes |
| GET | `/api/ordenes/{id}` | Consultar detalle de orden |
| GET/POST | `/api/revisionescda` | Consultar y registrar revisiones CDA |
| GET/PUT/DELETE | `/api/revisionescda/{id}` | Consultar/editar/eliminar revisión (DELETE solo administrador) |
| GET | `/api/items-inspeccion` | Consultar catálogo CDA |
| GET/POST | `/api/revisionescda/{id}/detalle` | Consultar/guardar checklist |
| GET/POST | `/api/historial` | Consultar/registrar historial MongoDB |
| GET/PUT/DELETE | `/api/historial/{id}` | Consultar/editar/eliminar evento (DELETE solo administrador) |
| GET | `/api/historial/vehiculo/{id}` | Consultar historial por vehículo |
| GET | `/api/historial/estadisticas` | Consultar estadísticas |

## Verificación de roles

Después de crear las cuentas:

1. Inicia sesión como administrador y verifica que aparezca **Usuarios**.
2. Crea una cuenta vinculada a un empleado.
3. Cierra sesión e inicia como empleado.
4. Comprueba que no aparece el módulo Usuarios ni los botones de eliminación.
5. Comprueba que un intento de DELETE desde una herramienta HTTP responde `403 Forbidden`.
6. Verifica que un empleado no puede consultar `/api/usuarios` (debe responder `403 Forbidden`).
7. Inicia de nuevo como administrador y verifica que puede administrar usuarios.

## Notas de seguridad

- Nunca uses una contraseña real en el repositorio.
- No compartas `MYSQL_DSN`, `MONGODB_URI` ni `SESSION_SECRET`.
- En producción usa HTTPS y `COOKIE_SECURE=true`.
- Las cuentas se desactivan en vez de borrarse para conservar trazabilidad.
- El backend vuelve a consultar el estado y rol de la cuenta en cada petición autenticada.
- Esta aplicación es un proyecto académico; antes de un despliegue público se recomienda añadir protección CSRF explícita, rate limiting de login, auditoría de acciones y gestión de recuperación de contraseña.

## Autores

- Federico Ospina
- Liseth Natalia Vergel Rodriguez
- Sergio Valderrama
