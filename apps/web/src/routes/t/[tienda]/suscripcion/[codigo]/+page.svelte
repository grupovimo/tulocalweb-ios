<script lang="ts">
	import Seo from '#lib/components/Seo.svelte';
	import { imagenTienda, textosCompartir } from '#lib/tienda/compartir.ts';
	import Plantilla from '#lib/tienda/Plantilla.svelte';

	let { data, form } = $props();
</script>

<Seo
	title="Tu suscripción · {data.tienda.nombre}"
	description={textosCompartir(data.tienda).descripcion}
	path="/t/{data.tienda.slug}"
	noindex
	imagen={imagenTienda(data.tienda)}
/>
<Plantilla
	tienda={data.tienda}
	pagina="mi-suscripcion"
	suscripcion={data.suscripcion}
	aviso={data.aviso}
	demo={data.demo}
	error={form?.error ?? null}
/>
