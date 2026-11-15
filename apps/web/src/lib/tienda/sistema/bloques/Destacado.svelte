<script lang="ts">
	import { campo, textoDe, type Bloque, type Fondo } from '../catalogo.ts';
	import { usarTienda } from '../contexto.ts';
	import { destacadoDe } from '../datos.ts';
	import UnProducto from './UnProducto.svelte';

	/** Producto destacado: uno solo, grande, con botón de compra (`partido` o `foto`; A-cafe «Café del mes»). */
	let { bloque }: { bloque: Bloque } = $props();
	const ctx = usarTienda();
	const p = $derived(destacadoDe(ctx.tienda, bloque));
</script>

{#if p}
	<UnProducto
		producto={p}
		ante={textoDe(bloque, 'etiqueta')}
		texto={textoDe(bloque, 'texto')}
		variante={bloque.variante === 'foto' ? 'foto' : 'partido'}
		fondo={campo<Fondo>(bloque, 'fondo') ?? 'oscuro'}
	/>
{/if}
