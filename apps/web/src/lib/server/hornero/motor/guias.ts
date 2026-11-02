/**
 * Los «¿cómo hago…?»: pasos cortos, la sección donde se hace (con el nombre que tiene en la app y en
 * la web) y un botón que lleva ahí.
 */
import type { Destino, Respuesta } from './respuesta.ts';

type Guia = { re: RegExp; texto: string; ir?: Destino; boton?: string; sugerencias?: string[] };

const GUIAS: Guia[] = [
	{
		re: /(cargo|cargar|agrego|agregar|subo|subir|crear|creo|nuevo|sumar|sumo|publicar|publico) (un |el |mis |los )?(producto|articulo)/,
		texto:
			'Tocá **Nuevo producto** (el «+»): primero elegís qué es (ropa, comida, hogar…) y después cargás nombre, precio, stock y fotos, con los datos que van con ese tipo. Si viene en talles, colores u otras opciones, las sumás abajo. Se publica solo en tu tienda.',
		ir: 'nuevo-producto',
		boton: 'Cargar un producto',
		sugerencias: ['¿Cómo cargo talles y colores?', '¿Cómo traigo mis productos de otra tienda?']
	},
	{
		re: /plantilla|diseno|colores? de (la|mi) tienda|cambiar la tienda|portada|logo|como se ve|colores/,
		texto:
			'En **Editar mi tienda** elegís la plantilla, los colores, la portada, el logo y qué secciones se ven. Publicás cuando te gusta cómo quedó.',
		ir: 'mi-tienda',
		boton: 'Editar mi tienda'
	},
	{
		re: /talle|variante|combinacion|colores? (de|del|en|a) (un |el |los |mis )?producto/,
		texto:
			'En el producto, en **Variantes y stock**, sumá las opciones (talle, color, sabor…) con sus valores: se arman las combinaciones y cargás el stock y, si querés, el precio de cada una.',
		ir: 'productos',
		boton: 'Ir a productos'
	},
	{
		re: /foto|imagen/,
		texto:
			'En el producto tocá **Sacar foto** o **Galería**: hasta 8 fotos, la primera es la portada. Con buena luz natural y fondo liso se venden más.',
		ir: 'productos',
		boton: 'Ir a productos'
	},
	{
		re: /cupon|descuento|codigo de descuento|promo/,
		texto:
			'En **Cupones** creás el código, el descuento (porcentaje, monto fijo o envío gratis) y hasta cuándo vale. Después lo compartís con un link.',
		ir: 'cupones',
		boton: 'Crear un cupón'
	},
	{
		re: /envio|envios|correo argentino|retiro|entrega/,
		texto:
			'En **Datos del negocio y cobro** (en la web, **Configuración → Envíos**) elegís cómo entregás: retiro en el local, Correo Argentino o tu propio envío, y desde cuánto es gratis.',
		ir: 'negocio',
		boton: 'Configurar envíos'
	},
	{
		re: /mercado ?pago|tarjeta|cuotas|qr/,
		texto:
			'En **Integraciones → Mercado Pago** conectás tu cuenta y tus clientes pagan con tarjeta, cuotas, débito o QR. La plata va directo a tu Mercado Pago.',
		ir: 'integraciones',
		boton: 'Conectar Mercado Pago'
	},
	{
		re: /transferencia|alias|cbu|cvu/,
		texto:
			'En **Datos del negocio y cobro** cargás tu alias o CBU: así aparece «Transferencia» al pagar. Cuando te llega la plata, confirmás el pedido con un toque.',
		ir: 'negocio',
		boton: 'Cargar mis datos'
	},
	{
		re: /link de pago|cobrar a alguien|mandar a pagar/,
		texto:
			'En **Links de pago** armás un link con el monto y el concepto, y se lo mandás por WhatsApp: paga con Mercado Pago.',
		ir: 'links',
		boton: 'Crear un link'
	},
	{
		re: /mercado ?libre|publicar en ml|\bml\b|meli/,
		texto:
			'En **Integraciones → Mercado Libre** conectás tu cuenta y publicás tus productos desde acá, con el stock y el precio sincronizados.',
		ir: 'integraciones',
		boton: 'Conectar Mercado Libre'
	},
	{
		re: /dominio|mi propia direccion|\.com|direccion de la tienda|cambiar el link/,
		texto:
			'En **Editar mi tienda** cambiás la dirección (tutienda.tulocalweb.com.ar). Un dominio propio (.com.ar) lo conectás desde el panel web en **Cuenta → Dominio**.',
		ir: 'mi-tienda',
		boton: 'Editar mi tienda'
	},
	{
		re: /factur|arca|afip|comprobante/,
		texto:
			'En **Facturación de ARCA** cargás tu CUIT y tu certificado; después cada venta se puede facturar con un toque (o sola, al cobrar).',
		ir: 'facturacion',
		boton: 'Ir a facturación'
	},
	{
		re: /equipo|empleado|usuario|permiso|invitar|socio/,
		texto:
			'En **Usuarios y permisos** invitás a tu gente con su mail y elegís el rol: administración, ventas o local.',
		ir: 'equipo',
		boton: 'Invitar a alguien'
	},
	{
		re: /importar|traer (mis|la|mi)|migrar|tiendanube|shopify|woocommerce|empretienda|otra plataforma/,
		texto:
			'En **Traer mi tienda de otra plataforma** elegís Tiendanube, Shopify, WooCommerce o Empretienda y traemos productos, clientes y pedidos. Los links viejos redirigen solos.',
		ir: 'importar',
		boton: 'Traer mi tienda'
	},
	{
		re: /compartir|difundir|promocionar|instagram|estado|historia|qr/,
		texto:
			'En **Compartí tu tienda** tenés el link, la imagen lista para estados e historias, el QR para imprimir y mensajes armados para WhatsApp.',
		ir: 'compartir',
		boton: 'Compartir la tienda'
	},
	{
		re: /plan|abono|pagar la suscripcion|precio de tu local web|cuanto cuesta tu local/,
		texto:
			'En **Plan** ves tu plan, qué incluye y lo cambiás cuando quieras. Se paga con Mercado Pago.',
		ir: 'plan',
		boton: 'Ver mi plan'
	},
	{
		re: /devolu|cambio|reembolso/,
		texto:
			'En **Devoluciones** registrás el cambio o la devolución: el stock vuelve solo y, si corresponde, hacés el reembolso.',
		ir: 'devoluciones',
		boton: 'Ver devoluciones'
	},
	{
		re: /vend(er|o) en el local|punto de venta|caja|cobr(ar|o) en el local|escanear|venta en el local/,
		texto:
			'Tocá **Vender** (el «+» › Venta en el local): escaneás o buscás los productos, cobrás en efectivo, QR o tarjeta y el stock se descuenta solo.',
		ir: 'venta',
		boton: 'Vender en el local'
	},
	{
		re: /stock|inventario|contar/,
		texto:
			'El stock de cada producto lo cambiás en el producto. Para un conteo, en **Inventario** escaneás los códigos y ajustás todo de una.',
		ir: 'productos',
		boton: 'Ir a productos'
	},
	{
		re: /precio|aumentar|subir los precios|actualizar precios/,
		texto:
			'Cambiás el precio en cada producto. Para todos juntos, en el panel web: **Productos → Seleccionar → Cambiar precio** (un porcentaje para todos).',
		ir: 'productos',
		boton: 'Ir a productos'
	},
	{
		re: /telegram/,
		texto:
			'En **Integraciones → Tincho en Telegram** lo vinculás: te aviso de cada venta y me preguntás lo que quieras por chat.',
		ir: 'integraciones',
		boton: 'Vincular Telegram'
	},
	{
		re: /whatsapp/,
		texto:
			'WhatsApp Business todavía no se puede conectar: está en camino. Mientras, el botón de WhatsApp de tu tienda abre un chat con el número de **Datos del negocio y cobro**.',
		ir: 'negocio',
		boton: 'Mis datos'
	},
	{
		re: /notificacion|aviso|alerta/,
		texto:
			'En **Perfil, seguridad y tema → Notificaciones** elegís qué te avisamos (ventas, stock, mensajes) y por dónde.',
		ir: 'negocio'
	},
	{
		re: /google|analytics|pixel|meta|facebook ads/,
		texto:
			'En **Integraciones → Google y Meta** cargás tu Google Analytics y tu píxel de Meta, y tenés el link del catálogo para Google Shopping y Meta.',
		ir: 'integraciones',
		boton: 'Ir a integraciones'
	}
];

/** La guía de la pregunta, si pide cómo hacer algo. */
export function guiaDe(q: string, pideComo: boolean): Respuesta | null {
	if (!pideComo) return null;
	const g = GUIAS.find((x) => x.re.test(q));
	if (!g) return null;
	return {
		texto: g.texto,
		...(g.ir ? { acciones: [{ texto: g.boton ?? 'Ir', ir: g.ir }] } : {}),
		...(g.sugerencias ? { sugerencias: g.sugerencias } : {}),
		tema: 'guia'
	};
}
