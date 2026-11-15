/**
 * Logos que se leen sobre cualquier fondo (bloque «logos»): se dibujan con sus colores, a pleno, y
 * sólo cuando uno no contrasta con el fondo en el que está (un logo negro sobre una tienda oscura,
 * uno blanco sobre una clara) se lo pasa entero a claro u oscuro. Los que traen su propio fondo
 * (sin transparencia) no se tocan: ya tienen con qué leerse.
 */

/** La luminancia relativa (WCAG) de un color en 0–255. */
export function luz(r: number, g: number, b: number): number {
	const c = [r, g, b].map((v) => {
		const x = v / 255;
		return x <= 0.03928 ? x / 12.92 : ((x + 0.055) / 1.055) ** 2.4;
	});
	return 0.2126 * c[0] + 0.7152 * c[1] + 0.0722 * c[2];
}

const contrasteDe = (a: number, b: number) => (Math.max(a, b) + 0.05) / (Math.min(a, b) + 0.05);

/**
 * Cómo pintar un logo: `null` con sus colores; 'claro' u 'oscuro' para pasarlo entero a ese tono.
 * `pixeles` son RGBA (de un canvas chico); `fondo`, la luminancia del fondo detrás.
 */
export function tonoLogo(pixeles: Uint8ClampedArray, fondo: number): 'claro' | 'oscuro' | null {
	let suma = 0;
	let peso = 0;
	let transparentes = 0;
	for (let i = 0; i < pixeles.length; i += 4) {
		const a = pixeles[i + 3] / 255;
		if (a < 0.5) transparentes++;
		if (a < 0.1) continue;
		suma += luz(pixeles[i], pixeles[i + 1], pixeles[i + 2]) * a;
		peso += a;
	}
	const total = pixeles.length / 4;
	// Sin transparencia (un JPG, un logo en su recuadro) o vacío: se deja como está.
	if (!peso || transparentes < total * 0.05) return null;
	const media = suma / peso;
	if (contrasteDe(media, fondo) >= 3) return null;
	return fondo < 0.18 ? 'claro' : 'oscuro';
}

/** La luminancia del primer fondo pintado detrás de un elemento (blanco si no hay ninguno). */
export function luzDelFondo(el: Element): number {
	for (let x: Element | null = el; x; x = x.parentElement) {
		const m = getComputedStyle(x).backgroundColor.match(/[\d.]+/g);
		if (!m) continue;
		const [r, g, b, a = '1'] = m;
		if (Number(a) > 0.5) return luz(Number(r), Number(g), Number(b));
	}
	return 1;
}

/**
 * Acción para el <img> de un logo: al cargar, lo mide en un canvas de 24 px y le pone
 * `data-tono` si hace falta. Si la foto viene de otro origen sin CORS, no se puede medir y queda
 * con sus colores.
 */
export function medirLogo(img: HTMLImageElement) {
	let vivo = true;
	const medir = () => {
		const src = img.currentSrc || img.src;
		if (!src || !vivo) return;
		const otra = new Image();
		if (new URL(src, location.href).origin !== location.origin) otra.crossOrigin = 'anonymous';
		otra.onload = () => {
			if (!vivo) return;
			try {
				const c = document.createElement('canvas');
				c.width = c.height = 24;
				const cx = c.getContext('2d', { willReadFrequently: true });
				if (!cx) return;
				cx.drawImage(otra, 0, 0, 24, 24);
				const tono = tonoLogo(cx.getImageData(0, 0, 24, 24).data, luzDelFondo(img));
				if (tono) img.dataset.tono = tono;
				else delete img.dataset.tono;
			} catch {
				// Canvas manchado por otro origen: queda con sus colores.
			}
		};
		otra.src = src;
	};
	if (img.complete) medir();
	img.addEventListener('load', medir);
	return {
		destroy() {
			vivo = false;
			img.removeEventListener('load', medir);
		}
	};
}
