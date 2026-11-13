/** Clientes en el motor de listas (handoff PanelListas `CFG.Clientes`). */
import * as I from '#lib/icons.ts';
import { cara } from '#lib/landing/formato.ts';
import { pesos, pesosCortos, plural } from '../formato.ts';
import { linkWhatsapp } from '#lib/whatsapp.ts';
import { rutaDe } from '../navegacion.ts';
import type { ConfigLista, Opcion } from './tipos.ts';

export type Cliente = {
	id: string;
	nombre: string;
	/** Cuando lo hay (es su id en la lista). */
	email?: string | null;
	foto: string | null;
	ini: number;
	ciudad: string;
	canal: string;
	telefono: string | null;
	pedidos: number;
	ultima: number;
	gastado: number;
	seg: 'vip' | 'freq' | 'new' | 'risk' | 'sleep';
	meses: number[];
	cuit: boolean;
	wa: boolean;
	cupon: boolean;
	etiquetas: string[];
};

/** Colores de las iniciales (handoff `INI`). */
const INI = [
	['#FCEDE5', '#C24E25'],
	['#E9F8EE', '#1F6F4A'],
	['#ECEBFB', '#4A47A3'],
	['#FFF4D9', '#8A5A08'],
	['#F3E9DF', '#6B4A33']
];
export const CANALES: Opcion[] = [
	{ k: 'web', t: 'Tienda', marca: 'web' },
	{ k: 'wa', t: 'WhatsApp', marca: 'wa' },
	{ k: 'ml', t: 'Mercado Libre', marca: 'ml' },
	{ k: 'ig', t: 'Instagram', marca: 'ig' },
	{ k: 'pos', t: 'Local', marca: 'pos' }
];
const SEG = {
	vip: ['VIP', 'rgba(201,138,18,.16)', '#8A5A08'],
	risk: ['En riesgo', 'rgba(210,88,47,.12)', '#C24E25'],
	sleep: ['Dormido', 'var(--chip)', 'var(--mu)'],
	new: ['Nuevo', 'rgba(62,107,253,.12)', '#3E6BFD']
} as const;

/** "hoy", "ayer", "hace 12 días", "hace 3 meses"… (handoff `ago`). */
export const hace = (d: number) =>
	d < 1
		? 'hoy'
		: d < 2
			? 'ayer'
			: d < 30
				? `hace ${Math.round(d)} días`
				: d < 365
					? `hace ${Math.round(d / 30)} ${Math.round(d / 30) === 1 ? 'mes' : 'meses'}`
					: 'hace más de un año';

export const iniciales = (n: string) =>
	n
		.split(' ')
		.map((w) => w[0])
		.slice(0, 2)
		.join('');

const nombreDe = (c: Cliente) => c.nombre.split(' ')[0];
const perdido = (c: Cliente) => c.seg === 'risk' || c.seg === 'sleep';

/** Abre WhatsApp con el cliente (hasta conectar WhatsApp Cloud, la conversación se abre en el teléfono). */
export function escribir(c: { telefono: string | null; nombre: string }) {
	const link = c.telefono ? linkWhatsapp(c.telefono) : null;
	if (!link) return false;
	window.open(link, '_blank', 'noopener');
	return true;
}

export function configClientes(
	abrirAlta: () => void,
	abrirEtiquetas: (ids: string[]) => void
): ConfigLista<Cliente> {
	return {
		lista: 'clientes',
		titulo: 'Clientes',
		unidad: 'clientes',
		buscar: 'Nombre, ciudad, teléfono o etiqueta',
		kpi: (l) =>
			`${plural(l.total, 'cliente', 'clientes')} · ${pesosCortos(l.resumen.gastado ?? 0)} en compras`,
		primario: { t: 'Nuevo cliente', icon: I.phUserPlusFill, hacer: abrirAlta },
		vacio: {
			hornero: 'saludo',
			titulo: 'Todavía no hay clientes',
			texto:
				'Cada persona que te compra queda acá con su historial. También podés cargar a los que ya tenés.',
			accion: { t: 'Cargar un cliente', hacer: abrirAlta }
		},
		tip: (l) => {
			const n = l.resumen.enRiesgo ?? 0;
			if (!n) return null;
			return {
				texto: `${n.toLocaleString('es-AR')} ${n === 1 ? 'cliente frecuente no compra' : 'clientes frecuentes no compran'} hace más de 75 días. ¿Les mando un cupón del 10% por WhatsApp?`,
				boton: 'Mandar cupón',
				async hacer(ctx) {
					const hechos = await ctx.accion([], 'cupon-riesgo');
					ctx.ui.toast({
						k: 'ai',
						title: 'Cupón enviado a clientes en riesgo',
						desc: `${hechos} ${hechos === 1 ? 'mensaje' : 'mensajes'} por WhatsApp · te aviso quién vuelve`
					});
				}
			};
		},
		vistas: [
			{ id: 'todos', t: 'Todos', icon: I.phUsersThreeFill, color: 'var(--mu)', filtros: {} },
			{
				id: 'riesgo',
				t: 'Para recuperar',
				icon: I.phHeartbeatFill,
				color: '#D2582F',
				filtros: { seg: ['risk'] }
			},
			{
				id: 'vip',
				t: 'VIP',
				icon: I.phCrownSimpleFill,
				color: '#C98A12',
				filtros: { seg: ['vip'] }
			},
			{
				id: 'nuevos',
				t: 'Nuevos este mes',
				icon: I.phSparkleFill,
				color: '#3E6BFD',
				filtros: { seg: ['new'] }
			}
		],
		segmentos: [
			{ k: 'vip', t: 'VIP', c: '#C98A12' },
			{ k: 'freq', t: 'Frecuentes', c: '#2F8F63' },
			{ k: 'new', t: 'Nuevos', c: '#3E6BFD' },
			{ k: 'risk', t: 'En riesgo', c: '#D2582F' },
			{ k: 'sleep', t: 'Dormidos', c: '#9C8B7E' }
		],
		facetas: [
			{ k: 'canal', t: 'Llegó por', opciones: CANALES },
			{ k: 'ciudad', t: 'Ciudad' }
		],
		rango: { t: 'Gastó en total, desde', max: () => 2_000_000, paso: 50_000, formato: pesosCortos },
		banderas: [
			{ k: 'wa', t: 'Acepta WhatsApp' },
			{ k: 'cuit', t: 'Con CUIT (factura A)' }
		],
		ordenes: [
			{ k: 'ltv', t: 'Más gastaron' },
			{ k: 'last', t: 'Compraron hace poco' },
			{ k: 'orders', t: 'Más pedidos' },
			{ k: 'az', t: 'A–Z' }
		],
		fila: (c) => {
			const seg = c.seg in SEG ? SEG[c.seg as keyof typeof SEG] : null;
			const [bg, fg] = INI[c.ini % INI.length];
			return {
				img: c.foto ? cara(c.foto) : null,
				ini: iniciales(c.nombre),
				iniBg: bg,
				iniFg: fg,
				radio: '50%',
				insignia: { marca: c.canal },
				t: c.nombre,
				tags: [
					...(c.cuit ? [{ icon: I.phIdentificationCardFill, col: 'var(--mu)', t: 'CUIT' }] : []),
					...c.etiquetas.map((e) => ({ icon: I.phTagFill, col: '#6A4FA0', t: e }))
				],
				// Sin ciudad no se dice «Sin ciudad»: lo que importa en la fila es cuántas veces compró.
				sub: `${c.ciudad ? `${c.ciudad} · ` : ''}${c.pedidos} ${c.pedidos === 1 ? 'pedido' : 'pedidos'}`,
				spark: c.meses.map((m) => ({ h: m ? 6 + m * 5 : 3, c: m ? '#D2582F' : 'var(--sf3)' })),
				vL: c.pedidos ? `Última compra ${hace(c.ultima)}` : 'Todavía no compró',
				vC: c.ultima > 75 ? '#C24E25' : 'var(--mu)',
				chip: seg ? { t: seg[0], bg: seg[1], fg: seg[2] } : undefined,
				v: pesosCortos(c.gastado),
				vs: c.pedidos ? `ticket ${pesos(c.gastado / c.pedidos)}` : 'sin compras',
				act: c.cupon
					? { t: 'Cupón enviado', icon: I.phCheckBold, bg: 'var(--chip)', fg: 'var(--mu)' }
					: perdido(c)
						? { t: 'Cupón', icon: I.phTicketFill, bg: 'rgba(210,88,47,.14)', fg: '#b0441d' }
						: { t: 'Escribir', icon: I.phWhatsappLogoFill, bg: '#E9F8EE', fg: '#1F6F4A' }
			};
		},
		async accionFila(c, ctx) {
			if (c.cupon) {
				ctx.ui.toast({
					k: 'info',
					title: `${nombreDe(c)} ya tiene su cupón`,
					desc: 'Vence a los 7 días de enviado'
				});
				return;
			}
			if (perdido(c)) {
				await ctx.accion([c.id], 'cupon');
				ctx.ui.toast({
					k: 'ok',
					title: `Cupón del 10% para ${nombreDe(c)}`,
					desc: 'Enviado por WhatsApp · vence en 7 días',
					action: 'Deshacer',
					onAction: async () => {
						await ctx.accion([c.id], 'sin-cupon');
					}
				});
				return;
			}
			if (!escribir(c))
				ctx.ui.toast({
					k: 'warn',
					title: `${nombreDe(c)} no tiene WhatsApp cargado`,
					desc: 'Sumalo desde su ficha'
				});
		},
		lote: [
			{
				t: 'Escribirles',
				icon: I.phWhatsappLogoFill,
				async hacer(ids, ctx) {
					ctx.ui.toast({
						k: 'wa',
						title: `Mensaje para ${ids.length} ${ids.length === 1 ? 'cliente' : 'clientes'}`,
						desc: 'Armalo como campaña: ves el costo exacto antes de mandarlo',
						action: 'Ir a Marketing',
						onAction: () => {
							location.href = rutaDe('Marketing');
						}
					});
				}
			},
			{
				t: 'Mandar cupón',
				icon: I.phTicketFill,
				async hacer(ids, ctx) {
					const n = await ctx.accion(ids, 'cupon');
					ctx.ui.toast({
						k: 'ok',
						title: `Cupón · ${n} ${n === 1 ? 'cliente' : 'clientes'}`,
						desc: 'Por WhatsApp · vence en 7 días',
						action: 'Deshacer',
						onAction: async () => {
							await ctx.accion(ids, 'sin-cupon');
						}
					});
				}
			},
			{ t: 'Etiquetar', icon: I.phTagFill, hacer: async (ids) => abrirEtiquetas(ids) }
		]
	};
}
