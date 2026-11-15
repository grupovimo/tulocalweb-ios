<script lang="ts">
	import { enhance } from '$app/forms';
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import CampoCupon from '../../CampoCupon.svelte';
	import EnvioCorreo from '../../EnvioCorreo.svelte';
	import { carritoDe } from '../../carrito.svelte.ts';
	import { combosDe, descuentoCombos } from '../../combos.ts';
	import { CuponCheckout, totales } from '../../cupon.svelte.ts';
	import { disenoDeTienda } from '../migrar.ts';
	import { cotizar, cpValido } from '../../envio.svelte.ts';
	import { camposFactura, OPCIONALES } from '../../factura.ts';
	import type { OpcionEnvio, TiendaPublica } from '../../tipos.ts';
	import { conTransferencia, cuotas, precio } from '../formato.ts';
	import Tienda from '../marco/Tienda.svelte';
	import Boton from '../primitivas/Boton.svelte';
	import Foto from '../primitivas/Foto.svelte';

	/**
	 * El checkout (A/B/C-celular › pago): 1 Datos, 2 Entrega y 3 Pago en una sola página, con el paso
	 * en el que se está marcado arriba; al costado (abajo en el celular), el pedido con el cupón y la
	 * cuenta, igual que la calcula el servidor (cupon.svelte.ts › totales). Factura A si la tienda la
	 * emite, «Quiero recibir novedades», Correo Argentino con provincia y sucursal (EnvioCorreo). El
	 * botón de pagar no se apaga por faltar algo: dice qué falta. Mercado Pago es otro sitio: se
	 * navega de verdad.
	 */
	let { tienda, error = null }: { tienda: TiendaPublica; error?: string | null } = $props();

	const base = $derived(tienda.base ?? `/t/${tienda.slug}`);
	const carrito = $derived(carritoDe(tienda.slug));
	const lineas = $derived(carrito.detalle(tienda.productos, tienda.mayorista ?? null));
	const subtotal = $derived(lineas.reduce((s, l) => s + l.subtotal, 0));

	let email = $state('');
	let nombre = $state('');
	let telefono = $state('');
	let cp = $state('');
	let direccion = $state('');
	let envio = $state<OpcionEnvio['id'] | null>(null);
	let pago = $state<'transferencia' | 'mercadopago'>('transferencia');
	let enviando = $state(false);
	let faltaEntrega = $state(false);

	const opciones = $derived(
		cpValido(cp)
			? cotizar(cp, subtotal, tienda)
			: tienda.envios.retiro
				? cotizar('0000', subtotal, tienda)
				: []
	);
	const elegida = $derived(opciones.find((o) => o.id === envio) ?? null);
	$effect(() => {
		if (envio && !opciones.some((o) => o.id === envio)) envio = null;
	});
	$effect(() => {
		if (envio) faltaEntrega = false;
	});

	const cupon = new CuponCheckout(() => tienda.slug);
	const cuenta = $derived(
		totales({
			subtotal,
			envio: elegida?.precio ?? 0,
			pago,
			transferencia: tienda.cobros.transferencia,
			cupon,
			combo: descuentoCombos(
				combosDe(disenoDeTienda(tienda)),
				lineas.map((l) => ({ ...l, precio: l.producto.precio }))
			).total
		})
	);

	const pagos = $derived([
		...(tienda.cobros.mercadoPago
			? [
					{
						id: 'mercadopago' as const,
						t: 'Mercado Pago',
						d:
							tienda.cobros.cuotas > 1
								? `Tarjetas, hasta ${tienda.cobros.cuotas} cuotas sin interés, o dinero en cuenta`
								: 'Tarjetas de crédito y débito, o dinero en cuenta'
					}
				]
			: []),
		...(tienda.cobros.datosTransferencia
			? [
					{
						id: 'transferencia' as const,
						t: tienda.cobros.transferencia
							? `Transferencia · ${tienda.cobros.transferencia} % off`
							: 'Transferencia',
						d: 'Te mostramos el alias al confirmar'
					}
				]
			: [])
	]);
	$effect(() => {
		if (pagos.length && !pagos.some((x) => x.id === pago)) pago = pagos[0].id;
	});

	const CAMPOS = [
		['email', 'Email', 'email', 'email'],
		['nombre', 'Nombre y apellido', 'text', 'name'],
		['telefono', 'Teléfono', 'tel', 'tel'],
		['dni', 'DNI (opcional)', 'text', 'off']
	] as const;
	const factura = $derived(camposFactura(tienda));

	// En qué paso va (los de arriba): datos completos, entrega elegida, pago.
	const datosOk = $derived(
		/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim()) &&
			nombre.trim().length >= 2 &&
			/^[\d\s+()-]{8,20}$/.test(telefono.trim())
	);
	const entregaOk = $derived(
		!!envio && (envio === 'retiro' || (cpValido(cp) && direccion.trim().length >= 4))
	);
	const paso = $derived(!datosOk ? 0 : !entregaOk ? 1 : 2);
	const PASOS = ['Datos', 'Entrega', 'Pago'];

	const textoPagar = $derived(
		tienda.vitrina
			? 'Es una tienda de ejemplo'
			: !pagos.length
				? 'Esta tienda todavía no tiene medios de pago'
				: `Pagar ${precio(cuenta.total)}`
	);
	const extra = $derived(
		pago === 'mercadopago' && tienda.cobros.mercadoPago
			? cuotas(cuenta.total, tienda.cobros.cuotas, true)
			: ''
	);
	let entrega = $state<HTMLElement>();
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
<Tienda {tienda} pagina="checkout" sinWhatsApp>
	<div class="tv-ancho tv-pag tv-checkout">
		{#if !lineas.length}
			<div class="tv-carrito-vacio tv-tarjeta">
				<Icon icon={I.phHandbag} size={28} />
				<p class="tv-lead">Tu carrito está vacío.</p>
				<Boton href="{base}/catalogo">Ver el catálogo</Boton>
			</div>
		{:else}
			<ol class="tv-pasos" aria-label="Pasos de la compra">
				{#each PASOS as p, i (p)}
					<li
						class:tv-paso--hecho={i < paso}
						class:tv-paso--ahora={i === paso}
						aria-current={i === paso ? 'step' : undefined}
					>
						<span>{i + 1}</span>{p}{#if i < paso}<Icon icon={I.phCheckBold} size={12} />{/if}
					</li>
				{/each}
			</ol>
			<form
				class="tv-checkout-cuerpo"
				method="POST"
				use:enhance={({ cancel }) => {
					if (!envio) {
						cancel();
						faltaEntrega = true;
						entrega?.scrollIntoView({ behavior: 'smooth', block: 'center' });
						return;
					}
					enviando = true;
					return async ({ result, update }) => {
						enviando = false;
						// Mercado Pago es otro sitio: se navega de verdad.
						if (result.type === 'redirect' && result.location.startsWith('https://'))
							return void (location.href = result.location);
						if (result.type === 'redirect') carrito.vaciar();
						await update();
					};
				}}
			>
				<input type="hidden" name="carrito" value={JSON.stringify(carrito.lineas)} />
				<div class="tv-checkout-pasos">
					<section class="tv-checkout-paso" aria-labelledby="tv-co-datos">
						<h2 class="tv-h3" id="tv-co-datos">Tus datos</h2>
						<div class="tv-campos">
							{#each CAMPOS as [n, l, tipo, ac] (n)}
								<div class="tv-campo">
									<label for="tv-co-{n}">{l}</label>
									{#if n === 'email'}<input
											id="tv-co-{n}"
											name={n}
											type={tipo}
											autocomplete={ac}
											required
											maxlength="120"
											bind:value={email}
										/>
									{:else if n === 'nombre'}<input
											id="tv-co-{n}"
											name={n}
											type={tipo}
											autocomplete={ac}
											required
											maxlength="80"
											bind:value={nombre}
										/>
									{:else if n === 'telefono'}<input
											id="tv-co-{n}"
											name={n}
											type={tipo}
											autocomplete={ac}
											required
											maxlength="20"
											bind:value={telefono}
										/>
									{:else}<input
											id="tv-co-{n}"
											name={n}
											type={tipo}
											autocomplete={ac}
											maxlength="12"
											inputmode="numeric"
										/>{/if}
								</div>
							{/each}
							{#each factura as [n, l, tipo, ac] (n)}
								<div class="tv-campo">
									<label for="tv-co-{n}">{l}</label>
									<input
										id="tv-co-{n}"
										name={n}
										type={tipo}
										autocomplete={ac}
										required={!OPCIONALES.has(n)}
										maxlength="120"
									/>
								</div>
							{/each}
						</div>
						<label class="tv-casilla tv-checkout-nl"
							><input type="checkbox" name="novedades" /><span>Quiero recibir novedades</span
							></label
						>
					</section>

					<section class="tv-checkout-paso" aria-labelledby="tv-co-entrega" bind:this={entrega}>
						<h2 class="tv-h3" id="tv-co-entrega">¿Cómo lo recibís?</h2>
						<div class="tv-campos tv-campos--envio">
							<div class="tv-campo">
								<label for="tv-co-cp">Código postal</label>
								<input
									id="tv-co-cp"
									name="cp"
									bind:value={cp}
									inputmode="numeric"
									maxlength="4"
									autocomplete="postal-code"
								/>
							</div>
							<div class="tv-campo">
								<label for="tv-co-dir">Dirección (calle, número, piso)</label>
								<input
									id="tv-co-dir"
									name="direccion"
									bind:value={direccion}
									maxlength="160"
									autocomplete="street-address"
									required={envio !== null && envio !== 'retiro'}
								/>
							</div>
						</div>
						<div class="tv-radios" role="radiogroup" aria-label="Forma de entrega">
							{#each opciones as s (s.id)}
								<label class="tv-radio tv-tarjeta" class:tv-radio--on={envio === s.id}>
									<input type="radio" name="envio" value={s.id} bind:group={envio} />
									<span class="tv-radio-txt"
										><b>{s.nombre}</b><span class="tv-chico tv-suave">{s.detalle}</span></span
									>
									<b class="tv-radio-precio" class:tv-ok={!s.precio}
										>{s.precio ? precio(s.precio) : 'Gratis'}</b
									>
								</label>
							{:else}
								<p class="tv-suave tv-chico">Poné tu código postal para ver cómo te&nbsp;llega.</p>
							{/each}
						</div>
						{#if faltaEntrega}<p class="tv-error" role="alert">
								Elegí cómo te llega para seguir.
							</p>{/if}
						<EnvioCorreo {tienda} {cp} {envio} {opciones} />
					</section>

					<section class="tv-checkout-paso" aria-labelledby="tv-co-pago">
						<h2 class="tv-h3" id="tv-co-pago">¿Cómo pagás?</h2>
						<div class="tv-radios" role="radiogroup" aria-label="Forma de pago">
							{#each pagos as s (s.id)}
								<label class="tv-radio tv-tarjeta" class:tv-radio--on={pago === s.id}>
									<input type="radio" name="pago" value={s.id} bind:group={pago} />
									<span class="tv-radio-txt"
										><b>{s.t}</b><span class="tv-chico tv-suave">{s.d}</span></span
									>
								</label>
							{:else}
								<p class="tv-suave tv-chico">
									Esta tienda todavía no cargó cómo cobrar. Escribile para coordinar.
								</p>
							{/each}
						</div>
					</section>

					{#if error}<p class="tv-error tv-error--caja" role="alert">{error}</p>{/if}
					<div class="tv-solo-compu">
						<button
							class="tv-btn tv-btn--primario tv-btn--bloque tv-checkout-pagar"
							disabled={!pagos.length || enviando || tienda.vitrina}
							><Icon icon={I.phLockSimpleFill} size={16} />{enviando
								? 'Un momento…'
								: textoPagar}</button
						>
					</div>
				</div>

				<aside class="tv-checkout-resumen tv-tarjeta" aria-label="Tu pedido">
					<div class="tv-checkout-resumen-cab">
						<h2 class="tv-h3">Tu pedido</h2>
						<a class="tv-link" href="{base}/carrito">Cambiar</a>
					</div>
					<ul class="tv-checkout-lineas">
						{#each lineas as l (l.clave)}
							<li>
								<Foto
									src={l.producto.foto}
									alt=""
									nombre={l.producto.nombre}
									ancho={160}
									proporcion="1/1"
								/>
								<span
									>{l.producto.nombre}<small class="tv-suave"
										>{l.variante ? `${l.variante} · ` : ''}× {l.cantidad}</small
									></span
								>
								<span class="tv-checkout-sub">{precio(l.subtotal)}</span>
							</li>
						{/each}
					</ul>
					<CampoCupon
						{cupon}
						lineas={carrito.lineas}
						{email}
						tinta="var(--t-texto)"
						suave="var(--t-suave)"
						borde="var(--t-borde)"
						fondo="var(--t-fondo)"
						acento="var(--t-texto)"
						sobreAcento="var(--t-fondo)"
						radio="var(--t-r-boton)"
						fuente="var(--t-f-texto)"
					/>
					<dl class="tv-cuentas">
						<div>
							<dt>Subtotal</dt>
							<dd>{precio(subtotal)}</dd>
						</div>
						<div>
							<dt>Envío{elegida ? ` · ${elegida.nombre}` : ''}</dt>
							<dd>{elegida ? (cuenta.envio ? precio(cuenta.envio) : 'Gratis') : '—'}</dd>
						</div>
						{#if cuenta.descCombo}<div class="tv-ok">
								<dt>Combo</dt>
								<dd>− {precio(cuenta.descCombo)}</dd>
							</div>{/if}
						{#if cuenta.descCupon}<div class="tv-ok">
								<dt>Cupón {cupon.aplicado}</dt>
								<dd>− {precio(cuenta.descCupon)}</dd>
							</div>{/if}
						{#if cuenta.transf}<div class="tv-ok">
								<dt>Transferencia −{tienda.cobros.transferencia} %</dt>
								<dd>− {precio(cuenta.transf)}</dd>
							</div>{/if}
					</dl>
					<div class="tv-cuentas-total"><span>Total</span><b>{precio(cuenta.total)}</b></div>
					{#if extra}<p class="tv-chico tv-suave">o {extra}</p>{/if}
					{#if !cuenta.transf && tienda.cobros.transferencia > 0 && tienda.cobros.datosTransferencia && pago !== 'transferencia'}<p
							class="tv-chico tv-suave"
						>
							{precio(
								conTransferencia(
									subtotal - cuenta.descCombo - cuenta.descCupon,
									tienda.cobros.transferencia
								) + cuenta.envio
							)} pagando con transferencia
						</p>{/if}
					<p class="tv-chico tv-suave tv-carrito-seguro">
						<Icon icon={I.phShieldCheck} size={16} />Compra protegida · 10 días
						para&nbsp;arrepentirte
					</p>
				</aside>

				<div class="tv-reserva-barra tv-solo-movil" aria-hidden="true"></div>
				<div class="tv-ficha-barra tv-solo-movil">
					<button
						class="tv-btn tv-btn--primario tv-btn--bloque"
						disabled={!pagos.length || enviando || tienda.vitrina}
						>{enviando ? 'Un momento…' : textoPagar}</button
					>
				</div>
			</form>
		{/if}
	</div>
</Tienda>
