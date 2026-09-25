import { Component, input } from '@angular/core';

/**
 * Formulario para crear (/usuarios/nuevo) o editar (/usuarios/:id/editar) un usuario.
 * EN CONSTRUCCIÓN: se implementa en el día 5.
 */
@Component({
  selector: 'app-usuario-form',
  styleUrl: './usuario-form.scss',
  templateUrl: './usuario-form.html',
})
export class UsuarioForm {
  /** :id de la ruta; no existe cuando se crea un usuario nuevo. */
  readonly id = input<string>();
}