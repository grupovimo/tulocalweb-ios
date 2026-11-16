/**
 * Por qué un bloque no se va a ver en la tienda (regla «Nada vacío» del handoff): el editor lo dice
 * en la lista y en el panel, con lo que le falta. La tienda decide lo mismo al dibujar
 * (lib/tienda/sistema/datos.ts › seDibuja); acá va el motivo en palabras.
 */
import { defDe, type Campo } from './catalogo.ts';
import { VIDEO_CORTO } from '#lib/tienda/sistema/catalogo.ts';
import type { Bloque } from './bloques.ts';

/** Lo que el editor sabe de la tienda para sugerir y avisar. */
export type Contexto = {
	productos: {
		id: string;
		nombre: string;
		foto: string | null;
		precio: number;
		categoria: string | null;
		/** Tiene precio anterior (está en oferta). */
		oferta: boolean;
	}[];
	categorias: string[];
	resenas: number;
	whatsapp: boolean;
	turnos: boolean;
	direccion: boolean;
	/** Fotos de productos (las que usa la galería si no tiene propias). */
	fotos: number;
	/** Lo que la tienda ofrece de verdad, para avisar si un texto promete otra cosa. */
	ofrece?: Ofrece;
};

/** Lo que la configuración de la tienda ofrece al que compra (sale de envíos y cobros). */
export type Ofrece = {
	envioGratis: boolean;
	todoElPais: boolean;
	retiro: boolean;
	mercadoPago: boolean;
	cuotasSinInteres: boolean;
	/** Cuántas cuotas sin interés eligió en Configuración › Cobros (0: ninguna). */
	cuotasMax?: number;
	transferencia: boolean;
};

const PROMESAS: { re: RegExp; si: (o: Ofrece) => boolean; dice: string; falta: string }[] = [
	{
		re: /env[ií]os?\s+(gratis|sin\s+cargo)|gratis\s+(el\s+|tu\s+)?env[ií]o/i,
		si: (o) => o.envioGratis,
		dice: 'envío gratis',
		falta: 'no tenés un monto de envío gratis (Envíos y retiro)'
	},
	{
		re: /todo\s+el\s+pa[ií]s/i,
		si: (o) => o.todoElPais,
		dice: 'envíos a todo el país',
		falta: 'no tenés el envío por correo (Envíos y retiro)'
	},
	{
		re: /sin\s+inter[eé]s/i,
		si: (o) => o.cuotasSinInteres,
		dice: 'cuotas sin interés',
		falta: 'al pagar no se ofrecen cuotas sin interés (Configuración › Cobros)'
	},
	{
		re: /cuotas/i,
		si: (o) => o.mercadoPago,
		dice: 'cuotas',
		falta: 'no tenés Mercado Pago conectado para cobrar con tarjeta'
	},
	{
		re: /retir(o|á|a|ás|ar)(?![a-zñáéíóú])/i,
		si: (o) => o.retiro,
		dice: 'retiro',
		falta: 'no tenés el retiro en el local (Envíos y retiro)'
	},
	{
		re: /\d\s*%[^.]*transferencia|transferencia[^.]*\d\s*%/i,
		si: (o) => o.transferencia,
		dice: 'descuento por transferencia',
		falta: 'no tenés un descuento por transferencia (Cupones y promociones)'
	}
];

/** Los textos de un valor cualquiera (los datos de un bloque, los mensajes de la barra). */
function textos(v: unknown): string[] {
	if (typeof v === 'string') return [v];
	if (Array.isArray(v)) return v.flatMap(textos);
	if (v && typeof v === 'object') return Object.values(v).flatMap(textos);
	return [];
}

/**
 * Si lo que escribió la tienda promete algo que su configuración no ofrece («3 cuotas sin interés»
 * sin cuotas, «envío gratis» sin monto): no se toca su texto, pero el editor lo avisa.
 */
export function queContradice(datos: unknown, cx: Pick<Contexto, 'ofrece'>): string | null {
	const o = cx.ofrece;
	if (!o) return null;
	const todo = textos(datos).join(' \n ');
	const p = PROMESAS.find((x) => x.re.test(todo) && !x.si(o));
	if (p) return `Dice «${p.dice}», pero ${p.falta}.`;
	// «12 cuotas sin interés» con 6 elegidas: promete más de lo que se cobra.
	if (o.cuotasSinInteres && o.cuotasMax) {
		const mas = [...todo.matchAll(/(\d+)\s*cuotas\s+sin\s+inter[eé]s/gi)]
			.map((m) => Number(m[1]))
			.find((n) => n > o.cuotasMax!);
		if (mas)
			return `Dice «${mas} cuotas sin interés», pero ofrecés hasta ${o.cuotasMax} (Configuración › Cobros).`;
	}
	return null;
}

const vacio = (v: unknown) =>
	v === undefined || v === null || v === '' || (Array.isArray(v) && !v.length);

function faltaRequerido(campos: Campo[], datos: Record<string, unknown>): string | null {
	for (const c of campos) {
		const v = datos[c.id];
		if (c.requerido && vacio(v) && c.porDefecto === undefined)
			return `Le falta: ${c.nombre.toLowerCase()}.`;
		if (c.min && Array.isArray(v) && v.length && v.length < c.min)
			return `${c.nombre}: van por lo menos ${c.min}.`;
		if (c.tipo === 'lista' && c.min && vacio(v)) return `Sumá ${c.nombre.toLowerCase()}.`;
	}
	return null;
}

type Fuente = { fuente?: string; cantidad?: number; ids?: string[]; categoria?: string };

function faltanProductos(f: Fuente, cx: Contexto): string | null {
	if (!cx.productos.length) return 'Todavía no cargaste productos.';
	if (f.fuente === 'ofertas' && !cx.productos.some((p) => p.oferta))
		return 'No tenés productos en oferta.';
	if (f.fuente === 'elegidos') {
		const hay = (f.ids ?? []).filter((id) => cx.productos.some((p) => p.id === id));
		if (!hay.length) return 'Elegí qué productos van.';
	}
	if (f.fuente === 'categoria') {
		if (!f.categoria) return 'Elegí la categoría.';
		if (!cx.productos.some((p) => p.categoria === f.categoria))
			return `La categoría ${f.categoria} no tiene productos.`;
	}
	return null;
}

/** El motivo por el que el bloque no se dibuja, o null si se ve. */
export function queFalta(b: Bloque, cx: Contexto, ahora = Date.now()): string | null {
	const def = defDe(b.tipo);
	if (!def) return 'Este bloque ya no existe: borralo.';
	const d = b.datos;
	const r = faltaRequerido(def.campos, d);
	if (r) return r;
	switch (b.tipo) {
		case 'productos':
			return faltanProductos(
				(d.productos as Fuente) ??
					(def.campos.find((c) => c.id === 'productos')?.porDefecto as Fuente),
				cx
			);
		case 'oferta': {
			const hasta = Date.parse(String(d.hasta ?? ''));
			if (Number.isFinite(hasta) && hasta <= ahora) return 'La oferta ya terminó: cambiá la fecha.';
			return faltanProductos((d.productos as Fuente) ?? { fuente: 'ofertas' }, cx);
		}
		case 'destacado':
			return cx.productos.some((p) => p.id === d.producto)
				? null
				: 'Ese producto ya no está en la tienda: elegí otro.';
		case 'combo': {
			const ids = ((d.productos as string[]) ?? []).filter((id) =>
				cx.productos.some((p) => p.id === id)
			);
			return ids.length >= 2 ? null : 'Elegí por lo menos 2 productos.';
		}
		case 'categorias':
			return cx.categorias.length ? null : 'Tus productos todavía no tienen categorías.';
		case 'menu':
			return cx.productos.length ? null : 'Todavía no cargaste productos.';
		case 'resenas':
			return cx.resenas ? null : 'Todavía no tenés reseñas: aparecen con la primera.';
		case 'galeria':
			return vacio(d.fotos) && !cx.fotos ? 'Subí fotos o cargá productos con foto.' : null;
		case 'whatsapp':
			return cx.whatsapp ? null : 'Falta tu WhatsApp: cargalo en Configuración › Negocio.';
		case 'turnos':
			return cx.turnos ? null : 'Activá los turnos para que aparezca.';
		case 'local':
			return cx.direccion || !vacio(d.horarios)
				? null
				: 'Falta la dirección: cargala en Configuración › Negocio.';
		case 'videos': {
			const l = (d.items as { video?: string }[] | undefined) ?? [];
			return l.length && !l.some((x) => VIDEO_CORTO.test(String(x.video ?? '')))
				? 'Los links tienen que ser de un .mp4, un reel de Instagram, un TikTok o un Short.'
				: null;
		}
		case 'video':
			return /^https:\/\/(www\.)?(youtube\.com|youtu\.be|vimeo\.com)\//.test(String(d.url ?? ''))
				? null
				: 'El link tiene que ser de YouTube o Vimeo.';
		default:
			return null;
	}
}
