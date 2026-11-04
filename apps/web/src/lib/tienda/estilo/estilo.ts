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
};

/** El título de una sección: el propio o el de la plantilla. */
export function tituloSeccion(
	estilo: EstiloTienda | null | undefined,
	seccion: SeccionConTitulo,
	original: string
): string {
	return estilo?.textos?.[seccion]?.trim() || original;
}
