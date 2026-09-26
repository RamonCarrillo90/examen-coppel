import { Component, DestroyRef, inject, OnInit, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { DatePipe } from '@angular/common';
import { FormControl, ReactiveFormsModule } from '@angular/forms';
import { catchError, debounceTime, of, startWith, switchMap, tap } from 'rxjs';

import { Tarea } from '../../core/models/tarea';
import { Tareas } from '../../core/services/tareas';
import { mensajeError } from '../../core/utils/errores';

/** Tareas sin usuario, visibles para cualquier usuario (el paso siguiente es solicitarlas). */
@Component({
  imports: [ReactiveFormsModule, DatePipe],
  selector: 'app-disponibles',
  styleUrl: './disponibles.scss',
  templateUrl: './disponibles.html',
})
export class Disponibles implements OnInit {
  private readonly tareasService = inject(Tareas);
  private readonly destroyRef = inject(DestroyRef);

  protected readonly tareas = signal<Tarea[]>([]);
  protected readonly cargando = signal(true);
  protected readonly error = signal<string | null>(null);

  /** Un solo campo de búsqueda: basta un FormControl, no hace falta un FormGroup. */
  protected readonly busqueda = new FormControl('', { nonNullable: true });

  ngOnInit(): void {
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
}