import { Component, computed, inject, input, OnInit, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { forkJoin, Observable, of } from 'rxjs';

import { ESTATUS, ETIQUETAS_ESTATUS, Estatus, Tarea } from '../../core/models/tarea';
import { Usuario } from '../../core/models/usuario';
import { Tareas } from '../../core/services/tareas';
import { Usuarios } from '../../core/services/usuarios';
import { mensajeError } from '../../core/utils/errores';

/**
 * Formulario para crear (/usuarios/:usuarioId/tareas/nueva) o editar
 * (/tareas/:id/editar) una tarea. Cambiar "Asignada a" al editar es REASIGNAR.
 */
@Component({
  imports: [ReactiveFormsModule, RouterLink],
  selector: 'app-tarea-form',
  styleUrl: './tarea-form.scss',
  templateUrl: './tarea-form.html',
})
export class TareaForm implements OnInit {
  private readonly tareasService = inject(Tareas);
  private readonly usuariosService = inject(Usuarios);
  private readonly router = inject(Router);

  /** :id de la tarea (solo al editar). */
  readonly id = input<string>();
  /** :usuarioId del dueño (solo al crear). */
  readonly usuarioId = input<string>();

  protected readonly esEdicion = computed(() => this.id() !== undefined);

  protected readonly estatus = ESTATUS;
  protected readonly etiquetas = ETIQUETAS_ESTATUS;

  /** Usuarios para el selector "Asignada a". */
  protected readonly usuarios = signal<Usuario[]>([]);
  /** Dueño original: a su detalle regresa "Cancelar". */
  protected readonly duenoOriginal = signal<number | null>(null);

  protected readonly form = inject(FormBuilder).nonNullable.group({
    titulo: ['', [Validators.required, Validators.maxLength(150)]],
    descripcion: ['', Validators.maxLength(1000)],
    fecha_limite: [''], // "AAAA-MM-DD", o vacío si no tiene
    estatus: ['pendiente' as Estatus],
    usuario_id: [0, Validators.min(1)],
  });

  protected readonly cargando = signal(true);
  protected readonly enviando = signal(false);
  protected readonly error = signal<string | null>(null);

  ngOnInit(): void {
    // Se piden en paralelo la lista de usuarios y, si se edita, la tarea.
    const tarea$: Observable<Tarea | null> = this.esEdicion()
      ? this.tareasService.obtener(Number(this.id()))
      : of(null);

    forkJoin({ usuarios: this.usuariosService.listar(), tarea: tarea$ }).subscribe({
      next: ({ usuarios, tarea }) => {
        this.usuarios.set(usuarios);
        if (tarea) {
          this.form.patchValue({
            titulo: tarea.titulo,
            descripcion: tarea.descripcion ?? '',
            // El backend manda "2026-09-30T00:00:00Z"; el <input type="date"> quiere "2026-09-30".
            fecha_limite: tarea.fecha_limite?.slice(0, 10) ?? '',
            estatus: tarea.estatus,
            usuario_id: tarea.usuario_id,
          });
          this.duenoOriginal.set(tarea.usuario_id);
        } else {
          const dueno = Number(this.usuarioId());
          this.form.patchValue({ usuario_id: dueno });
          this.duenoOriginal.set(dueno);
        }
        this.cargando.set(false);
      },
      error: (err) => {
        this.error.set(mensajeError(err, 'No se pudo cargar la información'));
        this.cargando.set(false);
      },
    });
  }

  /** true si el campo ya fue tocado y es inválido (para mostrar su mensaje). */
  protected invalido(campo: string): boolean {
    const control = this.form.get(campo);
    return !!control && control.touched && control.invalid;
  }

  protected guardar(): void {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    const v = this.form.getRawValue();
    const datos = {
      titulo: v.titulo.trim(),
      descripcion: v.descripcion.trim() || null,
      fecha_limite: v.fecha_limite || null,
      estatus: v.estatus,
    };

    const peticion: Observable<Tarea> = this.esEdicion()
      ? this.tareasService.actualizar(Number(this.id()), { ...datos, usuario_id: v.usuario_id })
      : this.tareasService.crear(v.usuario_id, datos);

    this.enviando.set(true);
    this.error.set(null);
    peticion.subscribe({
      // Se regresa al detalle del dueño FINAL (si se reasignó, al del nuevo dueño).
      next: (t) => this.router.navigate(['/usuarios', t.usuario_id]),
      error: (err) => {
        this.error.set(mensajeError(err, 'No se pudo guardar la tarea'));
        this.enviando.set(false);
      },
    });
  }
}