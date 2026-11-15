<script lang="ts">
	import { onMount } from 'svelte';
	import Icon from '#lib/components/Icon.svelte';
	import Interruptor from '#lib/components/Interruptor.svelte';
	import Modal from '#lib/components/Modal.svelte';
	import * as I from '#lib/icons.ts';
	import { precio } from '#lib/tienda/sistema/formato.ts';
	import {
		CONVERSION,
		MAX_VENTANAS,
		queFaltaVentana,
		VENTANAS,
		type PlantillaVentana,
		type Ventana
	} from '#lib/tienda/sistema/promos.ts';
	import type { ConversionGuardada } from '#lib/tienda/estilo/estilo.ts';
	import CampoBloque from './CampoBloque.svelte';
	import PanelVentana from './PanelVentana.svelte';
	import type { Ayuda } from './campos.ts';
	import type { DisenoV2 } from './diseno.ts';
	import * as V from './ventanas.ts';

	/**
	 * Mi tienda › Promociones (tienda-promociones.md › 4): la lista de ventanas (activar, ordenar,
	 * duplicar, borrar), crear desde plantillas, el panel de cada una, sus números de los últimos 30
	 * días (GET /panel/promos/metricas) y las herramientas para vender más (estilo.conversion), con
	 * los campos del catálogo. Cada cambio pasa por `oncambiar` (historial y borrador, como los
	 * bloques); `onver` abre una ventana en la vista previa.
	 */
	let {
		diseno,
		ayuda,
		envioGratis,
		abierta = $bindable(null),
		oncambiar,
		onver,
		onborrada
	}: {
		diseno: DisenoV2;
		ayuda: Ayuda;
		envioGratis: number | null;
		abierta?: string | null;
		oncambiar: (d: DisenoV2, confirmar?: boolean) => void;
		onver: (id: string) => void;
		onborrada: (nombre: string) => void;
	} = $props();

	const lista = $derived(V.ventanasDe(diseno));
	const ventana = $derived(lista.find((w) => w.id === abierta) ?? null);
	const poner = (l: Ventana[], confirmar = true) => oncambiar(V.conVentanas(diseno, l), confirmar);

	// ── Números de los últimos 30 días ──
	type Numeros = { vistas: number; envios: number; clics: number; ventas: number; monto: number };
	let numeros = $state<Record<string, Numeros>>({});
	let total = $state<Numeros | null>(null);
	onMount(() => {
		let vivo = true;
		fetch('/panel/promos/metricas')
			.then((r) => (r.ok ? r.json() : null))
			.then((x: unknown) => {
				const m = x as { total: Numeros; ventanas: (Numeros & { id: string })[] } | null;
				if (!vivo || !m) return;
				total = m.total;
				numeros = Object.fromEntries(m.ventanas.map((x) => [x.id, x]));
			})
			.catch(() => {});
		return () => (vivo = false);
	});

	// ── Crear ──
	let eligiendo = $state(false);
	function crear(p: PlantillaVentana | null) {
		eligiendo = false;
		const w = V.desdePlantilla(p, diseno, envioGratis);
		poner([...lista, w]);
		abierta = w.id;
	}

	function estado(w: Ventana): { t: string; k: 'ok' | 'pausa' | 'falta' } {
		if (!w.activa) return { t: 'Pausada', k: 'pausa' };
		return queFaltaVentana(w) ? { t: 'Incompleta', k: 'falta' } : { t: 'Activa', k: 'ok' };
	}
	const pct = (a: number, b: number) => (b ? `${Math.round((a / b) * 100)}\u00a0%` : '');
	function linea(w: Ventana) {
		const n = numeros[w.id];
		if (!n?.vistas) return V.resumen(w);
		const tasa =
			w.contenido === 'captacion' || w.contenido === 'ruleta'
				? `${pct(n.envios, n.vistas)} dejó el mail`
				: `${pct(n.clics, n.vistas)} tocó`;
		return `${n.vistas.toLocaleString('es-AR')} vistas · ${tasa}${n.ventas ? ` · ${n.ventas} ${n.ventas === 1 ? 'venta' : 'ventas'}` : ''}`;
	}

	function borrar(i: number) {
		const w = lista[i];
		poner(lista.filter((_, j) => j !== i));
		abierta = null;
		onborrada(w.nombre);
	}

	// ── Conversión ──
	const conv = $derived((diseno.estilo?.conversion ?? {}) as Record<string, unknown>);
	function ponerConversion(id: string, v: unknown, confirmar: boolean) {
		const c = { ...conv };
		if (v === undefined) delete c[id];
		else c[id] = v;
		oncambiar(
			{ ...diseno, estilo: { ...diseno.estilo, conversion: c as ConversionGuardada } },
			confirmar
		);
	}
</script>

{#if ventana}
	{@const i = lista.indexOf(ventana)}
	<PanelVentana
		{ventana}
		{ayuda}
		numeros={numeros[ventana.id] ?? null}
		puedeDuplicar={V.puedeSumar(diseno)}
		onventana={(w, c) =>
			poner(
				lista.map((x) => (x.id === w.id ? w : x)),
				c
			)}
		onvolver={() => (abierta = null)}
		onver={() => onver(ventana.id)}
		onduplicar={() => {
			const l = V.duplicar(lista, i, diseno);
			poner(l);
			abierta = l[i + 1].id;
		}}
		onborrar={() => borrar(i)}
	/>
{:else}
	<div class="cab">
		<div>
			<h2 class="t">Ventanas</h2>
			<p class="nota">
				Avisos que se abren solos: una promo, un cupón, pedir el mail o una ruleta. Sale una por
				página y nunca en el&nbsp;pago.
			</p>
		</div>
	</div>
	{#if total && total.vistas}
		<dl class="totales" aria-label="Últimos 30 días">
			<div>
				<dt>Vistas</dt>
				<dd>{total.vistas.toLocaleString('es-AR')}</dd>
			</div>
			<div>
				<dt>Mails</dt>
				<dd>{total.envios.toLocaleString('es-AR')}</dd>
			</div>
			<div>
				<dt>Ventas</dt>
				<dd>{precio(total.monto)}</dd>
			</div>
		</dl>
	{/if}
	{#if lista.length}
		<ul class="lista">
			{#each lista as w, i (w.id)}
				{@const e = estado(w)}
				<li class="fila" data-ventana={w.id}>
					<div class="orden">
						<button
							type="button"
							class="mini-btn icono"
							aria-label="Subir {w.nombre}"
							disabled={i === 0}
							onclick={() => poner(V.mover(lista, i, i - 1))}
							><Icon icon={I.phArrowUpBold} size={13} /></button
						>
						<button
							type="button"
							class="mini-btn icono"
							aria-label="Bajar {w.nombre}"
							disabled={i === lista.length - 1}
							onclick={() => poner(V.mover(lista, i, i + 1))}
							><Icon icon={I.phArrowDownBold} size={13} /></button
						>
					</div>
					<button type="button" class="abrir" onclick={() => (abierta = w.id)}>
						<span class="nombre"><b>{w.nombre}</b><span class="chip chip--{e.k}">{e.t}</span></span>
						<span class="det">{V.nombreContenido(w.contenido)} · {linea(w)}</span>
					</button>
					<Interruptor
						checked={w.activa}
						label="{w.activa ? 'Pausar' : 'Activar'} {w.nombre}"
						onchange={(v) => poner(lista.map((x) => (x.id === w.id ? { ...x, activa: v } : x)))}
					/>
				</li>
			{/each}
		</ul>
	{:else}
		<p class="vacio">Todavía no tenés ventanas. Empezá por una plantilla: en un minuto queda.</p>
	{/if}
	<button
		type="button"
		class="nueva"
		disabled={!V.puedeSumar(diseno)}
		onclick={() => (eligiendo = true)}
		><Icon icon={I.phPlusBold} size={14} />{V.puedeSumar(diseno)
			? 'Nueva ventana'
			: `Llegaste a ${MAX_VENTANAS} ventanas`}</button
	>

	<div class="cab sep">
		<div>
			<h2 class="t">{CONVERSION.nombre}</h2>
			<p class="nota">{CONVERSION.descripcion}</p>
		</div>
	</div>
	<div class="conversion">
		{#each CONVERSION.campos as c (c.id)}
			<div class="herr">
				<CampoBloque
					campo={c}
					valor={conv[c.id]}
					{ayuda}
					onvalor={(v, conf) => ponerConversion(c.id, v, conf)}
				/>
				{#if c.descripcion}<p class="nota">{c.descripcion}</p>{/if}
				{#if c.id === 'envioGratis' && !envioGratis}<p class="nota">
						Configurá el monto en Envíos para que se&nbsp;vea.
					</p>{/if}
			</div>
		{/each}
	</div>
{/if}

<Modal bind:open={eligiendo} titulo="Nueva ventana" ancho={720}>
	<div class="m-cab">
		<h2>Nueva ventana</h2>
		<p>Elegí una plantilla y cambiá lo que quieras. Se ve en tu tienda al publicar.</p>
	</div>
	<div class="plantillas">
		{#each VENTANAS.plantillas as p (p.id)}
			<button type="button" class="tpl" onclick={() => crear(p)}>
				<span class="mini mini--{p.ventana.variante}" aria-hidden="true"><span></span></span>
				<b>{p.nombre}</b>
				<small>{p.descripcion}</small>
			</button>
		{/each}
		<button type="button" class="tpl" onclick={() => crear(null)}>
			<span class="mini mini--blanco" aria-hidden="true"
				><Icon icon={I.phPlusBold} size={18} /></span
			>
			<b>En blanco</b>
			<small>Arrancá de cero, con todo para elegir.</small>
		</button>
	</div>
</Modal>

<style>
	.cab {
		display: flex;
		align-items: flex-start;
		gap: 8px;
		padding: 4px 4px 0;
	}
	.cab.sep {
		margin-top: 14px;
		padding-top: 14px;
		border-top: 1px solid var(--ln);
	}
	.t {
		margin: 0;
		font-size: 17px;
		font-weight: 600;
		letter-spacing: -0.02em;
	}
	.nota {
		margin: 2px 0 0;
		font-size: 12.5px;
		color: var(--mu);
		text-wrap: pretty;
	}
	.totales {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 6px;
		margin: 6px 0 2px;
	}
	.totales div {
		padding: 8px 10px;
		border-radius: 12px;
		background: var(--chip);
	}
	.totales dt {
		font-size: 11.5px;
		color: var(--mu);
	}
	.totales dd {
		margin: 2px 0 0;
		font-size: 15px;
		font-weight: 600;
		font-variant-numeric: tabular-nums;
		white-space: nowrap;
	}
	.lista {
		display: flex;
		flex-direction: column;
		gap: 6px;
		margin: 6px 0 0;
		padding: 0;
		list-style: none;
	}
	.fila {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 6px 8px 6px 4px;
		border: 1px solid var(--ln);
		border-radius: 14px;
	}
	.orden {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}
	.abrir {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 2px;
		padding: 6px 4px;
		border: 0;
		background: none;
		color: var(--tx);
		font: inherit;
		text-align: left;
		cursor: pointer;
	}
	.nombre {
		display: flex;
		align-items: center;
		gap: 6px;
		min-width: 0;
		font-size: 14px;
	}
	.nombre b {
		min-width: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.det {
		font-size: 12px;
		color: var(--mu);
		text-wrap: pretty;
	}
	.chip {
		flex: none;
		padding: 2px 8px;
		border-radius: 999px;
		font-size: 11px;
		font-weight: 600;
		background: var(--chip);
		color: var(--mu);
	}
	.chip--ok {
		background: color-mix(in srgb, #1f7a4d 14%, transparent);
		color: #1f7a4d;
	}
	.chip--falta {
		background: color-mix(in srgb, var(--alerta) 22%, transparent);
		color: #8a5a00;
	}
	.vacio {
		margin: 8px 4px 0;
		font-size: 13px;
		color: var(--mu);
		text-wrap: pretty;
	}
	.nueva {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		min-height: 44px;
		margin-top: 8px;
		border-radius: 999px;
		border: 1.5px dashed var(--ln2);
		background: none;
		color: var(--tx);
		font: inherit;
		font-size: 14px;
		font-weight: 600;
		cursor: pointer;
	}
	.nueva:disabled {
		opacity: 0.5;
		cursor: default;
	}
	.conversion {
		display: flex;
		flex-direction: column;
		gap: 12px;
		padding: 8px 4px 4px;
	}
	.herr {
		display: flex;
		flex-direction: column;
		gap: 4px;
	}
	.mini-btn {
		display: inline-grid;
		place-items: center;
		width: 30px;
		min-height: 26px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: var(--sf);
		color: var(--tx);
		cursor: pointer;
	}
	.mini-btn:disabled {
		opacity: 0.3;
		cursor: default;
	}
	.m-cab h2 {
		margin: 0;
		font-size: 22px;
		letter-spacing: -0.02em;
	}
	.m-cab p {
		margin: 4px 0 16px;
		color: var(--mu);
		font-size: 14px;
		text-wrap: pretty;
	}
	.plantillas {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(190px, 1fr));
		gap: 10px;
	}
	.tpl {
		display: flex;
		flex-direction: column;
		gap: 4px;
		padding: 8px;
		border: 2px solid var(--ln);
		border-radius: 16px;
		background: var(--sf);
		color: var(--tx);
		font: inherit;
		text-align: left;
		cursor: pointer;
	}
	.tpl:hover,
	.tpl:focus-visible {
		border-color: var(--terracota);
	}
	.tpl b {
		font-size: 14px;
	}
	.tpl small {
		font-size: 12.5px;
		color: var(--mu);
		text-wrap: pretty;
	}
	.mini {
		position: relative;
		display: grid;
		place-items: center;
		height: 70px;
		margin-bottom: 4px;
		border-radius: 10px;
		background:
			linear-gradient(var(--ln) 0 0) 10px 10px / 40% 6px no-repeat,
			linear-gradient(var(--ln) 0 0) 10px 24px / 70% 18px no-repeat,
			var(--sf2, #f4efe9);
		overflow: hidden;
		color: var(--mu);
	}
	.mini span {
		position: absolute;
		background: var(--sf);
		border: 1.5px solid var(--terracota);
		border-radius: 4px;
	}
	.mini--centrada span {
		inset: 14px 30%;
	}
	.mini--lateral span {
		left: 8px;
		bottom: 8px;
		width: 40%;
		height: 28px;
	}
	.mini--completa span {
		inset: 4px;
	}
	.mini--inferior span {
		left: 10%;
		right: 10%;
		bottom: -2px;
		height: 32px;
		border-radius: 6px 6px 0 0;
	}
	@media (max-width: 760px) {
		.mini-btn {
			width: 44px;
			min-height: 44px;
		}
	}
</style>
