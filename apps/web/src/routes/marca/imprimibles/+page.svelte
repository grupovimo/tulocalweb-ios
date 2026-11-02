<script lang="ts">
	import { page } from '$app/state';
	import Seo from '#lib/components/Seo.svelte';
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import Campos from '#lib/marca/Campos.svelte';
	import DescargarTodo from '#lib/marca/DescargarTodo.svelte';
	import Lamina from '#lib/marca/Lamina.svelte';
	import Qr from '#lib/marca/Qr.svelte';
	import {
		AHORRO_MAXIMO,
		DESCUENTO_ANUAL,
		DOMINIO,
		EVENTO_INICIAL,
		MAIL,
		PRUEBA_DIAS,
		qrCon,
		VERSUS,
		VS_PRO,
		fmt
	} from '#lib/marca/datos.ts';
	import { crearFamilia } from '#lib/marca/familia.svelte.ts';
	import { PLANTILLAS } from '#lib/plantillas.ts';

	/** Imprimibles (design/Imprimibles Mostrador.dc.html): A6 a A3 en mm; roll-up y backing a escala. */
	const fam = crearFamilia('imprimibles');
	// Con ?ref= (Materiales del panel de partners) los QR llevan al link del partner.
	const QR = $derived(qrCon(page.url.searchParams.get('ref')));

	let evento = $state({ ...EVENTO_INICIAL });
	const CAMPOS_EVENTO: [string, string][] = [
		['titulo', 'Evento'],
		['ciudad', 'Ciudad'],
		['fecha', 'Fecha'],
		['lugar', 'Lugar'],
		['bajada', 'Bajada'],
		['condiciones', 'Condiciones'],
		['link', 'Link del QR']
	];
	/** La tarjeta para paquetes es un modelo con la tienda de ejemplo (Café Bruma). */
	const RESENA_MUESTRA = 'https://cafebruma.com.ar/resena/10480';
</script>

<Seo
	title="Imprimibles · Marca · Tu Local Web"
	description="Flyers, afiche, tríptico, one-pager, roll-up, backing de stand y tarjeta para paquetes."
	path="/marca/imprimibles"
	noindex
/>

<section class="lienzo" bind:contentRect={fam.rect}>
	<header class="lienzo-cab">
		<div style="flex:1;min-width:min(100%,300px)">
			<img src="/wordmark-brown.png" alt="Tu Local Web" style="height:26px;display:block" />
			<h1
				style="margin:22px 0 0;font-size:56px;line-height:1;letter-spacing:-.04em;font-weight:600"
			>
				Imprimibles
			</h1>
			<p style="margin:12px 0 0;font-size:18px;color:#6B5244;max-width:680px;text-wrap:pretty">
				Flyers, afiche, tríptico, one-pager, roll-up, backing de stand y la tarjeta que las tiendas
				meten en sus paquetes. Las piezas A4 y A5 están en milímetros reales; las grandes, a escala
				(el PDF sale a la medida&nbsp;real).
			</p>
		</div>
		<DescargarTodo />
	</header>

	<!-- FLYERS A5 -->
	<div style="display:flex;flex-direction:column;gap:20px">
		<h2 style="margin:0;font-size:26px;letter-spacing:-.03em;font-weight:600">
			Flyer promo · A5 148 × 210 mm
		</h2>
		<div style="display:flex;flex-wrap:wrap;gap:48px;align-items:flex-start">
			<div id="1a" style="display:flex;flex-direction:column;gap:12px">
				<div style="display:flex;align-items:center;gap:10px">
					<a
						href="#1a"
						style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
						>1a</a
					><span style="font-size:14px;color:#6B5244">Institucional · foto en arco</span>
				</div>
				<Lamina nombre="imprimibles-1a-flyer-a5" ancho={559.37} alto={793.7} mm={[148, 210]}
					><div
						style="width:148mm;height:210mm;background:#FBF7F2;position:relative;overflow:hidden;padding:12mm;display:flex;flex-direction:column;box-shadow:0 24px 50px -26px rgba(42,20,12,.45)"
					>
						<img
							src="/wordmark-brown.png"
							alt="Tu Local Web"
							style="height:6mm;align-self:flex-start"
						/>
						<div
							style="margin-top:8mm;font-size:38px;line-height:.95;letter-spacing:-.045em;font-weight:600"
						>
							Tu tienda online, <span style="color:#D2582F">lista hoy.</span>
						</div>
						<div style="margin-top:4mm;font-size:12px;line-height:1.5;color:#6B5244">
							Web, Mercado Libre, WhatsApp, facturación y envíos en una sola cuenta.
						</div>
						<div
							style="margin-top:7mm;flex:1;border-radius:60mm 60mm 4mm 4mm;overflow:hidden;position:relative"
						>
							<img
								src="https://images.unsplash.com/photo-1556745753-b2904692b3cd?auto=format&fit=crop&w=800&q=75"
								alt=""
								style="width:100%;height:100%;object-fit:cover"
							/>
							<div
								style="position:absolute;left:5mm;bottom:5mm;padding:3mm 5mm;border-radius:999px;background:#fff;font-size:11px;font-weight:600;display:flex;align-items:center;gap:6px"
							>
								<Icon icon={I.phCheckCircleFill} size={14} style="color:#2F8F63" />Nueva venta · $
								42.900
							</div>
						</div>
						<div style="margin-top:7mm;display:flex;align-items:center;gap:5mm">
							<div style="flex:1">
								<div style="font-size:16px;font-weight:600">{PRUEBA_DIAS} días gratis</div>
								<div style="font-size:10px;color:#7A6252">
									Escaneá y empezá en 10 minutos · {DOMINIO}
								</div>
							</div>
							<Qr
								texto={QR.flyer}
								alt="QR · Crear mi tienda"
								style="width:22mm;height:22mm;border-radius:2mm;display:block;flex:none;background:#FBF7F2"
							/>
						</div>
					</div></Lamina
				>
			</div>
			<div id="1b" style="display:flex;flex-direction:column;gap:12px">
				<div style="display:flex;align-items:center;gap:10px">
					<a
						href="#1b"
						style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
						>1b</a
					><span style="font-size:14px;color:#6B5244">Jugada · terracota con recortable</span>
				</div>
				<Lamina nombre="imprimibles-1b-flyer-a5" ancho={559.37} alto={793.7} mm={[148, 210]}
					><div
						style="width:148mm;height:210mm;background:#D2582F;color:#FBF7F2;position:relative;overflow:hidden;padding:12mm;display:flex;flex-direction:column;box-shadow:0 24px 50px -26px rgba(42,20,12,.55)"
					>
						<div
							style="position:absolute;inset:0;background:url(/textura/arch-pattern-accent.svg) center/22mm;opacity:.55"
						></div>
						<div
							style="position:absolute;inset:0;background:url(/textura/grain.png);opacity:.22;mix-blend-mode:multiply"
						></div>
						<div style="position:relative;font-size:12px;font-weight:700;letter-spacing:.12em">
							¿VENDÉS POR INSTAGRAM?
						</div>
						<div
							style="position:relative;margin-top:4mm;font-size:40px;line-height:.92;letter-spacing:-.05em;font-weight:600"
						>
							Dejá de pasar el CBU por DM.
						</div>
						<img
							src="/hornero/whatsapp.png"
							alt=""
							style="position:relative;width:70mm;height:70mm;object-fit:contain;align-self:center;margin-top:4mm"
						/>
						<div
							style="position:relative;margin-top:auto;padding:6mm;border-radius:4mm;background:#FBF7F2;color:#2A140C;display:flex;align-items:center;gap:5mm;outline:1.5px dashed #FBF7F2;outline-offset:2.5mm"
						>
							<div style="flex:1">
								<div style="font-size:9px;font-weight:700;letter-spacing:.12em;color:#C24E25">
									PRUEBA GRATIS
								</div>
								<div style="font-size:24px;font-weight:700;letter-spacing:-.02em">
									{PRUEBA_DIAS} días
								</div>
								<div style="font-size:9.5px;color:#7A6252">
									Sin tarjeta y sin comisión por venta
								</div>
							</div>
							<Qr
								texto={QR.flyerJugada}
								alt="QR · Crear mi tienda"
								style="width:20mm;height:20mm;border-radius:2mm;display:block;flex:none;background:#FBF7F2"
							/>
						</div>
					</div></Lamina
				>
			</div>
		</div>
	</div>

	<!-- AFICHE A3 -->
	<div id="2" style="display:flex;flex-direction:column;gap:20px">
		<div style="display:flex;align-items:center;gap:12px">
			<a
				href="#2"
				style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
				>2</a
			>
			<h2 style="margin:0;font-size:26px;letter-spacing:-.03em;font-weight:600">
				Afiche para eventos · A3 297 × 420 mm · al 60 %
			</h2>
		</div>
		<Campos
			titulo="El evento"
			ayuda="Lo que escribas acá va en el&nbsp;afiche."
			campos={CAMPOS_EVENTO}
			bind:valores={evento}
		/>
		<Lamina
			nombre="imprimibles-2-afiche-a3"
			ancho={1122.52}
			alto={1587.4}
			mm={[297, 420]}
			escala={0.6}
			sombra="0 30px 60px -30px rgba(42,20,12,.55)"
			><div
				style="width:297mm;height:420mm;background:#2A140C;color:#FBF7F2;position:relative;overflow:hidden;padding:22mm"
			>
				<div style="position:absolute;inset:0;background:url(/textura/grain.png);opacity:.16"></div>
				<div
					style="position:absolute;left:50%;top:110mm;transform:translateX(-50%);width:230mm;height:250mm;border-radius:115mm 115mm 0 0;background:#D2582F;overflow:hidden"
				>
					<div
						style="position:absolute;inset:0;background:url(/textura/arch-pattern-accent.svg) center/30mm;opacity:.6"
					></div>
					<img
						src="/hornero/celebrando.png"
						alt=""
						style="position:absolute;left:50%;bottom:-6mm;transform:translateX(-50%);width:200mm;height:200mm;object-fit:contain"
					/>
				</div>
				<div
					style="position:relative;display:flex;justify-content:space-between;align-items:center"
				>
					<img src="/wordmark-cream.png" alt="Tu Local Web" style="height:11mm" /><span
						style="font-size:20px;font-weight:600;color:#E8875F"
						>{evento.ciudad} · {evento.fecha}</span
					>
				</div>
				<div
					style="position:relative;margin-top:16mm;font-size:110px;line-height:.86;letter-spacing:-.06em;font-weight:600"
				>
					{evento.titulo}
				</div>
				<div style="position:relative;margin-top:6mm;font-size:24px;color:#BDB2A8;max-width:180mm">
					{evento.bajada}
				</div>
				<div
					style="position:absolute;left:22mm;right:22mm;bottom:22mm;display:flex;justify-content:space-between;align-items:end;gap:10mm"
				>
					<div style="font-size:18px;line-height:1.5">
						{evento.lugar}<br /><b style="font-size:24px">{evento.condiciones}</b>
					</div>
					<Qr
						texto={evento.link}
						alt="QR · {evento.titulo}"
						style="width:40mm;height:40mm;border-radius:3mm;display:block;flex:none;background:#FBF7F2"
					/>
				</div>
			</div></Lamina
		>
	</div>

	<!-- TRIPTICO -->
	<div id="3" style="display:flex;flex-direction:column;gap:20px">
		<div style="display:flex;align-items:center;gap:12px">
			<a
				href="#3"
				style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
				>3</a
			>
			<h2 style="margin:0;font-size:26px;letter-spacing:-.03em;font-weight:600">
				Folleto tríptico · A4 apaisado, 3 cuerpos de 99 mm
			</h2>
		</div>
		<span style="font-size:14px;color:#6B5244">Exterior · solapa interna | contratapa | tapa</span>
		<Lamina nombre="imprimibles-3-triptico-exterior" ancho={1122.52} alto={793.7} mm={[297, 210]}
			><div
				style="width:297mm;height:210mm;display:grid;grid-template-columns:97mm 100mm 100mm;box-shadow:0 30px 60px -30px rgba(42,20,12,.5)"
			>
				<div
					style="background:#FCEDE5;padding:10mm;display:flex;flex-direction:column;border-right:1px dashed #E3C9B8"
				>
					<div style="font-size:9px;font-weight:700;letter-spacing:.12em;color:#C24E25">
						PASATE SIN COSTO
					</div>
					<div
						style="margin-top:3mm;font-size:22px;line-height:1;letter-spacing:-.03em;font-weight:600"
					>
						Migramos tu tienda en un&nbsp;día.
					</div>
					<div
						style="margin-top:6mm;display:flex;flex-direction:column;gap:4mm;font-size:10.5px;line-height:1.4"
					>
						<span style="display:flex;gap:6px"
							><b style="color:#D2582F">1</b>Pegás el link de tu tienda actual.</span
						><span style="display:flex;gap:6px"
							><b style="color:#D2582F">2</b>Copiamos productos, clientes, reseñas y dominio.</span
						><span style="display:flex;gap:6px"
							><b style="color:#D2582F">3</b>Revisás y publicás. Sin cortes de venta.</span
						>
					</div>
					<div style="margin-top:auto;display:flex;gap:8px;align-items:center">
						<span style="font-size:9px;color:#7A6252">Desde</span><b style="font-size:11px"
							>Tiendanube · Empretienda · Mercado Shops</b
						>
					</div>
				</div>
				<div
					style="background:#2A140C;color:#FBF7F2;padding:10mm;display:flex;flex-direction:column;align-items:center;justify-content:center;text-align:center;border-right:1px dashed #4A3428"
				>
					<img src="/hornero/saludo.png" alt="" style="width:46mm;height:46mm;object-fit:contain" />
					<div style="margin-top:4mm;font-size:16px;font-weight:600">¿Charlamos?</div>
					<div style="margin-top:3mm;font-size:10.5px;line-height:1.6;color:#BDB2A8">
						{MAIL.hola}<br />{DOMINIO}
					</div>
					<Qr
						texto={QR.triptico}
						alt="QR · Contacto"
						style="margin-top:6mm;width:24mm;height:24mm;border-radius:2mm;display:block;flex:none;background:#FBF7F2"
					/>
				</div>
				<div
					style="background:#D2582F;color:#FBF7F2;position:relative;overflow:hidden;padding:10mm;display:flex;flex-direction:column"
				>
					<div
						style="position:absolute;inset:0;background:url(/textura/arch-pattern-accent.svg) center/18mm;opacity:.55"
					></div>
					<img
						src="/wordmark-cream.png"
						alt="Tu Local Web"
						style="position:relative;height:5.5mm;align-self:flex-start"
					/>
					<div
						style="position:relative;margin-top:auto;width:60mm;height:80mm;border-radius:30mm 30mm 2mm 2mm;overflow:hidden;align-self:center"
					>
						<img
							src="https://images.unsplash.com/photo-1556745753-b2904692b3cd?auto=format&fit=crop&w=500&q=75"
							alt=""
							style="width:100%;height:100%;object-fit:cover"
						/>
					</div>
					<div
						style="position:relative;margin-top:6mm;font-size:24px;line-height:.95;letter-spacing:-.04em;font-weight:600"
					>
						Para los que venden en serio.
					</div>
				</div>
			</div></Lamina
		>
		<span style="font-size:14px;color:#6B5244"
			>Interior · qué incluye | precios | Tincho y Caja</span
		>
		<Lamina nombre="imprimibles-3-triptico-interior" ancho={1122.52} alto={793.7} mm={[297, 210]}
			><div
				style="width:297mm;height:210mm;display:grid;grid-template-columns:100mm 100mm 97mm;background:#FBF7F2;box-shadow:0 30px 60px -30px rgba(42,20,12,.45)"
			>
				<div style="padding:10mm;border-right:1px dashed #E3C9B8">
					<div style="font-size:9px;font-weight:700;letter-spacing:.12em;color:#C24E25">
						TODO EN UNA CUENTA
					</div>
					<div
						style="margin-top:3mm;font-size:20px;line-height:1.05;letter-spacing:-.03em;font-weight:600"
					>
						Lo que tienen las otras, y más.
					</div>
					<div style="margin-top:6mm;display:flex;flex-direction:column;gap:3.4mm;font-size:10.5px">
						<span style="display:flex;align-items:center;gap:8px"
							><Icon icon={I.phStorefrontBold} size={15} style="color:#D2582F" />Tienda online con {PLANTILLAS.length}
							plantillas</span
						>
						<span style="display:flex;align-items:center;gap:8px"
							><Icon icon={I.phCashRegisterBold} size={15} style="color:#D2582F" />Punto de venta
							para tu local</span
						>
						<span style="display:flex;align-items:center;gap:8px"
							><Icon icon={I.phHandshakeBold} size={15} style="color:#D2582F" />Mercado Libre
							sincronizado</span
						>
						<span style="display:flex;align-items:center;gap:8px"
							><Icon icon={I.phWhatsappLogoBold} size={15} style="color:#D2582F" />Ventas por
							WhatsApp e Instagram</span
						>
						<span style="display:flex;align-items:center;gap:8px"
							><Icon icon={I.phReceiptBold} size={15} style="color:#D2582F" />Factura electrónica
							con ARCA</span
						>
						<span style="display:flex;align-items:center;gap:8px"
							><Icon icon={I.phTruckBold} size={15} style="color:#D2582F" />Andreani, OCA, Correo y
							motos</span
						>
						<span style="display:flex;align-items:center;gap:8px"
							><Icon icon={I.phCreditCardBold} size={15} style="color:#D2582F" />Mercado Pago, MODO,
							Payway, cuotas</span
						>
						<span style="display:flex;align-items:center;gap:8px"
							><Icon icon={I.phShieldCheckBold} size={15} style="color:#D2582F" />Antifraude y
							respaldo diario</span
						>
					</div>
				</div>
				<div
					style="padding:10mm;border-right:1px dashed #E3C9B8;display:flex;flex-direction:column;gap:3.5mm"
				>
					<div style="font-size:9px;font-weight:700;letter-spacing:.12em;color:#C24E25">
						PLANES · SIN COMISIÓN POR VENTA
					</div>
					{#each VERSUS as v (v.plan.id)}
						{#if v.plan.id === 'pro'}
							<div
								style="padding:4mm 5mm;border-radius:3mm;background:#2A140C;color:#fff;display:flex;justify-content:space-between;align-items:center"
							>
								<div>
									<b style="font-size:13px">{v.plan.nombre}</b>
									<span style="font-size:8px;padding:1px 6px;border-radius:999px;background:#D2582F"
										>EL MÁS ELEGIDO</span
									>
									<div style="font-size:9.5px;color:#BDB2A8">{v.plan.incluye[0]}</div>
								</div>
								<b style="font-size:16px;color:#E8875F">{fmt(v.plan.precio)}</b>
							</div>
						{:else}
							<div
								style="padding:4mm 5mm;border-radius:3mm;background:#fff;display:flex;justify-content:space-between;align-items:center"
							>
								<div>
									<b style="font-size:13px">{v.plan.nombre}</b>
									<div style="font-size:9.5px;color:#7A6252">{v.plan.lema}</div>
								</div>
								<b style="font-size:16px">{fmt(v.plan.precio)}</b>
							</div>
						{/if}
					{/each}
					<div style="font-size:9px;color:#9C8676">
						Precios finales por mes, con IVA. Pagando por año, {DESCUENTO_ANUAL} %&nbsp;menos.
					</div>
					<div
						style="margin-top:auto;padding:4mm 5mm;border-radius:3mm;background:#FCEDE5;font-size:10.5px;line-height:1.4"
					>
						<b>Ahorrás hasta {fmt(AHORRO_MAXIMO)} por año</b> frente al plan equivalente de&nbsp;Tiendanube.
					</div>
				</div>
				<div style="padding:10mm;display:flex;flex-direction:column;gap:5mm">
					<div
						style="padding:5mm;border-radius:4mm;background:linear-gradient(180deg,#FCEADF,#F7D9C6);position:relative;overflow:hidden;min-height:62mm"
					>
						<div style="font-size:9px;font-weight:700;letter-spacing:.12em;color:#C24E25">
							HORNERO · IA
						</div>
						<div
							style="margin-top:2mm;font-size:15px;line-height:1.1;font-weight:600;max-width:46mm"
						>
							Responde, cobra y factura por vos, las 24 h.
						</div>
						<img
							src="/hornero/hablando.png"
							alt=""
							style="position:absolute;right:-3mm;bottom:-3mm;width:34mm;height:34mm;object-fit:contain"
						/>
					</div>
					<div style="padding:5mm;border-radius:4mm;background:#131211;color:#F4EEE6;flex:1">
						<div style="font-size:9px;font-weight:700;letter-spacing:.12em;color:#E8875F">
							+ CAJA
						</div>
						<div style="margin-top:2mm;font-size:15px;line-height:1.1;font-weight:600">
							Tus finanzas, ordenadas solas.
						</div>
						<div style="margin-top:3mm;font-size:10px;line-height:1.45;color:#BDB2A8">
							Ventas, comisiones y monotributo al día. Manejala también desde tu IA favorita.
						</div>
					</div>
				</div>
			</div></Lamina
		>
	</div>

	<!-- ONE PAGER -->
	<div id="4" style="display:flex;flex-direction:column;gap:20px">
		<div style="display:flex;align-items:center;gap:12px">
			<a
				href="#4"
				style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
				>4</a
			>
			<h2 style="margin:0;font-size:26px;letter-spacing:-.03em;font-weight:600">
				One-pager comercial · A4
			</h2>
		</div>
		<Lamina nombre="imprimibles-4-one-pager" ancho={793.7} alto={1122.52} mm={[210, 297]}
			><div
				style="width:210mm;height:297mm;background:#FBF7F2;position:relative;overflow:hidden;display:flex;flex-direction:column;box-shadow:0 30px 60px -30px rgba(42,20,12,.45)"
			>
				<div
					style="background:#2A140C;color:#FBF7F2;padding:14mm 16mm 12mm;position:relative;overflow:hidden"
				>
					<div
						style="position:absolute;inset:0;background:url(/textura/grain.png);opacity:.15"
					></div>
					<div
						style="position:relative;display:flex;justify-content:space-between;align-items:center"
					>
						<img src="/wordmark-cream.png" alt="Tu Local Web" style="height:6mm" /><span
							style="font-size:10px;color:#BDB2A8">Propuesta para tiendas · 2026</span
						>
					</div>
					<div
						style="position:relative;margin-top:9mm;font-size:34px;line-height:.95;letter-spacing:-.045em;font-weight:600;max-width:130mm"
					>
						Todo lo que tienen las otras. <span style="color:#E8875F"
							>Por menos plata y más seguro.</span
						>
					</div>
					<img
						src="/hornero/grafico.png"
						alt=""
						style="position:absolute;right:8mm;bottom:-4mm;width:46mm;height:46mm;object-fit:contain"
					/>
				</div>
				<div style="padding:9mm 16mm 0;display:grid;grid-template-columns:repeat(3,1fr);gap:5mm">
					<div style="padding:5mm;border-radius:3mm;background:#fff">
						<div style="font-size:24px;font-weight:600;color:#D2582F;letter-spacing:-.03em">
							−{VS_PRO.descuento} %
						</div>
						<div style="font-size:10px;color:#6B5244;margin-top:1mm">
							el plan {VS_PRO.plan.nombre} contra {VS_PRO.competencia.nombre}
						</div>
					</div>
					<div style="padding:5mm;border-radius:3mm;background:#fff">
						<div style="font-size:24px;font-weight:600;color:#D2582F;letter-spacing:-.03em">
							0 %
						</div>
						<div style="font-size:10px;color:#6B5244;margin-top:1mm">
							comisión por venta, en todos los planes
						</div>
					</div>
					<div style="padding:5mm;border-radius:3mm;background:#fff">
						<div style="font-size:24px;font-weight:600;color:#D2582F;letter-spacing:-.03em">
							1 día
						</div>
						<div style="font-size:10px;color:#6B5244;margin-top:1mm">
							para mudar tu tienda, gratis
						</div>
					</div>
				</div>
				<div style="padding:8mm 16mm 0;display:grid;grid-template-columns:1fr 1fr;gap:8mm">
					<div>
						<div style="font-size:12px;font-weight:700">Vendé en todos lados</div>
						<div style="margin-top:2mm;font-size:10px;line-height:1.55;color:#4A3428">
							Tienda online, Mercado Libre, Instagram, WhatsApp, Telegram y tu local con el mismo
							stock.
						</div>
					</div>
					<div>
						<div style="font-size:12px;font-weight:700">Cobrá como quieran</div>
						<div style="margin-top:2mm;font-size:10px;line-height:1.55;color:#4A3428">
							Mercado Pago, MODO, Payway, Ualá, transferencia con descuento y hasta 12 cuotas.
						</div>
					</div>
					<div>
						<div style="font-size:12px;font-weight:700">Tincho trabaja por vos</div>
						<div style="margin-top:2mm;font-size:10px;line-height:1.55;color:#4A3428">
							Responde consultas, recupera carritos y te avisa qué reponer antes de quedarte sin
							stock.
						</div>
					</div>
					<div>
						<div style="font-size:12px;font-weight:700">Seguro de verdad</div>
						<div style="margin-top:2mm;font-size:10px;line-height:1.55;color:#4A3428">
							Antifraude, doble verificación, respaldo diario y datos alojados con cifrado.
						</div>
					</div>
				</div>
				<div
					style="margin:8mm 16mm 0;border-radius:4mm;overflow:hidden;height:62mm;position:relative"
				>
					<img
						src="https://images.unsplash.com/photo-1556745753-b2904692b3cd?auto=format&fit=crop&w=1100&q=75"
						alt=""
						style="width:100%;height:100%;object-fit:cover"
					/>
					<div
						style="position:absolute;left:5mm;bottom:5mm;right:40%;padding:4mm 5mm;border-radius:3mm;background:rgba(251,247,242,.95);font-size:10px;line-height:1.45"
					>
						Tienda online, Mercado Libre, WhatsApp y el local con el mismo&nbsp;stock.<br /><b
							>Hecho en Argentina</b
						>
					</div>
				</div>
				<div
					style="margin-top:auto;padding:8mm 16mm;display:flex;align-items:center;gap:6mm;border-top:1px solid #E6D9CC"
				>
					<div style="flex:1">
						<div style="font-size:15px;font-weight:600">Empezá con {PRUEBA_DIAS} días gratis</div>
						<div style="font-size:10px;color:#7A6252">{DOMINIO} · {MAIL.hola}</div>
					</div>
					<Qr
						texto={QR.onepager}
						alt="QR · Crear mi tienda"
						style="width:22mm;height:22mm;border-radius:2mm;display:block;flex:none;background:#FBF7F2"
					/>
				</div>
			</div></Lamina
		>
	</div>

	<!-- ROLL-UP + BACKING -->
	<div style="display:flex;flex-direction:column;gap:20px">
		<h2 style="margin:0;font-size:26px;letter-spacing:-.03em;font-weight:600">
			Stand · roll-up y backing
		</h2>
		<div style="display:flex;flex-wrap:wrap;gap:48px;align-items:flex-start">
			<div id="5" style="display:flex;flex-direction:column;gap:12px">
				<div style="display:flex;align-items:center;gap:10px">
					<a
						href="#5"
						style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
						>5</a
					><span style="font-size:14px;color:#6B5244">Roll-up · 85 × 200 cm · escala 1:5</span>
				</div>
				<Lamina
					nombre="imprimibles-5-rollup"
					ancho={642.52}
					alto={1511.81}
					mm={[170, 400]}
					impresion={5}
					><div
						style="width:170mm;height:400mm;background:#FBF7F2;position:relative;overflow:hidden;display:flex;flex-direction:column;align-items:center;padding:16mm 12mm;text-align:center;box-shadow:0 30px 60px -30px rgba(42,20,12,.5)"
					>
						<img src="/wordmark-brown.png" alt="Tu Local Web" style="height:9mm" />
						<div
							style="margin-top:14mm;font-size:46px;line-height:.92;letter-spacing:-.05em;font-weight:600"
						>
							Vendé más.<br /><span style="color:#D2582F">Pagá menos.</span>
						</div>
						<div
							style="margin-top:6mm;font-size:14px;line-height:1.45;color:#6B5244;max-width:120mm"
						>
							La plataforma para tiendas argentinas que quieren crecer.
						</div>
						<div
							style="margin-top:12mm;width:130mm;height:150mm;border-radius:65mm 65mm 4mm 4mm;background:#D2582F;position:relative;overflow:hidden"
						>
							<div
								style="position:absolute;inset:0;background:url(/textura/arch-pattern-accent.svg) center/20mm;opacity:.6"
							></div>
							<img
								src="/hornero/celebrando.png"
								alt=""
								style="position:absolute;left:50%;bottom:-4mm;transform:translateX(-50%);width:120mm;height:120mm;object-fit:contain"
							/>
						</div>
						<div
							style="margin-top:10mm;display:flex;flex-direction:column;gap:3mm;font-size:13px;font-weight:600"
						>
							<span>Tienda · Mercado Libre · WhatsApp</span><span>Facturación · Envíos · IA</span>
						</div>
						<div
							style="margin-top:auto;width:100%;padding:6mm;border-radius:4mm;background:#2A140C;color:#fff;display:flex;align-items:center;gap:5mm;text-align:left"
						>
							<div style="flex:1">
								<b style="font-size:16px">{PRUEBA_DIAS} días gratis</b>
								<div style="font-size:10px;color:#BDB2A8">Escaneá y armá tu tienda acá</div>
							</div>
							<Qr
								texto={QR.rollup}
								alt="QR · Crear mi tienda"
								style="width:24mm;height:24mm;border-radius:2mm;display:block;flex:none;background:#FBF7F2"
							/>
						</div>
						<div
							style="position:absolute;left:0;right:0;bottom:0;height:12mm;background:repeating-linear-gradient(45deg,rgba(42,20,12,.06) 0 3mm,transparent 3mm 6mm)"
						></div>
					</div></Lamina
				>
				<span style="font-size:12px;color:#9C8676"
					>Los últimos 12 cm (franja rayada) quedan dentro del pie.</span
				>
			</div>
			<div id="6" style="display:flex;flex-direction:column;gap:12px">
				<div style="display:flex;align-items:center;gap:10px">
					<a
						href="#6"
						style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
						>6</a
					><span style="font-size:14px;color:#6B5244"
						>Backing / lona de stand · 300 × 225 cm · escala 1:10</span
					>
				</div>
				<Lamina
					nombre="imprimibles-6-backing"
					ancho={1133.86}
					alto={850.39}
					mm={[300, 225]}
					impresion={10}
					><div
						style="width:300mm;height:225mm;background:#D2582F;color:#FBF7F2;position:relative;overflow:hidden;box-shadow:0 30px 60px -30px rgba(42,20,12,.55)"
					>
						<div
							style="position:absolute;inset:0;background:url(/textura/arch-pattern-accent.svg) center/26mm;opacity:.6"
						></div>
						<div
							style="position:absolute;inset:0;background:url(/textura/grain.png);opacity:.22;mix-blend-mode:multiply"
						></div>
						<div style="position:absolute;left:18mm;top:22mm;width:150mm">
							<img src="/wordmark-cream.png" alt="Tu Local Web" style="height:10mm" />
							<div
								style="margin-top:12mm;font-size:62px;line-height:.9;letter-spacing:-.05em;font-weight:600"
							>
								Todo lo que tienen las otras.<br /><span style="color:#2A140C"
									>Por menos plata.</span
								>
							</div>
							<div style="margin-top:8mm;display:flex;gap:3mm;flex-wrap:wrap">
								<span
									style="padding:2.4mm 5mm;border-radius:999px;background:#FBF7F2;color:#2A140C;font-size:13px;font-weight:600"
									>0 % comisión</span
								><span
									style="padding:2.4mm 5mm;border-radius:999px;background:#FBF7F2;color:#2A140C;font-size:13px;font-weight:600"
									>Migración gratis</span
								><span
									style="padding:2.4mm 5mm;border-radius:999px;background:#FBF7F2;color:#2A140C;font-size:13px;font-weight:600"
									>IA incluida</span
								>
							</div>
						</div>
						<div
							style="position:absolute;right:20mm;bottom:0;width:100mm;height:170mm;border-radius:50mm 50mm 0 0;background:#2A140C;overflow:hidden"
						>
							<img
								src="/hornero/saludo.png"
								alt=""
								style="position:absolute;left:50%;bottom:6mm;transform:translateX(-50%);width:96mm;height:96mm;object-fit:contain"
							/>
							<div
								style="position:absolute;left:0;right:0;top:16mm;text-align:center;font-size:16px;font-weight:600"
							>
								Pasá, te armamos<br />la tienda acá
							</div>
						</div>
						<div style="position:absolute;left:18mm;bottom:16mm;font-size:14px;font-weight:600">
							{DOMINIO}
						</div>
						<div
							style="position:absolute;left:0;right:0;bottom:0;height:22mm;border-top:1px dashed rgba(255,255,255,.4)"
						></div>
					</div></Lamina
				>
				<span style="font-size:12px;color:#9C8676"
					>La línea punteada marca la altura de la mesa: nada importante por debajo.</span
				>
			</div>
		</div>
	</div>

	<!-- TARJETA PAQUETES -->
	<div style="display:flex;flex-direction:column;gap:20px">
		<h2 style="margin:0;font-size:26px;letter-spacing:-.03em;font-weight:600">
			Tarjeta para paquetes · A6 105 × 148 mm · la imprimen las tiendas desde el panel
		</h2>
		<div style="display:flex;flex-wrap:wrap;gap:48px;align-items:flex-start">
			<div id="7a" style="display:flex;flex-direction:column;gap:12px">
				<div style="display:flex;align-items:center;gap:10px">
					<a
						href="#7a"
						style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
						>7a</a
					><span style="font-size:14px;color:#6B5244"
						>Con la marca de la tienda · frente y dorso</span
					><span
						style="padding:4px 10px;border-radius:999px;background:#FCEDE5;color:#C24E25;font-size:12px;font-weight:700"
						>Modelo</span
					>
				</div>
				<div style="display:flex;gap:20px;flex-wrap:wrap">
					<Lamina
						nombre="imprimibles-7a-tarjeta-paquete-frente"
						ancho={396.85}
						alto={559.37}
						mm={[105, 148]}
						><div
							style="width:105mm;height:148mm;background:#F3E9DF;position:relative;overflow:hidden;padding:10mm;display:flex;flex-direction:column;align-items:center;text-align:center;box-shadow:0 24px 50px -26px rgba(42,20,12,.45)"
						>
							<div style="font-size:16px;font-weight:700;letter-spacing:.2em">CAFÉ BRUMA</div>
							<div
								style="margin-top:8mm;width:56mm;height:66mm;border-radius:28mm 28mm 2mm 2mm;overflow:hidden"
							>
								<img
									src="https://images.unsplash.com/photo-1495474472287-4d71bcdd2085?auto=format&fit=crop&w=500&q=75"
									alt=""
									style="width:100%;height:100%;object-fit:cover"
								/>
							</div>
							<div
								style="margin-top:8mm;font-size:26px;line-height:1;letter-spacing:-.03em;font-weight:600"
							>
								¡Gracias, Lucía!
							</div>
							<div style="margin-top:3mm;font-size:10.5px;line-height:1.5;color:#6B5244">
								Lo preparamos a mano el 20 de octubre. Que lo disfrutes.
							</div>
						</div></Lamina
					>
					<Lamina
						nombre="imprimibles-7a-tarjeta-paquete-dorso"
						ancho={396.85}
						alto={559.37}
						mm={[105, 148]}
						><div
							style="width:105mm;height:148mm;background:#FBF7F2;position:relative;overflow:hidden;padding:10mm;display:flex;flex-direction:column;box-shadow:0 24px 50px -26px rgba(42,20,12,.45)"
						>
							<div style="font-size:15px;font-weight:600">¿Te gustó?</div>
							<div style="font-size:10.5px;color:#6B5244;margin-top:1mm">
								Dejanos una reseña con foto y te regalamos un 15 % en tu próxima compra.
							</div>
							<Qr
								texto={RESENA_MUESTRA}
								alt="QR · Reseña del pedido 10480"
								style="margin-top:6mm;align-self:center;width:34mm;height:34mm;border-radius:2mm;display:block;flex:none;background:#FBF7F2"
							/>
							<div
								style="margin-top:6mm;padding:4mm;border-radius:3mm;background:#fff;text-align:center"
							>
								<div style="font-size:9px;font-weight:700;letter-spacing:.12em;color:#9C8676">
									TU CUPÓN
								</div>
								<div style="font-size:20px;font-weight:700;letter-spacing:.04em">LUCIA15</div>
							</div>
							<div
								style="margin-top:auto;display:flex;justify-content:space-between;align-items:center;font-size:9px;color:#9C8676"
							>
								<span>cafebruma.com.ar · @cafebruma</span><span
									style="display:flex;align-items:center;gap:4px"
									><img src="/logo.svg" alt="" style="width:3.5mm" />Hecho con Tu Local Web</span
								>
							</div>
						</div></Lamina
					>
				</div>
			</div>
			<div id="7b" style="display:flex;flex-direction:column;gap:12px">
				<div style="display:flex;align-items:center;gap:10px">
					<a
						href="#7b"
						style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
						>7b</a
					><span style="font-size:14px;color:#6B5244"
						>Tu Local Web · para que el comprador también arme su&nbsp;tienda</span
					>
				</div>
				<div style="display:flex;gap:20px;flex-wrap:wrap">
					<Lamina
						nombre="imprimibles-7b-tarjeta-paquete-frente"
						ancho={396.85}
						alto={559.37}
						mm={[105, 148]}
						><div
							style="width:105mm;height:148mm;background:#D2582F;color:#FBF7F2;position:relative;overflow:hidden;padding:10mm;display:flex;flex-direction:column;box-shadow:0 24px 50px -26px rgba(42,20,12,.55)"
						>
							<div
								style="position:absolute;inset:0;background:url(/textura/arch-pattern-accent.svg) center/16mm;opacity:.55"
							></div>
							<div
								style="position:relative;font-size:28px;line-height:.95;letter-spacing:-.04em;font-weight:600"
							>
								Este paquete salió de una tienda Tu Local Web.
							</div>
							<img
								src="/hornero/caja.png"
								alt=""
								style="position:relative;width:62mm;height:62mm;object-fit:contain;align-self:center;margin-top:auto"
							/><img
								src="/wordmark-cream.png"
								alt="Tu Local Web"
								style="position:relative;height:5mm;align-self:center;margin-top:4mm"
							/>
						</div></Lamina
					>
					<Lamina
						nombre="imprimibles-7b-tarjeta-paquete-dorso"
						ancho={396.85}
						alto={559.37}
						mm={[105, 148]}
						><div
							style="width:105mm;height:148mm;background:#2A140C;color:#FBF7F2;position:relative;overflow:hidden;padding:10mm;display:flex;flex-direction:column;box-shadow:0 24px 50px -26px rgba(42,20,12,.6)"
						>
							<div style="font-size:17px;font-weight:600;line-height:1.15">
								¿Vos también vendés algo?
							</div>
							<div style="margin-top:3mm;font-size:10.5px;line-height:1.5;color:#BDB2A8">
								Armá tu tienda en 10 minutos. Los primeros {PRUEBA_DIAS} días son gratis, sin tarjeta
								y sin comisión por&nbsp;venta.
							</div>
							<div
								style="margin-top:7mm;padding:5mm;border-radius:3mm;background:#FBF7F2;color:#2A140C;display:flex;align-items:center;gap:4mm"
							>
								<div style="flex:1">
									<div style="font-size:8.5px;font-weight:700;letter-spacing:.12em;color:#C24E25">
										PRUEBA GRATIS
									</div>
									<div style="font-size:18px;font-weight:700">{PRUEBA_DIAS} días</div>
								</div>
								<Qr
									texto={QR.paquete}
									alt="QR · Crear mi tienda"
									style="width:20mm;height:20mm;border-radius:2mm;display:block;flex:none;background:#FBF7F2"
								/>
							</div>
							<div style="margin-top:auto;font-size:9px;color:#9C8676">{DOMINIO}</div>
						</div></Lamina
					>
				</div>
			</div>
		</div>
	</div>
</section>
