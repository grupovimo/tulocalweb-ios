<script lang="ts">
	import PieEstilo from '../../PieEstilo.svelte';
	import { AltaNewsletter } from '../../newsletter.svelte.ts';
	import { usarTienda } from '../contexto.ts';
	import { categoriaHref, waHref } from '../datos.ts';
	import type { Pie } from '../presets.ts';
	import Boton from '../primitivas/Boton.svelte';

	/**
	 * El pie de la tienda (estilo.pie), igual en todas las páginas:
	 * - `simple` (A-vacia, C-moda): una línea con el botón de arrepentimiento, los legales, los medios
	 *   de pago y «Hecha con Tu Local Web».
	 * - `columnas` (A-moda, B-moda): el nombre con la newsletter (o el WhatsApp), Tienda, Ayuda y El
	 *   local; abajo, la línea legal.
	 * - `grande` (A-cafe): el nombre en grande, los links en fila y la línea legal.
	 * Siempre: botón de arrepentimiento, Defensa de las y los consumidores, términos y privacidad (lo
	 * que pide la ley y tenían las plantillas), y la línea y las redes de Mi tienda › Estilo.
	 *
	 * Uso: lo pone marco/Tienda.svelte. Props: `variante` para forzar una. Lleva `data-t-pie`
	 * (--t-pie y --t-pie-texto).
	 */
	let { variante }: { variante?: Pie } = $props();
	const ctx = usarTienda();
	const t = $derived(ctx.tienda);
	const v = $derived(variante ?? ctx.estilo.pie);
	const nl = new AltaNewsletter(() => ctx.tienda.slug);
	const wa = $derived(waHref(t));
	const anio = new Date().getFullYear();

	const medios = $derived(
		[
			...(t.cobros.mercadoPago ? ['Visa', 'Mastercard', 'Mercado Pago'] : []),
			...(t.cobros.datosTransferencia ? ['Transferencia'] : [])
		].join(' · ')
	);
	const tiendaLinks = $derived([
		{ t: 'Todos los productos', href: `${ctx.base}/catalogo` },
		...t.categorias
			.filter((c) => c.cantidad > 0)
			.slice(0, 4)
			.map((c) => ({ t: c.nombre, href: categoriaHref(t, c.nombre) }))
	]);
	const ayuda = $derived([
		{ t: 'Envíos', href: `${ctx.base}/nosotros` },
		{ t: 'Cambios y devoluciones', href: `${ctx.base}/arrepentimiento` },
		{ t: 'Medios de pago', href: `${ctx.base}/nosotros` },
		...(t.turnos ? [{ t: 'Reservar turno', href: `${ctx.base}/turnos` }] : [])
	]);
	const local = $derived(
		[
			t.direccion || t.envios.retiroDireccion,
			t.envios.retiroHorario,
			t.ciudad && !(t.direccion || '').includes(t.ciudad) ? t.ciudad : ''
		].filter(Boolean)
	);
	const lema = $derived(t.portada?.lema?.trim() ?? '');
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
{#snippet legales()}
	<div class="tv-pie-legal">
		<nav class="tv-pie-legales" aria-label="Legales">
			<a class="tv-pie-arrep" href="{ctx.base}/arrepentimiento">Botón de arrepentimiento</a>
			<a href="/legales/defensa-del-consumidor">Defensa de las y los consumidores</a>
			<a href="/legales/terminos">Términos</a>
			<a href="/legales/privacidad">Privacidad</a>
		</nav>
		{#if medios && v !== 'grande'}<span class="tv-pie-medios">{medios}</span>{/if}
		<a class="tv-pie-hecha" href="https://tulocalweb.com.ar" target="_blank" rel="noopener"
			>Hecha con Tu Local Web</a
		>
	</div>
{/snippet}

{#snippet newsletter()}
	<form class="tv-pie-nl" onsubmit={nl.enviar} novalidate>
		<label for="tv-pie-email" class="tv-oculto">Tu correo para las novedades</label>
		<input
			id="tv-pie-email"
			type="email"
			name="email"
			required
			maxlength="120"
			autocomplete="email"
			placeholder="Tu email"
			bind:value={nl.email}
		/>
		<button type="submit" disabled={nl.enviando}>Sumarme</button>
	</form>
	{#if nl.aviso}<p class="tv-pie-aviso" role="status">
			{nl.aviso.texto}{#if nl.aviso.link}&nbsp;<a href={nl.aviso.link}>Confirmar acá</a>{/if}
		</p>{/if}
{/snippet}

<footer class="tv-pie" data-variante={v} data-t-pie>
	<div class="tv-ancho">
		{#if v === 'columnas'}
			<div class="tv-pie-cols">
				<div class="tv-pie-marca">
					<p class="tv-pie-nombre">{t.nombre}</p>
					{#if ctx.dir === 'vidriera'}
						<p class="tv-pie-texto">Te avisamos cuando hay novedades. Un mail por mes, nada más.</p>
						{@render newsletter()}
					{:else}
						{#if lema}<p class="tv-pie-texto">{lema}</p>{/if}
						{#if wa}<Boton href={wa} tipo="wa" chico externo>WhatsApp</Boton>{/if}
					{/if}
				</div>
				<nav class="tv-pie-col" aria-label="Tienda">
					<p class="tv-pie-tit">Tienda</p>
					{#each tiendaLinks as l (l.href)}<a href={l.href}>{l.t}</a>{/each}
				</nav>
				<nav class="tv-pie-col" aria-label="Ayuda">
					<p class="tv-pie-tit">Ayuda</p>
					{#each ayuda as l (l.t)}<a href={l.href}>{l.t}</a>{/each}
				</nav>
				{#if local.length || wa}
					<div class="tv-pie-col">
						<p class="tv-pie-tit">El local</p>
						{#each local as l (l)}<span>{l}</span>{/each}
						{#if wa}<a href={wa} target="_blank" rel="noopener">WhatsApp</a>{/if}
					</div>
				{:else if medios}
					<div class="tv-pie-col">
						<p class="tv-pie-tit">Pagá como quieras</p>
						<span>{medios}</span>
					</div>
				{/if}
			</div>
		{:else if v === 'grande'}
			<div class="tv-pie-grande">
				<p class="tv-pie-nombre">{t.nombre}</p>
				<nav class="tv-pie-fila" aria-label="Tienda">
					{#each tiendaLinks.slice(0, 3) as l (l.href)}<a href={l.href}>{l.t}</a>{/each}
					<a href="{ctx.base}/nosotros">Nosotros</a>
					{#if wa}<a href={wa} target="_blank" rel="noopener">WhatsApp</a>{/if}
				</nav>
			</div>
		{/if}
		{@render legales()}
		<PieEstilo tienda={t} />
		<p class="tv-pie-copy">© {anio} {t.nombre}{t.ciudad ? ` · ${t.ciudad}` : ''}</p>
	</div>
</footer>
