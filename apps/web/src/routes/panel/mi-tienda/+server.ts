import { error } from '@sveltejs/kit';
import * as v from 'valibot';
import { plantilla } from '#lib/plantillas.ts';
import { FUENTES } from '#lib/tienda/estilo/fuentes.ts';
import { hrefDe, MAX_MENU } from '#lib/tienda/estilo/menu.ts';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

// Las fotos subidas a /panel/fotos vuelven con su origen (https en producción, http en local).
const foto = v.pipe(
	v.string(),
	v.maxLength(300),
	v.regex(/^$|^https?:\/\/[^\s"'<>]+$/, 'Esa foto no es válida: subila de nuevo.')
);
const color = v.pipe(v.string(), v.regex(/^(#[0-9A-Fa-f]{6})?$/, 'Ese color no es válido.'));
const texto = (max: number, que: string) =>
	v.pipe(
		v.string(),
		v.trim(),
		v.maxLength(max, `${que} va hasta ${max} letras.`),
		v.regex(/^[^<>]*$/, `${que} no puede llevar < ni >.`)
	);
const fuente = v.pipe(
	v.string(),
	v.check((f) => f === '' || f in FUENTES, 'Elegí una de las tipografías.')
);
const parcial = <T extends Record<string, v.GenericSchema<string, string>>>(campos: T) =>
	v.optional(
		v.strictObject(
			Object.fromEntries(Object.entries(campos).map(([k, s]) => [k, v.optional(s)])) as {
				[K in keyof T]: v.OptionalSchema<T[K], undefined>;
			}
		)
	);
const Estilo = v.object({
	colores: parcial({
		fondo: color,
		superficie: color,
		texto: color,
		textoSuave: color,
		borde: color,
		acento: color,
		sobreAcento: color,
		encabezado: color,
		encabezadoTexto: color,
		pie: color,
		pieTexto: color
	}),
	fuenteTitulos: v.optional(fuente),
	fuenteTexto: v.optional(fuente),
	logoOscuro: v.optional(foto),
	icono: v.optional(foto),
	logoAlto: v.optional(v.pipe(v.number(), v.integer(), v.minValue(0), v.maxValue(72))),
	anuncio: v.optional(texto(120, 'El anuncio')),
	anuncioLink: v.optional(foto),
	leyendaPie: v.optional(texto(160, 'La línea del pie')),
	textos: parcial({
		cats: texto(60, 'El título'),
		prods: texto(60, 'El título'),
		revs: texto(60, 'El título'),
		ig: texto(60, 'El título'),
		wa: texto(60, 'El título')
	}),
	menu: v.optional(
		v.pipe(
			v.array(
				v.object({
					t: v.pipe(
						v.string(),
						v.trim(),
						v.minLength(1, 'Cada pestaña necesita un texto.'),
						v.maxLength(30, 'El texto de la pestaña va hasta 30 letras.'),
						v.regex(/^[^<>]*$/, 'La pestaña no puede llevar < ni >.')
					),
					a: v.pipe(
						v.string(),
						v.check((a) => hrefDe(a, '') !== null, 'Elegí adónde lleva cada pestaña.')
					)
				})
			),
			v.maxLength(MAX_MENU, `El menú va hasta ${MAX_MENU} pestañas.`)
		)
	),
	redes: parcial({
		instagram: texto(60, 'Instagram'),
		tiktok: texto(60, 'TikTok'),
		facebook: texto(100, 'Facebook'),
		youtube: texto(100, 'YouTube'),
		email: texto(120, 'El email'),
		telefono: texto(30, 'El teléfono')
	})
});
const IDS = ['hero', 'cats', 'prods', 'revs', 'ig', 'wa'] as const;
const Diseno = v.object({
	secciones: v.pipe(
		v.array(v.object({ id: v.picklist(IDS), visible: v.boolean() })),
		v.check(
			(s) => s.length === IDS.length && new Set(s.map((x) => x.id)).size === IDS.length,
			'Faltan secciones.'
		)
	),
	paleta: v.pipe(v.number(), v.integer(), v.minValue(0), v.maxValue(3)),
	plantilla: v.optional(
		v.pipe(
			v.string(),
			v.check((id) => id === '' || !!plantilla(id), 'Elegí una de las plantillas.')
		)
	),
	titulo: v.optional(
		v.pipe(v.string(), v.trim(), v.maxLength(80, 'El título va hasta 80 letras.'))
	),
	lema: v.optional(v.pipe(v.string(), v.trim(), v.maxLength(120, 'El lema va hasta 120 letras.'))),
	portada: v.optional(foto),
	logo: v.optional(foto),
	compartirTitulo: v.optional(
		v.pipe(v.string(), v.trim(), v.maxLength(70, 'El título para compartir va hasta 70 letras.'))
	),
	compartirDescripcion: v.optional(
		v.pipe(
			v.string(),
			v.trim(),
			v.maxLength(160, 'La descripción para compartir va hasta 160 letras.')
		)
	),
	compartirFoto: v.optional(foto),
	estilo: v.optional(Estilo)
});

/** Guardar el borrador del diseño. */
export const PUT: RequestHandler = async ({ locals, request }) => {
	const r = v.safeParse(Diseno, await request.json().catch(() => null));
	if (!r.success) error(400, r.issues[0].message);
	await datosDe(locals).guardarMiTienda(r.output);
	return new Response(null, { status: 204 });
};
