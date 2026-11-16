<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { cuotasPara, precioTransferencia, type Ofrece } from '../cobros.ts';
	import { precio } from '../formato.ts';

	/**
	 * «Ver medios de pago» de la ficha: un panel al costado con lo que la tienda acepta de verdad,
	 * con el precio de este producto en cada caso: Mercado Pago (tarjetas, en un pago o en las cuotas
	 * sin interés que ofrece desde su monto) y transferencia (con su descuento, si tiene). Lo que no
	 * configuró no sale. Esc, tocar afuera o la cruz lo cierran.
	 */
	let {
		abierto,
		total,
		ofrece,
		alCerrar
	}: { abierto: boolean; total: number; ofrece: Ofrece; alCerrar: () => void } = $props();

	const enCuotas = $derived(cuotasPara(ofrece, total));
	const transf = $derived(precioTransferencia(ofrece, total));
	let dialogo = $state<HTMLDialogElement>();
	$effect(() => {
		if (!dialogo) return;
		if (abierto && !dialogo.open) dialogo.showModal();
		else if (!abierto && dialogo.open) dialogo.close();
	});
</script>

<dialog
	bind:this={dialogo}
	class="tv-panel tv-medios"
	aria-labelledby="tv-medios-titulo"
	onclose={() => abierto && alCerrar()}
	onclick={(e) => e.target === dialogo && alCerrar()}
>
	<div class="tv-panel-caja">
		<div class="tv-panel-cab">
			<h2 id="tv-medios-titulo" class="tv-h3">Medios de pago</h2>
			<button type="button" class="tv-icono" aria-label="Cerrar" onclick={alCerrar}
				><Icon icon={I.phXBold} size={18} /></button
			>
		</div>
		<div class="tv-panel-cuerpo">
			{#if ofrece.mercadoPago}
				<section class="tv-medio" aria-labelledby="tv-medio-mp">
					<h3 id="tv-medio-mp" class="tv-medio-tit">
						<Icon icon={I.phCreditCard} size={20} />Tarjetas y Mercado Pago
					</h3>
					<ul class="tv-medio-filas">
						<li><span>1 pago</span><b>{precio(total)}</b></li>
						{#if enCuotas}
							<li class="tv-ok">
								<span>{enCuotas.n} cuotas sin interés</span><b
									>{enCuotas.n}&nbsp;x {precio(enCuotas.valor)}</b
								>
							</li>
						{:else if ofrece.cuotas > 1}
							<li class="tv-suave">
								<span
									>{ofrece.cuotas} cuotas sin interés en compras desde {precio(
										ofrece.cuotasDesde
									)}</span
								>
							</li>
						{/if}
					</ul>
					<p class="tv-chico tv-suave">
						Tarjeta de crédito o débito, o dinero en tu cuenta de Mercado&nbsp;Pago.
					</p>
				</section>
			{/if}
			{#if ofrece.aceptaTransferencia}
				<section class="tv-medio" aria-labelledby="tv-medio-tr">
					<h3 id="tv-medio-tr" class="tv-medio-tit">
						<Icon icon={I.phBank} size={20} />Transferencia
					</h3>
					<ul class="tv-medio-filas">
						{#if transf !== null}
							<li class="tv-ok">
								<span>{ofrece.transferencia}&nbsp;% de descuento</span><b>{precio(transf)}</b>
							</li>
						{:else}
							<li><span>Transferencia bancaria</span><b>{precio(total)}</b></li>
						{/if}
					</ul>
					<p class="tv-chico tv-suave">Te pasamos los datos al terminar la&nbsp;compra.</p>
				</section>
			{/if}
		</div>
	</div>
</dialog>
