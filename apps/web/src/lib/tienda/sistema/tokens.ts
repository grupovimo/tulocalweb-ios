/**
 * Las variables CSS de una tienda: colores por rol (con el contraste AA de acentos.ts resuelto acá),
 * tipografías, forma, densidad, ancho y proporción de las fotos. Van en el style de la raíz
 * (`.tv`); todo lo demás está en las hojas de cada dirección (direcciones/*.css) y en base.css.
 */
import { contraste } from '../acentos.ts';
import { esOscuro } from '../estilo/css.ts';
import { FUENTES } from '../estilo/fuentes.ts';
import type { Direccion } from './catalogo.ts';
import type { Densidad, EstiloFinal, Forma } from './presets.ts';

// ---- Colores -----------------------------------------------------------------------------

const rgb = (h: string) => [1, 3, 5].map((i) => parseInt(h.slice(i, i + 2), 16));
const hex = (c: number[]) =>
	'#' +
	c
		.map((x) =>
			Math.round(Math.max(0, Math.min(255, x)))
				.toString(16)
				.padStart(2, '0')
		)
		.join('')
		.toUpperCase();

/** Mezcla en sRGB: p (0–1) del primero. */
export function mezclar(a: string, b: string, p: number): string {
	const [x, y] = [rgb(a), rgb(b)];
	return hex(x.map((v, i) => v * p + y[i] * (1 - p)));
}

/**
 * Un color que se lee sobre `fondo` (AA, 4,5:1 por defecto): el mismo si ya se lee; si no, se lo
 * va llevando hacia el blanco o el negro (el que más contraste dé) hasta que se lea.
 */
export function legible(color: string, fondo: string, min = 4.5): string {
	if (contraste(color, fondo) >= min) return color;
	const hacia = contraste('#FFFFFF', fondo) > contraste('#111111', fondo) ? '#FFFFFF' : '#111111';
	for (let p = 0.9; p >= 0; p -= 0.1) {
		const c = mezclar(color, hacia, p);
		if (contraste(c, fondo) >= min) return c;
	}
	return hacia;
}

/** El mejor texto para un fondo entre los candidatos (el primero que llegue a AA, o el de más contraste). */
export function textoSobre(fondo: string, candidatos: string[], min = 4.5): string {
	const ok = candidatos.find((c) => contraste(c, fondo) >= min);
	if (ok) return ok;
	return [...candidatos, '#FFFFFF', '#111111'].sort(
		(a, b) => contraste(b, fondo) - contraste(a, fondo)
	)[0];
}

export type Paleta = {
	fondo: string;
	superficie: string;
	texto: string;
	suave: string;
	borde: string;
	acento: string;
	sobreAcento: string;
	/** El acento como texto sobre el fondo (antetítulos, links): AA. */
	acentoTexto: string;
	/** Fondo suave del acento (chips, beneficios) y su texto. */
	acentoSuave: string;
	acentoSuaveTexto: string;
	/** La zona «oscuro» de los bloques (fondo oscuro) y su texto. */
	oscuro: string;
	sobreOscuro: string;
	/** El verde de «en stock», las cuotas y el precio con transferencia. */
	ok: string;
	enc: string;
	encTexto: string;
	pie: string;
	pieTexto: string;
};

export function paleta(e: EstiloFinal): Paleta {
	const c = e.colores;
	const oscuraLaTienda = esOscuro(c.fondo);
	const sobreAcento = textoSobre(c.acento, [c.sobreAcento, c.fondo, c.texto]);
	const acentoSuave = mezclar(c.acento, c.superficie, oscuraLaTienda ? 0.24 : 0.14);
	const oscuro = oscuraLaTienda
		? mezclar(c.superficie, '#000000', 0.75)
		: esOscuro(c.texto)
			? c.texto
			: '#16150F';
	return {
		fondo: c.fondo,
		superficie: c.superficie,
		texto: c.texto,
		suave: legible(c.textoSuave, c.fondo),
		borde: c.borde,
		acento: c.acento,
		sobreAcento,
		acentoTexto: legible(c.acento, c.fondo),
		acentoSuave,
		acentoSuaveTexto: legible(c.acento, acentoSuave),
		oscuro,
		sobreOscuro: textoSobre(oscuro, [oscuraLaTienda ? c.texto : c.fondo, '#F6F3EE']),
		ok: legible(oscuraLaTienda ? '#8FD4A4' : '#2F6B3E', c.superficie),
		enc: c.encabezado,
		encTexto: textoSobre(c.encabezado, [c.encabezadoTexto, c.texto, c.fondo]),
		pie: c.pie,
		pieTexto: textoSobre(c.pie, [c.pieTexto, c.texto, c.fondo])
	};
}

// ---- Forma, densidad y escala -------------------------------------------------------------

/** Radios por dirección y forma: [tarjeta, foto, botón, bloque grande]. */
const RADIOS: Record<Direccion, Record<Forma, [number, number, number, number]>> = {
	vidriera: { recta: [0, 0, 0, 0], redondeada: [8, 6, 8, 12], pildora: [12, 10, 999, 20] },
	mostrador: { recta: [4, 4, 4, 6], redondeada: [16, 12, 999, 24], pildora: [20, 14, 999, 28] },
	pedido: { recta: [2, 2, 2, 4], redondeada: [12, 10, 10, 16], pildora: [16, 12, 999, 20] }
};
/** Ritmo vertical entre bloques en la compu (README: 120 / 56–72 / 24–32). */
const RITMO: Record<Direccion, Record<Densidad, number>> = {
	vidriera: { aireada: 120, normal: 96, compacta: 72 },
	mostrador: { aireada: 88, normal: 72, compacta: 56 },
	pedido: { aireada: 40, normal: 32, compacta: 24 }
};
const MARGEN: Record<Direccion, number> = { vidriera: 56, mostrador: 40, pedido: 32 };
const VERTICAL: Record<Direccion, string> = { vidriera: '3/4', mostrador: '4/5', pedido: '4/5' };

/** El style de la raíz de la tienda: sólo variables (lo dinámico). */
export function variables(e: EstiloFinal): string {
	const p = paleta(e);
	const [rt, rf, rb, rg] = RADIOS[e.direccion][e.forma];
	const ritmo = RITMO[e.direccion][e.densidad];
	const foto =
		e.fotoProducto === 'cuadrada'
			? '1/1'
			: e.fotoProducto === 'apaisada'
				? '4/3'
				: VERTICAL[e.direccion];
	const f = (id: string, si: string) => FUENTES[id]?.css ?? si;
	const vars: Record<string, string> = {
		'--t-fondo': p.fondo,
		'--t-superficie': p.superficie,
		'--t-texto': p.texto,
		'--t-suave': p.suave,
		'--t-borde': p.borde,
		'--t-acento': p.acento,
		'--t-sobre-acento': p.sobreAcento,
		'--t-acento-texto': p.acentoTexto,
		'--t-acento-suave': p.acentoSuave,
		'--t-acento-suave-texto': p.acentoSuaveTexto,
		'--t-oscuro': p.oscuro,
		'--t-sobre-oscuro': p.sobreOscuro,
		'--t-ok': p.ok,
		'--t-enc': p.enc,
		'--t-enc-texto': p.encTexto,
		'--t-pie': p.pie,
		'--t-pie-texto': p.pieTexto,
		'--t-f-titulos': f(e.fuenteTitulos, 'serif'),
		'--t-f-texto': f(e.fuenteTexto, 'sans-serif'),
		'--t-r-tarjeta': `${rt}px`,
		'--t-r-foto': `${rf}px`,
		'--t-r-boton': `${rb}px`,
		'--t-r-grande': `${rg}px`,
		'--t-ritmo': `${ritmo}px`,
		'--t-margen': `${MARGEN[e.direccion]}px`,
		'--t-ancho': e.ancho === 'amplio' ? '1920px' : '1440px',
		'--t-foto': foto
	};
	return Object.entries(vars)
		.map(([k, v]) => `${k}:${v}`)
		.join(';');
}

/** La hoja de Google Fonts con las dos tipografías de la tienda (los pesos de la escala). */
export function hojaFuentes(e: EstiloFinal): string {
	const ids = [...new Set([e.fuenteTitulos, e.fuenteTexto])].filter((x) => FUENTES[x]);
	const familias = ids.map((x) => `family=${FUENTES[x].google}`);
	return familias.length
		? `https://fonts.googleapis.com/css2?${familias.join('&')}&display=swap`
		: '';
}
