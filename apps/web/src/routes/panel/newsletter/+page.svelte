<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { page } from '$app/state';
	import Alerta from '#lib/components/Alerta.svelte';
	import Icon from '#lib/components/Icon.svelte';
	import Modal from '#lib/components/Modal.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import Vacio from '#lib/components/Vacio.svelte';
	import * as I from '#lib/icons.ts';
	import { unsplash } from '#lib/landing/formato.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { diaDeMes, pesos, plural } from '#lib/panel/formato.ts';
	import { iniciales } from '#lib/panel/listas/clientes.ts';
	import { MAX_PRODUCTOS, origenDe, parrafos } from '#lib/panel/newsletter.ts';
	import { VIDA_TOAST, getUi } from '#lib/ui.svelte.ts';

	/**
	 * Crecer › Newsletter (§50), con las piezas de Marketing (design/PanelCrecer.dc.html, sMkt): la
	 * campaña nueva con su vista previa al costado, las campañas enviadas y la lista de suscriptos
	 * con las filas de Carritos abandonados. En la demo «Mandar» no le escribe a nadie.
	 */
	let { data } = $props();
	const ui = getUi();
	const n = $derived(data.newsletter);
	const demo = $derived(!!page.data.demo);

	let asunto = $state('');
	let texto = $state('');
	let elegidos = $state<string[]>([]);
	let buscar = $state('');
	let confirmar = $state(false);
	let enviando = $state(false);
	let error = $state<string | null>(null);
	let cuantos = $state(30);
	/** Los que se borraron y todavía se pueden recuperar con Deshacer. */
	let ocultos = $state<string[]>([]);

	const porId = $derived(new Map(n.productos.map((p) => [p.id, p])));
	const destacados = $derived(elegidos.map((id) => porId.get(id)).filter((p) => !!p));
	const opciones = $derived(
		n.productos
			.filter((p) => !buscar.trim() || p.nombre.toLowerCase().includes(buscar.trim().toLowerCase()))
			.slice(0, 12)
	);
	const suscriptos = $derived(n.suscriptores.filter((s) => !ocultos.includes(s.id)));
	const listo = $derived(asunto.trim().length >= 3 && texto.trim().length >= 10);
	const vistaParrafos = $derived(parrafos(texto));
	const nombreTienda = $derived(page.data.tienda?.nombre ?? 'Tu tienda');

	function alternar(id: string) {
		if (elegidos.includes(id)) elegidos = elegidos.filter((x) => x !== id);
		else if (elegidos.length < MAX_PRODUCTOS) elegidos = [...elegidos, id];
	}

	async function mandar() {
		if (enviando) return;
		enviando = true;
		error = null;
		const r = await fetch('/panel/newsletter/envios', {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ asunto, texto, productos: elegidos })
		});
		enviando = false;
		if (!r.ok) {
			error =
				((await r.json().catch(() => null)) as { message?: string } | null)?.message ??
				'No se mandó la novedad. Probá de nuevo.';
			return;
		}
		confirmar = false;
		const e = (await r.json()) as { destinatarios: number };
		asunto = texto = buscar = '';
		elegidos = [];
		await invalidateAll();
		ui.toast({
			k: 'ok',
			title: demo ? 'Novedad guardada' : 'Novedad en camino',
			desc: demo
				? 'Es la demo: no salió ningún mail'
				: `${plural(e.destinatarios, 'mail', 'mails')} de a tandas · hasta ${n.tope} por día`
		});
	}

	function borrar(s: (typeof n.suscriptores)[number]) {
		ocultos = [...ocultos, s.id];
		let deshecho = false;
		ui.undoable(`Sacaste a ${s.email}`, 'Ya no le llegan tus novedades', () => {
			deshecho = true;
			ocultos = ocultos.filter((x) => x !== s.id);
		});
		// Se borra cuando ya no se puede deshacer.
		setTimeout(async () => {
			if (deshecho) return;
			const r = await fetch(`/panel/newsletter/suscriptores/${encodeURIComponent(s.id)}`, {
				method: 'DELETE'
			});
			if (!r.ok && r.status !== 404) {
				ocultos = ocultos.filter((x) => x !== s.id);
				ui.toast({ k: 'err', title: 'No pudimos sacarlo de la lista', desc: 'Probá de nuevo.' });
				return;
			}
			await invalidateAll();
			ocultos = ocultos.filter((x) => x !== s.id);
		}, VIDA_TOAST);
	}

	const ESTADO: Record<string, string> = {
		enviando: 'Enviando',
		manana: 'Sigue mañana',
		enviado: 'Enviada'
	};
	const foto = (f: string | null) => (f ? unsplash(f, 240, '&h=240') : null);
</script>

<Seo
	title="Newsletter · Panel"
	description="Suscriptos y novedades por email"
	path="/panel/newsletter"
	noindex
/>

<div class="pagina">
	<div class="cab" use:anim={{ t: 'rise' }}>
		<div class="c-txt">
			<h1>Newsletter</h1>
			<div class="c-s">
				Quienes se sumaron desde tu tienda online y las novedades que les&nbsp;mandás
			</div>
		</div>
		{#if n.suscriptores.length}
			<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- descarga del panel -->
			<a class="btn btn-secondary btn-sm" href="/panel/newsletter/exportar" download
				><Icon icon={I.phDownloadSimpleBold} size={15} />Exportar CSV</a
			>
		{/if}
	</div>

	{#if demo}
		<div use:anim={{ t: 'rise', delay: 30 }}>
			<Alerta
				tipo="info"
				titulo="En la demo no sale ningún mail"
				texto="Probá todo: «Mandar» guarda la novedad en el historial, pero no le escribe a&nbsp;nadie."
			/>
		</div>
	{/if}

	<div class="kpis" use:anim={{ t: 'rise', delay: 60 }}>
		<div class="oscura">
			<div class="o-t">Suscriptos confirmados</div>
			<div class="o-v">{n.confirmados.toLocaleString('es-AR')}</div>
			<div class="o-s">
				{n.pendientes
					? `${plural(n.pendientes, 'persona', 'personas')} sin confirmar el mail`
					: 'Todos confirmaron su mail'}
			</div>
		</div>
		<div class="tarjeta">
			<div class="t-t">Mails de novedades hoy</div>
			<div class="t-v">{n.hoy.toLocaleString('es-AR')} <span>de {n.tope}</span></div>
			<div class="barra"><div style:width="{Math.min(100, (n.hoy / n.tope) * 100)}%"></div></div>
			<div class="t-s">
				Lo que no entra hoy sale mañana, así cuidamos que no lleguen como&nbsp;spam
			</div>
		</div>
	</div>

	<div class="arriba" use:anim={{ t: 'rise', delay: 90 }}>
		<div class="tarjeta nueva">
			<div class="n-cab">
				<span class="n-ico"><Icon icon={I.phEnvelopeSimpleFill} size={19} /></span><span
					>Mandar novedad</span
				>
			</div>
			<label class="field"
				>Asunto<input
					class="input"
					bind:value={asunto}
					maxlength="120"
					placeholder="Llegó la colección de otoño"
				/></label
			>
			<label class="field"
				>Texto<textarea
					bind:value={texto}
					maxlength="5000"
					placeholder="Contá la novedad. Dejá una línea en blanco para separar párrafos."
				></textarea></label
			>
			<div>
				<div class="et fila">
					Productos destacados<span>{elegidos.length} de {MAX_PRODUCTOS}</span>
				</div>
				{#if n.productos.length > 12}
					<input
						class="input buscar"
						bind:value={buscar}
						placeholder="Buscar un producto"
						aria-label="Buscar un producto"
					/>
				{/if}
				<div class="pills" role="group" aria-label="Productos destacados">
					{#each destacados as p (p.id)}
						<button
							type="button"
							class="pill chica"
							aria-pressed="true"
							onclick={() => alternar(p.id)}
							>{#if foto(p.foto)}<img src={foto(p.foto)} alt="" />{/if}{p.nombre}<Icon
								icon={I.phX}
								size={12}
							/></button
						>
					{/each}
					{#each opciones.filter((p) => !elegidos.includes(p.id)) as p (p.id)}
						<button
							type="button"
							class="pill chica"
							aria-pressed="false"
							disabled={elegidos.length >= MAX_PRODUCTOS}
							onclick={() => alternar(p.id)}
							>{#if foto(p.foto)}<img src={foto(p.foto)} alt="" />{/if}{p.nombre}</button
						>
					{:else}
						{#if !n.productos.length}<span class="nada">No tenés productos publicados.</span>{/if}
					{/each}
				</div>
			</div>
			<div class="acc">
				<button
					type="button"
					class="enviar"
					disabled={!listo || !n.confirmados}
					onclick={() => {
						error = null;
						confirmar = true;
					}}
					>{n.confirmados
						? `Mandar a ${plural(n.confirmados, 'persona', 'personas')}`
						: 'Sin suscriptos confirmados todavía'}</button
				>
			</div>
		</div>

		<div class="tarjeta previa">
			<div class="p-cab">
				<span class="p-t">Vista previa del mail</span>
				<span class="p-de">De: {nombreTienda}</span>
			</div>
			<div class="mail" aria-label="Vista previa del mail">
				<div class="m-tienda">{nombreTienda}</div>
				<div class="m-asunto" class:vacio={!asunto.trim()}>{asunto.trim() || 'El asunto'}</div>
				{#each vistaParrafos.length ? vistaParrafos : ['Acá va el texto de la novedad.'] as p, i (i)}
					<p class:vacio={!vistaParrafos.length}>{p}</p>
				{/each}
				{#if destacados.length}
					<div class="m-productos">
						{#each destacados as p (p.id)}
							<div class="m-prod">
								{#if foto(p.foto)}<img src={foto(p.foto)} alt="" />{/if}
								<div class="m-pn">{p.nombre}</div>
								<div class="m-pp">{pesos(p.precio)}</div>
							</div>
						{/each}
					</div>
				{/if}
				<span class="m-boton">Ir a la tienda</span>
				<div class="m-pie">
					Te llega porque te sumaste a las novedades de {nombreTienda}. <u>Darme de baja</u>.
				</div>
			</div>
		</div>
	</div>

	{#if n.envios.length}
		<div class="campanas" use:anim={{ t: 'rise', delay: 120 }}>
			{#each n.envios as e (e.id)}
				<div class="tarjeta campana">
					<div class="ca-cab">
						<Icon icon={I.phEnvelopeSimpleFill} size={18} />
						<span class="ca-t">{e.asunto}</span>
						<span class="estado" class:activa={e.estado !== 'enviado'}>{ESTADO[e.estado]}</span>
					</div>
					<div class="ca-d">Email · {diaDeMes(new Date(e.creado))}</div>
					<div class="metricas">
						<div>
							<div class="m-k">Destinatarios</div>
							<div class="m-v">{e.destinatarios.toLocaleString('es-AR')}</div>
						</div>
						<div>
							<div class="m-k">Enviados</div>
							<div class="m-v">{e.enviados.toLocaleString('es-AR')}</div>
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}

	<div class="lista" use:anim={{ t: 'rise', delay: 150 }}>
		<div class="l-cab">
			Suscriptos <span>{plural(suscriptos.length, 'persona', 'personas')}</span>
		</div>
		{#each suscriptos.slice(0, cuantos) as s (s.id)}
			<div class="fila">
				<span class="av">{iniciales(s.email.split('@')[0].replace(/[._-]+/g, ' '))}</span>
				<span class="quien"
					><span class="n">{s.email}</span><span class="s"
						>Desde el {diaDeMes(new Date(s.creado))} · {origenDe(s.origen)}</span
					></span
				>
				<span class="estado" class:activa={s.confirmado}
					>{s.confirmado ? 'Confirmado' : 'Sin confirmar'}</span
				>
				<button
					type="button"
					class="icon-btn"
					aria-label="Sacar a {s.email} de la lista"
					onclick={() => borrar(s)}><Icon icon={I.phTrash} size={17} /></button
				>
			</div>
		{:else}
			<Vacio
				hornero="saludo"
				titulo="Todavía no se sumó nadie"
				texto="Tu tienda online ya muestra dónde dejar el email. Cuando alguien se suma y confirma con el link del mail, aparece&nbsp;acá."
			/>
		{/each}
		{#if suscriptos.length > cuantos}
			<button type="button" class="btn btn-secondary btn-sm mas" onclick={() => (cuantos += 60)}
				>Ver más</button
			>
		{/if}
	</div>
</div>

<Modal bind:open={confirmar} titulo="Mandar la novedad" ancho={440}>
	<div class="m-ico"><Icon icon={I.phPaperPlaneTiltFill} size={24} /></div>
	<div class="m-t">¿Mandar «{asunto.trim()}»?</div>
	<div class="m-d">
		{#if demo}
			Le llegaría a {plural(n.confirmados, 'persona', 'personas')}. Es la demo: no sale
			ningún&nbsp;mail.
		{:else}
			Le llega a {plural(n.confirmados, 'persona', 'personas')} que confirmaron su mail. Sale de a tandas,
			hasta {n.tope} por día; lo que no entra hoy sale&nbsp;mañana.
		{/if}
	</div>
	{#if error}<div class="m-error" role="alert">{error}</div>{/if}
	<div class="m-acc">
		<button type="button" class="btn btn-secondary" onclick={() => (confirmar = false)}
			>Cancelar</button
		>
		<button type="button" class="btn btn-primary" disabled={enviando} onclick={mandar}
			>{enviando ? 'Mandando…' : `Mandar a ${plural(n.confirmados, 'persona', 'personas')}`}</button
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
		gap: 14px;
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
		text-wrap: pretty;
	}
	.cab .btn {
		display: inline-flex;
		align-items: center;
		gap: 6px;
	}
	.kpis {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
		gap: 12px;
	}
	.oscura {
		padding: 20px;
		border-radius: 22px;
		background: #2a140c;
		color: #fbf7f2;
		box-shadow: var(--shd);
	}
	.o-t {
		font-size: 13px;
		color: #d9c8ba;
	}
	.o-v,
	.t-v {
		font-size: 34px;
		font-weight: 600;
		letter-spacing: -0.04em;
	}
	.t-v span {
		font-size: 15px;
		font-weight: 500;
		letter-spacing: 0;
		color: var(--mu);
	}
	.o-s {
		font-size: 13px;
		color: #f5b796;
		margin-top: 2px;
	}
	.tarjeta {
		padding: 20px;
		border-radius: 22px;
		background: var(--sf);
		box-shadow: var(--shd);
	}
	.t-t {
		font-size: 13px;
		color: var(--mu);
	}
	.barra {
		margin-top: 8px;
		height: 8px;
		border-radius: 8px;
		background: var(--sf3);
		overflow: hidden;
	}
	.barra div {
		height: 100%;
		border-radius: 8px;
		background: #2f8f63;
	}
	.t-s {
		font-size: 12.5px;
		color: var(--mu);
		margin-top: 6px;
		text-wrap: pretty;
	}
	.arriba {
		display: grid;
		grid-template-columns: minmax(0, 1fr) minmax(0, 360px);
		gap: 16px;
		align-items: start;
	}
	.nueva {
		border-radius: 24px;
		display: flex;
		flex-direction: column;
		gap: 14px;
	}
	.n-cab {
		display: flex;
		align-items: center;
		gap: 10px;
		font-size: 17px;
		font-weight: 600;
	}
	.n-ico {
		width: 38px;
		height: 38px;
		border-radius: 12px;
		background: rgba(210, 88, 47, 0.12);
		color: #b0441d;
		display: grid;
		place-items: center;
	}
	textarea {
		width: 100%;
		min-height: 120px;
		padding: 12px;
		border-radius: 14px;
		border: 1px solid var(--ln2);
		background: var(--bg);
		font:
			500 14px/1.5 'Instrument Sans',
			sans-serif;
		color: var(--tx);
		outline: none;
		resize: vertical;
		box-sizing: border-box;
	}
	.et {
		font-size: 12.5px;
		font-weight: 600;
		color: var(--mu);
		margin-bottom: 6px;
	}
	.et.fila {
		display: flex;
		justify-content: space-between;
	}
	.buscar {
		margin-bottom: 8px;
	}
	.pills {
		display: flex;
		gap: 6px;
		flex-wrap: wrap;
	}
	.pill {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 7px 11px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: transparent;
		color: var(--tx);
		font:
			600 12.5px 'Instrument Sans',
			sans-serif;
		max-width: 100%;
		min-height: 36px;
	}
	.pill img {
		width: 22px;
		height: 22px;
		border-radius: 6px;
		object-fit: cover;
		flex: none;
	}
	.pill[aria-pressed='true'] {
		border-color: #d2582f;
		background: rgba(210, 88, 47, 0.08);
	}
	.pill:disabled {
		opacity: 0.5;
	}
	.nada {
		font-size: 13px;
		color: var(--mu);
	}
	.acc {
		display: flex;
	}
	.enviar {
		flex: 1;
		min-height: 46px;
		padding: 12px;
		border-radius: 999px;
		border: 0;
		background: linear-gradient(180deg, #c24e25, #b0441d);
		color: #fff;
		font:
			600 14px 'Instrument Sans',
			sans-serif;
	}
	.enviar:disabled {
		opacity: 0.55;
	}
	.previa {
		padding: 14px;
	}
	.p-cab {
		display: flex;
		justify-content: space-between;
		gap: 10px;
		flex-wrap: wrap;
		padding: 2px 4px 10px;
		font-size: 12.5px;
		color: var(--mu);
	}
	.p-t {
		font-weight: 600;
		color: var(--tx);
	}
	/* El mail tal como sale (los colores de los mails de la tienda, en claro también en modo oscuro). */
	.mail {
		padding: 22px 18px;
		border-radius: 16px;
		background: #fbf7f2;
		color: #2a140c;
		font-family: Helvetica, Arial, sans-serif;
	}
	.m-tienda {
		font-size: 16px;
		font-weight: 700;
	}
	.m-asunto {
		margin-top: 16px;
		font-size: 21px;
		line-height: 1.15;
		font-weight: 600;
		overflow-wrap: anywhere;
	}
	.mail p {
		margin: 10px 0 0;
		font-size: 14px;
		line-height: 1.5;
		color: #5c4536;
		white-space: pre-line;
		overflow-wrap: anywhere;
	}
	.vacio {
		opacity: 0.45;
	}
	.m-productos {
		margin-top: 14px;
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 8px;
	}
	.m-prod {
		padding: 8px;
		border-radius: 12px;
		background: #fff;
		min-width: 0;
	}
	.m-prod img {
		width: 100%;
		aspect-ratio: 1;
		object-fit: cover;
		border-radius: 8px;
		display: block;
	}
	.m-pn {
		margin-top: 6px;
		font-size: 12.5px;
		font-weight: 600;
		line-height: 1.3;
	}
	.m-pp {
		font-size: 12.5px;
		color: #5c4536;
	}
	.m-boton {
		display: inline-block;
		margin-top: 16px;
		padding: 10px 16px;
		border-radius: 999px;
		background: #2a140c;
		color: #fff;
		font-size: 13px;
		font-weight: 600;
	}
	.m-pie {
		margin-top: 18px;
		font-size: 11px;
		line-height: 1.5;
		color: #9c8676;
	}
	.campanas {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
		gap: 12px;
	}
	.campana {
		padding: 16px;
	}
	.ca-cab {
		display: flex;
		align-items: center;
		gap: 8px;
	}
	.ca-t {
		flex: 1;
		min-width: 0;
		font-size: 15px;
		font-weight: 600;
		overflow-wrap: anywhere;
	}
	.estado {
		font-size: 11.5px;
		font-weight: 700;
		padding: 3px 9px;
		border-radius: 999px;
		background: var(--chip);
		color: var(--mu);
		white-space: nowrap;
	}
	.estado.activa {
		background: rgba(47, 143, 99, 0.13);
		color: #237550;
	}
	.ca-d {
		font-size: 12.5px;
		color: var(--mu);
		margin-top: 3px;
	}
	.metricas {
		margin-top: 12px;
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 6px;
	}
	.m-k {
		font-size: 11.5px;
		color: var(--mu);
	}
	.m-v {
		font-size: 16px;
		font-weight: 600;
	}
	.lista {
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.l-cab {
		font-size: 16px;
		font-weight: 600;
	}
	.l-cab span {
		font-size: 13px;
		font-weight: 500;
		color: var(--mu);
		margin-left: 6px;
	}
	.fila {
		display: flex;
		align-items: center;
		gap: 14px;
		padding: 12px 14px 12px 16px;
		border-radius: 20px;
		background: var(--sf);
		box-shadow: var(--shd);
	}
	.av {
		width: 44px;
		height: 44px;
		border-radius: 50%;
		flex: none;
		display: grid;
		place-items: center;
		background: var(--sf3);
		color: var(--mu);
		font-size: 13px;
		font-weight: 700;
	}
	.quien {
		flex: 1;
		min-width: 0;
	}
	.n {
		display: block;
		font-size: 15px;
		font-weight: 600;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.s {
		display: block;
		font-size: 12.5px;
		color: var(--mu);
	}
	.mas {
		align-self: center;
	}
	.m-ico {
		width: 48px;
		height: 48px;
		display: grid;
		place-items: center;
		border-radius: 15px;
		background: rgba(210, 88, 47, 0.12);
		color: #b0441d;
	}
	.m-t {
		margin-top: 16px;
		font-size: 21px;
		font-weight: 600;
		letter-spacing: -0.02em;
		overflow-wrap: anywhere;
		text-wrap: balance;
	}
	.m-d {
		margin-top: 6px;
		font-size: 14.5px;
		line-height: 1.5;
		color: var(--mu);
		text-wrap: pretty;
	}
	.m-error {
		margin-top: 10px;
		font-size: 13px;
		font-weight: 600;
		color: #b83a2e;
	}
	.m-acc {
		margin-top: 20px;
		display: flex;
		gap: 8px;
		justify-content: flex-end;
		flex-wrap: wrap;
	}
	@media (max-width: 999px) {
		.arriba {
			grid-template-columns: minmax(0, 1fr);
		}
	}
	@media (max-width: 520px) {
		.fila {
			gap: 10px;
			padding-left: 12px;
		}
		.av {
			display: none;
		}
	}
</style>
