<script lang="ts">
	import Seo from '#lib/components/Seo.svelte';
	import Plantilla from '#lib/tienda/Plantilla.svelte';

	let { data } = $props();
</script>

<Seo
	title="Favoritos · {data.tienda.nombre}"
	description="Los productos que guardaste en {data.tienda.nombre}."
	path="/t/{data.tienda.slug}/favoritos"
	noindex
/>
<Plantilla tienda={data.tienda} pagina="favoritos" />
