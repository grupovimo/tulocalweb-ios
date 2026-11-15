/**
 * El diseño v2 como lo edita el panel (docs/especificaciones/tienda-bloques.md › Modelo). El
 * editor guarda sólo lo que la tienda cambió: lo demás sale del preset y de la dirección.
 */
import type { EstiloTienda } from '#lib/tienda/estilo/estilo.ts';
import type { Bloque, Paginas } from './bloques.ts';

export type Fondo = 'normal' | 'superficie' | 'acento' | 'oscuro';
export type MensajeBarra = { texto: string; link?: string };

export type EstiloV2 = Omit<EstiloTienda, 'textos' | 'anuncio' | 'anuncioLink'> & {
	direccion?: 'vidriera' | 'mostrador' | 'pedido';
	forma?: 'recta' | 'redondeada' | 'pildora';
	densidad?: 'aireada' | 'normal' | 'compacta';
	ancho?: 'normal' | 'amplio';
	fotoProducto?: 'cuadrada' | 'vertical' | 'apaisada';
	encabezado?: 'izquierda' | 'centrado' | 'buscador';
	pie?: 'simple' | 'columnas' | 'grande';
	barra?: { variante: 'fija' | 'rotativa'; mensajes: MensajeBarra[]; fondo?: Fondo };
	whatsappFlotante?: boolean;
};

export type DisenoV2 = {
	version: 2;
	plantilla?: string;
	paleta: number;
	titulo?: string;
	lema?: string;
	portada?: string;
	logo?: string;
	compartirTitulo?: string;
	compartirDescripcion?: string;
	compartirFoto?: string;
	estilo?: EstiloV2;
	inicio: Bloque[];
	paginas?: Paginas;
};

/** Lo que el estilo trae del preset y de la dirección: al cambiar el punto de partida se suelta. */
export const ESTILO_DEL_PRESET = [
	'direccion',
	'forma',
	'densidad',
	'ancho',
	'fotoProducto',
	'encabezado',
	'pie',
	'fuenteTitulos',
	'fuenteTexto',
	'colores'
] as const satisfies readonly (keyof EstiloV2)[];

/** Saca del diseño lo vacío (strings vacíos, objetos sin claves) para que pese poco. */
export function limpiar<T>(v: T): T {
	if (Array.isArray(v)) return v.map(limpiar) as T;
	if (v && typeof v === 'object') {
		const r: Record<string, unknown> = {};
		for (const [k, x] of Object.entries(v)) {
			const l = limpiar(x);
			if (l === undefined || l === '') continue;
			// Un bloque siempre lleva `datos`, aunque esté vacío.
			if (
				k !== 'datos' &&
				l &&
				typeof l === 'object' &&
				!Array.isArray(l) &&
				!Object.keys(l).length
			)
				continue;
			r[k] = l;
		}
		return r as T;
	}
	return v;
}
