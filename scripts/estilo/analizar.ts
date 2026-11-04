// Lista los colores de una plantilla con cuántas veces aparecen y en qué propiedad (para elegir sus roles).
import { readFileSync } from 'node:fs';
const f = process.argv[2];
const s = readFileSync(f, 'utf8');
const re = /([a-z-]+)\s*:\s*[^;"'`{}]*?(#[0-9a-fA-F]{6}\b|#[0-9a-fA-F]{3}\b)|(#[0-9a-fA-F]{6}\b|#[0-9a-fA-F]{3}\b)/g;
const cnt: Record<string, Record<string, number>> = {};
for (const m of s.matchAll(re)) {
	let hex = (m[2] || m[3]).toUpperCase();
	if (hex.length === 4) hex = '#' + [...hex.slice(1)].map((c) => c + c).join('');
	const prop = m[1] || '?';
	(cnt[hex] ??= {})[prop] = ((cnt[hex] ??= {})[prop] ?? 0) + 1;
}
const filas = Object.entries(cnt).map(([h, p]) => [h, Object.values(p).reduce((a, b) => a + b, 0), p] as const);
filas.sort((a, b) => b[1] - a[1]);
for (const [h, n, p] of filas) console.log(h, n, Object.entries(p).sort((a, b) => b[1] - a[1]).slice(0, 4).map(([k, v]) => `${k}:${v}`).join(' '));
