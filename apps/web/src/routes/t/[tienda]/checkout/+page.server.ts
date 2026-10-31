import { error, fail, redirect } from '@sveltejs/kit';
import * as v from 'valibot';
import { env } from 'cloudflare:workers';
import { codigoPedido } from '#lib/server/tienda/pedido.ts';
import { COOKIE, dbSandbox } from '#lib/server/sandbox/sandbox.ts';
import { avisarVenta } from '#lib/server/telegram.ts';
import { evaluarCupon } from '#lib/server/tienda/cupon.ts';
import { resolverTienda } from '#lib/server/tienda/publica.ts';
import { cotizar, cpValido } from '#lib/tienda/envio.ts';
import { precioDe } from '#lib/tienda/variantes.ts';
import type { Actions } from './$types';
import { irAfuera } from '#lib/server/afuera.ts';
import { cuitValido, limpiarCuit } from '#lib/panel/cuit.ts';

const Pedido = v.object({
	carrito: v.pipe(
		v.array(
			v.object({
				id: v.pipe(v.string(), v.maxLength(80)),
				varianteId: v.optional(v.nullable(v.pipe(v.string(), v.maxLength(120))), null),
				cantidad: v.pipe(v.number(), v.integer(), v.minValue(1), v.maxValue(99))
			})
		),
		v.minLength(1),
		v.maxLength(50)
	),
	email: v.pipe(v.string(), v.trim(), v.email('Revisá el email.'), v.maxLength(120)),
	nombre: v.pipe(v.string(), v.trim(), v.minLength(2, 'Poné tu nombre.'), v.maxLength(80)),
	telefono: v.pipe(v.string(), v.trim(), v.regex(/^[\d\s+()-]{8,20}$/, 'Revisá el teléfono.')),
	dni: v.pipe(v.string(), v.trim(), v.maxLength(12)),
	cp: v.pipe(v.string(), v.trim(), v.maxLength(4)),
	direccion: v.pipe(v.string(), v.trim(), v.maxLength(160)),
	envio: v.picklist(
		['andreani', 'correo-domicilio', 'correo-sucursal', 'moto', 'retiro'],
		'Elegí cómo te llega.'
	),
	// Correo Argentino: la provincia (código de MiCorreo), la localidad (a domicilio) y la sucursal.
	provincia: v.optional(v.pipe(v.string(), v.trim(), v.toUpperCase(), v.maxLength(1)), ''),
	localidad: v.optional(v.pipe(v.string(), v.trim(), v.maxLength(80)), ''),
	sucursal: v.optional(v.pipe(v.string(), v.trim(), v.maxLength(20)), ''),
	pago: v.picklist(['transferencia', 'mercadopago']),
	cupon: v.optional(v.pipe(v.string(), v.trim(), v.maxLength(30)), ''),
	// Factura A (sólo si la tienda la emite): el CUIT con su verificador; la razón social, opcional.
	cuit: v.optional(
		v.pipe(
			v.string(),
			v.transform(limpiarCuit),
			v.check((c) => c === '' || cuitValido(c), 'Revisá el CUIT: son 11 números.')
		),
		''
	),
	razonSocial: v.optional(v.pipe(v.string(), v.trim(), v.maxLength(120)), '')
});

export const actions: Actions = {
	default: async ({ request, params, cookies, platform }) => {
		const r = await resolverTienda(params.tienda, cookies.get(COOKIE));
		if (!r) error(404, 'Esta tienda no existe.');
		if (r.tienda.vitrina)
			return fail(409, { error: 'Es una tienda de ejemplo: no se compra de verdad.' });
		const f = await request.formData();
		let carrito: unknown = null;
		try {
			carrito = JSON.parse(String(f.get('carrito') ?? '[]'));
		} catch {
			// queda null y falla la validación
		}
		const b = v.safeParse(Pedido, { ...Object.fromEntries(f), carrito });
		if (!b.success) return fail(400, { error: b.issues[0].message });
		const p = b.output;
		if (p.pago === 'transferencia' && !r.tienda.cobros.datosTransferencia)
			return fail(400, { error: 'Esta tienda no tiene transferencia: elegí otro medio de pago.' });
		if (p.envio !== 'retiro' && (!cpValido(p.cp) || p.direccion.length < 4))
			return fail(400, { error: 'Poné el código postal y la dirección de entrega.' });
		if (p.envio === 'correo-domicilio' && (!p.provincia || !p.localidad))
			return fail(400, { error: 'Poné la provincia y la localidad de la entrega.' });
		if (p.envio === 'correo-sucursal' && (!p.provincia || !p.sucursal))
			return fail(400, { error: 'Elegí la sucursal de Correo Argentino donde retirás.' });
		// Con la API real, el pedido lo arma Go (precios, envío y Mercado Pago salen de ahí).
		if (!r.datos || !r.sandboxId) {
			const res = await fetch(
				`${String(env.API_URL)}/v1/publico/tiendas/${r.tienda.slug}/pedidos`,
				{
					method: 'POST',
					headers: { 'content-type': 'application/json' },
					body: JSON.stringify(p)
				}
			);
			if (!res.ok) {
				const e = (await res.json().catch(() => null)) as { detail?: string } | null;
				return fail(res.status === 409 ? 409 : 400, {
					error: e?.detail ?? 'No pudimos armar el pedido.'
				});
			}
			const g = (await res.json()) as { codigo: string; url?: string };
			if (g.url) irAfuera(g.url);
			redirect(303, `/t/${r.tienda.slug}/pedido/${g.codigo}`);
		}
		// La demo no tiene cuenta de MiCorreo: Correo Argentino no se ofrece (Go lo cotiza con la API).
		if (p.envio === 'correo-domicilio' || p.envio === 'correo-sucursal')
			return fail(400, { error: 'Esta tienda no envía por Correo Argentino.' });
		// Los precios y el envío salen del servidor, nunca del navegador. Con opciones, la variante
		// tiene que ser de ese producto, estar activa (las públicas lo están) y tener stock.
		const porId = new Map(r.tienda.productos.map((x) => [x.id, x]));
		let subtotal = 0;
		for (const l of p.carrito) {
			const x = porId.get(l.id);
			if (!x)
				return fail(409, { error: 'Algún producto ya no está publicado. Revisá el carrito.' });
			if (!x.opciones.length) {
				if (l.varianteId) return fail(400, { error: 'Revisá el carrito.' });
				subtotal += x.precio * l.cantidad;
				continue;
			}
			const variante = x.variantes.find((y) => y.id === l.varianteId);
			if (!variante)
				return fail(409, {
					error: `Elegí de nuevo la variante de ${x.nombre}: la que tenías ya no está.`
				});
			if (variante.stock < l.cantidad)
				return fail(409, { error: 'Algún producto se quedó sin stock. Revisá el carrito.' });
			subtotal += precioDe(variante, x.precio) * l.cantidad;
		}
		const opcion = cotizar(p.envio === 'retiro' ? '0000' : p.cp, subtotal, r.tienda).find(
			(o) => o.id === p.envio
		);
		if (!opcion) return fail(400, { error: 'Esa forma de entrega no llega a tu código postal.' });
		// Cupón: se vuelve a evaluar acá (el carrito sólo lo mostró).
		const cupon = p.cupon ? await evaluarCupon(r, p.cupon, p.email, p.carrito) : null;
		if (cupon && !cupon.ok) return fail(409, { error: cupon.motivo });
		const descCupon = cupon?.ok ? cupon.descuento : 0;
		const envio = cupon?.ok && cupon.envioGratis ? 0 : opcion.precio;
		const descuento =
			descCupon +
			(p.pago === 'transferencia' && (!cupon || (cupon.ok && cupon.combinable))
				? Math.round(((subtotal - descCupon) * r.tienda.cobros.transferencia) / 100)
				: 0);
		const total = subtotal + envio - descuento;
		const res = await r.datos.crearPedido(
			{
				items: p.carrito.map((l) => ({
					productoId: l.id,
					varianteId: l.varianteId,
					cantidad: l.cantidad
				})),
				pago: p.pago === 'transferencia' ? 'transf' : 'mp',
				canal: 'web',
				envio: p.envio,
				cliente: p.nombre,
				telefono: p.telefono,
				nota: [
					`Email: ${p.email}`,
					p.dni && `DNI: ${p.dni}`,
					r.tienda.cobros.facturaA &&
						p.cuit &&
						`Factura A: ${p.razonSocial || p.nombre}, CUIT ${p.cuit}`,
					p.envio !== 'retiro' && `Entrega: ${p.direccion} (${p.cp})`,
					cupon?.ok && `Cupón: ${p.cupon.toUpperCase()}`,
					`Total cobrado: $ ${total.toLocaleString('es-AR')}`
				]
					.filter(Boolean)
					.join(' · '),
				// En la demo, Mercado Pago se aprueba al toque; la transferencia queda esperando el comprobante.
				pagado: p.pago === 'mercadopago'
			},
			{
				email: p.email,
				cupon: cupon?.ok ? p.cupon.toUpperCase() : null,
				descuentoCupon: descCupon,
				total,
				entrega: p.envio === 'retiro' ? null : { direccion: p.direccion, cp: p.cp }
			}
		);
		if (res === 'sin-stock')
			return fail(409, { error: 'Algún producto se quedó sin stock. Revisá el carrito.' });
		if (res === 'invalido') return fail(400, { error: 'Revisá el carrito.' });
		// Si había dejado el carrito, queda recuperado (como en Go, que lo cierra al comprar).
		await r.datos.cerrarCarritoTienda(p.email, total).catch(() => {});
		const aviso = avisarVenta(
			dbSandbox(env.SANDBOX),
			r.sandboxId,
			`🛒 Nueva venta en la tienda: pedido #${res.numero} de ${p.nombre}.`
		).catch(() => {});
		if (platform?.ctx) platform.ctx.waitUntil(aviso);
		const q = new URLSearchParams({ pago: p.pago, total: String(total) });
		redirect(303, `/t/${r.tienda.slug}/pedido/${await codigoPedido(r.sandboxId, res.numero)}?${q}`);
	}
};
