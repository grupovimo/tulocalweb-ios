import { describe, expect, test } from 'vitest';
import { BLOQUES, campo, TIPOS, validarBloque, validarBloques, type TipoBloque } from './catalogo.ts';

const b = (tipo: TipoBloque, datos: Record<string, unknown> = {}, variante?: string) => ({
	id: 'abc123',
	tipo,
	variante: variante ?? BLOQUES.find((x) => x.tipo === tipo)!.variantes[0],
	datos
});

describe('catálogo de bloques', () => {
	test('los 19 tipos de bloques.json', () => {
		expect(BLOQUES.map((x) => x.tipo).sort()).toEqual([...TIPOS].sort());
	});
	test('valida tipo, variante e id', () => {
		expect(validarBloque(b('banner', { titulo: 'Hot Sale' }))).not.toBeNull();
		expect(validarBloque({ ...b('banner'), variante: 'otra' })).toBeNull();
		expect(validarBloque({ ...b('banner'), tipo: 'carrusel' as TipoBloque })).toBeNull();
		expect(validarBloque({ ...b('banner'), id: 'A b' })).toBeNull();
	});
	test('como Go: sin campos desconocidos, sin < ni >, sin saltos en los textos cortos', () => {
		expect(validarBloque(b('banner', { titulo: 'Hola', otro: 1 }))).toBeNull();
		expect(validarBloque(b('banner', { titulo: '<b>Hola</b>' }))).toBeNull();
		expect(validarBloque(b('banner', { titulo: 'Hola\nchau' }))).toBeNull();
		expect(validarBloque(b('textoImagen', { titulo: 'Hola', texto: 'Uno\n\nDos' }))).not.toBeNull();
	});
	test('botones con texto y link; links y videos válidos', () => {
		expect(validarBloque(b('banner', { titulo: 'x', boton: { t: 'Ver', a: 'catalogo' } }))).not.toBeNull();
		expect(validarBloque(b('banner', { titulo: 'x', boton: { t: '', a: 'catalogo' } }))).toBeNull();
		expect(validarBloque(b('banner', { titulo: 'x', boton: { t: 'Ver', a: '' } }))).toBeNull();
		expect(validarBloque(b('video', { url: 'https://youtu.be/abc123' }))).not.toBeNull();
		expect(validarBloque(b('video', { url: 'https://evil.com/youtube.com/x' }))).toBeNull();
		expect(validarBloque(b('destacado', { producto: 'prod_1' }))).toBeNull();
		expect(validarBloque(b('destacado', { producto: 'campera-sur' }))).not.toBeNull();
	});
	test('en las listas, los requeridos no pueden venir vacíos y hay tope', () => {
		expect(validarBloque(b('preguntas', { items: [{ p: '¿Envían?', r: 'Sí.' }] }))).not.toBeNull();
		expect(validarBloque(b('preguntas', { items: [{ p: '', r: 'Sí.' }] }))).toBeNull();
		const muchas = Array.from({ length: 13 }, () => ({ p: 'a', r: 'b' }));
		expect(validarBloque(b('preguntas', { items: muchas }))).toBeNull();
	});
	test('una sola portada, sin ids repetidos', () => {
		const l = validarBloques([
			{ ...b('portada'), id: 'portada' },
			{ ...b('portada'), id: 'portada2' },
			{ ...b('banner', { titulo: 'x' }), id: 'portada' },
			{ ...b('banner', { titulo: 'x' }), id: 'banner1' }
		]);
		expect(l.map((x) => x.id)).toEqual(['portada', 'banner1']);
	});
	test('los valores por defecto del catálogo', () => {
		expect(campo(b('productos'), 'titulo')).toBe('Lo más elegido');
		expect(campo(b('productos', { titulo: 'Mío' }), 'titulo')).toBe('Mío');
	});
});
