/** Los planes de Tu Local Web (PanelCuenta.dc.html): los mismos precios que la landing. */
export const PLANES = [
	{
		id: 'gratis',
		nombre: 'Gratis',
		lema: 'Para probar y empezar',
		precio: 0,
		incluye: ['Productos ilimitados', 'Tincho: 100 respuestas con IA', '1 persona · subdominio'],
		personas: 1,
		hornero: 100
	},
	{
		id: 'emprende',
		nombre: 'Emprende',
		lema: 'Para vender todos los días',
		precio: 19_900,
		incluye: [
			'Dominio propio y factura ARCA',
			'Todos los medios de pago y envío',
			'Tincho: 1.000 respuestas con IA',
			'2 personas'
		],
		personas: 2,
		hornero: 1000
	},
	{
		id: 'pro',
		nombre: 'Pro',
		lema: 'Para vender en todos tus canales',
		precio: 49_900,
		incluye: [
			'Mercado Libre, Instagram y local',
			'Acciones masivas y estadísticas',
			'Tincho: 3.000 respuestas con IA',
			'5 personas con permisos'
		],
		personas: 5,
		hornero: 3000
	},
	{
		id: 'negocio',
		nombre: 'Negocio',
		lema: 'Para varias sucursales',
		precio: 149_900,
		incluye: [
			'Sucursales y cajas ilimitadas',
			'API, webhooks y mayoristas',
			'Tincho: 15.000 respuestas con IA',
			'Personas ilimitadas'
		],
		personas: null,
		hornero: 15_000
	}
] as const;
export type IdPlan = (typeof PLANES)[number]['id'];
/** Pagando por año, 25 % menos. */
export const precioAnual = (mensual: number) => Math.round(mensual * 0.75);

/** El plan que rige hoy: durante la prueba gratis, una tienda en Gratis usa Pro. */
export const planVigente = (actual: string, enPrueba: boolean) =>
	actual === 'gratis' && enPrueba ? 'pro' : actual;

/**
 * El descuento para quedarse (§57, abono.Ofertas de la API) por motivo de baja: «es caro» → 50 % por 3
 * cobros; «otro» → 50 % en el próximo. Se cobra de verdad (baja el débito) y va una vez por tienda.
 */
export type Beneficio = { pct: number; cobros: number };
export const BENEFICIOS: Record<number, Beneficio> = {
	0: { pct: 50, cobros: 3 },
	4: { pct: 50, cobros: 1 }
};
export const conDescuento = (importe: number, b: Beneficio | null) =>
	b && b.cobros > 0 && b.pct > 0 ? Math.round((importe * (100 - b.pct)) / 100) : importe;
