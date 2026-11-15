import { error } from '@sveltejs/kit';
import * as v from 'valibot';
import { FONDOS, LIMITES, OPCIONES_ESTILO } from '#lib/panel/mi-tienda/catalogo.ts';
import type { DisenoV2 } from '#lib/panel/mi-tienda/diseno.ts';
import { linkValido, paraGuardar, problema } from '#lib/panel/mi-tienda/guardar.ts';
import { plantilla } from '#lib/plantillas.ts';
import { FUENTES } from '#lib/tienda/estilo/fuentes.ts';
import { MAX_MENU } from '#lib/tienda/estilo/menu.ts';
import { datosDe } from '#lib/server/datos/index.ts';
import { MAX_VENTANAS } from '#lib/tienda/sistema/promos.ts';
import type { RequestHandler } from './$types';

// Las fotos subidas a /panel/fotos vuelven con su origen (https en producción, http en local).
const foto = v.pipe(
	v.string(),
	v.maxLength(300),
	v.regex(/^$|^https?:\/\/[^\s"'<>]+$/, 'Esa foto no es válida: subila de nuevo.')
);
const enlace = v.pipe(
	v.string(),
	v.check((a) => linkValido(a), 'Elegí adónde lleva el link.')
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
const opcion = (o: string[]) =>
	v.optional(v.picklist(o as [string, ...string[]], 'Elegí una de las opciones.'));
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
	leyendaPie: v.optional(texto(160, 'La línea del pie')),
	// v2 (bloques.json › estilo): vacío es lo de la dirección y el preset.
	direccion: opcion(OPCIONES_ESTILO.direccion.opciones),
	forma: opcion(OPCIONES_ESTILO.forma.opciones),
	densidad: opcion(OPCIONES_ESTILO.densidad.opciones),
	ancho: opcion(OPCIONES_ESTILO.ancho.opciones),
	fotoProducto: opcion(OPCIONES_ESTILO.fotoProducto.opciones),
	encabezado: opcion(OPCIONES_ESTILO.encabezado.opciones),
	pie: opcion(OPCIONES_ESTILO.pie.opciones),
	barra: v.optional(
		v.object({
			variante: v.picklist(['fija', 'rotativa']),
			mensajes: v.pipe(
				v.array(
					v.object({
						texto: v.pipe(
							texto(100, 'Cada mensaje'),
							v.minLength(1, 'Cada mensaje necesita un texto.')
						),
						link: v.optional(enlace),
						hasta: v.optional(v.pipe(v.string(), v.isoTimestamp('Esa fecha no es válida.')))
					})
				),
				v.maxLength(5, 'La barra va hasta 5 mensajes.')
			),
			fondo: opcion(FONDOS.opciones)
		})
	),
	whatsappFlotante: v.optional(v.boolean()),
	// Las herramientas de conversión: los campos se limpian contra el catálogo en paraGuardar.
	conversion: v.optional(v.looseObject({})),
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
						v.check((a) => linkValido(a), 'Elegí adónde lleva cada pestaña.')
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
// Los bloques: forma mínima acá; los datos se limpian contra el catálogo en paraGuardar.
const Bloques = v.pipe(
	v.array(v.looseObject({ id: v.string(), tipo: v.string(), variante: v.string() })),
	v.maxLength(LIMITES.bloquesPorPagina, `Cada página va hasta ${LIMITES.bloquesPorPagina} bloques.`)
);
const Diseno = v.object({
	version: v.literal(2, 'Recargá la página: el editor cambió.'),
	inicio: Bloques,
	paginas: v.optional(
		v.object({
			catalogo: v.optional(v.looseObject({})),
			producto: v.optional(v.looseObject({ bloques: v.optional(Bloques) })),
			carrito: v.optional(v.looseObject({})),
			nosotros: v.optional(v.object({ bloques: v.optional(Bloques) }))
		})
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
	estilo: v.optional(Estilo),
	// Las ventanas: forma mínima acá; cada una se valida contra el catálogo en paraGuardar.
	ventanas: v.optional(
		v.pipe(
			v.array(
				v.looseObject({
					id: v.string(),
					contenido: v.string(),
					variante: v.string(),
					datos: v.optional(v.looseObject({}))
				})
			),
			v.maxLength(MAX_VENTANAS, `Van hasta ${MAX_VENTANAS} ventanas.`)
		)
	)
});

/**
 * Guardar el borrador del diseño (v2, hasta limites.bytesDiseno). Lo incompleto o fuera del
 * catálogo se saca acá también (paraGuardar), como en el editor: la API Go no lo aceptaría.
 */
export const PUT: RequestHandler = async ({ locals, request }) => {
	const cuerpo = await request.text();
	if (new TextEncoder().encode(cuerpo).length > LIMITES.bytesDiseno)
		error(413, 'El diseño quedó muy grande: sacá algunos bloques o fotos.');
	let json: unknown = null;
	try {
		json = JSON.parse(cuerpo);
	} catch {
		error(400, 'No se entendió el diseño.');
	}
	const r = v.safeParse(Diseno, json);
	if (!r.success) error(400, r.issues[0].message);
	const d = paraGuardar(r.output as unknown as DisenoV2);
	const p = problema(d);
	if (p) error(400, p);
	await datosDe(locals).guardarMiTienda(d);
	return new Response(null, { status: 204 });
};
