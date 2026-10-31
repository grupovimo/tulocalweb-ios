import { senar } from '#lib/server/tienda/lanzamientos.ts';
import type { RequestHandler } from './$types';

/** Reservar un lanzamiento o encargar con seña (§52): devuelve adónde pagar la seña. */
export const POST: RequestHandler = (event) => senar(event);
