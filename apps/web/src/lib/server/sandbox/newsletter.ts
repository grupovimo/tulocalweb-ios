/**
 * El newsletter en la demo (docs/decisiones.md §50). Los suscriptos de ejemplo salen de los clientes
 * de la tienda de ejemplo; en D1 sólo van los cambios del visitante (altas, confirmaciones, bajas y
 * novedades). «Mandar» guarda la novedad en el historial pero no manda ningún mail.
 */
import type { EnvioNewsletter, Newsletter } from '../datos/tipos.ts';
import type { Demo } from './demo.ts';

/** Mails de novedades por tienda por día (el mismo tope que Go). */
export const TOPE_DIARIO = 500;

export type SuscriptorDemo = {
	id: string;
	email: string;
	origen: 'tienda' | 'checkout';
	creado: number;
	confirmadoEn: number | null;
	token: string;
};

export type NovedadDemo = {
	id: string;
	asunto: string;
	creado: number;
	destinatarios: number;
};

const DIA = 86_400_000;

/** Los de ejemplo: los clientes con email (sin mayoristas), del más nuevo al más viejo. */
export function suscriptoresDemo(d: Demo, ahora: number): SuscriptorDemo[] {
	return d.clientes
		.filter((c) => c.email && !c.mayorista)
		.slice(0, 14)
		.map((c, i) => {
			const creado = ahora - (i * 3 + 1) * DIA - i * 3_600_000;
			// Los dos más nuevos todavía no confirmaron el mail.
			return {
				id: `ej-${c.id}`,
				email: c.email!.toLowerCase(),
				origen: i % 3 === 1 ? 'checkout' : 'tienda',
				creado,
				confirmadoEn: i < 2 ? null : creado + 20 * 60_000,
				token: `ejemplo-${c.id}-`.padEnd(24, 'x')
			};
		});
}

/** Una novedad de ejemplo en el historial, de hace dos semanas. */
export function novedadesDemo(ahora: number, confirmados: number): NovedadDemo[] {
	return [
		{
			id: 'ej-novedad',
			asunto: 'Llegó lo nuevo de la temporada',
			creado: ahora - 14 * DIA,
			destinatarios: Math.max(confirmados - 2, 0)
		}
	];
}

/** Junta los de ejemplo con los cambios del visitante (null: borrado). */
export function unirSuscriptores(
	ejemplo: SuscriptorDemo[],
	cambios: Map<string, unknown>
): SuscriptorDemo[] {
	const porId = new Map(ejemplo.map((s) => [s.id, s]));
	for (const [id, v] of cambios) {
		if (v) porId.set(id, v as SuscriptorDemo);
		else porId.delete(id);
	}
	return [...porId.values()].sort((a, b) => b.creado - a.creado);
}

/** El panel con los mismos campos que Go. En la demo una novedad queda enviada al toque. */
export function panelNewsletter(
	ss: SuscriptorDemo[],
	novedades: NovedadDemo[],
	inicioHoy: number,
	productos: Newsletter['productos']
): Newsletter {
	const confirmados = ss.filter((s) => s.confirmadoEn != null).length;
	const envios: EnvioNewsletter[] = [...novedades]
		.sort((a, b) => b.creado - a.creado)
		.map((n) => ({
			id: n.id,
			asunto: n.asunto,
			creado: new Date(n.creado).toISOString(),
			destinatarios: n.destinatarios,
			enviados: n.destinatarios,
			estado: 'enviado'
		}));
	const hoy = novedades
		.filter((n) => n.creado >= inicioHoy)
		.reduce((t, n) => t + n.destinatarios, 0);
	return {
		suscriptores: ss.map((s) => ({
			id: s.id,
			email: s.email,
			origen: s.origen,
			confirmado: s.confirmadoEn != null,
			creado: new Date(s.creado).toISOString(),
			confirmadoEn: s.confirmadoEn != null ? new Date(s.confirmadoEn).toISOString() : null
		})),
		confirmados,
		pendientes: ss.length - confirmados,
		envios,
		hoy: Math.min(hoy, TOPE_DIARIO),
		tope: TOPE_DIARIO,
		productos: [...productos].sort((a, b) => a.nombre.localeCompare(b.nombre, 'es')).slice(0, 300)
	};
}
