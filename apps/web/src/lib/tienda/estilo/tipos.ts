/** Los roles de color que la tienda puede cambiar en Mi tienda › Colores. */
export type Rol =
	'fondo' | 'superficie' | 'texto' | 'textoSuave' | 'borde' | 'acento' | 'sobreAcento';

/** Un tono de la plantilla: mezcla en OKLab de dos roles (p % del primero) y su transparencia (0–1). */
export type Tono = { a: Rol | 'ok'; b: Rol | null; p: number; alfa: number | null };

/** Lo que scripts/estilo/generar.ts sabe de cada plantilla (mapas.ts). */
export type MapaEstilo = {
	roles: Record<Rol, string>;
	/** El verde de «en stock» y del precio con transferencia (rol «ok»: se ajusta solo al fondo). */
	ok: string | null;
	/** El fondo del encabezado si no sale del fondo de la página (Pampa); null: sale del fondo. */
	encabezado: string | null;
	/** El pie como banda propia (fondo y texto originales); null: va con los colores de la página. */
	pie: { fondo: string; texto: string } | null;
	fuentes: { titulos: string | null; texto: string };
	tonos: Record<string, Tono>;
};
