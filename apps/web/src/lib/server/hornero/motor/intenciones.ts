/**
 * Lo que Tincho sabe responder sin IA, con los datos reales de la tienda. Cada intención mira los
 * conceptos de la pregunta (no frases exactas) y arma el texto, los bloques para ver los datos y los
 * botones para ir a hacer algo.
 */
import { conMontos, type DatosHornero } from '../contexto.ts';
import { nombrePeriodo, type Periodo } from './periodo.ts';
import {
	delta,
	diaCorto,
	diaLargo,
	enumerar,
	num,
	pesos,
	plural,
	variacion,
	type Accion,
	type Bloque,
	type Respuesta
} from './respuesta.ts';
import { alguno, normalizar, parecida, tiene, VACIAS, type Pregunta } from './texto.ts';

export type Producto = DatosHornero['contexto']['productos'][number];
type Dia = DatosHornero['contexto']['ventasPorDia'][number];

export type Ctx = {
	p: Pregunta;
	d: DatosHornero;
	periodo: Periodo | null;
	montos: boolean;
	producto: Producto | null;
	/** Si nombra algo que coincide con varios («yerba»), todos. */
	productos?: Producto[];
};

const SIN_PLATA = ' Tu rol no ve montos de dinero.';

// ---------------------------------------------------------------- Productos

/** El producto que nombra la pregunta (por palabras del nombre, con errores de tipeo), o null. */
export function productoDe(pregunta: string, productos: Producto[]): Producto | null {
	return productosDe(pregunta, productos)[0] ?? null;
}

/** Todos los que empatan en el mejor puntaje («yerba» → las dos yerbas). */
export function productosDe(pregunta: string, productos: Producto[]): Producto[] {
	const palabras = normalizar(pregunta)
		.split(' ')
		.filter((w) => w.length > 2 && !VACIAS.has(w));
	if (!palabras.length) return [];
	const puntaje: { p: Producto; puntos: number }[] = [];
	for (const p of productos) {
		const nombre = normalizar(p.nombre).split(' ');
		let puntos = 0;
		for (const w of palabras) {
			const s = w.replace(/(es|s)$/, '');
			if (nombre.some((n) => n === w || n === s || n.replace(/(es|s)$/, '') === s)) puntos += 2;
			else if (nombre.some((n) => n.length >= 5 && (n.startsWith(s) || parecida(w, n))))
				puntos += 1;
		}
		if (puntos) puntaje.push({ p, puntos });
	}
	const max = Math.max(0, ...puntaje.map((x) => x.puntos));
	return puntaje.filter((x) => x.puntos === max).map((x) => x.p);
}

const estadoStock = (p: Producto) =>
	p.stock <= 0 ? 'sin stock' : p.stock <= 3 ? `quedan ${num(p.stock)}` : `${num(p.stock)} en stock`;
const tonoStock = (p: Producto): 'mal' | 'aviso' | 'ok' =>
	p.stock <= 0 ? 'mal' : p.stock <= 3 ? 'aviso' : 'ok';

// ---------------------------------------------------------------- Ventas por período

type Cuenta = {
	nombre: string;
	ventas: number | null;
	pedidos: number;
	antes: { ventas: number | null; pedidos: number; nombre: string } | null;
	serie: Dia[];
	serieAntes: Dia[];
};

const suma = (xs: Dia[]) => ({
	ventas: xs.reduce((n, x) => n + (x.ventas ?? 0), 0),
	pedidos: xs.reduce((n, x) => n + x.pedidos, 0)
});

/** Ventas y pedidos del período, con con qué compararlo. null: no hay datos de ese período. */
export function cuentaDe(per: Periodo, d: DatosHornero): Cuenta | null {
	const dias = d.contexto.ventasPorDia;
	const nombre = nombrePeriodo(per);
	switch (per.tipo) {
		case 'hoy':
			return {
				nombre,
				ventas: d.hoy.ventas?.actual ?? null,
				pedidos: d.hoy.kpis.pedidos.actual,
				antes: {
					ventas: d.hoy.ventas?.anterior ?? null,
					pedidos: d.hoy.kpis.pedidos.anterior,
					nombre: 'ayer a esta hora'
				},
				serie: [],
				serieAntes: []
			};
		case 'dia': {
			const x = dias.at(-1 - per.atras);
			if (!x) return null;
			const semanaAntes = dias.at(-1 - per.atras - 7);
			return {
				nombre: per.nombre,
				ventas: x.ventas ?? null,
				pedidos: x.pedidos,
				antes: semanaAntes
					? {
							ventas: semanaAntes.ventas ?? null,
							pedidos: semanaAntes.pedidos,
							nombre: 'el mismo día de la semana anterior'
						}
					: null,
				serie: [x],
				serieAntes: []
			};
		}
		case 'semana':
			return {
				nombre,
				ventas: d.semana.ventas?.actual ?? null,
				pedidos: d.semana.kpis.pedidos.actual,
				antes: {
					ventas: d.semana.ventas?.anterior ?? null,
					pedidos: d.semana.kpis.pedidos.anterior,
					nombre: 'los 7 días anteriores'
				},
				serie: dias.slice(-7),
				serieAntes: dias.length >= 14 ? dias.slice(-14, -7) : []
			};
		case 'semanaPasada': {
			if (dias.length < 14) return null;
			const s = suma(dias.slice(-14, -7));
			const esta = suma(dias.slice(-7));
			return {
				nombre,
				ventas: d.semana.montos ? s.ventas : null,
				pedidos: s.pedidos,
				antes: {
					ventas: d.semana.montos ? esta.ventas : null,
					pedidos: esta.pedidos,
					nombre: 'esta semana'
				},
				serie: dias.slice(-14, -7),
				serieAntes: []
			};
		}
		case 'ultimos': {
			const n = Math.min(per.dias, dias.length);
			const s = suma(dias.slice(-n));
			const previos = dias.length >= n * 2 ? suma(dias.slice(-n * 2, -n)) : null;
			return {
				nombre: `en los últimos ${n} días`,
				ventas: d.semana.montos ? s.ventas : null,
				pedidos: s.pedidos,
				antes: previos
					? {
							ventas: d.semana.montos ? previos.ventas : null,
							pedidos: previos.pedidos,
							nombre: `los ${n} días anteriores`
						}
					: null,
				serie: dias.slice(-n),
				serieAntes: dias.length >= n * 2 ? dias.slice(-n * 2, -n) : []
			};
		}
		case 'mes':
			if (!d.mes) return null;
			return {
				nombre,
				ventas: d.mes.ventas?.actual ?? null,
				pedidos: d.mes.kpis.pedidos.actual,
				antes: {
					ventas: d.mes.ventas?.anterior ?? null,
					pedidos: d.mes.kpis.pedidos.anterior,
					nombre: 'los 30 días anteriores'
				},
				serie: dias,
				serieAntes: []
			};
	}
}

function graficoDe(c: Cuenta, montos: boolean): Bloque | null {
	if (c.serie.length < 2) return null;
	return {
		tipo: 'linea',
		titulo: montos ? 'Ventas por día' : 'Pedidos por día',
		unidad: montos ? 'pesos' : 'pedidos',
		puntos: c.serie.map((x) => ({
			etiqueta: diaCorto(x.fecha),
			valor: montos ? (x.ventas ?? 0) : x.pedidos
		})),
		...(c.serieAntes.length === c.serie.length
			? { anterior: c.serieAntes.map((x) => (montos ? (x.ventas ?? 0) : x.pedidos)) }
			: {})
	};
}

export function ventas(x: Ctx): Respuesta {
	const per = x.periodo ?? { tipo: 'hoy' };
	const c = cuentaDe(per, x.d);
	if (!c)
		return {
			texto:
				per.tipo === 'mes'
					? 'Del mes no tengo el detalle ahora. Te muestro las últimas dos semanas: preguntame «¿cuánto vendí en los últimos 14 días?».'
					: `Todavía no tengo datos de ${nombrePeriodo(per)}.`,
			sugerencias: ['¿Cuánto vendí esta semana?', '¿Cuánto vendí en los últimos 14 días?']
		};
	const soloPedidos = tiene(x.p, 'pedido') && !alguno(x.p, 'vender', 'plata');
	const verPlata = x.montos && !soloPedidos && c.ventas !== null;
	const titulo = c.nombre.charAt(0).toUpperCase() + c.nombre.slice(1);
	const pedidos = plural(c.pedidos, 'pedido', 'pedidos');
	let texto: string;
	if (verPlata) {
		const verbo = per.tipo === 'hoy' ? 'llevás' : 'vendiste';
		texto = `${titulo} ${verbo} **${pesos(c.ventas!)}** en ${pedidos}${c.antes?.ventas != null ? variacion(c.ventas!, c.antes.ventas, c.antes.nombre) : ''}.`;
		if (c.pedidos && per.tipo !== 'hoy')
			texto += ` Ticket promedio: ${pesos(c.ventas! / c.pedidos)}.`;
	} else {
		texto = `${titulo} entraron **${pedidos}**${c.antes ? variacion(c.pedidos, c.antes.pedidos, c.antes.nombre) : ''}.${!x.montos && !soloPedidos ? SIN_PLATA : ''}`;
	}
	if (!c.pedidos && per.tipo === 'hoy')
		texto += ' Todavía no entró ninguno: un buen momento para compartir la tienda.';
	const cifras: Bloque = {
		tipo: 'cifras',
		items: [
			...(verPlata
				? [
						{
							titulo: 'Ventas',
							valor: pesos(c.ventas!),
							...(c.antes?.ventas != null ? delta(c.ventas!, c.antes.ventas) : {})
						}
					]
				: []),
			{
				titulo: 'Pedidos',
				valor: num(c.pedidos),
				...(c.antes ? delta(c.pedidos, c.antes.pedidos) : {})
			},
			...(verPlata && c.pedidos
				? [{ titulo: 'Ticket promedio', valor: pesos(c.ventas! / c.pedidos) }]
				: [])
		]
	};
	const g = graficoDe(c, verPlata);
	return {
		texto,
		bloques: [cifras, ...(g ? [g] : [])],
		acciones: [
			{ texto: 'Ver pedidos', ir: 'pedidos' },
			{ texto: 'Estadísticas', ir: 'estadisticas' }
		],
		sugerencias:
			per.tipo === 'hoy'
				? ['¿Y ayer?', '¿Cómo vengo esta semana?', '¿Qué es lo que más se vende?']
				: per.tipo === 'semana'
					? ['¿Y la semana pasada?', '¿Cuál fue mi mejor día?', '¿Por qué canal vendo más?']
					: ['¿Cuánto vendí hoy?', '¿Cuál fue mi mejor día?', '¿Qué es lo que más se vende?'],
		tema: 'ventas'
	};
}

export function mejorDia(x: Ctx, peor = false): Respuesta {
	const dias = x.d.contexto.ventasPorDia;
	const valor = (y: Dia) => (x.montos ? (y.ventas ?? 0) : y.pedidos);
	if (!dias.length || !dias.some(valor))
		return { texto: 'En las últimas dos semanas todavía no hubo ventas.' };
	const orden = [...dias].sort((a, b) => valor(b) - valor(a));
	const m = peor ? orden.at(-1)! : orden[0];
	const total = dias.reduce((n, y) => n + valor(y), 0);
	const promedio = total / dias.length;
	return {
		texto: `Tu ${peor ? 'día más flojo' : 'mejor día'} de las últimas dos semanas fue el **${diaLargo(m.fecha)}**: ${x.montos ? pesos(valor(m)) : plural(m.pedidos, 'pedido', 'pedidos')}. El promedio es ${x.montos ? pesos(promedio) : `${promedio.toLocaleString('es-AR', { maximumFractionDigits: 1 })} pedidos`} por día.`,
		bloques: [
			{
				tipo: 'barras',
				titulo: x.montos ? 'Ventas por día (14 días)' : 'Pedidos por día (14 días)',
				unidad: x.montos ? 'pesos' : 'pedidos',
				items: dias.map((y) => ({
					etiqueta: diaCorto(y.fecha),
					valor: valor(y),
					destacado: y.fecha === m.fecha
				}))
			}
		],
		sugerencias: ['¿Cuánto vendí esta semana?', '¿Por qué canal vendo más?'],
		tema: 'dias'
	};
}

export function ticket(x: Ctx): Respuesta {
	const tk = x.d.semana.kpis.ticket;
	if (!x.montos || !tk) return { texto: `No tengo el ticket promedio para vos.${SIN_PLATA}` };
	return {
		texto: `En los últimos 7 días cada compra fue de **${pesos(tk.actual)}** en promedio${variacion(tk.actual, tk.anterior, 'la semana anterior')}.`,
		bloques: [
			{
				tipo: 'cifras',
				items: [
					{ titulo: 'Ticket promedio', valor: pesos(tk.actual), ...delta(tk.actual, tk.anterior) }
				]
			}
		],
		sugerencias: ['¿Cómo subo el ticket promedio?', '¿Qué es lo que más se vende?'],
		tema: 'ticket'
	};
}

export function meta(x: Ctx): Respuesta {
	const m = x.d.semana.meta;
	if (!x.montos) return { texto: `La meta del mes es en pesos.${SIN_PLATA}` };
	if (!m || !m.objetivo)
		return {
			texto:
				'Todavía no definiste una meta para el mes. Ponela desde **Inicio** (tocando la meta) y te voy diciendo cómo venís.',
			acciones: [{ texto: 'Ver estadísticas', ir: 'estadisticas' }]
		};
	const pct = Math.round((m.actual / m.objetivo) * 100);
	const llega = m.proyeccion >= m.objetivo;
	return {
		texto: `Llevás **${pesos(m.actual)}** de ${pesos(m.objetivo)} (${pct}\u202f%). A este ritmo cerrás el mes en **${pesos(m.proyeccion)}**: ${llega ? '¡llegás a la meta!' : `te faltarían ${pesos(m.objetivo - m.proyeccion)}.`}`,
		bloques: [
			{
				tipo: 'progreso',
				titulo: 'Meta del mes',
				actual: m.actual,
				objetivo: m.objetivo,
				unidad: 'pesos',
				detalle: `Proyección: ${pesos(m.proyeccion)}`
			}
		],
		sugerencias: llega
			? ['¿Qué es lo que más se vende?']
			: ['Ideas para vender más', '¿Qué productos no se venden?'],
		tema: 'meta'
	};
}

// ---------------------------------------------------------------- Productos y stock

export function masVendidos(x: Ctx): Respuesta {
	const top = [...x.d.contexto.productos]
		.filter((p) => p.vendidos30d > 0)
		.sort((a, b) => b.vendidos30d - a.vendidos30d)
		.slice(0, 5);
	if (!top.length)
		return {
			texto: 'Todavía no hay ventas en los últimos 30 días para armar el ranking.',
			acciones: [{ texto: 'Compartir la tienda', ir: 'compartir' }]
		};
	const sinStock = top.filter((p) => p.stock <= 0);
	return {
		texto: `Lo que más salió en 30 días es **${top[0].nombre}** (${plural(top[0].vendidos30d, 'unidad', 'unidades')}).${sinStock.length ? ` Ojo: ${enumerar(sinStock.map((p) => `**${p.nombre}**`))} ${sinStock.length === 1 ? 'está' : 'están'} sin stock.` : ''}`,
		bloques: [
			{
				tipo: 'barras',
				titulo: 'Más vendidos (30 días, unidades)',
				unidad: 'unidades',
				items: top.map((p, i) => ({ etiqueta: p.nombre, valor: p.vendidos30d, destacado: i === 0 }))
			},
			{
				tipo: 'tabla',
				titulo: 'Detalle',
				columnas: x.montos
					? ['Producto', 'Vendidos', 'Stock', 'Precio']
					: ['Producto', 'Vendidos', 'Stock'],
				filas: top.map((p) => [
					p.nombre,
					num(p.vendidos30d),
					p.stock <= 0 ? 'Sin stock' : num(p.stock),
					...(x.montos ? [pesos(p.precio)] : [])
				])
			}
		],
		acciones: sinStock.length
			? [{ texto: 'Reponer stock', ir: 'sin-stock' }]
			: [{ texto: 'Ver productos', ir: 'productos' }],
		sugerencias: ['¿Qué productos no se venden?', '¿Qué tengo que reponer?'],
		tema: 'top'
	};
}

export function quietos(x: Ctx): Respuesta {
	const qs = x.d.contexto.productos
		.filter((p) => p.vendidos30d === 0 && p.stock > 0)
		.sort((a, b) => b.stock - a.stock);
	if (!qs.length)
		return {
			texto: 'Todos tus productos con stock tuvieron alguna venta en los últimos 30 días. ¡Bien!',
			tema: 'quietos'
		};
	return {
		texto: `**${plural(qs.length, 'producto', 'productos')}** con stock no ${qs.length === 1 ? 'tuvo' : 'tuvieron'} ventas en 30 días. Un cupón, una foto nueva o compartirlos en tus estados puede moverlos.`,
		bloques: [
			{
				tipo: 'lista',
				titulo: 'Sin ventas en 30 días',
				items: qs.slice(0, 8).map((p) => ({
					titulo: p.nombre,
					detalle: x.montos ? pesos(p.precio) : undefined,
					insignia: `${num(p.stock)} en stock`,
					tono: 'aviso' as const
				}))
			}
		],
		acciones: [
			{ texto: 'Crear un cupón', ir: 'cupones' },
			{ texto: 'Compartir la tienda', ir: 'compartir' }
		],
		sugerencias: ['¿Qué es lo que más se vende?', 'Ideas para vender más'],
		tema: 'quietos'
	};
}

export function stockBajo(x: Ctx): Respuesta {
	const ps = x.d.contexto.productos;
	const sin = ps.filter((p) => p.stock <= 0).sort((a, b) => b.vendidos30d - a.vendidos30d);
	const pocos = ps.filter((p) => p.stock > 0 && p.stock <= 3).sort((a, b) => a.stock - b.stock);
	if (!ps.length)
		return {
			texto: 'Todavía no cargaste productos.',
			acciones: [{ texto: 'Cargar un producto', ir: 'nuevo-producto' }]
		};
	if (!sin.length && !pocos.length)
		return {
			texto: 'Está todo con stock: ningún producto tiene 3 unidades o menos.',
			sugerencias: ['¿Qué es lo que más se vende?'],
			tema: 'stock'
		};
	const urgentes = sin.filter((p) => p.vendidos30d > 0);
	const partes = [];
	if (sin.length)
		partes.push(`**${plural(sin.length, 'producto está', 'productos están')} sin stock**`);
	if (pocos.length) partes.push(`a ${plural(pocos.length, 'le quedan', 'les quedan')} 3 o menos`);
	return {
		texto: `${partes.join(' y ')}.${urgentes.length ? ` Reponé primero ${enumerar(urgentes.slice(0, 3).map((p) => `**${p.nombre}**`))}: se venden.` : ''}`,
		bloques: [
			{
				tipo: 'lista',
				titulo: 'Para reponer',
				items: [...sin, ...pocos].slice(0, 10).map((p) => ({
					titulo: p.nombre,
					detalle: p.vendidos30d
						? `${plural(p.vendidos30d, 'vendido', 'vendidos')} en 30 días`
						: 'Sin ventas en 30 días',
					insignia: p.stock <= 0 ? 'Sin stock' : `Quedan ${num(p.stock)}`,
					tono: tonoStock(p)
				}))
			}
		],
		acciones: [{ texto: 'Ver sin stock', ir: 'sin-stock' }],
		sugerencias: ['¿Qué es lo que más se vende?', '¿Qué productos no se venden?'],
		tema: 'stock'
	};
}

export function unProducto(x: Ctx): Respuesta | null {
	const p = x.producto;
	if (!p) return null;
	const varios = x.productos ?? [];
	if (varios.length > 1) {
		const sin = varios.filter((y) => y.stock <= 0);
		return {
			texto: `Encontré **${plural(varios.length, 'producto', 'productos')}**: ${enumerar(
				varios.map((y) => `**${y.nombre}** (${estadoStock(y)})`),
				4
			)}.${sin.length ? ` ${sin.length === 1 ? 'Uno está' : `${sin.length} están`} sin stock.` : ''}`,
			bloques: [
				{
					tipo: 'tabla',
					titulo: 'Productos',
					columnas: x.montos
						? ['Producto', 'Stock', 'Precio', 'Vendidos (30 días)']
						: ['Producto', 'Stock', 'Vendidos (30 días)'],
					filas: varios
						.slice(0, 8)
						.map((y) => [
							y.nombre,
							y.stock <= 0 ? 'Sin stock' : num(y.stock),
							...(x.montos ? [pesos(y.precio)] : []),
							num(y.vendidos30d)
						])
				}
			],
			acciones: [{ texto: 'Ver productos', ir: 'productos' }],
			tema: 'producto'
		};
	}
	const nombre = normalizar(p.nombre);
	const sobran = normalizar(x.p.original)
		.split(' ')
		.filter(
			(w) =>
				w.length > 3 &&
				!VACIAS.has(w) &&
				!x.p.conceptos.has(w as never) &&
				!nombre.includes(w.replace(/(es|s)$/, ''))
		)
		.filter((w) => !/^(cuant|preci|stock|quedan|tengo|tiene|vend|sale|hay|cuesta|vale)/.test(w));
	const aviso = sobran.length ? `No encontré «${sobran.join(' ')}»; lo más parecido es esto. ` : '';
	const pide = tiene(x.p, 'precio')
		? 'precio'
		: alguno(x.p, 'vender', 'pedido')
			? 'ventas'
			: 'stock';
	const texto =
		pide === 'precio' && x.montos
			? `**${p.nombre}** sale **${pesos(p.precio)}** (${estadoStock(p)}).`
			: pide === 'ventas'
				? `De **${p.nombre}** vendiste **${plural(p.vendidos30d, 'unidad', 'unidades')}** en 30 días; ${estadoStock(p)}.`
				: `De **${p.nombre}** ${p.stock <= 0 ? 'no te queda stock' : `tenés **${plural(p.stock, 'unidad', 'unidades')}**`}. Vendiste ${plural(p.vendidos30d, 'unidad', 'unidades')} en 30 días.`;
	const dias = p.vendidos30d > 0 && p.stock > 0 ? Math.round(p.stock / (p.vendidos30d / 30)) : null;
	return {
		texto:
			aviso +
			texto +
			(dias !== null && dias <= 14
				? ` A este ritmo te alcanza para **${plural(dias, 'día', 'días')}**.`
				: ''),
		bloques: [
			{
				tipo: 'cifras',
				items: [
					...(x.montos ? [{ titulo: 'Precio', valor: pesos(p.precio) }] : []),
					{ titulo: 'Stock', valor: p.stock <= 0 ? 'Sin stock' : num(p.stock), tono: tonoStock(p) },
					{ titulo: 'Vendidos (30 días)', valor: num(p.vendidos30d) },
					...(p.categoria ? [{ titulo: 'Categoría', valor: p.categoria }] : [])
				]
			}
		],
		acciones: [{ texto: 'Ver productos', ir: 'productos' }],
		sugerencias: [`¿Cuánto sale ${p.nombre}?`, '¿Qué tengo que reponer?'].filter(
			(s) => pide !== 'precio' || !s.startsWith('¿Cuánto sale')
		),
		tema: 'producto'
	};
}

export function catalogo(x: Ctx): Respuesta {
	const ps = x.d.contexto.productos;
	if (!ps.length)
		return {
			texto:
				'Todavía no cargaste productos. Con una foto y el precio alcanza: en un minuto tenés el primero.',
			acciones: [
				{ texto: 'Cargar un producto', ir: 'nuevo-producto' },
				{ texto: 'Traer mi tienda', ir: 'importar' }
			]
		};
	const porCat = new Map<string, Producto[]>();
	for (const p of ps) {
		const c = p.categoria || 'Sin categoría';
		porCat.set(c, [...(porCat.get(c) ?? []), p]);
	}
	const cats = [...porCat.entries()].sort((a, b) => b[1].length - a[1].length);
	const valor = ps.reduce((n, p) => n + Math.max(0, p.stock) * p.precio, 0);
	const unidades = ps.reduce((n, p) => n + Math.max(0, p.stock), 0);
	return {
		texto: `Tenés **${plural(ps.length, 'producto', 'productos')}** en ${plural(cats.length, 'categoría', 'categorías')}, con ${plural(unidades, 'unidad', 'unidades')} en stock${x.montos ? ` (a precio de venta, **${pesos(valor)}**)` : ''}.`,
		bloques: [
			{
				tipo: 'cifras',
				items: [
					{ titulo: 'Productos', valor: num(ps.length) },
					{ titulo: 'Unidades en stock', valor: num(unidades) },
					...(x.montos ? [{ titulo: 'Stock a precio de venta', valor: pesos(valor) }] : []),
					{
						titulo: 'Sin stock',
						valor: num(ps.filter((p) => p.stock <= 0).length),
						tono: ps.some((p) => p.stock <= 0) ? ('aviso' as const) : ('ok' as const)
					}
				]
			},
			{
				tipo: 'tabla',
				titulo: 'Por categoría',
				columnas: ['Categoría', 'Productos', 'Vendidos (30 días)'],
				filas: cats
					.slice(0, 8)
					.map(([c, xs]) => [c, num(xs.length), num(xs.reduce((n, p) => n + p.vendidos30d, 0))])
			}
		],
		acciones: [{ texto: 'Ver productos', ir: 'productos' }],
		sugerencias: ['¿Qué es lo que más se vende?', '¿Cuál es mi producto más caro?'],
		tema: 'catalogo'
	};
}

export function extremosPrecio(x: Ctx, caro: boolean): Respuesta {
	const ps = [...x.d.contexto.productos]
		.filter((p) => p.precio > 0)
		.sort((a, b) => (caro ? b.precio - a.precio : a.precio - b.precio));
	if (!x.montos) return { texto: `Los precios son montos en pesos.${SIN_PLATA}` };
	if (!ps.length) return { texto: 'Todavía no hay productos con precio.' };
	return {
		texto: `Tu producto más ${caro ? 'caro' : 'barato'} es **${ps[0].nombre}**: **${pesos(ps[0].precio)}**.`,
		bloques: [
			{
				tipo: 'tabla',
				titulo: caro ? 'Los más caros' : 'Los más baratos',
				columnas: ['Producto', 'Precio', 'Stock'],
				filas: ps
					.slice(0, 5)
					.map((p) => [p.nombre, pesos(p.precio), p.stock <= 0 ? 'Sin stock' : num(p.stock)])
			}
		],
		sugerencias: [caro ? '¿Y el más barato?' : '¿Y el más caro?', '¿Qué es lo que más se vende?'],
		tema: caro ? 'caro' : 'barato'
	};
}

export function mercadoLibre(x: Ctx): Respuesta {
	const ps = x.d.contexto.productos;
	const en = ps.filter((p) => p.enMl);
	return en.length
		? {
				texto: `Tenés **${plural(en.length, 'producto publicado', 'productos publicados')}** en Mercado Libre de ${num(ps.length)}. El stock y el precio se sincronizan solos.`,
				bloques: [
					{
						tipo: 'progreso',
						titulo: 'Publicados en Mercado Libre',
						actual: en.length,
						objetivo: Math.max(1, ps.length),
						unidad: 'numero'
					}
				],
				acciones: [{ texto: 'Integraciones', ir: 'integraciones' }],
				tema: 'ml'
			}
		: {
				texto:
					'Todavía no publicaste productos en Mercado Libre. Conectás tu cuenta en **Integraciones** y después los publicás desde cada producto.',
				acciones: [{ texto: 'Conectar Mercado Libre', ir: 'integraciones' }],
				tema: 'ml'
			};
}

// ---------------------------------------------------------------- Clientes, canales y visitas

export function clientes(x: Ctx): Respuesta {
	const c = x.d.contexto.clientes;
	const volver = tiene(x.p, 'volver');
	const nuevos = tiene(x.p, 'nuevo');
	const mejores = alguno(x.p, 'mejor', 'mas') && !volver && !nuevos;
	const texto = mejores
		? `Tenés **${plural(c.total, 'cliente', 'clientes')}**. Quiénes te compran más lo ves en **Clientes**, ordenado por lo que gastó cada uno; desde ahí les escribís por WhatsApp.`
		: volver
			? c.sinVolver
				? `**${plural(c.sinVolver, 'cliente compró', 'clientes compraron')}** más de una vez y hace un mes que no ${c.sinVolver === 1 ? 'vuelve' : 'vuelven'}. Un mensaje con un cupón suele traerlos de vuelta.`
				: 'No tenés clientes frecuentes que hayan dejado de comprar. ¡Bien!'
			: nuevos
				? `En los últimos 30 días se sumaron **${plural(c.nuevos30d, 'cliente nuevo', 'clientes nuevos')}**.`
				: `Tenés **${plural(c.total, 'cliente', 'clientes')}**: ${num(c.nuevos30d)} nuevos en 30 días y ${num(c.conWhatsApp)} con WhatsApp.`;
	return {
		texto,
		bloques: [
			{
				tipo: 'cifras',
				items: [
					{ titulo: 'Clientes', valor: num(c.total) },
					{
						titulo: 'Nuevos (30 días)',
						valor: num(c.nuevos30d),
						...delta(x.d.semana.kpis.clientesNuevos.actual, x.d.semana.kpis.clientesNuevos.anterior)
					},
					{ titulo: 'Con WhatsApp', valor: num(c.conWhatsApp) },
					{ titulo: 'Sin volver', valor: num(c.sinVolver), tono: c.sinVolver ? 'aviso' : 'ok' }
				]
			}
		],
		acciones:
			volver && c.sinVolver
				? [
						{ texto: 'Ver clientes', ir: 'clientes' },
						{ texto: 'Crear un cupón', ir: 'cupones' }
					]
				: [{ texto: 'Ver clientes', ir: 'clientes' }],
		sugerencias: volver
			? ['¿Cuántos clientes nuevos tengo?']
			: ['¿Qué clientes dejaron de comprar?', '¿Por qué canal vendo más?'],
		tema: 'clientes'
	};
}

const CANALES: Record<string, string> = {
	web: 'Tienda online',
	wa: 'WhatsApp',
	ml: 'Mercado Libre',
	ig: 'Instagram',
	pos: 'El local',
	local: 'El local',
	tg: 'Telegram'
};

export function canales(x: Ctx): Respuesta {
	const valor = (c: DatosHornero['contexto']['canales30d'][number]) =>
		x.montos ? (c.ventas ?? 0) : c.pedidos;
	const cs = [...x.d.contexto.canales30d]
		.filter((c) => valor(c) > 0)
		.sort((a, b) => valor(b) - valor(a));
	if (!cs.length)
		return {
			texto: 'Todavía no hay ventas en los últimos 30 días para ver de dónde vienen.',
			acciones: [{ texto: 'Compartir la tienda', ir: 'compartir' }]
		};
	const total = cs.reduce((n, c) => n + valor(c), 0);
	const top = cs[0];
	return {
		texto: `En 30 días vendiste más por **${CANALES[top.canal] ?? top.canal}**: ${Math.round((valor(top) / total) * 100)}\u202f% ${x.montos ? `(${pesos(valor(top))})` : `(${plural(top.pedidos, 'pedido', 'pedidos')})`}.`,
		bloques: [
			{
				tipo: 'barras',
				titulo: x.montos ? 'Ventas por canal (30 días)' : 'Pedidos por canal (30 días)',
				unidad: x.montos ? 'pesos' : 'pedidos',
				items: cs.map((c, i) => ({
					etiqueta: CANALES[c.canal] ?? c.canal,
					valor: valor(c),
					destacado: i === 0
				}))
			}
		],
		sugerencias: ['¿Cuánto vendí este mes?', '¿Cuántas visitas tuvo la tienda?'],
		tema: 'canales'
	};
}

export function visitas(x: Ctx): Respuesta {
	const v = x.d.semana.visitas;
	if (!v?.desde)
		return {
			texto:
				'Todavía no hubo visitas medidas en tu tienda online. Compartí el link en tus estados y redes: cada visita la cuento acá.',
			acciones: [{ texto: 'Compartir la tienda', ir: 'compartir' }],
			tema: 'visitas'
		};
	const conv = x.d.semana.kpis.conversion.actual;
	const masVistos = v.masVistos.slice(0, 5);
	return {
		texto: `En los últimos 7 días visitaron tu tienda **${plural(v.dias7.actual, 'persona', 'personas')}**${variacion(v.dias7.actual, v.dias7.anterior, 'la semana anterior')}. Hoy, ${num(x.d.hoy.visitas?.hoy.actual ?? v.hoy.actual)}.${conv ? ` De cada 100, **${conv.toLocaleString('es-AR', { maximumFractionDigits: 1 })}** compran.` : ''}`,
		bloques: [
			{
				tipo: 'cifras',
				items: [
					{
						titulo: 'Visitas hoy',
						valor: num(x.d.hoy.visitas?.hoy.actual ?? v.hoy.actual),
						...delta(v.hoy.actual, v.hoy.anterior)
					},
					{
						titulo: 'Últimos 7 días',
						valor: num(v.dias7.actual),
						...delta(v.dias7.actual, v.dias7.anterior)
					},
					{ titulo: 'Páginas vistas', valor: num(v.vistas.actual) },
					...(conv
						? [
								{
									titulo: 'Compran (de 100)',
									valor: conv.toLocaleString('es-AR', { maximumFractionDigits: 1 })
								}
							]
						: [])
				]
			},
			...(masVistos.length
				? [
						{
							tipo: 'barras' as const,
							titulo: 'Lo más visto (7 días)',
							unidad: 'visitas' as const,
							items: masVistos.map((p) => ({ etiqueta: p.nombre, valor: p.vistas }))
						}
					]
				: [])
		],
		acciones: [
			{ texto: 'Compartir la tienda', ir: 'compartir' },
			{ texto: 'Estadísticas', ir: 'estadisticas' }
		],
		sugerencias: ['¿Por qué canal vendo más?', 'Ideas para vender más'],
		tema: 'visitas'
	};
}

export function carritos(x: Ctx): Respuesta {
	const h = x.d.semana.hornero;
	if (!h.carritos)
		return { texto: 'No tenés carritos abandonados en los últimos días. ¡Bien!', tema: 'carritos' };
	return {
		texto: `Hay **${plural(h.carritos, 'carrito abandonado', 'carritos abandonados')}**${x.montos && h.carritosMonto ? ` por **${pesos(h.carritosMonto)}**` : ''}. ${h.recordados ? 'Ya les mandamos el recordatorio a algunos.' : 'Mandales un recordatorio: muchos vuelven a comprar.'}`,
		bloques: [
			{
				tipo: 'cifras',
				items: [
					{ titulo: 'Carritos', valor: num(h.carritos), tono: 'aviso' },
					...(x.montos && h.carritosMonto
						? [{ titulo: 'En juego', valor: pesos(h.carritosMonto) }]
						: [])
				]
			}
		],
		acciones: [{ texto: 'Ver carritos', ir: 'carritos' }],
		tema: 'carritos'
	};
}

export function resenas(x: Ctx): Respuesta {
	const r = x.d.semana.resenas;
	if (!r.promedio && !r.nuevas.length)
		return {
			texto: 'Todavía no tenés reseñas. Después de cada compra le pedimos una al cliente.',
			tema: 'resenas'
		};
	return {
		texto: `Tus clientes te ponen **${(r.promedio ?? 0).toLocaleString('es-AR', { maximumFractionDigits: 1 })} estrellas** en promedio.${r.nuevas.length ? ` Tenés ${plural(r.nuevas.length, 'reseña nueva', 'reseñas nuevas')} para responder.` : ''}`,
		bloques: r.nuevas.length
			? [
					{
						tipo: 'lista',
						titulo: 'Reseñas nuevas',
						items: r.nuevas.slice(0, 5).map((n) => ({
							titulo: `${'★'.repeat(n.estrellas)}${'☆'.repeat(5 - n.estrellas)} ${n.autor}`,
							detalle: `${n.producto}: «${n.texto}»`,
							tono:
								n.estrellas >= 4
									? ('ok' as const)
									: n.estrellas <= 2
										? ('mal' as const)
										: ('neutro' as const)
						}))
					}
				]
			: [],
		acciones: [{ texto: 'Responder reseñas', ir: 'resenas' }],
		tema: 'resenas'
	};
}

// ---------------------------------------------------------------- El día: pendientes, pedidos, caja, envíos

export function pendientes(x: Ctx): Respuesta {
	const h = x.d.hoy.paraHacer;
	const items: {
		titulo: string;
		detalle?: string;
		insignia?: string;
		tono?: 'ok' | 'mal' | 'aviso' | 'neutro';
	}[] = [];
	const acciones: Accion[] = [];
	if (h.preparar) {
		items.push({ titulo: 'Pedidos para preparar', insignia: num(h.preparar), tono: 'aviso' });
		acciones.push({ texto: 'Preparar pedidos', ir: 'preparar' });
	}
	if (h.transferencias) {
		items.push({
			titulo: 'Transferencias para confirmar',
			detalle: x.montos && h.transferenciasMonto ? pesos(h.transferenciasMonto) : undefined,
			insignia: num(h.transferencias),
			tono: 'aviso'
		});
		acciones.push({ texto: 'Confirmar transferencias', ir: 'transferencias' });
	}
	if (h.chats)
		items.push({ titulo: 'Consultas sin responder', insignia: num(h.chats), tono: 'aviso' });
	if (h.sinStock) {
		items.push({ titulo: 'Productos sin stock', insignia: num(h.sinStock), tono: 'mal' });
		acciones.push({ texto: 'Ver sin stock', ir: 'sin-stock' });
	}
	if (tiene(x.p, 'transferencia') && !tiene(x.p, 'preparar'))
		return h.transferencias
			? {
					texto: `Tenés **${plural(h.transferencias, 'transferencia', 'transferencias')}** para confirmar${x.montos && h.transferenciasMonto ? ` por **${pesos(h.transferenciasMonto)}**` : ''}. Revisá que te llegó la plata y confirmá cada pedido.`,
					acciones: [{ texto: 'Confirmar transferencias', ir: 'transferencias' }],
					tema: 'pendientes'
				}
			: { texto: 'No tenés transferencias para confirmar.', tema: 'pendientes' };
	if (!items.length)
		return {
			texto:
				'No tenés nada pendiente: pedidos preparados, transferencias confirmadas y todo con stock. 🙌',
			sugerencias: ['¿Cómo vengo hoy?', 'Ideas para vender más'],
			tema: 'pendientes'
		};
	return {
		texto: `Tenés ${enumerar(items.map((i) => `**${i.insignia}** ${i.titulo.toLowerCase()}`))}.`,
		bloques: [{ tipo: 'lista', titulo: 'Para hacer hoy', items }],
		acciones: acciones.slice(0, 3),
		tema: 'pendientes'
	};
}

const ESTADOS: Record<string, string> = {
	Nuevo: 'Nuevo',
	Pagado: 'Pagado',
	'Para preparar': 'Para preparar',
	Preparado: 'Preparado',
	Enviado: 'Enviado',
	Entregado: 'Entregado',
	Cancelado: 'Cancelado'
};

export function ultimosPedidos(x: Ctx): Respuesta {
	const ps = x.d.hoy.ultimosPedidos.length ? x.d.hoy.ultimosPedidos : x.d.semana.ultimosPedidos;
	if (!ps.length)
		return {
			texto: 'Todavía no entró ningún pedido.',
			acciones: [{ texto: 'Compartir la tienda', ir: 'compartir' }]
		};
	const u = ps[0];
	const hora = new Date(u.creado)
		.toLocaleString('es-AR', {
			timeZone: 'America/Argentina/Buenos_Aires',
			weekday: 'short',
			hour: '2-digit',
			minute: '2-digit',
			hourCycle: 'h23'
		})
		.replace('.', '');
	return {
		texto: `El último pedido es el **#${u.numero}** de ${u.cliente}${x.montos && u.total != null ? ` por **${pesos(u.total)}**` : ''} (${hora}).`,
		bloques: [
			{
				tipo: 'tabla',
				titulo: 'Últimos pedidos',
				columnas: x.montos ? ['#', 'Cliente', 'Total', 'Estado'] : ['#', 'Cliente', 'Estado'],
				filas: ps
					.slice(0, 6)
					.map((p) => [
						`#${p.numero}`,
						p.cliente,
						...(x.montos ? [p.total != null ? pesos(p.total) : '—'] : []),
						ESTADOS[p.estado] ?? p.estado
					])
			}
		],
		acciones: [{ texto: 'Ver pedidos', ir: 'pedidos' }],
		sugerencias: ['¿Qué tengo pendiente?', '¿Cuánto vendí hoy?'],
		tema: 'ultimos'
	};
}

export function caja(x: Ctx): Respuesta {
	const c = x.d.hoy.caja;
	if (!x.montos) return { texto: `La caja es en pesos.${SIN_PLATA}` };
	if (!c?.abierta)
		return {
			texto: 'La caja del día está cerrada. La abrís cuando empezás a vender en el local.',
			acciones: [{ texto: 'Abrir la caja', ir: 'caja' }],
			tema: 'caja'
		};
	return {
		texto: `La caja de hoy lleva **${pesos(c.total)}**.`,
		bloques: c.medios.length
			? [
					{
						tipo: 'barras',
						titulo: 'Por medio de pago',
						unidad: 'pesos',
						items: c.medios.map((m) => ({ etiqueta: m.medio, valor: m.monto }))
					}
				]
			: [],
		acciones: [
			{ texto: 'Ver la caja', ir: 'caja' },
			{ texto: 'Vender en el local', ir: 'venta' }
		],
		tema: 'caja'
	};
}

export function envios(x: Ctx): Respuesta {
	const e = x.d.hoy.envios;
	if (!e?.paquetes && !e?.retiros.length)
		return {
			texto: 'Hoy no tenés paquetes para despachar.',
			acciones: [{ texto: 'Ver pedidos', ir: 'pedidos' }],
			tema: 'envios'
		};
	const proximo = e.retiros.find((r) => r.proximo) ?? e.retiros[0];
	return {
		texto: `Hoy despachás **${plural(e.paquetes, 'paquete', 'paquetes')}**.${proximo ? ` El próximo retiro es a las **${proximo.hora}** (${proximo.transporte}).` : ''}`,
		bloques: e.retiros.length
			? [
					{
						tipo: 'tabla',
						titulo: 'Retiros de hoy',
						columnas: ['Hora', 'Transporte', 'Paquetes'],
						filas: e.retiros.map((r) => [r.hora, r.transporte, num(r.paquetes)])
					}
				]
			: [],
		acciones: [{ texto: 'Preparar pedidos', ir: 'preparar' }],
		tema: 'envios'
	};
}

// ---------------------------------------------------------------- Resumen e ideas

export function resumen(x: Ctx): Respuesta {
	const { hoy, semana } = x.d;
	const v = hoy.ventas;
	const h = hoy.paraHacer;
	const pend = h.preparar + h.transferencias;
	const visitasHoy = hoy.visitas?.hoy.actual ?? semana.visitas?.hoy.actual;
	const venta =
		x.montos && v
			? `llevás **${pesos(v.actual)}** en ${plural(hoy.kpis.pedidos.actual, 'pedido', 'pedidos')}${variacion(v.actual, v.anterior, 'ayer a esta hora')}`
			: `entraron **${plural(hoy.kpis.pedidos.actual, 'pedido', 'pedidos')}**`;
	return {
		texto: `${x.d.hola ? `${x.d.hola}, hoy` : 'Hoy'} ${venta}.${pend ? ` Tenés ${plural(pend, 'cosa pendiente', 'cosas pendientes')}.` : ' No tenés nada pendiente.'}`,
		bloques: [
			{
				tipo: 'cifras',
				items: [
					...(x.montos && v
						? [{ titulo: 'Ventas hoy', valor: pesos(v.actual), ...delta(v.actual, v.anterior) }]
						: []),
					{
						titulo: 'Pedidos hoy',
						valor: num(hoy.kpis.pedidos.actual),
						...delta(hoy.kpis.pedidos.actual, hoy.kpis.pedidos.anterior)
					},
					...(visitasHoy != null && semana.visitas?.desde
						? [{ titulo: 'Visitas hoy', valor: num(visitasHoy) }]
						: []),
					...(x.montos && semana.ventas
						? [
								{
									titulo: 'Últimos 7 días',
									valor: pesos(semana.ventas.actual),
									...delta(semana.ventas.actual, semana.ventas.anterior)
								}
							]
						: [])
				]
			},
			...(pend || h.sinStock
				? [
						{
							tipo: 'lista' as const,
							titulo: 'Para hacer',
							items: [
								...(h.preparar
									? [
											{
												titulo: 'Pedidos para preparar',
												insignia: num(h.preparar),
												tono: 'aviso' as const
											}
										]
									: []),
								...(h.transferencias
									? [
											{
												titulo: 'Transferencias para confirmar',
												insignia: num(h.transferencias),
												tono: 'aviso' as const
											}
										]
									: []),
								...(h.sinStock
									? [
											{
												titulo: 'Productos sin stock',
												insignia: num(h.sinStock),
												tono: 'mal' as const
											}
										]
									: [])
							]
						}
					]
				: [])
		],
		acciones: [
			...(h.preparar ? [{ texto: 'Preparar pedidos', ir: 'preparar' as const }] : []),
			...(h.transferencias
				? [{ texto: 'Confirmar transferencias', ir: 'transferencias' as const }]
				: []),
			{ texto: 'Ver pedidos', ir: 'pedidos' as const }
		].slice(0, 3),
		sugerencias: ['¿Cuánto vendí esta semana?', '¿Qué tengo que reponer?', 'Ideas para vender más'],
		tema: 'resumen'
	};
}

/** Ideas con los datos: lo más urgente o lo que más plata trae primero. */
export function ideas(x: Ctx): Respuesta {
	const { contexto, semana, hoy } = x.d;
	const ps = contexto.productos;
	const out: { titulo: string; detalle: string; accion: Accion; peso: number }[] = [];
	const agotados = ps.filter((p) => p.stock <= 0 && p.vendidos30d > 0);
	if (agotados.length)
		out.push({
			titulo: `Reponé ${agotados[0].nombre}`,
			detalle: `Se vende (${plural(agotados[0].vendidos30d, 'unidad', 'unidades')} en 30 días) y está sin stock.`,
			accion: { texto: 'Reponer', ir: 'sin-stock' },
			peso: 10
		});
	if (hoy.paraHacer.transferencias)
		out.push({
			titulo: 'Confirmá las transferencias',
			detalle: `Hay ${plural(hoy.paraHacer.transferencias, 'pedido esperando', 'pedidos esperando')} tu confirmación.`,
			accion: { texto: 'Confirmar', ir: 'transferencias' },
			peso: 9
		});
	if (semana.hornero.carritos)
		out.push({
			titulo: 'Recuperá los carritos',
			detalle: `${plural(semana.hornero.carritos, 'persona dejó', 'personas dejaron')} productos en el carrito${x.montos && semana.hornero.carritosMonto ? ` (${pesos(semana.hornero.carritosMonto)})` : ''}.`,
			accion: { texto: 'Ver carritos', ir: 'carritos' },
			peso: 8
		});
	if (contexto.clientes.sinVolver)
		out.push({
			titulo: 'Escribiles a los que no volvieron',
			detalle: `${plural(contexto.clientes.sinVolver, 'cliente frecuente', 'clientes frecuentes')} hace un mes que no compra${contexto.clientes.sinVolver === 1 ? '' : 'n'}: un cupón los trae.`,
			accion: { texto: 'Crear cupón', ir: 'cupones' },
			peso: 7
		});
	const quietos = ps.filter((p) => p.vendidos30d === 0 && p.stock > 0);
	if (quietos.length >= 2)
		out.push({
			titulo: 'Mové lo que está quieto',
			detalle: `${plural(quietos.length, 'producto con stock no se vendió', 'productos con stock no se vendieron')} en 30 días: probá un descuento.`,
			accion: { texto: 'Crear cupón', ir: 'cupones' },
			peso: 5
		});
	const visitas7 = semana.visitas?.dias7.actual ?? 0;
	if (visitas7 < 50)
		out.push({
			titulo: 'Compartí tu tienda',
			detalle: visitas7
				? `Tuviste ${plural(visitas7, 'visita', 'visitas')} en 7 días: un estado de WhatsApp o una historia te trae más.`
				: 'Todavía casi nadie la vio: compartila en tus estados y redes.',
			accion: { texto: 'Compartir', ir: 'compartir' },
			peso: 6
		});
	if (ps.length < 6)
		out.push({
			titulo: 'Sumá más productos',
			detalle: `Tenés ${plural(ps.length, 'producto', 'productos')}: las tiendas con más variedad venden más. Con una foto alcanza.`,
			accion: { texto: 'Cargar producto', ir: 'nuevo-producto' },
			peso: 6
		});
	if (ps.length && !ps.some((p) => p.enMl))
		out.push({
			titulo: 'Publicá en Mercado Libre',
			detalle: 'Tus productos pueden estar también ahí, con el stock sincronizado.',
			accion: { texto: 'Conectar', ir: 'integraciones' },
			peso: 3
		});
	if (semana.resenas.nuevas.length)
		out.push({
			titulo: 'Respondé las reseñas',
			detalle: `Tenés ${plural(semana.resenas.nuevas.length, 'reseña nueva', 'reseñas nuevas')}: responderlas da confianza.`,
			accion: { texto: 'Responder', ir: 'resenas' },
			peso: 4
		});
	const top = out.sort((a, b) => b.peso - a.peso).slice(0, 4);
	if (!top.length)
		return {
			texto:
				'Viene todo muy bien: no veo nada urgente. Seguí compartiendo la tienda y cargando novedades.',
			acciones: [{ texto: 'Compartir', ir: 'compartir' }],
			tema: 'ideas'
		};
	return {
		texto: `Te propongo ${top.length === 1 ? 'una cosa' : `${top.length} cosas`}, de la más importante a la menos: empezá por **${top[0].titulo.toLowerCase()}**.`,
		bloques: [
			{
				tipo: 'lista',
				titulo: 'Ideas para hoy',
				items: top.map((i) => ({ titulo: i.titulo, detalle: i.detalle }))
			}
		],
		acciones: top
			.slice(0, 3)
			.map((i) => i.accion)
			.filter((a, i, xs) => xs.findIndex((b) => b.ir === a.ir) === i),
		sugerencias: ['¿Qué tengo que reponer?', '¿Qué clientes dejaron de comprar?'],
		tema: 'ideas'
	};
}

export const montosDe = (d: DatosHornero) => conMontos(d);
