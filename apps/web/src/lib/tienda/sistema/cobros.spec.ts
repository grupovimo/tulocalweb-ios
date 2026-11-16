import { describe, expect, it } from 'vitest';
import type { TiendaPublica } from '../tipos.ts';
import {
	ivaDe,
	mensajesAuto,
	ofreceDe,
	precioCentavos,
	sinImpuestos,
	textoCuotas,
	textoSinImpuestos,
	textoTotalSinImpuestos,
	textoTransferencia
} from './cobros.ts';

const tienda = (cobros: Partial<TiendaPublica['cobros']>, gratisDesde: number | null = null) =>
	({
		cobros: {
			transferencia: 0,
			cuotas: 0,
			mercadoPago: false,
			datosTransferencia: null,
			...cobros
		},
		envios: { gratisDesde }
	}) as unknown as TiendaPublica;

describe('lo que la tienda ofrece al cobrar', () => {
	it('sin los campos nuevos: nada se promete (cuotas 0, impuestos apagados)', () => {
		const o = ofreceDe(tienda({ cuotas: 6 }));
		expect(o.cuotas).toBe(0);
		expect(o.impuestos).toBe(false);
		expect(o.cuotasDesde).toBe(30_000);
		expect(textoCuotas(o, 90_000)).toBe('');
	});
	it('cuotas sólo con Mercado Pago y desde el monto', () => {
		const o = ofreceDe(tienda({ cuotas: 3, cuotasDesde: 20_000, mercadoPago: true }));
		expect(textoCuotas(o, 19_999)).toBe('');
		expect(textoCuotas(o, 36_000)).toBe('3 cuotas sin interés de $ 12.000');
		expect(textoCuotas(o, 36_000, 'corto')).toBe('3 x $ 12.000 sin interés');
	});
	it('transferencia sólo con los datos cargados', () => {
		expect(textoTransferencia(ofreceDe(tienda({ transferencia: 10 })), 1000)).toBe('');
		const o = ofreceDe(
			tienda({ transferencia: 10, datosTransferencia: { alias: 'a', cbu: '', titular: 'x' } })
		);
		expect(textoTransferencia(o, 1000)).toBe('$ 900 con transferencia');
	});
	it('precio sin impuestos nacionales con centavos y coma', () => {
		expect(sinImpuestos(18_500, 21)).toBe(15_289.26);
		expect(precioCentavos(15_289.26)).toBe('$ 15.289,26');
		expect(ivaDe({ iva: 10.5 })).toBe(10.5);
		expect(ivaDe({ iva: 13 })).toBe(21);
		expect(ivaDe({})).toBe(21);
		const o = ofreceDe(tienda({ impuestos: { mostrar: true } }));
		expect(textoSinImpuestos(o, 18_500)).toBe('Precio sin impuestos nacionales $ 15.289,26');
		expect(textoSinImpuestos(ofreceDe(tienda({})), 18_500)).toBe('');
		expect(
			textoTotalSinImpuestos(o, [
				{ subtotal: 12_100, iva: 21 },
				{ subtotal: 1_105, iva: 10.5 }
			])
		).toBe('Total sin impuestos nacionales $ 11.000,00');
	});
	it('los mensajes automáticos de la barra no repiten lo que ya dicen los propios', () => {
		const o = ofreceDe(
			tienda(
				{
					cuotas: 6,
					mercadoPago: true,
					transferencia: 15,
					datosTransferencia: { alias: 'a', cbu: '', titular: 'x' }
				},
				80_000
			)
		);
		expect(mensajesAuto(o)).toEqual([
			'6 cuotas sin interés desde $ 30.000',
			'Envío gratis desde $ 80.000',
			'15 % off pagando con transferencia'
		]);
		expect(mensajesAuto(o, ['ENVÍO GRATIS en CABA'])).toHaveLength(2);
		expect(mensajesAuto(ofreceDe(tienda({})))).toEqual([]);
	});
});
