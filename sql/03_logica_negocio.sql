-- ============================================================
-- TALLER DE VEHICULOS & CDA
-- Script 03: Logica de negocio
--   - Funciones
--   - Procedimientos Almacenados
--   - Triggers
--   - Control de transacciones y excepciones
-- ============================================================

USE taller_vehiculos;

DELIMITER $$

-- ============================================================
-- FUNCIONES
-- ============================================================

-- ------------------------------------------------------------
-- fn_calcular_total_orden
-- Retorna el total (servicios + repuestos) de una orden dada.
-- ------------------------------------------------------------
DROP FUNCTION IF EXISTS fn_calcular_total_orden$$
CREATE FUNCTION fn_calcular_total_orden(p_orden_id INT)
RETURNS DECIMAL(12,2)
DETERMINISTIC
READS SQL DATA
BEGIN
    DECLARE v_total_servicios DECIMAL(12,2) DEFAULT 0;
    DECLARE v_total_repuestos DECIMAL(12,2) DEFAULT 0;

    SELECT COALESCE(SUM(subtotal), 0)
      INTO v_total_servicios
      FROM detalle_servicios
     WHERE orden_id = p_orden_id;

    SELECT COALESCE(SUM(subtotal), 0)
      INTO v_total_repuestos
      FROM detalle_repuestos
     WHERE orden_id = p_orden_id;

    RETURN v_total_servicios + v_total_repuestos;
END$$

-- ------------------------------------------------------------
-- fn_obtener_stock_disponible
-- Retorna el stock actual de un repuesto.
-- ------------------------------------------------------------
DROP FUNCTION IF EXISTS fn_obtener_stock_disponible$$
CREATE FUNCTION fn_obtener_stock_disponible(p_repuesto_id INT)
RETURNS INT
DETERMINISTIC
READS SQL DATA
BEGIN
    DECLARE v_stock INT DEFAULT 0;

    SELECT stock_actual INTO v_stock
      FROM repuestos
     WHERE repuesto_id = p_repuesto_id;

    RETURN COALESCE(v_stock, 0);
END$$

-- ------------------------------------------------------------
-- fn_dias_ultima_revision_cda
-- Retorna cuantos dias han pasado desde la ultima revision CDA.
-- Retorna -1 si el vehiculo no tiene revision registrada.
-- ------------------------------------------------------------
DROP FUNCTION IF EXISTS fn_dias_ultima_revision_cda$$
CREATE FUNCTION fn_dias_ultima_revision_cda(p_vehiculo_id INT)
RETURNS INT
DETERMINISTIC
READS SQL DATA
BEGIN
    DECLARE v_ultima_fecha DATE;
    DECLARE v_dias         INT DEFAULT -1;

    SELECT MAX(fecha_revision) INTO v_ultima_fecha
      FROM revisiones_cda
     WHERE vehiculo_id = p_vehiculo_id;

    IF v_ultima_fecha IS NOT NULL THEN
        SET v_dias = DATEDIFF(CURRENT_DATE, v_ultima_fecha);
    END IF;

    RETURN v_dias;
END$$

-- ============================================================
-- PROCEDIMIENTOS ALMACENADOS
-- ============================================================

-- ------------------------------------------------------------
-- sp_registrar_cliente_vehiculo
-- Registra un cliente y su vehiculo en una sola transaccion.
-- Si alguna parte falla se hace ROLLBACK completo.
-- ------------------------------------------------------------
DROP PROCEDURE IF EXISTS sp_registrar_cliente_vehiculo$$
CREATE PROCEDURE sp_registrar_cliente_vehiculo(
    IN  p_cedula      VARCHAR(20),
    IN  p_nombre      VARCHAR(80),
    IN  p_apellido    VARCHAR(80),
    IN  p_telefono    VARCHAR(20),
    IN  p_email       VARCHAR(120),
    IN  p_direccion   VARCHAR(200),
    IN  p_placa       VARCHAR(10),
    IN  p_marca       VARCHAR(50),
    IN  p_modelo      VARCHAR(60),
    IN  p_anio        SMALLINT,
    IN  p_color       VARCHAR(30),
    IN  p_cilindraje  DECIMAL(5,1),
    IN  p_km_actual   INT,
    OUT p_cliente_id  INT,
    OUT p_vehiculo_id INT,
    OUT p_mensaje     VARCHAR(300)
)
BEGIN
    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        ROLLBACK;
        SET p_cliente_id  = -1;
        SET p_vehiculo_id = -1;
        GET DIAGNOSTICS CONDITION 1 p_mensaje = MESSAGE_TEXT;
    END;

    START TRANSACTION;

    -- Verificar cedula duplicada
    IF EXISTS (SELECT 1 FROM clientes WHERE cedula = p_cedula) THEN
        SIGNAL SQLSTATE '45000'
            SET MESSAGE_TEXT = 'Ya existe un cliente con esa cedula.';
    END IF;

    -- Verificar placa duplicada
    IF EXISTS (SELECT 1 FROM vehiculos WHERE placa = p_placa) THEN
        SIGNAL SQLSTATE '45000'
            SET MESSAGE_TEXT = 'Ya existe un vehiculo con esa placa.';
    END IF;

    INSERT INTO clientes (cedula, nombre, apellido, telefono, email, direccion)
    VALUES (p_cedula, p_nombre, p_apellido, p_telefono, p_email, p_direccion);
    SET p_cliente_id = LAST_INSERT_ID();

    INSERT INTO vehiculos (cliente_id, placa, marca, modelo, anio, color, cilindraje, km_actual)
    VALUES (p_cliente_id, p_placa, p_marca, p_modelo, p_anio, p_color, p_cilindraje, p_km_actual);
    SET p_vehiculo_id = LAST_INSERT_ID();

    COMMIT;
    SET p_mensaje = 'OK';
END$$

-- ------------------------------------------------------------
-- sp_registrar_orden_trabajo
-- Crea una nueva orden de trabajo con validaciones.
-- ------------------------------------------------------------
DROP PROCEDURE IF EXISTS sp_registrar_orden_trabajo$$
CREATE PROCEDURE sp_registrar_orden_trabajo(
    IN  p_vehiculo_id       INT,
    IN  p_empleado_id       INT,
    IN  p_fecha_estimada    DATE,
    IN  p_km_ingreso        INT,
    IN  p_diagnostico       TEXT,
    OUT p_orden_id          INT,
    OUT p_mensaje           VARCHAR(300)
)
BEGIN
    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        ROLLBACK;
        SET p_orden_id = -1;
        GET DIAGNOSTICS CONDITION 1 p_mensaje = MESSAGE_TEXT;
    END;

    START TRANSACTION;

    -- Validar que el vehiculo exista
    IF NOT EXISTS (SELECT 1 FROM vehiculos WHERE vehiculo_id = p_vehiculo_id) THEN
        SIGNAL SQLSTATE '45000'
            SET MESSAGE_TEXT = 'El vehiculo especificado no existe.';
    END IF;

    -- Validar empleado activo
    IF NOT EXISTS (SELECT 1 FROM empleados WHERE empleado_id = p_empleado_id AND activo = 1) THEN
        SIGNAL SQLSTATE '45000'
            SET MESSAGE_TEXT = 'El empleado no existe o no esta activo.';
    END IF;

    -- Validar que no haya una orden EN_PROCESO para el mismo vehiculo
    IF EXISTS (
        SELECT 1 FROM ordenes_trabajo
         WHERE vehiculo_id = p_vehiculo_id
           AND estado IN ('PENDIENTE', 'EN_PROCESO')
    ) THEN
        SIGNAL SQLSTATE '45000'
            SET MESSAGE_TEXT = 'El vehiculo ya tiene una orden activa (PENDIENTE o EN_PROCESO).';
    END IF;

    INSERT INTO ordenes_trabajo (vehiculo_id, empleado_id, fecha_estimada, km_ingreso, diagnostico, estado)
    VALUES (p_vehiculo_id, p_empleado_id, p_fecha_estimada, p_km_ingreso, p_diagnostico, 'PENDIENTE');

    SET p_orden_id = LAST_INSERT_ID();

    -- Actualizar km del vehiculo si el nuevo km es mayor
    UPDATE vehiculos
       SET km_actual = p_km_ingreso
     WHERE vehiculo_id = p_vehiculo_id
       AND p_km_ingreso > km_actual;

    COMMIT;
    SET p_mensaje = 'OK';
END$$

-- ------------------------------------------------------------
-- sp_agregar_servicio_orden
-- Agrega un servicio a una orden existente.
-- ------------------------------------------------------------
DROP PROCEDURE IF EXISTS sp_agregar_servicio_orden$$
CREATE PROCEDURE sp_agregar_servicio_orden(
    IN  p_orden_id       INT,
    IN  p_servicio_id    INT,
    IN  p_cantidad       INT,
    IN  p_mano_obra      DECIMAL(12,2),
    OUT p_detalle_id     INT,
    OUT p_mensaje        VARCHAR(300)
)
BEGIN
    DECLARE v_precio_base DECIMAL(12,2);
    DECLARE v_subtotal    DECIMAL(12,2);

    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        ROLLBACK;
        SET p_detalle_id = -1;
        GET DIAGNOSTICS CONDITION 1 p_mensaje = MESSAGE_TEXT;
    END;

    START TRANSACTION;

    -- Validar orden activa
    IF NOT EXISTS (
        SELECT 1 FROM ordenes_trabajo
         WHERE orden_id = p_orden_id
           AND estado IN ('PENDIENTE', 'EN_PROCESO')
    ) THEN
        SIGNAL SQLSTATE '45000'
            SET MESSAGE_TEXT = 'La orden no existe o no esta en estado editable.';
    END IF;

    -- Obtener precio base del servicio
    SELECT precio_base INTO v_precio_base
      FROM servicios WHERE servicio_id = p_servicio_id;

    IF v_precio_base IS NULL THEN
        SIGNAL SQLSTATE '45000'
            SET MESSAGE_TEXT = 'El servicio especificado no existe.';
    END IF;

    SET v_subtotal = (v_precio_base * p_cantidad) + p_mano_obra;

    INSERT INTO detalle_servicios (orden_id, servicio_id, cantidad, precio_unitario, mano_obra, subtotal)
    VALUES (p_orden_id, p_servicio_id, p_cantidad, v_precio_base, p_mano_obra, v_subtotal);

    SET p_detalle_id = LAST_INSERT_ID();

    -- Cambiar estado a EN_PROCESO
    UPDATE ordenes_trabajo SET estado = 'EN_PROCESO'
     WHERE orden_id = p_orden_id AND estado = 'PENDIENTE';

    COMMIT;
    SET p_mensaje = 'OK';
END$$

-- ------------------------------------------------------------
-- sp_agregar_repuesto_orden
-- Agrega un repuesto a una orden. El trigger se encarga de
-- descontar el stock automaticamente.
-- ------------------------------------------------------------
DROP PROCEDURE IF EXISTS sp_agregar_repuesto_orden$$
CREATE PROCEDURE sp_agregar_repuesto_orden(
    IN  p_orden_id      INT,
    IN  p_repuesto_id   INT,
    IN  p_cantidad      INT,
    OUT p_detalle_id    INT,
    OUT p_mensaje       VARCHAR(300)
)
BEGIN
    DECLARE v_precio_venta DECIMAL(12,2);
    DECLARE v_stock_actual INT;
    DECLARE v_subtotal     DECIMAL(12,2);

    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        ROLLBACK;
        SET p_detalle_id = -1;
        GET DIAGNOSTICS CONDITION 1 p_mensaje = MESSAGE_TEXT;
    END;

    START TRANSACTION;

    -- Validar orden activa
    IF NOT EXISTS (
        SELECT 1 FROM ordenes_trabajo
         WHERE orden_id = p_orden_id
           AND estado IN ('PENDIENTE', 'EN_PROCESO')
    ) THEN
        SIGNAL SQLSTATE '45000'
            SET MESSAGE_TEXT = 'La orden no existe o no esta en estado editable.';
    END IF;

    -- Verificar stock (con lock para concurrencia)
    SELECT precio_venta, stock_actual
      INTO v_precio_venta, v_stock_actual
      FROM repuestos
     WHERE repuesto_id = p_repuesto_id
       FOR UPDATE;

    IF v_precio_venta IS NULL THEN
        SIGNAL SQLSTATE '45000'
            SET MESSAGE_TEXT = 'El repuesto especificado no existe.';
    END IF;

    IF v_stock_actual < p_cantidad THEN
        SIGNAL SQLSTATE '45000'
            SET MESSAGE_TEXT = 'Stock insuficiente para la cantidad solicitada.';
    END IF;

    SET v_subtotal = v_precio_venta * p_cantidad;

    INSERT INTO detalle_repuestos (orden_id, repuesto_id, cantidad, precio_unitario, subtotal)
    VALUES (p_orden_id, p_repuesto_id, p_cantidad, v_precio_venta, v_subtotal);

    SET p_detalle_id = LAST_INSERT_ID();

    COMMIT;
    SET p_mensaje = 'OK';
END$$

-- ------------------------------------------------------------
-- sp_cerrar_orden_y_facturar
-- Cierra una orden (estado COMPLETADO) y genera su factura.
-- Usa COMMIT/ROLLBACK para garantizar atomicidad.
-- ------------------------------------------------------------
DROP PROCEDURE IF EXISTS sp_cerrar_orden_y_facturar$$
CREATE PROCEDURE sp_cerrar_orden_y_facturar(
    IN  p_orden_id          INT,
    IN  p_observaciones     TEXT,
    IN  p_metodo_pago       ENUM('EFECTIVO','TARJETA','TRANSFERENCIA'),
    OUT p_factura_id        INT,
    OUT p_total             DECIMAL(12,2),
    OUT p_mensaje           VARCHAR(300)
)
BEGIN
    DECLARE v_sub_servicios DECIMAL(12,2);
    DECLARE v_sub_repuestos DECIMAL(12,2);
    DECLARE v_iva           DECIMAL(12,2);
    DECLARE v_total         DECIMAL(12,2);
    DECLARE c_iva_pct       DECIMAL(5,4) DEFAULT 0.19;

    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        ROLLBACK;
        SET p_factura_id = -1;
        SET p_total      = 0;
        GET DIAGNOSTICS CONDITION 1 p_mensaje = MESSAGE_TEXT;
    END;

    START TRANSACTION;

    -- Validar que la orden existe y esta en proceso
    IF NOT EXISTS (
        SELECT 1 FROM ordenes_trabajo
         WHERE orden_id = p_orden_id
           AND estado IN ('PENDIENTE', 'EN_PROCESO')
    ) THEN
        SIGNAL SQLSTATE '45000'
            SET MESSAGE_TEXT = 'La orden no existe o ya fue cerrada/cancelada.';
    END IF;

    -- Calcular totales
    SELECT COALESCE(SUM(subtotal), 0) INTO v_sub_servicios
      FROM detalle_servicios WHERE orden_id = p_orden_id;

    SELECT COALESCE(SUM(subtotal), 0) INTO v_sub_repuestos
      FROM detalle_repuestos WHERE orden_id = p_orden_id;

    SET v_iva   = (v_sub_servicios + v_sub_repuestos) * c_iva_pct;
    SET v_total = v_sub_servicios + v_sub_repuestos + v_iva;

    -- Actualizar orden
    UPDATE ordenes_trabajo
       SET estado        = 'COMPLETADO',
           fecha_salida  = NOW(),
           observaciones = p_observaciones
     WHERE orden_id = p_orden_id;

    -- Crear factura
    INSERT INTO facturas (orden_id, subtotal_servicios, subtotal_repuestos, iva, total, estado_pago, metodo_pago)
    VALUES (p_orden_id, v_sub_servicios, v_sub_repuestos, v_iva, v_total, 'PENDIENTE', p_metodo_pago);

    SET p_factura_id = LAST_INSERT_ID();
    SET p_total      = v_total;

    COMMIT;
    SET p_mensaje = 'OK';
END$$

-- ------------------------------------------------------------
-- sp_registrar_revision_cda
-- Registra una revision CDA para un vehiculo.
-- ------------------------------------------------------------
DROP PROCEDURE IF EXISTS sp_registrar_revision_cda$$
CREATE PROCEDURE sp_registrar_revision_cda(
    IN  p_vehiculo_id   INT,
    IN  p_empleado_id   INT,
    IN  p_resultado     ENUM('APROBADO','REPROBADO','APROBADO_CON_OBSERVACIONES'),
    IN  p_observaciones TEXT,
    OUT p_revision_id   INT,
    OUT p_mensaje       VARCHAR(300)
)
BEGIN
    DECLARE v_vencimiento DATE;

    DECLARE EXIT HANDLER FOR SQLEXCEPTION
    BEGIN
        ROLLBACK;
        SET p_revision_id = -1;
        GET DIAGNOSTICS CONDITION 1 p_mensaje = MESSAGE_TEXT;
    END;

    START TRANSACTION;

    IF NOT EXISTS (SELECT 1 FROM vehiculos WHERE vehiculo_id = p_vehiculo_id) THEN
        SIGNAL SQLSTATE '45000'
            SET MESSAGE_TEXT = 'El vehiculo no existe.';
    END IF;

    -- Fecha de vencimiento: 1 año si aprobado, NULL si reprobado
    IF p_resultado = 'REPROBADO' THEN
        SET v_vencimiento = NULL;
    ELSE
        SET v_vencimiento = DATE_ADD(CURRENT_DATE, INTERVAL 1 YEAR);
    END IF;

    INSERT INTO revisiones_cda (vehiculo_id, empleado_id, resultado, observaciones, fecha_vencimiento)
    VALUES (p_vehiculo_id, p_empleado_id, p_resultado, p_observaciones, v_vencimiento);

    SET p_revision_id = LAST_INSERT_ID();

    COMMIT;
    SET p_mensaje = 'OK';
END$$

-- ------------------------------------------------------------
-- sp_reporte_ordenes_por_periodo
-- Reporte de ordenes entre dos fechas con totales.
-- ------------------------------------------------------------
DROP PROCEDURE IF EXISTS sp_reporte_ordenes_por_periodo$$
CREATE PROCEDURE sp_reporte_ordenes_por_periodo(
    IN p_fecha_inicio DATE,
    IN p_fecha_fin    DATE
)
BEGIN
    SELECT
        o.orden_id,
        o.fecha_ingreso,
        o.estado,
        CONCAT(c.nombre, ' ', c.apellido)   AS cliente,
        v.placa,
        CONCAT(v.marca, ' ', v.modelo)      AS vehiculo,
        CONCAT(e.nombre, ' ', e.apellido)   AS responsable,
        fn_calcular_total_orden(o.orden_id) AS total_orden,
        f.total                             AS total_facturado,
        f.estado_pago
    FROM ordenes_trabajo o
    JOIN vehiculos  v ON v.vehiculo_id = o.vehiculo_id
    JOIN clientes   c ON c.cliente_id  = v.cliente_id
    JOIN empleados  e ON e.empleado_id = o.empleado_id
    LEFT JOIN facturas f ON f.orden_id = o.orden_id
    WHERE DATE(o.fecha_ingreso) BETWEEN p_fecha_inicio AND p_fecha_fin
    ORDER BY o.fecha_ingreso DESC;
END$$

-- ============================================================
-- TRIGGERS
-- ============================================================

-- ------------------------------------------------------------
-- trg_before_detalle_repuesto_insert
-- Valida que haya stock suficiente antes de insertar.
-- ------------------------------------------------------------
DROP TRIGGER IF EXISTS trg_before_detalle_repuesto_insert$$
CREATE TRIGGER trg_before_detalle_repuesto_insert
BEFORE INSERT ON detalle_repuestos
FOR EACH ROW
BEGIN
    DECLARE v_stock INT;

    SELECT stock_actual INTO v_stock
      FROM repuestos
     WHERE repuesto_id = NEW.repuesto_id;

    IF v_stock IS NULL THEN
        SIGNAL SQLSTATE '45000'
            SET MESSAGE_TEXT = 'El repuesto no existe.';
    END IF;

    IF v_stock < NEW.cantidad THEN
        SIGNAL SQLSTATE '45000'
            SET MESSAGE_TEXT = 'Stock insuficiente: no hay suficientes unidades del repuesto.';
    END IF;
END$$

-- ------------------------------------------------------------
-- trg_after_detalle_repuesto_insert
-- Descuenta el stock del repuesto al insertarse el detalle.
-- ------------------------------------------------------------
DROP TRIGGER IF EXISTS trg_after_detalle_repuesto_insert$$
CREATE TRIGGER trg_after_detalle_repuesto_insert
AFTER INSERT ON detalle_repuestos
FOR EACH ROW
BEGIN
    UPDATE repuestos
       SET stock_actual = stock_actual - NEW.cantidad
     WHERE repuesto_id = NEW.repuesto_id;
END$$

-- ------------------------------------------------------------
-- trg_after_detalle_repuesto_delete
-- Restaura el stock si se elimina un detalle de repuesto.
-- ------------------------------------------------------------
DROP TRIGGER IF EXISTS trg_after_detalle_repuesto_delete$$
CREATE TRIGGER trg_after_detalle_repuesto_delete
AFTER DELETE ON detalle_repuestos
FOR EACH ROW
BEGIN
    UPDATE repuestos
       SET stock_actual = stock_actual + OLD.cantidad
     WHERE repuesto_id = OLD.repuesto_id;
END$$

-- ------------------------------------------------------------
-- trg_after_orden_update_completado
-- Al completar una orden, registra la fecha_salida si no
-- fue asignada manualmente.
-- ------------------------------------------------------------
DROP TRIGGER IF EXISTS trg_after_orden_update_completado$$
CREATE TRIGGER trg_after_orden_update_completado
BEFORE UPDATE ON ordenes_trabajo
FOR EACH ROW
BEGIN
    IF NEW.estado = 'COMPLETADO' AND OLD.estado != 'COMPLETADO' THEN
        IF NEW.fecha_salida IS NULL THEN
            SET NEW.fecha_salida = NOW();
        END IF;
    END IF;

    -- Evitar retroceder a estados anteriores desde COMPLETADO/ENTREGADO
    IF OLD.estado IN ('COMPLETADO', 'ENTREGADO') AND NEW.estado = 'PENDIENTE' THEN
        SIGNAL SQLSTATE '45000'
            SET MESSAGE_TEXT = 'No se puede revertir una orden ya completada a PENDIENTE.';
    END IF;
END$$

DELIMITER ;

-- ============================================================
-- VISTAS UTILES PARA REPORTES
-- ============================================================

CREATE OR REPLACE VIEW v_ordenes_detalladas AS
SELECT
    o.orden_id,
    o.fecha_ingreso,
    o.fecha_estimada,
    o.fecha_salida,
    o.estado,
    o.km_ingreso,
    o.diagnostico,
    c.cliente_id,
    CONCAT(c.nombre, ' ', c.apellido) AS cliente,
    c.cedula                          AS cedula_cliente,
    c.telefono                        AS telefono_cliente,
    v.vehiculo_id,
    v.placa,
    CONCAT(v.marca, ' ', v.modelo, ' ', v.anio) AS vehiculo,
    CONCAT(e.nombre, ' ', e.apellido) AS responsable,
    fn_calcular_total_orden(o.orden_id) AS total_estimado,
    f.factura_id,
    f.total                           AS total_facturado,
    f.estado_pago,
    f.metodo_pago
FROM ordenes_trabajo o
JOIN vehiculos  v ON v.vehiculo_id = o.vehiculo_id
JOIN clientes   c ON c.cliente_id  = v.cliente_id
JOIN empleados  e ON e.empleado_id = o.empleado_id
LEFT JOIN facturas f ON f.orden_id = o.orden_id;

CREATE OR REPLACE VIEW v_repuestos_bajo_stock AS
SELECT
    r.repuesto_id,
    r.referencia,
    r.nombre,
    r.stock_actual,
    r.stock_minimo,
    (r.stock_minimo - r.stock_actual) AS unidades_requeridas,
    p.nombre AS proveedor,
    p.telefono AS tel_proveedor
FROM repuestos r
JOIN proveedores p ON p.proveedor_id = r.proveedor_id
WHERE r.stock_actual <= r.stock_minimo
ORDER BY unidades_requeridas DESC;

CREATE OR REPLACE VIEW v_vehiculos_con_cda_vencida AS
SELECT
    v.vehiculo_id,
    v.placa,
    CONCAT(v.marca, ' ', v.modelo) AS vehiculo,
    c.nombre AS cliente,
    c.telefono,
    MAX(r.fecha_revision)    AS ultima_revision,
    MAX(r.fecha_vencimiento) AS vencimiento_cda,
    DATEDIFF(CURRENT_DATE, MAX(r.fecha_vencimiento)) AS dias_vencida
FROM vehiculos vLista 
JOIN clientes c ON c.cliente_id = v.cliente_id
LEFT JOIN revisiones_cda r ON r.vehiculo_id = v.vehiculo_id
GROUP BY v.vehiculo_id, v.placa, v.marca, v.modelo, c.nombre, c.telefono
HAVING MAX(r.fecha_vencimiento) < CURRENT_DATE
    OR MAX(r.fecha_vencimiento) IS NULL;
