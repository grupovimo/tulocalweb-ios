<script lang="ts">
	import { proximamente } from '../../lanzamiento.ts';
	import { tarjetaDe } from '../../seleccion.svelte.ts';
	import type { ProductoPublico } from '../../tipos.ts';
	import { usarTienda } from '../contexto.ts';
	import { conTransferencia, cuotas, precio } from '../formato.ts';

	/**
	 * El precio de un producto como lo muestra cada dirección (README, fila «Precio»):
	 * - Vidriera: 600, tachado al lado y las cuotas abajo en 13 px.
	 * - Mostrador: 700 grande, cuotas sin interés en verde y el precio con transferencia.
	 * - Pedido: 600, todo en una línea (`linea`).
	 * Con variantes de precios distintos, «Desde». Un lanzamiento muestra «Próximamente».
	 */
	let {
		producto,
		tam = 'normal',
		linea = false,
		extras = true
	}: {
		producto: ProductoPublico;
		tam?: 'normal' | 'grande';
		/** Sólo el precio, sin cuotas ni transferencia. */
		linea?: boolean;
		extras?: boolean;
	} = $props();

	const ctx = usarTienda();
	const t = $derived(tarjetaDe(producto));
	const cobros = $derived(ctx.tienda.cobros);
	const anterior = $derived(
		!t.desde && producto.anterior && producto.anterior > t.precio ? producto.anterior : null
	);
	const enCuotas = $derived(
		cobros.mercadoPago ? cuotas(t.precio, cobros.cuotas, ctx.dir === 'mostrador') : ''
	);
	const transf = $derived(
		cobros.transferencia > 0 && cobros.datosTransferencia
			? `${precio(conTransferencia(t.precio, cobros.transferencia))} con transferencia`
			: ''
	);
	const ver = $derived(extras && !linea);
</script>

<span class="tv-precio tv-precio--{tam}">
	{#if proximamente(producto)}
		<span class="tv-precio-monto">Próximamente</span>
	{:else}
		<span class="tv-precio-fila"
			><b class="tv-precio-monto">{t.desde ? 'Desde ' : ''}{precio(t.precio)}</b>{#if anterior}<s
					class="tv-precio-antes"><span class="tv-oculto">Antes </span>{precio(anterior)}</s
				>{/if}</span
		>
		{#if ver && ctx.dir === 'mostrador'}
			{#if enCuotas}<span class="tv-precio-ok">{enCuotas}</span>{/if}
			{#if transf}<span class="tv-precio-ok">{transf}</span>{/if}
		{:else if ver && ctx.dir === 'vidriera' && enCuotas}
			<span class="tv-precio-cuotas">{enCuotas}</span>
		{/if}
	{/if}
</span>
