/**
 * API de conversiones de Meta desde el BFF (§47): el token de la tienda vive cifrado en Go y nunca
 * vuelve al navegador. La demo (sandbox) no la tiene: conectar responde que es una tienda de ejemplo.
 */
import { env } from 'cloudflare:workers';
import type { RequestEvent } from '@sveltejs/kit';
import type { components } from '@mostrador/contrato/esquema';
import { ClienteBff } from './bff.ts';
import { clienteDe } from './datos/index.ts';

export type Conversiones = components['schemas']['Conversiones'];

const RUTA = '/v1/panel/integraciones/medicion/conversiones';

export class ConversionesApi extends ClienteBff {
	estado(sesion: string) {
		return this.pedir<Conversiones>('GET', RUTA, undefined, sesion);
	}
	guardar(sesion: string, token: string, codigoPrueba: string) {
		return this.pedir<void>('PUT', RUTA, { token, codigoPrueba }, sesion);
	}
	desconectar(sesion: string) {
		return this.pedir<void>('DELETE', RUTA, undefined, sesion);
	}
}

export const conversionesDe = (event: Pick<RequestEvent, 'request' | 'getClientAddress'>) =>
	new ConversionesApi(String(env.API_URL), String(env.CLAVE_BFF ?? ''), clienteDe(event));
