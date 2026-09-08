-- ============================================================
-- Script 04: Recrear vistas (correr si el script 03 falló)
-- ============================================================

USE taller_vehiculos;

-- Vista ordenes detalladas
CREATE OR REPLACE VIEW v_ordenes_detalladas AS
SELECT
    o.orden_id,
    o.fecha_ingreso,
    o.fecha_estimada,
    o.fecha_salida,
    o.estado,
    o.km_ingreso,
    o.diagnostico,
    o.observaciones,
    c.cliente_id,
    CONCAT(c.nombre, ' ', c.apellido)           AS cliente,
    c.cedula                                    AS cedula_cliente,
    c.telefono                                  AS telefono_cliente,
    v.vehiculo_id,
    v.placa,
    CONCAT(v.marca, ' ', v.modelo, ' ', v.anio) AS vehiculo,
    CONCAT(e.nombre, ' ', e.apellido)           AS responsable,
    f.factura_id,
    f.total                                     AS total_facturado,
    f.estado_pago,
    f.metodo_pago
FROM ordenes_trabajo o
JOIN vehiculos  v ON v.vehiculo_id = o.vehiculo_id
JOIN clientes   c ON c.cliente_id  = v.cliente_id
JOIN empleados  e ON e.empleado_id = o.empleado_id
LEFT JOIN facturas f ON f.orden_id = o.orden_id;

-- Vista repuestos bajo stock
CREATE OR REPLACE VIEW v_repuestos_bajo_stock AS
SELECT
    r.repuesto_id,
    r.referencia,
    r.nombre,
    r.stock_actual,
    r.stock_minimo,
    (r.stock_minimo - r.stock_actual) AS unidades_requeridas,
    p.nombre    AS proveedor,
    p.telefono  AS tel_proveedor
FROM repuestos r
JOIN proveedores p ON p.proveedor_id = r.proveedor_id
WHERE r.stock_actual <= r.stock_minimo
ORDER BY unidades_requeridas DESC;

-- Vista vehiculos con CDA vencida
CREATE OR REPLACE VIEW v_vehiculos_con_cda_vencida AS
SELECT
    v.vehiculo_id,
    v.placa,
    CONCAT(v.marca, ' ', v.modelo) AS vehiculo,
    c.nombre    AS cliente,
    c.telefono,
    MAX(r.fecha_revision)    AS ultima_revision,
    MAX(r.fecha_vencimiento) AS vencimiento_cda,
    DATEDIFF(CURRENT_DATE, MAX(r.fecha_vencimiento)) AS dias_vencida
FROM vehiculos v
JOIN clientes c ON c.cliente_id = v.cliente_id
LEFT JOIN revisiones_cda r ON r.vehiculo_id = v.vehiculo_id
GROUP BY v.vehiculo_id, v.placa, v.marca, v.modelo, c.nombre, c.telefono
HAVING MAX(r.fecha_vencimiento) < CURRENT_DATE
    OR MAX(r.fecha_vencimiento) IS NULL;

-- Verificar que quedaron creadas
SHOW FULL TABLES WHERE Table_type = 'VIEW';
