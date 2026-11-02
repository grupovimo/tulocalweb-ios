/**
 * El motor de Tincho sin IA: entiende las preguntas de todos los días (ventas, pedidos, stock y
 * precio de un producto, lo que más y lo que menos sale, clientes, canales, Mercado Libre y los
 * "¿cómo hago para…?") y las contesta con los datos reales de la tienda, al instante y sin costo.
 * Lo que no entiende devuelve null y recién ahí responde el modelo (Workers AI).
 */
import { conMontos, type DatosHornero } from './contexto.ts';

type Producto = DatosHornero['contexto']['productos'][number];

const pesos = (n: number) => '$ ' + Math.round(n).toLocaleString('es-AR');
const num = (n: number) => Math.round(n).toLocaleString('es-AR');
const plural = (n: number, uno: string, varios: string) => `${num(n)} ${n === 1 ? uno : varios}`;

/** Minúsculas, sin tildes ni signos: "¿Cuánto vendí?" → "cuanto vendi". */
export const normalizar = (t: string) =>
	t
		.toLowerCase()
		.normalize('NFD')
		.replace(/[̀-ͯ]/g, '')
		.replace(/[^a-z0-9%$ ]+/g, ' ')
		.replace(/\s+/g, ' ')
		.trim();

const VACIAS = new Set(
	'el la los las de del un una unos unas y o a en por para con que cuanto cuantos cuanta cuantas cual cuales tengo tenes hay me queda quedan stock precio sale cuesta vale esta estan del al mi mis tu tus se lo le es son como'.split(
		' '
	)
);

/** El producto que nombra la pregunta (por palabras del nombre), o null. */
export function productoDe(pregunta: string, productos: Producto[]): Producto | null {
	const palabras = normalizar(pregunta)
		.split(' ')
		.filter((p) => p.length > 2 && !VACIAS.has(p));
	if (!palabras.length) return null;
	let mejor: { p: Producto; puntos: number } | null = null;
	for (const p of productos) {
		const nombre = normalizar(p.nombre);
		const puntos = palabras.filter(
			(w) => nombre.includes(w) || nombre.includes(w.replace(/s$/, ''))
		).length;
		if (puntos && (!mejor || puntos > mejor.puntos)) mejor = { p, puntos };
	}
	return mejor?.p ?? null;
}

const variacion = (actual: number, anterior: number) => {
	if (!anterior) return '';
	const pct = Math.round(((actual - anterior) / anterior) * 100);
	if (pct === 0) return ', igual que el período anterior';
	return pct > 0
		? `, un ${pct}% más que el período anterior`
		: `, un ${-pct}% menos que el período anterior`;
};

/** Los "¿cómo hago…?": pasos cortos y la sección del panel donde se hace. */
const GUIAS: [RegExp, string][] = [
	[
		/(cargo|cargar|agrego|agregar|subo|subir|crear|creo|nuevo) (un )?producto/,
		'Andá a **Productos → Nuevo producto**: primero elegís qué es (ropa, comida, hogar…) y después cargás nombre, precio, stock, fotos y los datos que van con ese tipo. Si viene en talles, colores u otras opciones, las sumás abajo. Se publica solo en tu tienda.'
	],
	[
		/cupon|descuento|codigo de descuento/,
		'En **Marketing → Cupones** creás el código, el descuento (porcentaje o monto) y hasta cuándo vale. Después lo compartís con un link.'
	],
	[
		/envio|envios|correo argentino|retiro/,
		'En **Configuración → Envíos** elegís cómo entregás: retiro en el local, Correo Argentino o tu propio envío, y desde cuánto es gratis.'
	],
	[
		/mercado ?pago|cobrar con tarjeta|tarjeta de credito/,
		'En **Integraciones → Mercado Pago** conectás tu cuenta y tus clientes pagan con tarjeta, débito o dinero en cuenta. La plata va directo a tu Mercado Pago.'
	],
	[
		/transferencia|alias|cbu/,
		'En **Configuración → Datos del negocio** cargás tu alias o CBU: así aparece «Transferencia» al pagar, con el descuento que elijas.'
	],
	[
		/mercado ?libre|publicar en ml/,
		'En **Integraciones → Mercado Libre** conectás tu cuenta y publicás tus productos desde acá, con el stock sincronizado.'
	],
	[
		/dominio|mi propia direccion|\.com\.ar/,
		'En **Cuenta → Dominio** ves tu dirección y podés conectar un dominio propio.'
	],
	[
		/factur|arca|afip/,
		'En **Integraciones → ARCA** cargás tu CUIT y tu certificado; después cada venta se puede facturar con un toque.'
	],
	[
		/plantilla|diseño|diseno|colores de la tienda|cambiar la tienda/,
		'En **Mi tienda** elegís la plantilla, los colores, el logo y qué secciones se ven. Los cambios se publican cuando los guardás.'
	],
	[
		/equipo|empleado|usuario|permiso/,
		'En **Configuración → Equipo** sumás a tu gente con su mail y elegís qué puede ver y hacer cada uno.'
	],
	[
		/importar|traer mis productos|migrar|tiendanube|shopify|woocommerce/,
		'En **Productos → Importar** traés tu catálogo desde Tiendanube, Shopify o WooCommerce, con fotos y variantes.'
	],
	[
		/whatsapp/,
		'WhatsApp Business todavía no se puede conectar al panel: está en camino. Mientras tanto, el botón de WhatsApp de tu tienda abre un chat con el número que cargás en **Configuración → Negocio**.'
	]
];

export function responderSinIa(pregunta: string, d: DatosHornero): string | null {
	const q = normalizar(pregunta);
	if (!q) return null;
	const { contexto, hoy, semana } = d;
	const productos = contexto.productos;
	const dias = contexto.ventasPorDia;
	// Ventas y Local no ven montos: Tincho les cuenta pedidos y unidades, nunca pesos.
	const montos = conMontos(d);
	const sinPlata = ' Tu rol no ve montos de dinero.';

	if (/^(gracias|genial|buenisimo|perfecto|joya|dale)\b/.test(q))
		return '¡De nada! Si necesitás algo más, acá estoy.';
	if (
		/^(hola|buenas|buen dia|buenos dias|buenas tardes|buenas noches|que tal)\b/.test(q) &&
		q.split(' ').length <= 4
	) {
		const nombre = d.hola ? `, ${d.hola}` : '';
		return montos
			? `¡Hola${nombre}! Hoy llevás **${pesos(hoy.ventas?.actual ?? 0)}** en ${plural(hoy.kpis.pedidos.actual, 'pedido', 'pedidos')}. Preguntame por ventas, stock, precios o cómo hacer algo en el panel.`
			: `¡Hola${nombre}! Hoy entraron **${plural(hoy.kpis.pedidos.actual, 'pedido', 'pedidos')}**. Preguntame por pedidos, stock o cómo hacer algo en el panel.`;
	}

	for (const [re, texto] of GUIAS)
		if (/(como|donde|quiero|puedo|ayuda|necesito)/.test(q) && re.test(q)) return texto;

	// Ventas y pedidos.
	if (/(vendi|venta|ventas|facture|gane|recaude|pedido|pedidos)/.test(q)) {
		const soloPedidos = /pedido/.test(q) && !/vend|venta/.test(q);
		const pedidos = soloPedidos || !montos;
		const nota = soloPedidos || montos ? '' : sinPlata;
		if (/ayer/.test(q)) {
			const ayer = dias.at(-2);
			if (!ayer) return 'Todavía no tengo datos de ayer.';
			return pedidos
				? `Ayer entraron **${plural(ayer.pedidos, 'pedido', 'pedidos')}**.${nota}`
				: `Ayer vendiste **${pesos(ayer.ventas ?? 0)}** en ${plural(ayer.pedidos, 'pedido', 'pedidos')}.`;
		}
		if (/semana|7 dias|siete dias/.test(q))
			return pedidos
				? `En los últimos 7 días entraron **${plural(semana.kpis.pedidos.actual, 'pedido', 'pedidos')}**${variacion(semana.kpis.pedidos.actual, semana.kpis.pedidos.anterior)}.${nota}`
				: `En los últimos 7 días vendiste **${pesos(semana.ventas?.actual ?? 0)}**${variacion(semana.ventas?.actual ?? 0, semana.ventas?.anterior ?? 0)}.`;
		if (/(14|catorce) dias|dos semanas|quincena/.test(q)) {
			const cuantos = dias.reduce((n, x) => n + x.pedidos, 0);
			if (pedidos)
				return `En los últimos 14 días entraron **${plural(cuantos, 'pedido', 'pedidos')}**.${nota}`;
			const total = dias.reduce((n, x) => n + (x.ventas ?? 0), 0);
			return `En los últimos 14 días vendiste **${pesos(total)}** en ${plural(cuantos, 'pedido', 'pedidos')}.`;
		}
		if (/mejor dia|dia que mas/.test(q) && dias.length) {
			const valor = (x: (typeof dias)[number]) => (montos ? (x.ventas ?? 0) : x.pedidos);
			const mejor = [...dias].sort((a, b) => valor(b) - valor(a))[0];
			const f = new Date(mejor.fecha + 'T12:00:00').toLocaleDateString('es-AR', {
				weekday: 'long',
				day: 'numeric',
				month: 'long'
			});
			if (!valor(mejor)) return 'En las últimas dos semanas todavía no hubo ventas.';
			return montos
				? `Tu mejor día de las últimas dos semanas fue el **${f}**: ${pesos(valor(mejor))}.`
				: `Tu mejor día de las últimas dos semanas fue el **${f}**: ${plural(mejor.pedidos, 'pedido', 'pedidos')}.`;
		}
		return pedidos
			? `Hoy entraron **${plural(hoy.kpis.pedidos.actual, 'pedido', 'pedidos')}**.${nota}`
			: `Hoy llevás **${pesos(hoy.ventas?.actual ?? 0)}** en ${plural(hoy.kpis.pedidos.actual, 'pedido', 'pedidos')}. Ayer a esta hora: ${pesos(hoy.ventas?.anterior ?? 0)}.`;
	}
	if (/ticket|promedio por compra|gasta cada/.test(q)) {
		const tk = semana.kpis.ticket;
		if (!tk) return `No tengo el ticket promedio para vos.${sinPlata}`;
		return `En los últimos 7 días cada compra fue de **${pesos(tk.actual)}** en promedio${variacion(tk.actual, tk.anterior)}.`;
	}
	if (/conversion|convierte|de cada 100/.test(q) && !semana.kpis.conversion.actual)
		return 'Todavía no mido las visitas de tu tienda, así que no puedo calcular la conversión. Si conectás **Google Analytics** en Integraciones, la ves en tu cuenta de Google.';
	if (/conversion|convierte|de cada 100/.test(q))
		return `De cada 100 visitas, **${semana.kpis.conversion.actual.toLocaleString('es-AR', { maximumFractionDigits: 1 })}** terminan comprando (últimos 7 días).`;

	// Productos: lo que más y lo que menos sale, sin stock, cantidad.
	if (/(mas vendido|mas se vende|mas sale|mas salio|lo que mas|top|estrella)/.test(q)) {
		const top = [...productos]
			.filter((p) => p.vendidos30d > 0)
			.sort((a, b) => b.vendidos30d - a.vendidos30d)
			.slice(0, 3);
		if (!top.length) return 'Todavía no hay ventas en los últimos 30 días para armar el ranking.';
		return `Lo que más salió en 30 días: ${top.map((p, i) => `${i + 1}. **${p.nombre}** (${plural(p.vendidos30d, 'unidad', 'unidades')})`).join(', ')}.`;
	}
	if (/(menos vendido|menos sale|no se vende|no sale|no se mueve|parado|quieto)/.test(q)) {
		const quietos = productos.filter((p) => p.vendidos30d === 0 && p.stock > 0);
		if (!quietos.length)
			return 'Todos tus productos con stock tuvieron alguna venta en los últimos 30 días.';
		return `Sin ventas en 30 días y con stock: ${quietos
			.slice(0, 5)
			.map((p) => `**${p.nombre}** (${num(p.stock)})`)
			.join(
				', '
			)}${quietos.length > 5 ? ` y ${quietos.length - 5} más` : ''}. Una promo o un cupón puede moverlos.`;
	}
	if (
		/(sin stock|agotado|se termino|falta stock|reponer|poco stock|bajo stock|quedan pocos)/.test(q)
	) {
		const sin = productos.filter((p) => p.stock <= 0);
		const pocos = productos.filter((p) => p.stock > 0 && p.stock <= 3);
		if (!sin.length && !pocos.length)
			return 'Está todo con stock: ningún producto tiene 3 unidades o menos.';
		const partes = [];
		if (sin.length)
			partes.push(
				`Sin stock: ${sin
					.slice(0, 6)
					.map((p) => `**${p.nombre}**`)
					.join(', ')}${sin.length > 6 ? ` y ${sin.length - 6} más` : ''}.`
			);
		if (pocos.length)
			partes.push(
				`Quedan pocos: ${pocos
					.slice(0, 6)
					.map((p) => `**${p.nombre}** (${num(p.stock)})`)
					.join(', ')}.`
			);
		return partes.join(' ');
	}
	if (/cuantos productos|cantidad de productos|productos tengo|productos cargados/.test(q))
		return productos.length
			? `Tenés **${plural(productos.length, 'producto', 'productos')}** cargados${(() => {
					const cats = new Set(productos.map((p) => p.categoria).filter(Boolean));
					return cats.size ? `, en ${plural(cats.size, 'categoría', 'categorías')}` : '';
				})()}.`
			: 'Todavía no cargaste productos. Andá a **Productos → Nuevo producto** y en un minuto tenés el primero.';
	if (/mercado ?libre|\bml\b/.test(q)) {
		const enMl = productos.filter((p) => p.enMl).length;
		return enMl
			? `Tenés **${plural(enMl, 'producto publicado', 'productos publicados')}** en Mercado Libre de ${num(productos.length)}.`
			: 'Todavía no publicaste productos en Mercado Libre. Lo conectás en **Integraciones → Mercado Libre**.';
	}

	// Un producto en particular: stock y precio.
	if (/(stock|cuantos quedan|cuantas quedan|quedan|precio|cuesta|sale|vale|tengo)/.test(q)) {
		const p = productoDe(pregunta, productos);
		if (p) {
			const stock = p.stock > 0 ? `${plural(p.stock, 'unidad', 'unidades')} en stock` : 'sin stock';
			if (/precio|cuesta|sale|vale/.test(q))
				return `**${p.nombre}** sale **${pesos(p.precio)}** (${stock}).`;
			return `De **${p.nombre}** tenés **${stock}**. Vendiste ${plural(p.vendidos30d, 'unidad', 'unidades')} en 30 días.`;
		}
	}

	// Clientes y canales.
	if (/cliente/.test(q)) {
		const c = contexto.clientes;
		if (/vuelv|perdi|no compran|recuperar/.test(q))
			return c.sinVolver
				? `**${plural(c.sinVolver, 'cliente compró', 'clientes compraron')}** más de una vez y hace un mes que no vuelven${c.conWhatsApp ? '; a los que dejaron WhatsApp les podés escribir desde el panel' : ''}.`
				: 'No tenés clientes frecuentes que hayan dejado de comprar.';
		if (/nuevo/.test(q))
			return `En los últimos 30 días se sumaron **${plural(c.nuevos30d, 'cliente nuevo', 'clientes nuevos')}**.`;
		return `Tenés **${plural(c.total, 'cliente', 'clientes')}**: ${num(c.nuevos30d)} nuevos en 30 días y ${num(c.conWhatsApp)} con WhatsApp.`;
	}
	if (/canal|de donde|por donde|donde vendo|donde compran/.test(q)) {
		const NOMBRES: Record<string, string> = {
			web: 'la tienda online',
			wa: 'WhatsApp',
			ml: 'Mercado Libre',
			ig: 'Instagram',
			pos: 'el local',
			tg: 'Telegram'
		};
		const valor = (x: (typeof contexto.canales30d)[number]) =>
			montos ? (x.ventas ?? 0) : x.pedidos;
		const c = [...contexto.canales30d]
			.filter((x) => valor(x) > 0)
			.sort((a, b) => valor(b) - valor(a));
		if (!c.length) return 'Todavía no hay ventas en los últimos 30 días para ver de dónde vienen.';
		return `En 30 días vendiste más por ${c
			.slice(0, 3)
			.map(
				(x) =>
					`**${NOMBRES[x.canal] ?? x.canal}** (${montos ? pesos(valor(x)) : plural(x.pedidos, 'pedido', 'pedidos')})`
			)
			.join(', ')}.`;
	}

	return null;
}

/** Sin IA y sin entender la pregunta: lo que sí sé responder. */
export const ayuda = () =>
	'Esa no la sé responder todavía. Preguntame, por ejemplo: «¿cuánto vendí esta semana?», «¿qué es lo que más sale?», «¿cuánto stock queda de…?», «¿qué productos están sin stock?» o «¿cómo cargo un cupón?».';
