<script lang="ts">
	import { onMount } from 'svelte';
	import MarcaIcono from '#lib/components/MarcaIcono.svelte';
	import { calma } from '#lib/anim.ts';
	import { fmt } from '#lib/landing/formato.ts';
	import Telefono from './Telefono.svelte';

	/**
	 * /l/avisos: la pantalla bloqueada a la noche y las ventas que van entrando, una por canal, como
	 * llegan los avisos de Tu Local Web. El de arriba es el último.
	 */
	const VENTAS = [
		['web', 'Venta en tu tienda online', 'Termo 1 L negro', 42900],
		['wa', 'Pedido por WhatsApp', 'Lucía · Kit matero', 58400],
		['ml', 'Venta en Mercado Libre', 'Mate de calabaza x2', 37000],
		['pos', 'Venta en el local', 'Yerba 1 kg x3', 20400],
		['ig', 'Pedido desde Instagram', 'Sofía · Bombilla de alpaca', 16900],
		['mp', 'Cobro aprobado', 'Mercado Pago · 3 cuotas', 128500]
	] as const;

	let n = $state(3);
	const avisos = $derived(
		[0, 1, 2].map((j) => {
			const i = n - 1 - j;
			const [k, t, s, v] = VENTAS[i % VENTAS.length];
			return { i, k, t, s: `${s} · ${fmt(v)}` };
		})
	);
	const total = $derived(
		Array.from({ length: n }, (_, i) => VENTAS[i % VENTAS.length][3]).reduce((a, b) => a + b, 0)
	);

	onMount(() => {
		if (calma()) return;
		const t = setInterval(() => (n = n >= 40 ? 3 : n + 1), 2600);
		return () => clearInterval(t);
	});
</script>

<Telefono fondo="linear-gradient(170deg, #6b2a14 0%, #2a140c 62%)" proporcion="9 / 14">
	<div class="bloqueo">
		<div class="hora">23:47</div>
		<div class="dia">Vendido hoy · {fmt(total)}</div>
		<div class="avisos" aria-live="off">
			{#each avisos as a, j (a.i)}
				<div class="aviso" class:nuevo={j === 0} style:opacity={1 - j * 0.16}>
					<span class="app"><img src="/logo.svg" alt="" width="22" height="22" /></span>
					<span class="txt">
						<span class="arriba"
							><b>Tu Local Web</b><span>{j === 0 ? 'ahora' : `hace ${j * 3} min`}</span></span
						>
						<span class="t"><MarcaIcono k={a.k} size={16} radius={5} />{a.t}</span>
						<span class="s">{a.s}</span>
					</span>
				</div>
			{/each}
		</div>
	</div>
</Telefono>

<style>
	.bloqueo {
		height: 100%;
		padding: 56px 12px 0;
		color: #fbf7f2;
		text-align: center;
	}
	.hora {
		font-size: 64px;
		font-weight: 600;
		letter-spacing: -0.04em;
		line-height: 1;
	}
	.dia {
		margin-top: 6px;
		font-size: 13px;
		color: rgba(251, 247, 242, 0.8);
		font-variant-numeric: tabular-nums;
	}
	.avisos {
		margin-top: 26px;
		display: grid;
		gap: 8px;
	}
	.aviso {
		display: flex;
		gap: 10px;
		padding: 10px 12px;
		border-radius: 18px;
		background: rgba(251, 247, 242, 0.92);
		color: #2a140c;
		text-align: left;
		transition: opacity 0.5s;
	}
	.nuevo {
		animation: entra 560ms cubic-bezier(0.2, 0.7, 0.1, 1);
	}
	@keyframes entra {
		from {
			opacity: 0;
			transform: translateY(-18px) scale(0.95);
		}
	}
	.app {
		flex: none;
		display: grid;
		place-items: center;
		width: 34px;
		height: 34px;
		border-radius: 9px;
		background: #fbf7f2;
		box-shadow: inset 0 0 0 1px rgba(42, 20, 12, 0.08);
	}
	.txt {
		min-width: 0;
		flex: 1;
		display: grid;
		gap: 1px;
	}
	.arriba {
		display: flex;
		justify-content: space-between;
		font-size: 11px;
		color: #7a6252;
	}
	.arriba b {
		font-weight: 600;
		color: #2a140c;
	}
	.t {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 13px;
		font-weight: 600;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.s {
		font-size: 12.5px;
		color: #5c4536;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	@media (prefers-reduced-motion: reduce) {
		.nuevo {
			animation: none;
		}
	}
</style>
