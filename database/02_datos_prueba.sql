-- ============================================================
-- TALLER DE VEHICULOS & CDA
-- Script 02: Datos de prueba
-- ============================================================

USE taller_vehiculos;

-- -----------------------------------------------------------
-- CARGOS
-- -----------------------------------------------------------
INSERT INTO cargos (nombre, descripcion) VALUES
    ('Mecanico General',      'Reparacion y mantenimiento general de vehiculos'),
    ('Electricista Automotriz','Diagnostico y reparacion de sistemas electricos'),
    ('Tecnico CDA',           'Inspector certificado para revision CDA'),
    ('Jefe de Taller',        'Supervision y coordinacion del taller'),
    ('Asesor Comercial',      'Atencion al cliente y presupuestos');

-- -----------------------------------------------------------
-- CATEGORIAS DE SERVICIO
-- -----------------------------------------------------------
INSERT INTO categorias_servicio (nombre) VALUES
    ('Mantenimiento Preventivo'),
    ('Mecanica General'),
    ('Electricidad Automotriz'),
    ('Carroceria y Pintura'),
    ('Revision CDA');

-- -----------------------------------------------------------
-- CLIENTES
-- -----------------------------------------------------------
INSERT INTO clientes (cedula, nombre, apellido, telefono, email, direccion) VALUES
    ('12345678',  'Carlos',    'Ramirez',    '3101234567', 'carlos.ramirez@email.com',    'Calle 45 #12-30, Bogota'),
    ('23456789',  'Maria',     'Lopez',      '3209876543', 'maria.lopez@email.com',       'Carrera 7 #80-15, Bogota'),
    ('34567890',  'Juan',      'Perez',      '3151234000', 'juan.perez@email.com',        'Av. El Dorado #68-40, Bogota'),
    ('45678901',  'Ana',       'Torres',     '3005678901', 'ana.torres@email.com',        'Calle 100 #50-20, Bogota'),
    ('56789012',  'Luis',      'Gomez',      '3124567890', 'luis.gomez@email.com',        'Carrera 15 #90-10, Bogota'),
    ('67890123',  'Sandra',    'Martinez',   '3187654321', 'sandra.m@email.com',          'Calle 72 #25-08, Bogota'),
    ('78901234',  'Roberto',   'Hernandez',  '3209087654', 'roberto.h@email.com',         'Cra 30 #45-55, Bogota'),
    ('89012345',  'Claudia',   'Diaz',       '3115432109', 'claudia.diaz@email.com',      'Calle 13 #100-60, Bogota');

-- -----------------------------------------------------------
-- VEHICULOS
-- -----------------------------------------------------------
INSERT INTO vehiculos (cliente_id, placa, marca, modelo, anio, color, cilindraje, km_actual) VALUES
    (1, 'ABC123', 'Chevrolet', 'Spark',    2020, 'Blanco',  1.0,  45000),
    (1, 'XYZ789', 'Mazda',    'CX-5',     2022, 'Gris',    2.0,  18000),
    (2, 'DEF456', 'Renault',  'Logan',    2019, 'Rojo',    1.6,  62000),
    (3, 'GHI321', 'Toyota',   'Corolla',  2021, 'Negro',   1.8,  30000),
    (4, 'JKL654', 'Ford',     'Explorer', 2018, 'Azul',    3.5,  88000),
    (5, 'MNO987', 'Nissan',   'Sentra',   2023, 'Plateado',1.8,   5000),
    (6, 'PQR147', 'Kia',      'Picanto',  2017, 'Verde',   1.2,  110000),
    (7, 'STU258', 'Hyundai',  'Tucson',   2020, 'Blanco',  2.0,  40000),
    (8, 'VWX369', 'Volkswagen','Polo',    2021, 'Rojo',    1.6,  25000);

-- -----------------------------------------------------------
-- EMPLEADOS
-- -----------------------------------------------------------
INSERT INTO empleados (cargo_id, cedula, nombre, apellido, telefono, salario, fecha_ingreso) VALUES
    (4, '11111111', 'Pedro',   'Vargas',    '3100001111', 4500000, '2018-03-01'),
    (1, '22222222', 'Diego',   'Castillo',  '3200002222', 2800000, '2019-07-15'),
    (1, '33333333', 'Andres',  'Moreno',    '3100003333', 2800000, '2020-01-10'),
    (2, '44444444', 'Felipe',  'Reyes',     '3200004444', 3200000, '2019-11-20'),
    (3, '55555555', 'Monica',  'Suarez',    '3100005555', 3000000, '2021-04-05'),
    (5, '66666666', 'Liliana', 'Ortiz',     '3200006666', 2500000, '2022-08-01');

-- -----------------------------------------------------------
-- PROVEEDORES
-- -----------------------------------------------------------
INSERT INTO proveedores (nit, nombre, telefono, email, ciudad) VALUES
    ('900111111-1', 'Autopartes Colombia S.A.',  '6014001111', 'ventas@autopartes.com',  'Bogota'),
    ('900222222-2', 'Repuestos del Norte Ltda.', '6017002222', 'info@repnorte.com',      'Medellin'),
    ('900333333-3', 'Distribuidora Motorica',    '6014503333', 'pedidos@motorica.com',   'Bogota'),
    ('900444444-4', 'TecniRepuestos SAS',        '6012504444', 'ventas@tecnirepuestos.co','Cali');

-- -----------------------------------------------------------
-- REPUESTOS
-- -----------------------------------------------------------
INSERT INTO repuestos (proveedor_id, referencia, nombre, descripcion, precio_compra, precio_venta, stock_actual, stock_minimo) VALUES
    (1, 'ACE-001', 'Aceite 5W30 1L',          'Aceite sintetico motor 5W30',            18000,  25000, 50, 10),
    (1, 'FIL-001', 'Filtro de Aceite',         'Filtro de aceite universal',             12000,  18000, 30,  5),
    (1, 'FIL-002', 'Filtro de Aire',           'Filtro de aire para motor',              22000,  32000, 20,  5),
    (2, 'PAD-001', 'Pastillas de Freno Del.',  'Juego pastillas freno delantero',        55000,  85000, 15,  3),
    (2, 'PAD-002', 'Pastillas de Freno Tra.',  'Juego pastillas freno trasero',          48000,  75000, 12,  3),
    (3, 'BAT-001', 'Bateria 60Ah',             'Bateria libre mantenimiento 60Ah',      180000, 260000,  8,  2),
    (3, 'BUJ-001', 'Bujias NGK x4',            'Set 4 bujias NGK iridium',              65000,  95000, 25,  5),
    (4, 'COR-001', 'Correa Distribucion',      'Correa de distribucion reforzada',       90000, 140000, 10,  2),
    (4, 'AMO-001', 'Amortiguador Delantero',   'Amortiguador delantero gas-presion',   120000, 195000,  6,  2),
    (1, 'LIQ-001', 'Liquido de Frenos DOT4',   'Liquido de frenos DOT4 500ml',          15000,  22000, 40,  8);

-- -----------------------------------------------------------
-- SERVICIOS
-- -----------------------------------------------------------
INSERT INTO servicios (categoria_id, nombre, descripcion, precio_base) VALUES
    (1, 'Cambio de Aceite',                'Cambio de aceite y filtro de aceite',                  35000),
    (1, 'Mantenimiento 5.000 km',          'Revision basica: aceite, filtros, niveles',            80000),
    (1, 'Mantenimiento 20.000 km',         'Revision completa: aceite, bujias, filtros, frenos',  180000),
    (2, 'Revision de Frenos',              'Inspeccion y ajuste sistema de frenos',                45000),
    (2, 'Cambio de Pastillas de Freno',    'Cambio juego de pastillas delanteras y traseras',      70000),
    (2, 'Alineacion y Balanceo',           'Alineacion 4 ruedas y balanceo',                       60000),
    (2, 'Cambio de Correa de Distribucion','Cambio correa con kit tensor',                        250000),
    (3, 'Diagnostico Electrico',           'Escaneo y diagnostico electronico',                    55000),
    (3, 'Cambio de Bateria',               'Cambio e instalacion de bateria nueva',                40000),
    (3, 'Reparacion Sistema de Carga',     'Revision y reparacion alternador/arranque',           120000),
    (5, 'Revision CDA Basica',             'Revision tecnico-mecanica para CDA',                   80000),
    (5, 'Revision CDA Completa',           'Revision integral para renovacion CDA',               120000);

-- -----------------------------------------------------------
-- ORDENES DE TRABAJO
-- -----------------------------------------------------------
INSERT INTO ordenes_trabajo (vehiculo_id, empleado_id, fecha_ingreso, fecha_estimada, fecha_salida, estado, km_ingreso, diagnostico, observaciones) VALUES
    (1, 2, '2026-08-01 08:00:00', '2026-08-01', '2026-08-01 11:30:00', 'ENTREGADO',   45000, 'Mantenimiento programado 5.000 km', 'Completado sin novedad'),
    (3, 3, '2026-08-05 09:00:00', '2026-08-06', '2026-08-06 16:00:00', 'ENTREGADO',   62000, 'Falla frenos delanteros, ruido al frenar', 'Se cambiaron pastillas y se purgo sistema'),
    (5, 2, '2026-08-10 10:00:00', '2026-08-12', '2026-08-12 14:00:00', 'ENTREGADO',   88000, 'Mantenimiento 20.000 km + correa distribucion', 'Servicio mayor completado'),
    (4, 4, '2026-08-15 11:00:00', '2026-08-15', '2026-08-15 17:00:00', 'ENTREGADO',   30000, 'Diagnostico electrico, luz check encendida', 'Se reemplazo sensor O2'),
    (6, 3, '2026-09-01 08:30:00', '2026-09-02', NULL,                  'EN_PROCESO',   5000, 'Revision general vehiculo nuevo', NULL),
    (7, 2, '2026-09-03 09:00:00', '2026-09-04', NULL,                  'PENDIENTE',  110000, 'Revision integral + posible cambio amortiguadores', NULL),
    (2, 5, '2026-09-05 10:00:00', '2026-09-05', NULL,                  'EN_PROCESO',  18000, 'Revision CDA para renovacion', NULL),
    (9, 3, '2026-09-06 08:00:00', '2026-09-07', NULL,                  'PENDIENTE',   25000, 'Alineacion, balanceo y frenos', NULL);

-- -----------------------------------------------------------
-- DETALLE_SERVICIOS
-- -----------------------------------------------------------
INSERT INTO detalle_servicios (orden_id, servicio_id, cantidad, precio_unitario, mano_obra, subtotal) VALUES
    -- Orden 1: Cambio aceite
    (1, 1, 1, 35000, 15000, 50000),
    -- Orden 2: Frenos
    (2, 4, 1, 45000, 20000, 65000),
    (2, 5, 1, 70000, 30000, 100000),
    -- Orden 3: Mantenimiento mayor + correa
    (3, 3, 1, 180000, 50000, 230000),
    (3, 7, 1, 250000, 80000, 330000),
    -- Orden 4: Diagnostico electrico
    (4, 8, 1, 55000, 25000, 80000),
    -- Orden 5: Revision general
    (5, 2, 1, 80000, 20000, 100000),
    -- Orden 7: CDA
    (7, 12, 1, 120000, 0, 120000),
    -- Orden 8: Alineacion y frenos
    (8, 6, 1, 60000, 20000, 80000),
    (8, 4, 1, 45000, 20000, 65000);

-- -----------------------------------------------------------
-- DETALLE_REPUESTOS
-- -----------------------------------------------------------
INSERT INTO detalle_repuestos (orden_id, repuesto_id, cantidad, precio_unitario, subtotal) VALUES
    -- Orden 1: Aceite (4L) + filtro aceite
    (1, 1, 4, 25000, 100000),
    (1, 2, 1, 18000,  18000),
    -- Orden 2: Pastillas freno del + tra
    (2, 4, 1, 85000,  85000),
    (2, 5, 1, 75000,  75000),
    -- Orden 3: Aceite (4L) + filtro + bujias + correa
    (3, 1, 4, 25000, 100000),
    (3, 2, 1, 18000,  18000),
    (3, 7, 1, 95000,  95000),
    (3, 8, 1, 140000, 140000),
    -- Orden 5: Aceite (4L) + filtro aire
    (5, 1, 4, 25000, 100000),
    (5, 3, 1, 32000,  32000),
    -- Orden 8: Liquido de frenos
    (8, 10, 1, 22000, 22000);

-- -----------------------------------------------------------
-- FACTURAS (solo para ordenes ENTREGADAS)
-- -----------------------------------------------------------
INSERT INTO facturas (orden_id, fecha_emision, subtotal_servicios, subtotal_repuestos, iva, total, estado_pago, metodo_pago) VALUES
    (1, '2026-08-01 12:00:00',  50000, 118000, 31844,  199844, 'PAGADO', 'EFECTIVO'),
    (2, '2026-08-06 16:30:00', 165000, 160000, 62100,  387100, 'PAGADO', 'TARJETA'),
    (3, '2026-08-12 15:00:00', 560000, 353000, 173070, 1086070,'PAGADO', 'TRANSFERENCIA'),
    (4, '2026-08-15 17:30:00',  80000,       0, 15200,   95200,'PAGADO', 'EFECTIVO');

-- -----------------------------------------------------------
-- REVISIONES CDA
-- -----------------------------------------------------------
INSERT INTO revisiones_cda (vehiculo_id, empleado_id, fecha_revision, resultado, observaciones, fecha_vencimiento) VALUES
    (1, 5, '2026-01-10', 'APROBADO',                   'Vehiculo en optimas condiciones',              '2027-01-10'),
    (3, 5, '2025-11-20', 'APROBADO_CON_OBSERVACIONES', 'Frenos delanteros en limite, programar cambio','2026-11-20'),
    (5, 5, '2025-06-15', 'REPROBADO',                  'Amortiguadores en mal estado, frenos desgastados', NULL),
    (4, 5, '2026-03-01', 'APROBADO',                   'Todos los sistemas en buen estado',            '2027-03-01'),
    (2, 5, '2026-09-05', 'APROBADO',                   'Vehiculo nuevo, sin observaciones',            '2027-09-05');

SELECT * FROM cargos;
SELECT * FROM categorias_servicio;
SELECT * FROM clientes;
SELECT * FROM vehiculos;
SELECT * FROM empleados;
SELECT * FROM proveedores;
SELECT * FROM repuestos;
SELECT * FROM servicios;
SELECT * FROM ordenes_trabajo;
SELECT * FROM detalle_servicios;
SELECT * FROM detalle_repuestos;
SELECT * FROM facturas;
SELECT * FROM revisiones_cda;

SHOW TABLES;