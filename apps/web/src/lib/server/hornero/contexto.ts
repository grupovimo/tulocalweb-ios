/**
 * Lo que Tincho sabe para responder: la tienda, el tablero de hoy y de la semana, ventas por día,
 * productos, clientes y canales. Va en el mensaje de sistema como JSON compacto (no inventa nada que
 * no esté acá).
 */
import type { ContextoHornero, Tablero, Tienda } from '../datos/tipos.ts';

const TZ = 'America/Argentina/Buenos_Aires';

export type DatosHornero = {
	tienda: Pick<Tienda, 'nombre' | 'rubro'>;
	hola: string | null;
	hoy: Tablero;
	semana: Tablero;
	/** Los últimos 30 días (para «¿cuánto vendí este mes?»); opcional para Telegram y la demo. */
	mes?: Tablero;
	contexto: ContextoHornero;
};

export function datosParaElModelo(d: DatosHornero) {
	const { hoy, semana, contexto } = d;
	// Las visitas las mide Tu Local Web (sin cookies) desde el primer día con alguna; antes, al
	// modelo no se le pasa un 0 % que después repetiría como dato. En la demo, la conversión de
	// ejemplo. Con la API real no hay WhatsApp para mandar mensajes: tampoco va su costo.
	const midoVisitas = !!semana.visitas?.desde || semana.kpis.conversion.actual > 0;
	const demo = !semana.visitas && semana.kpis.conversion.actual > 0;
	// Sin montos (Ventas, Local) no van las ventas en pesos, el ticket, la meta ni lo de los carritos:
	// la API ya no los manda y Tincho tampoco los tiene.
	const montos = conMontos(d);
	return {
		...(montos ? {} : { montos: 'quien pregunta no ve montos de dinero' }),
		// Hoy hasta ahora y ayer completo; el detalle de cada día va en ventasPorDia.
		hoyHastaAhora: { ventas: hoy.ventas?.actual, pedidos: hoy.kpis.pedidos.actual },
		ayerDiaCompleto: contexto.ventasPorDia.at(-2) ?? null,
		ultimos7Dias: {
			ventas: semana.ventas?.actual,
			semanaAnterior: semana.ventas?.anterior,
			pedidos: semana.kpis.pedidos.actual,
			ticketPromedio: semana.kpis.ticket?.actual,
			conversionPorcentaje: midoVisitas ? semana.kpis.conversion.actual : 'no se mide todavía',
			clientesNuevos: semana.kpis.clientesNuevos.actual,
			masVendidos: semana.masVendidos.map((p) => `${p.nombre}: ${p.unidades} u`)
		},
		metaDelMes: semana.meta,
		ventasPorDia: contexto.ventasPorDia,
		productos: contexto.productos,
		clientes: contexto.clientes,
		ventasPorCanal30Dias: contexto.canales30d,
		pendientes: semana.paraHacer,
		carritosAbandonados: { cantidad: semana.hornero.carritos, monto: semana.hornero.carritosMonto },
		// Una visita = una persona distinta en un día (sin cookies: no se sabe si volvió otro día).
		visitasTiendaOnline: semana.visitas?.desde
			? {
					hoy: hoy.visitas?.hoy.actual ?? semana.visitas.hoy.actual,
					ayer: semana.visitas.hoy.anterior,
					ultimos7Dias: semana.visitas.dias7.actual,
					semanaAnterior: semana.visitas.dias7.anterior,
					paginasVistas7Dias: semana.visitas.vistas.actual,
					masVistos7Dias: semana.visitas.masVistos.map((p) => `${p.nombre}: ${p.vistas} vistas`),
					...(contexto.visitas ?? {})
				}
			: demo
				? 'en la demo sólo hay la conversión de ejemplo'
				: 'todavía no hubo visitas medidas',
		...(demo ? { costoMensajeWhatsApp: semana.hornero.costoMensaje } : {}),
		resenas: { promedio: semana.resenas.promedio, nuevas: semana.resenas.nuevas }
	};
}

/** ¿Quien pregunta ve montos? (Ventas y Local no: la API los omite y `montos` viene en false). */
export const conMontos = (d: DatosHornero) => d.hoy.montos && d.semana.montos;

export function sistema(d: DatosHornero, ahora = new Date()) {
	const cuando = new Intl.DateTimeFormat('es-AR', {
		timeZone: TZ,
		weekday: 'long',
		day: 'numeric',
		month: 'long',
		hour: '2-digit',
		minute: '2-digit'
	}).format(ahora);
	return [
		`Sos Tincho, el asistente de la tienda «${d.tienda.nombre}» (${d.tienda.rubro}) en Tu Local Web, una plataforma argentina de venta online.`,
		d.hola ? `Hablás con ${d.hola}, la dueña.` : 'Hablás con quien maneja la tienda.',
		`Ahora es ${cuando} (hora de Argentina).`,
		'Hablá en español rioplatense, cálido y concreto, como alguien del equipo. Respondé en 1 a 4 oraciones (máximo 90 palabras), sin títulos ni listas largas.',
		'Marcá con **negrita** sólo el dato clave. Escribí la plata así: $ 12.345.',
		...(conMontos(d)
			? []
			: [
					'Quien pregunta no ve montos de dinero (su rol no tiene finanzas): nunca des montos en pesos. Si te pregunta por plata, decí que su rol no ve montos y ofrecé pedidos, unidades o stock.'
				]),
		'Usá sólo los DATOS de abajo. Si algo no está, decí que todavía no lo tenés; nunca inventes números, nombres ni fechas.',
		'Si te piden hacer algo (subir precios, mandar una campaña, pausar productos), contá cómo quedaría con los datos y en qué sección del panel se hace. Desde el chat no podés hacerlo ni dejarlo preparado: nunca digas que lo hiciste ni que quedó listo.',
		'Si te preguntan algo que no tiene que ver con la tienda, respondé con amabilidad que sólo ayudás con la tienda.',
		`DATOS: ${JSON.stringify(datosParaElModelo(d))}`
	].join('\n');
}
