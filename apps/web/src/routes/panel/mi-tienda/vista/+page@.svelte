<script lang="ts">
	import AnuncioTienda from '#lib/tienda/AnuncioTienda.svelte';
	import Plantilla from '#lib/tienda/Plantilla.svelte';

	/** Vista previa de Mi tienda: la portada con el borrador. Los links no navegan (es una vista). */
	let { data } = $props();
</script>

<svelte:head>
	<title>Vista previa · {data.tienda.nombre}</title>
	<meta name="robots" content="noindex" />
	<link
		href="https://fonts.googleapis.com/css2?family=Instrument+Sans:wght@400;500;600;700&family=Instrument+Serif:ital@0;1&display=swap"
		rel="stylesheet"
	/>
</svelte:head>

<div
	class="vista"
	onclickcapture={(e) => {
		if ((e.target as HTMLElement).closest('a,form button[type=submit]')) e.preventDefault();
	}}
>
	<AnuncioTienda tienda={data.tienda} />
	<Plantilla tienda={data.tienda} pagina="inicio" proximos={null} />
</div>
