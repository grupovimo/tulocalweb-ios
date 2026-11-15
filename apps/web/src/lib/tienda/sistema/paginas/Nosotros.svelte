<script lang="ts">
	import { comoLlega } from '../../envio.svelte.ts';
	import type { TiendaPublica } from '../../tipos.ts';
	import Bloques from '../bloques/Bloques.svelte';
	import { cuantos, promedio, sinViuda } from '../formato.ts';
	import Tienda from '../marco/Tienda.svelte';
	import { disenoDeTienda } from '../migrar.ts';
	import { NOSOTROS_BASE, opcionesPaginas } from '../paginas.ts';
	import { estiloFinal } from '../presets.ts';
	import Foto from '../primitivas/Foto.svelte';

	/**
	 * Nosotros: quiénes son (el nombre, el lema, la ciudad y cómo llega lo que venden), con la foto de
	 * la portada y tres datos reales; abajo, los bloques que la tienda armó para esta página
	 * (paginas.nosotros.bloques) o, si no armó nada, beneficios, el local, reseñas y WhatsApp.
	 */
	let { tienda }: { tienda: TiendaPublica } = $props();
	const diseno = $derived(disenoDeTienda(tienda));
	const dir = $derived(estiloFinal(tienda.plantilla, diseno.paleta, diseno.estilo).direccion);
	const propios = $derived(opcionesPaginas(tienda, dir).nosotros.bloques);
	const bloques = $derived(propios.length ? propios : NOSOTROS_BASE);
	const titulo = $derived(diseno.titulo?.trim() || tienda.portada.titulo?.trim() || tienda.nombre);
	const lema = $derived(diseno.lema?.trim() || tienda.portada.lema?.trim() || '');
	const foto = $derived(diseno.portada || tienda.portada.foto || tienda.fotos[0] || null);
	const datos = $derived(
		[
			tienda.productos.length > 1
				? {
						n: String(tienda.productos.length),
						t: tienda.productos.length === 1 ? 'producto' : 'productos'
					}
				: null,
			tienda.cobros.mercadoPago && tienda.cobros.cuotas > 1
				? { n: String(tienda.cobros.cuotas), t: 'cuotas sin interés' }
				: null,
			tienda.promedio && tienda.resenas.length
				? {
						n: promedio(tienda.promedio),
						t: `de promedio en ${cuantos(tienda.resenas.length, 'reseña', 'reseñas')}`
					}
				: null
		].filter((x) => !!x)
	);
</script>

<Tienda {tienda} pagina="nosotros">
	<section class="tv-ancho tv-pag tv-nos" class:tv-nos--foto={!!foto}>
		<div class="tv-nos-texto">
			<p class="tv-ante">{tienda.ciudad ? `${tienda.nombre} · ${tienda.ciudad}` : tienda.nombre}</p>
			<h1 class="tv-h1 tv-pag-titulo">{sinViuda(titulo)}</h1>
			{#if lema}<p class="tv-lead">{lema}</p>{/if}
			<p class="tv-lead">{comoLlega(tienda.envios)}</p>
			{#if datos.length}
				<dl class="tv-nos-datos">
					{#each datos as d (d.t)}<div>
							<dt>{d.t}</dt>
							<dd>{d.n}</dd>
						</div>{/each}
				</dl>
			{/if}
		</div>
		{#if foto}
			<div class="tv-nos-foto">
				<Foto src={foto} alt="" nombre={tienda.nombre} ancho={1100} proporcion="4/5" prioridad />
			</div>
		{/if}
	</section>
	<div class="tv-nos-bloques"><Bloques {bloques} suscripciones={false} /></div>
</Tienda>
