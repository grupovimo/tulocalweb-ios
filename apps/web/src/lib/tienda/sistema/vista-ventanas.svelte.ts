/**
 * La vista previa de Mi tienda y las ventanas: en la vista no salen solas (molestarían mientras se
 * edita) ni se miden; el editor abre una con `{ tipo: 'ventana', id }` (las apps, con
 * `window.tiendaVista.ventana(id)`) y se ve con el borrador hasta que se cierra o se pide otra.
 */
class VistaVentanas {
	/** Es la vista previa del editor. */
	vista = $state(false);
	/** La ventana que pidió el editor (null: ninguna). */
	forzada = $state<string | null>(null);
	/** Cambia en cada pedido: la misma ventana pedida dos veces se vuelve a abrir. */
	vez = $state(0);

	abrir(id: string | null) {
		this.forzada = id;
		this.vez++;
	}
}

export const vistaVentanas = new VistaVentanas();
