<script lang="ts">
	import CampoBloque from './CampoBloque.svelte';
	import { ajustesDe } from './bloques.ts';
	import type { Ayuda } from './campos.ts';
	import { PAGINAS } from './catalogo.ts';
	import type { DisenoV2 } from './diseno.ts';

	/** Los ajustes de Catálogo, Ficha y Carrito: el diseño de la página y sus opciones. */
	let {
		pagina,
		diseno,
		ayuda,
		oncambiar
	}: {
		pagina: 'catalogo' | 'producto' | 'carrito';
		diseno: DisenoV2;
		ayuda: Ayuda;
		oncambiar: (d: DisenoV2, confirmar: boolean) => void;
	} = $props();

	const def = $derived(PAGINAS[pagina]);
	const ajustes = $derived(ajustesDe(diseno, pagina) as Record<string, unknown>);

	function poner(k: string, v: unknown, confirmar: boolean) {
		oncambiar(
			{ ...diseno, paginas: { ...diseno.paginas, [pagina]: { ...ajustes, [k]: v } } },
			confirmar
		);
	}
</script>

<div class="ajustes">
	{#if def.variantes}
		<CampoBloque
			campo={{
				id: 'variante',
				tipo: 'opcion',
				nombre: 'Diseño',
				opciones: def.variantes,
				nombres: def.nombresVariantes
			}}
			valor={ajustes.variante}
			{ayuda}
			onvalor={(v, c) => poner('variante', v, c)}
		/>
	{/if}
	{#each def.campos ?? [] as c (c.id)}
		<CampoBloque
			campo={c}
			valor={ajustes[c.id]}
			{ayuda}
			onvalor={(v, conf) => poner(c.id, v, conf)}
		/>
	{/each}
</div>

<style>
	.ajustes {
		display: flex;
		flex-direction: column;
		gap: 14px;
		padding: 4px;
	}
</style>
