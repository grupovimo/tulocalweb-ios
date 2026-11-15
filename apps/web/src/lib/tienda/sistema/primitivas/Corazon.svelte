<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import type { ProductoPublico } from '../../tipos.ts';
	import { usarTienda } from '../contexto.ts';

	/**
	 * El corazón de favoritos (conversion.favoritos): en la tarjeta va sobre la foto, arriba a la
	 * derecha; `grande`, en la ficha, con texto. Es un botón con aria-pressed; los favoritos quedan en
	 * el dispositivo (favoritos.svelte.ts) y se ven en /favoritos.
	 */
	let { producto, grande = false }: { producto: ProductoPublico; grande?: boolean } = $props();
	const ctx = usarTienda();
	const on = $derived(ctx.favoritos.tiene(producto.id));
</script>

<button
	type="button"
	class={grande ? 'tv-btn tv-btn--suave tv-btn--chico tv-corazon-grande' : 'tv-corazon'}
	aria-pressed={on}
	aria-label={grande ? undefined : `Guardar ${producto.nombre} en favoritos`}
	onclick={() => ctx.favoritos.alternar(producto.id)}
	><Icon icon={on ? I.phHeartFill : I.phHeart} size={grande ? 18 : 20} />{#if grande}{on
			? 'En favoritos'
			: 'Guardar'}{/if}</button
>
