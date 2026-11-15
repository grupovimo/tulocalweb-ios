import { describe, expect, test } from 'vitest';
import {
	BLOQUES,
	campo,
	opcionValida,
	TIPOS,
	validarOpcionesPagina,
	validarBloque,
	validarBloques,
	type TipoBloque
} from './catalogo.ts';

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
		expect(
			validarBloque(b('banner', { titulo: 'x', boton: { t: 'Ver', a: 'catalogo' } }))
		).not.toBeNull();
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

describe('opciones de la ficha: combos y guía de talles (como Go)', () => {
	const fila8 = ['1', '2', '3', '4', '5', '6', '7', '8'];
	const guia = (tabla: unknown) => opcionValida('producto', 'guiaTalles', { tabla });
	test('la tabla: de 2 a 8 filas, de 1 a 8 columnas, parejas, celdas cortas en una línea', () => {
		expect(
			guia([
				['Talle', 'Busto', 'Largo'],
				['S', '88', '70'],
				['M', '', '72']
			])
		).toBe(true);
		expect(guia(Array.from({ length: 8 }, () => fila8))).toBe(true);
		expect(guia([['Talle', 'Busto']])).toBe(false);
		expect(guia(Array.from({ length: 9 }, () => ['a']))).toBe(false);
		expect(
			guia([
				[...fila8, '9'],
				[...fila8, '9']
			])
		).toBe(false);
		expect(guia([['Talle', 'Busto'], ['S']])).toBe(false);
		expect(guia([['Talle'], ['x'.repeat(31)]])).toBe(false);
		expect(guia([['Talle'], ['<b>S</b>']])).toBe(false);
		expect(guia([['Talle'], ['S\nM']])).toBe(false);
	});
	test('foto https, sin campos de más; combos es sí o no', () => {
		expect(
			opcionValida('producto', 'guiaTalles', { foto: 'https://fotos.tulocalweb.com.ar/g.jpg' })
		).toBe(true);
		expect(opcionValida('producto', 'guiaTalles', { pdf: 'https://x.com/a.pdf' })).toBe(false);
		expect(opcionValida('producto', 'combos', false)).toBe(true);
		expect(opcionValida('producto', 'combos', 'no')).toBe(false);
	});
	test('lo que no vale se descarta y toma lo de por defecto', () => {
		expect(
			validarOpcionesPagina('producto', {
				variante: 'mosaico',
				combos: false,
				guiaTalles: { tabla: [['Talle']] },
				relacionados: 'si'
			})
		).toEqual({ combos: false });
	});
});
