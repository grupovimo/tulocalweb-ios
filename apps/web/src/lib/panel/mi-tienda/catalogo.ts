/**
 * El catálogo de bloques visto desde el editor: lo de la tienda (lib/tienda/sistema/catalogo.ts)
 * más los nombres que muestra el panel (variantes, opciones, páginas y la barra de anuncios), todo
 * desde packages/contrato/bloques.json.
 */
import CATALOGO from '@mostrador/contrato/bloques.json';
import { DEF, type Campo, type DefBloque } from '#lib/tienda/sistema/catalogo.ts';

export { BLOQUES as DEFS, LIMITES, ICONOS } from '#lib/tienda/sistema/catalogo.ts';
export type { Campo, DefBloque };

/** Un campo con los nombres de sus opciones (los `opcion` de las páginas). */
export type CampoConNombres = Campo & { nombres?: string[] };

export type PaginaId = 'inicio' | 'catalogo' | 'producto' | 'carrito' | 'nosotros';

export type DefPagina = {
	nombre: string;
	bloques?: boolean;
	variantes?: string[];
	nombresVariantes?: string[];
	campos?: CampoConNombres[];
};

export const GRUPOS = CATALOGO.grupos;
export const FONDOS = CATALOGO.fondos;
export const FUENTES_PRODUCTOS = CATALOGO.fuentesProductos;
export const PAGINAS = CATALOGO.paginas as Record<PaginaId, DefPagina>;
export const ORDEN_PAGINAS: PaginaId[] = ['inicio', 'catalogo', 'producto', 'carrito', 'nosotros'];
export const BARRA = CATALOGO.barraAnuncios as {
	nombre: string;
	descripcion: string;
	variantes: string[];
	nombresVariantes: string[];
	campos: Campo[];
};
export type OpcionEstilo =
	'direccion' | 'forma' | 'densidad' | 'ancho' | 'fotoProducto' | 'encabezado' | 'pie';
export const OPCIONES_ESTILO = CATALOGO.estilo as Record<
	OpcionEstilo,
	{ opciones: string[]; nombres: string[]; porDefecto?: string }
>;

export const defDe = (tipo: string): DefBloque | undefined => DEF[tipo as keyof typeof DEF];

/** El nombre de una variante («Foto completa»), o el id si no está. */
export function nombreVariante(tipo: string, variante: string) {
	const d = defDe(tipo);
	return d?.nombresVariantes[d.variantes.indexOf(variante)] ?? variante;
}

/** Las páginas que se arman con bloques. */
export const conBloques = (p: PaginaId) => !!PAGINAS[p].bloques;
