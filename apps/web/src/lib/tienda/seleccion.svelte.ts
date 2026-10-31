/**
 * Lo que se eligió en la ficha de un producto: valores de cada opción, la variante que queda, la
 * foto de la galería y lo que hace "Agregar". Las 13 plantillas usan esto y sólo ponen su marcado
 * (talles, swatches, presentaciones, miniaturas), así el comportamiento es uno solo.
 */
import type { Carrito } from './carrito.svelte.ts';
import { proximamente } from './lanzamiento.ts';
import { mayoristaDeVariante, precioMayorista } from './mayorista.ts';
import type { ProductoPublico, VariantePublica } from './tipos.ts';
import {
	colorDe,
	elegirVariante,
	faltaElegir,
	fotoDelValor,
	precioDe,
	precioDesde,
	textoVariante,
	valoresDisponibles,
	type Valores
} from './variantes.ts';

export type ValorOpcion = {
	valor: string;
	elegido: boolean;
	/** Hay stock con lo demás que ya se eligió (si no, va tachado y deshabilitado). */
	disponible: boolean;
	/** La foto de las variantes con ese valor (para swatches con foto). */
	foto: string | null;
	/** El color, si la opción es Color y el valor uno conocido (para swatches de color). */
	color: string | null;
};

export class Seleccion {
	readonly producto: ProductoPublico;
	elegidos = $state<Valores>({});
	/** Índice de la foto que se ve en la galería. */
	fotoIdx = $state(0);
	/** Se tocó "Agregar" sin elegir todo: las plantillas resaltan lo que falta. */
	aviso = $state(false);

	constructor(producto: ProductoPublico) {
		this.producto = producto;
		// Las opciones con un solo valor ya vienen elegidas.
		const inicial: Valores = {};
		for (const o of this.opcionesBase) if (o.valores.length === 1) inicial[o.nombre] = o.valores[0];
		this.elegidos = inicial;
	}

	get opcionesBase() {
		return this.producto.opciones ?? [];
	}
	get variantes(): VariantePublica[] {
		return this.producto.variantes ?? [];
	}
	/** Tiene opciones para elegir. */
	get conOpciones() {
		return this.opcionesBase.length > 0;
	}

	/** Cada opción con sus valores, cuáles están elegidos y cuáles se pueden comprar. */
	get opciones() {
		return this.opcionesBase.map((o) => {
			const disp = valoresDisponibles(o, this.variantes, this.elegidos);
			return {
				nombre: o.nombre,
				elegido: this.elegidos[o.nombre] ?? null,
				valores: o.valores.map((valor): ValorOpcion => ({
					valor,
					elegido: this.elegidos[o.nombre] === valor,
					disponible: disp[valor],
					foto: fotoDelValor(this.variantes, o.nombre, valor),
					color: colorDe(o.nombre, valor)
				}))
			};
		});
	}

	get variante() {
		return elegirVariante(this.opcionesBase, this.variantes, this.elegidos);
	}
	/** "Talle M · Negro" de lo elegido (vacío si no hay nada). */
	get texto() {
		return textoVariante(this.opcionesBase, this.elegidos);
	}
	/** "Elegí un talle" mientras falte elegir algo; null cuando está todo. */
	get falta() {
		return faltaElegir(this.opcionesBase, this.elegidos);
	}

	/** El precio: el de la variante elegida o, sin elegir, el menor ("desde"). */
	get precio() {
		const v = this.variante;
		if (v) return precioDe(v, this.producto.precio);
		return precioDesde({ precio: this.producto.precio, variantes: this.variantes }).precio;
	}
	/** Sin elegir y con variantes de precios distintos: se muestra "desde". */
	get desde() {
		return (
			!this.variante &&
			precioDesde({ precio: this.producto.precio, variantes: this.variantes }).desde
		);
	}
	/**
	 * El precio mayorista de lo elegido (§49): el de la variante, o el del producto; null si no tiene
	 * o no es menor. Sin elegir, el del producto contra el precio «desde».
	 */
	get mayorista() {
		const v = this.variante;
		const mayor = v
			? mayoristaDeVariante(v.mayorista, this.producto.mayorista)
			: (this.producto.mayorista ?? null);
		const m = precioMayorista(this.precio, mayor);
		return m < this.precio ? m : null;
	}
	/** El tachado sólo cuando el precio es el del producto (las variantes con precio propio no). */
	get anterior() {
		return this.precio === this.producto.precio ? this.producto.anterior : null;
	}
	/** Stock de la variante elegida, o el total. */
	get stock() {
		const v = this.variante;
		if (v) return v.stock;
		// Elegido todo pero sin variante (esa combinación no se vende): no hay stock.
		if (this.conOpciones) return this.falta ? this.variantes.reduce((s, x) => s + x.stock, 0) : 0;
		return this.producto.stock;
	}
	/** Todavía no salió a la venta (lanzamiento, §52): no se agrega al carrito. */
	get proximo() {
		return proximamente(this.producto);
	}
	/** Se puede agregar ya: está todo elegido y hay stock. */
	get listo() {
		return !this.proximo && !this.falta && this.stock > 0 && (!this.conOpciones || !!this.variante);
	}
	/** El texto del botón: lo que falta elegir, sin stock o el "Agregar" de la plantilla. */
	boton(agregar: string, sinStock = 'Sin stock') {
		if (this.proximo) return 'Próximamente';
		if (this.stock <= 0) return sinStock;
		return this.falta ?? agregar;
	}

	/** La galería: las fotos del producto (o la única que tenga). */
	get fotos() {
		const f = this.producto.fotos?.length
			? this.producto.fotos
			: this.producto.foto
				? [this.producto.foto]
				: [];
		return f;
	}
	get foto() {
		return this.fotos[this.fotoIdx] ?? this.fotos[0] ?? null;
	}
	verFoto(i: number) {
		if (i >= 0 && i < this.fotos.length) this.fotoIdx = i;
	}

	/** Elegir (o, si ya estaba elegido, soltar) un valor. La foto sigue a lo elegido. */
	elegir(opcion: string, valor: string) {
		const o = this.opcionesBase.find((x) => x.nombre === opcion);
		if (!o) return;
		const nuevo = { ...this.elegidos };
		if (nuevo[opcion] === valor && o.valores.length > 1) delete nuevo[opcion];
		else nuevo[opcion] = valor;
		this.elegidos = nuevo;
		this.aviso = false;
		const conFoto = this.variantes.find(
			(v) => v.foto && Object.entries(nuevo).every(([k, x]) => v.valores[k] === x)
		);
		if (conFoto?.foto) {
			const i = this.fotos.indexOf(conFoto.foto);
			if (i >= 0) this.fotoIdx = i;
		}
	}

	/**
	 * Suma lo elegido al carrito. Devuelve false si falta elegir (y prende el aviso) o no hay stock.
	 */
	agregar(carrito: Carrito, cantidad = 1) {
		if (this.proximo) return false;
		if (this.falta) {
			this.aviso = true;
			return false;
		}
		if (this.stock <= 0) return false;
		const v = this.variante;
		carrito.sumar(this.producto.id, Math.min(cantidad, this.stock), this.stock, v?.id ?? null);
		return true;
	}
}

/** Datos para la tarjeta del catálogo: precio ("desde" si cambia según la variante) y agotado. */
export function tarjetaDe(p: ProductoPublico) {
	const d = precioDesde({ precio: p.precio, variantes: p.variantes ?? [] });
	const sinStock = p.opciones?.length
		? !(p.variantes ?? []).some((v) => v.stock > 0)
		: p.stock <= 0;
	return { precio: d.precio, desde: d.desde, agotado: sinStock };
}
