<script lang="ts">
	import Seo from '#lib/components/Seo.svelte';
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import Campos from '#lib/marca/Campos.svelte';
	import DescargarTodo from '#lib/marca/DescargarTodo.svelte';
	import Lamina from '#lib/marca/Lamina.svelte';
	import Qr from '#lib/marca/Qr.svelte';
	import {
		DOMINIO,
		EMPRESA,
		EVENTO_INICIAL,
		MAIL,
		PERSONA_INICIAL,
		PLAN,
		PRUEBA_DIAS,
		QR,
		VS_PRO,
		fechaCorta,
		fechaIso,
		fechaLarga,
		fmt,
		iniciales,
		mesAnio,
		qrArcaMuestra,
		vcard
	} from '#lib/marca/datos.ts';
	import { crearFamilia } from '#lib/marca/familia.svelte.ts';

	/** Papelería (design/Papeleria Mostrador.dc.html): medidas reales en mm, imprenta con sangrado. */
	const fam = crearFamilia('papeleria');

	let persona = $state({ ...PERSONA_INICIAL });
	let evento = $state({ ...EVENTO_INICIAL });
	const CAMPOS_PERSONA: [string, string][] = [
		['nombre', 'Nombre y apellido'],
		['cargo', 'Cargo'],
		['telefono', 'Teléfono'],
		['mail', 'Email']
	];
	const CAMPOS_EVENTO: [string, string][] = [
		['titulo', 'Evento'],
		['fecha', 'Fecha'],
		['link', 'Link del QR']
	];
	// La credencial parte el nombre en dos renglones, como el handoff.
	const nombre1 = $derived(persona.nombre.split(' ')[0]);
	const nombre2 = $derived(persona.nombre.split(' ').slice(1).join(' '));

	const ahora = new Date();
	const hoy = fechaLarga(ahora);
	const hoyIso = fechaIso(ahora);
	// Año y mes de Argentina, no los del entorno.
	const [anio, mes] = hoyIso.split('-').map(Number);
	// Mediodía UTC: en Argentina es el mismo día, sea cual sea la zona del navegador.
	const primero = new Date(Date.UTC(anio, mes - 1, 1, 12));
	const ultimo = new Date(Date.UTC(anio, mes, 0, 12));
	const periodo = `${fechaCorta(primero)} al ${fechaCorta(ultimo)}`;
	/** "$ 49.900,00", como un comprobante. */
	const pesos = (n: number) =>
		'$\u00a0' + n.toLocaleString('es-AR', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
</script>

<Seo
	title="Papelería · Marca · Tu Local Web"
	description="Tarjetas, hoja membretada, sobre, firmas de email, factura y remito modelo, carpeta, credenciales y stickers."
	path="/marca/papeleria"
	noindex
/>

<section class="lienzo" bind:contentRect={fam.rect}>
	<header class="lienzo-cab">
		<div style="flex:1;min-width:min(100%,300px)">
			<img src="/wordmark-brown.png" alt="Tu Local Web" style="height:26px;display:block" />
			<h1
				style="margin:22px 0 0;font-size:56px;line-height:1;letter-spacing:-.04em;font-weight:600"
			>
				Papelería corporativa
			</h1>
			<p style="margin:12px 0 0;font-size:18px;color:#6B5244;max-width:640px;text-wrap:pretty">
				Medidas reales en milímetros, listas para imprenta con 3 mm de sangrado. Las piezas clave
				tienen dos versiones (a/b). Los datos de la empresa salen de los legales; los tuyos, de los
				campos de&nbsp;abajo.
			</p>
			<div style="margin-top:20px"><DescargarTodo /></div>
		</div>
		<div style="display:flex;gap:8px;flex-wrap:wrap">
			<span
				style="display:flex;align-items:center;gap:8px;padding:8px 14px;border-radius:999px;background:#fff;font-size:13px;font-weight:600"
				><span style="width:14px;height:14px;border-radius:50%;background:#D2582F"></span>Terracota
				· Pantone 7592 C</span
			><span
				style="display:flex;align-items:center;gap:8px;padding:8px 14px;border-radius:999px;background:#fff;font-size:13px;font-weight:600"
				><span style="width:14px;height:14px;border-radius:50%;background:#2A140C"></span>Tostado ·
				Pantone 4625 C</span
			><span
				style="display:flex;align-items:center;gap:8px;padding:8px 14px;border-radius:999px;background:#fff;font-size:13px;font-weight:600"
				><span
					style="width:14px;height:14px;border-radius:50%;background:#FBF7F2;box-shadow:inset 0 0 0 1px #D9CBBE"
				></span>Crema · papel ahuesado</span
			>
		</div>
	</header>

	<Campos
		titulo="Tus datos"
		ayuda="Van en las tarjetas, la carta, la firma de email y la credencial."
		campos={CAMPOS_PERSONA}
		bind:valores={persona}
		foto
	/>

	<!-- TARJETAS -->
	<div style="display:flex;flex-direction:column;gap:20px">
		<h2 style="margin:0;font-size:26px;letter-spacing:-.03em;font-weight:600">
			Tarjetas personales · 85 × 55 mm
		</h2>
		<div style="display:flex;flex-wrap:wrap;gap:56px;align-items:flex-start">
			<div id="1a" style="display:flex;flex-direction:column;gap:14px">
				<div style="display:flex;align-items:center;gap:10px">
					<a
						href="#1a"
						style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
						>1a</a
					><span style="font-size:14px;color:#6B5244">Clásica · frente terracota con arco</span>
				</div>
				<div style="display:flex;gap:20px;flex-wrap:wrap">
					<Lamina nombre="papeleria-1a-tarjeta-frente" ancho={321.26} alto={207.87} mm={[85, 55]}
						><div
							style="width:85mm;height:55mm;border-radius:3mm;background:#D2582F url(/textura/arch-pattern-accent.svg) center/22mm;position:relative;overflow:hidden;box-shadow:0 1px 0 rgba(0,0,0,.05),0 18px 40px -18px rgba(42,20,12,.55)"
						>
							<div
								style="position:absolute;inset:0;background:url(/textura/grain.png);opacity:.25;mix-blend-mode:multiply"
							></div>
							<div
								style="position:absolute;left:50%;top:50%;transform:translate(-50%,-50%);display:flex;flex-direction:column;align-items:center;gap:3mm"
							>
								<img src="/logo.svg" alt="" style="width:16mm;filter:brightness(0) invert(1)" /><img
									src="/wordmark-cream.png"
									alt="Tu Local Web"
									style="height:4mm"
								/>
							</div>
						</div></Lamina
					>
					<Lamina nombre="papeleria-1a-tarjeta-dorso" ancho={321.26} alto={207.87} mm={[85, 55]}
						><div
							style="width:85mm;height:55mm;border-radius:3mm;background:#FBF7F2;position:relative;overflow:hidden;padding:6mm;display:flex;flex-direction:column;box-shadow:0 1px 0 rgba(0,0,0,.05),0 18px 40px -18px rgba(42,20,12,.45)"
						>
							<div style="display:flex;justify-content:space-between;align-items:start">
								<div>
									<div style="font-size:13px;font-weight:600;letter-spacing:-.01em">
										{persona.nombre}
									</div>
									<div style="font-size:8.5px;color:#7A6252;margin-top:1px">{persona.cargo}</div>
								</div>
								<img src="/logo.svg" alt="" style="width:7mm" />
							</div>
							<div
								style="margin-top:auto;display:flex;flex-direction:column;gap:1.2mm;font-size:8px;color:#2A140C"
							>
								<span style="display:flex;align-items:center;gap:5px"
									><Icon
										icon={I.phEnvelopeSimpleBold}
										size={9}
										style="color:#D2582F"
									/>{persona.mail}</span
								>
								<span style="display:flex;align-items:center;gap:5px"
									><Icon
										icon={I.phWhatsappLogoBold}
										size={9}
										style="color:#D2582F"
									/>{persona.telefono}</span
								>
								<span style="display:flex;align-items:center;gap:5px"
									><Icon
										icon={I.phGlobeSimpleBold}
										size={9}
										style="color:#D2582F"
									/>tulocalweb.com.ar</span
								>
							</div>
							<div
								style="position:absolute;right:0;bottom:0;width:22mm;height:15mm;border-radius:11mm 11mm 0 0;background:#D2582F;transform:translate(4mm,4mm)"
							></div>
						</div></Lamina
					>
				</div>
			</div>
			<div id="1b" style="display:flex;flex-direction:column;gap:14px">
				<div style="display:flex;align-items:center;gap:10px">
					<a
						href="#1b"
						style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
						>1b</a
					><span style="font-size:14px;color:#6B5244">Con hornero · dorso tostado, frase y QR</span>
				</div>
				<div style="display:flex;gap:20px;flex-wrap:wrap">
					<Lamina nombre="papeleria-1b-tarjeta-frente" ancho={321.26} alto={207.87} mm={[85, 55]}
						><div
							style="width:85mm;height:55mm;border-radius:3mm;background:#2A140C;position:relative;overflow:hidden;padding:6mm;color:#FBF7F2;box-shadow:0 18px 40px -18px rgba(42,20,12,.7)"
						>
							<div
								style="position:absolute;inset:0;background:url(/textura/grain.png);opacity:.18"
							></div>
							<img
								src="/wordmark-cream.png"
								alt="Tu Local Web"
								style="height:3.6mm;position:relative"
							/>
							<div
								style="position:relative;margin-top:5mm;font-size:15px;line-height:1.05;letter-spacing:-.03em;font-weight:600;max-width:44mm"
							>
								Todo lo que tienen las otras.<br /><span style="color:#E8875F"
									>Por menos plata.</span
								>
							</div>
							<img
								src="/hornero/saludo.png"
								alt=""
								style="position:absolute;right:-2mm;bottom:-3mm;width:30mm;height:30mm;object-fit:contain"
							/>
						</div></Lamina
					>
					<Lamina nombre="papeleria-1b-tarjeta-dorso" ancho={321.26} alto={207.87} mm={[85, 55]}
						><div
							style="width:85mm;height:55mm;border-radius:3mm;background:#FBF7F2;position:relative;overflow:hidden;padding:6mm;display:grid;grid-template-columns:1fr auto;gap:4mm;box-shadow:0 18px 40px -18px rgba(42,20,12,.45)"
						>
							<div style="display:flex;flex-direction:column">
								<div
									style="width:9mm;height:9mm;border-radius:50%;overflow:hidden;background:#F3E9DF"
								>
									{#if persona.foto}<img
											src={persona.foto}
											alt=""
											style="width:100%;height:100%;object-fit:cover"
										/>{:else}<span
											style="width:100%;height:100%;display:grid;place-items:center;background:#F3E9DF;color:#C24E25;font-weight:600;font-size:12px"
											>{iniciales(persona.nombre)}</span
										>{/if}
								</div>
								<div style="font-size:12.5px;font-weight:600;margin-top:2.5mm">
									{persona.nombre}
								</div>
								<div style="font-size:8px;color:#7A6252">{persona.cargo}</div>
								<div style="margin-top:auto;font-size:7.5px;line-height:1.55;color:#2A140C">
									{persona.mail}<br />{persona.telefono}
								</div>
							</div>
							<div
								style="display:flex;flex-direction:column;align-items:center;justify-content:flex-end;gap:1.5mm"
							>
								<Qr
									texto={vcard(persona)}
									alt="QR · Contacto de {persona.nombre}"
									style="width:18mm;height:18mm;border-radius:2mm;display:block;flex:none;background:#FBF7F2"
								/><span style="font-size:6.5px;color:#7A6252">Guardá mi contacto</span>
							</div>
						</div></Lamina
					>
				</div>
			</div>
		</div>
	</div>

	<!-- MEMBRETE + SOBRE -->
	<div style="display:flex;flex-direction:column;gap:20px">
		<h2 style="margin:0;font-size:26px;letter-spacing:-.03em;font-weight:600">
			Hoja membretada · A4 y Carta · Sobre DL
		</h2>
		<div style="display:flex;flex-wrap:wrap;gap:56px;align-items:flex-start">
			<div id="2a" style="display:flex;flex-direction:column;gap:14px">
				<div style="display:flex;align-items:center;gap:10px">
					<a
						href="#2a"
						style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
						>2a</a
					><span style="font-size:14px;color:#6B5244"
						>Sobria · A4 210 × 297 mm · el texto se edita con un&nbsp;clic</span
					>
				</div>
				<Lamina nombre="papeleria-2a-membrete-a4" ancho={793.7} alto={1122.52} mm={[210, 297]}
					><div
						style="width:210mm;height:297mm;background:#fff;position:relative;padding:20mm 22mm 18mm;display:flex;flex-direction:column;box-shadow:0 30px 60px -30px rgba(42,20,12,.45)"
					>
						<div style="display:flex;justify-content:space-between;align-items:center">
							<img src="/wordmark-brown.png" alt="Tu Local Web" style="height:7mm" />
							<div style="text-align:right;font-size:10px;line-height:1.5;color:#7A6252">
								{EMPRESA.ciudad}<br />{EMPRESA.region}
							</div>
						</div>
						<div style="height:1px;background:#EADFD4;margin:9mm 0 12mm"></div>
						<div style="font-size:11.5px;line-height:1.7;color:#2A140C;max-width:150mm">
							<div style="color:#7A6252">{EMPRESA.ciudadCarta}, {hoy}</div>
							<div
								contenteditable="true"
								spellcheck="false"
								role="textbox"
								aria-multiline="true"
								aria-label="Texto de la carta"
							>
								<div style="margin-top:9mm;font-weight:600">Sres. Nombre de la empresa</div>
								<div style="color:#7A6252">At. Área o persona</div>
								<p style="margin:8mm 0 0">De nuestra mayor consideración:</p>
								<p style="margin:4mm 0 0;text-wrap:pretty">
									Nos dirigimos a ustedes para contarles cómo podemos ayudarlos a vender en su
									tienda online, Mercado Libre, WhatsApp y el local con un solo stock, con la
									factura de ARCA y los envíos en el mismo&nbsp;panel.
								</p>
								<p style="margin:4mm 0 0;text-wrap:pretty">
									Adjuntamos el detalle de la propuesta. Quedamos a disposición para
									cualquier&nbsp;consulta.
								</p>
								<p style="margin:10mm 0 0">Saludos cordiales,</p>
							</div>
							<div style="margin-top:14mm;font-weight:600">{persona.nombre}</div>
							<div style="color:#7A6252">{persona.cargo}</div>
						</div>
						<div
							style="margin-top:auto;display:flex;justify-content:space-between;align-items:end;font-size:9px;color:#9C8676;gap:8mm"
						>
							<span>{EMPRESA.titular} · CUIT {EMPRESA.cuit} · IVA Responsable Inscripto</span><span
								style="display:flex;align-items:center;gap:6px;white-space:nowrap"
								><img src="/logo.svg" alt="" style="width:5mm" />tulocalweb.com.ar</span
							>
						</div>
						<div style="position:absolute;left:0;top:0;bottom:0;width:4mm;background:#D2582F"></div>
					</div></Lamina
				>
			</div>
			<div id="2b" style="display:flex;flex-direction:column;gap:14px">
				<div style="display:flex;align-items:center;gap:10px">
					<a
						href="#2b"
						style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
						>2b</a
					><span style="font-size:14px;color:#6B5244"
						>Con arco · Carta 8,5 × 11 in · el texto se edita con un&nbsp;clic</span
					>
				</div>
				<Lamina nombre="papeleria-2b-membrete-carta" ancho={816} alto={1056} mm={[215.9, 279.4]}
					><div
						style="width:8.5in;height:11in;background:#FBF7F2;position:relative;overflow:hidden;padding:22mm 22mm 18mm;display:flex;flex-direction:column;box-shadow:0 30px 60px -30px rgba(42,20,12,.45)"
					>
						<div
							style="position:absolute;right:-30mm;top:-46mm;width:110mm;height:90mm;border-radius:55mm 55mm 0 0;border:14mm solid #D2582F;border-bottom:0;opacity:.95"
						></div>
						<img src="/logo.svg" alt="" style="width:12mm;position:relative" />
						<div
							style="margin-top:4mm;font-size:10px;letter-spacing:.14em;text-transform:uppercase;font-weight:700;color:#C24E25"
						>
							Propuesta comercial
						</div>
						<div
							style="margin-top:20mm;font-size:11.5px;line-height:1.7;max-width:150mm;position:relative"
						>
							<div style="color:#7A6252">{hoy}</div>
							<div
								contenteditable="true"
								spellcheck="false"
								role="textbox"
								aria-multiline="true"
								aria-label="Título de la propuesta"
								style="margin-top:8mm;font-size:22px;line-height:1.15;letter-spacing:-.03em;font-weight:600;max-width:120mm"
							>
								Tu tienda en Tu Local Web, con todo lo que ya&nbsp;usás.
							</div>
							<div
								contenteditable="true"
								spellcheck="false"
								role="textbox"
								aria-multiline="true"
								aria-label="Texto de la propuesta"
							>
								<p style="margin:8mm 0 0;text-wrap:pretty">
									Hola. Como charlamos, la mudanza desde tu plataforma actual corre por nuestra
									cuenta: productos, clientes y dominio. Probás {PRUEBA_DIAS} días gratis y el plan {PLAN
										.pro.nombre} queda en {fmt(PLAN.pro.precio)} por mes, sin comisión por&nbsp;venta.
								</p>
								<p style="margin:4mm 0 0;text-wrap:pretty">Cualquier cosa, escribime.</p>
							</div>
							<div style="margin-top:12mm;font-weight:600">{persona.nombre}</div>
							<div style="color:#7A6252">{persona.mail} · {persona.telefono}</div>
						</div>
						<div
							style="margin-top:auto;padding-top:5mm;border-top:1px solid #E6D9CC;display:flex;justify-content:space-between;font-size:9px;color:#9C8676;gap:6mm;flex-wrap:wrap"
						>
							<span>{EMPRESA.titular} · CUIT {EMPRESA.cuit}</span><span>{EMPRESA.domicilio}</span
							><span>{DOMINIO}</span>
						</div>
					</div></Lamina
				>
			</div>
			<div id="3" style="display:flex;flex-direction:column;gap:14px">
				<div style="display:flex;align-items:center;gap:10px">
					<a
						href="#3"
						style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
						>3</a
					><span style="font-size:14px;color:#6B5244"
						>Sobre DL · 220 × 110 mm · frente y solapa · el destinatario se edita con un&nbsp;clic</span
					>
				</div>
				<Lamina nombre="papeleria-3-sobre-frente" ancho={831.5} alto={415.75} mm={[220, 110]}
					><div
						style="width:220mm;height:110mm;background:#FBF7F2;position:relative;overflow:hidden;padding:10mm 12mm;box-shadow:0 24px 50px -26px rgba(42,20,12,.45)"
					>
						<div style="display:flex;align-items:center;gap:3mm">
							<img src="/logo.svg" alt="" style="width:8mm" />
							<div style="font-size:8.5px;line-height:1.45;color:#7A6252">
								{EMPRESA.titular}<br />{EMPRESA.domicilio}
							</div>
						</div>
						<div
							contenteditable="true"
							spellcheck="false"
							role="textbox"
							aria-multiline="true"
							aria-label="Destinatario del sobre"
							style="position:absolute;left:110mm;top:52mm;font-size:12px;line-height:1.6"
						>
							<div style="font-weight:600">Nombre del destinatario</div>
							<div style="color:#7A6252">Calle y número<br />Código postal, ciudad y provincia</div>
						</div>
						<div
							style="position:absolute;left:0;right:0;bottom:0;height:6mm;background:#D2582F url(/textura/arch-pattern-accent.svg) center/10mm"
						></div>
					</div></Lamina
				>
				<Lamina nombre="papeleria-3-sobre-solapa" ancho={831.5} alto={415.75} mm={[220, 110]}
					><div
						style="width:220mm;height:110mm;background:#D2582F;position:relative;overflow:hidden;box-shadow:0 24px 50px -26px rgba(42,20,12,.55)"
					>
						<div
							style="position:absolute;inset:0;background:url(/textura/arch-pattern-accent.svg) center/24mm;opacity:.7"
						></div>
						<div
							style="position:absolute;left:0;right:0;top:0;height:46mm;background:#C24E25;clip-path:polygon(0 0,100% 0,50% 100%)"
						></div>
						<div
							style="position:absolute;left:50%;bottom:16mm;transform:translateX(-50%);display:flex;flex-direction:column;align-items:center;gap:2mm;color:#FBF7F2"
						>
							<img src="/wordmark-cream.png" alt="Tu Local Web" style="height:6mm" /><span
								style="font-size:10px;opacity:.9">Para los que venden en serio.</span
							>
						</div>
					</div></Lamina
				>
			</div>
		</div>
	</div>

	<!-- FIRMA EMAIL -->
	<div style="display:flex;flex-direction:column;gap:20px">
		<h2 style="margin:0;font-size:26px;letter-spacing:-.03em;font-weight:600">
			Firma de email · 600 px
		</h2>
		<div style="display:flex;flex-wrap:wrap;gap:40px;align-items:flex-start">
			<div id="4a" style="display:flex;flex-direction:column;gap:14px">
				<div style="display:flex;align-items:center;gap:10px">
					<a
						href="#4a"
						style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
						>4a</a
					><span style="font-size:14px;color:#6B5244"
						>Completa · foto, cargo y banner de campaña</span
					>
				</div>
				<Lamina nombre="papeleria-4a-firma-email" ancho={600}
					><div
						style="width:600px;padding:22px;background:#fff;border-radius:8px;box-shadow:0 18px 40px -24px rgba(42,20,12,.4)"
					>
						<div style="font-size:14px;color:#6B5244;margin-bottom:18px">Saludos,</div>
						<div style="display:flex;gap:16px;align-items:center">
							{#if persona.foto}<img
									src={persona.foto}
									alt=""
									style="width:64px;height:64px;border-radius:32px 32px 6px 6px;object-fit:cover"
								/>{:else}<span
									style="width:64px;height:64px;flex:none;border-radius:32px 32px 6px 6px;display:grid;place-items:center;background:#F3E9DF;color:#C24E25;font-weight:600;font-size:20px"
									>{iniciales(persona.nombre)}</span
								>{/if}
							<div style="flex:1;border-left:2px solid #D2582F;padding-left:16px">
								<div style="font-size:16px;font-weight:600">{persona.nombre}</div>
								<div style="font-size:13px;color:#7A6252">{persona.cargo} · Tu Local Web</div>
								<div
									style="margin-top:6px;font-size:12.5px;color:#2A140C;display:flex;gap:12px;flex-wrap:wrap"
								>
									<span>{persona.telefono}</span><span style="color:#C24E25">{DOMINIO}</span>
								</div>
							</div>
							<img src="/logo.svg" alt="" style="width:30px;align-self:flex-start" />
						</div>
						<div
							style="margin-top:18px;border-radius:12px;background:#FCEDE5;padding:12px 14px;display:flex;align-items:center;gap:12px"
						>
							<img
								src="/hornero/whatsapp.png"
								alt=""
								style="width:46px;height:46px;object-fit:contain"
							/>
							<div style="flex:1;font-size:13px;line-height:1.4">
								<b>Pasate de Tiendanube sin costo.</b> Migramos tus productos, clientes y dominio en un&nbsp;día.
							</div>
							<span
								style="padding:8px 12px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:600;white-space:nowrap"
								>Quiero migrar</span
							>
						</div>
						<div style="margin-top:12px;font-size:10px;line-height:1.5;color:#9C8676">
							{EMPRESA.titular} · CUIT {EMPRESA.cuit}. Este mensaje es confidencial. Si lo recibiste
							por error, avisanos y borralo.
						</div>
					</div></Lamina
				>
			</div>
			<div id="4b" style="display:flex;flex-direction:column;gap:14px">
				<div style="display:flex;align-items:center;gap:10px">
					<a
						href="#4b"
						style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
						>4b</a
					><span style="font-size:14px;color:#6B5244">Mínima · para respuestas y soporte</span>
				</div>
				<Lamina nombre="papeleria-4b-firma-email" ancho={600}
					><div
						style="width:600px;padding:22px;background:#fff;border-radius:8px;box-shadow:0 18px 40px -24px rgba(42,20,12,.4)"
					>
						<div style="display:flex;align-items:center;gap:12px">
							<img src="/logo.svg" alt="" style="width:26px" />
							<div style="font-size:13.5px;line-height:1.45">
								<b>Equipo de Soporte</b> · Tu Local Web<br /><span style="color:#7A6252"
									>De lunes a sábado de 8 a 22 h · {MAIL.ayuda}</span
								>
							</div>
						</div>
						<div style="margin-top:12px;display:flex;gap:6px;flex-wrap:wrap">
							<span
								style="display:flex;align-items:center;gap:6px;padding:6px 10px;border-radius:999px;background:#F3EBE2;font-size:12px;font-weight:600"
								><Icon icon={I.logosWhatsappIcon} size={13} />WhatsApp</span
							><span
								style="display:flex;align-items:center;gap:6px;padding:6px 10px;border-radius:999px;background:#F3EBE2;font-size:12px;font-weight:600"
								><Icon icon={I.phLifebuoyBold} size={13} style="color:#D2582F" />Centro de ayuda</span
							><span
								style="display:flex;align-items:center;gap:6px;padding:6px 10px;border-radius:999px;background:#F3EBE2;font-size:12px;font-weight:600"
								><Icon icon={I.phPulseBold} size={13} style="color:#2F8F63" />Estado: todo
								funcionando</span
							>
						</div>
					</div></Lamina
				>
			</div>
		</div>
	</div>

	<!-- FACTURA / REMITO -->
	<div style="display:flex;flex-direction:column;gap:20px">
		<h2 style="margin:0;font-size:26px;letter-spacing:-.03em;font-weight:600">
			Factura y remito · A4
		</h2>
		<div style="display:flex;flex-wrap:wrap;gap:56px;align-items:flex-start">
			<div id="5a" style="display:flex;flex-direction:column;gap:14px">
				<div style="display:flex;align-items:center;gap:10px">
					<a
						href="#5a"
						style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
						>5a</a
					><span style="font-size:14px;color:#6B5244"
						>Factura B · la que emite Tu Local Web a sus clientes</span
					><span
						style="padding:4px 10px;border-radius:999px;background:#FCEDE5;color:#C24E25;font-size:12px;font-weight:700"
						>Modelo</span
					>
				</div>
				<Lamina nombre="papeleria-5a-factura-b-modelo" ancho={793.7} alto={1122.52} mm={[210, 297]}
					><div
						style="width:210mm;height:297mm;background:#fff;padding:14mm 16mm;display:flex;flex-direction:column;font-size:10.5px;position:relative;overflow:hidden;box-shadow:0 30px 60px -30px rgba(42,20,12,.45)"
					>
						<div
							aria-hidden="true"
							style="position:absolute;left:50%;top:50%;transform:translate(-50%,-50%) rotate(-30deg);font-size:120px;font-weight:700;letter-spacing:.08em;color:rgba(210,88,47,.08);pointer-events:none;white-space:nowrap"
						>
							MODELO
						</div>
						<div
							style="display:grid;grid-template-columns:1fr auto 1fr;border:1px solid #2A140C;border-radius:2mm"
						>
							<div style="padding:5mm">
								<img src="/wordmark-brown.png" alt="Tu Local Web" style="height:6mm" />
								<div style="margin-top:3mm;line-height:1.55;color:#4A3428">
									{EMPRESA.titular}<br />{EMPRESA.domicilio}<br />IVA Responsable Inscripto
								</div>
							</div>
							<div
								style="display:flex;flex-direction:column;align-items:center;border-left:1px solid #2A140C;border-right:1px solid #2A140C;padding:3mm 6mm"
							>
								<div
									style="width:14mm;height:14mm;border:1.5px solid #2A140C;display:grid;place-items:center;font-size:28px;font-weight:700"
								>
									B
								</div>
								<div style="font-size:8px;margin-top:1mm">Cód. 006</div>
							</div>
							<div style="padding:5mm;line-height:1.6">
								<div style="font-size:16px;font-weight:700;letter-spacing:-.02em">FACTURA</div>
								<div>N.º <b>0001-00000000</b></div>
								<div>Fecha: {fechaCorta()}</div>
								<div style="color:#7A6252">
									CUIT {EMPRESA.cuit} · IIBB 000-000000-0<br />Inicio de actividades: MM/AAAA
								</div>
							</div>
						</div>
						<div
							style="margin-top:5mm;padding:4mm 5mm;border-radius:2mm;background:#F7F1EA;display:grid;grid-template-columns:1fr 1fr;gap:2mm 8mm;line-height:1.5"
						>
							<div><span style="color:#7A6252">Cliente:</span> <b>Cliente de muestra</b></div>
							<div><span style="color:#7A6252">CUIT:</span> 00-00000000-0</div>
							<div><span style="color:#7A6252">Domicilio:</span> Calle 123, Ciudad</div>
							<div><span style="color:#7A6252">Condición IVA:</span> Monotributo</div>
							<div><span style="color:#7A6252">Período:</span> {periodo}</div>
							<div><span style="color:#7A6252">Pago:</span> Débito automático Mercado Pago</div>
						</div>
						<div
							style="margin-top:6mm;display:grid;grid-template-columns:1fr 20mm 28mm 28mm;font-weight:600;padding:0 0 2mm;border-bottom:1.5px solid #2A140C"
						>
							<span>Concepto</span><span style="text-align:right">Cant.</span><span
								style="text-align:right">P. unitario</span
							><span style="text-align:right">Subtotal</span>
						</div>
						<div
							style="display:grid;grid-template-columns:1fr 20mm 28mm 28mm;padding:3mm 0;border-bottom:1px solid #EADFD4"
						>
							<span>Plan {PLAN.pro.nombre} · {mesAnio()}</span><span style="text-align:right"
								>1</span
							><span style="text-align:right">{pesos(PLAN.pro.precio)}</span><span
								style="text-align:right">{pesos(PLAN.pro.precio)}</span
							>
						</div>
						<div
							style="margin-top:6mm;margin-left:auto;width:80mm;display:flex;flex-direction:column;gap:1.5mm"
						>
							<div style="display:flex;justify-content:space-between">
								<span style="color:#7A6252">Subtotal</span><span>{pesos(PLAN.pro.precio)}</span>
							</div>
							<div
								style="display:flex;justify-content:space-between;padding-top:2mm;border-top:1.5px solid #2A140C;font-size:15px;font-weight:700"
							>
								<span>Total</span><span>{pesos(PLAN.pro.precio)}</span>
							</div>
							<div style="font-size:8.5px;color:#7A6252;line-height:1.4">
								Régimen de Transparencia Fiscal al Consumidor (Ley 27.743): IVA contenido {pesos(
									PLAN.pro.precio - PLAN.pro.precio / 1.21
								)}
							</div>
						</div>
						<div
							style="margin-top:auto;display:flex;gap:6mm;align-items:center;padding-top:6mm;border-top:1px solid #EADFD4"
						>
							<Qr
								texto={qrArcaMuestra(PLAN.pro.precio, hoyIso)}
								alt="QR de ARCA de muestra"
								style="width:22mm;height:22mm;border-radius:2mm;display:block;flex:none;background:#FBF7F2"
							/>
							<div style="flex:1;line-height:1.6">
								<div><b>CAE</b> 00000000000000 · Vto. —</div>
								<div style="color:#7A6252">
									Modelo sin validez fiscal: cliente, CAE y QR son de&nbsp;muestra.
								</div>
							</div>
							<div style="display:flex;align-items:center;gap:6px;color:#7A6252">
								<img src="/logo.svg" alt="" style="width:6mm" />Gracias por vender con nosotros
							</div>
						</div>
					</div></Lamina
				>
			</div>
			<div id="5b" style="display:flex;flex-direction:column;gap:14px">
				<div style="display:flex;align-items:center;gap:10px">
					<a
						href="#5b"
						style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
						>5b</a
					><span style="font-size:14px;color:#6B5244"
						>Remito R · para envíos de kits y hardware</span
					><span
						style="padding:4px 10px;border-radius:999px;background:#FCEDE5;color:#C24E25;font-size:12px;font-weight:700"
						>Modelo</span
					>
				</div>
				<Lamina nombre="papeleria-5b-remito-modelo" ancho={793.7} alto={1122.52} mm={[210, 297]}
					><div
						style="width:210mm;height:297mm;background:#fff;padding:14mm 16mm;display:flex;flex-direction:column;font-size:10.5px;position:relative;overflow:hidden;box-shadow:0 30px 60px -30px rgba(42,20,12,.45)"
					>
						<div
							aria-hidden="true"
							style="position:absolute;left:50%;top:50%;transform:translate(-50%,-50%) rotate(-30deg);font-size:120px;font-weight:700;letter-spacing:.08em;color:rgba(210,88,47,.08);pointer-events:none;white-space:nowrap"
						>
							MODELO
						</div>
						<div style="display:flex;align-items:start;gap:6mm">
							<img src="/logo.svg" alt="" style="width:12mm" />
							<div style="flex:1;line-height:1.55;color:#4A3428">
								<b style="color:#2A140C;font-size:12px">{EMPRESA.titular}</b><br
								/>{EMPRESA.domicilio}<br />CUIT {EMPRESA.cuit}
							</div>
							<div
								style="width:13mm;height:13mm;border:1.5px solid #2A140C;display:grid;place-items:center;font-size:24px;font-weight:700"
							>
								R
							</div>
							<div style="text-align:right;line-height:1.6">
								<div style="font-size:16px;font-weight:700">REMITO</div>
								<div>N.º <b>0001-00000000</b></div>
								<div>{fechaCorta()}</div>
							</div>
						</div>
						<div style="margin-top:8mm;display:grid;grid-template-columns:1fr 1fr;gap:5mm">
							<div style="padding:4mm 5mm;border-radius:3mm;background:#F7F1EA;line-height:1.55">
								<div
									style="font-size:8.5px;letter-spacing:.1em;text-transform:uppercase;font-weight:700;color:#C24E25"
								>
									Destinatario
								</div>
								<b>Destinatario de muestra</b><br />Calle 123, Ciudad<br />CUIT 00-00000000-0
							</div>
							<div style="padding:4mm 5mm;border-radius:3mm;background:#F7F1EA;line-height:1.55">
								<div
									style="font-size:8.5px;letter-spacing:.1em;text-transform:uppercase;font-weight:700;color:#C24E25"
								>
									Transporte
								</div>
								<b>Andreani</b> · Sucursal de destino<br />Seguimiento 000000000000000<br />2 bultos
								· 3,4 kg
							</div>
						</div>
						<div
							style="margin-top:7mm;display:grid;grid-template-columns:24mm 1fr 20mm;font-weight:600;padding-bottom:2mm;border-bottom:1.5px solid #2A140C"
						>
							<span>Código</span><span>Descripción</span><span style="text-align:right">Cant.</span>
						</div>
						<div
							style="display:grid;grid-template-columns:24mm 1fr 20mm;padding:3mm 0;border-bottom:1px solid #EADFD4"
						>
							<span>KIT-POS-01</span><span
								>Kit punto de venta Tu Local Web (lector QR + soporte)</span
							><span style="text-align:right">1</span>
						</div>
						<div
							style="display:grid;grid-template-columns:24mm 1fr 20mm;padding:3mm 0;border-bottom:1px solid #EADFD4"
						>
							<span>IMP-TER-80</span><span>Impresora térmica de etiquetas 80 mm</span><span
								style="text-align:right">1</span
							>
						</div>
						<div
							style="display:grid;grid-template-columns:24mm 1fr 20mm;padding:3mm 0;border-bottom:1px solid #EADFD4"
						>
							<span>ROL-ETQ-10</span><span>Rollos de etiquetas 100 × 150 mm</span><span
								style="text-align:right">10</span
							>
						</div>
						<div
							style="display:grid;grid-template-columns:24mm 1fr 20mm;padding:3mm 0;border-bottom:1px solid #EADFD4"
						>
							<span>STK-BIEN-50</span><span>Pack de stickers de bienvenida</span><span
								style="text-align:right">1</span
							>
						</div>
						<div
							style="margin-top:auto;display:grid;grid-template-columns:1fr 1fr 1fr;gap:8mm;padding-top:20mm"
						>
							<div style="border-top:1px solid #2A140C;padding-top:2mm;color:#7A6252">Firma</div>
							<div style="border-top:1px solid #2A140C;padding-top:2mm;color:#7A6252">
								Aclaración
							</div>
							<div style="border-top:1px solid #2A140C;padding-top:2mm;color:#7A6252">
								DNI y fecha de recepción
							</div>
						</div>
						<div style="margin-top:6mm;font-size:8.5px;color:#9C8676">
							Modelo · CAI 00000000000000 · Vto. — · Documento no válido como factura
						</div>
					</div></Lamina
				>
			</div>
		</div>
	</div>

	<!-- CARPETA -->
	<div id="6" style="display:flex;flex-direction:column;gap:20px">
		<div style="display:flex;align-items:center;gap:12px">
			<a
				href="#6"
				style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
				>6</a
			>
			<h2 style="margin:0;font-size:26px;letter-spacing:-.03em;font-weight:600">
				Carpeta institucional · 230 × 310 mm cerrada
			</h2>
		</div>
		<div style="display:flex;flex-wrap:wrap;gap:40px;align-items:flex-start">
			<div style="display:flex;flex-direction:column;gap:10px">
				<span style="font-size:14px;color:#6B5244">Exterior abierta · contratapa | tapa</span>
				<Lamina
					nombre="papeleria-6-carpeta-exterior"
					ancho={1738.58}
					alto={1171.65}
					mm={[460, 310]}
					escala={0.62}
					><div
						style="width:460mm;height:310mm;display:grid;grid-template-columns:1fr 1fr;background:linear-gradient(90deg,#2A140C 50%,#D2582F 50%);box-shadow:0 30px 60px -30px rgba(42,20,12,.55);"
					>
						<div
							style="background:#2A140C;color:#FBF7F2;position:relative;overflow:hidden;padding:22mm;display:flex;flex-direction:column"
						>
							<div
								style="position:absolute;inset:0;background:url(/textura/grain.png);opacity:.15"
							></div>
							<div
								style="position:relative;font-size:26px;line-height:1.1;letter-spacing:-.03em;font-weight:600;max-width:150mm"
							>
								Todo lo que tienen las otras, por menos plata y más&nbsp;seguro.
							</div>
							<div
								style="position:relative;margin-top:14mm;display:grid;grid-template-columns:repeat(3,1fr);gap:8mm"
							>
								<div>
									<div style="font-size:30px;font-weight:600;color:#E8875F">
										−{VS_PRO.descuento}%
									</div>
									<div style="font-size:12px;color:#BDB2A8">más barato que Tiendanube</div>
								</div>
								<div>
									<div style="font-size:30px;font-weight:600;color:#E8875F">0%</div>
									<div style="font-size:12px;color:#BDB2A8">comisión por venta</div>
								</div>
								<div>
									<div style="font-size:30px;font-weight:600;color:#E8875F">{PRUEBA_DIAS} días</div>
									<div style="font-size:12px;color:#BDB2A8">de prueba gratis</div>
								</div>
							</div>
							<div
								style="position:relative;margin-top:auto;display:flex;justify-content:space-between;align-items:end"
							>
								<div style="font-size:12px;line-height:1.6;color:#BDB2A8">
									{DOMINIO}<br />{MAIL.hola}
								</div>
								<Qr
									texto={QR.carpeta}
									alt="QR · {DOMINIO}"
									style="width:24mm;height:24mm;border-radius:2mm;display:block;flex:none;background:#FBF7F2"
								/>
							</div>
						</div>
						<div
							style="background:#D2582F;position:relative;overflow:hidden;display:flex;flex-direction:column;align-items:center;justify-content:center"
						>
							<div
								style="position:absolute;inset:0;background:url(/textura/arch-pattern-accent.svg) center/40mm;opacity:.75"
							></div>
							<div
								style="position:absolute;inset:0;background:url(/textura/grain.png);opacity:.22;mix-blend-mode:multiply"
							></div>
							<div
								style="position:relative;width:120mm;height:150mm;border-radius:60mm 60mm 0 0;overflow:hidden;box-shadow:0 30px 60px -20px rgba(42,20,12,.6)"
							>
								<img
									src="https://images.unsplash.com/photo-1556745753-b2904692b3cd?auto=format&fit=crop&w=900&q=75"
									alt=""
									style="width:100%;height:100%;object-fit:cover"
								/>
							</div>
							<img
								src="/wordmark-cream.png"
								alt="Tu Local Web"
								style="position:relative;height:11mm;margin-top:14mm"
							/>
							<div style="position:relative;margin-top:4mm;font-size:14px;color:#FBF7F2">
								Para los que venden en serio.
							</div>
						</div>
					</div></Lamina
				>
			</div>
			<div style="display:flex;flex-direction:column;gap:10px">
				<span style="font-size:14px;color:#6B5244">Interior · bolsillo con ranura para tarjeta</span
				>
				<Lamina
					nombre="papeleria-6-carpeta-interior"
					ancho={869.29}
					alto={1171.65}
					mm={[230, 310]}
					escala={0.62}
					><div
						style="width:230mm;height:310mm;background:#FBF7F2;position:relative;overflow:hidden;box-shadow:0 30px 60px -30px rgba(42,20,12,.45);"
					>
						<div style="padding:22mm">
							<img src="/logo.svg" alt="" style="width:14mm" />
							<div
								style="margin-top:8mm;font-size:22px;line-height:1.15;letter-spacing:-.03em;font-weight:600;max-width:150mm"
							>
								Todo lo que necesitás para vender, en una sola cuenta.
							</div>
							<div
								style="margin-top:8mm;display:grid;grid-template-columns:1fr 1fr;gap:4mm 10mm;font-size:12px;line-height:1.5;color:#4A3428"
							>
								<span>✓ Tienda online + punto de venta</span><span>✓ Tincho, asistente con IA</span
								><span>✓ Mercado Libre y WhatsApp</span><span>✓ Facturación con ARCA</span><span
									>✓ Envíos con Andreani y OCA</span
								><span>✓ Finanzas con Caja</span>
							</div>
						</div>
						<div
							style="position:absolute;left:0;right:0;bottom:0;height:115mm;background:#D2582F;clip-path:polygon(0 18%,100% 0,100% 100%,0 100%)"
						>
							<div
								style="position:absolute;inset:0;background:url(/textura/arch-pattern-accent.svg) center/30mm;opacity:.6"
							></div>
							<div
								style="position:absolute;left:24mm;top:40mm;width:90mm;height:55mm;border-radius:3mm;border:1.5px dashed rgba(255,255,255,.7);display:grid;place-items:center;color:#fff;font-size:12px"
							>
								Ranura tarjeta 85 × 55
							</div>
							<img
								src="/hornero/portapapeles.png"
								alt=""
								style="position:absolute;right:16mm;bottom:12mm;width:60mm;height:60mm;object-fit:contain"
							/>
						</div>
					</div></Lamina
				>
			</div>
		</div>
	</div>

	<!-- CREDENCIALES -->
	<div style="display:flex;flex-direction:column;gap:20px">
		<h2 style="margin:0;font-size:26px;letter-spacing:-.03em;font-weight:600">
			Credenciales · 54 × 86 mm vertical
		</h2>
		<Campos
			titulo="El evento"
			ayuda="Para la credencial de visitante; el nombre y el cargo salen de tus&nbsp;datos."
			campos={CAMPOS_EVENTO}
			bind:valores={evento}
		/>
		<div style="display:flex;flex-wrap:wrap;gap:56px;align-items:flex-start">
			<div id="7a" style="display:flex;flex-direction:column;gap:14px">
				<div style="display:flex;align-items:center;gap:10px">
					<a
						href="#7a"
						style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
						>7a</a
					><span style="font-size:14px;color:#6B5244">Equipo · foto en arco</span>
				</div>
				<Lamina nombre="papeleria-7a-credencial" ancho={204.09} alto={325.04} mm={[54, 86]}
					><div
						style="width:54mm;height:86mm;border-radius:3mm;background:#FBF7F2;position:relative;overflow:hidden;display:flex;flex-direction:column;align-items:center;padding:9mm 5mm 5mm;box-shadow:0 20px 40px -20px rgba(42,20,12,.5)"
					>
						<div
							style="position:absolute;top:3mm;left:50%;transform:translateX(-50%);width:12mm;height:2.6mm;border-radius:2mm;background:#E7DED3"
						></div>
						<div
							style="position:absolute;left:0;right:0;top:0;height:34mm;background:#D2582F url(/textura/arch-pattern-accent.svg) center/14mm"
						></div>
						<div
							style="position:relative;width:28mm;height:34mm;border-radius:14mm 14mm 2mm 2mm;overflow:hidden;box-shadow:0 0 0 1.4mm #FBF7F2"
						>
							{#if persona.foto}<img
									src={persona.foto}
									alt=""
									style="width:100%;height:100%;object-fit:cover"
								/>{:else}<span
									style="width:100%;height:100%;display:grid;place-items:center;background:#F3E9DF;color:#C24E25;font-weight:600;font-size:34px"
									>{iniciales(persona.nombre)}</span
								>{/if}
						</div>
						<div style="margin-top:4mm;font-size:13px;font-weight:600;text-align:center">
							{persona.nombre}
						</div>
						<div style="font-size:8px;color:#7A6252">{persona.cargo}</div>
						<div
							style="margin-top:auto;width:100%;display:flex;justify-content:space-between;align-items:center"
						>
							<img src="/wordmark-brown.png" alt="Tu Local Web" style="height:3mm" /><span
								style="font-size:7px;color:#9C8676">ID 0042</span
							>
						</div>
					</div></Lamina
				>
			</div>
			<div id="7b" style="display:flex;flex-direction:column;gap:14px">
				<div style="display:flex;align-items:center;gap:10px">
					<a
						href="#7b"
						style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
						>7b</a
					><span style="font-size:14px;color:#6B5244">Evento · visitante con QR de acceso</span>
				</div>
				<Lamina nombre="papeleria-7b-credencial-evento" ancho={204.09} alto={325.04} mm={[54, 86]}
					><div
						style="width:54mm;height:86mm;border-radius:3mm;background:#2A140C;color:#FBF7F2;position:relative;overflow:hidden;display:flex;flex-direction:column;padding:9mm 5mm 5mm;box-shadow:0 20px 40px -20px rgba(42,20,12,.6)"
					>
						<div
							style="position:absolute;top:3mm;left:50%;transform:translateX(-50%);width:12mm;height:2.6mm;border-radius:2mm;background:#4A3428"
						></div>
						<img
							src="/wordmark-cream.png"
							alt="Tu Local Web"
							style="height:3.2mm;align-self:flex-start"
						/>
						<div
							style="margin-top:3mm;font-size:7.5px;letter-spacing:.12em;text-transform:uppercase;font-weight:700;color:#E8875F"
						>
							{evento.titulo} · {evento.fecha}
						</div>
						<div
							style="margin-top:5mm;font-size:15px;line-height:1.05;font-weight:600;letter-spacing:-.02em"
						>
							{nombre1}<br />{nombre2}
						</div>
						<div style="font-size:8px;color:#BDB2A8;margin-top:1mm">{persona.cargo}</div>
						<div style="margin-top:auto;display:flex;justify-content:space-between;align-items:end">
							<Qr
								texto={evento.link}
								alt="QR · {evento.titulo}"
								style="width:17mm;height:17mm;border-radius:1.5mm;display:block;flex:none;background:#FBF7F2"
							/><span
								style="padding:1.2mm 2.4mm;border-radius:999px;background:#D2582F;font-size:7.5px;font-weight:700"
								>VISITANTE</span
							>
						</div>
						<img
							src="/hornero/saludo.png"
							alt=""
							style="position:absolute;right:-5mm;top:22mm;width:26mm;height:26mm;object-fit:contain;opacity:.95"
						/>
					</div></Lamina
				>
			</div>
		</div>
	</div>

	<!-- STICKERS -->
	<div id="8" style="display:flex;flex-direction:column;gap:20px">
		<div style="display:flex;align-items:center;gap:12px">
			<a
				href="#8"
				style="padding:4px 10px;border-radius:999px;background:#2A140C;color:#fff;font-size:12px;font-weight:700;text-decoration:none"
				>8</a
			>
			<h2 style="margin:0;font-size:26px;letter-spacing:-.03em;font-weight:600">
				Plancha de stickers · A5 troquelada
			</h2>
		</div>
		<Lamina nombre="papeleria-8-stickers" ancho={793.7} alto={559.37} mm={[210, 148]}
			><div
				style="width:210mm;height:148mm;background:#fff;border-radius:2mm;position:relative;padding:10mm;display:grid;grid-template-columns:repeat(4,1fr);grid-template-rows:1fr 1fr;gap:6mm;place-items:center;box-shadow:0 24px 50px -26px rgba(42,20,12,.45)"
			>
				<div
					style="width:38mm;height:30mm;border-radius:19mm 19mm 2mm 2mm;background:#D2582F;display:grid;place-items:center;box-shadow:0 0 0 1.6mm #fff,0 0 0 2mm #E7DED3"
				>
					<img src="/logo.svg" alt="" style="width:20mm;filter:brightness(0) invert(1)" />
				</div>
				<div
					style="width:40mm;height:40mm;border-radius:50%;background:#FBF7F2;display:grid;place-items:center;box-shadow:0 0 0 1.6mm #fff,0 0 0 2mm #E7DED3"
				>
					<img
						src="/hornero/celebrando.png"
						alt=""
						style="width:34mm;height:34mm;object-fit:contain"
					/>
				</div>
				<div
					style="padding:4mm 6mm;border-radius:999px;background:#2A140C;color:#FBF7F2;font-size:14px;font-weight:600;white-space:nowrap;box-shadow:0 0 0 1.6mm #fff,0 0 0 2mm #E7DED3"
				>
					Hecho con <span style="color:#E8875F">Tu Local Web</span>
				</div>
				<div
					style="width:36mm;height:36mm;border-radius:50%;background:#2F8F63;color:#fff;display:flex;flex-direction:column;align-items:center;justify-content:center;gap:1mm;box-shadow:0 0 0 1.6mm #fff,0 0 0 2mm #E7DED3"
				>
					<Icon icon={I.phCheckBold} size={30} /><span style="font-size:11px;font-weight:700"
						>Pago aprobado</span
					>
				</div>
				<div
					style="padding:4mm 6mm;border-radius:3mm;background:#FCEDE5;color:#C24E25;font-size:13px;font-weight:700;display:flex;align-items:center;gap:2mm;white-space:nowrap;box-shadow:0 0 0 1.6mm #fff,0 0 0 2mm #E7DED3"
				>
					<Icon icon={I.phTruckFill} size={18} />Envío gratis
				</div>
				<div
					style="width:40mm;height:40mm;border-radius:50%;background:#D2582F;display:grid;place-items:center;box-shadow:0 0 0 1.6mm #fff,0 0 0 2mm #E7DED3"
				>
					<img
						src="/hornero/whatsapp.png"
						alt=""
						style="width:34mm;height:34mm;object-fit:contain"
					/>
				</div>
				<div
					style="padding:4mm 6mm;border-radius:999px;background:#FBF7F2;font-size:13px;font-weight:600;white-space:nowrap;box-shadow:0 0 0 1.6mm #fff,0 0 0 2mm #E7DED3,inset 0 0 0 1px #E6D9CC"
				>
					Gracias por comprar local ♥
				</div>
				<div
					style="width:38mm;height:30mm;border-radius:19mm 19mm 2mm 2mm;background:#2A140C;display:flex;flex-direction:column;align-items:center;justify-content:center;gap:1.5mm;box-shadow:0 0 0 1.6mm #fff,0 0 0 2mm #E7DED3"
				>
					<span style="font-size:20px;font-weight:700;color:#E8875F;letter-spacing:-.03em">0%</span
					><span style="font-size:9px;color:#FBF7F2">de comisión</span>
				</div>
			</div></Lamina
		>
	</div>
</section>
