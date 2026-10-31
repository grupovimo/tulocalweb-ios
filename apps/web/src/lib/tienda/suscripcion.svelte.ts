/**
 * Lo que comparten las plantillas para dibujar un plan de suscripción y su alta (§51): cómo lo
 * recibe quien se suscribe, cuánto saldrían sueltos los productos del plan (para el «Descuento
 * socio» del club de Malbec y el «ahorrá» de Mostaza) y el estado del link privado. Cada plantilla
 * lo dibuja con sus piezas; la plata la calcula el servidor (esto sólo muestra).
 */
import { enhance } from '$app/forms';
import type { PlanPublico, SuscripcionPublica, TiendaPublica } from './tipos.ts';

/** El estado del formulario de alta: entrega elegida y si se está mandando. */
export class AltaPlan {
	entrega = $state<'envio' | 'retiro'>('envio');
	enviando = $state(false);

	constructor(plan: PlanPublico) {
		this.entrega = plan.envio ? 'envio' : 'retiro';
	}

	/** `use:alta.form` en el <form method="POST">: Mercado Pago es otro sitio, se navega de verdad. */
	form = (nodo: HTMLFormElement) =>
		enhance(nodo, () => {
			this.enviando = true;
			return async ({ result, update }) => {
				this.enviando = false;
				if (result.type === 'redirect' && result.location.startsWith('https://'))
					return void (location.href = result.location);
				await update();
			};
		});
}

/** Cuánto saldrían sueltos los productos del plan (con el precio de la tienda) y cuánto ahorra. */
export function cajaDe(plan: PlanPublico, tienda: Pick<TiendaPublica, 'productos'>) {
	let sueltos = 0;
	let unidades = 0;
	for (const x of plan.productos) {
		const p = tienda.productos.find((y) => y.id === x.id);
		if (p) sueltos += p.precio * x.cantidad;
		unidades += x.cantidad;
	}
	const ahorro = sueltos > plan.precio ? sueltos - plan.precio : 0;
	return {
		/** La suma suelta; 0 si no se sabe (productos fuera de la tienda) o no conviene mostrarla. */
		sueltos: ahorro ? sueltos : 0,
		ahorro,
		pct: ahorro ? Math.round((ahorro / sueltos) * 100) : 0,
		unidades,
		fotos: plan.productos.flatMap((x) => (x.foto ? [x.foto] : []))
	};
}

/** Lo que se puede hacer desde el link privado, según el estado. */
export function accionesDe(s: Pick<SuscripcionPublica, 'estado'>) {
	return {
		pausar: s.estado === 'activa',
		reanudar: s.estado === 'pausada',
		cancelar: s.estado !== 'cancelada'
	};
}

/** El título y el texto del link privado, según el estado (y si recién vuelve de Mercado Pago). */
export function textosSuscripcion(s: SuscripcionPublica, alta: boolean, demo: boolean) {
	switch (s.estado) {
		case 'pendiente':
			return {
				titulo: `Falta un paso, ${s.nombre}`,
				texto: alta
					? 'Estamos esperando que Mercado Pago confirme el débito automático. Apenas lo haga, te llega un mail.'
					: 'Todavía no autorizaste el débito automático en Mercado Pago. Sin eso no se cobra ni se prepara nada.'
			};
		case 'activa':
			return {
				titulo: alta ? `¡Listo, ${s.nombre}!` : `Hola, ${s.nombre}`,
				texto: demo
					? 'Es la demo: no se cobra nada. En una tienda de verdad, cada cobro de Mercado Pago arma tu pedido solo.'
					: 'Tu suscripción está activa. Cada cobro de Mercado Pago arma tu pedido y te avisamos por mail.'
			};
		case 'pausada':
			return {
				titulo: 'Suscripción pausada',
				texto: 'No se te cobra ni se prepara nada hasta que la reanudes.'
			};
		default:
			return {
				titulo: 'Suscripción cancelada',
				texto: 'No se te va a cobrar más. Si querés volver, te suscribís de nuevo cuando quieras.'
			};
	}
}
