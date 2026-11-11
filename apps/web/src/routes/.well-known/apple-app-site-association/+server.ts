import type { RequestHandler } from './$types';
import { aasa, respuestaJSON } from '#lib/enlaces.ts';

/** Universal Links de la app de iPhone: qué links de tulocalweb.com.ar abren la app (apps/pruebas/enlaces.json). */
export const GET: RequestHandler = () => respuestaJSON(aasa());
