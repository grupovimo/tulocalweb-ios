import type { FullAutoFill } from 'svelte/elements';
import type { TiendaPublica } from './tipos.ts';

/**
 * Campos de factura A del checkout (nombre, etiqueta, tipo, autocomplete), con la misma forma que los
 * del comprador de cada plantilla. Sólo si la tienda la emite; los dos son opcionales.
 */
export const camposFactura = (
	tienda: Pick<TiendaPublica, 'cobros'>
): [string, string, string, FullAutoFill][] =>
	tienda.cobros.facturaA
		? [
				['cuit', 'CUIT (si necesitás factura A)', 'text', 'off'],
				['razonSocial', 'Razón social (factura A)', 'text', 'organization']
			]
		: [];

/** Los campos del comprador que se pueden dejar vacíos. */
export const OPCIONALES = new Set(['dni', 'cuit', 'razonSocial']);
