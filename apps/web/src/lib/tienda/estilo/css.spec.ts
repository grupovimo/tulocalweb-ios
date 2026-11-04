import { describe, expect, it } from 'vitest';
import { cssEstilo, colorValido, esOscuro } from './css.ts';
import { MAPAS } from './mapas.ts';

describe('estilo de la tienda', () => {
	it('sin estilo no cambia nada (la plantilla se ve como siempre)', () => {
		for (const id of Object.keys(MAPAS))
			expect(cssEstilo(id, {})).toEqual({ raiz: '', reglas: '', fuentes: null });
		expect(cssEstilo('andes', null).raiz).toBe('');
	});

	it('un color elegido igual al de la plantilla tampoco cambia nada', () => {
		expect(cssEstilo('andes', { colores: { fondo: '#f3f5f8' } }).raiz).toBe('');
	});

	it('cambiar el fondo recalcula sus tonos y deja el resto', () => {
		const { raiz } = cssEstilo('andes', { colores: { fondo: '#101820' } });
		expect(raiz).toContain('--tc-f3f5f8:#101820');
		// El gris claro de las bandas sale del borde y el fondo: se mezcla con el fondo nuevo.
		expect(raiz).toMatch(/--tc-eef1f5:color-mix\(in oklab, #E3E8EF [\d.]+%, #101820\)/);
		// La cabecera translúcida conserva su transparencia.
		expect(raiz).toContain('--tc-f3f5f8-880:color-mix(in srgb, #101820 88%, transparent)');
		// El acento no se tocó.
		expect(raiz).not.toContain('--acento');
	});

	it('el acento elegido también va a --acento (lo que ya usaban las plantillas)', () => {
		expect(cssEstilo('miga', { colores: { acento: '#123456' } }).raiz).toContain(
			'--acento:#123456'
		);
	});

	it('lo que no es #RRGGBB no llega nunca al CSS', () => {
		const malo = 'red;}body{display:none';
		const r = cssEstilo('andes', { colores: { fondo: malo, pie: malo }, fuenteTitulos: "x';}" });
		expect(r).toEqual({ raiz: '', reglas: '', fuentes: null });
		expect(colorValido(' #abcdef ')).toBe('#ABCDEF');
		expect(colorValido('#abc')).toBeNull();
	});

	it('el pie con banda propia cambia sólo adentro del pie', () => {
		const { raiz, reglas } = cssEstilo('andes', {
			colores: { pie: '#5B3FD6', pieTexto: '#FFF7E0' }
		});
		expect(raiz).toBe('');
		expect(reglas).toMatch(/^\.t-estilo \[data-t-pie\]\{.*--tc-0e1116:#5B3FD6/);
		expect(reglas).toContain('#FFF7E0');
	});

	it('el encabezado propio recalcula el texto suave y los bordes adentro', () => {
		const { reglas } = cssEstilo('tijera', {
			colores: { encabezado: '#000000', encabezadoTexto: '#FFFFFF' }
		});
		expect(reglas).toMatch(/^\.t-estilo \[data-t-enc\]\{/);
		expect(reglas).toContain('--tc-f6f4f1:#000000');
	});

	it('las tipografías salen de la lista y cargan sus hojas', () => {
		const r = cssEstilo('andes', { fuenteTitulos: 'fraunces', fuenteTexto: 'inter' });
		expect(r.raiz).toContain("--tf-titulos:'Fraunces',serif");
		expect(r.raiz).toContain("--tf-texto:'Inter',sans-serif");
		expect(r.fuentes).toBe(
			'https://fonts.googleapis.com/css2?family=Fraunces:ital,opsz,wght@0,9..144,400;0,9..144,600;0,9..144,800;1,9..144,400&family=Inter:wght@400;500;600;700;800&display=swap'
		);
		// Yerbal tiene una sola: los títulos van con la de texto.
		expect(cssEstilo('yerbal', { fuenteTexto: 'lora' }).raiz).toContain(
			"--tf-titulos:'Lora',serif"
		);
	});

	it('cada plantilla mapea sus roles a tonos que existen', () => {
		for (const [id, m] of Object.entries(MAPAS)) {
			expect(Object.keys(m.tonos).length, id).toBeGreaterThan(10);
			// El acento va por var(--acento, …) (Plantilla.svelte), no por un tono.
			for (const [nombre, rol] of Object.entries(m.roles))
				if (nombre !== 'acento')
					expect(m.tonos[rol.slice(1).toLowerCase()], `${id} ${rol}`).toBeDefined();
		}
	});

	it('distingue fondos oscuros para elegir el logo', () => {
		expect(esOscuro('#0B0B0B')).toBe(true);
		expect(esOscuro('#F4EADB')).toBe(false);
	});
});
