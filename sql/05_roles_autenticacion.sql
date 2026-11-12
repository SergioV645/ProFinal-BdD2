-- ============================================================
-- AURUM MOTORS - Migracion 05: Autenticacion y roles
-- Ejecutar UNA SOLA VEZ sobre una base taller_vehiculos existente.
-- No elimina ni modifica los registros operativos.
-- ============================================================
USE taller_vehiculos;

CREATE TABLE IF NOT EXISTS roles (
    rol_id INT NOT NULL AUTO_INCREMENT,
    nombre VARCHAR(30) NOT NULL,
    descripcion VARCHAR(200) NULL,
    PRIMARY KEY (rol_id),
    UNIQUE KEY uq_roles_nombre (nombre)
) ENGINE=InnoDB;

INSERT INTO roles (nombre, descripcion) VALUES
    ('ADMINISTRADOR', 'Acceso completo al sistema'),
    ('EMPLEADO', 'Acceso operativo sin permisos de eliminacion ni administracion')
ON DUPLICATE KEY UPDATE descripcion=VALUES(descripcion);

CREATE TABLE IF NOT EXISTS usuarios (
    usuario_id INT NOT NULL AUTO_INCREMENT,
    username VARCHAR(60) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    rol_id INT NOT NULL,
    empleado_id INT NULL,
    activo TINYINT(1) NOT NULL DEFAULT 1,
    creado_en DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    actualizado_en DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (usuario_id),
    UNIQUE KEY uq_usuarios_username (username),
    UNIQUE KEY uq_usuarios_empleado (empleado_id),
    CONSTRAINT fk_usuarios_rol FOREIGN KEY (rol_id)
        REFERENCES roles (rol_id) ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_usuarios_empleado FOREIGN KEY (empleado_id)
        REFERENCES empleados (empleado_id) ON UPDATE CASCADE ON DELETE RESTRICT
) ENGINE=InnoDB;

SELECT * FROM roles;
SELECT usuario_id, username, rol_id, empleado_id, activo, creado_en FROM usuarios;
