import { inject, Service } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

import { ActualizarTareaPayload, Estatus, Tarea, TareaPayload } from '../models/tarea';

/** Tareas envuelve los endpoints de tareas de la API. */
@Service()
export class Tareas {
    private readonly http = inject(HttpClient);

    crear(usuarioId: number, datos: TareaPayload): Observable<Tarea> {
        return this.http.post<Tarea>(`/api/usuarios/${usuarioId}/tareas`, datos);
    }

    obtener(id: number): Observable<Tarea> {
        return this.http.get<Tarea>(`/api/tareas/${id}`);
    }

    actualizar(id: number, datos: ActualizarTareaPayload): Observable<Tarea> {
        return this.http.put<Tarea>(`/api/tareas/${id}`, datos);
    }

    cambiarEstatus(id: number, estatus: Estatus): Observable<Tarea> {
        return this.http.patch<Tarea>(`/api/tareas/${id}/estatus`, { estatus });
    }

    eliminar(id: number): Observable<void> {
        return this.http.delete<void>(`/api/tareas/${id}`);
    }
}