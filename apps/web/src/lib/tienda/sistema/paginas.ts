/**
 * Lo que comparten las páginas de la tienda en bloques (paginas/*.svelte): las opciones de cada
 * página del diseño (paginas.catalogo, .producto, .carrito, .nosotros) con sus valores por defecto,
 * el tema de las piezas de opiniones (FormResena, PreguntasProducto, LanzamientoFicha, SenaPedido,
 * VacioTienda) con las variables de la tienda, y las funciones propias de algunos puntos de partida
 * (comparador de Andes, aviso de +18 de Malbec).
 */
import type { TemaOpiniones } from '../opiniones.ts';
import type { TiendaPublica } from '../tipos.ts';
import type { Bloque, Direccion, PaginasDiseno } from './catalogo.ts';
import { disenoDeTienda } from './migrar.ts';

type Opciones = Required<{
	[K in keyof PaginasDiseno]-?: NonNullable<PaginasDiseno[K]>;
}>;

/** Lo que arranca cada dirección si la tienda no eligió (A-, B- y C-ficha, -catalogo, -celular). */
const PORDEFECTO: Record<Direccion, Opciones> = {
	vidriera: {
		catalogo: { variante: 'grilla', filtros: 'costado', columnas: 4 },
		producto: { variante: 'grilla', relacionados: true, calcularEnvio: true, bloques: [] },
		carrito: { variante: 'pagina', envioGratisDesde: true },
		nosotros: { bloques: [] }
	},
	mostrador: {
		catalogo: { variante: 'grilla', filtros: 'arriba', columnas: 4 },
		producto: { variante: 'miniaturas', relacionados: true, calcularEnvio: true, bloques: [] },
		carrito: { variante: 'pagina', envioGratisDesde: true },
		nosotros: { bloques: [] }
	},
	pedido: {
		catalogo: { variante: 'lista', filtros: 'arriba', columnas: 4 },
		producto: { variante: 'grilla', relacionados: true, calcularEnvio: true, bloques: [] },
		carrito: { variante: 'pagina', envioGratisDesde: true },
		nosotros: { bloques: [] }
	}
};

/** Las opciones de las páginas de la tienda (lo elegido sobre lo de la dirección). */
export function opcionesPaginas(t: TiendaPublica, dir: Direccion): Opciones {
	const p = disenoDeTienda(t).paginas ?? {};
	const d = PORDEFECTO[dir];
	return {
		catalogo: { ...d.catalogo, ...p.catalogo },
		producto: { ...d.producto, ...p.producto, bloques: p.producto?.bloques ?? [] },
		carrito: { ...d.carrito, ...p.carrito },
		nosotros: { bloques: p.nosotros?.bloques ?? [] }
	};
}

/**
 * Los bloques de Nosotros si la tienda no armó la página: cómo compra la gente (beneficios), el
 * local, las reseñas y WhatsApp. Los que no tienen datos no se dibujan (datos.ts › seDibuja).
 */
export const NOSOTROS_BASE: Bloque[] = [
	{ id: 'nos-beneficios', tipo: 'beneficios', variante: 'fila', datos: {} },
	{ id: 'nos-local', tipo: 'local', variante: 'mapa', datos: {} },
	{ id: 'nos-resenas', tipo: 'resenas', variante: 'tarjetas', datos: {} },
	{ id: 'nos-whatsapp', tipo: 'whatsapp', variante: 'banda', datos: {} }
];

const PESO: Record<Direccion, number> = { vidriera: 400, mostrador: 800, pedido: 700 };

/** El tema de las piezas de reseñas, preguntas, lanzamiento y seña con los tokens de la tienda. */
export function temaDe(dir: Direccion): TemaOpiniones {
	return {
		texto: 'var(--t-texto)',
		suave: 'var(--t-suave)',
		campo: 'var(--t-campo, var(--t-fondo))',
		tarjeta: 'var(--t-superficie)',
		borde: 'var(--t-borde)',
		acento: 'var(--t-texto)',
		sobreAcento: 'var(--t-fondo)',
		estrella: 'var(--t-estrella)',
		radio: 'var(--t-r-campo, var(--t-r-tarjeta))',
		radioBoton: 'var(--t-r-boton)',
		fuente: 'var(--t-f-texto)',
		titulos: 'var(--t-f-titulos)',
		pesoTitulos: PESO[dir]
	};
}

/** Las funciones propias que conserva cada punto de partida (spec › Presets). */
export function funcionesDe(t: Pick<TiendaPublica, 'plantilla'>) {
	return {
		/** Andes: comparar hasta 3 productos lado a lado. */
		comparador: t.plantilla === 'andes',
		/** Malbec: «¿Sos mayor de 18 años?» antes de mostrar la tienda (Ley 24.788). */
		mayor18: t.plantilla === 'malbec'
	};
}
