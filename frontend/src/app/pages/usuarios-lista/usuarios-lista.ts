import { Component, DestroyRef, inject, OnInit, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { catchError, debounceTime, of, startWith, switchMap, tap } from 'rxjs';

import { FiltroUsuarios, Usuario } from '../../core/models/usuario';
import { Usuarios } from '../../core/services/usuarios';
import { mensajeError } from '../../core/utils/errores';

/** Lista de usuarios con buscador (solo admin). */
@Component({
  imports: [ReactiveFormsModule, RouterLink],
  selector: 'app-usuarios-lista',
  styleUrl: './usuarios-lista.scss',
  templateUrl: './usuarios-lista.html',
})
export class UsuariosLista implements OnInit {
  private readonly usuariosService = inject(Usuarios);
  private readonly destroyRef = inject(DestroyRef);

  protected readonly usuarios = signal<Usuario[]>([]);
  protected readonly cargando = signal(true);
  protected readonly error = signal<string | null>(null);

  protected readonly filtros = inject(FormBuilder).nonNullable.group({
    q: [''],
    rol: [''],
  });

  ngOnInit(): void {
    this.filtros.valueChanges
      .pipe(
        startWith(this.filtros.getRawValue()),
        debounceTime(300),
        tap(() => this.cargando.set(true)),
        switchMap(() => {
          const f = this.filtros.getRawValue();
          const filtro: FiltroUsuarios = { q: f.q.trim(), rol: f.rol as FiltroUsuarios['rol'] };
          return this.usuariosService.listar(filtro).pipe(
            catchError((err) => {
              this.error.set(mensajeError(err, 'No se pudo cargar la lista de usuarios'));
              return of(null);
            }),
          );
        }),
        takeUntilDestroyed(this.destroyRef),
      )
      .subscribe((lista) => {
        if (lista) {
          this.usuarios.set(lista);
          this.error.set(null);
        }
        this.cargando.set(false);
      });
  }

  protected limpiar(): void {
    this.filtros.setValue({ q: '', rol: '' });
  }
}