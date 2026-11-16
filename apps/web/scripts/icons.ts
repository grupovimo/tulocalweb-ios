/**
 * Genera src/lib/icons.ts con los íconos que usa el handoff (design/*.dc.html y templates-data.js):
 * Phosphor (ph:*) y logos de marcas (logos:*, simple-icons:*), cada uno como export con nombre
 * para que el bundle sólo lleve los que se importan. Para cada ph:x también va ph:x-fill si existe
 * (el menú lateral muestra el ícono relleno del ítem activo).
 *
 * Uso: bun scripts/icons.ts [ph:otro-icono ...]. Lo que se use siempre va en EXTRA (si no, se pierde
 * al regenerar).
 */
import { readdirSync, readFileSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';

type Set = {
	prefix: string;
	width?: number;
	height?: number;
	icons: Record<string, { body: string; width?: number; height?: number }>;
	aliases?: Record<string, { parent: string }>;
};

const root = join(import.meta.dirname, '..');
const design = join(root, '../../design');
const load = (p: string): Set =>
	JSON.parse(readFileSync(join(root, '../../node_modules/@iconify-json', p, 'icons.json'), 'utf8'));
const sets: Record<string, Set> = {
	ph: load('ph'),
	logos: load('logos'),
	'simple-icons': load('simple-icons')
};

const files = readdirSync(design).filter(
	(f) => f.endsWith('.dc.html') || f === 'templates-data.js'
);
// Íconos que no aparecen en el handoff pero se usan (reemplazos de logos que no existen en Iconify).
const EXTRA = [
	// Tienda: bloques «videos» (play, sonido) y «clientes» (el ícono del editor).
	'ph:play-fill',
	'ph:film-strip-fill',
	'ph:speaker-simple-slash-fill',
	'ph:speaker-simple-high-fill',
	// Sesiones abiertas desde una tablet (Configuración › Seguridad).
	'ph:device-tablet-fill',
	// Salir del backoffice del equipo.
	'ph:sign-out-bold',
	'simple-icons:googlesheets',
	'simple-icons:microsoftexcel',
	// /estado cuando algo falla o todavía no hay datos (el handoff sólo dibuja "todo funciona").
	'ph:warning-bold',
	'ph:hourglass-bold',
	// La meta del mes en el Resumen.
	'ph:target',
	// "¿Te sirvió este artículo?" en la ayuda.
	'ph:thumbs-up-bold',
	'ph:thumbs-down-bold',
	// Exportar las piezas de /marca.
	'ph:image-bold',
	'ph:file-pdf-bold',
	'ph:file-zip-bold',
	'ph:upload-simple-bold',
	// Turnos: la agenda del panel y la reserva en la tienda.
	'ph:calendar-blank',
	'ph:arrow-down-bold',
	'ph:pencil-simple',
	'ph:calendar-x',
	// Compartir un producto (panel y ficha pública).
	'ph:share-network-bold',
	// Seguir con Apple en /ingresar (Clerk).
	'logos:apple',
	// Instalar el panel en el iPhone: Compartir › Agregar a inicio (los íconos que muestra Safari).
	'ph:export',
	'ph:plus-square'
];
const names = new Set<string>([...EXTRA, ...process.argv.slice(2)]);
for (const f of files)
	for (const m of readFileSync(join(design, f), 'utf8').matchAll(
		/\b(ph|logos|simple-icons):([a-z0-9-]+)/g
	))
		names.add(`${m[1]}:${m[2]}`);
for (const n of [...names])
	if (n.startsWith('ph:') && !n.endsWith('-fill'))
		names.add(n.replace(/-(bold|light|thin|duotone)$/, '') + '-fill');

const camel = (s: string) =>
	s.replace(/^simple-icons:/, 'si-').replace(/[:-]+([a-z0-9])/g, (_, c) => c.toUpperCase());

const out: string[] = [];
const missing: string[] = [];
for (const name of [...names].sort()) {
	const [prefix, id] = name.split(':');
	const set = sets[prefix];
	const icon = set.icons[id] ?? set.icons[set.aliases?.[id]?.parent ?? ''];
	if (!icon) {
		missing.push(name);
		continue;
	}
	const w = icon.width ?? set.width ?? 16;
	const h = icon.height ?? set.height ?? 16;
	out.push(`export const ${camel(name)}: IconData = [${w}, ${h}, ${JSON.stringify(icon.body)}];`);
}

writeFileSync(
	join(root, 'src/lib/icons.ts'),
	`// Generado por scripts/icons.ts — no editar a mano. Para sumar un ícono: bun scripts/icons.ts ph:nombre\n` +
		`/** [ancho, alto, cuerpo SVG] */\nexport type IconData = readonly [number, number, string];\n\n` +
		out.join('\n') +
		'\n'
);
console.log(`${out.length} íconos`);
// Los -fill agregados que no existen en Phosphor no son error: el ícono base no tiene variante.
const real = missing.filter((m) => !m.endsWith('-fill'));
if (real.length) console.log('No encontrados:', real.join(', '));
