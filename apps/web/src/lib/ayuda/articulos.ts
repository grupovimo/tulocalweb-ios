/**
 * Centro de ayuda (design/Mostrador Ayuda.dc.html): categorías y artículos. El handoff los deja en
 * "#"; acá son artículos de verdad, y los conteos y los minutos de lectura salen del contenido.
 */
import * as I from '#lib/icons.ts';

export type Bloque =
	| { t: 'p'; x: string }
	| { t: 'pasos'; x: string[] }
	| { t: 'lista'; x: string[] }
	| { t: 'nota'; x: string }
	| { t: 'h'; x: string };

export type Categoria = {
	slug: string;
	nombre: string;
	icono: I.IconData;
	bg: string;
	fg: string;
};

export type Articulo = {
	slug: string;
	categoria: string;
	titulo: string;
	resumen: string;
	cuerpo: Bloque[];
	/** Aparece en "Artículos más leídos". */
	destacado?: number;
	/** Palabras extra para el buscador. */
	claves?: string;
};

export const CATEGORIAS: Categoria[] = [
	{
		slug: 'primeros-pasos',
		nombre: 'Primeros pasos',
		icono: I.phRocketLaunchFill,
		bg: '#FCEDE5',
		fg: '#b0441d'
	},
	{
		slug: 'productos',
		nombre: 'Productos y stock',
		icono: I.phTagFill,
		bg: '#E8F1FF',
		fg: '#3E6BFD'
	},
	{
		slug: 'pedidos',
		nombre: 'Pedidos y envíos',
		icono: I.phPackageFill,
		bg: '#FFF4DC',
		fg: '#8a5f0e'
	},
	{
		slug: 'cobros',
		nombre: 'Cobros y medios de pago',
		icono: I.phCreditCardFill,
		bg: '#E6F4EC',
		fg: '#237550'
	},
	{
		slug: 'whatsapp',
		nombre: 'WhatsApp y Tincho',
		icono: I.phWhatsappLogoFill,
		bg: '#E9F8EE',
		fg: '#1F8A4C'
	},
	{
		slug: 'facturacion',
		nombre: 'Facturación y ARCA',
		icono: I.phReceiptFill,
		bg: '#F1ECFF',
		fg: '#6A4FA0'
	},
	{
		slug: 'diseno',
		nombre: 'Diseño de la tienda',
		icono: I.phPaintBrushBroadFill,
		bg: '#FCEEF3',
		fg: '#B4475B'
	},
	{ slug: 'cuenta', nombre: 'Cuenta y plan', icono: I.phGearSixFill, bg: '#EFEAE4', fg: '#5C4536' }
];

export const ARTICULOS: Articulo[] = [
	// ------------------------------------------------------------------ Primeros pasos
	{
		slug: 'crear-tu-tienda',
		categoria: 'primeros-pasos',
		titulo: 'Crear tu tienda en 10 minutos',
		resumen: 'Del registro a la tienda publicada, con tus productos adentro.',
		cuerpo: [
			{
				t: 'p',
				x: 'El registro es el mismo recorrido que arma tu tienda: ocho pasos cortos, y en cada uno podés seguir sin completar todo y volver después.'
			},
			{
				t: 'pasos',
				x: [
					'Entrá a Crear mi tienda y elegí cómo ingresar: email, Google, Mercado Pago o Mercado Libre.',
					'Contanos qué vendés. Si tenés CUIT, traemos tu razón social y tu condición fiscal desde ARCA.',
					'Conectá los canales que ya usás: WhatsApp, Instagram o Mercado Libre.',
					'Importá tus productos desde Excel, Tiendanube o Mercado Libre. Si un precio no coincide, elegís cuál queda.',
					'Elegí una plantilla, cómo cobrás y cómo enviás.',
					'Elegí el tono de Tincho y publicá. Te queda el QR de la tienda para imprimir o compartir.'
				]
			},
			{
				t: 'nota',
				x: 'Todos los planes arrancan con 14 días de prueba. No te pedimos tarjeta para empezar.'
			}
		],
		claves: 'registro onboarding empezar alta cuenta nueva'
	},
	{
		slug: 'probar-la-demo',
		categoria: 'primeros-pasos',
		titulo: 'Probar la demo antes de crear la cuenta',
		resumen: 'Una tienda de ejemplo con pedidos, clientes y Tincho, para tocar todo.',
		cuerpo: [
			{
				t: 'p',
				x: 'En la demo elegís una plantilla y entrás al panel de una tienda de ejemplo con dos meses de pedidos, clientes, stock y conversaciones. Podés cambiar precios, cobrar en el punto de venta, crear cupones o mandar un link de pago: todo funciona igual que en una cuenta real.'
			},
			{
				t: 'lista',
				x: [
					'Lo que cambiás queda guardado en tu navegador durante 24 horas.',
					'Nadie más ve tu demo: cada visitante tiene la suya.',
					'Los pagos y los mensajes se simulan; no se cobra ni se manda nada.'
				]
			},
			{
				t: 'p',
				x: 'Cuando quieras pasar a tu tienda, tocá Crear mi tienda: la demo no se mezcla con tu cuenta.'
			}
		],
		claves: 'demo ejemplo probar sandbox'
	},
	{
		slug: 'simple-y-completo',
		categoria: 'primeros-pasos',
		titulo: 'Recorrer el panel: modo Simple y modo Completo',
		resumen: 'Dos formas de ver el panel, según cuánto querés tener a mano.',
		cuerpo: [
			{
				t: 'p',
				x: 'El modo Simple deja sólo lo del día a día: Resumen, Pedidos, Productos, WhatsApp, Clientes y Cobros. El modo Completo suma todas las secciones, agrupadas en Ventas, Catálogo, Crecer, Operación, Tienda y Cuenta.'
			},
			{
				t: 'lista',
				x: [
					'Cambiás de modo desde abajo de la barra lateral; el panel lo recuerda.',
					'Con ⌘K (o Ctrl+K) buscás cualquier sección o le preguntás a Tincho.',
					'En el celular, la barra de abajo lleva a lo más usado y Más abre el resto.'
				]
			}
		],
		claves: 'panel menu navegar atajos cmd k'
	},

	// ------------------------------------------------------------------ Productos y stock
	{
		slug: 'importar-productos',
		categoria: 'productos',
		titulo: 'Importar productos desde Excel o Tiendanube',
		resumen: 'Subí tu planilla y reconocemos las columnas solas.',
		destacado: 6,
		cuerpo: [
			{
				t: 'p',
				x: 'Desde Productos › Importar subís un CSV o un Excel. Reconocemos las columnas de Tiendanube, Mercado Libre y de cualquier planilla con nombre, precio y stock.'
			},
			{
				t: 'pasos',
				x: [
					'Exportá tus productos desde la otra plataforma (o armá una planilla con nombre, precio, stock y SKU).',
					'En Productos tocá Importar y elegí el archivo.',
					'Revisá lo que encontramos: los productos nuevos se crean y los que ya existen (mismo SKU) se actualizan.'
				]
			},
			{
				t: 'nota',
				x: '¿Venís de otra plataforma con muchos productos? En Migrar te mudamos gratis en un día.'
			}
		],
		claves: 'csv excel planilla importar subir tiendanube mercado libre'
	},
	{
		slug: 'cargar-con-foto',
		categoria: 'productos',
		titulo: 'Cargar un producto con foto',
		resumen: 'Subí las fotos desde el celular y completá nombre, precio y stock.',
		cuerpo: [
			{
				t: 'p',
				x: 'Desde ⌘K › Cargar producto con foto, o desde Productos › Nuevo, subís las fotos (la primera es la portada). Después ponés el nombre, el precio y el stock y lo publicás.'
			},
			{
				t: 'p',
				x: 'En la ficha de cualquier producto, Mejorar con Tincho reescribe la descripción con el tono de tu tienda. Siempre la ves antes de guardarla.'
			}
		],
		claves: 'foto camara hornero ficha descripcion'
	},
	{
		slug: 'inventario',
		categoria: 'productos',
		titulo: 'Inventario: conteo, movimientos y pedido al proveedor',
		resumen: 'El mismo stock para la tienda, el local y Mercado Libre.',
		cuerpo: [
			{
				t: 'p',
				x: 'Cada venta descuenta stock en todos los canales a la vez, así que no vendés dos veces lo mismo. En Inventario ves cuánto te alcanza cada producto según lo que vendés por día.'
			},
			{
				t: 'lista',
				x: [
					'Modo conteo: escaneás o tipeás lo que hay en el estante y ajustamos la diferencia.',
					'Movimientos de hoy: ventas, ajustes e ingresos, con quién los hizo.',
					'Para reponer: armamos el pedido al proveedor y lo mandás por WhatsApp.',
					'Códigos de barra: imprimís las etiquetas Code 39 desde la misma pantalla.'
				]
			}
		],
		claves: 'stock conteo proveedor reponer codigo de barras etiquetas'
	},
	{
		slug: 'listas-de-precio',
		categoria: 'productos',
		titulo: 'Listas de precio y aumento por inflación',
		resumen: 'Mayorista, Mercado Libre y revendedoras sobre un mismo precio base.',
		cuerpo: [
			{
				t: 'p',
				x: 'Cada producto tiene un precio base (minorista) y las listas aplican una regla encima: por ejemplo −25 % mayorista o +10 % en Mercado Libre para cubrir la comisión.'
			},
			{
				t: 'p',
				x: 'Para actualizar por inflación, elegís el porcentaje y el redondeo (terminado en 90 o múltiplo de 500) y ves cómo queda cada lista antes de aplicar. Podés aplicarlo ya, con la opción de deshacer, o programarlo para el lunes a las 00:00.'
			}
		],
		claves: 'precios aumento inflacion mayorista redondeo'
	},

	{
		slug: 'mercado-libre',
		categoria: 'productos',
		titulo: 'Vender en Mercado Libre con el mismo stock',
		resumen: 'Precios, stock y preguntas sincronizados con tu tienda.',
		cuerpo: [
			{
				t: 'p',
				x: 'Conectás tu cuenta de Mercado Libre desde Integraciones y elegís qué productos publicar. El stock es uno solo: si vendés en el local o en la tienda, Mercado Libre se actualiza al instante.'
			},
			{
				t: 'lista',
				x: [
					'El precio sale de la lista Mercado Libre (por ejemplo +10 % para cubrir la comisión).',
					'Las preguntas llegan a Reseñas y preguntas, con la respuesta que propone Tincho.',
					'Los pedidos de Mercado Libre entran en Pedidos como cualquier otro.'
				]
			}
		],
		claves: 'mercado libre ml publicaciones preguntas sincronizar'
	},

	// ------------------------------------------------------------------ Pedidos y envíos
	{
		slug: 'envios-por-codigo-postal',
		categoria: 'pedidos',
		titulo: 'Configurar envíos por código postal con Correo Argentino',
		resumen: 'El cliente ve el costo y el plazo antes de pagar.',
		destacado: 2,
		cuerpo: [
			{
				t: 'p',
				x: 'Conectás tu cuenta de MiCorreo de Correo Argentino desde Integraciones. En el checkout, el cliente pone su código postal y ve el costo y cuándo le llega.'
			},
			{
				t: 'lista',
				x: [
					'Envío gratis desde un monto: se configura en Cupones › Automáticas.',
					'Moto propia: definís zonas y precio fijo.',
					'Retiro en el local: sin costo, con aviso cuando está listo.'
				]
			},
			{
				t: 'p',
				x: 'En Envíos imprimís las etiquetas de todos los pedidos listos de una vez y seguís cada paquete en el tablero.'
			}
		],
		claves: 'envio andreani oca correo argentino moto codigo postal etiquetas'
	},
	{
		slug: 'circuito-de-un-pedido',
		categoria: 'pedidos',
		titulo: 'El circuito de un pedido: de Sin pagar a Entregado',
		resumen: 'Cinco estados, y qué pasa en cada uno.',
		cuerpo: [
			{
				t: 'pasos',
				x: [
					'Sin pagar: el cliente armó el pedido y falta el pago (por ejemplo, una transferencia). Le podés recordar por WhatsApp.',
					'Para preparar: está cobrado. Si tenés ARCA conectado, la factura ya salió.',
					'Listo para despachar: el paquete está armado y la etiqueta impresa.',
					'En camino: lo tiene el correo o la moto; el cliente recibe el seguimiento.',
					'Entregado: llegó. Desde acá puede pedir una devolución.'
				]
			},
			{
				t: 'p',
				x: 'Cada pedido avanza o vuelve de a un paso, o se cancela. Desde la lista podés mover varios juntos, imprimir sus etiquetas o exportarlos a Excel.'
			}
		],
		claves: 'estados pedido preparar despachar entregado cancelar'
	},
	{
		slug: 'crear-pedido-a-mano',
		categoria: 'pedidos',
		titulo: 'Crear un pedido a mano y cobrarlo con un link',
		resumen: 'Para lo que te piden por WhatsApp, Instagram o en el mostrador.',
		cuerpo: [
			{
				t: 'p',
				x: 'En Pedidos › Crear pedido (o desde una charla de la bandeja) elegís el cliente, los productos y la entrega. Si ya te pagó, lo marcás pagado y queda para preparar.'
			},
			{
				t: 'p',
				x: 'Si no, elegís Mandar link de pago: el pedido queda sin pagar y armamos el link con el monto. Cuando el cliente paga, el pedido pasa solo a preparar.'
			}
		],
		claves: 'pedido manual link de pago whatsapp'
	},

	// ------------------------------------------------------------------ Cobros
	{
		slug: 'mercado-pago-y-cuotas',
		categoria: 'cobros',
		titulo: 'Cómo conectar Mercado Pago y cobrar en cuotas',
		resumen: 'Tu cuenta de Mercado Pago, con la plata directo ahí.',
		destacado: 1,
		cuerpo: [
			{
				t: 'p',
				x: 'En Integraciones › Mercado Pago tocás Conectar e iniciás sesión con tu cuenta. Las ventas se acreditan en tu Mercado Pago, sin pasar por Tu Local Web.'
			},
			{
				t: 'pasos',
				x: [
					'Conectá Mercado Pago.',
					'En Cupones › Automáticas prendé Cuotas sin interés y elegí desde qué monto.',
					'Listo: en el checkout aparecen 3 y 6 cuotas sin interés.'
				]
			},
			{
				t: 'nota',
				x: 'Tu Local Web no cobra comisión por venta en ningún plan. Lo que pagás de cuotas o acreditación lo cobra Mercado Pago según tu cuenta.'
			}
		],
		claves: 'mercado pago cuotas tarjeta cobrar acreditacion'
	},
	{
		slug: 'punto-de-venta',
		categoria: 'cobros',
		titulo: 'Vender en el local con el punto de venta',
		resumen: 'El mismo stock que la tienda online, con lector de códigos.',
		cuerpo: [
			{
				t: 'p',
				x: 'En Ventas en el local tocás los productos (o los escaneás con el lector) y cobrás con QR de Mercado Pago, efectivo, tarjeta, MODO o transferencia. Con Dividir cobrás una parte con un medio y el resto con otro.'
			},
			{ t: 'p', x: 'Cada venta descuenta stock, entra en la caja del día y sale con su factura B.' }
		],
		claves: 'local pos caja efectivo lector qr'
	},
	{
		slug: 'cupones-y-promociones',
		categoria: 'cobros',
		titulo: 'Cupones, promociones automáticas y ofertas relámpago',
		resumen: 'Descuentos que se miden: cuánto venden y cuánto te cuestan.',
		cuerpo: [
			{
				t: 'p',
				x: 'El armador de cupones te pregunta para qué es (primera compra, que vuelvan, liquidar stock, influencer) y antes de crearlo te muestra cuánto estima vender, cuánto te cuesta y qué margen te queda.'
			},
			{
				t: 'lista',
				x: [
					'Cupones: con código, mínimo de compra, límite de usos y vigencia.',
					'Automáticas: se aplican solas en el carrito (transferencia, cantidad, envío gratis, cuotas).',
					'Relámpago: un producto con descuento por unas horas y unidades contadas.'
				]
			}
		],
		claves: 'cupon descuento promocion relampago codigo'
	},

	// ------------------------------------------------------------------ WhatsApp y Tincho
	{
		slug: 'whatsapp-business',
		categoria: 'whatsapp',
		titulo: 'Conectar tu número de WhatsApp Business',
		resumen: 'El mismo número de siempre, con Tincho atendiendo.',
		destacado: 4,
		cuerpo: [
			{
				t: 'p',
				x: 'Desde Integraciones › WhatsApp conectás tu número con la API oficial de WhatsApp Business (Meta). Seguís usando el mismo número; los mensajes llegan a la bandeja del panel, donde atienden vos, tu equipo y Tincho.'
			},
			{
				t: 'pasos',
				x: [
					'Tené a mano el celular con el número y acceso a tu cuenta de Facebook.',
					'Tocá Conectar y seguí los pasos de Meta para verificar el número.',
					'Elegí si Tincho responde siempre, fuera de horario o nunca.'
				]
			},
			{
				t: 'nota',
				x: 'Los mensajes que inicia la tienda (campañas, recordatorios) los cobra Meta por mensaje. Antes de mandar te mostramos el costo exacto.'
			}
		],
		claves: 'whatsapp business numero meta bandeja conectar'
	},
	{
		slug: 'que-responde-hornero',
		categoria: 'whatsapp',
		titulo: 'Qué responde Tincho y cuándo te pasa la conversación',
		resumen: 'Responde con tus datos reales y te avisa cuando hace falta una persona.',
		cuerpo: [
			{
				t: 'p',
				x: 'Tincho conoce tu catálogo, tu stock, tus envíos y tus precios. Responde consultas de stock y envío, manda el producto con foto, arma el pedido y manda el link de pago.'
			},
			{
				t: 'lista',
				x: [
					'Te pasa la conversación cuando piden algo que no sabe, hay un reclamo o el cliente quiere hablar con una persona.',
					'Tomar una conversación la saca de Tincho hasta que la devolvés.',
					'En cada charla te sugiere la respuesta; la mandás tal cual o la cambiás.'
				]
			}
		],
		claves: 'hornero ia inteligencia artificial respuestas automaticas'
	},
	{
		slug: 'campanas-por-whatsapp',
		categoria: 'whatsapp',
		titulo: 'Campañas por WhatsApp con el costo exacto antes de enviar',
		resumen: 'A quién, qué mensaje y cuánto cuesta, antes de tocar Enviar.',
		cuerpo: [
			{
				t: 'p',
				x: 'En Marketing elegís la audiencia (todos, VIP, en riesgo o quienes compraron una categoría), el mensaje (Tincho te propone y reescribe) y si va con foto, botón al producto o cupón. Ves la vista previa en el teléfono, el costo exacto y las ventas estimadas según tus campañas anteriores.'
			},
			{
				t: 'p',
				x: 'Podés mandarla ya o programarla para el sábado a las 10, el mejor horario de la mayoría de las tiendas.'
			}
		],
		claves: 'campaña marketing whatsapp masivo difusion'
	},

	// ------------------------------------------------------------------ Facturación
	{
		slug: 'factura-automatica',
		categoria: 'facturacion',
		titulo: 'Emitir factura electrónica automática en cada venta',
		resumen: 'La factura sale sola, con CAE, y le llega al cliente.',
		destacado: 3,
		cuerpo: [
			{
				t: 'p',
				x: 'Con tu certificado de ARCA cargado, cada venta cobrada emite su factura electrónica con CAE y se la mandamos al cliente con el pedido. Las de transferencia salen cuando confirmás el pago.'
			},
			{
				t: 'lista',
				x: [
					'Factura manual: para ventas por fuera de la tienda, desde Facturación.',
					'Notas de crédito: al aprobar una devolución.',
					'Te avisamos cuando te acercás al tope de tu categoría de monotributo.'
				]
			}
		],
		claves: 'factura electronica arca afip cae automatica'
	},
	{
		slug: 'certificado-arca',
		categoria: 'facturacion',
		titulo: 'Subir el certificado de ARCA (y renovarlo)',
		resumen: 'Una vez, y te avisamos antes de que venza.',
		cuerpo: [
			{
				t: 'pasos',
				x: [
					'En ARCA, con clave fiscal, generá el certificado para web services de facturación.',
					'En Facturación › Certificado subí el archivo .crt y elegí el punto de venta.',
					'Hacemos una factura de prueba en homologación y, si sale bien, quedás facturando.'
				]
			},
			{
				t: 'p',
				x: 'El certificado dura dos años. Te avisamos con tiempo en el Resumen y en Para hacer hoy para que lo renueves sin cortes.'
			}
		],
		claves: 'certificado arca afip clave fiscal punto de venta renovar'
	},
	{
		slug: 'factura-a-b-c',
		categoria: 'facturacion',
		titulo: 'Factura A, B o C: cuál corresponde',
		resumen: 'Depende de tu condición y la de quien compra.',
		cuerpo: [
			{
				t: 'lista',
				x: [
					'Si sos monotributista, emitís factura C a todos.',
					'Si sos responsable inscripto, emitís A a otros responsables inscriptos (con su CUIT) y B a consumidores finales y monotributistas.'
				]
			},
			{
				t: 'p',
				x: 'Cuando el cliente carga su CUIT en el checkout, consultamos el padrón de ARCA y elegimos el tipo correcto solo.'
			}
		],
		claves: 'factura a b c monotributo responsable inscripto cuit'
	},

	// ------------------------------------------------------------------ Diseño
	{
		slug: 'elegir-plantilla',
		categoria: 'diseno',
		titulo: 'Elegir y cambiar de plantilla',
		resumen: 'Dieciséis plantillas pensadas por rubro, y cambiar no rompe nada.',
		cuerpo: [
			{
				t: 'p',
				x: 'Hay plantillas para cada rubro: moda, gastronomía, vinos, deco, tecnología, servicios con turnos y más. Algunas suman funciones propias, como la reserva de 10 minutos en los drops o la verificación de edad para vinos.'
			},
			{
				t: 'p',
				x: 'Desde Mi tienda cambiás de plantilla cuando quieras: tus productos, precios y pedidos quedan igual.'
			}
		],
		claves: 'plantilla tema diseño template'
	},
	{
		slug: 'colores-logo-portada',
		categoria: 'diseno',
		titulo: 'Personalizar colores, logo y portada',
		resumen: 'Lo ves cambiar en vivo antes de publicar.',
		cuerpo: [
			{
				t: 'p',
				x: 'En Mi tienda subís tu logo, elegís los colores y la tipografía, y cambiás la portada y los banners. Todo se ve en vivo en la vista de computadora y de celular; publicás cuando te gusta.'
			}
		],
		claves: 'logo colores portada banner tipografia editor'
	},
	{
		slug: 'promociones-en-la-tienda',
		categoria: 'diseno',
		titulo: 'Ventanas de promociones, captar mails y ruleta de premios',
		resumen: 'Avisos que se abren solos y ayudas que empujan la compra, sin molestar.',
		cuerpo: [
			{
				t: 'p',
				x: 'En Mi tienda › Promociones armás ventanas que se abren solas en tu tienda: una promo con botón, un cupón tuyo que se aplica solo al carrito, un formulario para captar mails a cambio de un premio o una ruleta de premios. Arrancás desde una plantilla y la ves en la vista previa antes de publicar.'
			},
			{
				t: 'pasos',
				x: [
					'Entrá a Mi tienda › Promociones y tocá «Nueva ventana».',
					'Elegí una plantilla: Bienvenida 10 %, Antes de irte, Ruleta, Envío gratis, Lanzamiento, Descargá la guía o Sumate por WhatsApp.',
					'Cambiá los textos, la foto y el premio; elegí el diseño (centrada, en un costado, pantalla completa o desde abajo).',
					'Decidí cuándo sale (al entrar, después de unos segundos, al bajar, antes de irse o al agregar al carrito), a quién y en qué páginas.',
					'Tocá «Ver en la vista previa» y, cuando te guste, «Publicar cambios».'
				]
			},
			{ t: 'h', x: 'Reglas que cuidan la compra' },
			{
				t: 'lista',
				x: [
					'Sale una sola ventana por página; si varias califican, la primera de la lista.',
					'Nunca en el pago, en la página de compra lista, en los legales ni en el botón de arrepentimiento.',
					'En el celular, la pantalla completa sube como una hoja desde abajo y deja ver la tienda.',
					'Siempre se cierra con la X, con Esc o tocando afuera, y no tapa WhatsApp ni la barra del pedido.',
					'Quien ya dejó su mail en una ventana no la vuelve a ver.'
				]
			},
			{ t: 'h', x: 'Captar mails y ruleta' },
			{
				t: 'p',
				x: 'El contacto entra a tu Newsletter con el nombre de la ventana como origen. El premio puede ser un cupón tuyo, un código propio de un solo uso (porcentaje, monto o envío gratis), una descarga o nada. En la ruleta, el premio lo sortea el servidor según el peso de cada porción, una vez por mail.'
			},
			{ t: 'h', x: 'Herramientas para vender más' },
			{
				t: 'lista',
				x: [
					'Cuánto falta para el envío gratis, en la bolsa y en el carrito.',
					'Compras recientes: avisos chicos con pedidos reales de los últimos 7 días (nombre de pila y provincia).',
					'Etiquetas solas en los productos: Nuevo, oferta, Últimas unidades y Más vendido.',
					'«Quedan 3» en la ficha, el botón de comprar fijo en el celular y «Sumá también» en el carrito.',
					'Favoritos: un corazón en los productos y la página Favoritos.'
				]
			},
			{
				t: 'nota',
				x: 'En Promociones ves cuántas personas vieron cada ventana, cuántas dejaron el mail y cuántas ventas trajo su cupón.'
			}
		],
		claves:
			'popup ventana emergente promo lead magnet captar mail ruleta premio cupon salida conversion favoritos'
	},

	// ------------------------------------------------------------------ Cuenta y plan
	{
		slug: 'dominio-propio',
		categoria: 'cuenta',
		titulo: 'Usar tu propio dominio (.com.ar)',
		resumen: 'Tu tienda en tunegocio.com.ar, con el candado incluido.',
		destacado: 5,
		cuerpo: [
			{
				t: 'pasos',
				x: [
					'Si todavía no lo tenés, registralo en NIC Argentina (.com.ar) o en el registrador que prefieras.',
					'En Cuenta › Dominio escribí tu dominio.',
					'Copiá los registros DNS que te mostramos en tu registrador.',
					'Cuando se propagan (de minutos a unas horas), la tienda queda en tu dominio con certificado SSL.'
				]
			},
			{
				t: 'nota',
				x: 'Mientras tanto, tu tienda sigue funcionando en la dirección de Tu Local Web.'
			}
		],
		claves: 'dominio com.ar dns nic ssl'
	},
	{
		slug: 'planes-y-prueba',
		categoria: 'cuenta',
		titulo: 'Planes y prueba de 14 días',
		resumen: 'Sin comisión por venta en ningún plan.',
		cuerpo: [
			{
				t: 'lista',
				x: [
					'Gratis: para empezar a vender.',
					'Emprende: $ 19.900 por mes.',
					'Pro: $ 49.900 por mes.',
					'Negocio: $ 149.900 por mes.'
				]
			},
			{
				t: 'p',
				x: 'Pagando anual tenés 25 % de descuento. Todos los planes arrancan con 14 días de prueba y ninguno cobra comisión por venta. Si Tincho se te queda corto, sumás el pack de 2.000 respuestas por $ 4.900.'
			},
			{
				t: 'p',
				x: 'Cambiás de plan cuando quieras desde Cuenta › Plan y facturación; la diferencia se calcula por los días que quedan.'
			}
		],
		claves: 'plan precio abono suscripcion prueba gratis comision'
	},
	{
		slug: 'equipo-y-permisos',
		categoria: 'cuenta',
		titulo: 'Invitar a tu equipo y elegir permisos',
		resumen: 'Cada persona ve sólo lo que necesita.',
		cuerpo: [
			{
				t: 'p',
				x: 'En Cuenta › Usuarios invitás por email o WhatsApp y elegís qué puede hacer cada uno: por ejemplo, alguien en la caja del local que cobra pero no ve finanzas, o quien atiende la bandeja sin tocar precios.'
			},
			{
				t: 'p',
				x: 'En Actividad ves quién hizo cada cambio y desde dónde, y la mayoría se puede deshacer.'
			}
		],
		claves: 'usuarios equipo permisos invitar roles'
	}
];

/** Minutos de lectura (unas 180 palabras por minuto). */
export function minutos(a: Articulo) {
	const texto = a.cuerpo.map((b) => (Array.isArray(b.x) ? b.x.join(' ') : b.x)).join(' ');
	return Math.max(1, Math.round(texto.split(/\s+/).length / 180));
}

export const categoria = (slug: string) => CATEGORIAS.find((c) => c.slug === slug);
export const delaCategoria = (slug: string) => ARTICULOS.filter((a) => a.categoria === slug);
export const destacados = () =>
	ARTICULOS.filter((a) => a.destacado).sort((a, b) => a.destacado! - b.destacado!);

const normal = (s: string) => s.toLowerCase().normalize('NFD').replace(/[̀-ͯ]/g, '');

/** Palabras que no dicen nada en una pregunta. */
const VACIAS = new Set(
	'como que cual cuando donde para por con los las del una uno unos mis mio tus sus hago hacer tengo pongo puedo quiero necesito esta este eso hay'.split(
		' '
	)
);

/** Palabra entera para las letras sueltas (A, B, C); si no, comienzo de palabra ("factur…"). */
const hay = (t: string, p: string) => new RegExp(p.length === 1 ? `\\b${p}\\b` : `\\b${p}`).test(t);

/** Buscador: puntúa título, resumen, claves y cuerpo. */
export function buscar(q: string, cuantos = 8) {
	const palabras = normal(q)
		.replace(/[^a-z0-9ñ ]/g, ' ')
		.split(/\s+/)
		.filter((p) => (p.length > 2 || /^[abc]$/.test(p)) && !VACIAS.has(p));
	if (!palabras.length) return [];
	return ARTICULOS.map((a) => {
		const titulo = normal(a.titulo);
		const resto = normal(`${a.resumen} ${a.claves ?? ''}`);
		const cuerpo = normal(
			a.cuerpo.map((b) => (Array.isArray(b.x) ? b.x.join(' ') : b.x)).join(' ')
		);
		const puntos = palabras.reduce(
			(s, p) => s + (hay(titulo, p) ? 5 : 0) + (hay(resto, p) ? 3 : 0) + (hay(cuerpo, p) ? 1 : 0),
			0
		);
		return { a, puntos };
	})
		.filter((x) => x.puntos > 0)
		.sort((x, y) => y.puntos - x.puntos)
		.slice(0, cuantos)
		.map((x) => x.a);
}

export const FAQ: [string, string][] = [
	[
		'¿Cuánto tarda en estar lista mi tienda?',
		'Si tenés tus productos en Excel, Tiendanube o Mercado Libre, en 10 minutos. El registro importa todo y te deja la tienda publicada.'
	],
	[
		'¿Puedo cobrar en cuotas sin interés?',
		'Sí. Conectás Mercado Pago o Payway y prendés las cuotas que quieras ofrecer desde Cupones › Automáticas.'
	],
	[
		'¿Tu Local Web cobra comisión por venta?',
		'No, en ningún plan. Pagás el abono del plan y nada más; las comisiones de cada medio de pago las cobra ese medio.'
	],
	[
		'¿Cómo hago la factura electrónica?',
		'Subís tu certificado de ARCA una vez y Tu Local Web emite la factura A, B o C sola en cada venta.'
	]
];
