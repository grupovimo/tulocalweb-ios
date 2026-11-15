import { datosDe } from '#lib/server/datos/index.ts';
import type { PageServerLoad } from './$types';

/** Crecer › Newsletter (§50): suscriptos, envíos y los productos para destacar. */
export const load: PageServerLoad = async ({ locals }) => {
	const datos = datosDe(locals);
	const [newsletter, mi] = await Promise.all([
		datos.newsletter(),
		datos.miTienda().catch(() => null)
	]);
	// Los nombres de las ventanas, para decir de cuál vino cada contacto captado.
	const d = mi?.diseno as { ventanas?: { id: string; nombre: string }[] } | null | undefined;
	const ventanas = Object.fromEntries((d?.ventanas ?? []).map((w) => [w.id, w.nombre]));
	return { newsletter, ventanas };
};
