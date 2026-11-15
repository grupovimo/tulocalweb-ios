<script lang="ts">
	import { onMount, tick } from 'svelte';
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { usarTienda } from '../contexto.ts';
	import { enlace } from '../datos.ts';
	import { sinViuda } from '../formato.ts';
	import Cuenta from '../primitivas/Cuenta.svelte';
	import Foto from '../primitivas/Foto.svelte';
	import {
		anotarCupon,
		captar,
		esFallo,
		evento,
		girar,
		type PremioGanado,
		type TipoEvento
	} from '../promos-api.ts';
	import { datoVentana, type VarianteVentana, type Ventana } from '../promos.ts';
	import { anotarHecha } from '../ventanas.ts';
	import Ruleta from './Ruleta.svelte';

	/**
	 * Una ventana de la tienda (tienda-promociones.md › Ventanas): las 4 variantes × los 4
	 * contenidos, con las primitivas de la dirección (botón, campos, antetítulo y títulos de su
	 * escala; la tarjeta con el radio, el borde o la sombra de cada una: direcciones/*.css).
	 *
	 * - `centrada`, `completa` e `inferior` son modales (`<dialog>` con showModal: foco atrapado y
	 *   devuelto, Esc y tocar afuera cierran); `lateral` no corta la compra: no es modal, va abajo a
	 *   la izquierda (WhatsApp está a la derecha) y Esc la cierra si el foco está adentro.
	 * - X de 44 px siempre visible; con «Reducir movimiento» no hay animaciones.
	 * - Con foto, el antetítulo y el título van sobre la foto con velo (regla 4 del handoff).
	 * - `medir`: manda vistas, cierres, clics y envíos (no en la vista previa ni en la vitrina).
	 *
	 * Uso: la dibuja marco/Ventanas.svelte (`variante` ya es la del dispositivo: ventanas.ts).
	 */
	let {
		ventana,
		variante,
		medir = true,
		abajo = 0,
		alCerrar
	}: {
		ventana: Ventana;
		variante: VarianteVentana;
		medir?: boolean;
		/** Lo que ocupa la barra fija de abajo (px): la hoja sube por encima (regla 7). */
		abajo?: number;
		alCerrar: () => void;
	} = $props();

	const ctx = usarTienda();
	const uid = $props.id();
	const w = $derived(ventana);
	const dato = <T = unknown,>(id: string) => datoVentana<T>(w, id);
	const txt = (id: string) => {
		const x = dato<string>(id);
		return typeof x === 'string' ? x.trim() : '';
	};
	const foto = $derived(w.contenido === 'ruleta' ? '' : txt('foto'));
	const titulo = $derived(txt('titulo'));
	const texto = $derived(txt('texto'));
	const etiqueta = $derived(txt('etiqueta'));
	const hasta = $derived(txt('hasta'));
	const modal = $derived(variante !== 'lateral');
	const pedirNombre = $derived(dato<boolean>('pedirNombre') === true);
	const pedirWa = $derived(dato<boolean>('pedirWhatsapp') === true);

	function medida(tipo: TipoEvento) {
		if (medir && !ctx.tienda.vitrina) evento(ctx.tienda.slug, w.id, tipo);
	}

	// ── Abrir y cerrar ──
	let dialogo = $state<HTMLDialogElement>();
	let antes: Element | null = null;
	let accion = false;
	onMount(() => {
		antes = document.activeElement;
		if (!dialogo) return;
		if (modal) dialogo.showModal();
		else dialogo.show();
		medida('vista');
		return () => {
			if (dialogo?.open) dialogo.close();
		};
	});
	function cerrar() {
		if (!accion) medida('cierre');
		if (dialogo?.open) dialogo.close();
		// El foco vuelve a donde estaba (también en la lateral, que no es modal).
		if (antes instanceof HTMLElement && antes.isConnected) antes.focus({ preventScroll: true });
		alCerrar();
	}
	function teclas(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			e.preventDefault();
			cerrar();
		}
	}

	// ── Promo: el botón ──
	const boton = $derived(dato<{ t: string; a: string }>('boton'));
	const href = $derived(boton ? enlace(ctx.tienda, boton.a) : null);
	function alTocarBoton() {
		accion = true;
		medida('clic');
		cerrar();
	}

	// ── Cupón: copiar y usar ──
	const codigo = $derived(txt('cupon'));
	let copiado = $state(false);
	let usado = $state(false);
	async function copiar(c: string) {
		try {
			await navigator.clipboard.writeText(c);
			copiado = true;
			setTimeout(() => (copiado = false), 2000);
		} catch {
			// Sin portapapeles: queda seleccionado para copiar a mano.
			const el = document.getElementById(`${uid}-codigo`);
			if (el) getSelection()?.selectAllChildren(el);
		}
	}
	function usar(c: string) {
		anotarCupon(ctx.tienda.slug, ctx.base, c);
		usado = true;
		if (!accion) {
			accion = true;
			medida('clic');
		}
	}

	// ── Captación y ruleta: el formulario ──
	let email = $state('');
	let nombre = $state('');
	let whatsapp = $state('');
	let acepto = $state(false);
	let enviando = $state(false);
	let error = $state('');
	let premio = $state<PremioGanado | null>(null);
	let porcion = $state<number | null>(null);
	let girando = $state(false);
	const textoBoton = $derived(txt('textoBoton') || 'Enviar');
	const listo = $derived(txt('listo'));
	const porciones = $derived(
		(dato<{ texto: string }[]>('porciones') ?? []).map((p) => p.texto ?? '')
	);

	async function enviar(e: SubmitEvent) {
		e.preventDefault();
		if (enviando || premio) return;
		error = '';
		const m = email.trim();
		if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(m) || m.length > 120) {
			error = 'Revisá el email.';
			return;
		}
		if (!acepto) {
			error = 'Para seguir, aceptá que te escribamos.';
			return;
		}
		if (ctx.tienda.vitrina || !medir) {
			error = medir
				? 'Es una tienda de ejemplo: acá no se anota nadie.'
				: 'En la vista previa no se anota nadie.';
			return;
		}
		enviando = true;
		const c = {
			ventana: w.id,
			email: m,
			...(pedirNombre && nombre.trim() ? { nombre: nombre.trim().slice(0, 80) } : {}),
			...(pedirWa && whatsapp.trim() ? { whatsapp: whatsapp.trim().slice(0, 30) } : {}),
			consentimiento: true as const
		};
		if (w.contenido === 'ruleta') {
			const r = await girar(ctx.tienda.slug, c);
			enviando = false;
			if (esFallo(r)) return void (error = r.error);
			accion = true;
			medida('envio');
			anotarHecha(ctx.tienda.slug, w.id);
			porcion = r.porcion;
			girando = true;
			premioPendiente = r.premio;
		} else {
			const r = await captar(ctx.tienda.slug, c);
			enviando = false;
			if (esFallo(r)) return void (error = r.error);
			accion = true;
			medida('envio');
			anotarHecha(ctx.tienda.slug, w.id);
			ganar(r.premio);
		}
	}
	let premioPendiente: PremioGanado | null = null;
	async function ganar(p: PremioGanado) {
		premio = p;
		girando = false;
		// El código ganado ya queda en el carrito (se aplica solo).
		if (p.codigo) {
			anotarCupon(ctx.tienda.slug, ctx.base, p.codigo);
			usado = true;
		}
		await tick();
		document.getElementById(`${uid}-listo`)?.focus();
	}
	const tituloPremio = $derived(
		!premio
			? ''
			: premio.tipo === 'nada' && w.contenido === 'ruleta'
				? '¡Gracias por jugar!'
				: '¡Listo!'
	);
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- links que eligió la tienda -->
{#snippet codigoCupon(c: string)}
	<div class="tv-vent-cupon">
		<span class="tv-vent-codigo" id="{uid}-codigo">{c}</span>
		<button type="button" class="tv-btn tv-btn--secundario tv-btn--chico" onclick={() => copiar(c)}
			><Icon icon={copiado ? I.phCheckBold : I.phCopy} size={16} />{copiado
				? 'Copiado'
				: 'Copiar'}</button
		>
	</div>
	<p class="tv-vent-estado" role="status">
		{#if usado}<Icon icon={I.phCheckCircleFill} size={16} />Ya está en tu carrito: se aplica
			solo&nbsp;al&nbsp;pagar.{/if}
	</p>
{/snippet}

{#snippet campos()}
	<form class="tv-vent-form" onsubmit={enviar} novalidate>
		{#if pedirNombre}
			<div class="tv-campo">
				<label for="{uid}-nombre">Tu nombre <span class="tv-suave">(opcional)</span></label>
				<input id="{uid}-nombre" bind:value={nombre} autocomplete="given-name" maxlength="80" />
			</div>
		{/if}
		<div class="tv-campo">
			<label for="{uid}-email">Tu email</label>
			<input
				id="{uid}-email"
				type="email"
				bind:value={email}
				autocomplete="email"
				inputmode="email"
				maxlength="120"
				required
				aria-invalid={error === 'Revisá el email.' || undefined}
				aria-describedby={error ? `${uid}-error` : undefined}
			/>
		</div>
		{#if pedirWa}
			<div class="tv-campo">
				<label for="{uid}-wa">Tu WhatsApp <span class="tv-suave">(opcional)</span></label>
				<input
					id="{uid}-wa"
					type="tel"
					bind:value={whatsapp}
					autocomplete="tel"
					inputmode="tel"
					maxlength="30"
				/>
			</div>
		{/if}
		<label class="tv-casilla tv-vent-acepto">
			<input type="checkbox" bind:checked={acepto} required />
			<span
				>Acepto que {ctx.tienda.nombre} me escriba con novedades. Podés darte de baja cuando quieras (<a
					href="/legales/privacidad"
					target="_blank"
					rel="noopener">privacidad</a
				>).</span
			>
		</label>
		{#if error}<p class="tv-vent-error" id="{uid}-error" role="alert">{error}</p>{/if}
		<button
			type="submit"
			class="tv-btn tv-btn--primario tv-btn--bloque"
			disabled={enviando || girando}>{enviando ? 'Un momento…' : textoBoton}</button
		>
	</form>
{/snippet}

{#snippet resultado()}
	{#if premio}
		<div class="tv-vent-listo" id="{uid}-listo" tabindex="-1">
			<p class="tv-h3">{tituloPremio}</p>
			{#if premio.tipo === 'nada'}
				<p class="tv-lead">
					{w.contenido === 'ruleta'
						? 'Esta vez no tocó premio, pero ya estás en la lista de novedades.'
						: listo || 'Ya estás en la lista: te vamos a escribir con las novedades.'}
				</p>
			{:else}
				{#if listo}<p class="tv-lead">{sinViuda(listo)}</p>{/if}
				{#if premio.codigo}{@render codigoCupon(premio.codigo)}{/if}
				{#if premio.descarga}
					<a
						class="tv-btn tv-btn--primario tv-btn--bloque"
						href={premio.descarga}
						target="_blank"
						rel="noopener"
						download
						><Icon icon={I.phDownloadSimpleBold} size={18} />Descargar {premio.nombreArchivo ||
							'el archivo'}</a
					>
				{/if}
			{/if}
			<button type="button" class="tv-link tv-vent-seguir" onclick={cerrar}>Seguir mirando</button>
		</div>
	{/if}
{/snippet}

<dialog
	bind:this={dialogo}
	class="tv-vent tv-vent--{variante}"
	class:tv-vent--foto={!!foto}
	data-contenido={w.contenido}
	aria-labelledby="{uid}-titulo"
	aria-describedby={texto ? `${uid}-texto` : undefined}
	aria-modal={modal ? 'true' : undefined}
	style:--tv-vent-abajo="{abajo}px"
	oncancel={(e) => {
		e.preventDefault();
		cerrar();
	}}
	onkeydown={modal ? undefined : teclas}
	onclick={(e) => e.target === dialogo && cerrar()}
>
	<!-- Al abrir, el foco va a la ventana (no a la X ni a un campo: abriría el teclado del celular). -->
	<div class="tv-vent-caja tv-tarjeta" tabindex="-1" autofocus>
		<button type="button" class="tv-icono tv-vent-x" aria-label="Cerrar" onclick={cerrar}
			><Icon icon={I.phXBold} size={20} /></button
		>
		{#if foto && variante === 'completa'}
			<div class="tv-vent-fondo">
				<Foto src={foto} alt="" nombre={titulo || ctx.tienda.nombre} ancho={1600} />
				<span class="tv-velo"></span>
			</div>
		{:else if foto}
			<div class="tv-vent-foto">
				<Foto src={foto} alt="" nombre={titulo || ctx.tienda.nombre} ancho={900} />
				<span class="tv-velo"></span>
				<div class="tv-vent-sobre">
					{#if etiqueta}<p class="tv-ante">{etiqueta}</p>{/if}
					<h2 class="tv-h2 tv-vent-titulo" id="{uid}-titulo">{sinViuda(titulo)}</h2>
				</div>
			</div>
		{/if}
		<div class="tv-vent-cuerpo">
			{#if !foto || variante === 'completa'}
				{#if etiqueta}<p class="tv-ante">{etiqueta}</p>{/if}
				<h2 class="tv-h2 tv-vent-titulo" id="{uid}-titulo">{sinViuda(titulo)}</h2>
			{/if}
			{#if texto && !premio}<p class="tv-lead tv-vent-texto" id="{uid}-texto">
					{sinViuda(texto)}
				</p>{/if}
			{#if hasta && !premio}<Cuenta {hasta} caja alVencer={cerrar} />{/if}

			{#if w.contenido === 'promo'}
				{#if boton?.t && href}
					<a
						class="tv-btn tv-btn--primario tv-btn--bloque"
						{href}
						target={href.startsWith('http') ? '_blank' : undefined}
						rel={href.startsWith('http') ? 'noopener' : undefined}
						onclick={alTocarBoton}>{boton.t}</a
					>
				{/if}
			{:else if w.contenido === 'cupon'}
				{#if codigo}
					{@render codigoCupon(codigo)}
					{#if txt('condiciones')}<p class="tv-chico tv-suave">{txt('condiciones')}</p>{/if}
					{#if usado}
						<a class="tv-btn tv-btn--primario tv-btn--bloque" href="{ctx.base}/carrito"
							>Ir al carrito</a
						>
					{:else}
						<button
							type="button"
							class="tv-btn tv-btn--primario tv-btn--bloque"
							onclick={() => usar(codigo)}>Usar en mi compra</button
						>
					{/if}
				{/if}
			{:else if w.contenido === 'ruleta'}
				<Ruleta {porciones} {porcion} onlisto={() => premioPendiente && ganar(premioPendiente)} />
				{#if premio}{@render resultado()}{:else}{@render campos()}{/if}
			{:else if premio}
				{@render resultado()}
			{:else}
				{@render campos()}
			{/if}
		</div>
	</div>
</dialog>
