/**
 * Los recordatorios del carrito abandonado (§62): hasta tres mails desde que dejó el carrito (por
 * defecto 1 h, 24 h y 72 h), con un cupón único opcional en el segundo o el tercero. Las reglas y
 * los mensajes son los de Go (ventas/recordatorios.go); el sandbox de la demo valida con esto.
 */
import type { components } from '@mostrador/contrato/esquema';

export type RecordatoriosCarritos = components['schemas']['RecordatoriosCarritos'];
export type PasoRecordatorio = RecordatoriosCarritos['pasos'][number];
export type CuponPaso = NonNullable<PasoRecordatorio['cupon']>;

export const RECORDATORIOS_POR_DEFECTO: RecordatoriosCarritos = {
	activo: false,
	pasos: [
		{ activo: true, horas: 1, cupon: null },
		{ activo: true, horas: 24, cupon: null },
		{ activo: true, horas: 72, cupon: null }
	]
};

/** Las demoras que se ofrecen (en horas), de 1 hora a 7 días. */
export const DEMORAS = [1, 2, 3, 6, 12, 24, 36, 48, 72, 96, 120, 168] as const;

/** Los cupones que se ofrecen en un toque (el editor no pide más que eso). */
export const CUPONES: { id: string; texto: string; cupon: CuponPaso | null }[] = [
	{ id: 'no', texto: 'Sin cupón', cupon: null },
	{ id: 'p5', texto: '5 % off', cupon: { descuento: 'porcentaje', valor: 5, dias: 3 } },
	{ id: 'p10', texto: '10 % off', cupon: { descuento: 'porcentaje', valor: 10, dias: 3 } },
	{ id: 'p15', texto: '15 % off', cupon: { descuento: 'porcentaje', valor: 15, dias: 3 } },
	{ id: 'envio', texto: 'Envío gratis', cupon: { descuento: 'envioGratis', valor: 0, dias: 3 } }
];

/** El id de CUPONES que corresponde a un cupón (o 'no'); uno distinto se muestra igual. */
export function idCupon(c: CuponPaso | null): string {
	if (!c) return 'no';
	if (c.descuento === 'envioGratis') return 'envio';
	return (
		CUPONES.find((x) => x.cupon?.descuento === c.descuento && x.cupon.valor === c.valor)?.id ??
		'otro'
	);
}

export function textoDemora(h: number) {
	if (h < 24) return h === 1 ? '1 hora' : `${h} horas`;
	const d = h / 24;
	return Number.isInteger(d) ? (d === 1 ? '1 día' : `${d} días`) : `${h} horas`;
}

export function textoCupon(c: CuponPaso | null) {
	if (!c) return '';
	if (c.descuento === 'envioGratis') return 'envío gratis';
	if (c.descuento === 'porcentaje') return `${c.valor} % off`;
	return `$ ${c.valor.toLocaleString('es-AR')} off`;
}

/** Lo que no deja guardar (o null), con los mismos mensajes que Go. */
export function problemaRecordatorios(r: RecordatoriosCarritos): string | null {
	if (r.pasos.length !== 3) return 'Revisá los recordatorios.';
	let previo = 0;
	for (const [i, p] of r.pasos.entries()) {
		if (!Number.isInteger(p.horas) || p.horas < 1 || p.horas > 168)
			return 'Cada recordatorio sale entre 1 hora y 7 días después.';
		if (p.horas <= previo) return 'Cada recordatorio tiene que salir después del anterior.';
		previo = p.horas;
		const c = p.cupon;
		if (!c) continue;
		if (i === 0) return 'El cupón va en el segundo o en el tercer recordatorio.';
		const ok =
			c.descuento === 'envioGratis' ||
			(c.descuento === 'porcentaje' && c.valor >= 1 && c.valor <= 90) ||
			(c.descuento === 'monto' && c.valor >= 100 && c.valor <= 100_000);
		if (!ok)
			return 'Revisá el descuento del cupón: de 1 a 90 %, de $ 100 a $ 100.000 o envío gratis.';
		if (!Number.isInteger(c.dias) || c.dias < 1 || c.dias > 30)
			return 'El cupón vence entre 1 y 30 días después.';
	}
	return null;
}
