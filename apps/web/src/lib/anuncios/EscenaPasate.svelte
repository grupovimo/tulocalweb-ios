<script lang="ts">
	import { onMount } from 'svelte';
	import Icon from '#lib/components/Icon.svelte';
	import { calma } from '#lib/anim.ts';
	import { phCheckCircleFill } from '#lib/icons.ts';
	import Telefono from './Telefono.svelte';

	/**
	 * /l/pasate: una tienda real hecha con Tu Local Web (Tienda Lumbre, la demo) y encima la
	 * importación desde Tiendanube, que avanza hasta quedar lista y vuelve a empezar.
	 */
	const FILAS = [
		['Productos', 248],
		['Fotos', 1120],
		['Clientes', 1904],
		['Pedidos', 3612]
	] as const;
	const PASOS = 40;

	let p = $state(PASOS);
	const lista = $derived(p >= PASOS);

	onMount(() => {
		if (calma()) return;
		p = 0;
		// 40 pasos de 80 ms (3,2 s) para importar y unos 3 s quieta, ya lista.
		const t = setInterval(() => (p = p >= PASOS + 38 ? 0 : p + 1), 80);
		return () => clearInterval(t);
	});
	const n = (total: number) => Math.round((total * Math.min(p, PASOS)) / PASOS);
</script>

<div class="escena">
	<Telefono fondo="#f3efe9">
		<img
			src="/anuncios/lumbre.webp"
			alt="La tienda online de Tienda Lumbre"
			width="560"
			height="1138"
		/>
	</Telefono>
	<div class="import">
		<div class="cab">
			{#if lista}
				<Icon icon={phCheckCircleFill} size={18} style="color:#237550" />Tu tienda ya está acá
			{:else}
				<span class="punto"></span>Trayendo desde Tiendanube
			{/if}
		</div>
		<div class="barra"><span style:width="{(Math.min(p, PASOS) / PASOS) * 100}%"></span></div>
		{#each FILAS as [nombre, total] (nombre)}
			<div class="fila">
				<span>{nombre}</span><b>{n(total).toLocaleString('es-AR')}</b>
			</div>
		{/each}
	</div>
</div>

<style>
	.escena {
		position: relative;
		padding-bottom: 40px;
	}
	img {
		display: block;
		width: 100%;
		/* Debajo de la isla, como una barra de estado. */
		margin-top: 44px;
		height: calc(100% - 44px);
		object-fit: cover;
		object-position: top;
	}
	.import {
		position: absolute;
		left: -34px;
		bottom: 0;
		width: 230px;
		padding: 14px 16px;
		border-radius: 20px;
		background: #2a140c;
		color: #fbf7f2;
		box-shadow: 0 30px 50px -20px rgba(42, 20, 12, 0.7);
	}
	.cab {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 13.5px;
		font-weight: 600;
	}
	.punto {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: #e8875f;
		animation: late 1s infinite;
	}
	@keyframes late {
		50% {
			opacity: 0.3;
		}
	}
	.barra {
		margin: 10px 0 8px;
		height: 5px;
		border-radius: 5px;
		background: rgba(251, 247, 242, 0.15);
		overflow: hidden;
	}
	.barra span {
		display: block;
		height: 100%;
		background: #e8875f;
	}
	.fila {
		display: flex;
		justify-content: space-between;
		font-size: 13px;
		color: #d9c8ba;
		padding: 2px 0;
	}
	.fila b {
		color: #fbf7f2;
		font-weight: 600;
		font-variant-numeric: tabular-nums;
	}
	@media (max-width: 520px) {
		.import {
			left: -8px;
		}
	}
	@media (prefers-reduced-motion: reduce) {
		.punto {
			animation: none;
		}
	}
</style>
