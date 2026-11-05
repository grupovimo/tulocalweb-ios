import { datosDe } from '#lib/server/datos/index.ts';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals, depends, url }) => {
	// Tiempo real (§44): se recarga con los avisos de lo que muestra.
	depends('tienda:plan');
	const datos = datosDe(locals);
	// La vuelta de Mercado Pago (§56): se confirma el pago de la diferencia o el débito sin esperar el
	// aviso, así la persona vuelve y ya ve su plan nuevo. Si falla, queda el aviso del webhook.
	const pago = url.searchParams.get('payment_id') ?? url.searchParams.get('collection_id');
	const suscripcion = url.searchParams.get('preapproval_id');
	const ok = (v: string | null) => (v && /^[\w-]{1,64}$/.test(v) ? v : undefined);
	if (ok(pago) || ok(suscripcion))
		await datos
			.volverAbono({ pago: ok(pago), suscripcion: ok(suscripcion) })
			.catch(() => undefined);
	return datos.planCuenta();
};
