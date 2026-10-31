import { error, json, redirect } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { EstadoPedido } from '#lib/server/datos/tipos.ts';
import type { RequestHandler } from './$types';

const ESTADOS: EstadoPedido[] = [
	'Sin pagar',
	'Para preparar',
	'Listo para despachar',
	'En camino',
	'Entregado',
	'Cancelado'
];

const numeroDe = (s: string) => {
	const n = Number(s);
	if (!Number.isSafeInteger(n) || n < 1) error(404, 'Ese pedido no existe.');
	return n;
};

/**
 * El detalle de un pedido (la hoja). Abierto en el navegador (el «Ver el pedido» de los mails y el
 * push), lleva a la lista con esa hoja abierta.
 */
export const GET: RequestHandler = async ({ locals, params, request }) => {
	const n = numeroDe(params.numero);
	if (request.headers.get('accept')?.includes('text/html'))
		redirect(303, `/panel/pedidos?pedido=${n}`);
	const p = await datosDe(locals).pedido(n);
	if (!p) error(404, 'Ese pedido no existe.');
	return json(p, { headers: { 'cache-control': 'no-store' } });
};

/**
 * {estado}: avanzar, deshacer o cancelar. {recordatorio: true}: reclamar el pago por WhatsApp.
 * {saldado}: con la seña pagada (§52), la tienda cobró el resto (false lo deshace).
 */
export const POST: RequestHandler = async ({ locals, params, request }) => {
	const n = numeroDe(params.numero);
	const cuerpo = (await request.json().catch(() => null)) as {
		estado?: unknown;
		recordatorio?: unknown;
		saldado?: unknown;
	} | null;
	const datos = datosDe(locals);
	let r;
	if (cuerpo?.recordatorio === true) r = await datos.recordarPago(n);
	else if (typeof cuerpo?.saldado === 'boolean') r = await datos.saldarPedido(n, cuerpo.saldado);
	else if (ESTADOS.includes(cuerpo?.estado as EstadoPedido))
		r = await datos.cambiarEstado(n, cuerpo!.estado as EstadoPedido);
	else error(400, 'Falta qué hacer con el pedido.');
	if (r === 'no-existe') error(404, 'Ese pedido no existe.');
	if (r === 'no-se-puede') error(409, 'El pedido ya cambió de etapa. Actualizá la lista.');
	return new Response(null, { status: 204 });
};
