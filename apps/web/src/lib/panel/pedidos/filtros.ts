/**
 * Filtros de Pedidos en la URL (`/panel/pedidos?estado=para-preparar&canal=wa,ml&fecha=hoy`): se
 * comparten, sobreviven a recargar y los usa también el Resumen para mandar a una vista.
 */
import type { components } from '@mostrador/contrato/esquema';

export type EstadoPedido = components['schemas']['EstadoPedido'];

export const ESTADO_SLUG: Record<EstadoPedido, string> = {
	'Sin pagar': 'sin-pagar',
	'Para preparar': 'para-preparar',
	'Listo para despachar': 'listo',
	'En camino': 'en-camino',
	Entregado: 'entregado',
	Cancelado: 'cancelado'
};
const SLUG_ESTADO = Object.fromEntries(
	Object.entries(ESTADO_SLUG).map(([e, s]) => [s, e])
) as Record<string, EstadoPedido>;

export type Fecha = 'hoy' | '7d' | '30d' | 'todo';
export type Orden = 'nuevos' | 'urgentes' | 'monto' | 'viejos';

export type Filtros = {
	q: string;
	estado: EstadoPedido[];
	canal: string[];
	pago: string[];
	envio: string[];
	fecha: Fecha;
	minimo: number;
	vip: boolean;
	hornero: boolean;
	mayorista: boolean;
	orden: Orden;
};

const lista = (v: string | null) => (v ? v.split(',').filter(Boolean) : []);

export function leerFiltros(p: Pick<URLSearchParams, 'get'>): Filtros {
	const fecha = p.get('fecha');
	const orden = p.get('orden');
	return {
		q: (p.get('q') ?? '').slice(0, 100),
		estado: lista(p.get('estado'))
			.map((s) => SLUG_ESTADO[s])
			.filter(Boolean),
		canal: lista(p.get('canal')),
		pago: lista(p.get('pago')),
		envio: lista(p.get('envio')),
		fecha: fecha === 'hoy' || fecha === '7d' || fecha === '30d' ? fecha : 'todo',
		minimo: Math.max(0, Number(p.get('minimo')) || 0),
		vip: p.get('vip') === '1',
		hornero: p.get('hornero') === '1',
		mayorista: p.get('mayorista') === '1',
		orden: orden === 'urgentes' || orden === 'monto' || orden === 'viejos' ? orden : 'nuevos'
	};
}

/** Los filtros como query de la URL (sin lo que está en su valor por defecto). */
export function escribirFiltros(f: Partial<Filtros>) {
	const p = new URLSearchParams();
	if (f.q?.trim()) p.set('q', f.q.trim());
	if (f.estado?.length) p.set('estado', f.estado.map((e) => ESTADO_SLUG[e]).join(','));
	for (const k of ['canal', 'pago', 'envio'] as const) if (f[k]?.length) p.set(k, f[k]!.join(','));
	if (f.fecha && f.fecha !== 'todo') p.set('fecha', f.fecha);
	if (f.minimo) p.set('minimo', String(f.minimo));
	for (const k of ['vip', 'hornero', 'mayorista'] as const) if (f[k]) p.set(k, '1');
	if (f.orden && f.orden !== 'nuevos') p.set('orden', f.orden);
	return p;
}

/** Vistas guardadas del handoff. */
export const VISTAS: { id: string; t: string; filtros: Partial<Filtros> }[] = [
	{ id: 'todos', t: 'Todos', filtros: {} },
	{
		id: 'hoy',
		t: 'Para despachar hoy',
		filtros: { estado: ['Para preparar', 'Listo para despachar'], fecha: 'hoy' }
	},
	{ id: 'sin-pagar', t: 'Sin pagar', filtros: { estado: ['Sin pagar'] } },
	// Compras de la tienda online con Mercado Pago sin acreditar (todavía no son ventas, §63).
	{
		id: 'esperando-pago',
		t: 'Esperando pago',
		filtros: { estado: ['Sin pagar'], pago: ['mp'], canal: ['web'] }
	},
	{ id: 'vip', t: 'Clientes VIP', filtros: { vip: true } },
	{ id: 'mayoristas', t: 'Mayoristas', filtros: { mayorista: true } },
	{ id: 'hornero', t: 'Vendidos por Tincho', filtros: { hornero: true } }
];

/** La vista que coincide exactamente con los filtros (sin contar el orden). */
export function vistaDe(f: Filtros) {
	const sinOrden = (x: Partial<Filtros>) => {
		const p = escribirFiltros({ ...x, orden: 'nuevos' });
		p.sort();
		return p.toString();
	};
	const actual = sinOrden(f);
	return VISTAS.find((v) => sinOrden(v.filtros) === actual)?.id ?? null;
}
