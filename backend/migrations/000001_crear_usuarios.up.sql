-- =============================================================================
-- Migración 000001: tabla de usuarios
-- -----------------------------------------------------------------------------
-- Un usuario es a la vez:
--   1) un registro que se gestiona desde el CRUD (nombre, email, teléfono...)
--   2) una cuenta que puede iniciar sesión (email + password_hash + rol)
-- =============================================================================

CREATE TABLE usuarios (
    id             SERIAL       PRIMARY KEY,
    nombre         VARCHAR(100) NOT NULL,
    apellido       VARCHAR(100) NOT NULL,
    -- El email identifica al usuario al iniciar sesión, por eso es único.
    email          VARCHAR(150) NOT NULL UNIQUE,
    telefono       VARCHAR(20),
    -- Nunca se guarda la contraseña en texto plano: solo su hash bcrypt.
    password_hash  VARCHAR(255) NOT NULL,
    -- 'admin' gestiona todo; 'usuario' solo ve y actualiza lo suyo.
    rol            VARCHAR(10)  NOT NULL DEFAULT 'usuario'
                   CONSTRAINT usuarios_rol_valido CHECK (rol IN ('admin', 'usuario')),
    creado_en      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    actualizado_en TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- -----------------------------------------------------------------------------
-- Función reutilizable: actualiza la columna "actualizado_en" en cada UPDATE.
-- Así la base de datos lo garantiza aunque el backend olvide hacerlo.
-- -----------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION fijar_actualizado_en()
RETURNS TRIGGER AS $$
BEGIN
    NEW.actualizado_en = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER usuarios_actualizado_en
    BEFORE UPDATE ON usuarios
    FOR EACH ROW
    EXECUTE FUNCTION fijar_actualizado_en();
