/**
 * Venta por mayor de la tienda online (docs/decisiones.md §49): cada producto (y cada variante) puede
 * tener un precio mayorista y, cuando el carrito llega al mínimo de la tienda (unidades y/o monto,
 * contra los precios minoristas), todos los que lo tienen pasan a ese precio. Es la misma regla que
 * Go (apps/api/internal/dominio/mayorista), con los mismos casos en los tests: el servidor la vuelve
 * a calcular al comprar, acá sólo se muestra.
 */

/** El mínimo del carrito (TiendaPublica.mayorista). Con los dos, tienen que cumplirse los dos. */
export type MinimoMayorista = { unidades: number | null; monto: number | null };

/** Configuración › Venta mayorista (components/schemas/VentaMayorista). */
export type VentaMayorista = MinimoMayorista & { activa: boolean };

export const MIN_UNIDADES = 2;
export const MAX_UNIDADES = 10_000;
export const MAX_MONTO = 100_000_000;

/** Lo que tiene una tienda que nunca la configuró (como la columna de Postgres). */
export const MAYORISTA_INICIAL: VentaMayorista = { activa: false, unidades: 12, monto: null };

/** Un monto no se corta: «$ 9.600» va entero en su renglón. */
const duro = (s: string) => s.replace(/ /g, '\u00a0');

const tieneMinimo = (m: MinimoMayorista | null | undefined): m is MinimoMayorista =>
	!!m && (m.unidades !== null || m.monto !== null);

/** El carrito (sus unidades y su monto a precio minorista) llega al mínimo. */
export function alcanza(m: MinimoMayorista | null | undefined, unidades: number, monto: number) {
	if (!tieneMinimo(m)) return false;
	return (m.unidades === null || unidades >= m.unidades) && (m.monto === null || monto >= m.monto);
}

/** El precio con el carrito mayorista: el mayorista si tiene y es menor (nunca sube). */
export const precioMayorista = (minorista: number, mayor: number | null | undefined) =>
	mayor && mayor > 0 && mayor < minorista ? mayor : minorista;

/**
 * El mayorista de una variante: el suyo o, si no tiene, el del producto, pero sólo si se vende al
 * precio del producto (una presentación con precio propio no hereda el mayorista de otra).
 */
export const mayoristaDeVariante = (
	variante: number | null | undefined,
	producto: number | null | undefined,
	precioVariante: number | null | undefined = null
) => variante ?? (precioVariante == null ? (producto ?? null) : null);

/** Cuánto le falta al carrito: unidades y pesos (0 si ese mínimo ya está o no existe). */
export function falta(m: MinimoMayorista, unidades: number, monto: number) {
	return {
		unidades: m.unidades === null ? 0 : Math.max(0, m.unidades - unidades),
		monto: m.monto === null ? 0 : Math.max(0, m.monto - monto)
	};
}

/**
 * El mínimo en palabras, para «Por mayor: $ X desde …»: «12 unidades», «$ 50.000 de compra» o «6
 * unidades y $ 50.000». `pesos` es el formato de la plantilla.
 */
export function textoMinimo(m: MinimoMayorista, pesos: (n: number) => string) {
	const u = m.unidades !== null ? `${m.unidades}\u00a0unidades` : '';
	if (m.monto === null) return u;
	return u ? `${u} y ${duro(pesos(m.monto))}` : `${duro(pesos(m.monto))} de\u00a0compra`;
}

/** Lo que falta en palabras («3 unidades más», «$ 4.000 más», o los dos). null: ya alcanza. */
export function textoFalta(f: { unidades: number; monto: number }, pesos: (n: number) => string) {
	const partes = [
		f.unidades ? (f.unidades === 1 ? '1\u00a0unidad' : `${f.unidades}\u00a0unidades`) : '',
		f.monto ? duro(pesos(f.monto)) : ''
	].filter(Boolean);
	return partes.length ? `${partes.join(' y ')}\u00a0más` : null;
}

/**
 * La línea discreta de la ficha: «Por mayor: $ 8.000 desde 12 unidades». null si no hay precio
 * mayorista o la tienda no vende por mayor.
 */
export function lineaPorMayor(
	mayor: number | null,
	minimo: MinimoMayorista | null | undefined,
	pesos: (n: number) => string
) {
	if (mayor === null || !minimo || (minimo.unidades === null && minimo.monto === null)) return null;
	return `Por mayor: ${duro(pesos(mayor))} desde ${textoMinimo(minimo, pesos)}`;
}

/**
 * El aviso del carrito en tres partes (la del medio va en negrita): «Sumando **3 unidades más**
 * pagás precio mayorista: ahorrás $ 9.600» o «**Precio mayorista** aplicado: ahorrás $ 9.600».
 */
export function textoAviso(
	a: { aplicado: boolean; falta: { unidades: number; monto: number }; ahorro: number },
	pesos: (n: number) => string
) {
	const ahorro = `ahorrás\u00a0${duro(pesos(a.ahorro))}`;
	const falta = textoFalta(a.falta, pesos);
	if (a.aplicado || !falta) return { pre: '', b: 'Precio mayorista', post: ` aplicado: ${ahorro}` };
	return { pre: 'Sumando ', b: falta, post: ` pagás precio mayorista: ${ahorro}` };
}

/** El problema de la configuración (null si va): al menos un mínimo, dentro de los topes. */
export function problemaMayorista(c: MinimoMayorista): string | null {
	if (c.unidades === null && c.monto === null)
		return 'Poné un mínimo de unidades, de monto o los dos.';
	if (
		(c.unidades !== null &&
			(!Number.isInteger(c.unidades) || c.unidades < MIN_UNIDADES || c.unidades > MAX_UNIDADES)) ||
		(c.monto !== null && (!Number.isInteger(c.monto) || c.monto < 1 || c.monto > MAX_MONTO))
	)
		return 'Revisá el mínimo: de 2 a 10.000 unidades y hasta $ 100.000.000.';
	return null;
}
