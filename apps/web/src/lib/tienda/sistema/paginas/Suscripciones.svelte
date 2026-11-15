<script lang="ts">
	import type { TiendaPublica } from '../../tipos.ts';
	import Planes from '../bloques/Suscripciones.svelte';
	import Tienda from '../marco/Tienda.svelte';

	/** Todos los planes de suscripción de la tienda (la pestaña «Suscripciones» del menú). */
	let { tienda }: { tienda: TiendaPublica } = $props();
</script>

<Tienda {tienda} pagina="suscripciones">
	<div class="tv-pag tv-pag--bloques"><Planes /></div>
</Tienda>
