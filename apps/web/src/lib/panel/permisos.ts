/**
 * Roles fijos del equipo de una tienda (docs/decisiones.md §37). Go los aplica en cada ruta del
 * panel (403 si no); acá sólo se decide qué mostrar. La tabla de qué puede cada rol es la misma que
 * la de Go: el test la compara con apps/api/internal/dominio/roles/testdata/roles.json.
 */
export const ROLES = ['duenio', 'admin', 'ventas', 'local'] as const;
export type Rol = (typeof ROLES)[number];
export const INVITABLES = ['admin', 'ventas', 'local'] as const;
export type RolInvitable = (typeof INVITABLES)[number];

export const NOMBRE_ROL: Record<Rol, string> = {
	duenio: 'Dueña',
	admin: 'Administración',
	ventas: 'Ventas',
	local: 'Local'
};

export const PERMISOS = [
	'panel',
	'pedidos',
	'clientes',
	'productos',
	'catalogo',
	'finanzas',
	'tienda',
	'cuenta'
] as const;
export type Permiso = (typeof PERMISOS)[number];

export const PERMISOS_DE: Record<Rol, Permiso[]> = {
	duenio: ['panel', 'pedidos', 'clientes', 'productos', 'catalogo', 'finanzas', 'tienda', 'cuenta'],
	admin: ['panel', 'pedidos', 'clientes', 'productos', 'catalogo', 'finanzas', 'tienda'],
	ventas: ['panel', 'pedidos', 'clientes', 'productos'],
	local: ['panel', 'pedidos']
};

/** Qué hace cada rol, en una línea (invitar y la lista del equipo). */
export const QUE_HACE: Record<Rol, string> = {
	duenio: 'Todo, también el plan y el equipo',
	admin: 'Todo menos el plan, la facturación de Tu Local Web y el equipo',
	ventas: 'Pedidos, clientes y productos, sin precios, costos ni finanzas',
	local: 'Punto de venta y pedidos'
};

/** Lo que cada rol puede y no puede, para la tarjeta «Qué puede cada rol». */
export const DETALLE_ROL: Record<Rol, { si: string[]; no: string[] }> = {
	duenio: { si: ['Todo el panel', 'Plan y facturación', 'Invitar y quitar personas'], no: [] },
	admin: {
		si: [
			'Pedidos, clientes y productos',
			'Precios, costos y finanzas',
			'Configuración de la tienda'
		],
		no: ['Plan y facturación', 'El equipo']
	},
	ventas: {
		si: ['Pedidos y punto de venta', 'Clientes y WhatsApp', 'Editar productos y stock'],
		no: ['Precios y costos', 'Finanzas', 'Configuración']
	},
	local: {
		si: ['Punto de venta', 'Pedidos y envíos', 'Ver productos'],
		no: ['Clientes', 'Editar productos', 'Finanzas y configuración']
	}
};

/** Chips cortos de cada persona en la lista del equipo. */
export const AREAS: Record<Rol, string[]> = {
	duenio: ['Todo'],
	admin: ['Todo menos plan y equipo'],
	ventas: ['Pedidos', 'Clientes', 'Productos'],
	local: ['Punto de venta', 'Pedidos']
};

/**
 * El permiso que hace falta para ver cada sección del panel (por su nombre en la navegación). Lo
 * que no está acá lo ve cualquiera del equipo.
 */
export const PERMISO_SECCION: Record<string, Permiso> = {
	// Quién hizo qué: la dueña y administración (docs/decisiones.md §39).
	Actividad: 'tienda',
	Pedidos: 'pedidos',
	Turnos: 'pedidos',
	// Planes y suscriptores (§51); crear y editar planes pide Tienda (lo decide Go).
	Suscripciones: 'pedidos',
	Devoluciones: 'pedidos',
	'Ventas en local': 'pedidos',
	Productos: 'pedidos',
	Inventario: 'pedidos',
	Envíos: 'pedidos',
	'Carritos abandonados': 'clientes',
	WhatsApp: 'clientes',
	Clientes: 'clientes',
	'Reseñas y preguntas': 'clientes',
	'Listas de precio': 'catalogo',
	Cupones: 'catalogo',
	Finanzas: 'finanzas',
	Cobros: 'finanzas',
	Facturación: 'finanzas',
	Informes: 'finanzas',
	Marketing: 'tienda',
	Newsletter: 'tienda',
	Canales: 'tienda',
	'Mi tienda': 'tienda',
	'Tarjeta para paquetes': 'tienda',
	Integraciones: 'tienda',
	Dominio: 'tienda',
	Notificaciones: 'tienda',
	'Mis datos': 'tienda',
	'Plan y facturación': 'cuenta',
	'Usuarios y permisos': 'cuenta'
};

/** ¿Estos permisos alcanzan para ver la sección? */
export const puedeVer = (permisos: readonly string[], seccion: string) => {
	const p = PERMISO_SECCION[seccion];
	return !p || permisos.includes(p);
};

/** Los motivos y la oferta de cada uno en la cancelación en 3 pasos. */
export const MOTIVOS = [
	'Es caro para lo que vendo',
	'Me falta una función',
	'Voy a pausar el negocio',
	'Me voy a otra plataforma',
	'Otro motivo'
] as const;
