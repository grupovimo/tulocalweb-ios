<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { insignia } from './notificaciones.ts';
	import { getShell } from './shell.svelte.ts';

	/** Barra de arriba (handoff): menú, dónde estás, estado de la tienda, tema y notificaciones. */
	let {
		area,
		seccion,
		foto,
		sinLeer,
		caida = false
	}: {
		area: string;
		seccion: string;
		foto: string | null;
		/** Notificaciones sin leer de la persona (§45). */
		sinLeer: number;
		/** Sin conexión de tiempo real hace más de unos segundos (§44). */
		caida?: boolean;
	} = $props();
	const shell = getShell();
	const numero = $derived(insignia(sinLeer));
</script>

<header class="top">
	<button
		type="button"
		class="ico menu"
		aria-label="Menú"
		aria-expanded={shell.barra}
		onclick={() => (shell.barra = true)}><Icon icon={I.phSidebarSimple} size={19} /></button
	>
	<span class="chico">
		{#if foto}<img src={foto} alt="" width="30" height="30" />{/if}
		<span class="sec">{seccion}</span>
	</span>
	<nav class="miga" aria-label="Ubicación">
		<span>{area}</span><Icon icon={I.phCaretRight} size={12} /><span class="actual">{seccion}</span>
	</nav>
	<span class="esp"></span>
	<button
		type="button"
		class="ico solo-chico"
		aria-label="Buscar"
		onclick={() => (shell.cmdk = true)}><Icon icon={I.phMagnifyingGlass} size={17} /></button
	>
	{#if caida}
		<span class="en-linea caida" role="status"
			><span class="punto"><span></span></span>Reconectando…</span
		>
	{:else}
		<span class="en-linea"
			><span class="punto"><span class="pulso"></span><span></span></span>Tienda en línea</span
		>
	{/if}
	<button
		type="button"
		class="ico"
		title="Cambiar tema"
		aria-label={shell.tema === 'light' ? 'Pasar a tema oscuro' : 'Pasar a tema claro'}
		onclick={() => shell.cambiarTema()}
		><Icon icon={shell.tema === 'light' ? I.phMoon : I.phSun} size={17} /></button
	>
	<button
		type="button"
		class="ico"
		aria-label={numero ? `Notificaciones, ${sinLeer} sin leer` : 'Notificaciones'}
		aria-expanded={shell.notificaciones}
		onclick={() => (shell.notificaciones = !shell.notificaciones)}
		><Icon icon={I.phBell} size={17} />{#if numero}<span class="badge" aria-hidden="true"
				>{numero}</span
			>{/if}</button
	>
</header>

<style>
	.top {
		position: sticky;
		top: 0;
		z-index: 10;
		height: 64px;
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 0 36px;
		background: var(--bga);
		backdrop-filter: blur(14px);
		-webkit-backdrop-filter: blur(14px);
	}
	.ico {
		position: relative;
		width: 38px;
		height: 38px;
		flex: none;
		border-radius: 999px;
		border: 0;
		background: var(--sf);
		color: var(--tx);
		display: grid;
		place-items: center;
		box-shadow: var(--sh);
	}
	.menu {
		width: 40px;
		height: 40px;
		border-radius: 12px;
		display: none;
	}
	/* La insignia de la barra de abajo (BarraAbajo), sobre la campanita. */
	.badge {
		position: absolute;
		top: -4px;
		right: -5px;
		min-width: 18px;
		height: 18px;
		padding: 0 5px;
		border-radius: 999px;
		background: #b0441d;
		color: #fff;
		font-size: 10.5px;
		font-weight: 700;
		display: grid;
		place-items: center;
		box-shadow: 0 0 0 2px var(--bg);
	}
	.chico {
		display: none;
		align-items: center;
		gap: 8px;
		min-width: 0;
	}
	.chico img {
		border-radius: 9px;
		object-fit: contain;
		background: var(--sf);
		flex: none;
	}
	.sec {
		font-size: 15px;
		font-weight: 600;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.miga {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 14px;
		color: var(--mu);
		white-space: nowrap;
		min-width: 0;
	}
	.actual {
		color: var(--tx);
		font-weight: 600;
	}
	.esp {
		flex: 1;
	}
	.solo-chico {
		display: none;
	}
	.en-linea {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 13px;
		font-weight: 500;
		color: var(--mu);
		padding: 7px 12px;
		border-radius: 999px;
		background: var(--sf);
		box-shadow: var(--sh);
		white-space: nowrap;
	}
	.punto {
		position: relative;
		width: 7px;
		height: 7px;
	}
	.punto span {
		position: absolute;
		inset: 0;
		border-radius: 50%;
		background: #2f8f63;
	}
	.pulso {
		animation: pulso 2s ease-out infinite;
	}
	.caida .punto span {
		background: #b8860b;
	}
	@keyframes pulso {
		to {
			transform: scale(2.6);
			opacity: 0;
		}
	}
	@media (prefers-reduced-motion: reduce) {
		.pulso {
			animation: none;
		}
	}
	@media (max-width: 1099px) {
		.top {
			padding: 0 24px;
		}
		.menu {
			display: grid;
		}
	}
	@media (max-width: 759px) {
		.top {
			padding: 0 12px;
			gap: 8px;
		}
		.chico,
		.solo-chico {
			display: flex;
		}
		.solo-chico {
			display: grid;
		}
		.miga,
		.en-linea {
			display: none;
		}
		/* Sin conexión se ve también en el celular (es lo único de este cartel que importa ahí). */
		/* En el celular no le saca lugar a la barra: va debajo, centrado. */
		.en-linea.caida {
			display: flex;
			position: absolute;
			top: 58px;
			left: 50%;
			transform: translateX(-50%);
			padding: 6px 12px;
			font-size: 12.5px;
		}
	}
</style>
