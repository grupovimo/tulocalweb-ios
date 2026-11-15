<script lang="ts">
	import { duracion } from '#lib/turnos/formato.ts';
	import type { ProximosTurnos } from '../../tipos.ts';
	import { textoDe, type Bloque } from '../catalogo.ts';
	import { usarTienda } from '../contexto.ts';
	import { precio, sinViuda } from '../formato.ts';
	import { cuandoCorto } from '../../lanzamiento.ts';
	import Boton from '../primitivas/Boton.svelte';
	import Foto from '../primitivas/Foto.svelte';

	/**
	 * Turnos: los servicios (hasta 4) con su duración y precio, los horarios libres de hoy (los trae
	 * la portada: +page.server.ts › proximos) y el botón para reservar. Sólo con turnos activos.
	 */
	let { bloque, proximos = null }: { bloque: Bloque; proximos?: ProximosTurnos | null } = $props();
	const ctx = usarTienda();
	const turnos = $derived(ctx.tienda.turnos);
	const servicios = $derived((turnos?.servicios ?? []).slice(0, 4));
	const reservar = (q: Record<string, string>) => {
		const s = new URLSearchParams(q).toString();
		return `${ctx.base}/turnos${s ? `?${s}` : ''}`;
	};
	// El equipo (lo que mostraba Tijera): quién atiende y su próximo horario libre, si se sabe.
	const equipo = $derived(
		(turnos?.profesionales ?? []).slice(0, 8).map((p) => {
			const prox = proximos?.porProfesional?.[p.id];
			return { ...p, libre: prox ? `Libre ${cuandoCorto(prox)}` : null };
		})
	);
	const hoy = $derived(
		(proximos?.hoy ?? []).slice(0, 6).map((inicio) => ({
			t: inicio.slice(11, 16),
			href: reservar({ servicio: servicios[0]?.id ?? '', dia: inicio.slice(0, 10), inicio })
		}))
	);
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
{#if turnos && servicios.length}
	<section class="tv-turnos tv-ancho">
		<div class="tv-cab">
			<h2 class="tv-h2">{sinViuda(textoDe(bloque, 'titulo'))}</h2>
			<Boton href={reservar({})}>Reservar</Boton>
		</div>
		{#if hoy.length}
			<div class="tv-turnos-hoy">
				<p class="tv-chico tv-suave">Libres hoy</p>
				<ul>
					{#each hoy as h (h.t)}<li><a href={h.href}>{h.t}</a></li>{/each}
				</ul>
			</div>
		{/if}
		<ul class="tv-turnos-servicios" style:--cols={servicios.length}>
			{#each servicios as s (s.id)}
				<li class="tv-tarjeta">
					<a href={reservar({ servicio: s.id })}>
						<b>{s.nombre}</b>
						<span class="tv-suave">{duracion(s.duracion)}</span>
						<span class="tv-turnos-precio">{precio(s.precio)}</span>
					</a>
				</li>
			{/each}
		</ul>
		{#if equipo.length}
			<div class="tv-turnos-equipo">
				<h3 class="tv-h3">El equipo</h3>
				<ul class="tv-equipo" aria-label="El equipo">
					{#each equipo as p (p.id)}
						<li>
							<a class="tv-equipo-uno tv-tarjeta" href={reservar({ profesional: p.id })}>
								<Foto src={p.foto} alt="" nombre={p.nombre} ancho={160} proporcion="1/1" />
								<b>{p.nombre}</b>
								<span class="tv-chico tv-suave">{p.libre ?? p.rol.split(' · ')[0]}</span>
							</a>
						</li>
					{/each}
				</ul>
			</div>
		{/if}
	</section>
{/if}
