import type { Rol } from './tipos.ts';
import type { Zona } from './estilo.ts';

/**
 * Los colores de cada punto de partida por rol y zona: lo que muestra el editor antes de cambiar
 * nada (fondo, texto, acento, texto sobre el acento, superficie, texto suave, borde, fondo y texto del
 * pie). Fuente: los mapas que tenían las 17 plantillas (estilo/mapas.ts, ya borrado).
 */
// prettier-ignore
const ROLES: Record<string, string[]> = {
	andes: ['#F3F5F8', '#0E1116', '#2F6BFF', '#FFFFFF', '#FFFFFF', '#69717E', '#E3E8EF', '#0E1116', '#FFFFFF'],
	ceibo: ['#F4EFEA', '#1C1A17', '#9C4A2F', '#F4EFEA', '#F8F4EF', '#6E655C', '#E2D9CF', '#F4EFEA', '#1C1A17'],
	malbec: ['#1E0E12', '#F4E9E1', '#E8B86A', '#1E0E12', '#24121A', '#BFA9A3', '#3A2026', '#1E0E12', '#F4E9E1'],
	miga: ['#F4EADB', '#2A1C12', '#D9452B', '#F4EADB', '#FFFFFF', '#7A5F4C', '#E2CDB6', '#2A1C12', '#F4EADB'],
	'miga-clasica': ['#FBF1E6', '#3A2414', '#B5651D', '#FBF1E6', '#FFFFFF', '#7A5F4C', '#E2CDB6', '#3A2414', '#FBF1E6'],
	mostaza: ['#FBF3E4', '#3B2416', '#C9821A', '#2A170C', '#FFFFFF', '#7A5A44', '#E9D6B8', '#3B2416', '#FBF3E4'],
	pampa: ['#0B0B0B', '#F2F2F2', '#E6FF4D', '#111111', '#161616', '#8A8A8A', '#2A2A2A', '#0B0B0B', '#F2F2F2'],
	quebracho: ['#EFE9E1', '#2B2520', '#4F5B3A', '#FFFFFF', '#F7F3EE', '#6F6358', '#D8CCBC', '#2B2520', '#EFE9E1'],
	rosedal: ['#FBF4F1', '#3A1F25', '#B4475B', '#FFFFFF', '#FFFFFF', '#8C6A70', '#EAD3CF', '#F3D9DE', '#3A1F25'],
	tijera: ['#F6F4F1', '#1E1B2E', '#5B3FD6', '#FFFFFF', '#FFFFFF', '#5A566B', '#E4E1EA', '#1E1B2E', '#F6F4F1'],
	yerbal: ['#EEF1E8', '#1E2B22', '#2F5D50', '#FFFFFF', '#FFFFFF', '#6B7A6E', '#DDE8DF', '#1E2B22', '#EEF1E8'],
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
	const [fondo, texto, acento, sobreAcento, superficie, textoSuave, borde, pie, pieTexto] =
		ROLES[plantilla] ?? ROLES.litoral;
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
