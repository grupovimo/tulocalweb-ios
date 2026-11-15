<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { usarTienda } from '../contexto.ts';
	import { waHref } from '../datos.ts';

	/**
	 * El botón flotante de WhatsApp (abajo a la derecha). No se dibuja sin número, con
	 * estilo.whatsappFlotante = false, en el checkout ni cuando la barra del pedido ocupa ese lugar.
	 * Regla 8 (lo flotante no tapa): deja su espacio al final de la página (.tv-reserva-wa), así el
	 * pie nunca queda abajo del botón.
	 *
	 * Uso: lo pone marco/Tienda.svelte. Props: `oculto` (la página lo saca, p. ej. con la barra fija).
	 */
	let { oculto = false }: { oculto?: boolean } = $props();
	const ctx = usarTienda();
	const href = $derived(waHref(ctx.tienda, '¡Hola! Quería hacer una consulta.'));
	const ver = $derived(
		!!href && ctx.estilo.whatsappFlotante && !oculto && ctx.pagina !== 'checkout'
	);
</script>

{#if ver}
	<div class="tv-reserva-wa" aria-hidden="true"></div>
	<a class="tv-wa" {href} target="_blank" rel="noopener" aria-label="Escribinos por WhatsApp"
		><Icon icon={I.phWhatsappLogoFill} size={28} /></a
	>
{/if}
