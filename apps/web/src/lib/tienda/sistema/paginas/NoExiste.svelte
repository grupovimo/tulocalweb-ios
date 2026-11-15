<script lang="ts">
	import VacioTienda from '../../VacioTienda.svelte';
	import type { TiendaPublica } from '../../tipos.ts';
	import Bloques from '../bloques/Bloques.svelte';
	import type { Bloque } from '../catalogo.ts';
	import Tienda from '../marco/Tienda.svelte';
	import { disenoDeTienda } from '../migrar.ts';
	import { temaDe } from '../paginas.ts';
	import { estiloFinal } from '../presets.ts';

	/**
	 * Un producto o una página que no existe, con el marco de la tienda (no la página de error de Tu
	 * Local Web): qué pasó, cómo seguir y, si hay, los más vendidos para no quedar en blanco.
	 */
	let { tienda }: { tienda: TiendaPublica } = $props();
	const diseno = $derived(disenoDeTienda(tienda));
	const dir = $derived(estiloFinal(tienda.plantilla, diseno.paleta, diseno.estilo).direccion);
	const sugeridos: Bloque[] = [
		{
			id: 'noexiste-prods',
			tipo: 'productos',
			variante: 'grilla',
			datos: { titulo: 'Lo más vendido', productos: { fuente: 'vendidos', cantidad: 4 } }
		}
	];
</script>

<Tienda {tienda} pagina="no-existe">
	<div class="tv-ancho tv-pag"><VacioTienda {tienda} tema={temaDe(dir)} motivo="no-existe" /></div>
	{#if tienda.productos.length > 1}<Bloques bloques={sugeridos} suscripciones={false} />{/if}
</Tienda>
