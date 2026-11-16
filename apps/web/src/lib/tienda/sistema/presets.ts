/**
 * Los puntos de partida (las 17 plantillas) y el estilo final de una tienda. Fuente de verdad:
 * packages/contrato/presets.json (lo leen la web, Go y las apps).
 *
 * Estilo final = tokens de la dirección ⊕ preset ⊕ colores de rolesDe(plantilla) ⊕ diseno.estilo.
 * El diseño guarda sólo lo que la tienda cambió.
 */
import PRESETS_JSON from '@mostrador/contrato/presets.json';
import { acentoElegido } from '../acentos.ts';
import { colorValido, esOscuro } from '../estilo/css.ts';
import { FUENTES } from '../estilo/fuentes.ts';
import { rolesDe } from '../estilo/roles.ts';
import type { EstiloTienda, Zona } from '../estilo/estilo.ts';
import type { Rol } from '../estilo/tipos.ts';
import type { Bloque, Direccion, TipoBloque } from './catalogo.ts';

export type Preset = {
	direccion: Direccion;
	estilo: {
		forma: Forma;
		densidad: Densidad;
		fotoProducto: FotoProducto;
		encabezado: Encabezado;
		pie: Pie;
		fuenteTitulos: string;
		fuenteTexto: string;
	};
	orden: TipoBloque[];
	variantes: Partial<Record<TipoBloque, string>>;
	gastronomia: boolean;
};
export type Forma = 'recta' | 'redondeada' | 'pildora';
export type Densidad = 'aireada' | 'normal' | 'compacta';
export type FotoProducto = 'cuadrada' | 'vertical' | 'apaisada';
export type Encabezado = 'izquierda' | 'centrado' | 'buscador';
export type Pie = 'simple' | 'columnas' | 'grande';

export const PRESETS = PRESETS_JSON.presets as Record<string, Preset>;

/** El preset de una plantilla (una que no existe toma Litoral, como el motor de hoy). */
export function presetDe(plantilla: string | null | undefined): Preset {
	return PRESETS[plantilla ?? ''] ?? PRESETS.litoral;
}

/** Los bloques con los que arranca la portada de un preset (id = tipo, datos vacíos). */
export function inicioDe(preset: Preset): Bloque[] {
	return preset.orden.map((tipo) => ({
		id: tipo.toLowerCase(),
		tipo,
		variante: preset.variantes[tipo] ?? '',
		datos: {}
	}));
}

/** Lo que trae cada dirección antes del preset (README del handoff, tabla de tokens). */
export const DIRECCIONES: Record<
	Direccion,
	Omit<Preset['estilo'], 'fuenteTitulos' | 'fuenteTexto'> & { titulos: string; texto: string }
> = {
	vidriera: {
		forma: 'recta',
		densidad: 'aireada',
		fotoProducto: 'vertical',
		encabezado: 'centrado',
		pie: 'columnas',
		titulos: 'fraunces',
		texto: 'figtree'
	},
	mostrador: {
		forma: 'redondeada',
		densidad: 'normal',
		fotoProducto: 'vertical',
		encabezado: 'buscador',
		pie: 'columnas',
		titulos: 'bricolage',
		texto: 'instrument-sans'
	},
	pedido: {
		forma: 'redondeada',
		densidad: 'compacta',
		fotoProducto: 'vertical',
		encabezado: 'buscador',
		pie: 'simple',
		titulos: 'geist',
		texto: 'geist'
	}
};

export type Colores = Record<Rol | Zona, string>;

/** El estilo de la tienda ya resuelto: lo que dibuja el sistema, sin nada opcional. */
export type EstiloFinal = {
	direccion: Direccion;
	forma: Forma;
	densidad: Densidad;
	ancho: 'normal' | 'amplio';
	fotoProducto: FotoProducto;
	encabezado: Encabezado;
	pie: Pie;
	fuenteTitulos: string;
	fuenteTexto: string;
	colores: Colores;
	barra: NonNullable<EstiloTienda['barra']> | null;
	/** Sumar a la barra las cuotas, el envío gratis y la transferencia (barra.automaticos ≠ false). */
	barraAuto: boolean;
	whatsappFlotante: boolean;
	gastronomia: boolean;
};

const de = <T extends string>(x: unknown, opciones: readonly T[], si: T): T =>
	opciones.includes(x as T) ? (x as T) : si;

/**
 * El estilo final de una tienda: la dirección y el preset de su plantilla, con los colores de la
 * plantilla (y el acento de la paleta) y lo que cambió en Mi tienda › Estilo encima.
 */
export function estiloFinal(
	plantilla: string,
	paleta: number,
	estilo: EstiloTienda | null | undefined
): EstiloFinal {
	const preset = presetDe(plantilla);
	const e = estilo ?? {};
	const direccion = de(e.direccion, ['vidriera', 'mostrador', 'pedido'], preset.direccion);
	const base = DIRECCIONES[direccion];
	// Si la tienda cambió de dirección, el preset de otra dirección no manda: van los tokens.
	const p = direccion === preset.direccion ? preset.estilo : null;
	const colores = { ...rolesDe(plantilla) };
	const acento = acentoElegido(plantilla, paleta);
	if (acento) colores.acento = acento;
	for (const [k, x] of Object.entries(e.colores ?? {})) {
		const c = colorValido(x);
		if (c) colores[k as Rol | Zona] = c;
	}
	// El encabezado y el pie siguen al fondo y al texto si no tienen color propio elegido.
	if (!colorValido(e.colores?.encabezado) && colorValido(e.colores?.fondo))
		colores.encabezado = colores.fondo;
	if (!colorValido(e.colores?.encabezadoTexto) && colorValido(e.colores?.texto))
		colores.encabezadoTexto = colores.texto;
	// Las zonas como en el handoff, si la tienda no eligió color: Mostrador y Pedido llevan el
	// encabezado en tarjeta (blanco); el pie de Vidriera con columnas o nombre grande es una banda
	// oscura, y el de Mostrador y Pedido va en tarjeta.
	const pie = de(e.pie, ['simple', 'columnas', 'grande'], p?.pie ?? base.pie);
	const propio = (z: Zona) => !!colorValido(e.colores?.[z]);
	if (direccion !== 'vidriera' && !propio('encabezado') && colores.encabezado === colores.fondo) {
		colores.encabezado = colores.superficie;
		if (!propio('encabezadoTexto')) colores.encabezadoTexto = colores.texto;
	}
	if (!propio('pie') && colores.pie === colores.fondo) {
		if (direccion === 'vidriera' && pie !== 'simple') {
			colores.pie = esOscuro(colores.texto) ? colores.texto : '#16150F';
			if (!propio('pieTexto')) colores.pieTexto = colores.fondo;
		} else if (direccion !== 'vidriera' && pie !== 'simple') {
			colores.pie = colores.superficie;
			if (!propio('pieTexto')) colores.pieTexto = colores.texto;
		}
	}
	const fuente = (x: string | undefined, si: string) => (x && FUENTES[x] ? x : si);
	const mensajes = (e.barra?.mensajes ?? [])
		.map((m) => ({
			texto: (m.texto ?? '').trim(),
			...(m.link ? { link: m.link } : {}),
			...(m.hasta && Number.isFinite(Date.parse(m.hasta)) ? { hasta: m.hasta } : {})
		}))
		.filter((m) => m.texto)
		.slice(0, 5);
	return {
		direccion,
		forma: de(e.forma, ['recta', 'redondeada', 'pildora'], p?.forma ?? base.forma),
		densidad: de(e.densidad, ['aireada', 'normal', 'compacta'], p?.densidad ?? base.densidad),
		ancho: de(e.ancho, ['normal', 'amplio'], 'normal'),
		fotoProducto: de(
			e.fotoProducto,
			['cuadrada', 'vertical', 'apaisada'],
			p?.fotoProducto ?? base.fotoProducto
		),
		encabezado: de(
			e.encabezado,
			['izquierda', 'centrado', 'buscador'],
			p?.encabezado ?? base.encabezado
		),
		pie,
		fuenteTitulos: fuente(e.fuenteTitulos, p?.fuenteTitulos ?? base.titulos),
		fuenteTexto: fuente(e.fuenteTexto, p?.fuenteTexto ?? base.texto),
		colores,
		barra: mensajes.length
			? {
					variante: e.barra?.variante === 'rotativa' ? 'rotativa' : 'fija',
					mensajes,
					fondo: e.barra?.fondo ?? 'oscuro'
				}
			: null,
		barraAuto: e.barra?.automaticos !== false,
		whatsappFlotante: e.whatsappFlotante !== false,
		gastronomia: preset.gastronomia
	};
}
