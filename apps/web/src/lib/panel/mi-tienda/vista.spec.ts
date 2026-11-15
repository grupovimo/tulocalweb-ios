import { describe, expect, it, vi } from 'vitest';
import type { TiendaPublica } from '#lib/tienda/tipos.ts';
import type { DisenoV2 } from './diseno.ts';
import { avisar, conDiseno, leerMensaje, paginaVista } from './vista.ts';

const tienda = {
	slug: 'lumbre',
	nombre: 'Lumbre',
	plantilla: 'ceibo',
	paleta: 0,
	secciones: [],
	portada: { titulo: 'Lumbre', lema: 'Desde Rosario', foto: 'https://x/p.webp' },
	logo: '',
	compartir: { titulo: '', descripcion: '', foto: '' },
	estilo: {}
} as unknown as TiendaPublica;

describe('la tienda con el diseño del editor', () => {
	const d: DisenoV2 = {
		version: 2,
		paleta: 2,
		plantilla: 'pampa',
		logo: 'https://x/l.webp',
		estilo: {
			barra: { variante: 'fija', mensajes: [{ texto: 'Envío gratis', link: 'catalogo' }] }
		},
		inicio: [
			{ id: 'b1', tipo: 'banner', variante: 'banda', datos: { titulo: 'Hot Sale' } },
			{
				id: 'b2',
				tipo: 'portada',
				variante: 'foto',
				datos: { titulo: 'Hola', fotos: ['https://x/f.webp'] }
			},
			{ id: 'b3', tipo: 'resenas', variante: 'tarjetas', oculto: true, datos: {} }
		]
	};

	it('lleva el diseño entero y deriva lo de v1', () => {
		const t = conDiseno(tienda, d) as TiendaPublica & { diseno: DisenoV2 };
		expect(t.diseno).toBe(d);
		expect(t.plantilla).toBe('pampa');
		expect(t.paleta).toBe(2);
		expect(t.logo).toBe('https://x/l.webp');
		expect(t.secciones).toEqual(['hero', 'wa']);
		expect(t.portada).toEqual({ titulo: 'Hola', lema: 'Desde Rosario', foto: 'https://x/f.webp' });
		expect(t.estilo?.anuncio).toBe('Envío gratis');
	});

	it('sin WhatsApp flotante no va la sección wa', () => {
		const t = conDiseno(tienda, { ...d, estilo: { whatsappFlotante: false } });
		expect(t.secciones).toEqual(['hero']);
	});
});

describe('mensajes', () => {
	it('lee sólo los del protocolo', () => {
		expect(leerMensaje({ tipo: 'diseno', diseno: { version: 2 }, pagina: 'producto' })).toEqual({
			tipo: 'diseno',
			diseno: { version: 2 },
			pagina: 'producto'
		});
		expect(leerMensaje({ tipo: 'seleccionar', id: 'b1' })).toEqual({
			tipo: 'seleccionar',
			id: 'b1'
		});
		expect(leerMensaje({ tipo: 'otra' })).toBeNull();
		expect(leerMensaje('hola')).toBeNull();
		expect(paginaVista('cualquiera')).toBe('inicio');
	});

	it('avisa al editor web, a la app de iPhone y a la de Android', () => {
		const parent = { postMessage: vi.fn() };
		const ios = vi.fn();
		const android = vi.fn();
		const w = {
			parent,
			location: { origin: 'https://tulocalweb.com.ar' },
			webkit: { messageHandlers: { tienda: { postMessage: ios } } },
			TiendaPuente: { bloque: android }
		} as unknown as Window;
		avisar({ tipo: 'bloque', id: 'b1' }, w);
		expect(parent.postMessage).toHaveBeenCalledWith(
			{ tipo: 'bloque', id: 'b1' },
			'https://tulocalweb.com.ar'
		);
		expect(ios).toHaveBeenCalledWith({ tipo: 'bloque', id: 'b1' });
		expect(android).toHaveBeenCalledWith('b1');
	});
});
