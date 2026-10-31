<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { page } from '$app/state';
	import { varsTema, type TemaOpiniones } from './opiniones.ts';

	/**
	 * La seña de un pedido reservado o encargado (§52) en su confirmación, con los colores de la
	 * plantilla: si ya se pagó y cuánto falta al retirarlo. La lee de la página del pedido.
	 */
	let { tema }: { tema: TemaOpiniones } = $props();

	type Sena = {
		tipo: 'reserva' | 'encargo';
		monto: number;
		falta: number;
		pagada: boolean;
		saldado: boolean;
	};
	const sena = $derived((page.data.pedido as { sena?: Sena | null } | undefined)?.sena ?? null);
	const demo = $derived(!!(page.data as { senaDemo?: boolean }).senaDemo);
	const fmt = (n: number) => '$ ' + Math.round(n).toLocaleString('es-AR');
</script>

{#if sena}
	<div class="sp" style={varsTema(tema)} role="status">
		<Icon icon={sena.pagada ? I.phCheckBold : I.phHandCoinsBold} size={18} />
		<p>
			{#if sena.saldado}
				Pedido pagado completo. ¡Gracias!
			{:else if sena.pagada}
				<b>Seña pagada · falta {fmt(sena.falta)}</b>, que pagás al&nbsp;retirarlo.
			{:else}
				<b>Falta pagar la seña de {fmt(sena.monto)}</b> con Mercado Pago para que quede {sena.tipo ===
				'reserva'
					? 'reservado'
					: 'encargado'}.
			{/if}
			{#if demo}<span class="demo"
					>En la demo no se cobra: en tu tienda, la seña se paga con Mercado Pago.</span
				>{/if}
		</p>
	</div>
{/if}

<style>
	.sp {
		box-sizing: border-box;
		margin-top: 14px;
		display: flex;
		gap: 10px;
		align-items: flex-start;
		padding: 12px 14px;
		border: 1px solid var(--op-ln);
		border-radius: var(--op-r);
		background: var(--op-sf);
		color: var(--op-tx);
		font: 400 14.5px/1.5 var(--op-f);
		text-align: left;
	}
	.sp :global(svg) {
		flex: none;
		margin-top: 2px;
	}
	p {
		margin: 0;
		text-wrap: pretty;
	}
	.demo {
		display: block;
		margin-top: 4px;
		font-size: 12.5px;
		color: var(--op-mu);
	}
</style>
