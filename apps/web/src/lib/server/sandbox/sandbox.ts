/**
 * Sandbox de la demo pública. Cada visitante tiene uno (cookie), que vence a las 24 h.
 *
 * "Copia al escribir": la tienda de ejemplo de cada rubro sale del código, igual para todos; D1
 * guarda sólo el sandbox (una fila) y lo que el visitante cambie (tabla `cambios`). Así crear una
 * demo cuesta una escritura y no cientos, y el cupo gratis de D1 (100.000 escrituras por día)
 * alcanza para muchas visitas.
 */
import { PERMISOS_DE, type Rol, type RolInvitable } from '#lib/panel/permisos.ts';
import { contextoSinMontos, sinMontos, type TableroConMontos } from '../datos/montos.ts';
import {
	configuracionDemo,
	type EstadoCuenta,
	type Invitado,
	type Negocio,
	type Plan
} from './cuenta.ts';
import { finanzasDemo, type Retiro } from './finanzas.ts';
import {
	actividadDemo,
	filasDemo,
	paginaDemo,
	type FilaCambio,
	type Nombres
} from './actividad.ts';
import { canalesDemo, informesDemo, marketingDemo, type CampanaDemo } from './crecer.ts';
import {
	novedadesDemo,
	panelNewsletter,
	suscriptoresDemo,
	unirSuscriptores,
	type NovedadDemo,
	type SuscriptorDemo
} from './newsletter.ts';
import { validarNovedad } from '#lib/panel/newsletter.ts';
import { cuando } from '#lib/panel/cupones.ts';
import { descripcionDe, lunes9, vigenciaDe } from '#lib/panel/cupones.ts';
import {
	cuponesDemo,
	ENVIO_GRATIS_DESDE,
	type CambioCupon,
	type EstadoRelampago
} from './cupones.ts';
import { EsquemaEntregas, problemaEntregas } from '#lib/panel/entregas.ts';
import { error } from '@sveltejs/kit';
import * as v from 'valibot';
import { aumentar, proximoLunes, type Redondeo } from '#lib/panel/precios.ts';
import { and, desc, eq, lt } from 'drizzle-orm';
import { drizzle, type DrizzleD1Database } from 'drizzle-orm/d1';
import { inArray } from './sql.ts';
import { fromB64url, randomToken, sha256 } from '../crypto.ts';
import { rubro as buscarRubro, slugDe } from './rubros.ts';
import { cambios, sandboxes } from './schema.ts';
import {
	conectadasDeMedicion,
	DE_MEDICION,
	estadoInicial,
	INTEGRACIONES,
	type AjusteIntegracion,
	type EstadoIntegraciones
} from '#lib/panel/integraciones.ts';
import {
	MAX_CAMBIOS_POR_DIA,
	MENSAJES,
	motivoDeFormato,
	type MotivoDireccion
} from '#lib/tienda/direccion.ts';
import type {
	CambioDireccion,
	DisponibilidadDireccion,
	AjustesTurnos,
	PanelSuscripciones,
	PlanSuscripcion,
	PlanSuscripcionEditable,
	Suscriptor,
	EstadoTurno,
	Profesional,
	ProfesionalEditable,
	Servicio,
	ServicioEditable,
	Turno,
	TurnoNuevo,
	ArcaAjustes,
	ArcaAlta,
	ArcaCertificado,
	ArcaConfig,
	ArcaPedido,
	ArcaPrueba,
	Devolucion,
	ConexionML,
	Medicion,
	Entregas,
	EstadoEntregas,
	ConexionMP,
	PublicacionML,
	Dominios,
	MisDatos,
	PreferenciasAvisos,
	EstadoPushWeb,
	Acceso,
	PlanCuenta,
	Configuracion,
	Miembro,
	DisenoTienda,
	MiTienda,
	Finanzas,
	Evento,
	FiltroActividad,
	ParaHacer,
	Campana,
	Canales,
	Informes,
	Marketing,
	Newsletter,
	EnvioNewsletter,
	NovedadNueva,
	ErrorNovedad,
	Cupon,
	Cupones,
	CuponNuevo as CuponNuevoT,
	ListasPrecio,
	AccionLote,
	PedidoNuevo,
	AutomaticosCarritos,
	Tarjeta,
	Resultado,
	MovimientoStock,
	FiltrosLista,
	ListaGenerica,
	NombreLista,
	Producto,
	ProductoEditable,
	ContextoHornero,
	Datos,
	ListaPedidos,
	PedidoDetalle,
	Idea,
	Onboarding,
	Periodo,
	PaginaNotificaciones,
	ResumenPanel,
	Tablero,
	Tienda
} from '../datos/tipos.ts';
import { demoDe, type Demo, type EstadoPedido, type Pedido } from './demo.ts';
import {
	COBRADO,
	devolucionesDemo,
	resenasDemo,
	todasLasResenas,
	type CambiosResenas,
	type EstadoDevolucion,
	type PedidoNuevoDemo,
	type PreguntaTiendaDemo,
	type ResenaNuevaDemo
} from './resenas.ts';
import { codigoPedido } from '../tienda/pedido.ts';
import {
	DEF_CLIENTES,
	DEF_ENVIOS,
	DEF_FACTURACION,
	DEF_INVENTARIO,
	DEF_PRODUCTOS,
	listarGenerico,
	registrosClientes,
	unirClientes,
	claveHistorial,
	registrosEnvios,
	registrosFacturacion,
	registrosInventario,
	registrosProductos,
	claveStock,
	slugProducto,
	type CambioCliente,
	type CambioFactura,
	type CambioProducto,
	type FacturaManual,
	type Movimiento,
	type RegProducto
} from './listas.ts';
import { inicioDelDia } from './tablero.ts';
import {
	aHistorial,
	avanzarImportacion,
	cancelarImportacion,
	conectarTiendanube,
	crearImportacion,
	importacionActual,
	importadosDe,
	volverTiendanube,
	type Cifra
} from './importacion.ts';
import type { AppTN } from '../importar/tiendanube.ts';
import type { Fuentes } from '../importacion/fuente.ts';
import type { ImportacionNueva, PedidoHistorial } from '../importacion/fuente.ts';
import { cuitValido, limpiarCuit } from '#lib/panel/cuit.ts';
import {
	bandejaDemo,
	carritosDemo,
	conCarritosTienda,
	conversacionDemo,
	idCarritoTienda,
	type CambioConversacion,
	type CarritoTienda,
	type EstadoCarritos
} from './ventas.ts';
import {
	detalleDemo,
	listarDemo,
	pedidosVivos,
	transicionValida,
	type CambioPedido,
	type FiltrosPedidos
} from './pedidos.ts';
import { DIAS_DE_PRUEBA, resumenDemo } from './resumen.ts';
import { COSTO_MENSAJE, contextoDemo, tableroDemo, type CambiosTablero } from './tablero.ts';
import RUBROS from '@mostrador/contrato/rubros.json';
import { plantilla as buscarPlantilla } from '#lib/plantillas.ts';
import { SITE } from '#lib/site.ts';
import { TARJETA_INICIAL } from '#lib/panel/tarjeta.ts';
import type {
	PlanPublico,
	ProximosTurnos,
	SuscripcionPublica,
	TiendaPublica,
	TurnoPublico,
	TurnosTienda
} from '#lib/tienda/tipos.ts';
import {
	asignar,
	diaAr,
	hace,
	horaAr,
	horariosDelDia,
	isoAr,
	libres,
	ocupa,
	sumarDias,
	superpone
} from '#lib/turnos/libres.ts';
import { ENTIDAD as ENTIDAD_TURNOS, turnosDemo } from './turnos.ts';
import {
	ENTIDAD as ENTIDAD_SUSCRIPCIONES,
	mezclar as mezclarSuscripciones,
	nuevoCodigo as codigoSuscripcion,
	planPublico,
	publica as suscripcionPublicaDe
} from './suscripciones.ts';
import { diasDe } from '#lib/tienda/suscripciones.ts';
import type { CuponTienda, Historial } from '#lib/tienda/cupon.ts';
import { validarGa4, validarPixel } from '#lib/tienda/medicion.ts';
import { generarVariantes, precioDe, textoVariante, type Variante } from '#lib/tienda/variantes.ts';
import {
	aplicarLanzamiento,
	errorLanzamiento,
	proximamente,
	senaDe,
	SIN_LANZAMIENTO,
	type ConfigLanzamiento,
	type LanzamientoProducto
} from '#lib/tienda/lanzamiento.ts';

export type DbSandbox = DrizzleD1Database<Record<string, never>>;

/** ARCA en la demo: el punto de venta y los cambios de "facturar al cobrar" ([cuándo, prendido]). */
type ArcaDemo = { puntoVenta: number; historial: [number, boolean][] };
/** El CUIT de la dueña de ejemplo (Valeria Sosa, Configuración › Negocio). */
const CUIT_DEMO = '27314589020';
const DEMO_ARCA =
	'Es una tienda de ejemplo: la demo no guarda certificados ni habla con ARCA. Con tu cuenta, este paso deja la facturación andando.';
export const SIN_FACTURAR_AL_COBRAR = 'Facturar al cobrar está apagado: se factura a mano.';

/** Si facturar al cobrar estaba prendido en ese momento (arranca prendido). */
export function alCobrarEn(historial: [number, boolean][], en: number) {
	let v = true;
	for (const [t, on] of historial) if (t <= en) v = on;
	return v;
}

/** Los mismos chequeos que Go (facturacion.GenerarPedido); devuelve el error para mostrar. */
export function validarAlta(a: ArcaAlta): string | null {
	if (!cuitValido(limpiarCuit(a.cuit))) return 'Revisá el CUIT: el dígito verificador no da.';
	const r = a.razonSocial.trim();
	if (!r || r.length > 64) return 'Poné la razón social como figura en ARCA (hasta 64 letras).';
	if (a.condicion !== 'monotributo' && a.condicion !== 'responsable_inscripto')
		return 'Elegí tu condición frente al IVA.';
	if (!Number.isInteger(a.puntoVenta) || a.puntoVenta < 1 || a.puntoVenta > 99998)
		return 'El punto de venta va de 1 a 99998.';
	if (a.ambiente !== 'homologacion' && a.ambiente !== 'produccion') return 'Elegí el ambiente.';
	return null;
}

export const COOKIE = 'mostrador_demo';
export const VIDA_MS = 24 * 60 * 60 * 1000;

export const dbSandbox = (d1: D1Database): DbSandbox => drizzle(d1);

export type Sandbox = {
	id: string;
	rubro: string;
	plantilla: string | null;
	creadoEn: number;
	venceEn: number;
};

/**
 * Crea un sandbox del rubro, con su plantilla por defecto o la elegida en la galería (que tiene que
 * ser de ese rubro). Devuelve null si el rubro o la plantilla no existen.
 */
export async function crear(
	db: DbSandbox,
	rubro: string,
	ahora = Date.now(),
	plantilla: string | null = null
): Promise<Sandbox | null> {
	const r = buscarRubro(rubro);
	if (!r) return null;
	if (plantilla && buscarPlantilla(plantilla)?.rubro !== r.nombre) return null;
	const sb = { id: randomToken(24), rubro, plantilla, creadoEn: ahora, venceEn: ahora + VIDA_MS };
	await db.insert(sandboxes).values(sb);
	return sb;
}

/** El sandbox de la cookie, si existe y no venció. */
export async function leer(db: DbSandbox, id: string | undefined, ahora = Date.now()) {
	if (!id || id.length > 64) return null;
	const [sb] = await db.select().from(sandboxes).where(eq(sandboxes.id, id)).limit(1);
	return sb && sb.venceEn > ahora ? sb : null;
}

/**
 * Borra los sandboxes vencidos y sus cambios. Va de a tandas chicas para no pasar el tope de CPU ni
 * de parámetros de D1; lo llama el cron.
 */
export async function purgar(
	db: DbSandbox,
	ahora = Date.now(),
	tanda = 50,
	// Lo que vive fuera de la D1 (las fotos en R2), antes de borrar los sandboxes.
	antes?: (ids: string[]) => Promise<unknown>
) {
	const viejos = await db
		.select({ id: sandboxes.id })
		.from(sandboxes)
		.where(lt(sandboxes.venceEn, ahora))
		.limit(tanda);
	if (!viejos.length) return 0;
	const ids = viejos.map((v) => v.id);
	await antes?.(ids);
	await db.batch([
		db.delete(cambios).where(inArray(cambios.sandboxId, ids)),
		db.delete(sandboxes).where(inArray(sandboxes.id, ids))
	]);
	return ids.length;
}

/** Lo que el visitante cambió de una entidad, por clave. `null` = borrado. */
export async function cambiosDe(db: DbSandbox, sandboxId: string, entidad: string) {
	const filas = await db
		.select()
		.from(cambios)
		.where(and(eq(cambios.sandboxId, sandboxId), eq(cambios.entidad, entidad)));
	return new Map(filas.map((f) => [f.clave, f.borrado ? null : JSON.parse(f.datos ?? 'null')]));
}

/** uuid estable a partir de un texto (para ids de ejemplo que no cambian entre pedidos). */
async function uuidDe(texto: string) {
	const h = [...fromB64url(await sha256(texto))]
		.map((b) => b.toString(16).padStart(2, '0'))
		.join('');
	return `${h.slice(0, 8)}-${h.slice(8, 12)}-4${h.slice(13, 16)}-8${h.slice(17, 20)}-${h.slice(20, 32)}`;
}

/** Guarda (o reemplaza) lo que el visitante cambió de una entidad. */
export async function guardarCambio(
	db: DbSandbox,
	sandboxId: string,
	entidad: string,
	clave: string,
	datos: unknown,
	ahora = Date.now()
) {
	const fila = {
		sandboxId,
		entidad,
		clave,
		datos: JSON.stringify(datos),
		borrado: false,
		actualizadoEn: ahora
	};
	await db
		.insert(cambios)
		.values(fila)
		.onConflictDoUpdate({
			target: [cambios.sandboxId, cambios.entidad, cambios.clave],
			set: { datos: fila.datos, borrado: false, actualizadoEn: ahora }
		});
}

/** Borra varios cambios de una entidad en una sola consulta. */
export async function borrarCambios(
	db: DbSandbox,
	sandboxId: string,
	entidad: string,
	claves: string[]
) {
	if (!claves.length) return;
	await db
		.delete(cambios)
		.where(
			and(
				eq(cambios.sandboxId, sandboxId),
				eq(cambios.entidad, entidad),
				inArray(cambios.clave, claves)
			)
		);
}

/** Borra lo que el visitante cambió (vuelve al dato de ejemplo). */
export async function borrarCambio(
	db: DbSandbox,
	sandboxId: string,
	entidad: string,
	clave: string
) {
	await db
		.delete(cambios)
		.where(
			and(eq(cambios.sandboxId, sandboxId), eq(cambios.entidad, entidad), eq(cambios.clave, clave))
		);
}

const IDEAS: Idea[] = ['carritos', 'stock', 'promo'];

/** "jueves 23/10 a las 17:30" (hora de Argentina), para avisos de turnos. */
export function cuandoTurno(iso: string) {
	const t = Date.parse(iso);
	const d = new Date(t - 3 * 3_600_000);
	const dia = ['domingo', 'lunes', 'martes', 'miércoles', 'jueves', 'viernes', 'sábado'][
		d.getUTCDay()
	];
	return `${dia} ${d.getUTCDate()}/${d.getUTCMonth() + 1} a las ${horaAr(t)}`;
}

/** Suma un mensaje a Tincho en el día del visitante y devuelve cuántos lleva. */
export async function sumarUsoHornero(db: DbSandbox, sandboxId: string, dia: string) {
	const uso = (await cambiosDe(db, sandboxId, 'hornero')).get('uso') as
		{ dia: string; n: number } | undefined;
	const n = uso?.dia === dia ? uso.n + 1 : 1;
	await guardarCambio(db, sandboxId, 'hornero', 'uso', { dia, n });
	return n;
}

/** Un link de pago de la demo. */
export type LinkDemo = {
	monto: number;
	concepto: string;
	cuotas: boolean;
	pedido: number | null;
	creado: number;
	pagado: number | null;
};

const ALFABETO = '23456789ABCDEFGHJKLMNPQRSTUVWXYZ';
/** Código corto y sin letras que se confunden (0/O, 1/I). */
function codigo() {
	const b = crypto.getRandomValues(new Uint8Array(6));
	return [...b].map((x) => ALFABETO[x % ALFABETO.length]).join('');
}

/** El link por su código (de cualquier visitante: lo abre quien paga, sin cookie). */
export async function buscarLink(db: DbSandbox, cod: string, ahora = Date.now()) {
	if (!/^[2-9A-Z]{6}$/.test(cod)) return null;
	const [f] = await db
		.select()
		.from(cambios)
		.where(and(eq(cambios.entidad, 'link'), eq(cambios.clave, cod)))
		.limit(1);
	if (!f || f.borrado) return null;
	const sb = await leer(db, f.sandboxId, ahora);
	if (!sb) return null;
	return { sb, link: JSON.parse(f.datos ?? 'null') as LinkDemo };
}

/**
 * Un cambio buscado por su clave en todos los sandboxes (el link de pago o el chat de Telegram no
 * traen la cookie de la demo). Sólo devuelve sandboxes que no vencieron.
 */
export async function buscarCambio<T>(
	db: DbSandbox,
	entidad: string,
	clave: string,
	ahora = Date.now()
) {
	const [f] = await db
		.select()
		.from(cambios)
		.where(and(eq(cambios.entidad, entidad), eq(cambios.clave, clave)))
		.limit(1);
	if (!f || f.borrado) return null;
	const sb = await leer(db, f.sandboxId, ahora);
	if (!sb) return null;
	return { sb, datos: JSON.parse(f.datos ?? 'null') as T };
}

/** El sábado que viene a las 10 de Argentina (el mejor horario para campañas). */
function sabado10(ahora: number) {
	const ar = new Date(ahora - 3 * 3_600_000);
	const dias = (6 - ar.getUTCDay() + 7) % 7 || 7;
	return Date.UTC(ar.getUTCFullYear(), ar.getUTCMonth(), ar.getUTCDate() + dias, 13);
}

/** Las secciones de la portada y los colores, como vienen en la plantilla. */
const DISENO_INICIAL: DisenoTienda = {
	secciones: (['hero', 'cats', 'prods', 'revs', 'ig', 'wa'] as const).map((id) => ({
		id,
		visible: true
	})),
	paleta: 0
};

const NOMBRE_PLAN: Record<Plan, string> = {
	gratis: 'Gratis',
	emprende: 'Emprende',
	pro: 'Pro',
	negocio: 'Negocio'
};

type DominioDemo = { nombre: string; en: number };
export type TipoExport =
	'productos' | 'clientes' | 'pedidos' | 'facturas' | 'resenas' | 'conversaciones';

type AumentoProgramado = { pct: number; redondeo: Redondeo; para: number };

/** La medición guardada en los cambios de Integraciones (la demo arranca sin IDs). */
function medicionDe(c: Map<string, unknown>): Medicion {
	const m = c.get('medicion') as Medicion | undefined;
	const demo = (c.get('conectadas') as Record<string, boolean> | undefined) ?? {};
	return {
		ga4: m?.ga4 ?? null,
		pixel: m?.pixel ?? null,
		feeds: { gm: m?.feeds.gm ?? demo.gm ?? true, meta: m?.feeds.meta ?? false }
	};
}

export class DatosSandbox implements Datos {
	constructor(
		private db: DbSandbox,
		private sb: Sandbox,
		/** Lo que viene del Worker: con qué cifrar credenciales y la app de Tiendanube (si está). */
		private ext: { cifra?: Cifra; tiendanube?: AppTN | null; fuentes?: Fuentes } = {}
	) {}

	/**
	 * Quien prueba la demo es la dueña de la tienda de ejemplo. Con otro rol (pruebas), el sandbox
	 * se comporta como la API: sin finanzas, el Inicio y Tincho no reciben montos.
	 */
	rol: Rol = 'duenio';

	async rubros() {
		return RUBROS;
	}

	async tienda(): Promise<Tienda> {
		const r = buscarRubro(this.sb.rubro)!;
		// La tienda de ejemplo es la de la plantilla elegida (o la del rubro).
		const p = buscarPlantilla(this.sb.plantilla ?? r.plantilla);
		const nombre = p?.tienda ?? r.tienda;
		const base: Tienda = {
			id: await uuidDe(`${this.sb.id}:tienda`),
			slug: slugDe(nombre),
			nombre,
			rubro: r.id,
			plantilla: p?.id ?? r.plantilla,
			plan: 'gratis',
			creadaEn: new Date(this.sb.creadoEn).toISOString()
		};
		const cambio = (await cambiosDe(this.db, this.sb.id, 'tienda')).get('tienda');
		return cambio ? { ...base, ...cambio } : base;
	}

	/** Las direcciones que usó la tienda del visitante y cuándo las cambió (decisiones §38). */
	async #direccion(): Promise<{ anteriores: string[]; cambios: number[] }> {
		const d = (await cambiosDe(this.db, this.sb.id, 'tienda')).get('direccion') as {
			anteriores?: string[];
			cambios?: number[];
		} | null;
		return { anteriores: d?.anteriores ?? [], cambios: d?.cambios ?? [] };
	}

	/** En la demo no hay otras tiendas: sólo las vitrinas, que están reservadas. */
	async disponibilidadDireccion(slug: string): Promise<DisponibilidadDireccion> {
		const [t, dir] = await Promise.all([this.tienda(), this.#direccion()]);
		// Una que ya usó gana a los reservados (la de la plantilla con la que arrancó es una vitrina).
		const formato = motivoDeFormato(slug);
		const motivo: MotivoDireccion =
			slug === t.slug
				? 'actual'
				: formato === 'invalido'
					? formato
					: dir.anteriores.includes(slug)
						? 'propio'
						: (formato ?? 'libre');
		return {
			slug,
			motivo,
			disponible: motivo === 'libre' || motivo === 'propio',
			mensaje: MENSAJES[motivo]
		};
	}

	async cambiarDireccion(slug: string, ahora = Date.now()): Promise<CambioDireccion> {
		const d = await this.disponibilidadDireccion(slug);
		if (d.motivo === 'actual') return { ok: true, slug, anterior: slug };
		if (d.motivo !== 'libre' && d.motivo !== 'propio')
			return { ok: false, motivo: d.motivo, mensaje: d.mensaje };
		const [t, dir, previo] = await Promise.all([
			this.tienda(),
			this.#direccion(),
			cambiosDe(this.db, this.sb.id, 'tienda').then((m) => m.get('tienda') ?? {})
		]);
		const recientes = dir.cambios.filter((x) => x > ahora - 86_400_000);
		if (recientes.length >= MAX_CAMBIOS_POR_DIA)
			return {
				ok: false,
				motivo: 'tope',
				mensaje: `Ya cambiaste la dirección ${MAX_CAMBIOS_POR_DIA} veces hoy. Probá de nuevo mañana.`
			};
		await guardarCambio(this.db, this.sb.id, 'tienda', 'tienda', { ...previo, slug }, ahora);
		await guardarCambio(
			this.db,
			this.sb.id,
			'tienda',
			'direccion',
			{
				anteriores: [...dir.anteriores.filter((x) => x !== slug), t.slug],
				cambios: [...recientes, ahora]
			},
			ahora
		);
		return { ok: true, slug, anterior: t.slug };
	}

	/** La dirección vigente si slug es una que la tienda del visitante usó antes, o null. */
	async slugVigente(slug: string): Promise<string | null> {
		const dir = await this.#direccion();
		return dir.anteriores.includes(slug) ? (await this.tienda()).slug : null;
	}

	async onboarding(): Promise<Onboarding> {
		const g = (await cambiosDe(this.db, this.sb.id, 'onboarding')).get('estado') as
			{ elecciones?: Record<string, unknown>; terminado?: boolean } | undefined;
		return {
			plantilla: (await this.tienda()).plantilla,
			elecciones: g?.elecciones ?? {},
			terminado: g?.terminado ?? false
		};
	}

	/** La plantilla ya la fija el sandbox al crearse; acá queda el resto de lo elegido. */
	async guardarOnboarding(o: Onboarding) {
		await guardarCambio(this.db, this.sb.id, 'onboarding', 'estado', {
			elecciones: o.elecciones,
			terminado: o.terminado ?? false
		});
	}

	/** Los datos de ejemplo con lo que cambió el visitante (estado y factura de cada pedido). */
	/**
	 * Los datos de ejemplo más los pedidos que cargó el visitante (punto de venta, crear pedido), sin
	 * los cambios de estado. Los pedidos nuevos traen su comprador como cliente.
	 */
	private async base(): Promise<Demo> {
		const d = demoDe(this.plantilla(), this.sb.creadoEn)!;
		const nuevos = [...(await cambiosDe(this.db, this.sb.id, 'pedido-nuevo'))]
			.map(([, v]) => v as PedidoNuevoDemo | null)
			.filter((x): x is PedidoNuevoDemo => !!x)
			.sort((a, b) => b.pedido.numero - a.pedido.numero);
		if (!nuevos.length) return d;
		return {
			...d,
			pedidos: [...nuevos.map((x) => x.pedido), ...d.pedidos],
			clientes: [...d.clientes, ...nuevos.filter((x) => x.cliente).map((x) => x.cliente!)]
		};
	}

	/** Los datos de ejemplo con lo que cambió el visitante (pedidos nuevos, estados y facturas). */
	private async demo(): Promise<Demo> {
		const d = await this.base();
		const c = await this.cambiosPedidos();
		if (!c.size) return d;
		return {
			...d,
			// eslint-disable-next-line @typescript-eslint/no-unused-vars -- se descarta `cambio`
			pedidos: pedidosVivos(d, c, Number.MAX_SAFE_INTEGER).map(({ cambio, ...p }) => p)
		};
	}

	private async cambiosPedidos() {
		const c = await cambiosDe(this.db, this.sb.id, 'pedido');
		return new Map([...c].filter(([, v]) => v).map(([k, v]) => [Number(k), v as CambioPedido]));
	}

	/** Como en Go: sólo lo comprado en la tienda online tiene la confirmación (y ahí, la reseña). */
	private async conCodigo<F extends { numero: number; canal: string; codigo: string | null }>(
		f: F
	) {
		return f.canal === 'web' ? { ...f, codigo: await codigoPedido(this.sb.id, f.numero) } : f;
	}

	async listarPedidos(f: FiltrosPedidos): Promise<ListaPedidos> {
		const d = await this.base();
		const l = listarDemo(d, await this.cambiosPedidos(), f, Date.now());
		return { ...l, filas: await Promise.all(l.filas.map((x) => this.conCodigo(x))) };
	}

	async pedido(numero: number): Promise<PedidoDetalle | null> {
		const d = await this.base();
		const p = detalleDemo(d, await this.cambiosPedidos(), numero, Date.now());
		return p && this.conCodigo(p);
	}

	/** Cambia el estado de un pedido (avanzar, deshacer o cancelar). */
	async cambiarEstado(
		numero: number,
		estado: EstadoPedido
	): Promise<'ok' | 'no-existe' | 'no-se-puede'> {
		const d = await this.base();
		const cambios = await this.cambiosPedidos();
		const p = pedidosVivos(d, cambios, Date.now()).find((x) => x.numero === numero);
		if (!p) return 'no-existe';
		if (!transicionValida(p.estado, estado, p.envio)) return 'no-se-puede';
		await guardarCambio(this.db, this.sb.id, 'pedido', String(numero), {
			...p.cambio,
			estado,
			en: { ...p.cambio.en, [estado]: Date.now() }
		} satisfies CambioPedido);
		return 'ok';
	}

	async pedidosEnLote(numeros: number[], accion: AccionLote) {
		const d = await this.base();
		const cambios = await this.cambiosPedidos();
		const pedidos = new Map(pedidosVivos(d, cambios, Date.now()).map((p) => [p.numero, p]));
		const hechos: { numero: number; antes: EstadoPedido }[] = [];
		const ahora = Date.now();
		for (const n of new Set(numeros)) {
			const p = pedidos.get(n);
			if (!p) continue;
			let cambio: CambioPedido | null = null;
			if (accion === 'preparar' && p.estado === 'Para preparar')
				cambio = {
					...p.cambio,
					estado: 'Listo para despachar',
					en: { ...p.cambio.en, 'Listo para despachar': ahora }
				};
			else if (accion === 'etiquetas' && p.estado === 'Listo para despachar')
				cambio = { ...p.cambio, estado: 'En camino', en: { ...p.cambio.en, 'En camino': ahora } };
			else if (
				accion === 'facturar' &&
				!p.facturado &&
				p.estado !== 'Sin pagar' &&
				p.estado !== 'Cancelado'
			)
				cambio = { ...p.cambio, facturado: true };
			else if (accion === 'avisar' && p.estado !== 'Cancelado')
				cambio = { ...p.cambio, recordado: ahora };
			if (!cambio) continue;
			await guardarCambio(this.db, this.sb.id, 'pedido', String(n), cambio);
			hechos.push({ numero: n, antes: p.estado });
		}
		return { hechos: hechos.length, cambios: hechos };
	}

	async recordarPago(numero: number): Promise<'ok' | 'no-existe' | 'no-se-puede'> {
		const d = await this.base();
		const cambios = await this.cambiosPedidos();
		const p = pedidosVivos(d, cambios, Date.now()).find((x) => x.numero === numero);
		if (!p) return 'no-existe';
		if (p.estado !== 'Sin pagar') return 'no-se-puede';
		await guardarCambio(this.db, this.sb.id, 'pedido', String(numero), {
			...p.cambio,
			recordado: Date.now()
		});
		return 'ok';
	}

	// ---------------------------------------------------------------- Listas

	private async cambiosProductos() {
		const c = await cambiosDe(this.db, this.sb.id, 'producto');
		return new Map([...c].filter(([, v]) => v).map(([k, v]) => [k, v as CambioProducto]));
	}
	private async ajustesStock() {
		const c = await cambiosDe(this.db, this.sb.id, 'stock');
		return new Map(
			[...c].map(([k, v]) => [k, Number((v as { delta: number } | null)?.delta ?? 0)])
		);
	}
	private async regProductos() {
		const d = await this.demo();
		const borrados = await cambiosDe(this.db, this.sb.id, 'producto-borrado');
		return registrosProductos(
			d,
			Date.now(),
			await this.cambiosProductos(),
			await this.ajustesStock()
		).filter((p) => !borrados.has(p.id));
	}
	async borrarProducto(id: string): Promise<Resultado> {
		const p = (await this.regProductos()).find((x) => x.id === id);
		if (!p) return 'no-existe';
		await guardarCambio(this.db, this.sb.id, 'producto-borrado', id, { borrado: true });
		return 'ok';
	}
	async duplicarProducto(id: string): Promise<Producto | null> {
		const p = await this.producto(id);
		if (!p) return null;
		const copia = await this.guardarProducto(null, {
			nombre: `${p.nombre} (copia)`.slice(0, 120),
			descripcion: p.descripcion,
			categoria: p.categoria,
			precio: p.precio,
			costo: p.costo,
			sku: '',
			stock: 0,
			fotos: p.fotos,
			opciones: p.opciones,
			variantes: p.variantes.map((v) => ({ ...v, sku: '', stock: 0 })),
			canales: { web: false }
		});
		return copia;
	}
	private async regClientes() {
		const [d, cambios, { fichas, historial }] = await Promise.all([
			this.demo(),
			this.cambiosClientes(),
			importadosDe(this.db, this.sb.id)
		]);
		const ahora = Date.now();
		const regs = registrosClientes(d, this.sb.creadoEn, ahora, cambios);
		if (!fichas.length && !historial.length) return regs;
		return unirClientes(regs, this.emailsClientes(d, cambios), fichas, historial, ahora);
	}
	/** El email de cada cliente de la lista (los de ejemplo y los cargados en el panel), por id. */
	private emailsClientes(d: Demo, cambios: Map<string, CambioCliente>) {
		const m = new Map<string, string | null>(d.clientes.map((c) => [c.id, c.email]));
		for (const [id, c] of cambios) if (c.nuevo) m.set(id, c.nuevo.email);
		return m;
	}

	// ---------------------------------------------------------------- Importación

	importacionActual() {
		return importacionActual(this.db, this.sb.id);
	}
	crearImportacion(n: ImportacionNueva) {
		return crearImportacion(this.db, this.sb.id, n, this.opcionesImp());
	}
	/** El OAuth real también en la demo: hace falta la app de socio y con qué cifrar el token. */
	async tiendanubeDisponible() {
		return !!(this.ext.tiendanube && this.ext.cifra);
	}
	async conectarTiendanube() {
		const app = this.ext.tiendanube;
		if (!app || !this.ext.cifra) return null;
		return conectarTiendanube(this.db, this.sb.id, app);
	}
	async volverTiendanube(codigo: string, estado: string) {
		const app = this.ext.tiendanube;
		if (!app) return 'sin-tiendanube' as const;
		return volverTiendanube(this.db, this.sb.id, app, codigo, estado, this.opcionesImp());
	}
	cancelarImportacion(id: string) {
		return cancelarImportacion(this.db, this.sb.id, id);
	}
	/** Sin outbox: un paso por pedido del navegador, con la misma función que el editor de productos. */
	avanzarImportacion() {
		return avanzarImportacion(
			this.db,
			this.sb.id,
			{
				guardarProducto: (p) => this.guardarProducto(null, p),
				skus: async () =>
					new Set((await this.regProductos()).map((p) => p.sku.trim()).filter(Boolean))
			},
			this.opcionesImp()
		);
	}
	private opcionesImp() {
		return { cifra: this.ext.cifra, fuentes: this.ext.fuentes };
	}
	/** Los pedidos importados de un cliente de la lista, del más nuevo al más viejo. */
	async historialCliente(id: string): Promise<PedidoHistorial[]> {
		const [d, cambios, { fichas, historial }] = await Promise.all([
			this.demo(),
			this.cambiosClientes(),
			importadosDe(this.db, this.sb.id)
		]);
		if (!historial.length) return [];
		const clave =
			this.emailsClientes(d, cambios).get(id)?.toLowerCase() ||
			fichas.find((f) => f.id === id)?.email ||
			id;
		return historial
			.filter((h) => claveHistorial(h) === clave)
			.sort((a, b) => b.fecha - a.fecha)
			.slice(0, 200)
			.map(aHistorial);
	}
	private async cambiosClientes() {
		const c = await cambiosDe(this.db, this.sb.id, 'cliente');
		return new Map([...c].filter(([, v]) => v).map(([k, v]) => [k, v as CambioCliente]));
	}

	async listar(lista: NombreLista, f: FiltrosLista): Promise<ListaGenerica> {
		if (lista === 'clientes') return listarGenerico(await this.regClientes(), DEF_CLIENTES, f);
		if (lista === 'productos') {
			const l = listarGenerico(await this.regProductos(), DEF_PRODUCTOS, f);
			// Como la API Go: cada fila trae el resumen de sus variantes (las inactivas también).
			const filas = (l.filas as unknown as RegProducto[]).map((r) => ({
				...r,
				variantes: r.variantes.map((v) => ({
					id: v.id,
					texto: textoVariante(r.opciones, v.valores),
					sku: v.sku,
					precio: v.precio,
					stock: v.stock,
					activa: v.activa
				}))
			}));
			return { ...l, filas };
		}
		if (lista === 'inventario')
			return listarGenerico(registrosInventario(await this.regProductos()), DEF_INVENTARIO, f);
		if (lista === 'facturacion')
			return listarGenerico(await this.regFacturacion(), DEF_FACTURACION, f);
		return listarGenerico(await this.regEnvios(), DEF_ENVIOS, f);
	}

	private async regFacturacion() {
		const d = await this.demo();
		const c = await cambiosDe(this.db, this.sb.id, 'factura');
		const cambios = new Map([...c].filter(([, v]) => v).map(([k, v]) => [k, v as CambioFactura]));
		const manuales = [...(await cambiosDe(this.db, this.sb.id, 'factura-manual'))]
			.filter(([, v]) => v)
			.map(([k, v]) => [k, v as FacturaManual] as [string, FacturaManual])
			.sort((a, b) => a[1].en - b[1].en);
		const arca = await this.arcaDemo();
		return registrosFacturacion(d, Date.now(), cambios, manuales, (en) =>
			alCobrarEn(arca.historial, en) ? null : SIN_FACTURAR_AL_COBRAR
		);
	}

	private async regEnvios() {
		const d = await this.base();
		const vivos = pedidosVivos(d, await this.cambiosPedidos(), Date.now());
		const reclamados = new Set((await cambiosDe(this.db, this.sb.id, 'reclamo')).keys());
		return registrosEnvios(vivos, d.clientes, Date.now(), reclamados);
	}

	/** Movimientos de stock de hoy: ventas (de los pedidos) y ajustes, conteos e ingresos. */
	async movimientosHoy(): Promise<MovimientoStock[]> {
		const d = await this.demo();
		const ahora = Date.now();
		const desde = inicioDelDia(ahora);
		const nombres = new Map(
			(await this.regProductos()).flatMap((p) => [
				[p.id, p.nombre] as const,
				...p.variantes.map(
					(v) =>
						[
							claveStock(p.id, v.id),
							`${p.nombre} (${textoVariante(p.opciones, v.valores)})`
						] as const
				)
			])
		);
		const ventas: MovimientoStock[] = d.pedidos
			.filter((p) => p.estado !== 'Cancelado' && p.creado >= desde && p.creado <= ahora)
			.flatMap((p) =>
				p.items.map((i) => ({
					tipo: p.canal === 'pos' ? 'local' : p.canal === 'ml' ? 'ml' : 'web',
					texto: `${p.canal === 'pos' ? 'Venta en local' : p.canal === 'ml' ? `Mercado Libre #${p.numero}` : `Venta #${p.numero}`} · ${i.nombre}${i.variante ? ` (${i.variante})` : ''}`,
					cantidad: -i.cantidad,
					en: new Date(p.creado).toISOString()
				}))
			);
		const propios = [...(await cambiosDe(this.db, this.sb.id, 'movimiento'))]
			.map(([, v]) => v as Movimiento | null)
			.filter((m): m is Movimiento => !!m && m.en >= desde)
			.map((m): MovimientoStock => ({
				tipo: m.motivo,
				texto: `${m.motivo === 'ingreso' ? 'Ingreso de proveedor' : m.motivo === 'conteo' ? 'Conteo' : 'Ajuste manual'} · ${nombres.get(m.productoId) ?? 'Producto'}`,
				cantidad: m.q,
				en: new Date(m.en).toISOString()
			}));
		return [...ventas, ...propios].sort((a, b) => b.en.localeCompare(a.en)).slice(0, 20);
	}

	/** Ajusta el stock de un producto sin opciones o de una variante (clave producto~variante). */
	private async ajustar(productoId: string, q: number, motivo: Movimiento['motivo']) {
		const actual = (await this.ajustesStock()).get(productoId) ?? 0;
		const ahora = Date.now();
		await guardarCambio(this.db, this.sb.id, 'stock', productoId, { delta: actual + q });
		await guardarCambio(this.db, this.sb.id, 'movimiento', `${ahora}-${productoId}`, {
			productoId,
			q,
			motivo,
			en: ahora
		} satisfies Movimiento);
	}

	async accionLista(lista: NombreLista, ids: string[], accion: string, valor?: unknown) {
		const ahora = Date.now();
		if (lista === 'clientes') {
			const todos = await this.regClientes();
			// La sugerencia de Tincho: cupón a los que están en riesgo.
			const elegidos =
				accion === 'cupon-riesgo'
					? todos.filter((c) => c.seg === 'risk' && !c.cupon)
					: todos.filter((c) => ids.includes(c.id));
			const cambios = await this.cambiosClientes();
			const guardar = (id: string, c: CambioCliente) =>
				guardarCambio(this.db, this.sb.id, 'cliente', id, { ...cambios.get(id), ...c });
			if (accion === 'cupon' || accion === 'cupon-riesgo') {
				for (const c of elegidos) await guardar(c.id, { cupon: ahora });
				return elegidos.length;
			}
			if (accion === 'sin-cupon') {
				for (const c of elegidos) await guardar(c.id, { cupon: undefined });
				return elegidos.length;
			}
			if (accion === 'etiquetar' && typeof valor === 'string' && valor.trim()) {
				const e = valor.trim().slice(0, 30);
				for (const c of elegidos)
					await guardar(c.id, {
						etiquetas: [...new Set([...(cambios.get(c.id)?.etiquetas ?? []), e])]
					});
				return elegidos.length;
			}
			if (accion === 'crear') {
				const v = valor as {
					nombre?: unknown;
					telefono?: unknown;
					email?: unknown;
					ciudad?: unknown;
				} | null;
				const nombre = typeof v?.nombre === 'string' ? v.nombre.trim().slice(0, 80) : '';
				if (nombre.length < 2) return 0;
				const texto = (x: unknown, max: number) =>
					typeof x === 'string' && x.trim() ? x.trim().slice(0, max) : null;
				await guardar(`n${ahora.toString(36)}`, {
					nuevo: {
						nombre,
						telefono: texto(v?.telefono, 30),
						email: texto(v?.email, 120),
						ciudad: texto(v?.ciudad, 60) ?? '',
						creado: ahora
					}
				});
				return 1;
			}
			// Escribir no cambia datos: abre la conversación.
			return elegidos.length;
		}
		if (lista === 'productos') {
			const todos = await this.regProductos();
			const cambios = await this.cambiosProductos();
			const elegidos =
				accion === 'pausar-sin-stock-ml'
					? todos.filter((p) => p.estado === 'out' && p.canales.includes('ml'))
					: todos.filter((p) => ids.includes(p.id));
			if (accion === 'precio' && typeof valor === 'number' && valor > -90 && valor <= 200) {
				for (const p of elegidos) {
					const precio = Math.max(0, Math.round((p.precio * (1 + valor / 100)) / 10) * 10);
					await guardarCambio(this.db, this.sb.id, 'producto', p.id, {
						...cambios.get(p.id),
						precio
					});
				}
				return elegidos.length;
			}
			const publicado =
				accion === 'publicar'
					? Boolean(valor)
					: accion === 'pausar' || accion === 'pausar-sin-stock-ml'
						? false
						: null;
			if (publicado === null) return elegidos.length;
			for (const p of elegidos)
				await guardarCambio(this.db, this.sb.id, 'producto', p.id, {
					...cambios.get(p.id),
					publicado
				});
			return elegidos.length;
		}
		if (lista === 'inventario') {
			const productos = await this.regProductos();
			// Lo que tiene stock propio: los productos sin opciones y cada variante (producto~variante).
			const unidades = productos.flatMap((p) =>
				p.opciones.length
					? p.variantes.map((v) => ({ id: claveStock(p.id, v.id), stock: v.stock }))
					: [{ id: p.id, stock: p.stock }]
			);
			if (
				accion === 'ajustar' &&
				typeof valor === 'number' &&
				Number.isInteger(valor) &&
				Math.abs(valor) <= 10_000
			) {
				let n = 0;
				for (const p of unidades.filter((x) => ids.includes(x.id))) {
					// No puede quedar stock negativo.
					const q = Math.max(-p.stock, valor);
					if (!q) continue;
					await this.ajustar(p.id, q, 'ajuste');
					n++;
				}
				return n;
			}
			if (accion === 'conteo' && Array.isArray(valor)) {
				let n = 0;
				const stock = new Map(unidades.map((p) => [p.id, p.stock]));
				for (const c of valor as { id?: unknown; cantidad?: unknown }[]) {
					const id = String(c?.id ?? '');
					const cantidad = Number(c?.cantidad);
					if (!stock.has(id) || !Number.isInteger(cantidad) || cantidad < 0 || cantidad > 1_000_000)
						continue;
					const q = cantidad - stock.get(id)!;
					if (!q) continue;
					await this.ajustar(id, q, 'conteo');
					n++;
				}
				return n;
			}
			if (accion === 'ingreso' && typeof valor === 'number' && valor > 0 && valor <= 100_000) {
				const validos = ids.filter((id) => unidades.some((u) => u.id === id));
				for (const id of validos) await this.ajustar(id, Math.round(valor), 'ingreso');
				return validos.length;
			}
			// Pedir al proveedor, mover e imprimir códigos no cambian el stock.
			return productos.filter((p) => ids.includes(p.id)).length;
		}
		if (lista === 'facturacion') {
			const todos = await this.regFacturacion();
			const c = await cambiosDe(this.db, this.sb.id, 'factura');
			const guardar = (id: string, v: CambioFactura) =>
				guardarCambio(this.db, this.sb.id, 'factura', id, {
					...(c.get(id) as CambioFactura | null),
					...v
				});
			const elegidos = todos.filter((f) => ids.includes(f.id));
			if (accion === 'enviar') {
				for (const f of elegidos.filter((x) => x.estado === 'ok'))
					await guardar(f.id, { enviado: true });
				return elegidos.filter((x) => x.estado === 'ok').length;
			}
			if (accion === 'reintentar') {
				const a = elegidos.filter((x) => x.estado === 'pend' || x.estado === 'rej');
				for (const f of a) await guardar(f.id, { autorizada: true });
				return a.length;
			}
			// Anular con nota de crédito: sólo facturas autorizadas (no notas) y una vez.
			if (accion === 'anular') {
				const a = elegidos.filter((x) => x.estado === 'ok' && !x.id.startsWith('NC-'));
				for (const f of a) await guardar(f.id, { anulada: ahora });
				return a.length;
			}
			if (accion === 'emitir') {
				const v = valor as Partial<FacturaManual> | null;
				const cliente = typeof v?.cliente === 'string' ? v.cliente.trim().slice(0, 80) : '';
				const total = Number(v?.total);
				const arca = await this.arca();
				// La letra sale de la condición de la tienda (como en Go): A sólo si es responsable
				// inscripta y el cliente también; si no, B; monotributo, siempre C.
				const tipo = arca.condicion !== 'responsable_inscripto' ? 'C' : v?.tipo === 'A' ? 'A' : 'B';
				const cuit = typeof v?.cuit === 'string' ? limpiarCuit(v.cuit) : '';
				if (cliente.length < 2 || !Number.isInteger(total) || total <= 0 || total > 1_000_000_000)
					return 0;
				if (tipo === 'A' && !cuitValido(cuit)) return 0;
				await guardarCambio(this.db, this.sb.id, 'factura-manual', `M-${ahora.toString(36)}`, {
					cliente,
					tipo,
					cuit: tipo === 'A' ? cuit : null,
					concepto: typeof v?.concepto === 'string' ? v.concepto.slice(0, 200) : '',
					total,
					en: ahora
				} satisfies FacturaManual);
				return 1;
			}
			return elegidos.length;
		}
		if (lista === 'envios') {
			const todos = await this.regEnvios();
			const elegidos =
				accion === 'reclamar-todos'
					? todos.filter((e) => e.estado === 'late' && !e.reclamado)
					: accion === 'etiquetas' && !ids.length
						? todos.filter((e) => e.estado === 'todo')
						: todos.filter((e) => ids.includes(e.id));
			if (accion === 'reclamar' || accion === 'reclamar-todos') {
				for (const e of elegidos)
					await guardarCambio(this.db, this.sb.id, 'reclamo', e.id, { en: ahora });
				return elegidos.length;
			}
			if (accion === 'etiquetas') {
				let n = 0;
				for (const e of elegidos.filter((x) => x.estado === 'todo'))
					if ((await this.cambiarEstado(Number(e.id), 'En camino')) === 'ok') n++;
				return n;
			}
			if (accion === 'sin-etiqueta') {
				let n = 0;
				for (const e of elegidos)
					if ((await this.cambiarEstado(Number(e.id), 'Listo para despachar')) === 'ok') n++;
				return n;
			}
			// Pedir retiro y avisar esperan a los transportes y a WhatsApp.
			return elegidos.length;
		}
		return 0;
	}

	async producto(id: string): Promise<Producto | null> {
		const p = (await this.regProductos()).find((x) => x.id === id);
		if (!p) return null;
		return {
			id: p.id,
			nombre: p.nombre,
			descripcion: p.descripcion,
			categoria: p.categoria === 'Sin categoría' ? null : p.categoria,
			precio: p.precio,
			costo: p.costo,
			sku: p.sku,
			stock: p.stock,
			foto: p.foto,
			fotos: p.fotos,
			opciones: p.opciones,
			variantes: p.variantes,
			canales: Object.fromEntries(
				['web', 'ml', 'ig', 'google'].map((k) => [k, p.canales.includes(k)])
			),
			publicado: p.publicado,
			vendidos30d: p.vendidos30,
			slug: slugProducto(p.nombre)
		};
	}

	async guardarProducto(id: string | null, p: ProductoEditable): Promise<Producto | null> {
		const cambios = await this.cambiosProductos();
		// El sufijo al azar: dos productos creados en el mismo milisegundo (una importación) no se pisan.
		const clave = id ?? `n${Date.now().toString(36)}${crypto.randomUUID().slice(0, 4)}`;
		const actual = id ? (await this.regProductos()).find((x) => x.id === id) : undefined;
		if (id && !actual) return null;
		const antes = cambios.get(clave) ?? {};
		// Galería (como la API Go): `fotos` la reemplaza (sin vacías ni repetidas); sin `fotos`, `foto`
		// cambia la portada y deja el resto (null la vacía); sin ninguna, queda la que había.
		const previa = actual?.fotos ?? [];
		const fotos =
			p.fotos !== undefined
				? [...new Set(p.fotos.map((f) => f.trim()).filter(Boolean))].slice(0, 8)
				: p.foto !== undefined
					? p.foto === null
						? []
						: [p.foto, ...previa.slice(1).filter((f) => f !== p.foto)]
					: previa;
		// Opciones: sin mandar quedan las que había; [] las saca y el stock vuelve a ser editable.
		const opciones =
			p.opciones?.map((o) => ({
				nombre: o.nombre.trim(),
				valores: o.valores.map((v) => v.trim())
			})) ??
			actual?.opciones ??
			[];
		// Una variante por combinación: las que siguen conservan stock, precio, SKU y foto; las nuevas
		// arrancan en 0 y activas; las que dejan de existir se van (sus pedidos guardaron el texto). Lo
		// que manda el editor (identificado por los valores) pisa lo que había.
		const nuevoId = (i: number) => `${clave}-n${Date.now().toString(36)}${i}`;
		const variantes = generarVariantes(opciones, actual?.variantes ?? []).map((v, i): Variante => {
			const e = p.variantes?.find((x) =>
				opciones.every((o) => x.valores[o.nombre] === v.valores[o.nombre])
			);
			const r = e
				? {
						...v,
						sku: e.sku?.trim() ?? v.sku,
						precio: e.precio ?? null,
						foto: e.foto ?? null,
						stock: e.stock ?? v.stock,
						activa: e.activa ?? v.activa
					}
				: v;
			return {
				...r,
				id: r.id || nuevoId(i),
				foto: r.foto && fotos.includes(r.foto) ? r.foto : null
			};
		});
		// El stock que se escribe en la ficha reemplaza al de ejemplo y a los ajustes (el de cada
		// variante ya viene con sus ajustes y ventas).
		if (p.stock !== undefined || opciones.length)
			await borrarCambio(this.db, this.sb.id, 'stock', clave);
		if (actual?.variantes.length || variantes.length)
			await borrarCambios(
				this.db,
				this.sb.id,
				'stock',
				[...(actual?.variantes ?? []), ...variantes].map((v) => claveStock(clave, v.id))
			);
		await guardarCambio(this.db, this.sb.id, 'producto', clave, {
			...antes,
			nombre: p.nombre,
			descripcion: p.descripcion ?? antes.descripcion,
			categoria: p.categoria ?? antes.categoria,
			precio: p.precio,
			costo: p.costo ?? antes.costo,
			sku: p.sku ?? antes.sku,
			stock: opciones.length ? antes.stock : (p.stock ?? antes.stock),
			foto: fotos[0] ?? null,
			fotos,
			opciones,
			variantes,
			canales: p.canales ?? antes.canales,
			publicado: true,
			...(id ? {} : { nuevo: true, creado: Date.now() })
		} satisfies CambioProducto);
		return this.producto(clave);
	}

	// ---------------------------------------------------------------- Bandeja y carritos

	private async cambiosConversaciones() {
		const [c, leidas] = await Promise.all([
			cambiosDe(this.db, this.sb.id, 'conversacion'),
			cambiosDe(this.db, this.sb.id, 'conversacion-leida')
		]);
		const m = new Map([...c].filter(([, v]) => v).map(([k, v]) => [k, v as CambioConversacion]));
		for (const [k, v] of leidas) if (v) m.set(k, { ...m.get(k), leida: true });
		return m;
	}

	async bandeja() {
		return bandejaDemo(await this.demo(), await this.cambiosConversaciones());
	}

	async conversacion(id: string) {
		return conversacionDemo(await this.demo(), id, await this.cambiosConversaciones());
	}

	async responder(
		id: string,
		c: { texto?: string; tomada?: boolean; leida?: boolean }
	): Promise<Resultado> {
		const cambios = await this.cambiosConversaciones();
		if (!conversacionDemo(await this.demo(), id, cambios)) return 'no-existe';
		// "Leída" va aparte: llega junto con el primer mensaje y no tiene que pisarlo.
		if (c.leida && !c.texto?.trim() && c.tomada === undefined) {
			await guardarCambio(this.db, this.sb.id, 'conversacion-leida', id, true);
			return 'ok';
		}
		const antes = cambios.get(id) ?? {};
		const texto = c.texto?.trim().slice(0, 2000);
		await guardarCambio(this.db, this.sb.id, 'conversacion', id, {
			...antes,
			...(texto
				? { mensajes: [...(antes.mensajes ?? []), { texto, en: Date.now() }], leida: true }
				: {}),
			...(c.tomada !== undefined ? { tomada: c.tomada } : {}),
			...(c.leida ? { leida: true } : {})
		} satisfies CambioConversacion);
		return 'ok';
	}

	private async estadoCarritos(): Promise<EstadoCarritos> {
		const c = await cambiosDe(this.db, this.sb.id, 'carrito');
		const auto = (await cambiosDe(this.db, this.sb.id, 'carritos')).get('automaticos') as
			EstadoCarritos['automaticos'] | undefined;
		return {
			recordados: new Set([...c].filter(([, v]) => v).map(([k]) => k)),
			todos: (await this.ideasHechas()).includes('carritos'),
			automaticos: auto ?? { hora: true, dia: true, email: false }
		};
	}

	async carritos() {
		const e = await this.estadoCarritos();
		const base = carritosDemo(await this.demo(), Date.now(), e, 120);
		const propios = await this.carritosTienda();
		if (!propios.length) return base;
		const fotos = new Map((await this.regProductos()).map((p) => [p.id, p.foto]));
		return conCarritosTienda(base, propios, (id) => fotos.get(id) ?? null, e, Date.now());
	}

	// ---------------------------------------------------------------- Carritos de la tienda online

	private async carritosTienda() {
		const c = await cambiosDe(this.db, this.sb.id, 'carrito-tienda');
		return [...c.values()].filter(Boolean) as CarritoTienda[];
	}

	/**
	 * Guarda el carrito que alguien dejó en el checkout de la tienda (uno abierto por email). Como
	 * Go: sólo productos publicados y variantes que se venden; el total, con los precios de la tienda.
	 */
	async guardarCarritoTienda(
		email: string,
		nombre: string,
		lineas: { id: string; varianteId: string | null; cantidad: number }[]
	) {
		const tienda = await this.tiendaPublica(false);
		const porId = new Map(tienda.productos.map((p) => [p.id, p]));
		let total = 0;
		const items: CarritoTienda['items'] = [];
		for (const l of lineas) {
			const p = porId.get(l.id);
			const v = p?.opciones.length ? p.variantes.find((x) => x.id === l.varianteId) : null;
			if (!p || (p.opciones.length && !v)) continue;
			total += (v ? precioDe(v, p.precio) : p.precio) * l.cantidad;
			items.push({ id: p.id, varianteId: v?.id ?? null, cantidad: l.cantidad });
		}
		if (!items.length) return false;
		const clave = email.trim().toLowerCase();
		const previo = (await cambiosDe(this.db, this.sb.id, 'carrito-tienda')).get(clave) as
			CarritoTienda | null | undefined;
		const abierto = previo && !previo.recuperado ? previo : null;
		const ahora = Date.now();
		await guardarCambio(this.db, this.sb.id, 'carrito-tienda', clave, {
			email: clave,
			nombre: nombre.trim(),
			items,
			total,
			token: abierto?.token ?? randomToken(24),
			creado: abierto?.creado ?? ahora,
			actualizado: ahora,
			recuperado: null
		} satisfies CarritoTienda);
		return true;
	}

	/** Las líneas del carrito del link de recordatorio (null si no existe o ya compró). */
	async carritoTienda(token: string) {
		const c = (await this.carritosTienda()).find((x) => x.token === token && !x.recuperado);
		return c?.items ?? null;
	}

	/** Ese email compró: su carrito queda recuperado y deja de aparecer. */
	async cerrarCarritoTienda(email: string, total: number) {
		const clave = email.trim().toLowerCase();
		const c = (await cambiosDe(this.db, this.sb.id, 'carrito-tienda')).get(clave) as
			CarritoTienda | null | undefined;
		if (!c || c.recuperado) return;
		await guardarCambio(this.db, this.sb.id, 'carrito-tienda', clave, {
			...c,
			recuperado: Date.now(),
			recuperadoTotal: total
		} satisfies CarritoTienda);
	}

	async tarjeta(): Promise<Tarjeta> {
		const t = (await cambiosDe(this.db, this.sb.id, 'tarjeta')).get('tarjeta') as
			Tarjeta | undefined;
		return { ...TARJETA_INICIAL, ...t };
	}

	async guardarTarjeta(t: Tarjeta) {
		await guardarCambio(this.db, this.sb.id, 'tarjeta', 'tarjeta', t);
	}

	async guardarAutomaticos(a: AutomaticosCarritos) {
		await guardarCambio(this.db, this.sb.id, 'carritos', 'automaticos', {
			hora: !!a.hora,
			dia: !!a.dia,
			email: !!a.email
		});
	}

	async recordarCarrito(id: string): Promise<Resultado> {
		const d = await this.demo();
		const deLaTienda = (await this.carritosTienda()).some(
			(c) => !c.recuperado && idCarritoTienda(c) === id
		);
		if (!deLaTienda && !d.carritos.some((k) => k.id === id)) return 'no-existe';
		const e = await this.estadoCarritos();
		if (e.todos || e.recordados.has(id)) return 'no-se-puede';
		await guardarCambio(this.db, this.sb.id, 'carrito', id, { en: Date.now() });
		return 'ok';
	}

	// ---------------------------------------------------------------- Pedidos nuevos

	async crearPedido(
		n: PedidoNuevo,
		/** Tienda online: email de quien compra y el cupón que usó (para "una vez por cliente"). */
		web?: {
			email: string;
			cupon: string | null;
			descuentoCupon: number;
			total?: number;
			entrega?: { direccion: string; cp: string } | null;
			/** El precio de cada línea (con la oferta vigente, §52), por claveStock. */
			precios?: Record<string, number>;
			/** Reserva o encargo con seña (§52): en la demo la seña se da por pagada. */
			sena?: { tipo: 'reserva' | 'encargo'; monto: number };
		}
	): Promise<{ numero: number; total: number } | 'sin-stock' | 'invalido'> {
		const d = await this.base();
		const productos = new Map((await this.regProductos()).map((p) => [p.id, p]));
		const items: Pedido['items'] = [];
		// Lo que se va a descontar, por clave de stock (dos líneas de la misma variante se suman).
		const aDescontar = new Map<string, number>();
		for (const i of n.items) {
			const p = productos.get(i.productoId);
			if (!p || !Number.isInteger(i.cantidad) || i.cantidad < 1) return 'invalido';
			// Con opciones se vende por variante: tiene que ser de este producto y estar activa.
			const v = p.opciones.length ? p.variantes.find((x) => x.id === i.varianteId) : undefined;
			if (p.opciones.length ? !v?.activa : i.varianteId) return 'invalido';
			const k = claveStock(p.id, v?.id);
			const total = (aDescontar.get(k) ?? 0) + i.cantidad;
			if ((v ? v.stock : p.stock) < total) return 'sin-stock';
			aDescontar.set(k, total);
			items.push({
				productoId: p.id,
				nombre: p.nombre,
				cantidad: i.cantidad,
				precio: web?.precios?.[k] ?? v?.precio ?? p.precio,
				varianteId: v?.id ?? null,
				variante: v ? textoVariante(p.opciones, v.valores) : null
			});
		}
		if (!items.length) return 'invalido';
		const ahora = Date.now();
		const numero = Math.max(...d.pedidos.map((p) => p.numero)) + 1;
		const nombre = n.cliente?.trim().slice(0, 80);
		// Un cliente que ya existe: el pedido se suma a su ficha.
		const existente = n.clienteId ? d.clientes.find((c) => c.id === n.clienteId) : undefined;
		const cliente: Demo['clientes'][number] | null = existente
			? null
			: {
					id: `v${numero}`,
					nombre: nombre || 'Consumidor final',
					email: web?.email.toLowerCase() ?? null,
					telefono: n.telefono?.trim() || null,
					ciudad: d.catalogo.ciudad,
					foto: null,
					mayorista: false,
					alta: ahora
				};
		const envio = n.envio ?? 'retiro';
		const pedido: Pedido = {
			numero,
			clienteId: existente?.id ?? cliente!.id,
			canal: n.canal,
			pago: n.pago,
			estadoPago:
				n.pagado === false
					? 'Pendiente'
					: n.pago === 'efectivo'
						? 'Cobrado'
						: n.pago === 'transf'
							? 'Acreditada'
							: 'Aprobado',
			envio,
			ciudadEnvio: d.catalogo.ciudad,
			entrega: envio === 'retiro' ? null : (web?.entrega ?? null),
			estado:
				n.pagado === false
					? 'Sin pagar'
					: n.canal === 'pos' && envio === 'retiro'
						? 'Entregado'
						: 'Para preparar',
			facturado: n.pagado !== false,
			items,
			// En la tienda online, el total cobrado (envío, cupón y transferencia incluidos).
			total: web?.total ?? items.reduce((s, i) => s + i.precio * i.cantidad, 0),
			creado: ahora,
			...(web?.cupon ? { cupon: web.cupon } : {}),
			...(web?.sena
				? {
						sena: { ...web.sena, pagada: true, saldado: false },
						estadoPago: 'Seña',
						estado: 'Para preparar' as const,
						facturado: false
					}
				: {})
		};
		if (web?.cupon) await this.usarCupon(web.cupon, pedido.total, web.descuentoCupon);
		await guardarCambio(this.db, this.sb.id, 'pedido-nuevo', String(numero), {
			pedido,
			cliente
		} satisfies PedidoNuevoDemo);
		// La venta descuenta stock, de la variante si tiene (el movimiento lo muestra el pedido).
		const ajustes = await this.ajustesStock();
		for (const [k, q] of aDescontar)
			await guardarCambio(this.db, this.sb.id, 'stock', k, { delta: (ajustes.get(k) ?? 0) - q });
		return { numero, total: pedido.total };
	}

	// ---------------------------------------------------------------- Cupones

	private async cuponesNuevos() {
		return [...(await cambiosDe(this.db, this.sb.id, 'cupon-nuevo'))]
			.filter(([, v]) => v)
			.map(([, v]) => v as Cupon)
			.sort((a, b) => Date.parse(b.creado) - Date.parse(a.creado));
	}

	async cupones(): Promise<Cupones> {
		const d = await this.demo();
		const [cambios, nuevos, consejo, autos, rel, clientes] = await Promise.all([
			cambiosDe(this.db, this.sb.id, 'cupon'),
			this.cuponesNuevos(),
			cambiosDe(this.db, this.sb.id, 'cupones'),
			cambiosDe(this.db, this.sb.id, 'automatica'),
			cambiosDe(this.db, this.sb.id, 'relampago'),
			this.regClientes()
		]);
		return cuponesDemo(
			d,
			Date.now(),
			this.sb.creadoEn,
			new Map([...cambios].filter(([, v]) => v).map(([k, v]) => [k, v as CambioCupon])),
			nuevos,
			!!consejo.get('consejo'),
			new Map(
				[...autos].filter(([, v]) => v).map(([k, v]) => [k, (v as { activa: boolean }).activa])
			),
			(rel.get('vivo') as EstadoRelampago | undefined) ?? {
				horas: 0,
				liberadas: 0,
				terminada: false
			},
			{
				vip: clientes.filter((c) => c.seg === 'vip').length,
				riesgo: clientes.filter((c) => c.seg === 'risk').length
			}
		);
	}

	async crearCupon(n: CuponNuevoT): Promise<Cupon | 'repetido'> {
		const todos = await this.cupones();
		if (todos.cupones.some((c) => c.codigo === n.codigo && c.estado !== 'terminado'))
			return 'repetido';
		const ahora = Date.now();
		const v = vigenciaDe(n.vigencia, ahora);
		const empieza = n.programar ? lunes9(ahora) : v.empieza;
		const nombres = todos.productos.filter((p) => n.productos?.includes(p.id)).map((p) => p.nombre);
		const c: Cupon = {
			id: `n${ahora.toString(36)}`,
			codigo: n.codigo,
			tipo: n.tipo,
			valor: n.tipo === 'pct' || n.tipo === 'fijo' ? n.valor : 0,
			descripcion: descripcionDe(n, todos.categoria, nombres),
			estado: empieza && empieza > ahora ? 'programado' : 'activo',
			usos: 0,
			limite: n.limite,
			vendio: 0,
			desconto: 0,
			canales: n.canales,
			empieza: empieza ? new Date(empieza).toISOString() : null,
			termina: v.termina
				? new Date(Math.max(v.termina, (empieza ?? 0) + 86_400_000)).toISOString()
				: null,
			minimo: n.minimo,
			consejo: n.opciones.hornero ? 'Tincho ya lo puede ofrecer en el chat.' : null,
			creado: new Date(ahora).toISOString()
		};
		// Las reglas que usa la tienda online (lib/tienda/cupon.ts) van guardadas con el cupón.
		const regla: Partial<CuponTienda> = {
			quien: n.quien,
			porCliente: n.porCliente,
			alcance: {
				tipo: n.alcance,
				productos: n.productos ?? [],
				...(n.alcance === 'categoria' && todos.categoria ? { categoria: todos.categoria } : {})
			},
			combinable: n.opciones.combinable
		};
		await guardarCambio(this.db, this.sb.id, 'cupon-nuevo', c.id, { ...c, regla });
		return c;
	}

	/** La tienda online: el cupón como lo evalúa lib/tienda/cupon.ts. */
	async cuponTienda(codigo: string): Promise<CuponTienda | null> {
		const c = (await this.cupones()).cupones.find(
			(x) => x.codigo === codigo && x.estado !== 'terminado'
		);
		if (!c) return null;
		// Los nuevos guardan sus reglas; los de ejemplo del handoff, abiertas.
		const r = (c as Cupon & { regla?: Partial<CuponTienda> }).regla ?? {};
		return {
			codigo: c.codigo,
			tipo: c.tipo,
			valor: c.valor,
			pausado: c.estado === 'pausado',
			usos: c.usos,
			limite: c.limite,
			canales: c.canales,
			empieza: c.empieza,
			termina: c.termina,
			minimo: c.minimo,
			quien: r.quien ?? 'todos',
			porCliente: r.porCliente ?? 'libre',
			alcance: r.alcance ?? { tipo: 'todo', productos: [] },
			combinable: r.combinable ?? false
		};
	}

	/** Lo que se sabe de quien compra por su email: pedidos, usos de un cupón, VIP y última compra. */
	async historialDe(email: string, codigo: string): Promise<Historial> {
		const d = await this.demo();
		const e = email.trim().toLowerCase();
		const suyos = new Set(d.clientes.filter((c) => c.email?.toLowerCase() === e).map((c) => c.id));
		const pedidos = d.pedidos.filter((p) => suyos.has(p.clienteId) && p.estado !== 'Cancelado');
		const ultima = Math.max(0, ...pedidos.map((p) => p.creado));
		return {
			pedidos: pedidos.length,
			usosDelCupon: pedidos.filter((p) => (p as { cupon?: string }).cupon === codigo).length,
			vip: pedidos.some((p) => (p as { vip?: boolean }).vip),
			ultimaCompra: ultima ? new Date(ultima).toISOString() : null
		};
	}

	private async usarCupon(codigo: string, vendio: number, desconto: number) {
		const c = (await this.cupones()).cupones.find((x) => x.codigo === codigo);
		if (!c) return;
		const cambio = { usos: c.usos + 1, vendio: c.vendio + vendio, desconto: c.desconto + desconto };
		if (c.id.startsWith('n'))
			await guardarCambio(this.db, this.sb.id, 'cupon-nuevo', c.id, { ...c, ...cambio });
		else {
			const antes = (await cambiosDe(this.db, this.sb.id, 'cupon')).get(c.id) as
				CambioCupon | undefined;
			await guardarCambio(this.db, this.sb.id, 'cupon', c.id, { ...antes, ...cambio });
		}
	}

	async accionCupon(
		id: string,
		accion: 'pausar' | 'activar' | 'duplicar' | 'minimo' | 'deshacer-minimo',
		minimo = 0
	): Promise<Cupon | 'no-existe' | 'no-se-puede'> {
		const c = (await this.cupones()).cupones.find((x) => x.id === id);
		if (!c) return 'no-existe';
		const nuevo = id.startsWith('n');
		const guardar = async (cambio: CambioCupon) => {
			if (nuevo) await guardarCambio(this.db, this.sb.id, 'cupon-nuevo', id, { ...c, ...cambio });
			else {
				const antes = (await cambiosDe(this.db, this.sb.id, 'cupon')).get(id) as
					CambioCupon | undefined;
				await guardarCambio(this.db, this.sb.id, 'cupon', id, { ...antes, ...cambio });
			}
			return { ...c, ...cambio };
		};
		if (accion === 'duplicar') {
			const copia: Cupon = {
				...c,
				id: `n${Date.now().toString(36)}`,
				codigo: `${c.codigo}-2`.slice(0, 20),
				estado: 'pausado',
				usos: 0,
				vendio: 0,
				desconto: 0,
				consejo: null,
				empieza: null,
				termina: null,
				creado: new Date().toISOString()
			};
			await guardarCambio(this.db, this.sb.id, 'cupon-nuevo', copia.id, copia);
			return copia;
		}
		if (c.estado === 'terminado') return 'no-se-puede';
		if (accion === 'pausar') return guardar({ estado: 'pausado' });
		if (accion === 'activar') return guardar({ estado: 'activo' });
		if (accion === 'minimo')
			return guardar({
				minimo,
				descripcion: `${c.descripcion.replace(/ · mín\. .*$/, '')} · mín. $ ${minimo.toLocaleString('es-AR')}`,
				consejo: null
			});
		return guardar({ minimo: 0, descripcion: c.descripcion.replace(/ · mín\. .*$/, '') });
	}

	async descartarConsejo() {
		await guardarCambio(this.db, this.sb.id, 'cupones', 'consejo', { oculto: true });
	}

	async fijarAutomatica(id: string, activa: boolean): Promise<Resultado> {
		if (!(await this.cupones()).automaticas.some((a) => a.id === id)) return 'no-existe';
		await guardarCambio(this.db, this.sb.id, 'automatica', id, { activa });
		return 'ok';
	}

	async accionRelampago(
		accion: 'sumar-hora' | 'liberar' | 'terminar' | 'reanudar'
	): Promise<Resultado> {
		const e = ((await cambiosDe(this.db, this.sb.id, 'relampago')).get('vivo') as
			EstadoRelampago | undefined) ?? { horas: 0, liberadas: 0, terminada: false };
		const r = (await this.cupones()).relampago.vivo;
		if (!r) return 'no-existe';
		if (r.terminada && accion !== 'reanudar') return 'no-se-puede';
		const nuevo: EstadoRelampago =
			accion === 'sumar-hora'
				? { ...e, horas: e.horas + 1 }
				: accion === 'liberar'
					? { ...e, liberadas: e.liberadas + 1 }
					: { ...e, terminada: accion === 'terminar' };
		await guardarCambio(this.db, this.sb.id, 'relampago', 'vivo', nuevo);
		return 'ok';
	}

	// ---------------------------------------------------------------- Crecer

	async marketing(): Promise<Marketing> {
		const d = await this.demo();
		const [clientes, cup, nuevas] = await Promise.all([
			this.regClientes(),
			this.cupones(),
			cambiosDe(this.db, this.sb.id, 'campana')
		]);
		const cat = d.catalogo.categorias[0];
		const deCat = new Set(d.catalogo.productos.filter((p) => p.categoria === cat).map((p) => p.id));
		const compraron = new Set(
			d.pedidos.filter((p) => p.items.some((i) => deCat.has(i.productoId))).map((p) => p.clienteId)
		);
		const conWa = clientes.filter((c) => c.wa);
		const activos = cup.cupones.filter((c) => c.estado === 'activo');
		return marketingDemo(
			d,
			Date.now(),
			{
				todos: conWa.length,
				vip: conWa.filter((c) => c.seg === 'vip').length,
				riesgo: conWa.filter((c) => c.seg === 'risk').length,
				categoria: conWa.filter((c) => compraron.has(c.id)).length
			},
			activos.slice(0, 3).map((c) => ({
				codigo: c.codigo,
				descripcion: c.descripcion,
				usos: `${c.usos} usos${c.termina ? ` · ${cuando(c, Date.now()).toLowerCase()}` : ''}`
			})),
			activos.find((c) => c.tipo === 'pct')?.codigo ?? null,
			[...nuevas]
				.filter(([, v]) => v)
				.map(([, v]) => v as CampanaDemo)
				.sort((a, b) => b.en - a.en),
			COSTO_MENSAJE
		);
	}

	async crearCampana(c: {
		audiencia: 'todos' | 'vip' | 'riesgo' | 'categoria';
		programar: boolean;
	}): Promise<Campana | 'sin-audiencia'> {
		const m = await this.marketing();
		const personas = m.audiencias.find((a) => a.id === c.audiencia)?.cantidad ?? 0;
		if (!personas) return 'sin-audiencia';
		const ahora = Date.now();
		const id = `n${ahora.toString(36)}`;
		await guardarCambio(this.db, this.sb.id, 'campana', id, {
			id,
			audiencia: c.audiencia,
			personas,
			en: ahora,
			programada: c.programar ? sabado10(ahora) : null
		} satisfies CampanaDemo);
		return (await this.marketing()).campanas.find((x) => x.id === id)!;
	}

	// ---------------------------------------------------------------- Newsletter (§50)

	private async suscriptores() {
		const [d, c] = await Promise.all([this.demo(), cambiosDe(this.db, this.sb.id, 'suscriptor')]);
		return unirSuscriptores(suscriptoresDemo(d, this.sb.creadoEn), c);
	}

	async newsletter(): Promise<Newsletter> {
		const [ss, c, t] = await Promise.all([
			this.suscriptores(),
			cambiosDe(this.db, this.sb.id, 'novedad'),
			this.tiendaPublica(false)
		]);
		const confirmados = ss.filter((s) => s.confirmadoEn != null).length;
		const propias = [...c.values()].filter(Boolean) as NovedadDemo[];
		return panelNewsletter(
			ss,
			[...novedadesDemo(this.sb.creadoEn, confirmados), ...propias],
			inicioDelDia(Date.now()),
			t.productos.map((p) => ({ id: p.id, nombre: p.nombre, precio: p.precio, foto: p.foto }))
		);
	}

	async borrarSuscriptor(id: string): Promise<Resultado> {
		if (!(await this.suscriptores()).some((s) => s.id === id)) return 'no-existe';
		await guardarCambio(this.db, this.sb.id, 'suscriptor', id, null);
		return 'ok';
	}

	/** En la demo la novedad queda en el historial y no sale ningún mail. */
	async mandarNovedad(n: NovedadNueva): Promise<EnvioNewsletter | ErrorNovedad> {
		const v = validarNovedad(n);
		if (!v.ok) return { error: v.error };
		const publicados = new Set((await this.tiendaPublica(false)).productos.map((p) => p.id));
		if (v.novedad.productos.some((id) => !publicados.has(id)))
			return { error: 'Algún producto ya no está publicado. Elegilos de nuevo.' };
		const destinatarios = (await this.suscriptores()).filter((s) => s.confirmadoEn != null).length;
		if (!destinatarios) return 'sin-suscriptos';
		const ahora = Date.now();
		const id = `n${ahora.toString(36)}`;
		await guardarCambio(this.db, this.sb.id, 'novedad', id, {
			id,
			asunto: v.novedad.asunto,
			creado: ahora,
			destinatarios
		} satisfies NovedadDemo);
		return {
			id,
			asunto: v.novedad.asunto,
			creado: new Date(ahora).toISOString(),
			destinatarios,
			enviados: destinatarios,
			estado: 'enviado'
		};
	}

	/**
	 * Alta desde la tienda de la demo: queda pendiente como con Go, pero no sale el mail: devuelve el
	 * token para que la tienda muestre el link de confirmar. null si el email no sirve.
	 */
	async suscribirNewsletter(email: string, origen: 'tienda' | 'checkout') {
		const e = email.trim().toLowerCase();
		if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(e) || e.length > 120) return null;
		const previo = (await this.suscriptores()).find((s) => s.email === e);
		if (previo) return { token: previo.token, confirmado: previo.confirmadoEn != null };
		const token = randomToken(24);
		const id = `s${Date.now().toString(36)}${token.slice(0, 4)}`;
		await guardarCambio(this.db, this.sb.id, 'suscriptor', id, {
			id,
			email: e,
			origen,
			creado: Date.now(),
			confirmadoEn: null,
			token
		} satisfies SuscriptorDemo);
		return { token, confirmado: false };
	}

	async confirmarNewsletter(token: string) {
		const s = (await this.suscriptores()).find((x) => x.token === token);
		if (!s) return false;
		if (!s.confirmadoEn)
			await guardarCambio(this.db, this.sb.id, 'suscriptor', s.id, {
				...s,
				confirmadoEn: Date.now()
			} satisfies SuscriptorDemo);
		return true;
	}

	async bajaNewsletter(token: string) {
		const s = (await this.suscriptores()).find((x) => x.token === token);
		if (!s) return false;
		await guardarCambio(this.db, this.sb.id, 'suscriptor', s.id, null);
		return true;
	}

	async canales(): Promise<Canales> {
		const d = await this.demo();
		const [pausados, productos] = await Promise.all([
			cambiosDe(this.db, this.sb.id, 'canal'),
			this.regProductos()
		]);
		return canalesDemo(
			d,
			Date.now(),
			new Map(
				[...pausados].filter(([, v]) => v).map(([k, v]) => [k, (v as { pausado: boolean }).pausado])
			),
			productos.filter((p) => p.canales.includes('ig')).length
		);
	}

	async fijarCanal(id: string, activo: boolean): Promise<Resultado> {
		const c = (await this.canales()).canales.find((x) => x.id === id);
		if (!c) return 'no-existe';
		if (!c.conectado) return 'no-se-puede';
		await guardarCambio(this.db, this.sb.id, 'canal', id, { pausado: !activo });
		return 'ok';
	}

	async informes(): Promise<Informes> {
		const d = await this.demo();
		const productos = await this.regProductos();
		const vendidos = productos.reduce((s, p) => s + p.vendidos30, 0);
		const margen = vendidos
			? Math.round(productos.reduce((s, p) => s + (p.margen ?? 0) * p.vendidos30, 0) / vendidos)
			: 38;
		return informesDemo(d, Date.now(), margen);
	}

	// ---------------------------------------------------------------- Actividad y Para hacer hoy

	/** Todos los cambios del visitante, del más nuevo al más viejo. */
	private async historial(): Promise<FilaCambio[]> {
		const filas = await this.db
			.select()
			.from(cambios)
			.where(and(eq(cambios.sandboxId, this.sb.id), eq(cambios.borrado, false)))
			.orderBy(desc(cambios.actualizadoEn))
			.limit(200);
		return filas.map((f) => ({
			entidad: f.entidad,
			clave: f.clave,
			datos: JSON.parse(f.datos ?? 'null'),
			en: f.actualizadoEn
		}));
	}

	private async nombres(): Promise<Nombres> {
		const [productos, cup, canales] = await Promise.all([
			this.regProductos(),
			this.cupones(),
			this.canales()
		]);
		const d = await this.base();
		return {
			productos: new Map(productos.map((p) => [p.id, p.nombre])),
			cupones: new Map(cup.cupones.map((c) => [c.id, c.codigo])),
			automaticas: new Map(cup.automaticas.map((a) => [a.id, a.titulo])),
			canales: new Map(canales.canales.map((c) => [c.id, c.nombre])),
			totales: new Map(d.pedidos.map((p) => [p.numero, p.total]))
		};
	}

	async actividad(f: FiltroActividad = {}) {
		return paginaDemo(await this.actividadToda(), f);
	}

	async exportarActividad(f: FiltroActividad = {}) {
		return filasDemo(await this.actividadToda(), f);
	}

	private async actividadToda(): Promise<Evento[]> {
		const [filas, nombres] = await Promise.all([this.historial(), this.nombres()]);
		const deshechos = filas
			.filter((f) => f.entidad === 'deshecho')
			.map((f) => f.datos as { id: string; texto: string; donde: string; en: number });
		return actividadDemo(
			await this.demo(),
			Date.now(),
			this.sb.creadoEn,
			filas,
			deshechos,
			nombres
		);
	}

	async deshacerEvento(id: string): Promise<Resultado> {
		const e = (await this.actividadToda()).find((x) => x.id === id);
		if (!e) return 'no-existe';
		if (!e.deshacer) return 'no-se-puede';
		const i = id.indexOf(':');
		const [entidad, clave] = [id.slice(0, i), id.slice(i + 1)];
		await borrarCambio(this.db, this.sb.id, entidad, clave);
		await guardarCambio(this.db, this.sb.id, 'deshecho', `${id}:${Date.parse(e.cuando)}`, {
			id: `${id}:${Date.parse(e.cuando)}`,
			texto: `${e.texto} — deshecho`,
			donde: e.donde,
			en: Date.parse(e.cuando)
		});
		return 'ok';
	}

	/** El día de Argentina (las tareas se marcan por día). */
	private dia() {
		return new Date(Date.now() - 3 * 3_600_000).toISOString().slice(0, 10);
	}

	private async tareasHechas() {
		const hoy = this.dia();
		const c = await cambiosDe(this.db, this.sb.id, 'tarea');
		return new Set(
			[...c].filter(([k, v]) => v && k.startsWith(`${hoy}:`)).map(([k]) => k.slice(11))
		);
	}

	async paraHacer(): Promise<ParaHacer> {
		const d = await this.base();
		const ahora = Date.now();
		const [vivosC, bandeja, carritos, tablero, productos, hechas] = await Promise.all([
			this.cambiosPedidos(),
			this.bandeja(),
			this.carritos(),
			cambiosDe(this.db, this.sb.id, 'tablero'),
			this.regProductos(),
			this.tareasHechas()
		]);
		const vivos = pedidosVivos(d, vivosC, ahora);
		const preparar = vivos.filter((p) => p.estado === 'Para preparar');
		const chats = bandeja.conversaciones.filter((c) => c.sinLeer > 0 || c.atiende === 'vos');
		const pendientes = carritos.carritos.filter((c) => !c.recordado);
		const transf = vivos.filter((p) => p.estado === 'Sin pagar' && p.pago === 'transf');
		const reponer = registrosInventario(productos).filter((p) =>
			['out', 'crit', 'low'].includes(p.st)
		);
		const conTicket = vivos.filter((p) => p.estado !== 'Cancelado');
		const ticket = conTicket.length
			? conTicket.reduce((s, p) => s + p.total, 0) / conTicket.length
			: 0;
		const pesosT = (n: number) => '$ ' + Math.round(n).toLocaleString('es-AR');
		type Tarea = ParaHacer['grupos'][number]['tareas'][number];
		const t = (x: Omit<Tarea, 'hecha'>, hay: boolean): Tarea[] =>
			hay ? [{ ...x, hecha: hechas.has(x.id) }] : [];
		const plural = (n: number, s: string, p: string) => `${n} ${n === 1 ? s : p}`;
		return {
			grupos: [
				{
					titulo: 'Urgente' as const,
					tareas: [
						...t(
							{
								id: 'preparar',
								tipo: 'preparar',
								titulo: `Despachar ${plural(preparar.length, 'pedido', 'pedidos')}`,
								detalle: 'Andreani retira a las 15:00',
								valor: preparar.reduce((s, p) => s + p.total, 0),
								accion: 'Preparar'
							},
							preparar.length > 0
						),
						...t(
							{
								id: 'chats',
								tipo: 'chats',
								titulo: `Responder ${plural(chats.length, 'chat', 'chats')}`,
								detalle: chats[0] ? `${chats[0].nombre.split(' ')[0]} espera respuesta` : '',
								valor: Math.round(chats.length * ticket),
								aproximado: true,
								accion: 'Abrir bandeja'
							},
							chats.length > 0
						)
					]
				},
				{
					titulo: 'Hoy' as const,
					tareas: [
						...t(
							{
								id: 'carritos',
								tipo: 'carritos',
								titulo: `Recuperar ${plural(pendientes.length, 'carrito', 'carritos')}`,
								detalle: `Recordatorio por WhatsApp · ${pesosT(pendientes.length * carritos.costoMensaje)}`,
								valor: pendientes.reduce((s, c) => s + c.total, 0),
								aproximado: true,
								accion: 'Enviar'
							},
							pendientes.length > 0
						),
						...t(
							{
								id: 'transferencias',
								tipo: 'transferencias',
								titulo: `Conciliar ${plural(transf.length, 'transferencia', 'transferencias')}`,
								detalle:
									transf.length === 1
										? `Pedido #${transf[0].numero}`
										: `Pedidos ${transf
												.slice(0, 3)
												.map((p) => `#${p.numero}`)
												.join(', ')
												.replace(/, ([^,]*)$/, ' y $1')}`,
								valor: transf.reduce((s, p) => s + p.total, 0),
								accion: 'Conciliar'
							},
							transf.length > 0
						)
					]
				},
				{
					titulo: 'Esta semana' as const,
					tareas: [
						...t(
							{
								id: 'certificado',
								tipo: 'certificado',
								titulo: 'Renovar certificado de ARCA',
								detalle: 'Vence en 5 días',
								valor: null,
								accion: 'Renovar'
							},
							!tablero.get('certificado')
						),
						...t(
							{
								id: 'reponer',
								tipo: 'reponer',
								titulo: `Reponer ${plural(reponer.length, 'producto', 'productos')}`,
								detalle: 'Tincho armó el pedido al proveedor',
								valor: null,
								accion: 'Revisar'
							},
							reponer.length > 0
						)
					]
				}
			].filter((g) => g.tareas.length)
		};
	}

	async marcarTarea(id: string, hecha: boolean): Promise<Resultado> {
		if (!['preparar', 'chats', 'carritos', 'transferencias', 'certificado', 'reponer'].includes(id))
			return 'no-existe';
		const clave = `${this.dia()}:${id}`;
		if (hecha) await guardarCambio(this.db, this.sb.id, 'tarea', clave, { en: Date.now() });
		else await borrarCambio(this.db, this.sb.id, 'tarea', clave);
		return 'ok';
	}

	// ---------------------------------------------------------------- Finanzas

	async finanzas(): Promise<Finanzas> {
		const [d, t, productos, retiros, campanas] = await Promise.all([
			this.demo(),
			this.tableroConMontos('7d'),
			this.regProductos(),
			cambiosDe(this.db, this.sb.id, 'retiro'),
			cambiosDe(this.db, this.sb.id, 'campana')
		]);
		const vendidos = productos.reduce((s, p) => s + p.vendidos30, 0);
		const margen = vendidos
			? Math.round(productos.reduce((s, p) => s + (p.margen ?? 0) * p.vendidos30, 0) / vendidos)
			: 38;
		return finanzasDemo(
			d,
			Date.now(),
			t,
			margen,
			[...retiros.values()].filter(Boolean).map((r) => r as Retiro),
			[...campanas.values()]
				.filter(Boolean)
				.map((c) => c as CampanaDemo)
				.filter((c) => !c.programada || c.programada <= Date.now())
				.map((c) => ({ personas: c.personas, en: c.programada ?? c.en })),
			COSTO_MENSAJE
		);
	}

	async retirar(): Promise<{ monto: number; cuenta: string } | 'nada'> {
		const f = await this.finanzas();
		if (!f.disponible) return 'nada';
		const en = Date.now();
		await guardarCambio(this.db, this.sb.id, 'retiro', String(en), {
			monto: f.disponible,
			en
		} satisfies Retiro);
		return { monto: f.disponible, cuenta: f.banco ?? 'tu banco' };
	}

	// ---------------------------------------------------------------- Configuración

	private async estadoCuenta(): Promise<EstadoCuenta> {
		const c = await cambiosDe(this.db, this.sb.id, 'cuenta');
		const g = <T>(k: string, def: T) => (c.get(k) as T | undefined) ?? def;
		return {
			negocio: g('negocio', {}),
			seguridad: g('seguridad', {}),
			avisos: g('avisos', {}),
			cerradas: new Set(g<string[]>('cerradas', [])),
			quitados: new Set(g<string[]>('quitados', [])),
			// «Depósito» pasó a «Local» (docs/decisiones.md §37): lo guardado antes se lee con el nombre nuevo.
			roles: new Map(
				Object.entries(g<Record<string, string>>('roles', {})).map(([k, r]) => [
					k,
					(r === 'deposito' ? 'local' : r) as Miembro['rol']
				])
			),
			invitados: g<Invitado[]>('invitados', []).map((i) => ({
				...i,
				rol: ((i.rol as string) === 'deposito' ? 'local' : i.rol) as Miembro['rol']
			})),
			plan: g('plan', {}),
			packs: g('packs', {}),
			cancelado: g('cancelado', null),
			oferta: g('oferta', null)
		};
	}

	async configuracion(): Promise<Configuracion> {
		const [d, t, e] = await Promise.all([this.demo(), this.tienda(), this.estadoCuenta()]);
		return configuracionDemo(d, t.nombre, Date.now(), DIAS_DE_PRUEBA, e);
	}

	async guardarNegocio(n: Negocio) {
		await guardarCambio(this.db, this.sb.id, 'cuenta', 'negocio', n);
	}

	async fijarOpcion(
		grupo: 'seguridad' | 'avisos',
		clave: string,
		activo: boolean
	): Promise<Resultado> {
		const c = await this.configuracion();
		if (!(clave in c[grupo])) return 'no-existe';
		await guardarCambio(this.db, this.sb.id, 'cuenta', grupo, {
			...(await this.estadoCuenta())[grupo],
			[clave]: activo
		});
		return 'ok';
	}

	async cerrarSesion(id: string): Promise<Resultado> {
		const s = (await this.configuracion()).sesiones.find((x) => x.id === id);
		if (!s) return 'no-existe';
		if (s.actual) return 'no-se-puede';
		const e = await this.estadoCuenta();
		await guardarCambio(this.db, this.sb.id, 'cuenta', 'cerradas', [...e.cerradas, id]);
		return 'ok';
	}

	async cerrarOtrasSesiones(): Promise<number> {
		const otras = (await this.configuracion()).sesiones.filter((s) => !s.actual).map((s) => s.id);
		if (!otras.length) return 0;
		const e = await this.estadoCuenta();
		await guardarCambio(this.db, this.sb.id, 'cuenta', 'cerradas', [...e.cerradas, ...otras]);
		return otras.length;
	}

	/** En la demo, quien prueba es la dueña de una sola tienda. */
	async acceso(): Promise<Acceso> {
		const t = await this.tienda();
		return {
			tiendaId: t.id,
			rol: this.rol,
			permisos: PERMISOS_DE[this.rol],
			tiendas: [{ id: t.id, nombre: t.nombre, slug: t.slug, rol: this.rol }]
		};
	}

	async elegirTienda(id: string): Promise<Resultado> {
		return id === (await this.tienda()).id ? 'ok' : 'no-existe';
	}

	/** En la demo no sale ningún mail: la invitación vuelve a vencer en 7 días desde ahora. */
	async reenviarInvitacion(id: string): Promise<Resultado> {
		const e = await this.estadoCuenta();
		if (!e.invitados.some((i) => i.id === id)) return 'no-existe';
		await guardarCambio(
			this.db,
			this.sb.id,
			'cuenta',
			'invitados',
			e.invitados.map((i) => (i.id === id ? { ...i, en: Date.now() } : i))
		);
		return 'ok';
	}

	async invitar(contacto: string, rol: RolInvitable): Promise<Miembro | 'repetido'> {
		const c = await this.configuracion();
		const v = contacto.trim().toLowerCase();
		if (c.equipo.some((m) => m.contacto.toLowerCase() === v)) return 'repetido';
		const e = await this.estadoCuenta();
		const nuevo: Invitado = {
			id: `i${Date.now().toString(36)}`,
			contacto: contacto.trim(),
			rol,
			en: Date.now()
		};
		await guardarCambio(this.db, this.sb.id, 'cuenta', 'invitados', [...e.invitados, nuevo]);
		return (await this.configuracion()).equipo.find((m) => m.id === nuevo.id)!;
	}

	async cambiarPlan(plan: Plan, ciclo?: 'mes' | 'anio') {
		const e = await this.estadoCuenta();
		await guardarCambio(this.db, this.sb.id, 'cuenta', 'plan', {
			...e.plan,
			actual: plan,
			...(ciclo ? { ciclo } : {})
		});
	}

	// ---------------------------------------------------------------- Cuenta

	private mes() {
		return new Date(Date.now() - 3 * 3_600_000).toISOString().slice(0, 7);
	}

	async planCuenta(): Promise<PlanCuenta> {
		const [c, e, productos, uso] = await Promise.all([
			this.configuracion(),
			this.estadoCuenta(),
			this.regProductos(),
			cambiosDe(this.db, this.sb.id, 'hornero')
		]);
		const usado = (uso.get('uso') as { cuantos?: number } | undefined)?.cuantos ?? 0;
		const fin = Date.now() + DIAS_DE_PRUEBA * 86_400_000;
		return {
			actual: c.plan.actual,
			ciclo: c.plan.ciclo,
			enPrueba: c.plan.enPrueba,
			diasPrueba: c.plan.diasPrueba,
			proximoCobro: new Date(fin).toISOString(),
			uso: {
				productos: productos.length,
				personas: c.equipo.length,
				// La tienda de ejemplo viene con su historia; se suma lo que le preguntó el visitante.
				hornero: 1842 + usado,
				packs: e.packs[this.mes()] ?? 0,
				sinIa: 3912
			},
			medio: { marca: 'visa', ultimos: '4821', vence: '08/29' },
			facturacion: { razonSocial: c.negocio.razonSocial, cuit: c.negocio.cuit, tipo: 'Factura A' },
			// Está en prueba: la primera factura de Tu Local Web llega cuando termina.
			facturas: [],
			cancelado: e.cancelado ? { hasta: new Date(fin).toISOString() } : null
		};
	}

	async comprarPack() {
		const e = await this.estadoCuenta();
		const mes = this.mes();
		await guardarCambio(this.db, this.sb.id, 'cuenta', 'packs', {
			...e.packs,
			[mes]: (e.packs[mes] ?? 0) + 1
		});
	}

	async devolverPack(): Promise<Resultado> {
		const e = await this.estadoCuenta();
		const mes = this.mes();
		if (!e.packs[mes]) return 'no-existe';
		await guardarCambio(this.db, this.sb.id, 'cuenta', 'packs', {
			...e.packs,
			[mes]: e.packs[mes] - 1
		});
		return 'ok';
	}

	async cancelarPlan(motivo: number, accion: 'cancelar' | 'oferta') {
		const en = Date.now();
		if (accion === 'oferta')
			await guardarCambio(this.db, this.sb.id, 'cuenta', 'oferta', { motivo, en });
		else await guardarCambio(this.db, this.sb.id, 'cuenta', 'cancelado', { motivo, en });
		return { hasta: new Date(en + DIAS_DE_PRUEBA * 86_400_000).toISOString() };
	}

	async reactivarPlan() {
		await guardarCambio(this.db, this.sb.id, 'cuenta', 'cancelado', null);
	}

	async cambiarRol(id: string, rol: RolInvitable): Promise<Resultado> {
		const m = (await this.configuracion()).equipo.find((x) => x.id === id);
		if (!m) return 'no-existe';
		if (m.vos || m.rol === 'duenio') return 'no-se-puede';
		const e = await this.estadoCuenta();
		await guardarCambio(this.db, this.sb.id, 'cuenta', 'roles', {
			...Object.fromEntries(e.roles),
			[id]: rol
		});
		return 'ok';
	}

	async quitarDelEquipo(id: string): Promise<Resultado> {
		const m = (await this.configuracion()).equipo.find((x) => x.id === id);
		if (!m) return 'no-existe';
		if (m.vos) return 'no-se-puede';
		const e = await this.estadoCuenta();
		if (m.pendiente)
			await guardarCambio(
				this.db,
				this.sb.id,
				'cuenta',
				'invitados',
				e.invitados.filter((i) => i.id !== id)
			);
		else await guardarCambio(this.db, this.sb.id, 'cuenta', 'quitados', [...e.quitados, id]);
		return 'ok';
	}

	// ---------------------------------------------------------------- Dominio, avisos y mis datos

	async dominios(): Promise<Dominios> {
		const [t, c] = await Promise.all([this.tienda(), cambiosDe(this.db, this.sb.id, 'cuenta')]);
		const lista = (c.get('dominios') as DominioDemo[] | undefined) ?? [];
		const ahora = Date.now();
		const gratis = `${SITE.replace(/^https?:\/\//, '')}/t/${t.slug}`;
		const principal = (c.get('principal') as string | undefined) ?? gratis;
		// En la demo los registros "aparecen" a los 1, 3 y 5 minutos de agregar el dominio.
		const paso = (d: DominioDemo, min: number) => ahora - d.en > min * 60_000;
		return {
			gratis,
			dominios: [
				{
					nombre: gratis,
					detalle: 'La dirección gratis de tu tienda',
					estado: 'activo',
					principal: principal === gratis
				},
				...lista.map((d) => {
					const listo = paso(d, 5);
					return {
						nombre: d.nombre,
						detalle: listo ? 'Conectado' : 'Agregado recién · revisando DNS',
						estado: listo ? ('activo' as const) : ('verificando' as const),
						principal: principal === d.nombre,
						chequeos: [
							{
								titulo: 'Registro CNAME de www',
								estado: paso(d, 1) ? ('ok' as const) : ('esperando' as const)
							},
							{
								titulo: 'Verificación TXT',
								estado: paso(d, 3)
									? ('ok' as const)
									: paso(d, 1)
										? ('esperando' as const)
										: ('pendiente' as const)
							},
							{
								titulo: 'Certificado SSL',
								estado: listo ? ('ok' as const) : ('pendiente' as const)
							}
						]
					};
				})
			],
			registros: [
				{ tipo: 'CNAME', nombre: 'www', valor: `tiendas.${SITE.replace(/^https?:\/\//, '')}` },
				{ tipo: 'TXT', nombre: '_tulocalweb', valor: `tulocalweb-verificacion=${t.slug}` }
			]
		};
	}

	async conectarDominio(nombre: string): Promise<Resultado> {
		const c = await cambiosDe(this.db, this.sb.id, 'cuenta');
		const lista = (c.get('dominios') as DominioDemo[] | undefined) ?? [];
		if (lista.some((d) => d.nombre === nombre)) return 'no-se-puede';
		await guardarCambio(this.db, this.sb.id, 'cuenta', 'dominios', [
			...lista,
			{ nombre, en: Date.now() }
		]);
		return 'ok';
	}

	async dominioPrincipal(nombre: string): Promise<Resultado> {
		const d = (await this.dominios()).dominios.find((x) => x.nombre === nombre);
		if (!d) return 'no-existe';
		if (d.estado !== 'activo') return 'no-se-puede';
		await guardarCambio(this.db, this.sb.id, 'cuenta', 'principal', nombre);
		return 'ok';
	}

	async quitarDominio(nombre: string): Promise<Resultado> {
		const c = await cambiosDe(this.db, this.sb.id, 'cuenta');
		const lista = (c.get('dominios') as DominioDemo[] | undefined) ?? [];
		if (!lista.some((d) => d.nombre === nombre)) return 'no-existe';
		await guardarCambio(
			this.db,
			this.sb.id,
			'cuenta',
			'dominios',
			lista.filter((d) => d.nombre !== nombre)
		);
		if (c.get('principal') === nombre)
			await guardarCambio(this.db, this.sb.id, 'cuenta', 'principal', null);
		return 'ok';
	}

	// ---------------------------------------------------------------- Mercado Pago

	async integraciones(): Promise<EstadoIntegraciones> {
		const [c, mp] = await Promise.all([
			cambiosDe(this.db, this.sb.id, 'integraciones'),
			this.mercadoPago()
		]);
		const e = estadoInicial(true);
		const g = (k: keyof EstadoIntegraciones) =>
			(c.get(k) as Record<string, never> | undefined) ?? {};
		Object.assign(e.conectadas, g('conectadas'));
		for (const [id, o] of Object.entries(g('opciones') as Record<string, Record<string, boolean>>))
			Object.assign((e.opciones[id] ??= {}), o);
		Object.assign(e.ia, g('ia'));
		Object.assign(e.permisosIa, g('permisosIa'));
		e.conectadas.mp = mp.conectada;
		e.conectadas.ml = (await this.mercadoLibre()).conectada;
		Object.assign(e.conectadas, conectadasDeMedicion(medicionDe(c)));
		const arca = await this.arca();
		e.conectadas.arca = arca.conectada;
		e.opciones.arca = { ...e.opciones.arca, auto: arca.facturarAlCobrar };
		// Telegram cuenta si hay un chat vinculado (lib/server/telegram.ts).
		e.conectadas.tg = !!(await cambiosDe(this.db, this.sb.id, 'cuenta')).get('telegram');
		return e;
	}
	async ajustarIntegracion(a: AjusteIntegracion): Promise<Resultado> {
		const e = await this.integraciones();
		if (a.tipo === 'conexion') {
			if (
				a.id === 'mp' ||
				a.id === 'ml' ||
				a.id === 'tg' ||
				(DE_MEDICION as readonly string[]).includes(a.id) ||
				a.id === 'arca' ||
				!INTEGRACIONES.some((i) => i.id === a.id && !i.pronto)
			)
				return 'no-se-puede';
			e.conectadas[a.id] = a.valor;
			await guardarCambio(this.db, this.sb.id, 'integraciones', 'conectadas', e.conectadas);
		} else if (a.tipo === 'opcion') {
			if (!(a.clave in (e.opciones[a.id] ?? {}))) return 'no-existe';
			e.opciones[a.id][a.clave] = a.valor;
			await guardarCambio(this.db, this.sb.id, 'integraciones', 'opciones', e.opciones);
		} else if (a.tipo === 'ia') {
			if (!(a.id in e.ia)) return 'no-existe';
			e.ia[a.id] = a.valor;
			await guardarCambio(this.db, this.sb.id, 'integraciones', 'ia', e.ia);
		} else {
			if (!(a.id in e.permisosIa) || a.id === 'leer') return 'no-se-puede';
			e.permisosIa[a.id] = a.valor;
			await guardarCambio(this.db, this.sb.id, 'integraciones', 'permisosIa', e.permisosIa);
		}
		return 'ok';
	}

	// ---------------------------------------------------------------- ARCA

	/** Lo que el visitante cambió de ARCA: punto de venta y cuándo prendió o apagó facturar al cobrar. */
	private async arcaDemo(): Promise<ArcaDemo> {
		const c = (await cambiosDe(this.db, this.sb.id, 'cuenta')).get('arca') as ArcaDemo | undefined;
		return { puntoVenta: c?.puntoVenta ?? 4, historial: c?.historial ?? [] };
	}
	/**
	 * La demo arranca conectada en homologación con un certificado de ejemplo, para que Facturación
	 * muestre comprobantes como siempre. Lo que no se puede simular sin mentir (guardar un
	 * certificado, hablar con ARCA) responde que es una tienda de ejemplo.
	 */
	async arca(): Promise<ArcaConfig> {
		const a = await this.arcaDemo();
		return {
			conectada: true,
			cuit: CUIT_DEMO,
			razonSocial: 'Valeria Sosa',
			condicion: 'responsable_inscripto',
			puntoVenta: a.puntoVenta,
			ambiente: 'homologacion',
			certificado: true,
			certificadoVence: new Date(this.sb.creadoEn + 2 * 365 * 86_400_000).toISOString(),
			facturarAlCobrar: alCobrarEn(a.historial, Date.now())
		};
	}
	/** Valida como Go; la clave no se genera acá (CPU de Workers Free): el navegador arma el CSR. */
	async arcaPedido(a: ArcaAlta): Promise<ArcaPedido> {
		const e = validarAlta(a);
		return e ? { error: e } : { demo: true };
	}
	async arcaCertificado(pem: string): Promise<ArcaCertificado | { error: string }> {
		if (!/-----BEGIN CERTIFICATE-----[\s\S]+-----END CERTIFICATE-----/.test(pem))
			return {
				error:
					'Eso no es un certificado: pegá el archivo .crt que te dio ARCA, desde -----BEGIN CERTIFICATE-----.'
			};
		return { error: DEMO_ARCA };
	}
	async arcaAjustes(a: ArcaAjustes): Promise<{ error: string } | null> {
		if (
			a.puntoVenta !== undefined &&
			(!Number.isInteger(a.puntoVenta) || a.puntoVenta < 1 || a.puntoVenta > 99998)
		)
			return { error: 'El punto de venta va de 1 a 99998.' };
		const antes = await this.arcaDemo();
		const historial =
			a.facturarAlCobrar === undefined
				? antes.historial
				: [...antes.historial, [Date.now(), a.facturarAlCobrar] as [number, boolean]].slice(-20);
		await guardarCambio(this.db, this.sb.id, 'cuenta', 'arca', {
			puntoVenta: a.puntoVenta ?? antes.puntoVenta,
			historial
		} satisfies ArcaDemo);
		return null;
	}
	async arcaProbar(): Promise<ArcaPrueba> {
		return { ok: false, ultimo: 0, detalle: DEMO_ARCA };
	}

	/** La demo arranca sin IDs (y con el feed de Google cargado, como la tienda del handoff). */
	async medicion(): Promise<Medicion> {
		return medicionDe(await cambiosDe(this.db, this.sb.id, 'integraciones'));
	}
	async guardarMedicion(m: Medicion): Promise<Resultado> {
		const ga4 = validarGa4(m.ga4);
		const pixel = validarPixel(m.pixel);
		if (!ga4.ok || !pixel.ok) return 'no-se-puede';
		await guardarCambio(this.db, this.sb.id, 'integraciones', 'medicion', {
			ga4: ga4.valor,
			pixel: pixel.valor,
			feeds: { gm: !!m.feeds.gm, meta: !!m.feeds.meta }
		} satisfies Medicion);
		return 'ok';
	}

	/** Envío gratis desde el monto de la promoción automática, si está prendida (como en Go). */
	private async envioGratisDesde(): Promise<number | null> {
		const autos = await cambiosDe(this.db, this.sb.id, 'automatica');
		const envio = autos.get('envio') as { activa: boolean } | null | undefined;
		return (envio?.activa ?? true) ? ENVIO_GRATIS_DESDE : null;
	}

	/**
	 * Envíos y retiro (como configuracion.LeerEntregas): sin configurar, retiro con la dirección y el
	 * horario del negocio y sin moto. La demo no tiene cuenta de MiCorreo: Correo nunca está conectado.
	 */
	async entregas(): Promise<EstadoEntregas> {
		const [c, conf, gratisDesde] = await Promise.all([
			cambiosDe(this.db, this.sb.id, 'entregas'),
			this.configuracion(),
			this.envioGratisDesde()
		]);
		const e = (c.get('config') as Entregas | undefined) ?? {
			retiro: true,
			retiroDireccion: '',
			retiroHorario: '',
			moto: false,
			motoCosto: 0,
			motoZona: '',
			motoDemora: ''
		};
		return {
			...e,
			retiroDireccion: e.retiroDireccion || conf.negocio.direccion || '',
			retiroHorario: e.retiroHorario || conf.negocio.horario || '',
			correo: false,
			gratisDesde
		};
	}
	async guardarEntregas(e: Entregas): Promise<void> {
		const r = v.safeParse(EsquemaEntregas, e);
		if (!r.success) error(400, r.issues[0].message);
		const conf = await this.configuracion();
		const mal = problemaEntregas(r.output, {
			correo: false,
			direccionNegocio: conf.negocio.direccion ?? ''
		});
		if (mal) error(400, mal);
		await guardarCambio(this.db, this.sb.id, 'entregas', 'config', r.output satisfies Entregas);
	}

	/** En la demo Mercado Libre arranca conectado (como en el handoff) y todo se simula. */
	async mercadoLibre(): Promise<ConexionML> {
		const c = await cambiosDe(this.db, this.sb.id, 'cuenta');
		const ml = c.get('ml') as
			{ conectada: boolean; desde: number; recargo?: number; pausar?: boolean } | undefined;
		const conectada = ml?.conectada ?? true;
		const productos = await this.regProductos();
		return {
			disponible: true,
			conectada,
			...(conectada
				? {
						apodo: 'MATEYCO_OFICIAL',
						desde: new Date(ml?.desde ?? Date.now() - 60 * 86_400_000).toISOString()
					}
				: {}),
			reautorizar: false,
			recargo: ml?.recargo ?? 10,
			pausarSinStock: ml?.pausar ?? true,
			publicaciones: conectada ? productos.filter((p) => p.canales.includes('ml')).length : 0,
			preguntas: conectada ? 3 : 0
		};
	}
	async conectarMercadoLibre() {
		return '/integraciones/mercadolibre/callback?code=demo&state=demo';
	}
	async volverMercadoLibre(codigo: string, estado: string): Promise<Resultado> {
		if (codigo !== 'demo' || estado !== 'demo') return 'no-se-puede';
		const c = await cambiosDe(this.db, this.sb.id, 'cuenta');
		const ml = (c.get('ml') as Record<string, unknown> | undefined) ?? {};
		await guardarCambio(this.db, this.sb.id, 'cuenta', 'ml', {
			...ml,
			conectada: true,
			desde: Date.now()
		});
		return 'ok';
	}
	async desconectarMercadoLibre() {
		const c = await cambiosDe(this.db, this.sb.id, 'cuenta');
		const ml = (c.get('ml') as Record<string, unknown> | undefined) ?? {};
		await guardarCambio(this.db, this.sb.id, 'cuenta', 'ml', { ...ml, conectada: false });
	}
	async ajustarMercadoLibre(recargo: number, pausar: boolean): Promise<Resultado> {
		if (!(await this.mercadoLibre()).conectada) return 'no-se-puede';
		const c = await cambiosDe(this.db, this.sb.id, 'cuenta');
		const ml = (c.get('ml') as Record<string, unknown> | undefined) ?? {};
		await guardarCambio(this.db, this.sb.id, 'cuenta', 'ml', { ...ml, recargo, pausar });
		return 'ok';
	}
	async publicarEnMercadoLibre(productoId: string): Promise<PublicacionML | { error: string }> {
		if (!(await this.mercadoLibre()).conectada) return { error: 'Conectá Mercado Libre primero.' };
		const n = parseInt(productoId.replace(/\D/g, '').slice(-6) || '1', 10);
		const item = `MLA${1_400_000_000 + n}`;
		return { itemId: item, permalink: `https://articulo.mercadolibre.com.ar/MLA-${item.slice(3)}` };
	}
	async despublicarDeMercadoLibre() {}

	/** En la demo la tienda arranca con Mercado Pago conectado (el onboarding lo conecta). */
	async mercadoPago(): Promise<ConexionMP> {
		const c = await cambiosDe(this.db, this.sb.id, 'cuenta');
		const mp = c.get('mp') as { conectada: boolean; desde: number } | undefined;
		const conectada = mp?.conectada ?? true;
		return {
			disponible: true,
			conectada,
			...(conectada
				? {
						usuario: '3666528660',
						desde: new Date(mp?.desde ?? Date.now() - 30 * 86_400_000).toISOString()
					}
				: {}),
			reautorizar: false,
			cajaQR: conectada
		};
	}
	/** La demo no sale a Mercado Pago: vuelve directo con un código de mentira. */
	async conectarMercadoPago() {
		return '/integraciones/mercadopago/callback?code=demo&state=demo';
	}
	async volverMercadoPago(codigo: string, estado: string): Promise<Resultado> {
		if (codigo !== 'demo' || estado !== 'demo') return 'no-se-puede';
		await guardarCambio(this.db, this.sb.id, 'cuenta', 'mp', {
			conectada: true,
			desde: Date.now()
		});
		return 'ok';
	}
	async desconectarMercadoPago() {
		await guardarCambio(this.db, this.sb.id, 'cuenta', 'mp', { conectada: false, desde: 0 });
	}

	async preferenciasAvisos(): Promise<PreferenciasAvisos> {
		const c = await cambiosDe(this.db, this.sb.id, 'cuenta');
		return (
			(c.get('avisos-canales') as PreferenciasAvisos | undefined) ?? {
				matriz: {},
				silencio: true,
				resumen: 'whatsapp'
			}
		);
	}

	async guardarPreferenciasAvisos(p: PreferenciasAvisos) {
		await guardarCambio(this.db, this.sb.id, 'cuenta', 'avisos-canales', p);
	}

	// La demo no manda avisos al celular: sin clave, el panel dice que no está disponible.
	async pushWeb(): Promise<EstadoPushWeb> {
		return { clave: '', suscripta: false };
	}
	async suscribirPushWeb(): Promise<Resultado> {
		return 'no-se-puede';
	}
	async bajaPushWeb() {}
	async probarPushWeb(): Promise<Resultado> {
		return 'no-se-puede';
	}

	private async inactivos() {
		// Sin compras en 3 años: en la demo, con dos meses de historia, no hay.
		return (await this.regClientes()).filter((c) => c.ultima > 3 * 365);
	}

	async misDatos(): Promise<MisDatos> {
		const [productos, clientes, d, facturas, conv, exps, c] = await Promise.all([
			this.regProductos(),
			this.regClientes(),
			this.base(),
			this.regFacturacion(),
			this.bandeja(),
			cambiosDe(this.db, this.sb.id, 'exportacion'),
			cambiosDe(this.db, this.sb.id, 'cuenta')
		]);
		return {
			conteos: {
				productos: productos.length,
				clientes: clientes.length,
				pedidos: d.pedidos.filter((p) => p.creado <= Date.now()).length,
				facturas: facturas.length,
				resenas: d.resenas.length,
				conversaciones: conv.conversaciones.length
			},
			recientes: [...exps.values()]
				.filter(Boolean)
				.map((x) => x as { tipo: string; formato: string; en: number })
				.sort((a, b) => b.en - a.en)
				.slice(0, 5)
				.map((x) => ({ ...x, en: new Date(x.en).toISOString() })),
			inactivos: clientes.filter((x) => x.ultima > 3 * 365).length,
			borrado: (c.get('borrado') as MisDatos['borrado'] | undefined) ?? null
		};
	}

	/** Filas de cada exportación. */
	async exportar(tipo: TipoExport): Promise<(string | number | null)[][]> {
		const fecha = (t: number) => new Date(t).toISOString().slice(0, 10);
		if (tipo === 'productos')
			return [
				['SKU', 'Nombre', 'Categoría', 'Precio', 'Costo', 'Stock', 'Publicado'],
				...(await this.regProductos()).map((p) => [
					p.sku,
					p.nombre,
					p.categoria,
					p.precio,
					p.costo,
					p.stock,
					p.publicado ? 'Sí' : 'No'
				])
			];
		if (tipo === 'clientes')
			return [
				['Nombre', 'Ciudad', 'Teléfono', 'Pedidos', 'Gastado'],
				...(await this.regClientes()).map((c) => [
					c.nombre,
					c.ciudad,
					c.telefono,
					c.pedidos,
					c.gastado
				])
			];
		if (tipo === 'facturas')
			return [
				['Número', 'Tipo', 'Cliente', 'Total', 'Estado', 'Pedido'],
				...(await this.regFacturacion()).map((f) => [
					f.numero,
					f.tipo,
					f.cliente,
					f.total,
					f.estado,
					f.pedido
				])
			];
		const d = await this.base();
		const nombre = new Map(d.clientes.map((c) => [c.id, c.nombre]));
		if (tipo === 'pedidos')
			return [
				['Número', 'Fecha', 'Cliente', 'Canal', 'Pago', 'Envío', 'Estado', 'Total'],
				...d.pedidos
					.filter((p) => p.creado <= Date.now())
					.map((p) => [
						p.numero,
						fecha(p.creado),
						nombre.get(p.clienteId) ?? '',
						p.canal,
						p.pago,
						p.envio,
						p.estado,
						p.total
					])
			];
		if (tipo === 'resenas')
			return [
				['Fecha', 'Cliente', 'Producto', 'Estrellas', 'Reseña'],
				...d.resenas.map((r) => [
					fecha(r.creada),
					nombre.get(r.clienteId) ?? '',
					r.producto,
					r.estrellas,
					r.texto
				])
			];
		const filas: (string | number | null)[][] = [
			['Conversación', 'Cliente', 'De', 'Mensaje', 'Fecha']
		];
		for (const c of (await this.bandeja()).conversaciones) {
			const conv = await this.conversacion(c.id);
			for (const m of conv?.mensajes ?? [])
				if (m.texto) filas.push([c.id, c.nombre, m.de, m.texto, m.en.slice(0, 16)]);
		}
		return filas;
	}

	async registrarExportacion(tipo: TipoExport | 'todo', formato: 'csv' | 'excel') {
		const en = Date.now();
		await guardarCambio(this.db, this.sb.id, 'exportacion', String(en), { tipo, formato, en });
	}

	async borrarInactivos(): Promise<number> {
		const n = (await this.inactivos()).length;
		await guardarCambio(
			this.db,
			this.sb.id,
			'cuenta',
			'borrado',
			n ? { clientes: n, el: new Date(Date.now() + 7 * 86_400_000).toISOString() } : null
		);
		return n;
	}

	async frenarBorrado() {
		await guardarCambio(this.db, this.sb.id, 'cuenta', 'borrado', null);
	}

	/** Cerrar la cuenta en la demo: se borra todo lo del visitante y vuelve a empezar. */
	async cerrarCuenta(confirmacion: string): Promise<Resultado> {
		const t = await this.tienda();
		if (confirmacion.trim().toLowerCase() !== t.nombre.toLowerCase()) return 'no-se-puede';
		await this.db.batch([
			this.db.delete(cambios).where(eq(cambios.sandboxId, this.sb.id)),
			this.db.delete(sandboxes).where(eq(sandboxes.id, this.sb.id))
		]);
		return 'ok';
	}

	// ---------------------------------------------------------------- Mi tienda

	async miTienda(): Promise<MiTienda> {
		const [d, t, c] = await Promise.all([
			this.demo(),
			this.tienda(),
			cambiosDe(this.db, this.sb.id, 'mi-tienda')
		]);
		const borrador = (c.get('borrador') as DisenoTienda | undefined) ?? DISENO_INICIAL;
		const publicado = (c.get('publicado') as DisenoTienda | undefined) ?? DISENO_INICIAL;
		// En la demo la plantilla de la tienda fija su nombre y su dirección: la elegida en Mi tienda
		// queda en el diseño y se usa para dibujarla.
		const plantillaId = borrador.plantilla || publicado.plantilla || t.plantilla;
		const p = buscarPlantilla(plantillaId);
		const vendidos = new Map<string, number>();
		for (const x of d.pedidos)
			for (const i of x.items)
				vendidos.set(i.productoId, (vendidos.get(i.productoId) ?? 0) + i.cantidad);
		const top = [...d.catalogo.productos].sort(
			(a, b) => (vendidos.get(b.id) ?? 0) - (vendidos.get(a.id) ?? 0)
		);
		const buenas = d.resenas.filter((r) => r.estrellas === 5 && r.texto.length > 30);
		const promedio = d.resenas.length
			? d.resenas.reduce((s, r) => s + r.estrellas, 0) / d.resenas.length
			: 0;
		return {
			plantilla: p?.nombre ?? plantillaId,
			plantillaId,
			direccion: `${SITE.replace(/^https?:\/\//, '')}/t/${t.slug}`,
			diseno: borrador,
			sinPublicar: JSON.stringify(borrador) !== JSON.stringify(publicado),
			vista: {
				tienda: t.nombre,
				categorias: d.catalogo.categorias,
				productos: top
					.slice(0, 4)
					.map((x) => ({ nombre: x.nombre, foto: x.foto, precio: x.precio })),
				resena: d.resenas.length
					? { promedio: Math.round(promedio * 10) / 10, texto: (buenas[0] ?? d.resenas[0]).texto }
					: null,
				fotos: [...new Set(top.flatMap((x) => (x.foto ? [x.foto] : [])))].slice(0, 6),
				portada: {
					foto: borrador.portada || (p?.img ?? top[0]?.foto ?? null),
					titulo: borrador.titulo || (p?.lema ?? t.nombre),
					lema: borrador.lema ?? ''
				},
				logo: borrador.logo ?? ''
			}
		};
	}

	/**
	 * La tienda online como la ve quien compra (lo publicado en Mi tienda). vitrina: es la tienda de
	 * ejemplo de una plantilla para mirar, sin carrito real.
	 */
	// ---------------------------------------------------------------- Lanzamientos (§52)

	private async lanzamientos() {
		const c = await cambiosDe(this.db, this.sb.id, 'lanzamiento');
		return new Map([...c].filter(([, v]) => v).map(([k, v]) => [k, v as ConfigLanzamiento]));
	}
	private async avisosLanzamiento() {
		return [...(await cambiosDe(this.db, this.sb.id, 'aviso-lanzamiento')).keys()];
	}

	/** Lanzamiento, señas y oferta del producto, con cuántos pidieron aviso y cuántos reservaron. */
	async lanzamiento(id: string): Promise<LanzamientoProducto | null> {
		const [productos, config, avisos, d] = await Promise.all([
			this.regProductos(),
			this.lanzamientos(),
			this.avisosLanzamiento(),
			this.demo()
		]);
		if (!productos.some((p) => p.id === id)) return null;
		const c = config.get(id) ?? SIN_LANZAMIENTO;
		const conSena = d.pedidos.filter(
			(p) => p.sena && p.estado !== 'Cancelado' && p.items.some((i) => i.productoId === id)
		);
		const suyos = avisos.filter((k) => k.startsWith(`${id}:`)).length;
		return {
			...c,
			avisos: suyos,
			// La demo no manda mails: a la hora del lanzamiento, todos cuentan como avisados.
			avisados: c.lanzamiento && Date.parse(c.lanzamiento) <= Date.now() ? suyos : 0,
			reservas: conSena.filter((p) => p.sena!.tipo === 'reserva').length,
			encargos: conSena.filter((p) => p.sena!.tipo === 'encargo').length
		};
	}

	async guardarLanzamiento(
		id: string,
		c: ConfigLanzamiento
	): Promise<LanzamientoProducto | { error: string } | null> {
		const p = (await this.regProductos()).find((x) => x.id === id);
		if (!p) return null;
		const antes = (await this.lanzamientos()).get(id) ?? SIN_LANZAMIENTO;
		const mal = errorLanzamiento(c, antes, p.precio);
		if (mal) return { error: mal };
		await guardarCambio(this.db, this.sb.id, 'lanzamiento', id, c);
		return this.lanzamiento(id);
	}

	/** «Avisame cuando salga» en la demo: se anota (no sale ningún mail). */
	async avisarLanzamiento(
		productoId: string,
		email: string
	): Promise<'ok' | 'no-existe' | 'ya-salio' | 'lleno'> {
		const t = await this.tiendaPublica(false);
		const p = t.productos.find((x) => x.id === productoId);
		if (!p) return 'no-existe';
		if (!proximamente(p)) return 'ya-salio';
		const avisos = await this.avisosLanzamiento();
		if (avisos.filter((k) => k.startsWith(`${productoId}:`)).length >= 200) return 'lleno';
		await guardarCambio(
			this.db,
			this.sb.id,
			'aviso-lanzamiento',
			`${productoId}:${email.toLowerCase()}`,
			{
				en: Date.now()
			}
		);
		return 'ok';
	}

	/**
	 * Reservar o encargar con seña en la demo: el pedido se arma igual que en Go (precio y seña del
	 * servidor, descuenta stock) y la seña se da por pagada, porque la demo no cobra.
	 */
	async senar(s: {
		productoId: string;
		varianteId: string | null;
		cantidad: number;
		nombre: string;
		email: string;
		telefono: string;
	}): Promise<
		{ numero: number; total: number; sena: number } | 'sin-stock' | 'invalido' | 'sin-sena'
	> {
		const t = await this.tiendaPublica(false);
		const p = t.productos.find((x) => x.id === s.productoId);
		if (!p) return 'invalido';
		const antes = proximamente(p);
		const pct = antes ? p.reserva : p.encargo;
		if (!pct) return 'sin-sena';
		const v = p.opciones.length ? p.variantes.find((x) => x.id === s.varianteId) : undefined;
		if (p.opciones.length ? !v : s.varianteId) return 'invalido';
		const precio = v ? precioDe(v, p.precio) : p.precio;
		const total = precio * s.cantidad;
		const sena = senaDe(total, pct);
		const tipo = antes ? 'reserva' : 'encargo';
		const r = await this.crearPedido(
			{
				items: [{ productoId: p.id, varianteId: v?.id ?? null, cantidad: s.cantidad }],
				pago: 'mp',
				canal: 'web',
				envio: 'retiro',
				cliente: s.nombre,
				telefono: s.telefono,
				nota: [
					`Email: ${s.email}`,
					`${tipo === 'reserva' ? 'Reserva del lanzamiento' : 'Encargo'} con seña: $ ${sena.toLocaleString('es-AR')} ahora, $ ${(total - sena).toLocaleString('es-AR')} al retirar`
				].join(' · '),
				pagado: false
			},
			{
				email: s.email,
				cupon: null,
				descuentoCupon: 0,
				total,
				precios: { [claveStock(p.id, v?.id)]: precio },
				sena: { tipo, monto: sena }
			}
		);
		if (typeof r === 'string') return r;
		return { ...r, sena };
	}

	/** «Saldado»: cobró el resto de un pedido con la seña pagada (o lo deshace). */
	async saldarPedido(numero: number, saldado: boolean): Promise<Resultado> {
		const n = (await cambiosDe(this.db, this.sb.id, 'pedido-nuevo')).get(String(numero)) as
			PedidoNuevoDemo | undefined;
		if (!n) return 'no-existe';
		const vivo = (await this.demo()).pedidos.find((x) => x.numero === numero);
		if (!n.pedido.sena?.pagada || !vivo || vivo.estado === 'Cancelado') return 'no-se-puede';
		await guardarCambio(this.db, this.sb.id, 'pedido-nuevo', String(numero), {
			...n,
			pedido: {
				...n.pedido,
				estadoPago: saldado ? 'Cobrado' : 'Seña',
				sena: { ...n.pedido.sena, saldado }
			}
		} satisfies PedidoNuevoDemo);
		return 'ok';
	}

	async tiendaPublica(vitrina = false): Promise<TiendaPublica> {
		const [d, t, c, conf, productos, mp, turnos, medicion, cr, respuestasQ, entregas, planes, lanz] =
			await Promise.all([
				this.demo(),
				this.tienda(),
				cambiosDe(this.db, this.sb.id, 'mi-tienda'),
				this.configuracion(),
				this.regProductos(),
				this.mercadoPago(),
				this.turnosTienda(),
				vitrina ? null : this.medicion(),
				this.cambiosResenas(),
				this.respuestasPreguntas(),
				this.entregas(),
				this.estadoSuscripciones(),
				this.lanzamientos()
			]);
		const diseno = (c.get('publicado') as DisenoTienda | undefined) ?? DISENO_INICIAL;
		const plantillaId = diseno.plantilla || t.plantilla;
		const p = buscarPlantilla(plantillaId);
		const visibles = productos.filter((x) => x.publicado && x.canales.includes('web'));
		const vendidos = new Map(visibles.map((x) => [x.id, x.vendidos30]));
		const masVendidos = [...visibles]
			.sort((a, b) => (vendidos.get(b.id) ?? 0) - (vendidos.get(a.id) ?? 0))
			.map((x) => x.id);
		const delCatalogo = new Map(d.catalogo.productos.map((x) => [x.id, x]));
		const categorias = [...new Set(visibles.map((x) => x.categoria).filter(Boolean))].map(
			(nombre) => {
				const de = visibles.filter((x) => x.categoria === nombre);
				return { nombre, cantidad: de.length, foto: de.find((x) => x.foto)?.foto ?? null };
			}
		);
		// Las del visitante se ven todas (como en Go); de las de ejemplo, las seis mejores escritas.
		cr.productos = new Map(productos.map((x) => [x.id, { nombre: x.nombre, foto: x.foto }]));
		const aLaVista = todasLasResenas(d, cr).filter((r) => !r.oculta);
		const resenas = [
			...aLaVista.filter((r) => !r.ejemplo),
			...aLaVista.filter((r) => r.ejemplo && r.estrellas >= 4 && r.texto.length > 25).slice(0, 6)
		].map((r) => ({
			texto: r.texto,
			autor: r.firma,
			producto: r.producto ?? '',
			estrellas: r.estrellas,
			conCompra: r.conCompra,
			respuesta: r.respuesta
		}));
		const promedio = aLaVista.length
			? Math.round((aLaVista.reduce((s, r) => s + r.estrellas, 0) / aLaVista.length) * 10) / 10
			: null;
		const publicados = new Set(visibles.map((x) => x.id));
		const preguntas = cr.preguntas
			.filter((q) => publicados.has(q.productoId) && respuestasQ.get(q.id))
			.map((q) => ({
				productoId: q.productoId,
				pregunta: q.pregunta,
				respuesta: respuestasQ.get(q.id)!
			}));
		return {
			slug: t.slug,
			nombre: t.nombre,
			plantilla: plantillaId,
			paleta: diseno.paleta,
			secciones: diseno.secciones.filter((s) => s.visible).map((s) => s.id),
			portada: {
				titulo: diseno.titulo || (p?.lema ?? t.nombre),
				lema: diseno.lema || (d.catalogo.ciudad ? `Desde ${d.catalogo.ciudad}` : ''),
				foto: diseno.portada || (p?.img ?? visibles.find((x) => x.foto)?.foto ?? null)
			},
			logo: diseno.logo ?? '',
			compartir: {
				titulo: diseno.compartirTitulo ?? '',
				descripcion: diseno.compartirDescripcion ?? '',
				foto: diseno.compartirFoto ?? ''
			},
			ciudad: d.catalogo.ciudad,
			direccion: conf.negocio.direccion,
			whatsapp: conf.negocio.whatsapp,
			categorias,
			productos: visibles.map((x) => ({
				id: x.id,
				nombre: x.nombre,
				descripcion: x.descripcion || delCatalogo.get(x.id)?.detalle || '',
				categoria: x.categoria || null,
				precio: x.precio,
				anterior: delCatalogo.get(x.id)?.anterior ?? null,
				foto: x.foto,
				etiqueta:
					delCatalogo.get(x.id)?.etiqueta ?? (x.stock > 0 && x.stock <= 3 ? 'Últimos' : null),
				// Lanzamiento, oferta con fin y señas (§52), como Go (vidriera.lanzamiento.aplicar).
				...aplicarLanzamiento(x.precio, lanz.get(x.id), mp.conectada, Date.now()),
				stock: x.stock,
				fotos: x.fotos,
				opciones: x.opciones,
				// Sólo las activas, sin SKU.
				variantes: x.variantes
					.filter((v) => v.activa)
					.map((v) => ({
						id: v.id,
						valores: v.valores,
						precio: v.precio,
						stock: v.stock,
						foto: v.foto
					}))
			})),
			masVendidos,
			resenas,
			promedio,
			preguntas,
			fotos: [...new Set(visibles.flatMap((x) => (x.foto ? [x.foto] : [])))].slice(0, 6),
			cobros: {
				transferencia: 15,
				cuotas: 6,
				mercadoPago: mp.conectada,
				// Lo que cargó el visitante en Configuración › Negocio, o uno de ejemplo rotulado así.
				datosTransferencia:
					conf.negocio.alias || conf.negocio.cbu
						? {
								alias: conf.negocio.alias ?? '',
								cbu: conf.negocio.cbu ?? '',
								titular: conf.negocio.titular ?? ''
							}
						: {
								alias: `${t.slug.slice(0, 12)}.ejemplo`,
								cbu: '',
								titular: 'Tienda de ejemplo',
								ejemplo: true
							}
			},
			// Lo de Envíos y retiro. Correo Argentino: la demo no tiene cuenta de MiCorreo, y sin
			// cotización real no se ofrece.
			envios: {
				gratisDesde: entregas.gratisDesde,
				retiro: entregas.retiro,
				retiroDireccion: entregas.retiro ? entregas.retiroDireccion : '',
				retiroHorario: entregas.retiro ? entregas.retiroHorario : '',
				moto: entregas.moto,
				motoCosto: entregas.moto ? entregas.motoCosto : 0,
				motoZona: entregas.moto ? entregas.motoZona : '',
				motoDemora: entregas.moto ? entregas.motoDemora : '',
				correo: false
			},
			turnos,
			// Como en Go: sin Mercado Pago conectado la tienda no ofrece planes.
			suscripciones: mp.conectada
				? planes.planes.filter((x) => x.activo).map((x) => planPublico(x, productos))
				: [],
			vitrina,
			// La vitrina de una plantilla nunca mide: no es de nadie.
			medicion: { ga4: medicion?.ga4 ?? null, pixel: medicion?.pixel ?? null }
		};
	}

	async guardarMiTienda(diseno: DisenoTienda) {
		await guardarCambio(this.db, this.sb.id, 'mi-tienda', 'borrador', diseno);
	}

	async publicarTienda() {
		const c = await cambiosDe(this.db, this.sb.id, 'mi-tienda');
		await guardarCambio(
			this.db,
			this.sb.id,
			'mi-tienda',
			'publicado',
			c.get('borrador') ?? DISENO_INICIAL
		);
	}

	// ---------------------------------------------------------------- Listas de precio

	/** Aplica lo programado si ya llegó el día (la demo no tiene worker: se mira al entrar). */
	private async programadoVigente() {
		const p = (await cambiosDe(this.db, this.sb.id, 'aumento')).get('programado') as
			AumentoProgramado | null | undefined;
		if (!p) return null;
		if (p.para > Date.now()) return p;
		await borrarCambio(this.db, this.sb.id, 'aumento', 'programado');
		await this.aplicarAumento(p.pct, p.redondeo);
		return null;
	}

	async listasPrecio(): Promise<ListasPrecio> {
		const programado = await this.programadoVigente();
		const d = await this.demo();
		const productos = await this.regProductos();
		const mayoristas = d.clientes.filter((c) => c.mayorista).length;
		return {
			listas: [
				{
					id: 'minorista',
					nombre: 'Minorista',
					quien: 'Tienda y local',
					ajuste: 0,
					detalle: 'Precio de referencia'
				},
				{
					id: 'mayorista',
					nombre: 'Mayorista',
					quien: 'Desde 10 unidades',
					ajuste: -25,
					detalle: `${mayoristas} ${mayoristas === 1 ? 'cliente mayorista' : 'clientes mayoristas'}`
				},
				{
					id: 'ml',
					nombre: 'Mercado Libre',
					quien: 'Cubre la comisión',
					ajuste: 10,
					detalle: 'Se actualiza solo'
				},
				{
					id: 'revendedoras',
					nombre: 'Revendedoras',
					quien: 'Para quienes revenden',
					ajuste: -15,
					detalle: 'Con link propio'
				}
			],
			productos: productos.length,
			muestra: [...productos]
				.sort((a, b) => b.vendidos30 - a.vendidos30)
				.slice(0, 4)
				.map((p) => ({ id: p.id, nombre: p.nombre, precio: p.precio })),
			programado: programado
				? {
						pct: programado.pct,
						redondeo: programado.redondeo,
						para: new Date(programado.para).toISOString()
					}
				: null
		};
	}

	/** Sube todos los precios y guarda los de antes para deshacer. */
	private async aplicarAumento(pct: number, redondeo: Redondeo) {
		const productos = await this.regProductos();
		const cambios = await this.cambiosProductos();
		const antes: Record<string, number | null> = {};
		for (const p of productos) {
			const c = cambios.get(p.id) ?? {};
			antes[p.id] = c.precio ?? null;
			await guardarCambio(this.db, this.sb.id, 'producto', p.id, {
				...c,
				precio: aumentar(p.precio, pct, redondeo)
			} satisfies CambioProducto);
		}
		await guardarCambio(this.db, this.sb.id, 'aumento', 'ultimo', { antes, en: Date.now() });
		return productos.length;
	}

	async aumentarPrecios(pct: number, redondeo: Redondeo, programar: boolean) {
		if (programar) {
			const para = proximoLunes(Date.now());
			await guardarCambio(this.db, this.sb.id, 'aumento', 'programado', {
				pct,
				redondeo,
				para
			} satisfies AumentoProgramado);
			return { productos: (await this.regProductos()).length, para: new Date(para).toISOString() };
		}
		return { productos: await this.aplicarAumento(pct, redondeo), para: null };
	}

	async deshacerAumento(): Promise<Resultado> {
		const u = (await cambiosDe(this.db, this.sb.id, 'aumento')).get('ultimo') as
			{ antes: Record<string, number | null> } | null | undefined;
		if (!u) return 'no-existe';
		const cambios = await this.cambiosProductos();
		for (const [id, precio] of Object.entries(u.antes)) {
			const resto: CambioProducto = { ...cambios.get(id) };
			if (precio === null) delete resto.precio;
			else resto.precio = precio;
			await guardarCambio(this.db, this.sb.id, 'producto', id, resto);
		}
		await borrarCambio(this.db, this.sb.id, 'aumento', 'ultimo');
		return 'ok';
	}

	async cancelarAumento(): Promise<Resultado> {
		if (!(await this.programadoVigente())) return 'no-existe';
		await borrarCambio(this.db, this.sb.id, 'aumento', 'programado');
		return 'ok';
	}

	// ---------------------------------------------------------------- Links de pago

	async crearLink(l: { monto: number; concepto: string; cuotas: boolean; pedido: number | null }) {
		const cod = codigo();
		await guardarCambio(this.db, this.sb.id, 'link', cod, {
			...l,
			creado: Date.now(),
			pagado: null
		} satisfies LinkDemo);
		return cod;
	}

	/** Pagar un link (en la demo se simula): queda pagado y, si cobra un pedido, el pedido pasa a preparar. */
	async pagarLink(cod: string): Promise<Resultado> {
		const c = (await cambiosDe(this.db, this.sb.id, 'link')).get(cod) as
			LinkDemo | null | undefined;
		if (!c) return 'no-existe';
		if (c.pagado) return 'no-se-puede';
		await guardarCambio(this.db, this.sb.id, 'link', cod, {
			...c,
			pagado: Date.now()
		} satisfies LinkDemo);
		if (c.pedido) {
			await this.cambiarEstado(c.pedido, 'Para preparar');
			// Si es un pedido cargado acá, queda cobrado y facturado.
			const n = (await cambiosDe(this.db, this.sb.id, 'pedido-nuevo')).get(String(c.pedido)) as
				PedidoNuevoDemo | undefined;
			if (n)
				await guardarCambio(this.db, this.sb.id, 'pedido-nuevo', String(c.pedido), {
					...n,
					pedido: { ...n.pedido, estadoPago: 'Aprobado', pago: 'mp', facturado: true }
				} satisfies PedidoNuevoDemo);
		}
		return 'ok';
	}

	// ---------------------------------------------------------------- Devoluciones y reseñas

	async devoluciones() {
		const d = await this.demo();
		const [c, nuevas] = await Promise.all([
			cambiosDe(this.db, this.sb.id, 'devolucion'),
			cambiosDe(this.db, this.sb.id, 'devolucion-nueva')
		]);
		const estados = new Map(
			[...c].filter(([, v]) => v).map(([k, v]) => [k, (v as { estado: EstadoDevolucion }).estado])
		);
		const base = devolucionesDemo(d, estados);
		const propias = [...nuevas.values()]
			.filter(Boolean)
			.map((v) => {
				const x = v as Devolucion;
				return { ...x, estado: estados.get(x.id) ?? x.estado };
			})
			.sort((a, b) => Date.parse(b.creada) - Date.parse(a.creada));
		// La tasa cuenta también las que pidió el visitante (botón de arrepentimiento).
		const entregados = d.pedidos.filter((x) => x.estado === 'Entregado').length;
		const todas = [...propias, ...base.devoluciones];
		return {
			devoluciones: todas,
			tasa: Math.round((todas.length / Math.max(1, entregados)) * 1000) / 10
		};
	}

	/**
	 * Botón de arrepentimiento de la tienda (demo): el pedido se identifica por el número que trae el
	 * código de la confirmación. Devuelve el código de trámite (el mismo si se repite).
	 */
	async arrepentirse(
		numero: number,
		/** `porNumero`: llegó con el número a secas (no el código firmado): el email tiene que coincidir. */
		p: { nombre: string; email: string; motivo: string; porNumero?: boolean },
		ahora = Date.now()
	): Promise<string | 'no-existe' | 'fuera-de-plazo'> {
		const [d, cambios] = await Promise.all([this.demo(), this.cambiosPedidos()]);
		const pedido = d.pedidos.find((x) => x.numero === numero);
		if (!pedido) return 'no-existe';
		if (p.porNumero) {
			const email = d.clientes.find((c) => c.id === pedido.clienteId)?.email;
			if (!email || email.toLowerCase() !== p.email.trim().toLowerCase()) return 'no-existe';
		}
		// Los 10 días corren desde que la recibió (si ya se entregó), si no desde la compra.
		const desde = cambios.get(numero)?.en?.Entregado ?? pedido.creado;
		if (ahora - desde > 10 * 86_400_000) return 'fuera-de-plazo';
		const nuevas = await cambiosDe(this.db, this.sb.id, 'devolucion-nueva');
		const ya = [...nuevas.values()].find((v) => (v as Devolucion | null)?.pedido === numero) as
			(Devolucion & { codigo: string }) | undefined;
		if (ya) return ya.codigo;
		// Si ya tiene una devolución de ejemplo, no se abre otra: el mismo trámite.
		const deEjemplo = (await this.devoluciones()).devoluciones.find((x) => x.pedido === numero);
		if (deEjemplo) return `ARR-${new Date(ahora).getFullYear()}-${deEjemplo.id.toUpperCase()}`;
		const codigo = `ARR-${new Date(ahora).getFullYear()}-${(nuevas.size + 1001).toString()}`;
		const dev: Devolucion & { codigo: string } = {
			id: `arr-${numero}`,
			codigo,
			producto:
				pedido.items[0].nombre +
				(pedido.items.length > 1 ? ` y ${pedido.items.length - 1} más` : ''),
			foto: null,
			cliente: p.nombre,
			pedido: numero,
			creada: new Date(ahora).toISOString(),
			motivo: p.motivo || 'Arrepentimiento (no hace falta motivo)',
			monto: pedido.total,
			resolucion: 'reintegro',
			estado: 'Pendiente'
		};
		await guardarCambio(this.db, this.sb.id, 'devolucion-nueva', dev.id, dev);
		return codigo;
	}

	async resolverDevolucion(
		id: string,
		accion: 'aprobar' | 'cambio' | 'deshacer'
	): Promise<Resultado> {
		const dev = (await this.devoluciones()).devoluciones.find((x) => x.id === id);
		if (!dev) return 'no-existe';
		if (accion === 'deshacer') {
			await borrarCambio(this.db, this.sb.id, 'devolucion', id);
			return 'ok';
		}
		await guardarCambio(this.db, this.sb.id, 'devolucion', id, {
			estado: accion === 'aprobar' ? 'Aprobada' : 'Cambio ofrecido'
		});
		return 'ok';
	}

	/** Respuestas publicadas a las preguntas (de Mercado Libre y de la tienda), por id. */
	private async respuestasPreguntas() {
		const c = await cambiosDe(this.db, this.sb.id, 'pregunta');
		return new Map(
			[...c].filter(([, v]) => v).map(([k, v]) => [k, (v as { respuesta: string }).respuesta])
		);
	}

	/**
	 * Reseñas y preguntas que dejó el visitante en su tienda online, más lo que hizo con ellas. Cada
	 * acción va en su propio cambio (ocultar y responder no se pisan si llegan juntas).
	 */
	private async cambiosResenas(): Promise<CambiosResenas> {
		const [nuevas, ocultas, respuestas, preguntas] = await Promise.all([
			cambiosDe(this.db, this.sb.id, 'resena-nueva'),
			cambiosDe(this.db, this.sb.id, 'resena-oculta'),
			cambiosDe(this.db, this.sb.id, 'resena-respuesta'),
			cambiosDe(this.db, this.sb.id, 'pregunta-tienda')
		]);
		return {
			nuevas: [...nuevas.values()].filter(Boolean) as ResenaNuevaDemo[],
			ocultas: new Set([...ocultas].filter(([, v]) => v).map(([k]) => k)),
			respuestas: new Map(
				[...respuestas]
					.filter(([, v]) => v)
					.map(([k, v]) => [k, (v as { respuesta: string }).respuesta])
			),
			preguntas: [...preguntas.values()].filter(Boolean) as PreguntaTiendaDemo[]
		};
	}

	async resenas() {
		const [d, respuestas, cambios, productos] = await Promise.all([
			this.demo(),
			this.respuestasPreguntas(),
			this.cambiosResenas(),
			this.regProductos()
		]);
		// Con los productos que sumó el visitante (el catálogo de ejemplo no los tiene).
		cambios.productos = new Map(productos.map((p) => [p.id, { nombre: p.nombre, foto: p.foto }]));
		return resenasDemo(d, Date.now(), respuestas, cambios);
	}

	async responderPregunta(id: string, respuesta: string): Promise<Resultado> {
		if (!(await this.resenas()).preguntas.some((p) => p.id === id)) return 'no-existe';
		await guardarCambio(this.db, this.sb.id, 'pregunta', id, {
			respuesta: respuesta.trim().slice(0, id.startsWith('t-') ? 1000 : 2000)
		});
		return 'ok';
	}

	/** Ocultar o mostrar una reseña (nunca se borra), o responderla en público (vacía, la saca). */
	async accionResena(
		id: string,
		accion: 'ocultar' | 'mostrar' | 'responder',
		respuesta = ''
	): Promise<Resultado> {
		const d = await this.demo();
		if (!todasLasResenas(d, await this.cambiosResenas()).some((r) => r.id === id))
			return 'no-existe';
		if (accion === 'ocultar') await guardarCambio(this.db, this.sb.id, 'resena-oculta', id, true);
		else if (accion === 'mostrar') await borrarCambio(this.db, this.sb.id, 'resena-oculta', id);
		else {
			const r = respuesta.trim();
			if (r.length > 1000) return 'no-se-puede';
			if (r) await guardarCambio(this.db, this.sb.id, 'resena-respuesta', id, { respuesta: r });
			else await borrarCambio(this.db, this.sb.id, 'resena-respuesta', id);
		}
		return 'ok';
	}

	/**
	 * La reseña que deja alguien en la tienda online. Con el código de un pedido cobrado (el firmado
	 * de la confirmación) queda con compra, una por pedido y producto; con un código que no es de un
	 * pedido cobrado, se guarda igual pero sin compra (como Go).
	 */
	async dejarResena(n: {
		codigo?: string;
		productoId?: string;
		autor: string;
		estrellas: number;
		texto?: string;
	}): Promise<'ok' | 'ya-reseno' | 'invalido'> {
		const autor = n.autor.trim();
		const texto = (n.texto ?? '').trim();
		if (
			autor.length < 2 ||
			autor.length > 60 ||
			!Number.isInteger(n.estrellas) ||
			n.estrellas < 1 ||
			n.estrellas > 5 ||
			texto.length > 1000
		)
			return 'invalido';
		let pedido: number | null = null;
		const numero = Number(n.codigo?.split('-')[0]);
		if (
			n.codigo &&
			Number.isInteger(numero) &&
			(await codigoPedido(this.sb.id, numero)) === n.codigo
		) {
			const p = await this.pedido(numero);
			if (p && COBRADO.includes(p.estadoPago)) pedido = numero;
		}
		const productoId =
			n.productoId && (await this.regProductos()).some((p) => p.id === n.productoId)
				? n.productoId
				: null;
		const ahora = Date.now();
		// Con compra, la clave es el pedido y el producto: así hay una sola por compra.
		const id =
			pedido !== null
				? `p${pedido}-${productoId ?? 'tienda'}`
				: `n${ahora.toString(36)}${randomToken(4).replace(/[^a-z0-9]/gi, '')}`;
		if (pedido !== null && (await cambiosDe(this.db, this.sb.id, 'resena-nueva')).get(id))
			return 'ya-reseno';
		await guardarCambio(this.db, this.sb.id, 'resena-nueva', id, {
			id,
			autor,
			estrellas: n.estrellas,
			texto,
			productoId,
			pedido,
			creada: ahora
		} satisfies ResenaNuevaDemo);
		return 'ok';
	}

	/** Una pregunta sobre un producto publicado; aparece en el panel junto a las de Mercado Libre. */
	async preguntar(q: {
		productoId: string;
		nombre: string;
		email?: string;
		pregunta: string;
	}): Promise<'ok' | 'invalido'> {
		const nombre = q.nombre.trim();
		const pregunta = q.pregunta.trim();
		const email = (q.email ?? '').trim();
		if (
			nombre.length < 2 ||
			nombre.length > 60 ||
			pregunta.length < 3 ||
			pregunta.length > 500 ||
			email.length > 120
		)
			return 'invalido';
		const p = (await this.regProductos()).find((x) => x.id === q.productoId);
		if (!p?.publicado) return 'invalido';
		const ahora = Date.now();
		const id = `t-${ahora.toString(36)}${randomToken(4).replace(/[^a-z0-9]/gi, '')}`;
		await guardarCambio(this.db, this.sb.id, 'pregunta-tienda', id, {
			id,
			productoId: p.id,
			nombre,
			email: email || null,
			pregunta,
			creada: ahora
		} satisfies PreguntaTiendaDemo);
		return 'ok';
	}

	/** La plantilla del sandbox (o la del rubro). */
	private plantilla() {
		return this.sb.plantilla ?? buscarRubro(this.sb.rubro)!.plantilla;
	}

	async resumen(): Promise<ResumenPanel> {
		const d = await this.demo();
		const leidas = (await cambiosDe(this.db, this.sb.id, 'notificaciones')).get('leidas') as
			{ hasta?: number } | undefined;
		const r = resumenDemo(d, this.sb.creadoEn, leidas?.hasta ?? 0, Date.now());
		const plan = (await this.estadoCuenta()).plan.actual ?? 'pro';
		r.prueba.plan = NOMBRE_PLAN[plan];
		// La insignia cuenta lo mismo que la pantalla: las tareas de hoy sin marcar.
		const tareas = (await this.paraHacer()).grupos.flatMap((g) => g.tareas);
		r.insignias.paraHacer = tareas.filter((t) => !t.hecha).length;
		// Los canales de la barra son los que están conectados en Integraciones.
		const { conectadas } = await this.integraciones();
		r.canales = r.canales.filter((c) => conectadas[c]);
		// Cada turno que entró desde la tienda avisa en el panel.
		const turnos = (await this.turnosNuevosDeLaTienda())
			.sort((a, b) => (a.creadoEn < b.creadoEn ? 1 : -1))
			.slice(0, 5)
			.map((t) => ({
				id: `turno-${t.id}`,
				tipo: 'ok',
				titulo: `Turno nuevo: ${t.servicio} con ${t.profesional}`,
				detalle: `${t.cliente.nombre} · ${cuandoTurno(t.inicio)}`,
				creada: t.creadoEn,
				leida: Date.parse(t.creadoEn) <= (leidas?.hasta ?? 0),
				grupo: 'ventas' as const,
				ruta: '/panel/turnos'
			}));
		r.notificaciones = [...turnos, ...r.notificaciones];
		return r;
	}

	/** La campanita de la demo: las del resumen, con las que el visitante tocó ya leídas. */
	async notificaciones(): Promise<PaginaNotificaciones> {
		const [r, cambios] = await Promise.all([
			this.resumen(),
			cambiosDe(this.db, this.sb.id, 'notificaciones')
		]);
		const tocadas = new Set((cambios.get('tocadas') as { ids?: string[] } | undefined)?.ids ?? []);
		const lista = r.notificaciones.map((n) => (tocadas.has(n.id) ? { ...n, leida: true } : n));
		return {
			notificaciones: lista,
			sinLeer: lista.filter((n) => !n.leida).length,
			siguiente: null
		};
	}

	async marcarNotificacionLeida(id: string) {
		const cambios = await cambiosDe(this.db, this.sb.id, 'notificaciones');
		const ids = (cambios.get('tocadas') as { ids?: string[] } | undefined)?.ids ?? [];
		if (ids.includes(id)) return;
		await guardarCambio(this.db, this.sb.id, 'notificaciones', 'tocadas', {
			ids: [...ids, id.slice(0, 80)].slice(-100)
		});
	}

	async marcarNotificacionesLeidas() {
		await guardarCambio(this.db, this.sb.id, 'notificaciones', 'leidas', { hasta: Date.now() });
	}

	async tablero(periodo: Periodo): Promise<Tablero> {
		const t = await this.tableroConMontos(periodo);
		return PERMISOS_DE[this.rol].includes('finanzas') ? t : sinMontos(t);
	}

	/** El tablero entero, con montos (Finanzas lo usa; sólo lo ve quien tiene finanzas). */
	private async tableroConMontos(periodo: Periodo): Promise<TableroConMontos> {
		const d = await this.demo();
		const c = await cambiosDe(this.db, this.sb.id, 'tablero');
		const cambios: CambiosTablero = {
			meta: (c.get('meta') as { objetivo: number } | undefined)?.objetivo,
			certificadoOculto: c.has('certificado'),
			ideasHechas: await this.ideasHechas()
		};
		// Quien prueba la demo es la dueña de la tienda de ejemplo (la del onboarding).
		return tableroDemo(d, periodo, {
			base: this.sb.creadoEn,
			ahora: Date.now(),
			hola: 'Valeria',
			cambios
		});
	}

	async guardarMeta(objetivo: number) {
		await guardarCambio(this.db, this.sb.id, 'tablero', 'meta', { objetivo });
	}

	async ocultarAviso(aviso: 'certificado') {
		await guardarCambio(this.db, this.sb.id, 'tablero', aviso, true);
	}

	private async ideasHechas() {
		const c = await cambiosDe(this.db, this.sb.id, 'hornero');
		return IDEAS.filter((k) => c.get(k));
	}

	async recordarCarritos() {
		const d = await this.demo();
		if ((await this.ideasHechas()).includes('carritos')) return 0;
		await guardarCambio(this.db, this.sb.id, 'hornero', 'carritos', { en: Date.now() });
		return d.carritos.length;
	}

	async marcarIdea(idea: Idea, hecha: boolean) {
		if (hecha) await guardarCambio(this.db, this.sb.id, 'hornero', idea, { en: Date.now() });
		else await borrarCambio(this.db, this.sb.id, 'hornero', idea);
	}

	async contextoHornero(): Promise<ContextoHornero> {
		const c = contextoDemo(await this.demo(), Date.now());
		return PERMISOS_DE[this.rol].includes('finanzas') ? c : contextoSinMontos(c);
	}

	// ---------------------------------------------------------------- Suscripciones (§51)

	/** Planes y suscriptores: los de ejemplo con lo que cambió el visitante. */
	private async estadoSuscripciones() {
		const c = await cambiosDe(this.db, this.sb.id, ENTIDAD_SUSCRIPCIONES);
		return mezclarSuscripciones(c, this.plantilla(), this.sb.creadoEn);
	}

	private async guardarSuscriptor(x: Suscriptor) {
		await guardarCambio(this.db, this.sb.id, ENTIDAD_SUSCRIPCIONES, `s:${x.id}`, x);
	}

	async suscripcionesPanel(): Promise<PanelSuscripciones> {
		const [e, mp] = await Promise.all([this.estadoSuscripciones(), this.mercadoPago()]);
		return { mercadoPago: mp.conectada, planes: e.planes, suscriptores: e.suscriptores };
	}

	async guardarPlanSuscripcion(
		id: string | null,
		p: PlanSuscripcionEditable
	): Promise<PlanSuscripcion | null | 'invalido'> {
		const [e, productos] = await Promise.all([this.estadoSuscripciones(), this.regProductos()]);
		const previo = id ? e.planes.find((x) => x.id === id) : null;
		if (id && !previo) return null;
		// Cada producto es de la tienda y, con opciones, trae una variante suya (como en Go).
		for (const x of p.productos) {
			const r = productos.find((y) => y.id === x.productoId);
			if (
				!r ||
				r.opciones.length > 0 !== !!x.varianteId ||
				(x.varianteId && !r.variantes.some((v) => v.id === x.varianteId))
			)
				return 'invalido';
		}
		const { suscriptores = 0, ...base } = previo ?? {};
		const guardado: Omit<PlanSuscripcion, 'suscriptores'> = {
			...base,
			...p,
			id: previo?.id ?? crypto.randomUUID(),
			orden: previo?.orden ?? Math.max(-1, ...e.planes.map((x) => x.orden)) + 1
		};
		await guardarCambio(this.db, this.sb.id, ENTIDAD_SUSCRIPCIONES, `p:${guardado.id}`, guardado);
		return { ...guardado, suscriptores };
	}

	async borrarPlanSuscripcion(id: string): Promise<Resultado> {
		const e = await this.estadoSuscripciones();
		if (!e.planes.some((x) => x.id === id)) return 'no-existe';
		if (e.suscriptores.some((x) => x.planId === id)) return 'no-se-puede';
		await guardarCambio(this.db, this.sb.id, ENTIDAD_SUSCRIPCIONES, `p:${id}`, null);
		return 'ok';
	}

	async cancelarSuscriptor(id: string): Promise<Suscriptor | 'no-existe' | 'mercadopago'> {
		const x = (await this.estadoSuscripciones()).suscriptores.find((y) => y.id === id);
		if (!x) return 'no-existe';
		if (x.estado === 'cancelada') return x;
		const nuevo: Suscriptor = {
			...x,
			estado: 'cancelada',
			canceladaPor: 'tienda',
			proximoCobro: null
		};
		await this.guardarSuscriptor(nuevo);
		return nuevo;
	}

	/** Los planes activos de la tienda (la página del plan los busca acá). */
	async planesTienda(): Promise<PlanPublico[]> {
		const [e, productos] = await Promise.all([this.estadoSuscripciones(), this.regProductos()]);
		return e.planes.filter((x) => x.activo).map((x) => planPublico(x, productos));
	}

	/**
	 * El alta desde la tienda de la demo: queda activa sin cobrar (en la demo no hay Mercado Pago de
	 * verdad). El precio sale del plan, como en Go.
	 */
	async suscribirPublico(
		a: {
			planId: string;
			nombre: string;
			email: string;
			telefono: string;
			entrega: 'envio' | 'retiro';
			direccion: string;
			cp: string;
		},
		ahora = Date.now()
	): Promise<{ codigo: string } | 'no-existe' | 'cupo' | 'invalido'> {
		const e = await this.estadoSuscripciones();
		const p = e.planes.find((x) => x.id === a.planId && x.activo);
		if (!p) return 'no-existe';
		if ((a.entrega === 'envio' && !p.envio) || (a.entrega === 'retiro' && !p.retiro))
			return 'invalido';
		if (p.cupo !== null && p.suscriptores >= p.cupo) return 'cupo';
		const codigo = codigoSuscripcion();
		const x: Suscriptor = {
			id: crypto.randomUUID(),
			planId: p.id,
			plan: p.nombre,
			precio: p.precio,
			frecuencia: p.frecuencia,
			nombre: a.nombre,
			email: a.email,
			telefono: a.telefono,
			entrega: a.entrega,
			direccion: a.entrega === 'envio' ? a.direccion : '',
			estado: 'activa',
			proximoCobro: new Date(ahora + diasDe(p.frecuencia) * 86_400_000).toISOString(),
			cobros: 0,
			ultimoCobro: null,
			creadaEn: new Date(ahora).toISOString(),
			canceladaPor: null
		};
		await Promise.all([
			this.guardarSuscriptor(x),
			guardarCambio(this.db, this.sb.id, ENTIDAD_SUSCRIPCIONES, `c:${codigo}`, { id: x.id })
		]);
		return { codigo };
	}

	private async suscriptorDeCodigo(codigo: string) {
		const ref = (await cambiosDe(this.db, this.sb.id, ENTIDAD_SUSCRIPCIONES)).get(`c:${codigo}`) as
			{ id: string } | undefined;
		if (!ref) return null;
		const e = await this.estadoSuscripciones();
		const x = e.suscriptores.find((y) => y.id === ref.id);
		return x ? { x, plan: e.planes.find((p) => p.id === x.planId) } : null;
	}

	async suscripcionPublica(codigo: string): Promise<SuscripcionPublica | null> {
		const [r, conf, t] = await Promise.all([
			this.suscriptorDeCodigo(codigo),
			this.configuracion(),
			this.tienda()
		]);
		return r
			? suscripcionPublicaDe(r.x, codigo, r.plan, {
					nombre: t.nombre,
					whatsapp: conf.negocio.whatsapp ?? ''
				})
			: null;
	}

	/** Pausar, reanudar o cancelar desde el link privado. */
	async cambiarSuscripcionPublica(
		codigo: string,
		accion: 'pausar' | 'reanudar' | 'cancelar',
		ahora = Date.now()
	): Promise<SuscripcionPublica | 'no-existe' | 'estado'> {
		const r = await this.suscriptorDeCodigo(codigo);
		if (!r) return 'no-existe';
		const { x } = r;
		if (!(x.estado === 'cancelada' && accion === 'cancelar')) {
			if (
				x.estado === 'cancelada' ||
				(accion === 'pausar' && x.estado !== 'activa') ||
				(accion === 'reanudar' && x.estado !== 'pausada')
			)
				return 'estado';
			const estado =
				accion === 'pausar' ? 'pausada' : accion === 'reanudar' ? 'activa' : 'cancelada';
			await this.guardarSuscriptor({
				...x,
				estado,
				canceladaPor: estado === 'cancelada' ? 'comprador' : x.canceladaPor,
				proximoCobro:
					estado === 'activa'
						? new Date(ahora + diasDe(x.frecuencia) * 86_400_000).toISOString()
						: null
			});
		}
		return (await this.suscripcionPublica(codigo)) ?? 'no-existe';
	}

	// ---------------------------------------------------------------- Turnos

	/** Ajustes, servicios, profesionales y turnos: los de ejemplo con lo que cambió el visitante. */
	private async estadoTurnos(ahora = Date.now()) {
		const c = await cambiosDe(this.db, this.sb.id, ENTIDAD_TURNOS);
		const ej = turnosDemo(this.plantilla(), this.sb.creadoEn);
		function mezclar<T extends { id: string }>(base: T[], pre: string) {
			const m = new Map(base.map((x) => [x.id, x]));
			for (const [k, v] of c)
				if (k.startsWith(pre)) {
					if (v) m.set(k.slice(pre.length), v as T);
					else m.delete(k.slice(pre.length));
				}
			return [...m.values()];
		}
		const porOrden = (a: { orden: number }, b: { orden: number }) => a.orden - b.orden;
		return {
			ajustes: (c.get('ajustes') as AjustesTurnos | undefined) ?? ej.ajustes,
			servicios: mezclar(ej.servicios, 's:').sort(porOrden),
			profesionales: mezclar(ej.profesionales, 'p:').sort(porOrden),
			turnos: mezclar(ej.turnos, 't:')
				.map((t) => ({ ...t, vencido: !ocupa(t, ahora) && t.estado === 'reservado' }))
				.sort((a, b) => (a.inicio < b.inicio ? -1 : a.inicio > b.inicio ? 1 : 0)),
			codigos: c
		};
	}

	private async guardarTurno(t: Turno) {
		await guardarCambio(this.db, this.sb.id, ENTIDAD_TURNOS, `t:${t.id}`, t);
	}

	async ajustesTurnos(): Promise<AjustesTurnos> {
		return (await this.estadoTurnos()).ajustes;
	}

	async guardarAjustesTurnos(a: AjustesTurnos) {
		await guardarCambio(this.db, this.sb.id, ENTIDAD_TURNOS, 'ajustes', a);
		return a;
	}

	async serviciosTurnos() {
		return (await this.estadoTurnos()).servicios;
	}

	async guardarServicio(id: string | null, s: ServicioEditable): Promise<Servicio | null> {
		const { servicios } = await this.estadoTurnos();
		const antes = id ? servicios.find((x) => x.id === id) : undefined;
		if (id && !antes) return null;
		const nuevo: Servicio = {
			...s,
			id: id ?? randomToken(9),
			orden: s.orden ?? antes?.orden ?? Math.max(-1, ...servicios.map((x) => x.orden)) + 1
		};
		await guardarCambio(this.db, this.sb.id, ENTIDAD_TURNOS, `s:${nuevo.id}`, nuevo);
		return nuevo;
	}

	async borrarServicio(id: string): Promise<Resultado> {
		if (!(await this.estadoTurnos()).servicios.some((x) => x.id === id)) return 'no-existe';
		await guardarCambio(this.db, this.sb.id, ENTIDAD_TURNOS, `s:${id}`, null);
		return 'ok';
	}

	async profesionalesTurnos() {
		return (await this.estadoTurnos()).profesionales;
	}

	async guardarProfesional(id: string | null, p: ProfesionalEditable): Promise<Profesional | null> {
		const { profesionales } = await this.estadoTurnos();
		const antes = id ? profesionales.find((x) => x.id === id) : undefined;
		if (id && !antes) return null;
		const nuevo: Profesional = {
			...p,
			id: id ?? randomToken(9),
			orden: p.orden ?? antes?.orden ?? Math.max(-1, ...profesionales.map((x) => x.orden)) + 1
		};
		await guardarCambio(this.db, this.sb.id, ENTIDAD_TURNOS, `p:${nuevo.id}`, nuevo);
		return nuevo;
	}

	async borrarProfesional(id: string): Promise<Resultado> {
		if (!(await this.estadoTurnos()).profesionales.some((x) => x.id === id)) return 'no-existe';
		await guardarCambio(this.db, this.sb.id, ENTIDAD_TURNOS, `p:${id}`, null);
		return 'ok';
	}

	async agendaTurnos(desde: string, hasta: string, profesional?: string) {
		return (await this.estadoTurnos()).turnos.filter((t) => {
			const dia = diaAr(Date.parse(t.inicio));
			return dia >= desde && dia <= hasta && (!profesional || t.profesionalId === profesional);
		});
	}

	async crearTurno(n: TurnoNuevo, ahora = Date.now()): Promise<Turno | 'ocupado' | 'invalido'> {
		const e = await this.estadoTurnos(ahora);
		const s = e.servicios.find((x) => x.id === n.servicioId);
		const p = e.profesionales.find((x) => x.id === n.profesionalId);
		const inicio = Date.parse(n.inicio);
		if (!s || !p || !Number.isFinite(inicio)) return 'invalido';
		const fin = inicio + s.duracion * 60_000;
		if (!n.forzar && superpone(e.turnos, p.id, inicio, fin, ahora)) return 'ocupado';
		const t: Turno = {
			id: randomToken(9),
			codigo: randomToken(18),
			servicioId: s.id,
			servicio: s.nombre,
			profesionalId: p.id,
			profesional: p.nombre,
			inicio: isoAr(inicio),
			fin: isoAr(fin),
			cliente: { nombre: n.nombre, telefono: n.telefono, email: n.email },
			precio: s.precio,
			sena: 0,
			estado: 'confirmado',
			senaPagada: false,
			senaDevuelta: false,
			origen: 'panel',
			nota: n.nota ?? null,
			creadoEn: new Date(ahora).toISOString(),
			venceEn: null,
			vencido: false
		};
		await this.guardarTurnoNuevo(t);
		return t;
	}

	/** El turno y su código (para encontrarlo sin la cookie, desde el link de la confirmación). */
	private async guardarTurnoNuevo(t: Turno) {
		await Promise.all([
			this.guardarTurno(t),
			guardarCambio(this.db, this.sb.id, ENTIDAD_TURNOS, `c:${t.codigo}`, { id: t.id })
		]);
	}

	async estadoTurno(id: string, estado: EstadoTurno): Promise<Turno | 'no-existe'> {
		const t = (await this.estadoTurnos()).turnos.find((x) => x.id === id);
		if (!t) return 'no-existe';
		const nuevo: Turno = { ...t, estado, venceEn: estado === 'reservado' ? t.venceEn : null };
		await this.guardarTurno(nuevo);
		return nuevo;
	}

	async devolverSena(id: string): Promise<Turno | 'no-existe' | 'sin-sena' | 'sin-cobros'> {
		const [e, mp] = await Promise.all([this.estadoTurnos(), this.mercadoPago()]);
		const t = e.turnos.find((x) => x.id === id);
		if (!t) return 'no-existe';
		if (!t.senaPagada || t.senaDevuelta || !t.sena) return 'sin-sena';
		if (!mp.conectada) return 'sin-cobros';
		const nuevo = { ...t, senaDevuelta: true };
		await this.guardarTurno(nuevo);
		return nuevo;
	}

	/** Lo que la tienda pública muestra de los turnos (null si no toma). */
	async turnosTienda(): Promise<TurnosTienda | null> {
		const e = await this.estadoTurnos();
		if (!e.ajustes.activo) return null;
		return {
			servicios: e.servicios.filter((s) => s.activo),
			profesionales: e.profesionales.filter((p) => p.activo),
			// En la demo no hay Mercado Pago de verdad: el turno se confirma sin seña.
			sena: 0,
			horario: e.ajustes.horario,
			anticipacionMin: e.ajustes.anticipacionMin,
			diasVista: e.ajustes.diasVista
		};
	}

	/** La grilla de un día para la tienda: cada horario con si está libre (los otros, tachados). */
	async horariosTurnos(
		servicio: string,
		profesional: string | null,
		dia: string,
		ahora = Date.now()
	) {
		const e = await this.estadoTurnos(ahora);
		const s = e.servicios.find((x) => x.id === servicio);
		if (!e.ajustes.activo || !s) return null;
		return horariosDelDia({ ...e, servicio: s, profesionalId: profesional, dia, ahora });
	}

	/** Los libres de hoy (portada) y el próximo libre de cada profesional (El equipo). */
	async proximosTurnos(ahora = Date.now()): Promise<ProximosTurnos> {
		const e = await this.estadoTurnos(ahora);
		const activos = e.servicios.filter((s) => s.activo);
		const hoy = diaAr(ahora);
		const base = { ...e, ahora };
		const porProfesional: Record<string, string | null> = {};
		for (const p of e.profesionales.filter((x) => x.activo)) {
			const s = activos.find((x) => hace(p, x.id));
			porProfesional[p.id] = null;
			for (let d = 0; s && d < e.ajustes.diasVista && !porProfesional[p.id]; d++)
				porProfesional[p.id] =
					libres({ ...base, servicio: s, profesionalId: p.id, dia: sumarDias(hoy, d) })[0]
						?.inicio ?? null;
		}
		return {
			hoy: activos[0]
				? libres({ ...base, servicio: activos[0], profesionalId: null, dia: hoy }).map(
						(h) => h.inicio
					)
				: [],
			porProfesional
		};
	}

	/**
	 * Reserva desde la tienda: precio, duración y seña salen del servicio, y el horario se re-chequea
	 * acá. En la demo no hay Mercado Pago real: queda confirmado sin seña.
	 */
	async reservarTurno(
		r: {
			servicioId: string;
			profesionalId: string | null;
			inicio: string;
			nombre: string;
			telefono: string;
			email: string;
		},
		ahora = Date.now()
	): Promise<Turno | 'ocupado' | 'invalido' | 'apagado'> {
		const e = await this.estadoTurnos(ahora);
		if (!e.ajustes.activo) return 'apagado';
		const s = e.servicios.find((x) => x.id === r.servicioId && x.activo);
		if (!s) return 'invalido';
		const quien = asignar({ ...e, servicio: s, profesionalId: r.profesionalId, ahora }, r.inicio);
		const p = quien && e.profesionales.find((x) => x.id === quien);
		if (!p) return 'ocupado';
		const inicio = Date.parse(r.inicio);
		const t: Turno = {
			id: randomToken(9),
			codigo: randomToken(18),
			servicioId: s.id,
			servicio: s.nombre,
			profesionalId: p.id,
			profesional: p.nombre,
			inicio: isoAr(inicio),
			fin: isoAr(inicio + s.duracion * 60_000),
			cliente: { nombre: r.nombre, telefono: r.telefono, email: r.email },
			precio: s.precio,
			sena: 0,
			estado: 'confirmado',
			senaPagada: false,
			senaDevuelta: false,
			origen: 'tienda',
			nota: null,
			creadoEn: new Date(ahora).toISOString(),
			venceEn: null,
			vencido: false
		};
		await this.guardarTurnoNuevo(t);
		return t;
	}

	/** El turno por su código público, como lo ve quien reservó. */
	async turnoPublico(codigo: string, ahora = Date.now()): Promise<TurnoPublico | null> {
		const [e, conf, t] = await Promise.all([
			this.estadoTurnos(ahora),
			this.configuracion(),
			this.tienda()
		]);
		const ref = e.codigos.get(`c:${codigo}`) as { id: string } | undefined;
		const x = ref && e.turnos.find((y) => y.id === ref.id);
		if (!x) return null;
		const cancelable =
			((x.estado === 'reservado' && !x.vencido) || x.estado === 'confirmado') &&
			ahora <= Date.parse(x.inicio) - e.ajustes.anticipacionMin * 60_000;
		return {
			codigo: x.codigo,
			servicio: x.servicio,
			profesional: x.profesional,
			inicio: x.inicio,
			fin: x.fin,
			nombre: x.cliente.nombre.split(' ')[0],
			precio: x.precio,
			sena: x.sena,
			senaPagada: x.senaPagada,
			estado: x.estado,
			venceEn: x.venceEn,
			vencido: x.vencido,
			cancelable,
			tienda: t.nombre,
			direccion: conf.negocio.direccion,
			whatsapp: conf.negocio.whatsapp
		};
	}

	/** Cancela desde la confirmación (hasta la anticipación mínima). Uno cancelado no es error. */
	async cancelarTurnoPublico(codigo: string, ahora = Date.now()) {
		const p = await this.turnoPublico(codigo, ahora);
		if (!p) return 'no-existe' as const;
		if (p.estado === 'cancelado') return p;
		if (!p.cancelable) return 'tarde' as const;
		const e = await this.estadoTurnos(ahora);
		const id = (e.codigos.get(`c:${codigo}`) as { id: string }).id;
		const t = e.turnos.find((y) => y.id === id)!;
		await this.guardarTurno({ ...t, estado: 'cancelado', venceEn: null });
		return { ...p, estado: 'cancelado' as const, cancelable: false };
	}

	/** Turnos que entraron desde la tienda (lo que reservó el visitante): avisos del panel. */
	private async turnosNuevosDeLaTienda() {
		const { turnos } = await this.estadoTurnos();
		return turnos.filter((t) => t.origen === 'tienda' && !t.id.startsWith('ej-'));
	}
}
