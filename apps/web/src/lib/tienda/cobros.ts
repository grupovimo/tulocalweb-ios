/**
 * Cuotas sin interés y precio sin impuestos nacionales (Ley 27.743), lo que la tienda elige en
 * Configuración › Cobros (docs/especificaciones/tienda-nivel-tiendanube.md). La misma regla que Go
 * (apps/api/internal/dominio/cobros): la usan el panel, el sandbox y la tienda.
 */

/** Las cuotas que se pueden ofrecer (0: ninguna). */
export const CUOTAS = [0, 3, 6, 9, 12] as const;
export type Cuotas = (typeof CUOTAS)[number];
/** Las de una tienda que nunca eligió (como la columna de Postgres) y el «desde» por defecto, en pesos. */
export const CUOTAS_INICIAL = 6;
export const DESDE_INICIAL = 30_000;
export const MAX_DESDE = 100_000_000;

/** Las alícuotas de IVA de un producto (21 por defecto). */
export const IVAS = [21, 10.5, 27, 0] as const;
export type Iva = (typeof IVAS)[number];
export const IVA_INICIAL: Iva = 21;

/** La configuración de MP donde la tienda activa las cuotas sin interés que absorbe. */
export const MP_CUOTAS = 'https://www.mercadopago.com.ar/costs-section';

/**
 * La condición fiscal de Configuración › Negocio es texto libre («Responsable inscripta»,
 * «responsable_inscripto», «RI»): todo lo demás (monotributo, exento, vacío) no discrimina IVA.
 */
export function responsableInscripto(condicion: string | null | undefined) {
	const c = (condicion ?? '').trim().toLowerCase();
	return c.includes('responsable') || c.includes('inscript') || c === 'ri';
}

/** El motivo por el que no se guarda, o null (los mismos topes que Go). */
export function problemaCobros(c: { cuotas: number; cuotasDesde: number }) {
	if (!(CUOTAS as readonly number[]).includes(c.cuotas))
		return 'Elegí ninguna, 3, 6, 9 o 12 cuotas.';
	if (!Number.isInteger(c.cuotasDesde) || c.cuotasDesde < 0 || c.cuotasDesde > MAX_DESDE)
		return 'El monto va de $ 0 a $ 100.000.000.';
	return null;
}

/** Precio sin impuestos nacionales: precio / (1 + iva/100), redondeado a centavos. */
export const sinImpuestos = (precio: number, iva: number = IVA_INICIAL) =>
	Math.round((precio / (1 + iva / 100)) * 100) / 100;
