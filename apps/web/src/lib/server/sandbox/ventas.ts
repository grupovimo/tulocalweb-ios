/**
 * Bandeja y carritos de la demo (design/PanelVentas.dc.html). Las conversaciones salen de quienes
 * compraron hace poco (las tres primeras son las del Resumen) y la sugerencia de Tincho se arma con
 * los datos de la tienda (envío, stock, lista mayorista), sin llamar al modelo.
 */
import type { components } from '@mostrador/contrato/esquema';
import type { Demo } from './demo.ts';

type ConversacionBreve = components['schemas']['ConversacionBreve'];
type Conversacion = components['schemas']['Conversacion'];
type Carritos = components['schemas']['Carritos'];
type Mensaje = Conversacion['mensajes'][number];

const MIN = 60_000;
const DIA = 1440 * MIN;
const iso = (t: number) => new Date(t).toISOString();

/** Lo que el visitante hizo en una conversación. */
export type CambioConversacion = {
	mensajes?: { texto: string; en: number }[];
	tomada?: boolean;
	leida?: boolean;
};

type Base = {
	id: string;
	clienteId: string;
	canal: ConversacionBreve['canal'];
	atiende: ConversacionBreve['atiende'];
	sinLeer: number;
	mensajes: Mensaje[];
};

/** Las conversaciones de ejemplo con su hilo. */
function base(d: Demo): Base[] {
	const c = d.conversaciones;
	const cliente = (id: string) => d.clientes.find((x) => x.id === id)!;
	const primero = (id: string) => cliente(id).nombre.split(' ')[0];
	const prod = (i: number) => d.catalogo.productos[i % d.catalogo.productos.length];
	const pregunta = d.pedidos.find((p) => p.clienteId === c[0].clienteId)!.items[0];
	const producto = d.catalogo.productos.find((p) => p.id === pregunta.productoId) ?? prod(0);
	const stock = d.stock[producto.id] ?? 3;
	// Otros que compraron hace poco (distintos de los tres de arriba).
	const otros = [
		...new Set(
			d.pedidos
				.map((p) => p.clienteId)
				.filter((id) => !c.some((x) => x.clienteId === id) && !cliente(id).mayorista)
		)
	].slice(0, 4);
	const t0 = c[0].creado;
	const m = (
		de: Mensaje['de'],
		texto: string,
		haceMin: number,
		extra: Partial<Mensaje> = {}
	): Mensaje => ({
		de,
		texto,
		en: iso(t0 - haceMin * MIN),
		...extra
	});
	return [
		{
			id: 'c1',
			clienteId: c[0].clienteId,
			canal: 'wa',
			atiende: 'hornero',
			sinLeer: 0,
			mensajes: [
				m('cliente', c[0].mensaje, 3),
				m(
					'hornero',
					`¡Hola ${primero(c[0].clienteId)}! Sí, ${stock ? `nos quedan ${stock}` : 'lo estamos reponiendo'} 🙌`,
					3
				),
				m('hornero', '', 3, {
					producto: { nombre: producto.nombre, foto: producto.foto, precio: producto.precio, stock }
				}),
				m('cliente', `¿Hacen envíos a ${cliente(c[0].clienteId).ciudad}?`, 1)
			]
		},
		{
			id: 'c2',
			clienteId: c[1].clienteId,
			canal: 'wa',
			atiende: 'vos',
			sinLeer: 2,
			mensajes: [m('cliente', 'Hola, compré ayer', 9), m('cliente', c[1].mensaje, 8)]
		},
		{
			id: 'c3',
			clienteId: c[2].clienteId,
			canal: 'ig',
			atiende: 'vos',
			sinLeer: 1,
			mensajes: [m('cliente', c[2].mensaje, 18)]
		},
		...otros.map((id, k): Base => {
			const p = prod(k + 3);
			const plantillas: [Base['canal'], Base['atiende'], Mensaje[]][] = [
				[
					'ig',
					'resuelto',
					[
						m('cliente', '¿Cuándo llega mi pedido?', 70),
						m(
							'hornero',
							'Sale hoy por Andreani y te llega en 48 h. Te paso el seguimiento por acá.',
							69
						),
						m('cliente', '¡Gracias! Ya lo recibí 😊', 56)
					]
				],
				[
					'tg',
					'hornero',
					[
						m('cliente', `¿Tienen ${p.nombre.toLowerCase()}?`, 120),
						m('hornero', `¡Sí! Está a $ ${p.precio.toLocaleString('es-AR')}. ¿Te lo reservo?`, 119)
					]
				],
				[
					'wa',
					'hornero',
					[
						m('cliente', 'Ya te transferí', 180),
						m('hornero', '¡Gracias! En cuanto se acredite te confirmo el pedido.', 179)
					]
				],
				[
					'ml',
					'resuelto',
					[
						m('cliente', '¿Hacen factura A?', 300),
						m('hornero', 'Sí, hacemos factura A. Pasanos el CUIT al comprar y sale sola.', 299)
					]
				]
			];
			const [canal, atiende, mensajes] = plantillas[k];
			return { id: `c${k + 4}`, clienteId: id, canal, atiende, sinLeer: 0, mensajes };
		})
	];
}

function aplicar(b: Base, cambio: CambioConversacion | undefined): Base {
	if (!cambio) return b;
	const propios = (cambio.mensajes ?? []).map((x): Mensaje => ({
		de: 'vos',
		texto: x.texto,
		en: iso(x.en)
	}));
	return {
		...b,
		atiende: propios.length ? 'resuelto' : cambio.tomada ? 'vos' : b.atiende,
		sinLeer: cambio.leida ? 0 : b.sinLeer,
		mensajes: [...b.mensajes, ...propios]
	};
}

export function bandejaDemo(d: Demo, cambios: Map<string, CambioConversacion>) {
	const conversaciones: ConversacionBreve[] = base(d).map((b0) => {
		const b = aplicar(b0, cambios.get(b0.id));
		const c = d.clientes.find((x) => x.id === b.clienteId)!;
		const ultimo = b.mensajes.filter((x) => x.texto).at(-1)!;
		return {
			id: b.id,
			nombre: c.nombre,
			foto: c.foto,
			canal: b.canal,
			ultimo: ultimo.texto,
			creado: ultimo.en,
			atiende: b.atiende,
			sinLeer: b.sinLeer
		};
	});
	const hoy = d.pedidos.filter(
		(p) => p.creado > d.conversaciones[0].creado - DIA && p.canal !== 'pos'
	).length;
	return { conversaciones, atendidasHoy: hoy * 2 + d.conversaciones.length };
}

const pesos = (n: number) => '$ ' + Math.round(n).toLocaleString('es-AR');

export function conversacionDemo(
	d: Demo,
	id: string,
	cambios: Map<string, CambioConversacion>
): Conversacion | null {
	const b0 = base(d).find((x) => x.id === id);
	if (!b0) return null;
	const cambio = cambios.get(id);
	const b = aplicar(b0, cambio);
	const c = d.clientes.find((x) => x.id === b.clienteId)!;
	const suyos = d.pedidos.filter((p) => p.clienteId === c.id && p.estado !== 'Cancelado');
	const ultimo = suyos[0];
	const ultimoTexto = b.mensajes.filter((x) => x.texto).at(-1)!;
	// La sugerencia responde a lo último que preguntó el cliente.
	const pregunta =
		b.mensajes.at(-1)?.de === 'cliente' ? b.mensajes.at(-1)!.texto.toLowerCase() : '';
	// Lo más específico primero: "cambiar la dirección de envío" no es una pregunta de envío.
	const sugerencia = !pregunta
		? null
		: /direcci[oó]n/.test(pregunta)
			? `¡Hola ${c.nombre.split(' ')[0]}! Pasame la dirección nueva y la cambio antes de despachar.`
			: /mayorista/.test(pregunta)
				? 'Te paso la lista mayorista: desde 10 unidades tenés 25% off. ¿Te armo el pedido?'
				: /env[ií]o/.test(pregunta)
					? `Sí, por Andreani te llega en 48 h a ${c.ciudad}. El envío sale ${pesos(4200)}, o es gratis en compras desde ${pesos(60_000)}.`
					: '¡Hola! Ya lo reviso y te confirmo en un minuto.';
	const pagos = suyos.map((p) => p.pago);
	const masUsado = pagos.sort(
		(a, x) => pagos.filter((y) => y === x).length - pagos.filter((y) => y === a).length
	)[0];
	const NOMBRE: Record<string, string> = {
		mp: 'Mercado Pago',
		transf: 'transferencia',
		tarjeta: 'tarjeta',
		modo: 'MODO',
		payway: 'tarjeta',
		uala: 'Ualá',
		efectivo: 'efectivo'
	};
	const ESTADOS = [
		'Sin pagar',
		'Para preparar',
		'Listo para despachar',
		'En camino',
		'Entregado',
		'Cancelado'
	] as const;
	return {
		id: b.id,
		nombre: c.nombre,
		foto: c.foto,
		canal: b.canal,
		ultimo: ultimoTexto.texto,
		creado: ultimoTexto.en,
		atiende: b.atiende,
		sinLeer: b.sinLeer,
		mensajes: b.mensajes,
		clienteId: c.id,
		sugerencia: cambio?.mensajes?.length ? null : sugerencia,
		tomada: !!cambio?.tomada,
		cliente: {
			ciudad: c.ciudad,
			desde: iso(c.alta),
			gastado: suyos.reduce((s, p) => s + p.total, 0),
			pedidos: suyos.length,
			etiquetas: [
				...(c.mayorista ? ['Mayorista'] : []),
				...(suyos.length >= 4 ? ['Recurrente'] : suyos.length <= 1 ? ['Nueva'] : []),
				...(suyos.some((p) => p.pago === 'tarjeta' || p.pago === 'payway') ? ['Cuotas'] : [])
			],
			ultimoPedido: ultimo
				? {
						numero: ultimo.numero,
						resumen: ultimo.items.map((i) => i.nombre).join(', '),
						total: ultimo.total,
						estado: ESTADOS.includes(ultimo.estado) ? ultimo.estado : 'Entregado'
					}
				: null,
			sabe: suyos.length
				? `prefiere ${suyos.filter((p) => p.envio === 'retiro').length > suyos.length / 2 ? 'retirar en el local' : 'envío a domicilio'} y paga con ${NOMBRE[masUsado] ?? masUsado}.`
				: null
		}
	};
}

/** Lo que se hizo con los carritos. */
export type EstadoCarritos = {
	recordados: Set<string>;
	todos: boolean;
	automaticos: { hora: boolean; dia: boolean; email: boolean };
	/** Los tres recordatorios por mail (§62). */
	recordatorios: Carritos['recordatorios'];
};

const DESDE: Record<string, string> = {
	web: 'la tienda',
	ig: 'Instagram',
	wa: 'WhatsApp',
	ml: 'Mercado Libre',
	tg: 'Telegram',
	pos: 'el local'
};

export function carritosDemo(
	d: Demo,
	ahora: number,
	e: EstadoCarritos,
	costoMensaje: number
): Carritos {
	const clientes = new Map(d.clientes.map((c) => [c.id, c]));
	const productos = new Map(d.catalogo.productos.map((p) => [p.id, p]));
	const carritos = d.carritos.map((k) => {
		const c = clientes.get(k.clienteId)!;
		const canal = d.pedidos.find((p) => p.clienteId === c.id)?.canal ?? 'web';
		return {
			id: k.id,
			nombre: c.nombre,
			foto: c.foto,
			desde: DESDE[canal] ?? 'la tienda',
			creado: iso(Math.min(k.creado, ahora)),
			fotos: k.items.map((i) => productos.get(i.productoId)?.foto ?? null),
			cantidad: k.items.reduce((s, i) => s + i.cantidad, 0),
			total: k.total,
			recordado: e.todos || e.recordados.has(k.id)
		};
	});
	// Lo recuperado este mes: los clientes con carrito que después compraron.
	const mes = ahora - 30 * DIA;
	const conCarrito = new Set(d.carritos.map((k) => k.clienteId));
	const recuperados = d.pedidos.filter(
		(p) => conCarrito.has(p.clienteId) && p.creado > mes && p.estado !== 'Cancelado'
	);
	const recuperado = recuperados.reduce((s, p) => s + p.total, 0);
	const monto = carritos.reduce((s, k) => s + k.total, 0);
	return {
		carritos,
		abiertos: carritos.length,
		monto,
		recuperado,
		tasa: Math.round((recuperado / Math.max(1, recuperado + monto)) * 100),
		costoMensaje,
		automaticos: e.automaticos,
		recordatorios: e.recordatorios,
		recuperados: { pedidos: recuperados.length, monto: recuperado }
	};
}

/**
 * Un carrito que alguien dejó en el checkout de la tienda del visitante (uno abierto por email,
 * como en la API Go). Se cierra cuando ese email compra.
 */
export type CarritoTienda = {
	email: string;
	nombre: string;
	items: { id: string; varianteId: string | null; cantidad: number }[];
	total: number;
	token: string;
	creado: number;
	actualizado: number;
	/** Cuándo compró ese email (y cuánto), o null si sigue abierto. */
	recuperado: number | null;
	recuperadoTotal?: number;
	/** Pidió no recibir más recordatorios (§62). */
	baja?: number;
};

/** El id con el que el panel nombra un carrito de la tienda (no expone el email ni el token). */
export const idCarritoTienda = (c: Pick<CarritoTienda, 'token'>) => `t-${c.token.slice(0, 12)}`;

/** Suma a los carritos de ejemplo los que se dejaron de verdad en la tienda: arriba, los más nuevos. */
export function conCarritosTienda(
	base: Carritos,
	propios: CarritoTienda[],
	fotoDe: (id: string) => string | null,
	e: EstadoCarritos,
	ahora: number
): Carritos {
	const abiertos = propios
		.filter((c) => !c.recuperado)
		.sort((a, b) => b.actualizado - a.actualizado)
		.map((c) => ({
			id: idCarritoTienda(c),
			nombre: c.nombre || c.email,
			foto: null,
			desde: 'la tienda',
			creado: iso(c.actualizado),
			fotos: c.items.map((i) => fotoDe(i.id)),
			cantidad: c.items.reduce((s, i) => s + i.cantidad, 0),
			total: c.total,
			recordado: e.todos || e.recordados.has(idCarritoTienda(c))
		}));
	const mes = ahora - 30 * DIA;
	const deVerdad = propios.filter((c) => c.recuperado && c.recuperado > mes);
	const recuperado =
		base.recuperado + deVerdad.reduce((s, c) => s + (c.recuperadoTotal ?? c.total), 0);
	const carritos = [...abiertos, ...base.carritos];
	const monto = carritos.reduce((s, k) => s + k.total, 0);
	return {
		...base,
		carritos,
		abiertos: carritos.length,
		monto,
		recuperado,
		recuperados: { pedidos: base.recuperados.pedidos + deVerdad.length, monto: recuperado },
		tasa: Math.round((recuperado / Math.max(1, recuperado + monto)) * 100)
	};
}
