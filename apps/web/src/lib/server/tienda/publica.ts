/**
 * Encuentra la tienda online de /t/[tienda]. En la demo: si quien mira tiene su sandbox y el slug es
 * el de su tienda, ve la suya (con lo que cambió y publicó); si no, ve la vitrina de la plantilla de
 * ese nombre (los datos de ejemplo, sólo para mirar). Con la API real, la pide a Go.
 */
import { env } from 'cloudflare:workers';
import RUBROS from '@mostrador/contrato/rubros.json';
import { PLANTILLAS } from '#lib/plantillas.ts';
import { validarGa4, validarPixel } from '#lib/tienda/medicion.ts';
import type { TiendaPublica } from '#lib/tienda/tipos.ts';
import { usaSandbox } from '../datos/index.ts';
import { slugDe } from '../sandbox/rubros.ts';
import { DatosSandbox, dbSandbox, leer, type Sandbox } from '../sandbox/sandbox.ts';

const DIA = 86_400_000;

/** El sandbox de mentira de una plantilla: arranca hoy (los datos de ejemplo quedan frescos). */
function vitrinaDe(slug: string): Sandbox | null {
	const p = PLANTILLAS.find((x) => slugDe(x.tienda) === slug);
	const r = p && RUBROS.find((x) => x.nombre === p.rubro);
	if (!p || !r) return null;
	const hoy = Math.floor(Date.now() / DIA) * DIA;
	return {
		id: `vitrina-${p.id}`,
		rubro: r.id,
		plantilla: p.id,
		creadoEn: hoy,
		venceEn: Number.MAX_SAFE_INTEGER
	};
}

/** Los IDs van a una URL de script: sólo pasan los que tienen el formato exacto. */
function medicionSegura(t: Partial<TiendaPublica>): TiendaPublica['medicion'] {
	const ga4 = validarGa4(t.medicion?.ga4);
	const pixel = validarPixel(t.medicion?.pixel);
	return {
		ga4: !t.vitrina && ga4.ok ? ga4.valor : null,
		pixel: !t.vitrina && pixel.ok ? pixel.valor : null
	};
}

export type TiendaResuelta = {
	tienda: TiendaPublica;
	datos: DatosSandbox | null;
	sandboxId: string | null;
};

export async function resolverTienda(
	slug: string,
	cookie: string | undefined
): Promise<TiendaResuelta | null> {
	if (!/^[a-z0-9-]{2,60}$/.test(slug)) return null;
	if (!usaSandbox()) {
		const r = await fetch(`${String(env.API_URL)}/v1/publico/tiendas/${slug}`);
		// Las vitrinas de las plantillas (tiendas de ejemplo) no están en la API: salen de sus datos.
		if (r.status === 404) return vitrina(slug);
		if (!r.ok) throw new Error(`API ${r.status}`);
		const tienda = (await r.json()) as TiendaPublica;
		// Una API sin variantes todavía: los productos se ven como sin opciones.
		tienda.productos = tienda.productos.map((p) => ({
			...p,
			fotos: p.fotos ?? (p.foto ? [p.foto] : []),
			opciones: p.opciones ?? [],
			variantes: p.variantes ?? []
		}));
		// Una API sin Envíos y retiro todavía: retiro sin dirección propia y sin moto.
		tienda.envios = {
			retiroDireccion: '',
			retiroHorario: '',
			motoCosto: 0,
			motoZona: '',
			motoDemora: '',
			...(tienda.envios as Partial<TiendaPublica['envios']>)
		} as TiendaPublica['envios'];
		// Sin alias ni CBU no se ofrece transferencia: tampoco su descuento (0 = no hay).
		if (!tienda.cobros.datosTransferencia) tienda.cobros.transferencia = 0;
		// Una API sin Al compartir todavía: lo sugerido.
		tienda.compartir = {
			titulo: tienda.compartir?.titulo ?? '',
			descripcion: tienda.compartir?.descripcion ?? '',
			foto: tienda.compartir?.foto ?? ''
		};
		return {
			tienda: {
				...tienda,
				turnos: tienda.turnos ?? null,
				// Una API sin suscripciones todavía: sin planes.
				suscripciones: tienda.suscripciones ?? [],
				medicion: medicionSegura(tienda)
			},
			datos: null,
			sandboxId: null
		};
	}
	const db = dbSandbox(env.SANDBOX);
	const propio = await leer(db, cookie);
	if (propio) {
		const datos = new DatosSandbox(db, propio);
		if ((await datos.tienda()).slug === slug)
			return { tienda: await datos.tiendaPublica(false), datos, sandboxId: propio.id };
	}
	return vitrina(slug);
}

/** La vitrina de una plantilla: sus datos de ejemplo, sólo para mirar (no se compra). */
async function vitrina(slug: string): Promise<TiendaResuelta | null> {
	const v = vitrinaDe(slug);
	if (!v) return null;
	const datos = new DatosSandbox(dbSandbox(env.SANDBOX), v);
	return { tienda: await datos.tiendaPublica(true), datos, sandboxId: null };
}
