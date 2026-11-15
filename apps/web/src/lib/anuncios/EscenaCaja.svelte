<script lang="ts">
	import { onMount } from 'svelte';
	import MarcaIcono from '#lib/components/MarcaIcono.svelte';
	import { calma } from '#lib/anim.ts';
	import { fmt } from '#lib/landing/formato.ts';

	/**
	 * /l/una-caja: un producto con su stock y los tres lugares donde se vende. Cada venta, del local,
	 * de la web o de Mercado Libre, descuenta del mismo número; en cero, la tienda lo muestra agotado y
	 * Mercado Libre queda pausado. Después se repone y vuelve a empezar.
	 */
	const CANALES = [
		{ k: 'pos', nombre: 'Local', venta: 'Venta en el local' },
		{ k: 'web', nombre: 'Tienda online', venta: 'Venta en la web' },
		{ k: 'ml', nombre: 'Mercado Libre', venta: 'Venta en Mercado Libre' }
	] as const;
	const INICIAL = 3;
	// El orden de las ventas: local, Mercado Libre, web; después se repone.
	const ORDEN = [0, 2, 1];

	let paso = $state(0);
	const stock = $derived(Math.max(INICIAL - paso, 0));
	const vendio = $derived(paso > 0 && paso <= INICIAL ? ORDEN[paso - 1] : -1);

	onMount(() => {
		if (calma()) return;
		const t = setInterval(() => (paso = paso > INICIAL ? 0 : paso + 1), 2400);
		return () => clearInterval(t);
	});
</script>

<div class="escena">
	<div class="producto">
		<img src="/demo/lumbre/tote-de-cuero.webp" alt="" width="120" height="150" />
		<div>
			<div class="nombre">Tote de cuero</div>
			<div class="precio">{fmt(89900)}</div>
			<div class="stock" class:cero={stock === 0}>
				{#key stock}<b>{stock}</b>{/key}
				{stock === 0 ? 'Sin stock' : 'en stock'}
			</div>
		</div>
	</div>
	<div class="canales">
		{#each CANALES as c, i (c.k)}
			<div class="canal" class:vendio={vendio === i}>
				<MarcaIcono k={c.k} size={34} radius={10} />
				<span class="txt">
					<b>{c.nombre}</b>
					<span>
						{#if vendio === i}
							{c.venta} · −1
						{:else if stock === 0}
							{c.k === 'ml' ? 'Publicación pausada' : 'Agotado'}
						{:else}
							{`${stock}\u00a0${stock === 1 ? 'disponible' : 'disponibles'}`}
						{/if}
					</span>
				</span>
			</div>
		{/each}
	</div>
</div>

<style>
	.escena {
		width: min(380px, 100%);
		display: grid;
		gap: 14px;
	}
	.producto {
		display: flex;
		align-items: center;
		gap: 18px;
		padding: 14px;
		border-radius: 26px;
		background: #fff;
		box-shadow: 0 40px 70px -40px rgba(42, 20, 12, 0.6);
	}
	img {
		width: 96px;
		height: 120px;
		object-fit: contain;
		border-radius: 16px;
		background: #fff;
	}
	.nombre {
		font-weight: 600;
		font-size: 17px;
	}
	.precio {
		color: #7a6252;
		font-size: 15px;
	}
	.stock {
		margin-top: 10px;
		display: flex;
		align-items: baseline;
		gap: 8px;
		color: #5c4536;
		font-size: 15px;
	}
	.stock b {
		display: inline-block;
		font-size: 44px;
		line-height: 1;
		font-weight: 600;
		letter-spacing: -0.04em;
		color: #2a140c;
		font-variant-numeric: tabular-nums;
		animation: baja 500ms cubic-bezier(0.2, 0.7, 0.1, 1);
	}
	.cero b {
		color: #d2582f;
	}
	@keyframes baja {
		from {
			opacity: 0;
			transform: translateY(-16px);
		}
	}
	.canales {
		display: grid;
		gap: 8px;
	}
	.canal {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 10px 14px;
		border-radius: 18px;
		background: rgba(255, 255, 255, 0.7);
		transition:
			background 0.4s,
			transform 0.4s;
	}
	.canal.vendio {
		background: #fcede5;
		transform: translateX(6px);
	}
	.txt {
		display: grid;
		font-size: 13.5px;
		color: #5c4536;
	}
	.txt b {
		color: #2a140c;
		font-size: 15px;
		font-weight: 600;
	}
	@media (prefers-reduced-motion: reduce) {
		.stock b {
			animation: none;
		}
		.canal {
			transition: none;
		}
	}
</style>
