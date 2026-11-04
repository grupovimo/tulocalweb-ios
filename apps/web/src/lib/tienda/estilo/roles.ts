import { MAPAS } from './mapas.ts';
import type { Rol } from './tipos.ts';
import type { Zona } from './estilo.ts';

/**
 * Los colores de cada plantilla por rol y zona: lo que muestra el editor antes de cambiar nada. Las
 * con archivo propio salen de mapas.ts; las del motor, de sus tokens (Motor.svelte, TH: fondo, tinta,
 * acento, tinta sobre acento, superficie, apagado, línea, panel y tinta del panel).
 */
const MOTOR: Record<string, string[]> = {
	patio: [
		'#FFF7E8',
		'#2A1E10',
		'#F08A24',
		'#2A1E10',
		'#FFFFFF',
		'#75644F',
		'#E6DCCB',
		'#2A1E10',
		'#FFF7E8'
	],
	taller: [
		'#F3EEE6',
		'#2A231C',
		'#6B4A2F',
		'#FFFFFF',
		'#FAF7F2',
		'#71665A',
		'#DCD5CB',
		'#2A231C',
		'#F3EEE6'
	],
	hora: [
		'#F5F5F2',
		'#161616',
		'#B08D57',
		'#161616',
		'#FFFFFF',
		'#66665F',
		'#DCDCD9',
		'#161616',
		'#F5F5F2'
	],
	litoral: [
		'#F3E9DC',
		'#3B2416',
		'#6B3A22',
		'#FFFFFF',
		'#FFF9F1',
		'#78604C',
		'#DBCDBD',
		'#3B2416',
		'#F3E9DC'
	],
	torno: [
		'#EDE6DD',
		'#2C2621',
		'#A4623D',
		'#FFFFFF',
		'#F6F1EB',
		'#73685E',
		'#D5CEC6',
		'#2C2621',
		'#EDE6DD'
	],
	lente: [
		'#FFF8EC',
		'#1F1A12',
		'#E0A31B',
		'#1F1A12',
		'#FFFFFF',
		'#736650',
		'#E8E1D5',
		'#1F1A12',
		'#FFF8EC'
	]
};

export function rolesDe(plantilla: string): Record<Rol | Zona, string> {
	const m = MAPAS[plantilla];
	if (m)
		return {
			...m.roles,
			encabezado: m.roles.fondo,
			encabezadoTexto: m.roles.texto,
			pie: m.pie?.fondo ?? m.roles.fondo,
			pieTexto: m.pie?.texto ?? m.roles.texto
		};
	const [fondo, texto, acento, sobreAcento, superficie, textoSuave, borde, pie, pieTexto] =
		MOTOR[plantilla] ?? MOTOR.litoral;
	return {
		fondo,
		superficie,
		texto,
		textoSuave,
		borde,
		acento,
		sobreAcento,
		encabezado: fondo,
		encabezadoTexto: texto,
		pie,
		pieTexto
	};
}

/** Combinaciones listas: cambian fondo, tarjetas, texto y bordes de una vez (el acento queda). */
export const COMBINACIONES: {
	id: string;
	nombre: string;
	colores: Partial<Record<Rol, string>>;
}[] = [
	{
		id: 'clara',
		nombre: 'Clara',
		colores: {
			fondo: '#FAFAF7',
			superficie: '#FFFFFF',
			texto: '#17171A',
			textoSuave: '#6B6B73',
			borde: '#E7E5E0'
		}
	},
	{
		id: 'crema',
		nombre: 'Crema',
		colores: {
			fondo: '#F6EEDF',
			superficie: '#FFFBF3',
			texto: '#3A2A1C',
			textoSuave: '#7C6652',
			borde: '#E6D6BE'
		}
	},
	{
		id: 'verde',
		nombre: 'Verde suave',
		colores: {
			fondo: '#EEF2EA',
			superficie: '#FFFFFF',
			texto: '#1F2A22',
			textoSuave: '#66756A',
			borde: '#D7E0D4'
		}
	},
	{
		id: 'rosa',
		nombre: 'Rosa',
		colores: {
			fondo: '#FBF1F0',
			superficie: '#FFFFFF',
			texto: '#3A1F25',
			textoSuave: '#8C6A70',
			borde: '#EFD9D6'
		}
	},
	{
		id: 'oscura',
		nombre: 'Oscura',
		colores: {
			fondo: '#111214',
			superficie: '#1B1C20',
			texto: '#F2F2F0',
			textoSuave: '#A1A1AA',
			borde: '#2C2D33'
		}
	},
	{
		id: 'noche',
		nombre: 'Noche azul',
		colores: {
			fondo: '#0F1A2B',
			superficie: '#16243A',
			texto: '#EEF3FA',
			textoSuave: '#9AABC4',
			borde: '#24344D'
		}
	}
];
