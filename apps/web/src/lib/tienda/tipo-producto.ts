/**
 * Tipos de producto (API: migración 00071). El editor pregunta primero qué es y según eso pide sus
 * datos: la ropa, talle y color; la comida, el contenido; un mueble, peso y medidas. La misma lista
 * está en Go (internal/dominio/listas/tipo.go) y en la app de iPhone (TipoProducto.swift).
 */

export type TipoProducto =
	'ropa' | 'calzado' | 'alimentos' | 'belleza' | 'hogar' | 'tecnologia' | 'otro';

export type Atributo =
	'peso' | 'alto' | 'ancho' | 'largo' | 'contenido' | 'material' | 'marca' | 'modelo' | 'garantia';

export type DefTipo = {
	id: TipoProducto;
	nombre: string;
	ejemplo: string;
	/** Los datos que pide, en orden. `medidas` son alto, ancho y largo juntos. */
	campos: (Atributo | 'medidas')[];
	/** Las opciones de variante que propone (con sus valores de ejemplo). */
	opciones: { nombre: string; ejemplo: string }[];
};

export const TIPOS: DefTipo[] = [
	{
		id: 'ropa',
		nombre: 'Ropa y accesorios',
		ejemplo: 'Remeras, buzos, carteras',
		campos: ['material', 'peso'],
		opciones: [
			{ nombre: 'Talle', ejemplo: 'S, M, L' },
			{ nombre: 'Color', ejemplo: 'Negro, Blanco' }
		]
	},
	{
		id: 'calzado',
		nombre: 'Calzado',
		ejemplo: 'Zapatillas, botas, sandalias',
		campos: ['material', 'peso'],
		opciones: [
			{ nombre: 'Talle', ejemplo: '38, 39, 40' },
			{ nombre: 'Color', ejemplo: 'Negro, Blanco' }
		]
	},
	{
		id: 'alimentos',
		nombre: 'Comida y bebida',
		ejemplo: 'Yerba, vinos, budines',
		campos: ['contenido', 'marca', 'peso'],
		opciones: [
			{ nombre: 'Presentación', ejemplo: '500 g, 1 kg' },
			{ nombre: 'Sabor', ejemplo: 'Limón, Naranja' }
		]
	},
	{
		id: 'belleza',
		nombre: 'Belleza y cuidado',
		ejemplo: 'Cremas, perfumes, maquillaje',
		campos: ['contenido', 'marca', 'peso'],
		opciones: [
			{ nombre: 'Presentación', ejemplo: '50 ml, 100 ml' },
			{ nombre: 'Fragancia', ejemplo: 'Vainilla, Coco' }
		]
	},
	{
		id: 'hogar',
		nombre: 'Hogar, deco y objetos',
		ejemplo: 'Muebles, mates, macetas',
		campos: ['material', 'peso', 'medidas'],
		opciones: [
			{ nombre: 'Color', ejemplo: 'Natural, Negro' },
			{ nombre: 'Tamaño', ejemplo: 'Chico, Grande' }
		]
	},
	{
		id: 'tecnologia',
		nombre: 'Tecnología',
		ejemplo: 'Celulares, auriculares, cables',
		campos: ['marca', 'modelo', 'garantia', 'peso', 'medidas'],
		opciones: [
			{ nombre: 'Color', ejemplo: 'Negro, Blanco' },
			{ nombre: 'Capacidad', ejemplo: '64 GB, 128 GB' }
		]
	},
	{
		id: 'otro',
		nombre: 'Otra cosa',
		ejemplo: 'Lo que no entra en ninguno',
		campos: ['peso', 'medidas'],
		opciones: [
			{ nombre: 'Modelo', ejemplo: 'Clásico, Premium' },
			{ nombre: 'Color', ejemplo: 'Negro, Blanco' }
		]
	}
];

export const CAMPOS: Record<
	Atributo,
	{ nombre: string; placeholder: string; unidad?: string; numero?: boolean }
> = {
	peso: { nombre: 'Peso', placeholder: '0', unidad: 'g', numero: true },
	alto: { nombre: 'Alto', placeholder: '0', unidad: 'cm', numero: true },
	ancho: { nombre: 'Ancho', placeholder: '0', unidad: 'cm', numero: true },
	largo: { nombre: 'Largo', placeholder: '0', unidad: 'cm', numero: true },
	contenido: { nombre: 'Contenido', placeholder: 'Por ejemplo: 500 g o 750 ml' },
	material: { nombre: 'Material', placeholder: 'Por ejemplo: algodón' },
	marca: { nombre: 'Marca', placeholder: 'La del fabricante o la tuya' },
	modelo: { nombre: 'Modelo', placeholder: 'Por ejemplo: A15' },
	garantia: { nombre: 'Garantía', placeholder: 'Por ejemplo: 6 meses' }
};

export const tipoDe = (id: string | null | undefined): DefTipo | null =>
	TIPOS.find((t) => t.id === id) ?? null;

/**
 * Para los productos de antes, sin tipo: lo que dicen sus opciones (un talle numérico de 34 a 46 es
 * calzado; con talle, ropa) o null si no hay pista.
 */
export function adivinarTipo(
	opciones: { nombre: string; valores: string[] }[]
): TipoProducto | null {
	const talle = opciones.find((o) => /^talle/i.test(o.nombre.trim()));
	if (!talle) return null;
	return talle.valores.length && talle.valores.every((v) => /^(3[4-9]|4[0-6])$/.test(v.trim()))
		? 'calzado'
		: 'ropa';
}
