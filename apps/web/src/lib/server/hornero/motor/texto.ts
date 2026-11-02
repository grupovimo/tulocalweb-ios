/**
 * Cómo entiende Tincho lo que le escriben: sin tildes ni signos, palabra por palabra, con sinónimos
 * del habla argentina y tolerancia a errores de tipeo («bendi», «stok», «facture»). Cada palabra se
 * lleva a un concepto (vender, stock, cliente…) y las intenciones se arman con conceptos, no con
 * frases exactas.
 */

/** Minúsculas, sin tildes ni signos: "¿Cuánto VENDÍ?" → "cuanto vendi". */
export const normalizar = (t: string) =>
	t
		.toLowerCase()
		.normalize('NFD')
		.replace(/[̀-ͯ]/g, '')
		.replace(/[^a-z0-9%$/ ]+/g, ' ')
		.replace(/\s+/g, ' ')
		.trim();

/** Distancia de edición (con transposición: «stokc» está a 1 de «stock»). */
export function distancia(a: string, b: string): number {
	if (a === b) return 0;
	if (Math.abs(a.length - b.length) > 2) return 3;
	const d: number[][] = Array.from({ length: a.length + 1 }, (_, i) =>
		Array.from({ length: b.length + 1 }, (_, j) => (i === 0 ? j : j === 0 ? i : 0))
	);
	for (let i = 1; i <= a.length; i++)
		for (let j = 1; j <= b.length; j++) {
			const c = a[i - 1] === b[j - 1] ? 0 : 1;
			d[i][j] = Math.min(d[i - 1][j] + 1, d[i][j - 1] + 1, d[i - 1][j - 1] + c);
			if (i > 1 && j > 1 && a[i - 1] === b[j - 2] && a[i - 2] === b[j - 1])
				d[i][j] = Math.min(d[i][j], d[i - 2][j - 2] + 1);
		}
	return d[a.length][b.length];
}

/** ¿La palabra es esta (o casi)? Las cortas tienen que ser exactas; las largas aceptan errores. */
export function parecida(palabra: string, objetivo: string): boolean {
	if (palabra === objetivo) return true;
	// Las cortas se confunden fácil («tengo» y «vengo»): desde 6 letras y con la misma inicial.
	if (objetivo.length < 6 || palabra.length < 5 || palabra[0] !== objetivo[0]) return false;
	return distancia(palabra, objetivo) <= (objetivo.length >= 9 ? 2 : 1);
}

export type Concepto =
	| 'vender'
	| 'plata'
	| 'pedido'
	| 'stock'
	| 'agotado'
	| 'poco'
	| 'reponer'
	| 'precio'
	| 'producto'
	| 'categoria'
	| 'cliente'
	| 'canal'
	| 'ticket'
	| 'conversion'
	| 'visita'
	| 'carrito'
	| 'resena'
	| 'meta'
	| 'mejor'
	| 'peor'
	| 'mas'
	| 'menos'
	| 'cuanto'
	| 'como'
	| 'resumen'
	| 'idea'
	| 'preparar'
	| 'transferencia'
	| 'envio'
	| 'caja'
	| 'mercadolibre'
	| 'ultimo'
	| 'dia'
	| 'promedio'
	| 'proyeccion'
	| 'caro'
	| 'barato'
	| 'valor'
	| 'nuevo'
	| 'volver'
	| 'whatsapp'
	| 'saludo'
	| 'gracias'
	| 'quien'
	| 'pendiente'
	| 'chat'
	| 'comparar';

/**
 * Las raíces de cada concepto. Una palabra cae en un concepto si empieza con la raíz (para las
 * conjugaciones: vend → vendí, vendimos, vendiendo) o si se le parece (errores de tipeo).
 */
const RAICES: [Concepto, string[]][] = [
	[
		'vender',
		[
			'vend',
			'bend',
			'factur',
			'recaud',
			'gane',
			'ganamos',
			'ganancia',
			'hice',
			'hicimos',
			'movimiento',
			'facturacion',
			'ingres',
			'cobr',
			'venta',
			'ventas'
		]
	],
	['plata', ['plata', 'guita', 'dinero', 'mango', 'pesos', 'monto', 'recaudacion', 'total']],
	['pedido', ['pedid', 'orden', 'ordenes', 'compra', 'compras']],
	['stock', ['stock', 'stok', 'inventario', 'existencia', 'unidades', 'quedan', 'queda', 'hay']],
	['agotado', ['agotad', 'sinstock', 'termino', 'terminaron', 'acabo', 'acabaron', 'cero']],
	['poco', ['poco', 'pocos', 'bajo', 'bajos', 'escaso', 'casi']],
	['reponer', ['repon', 'reposicion', 'comprar', 'encargar', 'faltan', 'falta', 'faltante']],
	['precio', ['precio', 'cuesta', 'vale', 'valen', 'cobro', 'sale', 'salen']],
	['producto', ['product', 'articul', 'item', 'cosas', 'catalogo', 'mercaderia']],
	['categoria', ['categor', 'rubros', 'seccion', 'tipos']],
	['cliente', ['client', 'compradores', 'comprador', 'gente', 'personas']],
	['canal', ['canal', 'canales', 'donde', 'medio', 'medios', 'origen']],
	['ticket', ['ticket', 'tiket', 'gastan', 'gasta']],
	['conversion', ['conversion', 'convierte', 'convierten', 'convertir']],
	['visita', ['visit', 'entraron', 'entran', 'entro', 'trafico', 'vieron', 'miraron', 'vistas']],
	['carrito', ['carrito', 'abandon']],
	['resena', ['resena', 'opinion', 'comentario', 'calificacion', 'puntaje', 'review']],
	['meta', ['meta', 'objetivo']],
	['mejor', ['mejor', 'record', 'pico']],
	['peor', ['peor', 'flojo', 'malo']],
	['mas', ['mas', 'top', 'estrella', 'ranking']],
	['menos', ['menos', 'parado', 'parados', 'quieto', 'quietos', 'clavado', 'clavados']],
	['cuanto', ['cuanto', 'cuantos', 'cuanta', 'cuantas', 'cantidad', 'numero']],
	[
		'como',
		[
			'como',
			'donde',
			'puedo',
			'hago',
			'hacer',
			'quiero',
			'necesito',
			'ayuda',
			'configur',
			'activ',
			'conect',
			'cambi',
			'agreg',
			'carg',
			'sub',
			'cre',
			'pon',
			'saco',
			'sacar',
			'borr',
			'elimin'
		]
	],
	[
		'resumen',
		[
			'resumen',
			'vengo',
			'venimos',
			'novedad',
			'novedades',
			'balance',
			'onda',
			'panorama',
			'situacion',
			'estado'
		]
	],
	['idea', ['idea', 'consejo', 'recomend', 'suger', 'tip', 'mejorar', 'aumentar', 'estrategia']],
	['preparar', ['prepar', 'armar', 'despach', 'empaquet']],
	['transferencia', ['transferen', 'confirm']],
	['envio', ['envio', 'envios', 'enviar', 'paquete', 'paquetes', 'retiro', 'correo', 'entrega']],
	['caja', ['caja', 'efectivo', 'arqueo']],
	['mercadolibre', ['mercadolibre', 'meli', 'ml']],
	['ultimo', ['ultim', 'reciente']],
	['dia', ['dia', 'dias', 'jornada']],
	['promedio', ['promedio', 'media']],
	['proyeccion', ['proyecc', 'cierro', 'cerrar', 'llego', 'llegar', 'llegamos']],
	['caro', ['caro', 'caros', 'carisimo']],
	['barato', ['barato', 'baratos', 'economico']],
	['valor', ['valor', 'valuado', 'vale', 'capital', 'invertido']],
	['nuevo', ['nuevo', 'nuevos', 'nueva', 'nuevas']],
	['volver', ['vuelv', 'volv', 'perdi', 'recuper', 'dejaron', 'inactiv']],
	['whatsapp', ['whatsapp', 'wsp', 'wpp', 'whats']],
	['saludo', ['hola', 'buenas', 'buen', 'buenos', 'holis', 'hey', 'que tal']],
	[
		'gracias',
		[
			'gracias',
			'genial',
			'buenisimo',
			'perfecto',
			'joya',
			'dale',
			'barbaro',
			'ok',
			'okey',
			'listo',
			'excelente',
			'groso',
			'capo',
			'crack'
		]
	],
	['quien', ['quien', 'sos', 'podes', 'sabes']],
	['pendiente', ['pendient', 'tarea', 'tareas', 'atrasad']],
	['chat', ['chat', 'mensaje', 'consulta', 'pregunta', 'responder']],
	['comparar', ['compara', 'versus', 'vs', 'contra', 'diferencia']]
];

/** Palabras que no dicen nada para buscar productos. */
export const VACIAS = new Set(
	'el la los las de del un una unos unas y o a en por para con que cual cuales tengo tenes hay me mi mis tu tus se lo le es son esta estan este esto eso al como cuanto cuantos cuanta cuantas queda quedan stock precio sale salen cuesta vale tiene tienen hoy ayer semana mes dias dia vendi vendo vendimos'.split(
		' '
	)
);

export type Pregunta = {
	original: string;
	q: string;
	palabras: string[];
	conceptos: Set<Concepto>;
};

/** Lleva la pregunta a conceptos. */
export function entender(texto: string): Pregunta {
	const q = normalizar(texto);
	const palabras = q.split(' ').filter(Boolean);
	const conceptos = new Set<Concepto>();
	for (const p of palabras)
		for (const [c, raices] of RAICES)
			if (raices.some((r) => (r.length >= 4 ? p.startsWith(r) || parecida(p, r) : p === r)))
				conceptos.add(c);
	// Frases de varias palabras.
	if (/mercado ?libre/.test(q)) conceptos.add('mercadolibre');
	if (/sin stock|no hay mas|no queda/.test(q)) conceptos.add('agotado');
	if (/que tal|como va|como andamos|como estamos|como viene/.test(q)) conceptos.add('resumen');
	if (/que (hago|puedo hacer) para|como (vendo|hago para vender) mas/.test(q))
		conceptos.add('idea');
	if (/a confirmar|por confirmar|sin confirmar/.test(q)) conceptos.add('transferencia');
	if (/\bestrellas\b/.test(q)) conceptos.add('resena');
	if (
		/(me|nos) (fue|va|anduvo|esta yendo)|como (anduvo|estuvo|salio)|que tal (fue|estuvo|anduvo)/.test(
			q
		)
	)
		conceptos.add('vender');
	return { original: texto, q, palabras, conceptos };
}

/** ¿Tiene todos estos conceptos? */
export const tiene = (p: Pregunta, ...cs: Concepto[]) => cs.every((c) => p.conceptos.has(c));
/** ¿Tiene alguno? */
export const alguno = (p: Pregunta, ...cs: Concepto[]) => cs.some((c) => p.conceptos.has(c));
