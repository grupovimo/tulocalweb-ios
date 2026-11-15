/**
 * Los favoritos de quien compra (estilo.conversion.favoritos): ids de producto por tienda, en el
 * dispositivo (localStorage con try/catch: en modo privado quedan sólo mientras dura la pestaña).
 * Uno por tienda y por pestaña, como el carrito.
 */
const MAX = 60;

export class Favoritos {
	ids = $state<string[]>([]);
	#clave: string;

	constructor(slug: string) {
		this.#clave = `tlw-favoritos:${slug}`;
		if (typeof window === 'undefined') return;
		// Después de hidratar: el servidor dibujó sin favoritos y el primer dibujo tiene que coincidir.
		queueMicrotask(() => {
			try {
				const x = JSON.parse(localStorage.getItem(this.#clave) ?? '[]');
				if (Array.isArray(x)) this.ids = x.filter((i) => typeof i === 'string').slice(0, MAX);
			} catch {
				/* sin almacenamiento: empieza vacío */
			}
		});
	}

	tiene = (id: string) => this.ids.includes(id);

	alternar(id: string): boolean {
		const esta = this.tiene(id);
		this.ids = esta ? this.ids.filter((x) => x !== id) : [id, ...this.ids].slice(0, MAX);
		try {
			localStorage.setItem(this.#clave, JSON.stringify(this.ids));
		} catch {
			/* sin almacenamiento: quedan hasta cerrar la pestaña */
		}
		return !esta;
	}
}

const porTienda: Record<string, Favoritos> = {};
export function favoritosDe(slug: string): Favoritos {
	let f = porTienda[slug];
	if (!f) {
		f = new Favoritos(slug);
		// En el servidor no se guarda: cada pedido arranca vacío.
		if (typeof window !== 'undefined') porTienda[slug] = f;
	}
	return f;
}
