<script lang="ts">
	import Newsletter from '#lib/tienda/sistema/paginas/Newsletter.svelte';
	import type { PageProps } from './$types';

	/** Confirmar la suscripción al newsletter con el link del mail (§50). */
	let { data, form }: PageProps = $props();
</script>

<Newsletter tienda={data.tienda} token={data.token} modo="confirmar" {form} />
