/**
 * Pedidos de la demo: la lista con filtros (los mismos de la URL del panel), los conteos por etapa y
 * por faceta (cada faceta se cuenta sin su propio filtro, como en el handoff), el detalle y los
 * cambios de estado. Los pedidos salen de demo.ts; lo que el visitante cambió (estado, factura)
 * llega en `cambios`.
 */
import type { components } from '@mostrador/contrato/esquema';
import type { Demo, EstadoPedido, Pedido } from './demo.ts';

type ListaPedidos = components['schemas']['ListaPedidos'];
type PedidoFila = components['schemas']['PedidoFila'];
type PedidoDetalle = components['schemas']['PedidoDetalle'];

const MIN = 60_000;
const DIA = 1440 * MIN;

/** El circuito del handoff (PanelPedidos `ST`), sin Cancelado. */
export const CIRCUITO: EstadoPedido[] = [
	'Sin pagar',
	'Para preparar',
	'Listo para despachar',
	'En camino',
	'Entregado'
];

/** Lo que el visitante cambió de un pedido. `en`: cuándo llegó a cada estado. */
export type CambioPedido = {
	estado?: EstadoPedido;
	en?: Partial<Record<EstadoPedido, number>>;
	facturado?: boolean;
	recordado?: number;
};

export type FiltrosPedidos = {
	q?: string;
	estado?: string[];
	canal?: string[];
	pago?: string[];
	envio?: string[];
	fecha?: 'hoy' | '7d' | '30d' | 'todo';
	minimo?: number;
	vip?: boolean;
	hornero?: boolean;
	mayorista?: boolean;
	orden?: 'nuevos' | 'urgentes' | 'monto' | 'viejos';
	desde?: number;
	cuantos?: number;
};

/**
 * Se puede avanzar una etapa, volver una (deshacer) o cancelar (y deshacer la cancelación). El
 * retiro en el local no pasa por "En camino" (como pedidos.TransicionValida de la API).
 */
export function transicionValida(de: EstadoPedido, a: EstadoPedido, envio = '') {
	if (de === a) return false;
	if (a === 'Cancelado') return de !== 'Entregado';
	if (de === 'Cancelado') return a !== 'Entregado';
	const c = envio === 'retiro' ? CIRCUITO.filter((e) => e !== 'En camino') : CIRCUITO;
	if (!c.includes(a)) return false;
	return Math.abs(c.indexOf(a) - c.indexOf(de)) === 1;
}

/** El pedido con lo que cambió el visitante. */
type PedidoVivo = Pedido & { cambio: CambioPedido };

const sinAcento = (s: string) => s.normalize('NFD').replace(/[̀-ͯ]/g, '').toLowerCase();

export function pedidosVivos(d: Demo, cambios: Map<number, CambioPedido>, ahora: number) {
	return d.pedidos
		.filter((p) => p.creado <= ahora)
		.map((p): PedidoVivo => {
			const c = cambios.get(p.numero) ?? {};
			return {
				...p,
				estado: c.estado ?? p.estado,
				facturado: c.facturado ?? p.facturado,
				cambio: c
			};
		});
}

/** Clientes VIP: el 10 % que más compró (sin mayoristas). */
function vipsDe(d: Demo) {
	const gastado = new Map<string, number>();
	for (const p of d.pedidos)
		if (p.estado !== 'Cancelado')
			gastado.set(p.clienteId, (gastado.get(p.clienteId) ?? 0) + p.total);
	const minoristas = [...gastado].filter(([id]) => !d.clientes.find((c) => c.id === id)?.mayorista);
	minoristas.sort((a, b) => b[1] - a[1]);
	return new Set(
		minoristas.slice(0, Math.max(1, Math.round(minoristas.length / 10))).map(([id]) => id)
	);
}

/** Lo vendió Tincho: 6 de cada 10 pedidos de WhatsApp e Instagram. */
const deHornero = (p: Pedido) => (p.canal === 'wa' || p.canal === 'ig') && p.numero % 10 < 6;

/** Urgencia del handoff: para despachar (3 hoy, 2 atrasado), sin pagar hace más de un día (1). */
function urgencia(p: PedidoVivo, ahora: number) {
	const horas = (ahora - p.creado) / 3_600_000;
	if (p.estado === 'Para preparar' || p.estado === 'Listo para despachar')
		return horas < 20 ? 3 : 2;
	return p.estado === 'Sin pagar' && horas > 24 ? 1 : 0;
}

export function listarDemo(
	d: Demo,
	cambios: Map<number, CambioPedido>,
	f: FiltrosPedidos,
	ahora: number
): ListaPedidos {
	const clientes = new Map(d.clientes.map((c) => [c.id, c]));
	const productos = new Map(d.catalogo.productos.map((p) => [p.id, p]));
	const vips = vipsDe(d);
	const q = sinAcento(f.q?.trim() ?? '');
	const limite =
		f.fecha === 'hoy'
			? ahora - DIA
			: f.fecha === '7d'
				? ahora - 7 * DIA
				: f.fecha === '30d'
					? ahora - 30 * DIA
					: 0;
	const en = (lista: string[] | undefined, v: string) => !lista?.length || lista.includes(v);

	// Los filtros que no son facetas.
	const base = pedidosVivos(d, cambios, ahora).filter((p) => {
		const c = clientes.get(p.clienteId)!;
		if (p.creado < limite || p.total < (f.minimo ?? 0)) return false;
		if (f.vip && !vips.has(p.clienteId)) return false;
		if (f.hornero && !deHornero(p)) return false;
		// Mayorista: el cliente lo es o el pedido salió con precios mayoristas (§49).
		if (f.mayorista && !c.mayorista && !p.mayorista) return false;
		if (!q) return true;
		const texto = `#${p.numero} ${p.numero} ${c.nombre} ${p.ciudadEnvio} ${p.items.map((i) => i.nombre).join(' ')}`;
		return sinAcento(texto).includes(q);
	});
	// Cada faceta se cuenta con los demás filtros, sin el suyo.
	const pasa = (p: PedidoVivo, sin?: 'estado' | 'canal' | 'pago' | 'envio') =>
		(sin === 'estado' || en(f.estado, p.estado)) &&
		(sin === 'canal' || en(f.canal, p.canal)) &&
		(sin === 'pago' || en(f.pago, p.pago)) &&
		(sin === 'envio' || en(f.envio, p.envio));
	const contar = (sin: 'estado' | 'canal' | 'pago' | 'envio', clave: (p: PedidoVivo) => string) => {
		const m: Record<string, number> = {};
		for (const p of base) if (pasa(p, sin)) m[clave(p)] = (m[clave(p)] ?? 0) + 1;
		return m;
	};

	let lista = base.filter((p) => pasa(p));
	if (f.orden === 'monto') lista = lista.toSorted((a, b) => b.total - a.total);
	else if (f.orden === 'viejos') lista = lista.toReversed();
	else if (f.orden === 'urgentes')
		lista = lista.toSorted(
			(a, b) => urgencia(b, ahora) - urgencia(a, ahora) || b.creado - a.creado
		);

	const desde = Math.max(0, f.desde ?? 0);
	const cuantos = Math.min(200, Math.max(1, f.cuantos ?? 60));
	return {
		total: lista.length,
		suma: lista.reduce((s, p) => s + p.total, 0),
		filas: lista.slice(desde, desde + cuantos).map((p) => fila(p, clientes, productos, vips)),
		etapas: contar('estado', (p) => p.estado),
		facetas: {
			canal: contar('canal', (p) => p.canal),
			pago: contar('pago', (p) => p.pago),
			envio: contar('envio', (p) => p.envio)
		}
	};
}

function fila(
	p: PedidoVivo,
	clientes: Map<string, Demo['clientes'][number]>,
	productos: Map<string, Demo['catalogo']['productos'][number]>,
	vips: Set<string>
): PedidoFila {
	const c = clientes.get(p.clienteId)!;
	return {
		numero: p.numero,
		cliente: c.nombre,
		foto: c.foto,
		ciudad: p.ciudadEnvio,
		canal: p.canal,
		pago: p.pago,
		envio: p.envio,
		estado: p.estado,
		total: p.total,
		creado: new Date(p.creado).toISOString(),
		productos: p.items
			.slice(0, 3)
			.map((i) => ({ nombre: i.nombre, foto: productos.get(i.productoId)?.foto ?? null })),
		cantidadProductos: p.items.length,
		vip: vips.has(p.clienteId),
		hornero: deHornero(p),
		mayorista: c.mayorista || !!p.mayorista,
		telefono: c.telefono,
		// El código firmado de la confirmación lo pone sandbox.ts (necesita el id del sandbox).
		codigo: null
	};
}

const NOTAS = [
	'pidió que vaya envuelto para regalo. Ya lo anoté en la etiqueta y le avisé que sale mañana.',
	'preguntó si llega antes del viernes. Le confirmé la fecha de entrega.',
	'quiere factura A. Le pedí el CUIT y lo cargué en el pedido.'
];
const TRANSPORTE: Record<string, string> = {
	andreani: 'Andreani',
	oca: 'OCA',
	correo: 'Correo Argentino',
	moto: 'Moto propia',
	retiro: 'Retiro en el local'
};
const CANAL: Record<string, string> = {
	web: 'la tienda online',
	wa: 'WhatsApp',
	ml: 'Mercado Libre',
	ig: 'Instagram',
	pos: 'el local',
	tg: 'Telegram'
};
const MEDIO: Record<string, string> = {
	mp: 'Mercado Pago',
	tarjeta: 'tarjeta',
	payway: 'Payway',
	modo: 'MODO',
	uala: 'Ualá',
	transf: 'transferencia',
	efectivo: 'efectivo'
};

export function detalleDemo(
	d: Demo,
	cambios: Map<number, CambioPedido>,
	numero: number,
	ahora: number
): PedidoDetalle | null {
	const p = pedidosVivos(d, cambios, ahora).find((x) => x.numero === numero);
	if (!p) return null;
	const clientes = new Map(d.clientes.map((c) => [c.id, c]));
	const productos = new Map(d.catalogo.productos.map((x) => [x.id, x]));
	const c = clientes.get(p.clienteId)!;
	const suyos = d.pedidos.filter(
		(x) => x.clienteId === p.clienteId && x.estado !== 'Cancelado' && x.creado <= ahora
	);
	const paso = CIRCUITO.indexOf(p.estado);
	const llegada = (e: EstadoPedido, minutos: number) =>
		p.cambio.en?.[e] ?? (paso >= CIRCUITO.indexOf(e) ? p.creado + minutos * MIN : null);
	const iso = (t: number | null) =>
		t === null ? null : new Date(Math.min(t, ahora)).toISOString();
	const conEnvio = p.envio !== 'retiro' && p.envio !== 'moto';
	const historial: PedidoDetalle['historial'] = [
		{ texto: `Pedido creado por ${CANAL[p.canal] ?? p.canal}`, cuando: iso(p.creado), hecho: true }
	];
	if (p.estado === 'Cancelado')
		historial.push({
			texto: 'Pedido cancelado',
			cuando: iso(p.cambio.en?.Cancelado ?? null),
			hecho: true
		});
	else {
		historial.push({
			texto:
				paso >= 1
					? `Pago ${p.estadoPago.toLowerCase()} · ${MEDIO[p.pago] ?? p.pago}`
					: 'Esperando el pago',
			cuando: iso(llegada('Para preparar', 2)),
			hecho: paso >= 1
		});
		if (p.facturado)
			historial.push({
				texto: 'Factura B emitida',
				cuando: iso(llegada('Para preparar', 2)),
				hecho: true
			});
		historial.push({
			texto: paso >= 2 ? 'Preparado' : 'Preparar y despachar',
			cuando: iso(llegada('Listo para despachar', 180)),
			hecho: paso >= 2
		});
		if (p.envio !== 'retiro')
			historial.push({
				texto: `Despachado con ${TRANSPORTE[p.envio]}`,
				cuando: iso(llegada('En camino', 1440)),
				hecho: paso >= 3
			});
		historial.push({
			texto: p.envio === 'retiro' ? 'Retirado' : 'Entregado',
			cuando: iso(llegada('Entregado', 3 * 1440)),
			hecho: paso >= 4
		});
	}
	const subtotal = p.items.reduce((s, i) => s + i.precio * i.cantidad, 0);
	return {
		...fila(p, clientes, productos, vipsDe(d)),
		items: p.items.map((i) => ({
			productoId: i.productoId,
			nombre: i.nombre,
			foto:
				d.variantes?.get(i.productoId)?.variantes.find((v) => v.id === i.varianteId)?.foto ??
				productos.get(i.productoId)?.foto ??
				null,
			cantidad: i.cantidad,
			precio: i.precio,
			varianteId: i.varianteId ?? null,
			variante: i.variante ?? null
		})),
		subtotal,
		costoEnvio: p.total - subtotal,
		descuentos: [],
		comprador: {
			nombre: c.nombre,
			foto: c.foto,
			telefono: c.telefono,
			email: c.email,
			desde: new Date(c.alta).toISOString(),
			pedidos: suyos.length,
			gastado: suyos.reduce((s, x) => s + x.total, 0)
		},
		estadoPago: p.estado === 'Sin pagar' ? 'Pendiente' : p.estadoPago,
		sena: p.sena
			? {
					...p.sena,
					falta: p.sena.saldado ? 0 : Math.max(0, p.total - p.sena.monto),
					vence: null
				}
			: null,
		factura: p.facturado ? `B 0004-${String(p.numero - 200).padStart(8, '0')}` : null,
		seguimiento:
			conEnvio && paso >= 3
				? `36000${String(p.numero * 7919)
						.padStart(6, '0')
						.slice(-6)}`
				: null,
		llega:
			conEnvio && p.estado === 'En camino'
				? new Date(p.creado + 3 * DIA - 3 * 3_600_000).toISOString().slice(0, 10)
				: null,
		nota: deHornero(p) && p.numero % 3 === 0 ? NOTAS[p.numero % NOTAS.length] : null,
		entrega: entregaDe(p),
		historial
	};
}

const CALLES = [
	'Av. Corrientes',
	'Gorriti',
	'Belgrano',
	'San Martín',
	'Mitre',
	'Rivadavia',
	'Sarmiento',
	'Urquiza'
];

/** Adónde va: la que cargó quien compró o, en los de ejemplo, una verosímil (siempre la misma). */
function entregaDe(p: {
	numero: number;
	envio: string;
	entrega?: { direccion: string; cp: string } | null;
}) {
	if (p.envio === 'retiro') return null;
	if (p.entrega !== undefined) return p.entrega;
	return {
		direccion: `${CALLES[p.numero % CALLES.length]} ${100 + ((p.numero * 37) % 4800)}`,
		cp: String(1000 + ((p.numero * 131) % 8000))
	};
}
