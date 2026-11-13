<script lang="ts">
	import Hoja from '#lib/components/Hoja.svelte';
	import Icon from '#lib/components/Icon.svelte';
	import MarcaIcono from '#lib/components/MarcaIcono.svelte';
	import * as I from '#lib/icons.ts';
	import { cara } from '#lib/landing/formato.ts';
	import { pesos, pesosCortos } from '../formato.ts';
	import { rutaDe } from '../navegacion.ts';
	import { CANALES, hace, iniciales, type Cliente } from './clientes.ts';

	/**
	 * Ficha de cliente. El handoff no la diseña: se arma con piezas de la hoja de pedido (encabezado,
	 * tarjetas, KPIs) y de la fila de la lista (barras por mes).
	 */
	let {
		open = $bindable(false),
		cliente,
		escribir,
		cupon,
		editar,
		suprimir
	}: {
		open?: boolean;
		cliente: Cliente | null;
		escribir: (c: Cliente) => void;
		cupon: (c: Cliente) => void;
		/** Corregir su nombre, email y WhatsApp. */
		editar: (c: Cliente) => void;
		/** El pedido de supresión: borrar sus datos. */
		suprimir: (c: Cliente) => void;
	} = $props();

	const SEG: Record<string, [string, string, string]> = {
		vip: ['VIP', 'rgba(201,138,18,.16)', '#8A5A08'],
		freq: ['Frecuente', 'rgba(47,143,99,.13)', '#2F8F63'],
		new: ['Nuevo', 'rgba(62,107,253,.12)', '#3E6BFD'],
		risk: ['En riesgo', 'rgba(210,88,47,.12)', '#C24E25'],
		sleep: ['Dormido', 'var(--chip)', 'var(--mu)']
	};
	const MESES = [
		'ene',
		'feb',
		'mar',
		'abr',
		'may',
		'jun',
		'jul',
		'ago',
		'sep',
		'oct',
		'nov',
		'dic'
	];
	const mesDe = (k: number) => MESES[(new Date().getMonth() - (11 - k) + 24) % 12];

	/** Pedidos importados de la tienda anterior (no son pedidos de Tu Local Web: sólo se muestran). */
	type Viejo = { id: string; plataforma: string; numero: string; fecha: string; total: number };
	const PLATAFORMA: Record<string, string> = {
		tiendanube: 'Tiendanube',
		shopify: 'Shopify',
		woocommerce: 'WooCommerce',
		mercadolibre: 'Mercado Libre',
		empretienda: 'Empretienda',
		csv: 'Planilla'
	};
	let historial = $state<Viejo[]>([]);
	$effect(() => {
		const id = open ? cliente?.id : null;
		historial = [];
		if (!id) return;
		let vivo = true;
		fetch(`/panel/clientes/${encodeURIComponent(id)}/historial`)
			.then((r) => (r.ok ? (r.json() as Promise<Viejo[]>) : []))
			.then((l) => {
				if (vivo) historial = l;
			})
			.catch(() => {});
		return () => {
			vivo = false;
		};
	});
	const fecha = (iso: string) =>
		new Date(iso).toLocaleDateString('es-AR', { day: 'numeric', month: 'short', year: 'numeric' });
</script>

<Hoja bind:open titulo={cliente?.nombre ?? 'Cliente'} ancho={520}>
	{#snippet cabecera()}
		{#if cliente}
			{@const s = SEG[cliente.seg]}
			<div class="cab">
				{#if cliente.foto}<img class="av" src={cara(cliente.foto)} alt="" />{:else}<span
						class="av ini">{iniciales(cliente.nombre)}</span
					>{/if}
				<div class="c-txt">
					<div class="c-fila">
						<span class="c-n">{cliente.nombre}</span><span
							class="chip"
							style:background={s[1]}
							style:color={s[2]}>{s[0]}</span
						>
					</div>
					<div class="c-s">
						{cliente.ciudad || 'Sin ciudad'} · llegó por {CANALES.find(
							(c) => c.k === cliente!.canal
						)?.t ?? cliente.canal}
					</div>
				</div>
				<button type="button" class="redondo" aria-label="Cerrar" onclick={() => (open = false)}
					><Icon icon={I.phXBold} size={15} /></button
				>
			</div>
		{/if}
	{/snippet}

	{#if cliente}
		<div class="cuerpo">
			<div class="kpis">
				<div class="kpi"><span>Pedidos</span><b>{cliente.pedidos}</b></div>
				<div class="kpi"><span>Gastó</span><b>{pesosCortos(cliente.gastado)}</b></div>
				<div class="kpi">
					<span>Ticket promedio</span><b
						>{cliente.pedidos ? pesos(cliente.gastado / cliente.pedidos) : '—'}</b
					>
				</div>
				<div class="kpi">
					<span>Última compra</span><b class:alerta={cliente.ultima > 75}
						>{cliente.pedidos ? hace(cliente.ultima) : '—'}</b
					>
				</div>
			</div>

			<div class="tarjeta">
				<div class="t">Compras de los últimos 12 meses</div>
				<div class="barras">
					{#each cliente.meses as m, k (k)}
						<div class="mes">
							<span
								class="b"
								style:height="{m ? 14 + m * 14 : 4}px"
								style:background={m ? '#D2582F' : 'var(--sf3)'}
							></span>
							<span class="m-l">{mesDe(k)}</span>
						</div>
					{/each}
				</div>
			</div>

			{#if historial.length}
				<div class="tarjeta">
					<div class="t">De tu tienda anterior</div>
					{#each historial as h (h.id)}
						<div class="dato">
							<span class="d-ico"><Icon icon={I.phClockCounterClockwiseFill} size={16} /></span>
							<span class="d-t"
								><span>{PLATAFORMA[h.plataforma] ?? h.plataforma} · {fecha(h.fecha)}</span><b
									>{h.numero ? `Pedido ${h.numero}` : 'Pedido'}</b
								></span
							>
							<b class="monto">{pesos(h.total)}</b>
						</div>
					{/each}
				</div>
			{/if}

			<div class="tarjeta">
				<div class="t">Contacto</div>
				<div class="dato">
					<MarcaIcono k="wa" size={30} radius={9} />
					<span class="d-t"><span>WhatsApp</span><b>{cliente.telefono ?? 'Sin cargar'}</b></span>
					{#if cliente.telefono}<button type="button" class="wa" onclick={() => escribir(cliente!)}
							>Escribir</button
						>{/if}
				</div>
				{#if cliente.cuit}
					<div class="dato">
						<span class="d-ico"><Icon icon={I.phIdentificationCardFill} size={16} /></span>
						<span class="d-t"><span>Facturación</span><b>Con CUIT · factura A</b></span>
					</div>
				{/if}
				{#if cliente.email}
					<div class="dato">
						<span class="d-ico"><Icon icon={I.phEnvelopeSimpleFill} size={16} /></span>
						<span class="d-t"><span>Email</span><b>{cliente.email}</b></span>
					</div>
				{/if}
				{#if cliente.etiquetas.length}
					<div class="etiquetas">
						{#each cliente.etiquetas as e (e)}<span><Icon icon={I.phTagFill} size={12} />{e}</span
							>{/each}
					</div>
				{/if}
			</div>

			<div class="tarjeta">
				<div class="t">Sus datos</div>
				<div class="datos-acc">
					<button type="button" class="chico" onclick={() => editar(cliente!)}
						><Icon icon={I.phPencilSimpleFill} size={14} />Corregir</button
					>
					<button type="button" class="chico peligro" onclick={() => suprimir(cliente!)}
						><Icon icon={I.phTrashFill} size={14} />Borrar sus datos</button
					>
				</div>
			</div>
		</div>
	{/if}

	{#snippet pie()}
		{#if cliente}
			<div class="pie">
				<!-- eslint-disable svelte/no-navigation-without-resolve -- ruta del panel con query -->
				<a
					class="borde"
					href="{rutaDe('Pedidos')}?q={encodeURIComponent(cliente.nombre)}"
					onclick={() => (open = false)}>Ver pedidos</a
				>
				<!-- eslint-enable svelte/no-navigation-without-resolve -->
				<button
					type="button"
					class="principal"
					disabled={cliente.cupon}
					onclick={() => cupon(cliente!)}
					>{cliente.cupon ? 'Cupón enviado' : 'Mandar cupón del 10%'}</button
				>
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
		font-size: 15px;
		font-weight: 700;
		color: var(--mu);
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
		font-size: 21px;
		font-weight: 600;
		letter-spacing: -0.03em;
	}
	.chip {
		font-size: 12px;
		font-weight: 700;
		padding: 3px 9px;
		border-radius: 999px;
		white-space: nowrap;
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
	.kpis {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 10px;
	}
	.kpi {
		padding: 12px 14px;
		border-radius: 16px;
		background: var(--sf);
		box-shadow: var(--sh);
	}
	.kpi span {
		display: block;
		font-size: 12.5px;
		color: var(--mu);
	}
	.kpi b {
		display: block;
		margin-top: 4px;
		font-size: 19px;
		font-weight: 600;
		letter-spacing: -0.02em;
	}
	.kpi b.alerta {
		color: #b0441d;
	}
	.tarjeta {
		padding: 14px 16px;
		border-radius: 18px;
		background: var(--sf);
		box-shadow: var(--sh);
	}
	.t {
		font-size: 13px;
		font-weight: 600;
		color: var(--mu);
	}
	.barras {
		margin-top: 12px;
		display: flex;
		align-items: end;
		gap: 4px;
		height: 72px;
	}
	.mes {
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 4px;
	}
	.b {
		width: 100%;
		border-radius: 3px;
	}
	.m-l {
		font-size: 10.5px;
		color: var(--fa);
	}
	.dato {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 10px 0 0;
	}
	.d-ico {
		width: 30px;
		height: 30px;
		border-radius: 9px;
		background: var(--sf3);
		display: grid;
		place-items: center;
		color: var(--mu);
	}
	.d-t {
		flex: 1;
		min-width: 0;
	}
	.d-t span {
		display: block;
		font-size: 12px;
		color: var(--mu);
	}
	.d-t b {
		display: block;
		font-size: 14px;
		font-weight: 600;
	}
	.monto {
		font-size: 14px;
		font-weight: 600;
		white-space: nowrap;
	}
	.wa {
		border: 0;
		font-size: 13px;
		font-weight: 600;
		padding: 8px 12px;
		border-radius: 999px;
		background: #e9f8ee;
		color: #1f6f4a;
	}
	.datos-acc {
		margin-top: 10px;
		display: flex;
		gap: 8px;
		flex-wrap: wrap;
	}
	.chico {
		min-height: 44px;
		display: inline-flex;
		align-items: center;
		gap: 6px;
		padding: 0 14px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: none;
		color: var(--tx);
		font-size: 13.5px;
		font-weight: 600;
	}
	.chico.peligro {
		color: #b83a2e;
	}
	.etiquetas {
		margin-top: 12px;
		display: flex;
		gap: 6px;
		flex-wrap: wrap;
	}
	.etiquetas span {
		display: flex;
		align-items: center;
		gap: 5px;
		font-size: 12.5px;
		font-weight: 600;
		padding: 4px 10px;
		border-radius: 999px;
		background: rgba(106, 79, 160, 0.12);
		color: #6a4fa0;
	}
	.pie {
		width: 100%;
		display: flex;
		gap: 8px;
	}
	.borde {
		padding: 13px 18px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		color: var(--tx);
		font-size: 14px;
		font-weight: 600;
		white-space: nowrap;
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
	}
	.principal:disabled {
		background: var(--chip);
		color: var(--mu);
	}
</style>
