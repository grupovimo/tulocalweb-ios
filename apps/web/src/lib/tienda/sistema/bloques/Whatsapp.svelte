<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { textoDe, type Bloque } from '../catalogo.ts';
	import { usarTienda } from '../contexto.ts';
	import { waHref } from '../datos.ts';
	import { sinViuda } from '../formato.ts';
	import Boton from '../primitivas/Boton.svelte';

	/**
	 * Escribinos: una invitación a consultar por WhatsApp. `banda`: franja con el botón al costado;
	 * `tarjeta`: tarjeta centrada. Sin número cargado no se dibuja.
	 */
	let { bloque }: { bloque: Bloque } = $props();
	const ctx = usarTienda();
	const href = $derived(waHref(ctx.tienda, '¡Hola! Quería hacer una consulta.'));
	const texto = $derived(textoDe(bloque, 'texto') || 'Consultanos lo que necesites por WhatsApp.');
</script>

{#if href}
	<section class="tv-escribinos tv-ancho" data-variante={bloque.variante}>
		<div class="tv-escribinos-caja tv-z" data-fondo="superficie">
			<span class="tv-escribinos-icono"><Icon icon={I.phWhatsappLogoFill} size={28} /></span>
			<div class="tv-escribinos-textos">
				<h2 class="tv-h3">{sinViuda(textoDe(bloque, 'titulo'))}</h2>
				<p>{texto}</p>
			</div>
			<!-- El verde de WhatsApp es de Mostrador y Pedido (B-moda, C-vacia); Vidriera no lo usa. -->
			<Boton {href} tipo={ctx.dir === 'vidriera' ? 'secundario' : 'wa'} externo>Escribinos</Boton>
		</div>
	</section>
{/if}
