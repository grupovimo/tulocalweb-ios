import { MAPAS } from './mapas.ts';
import { FUENTES } from './fuentes.ts';
import type { MapaEstilo, Rol, Tono } from './tipos.ts';
import type { EstiloTienda } from './estilo.ts';

/** Los roles de los tonos: los de la tienda más «ok», que se ajusta solo. */
type RolT = Rol | 'ok';

const ROLES: Rol[] = [
	'fondo',
	'superficie',
	'texto',
	'textoSuave',
	'borde',
	'acento',
	'sobreAcento'
];
const HEX = /^#[0-9A-F]{6}$/;

/** Un color elegido válido (#RRGGBB) o null. Lo que no valida no llega nunca al CSS. */
export function colorValido(c: string | undefined | null): string | null {
	const v = (c ?? '').trim().toUpperCase();
	return HEX.test(v) ? v : null;
}

/** El mapa de una plantilla con archivo propio (las del motor se pintan con sus tokens, Motor.svelte). */
export const mapaDe = (plantilla: string): MapaEstilo | null => MAPAS[plantilla] ?? null;

function expresion(t: Tono, v: Record<RolT, string>): string {
	const base =
		t.b === null || t.p >= 100 ? v[t.a] : `color-mix(in oklab, ${v[t.a]} ${t.p}%, ${v[t.b]})`;
	return t.alfa === null
		? base
		: `color-mix(in srgb, ${base} ${Math.round(t.alfa * 1000) / 10}%, transparent)`;
}

/** Las variables de los tonos que cambian si cambian esos roles. */
function tonos(mapa: MapaEstilo, v: Record<RolT, string>, cambiados: Set<RolT>): string[] {
	const out: string[] = [];
	for (const [clave, t] of Object.entries(mapa.tonos))
		if (cambiados.has(t.a) || (t.b && cambiados.has(t.b)))
			out.push(`--tc-${clave}:${expresion(t, v)}`);
	return out;
}

const mezcla = (a: string, p: number, b: string) => `color-mix(in oklab, ${a} ${p}%, ${b})`;

/**
 * Una zona con fondo y texto propios (encabezado o pie): los tonos de adentro se recalculan con
 * esos dos colores. `inv` es el pie de las plantillas con banda propia, donde el fondo del pie es un
 * rol (casi siempre el texto) y su texto otro (casi siempre el fondo).
 */
function zona(
	mapa: MapaEstilo,
	global: Record<RolT, string>,
	cambiados: Set<RolT>,
	fondo: string | null,
	texto: string | null,
	inv: { fondo: Rol; texto: Rol } | null,
	originales: { fondo: string; texto: string } | null = null
): string[] {
	if (!fondo && !texto) return [];
	const v = { ...global };
	const c = new Set(cambiados);
	const poner = (r: Rol, x: string) => {
		v[r] = x;
		c.add(r);
	};
	if (inv) {
		if (fondo) poner(inv.fondo, fondo);
		if (texto) poner(inv.texto, texto);
	} else {
		const f = fondo ?? global.fondo;
		const t = texto ?? global.texto;
		poner('fondo', f);
		poner('texto', t);
		poner('textoSuave', mezcla(t, 68, f));
		poner('borde', mezcla(t, 14, f));
	}
	const out = tonos(mapa, v, c);
	// Los colores originales de la zona (el rosa del pie de Rosedal no es ningún rol) van exactos.
	const clave = (h: string) => h.slice(1).toLowerCase();
	if (originales && fondo) out.push(`--tc-${clave(originales.fondo)}:${fondo}`);
	if (originales && texto) out.push(`--tc-${clave(originales.texto)}:${texto}`);
	// Y la zona se pinta: hay pies sin fondo propio (Ceibo, Malbec, Pampa), que si no quedan con el
	// texto nuevo sobre el fondo de la página. Un fondo propio en línea manda igual.
	if (fondo) out.push(`background-color:${fondo}`);
	if (texto) out.push(`color:${texto}`);
	return out;
}

/** El rol que es ese color en la plantilla. */
function rolDe(mapa: MapaEstilo, color: string): Rol | null {
	return ROLES.find((r) => mapa.roles[r].toUpperCase() === color.toUpperCase()) ?? null;
}

export type CssEstilo = {
	/** Para el style del envoltorio de la tienda. */
	raiz: string;
	/** Reglas para el encabezado y el pie ([data-t-enc] y [data-t-pie]) dentro de .t-estilo. */
	reglas: string;
	/** Hoja de Google Fonts con las fuentes elegidas (null: las de la plantilla). */
	fuentes: string | null;
};

/** El CSS del estilo elegido para una plantilla con archivo propio. Vacío si no cambió nada. */
export function cssEstilo(plantilla: string, estilo: EstiloTienda | null | undefined): CssEstilo {
	const mapa = mapaDe(plantilla);
	const raiz: string[] = [];
	const reglas: string[] = [];
	const c = estilo?.colores ?? {};
	if (mapa) {
		const global: Record<RolT, string> = { ...mapa.roles, ok: mapa.ok ?? '#2F7A43' };
		const cambiados = new Set<RolT>();
		for (const r of ROLES) {
			const x = colorValido(c[r]);
			if (x && x !== mapa.roles[r].toUpperCase()) {
				global[r] = x;
				cambiados.add(r);
			}
		}
		// El verde de «en stock» y del precio con transferencia se pensó para el fondo de la plantilla:
		// si el fondo pasa de claro a oscuro (o al revés), va uno que se lea.
		if (mapa.ok && esOscuro(global.fondo) !== esOscuro(mapa.roles.fondo)) {
			global.ok = esOscuro(global.fondo) ? '#8FD4A4' : '#2F7A43';
			cambiados.add('ok');
		}
		raiz.push(...tonos(mapa, global, cambiados));
		if (cambiados.has('acento')) raiz.push(`--acento:${global.acento}`);
		const enc = zona(
			mapa,
			global,
			cambiados,
			colorValido(c.encabezado),
			colorValido(c.encabezadoTexto),
			null
		);
		if (enc.length) reglas.push(`.t-estilo [data-t-enc]{${enc.join(';')}}`);
		const inv =
			mapa.pie && rolDe(mapa, mapa.pie.fondo) && rolDe(mapa, mapa.pie.texto)
				? { fondo: rolDe(mapa, mapa.pie.fondo)!, texto: rolDe(mapa, mapa.pie.texto)! }
				: null;
		const pie = zona(
			mapa,
			global,
			cambiados,
			colorValido(c.pie),
			colorValido(c.pieTexto),
			inv,
			mapa.pie
		);
		if (pie.length) reglas.push(`.t-estilo [data-t-pie]{${pie.join(';')}}`);
	}
	const titulos = FUENTES[estilo?.fuenteTitulos ?? ''];
	const texto = FUENTES[estilo?.fuenteTexto ?? ''];
	if (titulos) raiz.push(`--tf-titulos:${titulos.css}`);
	if (texto) {
		raiz.push(`--tf-texto:${texto.css}`);
		// Sin fuente propia de títulos (Yerbal), los títulos van con la de texto.
		if (mapa && !mapa.fuentes.titulos) raiz.push(`--tf-titulos:${texto.css}`);
	}
	const familias = [titulos, texto]
		.filter((f, i, a) => f && a.indexOf(f) === i)
		.map((f) => `family=${f!.google}`);
	return {
		raiz: raiz.join(';'),
		reglas: reglas.join(''),
		fuentes: familias.length
			? `https://fonts.googleapis.com/css2?${familias.join('&')}&display=swap`
			: null
	};
}

/** El color que se ve de un rol (el elegido o el de la plantilla), para el editor y el logo. */
export function colorDe(
	plantilla: string,
	estilo: EstiloTienda | null | undefined,
	rol: Rol
): string | null {
	return colorValido(estilo?.colores?.[rol]) ?? mapaDe(plantilla)?.roles[rol] ?? null;
}

/** ¿Es oscuro? (para elegir el logo para fondos oscuros). */
export function esOscuro(hex: string | null): boolean {
	if (!hex) return false;
	const [r, g, b] = [1, 3, 5].map((i) => parseInt(hex.slice(i, i + 2), 16) / 255);
	return 0.2126 * r + 0.7152 * g + 0.0722 * b < 0.4;
}
