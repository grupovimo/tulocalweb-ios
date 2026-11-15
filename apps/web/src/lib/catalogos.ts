// Generado por scripts/catalogos.ts desde el handoff (Plantillas Tienda y las plantillas trabajadas). No editar a mano.
/** Un producto de la tienda de ejemplo. Precios en pesos. */
export type ProductoEjemplo = {
	id: string;
	nombre: string;
	categoria: string | null;
	precio: number;
	anterior: number | null;
	/** Id de la foto en Unsplash o URL completa (las fotos propias de static/demo). */
	foto: string;
	etiqueta: string | null;
	detalle: string | null;
};
/** El catálogo de la tienda de ejemplo de una plantilla. */
export type Catalogo = {
	plantilla: string;
	tienda: string;
	rubro: string;
	ciudad: string;
	categorias: string[];
	productos: ProductoEjemplo[];
};

export const CATALOGOS: Catalogo[] = [
	{
		plantilla: 'andes',
		tienda: 'Sonido Sur',
		rubro: 'Tecnología',
		ciudad: 'CABA',
		categorias: ['Auriculares', 'Parlantes', 'In-ear'],
		productos: [
			{
				id: 'auricular-over-ear-anc',
				nombre: 'Auricular Over-Ear ANC',
				categoria: 'Auriculares',
				precio: 289000,
				anterior: 349000,
				foto: '1505740420928-5e560c06d30e',
				etiqueta: '−17%',
				detalle: null
			},
			{
				id: 'parlante-portatil-20-w',
				nombre: 'Parlante portátil 20 W',
				categoria: 'Parlantes',
				precio: 119000,
				anterior: null,
				foto: '1608043152269-423dbba4e7e1',
				etiqueta: 'Más vendido',
				detalle: null
			},
			{
				id: 'in-ear-inalambricos',
				nombre: 'In-ear inalámbricos',
				categoria: 'In-ear',
				precio: 94500,
				anterior: null,
				foto: '1590658268037-6bf12165a8df',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'auricular-studio',
				nombre: 'Auricular Studio',
				categoria: 'Auriculares',
				precio: 214000,
				anterior: null,
				foto: '1546435770-a3e426bf472b',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'parlante-hogar',
				nombre: 'Parlante Hogar',
				categoria: 'Parlantes',
				precio: 168000,
				anterior: 189000,
				foto: '1545454675-3531b543be5d',
				etiqueta: '−11%',
				detalle: null
			},
			{
				id: 'auricular-gamer',
				nombre: 'Auricular Gamer',
				categoria: 'Auriculares',
				precio: 132000,
				anterior: null,
				foto: '1583394838336-acd977736f90',
				etiqueta: 'Nuevo',
				detalle: null
			},
			{
				id: 'in-ear-deportivos',
				nombre: 'In-ear deportivos',
				categoria: 'In-ear',
				precio: 78900,
				anterior: null,
				foto: '1606220588913-b3aacb4d2f46',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'auricular-retro',
				nombre: 'Auricular Retro',
				categoria: 'Auriculares',
				precio: 98000,
				anterior: null,
				foto: '1484704849700-f032a568e944',
				etiqueta: null,
				detalle: null
			}
		]
	},
	{
		plantilla: 'ceibo',
		tienda: 'Tienda Lumbre',
		rubro: 'Moda',
		ciudad: 'CABA',
		categorias: ['Abrigos', 'Básicos', 'Accesorios'],
		productos: [
			{
				id: 'tapado-merino-ceibo',
				nombre: 'Tapado Merino Ceibo',
				categoria: 'Abrigos',
				precio: 186000,
				anterior: 248000,
				foto: 'https://tulocalweb.com.ar/demo/lumbre/tapado-merino-ceibo.webp',
				etiqueta: 'Hot Sale',
				detalle: null
			},
			{
				id: 'campera-de-cuero-sur',
				nombre: 'Campera de cuero Sur',
				categoria: 'Abrigos',
				precio: 312000,
				anterior: null,
				foto: '1551028719-00167b16eac5',
				etiqueta: 'Nuevo',
				detalle: null
			},
			{
				id: 'remera-pima',
				nombre: 'Remera Pima',
				categoria: 'Básicos',
				precio: 28500,
				anterior: null,
				foto: '1521572163474-6864f9cf17ab',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'jean-recto-delta',
				nombre: 'Jean Recto Delta',
				categoria: 'Básicos',
				precio: 64900,
				anterior: null,
				foto: '1542272604-787c3835535d',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'vestido-lino-ceibo',
				nombre: 'Vestido Lino Ceibo',
				categoria: 'Básicos',
				precio: 98400,
				anterior: 131200,
				foto: '1496747611176-843222e1e57c',
				etiqueta: 'Hot Sale',
				detalle: null
			},
			{
				id: 'sweater-trenzado',
				nombre: 'Sweater Trenzado',
				categoria: 'Abrigos',
				precio: 74800,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/lumbre/sweater-trenzado.webp',
				etiqueta: 'Nuevo',
				detalle: null
			},
			{
				id: 'tote-de-cuero',
				nombre: 'Tote de cuero',
				categoria: 'Accesorios',
				precio: 118000,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/lumbre/tote-de-cuero.webp',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'stiletto-noche',
				nombre: 'Stiletto Noche',
				categoria: 'Accesorios',
				precio: 142000,
				anterior: null,
				foto: '1543163521-1bf539c55dd2',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'camisa-oversize',
				nombre: 'Camisa Oversize',
				categoria: 'Básicos',
				precio: 52600,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/lumbre/camisa-oversize.webp',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'blazer-lumbre',
				nombre: 'Blazer Lumbre',
				categoria: 'Abrigos',
				precio: 168000,
				anterior: 224000,
				foto: 'https://tulocalweb.com.ar/demo/lumbre/blazer-lumbre.webp',
				etiqueta: 'Hot Sale',
				detalle: null
			},
			{
				id: 'remera-basica-negra',
				nombre: 'Remera Básica Negra',
				categoria: 'Básicos',
				precio: 24500,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/lumbre/remera-basica-negra.webp',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'conjunto-tejido',
				nombre: 'Conjunto Tejido',
				categoria: 'Básicos',
				precio: 132000,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/lumbre/conjunto-tejido.webp',
				etiqueta: 'Últimos',
				detalle: null
			}
		]
	},
	{
		plantilla: 'hora',
		tienda: 'Relojería Norte',
		rubro: 'Relojes',
		ciudad: 'CABA',
		categorias: ['Automáticos', 'Cuarzo', 'Mallas'],
		productos: [
			{
				id: 'automatico-clasico-40',
				nombre: 'Automático Clásico 40',
				categoria: 'Automáticos',
				precio: 642000,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/hora/automatico-clasico-40.webp',
				etiqueta: 'Garantía 2 años',
				detalle: null
			},
			{
				id: 'cronografo-norte',
				nombre: 'Cronógrafo Norte',
				categoria: 'Cuarzo',
				precio: 498000,
				anterior: null,
				foto: '1522312346375-d1a52e2b99b3',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'dress-watch-acero',
				nombre: 'Dress Watch Acero',
				categoria: 'Automáticos',
				precio: 356000,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/hora/dress-watch-acero.webp',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'diver-300-m',
				nombre: 'Diver 300 m',
				categoria: 'Automáticos',
				precio: 724000,
				anterior: 812000,
				foto: 'https://tulocalweb.com.ar/demo/hora/diver-300-m.webp',
				etiqueta: '−11%',
				detalle: null
			},
			{
				id: 'cuarzo-minimal',
				nombre: 'Cuarzo Minimal',
				categoria: 'Cuarzo',
				precio: 184000,
				anterior: null,
				foto: '1533139502658-0198f920d8e8',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'malla-de-cuero',
				nombre: 'Malla de cuero',
				categoria: 'Mallas',
				precio: 48000,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/hora/malla-de-cuero.webp',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'automatico-esqueleto',
				nombre: 'Automático Esqueleto',
				categoria: 'Automáticos',
				precio: 812000,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/hora/automatico-esqueleto.webp',
				etiqueta: 'Nuevo',
				detalle: null
			},
			{
				id: 'cuarzo-dorado',
				nombre: 'Cuarzo Dorado',
				categoria: 'Cuarzo',
				precio: 226000,
				anterior: null,
				foto: '1524592094714-0f0654e20314',
				etiqueta: null,
				detalle: null
			}
		]
	},
	{
		plantilla: 'lente',
		tienda: 'Óptica Faro',
		rubro: 'Anteojos',
		ciudad: 'Mar del Plata',
		categorias: ['De sol', 'Recetados', 'Lentes de contacto', 'Accesorios'],
		productos: [
			{
				id: 'sol-aviador-dorado',
				nombre: 'Sol Aviador Dorado',
				categoria: 'De sol',
				precio: 86000,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/lente/sol-aviador-dorado.webp',
				etiqueta: '2×1',
				detalle: null
			},
			{
				id: 'sol-cat-eye-carey',
				nombre: 'Sol Cat-Eye Carey',
				categoria: 'De sol',
				precio: 78000,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/lente/sol-cat-eye-carey.webp',
				etiqueta: '2×1',
				detalle: null
			},
			{
				id: 'recetado-redondo',
				nombre: 'Recetado Redondo',
				categoria: 'Recetados',
				precio: 64000,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/lente/recetado-redondo.webp',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'sol-wayfarer',
				nombre: 'Sol Wayfarer',
				categoria: 'De sol',
				precio: 72000,
				anterior: 84000,
				foto: '1572635196237-14b3f281503f',
				etiqueta: '−14%',
				detalle: null
			},
			{
				id: 'recetado-acetato',
				nombre: 'Recetado Acetato',
				categoria: 'Recetados',
				precio: 58000,
				anterior: null,
				foto: '1591076482161-42ce6da69f67',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'sol-deportivo',
				nombre: 'Sol Deportivo',
				categoria: 'De sol',
				precio: 94000,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/lente/sol-deportivo.webp',
				etiqueta: 'Nuevo',
				detalle: null
			},
			{
				id: 'estuche-rigido',
				nombre: 'Estuche rígido',
				categoria: 'Accesorios',
				precio: 9800,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/lente/estuche-rigido.webp',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'clip-on-polarizado',
				nombre: 'Clip-on polarizado',
				categoria: 'Accesorios',
				precio: 24000,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/lente/clip-on-polarizado.webp',
				etiqueta: null,
				detalle: null
			}
		]
	},
	{
		plantilla: 'litoral',
		tienda: 'Mate & Co',
		rubro: 'Regionales',
		ciudad: 'Posadas',
		categorias: ['Yerbas', 'Mates', 'Dulces'],
		productos: [
			{
				id: 'yerba-barbacua-1-kg',
				nombre: 'Yerba barbacuá 1 kg',
				categoria: 'Yerbas',
				precio: 8900,
				anterior: null,
				foto: '1675001077188-809370ce3279',
				etiqueta: 'Más vendida',
				detalle: null
			},
			{
				id: 'mate-de-calabaza',
				nombre: 'Mate de calabaza',
				categoria: 'Mates',
				precio: 18500,
				anterior: null,
				foto: '1609541994821-d909982e2f1b',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'termo-acero-1-l',
				nombre: 'Termo acero 1 L',
				categoria: 'Mates',
				precio: 42900,
				anterior: 49900,
				foto: '1634741011278-9dc2f5b6fc0b',
				etiqueta: '−14%',
				detalle: null
			},
			{
				id: 'bombilla-alpaca',
				nombre: 'Bombilla alpaca',
				categoria: 'Mates',
				precio: 12600,
				anterior: null,
				foto: '1585744259332-111384f5def9',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'miel-de-monte',
				nombre: 'Miel de monte',
				categoria: 'Dulces',
				precio: 7800,
				anterior: null,
				foto: '1558642452-9d2a7deb7f62',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'kit-matero-completo',
				nombre: 'Kit matero completo',
				categoria: 'Mates',
				precio: 58400,
				anterior: null,
				foto: '1634741011278-9dc2f5b6fc0b',
				etiqueta: 'Kit',
				detalle: null
			},
			{
				id: 'yerba-con-hierbas',
				nombre: 'Yerba con hierbas',
				categoria: 'Yerbas',
				precio: 6400,
				anterior: null,
				foto: '1675001077188-809370ce3279',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'dulce-de-mamon',
				nombre: 'Dulce de mamón',
				categoria: 'Dulces',
				precio: 5900,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/litoral/dulce-de-mamon.webp',
				etiqueta: 'Nuevo',
				detalle: null
			}
		]
	},
	{
		plantilla: 'malbec',
		tienda: 'Vinoteca Tres',
		rubro: 'Vinos',
		ciudad: 'Mendoza',
		categorias: ['Tintos', 'Blancos', 'Espumantes', 'Rosados', 'Cajas'],
		productos: [
			{
				id: 'malbec-valle-de-uco-2021',
				nombre: 'Malbec Valle de Uco 2021',
				categoria: 'Tintos',
				precio: 18900,
				anterior: null,
				foto: '1553361371-9b22f78e8b1d',
				etiqueta: 'Del mes',
				detalle: null
			},
			{
				id: 'cabernet-franc-gualtallary',
				nombre: 'Cabernet Franc Gualtallary',
				categoria: 'Tintos',
				precio: 32400,
				anterior: null,
				foto: '1584916201218-f4242ceb4809',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'blend-de-altura-2020',
				nombre: 'Blend de altura 2020',
				categoria: 'Tintos',
				precio: 26800,
				anterior: 31500,
				foto: '1568213816046-0ee1c42bd559',
				etiqueta: '−15%',
				detalle: null
			},
			{
				id: 'chardonnay-sin-madera',
				nombre: 'Chardonnay sin madera',
				categoria: 'Blancos',
				precio: 15600,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/malbec/chardonnay-sin-madera.webp',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'espumante-nature',
				nombre: 'Espumante Nature',
				categoria: 'Espumantes',
				precio: 21900,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/malbec/espumante-nature.webp',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'malbec-reserva-lujan',
				nombre: 'Malbec Reserva Luján',
				categoria: 'Tintos',
				precio: 42000,
				anterior: null,
				foto: '1510812431401-41d2bd2722f3',
				etiqueta: 'Nuevo',
				detalle: null
			},
			{
				id: 'rosado-de-malbec',
				nombre: 'Rosado de Malbec',
				categoria: 'Rosados',
				precio: 12800,
				anterior: null,
				foto: '1547595628-c61a29f496f0',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'caja-6-seleccion',
				nombre: 'Caja 6 · Selección',
				categoria: 'Cajas',
				precio: 118000,
				anterior: 147500,
				foto: '1516594915697-87eb3b1c14ea',
				etiqueta: 'Caja',
				detalle: null
			}
		]
	},
	{
		plantilla: 'miga',
		tienda: 'Panadería Lola',
		rubro: 'Panadería',
		ciudad: 'Rosario',
		categorias: ['Panes', 'Facturas', 'Dulces'],
		productos: [
			{
				id: 'hogaza-de-masa-madre',
				nombre: 'Hogaza de masa madre',
				categoria: 'Panes',
				precio: 6900,
				anterior: null,
				foto: '1586444248902-2f64eddc13df',
				etiqueta: 'Del día',
				detalle: null
			},
			{
				id: 'medialuna-de-manteca',
				nombre: 'Medialuna de manteca',
				categoria: 'Facturas',
				precio: 1200,
				anterior: null,
				foto: '1555507036-ab1f4038808a',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'budin-de-limon',
				nombre: 'Budín de limón',
				categoria: 'Dulces',
				precio: 6400,
				anterior: null,
				foto: '1691242720281-9269de4d9f86',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'rol-de-canela',
				nombre: 'Rol de canela',
				categoria: 'Facturas',
				precio: 2800,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/miga/rol-de-canela.webp',
				etiqueta: 'Nuevo',
				detalle: null
			},
			{
				id: 'pan-de-campo',
				nombre: 'Pan de campo',
				categoria: 'Panes',
				precio: 5400,
				anterior: null,
				foto: '1549931319-a545dcf3bc73',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'cookies-x6',
				nombre: 'Cookies x6',
				categoria: 'Dulces',
				precio: 7200,
				anterior: 8400,
				foto: 'https://tulocalweb.com.ar/demo/miga/cookies-x6.webp',
				etiqueta: '−15%',
				detalle: null
			},
			{
				id: 'croissant',
				nombre: 'Croissant',
				categoria: 'Facturas',
				precio: 2400,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/miga/croissant.webp',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'focaccia',
				nombre: 'Focaccia',
				categoria: 'Panes',
				precio: 6200,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/miga/focaccia.webp',
				etiqueta: null,
				detalle: null
			}
		]
	},
	{
		plantilla: 'miga-clasica',
		tienda: 'Panadería Lola Clásica',
		rubro: 'Panadería',
		ciudad: 'Rosario',
		categorias: ['Panes', 'Facturas', 'Dulces'],
		productos: [
			{
				id: 'hogaza-de-masa-madre',
				nombre: 'Hogaza de masa madre',
				categoria: 'Panes',
				precio: 6900,
				anterior: null,
				foto: '1586444248902-2f64eddc13df',
				etiqueta: 'Del día',
				detalle: null
			},
			{
				id: 'medialuna-de-manteca',
				nombre: 'Medialuna de manteca',
				categoria: 'Facturas',
				precio: 1200,
				anterior: null,
				foto: '1555507036-ab1f4038808a',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'budin-de-limon',
				nombre: 'Budín de limón',
				categoria: 'Dulces',
				precio: 6400,
				anterior: null,
				foto: '1691242720281-9269de4d9f86',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'rol-de-canela',
				nombre: 'Rol de canela',
				categoria: 'Facturas',
				precio: 2800,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/miga/rol-de-canela.webp',
				etiqueta: 'Nuevo',
				detalle: null
			},
			{
				id: 'pan-de-campo',
				nombre: 'Pan de campo',
				categoria: 'Panes',
				precio: 5400,
				anterior: null,
				foto: '1549931319-a545dcf3bc73',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'cookies-x6',
				nombre: 'Cookies x6',
				categoria: 'Dulces',
				precio: 7200,
				anterior: 8400,
				foto: 'https://tulocalweb.com.ar/demo/miga/cookies-x6.webp',
				etiqueta: '−15%',
				detalle: null
			},
			{
				id: 'croissant',
				nombre: 'Croissant',
				categoria: 'Facturas',
				precio: 2400,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/miga/croissant.webp',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'focaccia',
				nombre: 'Focaccia',
				categoria: 'Panes',
				precio: 6200,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/miga/focaccia.webp',
				etiqueta: null,
				detalle: null
			}
		]
	},
	{
		plantilla: 'mostaza',
		tienda: 'Café Bruma',
		rubro: 'Café de especialidad',
		ciudad: 'CABA',
		categorias: ['Cafés', 'Panadería', 'Salado', 'Café en grano'],
		productos: [
			{
				id: 'flat-white',
				nombre: 'Flat white',
				categoria: 'Cafés',
				precio: 4200,
				anterior: null,
				foto: '1509042239860-f550ce710b93',
				etiqueta: 'Más pedido',
				detalle: 'Doble ristretto y leche texturizada'
			},
			{
				id: 'capuccino',
				nombre: 'Capuccino',
				categoria: 'Cafés',
				precio: 3900,
				anterior: null,
				foto: '1572442388796-11668a67e53d',
				etiqueta: null,
				detalle: 'Con canela o cacao, vos elegís'
			},
			{
				id: 'espresso',
				nombre: 'Espresso',
				categoria: 'Cafés',
				precio: 2600,
				anterior: null,
				foto: '1510707577719-ae7c14805e3a',
				etiqueta: null,
				detalle: 'Blend de la casa · Brasil y Colombia'
			},
			{
				id: 'cold-brew',
				nombre: 'Cold brew',
				categoria: 'Cafés',
				precio: 4800,
				anterior: null,
				foto: '1461023058943-07fcbe16d735',
				etiqueta: 'Nuevo',
				detalle: '18 horas en frío, con o sin leche'
			},
			{
				id: 'medialunas-de-manteca',
				nombre: 'Medialunas de manteca',
				categoria: 'Panadería',
				precio: 2600,
				anterior: null,
				foto: '1555507036-ab1f4038808a',
				etiqueta: 'Más pedido',
				detalle: '2×1 hasta las 11'
			},
			{
				id: 'budin-de-limon',
				nombre: 'Budín de limón',
				categoria: 'Panadería',
				precio: 6900,
				anterior: null,
				foto: '1691242720281-9269de4d9f86',
				etiqueta: null,
				detalle: 'Con glaseado de limón y amapola'
			},
			{
				id: 'alfajor-de-maicena',
				nombre: 'Alfajor de maicena',
				categoria: 'Panadería',
				precio: 2400,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/cafebruma/alfajor-de-maicena.webp',
				etiqueta: 'Sin TACC',
				detalle: 'Dulce de leche y coco'
			},
			{
				id: 'tostado-en-masa-madre',
				nombre: 'Tostado en masa madre',
				categoria: 'Salado',
				precio: 8900,
				anterior: null,
				foto: '1528735602780-2552fd46c7af',
				etiqueta: null,
				detalle: 'Jamón cocido natural y queso tybo'
			},
			{
				id: 'avocado-toast',
				nombre: 'Avocado toast',
				categoria: 'Salado',
				precio: 9800,
				anterior: null,
				foto: '1588137378633-dea1336ce1e2',
				etiqueta: 'Vegano',
				detalle: 'Palta, huevo poché y semillas'
			},
			{
				id: 'jugo-de-naranja',
				nombre: 'Jugo de naranja',
				categoria: 'Salado',
				precio: 3600,
				anterior: null,
				foto: '1600271886742-f049cd451bba',
				etiqueta: null,
				detalle: 'Exprimido en el momento'
			},
			{
				id: 'cafe-finca-la-esperanza',
				nombre: 'Café Finca La Esperanza',
				categoria: 'Café en grano',
				precio: 17500,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/cafebruma/cafe-finca-la-esperanza.webp',
				etiqueta: null,
				detalle: 'Colombia · Panela · durazno · cacao'
			},
			{
				id: 'cafe-yirgacheffe',
				nombre: 'Café Yirgacheffe',
				categoria: 'Café en grano',
				precio: 21900,
				anterior: null,
				foto: '1447933601403-0c6688de566e',
				etiqueta: null,
				detalle: 'Etiopía · Jazmín · limón · té negro'
			},
			{
				id: 'cafe-cerrado-mineiro',
				nombre: 'Café Cerrado Mineiro',
				categoria: 'Café en grano',
				precio: 14900,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/cafebruma/cafe-cerrado-mineiro.webp',
				etiqueta: null,
				detalle: 'Brasil · Chocolate · avellana'
			},
			{
				id: 'cafe-blend-de-la-casa',
				nombre: 'Café Blend de la casa',
				categoria: 'Café en grano',
				precio: 13500,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/cafebruma/cafe-blend-de-la-casa.webp',
				etiqueta: null,
				detalle: 'Brasil + Colombia · Dulce de leche · nuez'
			},
			{
				id: 'cafe-huehuetenango',
				nombre: 'Café Huehuetenango',
				categoria: 'Café en grano',
				precio: 19400,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/cafebruma/cafe-huehuetenango.webp',
				etiqueta: null,
				detalle: 'Guatemala · Caramelo · manzana roja'
			},
			{
				id: 'cafe-decaf-sierra',
				nombre: 'Café Decaf Sierra',
				categoria: 'Café en grano',
				precio: 16200,
				anterior: null,
				foto: '1447933601403-0c6688de566e',
				etiqueta: null,
				detalle: 'México · Cacao · miel · sin cafeína'
			}
		]
	},
	{
		plantilla: 'pampa',
		tienda: 'PMPA',
		rubro: 'Moda urbana',
		ciudad: 'CABA',
		categorias: ['Buzos', 'Zapatillas', 'Gorras', 'Remeras', 'Pantalones', 'Camperas'],
		productos: [
			{
				id: 'buzo-heavy-frisa',
				nombre: 'Buzo Heavy Frisa',
				categoria: 'Buzos',
				precio: 68900,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/pampa/buzo-heavy-frisa.webp',
				etiqueta: 'Drop 07',
				detalle: null
			},
			{
				id: 'zapa-runner-90',
				nombre: 'Zapa Runner 90',
				categoria: 'Zapatillas',
				precio: 149000,
				anterior: null,
				foto: '1542291026-7eec264c27ff',
				etiqueta: 'Últimos',
				detalle: null
			},
			{
				id: 'remera-boxy-negra',
				nombre: 'Remera Boxy Negra',
				categoria: 'Remeras',
				precio: 24900,
				anterior: null,
				foto: '1503342217505-b0a15ec3261c',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'gorra-5-paneles',
				nombre: 'Gorra 5 paneles',
				categoria: 'Gorras',
				precio: 21500,
				anterior: null,
				foto: '1588850561407-ed78c282e89b',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'cargo-ripstop',
				nombre: 'Cargo Ripstop',
				categoria: 'Pantalones',
				precio: 74900,
				anterior: 89900,
				foto: 'https://tulocalweb.com.ar/demo/pampa/cargo-ripstop.webp',
				etiqueta: '−17%',
				detalle: null
			},
			{
				id: 'zapa-court-blanca',
				nombre: 'Zapa Court Blanca',
				categoria: 'Zapatillas',
				precio: 128000,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/pampa/zapa-court-blanca.webp',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'campera-coach',
				nombre: 'Campera Coach',
				categoria: 'Camperas',
				precio: 112000,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/pampa/campera-coach.webp',
				etiqueta: 'Nuevo',
				detalle: null
			},
			{
				id: 'remera-boxy-blanca',
				nombre: 'Remera Boxy Blanca',
				categoria: 'Remeras',
				precio: 24900,
				anterior: null,
				foto: '1521572163474-6864f9cf17ab',
				etiqueta: null,
				detalle: null
			}
		]
	},
	{
		plantilla: 'patio',
		tienda: 'Huellitas',
		rubro: 'Mascotas',
		ciudad: 'La Plata',
		categorias: ['Perros', 'Gatos', 'Juguetes'],
		productos: [
			{
				id: 'alimento-adulto-15-kg',
				nombre: 'Alimento adulto 15 kg',
				categoria: 'Perros',
				precio: 68900,
				anterior: 79900,
				foto: '1589924691995-400dc9ecc119',
				etiqueta: '−14%',
				detalle: null
			},
			{
				id: 'cama-redonda',
				nombre: 'Cama redonda',
				categoria: 'Perros',
				precio: 42500,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/patio/cama-redonda.webp',
				etiqueta: 'Más vendido',
				detalle: null
			},
			{
				id: 'pelota-mordible',
				nombre: 'Pelota mordible',
				categoria: 'Juguetes',
				precio: 8900,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/patio/pelota-mordible.webp',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'rascador-torre',
				nombre: 'Rascador torre',
				categoria: 'Gatos',
				precio: 54000,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/patio/rascador-torre.webp',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'correa-extensible',
				nombre: 'Correa extensible',
				categoria: 'Perros',
				precio: 18600,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/patio/correa-extensible.webp',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'snacks-dentales',
				nombre: 'Snacks dentales',
				categoria: 'Perros',
				precio: 7400,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/patio/snacks-dentales.webp',
				etiqueta: 'Nuevo',
				detalle: null
			},
			{
				id: 'comedero-doble',
				nombre: 'Comedero doble',
				categoria: 'Gatos',
				precio: 14900,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/patio/comedero-doble.webp',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'arena-aglomerante',
				nombre: 'Arena aglomerante',
				categoria: 'Gatos',
				precio: 16800,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/patio/arena-aglomerante.webp',
				etiqueta: null,
				detalle: null
			}
		]
	},
	{
		plantilla: 'quebracho',
		tienda: 'Casa Nido',
		rubro: 'Muebles',
		ciudad: 'Córdoba',
		categorias: ['Living', 'Comedor', 'Iluminación'],
		productos: [
			{
				id: 'sillon-nido',
				nombre: 'Sillón Nido',
				categoria: 'Living',
				precio: 486000,
				anterior: null,
				foto: '1567538096630-e0c55bd6374c',
				etiqueta: 'A pedido',
				detalle: null
			},
			{
				id: 'silla-quebracho',
				nombre: 'Silla Quebracho',
				categoria: 'Comedor',
				precio: 128000,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/quebracho/silla-quebracho.webp',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'mesa-ratona-paraiso',
				nombre: 'Mesa ratona Paraíso',
				categoria: 'Living',
				precio: 214000,
				anterior: 252000,
				foto: 'https://tulocalweb.com.ar/demo/quebracho/mesa-ratona-paraiso.webp',
				etiqueta: '−15%',
				detalle: null
			},
			{
				id: 'lampara-de-pie-lino',
				nombre: 'Lámpara de pie Lino',
				categoria: 'Iluminación',
				precio: 96500,
				anterior: null,
				foto: '1507473885765-e6ed057f782c',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'sofa-siesta-3-cuerpos',
				nombre: 'Sofá Siesta 3 cuerpos',
				categoria: 'Living',
				precio: 1240000,
				anterior: null,
				foto: '1555041469-a586c61ea9bc',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'butaca-pampa',
				nombre: 'Butaca Pampa',
				categoria: 'Living',
				precio: 342000,
				anterior: null,
				foto: '1586023492125-27b2c045efd7',
				etiqueta: 'Nuevo',
				detalle: null
			},
			{
				id: 'velador-ceramica',
				nombre: 'Velador Cerámica',
				categoria: 'Iluminación',
				precio: 58900,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/quebracho/velador-ceramica.webp',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'silla-taller',
				nombre: 'Silla Taller',
				categoria: 'Comedor',
				precio: 112000,
				anterior: null,
				foto: '1592078615290-033ee584e267',
				etiqueta: null,
				detalle: null
			}
		]
	},
	{
		plantilla: 'rosedal',
		tienda: 'Rosa Mosqueta',
		rubro: 'Cosmética natural',
		ciudad: 'Mendoza',
		categorias: ['Rostro', 'Cuerpo', 'Kits'],
		productos: [
			{
				id: 'serum-de-rosa-mosqueta',
				nombre: 'Sérum de rosa mosqueta',
				categoria: 'Rostro',
				precio: 18900,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/rosedal/serum-de-rosa-mosqueta.webp',
				etiqueta: 'Más vendido',
				detalle: null
			},
			{
				id: 'crema-de-noche',
				nombre: 'Crema de noche',
				categoria: 'Rostro',
				precio: 21400,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/rosedal/crema-de-noche.webp',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'tonico-de-rosas',
				nombre: 'Tónico de rosas',
				categoria: 'Rostro',
				precio: 12600,
				anterior: 15800,
				foto: 'https://tulocalweb.com.ar/demo/rosedal/tonico-de-rosas.webp',
				etiqueta: '−20%',
				detalle: null
			},
			{
				id: 'kit-rutina-completa',
				nombre: 'Kit rutina completa',
				categoria: 'Kits',
				precio: 46800,
				anterior: 58500,
				foto: 'https://tulocalweb.com.ar/demo/rosedal/kit-rutina-completa.webp',
				etiqueta: 'Kit',
				detalle: null
			},
			{
				id: 'aceite-corporal',
				nombre: 'Aceite corporal',
				categoria: 'Cuerpo',
				precio: 15200,
				anterior: null,
				foto: '1612817288484-6f916006741a',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'balsamo-labial',
				nombre: 'Bálsamo labial',
				categoria: 'Cuerpo',
				precio: 6400,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/rosedal/balsamo-labial.webp',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'limpiador-en-gel',
				nombre: 'Limpiador en gel',
				categoria: 'Rostro',
				precio: 13800,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/rosedal/limpiador-en-gel.webp',
				etiqueta: 'Nuevo',
				detalle: null
			},
			{
				id: 'contorno-de-ojos',
				nombre: 'Contorno de ojos',
				categoria: 'Rostro',
				precio: 16900,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/rosedal/contorno-de-ojos.webp',
				etiqueta: null,
				detalle: null
			}
		]
	},
	{
		plantilla: 'taller',
		tienda: 'Madera Viva',
		rubro: 'Artesanías',
		ciudad: 'Bariloche',
		categorias: ['Cocina', 'Mesa', 'Regalos'],
		productos: [
			{
				id: 'tabla-de-asado-lenga',
				nombre: 'Tabla de asado lenga',
				categoria: 'Cocina',
				precio: 38900,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/taller/tabla-de-asado-lenga.webp',
				etiqueta: 'Pieza única',
				detalle: null
			},
			{
				id: 'set-de-cucharas-x3',
				nombre: 'Set de cucharas x3',
				categoria: 'Cocina',
				precio: 18400,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/taller/set-de-cucharas-x3.webp',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'bowl-de-cipres',
				nombre: 'Bowl de ciprés',
				categoria: 'Mesa',
				precio: 26500,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/taller/bowl-de-cipres.webp',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'tabla-de-picada',
				nombre: 'Tabla de picada',
				categoria: 'Mesa',
				precio: 32000,
				anterior: 36800,
				foto: 'https://tulocalweb.com.ar/demo/taller/tabla-de-picada.webp',
				etiqueta: '−13%',
				detalle: null
			},
			{
				id: 'mate-de-algarrobo',
				nombre: 'Mate de algarrobo',
				categoria: 'Regalos',
				precio: 14800,
				anterior: null,
				foto: '1609541994821-d909982e2f1b',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'posafuentes-x4',
				nombre: 'Posafuentes x4',
				categoria: 'Mesa',
				precio: 12600,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/taller/posafuentes-x4.webp',
				etiqueta: 'Nuevo',
				detalle: null
			},
			{
				id: 'fuente-ovalada',
				nombre: 'Fuente ovalada',
				categoria: 'Mesa',
				precio: 44000,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/taller/fuente-ovalada.webp',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'kit-regalo-asador',
				nombre: 'Kit regalo asador',
				categoria: 'Regalos',
				precio: 64000,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/taller/kit-regalo-asador.webp',
				etiqueta: 'Kit',
				detalle: null
			}
		]
	},
	{
		plantilla: 'tijera',
		tienda: 'Navaja',
		rubro: 'Barbería',
		ciudad: 'CABA',
		categorias: ['Pelo', 'Barba', 'Kits'],
		productos: [
			{
				id: 'pomada-mate',
				nombre: 'Pomada mate',
				categoria: 'Pelo',
				precio: 14900,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/tijera/pomada-mate.webp',
				etiqueta: 'Más vendido',
				detalle: 'Navaja Studio'
			},
			{
				id: 'aceite-para-barba',
				nombre: 'Aceite para barba',
				categoria: 'Barba',
				precio: 11800,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/tijera/aceite-para-barba.webp',
				etiqueta: null,
				detalle: 'Navaja Studio'
			},
			{
				id: 'reuzel-shampoo-diario',
				nombre: 'Reuzel Shampoo diario',
				categoria: 'Pelo',
				precio: 16500,
				anterior: null,
				foto: '1585751119414-ef2636f8aede',
				etiqueta: null,
				detalle: 'Reuzel'
			},
			{
				id: 'cera-brillo',
				nombre: 'Cera brillo',
				categoria: 'Pelo',
				precio: 13200,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/tijera/cera-brillo.webp',
				etiqueta: 'Nuevo',
				detalle: 'Navaja Studio'
			},
			{
				id: 'proraso-crema-de-afeitar',
				nombre: 'Proraso Crema de afeitar',
				categoria: 'Barba',
				precio: 9800,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/tijera/proraso-crema-de-afeitar.webp',
				etiqueta: null,
				detalle: 'Proraso'
			},
			{
				id: 'kit-barba-completo',
				nombre: 'Kit barba completo',
				categoria: 'Kits',
				precio: 38900,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/tijera/kit-barba-completo.webp',
				etiqueta: 'Combo −20%',
				detalle: 'Navaja Studio'
			},
			{
				id: 'wahl-peine-de-madera',
				nombre: 'Wahl Peine de madera',
				categoria: 'Pelo',
				precio: 6400,
				anterior: null,
				foto: '1590159763121-7c9fd312190d',
				etiqueta: null,
				detalle: 'Wahl'
			},
			{
				id: 'spray-texturizador',
				nombre: 'Spray texturizador',
				categoria: 'Pelo',
				precio: 12600,
				anterior: null,
				foto: '1598524374912-6b0b0bab43dd',
				etiqueta: null,
				detalle: 'Navaja Studio'
			}
		]
	},
	{
		plantilla: 'torno',
		tienda: 'Estudio Arcilla',
		rubro: 'Cerámica',
		ciudad: 'CABA',
		categorias: ['Tazas', 'Platos', 'Jarrones'],
		productos: [
			{
				id: 'taza-arena',
				nombre: 'Taza Arena',
				categoria: 'Tazas',
				precio: 14800,
				anterior: null,
				foto: '1610701596007-11502861dcfa',
				etiqueta: 'Hornada nueva',
				detalle: null
			},
			{
				id: 'plato-playo-gres',
				nombre: 'Plato playo gres',
				categoria: 'Platos',
				precio: 16400,
				anterior: null,
				foto: '1578749556568-bc2c40e68b61',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'jarron-bruma',
				nombre: 'Jarrón Bruma',
				categoria: 'Jarrones',
				precio: 38900,
				anterior: null,
				foto: '1493106641515-6b5631de4bb9',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'bowl-esmaltado',
				nombre: 'Bowl esmaltado',
				categoria: 'Platos',
				precio: 18200,
				anterior: 21400,
				foto: 'https://tulocalweb.com.ar/demo/torno/bowl-esmaltado.webp',
				etiqueta: '−15%',
				detalle: null
			},
			{
				id: 'set-de-4-tazas',
				nombre: 'Set de 4 tazas',
				categoria: 'Tazas',
				precio: 52000,
				anterior: null,
				foto: '1610701596007-11502861dcfa',
				etiqueta: 'Set',
				detalle: null
			},
			{
				id: 'mate-de-ceramica',
				nombre: 'Mate de cerámica',
				categoria: 'Tazas',
				precio: 18500,
				anterior: null,
				foto: '1609541994821-d909982e2f1b',
				etiqueta: null,
				detalle: null
			},
			{
				id: 'fuente-tierra',
				nombre: 'Fuente Tierra',
				categoria: 'Platos',
				precio: 46000,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/torno/fuente-tierra.webp',
				etiqueta: 'Nuevo',
				detalle: null
			},
			{
				id: 'florero-mini',
				nombre: 'Florero mini',
				categoria: 'Jarrones',
				precio: 12400,
				anterior: null,
				foto: '1565193566173-7a0ee3dbe261',
				etiqueta: null,
				detalle: null
			}
		]
	},
	{
		plantilla: 'yerbal',
		tienda: 'Almacén Verde',
		rubro: 'Almacén natural',
		ciudad: 'Córdoba',
		categorias: ['Frutos secos', 'Almacén', 'Orgánicos'],
		productos: [
			{
				id: 'almendras-peladas-non-pareil',
				nombre: 'Almendras peladas Non Pareil',
				categoria: 'Frutos secos',
				precio: 14990,
				anterior: 21400,
				foto: '1508061253366-f7da158b6d46',
				etiqueta: null,
				detalle: '1 kg'
			},
			{
				id: 'nueces-mariposa',
				nombre: 'Nueces mariposa',
				categoria: 'Frutos secos',
				precio: 12800,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/almacenverde/nueces-mariposa.webp',
				etiqueta: null,
				detalle: '500 g'
			},
			{
				id: 'mani-tostado-sin-sal',
				nombre: 'Maní tostado sin sal',
				categoria: 'Frutos secos',
				precio: 3900,
				anterior: 4600,
				foto: '1567892737950-30c4db37cd89',
				etiqueta: null,
				detalle: '1 kg'
			},
			{
				id: 'mix-energetico',
				nombre: 'Mix energético',
				categoria: 'Frutos secos',
				precio: 8400,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/almacenverde/mix-energetico.webp',
				etiqueta: null,
				detalle: '500 g'
			},
			{
				id: 'yerba-organica-misionera',
				nombre: 'Yerba orgánica Misionera',
				categoria: 'Almacén',
				precio: 6200,
				anterior: null,
				foto: '1675001077188-809370ce3279',
				etiqueta: null,
				detalle: '1 kg'
			},
			{
				id: 'harina-integral-agroecologica',
				nombre: 'Harina integral agroecológica',
				categoria: 'Orgánicos',
				precio: 2900,
				anterior: 3400,
				foto: 'https://tulocalweb.com.ar/demo/almacenverde/harina-integral-agroecologica.webp',
				etiqueta: null,
				detalle: '1 kg'
			},
			{
				id: 'avena-arrollada',
				nombre: 'Avena arrollada',
				categoria: 'Almacén',
				precio: 2400,
				anterior: null,
				foto: '1614961233913-a5113a4a34ed',
				etiqueta: null,
				detalle: '1 kg'
			},
			{
				id: 'miel-de-monte',
				nombre: 'Miel de monte',
				categoria: 'Orgánicos',
				precio: 7600,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/almacenverde/miel-de-monte.webp',
				etiqueta: null,
				detalle: '500 g'
			},
			{
				id: 'aceite-de-oliva-extra-virgen',
				nombre: 'Aceite de oliva extra virgen',
				categoria: 'Almacén',
				precio: 15900,
				anterior: 18200,
				foto: '1474979266404-7eaacbcd87c5',
				etiqueta: null,
				detalle: '1 L'
			},
			{
				id: 'arroz-yamani',
				nombre: 'Arroz yamaní',
				categoria: 'Orgánicos',
				precio: 3600,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/almacenverde/arroz-yamani.webp',
				etiqueta: null,
				detalle: '1 kg'
			},
			{
				id: 'lentejas',
				nombre: 'Lentejas',
				categoria: 'Almacén',
				precio: 3200,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/almacenverde/lentejas.webp',
				etiqueta: null,
				detalle: '1 kg'
			},
			{
				id: 'semillas-de-chia',
				nombre: 'Semillas de chía',
				categoria: 'Orgánicos',
				precio: 4800,
				anterior: 5600,
				foto: 'https://tulocalweb.com.ar/demo/almacenverde/semillas-de-chia.webp',
				etiqueta: null,
				detalle: '500 g'
			},
			{
				id: 'pasas-de-uva-rubias',
				nombre: 'Pasas de uva rubias',
				categoria: 'Frutos secos',
				precio: 5200,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/almacenverde/pasas-de-uva-rubias.webp',
				etiqueta: null,
				detalle: '500 g'
			},
			{
				id: 'cafe-de-especialidad',
				nombre: 'Café de especialidad',
				categoria: 'Almacén',
				precio: 17500,
				anterior: null,
				foto: 'https://tulocalweb.com.ar/demo/almacenverde/cafe-de-especialidad.webp',
				etiqueta: null,
				detalle: '250 g'
			}
		]
	}
];

export const catalogoDe = (plantilla: string) => CATALOGOS.find((c) => c.plantilla === plantilla);
