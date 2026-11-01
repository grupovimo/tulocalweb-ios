<script lang="ts">
	import { page } from '$app/state';
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { renderSVG } from 'uqr';
	import { onMount } from 'svelte';
	import { compartirTienda } from '#lib/panel/compartir.ts';
	import { getUi } from '#lib/ui.svelte.ts';
	import { RESUMEN } from './datos.ts';
	import { getOnboarding } from './estado.svelte.ts';

	/**
	 * Último paso, ¡Listo!. En la demo, el del handoff: lo que quedó armado, lo que falta, el QR y el
	 * equipo. Con la API, sólo lo accionable: el link y el QR de verdad para compartir, y el siguiente
	 * paso más útil (cargar el primer producto, o seguir la importación si eligió traer su tienda).
	 * Los botones mandan las elecciones al servidor (form del padre) y siguen en el panel.
	 */
	let {
		terminar,
		terminando = false
	}: { terminar: (a?: 'panel' | 'producto') => void; terminando?: boolean } = $props();
	const o = getOnboarding();
	const ui = getUi();
	// En el celular, la hoja de compartir del sistema; en la compu, el link copiado.
	let tactil = $state(false);
	onMount(() => {
		tactil = matchMedia('(pointer: coarse)').matches && typeof navigator.share === 'function';
	});
	const importa = $derived(o.e.importar);
	const conSesion = $derived(!!page.data.conSesion);
	// Con la API, la tienda de verdad; en la demo, la del handoff.
	const real = $derived(page.data.tienda as { nombre: string; slug: string } | null);
	const link = $derived(
		real ? `https://${real.slug}.tulocalweb.com.ar` : 'https://laesquina.com.ar'
	);
	const qrSvg = $derived(
		'data:image/svg+xml;charset=utf-8,' +
			encodeURIComponent(
				renderSVG(link, { border: 2, ecc: 'M', whiteColor: '#FBF7F2', blackColor: '#2A140C' })
			)
	);
	let copiado = $state(false);
	let invitacion = $state('');

	function confeti(host: HTMLElement) {
		if (matchMedia('(prefers-reduced-motion: reduce)').matches) return;
		const cols = ['#D2582F', '#E8875F', '#E3A93B', '#2F8F63', '#2A140C'];
		const t = setTimeout(() => {
			for (let i = 0; i < 70; i++) {
				const p = document.createElement('span');
				const w = 6 + Math.random() * 6;
				p.style.cssText = `position:absolute;left:0;top:0;width:${w}px;height:${w * 0.45}px;border-radius:2px;background:${cols[i % 5]}`;
				host.appendChild(p);
				const a = Math.random() * Math.PI * 2;
				const d = 160 + Math.random() * 260;
				p.animate(
					[
						{ transform: 'translate(0,0) rotate(0)', opacity: 1 },
						{
							transform: `translate(${Math.cos(a) * d}px,${Math.sin(a) * d * 0.6 + 260}px) rotate(${Math.random() * 720}deg)`,
							opacity: 0
						}
					],
					{
						duration: 1600 + Math.random() * 1200,
						easing: 'cubic-bezier(.15,.6,.3,1)',
						fill: 'forwards'
					}
				).onfinish = () => p.remove();
			}
		}, 300);
		return { destroy: () => clearTimeout(t) };
	}
	async function copiar() {
		try {
			await navigator.clipboard.writeText(link);
		} catch {
			// Sin permiso de portapapeles: igual se muestra el aviso, el link está a la vista.
		}
		copiado = true;
		setTimeout(() => (copiado = false), 1800);
	}
	const iniciales = (e: string) => e.slice(0, 2).toUpperCase();
	const compartir = () => real && compartirTienda(ui, real);
</script>

<div class="todo">
	<div class="confeti" use:confeti></div>
	<div class="anillo">
		<svg viewBox="0 0 120 120" aria-hidden="true"
			><circle cx="60" cy="60" r="52" fill="none" stroke="#F1E7DC" stroke-width="10"
			></circle><circle
				cx="60"
				cy="60"
				r="52"
				fill="none"
				stroke="#D2582F"
				stroke-width="10"
				stroke-linecap="round"
				stroke-dasharray="326.7"
				stroke-dashoffset="0"
			></circle></svg
		>
		<div class="ave"><img src="/hornero/celebrando.webp" alt="Tincho celebrando" /></div>
	</div>
	{#if o.demo}
		<h1>Tu Local Web ya está listo.</h1>
		<p class="bajada">
			Y no tuviste que cargar nada a mano. Ya podés vender por todos tus canales.
		</p>
		<div class="resumen">
			{#each RESUMEN as [v, t] (t)}<div>
					<div class="r-v">{v}</div>
					<div class="r-t">{t}</div>
				</div>{/each}
		</div>
		<div class="faltan">
			<div class="f-t">Te quedan 2 cosas, cuando quieras</div>
			<div class="f-fila">
				<span class="arca">ARCA</span><span class="f-x"
					>Subir el certificado para facturar <span>· 2 minutos</span></span
				><button type="button" class="f-link" onclick={() => terminar()}>Hacerlo</button>
			</div>
			<div class="f-fila">
				<span class="corona"><Icon icon={I.phCrownSimpleFill} size={15} /></span><span class="f-x"
					>Elegir tu plan <span>· te quedan 14 días gratis</span></span
				>
				<!-- Con sesión, los planes del panel; sin sesión (la demo), los de la landing. -->
				<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- rutas fijas -->
				<a class="f-link" href={conSesion ? '/panel/plan-y-facturacion' : '/#precios'}>Ver planes</a
				>
			</div>
		</div>
		<div class="dos">
			<div class="online">
				<!-- El QR real del link de la tienda: tocándolo se descarga para imprimir. -->
				<!-- eslint-disable svelte/no-navigation-without-resolve -- descarga de un data: URL -->
				<a
					class="qr"
					href={qrSvg}
					download="qr-{real?.slug ?? 'tienda'}.svg"
					aria-label="Descargar el QR de la tienda"
					><img src={qrSvg} alt="QR de {link.replace('https://', '')}" /></a
				>
				<!-- eslint-enable svelte/no-navigation-without-resolve -->
				<div class="on-txt">
					<div class="on-t">Tu tienda ya está en línea</div>
					<div class="on-d">{link.replace('https://', '')} · imprimí el QR para el mostrador</div>
					<div class="on-acc">
						<button type="button" onclick={copiar}
							><Icon icon={copiado ? I.phCheckBold : I.phCopyBold} size={13} />{copiado
								? 'Copiado'
								: 'Copiar link'}</button
						>
						<a
							class="wa"
							href="https://wa.me/?text={encodeURIComponent(`Mirá mi tienda: ${link}`)}"
							target="_blank"
							rel="noopener"><Icon icon={I.phWhatsappLogoFill} size={13} />Compartir</a
						>
					</div>
				</div>
			</div>
			<div class="equipo">
				<div class="eq-t">Sumá a tu equipo</div>
				<div class="eq-d">Cada uno con sus permisos y doble verificación.</div>
				<div class="eq-l">
					{#each o.e.equipo as m (m.email)}
						<div>
							<span class="ini">{iniciales(m.email)}</span><span class="mail">{m.email}</span><span
								class="rol">{m.rol}</span
							>
						</div>
					{/each}
				</div>
				<form
					class="invitar"
					onsubmit={(ev) => {
						ev.preventDefault();
						o.invitar(invitacion);
						invitacion = '';
					}}
				>
					<input
						type="email"
						bind:value={invitacion}
						placeholder="email de tu socio o empleado"
						aria-label="Email de tu socio o empleado"
					/>
					<button type="submit">Invitar</button>
				</form>
			</div>
		</div>
		<div class="ctas">
			<button type="button" class="pri" onclick={() => terminar()}
				>Ir a mi panel <Icon icon={I.phArrowRightBold} size={16} /></button
			>
			<button type="button" class="sec" onclick={() => terminar()}>Ver mi tienda</button>
		</div>
	{:else}
		<!-- Con la API: el link de verdad para compartir y el siguiente paso, nada más. -->
		<h1>Tu tienda ya está en línea.</h1>
		<p class="bajada">
			{importa
				? 'Compartí el link cuando quieras y traé tus productos: empieza apenas sigas.'
				: 'Compartí el link con tus clientes y cargá tu primer producto.'}
		</p>
		<div class="online real-online">
			<!-- El QR real del link de la tienda: tocándolo se descarga para imprimir. -->
			<!-- eslint-disable svelte/no-navigation-without-resolve -- descarga de un data: URL -->
			<a
				class="qr"
				href={qrSvg}
				download="qr-{real?.slug ?? 'tienda'}.svg"
				aria-label="Descargar el QR de la tienda"
				><img src={qrSvg} alt="QR de {link.replace('https://', '')}" /></a
			>
			<!-- eslint-enable svelte/no-navigation-without-resolve -->
			<div class="on-txt">
				<div class="on-t">Tu link</div>
				<div class="on-d link">{link.replace('https://', '')}</div>
				<div class="on-acc">
					{#if tactil}
						<button type="button" class="compartir" onclick={compartir}
							><Icon icon={I.phPaperPlaneTiltFill} size={15} />Compartir</button
						>
					{:else}
						<button type="button" class="compartir" onclick={copiar}
							><Icon icon={copiado ? I.phCheckBold : I.phCopyBold} size={15} />{copiado
								? 'Copiado'
								: 'Copiar link'}</button
						>
						<a
							class="wa"
							href="https://wa.me/?text={encodeURIComponent(
								`¡Hola! Mirá ${real?.nombre ?? 'mi tienda'}, ya podés comprar online: ${link}`
							)}"
							target="_blank"
							rel="noopener"><Icon icon={I.phWhatsappLogoFill} size={15} />WhatsApp</a
						>
					{/if}
					<!-- eslint-disable svelte/no-navigation-without-resolve -- descarga de un data: URL -->
					<a href={qrSvg} download="qr-{real?.slug ?? 'tienda'}.svg"
						><Icon icon={I.phQrCodeFill} size={15} />Descargar QR</a
					>
					<!-- eslint-enable svelte/no-navigation-without-resolve -->
				</div>
			</div>
		</div>
		<div class="ctas">
			<button
				type="button"
				class="pri"
				disabled={terminando}
				onclick={() => terminar(importa ? 'panel' : 'producto')}
				>{terminando ? 'Abriendo…' : importa ? 'Traer mis productos' : 'Cargar mi primer producto'}
				<Icon icon={I.phArrowRightBold} size={16} /></button
			>
			<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- la tienda pública, en otra pestaña -->
			<a class="sec" href={link} target="_blank" rel="noopener">Ver mi tienda</a>
			{#if !importa}
				<button type="button" class="sec" disabled={terminando} onclick={() => terminar('panel')}
					>Ir a mi panel</button
				>
			{/if}
		</div>
	{/if}
</div>

<style>
	.todo {
		position: relative;
		text-align: center;
		max-width: 760px;
		margin: 0 auto;
	}
	.confeti {
		position: absolute;
		left: 50%;
		top: 60px;
		width: 0;
		height: 0;
	}
	.anillo {
		position: relative;
		width: 200px;
		height: 200px;
		margin: 0 auto;
	}
	.anillo svg {
		width: 100%;
		height: 100%;
		transform: rotate(-90deg);
	}
	.ave {
		position: absolute;
		inset: 22px;
		border-radius: 50%;
		overflow: hidden;
		background: #fcede5;
	}
	.ave img {
		display: block;
		width: 100%;
		height: 100%;
		object-fit: contain;
	}
	h1 {
		margin: 22px 0 0;
		font-size: clamp(34px, 4.2vw, 56px);
		line-height: 1;
		letter-spacing: -0.045em;
		font-weight: 600;
	}
	.bajada {
		margin: 14px auto 0;
		font-size: 17px;
		line-height: 1.55;
		color: #5c4536;
		max-width: 520px;
	}
	.resumen {
		margin-top: 28px;
		display: grid;
		grid-template-columns: repeat(4, minmax(0, 1fr));
		gap: 10px;
	}
	.resumen > div {
		padding: 16px 12px;
		border-radius: 18px;
		background: #fff;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.06);
	}
	.r-v {
		font-size: 28px;
		font-weight: 600;
		letter-spacing: -0.04em;
	}
	.r-t {
		font-size: 13px;
		color: #7a6252;
		margin-top: 2px;
	}
	.faltan {
		margin-top: 20px;
		padding: 16px 18px;
		border-radius: 20px;
		background: #fff;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.06);
		text-align: left;
	}
	.f-t {
		font-size: 14.5px;
		font-weight: 600;
	}
	.f-fila {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 10px 0;
		border-top: 1px solid rgba(58, 29, 18, 0.07);
	}
	.f-t + .f-fila {
		margin-top: 8px;
	}
	.arca,
	.corona {
		width: 30px;
		height: 30px;
		flex: none;
		border-radius: 9px;
		display: grid;
		place-items: center;
	}
	.arca {
		background: #1f3a6e;
		font-size: 7px;
		font-weight: 800;
		color: #fff;
	}
	.corona {
		background: #fcede5;
		color: #b0441d;
	}
	.f-x {
		flex: 1;
		font-size: 14px;
	}
	.f-x span {
		color: #7a6252;
	}
	.f-link {
		font-size: 13.5px;
		font-weight: 600;
		color: #b0441d;
		background: none;
		border: 0;
		padding: 0;
	}
	.dos {
		margin-top: 14px;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 300px), 1fr));
		gap: 12px;
		text-align: left;
	}
	.online {
		padding: 16px;
		border-radius: 20px;
		background: #2a140c;
		color: #fbf7f2;
		display: flex;
		gap: 14px;
		align-items: center;
	}
	.qr {
		width: 96px;
		height: 96px;
		flex: none;
		border-radius: 14px;
		background: #fbf7f2;
		padding: 6px;
		display: block;
	}
	.qr img {
		width: 100%;
		height: 100%;
		display: block;
	}
	.on-txt {
		flex: 1;
		min-width: 0;
	}
	.on-t {
		font-size: 15px;
		font-weight: 600;
	}
	.on-d {
		font-size: 12.5px;
		color: #d9c8ba;
		margin-top: 2px;
	}
	.on-acc {
		margin-top: 10px;
		display: flex;
		gap: 6px;
		flex-wrap: wrap;
	}
	.on-acc button,
	.on-acc a {
		display: flex;
		align-items: center;
		gap: 5px;
		padding: 7px 12px;
		border-radius: 999px;
		border: 0;
		background: rgba(251, 247, 242, 0.12);
		color: #fff;
		font-size: 12.5px;
		font-weight: 600;
	}
	.on-acc .wa {
		background: #25d366;
	}
	.real-online {
		margin-top: 22px;
		text-align: left;
	}
	.real-online .qr {
		width: 112px;
		height: 112px;
	}
	.on-d.link {
		font-size: 14px;
		color: #fbf7f2;
		font-weight: 600;
		overflow-wrap: anywhere;
	}
	.on-acc .compartir {
		background: #fbf7f2;
		color: #2a140c;
	}
	.ctas button:disabled {
		opacity: 0.7;
	}
	.ctas a {
		display: flex;
		align-items: center;
		gap: 10px;
		font-size: 16px;
		font-weight: 600;
		border-radius: 999px;
	}
	.equipo {
		padding: 16px;
		border-radius: 20px;
		background: #fff;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.06);
	}
	.eq-t {
		font-size: 15px;
		font-weight: 600;
	}
	.eq-d {
		font-size: 12.5px;
		color: #7a6252;
		margin-top: 2px;
	}
	.eq-l {
		margin-top: 10px;
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.eq-l > div {
		display: flex;
		align-items: center;
		gap: 10px;
		font-size: 13.5px;
	}
	.ini {
		width: 28px;
		height: 28px;
		flex: none;
		border-radius: 50%;
		background: #f1e7dc;
		display: grid;
		place-items: center;
		font-size: 11px;
		font-weight: 700;
		color: #7a6252;
	}
	.mail {
		flex: 1;
		min-width: 0;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.rol {
		font-size: 12px;
		font-weight: 600;
		padding: 3px 9px;
		border-radius: 999px;
		background: #f4ece2;
		color: #5c4536;
	}
	.invitar {
		margin-top: 10px;
		display: flex;
		gap: 6px;
	}
	.invitar input {
		flex: 1;
		min-width: 0;
		padding: 9px 12px;
		border-radius: 12px;
		border: 1px solid rgba(58, 29, 18, 0.16);
		font-size: 13.5px;
		font-weight: 500;
		outline: none;
		color: #2a140c;
		background: #fff;
	}
	.invitar button {
		padding: 9px 14px;
		border-radius: 12px;
		border: 0;
		background: #2a140c;
		color: #fff;
		font-size: 13px;
		font-weight: 600;
	}
	.ctas {
		margin-top: 24px;
		display: flex;
		gap: 10px;
		justify-content: center;
		flex-wrap: wrap;
	}
	.ctas button {
		display: flex;
		align-items: center;
		gap: 10px;
		font-size: 16px;
		font-weight: 600;
		border-radius: 999px;
	}
	.pri {
		color: #fff;
		border: 0;
		background: linear-gradient(180deg, #c24e25, #b0441d);
		padding: 16px 26px;
		box-shadow:
			inset 0 1px 0 rgba(255, 255, 255, 0.3),
			0 16px 34px -14px rgba(210, 88, 47, 0.75);
	}
	.sec {
		color: #2a140c;
		background: #fff;
		border: 1px solid rgba(58, 29, 18, 0.14);
		padding: 16px 24px;
	}
	@media (max-width: 560px) {
		.resumen {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
		/* En el celular: el QR arriba, los botones a lo ancho y de 44 px. */
		.real-online {
			flex-direction: column;
			align-items: stretch;
			text-align: center;
		}
		.real-online .qr {
			margin: 0 auto;
			width: 160px;
			height: 160px;
		}
		.real-online .on-acc {
			justify-content: center;
		}
		.on-acc button,
		.on-acc a {
			min-height: 44px;
			padding: 10px 16px;
			font-size: 14px;
		}
		.ctas {
			flex-direction: column;
			align-items: stretch;
		}
		.ctas button,
		.ctas a {
			justify-content: center;
		}
		.anillo {
			width: 150px;
			height: 150px;
		}
		.ave {
			inset: 16px;
		}
	}
</style>
