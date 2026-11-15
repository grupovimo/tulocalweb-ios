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
import type { TiendaPublica } from '../tipos.ts';
import { validarBloques, type Bloque, type DisenoV2, type TipoBloque } from './catalogo.ts';
import { inicioDe, presetDe } from './presets.ts';

export type SeccionV1 = 'hero' | 'cats' | 'prods' | 'revs' | 'ig' | 'wa';
export type DisenoV1 = {
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

/**
 * El diseño v2 de una tienda pública: el que trae (`tienda.diseno`, ya migrado por la API) o, si no
 * viene, el v1 armado con lo que trae la tienda y migrado al vuelo. Los bloques inválidos se
 * descartan (validarBloques): la tienda nunca se rompe por un dato raro.
 */
export function disenoDeTienda(t: TiendaPublica): DisenoV2 {
	const d = esV2(t.diseno)
		? t.diseno
		: migrar(
				{
					secciones: [
						...t.secciones.map((id) => ({ id, visible: true })),
						...INICIAL.filter((s) => !t.secciones.includes(s)).map((id) => ({
							id,
							visible: false
						}))
					],
					paleta: t.paleta,
					titulo: t.portada?.titulo,
					lema: t.portada?.lema,
					portada: t.portada?.foto ?? undefined,
					logo: t.logo,
					estilo: t.estilo
				},
				t.plantilla
			);
	return { ...d, inicio: validarBloques(d.inicio) };
}

// ---------------------------------------------------------------------------------------------
// Convivencia con lo v1 (como mitienda/convivencia.go): mientras las plantillas viejas y el editor
// el onboarding sigan, se derivan del v2. Cuando todo migró, esto se borra.

/** Las secciones visibles como las esperan las plantillas viejas («wa» al final, si no se apagó). */
export function seccionesV1(d: DisenoV2): SeccionV1[] {
	const out: SeccionV1[] = [];
	for (const b of d.inicio) {
		const s = DE_BLOQUE[b.tipo];
		if (s && s !== 'wa' && !b.oculto && !out.includes(s)) out.push(s);
	}
	if (d.estilo?.whatsappFlotante !== false) out.push('wa');
	return out;
}

/** El estilo con el anuncio (el primer mensaje de la barra) y los títulos de las secciones. */
export function estiloV1(d: DisenoV2): EstiloTienda {
	const e: EstiloTienda = { ...(d.estilo ?? {}) };
	const m = e.barra?.mensajes?.[0];
	if (m?.texto) {
		e.anuncio = m.texto;
		if (m.link) e.anuncioLink = m.link;
	}
	const textos: Partial<Record<SeccionConTitulo, string>> = {};
	for (const b of d.inicio) {
		const s = DE_BLOQUE[b.tipo] as SeccionConTitulo | 'hero' | undefined;
		const t = b.datos.titulo;
		if (s && s !== 'hero' && typeof t === 'string' && t && !textos[s]) textos[s] = t;
	}
	if (Object.keys(textos).length) e.textos = textos;
	return e;
}

/** Prende o apaga los bloques de esas secciones (el onboarding: «mostrar destacados», etc.). */
export function conVisibles(d: DisenoV2, ver: Partial<Record<SeccionV1, boolean>>): DisenoV2 {
	return {
		...d,
		inicio: d.inicio.map((b) => {
			const s = DE_BLOQUE[b.tipo];
			const v = s ? ver[s] : undefined;
			if (typeof v !== 'boolean') return b;
			const resto: Bloque = { id: b.id, tipo: b.tipo, variante: b.variante, datos: b.datos };
			return v ? resto : { ...resto, oculto: true };
		})
	};
}
