<script lang="ts">
	import Seo from '#lib/components/Seo.svelte';
	import DescargarTodo from '#lib/marca/DescargarTodo.svelte';
	import Lamina from '#lib/marca/Lamina.svelte';
	import { VERSUS, fmt } from '#lib/marca/datos.ts';
	import { crearFamilia } from '#lib/marca/familia.svelte.ts';

	/** Las 9 imágenes OG (design/OG Mostrador.dc.html), 1200 × 630. */
	const fam = crearFamilia('og');
</script>

<Seo
	title="Imágenes para compartir · Marca · Tu Local Web"
	description="Las imágenes de 1200 × 630 que se ven al compartir el sitio."
	path="/marca/og"
	noindex
/>

<section class="lienzo" bind:contentRect={fam.rect}>
	<div class="lienzo-cab">
		<div style="flex:1;min-width:min(100%,300px)">
			<div style="font-size:28px;font-weight:600;letter-spacing:-.03em">
				Imágenes para compartir (OG) · 1200&nbsp;×&nbsp;630
			</div>
			<div style="font-size:15px;color:#7A6252;margin-top:4px">
				Son las que usa el sitio (static/og/, en JPG): cada página enlaza la&nbsp;suya.
			</div>
		</div>
		<DescargarTodo />
	</div>
	<div style="display:flex;flex-wrap:wrap;gap:40px">
		<div id="1a" style="display:flex;flex-direction:column;gap:10px">
			<span
				style="align-self:flex-start;padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:13px;font-weight:700"
				>1a · Inicio</span
			>
			<Lamina nombre="og-home" ancho={1200} alto={630}
				><div
					data-og="home"
					style="width:1200px;height:630px;position:relative;overflow:hidden;background:#FBF7F2;flex:none"
				>
					<div
						style="position:absolute;inset:0;background:radial-gradient(700px 520px at 88% 30%,rgba(232,135,95,.35),transparent 70%)"
					></div>
					<div
						style="position:absolute;inset:0;background-image:url(/textura/grain.png);opacity:.5;mix-blend-mode:multiply"
					></div>
					<div
						style="position:absolute;left:72px;top:64px;width:640px;display:flex;flex-direction:column;height:502px"
					>
						<img
							src="/wordmark-brown.png"
							alt=""
							style="height:34px;width:auto;align-self:flex-start"
						/>
						<div
							style="margin-top:auto;font-size:68px;line-height:.95;letter-spacing:-.05em;font-weight:600"
						>
							Todo lo que tienen las otras. <span style="color:#D2582F">Por menos plata.</span>
						</div>
						<div style="margin-top:22px;font-size:25px;line-height:1.35;color:#5C4536">
							Tienda online, WhatsApp, Mercado Libre, local y factura de ARCA en un solo panel.
						</div>
						<div style="margin-top:28px;display:flex;gap:10px">
							<span
								style="width:52px;height:52px;border-radius:16px;background:#fff;display:grid;place-items:center;box-shadow:0 6px 16px -8px rgba(58,29,18,.35)"
								><span
									style="display:block;flex:none;width:28px;height:28px;background:url(/marca/iconos/logos-whatsapp-icon.png) center/contain no-repeat;"
								></span></span
							>
							<span
								style="width:52px;height:52px;border-radius:16px;background:#FFE600;display:grid;place-items:center;font-weight:800;font-size:15px;color:#2D3277"
								>ML</span
							>
							<span
								style="width:52px;height:52px;border-radius:16px;background:#fff;display:grid;place-items:center;box-shadow:0 6px 16px -8px rgba(58,29,18,.35)"
								><span
									style="display:block;flex:none;width:30px;height:30px;background:url(/marca/iconos/simple-icons-mercadopago.png) center/contain no-repeat;"
								></span></span
							>
							<span
								style="width:52px;height:52px;border-radius:16px;background:#fff;display:grid;place-items:center;box-shadow:0 6px 16px -8px rgba(58,29,18,.35)"
								><span
									style="display:block;flex:none;width:26px;height:26px;background:url(/marca/iconos/logos-instagram-icon.png) center/contain no-repeat;"
								></span></span
							>
							<span
								style="width:52px;height:52px;border-radius:16px;background:#FFCE00;display:grid;place-items:center;font-weight:800;font-size:12px;color:#003DA5"
								>CA</span
							>
							<span
								style="width:52px;height:52px;border-radius:16px;background:#1F3A8A;display:grid;place-items:center;font-weight:800;font-size:11px;color:#fff"
								>ARCA</span
							>
						</div>
					</div>
					<div
						style="position:absolute;right:70px;bottom:0;width:380px;height:480px;border-radius:190px 190px 0 0;background:linear-gradient(180deg,#E06A3E,#B9481F);overflow:hidden"
					>
						<div
							style="position:absolute;inset:0;background-image:url(/textura/arch-pattern-light.svg);opacity:.9"
						></div>
					</div>
					<img
						src="/hornero/saludo.png"
						alt=""
						style="position:absolute;right:80px;bottom:-6px;width:370px;height:auto"
					/>
					<div
						style="position:absolute;right:300px;top:64px;width:300px;padding:16px 18px;border-radius:20px;background:#fff;box-shadow:0 24px 50px -18px rgba(58,29,18,.45);display:flex;gap:12px;align-items:center"
					>
						<span
							style="width:44px;height:44px;border-radius:14px;background:#E9F8EE;display:grid;place-items:center;flex:none"
							><span
								style="display:block;flex:none;width:24px;height:24px;background:url(/marca/iconos/logos-whatsapp-icon.png) center/contain no-repeat;"
							></span></span
						>
						<div>
							<div style="font-size:17px;font-weight:700">Venta nueva · $ 44.600</div>
							<div style="font-size:14.5px;color:#7A6252;white-space:nowrap">
								Lucía · Termo 1 L · hace 1 min
							</div>
						</div>
					</div>
				</div></Lamina
			>
		</div>

		<div id="1b" style="display:flex;flex-direction:column;gap:10px">
			<span
				style="align-self:flex-start;padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:13px;font-weight:700"
				>1b · Precios</span
			>
			<Lamina nombre="og-precios" ancho={1200} alto={630}
				><div
					data-og="precios"
					style="width:1200px;height:630px;position:relative;overflow:hidden;background:#2A140C;color:#FBF7F2;flex:none"
				>
					<div
						style="position:absolute;inset:0;background-image:url(/textura/arch-pattern-dark.svg)"
					></div>
					<div
						style="position:absolute;inset:0;background:radial-gradient(640px 480px at 100% 0%,rgba(210,88,47,.45),transparent 70%)"
					></div>
					<div
						style="position:absolute;left:72px;top:64px;bottom:64px;width:480px;display:flex;flex-direction:column"
					>
						<img
							src="/wordmark-cream.png"
							alt=""
							style="height:32px;width:auto;align-self:flex-start"
						/>
						<div
							style="margin-top:auto;font-size:78px;line-height:.94;letter-spacing:-.055em;font-weight:600"
						>
							Más barato en cada escalón.
						</div>
						<div style="margin-top:20px;font-size:23px;line-height:1.4;color:#D9C8BA">
							Y con Tincho, factura de ARCA y punto de venta adentro.
						</div>
					</div>
					<div
						style="position:absolute;right:64px;top:64px;bottom:64px;width:560px;display:flex;flex-direction:column;justify-content:center;gap:14px"
					>
						{#each VERSUS as v, i (v.plan.id)}
							<div
								style="padding:22px 26px;border-radius:24px;{i === 0
									? 'background:#FBF7F2;color:#2A140C'
									: 'background:rgba(251,247,242,.07);border:1px solid rgba(251,247,242,.14)'};display:flex;align-items:center;gap:16px"
							>
								<div style="flex:1">
									<div style="font-size:16px;color:{i === 0 ? '#7A6252' : '#D9C8BA'}">
										{v.plan.nombre}
									</div>
									<div style="font-size:42px;font-weight:600;letter-spacing:-.04em">
										{fmt(v.plan.precio)}
									</div>
								</div>
								<div style="text-align:right">
									<div style="font-size:15px;color:{i === 0 ? '#7A6252' : '#B8A293'}">
										{v.competencia.nombre}
									</div>
									<div
										style="font-size:22px;color:{i === 0
											? '#7A6252'
											: '#B8A293'};text-decoration:line-through"
									>
										{fmt(v.competencia.precio)}
									</div>
								</div>
								<span
									style="padding:8px 12px;border-radius:999px;background:#D2582F;color:#fff;font-size:19px;font-weight:700;white-space:nowrap"
									>−{v.descuento} %</span
								>
							</div>
						{/each}
						<div style="font-size:13px;color:#B8A293">
							Por mes con IVA · precios publicados en octubre de 2026
						</div>
					</div>
				</div></Lamina
			>
		</div>

		<div id="1c" style="display:flex;flex-direction:column;gap:10px">
			<span
				style="align-self:flex-start;padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:13px;font-weight:700"
				>1c · Tincho</span
			>
			<Lamina nombre="og-hornero" ancho={1200} alto={630}
				><div
					data-og="hornero"
					style="width:1200px;height:630px;position:relative;overflow:hidden;background:linear-gradient(160deg,#E06A3E,#A63F1B);color:#fff;flex:none"
				>
					<div
						style="position:absolute;inset:0;background-image:url(/textura/arch-pattern-light.svg);opacity:.7"
					></div>
					<div
						style="position:absolute;left:72px;top:64px;bottom:64px;width:560px;display:flex;flex-direction:column"
					>
						<img
							src="/wordmark-cream.png"
							alt=""
							style="height:32px;width:auto;align-self:flex-start"
						/>
						<div
							style="margin-top:auto;display:inline-flex;align-self:flex-start;align-items:center;gap:8px;padding:8px 14px;border-radius:999px;background:rgba(255,255,255,.18);font-size:18px;font-weight:600"
						>
							<span
								style="display:block;flex:none;width:18px;height:18px;background:url(/marca/iconos/ph-sparkle-fill.png) center/contain no-repeat;"
							></span>Asistente con IA
						</div>
						<div
							style="margin-top:18px;font-size:76px;line-height:.94;letter-spacing:-.055em;font-weight:600"
						>
							Tincho atiende tu WhatsApp mientras dormís.
						</div>
						<div style="margin-top:18px;font-size:23px;line-height:1.4;color:#FFE3D6">
							Responde con tu stock real, cobra y te avisa lo importante.
						</div>
					</div>
					<div
						style="position:absolute;right:300px;top:96px;width:330px;display:flex;flex-direction:column;gap:10px"
					>
						<div
							style="align-self:flex-start;max-width:280px;padding:14px 16px;border-radius:20px 20px 20px 6px;background:#fff;color:#2A140C;font-size:18px;line-height:1.35;box-shadow:0 16px 36px -16px rgba(0,0,0,.45)"
						>
							¿Tenés el termo en negro?
						</div>
						<div
							style="align-self:flex-end;max-width:300px;padding:14px 16px;border-radius:20px 20px 6px 20px;background:#2A140C;color:#FBF7F2;font-size:18px;line-height:1.35;box-shadow:0 16px 36px -16px rgba(0,0,0,.5)"
						>
							¡Sí! Quedan 3. Te llega el jueves a Córdoba. ¿Te mando el link de pago?
						</div>
						<div
							style="align-self:flex-end;display:flex;align-items:center;gap:8px;padding:10px 14px;border-radius:16px;background:rgba(255,255,255,.95);color:#2A140C;font-size:16px;font-weight:600"
						>
							<span
								style="display:block;flex:none;width:20px;height:20px;background:url(/marca/iconos/simple-icons-mercadopago.png) center/contain no-repeat;"
							></span>Pago aprobado · $ 42.900
						</div>
					</div>
					<img
						src="/hornero/hablando.png"
						alt=""
						style="position:absolute;right:30px;bottom:-10px;width:330px;height:auto"
					/>
				</div></Lamina
			>
		</div>

		<div id="1d" style="display:flex;flex-direction:column;gap:10px">
			<span
				style="align-self:flex-start;padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:13px;font-weight:700"
				>1d · Migrar</span
			>
			<Lamina nombre="og-migrar" ancho={1200} alto={630}
				><div
					data-og="migrar"
					style="width:1200px;height:630px;position:relative;overflow:hidden;background:#F4ECE2;flex:none"
				>
					<div
						style="position:absolute;inset:0;background-image:url(/textura/arch-pattern-light.svg)"
					></div>
					<div
						style="position:absolute;left:72px;top:64px;bottom:64px;width:600px;display:flex;flex-direction:column"
					>
						<img
							src="/wordmark-brown.png"
							alt=""
							style="height:32px;width:auto;align-self:flex-start"
						/>
						<div
							style="margin-top:auto;font-size:82px;line-height:.92;letter-spacing:-.055em;font-weight:600"
						>
							Mudate a Tu Local Web <span style="color:#D2582F">en un día.</span>
						</div>
						<div style="margin-top:20px;font-size:23px;line-height:1.4;color:#5C4536">
							Productos, clientes y pedidos. Sin perder Google. Gratis.
						</div>
						<div style="margin-top:26px;display:flex;align-items:center;gap:10px">
							<span
								style="padding:12px 18px;border-radius:999px;background:#fff;font-size:18px;font-weight:600;box-shadow:0 6px 16px -8px rgba(58,29,18,.3)"
								>Tiendanube</span
							>
							<span
								style="padding:12px 18px;border-radius:999px;background:#fff;font-size:18px;font-weight:600;box-shadow:0 6px 16px -8px rgba(58,29,18,.3)"
								>Empretienda</span
							>
							<span
								style="padding:12px 18px;border-radius:999px;background:#fff;font-size:18px;font-weight:600;box-shadow:0 6px 16px -8px rgba(58,29,18,.3)"
								>Mercado Shops</span
							>
							<span
								style="display:block;flex:none;width:26px;height:26px;background:url(/marca/iconos/ph-arrow-right-bold.png) center/contain no-repeat;"
							></span>
							<span
								style="width:50px;height:50px;border-radius:16px;background:#2A140C;display:grid;place-items:center"
								><img src="/logo.svg" alt="" style="width:26px" /></span
							>
						</div>
					</div>
					<div
						style="position:absolute;right:90px;top:90px;width:420px;height:420px;border-radius:50%;background:radial-gradient(closest-side,rgba(210,88,47,.28),rgba(210,88,47,0))"
					></div>
					<img
						src="/hornero/caja.png"
						alt=""
						style="position:absolute;right:70px;bottom:40px;width:430px;height:auto"
					/>
				</div></Lamina
			>
		</div>

		<div id="1e" style="display:flex;flex-direction:column;gap:10px">
			<span
				style="align-self:flex-start;padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:13px;font-weight:700"
				>1e · Seguridad</span
			>
			<Lamina nombre="og-seguridad" ancho={1200} alto={630}
				><div
					data-og="seguridad"
					style="width:1200px;height:630px;position:relative;overflow:hidden;background:#1E0F09;color:#FBF7F2;flex:none"
				>
					<div
						style="position:absolute;inset:0;background-image:url(/textura/arch-pattern-dark.svg)"
					></div>
					<div
						style="position:absolute;right:-80px;top:-80px;width:720px;height:720px;border-radius:50%;background:radial-gradient(closest-side,rgba(47,143,99,.32),rgba(47,143,99,0))"
					></div>
					<div
						style="position:absolute;left:72px;top:64px;bottom:64px;width:620px;display:flex;flex-direction:column"
					>
						<img
							src="/wordmark-cream.png"
							alt=""
							style="height:32px;width:auto;align-self:flex-start"
						/>
						<div
							style="margin-top:auto;font-size:80px;line-height:.93;letter-spacing:-.055em;font-weight:600"
						>
							Tu negocio no se cae. Y nada se pierde.
						</div>
						<div style="margin-top:28px;display:flex;gap:12px">
							<div
								style="padding:16px 20px;border-radius:20px;background:rgba(251,247,242,.07);border:1px solid rgba(251,247,242,.12)"
							>
								<div style="font-size:34px;font-weight:600;letter-spacing:-.03em">99,98 %</div>
								<div style="font-size:15px;color:#D9C8BA;white-space:nowrap">
									en línea, 12 meses
								</div>
							</div>
							<div
								style="padding:16px 20px;border-radius:20px;background:rgba(251,247,242,.07);border:1px solid rgba(251,247,242,.12)"
							>
								<div style="font-size:34px;font-weight:600;letter-spacing:-.03em">30 días</div>
								<div style="font-size:15px;color:#D9C8BA;white-space:nowrap">
									de copias restaurables
								</div>
							</div>
							<div
								style="padding:16px 20px;border-radius:20px;background:rgba(251,247,242,.07);border:1px solid rgba(251,247,242,.12)"
							>
								<div style="font-size:34px;font-weight:600;letter-spacing:-.03em">2 pasos</div>
								<div style="font-size:15px;color:#D9C8BA;white-space:nowrap">incluido en todos</div>
							</div>
						</div>
					</div>
					<img
						src="/hornero/seguridad.png"
						alt=""
						style="position:absolute;right:70px;bottom:40px;width:400px;height:auto"
					/>
				</div></Lamina
			>
		</div>

		<div id="1f" style="display:flex;flex-direction:column;gap:10px">
			<span
				style="align-self:flex-start;padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:13px;font-weight:700"
				>1f · Plantillas</span
			>
			<Lamina nombre="og-plantillas" ancho={1200} alto={630}
				><div
					data-og="plantillas"
					style="width:1200px;height:630px;position:relative;overflow:hidden;background:#FBF7F2;flex:none"
				>
					<div
						style="position:absolute;inset:0;background-image:url(/textura/grain.png);opacity:.5;mix-blend-mode:multiply"
					></div>
					<div
						style="position:absolute;left:72px;top:64px;bottom:64px;width:470px;display:flex;flex-direction:column"
					>
						<img
							src="/wordmark-brown.png"
							alt=""
							style="height:32px;width:auto;align-self:flex-start"
						/>
						<div
							style="margin-top:auto;font-size:76px;line-height:.93;letter-spacing:-.055em;font-weight:600"
						>
							Una tienda que <span style="color:#D2582F">se ve como tu marca.</span>
						</div>
						<div style="margin-top:18px;font-size:22px;line-height:1.4;color:#5C4536">
							Plantillas pensadas para cada rubro: moda, cafés, almacenes, vinotecas, belleza y más.
						</div>
					</div>
					<div
						style="position:absolute;right:56px;top:40px;bottom:0;width:590px;display:grid;grid-template-columns:repeat(3,1fr);gap:14px"
					>
						<div style="display:flex;flex-direction:column;gap:14px;margin-top:70px">
							<div
								style="height:260px;border-radius:999px 999px 24px 24px;overflow:hidden;background:#E9DDCF"
							>
								<img
									src="https://images.unsplash.com/photo-1495474472287-4d71bcdd2085?auto=format&fit=crop&w=400&h=520&q=72"
									alt=""
									style="width:100%;height:100%;object-fit:cover"
								/>
							</div>
							<div style="height:240px;border-radius:24px;overflow:hidden;background:#E9DDCF">
								<img
									src="https://images.unsplash.com/photo-1609541994821-d909982e2f1b?auto=format&fit=crop&w=400&h=480&q=72"
									alt=""
									style="width:100%;height:100%;object-fit:cover"
								/>
							</div>
						</div>
						<div style="display:flex;flex-direction:column;gap:14px">
							<div style="height:300px;border-radius:24px;overflow:hidden;background:#E9DDCF">
								<img
									src="https://images.unsplash.com/photo-1555041469-a586c61ea9bc?auto=format&fit=crop&w=400&h=600&q=72"
									alt=""
									style="width:100%;height:100%;object-fit:cover"
								/>
							</div>
							<div
								style="height:280px;border-radius:999px 999px 24px 24px;overflow:hidden;background:#E9DDCF"
							>
								<img
									src="https://images.unsplash.com/photo-1634741011278-9dc2f5b6fc0b?auto=format&fit=crop&w=400&h=560&q=72"
									alt=""
									style="width:100%;height:100%;object-fit:cover"
								/>
							</div>
						</div>
						<div style="display:flex;flex-direction:column;gap:14px;margin-top:120px">
							<div
								style="height:230px;border-radius:999px 999px 24px 24px;overflow:hidden;background:#E9DDCF"
							>
								<img
									src="https://images.unsplash.com/photo-1567538096630-e0c55bd6374c?auto=format&fit=crop&w=400&h=460&q=72"
									alt=""
									style="width:100%;height:100%;object-fit:cover"
								/>
							</div>
							<div style="height:240px;border-radius:24px;overflow:hidden;background:#E9DDCF">
								<img
									src="https://images.unsplash.com/photo-1675001077188-809370ce3279?auto=format&fit=crop&w=400&h=480&q=72"
									alt=""
									style="width:100%;height:100%;object-fit:cover"
								/>
							</div>
						</div>
					</div>
				</div></Lamina
			>
		</div>

		<div id="1g" style="display:flex;flex-direction:column;gap:10px">
			<span
				style="align-self:flex-start;padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:13px;font-weight:700"
				>1g · Partners</span
			>
			<Lamina nombre="og-partners" ancho={1200} alto={630}
				><div
					data-og="partners"
					style="width:1200px;height:630px;position:relative;overflow:hidden;background:#2A140C;color:#FBF7F2;flex:none"
				>
					<div
						style="position:absolute;inset:0;background-image:url(/textura/arch-pattern-dark.svg)"
					></div>
					<div
						style="position:absolute;left:72px;top:64px;bottom:64px;width:580px;display:flex;flex-direction:column"
					>
						<img
							src="/wordmark-cream.png"
							alt=""
							style="height:32px;width:auto;align-self:flex-start"
						/>
						<div
							style="margin-top:auto;font-size:18px;font-weight:700;letter-spacing:.14em;text-transform:uppercase;color:#F5B796"
						>
							Programa de partners
						</div>
						<div
							style="margin-top:14px;font-size:80px;line-height:.93;letter-spacing:-.055em;font-weight:600"
						>
							Armás tiendas. Ganás cada mes.
						</div>
						<div style="margin-top:18px;font-size:22px;line-height:1.4;color:#D9C8BA">
							Para agencias, diseñadores, desarrolladores y contadores.
						</div>
					</div>
					<div
						style="position:absolute;right:72px;top:96px;width:420px;height:420px;border-radius:50%;background:#D2582F;display:flex;flex-direction:column;align-items:center;justify-content:center;box-shadow:0 40px 80px -30px rgba(0,0,0,.7)"
					>
						<div style="font-size:150px;font-weight:600;letter-spacing:-.07em;line-height:.9">
							30 %
						</div>
						<div style="font-size:22px;font-weight:600;margin-top:8px">comisión recurrente</div>
					</div>
					<img
						src="/hornero/celebrando.png"
						alt=""
						style="position:absolute;right:360px;bottom:-12px;width:230px;height:auto"
					/>
				</div></Lamina
			>
		</div>

		<div id="1h" style="display:flex;flex-direction:column;gap:10px">
			<span
				style="align-self:flex-start;padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:13px;font-weight:700"
				>1h · Tu Local Web + Caja</span
			>
			<Lamina nombre="og-caja" ancho={1200} alto={630}
				><div
					data-og="caja"
					style="width:1200px;height:630px;position:relative;overflow:hidden;background:#131211;color:#F4EEE6;flex:none"
				>
					<div
						style="position:absolute;left:-140px;bottom:-200px;width:700px;height:700px;border-radius:50%;background:radial-gradient(closest-side,rgba(210,88,47,.35),transparent)"
					></div>
					<div
						style="position:absolute;left:72px;top:64px;bottom:64px;width:620px;display:flex;flex-direction:column"
					>
						<div style="display:flex;align-items:center;gap:14px">
							<img src="/logo.svg" alt="" style="width:40px" /><span
								style="font-size:28px;color:#8A7F76">+</span
							><span
								style="display:flex;align-items:center;gap:10px;font-size:30px;font-weight:700;letter-spacing:-.03em"
								><img
									src="/caja/01-saludo.png"
									alt=""
									style="width:52px;height:52px;object-fit:contain"
								/>Caja</span
							>
						</div>
						<div
							style="margin-top:auto;font-size:80px;line-height:.93;letter-spacing:-.055em;font-weight:600"
						>
							Tu Local Web vende. <span style="color:#E8875F">Caja ordena la plata.</span>
						</div>
						<div style="margin-top:18px;font-size:22px;line-height:1.4;color:#BDB2A8">
							Cada venta, comisión y factura pasa sola a tus finanzas.
						</div>
					</div>
					<div
						style="position:absolute;right:72px;top:110px;width:380px;padding:24px;border-radius:28px;background:#F4EEE6;color:#131211;box-shadow:0 40px 80px -30px rgba(0,0,0,.8)"
					>
						<div style="display:flex;align-items:center;gap:10px">
							<span
								style="width:48px;height:48px;border-radius:14px;background:#131211;color:#F4EEE6;display:grid;place-items:center"
								><span
									style="display:block;flex:none;width:24px;height:24px;background:url(/marca/iconos/ph-wallet-fill.png) center/contain no-repeat;"
								></span></span
							>
							<div>
								<div style="font-size:18px;font-weight:700">Caja · Negocio</div>
								<div style="font-size:14px;color:#6E655D">octubre</div>
							</div>
						</div>
						<div style="margin-top:18px;font-size:16px;color:#6E655D">Ganancia del mes</div>
						<div style="font-size:48px;font-weight:600;letter-spacing:-.04em">$ 6.812.400</div>
						<div style="font-size:16px;color:#6E655D">≈ US$ 4.412 al MEP</div>
						<div
							style="margin-top:18px;display:flex;justify-content:space-between;font-size:15px;color:#6E655D"
						>
							<span>Monotributo · Cat. D</span><b style="color:#131211">70 %</b>
						</div>
						<div
							style="margin-top:8px;height:10px;border-radius:5px;background:#E2D8CD;overflow:hidden"
						>
							<div style="width:70%;height:100%;background:#D2582F"></div>
						</div>
					</div>
				</div></Lamina
			>
		</div>

		<div id="1i" style="display:flex;flex-direction:column;gap:10px">
			<span
				style="align-self:flex-start;padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:13px;font-weight:700"
				>1i · Ayuda y legales</span
			>
			<Lamina nombre="og-general" ancho={1200} alto={630}
				><div
					data-og="general"
					style="width:1200px;height:630px;position:relative;overflow:hidden;background:#F4ECE2;flex:none"
				>
					<div
						style="position:absolute;inset:0;background-image:url(/textura/arch-pattern-light.svg)"
					></div>
					<div
						style="position:absolute;left:72px;top:64px;bottom:64px;width:640px;display:flex;flex-direction:column"
					>
						<img
							src="/wordmark-brown.png"
							alt=""
							style="height:32px;width:auto;align-self:flex-start"
						/>
						<div
							style="margin-top:auto;font-size:80px;line-height:.93;letter-spacing:-.055em;font-weight:600"
						>
							Claro, simple y <span style="color:#D2582F">sin letra chica.</span>
						</div>
						<div style="margin-top:18px;font-size:23px;line-height:1.4;color:#5C4536">
							Ayuda, legales y todo lo que necesitás saber para vender tranquilo.
						</div>
					</div>
					<div
						style="position:absolute;right:80px;bottom:0;width:360px;height:440px;border-radius:180px 180px 0 0;background:#fff;box-shadow:0 30px 60px -30px rgba(58,29,18,.4)"
					></div>
					<img
						src="/hornero/portapapeles.png"
						alt=""
						style="position:absolute;right:70px;bottom:20px;width:380px;height:auto"
					/>
				</div></Lamina
			>
		</div>
	</div>
</section>

<style>
	/* El handoff de las OG no fija border-box (las otras familias sí): el padding suma al ancho. */
	.lienzo :global([data-og]),
	.lienzo :global([data-og] *) {
		box-sizing: content-box;
	}
</style>
