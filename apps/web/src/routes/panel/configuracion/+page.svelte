<script lang="ts">
	import { deslizable } from '#lib/panel/deslizable.ts';
	import { invalidateAll } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import CampoDireccion, { type EstadoDireccion } from '#lib/components/CampoDireccion.svelte';
	import Icon from '#lib/components/Icon.svelte';
	import Interruptor from '#lib/components/Interruptor.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { unsplash } from '#lib/landing/formato.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { diaDeMes, haceCorto, pesos } from '#lib/panel/formato.ts';
	import { AREAS, NOMBRE_ROL } from '#lib/panel/permisos.ts';
	import { PLANES } from '#lib/panel/planes.ts';
	import { HOST, MAX_CAMBIOS_POR_DIA, slugFinal } from '#lib/tienda/direccion.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/** Configuración (design/PanelOperacion.dc.html, sCfg): negocio, equipo, seguridad, avisos y plan. */
	let { data } = $props();
	/**
	 * Con la API real, los interruptores de Avisos se guardaban pero nada los usaba: los avisos se
	 * manejan en Notificaciones.
	 */
	const demo = $derived(!!page.data.demo);
	const ui = getUi();

	// Cada pestaña con el permiso que pide (docs/decisiones.md §37): la seguridad y las sesiones son
	// de cada persona, así que las ve cualquiera del equipo.
	const TODAS = [
		['negocio', I.phStorefrontFill, 'Negocio', 'tienda'],
		['equipo', I.phUsersThreeFill, 'Equipo y permisos', 'cuenta'],
		['seguridad', I.phShieldCheckFill, 'Seguridad', 'panel'],
		['avisos', I.phBellRingingFill, 'Notificaciones', 'tienda'],
		['mayorista', I.phPackageFill, 'Venta mayorista', 'catalogo'],
		['plan', I.phCrownSimpleFill, 'Plan', 'cuenta']
	] as const;
	type Tab = (typeof TODAS)[number][0];
	const permisos = $derived(page.data.acceso?.permisos ?? []);
	const TABS = $derived(TODAS.filter(([, , , p]) => permisos.includes(p)));
	// ?tab=seguridad: el link del mail «Entraron a tu cuenta» abre directo en las sesiones.
	const pedida = page.url.searchParams.get('tab');
	let elegida = $state<Tab | null>(null);
	const tab = $derived<Tab>(
		elegida ?? TABS.find(([k]) => k === pedida)?.[0] ?? TABS[0]?.[0] ?? 'seguridad'
	);

	async function pedir(url: string, method: string, cuerpo?: unknown) {
		const r = await fetch(url, {
			method,
			headers: cuerpo === undefined ? {} : { 'content-type': 'application/json' },
			body: cuerpo === undefined ? undefined : JSON.stringify(cuerpo)
		});
		if (!r.ok) {
			const m = ((await r.json().catch(() => null)) as { message?: string } | null)?.message;
			ui.toast({ k: 'err', title: 'No se pudo', desc: m ?? 'Probá de nuevo.' });
			return null;
		}
		return r;
	}

	// ---- Negocio ----
	const CAMPOS = [
		['nombre', 'Nombre de fantasía'],
		['razonSocial', 'Razón social'],
		['cuit', 'CUIT'],
		['condicion', 'Condición fiscal'],
		['direccion', 'Dirección'],
		['whatsapp', 'WhatsApp'],
		['email', 'Email'],
		['horario', 'Horario'],
		['alias', 'Alias para transferencias'],
		['cbu', 'CBU o CVU'],
		['titular', 'Titular de la cuenta']
	] as const;
	let negocio = $derived({ ...data.negocio });
	const cambiado = $derived(CAMPOS.some(([k]) => negocio[k] !== data.negocio[k]));
	async function guardarNegocio() {
		if (!(await pedir('/panel/configuracion/negocio', 'PUT', negocio))) return;
		await invalidateAll();
		ui.toast({ k: 'ok', title: 'Datos guardados', desc: 'Se ven en la tienda y en tus facturas' });
	}

	// ---- Dirección de la tienda (decisiones §38) ----
	const slugActual = $derived((page.data.tienda as { slug: string } | undefined)?.slug ?? '');
	let slug = $derived(slugActual);
	let estadoDir = $state<EstadoDireccion>('actual');
	let cambiando = $state(false);
	async function cambiarDireccion(nuevo: string, deshacer = false) {
		cambiando = true;
		const r = await fetch('/panel/configuracion/direccion', {
			method: 'PUT',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ slug: nuevo })
		}).catch(() => null);
		cambiando = false;
		const res = (await r?.json().catch(() => null)) as
			{ ok: true; slug: string; anterior: string } | { ok: false; mensaje: string } | null;
		if (!res?.ok) {
			ui.toast({
				k: 'err',
				title: 'No se pudo cambiar la dirección',
				desc: res && !res.ok ? res.mensaje : 'Probá de nuevo.'
			});
			return;
		}
		await invalidateAll();
		if (deshacer) {
			ui.toast({ k: 'ok', title: 'Volvió la dirección de antes', desc: `${HOST}/t/${res.slug}` });
			return;
		}
		const anterior = res.anterior;
		ui.undoable('Dirección cambiada', `Ahora es ${HOST}/t/${res.slug}`, () =>
			cambiarDireccion(anterior, true)
		);
	}

	// ---- Seguridad y avisos ----
	// Opciones de la persona (no de la tienda): las dos funcionan de verdad en la API.
	const SEGURIDAD = [
		[
			'avisoIngreso',
			I.phEnvelopeSimpleFill,
			'Avisarme por mail cuando entren a mi cuenta',
			'Con el dispositivo, el lugar y la hora'
		],
		[
			'inactivas',
			I.phClockCounterClockwiseFill,
			'Cerrar sesiones inactivas a los 7\u00a0días',
			'Recomendado para cajas compartidas'
		]
	] as const;
	const AVISOS = [
		['venta', 'Cada venta', 'Push y sonido de caja'],
		['chats', 'Chats que te esperan', 'Solo si Tincho no puede responder'],
		['stock', 'Stock bajo', 'Cuando quede menos de 1 semana'],
		['diario', 'Resumen diario por WhatsApp', 'Todos los días a las 21 h'],
		['ideas', 'Ideas de Tincho', 'Máximo 3 por día'],
		['novedades', 'Novedades de Tu Local Web', 'Una vez por mes']
	] as const;
	async function opcion(grupo: 'seguridad' | 'avisos', clave: string, activo: boolean) {
		if (await pedir(`/panel/configuracion/${grupo}/${clave}`, 'PUT', { activo }))
			await invalidateAll();
	}
	const SESION = {
		compu: I.phDesktopFill,
		celu: I.phDeviceMobileFill,
		tablet: I.phDeviceTabletFill,
		caja: I.phDesktopTowerFill
	} as const;
	const ultimaVez = (iso: string) => {
		const h = haceCorto(iso);
		return h === 'ayer' ? 'Última vez ayer' : `Última vez hace ${h}`;
	};
	// Cada dato entero en su renglón: sólo se corta en el «·».
	const duro = (t: string) => t.replaceAll(' ', '\u00a0');
	const otras = $derived(data.sesiones.filter((s) => !s.actual).length);
	async function cerrar(id: string, nombre: string) {
		if (!(await pedir(`/panel/sesiones/${id}`, 'DELETE'))) return;
		await invalidateAll();
		ui.toast({ k: 'ok', title: 'Sesión cerrada', desc: nombre });
	}
	async function cerrarOtras() {
		const r = await pedir('/panel/sesiones', 'DELETE');
		if (!r) return;
		const { cerradas } = (await r.json()) as { cerradas: number };
		await invalidateAll();
		ui.toast({
			k: 'ok',
			title: cerradas === 1 ? 'Cerramos 1 sesión' : `Cerramos ${cerradas} sesiones`,
			desc: 'Sólo queda abierta esta'
		});
	}

	// ---- Venta mayorista (§49) ----
	// El mínimo se edita en el formulario; prenderla o apagarla se guarda al toque (con Deshacer).
	const numeroOVacio = (s: string) => {
		const n = Number(s.replace(/\D/g, ''));
		return s.trim() && n > 0 ? n : null;
	};
	let minimo = $derived({ unidades: data.mayorista.unidades, monto: data.mayorista.monto });
	const minimoCambiado = $derived(
		minimo.unidades !== data.mayorista.unidades || minimo.monto !== data.mayorista.monto
	);
	async function guardarMayorista(activa: boolean, m = minimo, deshacer = false) {
		const antes = { ...data.mayorista };
		const r = await pedir('/panel/configuracion/mayorista', 'PUT', { activa, ...m });
		if (!r) {
			await invalidateAll();
			return;
		}
		await invalidateAll();
		if (deshacer) return;
		if (antes.activa === activa) {
			ui.toast({ k: 'ok', title: 'Mínimo guardado', desc: 'La tienda ya usa el nuevo mínimo' });
			return;
		}
		ui.undoable(
			activa ? 'Venta mayorista prendida' : 'Venta mayorista apagada',
			activa ? 'Los precios mayoristas ya se ven en tu tienda' : 'Tu tienda vende sólo minorista',
			() => guardarMayorista(antes.activa, { unidades: antes.unidades, monto: antes.monto }, true)
		);
	}

	// ---- Plan ----
	async function elegir(id: (typeof PLANES)[number]['id'], nombre: string) {
		if (id === data.plan.actual) return;
		const r0 = await pedir('/panel/plan', 'PUT', { plan: id });
		if (!r0) return;
		// Plan pago con el abono real: se carga la tarjeta en Mercado Pago.
		if (r0.status === 200) {
			location.href = ((await r0.json()) as { url: string }).url;
			return;
		}
		await invalidateAll();
		ui.toast({
			k: 'ok',
			title: `Plan ${nombre} elegido`,
			desc: data.plan.enPrueba
				? 'Se cobra cuando termina tu prueba · sin permanencia'
				: 'Cobramos la diferencia prorrateada'
		});
	}
</script>

<Seo
	title="Configuración · Panel"
	description="Configuración de la tienda"
	path="/panel/configuracion"
	noindex
/>

<div class="pagina">
	<div use:anim={{ t: 'rise' }}>
		<h1>Configuración</h1>
		<div class="c-s">Tu negocio, tu equipo y tu seguridad</div>
	</div>

	<div class="grilla" use:anim={{ t: 'rise', delay: 60 }}>
		<nav class="tabs" aria-label="Secciones" use:deslizable>
			{#each TABS as [id, icono, t] (id)}
				<button
					type="button"
					aria-current={tab === id ? 'page' : undefined}
					onclick={() => (elegida = id)}><Icon icon={icono} size={17} />{t}</button
				>
			{/each}
			<!-- Envíos y retiro tiene su pantalla (también se llega desde Envíos). -->
			{#if permisos.includes('tienda')}
				<a href={resolve('/panel/envios/configurar')}
					><Icon icon={I.phTruckFill} size={17} />Envíos y retiro<Icon
						icon={I.phArrowRightBold}
						size={12}
					/></a
				>
			{/if}
		</nav>

		<div class="contenido">
			{#if tab === 'negocio'}
				<form
					class="tarjeta campos"
					onsubmit={(e) => {
						e.preventDefault();
						void guardarNegocio();
					}}
				>
					{#each CAMPOS as [k, t] (k)}
						<label
							>{t}<input
								value={negocio[k]}
								oninput={(e) => (negocio = { ...negocio, [k]: e.currentTarget.value })}
								maxlength={k === 'direccion' ? 160 : 120}
							/></label
						>
					{/each}
					{#if cambiado}
						<div class="guardar">
							<button
								type="button"
								class="btn btn-secondary"
								onclick={() => (negocio = { ...data.negocio })}>Descartar</button
							>
							<button class="btn btn-primary">Guardar cambios</button>
						</div>
					{/if}
				</form>
				<form
					class="tarjeta direccion"
					onsubmit={(e) => {
						e.preventDefault();
						const nuevo = slugFinal(slug);
						if (nuevo && nuevo !== slugActual && estadoDir !== 'mal') void cambiarDireccion(nuevo);
					}}
				>
					<div class="t-cab"><span>Dirección de tu tienda</span></div>
					<p class="d-txt">
						Es la que va en tus links, el QR y lo que compartas. Si la cambiás, los links viejos
						siguen andando: quien entre por la dirección anterior llega a la nueva.
					</p>
					<CampoDireccion
						id="cfg-direccion"
						etiqueta="Nueva dirección"
						bind:slug
						bind:estado={estadoDir}
						actual={slugActual}
					/>
					<p class="d-nota">
						Podés cambiarla hasta {MAX_CAMBIOS_POR_DIA} veces por día. Una dirección que dejás no la puede
						usar otra tienda.
					</p>
					{#if slugFinal(slug) !== slugActual}
						<div class="guardar">
							<button type="button" class="btn btn-secondary" onclick={() => (slug = slugActual)}
								>Descartar</button
							>
							<button
								class="btn btn-primary"
								disabled={cambiando || estadoDir === 'mal' || estadoDir === 'revisando'}
								>{cambiando ? 'Cambiando…' : 'Cambiar dirección'}</button
							>
						</div>
					{/if}
				</form>
			{:else if tab === 'equipo'}
				<div class="tarjeta">
					<div class="t-cab">
						<span
							>Equipo · {data.equipo.length}
							{data.equipo.length === 1 ? 'persona' : 'personas'}</span
						>
						<!-- Se invita desde Usuarios y permisos (un solo lugar, con reenviar y cancelar). -->
						<a class="oscuro" href={resolve('/panel/[seccion]', { seccion: 'usuarios-y-permisos' })}
							>Invitar</a
						>
					</div>
					<div class="miembros">
						{#each data.equipo as m (m.id)}
							<div class="miembro" class:pendiente={m.pendiente}>
								{#if m.foto}<img
										src={unsplash(m.foto, 76, '&h=76&crop=faces')}
										alt=""
									/>{:else}<span class="av">{m.nombre[0].toUpperCase()}</span>{/if}
								<span class="m-txt"
									><span class="m-n">{m.nombre}{m.vos ? ' (vos)' : ''}</span><span class="m-e"
										>{m.pendiente
											? m.vence
												? 'Invitación enviada'
												: 'Invitación sin mandar'
											: m.contacto}</span
									></span
								>
								<span class="permisos"
									>{#each AREAS[m.rol] as p (p)}<span>{p}</span>{/each}</span
								>
								<span class="rol" class:duenio={m.rol === 'duenio'}>{NOMBRE_ROL[m.rol]}</span>
								{#if m.dobleFactor}<span title="Doble verificación activa"
										><Icon icon={I.phShieldCheckFill} size={17} style="color:#237550" /></span
									>{/if}
							</div>
						{/each}
					</div>
					<a class="mas" href={resolve('/panel/[seccion]', { seccion: 'usuarios-y-permisos' })}
						>Roles y permisos de cada uno <Icon icon={I.phArrowRightBold} size={13} /></a
					>
				</div>
			{:else if tab === 'seguridad'}
				<div class="tarjeta lista">
					{#each SEGURIDAD as [k, icono, t, d] (k)}
						<div class="opcion">
							<span class="o-ico"><Icon icon={icono} size={18} /></span>
							<span class="o-txt"><span class="o-t">{t}</span><span class="o-d">{d}</span></span>
							<Interruptor
								checked={data.seguridad[k]}
								label={t}
								onchange={(v) => opcion('seguridad', k, v)}
							/>
						</div>
					{/each}
					<div class="sesiones-cab">
						<div class="sesiones-t">Sesiones activas</div>
						{#if otras > 0}<button type="button" class="s-btn" onclick={cerrarOtras}
								>Cerrar todas las demás</button
							>{/if}
					</div>
					{#each data.sesiones as s (s.id)}
						{@const donde = [s.ubicacion, s.ip].filter(Boolean).join(' · ')}
						<div class="sesion">
							<Icon icon={SESION[s.tipo]} size={18} style="color:var(--mu)" />
							<span class="s-t">
								<span class="o-t">{s.nombre}</span>
								{#if donde}<span class="o-d">{donde}</span>{/if}
								<span class="o-d"
									>{duro(`Abierta el ${diaDeMes(new Date(s.creada))}`)} · {duro(
										s.actual ? 'En uso ahora' : ultimaVez(s.cuando)
									)}</span
								>
							</span>
							{#if s.actual}<span class="esta">Esta</span>{:else}<button
									type="button"
									class="s-btn"
									onclick={() => cerrar(s.id, s.nombre)}>Cerrar</button
								>{/if}
						</div>
					{/each}
				</div>
			{:else if tab === 'avisos'}
				<div class="tarjeta lista">
					{#if !demo}
						<p class="nota-real">
							Qué te avisamos y por dónde se elige en
							<a href={resolve('/panel/[seccion]', { seccion: 'notificaciones' })}>Notificaciones</a
							>.
						</p>
					{/if}
					{#each demo ? AVISOS : [] as [k, t, d] (k)}
						<div class="opcion">
							<span class="o-txt"><span class="o-t">{t}</span><span class="o-d">{d}</span></span>
							<Interruptor
								checked={data.avisos[k]}
								label={t}
								onchange={(v) => opcion('avisos', k, v)}
							/>
						</div>
					{/each}
				</div>
			{:else if tab === 'mayorista'}
				<div class="tarjeta lista">
					<div class="opcion">
						<span class="o-ico"><Icon icon={I.phPackageFill} size={18} /></span>
						<span class="o-txt"
							><span class="o-t">Venta mayorista</span><span class="o-d"
								>Con el carrito al mínimo, los productos con precio mayorista pasan a ese
								precio&nbsp;solos</span
							></span
						>
						<Interruptor
							checked={data.mayorista.activa}
							label="Venta mayorista"
							onchange={(v) => guardarMayorista(v, data.mayorista)}
						/>
					</div>
				</div>
				<form
					class="tarjeta campos"
					onsubmit={(e) => {
						e.preventDefault();
						void guardarMayorista(data.mayorista.activa);
					}}
				>
					<div class="t-cab"><span>Mínimo del carrito</span></div>
					<label
						>Unidades<input
							inputmode="numeric"
							placeholder="Sin mínimo"
							value={minimo.unidades ?? ''}
							oninput={(e) =>
								(minimo = { ...minimo, unidades: numeroOVacio(e.currentTarget.value) })}
						/></label
					>
					<label
						>Monto<input
							inputmode="numeric"
							placeholder="Sin mínimo"
							value={minimo.monto === null ? '' : pesos(minimo.monto)}
							onchange={(e) => (minimo = { ...minimo, monto: numeroOVacio(e.currentTarget.value) })}
						/></label
					>
					<p class="d-nota">
						Poné al menos uno; con los dos, el carrito tiene que cumplir los dos. El monto se cuenta
						con los precios de siempre. El precio mayorista de cada producto se carga en su ficha,
						en&nbsp;Productos.
					</p>
					{#if minimoCambiado}
						<div class="guardar">
							<button
								type="button"
								class="btn btn-secondary"
								onclick={() =>
									(minimo = { unidades: data.mayorista.unidades, monto: data.mayorista.monto })}
								>Descartar</button
							>
							<button class="btn btn-primary">Guardar mínimo</button>
						</div>
					{/if}
				</form>
			{:else}
				<div class="planes">
					{#each PLANES as p (p.id)}
						{@const actual = data.plan.actual === p.id}
						<div class="plan" class:actual>
							<div class="p-cab">
								<span>{p.nombre}</span>{#if actual}<span class="tuyo"
										>{data.plan.enPrueba ? 'Tu prueba' : 'Tu plan'}</span
									>{/if}
							</div>
							<div class="precio">
								{pesos(p.precio)}<span> /mes</span>
							</div>
							<div class="incluye">
								{#each p.incluye as f (f)}<span
										><Icon icon={I.phCheckBold} size={13} style="color:#237550" />{f}</span
									>{/each}
							</div>
							<button type="button" onclick={() => elegir(p.id, p.nombre)}
								>{actual ? `Quedarme en ${p.nombre}` : `Elegir ${p.nombre}`}</button
							>
						</div>
					{/each}
				</div>
			{/if}
		</div>
	</div>
</div>

<style>
	.pagina {
		display: flex;
		flex-direction: column;
		gap: 18px;
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
	.grilla {
		display: grid;
		grid-template-columns: 240px minmax(0, 1fr);
		gap: 16px;
		align-items: start;
	}
	.tabs {
		padding: 8px;
		border-radius: 22px;
		background: var(--sf);
		box-shadow: var(--shd);
		display: flex;
		flex-direction: column;
		gap: 2px;
		overflow-x: auto;
		scrollbar-width: none;
	}
	.tabs button,
	.tabs a {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 10px 12px;
		border-radius: 14px;
		border: 0;
		background: transparent;
		color: var(--tx);
		font:
			500 14px 'Instrument Sans',
			sans-serif;
		text-align: left;
		white-space: nowrap;
	}
	.tabs button :global(svg),
	.tabs a :global(svg) {
		color: var(--mu);
	}
	.tabs a {
		text-decoration: none;
	}
	.tabs a :global(svg:last-child) {
		margin-left: auto;
	}
	.tabs button[aria-current='page'] {
		background: var(--chip);
		font-weight: 600;
	}
	.tabs button[aria-current='page'] :global(svg) {
		color: #b0441d;
	}
	.contenido {
		display: flex;
		flex-direction: column;
		gap: 12px;
		min-width: 0;
	}
	.tarjeta {
		padding: 18px;
		border-radius: 22px;
		background: var(--sf);
		box-shadow: var(--shd);
	}
	.direccion {
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.d-txt,
	.d-nota {
		margin: 0;
		font-size: 13.5px;
		line-height: 1.45;
		color: var(--mu);
	}
	.d-nota {
		font-size: 12.5px;
	}
	.campos .t-cab,
	.campos .d-nota {
		grid-column: 1 / -1;
	}
	.campos {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
		gap: 12px;
	}
	.campos label {
		display: flex;
		flex-direction: column;
		gap: 5px;
		font-size: 12.5px;
		font-weight: 600;
		color: var(--mu);
	}
	.campos input {
		padding: 11px 12px;
		border-radius: 12px;
		border: 1px solid var(--ln2);
		background: var(--bg);
		font:
			500 14.5px 'Instrument Sans',
			sans-serif;
		color: var(--tx);
		outline: none;
	}
	.guardar {
		grid-column: 1 / -1;
		display: flex;
		justify-content: flex-end;
		gap: 8px;
	}
	.t-cab {
		display: flex;
		align-items: center;
		justify-content: space-between;
		font-size: 16px;
		font-weight: 600;
	}
	.oscuro {
		text-decoration: none;
		padding: 9px 14px;
		border-radius: 999px;
		border: 0;
		background: var(--tx);
		color: var(--sf);
		font:
			600 13px 'Instrument Sans',
			sans-serif;
	}
	.miembros {
		margin-top: 10px;
		display: flex;
		flex-direction: column;
	}
	.miembro {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 11px 0;
		border-bottom: 1px solid var(--ln);
		flex-wrap: wrap;
	}
	.miembro.pendiente {
		opacity: 0.6;
	}
	.miembro img,
	.av {
		width: 38px;
		height: 38px;
		border-radius: 50%;
		object-fit: cover;
	}
	.av {
		display: grid;
		place-items: center;
		background: var(--chip);
		font-weight: 600;
	}
	.m-txt {
		flex: 1 1 160px;
		min-width: 0;
	}
	.m-n {
		display: block;
		font-size: 14px;
		font-weight: 600;
		overflow-wrap: anywhere;
	}
	.m-e {
		display: block;
		font-size: 12.5px;
		color: var(--mu);
		overflow-wrap: anywhere;
	}
	.permisos {
		display: flex;
		gap: 4px;
		flex-wrap: wrap;
	}
	.permisos span {
		font-size: 11.5px;
		font-weight: 600;
		padding: 3px 8px;
		border-radius: 999px;
		background: var(--chip);
	}
	.rol {
		font-size: 12.5px;
		font-weight: 600;
		padding: 5px 11px;
		border-radius: 999px;
		background: var(--chip);
	}
	.rol.duenio {
		background: rgba(210, 88, 47, 0.12);
		color: #b0441d;
	}
	.mas {
		margin-top: 12px;
		display: inline-flex;
		align-items: center;
		gap: 6px;
		font-size: 13.5px;
		font-weight: 600;
		color: #b0441d;
		text-decoration: none;
	}
	.lista {
		display: flex;
		flex-direction: column;
		gap: 4px;
	}
	.opcion {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 12px 0;
		border-bottom: 1px solid var(--ln);
	}
	.o-ico {
		width: 36px;
		height: 36px;
		flex: none;
		border-radius: 11px;
		background: rgba(47, 143, 99, 0.12);
		color: #237550;
		display: grid;
		place-items: center;
	}
	.o-txt {
		flex: 1;
		min-width: 0;
	}
	.o-t {
		display: block;
		font-size: 14.5px;
		font-weight: 600;
	}
	.o-d {
		display: block;
		font-size: 12.5px;
		color: var(--mu);
	}
	.nota-real {
		margin: 0 0 6px;
		font-size: 13.5px;
		line-height: 1.45;
		color: var(--mu);
		text-wrap: pretty;
	}
	.nota-real a {
		color: var(--terracota);
		font-weight: 600;
	}
	.sesiones-cab {
		margin-top: 10px;
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 10px;
		flex-wrap: wrap;
	}
	.sesiones-t {
		font-size: 14px;
		font-weight: 600;
	}
	.sesion {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 9px 0;
		font-size: 13.5px;
		border-bottom: 1px solid var(--ln);
	}
	.sesion:last-child {
		border-bottom: 0;
	}
	.s-t {
		flex: 1;
		min-width: 0;
		overflow-wrap: anywhere;
	}
	.esta {
		font-size: 12px;
		font-weight: 600;
		color: #237550;
	}
	.s-btn {
		min-height: 44px;
		padding: 0 4px;
		white-space: nowrap;
		border: 0;
		background: transparent;
		font:
			600 12.5px 'Instrument Sans',
			sans-serif;
		color: #b83a2e;
	}
	.planes {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
		gap: 12px;
	}
	.plan {
		padding: 18px;
		border-radius: 22px;
		background: var(--sf);
		color: var(--tx);
		box-shadow: var(--shd);
		border: 2px solid transparent;
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.plan.actual {
		background: #2a140c;
		color: #fbf7f2;
		border-color: #d2582f;
	}
	.p-cab {
		display: flex;
		justify-content: space-between;
		align-items: center;
		font-size: 17px;
		font-weight: 600;
	}
	.tuyo {
		font-size: 11.5px;
		font-weight: 700;
		padding: 3px 9px;
		border-radius: 999px;
		background: #b0441d;
		color: #fff;
	}
	.precio {
		font-size: 30px;
		font-weight: 600;
		letter-spacing: -0.04em;
	}
	.precio span {
		font-size: 13px;
		font-weight: 500;
		opacity: 0.7;
		letter-spacing: 0;
	}
	.incluye {
		display: flex;
		flex-direction: column;
		gap: 6px;
		font-size: 13.5px;
	}
	.incluye span {
		display: flex;
		gap: 7px;
		align-items: center;
	}
	.plan button {
		margin-top: auto;
		padding: 11px;
		border-radius: 999px;
		border: 0;
		background: var(--chip);
		color: var(--tx);
		font:
			600 13.5px 'Instrument Sans',
			sans-serif;
	}
	.plan.actual button {
		background: #b0441d;
		color: #fff;
	}
	@media (max-width: 899px) {
		.grilla {
			grid-template-columns: minmax(0, 1fr);
		}
		.tabs {
			flex-direction: row;
		}
	}
</style>
