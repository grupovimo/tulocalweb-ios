/**
 * El comparador (función propia de Andes): hasta 3 productos de la tienda, recordados en este
 * navegador. Lo usan el botón «Comparar» de cada producto (marco/BotonComparar) y el panel
 * (marco/Comparador), que se abre sobre cualquier página.
 */
const MAX = 3;

export class Comparar {
	ids = $state<string[]>([]);
	abierto = $state(false);

	constructor(private clave: string) {
		try {
			const g = JSON.parse(localStorage.getItem(clave) ?? '[]');
			this.ids = Array.isArray(g) ? g.filter((x) => typeof x === 'string').slice(0, MAX) : [];
		} catch {
			this.ids = [];
		}
	}

	tiene = (id: string) => this.ids.includes(id);

	alternar(id: string) {
		this.ids = this.tiene(id) ? this.ids.filter((x) => x !== id) : [...this.ids, id].slice(-MAX);
		try {
			localStorage.setItem(this.clave, JSON.stringify(this.ids));
		} catch {
			// Sin almacenamiento (modo privado): la comparación dura lo que la página.
		}
	}
}

const porTienda: Record<string, Comparar> = {};

/** El comparador de una tienda (uno por pestaña). En el servidor, uno vacío que no se guarda. */
export function compararDe(slug: string): Comparar {
	if (typeof localStorage === 'undefined') return new Comparar(`andes-comparar:${slug}`);
	return (porTienda[slug] ??= new Comparar(`andes-comparar:${slug}`));
}
