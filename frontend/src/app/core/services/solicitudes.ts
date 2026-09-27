import { inject, Service } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';

import { EstatusSolicitud, Solicitud } from '../models/solicitud';

/** Solicitudes envuelve los endpoints de solicitudes de tareas. */
@Service()
export class Solicitudes {
    private readonly http = inject(HttpClient);

   /** El usuario pide quedarse con una tarea disponible. */
    solicitar(tareaId: number): Observable<unknown> {
        return this.http.post(`/api/tareas/${tareaId}/solicitudes`, {});
    }

   /** Las solicitudes del usuario en sesión. */
    mias(): Observable<Solicitud[]> {
        return this.http.get<Solicitud[]>('/api/solicitudes/mias');
    }

   /** Todas las solicitudes, opcionalmente por estatus (solo admin). */
    listar(estatus: EstatusSolicitud | '' = ''): Observable<Solicitud[]> {
        const params = estatus ? new HttpParams().set('estatus', estatus) : undefined;
        return this.http.get<Solicitud[]>('/api/solicitudes', { params });
    }

    aprobar(id: number): Observable<void> {
        return this.http.post<void>(`/api/solicitudes/${id}/aprobar`, {});
    }

    rechazar(id: number): Observable<void> {
        return this.http.post<void>(`/api/solicitudes/${id}/rechazar`, {});
    }
}