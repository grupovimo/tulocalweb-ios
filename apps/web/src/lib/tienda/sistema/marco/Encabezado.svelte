<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import LogoTienda from '../../LogoTienda.svelte';
	import { usarTienda } from '../contexto.ts';
	import { inicial } from '../datos.ts';
	import type { Encabezado } from '../presets.ts';
	import Buscador from './Buscador.svelte';
	import MenuMovil from './MenuMovil.svelte';

	/**
	 * El encabezado de la tienda (estilo.encabezado), igual en todas las páginas:
	 * - `centrado` (Vidriera, A-moda): menú a la izquierda, el nombre o el logo al medio, «Nosotros» y
	 *   la bolsa a la derecha; 84 px.
	 * - `izquierda`: logo y menú a la izquierda, la bolsa a la derecha.
	 * - `buscador` (Mostrador, B-moda): logo, buscador píldora y carrito oscuro; abajo, las pestañas.
	 * En Pedido es siempre el compacto de 64 px (C-moda): logo, una línea de la tienda, buscador y
	 * bolsa. En el celular: menú, marca y bolsa en 56 px, y el buscador abajo si corresponde.
	 * Debajo, el aviso de la vitrina («Es una tienda de ejemplo…»).
	 *
	 * Uso: lo pone marco/Tienda.svelte. Props: `variante` para forzar una (si no, la del estilo).
	 * Lleva `data-t-enc` (los colores de la zona: --t-enc y --t-enc-texto).
	 */
	let { variante }: { variante?: Encabezado } = $props();
	const ctx = usarTienda();
	const v = $derived(variante ?? ctx.estilo.encabezado);
	const t = $derived(ctx.tienda);
	const unidades = $derived(ctx.carrito.unidades);
	const nombreBolsa = $derived(
		ctx.estilo.gastronomia ? 'Pedido' : ctx.dir === 'mostrador' ? 'Carrito' : 'Bolsa'
	);
	const conBuscador = $derived(
		(v === 'buscador' || ctx.dir === 'pedido') && t.productos.length > 1
	);
	// En el centrado, «Nosotros» va a la derecha (A-moda); el resto del menú, a la izquierda.
	const nosotros = $derived(ctx.nav.find((n) => n.href.endsWith('/nosotros')) ?? null);
	const menu = $derived(v === 'izquierda' ? ctx.nav : ctx.nav.filter((n) => n !== nosotros));
	const linea = $derived(
		[
			t.ciudad,
			t.envios.correo ? 'envíos a todo el país' : t.envios.moto ? 'envío en moto' : '',
			t.envios.retiro ? 'retiro en el local' : ''
		]
			.filter(Boolean)
			.join(' · ')
	);
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
{#snippet marca()}
	<a class="tv-marca" href={ctx.base || '/'} aria-label="{t.nombre}, inicio">
		<LogoTienda tienda={t} alto={ctx.dir === 'pedido' ? 32 : 40}>
			{#if ctx.dir !== 'vidriera'}<span class="tv-marca-inicial" aria-hidden="true"
					>{inicial(t.nombre)}</span
				>{/if}<span class="tv-marca-nombre">{t.nombre}</span>
		</LogoTienda>
	</a>
{/snippet}

{#snippet nav(items: typeof ctx.nav, clase: string)}
	{#if items.length}
		<nav class="tv-nav {clase}" aria-label="Principal">
			{#each items as n (n.href + n.t)}<a href={n.href} aria-current={n.on ? 'page' : undefined}
					>{n.t}</a
				>{/each}
		</nav>
	{/if}
{/snippet}

{#snippet bolsa()}
	<a
		class="tv-bolsa"
		href="{ctx.base}/carrito"
		aria-label="{nombreBolsa}{unidades ? `, ${unidades} ${unidades === 1 ? 'producto' : 'productos'}` : ', vacío'}"
	>
		<Icon icon={I.phHandbag} size={20} />
		<span class="tv-bolsa-texto">{nombreBolsa}</span>
		<span class="tv-bolsa-n" class:tv-bolsa-n--vacia={!unidades}>{unidades}</span>
	</a>
{/snippet}

<header class="tv-enc" data-variante={ctx.dir === 'pedido' ? 'pedido' : v} data-t-enc>
	<div class="tv-enc-fila tv-ancho">
		<MenuMovil />
		<!-- En el celular, reservar turno a mano (en la compu va en el menú). Va primero en el DOM
		para que el lector lo encuentre antes que el menú escondido; se ubica al lado de la bolsa. -->
		{#if t.turnos}<a class="tv-enc-turnos tv-solo-movil" href="{ctx.base}/turnos" aria-label="Reservar turno"
				><Icon icon={I.phCalendarPlusBold} size={22} /></a
			>{/if}
		{#if ctx.dir === 'pedido'}
			{@render marca()}
			{#if linea}<span class="tv-enc-linea tv-solo-compu">{linea}</span>{/if}
			{#if conBuscador}<div class="tv-enc-buscar tv-solo-compu"><Buscador /></div>{/if}
			<div class="tv-enc-acciones">{@render bolsa()}</div>
		{:else if v === 'centrado'}
			<!-- Sin pestañas, igual ocupa su columna: el logo queda al medio. -->
			{#if menu.length}{@render nav(menu, 'tv-solo-compu')}{:else}<span class="tv-solo-compu"></span>{/if}
			{@render marca()}
			<div class="tv-enc-acciones">
				{#if nosotros}<a class="tv-solo-compu tv-enc-link" href={nosotros.href}>{nosotros.t}</a>{/if}
				{@render bolsa()}
			</div>
		{:else if v === 'buscador'}
			{@render marca()}
			{#if conBuscador}<div class="tv-enc-buscar tv-solo-compu"><Buscador /></div>{/if}
			<div class="tv-enc-acciones">
				{#if nosotros}<a class="tv-solo-compu tv-enc-link" href={nosotros.href}>{nosotros.t}</a>{/if}
				{@render bolsa()}
			</div>
		{:else}
			{@render marca()}
			{@render nav(ctx.nav, 'tv-solo-compu')}
			<div class="tv-enc-acciones">{@render bolsa()}</div>
		{/if}
	</div>
	{#if v === 'buscador' && ctx.dir !== 'pedido' && menu.length}
		<div class="tv-enc-pestanas tv-solo-compu">
			<div class="tv-ancho">{@render nav(menu, '')}</div>
		</div>
	{/if}
	{#if conBuscador}<div class="tv-enc-buscar-movil tv-solo-movil"><Buscador id="tv-buscar-m" /></div>{/if}
</header>

{#if t.vitrina}
	<p class="tv-vitrina">
		Es una tienda de ejemplo: podés recorrerla, pero no se compra de&nbsp;verdad.
	</p>
{/if}
