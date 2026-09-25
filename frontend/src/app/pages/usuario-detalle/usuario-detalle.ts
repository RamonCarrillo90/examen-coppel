import { Component, computed, inject, input, OnInit, signal } from '@angular/core';
import { DatePipe } from '@angular/common';
import { Router, RouterLink } from '@angular/router';

import { ESTATUS, ETIQUETAS_ESTATUS, Estatus, Tarea } from '../../core/models/tarea';
import { UsuarioConTareas } from '../../core/models/usuario';
import { Auth } from '../../core/services/auth';
import { Tareas } from '../../core/services/tareas';
import { Usuarios } from '../../core/services/usuarios';
import { mensajeError } from '../../core/utils/errores';

/** Detalle de un usuario con sus tareas y las acciones sobre ellas. */
@Component({
  imports: [RouterLink, DatePipe],
  selector: 'app-usuario-detalle',
  styleUrl: './usuario-detalle.scss',
  templateUrl: './usuario-detalle.html',
})
export class UsuarioDetalle implements OnInit {
  private readonly usuariosService = inject(Usuarios);
  private readonly tareasService = inject(Tareas);
  private readonly router = inject(Router);
  protected readonly auth = inject(Auth);

  /** :id de la ruta (llega gracias a withComponentInputBinding). */
  readonly id = input.required<string>();

  protected readonly usuario = signal<UsuarioConTareas | null>(null);
  /** Error al cargar la pantalla (no hay nada que mostrar). */
  protected readonly error = signal<string | null>(null);
  /** Error de una acción (la pantalla sigue visible). */
  protected readonly aviso = signal<string | null>(null);

  protected readonly estatus = ESTATUS;
  protected readonly etiquetas = ETIQUETAS_ESTATUS;

  /** El admin no puede eliminarse a sí mismo (regla del backend). */
  protected readonly puedeEliminarUsuario = computed(
    () => this.auth.esAdmin() && this.usuario()?.id !== this.auth.usuario()?.id,
  );

  ngOnInit(): void {
    this.cargar();
  }

  private cargar(): void {
    this.usuariosService.obtener(Number(this.id())).subscribe({
      next: (u) => this.usuario.set(u),
      error: (err) => this.error.set(mensajeError(err, 'No se pudo cargar el usuario')),
    });
  }

  /** Cambia el estatus con PATCH (lo puede hacer el admin o el dueño de la tarea). */
  protected cambiarEstatus(tarea: Tarea, valor: string): void {
    this.aviso.set(null);
    this.tareasService.cambiarEstatus(tarea.id, valor as Estatus).subscribe({
      next: (actualizada) =>
        this.usuario.update(
          (u) => u && { ...u, tareas: u.tareas.map((t) => (t.id === actualizada.id ? actualizada : t)) },
        ),
      error: (err) => {
        this.aviso.set(mensajeError(err, 'No se pudo cambiar el estatus'));
        this.cargar(); // regresa el selector al valor real
      },
    });
  }

  protected eliminarTarea(tarea: Tarea): void {
    if (!confirm(`¿Eliminar la tarea "${tarea.titulo}"? Esta acción no se puede deshacer.`)) return;

    this.aviso.set(null);
    this.tareasService.eliminar(tarea.id).subscribe({
      next: () =>
        this.usuario.update((u) => u && { ...u, tareas: u.tareas.filter((t) => t.id !== tarea.id) }),
      error: (err) => this.aviso.set(mensajeError(err, 'No se pudo eliminar la tarea')),
    });
  }

  protected eliminarUsuario(u: UsuarioConTareas): void {
    const n = u.tareas.length;
    const aviso = n > 0 ? ` También se eliminarán sus ${n} tarea(s).` : '';
    if (!confirm(`¿Eliminar a ${u.nombre} ${u.apellido}?${aviso} Esta acción no se puede deshacer.`)) return;

    this.aviso.set(null);
    this.usuariosService.eliminar(u.id).subscribe({
      next: () => this.router.navigate(['/usuarios']),
      error: (err) => this.aviso.set(mensajeError(err, 'No se pudo eliminar el usuario')),
    });
  }
}