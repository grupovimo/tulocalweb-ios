/**
 * Números argentinos en la tienda (regla 9 del handoff): `$ 186.000`, `4,8`, `3 cuotas de $ 62.000`.
 * Reemplaza a los `fmt` copiados en cada plantilla. No depende de Intl: en Workers y en Node da
 * siempre lo mismo.
 */

/** Miles con punto: 186000 → «186.000». */
export function miles(n: number): string {
	const s = String(Math.round(Math.abs(n)));
	const con = s.replace(/\B(?=(\d{3})+(?!\d))/g, '.');
	return n < 0 && Math.round(Math.abs(n)) > 0 ? `−${con}` : con;
}

/** Un precio en pesos enteros: «$ 186.000» (con espacio duro: nunca se corta). */
export const precio = (n: number) => `$ ${miles(n)}`;

/** Un promedio con un decimal y coma: 4,8 (5 → «5,0»). */
export function promedio(n: number): string {
	return (Math.round(n * 10) / 10).toFixed(1).replace('.', ',');
}

/** El valor de cada cuota (redondeado al peso). */
export const cuota = (total: number, n: number) => Math.round(total / Math.max(1, n));

/** «3 cuotas de $ 62.000» o «6 cuotas sin interés de $ 31.000». Vacío si no hay cuotas. */
export function cuotas(total: number, n: number, sinInteres = false): string {
	if (!(n > 1) || !(total > 0)) return '';
	return `${n} cuotas ${sinInteres ? 'sin interés ' : ''}de ${precio(cuota(total, n))}`;
}

/** El precio pagando con transferencia (pct de descuento). */
export const conTransferencia = (total: number, pct: number) =>
	Math.round(total * (1 - Math.max(0, Math.min(100, pct)) / 100));

/** El descuento de una oferta: «−25 %» (vacío si no hay o no llega al 1 %). */
export function descuento(actual: number, anterior: number | null | undefined): string {
	if (!anterior || anterior <= actual) return '';
	const pct = Math.round((1 - actual / anterior) * 100);
	return pct >= 1 ? `−${pct} %` : '';
}

/** «1 producto», «8 productos». */
export const cuantos = (n: number, uno: string, varios: string) =>
	`${miles(n)} ${n === 1 ? uno : varios}`;

/**
 * La URL de una foto con el ancho pedido. Las de la demo son ids de Unsplash; las subidas, URLs.
 * null: no hay foto (el placeholder lo dibuja Foto.svelte).
 */
export function urlFoto(f: string | null | undefined, ancho: number, alto?: number): string | null {
	if (!f) return null;
	if (/^https?:\/\//.test(f) || f.startsWith('/')) return f;
	return `https://images.unsplash.com/photo-${f}?auto=format&fit=crop&w=${ancho}${alto ? `&h=${alto}` : ''}&q=72`;
}

/**
 * El texto con la última palabra pegada a la anterior (regla 7: sin renglones viudos). Se usa en
 * títulos y frases cortas; los textos largos van con `text-wrap: pretty`.
 */
export function sinViuda(t: string): string {
	const s = t.trim().replace(/\s+/g, ' ');
	const i = s.lastIndexOf(' ');
	if (i <= 0) return s;
	// Si la última palabra es larga y la anterior también, igual conviene pegarlas: el título no
	// queda con una palabra sola. Sólo se evita en textos de una línea muy corta.
	return `${s.slice(0, i)} ${s.slice(i + 1)}`;
}
