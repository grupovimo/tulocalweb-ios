/**
 * El motor de Tincho sin IA: entiende las preguntas de todos los días por conceptos (con sinónimos y
 * errores de tipeo), el período del que hablan y el producto que nombran, sigue la charla («¿y
 * ayer?») y responde con los datos reales: texto, cifras, gráficos, tablas, botones y preguntas para
 * seguir. Lo que no entiende devuelve null y recién ahí responde el modelo (si lo hay).
 */
import type { DatosHornero } from '../contexto.ts';
import { guiaDe } from './guias.ts';
import * as I from './intenciones.ts';
import { periodoDe } from './periodo.ts';
import type { Respuesta } from './respuesta.ts';
import { alguno, entender, tiene, type Pregunta } from './texto.ts';

export { normalizar } from './texto.ts';
export { productoDe } from './intenciones.ts';
export type { Respuesta, Bloque, Accion, Destino } from './respuesta.ts';

/** ¿Pide cómo hacer algo en el panel? */
const PIDE_COMO =
	/(como (hago|puedo|se|cargo|creo|agrego|subo|pongo|cambio|conecto|activo|configuro|comparto|publico|traigo|importo|invito|facturo|cobro|vendo en|armo|saco|borro)|donde (se|esta|cambio|veo|cargo|pongo|configuro|activo|creo)|quiero (cargar|crear|agregar|subir|bajar|aumentar|poner|cambiar|conectar|activar|publicar|traer|importar|invitar|facturar|cobrar|compartir|vender|hacer|armar|sacar)|se puede|hay forma|necesito (cargar|crear|cambiar|conectar|hacer|subir|bajar|aumentar|poner|sacar)|ayuda con|tutorial|paso a paso|me ensenas|explicame)/;

const NO_SE_VENDE =
	/no (se )?(vende|venden|sale|salen|mueve|mueven|rota|rotan)|sin ventas|nadie compra/;

function intencion(x: I.Ctx): Respuesta | null {
	const { p } = x;
	const q = p.q;
	const como = PIDE_COMO.test(q);

	const guia = guiaDe(q, como);
	if (guia) return guia;

	if (alguno(p, 'idea') || /vender mas|como mejoro|que me recomendas|que hago/.test(q))
		return I.ideas(x);
	if (alguno(p, 'meta', 'proyeccion') && !tiene(p, 'producto')) return I.meta(x);
	if (tiene(p, 'comparar') && !x.producto)
		return I.ventas({ ...x, periodo: x.periodo ?? { tipo: 'semana' } });
	if (
		tiene(p, 'resumen') &&
		!alguno(p, 'stock', 'cliente', 'canal', 'visita', 'vender', 'plata', 'pedido')
	)
		return I.resumen(x);
	if (
		tiene(p, 'vender') &&
		x.periodo &&
		!alguno(p, 'producto', 'mas', 'menos', 'canal', 'cliente') &&
		!x.producto
	)
		return I.ventas(x);
	if (
		(alguno(p, 'pendiente', 'preparar') || (tiene(p, 'transferencia') && !como)) &&
		!alguno(p, 'reponer', 'agotado', 'poco')
	)
		return I.pendientes(x);
	if (tiene(p, 'carrito')) return I.carritos(x);
	if (tiene(p, 'resena')) return I.resenas(x);
	if (tiene(p, 'caja')) return I.caja(x);
	if (tiene(p, 'envio') && !como) return I.envios(x);
	if (alguno(p, 'visita', 'conversion')) return I.visitas(x);
	if (
		tiene(p, 'ticket') ||
		(tiene(p, 'promedio') && alguno(p, 'vender', 'plata', 'pedido') && !tiene(p, 'dia'))
	)
		return I.ticket(x);
	if (tiene(p, 'mercadolibre')) return I.mercadoLibre(x);
	if ((tiene(p, 'cliente') && !tiene(p, 'canal')) || (tiene(p, 'volver') && !x.producto))
		return I.clientes(x);
	if (tiene(p, 'canal') && alguno(p, 'vender', 'pedido', 'mas', 'cliente')) return I.canales(x);
	if (alguno(p, 'mejor', 'peor') && alguno(p, 'dia', 'vender', 'pedido') && !tiene(p, 'producto'))
		return I.mejorDia(x, tiene(p, 'peor'));
	if (tiene(p, 'ultimo') && tiene(p, 'pedido')) return I.ultimosPedidos(x);
	if (alguno(p, 'caro', 'barato')) return I.extremosPrecio(x, tiene(p, 'caro'));
	if (NO_SE_VENDE.test(q) || (tiene(p, 'menos') && alguno(p, 'vender', 'producto', 'pedido')))
		return I.quietos(x);
	if ((tiene(p, 'mas') || tiene(p, 'mejor')) && alguno(p, 'vender', 'producto', 'pedido'))
		return I.masVendidos(x);
	if (x.producto && alguno(p, 'stock', 'precio', 'cuanto', 'vender', 'pedido'))
		return I.unProducto(x);
	if (tiene(p, 'valor') && tiene(p, 'stock')) return I.catalogo(x);
	if (alguno(p, 'agotado', 'poco', 'reponer') || (tiene(p, 'stock') && !x.producto))
		return I.stockBajo(x);
	if (tiene(p, 'categoria') || (alguno(p, 'cuanto', 'valor') && tiene(p, 'producto')))
		return I.catalogo(x);
	if (alguno(p, 'vender', 'plata', 'pedido')) return I.ventas(x);
	if (x.producto && p.palabras.length <= 6) return I.unProducto(x);
	return null;
}

function contexto(p: Pregunta, d: DatosHornero, ahora: Date): I.Ctx {
	return {
		p,
		d,
		periodo: periodoDe(p, ahora),
		montos: I.montosDe(d),
		producto: I.productoDe(p.original, d.contexto.productos),
		productos: I.productosDe(p.original, d.contexto.productos)
	};
}

const SUGERENCIAS = ['¿Cómo vengo hoy?', '¿Qué tengo que reponer?', 'Ideas para vender más'];

/**
 * Responde con los datos. `anteriores`: las preguntas previas de la persona (la última al final),
 * para entender los seguimientos («¿y ayer?», «¿y del mate?»).
 */
export function responder(
	pregunta: string,
	d: DatosHornero,
	anteriores: string[] = [],
	ahora = new Date()
): Respuesta | null {
	const p = entender(pregunta);
	if (!p.q) return null;
	const corta = p.palabras.length <= 4;

	if (corta && tiene(p, 'gracias') && p.conceptos.size <= 2)
		return { texto: '¡De nada! Si necesitás algo más, acá estoy.', sugerencias: SUGERENCIAS };
	if (/^(chau|nos vemos|hasta luego|adios)\b/.test(p.q))
		return { texto: '¡Hasta luego! Que se venda mucho. 🙌' };
	if (
		/(quien sos|que (sos|podes|sabes|haces)|para que servis|que te puedo preguntar|en que me ayudas)/.test(
			p.q
		)
	)
		return {
			texto:
				'Soy **Tincho**, conozco tu tienda al día: ventas, pedidos, stock, clientes, visitas y cómo hacer cualquier cosa en el panel. Preguntame como le preguntarías a alguien del equipo.',
			sugerencias: [
				'¿Cómo vengo hoy?',
				'¿Qué es lo que más se vende?',
				'¿Qué tengo pendiente?',
				'¿Cómo creo un cupón?'
			]
		};

	const x = contexto(p, d, ahora);
	if (corta && tiene(p, 'saludo') && p.conceptos.size <= 2) return I.resumen(x);

	let r = intencion(x);
	// Seguimiento: «¿y ayer?», «¿y la semana pasada?», «¿y del mate?» → la pregunta anterior con lo nuevo.
	const previa = anteriores.at(-1);
	const sigue = /^(y|e|tambien|ahora)\b/.test(p.q) || (corta && !r);
	if (previa && sigue && (x.periodo || x.producto || !r)) {
		const pp = entender(previa);
		const unida = contexto(entender(`${previa} ${pregunta}`), d, ahora);
		// Lo nuevo manda: el período y el producto de esta pregunta (si los nombra).
		unida.periodo = x.periodo ?? periodoDe(pp, ahora);
		unida.producto = x.producto ?? I.productoDe(previa, d.contexto.productos);
		unida.productos = x.producto ? x.productos : I.productosDe(previa, d.contexto.productos);
		// El aviso de «no encontré…» mira sólo lo que se preguntó ahora.
		if (x.producto) unida.p = { ...unida.p, original: pregunta };
		if (x.periodo) unida.p = { ...pp, original: `${previa} ${pregunta}`, q: `${pp.q} ${p.q}` };
		r = intencion(unida) ?? r;
	}
	if (!r) return null;
	return { ...r, sugerencias: r.sugerencias?.length ? r.sugerencias : SUGERENCIAS };
}

/** Sólo el texto (Telegram y los clientes que no dibujan bloques). */
export function responderSinIa(
	pregunta: string,
	d: DatosHornero,
	anteriores: string[] = []
): string | null {
	return responder(pregunta, d, anteriores)?.texto ?? null;
}

/** Sin IA y sin entender la pregunta: lo que sí sé responder. */
export const ayuda = () =>
	'Esa no la sé responder todavía. Preguntame, por ejemplo: «¿cuánto vendí esta semana?», «¿qué es lo que más sale?», «¿cuánto stock queda de…?», «¿qué tengo pendiente?» o «¿cómo creo un cupón?».';

export const ayudaConBotones = (): Respuesta => ({
	texto:
		'Esa todavía no la sé responder. Sé de ventas, pedidos, stock y precios, clientes, visitas y de cómo hacer cualquier cosa en el panel.',
	sugerencias: [
		'¿Cuánto vendí esta semana?',
		'¿Qué es lo que más se vende?',
		'¿Qué tengo pendiente?',
		'¿Cómo creo un cupón?'
	]
});
