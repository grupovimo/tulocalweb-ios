/**
 * Deshacer y rehacer del editor: una pila de estados ya guardados. Lo que se escribe en un campo
 * entra recién al confirmarlo (al salir del campo), no letra por letra.
 */
export type Historial<T> = { pasado: T[]; futuro: T[] };

export const TOPE_HISTORIAL = 60;

export const vacio = <T>(): Historial<T> => ({ pasado: [], futuro: [] });

/** Anota el estado de antes de un cambio. Un cambio nuevo borra lo que se podía rehacer. */
export function registrar<T>(h: Historial<T>, antes: T, tope = TOPE_HISTORIAL): Historial<T> {
	return { pasado: [...h.pasado, antes].slice(-tope), futuro: [] };
}

/** Vuelve al estado anterior; `actual` queda para rehacer. null si no hay nada. */
export function deshacer<T>(h: Historial<T>, actual: T): { h: Historial<T>; estado: T } | null {
	if (!h.pasado.length) return null;
	const estado = h.pasado[h.pasado.length - 1];
	return { h: { pasado: h.pasado.slice(0, -1), futuro: [...h.futuro, actual] }, estado };
}

export function rehacer<T>(h: Historial<T>, actual: T): { h: Historial<T>; estado: T } | null {
	if (!h.futuro.length) return null;
	const estado = h.futuro[h.futuro.length - 1];
	return { h: { pasado: [...h.pasado, actual], futuro: h.futuro.slice(0, -1) }, estado };
}
