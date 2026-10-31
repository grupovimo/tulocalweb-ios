<script lang="ts">
	import Seo from '#lib/components/Seo.svelte';
	import { imagenTienda } from '#lib/tienda/compartir.ts';
	import Plantilla from '#lib/tienda/Plantilla.svelte';
	import { CADA } from '#lib/tienda/suscripciones.ts';

	let { data, form } = $props();
</script>

<Seo
	title="{data.plan.nombre} · {data.tienda.nombre}"
	description="Suscribite a {data.plan.nombre} de {data.tienda.nombre}: {CADA[
		data.plan.frecuencia
	]}, con débito automático de Mercado Pago."
	path="/t/{data.tienda.slug}/suscripciones/{data.plan.id}"
	noindex={false}
	imagen={imagenTienda(data.tienda)}
/>
<Plantilla
	tienda={data.tienda}
	pagina="suscripcion"
	plan={data.plan}
	demo={data.demo}
	error={form?.error ?? null}
/>
