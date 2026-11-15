/**
 * El diseño v2 como lo edita el panel (docs/especificaciones/tienda-bloques.md › Modelo). El
 * editor guarda sólo lo que la tienda cambió: lo demás sale del preset y de la dirección.
 */
import type { EstiloTienda } from '#lib/tienda/estilo/estilo.ts';
import type { DisenoV2 } from '#lib/tienda/sistema/catalogo.ts';

export type { DisenoV2 };
export type EstiloV2 = EstiloTienda;

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
