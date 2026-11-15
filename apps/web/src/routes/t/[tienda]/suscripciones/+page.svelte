<script lang="ts">
	import Seo from '#lib/components/Seo.svelte';
	import { imagenTienda } from '#lib/tienda/compartir.ts';
	import Plantilla from '#lib/tienda/Plantilla.svelte';

	let { data } = $props();
</script>

<Seo
	title="Suscripciones · {data.tienda.nombre}"
	description="Los planes de suscripción de {data.tienda
		.nombre}, con débito automático de Mercado Pago."
	path="/t/{data.tienda.slug}/suscripciones"
	imagen={imagenTienda(data.tienda)}
/>
<Plantilla tienda={data.tienda} pagina="suscripciones" />
