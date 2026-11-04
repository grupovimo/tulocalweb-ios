/**
 * Las tipografías que se pueden elegir en Mi tienda (Google Fonts). Lista cerrada: el id es lo que
 * se guarda y `css` es lo único que llega al CSS de la tienda.
 */
export type Fuente = {
	nombre: string;
	css: string;
	google: string;
	tipo: 'sans' | 'serif' | 'display';
};

export const FUENTES: Record<string, Fuente> = {
	inter: {
		nombre: 'Inter',
		css: "'Inter',sans-serif",
		google: 'Inter:wght@400;500;600;700;800',
		tipo: 'sans'
	},
	'instrument-sans': {
		nombre: 'Instrument Sans',
		css: "'Instrument Sans',sans-serif",
		google: 'Instrument+Sans:wght@400;500;600;700',
		tipo: 'sans'
	},
	'dm-sans': {
		nombre: 'DM Sans',
		css: "'DM Sans',sans-serif",
		google: 'DM+Sans:wght@400;500;600;700;800',
		tipo: 'sans'
	},
	manrope: {
		nombre: 'Manrope',
		css: "'Manrope',sans-serif",
		google: 'Manrope:wght@400;500;600;700;800',
		tipo: 'sans'
	},
	poppins: {
		nombre: 'Poppins',
		css: "'Poppins',sans-serif",
		google: 'Poppins:wght@400;500;600;700;800',
		tipo: 'sans'
	},
	montserrat: {
		nombre: 'Montserrat',
		css: "'Montserrat',sans-serif",
		google: 'Montserrat:wght@400;500;600;700;800',
		tipo: 'sans'
	},
	'space-grotesk': {
		nombre: 'Space Grotesk',
		css: "'Space Grotesk',sans-serif",
		google: 'Space+Grotesk:wght@400;500;600;700',
		tipo: 'sans'
	},
	sora: {
		nombre: 'Sora',
		css: "'Sora',sans-serif",
		google: 'Sora:wght@400;500;600;700;800',
		tipo: 'display'
	},
	syne: {
		nombre: 'Syne',
		css: "'Syne',sans-serif",
		google: 'Syne:wght@500;600;700;800',
		tipo: 'display'
	},
	bricolage: {
		nombre: 'Bricolage Grotesque',
		css: "'Bricolage Grotesque',sans-serif",
		google: 'Bricolage+Grotesque:opsz,wght@12..96,400;12..96,600;12..96,800',
		tipo: 'display'
	},
	'archivo-black': {
		nombre: 'Archivo Black',
		css: "'Archivo Black',sans-serif",
		google: 'Archivo+Black',
		tipo: 'display'
	},
	fredoka: {
		nombre: 'Fredoka',
		css: "'Fredoka',sans-serif",
		google: 'Fredoka:wght@400;500;600;700',
		tipo: 'display'
	},
	playfair: {
		nombre: 'Playfair Display',
		css: "'Playfair Display',serif",
		google: 'Playfair+Display:ital,wght@0,400;0,600;0,800;1,400;1,600',
		tipo: 'serif'
	},
	fraunces: {
		nombre: 'Fraunces',
		css: "'Fraunces',serif",
		google: 'Fraunces:ital,opsz,wght@0,9..144,400;0,9..144,600;0,9..144,800;1,9..144,400',
		tipo: 'serif'
	},
	'dm-serif': {
		nombre: 'DM Serif Display',
		css: "'DM Serif Display',serif",
		google: 'DM+Serif+Display:ital@0;1',
		tipo: 'serif'
	},
	cormorant: {
		nombre: 'Cormorant Garamond',
		css: "'Cormorant Garamond',serif",
		google: 'Cormorant+Garamond:ital,wght@0,500;0,600;0,700;1,500;1,600',
		tipo: 'serif'
	},
	lora: {
		nombre: 'Lora',
		css: "'Lora',serif",
		google: 'Lora:ital,wght@0,400;0,600;0,700;1,400',
		tipo: 'serif'
	},
	'young-serif': {
		nombre: 'Young Serif',
		css: "'Young Serif',serif",
		google: 'Young+Serif',
		tipo: 'serif'
	},
	caveat: {
		nombre: 'Caveat',
		css: "'Caveat',cursive",
		google: 'Caveat:wght@500;700',
		tipo: 'display'
	}
};

export const IDS_FUENTES = Object.keys(FUENTES);
