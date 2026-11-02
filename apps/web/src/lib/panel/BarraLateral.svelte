<script lang="ts">
	import { afterNavigate } from '$app/navigation';
	import Icon from '#lib/components/Icon.svelte';
	import MarcaIcono from '#lib/components/MarcaIcono.svelte';
	import * as I from '#lib/icons.ts';
	import type { components } from '@mostrador/contrato/esquema';
	import { AREAS, rutaDe, visible, type Area, type Item } from './navegacion.ts';
	import { NOMBRE_ROL, type Rol } from './permisos.ts';
	import { getShell } from './shell.svelte.ts';

	/**
	 * Barra lateral (handoff): la tienda, buscar, las secciones del área, los canales conectados y la
	 * prueba. Fija desde 1100 px; más angosta, se abre encima. En el celular (sin rail) las áreas van
	 * en acordeón —tocar una despliega sus secciones— y todo el menú se desplaza junto; la prueba queda
	 * en una línea al final. El cambio Simple/Completo del handoff se sacó: confundía más de lo que
	 * ayudaba (Nahuel, 23/10).
	 */
	let {
		area,
		seccion,
		tienda,
		foto,
		dominio,
		resumen
	}: {
		area: Area;
		seccion: string;
		tienda: string;
		foto: string | null;
		dominio: string;
		resumen: components['schemas']['ResumenPanel'];
	} = $props();
	const shell = getShell();
	// Una persona con varias tiendas elige cuál ver: la sesión pasa a esa tienda y el panel se carga
	// de nuevo entero (nada de la otra tienda queda en pantalla).
	let eligiendo = $state(false);
	async function elegir(id: string) {
		if (id === shell.tiendaId) return void (eligiendo = false);
		const r = await fetch('/panel/tienda', {
			method: 'PUT',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ tiendaId: id })
		});
		if (r.ok) location.assign('/panel');
	}
	// Sin foto de la tienda, sus iniciales (nunca una foto de stock).
	const iniciales = $derived(
		tienda
			.split(/\s+/)
			.filter(Boolean)
			.map((w) => w[0])
			.slice(0, 2)
			.join('')
			.toUpperCase()
	);
	// Celular (sin rail): acordeón. Se dibuja una sola de las dos navegaciones, no las dos escondidas.
	let ancho = $state(1200);
	const celular = $derived(ancho < 760);
	// Acordeón del celular: abierta el área actual; se puede abrir otra.
	let abierta = $state<string | null>(null);
	const desplegada = $derived(abierta ?? area.id);
	// Cualquier navegación (una sección, el "+", Elegir plan…) cierra el menú.
	afterNavigate(() => {
		shell.barra = false;
		abierta = null;
	});
	const activo = (label: string) =>
		label === seccion ||
		(label === 'Inicio' && seccion === 'Resumen') ||
		(label === 'Cobros' && seccion === 'Finanzas');
	const insignia = (i: Item) => (i.insignia ? resumen.insignias[i.insignia] : 0);
	const pendientes = (a: Area) => a.items.reduce((s, i) => s + insignia(i), 0);
	const NOMBRES: Record<string, string> = {
		wa: 'WhatsApp',
		ml: 'Mercado Libre',
		mp: 'Mercado Pago',
		ig: 'Instagram',
		tg: 'Telegram'
	};
	const dias = $derived(
		Array.from(
			{ length: resumen.prueba.diasTotales },
			(_, i) => i < resumen.prueba.diasTotales - resumen.prueba.diasRestantes
		)
	);
	const restan = $derived(
		`Quedan ${resumen.prueba.diasRestantes} ${resumen.prueba.diasRestantes === 1 ? 'día' : 'días'}`
	);
</script>

<svelte:window bind:innerWidth={ancho} />

<!-- eslint-disable svelte/no-navigation-without-resolve -- rutas del panel armadas con rutaDe -->
{#snippet seccionLink(i: Item)}
	{@const on = activo(i.label)}
	{@const n = insignia(i)}
	<a
		href={rutaDe(i.label)}
		class:on
		aria-current={on ? 'page' : undefined}
		onclick={() => (shell.barra = false)}
	>
		<Icon icon={on ? i.iconFill : i.icon} size={18} />
		<span class="i-l">{i.label}</span>
		{#if n}<span class="badge">{n}</span>{/if}
	</a>
{/snippet}

{#if shell.barra}
	<div class="fondo" onclick={() => (shell.barra = false)} aria-hidden="true"></div>
{/if}
<aside class="barra" class:abierta={shell.barra} aria-label="Secciones">
	<div class="arriba">
		<div class="tienda">
			{#if foto}<img src={foto} alt="" width="34" height="34" />{:else}<span
					class="t-ini"
					aria-hidden="true">{iniciales}</span
				>{/if}
			<span class="t-txt"
				><span class="t-n">{tienda}</span><span class="t-d"
					><span class="vivo"></span>{dominio}</span
				></span
			>
			{#if shell.tiendas.length > 1}
				<button
					type="button"
					class="t-cambiar"
					aria-expanded={eligiendo}
					aria-label="Cambiar de tienda"
					onclick={() => (eligiendo = !eligiendo)}><Icon icon={I.phCaretUpDown} size={17} /></button
				>
			{/if}
		</div>
		{#if eligiendo}
			<div class="t-lista" role="group" aria-label="Tus tiendas">
				{#each shell.tiendas as t (t.id)}
					<button type="button" aria-current={t.id === shell.tiendaId} onclick={() => elegir(t.id)}
						><span class="t-n">{t.nombre}</span><span class="t-r"
							>{NOMBRE_ROL[t.rol as Rol] ?? ''}</span
						>{#if t.id === shell.tiendaId}<Icon icon={I.phCheckBold} size={14} />{/if}</button
					>
				{/each}
			</div>
		{/if}
		<button
			type="button"
			class="buscar"
			onclick={() => ((shell.cmdk = true), (shell.barra = false))}
		>
			<Icon icon={I.phMagnifyingGlass} size={15} /><span>Buscar o preguntar…</span><kbd>⌘K</kbd>
		</button>
	</div>
	<div class="medio">
		<!-- Compu y tablet: las secciones del área elegida en el rail. -->
		{#if !celular}
			<div class="solo-ancho">
				<div class="area">{area.label}</div>
				<nav class="items">
					{#each area.items.filter( (i) => visible(i.label, shell.turnos, shell.permisos) ) as i (i.label)}{@render seccionLink(
							i
						)}{/each}
				</nav>
			</div>
		{:else}
			<!-- Celular: todas las áreas, en acordeón. -->
			<nav class="acordeon" aria-label="Áreas del panel">
				{#each AREAS.filter( (x) => x.items.some( (i) => visible(i.label, shell.turnos, shell.permisos) ) ) as a (a.id)}
					{@const on = desplegada === a.id}
					{@const n = pendientes(a)}
					<button
						type="button"
						class="a-fila"
						class:on
						aria-expanded={on}
						onclick={() => (abierta = on ? '' : a.id)}
					>
						<Icon icon={on ? a.iconFill : a.icon} size={19} />
						<span class="a-l">{a.label}</span>
						{#if n && !on}<span class="badge">{n}</span>{/if}
						<Icon icon={on ? I.phCaretUpBold : I.phCaretDownBold} size={13} />
					</button>
					{#if on}
						<div class="items sub">
							{#each a.items.filter( (i) => visible(i.label, shell.turnos, shell.permisos) ) as i (i.label)}{@render seccionLink(
									i
								)}{/each}
						</div>
					{/if}
				{/each}
			</nav>
		{/if}
		{#if resumen.canales.length}
			<div class="c-t">Canales conectados</div>
			<div class="canales">
				{#each resumen.canales as c (c)}<span title={NOMBRES[c] ?? c} class="canal"
						><MarcaIcono
							k={c}
							size={34}
							radius={10}
							fondo={c === 'ml' ? undefined : 'var(--sf)'}
						/></span
					>{/each}
				<a class="mas" href={rutaDe('Integraciones')} aria-label="Conectar otro canal"
					><Icon icon={I.phPlus} size={15} /></a
				>
			</div>
		{/if}
		{#if resumen.prueba.diasRestantes > 0 && shell.permisos.includes('cuenta')}
			<a class="prueba-corta" href={rutaDe('Plan y facturación')}>
				<span><b>Plan {resumen.prueba.plan} · prueba</b><span>{restan}</span></span><span
					class="pc-cta">Elegir plan</span
				>
			</a>
		{/if}
		<!-- En el celular «abajo» no se muestra: cerrar sesión va al final del menú. -->
		<form class="salir-movil" method="POST" action="/salir" data-sveltekit-reload>
			<button class="salir"><Icon icon={I.phSignOutBold} size={17} />Cerrar sesión</button>
		</form>
	</div>
	<div class="abajo">
		{#if resumen.prueba.diasRestantes > 0 && shell.permisos.includes('cuenta')}
			<div class="prueba">
				<div class="p-in">
					<div class="p-plan">Plan {resumen.prueba.plan} · prueba</div>
					<div class="p-dias">{restan}</div>
					<div class="p-barras">
						{#each dias as usado, i (i)}<span class:usado></span>{/each}
					</div>
					<a href={rutaDe('Plan y facturación')}>Elegir plan</a>
				</div>
			</div>
		{/if}
		<!-- Un POST: un link no puede sacar a nadie de su cuenta. -->
		<form method="POST" action="/salir" data-sveltekit-reload>
			<button class="salir"><Icon icon={I.phSignOutBold} size={17} />Cerrar sesión</button>
		</form>
	</div>
</aside>

<!-- eslint-enable svelte/no-navigation-without-resolve -->

<style>
	.fondo {
		position: fixed;
		inset: 0;
		z-index: 74;
		background: rgba(42, 20, 12, 0.4);
		backdrop-filter: blur(2px);
		animation: aparece 0.25s both;
	}
	@keyframes aparece {
		from {
			opacity: 0;
		}
	}
	.barra {
		width: 260px;
		flex: none;
		position: sticky;
		top: 0;
		height: 100vh;
		height: 100dvh;
		/* En escritorio, debajo de los modales (70/71); encima del contenido. */
		z-index: 50;
		background: var(--sf2);
		border-right: 1px solid var(--ln);
		display: flex;
		flex-direction: column;
		transition: transform 0.4s var(--ease);
	}
	.arriba {
		padding: 16px 14px 8px;
	}
	.tienda {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 8px;
		border-radius: 14px;
		border: 1px solid var(--ln);
		background: var(--sf);
		box-shadow: var(--sh);
	}
	.t-cambiar {
		flex: none;
		width: 36px;
		height: 36px;
		display: grid;
		place-items: center;
		border: 0;
		border-radius: 10px;
		background: transparent;
		color: var(--mu);
	}
	.t-lista {
		margin-top: 6px;
		display: flex;
		flex-direction: column;
		padding: 4px;
		border-radius: 14px;
		border: 1px solid var(--ln);
		background: var(--sf);
		box-shadow: var(--sh);
	}
	.t-lista button {
		display: flex;
		align-items: center;
		gap: 8px;
		min-height: 44px;
		padding: 8px 10px;
		border: 0;
		border-radius: 10px;
		background: transparent;
		color: var(--tx);
		font: inherit;
		text-align: left;
	}
	.t-lista button[aria-current='true'] {
		background: var(--chip);
	}
	.t-lista .t-n {
		flex: 1;
		min-width: 0;
	}
	.t-r {
		font-size: 12px;
		color: var(--mu);
	}
	.t-ini {
		flex: none;
		display: grid;
		place-items: center;
		width: 34px;
		height: 34px;
		border-radius: 10px;
		background: var(--durazno);
		color: #8a3a1c;
		font-size: 13px;
		font-weight: 700;
	}
	.tienda img {
		border-radius: 10px;
		/* Un logo no se recorta: entra entero sobre la superficie. */
		object-fit: contain;
		background: var(--sf);
		flex: none;
	}
	.t-txt {
		flex: 1;
		min-width: 0;
	}
	.t-n {
		display: block;
		font-size: 14px;
		font-weight: 600;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.t-d {
		display: flex;
		align-items: center;
		gap: 5px;
		font-size: 12px;
		color: var(--mu);
		white-space: nowrap;
		overflow: hidden;
	}
	.vivo {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		background: #2f8f63;
		flex: none;
	}
	.buscar {
		margin-top: 10px;
		width: 100%;
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 9px 12px;
		border-radius: 999px;
		border: 1px solid var(--ln);
		background: var(--sf);
		color: var(--mu);
		font-size: 13.5px;
	}
	.buscar span {
		flex: 1;
		text-align: left;
	}
	kbd {
		font-family: inherit;
		font-size: 11.5px;
		font-weight: 600;
		padding: 2px 7px;
		border-radius: 6px;
		background: var(--chip);
	}
	.medio {
		padding: 10px 12px;
		flex: 1;
		overflow: auto;
	}
	.acordeon,
	.prueba-corta {
		display: none;
	}
	.area {
		padding: 6px 12px 10px;
		font-size: 22px;
		font-weight: 600;
		letter-spacing: -0.03em;
	}
	.items {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}
	.items a {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 9px 12px;
		border-radius: 12px;
		color: var(--mu);
		font-size: 14px;
		font-weight: 500;
	}
	.items a:hover {
		color: var(--tx);
		background: var(--chip);
	}
	.items a.on {
		background: var(--sf);
		color: var(--tx);
		font-weight: 600;
		box-shadow: var(--sh);
	}
	.items a.on :global(svg) {
		color: #b0441d;
	}
	.i-l {
		flex: 1;
	}
	.badge {
		font-size: 11.5px;
		font-weight: 700;
		padding: 1px 8px;
		border-radius: 999px;
		background: rgba(210, 88, 47, 0.13);
		color: #b0441d;
	}
	.on .badge {
		background: #b0441d;
		color: #fff;
	}
	.c-t {
		margin: 22px 12px 8px;
		font-size: 12.5px;
		font-weight: 600;
		color: var(--fa);
	}
	.canales {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
		padding: 0 12px;
	}
	.canal {
		border-radius: 10px;
		box-shadow: var(--sh);
	}
	.mas {
		width: 34px;
		height: 34px;
		border-radius: 10px;
		border: 1px dashed var(--ln2);
		display: grid;
		place-items: center;
		color: var(--mu);
	}
	.abajo {
		padding: 14px;
	}
	.salir {
		margin-top: 10px;
		width: 100%;
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 10px 12px;
		border-radius: 12px;
		font-size: 14px;
		font-weight: 500;
		color: var(--mu);
	}
	.salir:hover {
		background: var(--sf);
		color: var(--tx);
	}
	.salir-movil {
		display: none;
	}
	@media (max-width: 759px) {
		.salir-movil {
			display: block;
			padding: 0 8px 20px;
		}
		.salir-movil .salir {
			min-height: 44px;
		}
	}
	.prueba {
		position: relative;
		overflow: hidden;
		border-radius: 18px;
		padding: 16px;
		background: linear-gradient(140deg, #d2582f, #b9481f);
		color: #fff;
		box-shadow: 0 14px 30px -14px rgba(210, 88, 47, 0.6);
	}
	.p-in {
		position: relative;
	}
	.p-plan {
		font-size: 12.5px;
		font-weight: 600;
		color: #fce3d6;
	}
	.p-dias {
		margin-top: 4px;
		font-size: 15px;
		font-weight: 600;
	}
	.p-barras {
		margin-top: 10px;
		display: flex;
		gap: 3px;
	}
	.p-barras span {
		flex: 1;
		height: 5px;
		border-radius: 2px;
		background: rgba(255, 255, 255, 0.28);
	}
	.p-barras span.usado {
		background: #fff;
	}
	.prueba a {
		margin-top: 12px;
		display: block;
		text-align: center;
		font-size: 13px;
		font-weight: 600;
		color: #b0441d;
		background: #fff;
		padding: 8px;
		border-radius: 999px;
	}
	/* Menos de 1100 px: se abre encima, a la derecha del rail. */
	@media (max-width: 1099px) {
		.barra {
			z-index: 75;
			position: fixed;
			left: 76px;
			width: 280px;
			transform: translateX(-115%);
		}
		.barra.abierta {
			transform: none;
			box-shadow: 30px 0 60px -20px rgba(42, 20, 12, 0.5);
		}
	}
	/* Celular: sin rail, desde el borde. Las áreas en acordeón y todo el menú se desplaza junto. */
	@media (max-width: 759px) {
		.barra {
			left: 0;
			width: min(320px, 86vw);
			overflow-y: auto;
			overscroll-behavior: contain;
		}
		.medio {
			flex: none;
			overflow: visible;
			padding-bottom: 20px;
		}
		.solo-ancho,
		.abajo {
			display: none;
		}
		.acordeon {
			display: flex;
			flex-direction: column;
			gap: 2px;
			padding: 4px 8px 0;
		}
		.a-fila {
			display: flex;
			align-items: center;
			gap: 12px;
			width: 100%;
			padding: 12px;
			border: 0;
			border-radius: 12px;
			background: transparent;
			color: var(--tx);
			font:
				600 15px 'Instrument Sans',
				sans-serif;
			text-align: left;
			cursor: pointer;
		}
		.a-fila.on {
			background: var(--sf);
			box-shadow: var(--sh);
		}
		.a-l {
			flex: 1;
		}
		.items.sub {
			padding: 2px 0 8px 18px;
		}
		.prueba-corta {
			display: flex;
			align-items: center;
			gap: 10px;
			margin: 16px 12px 0;
			padding: 12px 14px;
			border-radius: 14px;
			background: linear-gradient(140deg, #c95a32, #a8461f);
			color: #fff;
			text-decoration: none;
			font-size: 13px;
		}
		.prueba-corta > span:first-child {
			flex: 1;
			min-width: 0;
			display: flex;
			flex-direction: column;
		}
		.prueba-corta b {
			font-size: 12px;
			font-weight: 600;
			opacity: 0.85;
		}
		.pc-cta {
			flex: none;
			padding: 7px 12px;
			border-radius: 999px;
			background: #fff;
			color: #b8461f;
			font-weight: 600;
			font-size: 12.5px;
		}
	}
	@media (min-width: 1100px) {
		.fondo {
			display: none;
		}
	}
</style>
