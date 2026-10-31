<script lang="ts">
	import PaginaNewsletter from '#lib/tienda/PaginaNewsletter.svelte';
	import type { PageProps } from './$types';

	/** Confirmar la suscripción al newsletter con el link del mail (§50). */
	let { data, form }: PageProps = $props();
</script>

<PaginaNewsletter tienda={data.tienda} token={data.token} modo="confirmar" {form} />
