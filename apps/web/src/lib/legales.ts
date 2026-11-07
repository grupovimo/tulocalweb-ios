// Generado por scripts/legales.ts desde design/Legal *.dc.html. No editar a mano.
import { phBankFill, phBuildingsFill, phChatsCircleFill } from '#lib/icons.ts';
import type { IconData } from '#lib/icons.ts';

export type BloqueLegal =
	| { t: 'resumen'; items: string[] }
	| {
			t: 'seccion';
			id: string;
			n: string;
			titulo: string;
			cuerpo: ({ p: string } | { ul: string[] })[];
	  }
	| { t: 'organismos'; items: { href: string; icono: IconData; titulo: string; texto: string }[] }
	| { t: 'creditos'; items: { etiqueta: string; umbral: string; credito: string; pie: string }[] }
	| {
			t: 'cookies';
			filas: { nombre: string; tipo: string; color: string; para: string; dura: string }[];
	  }
	| { t: 'configurar-cookies' }
	| { t: 'arrepentimiento' };

export type Legal = {
	slug: string;
	titulo: string;
	/** Cómo figura en la lista lateral "Legales". */
	corto: string;
	miga: string;
	descripcion: string;
	chips: string[];
	ilustracion: string;
	indice: [string, string][];
	bloques: BloqueLegal[];
	cierre: { titulo: string; texto: string };
};

export const LEGALES: Legal[] = [
	{
		slug: 'terminos',
		titulo: 'Términos y condiciones',
		corto: 'Términos y condiciones',
		miga: 'Términos',
		descripcion: 'Legales de Tu Local Web, claros y sin letra chica.',
		chips: ['Vigente desde el 1 de octubre de 2026', 'Versión 3.2', '3 min de lectura'],
		ilustracion: 'portapapeles',
		indice: [
			['s1', '1. Quiénes somos'],
			['s2', '2. Tu cuenta'],
			['s3', '3. Planes, precios y facturación'],
			['s4', '4. Prueba gratis y cancelación'],
			['s5', '5. Tu contenido y tus ventas'],
			['s6', '6. Cobros y medios de pago'],
			['s7', '7. Disponibilidad del servicio'],
			['s8', '8. Asistente con inteligencia artificial'],
			['s9', '9. Suspensión de cuentas'],
			['s10', '10. Responsabilidad'],
			['s11', '11. Ley aplicable']
		],
		bloques: [
			{
				t: 'resumen',
				items: [
					'Tu tienda, tus productos y los datos de tus clientes son tuyos. Los exportás o los borrás cuando quieras.',
					'Cancelás con un clic desde el panel, sin penalidad ni llamados.',
					'Te avisamos con 30 días de anticipación cualquier cambio de precio o de estos términos.'
				]
			},
			{
				t: 'seccion',
				id: 's1',
				n: '01',
				titulo: 'Quiénes somos',
				cuerpo: [
					{
						p: 'Nahuel Santillán, CUIT 20-44859234-1 (“Tu Local Web”), ofrece una plataforma para crear y administrar tiendas online, vender por WhatsApp y redes, cobrar en el local y facturar. Al crear una cuenta aceptás estos términos en nombre propio o de la empresa que representás.'
					}
				]
			},
			{
				t: 'seccion',
				id: 's2',
				n: '02',
				titulo: 'Tu cuenta',
				cuerpo: [
					{
						p: 'Tenés que ser mayor de 18 años o una persona jurídica, y cargar datos verdaderos. Sos responsable de lo que pase con tu usuario y contraseña: te recomendamos activar la verificación en dos pasos.'
					},
					{
						ul: [
							'Una cuenta puede tener varias personas con distintos permisos.',
							'Si detectás un acceso que no reconocés, avisanos a seguridad@tulocalweb.com.ar.'
						]
					}
				]
			},
			{
				t: 'seccion',
				id: 's3',
				n: '03',
				titulo: 'Planes, precios y facturación',
				cuerpo: [
					{
						p: 'Los planes se pagan por mes o por año, en pesos y con IVA incluido, por débito automático. Te enviamos factura electrónica por cada cobro. Si actualizamos un precio, te avisamos por email y en el panel con 30 días de anticipación; podés cambiar de plan o cancelar antes de que se aplique.'
					}
				]
			},
			{
				t: 'seccion',
				id: 's4',
				n: '04',
				titulo: 'Prueba gratis y cancelación',
				cuerpo: [
					{
						p: 'Los primeros 14 días son gratis y no pedimos tarjeta. Podés cancelar cuando quieras desde Configuración › Plan. Seguís usando Tu Local Web hasta el final del período que ya pagaste y no se te cobra de nuevo.'
					}
				]
			},
			{
				t: 'seccion',
				id: 's5',
				n: '05',
				titulo: 'Tu contenido y tus ventas',
				cuerpo: [
					{
						p: 'Lo que subís (productos, fotos, textos, marcas) sigue siendo tuyo. Nos das permiso para alojarlo y mostrarlo solo para prestarte el servicio. Frente a tus compradores, el vendedor sos vos: sos responsable de precios, stock, envíos, garantías y de cumplir la Ley 24.240 de Defensa del Consumidor.'
					}
				]
			},
			{
				t: 'seccion',
				id: 's6',
				n: '06',
				titulo: 'Cobros y medios de pago',
				cuerpo: [
					{
						p: 'Los pagos de tus ventas los procesan proveedores como Mercado Pago, Payway, MODO o tu banco. Tu Local Web no es una entidad financiera, no recibe ni retiene tu dinero: cada proveedor te liquida según sus propias condiciones.'
					}
				]
			},
			{
				t: 'seccion',
				id: 's7',
				n: '07',
				titulo: 'Disponibilidad del servicio',
				cuerpo: [
					{
						p: 'Nos comprometemos a una disponibilidad mensual del 99,9 % de la tienda y el panel. Podés ver el estado en tiempo real en la página Estado del sistema. Si no cumplimos, te acreditamos parte del plan según la política de Reembolsos.'
					}
				]
			},
			{
				t: 'seccion',
				id: 's8',
				n: '08',
				titulo: 'Asistente con inteligencia artificial',
				cuerpo: [
					{
						p: 'Tincho, el asistente de Tu Local Web, puede redactar respuestas, descripciones y sugerencias. Puede equivocarse: revisá lo importante antes de enviarlo. No usamos tus datos ni los de tus clientes para entrenar modelos de terceros.'
					}
				]
			},
			{
				t: 'seccion',
				id: 's9',
				n: '09',
				titulo: 'Suspensión de cuentas',
				cuerpo: [
					{
						p: 'Podemos suspender una tienda que incumpla la política de Uso aceptable, que no pague o que ponga en riesgo a otros usuarios. Salvo urgencia, te avisamos antes y te damos un plazo para corregirlo.'
					}
				]
			},
			{
				t: 'seccion',
				id: 's10',
				n: '10',
				titulo: 'Responsabilidad',
				cuerpo: [
					{
						p: 'Hacemos lo posible para que Tu Local Web funcione bien, pero no respondemos por pérdidas indirectas ni por fallas de servicios de terceros. Nuestra responsabilidad máxima es el total que nos pagaste en los últimos 12 meses, sin perjuicio de los derechos que te reconoce la ley.'
					}
				]
			},
			{
				t: 'seccion',
				id: 's11',
				n: '11',
				titulo: 'Ley aplicable',
				cuerpo: [
					{
						p: 'Estos términos se rigen por las leyes de la República Argentina. Cualquier conflicto se resuelve ante los tribunales ordinarios de la Ciudad Autónoma de Buenos Aires, salvo que la ley te permita elegir los de tu domicilio.'
					}
				]
			}
		],
		cierre: {
			titulo: '¿Te quedó alguna duda?',
			texto:
				'Escribinos por WhatsApp al +54 223 592-6772 o a legales@tulocalweb.com.ar y te respondemos en 48 horas hábiles.'
		}
	},
	{
		slug: 'privacidad',
		titulo: 'Política de privacidad',
		corto: 'Política de privacidad',
		miga: 'Privacidad',
		descripcion: 'Legales de Tu Local Web, claros y sin letra chica.',
		chips: ['Vigente desde el 30 de octubre de 2026', 'Versión 2.5', '2 min de lectura'],
		ilustracion: 'portapapeles',
		indice: [
			['s1', '1. Responsable'],
			['s2', '2. Qué datos juntamos'],
			['s2b', '3. Visitas a las tiendas'],
			['s3', '4. Para qué los usamos'],
			['s4', '5. Con quién los compartimos'],
			['s5', '6. Transferencias internacionales'],
			['s6', '7. Tus derechos'],
			['s7', '8. Cuánto tiempo los guardamos'],
			['s8', '9. Seguridad'],
			['s9', '10. Inteligencia artificial'],
			['s10', '11. Cambios']
		],
		bloques: [
			{
				t: 'resumen',
				items: [
					'Usamos tus datos para prestarte el servicio, facturarte y mantener tu cuenta segura. Nada más.',
					'No vendemos datos personales, ni los tuyos ni los de tus clientes.',
					'Podés pedir una copia, corregirlos o borrarlos cuando quieras, gratis.'
				]
			},
			{
				t: 'seccion',
				id: 's1',
				n: '01',
				titulo: 'Responsable',
				cuerpo: [
					{
						p: 'Nahuel Santillán, CUIT 20-44859234-1, con domicilio en Mar del Plata, provincia de Buenos Aires, es responsable de la base de datos de usuarios de Tu Local Web.'
					}
				]
			},
			{
				t: 'seccion',
				id: 's2',
				n: '02',
				titulo: 'Qué datos juntamos',
				cuerpo: [
					{
						p: 'Solo los que necesitamos para que funcione tu tienda.'
					},
					{
						ul: [
							'De tu cuenta: nombre, email, teléfono, CUIT y datos de facturación.',
							'De uso: páginas del panel que visitás, dispositivo y dirección IP, para seguridad y para mejorar el producto.',
							'De tus clientes: los que vos cargás o recibís en tus ventas. Ahí sos el responsable y Tu Local Web actúa como encargado (ver Tratamiento de datos).',
							'De quienes visitan las tiendas: sólo cuántas veces se vio cada página, sin cookies y sin guardar la IP ni el navegador (ver Visitas a las tiendas).'
						]
					}
				]
			},
			{
				t: 'seccion',
				id: 's2b',
				n: '03',
				titulo: 'Visitas a las tiendas',
				cuerpo: [
					{
						p: 'Para que cada comercio sepa cuánta gente entra a su tienda, Tu Local Web cuenta las páginas vistas desde su propio servidor, sin cookies, sin guardar nada en el navegador y sin servicios de terceros.'
					},
					{
						ul: [
							'Para no contar dos veces a la misma persona en un día, calculamos un código (hash) con la IP, el navegador, la tienda y una clave al azar que cambia todos los días y nunca se guarda. Con la clave borrada, ese código no permite saber quién sos ni unir tus visitas de distintos días.',
							'No guardamos la IP ni el navegador. Lo que queda es el total por tienda, día y página (por ejemplo, «el mate se vio 40 veces el martes»).',
							'El código del día se borra al día siguiente; los totales, a los 400 días.',
							'No cuentan los buscadores ni los robots, las vistas previas de links (WhatsApp, Facebook, Telegram) ni la dueña de la tienda mirando la suya con su cuenta.'
						]
					}
				]
			},
			{
				t: 'seccion',
				id: 's3',
				n: '04',
				titulo: 'Para qué los usamos',
				cuerpo: [
					{
						p: 'Para prestarte el servicio, cobrarte y facturarte, darte soporte, prevenir fraudes y accesos indebidos, y mandarte novedades del producto. Las comunicaciones comerciales las podés desactivar desde Notificaciones.'
					}
				]
			},
			{
				t: 'seccion',
				id: 's4',
				n: '05',
				titulo: 'Con quién los compartimos',
				cuerpo: [
					{
						p: 'Con proveedores que nos ayudan a operar (alojamiento en la nube, envío de emails, procesamiento de pagos), con las integraciones que vos activás, y con autoridades cuando una ley o un juez lo exige. Todos firman acuerdos de confidencialidad.'
					}
				]
			},
			{
				t: 'seccion',
				id: 's5',
				n: '06',
				titulo: 'Transferencias internacionales',
				cuerpo: [
					{
						p: 'Algunos proveedores alojan datos fuera de la Argentina, principalmente en Brasil y Estados Unidos. En esos casos usamos cláusulas contractuales que garantizan un nivel de protección adecuado.'
					}
				]
			},
			{
				t: 'seccion',
				id: 's6',
				n: '07',
				titulo: 'Tus derechos',
				cuerpo: [
					{
						p: 'Podés acceder a tus datos, rectificarlos, actualizarlos o pedir que los borremos escribiendo a privacidad@tulocalweb.com.ar o desde Configuración › Mis datos. El acceso es gratuito en intervalos de seis meses.'
					},
					{
						p: 'LA AGENCIA DE ACCESO A LA INFORMACIÓN PÚBLICA, en su carácter de Órgano de Control de la Ley N° 25.326, tiene la atribución de atender las denuncias y reclamos que interpongan quienes resulten afectados en sus derechos por incumplimiento de las normas vigentes en materia de protección de datos personales.'
					}
				]
			},
			{
				t: 'seccion',
				id: 's7',
				n: '08',
				titulo: 'Cuánto tiempo los guardamos',
				cuerpo: [
					{
						p: 'Mientras tu cuenta esté activa. Si la cerrás, borramos tu tienda a los 90 días, salvo la información contable que la ley nos obliga a conservar durante 10 años.'
					}
				]
			},
			{
				t: 'seccion',
				id: 's8',
				n: '09',
				titulo: 'Seguridad',
				cuerpo: [
					{
						p: 'Ciframos los datos en tránsito y en reposo, hacemos copias de seguridad diarias y limitamos el acceso interno. Más detalle en la página de Seguridad.'
					}
				]
			},
			{
				t: 'seccion',
				id: 's9',
				n: '10',
				titulo: 'Inteligencia artificial',
				cuerpo: [
					{
						p: 'Cuando Tincho responde un mensaje, el texto se procesa solo para generar esa respuesta. Trabajamos con proveedores de IA que no retienen ni usan esos datos para entrenar sus modelos.'
					}
				]
			},
			{
				t: 'seccion',
				id: 's10',
				n: '11',
				titulo: 'Cambios',
				cuerpo: [
					{
						p: 'Si cambiamos esta política te avisamos por email y en el panel. La versión vigente siempre está en esta página.'
					}
				]
			}
		],
		cierre: {
			titulo: '¿Te quedó alguna duda?',
			texto:
				'Escribinos por WhatsApp al +54 223 592-6772 o a legales@tulocalweb.com.ar y te respondemos en 48 horas hábiles.'
		}
	},
	{
		slug: 'cookies',
		titulo: 'Política de cookies',
		corto: 'Política de cookies',
		miga: 'Cookies',
		descripcion: 'Legales de Tu Local Web, claros y sin letra chica.',
		chips: ['Vigente desde el 30 de octubre de 2026', 'Versión 1.4', '2 min de lectura'],
		ilustracion: 'portapapeles',
		indice: [
			['s1', '1. Qué son las cookies'],
			['s2', '2. Qué cookies usamos'],
			['s3', '3. Cómo gestionarlas'],
			['s4', '4. Cookies en las tiendas'],
			['s5', '5. Visitas medidas sin cookies']
		],
		bloques: [
			{
				t: 'resumen',
				items: [
					'Usamos cookies necesarias para que el sitio funcione; esas no se pueden apagar.',
					'Las analíticas y de marketing solo se activan si las aceptás.',
					'Podés cambiar tu elección cuando quieras desde el botón de abajo o el pie de la página.',
					'Las visitas a las tiendas las contamos sin cookies: no hace falta aceptar nada.'
				]
			},
			{
				t: 'seccion',
				id: 's1',
				n: '01',
				titulo: 'Qué son las cookies',
				cuerpo: [
					{
						p: 'Son pequeños archivos que el navegador guarda cuando visitás un sitio. Sirven para recordar tu sesión, tus preferencias y, si lo permitís, para medir cómo se usa Tu Local Web.'
					}
				]
			},
			{
				t: 'seccion',
				id: 's2',
				n: '02',
				titulo: 'Qué cookies usamos',
				cuerpo: [
					{
						p: 'Las agrupamos en tres categorías. La tabla de abajo detalla cada una.'
					}
				]
			},
			{
				t: 'cookies',
				filas: [
					{
						nombre: '__Host-tlw_sesion',
						color: '#2F8F63',
						tipo: 'Necesaria',
						para: 'Mantiene tu sesión iniciada',
						dura: '12 meses'
					},
					{
						nombre: 'tlw-cookies',
						color: '#2F8F63',
						tipo: 'Necesaria',
						para: 'Guarda tu elección de cookies',
						dura: '12 meses'
					},
					{
						nombre: '_ga',
						color: '#3E6BFD',
						tipo: 'Analítica',
						para: 'Google Analytics: visitas agregadas',
						dura: '13 meses'
					},
					{
						nombre: 'ph_*',
						color: '#3E6BFD',
						tipo: 'Analítica',
						para: 'Uso anónimo del panel',
						dura: '12 meses'
					},
					{
						nombre: '_fbp',
						color: '#C24E25',
						tipo: 'Marketing',
						para: 'Meta: medir campañas',
						dura: '3 meses'
					},
					{
						nombre: '_gcl_au',
						color: '#C24E25',
						tipo: 'Marketing',
						para: 'Google Ads: conversiones',
						dura: '3 meses'
					}
				]
			},
			{
				t: 'configurar-cookies'
			},
			{
				t: 'seccion',
				id: 's3',
				n: '03',
				titulo: 'Cómo gestionarlas',
				cuerpo: [
					{
						p: 'Podés aceptar o rechazar las analíticas y de marketing desde el panel de cookies. También podés borrarlas o bloquearlas desde la configuración de tu navegador; si bloqueás las necesarias, algunas partes del sitio pueden dejar de funcionar.'
					}
				]
			},
			{
				t: 'seccion',
				id: 's4',
				n: '04',
				titulo: 'Cookies en las tiendas',
				cuerpo: [
					{
						p: 'Cada tienda creada con Tu Local Web puede sumar sus propios píxeles (por ejemplo, de Meta o Google) para medir sus campañas. En ese caso el responsable es el comercio. Tu Local Web incluye en todas las tiendas un aviso de cookies configurable para que cumplan la normativa.'
					}
				]
			},
			{
				t: 'seccion',
				id: 's5',
				n: '05',
				titulo: 'Visitas medidas sin cookies',
				cuerpo: [
					{
						p: 'Tu Local Web cuenta cuántas personas entran a cada tienda desde su propio servidor, sin cookies ni nada guardado en el navegador, y sin scripts de terceros. Por eso esa medición no aparece en el aviso de cookies ni necesita tu permiso.'
					},
					{
						p: 'Para no contar dos veces a la misma persona en el día usamos un código calculado con la IP, el navegador, la tienda y una clave al azar que cambia cada día y no se guarda. No guardamos la IP ni el navegador, y el código se borra al día siguiente: no sirve para identificarte ni para seguirte entre días o tiendas. Más detalle en la Política de privacidad.'
					}
				]
			}
		],
		cierre: {
			titulo: '¿Te quedó alguna duda?',
			texto:
				'Escribinos por WhatsApp al +54 223 592-6772 o a legales@tulocalweb.com.ar y te respondemos en 48 horas hábiles.'
		}
	},
	{
		slug: 'arrepentimiento',
		titulo: 'Botón de arrepentimiento',
		corto: 'Botón de arrepentimiento',
		miga: 'Arrepentimiento',
		descripcion: 'Legales de Tu Local Web, claros y sin letra chica.',
		chips: ['Vigente desde el 1 de octubre de 2026', 'Versión 1.1', '2 min de lectura'],
		ilustracion: 'portapapeles',
		indice: [
			['s1', '1. Qué es'],
			['s2', '2. Plazo'],
			['s3', '3. Cómo sigue'],
			['s4', '4. Si compraste en una tienda hecha con Tu Local Web']
		],
		bloques: [
			{
				t: 'arrepentimiento'
			},
			{
				t: 'resumen',
				items: [
					'Tenés 10 días corridos desde que contrataste para arrepentirte, sin dar explicaciones.',
					'No tiene costo y te devolvemos todo lo que pagaste, por el mismo medio de pago.',
					'Te damos un código de trámite en menos de 24 horas.'
				]
			},
			{
				t: 'seccion',
				id: 's1',
				n: '01',
				titulo: 'Qué es',
				cuerpo: [
					{
						p: 'Es tu derecho a revocar la contratación de un servicio hecho a distancia, previsto en la Ley 24.240 y la Resolución 424/2020 de la Secretaría de Comercio Interior. Este formulario es la forma más rápida de ejercerlo.'
					}
				]
			},
			{
				t: 'seccion',
				id: 's2',
				n: '02',
				titulo: 'Plazo',
				cuerpo: [
					{
						p: 'Podés usarlo dentro de los 10 días corridos desde que contrataste un plan pago de Tu Local Web o desde que se activó, lo que ocurra último.'
					}
				]
			},
			{
				t: 'seccion',
				id: 's3',
				n: '03',
				titulo: 'Cómo sigue',
				cuerpo: [
					{
						p: 'Cuando lo enviás te mandamos por email un código de identificación del trámite. Cancelamos el débito automático y te devolvemos el total pagado en un plazo máximo de 10 días hábiles, por el mismo medio que usaste para pagar.'
					}
				]
			},
			{
				t: 'seccion',
				id: 's4',
				n: '04',
				titulo: 'Si compraste en una tienda hecha con Tu Local Web',
				cuerpo: [
					{
						p: 'El vendedor es esa tienda. Cada tienda tiene su propio botón de arrepentimiento en el pie de página, y es el comercio quien procesa la devolución. Si no te responde en 10 días, escribinos y te ayudamos a resolverlo.'
					}
				]
			}
		],
		cierre: {
			titulo: '¿Te quedó alguna duda?',
			texto:
				'Escribinos por WhatsApp al +54 223 592-6772 o a legales@tulocalweb.com.ar y te respondemos en 48 horas hábiles.'
		}
	},
	{
		slug: 'defensa-del-consumidor',
		titulo: 'Defensa de las y los consumidores',
		corto: 'Defensa del consumidor',
		miga: 'Consumidor',
		descripcion: 'Legales de Tu Local Web, claros y sin letra chica.',
		chips: ['Vigente desde el 1 de octubre de 2026', 'Versión 1.0', '2 min de lectura'],
		ilustracion: 'portapapeles',
		indice: [
			['s1', '1. Atención al cliente'],
			['s2', '2. Si no estás conforme con la respuesta'],
			['s3', '3. Datos del proveedor'],
			['s4', '4. Compras en tiendas creadas con Tu Local Web']
		],
		bloques: [
			{
				t: 'resumen',
				items: [
					'Respondemos todos los reclamos en menos de 48 horas hábiles.',
					'Si no quedás conforme, podés ir a la Ventanilla Única Federal o a Defensa del Consumidor de tu ciudad.',
					'Si el problema es con una tienda hecha con Tu Local Web, mediamos con el comercio.'
				]
			},
			{
				t: 'seccion',
				id: 's1',
				n: '01',
				titulo: 'Atención al cliente',
				cuerpo: [
					{
						p: 'Podés escribirnos por WhatsApp al +54 223 592-6772 o por email a soporte@tulocalweb.com.ar. Te respondemos en menos de 48 horas hábiles y te damos un número de reclamo.'
					}
				]
			},
			{
				t: 'seccion',
				id: 's2',
				n: '02',
				titulo: 'Si no estás conforme con la respuesta',
				cuerpo: [
					{
						p: 'Podés iniciar un reclamo gratuito en la Ventanilla Única Federal de Defensa del Consumidor o ante la oficina de Defensa del Consumidor de tu jurisdicción. También podés pedir una conciliación previa en el COPREC.'
					}
				]
			},
			{
				t: 'organismos',
				items: [
					{
						href: 'https://www.argentina.gob.ar/produccion/defensadelconsumidor/formulario',
						icono: phBankFill,
						titulo: 'Ventanilla Única Federal',
						texto: 'Reclamo online en argentina.gob.ar'
					},
					{
						href: 'https://buenosaires.gob.ar/defensaconsumidor',
						icono: phBuildingsFill,
						titulo: 'Defensa del Consumidor CABA',
						texto: 'Si vivís en la Ciudad de Buenos Aires'
					},
					{
						href: '/contacto',
						icono: phChatsCircleFill,
						titulo: 'Atención Tu Local Web',
						texto: 'Respuesta en menos de 48 h hábiles'
					}
				]
			},
			{
				t: 'seccion',
				id: 's3',
				n: '03',
				titulo: 'Datos del proveedor',
				cuerpo: [
					{
						p: 'Nahuel Santillán · CUIT 20-44859234-1 · Mar del Plata, provincia de Buenos Aires · legales@tulocalweb.com.ar · WhatsApp +54 223 592-6772.'
					}
				]
			},
			{
				t: 'seccion',
				id: 's4',
				n: '04',
				titulo: 'Compras en tiendas creadas con Tu Local Web',
				cuerpo: [
					{
						p: 'Cuando comprás en una tienda que usa Tu Local Web, el vendedor responsable es ese comercio y sus datos figuran en el pie de su tienda. Si no te responde, escribinos con el número de pedido: intervenimos y, si corresponde, suspendemos la tienda.'
					}
				]
			}
		],
		cierre: {
			titulo: '¿Te quedó alguna duda?',
			texto:
				'Escribinos por WhatsApp al +54 223 592-6772 o a legales@tulocalweb.com.ar y te respondemos en 48 horas hábiles.'
		}
	},
	{
		slug: 'tratamiento-de-datos',
		titulo: 'Acuerdo de tratamiento de datos',
		corto: 'Tratamiento de datos (para tiendas)',
		miga: 'Tratamiento de datos',
		descripcion: 'Legales de Tu Local Web, claros y sin letra chica.',
		chips: ['Vigente desde el 1 de octubre de 2026', 'Versión 1.2', '2 min de lectura'],
		ilustracion: 'portapapeles',
		indice: [
			['s1', '1. Roles'],
			['s2', '2. Qué datos tratamos'],
			['s3', '3. Instrucciones'],
			['s4', '4. Confidencialidad'],
			['s5', '5. Medidas de seguridad'],
			['s6', '6. Subencargados'],
			['s7', '7. Incidentes'],
			['s8', '8. Fin del servicio'],
			['s9', '9. Auditoría']
		],
		bloques: [
			{
				t: 'resumen',
				items: [
					'Tus clientes son tuyos: vos sos el responsable de sus datos y Tu Local Web solo los trata para prestarte el servicio.',
					'Te avisamos en menos de 48 horas si hay un incidente de seguridad que los afecte.',
					'Cuando cerrás tu cuenta te los devolvemos y los borramos en 90 días.'
				]
			},
			{
				t: 'seccion',
				id: 's1',
				n: '01',
				titulo: 'Roles',
				cuerpo: [
					{
						p: 'En este acuerdo, el comercio que usa Tu Local Web es el responsable de los datos de sus compradores y Tu Local Web es el encargado del tratamiento, en los términos del artículo 25 de la Ley 25.326.'
					}
				]
			},
			{
				t: 'seccion',
				id: 's2',
				n: '02',
				titulo: 'Qué datos tratamos',
				cuerpo: [
					{
						p: 'Nombre, documento, email, teléfono, dirección de entrega, historial de compras y mensajes de tus clientes, en la medida en que vos los recibas o cargues en Tu Local Web.'
					}
				]
			},
			{
				t: 'seccion',
				id: 's3',
				n: '03',
				titulo: 'Instrucciones',
				cuerpo: [
					{
						p: 'Tratamos esos datos solo para prestarte el servicio y siguiendo tus instrucciones, que son las configuraciones que elegís en el panel. No los usamos para fines propios ni los cedemos a terceros.'
					}
				]
			},
			{
				t: 'seccion',
				id: 's4',
				n: '04',
				titulo: 'Confidencialidad',
				cuerpo: [
					{
						p: 'Todas las personas de Tu Local Web con acceso a datos firmaron un compromiso de confidencialidad y reciben capacitación periódica.'
					}
				]
			},
			{
				t: 'seccion',
				id: 's5',
				n: '05',
				titulo: 'Medidas de seguridad',
				cuerpo: [
					{
						p: 'Aplicamos, como mínimo:'
					},
					{
						ul: [
							'Cifrado TLS en tránsito y AES-256 en reposo.',
							'Verificación en dos pasos obligatoria para el equipo de Tu Local Web.',
							'Copias de seguridad diarias con retención de 14 días.',
							'Registro de accesos y revisión trimestral de permisos.'
						]
					}
				]
			},
			{
				t: 'seccion',
				id: 's6',
				n: '06',
				titulo: 'Subencargados',
				cuerpo: [
					{
						p: 'Usamos proveedores para alojamiento en la nube, envío de emails y mensajes de WhatsApp, y procesamiento con IA sin retención. Publicamos la lista vigente y te avisamos con 30 días de anticipación si sumamos uno nuevo; podés oponerte por motivos fundados.'
					},
					{
						p: 'Para mostrarte desde dónde está abierta cada sesión de tu cuenta, la ubicación aproximada sale de la dirección IP con la base IP Geolocation by DB-IP (db-ip.com, licencia CC BY 4.0), que corre en nuestros servidores: tu IP no se manda a nadie.'
					}
				]
			},
			{
				t: 'seccion',
				id: 's7',
				n: '07',
				titulo: 'Incidentes',
				cuerpo: [
					{
						p: 'Si detectamos un incidente que afecte datos de tus clientes, te avisamos en menos de 48 horas, con lo que sabemos y lo que estamos haciendo, para que puedas cumplir con tus propias obligaciones.'
					}
				]
			},
			{
				t: 'seccion',
				id: 's8',
				n: '08',
				titulo: 'Fin del servicio',
				cuerpo: [
					{
						p: 'Cuando cerrás tu cuenta podés exportar todos tus datos. A los 90 días los borramos de nuestros sistemas, salvo lo que la ley nos obligue a conservar.'
					}
				]
			},
			{
				t: 'seccion',
				id: 's9',
				n: '09',
				titulo: 'Auditoría',
				cuerpo: [
					{
						p: 'A pedido, te enviamos una vez por año un informe de las medidas de seguridad vigentes.'
					}
				]
			}
		],
		cierre: {
			titulo: '¿Te quedó alguna duda?',
			texto:
				'Escribinos por WhatsApp al +54 223 592-6772 o a legales@tulocalweb.com.ar y te respondemos en 48 horas hábiles.'
		}
	},
	{
		slug: 'uso-aceptable',
		titulo: 'Uso aceptable y productos prohibidos',
		corto: 'Uso aceptable',
		miga: 'Uso aceptable',
		descripcion: 'Legales de Tu Local Web, claros y sin letra chica.',
		chips: ['Vigente desde el 1 de octubre de 2026', 'Versión 2.0', '2 min de lectura'],
		ilustracion: 'portapapeles',
		indice: [
			['s1', '1. Principio general'],
			['s2', '2. Productos prohibidos'],
			['s3', '3. Productos con requisitos'],
			['s4', '4. Conductas prohibidas'],
			['s5', '5. Qué pasa si se incumple'],
			['s6', '6. Denunciar una tienda']
		],
		bloques: [
			{
				t: 'resumen',
				items: [
					'Podés vender casi todo lo que es legal en la Argentina.',
					'Algunos productos necesitan habilitación (alcohol, cosméticos, alimentos) y otros están prohibidos.',
					'Si algo no está claro, consultanos antes de publicarlo.'
				]
			},
			{
				t: 'seccion',
				id: 's1',
				n: '01',
				titulo: 'Principio general',
				cuerpo: [
					{
						p: 'Tu Local Web es para negocios legales y honestos. No podés usarlo para vender productos ilegales, engañar a compradores ni dañar a otros usuarios o a la plataforma.'
					}
				]
			},
			{
				t: 'seccion',
				id: 's2',
				n: '02',
				titulo: 'Productos prohibidos',
				cuerpo: [
					{
						p: 'No se pueden publicar:'
					},
					{
						ul: [
							'Armas, municiones y explosivos.',
							'Drogas, estupefacientes y medicamentos de venta bajo receta.',
							'Productos falsificados, réplicas o que infrinjan marcas registradas.',
							'Animales vivos de especies protegidas.',
							'Contenido sexual explícito o servicios sexuales.',
							'Apuestas, rifas o juegos de azar no autorizados.',
							'Esquemas piramidales o inversiones no reguladas.',
							'Documentos de identidad, bases de datos o datos personales.'
						]
					}
				]
			},
			{
				t: 'seccion',
				id: 's3',
				n: '03',
				titulo: 'Productos con requisitos',
				cuerpo: [
					{
						p: 'Se pueden vender cumpliendo la normativa correspondiente:'
					},
					{
						ul: [
							'Bebidas alcohólicas: solo a mayores de 18, con verificación de edad.',
							'Tabaco y vapeadores, según la regulación vigente.',
							'Cosméticos, suplementos y productos médicos: con registro ANMAT.',
							'Alimentos envasados: con RNE y RNPA.',
							'Pirotecnia: con habilitación de RENAR/ANMaC.'
						]
					}
				]
			},
			{
				t: 'seccion',
				id: 's4',
				n: '04',
				titulo: 'Conductas prohibidas',
				cuerpo: [
					{
						p: 'Tampoco está permitido:'
					},
					{
						ul: [
							'Mandar mensajes masivos de WhatsApp a quien no aceptó recibirlos, o incumplir las políticas de Meta.',
							'Hacer phishing, suplantar marcas o personas.',
							'Revender el servicio o usarlo para atacar o sobrecargar otros sistemas.'
						]
					}
				]
			},
			{
				t: 'seccion',
				id: 's5',
				n: '05',
				titulo: 'Qué pasa si se incumple',
				cuerpo: [
					{
						p: 'Según la gravedad, te avisamos para que lo corrijas, ocultamos la publicación, suspendemos la tienda o damos de baja la cuenta. Si hay un delito, colaboramos con las autoridades.'
					}
				]
			},
			{
				t: 'seccion',
				id: 's6',
				n: '06',
				titulo: 'Denunciar una tienda',
				cuerpo: [
					{
						p: 'Si ves una tienda que incumple estas reglas, escribinos a abuso@tulocalweb.com.ar con el link. Revisamos todas las denuncias en menos de 72 horas.'
					}
				]
			}
		],
		cierre: {
			titulo: '¿Te quedó alguna duda?',
			texto:
				'Escribinos por WhatsApp al +54 223 592-6772 o a legales@tulocalweb.com.ar y te respondemos en 48 horas hábiles.'
		}
	},
	{
		slug: 'reembolsos',
		titulo: 'Política de reembolsos de planes',
		corto: 'Reembolsos de planes',
		miga: 'Reembolsos',
		descripcion: 'Legales de Tu Local Web, claros y sin letra chica.',
		chips: ['Vigente desde el 1 de octubre de 2026', 'Versión 1.4', '2 min de lectura'],
		ilustracion: 'portapapeles',
		indice: [
			['s1', '1. Prueba gratis'],
			['s2', '2. Arrepentimiento'],
			['s3', '3. Planes mensuales'],
			['s4', '4. Planes anuales'],
			['s5', '5. Créditos por disponibilidad'],
			['s6', '6. Cómo pedirlo']
		],
		bloques: [
			{
				t: 'resumen',
				items: [
					'La prueba de 14 días es gratis: no te cobramos nada si cancelás antes.',
					'En los primeros 10 días de un plan pago te devolvemos todo.',
					'Si el servicio falla más de lo prometido, te acreditamos parte del plan automáticamente.'
				]
			},
			{
				t: 'seccion',
				id: 's1',
				n: '01',
				titulo: 'Prueba gratis',
				cuerpo: [
					{
						p: 'Durante los 14 días de prueba no se cobra nada. Si no elegís un plan, la tienda queda en pausa y conservamos todo por 30 días.'
					}
				]
			},
			{
				t: 'seccion',
				id: 's2',
				n: '02',
				titulo: 'Arrepentimiento',
				cuerpo: [
					{
						p: 'Dentro de los 10 días corridos desde que contrataste un plan, te devolvemos el 100 % de lo pagado. Podés pedirlo desde el Botón de arrepentimiento.'
					}
				]
			},
			{
				t: 'seccion',
				id: 's3',
				n: '03',
				titulo: 'Planes mensuales',
				cuerpo: [
					{
						p: 'No devolvemos meses parciales. Si cancelás, seguís usando Tu Local Web hasta el final del mes ya pagado y no se te vuelve a cobrar.'
					}
				]
			},
			{
				t: 'seccion',
				id: 's4',
				n: '04',
				titulo: 'Planes anuales',
				cuerpo: [
					{
						p: 'Si cancelás dentro de los primeros 30 días, te devolvemos el total. Después, te reintegramos los meses completos que no usaste, calculados a precio mensual.'
					}
				]
			},
			{
				t: 'seccion',
				id: 's5',
				n: '05',
				titulo: 'Créditos por disponibilidad',
				cuerpo: [
					{
						p: 'Si la disponibilidad mensual de tu tienda queda por debajo del 99,9 %, te acreditamos automáticamente un porcentaje del plan en la próxima factura, según la tabla de abajo.'
					}
				]
			},
			{
				t: 'creditos',
				items: [
					{
						etiqueta: 'Disponibilidad mensual',
						umbral: 'Menos de 99,9 %',
						credito: '10 %',
						pie: 'de crédito'
					},
					{
						etiqueta: 'Disponibilidad mensual',
						umbral: 'Menos de 99 %',
						credito: '25 %',
						pie: 'de crédito'
					},
					{
						etiqueta: 'Disponibilidad mensual',
						umbral: 'Menos de 95 %',
						credito: '50 %',
						pie: 'de crédito'
					}
				]
			},
			{
				t: 'seccion',
				id: 's6',
				n: '06',
				titulo: 'Cómo pedirlo',
				cuerpo: [
					{
						p: 'Desde Configuración › Plan › Pedir reembolso, o escribiendo a facturacion@tulocalweb.com.ar. Lo procesamos en un máximo de 10 días hábiles y lo devolvemos por el mismo medio de pago.'
					}
				]
			}
		],
		cierre: {
			titulo: '¿Te quedó alguna duda?',
			texto:
				'Escribinos por WhatsApp al +54 223 592-6772 o a legales@tulocalweb.com.ar y te respondemos en 48 horas hábiles.'
		}
	}
];
