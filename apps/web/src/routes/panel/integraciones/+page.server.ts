import { env } from 'cloudflare:workers';
import { error } from '@sveltejs/kit';
import { IAS, iaDe } from '#lib/panel/integraciones.ts';
import { datosDe, usaSandbox } from '#lib/server/datos/index.ts';
import { conversionesDe } from '#lib/server/conversiones.ts';
import { correoDe } from '#lib/server/correo-argentino.ts';
import { panelApi } from '#lib/server/mcp/api.ts';
import { conexionesDe } from '#lib/server/mcp/oauth.ts';
import { dbSandbox } from '#lib/server/sandbox/sandbox.ts';
import { BOT, habilitado, vinculoApi, vinculoDe } from '#lib/server/telegram.ts';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async (event) => {
	// Tiempo real (§44): se recarga con los avisos de lo que muestra.
	event.depends('tienda:integraciones');
	const { locals, url } = event;
	const d = datosDe(locals);
	const db = locals.sandbox ? dbSandbox(env.SANDBOX) : null;
	const [estado, mp, ml, arca, tg, conexiones, medicion, tienda, correo, conversiones] =
		await Promise.all([
			d.integraciones(),
			d.mercadoPago(),
			d.mercadoLibre(),
			d.arca(),
			db && locals.sandbox
				? vinculoDe(db, locals.sandbox.id)
				: !usaSandbox() && locals.token && habilitado()
					? vinculoApi(locals.token).catch((e) => {
							// Que una falla de Go no tire Integraciones entera: Telegram se ve sin vincular.
							console.error('telegram', e instanceof Error ? e.message : e);
							return null;
						})
					: null,
			db && locals.sandbox
				? conexionesDe(db, locals.sandbox.id)
				: panelApi(locals.token).then((p) => {
						// Que una falla de Go no muestre "sin conexiones" cuando hay IA conectadas.
						if (p && 'fallo' in p)
							error(503, 'No pudimos traer las conexiones de IA. Probá recargar en un rato.');
						return (p?.conexiones ?? []).map((c) => ({
							id: c.id,
							nombre: c.cliente,
							permisos: c.acciones ? 'lectura acciones' : 'lectura',
							creadoEn: Date.parse(c.creadaEn),
							usadoEn: Date.parse(c.usadaEn)
						}));
					}),
			d.medicion(),
			d.tienda(),
			// Correo Argentino vive en Go; la demo no tiene cuenta de MiCorreo.
			!usaSandbox() && locals.token
				? correoDe(event)
						.estado(locals.token)
						.catch((e) => {
							console.error('correo', e instanceof Error ? e.message : e);
							return null;
						})
				: null,
			// La API de conversiones de Meta (§47) vive en Go; la demo no la tiene.
			!usaSandbox() && locals.token
				? conversionesDe(event)
						.estado(locals.token)
						.catch((e) => {
							console.error('conversiones', e instanceof Error ? e.message : e);
							return null;
						})
				: null
		]);
	// Telegram (Tincho para la dueña) está conectado si hay un chat vinculado.
	estado.conectadas.tg = !!tg;
	estado.conectadas.correo = !!correo?.conectada;
	// Manejá Tu Local Web desde tu IA: las conexiones reales por OAuth (servidor MCP).
	const ias = conexiones.map((c) => ({
		id: c.id,
		nombre: c.nombre,
		ia: iaDe(c.nombre).id,
		acciones: c.permisos.split(' ').includes('acciones'),
		desde: c.creadoEn,
		usada: c.usadoEn
	}));
	estado.ia = Object.fromEntries(IAS.map((p) => [p.id, ias.some((c) => c.ia === p.id)]));
	return {
		estado,
		mp,
		ml,
		medicion,
		conversiones,
		// Google Merchant Center y Meta Commerce Manager leen el mismo feed.
		feed: `${url.origin}/t/${tienda.slug}/feed.xml`,
		tienda: `${url.origin}/t/${tienda.slug}`,
		arca,
		correo,
		telegram: {
			disponible: habilitado() && (usaSandbox() ? !!locals.sandbox : !!locals.token),
			nombre: tg?.nombre ?? null,
			bot: BOT
		},
		mcp: {
			url: `${env.PUBLIC_SITE_URL}/api/mcp`,
			disponible: true,
			conexiones: ias
		},
		demo: usaSandbox()
	};
};
