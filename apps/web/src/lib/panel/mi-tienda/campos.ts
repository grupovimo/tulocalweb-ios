/**
 * Ayudas de los campos del panel del bloque: los destinos de un link, las fechas en hora de
 * Argentina y el buscador de productos.
 */
import { NOMBRE_PAGINA, PAGINAS as PAGINAS_MENU } from '#lib/tienda/estilo/menu.ts';
import type { Contexto } from './faltan.ts';

/** Lo que necesitan los campos además de su valor. */
export type Ayuda = {
	cx: Contexto;
	/** El nombre y el lema de la tienda: lo sugerido si el campo queda vacío. */
	nombre: string;
	lema: string;
	/** Sube una foto como las de los productos (POST /panel/fotos) y devuelve su URL. */
	subir: (f: File) => Promise<string>;
};

export type Destino = { valor: string; nombre: string };

/** Los destinos de un link agrupados: páginas, categorías y productos. */
export function destinos(cx: Pick<Contexto, 'categorias' | 'productos'>) {
	return {
		paginas: PAGINAS_MENU.map((p) => ({ valor: p, nombre: NOMBRE_PAGINA[p] })),
		categorias: cx.categorias.map((c) => ({ valor: `cat:${c}`, nombre: c })),
		productos: cx.productos.map((p) => ({ valor: `prod:${p.id}`, nombre: p.nombre }))
	};
}

/** Si un link va a un sitio de afuera (https://…) y no a una página de la tienda. */
export const esPropio = (a: string) => a.startsWith('https://') || a === 'https:';

export const URL_HTTPS = /^https:\/\/[^\s"'<>]{3,}$/;
export const URL_VIDEO =
	/^https:\/\/(www\.)?(youtube\.com|youtu\.be|vimeo\.com|player\.vimeo\.com)\/[^\s"'<>]+$/;

const AR = -3 * 60 * 60 * 1000;
/** ISO (cualquier zona) → «YYYY-MM-DDTHH:mm» en hora de Argentina, para un datetime-local. */
export function fechaLocal(iso: unknown): string {
	const t = typeof iso === 'string' ? Date.parse(iso) : NaN;
	if (!Number.isFinite(t)) return '';
	return new Date(t + AR).toISOString().slice(0, 16);
}
/** «YYYY-MM-DDTHH:mm» de Argentina → ISO con su zona (-03:00). */
export function isoDeLocal(v: string): string {
	return /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}$/.test(v) ? `${v}:00-03:00` : '';
}

const plano = (s: string) => s.normalize('NFD').replace(/[̀-ͯ]/g, '').toLowerCase();
/** Los productos que coinciden con lo buscado (por nombre o categoría), sin los ya elegidos. */
export function buscar(
	productos: Contexto['productos'],
	q: string,
	fuera: readonly string[] = [],
	tope = 8
) {
	const t = plano(q.trim());
	return productos
		.filter((p) => !fuera.includes(p.id))
		.filter((p) => !t || plano(`${p.nombre} ${p.categoria ?? ''}`).includes(t))
		.slice(0, tope);
}
