<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { page } from '$app/state';
	import Icon from '#lib/components/Icon.svelte';
	import Interruptor from '#lib/components/Interruptor.svelte';
	import Modal from '#lib/components/Modal.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { cara, unsplash } from '#lib/landing/formato.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { pesos } from '#lib/panel/formato.ts';
	import { hace, iniciales } from '#lib/panel/listas/clientes.ts';
	import {
		CUPONES,
		DEMORAS,
		idCupon,
		problemaRecordatorios,
		textoCupon,
		textoDemora,
		type RecordatoriosCarritos
	} from '#lib/panel/recordatorios.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/** Carritos abandonados (design/PanelVentas.dc.html): lo que quedó, lo recuperado y los recordatorios. */
	let { data } = $props();
	const ui = getUi();
	const k = $derived(data.carritos);
	const pendientes = $derived(k.carritos.filter((x) => !x.recordado));
	const ticket = $derived(k.abiertos ? k.monto / k.abiertos : 0);
	/**
	 * Con la API real, los recordatorios van por mail y no cuestan nada (WhatsApp Business todavía no
	 * se conecta). Los de WhatsApp y el segundo aviso con descuento sólo existen en la demo.
	 */
	const demo = $derived(!!page.data.demo);

	async function post(ruta: string, cuerpo: object = {}) {
		const r = await fetch(`/panel/carritos-abandonados/${ruta}`, {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify(cuerpo)
		});
		await invalidateAll();
		return r;
	}
	async function todos() {
		const r = await post('todos');
		const { enviados } = (await r.json()) as { enviados: number };
		ui.toast({
			k: demo ? 'wa' : 'ok',
			title: enviados
				? `${enviados} ${enviados === 1 ? 'recordatorio' : 'recordatorios'} en camino`
				: 'Ya estaban todos recordados',
			desc: !enviados
				? 'No se mandó nada de nuevo'
				: demo
					? `Costo total ${pesos(enviados * k.costoMensaje)} · te avisamos cuando compren`
					: 'Por mail, con el link a su carrito'
		});
	}
	async function uno(c: (typeof k.carritos)[number]) {
		const r = await post(c.id);
		if (r.ok)
			ui.toast({
				k: demo ? 'wa' : 'ok',
				title: 'Recordatorio enviado',
				desc: demo
					? `${c.nombre} · ${pesos(k.costoMensaje)} de WhatsApp`
					: `${c.nombre} · por mail, con el link a su carrito`
			});
	}
	/**
	 * Los tres recordatorios por mail (§62): se prenden y se apagan acá, y «Configurar» abre la demora
	 * de cada uno y el cupón del segundo o el tercero. La demo los guarda sin mandar mails.
	 */
	const rec = $derived(k.recordatorios);
	const ORDEN = ['Primero', 'Segundo', 'Tercero'];
	async function guardarRec(r: RecordatoriosCarritos, titulo: string, desc: string) {
		const res = await fetch('/panel/carritos-abandonados/recordatorios', {
			method: 'PUT',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify(r)
		}).catch(() => null);
		if (!res?.ok) {
			const b = (await res?.json().catch(() => null)) as { message?: string } | null;
			ui.toast({ k: 'err', title: 'No se guardó', desc: b?.message ?? 'Revisá la conexión.' });
			return false;
		}
		await invalidateAll();
		ui.toast({ k: 'ok', title: titulo, desc });
		return true;
	}
	const resumenPaso = (p: RecordatoriosCarritos['pasos'][number]) =>
		`${textoDemora(p.horas)} después${p.cupon ? ` · ${textoCupon(p.cupon)}` : ''}`;
	async function prender(on: boolean) {
		await guardarRec(
			{ ...rec, activo: on },
			on ? 'Recordatorios prendidos' : 'Recordatorios apagados',
			on ? 'Les llega un mail con el link a su carrito' : 'No sale ningún recordatorio automático'
		);
	}
	async function paso(i: number, on: boolean) {
		const pasos = rec.pasos.map((p, j) => (j === i ? { ...p, activo: on } : p));
		await guardarRec(
			{ ...rec, pasos },
			on ? `${ORDEN[i]} recordatorio prendido` : `${ORDEN[i]} recordatorio apagado`,
			resumenPaso(rec.pasos[i])
		);
	}

	let editando = $state(false);
	let borrador = $state<RecordatoriosCarritos | null>(null);
	const problema = $derived(borrador ? problemaRecordatorios(borrador) : null);
	function configurar() {
		borrador = structuredClone($state.snapshot(rec));
		editando = true;
	}
	async function guardarBorrador() {
		if (!borrador || problema) return;
		if (await guardarRec(borrador, 'Recordatorios guardados', 'Se aplican a los próximos carritos'))
			editando = false;
	}
</script>

<Seo
	title="Carritos abandonados · Panel"
	description="Carritos de tu tienda"
	path="/panel/carritos-abandonados"
	noindex
/>

<div class="pagina">
	<div class="cab" use:anim={{ t: 'rise' }}>
		<div class="c-txt">
			<h1>Carritos abandonados</h1>
			<div class="c-s">Gente que estuvo a un paso de comprarte</div>
		</div>
		<button type="button" class="primario" disabled={!pendientes.length} onclick={todos}
			>{#if demo}<Icon icon={I.phWhatsappLogoFill} size={15} />Recordar a todos · {pesos(
					pendientes.length * k.costoMensaje
				)}{:else}<Icon icon={I.phEnvelopeSimpleFill} size={15} />Recordar a todos por mail{/if}</button
		>
	</div>

	<div class="kpis" use:anim={{ t: 'rise', delay: 60 }}>
		<div class="oscura">
			<div class="o-in">
				<div class="o-t">En carritos abiertos</div>
				<div class="o-v">{pesos(k.monto)}</div>
				<div class="o-s">
					{k.abiertos}
					{k.abiertos === 1 ? 'carrito' : 'carritos'} · ticket promedio {pesos(ticket)}
				</div>
			</div>
		</div>
		<div class="tarjeta">
			<div class="t-t">Recuperado este mes</div>
			<div class="t-v">{pesos(k.recuperado)}</div>
			<div class="barra"><div style:width="{k.tasa}%"></div></div>
			<div class="t-s">{k.tasa}% de recuperación</div>
		</div>
		<div class="tarjeta autos">
			<div class="auto a-cab">
				<span class="a-t">Recordatorios por mail</span>
				<Interruptor label="Recordatorios por mail" checked={rec.activo} onchange={prender} />
			</div>
			{#each rec.pasos as p, i (i)}
				<div class="auto" class:apagado={!rec.activo}>
					<span>{ORDEN[i]} · {resumenPaso(p)}</span>
					<Interruptor
						label="{ORDEN[i]} recordatorio"
						checked={p.activo}
						disabled={!rec.activo}
						onchange={(on: boolean) => paso(i, on)}
					/>
				</div>
			{/each}
			<button type="button" class="configurar" onclick={configurar}
				><Icon icon={I.phTimerBold} size={14} />Configurar demoras y cupones</button
			>
		</div>
	</div>

	<div class="recup" use:anim={{ t: 'rise', delay: 90 }}>
		<Icon icon={I.phCheckCircleFill} size={16} />
		{#if k.recuperados.pedidos}
			<span
				>Este mes {k.recuperados.pedidos === 1
					? 'volvió 1 persona'
					: `volvieron ${k.recuperados.pedidos} personas`} después del recordatorio:
				<b>{pesos(k.recuperados.monto)}</b> en&nbsp;ventas.</span
			>
		{:else}
			<span
				>Cuando alguien compre después de un recordatorio, lo vas a ver acá con lo
				que&nbsp;cobraste.</span
			>
		{/if}
	</div>

	<div class="carritos" use:anim={{ t: 'rise', delay: 120 }}>
		{#each k.carritos as c (c.id)}
			<div class="carrito">
				{#if c.foto}<img class="av" src={cara(c.foto)} alt="" />{:else}<span class="av ini"
						>{iniciales(c.nombre)}</span
					>{/if}
				<span class="quien"
					><span class="n">{c.nombre}</span><span class="s"
						>{hace((Date.now() - Date.parse(c.creado)) / 86_400_000)} · desde {c.desde}</span
					></span
				>
				<span class="fotos"
					>{#each c.fotos as f, i (i)}{#if f}<img
								src={unsplash(f, 80, '&h=80')}
								alt=""
							/>{/if}{/each}</span
				>
				<span class="monto"
					><span class="v">{pesos(c.total)}</span><span class="q"
						>{c.cantidad} {c.cantidad === 1 ? 'producto' : 'productos'}</span
					></span
				>
				{#if c.recordado}
					<span class="hecho"><Icon icon={I.phCheckCircleFill} size={16} />Recordado</span>
				{:else}
					<button type="button" class="recordar" onclick={() => uno(c)}
						><Icon
							icon={demo ? I.logosWhatsappIcon : I.phEnvelopeSimpleFill}
							size={14}
						/>Recordar</button
					>
				{/if}
			</div>
		{:else}
			<div class="vacio">
				No hay carritos abandonados. Cuando alguien deje su mail en el checkout y no termine la
				compra, aparece acá.
			</div>
		{/each}
	</div>
</div>

<Modal bind:open={editando} titulo="Recordatorios del carrito" ancho={520}>
	{#if borrador}
		<div class="ed">
			<h2>Recordatorios del carrito</h2>
			<p class="ed-s">
				Salen desde que dejó el carrito. Si compra o se da de baja, no le llega&nbsp;ninguno más.
			</p>
			{#each borrador.pasos as p, i (i)}
				<div class="ed-paso">
					<div class="ed-cab">
						<b>{ORDEN[i]}</b>
						<Interruptor label="{ORDEN[i]} recordatorio" bind:checked={p.activo} />
					</div>
					<div class="ed-campos">
						<label class="campo"
							>Cuándo sale<select bind:value={p.horas}>
								{#each DEMORAS as h (h)}<option value={h}>{textoDemora(h)} después</option>{/each}
							</select></label
						>
						{#if i > 0}
							<label class="campo"
								>Cupón de un uso<select
									value={idCupon(p.cupon)}
									onchange={(e) => {
										const c = CUPONES.find((x) => x.id === e.currentTarget.value);
										p.cupon = c?.cupon ? { ...c.cupon } : null;
									}}
								>
									{#each CUPONES as c (c.id)}<option value={c.id}>{c.texto}</option>{/each}
									{#if idCupon(p.cupon) === 'otro'}<option value="otro"
											>{textoCupon(p.cupon)}</option
										>{/if}
								</select></label
							>
						{/if}
					</div>
				</div>
			{/each}
			<p class="ed-nota">
				<Icon icon={I.phTicketFill} size={14} />El cupón es único para cada carrito, vale 3 días y
				se aplica solo con el link del&nbsp;mail.
			</p>
			{#if problema}<p class="ed-error" role="alert">{problema}</p>{/if}
			<div class="ed-acc">
				<button type="button" class="sec" onclick={() => (editando = false)}>Cancelar</button>
				<button type="button" class="primario" disabled={!!problema} onclick={guardarBorrador}
					>Guardar</button
				>
			</div>
		</div>
	{/if}
</Modal>

<style>
	.pagina {
		display: flex;
		flex-direction: column;
		gap: 18px;
	}
	.cab {
		display: flex;
		align-items: end;
		gap: 14px;
		flex-wrap: wrap;
	}
	.c-txt {
		flex: 1;
		min-width: 220px;
	}
	h1 {
		margin: 0;
		font-size: 30px;
		letter-spacing: -0.035em;
		font-weight: 600;
	}
	.c-s {
		margin-top: 2px;
		font-size: 15px;
		color: var(--mu);
	}
	.primario {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 11px 20px;
		border-radius: 999px;
		border: 0;
		background: linear-gradient(180deg, #c24e25, #b0441d);
		color: #fff;
		font-size: 14px;
		font-weight: 600;
		white-space: nowrap;
		box-shadow:
			inset 0 1px 0 rgba(255, 255, 255, 0.3),
			0 10px 24px -10px rgba(210, 88, 47, 0.8);
	}
	.primario:disabled {
		background: var(--chip);
		color: var(--mu);
		box-shadow: none;
	}
	.kpis {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
		gap: 12px;
	}
	.oscura {
		position: relative;
		overflow: hidden;
		padding: 20px;
		border-radius: 22px;
		background: #2a140c;
		color: #fbf7f2;
		box-shadow: var(--shd);
	}
	.o-in {
		position: relative;
	}
	.o-t {
		font-size: 13px;
		color: #d9c8ba;
	}
	.o-v,
	.t-v {
		font-size: 34px;
		font-weight: 600;
		letter-spacing: -0.04em;
	}
	.o-s {
		font-size: 13px;
		color: #f5b796;
		margin-top: 2px;
	}
	.tarjeta {
		padding: 20px;
		border-radius: 22px;
		background: var(--sf);
		box-shadow: var(--shd);
	}
	.t-t {
		font-size: 13px;
		color: var(--mu);
	}
	.t-v {
		color: #237550;
	}
	.barra {
		margin-top: 8px;
		height: 8px;
		border-radius: 8px;
		background: var(--sf3);
		overflow: hidden;
	}
	.barra div {
		height: 100%;
		border-radius: 8px;
		background: #2f8f63;
	}
	.t-s {
		font-size: 12.5px;
		color: var(--mu);
		margin-top: 6px;
	}
	.autos {
		padding: 16px 18px;
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.a-t {
		font-size: 14px;
		font-weight: 600;
		margin-bottom: 4px;
	}
	.auto {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 7px 0;
	}
	.auto span {
		flex: 1;
		font-size: 13.5px;
		line-height: 1.3;
	}
	.a-cab {
		padding-top: 0;
	}
	.a-cab .a-t {
		margin: 0;
	}
	.auto.apagado span {
		color: var(--mu);
	}
	.configurar {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		align-self: flex-start;
		min-height: 36px;
		padding: 0;
		border: 0;
		background: none;
		color: var(--ac, #d2582f);
		font: 600 13px inherit;
		font-family: inherit;
		cursor: pointer;
	}
	.recup {
		display: flex;
		align-items: flex-start;
		gap: 8px;
		padding: 12px 16px;
		border-radius: 16px;
		background: rgba(47, 143, 99, 0.1);
		color: #237550;
		font-size: 13.5px;
		line-height: 1.4;
		text-wrap: pretty;
	}
	.recup :global(svg) {
		flex: none;
		margin-top: 1px;
	}
	.ed {
		display: flex;
		flex-direction: column;
		gap: 14px;
	}
	.ed h2 {
		margin: 0;
		font-size: 22px;
		font-weight: 600;
		letter-spacing: -0.02em;
	}
	.ed-s,
	.ed-nota {
		margin: -6px 0 0;
		font-size: 13.5px;
		color: var(--mu);
		text-wrap: pretty;
	}
	.ed-nota {
		margin: 0;
		display: flex;
		gap: 6px;
		align-items: flex-start;
	}
	.ed-nota :global(svg) {
		flex: none;
		margin-top: 2px;
	}
	.ed-paso {
		display: flex;
		flex-direction: column;
		gap: 10px;
		padding: 14px;
		border-radius: 16px;
		background: var(--sf2, var(--sf3));
	}
	.ed-cab {
		display: flex;
		align-items: center;
		justify-content: space-between;
		font-size: 14.5px;
	}
	.ed-campos {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 180px), 1fr));
		gap: 10px;
	}
	.campo {
		display: flex;
		flex-direction: column;
		gap: 6px;
		font-size: 12.5px;
		font-weight: 600;
		color: var(--mu);
	}
	.campo select {
		min-height: 44px;
		padding: 0 12px;
		border-radius: 12px;
		border: 1px solid var(--ln, rgba(0, 0, 0, 0.12));
		background: var(--sf);
		color: var(--tx);
		font: 500 14.5px inherit;
		font-family: inherit;
	}
	.ed-error {
		margin: 0;
		font-size: 13.5px;
		color: #b83a2e;
	}
	.ed-acc {
		display: flex;
		justify-content: flex-end;
		gap: 8px;
		flex-wrap: wrap;
	}
	.sec {
		min-height: 44px;
		padding: 0 18px;
		border-radius: 999px;
		border: 1px solid var(--ln, rgba(0, 0, 0, 0.12));
		background: transparent;
		color: var(--tx);
		font: 600 14px inherit;
		font-family: inherit;
		cursor: pointer;
	}
	.carritos {
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.carrito {
		display: flex;
		align-items: center;
		gap: 14px;
		padding: 14px 16px;
		border-radius: 20px;
		background: var(--sf);
		box-shadow: var(--shd);
		flex-wrap: wrap;
	}
	.av {
		width: 44px;
		height: 44px;
		border-radius: 50%;
		object-fit: cover;
		flex: none;
	}
	.ini {
		display: grid;
		place-items: center;
		background: var(--sf3);
		color: var(--mu);
		font-size: 13px;
		font-weight: 700;
	}
	.quien {
		flex: 1 1 180px;
		min-width: 0;
	}
	.n {
		display: block;
		font-size: 15px;
		font-weight: 600;
	}
	.s {
		display: block;
		font-size: 12.5px;
		color: var(--mu);
	}
	.fotos {
		display: flex;
		padding-left: 8px;
	}
	.fotos img {
		width: 40px;
		height: 40px;
		border-radius: 10px;
		object-fit: cover;
		margin-left: -8px;
		box-shadow: 0 0 0 2px var(--sf);
	}
	.monto {
		min-width: 90px;
		text-align: right;
	}
	.v {
		display: block;
		font-size: 16px;
		font-weight: 600;
		white-space: nowrap;
	}
	.q {
		display: block;
		font-size: 12px;
		color: var(--mu);
	}
	.recordar {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 9px 14px;
		border-radius: 999px;
		border: 0;
		background: #e9f8ee;
		color: #1f6f4a;
		font-size: 13px;
		font-weight: 600;
		white-space: nowrap;
	}
	.hecho {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 13px;
		font-weight: 600;
		color: #237550;
		white-space: nowrap;
	}
	.vacio {
		padding: 28px;
		text-align: center;
		color: var(--mu);
		font-size: 14px;
	}
</style>
