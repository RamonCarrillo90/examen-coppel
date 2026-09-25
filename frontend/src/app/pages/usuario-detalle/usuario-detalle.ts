import { Component, inject, input, OnInit, signal } from '@angular/core';
import { DatePipe } from '@angular/common';
import { RouterLink } from '@angular/router';

import { Estatus } from '../../core/models/tarea';
import { UsuarioConTareas } from '../../core/models/usuario';
import { Auth } from '../../core/services/auth';
import { Usuarios } from '../../core/services/usuarios';

/** Detalle de un usuario con sus tareas. */
@Component({
  imports: [RouterLink, DatePipe],
  selector: 'app-usuario-detalle',
  styleUrl: './usuario-detalle.scss',
  templateUrl: './usuario-detalle.html',
})
export class UsuarioDetalle implements OnInit {
  private readonly usuariosService = inject(Usuarios);
  protected readonly auth = inject(Auth);

  /** :id de la ruta (llega gracias a withComponentInputBinding). */
  readonly id = input.required<string>();

  protected readonly usuario = signal<UsuarioConTareas | null>(null);
  protected readonly error = signal<string | null>(null);

  protected readonly etiquetas: Record<Estatus, string> = {
    pendiente: 'Pendiente',
    en_progreso: 'En progreso',
    completada: 'Completada',
  };

  ngOnInit(): void {
    this.usuariosService.obtener(Number(this.id())).subscribe({
      next: (u) => this.usuario.set(u),
      error: () => this.error.set('No se pudo cargar el usuario'),
    });
  }
}