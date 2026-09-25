import { inject, Service } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

import { Usuario, UsuarioConTareas, UsuarioPayload } from '../models/usuario';

/** Usuarios envuelve los endpoints /api/usuarios de la API. */
@Service()
export class Usuarios {
    private readonly http = inject(HttpClient);
    private readonly url = '/api/usuarios';

    listar(): Observable<Usuario[]> {
        return this.http.get<Usuario[]>(this.url);
    }

    obtener(id: number): Observable<UsuarioConTareas> {
        return this.http.get<UsuarioConTareas>(`${this.url}/${id}`);
    }

    crear(datos: UsuarioPayload): Observable<Usuario> {
        return this.http.post<Usuario>(this.url, datos);
    }

    actualizar(id: number, datos: UsuarioPayload): Observable<Usuario> {
        return this.http.put<Usuario>(`${this.url}/${id}`, datos);
    }

    eliminar(id: number): Observable<void> {
        return this.http.delete<void>(`${this.url}/${id}`);
    }
}