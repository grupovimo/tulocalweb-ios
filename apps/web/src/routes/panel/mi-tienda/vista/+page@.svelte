<script lang="ts">
	import { onMount, tick } from 'svelte';
	import AnuncioTienda from '#lib/tienda/AnuncioTienda.svelte';
	import Plantilla from '#lib/tienda/Plantilla.svelte';
	import {
		avisar,
		conDiseno,
		leerMensaje,
		paginaVista,
		type PaginaVista
	} from '#lib/panel/mi-tienda/vista.ts';
	import type { DisenoV2 } from '#lib/panel/mi-tienda/diseno.ts';

	/**
	 * Vista previa de Mi tienda: la página pedida con el borrador. Escucha al editor
	 * (postMessage de la web, window.tiendaVista de las apps) y vuelve a dibujar sin recargar. Con
	 * marcos, cada bloque se resalta al pasar y al tocarlo avisa cuál es. Los links no navegan.
	 */
	let { data } = $props();
	// svelte-ignore state_referenced_locally
	let diseno = $state.raw<DisenoV2>(data.diseno);
	// svelte-ignore state_referenced_locally
	let pagina = $state<PaginaVista>(data.pagina);
	let elegido = $state<string | null>(null);
	const tienda = $derived(conDiseno(data.tienda, diseno, data.plantilla));

	async function seleccionar(id: string | null) {
		elegido = id;
		await tick();
		marcar();
		if (!id) return;
		document
			.querySelector(`[data-bloque="${CSS.escape(id)}"]`)
			?.scrollIntoView({ behavior: 'smooth', block: 'center' });
	}
	/** El bloque elegido lleva su contorno (también después de volver a dibujar). */
	function marcar() {
		for (const el of document.querySelectorAll('[data-bloque].tv-sel'))
			el.classList.remove('tv-sel');
		if (elegido)
			document.querySelector(`[data-bloque="${CSS.escape(elegido)}"]`)?.classList.add('tv-sel');
	}
	async function mostrar(d: DisenoV2, p?: string) {
		diseno = d;
		if (p) pagina = paginaVista(p);
		await tick();
		marcar();
	}

	onMount(() => {
		const w = window as unknown as {
			tiendaVista?: {
				mostrar: (d: DisenoV2 | string, p?: string) => void;
				seleccionar: (id: string | null) => void;
			};
		};
		// Las apps: evaluateJavaScript("window.tiendaVista.mostrar(<json>, '<pagina>')").
		w.tiendaVista = {
			mostrar: (d, p) => void mostrar(typeof d === 'string' ? JSON.parse(d) : d, p),
			seleccionar: (id) => void seleccionar(id)
		};
		const alMensaje = (e: MessageEvent) => {
			if (e.origin !== location.origin || e.source !== window.parent) return;
			const m = leerMensaje(e.data);
			if (m?.tipo === 'diseno') void mostrar(m.diseno, m.pagina);
			else if (m?.tipo === 'seleccionar') void seleccionar(m.id);
		};
		addEventListener('message', alMensaje);
		avisar({ tipo: 'lista' });
		return () => {
			removeEventListener('message', alMensaje);
			delete w.tiendaVista;
		};
	});

	function tocar(e: MouseEvent) {
		const t = e.target as HTMLElement;
		const bloque = data.marcos ? t.closest<HTMLElement>('[data-bloque]') : null;
		if (t.closest('a,form button[type=submit]') || bloque) e.preventDefault();
		if (bloque?.dataset.bloque) {
			e.stopPropagation();
			elegido = bloque.dataset.bloque;
			marcar();
			avisar({ tipo: 'bloque', id: bloque.dataset.bloque });
		}
	}
</script>

<svelte:head>
	<title>Vista previa · {data.tienda.nombre}</title>
	<meta name="robots" content="noindex" />
	<link
		href="https://fonts.googleapis.com/css2?family=Instrument+Sans:wght@400;500;600;700&family=Instrument+Serif:ital@0;1&display=swap"
		rel="stylesheet"
	/>
</svelte:head>

<div class="vista" class:marcos={data.marcos} onclickcapture={tocar}>
	<AnuncioTienda {tienda} />
	{#if pagina === 'producto'}
		{#if data.producto}
			<Plantilla {tienda} pagina="producto" producto={data.producto} />
		{:else}
			<p class="sin">Cargá un producto para ver cómo queda la ficha.</p>
		{/if}
	{:else if pagina === 'inicio'}
		<Plantilla {tienda} pagina="inicio" proximos={null} />
	{:else}
		<Plantilla {tienda} {pagina} />
	{/if}
</div>

<style>
	.sin {
		margin: 0;
		padding: 64px 24px;
		text-align: center;
		font:
			500 16px/1.5 'Instrument Sans',
			system-ui,
			sans-serif;
		color: #6b5244;
	}
	/* Con marcos: el bloque bajo el mouse y el elegido llevan contorno, sin mover nada. */
	.marcos :global([data-bloque]) {
		cursor: pointer;
		outline-offset: -3px;
	}
	.marcos :global([data-bloque]:hover) {
		outline: 2px dashed #d2582f;
	}
	.marcos :global([data-bloque].tv-sel) {
		outline: 3px solid #d2582f;
	}
</style>
