import { error, fail } from '@sveltejs/kit';
import { usaSandbox } from '#lib/server/datos/index.ts';
import { cambiar, suscripcionDe, type Accion } from '#lib/server/tienda/suscripciones.ts';
import type { Actions, PageServerLoad } from './$types';

/**
 * El link privado de quien se suscribió (§51): ver la suscripción, pausarla, reanudarla o darla de
 * baja, sin cuenta. El código es aleatorio (no se adivina el de otro). Mercado Pago vuelve acá
 * después de autorizar el débito (back_url, con preapproval_id).
 */
export const load: PageServerLoad = async ({ params, url }) => {
	const suscripcion = await suscripcionDe(params.tienda, params.codigo, usaSandbox());
	if (!suscripcion) error(404, 'No encontramos esa suscripción.');
	const alta = url.searchParams.has('alta') || url.searchParams.has('preapproval_id');
	return { suscripcion, aviso: alta ? 'alta' : null, demo: usaSandbox() };
};

const accion =
	(a: Accion) =>
	async ({ params }: { params: { tienda: string; codigo: string } }) => {
		const r = await cambiar(params.tienda, params.codigo, a, usaSandbox());
		if (r === 'no-existe') error(404, 'No encontramos esa suscripción.');
		if (r === 'estado')
			return fail(409, { error: 'La suscripción cambió mientras tanto. Recargá la página.' });
		if (r === 'mercadopago')
			return fail(502, {
				error: 'Mercado Pago no respondió y no cambió nada. Probá de nuevo en un rato.'
			});
		return { hecho: a };
	};

export const actions: Actions = {
	pausar: accion('pausar'),
	reanudar: accion('reanudar'),
	cancelar: accion('cancelar')
};
