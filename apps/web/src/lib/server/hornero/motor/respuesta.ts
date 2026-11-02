/**
 * Lo que responde Tincho: un texto corto (con **negritas**) y, cuando corresponde, bloques para ver
 * los datos (cifras, gráficos, tablas, listas, una barra de progreso), botones que llevan a la
 * sección del panel y preguntas para seguir. La app y el panel dibujan los bloques; Telegram y los
 * clientes viejos usan sólo el texto.
 */

/** Adónde lleva un botón (la app y la web lo traducen a su pantalla). */
export type Destino =
	| 'pedidos'
	| 'preparar'
	| 'transferencias'
	| 'productos'
	| 'sin-stock'
	| 'nuevo-producto'
	| 'importar'
	| 'clientes'
	| 'cupones'
	| 'carritos'
	| 'compartir'
	| 'estadisticas'
	| 'integraciones'
	| 'mi-tienda'
	| 'resenas'
	| 'negocio'
	| 'caja'
	| 'finanzas'
	| 'facturacion'
	| 'equipo'
	| 'plan'
	| 'links'
	| 'devoluciones'
	| 'venta';

export type Tono = 'ok' | 'mal' | 'aviso' | 'neutro';
export type Unidad = 'pesos' | 'pedidos' | 'unidades' | 'visitas' | 'numero';

export type Bloque =
	| { tipo: 'cifras'; items: { titulo: string; valor: string; detalle?: string; tono?: Tono }[] }
	| {
			tipo: 'barras';
			titulo: string;
			unidad: Unidad;
			items: { etiqueta: string; valor: number; destacado?: boolean }[];
	  }
	| {
			tipo: 'linea';
			titulo: string;
			unidad: Unidad;
			puntos: { etiqueta: string; valor: number }[];
			/** El período anterior, para comparar (misma cantidad de puntos). */
			anterior?: number[];
	  }
	| { tipo: 'tabla'; titulo: string; columnas: string[]; filas: string[][] }
	| {
			tipo: 'lista';
			titulo: string;
			items: { titulo: string; detalle?: string; insignia?: string; tono?: Tono }[];
	  }
	| {
			tipo: 'progreso';
			titulo: string;
			actual: number;
			objetivo: number;
			unidad: Unidad;
			detalle?: string;
	  };

export type Accion = { texto: string; ir: Destino };

export type Respuesta = {
	texto: string;
	bloques?: Bloque[];
	acciones?: Accion[];
	sugerencias?: string[];
	/** Lo que entendió (para seguir la charla: «¿y ayer?»). */
	tema?: string;
};

/** Con espacio fino que no se corta, como en todo el panel: «$ 12.345». */
export const pesos = (n: number) => '$\u00a0' + Math.round(n).toLocaleString('es-AR');
export const num = (n: number) => Math.round(n).toLocaleString('es-AR');
export const plural = (n: number, uno: string, varios: string) =>
	`${num(n)} ${n === 1 ? uno : varios}`;

/** «, un 12 % más que …» o vacío si no hay con qué comparar. */
export function variacion(actual: number, anterior: number, contra = 'el período anterior') {
	if (!anterior) return '';
	const pct = Math.round(((actual - anterior) / anterior) * 100);
	if (pct === 0) return `, igual que ${contra}`;
	return pct > 0
		? `, un ${pct}\u202f% más que ${contra}`
		: `, un ${-pct}\u202f% menos que ${contra}`;
}

/** La variación como detalle de una cifra: «▲ 12 %» y su tono. */
export function delta(actual: number, anterior: number): { detalle?: string; tono?: Tono } {
	if (!anterior) return {};
	const pct = Math.round(((actual - anterior) / anterior) * 100);
	return {
		detalle: pct === 0 ? 'igual' : pct > 0 ? `▲ ${pct}\u202f%` : `▼ ${-pct}\u202f%`,
		tono: pct > 0 ? 'ok' : pct < 0 ? 'mal' : 'neutro'
	};
}

const TZ = 'America/Argentina/Buenos_Aires';
/** «lun 28», para los ejes. */
export const diaCorto = (iso: string) =>
	new Date(iso + 'T12:00:00Z')
		.toLocaleDateString('es-AR', { timeZone: TZ, weekday: 'short', day: 'numeric' })
		.replace('.', '')
		.replace(',', '');
/** «lunes 28 de octubre». */
export const diaLargo = (iso: string) =>
	new Date(iso + 'T12:00:00Z').toLocaleDateString('es-AR', {
		timeZone: TZ,
		weekday: 'long',
		day: 'numeric',
		month: 'long'
	});

/** «A, B y C» (y «y N más» si hay más). */
export function enumerar(xs: string[], max = 5) {
	const v = xs.slice(0, max);
	const resto = xs.length - v.length;
	if (resto > 0) return `${v.join(', ')} y ${resto} más`;
	return v.length > 1 ? `${v.slice(0, -1).join(', ')} y ${v.at(-1)}` : (v[0] ?? '');
}
