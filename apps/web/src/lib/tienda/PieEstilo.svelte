<script lang="ts">
	import type { TiendaPublica } from './tipos.ts';

	/**
	 * Mi tienda › Estilo en el pie de cualquier plantilla: la línea propia y las redes. Va con el color
	 * y la fuente del pie (currentColor, inherit), así queda bien en todas.
	 */
	let { tienda }: { tienda: Pick<TiendaPublica, 'estilo'> } = $props();

	const limpio = (s: string | undefined) => (s ?? '').trim();
	const usuario = (s: string | undefined) =>
		limpio(s)
			.replace(/^@/, '')
			.replace(/^https?:\/\/[^/]+\//, '');
	const web = (s: string, base: string) =>
		/^https?:\/\//.test(s) ? s : base + s.replace(/^@/, '');

	const r = $derived(tienda.estilo?.redes ?? {});
	const links = $derived(
		[
			limpio(r.instagram) && { t: 'Instagram', u: `https://instagram.com/${usuario(r.instagram)}` },
			limpio(r.tiktok) && { t: 'TikTok', u: `https://www.tiktok.com/@${usuario(r.tiktok)}` },
			limpio(r.facebook) && { t: 'Facebook', u: web(limpio(r.facebook), 'https://facebook.com/') },
			limpio(r.youtube) && { t: 'YouTube', u: web(limpio(r.youtube), 'https://youtube.com/@') },
			limpio(r.email) && { t: limpio(r.email), u: `mailto:${limpio(r.email)}` },
			limpio(r.telefono) && {
				t: limpio(r.telefono),
				u: `tel:${limpio(r.telefono).replace(/[^\d+]/g, '')}`
			}
		].filter((x): x is { t: string; u: string } => !!x && /^(https:|mailto:|tel:)/.test(x.u))
	);
	const leyenda = $derived(limpio(tienda.estilo?.leyendaPie));
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->

{#if leyenda || links.length}
	<div class="pie-estilo">
		{#if links.length}
			<nav aria-label="Redes y contacto">
				{#each links as l (l.u)}<a
						href={l.u}
						target={l.u.startsWith('https') ? '_blank' : undefined}
						rel="noopener">{l.t}</a
					>{/each}
			</nav>
		{/if}
		{#if leyenda}<p>{leyenda}</p>{/if}
	</div>
{/if}

<style>
	.pie-estilo {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 10px;
		padding: 18px 16px 4px;
		font-size: 14px;
		text-align: center;
	}
	nav {
		display: flex;
		flex-wrap: wrap;
		justify-content: center;
		gap: 6px 18px;
	}
	a {
		color: inherit;
		font-weight: 600;
		text-decoration: underline;
		text-underline-offset: 3px;
		text-decoration-thickness: 1px;
	}
	p {
		margin: 0;
		opacity: 0.8;
		text-wrap: balance;
		max-width: 60ch;
	}
</style>
