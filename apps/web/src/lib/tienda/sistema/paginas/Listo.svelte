<script lang="ts" module>
	export type PedidoListo = {
		numero: number;
		total: number;
		pago: string;
		nombre: string;
		alias: string | null;
		esCbu?: boolean;
		/** El código del pedido: con él, la reseña queda con compra. */
		codigo?: string;
		/** La reseña se pide recién cuando lo recibió (Entregado o Retirado). */
		estado?: string;
		/** Con Mercado Pago y todavía sin pagar (Mercado Pago no armó el pago o lo rechazó): se puede pagar acá. */
		porPagar?: boolean;
		/** Volvió de Mercado Pago con el pago rechazado. */
		rechazado?: boolean;
	};
</script>

<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import FormResena from '../../FormResena.svelte';
	import SenaPedido from '../../SenaPedido.svelte';
	import type { TiendaPublica } from '../../tipos.ts';
	import { waHref } from '../datos.ts';
	import { precio } from '../formato.ts';
	import Tienda from '../marco/Tienda.svelte';
	import { disenoDeTienda } from '../migrar.ts';
	import { temaDe } from '../paginas.ts';
	import { estiloFinal } from '../presets.ts';
	import Boton from '../primitivas/Boton.svelte';

	/**
	 * El pedido confirmado: gracias, el número y el total; la seña si la tiene (SenaPedido); los datos
	 * para transferir con el comprobante por WhatsApp; y, cuando ya lo recibió, la reseña con compra.
	 */
	let { tienda, pedido }: { tienda: TiendaPublica; pedido: PedidoListo } = $props();
	const diseno = $derived(disenoDeTienda(tienda));
	const dir = $derived(estiloFinal(tienda.plantilla, diseno.paleta, diseno.estilo).direccion);
	const tema = $derived(temaDe(dir));
	const base = $derived(tienda.base ?? `/t/${tienda.slug}`);
	const wa = $derived(
		waHref(tienda, `¡Hola! Te mando el comprobante del pedido #${pedido.numero}.`)
	);
	const recibido = $derived(pedido.estado === 'Entregado' || pedido.estado === 'Retirado');
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
<Tienda {tienda} pagina="listo">
	<section class="tv-ancho tv-pag tv-confirma">
		<span class="tv-confirma-ico"><Icon icon={I.phCheckBold} size={28} /></span>
		<h1 class="tv-h1 tv-pag-titulo">¡Gracias, {pedido.nombre}!</h1>
		<p class="tv-lead">
			Recibimos tu pedido <b>#{pedido.numero}</b> por <b>{precio(pedido.total)}</b>. Te escribimos
			por mail con cada&nbsp;novedad.
		</p>
		<SenaPedido {tema} />
		{#if pedido.porPagar}
			<form class="tv-confirma-caja tv-tarjeta" method="POST" action="?/pagar">
				<p>
					{#if pedido.rechazado}<b>El pago no se completó.</b> Probá de nuevo, con otra tarjeta o
						medio.{:else}<b>Falta pagar {precio(pedido.total)}</b> con Mercado Pago.{/if} El pedido te
					lo guardamos 24&nbsp;horas.
				</p>
				<button class="tv-btn tv-btn--primario tv-btn--bloque" type="submit">Pagar con Mercado&nbsp;Pago</button>
			</form>
		{:else if pedido.pago === 'transferencia'}
			<div class="tv-confirma-caja tv-tarjeta">
				<p>
					<b>Transferí {precio(pedido.total)}</b>
					{#if pedido.alias}al {pedido.esCbu ? 'CBU/CVU' : 'alias'}
						<b class="tv-confirma-alias">{pedido.alias}</b>{:else}con los datos que te mandamos por
						mail{/if} y mandanos el comprobante{#if wa}&nbsp;<a
							href={wa}
							target="_blank"
							rel="noopener">por WhatsApp</a
						>{:else}&nbsp;respondiendo el mail del pedido{/if}. Lo preparamos apenas
					se&nbsp;acredite.
				</p>
			</div>
		{/if}
		<div class="tv-botones tv-confirma-botones">
			<Boton href={base || '/'}>Volver a la tienda</Boton>
		</div>
		{#if recibido}
			<div class="tv-confirma-resena">
				<FormResena slug={tienda.slug} {tema} codigo={pedido.codigo} nombre={pedido.nombre} />
			</div>
		{/if}
	</section>
</Tienda>
