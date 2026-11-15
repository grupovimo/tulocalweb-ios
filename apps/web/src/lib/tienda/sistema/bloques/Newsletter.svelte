<script lang="ts">
	import { AltaNewsletter } from '../../newsletter.svelte.ts';
	import { campo, textoDe, type Bloque, type Fondo } from '../catalogo.ts';
	import { usarTienda } from '../contexto.ts';
	import { sinViuda } from '../formato.ts';

	/**
	 * Newsletter (§50): junta emails; queda pendiente hasta confirmar por mail (en la demo, el aviso
	 * trae el link). `centrado`: título, texto y el campo al medio; `banda`: franja con el texto a la
	 * izquierda y el campo a la derecha (B-moda «10 % off en tu primera compra»).
	 */
	let { bloque }: { bloque: Bloque } = $props();
	const ctx = usarTienda();
	const nl = new AltaNewsletter(() => ctx.tienda.slug);
	const fondo = $derived(campo<Fondo>(bloque, 'fondo') ?? 'oscuro');
	const texto = $derived(textoDe(bloque, 'texto') || 'Un mail cada tanto con lo nuevo y las promos. Nada más.');
	const id = $derived(`nl-${bloque.id}`);
</script>

<section class="tv-nl tv-ancho" data-variante={bloque.variante}>
	<div class="tv-nl-caja tv-z" data-fondo={fondo}>
		<div class="tv-nl-textos">
			<h2 class="tv-h2">{sinViuda(textoDe(bloque, 'titulo'))}</h2>
			<p class="tv-lead">{texto}</p>
		</div>
		<div class="tv-nl-form-caja">
			<form class="tv-nl-form" onsubmit={nl.enviar} novalidate>
				<label for={id} class="tv-oculto">Tu email</label>
				<input
					{id}
					type="email"
					name="email"
					required
					maxlength="120"
					autocomplete="email"
					placeholder="Tu email"
					bind:value={nl.email}
				/>
				<button type="submit" class="tv-btn tv-btn--primario" disabled={nl.enviando}>Sumarme</button>
			</form>
			{#if nl.aviso}<p class="tv-chico" role="status">
					<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- link de confirmación -->
					{nl.aviso.texto}{#if nl.aviso.link}&nbsp;<a href={nl.aviso.link}>Confirmar acá</a>{/if}
				</p>{/if}
		</div>
	</div>
</section>
