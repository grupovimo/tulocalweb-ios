/**
 * Datos desde la API Go (DATA_SOURCE=api). El navegador nunca habla con Go: lo hace este servidor
 * (BFF), con el token de la sesión. Los errores problem+json se convierten en errores de SvelteKit.
 */
import { error } from '@sveltejs/kit';
import {
	conectadasDeMedicion,
	estadoInicial,
	type AjusteIntegracion
} from '#lib/panel/integraciones.ts';
import { consultaDe } from '#lib/panel/actividad.ts';
import type { Beneficio } from '#lib/panel/planes.ts';
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
	AccionCupon,
	ArcaAjustes,
	ArcaAlta,
	ArcaCertificado,
	ArcaConfig,
	ArcaPedido,
	ArcaPrueba,
	ConexionML,
	Medicion,
	Entregas,
	EstadoEntregas,
	VentaMayorista,
	ConexionMP,
	PublicacionML,
	Dominios,
	MisDatos,
	PreferenciasAvisos,
	EstadoPushWeb,
	SuscripcionWeb,
	TipoExport,
	Acceso,
	PlanCuenta,
	Configuracion,
	Miembro,
	Negocio,
	IdPlan,
	DisenoTienda,
	MiTienda,
	Finanzas,
	FiltroActividad,
	PaginaActividad,
	ParaHacer,
	Campana,
	Canales,
	Informes,
	Marketing,
	Newsletter,
	EnvioNewsletter,
	NovedadNueva,
	AccionLote,
	AccionRelampago,
	Cupon,
	Cupones,
	CuponNuevo,
	ListasPrecio,
	Devolucion,
	PedidoNuevo,
	ResenasYPreguntas,
	AutomaticosCarritos,
	Tarjeta,
	Carritos,
	Conversacion,
	ConversacionBreve,
	MovimientoStock,
	FiltrosLista,
	ListaGenerica,
	NombreLista,
	Producto,
	ProductoEditable,
	ContextoHornero,
	EstadoPedido,
	FiltrosPedidos,
	ListaPedidos,
	PedidoDetalle,
	Resultado,
	Datos,
	Idea,
	Onboarding,
	Periodo,
	PaginaNotificaciones,
	ResumenPanel,
	Rubro,
	Tablero,
	Tienda
} from './tipos.ts';
import type { Importacion, ImportacionNueva, PedidoHistorial } from '../importacion/fuente.ts';
import type { ErrorVueltaTN } from './tipos.ts';
import type { ConfigLanzamiento, LanzamientoProducto } from '#lib/tienda/lanzamiento.ts';
import type { RolInvitable } from '#lib/panel/permisos.ts';

/** Un producto de una API sin galería ni variantes todavía se ve como uno sin opciones. */
function conVariantes(p: Producto): Producto {
	return {
		...p,
		fotos: p.fotos ?? (p.foto ? [p.foto] : []),
		opciones: p.opciones ?? [],
		variantes: p.variantes ?? []
	};
}

export class DatosApi implements Datos {
	constructor(
		private base: string,
		private token: string | undefined,
		// Envuelto: en Workers, `fetch` llamado como método de otro objeto da "Illegal invocation".
		private f: typeof fetch = (i, init) => fetch(i, init)
	) {}

	private async pedir<T>(
		ruta: string,
		metodo: 'GET' | 'PUT' | 'POST' | 'PATCH' | 'DELETE' = 'GET',
		cuerpo?: unknown
	): Promise<T> {
		const res = await this.f(this.base + ruta, {
			method: metodo,
			headers: {
				accept: 'application/json',
				...(cuerpo === undefined ? {} : { 'content-type': 'application/json' }),
				...(this.token ? { authorization: `Bearer ${this.token}` } : {})
			},
			body: cuerpo === undefined ? undefined : JSON.stringify(cuerpo)
		});
		if (!res.ok) {
			const p = (await res.json().catch(() => null)) as { detail?: string; code?: string } | null;
			error(res.status, { message: p?.detail ?? 'La API no respondió bien.', code: p?.code });
		}
		if (res.status === 204) return undefined as T;
		return (await res.json()) as T;
	}

	rubros() {
		return this.pedir<Rubro[]>('/v1/rubros');
	}

	tienda() {
		return this.pedir<Tienda>('/v1/panel/tienda');
	}

	onboarding() {
		return this.pedir<Onboarding>('/v1/panel/onboarding');
	}

	guardarOnboarding(o: Onboarding) {
		return this.pedir<void>('/v1/panel/onboarding', 'PUT', o);
	}

	disponibilidadDireccion(slug: string) {
		return this.pedir<DisponibilidadDireccion>(
			`/v1/panel/tienda/direccion?slug=${encodeURIComponent(slug)}`
		);
	}

	async cambiarDireccion(slug: string): Promise<CambioDireccion> {
		const r = await this.intentar('/v1/panel/tienda/direccion', 'PUT', { slug }, [400]);
		if (r.ok) return { ok: true, ...((await r.json()) as { slug: string; anterior: string }) };
		const p = (await r.json().catch(() => null)) as { detail?: string; code?: string } | null;
		const motivo =
			p?.code === 'usado' || p?.code === 'reservado' || p?.code === 'tope' ? p.code : 'invalido';
		return { ok: false, motivo, mensaje: p?.detail ?? 'No pudimos cambiar la dirección.' };
	}

	resumen() {
		return this.pedir<ResumenPanel>('/v1/panel/resumen');
	}
	notificaciones(cursor?: string) {
		const q = cursor ? `?cursor=${encodeURIComponent(cursor)}` : '';
		return this.pedir<PaginaNotificaciones>(`/v1/panel/notificaciones${q}`);
	}
	/** Una que ya no está (404: se purgó) no es un error para quien la tocó. */
	async marcarNotificacionLeida(id: string) {
		await this.intentar(`/v1/panel/notificaciones/${encodeURIComponent(id)}/leida`, 'POST', {});
	}
	marcarNotificacionesLeidas() {
		return this.pedir<void>('/v1/panel/notificaciones/leidas', 'POST', {});
	}

	tablero(periodo: Periodo) {
		return this.pedir<Tablero>(`/v1/panel/tablero?periodo=${periodo}`);
	}
	guardarMeta(objetivo: number) {
		return this.pedir<void>('/v1/panel/meta', 'PUT', { objetivo });
	}
	ocultarAviso(aviso: 'certificado') {
		return this.pedir<void>(`/v1/panel/avisos/${aviso}/ocultar`, 'POST', {});
	}
	async recordarCarritos() {
		return (await this.pedir<{ enviados: number }>('/v1/panel/carritos/recordar', 'POST', {}))
			.enviados;
	}

	contextoHornero() {
		return this.pedir<ContextoHornero>('/v1/panel/hornero/contexto');
	}
	marcarIdea(idea: Idea, hecha: boolean) {
		return this.pedir<void>(`/v1/panel/hornero/ideas/${idea}`, 'POST', { hecha });
	}

	listarPedidos(f: FiltrosPedidos) {
		const q = new URLSearchParams();
		for (const [k, v] of Object.entries(f)) {
			if (v === undefined || v === '' || v === false || (Array.isArray(v) && !v.length)) continue;
			q.set(k, Array.isArray(v) ? v.join(',') : String(v));
		}
		return this.pedir<ListaPedidos>(`/v1/panel/pedidos?${q}`);
	}
	async pedido(numero: number) {
		const r = await this.intentar(`/v1/panel/pedidos/${numero}`, 'GET');
		return r.ok ? ((await r.json()) as PedidoDetalle) : null;
	}
	async cambiarEstado(numero: number, estado: EstadoPedido) {
		return resultado(await this.intentar(`/v1/panel/pedidos/${numero}/estado`, 'POST', { estado }));
	}
	pedidosEnLote(numeros: number[], accion: AccionLote) {
		return this.pedir<{ hechos: number; cambios: { numero: number; antes: EstadoPedido }[] }>(
			'/v1/panel/pedidos/lote',
			'POST',
			{ numeros, accion }
		);
	}
	async recordarPago(numero: number) {
		return resultado(await this.intentar(`/v1/panel/pedidos/${numero}/recordatorio`, 'POST', {}));
	}
	async saldarPedido(numero: number, saldado: boolean) {
		return resultado(
			await this.intentar(`/v1/panel/pedidos/${numero}/saldado`, saldado ? 'POST' : 'DELETE')
		);
	}
	async lanzamiento(id: string) {
		const r = await this.intentar(
			`/v1/panel/productos/${encodeURIComponent(id)}/lanzamiento`,
			'GET'
		);
		return r.ok ? ((await r.json()) as LanzamientoProducto) : null;
	}
	async guardarLanzamiento(id: string, c: ConfigLanzamiento) {
		const r = await this.intentar(
			`/v1/panel/productos/${encodeURIComponent(id)}/lanzamiento`,
			'PUT',
			c,
			[400]
		);
		if (r.status === 404) return null;
		if (!r.ok)
			return {
				error:
					((await r.json().catch(() => null)) as { detail?: string } | null)?.detail ??
					'No pudimos guardar el lanzamiento.'
			};
		return (await r.json()) as LanzamientoProducto;
	}

	listar(lista: NombreLista, f: FiltrosLista) {
		const q = new URLSearchParams();
		if (f.q) q.set('q', f.q);
		if (f.seg?.length) q.set('seg', f.seg.join(','));
		const facetas = Object.entries(f.facetas ?? {}).flatMap(([k, vs]) =>
			vs.map((v) => `${k}:${v}`)
		);
		if (facetas.length) q.set('facetas', facetas.join(','));
		if (f.rango) q.set('rango', String(f.rango));
		if (f.banderas?.length) q.set('banderas', f.banderas.join(','));
		if (f.orden) q.set('orden', f.orden);
		q.set('desde', String(f.desde ?? 0));
		q.set('cuantos', String(f.cuantos ?? 60));
		return this.pedir<ListaGenerica>(`/v1/panel/listas/${lista}?${q}`);
	}
	async accionLista(lista: NombreLista, ids: string[], accion: string, valor?: unknown) {
		return (
			await this.pedir<{ hechos: number }>(`/v1/panel/listas/${lista}/acciones`, 'POST', {
				ids,
				accion,
				valor
			})
		).hechos;
	}
	async producto(id: string) {
		const r = await this.intentar(`/v1/panel/productos/${encodeURIComponent(id)}`, 'GET');
		return r.ok ? conVariantes((await r.json()) as Producto) : null;
	}
	async guardarProducto(id: string | null, p: ProductoEditable) {
		if (!id) return conVariantes(await this.pedir<Producto>('/v1/panel/productos', 'POST', p));
		const r = await this.intentar(`/v1/panel/productos/${encodeURIComponent(id)}`, 'PUT', p);
		return r.ok ? conVariantes((await r.json()) as Producto) : null;
	}
	async borrarProducto(id: string) {
		return resultado(
			await this.intentar(`/v1/panel/productos/${encodeURIComponent(id)}`, 'DELETE')
		);
	}
	async duplicarProducto(id: string) {
		const r = await this.intentar(`/v1/panel/productos/${encodeURIComponent(id)}/duplicar`, 'POST');
		return r.ok ? conVariantes((await r.json()) as Producto) : null;
	}

	movimientosHoy() {
		return this.pedir<MovimientoStock[]>('/v1/panel/inventario/movimientos');
	}

	bandeja() {
		return this.pedir<{ conversaciones: ConversacionBreve[]; atendidasHoy: number }>(
			'/v1/panel/bandeja'
		);
	}
	async conversacion(id: string) {
		const r = await this.intentar(`/v1/panel/bandeja/${encodeURIComponent(id)}`, 'GET');
		return r.ok ? ((await r.json()) as Conversacion) : null;
	}
	async responder(id: string, c: { texto?: string; tomada?: boolean; leida?: boolean }) {
		return resultado(await this.intentar(`/v1/panel/bandeja/${encodeURIComponent(id)}`, 'POST', c));
	}
	carritos() {
		return this.pedir<Carritos>('/v1/panel/carritos');
	}
	guardarAutomaticos(a: AutomaticosCarritos) {
		return this.pedir<void>('/v1/panel/carritos', 'PUT', a);
	}
	tarjeta() {
		return this.pedir<Tarjeta>('/v1/panel/tarjeta');
	}
	guardarTarjeta(t: Tarjeta) {
		return this.pedir<void>('/v1/panel/tarjeta', 'PUT', t);
	}
	async recordarCarrito(id: string) {
		return resultado(
			await this.intentar(`/v1/panel/carritos/${encodeURIComponent(id)}/recordar`, 'POST', {})
		);
	}

	async crearPedido(n: PedidoNuevo) {
		const r = await this.intentar('/v1/panel/pedidos', 'POST', n);
		if (r.status === 409) return 'sin-stock' as const;
		// Una variante que no es del producto, inactiva o que falta: el pedido no se arma.
		if (r.status === 400 || r.status === 422) return 'invalido' as const;
		return (await r.json()) as { numero: number; total: number };
	}
	devoluciones() {
		return this.pedir<{ devoluciones: Devolucion[]; tasa: number }>('/v1/panel/devoluciones');
	}
	async resolverDevolucion(id: string, accion: 'aprobar' | 'cambio' | 'deshacer') {
		return resultado(
			await this.intentar(`/v1/panel/devoluciones/${encodeURIComponent(id)}`, 'POST', { accion })
		);
	}
	resenas() {
		return this.pedir<ResenasYPreguntas>('/v1/panel/resenas');
	}
	async responderPregunta(id: string, respuesta: string) {
		return resultado(
			await this.intentar(`/v1/panel/preguntas/${encodeURIComponent(id)}`, 'POST', { respuesta })
		);
	}
	async accionResena(id: string, accion: 'ocultar' | 'mostrar' | 'responder', respuesta = '') {
		return resultado(
			await this.intentar(`/v1/panel/resenas/${encodeURIComponent(id)}`, 'POST', {
				accion,
				respuesta
			})
		);
	}

	// ---------------------------------------------------------------- Importación

	async importacionActual() {
		const r = await this.pedir<Importacion | undefined>('/v1/panel/importaciones/actual');
		return r ?? null;
	}
	async crearImportacion(n: ImportacionNueva) {
		const r = await this.intentar('/v1/panel/importaciones', 'POST', n, [400]);
		if (r.ok) return (await r.json()) as Importacion;
		if (r.status === 400) return 'invalido';
		const p = (await r.json().catch(() => null)) as { code?: string } | null;
		if (p?.code === 'importacion_activa') return 'activa';
		if (p?.code === 'sin_fuente') return 'sin-fuente';
		if (p?.code === 'sin_cifrado') return 'sin-cifrado';
		error(r.status, 'No pudimos empezar la importación.');
	}
	async tiendanubeDisponible() {
		const r = await this.pedir<{ disponible: boolean }>('/v1/panel/importaciones/tiendanube');
		return !!r?.disponible;
	}
	async conectarTiendanube() {
		const r = await this.intentar('/v1/panel/importaciones/tiendanube/conectar', 'POST', {});
		if (!r.ok) return null;
		return ((await r.json()) as { url: string }).url;
	}
	/** Go consume el state, canjea el código y crea la importación: el token no pasa por acá. */
	async volverTiendanube(codigo: string, estado: string): Promise<Importacion | ErrorVueltaTN> {
		const r = await this.intentar(
			'/v1/panel/importaciones/tiendanube/vuelta',
			'POST',
			{ codigo, estado },
			[400, 409, 502]
		);
		if (r.ok) return (await r.json()) as Importacion;
		if (r.status === 400) return 'estado';
		if (r.status === 502) return 'canje';
		const p = (await r.json().catch(() => null)) as { code?: string } | null;
		if (p?.code === 'importacion_activa') return 'activa';
		if (p?.code === 'sin_cifrado') return 'sin-cifrado';
		return 'sin-tiendanube';
	}
	async cancelarImportacion(id: string) {
		const r = await this.intentar(
			`/v1/panel/importaciones/${encodeURIComponent(id)}/cancelar`,
			'POST',
			{}
		);
		if (r.ok) return (await r.json()) as Importacion;
		return r.status === 404 ? 'no-existe' : 'terminada';
	}
	/** Con la API avanza el worker por el outbox: acá sólo se lee cómo va. */
	avanzarImportacion() {
		return this.importacionActual();
	}
	historialCliente(id: string) {
		return this.pedir<PedidoHistorial[]>(`/v1/panel/clientes/${encodeURIComponent(id)}/historial`);
	}

	async crearLink(l: { monto: number; concepto: string; cuotas: boolean; pedido: number | null }) {
		return (await this.pedir<{ codigo: string }>('/v1/panel/links', 'POST', l)).codigo;
	}

	dominios() {
		return this.pedir<Dominios>('/v1/panel/dominios');
	}
	async conectarDominio(dominio: string) {
		return resultado(await this.intentar('/v1/panel/dominios', 'POST', { dominio }));
	}
	async dominioPrincipal(dominio: string) {
		return resultado(
			await this.intentar(`/v1/panel/dominios/${encodeURIComponent(dominio)}`, 'PUT')
		);
	}
	async quitarDominio(dominio: string) {
		return resultado(
			await this.intentar(`/v1/panel/dominios/${encodeURIComponent(dominio)}`, 'DELETE')
		);
	}
	preferenciasAvisos() {
		return this.pedir<PreferenciasAvisos>('/v1/panel/notificaciones/preferencias');
	}
	async guardarPreferenciasAvisos(p: PreferenciasAvisos) {
		await this.pedir<void>('/v1/panel/notificaciones/preferencias', 'PUT', p);
	}
	pushWeb() {
		return this.pedir<EstadoPushWeb>('/v1/panel/push-web');
	}
	async suscribirPushWeb(s: SuscripcionWeb) {
		return resultado(await this.intentar('/v1/panel/push-web', 'POST', s, [400]));
	}
	async bajaPushWeb() {
		await this.pedir<void>('/v1/panel/push-web', 'DELETE');
	}
	async probarPushWeb() {
		return resultado(await this.intentar('/v1/panel/push-web/prueba', 'POST'));
	}
	misDatos() {
		return this.pedir<MisDatos>('/v1/panel/datos');
	}
	async exportar(tipo: TipoExport) {
		return (await this.pedir<{ filas: (string | number | null)[][] }>(`/v1/panel/datos/${tipo}`))
			.filas;
	}
	async registrarExportacion(tipo: TipoExport | 'todo', formato: 'csv' | 'excel') {
		await this.pedir<void>('/v1/panel/datos/exportaciones', 'POST', { tipo, formato });
	}
	async borrarInactivos() {
		return (await this.pedir<{ clientes: number }>('/v1/panel/datos/inactivos', 'DELETE')).clientes;
	}
	async frenarBorrado() {
		await this.pedir<void>('/v1/panel/datos/inactivos', 'POST');
	}
	async cerrarCuenta(confirmacion: string) {
		return resultado(await this.intentar('/v1/panel/cuenta', 'DELETE', { confirmacion }));
	}
	planCuenta() {
		return this.pedir<PlanCuenta>('/v1/panel/cuenta/plan');
	}
	async comprarPack() {
		await this.pedir<void>('/v1/panel/cuenta/pack', 'POST');
	}
	async devolverPack() {
		return resultado(await this.intentar('/v1/panel/cuenta/pack', 'DELETE'));
	}
	cancelarPlan(motivo: number, accion: 'cancelar' | 'oferta') {
		return this.pedir<{ hasta?: string | null; beneficio?: Beneficio }>(
			'/v1/panel/cuenta/cancelacion',
			'POST',
			{
				motivo,
				accion
			}
		);
	}
	async reactivarPlan() {
		await this.pedir<void>('/v1/panel/cuenta/cancelacion', 'DELETE');
	}
	acceso() {
		return this.pedir<Acceso>('/v1/auth/tiendas');
	}
	async elegirTienda(id: string) {
		return resultado(await this.intentar('/v1/auth/tienda', 'PUT', { tiendaId: id }));
	}
	async reenviarInvitacion(id: string) {
		return resultado(
			await this.intentar(`/v1/panel/equipo/${encodeURIComponent(id)}/reenviar`, 'POST')
		);
	}
	async cambiarRol(id: string, rol: RolInvitable) {
		return resultado(
			await this.intentar(`/v1/panel/equipo/${encodeURIComponent(id)}`, 'PUT', { rol })
		);
	}
	async quitarDelEquipo(id: string) {
		return resultado(await this.intentar(`/v1/panel/equipo/${encodeURIComponent(id)}`, 'DELETE'));
	}
	configuracion() {
		return this.pedir<Configuracion>('/v1/panel/configuracion');
	}
	async guardarNegocio(n: Negocio) {
		await this.pedir<void>('/v1/panel/configuracion/negocio', 'PUT', n);
	}
	async fijarOpcion(grupo: 'seguridad' | 'avisos', clave: string, activo: boolean) {
		return resultado(
			await this.intentar(`/v1/panel/configuracion/${grupo}/${encodeURIComponent(clave)}`, 'PUT', {
				activo
			})
		);
	}
	async cerrarSesion(id: string) {
		return resultado(await this.intentar(`/v1/panel/sesiones/${encodeURIComponent(id)}`, 'DELETE'));
	}
	async cerrarOtrasSesiones() {
		return (await this.pedir<{ cerradas: number }>('/v1/panel/sesiones', 'DELETE')).cerradas;
	}
	async invitar(email: string, rol: RolInvitable) {
		const r = await this.intentar('/v1/panel/equipo', 'POST', { contacto: email, rol });
		if (r.status === 409) return 'repetido' as const;
		return (await r.json()) as Miembro;
	}
	/** El abono (§56): la API cotiza y aplica; url vacía si quedó hecho, si no adónde ir en Mercado Pago. */
	async cambiarPlan(plan: IdPlan, ciclo?: 'mes' | 'anio') {
		const ciclo_ = ciclo ?? (await this.planCuenta()).ciclo;
		const r = await this.pedir<{ url: string }>('/v1/panel/cuenta/abono', 'POST', {
			plan,
			ciclo: ciclo_
		});
		return r.url ? r : undefined;
	}
	async volverAbono(v: { pago?: string; suscripcion?: string }) {
		await this.pedir<void>('/v1/panel/cuenta/abono/vuelta', 'POST', v);
	}
	// Las demás integraciones llegan una por PR: Mercado Pago, Mercado Libre, la medición y ARCA son reales.
	async integraciones() {
		const e = estadoInicial(false);
		const [mp, ml, ia, m, arca] = await Promise.all([
			this.mercadoPago(),
			this.mercadoLibre(),
			this.pedir<{ permisos: Record<string, boolean> }>('/v1/panel/ia'),
			this.medicion(),
			this.arca()
		]);
		e.conectadas.mp = mp.conectada;
		e.conectadas.ml = ml.conectada;
		Object.assign(e.conectadas, conectadasDeMedicion(m));
		e.conectadas.arca = arca.conectada;
		e.opciones.arca = { ...e.opciones.arca, auto: arca.facturarAlCobrar };
		e.permisosIa = { ...e.permisosIa, ...ia.permisos };
		return e;
	}
	arca() {
		return this.pedir<ArcaConfig>('/v1/panel/arca');
	}
	async arcaPedido(a: ArcaAlta): Promise<ArcaPedido> {
		const r = await this.conProblema<{ pedido: string }>('/v1/panel/arca/pedido', 'POST', a);
		return 'error' in r ? r : { pedido: r.pedido };
	}
	arcaCertificado(pem: string) {
		return this.conProblema<ArcaCertificado>('/v1/panel/arca/certificado', 'PUT', {
			certificado: pem
		});
	}
	async arcaAjustes(a: ArcaAjustes) {
		const r = await this.conProblema<null>('/v1/panel/arca/ajustes', 'PUT', a);
		return r && 'error' in r ? r : null;
	}
	async arcaProbar(): Promise<ArcaPrueba> {
		const r = await this.conProblema<ArcaPrueba>('/v1/panel/arca/probar', 'POST', {});
		return 'error' in r ? { ok: false, ultimo: 0, detalle: r.error } : r;
	}
	medicion() {
		return this.pedir<Medicion>('/v1/panel/integraciones/medicion');
	}
	async guardarMedicion(m: Medicion) {
		return resultado(await this.intentar('/v1/panel/integraciones/medicion', 'PUT', m));
	}
	entregas() {
		return this.pedir<EstadoEntregas>('/v1/panel/configuracion/entregas');
	}
	guardarEntregas(e: Entregas) {
		return this.pedir<void>('/v1/panel/configuracion/entregas', 'PUT', e);
	}
	guardarMayorista(m: VentaMayorista) {
		return this.pedir<void>('/v1/panel/configuracion/mayorista', 'PUT', m);
	}
	async ajustarIntegracion(a: AjusteIntegracion) {
		// Por ahora sólo los permisos de las IA se guardan en la API; el resto llega con cada integración.
		if (a.tipo !== 'permisoIa') return 'no-se-puede' as const;
		return resultado(
			await this.intentar(`/v1/panel/ia/permisos/${encodeURIComponent(a.id)}`, 'PUT', {
				valor: a.valor
			})
		);
	}
	mercadoLibre() {
		return this.pedir<ConexionML>('/v1/panel/integraciones/mercadolibre');
	}
	async conectarMercadoLibre() {
		const res = await this.intentar('/v1/panel/integraciones/mercadolibre/conectar', 'POST');
		if (!res.ok) return null;
		return ((await res.json()) as { url: string }).url;
	}
	async volverMercadoLibre(codigo: string, estado: string) {
		const res = await this.intentar('/v1/panel/integraciones/mercadolibre/vuelta', 'POST', {
			codigo,
			estado
		});
		return res.ok ? 'ok' : 'no-se-puede';
	}
	async desconectarMercadoLibre() {
		await this.pedir<void>('/v1/panel/integraciones/mercadolibre', 'DELETE');
	}
	async ajustarMercadoLibre(recargo: number, pausarSinStock: boolean) {
		return resultado(
			await this.intentar('/v1/panel/integraciones/mercadolibre', 'PUT', {
				recargo,
				pausarSinStock
			})
		);
	}
	async publicarEnMercadoLibre(productoId: string) {
		const res = await this.f(
			`${this.base}/v1/panel/productos/${encodeURIComponent(productoId)}/mercadolibre`,
			{
				method: 'POST',
				headers: {
					accept: 'application/json',
					...(this.token ? { authorization: `Bearer ${this.token}` } : {})
				}
			}
		);
		if (res.ok) return (await res.json()) as PublicacionML;
		const p = (await res.json().catch(() => null)) as { detail?: string } | null;
		return { error: p?.detail ?? 'Mercado Libre no respondió.' };
	}
	async despublicarDeMercadoLibre(productoId: string) {
		await this.intentar(
			`/v1/panel/productos/${encodeURIComponent(productoId)}/mercadolibre`,
			'DELETE'
		);
	}
	mercadoPago() {
		return this.pedir<ConexionMP>('/v1/panel/integraciones/mercadopago');
	}
	async conectarMercadoPago() {
		const res = await this.intentar('/v1/panel/integraciones/mercadopago/conectar', 'POST');
		if (!res.ok) return null;
		return ((await res.json()) as { url: string }).url;
	}
	async volverMercadoPago(codigo: string, estado: string) {
		const res = await this.intentar('/v1/panel/integraciones/mercadopago/vuelta', 'POST', {
			codigo,
			estado
		});
		return res.ok ? 'ok' : 'no-se-puede';
	}
	async desconectarMercadoPago() {
		await this.pedir<void>('/v1/panel/integraciones/mercadopago', 'DELETE');
	}
	miTienda() {
		return this.pedir<MiTienda>('/v1/panel/mi-tienda');
	}
	async guardarMiTienda(d: DisenoTienda) {
		await this.pedir<void>('/v1/panel/mi-tienda', 'PUT', d);
	}
	async publicarTienda() {
		await this.pedir<void>('/v1/panel/mi-tienda/publicar', 'POST');
	}
	finanzas() {
		return this.pedir<Finanzas>('/v1/panel/finanzas');
	}
	async retirar() {
		const r = await this.intentar('/v1/panel/finanzas/retiro', 'POST');
		if (r.status === 409) return 'sin-cobros' as const;
		return (await r.json()) as { monto: number; cuenta: string };
	}
	actividad(f: FiltroActividad = {}) {
		const q = consultaDe(f);
		return this.pedir<PaginaActividad>('/v1/panel/actividad' + (q ? `?${q}` : ''));
	}
	async exportarActividad(f: FiltroActividad = {}) {
		const q = consultaDe({ ...f, cursor: undefined });
		return (
			await this.pedir<{ filas: string[][] }>('/v1/panel/actividad/exportar' + (q ? `?${q}` : ''))
		).filas;
	}
	async deshacerEvento(id: string) {
		return resultado(
			await this.intentar(`/v1/panel/actividad/${encodeURIComponent(id)}/deshacer`, 'POST')
		);
	}
	paraHacer() {
		return this.pedir<ParaHacer>('/v1/panel/para-hacer');
	}
	async marcarTarea(id: string, hecha: boolean) {
		return resultado(
			await this.intentar(`/v1/panel/para-hacer/${encodeURIComponent(id)}`, 'PUT', { hecha })
		);
	}
	marketing() {
		return this.pedir<Marketing>('/v1/panel/marketing');
	}
	newsletter() {
		return this.pedir<Newsletter>('/v1/panel/newsletter');
	}
	async borrarSuscriptor(id: string) {
		return resultado(
			await this.intentar(`/v1/panel/newsletter/suscriptores/${encodeURIComponent(id)}`, 'DELETE')
		);
	}
	async mandarNovedad(n: NovedadNueva) {
		const r = await this.intentar('/v1/panel/newsletter/envios', 'POST', n, [400]);
		if (r.status === 409) return 'sin-suscriptos' as const;
		if (!r.ok) {
			const p = (await r.json().catch(() => null)) as { detail?: string } | null;
			return { error: p?.detail ?? 'Revisá la novedad.' };
		}
		return (await r.json()) as EnvioNewsletter;
	}
	async crearCampana(c: Parameters<Datos['crearCampana']>[0]) {
		const r = await this.intentar('/v1/panel/campanas', 'POST', c);
		if (r.status === 409) return 'sin-whatsapp' as const;
		return (await r.json()) as Campana;
	}
	canales() {
		return this.pedir<Canales>('/v1/panel/canales');
	}
	async fijarCanal(id: string, activo: boolean) {
		return resultado(
			await this.intentar(`/v1/panel/canales/${encodeURIComponent(id)}`, 'PUT', { activo })
		);
	}
	informes() {
		return this.pedir<Informes>('/v1/panel/informes');
	}
	cupones() {
		return this.pedir<Cupones>('/v1/panel/cupones');
	}
	async crearCupon(n: CuponNuevo) {
		const r = await this.intentar('/v1/panel/cupones', 'POST', n);
		if (r.status === 409) return 'repetido' as const;
		return (await r.json()) as Cupon;
	}
	async accionCupon(id: string, accion: AccionCupon, minimo?: number) {
		const r = await this.intentar(`/v1/panel/cupones/${encodeURIComponent(id)}`, 'POST', {
			accion,
			minimo
		});
		if (r.status === 404) return 'no-existe' as const;
		if (r.status === 409) return 'no-se-puede' as const;
		return (await r.json()) as Cupon;
	}
	async descartarConsejo() {
		await this.intentar('/v1/panel/cupones/consejo', 'DELETE');
	}
	async fijarAutomatica(id: string, activa: boolean) {
		return resultado(
			await this.intentar(`/v1/panel/automaticas/${encodeURIComponent(id)}`, 'PUT', { activa })
		);
	}
	async accionRelampago(accion: AccionRelampago) {
		return resultado(await this.intentar('/v1/panel/relampago', 'POST', { accion }));
	}
	listasPrecio() {
		return this.pedir<ListasPrecio>('/v1/panel/listas-precio');
	}
	async aumentarPrecios(pct: number, redondeo: 'no' | '90' | '500', programar: boolean) {
		const r = await this.pedir<{ productos: number; para?: string | null }>(
			'/v1/panel/listas-precio/aumento',
			'POST',
			{ pct, redondeo, programar }
		);
		return { productos: r.productos, para: r.para ?? null };
	}
	async deshacerAumento() {
		return resultado(await this.intentar('/v1/panel/listas-precio/aumento/deshacer', 'POST'));
	}
	async cancelarAumento() {
		return resultado(await this.intentar('/v1/panel/listas-precio/programado', 'DELETE'));
	}

	// ---- Turnos con seña

	ajustesTurnos() {
		return this.pedir<AjustesTurnos>('/v1/panel/turnos/ajustes');
	}
	guardarAjustesTurnos(a: AjustesTurnos) {
		return this.pedir<AjustesTurnos>('/v1/panel/turnos/ajustes', 'PUT', a);
	}
	serviciosTurnos() {
		return this.pedir<Servicio[]>('/v1/panel/turnos/servicios');
	}
	async guardarServicio(id: string | null, s: ServicioEditable) {
		if (!id) return this.pedir<Servicio>('/v1/panel/turnos/servicios', 'POST', s);
		const r = await this.intentar(`/v1/panel/turnos/servicios/${encodeURIComponent(id)}`, 'PUT', s);
		return r.ok ? ((await r.json()) as Servicio) : null;
	}
	async borrarServicio(id: string) {
		return resultado(
			await this.intentar(`/v1/panel/turnos/servicios/${encodeURIComponent(id)}`, 'DELETE')
		);
	}
	profesionalesTurnos() {
		return this.pedir<Profesional[]>('/v1/panel/turnos/profesionales');
	}
	async guardarProfesional(id: string | null, p: ProfesionalEditable) {
		if (!id) return this.pedir<Profesional>('/v1/panel/turnos/profesionales', 'POST', p);
		const r = await this.intentar(
			`/v1/panel/turnos/profesionales/${encodeURIComponent(id)}`,
			'PUT',
			p
		);
		return r.ok ? ((await r.json()) as Profesional) : null;
	}
	async borrarProfesional(id: string) {
		return resultado(
			await this.intentar(`/v1/panel/turnos/profesionales/${encodeURIComponent(id)}`, 'DELETE')
		);
	}
	agendaTurnos(desde: string, hasta: string, profesional?: string) {
		const q = new URLSearchParams({ desde, hasta });
		if (profesional) q.set('profesional', profesional);
		return this.pedir<Turno[]>(`/v1/panel/turnos?${q}`);
	}
	async crearTurno(n: TurnoNuevo) {
		const r = await this.intentar('/v1/panel/turnos', 'POST', n, [400, 422]);
		if (r.status === 409) return 'ocupado' as const;
		// Como en el sandbox: un servicio o un horario que no van es "inválido", no un error de la API.
		if (r.status === 404 || r.status === 400 || r.status === 422) return 'invalido' as const;
		return (await r.json()) as Turno;
	}
	async estadoTurno(id: string, estado: EstadoTurno) {
		const r = await this.intentar(`/v1/panel/turnos/${encodeURIComponent(id)}`, 'PATCH', {
			estado
		});
		return r.ok ? ((await r.json()) as Turno) : ('no-existe' as const);
	}
	async devolverSena(id: string) {
		const r = await this.intentar(
			`/v1/panel/turnos/${encodeURIComponent(id)}/devolver-sena`,
			'POST',
			{}
		);
		if (r.status === 404) return 'no-existe' as const;
		if (r.status === 409) {
			const p = (await r.json().catch(() => null)) as { code?: string } | null;
			return p?.code === 'sin_cobros' ? ('sin-cobros' as const) : ('sin-sena' as const);
		}
		return (await r.json()) as Turno;
	}

	// ---- Suscripciones de la tienda (§51)

	suscripcionesPanel() {
		return this.pedir<PanelSuscripciones>('/v1/panel/suscripciones');
	}
	async guardarPlanSuscripcion(id: string | null, p: PlanSuscripcionEditable) {
		const r = id
			? await this.intentar(
					`/v1/panel/suscripciones/planes/${encodeURIComponent(id)}`,
					'PUT',
					p,
					[400]
				)
			: await this.intentar('/v1/panel/suscripciones/planes', 'POST', p, [400]);
		if (r.status === 400) return 'invalido' as const;
		return r.ok ? ((await r.json()) as PlanSuscripcion) : null;
	}
	async borrarPlanSuscripcion(id: string) {
		return resultado(
			await this.intentar(`/v1/panel/suscripciones/planes/${encodeURIComponent(id)}`, 'DELETE')
		);
	}
	async cancelarSuscriptor(id: string) {
		const r = await this.intentar(
			`/v1/panel/suscripciones/${encodeURIComponent(id)}/cancelar`,
			'POST',
			{},
			[502]
		);
		if (r.status === 502) return 'mercadopago' as const;
		return r.ok ? ((await r.json()) as Suscriptor) : ('no-existe' as const);
	}

	/** Como `pedir`, pero 404 y 409 vuelven como respuesta (son resultados, no errores). */
	/** Como `pedir`, pero un 4xx/5xx vuelve como { error } con el detalle de Go, para mostrarlo. */
	private async conProblema<T>(
		ruta: string,
		metodo: 'POST' | 'PUT',
		cuerpo: unknown
	): Promise<T | { error: string }> {
		const res = await this.f(this.base + ruta, {
			method: metodo,
			headers: {
				accept: 'application/json',
				'content-type': 'application/json',
				...(this.token ? { authorization: `Bearer ${this.token}` } : {})
			},
			body: JSON.stringify(cuerpo)
		});
		if (res.status === 401) error(401, 'Tu sesión venció.');
		if (!res.ok) {
			const p = (await res.json().catch(() => null)) as { detail?: string } | null;
			return { error: p?.detail ?? 'ARCA no respondió bien. Probá de nuevo en un rato.' };
		}
		if (res.status === 204) return null as T;
		return (await res.json()) as T;
	}

	private async intentar(
		ruta: string,
		metodo: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE',
		cuerpo?: unknown,
		/** Estados de error que resuelve quien llama, además de 404 y 409. */
		aceptar: number[] = []
	) {
		const res = await this.f(this.base + ruta, {
			method: metodo,
			headers: {
				accept: 'application/json',
				...(cuerpo === undefined ? {} : { 'content-type': 'application/json' }),
				...(this.token ? { authorization: `Bearer ${this.token}` } : {})
			},
			body: cuerpo === undefined ? undefined : JSON.stringify(cuerpo)
		});
		if (!res.ok && res.status !== 404 && res.status !== 409 && !aceptar.includes(res.status)) {
			const p = (await res.json().catch(() => null)) as { detail?: string; code?: string } | null;
			error(res.status, { message: p?.detail ?? 'La API no respondió bien.', code: p?.code });
		}
		return res;
	}
}

const resultado = (r: Response): Resultado =>
	r.ok ? 'ok' : r.status === 404 ? 'no-existe' : 'no-se-puede';
