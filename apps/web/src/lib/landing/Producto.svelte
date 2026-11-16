<script lang="ts">
	import CtaTienda from './CtaTienda.svelte';
	import Icon from '#lib/components/Icon.svelte';
	import { anim } from '#lib/anim.ts';
	import * as I from '#lib/icons.ts';
	import { INTEG } from './datos.ts';
	import { cara, unsplash } from './formato.ts';
	import { spot } from './spot.ts';

	/**
	 * "Todo lo que tienen las otras. Y lo que les falta." — el bento de producto del handoff, y abajo,
	 * en dos renglones, con qué se conecta y desde qué IA se maneja (antes eran secciones aparte).
	 */
	const LOGOS = (INTEG as unknown as [string, string, { icon?: unknown; col?: string }][])
		.filter(([, cat, m]) => cat !== 'ia' && 'icon' in m)
		.map(([n, , m]) => ({ n, icon: m.icon as I.IconData, col: m.col }));
	/** Las que no están en los sets de íconos: el logo oficial (static/marcas), con su proporción. */
	const MARCAS = [
		['Mercado Libre', '/marcas/mercadolibre.svg', 2.4],
		['Tiendanube', '/marcas/tiendanube.svg', 5.94],
		['Payway', '/marcas/payway.svg', 3.51],
		['Correo Argentino', '/marcas/correo-argentino.svg', 3.4],
		['ARCA', '/marcas/arca.webp', 9.83],
		['Empretienda', '/marcas/empretienda.svg', 7.07]
	] as const;
	const IAS = [
		['Claude', I.siClaude, '#D97757'],
		['ChatGPT', I.siOpenai, '#10A37F'],
		['Cualquier IA con MCP', I.phPlugsConnectedBold, '#8A5A3E']
	] as const;
	const CHATS = [
		[
			'1531058240690-006c446962d8',
			'Lucía F.',
			'¿Tenés el termo en negro?',
			I.logosWhatsappIcon,
			true
		],
		['1687422811062-a966b55cb217', 'Martín G.', '¡Llegó perfecto!', I.logosTelegram, false],
		['1764173039506-c5cc7b8ec14b', 'Norma R.', 'Pedido mayorista', I.logosInstagramIcon, false]
	] as const;
</script>

<section id="producto">
	<div class="in">
		<div class="cab" use:anim={'reveal'}>
			<div class="etiq"><span></span>Un solo panel</div>
			<h2>Todo lo que tienen las otras. Y lo que les falta.</h2>
		</div>
		<div class="bento">
			<div class="caja wa" use:anim={'reveal'} use:spot>
				<div class="sello">
					<Icon icon={I.logosWhatsappIcon} size={22} /><span class="s-n">WhatsApp y Telegram</span
					><span class="incl">Incluido</span>
				</div>
				<h3>Vendé donde ya te escriben</h3>
				<p>
					Bandeja compartida para todo el equipo, catálogo dentro del chat, link de pago y campañas
					con el costo exacto antes de enviar.
				</p>
				<div class="bandeja">
					<div class="lista">
						{#each CHATS as [id, n, m, ico, act] (n)}
							<div class="chat-f" class:act>
								<img src={cara(id)} alt={n} />
								<div class="cf-txt">
									<div class="cf-n">{n}</div>
									<div class="cf-m">{m}</div>
								</div>
								<Icon icon={ico} size={13} />
							</div>
						{/each}
					</div>
					<div class="hilo">
						<div class="b-c">Hola! ¿Tenés el termo de 1 L en negro?</div>
						<div class="b-v">¡Sí! Te paso el link de pago 👇</div>
						<div class="b-link">
							<div class="bl-prod">
								<img
									src={unsplash('1573612664822-d7d347da7b80', 100, '&h=100')}
									alt="Interior de un local de ropa y regalos"
								/>
								<div class="bl-txt">
									<div class="a">Termo 1 L · Negro</div>
									<div class="b">$ 42.900</div>
								</div>
							</div>
							<div class="bl-pago">
								<Icon icon={I.siMercadopago} size={16} style="color:#00B1EA" />Pagar con Mercado
								Pago
							</div>
						</div>
					</div>
				</div>
			</div>

			<div class="caja pos" use:anim={{ t: 'reveal', delay: 100 }}>
				<img
					data-pxi
					src={unsplash('1556742049-0cfed4f6a45d', 900)}
					alt="Cobro en mostrador"
					loading="lazy"
				/>
				<div class="pos-sombra"></div>
				<div class="pos-in">
					<div class="sello">
						<Icon icon={I.phStorefrontFill} size={22} style="color:#F5B796" /><span
							class="s-n blanco">Punto de venta</span
						><span class="extra">Sin costo extra</span>
					</div>
					<h3 class="pos-h">El local y la web, con un solo stock</h3>
					<p class="pos-p">Cobrás con QR, tarjeta o efectivo y la factura sale sola.</p>
					<div class="medios">
						<span><Icon icon={I.siMercadopago} size={16} style="color:#00B1EA" />QR</span>
						<span><Icon icon={I.logosVisa} size={30} /></span>
						<span><Icon icon={I.logosMastercard} size={22} /></span>
					</div>
				</div>
			</div>

			<div id="asistente" class="caja ia" use:anim={'reveal'}>
				<div class="ia-txt">
					<div class="sello">
						<Icon icon={I.phSparkleFill} size={20} /><span class="s-n blanco">Asistente con IA</span
						><span class="excl">Exclusivo</span>
					</div>
					<h3>Tincho te avisa qué hacer</h3>
					<p class="ia-p">
						Responde consultas, recupera carritos y arma promos con tus números reales. Vos decidís
						qué manda solo y qué te consulta.
					</p>
					<div class="preguntas">
						<span>¿Qué vendí más ayer?</span><span>Armá una promo de fin de mes</span>
					</div>
				</div>
			</div>

			<div class="caja fact" use:anim={{ t: 'reveal', delay: 100 }} use:spot>
				<div class="sello">
					<Icon icon={I.phReceiptFill} size={22} style="color:#b0441d" /><span class="s-n"
						>Facturación y cobros</span
					>
				</div>
				<h3 class="fact-h">Factura electrónica ARCA en cada venta. Cobros conciliados solos.</h3>
				<div class="filas">
					<div class="fila">
						<span class="f-n"
							><Icon icon={I.siMercadopago} size={18} style="color:#00B1EA" />Mercado Pago</span
						><span class="ok">Conciliado</span>
					</div>
					<div class="fila">
						<span class="f-n"
							><img src="/marcas/payway.svg" alt="Payway" height="18" width="63" /></span
						><span class="ok">Conciliado</span>
					</div>
					<div class="fila arca">
						<span class="f-n"
							><img src="/marcas/arca.webp" alt="ARCA" height="14" width="138" />Factura&nbsp;B
							<span style="white-space:nowrap">#0004-00812</span></span
						><span class="emitida">Emitida</span>
					</div>
				</div>
			</div>
		</div>
		<div class="conecta" use:anim={'reveal'}>
			<div class="c-fila">
				<span class="c-t">Se conecta con</span>
				<span class="c-logos">
					{#each MARCAS as [n, src, prop] (n)}
						<img class="c-marca" {src} alt={n} height="20" width={Math.round(20 * prop)} />
					{/each}
					{#each LOGOS as l (l.n)}
						<span class="c-logo" title={l.n}
							><Icon icon={l.icon} size={22} style={l.col ? `color:${l.col}` : ''} /><span
								class="sr">{l.n}</span
							></span
						>
					{/each}
				</span>
			</div>
			<div class="c-fila">
				<span class="c-t">Y la manejás desde tu IA</span>
				<span class="c-logos">
					{#each IAS as [n, ico, col] (n)}
						<span class="c-ia"><Icon icon={ico} size={18} style="color:{col}" />{n}</span>
					{/each}
				</span>
			</div>
		</div>
	</div>
	<CtaTienda />
</section>

<style>
	section {
		background: #f6f5f3;
		border-top: 1px solid rgba(58, 29, 18, 0.08);
		border-bottom: 1px solid rgba(58, 29, 18, 0.08);
	}
	.in {
		max-width: 1240px;
		margin: 0 auto;
		padding: 120px var(--lado);
	}
	.cab {
		max-width: 720px;
	}
	.etiq {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 14px;
		font-weight: 600;
		color: #b0441d;
	}
	.etiq span {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		background: #d2582f;
	}
	h2 {
		margin: 14px 0 0;
		font-size: clamp(36px, 4.6vw, 60px);
		line-height: 1;
		letter-spacing: -0.04em;
		font-weight: 600;
		text-wrap: balance;
	}
	.bento {
		margin-top: 56px;
		display: flex;
		flex-wrap: wrap;
		gap: 16px;
	}
	.caja {
		position: relative;
		overflow: hidden;
		border-radius: 26px;
		box-shadow:
			0 1px 2px rgba(58, 29, 18, 0.06),
			0 24px 50px -30px rgba(58, 29, 18, 0.4);
	}
	.sello {
		display: flex;
		align-items: center;
		gap: 10px;
	}
	.s-n {
		font-size: 14px;
		font-weight: 600;
		color: #5c4536;
	}
	.s-n.blanco {
		color: inherit;
	}
	.incl {
		font-size: 12px;
		font-weight: 600;
		color: #237550;
		background: rgba(47, 143, 99, 0.12);
		padding: 3px 9px;
		border-radius: 999px;
	}
	h3 {
		margin: 14px 0 0;
		font-size: 30px;
		letter-spacing: -0.03em;
		font-weight: 600;
		max-width: 440px;
		line-height: 1.08;
	}
	.wa {
		flex: 1 1 620px;
		min-height: 460px;
		background:
			radial-gradient(
				520px circle at var(--mx, 70%) var(--my, 0%),
				rgba(232, 135, 95, 0.16),
				transparent 60%
			),
			#fff;
		padding: 32px;
		display: flex;
		flex-direction: column;
	}
	.wa > p {
		margin: 12px 0 0;
		color: #5c4536;
		font-size: 15.5px;
		line-height: 1.55;
		max-width: 440px;
	}
	.bandeja {
		margin-top: auto;
		padding-top: 28px;
		display: grid;
		grid-template-columns: 210px 1fr;
		border-radius: 18px;
		overflow: hidden;
		border: 1px solid rgba(58, 29, 18, 0.08);
		background: #fbf7f2;
		box-shadow: 0 20px 40px -24px rgba(58, 29, 18, 0.4);
	}
	.lista {
		border-right: 1px solid rgba(58, 29, 18, 0.08);
		padding: 10px;
		display: flex;
		flex-direction: column;
		gap: 4px;
		background: #fff;
	}
	.chat-f {
		padding: 9px;
		border-radius: 12px;
		display: flex;
		gap: 10px;
		align-items: center;
	}
	.chat-f.act {
		background: #fcede5;
	}
	.chat-f img {
		width: 30px;
		height: 30px;
		border-radius: 50%;
		object-fit: cover;
	}
	.cf-txt {
		flex: 1;
		min-width: 0;
	}
	.cf-n {
		font-size: 13px;
		font-weight: 600;
	}
	.cf-m {
		font-size: 12px;
		color: #7a6252;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.hilo {
		padding: 14px;
		display: flex;
		flex-direction: column;
		gap: 8px;
		background: #efe7dc;
	}
	.b-c {
		align-self: flex-start;
		max-width: 80%;
		font-size: 13px;
		line-height: 1.4;
		background: #fff;
		padding: 8px 12px;
		border-radius: 14px 14px 14px 4px;
		box-shadow: 0 1px 1px rgba(0, 0, 0, 0.05);
	}
	.b-v {
		align-self: flex-end;
		max-width: 80%;
		font-size: 13px;
		line-height: 1.4;
		background: #d8f3df;
		padding: 8px 12px;
		border-radius: 14px 14px 4px 14px;
	}
	.b-link {
		align-self: flex-end;
		width: 230px;
		border-radius: 14px;
		background: #d8f3df;
		padding: 6px;
	}
	.bl-prod {
		border-radius: 10px;
		background: #fff;
		padding: 10px;
		display: flex;
		gap: 10px;
		align-items: center;
	}
	.bl-prod img {
		width: 38px;
		height: 38px;
		border-radius: 8px;
		object-fit: cover;
	}
	.bl-txt {
		flex: 1;
	}
	.bl-txt .a {
		font-size: 12.5px;
		font-weight: 600;
	}
	.bl-txt .b {
		font-size: 13px;
		font-weight: 600;
		color: #b0441d;
	}
	.bl-pago {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 6px;
		font-size: 12.5px;
		font-weight: 600;
		color: #1f6f4a;
		padding: 8px 0 2px;
	}
	.pos {
		flex: 1 1 360px;
		min-height: 460px;
	}
	.pos > img {
		position: absolute;
		inset: 0;
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	.pos-sombra {
		position: absolute;
		inset: 0;
		background: linear-gradient(180deg, rgba(42, 20, 12, 0.15), rgba(42, 20, 12, 0.82) 75%);
	}
	.pos-in {
		position: relative;
		height: 100%;
		min-height: 460px;
		padding: 32px;
		display: flex;
		flex-direction: column;
		color: #fff;
	}
	.extra {
		font-size: 12px;
		font-weight: 600;
		color: #2a140c;
		background: #f5b796;
		padding: 3px 9px;
		border-radius: 999px;
	}
	.pos-h {
		margin: auto 0 0;
	}
	.pos-p {
		margin: 10px 0 0;
		font-size: 15.5px;
		line-height: 1.55;
		color: #eadbcf;
	}
	.medios {
		margin-top: 18px;
		display: flex;
		gap: 8px;
		flex-wrap: wrap;
	}
	.medios span {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 7px 12px;
		border-radius: 999px;
		background: rgba(255, 255, 255, 0.95);
		color: #2a140c;
		font-size: 13px;
		font-weight: 600;
	}
	.ia {
		flex: 1 1 520px;
		min-height: 340px;
		background: linear-gradient(140deg, #d2582f, #b9481f);
		color: #fff;
		box-shadow:
			0 1px 2px rgba(58, 29, 18, 0.06),
			0 24px 50px -24px rgba(210, 88, 47, 0.6);
		padding: 32px;
		display: flex;
		gap: 28px;
		flex-wrap: wrap;
		align-items: center;
	}
	.ia-txt {
		position: relative;
		flex: 1 1 260px;
	}
	.excl {
		font-size: 12px;
		font-weight: 700;
		color: #b0441d;
		background: #fff;
		padding: 3px 9px;
		border-radius: 999px;
	}
	.ia-p {
		margin: 12px 0 0;
		color: #fce3d6;
		font-size: 15.5px;
		line-height: 1.55;
	}
	.preguntas {
		margin-top: 18px;
		display: flex;
		gap: 8px;
		flex-wrap: wrap;
	}
	.preguntas span {
		font-size: 13px;
		font-weight: 500;
		padding: 7px 12px;
		border-radius: 999px;
		background: rgba(255, 255, 255, 0.16);
		border: 1px solid rgba(255, 255, 255, 0.3);
	}
	.fact {
		flex: 1 1 340px;
		min-height: 340px;
		background:
			radial-gradient(
				420px circle at var(--mx, 50%) var(--my, 0%),
				rgba(232, 135, 95, 0.14),
				transparent 60%
			),
			#fff;
		padding: 32px;
		display: flex;
		flex-direction: column;
	}
	.fact-h {
		font-size: 26px;
		line-height: 1.12;
		max-width: none;
	}
	.filas {
		margin-top: auto;
		padding-top: 20px;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.fila {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 11px 14px;
		border-radius: 12px;
		background: #fbf7f2;
	}
	.fila.arca {
		background: #fcede5;
	}
	.f-n {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 14px;
		font-weight: 500;
	}
	.ok {
		font-size: 13px;
		font-weight: 600;
		color: #237550;
	}
	.emitida {
		font-size: 13px;
		font-weight: 600;
		color: #b0441d;
	}
	@media (max-width: 520px) {
		.bandeja {
			grid-template-columns: 1fr;
		}
	}
	.conecta {
		margin-top: 40px;
		display: flex;
		flex-direction: column;
		gap: 14px;
	}
	.c-fila {
		display: flex;
		align-items: center;
		gap: 12px 20px;
		flex-wrap: wrap;
		padding: 16px 20px;
		border-radius: 22px;
		background: #fff;
		border: 1px solid rgba(58, 29, 18, 0.08);
	}
	.c-t {
		font-size: 14px;
		font-weight: 600;
		color: #7a6252;
		white-space: nowrap;
	}
	.c-logos {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 10px 18px;
	}
	.c-logo {
		display: grid;
		place-items: center;
	}
	.c-ia {
		display: flex;
		align-items: center;
		gap: 7px;
		font-size: 14.5px;
		font-weight: 600;
	}
	.sr {
		position: absolute;
		width: 1px;
		height: 1px;
		overflow: hidden;
		clip-path: inset(50%);
		white-space: nowrap;
	}
	.c-marca {
		height: 20px;
		width: auto;
		display: block;
	}
</style>
