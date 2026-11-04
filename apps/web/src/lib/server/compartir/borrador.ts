/**
 * La tienda como quedaría con el borrador de Mi tienda: lo publicado (productos, categorías,
 * envíos) con la plantilla, los colores, la portada, el logo y Al compartir del borrador. Sirve para
 * la vista previa del editor. Si la tienda todavía no se puede leer como pública, se arma con lo que
 * muestra el editor.
 */
import type { Cookies } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { MiTienda } from '#lib/server/datos/tipos.ts';
import { COOKIE } from '#lib/server/sandbox/sandbox.ts';
import { resolverTienda } from '#lib/server/tienda/publica.ts';
import { COMPARTIR_VACIO } from '#lib/tienda/compartir.ts';
import type { TiendaPublica } from '#lib/tienda/tipos.ts';

export function conBorrador(t: TiendaPublica, m: MiTienda): TiendaPublica {
	return {
		...t,
		plantilla: m.plantillaId,
		paleta: m.diseno.paleta,
		secciones: m.diseno.secciones.filter((s) => s.visible).map((s) => s.id),
		estilo: m.diseno.estilo ?? {},
		logo: m.diseno.logo ?? '',
		portada: {
			titulo: m.vista.portada.titulo,
			lema: m.vista.portada.lema ?? '',
			foto: m.vista.portada.foto
		},
		compartir: {
			titulo: m.diseno.compartirTitulo ?? '',
			descripcion: m.diseno.compartirDescripcion ?? '',
			foto: m.diseno.compartirFoto ?? ''
		}
	};
}

/** Sólo con lo del editor (una tienda que todavía no se ve en público). */
function desdeElEditor(slug: string, m: MiTienda): TiendaPublica {
	const cats = new Map<string, number>();
	for (const c of m.vista.categorias) cats.set(c, 0);
	return {
		slug,
		nombre: m.vista.tienda,
		plantilla: m.plantillaId,
		paleta: m.diseno.paleta,
		secciones: [],
		portada: { titulo: '', lema: '', foto: null },
		logo: '',
		compartir: COMPARTIR_VACIO,
		ciudad: '',
		direccion: '',
		whatsapp: '',
		// Sin saber cuántos productos tiene cada una, no se cuentan (nada inventado).
		categorias: [],
		productos: m.vista.productos.map((p, i) => ({
			id: String(i),
			nombre: p.nombre,
			descripcion: '',
			categoria: null,
			precio: p.precio,
			anterior: null,
			foto: p.foto,
			etiqueta: null,
			stock: 0,
			fotos: [],
			opciones: [],
			variantes: []
		})),
		masVendidos: [],
		resenas: [],
		promedio: null,
		preguntas: [],
		fotos: m.vista.fotos,
		cobros: { transferencia: 0, cuotas: 0, mercadoPago: false, datosTransferencia: null },
		envios: {
			gratisDesde: null,
			retiro: false,
			retiroDireccion: '',
			retiroHorario: '',
			moto: false,
			motoCosto: 0,
			motoZona: '',
			motoDemora: '',
			correo: false
		},
		turnos: null,
		suscripciones: [],
		vitrina: false,
		dominio: null,
		medicion: { ga4: null, pixel: null }
	};
}

/** La tienda de la sesión con su borrador, y lo que muestra el editor. */
export async function tiendaDelBorrador(locals: App.Locals, cookies: Cookies) {
	const datos = datosDe(locals);
	const [m, t] = await Promise.all([datos.miTienda(), datos.tienda()]);
	const r = await resolverTienda(t.slug, cookies.get(COOKIE)).catch(() => null);
	// La vitrina de la plantilla no es esta tienda: sirve sólo si es la suya.
	const publica = r && !r.tienda.vitrina ? r.tienda : null;
	const base = publica ?? desdeElEditor(t.slug, m);
	// Con los productos del editor, los cuenta igual (la pública ya los trae todos).
	return { mi: m, tienda: conBorrador(base, m), publica: !!publica };
}
