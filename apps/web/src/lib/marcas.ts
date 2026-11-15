/**
 * Identidad visual de cada canal, medio de pago, transportista y tipo de aviso: el cuadradito con
 * ícono (o sigla) que acompaña toasts, filas y el feed. Copia del mapa `L` del handoff
 * (design/Mostrador v4.dc.html), que es la fuente de verdad de estos colores.
 */
import * as I from '#lib/icons.ts';
import type { IconData } from '#lib/icons.ts';

export type Marca = {
	/** Ícono (Phosphor o logo). Si no hay, se muestra `sigla`. */
	icon?: IconData;
	sigla?: string;
	/** Color del ícono o de la sigla. */
	col?: string;
	/** Fondo del cuadradito (puede ser un degradé). */
	bg: string;
};

export const MARCAS = {
	wa: { icon: I.logosWhatsappIcon, bg: '#E9F8EE' },
	tg: { icon: I.logosTelegram, bg: '#E8F4FB' },
	ig: { icon: I.logosInstagramIcon, bg: '#FCEEF3' },
	mp: { icon: I.siMercadopago, col: '#00B1EA', bg: '#E6F7FD' },
	ml: { sigla: 'ML', col: '#2D3277', bg: '#FFE600' },
	arca: { sigla: 'ARCA', col: '#fff', bg: '#1F3A6E' },
	andreani: { sigla: 'AND', col: '#fff', bg: '#E2001A' },
	web: { icon: I.phGlobeSimpleFill, col: '#D2582F', bg: '#FCEDE5' },
	pos: { icon: I.phStorefrontFill, col: '#8A5A3E', bg: '#F3E9DF' },
	transf: { icon: I.phBankFill, col: '#6B5244', bg: '#F1E9DF' },
	efectivo: { icon: I.phMoneyFill, col: '#2F8F63', bg: '#E9F5EE' },
	payway: { sigla: 'PW', col: '#fff', bg: '#0E1E4A' },
	modo: { sigla: 'MODO', col: '#fff', bg: '#00A88F' },
	uala: { sigla: 'ualá', col: '#fff', bg: '#3E6BFD' },
	oca: { sigla: 'OCA', col: '#fff', bg: '#4B2A85' },
	correo: { sigla: 'CA', col: '#003DA5', bg: '#FFCE00' },
	moto: { icon: I.phMopedFill, col: '#9A6A10', bg: '#FDF3DD' },
	ga: { icon: I.logosGoogleAnalytics, bg: '#FFF4E5' },
	meta: { icon: I.logosMetaIcon, bg: '#EEF4FD' },
	zap: { icon: I.logosZapier, bg: '#FFF1EB' },
	gg: { icon: I.logosGoogleIcon, bg: '#EEF4FD' },
	caja: { icon: I.phWalletFill, col: '#2F8F63', bg: '#E9F5EE' },
	retiro: { icon: I.phStorefrontFill, col: '#8A5A3E', bg: '#F3E9DF' },
	visa: { icon: I.logosVisa, bg: '#fff' },
	master: { icon: I.logosMastercard, bg: '#fff' },
	sys: { icon: I.phGearSixFill, col: '#6B5244', bg: '#F1E9DF' },
	star: { icon: I.phStarFill, col: '#C98A12', bg: '#FDF3DD' },
	ok: { icon: I.phCheckCircleFill, col: '#2F8F63', bg: 'rgba(47,143,99,.14)' },
	warn: { icon: I.phWarningFill, col: '#9A6A10', bg: 'rgba(227,169,59,.2)' },
	err: { icon: I.phXCircleFill, col: '#B83A2E', bg: 'rgba(184,58,46,.12)' },
	info: { icon: I.phInfoFill, col: '#C24E25', bg: 'rgba(210,88,47,.12)' },
	ai: { icon: I.phSparkleFill, col: '#fff', bg: 'linear-gradient(140deg,#E8875F,#C24E25)' },
	link: { icon: I.phLinkBold, col: '#C24E25', bg: 'rgba(210,88,47,.12)' }
} satisfies Record<string, Marca>;

export type ClaveMarca = keyof typeof MARCAS;

export const marca = (k: string): Marca => MARCAS[k as ClaveMarca] ?? MARCAS.sys;

/** Color de la barra de vida de un toast según su tipo (handoff: `BAR`). */
export const BARRA: Partial<Record<ClaveMarca, string>> = {
	ok: '#2F8F63',
	warn: '#E3A93B',
	err: '#B83A2E',
	info: '#D2582F',
	ai: '#D2582F'
};

/** Estados de pedido: fondo y color del chip (handoff: `ST` y la fila "Botones y estados"). */
export const ESTADOS = {
	'Sin pagar': ['rgba(227,169,59,.18)', '#9A6A10'],
	'Pago pendiente': ['rgba(227,169,59,.18)', '#9A6A10'],
	// Mercado Pago de la tienda online sin acreditar: todavía no es una venta.
	'Esperando pago': ['rgba(126,102,86,.12)', '#7E6656'],
	'Para preparar': ['rgba(210,88,47,.13)', '#C24E25'],
	Pagado: ['rgba(47,143,99,.13)', '#2F8F63'],
	'Listo para despachar': ['rgba(201,138,18,.16)', '#9A6A10'],
	Enviado: ['rgba(227,169,59,.18)', '#9A6A10'],
	'En camino': ['rgba(227,169,59,.18)', '#9A6A10'],
	Entregado: ['rgba(126,102,86,.14)', '#7E6656'],
	Cancelado: ['rgba(184,58,46,.1)', '#B83A2E']
} as const satisfies Record<string, readonly [string, string]>;

export type Estado = keyof typeof ESTADOS;
