<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { promedio } from '../formato.ts';

	/** Cinco estrellas (llenas, media o vacías) con el valor dicho en palabras para los lectores. */
	let { valor, tam = 15 }: { valor: number; tam?: number } = $props();
	const iconos = $derived(
		Array.from({ length: 5 }, (_, i) =>
			valor >= i + 0.75 ? I.phStarFill : valor >= i + 0.25 ? I.phStarHalfFill : I.phStar
		)
	);
</script>

<span class="tv-estrellas" role="img" aria-label="{promedio(valor)} de 5 estrellas"
	>{#each iconos as e, i (i)}<Icon icon={e} size={tam} />{/each}</span
>
