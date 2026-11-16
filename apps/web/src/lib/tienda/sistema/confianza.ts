/**
 * La caja de confianza de la ficha (paginas.producto.confianza, tienda-nivel-tiendanube.md) y la
 * fecha estimada de llegada al cotizar. El envío se arma solo con la configuración real de envíos
 * (moto con su zona y demora, correo, retiro, envío gratis desde $ X) salvo que la tienda haya
 * escrito el suyo; garantía y cambios salen sólo si la tienda los escribió. La compra protegida
 * (10 días para arrepentirse, Ley 24.240) vale para toda tienda: siempre va.
 */
import type { EnviosTienda, TiendaPublica } from '../tipos.ts';
import type { Confianza } from './catalogo.ts';
import { precio } from './formato.ts';

export type LineaConfianza = {
	tipo: 'envio' | 'garantia' | 'cambios' | 'protegida';
	titulo: string;
	lineas: string[];
};

const limpio = (x: unknown, max = 120) =>
	typeof x === 'string'
		? x
				.replace(/[<>\r\n]/g, ' ')
				.trim()
				.slice(0, max)
		: '';

/** Las líneas de envío que salen de la configuración (vacío: no envía ni tiene retiro). */
export function envioAuto(e: EnviosTienda): string[] {
	const out: string[] = [];
	if (e.gratisDesde && e.gratisDesde > 0) out.push(`Envío gratis desde ${precio(e.gratisDesde)}`);
	if (e.moto) {
		const zona = e.motoZona.trim();
		const demora = e.motoDemora.trim();
		out.push(
			[zona ? `En moto a ${zona}` : 'En moto', demora ? demora.toLowerCase() : '']
				.filter(Boolean)
				.join(', ')
		);
	}
	if (e.correo) out.push('A todo el país por Correo Argentino');
	if (e.retiro) {
		const donde = e.retiroDireccion.trim();
		out.push(donde ? `Retiro sin cargo en ${donde}` : 'Retiro sin cargo en el local');
	}
	return out;
}

export function confianzaDe(
	t: Pick<TiendaPublica, 'envios'>,
	c: Confianza | undefined | null
): LineaConfianza[] {
	const out: LineaConfianza[] = [];
	const envio = limpio(c?.envio);
	const auto = envioAuto(t.envios);
	if (envio || auto.length)
		out.push({ tipo: 'envio', titulo: 'Envíos', lineas: envio ? [envio] : auto });
	const garantia = limpio(c?.garantia);
	if (garantia) out.push({ tipo: 'garantia', titulo: 'Garantía', lineas: [garantia] });
	const cambios = limpio(c?.cambios);
	if (cambios) out.push({ tipo: 'cambios', titulo: 'Cambios', lineas: [cambios] });
	out.push({
		tipo: 'protegida',
		titulo: 'Compra protegida',
		lineas: ['10 días para arrepentirte de la compra']
	});
	return out;
}

const DIAS = ['dom', 'lun', 'mar', 'mié', 'jue', 'vie', 'sáb'];
const AR = -3 * 60 * 60 * 1000;

/** Suma días hábiles (lunes a viernes) a una fecha de Argentina (sin feriados). */
function habiles(desde: Date, n: number): Date {
	const d = new Date(desde.getTime());
	let quedan = n;
	while (quedan > 0) {
		d.setUTCDate(d.getUTCDate() + 1);
		const dia = d.getUTCDay();
		if (dia !== 0 && dia !== 6) quedan--;
	}
	return d;
}

const corta = (d: Date) => `${DIAS[d.getUTCDay()]} ${d.getUTCDate()}/${d.getUTCMonth() + 1}`;

/**
 * «Llega entre el jue 19/11 y el lun 23/11» (o «Llega el jue 19/11»), contando días hábiles desde
 * hoy en hora de Argentina. Es una estimación de quien lleva: no cuenta feriados.
 */
export function fechaLlegada(dias: { min: number; max: number }, hoy = new Date()): string {
	if (!(dias.min > 0) || !(dias.max >= dias.min)) return '';
	// La fecha de Argentina como UTC (así getUTC* es el día de acá).
	const ar = new Date(hoy.getTime() + AR);
	const a = habiles(ar, dias.min);
	const b = habiles(ar, dias.max);
	return dias.min === dias.max
		? `Llega el ${corta(a)}`
		: `Llega entre el ${corta(a)} y el ${corta(b)}`;
}
