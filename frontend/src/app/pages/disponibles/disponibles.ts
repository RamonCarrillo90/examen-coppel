import { Component, computed, DestroyRef, inject, OnInit, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { DatePipe } from '@angular/common';
import { FormControl, ReactiveFormsModule } from '@angular/forms';
import { catchError, debounceTime, of, startWith, switchMap, tap } from 'rxjs';

import { ETIQUETAS_SOLICITUD, Solicitud } from '../../core/models/solicitud';
import { Tarea } from '../../core/models/tarea';
import { Solicitudes } from '../../core/services/solicitudes';
import { Tareas } from '../../core/services/tareas';
import { mensajeError } from '../../core/utils/errores';

/**
 * Tareas sin usuario: el usuario puede solicitarlas, y el admin decide.
 * Abajo se muestran sus solicitudes y cómo van.
 */
@Component({
  imports: [ReactiveFormsModule, DatePipe],
  selector: 'app-disponibles',
  styleUrl: './disponibles.scss',
  templateUrl: './disponibles.html',
})
export class Disponibles implements OnInit {
  private readonly tareasService = inject(Tareas);
  private readonly solicitudesService = inject(Solicitudes);
  private readonly destroyRef = inject(DestroyRef);

  protected readonly etiquetas = ETIQUETAS_SOLICITUD;

  protected readonly tareas = signal<Tarea[]>([]);
  protected readonly misSolicitudes = signal<Solicitud[]>([]);
  protected readonly cargando = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly aviso = signal<string | null>(null);
  /** id de la tarea que se está solicitando (para deshabilitar su botón). */
  protected readonly solicitando = signal<number | null>(null);

  /** Tareas que ya pedí y siguen esperando respuesta. */
  protected readonly yaSolicitadas = computed(
    () =>
      new Set(
        this.misSolicitudes()
          .filter((s) => s.estatus === 'pendiente')
          .map((s) => s.tarea_id),
      ),
  );

  /** Un solo campo de búsqueda: basta un FormControl, no hace falta un FormGroup. */
  protected readonly busqueda = new FormControl('', { nonNullable: true });

  ngOnInit(): void {
    this.cargarMisSolicitudes();

    this.busqueda.valueChanges
      .pipe(
        startWith(''),
        debounceTime(300),
        tap(() => this.cargando.set(true)),
        switchMap((q) =>
          this.tareasService.disponibles(q.trim()).pipe(
            catchError((err) => {
              this.error.set(mensajeError(err, 'No se pudieron cargar las tareas disponibles'));
              return of(null);
            }),
          ),
        ),
        takeUntilDestroyed(this.destroyRef),
      )
      .subscribe((lista) => {
        if (lista) {
          this.tareas.set(lista);
          this.error.set(null);
        }
        this.cargando.set(false);
      });
  }

  private cargarMisSolicitudes(): void {
    this.solicitudesService.mias().subscribe({
      next: (lista) => this.misSolicitudes.set(lista),
      error: (err) => this.aviso.set(mensajeError(err, 'No se pudieron cargar tus solicitudes')),
    });
  }

  protected solicitar(tarea: Tarea): void {
    this.aviso.set(null);
    this.solicitando.set(tarea.id);
    this.solicitudesService.solicitar(tarea.id).subscribe({
      next: () => {
        this.solicitando.set(null);
        this.cargarMisSolicitudes();
      },
      error: (err) => {
        this.solicitando.set(null);
        this.aviso.set(mensajeError(err, 'No se pudo enviar la solicitud'));
        // Si ya no está disponible, se vuelve a pedir la lista (setValue dispara valueChanges).
        this.busqueda.setValue(this.busqueda.value);
      },
    });
  }
}