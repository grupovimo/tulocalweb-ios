import { error, json } from '@sveltejs/kit';
import * as v from 'valibot';
import { datosDe } from '#lib/server/datos/index.ts';
import { errorOpciones, errorVariantes, MAX_FOTOS, MAX_VARIANTES } from '#lib/tienda/variantes.ts';
import type { RequestHandler } from './$types';

// Una foto de Unsplash (id), una URL https (las subidas a /f/ y las importadas de Empretienda) o,
// en las pruebas locales, una subida a http://localhost.
const Foto = v.union([
	v.pipe(v.string(), v.regex(/^\d{10,}-[0-9a-f]+$/)),
	v.pipe(v.string(), v.maxLength(500), v.url(), v.startsWith('https://')),
	v.pipe(v.string(), v.maxLength(200), v.regex(/^http:\/\/(localhost|127\.0\.0\.1):\d+\/f\//))
]);
const Texto20 = v.pipe(v.string(), v.trim(), v.minLength(1), v.maxLength(20));

const Editable = v.object({
	nombre: v.pipe(v.string(), v.trim(), v.minLength(2, 'Poné un nombre.'), v.maxLength(120)),
	descripcion: v.optional(v.pipe(v.string(), v.maxLength(4000))),
	categoria: v.optional(v.nullable(v.pipe(v.string(), v.trim(), v.maxLength(60)))),
	precio: v.pipe(v.number(), v.integer(), v.minValue(0), v.maxValue(1_000_000_000)),
	// Precio mayorista (§49): sin venir (o null) queda el que había; 0 lo saca.
	precioMayorista: v.optional(
		v.nullable(v.pipe(v.number(), v.integer(), v.minValue(0), v.maxValue(1_000_000_000)))
	),
	costo: v.optional(v.pipe(v.number(), v.integer(), v.minValue(0), v.maxValue(1_000_000_000))),
	sku: v.optional(v.pipe(v.string(), v.trim(), v.maxLength(40))),
	stock: v.optional(v.pipe(v.number(), v.integer(), v.minValue(0), v.maxValue(1_000_000))),
	foto: v.optional(v.nullable(Foto)),
	fotos: v.optional(v.pipe(v.array(Foto), v.maxLength(MAX_FOTOS, `Hasta ${MAX_FOTOS} fotos.`))),
	opciones: v.optional(
		v.pipe(
			v.array(
				v.object({
					nombre: Texto20,
					valores: v.pipe(v.array(Texto20), v.minLength(1), v.maxLength(20))
				})
			),
			v.maxLength(3, 'Hasta 3 opciones.')
		)
	),
	// Sólo las que se tocan, identificadas por sus valores (el id no cuenta).
	variantes: v.optional(
		v.pipe(
			v.array(
				v.object({
					valores: v.record(Texto20, Texto20),
					sku: v.optional(v.pipe(v.string(), v.trim(), v.maxLength(40))),
					precio: v.optional(
						v.nullable(v.pipe(v.number(), v.integer(), v.minValue(0), v.maxValue(1_000_000_000)))
					),
					precioMayorista: v.optional(
						v.nullable(v.pipe(v.number(), v.integer(), v.minValue(0), v.maxValue(1_000_000_000)))
					),
					stock: v.optional(v.pipe(v.number(), v.integer(), v.minValue(0), v.maxValue(1_000_000))),
					foto: v.optional(v.nullable(Foto)),
					activa: v.optional(v.boolean())
				})
			),
			v.maxLength(MAX_VARIANTES)
		)
	),
	canales: v.optional(v.record(v.picklist(['web', 'ml', 'ig', 'google']), v.boolean()))
});

/** La ficha del producto (el editor). */
export const GET: RequestHandler = async ({ locals, params }) => {
	const p = await datosDe(locals).producto(params.id);
	if (!p) error(404, 'Ese producto no existe.');
	return json(p, { headers: { 'cache-control': 'no-store' } });
};

/** Guardar (o crear, con id "nuevo"). */
export const PUT: RequestHandler = async ({ locals, params, request }) => {
	// Como la API Go: el cuerpo del editor llega hasta 128 KB.
	const texto = await request.text();
	if (texto.length > 128 * 1024) error(413, 'El producto tiene demasiados datos para guardar.');
	let cuerpo: unknown = null;
	try {
		cuerpo = JSON.parse(texto);
	} catch {
		// queda null y falla la validación
	}
	const r = v.safeParse(Editable, cuerpo);
	if (!r.success) error(400, r.issues[0].message);
	// Como Go: el mayorista tiene que ser menor que el precio (0 o null no es un precio).
	if (r.output.precioMayorista && r.output.precioMayorista >= r.output.precio)
		error(400, 'El precio mayorista tiene que ser menor que el precio.');
	const { opciones, variantes, fotos } = r.output;
	// Las combinaciones las rearma quien guarda; acá se rechaza lo que no puede corresponder.
	const mal = opciones
		? (errorOpciones(opciones) ??
			(variantes && fotos ? errorVariantes(opciones, variantes, fotos) : null))
		: variantes?.length
			? 'Sin opciones no hay variantes.'
			: null;
	if (mal) error(400, mal);
	const datos = datosDe(locals);
	const antes = params.id === 'nuevo' ? null : await datos.producto(params.id);
	const p = await datos.guardarProducto(params.id === 'nuevo' ? null : params.id, r.output);
	if (!p) error(404, 'Ese producto no existe.');
	// Mercado Libre: prenderlo publica el producto y apagarlo pausa la publicación.
	const ml = !!r.output.canales?.ml;
	let mercadoLibre: { publicado?: string; error?: string } | undefined;
	if (ml !== !!antes?.canales?.ml && r.output.canales) {
		if (ml) {
			const pub = await datos.publicarEnMercadoLibre(p.id);
			mercadoLibre = 'error' in pub ? { error: pub.error } : { publicado: pub.permalink };
			if ('error' in pub)
				await datos.guardarProducto(p.id, {
					...r.output,
					canales: { ...r.output.canales, ml: false }
				});
		} else await datos.despublicarDeMercadoLibre(p.id);
	}
	return json({ ...p, ...(mercadoLibre ? { mercadoLibre } : {}) });
};

/** Borrar el producto. Si está publicado en Mercado Libre, primero hay que sacarlo de ahí. */
export const DELETE: RequestHandler = async ({ locals, params }) => {
	const r = await datosDe(locals).borrarProducto(params.id);
	if (r === 'no-existe') error(404, 'Ese producto no existe.');
	if (r === 'no-se-puede') error(409, 'Está publicado en Mercado Libre: sacalo de ahí primero.');
	return new Response(null, { status: 204 });
};
