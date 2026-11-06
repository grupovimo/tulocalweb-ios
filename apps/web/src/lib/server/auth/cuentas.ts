/**
 * Cuentas contra la API Go (DATA_SOURCE=api). El navegador nunca ve el token de Go: lo guarda este
 * servidor en una cookie httpOnly.
 */
import type { components } from '@mostrador/contrato/esquema';
import { ClienteBff } from '../bff.ts';

export { ErrorApi as ErrorCuenta } from '../bff.ts';
export type Sesion = components['schemas']['Sesion'];
export type DatosSesion = components['schemas']['DatosSesion'];
export type Invitacion = components['schemas']['Invitacion'];

export class CuentasApi extends ClienteBff {
	registrar(a: components['schemas']['Alta']) {
		return this.pedir<Sesion>('POST', '/v1/auth/registro', a);
	}
	ingresar(email: string, clave: string, recordarme: boolean) {
		return this.pedir<Sesion>('POST', '/v1/auth/ingresar', { email, clave, recordarme });
	}
	sesion(token: string) {
		return this.pedir<DatosSesion>('GET', '/v1/auth/sesion', undefined, token);
	}
	salir(token: string) {
		return this.pedir<void>('POST', '/v1/auth/salir', undefined, token);
	}
	/** El link de contraseña nueva, por email o al WhatsApp de la cuenta. */
	recuperar(a: { email: string } | { telefono: string }) {
		return this.pedir<void>('POST', '/v1/auth/recuperar', a);
	}
	/** Qué ingresos están configurados en la API. */
	proveedores() {
		return this.pedir<{ oauth: string[]; clerk: boolean; whatsapp: boolean }>(
			'GET',
			'/v1/auth/proveedores'
		);
	}
	/** URL del proveedor a la que se manda a la persona (con state y PKCE guardados en Go). */
	iniciarOAuth(proveedor: string) {
		return this.pedir<{ url: string }>(
			'POST',
			`/v1/auth/oauth/${encodeURIComponent(proveedor)}/inicio`,
			{}
		);
	}
	volverOAuth(proveedor: string, codigo: string, estado: string) {
		return this.pedir<Sesion & { nueva: boolean }>(
			'POST',
			`/v1/auth/oauth/${encodeURIComponent(proveedor)}/vuelta`,
			{ codigo, estado }
		);
	}
	/**
	 * Entra con el token de sesión de Clerk (Google, Apple): Go lo verifica con la instancia. Con una
	 * invitación, la persona queda sumada a esa tienda (y una cuenta nueva no crea otra tienda).
	 */
	ingresarConClerk(token: string, invitacion?: string) {
		return this.pedir<Sesion & { nueva: boolean }>('POST', '/v1/auth/clerk', {
			token,
			...(invitacion ? { invitacion } : {})
		});
	}
	/** De qué tienda y con qué rol es la invitación (sin usarla). */
	verInvitacion(token: string) {
		return this.pedir<Invitacion>('POST', '/v1/auth/invitacion', { token });
	}
	/** Suma a la persona de la sesión a la tienda de la invitación. */
	aceptarInvitacion(token: string, sesion: string) {
		return this.pedir<{ tiendaId: string }>(
			'POST',
			'/v1/auth/invitacion/aceptar',
			{ token },
			sesion
		);
	}
	/** Crea la cuenta con el email de la invitación, sin tienda propia, y la suma a esa tienda. */
	registrarConInvitacion(token: string, nombre: string, clave: string) {
		return this.pedir<Sesion>('POST', '/v1/auth/invitacion/registro', { token, nombre, clave });
	}
	/** Completa un ingreso con verificación en dos pasos (el desafío y el código). */
	completarSegundoFactor(desafio: string, codigo: string) {
		return this.pedir<Sesion>('POST', '/v1/auth/segundo-factor', { desafio, codigo });
	}
	estadoSegundoFactor(sesion: string) {
		return this.pedir<components['schemas']['EstadoSegundoFactor']>(
			'GET',
			'/v1/auth/segundo-factor',
			undefined,
			sesion
		);
	}
	activarSegundoFactor(sesion: string) {
		return this.pedir<{ secreto: string; uri: string }>(
			'POST',
			'/v1/auth/segundo-factor/activar',
			{},
			sesion
		);
	}
	/** confirmar (devuelve los códigos de recuperación), codigos (nuevos) o desactivar. */
	conCodigoSegundoFactor(
		accion: 'confirmar' | 'codigos' | 'desactivar',
		codigo: string,
		sesion: string
	) {
		return this.pedir<components['schemas']['CodigosRecuperacion'] | undefined>(
			'POST',
			`/v1/auth/segundo-factor/${accion}`,
			{ codigo },
			sesion
		);
	}
	/** Link para entrar sin contraseña, al WhatsApp de la cuenta. */
	pedirEnlace(telefono: string) {
		return this.pedir<void>('POST', '/v1/auth/enlace', { telefono });
	}
	usarEnlace(token: string) {
		return this.pedir<Sesion>('POST', '/v1/auth/enlace/usar', { token });
	}
	/** Si el link de contraseña nueva sirve todavía (sin usarlo). */
	verificarRestablecer(token: string) {
		return this.pedir<void>('POST', '/v1/auth/restablecer/verificar', { token });
	}
	restablecer(token: string, clave: string) {
		return this.pedir<void>('POST', '/v1/auth/restablecer', { token, clave });
	}
}
