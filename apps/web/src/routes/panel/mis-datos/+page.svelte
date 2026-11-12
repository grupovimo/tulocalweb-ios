<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import Icon from '#lib/components/Icon.svelte';
	import Modal from '#lib/components/Modal.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { plural } from '#lib/panel/formato.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/** Mis datos (design/PanelCuenta.dc.html, sData): exportar, borrar inactivos y cerrar la cuenta. */
	let { data } = $props();
	const ui = getUi();

	const EXPORTS = $derived([
		['productos', I.phTag, 'Productos', plural(data.conteos.productos, 'producto', 'productos')],
		['clientes', I.phUsersThree, 'Clientes', plural(data.conteos.clientes, 'cliente', 'clientes')],
		['pedidos', I.phPackage, 'Pedidos', plural(data.conteos.pedidos, 'pedido', 'pedidos')],
		[
			'facturas',
			I.phReceipt,
			'Facturas',
			plural(data.conteos.facturas, 'comprobante', 'comprobantes')
		],
		['resenas', I.phStar, 'Reseñas', plural(data.conteos.resenas, 'reseña', 'reseñas')],
		// Con la API todavía no hay conversaciones (WhatsApp e Instagram no se conectan): no se ofrece
		// bajar un archivo que sólo tendría el encabezado.
		...(page.data.demo
			? ([['conversaciones', I.phChatsCircle, 'Conversaciones', 'WhatsApp e Instagram']] as const)
			: [])
	] as const);
	const NOMBRES: Record<string, string> = {
		productos: 'Productos',
		clientes: 'Clientes',
		pedidos: 'Pedidos',
		facturas: 'Facturas',
		resenas: 'Reseñas',
		conversaciones: 'Conversaciones',
		todo: 'Exportación completa'
	};
	const cuando = (iso: string) =>
		new Date(iso).toLocaleString('es-AR', {
			day: 'numeric',
			month: 'long',
			hour: '2-digit',
			minute: '2-digit',
			timeZone: 'America/Argentina/Buenos_Aires'
		});
	/**
	 * Pide el archivo y lo baja recién si salió bien: si la exportación falla, avisa en vez de bajar
	 * la página de error como .html. Después refresca "Exportaciones recientes".
	 */
	let bajando = $state('');
	async function bajar(tipo: string, formato: 'csv' | 'excel') {
		if (bajando) return;
		bajando = `${tipo}-${formato}`;
		try {
			const r = await fetch(`/panel/mis-datos/exportar/${tipo}?formato=${formato}`);
			if (!r.ok) throw new Error(String(r.status));
			const nombre =
				/filename="([^"]+)"/.exec(r.headers.get('content-disposition') ?? '')?.[1] ??
				`${tipo}.${formato === 'excel' ? 'xlsx' : 'csv'}`;
			const url = URL.createObjectURL(await r.blob());
			const a = document.createElement('a');
			a.href = url;
			a.download = nombre;
			document.body.appendChild(a);
			a.click();
			a.remove();
			setTimeout(() => URL.revokeObjectURL(url), 10_000);
			ui.toast({
				k: 'ok',
				title:
					tipo === 'todo'
						? 'Exportación completa'
						: `${NOMBRES[tipo]} en ${formato === 'csv' ? 'CSV' : 'Excel'}`,
				desc: tipo === 'todo' ? 'Un ZIP con todo en CSV' : `Se bajó ${nombre}`
			});
			await invalidateAll();
		} catch {
			ui.toast({
				k: 'err',
				title: 'No se pudo exportar',
				desc: 'Probá de nuevo en un rato. Si sigue, escribinos.'
			});
		} finally {
			bajando = '';
		}
	}
	async function borrarInactivos() {
		const r = await fetch('/panel/mis-datos/inactivos', { method: 'DELETE' });
		const { clientes } = (await r.json()) as { clientes: number };
		await invalidateAll();
		ui.toast({
			k: 'warn',
			title: `${plural(clientes, 'cliente se borra', 'clientes se borran')} en 7 días`,
			desc: 'Podés frenarlo hasta entonces',
			action: 'Deshacer',
			onAction: async () => {
				await fetch('/panel/mis-datos/inactivos', { method: 'POST' });
				await invalidateAll();
			}
		});
	}

	// ---- Cerrar la cuenta ----
	let cerrando = $state(false);
	let confirmacion = $state('');
	const tienda = $derived(data.tienda);
	function abrirCerrar() {
		confirmacion = '';
		cerrando = true;
	}
	const ok = $derived(!!tienda && confirmacion.trim().toLowerCase() === tienda.toLowerCase());
	async function cerrar() {
		if (!ok) return;
		const r = await fetch('/panel/mis-datos/cerrar', {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ confirmacion })
		});
		if (!r.ok)
			return ui.toast({ k: 'err', title: 'No se cerró', desc: 'Revisá el nombre de la tienda.' });
		cerrando = false;
		// Carga completa: sin sandbox, cualquier pedido del panel que siga en camino redirige a /demo.
		location.assign(resolve('/'));
	}

	// ---- Borrar mi cuenta (la persona; Google Play pide esta página) ----
	let borrando = $state(false);
	let confirmaEmail = $state('');
	const email = $derived(data.email);
	const okEmail = $derived(!!email && confirmaEmail.trim().toLowerCase() === email.toLowerCase());
	function abrirBorrar() {
		confirmaEmail = '';
		borrando = true;
	}
	async function borrarCuenta() {
		if (!okEmail) return;
		const r = await fetch('/panel/mis-datos/borrar-cuenta', {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ confirmacion: email })
		});
		if (!r.ok)
			return ui.toast({ k: 'err', title: 'No se borró', desc: 'Probá de nuevo en un rato.' });
		borrando = false;
		location.assign(resolve('/ingresar'));
	}
</script>

<Seo title="Mis datos · Panel" description="Tus datos" path="/panel/mis-datos" noindex />

<div class="pagina">
	<div class="cab" use:anim={{ t: 'rise' }}>
		<div class="c-txt">
			<h1>Mis datos</h1>
			<div class="c-s">
				Todo lo que guardaste en Tu Local Web es tuyo. Llevátelo cuando quieras.
			</div>
		</div>
		<button
			type="button"
			class="todo"
			disabled={!!bajando}
			aria-busy={bajando === 'todo-csv'}
			onclick={() => bajar('todo', 'csv')}
			><Icon icon={I.phArchiveBold} size={15} />Exportar todo</button
		>
	</div>

	<div class="exports" use:anim={{ t: 'rise', delay: 50 }}>
		{#each EXPORTS as [id, icono, t, d] (id)}
			<div class="export">
				<div class="e-cab">
					<span class="e-ico"><Icon icon={icono} size={19} /></span>
					<div>
						<b>{t}</b>
						<div class="e-d">{d}</div>
					</div>
				</div>
				<div class="e-acc">
					<button
						type="button"
						disabled={!!bajando}
						aria-busy={bajando === `${id}-csv`}
						onclick={() => bajar(id, 'csv')}>CSV</button
					>
					<button
						type="button"
						disabled={!!bajando}
						aria-busy={bajando === `${id}-excel`}
						onclick={() => bajar(id, 'excel')}>Excel</button
					>
				</div>
			</div>
		{/each}
	</div>

	<div class="dos" use:anim={{ t: 'rise', delay: 100 }}>
		<div class="tarjeta">
			<div class="t">Exportaciones recientes</div>
			<div class="recientes">
				{#each data.recientes as r (r.en)}
					<div>
						<Icon icon={I.phCheckCircleFill} size={18} style="color:#237550" /><span
							>{NOMBRES[r.tipo] ?? r.tipo}{r.tipo !== 'todo' && r.formato
								? ` · ${r.formato === 'excel' ? 'Excel' : 'CSV'}`
								: ''}<span class="rd">{cuando(r.en)}</span></span
						>
					</div>
				{:else}
					<div class="vacio">Todavía no exportaste nada.</div>
				{/each}
			</div>
		</div>
		<div class="tarjeta">
			<div class="t">Tus derechos sobre tus datos</div>
			<div class="derechos">
				Podés pedir acceso, corrección o borrado de tus datos personales según la Ley 25.326. Más
				info en la
				<a href={resolve('/(sitio)/legales/[slug]', { slug: 'privacidad' })}
					>política de privacidad</a
				>
				y el
				<a href={resolve('/(sitio)/legales/[slug]', { slug: 'tratamiento-de-datos' })}
					>acuerdo de tratamiento de datos</a
				>.
			</div>
		</div>
	</div>

	<div class="peligro" use:anim={{ t: 'rise', delay: 150 }}>
		<div class="p-cab">
			<Icon icon={I.phWarningOctagonFill} size={22} style="color:#B83A2E" /><b>Zona de peligro</b>
		</div>
		<div class="p-filas">
			<div class="p-fila">
				<div class="p-txt">
					<b>Borrar clientes inactivos</b>
					<div>
						{data.borrado
							? `${plural(data.borrado.clientes, 'cliente se borra', 'clientes se borran')} el ${cuando(data.borrado.el)}`
							: data.inactivos
								? `Elimina los datos de ${plural(data.inactivos, 'cliente', 'clientes')} sin compras en 3 años`
								: 'No hay clientes sin compras en 3 años'}
					</div>
				</div>
				{#if data.borrado}
					<button
						type="button"
						class="borde"
						onclick={async () => {
							await fetch('/panel/mis-datos/inactivos', { method: 'POST' });
							await invalidateAll();
						}}>Frenar</button
					>
				{:else}
					<button type="button" class="borde" disabled={!data.inactivos} onclick={borrarInactivos}
						>Borrar</button
					>
				{/if}
			</div>
			<div class="p-fila">
				<div class="p-txt">
					<b>Cerrar la cuenta y borrar la tienda</b>
					<div>Se cancela el plan y a los 90 días borramos todo, salvo lo contable</div>
				</div>
				<button type="button" class="rojo" onclick={abrirCerrar}>Cerrar cuenta</button>
			</div>
			{#if email}
				<div class="p-fila">
					<div class="p-txt">
						<b>Borrar mi cuenta</b>
						<div>Tu usuario se desactiva hoy y tus datos personales se borran a los 30 días</div>
					</div>
					<button type="button" class="rojo" onclick={abrirBorrar}>Borrar mi cuenta</button>
				</div>
			{/if}
		</div>
	</div>
</div>

<Modal bind:open={cerrando} titulo="Cerrar la cuenta" ancho={460} padding={24}>
	<span class="m-ico"><Icon icon={I.phWarningOctagonFill} size={26} style="color:#B83A2E" /></span>
	<div class="m-t">¿Cerrar {tienda}?</div>
	<div class="m-d">
		La tienda deja de estar online hoy. Tenés 90 días para arrepentirte; después borramos productos,
		clientes y pedidos. Te recomendamos
		<b>exportar todo</b> antes.
	</div>
	<div class="m-c">Escribí <b>{tienda.toLowerCase()}</b> para confirmar</div>
	<input
		class="m-in"
		bind:value={confirmacion}
		placeholder={tienda.toLowerCase()}
		aria-label="Nombre de la tienda"
	/>
	<div class="m-acc">
		<button type="button" class="b-sec" onclick={() => (cerrando = false)}>No, volver</button>
		<button type="button" class="b-roja" disabled={!ok} onclick={cerrar}>Cerrar cuenta</button>
	</div>
</Modal>

<Modal bind:open={borrando} titulo="Borrar mi cuenta" ancho={460} padding={24}>
	<span class="m-ico"><Icon icon={I.phWarningOctagonFill} size={26} style="color:#B83A2E" /></span>
	<div class="m-t">¿Borrar tu cuenta?</div>
	<div class="m-d">
		Tu cuenta se desactiva ahora y tus datos personales se borran a los 30 días. Las tiendas donde
		sos la única dueña se cierran. En esos 30 días la recuperás escribiendo a
		<b>hola@tulocalweb.com.ar</b>; después no se puede deshacer.
	</div>
	<div class="m-c">Escribí <b>{email}</b> para confirmar</div>
	<input class="m-in" bind:value={confirmaEmail} placeholder={email} aria-label="Tu email" />
	<div class="m-acc">
		<button type="button" class="b-sec" onclick={() => (borrando = false)}>No, volver</button>
		<button type="button" class="b-roja" disabled={!okEmail} onclick={borrarCuenta}
			>Borrar mi cuenta</button
		>
	</div>
</Modal>

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
	.todo {
		display: flex;
		align-items: center;
		gap: 8px;
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
	.exports {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 240px), 1fr));
		gap: 10px;
	}
	.export {
		padding: 16px;
		border-radius: 20px;
		background: var(--sf);
		box-shadow: var(--shd);
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.e-cab {
		display: flex;
		align-items: center;
		gap: 10px;
	}
	.e-cab b {
		font-size: 14.5px;
	}
	.e-ico {
		width: 38px;
		height: 38px;
		border-radius: 12px;
		background: var(--chip);
		display: grid;
		place-items: center;
		flex: none;
	}
	.e-d {
		font-size: 12.5px;
		color: var(--mu);
	}
	.e-acc {
		display: flex;
		gap: 6px;
	}
	.e-acc button:disabled,
	.todo:disabled {
		opacity: 0.6;
	}
	.e-acc button {
		flex: 1;
		padding: 8px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: transparent;
		color: var(--tx);
		font:
			600 12.5px 'Instrument Sans',
			sans-serif;
	}
	.dos {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 320px), 1fr));
		gap: 12px;
		align-items: start;
	}
	.tarjeta {
		padding: 18px;
		border-radius: 22px;
		background: var(--sf);
		box-shadow: var(--shd);
	}
	.t {
		font-size: 15.5px;
		font-weight: 600;
	}
	.recientes {
		margin-top: 10px;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.recientes > div {
		display: flex;
		align-items: center;
		gap: 10px;
		font-size: 13.5px;
	}
	.rd {
		display: block;
		font-size: 12px;
		color: var(--mu);
	}
	.vacio {
		color: var(--mu);
	}
	.derechos {
		margin-top: 8px;
		font-size: 13.5px;
		line-height: 1.55;
		color: var(--mu);
	}
	.derechos a {
		color: #b0441d;
		font-weight: 600;
	}
	.peligro {
		padding: 20px;
		border-radius: 24px;
		border: 1.5px solid rgba(184, 58, 46, 0.3);
		background: rgba(184, 58, 46, 0.04);
	}
	.p-cab {
		display: flex;
		align-items: center;
		gap: 10px;
		font-size: 16px;
		color: #b83a2e;
	}
	.p-filas {
		margin-top: 12px;
		display: flex;
		flex-direction: column;
		gap: 2px;
		border-radius: 16px;
		overflow: hidden;
	}
	.p-fila {
		display: flex;
		align-items: center;
		gap: 12px;
		flex-wrap: wrap;
		padding: 14px;
		background: var(--sf);
	}
	.p-txt {
		flex: 1;
		min-width: 220px;
	}
	.p-txt b {
		font-size: 14.5px;
	}
	.p-txt div {
		font-size: 12.5px;
		color: var(--mu);
	}
	.borde {
		padding: 9px 14px;
		border-radius: 999px;
		border: 1px solid rgba(184, 58, 46, 0.4);
		background: transparent;
		color: #b83a2e;
		font:
			600 13px 'Instrument Sans',
			sans-serif;
	}
	.borde:disabled {
		opacity: 0.45;
	}
	.rojo {
		padding: 9px 14px;
		border-radius: 999px;
		border: 0;
		background: #b83a2e;
		color: #fff;
		font:
			600 13px 'Instrument Sans',
			sans-serif;
	}
	.m-ico {
		width: 48px;
		height: 48px;
		border-radius: 16px;
		background: rgba(184, 58, 46, 0.1);
		display: grid;
		place-items: center;
	}
	.m-t {
		margin-top: 12px;
		font-size: 19px;
		font-weight: 600;
	}
	.m-d {
		margin-top: 6px;
		font-size: 14px;
		line-height: 1.55;
		color: var(--mu);
	}
	.m-d b {
		color: var(--tx);
	}
	.m-c {
		margin-top: 14px;
		font-size: 13px;
	}
	.m-in {
		margin-top: 6px;
		width: 100%;
		padding: 12px 14px;
		border-radius: 14px;
		border: 1.5px solid var(--ln2);
		background: var(--sf2);
		font:
			500 15px 'Instrument Sans',
			sans-serif;
		color: var(--tx);
		outline: none;
	}
	.m-acc {
		margin-top: 16px;
		display: flex;
		gap: 8px;
	}
	.m-acc button {
		flex: 1;
		padding: 13px;
		border-radius: 999px;
		font:
			600 14px 'Instrument Sans',
			sans-serif;
	}
	.b-sec {
		border: 1px solid var(--ln2);
		background: transparent;
		color: var(--tx);
	}
	.b-roja {
		border: 0;
		background: #b83a2e;
		color: #fff;
	}
	.b-roja:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}
</style>
