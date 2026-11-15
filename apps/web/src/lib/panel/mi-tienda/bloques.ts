/**
 * Lo que hace la lista de bloques del editor, como funciones puras: agregar entre medio, mover,
 * duplicar, ocultar y borrar. Cada una devuelve una lista nueva (el editor guarda la anterior para
 * deshacer).
 */
import type { Bloque, DisenoV2, PaginasDiseno, TipoBloque } from '#lib/tienda/sistema/catalogo.ts';
import { defDe, LIMITES, PAGINAS, type PaginaId } from './catalogo.ts';

export type { Bloque };

type ConBloques = Pick<DisenoV2, 'inicio' | 'paginas'>;

/** Los ajustes de una página con lo que falta tomado del catálogo (variante y campos). */
export function ajustesDe<P extends 'catalogo' | 'producto' | 'carrito'>(
	d: Pick<DisenoV2, 'paginas'>,
	p: P
): NonNullable<PaginasDiseno[P]> {
	const def = PAGINAS[p];
	const base: Record<string, unknown> = { variante: def.variantes?.[0] };
	for (const c of def.campos ?? []) base[c.id] = c.porDefecto;
	if (p === 'producto') base.bloques = [];
	return { ...base, ...d.paginas?.[p] } as NonNullable<PaginasDiseno[P]>;
}

/** La lista de bloques de una página (vacía si la página no tiene). */
export function bloquesDe(d: ConBloques, p: PaginaId): Bloque[] {
	if (p === 'inicio') return d.inicio;
	if (p === 'producto') return d.paginas?.producto?.bloques ?? [];
	if (p === 'nosotros') return d.paginas?.nosotros?.bloques ?? [];
	return [];
}

/** El diseño con otra lista en esa página (sin tocar el resto). */
export function conBloquesEn<T extends ConBloques>(d: T, p: PaginaId, lista: Bloque[]): T {
	if (p === 'inicio') return { ...d, inicio: lista };
	if (p === 'producto')
		return {
			...d,
			paginas: { ...d.paginas, producto: { ...ajustesDe(d, 'producto'), bloques: lista } }
		};
	if (p === 'nosotros') return { ...d, paginas: { ...d.paginas, nosotros: { bloques: lista } } };
	return d;
}

/** Todos los ids del diseño: no se repiten entre páginas. */
export function idsDe(d: ConBloques): Set<string> {
	return new Set(
		[...d.inicio, ...bloquesDe(d, 'producto'), ...bloquesDe(d, 'nosotros')].map((b) => b.id)
	);
}

const LETRAS = 'abcdefghijklmnopqrstuvwxyz0123456789';
/** Un id nuevo ([a-z0-9]{8}) que no esté usado. */
export function nuevoId(usados: Set<string>, azar = Math.random): string {
	for (;;) {
		let id = '';
		for (let i = 0; i < 8; i++) id += LETRAS[Math.floor(azar() * LETRAS.length)];
		if (!usados.has(id)) return id;
	}
}

/** Si se puede sumar otro bloque de ese tipo a la lista; si no, por qué. */
export function porQueNo(lista: Bloque[], tipo: string): string | null {
	const def = defDe(tipo);
	if (!def) return 'Ese bloque no existe.';
	if (lista.length >= LIMITES.bloquesPorPagina)
		return `Cada página va hasta ${LIMITES.bloquesPorPagina} bloques.`;
	if (def.max && lista.filter((b) => b.tipo === tipo).length >= def.max)
		return def.max === 1 ? `Ya hay un bloque de ${def.nombre.toLowerCase()}.` : 'Ya están todos.';
	return null;
}

/** Un bloque nuevo, con su primera variante (o la pedida) y sin datos: toma lo sugerido. */
export function crear(tipo: TipoBloque, usados: Set<string>, variante?: string): Bloque {
	const def = defDe(tipo);
	if (!def) throw new Error(`No existe el bloque ${tipo}`);
	return {
		id: nuevoId(usados),
		tipo,
		variante: variante && def.variantes.includes(variante) ? variante : def.variantes[0],
		datos: {}
	};
}

/** Inserta en la posición `en` (0: arriba de todo; lista.length: al final). */
export function agregar(lista: Bloque[], b: Bloque, en = lista.length): Bloque[] {
	const i = Math.max(0, Math.min(en, lista.length));
	return [...lista.slice(0, i), b, ...lista.slice(i)];
}

/** Mueve el bloque de `de` a `a` (los dos índices en la lista de antes). */
export function mover(lista: Bloque[], de: number, a: number): Bloque[] {
	if (de === a || de < 0 || de >= lista.length) return lista;
	const r = [...lista];
	const [b] = r.splice(de, 1);
	r.splice(Math.max(0, Math.min(a, r.length)), 0, b);
	return r;
}

/** Una copia del bloque justo abajo, con id nuevo. */
export function duplicar(lista: Bloque[], i: number, usados: Set<string>): Bloque[] {
	const b = lista[i];
	if (!b) return lista;
	return agregar(lista, { ...structuredClone(b), id: nuevoId(usados) }, i + 1);
}

export function alternar(lista: Bloque[], i: number): Bloque[] {
	return lista.map((b, k) => {
		if (k !== i) return b;
		const { oculto, ...resto } = b;
		return oculto ? resto : { ...resto, oculto: true };
	});
}

export function borrar(lista: Bloque[], i: number): Bloque[] {
	return lista.filter((_, k) => k !== i);
}

/** El bloque con un dato cambiado; vacío ('' / [] / undefined) lo saca, para que tome lo sugerido. */
export function conDato(b: Bloque, campo: string, valor: unknown): Bloque {
	const datos = { ...b.datos };
	if (valor === undefined || valor === '' || (Array.isArray(valor) && !valor.length))
		delete datos[campo];
	else datos[campo] = valor;
	return { ...b, datos };
}
