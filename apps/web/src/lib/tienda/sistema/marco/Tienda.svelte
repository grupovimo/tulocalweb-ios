<script lang="ts">
	import type { Snippet } from 'svelte';
	import { carritoDe } from '../../carrito.svelte.ts';
	import { menuPropio } from '../../estilo/menu.ts';
	import type { TiendaPublica } from '../../tipos.ts';
	import { ponerTienda, type ItemNav } from '../contexto.ts';
	import { baseDe, categoriaHref } from '../datos.ts';
	import { disenoDeTienda } from '../migrar.ts';
	import { estiloFinal } from '../presets.ts';
	import { hojaFuentes, variables } from '../tokens.ts';
	import Barra from './Barra.svelte';
	import BotonWhatsApp from './BotonWhatsApp.svelte';
	import Encabezado from './Encabezado.svelte';
	import Pie from './Pie.svelte';
	import '../base.css';
	import '../direcciones/vidriera.css';
	import '../direcciones/mostrador.css';
	import '../direcciones/pedido.css';

	/**
	 * El marco de cualquier página de la tienda en el sistema de bloques: pone el contexto (tienda,
	 * estilo final, dirección, links, menú y carrito), las variables CSS de la tienda, las fuentes, y
	 * dibuja Barra, Encabezado, la página (children), Pie y el botón de WhatsApp.
	 *
	 * Cómo se arma una página (paginas/Inicio.svelte es el ejemplo):
	 *
	 *   <Tienda {tienda} pagina="catalogo">
	 *     <Seccion titulo="Productos"><Grilla items={lista}>{#snippet children(p)}
	 *       <TarjetaProducto producto={p} />{/snippet}</Grilla></Seccion>
	 *   </Tienda>
	 *
	 * Adentro, cualquier pieza lee `usarTienda()` (contexto.ts). Props: `sinWhatsApp` saca el botón
	 * flotante (la barra fija del pedido ocupa ese lugar).
	 */
	let {
		tienda,
		pagina,
		sinWhatsApp = false,
		children
	}: {
		tienda: TiendaPublica;
		pagina: string;
		sinWhatsApp?: boolean;
		children: Snippet;
	} = $props();

	const diseno = $derived(disenoDeTienda(tienda));
	const estilo = $derived(estiloFinal(tienda.plantilla, diseno.paleta ?? tienda.paleta, diseno.estilo));
	const base = $derived(baseDe(tienda));
	const carrito = $derived(carritoDe(tienda.slug));
	const nav = $derived.by((): ItemNav[] => {
		const propio = menuPropio({ estilo: diseno.estilo ?? tienda.estilo }, base);
		if (propio) return propio;
		const cats = tienda.categorias.filter((c) => c.cantidad > 0);
		return [
			...(tienda.productos.length > 1
				? [{ t: cats.length > 1 ? 'Todo' : 'Productos', href: `${base}/catalogo`, on: pagina === 'catalogo' }]
				: []),
			...(cats.length > 1
				? cats.slice(0, 4).map((c) => ({ t: c.nombre, href: categoriaHref(tienda, c.nombre) }))
				: []),
			...(tienda.turnos ? [{ t: 'Turnos', href: `${base}/turnos`, on: pagina === 'turnos' }] : []),
			...(tienda.suscripciones.length
				? [{ t: 'Suscripciones', href: `${base}/suscripciones`, on: pagina === 'suscripciones' }]
				: []),
			{ t: 'Nosotros', href: `${base}/nosotros`, on: pagina === 'nosotros' }
		];
	});

	ponerTienda({
		get tienda() {
			return tienda;
		},
		get estilo() {
			return estilo;
		},
		get dir() {
			return estilo.direccion;
		},
		get base() {
			return base;
		},
		get nav() {
			return nav;
		},
		get carrito() {
			return carrito;
		},
		get pagina() {
			return pagina;
		}
	});
	const fuentes = $derived(hojaFuentes(estilo));
</script>

<svelte:head>
	{#if fuentes}<link rel="stylesheet" href={fuentes} />{/if}
</svelte:head>

<div
	class="tv"
	data-dir={estilo.direccion}
	data-forma={estilo.forma}
	data-densidad={estilo.densidad}
	style={variables(estilo)}
>
	<Barra />
	<Encabezado />
	<main class="tv-main">
		{@render children()}
	</main>
	<Pie />
	<!-- En Pedido, con algo en la bolsa, la barra fija del celular ocupa su lugar (y la bolsa del
	costado ya ofrece consultar por WhatsApp). -->
	<BotonWhatsApp oculto={sinWhatsApp || (estilo.direccion === 'pedido' && carrito.unidades > 0)} />
</div>
