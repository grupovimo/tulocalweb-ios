import { readFileSync } from 'node:fs';
import { describe, expect, it } from 'vitest';
import { PERMISOS_DE, puedeVer } from './permisos.ts';
import { SECCIONES } from './navegacion.ts';

describe('roles del equipo de una tienda', () => {
	it('son los mismos que aplica Go', () => {
		const url = new URL(
			'../../../../api/internal/dominio/roles/testdata/roles.json',
			import.meta.url
		);
		expect(PERMISOS_DE).toEqual(JSON.parse(readFileSync(url, 'utf8')));
	});
	it('cada rol ve sólo sus secciones', () => {
		const ve = (rol: keyof typeof PERMISOS_DE) =>
			SECCIONES.filter((s) => puedeVer(PERMISOS_DE[rol], s));
		expect(ve('duenio')).toEqual(SECCIONES);
		expect(ve('admin')).not.toContain('Usuarios y permisos');
		expect(ve('admin')).not.toContain('Plan y facturación');
		expect(ve('admin')).toContain('Finanzas');
		expect(ve('ventas')).toEqual(expect.arrayContaining(['Pedidos', 'Clientes', 'Productos']));
		expect(ve('ventas')).not.toContain('Finanzas');
		expect(ve('ventas')).not.toContain('Listas de precio');
		expect(ve('ventas')).not.toContain('Actividad');
		expect(ve('admin')).toContain('Actividad');
		expect(ve('local')).toEqual([
			'Resumen',
			'Para hacer hoy',
			'Pedidos',
			'Turnos',
			'Suscripciones',
			'Devoluciones',
			'Ventas en local',
			'Productos',
			'Inventario',
			'Envíos',
			'Configuración',
			'Ayuda'
		]);
	});
});
