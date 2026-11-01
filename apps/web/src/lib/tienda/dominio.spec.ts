import { describe, expect, it } from 'vitest';
import { esDominioPropio, pathInterno } from './dominio.ts';

describe('dominio propio', () => {
	it('distingue los hosts de Tu Local Web de los de las tiendas', () => {
		for (const h of [
			'tulocalweb.com.ar',
			'panel.tulocalweb.com.ar',
			'WWW.tulocalweb.com.ar',
			'localhost',
			'127.0.0.1',
			'mostrador.nawe.workers.dev'
		])
			expect(esDominioPropio(h), h).toBe(false);
		for (const h of [
			'lunamates.com.ar',
			'www.lunamates.com.ar',
			'tulocalweb.com.ar.evil.com',
			'tienda-lumbre.localhost'
		])
			expect(esDominioPropio(h), h).toBe(true);
	});

	it('lleva el path del dominio a /t/<slug> y deja lo que ya es de la tienda', () => {
		expect(pathInterno('luna', '/')).toBe('/t/luna');
		expect(pathInterno('luna', '/producto/mate')).toBe('/t/luna/producto/mate');
		expect(pathInterno('luna', '/catalogo')).toBe('/t/luna/catalogo');
		// Lo que la tienda ya pide con su path largo, y lo global (legales, el propio resolver).
		expect(pathInterno('luna', '/t/luna/pregunta')).toBeUndefined();
		expect(pathInterno('luna', '/t/luna')).toBeUndefined();
		expect(pathInterno('luna', '/legales/defensa-del-consumidor')).toBeUndefined();
		expect(pathInterno('luna', '/api/dominio')).toBeUndefined();
		// Otra tienda por su path largo no se sirve tal cual: queda adentro de esta (y da 404).
		expect(pathInterno('luna', '/t/sol')).toBe('/t/luna/t/sol');
		expect(pathInterno('luna', '/t/lunares')).toBe('/t/luna/t/lunares');
	});
});

describe('slugDeSubdominio', () => {
	it('saca la tienda del subdominio, salvo los de la plataforma', async () => {
		const { slugDeSubdominio } = await import('./dominio.ts');
		expect(slugDeSubdominio('la-esquina.tulocalweb.com.ar')).toBe('la-esquina');
		expect(slugDeSubdominio('LA-ESQUINA.tulocalweb.com.ar')).toBe('la-esquina');
		expect(slugDeSubdominio('api.tulocalweb.com.ar')).toBeNull();
		expect(slugDeSubdominio('www.tulocalweb.com.ar')).toBeNull();
		expect(slugDeSubdominio('tulocalweb.com.ar')).toBeNull();
		expect(slugDeSubdominio('a.b.tulocalweb.com.ar')).toBeNull();
		expect(slugDeSubdominio('mitienda.com.ar')).toBeNull();
	});
});
