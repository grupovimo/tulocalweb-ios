/**
 * De qué período habla la pregunta: hoy, ayer, anteayer, esta semana (7 días), la semana pasada,
 * este mes (30 días), últimos N días, un día de la semana («el lunes») o una fecha («el 28»).
 * Los días puntuales y la semana pasada salen de las ventas por día (14 días); el mes, del tablero
 * de 30 días.
 */
import type { Pregunta } from './texto.ts';

export type Periodo =
	| { tipo: 'hoy' }
	| { tipo: 'dia'; atras: number; nombre: string }
	| { tipo: 'semana' }
	| { tipo: 'semanaPasada' }
	| { tipo: 'mes' }
	| { tipo: 'ultimos'; dias: number };

const DIAS = ['domingo', 'lunes', 'martes', 'miercoles', 'jueves', 'viernes', 'sabado'];
const NUMEROS: Record<string, number> = {
	dos: 2,
	tres: 3,
	cuatro: 4,
	cinco: 5,
	seis: 6,
	siete: 7,
	diez: 10,
	catorce: 14,
	quince: 15,
	treinta: 30
};

/** El día de la semana de hoy en Argentina (0 = domingo) y el número del día del mes. */
function hoyAr(ahora: Date) {
	const f = new Intl.DateTimeFormat('en-US', {
		timeZone: 'America/Argentina/Buenos_Aires',
		weekday: 'short',
		day: 'numeric'
	}).formatToParts(ahora);
	const dia = Number(f.find((x) => x.type === 'day')?.value ?? 1);
	const sem = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'].indexOf(
		f.find((x) => x.type === 'weekday')?.value ?? 'Sun'
	);
	return { dia, sem };
}

export function periodoDe(p: Pregunta, ahora = new Date()): Periodo | null {
	const q = p.q;
	if (/\banteayer\b|antes de ayer/.test(q)) return { tipo: 'dia', atras: 2, nombre: 'anteayer' };
	if (/\bayer\b/.test(q)) return { tipo: 'dia', atras: 1, nombre: 'ayer' };
	if (/semana pasada|semana anterior|la otra semana/.test(q)) return { tipo: 'semanaPasada' };
	const n = q.match(/ultim[oa]s? (\d{1,2}|[a-z]+) dias/);
	if (n) {
		const dias = Number(n[1]) || NUMEROS[n[1]] || 0;
		if (dias === 1) return { tipo: 'hoy' };
		if (dias === 7) return { tipo: 'semana' };
		if (dias >= 28) return { tipo: 'mes' };
		if (dias >= 2) return { tipo: 'ultimos', dias: Math.min(dias, 14) };
	}
	if (/quincena|dos semanas|14 dias/.test(q)) return { tipo: 'ultimos', dias: 14 };
	if (/\bmes\b|30 dias|treinta dias|mensual/.test(q)) return { tipo: 'mes' };
	if (/\bsemana\b|7 dias|siete dias|semanal/.test(q)) return { tipo: 'semana' };
	if (/\bhoy\b|en el dia|al momento|hasta ahora|ahora/.test(q)) return { tipo: 'hoy' };
	const { dia, sem } = hoyAr(ahora);
	// «el lunes», «el martes pasado»: el último que pasó (hoy no cuenta).
	for (const [i, nombre] of DIAS.entries())
		if (new RegExp(`\\b${nombre}\\b`).test(q)) {
			const atras = (sem - i + 7) % 7 || 7;
			return {
				tipo: 'dia',
				atras,
				nombre: `el ${nombre === 'miercoles' ? 'miércoles' : nombre === 'sabado' ? 'sábado' : nombre}`
			};
		}
	// «el 28», «el 28/10»: un día de este mes (o del anterior, si todavía no llegó).
	const f = q.match(/\bel (\d{1,2})(?:\/\d{1,2})?\b/);
	if (f) {
		const d = Number(f[1]);
		if (d >= 1 && d <= 31) {
			const atras = d <= dia ? dia - d : dia + (30 - d);
			if (atras === 0) return { tipo: 'hoy' };
			if (atras <= 13) return { tipo: 'dia', atras, nombre: `el ${d}` };
		}
	}
	return null;
}

export const nombrePeriodo = (p: Periodo): string => {
	switch (p.tipo) {
		case 'hoy':
			return 'hoy';
		case 'dia':
			return p.nombre;
		case 'semana':
			return 'en los últimos 7 días';
		case 'semanaPasada':
			return 'la semana pasada';
		case 'mes':
			return 'en los últimos 30 días';
		case 'ultimos':
			return `en los últimos ${p.dias} días`;
	}
};
