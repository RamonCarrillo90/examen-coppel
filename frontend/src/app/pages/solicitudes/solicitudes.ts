import { Component, DestroyRef, inject, OnInit, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { DatePipe } from '@angular/common';
import { FormControl, ReactiveFormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';

import { ETIQUETAS_SOLICITUD, EstatusSolicitud, Solicitud } from '../../core/models/solicitud';
import { Solicitudes } from '../../core/services/solicitudes';
import { mensajeError } from '../../core/utils/errores';

/** Bandeja de solicitudes del admin: aprobar o rechazar. */
@Component({
  imports: [ReactiveFormsModule, RouterLink, DatePipe],
  selector: 'app-solicitudes',
  styleUrl: './solicitudes.scss',
  templateUrl: './solicitudes.html',
})
export class SolicitudesPagina implements OnInit {
  private readonly solicitudesService = inject(Solicitudes);
  private readonly destroyRef = inject(DestroyRef);

  protected readonly etiquetas = ETIQUETAS_SOLICITUD;

  protected readonly solicitudes = signal<Solicitud[]>([]);
  protected readonly cargando = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly aviso = signal<string | null>(null);

  /** Por defecto se ven las pendientes: es lo que el admin tiene que atender. */
  protected readonly estatus = new FormControl<EstatusSolicitud | ''>('pendiente', { nonNullable: true });

  ngOnInit(): void {
    this.cargar();
    this.estatus.valueChanges.pipe(takeUntilDestroyed(this.destroyRef)).subscribe(() => this.cargar());
  }

  private cargar(): void {
    this.cargando.set(true);
    this.solicitudesService.listar(this.estatus.value).subscribe({
      next: (lista) => {
        this.solicitudes.set(lista);
        this.error.set(null);
        this.cargando.set(false);
      },
      error: (err) => {
        this.error.set(mensajeError(err, 'No se pudieron cargar las solicitudes'));
        this.cargando.set(false);
      },
    });
  }

  protected aprobar(s: Solicitud): void {
    const texto = `¿Asignar "${s.tarea_titulo}" a ${s.usuario_nombre}? Las demás solicitudes de esta tarea se rechazarán.`;
    if (!confirm(texto)) return;
    this.aviso.set(null);
    this.solicitudesService.aprobar(s.id).subscribe({
      // Se recarga todo: aprobar una solicitud puede haber rechazado otras.
      next: () => this.cargar(),
      error: (err) => {
        this.aviso.set(mensajeError(err, 'No se pudo aprobar la solicitud'));
        this.cargar();
      },
    });
  }

  protected rechazar(s: Solicitud): void {
    if (!confirm(`¿Rechazar la solicitud de ${s.usuario_nombre} para "${s.tarea_titulo}"?`)) return;
    this.aviso.set(null);
    this.solicitudesService.rechazar(s.id).subscribe({
      next: () => this.cargar(),
      error: (err) => {
        this.aviso.set(mensajeError(err, 'No se pudo rechazar la solicitud'));
        this.cargar();
      },
    });
  }
}