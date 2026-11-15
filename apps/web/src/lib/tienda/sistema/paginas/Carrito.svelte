<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import CampoCupon from '../../CampoCupon.svelte';
	import { carritoDe } from '../../carrito.svelte.ts';
	import { combosDe, descuentoCombos } from '../../combos.ts';
	import { CuponCheckout } from '../../cupon.svelte.ts';
	import { cotizar, cpValido } from '../../envio.svelte.ts';
	import { envia } from '../../envio.ts';
	import { tarjetaDe } from '../../seleccion.svelte.ts';
	import type { TiendaPublica } from '../../tipos.ts';
	import { productosDe } from '../datos.ts';
	import { conTransferencia, cuantos, cuotas, precio } from '../formato.ts';
	import AvisosCarrito from '../marco/AvisosCarrito.svelte';
	import LineasCarrito from '../marco/LineasCarrito.svelte';
	import Tienda from '../marco/Tienda.svelte';
	import { disenoDeTienda } from '../migrar.ts';
	import { opcionesPaginas } from '../paginas.ts';
	import { estiloFinal } from '../presets.ts';
	import Boton from '../primitivas/Boton.svelte';
	import Grilla from '../primitivas/Grilla.svelte';
	import TarjetaProducto from '../primitivas/TarjetaProducto.svelte';

	/**
	 * El carrito (A/B/C-celular: bolsa, carrito y pedido): cuánto falta para el envío gratis y para el
	 * precio mayorista, las líneas con cantidades y lo que se agotó en el momento, el cupón (queda
	 * anotado para el checkout), el envío estimado por código postal y el total con cuotas y con
	 * transferencia. En el celular, el total y «Finalizar compra» van fijos abajo. Vacío, lo dice y
	 * ofrece el catálogo; abajo, «Te puede gustar».
	 */
	let { tienda }: { tienda: TiendaPublica } = $props();

	const diseno = $derived(disenoDeTienda(tienda));
	const estilo = $derived(estiloFinal(tienda.plantilla, diseno.paleta, diseno.estilo));
	const dir = $derived(estilo.direccion);
	const op = $derived(opcionesPaginas(tienda, dir).carrito);
	const base = $derived(tienda.base ?? `/t/${tienda.slug}`);
	const carrito = $derived(carritoDe(tienda.slug));
	const lineas = $derived(carrito.detalle(tienda.productos, tienda.mayorista ?? null));
	const unidades = $derived(lineas.reduce((s, l) => s + l.cantidad, 0));
	const subtotal = $derived(lineas.reduce((s, l) => s + l.subtotal, 0));
	const titulo = $derived(
		estilo.gastronomia ? 'Tu pedido' : dir === 'mostrador' ? 'Tu carrito' : 'Tu bolsa'
	);

	// El cupón: se revisa acá y queda anotado (como el link de un cupón) para el checkout.
	const cupon = new CuponCheckout(() => tienda.slug);
	$effect(() => {
		if (!cupon.aplicado) return;
		document.cookie = `cupon-${tienda.slug}=${cupon.aplicado}; path=/t/${tienda.slug}; max-age=${7 * 86_400}; samesite=lax`;
	});
	const descCupon = $derived(cupon.ok?.descuento ?? 0);
	// Los combos completos (agregados juntos desde su bloque), como los cobra el servidor.
	const descCombo = $derived(
		descuentoCombos(
			combosDe(diseno),
			lineas.map((l) => ({ ...l, precio: l.producto.precio }))
		).total
	);
	const total = $derived(Math.max(0, subtotal - descCombo - descCupon));

	// El envío estimado: el más barato a domicilio para ese código postal.
	let cp = $state('');
	let cpPedido = $state('');
	const envio = $derived.by(() => {
		if (!cpValido(cpPedido)) return null;
		const o = cotizar(cpPedido, subtotal, tienda).filter((x) => x.id !== 'retiro');
		return o.length ? Math.min(...o.map((x) => x.precio)) : -1;
	});
	const extra = $derived(
		[
			tienda.cobros.mercadoPago ? cuotas(total, tienda.cobros.cuotas, dir === 'mostrador') : '',
			tienda.cobros.transferencia > 0 && tienda.cobros.datosTransferencia
				? `${precio(conTransferencia(total, tienda.cobros.transferencia))} con transferencia`
				: ''
		].filter(Boolean)
	);

	const sugeridos = $derived(
		productosDe(tienda, { fuente: 'vendidos', cantidad: 12 })
			.filter((p) => !carrito.lineas.some((l) => l.id === p.id) && !tarjetaDe(p).agotado)
			.slice(0, 4)
	);
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
<Tienda {tienda} pagina="carrito">
	<div class="tv-ancho tv-pag tv-carrito">
		<div class="tv-carrito-cab">
			<h1 class="tv-h1 tv-pag-titulo">{titulo}</h1>
			{#if unidades}<p class="tv-suave">{cuantos(unidades, 'producto', 'productos')}</p>{/if}
		</div>
		{#if !lineas.length && !carrito.lineas.length}
			<div class="tv-carrito-vacio tv-tarjeta">
				<Icon icon={I.phHandbag} size={28} />
				<p class="tv-lead">Todavía no sumaste nada.</p>
				<Boton href="{base}/catalogo">Ver el catálogo</Boton>
			</div>
		{:else}
			<div class="tv-carrito-cuerpo">
				<div class="tv-carrito-lineas">
					<AvisosCarrito {lineas} envioGratis={op.envioGratisDesde} />
					<LineasCarrito {lineas} />
				</div>
				<aside class="tv-carrito-resumen tv-tarjeta" aria-label="Resumen">
					<h2 class="tv-h3">Resumen</h2>
					<div class="tv-carrito-cupon">
						<CampoCupon
							{cupon}
							lineas={carrito.lineas}
							tinta="var(--t-texto)"
							suave="var(--t-suave)"
							borde="var(--t-borde)"
							fondo="var(--t-fondo)"
							acento="var(--t-texto)"
							sobreAcento="var(--t-fondo)"
							radio="var(--t-r-boton)"
							fuente="var(--t-f-texto)"
						/>
					</div>
					<dl class="tv-cuentas">
						<div>
							<dt>Subtotal</dt>
							<dd>{precio(subtotal)}</dd>
						</div>
						{#if descCombo}<div class="tv-ok">
								<dt>Combo</dt>
								<dd>− {precio(descCombo)}</dd>
							</div>{/if}
						{#if descCupon}<div class="tv-ok">
								<dt>Cupón {cupon.aplicado}</dt>
								<dd>− {precio(descCupon)}</dd>
							</div>{/if}
						{#if envia(tienda.envios)}
							<div>
								<dt>Envío</dt>
								<dd class:tv-suave={envio === null}>
									{envio === null
										? 'Se calcula al pagar'
										: envio < 0
											? 'No llegamos a ese código'
											: envio === 0
												? 'Gratis'
												: `Desde ${precio(envio)}`}
								</dd>
							</div>
						{/if}
					</dl>
					{#if envia(tienda.envios)}
						<form
							class="tv-carrito-cp"
							onsubmit={(e) => {
								e.preventDefault();
								cpPedido = cp;
							}}
						>
							<label for="tv-cp-carrito" class="tv-oculto">Código postal</label>
							<input
								id="tv-cp-carrito"
								bind:value={cp}
								inputmode="numeric"
								maxlength="4"
								autocomplete="postal-code"
								placeholder="Tu código postal"
							/>
							<button type="submit" class="tv-btn tv-btn--secundario tv-btn--chico">Calcular</button
							>
						</form>
					{/if}
					<div class="tv-cuentas-total"><span>Total</span><b>{precio(total)}</b></div>
					{#if extra.length}<p class="tv-chico tv-suave tv-carrito-extra">
							o {extra.join(' · ')}
						</p>{/if}
					<div class="tv-solo-compu">
						<Boton href="{base}/checkout" bloque>Finalizar compra</Boton>
					</div>
					<a class="tv-link tv-carrito-seguir" href="{base}/catalogo">Seguir comprando</a>
					<p class="tv-chico tv-suave tv-carrito-seguro">
						<Icon icon={I.phShieldCheck} size={16} />Compra protegida · 10 días
						para&nbsp;arrepentirte
					</p>
				</aside>
			</div>
		{/if}
	</div>

	{#if sugeridos.length && lineas.length}
		<section class="tv-ancho tv-ficha-rel" aria-labelledby="tv-sug">
			<h2 class="tv-h2" id="tv-sug">Te puede gustar</h2>
			<Grilla items={sugeridos} columnas={4} etiqueta="Te puede gustar">
				{#snippet children(p)}<TarjetaProducto producto={p} ancho={500} />{/snippet}
			</Grilla>
		</section>
	{/if}

	{#if lineas.length}
		<div class="tv-reserva-barra tv-solo-movil" aria-hidden="true"></div>
		<div class="tv-ficha-barra tv-carrito-barra tv-solo-movil">
			<div class="tv-carrito-barra-total">
				<span>Total</span><b>{precio(total)}</b>
			</div>
			<Boton href="{base}/checkout" bloque>Finalizar compra</Boton>
		</div>
	{/if}
</Tienda>
