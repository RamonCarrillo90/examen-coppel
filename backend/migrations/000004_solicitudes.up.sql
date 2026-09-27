-- Solicitudes de los usuarios para quedarse con una tarea sin asignar.
-- El admin las aprueba o rechaza.
CREATE TABLE solicitudes (
    id          SERIAL PRIMARY KEY,
    tarea_id    INT NOT NULL REFERENCES tareas(id) ON DELETE CASCADE,
    usuario_id  INT NOT NULL REFERENCES usuarios(id) ON DELETE CASCADE,
    estatus     VARCHAR(20) NOT NULL DEFAULT 'pendiente'
                CHECK (estatus IN ('pendiente', 'aprobada', 'rechazada')),
    creado_en   TIMESTAMPTZ NOT NULL DEFAULT now(),
    resuelto_en TIMESTAMPTZ
);

-- Un usuario no puede tener dos solicitudes PENDIENTES para la misma tarea.
-- (Índice único PARCIAL: solo cuenta las filas con estatus 'pendiente'.)
CREATE UNIQUE INDEX solicitudes_una_pendiente
    ON solicitudes (tarea_id, usuario_id)
    WHERE estatus = 'pendiente';