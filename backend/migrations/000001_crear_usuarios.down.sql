-- Revierte la migración 000001 (en orden inverso a como se creó).
DROP TRIGGER IF EXISTS usuarios_actualizado_en ON usuarios;
DROP TABLE IF EXISTS usuarios;
DROP FUNCTION IF EXISTS fijar_actualizado_en();
