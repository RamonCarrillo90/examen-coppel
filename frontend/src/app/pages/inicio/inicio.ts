import { Component, computed, inject, OnInit, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { forkJoin } from 'rxjs';

import { Tarea } from '../../core/models/tarea';
import { Auth } from '../../core/services/auth';
import { Solicitudes } from '../../core/services/solicitudes';
import { Tareas } from '../../core/services/tareas';
import { Usuarios } from '../../core/services/usuarios';

/** Una tarjeta del menú: a dónde lleva y qué resumen muestra. */
interface Tarjeta {
  titulo: string;
  resumen: string;
  ruta: string;
}

/**
 * Menú de inicio. El admin elige entre Usuarios y Tareas; el usuario normal,
 * entre sus tareas y las disponibles. Cada tarjeta muestra un pequeño resumen.
 */
@Component({
  imports: [RouterLink],
  selector: 'app-inicio',
  styleUrl: './inicio.scss',
  templateUrl: './inicio.html',
})
export class Inicio implements OnInit {
  private readonly usuariosService = inject(Usuarios);
  private readonly tareasService = inject(Tareas);
  private readonly solicitudesService = inject(Solicitudes);
  protected readonly auth = inject(Auth);

  protected readonly tarjetas = signal<Tarjeta[]>([]);
  protected readonly error = signal<string | null>(null);

  protected readonly saludo = computed(() => `Hola, ${this.auth.usuario()?.nombre ?? ''}`);

  ngOnInit(): void {
    const usuario = this.auth.usuario();
    if (!usuario) return;

    if (this.auth.esAdmin()) {
      forkJoin({
        usuarios: this.usuariosService.listar(),
        tareas: this.tareasService.listar(),
        pendientes: this.solicitudesService.listar('pendiente'),
      }).subscribe({
        next: ({ usuarios, tareas, pendientes }) =>
          this.tarjetas.set([
            { titulo: 'Usuarios', resumen: plural(usuarios.length, 'usuario registrado', 'usuarios registrados'), ruta: '/usuarios' },
            { titulo: 'Tareas', resumen: resumenTareas(tareas, true), ruta: '/tareas' },
            {
              titulo: 'Solicitudes',
              resumen: plural(pendientes.length, 'solicitud por revisar', 'solicitudes por revisar'),
              ruta: '/solicitudes',
            },
          ]),
        error: () => this.error.set('No se pudo cargar el resumen'),
      });
    } else {
      forkJoin({
        mias: this.usuariosService.obtener(usuario.id),
        disponibles: this.tareasService.disponibles(),
      }).subscribe({
        next: ({ mias, disponibles }) =>
          this.tarjetas.set([
            { titulo: 'Mis tareas', resumen: resumenTareas(mias.tareas, false), ruta: `/usuarios/${usuario.id}` },
            { titulo: 'Tareas disponibles', resumen: plural(disponibles.length, 'tarea sin asignar', 'tareas sin asignar'), ruta: '/disponibles' },
          ]),
        error: () => this.error.set('No se pudo cargar el resumen'),
      });
    }
  }
}

/** "12 tareas · 5 pendientes · 2 sin asignar" (lo último solo para el admin). */
function resumenTareas(tareas: Tarea[], incluirSinAsignar: boolean): string {
  const pendientes = tareas.filter((t) => t.estatus === 'pendiente').length;
  const partes = [plural(tareas.length, 'tarea', 'tareas'), plural(pendientes, 'pendiente', 'pendientes')];
  if (incluirSinAsignar) {
    partes.push(`${tareas.filter((t) => t.usuario_id === null).length} sin asignar`);
  }
  return partes.join(' · ');
}

/** "1 tarea" / "3 tareas": elige singular o plural según la cantidad. */
function plural(n: number, singular: string, varios: string): string {
  return `${n} ${n === 1 ? singular : varios}`;
}