<script lang="ts">
	import { avisoMayorista, type LineaDetalle } from '../../lineas.ts';
	import { textoAviso } from '../../mayorista.ts';
	import { usarTienda } from '../contexto.ts';
	import { precio } from '../formato.ts';

	/**
	 * Las barritas de arriba del carrito (página y panel): cuánto falta para el envío gratis (o «¡Ya
	 * tenés envío gratis!») si la tienda lo ofrece y la página lo muestra, y cuánto falta para el
	 * precio mayorista (§49) o lo que ya se ahorra.
	 */
	let { lineas, envioGratis = true }: { lineas: LineaDetalle[]; envioGratis?: boolean } = $props();
	const ctx = usarTienda();
	const t = $derived(ctx.tienda);
	const subtotal = $derived(lineas.reduce((s, l) => s + l.subtotal, 0));
	// La página lo muestra (paginas.carrito.envioGratisDesde) y la tienda no lo apagó (conversion).
	const desde = $derived(envioGratis && ctx.conversion.envioGratis ? t.envios.gratisDesde : null);
	const falta = $derived(desde ? Math.max(0, desde - subtotal) : 0);
	const mayor = $derived(avisoMayorista(lineas, t.mayorista ?? null));
	const textoMayor = $derived(mayor && textoAviso(mayor, precio));
</script>

{#if desde && lineas.length}
	<div class="tv-aviso-barra" class:tv-aviso-barra--ok={!falta}>
		<p>
			{#if falta}Te faltan <b>{precio(falta)}</b> para el envío&nbsp;gratis.{:else}<b
					>¡Ya tenés envío gratis!</b
				>{/if}
		</p>
		{#if falta}<span class="tv-progreso" aria-hidden="true"
				><span style:width="{Math.min(100, (subtotal / desde) * 100)}%"></span></span
			>{/if}
	</div>
{/if}
{#if mayor && textoMayor}
	<div class="tv-aviso-barra" class:tv-aviso-barra--ok={mayor.aplicado}>
		<p>{textoMayor.pre}<b>{textoMayor.b}</b>{textoMayor.post}</p>
		{#if !mayor.aplicado}<span class="tv-progreso" aria-hidden="true"
				><span style:width="{mayor.progreso}%"></span></span
			>{/if}
	</div>
{/if}
