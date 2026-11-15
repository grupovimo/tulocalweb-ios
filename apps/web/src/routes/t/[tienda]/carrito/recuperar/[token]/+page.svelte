<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import Seo from '#lib/components/Seo.svelte';
	import { imagenTienda, textosCompartir } from '#lib/tienda/compartir.ts';
	import { carritoDe } from '#lib/tienda/carrito.svelte.ts';
	import Plantilla from '#lib/tienda/Plantilla.svelte';

	/**
	 * El link de "¿Te olvidaste algo?": vuelve a armar el carrito en el navegador y lleva al carrito
	 * de la plantilla. Si el link ya no sirve (compró o venció), el carrito queda como estaba.
	 */
	let { data } = $props();
	onMount(() => {
		if (data.lineas?.length) carritoDe(data.tienda.slug).restaurar(data.lineas);
		// Con ?cupon= (el recordatorio con cupón, §62) llega al carrito con el cupón aplicado.
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- ruta de la tienda con su slug
		goto(`/t/${data.tienda.slug}/carrito${location.search}`, { replaceState: true });
	});
</script>

<Seo
	title="Tu carrito · {data.tienda.nombre}"
	description={textosCompartir(data.tienda).descripcion}
	path="/t/{data.tienda.slug}/carrito"
	noindex={true}
	imagen={imagenTienda(data.tienda)}
/>
<Plantilla tienda={data.tienda} pagina="carrito" />
