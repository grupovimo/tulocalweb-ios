import { describe, expect, it } from 'vitest';
import {
	agregar,
	alternar,
	bloquesDe,
	borrar,
	conBloquesEn,
	conDato,
	crear,
	duplicar,
	idsDe,
	mover,
	nuevoId,
	porQueNo,
	type Bloque
} from './bloques.ts';
import type { TipoBloque } from '#lib/tienda/sistema/catalogo.ts';
import { DEFS, LIMITES } from './catalogo.ts';
import { paraGuardar } from './guardar.ts';
import { fechaLocal, isoDeLocal, buscar } from './campos.ts';
import { limpiar, type DisenoV2 } from './diseno.ts';
import { queFalta, type Contexto } from './faltan.ts';
import { deshacer, registrar, rehacer, vacio } from './historial.ts';
import { aplicarPreset, PRESETS_TIENDA } from './partida.ts';

const b = (
	id: string,
	tipo: TipoBloque = 'banner',
	datos: Record<string, unknown> = {}
): Bloque => ({
	id,
	tipo,
	variante: DEFS.find((d) => d.tipo === tipo)!.variantes[0],
	datos
});

describe('lista de bloques', () => {
	const lista = [b('aaaaaaaa'), b('bbbbbbbb'), b('cccccccc')];

	it('agrega entre medio y en los bordes', () => {
		const n = b('nnnnnnnn');
		expect(agregar(lista, n, 1).map((x) => x.id)).toEqual([
			'aaaaaaaa',
			'nnnnnnnn',
			'bbbbbbbb',
			'cccccccc'
		]);
		expect(agregar(lista, n, 0)[0].id).toBe('nnnnnnnn');
		expect(agregar(lista, n).at(-1)!.id).toBe('nnnnnnnn');
		expect(agregar(lista, n, 99).at(-1)!.id).toBe('nnnnnnnn');
	});

	it('mueve arriba y abajo sin perder ninguno', () => {
		expect(mover(lista, 0, 2).map((x) => x.id)).toEqual(['bbbbbbbb', 'cccccccc', 'aaaaaaaa']);
		expect(mover(lista, 2, 0).map((x) => x.id)).toEqual(['cccccccc', 'aaaaaaaa', 'bbbbbbbb']);
		expect(mover(lista, 1, 1)).toBe(lista);
	});

	it('duplica abajo con id nuevo y datos copiados (no compartidos)', () => {
		const l = [b('aaaaaaaa', 'banner', { titulo: 'Hot Sale' })];
		const d = duplicar(l, 0, idsDe({ inicio: l }));
		expect(d).toHaveLength(2);
		expect(d[1].id).not.toBe('aaaaaaaa');
		expect(d[1].datos).toEqual({ titulo: 'Hot Sale' });
		expect(d[1].datos).not.toBe(d[0].datos);
	});

	it('oculta, muestra y borra', () => {
		const o = alternar(lista, 1);
		expect(o[1].oculto).toBe(true);
		expect('oculto' in alternar(o, 1)[1]).toBe(false);
		expect(borrar(lista, 0).map((x) => x.id)).toEqual(['bbbbbbbb', 'cccccccc']);
	});

	it('respeta el máximo por tipo y por página', () => {
		expect(porQueNo([b('aaaaaaaa', 'portada')], 'portada')).toMatch(/Ya hay/);
		expect(porQueNo([], 'portada')).toBeNull();
		const llena = Array.from({ length: LIMITES.bloquesPorPagina }, (_, i) => b(`x${i}xxxxxx`));
		expect(porQueNo(llena, 'banner')).toMatch(/hasta/);
		expect(porQueNo([], 'nada')).toMatch(/no existe/);
	});

	it('ids de 8 letras que no se repiten', () => {
		const usados = new Set(['aaaaaaaa']);
		let n = 0;
		const azar = () => (n++ < 8 ? 0 : 0.5);
		const id = nuevoId(usados, azar);
		expect(id).toMatch(/^[a-z0-9]{8}$/);
		expect(id).not.toBe('aaaaaaaa');
		const c = crear('portada', usados, 'collage');
		expect(c.variante).toBe('collage');
		expect(crear('portada', usados, 'inventada').variante).toBe('foto');
	});

	it('cada página tiene su lista', () => {
		const d = { inicio: lista, paginas: {} };
		const conN = conBloquesEn(d, 'nosotros', [b('zzzzzzzz', 'textoImagen')]);
		expect(bloquesDe(conN, 'nosotros')).toHaveLength(1);
		expect(bloquesDe(conN, 'inicio')).toBe(lista);
		expect(conBloquesEn(d, 'carrito', [])).toBe(d);
		expect(idsDe(conN).has('zzzzzzzz')).toBe(true);
	});

	it('un dato vacío se saca para que tome lo sugerido', () => {
		const x = conDato(b('aaaaaaaa', 'banner', { titulo: 'A' }), 'titulo', '');
		expect(x.datos).toEqual({});
		expect(conDato(x, 'fotos', ['u']).datos).toEqual({ fotos: ['u'] });
	});
});

describe('historial', () => {
	it('deshace y rehace en orden; un cambio nuevo corta el rehacer', () => {
		let h = vacio<number>();
		h = registrar(h, 1);
		h = registrar(h, 2);
		const d1 = deshacer(h, 3)!;
		expect(d1.estado).toBe(2);
		const d2 = deshacer(d1.h, 2)!;
		expect(d2.estado).toBe(1);
		expect(deshacer(d2.h, 1)).toBeNull();
		const r = rehacer(d2.h, 1)!;
		expect(r.estado).toBe(2);
		expect(registrar(r.h, 2).futuro).toEqual([]);
	});

	it('tiene tope', () => {
		let h = vacio<number>();
		for (let i = 0; i < 100; i++) h = registrar(h, i, 10);
		expect(h.pasado).toEqual([90, 91, 92, 93, 94, 95, 96, 97, 98, 99]);
	});
});

describe('qué le falta a un bloque', () => {
	const cx: Contexto = {
		productos: [
			{ id: 'p1', nombre: 'Taza', foto: null, precio: 1000, categoria: 'Cocina', oferta: false }
		],
		categorias: ['Cocina'],
		resenas: 0,
		whatsapp: false,
		turnos: false,
		direccion: false,
		fotos: 0
	};

	it('pide los campos obligatorios', () => {
		expect(queFalta(b('a', 'banner'), cx)).toMatch(/título/);
		expect(queFalta(b('a', 'banner', { titulo: 'Hot Sale' }), cx)).toBeNull();
		expect(queFalta(b('a', 'preguntas'), cx)).toMatch(/preguntas/i);
	});

	it('avisa lo que la tienda no tiene', () => {
		expect(queFalta(b('a', 'resenas'), cx)).toMatch(/reseñas/);
		expect(queFalta(b('a', 'whatsapp'), cx)).toMatch(/WhatsApp/);
		expect(queFalta(b('a', 'turnos'), cx)).toMatch(/turnos/);
		expect(queFalta(b('a', 'galeria'), cx)).toMatch(/fotos/);
		expect(queFalta(b('a', 'productos'), cx)).toBeNull();
		expect(queFalta(b('a', 'productos'), { ...cx, productos: [] })).toMatch(/productos/);
		expect(
			queFalta(b('a', 'productos', { productos: { fuente: 'ofertas', cantidad: 4 } }), cx)
		).toMatch(/oferta/);
		expect(
			queFalta(b('a', 'productos', { productos: { fuente: 'elegidos', cantidad: 4, ids: [] } }), cx)
		).toMatch(/Elegí/);
	});

	it('la oferta vencida y el destacado que ya no está', () => {
		const ahora = Date.parse('2026-11-15T12:00:00-03:00');
		expect(queFalta(b('a', 'oferta', { hasta: '2026-11-14T23:59:00-03:00' }), cx, ahora)).toMatch(
			/terminó/
		);
		expect(queFalta(b('a', 'destacado', { producto: 'p9' }), cx)).toMatch(/otro/);
		expect(queFalta(b('a', 'destacado', { producto: 'p1' }), cx)).toBeNull();
		expect(queFalta(b('a', 'combo', { productos: ['p1'] }), cx)).toMatch(/2/);
	});
});

describe('punto de partida', () => {
	const d: DisenoV2 = {
		version: 2,
		paleta: 2,
		plantilla: 'ceibo',
		estilo: { forma: 'pildora', colores: { acento: '#112233' }, logoOscuro: 'https://x/l.webp' },
		inicio: [
			{ id: 'b1', tipo: 'productos', variante: 'lista', datos: { titulo: 'Lo nuestro' } },
			{ id: 'b2', tipo: 'banner', variante: 'banda', oculto: true, datos: { titulo: 'Hot Sale' } }
		]
	};

	it('arma el orden del preset y conserva los datos del mismo tipo', () => {
		const r = aplicarPreset(d, 'mostaza');
		const p = PRESETS_TIENDA.mostaza;
		expect(r.plantilla).toBe('mostaza');
		// El orden del preset, con sus agregados (novedades, ofertas) después de los productos.
		const sinAgregados = r.inicio.filter((x) => !p.agregados?.some((a) => a.id === x.id));
		expect(sinAgregados.slice(0, p.orden.length).map((x) => x.tipo)).toEqual(p.orden);
		expect(r.inicio.map((x) => x.id).slice(3, 5)).toEqual(['novedades', 'ofertas']);
		expect(r.inicio.find((x) => x.id === 'novedades')!.datos).toMatchObject({
			productos: { fuente: 'nuevos' }
		});
		const prods = r.inicio.find((x) => x.tipo === 'productos')!;
		expect(prods.id).toBe('b1');
		expect(prods.datos).toEqual({ titulo: 'Lo nuestro' });
		expect(prods.variante).toBe(p.variantes.productos);
		// El banner no está en el preset: sigue, al final.
		expect(r.inicio.at(-1)!.id).toBe('b2');
		expect(new Set(r.inicio.map((x) => x.id)).size).toBe(r.inicio.length);
	});

	it('el look pasa a ser el del preset; la marca queda', () => {
		const r = aplicarPreset(d, 'pampa');
		expect(r.paleta).toBe(0);
		expect(r.estilo).toEqual({ logoOscuro: 'https://x/l.webp' });
		expect(aplicarPreset(d, 'inventado')).toBe(d);
	});

	it('los 17 existen y sus bloques están en el catálogo', () => {
		expect(Object.keys(PRESETS_TIENDA)).toHaveLength(17);
		for (const p of Object.values(PRESETS_TIENDA))
			for (const t of p.orden) expect(DEFS.some((x) => x.tipo === t)).toBe(true);
	});
});

describe('limpiar', () => {
	it('saca lo vacío pero deja los datos de cada bloque', () => {
		expect(
			limpiar({ a: '', b: {}, c: { d: '' }, inicio: [{ id: 'x', datos: {} }], e: 0, f: false })
		).toEqual({ inicio: [{ id: 'x', datos: {} }], e: 0, f: false });
	});
});

describe('para guardar', () => {
	const base: DisenoV2 = { version: 2, paleta: 0, inicio: [] };

	it('saca lo incompleto y lo que no es del catálogo', () => {
		const r = paraGuardar({
			...base,
			inicio: [
				b('aaaaaaaa', 'preguntas', {
					titulo: 'Dudas <b>',
					items: [{ p: '¿Envían?', r: 'Sí' }, { p: 'Sin respuesta' }],
					inventado: 1
				}),
				b('bbbbbbbb', 'portada', {
					titulo: 'Hola\nmundo',
					boton: { t: 'Ver', a: '' },
					boton2: { t: '', a: 'catalogo' },
					fotos: ['https://x.com/a.webp', 'javascript:alert(1)']
				}),
				b('cccccccc', 'logos', { items: [] }),
				b('dddddddd', 'productos', { productos: { fuente: 'elegidos', ids: ['p1', 'p 2'] } }),
				b('eeeeeeee', 'combo', { productos: ['p1'] }),
				b('ffffffff', 'banner', { boton: { t: 'Ver', a: 'prod:p1' }, segundo: { titulo: '' } })
			]
		});
		expect(r.inicio[0].datos).toEqual({ titulo: 'Dudas b', items: [{ p: '¿Envían?', r: 'Sí' }] });
		expect(r.inicio[1].datos).toEqual({ titulo: 'Hola mundo', fotos: ['https://x.com/a.webp'] });
		expect(r.inicio[2].datos).toEqual({});
		expect(r.inicio[3].datos).toEqual({
			productos: { fuente: 'elegidos', cantidad: 8, ids: ['p1'] }
		});
		expect(r.inicio[4].datos).toEqual({});
		expect(r.inicio[5].datos).toEqual({ boton: { t: 'Ver', a: 'prod:p1' } });
	});

	it('siempre v2, sin lo de v1, con la barra y el menú limpios', () => {
		const r = paraGuardar({
			...base,
			lema: '  ',
			estilo: {
				anuncio: 'viejo',
				textos: { prods: 'x' },
				barra: { variante: 'fija', mensajes: [{ texto: ' ' }] },
				menu: [
					{ t: 'Ofertas', a: 'cat:Ofertas' },
					{ t: 'Mal', a: 'https://' }
				],
				redes: { instagram: '' }
			}
		});
		expect(r).toEqual({
			version: 2,
			paleta: 0,
			inicio: [],
			estilo: { menu: [{ t: 'Ofertas', a: 'cat:Ofertas' }] }
		});
	});

	it('las páginas sólo con lo del catálogo', () => {
		const r = paraGuardar({
			...base,
			paginas: {
				catalogo: { variante: 'lista', filtros: 'arriba', columnas: 9 },
				nosotros: { bloques: [b('aaaaaaaa', 'textoImagen', { titulo: 'Somos' })] }
			}
		});
		expect(r.paginas).toEqual({
			catalogo: { variante: 'lista', filtros: 'arriba', columnas: 5 },
			nosotros: { bloques: [b('aaaaaaaa', 'textoImagen', { titulo: 'Somos' })] }
		});
	});

	it('la guía de talles: la tabla pareja, sin filas vacías; sin cuerpo no va', () => {
		const r = paraGuardar({
			...base,
			paginas: {
				producto: {
					variante: 'grilla',
					relacionados: true,
					calcularEnvio: true,
					combos: false,
					guiaTalles: {
						tabla: [['Talle', ' Busto <cm> '], ['S', '88', 'de más'], ['', ''], ['M']]
					},
					bloques: []
				}
			}
		});
		expect(r.paginas?.producto).toMatchObject({
			combos: false,
			guiaTalles: {
				tabla: [
					['Talle', 'Busto cm'],
					['S', '88'],
					['M', '']
				]
			}
		});
		const vacia = paraGuardar({
			...base,
			paginas: {
				producto: {
					variante: 'grilla',
					relacionados: true,
					calcularEnvio: true,
					guiaTalles: {
						tabla: [
							['Talle', ''],
							['', '']
						]
					},
					bloques: []
				}
			}
		});
		expect(vacia.paginas?.producto?.guiaTalles).toBeUndefined();
	});
});

describe('campos', () => {
	it('las fechas van y vuelven en hora de Argentina', () => {
		expect(isoDeLocal('2026-11-20T23:59')).toBe('2026-11-20T23:59:00-03:00');
		expect(fechaLocal('2026-11-21T02:59:00Z')).toBe('2026-11-20T23:59');
		expect(fechaLocal('')).toBe('');
		expect(isoDeLocal('mañana')).toBe('');
	});

	it('el buscador ignora tildes y saca los elegidos', () => {
		const ps = [
			{
				id: 'a',
				nombre: 'Café de especialidad',
				foto: null,
				precio: 1,
				categoria: null,
				oferta: false
			},
			{ id: 'b', nombre: 'Taza', foto: null, precio: 1, categoria: 'Cafe', oferta: false }
		];
		expect(buscar(ps, 'cafe').map((p) => p.id)).toEqual(['a', 'b']);
		expect(buscar(ps, 'cafe', ['a']).map((p) => p.id)).toEqual(['b']);
	});
});
