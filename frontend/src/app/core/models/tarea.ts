/** Estatus posibles de una tarea (coinciden con el backend). */
export type Estatus = 'pendiente' | 'en_progreso' | 'completada';

/** Tarea tal como la devuelve la API. */
export interface Tarea {
  id: number;
  usuario_id: number | null; // null = sin asignar
  titulo: string;
  descripcion: string | null;
  fecha_limite: string | null;
  estatus: Estatus;
  creado_en: string;
  actualizado_en: string;
}

/** Campos comunes para crear una tarea (siempre nace "pendiente"). */
export interface TareaPayload {
  titulo: string;
  descripcion: string | null;
  fecha_limite: string | null; // "AAAA-MM-DD"
}

/** Cuerpo de POST /api/tareas: usuario_id null = sin asignar. */
export interface CrearTareaPayload extends TareaPayload {
  usuario_id: number | null;
}

/** Cuerpo para actualizar (reasignar o desasignar) una tarea. */
export interface ActualizarTareaPayload extends TareaPayload {
  usuario_id: number | null;
  estatus: Estatus;
}

/** Filtros opcionales de GET /api/tareas (los vacíos no se mandan). */
export interface FiltroTareas {
  q?: string;
  estatus?: Estatus | '';
  usuario_id?: number;
  sin_asignar?: boolean;
}

/** Todos los estatus, en el orden en que se muestran en las listas. */
export const ESTATUS: Estatus[] = ['pendiente', 'en_progreso', 'completada'];

/** Texto para mostrar de cada estatus. */
export const ETIQUETAS_ESTATUS: Record<Estatus, string> = {
  pendiente: 'Pendiente',
  en_progreso: 'En progreso',
  completada: 'Completada',
};