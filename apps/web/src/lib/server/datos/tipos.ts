/**
 * Lo que las páginas pueden pedir. Las dos fuentes (sandbox de la demo y API Go) implementan esta
 * misma interfaz con los tipos del contrato (packages/contrato/openapi.yaml), y los casos de
 * packages/contrato/casos corren contra las dos.
 */
import type { components } from '@mostrador/contrato/esquema';
import type { RolInvitable } from '#lib/panel/permisos.ts';
import type { AjusteIntegracion, EstadoIntegraciones } from '#lib/panel/integraciones.ts';
import type { Importacion, ImportacionNueva, PedidoHistorial } from '../importacion/fuente.ts';

/** Por qué no se pudo volver de Tiendanube (el panel lo traduce). */
export type ErrorVueltaTN =
	'estado' | 'canje' | 'invalido' | 'sin-fuente' | 'activa' | 'sin-cifrado' | 'sin-tiendanube';

export type Rubro = components['schemas']['Rubro'];
export type Tienda = components['schemas']['Tienda'];
export type DisponibilidadDireccion = components['schemas']['DisponibilidadDireccion'];
/** Cómo terminó un cambio de dirección: la nueva y la anterior, o por qué no se pudo. */
export type CambioDireccion =
	| { ok: true; slug: string; anterior: string }
	| { ok: false; motivo: 'invalido' | 'reservado' | 'usado' | 'tope'; mensaje: string };
export type Plan = components['schemas']['Plan'];

export type Onboarding = components['schemas']['Onboarding'];
export type ResumenPanel = components['schemas']['ResumenPanel'];
export type Notificacion = components['schemas']['Notificacion'];
export type PaginaNotificaciones = components['schemas']['PaginaNotificaciones'];
export type Tablero = components['schemas']['Tablero'];
export type Periodo = Tablero['periodo'];
export type ContextoHornero = components['schemas']['ContextoHornero'];
export type Idea = 'carritos' | 'stock' | 'promo';
export type ListaPedidos = components['schemas']['ListaPedidos'];
export type PedidoDetalle = components['schemas']['PedidoDetalle'];
export type EstadoPedido = components['schemas']['EstadoPedido'];
export type AccionLote = 'preparar' | 'etiquetas' | 'facturar' | 'avisar';
export type FiltrosPedidos = import('../sandbox/pedidos.ts').FiltrosPedidos;
/** Resultado de una acción sobre un pedido. */
export type Resultado = 'ok' | 'no-existe' | 'no-se-puede';
export type ListaGenerica = components['schemas']['ListaGenerica'];
export type NombreLista = components['schemas']['NombreLista'];
/** El de la API: ProductoEditable manda variantes editables; la ficha devuelve las completas. */
export type Producto = Omit<components['schemas']['Producto'], 'variantes'> & {
	variantes: components['schemas']['Variante'][];
};
export type ProductoEditable = components['schemas']['ProductoEditable'];
export type FiltrosLista = import('../sandbox/listas.ts').FiltrosLista;
export type MovimientoStock = components['schemas']['MovimientoStock'];
export type ConversacionBreve = components['schemas']['ConversacionBreve'];
export type Conversacion = components['schemas']['Conversacion'];
export type Carritos = components['schemas']['Carritos'];
export type AutomaticosCarritos = components['schemas']['AutomaticosCarritos'];
export type PedidoNuevo = components['schemas']['PedidoNuevo'];
export type ListasPrecio = components['schemas']['ListasPrecio'];
export type Cupon = components['schemas']['Cupon'];
export type Marketing = components['schemas']['Marketing'];
export type Evento = components['schemas']['Evento'];
export type { FiltroActividad, PaginaActividad } from '#lib/panel/actividad.ts';
import type { FiltroActividad, PaginaActividad } from '#lib/panel/actividad.ts';
export type Finanzas = components['schemas']['Finanzas'];
export type MiTienda = components['schemas']['MiTienda'];
export type Tarjeta = components['schemas']['Tarjeta'];
export type Configuracion = components['schemas']['Configuracion'];
export type PlanCuenta = components['schemas']['PlanCuenta'];
export type Dominios = components['schemas']['Dominios'];
export type ConexionMP = components['schemas']['ConexionMP'];
export type ConexionML = components['schemas']['ConexionML'];
export type Medicion = components['schemas']['Medicion'];
export type Entregas = components['schemas']['Entregas'];
export type EstadoEntregas = components['schemas']['EstadoEntregas'];
export type PublicacionML = components['schemas']['PublicacionML'];
export type ArcaConfig = components['schemas']['ArcaConfig'];
export type ArcaAlta = components['schemas']['ArcaAlta'];
export type ArcaCertificado = components['schemas']['ArcaCertificado'];
export type ArcaAjustes = { puntoVenta?: number; facturarAlCobrar?: boolean };
export type ArcaPrueba = { ok: boolean; ultimo: number; detalle: string };
/**
 * El pedido de certificado (CSR en PEM). En la demo no hay clave que guardar: `demo` le dice al
 * panel que lo arme el navegador (lib/panel/csr.ts) con los datos ya validados.
 */
export type ArcaPedido = { pedido: string } | { demo: true } | { error: string };
export type PreferenciasAvisos = components['schemas']['PreferenciasAvisos'];
/** Push web del panel instalado (docs/decisiones.md §43): clave VAPID (vacía: apagado) y si esta sesión está suscripta. */
export type EstadoPushWeb = components['schemas']['EstadoPushWeb'];
export type SuscripcionWeb = components['schemas']['SuscripcionWeb'];
export type MisDatos = components['schemas']['MisDatos'];
export type TipoExport =
	'productos' | 'clientes' | 'pedidos' | 'facturas' | 'resenas' | 'conversaciones';
export type Acceso = components['schemas']['Acceso'];
export type Negocio = components['schemas']['Negocio'];
export type Miembro = components['schemas']['Miembro'];
export type IdPlan = Configuracion['plan']['actual'];
export type DisenoTienda = components['schemas']['DisenoTienda'];
export type ParaHacer = components['schemas']['ParaHacer'];
export type Campana = components['schemas']['Campana'];
export type Canales = components['schemas']['Canales'];
export type Informes = components['schemas']['Informes'];
export type Cupones = components['schemas']['Cupones'];
export type CuponNuevo = components['schemas']['CuponNuevo'];
export type AccionCupon = 'pausar' | 'activar' | 'duplicar' | 'minimo' | 'deshacer-minimo';
export type AccionRelampago = 'sumar-hora' | 'liberar' | 'terminar' | 'reanudar';
export type Devolucion = components['schemas']['Devolucion'];
export type ResenasYPreguntas = components['schemas']['ResenasYPreguntas'];
export type AjustesTurnos = components['schemas']['AjustesTurnos'];
export type Servicio = components['schemas']['Servicio'];
export type ServicioEditable = components['schemas']['ServicioEditable'];
export type Profesional = components['schemas']['Profesional'];
export type ProfesionalEditable = components['schemas']['ProfesionalEditable'];
export type Turno = components['schemas']['Turno'];
export type TurnoNuevo = components['schemas']['TurnoNuevo'];
export type EstadoTurno = components['schemas']['EstadoTurno'];
export type PanelSuscripciones = components['schemas']['PanelSuscripciones'];
export type PlanSuscripcion = components['schemas']['PlanSuscripcion'];
export type PlanSuscripcionEditable = components['schemas']['PlanSuscripcionEditable'];
export type Suscriptor = components['schemas']['Suscriptor'];

export interface Datos {
	rubros(): Promise<Rubro[]>;
	/** La tienda de la sesión. */
	tienda(): Promise<Tienda>;
	/** Lo guardado del onboarding, para retomarlo donde quedó, y si ya terminó. */
	onboarding(): Promise<Onboarding>;
	/** Lo elegido en el onboarding (y si terminó). */
	guardarOnboarding(o: Onboarding): Promise<void>;
	/** Si la tienda puede pasar a tulocalweb.com.ar/t/<slug> (decisiones §38). */
	disponibilidadDireccion(slug: string): Promise<DisponibilidadDireccion>;
	/** Cambia la dirección; la anterior sigue andando (301) para siempre. */
	cambiarDireccion(slug: string): Promise<CambioDireccion>;
	/** Insignias del menú, notificaciones, canales y prueba: lo que muestra el armazón del panel. */
	resumen(): Promise<ResumenPanel>;
	/** La campanita (§45): las notificaciones de la persona, de la más nueva a la más vieja. */
	notificaciones(cursor?: string): Promise<PaginaNotificaciones>;
	marcarNotificacionLeida(id: string): Promise<void>;
	marcarNotificacionesLeidas(): Promise<void>;
	/** El Resumen (dashboard) del período. */
	tablero(periodo: Periodo): Promise<Tablero>;
	/** La meta de ventas del mes, en pesos (0 la saca). */
	guardarMeta(objetivo: number): Promise<void>;
	ocultarAviso(aviso: 'certificado'): Promise<void>;
	/** Manda el recordatorio a los carritos abandonados; devuelve cuántos mensajes salieron. */
	recordarCarritos(): Promise<number>;
	/** Lo que Tincho sabe de la tienda para responder. */
	contextoHornero(): Promise<ContextoHornero>;
	/** Aprueba (o deshace) una idea de Tincho. */
	marcarIdea(idea: Idea, hecha: boolean): Promise<void>;
	listarPedidos(f: FiltrosPedidos): Promise<ListaPedidos>;
	/** null si no existe. */
	pedido(numero: number): Promise<PedidoDetalle | null>;
	cambiarEstado(numero: number, estado: EstadoPedido): Promise<Resultado>;
	pedidosEnLote(
		numeros: number[],
		accion: AccionLote
	): Promise<{ hechos: number; cambios: { numero: number; antes: EstadoPedido }[] }>;
	/** Recordatorio de pago por WhatsApp. */
	recordarPago(numero: number): Promise<Resultado>;
	/** Motor de listas: Clientes, Productos, Inventario, Facturación y Envíos. */
	listar(lista: NombreLista, f: FiltrosLista): Promise<ListaGenerica>;
	/** Una acción sobre filas de una lista (o la sugerencia de Tincho); devuelve cuántas cambiaron. */
	accionLista(lista: NombreLista, ids: string[], accion: string, valor?: unknown): Promise<number>;
	producto(id: string): Promise<Producto | null>;
	/** Crea (id null) o guarda un producto. null si el id no existe. */
	guardarProducto(id: string | null, p: ProductoEditable): Promise<Producto | null>;
	/** Borra un producto. 'no-se-puede': está publicado en Mercado Libre. */
	borrarProducto(id: string): Promise<Resultado>;
	/** Una copia pausada (sin stock ni SKU). null si el id no existe. */
	duplicarProducto(id: string): Promise<Producto | null>;
	/** Movimientos de stock de hoy (ventas, ajustes, conteos, ingresos). */
	movimientosHoy(): Promise<MovimientoStock[]>;
	bandeja(): Promise<{ conversaciones: ConversacionBreve[]; atendidasHoy: number }>;
	conversacion(id: string): Promise<Conversacion | null>;
	/** Mandar un mensaje, tomar o devolver la conversación, o marcarla leída. */
	responder(
		id: string,
		c: { texto?: string; tomada?: boolean; leida?: boolean }
	): Promise<Resultado>;
	carritos(): Promise<Carritos>;
	guardarAutomaticos(a: AutomaticosCarritos): Promise<void>;
	recordarCarrito(id: string): Promise<Resultado>;
	/** Un pedido nuevo (punto de venta o cargado a mano); descuenta stock. */
	crearPedido(
		n: PedidoNuevo
	): Promise<{ numero: number; total: number } | 'sin-stock' | 'invalido'>;
	devoluciones(): Promise<{ devoluciones: Devolucion[]; tasa: number }>;
	resolverDevolucion(id: string, accion: 'aprobar' | 'cambio' | 'deshacer'): Promise<Resultado>;
	resenas(): Promise<ResenasYPreguntas>;
	/** Responder una pregunta de Mercado Libre o de la tienda online (id "t-…"). */
	responderPregunta(id: string, respuesta: string): Promise<Resultado>;
	/** Ocultar o mostrar una reseña (no se borra), o responderla en público. */
	accionResena(
		id: string,
		accion: 'ocultar' | 'mostrar' | 'responder',
		respuesta?: string
	): Promise<Resultado>;
	listasPrecio(): Promise<ListasPrecio>;
	cupones(): Promise<Cupones>;
	marketing(): Promise<Marketing>;
	/** Actividad (§39): una página con los filtros. */
	actividad(f?: FiltroActividad): Promise<PaginaActividad>;
	/** Las filas del CSV de la actividad (la primera es el encabezado). */
	exportarActividad(f?: FiltroActividad): Promise<string[][]>;
	finanzas(): Promise<Finanzas>;
	miTienda(): Promise<MiTienda>;
	/** La tarjeta A6 para los paquetes (modelo, textos y QR del dorso). */
	tarjeta(): Promise<Tarjeta>;
	guardarTarjeta(t: Tarjeta): Promise<void>;
	configuracion(): Promise<Configuracion>;
	planCuenta(): Promise<PlanCuenta>;
	/** El hub de Integraciones (Mercado Pago se lee de mercadoPago()). */
	integraciones(): Promise<EstadoIntegraciones>;
	ajustarIntegracion(a: AjusteIntegracion): Promise<Resultado>;
	/** GA4, el píxel de Meta y los feeds de productos cargados (Google Shopping y Meta Ads). */
	medicion(): Promise<Medicion>;
	guardarMedicion(m: Medicion): Promise<Resultado>;
	/** Envíos y retiro: retiro y moto, con el estado de Correo y del envío gratis. */
	entregas(): Promise<EstadoEntregas>;
	/** Guarda retiro y moto; un dato que no va o todo apagado: error 400 con el motivo. */
	guardarEntregas(e: Entregas): Promise<void>;
	/** Mercado Libre como canal (OAuth de la tienda). */
	mercadoLibre(): Promise<ConexionML>;
	conectarMercadoLibre(): Promise<string | null>;
	volverMercadoLibre(codigo: string, estado: string): Promise<Resultado>;
	desconectarMercadoLibre(): Promise<void>;
	ajustarMercadoLibre(recargo: number, pausarSinStock: boolean): Promise<Resultado>;
	/** Publicar o despublicar un producto; el texto de error es el que da Mercado Libre. */
	publicarEnMercadoLibre(productoId: string): Promise<PublicacionML | { error: string }>;
	despublicarDeMercadoLibre(productoId: string): Promise<void>;
	/** La conexión con ARCA (facturación electrónica); nunca trae la clave. */
	arca(): Promise<ArcaConfig>;
	/** Genera la clave (queda en Tu Local Web) y devuelve el pedido de certificado para subir a ARCA. */
	arcaPedido(a: ArcaAlta): Promise<ArcaPedido>;
	/** El certificado que devolvió ARCA (PEM); el error es el texto para mostrar. */
	arcaCertificado(pem: string): Promise<ArcaCertificado | { error: string }>;
	/** Punto de venta y facturar al cobrar; null si se guardó. */
	arcaAjustes(a: ArcaAjustes): Promise<{ error: string } | null>;
	/** Pide el ticket y el último comprobante autorizado. */
	arcaProbar(): Promise<ArcaPrueba>;
	/** Mercado Pago de la tienda (marketplace por OAuth). */
	mercadoPago(): Promise<ConexionMP>;
	/** Adónde mandar a la persona para autorizar (o null si la aplicación no tiene credenciales). */
	conectarMercadoPago(): Promise<string | null>;
	volverMercadoPago(codigo: string, estado: string): Promise<Resultado>;
	desconectarMercadoPago(): Promise<void>;
	dominios(): Promise<Dominios>;
	conectarDominio(nombre: string): Promise<Resultado>;
	dominioPrincipal(nombre: string): Promise<Resultado>;
	quitarDominio(nombre: string): Promise<Resultado>;
	preferenciasAvisos(): Promise<PreferenciasAvisos>;
	guardarPreferenciasAvisos(p: PreferenciasAvisos): Promise<void>;
	/** Push web en este celular: la clave pública y si esta sesión ya se suscribió. */
	pushWeb(): Promise<EstadoPushWeb>;
	/** Guarda la suscripción del navegador atada a la sesión ('no-se-puede': apagado o inválida). */
	suscribirPushWeb(s: SuscripcionWeb): Promise<Resultado>;
	bajaPushWeb(): Promise<void>;
	/** «Mandame un aviso de prueba» ('no-se-puede': esta sesión no está suscripta). */
	probarPushWeb(): Promise<Resultado>;
	misDatos(): Promise<MisDatos>;
	exportar(tipo: TipoExport): Promise<(string | number | null)[][]>;
	/** Anota una descarga en "Exportaciones recientes" ("todo" es la exportación completa). */
	registrarExportacion(tipo: TipoExport | 'todo', formato: 'csv' | 'excel'): Promise<void>;
	borrarInactivos(): Promise<number>;
	frenarBorrado(): Promise<void>;
	cerrarCuenta(confirmacion: string): Promise<Resultado>;
	comprarPack(): Promise<void>;
	devolverPack(): Promise<Resultado>;
	cancelarPlan(motivo: number, accion: 'cancelar' | 'oferta'): Promise<{ hasta: string }>;
	reactivarPlan(): Promise<void>;
	/** La tienda que se ve, el rol con sus permisos y las tiendas de la persona. */
	acceso(): Promise<Acceso>;
	/** Cambia la tienda de la sesión (una persona con varias tiendas). */
	elegirTienda(id: string): Promise<Resultado>;
	cambiarRol(id: string, rol: RolInvitable): Promise<Resultado>;
	/** Quita a alguien del equipo (cierra sus sesiones en la tienda) o cancela la invitación. */
	quitarDelEquipo(id: string): Promise<Resultado>;
	/** Manda otra vez la invitación con un link nuevo. */
	reenviarInvitacion(id: string): Promise<Resultado>;
	guardarNegocio(n: Negocio): Promise<void>;
	fijarOpcion(grupo: 'seguridad' | 'avisos', clave: string, activo: boolean): Promise<Resultado>;
	cerrarSesion(id: string): Promise<Resultado>;
	/** Cierra todas las sesiones de la persona menos la actual; devuelve cuántas cerró. */
	cerrarOtrasSesiones(): Promise<number>;
	invitar(email: string, rol: RolInvitable): Promise<Miembro | 'repetido'>;
	/** Con el abono por Mercado Pago, un plan pago devuelve adónde ir a cargar la tarjeta. */
	cambiarPlan(plan: IdPlan, ciclo?: 'mes' | 'anio'): Promise<void | { url: string }>;
	guardarMiTienda(d: DisenoTienda): Promise<void>;
	publicarTienda(): Promise<void>;
	/** Pasa lo disponible al banco. */
	retirar(): Promise<{ monto: number; cuenta: string } | 'nada' | 'sin-cobros'>;
	deshacerEvento(id: string): Promise<Resultado>;
	paraHacer(): Promise<ParaHacer>;
	marcarTarea(id: string, hecha: boolean): Promise<Resultado>;
	crearCampana(c: {
		audiencia: 'todos' | 'vip' | 'riesgo' | 'categoria';
		mensaje: string;
		foto: boolean;
		boton: boolean;
		cupon: boolean;
		programar: boolean;
	}): Promise<Campana | 'sin-audiencia' | 'sin-whatsapp'>;
	canales(): Promise<Canales>;
	fijarCanal(id: string, activo: boolean): Promise<Resultado>;
	informes(): Promise<Informes>;
	crearCupon(n: CuponNuevo): Promise<Cupon | 'repetido'>;
	accionCupon(
		id: string,
		accion: AccionCupon,
		minimo?: number
	): Promise<Cupon | 'no-existe' | 'no-se-puede'>;
	descartarConsejo(): Promise<void>;
	fijarAutomatica(id: string, activa: boolean): Promise<Resultado>;
	accionRelampago(accion: AccionRelampago): Promise<Resultado>;
	/** Sube todos los precios ya, o lo programa para el lunes. */
	aumentarPrecios(
		pct: number,
		redondeo: 'no' | '90' | '500',
		programar: boolean
	): Promise<{ productos: number; para: string | null }>;
	deshacerAumento(): Promise<Resultado>;
	cancelarAumento(): Promise<Resultado>;
	/** Turnos con seña: ajustes (activo, seña, anticipación, días de vista y horario). */
	ajustesTurnos(): Promise<AjustesTurnos>;
	guardarAjustesTurnos(a: AjustesTurnos): Promise<AjustesTurnos>;
	/** Servicios que se reservan (activos e inactivos), por orden. */
	serviciosTurnos(): Promise<Servicio[]>;
	/** Crea (id null) o guarda un servicio; null si el id no existe. */
	guardarServicio(id: string | null, s: ServicioEditable): Promise<Servicio | null>;
	borrarServicio(id: string): Promise<Resultado>;
	profesionalesTurnos(): Promise<Profesional[]>;
	guardarProfesional(id: string | null, p: ProfesionalEditable): Promise<Profesional | null>;
	borrarProfesional(id: string): Promise<Resultado>;
	/** La agenda entre dos días de Argentina (inclusive), por inicio. */
	agendaTurnos(desde: string, hasta: string, profesional?: string): Promise<Turno[]>;
	/** Un turno a mano (sin seña); 'ocupado' si se superpone y no se pidió forzar. */
	crearTurno(n: TurnoNuevo): Promise<Turno | 'ocupado' | 'invalido'>;
	/** Cualquiera de los cinco estados (también para deshacer). */
	estadoTurno(id: string, estado: EstadoTurno): Promise<Turno | 'no-existe'>;
	/** Devuelve la seña pagada con Mercado Pago. */
	devolverSena(id: string): Promise<Turno | 'no-existe' | 'sin-sena' | 'sin-cobros'>;
	/** Suscripciones (§51): los planes, los suscriptores y si hay Mercado Pago para cobrarlos. */
	suscripcionesPanel(): Promise<PanelSuscripciones>;
	/** Crea (id null) o guarda un plan; null si el id no existe o 'invalido'. */
	guardarPlanSuscripcion(
		id: string | null,
		p: PlanSuscripcionEditable
	): Promise<PlanSuscripcion | null | 'invalido'>;
	/** 'no-se-puede': tiene suscriptores (se pausa en vez de borrarlo). */
	borrarPlanSuscripcion(id: string): Promise<Resultado>;
	/** La tienda da de baja una suscripción. */
	cancelarSuscriptor(id: string): Promise<Suscriptor | 'no-existe' | 'mercadopago'>;
	/** La importación en curso o, si no hay, la última (null: nunca hubo). */
	importacionActual(): Promise<Importacion | null>;
	/** Empieza a importar desde otra plataforma (una sola activa). */
	crearImportacion(
		n: ImportacionNueva
	): Promise<Importacion | 'invalido' | 'sin-fuente' | 'activa' | 'sin-cifrado'>;
	/** Si se puede conectar una tienda de Tiendanube (la app de socio está configurada). */
	tiendanubeDisponible(): Promise<boolean>;
	/** Adónde mandar a la dueña para autorizar la app (state de un solo uso); null: no disponible. */
	conectarTiendanube(): Promise<string | null>;
	/** La vuelta del OAuth: canjea el código y crea la importación con la credencial. */
	volverTiendanube(codigo: string, estado: string): Promise<Importacion | ErrorVueltaTN>;
	cancelarImportacion(id: string): Promise<Importacion | 'no-existe' | 'terminada'>;
	/**
	 * En la demo hace un paso (no hay outbox: avanza por pedidos del navegador); con la API sólo
	 * devuelve cómo va, porque avanza el worker.
	 */
	avanzarImportacion(): Promise<Importacion | null>;
	/** Los pedidos importados de otra plataforma de un cliente (el id de la lista). */
	historialCliente(id: string): Promise<PedidoHistorial[]>;
	/** Un link de pago; devuelve el código. */
	crearLink(l: {
		monto: number;
		concepto: string;
		cuotas: boolean;
		pedido: number | null;
	}): Promise<string>;
}
