-- Para regresar, primero hay que eliminar las tareas sin usuario:
-- si no, SET NOT NULL fallaría.
DELETE FROM tareas WHERE usuario_id IS NULL;
ALTER TABLE tareas ALTER COLUMN usuario_id SET NOT NULL;