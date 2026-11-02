<script lang="ts">
	import Seo from '#lib/components/Seo.svelte';
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import DescargarTodo from '#lib/marca/DescargarTodo.svelte';
	import Lamina from '#lib/marca/Lamina.svelte';
	import { DESDE, PRUEBA_DIAS, VS_PRO, anioHotSale, fmt } from '#lib/marca/datos.ts';
	import { crearFamilia } from '#lib/marca/familia.svelte.ts';

	/** Redes (design/Redes Mostrador.dc.html): cada pieza a su medida en px, mostrada al 30–50 %. */
	const fam = crearFamilia('redes');
	const ANIO_HOT_SALE = anioHotSale();
</script>

<Seo
	title="Redes · Marca · Tu Local Web"
	description="Avatares, tapas de destacadas, portadas, anuncios para Meta y miniaturas de YouTube."
	path="/marca/redes"
	noindex
/>

<section class="lienzo" bind:contentRect={fam.rect}>
	<header class="lienzo-cab">
		<div style="flex:1;min-width:min(100%,300px)">
			<img src="/wordmark-brown.png" alt="Tu Local Web" style="height:26px;display:block" />
			<h1
				style="margin:22px 0 0;font-size:56px;line-height:1;letter-spacing:-.04em;font-weight:600"
			>
				Redes
			</h1>
			<p style="margin:12px 0 0;font-size:18px;color:#6B5244;max-width:680px;text-wrap:pretty">
				Portadas, avatares, anuncios para Meta y miniaturas de YouTube. Se muestran al 40–50 %; cada
				pieza está armada en su medida real. Las institucionales usan crema y tostado; las jugadas,
				terracota a&nbsp;pleno.
			</p>
		</div>
		<DescargarTodo />
	</header>

	<!-- AVATARES -->
	<div style="display:flex;flex-direction:column;gap:20px">
		<h2 style="margin:0;font-size:26px;letter-spacing:-.03em;font-weight:600">
			Avatares de perfil · 400 × 400
		</h2>
		<div style="display:flex;flex-wrap:wrap;gap:40px;align-items:flex-start">
			<div id="1a" style="display:flex;flex-direction:column;gap:12px">
				<div style="display:flex;align-items:center;gap:10px">
					<a
						href="#1a"
						style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
						>1a</a
					><span style="font-size:14px;color:#6B5244">Marca · todas las redes</span>
				</div>
				<div style="display:flex;gap:20px;align-items:center">
					<Lamina nombre="redes-1a-avatar" ancho={200} alto={200} escalaPng={2}
						><div
							style="width:200px;height:200px;border-radius:50%;background:#D2582F;display:grid;place-items:center;position:relative;overflow:hidden;box-shadow:0 18px 40px -18px rgba(42,20,12,.5)"
						>
							<div
								style="position:absolute;inset:0;background:url(/textura/grain.png);opacity:.22;mix-blend-mode:multiply"
							></div>
							<img
								src="/logo.svg"
								alt=""
								style="width:96px;position:relative;filter:brightness(0) invert(1);margin-top:6px"
							/>
						</div></Lamina
					>
					<div
						style="width:64px;height:64px;border-radius:50%;background:#D2582F;display:grid;place-items:center"
					>
						<img src="/logo.svg" alt="" style="width:31px;filter:brightness(0) invert(1)" />
					</div>
					<div
						style="width:32px;height:32px;border-radius:50%;background:#D2582F;display:grid;place-items:center"
					>
						<img src="/logo.svg" alt="" style="width:16px;filter:brightness(0) invert(1)" />
					</div>
				</div>
			</div>
			<div id="1b" style="display:flex;flex-direction:column;gap:12px">
				<div style="display:flex;align-items:center;gap:10px">
					<a
						href="#1b"
						style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
						>1b</a
					><span style="font-size:14px;color:#6B5244">Tincho · soporte y cuenta de tips</span>
				</div>
				<div style="display:flex;gap:20px;align-items:center">
					<Lamina nombre="redes-1b-avatar-hornero" ancho={200} alto={200} escalaPng={2}
						><div
							style="width:200px;height:200px;border-radius:50%;background:#FBF7F2;display:grid;place-items:center;overflow:hidden;box-shadow:inset 0 0 0 10px #D2582F,0 18px 40px -18px rgba(42,20,12,.5)"
						>
							<img
								src="/hornero/avatar.png"
								alt=""
								style="width:170px;height:170px;object-fit:contain"
							/>
						</div></Lamina
					>
					<div
						style="width:64px;height:64px;border-radius:50%;background:#FBF7F2;display:grid;place-items:center;overflow:hidden;box-shadow:inset 0 0 0 3px #D2582F"
					>
						<img
							src="/hornero/avatar.png"
							alt=""
							style="width:54px;height:54px;object-fit:contain"
						/>
					</div>
				</div>
			</div>
			<div id="1c" style="display:flex;flex-direction:column;gap:12px">
				<div style="display:flex;align-items:center;gap:10px">
					<a
						href="#1c"
						style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
						>1c</a
					><span style="font-size:14px;color:#6B5244">Tapas de destacadas de Instagram</span>
				</div>
				<div style="display:flex;gap:16px;flex-wrap:wrap">
					<div style="display:flex;flex-direction:column;align-items:center;gap:6px">
						<Lamina nombre="redes-1c-precios" ancho={84} alto={84} escalaPng={1080 / 84} compacta
							><div
								style="width:84px;height:84px;border-radius:50%;background:#2A140C;color:#E8875F;display:grid;place-items:center;box-shadow:0 0 0 3px #E7DED3,0 0 0 5px #D9CBBE"
							>
								<Icon icon={I.phTagBold} size={34} />
							</div></Lamina
						><span style="font-size:12px">Precios</span>
					</div>
					<div style="display:flex;flex-direction:column;align-items:center;gap:6px">
						<Lamina nombre="redes-1c-hornero" ancho={84} alto={84} escalaPng={1080 / 84} compacta
							><div
								style="width:84px;height:84px;border-radius:50%;background:#2A140C;color:#E8875F;display:grid;place-items:center;box-shadow:0 0 0 3px #E7DED3,0 0 0 5px #D9CBBE"
							>
								<Icon icon={I.phSparkleBold} size={34} />
							</div></Lamina
						><span style="font-size:12px">Tincho</span>
					</div>
					<div style="display:flex;flex-direction:column;align-items:center;gap:6px">
						<Lamina nombre="redes-1c-tiendas" ancho={84} alto={84} escalaPng={1080 / 84} compacta
							><div
								style="width:84px;height:84px;border-radius:50%;background:#2A140C;color:#E8875F;display:grid;place-items:center;box-shadow:0 0 0 3px #E7DED3,0 0 0 5px #D9CBBE"
							>
								<Icon icon={I.phStorefrontBold} size={34} />
							</div></Lamina
						><span style="font-size:12px">Tiendas</span>
					</div>
					<div style="display:flex;flex-direction:column;align-items:center;gap:6px">
						<Lamina nombre="redes-1c-tips" ancho={84} alto={84} escalaPng={1080 / 84} compacta
							><div
								style="width:84px;height:84px;border-radius:50%;background:#2A140C;color:#E8875F;display:grid;place-items:center;box-shadow:0 0 0 3px #E7DED3,0 0 0 5px #D9CBBE"
							>
								<Icon icon={I.phLightbulbBold} size={34} />
							</div></Lamina
						><span style="font-size:12px">Tips</span>
					</div>
					<div style="display:flex;flex-direction:column;align-items:center;gap:6px">
						<Lamina nombre="redes-1c-migrar" ancho={84} alto={84} escalaPng={1080 / 84} compacta
							><div
								style="width:84px;height:84px;border-radius:50%;background:#2A140C;color:#E8875F;display:grid;place-items:center;box-shadow:0 0 0 3px #E7DED3,0 0 0 5px #D9CBBE"
							>
								<Icon icon={I.phArrowsLeftRightBold} size={34} />
							</div></Lamina
						><span style="font-size:12px">Migrar</span>
					</div>
				</div>
			</div>
		</div>
	</div>

	<!-- PORTADAS -->
	<div style="display:flex;flex-direction:column;gap:20px">
		<h2 style="margin:0;font-size:26px;letter-spacing:-.03em;font-weight:600">Portadas</h2>
		<div style="display:flex;flex-direction:column;gap:40px">
			<div id="2a" style="display:flex;flex-direction:column;gap:12px">
				<div style="display:flex;align-items:center;gap:10px">
					<a
						href="#2a"
						style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
						>2a</a
					><span style="font-size:14px;color:#6B5244">LinkedIn · 1584 × 396 · al 50 %</span>
				</div>
				<Lamina
					nombre="redes-2a-portada-linkedin"
					ancho={1584}
					alto={396}
					escala={0.5}
					radio={6}
					sombra="0 18px 40px -20px rgba(42,20,12,.45)"
					><div
						style="width:1584px;height:396px;background:#FBF7F2;position:relative;overflow:hidden"
					>
						<div
							style="position:absolute;inset:0;background:url(/textura/arch-pattern-light.svg) right center/80px;opacity:.6;-webkit-mask-image:linear-gradient(90deg,transparent 35%,#000 75%);mask-image:linear-gradient(90deg,transparent 35%,#000 75%)"
						></div>
						<div style="position:absolute;left:520px;top:92px;width:560px">
							<div style="font-size:64px;line-height:1;letter-spacing:-.04em;font-weight:600">
								La plataforma para los que <span style="color:#D2582F">venden en serio.</span>
							</div>
							<div style="margin-top:20px;font-size:22px;color:#6B5244">
								Tienda online, Mercado Libre, WhatsApp y facturación. Una sola cuenta.
							</div>
						</div>
						<div
							style="position:absolute;right:120px;top:40px;width:250px;height:356px;border-radius:125px 125px 0 0;overflow:hidden"
						>
							<img
								src="https://images.unsplash.com/photo-1556745753-b2904692b3cd?auto=format&fit=crop&w=600&q=75"
								alt=""
								style="width:100%;height:100%;object-fit:cover"
							/>
						</div>
						<img
							src="/hornero/saludo.png"
							alt=""
							style="position:absolute;right:330px;bottom:-10px;width:170px;height:170px;object-fit:contain"
						/>
					</div></Lamina
				>
			</div>
			<div id="2b" style="display:flex;flex-direction:column;gap:12px">
				<div style="display:flex;align-items:center;gap:10px">
					<a
						href="#2b"
						style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
						>2b</a
					><span style="font-size:14px;color:#6B5244">X · 1500 × 500 · al 50 %</span>
				</div>
				<Lamina
					nombre="redes-2b-portada-x"
					ancho={1500}
					alto={500}
					escala={0.5}
					radio={6}
					sombra="0 18px 40px -20px rgba(42,20,12,.5)"
					><div
						style="width:1500px;height:500px;background:#D2582F;position:relative;overflow:hidden;color:#FBF7F2"
					>
						<div
							style="position:absolute;inset:0;background:url(/textura/arch-pattern-accent.svg) center/110px;opacity:.7"
						></div>
						<div
							style="position:absolute;inset:0;background:url(/textura/grain.png);opacity:.22;mix-blend-mode:multiply"
						></div>
						<div style="position:absolute;left:560px;top:120px">
							<div style="font-size:88px;line-height:.95;letter-spacing:-.045em;font-weight:600">
								Todo lo de las otras.<br /><span style="color:#2A140C">Por menos plata.</span>
							</div>
						</div>
						<img
							src="/hornero/celebrando.png"
							alt=""
							style="position:absolute;right:60px;bottom:-20px;width:300px;height:300px;object-fit:contain"
						/>
					</div></Lamina
				>
			</div>
			<div id="2c" style="display:flex;flex-direction:column;gap:12px">
				<div style="display:flex;align-items:center;gap:10px">
					<a
						href="#2c"
						style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
						>2c</a
					><span style="font-size:14px;color:#6B5244"
						>YouTube · 2560 × 1440 · al 30 % · la franja punteada es lo que se ve en todos los
						dispositivos</span
					>
				</div>
				<Lamina
					nombre="redes-2c-portada-youtube"
					ancho={2560}
					alto={1440}
					escala={0.3}
					radio={6}
					sombra="0 18px 40px -20px rgba(42,20,12,.5)"
					><div
						style="width:2560px;height:1440px;background:#2A140C;position:relative;overflow:hidden;color:#FBF7F2"
					>
						<div
							style="position:absolute;inset:0;background:url(/textura/grain.png);opacity:.15"
						></div>
						<div
							style="position:absolute;inset:0;background:url(/textura/arch-pattern-dark.svg) center/160px;opacity:.5"
						></div>
						<div
							style="position:absolute;left:507px;right:507px;top:509px;height:423px;border:4px dashed rgba(232,135,95,.6);display:flex;align-items:center;gap:60px;padding:0 70px;background:#2A140C"
						>
							<img src="/logo.svg" alt="" style="width:170px;flex:none" />
							<div style="flex:1">
								<div style="font-size:96px;line-height:1;letter-spacing:-.04em;font-weight:600">
									Aprendé a vender más.
								</div>
								<div style="margin-top:18px;font-size:38px;color:#BDB2A8">
									Tutoriales y casos reales, todos los jueves.
								</div>
							</div>
							<img
								src="/hornero/hablando.png"
								alt=""
								style="width:300px;height:300px;object-fit:contain;flex:none"
							/>
						</div>
					</div></Lamina
				>
			</div>
		</div>
	</div>

	<!-- ADS -->
	<div style="display:flex;flex-direction:column;gap:20px">
		<h2 style="margin:0;font-size:26px;letter-spacing:-.03em;font-weight:600">
			Anuncios para Meta · feed 1080 × 1350 y historia 1080 × 1920 · al 40 %
		</h2>
		<div style="display:flex;flex-wrap:wrap;gap:36px;align-items:flex-start">
			<div id="3a" style="display:flex;flex-direction:column;gap:12px">
				<div style="display:flex;align-items:center;gap:10px">
					<a
						href="#3a"
						style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
						>3a</a
					><span style="font-size:14px;color:#6B5244">Lanzamiento · institucional</span>
				</div>
				<Lamina
					nombre="redes-3a-anuncio-meta"
					ancho={1080}
					alto={1350}
					escala={0.4}
					radio={6}
					sombra="0 18px 40px -20px rgba(42,20,12,.45)"
					><div
						style="width:1080px;height:1350px;background:#FBF7F2;position:relative;overflow:hidden;padding:90px"
					>
						<img src="/wordmark-brown.png" alt="Tu Local Web" style="height:46px" />
						<div
							style="margin-top:70px;font-size:112px;line-height:.95;letter-spacing:-.045em;font-weight:600"
						>
							Llegó la tienda que hace todo.
						</div>
						<div style="margin-top:30px;font-size:36px;color:#6B5244;max-width:760px">
							Web, Mercado Libre, WhatsApp, facturación y envíos. Desde {fmt(DESDE)} por mes.
						</div>
						<div
							style="position:absolute;left:90px;right:90px;bottom:90px;height:520px;border-radius:260px 260px 24px 24px;overflow:hidden"
						>
							<img
								src="https://images.unsplash.com/photo-1556745753-b2904692b3cd?auto=format&fit=crop&w=1100&q=75"
								alt=""
								style="width:100%;height:100%;object-fit:cover"
							/>
						</div>
						<div
							style="position:absolute;right:130px;bottom:130px;padding:24px 34px;border-radius:999px;background:#2A140C;color:#fff;font-size:34px;font-weight:600"
						>
							Probá {PRUEBA_DIAS} días gratis
						</div>
					</div></Lamina
				>
			</div>

			<div id="3b" style="display:flex;flex-direction:column;gap:12px">
				<div style="display:flex;align-items:center;gap:10px">
					<a
						href="#3b"
						style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
						>3b</a
					><span style="font-size:14px;color:#6B5244">Lanzamiento · jugada</span>
				</div>
				<Lamina
					nombre="redes-3b-anuncio-meta"
					ancho={1080}
					alto={1350}
					escala={0.4}
					radio={6}
					sombra="0 18px 40px -20px rgba(42,20,12,.5)"
					><div
						style="width:1080px;height:1350px;background:#D2582F;position:relative;overflow:hidden;padding:90px;color:#FBF7F2"
					>
						<div
							style="position:absolute;inset:0;background:url(/textura/grain.png);opacity:.22;mix-blend-mode:multiply"
						></div>
						<div style="position:relative;font-size:40px;font-weight:600;opacity:.9">
							Vos a las 2 AM:
						</div>
						<div
							style="position:relative;margin-top:20px;padding:40px 46px;border-radius:40px 40px 40px 8px;background:#FBF7F2;color:#2A140C;font-size:52px;line-height:1.15;letter-spacing:-.02em;font-weight:500;max-width:820px"
						>
							“¿Hay stock del termo negro?”<br />“¿Hacen envíos a Salta?”<br />“¿Puedo pagar en
							cuotas?”
						</div>
						<div
							style="position:relative;margin-top:44px;font-size:104px;line-height:.95;letter-spacing:-.045em;font-weight:600"
						>
							Dormí. Tincho responde.
						</div>
						<img
							src="/hornero/whatsapp.png"
							alt=""
							style="position:absolute;right:30px;bottom:20px;width:440px;height:440px;object-fit:contain"
						/>
						<img
							src="/wordmark-cream.png"
							alt="Tu Local Web"
							style="position:absolute;left:90px;bottom:96px;height:46px"
						/>
					</div></Lamina
				>
			</div>

			<div id="3c" style="display:flex;flex-direction:column;gap:12px">
				<div style="display:flex;align-items:center;gap:10px">
					<a
						href="#3c"
						style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
						>3c</a
					><span style="font-size:14px;color:#6B5244">Precio vs. competencia · institucional</span>
				</div>
				<Lamina
					nombre="redes-3c-anuncio-meta"
					ancho={1080}
					alto={1350}
					escala={0.4}
					radio={6}
					sombra="0 18px 40px -20px rgba(42,20,12,.45)"
					><div
						style="width:1080px;height:1350px;background:#FBF7F2;position:relative;overflow:hidden;padding:90px"
					>
						<div style="font-size:96px;line-height:.95;letter-spacing:-.045em;font-weight:600">
							Lo mismo.<br /><span style="color:#D2582F">Pagando menos.</span>
						</div>
						<div style="margin-top:70px;display:flex;flex-direction:column;gap:26px">
							<div
								style="display:flex;align-items:center;gap:30px;padding:34px 40px;border-radius:32px;background:#fff;box-shadow:0 20px 50px -30px rgba(42,20,12,.4)"
							>
								<span style="flex:1;font-size:38px;color:#6B5244"
									>Tiendanube · plan {VS_PRO.competencia.plan}</span
								><span
									style="font-size:52px;font-weight:600;color:#9C8676;text-decoration:line-through"
									>{fmt(VS_PRO.competencia.precio)}</span
								>
							</div>
							<div
								style="display:flex;align-items:center;gap:30px;padding:40px;border-radius:32px;background:#2A140C;color:#fff"
							>
								<img src="/logo.svg" alt="" style="width:60px" /><span
									style="flex:1;font-size:40px;font-weight:600"
									>Tu Local Web · {VS_PRO.plan.nombre}</span
								><span style="font-size:64px;font-weight:600;color:#E8875F"
									>{fmt(VS_PRO.plan.precio)}</span
								>
							</div>
						</div>
						<div style="position:absolute;left:90px;bottom:90px;font-size:26px;color:#9C8676">
							Precios por mes con IVA, octubre 2026, sin comisión por&nbsp;venta.
						</div>
					</div></Lamina
				>
			</div>

			<div id="3d" style="display:flex;flex-direction:column;gap:12px">
				<div style="display:flex;align-items:center;gap:10px">
					<a
						href="#3d"
						style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
						>3d</a
					><span style="font-size:14px;color:#6B5244">Precio vs. competencia · jugada</span>
				</div>
				<Lamina
					nombre="redes-3d-anuncio-meta"
					ancho={1080}
					alto={1350}
					escala={0.4}
					radio={6}
					sombra="0 18px 40px -20px rgba(42,20,12,.5)"
					><div
						style="width:1080px;height:1350px;background:#2A140C;position:relative;overflow:hidden;padding:90px;color:#FBF7F2"
					>
						<div
							style="position:absolute;inset:0;background:url(/textura/grain.png);opacity:.15"
						></div>
						<div style="position:relative;font-size:44px;color:#BDB2A8">
							Con {VS_PRO.plan.nombre} en vez de {VS_PRO.competencia.nombre}, en un
							año&nbsp;ahorrás:
						</div>
						<div
							style="position:relative;margin-top:24px;font-size:200px;line-height:.9;letter-spacing:-.06em;font-weight:600;color:#E8875F"
						>
							{fmt(VS_PRO.ahorroAnual)}
						</div>
						<div
							style="position:relative;margin-top:50px;display:flex;flex-direction:column;gap:22px;font-size:44px;font-weight:500"
						>
							<span style="display:flex;align-items:center;gap:20px"
								><Icon icon={I.phCheckCircleFill} size={50} style="color:#E8875F" />{VS_PRO.plan
									.nombre}: {fmt(VS_PRO.plan.precio)} por mes</span
							><span style="display:flex;align-items:center;gap:20px"
								><Icon icon={I.phCheckCircleFill} size={50} style="color:#E8875F" />{VS_PRO
									.competencia.plan}: {fmt(VS_PRO.competencia.precio)} por mes</span
							><span style="display:flex;align-items:center;gap:20px"
								><Icon icon={I.phCheckCircleFill} size={50} style="color:#E8875F" />0 % de comisión
								por venta</span
							>
						</div>
						<img
							src="/hornero/caja.png"
							alt=""
							style="position:absolute;right:20px;bottom:10px;width:420px;height:420px;object-fit:contain"
						/>
						<img
							src="/wordmark-cream.png"
							alt="Tu Local Web"
							style="position:absolute;left:90px;bottom:96px;height:46px"
						/>
					</div></Lamina
				>
			</div>

			<div id="3e" style="display:flex;flex-direction:column;gap:12px">
				<div style="display:flex;align-items:center;gap:10px">
					<a
						href="#3e"
						style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
						>3e</a
					><span style="font-size:14px;color:#6B5244">Tincho (IA)</span>
				</div>
				<Lamina
					nombre="redes-3e-anuncio-meta"
					ancho={1080}
					alto={1350}
					escala={0.4}
					radio={6}
					sombra="0 18px 40px -20px rgba(42,20,12,.45)"
					><div
						style="width:1080px;height:1350px;background:linear-gradient(180deg,#FBF7F2 0%,#FCEADF 60%,#F7D9C6 100%);position:relative;overflow:hidden;padding:90px"
					>
						<div
							style="display:inline-flex;align-items:center;gap:14px;padding:14px 26px;border-radius:999px;background:#fff;font-size:30px;font-weight:600;color:#C24E25"
						>
							✦ Tincho · tu vendedor con IA
						</div>
						<div
							style="margin-top:40px;font-size:100px;line-height:.95;letter-spacing:-.045em;font-weight:600"
						>
							Vende, cobra y factura mientras vos hacés otra cosa.
						</div>
						<div
							style="position:absolute;left:90px;right:440px;bottom:110px;display:flex;flex-direction:column;gap:18px"
						>
							<div
								style="align-self:flex-start;padding:26px 32px;border-radius:30px 30px 30px 6px;background:#fff;font-size:32px;box-shadow:0 16px 40px -24px rgba(42,20,12,.5)"
							>
								¿Tenés el termo en negro?
							</div>
							<div
								style="align-self:flex-end;padding:26px 32px;border-radius:30px 30px 6px 30px;background:#2A140C;color:#fff;font-size:32px"
							>
								¡Sí! Quedan 4. ¿Te paso el link de pago?
							</div>
							<div
								style="align-self:flex-end;display:flex;align-items:center;gap:12px;padding:20px 28px;border-radius:999px;background:#2F8F63;color:#fff;font-size:30px;font-weight:600"
							>
								<Icon icon={I.phCheckCircleFill} size={34} />Pago aprobado · $ 42.900
							</div>
						</div>
						<img
							src="/hornero/hablando.png"
							alt=""
							style="position:absolute;right:10px;bottom:20px;width:440px;height:440px;object-fit:contain"
						/>
					</div></Lamina
				>
			</div>

			<div id="3f" style="display:flex;flex-direction:column;gap:12px">
				<div style="display:flex;align-items:center;gap:10px">
					<a
						href="#3f"
						style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
						>3f</a
					><span style="font-size:14px;color:#6B5244">Integración con Caja</span>
				</div>
				<Lamina
					nombre="redes-3f-anuncio-meta"
					ancho={1080}
					alto={1350}
					escala={0.4}
					radio={6}
					sombra="0 18px 40px -20px rgba(42,20,12,.5)"
					><div
						style="width:1080px;height:1350px;background:#131211;position:relative;overflow:hidden;padding:90px;color:#F4EEE6"
					>
						<div style="display:flex;align-items:center;gap:20px;font-size:44px;font-weight:600">
							<img src="/logo.svg" alt="" style="width:64px" /><span style="color:#8A7F76">+</span
							><img
								src="/caja/01-saludo.png"
								alt=""
								style="width:84px;height:84px;object-fit:contain"
							/>Caja
						</div>
						<div
							style="margin-top:60px;font-size:104px;line-height:.95;letter-spacing:-.045em;font-weight:600"
						>
							Tu Local Web vende.<br /><span style="color:#E8875F">Caja ordena la plata.</span>
						</div>
						<div
							style="margin-top:60px;padding:46px;border-radius:40px;background:#F4EEE6;color:#131211;max-width:700px"
						>
							<div style="font-size:28px;color:#6E655D">Ganancia real de octubre</div>
							<div style="font-size:84px;font-weight:600;letter-spacing:-.04em">$ 6.812.400</div>
							<div style="font-size:28px;color:#6E655D">
								≈ US$ 4.412 al MEP · Monotributo al 70 %
							</div>
						</div>
						<div style="position:absolute;left:90px;bottom:90px;font-size:32px;color:#BDB2A8">
							3 meses de Caja Pro Max con cualquier plan.
						</div>
					</div></Lamina
				>
			</div>

			<div id="3g" style="display:flex;flex-direction:column;gap:12px">
				<div style="display:flex;align-items:center;gap:10px">
					<a
						href="#3g"
						style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
						>3g</a
					><span style="font-size:14px;color:#6B5244">Migrar desde Tiendanube · jugada</span>
				</div>
				<Lamina
					nombre="redes-3g-anuncio-meta"
					ancho={1080}
					alto={1350}
					escala={0.4}
					radio={6}
					sombra="0 18px 40px -20px rgba(42,20,12,.45)"
					><div
						style="width:1080px;height:1350px;background:#FBF7F2;position:relative;overflow:hidden;padding:90px"
					>
						<div style="font-size:42px;color:#6B5244">Mudarse de plataforma:</div>
						<div style="margin-top:40px;display:grid;grid-template-columns:1fr 1fr;gap:28px">
							<div style="padding:40px;border-radius:36px;background:#EFE6DC">
								<div style="font-size:32px;font-weight:700;color:#9C8676">Lo que pensás</div>
								<div style="margin-top:20px;font-size:38px;line-height:1.3;color:#6B5244">
									Cargar 400 productos a mano. Perder las reseñas. Un mes sin vender.
								</div>
							</div>
							<div style="padding:40px;border-radius:36px;background:#D2582F;color:#fff">
								<div style="font-size:32px;font-weight:700;opacity:.85">Con Tu Local Web</div>
								<div style="margin-top:20px;font-size:38px;line-height:1.3">
									Pegás el link de tu tienda. En un día está todo. Tu dominio, también.
								</div>
							</div>
						</div>
						<div
							style="margin-top:60px;font-size:100px;line-height:.95;letter-spacing:-.045em;font-weight:600"
						>
							La mudanza la hacemos nosotros.
						</div>
						<img
							src="/hornero/envio.png"
							alt=""
							style="position:absolute;right:30px;bottom:20px;width:400px;height:400px;object-fit:contain"
						/>
						<img
							src="/wordmark-brown.png"
							alt="Tu Local Web"
							style="position:absolute;left:90px;bottom:96px;height:46px"
						/>
					</div></Lamina
				>
			</div>

			<div id="3h" style="display:flex;flex-direction:column;gap:12px">
				<div style="display:flex;align-items:center;gap:10px">
					<a
						href="#3h"
						style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
						>3h</a
					><span style="font-size:14px;color:#6B5244">Tip para vender · carrusel, tapa</span>
				</div>
				<Lamina
					nombre="redes-3h-anuncio-meta"
					ancho={1080}
					alto={1350}
					escala={0.4}
					radio={6}
					sombra="0 18px 40px -20px rgba(42,20,12,.45)"
					><div
						style="width:1080px;height:1350px;background:#FCEDE5;position:relative;overflow:hidden;padding:90px"
					>
						<div style="display:flex;justify-content:space-between;align-items:center">
							<span
								style="padding:14px 26px;border-radius:999px;background:#2A140C;color:#fff;font-size:30px;font-weight:600"
								>Tip #14</span
							><span style="font-size:30px;color:#C24E25;font-weight:600">Deslizá →</span>
						</div>
						<div
							style="margin-top:80px;font-size:120px;line-height:.92;letter-spacing:-.05em;font-weight:600"
						>
							3 fotos que duplican tus ventas.
						</div>
						<div
							style="position:absolute;left:90px;right:90px;bottom:90px;display:grid;grid-template-columns:repeat(3,1fr);gap:20px"
						>
							<div style="height:380px;border-radius:190px 190px 20px 20px;overflow:hidden">
								<img
									src="https://images.unsplash.com/photo-1634741011278-9dc2f5b6fc0b?auto=format&fit=crop&w=400&h=560&q=70"
									alt=""
									style="width:100%;height:100%;object-fit:cover"
								/>
							</div>
							<div style="height:380px;border-radius:190px 190px 20px 20px;overflow:hidden">
								<img
									src="https://images.unsplash.com/photo-1609541994821-d909982e2f1b?auto=format&fit=crop&w=400&h=560&q=70"
									alt=""
									style="width:100%;height:100%;object-fit:cover"
								/>
							</div>
							<div style="height:380px;border-radius:190px 190px 20px 20px;overflow:hidden">
								<img
									src="https://images.unsplash.com/photo-1495474472287-4d71bcdd2085?auto=format&fit=crop&w=400&h=560&q=70"
									alt=""
									style="width:100%;height:100%;object-fit:cover"
								/>
							</div>
						</div>
						<img
							src="/hornero/foto.png"
							alt=""
							style="position:absolute;right:60px;top:520px;width:260px;height:260px;object-fit:contain"
						/>
					</div></Lamina
				>
			</div>

			<div id="3i" style="display:flex;flex-direction:column;gap:12px">
				<div style="display:flex;align-items:center;gap:10px">
					<a
						href="#3i"
						style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
						>3i</a
					><span style="font-size:14px;color:#6B5244">Hot Sale · historia 9:16</span>
				</div>
				<Lamina
					nombre="redes-3i-anuncio-meta"
					ancho={1080}
					alto={1920}
					escala={0.3}
					radio={6}
					sombra="0 18px 40px -20px rgba(42,20,12,.5)"
					><div
						style="width:1080px;height:1920px;background:#D2582F;position:relative;overflow:hidden;padding:250px 90px;color:#FBF7F2"
					>
						<div
							style="position:absolute;inset:0;background:url(/textura/arch-pattern-accent.svg) center/130px;opacity:.6"
						></div>
						<div style="position:relative;font-size:60px;font-weight:700;letter-spacing:.06em">
							HOT SALE {ANIO_HOT_SALE}
						</div>
						<div
							style="position:relative;margin-top:30px;font-size:150px;line-height:.9;letter-spacing:-.05em;font-weight:600"
						>
							Tu tienda lista en 1 tarde.
						</div>
						<div
							style="position:relative;margin-top:50px;display:inline-flex;padding:26px 40px;border-radius:999px;background:#2A140C;font-size:44px;font-weight:600"
						>
							Probá {PRUEBA_DIAS} días gratis
						</div>
						<img
							src="/hornero/celebrando.png"
							alt=""
							style="position:absolute;left:50%;transform:translateX(-50%);bottom:260px;width:640px;height:640px;object-fit:contain"
						/>
						<div
							style="position:absolute;left:90px;right:90px;bottom:250px;text-align:center;font-size:36px;opacity:.9"
						>
							Deslizá para arriba
						</div>
					</div></Lamina
				>
			</div>

			<div id="3j" style="display:flex;flex-direction:column;gap:12px">
				<div style="display:flex;align-items:center;gap:10px">
					<a
						href="#3j"
						style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
						>3j</a
					><span style="font-size:14px;color:#6B5244">CyberMonday · historia 9:16</span>
				</div>
				<Lamina
					nombre="redes-3j-anuncio-meta"
					ancho={1080}
					alto={1920}
					escala={0.3}
					radio={6}
					sombra="0 18px 40px -20px rgba(42,20,12,.5)"
					><div
						style="width:1080px;height:1920px;background:#2A140C;position:relative;overflow:hidden;padding:250px 90px;color:#FBF7F2"
					>
						<div
							style="position:absolute;inset:0;background:url(/textura/grain.png);opacity:.15"
						></div>
						<div
							style="position:relative;font-size:60px;font-weight:700;letter-spacing:.06em;color:#E8875F"
						>
							CYBERMONDAY
						</div>
						<div
							style="position:relative;margin-top:30px;font-size:140px;line-height:.9;letter-spacing:-.05em;font-weight:600"
						>
							Que el tráfico no te tire la tienda.
						</div>
						<div
							style="position:relative;margin-top:50px;font-size:44px;line-height:1.35;color:#BDB2A8"
						>
							Copia diaria de tus datos y aviso por mail de cada ingreso, en todos los&nbsp;planes.
						</div>
						<img
							src="/hornero/seguridad.png"
							alt=""
							style="position:absolute;left:50%;transform:translateX(-50%);bottom:250px;width:600px;height:600px;object-fit:contain"
						/>
					</div></Lamina
				>
			</div>

			<div id="3k" style="display:flex;flex-direction:column;gap:12px">
				<div style="display:flex;align-items:center;gap:10px">
					<a
						href="#3k"
						style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
						>3k</a
					><span style="font-size:14px;color:#6B5244">Navidad · feed 1:1</span>
				</div>
				<Lamina
					nombre="redes-3k-anuncio-meta"
					ancho={1080}
					alto={1080}
					escala={0.4}
					radio={6}
					sombra="0 18px 40px -20px rgba(42,20,12,.45)"
					><div
						style="width:1080px;height:1080px;background:#FBF7F2;position:relative;overflow:hidden;padding:90px"
					>
						<div
							style="position:absolute;right:-60px;top:-60px;width:520px;height:600px;border-radius:260px 260px 0 0;background:#2F6B4A"
						></div>
						<div style="position:relative;font-size:40px;font-weight:700;color:#2F6B4A">
							Navidad en Tu Local Web
						</div>
						<div
							style="position:relative;margin-top:30px;font-size:110px;line-height:.92;letter-spacing:-.05em;font-weight:600;max-width:700px"
						>
							Armá combos de regalo en 2 clics.
						</div>
						<div
							style="position:relative;margin-top:30px;font-size:34px;color:#6B5244;max-width:560px"
						>
							Con envoltorio, tarjeta y fecha de entrega antes del 24.
						</div>
						<img
							src="/hornero/ticket.png"
							alt=""
							style="position:absolute;right:50px;bottom:40px;width:400px;height:400px;object-fit:contain"
						/>
					</div></Lamina
				>
			</div>
		</div>
	</div>

	<!-- THUMBNAILS -->
	<div style="display:flex;flex-direction:column;gap:20px">
		<h2 style="margin:0;font-size:26px;letter-spacing:-.03em;font-weight:600">
			Miniaturas de YouTube · 1280 × 720 · al 40 %
		</h2>
		<div style="display:flex;flex-wrap:wrap;gap:32px">
			<div id="4a" style="display:flex;flex-direction:column;gap:12px">
				<div style="display:flex;align-items:center;gap:10px">
					<a
						href="#4a"
						style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
						>4a</a
					><span style="font-size:14px;color:#6B5244">Tutorial · cara + texto grande</span>
				</div>
				<Lamina
					nombre="redes-4a-miniatura-youtube"
					ancho={1280}
					alto={720}
					escala={0.4}
					radio={10}
					sombra="0 18px 40px -20px rgba(42,20,12,.5)"
					><div
						style="width:1280px;height:720px;background:#D2582F;position:relative;overflow:hidden;color:#fff"
					>
						<div
							style="position:absolute;inset:0;background:url(/textura/grain.png);opacity:.22;mix-blend-mode:multiply"
						></div>
						<div style="position:absolute;right:0;top:0;bottom:0;width:560px;overflow:hidden">
							<img
								src="https://images.unsplash.com/photo-1531058240690-006c446962d8?auto=format&fit=crop&w=700&h=900&q=75&crop=faces"
								alt=""
								style="width:100%;height:100%;object-fit:cover"
							/>
						</div>
						<div
							style="position:absolute;left:70px;top:90px;width:700px;font-size:120px;line-height:.9;letter-spacing:-.05em;font-weight:700;text-shadow:0 6px 0 #2A140C"
						>
							TU TIENDA EN 10 MIN
						</div>
						<div
							style="position:absolute;left:70px;bottom:70px;display:flex;align-items:center;gap:16px;padding:16px 28px;border-radius:999px;background:#2A140C;font-size:34px;font-weight:600"
						>
							<img src="/logo.svg" alt="" style="width:40px" />Paso a paso
						</div>
					</div></Lamina
				>
			</div>
			<div id="4b" style="display:flex;flex-direction:column;gap:12px">
				<div style="display:flex;align-items:center;gap:10px">
					<a
						href="#4b"
						style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
						>4b</a
					><span style="font-size:14px;color:#6B5244">Comparativa · versus</span>
				</div>
				<Lamina
					nombre="redes-4b-miniatura-youtube"
					ancho={1280}
					alto={720}
					escala={0.4}
					radio={10}
					sombra="0 18px 40px -20px rgba(42,20,12,.5)"
					><div
						style="width:1280px;height:720px;background:#2A140C;position:relative;overflow:hidden;color:#fff;display:grid;grid-template-columns:1fr auto 1fr;align-items:center;padding:0 80px;gap:40px"
					>
						<div style="text-align:center">
							<div style="font-size:60px;font-weight:600;color:#9C8676">Tiendanube</div>
							<div
								style="font-size:100px;font-weight:700;color:#9C8676;text-decoration:line-through"
							>
								{fmt(VS_PRO.competencia.precio)}
							</div>
						</div>
						<div
							style="width:150px;height:150px;border-radius:50%;background:#E8875F;color:#2A140C;display:grid;place-items:center;font-size:64px;font-weight:800"
						>
							VS
						</div>
						<div style="text-align:center">
							<img src="/logo.svg" alt="" style="width:110px" />
							<div style="font-size:120px;font-weight:700;color:#E8875F;letter-spacing:-.04em">
								{fmt(VS_PRO.plan.precio)}
							</div>
						</div>
						<img
							src="/hornero/pensando.png"
							alt=""
							style="position:absolute;left:50%;bottom:-40px;transform:translateX(-50%);width:260px;height:260px;object-fit:contain"
						/>
					</div></Lamina
				>
			</div>
			<div id="4c" style="display:flex;flex-direction:column;gap:12px">
				<div style="display:flex;align-items:center;gap:10px">
					<a
						href="#4c"
						style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
						>4c</a
					><span style="font-size:14px;color:#6B5244"
						>Caso real · número protagonista · el texto se edita con un&nbsp;clic</span
					>
				</div>
				<Lamina
					nombre="redes-4c-miniatura-youtube"
					ancho={1280}
					alto={720}
					escala={0.4}
					radio={10}
					sombra="0 18px 40px -20px rgba(42,20,12,.5)"
					><div
						style="width:1280px;height:720px;background:#FBF7F2;position:relative;overflow:hidden"
					>
						<div
							style="position:absolute;left:60px;top:60px;bottom:60px;width:470px;border-radius:235px 235px 24px 24px;overflow:hidden"
						>
							<img
								src="https://images.unsplash.com/photo-1495474472287-4d71bcdd2085?auto=format&fit=crop&w=600&h=800&q=75"
								alt=""
								style="width:100%;height:100%;object-fit:cover"
							/>
						</div>
						<div
							contenteditable="true"
							spellcheck="false"
							role="textbox"
							aria-multiline="true"
							aria-label="Nombre de la tienda"
							style="position:absolute;left:590px;top:110px"
						>
							<div style="font-size:50px;font-weight:600;color:#6B5244">Nombre de la tienda</div>
							<div
								style="font-size:220px;line-height:.9;letter-spacing:-.06em;font-weight:700;color:#D2582F"
							>
								×N
							</div>
							<div style="font-size:60px;line-height:1;letter-spacing:-.03em;font-weight:600">
								el resultado del caso
							</div>
						</div>
					</div></Lamina
				>
			</div>
		</div>
	</div>
</section>
