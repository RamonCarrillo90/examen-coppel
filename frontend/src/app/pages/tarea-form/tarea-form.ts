import { Component, input } from '@angular/core';

/**
 * Formulario para crear (/usuarios/:usuarioId/tareas/nueva) o editar
 * (/tareas/:id/editar) una tarea, incluida la reasignación.
 * EN CONSTRUCCIÓN: se implementa en el día 5.
 */
@Component({
  selector: 'app-tarea-form',
  styleUrl: './tarea-form.scss',
  templateUrl: './tarea-form.html',
})
export class TareaForm {
  /** :id de la tarea (solo al editar). */
  readonly id = input<string>();
  /** :usuarioId del dueño (solo al crear). */
  readonly usuarioId = input<string>();
}