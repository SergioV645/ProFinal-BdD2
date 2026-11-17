-- ============================================================
-- TALLER DE VEHICULOS & CDA
-- Script 01: Creacion de base de datos y tablas (incluye proveedores multiples y checklist CDA)
-- Motor: MySQL 8.0+
-- ============================================================

CREATE DATABASE taller_vehiculos
    CHARACTER SET utf8mb4
    COLLATE utf8mb4_unicode_ci;

USE taller_vehiculos;

-- -----------------------------------------------------------
-- 1. CARGOS (normaliza el tipo de cargo del empleado)
-- -----------------------------------------------------------
CREATE TABLE cargos (
    cargo_id    INT           NOT NULL AUTO_INCREMENT,
    nombre      VARCHAR(60)   NOT NULL,
    descripcion VARCHAR(200)  NULL,
    PRIMARY KEY (cargo_id),
    UNIQUE KEY uq_cargo_nombre (nombre)
) ENGINE=InnoDB;

-- -----------------------------------------------------------
-- 2. CATEGORIAS_SERVICIO (normaliza categoría de servicio)
-- -----------------------------------------------------------
CREATE TABLE categorias_servicio (
    categoria_id INT          NOT NULL AUTO_INCREMENT,
    nombre       VARCHAR(60)  NOT NULL,
    PRIMARY KEY (categoria_id),
    UNIQUE KEY uq_cat_nombre (nombre)
) ENGINE=InnoDB;

-- -----------------------------------------------------------
-- 3. CLIENTES
-- -----------------------------------------------------------
CREATE TABLE clientes (
    cliente_id      INT           NOT NULL AUTO_INCREMENT,
    cedula          VARCHAR(20)   NOT NULL,
    nombre          VARCHAR(80)   NOT NULL,
    apellido        VARCHAR(80)   NOT NULL,
    telefono        VARCHAR(20)   NULL,
    email           VARCHAR(120)  NULL,
    direccion       VARCHAR(200)  NULL,
    fecha_registro  DATE          NOT NULL DEFAULT (CURRENT_DATE),
    PRIMARY KEY (cliente_id),
    UNIQUE KEY uq_cliente_cedula (cedula)
) ENGINE=InnoDB;

-- -----------------------------------------------------------
-- 4. VEHICULOS  (1:N con clientes)
-- -----------------------------------------------------------
CREATE TABLE vehiculos (
    vehiculo_id  INT          NOT NULL AUTO_INCREMENT,
    cliente_id   INT          NOT NULL,
    placa        VARCHAR(10)  NOT NULL,
    marca        VARCHAR(50)  NOT NULL,
    modelo       VARCHAR(60)  NOT NULL,
    anio         SMALLINT     NOT NULL,
    color        VARCHAR(30)  NULL,
    cilindraje   DECIMAL(5,1) NULL,
    km_actual    INT          NOT NULL DEFAULT 0,
    PRIMARY KEY (vehiculo_id),
    UNIQUE KEY uq_vehiculo_placa (placa),
    CONSTRAINT fk_veh_cliente FOREIGN KEY (cliente_id)
        REFERENCES clientes (cliente_id)
        ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB;

-- -----------------------------------------------------------
-- 5. EMPLEADOS  (1:N con cargos)
-- -----------------------------------------------------------
CREATE TABLE empleados (
    empleado_id   INT           NOT NULL AUTO_INCREMENT,
    cargo_id      INT           NOT NULL,
    cedula        VARCHAR(20)   NOT NULL,
    nombre        VARCHAR(80)   NOT NULL,
    apellido      VARCHAR(80)   NOT NULL,
    telefono      VARCHAR(20)   NULL,
    salario       DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    fecha_ingreso DATE          NOT NULL DEFAULT (CURRENT_DATE),
    activo        TINYINT(1)    NOT NULL DEFAULT 1,
    PRIMARY KEY (empleado_id),
    UNIQUE KEY uq_emp_cedula (cedula),
    CONSTRAINT fk_emp_cargo FOREIGN KEY (cargo_id)
        REFERENCES cargos (cargo_id)
        ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB;

-- -----------------------------------------------------------
-- 6. PROVEEDORES
-- -----------------------------------------------------------
CREATE TABLE proveedores (
    proveedor_id INT          NOT NULL AUTO_INCREMENT,
    nit          VARCHAR(20)  NOT NULL,
    nombre       VARCHAR(120) NOT NULL,
    telefono     VARCHAR(20)  NULL,
    email        VARCHAR(120) NULL,
    ciudad       VARCHAR(60)  NULL,
    PRIMARY KEY (proveedor_id),
    UNIQUE KEY uq_prov_nit (nit)
) ENGINE=InnoDB;

-- -----------------------------------------------------------
-- 7. REPUESTOS  (proveedor principal; N:M ampliada en 05_tablas_faltantes.sql)
-- -----------------------------------------------------------
CREATE TABLE repuestos (
    repuesto_id   INT            NOT NULL AUTO_INCREMENT,
    proveedor_id  INT            NOT NULL,
    referencia    VARCHAR(40)    NOT NULL,
    nombre        VARCHAR(120)   NOT NULL,
    descripcion   VARCHAR(300)   NULL,
    precio_compra DECIMAL(12,2)  NOT NULL DEFAULT 0.00,
    precio_venta  DECIMAL(12,2)  NOT NULL DEFAULT 0.00,
    stock_actual  INT            NOT NULL DEFAULT 0,
    stock_minimo  INT            NOT NULL DEFAULT 5,
    PRIMARY KEY (repuesto_id),
    UNIQUE KEY uq_rep_referencia (referencia),
    CONSTRAINT fk_rep_proveedor FOREIGN KEY (proveedor_id)
        REFERENCES proveedores (proveedor_id)
        ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB;

-- -----------------------------------------------------------
-- 8. SERVICIOS  (1:N con categorias_servicio)
-- -----------------------------------------------------------
CREATE TABLE servicios (
    servicio_id  INT            NOT NULL AUTO_INCREMENT,
    categoria_id INT            NOT NULL,
    nombre       VARCHAR(120)   NOT NULL,
    descripcion  VARCHAR(300)   NULL,
    precio_base  DECIMAL(12,2)  NOT NULL DEFAULT 0.00,
    PRIMARY KEY (servicio_id),
    CONSTRAINT fk_ser_categoria FOREIGN KEY (categoria_id)
        REFERENCES categorias_servicio (categoria_id)
        ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB;

-- -----------------------------------------------------------
-- 9. ORDENES_TRABAJO  (1:N con vehiculos y empleados)
-- -----------------------------------------------------------
CREATE TABLE ordenes_trabajo (
    orden_id              INT           NOT NULL AUTO_INCREMENT,
    vehiculo_id           INT           NOT NULL,
    empleado_id           INT           NOT NULL,
    fecha_ingreso         DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    fecha_estimada        DATE          NULL,
    fecha_salida          DATETIME      NULL,
    estado                ENUM('PENDIENTE','EN_PROCESO','COMPLETADO','ENTREGADO','CANCELADO')
                          NOT NULL DEFAULT 'PENDIENTE',
    km_ingreso            INT           NOT NULL DEFAULT 0,
    diagnostico           TEXT          NULL,
    observaciones         TEXT          NULL,
    PRIMARY KEY (orden_id),
    CONSTRAINT fk_ord_vehiculo  FOREIGN KEY (vehiculo_id)
        REFERENCES vehiculos (vehiculo_id) ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_ord_empleado  FOREIGN KEY (empleado_id)
        REFERENCES empleados (empleado_id) ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB;

-- -----------------------------------------------------------
-- 10. DETALLE_SERVICIOS  (N:M entre ordenes_trabajo y servicios)
-- -----------------------------------------------------------
CREATE TABLE detalle_servicios (
    detalle_id      INT            NOT NULL AUTO_INCREMENT,
    orden_id        INT            NOT NULL,
    servicio_id     INT            NOT NULL,
    cantidad        INT            NOT NULL DEFAULT 1,
    precio_unitario DECIMAL(12,2)  NOT NULL,
    mano_obra       DECIMAL(12,2)  NOT NULL DEFAULT 0.00,
    subtotal        DECIMAL(12,2)  NOT NULL,
    PRIMARY KEY (detalle_id),
    CONSTRAINT fk_ds_orden    FOREIGN KEY (orden_id)
        REFERENCES ordenes_trabajo (orden_id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_ds_servicio FOREIGN KEY (servicio_id)
        REFERENCES servicios (servicio_id) ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB;

-- -----------------------------------------------------------
-- 11. DETALLE_REPUESTOS  (N:M entre ordenes_trabajo y repuestos)
-- -----------------------------------------------------------
CREATE TABLE detalle_repuestos (
    detalle_id      INT            NOT NULL AUTO_INCREMENT,
    orden_id        INT            NOT NULL,
    repuesto_id     INT            NOT NULL,
    cantidad        INT            NOT NULL DEFAULT 1,
    precio_unitario DECIMAL(12,2)  NOT NULL,
    subtotal        DECIMAL(12,2)  NOT NULL,
    PRIMARY KEY (detalle_id),
    CONSTRAINT fk_dr_orden    FOREIGN KEY (orden_id)
        REFERENCES ordenes_trabajo (orden_id) ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_dr_repuesto FOREIGN KEY (repuesto_id)
        REFERENCES repuestos (repuesto_id) ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB;

-- -----------------------------------------------------------
-- 12. FACTURAS  (1:1 con ordenes_trabajo)
-- -----------------------------------------------------------
CREATE TABLE facturas (
    factura_id         INT            NOT NULL AUTO_INCREMENT,
    orden_id           INT            NOT NULL,
    fecha_emision      DATETIME       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    subtotal_servicios DECIMAL(12,2)  NOT NULL DEFAULT 0.00,
    subtotal_repuestos DECIMAL(12,2)  NOT NULL DEFAULT 0.00,
    iva                DECIMAL(12,2)  NOT NULL DEFAULT 0.00,
    total              DECIMAL(12,2)  NOT NULL DEFAULT 0.00,
    estado_pago        ENUM('PENDIENTE','PAGADO','ANULADO') NOT NULL DEFAULT 'PENDIENTE',
    metodo_pago        ENUM('EFECTIVO','TARJETA','TRANSFERENCIA') NULL,
    PRIMARY KEY (factura_id),
    UNIQUE KEY uq_factura_orden (orden_id),        -- 1:1 con ordenes_trabajo
    CONSTRAINT fk_fac_orden FOREIGN KEY (orden_id)
        REFERENCES ordenes_trabajo (orden_id) ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB;

-- -----------------------------------------------------------
-- 13. REVISIONES_CDA
-- -----------------------------------------------------------
CREATE TABLE revisiones_cda (
    revision_id       INT          NOT NULL AUTO_INCREMENT,
    vehiculo_id       INT          NOT NULL,
    empleado_id       INT          NOT NULL,
    fecha_revision    DATE         NOT NULL DEFAULT (CURRENT_DATE),
    resultado         ENUM('APROBADO','REPROBADO','APROBADO_CON_OBSERVACIONES')
                      NOT NULL DEFAULT 'APROBADO',
    observaciones     TEXT         NULL,
    fecha_vencimiento DATE         NULL,
    PRIMARY KEY (revision_id),
    CONSTRAINT fk_cda_vehiculo  FOREIGN KEY (vehiculo_id)
        REFERENCES vehiculos (vehiculo_id) ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_cda_empleado  FOREIGN KEY (empleado_id)
        REFERENCES empleados (empleado_id) ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB;


-- -----------------------------------------------------------
-- 14. REPUESTO_PROVEEDOR (N:M entre repuestos y proveedores)
-- Se conserva repuestos.proveedor_id como proveedor principal
-- para compatibilidad con consultas y funcionalidades existentes.
-- -----------------------------------------------------------
CREATE TABLE repuesto_proveedor (
    repuesto_id   INT           NOT NULL,
    proveedor_id  INT           NOT NULL,
    precio_compra DECIMAL(12,2) NOT NULL,
    dias_entrega  SMALLINT      NOT NULL DEFAULT 1,
    PRIMARY KEY (repuesto_id, proveedor_id),
    CONSTRAINT ck_rp_precio CHECK (precio_compra > 0),
    CONSTRAINT ck_rp_dias CHECK (dias_entrega >= 0),
    CONSTRAINT fk_rp_repuesto FOREIGN KEY (repuesto_id)
        REFERENCES repuestos (repuesto_id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_rp_proveedor FOREIGN KEY (proveedor_id)
        REFERENCES proveedores (proveedor_id)
        ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB;

CREATE INDEX ix_rp_proveedor ON repuesto_proveedor (proveedor_id);

-- -----------------------------------------------------------
-- 15. ITEMS_INSPECCION (catalogo de elementos del checklist CDA)
-- -----------------------------------------------------------
CREATE TABLE items_inspeccion (
    item_id     INT          NOT NULL AUTO_INCREMENT,
    nombre      VARCHAR(100) NOT NULL,
    grupo       VARCHAR(40)  NOT NULL,
    es_critico  BOOLEAN      NOT NULL DEFAULT FALSE,
    PRIMARY KEY (item_id),
    UNIQUE KEY uq_item_nombre (nombre)
) ENGINE=InnoDB;

-- -----------------------------------------------------------
-- 16. DETALLE_REVISION_CDA (resultado por item y revision)
-- -----------------------------------------------------------
CREATE TABLE detalle_revision_cda (
    revision_id INT NOT NULL,
    item_id     INT NOT NULL,
    resultado   ENUM('CONFORME','NO_CONFORME') NOT NULL,
    observacion VARCHAR(200) NULL,
    PRIMARY KEY (revision_id, item_id),
    CONSTRAINT fk_drc_revision FOREIGN KEY (revision_id)
        REFERENCES revisiones_cda (revision_id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_drc_item FOREIGN KEY (item_id)
        REFERENCES items_inspeccion (item_id)
        ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB;

CREATE INDEX ix_drc_item ON detalle_revision_cda (item_id, resultado);

-- -----------------------------------------------------------
-- Indices adicionales para consultas frecuentes
-- -----------------------------------------------------------
CREATE INDEX idx_ord_estado      ON ordenes_trabajo (estado);
CREATE INDEX idx_ord_fecha       ON ordenes_trabajo (fecha_ingreso);
CREATE INDEX idx_cda_vehiculo    ON revisiones_cda  (vehiculo_id, fecha_revision);
CREATE INDEX idx_rep_stock       ON repuestos        (stock_actual);

SHOW TABLES;
