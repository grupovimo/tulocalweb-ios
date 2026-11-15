import { describe, expect, it } from 'vitest';
import { escribirFiltros, leerFiltros, vistaDe } from './filtros.ts';
import { esperaPago, etapaDePedido, siguienteDe, textoEstado } from './nombres.ts';

// Una compra de la tienda online con Mercado Pago sin acreditar no es una venta (§63).
describe('esperando pago', () => {
	const mp = { estado: 'Sin pagar' as const, envio: 'correo', pago: 'mp', canal: 'web' };

	it('se ve «Esperando pago» y no se marca pagado a mano', () => {
		expect(esperaPago(mp)).toBe(true);
		expect(textoEstado(mp)).toBe('Esperando pago');
		expect(etapaDePedido(mp).t).toBe('Esperando pago');
		expect(siguienteDe(mp)).toBeUndefined();
	});

	it('la transferencia y lo cargado a mano siguen «Sin pagar»', () => {
		for (const p of [
			{ ...mp, pago: 'transf' },
			{ ...mp, canal: 'wa' }
		]) {
			expect(esperaPago(p)).toBe(false);
			expect(textoEstado(p)).toBe('Sin pagar');
			expect(siguienteDe(p)).toBe('Marcar como pagado');
		}
		expect(esperaPago({ ...mp, estado: 'Para preparar' })).toBe(false);
	});

	it('tiene su vista en Pedidos', () => {
		const f = leerFiltros(new URLSearchParams('estado=sin-pagar&pago=mp&canal=web'));
		expect(vistaDe(f)).toBe('esperando-pago');
		expect(escribirFiltros(f).toString()).toBe('estado=sin-pagar&canal=web&pago=mp');
	});
});
