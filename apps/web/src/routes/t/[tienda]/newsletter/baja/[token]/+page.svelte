<script lang="ts">
	import Newsletter from '#lib/tienda/sistema/paginas/Newsletter.svelte';
	import type { PageProps } from './$types';

	/** Darse de baja del newsletter (link y List-Unsubscribe de cada mail) (§50). */
	let { data, form }: PageProps = $props();
</script>

<Newsletter tienda={data.tienda} token={data.token} modo="baja" {form} />
