import { tiendaDelBorrador } from '#lib/server/compartir/borrador.ts';
import type { Contexto } from '#lib/panel/mi-tienda/faltan.ts';
import { hechosDe } from '#lib/tienda/compartir.ts';
import { migrar } from '#lib/tienda/sistema/migrar.ts';
import type { PageServerLoad } from './$types';

/**
 * El editor: el borrador en v2 (si todavía es v1, migrado al vuelo) y lo que se sabe de la tienda
 * para los buscadores, lo sugerido y el aviso de qué le falta a cada bloque.
 */
export const load: PageServerLoad = async ({ locals, cookies }) => {
	const { mi, tienda } = await tiendaDelBorrador(locals, cookies);
	const contexto: Contexto = {
		productos: tienda.productos.map((p) => ({
			id: p.id,
			nombre: p.nombre,
			foto: p.foto,
			precio: p.precio,
			categoria: p.categoria,
			oferta: p.anterior != null && p.anterior > p.precio
		})),
		categorias: [
			...new Set([
				...tienda.categorias.map((c) => c.nombre),
				...mi.vista.categorias,
				...tienda.productos.flatMap((p) => (p.categoria ? [p.categoria] : []))
			])
		],
		resenas: tienda.resenas.length,
		whatsapp: !!tienda.whatsapp,
		turnos: !!tienda.turnos,
		direccion: !!tienda.direccion,
		fotos: tienda.fotos.length || tienda.productos.filter((p) => p.foto).length
	};
	return {
		...mi,
		diseno: migrar(mi.diseno, mi.plantillaId),
		contexto,
		hechos: hechosDe(tienda)
	};
};
