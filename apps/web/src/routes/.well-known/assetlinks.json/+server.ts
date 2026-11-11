import type { RequestHandler } from './$types';
import { assetlinks, respuestaJSON } from '#lib/enlaces.ts';

/** App Links de Android: la app que puede abrir los links de tulocalweb.com.ar (apps/pruebas/enlaces.json). */
export const GET: RequestHandler = () => respuestaJSON(assetlinks());
