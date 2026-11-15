import type { ItemMenu } from './menu.ts';
import type { Rol } from './tipos.ts';

/** Zonas con colores propios además de los roles de la página. */
export type Zona = 'encabezado' | 'encabezadoTexto' | 'pie' | 'pieTexto';
export const ZONAS: Zona[] = ['encabezado', 'encabezadoTexto', 'pie', 'pieTexto'];

/** Las secciones de la portada que llevan título propio. */
export type SeccionConTitulo = 'cats' | 'prods' | 'revs' | 'ig' | 'wa';
export const CON_TITULO: SeccionConTitulo[] = ['cats', 'prods', 'revs', 'ig', 'wa'];

export type Redes = {
	instagram?: string;
	tiktok?: string;
	facebook?: string;
	youtube?: string;
	email?: string;
	telefono?: string;
};
export const REDES: (keyof Redes)[] = [
	'instagram',
	'tiktok',
	'facebook',
	'youtube',
	'email',
	'telefono'
];

/**
 * Mi tienda › Estilo (components/schemas/EstiloTienda). Todo opcional: lo vacío es lo de la
 * plantilla. Igual en la web, en Go (mitienda.Estilo) y en la app de iPhone (EstiloTienda.swift).
 */
export type EstiloTienda = {
	/** #RRGGBB por rol y por zona. */
	colores?: Partial<Record<Rol | Zona, string>>;
	/** Ids de lib/tienda/estilo/fuentes.ts. */
	fuenteTitulos?: string;
	fuenteTexto?: string;
	/** El logo para fondos oscuros (URL). */
	logoOscuro?: string;
	/** El ícono de la pestaña y del acceso directo (URL, cuadrado). */
	icono?: string;
	/** Alto del logo en la cabecera, en px (24–72; 0 o ausente: el de la plantilla). */
	logoAlto?: number;
	/** La barra de arriba de todo (vacía: no hay). */
	anuncio?: string;
	anuncioLink?: string;
	/** Una línea propia en el pie. */
	leyendaPie?: string;
	/** El título de cada sección de la portada (vacío: el de la plantilla). */
	textos?: Partial<Record<SeccionConTitulo, string>>;
	redes?: Redes;
	/** Las pestañas del menú (vacío: las de la plantilla). lib/tienda/estilo/menu.ts. */
	menu?: ItemMenu[];
	// ---- Diseño v2 (docs/especificaciones/tienda-bloques.md). Vacío: lo de la dirección y el preset.
	direccion?: 'vidriera' | 'mostrador' | 'pedido';
	forma?: 'recta' | 'redondeada' | 'pildora';
	densidad?: 'aireada' | 'normal' | 'compacta';
	ancho?: 'normal' | 'amplio';
	fotoProducto?: 'cuadrada' | 'vertical' | 'apaisada';
	encabezado?: 'izquierda' | 'centrado' | 'buscador';
	pie?: 'simple' | 'columnas' | 'grande';
	/** La barra de anuncios de arriba de todo (en v1 era `anuncio` y `anuncioLink`). */
	barra?: {
		/** Vacío: fija. */
		variante?: 'fija' | 'rotativa';
		/** `hasta` (ISO): cuenta regresiva en el mensaje, que se esconde al vencer. */
		mensajes: { texto: string; link?: string; hasta?: string }[];
		fondo?: 'normal' | 'superficie' | 'acento' | 'oscuro';
	};
	/** false: sin el botón flotante de WhatsApp (la sección «wa» oculta de v1). */
	whatsappFlotante?: boolean;
	/** Las herramientas para vender más (bloques.json › conversion; lo que falta, su porDefecto). */
	conversion?: ConversionGuardada;
};

/** estilo.conversion como se guarda (todo opcional): sistema/promos.ts › conversionDe lo completa. */
export type ConversionGuardada = {
	envioGratis?: boolean;
	pruebaSocial?: { activa?: boolean; cada?: number; paginas?: string[] };
	etiquetas?: { nuevo?: boolean; oferta?: boolean; ultimas?: boolean; masVendido?: boolean };
	stockBajo?: number;
	compraFija?: boolean;
	sumaEnCarrito?: boolean;
	favoritos?: boolean;
};

/** El título de una sección: el propio o el de la plantilla. */
export function tituloSeccion(
	estilo: EstiloTienda | null | undefined,
	seccion: SeccionConTitulo,
	original: string
): string {
	return estilo?.textos?.[seccion]?.trim() || original;
}
