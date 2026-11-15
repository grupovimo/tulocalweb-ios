<script lang="ts">
	import { browser } from '$app/env';
	import { campo, textoDe, type Bloque, type Fondo } from '../catalogo.ts';
	import { usarTienda } from '../contexto.ts';
	import {
		botonesPortada,
		fotosPortada,
		fraseDe,
		inicial,
		retiroDe,
		tituloPortada
	} from '../datos.ts';
	import { promedio, sinViuda } from '../formato.ts';
	import Boton from '../primitivas/Boton.svelte';
	import Estrellas from '../primitivas/Estrellas.svelte';
	import Foto from '../primitivas/Foto.svelte';

	/**
	 * Portada: antetítulo, título (h1, sin viudas), frase y hasta dos botones.
	 * - `foto`: la foto entera con velo y el texto abajo (A-cafe, C-moda).
	 * - `partida`: texto a un lado y fotos al otro (A-moda; B y C en tarjetas).
	 * - `collage`: el texto en un bloque de color y tres fotos alrededor; con reseñas, el promedio
	 *   en una de las celdas (B-cafe).
	 * - `texto`: sólo texto sobre el color elegido (Bloques.dc.html).
	 * - `carrusel`: las fotos una tras otra con puntos, el texto encima.
	 * Sin fotos, cualquier variante queda como texto sobre el color de la marca (A/B/C-vacia).
	 */
	let { bloque }: { bloque: Bloque } = $props();
	const ctx = usarTienda();
	const t = $derived(ctx.tienda);
	const titulo = $derived(sinViuda(tituloPortada(t, bloque)));
	const frase = $derived(fraseDe(t, bloque));
	const ante = $derived(textoDe(bloque, 'etiqueta'));
	const fotos = $derived(fotosPortada(t, bloque));
	const botones = $derived(botonesPortada(t, bloque));
	const fondo = $derived((campo<Fondo>(bloque, 'fondo') ?? 'acento') as Fondo);
	const variante = $derived(fotos.length ? bloque.variante : 'texto');
	const retiro = $derived(retiroDe(t));
	const promedioVisible = $derived(
		t.promedio && t.resenas.filter((r) => r.conCompra).length >= 3 ? t.promedio : null
	);

	// Carrusel: una foto por vez, con puntos; cambia sola cada 6 s salvo «Reducir movimiento».
	let actual = $state(0);
	$effect(() => {
		if (!browser || variante !== 'carrusel' || fotos.length < 2) return;
		if (matchMedia('(prefers-reduced-motion: reduce)').matches) return;
		const i = setInterval(() => (actual = (actual + 1) % fotos.length), 6000);
		return () => clearInterval(i);
	});
</script>

{#snippet textos(claro: boolean)}
	<div class="tv-portada-textos">
		{#if ante}<p class="tv-ante">{ante}</p>{/if}
		<h1 class="tv-h1">{titulo}</h1>
		{#if frase}<p class="tv-lead">{frase}</p>{/if}
		{#if botones.length}
			<div class="tv-botones">
				{#each botones as b, i (b.href)}<Boton
						href={b.href}
						tipo={i === 0 ? (claro ? 'claro' : 'primario') : 'secundario'}
						externo={b.href.startsWith('https://wa.me')}>{b.t}</Boton
					>{/each}
			</div>
		{/if}
	</div>
{/snippet}

<section class="tv-portada" data-variante={variante}>
	{#if variante === 'foto'}
		<div class="tv-portada-foto tv-ancho-foto">
			<Foto src={fotos[0]} alt="" nombre={t.nombre} ancho={1800} prioridad />
			<div class="tv-velo"></div>
			<div class="tv-portada-sobre tv-ancho">{@render textos(true)}</div>
		</div>
	{:else if variante === 'carrusel'}
		<div class="tv-portada-foto tv-ancho-foto">
			{#each fotos as f, i (f)}
				<div class="tv-portada-slide" class:tv-activa={i === actual} aria-hidden={i !== actual}>
					<Foto src={f} alt="" nombre={t.nombre} ancho={1800} prioridad={i === 0} />
				</div>
			{/each}
			<div class="tv-velo"></div>
			<div class="tv-portada-sobre tv-ancho">{@render textos(true)}</div>
			{#if fotos.length > 1}
				<div class="tv-puntos" role="tablist" aria-label="Fotos de la portada">
					{#each fotos as f, i (f)}<button
							type="button"
							role="tab"
							aria-selected={i === actual}
							aria-label="Foto {i + 1} de {fotos.length}"
							class:tv-activa={i === actual}
							onclick={() => (actual = i)}
						></button>{/each}
				</div>
			{/if}
		</div>
	{:else if variante === 'partida'}
		<div class="tv-portada-partida tv-ancho">
			<div class="tv-portada-col">
				{@render textos(false)}
				{#if retiro}<p class="tv-portada-nota">{retiro}</p>{/if}
			</div>
			<div class="tv-portada-fotos" data-n={Math.min(fotos.length, 3)}>
				{#each fotos.slice(0, 3) as f, i (f)}<Foto
						src={f}
						alt=""
						nombre={t.nombre}
						ancho={i === 0 ? 1100 : 700}
						prioridad={i === 0}
					/>{/each}
			</div>
		</div>
	{:else if variante === 'collage'}
		<div class="tv-portada-collage tv-ancho" data-n={Math.min(fotos.length, 3)}>
			<div class="tv-portada-tile tv-z" data-fondo="oscuro">{@render textos(true)}</div>
			{#each fotos.slice(0, 2) as f, i (f)}<div class="tv-portada-celda tv-c{i + 1}">
					<Foto src={f} alt="" nombre={t.nombre} ancho={900} prioridad={i === 0} />
				</div>{/each}
			{#if promedioVisible}
				<div class="tv-portada-celda tv-c3 tv-z tv-portada-nota-tile" data-fondo="acento">
					<Estrellas valor={promedioVisible} tam={20} />
					<div>
						<p class="tv-portada-num">{promedio(promedioVisible)}</p>
						<p class="tv-chico">{t.resenas.filter((r) => r.conCompra).length} reseñas con compra</p>
					</div>
				</div>
			{:else if fotos[2]}
				<div class="tv-portada-celda tv-c3"><Foto src={fotos[2]} alt="" nombre={t.nombre} ancho={700} /></div>
			{/if}
		</div>
	{:else}
		<div class="tv-portada-texto tv-ancho">
			<div class="tv-z tv-portada-caja" data-fondo={fondo}>
				{#if ctx.dir === 'pedido'}<span class="tv-portada-inicial" aria-hidden="true"
						>{inicial(t.nombre)}</span
					>{/if}
				{@render textos(fondo === 'acento' || fondo === 'oscuro')}
			</div>
		</div>
	{/if}
</section>
