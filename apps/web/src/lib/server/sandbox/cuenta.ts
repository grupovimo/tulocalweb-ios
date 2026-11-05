/**
 * La cuenta de la tienda de ejemplo: negocio, equipo, seguridad, sesiones, avisos y plan. La usan
 * Configuración y Cuenta (design/PanelOperacion.dc.html y PanelCuenta.dc.html). La dueña es Valeria
 * (la que saluda el Resumen); quien la prueba es ella.
 */
import type { components } from '@mostrador/contrato/esquema';
import type { Demo } from './demo.ts';

type Configuracion = components['schemas']['Configuracion'];
type Miembro = components['schemas']['Miembro'];
export type Negocio = components['schemas']['Negocio'];
export type Plan = Configuracion['plan']['actual'];

const MIN = 60_000;
const iso = (t: number) => new Date(t).toISOString();

/** El equipo de la demo. Carla es quien cobra en la caja del local. */
export const EQUIPO_DEMO = [
	{
		id: 'u1',
		nombre: 'Valeria Sosa',
		usuario: 'valeria',
		rol: 'duenio',
		foto: '1556745753-b2904692b3cd',
		haceMin: 0
	},
	{
		id: 'u2',
		nombre: 'Nahuel Sosa',
		usuario: 'nahuel',
		rol: 'admin',
		foto: '1545262722-9e0d80a0bc01',
		haceMin: 120
	},
	{
		id: 'u3',
		nombre: 'Martín Ríos',
		usuario: 'martin',
		rol: 'ventas',
		foto: '1687422811062-a966b55cb217',
		haceMin: 20
	},
	{
		id: 'u4',
		nombre: 'Carla Díaz',
		usuario: 'carla',
		rol: 'ventas',
		foto: '1520333789090-1afc82db536a',
		haceMin: 10
	}
] as const;

export type Invitado = { id: string; contacto: string; rol: Miembro['rol']; en: number };
export type EstadoCuenta = {
	negocio: Partial<Negocio>;
	seguridad: Partial<Configuracion['seguridad']>;
	avisos: Partial<Configuracion['avisos']>;
	cerradas: Set<string>;
	quitados: Set<string>;
	roles: Map<string, Miembro['rol']>;
	invitados: Invitado[];
	/** activo: ya «autorizó el débito» en la demo; programado: el cambio que rige al terminar la prueba. */
	plan: {
		actual?: Plan;
		ciclo?: 'mes' | 'anio';
		activo?: boolean;
		programado?: { plan: Plan; ciclo: 'mes' | 'anio'; desde: string } | null;
	};
	/** Packs de 2.000 respuestas comprados, por mes ("2026-10"). */
	packs: Record<string, number>;
	/** Cancelado al final del período (o null). */
	cancelado: { en: number; motivo: number } | null;
	/** Oferta aceptada para quedarse. */
	oferta: { motivo: number; en: number } | null;
};

const dominio = (tienda: string) =>
	tienda
		.toLowerCase()
		.normalize('NFD')
		.replace(/[^a-z0-9]/g, '');

export function configuracionDemo(
	d: Demo,
	tienda: string,
	ahora: number,
	diasPrueba: number,
	e: EstadoCuenta
): Omit<Configuracion, 'mayorista'> {
	const dom = `${dominio(tienda)}.com.ar`;
	const negocio: Negocio = {
		nombre: tienda,
		razonSocial: 'Valeria Sosa',
		cuit: '27-31458902-0',
		condicion: 'Responsable inscripta',
		direccion: `Av. San Martín 1240, ${d.catalogo.ciudad}`,
		whatsapp: '+54 341 555-0142',
		email: `hola@${dom}`,
		horario: 'Lun a sáb · 8 a 20 h',
		...e.negocio
	};
	const equipo: Miembro[] = [
		...EQUIPO_DEMO.filter((m) => !e.quitados.has(m.id)).map((m) => ({
			id: m.id,
			nombre: m.nombre,
			contacto: `${m.usuario}@${dom}`,
			rol: e.roles.get(m.id) ?? m.rol,
			foto: m.foto,
			ultimo: iso(ahora - m.haceMin * MIN),
			pendiente: false,
			vence: null,
			vos: m.id === 'u1',
			dobleFactor: m.id !== 'u3'
		})),
		...e.invitados.map((i) => ({
			id: i.id,
			nombre: i.contacto,
			contacto: i.contacto,
			rol: e.roles.get(i.id) ?? i.rol,
			foto: null,
			ultimo: null,
			pendiente: true,
			vence: iso(i.en + 7 * 24 * 60 * MIN),
			vos: false,
			dobleFactor: false
		}))
	];
	// En la demo no hay IP: las sesiones de ejemplo van sin ella.
	const sesiones = [
		{
			id: 's1',
			tipo: 'compu' as const,
			nombre: 'Este navegador',
			ubicacion: d.catalogo.ciudad,
			ip: '',
			creada: iso(ahora - 3 * 24 * 60 * MIN),
			cuando: iso(ahora),
			actual: true
		},
		{
			id: 's3',
			tipo: 'caja' as const,
			nombre: 'Caja del local',
			ubicacion: d.catalogo.ciudad,
			ip: '',
			creada: iso(ahora - 20 * 24 * 60 * MIN),
			cuando: iso(ahora - 10 * MIN),
			actual: false
		},
		{
			id: 's2',
			tipo: 'celu' as const,
			nombre: 'App de Tu Local Web en iPhone',
			ubicacion: d.catalogo.ciudad,
			ip: '',
			creada: iso(ahora - 9 * 24 * 60 * MIN),
			cuando: iso(ahora - 120 * MIN),
			actual: false
		}
	].filter((s) => !e.cerradas.has(s.id));
	return {
		negocio,
		equipo,
		seguridad: {
			avisoIngreso: true,
			inactivas: false,
			...e.seguridad
		},
		sesiones,
		avisos: {
			venta: true,
			chats: true,
			stock: true,
			diario: true,
			ideas: true,
			novedades: false,
			...e.avisos
		},
		plan: {
			actual: e.plan.actual ?? 'pro',
			ciclo: e.plan.ciclo ?? 'mes',
			enPrueba: true,
			diasPrueba
		}
	};
}
