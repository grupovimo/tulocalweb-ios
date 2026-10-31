<script lang="ts">
	import PaginaNewsletter from '#lib/tienda/PaginaNewsletter.svelte';
	import type { PageProps } from './$types';

	/** Darse de baja del newsletter (link y List-Unsubscribe de cada mail) (§50). */
	let { data, form }: PageProps = $props();
</script>

<PaginaNewsletter tienda={data.tienda} token={data.token} modo="baja" {form} />
