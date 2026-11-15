<script lang="ts">
	import type { ProximosTurnos, TiendaPublica } from '../../tipos.ts';
	import Bloques from '../bloques/Bloques.svelte';
	import type { Bloque } from '../catalogo.ts';
	import { categoriaHref, localDe, seDibuja } from '../datos.ts';
	import { cuantos } from '../formato.ts';
	import BarraPedido from '../marco/BarraPedido.svelte';
	import Bolsa from '../marco/Bolsa.svelte';
	import Tienda from '../marco/Tienda.svelte';
	import { disenoDeTienda } from '../migrar.ts';
	import { estiloFinal } from '../presets.ts';

	/**
	 * La portada de la tienda: los bloques de `diseno.inicio` dentro del marco. Si con lo que hay se
	 * dibujan menos de tres bloques (una tienda recién abierta), se suman los beneficios en tarjetas
	 * y la newsletter, así nunca se ve vacía (regla 1, A/B/C-vacia).
	 *
	 * En Pedido, como C-moda: las categorías al costado (si hay dos o más), los bloques al medio y la
	 * bolsa siempre visible a la derecha; en el celular, la barra fija del pedido.
	 */
	let { tienda, proximos = null }: { tienda: TiendaPublica; proximos?: ProximosTurnos | null } =
		$props();

	const diseno = $derived(disenoDeTienda(tienda));
	const dir = $derived(estiloFinal(tienda.plantilla, diseno.paleta, diseno.estilo).direccion);
	const RELLENO: Bloque[] = [
		{ id: 'relleno1', tipo: 'beneficios', variante: 'tarjetas', datos: {} },
		{ id: 'relleno2', tipo: 'newsletter', variante: 'banda', datos: {} }
	];
	const bloques = $derived.by(() => {
		const lista = [...diseno.inicio];
		for (const r of RELLENO) {
			if (lista.filter((b) => seDibuja(b, tienda)).length >= 3) break;
			if (!lista.some((b) => b.tipo === r.tipo && !b.oculto)) lista.push(r);
		}
		return lista;
	});
	const cats = $derived(tienda.categorias.filter((c) => c.cantidad > 0));
	const local = $derived(localDe(tienda, { id: 'local', tipo: 'local', variante: 'tarjeta', datos: {} }));
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
<Tienda {tienda} pagina="inicio">
	{#if dir === 'pedido'}
		<div class="tv-pedido tv-ancho" data-cats={cats.length >= 2}>
			{#if cats.length >= 2}
				<nav class="tv-pedido-cats tv-solo-compu" aria-label="Categorías">
					{#each cats as c (c.nombre)}<a href={categoriaHref(tienda, c.nombre)}
							>{c.nombre}<span>{cuantos(c.cantidad, 'producto', 'productos').split(' ')[0]}</span></a
						>{/each}
					{#if local.direccion}
						<div class="tv-pedido-local tv-tarjeta">
							<b>El local</b>
							<span>{local.direccion}</span>
							{#each local.horarios as h, i (i)}<span>{h.dias} · {h.horas}</span>{/each}
							{#if local.mapa}<a href={local.mapa} target="_blank" rel="noopener">Cómo llegar</a>{/if}
						</div>
					{/if}
				</nav>
			{/if}
			<div class="tv-pedido-main"><Bloques {bloques} {proximos} /></div>
			<div class="tv-pedido-lado tv-solo-compu"><Bolsa /></div>
		</div>
		<BarraPedido />
	{:else}
		<Bloques {bloques} {proximos} />
	{/if}
</Tienda>
