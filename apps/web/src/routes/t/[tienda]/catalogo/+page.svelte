<script lang="ts">
	import { page } from '$app/state';
	import Seo from '#lib/components/Seo.svelte';
	import { imagenTienda } from '#lib/tienda/compartir.ts';
	import Plantilla from '#lib/tienda/Plantilla.svelte';

	let { data } = $props();
	const cat = $derived(page.url.searchParams.get('cat'));
	const categoria = $derived(data.tienda.categorias.some((c) => c.nombre === cat) ? cat : null);
</script>

<Seo
	title="{categoria ?? 'Productos'} · {data.tienda.nombre}"
	description="Todos los productos de {data.tienda.nombre}"
	path="/t/{data.tienda.slug}/catalogo"
	imagen={imagenTienda(data.tienda)}
/>
<Plantilla tienda={data.tienda} pagina="catalogo" />
