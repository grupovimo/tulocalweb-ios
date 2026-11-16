import { describe, expect, it } from 'vitest';
import type { EnviosTienda } from '../tipos.ts';
import { confianzaDe, envioAuto, fechaLlegada } from './confianza.ts';
import { videoCortoDe } from './datos.ts';
import { VIDEO_CORTO } from './catalogo.ts';
import { diasDe } from '../envio.ts';

const envios: EnviosTienda = {
	gratisDesde: 60_000,
	retiro: true,
	retiroDireccion: 'Gorriti 4500',
	retiroHorario: '',
	moto: true,
	motoCosto: 3500,
	motoZona: 'Palermo y Belgrano',
	motoDemora: 'En el día',
	correo: true
};

describe('caja de confianza', () => {
	it('el envío sale de la configuración real', () => {
		expect(envioAuto(envios)).toEqual([
			'Envío gratis desde $ 60.000',
			'En moto a Palermo y Belgrano, en el día',
			'A todo el país por Correo Argentino',
			'Retiro sin cargo en Gorriti 4500'
		]);
	});
	it('garantía y cambios sólo si la tienda los escribió', () => {
		const sin = confianzaDe({ envios }, {});
		expect(sin.map((x) => x.tipo)).toEqual(['envio', 'protegida']);
		const con = confianzaDe({ envios }, { envio: 'Despachamos en 24 h', garantia: '6 meses' });
		expect(con.map((x) => x.tipo)).toEqual(['envio', 'garantia', 'protegida']);
		expect(con[0].lineas).toEqual(['Despachamos en 24 h']);
	});
	it('sin envíos ni retiro no promete envío', () => {
		const nada = { ...envios, gratisDesde: null, retiro: false, moto: false, correo: false };
		expect(confianzaDe({ envios: nada }, null).map((x) => x.tipo)).toEqual(['protegida']);
	});
});

describe('fecha estimada de llegada', () => {
	// Lunes 16/11/2026 a las 10 de Argentina.
	const lunes = new Date('2026-11-16T13:00:00Z');
	it('cuenta días hábiles desde hoy', () => {
		expect(fechaLlegada({ min: 2, max: 5 }, lunes)).toBe('Llega entre el mié 18/11 y el lun 23/11');
		expect(fechaLlegada({ min: 1, max: 1 }, lunes)).toBe('Llega el mar 17/11');
	});
	it('los días de Correo como números', () => {
		expect(diasDe({ precio: 1, producto: 'x', diasMin: '2', diasMax: '5' })).toEqual({
			min: 2,
			max: 5
		});
		expect(diasDe({ precio: 1, producto: 'x' })).toBeUndefined();
		expect(diasDe({ precio: 1, producto: 'x', diasMax: '3' })).toEqual({ min: 3, max: 3 });
	});
});

describe('videos verticales', () => {
	it('reconoce archivos y redes', () => {
		for (const u of [
			'https://fotos.tulocalweb.com.ar/v/a.mp4',
			'https://cdn.x.com/a.webm?v=2',
			'https://www.instagram.com/reel/C9xYz12AbCd/',
			'https://www.tiktok.com/@bylula/video/7301234567890123456',
			'https://youtube.com/shorts/abcDEF123'
		]) {
			expect(VIDEO_CORTO.test(u), u).toBe(true);
			expect(videoCortoDe(u), u).not.toBeNull();
		}
		for (const u of [
			'http://x.com/a.mp4',
			'https://www.youtube.com/watch?v=abc123',
			'https://x.com/a.gif',
			'https://instagram.com/bylula'
		]) {
			expect(VIDEO_CORTO.test(u), u).toBe(false);
			expect(videoCortoDe(u), u).toBeNull();
		}
		expect(videoCortoDe('https://www.instagram.com/p/C9xYz12AbCd/')).toEqual({
			tipo: 'red',
			red: 'instagram',
			src: 'https://www.instagram.com/p/C9xYz12AbCd/embed/'
		});
	});
});
