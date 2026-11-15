import { manejar } from '../participar.ts';
import type { RequestHandler } from './$types';

/** POST { ventana, email, nombre?, whatsapp?, consentimiento } → PromoResultado con porcion y texto. */
export const POST: RequestHandler = (event) => manejar(event, 'ruleta');
