<script lang="ts">
	import { deserialize } from '$app/forms';
	import { goto, invalidate } from '$app/navigation';
	import { iniciales } from '#lib/formato.ts';
	import { abrirVentana, cerrarVentana, ventanas } from '#lib/redactar.svelte.ts';
	import type { Fila } from '#lib/server/bandeja.ts';
	import Carpetas from '#lib/ui/Carpetas.svelte';
	import Editor from '#lib/ui/Editor.svelte';
	import Lectura from '#lib/ui/Lectura.svelte';
	import Lista from '#lib/ui/Lista.svelte';
	import Marca from '#lib/ui/Marca.svelte';
	import Bienvenida from '#lib/ui/Bienvenida.svelte';
	import { activarAvisos, app } from '#lib/avisos.svelte.ts';
	import { onMount } from 'svelte';
	import { page } from '$app/state';

	let { data } = $props();

	// Escribibles: el tema cambia con el botón y la búsqueda mientras se tipea.
	let tema = $derived(data.tema);
	let buscar = $derived(data.buscar);
	let cajon = $state(false);
	let usuario = $state(false);
	let aviso = $state('');
	let avisoT: ReturnType<typeof setTimeout> | undefined;

	/** La URL de la bandeja con lo que cambia (sin el correo abierto si cambia la lista). */
	function href(o: {
		c?: string;
		cas?: string | null;
		p?: string;
		q?: string;
		m?: string | null;
		img?: boolean;
	}) {
		const u = new URLSearchParams();
		const c = o.c ?? data.carpeta;
		const cas = o.cas === undefined ? data.casilla : o.cas;
		const p = o.p ?? (o.c || o.cas !== undefined ? 'todos' : data.pestana);
		const q = o.q ?? (o.c || o.cas !== undefined ? '' : data.buscar);
		if (c !== 'entrada') u.set('c', c);
		if (cas) u.set('cas', cas);
		if (p !== 'todos') u.set('p', p);
		if (q) u.set('q', q);
		if (o.m) u.set('m', o.m);
		if (o.img) u.set('img', '1');
		const s = u.toString();
		return s ? `/?${s}` : '/';
	}
	const hrefFila = (f: Fila) => href({ m: f.id });

	function avisar(t: string) {
		aviso = t;
		clearTimeout(avisoT);
		avisoT = setTimeout(() => (aviso = ''), 3500);
	}
	async function cambiarTema() {
		tema = tema === 'oscuro' ? 'claro' : 'oscuro';
		document.documentElement.dataset.tema = tema;
		await fetch('/tema', { method: 'POST', body: JSON.stringify({ tema }) });
	}
	async function accion(que: string, valor = '') {
		if (!data.leido) return;
		// Desde la papelera, borrar es para siempre: se pregunta.
		if (
			que === 'borrar' &&
			data.carpeta === 'papelera' &&
			!confirm('¿Borrar este correo para siempre?')
		)
			return;
		const fd = new FormData();
		const bloquear = que === 'bloquear';
		fd.set('que', que);
		fd.set('valor', valor);
		fd.set('email', data.leido.de.email);
		fd.append('id', data.leido.id);
		const r = await fetch(bloquear ? '?/bloquear' : '?/accion', {
			method: 'POST',
			body: fd,
			headers: { 'x-sveltekit-action': 'true' }
		});
		const res = deserialize(await r.text());
		if (res.type === 'failure') return avisar(String(res.data?.error ?? 'No se pudo.'));
		if (res.type === 'error') return avisar('No se pudo.');
		const sale = [
			'archivar',
			'borrar',
			'spam',
			'no-leido',
			'mover',
			'bloquear',
			'cancelar-envio'
		].includes(que);
		avisar(
			{
				archivar: 'Archivado',
				borrar: data.carpeta === 'papelera' ? 'Borrado para siempre' : 'Movido a la papelera',
				spam: 'Movido a spam',
				'no-leido': 'Marcado como no leído',
				destacar: 'Destacado',
				'quitar-destacado': 'Sacado de destacados',
				etiquetar: 'Etiqueta puesta',
				desetiquetar: 'Etiqueta sacada',
				mover: 'Movido',
				bloquear: 'Bloqueado: lo que mande va a Spam',
				'cancelar-envio': 'Envío cancelado: quedó en Borradores'
			}[que] ?? 'Listo'
		);
		if (sale) await goto(href({}), { reset: false, refreshAll: true });
		else await invalidate('correo:bandeja');
	}
	function redactar() {
		cajon = false;
		const cuenta =
			data.casilla ?? data.casillas.find((c) => c.personal)?.cuenta ?? data.casillas[0]?.cuenta;
		if (cuenta) abrirVentana({ cuenta });
	}
	// El atajo «Redactar» del ícono instalado (manifest: /?redactar=1).
	onMount(() => {
		if (page.url.searchParams.get('redactar') === '1') {
			redactar();
			goto(href({}), { replace: true, reset: false });
		}
	});
	async function avisos() {
		usuario = false;
		if (await activarAvisos()) avisar('Avisos activados en este dispositivo');
		else if (app.error) avisar(app.error);
	}
	async function enviado(programado: boolean) {
		avisar(programado ? 'Programado: sale a la hora elegida' : 'Enviado');
		await invalidate('correo:bandeja');
	}

	const contactos = $derived.by(() => {
		const m = new Map<string, string>();
		for (const f of data.filas) if (f.email && !m.has(f.email)) m.set(f.email, f.nombre);
		return [...m].map(([email, nombre]) => ({ email, nombre }));
	});
	const titulo = $derived(
		`${data.carpetas.find((c) => c.id === data.carpeta)?.l ?? 'Entrada'}${data.casilla ? ` · ${data.casillas.find((c) => c.cuenta === data.casilla)?.email.split('@')[0]}@` : data.casillas.length > 1 ? ' · todas las casillas' : ''}`
	);
	const vacio = $derived(
		data.buscar
			? `No encontramos correos con «${data.buscar}».`
			: data.pestana === 'sinleer'
				? 'No hay correos sin leer.'
				: 'No hay correos acá.'
	);
	const yo = $derived(iniciales(data.yo.email.split('@')[0]));
	// Qué se está viendo: una casilla o todas juntas (las compartidas se mezclan si no se dice).
	const viendo = $derived(
		data.casilla
			? (data.casillas.find((c) => c.cuenta === data.casilla)?.email ?? '')
			: data.casillas.length > 1
				? `Todas las casillas (${data.casillas.length})`
				: (data.casillas[0]?.email ?? '')
	);
	async function vaciar() {
		const cual = data.casilla ? 'de esta casilla' : 'de todas las casillas';
		if (!confirm(`¿Vaciar la papelera ${cual}? Se borra para siempre.`)) return;
		const fd = new FormData();
		fd.set('casilla', data.casilla ?? '');
		const r = await fetch('?/vaciar', {
			method: 'POST',
			body: fd,
			headers: { 'x-sveltekit-action': 'true' }
		});
		const res = deserialize(await r.text());
		if (res.type !== 'success') return avisar('No se pudo vaciar la papelera.');
		const n = Number(res.data?.n ?? 0);
		avisar(
			n
				? `Papelera vacía: ${n} ${n === 1 ? 'correo borrado' : 'correos borrados'}`
				: 'La papelera ya estaba vacía'
		);
		await invalidate('correo:bandeja');
	}
</script>

<svelte:head><title>{data.leido ? data.leido.asunto : titulo} · Correo</title></svelte:head>

<div class="app" class:leyendo={!!data.leido}>
	<header class="barra">
		<button class="menu-tel" aria-label="Carpetas y casillas" onclick={() => (cajon = true)}
			><i class="icon-menu"></i></button
		>
		<div class="marca">
			<Marca marca={data.marca} />
			<div class="mt">
				<div class="mc">Correo</div>
				<div class="mm">{data.marca}</div>
			</div>
		</div>
		<form
			class="buscar vidrio"
			role="search"
			onsubmit={(e) => (
				e.preventDefault(),
				goto(href({ q: buscar.trim(), m: null }), { reset: false })
			)}
		>
			<i class="icon-search"></i>
			<input
				bind:value={buscar}
				placeholder="Buscar en el correo"
				aria-label="Buscar en el correo"
			/>
			{#if data.buscar}
				<a href={href({ q: '' })} aria-label="Borrar la búsqueda"><i class="icon-x"></i></a>
			{/if}
		</form>
		<div class="der">
			<button
				class="ic"
				aria-label={tema === 'oscuro' ? 'Tema claro' : 'Tema oscuro'}
				onclick={cambiarTema}><i class={tema === 'oscuro' ? 'icon-sun' : 'icon-moon'}></i></button
			>
			<a class="ic" href="/ajustes" aria-label="Ajustes"><i class="icon-settings"></i></a>
			<span class="usuario-w">
				<button
					class="av"
					aria-label="Tu cuenta"
					aria-expanded={usuario}
					onclick={() => (usuario = !usuario)}>{yo}</button
				>
				{#if usuario}
					<div class="pop vidrio">
						<div class="pe">{data.yo.email}</div>
						{#if app.permiso !== 'granted' && app.permiso !== 'no-soportado'}
							<button onclick={avisos}><i class="icon-bell"></i>Activar avisos</button>
						{/if}
						<form method="POST" action="/salir">
							<button><i class="icon-log-out"></i>Salir</button>
						</form>
					</div>
				{/if}
			</span>
		</div>
	</header>

	<div class="cuerpo">
		<div class="lado" class:abierto={cajon}>
			<Carpetas
				datos={data}
				carpeta={data.carpeta}
				casilla={data.casilla}
				href={(o) => href(o)}
				onredactar={redactar}
			/>
		</div>
		{#if cajon}<button class="velo" aria-label="Cerrar el menú" onclick={() => (cajon = false)}
			></button>{/if}

		<main class="tarjeta vidrio">
			<div class="col-lista">
				<div class="tit-tel">{titulo}</div>
				<Lista
					filas={data.filas}
					pestana={data.pestana}
					contadores={data.contadores}
					casillas={data.casillas}
					casilla={data.casilla}
					{viendo}
					papelera={data.carpeta === 'papelera'}
					onvaciar={vaciar}
					activa={data.leido?.id ?? null}
					href={hrefFila}
					hrefPestana={(p) => href({ p })}
					onrefrescar={() => invalidate('correo:bandeja')}
					{vacio}
				/>
			</div>
			<div class="col-leer">
				{#if data.leido}
					{#key data.leido.id}
						<Lectura
							m={data.leido}
							marca={data.marca}
							envio={data.filas.find((f) => f.id === data.leido?.id)?.envio ?? null}
							casillas={data.casillas}
							{contactos}
							hrefImagenes={href({ m: data.leido.id, img: true })}
							onaccion={accion}
							onvolver={() => goto(href({}), { reset: false })}
							onenviado={enviado}
						/>
					{/key}
				{:else}
					<div class="nada">
						<i class="icon-mail-open"></i>
						<p>Elegí un correo para leerlo.</p>
					</div>
				{/if}
			</div>
		</main>
	</div>

	<button class="fab" onclick={redactar}><i class="icon-pencil"></i>Redactar</button>

	<div class="ventanas">
		{#each ventanas as v (v.clave)}
			<Editor
				b={v}
				casillas={data.casillas}
				{contactos}
				forma="ventana"
				oncerrar={() => cerrarVentana(v.clave)}
				onenviado={(p) => (cerrarVentana(v.clave), enviado(p))}
			/>
		{/each}
	</div>

	{#if aviso}<div class="toast" role="status">{aviso}</div>{/if}
	<Bienvenida />
</div>

<style>
	/* La app instalada dibuja debajo de la barra de estado (black-translucent): se respetan las
	   zonas seguras del teléfono. */
	.app {
		height: 100dvh;
		display: flex;
		flex-direction: column;
		overflow: hidden;
		padding-top: env(safe-area-inset-top);
		padding-left: env(safe-area-inset-left);
		padding-right: env(safe-area-inset-right);
	}
	.barra {
		position: relative;
		height: 72px;
		flex: none;
		display: flex;
		align-items: center;
		gap: 16px;
		padding: 0 20px;
	}
	.menu-tel {
		display: none;
		font-size: 22px;
		padding: 8px;
	}
	.marca {
		width: 236px;
		flex: none;
		display: flex;
		align-items: center;
		gap: 12px;
		padding-left: 6px;
	}
	.mt {
		line-height: 1.2;
	}
	.mc {
		font-size: 19px;
		font-weight: 600;
	}
	.mm {
		font-size: 14px;
		color: var(--mut);
	}
	.buscar {
		width: 720px;
		max-width: 100%;
		min-width: 0;
		height: 50px;
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 0 16px;
		border-radius: 12px;
		background: var(--srch);
		color: var(--mut);
		font-size: 17px;
	}
	.buscar:focus-within {
		box-shadow: 0 0 0 2px var(--acc);
	}
	.buscar i {
		font-size: 20px;
	}
	.buscar input {
		flex: 1;
		min-width: 0;
		border: 0;
		outline: 0;
		background: transparent;
		color: var(--fg);
		font-size: 17px;
	}
	.buscar input::placeholder {
		color: var(--mut);
	}
	.buscar a {
		color: var(--mut);
		display: inline-flex;
	}
	.der {
		margin-left: auto;
		display: flex;
		align-items: center;
		gap: 6px;
		color: var(--mut);
	}
	.ic {
		width: 42px;
		height: 42px;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		font-size: 21px;
		color: var(--mut);
	}
	.usuario-w {
		position: relative;
	}
	.av {
		width: 40px;
		height: 40px;
		margin-left: 6px;
		border-radius: 50%;
		background: oklch(0.68 calc(0.19 * var(--ac)) calc(var(--h) + 0));
		color: #fff;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 15px;
		font-weight: 600;
	}
	.pop {
		position: absolute;
		right: 0;
		top: 50px;
		z-index: 20;
		min-width: 260px;
		padding: 6px 0;
		border-radius: 12px;
		background: var(--pop);
		box-shadow:
			var(--popsh),
			0 0 0 1px var(--ls);
		color: var(--fg);
	}
	.pe {
		padding: 10px 16px;
		font-size: 14px;
		color: var(--mut);
		border-bottom: 1px solid var(--ls);
	}
	.pop button {
		width: 100%;
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 10px 16px;
		font-size: 16px;
	}
	.pop button:hover {
		background: var(--sel);
	}
	.cuerpo {
		flex: 1;
		min-height: 0;
		display: flex;
	}
	.lado {
		display: flex;
		min-height: 0;
	}
	.tarjeta {
		flex: 1;
		min-width: 0;
		margin: 0 16px 16px 0;
		display: grid;
		grid-template-columns: 460px minmax(0, 1fr);
		background-color: var(--card);
		background-image: var(--texC);
		background-size: 180px 180px;
		border-radius: 14px;
		overflow: hidden;
		box-shadow: var(--sh);
	}
	.col-lista,
	.col-leer {
		display: flex;
		flex-direction: column;
		min-height: 0;
		min-width: 0;
	}
	.col-lista > :global(*) {
		flex: 1;
	}
	.col-leer > :global(*) {
		flex: 1;
	}
	.tit-tel {
		display: none;
	}
	.nada {
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		color: var(--mut);
		gap: 10px;
	}
	.nada i {
		font-size: 36px;
		color: var(--t3);
	}
	.nada p {
		margin: 0;
	}
	.fab {
		display: none;
	}
	.velo {
		display: none;
	}
	.ventanas {
		position: fixed;
		right: 24px;
		bottom: 0;
		z-index: 30;
		display: flex;
		flex-direction: row-reverse;
		align-items: flex-end;
		gap: 16px;
		pointer-events: none;
	}
	.ventanas > :global(*) {
		pointer-events: auto;
	}
	.toast {
		position: fixed;
		left: 50%;
		bottom: calc(28px + env(safe-area-inset-bottom));
		transform: translateX(-50%);
		z-index: 40;
		padding: 12px 18px;
		border-radius: 12px;
		background: var(--toast);
		color: var(--toastT);
		font-size: 15px;
		font-weight: 500;
		box-shadow: var(--popsh);
	}
	@media (max-width: 1280px) {
		.tarjeta {
			grid-template-columns: 380px minmax(0, 1fr);
		}
		.marca {
			width: auto;
		}
	}
	@media (max-width: 1100px) {
		.mt {
			display: none;
		}
	}
	/* Teléfono y tablet: una pantalla a la vez; carpetas y casillas en el menú de ☰. */
	@media (max-width: 900px) {
		.barra {
			height: 64px;
			gap: 8px;
			padding: 0 16px;
		}
		.marca,
		.der .ic {
			display: none;
		}
		.menu-tel {
			display: inline-flex;
			position: absolute;
			left: 22px;
			z-index: 2;
			padding: 6px;
		}
		.buscar {
			flex: 1;
			width: auto;
			padding-left: 52px;
			border-radius: 999px;
		}
		.buscar > i {
			display: none;
		}
		.der {
			position: absolute;
			right: 22px;
		}
		.av {
			width: 36px;
			height: 36px;
			font-size: 14px;
		}
		.buscar {
			padding-right: 54px;
		}
		.lado {
			position: fixed;
			inset: 0 auto 0 0;
			z-index: 25;
			transform: translateX(-105%);
			transition: transform 0.2s ease;
			background-color: var(--side);
			background-image: var(--tex);
			padding-top: calc(16px + env(safe-area-inset-top));
			padding-bottom: env(safe-area-inset-bottom);
			max-width: 86vw;
		}
		.lado.abierto {
			transform: none;
		}
		.velo {
			display: block;
			position: fixed;
			inset: 0;
			z-index: 24;
			background: rgb(0 0 0 / 0.4);
		}
		.tarjeta {
			margin: 0;
			border-radius: 0;
			grid-template-columns: minmax(0, 1fr);
			box-shadow: none;
		}
		.leyendo .col-lista,
		.app:not(.leyendo) .col-leer {
			display: none;
		}
		.tit-tel {
			display: block;
			flex: none !important;
			padding: 12px 20px 0;
			font-size: 15px;
			font-weight: 600;
			color: var(--mut);
		}
		.fab {
			position: fixed;
			right: 20px;
			bottom: calc(24px + env(safe-area-inset-bottom));
			z-index: 20;
			display: inline-flex;
			align-items: center;
			gap: 10px;
			height: 56px;
			padding: 0 22px;
			border-radius: 16px;
			background: var(--acc);
			color: #fff;
			font-size: 17px;
			font-weight: 600;
			box-shadow:
				0 2px 0 oklch(0.45 calc(0.17 * var(--ac)) calc(var(--h) - 6)),
				0 8px 20px -8px oklch(0.45 calc(0.17 * var(--ac)) var(--h) / 0.55);
		}
		.fab i {
			font-size: 20px;
		}
		.leyendo .fab {
			display: none;
		}
		.ventanas {
			right: 0;
			left: 0;
			flex-direction: column;
		}
	}
</style>
