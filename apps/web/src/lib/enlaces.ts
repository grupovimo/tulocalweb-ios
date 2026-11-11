import enlaces from '../../../pruebas/enlaces.json';

/**
 * Qué links abren la app, desde `apps/pruebas/enlaces.json` (lo mismo que leen las pruebas de iOS y Android):
 * el AASA de Apple y el assetlinks de Android salen de ahí, así nunca se separan de lo que declaran las apps.
 */

const json = { 'content-type': 'application/json', 'cache-control': 'public, max-age=3600' };

/** Universal Links: primero `?web=1` y la lista «web» (excluidas), después lo que abre la app. */
export function aasa() {
	return {
		applinks: {
			details: [
				{
					appIDs: [enlaces.ios],
					components: [
						{ '/': '*', '?': { web: '1' }, exclude: true, comment: 'Forzar la web' },
						...enlaces.web.map((p) => ({ '/': p, exclude: true })),
						...enlaces.app.map((p) => ({ '/': p }))
					]
				}
			]
		},
		webcredentials: { apps: [enlaces.ios] }
	};
}

/** App Links: la app de Android, firmada con la llave de subida o con la de Play. */
export function assetlinks() {
	return [
		{
			relation: [
				'delegate_permission/common.handle_all_urls',
				'delegate_permission/common.get_login_creds'
			],
			target: {
				namespace: 'android_app',
				package_name: enlaces.android,
				sha256_cert_fingerprints: enlaces.huellas
			}
		}
	];
}

export const respuestaJSON = (cuerpo: unknown) =>
	new Response(JSON.stringify(cuerpo), { headers: json });
