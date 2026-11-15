/**
 * Las ventanas de la tienda en la demo (docs/especificaciones/tienda-promociones.md), con las mismas
 * reglas que Go (dominio/promos y mitienda/ventanas.go): cuándo una ventana está completa, el sorteo
 * de la ruleta por pesos, el premio, las compras recientes y los números por ventana y día. Las
 * ventanas de las demos salen de las plantillas del catálogo (bloques.json › ventanas.plantillas),
 * como en dominio/demo.
 */
import CATALOGO from '@mostrador/contrato/bloques.json';
import type { components } from '@mostrador/contrato/esquema';

export type Ventana = components['schemas']['Ventana'];
export type Premio = components['schemas']['PromoPremio'];
export type ResultadoPromo = components['schemas']['PromoResultado'];
export type CompraReciente = components['schemas']['PromoReciente'];
export type MetricasPromos = components['schemas']['PromoMetricas'];
export type NumerosPromo = components['schemas']['PromoNumeros'];
export type Conversion = components['schemas']['ConversionTienda'];

type Datos = Record<string, unknown>;
type Campo = { id: string; tipo: string; requerido?: boolean };

const vacio = (x: unknown) =>
	x == null ||
	(typeof x === 'string' && x.trim() === '') ||
	(Array.isArray(x) && x.length === 0) ||
	(typeof x === 'object' && !Array.isArray(x) && Object.keys(x as object).length === 0);

/** El premio trae lo que necesita para entregarse (casos/ventanas.json › completa). */
export function premioCompleto(p: unknown): boolean {
	if (!p || typeof p !== 'object') return false;
	const x = p as Datos;
	switch (x.tipo) {
		case 'cuponFijo':
			return !vacio(x.cupon);
		case 'cuponUnico':
			return (
				x.descuento === 'envioGratis' ||
				((x.descuento === 'porcentaje' || x.descuento === 'monto') && x.valor != null)
			);
		case 'descarga':
			return !vacio(x.archivo);
		case 'nada':
			return true;
	}
	return false;
}

/**
 * Completa: los requeridos del contenido cargados, el premio con lo suyo y, en la ruleta, las
 * porciones completas con algún peso mayor a 0. (Que sea válida lo dice la validación del diseño.)
 */
export function ventanaCompleta(v: Ventana): boolean {
	const cont = CATALOGO.ventanas.contenidos.find((c) => c.tipo === v.contenido);
	if (!cont) return false;
	const datos = (v.datos ?? {}) as Datos;
	for (const c of cont.campos as Campo[]) {
		if (c.requerido && vacio(datos[c.id])) return false;
		if (c.tipo === 'premio' && !premioCompleto(datos[c.id])) return false;
	}
	if (v.contenido === 'ruleta') {
		const porciones = (datos.porciones ?? []) as Datos[];
		if (porciones.some((p) => !premioCompleto(p.premio))) return false;
		if (porciones.reduce((s, p) => s + (Number(p.peso) || 0), 0) <= 0) return false;
	}
	return true;
}

/** Activa, completa y dentro de sus fechas. */
export function vigente(v: Ventana, ahora: number): boolean {
	if (!v.activa || !ventanaCompleta(v)) return false;
	if (v.desde && ahora < Date.parse(v.desde)) return false;
	if (v.hasta && ahora >= Date.parse(v.hasta)) return false;
	return true;
}

/** Una porción según los pesos (-1 si ninguno es mayor a 0), como promos.Sortear. */
export function sortear(pesos: number[], azar: () => number = aleatorio): number {
	const total = pesos.reduce((s, p) => s + Math.max(p, 0), 0);
	if (total <= 0) return -1;
	let x = Math.floor(azar() * total);
	for (let i = 0; i < pesos.length; i++) {
		if (pesos[i] <= 0) continue;
		if (x < pesos[i]) return i;
		x -= pesos[i];
	}
	return pesos.length - 1;
}

function aleatorio() {
	const b = new Uint32Array(1);
	crypto.getRandomValues(b);
	return b[0] / 2 ** 32;
}

const ALFABETO = 'ABCDEFGHJKMNPQRSTUVWXYZ23456789';

/** BIENVE-7K2M9QXH: las primeras letras del nombre de la ventana y 8 al azar (como cupones.CrearUnico). */
export function codigoUnico(nombre: string): string {
	let prefijo = nombre
		.normalize('NFD')
		.replace(/[̀-ͯ]/g, '')
		.toUpperCase()
		.replace(/[^A-Z0-9]/g, '')
		.slice(0, 6);
	if (prefijo.length < 2) prefijo = 'PROMO';
	const b = new Uint8Array(8);
	crypto.getRandomValues(b);
	return prefijo + '-' + [...b].map((n) => ALFABETO[n % ALFABETO.length]).join('');
}

const pesos = (n: number) => '$ ' + Math.round(n).toLocaleString('es-AR');

/** El tipo del cupón y el texto corto del premio, como cupones.TipoYDescripcion. */
export function tipoYDescripcion(
	descuento: unknown,
	valor: number
): { tipo: 'pct' | 'fijo' | 'envio'; descripcion: string } | null {
	if (descuento === 'porcentaje' && valor >= 1 && valor <= 90)
		return { tipo: 'pct', descripcion: `${valor} % off` };
	if (descuento === 'monto' && valor >= 100 && valor <= 100_000)
		return { tipo: 'fijo', descripcion: `${pesos(valor)} off` };
	if (descuento === 'envioGratis') return { tipo: 'envio', descripcion: 'Envío gratis' };
	return null;
}

const CIUDAD_PROVINCIA: Record<string, string> = {
	CABA: 'CABA',
	Córdoba: 'Córdoba',
	'Villa Carlos Paz': 'Córdoba',
	'Río Cuarto': 'Córdoba',
	'Alta Gracia': 'Córdoba',
	'Villa Allende': 'Córdoba',
	Rosario: 'Santa Fe',
	Mendoza: 'Mendoza',
	'La Plata': 'Buenos Aires',
	'Mar del Plata': 'Buenos Aires',
	Tucumán: 'Tucumán',
	Salta: 'Salta',
	Neuquén: 'Neuquén',
	Posadas: 'Misiones',
	Lanús: 'Buenos Aires',
	Quilmes: 'Buenos Aires',
	'San Isidro': 'Buenos Aires',
	'Vicente López': 'Buenos Aires',
	Avellaneda: 'Buenos Aires',
	Morón: 'Buenos Aires',
	'Bahía Blanca': 'Buenos Aires'
};

const LETRAS: Record<string, string> = {
	A: 'Salta',
	B: 'Buenos Aires',
	C: 'CABA',
	D: 'San Luis',
	E: 'Entre Ríos',
	F: 'La Rioja',
	G: 'Santiago del Estero',
	H: 'Chaco',
	J: 'San Juan',
	K: 'Catamarca',
	L: 'La Pampa',
	M: 'Mendoza',
	N: 'Misiones',
	P: 'Formosa',
	Q: 'Neuquén',
	R: 'Río Negro',
	S: 'Santa Fe',
	T: 'Tucumán',
	U: 'Chubut',
	V: 'Tierra del Fuego',
	W: 'Corrientes',
	X: 'Córdoba',
	Y: 'Jujuy',
	Z: 'Santa Cruz'
};

/** La provincia de una ciudad de la demo o de un CPA; '' si no se sabe (como promos.provinciaDe). */
export function provinciaDe(ciudad: string, cp = ''): string {
	const c = cp.trim().toUpperCase();
	if (/^[A-Z]\d{4}([A-Z]{3})?$/.test(c)) return LETRAS[c[0]] ?? '';
	return CIUDAD_PROVINCIA[ciudad.trim()] ?? '';
}

/** La primera palabra, si parece un nombre («Consumidor final» no sale). */
export function nombrePila(n: string): string {
	const t = n.trim();
	if (!t || /^consumidor final$/i.test(t) || t.includes('@')) return '';
	const p = t.split(/\s+/)[0];
	if (p.length < 2 || p.length > 20 || !/^[\p{L}'-]+$/u.test(p)) return '';
	return p[0].toUpperCase() + p.slice(1).toLowerCase();
}

/** estilo.conversion.pruebaSocial.activa (por defecto, apagada). */
export function pruebaSocialActiva(estilo: { conversion?: Conversion } | undefined | null) {
	return !!estilo?.conversion?.pruebaSocial?.activa;
}

const DIA = 86_400_000;

/** El día de Argentina (YYYY-MM-DD) de un instante. */
export const diaAR = (t: number) => new Date(t - 3 * 3_600_000).toISOString().slice(0, 10);

/** Rango del panel: vacío, los últimos 30 días; hasta un año (como promos.Rango). */
export function rango(desde: string | null, hasta: string | null, ahora: number) {
	const fecha = /^\d{4}-\d{2}-\d{2}$/;
	const h = hasta || diaAR(ahora);
	if (!fecha.test(h)) return null;
	const d = desde || diaAR(Date.parse(h + 'T12:00:00Z') - 29 * DIA);
	if (!fecha.test(d)) return null;
	const dias = (Date.parse(h) - Date.parse(d)) / DIA;
	if (!(dias >= 0) || dias > 366) return null;
	return { desde: d, hasta: h };
}

// ---------------------------------------------------------------- Demos

type Plantilla = (typeof CATALOGO.ventanas.plantillas)[number];

/** Una plantilla del catálogo como ventana, con su id y el envío gratis de la tienda. */
export function deLaPlantilla(
	plantilla: string,
	id: string,
	gratisDesde: number | null,
	ahora: number
): Ventana | null {
	const p = CATALOGO.ventanas.plantillas.find((x) => x.id === plantilla) as Plantilla | undefined;
	if (!p) return null;
	const v = structuredClone(p.ventana) as unknown as Ventana;
	const datos = v.datos as Datos;
	const monto = gratisDesde ? pesos(gratisDesde) : '';
	for (const k of ['titulo', 'texto', 'etiqueta'])
		if (typeof datos[k] === 'string')
			datos[k] = (datos[k] as string).replace('{envioGratis}', monto);
	if ('hastaEnDias' in p && p.hastaEnDias)
		datos.hasta = new Date(ahora + (p.hastaEnDias as number) * DIA).toISOString();
	return { ...v, id };
}

/**
 * Lo que estrenan las demos (como demo.Diseno en Go): Lumbre (ceibo) con la Bienvenida 10 %, Café
 * Bruma (mostaza) con la ruleta y Almacén Verde (yerbal) con el envío gratis y la prueba social.
 */
export function promosDemo(
	plantilla: string,
	gratisDesde: number | null,
	ahora: number
): { ventanas: Ventana[]; conversion?: Conversion } {
	const una = (p: string, id: string) => {
		const v = deLaPlantilla(p, id, gratisDesde, ahora);
		return v ? [v] : [];
	};
	switch (plantilla) {
		case 'ceibo':
			return { ventanas: una('bienvenida', 'bienvenida') };
		case 'mostaza':
			return { ventanas: una('ruleta', 'ruletabruma') };
		case 'yerbal':
			return {
				ventanas: gratisDesde ? una('envio-gratis', 'enviogratis') : [],
				conversion: { envioGratis: true, pruebaSocial: { activa: true, cada: 25, paginas: [] } }
			};
	}
	return { ventanas: [] };
}

// ---------------------------------------------------------------- Métricas

export type MetricaDia = { vistas: number; cierres: number; clics: number };
export type Captacion = { ventana: string; creado: number; premio: Premio; porcion?: number };
export type VentaPromo = { cupon: string; creado: number; total: number; estado: string };

const ceros = (): NumerosPromo => ({
	vistas: 0,
	cierres: 0,
	clics: 0,
	envios: 0,
	ventas: 0,
	monto: 0
});

/** Los números del panel con las mismas cuentas que promos.LeerMetricas. */
export function metricas(
	ventanas: Ventana[],
	eventos: Map<string, MetricaDia>,
	captaciones: Captacion[],
	ventas: VentaPromo[],
	unicos: Map<string, string>,
	desde: string,
	hasta: string
): MetricasPromos {
	const por = new Map<string, Map<string, NumerosPromo>>();
	const sumar = (v: string, dia: string, f: (n: NumerosPromo) => void) => {
		if (dia < desde || dia > hasta || !v) return;
		if (!por.has(v)) por.set(v, new Map());
		const m = por.get(v)!;
		if (!m.has(dia)) m.set(dia, ceros());
		f(m.get(dia)!);
	};
	for (const [clave, e] of eventos) {
		const [v, dia] = clave.split(':');
		sumar(v, dia, (n) => {
			n.vistas += e.vistas;
			n.cierres += e.cierres;
			n.clics += e.clics;
		});
	}
	for (const c of captaciones) sumar(c.ventana, diaAR(c.creado), (n) => n.envios++);
	const fijos = new Map<string, string>();
	for (const v of ventanas) {
		const d = v.datos as Datos;
		const premio = d.premio as Datos | undefined;
		const codigo =
			v.contenido === 'cupon'
				? d.cupon
				: v.contenido === 'captacion' && premio?.tipo === 'cuponFijo'
					? premio.cupon
					: null;
		if (typeof codigo === 'string' && codigo && !fijos.has(codigo)) fijos.set(codigo, v.id);
	}
	for (const p of ventas) {
		if (p.estado === 'Cancelado') continue;
		const v = unicos.get(p.cupon) ?? fijos.get(p.cupon);
		if (v)
			sumar(v, diaAR(p.creado), (n) => {
				n.ventas++;
				n.monto += p.total;
			});
	}
	const out: MetricasPromos = { desde, hasta, total: ceros(), ventanas: [] };
	const agregar = (
		id: string,
		nombre: string,
		contenido: string,
		activa: boolean,
		borrada: boolean
	) => {
		const dias = [...(por.get(id) ?? new Map<string, NumerosPromo>())]
			.map(([dia, n]) => ({ dia, ...n }))
			.sort((a, b) => a.dia.localeCompare(b.dia));
		const n = ceros();
		for (const d of dias) for (const k of Object.keys(n) as (keyof NumerosPromo)[]) n[k] += d[k];
		for (const k of Object.keys(n) as (keyof NumerosPromo)[]) out.total[k] += n[k];
		out.ventanas.push({
			id,
			nombre,
			contenido,
			activa,
			borrada,
			...n,
			tasaEnvio: n.vistas > 0 ? n.envios / n.vistas : null,
			dias
		});
	};
	const vistas = new Set<string>();
	for (const v of ventanas) {
		if (vistas.has(v.id)) continue;
		vistas.add(v.id);
		agregar(v.id, v.nombre, v.contenido, v.activa, false);
	}
	for (const id of [...por.keys()].filter((x) => !vistas.has(x)).sort())
		agregar(id, id, '', false, true);
	return out;
}
