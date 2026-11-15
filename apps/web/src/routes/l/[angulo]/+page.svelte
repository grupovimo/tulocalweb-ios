<script lang="ts">
	import Aterrizaje from '#lib/anuncios/Aterrizaje.svelte';

	let { data } = $props();
</script>

{#key data.a.id}
	<Aterrizaje a={data.a} pixel={data.pixel} />
{/key}
