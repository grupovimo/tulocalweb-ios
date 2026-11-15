<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { usarTienda } from '../contexto.ts';
	import { waHref } from '../datos.ts';
	import { conTransferencia, cuotas, cuantos, precio } from '../formato.ts';
	import Boton from '../primitivas/Boton.svelte';
	import Foto from '../primitivas/Foto.svelte';

	/**
	 * La bolsa (o el pedido) siempre visible al costado en Pedido (C-moda, C-cafe, C-vacia): las
	 * líneas del carrito con su foto, envío gratis si llega, total con cuotas y transferencia, y los
	 * botones para comprar o consultar. Vacía, lo dice. Sólo en la compu (en el celular va
	 * marco/BarraPedido).
	 *
	 * Uso: `<Bolsa />` en la columna derecha de una página de Pedido (paginas/Inicio.svelte).
	 */
	const ctx = usarTienda();
	const t = $derived(ctx.tienda);
	const lineas = $derived(ctx.carrito.detalle(t.productos, t.mayorista ?? null));
	const total = $derived(lineas.reduce((s, l) => s + l.subtotal, 0));
	const unidades = $derived(lineas.reduce((s, l) => s + l.cantidad, 0));
	const titulo = $derived(ctx.estilo.gastronomia ? 'Tu pedido' : 'Tu bolsa');
	const gratis = $derived(t.envios.gratisDesde && total >= t.envios.gratisDesde);
	const falta = $derived(t.envios.gratisDesde ? Math.max(0, t.envios.gratisDesde - total) : 0);
	const extra = $derived(
		[
			t.cobros.mercadoPago ? cuotas(total, t.cobros.cuotas) : '',
			t.cobros.transferencia > 0 && t.cobros.datosTransferencia
				? `${precio(conTransferencia(total, t.cobros.transferencia))} con transferencia`
				: ''
		]
			.filter(Boolean)
			.join(' · ')
	);
	const wa = $derived(waHref(t, '¡Hola! Quería consultar por mi pedido.'));
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
<aside class="tv-bolsa-lateral tv-tarjeta" aria-label={titulo}>
	<div class="tv-bolsa-cab">
		<h2 class="tv-h3">{titulo}</h2>
		{#if unidades}<span class="tv-chico tv-suave">{cuantos(unidades, 'producto', 'productos')}</span
			>{/if}
	</div>
	{#if lineas.length}
		<ul class="tv-bolsa-lineas">
			{#each lineas as l (l.clave)}
				<li>
					<Foto
						src={l.producto.foto}
						alt=""
						nombre={l.producto.nombre}
						ancho={120}
						proporcion="4/5"
					/>
					<div>
						<p class="tv-bolsa-nombre">{l.producto.nombre}</p>
						<p class="tv-chico tv-suave">
							{l.variante ? `${l.variante} · ` : ''}{l.cantidad}&nbsp;×&nbsp;{precio(
								l.producto.precio
							)}
						</p>
					</div>
					<b>{precio(l.subtotal)}</b>
					<button
						type="button"
						class="tv-bolsa-quitar"
						aria-label="Quitar uno de {l.producto.nombre}"
						onclick={() => ctx.carrito.fijar(l.clave, l.cantidad - 1)}
						><Icon icon={I.phMinus} size={14} /></button
					>
				</li>
			{/each}
		</ul>
		{#if t.envios.gratisDesde && ctx.conversion.envioGratis}
			<div class="tv-aviso-barra tv-bolsa-aviso" class:tv-aviso-barra--ok={!falta}>
				<p>
					{#if falta}Te faltan <b>{precio(falta)}</b> para el envío&nbsp;gratis.{:else}<b
							>¡Ya tenés envío gratis!</b
						>{/if}
				</p>
				{#if falta}<span class="tv-progreso" aria-hidden="true"
						><span style:width="{Math.min(100, (total / t.envios.gratisDesde) * 100)}%"
						></span></span
					>{/if}
			</div>
		{/if}
		<div class="tv-bolsa-total">
			{#if t.envios.gratisDesde}<div class="tv-bolsa-fila">
					<span>Envío</span>{#if gratis}<b class="tv-ok">Gratis</b>{:else}<span class="tv-suave"
							>Gratis desde {precio(t.envios.gratisDesde)}</span
						>{/if}
				</div>{/if}
			<div class="tv-bolsa-fila tv-bolsa-fila--total"><span>Total</span><b>{precio(total)}</b></div>
			{#if extra}<span class="tv-chico tv-suave">{extra}</span>{/if}
		</div>
		<Boton href="{ctx.base}/carrito" bloque
			>{ctx.estilo.gastronomia ? 'Confirmar pedido' : 'Comprar'}</Boton
		>
		{#if wa}<Boton href={wa} tipo="suave" bloque externo>Consultar por WhatsApp</Boton>{/if}
	{:else}
		<p class="tv-bolsa-vacia">Todavía está vacía.</p>
	{/if}
</aside>
