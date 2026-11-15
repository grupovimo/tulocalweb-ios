<script lang="ts">
	import { enhance } from '$app/forms';
	import { page } from '$app/state';
	import { untrack } from 'svelte';
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { duracion } from '#lib/turnos/formato.ts';
	import { Reserva } from '../../reserva.svelte.ts';
	import type { TiendaPublica } from '../../tipos.ts';
	import { precio } from '../formato.ts';
	import Tienda from '../marco/Tienda.svelte';
	import Foto from '../primitivas/Foto.svelte';

	/**
	 * Reservar un turno (no está en el handoff: con las piezas del checkout). Servicio, profesional
	 * (o el primero libre), día y horario (reserva.svelte.ts; lo que venga en la URL queda elegido),
	 * los datos y la seña con Mercado Pago si la tienda la pide. Al costado, la reserva.
	 */
	let { tienda, error = null }: { tienda: TiendaPublica; error?: string | null } = $props();
	const turnos = $derived(tienda.turnos!);
	const reserva = untrack(
		() =>
			new Reserva(tienda.slug, tienda.turnos!, {
				servicio: page.url.searchParams.get('servicio'),
				profesional: page.url.searchParams.get('profesional'),
				dia: page.url.searchParams.get('dia'),
				inicio: page.url.searchParams.get('inicio')
			})
	);
	const PASOS = ['Servicio', 'Profesional', 'Horario', 'Pago'];
	const anticipacion = $derived.by(() => {
		const m = turnos.anticipacionMin ?? 0;
		return m >= 120 && m % 60 === 0 ? `${m / 60} h` : `${m} min`;
	});
	let reservando = $state(false);
</script>

<Tienda {tienda} pagina="turnos">
	<div class="tv-ancho tv-pag tv-checkout">
		<ol class="tv-pasos" aria-label="Pasos de la reserva">
			{#each PASOS as p, i (p)}
				<li
					class:tv-paso--hecho={i < reserva.paso}
					class:tv-paso--ahora={i === reserva.paso}
					aria-current={i === reserva.paso ? 'step' : undefined}
				>
					<span>{i + 1}</span>{p}
				</li>
			{/each}
		</ol>
		<h1 class="tv-h1 tv-pag-titulo">Reservá tu turno</h1>
		<form
			class="tv-checkout-cuerpo"
			method="POST"
			use:enhance={() => {
				reservando = true;
				return async ({ result, update }) => {
					reservando = false;
					// Mercado Pago es otro sitio: se navega de verdad.
					if (result.type === 'redirect' && result.location.startsWith('https://'))
						return void (location.href = result.location);
					await update();
					// Si otro tomó el horario mientras tanto, se ve al toque.
					if (result.type === 'failure') void reserva.cargar();
				};
			}}
		>
			<input type="hidden" name="servicioId" value={reserva.servicioId ?? ''} />
			<input type="hidden" name="profesionalId" value={reserva.profesionalId ?? ''} />
			<input type="hidden" name="inicio" value={reserva.inicio ?? ''} />
			<div class="tv-checkout-pasos">
				<section class="tv-checkout-paso" aria-labelledby="tv-tu-servicio">
					<h2 class="tv-h3" id="tv-tu-servicio">Servicio</h2>
					<div class="tv-radios" role="radiogroup" aria-label="Servicio">
						{#each turnos.servicios as s (s.id)}
							{@const on = reserva.servicioId === s.id}
							<button
								type="button"
								role="radio"
								aria-checked={on}
								class="tv-radio tv-tarjeta"
								class:tv-radio--on={on}
								onclick={() => reserva.elegirServicio(s.id)}
								><span class="tv-radio-punto" aria-hidden="true"></span><span class="tv-radio-txt"
									><b>{s.nombre}</b><span class="tv-chico tv-suave">{duracion(s.duracion)}</span
									></span
								><b class="tv-radio-precio">{precio(s.precio)}</b></button
							>
						{/each}
					</div>
				</section>

				<section class="tv-checkout-paso" aria-labelledby="tv-tu-prof">
					<h2 class="tv-h3" id="tv-tu-prof">Profesional</h2>
					<div class="tv-equipo" role="radiogroup" aria-label="Profesional">
						{#each [null, ...reserva.profesionales] as p (p?.id ?? 'cualquiera')}
							{@const on = reserva.profesionalId === (p?.id ?? null)}
							<button
								type="button"
								role="radio"
								aria-checked={on}
								class="tv-equipo-uno tv-tarjeta"
								class:tv-radio--on={on}
								onclick={() => reserva.elegirProfesional(p?.id ?? null)}
							>
								{#if p}<Foto
										src={p.foto}
										alt=""
										nombre={p.nombre}
										ancho={160}
										proporcion="1/1"
									/>{:else}<span class="tv-equipo-todos"
										><Icon icon={I.phShuffleBold} size={22} /></span
									>{/if}
								<b>{p?.nombre ?? 'Cualquiera'}</b>
								<span class="tv-chico tv-suave"
									>{p ? p.rol.split(' · ')[0] : 'El primero libre'}</span
								>
							</button>
						{/each}
					</div>
				</section>

				<section class="tv-checkout-paso" aria-labelledby="tv-tu-dia">
					<div class="tv-turnos-mes">
						<h2 class="tv-h3" id="tv-tu-dia">Día y horario</h2>
						<span>
							{#if reserva.hayAntes || reserva.hayDespues}<button
									type="button"
									class="tv-icono"
									aria-label="Días anteriores"
									disabled={!reserva.hayAntes}
									onclick={() => reserva.mover(-1)}
									><Icon icon={I.phCaretLeftBold} size={14} /></button
								>{/if}<b>{reserva.mes}</b>{#if reserva.hayAntes || reserva.hayDespues}<button
									type="button"
									class="tv-icono"
									aria-label="Días siguientes"
									disabled={!reserva.hayDespues}
									onclick={() => reserva.mover(1)}
									><Icon icon={I.phCaretRightBold} size={14} /></button
								>{/if}
						</span>
					</div>
					<div class="tv-dias" role="radiogroup" aria-label="Día">
						{#each reserva.dias as d (d.dia)}
							{@const on = reserva.dia === d.dia}
							<button
								type="button"
								role="radio"
								aria-checked={on}
								aria-label="{d.w} {d.n}{d.abierto ? '' : ', cerrado'}"
								disabled={!d.abierto}
								class="tv-dia"
								class:tv-radio--on={on}
								onclick={() => reserva.elegirDia(d.dia)}><small>{d.w}</small><b>{d.n}</b></button
							>
						{/each}
					</div>
					<div class="tv-horarios">
						{#if !reserva.servicioId}
							<p class="tv-suave">Elegí un servicio para ver los horarios.</p>
						{:else if reserva.horarios && !reserva.grupos.length}
							<p class="tv-suave">
								No quedan horarios este día. Probá con otro día u otro&nbsp;profesional.
							</p>
						{/if}
						{#each reserva.grupos as g (g.k)}
							<div>
								<p class="tv-chico tv-suave">{g.t}</p>
								<div
									class="tv-horas"
									role="radiogroup"
									aria-label="Horarios a la {g.t.toLowerCase()}"
								>
									{#each g.s as h (h.inicio)}
										{@const on = reserva.inicio === h.inicio}
										<button
											type="button"
											role="radio"
											aria-checked={on}
											aria-label="{h.hora}{h.libre ? '' : ', ocupado'}"
											disabled={!h.libre}
											class="tv-hora"
											class:tv-radio--on={on}
											onclick={() => reserva.elegirHorario(h)}>{h.hora}</button
										>
									{/each}
								</div>
							</div>
						{/each}
					</div>
				</section>

				{#if reserva.listo}
					<section class="tv-checkout-paso" aria-labelledby="tv-tu-datos">
						<h2 class="tv-h3" id="tv-tu-datos">Tus datos</h2>
						<div class="tv-campos">
							{#each [['nombre', 'Nombre', 'text', 'name'], ['telefono', 'WhatsApp', 'tel', 'tel'], ['email', 'Email', 'email', 'email']] as const as [n, l, tipo, ac] (n)}
								<div class="tv-campo">
									<label for="tv-tu-{n}">{l}</label>
									<input
										id="tv-tu-{n}"
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
					<section class="tv-checkout-paso" aria-labelledby="tv-tu-pago">
						<h2 class="tv-h3" id="tv-tu-pago">Pago</h2>
						<div class="tv-radio tv-radio--on tv-tarjeta">
							<Icon icon={reserva.sena ? I.phCreditCard : I.phStorefrontFill} size={22} />
							<span class="tv-radio-txt"
								>{#if reserva.sena}<b>Mercado Pago</b><span class="tv-chico tv-suave"
										>Seña de {precio(reserva.sena)} para confirmar · el resto, en el local</span
									>{:else}<b>Pagás en el local</b><span class="tv-chico tv-suave"
										>No hace falta seña: el turno queda confirmado</span
									>{/if}</span
							>
						</div>
					</section>
				{/if}

				{#if error}<p class="tv-error tv-error--caja" role="alert">{error}</p>{/if}
				<button
					class="tv-btn tv-btn--primario tv-btn--bloque tv-checkout-pagar"
					disabled={!reserva.listo || reservando || tienda.vitrina}
					><Icon icon={I.phCalendarCheckBold} size={16} />{tienda.vitrina
						? 'Es una tienda de ejemplo'
						: !reserva.listo
							? 'Elegí servicio y horario'
							: reserva.sena
								? 'Confirmar y pagar seña'
								: 'Confirmar turno'}</button
				>
			</div>

			<aside class="tv-checkout-resumen tv-tarjeta" aria-label="Tu reserva">
				<h2 class="tv-h3">Tu reserva</h2>
				<div class="tv-turno-quien">
					{#if reserva.profesional?.foto}<Foto
							src={reserva.profesional.foto}
							alt=""
							nombre={reserva.profesional.nombre}
							ancho={120}
							proporcion="1/1"
						/>{/if}
					<div>
						<b>{reserva.conQuien ? `Con ${reserva.conQuien}` : 'Con el primero libre'}</b>
						<p class="tv-chico tv-suave">{reserva.cuando ?? 'Elegí el día y el horario'}</p>
					</div>
				</div>
				{#if reserva.servicio}
					<dl class="tv-cuentas">
						<div>
							<dt>{reserva.servicio.nombre}</dt>
							<dd>{precio(reserva.servicio.precio)}</dd>
						</div>
						<div class="tv-suave">
							<dt>Duración total</dt>
							<dd>{reserva.duracion}</dd>
						</div>
					</dl>
					<div class="tv-cuentas-total">
						<span>Total</span><b>{precio(reserva.servicio.precio)}</b>
					</div>
				{/if}
				<p class="tv-chico tv-suave tv-carrito-seguro">
					<Icon icon={I.phShieldCheck} size={16} /><span
						>{#if reserva.sena}Seña de {precio(reserva.sena)} para confirmar. Podés cancelar desde el
							link del turno hasta {anticipacion} antes; la seña te la devuelve la&nbsp;tienda.{:else}No
							pagás nada ahora. Podés cancelar desde el link del turno hasta {anticipacion}&nbsp;antes.{/if}</span
					>
				</p>
			</aside>
		</form>
	</div>
</Tienda>
