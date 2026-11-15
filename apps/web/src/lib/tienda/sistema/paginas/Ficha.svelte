<script lang="ts">
	import { goto } from '$app/navigation';
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import CompartirProducto from '../../CompartirProducto.svelte';
	import LanzamientoFicha from '../../LanzamientoFicha.svelte';
	import PreguntasProducto from '../../PreguntasProducto.svelte';
	import { carritoDe } from '../../carrito.svelte.ts';
	import { envia } from '../../envio.ts';
	import { cotizarYa, cpValido, textoEnvios } from '../../envio.svelte.ts';
	import { lineaPorMayor } from '../../mayorista.ts';
	import { Seleccion, tarjetaDe } from '../../seleccion.svelte.ts';
	import type { OpcionEnvio, ProductoPublico, TiendaPublica } from '../../tipos.ts';
	import Bloques from '../bloques/Bloques.svelte';
	import { categoriaHref, numeroWa, productosDe, waHref } from '../datos.ts';
	import {
		conTransferencia,
		cuantos,
		cuotas,
		descuento,
		precio,
		promedio,
		sinViuda
	} from '../formato.ts';
	import BotonComparar from '../marco/BotonComparar.svelte';
	import Tienda from '../marco/Tienda.svelte';
	import { disenoDeTienda } from '../migrar.ts';
	import { panelCarrito } from '../panel.svelte.ts';
	import { funcionesDe, opcionesPaginas, temaDe } from '../paginas.ts';
	import { estiloFinal } from '../presets.ts';
	import Boton from '../primitivas/Boton.svelte';
	import Cantidad from '../primitivas/Cantidad.svelte';
	import Estrellas from '../primitivas/Estrellas.svelte';
	import Foto from '../primitivas/Foto.svelte';
	import Grilla from '../primitivas/Grilla.svelte';
	import TarjetaProducto from '../primitivas/TarjetaProducto.svelte';

	/**
	 * La ficha de un producto (A-, B- y C-ficha; en el celular, A/B/C-celular):
	 * - la galería en grilla, con miniaturas o en carrusel (paginas.producto.variante; en el
	 *   celular, siempre carrusel con puntos);
	 * - talle, color y cualquier opción (seleccion.svelte.ts), con lo que queda en stock;
	 * - el precio con cuotas y con transferencia, y el mayorista;
	 * - el botón principal nunca gris si se puede tocar: si falta elegir, dice qué falta y lo marca;
	 *   en el celular va en la barra fija de abajo;
	 * - WhatsApp, calcular el envío por código postal, compartir, lanzamiento, reserva y encargo
	 *   con seña (LanzamientoFicha);
	 * - descripción, envíos y cambios; reseñas y preguntas del producto; relacionados y los bloques
	 *   que la tienda sumó a la ficha (paginas.producto.bloques).
	 */
	let { tienda, producto }: { tienda: TiendaPublica; producto: ProductoPublico } = $props();

	const diseno = $derived(disenoDeTienda(tienda));
	const estilo = $derived(estiloFinal(tienda.plantilla, diseno.paleta, diseno.estilo));
	const dir = $derived(estilo.direccion);
	const op = $derived(opcionesPaginas(tienda, dir));
	const tema = $derived(temaDe(dir));
	const base = $derived(tienda.base ?? `/t/${tienda.slug}`);
	const carrito = $derived(carritoDe(tienda.slug));
	const sel = $derived(new Seleccion(producto));
	const cobros = $derived(tienda.cobros);

	const verbo = $derived(
		dir === 'mostrador'
			? 'Agregar al carrito'
			: estilo.gastronomia
				? 'Agregar al pedido'
				: 'Agregar a la bolsa'
	);
	const tachado = $derived(
		sel && !sel.desde && sel.anterior && sel.anterior > sel.precio ? sel.anterior : null
	);
	const off = $derived(tachado ? descuento(sel.precio, tachado) : '');
	const enCuotas = $derived(
		cobros.mercadoPago ? cuotas(sel.precio, cobros.cuotas, dir !== 'pedido') : ''
	);
	const transf = $derived(
		cobros.transferencia > 0 && cobros.datosTransferencia
			? precio(conTransferencia(sel.precio, cobros.transferencia))
			: ''
	);
	const porMayor = $derived(lineaPorMayor(sel.mayorista, tienda.mayorista, precio));
	const resenas = $derived(tienda.resenas.filter((r) => r.producto === producto.nombre));
	const prom = $derived(
		resenas.length ? resenas.reduce((s, r) => s + r.estrellas, 0) / resenas.length : null
	);
	const wa = $derived(
		waHref(
			tienda,
			`¡Hola! Quiero consultar por ${producto.nombre}${sel.variante ? ` (${sel.texto})` : ''} (${precio(sel.precio)}).`
		)
	);
	const comparador = $derived(funcionesDe(tienda).comparador);
	const chips = $derived(
		[
			tienda.envios.gratisDesde && sel.precio >= tienda.envios.gratisDesde ? 'Envío gratis' : '',
			producto.etiqueta?.trim() ?? ''
		].filter(Boolean)
	);

	// Cuántos (Mostrador: el «− 1 +» al lado del botón).
	let cantidad = $state(1);
	$effect(() => {
		if (cantidad > Math.max(1, sel.stock)) cantidad = Math.max(1, sel.stock);
	});

	// El botón principal: gris sólo si no se puede tocar (sin stock, próximamente o la vitrina).
	const sePuede = $derived(!tienda.vitrina && !sel.proximo && sel.stock > 0);
	const textoBoton = $derived(
		tienda.vitrina && !sel.falta
			? 'Es una tienda de ejemplo'
			: sel.boton(
					dir === 'vidriera' && !sel.falta && sel.stock > 0
						? `${verbo} · ${precio(sel.precio * cantidad)}`
						: verbo
				)
	);
	function agregar() {
		if (!sel.agregar(carrito, cantidad)) return;
		if (op.carrito.variante === 'panel') return void (panelCarrito.abierto = true);
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- ruta de la tienda con su slug
		goto(`${base}/carrito`);
	}
	function comprarYa() {
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- ruta de la tienda con su slug
		if (sel.agregar(carrito, cantidad)) goto(`${base}/checkout`);
	}

	// Calcular el envío por código postal.
	const calcula = $derived(
		op.producto.calcularEnvio && (envia(tienda.envios) || tienda.envios.retiro)
	);
	let cp = $state('');
	let cotizado = $state<OpcionEnvio[] | null>(null);
	let cotizando = $state(false);
	async function calcular(e: SubmitEvent) {
		e.preventDefault();
		if (!cpValido(cp)) return;
		cotizando = true;
		cotizado = await cotizarYa(cp, sel.precio * cantidad, tienda);
		cotizando = false;
	}

	const detalles = $derived(
		[
			{ t: 'Descripción', d: producto.descripcion?.trim() || '' },
			{ t: 'Envíos', d: textoEnvios(tienda.envios) },
			{
				t: 'Cambios y devoluciones',
				d: `${tienda.politica ? `${tienda.politica} ` : ''}Tenés 10 días desde que lo recibís para arrepentirte de la compra, sin dar explicaciones (botón de arrepentimiento, abajo de todo).`
			}
		]
			.filter((x) => x.d)
			.map((x) => ({ ...x, d: sinViuda(x.d) }))
	);

	// Relacionados: de la misma categoría con stock y, si faltan, los más vendidos.
	const relacionados = $derived.by(() => {
		if (!op.producto.relacionados) return [];
		const tope = dir === 'pedido' ? 6 : 4;
		const misma = tienda.productos.filter(
			(p) =>
				p.id !== producto.id &&
				p.nombre &&
				p.categoria === producto.categoria &&
				!tarjetaDe(p).agotado
		);
		const mas = productosDe(tienda, { fuente: 'vendidos', cantidad: 12 }).filter(
			(p) => p.id !== producto.id && !misma.includes(p) && !tarjetaDe(p).agotado
		);
		return [...misma, ...mas].slice(0, tope);
	});
	const tituloRel = $derived(
		dir === 'mostrador'
			? 'Llevalo con'
			: dir === 'pedido'
				? 'Te puede gustar'
				: 'También te puede gustar'
	);

	// Galería: el índice de la foto a la vista en el carrusel del celular.
	let pista = $state<HTMLElement>();
	let vista = $state(0);
	function alDeslizar() {
		if (!pista) return;
		vista = Math.round(pista.scrollLeft / Math.max(1, pista.clientWidth));
	}
	function irA(i: number) {
		sel.verFoto(i);
		pista?.scrollTo({ left: i * pista.clientWidth, behavior: 'smooth' });
	}
	$effect(() => {
		// La foto sigue a lo elegido (color con foto): en el carrusel, también.
		const i = sel.fotoIdx;
		if (pista && Math.round(pista.scrollLeft / Math.max(1, pista.clientWidth)) !== i)
			pista.scrollTo({ left: i * pista.clientWidth });
	});
	const fotos = $derived(sel.fotos.length ? sel.fotos : [null]);
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
{#snippet chipOff()}
	{#if off}<span class="tv-chip tv-chip--off tv-ficha-off">{off}</span>{/if}
{/snippet}

{#snippet galeria()}
	<div class="tv-ficha-galeria" data-variante={op.producto.variante}>
		<!-- Celular (y la variante carrusel): una pista que se desliza, con puntos. -->
		<div class="tv-ficha-carrusel" class:tv-solo-movil={op.producto.variante !== 'carrusel'}>
			<div class="tv-ficha-pista" bind:this={pista} onscroll={alDeslizar}>
				{#each fotos as f, i (i)}
					<div class="tv-ficha-cuadro">
						<Foto
							src={f}
							alt={i === 0 ? producto.nombre : `${producto.nombre}, foto ${i + 1}`}
							nombre={producto.nombre}
							ancho={1100}
							proporcion="var(--t-foto)"
							prioridad={i === 0}
						/>
					</div>
				{/each}
			</div>
			{@render chipOff()}
			{#if fotos.length > 1}
				<div class="tv-ficha-puntos" role="group" aria-label="Fotos">
					{#each fotos.map((_f, i) => i) as i (i)}<button
							type="button"
							aria-label="Ver la foto {i + 1}"
							aria-pressed={i === vista}
							onclick={() => irA(i)}><span></span></button
						>{/each}
				</div>
			{/if}
		</div>
		{#if op.producto.variante === 'miniaturas'}
			<div class="tv-ficha-mini tv-solo-compu">
				{#if fotos.length > 1}
					<div class="tv-ficha-miniaturas" role="group" aria-label="Fotos">
						{#each fotos as f, i (i)}<button
								type="button"
								aria-label="Ver la foto {i + 1}"
								aria-pressed={i === sel.fotoIdx}
								onclick={() => sel.verFoto(i)}
								><Foto
									src={f}
									alt=""
									nombre={producto.nombre}
									ancho={200}
									proporcion="var(--t-foto)"
								/></button
							>{/each}
					</div>
				{/if}
				<div class="tv-ficha-principal">
					<Foto
						src={sel.foto}
						alt={producto.nombre}
						nombre={producto.nombre}
						ancho={1200}
						proporcion="var(--t-foto)"
						prioridad
					/>
					{@render chipOff()}
				</div>
			</div>
		{:else if op.producto.variante === 'grilla'}
			<div class="tv-ficha-grilla tv-solo-compu" data-n={Math.min(fotos.length, 5)}>
				{#each fotos.slice(0, 5) as f, i (i)}
					<div class="tv-ficha-cuadro">
						<Foto
							src={f}
							alt={i === 0 ? producto.nombre : `${producto.nombre}, foto ${i + 1}`}
							nombre={producto.nombre}
							ancho={i === 0 ? 1200 : 700}
							proporcion="var(--t-foto)"
							prioridad={i === 0}
						/>
						{#if i === 0}{@render chipOff()}{/if}
					</div>
				{/each}
			</div>
		{/if}
	</div>
{/snippet}

{#snippet opciones()}
	{#each sel.opciones as o (o.nombre)}
		{@const muestras = o.valores.every((x) => x.color || x.foto)}
		{@const falta = sel.aviso && !o.elegido}
		<fieldset class="tv-opcion" class:tv-opcion--falta={falta}>
			<legend
				><b>{o.nombre}{o.elegido ? ':' : ''}</b
				>{#if o.elegido}&nbsp;{o.elegido}{:else if falta}&nbsp;· elegí uno{/if}</legend
			>
			<div class="tv-opcion-valores" class:tv-opcion-valores--muestras={muestras}>
				{#each o.valores as x (x.valor)}
					{#if muestras && dir !== 'pedido'}
						<button
							type="button"
							class="tv-muestra"
							aria-label="{o.nombre} {x.valor}{x.disponible ? '' : ', sin stock'}"
							aria-pressed={x.elegido}
							disabled={!x.disponible && !x.elegido}
							onclick={() => sel.elegir(o.nombre, x.valor)}
							>{#if x.color}<span style:background={x.color}></span>{:else}<Foto
									src={x.foto}
									alt=""
									nombre={x.valor}
									ancho={120}
									proporcion="1/1"
								/>{/if}</button
						>
					{:else}
						<button
							type="button"
							class="tv-valor"
							aria-pressed={x.elegido}
							disabled={!x.disponible && !x.elegido}
							onclick={() => sel.elegir(o.nombre, x.valor)}
							>{#if muestras && x.color}<span class="tv-valor-punto" style:background={x.color}
								></span>{/if}{x.valor}{#if !x.disponible}<span class="tv-oculto">, sin stock</span
								>{/if}</button
						>
					{/if}
				{/each}
			</div>
		</fieldset>
	{/each}
{/snippet}

{#snippet botonPrincipal(clase: string)}
	<button
		type="button"
		class="tv-btn tv-btn--primario tv-btn--bloque tv-ficha-agregar {clase}"
		disabled={!sePuede}
		onclick={agregar}>{textoBoton}</button
	>
{/snippet}

<Tienda {tienda} pagina="producto" sinWhatsApp>
	<article class="tv-ficha tv-ancho" data-dir={dir}>
		<nav class="tv-migas tv-solo-compu" aria-label="Estás en">
			<a href={base || '/'}>Inicio</a>{#if producto.categoria}<span aria-hidden="true"
					>{dir === 'mostrador' ? '›' : '/'}</span
				><a href={categoriaHref(tienda, producto.categoria)}>{producto.categoria}</a
				>{/if}{#if dir === 'mostrador'}<span aria-hidden="true">›</span><span
					>{producto.nombre}</span
				>{/if}
		</nav>
		<div class="tv-ficha-cuerpo">
			{@render galeria()}
			<div class="tv-ficha-info" class:tv-tarjeta={dir !== 'vidriera'}>
				{#if dir === 'mostrador' && chips.length}
					<div class="tv-ficha-chips">
						{#each chips as c, i (c)}<span
								class="tv-chip"
								class:tv-chip--ok={i === 0 && c === 'Envío gratis'}>{c}</span
							>{/each}
					</div>
				{/if}
				{#if dir === 'pedido' && (producto.categoria || prom)}
					<p class="tv-chico tv-suave tv-ficha-sobre">
						{producto.categoria ?? ''}{#if producto.categoria && prom}&#32;·&#32;{/if}{#if prom}
							{promedio(prom)}&nbsp;<Icon icon={I.phStarFill} size={12} /> ({resenas.length}){/if}
					</p>
				{/if}
				<h1 class="tv-h1 tv-ficha-titulo">{sinViuda(producto.nombre)}</h1>
				{#if prom && dir !== 'pedido'}
					<a class="tv-ficha-estrellas" href="#resenas"
						><Estrellas valor={prom} tam={14} /><b>{promedio(prom)}</b><span
							>· {cuantos(resenas.length, 'reseña', 'reseñas')}</span
						></a
					>
				{/if}

				<div class="tv-ficha-precio" class:tv-ficha-precio--caja={dir === 'mostrador'}>
					<p class="tv-ficha-monto">
						{#if sel.proximo}<b>Próximamente</b>{:else}<b
								>{sel.desde ? 'Desde ' : ''}{precio(sel.precio)}</b
							>{#if tachado}<s><span class="tv-oculto">Antes </span>{precio(tachado)}</s
								>{/if}{#if off && dir === 'pedido'}<span class="tv-chip tv-chip--oscuro">{off}</span
								>{/if}{/if}
					</p>
					{#if !sel.proximo && (enCuotas || transf)}
						{#if dir === 'mostrador'}
							{#if enCuotas}<p class="tv-ok tv-ficha-cuotas"><b>{enCuotas}</b></p>{/if}
							{#if transf}<p class="tv-ficha-transf">{transf} pagando con transferencia</p>{/if}
						{:else}
							<p class="tv-ficha-cuotas">
								{enCuotas}{#if enCuotas && transf}&#32;·&#32;{/if}{#if transf}<span
										class={dir === 'vidriera' ? 'tv-ok' : ''}>{transf} con transferencia</span
									>{/if}
							</p>
						{/if}
					{/if}
					{#if porMayor}<p class="tv-ficha-mayor">
							<Icon icon={I.phPackage} size={16} />{porMayor}
						</p>{/if}
				</div>

				{@render opciones()}
				{#if sel.stock > 0 && sel.stock <= 3 && !sel.proximo && (!sel.conOpciones || sel.variante)}
					<p class="tv-ficha-quedan">
						{dir === 'mostrador' ? '¡' : ''}{sel.stock === 1
							? 'Queda 1'
							: `Quedan ${sel.stock}`}{sel.variante ? ` en ${sel.texto.toLowerCase()}` : ''}{dir ===
						'mostrador'
							? '!'
							: ''}
					</p>
				{:else if sel.stock === 0 && !sel.proximo && (!sel.conOpciones || sel.variante)}
					<p class="tv-ficha-quedan">
						Sin stock{sel.variante ? ` en ${sel.texto.toLowerCase()}` : ''}{wa
							? '. Escribinos y te avisamos cuando vuelva.'
							: '.'}
					</p>
				{/if}

				<div class="tv-ficha-acciones">
					<div class="tv-ficha-comprar tv-solo-compu">
						{#if dir === 'mostrador' && sePuede && sel.stock > 1}
							<Cantidad
								valor={cantidad}
								max={sel.stock}
								nombre={producto.nombre}
								alCambiar={(n) => (cantidad = Math.max(1, n))}
							/>
						{/if}
						{@render botonPrincipal('')}
					</div>
					{#if dir === 'pedido' && cobros.mercadoPago && sePuede}
						<Boton tipo="suave" bloque onclick={comprarYa}>Comprar ahora con Mercado Pago</Boton>
					{/if}
					{#if wa}
						<Boton href={wa} tipo={dir === 'mostrador' ? 'wa' : 'secundario'} bloque externo
							><Icon icon={I.phWhatsappLogo} size={18} />Consultar por WhatsApp</Boton
						>
					{/if}
					<LanzamientoFicha {tienda} {producto} {sel} {tema} />
					<div class="tv-ficha-extras">
						<CompartirProducto {tienda} {producto} {tema} />
						{#if comparador}<BotonComparar {producto} grande />{/if}
					</div>
				</div>

				{#if calcula}
					<form class="tv-ficha-envio" class:tv-tarjeta={dir === 'vidriera'} onsubmit={calcular}>
						<label for="tv-cp-ficha"
							><b>{dir === 'mostrador' ? '¿Cuándo te llega?' : 'Calculá el envío'}</b></label
						>
						<div class="tv-ficha-envio-fila">
							<input
								id="tv-cp-ficha"
								bind:value={cp}
								inputmode="numeric"
								maxlength="4"
								autocomplete="postal-code"
								placeholder="Tu código postal"
							/>
							<button
								type="submit"
								class="tv-btn tv-btn--secundario tv-btn--chico"
								disabled={cotizando}>{cotizando ? 'Calculando…' : 'Calcular'}</button
							>
						</div>
						{#if cotizado}
							<ul class="tv-ficha-envios" aria-live="polite">
								{#each cotizado as o (o.id)}
									<li>
										<span>{o.nombre} · {o.detalle}</span><b class:tv-ok={!o.precio}
											>{o.precio ? precio(o.precio) : 'Gratis'}</b
										>
									</li>
								{:else}
									<li class="tv-suave">No llegamos a ese código postal por&nbsp;ahora.</li>
								{/each}
							</ul>
						{/if}
					</form>
				{/if}

				{#if dir !== 'mostrador'}
					<div class="tv-ficha-detalles">
						{#if dir === 'pedido'}
							<details>
								<summary>Detalles</summary>
								{#each detalles as d (d.t)}<h2 class="tv-ficha-det-tit">{d.t}</h2>
									<p>{d.d}</p>{/each}
							</details>
						{:else}
							{#each detalles as d, i (d.t)}
								<details open={i === 0 && !!producto.descripcion}>
									<summary>{d.t}</summary>
									<p>{d.d}</p>
								</details>
							{/each}
						{/if}
					</div>
				{/if}
			</div>
		</div>
	</article>

	<!-- Celular: el botón principal fijo abajo (reserva su lugar al final: regla 8). Va acá, antes
	que los relacionados, así es el primero de la página para el teclado y los lectores. -->
	<div class="tv-ficha-barra tv-solo-movil">
		{#if dir === 'mostrador' && numeroWa(tienda)}
			<a
				class="tv-ficha-wa"
				href={wa}
				target="_blank"
				rel="noopener"
				aria-label="Consultar por WhatsApp"><Icon icon={I.phWhatsappLogoFill} size={24} /></a
			>
		{/if}
		{@render botonPrincipal('')}
	</div>

	{#if dir === 'mostrador'}
		<section class="tv-ancho tv-ficha-cajas" aria-label="Detalles">
			{#each detalles as d (d.t)}
				<div class="tv-tarjeta">
					<h2 class="tv-h3">{d.t}</h2>
					<p>{d.d}</p>
				</div>
			{/each}
		</section>
	{/if}

	<section
		class="tv-ancho tv-ficha-opiniones"
		class:tv-ficha-opiniones--sin={!prom}
		id="resenas"
		aria-label="Reseñas y preguntas"
	>
		{#if prom}
			<div class="tv-ficha-op-cab">
				<h2 class="tv-h2">Reseñas</h2>
				<p class="tv-ficha-prom">{promedio(prom)}</p>
				<p class="tv-suave tv-chico">{cuantos(resenas.length, 'reseña', 'reseñas')}</p>
			</div>
		{/if}
		<div class="tv-ficha-op-lista">
			{#each resenas.slice(0, 6) as r, i (i)}
				<figure class="tv-ficha-resena">
					<figcaption>
						<b>{r.autor}</b>{#if r.conCompra}<span class="tv-suave"
								>&#32;·&#32;compra verificada</span
							>{/if}
					</figcaption>
					<Estrellas valor={r.estrellas} tam={14} />
					<blockquote>{r.texto}</blockquote>
					{#if r.respuesta}<p class="tv-suave tv-chico">
							<b>Respuesta de {tienda.nombre}:</b>
							{r.respuesta}
						</p>{/if}
				</figure>
			{/each}
			<PreguntasProducto
				slug={tienda.slug}
				productoId={producto.id}
				preguntas={tienda.preguntas}
				tema={{ ...tema, tarjeta: 'transparent' }}
				titulo="Preguntas"
			/>
		</div>
	</section>

	{#if relacionados.length}
		<section class="tv-ancho tv-ficha-rel" aria-labelledby="tv-rel">
			<h2 class="tv-h2" id="tv-rel">{tituloRel}</h2>
			<Grilla items={relacionados} columnas={dir === 'pedido' ? 6 : 4} etiqueta={tituloRel}>
				{#snippet children(p)}<TarjetaProducto producto={p} ancho={500} />{/snippet}
			</Grilla>
		</section>
	{/if}

	{#if op.producto.bloques.length}
		<div class="tv-ficha-bloques">
			<Bloques bloques={op.producto.bloques} suscripciones={false} />
		</div>
	{/if}

	<div class="tv-reserva-barra tv-solo-movil" aria-hidden="true"></div>
</Tienda>
