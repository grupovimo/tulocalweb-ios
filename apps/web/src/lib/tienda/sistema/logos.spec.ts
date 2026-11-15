import { describe, expect, it } from 'vitest';
import { luz, tonoLogo } from './logos.ts';

/** Un logo de 10×10: la mitad transparente y la otra mitad de un color. */
const logo = (r: number, g: number, b: number, transparente = true) => {
	const px = new Uint8ClampedArray(100 * 4);
	for (let i = 0; i < 100; i++) {
		const lleno = !transparente || i < 50;
		px.set(lleno ? [r, g, b, 255] : [0, 0, 0, 0], i * 4);
	}
	return px;
};
const OSCURO = luz(14, 14, 14);
const CLARO = luz(246, 243, 238);

describe('logos que se leen', () => {
	it('uno negro sobre fondo oscuro pasa a claro', () => {
		expect(tonoLogo(logo(20, 20, 20), OSCURO)).toBe('claro');
	});
	it('uno blanco sobre fondo claro pasa a oscuro', () => {
		expect(tonoLogo(logo(255, 255, 255), CLARO)).toBe('oscuro');
	});
	it('los que contrastan quedan con sus colores', () => {
		expect(tonoLogo(logo(29, 185, 84), OSCURO)).toBeNull();
		expect(tonoLogo(logo(20, 20, 20), CLARO)).toBeNull();
		expect(tonoLogo(logo(255, 255, 255), OSCURO)).toBeNull();
	});
	it('los que traen su propio fondo no se tocan', () => {
		expect(tonoLogo(logo(20, 20, 20, false), OSCURO)).toBeNull();
	});
});
