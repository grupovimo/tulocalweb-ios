<script lang="ts">
	import { campo, textoDe, type Bloque } from '../catalogo.ts';
	import { fotosUsadasDe, usarTienda } from '../contexto.ts';
	import { enlace, fotosGaleria } from '../datos.ts';
	import Foto from '../primitivas/Foto.svelte';
	import Seccion from '../primitivas/Seccion.svelte';

	/**
	 * Galería: fotos propias o las de los productos, siempre en filas completas (datos.ts ›
	 * fotosGaleria). `fila`: cuadradas, 6 por fila en la compu y 3 en el celular (A-moda «Usado de
	 * verdad»); `mosaico`: la primera grande y las demás alrededor (de a 6).
	 */
	let { bloque }: { bloque: Bloque } = $props();
	const ctx = usarTienda();
	const fotos = $derived(fotosGaleria(ctx.tienda, bloque, fotosUsadasDe(bloque.id)));
	const n = $derived(fotos.length);
	const mosaico = $derived(bloque.variante === 'mosaico' && n % 6 === 0);
	const href = $derived(enlace(ctx.tienda, campo<string>(bloque, 'link')));
	const ig = $derived(ctx.tienda.estilo?.redes?.instagram?.trim().replace(/^@/, '') ?? '');
	const link = $derived(
		href
			? {
					t: href.includes('instagram.com/')
						? `@${href.split('instagram.com/')[1].replace(/\/.*$/, '')}`
						: 'Ver más',
					href
				}
			: ig
				? {
						t: `@${ig.replace(/^https?:\/\/[^/]+\//, '')}`,
						href: `https://instagram.com/${ig.replace(/^https?:\/\/[^/]+\//, '')}`
					}
				: null
	);
	const cols = $derived(n >= 6 ? 6 : n);
	const colsM = $derived(n === 4 ? 2 : 3);
</script>

<Seccion titulo={textoDe(bloque, 'titulo')} {link} chico>
	<ul
		class="tv-galeria"
		class:tv-galeria--mosaico={mosaico}
		style:--cols={cols}
		style:--cols-m={colsM}
		aria-label="Fotos"
	>
		{#each fotos as f, i (f)}<li>
				<Foto
					src={f}
					alt=""
					nombre={ctx.tienda.nombre}
					ancho={mosaico && i % 6 === 0 ? 900 : 480}
				/>
			</li>{/each}
	</ul>
</Seccion>
