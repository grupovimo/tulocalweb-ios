import { avisarStock } from '#lib/server/tienda/lanzamientos.ts';
import type { RequestHandler } from './$types';

/** «Avisame cuando vuelva»: el email (y el WhatsApp) de quien espera un producto sin stock (§60). */
export const POST: RequestHandler = (event) => avisarStock(event);
