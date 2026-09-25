import { HttpErrorResponse } from '@angular/common/http';

/**
 * mensajeError devuelve el mensaje que mandó el backend ({"error": "..."}) cuando
 * es un error "del usuario" (400, 403, 404, 409). Para cualquier otro caso
 * (500, backend apagado...) devuelve el mensaje genérico que le pases.
 */
export function mensajeError(err: unknown, porDefecto: string): string {
  const esDelUsuario =
    err instanceof HttpErrorResponse && [400, 403, 404, 409].includes(err.status);
  if (esDelUsuario && typeof err.error?.error === 'string') {
    return err.error.error;
  }
  return porDefecto;
}