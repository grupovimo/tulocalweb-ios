<script lang="ts">
	import { campo, textoDe, type Bloque, type BotonDatos, type Fondo } from '../catalogo.ts';
	import { usarTienda } from '../contexto.ts';
	import { enlace } from '../datos.ts';
	import { sinViuda } from '../formato.ts';
	import Boton from '../primitivas/Boton.svelte';
	import Foto from '../primitivas/Foto.svelte';

	/**
	 * Texto con imagen: la historia, el taller, el equipo (A-moda «El taller»). `derecha` o
	 * `izquierda`: la foto a ese lado, en mitades; `fondo`: la foto detrás con velo; `texto`: sólo el
	 * texto, centrado. Sin foto, va sólo el texto.
	 */
	let { bloque }: { bloque: Bloque } = $props();
	const ctx = usarTienda();
	const foto = $derived(campo<string>(bloque, 'foto') ?? null);
	const variante = $derived(foto ? bloque.variante : 'texto');
	const boton = $derived.by(() => {
		const b = campo<BotonDatos>(bloque, 'boton');
		const href = b?.t?.trim() ? enlace(ctx.tienda, b.a) : null;
		return href ? { t: b!.t, href } : null;
	});
	const fondo = $derived(
		variante === 'fondo' ? 'oscuro' : (campo<Fondo>(bloque, 'fondo') ?? 'superficie')
	);
	const parrafos = $derived(
		textoDe(bloque, 'texto')
			.split(/\n{2,}|\n/)
			.map((p) => p.trim())
			.filter(Boolean)
	);
</script>

<section class="tv-ti tv-ancho" data-variante={variante}>
	<div class="tv-ti-caja tv-z" data-fondo={fondo}>
		{#if foto}<div class="tv-ti-foto">
				<Foto src={foto} alt="" nombre={textoDe(bloque, 'titulo')} ancho={1200} />
			</div>{/if}
		{#if variante === 'fondo'}<div class="tv-velo"></div>{/if}
		<div class="tv-ti-textos">
			{#if textoDe(bloque, 'etiqueta')}<p class="tv-ante">{textoDe(bloque, 'etiqueta')}</p>{/if}
			<h2 class="tv-h2">{sinViuda(textoDe(bloque, 'titulo'))}</h2>
			{#each parrafos as p, i (i)}<p class="tv-lead">{p}</p>{/each}
			{#if boton}<div class="tv-botones">
					<Boton href={boton.href} tipo={variante === 'fondo' ? 'claro' : 'secundario'}
						>{boton.t}</Boton
					>
				</div>{/if}
		</div>
	</div>
</section>
