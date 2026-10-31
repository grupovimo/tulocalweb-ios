import { describe, expect, it } from 'vitest';
import { camposFactura, OPCIONALES } from './factura.ts';

const cobros = { transferencia: 0, cuotas: 0, mercadoPago: false, datosTransferencia: null };

describe('camposFactura', () => {
	it('sólo pide CUIT y razón social si la tienda emite factura A', () => {
		expect(camposFactura({ cobros })).toEqual([]);
		expect(camposFactura({ cobros: { ...cobros, facturaA: false } })).toEqual([]);
		expect(camposFactura({ cobros: { ...cobros, facturaA: true } }).map(([n]) => n)).toEqual([
			'cuit',
			'razonSocial'
		]);
	});
	it('los dos se pueden dejar vacíos', () => {
		expect(
			OPCIONALES.has('cuit') && OPCIONALES.has('razonSocial') && !OPCIONALES.has('email')
		).toBe(true);
	});
});
