import { inject, Service } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';

import {
  ActualizarTareaPayload,
  CrearTareaPayload,
  Estatus,
  FiltroTareas,
  Tarea,
  TareaPayload,
} from '../models/tarea';

/** Tareas envuelve los endpoints de tareas de la API. */
@Service()
export class Tareas {
  private readonly http = inject(HttpClient);

  /** Todas las tareas con filtros opcionales (solo admin). */
  listar(filtros: FiltroTareas = {}): Observable<Tarea[]> {
    let params = new HttpParams();
    if (filtros.q) params = params.set('q', filtros.q);
    if (filtros.estatus) params = params.set('estatus', filtros.estatus);
    if (filtros.usuario_id) params = params.set('usuario_id', filtros.usuario_id);
    if (filtros.sin_asignar) params = params.set('sin_asignar', 'true');
    return this.http.get<Tarea[]>('/api/tareas', { params });
  }

  /** Tareas sin usuario, que cualquier usuario puede ver (y, en el futuro, solicitar). */
  disponibles(q = ''): Observable<Tarea[]> {
    const params = q ? new HttpParams().set('q', q) : undefined;
    return this.http.get<Tarea[]>('/api/tareas/disponibles', { params });
  }

  /** Crea una tarea para el usuario de un perfil. */
  crear(usuarioId: number, datos: TareaPayload): Observable<Tarea> {
    return this.http.post<Tarea>(`/api/usuarios/${usuarioId}/tareas`, datos);
  }

  /** Crea una tarea con o sin usuario (desde la página de tareas). */
  crearGeneral(datos: CrearTareaPayload): Observable<Tarea> {
    return this.http.post<Tarea>('/api/tareas', datos);
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