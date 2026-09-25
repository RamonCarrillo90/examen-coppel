import { Component, inject, OnInit, signal } from '@angular/core';
import { RouterLink } from '@angular/router';

import { Usuario } from '../../core/models/usuario';
import { Usuarios } from '../../core/services/usuarios';

/** Lista de todos los usuarios (solo admin). */
@Component({
  imports: [RouterLink],
  selector: 'app-usuarios-lista',
  styleUrl: './usuarios-lista.scss',
  templateUrl: './usuarios-lista.html',
})
export class UsuariosLista implements OnInit {
  private readonly usuariosService = inject(Usuarios);

  protected readonly usuarios = signal<Usuario[]>([]);
  protected readonly cargando = signal(true);
  protected readonly error = signal<string | null>(null);

  ngOnInit(): void {
    this.usuariosService.listar().subscribe({
      next: (lista) => {
        this.usuarios.set(lista);
        this.cargando.set(false);
      },
      error: () => {
        this.error.set('No se pudo cargar la lista de usuarios');
        this.cargando.set(false);
      },
    });
  }
}