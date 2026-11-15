import { describe, expect, it } from 'vitest';
import type { Cookies } from '@sveltejs/kit';
import { COOKIE_ORIGEN, leerOrigen, origenDe } from './origen.ts';

describe('origenDe', () => {
	const ahora = new Date('2026-10-23T15:00:00Z');
	it('toma utm, ref y cupón válidos', () => {
		expect(
			origenDe(new URL('https://x/empezar?utm_source=flyer&ref=BRUMA-REF&cupon=arranco50'), ahora)
		).toEqual({
			fuente: 'flyer',
			ref: 'BRUMA-REF',
			cupon: 'ARRANCO50',
			llego: ahora.toISOString()
		});
	});
	it('toma el anuncio (utm_content)', () => {
		expect(
			origenDe(
				new URL('https://x/l/avisos?utm_source=facebook&utm_campaign=avisos&utm_content=1203'),
				ahora
			)
		).toEqual({
			fuente: 'facebook',
			campana: 'avisos',
			contenido: '1203',
			llego: ahora.toISOString()
		});
	});
	it('descarta lo que no es un código', () => {
		expect(origenDe(new URL('https://x/empezar?ref=<script>&utm_source=a b'), ahora)).toBeNull();
	});
	it('sin nada no anota', () => {
		expect(origenDe(new URL('https://x/empezar'), ahora)).toBeNull();
	});
});

describe('leerOrigen', () => {
	const con = (v: string) =>
		({ get: (k: string) => (k === COOKIE_ORIGEN ? v : undefined) }) as Cookies;
	it('descarta cookies sin un dato de atribución válido', () => {
		for (const v of [
			'{}',
			'[]',
			'null',
			'"x"',
			'{"llego":"2026-10-23"}',
			'{"ref":"<script>"}',
			'no-json'
		])
			expect(leerOrigen(con(v)), v).toBeNull();
	});
	it('se queda sólo con los campos conocidos y válidos', () => {
		expect(
			leerOrigen(
				con('{"ref":"BRUMA-REF","otro":1,"medio":"a b","llego":"2026-10-23T08:00:00.000Z"}')
			)
		).toEqual({
			ref: 'BRUMA-REF',
			llego: '2026-10-23T08:00:00.000Z'
		});
	});
});
