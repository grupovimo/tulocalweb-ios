/**
 * Carritos abandonados de la tienda online: el checkout guarda el carrito cuando quien compra deja
 * su email, y el link del recordatorio (/t/{tienda}/carrito/recuperar/{token}) lo vuelve a armar.
 * Con la API real lo hace Go (con su tope por IP); en la demo, el sandbox del visitante.
 */
import { env } from 'cloudflare:workers';
import { ClienteBff, ErrorApi } from '../bff.ts';
import { dbSandbox } from '../sandbox/sandbox.ts';
import { golpe } from '../limite.ts';
import type { TiendaResuelta } from './publica.ts';

export type LineaGuardada = { id: string; varianteId: string | null; cantidad: number };

class CarritosApi extends ClienteBff {
	guardar(slug: string, email: string, nombre: string, carrito: LineaGuardada[]) {
		return this.pedir<void>('POST', `/v1/publico/tiendas/${slug}/carritos`, {
			email,
			nombre,
			carrito
		});
	}
	baja(slug: string, token: string) {
		return this.pedir<void>(
			'POST',
			`/v1/publico/tiendas/${slug}/carritos/${encodeURIComponent(token)}/baja`
		);
	}
	async porToken(slug: string, token: string) {
		try {
			const r = await this.pedir<{ carrito: LineaGuardada[] }>(
				'GET',
				`/v1/publico/tiendas/${slug}/carritos/${encodeURIComponent(token)}`
			);
			return r.carrito;
		} catch (e) {
			if (e instanceof ErrorApi && e.estado === 404) return null;
			throw e;
		}
	}
}

const api = (cliente: { ip: string; agente: string }) =>
	new CarritosApi(String(env.API_URL), String(env.CLAVE_BFF ?? ''), cliente);

/** Guarda el carrito (la vitrina de una plantilla no guarda nada). Devuelve false si se rechazó. */
export async function guardarCarrito(
	r: TiendaResuelta,
	cliente: { ip: string; agente: string },
	email: string,
	nombre: string,
	carrito: LineaGuardada[]
) {
	if (r.tienda.vitrina) return false;
	if (!r.datos || !r.sandboxId) {
		await api(cliente).guardar(r.tienda.slug, email, nombre, carrito);
		return true;
	}
	// En la demo, como mucho 30 por hora por visitante (cada uno escribe en la D1).
	if (!(await golpe(dbSandbox(env.SANDBOX), `carrito:${r.sandboxId}`, 30, 3600))) return false;
	return r.datos.guardarCarritoTienda(email, nombre, carrito);
}

/** Las líneas del link del recordatorio, o null si no existe o ya se compró. */
export async function carritoPorToken(
	r: TiendaResuelta,
	cliente: { ip: string; agente: string },
	token: string
) {
	if (!/^[A-Za-z0-9_-]{20,64}$/.test(token) || r.tienda.vitrina) return null;
	if (!r.datos || !r.sandboxId) return api(cliente).porToken(r.tienda.slug, token);
	return r.datos.carritoTienda(token);
}

/**
 * «No quiero más estos mails» del recordatorio (§62): corta los recordatorios de ese email. Siempre
 * responde lo mismo (no se filtra si el token existe).
 */
export async function bajaCarrito(
	r: TiendaResuelta,
	cliente: { ip: string; agente: string },
	token: string
) {
	if (!/^[A-Za-z0-9_-]{20,64}$/.test(token) || r.tienda.vitrina) return;
	if (!r.datos || !r.sandboxId) return api(cliente).baja(r.tienda.slug, token);
	await r.datos.bajaCarritoTienda(token);
}
