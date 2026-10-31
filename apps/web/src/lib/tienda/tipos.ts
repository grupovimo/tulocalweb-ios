import type { components } from '@mostrador/contrato/esquema';

/** Los turnos de la tienda (servicios y profesionales activos, seña, horario); null si no toma. */
export type TurnosTienda = components['schemas']['TurnosPublicos'];
export type TurnoPublico = components['schemas']['TurnoPublico'];
export type HorarioLibre = components['schemas']['HorarioLibre'];

/**
 * La tienda online tal como la ve quien compra: lo publicado, sin costos, márgenes ni datos de
 * clientes. La arma el servidor (demo o API) y la dibuja la plantilla elegida en Mi tienda.
 */
export type ProductoPublico = {
	id: string;
	/** El link (/producto/<slug>), fijo desde que se creó. En la demo no viene: el id ya es legible. */
	slug?: string;
	nombre: string;
	descripcion: string;
	categoria: string | null;
	precio: number;
	/** Precio tachado (oferta). */
	anterior: number | null;
	/** URL de la foto (o id de Unsplash en la demo). */
	foto: string | null;
	etiqueta: string | null;
	/** Con opciones, la suma del stock de sus variantes. */
	stock: number;
	/** Galería (la primera es `foto`). */
	fotos: string[];
	/** Talle, Color, Presentación… (vacío: se vende sin elegir nada). */
	opciones: OpcionProducto[];
	/** Sólo las activas. `precio` null: el del producto. */
	variantes: VariantePublica[];
};

export type OpcionProducto = { nombre: string; valores: string[] };

export type VariantePublica = {
	id: string;
	valores: Record<string, string>;
	precio: number | null;
	stock: number;
	foto: string | null;
};

export type ResenaPublica = {
	texto: string;
	autor: string;
	/** El nombre del producto ('' si es de la tienda en general). */
	producto: string;
	estrellas: number;
	/** La dejó quien compró (con el código de un pedido cobrado). */
	conCompra: boolean;
	/** La respuesta pública de la tienda. */
	respuesta: string | null;
};

export type PreguntaPublica = { productoId: string; pregunta: string; respuesta: string };

export type TiendaPublica = {
	slug: string;
	nombre: string;
	plantilla: string;
	/** 0–3: la paleta elegida en Mi tienda. */
	paleta: number;
	/** Secciones visibles de la portada, en orden. */
	secciones: ('hero' | 'cats' | 'prods' | 'revs' | 'ig' | 'wa')[];
	portada: { titulo: string; lema: string; foto: string | null };
	/** El logo subido en Mi tienda (URL); vacío: el nombre en letras. */
	logo: string;
	/**
	 * Mi tienda › Al compartir: título, descripción y foto propia de la vista previa (vacío: lo
	 * sugerido y la imagen dibujada con la plantilla, lib/tienda/compartir.ts).
	 */
	compartir: { titulo: string; descripcion: string; foto: string };
	ciudad: string;
	direccion: string;
	whatsapp: string;
	categorias: { nombre: string; cantidad: number; foto: string | null }[];
	productos: ProductoPublico[];
	/** Ids de los más vendidos, del más al menos. */
	masVendidos: string[];
	/** Las visibles, de la más nueva a la más vieja. */
	resenas: ResenaPublica[];
	promedio: number | null;
	/** Preguntas ya respondidas por la tienda (las demás no se muestran). */
	preguntas: PreguntaPublica[];
	fotos: string[];
	cobros: {
		transferencia: number;
		cuotas: number;
		mercadoPago: boolean;
		/** Datos para transferir que cargó la tienda (null: no ofrece transferencia). `ejemplo` en la demo. */
		datosTransferencia: { alias: string; cbu: string; titular: string; ejemplo?: boolean } | null;
		/** Responsable inscripta con ARCA conectado: el checkout pide CUIT y razón social para factura A. */
		facturaA?: boolean;
	};
	/**
	 * Cómo entrega (Configuración › Envíos y retiro). Sin retiro, su dirección y horario van vacíos;
	 * sin moto, su costo (en pesos), zona y demora. correo: la tienda conectó su cuenta de MiCorreo
	 * (cotiza la API, envio.svelte.ts).
	 */
	envios: EnviosTienda;
	/** Turnos con seña; null si la tienda no los tomó. `sena` es el % que se cobra de verdad. */
	turnos: TurnosTienda | null;
	/** Sólo mirar: la vitrina de una plantilla (nadie la administra, no se compra). */
	vitrina: boolean;
	/** El dominio propio principal y verificado (mitienda.com.ar), o null: sólo /t/<slug>. */
	dominio?: string | null;
	/** Prefijo de los links: '' en el dominio propio; si falta, /t/<slug>. Lo pone el layout. */
	base?: string;
	/**
	 * Los IDs de GA4 y del píxel de Meta de la tienda (lib/tienda/medicion.ts). Se cargan sólo con
	 * el consentimiento de quien compra; la vitrina nunca los tiene.
	 */
	medicion: { ga4: string | null; pixel: string | null };
};

/** Lo que muestra la portada de los turnos: los libres de hoy y el próximo libre de cada uno. */
export type ProximosTurnos = {
	/** Inicios (ISO −03:00) libres hoy, con cualquiera, para el primer servicio. */
	hoy: string[];
	/** Por profesional: su próximo horario libre (ISO) o null. */
	porProfesional: Record<string, string | null>;
};

export type EnviosTienda = {
	gratisDesde: number | null;
	retiro: boolean;
	retiroDireccion: string;
	retiroHorario: string;
	moto: boolean;
	motoCosto: number;
	motoZona: string;
	motoDemora: string;
	correo: boolean;
};

/** Una opción de envío cotizada para un código postal. */
export type OpcionEnvio = {
	id: 'andreani' | 'correo-domicilio' | 'correo-sucursal' | 'moto' | 'retiro';
	nombre: string;
	detalle: string;
	precio: number;
};
