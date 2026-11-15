<script lang="ts">
	import { browser } from '$app/env';
	import { usarTienda } from '../contexto.ts';
	import { enlace } from '../datos.ts';
	import Cuenta from '../primitivas/Cuenta.svelte';

	/**
	 * La barra de anuncios de arriba de todo (estilo.barra, en todas las páginas). `fija`: los
	 * mensajes uno al lado del otro con un punto en el medio (A-moda; en el celular, el primero);
	 * `rotativa`: de a uno, con flechas y cambio solo cada 5 s, que se frena al pasar el mouse o con
	 * «Reducir movimiento» (B-moda). Sin mensajes no se dibuja.
	 *
	 * Un mensaje con `hasta` lleva la cuenta regresiva y se esconde al vencer (también si la página
	 * quedó abierta).
	 *
	 * Uso: `<Barra />` dentro de marco/Tienda.svelte (ya la pone Tienda; no hace falta en cada página).
	 */
	const ctx = usarTienda();
	const barra = $derived(ctx.estilo.barra);
	// Los vencidos no salen: al cargar (el servidor) y al vencer mientras se mira (la cuenta avisa).
	let vencidos = $state<string[]>([]);
	const ahora = Date.now();
	const mensajes = $derived(
		(barra?.mensajes ?? [])
			.filter((m) => !m.hasta || (Date.parse(m.hasta) > ahora && !vencidos.includes(m.hasta)))
			.map((m) => ({ texto: m.texto, href: enlace(ctx.tienda, m.link), hasta: m.hasta ?? '' }))
	);
	$effect(() => {
		if (i >= mensajes.length) i = 0;
	});
	let i = $state(0);
	let quieta = $state(false);
	const rota = $derived(barra?.variante === 'rotativa' && mensajes.length > 1);

	$effect(() => {
		if (!browser || !rota || quieta) return;
		if (matchMedia('(prefers-reduced-motion: reduce)').matches) return;
		const t = setInterval(() => (i = (i + 1) % mensajes.length), 5000);
		return () => clearInterval(t);
	});
	const mover = (d: number) => (i = (i + d + mensajes.length) % mensajes.length);
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- links que eligió la tienda -->
{#snippet mensaje(m: { texto: string; href: string | null; hasta: string })}
	{#if m.href}<a href={m.href}>{m.texto}</a>{:else}<span>{m.texto}</span>{/if}{#if m.hasta}<span
			class="tv-barra-cuenta"
			>&nbsp;· <Cuenta hasta={m.hasta} alVencer={() => (vencidos = [...vencidos, m.hasta])} /></span
		>{/if}
{/snippet}

{#if mensajes.length}
	<div
		class="tv-barra"
		data-fondo={barra?.fondo ?? 'oscuro'}
		role="region"
		aria-label="Anuncios"
		onmouseenter={() => (quieta = true)}
		onmouseleave={() => (quieta = false)}
		onfocusin={() => (quieta = true)}
		onfocusout={() => (quieta = false)}
	>
		{#if rota}
			<button
				type="button"
				class="tv-barra-flecha"
				aria-label="Anuncio anterior"
				onclick={() => mover(-1)}>‹</button
			>
			<p class="tv-barra-msj" aria-live="polite">{@render mensaje(mensajes[i])}</p>
			<button
				type="button"
				class="tv-barra-flecha"
				aria-label="Anuncio siguiente"
				onclick={() => mover(1)}>›</button
			>
		{:else}
			<p class="tv-barra-msj">
				{#each mensajes as m, j (j)}{#if j}<span class="tv-barra-punto" aria-hidden="true">·</span
						>{/if}<span class="tv-barra-item" class:tv-solo-compu={j > 0}>{@render mensaje(m)}</span
					>{/each}
			</p>
		{/if}
	</div>
{/if}
