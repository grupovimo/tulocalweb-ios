import { avisar } from '#lib/server/tienda/lanzamientos.ts';
import type { RequestHandler } from './$types';

/** «Avisame cuando salga»: el email para el aviso del lanzamiento (§52). */
export const POST: RequestHandler = (event) => avisar(event);
