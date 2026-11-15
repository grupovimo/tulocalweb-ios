<script lang="ts">
	import { campo, textoDe, type Bloque } from '../catalogo.ts';
	import { usarTienda } from '../contexto.ts';
	import { enlace } from '../datos.ts';
	import { urlFoto } from '../formato.ts';
	import { medirLogo } from '../logos.ts';

	/**
	 * Logos de prensa, marcas o con quién trabaja la tienda: una fila centrada, con sus colores y a
	 * pleno. El que no contrasta con el fondo (uno negro en una tienda oscura) se pasa entero a claro
	 * u oscuro (logos.ts).
	 */
	let { bloque }: { bloque: Bloque } = $props();
	const ctx = usarTienda();
	type Logo = { foto: string; nombre: string; link?: string };
	const items = $derived(
		(campo<Logo[]>(bloque, 'items') ?? [])
			.filter((l) => l.foto && l.nombre)
			.map((l) => ({ ...l, href: enlace(ctx.tienda, l.link) }))
	);
	const titulo = $derived(textoDe(bloque, 'titulo'));
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- links que eligió la tienda -->
<section class="tv-logos tv-ancho" aria-label={titulo || 'Logos'}>
	{#if titulo}<p class="tv-ante tv-logos-tit">{titulo}</p>{/if}
	<ul>
		{#each items as l (l.foto)}
			<li>
				{#if l.href}<a
						href={l.href}
						target={l.href.startsWith('https://') ? '_blank' : undefined}
						rel="noopener"
						><img src={urlFoto(l.foto, 320)} alt={l.nombre} loading="lazy" use:medirLogo /></a
					>{:else}<img
						src={urlFoto(l.foto, 320)}
						alt={l.nombre}
						loading="lazy"
						use:medirLogo
					/>{/if}
			</li>
		{/each}
	</ul>
</section>
