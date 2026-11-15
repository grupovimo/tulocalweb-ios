/**
 * Genera src/lib/catalogos.ts: el catálogo de la tienda de ejemplo de cada una de las 16 plantillas,
 * sacado del handoff. Doce vienen de design/Plantillas Tienda.dc.html (const ST: tienda, ciudad,
 * categorías y productos); Ceibo, Mostaza, Yerbal y Tijera tienen su propio archivo trabajado.
 * Es la base de los datos de la demo (productos, pedidos, clientes) y del storefront.
 * Uso: bun scripts/catalogos.ts
 */
import { readFileSync, writeFileSync } from 'node:fs';
import { conMarca } from './marca.ts';
import { join } from 'node:path';
import { runInNewContext } from 'node:vm';

const DISENO = join(import.meta.dirname, '../../../design');
const leer = (f: string) => readFileSync(join(DISENO, f), 'utf8');

/** Hasta dónde llega el literal que empieza en `desde` ([ o {), respetando comillas. */
function fin(fuente: string, desde: number) {
	let prof = 0;
	let comilla = '';
	for (let k = desde; k < fuente.length; k++) {
		const c = fuente[k];
		if (comilla) {
			if (c === '\\') k++;
			else if (c === comilla) comilla = '';
			continue;
		}
		if (c === "'" || c === '"' || c === '`') comilla = c;
		else if (c === '[' || c === '{') prof++;
		else if (c === ']' || c === '}') if (--prof === 0) return k + 1;
	}
	throw new Error('literal sin cerrar');
}

/** Evalúa un literal de JS del handoff (son archivos locales del diseño). */
function literal<T>(fuente: string, inicio: string): T {
	const i = fuente.indexOf(inicio);
	if (i < 0) throw new Error(`no está: ${inicio}`);
	const a = i + inicio.length - 1;
	const ctx: { __v?: T } = {};
	runInNewContext(`__v = ${fuente.slice(a, fin(fuente, a))}`, ctx);
	return ctx.__v as T;
}

/** Un literal dentro de un objeto de renderVals ("clave: [[…]]"). */
function lista<T>(fuente: string, clave: string): T {
	return literal<T>(fuente, `${clave}: [`);
}

type Producto = {
	id: string;
	nombre: string;
	categoria: string | null;
	precio: number;
	anterior: number | null;
	/** Id de la foto en Unsplash o URL completa (las fotos propias de static/demo). */
	foto: string;
	etiqueta: string | null;
	detalle: string | null;
};
type Catalogo = {
	plantilla: string;
	tienda: string;
	rubro: string;
	ciudad: string;
	categorias: string[];
	productos: Producto[];
};

const idDe = (s: string) =>
	s
		.normalize('NFD')
		.replace(/[̀-ͯ]/g, '')
		.toLowerCase()
		.replace(/[^a-z0-9]+/g, '-')
		.replace(/^-|-$/g, '')
		.slice(0, 40);

function producto(p: Omit<Producto, 'id'>): Producto {
	return { id: idDe(p.nombre), ...p, anterior: p.anterior || null, etiqueta: p.etiqueta || null };
}

const catalogos: Catalogo[] = [];

// Doce tiendas del motor de plantillas.
type St = {
	store: string;
	rubro: string;
	city: string;
	cats: [string, string, string][];
	prods: [string, number, number, string, string | 0][];
};
const ST = literal<Record<string, St>>(leer('Plantillas Tienda.dc.html'), 'const ST = {');
for (const [id, t] of Object.entries(ST)) {
	catalogos.push({
		plantilla: id,
		tienda: t.store,
		rubro: t.rubro,
		ciudad: t.city,
		categorias: t.cats.map((c) => c[0]),
		productos: t.prods.map(([n, p, ant, foto, tag]) =>
			producto({
				nombre: n,
				categoria: null,
				precio: p,
				anterior: ant,
				foto,
				etiqueta: tag || null,
				detalle: null
			})
		)
	});
}

// Ceibo (moda).
{
	const s = leer('Plantilla Ceibo.dc.html');
	const P = literal<[string, number, number | null, string, string[], string | null][]>(
		s,
		'const P = ['
	);
	catalogos.push({
		plantilla: 'ceibo',
		tienda: 'Tienda Lumbre',
		rubro: 'Moda',
		ciudad: 'CABA',
		categorias: ['Abrigos', 'Básicos', 'Accesorios'],
		productos: P.map(([n, p, ant, foto, , tag]) =>
			producto({
				nombre: n,
				categoria: null,
				precio: p,
				anterior: ant,
				foto,
				etiqueta: tag,
				detalle: null
			})
		)
	});
}

// Yerbal (almacén natural).
{
	const s = leer('Plantilla Yerbal.dc.html');
	const IMG = literal<Record<string, string>>(s, 'const IMG = {');
	const P = literal<[string, string, number, number | null, string][]>(s, 'const P = [');
	catalogos.push({
		plantilla: 'yerbal',
		tienda: 'Almacén Verde',
		rubro: 'Almacén natural',
		ciudad: 'Córdoba',
		categorias: ['Frutos secos', 'Almacén', 'Orgánicos'],
		productos: P.map(([n, img, p, ant, unidad]) =>
			producto({
				nombre: n,
				categoria: null,
				precio: p,
				anterior: ant,
				foto: IMG[img],
				etiqueta: null,
				detalle: unidad
			})
		)
	});
}

// Mostaza (café de especialidad): el menú, el café en grano y lo que acompaña.
{
	const s = leer('Plantilla Mostaza.dc.html');
	const IMG = literal<Record<string, string>>(s, 'const IMG = {');
	const menu = lista<[string, [string, string, number, string, string?][]][]>(s, 'menu');
	const beans = lista<[string, string, string, number, string][]>(s, 'beans');
	const productos: Producto[] = [];
	for (const [cat, items] of menu)
		for (const [n, d, p, img, tag] of items)
			productos.push(
				producto({
					nombre: n,
					categoria: cat,
					precio: p,
					anterior: null,
					foto: IMG[img],
					etiqueta: tag ?? null,
					detalle: d
				})
			);
	for (const [n, origen, notas, p, img] of beans)
		productos.push(
			producto({
				nombre: `Café ${n}`,
				categoria: 'Café en grano',
				precio: p,
				anterior: null,
				foto: IMG[img],
				etiqueta: null,
				detalle: `${origen} · ${notas}`
			})
		);
	catalogos.push({
		plantilla: 'mostaza',
		tienda: 'Café Bruma',
		rubro: 'Café de especialidad',
		ciudad: 'CABA',
		categorias: [...menu.map(([c]) => c), 'Café en grano'],
		productos
	});
}

// Tijera (barbería): los productos de la tienda (los turnos son servicios, llegan con la agenda).
{
	const s = leer('Plantilla Tijera.dc.html');
	const PR = literal<[string, string, number, string, string | null][]>(s, 'const PR = [');
	catalogos.push({
		plantilla: 'tijera',
		tienda: 'Navaja',
		rubro: 'Barbería',
		ciudad: 'CABA',
		categorias: ['Pelo', 'Barba', 'Kits'],
		productos: PR.map(([marca, n, p, foto, tag]) =>
			producto({
				nombre: marca === 'Navaja Studio' ? n : `${marca} ${n}`,
				categoria: null,
				precio: p,
				anterior: null,
				foto,
				etiqueta: tag,
				detalle: marca
			})
		)
	});
}

/*
 * Categoría de cada producto (por posición). El handoff da las categorías de la tienda pero no en
 * cuál va cada producto; cuando uno no entra en ninguna, se suma una categoría al final.
 */
const CATEGORIAS: Record<string, string[]> = {
	andes: [
		'Auriculares',
		'Parlantes',
		'In-ear',
		'Auriculares',
		'Parlantes',
		'Auriculares',
		'In-ear',
		'Auriculares'
	],
	ceibo: [
		'Abrigos',
		'Abrigos',
		'Básicos',
		'Básicos',
		'Básicos',
		'Abrigos',
		'Accesorios',
		'Accesorios',
		'Básicos',
		'Abrigos',
		'Básicos',
		'Básicos'
	],
	hora: [
		'Automáticos',
		'Cuarzo',
		'Automáticos',
		'Automáticos',
		'Cuarzo',
		'Mallas',
		'Automáticos',
		'Cuarzo'
	],
	lente: [
		'De sol',
		'De sol',
		'Recetados',
		'De sol',
		'Recetados',
		'De sol',
		'Accesorios',
		'Accesorios'
	],
	litoral: ['Yerbas', 'Mates', 'Mates', 'Mates', 'Dulces', 'Mates', 'Yerbas', 'Dulces'],
	malbec: ['Tintos', 'Tintos', 'Tintos', 'Blancos', 'Espumantes', 'Tintos', 'Rosados', 'Cajas'],
	miga: ['Panes', 'Facturas', 'Dulces', 'Facturas', 'Panes', 'Dulces', 'Facturas', 'Panes'],
	pampa: [
		'Buzos',
		'Zapatillas',
		'Remeras',
		'Gorras',
		'Pantalones',
		'Zapatillas',
		'Camperas',
		'Remeras'
	],
	patio: ['Perros', 'Perros', 'Juguetes', 'Gatos', 'Perros', 'Perros', 'Gatos', 'Gatos'],
	quebracho: [
		'Living',
		'Comedor',
		'Living',
		'Iluminación',
		'Living',
		'Living',
		'Iluminación',
		'Comedor'
	],
	rosedal: ['Rostro', 'Rostro', 'Rostro', 'Kits', 'Cuerpo', 'Cuerpo', 'Rostro', 'Rostro'],
	taller: ['Cocina', 'Cocina', 'Mesa', 'Mesa', 'Regalos', 'Mesa', 'Mesa', 'Regalos'],
	tijera: ['Pelo', 'Barba', 'Pelo', 'Pelo', 'Barba', 'Kits', 'Pelo', 'Pelo'],
	torno: ['Tazas', 'Platos', 'Jarrones', 'Platos', 'Tazas', 'Tazas', 'Platos', 'Jarrones'],
	yerbal: [
		'Frutos secos',
		'Frutos secos',
		'Frutos secos',
		'Frutos secos',
		'Almacén',
		'Orgánicos',
		'Almacén',
		'Orgánicos',
		'Almacén',
		'Orgánicos',
		'Almacén',
		'Orgánicos',
		'Frutos secos',
		'Almacén'
	]
};
for (const c of catalogos) {
	const cats = CATEGORIAS[c.plantilla];
	if (!cats) continue;
	if (cats.length !== c.productos.length)
		throw new Error(
			`${c.plantilla}: ${cats.length} categorías para ${c.productos.length} productos`
		);
	c.productos.forEach((p, i) => (p.categoria = cats[i]));
	for (const k of cats) if (!c.categorias.includes(k)) c.categorias.push(k);
}

// Ids únicos dentro de cada catálogo.
for (const c of catalogos) {
	const vistos = new Map<string, number>();
	for (const p of c.productos) {
		const n = (vistos.get(p.id) ?? 0) + 1;
		vistos.set(p.id, n);
		if (n > 1) p.id = `${p.id}-${n}`;
	}
}
/*
 * Fotos que muestran el producto: en el handoff muchas no coincidían (una sandía para la miel, un
 * lavarropas para la tabla de asado). fotos-demo.json las reemplaza por plantilla y id de producto:
 * un id de Unsplash que ya estaba en el catálogo o una foto libre propia en static/demo (créditos en
 * static/demo/CREDITOS.md). La API usa las mismas en internal/dominio/demo/catalogos.json.
 */
const FOTOS: Record<string, Record<string, string>> = JSON.parse(
	readFileSync(join(import.meta.dirname, 'fotos-demo.json'), 'utf8')
);
for (const c of catalogos)
	for (const p of c.productos) {
		const foto = FOTOS[c.plantilla]?.[p.id];
		if (foto) p.foto = foto;
	}
for (const [plantilla, fotos] of Object.entries(FOTOS))
	for (const id of Object.keys(fotos))
		if (!catalogos.some((c) => c.plantilla === plantilla && c.productos.some((p) => p.id === id)))
			throw new Error(`fotos-demo.json: no existe ${plantilla}/${id}`);
// Miga Clásica (la otra versión de Miga en el handoff) vende lo mismo que Miga.
const miga = catalogos.find((c) => c.plantilla === 'miga');
if (miga)
	catalogos.push({
		...structuredClone(miga),
		plantilla: 'miga-clasica',
		tienda: 'Panadería Lola Clásica'
	});
catalogos.sort((a, b) => a.plantilla.localeCompare(b.plantilla));

const salida = `// Generado por scripts/catalogos.ts desde el handoff (Plantillas Tienda y las plantillas trabajadas). No editar a mano.
/** Un producto de la tienda de ejemplo. Precios en pesos. */
export type ProductoEjemplo = {
	id: string;
	nombre: string;
	categoria: string | null;
	precio: number;
	anterior: number | null;
	/** Id de la foto en Unsplash o URL completa (las fotos propias de static/demo). */
	foto: string;
	etiqueta: string | null;
	detalle: string | null;
};
/** El catálogo de la tienda de ejemplo de una plantilla. */
export type Catalogo = {
	plantilla: string;
	tienda: string;
	rubro: string;
	ciudad: string;
	categorias: string[];
	productos: ProductoEjemplo[];
};

export const CATALOGOS: Catalogo[] = ${JSON.stringify(catalogos, null, '\t')};

export const catalogoDe = (plantilla: string) => CATALOGOS.find((c) => c.plantilla === plantilla);
`;
writeFileSync(join(import.meta.dirname, '../src/lib/catalogos.ts'), conMarca(salida));
console.log(
	catalogos.map((c) => `${c.plantilla}: ${c.tienda} · ${c.productos.length} productos`).join('\n')
);
