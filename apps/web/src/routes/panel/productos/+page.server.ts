import { datosDe } from '#lib/server/datos/index.ts';
import { leer } from '#lib/panel/listas/filtros.ts';
import type { PageServerLoad } from './$types';

/**
 * La primera ventana de la lista con los filtros de la URL, y el recargo de Mercado Libre para el
 * editor: sólo si está conectado (con la API, la lista «ml» aparece sólo con la cuenta conectada).
 */
export const load: PageServerLoad = async ({ locals, url, depends }) => {
	// Tiempo real (§44): stock, altas, cambios de otra persona del equipo e importaciones.
	depends('tienda:productos');
	const datos = datosDe(locals);
	const [lista, precios, mayorista] = await Promise.all([
		datos.listar('productos', { ...leer(url.searchParams), desde: 0, cuantos: 60 }),
		// Sin permiso para ver precios (o si falla), el editor sigue sin la fila de Mercado Libre.
		datos.listasPrecio().catch(() => ({ listas: [] as { id: string; ajuste: number }[] })),
		// El mínimo mayorista (§49) para la caja «Mayorista» del editor; si falla, sin la caja.
		datos
			.configuracion()
			.then((c) => c.mayorista)
			.catch(() => null)
	]);
	return {
		lista,
		recargoML: precios.listas.find((l) => l.id === 'ml')?.ajuste ?? null,
		mayorista
	};
};
