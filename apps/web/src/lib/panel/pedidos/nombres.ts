/** Nombres, colores e íconos de Pedidos (handoff PanelPedidos: `CH`, `PAY`, `SHIP`, `ST`). */
import * as I from '#lib/icons.ts';
import type { IconData } from '#lib/icons.ts';
import type { EstadoPedido } from './filtros.ts';

export const NOMBRE_CANAL: Record<string, string> = {
	web: 'Tienda',
	wa: 'WhatsApp',
	ml: 'Mercado Libre',
	ig: 'Instagram',
	pos: 'Local',
	tg: 'Telegram'
};
export const NOMBRE_PAGO: Record<string, string> = {
	mp: 'Mercado Pago',
	tarjeta: 'Tarjeta',
	payway: 'Payway',
	modo: 'MODO',
	uala: 'Ualá',
	transf: 'Transferencia',
	efectivo: 'Efectivo'
};
export const NOMBRE_ENVIO: Record<string, string> = {
	andreani: 'Andreani',
	oca: 'OCA',
	correo: 'Correo Argentino',
	moto: 'Moto',
	retiro: 'Retiro',
	mercadoenvios: 'Mercado Envíos'
};
/** Ícono y color de cada medio de pago en la lista (handoff `PAY`). */
export const ICONO_PAGO: Record<string, [IconData, string]> = {
	mp: [I.siMercadopago, '#00B1EA'],
	tarjeta: [I.phCreditCardFill, '#3E6BFD'],
	payway: [I.phCreditCardFill, '#0E1E4A'],
	modo: [I.phDeviceMobileFill, '#00A88F'],
	uala: [I.phDeviceMobileFill, '#3E6BFD'],
	transf: [I.phBankFill, '#6B5244'],
	efectivo: [I.phMoneyFill, '#2F8F63']
};
export const ICONO_ENVIO: Record<string, IconData> = {
	andreani: I.phTruckFill,
	oca: I.phTruckFill,
	correo: I.phTruckFill,
	moto: I.phMopedFill,
	retiro: I.phStorefrontFill,
	mercadoenvios: I.phTruckFill
};

/** Las etapas del handoff (`ST`) con su color e ícono; Cancelado aparte. */
export const ETAPAS: { estado: EstadoPedido; t: string; c: string; i: IconData }[] = [
	{ estado: 'Sin pagar', t: 'Sin pagar', c: '#E3A93B', i: I.phHourglassMediumFill },
	{ estado: 'Para preparar', t: 'Para preparar', c: '#D2582F', i: I.phPackageFill },
	{ estado: 'Listo para despachar', t: 'Listo para despachar', c: '#C98A12', i: I.phTagFill },
	{ estado: 'En camino', t: 'En camino', c: '#3E6BFD', i: I.phTruckFill },
	{ estado: 'Entregado', t: 'Entregado', c: '#2F8F63', i: I.phCheckCircleFill }
];
export const CANCELADO = {
	estado: 'Cancelado' as const,
	t: 'Cancelado',
	c: '#B83A2E',
	i: I.phXCircleFill
};
export const etapaDe = (e: EstadoPedido) => ETAPAS.find((x) => x.estado === e) ?? CANCELADO;

/**
 * Una compra de la tienda online con Mercado Pago que todavía no se acreditó: no es una venta
 * (docs/decisiones.md §63). Se ve «Esperando pago», no se marca pagada a mano (la confirma Mercado
 * Pago) y a las 24 h se cancela sola.
 */
export const esperaPago = (p: { estado: EstadoPedido; pago?: string; canal?: string }) =>
	p.estado === 'Sin pagar' && p.pago === 'mp' && p.canal === 'web';
export const ESPERANDO = {
	estado: 'Sin pagar' as const,
	t: 'Esperando pago',
	c: '#7E6656',
	i: I.phHourglassMediumFill
};
/** La etapa con que se dibuja el pedido (la de «Esperando pago» si corresponde). */
export const etapaDePedido = (p: { estado: EstadoPedido; pago?: string; canal?: string }) =>
	esperaPago(p) ? ESPERANDO : etapaDe(p.estado);

/** El botón principal de la hoja: lleva a la etapa siguiente. */
export const SIGUIENTE: Partial<Record<EstadoPedido, string>> = {
	'Sin pagar': 'Marcar como pagado',
	'Para preparar': 'Marcar como preparado',
	'Listo para despachar': 'Marcar como despachado',
	'En camino': 'Marcar como entregado'
};
export const PROXIMO: Partial<Record<EstadoPedido, EstadoPedido>> = {
	'Sin pagar': 'Para preparar',
	'Para preparar': 'Listo para despachar',
	'Listo para despachar': 'En camino',
	'En camino': 'Entregado'
};

/**
 * El retiro en el local no se despacha: Listo para retirar → Retirado, sin "En camino" (la API
 * tampoco lo deja pasar por ahí ni manda el mail de despacho).
 */
const TEXTO_RETIRO: Partial<Record<EstadoPedido, string>> = {
	'Listo para despachar': 'Listo para retirar',
	Entregado: 'Retirado'
};
const SIGUIENTE_RETIRO: Partial<Record<EstadoPedido, string>> = {
	'Sin pagar': 'Marcar como pagado',
	'Para preparar': 'Marcar como listo para retirar',
	'Listo para despachar': 'Marcar como retirado'
};
const PROXIMO_RETIRO: Partial<Record<EstadoPedido, EstadoPedido>> = {
	'Sin pagar': 'Para preparar',
	'Para preparar': 'Listo para despachar',
	'Listo para despachar': 'Entregado'
};
type ConEnvio = { estado: EstadoPedido; envio: string; pago?: string; canal?: string };
export const textoEstado = (p: ConEnvio) =>
	(p.envio === 'retiro' && TEXTO_RETIRO[p.estado]) || etapaDePedido(p).t;
export const siguienteDe = (p: ConEnvio) =>
	esperaPago(p)
		? undefined
		: p.envio === 'retiro'
			? SIGUIENTE_RETIRO[p.estado]
			: SIGUIENTE[p.estado];
export const proximoDe = (p: ConEnvio) =>
	p.envio === 'retiro' ? PROXIMO_RETIRO[p.estado] : PROXIMO[p.estado];
