/** Productos en el motor de listas (handoff PanelListas `CFG.Productos`). */
import * as I from '#lib/icons.ts';
import { unsplash } from '#lib/landing/formato.ts';
import { pesos } from '../formato.ts';
import { rutaDe } from '../navegacion.ts';
import { CANALES } from './clientes.ts';
import type { ConfigLista, Ctx } from './tipos.ts';

export type ProductoFila = {
	id: string;
	nombre: string;
	foto: string | null;
	sku: string;
	categoria: string;
	precio: number;
	costo: number;
	/** null: sin costo cargado, el margen no se sabe. */
	margen: number | null;
	stock: number;
	ventas: number[];
	vendidos30: number;
	canales: string[];
	publicado: boolean;
	/** La parte del link público (/t/<tienda>/producto/<slug>). */
	slug?: string;
	estado: 'active' | 'out' | 'paused' | 'draft';
	lowm: boolean;
	star: boolean;
	/** Cuántos pidieron «Avisame cuando vuelva» y todavía esperan (§60). */
	esperan?: number;
};

/** Ícono de cada canal en la fila (handoff `CHI`). */
const ICONO_CANAL: Record<string, [typeof I.phGlobeSimpleFill, string, string]> = {
	web: [I.phGlobeSimpleFill, '#D2582F', 'Tienda'],
	wa: [I.logosWhatsappIcon, 'inherit', 'WhatsApp'],
	ml: [I.phHandshakeFill, '#2D3277', 'Mercado Libre'],
	ig: [I.logosInstagramIcon, 'inherit', 'Instagram'],
	pos: [I.phStorefrontFill, '#8A5A3E', 'Local']
};
const ESTADO = {
	out: ['Sin stock', 'rgba(184,58,46,.1)', '#B83A2E'],
	paused: ['Pausado', 'var(--chip)', 'var(--mu)'],
	draft: ['Borrador', 'rgba(62,107,253,.12)', '#3E6BFD']
} as const;

export function configProductos(
	nuevo: () => void,
	pedirPorcentaje: (ids: string[]) => void,
	compartir: (p: ProductoFila, ctx: Ctx) => Promise<void>
): ConfigLista<ProductoFila> {
	return {
		lista: 'productos',
		titulo: 'Productos',
		unidad: 'productos',
		buscar: 'Nombre, SKU o categoría',
		kpi: (l) =>
			`${l.total.toLocaleString('es-AR')} ${l.total === 1 ? 'producto' : 'productos'} · ${(l.resumen.unidades30 ?? 0).toLocaleString('es-AR')} unidades vendidas en 30 días`,
		primario: { t: 'Nuevo producto', icon: I.phCameraPlusFill, hacer: nuevo },
		tip: (l) => {
			const n = l.resumen.sinStockEnMl ?? 0;
			if (!n) return null;
			return {
				texto: `${n} ${n === 1 ? 'producto sin stock sigue publicado' : 'productos sin stock siguen publicados'} en Mercado Libre. ¿${n === 1 ? 'Lo pauso' : 'Los pauso'} hasta que repongas?`,
				boton: 'Pausar',
				async hacer(ctx) {
					await ctx.accion([], 'pausar-sin-stock-ml');
					ctx.ui.toast({
						k: 'ai',
						title: 'Productos pausados en Mercado Libre',
						desc: 'Los volvés a publicar con un toque cuando cargues stock'
					});
				}
			};
		},
		vistas: [
			{ id: 'todos', t: 'Todos', icon: I.phSquaresFourFill, color: 'var(--mu)', filtros: {} },
			{
				id: 'sin-stock',
				t: 'Sin stock',
				icon: I.phWarningCircleFill,
				color: '#B83A2E',
				filtros: { seg: ['out'] }
			},
			{
				id: 'estrellas',
				t: 'Estrellas',
				icon: I.phStarFill,
				color: '#C98A12',
				filtros: { banderas: ['star'] }
			},
			{
				id: 'margen',
				t: 'Margen bajo',
				icon: I.phTrendDownBold,
				color: '#D2582F',
				filtros: { banderas: ['lowm'] }
			}
		],
		segmentos: [
			{ k: 'active', t: 'Publicados', c: '#2F8F63' },
			{ k: 'out', t: 'Sin stock', c: '#B83A2E' },
			{ k: 'paused', t: 'Pausados', c: '#9C8B7E' },
			{ k: 'draft', t: 'Borradores', c: '#3E6BFD' }
		],
		facetas: [
			{ k: 'categoria', t: 'Categoría' },
			{ k: 'canales', t: 'Publicado en', opciones: CANALES }
		],
		rango: {
			t: 'Precio desde',
			max: (l) => Math.ceil((l.resumen.precioMax ?? 60_000) / 1000) * 1000,
			paso: 1000,
			formato: pesos
		},
		banderas: [
			{ k: 'lowm', t: 'Margen menor al 25%' },
			{ k: 'star', t: 'Se vende mucho (+60 en 30 días)' }
		],
		ordenes: [
			{ k: 'sold', t: 'Más vendidos' },
			{ k: 'stock', t: 'Menos stock' },
			{ k: 'price', t: 'Mayor precio' },
			{ k: 'margin', t: 'Mejor margen' }
		],
		fila: (p) => {
			const mx = Math.max(1, ...p.ventas);
			const e = p.estado in ESTADO ? ESTADO[p.estado as keyof typeof ESTADO] : null;
			return {
				img: p.foto ? unsplash(p.foto, 96, '&h=96') : null,
				ini: p.nombre.slice(0, 2).toUpperCase(),
				iniBg: 'var(--sf3)',
				iniFg: 'var(--mu)',
				radio: '12px',
				t: p.nombre,
				tags: p.canales
					.filter((k) => ICONO_CANAL[k])
					.map((k) => ({ icon: ICONO_CANAL[k][0], col: ICONO_CANAL[k][1], t: ICONO_CANAL[k][2] })),
				// Sin costo cargado, «margen sin dato» sólo ocupa lugar (y cortaba el stock a 375 px).
				sub: `${p.sku ? `${p.sku} · ` : ''}${p.stock} en stock${p.margen === null ? '' : ` · margen ${p.margen} %`}`,
				spark: p.ventas.map((x) => ({
					h: Math.max(3, (x / mx) * 22),
					c: x ? '#D2582F' : 'var(--sf3)'
				})),
				// Sin stock y con gente esperando, eso es lo que importa: se avisa sola al reponer.
				vL: p.esperan
					? `${p.esperan} ${p.esperan === 1 ? 'espera' : 'esperan'} que vuelva`
					: `${p.vendidos30} vendidos en 30 días`,
				vC: p.esperan ? '#B83A2E' : 'var(--mu)',
				vI: p.esperan ? I.phBellRingingFill : p.star ? I.phFireFill : undefined,
				chip: e ? { t: e[0], bg: e[1], fg: e[2] } : undefined,
				v: pesos(p.precio),
				vs: 'precio',
				toggle: { on: p.publicado, t: p.publicado ? 'Publicado' : 'Pausado' },
				act: { t: 'Compartir', icon: I.phShareNetworkBold, bg: 'var(--sf3)', fg: 'var(--tx)' }
			};
		},
		accionFila: compartir,
		async alternar(p, on, ctx) {
			await ctx.accion([p.id], 'publicar', on);
			ctx.ui.toast({
				k: 'ok',
				title: `${p.nombre} ${on ? 'publicado' : 'pausado'}`,
				desc: on ? 'Visible en todos tus canales' : 'Ya no se ve en la tienda ni en Mercado Libre',
				action: 'Deshacer',
				onAction: async () => {
					await ctx.accion([p.id], 'publicar', !on);
				}
			});
		},
		lote: [
			{ t: 'Cambiar precio', icon: I.phPercentBold, hacer: async (ids) => pedirPorcentaje(ids) },
			{
				t: 'Pausar',
				icon: I.phPauseFill,
				async hacer(ids, ctx) {
					const n = await ctx.accion(ids, 'pausar');
					ctx.ui.toast({
						k: 'ok',
						title: `${n} ${n === 1 ? 'producto pausado' : 'productos pausados'}`,
						desc: 'Ya no se ven en la tienda ni en Mercado Libre',
						action: 'Deshacer',
						onAction: async () => {
							await ctx.accion(ids, 'publicar', true);
						}
					});
				}
			},
			{
				t: 'Publicar en ML',
				icon: I.phMegaphoneFill,
				async hacer(_, ctx) {
					ctx.ui.toast({
						k: 'ml',
						title: 'Conectá Mercado Libre',
						desc: 'Para publicar desde acá, primero vinculá tu cuenta',
						action: 'Conectar',
						onAction: () => {
							location.href = rutaDe('Integraciones');
						}
					});
				}
			}
		]
	};
}
