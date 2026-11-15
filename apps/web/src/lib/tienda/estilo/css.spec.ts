import { describe, expect, it } from 'vitest';
import { colorDe, colorValido, esOscuro } from './css.ts';

describe('colores de Mi tienda', () => {
	it('sólo valida #RRGGBB (en mayúsculas)', () => {
		expect(colorValido(' #a1b2c3 ')).toBe('#A1B2C3');
		expect(colorValido('red')).toBeNull();
		expect(colorValido('#abc')).toBeNull();
		expect(colorValido('#123456;}body{')).toBeNull();
		expect(colorValido(null)).toBeNull();
	});

	it('el color de un rol: el elegido o el del punto de partida', () => {
		expect(colorDe('andes', { colores: { fondo: '#101820' } }, 'fondo')).toBe('#101820');
		expect(colorDe('andes', null, 'acento')).toBe('#2F6BFF');
		expect(colorDe('no-existe', null, 'fondo')).toBe(colorDe('litoral', null, 'fondo'));
	});

	it('distingue fondos oscuros', () => {
		expect(esOscuro('#0B0B0B')).toBe(true);
		expect(esOscuro('#F4EFEA')).toBe(false);
		expect(esOscuro(null)).toBe(false);
	});
});
