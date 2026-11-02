<script lang="ts">
	import Qr from '#lib/marca/Qr.svelte';
	import { unsplash } from '#lib/landing/formato.ts';
	import type { Plantilla } from '#lib/plantillas.ts';
	import { completar, linkQr, type Tarjeta, type Variables } from './tarjeta.ts';

	/**
	 * La tarjeta A6 (105 × 148 mm) de la pieza 7a de Imprimibles, con la marca de la tienda: sus
	 * colores y tipografía salen de la plantilla. Mide en milímetros reales: la vista previa la
	 * escala y la impresión la usa tal cual.
	 */
	let {
		t,
		cara,
		tienda,
		paleta,
		foto,
		v
	}: {
		t: Tarjeta;
		cara: 'frente' | 'dorso';
		tienda: { nombre: string; url: string; whatsapp: string; dominio: string };
		paleta: Pick<Plantilla, 'bg' | 'tx' | 'ac' | 'ff' | 'oscura'>;
		foto: string | null;
		v: Variables;
	} = $props();

	const titulo = $derived(completar(t.titulo, v));
	const mensaje = $derived(completar(t.mensaje, v));
	const dorsoTitulo = $derived(completar(t.dorsoTitulo, v));
	const dorsoTexto = $derived(completar(t.dorsoTexto, v));
	const qr = $derived(linkQr(t, tienda));
	// Sobre el color de acento, el texto va con el fondo de la plantilla (como 7b sobre terracota).
	const sello = $derived(t.modelo === 'sello');
	const fondo = $derived(sello ? paleta.ac : paleta.bg);
	const tinta = $derived(sello ? paleta.bg : paleta.tx);
</script>

{#if cara === 'frente'}
	<div class="a6 frente m-{t.modelo}" style:background={fondo} style:color={tinta}>
		{#if t.modelo === 'arco'}
			<div class="marca" style:font-family={paleta.ff}>{tienda.nombre}</div>
			<div class="arco" style:background="color-mix(in srgb, {paleta.tx} 10%, {paleta.bg})">
				{#if foto}<img src={unsplash(foto, 500)} alt="" />{/if}
			</div>
			<div class="titulo" style:font-family={paleta.ff}>{titulo}</div>
			<div class="mensaje">{mensaje}</div>
		{:else if t.modelo === 'sello'}
			<div class="titulo grande" style:font-family={paleta.ff}>{titulo}</div>
			<div class="mensaje">{mensaje}</div>
			<div class="marca abajo" style:font-family={paleta.ff}>{tienda.nombre}</div>
		{:else}
			<div class="borde" style:border-color="color-mix(in srgb, {paleta.tx} 25%, transparent)">
				<div class="marca" style:font-family={paleta.ff}>{tienda.nombre}</div>
				<div class="titulo" style:font-family={paleta.ff}>{titulo}</div>
				<div class="mensaje">{mensaje}</div>
				<div class="firma" style:font-family={paleta.ff}>— {tienda.nombre}</div>
			</div>
		{/if}
	</div>
{:else}
	<div class="a6 dorso" style:color={paleta.tx}>
		<div class="d-t">{dorsoTitulo}</div>
		<div class="d-x">{dorsoTexto}</div>
		{#if qr}
			<Qr
				texto={qr}
				alt="QR"
				style="margin-top:6mm;align-self:center;width:34mm;height:34mm;display:block"
			/>
		{/if}
		{#if t.qr === 'cupon' && t.cupon}
			<div class="cupon">
				<div class="c-t">TU CUPÓN</div>
				<div class="c-c">{t.cupon}</div>
			</div>
		{/if}
		<div class="pie">
			<span>{tienda.dominio}{t.instagram ? ` · @${t.instagram}` : ''}</span><span class="hecho"
				><img src="/logo.svg" alt="" />Hecho con Tu Local Web</span
			>
		</div>
	</div>
{/if}

<style>
	.a6 {
		width: 105mm;
		height: 148mm;
		position: relative;
		overflow: hidden;
		padding: 10mm;
		display: flex;
		flex-direction: column;
		box-sizing: border-box;
		font-family: 'Instrument Sans', system-ui, sans-serif;
	}
	.frente.m-arco,
	.frente.m-carta {
		align-items: center;
		text-align: center;
	}
	.marca {
		font-size: 16px;
		font-weight: 700;
		letter-spacing: 0.2em;
		text-transform: uppercase;
	}
	.arco {
		margin-top: 8mm;
		width: 56mm;
		height: 66mm;
		border-radius: 28mm 28mm 2mm 2mm;
		overflow: hidden;
		flex: none;
	}
	.arco img {
		width: 100%;
		height: 100%;
		object-fit: cover;
		display: block;
	}
	.titulo {
		margin-top: 8mm;
		font-size: 26px;
		line-height: 1.05;
		letter-spacing: -0.02em;
		font-weight: 600;
		text-wrap: balance;
	}
	.titulo.grande {
		margin-top: 0;
		font-size: 34px;
		line-height: 1;
	}
	.mensaje {
		margin-top: 3mm;
		font-size: 10.5px;
		line-height: 1.5;
		opacity: 0.8;
		text-wrap: pretty;
	}
	.abajo {
		margin-top: auto;
	}
	.borde {
		flex: 1;
		border: 1px solid;
		padding: 8mm 6mm;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
	}
	.borde .titulo {
		margin-top: 12mm;
	}
	.firma {
		margin-top: 10mm;
		font-size: 13px;
		font-style: italic;
	}
	.dorso {
		background: #fbf7f2;
	}
	.d-t {
		font-size: 15px;
		font-weight: 600;
	}
	.d-x {
		margin-top: 1mm;
		font-size: 10.5px;
		line-height: 1.5;
		color: #6b5244;
		text-wrap: pretty;
	}
	.cupon {
		margin-top: 6mm;
		padding: 4mm;
		border-radius: 3mm;
		background: #fff;
		text-align: center;
	}
	.c-t {
		font-size: 9px;
		font-weight: 700;
		letter-spacing: 0.12em;
		color: #7a6252;
	}
	.c-c {
		font-size: 20px;
		font-weight: 700;
		letter-spacing: 0.04em;
	}
	.pie {
		margin-top: auto;
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 4mm;
		font-size: 9px;
		color: #7a6252;
	}
	.hecho {
		display: flex;
		align-items: center;
		gap: 4px;
		white-space: nowrap;
	}
	.hecho img {
		width: 3.5mm;
	}
</style>
