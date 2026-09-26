import { Component, computed, inject, input, OnInit, signal } from '@angular/core';
import { Location } from '@angular/common';
import {
  AbstractControl,
  FormBuilder,
  ReactiveFormsModule,
  ValidationErrors,
  Validators,
} from '@angular/forms';
import { Router } from '@angular/router';
import { forkJoin, Observable, of } from 'rxjs';

import { ESTATUS, ETIQUETAS_ESTATUS, Estatus, Tarea } from '../../core/models/tarea';
import { Usuario } from '../../core/models/usuario';
import { Tareas } from '../../core/services/tareas';
import { Usuarios } from '../../core/services/usuarios';
import { mensajeError } from '../../core/utils/errores';
import { hoyLocal } from '../../core/utils/fechas';

/**
 * Formulario de tareas. Tiene tres modos, según la ruta:
 * - /usuarios/:usuarioId/tareas/nueva → crear desde un perfil: el usuario es fijo.
 * - /tareas/nueva                     → crear desde la página de tareas: se elige usuario o "Sin asignar".
 * - /tareas/:id/editar                → editar: se puede reasignar, desasignar y cambiar estatus.
 * Al crear, el estatus siempre es "pendiente" (lo fija el backend).
 */
@Component({
  imports: [ReactiveFormsModule],
  selector: 'app-tarea-form',
  styleUrl: './tarea-form.scss',
  templateUrl: './tarea-form.html',
})
export class TareaForm implements OnInit {
  private readonly tareasService = inject(Tareas);
  private readonly usuariosService = inject(Usuarios);
  private readonly router = inject(Router);
  private readonly location = inject(Location);

  /** :id de la tarea (solo al editar). */
  readonly id = input<string>();
  /** :usuarioId del dueño (solo al crear desde un perfil). */
  readonly usuarioId = input<string>();

  protected readonly esEdicion = computed(() => this.id() !== undefined);
  protected readonly desdePerfil = computed(() => this.usuarioId() !== undefined);

  protected readonly estatus = ESTATUS;
  protected readonly etiquetas = ETIQUETAS_ESTATUS;

  /** Usuarios para el selector "Asignada a". */
  protected readonly usuarios = signal<Usuario[]>([]);

  /** Dueño fijo al crear desde un perfil. */
  protected readonly dueno = computed(() =>
    this.usuarios().find((u) => u.id === Number(this.usuarioId())),
  );

  /** Hoy en "AAAA-MM-DD": también es el [min] del calendario. */
  protected readonly hoy = hoyLocal();
  /** Fecha que ya tenía la tarea al editar (puede estar vencida y se permite conservarla). */
  private fechaOriginal = '';

  protected readonly form = inject(FormBuilder).nonNullable.group({
    titulo: ['', [Validators.required, Validators.maxLength(150)]],
    descripcion: ['', Validators.maxLength(1000)],
    fecha_limite: ['', (c: AbstractControl<string>) => this.validarFecha(c)], // "AAAA-MM-DD" o vacío
    estatus: ['pendiente' as Estatus],
    usuario_id: [null as number | null], // null = sin asignar
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
          // El backend manda "2026-09-30T00:00:00Z"; el <input type="date"> quiere "2026-09-30".
          this.fechaOriginal = tarea.fecha_limite?.slice(0, 10) ?? '';
          this.form.patchValue({
            titulo: tarea.titulo,
            descripcion: tarea.descripcion ?? '',
            fecha_limite: this.fechaOriginal,
            estatus: tarea.estatus,
            usuario_id: tarea.usuario_id,
          });
        } else if (this.desdePerfil()) {
          this.form.patchValue({ usuario_id: Number(this.usuarioId()) });
        }
        this.cargando.set(false);
      },
      error: (err) => {
        this.error.set(mensajeError(err, 'No se pudo cargar la información'));
        this.cargando.set(false);
      },
    });
  }

  /**
   * La fecha límite no puede ser anterior a hoy. Excepción: al editar, se puede
   * conservar la fecha que ya tenía (si no, una tarea vencida no se podría editar).
   * Las fechas "AAAA-MM-DD" se pueden comparar como texto.
   */
  private validarFecha(control: AbstractControl<string>): ValidationErrors | null {
    const fecha = control.value;
    if (!fecha || fecha === this.fechaOriginal || fecha >= this.hoy) return null;
    return { fechaPasada: true };
  }

  /** true si el campo ya fue tocado y es inválido (para mostrar su mensaje). */
  protected invalido(campo: string): boolean {
    const control = this.form.get(campo);
    return !!control && control.touched && control.invalid;
  }

  /** Regresa a la pantalla desde donde se llegó (perfil, lista de tareas...). */
  protected volver(): void {
    this.location.back();
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
    };

    // Al crear NO se manda estatus: el backend siempre la crea "pendiente".
    let peticion: Observable<Tarea>;
    if (this.esEdicion()) {
      peticion = this.tareasService.actualizar(Number(this.id()), {
        ...datos,
        estatus: v.estatus,
        usuario_id: v.usuario_id,
      });
    } else if (this.desdePerfil()) {
      peticion = this.tareasService.crear(Number(this.usuarioId()), datos);
    } else {
      peticion = this.tareasService.crearGeneral({ ...datos, usuario_id: v.usuario_id });
    }

    this.enviando.set(true);
    this.error.set(null);
    peticion.subscribe({
      next: () => this.volver(),
      error: (err) => {
        this.error.set(mensajeError(err, 'No se pudo guardar la tarea'));
        this.enviando.set(false);
      },
    });
  }
}