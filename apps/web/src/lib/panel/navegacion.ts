/**
 * Navegación del panel (handoff: AREAS y SIMPLE). Cada sección tiene su ruta: /panel es el Resumen
 * (o el Inicio del modo Simple) y el resto /panel/<slug>.
 */
import * as I from '#lib/icons.ts';
import type { IconData } from '#lib/icons.ts';
import { PERMISOS, puedeVer } from './permisos.ts';

export type Insignia = 'pedidos' | 'carritos' | 'whatsapp' | 'paraHacer';
export type Item = { label: string; icon: IconData; iconFill: IconData; insignia?: Insignia };
export type Area = {
	id: string;
	label: string;
	icon: IconData;
	iconFill: IconData;
	punto?: boolean;
	items: Item[];
};

const it = (label: string, icon: IconData, iconFill: IconData, insignia?: Insignia): Item => ({
	label,
	icon,
	iconFill,
	insignia
});

export const AREAS: Area[] = [
	{
		id: 'inicio',
		label: 'Inicio',
		icon: I.phHouseSimple,
		iconFill: I.phHouseSimpleFill,
		items: [
			it('Resumen', I.phSquaresFour, I.phSquaresFourFill),
			it('Para hacer hoy', I.phCheckSquareOffset, I.phCheckSquareOffsetFill, 'paraHacer'),
			it('Actividad', I.phPulse, I.phPulseFill)
		]
	},
	{
		id: 'ventas',
		label: 'Ventas',
		icon: I.phShoppingBagOpen,
		iconFill: I.phShoppingBagOpenFill,
		punto: true,
		items: [
			it('Pedidos', I.phPackage, I.phPackageFill, 'pedidos'),
			it('Turnos', I.phCalendarBlank, I.phCalendarBlankFill),
			it('Suscripciones', I.phArrowsClockwise, I.phArrowsClockwiseFill),
			it('Devoluciones', I.phArrowUUpLeft, I.phArrowUUpLeftFill),
			it('Ventas en local', I.phStorefront, I.phStorefrontFill),
			it('Carritos abandonados', I.phShoppingCartSimple, I.phShoppingCartSimpleFill, 'carritos'),
			it('WhatsApp', I.phWhatsappLogo, I.phWhatsappLogoFill, 'whatsapp'),
			it('Clientes', I.phUsersThree, I.phUsersThreeFill),
			it('Reseñas y preguntas', I.phStar, I.phStarFill)
		]
	},
	{
		id: 'catalogo',
		label: 'Catálogo',
		icon: I.phTag,
		iconFill: I.phTagFill,
		items: [
			it('Productos', I.phTag, I.phTagFill),
			it('Inventario', I.phStack, I.phStackFill),
			it('Listas de precio', I.phListDashes, I.phListDashesFill)
		]
	},
	{
		id: 'crecer',
		label: 'Crecer',
		icon: I.phChartLineUp,
		iconFill: I.phChartLineUpFill,
		items: [
			it('Marketing', I.phMegaphone, I.phMegaphoneFill),
			it('Cupones', I.phTicket, I.phTicketFill),
			it('Canales', I.phLinkSimple, I.phLinkSimpleFill),
			it('Informes', I.phChartBar, I.phChartBarFill)
		]
	},
	{
		id: 'operacion',
		label: 'Operación',
		icon: I.phWallet,
		iconFill: I.phWalletFill,
		items: [
			it('Finanzas', I.phWallet, I.phWalletFill),
			it('Facturación', I.phReceipt, I.phReceiptFill),
			it('Envíos', I.phTruck, I.phTruckFill)
		]
	},
	{
		id: 'tienda',
		label: 'Tienda',
		icon: I.phStorefront,
		iconFill: I.phStorefrontFill,
		items: [
			it('Mi tienda', I.phPaintBrushBroad, I.phPaintBrushBroadFill),
			it('Tarjeta para paquetes', I.phEnvelopeSimple, I.phEnvelopeSimpleFill),
			it('Configuración', I.phGearSix, I.phGearSixFill),
			it('Integraciones', I.phPlugsConnected, I.phPlugsConnectedFill)
		]
	},
	{
		id: 'cuenta',
		label: 'Cuenta',
		icon: I.phUserCircle,
		iconFill: I.phUserCircleFill,
		items: [
			it('Plan y facturación', I.phCreditCard, I.phCreditCardFill),
			it('Usuarios y permisos', I.phUsers, I.phUsersFill),
			it('Dominio', I.phGlobe, I.phGlobeFill),
			it('Notificaciones', I.phBell, I.phBellFill),
			it('Mis datos', I.phDatabase, I.phDatabaseFill)
		]
	}
];

/**
 * Secciones que se muestran: Turnos, con turnos activos o rubro de servicios; y sólo las que el rol
 * de la persona puede ver (Go igual responde 403 a lo demás). La ruta existe igual; esto decide qué
 * se muestra en el menú y en ⌘K.
 */
export const visible = (label: string, turnos: boolean, permisos: readonly string[] = PERMISOS) =>
	(label !== 'Turnos' || turnos) && puedeVer(permisos, label);

/** Modo Simple: una sola lista ("Día a día"). */
export const SIMPLE: Item[] = [
	it('Inicio', I.phHouseSimple, I.phHouseSimpleFill),
	it('Pedidos', I.phPackage, I.phPackageFill, 'pedidos'),
	it('Productos', I.phTag, I.phTagFill),
	it('WhatsApp', I.phWhatsappLogo, I.phWhatsappLogoFill, 'whatsapp'),
	it('Clientes', I.phUsersThree, I.phUsersThreeFill),
	it('Cobros', I.phWallet, I.phWalletFill),
	it('Envíos', I.phTruck, I.phTruckFill),
	it('Mi tienda', I.phStorefront, I.phStorefrontFill),
	it('Configuración', I.phGearSix, I.phGearSixFill),
	it('Ayuda', I.phQuestion, I.phQuestionFill)
];

/** "Reseñas y preguntas" → "resenas-y-preguntas". */
export const slug = (label: string) =>
	label
		.normalize('NFD')
		.replace(/[̀-ͯ]/g, '')
		.toLowerCase()
		.replace(/[^a-z0-9]+/g, '-')
		.replace(/^-|-$/g, '');

/** "Cobros" del modo Simple es Finanzas. */
const ALIAS: Record<string, string> = { Cobros: 'Finanzas', Inicio: 'Resumen' };

/** Ruta de una sección. */
export const rutaDe = (label: string) => {
	const l = ALIAS[label] ?? label;
	return l === 'Resumen' ? '/panel' : `/panel/${slug(l)}`;
};

/** Todas las secciones con ruta (las del modo Completo, más Ayuda). */
export const SECCIONES = [...AREAS.flatMap((a) => a.items.map((i) => i.label)), 'Ayuda'];

/** La sección y el área de una ruta del panel. */
export function ubicar(pathname: string): { seccion: string; area: Area } | null {
	const resto = pathname.replace(/^\/panel\/?/, '').split('/')[0];
	const seccion = resto ? SECCIONES.find((s) => slug(s) === resto) : 'Resumen';
	if (!seccion) return null;
	const area = AREAS.find((a) => a.items.some((i) => i.label === seccion)) ?? AREAS[0];
	return { seccion, area };
}
