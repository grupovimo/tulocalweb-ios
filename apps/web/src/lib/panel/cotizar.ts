/**
 * La cotización de un cambio de plan (§56): qué se paga hoy, desde cuándo rige y cuánto se cobra después.
 * Espejo de apps/api/internal/dominio/abono/abono.go, con la misma tabla de casos
 * (apps/api/internal/dominio/abono/testdata/cotizar.json). La API manda las cotizaciones hechas; esto lo
 * usa el sandbox de la demo.
 */
import { PLANES, precioAnual, type IdPlan } from './planes.ts';

export type Ciclo = 'mes' | 'anio';
export type TipoCambio =
	'actual' | 'prueba' | 'alta' | 'sube' | 'baja' | 'ciclo' | 'gratis' | 'programado' | 'mantener';
export type ComoPaga = 'ninguno' | 'suscripcion' | 'diferencia' | 'actualizar' | 'cancelar';

export type EstadoAbono = {
	plan: IdPlan;
	ciclo: Ciclo;
	finPrueba: string;
	/** Hay un débito autorizado que va a cobrar en `proximo`. */
	activo: boolean;
	/** Hasta cuándo está pago el plan vigente. */
	proximo: string | null;
	pendiente: { plan: IdPlan; ciclo: Ciclo; desde: string } | null;
};

export type Cotizacion = {
	plan: IdPlan;
	ciclo: Ciclo;
	tipo: TipoCambio;
	hoy: number;
	rige: string;
	hasta: string | null;
	cobro: { fecha: string; importe: number } | null;
	porCiclo: number;
	/** El detalle de `hoy` (suman `hoy`). */
	lineas: { texto: string; importe: number }[];
	pago: ComoPaga;
	inicio: string | null;
};

export class ErrorCotizar extends Error {
	constructor(public motivo: 'invalido' | 'programado') {
		super(motivo);
	}
}

const NOMBRES: Record<IdPlan, string> = {
	gratis: 'Gratis',
	emprende: 'Emprende',
	pro: 'Pro',
	negocio: 'Negocio'
};

/** Lo que se cobra cada vez: el anual, una vez por año (25 % menos). */
export const precioDe = (plan: IdPlan, ciclo: Ciclo) => {
	const p = PLANES.find((x) => x.id === plan);
	if (!p) return null;
	return ciclo === 'anio' ? precioAnual(p.precio) * 12 : p.precio;
};

const D = 86_400_000;
const diasEntre = (a: Date, b: Date) => (b > a ? Math.ceil((b.getTime() - a.getTime()) / D) : 0);
const mover = (iso: string | Date, ciclo: Ciclo, signo: 1 | -1) => {
	const d = new Date(iso);
	if (ciclo === 'anio') d.setUTCFullYear(d.getUTCFullYear() + signo);
	else d.setUTCMonth(d.getUTCMonth() + signo);
	return d;
};
export const siguiente = (iso: string | Date, ciclo: Ciclo) => mover(iso, ciclo, 1);

export function cotizar(e: EstadoAbono, plan: IdPlan, ciclo: Ciclo, ahora: Date): Cotizacion {
	const precio = precioDe(plan, ciclo);
	if (precio === null || (ciclo !== 'mes' && ciclo !== 'anio')) throw new ErrorCotizar('invalido');
	const ya = ahora.toISOString();
	const c: Cotizacion = {
		plan,
		ciclo,
		tipo: 'actual',
		hoy: 0,
		rige: ya,
		hasta: null,
		cobro: null,
		porCiclo: precio,
		lineas: [],
		pago: 'ninguno',
		inicio: null
	};
	const enPrueba = ahora < new Date(e.finPrueba);
	const pago = e.proximo !== null && new Date(e.proximo) > ahora;
	const actual = precioDe(e.plan, e.ciclo) ?? 0;
	const elegido = `${plan}/${ciclo}`;
	const vigente = `${e.plan}/${e.ciclo}`;

	if (e.pendiente) {
		if (elegido === `${e.pendiente.plan}/${e.pendiente.ciclo}`) {
			c.tipo = 'programado';
			c.rige = e.pendiente.desde;
			c.hasta = e.proximo;
			if (plan !== 'gratis') c.cobro = { fecha: e.pendiente.desde, importe: precio };
			return c;
		}
		if (elegido === vigente) {
			c.tipo = 'mantener';
			if (e.plan !== 'gratis' && pago) {
				c.cobro = { fecha: e.proximo!, importe: precio };
				if (!e.activo || e.pendiente.ciclo !== e.ciclo) {
					c.pago = 'suscripcion';
					c.inicio = e.proximo;
				} else c.pago = 'actualizar';
			}
			return c;
		}
		throw new ErrorCotizar('programado');
	}

	if (elegido === vigente || (plan === 'gratis' && e.plan === 'gratis')) {
		if (plan !== 'gratis' && pago) c.cobro = { fecha: e.proximo!, importe: precio };
		return c;
	}

	if (plan === 'gratis') {
		c.tipo = 'gratis';
		if (e.activo && pago) {
			c.rige = e.proximo!;
			c.hasta = e.proximo;
			c.pago = 'cancelar';
		} else if (e.activo) c.pago = 'cancelar';
		return c;
	}

	const linea = `${NOMBRES[plan]} ${ciclo === 'anio' ? 'anual' : 'mensual'}`;
	if (enPrueba) {
		c.tipo = 'prueba';
		c.cobro = { fecha: e.finPrueba, importe: precio };
		c.pago = e.activo && ciclo === e.ciclo ? 'actualizar' : 'suscripcion';
	} else if (!e.activo && pago && e.plan !== 'gratis') {
		c.tipo = plan === e.plan ? 'mantener' : 'ciclo';
		c.rige = e.proximo!;
		c.hasta = e.proximo;
		c.cobro = { fecha: e.proximo!, importe: precio };
		c.pago = 'suscripcion';
		c.inicio = e.proximo;
	} else if (!e.activo || !pago) {
		c.tipo = 'alta';
		c.hoy = precio;
		c.cobro = { fecha: siguiente(ahora, ciclo).toISOString(), importe: precio };
		c.lineas.push({ texto: linea, importe: precio });
		c.pago = 'suscripcion';
	} else if (ciclo !== e.ciclo) {
		c.tipo = 'ciclo';
		c.rige = e.proximo!;
		c.hasta = e.proximo;
		c.cobro = { fecha: e.proximo!, importe: precio };
		c.pago = 'suscripcion';
		c.inicio = e.proximo;
	} else if (precio > actual) {
		c.tipo = 'sube';
		const prox = new Date(e.proximo!);
		const dias = diasEntre(ahora, prox);
		const periodo = diasEntre(mover(prox, e.ciclo, -1), prox);
		const nuevo = Math.round((precio * dias) / periodo);
		const credito = Math.round((actual * dias) / periodo);
		c.hoy = nuevo - credito;
		c.cobro = { fecha: e.proximo!, importe: precio };
		const uno = dias === 1;
		c.lineas.push(
			{ texto: `${NOMBRES[plan]} por ${uno ? '1 día' : `${dias} días`}`, importe: nuevo },
			{
				texto: `${NOMBRES[e.plan]} ya pagado por ${uno ? 'ese día' : 'esos días'}`,
				importe: -credito
			}
		);
		c.pago = 'diferencia';
		if (c.hoy <= 0) {
			c.hoy = 0;
			c.pago = 'actualizar';
		}
	} else {
		c.tipo = 'baja';
		c.rige = e.proximo!;
		c.hasta = e.proximo;
		c.cobro = { fecha: e.proximo!, importe: precio };
		c.pago = 'actualizar';
	}
	return c;
}

/** Todas las cotizaciones por "plan/ciclo" (las de un cambio no permitido no van). */
export function todas(e: EstadoAbono, ahora: Date): Record<string, Cotizacion> {
	const r: Record<string, Cotizacion> = {};
	for (const p of PLANES) {
		for (const ci of ['mes', 'anio'] as const) {
			try {
				r[`${p.id}/${ci}`] = cotizar(e, p.id, ci, ahora);
			} catch {
				/* un cambio no permitido no se cotiza */
			}
		}
	}
	return r;
}
