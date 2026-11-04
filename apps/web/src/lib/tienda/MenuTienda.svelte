<script lang="ts">
	import { afterNavigate } from '$app/navigation';
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { carritoDe } from './carrito.svelte.ts';
	import type { TiendaPublica } from './tipos.ts';

	/**
	 * El menú del celular en todas las plantillas: el ícono abre un panel con la navegación de la
	 * plantilla (nav), Inicio, WhatsApp y el carrito. Toma el fondo, la letra y la fuente que tiene
	 * el encabezado en ese momento, así respeta la plantilla y Mi tienda › Estilo.
	 */
	let {
		tienda,
		base,
		nav,
		estilo = 'display:grid',
		tam = 22,
		negrita = false
	}: {
		tienda: Pick<TiendaPublica, 'slug' | 'nombre' | 'whatsapp'>;
		base: string;
		nav: { t: string; href: string; on?: boolean }[];
		estilo?: string;
		tam?: number;
		/** El ícono de la plantilla en negrita (Miga, Mostaza, Pampa, Tijera). */
		negrita?: boolean;
	} = $props();

	let abierto = $state(false);
	let boton = $state<HTMLButtonElement>();
	let colores = $state({ fondo: '#fff', texto: '#111', fuente: 'inherit' });
	const carrito = $derived(carritoDe(tienda.slug));

	/** El primer fondo opaco desde el botón hacia arriba (el encabezado suele ser translúcido). */
	function leerColores() {
		if (!boton) return;
		let el: HTMLElement | null = boton;
		let fondo = '';
		while (el && !fondo) {
			const bg = getComputedStyle(el).backgroundColor;
			const m = bg.match(/rgba?\(([^)]+)\)/);
			const a = m ? Number(m[1].split(/[\s,/]+/)[3] ?? 1) : 0;
			if (m && a > 0.5)
				fondo = `rgb(${m[1]
					.split(/[\s,/]+/)
					.slice(0, 3)
					.join(',')})`;
			el = el.parentElement;
		}
		const cs = getComputedStyle(boton);
		colores = { fondo: fondo || '#fff', texto: cs.color, fuente: cs.fontFamily };
	}
	function abrir() {
		leerColores();
		abierto = true;
	}
	afterNavigate(() => (abierto = false));
	/**
	 * El panel va directo al body: los encabezados con backdrop-filter o transform hacen que un
	 * position:fixed quede atrapado adentro (se veía debajo del contenido).
	 */
	function alBody(nodo: HTMLElement) {
		document.body.appendChild(nodo);
		return { destroy: () => nodo.remove() };
	}
	const wa = $derived(tienda.whatsapp.replace(/\D/g, ''));
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->

<svelte:window onkeydown={(e) => e.key === 'Escape' && (abierto = false)} />

<button
	bind:this={boton}
	type="button"
	aria-label="Menú"
	aria-expanded={abierto}
	style="{estilo};border:0;background:none;padding:0;cursor:pointer;color:inherit"
	onclick={abrir}><Icon icon={negrita ? I.phListBold : I.phList} size={tam} /></button
>

{#if abierto}
	<div class="velo" role="presentation" use:alBody onclick={() => (abierto = false)}></div>
	<div
		use:alBody
		class="cajon"
		role="dialog"
		aria-modal="true"
		aria-label="Menú de {tienda.nombre}"
		style:background={colores.fondo}
		style:color={colores.texto}
		style:font-family={colores.fuente}
	>
		<div class="cab">
			<b>{tienda.nombre}</b>
			<button
				type="button"
				class="cerrar"
				aria-label="Cerrar el menú"
				onclick={() => (abierto = false)}><Icon icon={I.phXBold} size={20} /></button
			>
		</div>
		<nav>
			<a href={base || '/'}>Inicio</a>
			{#each nav as n (n.href + n.t)}<a href={n.href} aria-current={n.on ? 'page' : undefined}
					>{n.t}</a
				>{/each}
		</nav>
		<div class="pie">
			<a class="accion" href="{base}/carrito"
				><Icon
					icon={I.phHandbag}
					size={18}
				/>Carrito{#if carrito.unidades}&nbsp;({carrito.unidades}){/if}</a
			>
			{#if wa}<a class="accion" href="https://wa.me/{wa}" target="_blank" rel="noopener"
					><Icon icon={I.phWhatsappLogoFill} size={18} />WhatsApp</a
				>{/if}
		</div>
	</div>
{/if}

<style>
	.velo {
		position: fixed;
		inset: 0;
		z-index: 90;
		background: rgba(0, 0, 0, 0.4);
		animation: aparece 0.18s ease-out;
	}
	.cajon {
		position: fixed;
		inset: 0 auto 0 0;
		z-index: 91;
		width: min(320px, 86vw);
		display: flex;
		flex-direction: column;
		padding: 18px 20px calc(20px + env(safe-area-inset-bottom));
		box-shadow: 8px 0 40px rgba(0, 0, 0, 0.25);
		overflow-y: auto;
		animation: entra 0.22s cubic-bezier(0.2, 0.8, 0.2, 1);
	}
	.cab {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
		padding-bottom: 14px;
		font-size: 18px;
	}
	.cerrar {
		display: grid;
		place-items: center;
		width: 40px;
		height: 40px;
		margin-right: -8px;
		border: 0;
		border-radius: 999px;
		background: none;
		color: inherit;
		cursor: pointer;
	}
	nav {
		display: flex;
		flex-direction: column;
		border-top: 1px solid color-mix(in srgb, currentColor 14%, transparent);
	}
	nav a {
		padding: 14px 2px;
		font-size: 17px;
		font-weight: 500;
		color: inherit;
		text-decoration: none;
		border-bottom: 1px solid color-mix(in srgb, currentColor 14%, transparent);
	}
	nav a[aria-current='page'] {
		font-weight: 700;
	}
	.pie {
		display: flex;
		flex-direction: column;
		gap: 8px;
		margin-top: auto;
		padding-top: 20px;
	}
	.accion {
		display: flex;
		align-items: center;
		gap: 10px;
		min-height: 46px;
		padding: 0 14px;
		border-radius: 12px;
		background: color-mix(in srgb, currentColor 8%, transparent);
		color: inherit;
		text-decoration: none;
		font-weight: 600;
	}
	@keyframes entra {
		from {
			transform: translateX(-100%);
		}
	}
	@keyframes aparece {
		from {
			opacity: 0;
		}
	}
	@media (prefers-reduced-motion: reduce) {
		.cajon,
		.velo {
			animation: none;
		}
	}
</style>
