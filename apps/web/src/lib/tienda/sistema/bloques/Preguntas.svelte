<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { campo, textoDe, type Bloque } from '../catalogo.ts';
	import { sinViuda } from '../formato.ts';

	/**
	 * Preguntas frecuentes en acordeón (details/summary reales): el título a un lado y las preguntas
	 * al otro en la compu (A-cafe), una debajo de la otra en el celular. La primera, abierta.
	 */
	let { bloque }: { bloque: Bloque } = $props();
	const items = $derived(
		(campo<{ p: string; r: string }[]>(bloque, 'items') ?? []).filter(
			(x) => x.p?.trim() && x.r?.trim()
		)
	);
</script>

<section class="tv-faq tv-ancho">
	<h2 class="tv-h2 tv-h2--chico">{sinViuda(textoDe(bloque, 'titulo'))}</h2>
	<div class="tv-faq-lista">
		{#each items as q, i (i)}
			<details open={i === 0}>
				<summary><span>{q.p}</span><Icon icon={I.phPlus} size={18} /></summary>
				{#each q.r.split(/\n+/) as r, j (j)}<p>{r}</p>{/each}
			</details>
		{/each}
	</div>
</section>
