import { env } from 'cloudflare:workers';
import { error, fail } from '@sveltejs/kit';
import { COOKIE } from '#lib/server/sandbox/sandbox.ts';
import { resolverTienda } from '#lib/server/tienda/publica.ts';
import { codigoPedido } from '#lib/server/tienda/pedido.ts';
import { irAfuera } from '#lib/server/afuera.ts';
import type { Actions, PageServerLoad } from './$types';

/** La confirmación del pedido: el código lleva la firma, así nadie mira el pedido de otro. */
export const load: PageServerLoad = async ({ params, cookies, url, depends }) => {
	// Tiempo real (§44): al acreditarse el pago (o cambiar el estado) la página se actualiza sola.
	depends('tienda:pedido-publico');
	const r = await resolverTienda(params.tienda, cookies.get(COOKIE));
	// Con la API real, el código es el público del pedido.
	if (r && !r.datos) {
		const res = await fetch(
			`${String(env.API_URL)}/v1/publico/tiendas/${r.tienda.slug}/pedidos/${encodeURIComponent(params.codigo)}`
		);
		if (!res.ok) error(404, 'No encontramos ese pedido.');
		const c = (await res.json()) as {
			numero: number;
			total: number;
			pago: string;
			nombre: string;
			estado: string;
			estadoPago?: string;
			envio: string;
			seguimiento: string | null;
			alias: string | null;
			cbu: string | null;
			titular: string | null;
			/** Reservado o encargado con seña (§52). */
			sena?: {
				tipo: 'reserva' | 'encargo';
				monto: number;
				falta: number;
				pagada: boolean;
				saldado: boolean;
			} | null;
		};
		// Vuelve de Mercado Pago con el pago aprobado y el pedido todavía sin pagar: se le pide a la API
		// que lo confirme (la API lo consulta en Mercado Pago; no confía en lo que dice la URL).
		const idPago = url.searchParams.get('payment_id') ?? url.searchParams.get('collection_id');
		const aprobado =
			(url.searchParams.get('status') ?? url.searchParams.get('collection_status')) === 'approved';
		if (
			c.pago === 'mercadopago' &&
			c.estado === 'Sin pagar' &&
			aprobado &&
			idPago &&
			/^\d{1,20}$/.test(idPago)
		)
			await fetch(`${String(env.API_URL)}/v1/publico/tiendas/${r.tienda.slug}/pagos/${idPago}`, {
				method: 'POST',
				signal: AbortSignal.timeout(5000)
			}).catch(() => null);
		// Los datos para transferir son los que cargó la tienda (nunca inventados). Sin alias, el CBU/CVU.
		const tr = c.pago === 'transferencia';
		return {
			pedido: {
				...c,
				codigo: params.codigo,
				alias: tr ? (c.alias ?? c.cbu) : null,
				esCbu: tr && !c.alias && !!c.cbu,
				// Los mismos estados de «pagado» que los mails a quien compra (Go, 00034).
				pagado: ['Aprobado', 'Acreditada', 'Cobrado'].includes(c.estadoPago ?? ''),
				// Mercado Pago sin pagar todavía: la página ofrece pagarlo (si no, quedaba sin forma de pagar).
				porPagar: c.pago === 'mercadopago' && c.estado === 'Sin pagar' && c.estadoPago === 'Pendiente' && !c.sena,
				rechazado: ['rejected', 'failure', 'null'].includes(
					url.searchParams.get('status') ?? url.searchParams.get('collection_status') ?? ''
				)
			}
		};
	}
	const numero = Number(params.codigo.split('-')[0]);
	if (
		!r?.datos ||
		!r.sandboxId ||
		!Number.isInteger(numero) ||
		(await codigoPedido(r.sandboxId, numero)) !== params.codigo
	)
		error(404, 'No encontramos ese pedido.');
	const p = await r.datos.pedido(numero);
	if (!p) error(404, 'No encontramos ese pedido.');
	// Con seña (§52), se paga por Mercado Pago (en la demo, simulado).
	const pago =
		p.sena || url.searchParams.get('pago') === 'mercadopago' ? 'mercadopago' : 'transferencia';
	const total = Number(url.searchParams.get('total')) || p.total;
	const dt = pago === 'transferencia' ? r.tienda.cobros.datosTransferencia : null;
	return {
		pedido: {
			numero,
			codigo: params.codigo,
			total,
			pago,
			nombre: p.cliente.split(' ')[0],
			alias: dt?.alias || dt?.cbu || null,
			esCbu: !!dt && !dt.alias && !!dt.cbu,
			estado: p.estado,
			titular: r.tienda.cobros.datosTransferencia?.titular ?? null,
			pagado: false,
			sena: p.sena ?? null
		},
		// La demo no cobra la seña: la confirmación lo dice.
		senaDemo: !!p.sena
	};
};

export const actions: Actions = {
	/** Pagar después: arma un pago nuevo de Mercado Pago para el pedido y lleva a pagarlo. */
	pagar: async ({ params, cookies, request, getClientAddress }) => {
		const r = await resolverTienda(params.tienda, cookies.get(COOKIE));
		if (!r || r.datos) return fail(400, { error: 'Esta tienda no cobra con Mercado Pago.' });
		const res = await fetch(
			`${String(env.API_URL)}/v1/publico/tiendas/${r.tienda.slug}/pedidos/${encodeURIComponent(params.codigo)}/pagar`,
			{
				method: 'POST',
				headers: {
					'x-mostrador-bff': String(env.CLAVE_BFF ?? ''),
					'x-cliente-ip': request.headers.get('cf-connecting-ip') ?? getClientAddress()
				}
			}
		).catch(() => null);
		const g = res?.ok ? ((await res.json()) as { url?: string }) : null;
		if (!g?.url) return fail(409, { error: 'Mercado Pago no respondió. Probá de nuevo en un rato.' });
		irAfuera(g.url);
	}
};
