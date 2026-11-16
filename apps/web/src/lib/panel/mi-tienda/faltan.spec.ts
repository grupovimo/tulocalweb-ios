import { describe, expect, it } from 'vitest';
import { queContradice, type Ofrece } from './faltan.ts';

const nada: Ofrece = {
	envioGratis: false,
	todoElPais: false,
	retiro: false,
	mercadoPago: false,
	cuotasSinInteres: false,
	transferencia: false
};
const todo = Object.fromEntries(Object.keys(nada).map((k) => [k, true])) as unknown as Ofrece;

describe('textos que prometen lo que la tienda no ofrece', () => {
	const beneficios = {
		items: [
			{ icono: 'envio', titulo: 'Envíos a todo el país', texto: 'Gratis desde $ 80.000' },
			{ icono: 'cuotas', titulo: '3 cuotas sin interés', texto: 'Con todas las tarjetas' }
		]
	};
	it('avisa lo primero que no se cumple', () => {
		expect(queContradice(beneficios, { ofrece: nada })).toBe(
			'Dice «envíos a todo el país», pero no tenés el envío por correo (Envíos y retiro).'
		);
		expect(queContradice(beneficios, { ofrece: { ...nada, todoElPais: true } })).toMatch(
			/cuotas sin interés/
		);
	});
	it('con la configuración que lo respalda, nada', () => {
		expect(queContradice(beneficios, { ofrece: todo })).toBeNull();
		expect(queContradice(['Envío gratis desde $ 50.000'], { ofrece: todo })).toBeNull();
	});
	it('envío gratis, retiro y transferencia', () => {
		expect(queContradice(['Envío gratis en CABA'], { ofrece: nada })).toMatch(/envío gratis/);
		expect(queContradice(['Retirá en el local'], { ofrece: nada })).toMatch(/retiro/);
		expect(queContradice(['10 % off pagando con transferencia'], { ofrece: nada })).toMatch(
			/transferencia/
		);
		expect(queContradice(['Hecho a mano en Córdoba'], { ofrece: nada })).toBeNull();
	});
	it('más cuotas sin interés que las elegidas en Cobros', () => {
		const seis = { ...todo, cuotasMax: 6 };
		expect(queContradice(['12 cuotas sin interés'], { ofrece: seis })).toBe(
			'Dice «12 cuotas sin interés», pero ofrecés hasta 6 (Configuración › Cobros).'
		);
		expect(queContradice(['3 y 6 cuotas sin interés'], { ofrece: seis })).toBeNull();
	});
	it('sin saber qué ofrece la tienda, no avisa', () => {
		expect(queContradice(beneficios, {})).toBeNull();
	});
});
