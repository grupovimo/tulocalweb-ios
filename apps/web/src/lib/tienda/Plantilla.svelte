<script lang="ts">
	import Carrito from './sistema/paginas/Carrito.svelte';
	import Catalogo from './sistema/paginas/Catalogo.svelte';
	import Checkout from './sistema/paginas/Checkout.svelte';
	import Favoritos from './sistema/paginas/Favoritos.svelte';
	import Ficha from './sistema/paginas/Ficha.svelte';
	import Inicio from './sistema/paginas/Inicio.svelte';
	import Listo, { type PedidoListo } from './sistema/paginas/Listo.svelte';
	import MiSuscripcion from './sistema/paginas/MiSuscripcion.svelte';
	import NoExiste from './sistema/paginas/NoExiste.svelte';
	import Nosotros from './sistema/paginas/Nosotros.svelte';
	import Suscripcion from './sistema/paginas/Suscripcion.svelte';
	import Suscripciones from './sistema/paginas/Suscripciones.svelte';
	import Turno from './sistema/paginas/Turno.svelte';
	import Turnos from './sistema/paginas/Turnos.svelte';
	import type {
		PlanPublico,
		ProductoPublico,
		ProximosTurnos,
		SuscripcionPublica,
		TiendaPublica,
		TurnoPublico
	} from './tipos.ts';

	/**
	 * La tienda online: cada página dibujada con el sistema de bloques (lib/tienda/sistema, diseño
	 * v2: `tienda.diseno`, o migrado al vuelo si no viene). Las rutas de /t/[tienda] y la vista
	 * previa de Mi tienda le pasan la página y lo que esa página necesita.
	 */
	type Pagina =
		| 'inicio'
		| 'catalogo'
		| 'producto'
		| 'carrito'
		| 'checkout'
		| 'nosotros'
		| 'favoritos'
		| 'no-existe'
		| 'listo'
		| 'turnos'
		| 'turno'
		| 'suscripcion'
		| 'suscripciones'
		| 'mi-suscripcion';

	let {
		tienda,
		pagina,
		producto = null,
		pedido = null,
		turno = null,
		proximos = null,
		aviso = null,
		plan = null,
		suscripcion = null,
		demo = false,
		error = null
	}: {
		tienda: TiendaPublica;
		pagina: Pagina;
		producto?: ProductoPublico | null;
		pedido?: PedidoListo | null;
		/** La confirmación de un turno (pagina 'turno'). */
		turno?: TurnoPublico | null;
		/** Los libres de hoy y el próximo de cada profesional (portada). */
		proximos?: ProximosTurnos | null;
		/** Vuelta de Mercado Pago (la seña del turno) o el alta de una suscripción. */
		aviso?: string | null;
		/** El plan de suscripción (pagina 'suscripcion', §51). */
		plan?: PlanPublico | null;
		/** La suscripción del link privado (pagina 'mi-suscripcion'). */
		suscripcion?: SuscripcionPublica | null;
		/** Es la demo: suscribirse no cobra (se dice al lado del botón). */
		demo?: boolean;
		error?: string | null;
	} = $props();
</script>

{#if pagina === 'inicio'}
	<Inicio {tienda} {proximos} />
{:else if pagina === 'catalogo'}
	<Catalogo {tienda} />
{:else if pagina === 'producto' && producto}
	<Ficha {tienda} {producto} />
{:else if pagina === 'carrito'}
	<Carrito {tienda} />
{:else if pagina === 'checkout'}
	<Checkout {tienda} {error} />
{:else if pagina === 'listo' && pedido}
	<Listo {tienda} {pedido} />
{:else if pagina === 'favoritos'}
	<Favoritos {tienda} />
{:else if pagina === 'nosotros'}
	<Nosotros {tienda} />
{:else if pagina === 'turnos' && tienda.turnos}
	<Turnos {tienda} {error} />
{:else if pagina === 'turno' && turno}
	<Turno {tienda} {turno} {aviso} {error} />
{:else if pagina === 'suscripcion' && plan}
	<Suscripcion {tienda} {plan} {demo} {error} />
{:else if pagina === 'suscripciones'}
	<Suscripciones {tienda} />
{:else if pagina === 'mi-suscripcion' && suscripcion}
	<MiSuscripcion {tienda} {suscripcion} {aviso} {demo} {error} />
{:else}
	<NoExiste {tienda} />
{/if}
