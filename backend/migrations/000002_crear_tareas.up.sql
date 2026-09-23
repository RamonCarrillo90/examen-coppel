-- =============================================================================
-- Migración 000002: tabla de tareas
-- -----------------------------------------------------------------------------
-- Relación: un usuario tiene muchas tareas (1 a N).
-- Cada tarea pertenece exactamente a un usuario (usuario_id NOT NULL).
-- =============================================================================

CREATE TABLE tareas (
    id             SERIAL       PRIMARY KEY,
    -- ON DELETE CASCADE: al eliminar un usuario, PostgreSQL elimina sus tareas.
    -- Cumple el requisito del examen directamente en la base de datos.
    usuario_id     INT          NOT NULL
                   REFERENCES usuarios(id) ON DELETE CASCADE,
    titulo         VARCHAR(150) NOT NULL,
    descripcion    TEXT,
    fecha_limite   DATE,
    estatus        VARCHAR(20)  NOT NULL DEFAULT 'pendiente'
                   CONSTRAINT tareas_estatus_valido
                   CHECK (estatus IN ('pendiente', 'en_progreso', 'completada')),
    creado_en      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    actualizado_en TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- PostgreSQL NO crea índices automáticamente para llaves foráneas.
-- Este índice acelera "dame las tareas del usuario X", la consulta más común.
CREATE INDEX idx_tareas_usuario_id ON tareas (usuario_id);

-- Reutiliza la función creada en la migración 000001.
CREATE TRIGGER tareas_actualizado_en
    BEFORE UPDATE ON tareas
    FOR EACH ROW
    EXECUTE FUNCTION fijar_actualizado_en();
