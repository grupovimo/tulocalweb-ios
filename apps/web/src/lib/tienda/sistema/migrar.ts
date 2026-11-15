/**
 * La migración del diseño v1 (secciones fijas) al v2 (bloques). Función pura, igual que
 * mitienda.Migrar en Go: las dos corren los casos de packages/contrato/casos/diseno-v2.json
 * (migrar.spec.ts). Mientras convivan, la tienda migra al vuelo cualquier diseño v1 que lea.
 *
 * Reglas: un bloque por sección del orden del preset (id = tipo), en el orden del preset salvo que
 * la tienda haya reordenado (entonces el suyo, filtrado a las del preset); las ocultas quedan
 * oculto:true; hero → portada con titulo/lema/portada; textos[sección] → datos.titulo; wa oculta →
 * estilo.whatsappFlotante=false; anuncio → estilo.barra fija con un mensaje; se sacan secciones,
 * textos, anuncio y anuncioLink. Un diseño ya v2 vuelve igual. null → el inicial del preset.
 */
import type { EstiloTienda, SeccionConTitulo } from '../estilo/estilo.ts';
import type { Bloque, DisenoV2, TipoBloque } from './catalogo.ts';
import { inicioDe, presetDe } from './presets.ts';

type SeccionV1 = 'hero' | 'cats' | 'prods' | 'revs' | 'ig' | 'wa';
type DisenoV1 = {
	version?: number;
	secciones?: { id: SeccionV1; visible: boolean }[];
	paleta?: number;
	plantilla?: string;
	titulo?: string;
	lema?: string;
	portada?: string;
	logo?: string;
	compartirTitulo?: string;
	compartirDescripcion?: string;
	compartirFoto?: string;
	estilo?: EstiloTienda | null;
};

const INICIAL: SeccionV1[] = ['hero', 'cats', 'prods', 'revs', 'ig', 'wa'];
const A_BLOQUE: Record<SeccionV1, TipoBloque> = {
	hero: 'portada',
	cats: 'categorias',
	prods: 'productos',
	revs: 'resenas',
	ig: 'galeria',
	wa: 'whatsapp'
};
const DE_BLOQUE = Object.fromEntries(
	Object.entries(A_BLOQUE).map(([s, t]) => [t, s as SeccionV1])
) as Partial<Record<TipoBloque, SeccionV1>>;
const MARCA = [
	'plantilla',
	'titulo',
	'lema',
	'portada',
	'logo',
	'compartirTitulo',
	'compartirDescripcion',
	'compartirFoto'
] as const;

/** ¿Es un diseño v2? */
export const esV2 = (d: unknown): d is DisenoV2 =>
	!!d && typeof d === 'object' && (d as { version?: unknown }).version === 2;

/**
 * El diseño en v2. `plantilla` es la de la tienda (tiendas.plantilla); si el borrador eligió otra
 * (`diseno.plantilla`), manda esa.
 */
export function migrar(diseno: unknown, plantilla: string): DisenoV2 {
	if (esV2(diseno)) return diseno;
	if (!diseno || typeof diseno !== 'object')
		return { version: 2, paleta: 0, inicio: inicioDe(presetDe(plantilla)) };
	const v1 = diseno as DisenoV1;
	const preset = presetDe(v1.plantilla || plantilla);
	const secciones = (v1.secciones ?? []).filter((s) => s && INICIAL.includes(s.id));
	const visible = new Map(secciones.map((s) => [s.id, s.visible !== false]));

	// El orden: el del preset si la tienda nunca reordenó; si reordenó, el suyo (sólo las del preset).
	const pos = secciones.map((s) => INICIAL.indexOf(s.id));
	const reordenada = pos.some((p, i) => i > 0 && p < pos[i - 1]);
	const tipos: TipoBloque[] = reordenada
		? secciones.map((s) => A_BLOQUE[s.id]).filter((t) => preset.orden.includes(t))
		: preset.orden;

	const textos = v1.estilo?.textos ?? {};
	const inicio: Bloque[] = inicioDe({ ...preset, orden: tipos }).map((b) => {
		const sec = DE_BLOQUE[b.tipo];
		const datos: Record<string, unknown> = {};
		if (b.tipo === 'portada') {
			if (v1.titulo?.trim()) datos.titulo = v1.titulo.trim();
			if (v1.lema?.trim()) datos.texto = v1.lema.trim();
			if (v1.portada?.trim()) datos.fotos = [v1.portada.trim()];
		} else if (sec && sec !== 'hero') {
			const t = textos[sec as SeccionConTitulo]?.trim();
			if (t) datos.titulo = t;
		}
		return {
			...b,
			...(sec && visible.get(sec) === false ? { oculto: true } : {}),
			datos
		};
	});

	const out: DisenoV2 = { version: 2, paleta: v1.paleta ?? 0, inicio };
	for (const k of MARCA) if (v1[k]) out[k] = v1[k];

	// El estilo queda igual, salvo lo que pasa a bloques o a la barra.
	const estilo: EstiloTienda = { ...(v1.estilo ?? {}) };
	const anuncio = estilo.anuncio?.trim();
	const anuncioLink = estilo.anuncioLink?.trim();
	delete estilo.textos;
	delete estilo.anuncio;
	delete estilo.anuncioLink;
	if (anuncio && !estilo.barra)
		estilo.barra = {
			variante: 'fija',
			mensajes: [{ texto: anuncio, ...(anuncioLink ? { link: anuncioLink } : {}) }]
		};
	if (visible.get('wa') === false) estilo.whatsappFlotante = false;
	if (Object.keys(estilo).length) out.estilo = estilo;
	return out;
}
