<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import VacioTienda from '../../VacioTienda.svelte';
	import { textoMinimo } from '../../mayorista.ts';
	import type { TiendaPublica } from '../../tipos.ts';
	import { tarjetaDe } from '../../seleccion.svelte.ts';
	import { waHref } from '../datos.ts';
	import { cuantos, precio } from '../formato.ts';
	import {
		enOferta,
		filtrar,
		hayFiltros,
		hrefFiltros,
		leerFiltros,
		ORDENES,
		porCategoria,
		SIN_FILTROS,
		type Filtros
	} from '../listado.ts';
	import BarraPedido from '../marco/BarraPedido.svelte';
	import Bolsa from '../marco/Bolsa.svelte';
	import BotonComparar from '../marco/BotonComparar.svelte';
	import Tienda from '../marco/Tienda.svelte';
	import { disenoDeTienda } from '../migrar.ts';
	import { funcionesDe, opcionesPaginas, temaDe } from '../paginas.ts';
	import { estiloFinal } from '../presets.ts';
	import Boton from '../primitivas/Boton.svelte';
	import Grilla from '../primitivas/Grilla.svelte';
	import TarjetaProducto from '../primitivas/TarjetaProducto.svelte';

	/**
	 * El catálogo (A-, B- y C-catalogo): categorías, búsqueda, orden, ofertas, con stock, por mayor y
	 * precio, todo en la URL (listado.ts). Los filtros van al costado (casillas, A) o arriba en chips
	 * (B y C), según paginas.catalogo.filtros; en el celular siempre en chips. Grilla con las columnas
	 * elegidas o lista (C), de a 24 con «Ver más». Sin productos, el vacío de la tienda; sin
	 * resultados, qué hacer (limpiar los filtros o preguntar por WhatsApp). En Pedido, la bolsa al
	 * costado y la barra fija en el celular.
	 */
	let { tienda }: { tienda: TiendaPublica } = $props();

	const diseno = $derived(disenoDeTienda(tienda));
	const dir = $derived(estiloFinal(tienda.plantilla, diseno.paleta, diseno.estilo).direccion);
	const op = $derived(opcionesPaginas(tienda, dir).catalogo);
	const base = $derived(tienda.base ?? `/t/${tienda.slug}`);
	const f = $derived(leerFiltros(page.url.searchParams, tienda));
	const lista = $derived(filtrar(tienda, f));
	const cuenta = $derived(porCategoria(tienda, f));
	const cats = $derived(tienda.categorias.filter((c) => c.cantidad > 0));
	const comparador = $derived(funcionesDe(tienda).comparador);
	const wa = $derived(waHref(tienda, '¡Hola! Estoy buscando algo que no encontré en la tienda.'));

	// Qué filtros tienen sentido con lo que hay (nada que no cambie el resultado).
	const hayOfertas = $derived(tienda.productos.some(enOferta));
	const hayAgotados = $derived(tienda.productos.some((p) => tarjetaDe(p).agotado));
	const hayMayor = $derived(
		!!tienda.mayorista && tienda.productos.some((p) => p.mayorista && p.mayorista < p.precio)
	);
	const precios = $derived(tienda.productos.map((p) => tarjetaDe(p).precio));
	const conPrecio = $derived(precios.length > 3 && Math.max(...precios) > Math.min(...precios));

	let mostrar = $state(24);
	$effect(() => {
		void page.url.search;
		mostrar = 24;
	});
	const visibles = $derived(lista.slice(0, mostrar));

	const ir = (g: Partial<Filtros>) => hrefFiltros(base, { ...f, ...g });
	const navegar = (href: string) =>
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- ruta de la tienda con su slug
		goto(href, { replace: true, reset: false });
	const conCat = (c: string) =>
		ir({ cats: f.cats.includes(c) ? f.cats.filter((x) => x !== c) : [...f.cats, c] });

	const titulo = $derived(
		f.q
			? `${cuantos(lista.length, dir === 'mostrador' ? 'producto' : 'resultado', dir === 'mostrador' ? 'productos' : 'resultados')} para «${f.q}»`
			: f.cats.length === 1
				? f.cats[0]
				: f.ofertas && !f.cats.length
					? 'Ofertas'
					: 'Todos los productos'
	);
	const miga = $derived(f.q ? 'Buscar' : f.cats.length === 1 ? f.cats[0] : 'Productos');

	type Chip = { t: string; on: boolean; href: string };
	const activos = $derived<Chip[]>([
		...f.cats.map((c) => ({ t: c, on: true, href: conCat(c) })),
		...(f.ofertas ? [{ t: 'En oferta', on: true, href: ir({ ofertas: false }) }] : []),
		...(f.stock ? [{ t: 'Con stock', on: true, href: ir({ stock: false }) }] : []),
		...(f.mayor ? [{ t: 'Por mayor', on: true, href: ir({ mayor: false }) }] : []),
		...(f.min || f.max
			? [
					{
						t:
							f.min && f.max
								? `${precio(f.min)} a ${precio(f.max)}`
								: f.min
									? `Desde ${precio(f.min)}`
									: `Hasta ${precio(f.max!)}`,
						on: true,
						href: ir({ min: null, max: null })
					}
				]
			: [])
	]);
	const opcionales = $derived<Chip[]>([
		...(cats.length > 1
			? cats
					.filter((c) => !f.cats.includes(c.nombre))
					.map((c) => ({ t: c.nombre, on: false, href: conCat(c.nombre) }))
			: []),
		...(hayOfertas && !f.ofertas
			? [{ t: 'En oferta', on: false, href: ir({ ofertas: true }) }]
			: []),
		...(hayAgotados && !f.stock ? [{ t: 'Con stock', on: false, href: ir({ stock: true }) }] : []),
		...(hayMayor && !f.mayor ? [{ t: 'Por mayor', on: false, href: ir({ mayor: true }) }] : [])
	]);
	const limpiar = $derived(hrefFiltros(base, { ...SIN_FILTROS, orden: f.orden }));

	/** Las casillas y el precio del costado (o del panel de precio): al cambiar, se navega. */
	function alCambiar(e: Event) {
		const form = e.currentTarget as HTMLFormElement;
		const d = new FormData(form);
		const num = (k: string) => {
			const n = Number(String(d.get(k) ?? '').replace(/\D/g, ''));
			return n > 0 ? n : null;
		};
		navegar(
			ir({
				cats: d.has('cat')
					? (d.getAll('cat') as string[])
					: form.querySelector('[name=cat]')
						? []
						: f.cats,
				ofertas: d.has('ofertas') || (!form.querySelector('[name=ofertas]') && f.ofertas),
				stock: d.has('stock') || (!form.querySelector('[name=stock]') && f.stock),
				mayor: d.has('mayor') || (!form.querySelector('[name=mayor]') && f.mayor),
				min: form.querySelector('[name=min]') ? num('min') : f.min,
				max: form.querySelector('[name=max]') ? num('max') : f.max
			})
		);
	}
	function enviar(e: SubmitEvent) {
		e.preventDefault();
		alCambiar(e);
	}
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
{#snippet ordenar()}
	<label class="tv-orden">
		<span>Ordenar</span>
		<select
			value={f.orden}
			onchange={(e) => navegar(ir({ orden: e.currentTarget.value as Filtros['orden'] }))}
		>
			{#each ORDENES as [k, t] (k)}<option value={k}>{t}</option>{/each}
		</select>
	</label>
{/snippet}

{#snippet chip(c: Chip)}
	<a
		class="tv-fchip"
		class:tv-fchip--on={c.on}
		href={c.href}
		data-sveltekit-reset="false"
		data-sveltekit-replacestate
		aria-label={c.on ? `Quitar el filtro ${c.t}` : `Filtrar por ${c.t}`}
		>{c.t}{#if c.on}<Icon icon={I.phXBold} size={11} />{/if}</a
	>
{/snippet}

{#snippet panelPrecio()}
	{#if conPrecio}
		<details class="tv-fchip-precio">
			<summary class="tv-fchip">Precio<Icon icon={I.phCaretDownBold} size={11} /></summary>
			<form class="tv-fprecio tv-tarjeta" action="{base}/catalogo" method="get" onsubmit={enviar}>
				{@render rango('chip')}
				<button class="tv-btn tv-btn--primario tv-btn--chico" type="submit">Aplicar</button>
			</form>
		</details>
	{/if}
{/snippet}

{#snippet rango(id: string)}
	<div class="tv-rango">
		<label
			><span class="tv-oculto">Precio mínimo</span><input
				id="min-{id}"
				name="min"
				inputmode="numeric"
				placeholder="$ mín."
				value={f.min ?? ''}
				autocomplete="off"
			/></label
		>
		<label
			><span class="tv-oculto">Precio máximo</span><input
				id="max-{id}"
				name="max"
				inputmode="numeric"
				placeholder="$ máx."
				value={f.max ?? ''}
				autocomplete="off"
			/></label
		>
	</div>
{/snippet}

{#snippet chips(clase: string)}
	{#if activos.length || opcionales.length || conPrecio}
		<div class="tv-fchips {clase}" role="group" aria-label="Filtros">
			{#each activos as c (c.t)}{@render chip(c)}{/each}
			{#each opcionales as c (c.t)}{@render chip(c)}{/each}
			{@render panelPrecio()}
		</div>
	{/if}
{/snippet}

{#snippet costado()}
	<form
		class="tv-filtros"
		action="{base}/catalogo"
		method="get"
		onchange={alCambiar}
		onsubmit={enviar}
	>
		{#if f.q}<input type="hidden" name="q" value={f.q} />{/if}
		{#if cats.length > 1}
			<fieldset>
				<legend class="tv-filtros-tit">Categoría</legend>
				{#each cats as c (c.nombre)}
					<label class="tv-casilla"
						><input
							type="checkbox"
							name="cat"
							value={c.nombre}
							checked={f.cats.includes(c.nombre)}
						/><span>{c.nombre}</span><small>{cuenta.get(c.nombre) ?? 0}</small></label
					>
				{/each}
			</fieldset>
		{/if}
		{#if conPrecio}
			<fieldset>
				<legend class="tv-filtros-tit">Precio</legend>
				{@render rango('costado')}
			</fieldset>
		{/if}
		{#if hayOfertas || hayAgotados || hayMayor}
			<fieldset>
				<legend class="tv-filtros-tit">Mostrar</legend>
				{#if hayAgotados}<label class="tv-casilla"
						><input type="checkbox" name="stock" checked={f.stock} /><span>Sólo con stock</span
						></label
					>{/if}
				{#if hayOfertas}<label class="tv-casilla"
						><input type="checkbox" name="ofertas" checked={f.ofertas} /><span>En oferta</span
						></label
					>{/if}
				{#if hayMayor}<label class="tv-casilla"
						><input type="checkbox" name="mayor" checked={f.mayor} /><span>Precio por mayor</span
						></label
					>{/if}
			</fieldset>
		{/if}
		<button type="submit" class="tv-oculto">Filtrar</button>
		{#if tienda.mayorista && hayMayor}
			<div class="tv-filtros-caja">
				<b>¿Comprás por cantidad?</b>
				<p>
					Desde {textoMinimo(tienda.mayorista, precio)} tenés precio&nbsp;mayorista.
				</p>
				<a href={ir({ mayor: true })} data-sveltekit-reset="false">Ver la lista</a>
			</div>
		{/if}
	</form>
{/snippet}

{#snippet resultados()}
	{#if !lista.length}
		<div class="tv-sinres tv-tarjeta">
			<Icon icon={I.phMagnifyingGlassBold} size={26} />
			<h2 class="tv-h3">
				{f.q ? `No encontramos productos con «${f.q}»` : 'No hay productos con estos filtros'}
			</h2>
			<p class="tv-suave">
				Probá con menos filtros o con otra palabra. {wa
					? 'Si no está, preguntanos: lo buscamos.'
					: ''}
			</p>
			<div class="tv-botones">
				<Boton href={hrefFiltros(base, SIN_FILTROS)}>Ver todos los productos</Boton>
				{#if wa}<Boton href={wa} tipo="secundario" externo>Preguntar por WhatsApp</Boton>{/if}
			</div>
		</div>
	{:else if op.variante === 'lista'}
		<ul class="tv-cat-filas tv-tarjeta" aria-label="Productos">
			{#each visibles as p, i (p.id)}
				<li>
					<TarjetaProducto
						producto={p}
						forma="lista"
						accion="agregar"
						ancho={200}
						prioridad={i < 4}
					/>
					{#if comparador}<BotonComparar producto={p} />{/if}
				</li>
			{/each}
		</ul>
	{:else}
		<Grilla
			items={visibles}
			columnas={op.columnas}
			recortar={false}
			clase="tv-cat-grilla"
			etiqueta="Productos"
		>
			{#snippet children(p, i)}
				<TarjetaProducto producto={p} accion="agregar" prioridad={i < 4} />
				{#if comparador}<BotonComparar producto={p} />{/if}
			{/snippet}
		</Grilla>
		{#if dir === 'mostrador' && wa && (f.q || hayFiltros(f)) && lista.length <= mostrar}
			<div class="tv-cat-wa tv-z" data-fondo="oscuro">
				<h2 class="tv-h3">¿No encontraste lo que&nbsp;buscabas?</h2>
				<p>Escribinos y te lo conseguimos.</p>
				<Boton href={wa} tipo="wa" externo>Pedir por WhatsApp</Boton>
			</div>
		{/if}
	{/if}
	{#if lista.length > mostrar}
		<div class="tv-cat-mas">
			<Boton tipo="secundario" onclick={() => (mostrar += 24)}
				>Ver más ({lista.length - mostrar})</Boton
			>
		</div>
	{/if}
{/snippet}

<Tienda {tienda} pagina="catalogo">
	{#if !tienda.productos.length}
		<div class="tv-ancho tv-pag">
			<VacioTienda {tienda} tema={temaDe(dir)} motivo="sin-productos" />
		</div>
	{:else}
		<div class="tv-pag tv-ancho" class:tv-con-bolsa={dir === 'pedido'}>
			<div class="tv-pag-main">
				<nav class="tv-migas" aria-label="Estás en">
					<a href={base || '/'}>Inicio</a><span aria-hidden="true"
						>{dir === 'mostrador' ? '›' : '/'}</span
					><span>{miga}</span>
				</nav>
				<div class="tv-cat-cab">
					<div>
						<h1 class="tv-h1 tv-pag-titulo">{titulo}</h1>
						{#if !f.q}<p class="tv-suave tv-cat-cuantos">
								{cuantos(lista.length, 'producto', 'productos')}
							</p>{/if}
					</div>
					{@render ordenar()}
				</div>
				{#if op.filtros === 'arriba'}
					{@render chips('')}
				{:else}
					{#if activos.length}
						<div class="tv-fchips tv-solo-compu" role="group" aria-label="Filtros puestos">
							{#each activos as c (c.t)}{@render chip(c)}{/each}
							<a class="tv-flimpiar" href={limpiar} data-sveltekit-reset="false">Limpiar filtros</a>
						</div>
					{/if}
					{@render chips('tv-solo-movil')}
				{/if}
				<div class="tv-cat-cuerpo" class:tv-cat-cuerpo--costado={op.filtros === 'costado'}>
					{#if op.filtros === 'costado'}<div class="tv-solo-compu">{@render costado()}</div>{/if}
					<div class="tv-cat-res" style:--cat-cols={op.columnas}>{@render resultados()}</div>
				</div>
			</div>
			{#if dir === 'pedido'}<div class="tv-pag-lado tv-solo-compu"><Bolsa /></div>{/if}
		</div>
		{#if dir === 'pedido'}<BarraPedido />{/if}
	{/if}
</Tienda>
