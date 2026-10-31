import { datosDe } from '#lib/server/datos/index.ts';
import type { PageServerLoad } from './$types';

/** Crecer › Newsletter (§50): suscriptos, envíos y los productos para destacar. */
export const load: PageServerLoad = async ({ locals }) => ({
	newsletter: await datosDe(locals).newsletter()
});
