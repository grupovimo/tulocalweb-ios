/**
 * El píxel de Meta de Tu Local Web, sólo en las páginas de los anuncios (/l/) y sólo si la persona
 * aceptó las cookies de marketing (#lib/consentimiento.svelte.ts). Si acepta después de entrar, se
 * carga en ese momento. El alta se mide además desde el servidor (#lib/server/meta.ts).
 */
import { leerEleccion } from '#lib/consentimiento.svelte.ts';

type Fbq = ((...a: unknown[]) => void) & Record<string, unknown>;
type Ventana = Window & { fbq?: Fbq; _fbq?: unknown };

let pixel: string | null = null;
let cargado = false;
const pendientes: [string, Record<string, unknown>][] = [];

const permitido = () => leerEleccion()?.mk === true;

function cargar() {
	if (cargado || !pixel || !permitido()) return;
	cargado = true;
	const w = window as Ventana;
	if (!w.fbq) {
		// El arranque oficial de fbevents.js: una cola hasta que carga el script.
		const n = function (...a: unknown[]) {
			if (n.callMethod) (n.callMethod as (...x: unknown[]) => void)(...a);
			else (n.queue as unknown[]).push(a);
		} as Fbq;
		n.push = n;
		n.loaded = true;
		n.version = '2.0';
		n.queue = [];
		w.fbq = n;
		w._fbq ??= n;
		const s = document.createElement('script');
		s.async = true;
		s.src = 'https://connect.facebook.net/en_US/fbevents.js';
		document.head.appendChild(s);
	}
	w.fbq('init', pixel);
	w.fbq('track', 'PageView');
	for (const [nombre, datos] of pendientes.splice(0)) w.fbq('track', nombre, datos);
}

/** Arranca en la página: con el ID vacío (sin cargar en el servidor) no hace nada. */
export function iniciarPixel(id: string | null) {
	if (!id || typeof window === 'undefined') return () => {};
	pixel = id;
	cargar();
	const alElegir = () => cargar();
	window.addEventListener('mostrador:cookies', alElegir);
	return () => window.removeEventListener('mostrador:cookies', alElegir);
}

/** Un evento estándar. Sin permiso no sale; si el script todavía no cargó, espera en la cola. */
export function evento(nombre: 'ViewContent' | 'Lead', datos: Record<string, unknown> = {}) {
	if (!pixel || !permitido()) return;
	const w = window as Ventana;
	if (cargado && w.fbq) w.fbq('track', nombre, datos);
	else pendientes.push([nombre, datos]);
}
