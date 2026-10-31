/**
 * Suscripciones de la demo (§51). Malbec arranca con el club de vinos de su diseño
 * (design/Plantilla Malbec.dc.html: `plans` — Descubrir, Explorador, Coleccionista) y Mostaza con el
 * café en grano por suscripción (design/Plantilla Mostaza.dc.html: «Suscribite y ahorrá 15% · Te llega
 * cada 2 semanas»), con unos suscriptores de ejemplo fechados respecto de cuándo se creó el sandbox.
 * Las demás plantillas arrancan sin planes.
 *
 * Como todo el sandbox: esto sale del código y D1 guarda sólo lo que cambia el visitante (entidad
 * `suscripciones`, claves `p:<id>`, `s:<id>` y `c:<codigo>` para encontrar el link privado sin la
 * cookie). En la demo nadie paga: suscribirse la deja activa sin cobrar (la página lo dice).
 */
import type { PlanSuscripcion, Suscriptor } from '../datos/tipos.ts';
import type { PlanPublico, SuscripcionPublica } from '#lib/tienda/tipos.ts';
import { diasDe } from '#lib/tienda/suscripciones.ts';
import { textoVariante } from '#lib/tienda/variantes.ts';
import type { RegProducto } from './listas.ts';

export const ENTIDAD = 'suscripciones';

const DIA = 86_400_000;

type Base = Omit<PlanSuscripcion, 'suscriptores'>;

const plan = (
	id: string,
	nombre: string,
	descripcion: string,
	foto: string,
	precio: number,
	frecuencia: Base['frecuencia'],
	incluye: string,
	productos: [string, number][],
	orden: number
): Base => ({
	id,
	nombre,
	descripcion,
	foto,
	precio,
	frecuencia,
	incluye,
	productos: productos.map(([productoId, cantidad]) => ({ productoId, varianteId: '', cantidad })),
	envio: true,
	retiro: true,
	cupo: null,
	activo: true,
	orden
});

const PLANES: Record<string, Base[]> = {
	malbec: [
		plan(
			'club-descubrir',
			'Descubrir',
			'Tres vinos para empezar a conocer Mendoza, con la ficha de cata de cada uno.',
			'1510812431401-41d2bd2722f3',
			64000,
			'mensual',
			'3 botellas por mes',
			[
				['malbec-valle-de-uco-2021', 1],
				['chardonnay-sin-madera', 1],
				['rosado-de-malbec', 1]
			],
			0
		),
		plan(
			'club-explorador',
			'Explorador',
			'Tintos de altura, rango medio. Llega el primer martes de cada mes.',
			'1444157545135-c045be691b05',
			118000,
			'mensual',
			'6 botellas · el más elegido',
			[
				['malbec-valle-de-uco-2021', 2],
				['cabernet-franc-gualtallary', 1],
				['blend-de-altura-2020', 1],
				['chardonnay-sin-madera', 1],
				['espumante-nature', 1]
			],
			1
		),
		plan(
			'club-coleccionista',
			'Coleccionista',
			'Las etiquetas de guarda que elige la sommelier, para abrir o para guardar.',
			'1553361371-9b22f78e8b1d',
			196000,
			'mensual',
			'6 de alta gama',
			[
				['malbec-reserva-lujan', 3],
				['cabernet-franc-gualtallary', 3]
			],
			2
		)
	],
	mostaza: [
		plan(
			'cafe-cada-dos-semanas',
			'Café cada 2 semanas',
			'Grano o molido, tostado esta semana. Pausás cuando quieras.',
			'1559056199-641a0ac8b55e',
			14875,
			'quincenal',
			'250 g de Finca La Esperanza, en grano o molido',
			[['cafe-finca-la-esperanza', 1]],
			0
		),
		plan(
			'cafe-doble-origen',
			'Doble origen',
			'Dos orígenes por mes para probar en casa, con 15% menos que comprarlos sueltos.',
			'1447933601403-0c6688de566e',
			33490,
			'mensual',
			'2 bolsas de 250 g: Finca La Esperanza y Yirgacheffe',
			[
				['cafe-finca-la-esperanza', 1],
				['cafe-yirgacheffe', 1]
			],
			1
		)
	]
};

// Quienes ya se suscribieron (nombre, email, plan, días desde que se sumaron, cobros, estado).
const SUSCRIPTORES: Record<string, [string, string, string, number, Suscriptor['estado']][]> = {
	malbec: [
		['Gustavo Rinaldi', 'gustavo.rinaldi@mail.com', 'club-explorador', 120, 'activa'],
		['Sole Ferreyra', 'sole.ferreyra@mail.com', 'club-descubrir', 64, 'activa'],
		['Martín Ochoa', 'martin.ochoa@mail.com', 'club-coleccionista', 200, 'activa'],
		['Paula Ríos', 'paula.rios@mail.com', 'club-explorador', 95, 'pausada'],
		['Diego Sosa', 'diego.sosa@mail.com', 'club-descubrir', 150, 'cancelada']
	],
	mostaza: [
		['Caro Méndez', 'caro.mendez@mail.com', 'cafe-cada-dos-semanas', 300, 'activa'],
		['Juli Bianchi', 'juli.bianchi@mail.com', 'cafe-cada-dos-semanas', 40, 'activa'],
		['Tomás Herrera', 'tomas.herrera@mail.com', 'cafe-doble-origen', 75, 'activa'],
		['Ana Leiva', 'ana.leiva@mail.com', 'cafe-doble-origen', 33, 'pausada']
	]
};

const iso = (t: number) => new Date(t).toISOString();

/** Los planes y suscriptores de ejemplo de la plantilla. */
export function suscripcionesDemo(plantilla: string, creadoEn: number) {
	const planes = PLANES[plantilla] ?? [];
	const porId = new Map(planes.map((p) => [p.id, p]));
	const suscriptores: Suscriptor[] = (SUSCRIPTORES[plantilla] ?? []).map(
		([nombre, email, planId, hace, estado], i) => {
			const p = porId.get(planId)!;
			const periodo = diasDe(p.frecuencia);
			const desde = creadoEn - hace * DIA;
			const cobros = Math.max(1, Math.floor(hace / periodo));
			const ultimo = desde + (cobros - 1) * periodo * DIA;
			return {
				id: `demo-${plantilla}-${i + 1}`,
				planId,
				plan: p.nombre,
				precio: p.precio,
				frecuencia: p.frecuencia,
				nombre,
				email,
				telefono: `11 5${String(100 + i * 37).padStart(3, '0')} ${String(2000 + i * 811).slice(0, 4)}`,
				entrega: i % 3 === 1 ? 'retiro' : 'envio',
				direccion:
					i % 3 === 1 ? '' : ['Av. Corrientes 4120', 'Gurruchaga 1650', 'Thames 980'][i % 3],
				estado,
				proximoCobro: estado === 'activa' ? iso(ultimo + periodo * DIA) : null,
				cobros,
				ultimoCobro: iso(ultimo),
				creadaEn: iso(desde),
				canceladaPor: estado === 'cancelada' ? 'comprador' : null
			};
		}
	);
	return { planes, suscriptores };
}

type Cambios = Map<string, unknown>;

/** Lo de ejemplo con lo que cambió el visitante. */
export function mezclar(c: Cambios, plantilla: string, creadoEn: number) {
	const ej = suscripcionesDemo(plantilla, creadoEn);
	function con<T extends { id: string }>(base: T[], pre: string) {
		const m = new Map(base.map((x) => [x.id, x]));
		for (const [k, v] of c)
			if (k.startsWith(pre)) {
				if (v) m.set(k.slice(pre.length), v as T);
				else m.delete(k.slice(pre.length));
			}
		return [...m.values()];
	}
	const suscriptores = con(ej.suscriptores, 's:').sort((a, b) =>
		a.creadaEn < b.creadaEn ? 1 : a.creadaEn > b.creadaEn ? -1 : 0
	);
	const vigentes = (id: string) =>
		suscriptores.filter((s) => s.planId === id && (s.estado === 'activa' || s.estado === 'pausada'))
			.length;
	const planes: PlanSuscripcion[] = con(ej.planes, 'p:')
		.sort((a, b) => a.orden - b.orden)
		.map((p) => ({ ...p, suscriptores: vigentes(p.id) }));
	return { planes, suscriptores };
}

/** Un plan como lo ve quien compra: sólo los productos que siguen en el catálogo. */
export function planPublico(p: PlanSuscripcion, productos: RegProducto[]): PlanPublico {
	const porId = new Map(productos.map((x) => [x.id, x]));
	return {
		id: p.id,
		nombre: p.nombre,
		descripcion: p.descripcion,
		foto: p.foto,
		precio: p.precio,
		frecuencia: p.frecuencia,
		incluye: p.incluye,
		productos: p.productos.flatMap((x) => {
			const r = porId.get(x.productoId);
			if (!r) return [];
			const v = x.varianteId ? r.variantes.find((y) => y.id === x.varianteId) : null;
			return [
				{
					id: r.id,
					nombre: r.nombre,
					foto: v?.foto ?? r.foto,
					cantidad: x.cantidad,
					variante: v ? textoVariante(r.opciones, v.valores) : null
				}
			];
		}),
		envio: p.envio,
		retiro: p.retiro,
		agotado: p.cupo !== null && p.suscriptores >= p.cupo
	};
}

/** Lo que ve quien se suscribió en su link privado. */
export function publica(
	s: Suscriptor,
	codigo: string,
	p: PlanSuscripcion | undefined,
	tienda: { nombre: string; whatsapp: string }
): SuscripcionPublica {
	return {
		codigo,
		planId: s.planId,
		plan: s.plan,
		foto: p?.foto ?? null,
		incluye: p?.incluye ?? '',
		precio: s.precio,
		frecuencia: s.frecuencia,
		nombre: s.nombre.split(' ')[0],
		entrega: s.entrega,
		direccion: s.direccion,
		estado: s.estado,
		proximoCobro: s.proximoCobro,
		cobros: s.cobros,
		creadaEn: s.creadaEn,
		tienda: tienda.nombre,
		whatsapp: tienda.whatsapp
	};
}

/** El código del link privado (como el de los pedidos: 24 caracteres). */
export function nuevoCodigo() {
	const b = crypto.getRandomValues(new Uint8Array(18));
	return btoa(String.fromCharCode(...b))
		.replace(/\+/g, '-')
		.replace(/\//g, '_')
		.replace(/=+$/, '');
}

export { DIA };
