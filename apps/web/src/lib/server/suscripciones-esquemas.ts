/**
 * Validación de los planes de suscripción del panel y del alta desde la tienda (§51), con las
 * mismas reglas que Go (suscripciones.Plan.validar y Alta.validar).
 */
import * as v from 'valibot';

const uuid = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;
// En la demo los productos tienen ids del catálogo de ejemplo ("malbec-valle-de-uco-2021").
const idProducto = v.pipe(v.string(), v.regex(/^[A-Za-z0-9_-]{1,64}$/, 'Elegí un producto.'));

export const PlanSuscripcion = v.pipe(
	v.object({
		nombre: v.pipe(
			v.string(),
			v.trim(),
			v.minLength(2, 'Poné el nombre del plan.'),
			v.maxLength(60, 'El nombre puede tener hasta 60 letras.')
		),
		descripcion: v.pipe(v.string(), v.trim(), v.maxLength(400)),
		foto: v.pipe(
			v.nullable(v.pipe(v.string(), v.trim(), v.maxLength(2048))),
			v.transform((s) => s || null)
		),
		precio: v.pipe(
			v.number(),
			v.integer(),
			v.minValue(100, 'Poné el precio (desde $ 100).'),
			v.maxValue(100_000_000)
		),
		frecuencia: v.picklist(['quincenal', 'mensual', 'bimestral'], 'Elegí cada cuánto se cobra.'),
		productos: v.pipe(
			v.array(
				v.object({
					productoId: idProducto,
					varianteId: v.pipe(v.string(), v.maxLength(64)),
					cantidad: v.pipe(v.number(), v.integer(), v.minValue(1), v.maxValue(99))
				})
			),
			v.maxLength(30, 'Hasta 30 productos por plan.')
		),
		incluye: v.pipe(v.string(), v.trim(), v.maxLength(400)),
		envio: v.boolean(),
		retiro: v.boolean(),
		cupo: v.nullable(
			v.pipe(
				v.number(),
				v.integer(),
				v.minValue(1, 'El cupo es de 1 en adelante.'),
				v.maxValue(100_000)
			)
		),
		activo: v.boolean()
	}),
	v.check((p) => p.envio || p.retiro, 'Elegí si se envía, se retira o las dos.'),
	v.check(
		(p) => p.productos.length > 0 || p.incluye.length > 0,
		'Contá qué incluye: elegí productos o escribilo.'
	)
);

export const AltaSuscripcion = v.pipe(
	v.object({
		planId: v.pipe(v.string(), v.regex(/^[A-Za-z0-9_-]{1,64}$/, 'Elegí un plan.')),
		nombre: v.pipe(
			v.string(),
			v.trim(),
			v.minLength(2, 'Poné tu nombre.'),
			v.maxLength(80, 'El nombre puede tener hasta 80 letras.')
		),
		email: v.pipe(v.string(), v.trim(), v.email('Revisá el email.'), v.maxLength(120)),
		telefono: v.pipe(v.string(), v.trim(), v.regex(/^[\d\s+()-]{8,20}$/, 'Revisá el teléfono.')),
		entrega: v.picklist(['envio', 'retiro'], 'Elegí cómo lo recibís.'),
		direccion: v.optional(v.pipe(v.string(), v.trim(), v.maxLength(200)), ''),
		cp: v.optional(v.pipe(v.string(), v.trim(), v.maxLength(10)), '')
	}),
	v.check((a) => a.entrega === 'retiro' || a.direccion.length >= 4, 'Poné la dirección de entrega.')
);

export const esUuid = (s: string) => uuid.test(s);
