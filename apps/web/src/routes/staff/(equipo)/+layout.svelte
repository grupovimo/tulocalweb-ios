<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { page } from '$app/state';
	import Icon from '#lib/components/Icon.svelte';
	import Segmentado from '#lib/components/Segmentado.svelte';
	import * as I from '#lib/icons.ts';
	import { nombreRol, ROLES, seccionesDe, type Rol } from '#lib/staff/permisos.ts';

	/**
	 * Backoffice del equipo de Tu Local Web (/staff). El handoff no lo diseña: se arma con las piezas del
	 * panel (navegación, tarjetas, KPIs, listas). La barra oscura de arriba deja claro que es interno.
	 */
	let { data, children } = $props();
	const secciones = $derived(seccionesDe(data.sesion.permisos));
	const actual = (ruta: string) =>
		ruta === '/staff' ? page.url.pathname === '/staff' : page.url.pathname.startsWith(ruta);

	async function cambiarRol(rol: Rol) {
		await fetch('/staff/rol', {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ rol })
		});
		await invalidateAll();
	}
</script>

<div class="staff">
	<header class="equipo">
		<span class="marca"><img src="/logo.svg" alt="" width="20" height="20" />Equipo</span>
		<span class="quien"
			>{data.sesion.miembro.nombre}<span class="rol">{nombreRol(data.sesion.rol)}</span></span
		>
		<form method="POST" action="/staff/salir">
			<button type="submit" class="salir"
				><Icon icon={I.phSignOutBold} size={14} /><span>Salir</span></button
			>
		</form>
	</header>

	{#if data.demo}
		<div class="aviso-demo">
			<span
				>Es un backoffice de ejemplo: los datos son inventados y las acciones no se aplican.</span
			>
			<Segmentado
				opciones={ROLES.map((r) => [r.id, r.nombre] as const)}
				value={data.sesion.rol}
				label="Ver como"
				onchange={cambiarRol}
			/>
		</div>
	{/if}

	<div class="cuerpo">
		<nav class="nav" aria-label="Secciones del equipo">
			<!-- eslint-disable svelte/no-navigation-without-resolve -- rutas del equipo de lib/staff/permisos.ts -->
			{#each secciones as s (s.id)}
				<a
					href={s.ruta}
					class:on={actual(s.ruta)}
					aria-current={actual(s.ruta) ? 'page' : undefined}
					><Icon icon={s.icon} size={17} /><span>{s.t}</span></a
				>
			{/each}
			<!-- eslint-enable svelte/no-navigation-without-resolve -->
		</nav>
		<main class="contenido">{@render children()}</main>
	</div>
</div>

<style>
	.staff {
		min-height: 100vh;
		background: var(--bg);
		color: var(--tx);
		font-family: 'Instrument Sans', system-ui, sans-serif;
	}
	.equipo {
		position: sticky;
		top: 0;
		z-index: 5;
		display: flex;
		align-items: center;
		gap: 14px;
		padding: 10px 20px;
		background: #2a140c;
		color: #fbf7f2;
	}
	.marca {
		display: flex;
		align-items: center;
		gap: 8px;
		font-weight: 700;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		font-size: 12.5px;
	}
	.quien {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: flex-end;
		gap: 8px;
		font-size: 13.5px;
		font-weight: 600;
		min-width: 0;
	}
	.rol {
		padding: 3px 9px;
		border-radius: 999px;
		background: rgba(251, 247, 242, 0.14);
		font-size: 11.5px;
		font-weight: 600;
		white-space: nowrap;
	}
	.salir {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 7px 12px;
		border-radius: 999px;
		border: 1px solid rgba(251, 247, 242, 0.25);
		background: transparent;
		color: inherit;
		font: inherit;
		font-size: 13px;
		font-weight: 600;
	}
	.aviso-demo {
		display: flex;
		align-items: center;
		justify-content: space-between;
		flex-wrap: wrap;
		gap: 10px;
		padding: 10px 20px;
		background: #fcede5;
		color: #a8401c;
		font-size: 13.5px;
	}
	.cuerpo {
		display: grid;
		grid-template-columns: 220px minmax(0, 1fr);
		gap: 24px;
		max-width: 1320px;
		margin: 0 auto;
		padding: 24px 20px 60px;
	}
	.nav {
		position: sticky;
		top: 70px;
		align-self: start;
		display: flex;
		flex-direction: column;
		gap: 4px;
	}
	.nav a {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 10px 12px;
		border-radius: 12px;
		color: var(--mu);
		font-size: 14.5px;
		font-weight: 600;
		text-decoration: none;
	}
	.nav a.on {
		background: var(--sf);
		color: var(--tx);
		box-shadow: var(--sh);
	}
	.nav a.on :global(svg) {
		color: #b0441d;
	}
	.contenido {
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 18px;
	}
	/* Piezas que comparten las pantallas del equipo (las mismas del panel). */
	.staff :global(.cab) {
		display: flex;
		align-items: end;
		gap: 12px;
		flex-wrap: wrap;
	}
	.staff :global(.cab h1) {
		flex: 1;
		min-width: 200px;
		margin: 0;
		font-size: 30px;
		letter-spacing: -0.035em;
		font-weight: 600;
	}
	.staff :global(.sub) {
		margin: 2px 0 0;
		font-size: 15px;
		color: var(--mu);
	}
	.staff :global(.tarjeta) {
		padding: 16px;
		border-radius: 20px;
		background: var(--sf);
		box-shadow: var(--sh);
	}
	.staff :global(.t-t) {
		font-size: 15px;
		font-weight: 600;
		margin-bottom: 10px;
	}
	.staff :global(.kpis) {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 150px), 1fr));
		gap: 12px;
	}
	.staff :global(.kpi) {
		padding: 14px 16px;
		border-radius: 18px;
		background: var(--sf);
		box-shadow: var(--sh);
	}
	.staff :global(.kpi .k-t) {
		font-size: 12.5px;
		font-weight: 600;
		color: var(--mu);
	}
	.staff :global(.kpi .k-n) {
		margin-top: 4px;
		font-size: 24px;
		font-weight: 600;
		letter-spacing: -0.02em;
		font-variant-numeric: tabular-nums;
	}
	.staff :global(.fila) {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 11px 0;
		border-bottom: 1px solid var(--ln);
		font-size: 14px;
	}
	.staff :global(.fila:last-child) {
		border-bottom: 0;
	}
	.staff :global(.f-txt) {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 2px;
	}
	.staff :global(.f-s) {
		font-size: 12.5px;
		color: var(--mu);
	}
	.staff :global(.chip) {
		display: inline-flex;
		padding: 4px 10px;
		border-radius: 999px;
		font-size: 12px;
		font-weight: 600;
		white-space: nowrap;
	}
	.staff :global(.vacio) {
		padding: 18px 0;
		text-align: center;
		color: var(--mu);
		font-size: 14px;
	}
	.staff :global(.boton) {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		padding: 10px 18px;
		border-radius: 999px;
		border: 0;
		background: linear-gradient(180deg, #c24e25, #b0441d);
		color: #fff;
		font:
			600 14px 'Instrument Sans',
			sans-serif;
		text-decoration: none;
		white-space: nowrap;
	}
	.staff :global(.borde) {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		padding: 9px 16px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: var(--sf);
		color: var(--tx);
		font:
			600 13.5px 'Instrument Sans',
			sans-serif;
		text-decoration: none;
		white-space: nowrap;
	}
	.staff :global(.campo) {
		display: flex;
		flex-direction: column;
		gap: 4px;
		font-size: 12.5px;
		font-weight: 600;
		color: var(--mu);
	}
	.staff :global(.campo input),
	.staff :global(.campo textarea),
	.staff :global(.campo select) {
		padding: 11px 12px;
		border-radius: 12px;
		border: 1px solid var(--ln2);
		background: var(--bg);
		font:
			500 15px 'Instrument Sans',
			sans-serif;
		color: var(--tx);
		outline: none;
	}
	.staff :global(.campo textarea) {
		min-height: 90px;
		resize: vertical;
		font-size: 14px;
		line-height: 1.5;
	}
	@media (max-width: 860px) {
		.cuerpo {
			grid-template-columns: minmax(0, 1fr);
			padding: 12px 16px 60px;
			gap: 14px;
		}
		.nav {
			position: static;
			flex-direction: row;
			overflow-x: auto;
			margin: 0 -16px;
			padding: 0 16px 4px;
		}
		.nav a {
			flex: none;
			padding: 8px 12px;
		}
		.equipo {
			padding: 10px 16px;
		}
		.salir span {
			display: none;
		}
		.aviso-demo {
			padding: 10px 16px;
		}
		.staff :global(.cab h1) {
			font-size: 26px;
		}
		.staff :global(.kpis) {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
	}
</style>
