import { Component, computed, inject, input, OnInit, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { Observable } from 'rxjs';

import { Rol, Usuario, UsuarioPayload } from '../../core/models/usuario';
import { Auth } from '../../core/services/auth';
import { Usuarios } from '../../core/services/usuarios';
import { mensajeError } from '../../core/utils/errores';

/**
 * Formulario para crear (/usuarios/nuevo) o editar (/usuarios/:id/editar) un usuario.
 * Si la ruta trae :id, edita; si no, crea.
 */
@Component({
  imports: [ReactiveFormsModule, RouterLink],
  selector: 'app-usuario-form',
  styleUrl: './usuario-form.scss',
  templateUrl: './usuario-form.html',
})
export class UsuarioForm implements OnInit {
  private readonly usuariosService = inject(Usuarios);
  protected readonly auth = inject(Auth);
  private readonly router = inject(Router);

  /** :id de la ruta; no existe cuando se crea un usuario nuevo. */
  readonly id = input<string>();

  protected readonly esEdicion = computed(() => this.id() !== undefined);

  /** Editar a uno mismo: no se puede cambiar el propio rol (regla del backend). */
  protected readonly esUnoMismo = computed(() => Number(this.id()) === this.auth.usuario()?.id);

  /** Solo un admin cambia roles, y nunca el suyo. */
  protected readonly puedeCambiarRol = computed(() => this.auth.esAdmin() && !this.esUnoMismo());

  protected readonly form = inject(FormBuilder).nonNullable.group({
    nombre: ['', [Validators.required, Validators.maxLength(100)]],
    apellido: ['', [Validators.required, Validators.maxLength(100)]],
    email: ['', [Validators.required, Validators.email, Validators.maxLength(150)]],
    telefono: ['', Validators.maxLength(20)],
    password: ['', [Validators.minLength(8), Validators.maxLength(72)]],
    rol: ['usuario' as Rol],
  });

  protected readonly cargando = signal(false);
  protected readonly enviando = signal(false);
  protected readonly error = signal<string | null>(null);

  ngOnInit(): void {
    if (!this.esEdicion()) {
      // Al crear, la contraseña es obligatoria; al editar, solo si se quiere cambiar.
      this.form.controls.password.addValidators(Validators.required);
      this.form.controls.password.updateValueAndValidity();
      return;
    }

    if (!this.puedeCambiarRol()) {
      this.form.controls.rol.disable();
    }

    this.cargando.set(true);
    this.usuariosService.obtener(Number(this.id())).subscribe({
      next: (u) => {
        this.form.patchValue({
          nombre: u.nombre,
          apellido: u.apellido,
          email: u.email,
          telefono: u.telefono ?? '',
          rol: u.rol,
        });
        this.cargando.set(false);
      },
      error: (err) => {
        this.error.set(mensajeError(err, 'No se pudo cargar el usuario'));
        this.cargando.set(false);
      },
    });
  }

  /** true si el campo ya fue tocado y es inválido (para mostrar su mensaje). */
  protected invalido(campo: string): boolean {
    const control = this.form.get(campo);
    return !!control && control.touched && control.invalid;
  }

  /** A dónde regresa "Cancelar". */
  protected rutaCancelar(): string {
    return this.esEdicion() ? `/usuarios/${this.id()}` : '/usuarios';
  }

  protected guardar(): void {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    const v = this.form.getRawValue();
    // Se arma el cuerpo campo por campo: solo se manda lo que el backend espera.
    const datos: UsuarioPayload = {
      nombre: v.nombre.trim(),
      apellido: v.apellido.trim(),
      email: v.email.trim(),
      telefono: v.telefono.trim() || null,
    };
    if (v.password) datos.password = v.password;
    if (this.puedeCambiarRol()) datos.rol = v.rol;

    const peticion: Observable<Usuario> = this.esEdicion()
      ? this.usuariosService.actualizar(Number(this.id()), datos)
      : this.usuariosService.crear(datos);

    this.enviando.set(true);
    this.error.set(null);
    peticion.subscribe({
      next: (u) => {
        // Si alguien editó su propio perfil, la navbar debe mostrar los datos nuevos.
        if (u.id === this.auth.usuario()?.id) this.auth.actualizarUsuario(u);
        this.router.navigate(['/usuarios', u.id]);
      },
      error: (err) => {
        this.error.set(mensajeError(err, 'No se pudo guardar el usuario'));
        this.enviando.set(false);
      },
    });
  }
}