import { manejar } from '../participar.ts';
import type { RequestHandler } from './$types';

/** POST { ventana, email, nombre?, whatsapp?, consentimiento } → PromoResultado. */
export const POST: RequestHandler = (event) => manejar(event, 'captar');
