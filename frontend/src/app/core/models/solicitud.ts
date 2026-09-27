/** Estatus posibles de una solicitud (coinciden con el backend). */
export type EstatusSolicitud = 'pendiente' | 'aprobada' | 'rechazada';

export const ETIQUETAS_SOLICITUD: Record<EstatusSolicitud, string> = {
    pendiente: 'Pendiente',
    aprobada: 'Aprobada',
    rechazada: 'Rechazada',
};

/** Solicitud tal como la devuelve la API al listar (con datos para mostrar). */
export interface Solicitud {
    id: number;
    tarea_id: number;
    usuario_id: number;
    estatus: EstatusSolicitud;
    creado_en: string;
    resuelto_en: string | null;
    tarea_titulo: string;
    usuario_nombre: string;
}