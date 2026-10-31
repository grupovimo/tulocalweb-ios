import { datosDe } from '#lib/server/datos/index.ts';
import type { VarianteFila } from '#lib/panel/vendibles.ts';
import type { PageServerLoad } from './$types';

/** Suscripciones (§51): los planes, los suscriptores y los productos para armar un plan. */
export const load: PageServerLoad = async ({ locals }) => {
	const datos = datosDe(locals);
	const [panel, productos] = await Promise.all([
		datos.suscripcionesPanel(),
		datos.listar('productos', { cuantos: 200 })
	]);
	return {
		...panel,
		productos: (
			productos.filas as {
				id: string;
				nombre: string;
				foto: string | null;
				precio: number;
				variantes: VarianteFila[];
			}[]
		).map((p) => ({
			id: p.id,
			nombre: p.nombre,
			foto: p.foto,
			precio: p.precio,
			variantes: (p.variantes ?? [])
				.filter((v) => v.activa)
				.map((v) => ({ id: v.id, texto: v.texto }))
		}))
	};
};
