/**
 * El catálogo de bloques visto desde el editor (packages/contrato/bloques.json): las definiciones
 * con las que se arman la biblioteca, las variantes y los campos de cada panel. La validación y
 * los valores por defecto de la tienda están en lib/tienda/sistema/catalogo.ts.
 */
import BLOQUES from '@mostrador/contrato/bloques.json';

export type TipoCampo = keyof typeof BLOQUES.tiposCampo;

export type Campo = {
	id: string;
	tipo: TipoCampo;
	nombre: string;
	max?: number;
	min?: number;
	requerido?: boolean;
	porDefecto?: unknown;
	/** De dónde sale el texto si se deja vacío: el nombre o el lema de la tienda. */
	sugerido?: 'nombre' | 'lema';
	opciones?: string[];
	nombres?: string[];
	item?: Campo[];
};

export type DefBloque = {
	tipo: string;
	grupo: string;
	nombre: string;
	descripcion: string;
	max?: number;
	variantes: string[];
	nombresVariantes: string[];
	campos: Campo[];
};

export type PaginaId = 'inicio' | 'catalogo' | 'producto' | 'carrito' | 'nosotros';

export type DefPagina = {
	nombre: string;
	bloques?: boolean;
	variantes?: string[];
	nombresVariantes?: string[];
	campos?: Campo[];
};

export const CATALOGO = BLOQUES;
export const LIMITES = BLOQUES.limites;
export const DEFS = BLOQUES.bloques as DefBloque[];
export const GRUPOS = BLOQUES.grupos;
export const FONDOS = BLOQUES.fondos;
export const ICONOS = BLOQUES.iconos;
export const FUENTES_PRODUCTOS = BLOQUES.fuentesProductos;
export const PAGINAS = BLOQUES.paginas as Record<PaginaId, DefPagina>;
export const ORDEN_PAGINAS: PaginaId[] = ['inicio', 'catalogo', 'producto', 'carrito', 'nosotros'];
export const BARRA = BLOQUES.barraAnuncios as {
	nombre: string;
	descripcion: string;
	variantes: string[];
	nombresVariantes: string[];
	campos: Campo[];
};
export const OPCIONES_ESTILO = BLOQUES.estilo as Record<
	'direccion' | 'forma' | 'densidad' | 'ancho' | 'fotoProducto' | 'encabezado' | 'pie',
	{ opciones: string[]; nombres: string[]; porDefecto?: string }
>;

const POR_TIPO = new Map(DEFS.map((d) => [d.tipo, d]));
export const defDe = (tipo: string) => POR_TIPO.get(tipo);

/** El nombre de una variante («Foto completa»), o el id si no está. */
export function nombreVariante(tipo: string, variante: string) {
	const d = defDe(tipo);
	return d?.nombresVariantes[d.variantes.indexOf(variante)] ?? variante;
}

/** Las páginas que se arman con bloques. */
export const conBloques = (p: PaginaId) => !!PAGINAS[p].bloques;
