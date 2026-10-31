import { env } from 'cloudflare:workers';
import { aplicarCupon, normalizarCodigo, type ResultadoCupon } from '#lib/tienda/cupon.ts';
import { detalleLineas } from '#lib/tienda/lineas.ts';
import type { TiendaResuelta } from './publica.ts';

type Linea = { id: string; varianteId?: string | null; cantidad: number };

/**
 * Evalúa un cupón para un carrito de la tienda online. En la demo, con los cupones y pedidos del
 * sandbox; con la API real, lo hace Go (que vuelve a hacerlo al comprar). Los precios salen siempre
 * de la tienda, nunca del navegador.
 */
export async function evaluarCupon(
	r: TiendaResuelta,
	codigo: string,
	email: string,
	carrito: Linea[]
): Promise<ResultadoCupon> {
	const c = normalizarCodigo(codigo);
	if (!/^[A-Z0-9-]{3,20}$/.test(c))
		return { ok: false, motivo: 'Ese cupón no existe. Revisá cómo está escrito.' };
	if (r.tienda.vitrina)
		return { ok: false, motivo: 'Es una tienda de ejemplo: los cupones no se aplican.' };
	if (!r.datos) {
		const res = await fetch(`${String(env.API_URL)}/v1/publico/tiendas/${r.tienda.slug}/cupon`, {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ cupon: c, email, carrito })
		});
		if (!res.ok) return { ok: false, motivo: 'No pudimos revisar el cupón. Probá de nuevo.' };
		const g = (await res.json()) as {
			ok: boolean;
			descuento?: number;
			envioGratis?: boolean;
			combinable?: boolean;
			motivo?: string;
		};
		return g.ok
			? {
					ok: true,
					descuento: g.descuento ?? 0,
					envioGratis: !!g.envioGratis,
					combinable: !!g.combinable
				}
			: { ok: false, motivo: g.motivo ?? 'Ese cupón no vale para esta compra.' };
	}
	const cupon = await r.datos.cuponTienda(c);
	if (!cupon) return { ok: false, motivo: 'Ese cupón no existe. Revisá cómo está escrito.' };
	// Con los precios que se cobran (el de la variante y el mayorista, §49), como Go.
	const lineas = detalleLineas(
		carrito.map((l) => ({ ...l, varianteId: l.varianteId ?? null })),
		r.tienda.productos,
		r.tienda.mayorista ?? null
	).map((l) => ({
		id: l.id,
		precio: l.producto.precio,
		cantidad: l.cantidad,
		categoria: l.producto.categoria
	}));
	const h = email
		? await r.datos.historialDe(email, c)
		: { pedidos: 0, usosDelCupon: 0, vip: false, ultimaCompra: null };
	return aplicarCupon(cupon, lineas, h, new Date());
}
