import { PLANES } from '#lib/panel/planes.ts';
import { fmt } from '#lib/landing/formato.ts';

/**
 * Las páginas de aterrizaje de los anuncios de Meta (/l/<id>): una por ángulo, que sigue lo que dijo
 * el anuncio y lleva a una sola acción, crear la cuenta. Cada promesa sale de lo que el producto ya
 * hace (marketing/tulocalweb/competencia.md › «Ángulos libres»); nada de «más ventas» ni «sin costos
 * de cobro».
 */
export type Escena = 'avisos' | 'caja' | 'pasate';

export type Angulo = {
	id: string;
	escena: Escena;
	/** Lo que se ve al compartir el link y en la pestaña. */
	titulo: string;
	descripcion: string;
	etiqueta: string;
	/** El título en dos renglones: el segundo va en terracota. */
	h1: [string, string];
	bajada: string;
	pruebas: { t: string; d: string }[];
	/** La aclaración del precio o del plan, cerca del botón. */
	nota: string;
	preguntas: { p: string; r: string }[];
	cierre: string;
};

const pro = PLANES.find((p) => p.id === 'pro')!;

const COMISION = {
	p: '¿Cobran comisión por venta?',
	r: 'No: 0 % por venta en todos los planes. Mercado Pago cobra lo suyo, como siempre.'
};
const PRUEBA = '14 días gratis, sin tarjeta. Si no te sirve, no pagás nada.';

export const ANGULOS: Angulo[] = [
	{
		id: 'avisos',
		escena: 'avisos',
		titulo: 'Te enterás de cada venta · Tu Local Web',
		descripcion:
			'Cada venta de tu tienda online, WhatsApp, Mercado Libre y el local te avisa al instante.',
		etiqueta: 'Avisos de venta',
		h1: ['Te enterás de cada venta.', 'Al instante, estés donde estés.'],
		bajada:
			'La tienda online, WhatsApp, Mercado Libre y el local avisan en el mismo lugar: al celular, al mail o por Telegram.',
		pruebas: [
			{
				t: 'Un aviso por venta',
				d: 'Con el producto, el total y cómo pagó. Lo tocás y estás en el pedido.'
			},
			{
				t: 'Todos tus canales juntos',
				d: 'Tienda online, WhatsApp, Instagram, Mercado Libre y el local, en una sola lista.'
			},
			{
				t: 'Se atiende desde el celular',
				d: 'Marcás listo, cobrado o despachado con un toque, sin prender la compu.'
			}
		],
		nota: PRUEBA,
		preguntas: [
			{
				p: '¿Tengo que instalar algo?',
				r: 'No. Entrás desde el navegador y elegís dónde te llegan los avisos: el celular, el mail o Telegram.'
			},
			COMISION,
			{
				p: '¿Y si ya tengo tienda en otra plataforma?',
				r: 'Te la traemos gratis: productos con fotos y variantes, y desde Tiendanube también clientes y pedidos.'
			}
		],
		cierre: 'Que la próxima venta no te la cuente nadie.'
	},
	{
		id: 'una-caja',
		escena: 'caja',
		titulo: 'El local y la web, con un solo stock · Tu Local Web',
		descripcion:
			'Vendés en el mostrador y se descuenta en la tienda online y en Mercado Libre. Todo en un solo panel.',
		etiqueta: 'Local + tienda online',
		h1: ['El local y la web,', 'con un solo stock.'],
		bajada:
			'Vendés en el mostrador y se descuenta en la tienda online y en Mercado Libre. Nunca más vender lo que ya no tenés.',
		pruebas: [
			{
				t: 'La caja del local',
				d: 'Cobrás con QR, tarjeta, efectivo o transferencia, y dividís el pago si hace falta.'
			},
			{
				t: 'Un solo número de stock',
				d: 'Cada venta, venga de donde venga, descuenta del mismo lugar.'
			},
			{
				t: 'Mercado Libre al día',
				d: 'Precio y stock se sincronizan solos, y lo agotado se pausa.'
			}
		],
		nota: `Local y Mercado Libre vienen en el plan ${pro.nombre}: ${fmt(pro.precio)} por mes. ${PRUEBA}`,
		preguntas: [
			{
				p: '¿Necesito una compu en el local?',
				r: 'No. La caja anda en el celular o en la tablet que ya tenés.'
			},
			COMISION,
			{
				p: '¿Puedo cargar lo que ya tengo?',
				r: 'Sí: importás tus productos desde una planilla o te los traemos de tu tienda actual, gratis.'
			}
		],
		cierre: 'Una caja para todo lo que vendés.'
	},
	{
		id: 'pasate',
		escena: 'pasate',
		titulo: 'Te traemos tu tienda gratis · Tu Local Web',
		descripcion:
			'Productos, fotos y variantes, y desde Tiendanube también clientes y pedidos. Sin comisión por venta.',
		etiqueta: 'Migración gratis',
		h1: ['Te traemos tu tienda.', 'Gratis, en\u00a0una\u00a0tarde.'],
		bajada:
			'Desde Tiendanube, Shopify, WooCommerce o Mercado Libre. Tus links viejos siguen llevando a tus productos.',
		pruebas: [
			{
				t: 'Todo lo que tenés',
				d: 'Productos con fotos y variantes, y desde Tiendanube también clientes y pedidos.'
			},
			{
				t: 'Sin perder Google',
				d: 'Las direcciones viejas redirigen a las nuevas y tu dominio sigue siendo tuyo.'
			},
			{
				t: '0 % por venta',
				d: 'Pagás el plan y nada más. Ninguna venta te descuenta un porcentaje.'
			}
		],
		nota: PRUEBA,
		preguntas: [
			{
				p: '¿Tengo que cerrar mi tienda de ahora?',
				r: 'No. Las dos conviven hasta que decidas pasar el dominio.'
			},
			COMISION,
			{
				p: '¿Cuánto tarda?',
				r: 'La importación corre sola y la ves avanzar. Una tienda común queda lista en una tarde.'
			}
		],
		cierre: 'Mudate sin cajas ni cinta.'
	}
];

export const angulo = (id: string) => ANGULOS.find((a) => a.id === id) ?? null;
