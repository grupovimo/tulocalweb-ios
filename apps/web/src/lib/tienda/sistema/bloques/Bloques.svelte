<script lang="ts">
	import type { Component } from 'svelte';
	import type { ProximosTurnos } from '../../tipos.ts';
	import type { Bloque, TipoBloque } from '../catalogo.ts';
	import { ponerFotosUsadas, usarTienda } from '../contexto.ts';
	import { fotosGaleria, fotosUsadas, productosQueSuman, seDibuja } from '../datos.ts';
	import Banner from './Banner.svelte';
	import Beneficios from './Beneficios.svelte';
	import Categorias from './Categorias.svelte';
	import Clientes from './Clientes.svelte';
	import Combo from './Combo.svelte';
	import Destacado from './Destacado.svelte';
	import Galeria from './Galeria.svelte';
	import Local from './Local.svelte';
	import Logos from './Logos.svelte';
	import Menu from './Menu.svelte';
	import Newsletter from './Newsletter.svelte';
	import Oferta from './Oferta.svelte';
	import Portada from './Portada.svelte';
	import Preguntas from './Preguntas.svelte';
	import Productos from './Productos.svelte';
	import Resenas from './Resenas.svelte';
	import Suscripciones from './Suscripciones.svelte';
	import TextoImagen from './TextoImagen.svelte';
	import Turnos from './Turnos.svelte';
	import Video from './Video.svelte';
	import Videos from './Videos.svelte';
	import Whatsapp from './Whatsapp.svelte';

	/**
	 * Recorre una lista de bloques (la portada, la ficha o Nosotros) y dibuja los que tienen con qué
	 * (datos.ts › seDibuja), cada uno envuelto con `data-bloque={id}`: la vista previa del editor lo
	 * usa para saber qué bloque se tocó. Después del primer bloque de productos van los planes de
	 * suscripción, si la tienda tiene (no son un bloque del catálogo: salen de la tienda).
	 */
	let {
		bloques,
		proximos = null,
		suscripciones = true
	}: {
		bloques: Bloque[];
		/** Los libres de hoy, para el bloque de turnos. */
		proximos?: ProximosTurnos | null;
		/** Dibujar los planes de suscripción después del primer bloque de productos. */
		suscripciones?: boolean;
	} = $props();
	const ctx = usarTienda();

	const PIEZAS: Record<
		TipoBloque,
		Component<{ bloque: Bloque; proximos?: ProximosTurnos | null }>
	> = {
		portada: Portada,
		banner: Banner,
		oferta: Oferta,
		textoImagen: TextoImagen,
		video: Video,
		beneficios: Beneficios,
		productos: Productos,
		destacado: Destacado,
		combo: Combo,
		categorias: Categorias,
		menu: Menu,
		local: Local,
		resenas: Resenas,
		preguntas: Preguntas,
		logos: Logos,
		galeria: Galeria,
		clientes: Clientes,
		videos: Videos,
		newsletter: Newsletter,
		whatsapp: Whatsapp,
		turnos: Turnos
	};
	const conDatos = $derived(bloques.filter((b) => seDibuja(b, ctx.tienda)));
	// Las fotos que ya muestra la página, para que lo automático no las repita (la galería que sólo
	// repetiría la grilla no se dibuja).
	const usadas = $derived(fotosUsadas(ctx.tienda, conDatos));
	ponerFotosUsadas(() => usadas);
	// Una sección de productos que repite lo de las anteriores no se dibuja (datos.ts).
	const repetidos = $derived(productosQueSuman(ctx.tienda, conDatos));
	const visibles = $derived(
		conDatos.filter(
			(b) =>
				!repetidos.has(b.id) &&
				(b.tipo !== 'galeria' || fotosGaleria(ctx.tienda, b, usadas.get(b.id)).length > 0)
		)
	);
	const conPlanes = $derived(
		suscripciones && ctx.tienda.suscripciones.length
			? (visibles.find((b) => b.tipo === 'productos' || b.tipo === 'menu')?.id ??
					visibles.at(-1)?.id)
			: null
	);
</script>

<div class="tv-bloques">
	{#each visibles as b (b.id)}
		{@const Pieza = PIEZAS[b.tipo]}
		<div class="tv-b" data-bloque={b.id} data-tipo={b.tipo}>
			<Pieza bloque={b} {proximos} />
		</div>
		{#if b.id === conPlanes}
			<div class="tv-b" data-tipo="suscripciones"><Suscripciones /></div>
		{/if}
	{/each}
</div>
