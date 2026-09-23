-- Revierte la migración 000002.
DROP TRIGGER IF EXISTS tareas_actualizado_en ON tareas;
DROP TABLE IF EXISTS tareas;
