import { describe, expect, it } from 'vitest';
import { csrfForbidden } from './csrf';

const ORIGIN = 'https://tulocalweb.com.ar';
const U = (p: string) => new URL(`${ORIGIN}${p}`);
const req = (method: string, type: string | null, origin: string | null) =>
	new Request(`${ORIGIN}/x`, {
		method,
		headers: {
			...(type ? { 'content-type': type } : {}),
			...(origin ? { origin } : {})
		}
	});
const FORM = 'application/x-www-form-urlencoded';

describe('csrfForbidden (misma regla que SvelteKit)', () => {
	it('un formulario de otro origen o sin origen se rechaza', () => {
		expect(csrfForbidden(req('POST', FORM, 'https://malo.com'), U('/panel/pedidos'))).toBe(true);
		expect(csrfForbidden(req('POST', FORM, null), U('/panel/pedidos'))).toBe(true);
		expect(csrfForbidden(req('POST', null, null), U('/panel/pedidos'))).toBe(true);
		expect(csrfForbidden(req('POST', 'multipart/form-data; boundary=x', null), U('/x'))).toBe(true);
	});
	it('desde nuestro origen, JSON o GET pasan', () => {
		expect(csrfForbidden(req('POST', FORM, ORIGIN), U('/panel/pedidos'))).toBe(false);
		expect(csrfForbidden(req('POST', 'application/json', null), U('/api/x'))).toBe(false);
		expect(csrfForbidden(req('GET', null, null), U('/panel'))).toBe(false);
	});
	it('OAuth del conector, el servidor MCP y webhooks pasan sin origen', () => {
		for (const p of [
			'/api/mcp',
			'/api/oauth/token',
			'/api/oauth/revoke',
			'/api/oauth/register',
			'/api/webhooks/mercadopago',
			'/api/webhooks/caja'
		])
			expect(csrfForbidden(req('POST', FORM, null), U(p))).toBe(false);
		// La pantalla de autorización (con cookies) sigue protegida.
		expect(csrfForbidden(req('POST', FORM, null), U('/oauth/autorizar'))).toBe(true);
		expect(csrfForbidden(req('POST', FORM, null), U('/api/webhooks/a/b'))).toBe(true);
	});
	it('la baja de un clic del newsletter (RFC 8058) pasa sin origen; confirmar, no', () => {
		const token = 'aB3_dE5-fG7hI9jK1lM3nO5p';
		expect(csrfForbidden(req('POST', FORM, null), U(`/t/luna/newsletter/baja/${token}`))).toBe(
			false
		);
		expect(csrfForbidden(req('POST', FORM, null), U(`/t/luna/newsletter/confirmar/${token}`))).toBe(
			true
		);
		expect(csrfForbidden(req('POST', FORM, null), U('/t/luna/newsletter/baja/corto'))).toBe(true);
	});
});
