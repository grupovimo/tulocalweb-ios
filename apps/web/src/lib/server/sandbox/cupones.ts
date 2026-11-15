/**
 * Cupones de la demo (design/PanelCupones.dc.html). Los nueve de ejemplo del handoff, con los
 * productos y categorías de la tienda elegida y lo vendido en proporción a sus pedidos. Lo que hace
 * el visitante (pausar, crear, duplicar, poner mínimo) se guarda como cambios.
 */
import type { components } from '@mostrador/contrato/esquema';
import type { Demo } from './demo.ts';

type Cupon = components['schemas']['Cupon'];
type Cupones = components['schemas']['Cupones'];
type Automatica = components['schemas']['Automatica'];

const DIA = 86_400_000;
/**
 * El envío gratis de una tienda que nunca cambió el monto (como cupones.EnvioGratisPorDefecto en Go):
 * cada tienda pone el suyo en Envíos y retiro.
 */
export const ENVIO_GRATIS_DESDE = 40_000;
const iso = (t: number) => new Date(t).toISOString();
const pesos = (n: number) => '$ ' + Math.round(n).toLocaleString('es-AR');

/** Lo que el visitante cambió de un cupón de ejemplo. */
export type CambioCupon = Partial<
	Pick<Cupon, 'estado' | 'minimo' | 'descripcion' | 'consejo' | 'usos' | 'vendio' | 'desconto'>
>;
export type EstadoRelampago = { horas: number; liberadas: number; terminada: boolean };

/** El domingo que viene a las 23:59 de Argentina. */
function domingo(ahora: number) {
	const ar = new Date(ahora - 3 * 3_600_000);
	const dias = (7 - ar.getUTCDay()) % 7;
	return Date.UTC(ar.getUTCFullYear(), ar.getUTCMonth(), ar.getUTCDate() + dias, 26, 59);
}

/** Los nueve del handoff sobre la tienda: [peso en lo vendido, descuento/vendido]. */
function ejemplos(d: Demo, ahora: number, creado: number): Cupon[] {
	const [cat1, cat2] = [
		d.catalogo.categorias[0] ?? 'productos',
		d.catalogo.categorias[1] ?? 'la tienda'
	];
	const prod = d.catalogo.productos[0];
	const mes = d.pedidos
		.filter((p) => p.estado !== 'Cancelado' && p.creado > ahora - 30 * DIA && p.creado <= ahora)
		.reduce((s, p) => s + p.total, 0);
	const cliente = d.clientes.find((c) => !c.mayorista)?.nombre.split(' ')[0] ?? 'LUCIA';
	const handle = d.catalogo.tienda
		.toLowerCase()
		.replace(/[^a-z]/g, '')
		.slice(0, 6);
	const base = (
		id: string,
		codigo: string,
		tipo: Cupon['tipo'],
		valor: number,
		descripcion: string,
		estado: Cupon['estado'],
		usos: number,
		limite: number,
		peso: number,
		ratio: number,
		canales: Cupon['canales'],
		empieza: number | null,
		termina: number | null,
		consejo: string | null
	): Cupon => {
		const vendio = Math.round((mes * peso) / 1000) * 1000;
		return {
			id,
			codigo,
			tipo,
			valor,
			descripcion,
			estado,
			usos,
			limite,
			vendio,
			desconto: Math.round(vendio * ratio),
			canales,
			empieza: empieza === null ? null : iso(empieza),
			termina: termina === null ? null : iso(termina),
			minimo: tipo === 'envio' ? 40_000 : 0,
			consejo,
			creado: iso(creado - (10 - Number(id.slice(1))) * 9 * DIA)
		};
	};
	const lower = (s: string) => s.toLowerCase();
	return [
		base(
			'c1',
			'FINDE10',
			'pct',
			10,
			`10 % en ${lower(cat1)}`,
			'activo',
			41,
			100,
			0.046,
			0.1,
			['web', 'wa', 'ig'],
			null,
			domingo(ahora),
			'Va 3 veces mejor por WhatsApp. ¿Lo mandamos a tus clientes VIP?'
		),
		base(
			'c2',
			'BIENVENIDA',
			'fijo',
			3000,
			'$ 3.000 en la primera compra',
			'activo',
			128,
			0,
			0.058,
			0.166,
			['web'],
			null,
			null,
			'El 71 % ya tenía el carrito armado. Pedí un mínimo.'
		),
		base(
			'c3',
			'ENVIOGRATIS',
			'envio',
			0,
			'Envío gratis desde $ 40.000',
			'activo',
			96,
			0,
			0.13,
			0.096,
			['web', 'wa'],
			null,
			null,
			null
		),
		base(
			'c4',
			`${cliente
				.toUpperCase()
				.normalize('NFD')
				.replace(/[^A-Z]/g, '')}15`,
			'pct',
			15,
			'Reseña con foto · 1 uso',
			'activo',
			1,
			1,
			0.0015,
			0.15,
			['web'],
			null,
			ahora + 30 * DIA,
			null
		),
		base(
			'c5',
			'HOTSALE30',
			'pct',
			30,
			'30 % en toda la tienda',
			'programado',
			0,
			300,
			0,
			0,
			['web', 'wa', 'ig', 'pos'],
			ahora + 20 * DIA,
			ahora + 23 * DIA,
			`Con 30 % ${prod ? `${prod.nombre} queda` : 'algunos productos quedan'} con 9 % de margen. Probá 25 %.`
		),
		base(
			'c6',
			`INFLU-${handle.toUpperCase()}`,
			'pct',
			12,
			`12 % · código de @${handle}.ok`,
			'activo',
			63,
			200,
			0.069,
			0.12,
			['web', 'ig'],
			null,
			ahora + 12 * DIA,
			'Es tu mejor código: cada $ 1 de descuento trae $ 8,3.'
		),
		base(
			'c7',
			'VUELVE20',
			'pct',
			20,
			'Clientes que no compran hace 90 días',
			'pausado',
			22,
			150,
			0.015,
			0.2,
			['wa'],
			null,
			null,
			null
		),
		base(
			'c8',
			'PRIMAVERA',
			'pct',
			15,
			`15 % en ${lower(cat2)}`,
			'terminado',
			210,
			210,
			0.099,
			0.15,
			['web', 'wa'],
			null,
			ahora - 22 * DIA,
			null
		),
		base(
			'c9',
			'2X1',
			'regalo',
			0,
			`Llevá 2 de ${lower(cat1)}, pagá 1`,
			'terminado',
			54,
			60,
			0.05,
			0.5,
			['web', 'pos'],
			null,
			ahora - 68 * DIA,
			null
		)
	];
}

/** Promociones automáticas: [id, título, detalle, ejemplo, estado, peso, prendida de entrada]. */
function automaticas(d: Demo, mes: number, prendidas: Map<string, boolean>): Automatica[] {
	const prods = [...d.catalogo.productos].sort((a, b) => b.precio - a.precio);
	const caro = prods[0];
	const barato = prods.at(-1);
	const AUTOS: [Automatica['id'], string, string, string, string, number, boolean][] = [
		[
			'transf',
			'Descuento por transferencia',
			// Lo mismo que aplica la tienda de la demo (cobros.transferencia = 15).
			'15 % si pagan con transferencia',
			'Carrito de $ 50.000 → paga $ 42.500',
			'312 pedidos este mes',
			0.23,
			true
		],
		[
			'envio',
			'Envío gratis desde un monto',
			'Desde $ 40.000 a todo el país',
			'Le faltan $ 6.000 → “Sumá algo y el envío es gratis”',
			'Ticket subió 18 %',
			0.13,
			true
		],
		[
			'cantidad',
			'Descuento por cantidad',
			'3 o más unidades: 15 % off',
			`3 × ${barato?.nombre ?? 'producto'} → 15 % off automático`,
			'86 pedidos',
			0.05,
			true
		],
		[
			'3x2',
			'3x2 en una categoría',
			`${d.catalogo.categorias.at(-1) ?? 'Una categoría'}: llevás 3, pagás 2`,
			'3 unidades → la más barata gratis',
			'',
			0,
			false
		],
		[
			'cuotas',
			'Cuotas sin interés',
			'3 y 6 cuotas desde $ 30.000',
			caro
				? `${caro.nombre} ${pesos(caro.precio)} → 6 × ${pesos(Math.ceil(caro.precio / 6))}`
				: '6 cuotas sin interés',
			'41 % de las ventas con tarjeta',
			0.19,
			true
		],
		[
			'mayorista',
			'Precio mayorista',
			'Desde 12 unidades, lista Mayorista',
			'12 unidades → precio de lista mayorista',
			`${d.clientes.filter((c) => c.mayorista).length} clientes mayoristas`,
			0.09,
			true
		],
		[
			'carrito',
			'Rescate de carrito con Tincho',
			'Ofrece hasta 10 % si el carrito se enfría',
			'Carrito abandonado 2 h → “Te guardo 10 % hasta las 20 h”',
			`${d.carritos.length * 3} carritos recuperados`,
			0.047,
			true
		],
		[
			'cumple',
			'Cumpleaños',
			'15 % el día del cumple del cliente',
			'Le llega por WhatsApp a las 9 h',
			'22 este mes',
			0.01,
			false
		]
	];
	return AUTOS.map(([id, titulo, detalle, ejemplo, estado, peso, deEntrada]) => {
		const activa = prendidas.get(id) ?? deEntrada;
		return {
			id,
			titulo,
			detalle,
			ejemplo,
			estado: activa && estado ? estado : activa ? 'Recién prendida' : 'Apagada',
			aporte: activa && peso ? Math.round((mes * peso) / 1000) * 1000 : null,
			activa
		};
	});
}

export function cuponesDemo(
	d: Demo,
	ahora: number,
	creado: number,
	cambios: Map<string, CambioCupon>,
	nuevos: Cupon[],
	consejoOculto: boolean,
	prendidas: Map<string, boolean>,
	relampago: EstadoRelampago,
	segmentos: { vip: number; riesgo: number }
): Cupones {
	const cupones = [
		...nuevos,
		...ejemplos(d, ahora, creado).map((c) => ({ ...c, ...cambios.get(c.id) }))
	].map((c) =>
		// Lo programado arranca solo cuando llega el día; lo vencido termina.
		c.estado === 'programado' && c.empieza && Date.parse(c.empieza) <= ahora
			? { ...c, estado: 'activo' as const }
			: c.estado === 'activo' && c.termina && Date.parse(c.termina) < ahora
				? { ...c, estado: 'terminado' as const }
				: c
	);
	const mes = d.pedidos
		.filter((p) => p.estado !== 'Cancelado' && p.creado > ahora - 30 * DIA && p.creado <= ahora)
		.reduce((s, p) => s + p.total, 0);
	const activos = cupones.filter((c) => c.estado === 'activo');
	const vendido = cupones.reduce((s, c) => s + c.vendio, 0);
	const bienvenida = cupones.find((c) => c.id === 'c2');
	const vendidos = new Map<string, number>();
	for (const p of d.pedidos)
		for (const i of p.items)
			vendidos.set(i.productoId, (vendidos.get(i.productoId) ?? 0) + i.cantidad);
	const porVenta = [...d.catalogo.productos].sort(
		(a, b) => (vendidos.get(b.id) ?? 0) - (vendidos.get(a.id) ?? 0)
	);
	const estrella = porVenta[0];
	const cats = new Map<string, number>();
	for (const p of porVenta)
		if (p.categoria)
			cats.set(p.categoria, (cats.get(p.categoria) ?? 0) + (vendidos.get(p.id) ?? 0));
	const unidades = 50 + relampago.liberadas * 20;
	const termina = creado + 7322_000 + relampago.horas * 3_600_000;
	return {
		cupones,
		kpis: {
			activos: activos.length,
			usosMes: activos.reduce((s, c) => s + c.usos, 0),
			vendido,
			descontado: cupones.reduce((s, c) => s + c.desconto, 0),
			vsMesAnterior: 24
		},
		consejo:
			!consejoOculto && bienvenida && !bienvenida.minimo
				? {
						cuponId: 'c2',
						codigo: 'BIENVENIDA',
						titulo: 'BIENVENIDA te cuesta más de lo que trae',
						texto: `El 71 % de quienes lo usan ya tenían el carrito armado. Si lo bajás a $ 2.000 o pedís un mínimo de $ 25.000, ahorrás unos ${pesos(Math.round((bienvenida.desconto * 0.25) / 1000) * 1000)} por mes.`,
						minimo: 25_000
					}
				: null,
		segmentos,
		automaticas: automaticas(d, mes, prendidas),
		relampago: {
			vivo: estrella
				? {
						producto: estrella.nombre,
						foto: estrella.foto,
						pct: 25,
						precio: estrella.precio,
						oferta: Math.round(estrella.precio * 0.75),
						termina: iso(termina),
						vendidos: 38,
						unidades,
						terminada: relampago.terminada || termina < ahora
					}
				: null,
			proximas: [
				[porVenta[1], '20 % off', 'Viernes 18 h · 3 horas · 30 unidades', 'Vie'],
				[porVenta[2], '3x2', 'Sábado 10 h · todo el día', 'Sáb'],
				[porVenta[3], '15 %', 'Black Friday · 27 de noviembre', 'Nov']
			]
				.filter(([p]) => p)
				.map(([p, t, detalle, cuando]) => {
					const x = p as (typeof porVenta)[number];
					return {
						titulo: `${x.nombre} · ${t as string}`,
						detalle: detalle as string,
						cuando: cuando as string,
						foto: x.foto
					};
				})
		},
		tienda: d.catalogo.tienda,
		productos: porVenta.slice(0, 6).map((p) => ({ id: p.id, nombre: p.nombre, foto: p.foto })),
		categoria: [...cats].sort((a, b) => b[1] - a[1])[0]?.[0] ?? null
	};
}
