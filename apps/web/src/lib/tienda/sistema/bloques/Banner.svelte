<script lang="ts">
	import { campo, textoDe, type Bloque, type BotonDatos, type Fondo } from '../catalogo.ts';
	import { usarTienda } from '../contexto.ts';
	import { enlace } from '../datos.ts';
	import { sinViuda } from '../formato.ts';
	import Boton from '../primitivas/Boton.svelte';
	import Foto from '../primitivas/Foto.svelte';

	/**
	 * Banner: una promo con color o foto, texto y botón.
	 * - `banda`: una franja de color con el texto y el botón al costado (B-moda «10 % off…»).
	 * - `foto`: la foto con velo y el texto encima.
	 * - `doble`: dos promos lado a lado (la segunda es `segundo`; sin ella, va la primera sola).
	 */
	let { bloque }: { bloque: Bloque } = $props();
	const ctx = usarTienda();
	type Promo = { titulo?: string; texto?: string; foto?: string; boton?: BotonDatos };
	const fondo = $derived(campo<Fondo>(bloque, 'fondo') ?? 'oscuro');
	const promo = (p: Promo) => ({
		titulo: (p.titulo ?? '').trim(),
		texto: (p.texto ?? '').trim(),
		foto: p.foto ?? null,
		boton:
			p.boton?.t?.trim() && enlace(ctx.tienda, p.boton.a)
				? { t: p.boton.t, href: enlace(ctx.tienda, p.boton.a)! }
				: null
	});
	const uno = $derived(
		promo({
			titulo: textoDe(bloque, 'titulo'),
			texto: textoDe(bloque, 'texto'),
			foto: campo<string>(bloque, 'foto'),
			boton: campo<BotonDatos>(bloque, 'boton')
		})
	);
	const dos = $derived(promo(campo<Promo>(bloque, 'segundo') ?? {}));
	const promos = $derived(bloque.variante === 'doble' && dos.titulo ? [uno, dos] : [uno]);
	const conFoto = $derived(bloque.variante !== 'banda');
</script>

<section class="tv-banner tv-ancho" data-variante={bloque.variante} data-n={promos.length}>
	{#each promos as p, i (i)}
		<div
			class="tv-banner-caja tv-z"
			data-fondo={conFoto && p.foto ? 'oscuro' : i === 1 ? 'acento' : fondo}
		>
			{#if conFoto && p.foto}<Foto src={p.foto} alt="" nombre={p.titulo} ancho={1400} />
				<div class="tv-velo"></div>{/if}
			<div class="tv-banner-textos">
				<h2 class="tv-h2">{sinViuda(p.titulo)}</h2>
				{#if p.texto}<p class="tv-lead">{p.texto}</p>{/if}
			</div>
			{#if p.boton}<Boton href={p.boton.href} tipo={conFoto && p.foto ? 'claro' : 'primario'}
					>{p.boton.t}</Boton
				>{/if}
		</div>
	{/each}
</section>
