-- Una tarea puede existir sin usuario asignado (NULL = "sin asignar").
-- La llave foránea y el ON DELETE CASCADE no cambian.
ALTER TABLE tareas ALTER COLUMN usuario_id DROP NOT NULL;