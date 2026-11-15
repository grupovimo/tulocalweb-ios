<script lang="ts">
	import { untrack } from 'svelte';
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { AltaPlan, cajaDe } from '../../suscripcion.svelte.ts';
	import { CADA, porPeriodo, queIncluye } from '../../suscripciones.ts';
	import type { PlanPublico, TiendaPublica } from '../../tipos.ts';
	import { precio } from '../formato.ts';
	import Tienda from '../marco/Tienda.svelte';
	import Foto from '../primitivas/Foto.svelte';

	/**
	 * Un plan de suscripción y su alta (§51; no está en el handoff: con las piezas del checkout). Los
	 * otros planes para cambiar, los datos, cómo llega (envío o retiro) y, al costado, la caja con lo
	 * que incluye, cuánto se ahorra contra comprarlo suelto y el precio por período.
	 */
	let {
		tienda,
		plan,
		demo = false,
		error = null
	}: { tienda: TiendaPublica; plan: PlanPublico; demo?: boolean; error?: string | null } = $props();
	const base = $derived(tienda.base ?? `/t/${tienda.slug}`);
	const alta = untrack(() => new AltaPlan(plan));
	const caja = $derived(cajaDe(plan, tienda));
	const entregas = $derived(
		[
			plan.envio ? { id: 'envio' as const, t: 'Envío', d: 'A tu casa, incluido' } : null,
			plan.retiro
				? { id: 'retiro' as const, t: 'Retiro', d: tienda.envios.retiroDireccion || 'En el local' }
				: null
		].filter((x) => !!x)
	);
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
<Tienda {tienda} pagina="suscripciones">
	<div class="tv-ancho tv-pag tv-checkout">
		<div class="tv-sus-cab">
			<h1 class="tv-h1 tv-pag-titulo">{plan.nombre}</h1>
			{#if plan.descripcion}<p class="tv-lead">{plan.descripcion}</p>{/if}
		</div>
		<form class="tv-checkout-cuerpo" method="POST" use:alta.form>
			<input type="hidden" name="entrega" value={alta.entrega} />
			<div class="tv-checkout-pasos">
				{#if tienda.suscripciones.length > 1}
					<section class="tv-checkout-paso" aria-labelledby="tv-sus-planes">
						<h2 class="tv-h3" id="tv-sus-planes">Elegí tu plan</h2>
						<div class="tv-sus-planes">
							{#each tienda.suscripciones as p (p.id)}
								<a
									href="{base}/suscripciones/{p.id}"
									class="tv-radio tv-tarjeta"
									class:tv-radio--on={p.id === plan.id}
									aria-current={p.id === plan.id ? 'page' : undefined}
									><span class="tv-radio-txt"
										><b>{p.nombre}</b><span class="tv-chico tv-suave"
											>{precio(p.precio)} {CADA[p.frecuencia]}</span
										></span
									></a
								>
							{/each}
						</div>
					</section>
				{/if}
				<section class="tv-checkout-paso" aria-labelledby="tv-sus-datos">
					<h2 class="tv-h3" id="tv-sus-datos">Tus datos</h2>
					<div class="tv-campos">
						{#each [['email', 'Email', 'email', 'email'], ['nombre', 'Nombre y apellido', 'text', 'name'], ['telefono', 'WhatsApp', 'tel', 'tel']] as const as [n, l, tipo, ac] (n)}
							<div class="tv-campo">
								<label for="tv-sus-{n}">{l}</label>
								<input
									id="tv-sus-{n}"
									name={n}
									type={tipo}
									autocomplete={ac}
									required
									maxlength="120"
								/>
							</div>
						{/each}
					</div>
				</section>
				<section class="tv-checkout-paso" aria-labelledby="tv-sus-entrega">
					<h2 class="tv-h3" id="tv-sus-entrega">Cómo te llega</h2>
					<div class="tv-radios tv-radios--fila" role="radiogroup" aria-label="Cómo te llega">
						{#each entregas as e (e.id)}
							<button
								type="button"
								role="radio"
								aria-checked={alta.entrega === e.id}
								class="tv-radio tv-tarjeta"
								class:tv-radio--on={alta.entrega === e.id}
								onclick={() => (alta.entrega = e.id)}
								><span class="tv-radio-punto" aria-hidden="true"></span><span class="tv-radio-txt"
									><b>{e.t}</b><span class="tv-chico tv-suave">{e.d}</span></span
								></button
							>
						{/each}
					</div>
					{#if alta.entrega === 'envio'}
						<div class="tv-campos tv-campos--envio">
							<div class="tv-campo">
								<label for="tv-sus-cp">Código postal</label>
								<input
									id="tv-sus-cp"
									name="cp"
									inputmode="numeric"
									maxlength="8"
									autocomplete="postal-code"
								/>
							</div>
							<div class="tv-campo">
								<label for="tv-sus-dir">Dirección (calle, número, piso)</label>
								<input
									id="tv-sus-dir"
									name="direccion"
									maxlength="160"
									autocomplete="street-address"
									required
								/>
							</div>
						</div>
					{/if}
				</section>
				{#if error}<p class="tv-error tv-error--caja" role="alert">{error}</p>{/if}
			</div>

			<aside class="tv-checkout-resumen tv-tarjeta" aria-label="Tu suscripción">
				<h2 class="tv-h3">Tu suscripción</h2>
				{#if caja.fotos.length}
					<div class="tv-sus-fotos">
						{#each caja.fotos.slice(0, 6) as f, i (i)}<Foto
								src={f}
								alt=""
								nombre={plan.nombre}
								ancho={120}
								proporcion="1/1"
							/>{/each}
					</div>
				{/if}
				<p><b>{plan.nombre}</b></p>
				{#if queIncluye(plan)}<p class="tv-chico tv-suave">{queIncluye(plan)}</p>{/if}
				<dl class="tv-cuentas">
					{#if caja.ahorro}
						<div>
							<dt>Sueltos</dt>
							<dd><s class="tv-suave">{precio(caja.sueltos)}</s></dd>
						</div>
						<div class="tv-ok">
							<dt>Ahorrás {caja.pct} %</dt>
							<dd>− {precio(caja.ahorro)}</dd>
						</div>
					{/if}
					<div>
						<dt>{alta.entrega === 'envio' ? 'Envío' : 'Retiro'}</dt>
						<dd class="tv-ok">{alta.entrega === 'envio' ? 'Incluido' : 'En el local'}</dd>
					</div>
				</dl>
				<div class="tv-cuentas-total">
					<span>{porPeriodo(plan.frecuencia)}</span><b>{precio(plan.precio)}</b>
				</div>
				<button
					class="tv-btn tv-btn--primario tv-btn--bloque tv-checkout-pagar"
					disabled={alta.enviando || tienda.vitrina || plan.agotado}
					>{tienda.vitrina
						? 'Es una tienda de ejemplo'
						: plan.agotado
							? 'Sin lugar por ahora'
							: `Suscribirme · ${precio(plan.precio)}`}</button
				>
				<ul class="tv-sus-notas tv-chico tv-suave">
					{#if demo}
						<li><Icon icon={I.phCheck} size={15} />Es la demo: no se cobra nada</li>
					{:else}
						<li><Icon icon={I.phCheck} size={15} />Débito automático con Mercado&nbsp;Pago</li>
						<li><Icon icon={I.phCheck} size={15} />Pausás o cancelás cuando quieras</li>
					{/if}
				</ul>
			</aside>
		</form>
	</div>
</Tienda>
