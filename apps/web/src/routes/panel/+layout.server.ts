import { env } from 'cloudflare:workers';
import { datosDe, usaSandbox } from '#lib/server/datos/index.ts';
import { plantilla } from '#lib/plantillas.ts';
import type { PaginaNotificaciones } from '#lib/server/datos/tipos.ts';
import { unsplash } from '#lib/landing/formato.ts';
import type { LayoutServerLoad } from './$types';

/** El logo de Mi tienda (una foto subida, que es una URL, o un id de Unsplash), listo para un <img>. */
const logoTienda = (logo: string | null | undefined) => (logo ? unsplash(logo, 120) : null);

/**
 * El armazón del panel: la tienda, el resumen (insignias, notificaciones, prueba) y las preferencias
 * de la persona (tema y modo, en cookies que escribe el navegador, así el primer render ya sale bien).
 */
export const load: LayoutServerLoad = async ({ locals, cookies, depends }) => {
	// Tiempo real (§44): las insignias, la campanita (§45) y el plan se recargan con sus avisos.
	depends('tienda:resumen', 'tienda:armazon', 'tienda:plan', 'tienda:notificaciones');
	const datos = datosDe(locals);
	const [tienda, resumen, notificaciones, ajustes, acceso] = await Promise.all([
		datos.tienda(),
		datos.resumen(),
		// Si la API no responde la campanita, el panel anda igual (la campanita queda vacía).
		datos
			.notificaciones()
			.catch(() => ({ notificaciones: [], sinLeer: 0, siguiente: null }) as PaginaNotificaciones),
		// Si la API todavía no responde turnos, la sección queda como la del rubro.
		datos.ajustesTurnos().catch(() => null),
		// El rol de la persona y sus tiendas: el menú muestra sólo lo que puede (Go igual responde 403).
		datos.acceso()
	]);
	return {
		tienda,
		resumen,
		notificaciones,
		acceso,
		// El logo que se cargó en Mi tienda; sin logo, sus iniciales (nunca la foto de stock de la
		// plantilla). En la demo, la portada de su plantilla.
		fotoTienda: usaSandbox()
			? (plantilla(tienda.plantilla)?.img ?? null)
			: logoTienda(acceso.tiendas.find((t) => t.id === acceso.tiendaId)?.logo),
		demo: usaSandbox(),
		// Alguien del equipo de Tu Local Web viendo la tienda en sólo lectura (aviso fijo).
		mirada: locals.mirada,
		// La sección Turnos: con turnos activos o si la tienda es de servicios.
		turnos: !!ajustes?.activo || tienda.rubro === 'servicios',
		tema: cookies.get('panel_tema') === 'dark' ? ('dark' as const) : ('light' as const),
		modo: cookies.get('panel_modo') === 'simple' ? ('simple' as const) : ('completo' as const),
		// El tiempo real sólo existe con la API (la demo no tiene servidor de avisos). El origen va
		// vacío en producción (mismo sitio, así viaja la cookie); en las pruebas, el de la API.
		tiempoReal: usaSandbox() ? null : { origen: String(env.TIEMPO_REAL_ORIGEN ?? '') || null }
	};
};
