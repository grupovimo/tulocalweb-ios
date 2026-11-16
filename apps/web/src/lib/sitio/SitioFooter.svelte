<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import { phArrowCounterClockwiseBold, phArrowRightBold, phQrCodeBold } from '#lib/icons.ts';
	import Marca from '#lib/landing/Marca.svelte';

	/**
	 * El footer de todo el sitio (la landing y las páginas públicas). Arriba, Tincho invita a crear la
	 * tienda (`cta`: la landing ya cierra con su cartel, así que lo apaga); en el medio, la marca y las
	 * columnas; abajo, lo legal, con el botón de arrepentimiento a la vista (es obligatorio).
	 * "Configurar cookies" reabre el banner.
	 */
	let { cta = true }: { cta?: boolean } = $props();

	const COLUMNAS = [
		[
			'Producto',
			[
				['/#producto', 'Tienda online'],
				['/#hornero', 'Tincho, el asistente'],
				['/#plantillas', 'Plantillas'],
				['/#precios', 'Precios'],
				['/novedades', 'Novedades']
			]
		],
		[
			'Migrar',
			[
				['/migrar?desde=tiendanube', 'Desde Tiendanube'],
				['/migrar?desde=empretienda', 'Desde Empretienda'],
				['/migrar?desde=mercadolibre', 'Desde Mercado Libre'],
				['/migrar', 'Desde otra plataforma']
			]
		],
		[
			'Empresa',
			[
				['/nosotros', 'Sobre nosotros'],
				['/partners', 'Partners y agencias'],
				['/seguridad', 'Seguridad'],
				['/estado', 'Estado del sistema'],
				['/contacto', 'Contacto y ventas']
			]
		],
		[
			'Ayuda',
			[
				['/ayuda', 'Centro de ayuda'],
				['/legales/terminos', 'Términos y condiciones'],
				['/legales/privacidad', 'Privacidad'],
				['/legales/tratamiento-de-datos', 'Datos de tus clientes'],
				['/legales/reembolsos', 'Reembolsos']
			]
		]
	] as const;

	const configurar = () => window.dispatchEvent(new CustomEvent('mostrador:configurar-cookies'));
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- páginas públicas y anclas de la landing -->
<footer>
	<div class="curvas" aria-hidden="true"></div>
	<div class="in">
		{#if cta}
			<div class="invita">
				<img src="/hornero/saludo.webp" alt="" width="120" height="120" />
				<div class="inv-txt">
					<b>¿Arrancamos con tu tienda?</b>
					<span>Te la dejamos andando hoy. 14 días gratis, sin tarjeta.</span>
				</div>
				<div class="inv-acc">
					<a class="btn-inflado btn-grande" href="/onboarding"
						>Crear mi tienda gratis<Icon icon={phArrowRightBold} size={16} /></a
					>
					<a class="vidrio" href="/contacto">Hablar con alguien</a>
				</div>
			</div>
		{/if}

		<div class="medio">
			<div class="marca">
				<a class="logo" href="/" aria-label="Tu Local Web, inicio">
					<Marca color="#FBF7F2" alto={34} /><img
						src="/wordmark-cream.webp"
						alt="Tu Local Web"
						width="210"
						height="24"
					/>
				</a>
				<p>Tu negocio en internet, sin complicaciones. Hecho en Argentina.</p>
				<a class="mail" href="mailto:hola@tulocalweb.com.ar">hola@tulocalweb.com.ar</a>
			</div>
			<nav class="cols" aria-label="Pie de página">
				{#each COLUMNAS as [titulo, links] (titulo)}
					<div class="col">
						<span class="col-t">{titulo}</span>
						{#each links as [href, t] (t)}<a
								{href}
								rel={href.includes('?') ? 'nofollow' : undefined}>{t}</a
							>{/each}
					</div>
				{/each}
			</nav>
		</div>

		<div class="grande" aria-hidden="true">Tu Local Web</div>

		<div class="barra">
			<span
				>© 2026 Tu Local Web · Nahuel Santillán · <span class="nw">CUIT 20-44859234-1</span></span
			>
			<span class="legal">
				<a class="arrep" href="/legales/arrepentimiento"
					><Icon icon={phArrowCounterClockwiseBold} size={13} />Botón de arrepentimiento</a
				>
				<a href="/legales/defensa-del-consumidor">Defensa del consumidor</a>
				<a href="/legales/cookies">Cookies</a>
				<button type="button" onclick={configurar}>Configurar cookies</button>
				<span class="fiscal"><Icon icon={phQrCodeBold} size={15} />Data fiscal</span>
			</span>
		</div>
	</div>
</footer>

<!-- eslint-enable svelte/no-navigation-without-resolve -->

<style>
	footer {
		position: relative;
		overflow: hidden;
		margin-top: 24px;
		background: #2a140c;
		color: #fbf7f2;
		border-radius: 36px 36px 0 0;
	}
	.curvas {
		position: absolute;
		inset: 0;
		background: url('/textura/curvas-oscuras.svg') center / cover no-repeat;
	}
	.in {
		position: relative;
		max-width: 1240px;
		margin: 0 auto;
		padding: 0 var(--lado, 28px);
	}

	/* Tincho invita */
	.invita {
		display: flex;
		align-items: center;
		gap: 20px 28px;
		flex-wrap: wrap;
		padding: 40px 0 36px;
		border-bottom: 1px solid rgba(251, 247, 242, 0.12);
	}
	.invita img {
		width: 96px;
		height: 96px;
		flex: none;
		object-fit: contain;
	}
	.inv-txt {
		flex: 1 1 280px;
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.inv-txt b {
		font-size: clamp(24px, 3vw, 34px);
		line-height: 1.05;
		letter-spacing: -0.03em;
		font-weight: 600;
	}
	.inv-txt span {
		font-size: 15px;
		color: #d9c8ba;
	}
	.inv-acc {
		display: flex;
		align-items: center;
		gap: 12px;
		flex-wrap: wrap;
	}
	.vidrio {
		display: inline-flex;
		align-items: center;
		min-height: 56px;
		padding: 0 24px;
		border-radius: 999px;
		color: #fff;
		font-size: 15.5px;
		font-weight: 600;
		white-space: nowrap;
		background: linear-gradient(180deg, rgba(255, 255, 255, 0.18), rgba(255, 255, 255, 0.06));
		box-shadow: var(--inflado-vidrio);
	}
	.vidrio:hover {
		color: #fff;
	}

	/* Marca y columnas */
	.medio {
		display: flex;
		gap: 40px 64px;
		flex-wrap: wrap;
		padding: 48px 0 24px;
	}
	.marca {
		flex: 1 1 260px;
		max-width: 340px;
	}
	.logo {
		display: flex;
		align-items: center;
		gap: 10px;
	}
	.logo img {
		height: 24px;
		width: auto;
		display: block;
	}
	.marca p {
		margin: 16px 0 0;
		font-size: 15px;
		line-height: 1.5;
		color: #d9c8ba;
	}
	.mail {
		display: inline-block;
		margin-top: 12px;
		font-size: 15px;
		font-weight: 600;
		color: #f5b796;
	}
	.cols {
		flex: 3 1 520px;
		display: grid;
		grid-template-columns: repeat(4, minmax(0, 1fr));
		gap: 28px 24px;
	}
	.col {
		display: flex;
		flex-direction: column;
		gap: 10px;
		min-width: 0;
	}
	.col-t {
		font-size: 13px;
		font-weight: 700;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		color: #b8a293;
	}
	.col a {
		font-size: 14.5px;
		color: #fbf7f2;
	}
	.col a:hover,
	.mail:hover {
		color: #f5b796;
	}

	/* La marca grande, de fondo */
	.grande {
		font-size: clamp(64px, 15.5vw, 210px);
		font-weight: 700;
		letter-spacing: -0.06em;
		line-height: 0.8;
		white-space: nowrap;
		color: rgba(251, 247, 242, 0.07);
		margin: 8px 0 -0.08em;
		user-select: none;
	}

	/* Lo legal */
	.barra {
		display: flex;
		gap: 12px 24px;
		flex-wrap: wrap;
		justify-content: space-between;
		align-items: center;
		padding: 18px 0 24px;
		border-top: 1px solid rgba(251, 247, 242, 0.12);
		font-size: 13px;
		color: #b8a293;
	}
	.nw {
		white-space: nowrap;
	}
	.legal {
		display: flex;
		gap: 8px 18px;
		flex-wrap: wrap;
		align-items: center;
	}
	.legal a,
	.legal button {
		color: #d9c8ba;
		font: inherit;
		background: none;
		border: 0;
		padding: 0;
		cursor: pointer;
	}
	.legal a:hover,
	.legal button:hover {
		color: #fff;
	}
	.arrep {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		padding: 6px 12px !important;
		border-radius: 999px;
		background: rgba(251, 247, 242, 0.1) !important;
		color: #fff !important;
		font-weight: 600;
	}
	.fiscal {
		display: inline-flex;
		align-items: center;
		gap: 6px;
	}

	@media (max-width: 760px) {
		footer {
			border-radius: 28px 28px 0 0;
		}
		.invita {
			padding-top: 32px;
		}
		.invita img {
			width: 72px;
			height: 72px;
		}
		.inv-acc,
		.inv-acc a {
			width: 100%;
			justify-content: center;
		}
		.cols {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
	}
</style>
