<script lang="ts">
	import { campo, textoDe, type Bloque } from '../catalogo.ts';
	import { videoDe } from '../datos.ts';
	import { sinViuda } from '../formato.ts';

	/**
	 * Video de YouTube (sin cookies) o Vimeo. `ancho`: el video a todo el ancho del contenido;
	 * `conTexto`: el texto a un lado y el video al otro. Se carga recién al acercarse (loading=lazy).
	 */
	let { bloque }: { bloque: Bloque } = $props();
	const src = $derived(videoDe(campo<string>(bloque, 'url')));
	const titulo = $derived(textoDe(bloque, 'titulo'));
	const texto = $derived(textoDe(bloque, 'texto'));
</script>

{#if src}
	<section class="tv-video tv-ancho" data-variante={bloque.variante}>
		{#if bloque.variante === 'conTexto' && (titulo || texto)}
			<div class="tv-video-textos">
				{#if titulo}<h2 class="tv-h2">{sinViuda(titulo)}</h2>{/if}
				{#if texto}<p class="tv-lead">{texto}</p>{/if}
			</div>
		{:else if titulo}
			<div class="tv-cab"><h2 class="tv-h2">{sinViuda(titulo)}</h2></div>
		{/if}
		<div class="tv-video-marco">
			<iframe
				{src}
				title={titulo || 'Video'}
				loading="lazy"
				allow="accelerometer; encrypted-media; gyroscope; picture-in-picture; fullscreen"
				referrerpolicy="strict-origin-when-cross-origin"
			></iframe>
		</div>
	</section>
{/if}
