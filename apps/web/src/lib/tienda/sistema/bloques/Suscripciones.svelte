<script lang="ts">
	import { CADA, queIncluye } from '../../suscripciones.ts';
	import { usarTienda } from '../contexto.ts';
	import { precio } from '../formato.ts';
	import Foto from '../primitivas/Foto.svelte';
	import Grilla from '../primitivas/Grilla.svelte';
	import Seccion from '../primitivas/Seccion.svelte';

	/**
	 * Los planes de suscripción de la tienda (§51), con las piezas de las tarjetas de producto. No es
	 * un bloque del catálogo: lo pone Bloques.svelte después de los productos si la tienda tiene planes.
	 */
	const ctx = usarTienda();
	const planes = $derived(ctx.tienda.suscripciones);
	// El club de vinos de Malbec (su función propia); en las demás, «Suscripciones».
	const titulo = $derived(ctx.tienda.plantilla === 'malbec' ? 'Club de vinos' : 'Suscripciones');
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
<Seccion {titulo} link={{ t: 'Ver los planes', href: `${ctx.base}/suscripciones` }}>
	<Grilla items={planes} columnas={4} etiqueta="Planes de suscripción">
		{#snippet children(p)}
			<article class="tv-prod tv-prod--grilla">
				<a
					class="tv-prod-foto"
					href="{ctx.base}/suscripciones/{p.id}"
					tabindex="-1"
					aria-hidden="true"
				>
					<Foto
						src={p.foto ?? p.productos.find((x) => x.foto)?.foto ?? null}
						alt=""
						nombre={p.nombre}
						ancho={600}
						proporcion="var(--t-foto)"
					/>
					{#if p.agotado}<span class="tv-chip tv-prod-chip">Sin lugar</span>{/if}
				</a>
				<div class="tv-prod-info">
					<a class="tv-prod-nombre" href="{ctx.base}/suscripciones/{p.id}">{p.nombre}</a>
					{#if queIncluye(p)}<span class="tv-prod-detalle">{queIncluye(p)}</span>{/if}
					<span class="tv-precio"
						><span class="tv-precio-fila"
							><b class="tv-precio-monto">{precio(p.precio)}</b><span class="tv-suave tv-chico"
								>{CADA[p.frecuencia]}</span
							></span
						></span
					>
				</div>
			</article>
		{/snippet}
	</Grilla>
</Seccion>
