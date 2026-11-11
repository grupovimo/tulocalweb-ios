import { describe, expect, it } from 'vitest';
import enlaces from '../../../pruebas/enlaces.json';
import { aasa, assetlinks } from './enlaces.ts';

/** Cómo decide iOS: el primer componente que coincide gana; `*` es cualquier cosa. */
function abreLaApp(url: string): boolean {
	const u = new URL(url, 'https://' + enlaces.dominio);
	for (const c of aasa().applinks.details[0].components as { '/': string; '?'?: Record<string, string>; exclude?: boolean }[]) {
		const camino = new RegExp('^' + c['/'].replace(/[.+^${}()|[\]\\]/g, '\\$&').replace(/\*/g, '.*').replace(/\?/g, '.') + '$');
		if (!camino.test(u.pathname)) continue;
		if (c['?'] && !Object.entries(c['?']).every(([k, v]) => u.searchParams.get(k) === v)) continue;
		return !c.exclude;
	}
	return false;
}

describe('enlaces', () => {
	it('cada ejemplo abre la app o se queda en la web como dice enlaces.json', () => {
		for (const [url, destino] of Object.entries(enlaces.ejemplos)) {
			expect(abreLaApp(url), url).toBe(destino !== 'web');
		}
	});
	it('la landing y las páginas públicas nunca abren la app', () => {
		for (const url of ['/', '/ayuda', '/legales/terminos', '/t/una-tienda', '/t/una-tienda/carrito', '/invitacion']) {
			expect(abreLaApp(url), url).toBe(false);
		}
	});
	it('el assetlinks lleva el paquete y las huellas', () => {
		const [a] = assetlinks();
		expect(a.target.package_name).toBe('ar.com.tulocalweb.app');
		expect(a.target.sha256_cert_fingerprints.length).toBeGreaterThan(0);
	});
});
