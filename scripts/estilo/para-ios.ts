/**
 * Lo que la app de iPhone necesita para editar el estilo igual que la web: tipografías, combinaciones,
 * los colores de cada plantilla por rol y las páginas del menú. Sale a
 * apps/ios/TuLocalWeb/Recursos/estilo.json. Además copia el catálogo de bloques y los presets de la tienda
 * (packages/contrato/bloques.json y presets.json) a Recursos/ y a los assets de Android, para que los editores
 * de las apps generen los campos igual que la web. Correrlo cuando cambie algo de lib/tienda/estilo o del contrato:
 *   bun scripts/estilo/para-ios.ts
 */
import { copyFileSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';
import { PLANTILLAS } from '../../apps/web/src/lib/plantillas.ts';
import { FUENTES } from '../../apps/web/src/lib/tienda/estilo/fuentes.ts';
import { MAX_MENU, NOMBRE_PAGINA, PAGINAS } from '../../apps/web/src/lib/tienda/estilo/menu.ts';
import { COMBINACIONES, rolesDe } from '../../apps/web/src/lib/tienda/estilo/roles.ts';
import { acentos } from '../../apps/web/src/lib/tienda/acentos.ts';

const salida = {
	fuentes: Object.entries(FUENTES)
		.map(([id, f]) => ({ id, nombre: f.nombre, tipo: f.tipo }))
		.sort((a, b) => a.nombre.localeCompare(b.nombre)),
	combinaciones: COMBINACIONES,
	roles: Object.fromEntries(PLANTILLAS.map((p) => [p.id, rolesDe(p.id)])),
	acentos: Object.fromEntries(PLANTILLAS.map((p) => [p.id, acentos(p.id).map((a) => a.color)])),
	paginas: PAGINAS.map((id) => ({ id, nombre: NOMBRE_PAGINA[id] })),
	maxMenu: MAX_MENU
};
const ruta = join(import.meta.dir, '../../apps/ios/TuLocalWeb/Recursos/estilo.json');
writeFileSync(ruta, JSON.stringify(salida));
console.log('estilo.json:', Object.keys(salida.roles).length, 'plantillas,', salida.fuentes.length, 'fuentes');

// El catálogo de bloques y los presets, tal cual (los leen CatalogoBloques en iOS y en Android).
const contrato = join(import.meta.dir, '../../packages/contrato');
const destinos = [
	join(import.meta.dir, '../../apps/ios/TuLocalWeb/Recursos'),
	join(import.meta.dir, '../../apps/android/app/src/main/assets')
];
for (const archivo of ['bloques.json', 'presets.json']) {
	for (const d of destinos) copyFileSync(join(contrato, archivo), join(d, archivo));
}
console.log('bloques.json y presets.json copiados a', destinos.length, 'apps');
