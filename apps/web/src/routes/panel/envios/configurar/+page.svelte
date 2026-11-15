<script lang="ts">
	import { untrack } from 'svelte';
	import { invalidateAll } from '$app/navigation';
	import { resolve } from '$app/paths';
	import Icon from '#lib/components/Icon.svelte';
	import Interruptor from '#lib/components/Interruptor.svelte';
	import MarcaIcono from '#lib/components/MarcaIcono.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { problemaEntregas, TOPES_ENTREGAS, type Entregas } from '#lib/panel/entregas.ts';
	import { pesos } from '#lib/panel/formato.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/**
	 * Envíos y retiro (Configuración): cómo le llega a quien compra. El handoff no la diseña: se arma
	 * con las piezas de Configuración (tarjetas, opciones con Interruptor y campos) y de la tarjeta para
	 * paquetes (cabecera con Guardar). Correo Argentino se conecta en Integraciones; el envío gratis
	 * (la misma promoción automática de Cupones) se prende acá con el monto de la tienda.
	 */
	let { data } = $props();
	const ui = getUi();

	const copia = (e: Entregas): Entregas => ({
		retiro: e.retiro,
		retiroDireccion: e.retiroDireccion,
		retiroHorario: e.retiroHorario,
		moto: e.moto,
		motoCosto: e.motoCosto,
		motoZona: e.motoZona,
		motoDemora: e.motoDemora,
		envioGratis: e.envioGratis ?? false,
		envioGratisMonto: e.envioGratisMonto ?? 40_000
	});
	let e = $state<Entregas>(untrack(() => copia(data.entregas)));
	let guardando = $state(false);
	const guardado = $derived(copia(data.entregas));
	const cambiado = $derived(JSON.stringify(e) !== JSON.stringify(guardado));
	const falta = $derived(
		problemaEntregas(e, { correo: data.entregas.correo, direccionNegocio: data.negocio.direccion })
	);

	async function guardar() {
		if (falta) return ui.toast({ k: 'err', title: 'Falta algo', desc: falta });
		guardando = true;
		const r = await fetch('/panel/envios/configurar', {
			method: 'PUT',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify(e)
		})
			.catch(() => null)
			.finally(() => (guardando = false));
		if (!r) return ui.toast({ k: 'err', title: 'No se guardó', desc: 'Revisá la conexión.' });
		if (!r.ok)
			return ui.toast({
				k: 'err',
				title: 'No se guardó',
				desc:
					((await r.json().catch(() => null)) as { message?: string } | null)?.message ??
					'Probá de nuevo.'
			});
		await invalidateAll();
		e = copia(data.entregas);
		ui.toast({ k: 'ok', title: 'Envíos y retiro guardados', desc: 'Ya se ven en tu tienda' });
	}
</script>

<Seo
	title="Envíos y retiro · Panel"
	description="Cómo les llega lo que compran"
	path="/panel/envios/configurar"
	noindex
/>

<div class="pagina">
	<div class="cab" use:anim={{ t: 'rise' }}>
		<div class="c-txt">
			<a class="volver" href={resolve('/panel/envios')}
				><Icon icon={I.phArrowLeftBold} size={13} />Envíos</a
			>
			<h1>Envíos y retiro</h1>
			<div class="c-s">Cómo les llega lo que compran en tu tienda</div>
		</div>
		<div class="acciones">
			{#if cambiado}
				<button
					type="button"
					class="cambiar"
					onclick={() => (e = copia(data.entregas))}
					disabled={guardando}>Descartar</button
				>
			{/if}
			<button type="button" class="publicar" onclick={guardar} disabled={guardando || !cambiado}
				>{guardando ? 'Guardando…' : 'Guardar'}</button
			>
		</div>
	</div>

	<div class="grilla" use:anim={{ t: 'rise', delay: 60 }}>
		<section class="tarjeta">
			<div class="opcion">
				<span class="o-ico"><Icon icon={I.phStorefrontFill} size={18} /></span>
				<span class="o-txt"
					><span class="o-t">Retiro en el local</span><span class="o-d"
						>Sin costo · te avisamos cuando está listo</span
					></span
				>
				<Interruptor bind:checked={e.retiro} label="Retiro en el local" />
			</div>
			{#if e.retiro}
				<div class="campos">
					<label class="campo"
						>Dirección<input
							bind:value={e.retiroDireccion}
							maxlength={TOPES_ENTREGAS.retiroDireccion}
							placeholder={data.negocio.direccion || 'Calle, número y localidad'}
							autocomplete="street-address"
						/></label
					>
					<label class="campo"
						>Horario<input
							bind:value={e.retiroHorario}
							maxlength={TOPES_ENTREGAS.retiroHorario}
							placeholder={data.negocio.horario || 'Lunes a viernes de 10 a 19'}
						/></label
					>
				</div>
				<p class="nota">
					<Icon icon={I.phInfoFill} size={15} />Se muestran al elegir retiro en la tienda y en la
					compra.
				</p>
			{/if}
		</section>

		<section class="tarjeta">
			<div class="opcion">
				<span class="o-ico"><Icon icon={I.phMopedFill} size={18} /></span>
				<span class="o-txt"
					><span class="o-t">Envío propio en moto</span><span class="o-d"
						>Con tu moto o tu cadete, a un costo fijo</span
					></span
				>
				<Interruptor bind:checked={e.moto} label="Envío propio en moto" />
			</div>
			{#if e.moto}
				<div class="campos">
					<label class="campo"
						>Costo en pesos<span class="pesos"
							><span>$</span><input
								type="number"
								inputmode="numeric"
								min="0"
								max={TOPES_ENTREGAS.motoCosto}
								step="1"
								bind:value={e.motoCosto}
							/></span
						></label
					>
					<label class="campo"
						>Cuánto tarda<input
							bind:value={e.motoDemora}
							maxlength={TOPES_ENTREGAS.motoDemora}
							placeholder="En el día"
						/></label
					>
					<label class="campo ancho"
						>Hasta dónde llega<textarea
							bind:value={e.motoZona}
							maxlength={TOPES_ENTREGAS.motoZona}
							rows="2"
							placeholder="Barrios o localidades: Palermo, Belgrano, Núñez"></textarea></label
					>
				</div>
				<p class="nota">
					<Icon icon={I.phInfoFill} size={15} />Quien compra ve la zona y el costo antes de
					elegirla. Si su dirección queda afuera, avisale por WhatsApp.
				</p>
			{/if}
		</section>

		<section class="tarjeta">
			<div class="opcion">
				<MarcaIcono k="correo" size={36} radius={11} />
				<span class="o-txt"
					><span class="o-t">Correo Argentino</span><span class="o-d"
						>{data.entregas.correo
							? 'Cotiza con tu cuenta de MiCorreo a domicilio y a sucursal'
							: 'Conectá tu cuenta de MiCorreo para enviar a todo el país'}</span
					></span
				>
				<span class="estado" class:on={data.entregas.correo}
					>{data.entregas.correo ? 'Conectado' : 'Sin conectar'}</span
				>
			</div>
			<a class="mas" href={resolve('/panel/integraciones')}
				>{data.entregas.correo ? 'Ver en Integraciones' : 'Conectar en Integraciones'}
				<Icon icon={I.phArrowRightBold} size={13} /></a
			>
		</section>

		<section class="tarjeta">
			<div class="opcion">
				<span class="o-ico"><Icon icon={I.phTruckFill} size={18} /></span>
				<span class="o-txt"
					><span class="o-t">Envío gratis</span><span class="o-d"
						>{e.envioGratis
							? `Desde ${pesos(e.envioGratisMonto ?? 0)}, en moto y por Correo`
							: 'Para que sumen algo más al carrito'}</span
					></span
				>
				<Interruptor bind:checked={e.envioGratis} label="Envío gratis" />
			</div>
			{#if e.envioGratis}
				<div class="campos">
					<label class="campo"
						>Desde qué monto<span class="pesos"
							><span>$</span><input
								type="number"
								inputmode="numeric"
								min={TOPES_ENTREGAS.gratisMinimo}
								max={TOPES_ENTREGAS.gratisMaximo}
								step="1"
								bind:value={e.envioGratisMonto}
							/></span
						></label
					>
				</div>
				<p class="nota">
					<Icon icon={I.phInfoFill} size={15} />La tienda muestra cuánto les falta en la bolsa y en
					el carrito, y el envío sale gratis al&nbsp;pagar.
				</p>
			{/if}
		</section>
	</div>

	{#if falta}
		<p class="aviso" role="status"><Icon icon={I.phInfoFill} size={15} />{falta}</p>
	{/if}
</div>

<style>
	.pagina {
		display: flex;
		flex-direction: column;
		gap: 18px;
	}
	.cab {
		display: flex;
		align-items: end;
		gap: 12px;
		flex-wrap: wrap;
	}
	.c-txt {
		flex: 1;
		min-width: 220px;
	}
	.volver {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		min-height: 32px;
		font-size: 13.5px;
		font-weight: 600;
		color: var(--mu);
		text-decoration: none;
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
		text-wrap: pretty;
	}
	.acciones {
		display: flex;
		gap: 8px;
	}
	.cambiar {
		padding: 10px 16px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: var(--sf);
		color: var(--tx);
		font:
			600 13.5px 'Instrument Sans',
			sans-serif;
		box-shadow: var(--sh);
		white-space: nowrap;
	}
	.publicar {
		padding: 11px 20px;
		border-radius: 999px;
		border: 0;
		background: linear-gradient(180deg, #c24e25, #b0441d);
		color: #fff;
		font:
			600 14px 'Instrument Sans',
			sans-serif;
		white-space: nowrap;
	}
	.publicar:disabled {
		opacity: 0.5;
	}
	.grilla {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 12px;
		align-items: start;
	}
	.tarjeta {
		padding: 6px 18px 16px;
		border-radius: 22px;
		background: var(--sf);
		box-shadow: var(--shd);
		display: flex;
		flex-direction: column;
		gap: 10px;
		min-width: 0;
	}
	.opcion {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 12px 0 4px;
	}
	.o-ico {
		width: 36px;
		height: 36px;
		flex: none;
		border-radius: 11px;
		background: rgba(210, 88, 47, 0.12);
		color: #b0441d;
		display: grid;
		place-items: center;
	}
	.o-txt {
		flex: 1;
		min-width: 0;
	}
	.o-t {
		display: block;
		font-size: 15px;
		font-weight: 600;
	}
	.o-d {
		display: block;
		font-size: 12.5px;
		color: var(--mu);
		text-wrap: pretty;
	}
	.estado {
		flex: none;
		font-size: 12px;
		font-weight: 600;
		padding: 4px 10px;
		border-radius: 999px;
		background: var(--chip);
		color: var(--mu);
		white-space: nowrap;
	}
	.estado.on {
		background: rgba(47, 143, 99, 0.12);
		color: #237550;
	}
	.campos {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
		gap: 10px;
	}
	.campo {
		display: flex;
		flex-direction: column;
		gap: 5px;
		font-size: 12.5px;
		font-weight: 600;
		color: var(--mu);
		min-width: 0;
	}
	.campo.ancho {
		grid-column: 1 / -1;
	}
	.campo input,
	.campo textarea {
		width: 100%;
		box-sizing: border-box;
		padding: 11px 12px;
		border-radius: 12px;
		border: 1px solid var(--ln2);
		background: var(--bg);
		font:
			500 15px 'Instrument Sans',
			sans-serif;
		color: var(--tx);
		outline: none;
		resize: vertical;
	}
	.pesos {
		position: relative;
		display: block;
	}
	.pesos span {
		position: absolute;
		left: 12px;
		top: 50%;
		transform: translateY(-50%);
		font-size: 15px;
		color: var(--mu);
	}
	.pesos input {
		padding-left: 26px;
	}
	.nota,
	.aviso {
		margin: 0;
		display: flex;
		gap: 8px;
		font-size: 12.5px;
		line-height: 1.5;
		color: var(--mu);
		text-wrap: pretty;
	}
	.nota :global(svg),
	.aviso :global(svg) {
		flex: none;
		margin-top: 2px;
	}
	.aviso {
		padding: 12px 14px;
		border-radius: 16px;
		background: rgba(210, 88, 47, 0.08);
		color: #8a3416;
		font-size: 13.5px;
	}
	.mas {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		min-height: 32px;
		font-size: 13.5px;
		font-weight: 600;
		color: #b0441d;
		text-decoration: none;
	}
	@media (max-width: 899px) {
		.grilla {
			grid-template-columns: minmax(0, 1fr);
		}
	}
	@media (max-width: 760px) {
		.mas,
		.volver {
			min-height: 44px;
		}
		.acciones {
			width: 100%;
		}
		.cab .publicar,
		.cab .cambiar {
			flex: 1;
			min-height: 44px;
		}
	}
</style>
