import { Component, computed, DestroyRef, inject, OnInit, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { DatePipe } from '@angular/common';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { catchError, debounceTime, of, startWith, switchMap, tap } from 'rxjs';

import { ESTATUS, ETIQUETAS_ESTATUS, FiltroTareas, Tarea } from '../../core/models/tarea';
import { Usuario } from '../../core/models/usuario';
import { Tareas } from '../../core/services/tareas';
import { Usuarios } from '../../core/services/usuarios';
import { mensajeError } from '../../core/utils/errores';

/** Todas las tareas con buscador y filtros (solo admin). */
@Component({
  imports: [ReactiveFormsModule, RouterLink, DatePipe],
  selector: 'app-tareas-lista',
  styleUrl: './tareas-lista.scss',
  templateUrl: './tareas-lista.html',
})
export class TareasLista implements OnInit {
  private readonly tareasService = inject(Tareas);
  private readonly usuariosService = inject(Usuarios);
  private readonly destroyRef = inject(DestroyRef);

  protected readonly estatus = ESTATUS;
  protected readonly etiquetas = ETIQUETAS_ESTATUS;

  protected readonly tareas = signal<Tarea[]>([]);
  protected readonly usuarios = signal<Usuario[]>([]);
  protected readonly cargando = signal(true);
  protected readonly error = signal<string | null>(null);

  /** id → "Nombre Apellido", para mostrar a quién está asignada cada tarea. */
  protected readonly nombres = computed(
    () => new Map(this.usuarios().map((u) => [u.id, `${u.nombre} ${u.apellido}`])),
  );

  /** usuario: '' = todos, 'sin' = sin asignar, o el id como texto. */
  protected readonly filtros = inject(FormBuilder).nonNullable.group({
    q: [''],
    estatus: [''],
    usuario: [''],
  });

  ngOnInit(): void {
    // Los usuarios solo se necesitan para el filtro y para mostrar nombres.
    this.usuariosService.listar().subscribe({
      next: (lista) => this.usuarios.set(lista),
    });

    this.filtros.valueChanges
      .pipe(
        startWith(this.filtros.getRawValue()), // la primera búsqueda, sin esperar a que escriban
        debounceTime(300), // espera 300 ms sin cambios antes de buscar
        tap(() => this.cargando.set(true)),
        // switchMap cancela la búsqueda anterior si llega una nueva:
        // así nunca se muestra el resultado viejo de una búsqueda lenta.
        switchMap(() =>
          this.tareasService.listar(this.filtroActual()).pipe(
            catchError((err) => {
              this.error.set(mensajeError(err, 'No se pudo cargar la lista de tareas'));
              return of(null);
            }),
          ),
        ),
        takeUntilDestroyed(this.destroyRef), // se desuscribe solo al salir de la pantalla
      )
      .subscribe((lista) => {
        if (lista) {
          this.tareas.set(lista);
          this.error.set(null);
        }
        this.cargando.set(false);
      });
  }

  /** Convierte lo que hay en el formulario de filtros en lo que espera la API. */
  private filtroActual(): FiltroTareas {
    const f = this.filtros.getRawValue();
    return {
      q: f.q.trim(),
      estatus: f.estatus as FiltroTareas['estatus'],
      sin_asignar: f.usuario === 'sin',
      usuario_id: f.usuario && f.usuario !== 'sin' ? Number(f.usuario) : undefined,
    };
  }

  protected limpiar(): void {
    this.filtros.setValue({ q: '', estatus: '', usuario: '' });
  }

  protected eliminar(tarea: Tarea): void {
    if (!confirm(`¿Eliminar la tarea "${tarea.titulo}"? Esta acción no se puede deshacer.`)) return;
    this.tareasService.eliminar(tarea.id).subscribe({
      next: () => this.tareas.update((lista) => lista.filter((t) => t.id !== tarea.id)),
      error: (err) => this.error.set(mensajeError(err, 'No se pudo eliminar la tarea')),
    });
  }
}