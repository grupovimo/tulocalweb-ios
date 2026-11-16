/**
 * Lo que la tienda ofrece al cobrar, como lo muestran la barra, las tarjetas, la ficha y el carrito
 * (docs/especificaciones/tienda-nivel-tiendanube.md › Contrato): cuotas sin interés desde un monto,
 * el descuento por transferencia, el envío gratis y el precio sin impuestos nacionales. Lee los
 * campos con valores seguros si la API todavía no los manda (cuotas 0, impuestos apagados, IVA 21):
 * nada se promete si la tienda no lo configuró.
 */
import { IVA_INICIAL, IVAS, sinImpuestos } from '../cobros.ts';
import type { ProductoPublico, TiendaPublica } from '../tipos.ts';
import { CUOTAS_DESDE, conTransferencia, cuota, miles, precio } from './formato.ts';

type ConCobros = Pick<TiendaPublica, 'cobros'> & Partial<Pick<TiendaPublica, 'envios'>>;

export type Ofrece = {
	/** Cuotas sin interés (0: ninguna). Sólo con Mercado Pago conectado. */
	cuotas: number;
	/** Desde qué total se ofrecen. */
	cuotasDesde: number;
	/** % de descuento pagando con transferencia (0: no hay o no cargó los datos). */
	transferencia: number;
	/** Acepta transferencia (cargó los datos), con o sin descuento. */
	aceptaTransferencia: boolean;
	mercadoPago: boolean;
	/** Muestra el precio sin impuestos nacionales. */
	impuestos: boolean;
	/** Desde qué total el envío es gratis (null: no hay). */
	envioGratis: number | null;
};

const entero = (x: unknown, def: number) =>
	typeof x === 'number' && Number.isFinite(x) ? Math.round(x) : def;

export function ofreceDe(t: ConCobros): Ofrece {
	const c = t.cobros;
	const mp = !!c?.mercadoPago;
	const n = entero(c?.cuotas, 0);
	const acepta = !!c?.datosTransferencia;
	const gratis = t.envios?.gratisDesde;
	return {
		cuotas: mp && n > 1 ? Math.min(n, 24) : 0,
		cuotasDesde: Math.max(0, entero(c?.cuotasDesde, CUOTAS_DESDE)),
		transferencia: acepta ? Math.max(0, Math.min(90, entero(c?.transferencia, 0))) : 0,
		aceptaTransferencia: acepta,
		mercadoPago: mp,
		impuestos: c?.impuestos?.mostrar === true,
		envioGratis: typeof gratis === 'number' && gratis > 0 ? gratis : null
	};
}

/** Las cuotas para un total: `{ n, valor }`, o null si no hay o el total no llega al mínimo. */
export function cuotasPara(o: Ofrece, total: number) {
	if (o.cuotas < 2 || !(total > 0) || total < o.cuotasDesde) return null;
	return { n: o.cuotas, valor: cuota(total, o.cuotas) };
}

/**
 * Las cuotas en palabras: «6 cuotas sin interés de $ 31.000» (largo) o «6 x $ 31.000 sin interés»
 * (corto, para las tarjetas). Vacío si no corresponden.
 */
export function textoCuotas(o: Ofrece, total: number, forma: 'largo' | 'corto' = 'largo') {
	const c = cuotasPara(o, total);
	if (!c) return '';
	return forma === 'corto'
		? `${c.n}\u00a0x ${precio(c.valor)} sin\u00a0interés`
		: `${c.n} cuotas sin interés de ${precio(c.valor)}`;
}

/** El precio pagando con transferencia (null si no hay descuento). */
export const precioTransferencia = (o: Ofrece, total: number) =>
	o.transferencia > 0 ? conTransferencia(total, o.transferencia) : null;

/** «$ 15.300 con transferencia» (vacío sin descuento). */
export function textoTransferencia(o: Ofrece, total: number) {
	const p = precioTransferencia(o, total);
	return p === null ? '' : `${precio(p)} con transferencia`;
}

export { IVAS, sinImpuestos };

/** La alícuota de un producto (21 si no viene o no es una de las del contrato). */
export function ivaDe(p: Pick<ProductoPublico, 'iva'> | null | undefined): number {
	const x = p?.iva;
	return typeof x === 'number' && (IVAS as readonly number[]).includes(x) ? x : IVA_INICIAL;
}

/** Un precio con centavos y coma: «$ 15.289,26» (con espacio duro). */
export function precioCentavos(n: number): string {
	const c = Math.round(Math.abs(n) * 100);
	const pesos = Math.floor(c / 100);
	const cent = String(c % 100).padStart(2, '0');
	return `${n < 0 && c > 0 ? '−' : ''}$\u00a0${miles(pesos)},${cent}`;
}

/** «Precio sin impuestos nacionales $ 15.289,26», o vacío si la tienda no lo muestra. */
export function textoSinImpuestos(o: Ofrece, total: number, iva = 21) {
	if (!o.impuestos || !(total > 0)) return '';
	return `Precio sin impuestos nacionales ${precioCentavos(sinImpuestos(total, iva))}`;
}

/**
 * El total sin impuestos de varias líneas (cada una con su IVA), para el carrito: «Total sin
 * impuestos nacionales $ X». Vacío si la tienda no lo muestra.
 */
export function textoTotalSinImpuestos(
	o: Ofrece,
	lineas: { subtotal: number; iva?: number }[]
): string {
	if (!o.impuestos || !lineas.length) return '';
	const total = lineas.reduce((s, l) => s + sinImpuestos(l.subtotal, l.iva ?? 21), 0);
	return total > 0 ? `Total sin impuestos nacionales ${precioCentavos(total)}` : '';
}

/** El precio alcanza solo para el envío gratis. */
export const conEnvioGratis = (o: Ofrece, total: number) =>
	o.envioGratis !== null && total >= o.envioGratis;

/**
 * Los mensajes automáticos de la barra de anuncios, con lo que la tienda ofrece de verdad: cuotas,
 * envío gratis y descuento por transferencia. Los que ya dicen los mensajes propios no se repiten.
 */
export function mensajesAuto(o: Ofrece, propios: string[] = []): string[] {
	const ya = propios.join(' \n ').toLowerCase();
	const out: string[] = [];
	if (o.cuotas > 1 && !/cuotas/.test(ya))
		out.push(
			o.cuotasDesde > 0
				? `${o.cuotas} cuotas sin interés desde ${precio(o.cuotasDesde)}`
				: `${o.cuotas} cuotas sin interés`
		);
	if (o.envioGratis !== null && !/env[ií]os?\s+gratis/.test(ya))
		out.push(`Envío gratis desde ${precio(o.envioGratis)}`);
	if (o.transferencia > 0 && !/transferencia/.test(ya))
		out.push(`${o.transferencia}\u00a0% off pagando con transferencia`);
	return out;
}
