<script lang="ts">
	import Seo from '#lib/components/Seo.svelte';
	import { imagenTienda } from '#lib/tienda/compartir.ts';
	import Arrepentimiento from '#lib/tienda/sistema/paginas/Arrepentimiento.svelte';
	import type { PageProps } from './$types';

	/** Botón de arrepentimiento de la tienda, con su marco (sistema/paginas/Arrepentimiento). */
	let { data, form }: PageProps = $props();
</script>

<Seo
	title="Botón de arrepentimiento · {data.tienda.nombre}"
	description="Revocá tu compra en {data.tienda.nombre} hasta 10 días después de recibirla."
	path="/t/{data.tienda.slug}/arrepentimiento"
	noindex
	imagen={imagenTienda(data.tienda)}
/>
<Arrepentimiento tienda={data.tienda} codigo={data.codigo} {form} />
