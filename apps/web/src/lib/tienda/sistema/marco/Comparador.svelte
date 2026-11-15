<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { getUi } from '#lib/ui.svelte.ts';
	import { tarjetaDe } from '../../seleccion.svelte.ts';
	import type { ProductoPublico } from '../../tipos.ts';
	import { agregar, sumaDirecto } from '../acciones.ts';
	import { compararDe } from '../comparar.svelte.ts';
	import { usarTienda } from '../contexto.ts';
	import { productoHref, waHref } from '../datos.ts';
	import { conTransferencia, cuotas, descuento, precio } from '../formato.ts';
	import Boton from '../primitivas/Boton.svelte';
	import Foto from '../primitivas/Foto.svelte';

	/**
	 * El comparador de Andes sobre cualquier página: una barra abajo con cuántos hay y, al abrirla,
	 * los productos lado a lado (precio, con transferencia, descuento, categoría, stock y reseñas) con
	 * lo mejor de cada fila marcado. Lo pone marco/Tienda.svelte sólo si el preset lo trae.
	 */
	const ctx = usarTienda();
	const ui = getUi();
	const t = $derived(ctx.tienda);
	const cmp = $derived(compararDe(t.slug));
	const ps = $derived(
		cmp.ids.flatMap((id) => {
			const p = t.productos.find((x) => x.id === id);
			return p ? [p] : [];
		})
	);
	let dialogo = $state<HTMLDialogElement>();
	$effect(() => {
		if (!dialogo) return;
		if (cmp.abierto && !dialogo.open) dialogo.showModal();
		else if (!cmp.abierto && dialogo.open) dialogo.close();
	});

	const pct = (p: ProductoPublico) => {
		const d = descuento(p.precio, p.anterior);
		return d ? Number(d.replace(/\D/g, '')) : 0;
	};
	const resenas = (p: ProductoPublico) => t.resenas.filter((r) => r.producto === p.nombre).length;
	const desde = (p: ProductoPublico) => tarjetaDe(p).precio;
	const filas = $derived.by(() => {
		const varios = ps.length > 1;
		const min = Math.min(...ps.map(desde));
		const maxOff = Math.max(...ps.map(pct));
		const maxRes = Math.max(...ps.map(resenas));
		const algunoAgotado = ps.some((p) => tarjetaDe(p).agotado);
		const transf = t.cobros.transferencia > 0 && !!t.cobros.datosTransferencia;
		return ps.map((p) => ({
			p,
			filas: [
				{ t: 'Precio', v: precio(desde(p)), mejor: varios && desde(p) === min },
				...(transf
					? [
							{
								t: 'Con transferencia',
								v: precio(conTransferencia(desde(p), t.cobros.transferencia)),
								mejor: varios && desde(p) === min
							}
						]
					: []),
				{
					t: 'Descuento',
					v: pct(p) ? `−${pct(p)} %` : '—',
					mejor: varios && pct(p) > 0 && pct(p) === maxOff
				},
				{ t: 'Categoría', v: p.categoria ?? '—', mejor: false },
				{
					t: 'Stock',
					v: tarjetaDe(p).agotado ? 'Sin stock' : 'Disponible',
					mejor: varios && algunoAgotado && !tarjetaDe(p).agotado
				},
				{
					t: 'Reseñas',
					v: String(resenas(p)),
					mejor: varios && resenas(p) > 0 && resenas(p) === maxRes
				}
			]
		}));
	});
	const wa = $derived(
		waHref(
			t,
			ps.length
				? `¡Hola! Estoy comparando ${ps.map((p) => p.nombre).join(', ')}. ¿Cuál me recomiendan?`
				: '¡Hola! ¿Me ayudan a elegir?'
		)
	);
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
{#if ps.length}
	<div class="tv-cmp-barra">
		<span
			><b>{ps.length}</b>
			{ps.length === 1 ? 'producto para comparar' : 'productos para comparar'}</span
		>
		<button
			type="button"
			class="tv-btn tv-btn--primario tv-btn--chico"
			onclick={() => (cmp.abierto = true)}>Comparar</button
		>
	</div>
{/if}

<dialog
	bind:this={dialogo}
	class="tv-cmp"
	aria-labelledby="tv-cmp-titulo"
	onclose={() => (cmp.abierto = false)}
>
	<div class="tv-cmp-cab">
		<h2 id="tv-cmp-titulo" class="tv-h2">Compará lado a lado</h2>
		<button type="button" class="tv-icono" aria-label="Cerrar" onclick={() => (cmp.abierto = false)}
			><Icon icon={I.phXBold} size={18} /></button
		>
	</div>
	<p class="tv-suave">
		{ps.length > 1 ? 'Marcamos lo mejor de cada fila.' : 'Sumá otro producto para compararlo.'}
	</p>
	<ul class="tv-cmp-lista">
		{#each filas as c (c.p.id)}
			<li class="tv-tarjeta">
				<a href={productoHref(t, c.p)} tabindex="-1" aria-hidden="true"
					><Foto src={c.p.foto} alt="" nombre={c.p.nombre} ancho={500} proporcion="4/3" /></a
				>
				<a class="tv-cmp-nombre" href={productoHref(t, c.p)}>{c.p.nombre}</a>
				{#if t.cobros.mercadoPago && cuotas(desde(c.p), t.cobros.cuotas)}<span
						class="tv-chico tv-suave">{cuotas(desde(c.p), t.cobros.cuotas)}</span
					>{/if}
				<dl>
					{#each c.filas as f (f.t)}
						<div class:tv-cmp-mejor={f.mejor}>
							<dt>{f.t}</dt>
							<dd>
								{#if f.mejor}<Icon icon={I.phTrophyFill} size={13} />{/if}{f.v}
							</dd>
						</div>
					{/each}
				</dl>
				<div class="tv-cmp-acciones">
					{#if sumaDirecto(ctx, c.p)}
						<Boton chico bloque onclick={() => agregar(ctx, c.p, ui)}>Agregar al carrito</Boton>
					{:else}
						<Boton chico bloque href={productoHref(t, c.p)}
							>{tarjetaDe(c.p).agotado ? 'Ver el producto' : 'Elegir opciones'}</Boton
						>
					{/if}
					<button type="button" class="tv-cmp-quitar" onclick={() => cmp.alternar(c.p.id)}
						>Quitar de la comparación</button
					>
				</div>
			</li>
		{/each}
	</ul>
	{#if wa}
		<div class="tv-cmp-wa">
			<div>
				<b>¿No te decidís?</b>
				<p>Contanos para qué lo vas a usar y te recomendamos uno por&nbsp;WhatsApp.</p>
			</div>
			<Boton href={wa} tipo="wa" chico externo>Preguntar</Boton>
		</div>
	{/if}
</dialog>
