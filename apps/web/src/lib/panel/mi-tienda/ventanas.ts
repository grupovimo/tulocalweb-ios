/**
 * Mi tienda › Promociones: armar y editar las ventanas del diseño sin mutar nada (cada cambio da un
 * diseño nuevo, como los bloques: entra al historial y se puede deshacer). Las plantillas y los
 * campos salen del catálogo (lib/tienda/sistema/promos.ts).
 */
import {
	CONTENIDO,
	MAX_VENTANAS,
	VENTANAS,
	type ContenidoVentana,
	type PlantillaVentana,
	type Ventana
} from '#lib/tienda/sistema/promos.ts';
import { precio } from '#lib/tienda/sistema/formato.ts';
import { idsDe, nuevoId } from './bloques.ts';
import type { DisenoV2 } from './diseno.ts';

export const ventanasDe = (d: DisenoV2): Ventana[] => d.ventanas ?? [];
export const conVentanas = (d: DisenoV2, l: Ventana[]): DisenoV2 => ({ ...d, ventanas: l });

/** Los ids usados en el diseño (bloques y ventanas): uno nuevo no choca con ninguno. */
export function usados(d: DisenoV2): Set<string> {
	const s = idsDe(d);
	for (const w of ventanasDe(d)) s.add(w.id);
	return s;
}

/** Reemplaza {envioGratis} por el monto de la tienda en los textos de una plantilla. */
function conMonto(x: unknown, monto: string): unknown {
	if (typeof x === 'string') return x.replaceAll('{envioGratis}', monto);
	if (Array.isArray(x)) return x.map((y) => conMonto(y, monto));
	if (x && typeof x === 'object')
		return Object.fromEntries(Object.entries(x).map(([k, v]) => [k, conMonto(v, monto)]));
	return x;
}

/**
 * Una ventana desde una plantilla (o en blanco con `null`): id nuevo, el monto del envío gratis en
 * el texto y, si trae `hastaEnDias`, la cuenta regresiva hasta ahí. Sin envío gratis configurado,
 * la frase queda con «$ X» para que se vea qué completar.
 */
export function desdePlantilla(
	p: PlantillaVentana | null,
	d: DisenoV2,
	envioGratis: number | null,
	ahora = Date.now()
): Ventana {
	const id = nuevoId(usados(d));
	if (!p)
		return {
			id,
			nombre: 'Ventana nueva',
			activa: false,
			variante: 'centrada',
			contenido: 'promo',
			datos: {},
			disparo: { cuando: 'segundos', valor: 10 },
			frecuencia: { cada: 'sesion' },
			publico: 'todos',
			paginas: [],
			dispositivos: 'todos'
		};
	const w = conMonto(
		structuredClone(p.ventana),
		envioGratis ? precio(envioGratis) : '$\u00a0X'
	) as Omit<Ventana, 'id'>;
	const datos = { ...w.datos };
	if (p.hastaEnDias) datos.hasta = new Date(ahora + p.hastaEnDias * 86_400_000).toISOString();
	return { ...w, id, datos };
}

export const puedeSumar = (d: DisenoV2) => ventanasDe(d).length < MAX_VENTANAS;

export function duplicar(l: Ventana[], i: number, d: DisenoV2): Ventana[] {
	const w = l[i];
	const copia: Ventana = {
		...structuredClone(w),
		id: nuevoId(usados(d)),
		nombre: `${w.nombre} (copia)`.slice(0, 40),
		activa: false
	};
	return [...l.slice(0, i + 1), copia, ...l.slice(i + 1)];
}

export function mover(l: Ventana[], de: number, a: number): Ventana[] {
	if (a < 0 || a >= l.length || de === a) return l;
	const r = [...l];
	const [x] = r.splice(de, 1);
	r.splice(a, 0, x);
	return r;
}

/**
 * Cambiar el contenido: se conservan los campos con el mismo id y tipo (título, texto, foto…); lo
 * demás se suelta.
 */
export function conContenido(w: Ventana, c: ContenidoVentana): Ventana {
	if (w.contenido === c) return w;
	const antes = CONTENIDO[w.contenido].campos;
	const datos = Object.fromEntries(
		CONTENIDO[c].campos
			.filter((x) => antes.some((y) => y.id === x.id && y.tipo === x.tipo))
			.filter((x) => w.datos[x.id] !== undefined)
			.map((x) => [x.id, w.datos[x.id]])
	);
	return { ...w, contenido: c, datos };
}

export const nombreContenido = (c: ContenidoVentana) => CONTENIDO[c]?.nombre ?? c;
export const nombreVariante = (v: string) =>
	VENTANAS.nombresVariantes[VENTANAS.variantes.indexOf(v as Ventana['variante'])] ?? v;

/** Un resumen de cuándo sale: «A los 8 s · Primera visita · Cada 14 días». */
export function resumen(w: Ventana): string {
	const d = w.disparo;
	const cuando =
		d.cuando === 'segundos'
			? `A los ${d.valor ?? 10}\u00a0s`
			: d.cuando === 'scroll'
				? `Al bajar el ${d.valor ?? 50}\u00a0%`
				: (VENTANAS.disparos.find((x) => x.cuando === d.cuando)?.nombre ?? d.cuando);
	const publico =
		w.publico === 'todos'
			? ''
			: VENTANAS.publicos.nombres[VENTANAS.publicos.opciones.indexOf(w.publico)];
	const f = w.frecuencia;
	const cada =
		f.cada === 'dias'
			? `Cada ${f.dias ?? 7}\u00a0días`
			: (VENTANAS.frecuencias.find((x) => x.cada === f.cada)?.nombre ?? f.cada);
	return [cuando, publico, cada].filter(Boolean).join(' · ');
}
