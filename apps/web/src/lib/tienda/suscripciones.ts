/**
 * Las suscripciones de la tienda (§51), en palabras: cada cuánto se cobra, el estado y la fecha del
 * próximo cobro. Lo usan las plantillas, el link privado y el panel.
 */
import type { PlanPublico, SuscripcionPublica } from './tipos.ts';

export type Frecuencia = PlanPublico['frecuencia'];
export type EstadoSuscripcion = SuscripcionPublica['estado'];

/** "por mes", "cada 2 semanas", "cada 2 meses". */
export const CADA: Record<Frecuencia, string> = {
	quincenal: 'cada 2 semanas',
	mensual: 'por mes',
	bimestral: 'cada 2 meses'
};

/** Para el selector del panel. */
export const FRECUENCIAS: [Frecuencia, string][] = [
	['mensual', 'Mensual'],
	['quincenal', 'Cada 2 semanas'],
	['bimestral', 'Cada 2 meses']
];

/** "Por mes", "Cada 2 semanas": la etiqueta del total. */
export const porPeriodo = (f: Frecuencia) => CADA[f].charAt(0).toUpperCase() + CADA[f].slice(1);

export const ESTADO_SUSCRIPCION: Record<EstadoSuscripcion, string> = {
	pendiente: 'Falta autorizar',
	activa: 'Activa',
	pausada: 'Pausada',
	cancelada: 'Cancelada'
};

/** "martes 2/12" en hora de Argentina. */
export function diaCobro(iso: string | null) {
	if (!iso) return null;
	const d = new Date(Date.parse(iso) - 3 * 3_600_000);
	const dia = ['domingo', 'lunes', 'martes', 'miércoles', 'jueves', 'viernes', 'sábado'][
		d.getUTCDay()
	];
	return `${dia} ${d.getUTCDate()}/${d.getUTCMonth() + 1}`;
}

/** Cuántos días dura un período (para el próximo cobro de la demo). */
export const diasDe = (f: Frecuencia) => (f === 'quincenal' ? 14 : f === 'mensual' ? 30 : 61);

/** Qué incluye, en una línea: el texto de la tienda o los productos con su cantidad. */
export function queIncluye(p: Pick<PlanPublico, 'incluye' | 'productos'>) {
	if (p.incluye) return p.incluye;
	return p.productos
		.map(
			(x) =>
				(x.cantidad > 1 ? `${x.cantidad} × ` : '') +
				x.nombre +
				(x.variante ? ` (${x.variante})` : '')
		)
		.join(' · ');
}

/** Cómo lo recibe, según lo que ofrece el plan. */
export function entregasDe(p: Pick<PlanPublico, 'envio' | 'retiro'>) {
	return p.envio && p.retiro ? 'Envío o retiro' : p.envio ? 'Con envío' : 'Retirás en el local';
}
