<script lang="ts">
	import CtaTienda from './CtaTienda.svelte';
	import Icon from '#lib/components/Icon.svelte';
	import { anim } from '#lib/anim.ts';
	import * as I from '#lib/icons.ts';
	import { CHAT } from './datos.ts';
	import { getLanding } from './estado.svelte.ts';
	import { cara, unsplash } from './formato.ts';

	/** Sección "Conocé a Tincho": el título, lo que hace en cuatro renglones y el chat demo animado. */
	const l = getLanding();
	const visibles = $derived(CHAT.slice(0, Math.min(l.chat, CHAT.length)));
	const siguiente = $derived(CHAT[l.chat]);
	const escribiendo = $derived(!!siguiente && ['h', 'prod', 'pay'].includes(siguiente.k));

	const CAPACIDADES = [
		[I.phChatsCircleFill, 'Responde y cierra ventas'],
		[I.phShoppingCartSimpleFill, 'Recupera carritos abandonados'],
		[I.phCameraFill, 'Carga productos con una foto'],
		[I.phLightbulbFilamentFill, 'Te dice qué hacer hoy']
	] as const;

	/** Cada mensaje nuevo entra como en el handoff (translateY 12px, scale .97, 420 ms). */
	function entra(el: HTMLElement) {
		el.animate(
			[
				{ opacity: 0, transform: 'translateY(12px) scale(.97)' },
				{ opacity: 1, transform: 'none' }
			],
			{ duration: 420, easing: 'cubic-bezier(.2,.7,.1,1)' }
		);
	}
</script>

<section id="hornero">
	<div class="in">
		<div class="dos">
			<div class="cab" use:anim={'reveal'}>
				<div class="sello"><Icon icon={I.phSparkleFill} size={15} />Asistente con IA</div>
				<h2>Conocé a Tincho. Atiende y vende las 24&nbsp;horas.</h2>
				<p>
					Contesta en WhatsApp, Instagram y Telegram con tus precios y tu stock reales. Lo
					importante te lo&nbsp;consulta.
				</p>
				<ul class="hace">
					{#each CAPACIDADES as [ico, t] (t)}
						<li><span class="h-ico"><Icon icon={ico} size={18} /></span>{t}</li>
					{/each}
				</ul>
				<CtaTienda texto="Probar a Tincho gratis" izquierda />
			</div>
			<div class="chat" use:anim={{ t: 'reveal', delay: 120 }}>
				<div class="chat-cab">
					<img src={cara('1531058240690-006c446962d8')} alt="Lucía Fernández" />
					<div class="cab-txt">
						<div class="nom">Lucía Fernández <Icon icon={I.logosWhatsappIcon} size={14} /></div>
						<div class="sub">Cliente nueva · llegó por Instagram</div>
					</div>
					<span class="atiende"
						><span class="av"><img src="/hornero/avatar.webp" alt="Tincho" /></span>Tincho
						atendiendo</span
					>
				</div>
				<div class="mensajes" aria-live="polite">
					{#each visibles as m, i (i)}
						{#if m.k === 'c'}
							<div class="m-c" use:entra>{m.t}</div>
						{:else if m.k === 'h'}
							<div class="m-h" use:entra>
								<span class="firma"><Icon icon={I.phSparkleFill} size={11} />Tincho</span><span
									class="burbuja">{m.t}</span
								>
							</div>
						{:else if m.k === 'prod'}
							<div class="m-tarjeta" use:entra>
								<div class="prod">
									<img
										src={unsplash('1573612664822-d7d347da7b80', 500, '&h=260')}
										alt="Interior de un local de ropa y regalos"
									/>
									<div class="prod-pie">
										<div>
											<div class="prod-n">Termo 1 L · Negro</div>
											<div class="prod-s">Quedan 3</div>
										</div>
										<span class="prod-p">$ 42.900</span>
									</div>
								</div>
							</div>
						{:else if m.k === 'pay'}
							<div class="m-tarjeta" use:entra>
								<div class="pago">
									<span class="mp"
										><Icon icon={I.siMercadopago} size={20} style="color:#00B1EA" /></span
									>
									<div class="pago-txt">
										<div class="a">Pagar $ 44.600</div>
										<div class="b">Hasta 6 cuotas sin interés</div>
									</div>
									<Icon icon={I.phArrowUpRightBold} size={15} style="color:#b0441d" />
								</div>
							</div>
						{:else}
							<div class="m-sys" use:entra><Icon icon={I.phCheckCircleFill} size={15} />{m.t}</div>
						{/if}
					{/each}
					{#if escribiendo}
						<div class="puntos" role="status" aria-label="Tincho está escribiendo">
							<span use:anim={'dots'}></span><span use:anim={{ t: 'dots', delay: 180 }}></span><span
								use:anim={{ t: 'dots', delay: 360 }}
							></span>
						</div>
					{/if}
				</div>
				<div class="chat-pie">
					<span>Tincho contesta solo. Vos podés intervenir cuando quieras.</span>
					<span class="tomar">Tomar la conversación</span>
				</div>
			</div>
		</div>
	</div>
</section>

<style>
	section {
		position: relative;
		overflow: hidden;
		background: #fff;
		border-bottom: 1px solid rgba(58, 29, 18, 0.08);
	}
	.in {
		position: relative;
		max-width: 1240px;
		margin: 0 auto;
		padding: 96px var(--lado);
	}
	.cab {
		flex: 1 1 380px;
		max-width: 520px;
	}
	.hace {
		margin: 26px 0 0;
		padding: 0;
		list-style: none;
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.hace li {
		display: flex;
		align-items: center;
		gap: 12px;
		font-size: 16px;
		font-weight: 600;
	}
	.h-ico {
		width: 34px;
		height: 34px;
		flex: none;
		border-radius: 11px;
		display: grid;
		place-items: center;
		background: rgba(210, 88, 47, 0.12);
		color: #c24e25;
	}
	.sello {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		font-size: 14px;
		font-weight: 600;
		color: #fff;
		background: #d2582f;
		padding: 7px 14px;
		border-radius: 999px;
		box-shadow: 0 10px 24px -12px rgba(210, 88, 47, 0.8);
	}
	h2 {
		text-wrap: balance;
		margin: 18px 0 0;
		font-size: clamp(36px, 4.4vw, 58px);
		line-height: 0.98;
		letter-spacing: -0.045em;
		font-weight: 600;
	}
	.cab p {
		text-wrap: pretty;
		margin: 18px 0 0;
		font-size: 19px;
		line-height: 1.55;
		color: #5c4536;
		max-width: 640px;
	}
	.dos {
		display: flex;
		gap: 48px 64px;
		flex-wrap: wrap;
		align-items: center;
	}
	.chat {
		flex: 1 1 420px;
		max-width: 540px;
		margin: 0 auto;
		border-radius: 28px;
		background: #fff;
		overflow: hidden;
		box-shadow:
			0 2px 4px rgba(58, 29, 18, 0.06),
			0 40px 80px -30px rgba(58, 29, 18, 0.5);
	}
	.chat-cab {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 14px 16px;
		border-bottom: 1px solid rgba(58, 29, 18, 0.08);
	}
	.chat-cab > img {
		width: 40px;
		height: 40px;
		border-radius: 50%;
		object-fit: cover;
	}
	.cab-txt {
		flex: 1;
		min-width: 0;
	}
	.nom {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 15px;
		font-weight: 600;
	}
	.sub {
		font-size: 12.5px;
		color: #7a6252;
	}
	.atiende {
		display: flex;
		align-items: center;
		gap: 7px;
		font-size: 12.5px;
		font-weight: 600;
		color: #b0441d;
		background: rgba(210, 88, 47, 0.1);
		padding: 6px 10px 6px 6px;
		border-radius: 999px;
		white-space: nowrap;
	}
	.av {
		width: 22px;
		height: 22px;
		border-radius: 50%;
		overflow: hidden;
		background: #fcede5;
		box-shadow: 0 0 0 1.5px #d2582f;
	}
	.av img {
		display: block;
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	.mensajes {
		height: 440px;
		padding: 18px 16px;
		display: flex;
		flex-direction: column;
		justify-content: flex-end;
		gap: 8px;
		overflow: hidden;
		background: #f4ece2;
	}
	.m-c {
		align-self: flex-start;
		max-width: 78%;
		font-size: 14px;
		line-height: 1.45;
		background: #fff;
		padding: 9px 13px;
		border-radius: 16px 16px 16px 4px;
		box-shadow: 0 1px 1px rgba(0, 0, 0, 0.05);
	}
	.m-h {
		align-self: flex-end;
		max-width: 80%;
		display: flex;
		flex-direction: column;
		align-items: flex-end;
		gap: 3px;
	}
	.firma {
		font-size: 11px;
		font-weight: 600;
		color: #b0441d;
		display: flex;
		align-items: center;
		gap: 4px;
	}
	.burbuja {
		font-size: 14px;
		line-height: 1.45;
		background: #fce3d6;
		padding: 9px 13px;
		border-radius: 16px 16px 4px 16px;
	}
	.m-tarjeta {
		align-self: flex-end;
		width: 250px;
		border-radius: 16px;
		background: #fce3d6;
		padding: 6px;
	}
	.prod {
		border-radius: 12px;
		background: #fff;
		overflow: hidden;
	}
	.prod img {
		width: 100%;
		height: 110px;
		object-fit: cover;
		display: block;
	}
	.prod-pie {
		padding: 10px 12px;
		display: flex;
		justify-content: space-between;
		align-items: center;
	}
	.prod-n {
		font-size: 13.5px;
		font-weight: 600;
	}
	.prod-s {
		font-size: 12px;
		color: #7a6252;
	}
	.prod-p {
		font-size: 15px;
		font-weight: 600;
		color: #b0441d;
	}
	.pago {
		border-radius: 12px;
		background: #fff;
		padding: 12px;
		display: flex;
		align-items: center;
		gap: 10px;
	}
	.mp {
		width: 38px;
		height: 38px;
		border-radius: 10px;
		background: #e6f7fd;
		display: grid;
		place-items: center;
	}
	.pago-txt {
		flex: 1;
	}
	.pago-txt .a {
		font-size: 13.5px;
		font-weight: 600;
	}
	.pago-txt .b {
		font-size: 12px;
		color: #7a6252;
	}
	.m-sys {
		align-self: center;
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 12.5px;
		font-weight: 600;
		color: #1f6f4a;
		background: #e3f4ea;
		padding: 8px 14px;
		border-radius: 999px;
		margin-top: 4px;
	}
	.puntos {
		align-self: flex-end;
		display: flex;
		gap: 4px;
		padding: 12px 14px;
		border-radius: 16px 16px 4px 16px;
		background: #fce3d6;
	}
	.puntos span {
		width: 7px;
		height: 7px;
		border-radius: 50%;
		background: #c24e25;
	}
	.chat-pie {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 12px 14px;
		border-top: 1px solid rgba(58, 29, 18, 0.08);
	}
	.chat-pie > span:first-child {
		flex: 1;
		font-size: 13px;
		color: #7a6252;
	}
	.tomar {
		font-size: 13px;
		font-weight: 600;
		padding: 8px 14px;
		border-radius: 999px;
		background: #2a140c;
		color: #fff;
		white-space: nowrap;
	}
	/* Celular: la escena se compacta para que las métricas no se salgan, y la cabecera del chat
	   pasa el "Tincho atendiendo" abajo en vez de estrujar el nombre. */
	@media (max-width: 640px) {
		.dos {
			gap: 32px;
		}
		.chat-cab {
			flex-wrap: wrap;
			row-gap: 8px;
		}
		.cab-txt {
			flex: 1 1 calc(100% - 52px);
		}
		.atiende {
			margin-left: 52px;
		}
	}
</style>
