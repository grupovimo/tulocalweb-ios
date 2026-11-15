/**
 * La vista previa de Mi tienda (docs/especificaciones/tienda-bloques.md › Protocolo de la vista
 * previa): la tienda con un diseño del editor, sin ir al servidor, y los mensajes entre la página y
 * quien la muestra (el editor web por postMessage; las apps por evaluateJavaScript y sus puentes).
 */
import type { TiendaPublica } from '#lib/tienda/tipos.ts';
import type { DisenoV2 } from './diseno.ts';

export const PAGINAS_VISTA = ['inicio', 'catalogo', 'producto', 'carrito', 'nosotros'] as const;
export type PaginaVista = (typeof PAGINAS_VISTA)[number];
export const paginaVista = (p: unknown): PaginaVista =>
	PAGINAS_VISTA.includes(p as PaginaVista) ? (p as PaginaVista) : 'inicio';

/** Lo que manda el editor y lo que avisa la página. */
export type AlaVista =
	| { tipo: 'diseno'; diseno: DisenoV2; pagina?: PaginaVista }
	| { tipo: 'seleccionar'; id: string | null };
export type DeLaVista = { tipo: 'lista' } | { tipo: 'bloque'; id: string };

type Seccion = TiendaPublica['secciones'][number];
const SECCION: Record<string, Seccion> = {
	portada: 'hero',
	categorias: 'cats',
	productos: 'prods',
	resenas: 'revs',
	galeria: 'ig',
	whatsapp: 'wa'
};

/**
 * La tienda con ese diseño. Además del diseño v2 entero (lo que dibuja el sistema de bloques),
 * deja derivado lo de v1 (secciones, portada, anuncio) mientras convivan las plantillas viejas.
 */
export function conDiseno(t: TiendaPublica, d: DisenoV2, plantilla?: string): TiendaPublica {
	const visibles = d.inicio.filter((b) => !b.oculto);
	const secciones = visibles.flatMap((b) => (SECCION[b.tipo] ? [SECCION[b.tipo]] : []));
	if (d.estilo?.whatsappFlotante !== false && !secciones.includes('wa')) secciones.push('wa');
	const portada = visibles.find((b) => b.tipo === 'portada')?.datos ?? {};
	const fotos = Array.isArray(portada.fotos) ? (portada.fotos as string[]) : [];
	const m = d.estilo?.barra?.mensajes?.[0];
	return {
		...t,
		plantilla: d.plantilla || plantilla || t.plantilla,
		paleta: d.paleta,
		secciones: [...new Set(secciones)],
		estilo: { ...d.estilo, ...(m ? { anuncio: m.texto, anuncioLink: m.link ?? '' } : {}) },
		logo: d.logo ?? '',
		portada: {
			titulo: (portada.titulo as string) || d.titulo || t.portada.titulo,
			lema: (portada.texto as string) || d.lema || t.portada.lema,
			foto: fotos[0] || d.portada || t.portada.foto
		},
		compartir: {
			titulo: d.compartirTitulo ?? '',
			descripcion: d.compartirDescripcion ?? '',
			foto: d.compartirFoto ?? ''
		},
		diseno: d
	} as TiendaPublica;
}

/** Avisa a quien muestra la vista: el editor web, la app de iPhone o la de Android. */
export function avisar(msg: DeLaVista, w: Window = window) {
	if (w.parent && w.parent !== w) w.parent.postMessage(msg, w.location.origin);
	const ww = w as unknown as {
		webkit?: { messageHandlers?: { tienda?: { postMessage: (m: unknown) => void } } };
		TiendaPuente?: { bloque?: (id: string) => void; lista?: () => void };
	};
	ww.webkit?.messageHandlers?.tienda?.postMessage(msg);
	if (msg.tipo === 'bloque') ww.TiendaPuente?.bloque?.(msg.id);
	else ww.TiendaPuente?.lista?.();
}

/** Un mensaje del editor, si lo es (cualquier otro postMessage se ignora). */
export function leerMensaje(x: unknown): AlaVista | null {
	if (!x || typeof x !== 'object') return null;
	const m = x as { tipo?: unknown; diseno?: unknown; pagina?: unknown; id?: unknown };
	if (m.tipo === 'diseno' && m.diseno && typeof m.diseno === 'object')
		return {
			tipo: 'diseno',
			diseno: m.diseno as DisenoV2,
			...(m.pagina ? { pagina: paginaVista(m.pagina) } : {})
		};
	if (m.tipo === 'seleccionar')
		return { tipo: 'seleccionar', id: typeof m.id === 'string' ? m.id : null };
	return null;
}
