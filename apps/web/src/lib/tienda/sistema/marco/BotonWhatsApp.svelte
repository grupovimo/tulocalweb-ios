<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { usarTienda } from '../contexto.ts';
	import { waHref } from '../datos.ts';

	/**
	 * El botón flotante de WhatsApp: en la compu, una pestaña pegada al borde derecho que entra en el
	 * margen lateral (regla 8: lo flotante no tapa). En el celular no se dibuja: no hay margen que
	 * alcance, y WhatsApp ya está en el menú, el pie y la barra de la ficha. Tampoco sin número, con
	 * estilo.whatsappFlotante = false, en el checkout ni cuando la página lo saca.
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

<!-- eslint-disable svelte/no-navigation-without-resolve -- link a WhatsApp -->
{#if ver}
	<a class="tv-wa" {href} target="_blank" rel="noopener" aria-label="Escribinos por WhatsApp"
		><Icon icon={I.phWhatsappLogoFill} size={22} /></a
	>
{/if}
