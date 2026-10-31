/**
 * Newsletter de la tienda (docs/decisiones.md §50): la validación de una novedad (la misma que Go),
 * las filas del CSV de suscriptos (lo arma `csv` de server/zip.ts) y los párrafos de la vista previa.
 */
import * as v from 'valibot';

export const MAX_PRODUCTOS = 4;

export const EsquemaNovedad = v.object({
	asunto: v.pipe(
		v.string(),
		v.transform((s) => s.split(/\s+/).filter(Boolean).join(' ')),
		v.minLength(3, 'Poné un asunto.'),
		v.maxLength(120, 'El asunto puede tener hasta 120 caracteres.')
	),
	texto: v.pipe(
		v.string(),
		v.transform((s) => s.replaceAll('\r\n', '\n').trim()),
		v.minLength(10, 'Escribí la novedad (al menos 10 caracteres).'),
		v.maxLength(5000, 'El texto puede tener hasta 5000 caracteres.')
	),
	productos: v.pipe(
		v.array(v.pipe(v.string(), v.maxLength(80))),
		v.maxLength(MAX_PRODUCTOS, 'Elegí hasta 4 productos.'),
		v.transform((ps) => [...new Set(ps.filter(Boolean))])
	)
});

/** Valida la novedad: los datos normalizados o el primer problema. */
export function validarNovedad(
	n: unknown
): { ok: true; novedad: v.InferOutput<typeof EsquemaNovedad> } | { ok: false; error: string } {
	const r = v.safeParse(EsquemaNovedad, n);
	return r.success ? { ok: true, novedad: r.output } : { ok: false, error: r.issues[0].message };
}

/** Los párrafos del texto, como los arma el mail (separados por una línea en blanco). */
export const parrafos = (texto: string) =>
	texto
		.replaceAll('\r\n', '\n')
		.split(/\n\s*\n/)
		.map((p) => p.trim())
		.filter(Boolean);

type Fila = { email: string; confirmado: boolean; creado: string; origen: string };

const ORIGEN: Record<string, string> = { tienda: 'Tienda online', checkout: 'Checkout' };

/** El origen como lo lee la tienda. */
export const origenDe = (o: string) => ORIGEN[o] ?? o;

/** Las filas del CSV (la primera es el encabezado), con la fecha en hora de Argentina. */
export function filasCsv(ss: Fila[]): string[][] {
	const fecha = (iso: string) =>
		new Date(iso).toLocaleDateString('en-CA', { timeZone: 'America/Argentina/Buenos_Aires' });
	return [
		['Email', 'Estado', 'Desde', 'Origen'],
		...ss.map((s) => [
			// Excel ejecuta lo que empieza con =, +, - o @ (un email puede empezar así).
			/^[=+\-@]/.test(s.email) ? `'${s.email}` : s.email,
			s.confirmado ? 'Confirmado' : 'Sin confirmar',
			fecha(s.creado),
			origenDe(s.origen)
		])
	];
}
