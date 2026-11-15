/**
 * Mi tienda › Punto de partida: cambiar a uno de los 17 presets (packages/contrato/presets.json)
 * con la tienda armada. Arma la portada del preset y conserva lo cargado en los bloques del mismo
 * tipo (el título del banner, las preguntas, las fotos); los bloques que el preset no trae siguen
 * al final, en su orden. El look (dirección, forma, fuentes, colores) pasa a ser el del preset; la
 * marca (logos, menú, redes, barra) queda. Se puede deshacer.
 */
import PRESETS from '@mostrador/contrato/presets.json';
import { idsDe, nuevoId, type Bloque } from './bloques.ts';
import { defDe } from './catalogo.ts';
import { ESTILO_DEL_PRESET, type DisenoV2 } from './diseno.ts';

export type Preset = {
	direccion: 'vidriera' | 'mostrador' | 'pedido';
	estilo: Record<string, string>;
	orden: string[];
	variantes: Record<string, string>;
	gastronomia: boolean;
};

export const PRESETS_TIENDA = PRESETS.presets as Record<string, Preset>;

export function aplicarPreset(d: DisenoV2, id: string): DisenoV2 {
	const p = PRESETS_TIENDA[id];
	if (!p) return d;
	// Los bloques de la tienda por tipo, en orden: el primero de cada tipo pasa al preset.
	const porTipo = new Map<string, Bloque[]>();
	for (const b of d.inicio) porTipo.set(b.tipo, [...(porTipo.get(b.tipo) ?? []), b]);
	const usados = idsDe(d);
	const inicio: Bloque[] = p.orden
		.filter((tipo) => defDe(tipo))
		.map((tipo) => {
			const viejo = porTipo.get(tipo)?.shift();
			const variante = p.variantes[tipo] ?? defDe(tipo)!.variantes[0];
			if (viejo) {
				const { oculto: _, ...b } = viejo;
				return { ...b, variante };
			}
			const id = nuevoId(usados);
			usados.add(id);
			return { id, tipo, variante, datos: {} };
		});
	// Lo que la tienda sumó y el preset no trae (un banner, preguntas) no se pierde.
	for (const resto of porTipo.values()) inicio.push(...resto);
	const estilo = { ...d.estilo };
	for (const k of ESTILO_DEL_PRESET) delete estilo[k];
	return { ...d, plantilla: id, paleta: 0, estilo, inicio };
}
