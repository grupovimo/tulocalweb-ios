<script lang="ts">
	import BajaCarrito from '#lib/tienda/sistema/paginas/BajaCarrito.svelte';
	import type { PageProps } from './$types';

	/** Dejar de recibir los recordatorios del carrito (link y List-Unsubscribe de cada mail) (§62). */
	let { data, form }: PageProps = $props();
</script>

<BajaCarrito tienda={data.tienda} token={data.token} {form} />
