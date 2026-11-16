/**
 * El carrito como panel al costado (paginas.carrito.variante = 'panel'): agregar desde la ficha o
 * tocar la bolsa del encabezado lo abre sin salir de la página. Uno solo por pestaña.
 */
class PanelCarrito {
	abierto = $state(false);
	/** El panel está en la página (marco/PanelCarrito lo marca al montarse). */
	montado = false;
}

export const panelCarrito = new PanelCarrito();
