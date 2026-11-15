/**
 * El descuento de los combos (bloque «combo» con descuento %): cuando se agregan juntos, las
 * líneas quedan marcadas con el combo y el descuento vale mientras estén todos sus productos. La
 * misma regla corre en el servidor (sandbox: checkout/+page.server.ts; Go: vidriera/combo.go) con
 * los casos de packages/contrato/casos/combos.json.
 */
import type { Bloque, DisenoV2 } from './sistema/catalogo.ts';

export type Combo = { id: string; productos: string[]; descuento: number };

/** Las líneas que mira la regla: con el precio que se cobra de la unidad. */
export type LineaCombo = {
	id: string;
	varianteId: string | null;
	cantidad: number;
	combo?: string | null;
	precio: number;
};

/** Los combos del diseño publicado (portada, ficha y Nosotros) que descuentan algo. */
export function combosDe(d: Pick<DisenoV2, 'inicio' | 'paginas'> | null | undefined): Combo[] {
	if (!d) return [];
	const bloques: Bloque[] = [
		...(d.inicio ?? []),
		...(d.paginas?.producto?.bloques ?? []),
		...(d.paginas?.nosotros?.bloques ?? [])
	];
	return bloques.flatMap((b) => {
		if (b.tipo !== 'combo' || b.oculto) return [];
		const productos = Array.isArray(b.datos?.productos)
			? (b.datos.productos as unknown[]).filter((x): x is string => typeof x === 'string')
			: [];
		const descuento = Number(b.datos?.descuento ?? 0);
		if (productos.length < 2 || !(descuento > 0) || descuento > 50) return [];
		return [{ id: b.id, productos: [...new Set(productos)], descuento: Math.floor(descuento) }];
	});
}

/**
 * El combo del que forma parte un producto, para «Completá el look» en su ficha: el primero que lo
 * incluye entre los bloques visibles de la tienda (portada, ficha y Nosotros), con o sin descuento.
 */
export function comboDelProducto(
	d: Pick<DisenoV2, 'inicio' | 'paginas'> | null | undefined,
	productoId: string
): Bloque | null {
	if (!d) return null;
	const bloques: Bloque[] = [
		...(d.inicio ?? []),
		...(d.paginas?.producto?.bloques ?? []),
		...(d.paginas?.nosotros?.bloques ?? [])
	];
	return (
		bloques.find(
			(b) =>
				b.tipo === 'combo' &&
				!b.oculto &&
				Array.isArray(b.datos?.productos) &&
				(b.datos.productos as unknown[]).includes(productoId)
		) ?? null
	);
}

/** Lo que descuenta cada combo completo del carrito y el total. */
export function descuentoCombos(combos: Combo[], lineas: LineaCombo[]) {
	const porCombo = combos.flatMap((c) => {
		const suyas = lineas.filter((l) => l.combo === c.id && !l.varianteId);
		const de = (id: string) => suyas.find((l) => l.id === id);
		if (!c.productos.every((id) => de(id))) return [];
		const juegos = Math.min(...c.productos.map((id) => de(id)!.cantidad));
		const suma = c.productos.reduce((s, id) => s + de(id)!.precio, 0);
		const descuento = Math.floor((suma * juegos * c.descuento + 50) / 100);
		return descuento > 0 ? [{ id: c.id, juegos, descuento }] : [];
	});
	return { total: porCombo.reduce((s, c) => s + c.descuento, 0), porCombo };
}
