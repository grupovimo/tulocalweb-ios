import { createHash } from 'node:crypto';
import { describe, expect, it, vi } from 'vitest';
import { eventoAlta, mandarAlta, medible } from './meta.ts';

const ALTA = {
	tiendaId: 't1',
	email: '  Lucia@Mail.com ',
	origen: { fuente: 'facebook', campana: 'avisos' },
	ip: '190.1.2.3',
	agente: 'Instagram',
	fbp: 'fb.1.1.2',
	url: 'https://tulocalweb.com.ar/onboarding',
	hora: 1_800_000_000
};
const galleta = (v?: string) => ({ get: () => v });

describe('eventoAlta', () => {
	it('hashea el email normalizado y no manda campos vacíos', async () => {
		const e = await eventoAlta(ALTA);
		expect(e.event_name).toBe('CompleteRegistration');
		expect(e.event_id).toBe('alta-t1');
		expect(e.user_data).toEqual({
			em: [createHash('sha256').update('lucia@mail.com').digest('hex')],
			client_ip_address: '190.1.2.3',
			client_user_agent: 'Instagram',
			fbp: 'fb.1.1.2'
		});
		expect(e.custom_data).toEqual({ status: 'registrada', content_name: 'avisos' });
	});
});

describe('medible', () => {
	it('sólo con píxel, token y el permiso de marketing', () => {
		const c = { pixel: '123456789012345', token: 'token' };
		expect(medible(galleta('si'), null)).toBe(false);
		expect(medible(galleta('no'), c)).toBe(false);
		expect(medible(galleta(), c)).toBe(false);
		expect(medible(galleta('si'), c)).toBe(true);
	});
});

describe('mandarAlta', () => {
	it('va al píxel con el código de prueba, y sin token no sale nada', async () => {
		const f = vi.fn(async () => new Response('{}'));
		await mandarAlta(ALTA, null, f);
		expect(f).not.toHaveBeenCalled();
		await mandarAlta(ALTA, { pixel: '123', token: 'tk', prueba: 'TEST1' }, f);
		const [url, init] = f.mock.calls[0] as unknown as [string, RequestInit];
		expect(url).toBe('https://graph.facebook.com/v23.0/123/events');
		const cuerpo = JSON.parse(String(init.body));
		expect(cuerpo.test_event_code).toBe('TEST1');
		expect(cuerpo.access_token).toBe('tk');
		expect(cuerpo.data).toHaveLength(1);
	});
});
