import { Tarea } from './tarea';

/** Roles posibles de un usuario (coinciden con el backend). */
export type Rol = 'admin' | 'usuario';

/** Usuario tal como lo devuelve la API (sin password_hash). */
export interface Usuario {
  id: number;
  nombre: string;
  apellido: string;
  email: string;
  telefono: string | null;
  rol: Rol;
  creado_en: string;
  actualizado_en: string;
}

/** Respuesta de GET /api/usuarios/{id}: el usuario con sus tareas. */
export interface UsuarioConTareas extends Usuario {
  tareas: Tarea[];
}

/** Cuerpo para crear o actualizar un usuario. */
export interface UsuarioPayload {
  nombre: string;
  apellido: string;
  email: string;
  telefono: string | null;
  password?: string;
  rol?: Rol;
}

/** Filtros opcionales de GET /api/usuarios (los vacíos no se mandan). */
export interface FiltroUsuarios {
  q?: string;
  rol?: Rol | '';
}