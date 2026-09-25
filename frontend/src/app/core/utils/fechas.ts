/**
 * hoyLocal devuelve la fecha de hoy como "AAAA-MM-DD" según el reloj de la computadora.
 * OJO: no se usa toISOString(), porque esa da la fecha en UTC; a las 11 p.m. en
 * Mazatlán, en UTC ya es "mañana".
 */
export function hoyLocal(): string {
  const d = new Date();
  const mes = String(d.getMonth() + 1).padStart(2, '0');
  const dia = String(d.getDate()).padStart(2, '0');
  return `${d.getFullYear()}-${mes}-${dia}`;
}