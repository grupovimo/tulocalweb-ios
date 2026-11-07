import { fail, redirect, type RequestEvent } from '@sveltejs/kit';
import { env } from 'cloudflare:workers';
import * as v from 'valibot';
import { fapiDe } from '#lib/clerk.ts';
import { cuentasDe, usaSandbox } from '#lib/server/datos/index.ts';
import { ErrorCuenta } from '#lib/server/auth/cuentas.ts';
import {
	Enlace,
	errores,
	Ingreso,
	Recuperar,
	Registro,
	Restablecer
} from '#lib/server/auth/esquemas.ts';
import { esProveedor, irAlProveedor } from '#lib/server/auth/social.ts';
import {
	borrarDesafio,
	guardarIngreso,
	guardarSesion,
	leerDesafio
} from '#lib/server/auth/sesion.ts';
import { destino, VOLVER, volverValido } from '#lib/server/auth/volver.ts';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async (event) => {
	const { url, cookies } = event;
	// El código de la verificación en dos pasos, sólo con un ingreso a medias.
	if (url.searchParams.get('m') === 'codigo' && !leerDesafio(cookies)) redirect(303, '/ingresar');
	const v = url.searchParams.get('volver');
	if (volverValido(v))
		cookies.set(VOLVER, v, { path: '/', httpOnly: true, sameSite: 'lax', maxAge: 600 });
	// El link de contraseña nueva se mira al abrirlo: si venció o ya se usó, se avisa de entrada en
	// vez de pedir la contraseña y recién ahí decirlo. Si la API no contesta, se muestra el
	// formulario (al enviarlo, Go vuelve a validar).
	let linkVencido = false;
	const token = url.searchParams.get('token');
	if (url.searchParams.get('m') === 'restablecer' && !usaSandbox()) {
		if (!token) linkVencido = true;
		else
			try {
				await cuentasDe(event).verificarRestablecer(token);
			} catch (e) {
				linkVencido = e instanceof ErrorCuenta && e.estado >= 400 && e.estado < 500;
			}
	}
	// Google y Apple van por Clerk (docs/decisiones.md §31): con la API, sólo si la instancia está
	// configurada; en la demo llevan al recorrido de ejemplo. WhatsApp, sólo en la demo.
	const clerk = usaSandbox() ? null : fapiDe(env.CLERK_PUBLISHABLE_KEY);
	return { clerk, sociales: usaSandbox() || !!clerk, whatsapp: usaSandbox(), linkVencido };
};

const datos = async (request: Request) => Object.fromEntries(await request.formData());

/** Un error de la API como errores del formulario: por campo si los trae, si no, general. */
function deApi(e: unknown) {
	if (!(e instanceof ErrorCuenta)) throw e;
	const errs = Object.keys(e.campos).length ? e.campos : { form: e.message };
	return { estado: e.estado >= 500 ? 503 : e.estado, errores: errs };
}

/*
 * En la demo (sandbox) no hay cuentas: entrar o registrarse lleva a la demo, y recuperar muestra la
 * pantalla de "Revisá tu email" sin mandar nada (docs/decisiones.md §12). Con la API Go
 * (DATA_SOURCE=api) autentican de verdad y la sesión queda en una cookie httpOnly.
 */
export const actions: Actions = {
	ingresar: async (event) => {
		const f = await datos(event.request);
		const r = v.safeParse(Ingreso, f);
		if (!r.success)
			return fail(400, { modo: 'login', email: String(f.email ?? ''), errores: errores(r.issues) });
		if (usaSandbox()) redirect(303, '/prueba');
		let hacia: string;
		try {
			const s = await cuentasDe(event).ingresar(
				r.output.email,
				r.output.clave,
				r.output.recordarme === 'on'
			);
			hacia = destino(event.cookies);
			if (guardarIngreso(event.cookies, s, hacia)) hacia = '/ingresar?m=codigo';
		} catch (e) {
			const x = deApi(e);
			return fail(x.estado, { modo: 'login', email: r.output.email, errores: x.errores });
		}
		redirect(303, hacia);
	},
	registro: async (event) => {
		const f = await datos(event.request);
		const r = v.safeParse(Registro, f);
		const previo = {
			modo: 'registro',
			nombre: String(f.nombre ?? ''),
			tienda: String(f.tienda ?? ''),
			email: String(f.email ?? ''),
			// Si el registro falla, "Acepto los términos" sigue como estaba.
			terminos: f.terminos === 'on'
		};
		if (!r.success) return fail(400, { ...previo, errores: errores(r.issues) });
		// El registro sigue en el onboarding (decisión: el onboarding es el registro). En la demo no
		// hay cuentas: se sigue igual, con el negocio de ejemplo.
		if (usaSandbox()) redirect(303, '/onboarding?paso=1');
		try {
			const { nombre, tienda, email, clave } = r.output;
			const s = await cuentasDe(event).registrar({ nombre, tienda, email, clave });
			guardarSesion(event.cookies, s.token, s.venceEn);
		} catch (e) {
			const x = deApi(e);
			return fail(x.estado, { ...previo, errores: x.errores });
		}
		redirect(303, '/onboarding?paso=1');
	},
	/**
	 * Sin JavaScript o en la demo. Google y Apple van por Clerk desde el navegador (en la demo, al
	 * recorrido de ejemplo); Mercado Pago y Mercado Libre, al proveedor; WhatsApp pide el número.
	 */
	social: async (event) => {
		const proveedor = String((await event.request.formData()).get('proveedor') ?? '');
		if (proveedor === 'whatsapp') return { modo: 'whatsapp' as const };
		if (proveedor === 'google' || proveedor === 'apple') {
			if (usaSandbox()) redirect(303, '/prueba');
			return fail(400, {
				modo: 'login',
				errores: { form: 'Para seguir con Google o Apple, activá JavaScript en tu navegador.' }
			});
		}
		if (!esProveedor(proveedor))
			return fail(400, { modo: 'login', errores: { form: 'Elegí cómo querés ingresar.' } });
		return irAlProveedor(event, proveedor, '/prueba', 'login');
	},
	/** Pide el link para entrar al WhatsApp de la cuenta. */
	enlace: async (event) => {
		const f = await datos(event.request);
		const r = v.safeParse(Enlace, f);
		if (!r.success) return fail(400, { modo: 'whatsapp', errores: errores(r.issues) });
		if (usaSandbox()) redirect(303, '/prueba');
		try {
			await cuentasDe(event).pedirEnlace(r.output.telefono);
		} catch (e) {
			const x = deApi(e);
			return fail(x.estado, { modo: 'whatsapp', errores: x.errores });
		}
		return {
			modo: 'enviado' as const,
			canal: 'whatsapp' as const,
			tipo: 'ingreso' as const,
			a: r.output.telefono
		};
	},
	/** El código de la app autenticadora (o uno de recuperación) que completa el ingreso. */
	codigo: async (event) => {
		const codigo = String((await event.request.formData()).get('codigo') ?? '').trim();
		const desafio = leerDesafio(event.cookies);
		if (!desafio)
			return fail(401, {
				modo: 'login',
				errores: { form: 'Se venció el ingreso. Ingresá de nuevo.' }
			});
		if (!codigo)
			return fail(400, { modo: 'codigo', errores: { codigo: 'Escribí el código de tu app.' } });
		try {
			const s = await cuentasDe(event).completarSegundoFactor(desafio.d, codigo);
			guardarSesion(event.cookies, s.token, s.venceEn);
			borrarDesafio(event.cookies);
		} catch (e) {
			const x = deApi(e);
			if (x.estado === 401) {
				borrarDesafio(event.cookies);
				return fail(401, { modo: 'login', errores: x.errores });
			}
			return fail(x.estado, {
				modo: 'codigo',
				errores:
					x.estado === 400
						? { codigo: 'El código no es correcto. Probá con el que muestra ahora.' }
						: x.errores
			});
		}
		redirect(303, desafio.h);
	},
	/** Entra con el link del WhatsApp (un clic: así la vista previa del link no lo consume). */
	entrar: async (event) => {
		const token = String((await event.request.formData()).get('token') ?? '');
		if (usaSandbox()) redirect(303, '/prueba');
		let hacia: string;
		try {
			const s = await cuentasDe(event).usarEnlace(token);
			hacia = destino(event.cookies);
			if (guardarIngreso(event.cookies, s, hacia)) hacia = '/ingresar?m=codigo';
		} catch (e) {
			const x = deApi(e);
			return fail(x.estado, { modo: 'enlace', errores: x.errores });
		}
		redirect(303, hacia);
	},
	recuperar: async (event: RequestEvent) => {
		const f = await datos(event.request);
		const r = v.safeParse(Recuperar, f);
		if (!r.success)
			return fail(400, {
				modo: 'recuperar',
				canal: String(f.canal ?? 'email'),
				errores: errores(r.issues)
			});
		if (!usaSandbox()) {
			try {
				await cuentasDe(event).recuperar(
					r.output.canal === 'email' ? { email: r.output.email } : { telefono: r.output.telefono }
				);
			} catch (e) {
				const x = deApi(e);
				return fail(x.estado, { modo: 'recuperar', canal: r.output.canal, errores: x.errores });
			}
		}
		// El destino vuelve en la respuesta, no en la URL: así no queda en el historial ni en logs.
		const destino = r.output.canal === 'email' ? r.output.email : r.output.telefono;
		return { modo: 'enviado' as const, canal: r.output.canal, a: destino };
	},
	restablecer: async (event) => {
		const f = await datos(event.request);
		const r = v.safeParse(Restablecer, f);
		if (!r.success) return fail(400, { modo: 'restablecer', errores: errores(r.issues) });
		// En la demo no hay cuentas ni links reales.
		if (usaSandbox()) redirect(303, '/prueba');
		try {
			await cuentasDe(event).restablecer(r.output.token, r.output.clave);
		} catch (e) {
			const x = deApi(e);
			return fail(x.estado, { modo: 'restablecer', errores: x.errores });
		}
		redirect(303, '/ingresar?m=login&listo=clave');
	}
};
