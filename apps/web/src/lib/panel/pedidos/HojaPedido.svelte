<script lang="ts">
	import { resolve } from '$app/paths';
	import ChipEstado from '#lib/components/ChipEstado.svelte';
	import Hoja from '#lib/components/Hoja.svelte';
	import Icon from '#lib/components/Icon.svelte';
	import MarcaIcono from '#lib/components/MarcaIcono.svelte';
	import EnvioCorreo from './EnvioCorreo.svelte';
	import * as I from '#lib/icons.ts';
	import { cara, unsplash } from '#lib/landing/formato.ts';
	import type { components } from '@mostrador/contrato/esquema';
	import { pesos } from '../formato.ts';
	import { getUi } from '#lib/ui.svelte.ts';
	import { whatsappDe, type Contexto } from './acciones.ts';
	import { NOMBRE_CANAL, NOMBRE_ENVIO, NOMBRE_PAGO, siguienteDe, textoEstado } from './nombres.ts';

	type Detalle = components['schemas']['PedidoDetalle'];

	/** Hoja del pedido (handoff `ovSheet`): cliente, nota de Tincho, productos, pago, factura, envío e historial. */
	let {
		open = $bindable(false),
		pedido,
		avanzar,
		imprimir,
		contexto
	}: {
		open?: boolean;
		pedido: Detalle | null;
		avanzar: (p: Detalle) => void;
		imprimir: (p: Detalle) => void;
		/** Dónde se retira y la dirección de la tienda, para el mensaje de WhatsApp. */
		contexto: Contexto;
	} = $props();

	const TZ = 'America/Argentina/Buenos_Aires';
	const cuando = (iso: string) => {
		const d = new Date(iso);
		const dia = new Intl.DateTimeFormat('en-CA', { timeZone: TZ }).format(d);
		const hoy = new Intl.DateTimeFormat('en-CA', { timeZone: TZ }).format(new Date());
		const hora = new Intl.DateTimeFormat('es-AR', {
			timeZone: TZ,
			hour: '2-digit',
			minute: '2-digit'
		}).format(d);
		return dia === hoy
			? `Hoy ${hora}`
			: new Intl.DateTimeFormat('es-AR', { timeZone: TZ, day: 'numeric', month: 'short' })
					.format(d)
					.replace('.', '') + ` ${hora}`;
	};
	const horaCorta = (iso: string) =>
		new Intl.DateTimeFormat('es-AR', {
			timeZone: TZ,
			hour: '2-digit',
			minute: '2-digit',
			hourCycle: 'h23'
		}).format(new Date(iso));
	const anio = (iso: string) => new Date(iso).getUTCFullYear();
	const llega = (fecha: string) =>
		new Intl.DateTimeFormat('es-AR', { timeZone: 'UTC', weekday: 'long', day: 'numeric' }).format(
			new Date(fecha)
		);
	const iniciales = (n: string) =>
		n
			.split(' ')
			.map((w) => w[0])
			.slice(0, 2)
			.join('');
	// El WhatsApp sale con el mensaje del estado (listo para retirar, en camino, la reseña…).
	const wa = $derived(
		pedido
			? whatsappDe(
					{ ...pedido, telefono: pedido.comprador.telefono },
					{ ...contexto, seguimiento: pedido.seguimiento }
				)
			: null
	);
	const siguiente = $derived(pedido ? siguienteDe(pedido) : undefined);

	// La seña (§52): «Seña pagada · falta $X» y «Marcar como saldado» (con Deshacer).
	const ui = getUi();
	let saldado = $state<boolean | null>(null);
	$effect(() => {
		void pedido?.numero;
		saldado = null;
	});
	const sena = $derived(
		pedido?.sena ? { ...pedido.sena, saldado: saldado ?? pedido.sena.saldado } : null
	);
	async function saldar(valor: boolean) {
		if (!pedido) return;
		const n = pedido.numero;
		const r = await fetch(`/panel/pedidos/${n}`, {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ saldado: valor })
		});
		if (!r.ok) {
			ui.toast({ k: 'warn', title: 'No se pudo', desc: 'El pedido cambió. Actualizá la lista.' });
			return;
		}
		saldado = valor;
		if (valor)
			ui.toast({
				k: 'ok',
				title: `Pedido #${n} saldado`,
				desc: 'Quedó cobrado completo',
				action: 'Deshacer',
				onAction: () => saldar(false)
			});
	}
</script>

<Hoja bind:open titulo={pedido ? `Pedido #${pedido.numero}` : 'Pedido'} ancho={520}>
	{#snippet cabecera()}
		{#if pedido}
			<div class="cab">
				<div class="c-txt">
					<div class="c-fila">
						<span class="c-n">Pedido #{pedido.numero}</span><ChipEstado
							estado={pedido.estado}
							texto={textoEstado(pedido)}
						/>
					</div>
					<div class="c-s">
						{cuando(pedido.creado)} · por {NOMBRE_CANAL[pedido.canal] ?? pedido.canal}
					</div>
				</div>
				<a
					class="redondo"
					title="Imprimir la tarjeta del paquete"
					aria-label="Imprimir la tarjeta del paquete"
					href="{resolve('/panel/pedidos/imprimir')}?n={pedido.numero}"
					target="_blank"><Icon icon={I.phPrinter} size={17} /></a
				>
				<button
					type="button"
					class="redondo"
					title="Cerrar"
					aria-label="Cerrar"
					onclick={() => (open = false)}><Icon icon={I.phXBold} size={15} /></button
				>
			</div>
		{/if}
	{/snippet}

	{#if pedido}
		<div class="cuerpo">
			<div class="tarjeta fila">
				{#if pedido.comprador.foto}<img
						class="av"
						src={cara(pedido.comprador.foto)}
						alt=""
					/>{:else}<span class="av ini">{iniciales(pedido.comprador.nombre)}</span>{/if}
				<div class="f-txt">
					<div class="f-n">{pedido.comprador.nombre}</div>
					<div class="f-s">
						<!-- Cada dato entero en su renglón: el «·» nunca arranca uno. -->
						<span class="nw">Cliente desde {anio(pedido.comprador.desde)} ·</span>
						<span class="nw"
							>{pedido.comprador.pedidos}
							{pedido.comprador.pedidos === 1 ? 'pedido' : 'pedidos'} ·</span
						>
						<span class="nw">{pesos(pedido.comprador.gastado)}</span>
					</div>
				</div>
				{#if wa}
					<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- link externo a WhatsApp -->
					<a class="wa" href={wa} target="_blank" rel="noopener noreferrer"
						><Icon icon={I.logosWhatsappIcon} size={14} />Escribir</a
					>
				{/if}
			</div>

			{#if pedido.nota}
				<div class="nota">
					<Icon icon={I.phSparkleFill} size={16} style="color:#b0441d;margin-top:2px;flex:none" />
					<div>
						{#if pedido.hornero}<b>Tincho:</b>
						{/if}{pedido.nota}
					</div>
				</div>
			{/if}

			<div class="tarjeta">
				<div class="t">Productos</div>
				{#each pedido.items as it, i (i)}
					<div class="item">
						{#if it.foto}<img src={unsplash(it.foto, 96, '&h=96')} alt="" />{:else}<span
								class="i-ico"><Icon icon={I.phPackageFill} size={19} /></span
							>{/if}
						<span class="i-txt"
							><span class="i-n">{it.nombre}</span><span class="i-q"
								>{#if it.variante}{it.variante} ·
								{/if}{it.cantidad} × {pesos(it.precio)}</span
							></span
						>
						<span class="i-t">{pesos(it.cantidad * it.precio)}</span>
					</div>
				{/each}
				<div class="suma"><span>Subtotal</span><span>{pesos(pedido.subtotal)}</span></div>
				<div class="suma">
					<span>Envío</span>
					{#if pedido.costoEnvio}<span>{pesos(pedido.costoEnvio)}</span>{:else}<span class="gratis"
							>Gratis</span
						>{/if}
				</div>
				{#each pedido.descuentos as d, i (i)}
					<div class="suma desc"><span>{d.concepto}</span><span>−{pesos(d.monto)}</span></div>
				{/each}
				<div class="total"><span>Total</span><span>{pesos(pedido.total)}</span></div>
			</div>

			<div class="dos" class:uno={!pedido.factura}>
				<div class="tarjeta">
					<div class="t">Pago</div>
					<div class="medio">
						<MarcaIcono
							k={pedido.pago === 'tarjeta' ? 'visa' : pedido.pago}
							size={34}
							radius={10}
						/>
						<span
							><span class="m-n">{NOMBRE_PAGO[pedido.pago] ?? pedido.pago}</span><span class="m-s"
								>{#if sena?.saldado}Seña y saldo cobrados{:else if sena?.pagada}Seña pagada · falta {pesos(
										sena.falta
									)}{:else if sena}Esperando la seña de {pesos(
										sena.monto
									)}{:else}{pedido.estadoPago}{/if}</span
							></span
						>
					</div>
					{#if sena?.pagada && !sena.saldado && pedido.estado !== 'Cancelado'}
						<button type="button" class="borde saldar" onclick={() => saldar(true)}
							>Marcar como saldado</button
						>
					{/if}
				</div>
				<!-- Sin factura no se muestra nada: «se emite al cobrar» no es cierto sin ARCA conectado. -->
				{#if pedido.factura}
					<div class="tarjeta">
						<div class="t">Factura</div>
						<div class="medio">
							<MarcaIcono k="arca" size={34} radius={10} />
							<span
								><span class="m-n">{pedido.factura}</span><span class="m-s ok">Emitida</span></span
							>
						</div>
					</div>
				{/if}
			</div>

			<div class="tarjeta fila">
				<MarcaIcono k={pedido.envio} size={40} radius={12} />
				<span class="f-txt">
					<span class="f-n">{NOMBRE_ENVIO[pedido.envio] ?? pedido.envio}</span>
					<span class="f-s">
						{#if pedido.seguimiento}Seguimiento {pedido.seguimiento}{#if pedido.llega}
								· llega el {llega(pedido.llega)}{/if}
						{:else if pedido.envio === 'retiro'}Retira en el local
						{:else}{pedido.ciudad}{/if}
					</span>
				</span>
				{#if pedido.envio !== 'retiro'}
					<button type="button" class="borde" onclick={() => imprimir(pedido!)}>Etiqueta</button>
				{/if}
			</div>

			{#if pedido.envio === 'correo'}
				<div class="tarjeta">
					<div class="t">Correo Argentino</div>
					<EnvioCorreo numero={pedido.numero} entrega={pedido.entrega ?? null} />
				</div>
			{/if}

			<div class="tarjeta">
				<div class="t">Historial</div>
				<div class="hist">
					{#each pedido.historial as h, i (i)}
						<div class="h">
							{#if h.hecho}<span class="h-ok"><Icon icon={I.phCheckBold} size={10} /></span
								>{:else}<span class="h-no"></span>{/if}
							<span class="h-t">{h.texto}</span><span class="h-c"
								>{h.hecho && h.cuando ? horaCorta(h.cuando) : 'pendiente'}</span
							>
						</div>
					{/each}
				</div>
			</div>
		</div>
	{:else}
		<div class="cargando"><span></span><span></span><span></span></div>
	{/if}

	{#snippet pie()}
		{#if pedido}
			<div class="pie">
				<!-- Un pedido nuevo para el mismo cliente, con lo mismo: se revisa en Crear pedido. -->
				<a
					class="secundario"
					href="{resolve('/panel/pedidos/nuevo')}?desde={pedido.numero}"
					data-sveltekit-preload-data="off"
					><Icon icon={I.phArrowClockwiseBold} size={15} />Volver a vender</a
				>
				{#if siguiente}
					<button type="button" class="principal" onclick={() => avanzar(pedido!)}
						>{siguiente}</button
					>
				{/if}
			</div>
		{/if}
	{/snippet}
</Hoja>

<style>
	.cab {
		display: flex;
		align-items: center;
		gap: 12px;
	}
	.c-txt {
		flex: 1;
		min-width: 0;
	}
	.c-fila {
		display: flex;
		align-items: center;
		gap: 10px;
		flex-wrap: wrap;
	}
	.c-n {
		font-size: 22px;
		font-weight: 600;
		letter-spacing: -0.03em;
	}
	.c-s {
		font-size: 13px;
		color: var(--mu);
		margin-top: 2px;
	}
	.redondo {
		width: 36px;
		height: 36px;
		flex: none;
		border: 0;
		border-radius: 50%;
		background: var(--chip);
		color: var(--tx);
		display: grid;
		place-items: center;
	}
	.cuerpo {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}
	.tarjeta {
		padding: 14px 16px;
		border-radius: 18px;
		background: var(--sf);
		box-shadow: var(--sh);
	}
	.fila {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 14px;
	}
	.av {
		width: 48px;
		height: 48px;
		flex: none;
		border-radius: 50%;
		object-fit: cover;
	}
	.ini {
		background: var(--sf3);
		display: grid;
		place-items: center;
		font-size: 14px;
		font-weight: 600;
		color: var(--mu);
	}
	.f-txt {
		flex: 1;
		min-width: 0;
	}
	.f-n {
		display: block;
		font-size: 15.5px;
		font-weight: 600;
	}
	.fila .f-n:only-child,
	.f-txt .f-n {
		font-size: 14px;
	}
	.fila:first-child .f-n {
		font-size: 15.5px;
	}
	.f-s {
		display: block;
		font-size: 12.5px;
		color: var(--mu);
	}
	.nw {
		white-space: nowrap;
	}
	.wa {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 13px;
		font-weight: 600;
		padding: 8px 12px;
		border-radius: 999px;
		background: #e9f8ee;
		color: #1f6f4a;
		white-space: nowrap;
	}
	.nota {
		display: flex;
		gap: 10px;
		padding: 12px 14px;
		border-radius: 16px;
		background: rgba(210, 88, 47, 0.1);
		border: 1px solid rgba(210, 88, 47, 0.22);
		font-size: 13.5px;
		line-height: 1.45;
	}
	.t {
		font-size: 13px;
		font-weight: 600;
		color: var(--mu);
	}
	.item {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 10px 0;
		border-bottom: 1px solid var(--ln);
	}
	.item img,
	.i-ico {
		width: 40px;
		height: 40px;
		flex: none;
		border-radius: 11px;
		object-fit: cover;
	}
	.i-ico {
		background: var(--sf3);
		display: grid;
		place-items: center;
		color: #b0441d;
	}
	.i-txt {
		flex: 1;
		min-width: 0;
	}
	.i-n {
		display: block;
		font-size: 14px;
		font-weight: 600;
	}
	.i-q {
		display: block;
		font-size: 12.5px;
		color: var(--mu);
	}
	.i-t {
		font-size: 14px;
		font-weight: 600;
		white-space: nowrap;
	}
	.suma {
		display: flex;
		justify-content: space-between;
		font-size: 13.5px;
		color: var(--mu);
		padding-top: 10px;
	}
	.suma + .suma {
		padding-top: 4px;
	}
	.desc span:last-child {
		color: #237550;
	}
	.gratis {
		color: #237550;
		font-weight: 600;
	}
	.total {
		display: flex;
		justify-content: space-between;
		font-size: 17px;
		font-weight: 600;
		padding-top: 8px;
	}
	.dos {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 12px;
	}
	.dos.uno {
		grid-template-columns: 1fr;
	}
	.dos .tarjeta {
		padding: 14px;
	}
	.medio {
		margin-top: 10px;
		display: flex;
		align-items: center;
		gap: 10px;
	}
	.medio :global(.marca) {
		box-shadow: var(--sh);
	}
	.m-n {
		display: block;
		font-size: 14px;
		font-weight: 600;
	}
	.m-s {
		display: block;
		font-size: 12.5px;
		color: var(--fa);
	}
	.m-s.ok {
		color: #237550;
		font-weight: 600;
	}
	.borde {
		border: 1px solid var(--ln2);
		background: transparent;
		font-size: 13px;
		font-weight: 600;
		padding: 8px 12px;
		border-radius: 999px;
		color: var(--tx);
		white-space: nowrap;
	}
	.saldar {
		margin-top: 10px;
		min-height: 44px;
		width: 100%;
	}
	.hist {
		margin-top: 10px;
		display: flex;
		flex-direction: column;
	}
	.h {
		display: flex;
		gap: 12px;
		align-items: flex-start;
		padding-bottom: 12px;
	}
	.h-ok {
		width: 18px;
		height: 18px;
		flex: none;
		border-radius: 50%;
		background: #2f8f63;
		color: #fff;
		display: grid;
		place-items: center;
		margin-top: 1px;
	}
	.h-no {
		width: 18px;
		height: 18px;
		flex: none;
		border-radius: 50%;
		border: 2px dashed var(--ln2);
		margin-top: 1px;
	}
	.h-t {
		flex: 1;
		font-size: 13.5px;
	}
	.h-c {
		font-size: 12.5px;
		color: var(--fa);
		white-space: nowrap;
	}
	.pie {
		width: 100%;
		display: flex;
		gap: 8px;
	}
	.secundario {
		flex: none;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 7px;
		padding: 12px 16px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		color: var(--tx);
		font-size: 14px;
		font-weight: 600;
		white-space: nowrap;
		text-decoration: none;
	}
	.pie .secundario:only-child {
		flex: 1;
	}
	.principal {
		flex: 1;
		padding: 13px;
		border-radius: 999px;
		border: 0;
		background: linear-gradient(180deg, #c24e25, #b0441d);
		color: #fff;
		font-size: 14.5px;
		font-weight: 600;
		box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.3);
	}
	.cargando {
		display: flex;
		gap: 6px;
		justify-content: center;
		padding: 40px;
	}
	.cargando span {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: var(--fa);
		animation: puntos 1.1s infinite;
	}
	.cargando span:nth-child(2) {
		animation-delay: 0.15s;
	}
	.cargando span:nth-child(3) {
		animation-delay: 0.3s;
	}
	@keyframes puntos {
		50% {
			opacity: 0.25;
		}
	}
	@media (max-width: 420px) {
		.dos {
			grid-template-columns: 1fr;
		}
	}
</style>
