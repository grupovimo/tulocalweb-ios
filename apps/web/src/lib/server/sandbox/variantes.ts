/**
 * Opciones y variantes de la tienda de ejemplo (moda, calzado, cosmética, gastronomía y regionales),
 * armadas desde el catálogo (src/lib/catalogos.ts, que es generado y no se toca). Talles, números,
 * colores y presentaciones verosímiles, con stock por variante. Son siempre iguales para un producto
 * (semilla por id), así no corren la semilla de los pedidos de demo.ts.
 */
import type { Catalogo, ProductoEjemplo } from '#lib/catalogos.ts';
import type { OpcionProducto } from '#lib/tienda/tipos.ts';
import { combinaciones, type Valores, type Variante } from '#lib/tienda/variantes.ts';

export type VariantesDemo = { opciones: OpcionProducto[]; fotos: string[]; variantes: Variante[] };

const hash = (s: string) =>
	[...s].reduce((h, c) => Math.imul(h ^ c.charCodeAt(0), 16777619), 2166136261) >>> 0;
const semilla = (s: string) => (hash(s) % 10_000) / 10_000;

const TALLES = ['S', 'M', 'L', 'XL'];
const redondo = (n: number) => Math.round(n / 100) * 100;

type Regla = {
	si: (p: ProductoEjemplo, c: Catalogo) => boolean;
	opciones: (p: ProductoEjemplo) => OpcionProducto[];
	/** Precio de la variante (null: el del producto). */
	precio?: (v: Valores, p: ProductoEjemplo) => number | null;
};

const en =
	(...ids: string[]) =>
	(p: ProductoEjemplo) =>
		ids.includes(p.id);
const deCategoria =
	(...cs: string[]) =>
	(p: ProductoEjemplo) =>
		!!p.categoria && cs.includes(p.categoria);
const plantilla =
	(...ps: string[]) =>
	(_: ProductoEjemplo, c: Catalogo) =>
		ps.includes(c.plantilla);
const y =
	(...fs: ((p: ProductoEjemplo, c: Catalogo) => boolean)[]) =>
	(p: ProductoEjemplo, c: Catalogo) =>
		fs.every((f) => f(p, c));

/** Presentación por peso: el precio del producto es el de la más chica. */
const porPeso = (valores: string[], factores: number[]) => ({
	opciones: () => [{ nombre: 'Presentación', valores }],
	precio: (v: Valores, p: ProductoEjemplo) => {
		const i = valores.indexOf(v['Presentación']);
		return i > 0 ? redondo(p.precio * factores[i]) : null;
	}
});

const REGLAS: Regla[] = [
	// ---- Moda (Ceibo, Pampa)
	{
		si: en('remera-pima', 'remera-boxy-negra', 'remera-boxy-blanca'),
		opciones: (p) => [
			{ nombre: 'Talle', valores: TALLES },
			...(p.id === 'remera-pima'
				? [{ nombre: 'Color', valores: ['Blanco', 'Negro', 'Arena'] }]
				: [])
		],
		// El XL lleva más tela.
		precio: (v, p) => (v['Talle'] === 'XL' ? redondo(p.precio * 1.12) : null)
	},
	{
		si: en('sweater-trenzado', 'buzo-heavy-frisa', 'tapado-merino-ceibo'),
		opciones: (p) => [
			{ nombre: 'Talle', valores: TALLES },
			{
				nombre: 'Color',
				valores:
					p.id === 'buzo-heavy-frisa' ? ['Negro', 'Gris melange'] : ['Crudo', 'Camel', 'Negro']
			}
		]
	},
	{
		si: en('jean-recto-delta', 'cargo-ripstop'),
		opciones: () => [{ nombre: 'Talle', valores: ['38', '40', '42', '44', '46'] }]
	},
	{
		si: y(
			plantilla('ceibo', 'pampa'),
			deCategoria('Abrigos', 'Básicos', 'Buzos', 'Remeras', 'Pantalones', 'Camperas')
		),
		opciones: () => [{ nombre: 'Talle', valores: TALLES }]
	},
	// ---- Calzado
	{
		si: en('stiletto-noche'),
		opciones: () => [
			{ nombre: 'Número', valores: ['35', '36', '37', '38', '39', '40'] },
			{ nombre: 'Color', valores: ['Negro', 'Nude'] }
		]
	},
	{
		si: deCategoria('Zapatillas'),
		opciones: () => [{ nombre: 'Número', valores: ['38', '39', '40', '41', '42', '43', '44'] }]
	},
	{
		si: y(plantilla('ceibo', 'pampa'), deCategoria('Accesorios', 'Gorras')),
		opciones: (p) => [
			{
				nombre: 'Color',
				valores: p.categoria === 'Gorras' ? ['Negro', 'Arena', 'Verde'] : ['Suela', 'Negro']
			}
		]
	},
	// ---- Cosmética (Rosedal)
	{
		si: en('serum-de-rosa-mosqueta', 'contorno-de-ojos'),
		...porPeso(['15 ml', '30 ml'], [1, 1.8])
	},
	{
		si: en('aceite-corporal', 'tonico-de-rosas', 'limpiador-en-gel'),
		...porPeso(['100 ml', '250 ml'], [1, 2.1])
	},
	{
		si: en('balsamo-labial'),
		opciones: () => [{ nombre: 'Aroma', valores: ['Rosa', 'Vainilla', 'Neutro'] }]
	},
	// ---- Gastronomía (Mostaza, Miga, Yerbal)
	{
		si: deCategoria('Café en grano'),
		opciones: () => [
			{ nombre: 'Presentación', valores: ['250 g', '500 g', '1 kg'] },
			{ nombre: 'Molienda', valores: ['En grano', 'Filtro', 'Espresso'] }
		],
		precio: (v, p) => {
			const f = [1, 1.9, 3.6][['250 g', '500 g', '1 kg'].indexOf(v['Presentación'])];
			return f > 1 ? redondo(p.precio * f) : null;
		}
	},
	{
		si: en('cafe-de-especialidad'),
		...porPeso(['250 g', '500 g', '1 kg'], [1, 1.9, 3.6])
	},
	{
		si: y(plantilla('mostaza'), en('flat-white', 'capuccino', 'cold-brew')),
		opciones: () => [{ nombre: 'Tamaño', valores: ['Chico', 'Mediano', 'Grande'] }],
		precio: (v, p) =>
			v['Tamaño'] === 'Mediano'
				? redondo(p.precio * 1.2)
				: v['Tamaño'] === 'Grande'
					? redondo(p.precio * 1.4)
					: null
	},
	{
		si: en('medialuna-de-manteca'),
		...porPeso(['Unidad', 'Media docena', 'Docena'], [1, 5.5, 10])
	},
	{
		si: y(plantilla('yerbal'), deCategoria('Frutos secos')),
		...porPeso(['250 g', '500 g', '1 kg'], [1, 1.9, 3.6])
	},
	{
		si: en('arroz-yamani', 'lentejas', 'semillas-de-chia', 'avena-arrollada'),
		...porPeso(['500 g', '1 kg'], [1, 1.85])
	},
	// ---- Regionales (Litoral)
	{
		si: y(plantilla('litoral'), deCategoria('Yerbas')),
		...porPeso(['500 g', '1 kg'], [1, 1.85])
	},
	{
		si: en('mate-de-calabaza'),
		opciones: () => [{ nombre: 'Color', valores: ['Natural', 'Negro', 'Marrón'] }]
	},
	{
		si: en('termo-acero-1-l'),
		opciones: () => [{ nombre: 'Color', valores: ['Negro', 'Verde', 'Blanco'] }]
	},
	// ---- Tecnología (Andes)
	{
		si: en('auricular-over-ear-anc', 'in-ear-inalambricos'),
		opciones: () => [{ nombre: 'Color', valores: ['Negro', 'Blanco', 'Azul marino'] }]
	},
	// ---- Vinos (Malbec): la caja de 6 con 10% menos por botella.
	{
		si: en('malbec-valle-de-uco-2021', 'malbec-reserva-lujan'),
		opciones: () => [{ nombre: 'Presentación', valores: ['Botella', 'Caja x 6'] }],
		precio: (v, p) => (v['Presentación'] === 'Caja x 6' ? redondo(p.precio * 5.4) : null)
	},
	// ---- Muebles (Quebracho)
	{
		si: en('silla-quebracho', 'butaca-pampa'),
		opciones: () => [{ nombre: 'Color', valores: ['Natural', 'Chocolate', 'Negro'] }]
	},
	{
		si: en('mesa-ratona-paraiso'),
		opciones: () => [{ nombre: 'Medida', valores: ['80 cm', '100 cm', '120 cm'] }],
		precio: (v, p) =>
			v['Medida'] === '100 cm'
				? redondo(p.precio * 1.2)
				: v['Medida'] === '120 cm'
					? redondo(p.precio * 1.4)
					: null
	},
	// ---- Las plantillas del motor: mascotas, artesanías, relojes, cerámica y anteojos.
	{
		si: en('cama-redonda'),
		opciones: () => [{ nombre: 'Talle', valores: ['S', 'M', 'L'] }],
		precio: (v, p) => (v['Talle'] === 'L' ? redondo(p.precio * 1.3) : null)
	},
	{
		si: en('correa-extensible', 'taza-arena', 'mate-de-ceramica'),
		opciones: (p) => [
			{
				nombre: 'Color',
				valores:
					p.id === 'correa-extensible' ? ['Negro', 'Rojo', 'Azul'] : ['Arena', 'Blanco', 'Verde']
			}
		]
	},
	{
		si: en('tabla-de-asado-lenga', 'tabla-de-picada'),
		opciones: () => [{ nombre: 'Medida', valores: ['30 × 20 cm', '40 × 25 cm'] }],
		precio: (v, p) => (v['Medida'] === '40 × 25 cm' ? redondo(p.precio * 1.35) : null)
	},
	{
		si: en('automatico-clasico-40', 'cuarzo-minimal'),
		opciones: () => [{ nombre: 'Malla', valores: ['Cuero negro', 'Cuero suela', 'Acero'] }],
		precio: (v, p) => (v['Malla'] === 'Acero' ? redondo(p.precio * 1.15) : null)
	},
	{
		si: en('sol-aviador-dorado', 'sol-wayfarer'),
		opciones: (p) => [
			{
				nombre: 'Color',
				valores: p.id === 'sol-wayfarer' ? ['Negro', 'Carey'] : ['Dorado', 'Plateado']
			}
		]
	}
];

/**
 * Las variantes de ejemplo de un catálogo, por producto. Los demás productos no tienen opciones.
 * En cada catálogo, el primer producto con talle o número tiene agotado el primero (para que la
 * demo muestre un talle tachado); el resto, stock al azar con alguno agotado.
 */
export function variantesDemo(catalogo: Catalogo): Map<string, VariantesDemo> {
	const r = new Map<string, VariantesDemo>();
	let tachado = false;
	for (const p of catalogo.productos) {
		const regla = REGLAS.find((x) => x.si(p, catalogo));
		if (!regla) continue;
		const opciones = regla.opciones(p);
		// La galería sólo muestra este producto: su foto. Ningún color toma la foto de otro modelo
		// (antes salían las de su categoría); sin foto propia, la variante queda con la principal.
		const fotos = [p.foto];
		const talle = opciones.find((o) => o.nombre === 'Talle' || o.nombre === 'Número');
		// La remera de Ceibo, la de la prueba de punta a punta, también tiene el S agotado.
		const agotar = talle && (!tachado || p.id === 'remera-pima') ? talle.valores[0] : null;
		if (agotar) tachado = true;
		const variantes = combinaciones(opciones).map((valores, i): Variante => {
			const s = semilla(`${p.id}:${i}`);
			return {
				id: `${p.id}-v${i + 1}`,
				valores,
				sku: `${p.id.slice(0, 10).toUpperCase()}-${Object.values(valores)
					.map((x) => x.replace(/\s+/g, '').slice(0, 3).toUpperCase())
					.join('-')}`,
				precio: regla.precio?.(valores, p) ?? null,
				stock:
					agotar && valores[talle!.nombre] === agotar ? 0 : s < 0.14 ? 0 : 2 + Math.floor(s * 14),
				foto: null,
				activa: true
			};
		});
		r.set(p.id, { opciones, fotos, variantes });
	}
	return r;
}
