import { describe, expect, it } from 'vitest';
import {
	hechosDe,
	imagenProducto,
	imagenTienda,
	lista,
	recortar,
	tarjetaDe,
	textosCompartir,
	textosPorDefecto,
	textosProducto,
	versionDe,
	type Hechos
} from './compartir.ts';
import type { ProductoPublico, TiendaPublica } from './tipos.ts';

const hechos = (h: Partial<Hechos> = {}): Hechos => ({
	nombre: 'La tienda de Nahuel',
	titulo: 'La tienda de Nahuel',
	lema: '',
	ciudad: '',
	categorias: [],
	productos: 0,
	servicios: [],
	envios: { correo: false, moto: false, motoZona: '', retiro: false },
	mercadoPago: false,
	...h
});

const producto = (p: Partial<ProductoPublico> = {}): ProductoPublico => ({
	id: 'p1',
	slug: 'medialunas',
	nombre: 'Medialunas',
	descripcion: '',
	categoria: 'Panes',
	precio: 7900,
	anterior: null,
	foto: null,
	etiqueta: null,
	stock: 4,
	fotos: [],
	opciones: [],
	variantes: [],
	...p
});

const tienda = (t: Partial<TiendaPublica> = {}): TiendaPublica => ({
	slug: 'la-tienda-de-nahuel',
	nombre: 'La tienda de Nahuel',
	plantilla: 'miga',
	paleta: 0,
	secciones: [],
	portada: { titulo: 'La tienda de Nahuel', lema: '', foto: null },
	logo: '',
	compartir: { titulo: '', descripcion: '', foto: '' },
	ciudad: '',
	direccion: '',
	whatsapp: '',
	categorias: [],
	productos: [],
	masVendidos: [],
	resenas: [],
	promedio: null,
	preguntas: [],
	fotos: [],
	cobros: { transferencia: 0, cuotas: 0, mercadoPago: false, datosTransferencia: null },
	envios: {
		gratisDesde: null,
		retiro: false,
		retiroDireccion: '',
		retiroHorario: '',
		moto: false,
		motoCosto: 0,
		motoZona: '',
		motoDemora: '',
		correo: false
	},
	turnos: null,
	suscripciones: [],
	vitrina: false,
	dominio: null,
	medicion: { ga4: null, pixel: null },
	...t
});

describe('textosPorDefecto', () => {
	it('sin lema ni fotos no repite el nombre: suma lo que vende y lo cuenta abajo', () => {
		const t = textosPorDefecto(
			hechos({ categorias: ['Panes', 'Tortas', 'Café'], productos: 24, ciudad: 'Córdoba' })
		);
		expect(t.titulo).toBe('La tienda de Nahuel · Panes y Tortas');
		expect(t.descripcion).toBe('24 productos para comprar online.');
		expect(t.descripcion).not.toContain(t.titulo);
	});
	it('con lema: nombre y lema arriba, los datos abajo', () => {
		const t = textosPorDefecto(
			hechos({
				lema: 'Hecho a mano en Rosario',
				categorias: ['Velas'],
				productos: 1,
				envios: { correo: true, moto: false, motoZona: '', retiro: true },
				mercadoPago: true
			})
		);
		expect(t.titulo).toBe('La tienda de Nahuel · Hecho a mano en Rosario');
		expect(t.descripcion).toBe(
			'Velas: 1 producto para comprar online. Envíos a todo el país. Pagás con Mercado Pago.'
		);
	});
	it('una tienda vacía no inventa nada', () => {
		const t = textosPorDefecto(hechos());
		expect(t.titulo).toBe('La tienda de Nahuel · Tienda online');
		expect(t.descripcion).toBe('Conocé La tienda de Nahuel y escribinos.');
		expect(textosPorDefecto(hechos({ ciudad: 'Salta' })).descripcion).toBe(
			'Tienda online en Salta.'
		);
	});
	it('sólo turnos', () => {
		const t = textosPorDefecto(hechos({ servicios: ['Corte', 'Color'] }));
		expect(t.titulo).toBe('La tienda de Nahuel · Turnos online');
		expect(t.descripcion).toBe('Reservá tu turno: Corte y Color.');
	});
	it('un lema que es el nombre no cuenta, y un nombre largo va solo', () => {
		expect(textosPorDefecto(hechos({ lema: 'la tienda de nahuel' })).titulo).toBe(
			'La tienda de Nahuel · Tienda online'
		);
		const largo = 'Almacén de ramos generales y dietética natural del barrio de Palermo Soho';
		const t = textosPorDefecto(hechos({ nombre: largo, lema: 'Todo suelto' }));
		expect(t.titulo.length).toBeLessThanOrEqual(70);
		expect(t.descripcion).toContain('Todo suelto');
	});
	it('la descripción entra en 160 letras', () => {
		const t = textosPorDefecto(
			hechos({
				lema: 'x'.repeat(60),
				nombre: 'n'.repeat(30),
				categorias: ['Categoría uno larga', 'Categoría dos larga', 'Categoría tres larga'],
				productos: 300,
				envios: { correo: true, moto: false, motoZona: '', retiro: false },
				mercadoPago: true
			})
		);
		expect(t.descripcion.length).toBeLessThanOrEqual(160);
	});
});

describe('textosCompartir y textosProducto', () => {
	it('lo escrito por la tienda pisa lo sugerido', () => {
		const t = tienda({ compartir: { titulo: ' Mío ', descripcion: '', foto: '' } });
		expect(textosCompartir(t).titulo).toBe('Mío');
		expect(textosCompartir(t).descripcion).toBe('Conocé La tienda de Nahuel y escribinos.');
	});
	it('el producto: precio, oferta y la descripción, sin repetir el título', () => {
		const t = tienda({ productos: [producto()] });
		expect(textosProducto(t, producto({ anterior: 9200 }))).toEqual({
			titulo: 'Medialunas · La tienda de Nahuel',
			descripcion: '$ 7.900 (antes $ 9.200). Disponible para comprar online.'
		});
		expect(textosProducto(t, producto({ descripcion: 'De manteca.' })).descripcion).toBe(
			'$ 7.900. De manteca.'
		);
	});
});

describe('imagen', () => {
	it('junta portada y fotos sin repetir, y cuenta categorías reales', () => {
		const t = tienda({
			portada: { titulo: 'La tienda de Nahuel', lema: '', foto: 'a' },
			productos: [producto({ foto: 'a' }), producto({ id: 'p2', foto: 'b' })],
			categorias: [
				{ nombre: 'Tortas', cantidad: 1, foto: null },
				{ nombre: 'Panes', cantidad: 3, foto: null },
				{ nombre: 'Vacía', cantidad: 0, foto: null }
			]
		});
		const c = tarjetaDe(t);
		expect(c.fotos).toEqual(['a', 'b']);
		expect(c.categorias).toEqual([
			{ nombre: 'Panes', cantidad: 3 },
			{ nombre: 'Tortas', cantidad: 1 }
		]);
		expect(c.detalle).toBe('2 productos');
		expect(c.frase).toBe('Panes y Tortas');
		expect(c.direccion).toBe('la-tienda-de-nahuel.tulocalweb.com.ar');
		expect(tarjetaDe({ ...t, dominio: 'nahuel.com.ar' }).direccion).toBe('nahuel.com.ar');
	});
	it('la URL es absoluta y cambia cuando cambia lo que se dibuja', () => {
		const t = tienda();
		const a = imagenTienda(t);
		expect(a).toMatch(
			/^https:\/\/tulocalweb\.com\.ar\/t\/la-tienda-de-nahuel\/compartir\/[a-z0-9]+\.jpg$/
		);
		expect(imagenTienda(tienda())).toBe(a);
		expect(imagenTienda(tienda({ paleta: 2 }))).not.toBe(a);
		expect(imagenTienda(tienda({ logo: 'https://x.com/l.png' }))).not.toBe(a);
		expect(
			imagenTienda(
				tienda({ compartir: { titulo: '', descripcion: '', foto: 'https://x.com/f.jpg' } })
			)
		).not.toBe(a);
		// Los textos para compartir no cambian la imagen.
		expect(imagenTienda(tienda({ compartir: { titulo: 'Otro', descripcion: '', foto: '' } }))).toBe(
			a
		);
		expect(imagenProducto(t, producto())).toMatch(
			/\/t\/la-tienda-de-nahuel\/producto\/medialunas\/compartir\/[a-z0-9]+\.jpg$/
		);
		expect(versionDe(tarjetaDe(t))).toMatch(/^[a-z0-9]{1,7}$/);
	});
	it('hechosDe ordena las categorías por cantidad', () => {
		const h = hechosDe(
			tienda({
				categorias: [
					{ nombre: 'B', cantidad: 1, foto: null },
					{ nombre: 'A', cantidad: 5, foto: null }
				]
			})
		);
		expect(h.categorias).toEqual(['A', 'B']);
	});
});

describe('lista y recortar', () => {
	it('lista', () => {
		expect(lista([])).toBe('');
		expect(lista(['a'])).toBe('a');
		expect(lista(['a', 'b', 'c'])).toBe('a, b y c');
	});
	it('recortar corta en una palabra entera', () => {
		expect(recortar('Hola mundo', 20)).toBe('Hola mundo');
		expect(recortar('Medialunas de manteca por docena', 20)).toBe('Medialunas de…');
	});
});
