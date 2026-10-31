<script lang="ts">
	import Interruptor from '#lib/components/Interruptor.svelte';
	import {
		aLocal,
		deLocal,
		errorLanzamiento,
		SIN_LANZAMIENTO,
		type ConfigLanzamiento,
		type LanzamientoProducto
	} from '#lib/tienda/lanzamiento.ts';
	import { pesos } from '../formato.ts';

	/**
	 * Las secciones «Lanzamiento», «Encargo con seña» y «Oferta» del editor del producto (§52), con las
	 * piezas del editor (tarjeta, precio, interruptor). Se guardan aparte del producto, después de él
	 * (un producto nuevo recién tiene id al crearse). Las fechas van en hora de Argentina.
	 */
	let {
		id,
		abierto,
		precio,
		conPrecios
	}: {
		/** El producto (null: el alta, todavía sin id). */
		id: string | null;
		abierto: boolean;
		/** El precio de lista que se está editando (la oferta tiene que ser menor). */
		precio: number;
		/** Quien no tiene Catálogo ve los números pero no cambia precios ni señas. */
		conPrecios: boolean;
	} = $props();

	let antes = $state<LanzamientoProducto>({
		...SIN_LANZAMIENTO,
		avisos: 0,
		avisados: 0,
		reservas: 0,
		encargos: 0
	});
	let lanzar = $state(false);
	let fecha = $state('');
	let reservar = $state(false);
	let reserva = $state(30);
	let encargar = $state(false);
	let encargo = $state(50);
	let ofertar = $state(false);
	let oferta = $state(0);
	let conFin = $state(false);
	let fin = $state('');

	function cargar(l: LanzamientoProducto) {
		antes = l;
		lanzar = !!l.lanzamiento;
		fecha = aLocal(l.lanzamiento);
		reservar = l.reserva !== null;
		reserva = l.reserva ?? 30;
		encargar = l.encargo !== null;
		encargo = l.encargo ?? 50;
		ofertar = l.ofertaPrecio !== null;
		oferta = l.ofertaPrecio ?? 0;
		conFin = !!l.ofertaHasta;
		fin = aLocal(l.ofertaHasta);
	}

	$effect(() => {
		if (!abierto) return;
		cargar({ ...SIN_LANZAMIENTO, avisos: 0, avisados: 0, reservas: 0, encargos: 0 });
		if (!id) return;
		const pedido = id;
		void fetch(`/panel/productos/${encodeURIComponent(pedido)}/lanzamiento`).then(async (r) => {
			if (r.ok && pedido === id) cargar((await r.json()) as LanzamientoProducto);
		});
	});

	const numero = (s: string) => Math.max(0, Number(s.replace(/\D/g, '')) || 0);
	const pct = (s: string) => Math.min(100, Math.max(1, numero(s) || 1));

	/** Lo que se va a guardar, con las mismas reglas que el servidor. */
	function config(): ConfigLanzamiento {
		return {
			lanzamiento: lanzar ? deLocal(fecha) : null,
			reserva: lanzar && reservar ? reserva : null,
			encargo: encargar ? encargo : null,
			ofertaPrecio: ofertar && oferta ? oferta : null,
			ofertaHasta: ofertar && conFin ? deLocal(fin) : null
		};
	}
	const igual = (a: ConfigLanzamiento, b: ConfigLanzamiento) =>
		(Object.keys(SIN_LANZAMIENTO) as (keyof ConfigLanzamiento)[]).every((k) =>
			typeof a[k] === 'string' && typeof b[k] === 'string'
				? Date.parse(a[k] as string) === Date.parse(b[k] as string)
				: a[k] === b[k]
		);

	/** Revisa antes de guardar el producto: el error para mostrar, o null. */
	export function revisar(): string | null {
		if (!conPrecios) return null;
		if (lanzar && !deLocal(fecha)) return 'Poné la fecha y la hora del lanzamiento.';
		if (ofertar && conFin && !deLocal(fin)) return 'Poné hasta cuándo dura la oferta.';
		return errorLanzamiento(config(), antes, precio);
	}

	/** Guarda (si cambió algo) con el id del producto ya guardado. El error para mostrar, o null. */
	export async function guardar(producto: string): Promise<string | null> {
		const c = config();
		if (!conPrecios || igual(c, antes)) return null;
		const r = await fetch(`/panel/productos/${encodeURIComponent(producto)}/lanzamiento`, {
			method: 'PUT',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify(c)
		});
		if (!r.ok)
			return (
				((await r.json().catch(() => null)) as { message?: string } | null)?.message ??
				'No pudimos guardar el lanzamiento.'
			);
		cargar((await r.json()) as LanzamientoProducto);
		return null;
	}
</script>

<div class="tarjeta">
	<div class="canal">
		<span class="t">Lanzamiento</span>
		<Interruptor label="Programar el lanzamiento" bind:checked={lanzar} disabled={!conPrecios} />
	</div>
	{#if lanzar}
		<label class="campo"
			>Sale a la venta (hora de Argentina)<input
				type="datetime-local"
				bind:value={fecha}
				disabled={!conPrecios}
			/></label
		>
		<p class="nota">
			Hasta esa hora se ve «Próximamente» con la cuenta regresiva y «Avisame cuando salga». A la
			hora, sale a la venta solo y les escribimos a quienes pidieron aviso.
		</p>
		<div class="canal">
			<span>Reservar con seña</span>
			<Interruptor label="Reservar con seña" bind:checked={reservar} disabled={!conPrecios} />
		</div>
		{#if reservar}
			<div class="precios">
				<div class="precio">
					<div class="p-t">Seña</div>
					<input
						class="p-in"
						inputmode="numeric"
						aria-label="Porcentaje de seña de la reserva"
						readonly={!conPrecios}
						value="{reserva} %"
						onchange={(e) => (reserva = pct(e.currentTarget.value))}
					/>
					<div class="p-n">
						{precio
							? `${pesos(Math.max(1, Math.round((precio * reserva) / 100)))} por unidad`
							: 'Del precio'}
					</div>
				</div>
			</div>
			<p class="nota">El cupo es el stock: cada reserva lo descuenta.</p>
		{/if}
	{/if}
	{#if id && (antes.avisos || antes.reservas || antes.lanzamiento)}
		<p class="numeros">
			<b>{antes.avisos}</b>
			{antes.avisos === 1 ? 'pidió' : 'pidieron'} aviso{#if antes.avisados}&nbsp;({antes.avisados}
				{antes.avisados === 1 ? 'avisado' : 'avisados'}){/if} · <b>{antes.reservas}</b>
			{antes.reservas === 1 ? 'reservó' : 'reservaron'}
		</p>
	{/if}
</div>

<div class="tarjeta">
	<div class="canal">
		<span class="t">Encargo con seña</span>
		<Interruptor label="Encargo con seña" bind:checked={encargar} disabled={!conPrecios} />
	</div>
	{#if encargar}
		<div class="precios">
			<div class="precio">
				<div class="p-t">Seña</div>
				<input
					class="p-in"
					inputmode="numeric"
					aria-label="Porcentaje de seña del encargo"
					readonly={!conPrecios}
					value="{encargo} %"
					onchange={(e) => (encargo = pct(e.currentTarget.value))}
				/>
				<div class="p-n">El resto, al retirarlo</div>
			</div>
		</div>
		<p class="nota">
			Se encarga pagando la seña con tu Mercado Pago. Con talles o colores, en Quebracho se arma
			paso a paso.
		</p>
		{#if id && antes.encargos}<p class="numeros">
				<b>{antes.encargos}</b>
				{antes.encargos === 1 ? 'encargo' : 'encargos'} con seña
			</p>{/if}
	{/if}
</div>

<div class="tarjeta">
	<div class="canal">
		<span class="t">Oferta</span>
		<Interruptor label="Precio de oferta" bind:checked={ofertar} disabled={!conPrecios} />
	</div>
	{#if ofertar}
		<div class="precios">
			<div class="precio">
				<div class="p-t">Precio de oferta</div>
				<input
					class="p-in"
					inputmode="numeric"
					aria-label="Precio de oferta"
					readonly={!conPrecios}
					placeholder="$ 0"
					value={oferta ? pesos(oferta) : ''}
					onchange={(e) => (oferta = numero(e.currentTarget.value))}
				/>
				<div class="p-n">
					{oferta && precio > oferta
						? `${Math.round((1 - oferta / precio) * 100)} % off sobre ${pesos(precio)}`
						: `Menor que ${pesos(precio)}`}
				</div>
			</div>
		</div>
		<div class="canal">
			<span>Con fecha de fin</span>
			<Interruptor label="Oferta con fecha de fin" bind:checked={conFin} disabled={!conPrecios} />
		</div>
		{#if conFin}
			<label class="campo"
				>Termina (hora de Argentina)<input
					type="datetime-local"
					bind:value={fin}
					disabled={!conPrecios}
				/></label
			>
			<p class="nota">Las plantillas con cuenta regresiva la muestran hasta esa hora.</p>
		{/if}
	{/if}
</div>

<style>
	.tarjeta {
		padding: 14px;
		border-radius: 18px;
		background: var(--sf);
		box-shadow: var(--sh);
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.t {
		font-size: 14px;
		font-weight: 600;
	}
	.canal {
		display: flex;
		align-items: center;
		gap: 10px;
		min-height: 44px;
	}
	.canal span {
		flex: 1;
		font-size: 14px;
	}
	.campo {
		display: flex;
		flex-direction: column;
		gap: 4px;
		font-size: 12.5px;
		font-weight: 600;
		color: var(--mu);
	}
	.campo input {
		min-height: 44px;
		padding: 11px 12px;
		border-radius: 12px;
		border: 1px solid var(--ln2);
		background: var(--bg);
		font-size: 15px;
		font-weight: 500;
		color: var(--tx);
		outline: none;
	}
	.precios {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
		gap: 8px;
	}
	.precio {
		padding: 10px 12px;
		border-radius: 12px;
		background: var(--sf3);
	}
	.p-t {
		font-size: 12px;
		color: var(--mu);
	}
	.p-in {
		width: 100%;
		font-size: 18px;
		font-weight: 600;
		border: 0;
		border-bottom: 1.5px dashed var(--ln2);
		background: transparent;
		color: var(--tx);
		padding: 0;
		outline: none;
	}
	.p-in:focus {
		border-bottom-color: #d2582f;
	}
	.p-n {
		font-size: 11.5px;
		color: var(--fa);
	}
	.nota,
	.numeros {
		margin: 0;
		font-size: 12.5px;
		line-height: 1.45;
		color: var(--mu);
		text-wrap: pretty;
	}
	.numeros b {
		color: var(--tx);
	}
</style>
