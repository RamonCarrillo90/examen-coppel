/** Estatus posibles de una tarea (coinciden con el backend). */
export type Estatus = 'pendiente' | 'en_progreso' | 'completada';

/** Tarea tal como la devuelve la API. */
export interface Tarea {
  id: number;
  usuario_id: number;
  titulo: string;
  descripcion: string | null;
  fecha_limite: string | null;
  estatus: Estatus;
  creado_en: string;
  actualizado_en: string;
}

/** Cuerpo para crear una tarea (el usuario va en la URL). */
export interface TareaPayload {
  titulo: string;
  descripcion: string | null;
  fecha_limite: string | null; // "AAAA-MM-DD"
  estatus?: Estatus;
}

/** Cuerpo para actualizar (y reasignar) una tarea. */
export interface ActualizarTareaPayload extends TareaPayload {
  usuario_id: number;
  estatus: Estatus;
}

/** Todos los estatus, en el orden en que se muestran en las listas. */
export const ESTATUS: Estatus[] = ['pendiente', 'en_progreso', 'completada'];

/** Texto para mostrar de cada estatus. */
export const ETIQUETAS_ESTATUS: Record<Estatus, string> = {
  pendiente: 'Pendiente',
  en_progreso: 'En progreso',
  completada: 'Completada',
};