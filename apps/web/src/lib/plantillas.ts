// Generado por scripts/plantillas.ts desde design/templates-data.js — no editar a mano.
/** Una plantilla de tienda del handoff. */
export type Plantilla = {
	id: string;
	nombre: string;
	/** Rubro tal como se muestra ("Gastronomía"). */
	rubro: string;
	layout: 'ed' | 'grid' | 'cat';
	tienda: string;
	img: string;
	bg: string;
	tx: string;
	ac: string;
	ff: string;
	radio: string;
	lema: string;
	oscura: boolean;
	productos: { pos: string; precio: string }[];
};

export const PLANTILLAS: Plantilla[] = [
	{
		id: 'ceibo',
		nombre: 'Ceibo',
		rubro: 'Moda',
		layout: 'ed',
		tienda: 'Tienda Lumbre',
		img: 'https://images.unsplash.com/photo-1573612664822-d7d347da7b80?auto=format&fit=crop&w=700&q=72',
		bg: '#F4EFEA',
		tx: '#1C1A17',
		ac: '#1C1A17',
		ff: "Georgia, 'Times New Roman', serif",
		radio: '2px',
		lema: 'Nueva temporada',
		oscura: false,
		productos: [
			{
				pos: '20% 30%',
				precio: '$ 8.900'
			},
			{
				pos: '80% 40%',
				precio: '$ 21.900'
			},
			{
				pos: '50% 80%',
				precio: '$ 34.900'
			},
			{
				pos: '30% 70%',
				precio: '$ 47.900'
			},
			{
				pos: '70% 20%',
				precio: '$ 20.900'
			},
			{
				pos: '50% 50%',
				precio: '$ 33.900'
			}
		]
	},
	{
		id: 'mostaza',
		nombre: 'Mostaza',
		rubro: 'Gastronomía',
		layout: 'grid',
		tienda: 'Café Bruma',
		img: 'https://images.unsplash.com/photo-1495474472287-4d71bcdd2085?auto=format&fit=crop&w=700&q=72',
		bg: '#FBF3E4',
		tx: '#3B2416',
		ac: '#C9821A',
		ff: "'Instrument Sans', system-ui, sans-serif",
		radio: '14px',
		lema: 'Tostado esta semana',
		oscura: false,
		productos: [
			{
				pos: '20% 30%',
				precio: '$ 15.900'
			},
			{
				pos: '80% 40%',
				precio: '$ 28.900'
			},
			{
				pos: '50% 80%',
				precio: '$ 41.900'
			},
			{
				pos: '30% 70%',
				precio: '$ 14.900'
			},
			{
				pos: '70% 20%',
				precio: '$ 27.900'
			},
			{
				pos: '50% 50%',
				precio: '$ 40.900'
			}
		]
	},
	{
		id: 'yerbal',
		nombre: 'Yerbal',
		rubro: 'Almacén',
		layout: 'cat',
		tienda: 'Almacén Verde',
		img: 'https://tulocalweb.com.ar/demo/almacenverde/portada.webp',
		bg: '#EEF1E8',
		tx: '#1E2B22',
		ac: '#2F5D50',
		ff: "'Instrument Sans', system-ui, sans-serif",
		radio: '10px',
		lema: 'Orgánico y a granel',
		oscura: false,
		productos: [
			{
				pos: '20% 30%',
				precio: '$ 22.900'
			},
			{
				pos: '80% 40%',
				precio: '$ 35.900'
			},
			{
				pos: '50% 80%',
				precio: '$ 8.900'
			},
			{
				pos: '30% 70%',
				precio: '$ 21.900'
			},
			{
				pos: '70% 20%',
				precio: '$ 34.900'
			},
			{
				pos: '50% 50%',
				precio: '$ 47.900'
			}
		]
	},
	{
		id: 'pampa',
		nombre: 'Pampa',
		rubro: 'Moda',
		layout: 'grid',
		tienda: 'PMPA Store',
		img: 'https://images.unsplash.com/photo-1542291026-7eec264c27ff?auto=format&fit=crop&w=700&q=72',
		bg: '#141414',
		tx: '#FFFFFF',
		ac: '#E6FF4D',
		ff: "'Instrument Sans', system-ui, sans-serif",
		radio: '6px',
		lema: 'Drop 07 · ya disponible',
		oscura: true,
		productos: [
			{
				pos: '20% 30%',
				precio: '$ 29.900'
			},
			{
				pos: '80% 40%',
				precio: '$ 42.900'
			},
			{
				pos: '50% 80%',
				precio: '$ 15.900'
			},
			{
				pos: '30% 70%',
				precio: '$ 28.900'
			},
			{
				pos: '70% 20%',
				precio: '$ 41.900'
			},
			{
				pos: '50% 50%',
				precio: '$ 14.900'
			}
		]
	},
	{
		id: 'rosedal',
		nombre: 'Rosedal',
		rubro: 'Belleza',
		layout: 'ed',
		tienda: 'Rosa Mosqueta',
		img: 'https://images.unsplash.com/photo-1612817288484-6f916006741a?auto=format&fit=crop&w=700&q=72',
		bg: '#F8EDEB',
		tx: '#3A1F25',
		ac: '#C26A7A',
		ff: "Georgia, 'Times New Roman', serif",
		radio: '18px',
		lema: 'Rutina de otoño',
		oscura: false,
		productos: [
			{
				pos: '20% 30%',
				precio: '$ 36.900'
			},
			{
				pos: '80% 40%',
				precio: '$ 9.900'
			},
			{
				pos: '50% 80%',
				precio: '$ 22.900'
			},
			{
				pos: '30% 70%',
				precio: '$ 35.900'
			},
			{
				pos: '70% 20%',
				precio: '$ 8.900'
			},
			{
				pos: '50% 50%',
				precio: '$ 21.900'
			}
		]
	},
	{
		id: 'quebracho',
		nombre: 'Quebracho',
		rubro: 'Hogar',
		layout: 'ed',
		tienda: 'Casa Nido',
		img: 'https://images.unsplash.com/photo-1555041469-a586c61ea9bc?auto=format&fit=crop&w=700&q=72',
		bg: '#EFE9E1',
		tx: '#2B2520',
		ac: '#8A5A3E',
		ff: "Georgia, 'Times New Roman', serif",
		radio: '4px',
		lema: 'Muebles hechos a mano',
		oscura: false,
		productos: [
			{
				pos: '20% 30%',
				precio: '$ 43.900'
			},
			{
				pos: '80% 40%',
				precio: '$ 16.900'
			},
			{
				pos: '50% 80%',
				precio: '$ 29.900'
			},
			{
				pos: '30% 70%',
				precio: '$ 42.900'
			},
			{
				pos: '70% 20%',
				precio: '$ 15.900'
			},
			{
				pos: '50% 50%',
				precio: '$ 28.900'
			}
		]
	},
	{
		id: 'andes',
		nombre: 'Andes',
		rubro: 'Tecnología',
		layout: 'grid',
		tienda: 'Sonido Sur',
		img: 'https://images.unsplash.com/photo-1505740420928-5e560c06d30e?auto=format&fit=crop&w=700&q=72',
		bg: '#F2F4F7',
		tx: '#0F1720',
		ac: '#2563EB',
		ff: "'Instrument Sans', system-ui, sans-serif",
		radio: '16px',
		lema: 'Envío en 24 h',
		oscura: false,
		productos: [
			{
				pos: '20% 30%',
				precio: '$ 10.900'
			},
			{
				pos: '80% 40%',
				precio: '$ 23.900'
			},
			{
				pos: '50% 80%',
				precio: '$ 36.900'
			},
			{
				pos: '30% 70%',
				precio: '$ 9.900'
			},
			{
				pos: '70% 20%',
				precio: '$ 22.900'
			},
			{
				pos: '50% 50%',
				precio: '$ 35.900'
			}
		]
	},
	{
		id: 'malbec',
		nombre: 'Malbec',
		rubro: 'Bebidas',
		layout: 'ed',
		tienda: 'Vinoteca Tres',
		img: 'https://images.unsplash.com/photo-1506377247377-2a5b3b417ebb?auto=format&fit=crop&w=700&q=72',
		bg: '#1E0E12',
		tx: '#F4E9E1',
		ac: '#C24A5A',
		ff: "Georgia, 'Times New Roman', serif",
		radio: '2px',
		lema: 'Selección del mes',
		oscura: true,
		productos: [
			{
				pos: '20% 30%',
				precio: '$ 17.900'
			},
			{
				pos: '80% 40%',
				precio: '$ 30.900'
			},
			{
				pos: '50% 80%',
				precio: '$ 43.900'
			},
			{
				pos: '30% 70%',
				precio: '$ 16.900'
			},
			{
				pos: '70% 20%',
				precio: '$ 29.900'
			},
			{
				pos: '50% 50%',
				precio: '$ 42.900'
			}
		]
	},
	{
		id: 'patio',
		nombre: 'Patio',
		rubro: 'Mascotas',
		layout: 'grid',
		tienda: 'Huellitas',
		img: 'https://images.unsplash.com/photo-1543466835-00a7907e9de1?auto=format&fit=crop&w=700&q=72',
		bg: '#FFF7E8',
		tx: '#2A1E10',
		ac: '#F08A24',
		ff: "'Instrument Sans', system-ui, sans-serif",
		radio: '22px',
		lema: 'Todo para tu compañero',
		oscura: false,
		productos: [
			{
				pos: '20% 30%',
				precio: '$ 24.900'
			},
			{
				pos: '80% 40%',
				precio: '$ 37.900'
			},
			{
				pos: '50% 80%',
				precio: '$ 10.900'
			},
			{
				pos: '30% 70%',
				precio: '$ 23.900'
			},
			{
				pos: '70% 20%',
				precio: '$ 36.900'
			},
			{
				pos: '50% 50%',
				precio: '$ 9.900'
			}
		]
	},
	{
		id: 'miga',
		nombre: 'Miga',
		rubro: 'Gastronomía',
		layout: 'cat',
		tienda: 'Panadería Lola',
		img: 'https://images.unsplash.com/photo-1509440159596-0249088772ff?auto=format&fit=crop&w=700&q=72',
		bg: '#FBF1E6',
		tx: '#3A2414',
		ac: '#B5651D',
		ff: "Georgia, 'Times New Roman', serif",
		radio: '12px',
		lema: 'Masa madre todos los días',
		oscura: false,
		productos: [
			{
				pos: '20% 30%',
				precio: '$ 31.900'
			},
			{
				pos: '80% 40%',
				precio: '$ 44.900'
			},
			{
				pos: '50% 80%',
				precio: '$ 17.900'
			},
			{
				pos: '30% 70%',
				precio: '$ 30.900'
			},
			{
				pos: '70% 20%',
				precio: '$ 43.900'
			},
			{
				pos: '50% 50%',
				precio: '$ 16.900'
			}
		]
	},
	{
		id: 'miga-clasica',
		nombre: 'Miga Clásica',
		rubro: 'Gastronomía',
		layout: 'cat',
		tienda: 'Panadería Lola Clásica',
		img: 'https://images.unsplash.com/photo-1509440159596-0249088772ff?auto=format&fit=crop&w=700&q=72',
		bg: '#FBF1E6',
		tx: '#3A2414',
		ac: '#B5651D',
		ff: "'Young Serif', Georgia, serif",
		radio: '24px',
		lema: 'Masa madre todos los días',
		oscura: false,
		productos: [
			{
				pos: '20% 30%',
				precio: '$ 31.900'
			},
			{
				pos: '80% 40%',
				precio: '$ 44.900'
			},
			{
				pos: '50% 80%',
				precio: '$ 17.900'
			},
			{
				pos: '30% 70%',
				precio: '$ 30.900'
			},
			{
				pos: '70% 20%',
				precio: '$ 43.900'
			},
			{
				pos: '50% 50%',
				precio: '$ 16.900'
			}
		]
	},
	{
		id: 'taller',
		nombre: 'Taller',
		rubro: 'Artesanías',
		layout: 'ed',
		tienda: 'Madera Viva',
		img: 'https://images.unsplash.com/photo-1687422810663-c316494f725a?auto=format&fit=crop&w=700&q=72',
		bg: '#F3EEE6',
		tx: '#2A231C',
		ac: '#6B4A2F',
		ff: "'Instrument Sans', system-ui, sans-serif",
		radio: '8px',
		lema: 'Piezas únicas',
		oscura: false,
		productos: [
			{
				pos: '20% 30%',
				precio: '$ 38.900'
			},
			{
				pos: '80% 40%',
				precio: '$ 11.900'
			},
			{
				pos: '50% 80%',
				precio: '$ 24.900'
			},
			{
				pos: '30% 70%',
				precio: '$ 37.900'
			},
			{
				pos: '70% 20%',
				precio: '$ 10.900'
			},
			{
				pos: '50% 50%',
				precio: '$ 23.900'
			}
		]
	},
	{
		id: 'hora',
		nombre: 'Hora',
		rubro: 'Accesorios',
		layout: 'grid',
		tienda: 'Relojería Norte',
		img: 'https://images.unsplash.com/photo-1522312346375-d1a52e2b99b3?auto=format&fit=crop&w=700&q=72',
		bg: '#F5F5F2',
		tx: '#161616',
		ac: '#B08D57',
		ff: "Georgia, 'Times New Roman', serif",
		radio: '0px',
		lema: 'Clásicos que duran',
		oscura: false,
		productos: [
			{
				pos: '20% 30%',
				precio: '$ 45.900'
			},
			{
				pos: '80% 40%',
				precio: '$ 18.900'
			},
			{
				pos: '50% 80%',
				precio: '$ 31.900'
			},
			{
				pos: '30% 70%',
				precio: '$ 44.900'
			},
			{
				pos: '70% 20%',
				precio: '$ 17.900'
			},
			{
				pos: '50% 50%',
				precio: '$ 30.900'
			}
		]
	},
	{
		id: 'litoral',
		nombre: 'Litoral',
		rubro: 'Regionales',
		layout: 'cat',
		tienda: 'Mate & Co',
		img: 'https://images.unsplash.com/photo-1687920492018-e21cad9023d9?auto=format&fit=crop&w=700&q=72',
		bg: '#F3E9DC',
		tx: '#3B2416',
		ac: '#6B3A22',
		ff: "'Instrument Sans', system-ui, sans-serif",
		radio: '14px',
		lema: 'El mate de cada mañana',
		oscura: false,
		productos: [
			{
				pos: '20% 30%',
				precio: '$ 12.900'
			},
			{
				pos: '80% 40%',
				precio: '$ 25.900'
			},
			{
				pos: '50% 80%',
				precio: '$ 38.900'
			},
			{
				pos: '30% 70%',
				precio: '$ 11.900'
			},
			{
				pos: '70% 20%',
				precio: '$ 24.900'
			},
			{
				pos: '50% 50%',
				precio: '$ 37.900'
			}
		]
	},
	{
		id: 'torno',
		nombre: 'Torno',
		rubro: 'Arte y diseño',
		layout: 'ed',
		tienda: 'Estudio Arcilla',
		img: 'https://images.unsplash.com/photo-1493106641515-6b5631de4bb9?auto=format&fit=crop&w=700&q=72',
		bg: '#EDE6DD',
		tx: '#2C2621',
		ac: '#A4623D',
		ff: "Georgia, 'Times New Roman', serif",
		radio: '24px',
		lema: 'Cerámica de autor',
		oscura: false,
		productos: [
			{
				pos: '20% 30%',
				precio: '$ 19.900'
			},
			{
				pos: '80% 40%',
				precio: '$ 32.900'
			},
			{
				pos: '50% 80%',
				precio: '$ 45.900'
			},
			{
				pos: '30% 70%',
				precio: '$ 18.900'
			},
			{
				pos: '70% 20%',
				precio: '$ 31.900'
			},
			{
				pos: '50% 50%',
				precio: '$ 44.900'
			}
		]
	},
	{
		id: 'tijera',
		nombre: 'Tijera',
		rubro: 'Servicios',
		layout: 'grid',
		tienda: 'Navaja',
		img: 'https://images.unsplash.com/photo-1503951914875-452162b0f3f1?auto=format&fit=crop&w=700&q=72',
		bg: '#F6F4F1',
		tx: '#1E1B2E',
		ac: '#5B3FD6',
		ff: "'Instrument Sans', system-ui, sans-serif",
		radio: '18px',
		lema: 'Cortes, barba y productos',
		oscura: false,
		productos: [
			{
				pos: '20% 30%',
				precio: '$ 26.900'
			},
			{
				pos: '80% 40%',
				precio: '$ 39.900'
			},
			{
				pos: '50% 80%',
				precio: '$ 12.900'
			},
			{
				pos: '30% 70%',
				precio: '$ 25.900'
			},
			{
				pos: '70% 20%',
				precio: '$ 38.900'
			},
			{
				pos: '50% 50%',
				precio: '$ 11.900'
			}
		]
	},
	{
		id: 'lente',
		nombre: 'Lente',
		rubro: 'Accesorios',
		layout: 'grid',
		tienda: 'Óptica Faro',
		img: 'https://images.unsplash.com/photo-1572635196237-14b3f281503f?auto=format&fit=crop&w=700&q=72',
		bg: '#FFF8EC',
		tx: '#1F1A12',
		ac: '#E0A31B',
		ff: "'Instrument Sans', system-ui, sans-serif",
		radio: '12px',
		lema: '2×1 en anteojos de sol',
		oscura: false,
		productos: [
			{
				pos: '20% 30%',
				precio: '$ 33.900'
			},
			{
				pos: '80% 40%',
				precio: '$ 46.900'
			},
			{
				pos: '50% 80%',
				precio: '$ 19.900'
			},
			{
				pos: '30% 70%',
				precio: '$ 32.900'
			},
			{
				pos: '70% 20%',
				precio: '$ 45.900'
			},
			{
				pos: '50% 50%',
				precio: '$ 18.900'
			}
		]
	}
];

export const plantilla = (id: string) => PLANTILLAS.find((p) => p.id === id);

/** Nombre del layout (handoff: TPL_LAYOUT). */
export const LAYOUT = { ed: 'Editorial', grid: 'Grilla', cat: 'Catálogo' } as const;

/** Rubros en el orden del handoff (MOSTRADOR_RUBROS). */
export const RUBROS_PLANTILLAS = [
	'Todos',
	'Moda',
	'Gastronomía',
	'Almacén',
	'Belleza',
	'Hogar',
	'Tecnología',
	'Bebidas',
	'Mascotas',
	'Artesanías',
	'Accesorios',
	'Regionales',
	'Arte y diseño',
	'Servicios'
] as const;

/**
 * Plantillas con su versión trabajada en el handoff (Plantilla X.dc.html): las tarjetas muestran la
 * captura de su portada (static/plantillas/<id>.jpg) en vez de la miniatura genérica.
 */
export const CON_PORTADA: ReadonlySet<string> = new Set([
	'ceibo',
	'mostaza',
	'yerbal',
	'tijera',
	'pampa',
	'malbec',
	'miga',
	'rosedal',
	'andes',
	'quebracho'
]);

/** Color legible sobre el acento (handoff: lumOf(acc) > 0.62 ? '#141414' : '#FFFFFF'). */
export function sobreAcento(hex: string) {
	const n = parseInt(hex.slice(1), 16);
	const lum = (0.299 * ((n >> 16) & 255) + 0.587 * ((n >> 8) & 255) + 0.114 * (n & 255)) / 255;
	return lum > 0.62 ? '#141414' : '#FFFFFF';
}
