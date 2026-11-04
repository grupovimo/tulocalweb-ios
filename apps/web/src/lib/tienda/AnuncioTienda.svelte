<script lang="ts">
	import { acentos } from './acentos.ts';
	import { colorValido } from './estilo/css.ts';
	import type { TiendaPublica } from './tipos.ts';

	/** Mi tienda › Estilo › Anuncio: la barra de arriba de todo, con el acento de la tienda. */
	let { tienda }: { tienda: Pick<TiendaPublica, 'estilo' | 'plantilla' | 'paleta'> } = $props();

	const estilo = $derived(tienda.estilo);
	const texto = $derived(estilo?.anuncio?.trim() || '');
	const link = $derived(estilo?.anuncioLink || '');
	const color = $derived.by(() => {
		const op = acentos(tienda.plantilla);
		const a = op[tienda.paleta] ?? op[0];
		return {
			fondo: colorValido(estilo?.colores?.acento) ?? a.color,
			texto: colorValido(estilo?.colores?.sobreAcento) ?? a.sobre
		};
	});
</script>

{#if texto}
	<div class="anuncio" style:background={color.fondo} style:color={color.texto}>
		<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- link que eligió la tienda -->
		{#if link}<a href={link}>{texto}</a>{:else}{texto}{/if}
	</div>
{/if}

<style>
	.anuncio {
		padding: 9px 16px;
		text-align: center;
		font:
			600 13.5px/1.35 'Instrument Sans',
			system-ui,
			sans-serif;
		text-wrap: balance;
	}
	.anuncio a {
		color: inherit;
		text-decoration: underline;
		text-underline-offset: 3px;
	}
</style>
