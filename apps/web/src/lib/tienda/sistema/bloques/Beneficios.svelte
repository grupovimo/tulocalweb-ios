<script lang="ts" module>
	import * as I from '#lib/icons.ts';
	import type { Icono } from '../catalogo.ts';

	/** El ícono de Phosphor de cada beneficio del catálogo (bloques.json › iconos). */
	export const ICONO: Record<Icono, I.IconData> = {
		envio: I.phTruck,
		cuotas: I.phCreditCard,
		transferencia: I.phBank,
		cambios: I.phArrowsClockwise,
		retiro: I.phStorefront,
		seguro: I.phShieldCheck,
		regalo: I.phGift,
		'hecho-a-mano': I.phHandHeartBold,
		reloj: I.phClock,
		estrella: I.phStar
	};
</script>

<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import type { Bloque } from '../catalogo.ts';
	import { usarTienda } from '../contexto.ts';
	import { beneficiosDe } from '../datos.ts';

	/**
	 * Beneficios: envío, cuotas, cambios, retiro. Vacío, salen solos de la configuración de la tienda
	 * (datos.ts › beneficiosAuto). `fila`: una franja con íconos y bordes (A-moda); `tarjetas`: una
	 * tarjeta por beneficio (B-moda, C-moda). Las columnas, según cuántos son.
	 */
	let { bloque }: { bloque: Bloque } = $props();
	const ctx = usarTienda();
	const items = $derived(beneficiosDe(ctx.tienda, bloque));
</script>

<section class="tv-benef tv-ancho" data-variante={bloque.variante} style:--cols={items.length}>
	<ul class="tv-benef-lista" aria-label="Beneficios">
		{#each items as b (b.titulo)}
			<li class="tv-benef-item">
				<span class="tv-benef-icono"><Icon icon={ICONO[b.icono] ?? I.phStar} size={24} /></span>
				<span class="tv-benef-textos"
					><b>{b.titulo}</b>{#if b.texto}<span>{b.texto}</span>{/if}</span
				>
			</li>
		{/each}
	</ul>
</section>
