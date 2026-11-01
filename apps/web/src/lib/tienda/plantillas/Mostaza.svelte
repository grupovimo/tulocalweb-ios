<script lang="ts">
	import { camposFactura, OPCIONALES } from '#lib/tienda/factura.ts';
	import { envia } from '../envio.ts';
	import { enhance } from '$app/forms';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { untrack } from 'svelte';
	import { Reserva } from '../reserva.svelte.ts';
	import { accionesDe, AltaPlan, cajaDe, textosSuscripcion } from '../suscripcion.svelte.ts';
	import { CADA, diaCobro, entregasDe, porPeriodo, queIncluye } from '../suscripciones.ts';
	import { cuandoLargo, duracion } from '#lib/turnos/formato.ts';
	import Icon from '#lib/components/Icon.svelte';
	import LogoTienda from '../LogoTienda.svelte';
	import VacioTienda from '../VacioTienda.svelte';
	import { enOrden } from '../secciones.ts';
	import * as I from '#lib/icons.ts';
	import FormResena from '../FormResena.svelte';
	import CompartirProducto from '../CompartirProducto.svelte';
	import LanzamientoFicha from '../LanzamientoFicha.svelte';
	import SenaPedido from '../SenaPedido.svelte';
	import { Conteo } from '../conteo.svelte.ts';
	import PreguntasProducto from '../PreguntasProducto.svelte';
	import type { TemaOpiniones } from '../opiniones.ts';
	import { carritoDe } from '../carrito.svelte.ts';
	import CampoCupon from '../CampoCupon.svelte';
	import { CuponCheckout, totales } from '../cupon.svelte.ts';
	import EnvioCorreo from '../EnvioCorreo.svelte';
	import {
		comoLlega,
		cotizar,
		cotizarYa,
		cpValido,
		resumenMoto,
		textoEnvios
	} from '../envio.svelte.ts';
	import { Seleccion, tarjetaDe } from '../seleccion.svelte.ts';
	import { avisoMayorista } from '../lineas.ts';
	import { lineaPorMayor, textoAviso } from '../mayorista.ts';
	import { partirTitulo } from '../titulo.ts';
	import type { OpcionEnvio, ProductoPublico, TiendaPublica } from '../tipos.ts';
	import { AltaNewsletter } from '../newsletter.svelte.ts';

	/**
	 * Plantilla Mostaza (design/Plantilla Mostaza.dc.html), con los datos de la tienda en lugar de los
	 * de ejemplo. El catálogo es el "Menú y pedido" del diseño (se suma desde la carta y el pedido se
	 * arma al costado). Lo que el diseño muestra sin datos atrás (cuenta regresiva, horarios, turnos de
	 * retiro, propina, mesa con QR, molienda y tamaños, club, seguidores) no se dibuja. La suscripción
	 * sale de los planes de la tienda (§51): la tarjeta «para tu casa» de la portada y la página de cada
	 * plan con las piezas del producto; sin planes, no se dibuja.
	 */
	type Pagina =
		| 'inicio'
		| 'catalogo'
		| 'producto'
		| 'carrito'
		| 'checkout'
		| 'nosotros'
		| 'no-existe'
		| 'listo'
		| 'turnos'
		| 'turno'
		| 'suscripcion'
		| 'mi-suscripcion';
	// eslint-disable-next-line svelte/no-unused-props -- `proximos`: esta plantilla no dibuja el equipo ni los libres de hoy
	let {
		tienda,
		pagina,
		producto = null,
		categoria = null,
		orden = 'vendidos',
		pedido = null,
		turno = null,
		aviso = null,
		plan = null,
		suscripcion = null,
		demo = false,
		error = null
	}: {
		tienda: TiendaPublica;
		pagina: Pagina;
		producto?: ProductoPublico | null;
		categoria?: string | null;
		orden?: string;
		pedido?: {
			numero: number;
			total: number;
			pago: string;
			nombre: string;
			alias: string | null;
			esCbu?: boolean;
			/** El código del pedido: con él, la reseña queda con compra. */
			codigo?: string;
			/** La reseña se pide recién cuando lo recibió (Entregado o Retirado). */
			estado?: string;
		} | null;
		/** La confirmación de un turno (pagina 'turno'). */
		turno?: import('../tipos.ts').TurnoPublico | null;
		/** Los libres de hoy y el próximo de cada profesional (portada). */
		proximos?: import('../tipos.ts').ProximosTurnos | null;
		/** Vuelta de Mercado Pago al pagar la seña: ok, pendiente o error. */
		aviso?: string | null;
		/** El plan de suscripción (pagina 'suscripcion', §51). */
		plan?: import('../tipos.ts').PlanPublico | null;
		/** La suscripción del link privado (pagina 'mi-suscripcion'). */
		suscripcion?: import('../tipos.ts').SuscripcionPublica | null;
		/** Es la demo: suscribirse no cobra (se dice al lado del botón). */
		demo?: boolean;
		error?: string | null;
	} = $props();

	// En el dominio propio de la tienda los links no llevan /t/<slug> (tienda.base = '').
	const base = $derived(tienda.base ?? `/t/${tienda.slug}`);
	// La cuenta regresiva (§52): sólo un lanzamiento próximo o una oferta con fin, nunca inventada.
	const conteo = new Conteo(() => tienda);
	const carrito = $derived(carritoDe(tienda.slug));
	const nl = new AltaNewsletter(() => tienda.slug);
	let ancho = $state(page.data.movil ? 390 : 1200);
	const nar = $derived(ancho < 820);
	const wide = $derived(!nar);

	const fmt = (n: number) => '$ ' + Math.round(n).toLocaleString('es-AR');
	const img = (f: string | null, w: number, h?: number) =>
		!f
			? null
			: /^https?:\/\//.test(f)
				? f
				: `https://images.unsplash.com/photo-${f}?auto=format&fit=crop&w=${w}${h ? '&h=' + h : ''}&q=72`;
	const porId = $derived(new Map(tienda.productos.map((p) => [p.id, p])));
	// Etiquetas con los colores del diseño ("Más pedido", "Vegano", "Nuevo").
	const tagColor = (p: ProductoPublico) =>
		tarjetaDe(p).agotado
			? ['#3B2416', '#FBF3E4']
			: p.anterior
				? ['#F5C26B', '#2A170C']
				: ['#E8F2E2', '#2F6B3A'];
	// Con variantes de precios distintos, "Desde"; "Sin stock" sólo si no queda ninguna.
	const tarjeta = (p: ProductoPublico) => {
		const t = tarjetaDe(p);
		return {
			id: p.id,
			n: p.nombre,
			d: p.descripcion,
			price: (t.desde ? 'Desde ' : '') + fmt(t.precio),
			old: p.anterior && !t.desde ? fmt(p.anterior) : null,
			img: img(p.foto, 500, 500),
			tag: t.agotado ? 'Sin stock' : p.etiqueta,
			tagBg: tagColor(p)[0],
			tagFg: tagColor(p)[1],
			stock: p.stock,
			agotado: t.agotado,
			href: `${base}/producto/${p.slug ?? p.id}`
		};
	};
	type Tarjeta = ReturnType<typeof tarjeta>;
	const ver = (s: string) => tienda.secciones.includes(s as TiendaPublica['secciones'][number]);
	const whatsapp = $derived(tienda.whatsapp.replace(/\D/g, ''));
	const wa = (texto: string) =>
		`https://wa.me/${whatsapp.startsWith('54') ? whatsapp : '54' + whatsapp}?text=${encodeURIComponent(texto)}`;
	// Las líneas con variante no se tocan desde el menú: ahí se elige en la ficha.
	const enCarrito = (id: string) =>
		carrito.lineas.find((l) => l.id === id && !l.varianteId)?.cantidad ?? 0;
	/** Sumar desde el menú o "¿Le sumás algo?": con opciones, primero se elige en la ficha. */
	const sumar = (id: string) => {
		const p = porId.get(id);
		if (!p || tienda.vitrina) return;
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- ruta de la tienda con su slug
		if (!carrito.sumarProducto(p)) goto(`${base}/producto/${p.slug ?? p.id}`);
	};

	// Título de la portada: el último tramo va en mostaza, como en el diseño ("Café rico, / sin vueltas.").
	const titulo = $derived(partirTitulo(tienda.portada.titulo));
	// Sin logo, la inicial en el círculo de la marca (nada que sugiera un rubro que la tienda no eligió).
	const inicial = $derived(tienda.nombre.trim().charAt(0).toUpperCase());

	// Las tres formas de recibirlo del diseño, con los envíos que la tienda tiene de verdad.
	const modes = $derived([
		...(tienda.envios.retiro
			? [
					{
						t: 'Retirá en el local',
						d: tienda.envios.retiroDireccion || tienda.ciudad || 'Sin costo',
						i: I.phStorefrontFill,
						bg: '#F1E2C8',
						href: `${base}/catalogo`
					}
				]
			: []),
		...(tienda.envios.moto
			? [
					{
						t: 'Envío en moto',
						d: [resumenMoto(tienda.envios), tienda.envios.motoDemora].filter(Boolean).join(' · '),
						i: I.phMopedFill,
						bg: '#F6E7CC',
						href: `${base}/catalogo`
					}
				]
			: []),
		...(tienda.envios.correo
			? [
					{
						t: 'Envío a todo el país',
						d: tienda.envios.gratisDesde
							? `Gratis desde ${fmt(tienda.envios.gratisDesde)}`
							: 'Por Correo Argentino',
						i: I.phPackageFill,
						bg: '#EFDDBF',
						href: `${base}/catalogo`
					}
				]
			: []),
		...(tienda.suscripciones.length
			? [
					{
						t: 'Para tu casa',
						d:
							tienda.suscripciones.length === 1
								? `${tienda.suscripciones[0].nombre}, con suscripción`
								: `${tienda.suscripciones.length} planes, con suscripción`,
						i: I.phArrowsClockwiseBold,
						bg: '#EFDDBF',
						href: `${base}/suscripciones/${tienda.suscripciones[0].id}`
					}
				]
			: [])
	]);
	const best = $derived(
		tienda.masVendidos.slice(0, 4).flatMap((id) => (porId.get(id) ? [tarjeta(porId.get(id)!)] : []))
	);
	// El bloque "Combo" del diseño, con las ofertas reales (precio tachado), la de mayor descuento primero.
	const ofertas = $derived(
		tienda.productos
			.filter((p) => p.anterior && !tarjetaDe(p).agotado)
			.sort((a, b) => a.precio / a.anterior! - b.precio / b.anterior!)
			.slice(0, 3)
	);
	const oferta = $derived(ofertas[0] ?? null);
	const fotoGrande = $derived(img(tienda.portada.foto, 900) ?? best[0]?.img ?? null);
	const fotoChica = $derived(
		tienda.fotos[0] ? img(tienda.fotos[0], 600, 600) : (best[1]?.img ?? null)
	);

	// Catálogo (el menú del diseño): categoría, orden y de a 24, agrupado por categoría.
	const ORDENES = [
		['vendidos', 'Más vendidos'],
		['menor', 'Menor precio'],
		['mayor', 'Mayor precio'],
		['az', 'De la A a la Z']
	] as const;
	const listado = $derived.by(() => {
		const l = tienda.productos.filter((p) => !categoria || p.categoria === categoria);
		const rank = new Map(tienda.masVendidos.map((id, i) => [id, i]));
		const f =
			{
				vendidos: (a: ProductoPublico, b: ProductoPublico) =>
					(rank.get(a.id) ?? 999) - (rank.get(b.id) ?? 999),
				menor: (a: ProductoPublico, b: ProductoPublico) => a.precio - b.precio,
				mayor: (a: ProductoPublico, b: ProductoPublico) => b.precio - a.precio,
				az: (a: ProductoPublico, b: ProductoPublico) => a.nombre.localeCompare(b.nombre, 'es')
			}[orden as 'vendidos'] ?? (() => 0);
		return [...l].sort(f);
	});
	let mostrar = $state(24);
	const grupos = $derived.by(() => {
		// eslint-disable-next-line svelte/prefer-svelte-reactivity -- local, sólo agrupa
		const g = new Map<string, Tarjeta[]>();
		for (const p of listado.slice(0, mostrar)) {
			const k = categoria ? '' : (p.categoria ?? 'Otros');
			g.set(k, [...(g.get(k) ?? []), tarjeta(p)]);
		}
		return [...g].map(([t, items]) => ({ t, items }));
	});
	const enlaceCat = (c: string | null, o = orden) => {
		// eslint-disable-next-line svelte/prefer-svelte-reactivity -- local, sólo arma el link
		const q = new URLSearchParams();
		if (c) q.set('cat', c);
		if (o !== 'vendidos') q.set('orden', o);
		return `${base}/catalogo${q.size ? '?' + q : ''}`;
	};

	// Ficha.
	let abierto = $state<string>('Descripción');
	const acc = $derived(
		producto
			? [
					['Descripción', producto.descripcion || `${producto.nombre}, de ${tienda.nombre}.`],
					['Envíos', textoEnvios(tienda.envios)],
					[
						'Cambios y devoluciones',
						`${tienda.politica ? `${tienda.politica} ` : ''}Tenés 10 días desde que lo recibís para arrepentirte de la compra, sin dar explicaciones (botón de arrepentimiento, abajo de todo).`
					]
				].map(([t, d]) => ({
					t,
					d,
					open: abierto === t,
					i: abierto === t ? I.phMinusBold : I.phPlusBold
				}))
			: []
	);
	const related = $derived(
		producto
			? tienda.productos
					.filter(
						(p) =>
							p.id !== producto!.id && p.categoria === producto!.categoria && !tarjetaDe(p).agotado
					)
					.slice(0, 4)
					.map(tarjeta)
			: []
	);
	const resenasDe = $derived(
		producto ? tienda.resenas.filter((r) => r.producto === producto!.nombre).length : 0
	);
	let cantidad = $state(1);
	// Tamaño, molienda, presentación…: lo elegido, la variante, su precio, stock y foto (seleccion.svelte.ts).
	const sel = $derived(producto ? new Seleccion(producto) : null);
	// El tachado no va con "Desde": no se sabe a qué presentación se compara.
	const tachado = $derived(sel && !sel.desde ? sel.anterior : null);
	let cp = $state('');
	let cotizado = $state<OpcionEnvio[] | null>(null);
	async function calcular() {
		if (cpValido(cp)) cotizado = await cotizarYa(cp, sel?.precio ?? subtotal, tienda);
	}
	function agregar() {
		if (!sel || tienda.vitrina || !sel.agregar(carrito, cantidad)) return;
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- ruta de la tienda con su slug
		goto(`${base}/carrito`);
	}
	function sumarOferta() {
		if (!oferta || tienda.vitrina) return;
		// Con opciones, la oferta lleva a la ficha para elegir.
		const a = carrito.sumarProducto(oferta)
			? `${base}/carrito`
			: `${base}/producto/${oferta.slug ?? oferta.id}`;
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- ruta de la tienda con su slug
		goto(a);
	}

	// Carrito.
	const lineas = $derived(carrito.detalle(tienda.productos, tienda.mayorista ?? null));
	const subtotal = $derived(lineas.reduce((s, l) => s + l.subtotal, 0));
	const falta = $derived(
		tienda.envios.gratisDesde ? Math.max(0, tienda.envios.gratisDesde - subtotal) : 0
	);
	// Venta por mayor (§49): cuánto falta para el precio mayorista, o lo que ya se ahorra.
	const porMayor = $derived(avisoMayorista(lineas, tienda.mayorista ?? null));
	const avisoMayor = $derived(porMayor && textoAviso(porMayor, fmt));
	// «Por mayor: $ X desde 12 unidades» en la ficha.
	const porMayorFicha = $derived(sel ? lineaPorMayor(sel.mayorista, tienda.mayorista, fmt) : null);
	const upsell = $derived(
		tienda.masVendidos
			.filter(
				(id) =>
					!carrito.lineas.some((l) => l.id === id) &&
					porId.has(id) &&
					!tarjetaDe(porId.get(id)!).agotado
			)
			.slice(0, 3)
			.map((id) => porId.get(id)!)
	);
	// El pedido armado, para mandarlo por WhatsApp desde el menú.
	const textoPedido = $derived(
		`¡Hola! Quiero pedir:\n${lineas.map((l) => `${l.cantidad} × ${l.producto.nombre}${l.variante ? ` (${l.variante})` : ''} (${fmt(l.subtotal)})`).join('\n')}\nTotal: ${fmt(subtotal)}`
	);

	// Checkout.
	let envio = $state<OpcionEnvio['id'] | null>(null);
	let pago = $state<'transferencia' | 'mercadopago'>('transferencia');
	let enviando = $state(false);
	const opciones = $derived(
		cpValido(cp)
			? cotizar(cp, subtotal, tienda)
			: tienda.envios.retiro
				? cotizar('0000', subtotal, tienda)
				: []
	);
	const elegida = $derived(opciones.find((o) => o.id === envio) ?? null);
	const costoEnvio = $derived(elegida?.precio ?? 0);
	const cupon = new CuponCheckout(() => tienda.slug);
	const cuenta = $derived(
		totales({
			subtotal,
			envio: costoEnvio,
			pago,
			transferencia: tienda.cobros.transferencia,
			cupon
		})
	);
	const descuento = $derived(cuenta.transf);
	const total = $derived(cuenta.total);
	$effect(() => {
		if (envio && !opciones.some((o) => o.id === envio)) envio = null;
	});
	const pagos = $derived([
		...(tienda.cobros.datosTransferencia
			? [
					{
						id: 'transferencia' as const,
						t: 'Transferencia',
						d: tienda.cobros.transferencia
							? `${tienda.cobros.transferencia}% off · te mandamos el alias al confirmar`
							: 'Te mandamos el alias al confirmar',
						i: I.phBankFill,
						c: '#2F6B3A'
					}
				]
			: []),
		...(tienda.cobros.mercadoPago
			? [
					{
						id: 'mercadopago' as const,
						t: 'Mercado Pago',
						d:
							tienda.cobros.cuotas > 1
								? `Tarjetas en hasta ${tienda.cobros.cuotas} cuotas sin interés, dinero en cuenta`
								: 'Tarjetas de crédito y débito, y dinero en cuenta',
						i: I.siMercadopago,
						c: '#00B1EA'
					}
				]
			: [])
	]);
	// Sin datos de transferencia cargados, el medio de pago es el primero disponible.
	$effect(() => {
		if (pagos.length && !pagos.some((x) => x.id === pago)) pago = pagos[0].id;
	});

	const foot = $derived([
		{
			t: 'Tienda',
			l: [
				{ t: 'Todo', href: `${base}/catalogo` },
				...tienda.categorias.slice(0, 4).map((c) => ({ t: c.nombre, href: enlaceCat(c.nombre) }))
			]
		},
		{
			t: 'Ayuda',
			l: [
				{ t: 'Envíos', href: `${base}/nosotros` },
				{ t: 'Cambios y devoluciones', href: `/t/${tienda.slug}/arrepentimiento` },
				{ t: 'Medios de pago', href: `${base}/nosotros` }
			]
		},
		{
			t: tienda.nombre,
			l: [
				{ t: 'Nosotros', href: `${base}/nosotros` },
				...(whatsapp ? [{ t: 'WhatsApp', href: wa('¡Hola!') }] : [])
			]
		}
	]);
	const nav = $derived([
		...(tienda.turnos
			? [
					{
						t: 'Reservar turno',
						href: `${base}/turnos`,
						on: pagina === 'turnos' || pagina === 'turno'
					}
				]
			: []),
		{ t: 'Menú', href: `${base}/catalogo`, on: pagina === 'catalogo' && !categoria },
		...tienda.categorias.slice(0, 3).map((c) => ({
			t: c.nombre,
			href: enlaceCat(c.nombre),
			on: pagina === 'catalogo' && categoria === c.nombre
		})),
		{ t: 'Nosotros', href: `${base}/nosotros`, on: pagina === 'nosotros' }
	]);
	const estrellas = (n: number) =>
		Array.from({ length: 5 }, (_, i) =>
			n >= i + 1 ? I.phStarFill : n >= i + 0.5 ? I.phStarHalfFill : I.phStar
		);
	const B = "font-family:'Bricolage Grotesque',sans-serif";
	// ---- Turnos (no están en el diseño de esta plantilla: se arman con sus piezas) ----
	// Se arma una vez al entrar a /turnos (lo que venga en la URL queda elegido).
	const reserva = untrack(() =>
		pagina === 'turnos' && tienda.turnos
			? new Reserva(tienda.slug, tienda.turnos, {
					servicio: page.url.searchParams.get('servicio'),
					profesional: page.url.searchParams.get('profesional'),
					dia: page.url.searchParams.get('dia'),
					inicio: page.url.searchParams.get('inicio')
				})
			: null
	);
	const FRANJA = { manana: I.phSunHorizon, tarde: I.phSun, noche: I.phMoon } as const;
	const anticipacion = $derived.by(() => {
		const m = tienda.turnos?.anticipacionMin ?? 0;
		return m >= 120 && m % 60 === 0 ? `${m / 60} h` : `${m} min`;
	});
	let reservando = $state(false);
	let cancelando = $state(false);
	let seguro = $state(false);

	// ---- Suscripción (los planes de la tienda, §51) ----
	const planes = $derived(tienda.suscripciones);
	const alta = untrack(() => (pagina === 'suscripcion' && plan ? new AltaPlan(plan) : null));
	const caja = $derived(plan ? cajaDe(plan, tienda) : null);
	let cambiando = $state(false);
	let seguroBaja = $state(false);
	// Reseñas y preguntas con las piezas de Mostaza: Bricolage pesada, tarjetas blancas, píldoras.
	const tema: TemaOpiniones = {
		texto: '#3B2416',
		suave: '#7A5A44',
		campo: '#FBF3E4',
		tarjeta: '#fff',
		borde: 'rgba(59,36,22,.14)',
		acento: '#3B2416',
		sobreAcento: '#FBF3E4',
		estrella: '#C9821A',
		radio: '14px',
		radioBoton: '999px',
		fuente: "'Instrument Sans', sans-serif",
		titulos: "'Bricolage Grotesque', sans-serif",
		pesoTitulos: 800
	};
</script>

<svelte:window bind:innerWidth={ancho} />

<svelte:head>
	<link
		href="https://fonts.googleapis.com/css2?family=Bricolage+Grotesque:opsz,wght@12..96,500;12..96,700;12..96,800&display=swap"
		rel="stylesheet"
	/>
</svelte:head>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
<div
	class="mostaza"
	style="min-height:100vh;font-family:'Instrument Sans',system-ui,sans-serif;color:#3B2416;background:#FBF3E4"
>
	<div
		style="container-type:inline-size;width:100%;background:#FBF3E4;position:relative;overflow-x:clip"
	>
		<!-- barra promo -->
		<div
			style="display:flex;justify-content:center;align-items:center;gap:12px;flex-wrap:wrap;padding:9px 16px;background:var(--acento, #C9821A);color:#2A170C;font-size:13px;font-weight:600;text-align:center"
		>
			{#if conteo.actual}<span>{conteo.texto}</span><span
					style="display:flex;gap:4px;font-variant-numeric:tabular-nums"
					>{#each conteo.partes(3) as c (c.u)}<span
							style="padding:2px 7px;border-radius:6px;background:#2A170C;color:#FBF3E4"
							>{c.v}{c.u}</span
						>{/each}</span
				>{:else}<span
					>{tienda.envios.gratisDesde
						? `Envío gratis desde ${fmt(tienda.envios.gratisDesde)}`
						: tienda.cobros.cuotas > 1
							? `${tienda.cobros.cuotas} cuotas sin interés`
							: 'Compra protegida'}</span
				>{/if}{#if wide && tienda.cobros.transferencia > 0}<span
					style="padding:2px 7px;border-radius:6px;background:#2A170C;color:#FBF3E4"
					>{tienda.cobros.transferencia}% off por transferencia</span
				>{/if}
		</div>

		<!-- header -->
		<header
			style="position:sticky;top:0;z-index:20;display:flex;align-items:center;gap:16px;padding:12px clamp(14px,3.5cqw,40px);background:rgba(251,243,228,.92);backdrop-filter:blur(10px)"
		>
			<a
				href={base || '/'}
				style="display:flex;align-items:center;gap:9px;min-width:0;text-decoration:none;color:#3B2416"
				>{#if !tienda.logo}<span
						style="width:38px;height:38px;border-radius:50%;background:#3B2416;color:#F5C26B;display:grid;place-items:center;flex:none;{B};font-weight:800;font-size:20px;letter-spacing:-.03em"
						>{inicial}</span
					>{/if}<span
					style="{B};font-weight:800;font-size:23px;letter-spacing:-.03em;white-space:nowrap;overflow:hidden;text-overflow:ellipsis"
					><LogoTienda {tienda} /></span
				></a
			>
			{#if wide}<nav
					style="display:flex;gap:4px;margin-left:14px;min-width:0;overflow-x:auto;scrollbar-width:none"
				>
					{#each nav as n (n.t)}<a
							href={n.href}
							style="text-decoration:none;font:600 14px 'Instrument Sans';padding:9px 14px;border-radius:999px;background:{n.on
								? '#F1E2C8'
								: 'transparent'};color:#3B2416;white-space:nowrap">{n.t}</a
						>{/each}
				</nav>{/if}
			<span style="flex:1"></span>
			<a
				href="{base}/carrito"
				aria-label="Carrito"
				style="display:flex;align-items:center;gap:8px;flex:none;text-decoration:none;padding:10px 16px;border-radius:999px;background:#3B2416;color:#FBF3E4;font:700 14px 'Instrument Sans';white-space:nowrap"
				><Icon icon={I.phBagSimpleFill} size={17} />{carrito.unidades}{#if wide && subtotal}<span
						style="opacity:.7;font-weight:500">· {fmt(subtotal)}</span
					>{/if}</a
			>
			{#if nar}<a
					href="{base}/catalogo"
					aria-label="Menú"
					style="display:grid;flex:none;color:#3B2416"><Icon icon={I.phListBold} size={24} /></a
				>{/if}{#if nar && tienda.turnos}<a
					href="{base}/turnos"
					aria-label="Reservar turno"
					style="display:grid;flex:none;color:#3B2416"
					><Icon icon={I.phCalendarPlusBold} size={24} /></a
				>{/if}
		</header>

		{#if tienda.vitrina}
			<div
				style="padding:10px 16px;background:#F1E2C8;font-size:13px;font-weight:600;text-align:center"
			>
				Es una tienda de ejemplo de la plantilla Mostaza: podés recorrerla, pero no se compra
				de&nbsp;verdad.
			</div>
		{/if}

		<!-- ============ INICIO ============ -->
		{#if pagina === 'inicio'}
			<!-- Las secciones en el orden que eligió la tienda en Mi tienda. -->
			{#each enOrden(tienda.secciones, ['hero', 'prods', 'revs', 'ig']) as seccion (seccion)}
				{#if seccion === 'hero'}
					{#if ver('hero')}
						<section
							style="padding:clamp(16px,3cqw,40px) clamp(14px,3.5cqw,40px) 0;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,380px),1fr));gap:clamp(12px,2cqw,24px)"
						>
							<div
								style="padding:clamp(26px,4.5cqw,60px);border-radius:28px;background:#3B2416;color:#FBF3E4;display:flex;flex-direction:column;justify-content:space-between;gap:30px;min-height:min(560px,120cqw);min-width:0"
							>
								{#if tienda.ciudad}<span
										style="align-self:flex-start;display:flex;align-items:center;gap:8px;padding:7px 12px;border-radius:999px;background:rgba(245,194,107,.16);color:#F5C26B;font-size:13px;font-weight:700"
										><Icon icon={I.phMapPinFill} size={15} />{tienda.ciudad}</span
									>{/if}
								<div>
									<h1
										style="margin:0;{B};font-weight:800;font-size:clamp(52px,8.4cqw,112px);line-height:.88;letter-spacing:-.05em;overflow-wrap:break-word"
									>
										{titulo[0]}{#if titulo[1]}<br /><span style="color:#F5C26B">{titulo[1]}</span
											>{/if}
									</h1>
									{#if tienda.portada.lema}<p
											style="margin:18px 0 0;font-size:17px;line-height:1.5;color:#E6D3BC;max-width:440px"
										>
											{tienda.portada.lema}
										</p>{/if}
								</div>
								<div style="display:flex;gap:10px;flex-wrap:wrap">
									<a
										href="{base}/catalogo"
										style="display:flex;align-items:center;gap:8px;padding:15px 24px;border-radius:999px;background:#F5C26B;color:#2A170C;font:700 15px 'Instrument Sans';text-decoration:none"
										><Icon icon={I.phHandTapFill} size={17} />Pedir ahora</a
									>{#if tienda.categorias[0]}<a
											href={enlaceCat(tienda.categorias[0].nombre)}
											style="padding:15px 22px;border:1.5px solid rgba(251,243,228,.35);border-radius:999px;color:#FBF3E4;font:700 15px 'Instrument Sans';text-decoration:none"
											>{tienda.categorias[0].nombre}</a
										>{/if}
								</div>
							</div>
							<!-- Sin fotos ni reseñas no hay mosaico: la portada ocupa todo el ancho. -->
							{#if fotoGrande || fotoChica || tienda.promedio}<div
									style="display:grid;grid-template-columns:1fr 1fr;grid-template-rows:1fr 1fr;gap:clamp(10px,1.4cqw,16px);min-height:min(560px,110cqw)"
								>
									<div
										style="grid-row:span 2;border-radius:28px;overflow:hidden;background:#D9B98C;position:relative"
									>
										{#if fotoGrande}<img
												src={fotoGrande}
												alt=""
												style="position:absolute;inset:0;width:100%;height:100%;object-fit:cover"
											/>{/if}{#if !tienda.portada.foto && best[0]}<a
												href={best[0].href}
												style="position:absolute;left:14px;bottom:14px;max-width:calc(100% - 28px);padding:8px 12px;border-radius:14px;background:#FBF3E4;font-size:13px;font-weight:700;color:#3B2416;text-decoration:none"
												>{best[0].n} · {best[0].price}</a
											>{/if}
									</div>
									<div
										style="border-radius:28px;overflow:hidden;background:#D9B98C;position:relative;grid-row:{tienda.promedio
											? 'auto'
											: 'span 2'}"
									>
										{#if fotoChica}<img
												src={fotoChica}
												alt=""
												style="position:absolute;inset:0;width:100%;height:100%;object-fit:cover"
											/>{/if}
									</div>
									{#if tienda.promedio}
										<div
											style="border-radius:28px;background:var(--acento, #C9821A);color:#2A170C;padding:18px;display:flex;flex-direction:column;justify-content:space-between"
										>
											<Icon icon={I.phStarFill} size={24} />
											<div>
												<div
													style="{B};font-weight:800;font-size:clamp(34px,4cqw,48px);line-height:1;letter-spacing:-.04em"
												>
													{tienda.promedio.toLocaleString('es-AR')}
												</div>
												<div style="font-size:13px;font-weight:600;margin-top:4px">
													de promedio en reseñas
												</div>
											</div>
										</div>
									{/if}
								</div>{/if}
						</section>
					{/if}

					<section
						style="padding:clamp(12px,2cqw,24px) clamp(14px,3.5cqw,40px) 0;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,220px),1fr));gap:clamp(10px,1.4cqw,16px)"
					>
						{#each modes as m (m.t)}<a
								href={m.href}
								style="display:flex;align-items:center;gap:14px;padding:18px;border-radius:22px;background:{m.bg};color:#3B2416;text-align:left;text-decoration:none;font-family:inherit"
								><span
									style="width:48px;height:48px;border-radius:16px;background:#FBF3E4;display:grid;place-items:center;flex:none"
									><Icon icon={m.i} size={24} style="color:#C9821A" /></span
								><span style="flex:1;min-width:0"
									><span
										style="display:block;{B};font-size:19px;font-weight:700;letter-spacing:-.02em"
										>{m.t}</span
									><span style="display:block;font-size:13px;color:#7A5A44;margin-top:2px"
										>{m.d}</span
									></span
								><Icon icon={I.phArrowRightBold} size={17} /></a
							>{/each}
					</section>
				{:else if seccion === 'prods'}
					{#if ver('prods') && best.length}
						<section style="padding:clamp(48px,7cqw,90px) clamp(14px,3.5cqw,40px) 0">
							<div
								style="display:flex;justify-content:space-between;align-items:end;gap:16px;flex-wrap:wrap"
							>
								<h2
									style="margin:0;{B};font-weight:800;font-size:clamp(36px,5cqw,64px);line-height:.95;letter-spacing:-.045em"
								>
									Lo que más sale
								</h2>
								<a
									href="{base}/catalogo"
									style="text-decoration:none;padding:11px 18px;border-radius:999px;background:#F1E2C8;color:#3B2416;font:700 14px 'Instrument Sans'"
									>Ver menú completo</a
								>
							</div>
							<div
								style="margin-top:24px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(46%,230px),1fr));gap:clamp(10px,1.4cqw,16px)"
							>
								{#each best as p (p.id)}<div
										style="padding:10px;border-radius:24px;background:#fff;box-shadow:0 1px 2px rgba(59,36,22,.06);min-width:0"
									>
										<a
											href={p.href}
											class="zoom"
											style="position:relative;display:block;aspect-ratio:1;border-radius:18px;overflow:hidden;background:#E9D6B8"
											>{#if p.img}<img
													src={p.img}
													alt=""
													style="width:100%;height:100%;object-fit:cover"
												/>{/if}{#if p.tag}<span
													style="position:absolute;top:8px;left:8px;padding:4px 9px;border-radius:999px;background:#FBF3E4;color:#3B2416;font-size:11.5px;font-weight:700"
													>{p.tag}</span
												>{/if}</a
										>
										<div style="padding:12px 6px 4px">
											<a
												href={p.href}
												style="display:block;{B};font-size:18px;font-weight:700;letter-spacing:-.02em;color:#3B2416;text-decoration:none"
												>{p.n}</a
											>
											{#if p.d}<div
													class="dos"
													style="font-size:12.5px;color:#7A5A44;margin-top:2px;line-height:1.4"
												>
													{p.d}
												</div>{/if}
											<div
												style="margin-top:10px;display:flex;justify-content:space-between;align-items:center;gap:8px"
											>
												<span
													><b style="font-size:16px">{p.price}</b>{#if p.old}
														<s style="font-size:13px;color:#9C8270">{p.old}</s>{/if}</span
												><button
													type="button"
													onclick={() => sumar(p.id)}
													disabled={p.agotado || tienda.vitrina}
													aria-label="Sumar {p.n}"
													style="width:36px;height:36px;flex:none;border:0;border-radius:50%;background:#3B2416;color:#FBF3E4;display:grid;place-items:center;cursor:pointer"
													><Icon icon={I.phPlusBold} size={15} /></button
												>
											</div>
										</div>
									</div>{/each}
							</div>
						</section>

						{#if oferta}
							<section
								style="margin:clamp(48px,7cqw,90px) clamp(14px,3.5cqw,40px) 0;padding:clamp(24px,4cqw,48px);border-radius:28px;background:#F5C26B;color:#2A170C;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,320px),1fr));gap:28px;align-items:center"
							>
								<div>
									<div
										style="font-size:13px;font-weight:700;letter-spacing:.08em;text-transform:uppercase"
									>
										Oferta · −{Math.round((1 - oferta.precio / oferta.anterior!) * 100)}%
									</div>
									<h2
										style="margin:10px 0 0;{B};font-weight:800;font-size:clamp(38px,5.4cqw,70px);line-height:.92;letter-spacing:-.045em;overflow-wrap:break-word"
									>
										{oferta.nombre}
									</h2>
									<div
										style="margin-top:16px;display:flex;align-items:baseline;gap:12px;flex-wrap:wrap"
									>
										<b style="{B};font-size:40px;font-weight:800;letter-spacing:-.03em"
											>{fmt(oferta.precio)}</b
										><s style="font-size:17px;opacity:.6">{fmt(oferta.anterior!)}</s>
									</div>
									<button
										type="button"
										onclick={sumarOferta}
										disabled={tienda.vitrina}
										style="margin-top:18px;padding:15px 24px;border:0;border-radius:999px;background:#2A170C;color:#FBF3E4;font:700 15px 'Instrument Sans';cursor:pointer"
										>Sumar al pedido</button
									>
								</div>
								<div
									style="display:grid;grid-template-columns:repeat({ofertas.length},1fr);gap:10px"
								>
									{#each ofertas as o (o.id)}<a
											href="{base}/producto/{o.slug ?? o.id}"
											class="zoom"
											aria-label={o.nombre}
											style="display:block;aspect-ratio:3/4;border-radius:20px;overflow:hidden;background:#E3A94F"
											>{#if o.foto}<img
													src={img(o.foto, 300, 400)}
													alt=""
													style="width:100%;height:100%;object-fit:cover"
												/>{/if}</a
										>{/each}
								</div>
							</section>
						{/if}
					{/if}
				{:else if seccion === 'revs'}
					{#if ver('revs') && tienda.resenas.length}
						<section style="padding:clamp(48px,7cqw,90px) clamp(14px,3.5cqw,40px) 0">
							<h2
								style="margin:0;{B};font-weight:800;font-size:clamp(36px,5cqw,64px);line-height:.95;letter-spacing:-.045em"
							>
								Dicen en el barrio
							</h2>
							<div
								style="margin-top:24px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,280px),1fr));gap:14px"
							>
								{#each tienda.resenas.slice(0, 3) as r, i (i)}
									{@const pr = tienda.productos.find((p) => p.nombre === r.producto)}
									<div
										style="padding:14px;border-radius:24px;background:#fff;display:flex;flex-direction:column;gap:12px"
									>
										{#if pr?.foto}<div
												style="aspect-ratio:16/10;border-radius:16px;overflow:hidden;background:#E9D6B8"
											>
												<img
													src={img(pr.foto, 600, 375)}
													alt=""
													style="width:100%;height:100%;object-fit:cover"
												/>
											</div>{/if}
										<div style="display:flex;gap:2px;color:#C9821A">
											{#each estrellas(r.estrellas) as e, j (j)}<Icon icon={e} size={15} />{/each}
										</div>
										<p style="margin:0;font-size:15px;line-height:1.5">“{r.texto}”</p>
										<span style="font-size:13px;color:#7A5A44;font-weight:600"
											>{r.autor}{#if r.producto}
												· {r.conCompra ? 'compró' : 'sobre'} {r.producto}{/if}</span
										>
										{#if r.conCompra}<span
												style="align-self:flex-start;display:inline-flex;align-items:center;gap:5px;padding:5px 10px;border-radius:999px;background:#E7F0DF;font-size:12px;font-weight:700;color:#2F6B3A"
												><Icon icon={I.phSealCheckFill} size={13} />Compra verificada</span
											>{/if}
										{#if r.respuesta}<div
												style="padding:10px 12px;border-radius:14px;background:#F1E2C8;font-size:13.5px;line-height:1.5;color:#5E4332"
											>
												<b style="color:#3B2416">Respuesta de {tienda.nombre}</b><br />{r.respuesta}
											</div>{/if}
									</div>
								{/each}
							</div>
						</section>
					{/if}
				{:else if seccion === 'ig'}
					{#if ver('ig') && tienda.fotos.length >= 3}
						<section style="padding:clamp(48px,7cqw,90px) clamp(14px,3.5cqw,40px)">
							<h2
								style="margin:0;{B};font-weight:800;font-size:clamp(36px,5cqw,64px);line-height:.95;letter-spacing:-.045em"
							>
								De cerca
							</h2>
							<div
								style="margin-top:22px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(30%,170px),1fr));gap:8px"
							>
								{#each tienda.fotos as g (g)}<div
										style="aspect-ratio:1;border-radius:18px;overflow:hidden;background:#E9D6B8"
									>
										<img
											src={img(g, 400, 400)}
											alt=""
											style="width:100%;height:100%;object-fit:cover"
										/>
									</div>{/each}
							</div>
						</section>
					{/if}
				{/if}
			{/each}
			{#if !(ver('ig') && tienda.fotos.length >= 3)}
				<div style="height:clamp(48px,7cqw,90px)"></div>
			{/if}
		{/if}

		<!-- ============ CATÁLOGO (el "Menú y pedido" del diseño) ============ -->
		{#if pagina === 'catalogo' && !tienda.productos.length}
			<section style="padding:clamp(40px,7cqw,96px) clamp(16px,4cqw,48px)">
				<VacioTienda {tienda} {tema} motivo="sin-productos" />
			</section>
		{:else if pagina === 'catalogo'}
			<section style="padding:clamp(18px,3cqw,32px) clamp(14px,3.5cqw,40px) clamp(48px,7cqw,90px)">
				<div style="font-size:13px;color:#7A5A44">
					<a href={base || '/'} style="color:inherit;text-decoration:none">Inicio</a> /
					<span style="color:#3B2416">{categoria ?? 'Menú'}</span>
				</div>
				<div
					style="margin-top:10px;display:flex;justify-content:space-between;align-items:end;gap:16px;flex-wrap:wrap"
				>
					<h1
						style="margin:0;{B};font-weight:800;font-size:clamp(46px,7cqw,92px);line-height:.9;letter-spacing:-.05em;overflow-wrap:break-word;min-width:0"
					>
						{categoria ?? 'Menú'}
					</h1>
					<label
						style="display:flex;align-items:center;gap:6px;padding:10px 15px;border-radius:999px;border:1.5px solid rgba(59,36,22,.12);background:#fff;color:#3B2416;font-size:13.5px;font-weight:700;white-space:nowrap"
						>Ordenar<select
							value={orden}
							onchange={(e) => goto(enlaceCat(categoria, e.currentTarget.value), { replace: true })}
							style="border:0;background:transparent;font:700 13.5px 'Instrument Sans';color:#3B2416;cursor:pointer;padding:0"
							>{#each ORDENES as [k, t] (k)}<option value={k}>{t}</option>{/each}</select
						></label
					>
				</div>
				<div
					style="margin-top:12px;display:flex;align-items:center;gap:8px;font-size:13.5px;color:#7A5A44;flex-wrap:wrap"
				>
					<Icon icon={I.phStorefrontFill} size={15} /><b style="color:#3B2416"
						>{listado.length} {listado.length === 1 ? 'producto' : 'productos'}</b
					>{#if tienda.envios.retiro && tienda.direccion}<span>· Retirás en {tienda.direccion}</span
						>{/if}{#if tienda.envios.gratisDesde}<span
							>· Envío gratis desde {fmt(tienda.envios.gratisDesde)}</span
						>{/if}
				</div>
				<div
					style="position:sticky;top:62px;z-index:10;margin:18px calc(-1 * clamp(14px,3.5cqw,40px)) 0;padding:10px clamp(14px,3.5cqw,40px);background:rgba(251,243,228,.95);backdrop-filter:blur(8px);display:flex;gap:6px;overflow-x:auto;scrollbar-width:none"
				>
					{#each [null, ...tienda.categorias.map((c) => c.nombre)] as c (c ?? '')}{@const on =
							c === categoria}<a
							href={enlaceCat(c)}
							style="flex:none;text-decoration:none;padding:9px 15px;border-radius:999px;background:{on
								? '#3B2416'
								: '#F1E2C8'};color:{on
								? '#FBF3E4'
								: '#3B2416'};font-size:14px;font-weight:700;white-space:nowrap">{c ?? 'Todo'}</a
						>{/each}
				</div>
				<div style="margin-top:14px;display:flex;gap:clamp(20px,3cqw,36px);align-items:flex-start">
					<div style="flex:1;min-width:0;display:flex;flex-direction:column;gap:30px">
						{#each grupos as s (s.t)}
							<div>
								{#if s.t}<h2
										style="margin:0 0 12px;{B};font-weight:800;font-size:30px;letter-spacing:-.03em"
									>
										{s.t}
									</h2>{/if}
								<div
									style="display:grid;grid-template-columns:repeat(auto-fill,minmax(min(100%,320px),1fr));gap:10px"
								>
									{#each s.items as p (p.id)}{@const q = enCarrito(p.id)}
										<div
											style="display:flex;gap:12px;padding:10px;border-radius:22px;background:#fff;border:1.5px solid {q
												? '#C9821A'
												: 'transparent'};min-width:0"
										>
											<div
												style="flex:1;min-width:0;padding:4px 2px 4px 6px;display:flex;flex-direction:column"
											>
												<div style="display:flex;align-items:center;gap:6px;flex-wrap:wrap">
													<a
														href={p.href}
														style="font-size:16px;font-weight:700;color:#3B2416;text-decoration:none"
														>{p.n}</a
													>{#if p.tag}<span
															style="padding:2px 8px;border-radius:999px;background:{p.tagBg};color:{p.tagFg};font-size:11px;font-weight:700"
															>{p.tag}</span
														>{/if}
												</div>
												{#if p.d}<span
														class="dos"
														style="font-size:13px;color:#7A5A44;line-height:1.4;margin-top:3px"
														>{p.d}</span
													>{/if}
												<span style="flex:1"></span>
												<span style="margin-top:8px"
													><b style="font-size:15.5px">{p.price}</b>{#if p.old}
														<s style="font-size:13px;color:#9C8270">{p.old}</s>{/if}</span
												>
											</div>
											<div
												style="position:relative;width:110px;height:110px;flex:none;border-radius:16px;overflow:hidden;background:#E9D6B8"
											>
												<a
													href={p.href}
													aria-label={p.n}
													style="display:block;width:100%;height:100%"
													>{#if p.img}<img
															src={img(porId.get(p.id)?.foto ?? null, 240, 240)}
															alt=""
															style="width:100%;height:100%;object-fit:cover"
														/>{/if}</a
												>
												<span
													style="position:absolute;right:6px;bottom:6px;display:flex;align-items:center;gap:2px;padding:3px;border-radius:999px;background:#fff;box-shadow:0 4px 10px rgba(59,36,22,.2)"
													>{#if q}<button
															type="button"
															aria-label="Uno menos"
															onclick={() => carrito.fijar(p.id, q - 1)}
															style="width:26px;height:26px;display:grid;place-items:center;border:0;background:transparent;padding:0;cursor:pointer;color:#3B2416"
															><Icon icon={I.phMinusBold} size={12} /></button
														><b style="min-width:16px;text-align:center;font-size:13px">{q}</b
														>{/if}<button
														type="button"
														aria-label="Sumar {p.n}"
														onclick={() => sumar(p.id)}
														disabled={p.agotado || tienda.vitrina || (q > 0 && q >= p.stock)}
														style="width:26px;height:26px;border:0;padding:0;border-radius:50%;background:#3B2416;color:#FBF3E4;display:grid;place-items:center;cursor:pointer"
														><Icon icon={I.phPlusBold} size={12} /></button
													></span
												>
											</div>
										</div>{/each}
								</div>
							</div>
						{:else}
							<p style="margin:20px 0;font-size:15px;color:#7A5A44">
								Todavía no hay productos en esta&nbsp;categoría.
							</p>
						{/each}
						{#if listado.length > mostrar}
							<div style="display:flex;justify-content:center">
								<button
									type="button"
									onclick={() => (mostrar += 24)}
									style="border:0;cursor:pointer;padding:11px 22px;border-radius:999px;background:#F1E2C8;color:#3B2416;font:700 14px 'Instrument Sans'"
									>Ver más</button
								>
							</div>
						{/if}
					</div>
					{#if wide}
						<aside
							style="width:330px;flex:none;position:sticky;top:130px;padding:20px;border-radius:26px;background:#fff;box-shadow:0 24px 50px -30px rgba(59,36,22,.4);display:flex;flex-direction:column;gap:12px"
						>
							<div style="{B};font-weight:800;font-size:24px;letter-spacing:-.03em">Tu pedido</div>
							{#each lineas as o (o.clave)}<div
									style="display:flex;align-items:center;gap:10px;font-size:14px"
								>
									<b
										style="width:26px;height:26px;border-radius:8px;background:#F1E2C8;display:grid;place-items:center;font-size:12.5px;flex:none"
										>{o.cantidad}</b
									><span style="flex:1;min-width:0"
										>{o.producto.nombre}{#if o.variante}<span
												style="display:block;font-size:12.5px;color:#7A5A44">{o.variante}</span
											>{/if}</span
									><b>{fmt(o.subtotal)}</b>
								</div>{:else}<span style="font-size:13.5px;color:#7A5A44"
									>Sumá productos del menú con el&nbsp;+.</span
								>{/each}
							{#if lineas.length}
								<div style="height:1px;background:rgba(59,36,22,.1)"></div>
								<div style="display:flex;justify-content:space-between;align-items:baseline">
									<b>Total</b><b style="{B};font-size:28px;font-weight:800">{fmt(subtotal)}</b>
								</div>
								{#if tienda.cobros.transferencia > 0}<span
										style="font-size:13px;color:#2F6B3A;font-weight:600"
										>{fmt(subtotal * (1 - tienda.cobros.transferencia / 100))} pagando por transferencia</span
									>{/if}
								<a
									href="{base}/checkout"
									style="padding:16px;border-radius:999px;background:#3B2416;color:#FBF3E4;font:700 15px 'Instrument Sans';text-align:center;text-decoration:none"
									>Finalizar pedido</a
								>{#if whatsapp}<a
										href={wa(textoPedido)}
										target="_blank"
										rel="noopener"
										style="display:flex;justify-content:center;align-items:center;gap:8px;padding:14px;border-radius:999px;background:#E8F7EC;color:#1F6F3A;font-size:14.5px;font-weight:700;text-decoration:none"
										><Icon icon={I.phWhatsappLogoFill} size={18} />Mandar por WhatsApp</a
									>{/if}
							{/if}
						</aside>
					{/if}
				</div>
				{#if nar && carrito.unidades}<a
						href="{base}/carrito"
						style="position:sticky;bottom:14px;margin-top:20px;width:100%;box-sizing:border-box;display:flex;align-items:center;gap:10px;padding:14px 18px;border-radius:999px;background:#3B2416;color:#FBF3E4;font:700 15px 'Instrument Sans';text-decoration:none;box-shadow:0 16px 30px -12px rgba(59,36,22,.6)"
						><span
							style="width:26px;height:26px;border-radius:50%;background:#F5C26B;color:#2A170C;display:grid;place-items:center;font-size:13px"
							>{carrito.unidades}</span
						><span style="flex:1;text-align:left">Ver pedido</span>{fmt(subtotal)}</a
					>{/if}
			</section>
		{/if}

		<!-- ============ PRODUCTO ============ -->
		{#if pagina === 'producto' && producto && sel}
			<section
				style="padding:clamp(18px,3cqw,32px) clamp(14px,3.5cqw,40px) 0;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,380px),1fr));gap:clamp(20px,4cqw,56px);align-items:start"
			>
				<div style="display:flex;flex-direction:column;gap:10px;min-width:0">
					<div
						style="position:relative;aspect-ratio:1;border-radius:30px;overflow:hidden;background:#E9C98F"
					>
						{#if sel.foto}<img
								src={img(sel.foto, 1000, 1000)}
								alt={producto.nombre}
								style="width:100%;height:100%;object-fit:cover"
							/>{/if}{#if tachado || producto.etiqueta}<span
								style="position:absolute;top:14px;left:14px;padding:4px 10px;border-radius:999px;background:#FBF3E4;font-size:11.5px;font-weight:700"
								>{tachado
									? `−${Math.round((1 - sel.precio / tachado) * 100)}%`
									: producto.etiqueta}</span
							>{/if}
					</div>
					{#if sel.fotos.length > 1}<div
							style="display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:10px"
						>
							{#each sel.fotos as f, i (f)}<button
									type="button"
									aria-label="Ver la foto {i + 1}"
									aria-pressed={i === sel.fotoIdx}
									onclick={() => sel.verFoto(i)}
									style="aspect-ratio:1;padding:0;border:0;border-radius:18px;overflow:hidden;background:#E9D6B8;cursor:pointer;box-shadow:0 0 0 {i ===
									sel.fotoIdx
										? '2.5px'
										: '0px'} #3B2416"
									><img
										src={img(f, 200, 200)}
										alt=""
										style="display:block;width:100%;height:100%;object-fit:cover"
									/></button
								>{/each}
						</div>{/if}
				</div>
				<div style="position:sticky;top:80px;min-width:0">
					<div style="font-size:13px;color:#7A5A44">
						<a href={base || '/'} style="color:inherit;text-decoration:none">Inicio</a> /
						<a href="{base}/catalogo" style="color:inherit;text-decoration:none">Menú</a>
					</div>
					{#if producto.categoria}<a
							href={enlaceCat(producto.categoria)}
							style="display:inline-flex;margin-top:12px;padding:5px 11px;border-radius:999px;background:#F1E2C8;font-size:12.5px;font-weight:700;color:#3B2416;text-decoration:none"
							>{producto.categoria}</a
						>{/if}
					<h1
						style="margin:12px 0 0;{B};font-weight:800;font-size:clamp(42px,5cqw,68px);line-height:.92;letter-spacing:-.045em;overflow-wrap:break-word"
					>
						{producto.nombre}
					</h1>
					{#if resenasDe}<div
							style="margin-top:10px;display:flex;align-items:center;gap:8px;font-size:13px;font-weight:600"
						>
							<span style="display:flex;color:#C9821A"
								>{#each estrellas(5) as e, i (i)}<Icon icon={e} size={13} />{/each}</span
							><span>{resenasDe} {resenasDe === 1 ? 'reseña' : 'reseñas'}</span>
						</div>{/if}
					{#if producto.descripcion}<p
							style="margin:12px 0 0;font-size:16px;line-height:1.55;color:#5E4332"
						>
							{producto.descripcion}
						</p>{/if}
					<div style="margin-top:18px;display:flex;align-items:baseline;gap:10px">
						<b style="{B};font-size:36px;font-weight:800;letter-spacing:-.03em"
							>{sel.desde ? 'Desde ' : ''}{fmt(sel.precio)}</b
						>{#if tachado}<s style="color:#9C8270">{fmt(tachado)}</s>{/if}
					</div>
					{#if tienda.cobros.cuotas > 1 || tienda.cobros.transferencia > 0}<div
							style="font-size:13.5px;color:#5E4332"
						>
							{#if tienda.cobros.cuotas > 1}{tienda.cobros.cuotas} cuotas sin interés de {fmt(
									sel.precio / tienda.cobros.cuotas
								)}{/if}{#if tienda.cobros.transferencia > 0}
								{#if tienda.cobros.cuotas > 1}·{/if}
								<b style="color:#2F6B3A"
									>{fmt(sel.precio * (1 - tienda.cobros.transferencia / 100))} por transferencia</b
								>{/if}
						</div>{/if}
					<!-- Venta por mayor (§49): con el estilo del renglón de las cuotas. -->
					{#if porMayorFicha}<div style="font-size:13.5px;color:#5E4332">{porMayorFicha}</div>{/if}
					{#each sel.opciones as o (o.nombre)}{@const swatch = o.valores.every(
							(x) => x.color || x.foto
						)}
						<div style="margin-top:18px">
							<div
								style="font-size:13.5px;font-weight:700;margin-bottom:8px;color:{sel.aviso &&
								!o.elegido
									? '#C9821A'
									: '#3B2416'}"
							>
								{o.nombre}{#if o.elegido}<span style="font-weight:400;color:#7A5A44"
										>: {o.elegido}</span
									>{/if}
							</div>
							{#if swatch}<div style="display:flex;gap:10px;flex-wrap:wrap">
									{#each o.valores as x (x.valor)}<button
											type="button"
											aria-label="{o.nombre} {x.valor}"
											aria-pressed={x.elegido}
											disabled={!x.disponible && !x.elegido}
											onclick={() => sel.elegir(o.nombre, x.valor)}
											style="width:42px;height:42px;padding:0;border:0;border-radius:50%;overflow:hidden;background:#E9D6B8;cursor:pointer;opacity:{x.disponible
												? 1
												: 0.4};box-shadow:inset 0 0 0 1px rgba(59,36,22,.14),0 0 0 {x.elegido
												? '2.5px'
												: '0px'} #FBF3E4,0 0 0 {x.elegido ? '5px' : '0px'} #3B2416"
											>{#if x.color}<span
													style="display:block;width:100%;height:100%;background:{x.color}"
												></span>{:else}<img
													src={img(x.foto, 80, 80)}
													alt=""
													style="display:block;width:100%;height:100%;object-fit:cover"
												/>{/if}</button
										>{/each}
								</div>{:else}<div style="display:flex;gap:8px;flex-wrap:wrap">
									{#each o.valores as x (x.valor)}<button
											type="button"
											aria-pressed={x.elegido}
											disabled={!x.disponible && !x.elegido}
											onclick={() => sel.elegir(o.nombre, x.valor)}
											style="min-height:44px;padding:10px 16px;border-radius:999px;border:0;background:{x.elegido
												? '#3B2416'
												: '#fff'};color:{x.elegido
												? '#FBF3E4'
												: x.disponible
													? '#3B2416'
													: '#9C8270'};font:700 14px 'Instrument Sans';text-decoration:{x.disponible
												? 'none'
												: 'line-through'};opacity:{x.disponible ? 1 : 0.6};cursor:{x.disponible
												? 'pointer'
												: 'default'}">{x.valor}</button
										>{/each}
								</div>{/if}
						</div>{/each}
					{#if sel.stock > 0 && sel.stock <= 3 && (!sel.conOpciones || sel.variante)}<div
							style="margin-top:12px;font-size:13px;font-weight:700;color:#C9821A"
						>
							{sel.stock === 1 ? 'Queda 1' : `Quedan ${sel.stock}`}{#if sel.variante}
								en {sel.texto.toLowerCase()}{/if}
						</div>{/if}
					<div style="margin-top:14px;display:flex;gap:8px">
						{#if sel.stock > 0}<span
								style="display:flex;align-items:center;gap:4px;padding:4px;border-radius:999px;background:#fff;flex:none"
								><button
									type="button"
									aria-label="Uno menos"
									disabled={cantidad <= 1}
									onclick={() => (cantidad = Math.max(1, cantidad - 1))}
									style="width:40px;height:40px;display:grid;place-items:center;border:0;background:transparent;padding:0;cursor:pointer;color:#3B2416"
									><Icon icon={I.phMinusBold} size={14} /></button
								><b style="width:24px;text-align:center">{cantidad}</b><button
									type="button"
									aria-label="Uno más"
									disabled={cantidad >= sel.stock}
									onclick={() => (cantidad = Math.min(sel.stock, cantidad + 1))}
									style="width:40px;height:40px;display:grid;place-items:center;border:0;background:transparent;padding:0;cursor:pointer;color:#3B2416"
									><Icon icon={I.phPlusBold} size={14} /></button
								></span
							>{/if}<button
							type="button"
							onclick={agregar}
							disabled={sel.stock === 0 || tienda.vitrina}
							style="flex:1;padding:15px;border:0;border-radius:999px;background:#3B2416;color:#FBF3E4;font:700 15px 'Instrument Sans';cursor:pointer"
							>{sel.boton(`Agregar · ${fmt(sel.precio * cantidad)}`)}</button
						>
					</div>
					{#if whatsapp}<a
							href={wa(
								`¡Hola! Quiero consultar por ${producto.nombre}${sel.variante ? ` (${sel.texto})` : ''} (${fmt(sel.precio)}).`
							)}
							target="_blank"
							rel="noopener"
							style="margin-top:8px;display:flex;justify-content:center;align-items:center;gap:8px;padding:14px;border-radius:999px;background:#E8F7EC;color:#1F6F3A;font-size:14.5px;font-weight:700;text-decoration:none"
							><Icon icon={I.phWhatsappLogoFill} size={18} />Consultar por WhatsApp</a
						>{/if}
					<LanzamientoFicha {tienda} {producto} {sel} {tema} />
					<CompartirProducto {tienda} {producto} {tema} />
					{#if envia(tienda.envios)}<div
							style="margin-top:14px;padding:14px;border-radius:18px;background:#fff;display:flex;flex-direction:column;gap:8px;font-size:13.5px"
						>
							<label for="cp-ficha" style="font-weight:700">Calculá el envío</label>
							<div style="display:flex;gap:6px">
								<input
									id="cp-ficha"
									bind:value={cp}
									inputmode="numeric"
									maxlength="4"
									placeholder="Código postal"
									onkeydown={(e) => e.key === 'Enter' && calcular()}
									style="flex:1;min-width:0;padding:10px 12px;border-radius:12px;border:0;background:#FBF3E4;font:400 13.5px 'Instrument Sans';color:#3B2416"
								/><button
									type="button"
									onclick={calcular}
									style="padding:10px 14px;border:0;border-radius:12px;background:#3B2416;color:#FBF3E4;font:700 13.5px 'Instrument Sans';cursor:pointer"
									>Calcular</button
								>
							</div>
							{#if cotizado}{#each cotizado as o (o.id)}<span
										style="display:flex;justify-content:space-between;gap:10px"
										><span>{o.nombre} · {o.detalle}</span><b
											style="white-space:nowrap;color:{o.precio ? '#3B2416' : '#2F6B3A'}"
											>{o.precio ? fmt(o.precio) : 'Gratis'}</b
										></span
									>{/each}{/if}
						</div>{/if}
					<div style="margin-top:14px;display:flex;flex-direction:column;gap:6px">
						{#each acc as a (a.t)}<div
								style="padding:12px 16px;border-radius:14px;background:{a.open
									? '#F1E2C8'
									: '#fff'};font-size:14.5px"
							>
								<button
									type="button"
									onclick={() => (abierto = a.open ? '' : a.t)}
									aria-expanded={a.open}
									style="width:100%;display:flex;justify-content:space-between;align-items:center;padding:0;border:0;background:transparent;font:600 14.5px 'Instrument Sans';color:#3B2416;cursor:pointer;text-align:left"
									>{a.t}<Icon icon={a.i} size={14} /></button
								>{#if a.open}<p
										style="margin:8px 0 0;font-size:14px;line-height:1.55;color:#5E4332;text-wrap:pretty"
									>
										{a.d}
									</p>{/if}
							</div>{/each}
					</div>
				</div>
			</section>
			<section style="padding:clamp(40px,6cqw,72px) clamp(14px,3.5cqw,40px) 0">
				<div
					style="max-width:820px;padding:clamp(18px,3cqw,28px);border-radius:24px;background:#fff"
				>
					<PreguntasProducto
						slug={tienda.slug}
						productoId={producto.id}
						preguntas={tienda.preguntas}
						{tema}
						titulo="Preguntas sobre este producto"
					/>
				</div>
			</section>
			{#if related.length}
				<section style="padding:clamp(48px,7cqw,90px) clamp(14px,3.5cqw,40px)">
					<h2
						style="margin:0;{B};font-weight:800;font-size:clamp(32px,4.4cqw,52px);letter-spacing:-.04em"
					>
						Va bien con
					</h2>
					<div
						style="margin-top:20px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(46%,220px),1fr));gap:12px"
					>
						{#each related as p (p.id)}<a
								href={p.href}
								class="zoom"
								style="display:block;padding:10px;border-radius:22px;background:#fff;color:#3B2416;text-decoration:none;min-width:0"
								><span
									style="display:block;aspect-ratio:1;border-radius:16px;overflow:hidden;background:#E9D6B8"
									>{#if p.img}<img
											src={p.img}
											alt=""
											style="width:100%;height:100%;object-fit:cover"
										/>{/if}</span
								><span
									style="padding:10px 4px 2px;display:flex;justify-content:space-between;gap:8px;flex-wrap:wrap"
									><b style="font-size:14.5px">{p.n}</b><span style="font-size:14px;font-weight:700"
										>{p.price}</span
									></span
								></a
							>{/each}
					</div>
				</section>
			{:else}
				<div style="height:clamp(48px,7cqw,90px)"></div>
			{/if}
		{/if}

		<!-- ============ CARRITO ============ -->
		{#if pagina === 'carrito'}
			<section
				style="padding:clamp(18px,3cqw,32px) clamp(14px,3.5cqw,40px) clamp(48px,7cqw,90px);display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,360px),1fr));gap:clamp(20px,4cqw,48px);align-items:start"
			>
				<div style="min-width:0">
					<h1
						style="margin:0;{B};font-weight:800;font-size:clamp(44px,6cqw,80px);line-height:.9;letter-spacing:-.05em"
					>
						Tu pedido
					</h1>
					{#if !lineas.length}
						<p style="margin:18px 0 0;font-size:15px;color:#7A5A44">Todavía no sumaste nada.</p>
						<a
							href="{base}/catalogo"
							style="display:inline-block;margin-top:16px;padding:15px 24px;border-radius:999px;background:#3B2416;color:#FBF3E4;font:700 15px 'Instrument Sans';text-decoration:none"
							>Ver el menú</a
						>
					{:else}
						{#if tienda.envios.gratisDesde}
							<div
								style="margin-top:14px;padding:12px 14px;border-radius:16px;background:#F1E2C8;font-size:13.5px;display:flex;flex-direction:column;gap:10px"
							>
								<span style="display:flex;align-items:center;gap:10px"
									><Icon icon={I.phPackageFill} size={20} style="color:#C9821A" /><span
										style="flex:1"
										>{#if falta}Sumá <b>{fmt(falta)}</b> y el envío es gratis{:else}<b
												>Tenés envío gratis</b
											>{/if}</span
									></span
								><span
									style="height:5px;border-radius:3px;background:#EADBC2;overflow:hidden;display:block"
									><span
										style="display:block;height:100%;width:{Math.min(
											100,
											(subtotal / tienda.envios.gratisDesde) * 100
										)}%;background:var(--acento, #C9821A)"
									></span></span
								>
							</div>
						{/if}
						{#if avisoMayor && porMayor}
							<div
								style="margin-top:14px;padding:12px 14px;border-radius:16px;background:#F1E2C8;font-size:13.5px;display:flex;flex-direction:column;gap:10px"
							>
								<span style="display:flex;align-items:center;gap:10px"
									><Icon icon={I.phPackageFill} size={20} style="color:#C9821A" /><span
										style="flex:1">{avisoMayor.pre}<b>{avisoMayor.b}</b>{avisoMayor.post}</span
									></span
								><span
									style="height:5px;border-radius:3px;background:#EADBC2;overflow:hidden;display:block"
									><span
										style="display:block;height:100%;width:{porMayor.progreso}%;background:var(--acento, #C9821A)"
									></span></span
								>
							</div>
						{/if}
						<div style="margin-top:16px;display:flex;flex-direction:column;gap:10px">
							{#each lineas as o (o.clave)}<div
									style="display:flex;gap:12px;align-items:center;padding:10px;border-radius:20px;background:#fff"
								>
									{#if o.producto.foto}<img
											src={img(o.producto.foto, 160, 160)}
											alt=""
											style="width:72px;height:72px;border-radius:14px;object-fit:cover;flex:none;background:#E9D6B8"
										/>{/if}
									<div style="flex:1;min-width:0">
										<a
											href="{base}/producto/{o.producto.slug ?? o.id}"
											style="font-size:15px;font-weight:700;color:#3B2416;text-decoration:none"
											>{o.producto.nombre}</a
										>
										<div style="font-size:12.5px;color:#7A5A44">
											{o.variante ? `${o.variante} · ` : ''}{fmt(o.producto.precio)} c/u
										</div>
										<div style="margin-top:4px;display:flex;align-items:center;gap:10px">
											<b style="font-size:14.5px">{fmt(o.subtotal)}</b><button
												type="button"
												onclick={() => carrito.fijar(o.clave, 0)}
												style="border:0;background:transparent;padding:0;font:600 12.5px 'Instrument Sans';text-decoration:underline;color:#7A5A44;cursor:pointer"
												>Quitar</button
											>
										</div>
									</div>
									<span
										style="display:flex;align-items:center;gap:2px;padding:3px;border-radius:999px;background:#FBF3E4;flex:none"
										><button
											type="button"
											aria-label="Uno menos"
											onclick={() => carrito.fijar(o.clave, o.cantidad - 1)}
											style="width:30px;height:30px;display:grid;place-items:center;border:0;background:transparent;padding:0;cursor:pointer;color:#3B2416"
											><Icon icon={I.phMinusBold} size={12} /></button
										><b style="min-width:18px;text-align:center">{o.cantidad}</b><button
											type="button"
											aria-label="Uno más"
											disabled={o.cantidad >= o.producto.stock}
											onclick={() => carrito.fijar(o.clave, o.cantidad + 1)}
											style="width:30px;height:30px;border:0;padding:0;border-radius:50%;background:#3B2416;color:#FBF3E4;display:grid;place-items:center;cursor:pointer"
											><Icon icon={I.phPlusBold} size={12} /></button
										></span
									>
								</div>{/each}
						</div>
						{#if upsell.length}
							<div style="margin-top:20px">
								<b style="font-size:15px">¿Le sumás algo?</b>
								<div
									style="margin-top:10px;display:flex;gap:8px;overflow-x:auto;scrollbar-width:none"
								>
									{#each upsell as u (u.id)}<div
											style="flex:none;display:flex;align-items:center;gap:10px;padding:8px 12px 8px 8px;border-radius:18px;background:#fff"
										>
											{#if u.foto}<img
													src={img(u.foto, 120, 120)}
													alt=""
													style="width:48px;height:48px;border-radius:12px;object-fit:cover"
												/>{/if}
											<div>
												<a
													href="{base}/producto/{u.slug ?? u.id}"
													style="display:block;font-size:13.5px;font-weight:700;white-space:nowrap;color:#3B2416;text-decoration:none"
													>{u.nombre}</a
												>
												<div style="font-size:13px">
													{tarjeta(u).price}
												</div>
											</div>
											<button
												type="button"
												aria-label="Sumar {u.nombre}"
												disabled={tienda.vitrina}
												onclick={() => sumar(u.id)}
												style="width:30px;height:30px;border:0;padding:0;border-radius:50%;background:#F5C26B;color:#2A170C;display:grid;place-items:center;cursor:pointer"
												><Icon icon={I.phPlusBold} size={13} /></button
											>
										</div>{/each}
								</div>
							</div>
						{/if}
					{/if}
				</div>
				{#if lineas.length}
					<div
						style="position:sticky;top:80px;padding:22px;border-radius:26px;background:#3B2416;color:#FBF3E4;display:flex;flex-direction:column;gap:12px;font-size:14.5px"
					>
						<span style="display:flex;justify-content:space-between;gap:10px"
							><span style="color:#E6D3BC">Subtotal</span><span>{fmt(subtotal)}</span></span
						>
						<span style="display:flex;justify-content:space-between;gap:10px"
							><span style="color:#E6D3BC">Envío</span><span style="text-align:right"
								>Se calcula en el próximo paso</span
							></span
						>
						<div style="height:1px;background:rgba(251,243,228,.14)"></div>
						<span style="display:flex;justify-content:space-between;align-items:baseline"
							><b>Total</b><b style="{B};font-size:32px;font-weight:800">{fmt(subtotal)}</b></span
						>
						{#if tienda.cobros.cuotas > 1}<span style="font-size:13px;color:#E6D3BC"
								>{tienda.cobros.cuotas} cuotas sin interés de {fmt(
									subtotal / tienda.cobros.cuotas
								)}</span
							>{/if}
						{#if tienda.cobros.transferencia > 0}<span
								style="font-size:13px;color:#F5C26B;font-weight:600"
								>{fmt(subtotal * (1 - tienda.cobros.transferencia / 100))} pagando por transferencia</span
							>{/if}
						<a
							href="{base}/checkout"
							style="padding:16px;border-radius:999px;background:#F5C26B;color:#2A170C;font:700 15px 'Instrument Sans';text-align:center;text-decoration:none"
							>Continuar</a
						>
					</div>
				{/if}
			</section>
		{/if}

		<!-- ============ CHECKOUT ============ -->
		{#if pagina === 'checkout'}
			{#if !lineas.length}
				<section
					style="padding:clamp(18px,3cqw,32px) clamp(14px,3.5cqw,40px) clamp(48px,7cqw,90px);max-width:760px;margin:0 auto"
				>
					<p style="font-size:15px;color:#7A5A44">Tu pedido está vacío.</p>
					<a href="{base}/catalogo" style="color:#3B2416;font-weight:700">Ver el menú</a>
				</section>
			{:else}
				<form
					method="POST"
					use:enhance={() => {
						enviando = true;
						return async ({ result, update }) => {
							enviando = false;
							// Mercado Pago es otro sitio: se navega de verdad.
							if (result.type === 'redirect' && result.location.startsWith('https://'))
								return void (location.href = result.location);
							if (result.type === 'redirect') carrito.vaciar();
							await update();
						};
					}}
					style="padding:clamp(18px,3cqw,32px) clamp(14px,3.5cqw,40px) clamp(48px,7cqw,90px);max-width:760px;margin:0 auto;display:flex;flex-direction:column;gap:16px"
				>
					<input type="hidden" name="carrito" value={JSON.stringify(carrito.lineas)} />
					<input type="hidden" name="envio" value={envio ?? ''} />
					<input type="hidden" name="pago" value={pago} />
					<h1
						style="margin:0;{B};font-weight:800;font-size:clamp(40px,5.4cqw,68px);line-height:.92;letter-spacing:-.05em"
					>
						Casi listo
					</h1>
					<div
						style="padding:18px;border-radius:24px;background:#fff;display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));gap:10px"
					>
						<b style="grid-column:1/-1;font-size:16px">Tus datos</b>
						{#each [['email', 'Email', 'email', 'email'], ['nombre', 'Nombre y apellido', 'text', 'name'], ['telefono', 'WhatsApp', 'tel', 'tel'], ['dni', 'DNI (opcional)', 'text', 'off'], ...camposFactura(tienda)] as [n, l, tipo, ac] (n)}
							<label style="display:block;padding:10px 12px;border-radius:14px;background:#FBF3E4"
								><span style="display:block;font-size:11.5px;color:#7A5A44">{l}</span><input
									name={n}
									type={tipo}
									autocomplete={ac as 'email'}
									required={!OPCIONALES.has(n)}
									maxlength="120"
									style="width:100%;margin-top:2px;padding:0;border:0;background:transparent;font:400 15px 'Instrument Sans';color:#3B2416;outline:none"
								/></label
							>
						{/each}
					</div>
					<div
						style="padding:18px;border-radius:24px;background:#fff;display:flex;flex-direction:column;gap:8px"
					>
						<b style="font-size:16px">¿Cómo te llega?</b>
						<div style="display:grid;grid-template-columns:minmax(0,130px) minmax(0,1fr);gap:10px">
							<label style="display:block;padding:10px 12px;border-radius:14px;background:#FBF3E4"
								><span style="display:block;font-size:11.5px;color:#7A5A44">Código postal</span
								><input
									name="cp"
									bind:value={cp}
									inputmode="numeric"
									maxlength="4"
									style="width:100%;margin-top:2px;padding:0;border:0;background:transparent;font:400 15px 'Instrument Sans';color:#3B2416;outline:none"
								/></label
							>
							<label style="display:block;padding:10px 12px;border-radius:14px;background:#FBF3E4"
								><span style="display:block;font-size:11.5px;color:#7A5A44"
									>Dirección (calle, número, piso)</span
								><input
									name="direccion"
									maxlength="160"
									autocomplete="street-address"
									required={envio !== 'retiro'}
									style="width:100%;margin-top:2px;padding:0;border:0;background:transparent;font:400 15px 'Instrument Sans';color:#3B2416;outline:none"
								/></label
							>
						</div>
						<div
							style="display:flex;flex-direction:column;gap:8px"
							role="radiogroup"
							aria-label="Forma de entrega"
						>
							{#each opciones as s (s.id)}{@const on = envio === s.id}<button
									type="button"
									role="radio"
									aria-checked={on}
									onclick={() => (envio = s.id)}
									style="display:flex;align-items:center;gap:12px;padding:14px;border-radius:16px;border:1.5px solid {on
										? '#3B2416'
										: 'rgba(59,36,22,.12)'};background:{on
										? '#FBF3E4'
										: 'transparent'};text-align:left;cursor:pointer;font-family:inherit;color:#3B2416"
									><Icon
										icon={s.id === 'retiro'
											? I.phStorefrontFill
											: s.id === 'moto'
												? I.phMopedFill
												: I.phPackageFill}
										size={22}
										style="color:#C9821A"
									/><span style="flex:1;min-width:0"
										><b style="display:block;font-size:14.5px">{s.nombre}</b><span
											style="display:block;font-size:12.5px;color:#7A5A44">{s.detalle}</span
										></span
									><b
										style="font-size:14px;white-space:nowrap;color:{s.precio
											? '#3B2416'
											: '#2F6B3A'}">{s.precio ? fmt(s.precio) : 'Gratis'}</b
									><span
										style="width:20px;height:20px;border-radius:50%;border:2px solid #3B2416;display:grid;place-items:center;flex:none"
										>{#if on}<span
												style="width:10px;height:10px;border-radius:50%;background:#3B2416"
											></span>{/if}</span
									></button
								>{:else}<span style="font-size:13.5px;color:#7A5A44"
									>Poné tu código postal para ver cómo te&nbsp;llega.</span
								>{/each}
						</div>
						<EnvioCorreo {tienda} {cp} {envio} {opciones} />
					</div>
					<div
						style="padding:18px;border-radius:24px;background:#fff;display:flex;flex-direction:column;gap:8px"
						role="radiogroup"
						aria-label="Forma de pago"
					>
						<b style="font-size:16px">Pago</b>
						{#each pagos as p (p.id)}{@const on = pago === p.id}<button
								type="button"
								role="radio"
								aria-checked={on}
								onclick={() => (pago = p.id)}
								style="display:flex;align-items:center;gap:12px;padding:14px;border-radius:16px;border:1.5px solid {on
									? '#3B2416'
									: 'rgba(59,36,22,.12)'};background:{on
									? '#FBF3E4'
									: 'transparent'};text-align:left;cursor:pointer;font-family:inherit;color:#3B2416"
								><Icon icon={p.i} size={22} style="color:{p.c}" /><span style="flex:1;min-width:0"
									><b style="display:block;font-size:14.5px">{p.t}</b><span
										style="display:block;font-size:12.5px;color:#7A5A44">{p.d}</span
									></span
								><span
									style="width:20px;height:20px;border-radius:50%;border:2px solid #3B2416;display:grid;place-items:center;flex:none"
									>{#if on}<span style="width:10px;height:10px;border-radius:50%;background:#3B2416"
										></span>{/if}</span
								></button
							>{/each}
					</div>
					<div
						style="padding:18px;border-radius:24px;background:#fff;display:flex;flex-direction:column;gap:12px"
					>
						<b style="font-size:16px">Tu pedido</b>
						{#each lineas as o (o.clave)}<div
								style="display:flex;align-items:center;gap:10px;font-size:14px"
							>
								<b
									style="width:26px;height:26px;border-radius:8px;background:#F1E2C8;display:grid;place-items:center;font-size:12.5px;flex:none"
									>{o.cantidad}</b
								><span style="flex:1;min-width:0"
									>{o.producto.nombre}{#if o.variante}<span
											style="display:block;font-size:12.5px;color:#7A5A44">{o.variante}</span
										>{/if}</span
								><b>{fmt(o.subtotal)}</b>
							</div>{/each}
						<div style="height:1px;background:rgba(59,36,22,.1)"></div>
						<span style="display:flex;justify-content:space-between;font-size:14px"
							><span>Subtotal</span><span>{fmt(subtotal)}</span></span
						><span style="display:flex;justify-content:space-between;font-size:14px"
							><span>Envío</span><span
								>{elegida ? (cuenta.envio ? fmt(cuenta.envio) : 'Gratis') : '—'}</span
							></span
						>{#if cuenta.descCupon}<span
								style="display:flex;justify-content:space-between;font-size:14px;color:#2F6B3A;font-weight:600"
								><span>Cupón {cupon.aplicado}</span><span>− {fmt(cuenta.descCupon)}</span></span
							>{/if}{#if descuento}<span
								style="display:flex;justify-content:space-between;font-size:14px;color:#2F6B3A;font-weight:600"
								><span>Transferencia −{tienda.cobros.transferencia}%</span><span
									>− {fmt(descuento)}</span
								></span
							>{/if}
						<div style="margin:6px 0 2px">
							<CampoCupon
								{cupon}
								lineas={carrito.lineas}
								tinta="#3B2416"
								suave="#7A5A44"
								borde="rgba(59,36,22,.14)"
								fondo="#FBF3E4"
								acento="#3B2416"
								sobreAcento="#FBF3E4"
								radio="14px"
								fuente="'Instrument Sans',sans-serif"
							/>
						</div>
						<div style="display:flex;justify-content:space-between;align-items:baseline">
							<b>Total</b><b style="{B};font-size:28px;font-weight:800">{fmt(total)}</b>
						</div>
					</div>
					{#if error}<div
							role="alert"
							style="padding:12px 14px;border-radius:16px;background:#F6E2D2;color:#9C3B1F;font-size:14px;font-weight:600"
						>
							{error}
						</div>{/if}
					<button
						disabled={!envio || !pagos.length || enviando || tienda.vitrina}
						style="padding:18px;border:0;border-radius:999px;background:#3B2416;color:#FBF3E4;font:700 16px 'Instrument Sans';cursor:pointer;display:flex;justify-content:center;align-items:center;gap:10px"
						>{tienda.vitrina
							? 'Es una tienda de ejemplo'
							: !pagos.length
								? 'Esta tienda todavía no tiene medios de pago'
								: envio
									? `Confirmar pedido · ${fmt(total)}`
									: 'Elegí cómo te llega'}<Icon icon={I.phArrowRightBold} size={16} /></button
					>
					<span
						style="display:flex;justify-content:center;align-items:center;gap:8px;text-align:center;font-size:13px;color:#7A5A44"
						><Icon icon={I.phShieldCheckFill} size={16} />Compra protegida · 10 días
						para&nbsp;arrepentirte</span
					>
				</form>
			{/if}
		{/if}

		<!-- ============ LISTO (no está en el diseño: con las piezas del checkout) ============ -->
		{#if pagina === 'listo' && pedido}
			<section
				style="padding:clamp(32px,5cqw,64px) clamp(14px,3.5cqw,40px) clamp(48px,7cqw,90px);max-width:760px;margin:0 auto;display:flex;flex-direction:column;gap:16px;text-align:center;align-items:center"
			>
				<span
					style="width:64px;height:64px;border-radius:50%;background:#3B2416;color:#F5C26B;display:grid;place-items:center"
					><Icon icon={I.phCheckBold} size={28} /></span
				>
				<h1
					style="margin:0;{B};font-weight:800;font-size:clamp(40px,5.4cqw,68px);line-height:.92;letter-spacing:-.05em;overflow-wrap:break-word;max-width:100%"
				>
					¡Gracias, {pedido.nombre}!
				</h1>
				<p style="margin:0;font-size:16px;line-height:1.55;color:#5E4332">
					Recibimos tu pedido <b>#{pedido.numero}</b> por <b>{fmt(pedido.total)}</b>. Te escribimos
					por mail con cada&nbsp;novedad.
				</p>
				<SenaPedido {tema} />
				{#if pedido.pago === 'transferencia'}
					<div
						style="padding:18px;border-radius:24px;background:#fff;text-align:left;font-size:14.5px;line-height:1.6;align-self:stretch"
					>
						<b>Transferí {fmt(pedido.total)}</b>
						{#if pedido.alias}al {pedido.esCbu ? 'CBU/CVU' : 'alias'}
							<b>{pedido.alias}</b>{:else}con los datos que te mandamos por mail{/if} y mandanos el comprobante{#if whatsapp}&nbsp;<a
								href={wa(`¡Hola! Te mando el comprobante del pedido #${pedido.numero}.`)}
								target="_blank"
								rel="noopener"
								style="color:#3B2416;font-weight:700">por WhatsApp</a
							>{:else}&nbsp;respondiendo el mail del pedido{/if}. Lo preparamos apenas
						se&nbsp;acredite.
					</div>
				{/if}
				<a
					href={base || '/'}
					style="margin-top:6px;padding:15px 24px;border-radius:999px;background:#3B2416;color:#FBF3E4;font:700 15px 'Instrument Sans';text-decoration:none"
					>Volver a la tienda</a
				>
				{#if pedido.estado === 'Entregado'}
					<div style="margin-top:clamp(24px,5cqw,48px);align-self:stretch">
						<FormResena
							slug={tienda.slug}
							tema={{ ...tema, borde: 'transparent' }}
							codigo={pedido.codigo}
							nombre={pedido.nombre}
						/>
					</div>
				{/if}
			</section>
		{/if}

		<!-- ============ TURNOS (no está en el diseño de esta plantilla: con las piezas de su checkout) ============ -->
		{#if pagina === 'turnos' && tienda.turnos && reserva}
			<form
				method="POST"
				use:enhance={() => {
					reservando = true;
					return async ({ result, update }) => {
						reservando = false;
						// Mercado Pago es otro sitio: se navega de verdad.
						if (result.type === 'redirect' && result.location.startsWith('https://'))
							return void (location.href = result.location);
						await update();
						// Si otro tomó el horario mientras tanto, se ve al toque.
						if (result.type === 'failure') void reserva.cargar();
					};
				}}
				style="display:flex;flex-direction:column;gap:16px;padding:clamp(18px,3cqw,32px) clamp(14px,3.5cqw,40px) clamp(48px,7cqw,90px);max-width:760px;margin:0 auto"
			>
				<input type="hidden" name="servicioId" value={reserva.servicioId ?? ''} />
				<input type="hidden" name="profesionalId" value={reserva.profesionalId ?? ''} />
				<input type="hidden" name="inicio" value={reserva.inicio ?? ''} />
				<h1
					style="margin:0;{B};font-weight:800;font-size:clamp(40px,5.4cqw,68px);line-height:.92;letter-spacing:-.05em"
				>
					Reservá tu turno
				</h1>
				<div
					style="padding:18px;border-radius:24px;background:#fff;display:flex;flex-direction:column;gap:8px"
				>
					<b style="font-size:16px">1 · Servicio</b>
					<div
						style="display:flex;flex-direction:column;gap:8px"
						role="radiogroup"
						aria-label="Servicio"
					>
						{#each tienda.turnos.servicios as s (s.id)}{@const on =
								reserva.servicioId === s.id}<button
								type="button"
								role="radio"
								aria-checked={on}
								onclick={() => reserva.elegirServicio(s.id)}
								style="display:flex;align-items:center;gap:12px;padding:15px;text-align:left;border-radius:16px;border:1.5px solid {on
									? '#3B2416'
									: 'rgba(59,36,22,.12)'};background:{on
									? '#FBF3E4'
									: 'transparent'};cursor:pointer;font-family:inherit;color:#3B2416"
								><span
									style="width:20px;height:20px;border-radius:50%;border:2px solid #3B2416;display:grid;place-items:center;flex:none"
									>{#if on}<span style="width:10px;height:10px;border-radius:50%;background:#3B2416"
										></span>{/if}</span
								><span style="flex:1;min-width:0"
									><span style="display:block;font-size:14.5px;font-weight:600">{s.nombre}</span
									><span style="display:block;font-size:12.5px;color:#7A5A44"
										>{duracion(s.duracion)}</span
									></span
								><b style="font-size:14.5px;white-space:nowrap">{fmt(s.precio)}</b></button
							>{/each}
					</div>
				</div>
				<div
					style="padding:18px;border-radius:24px;background:#fff;display:flex;flex-direction:column;gap:8px"
				>
					<b style="font-size:16px">2 · Profesional</b>
					<div
						style="display:flex;gap:8px;overflow-x:auto;scrollbar-width:none;padding-bottom:2px"
						role="radiogroup"
						aria-label="Profesional"
					>
						{#each [null, ...reserva.profesionales] as p (p?.id ?? 'cualquiera')}{@const on =
								reserva.profesionalId === (p?.id ?? null)}<button
								type="button"
								role="radio"
								aria-checked={on}
								onclick={() => reserva.elegirProfesional(p?.id ?? null)}
								style="flex:none;width:112px;padding:10px;text-align:center;border-radius:16px;border:1.5px solid {on
									? '#3B2416'
									: 'rgba(59,36,22,.12)'};background:{on
									? '#FBF3E4'
									: 'transparent'};cursor:pointer;font-family:inherit;color:#3B2416"
								>{#if p?.foto}<img
										src={img(p.foto, 112, 112)}
										alt=""
										style="width:52px;height:52px;border-radius:50%;object-fit:cover;display:block;margin:0 auto"
									/>{:else}<span
										style="width:52px;height:52px;border-radius:50%;background:#F1E2C8;color:#3B2416;display:grid;place-items:center;margin:0 auto"
										><Icon icon={p ? I.phUserFill : I.phShuffleBold} size={22} /></span
									>{/if}<span
									style="display:block;margin-top:8px;font-size:13.5px;font-weight:700;overflow:hidden;text-overflow:ellipsis;white-space:nowrap"
									>{p?.nombre ?? 'Cualquiera'}</span
								><span
									style="display:block;font-size:11.5px;color:#7A5A44;overflow:hidden;text-overflow:ellipsis;white-space:nowrap"
									>{p ? p.rol.split(' · ')[0] : 'El primero libre'}</span
								></button
							>{/each}
					</div>
				</div>
				<div
					style="padding:18px;border-radius:24px;background:#fff;display:flex;flex-direction:column;gap:8px"
				>
					<b style="font-size:16px">3 · Día y horario</b>
					<div
						style="display:flex;justify-content:flex-end;align-items:center;gap:8px;margin-bottom:8px;font-size:14px;font-weight:600"
					>
						{#if reserva.hayAntes || reserva.hayDespues}<button
								type="button"
								aria-label="Días anteriores"
								disabled={!reserva.hayAntes}
								onclick={() => reserva.mover(-1)}
								style="display:grid;place-items:center;min-width:44px;min-height:44px;padding:4px;border:0;background:transparent;color:inherit;cursor:pointer"
								><Icon icon={I.phCaretLeftBold} size={14} /></button
							>{/if}{reserva.mes}{#if reserva.hayAntes || reserva.hayDespues}<button
								type="button"
								aria-label="Días siguientes"
								disabled={!reserva.hayDespues}
								onclick={() => reserva.mover(1)}
								style="display:grid;place-items:center;min-width:44px;min-height:44px;padding:4px;border:0;background:transparent;color:inherit;cursor:pointer"
								><Icon icon={I.phCaretRightBold} size={14} /></button
							>{/if}
					</div>
					<div
						style="display:grid;grid-template-columns:repeat(7,minmax(0,1fr));gap:6px"
						role="radiogroup"
						aria-label="Día"
					>
						{#each reserva.dias as d (d.dia)}{@const sel = reserva.dia === d.dia}<button
								type="button"
								role="radio"
								aria-checked={sel}
								aria-label="{d.w} {d.n}{d.abierto ? '' : ', cerrado'}"
								disabled={!d.abierto}
								onclick={() => reserva.elegirDia(d.dia)}
								style="padding:9px 0;text-align:center;min-width:0;opacity:{d.abierto
									? 1
									: 0.35};border-radius:16px;border:1.5px solid {sel
									? '#3B2416'
									: 'rgba(59,36,22,.12)'};background:{sel
									? '#FBF3E4'
									: 'transparent'};cursor:pointer;font-family:inherit;color:#3B2416"
								><span style="display:block;font-size:11px;font-weight:600;text-transform:uppercase"
									>{d.w}</span
								><span style="display:block;font-size:17px;font-weight:700;margin-top:1px"
									>{d.n}</span
								></button
							>{/each}
					</div>
					<div style="margin-top:16px;display:flex;flex-direction:column;gap:12px">
						{#if !reserva.servicioId}<p style="margin:0;font-size:14px;color:#7A5A44">
								Elegí un servicio para ver los horarios.
							</p>{:else if reserva.horarios && !reserva.grupos.length}<p
								style="margin:0;font-size:14px;color:#7A5A44"
							>
								No quedan horarios este día. Probá con otro día u otro&nbsp;profesional.
							</p>{/if}
						{#each reserva.grupos as g (g.k)}<div>
								<div
									style="font-size:12.5px;font-weight:700;color:#7A5A44;margin-bottom:7px;display:flex;align-items:center;gap:6px"
								>
									<Icon icon={FRANJA[g.k]} size={14} />{g.t}
								</div>
								<div
									style="display:grid;grid-template-columns:repeat(auto-fill,minmax(76px,1fr));gap:6px"
									role="radiogroup"
									aria-label="Horarios a la {g.t.toLowerCase()}"
								>
									{#each g.s as h (h.inicio)}{@const on = reserva.inicio === h.inicio}<button
											type="button"
											role="radio"
											aria-checked={on}
											aria-label="{h.hora}{h.libre ? '' : ', ocupado'}"
											disabled={!h.libre}
											onclick={() => reserva.elegirHorario(h)}
											style="padding:10px 0;text-align:center;font-weight:700;font-size:14px;text-decoration:{h.libre
												? 'none'
												: 'line-through'};opacity:{h.libre
												? 1
												: 0.4};border-radius:16px;border:1.5px solid {on
												? '#3B2416'
												: 'rgba(59,36,22,.12)'};background:{on
												? '#FBF3E4'
												: 'transparent'};cursor:pointer;font-family:inherit;color:#3B2416"
											>{h.hora}</button
										>{/each}
								</div>
							</div>{/each}
					</div>
				</div>
				{#if reserva.listo}<div
						style="padding:18px;border-radius:24px;background:#fff;display:flex;flex-direction:column;gap:8px"
					>
						<b style="font-size:16px">4 · Tus datos</b>
						<div
							style="display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));gap:8px"
						>
							{#each [['nombre', 'Nombre', 'text', 'name'], ['telefono', 'WhatsApp', 'tel', 'tel'], ['email', 'Email', 'email', 'email']] as [n, l, tipo, ac] (n)}
								<label style="display:block;padding:10px 12px;border-radius:14px;background:#FBF3E4"
									><span style="display:block;font-size:11.5px;color:#7A5A44">{l}</span><input
										name={n}
										type={tipo}
										autocomplete={ac as 'email'}
										required
										maxlength="120"
										style="width:100%;margin-top:2px;padding:0;border:0;background:transparent;font:400 15px 'Instrument Sans';color:#3B2416;outline:none"
									/></label
								>
							{/each}
						</div>
					</div>
					<div
						style="padding:18px;border-radius:24px;background:#fff;display:flex;flex-direction:column;gap:8px"
					>
						<b style="font-size:16px">Pago</b>
						<div
							style="display:flex;align-items:center;gap:12px;padding:15px;text-align:left;border-radius:16px;border:1.5px solid #3B2416;background:#FBF3E4;cursor:pointer;font-family:inherit;color:#3B2416"
						>
							<span
								style="width:20px;height:20px;border-radius:50%;border:2px solid #3B2416;display:grid;place-items:center;flex:none"
								><span style="width:10px;height:10px;border-radius:50%;background:#3B2416"
								></span></span
							>{#if reserva.sena}<Icon
									icon={I.siMercadopago}
									size={22}
									style="color:#00B1EA;flex:none"
								/><span style="flex:1;min-width:0"
									><span style="display:block;font-size:14.5px;font-weight:600">Mercado Pago</span
									><span style="display:block;font-size:12.5px;color:#7A5A44"
										>Seña de {fmt(reserva.sena)} para confirmar · el resto, en el local</span
									></span
								>{:else}<Icon
									icon={I.phStorefrontFill}
									size={22}
									style="color:#3B2416;flex:none"
								/><span style="flex:1;min-width:0"
									><span style="display:block;font-size:14.5px;font-weight:600"
										>Pagás en el local</span
									><span style="display:block;font-size:12.5px;color:#7A5A44"
										>No hace falta seña: el turno queda confirmado</span
									></span
								>{/if}
						</div>
					</div>{/if}
				<div style="padding:18px;border-radius:24px;background:#fff">
					<b style="font-size:16px">Tu reserva</b>
					<div style="margin-top:14px;display:flex;align-items:center;gap:12px">
						{#if reserva.profesional?.foto}<img
								src={img(reserva.profesional.foto, 104, 104)}
								alt=""
								style="width:52px;height:52px;border-radius:50%;object-fit:cover;flex:none"
							/>{/if}
						<div style="min-width:0">
							<div style="font-size:15px;font-weight:700">
								{reserva.conQuien ? `Con ${reserva.conQuien}` : 'Con el primero libre'}
							</div>
							<div style="font-size:13px;color:#7A5A44">
								{reserva.cuando ?? 'Elegí el día y el horario'}
							</div>
						</div>
					</div>
					{#if reserva.servicio}<div
							style="margin-top:16px;padding-top:14px;border-top:1px solid rgba(59,36,22,.1);display:flex;flex-direction:column;gap:8px;font-size:14px"
						>
							<span style="display:flex;justify-content:space-between;gap:10px"
								><span>{reserva.servicio.nombre}</span><span style="white-space:nowrap"
									>{fmt(reserva.servicio.precio)}</span
								></span
							><span style="display:flex;justify-content:space-between;color:#7A5A44"
								><span>Duración total</span><span>{reserva.duracion}</span></span
							><span
								style="display:flex;justify-content:space-between;align-items:baseline;margin-top:6px"
								><b>Total</b><b style="{B};font-size:28px;font-weight:800"
									>{fmt(reserva.servicio.precio)}</b
								></span
							>
						</div>{/if}
					<div
						style="margin-top:14px;display:flex;gap:8px;font-size:13px;line-height:1.45;color:#7A5A44"
					>
						<Icon icon={I.phShieldCheckFill} size={16} style="flex:none;margin-top:1px" /><span
							>{#if reserva.sena}Seña de {fmt(reserva.sena)} para confirmar. Podés cancelar desde el link
								del turno hasta {anticipacion} antes; la seña te la devuelve la&nbsp;tienda.{:else}No
								pagás nada ahora. Podés cancelar desde el link del turno hasta {anticipacion}&nbsp;antes.{/if}</span
						>
					</div>
				</div>
				{#if error}<div
						role="alert"
						style="padding:12px 14px;border-radius:16px;background:#F6E2D2;color:#9C3B1F;font-size:14px;font-weight:600"
					>
						{error}
					</div>{/if}
				<button
					disabled={!reserva.listo || reservando || tienda.vitrina}
					style="padding:18px;border:0;border-radius:999px;background:#3B2416;color:#FBF3E4;font:700 16px 'Instrument Sans';cursor:pointer"
					>{tienda.vitrina
						? 'Es una tienda de ejemplo'
						: !reserva.listo
							? 'Elegí servicio y horario'
							: reserva.sena
								? 'Confirmar y pagar seña'
								: 'Confirmar turno'}</button
				>
			</form>
		{/if}

		<!-- ============ TURNO (con las piezas de la confirmación del pedido) ============ -->
		{#if pagina === 'turno' && turno}
			{@const cancelado = turno.estado === 'cancelado'}
			<section
				style="padding:clamp(32px,5cqw,64px) clamp(14px,3.5cqw,40px) clamp(48px,7cqw,90px);max-width:760px;margin:0 auto;text-align:center"
			>
				<span
					style="display:inline-grid;place-items:center;width:64px;height:64px;border-radius:50%;background:#3B2416;color:#F5C26B"
					><Icon icon={cancelado ? I.phCalendarX : I.phCheckBold} size={28} /></span
				>
				<h1
					style="margin:16px 0 0;{B};font-weight:800;font-size:clamp(40px,5.4cqw,68px);line-height:.92;letter-spacing:-.05em;overflow-wrap:break-word;max-width:100%"
				>
					{cancelado
						? 'Turno cancelado'
						: turno.estado === 'reservado'
							? `Falta la seña, ${turno.nombre}`
							: `¡Listo, ${turno.nombre}!`}
				</h1>
				<p style="margin:16px 0 0;font-size:16px;line-height:1.55;color:#5E4332">
					{#if cancelado}Liberamos el horario{turno.senaPagada
							? '. La seña te la devuelve la tienda por Mercado Pago'
							: ''}. Si querés, reservá&nbsp;otro.{:else if turno.estado === 'reservado'}Te
						guardamos el horario unos minutos mientras pagás la seña con Mercado&nbsp;Pago.{:else}Tu
						turno está confirmado. Te esperamos en&nbsp;{turno.tienda}.{/if}
				</p>
				{#if aviso === 'pendiente' || aviso === 'error'}<div
						role="status"
						style="margin-top:16px;padding:12px 14px;border-radius:16px;background:#F6E2D2;color:#9C3B1F;font-size:14px;font-weight:600"
					>
						{aviso === 'pendiente'
							? 'Mercado Pago todavía no acreditó la seña: te avisamos cuando entre.'
							: 'El pago de la seña no salió. Podés reservar de nuevo.'}
					</div>{/if}
				<div
					style="margin-top:16px;padding:18px;border-radius:24px;background:#fff;text-align:left;font-size:14.5px;line-height:1.6"
				>
					<b style="text-decoration:{cancelado ? 'line-through' : 'none'}"
						>{cuandoLargo(turno.inicio)} con {turno.profesional}</b
					>
					<div style="margin-top:4px">
						{turno.servicio} · {fmt(turno.precio)}{turno.sena
							? ` · seña ${fmt(turno.sena)}${turno.senaPagada ? ' pagada' : ''}`
							: ' · pagás en el local'}
					</div>
					{#if turno.direccion}<div style="margin-top:8px;display:flex;gap:8px;align-items:center">
							<Icon
								icon={I.phMapPinFill}
								size={16}
								style="color:#3B2416;flex:none"
							/>{turno.direccion}
						</div>{/if}
				</div>
				{#if error}<div
						role="alert"
						style="margin-top:16px;padding:12px 14px;border-radius:16px;background:#F6E2D2;color:#9C3B1F;font-size:14px;font-weight:600"
					>
						{error}
					</div>{/if}
				<div style="margin-top:22px;display:flex;justify-content:center;gap:10px;flex-wrap:wrap">
					{#if !cancelado}<a
							href="{base}/turno/{turno.codigo}/turno.ics"
							download="turno.ics"
							data-sveltekit-reload
							style="display:flex;align-items:center;gap:8px;padding:15px 24px;border-radius:999px;background:#3B2416;color:#FBF3E4;font:700 15px 'Instrument Sans';text-decoration:none;border:0"
							><Icon icon={I.phCalendarPlusBold} size={17} />Agregar al calendario</a
						>{/if}
					{#if turno.cancelable}<form
							method="POST"
							action="?/cancelar"
							use:enhance={() => {
								cancelando = true;
								return async ({ update }) => {
									cancelando = false;
									seguro = false;
									await update();
								};
							}}
						>
							{#if seguro}<button
									disabled={cancelando}
									style="cursor:pointer;padding:15px 24px;border-radius:999px;background:#3B2416;color:#FBF3E4;font:700 15px 'Instrument Sans';text-decoration:none;border:0"
									>Sí, cancelar el turno</button
								>{:else}<button
									type="button"
									onclick={() => (seguro = true)}
									style="cursor:pointer;padding:15px 24px;border-radius:999px;border:1.5px solid rgba(59,36,22,.12);background:transparent;color:#3B2416;font:700 15px 'Instrument Sans';text-decoration:none"
									>Cancelar turno</button
								>{/if}
						</form>{:else if cancelado}<a
							href="{base}/turnos"
							style="padding:15px 24px;border-radius:999px;background:#3B2416;color:#FBF3E4;font:700 15px 'Instrument Sans';text-decoration:none;border:0"
							>Reservar otro</a
						>{/if}
					<a
						href={base || '/'}
						style="padding:15px 24px;border-radius:999px;border:1.5px solid rgba(59,36,22,.12);background:transparent;color:#3B2416;font:700 15px 'Instrument Sans';text-decoration:none"
						>Volver a la tienda</a
					>
				</div>
				{#if !cancelado && !turno.cancelable}<p
						style="margin:16px 0 0;font-size:16px;line-height:1.55;color:#5E4332"
					>
						Para cambiarlo o cancelarlo, escribile a la&nbsp;tienda{#if whatsapp}&nbsp;<a
								href={wa(`¡Hola! Tengo un turno el ${cuandoLargo(turno.inicio)}.`)}
								target="_blank"
								rel="noopener"
								style="color:#3B2416">por WhatsApp</a
							>{/if}.
					</p>{/if}
			</section>
		{/if}

		<!-- ============ SUSCRIPCIÓN (un plan y su alta, con las piezas del producto y del checkout) ============ -->
		{#if pagina === 'suscripcion' && plan && alta && caja}
			<form
				method="POST"
				use:alta.form
				style="padding:clamp(18px,3cqw,32px) clamp(14px,3.5cqw,40px) clamp(48px,7cqw,90px);display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,380px),1fr));gap:clamp(20px,4cqw,56px);align-items:start"
			>
				<input type="hidden" name="entrega" value={alta.entrega} />
				<div style="display:flex;flex-direction:column;gap:10px;min-width:0">
					<div
						style="position:relative;aspect-ratio:1;border-radius:30px;overflow:hidden;background:#E9C98F"
					>
						{#if plan.foto ?? caja.fotos[0]}<img
								src={img(plan.foto ?? caja.fotos[0], 1000, 1000)}
								alt={plan.nombre}
								style="width:100%;height:100%;object-fit:cover"
							/>{/if}{#if caja.pct}<span
								style="position:absolute;top:14px;left:14px;padding:4px 10px;border-radius:999px;background:#FBF3E4;font-size:11.5px;font-weight:700"
								>−{caja.pct}%</span
							>{/if}
					</div>
					{#if caja.fotos.length > 1}<div
							style="display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:10px"
						>
							{#each caja.fotos.slice(0, 4) as f, i (i)}<div
									style="aspect-ratio:1;border-radius:18px;overflow:hidden;background:#E9D6B8"
								>
									<img
										src={img(f, 200, 200)}
										alt=""
										style="display:block;width:100%;height:100%;object-fit:cover"
									/>
								</div>{/each}
						</div>{/if}
				</div>
				<div style="position:sticky;top:80px;min-width:0;display:flex;flex-direction:column">
					<div style="font-size:13px;color:#7A5A44">
						<a href={base || '/'} style="color:inherit;text-decoration:none">Inicio</a> / Suscripción
					</div>
					<span
						style="align-self:flex-start;display:inline-flex;margin-top:12px;padding:5px 11px;border-radius:999px;background:#F1E2C8;font-size:12.5px;font-weight:700"
						>{entregasDe(plan)}</span
					>
					<h1
						style="margin:12px 0 0;{B};font-weight:800;font-size:clamp(42px,5cqw,68px);line-height:.92;letter-spacing:-.045em;overflow-wrap:break-word"
					>
						{plan.nombre}
					</h1>
					{#if plan.descripcion}<p
							style="margin:12px 0 0;font-size:16px;line-height:1.55;color:#5E4332;text-wrap:pretty"
						>
							{plan.descripcion}
						</p>{/if}
					{#if planes.length > 1}<div style="margin-top:18px">
							<div style="font-size:13.5px;font-weight:700;margin-bottom:8px">Plan</div>
							<div style="display:flex;gap:8px;flex-wrap:wrap">
								{#each planes as p (p.id)}{@const on = p.id === plan.id}<a
										href="{base}/suscripciones/{p.id}"
										aria-current={on ? 'page' : undefined}
										style="display:inline-flex;align-items:center;min-height:44px;padding:10px 16px;border-radius:999px;background:{on
											? '#3B2416'
											: '#fff'};color:{on
											? '#FBF3E4'
											: '#3B2416'};font:700 14px 'Instrument Sans';text-decoration:none"
										>{p.nombre}</a
									>{/each}
							</div>
						</div>{/if}
					<div
						style="margin-top:18px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,120px),1fr));gap:8px"
					>
						{#each [['Te llega', CADA[plan.frecuencia]], ['Incluye', caja.unidades ? `${caja.unidades} ${caja.unidades === 1 ? 'producto' : 'productos'}` : 'Lo de abajo'], ['Entrega', plan.envio ? 'Envío incluido' : 'Retiro en el local']] as [l, v] (l)}<div
								style="padding:12px;border-radius:16px;background:#fff"
							>
								<div style="font-size:11.5px;color:#7A5A44;font-weight:600">{l}</div>
								<div style="font-size:14.5px;font-weight:700;margin-top:2px">{v}</div>
							</div>{/each}
					</div>
					{#if queIncluye(plan)}<div
							style="margin-top:10px;font-size:14px;line-height:1.5;color:#5E4332;text-wrap:pretty"
						>
							{queIncluye(plan)}
						</div>{/if}
					<div
						style="margin-top:16px;padding:14px;border-radius:18px;background:#3B2416;color:#FBF3E4;display:flex;align-items:center;gap:12px"
					>
						<Icon icon={I.phArrowsClockwiseBold} size={22} style="color:#F5C26B;flex:none" />
						<div style="flex:1;min-width:0">
							<b style="font-size:14.5px"
								>{caja.pct ? `Suscribite y ahorrá ${caja.pct}%` : 'Suscribite y olvidate'}</b
							>
							<div style="font-size:12.5px;color:#E6D3BC">
								Te llega {CADA[plan.frecuencia]} · pausás cuando&nbsp;quieras
							</div>
						</div>
					</div>
					<div style="margin-top:18px;display:flex;align-items:baseline;gap:10px;flex-wrap:wrap">
						<b style="{B};font-size:36px;font-weight:800;letter-spacing:-.03em"
							>{fmt(plan.precio)}</b
						>{#if caja.sueltos}<s style="color:#9C8270">{fmt(caja.sueltos)}</s>{/if}
					</div>
					<div style="font-size:13.5px;color:#5E4332">
						{porPeriodo(plan.frecuencia)}, con débito automático de Mercado&nbsp;Pago
					</div>
					<div
						style="margin-top:16px;padding:18px;border-radius:24px;background:#fff;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,200px),1fr));gap:10px"
					>
						<b style="grid-column:1/-1;font-size:16px">Tus datos</b>
						{#each [['email', 'Email', 'email', 'email'], ['nombre', 'Nombre y apellido', 'text', 'name'], ['telefono', 'WhatsApp', 'tel', 'tel']] as [n, l, tipo, ac] (n)}
							<label style="display:block;padding:10px 12px;border-radius:14px;background:#FBF3E4"
								><span style="display:block;font-size:11.5px;color:#7A5A44">{l}</span><input
									name={n}
									type={tipo}
									autocomplete={ac as 'email'}
									required
									maxlength="120"
									style="width:100%;margin-top:2px;padding:0;border:0;background:transparent;font:400 15px 'Instrument Sans';color:#3B2416;outline:none"
								/></label
							>
						{/each}
					</div>
					<div
						style="margin-top:10px;padding:18px;border-radius:24px;background:#fff;display:flex;flex-direction:column;gap:8px"
					>
						<b style="font-size:16px">¿Cómo te llega?</b>
						<div
							style="display:flex;flex-direction:column;gap:8px"
							role="radiogroup"
							aria-label="Cómo te llega"
						>
							{#each [['envio', 'Envío', 'A tu casa, ya incluido en el precio', plan.envio, I.phMopedFill], ['retiro', 'Retirá en el local', tienda.envios.retiroDireccion || tienda.ciudad || 'Sin costo', plan.retiro, I.phStorefrontFill]] as const as [id, t, d, hay, ic] (id)}{#if hay}{@const on =
										alta.entrega === id}<button
										type="button"
										role="radio"
										aria-checked={on}
										onclick={() => (alta.entrega = id)}
										style="display:flex;align-items:center;gap:12px;padding:14px;border-radius:16px;border:1.5px solid {on
											? '#3B2416'
											: 'rgba(59,36,22,.12)'};background:{on
											? '#FBF3E4'
											: 'transparent'};text-align:left;cursor:pointer;font-family:inherit;color:#3B2416"
										><Icon icon={ic} size={22} style="color:#C9821A" /><span
											style="flex:1;min-width:0"
											><b style="display:block;font-size:14.5px">{t}</b><span
												style="display:block;font-size:12.5px;color:#7A5A44">{d}</span
											></span
										><span
											style="width:20px;height:20px;border-radius:50%;border:2px solid #3B2416;display:grid;place-items:center;flex:none"
											>{#if on}<span
													style="width:10px;height:10px;border-radius:50%;background:#3B2416"
												></span>{/if}</span
										></button
									>{/if}{/each}
						</div>
						{#if alta.entrega === 'envio'}<div
								style="display:grid;grid-template-columns:minmax(0,130px) minmax(0,1fr);gap:10px"
							>
								<label style="display:block;padding:10px 12px;border-radius:14px;background:#FBF3E4"
									><span style="display:block;font-size:11.5px;color:#7A5A44">Código postal</span
									><input
										name="cp"
										inputmode="numeric"
										maxlength="8"
										autocomplete="postal-code"
										style="width:100%;margin-top:2px;padding:0;border:0;background:transparent;font:400 15px 'Instrument Sans';color:#3B2416;outline:none"
									/></label
								>
								<label style="display:block;padding:10px 12px;border-radius:14px;background:#FBF3E4"
									><span style="display:block;font-size:11.5px;color:#7A5A44"
										>Dirección (calle, número, piso)</span
									><input
										name="direccion"
										maxlength="160"
										autocomplete="street-address"
										required
										style="width:100%;margin-top:2px;padding:0;border:0;background:transparent;font:400 15px 'Instrument Sans';color:#3B2416;outline:none"
									/></label
								>
							</div>{/if}
					</div>
					{#if error}<div
							role="alert"
							style="margin-top:10px;padding:12px 14px;border-radius:16px;background:#F6E2D2;color:#9C3B1F;font-size:14px;font-weight:600"
						>
							{error}
						</div>{/if}
					<button
						disabled={alta.enviando || tienda.vitrina || plan.agotado}
						style="margin-top:14px;padding:15px;border:0;border-radius:999px;background:#3B2416;color:#FBF3E4;font:700 15px 'Instrument Sans';cursor:pointer;display:flex;justify-content:center;align-items:center;gap:10px"
						>{tienda.vitrina
							? 'Es una tienda de ejemplo'
							: plan.agotado
								? 'Sin lugar por ahora'
								: `Suscribirme · ${fmt(plan.precio)}`}<Icon
							icon={I.phArrowRightBold}
							size={16}
						/></button
					>
					<span
						style="margin-top:10px;display:flex;justify-content:center;align-items:center;gap:8px;text-align:center;font-size:13px;color:#7A5A44"
						><Icon icon={I.phShieldCheckFill} size={16} />{demo
							? 'Es la demo: no se cobra nada'
							: 'Débito automático con Mercado Pago · pausás o cancelás cuando quieras'}</span
					>
				</div>
			</form>
		{/if}

		<!-- ============ TU SUSCRIPCIÓN (el link privado; con las piezas de la confirmación) ============ -->
		{#if pagina === 'mi-suscripcion' && suscripcion}
			{@const tx = textosSuscripcion(suscripcion, aviso === 'alta', demo)}
			{@const puede = accionesDe(suscripcion)}
			<section
				style="padding:clamp(32px,5cqw,64px) clamp(14px,3.5cqw,40px) clamp(48px,7cqw,90px);max-width:760px;margin:0 auto;text-align:center"
			>
				<span
					style="display:inline-grid;place-items:center;width:64px;height:64px;border-radius:50%;background:#3B2416;color:#F5C26B"
					><Icon
						icon={suscripcion.estado === 'cancelada'
							? I.phXBold
							: suscripcion.estado === 'pausada'
								? I.phPauseFill
								: I.phCheckBold}
						size={28}
					/></span
				>
				<h1
					style="margin:16px 0 0;{B};font-weight:800;font-size:clamp(40px,5.4cqw,68px);line-height:.92;letter-spacing:-.05em;overflow-wrap:break-word;max-width:100%"
				>
					{tx.titulo}
				</h1>
				<p style="margin:16px 0 0;font-size:16px;line-height:1.55;color:#5E4332;text-wrap:pretty">
					{tx.texto}
				</p>
				<div
					style="margin-top:16px;padding:18px;border-radius:24px;background:#fff;text-align:left;font-size:14.5px;line-height:1.6"
				>
					<b>{suscripcion.plan}</b>
					<div style="margin-top:4px">
						{fmt(suscripcion.precio)}
						{CADA[suscripcion.frecuencia]}{suscripcion.cobros
							? ` · ${suscripcion.cobros} ${suscripcion.cobros === 1 ? 'cobro' : 'cobros'}`
							: ''}
					</div>
					{#if suscripcion.incluye}<div style="color:#7A5A44">{suscripcion.incluye}</div>{/if}
					{#if suscripcion.estado === 'activa' && suscripcion.proximoCobro}<div
							style="margin-top:8px;display:flex;gap:8px;align-items:center"
						>
							<Icon icon={I.phCalendarPlusBold} size={16} style="color:#3B2416;flex:none" />Próximo
							cobro: {diaCobro(suscripcion.proximoCobro)}
						</div>{/if}
					<div style="margin-top:8px;display:flex;gap:8px;align-items:center">
						<Icon
							icon={I.phMapPinFill}
							size={16}
							style="color:#3B2416;flex:none"
						/>{suscripcion.entrega === 'envio'
							? `Te llega a ${suscripcion.direccion}`
							: `Retirás en ${suscripcion.tienda}`}
					</div>
				</div>
				{#if error}<div
						role="alert"
						style="margin-top:16px;padding:12px 14px;border-radius:16px;background:#F6E2D2;color:#9C3B1F;font-size:14px;font-weight:600"
					>
						{error}
					</div>{/if}
				<div style="margin-top:22px;display:flex;justify-content:center;gap:10px;flex-wrap:wrap">
					{#if puede.pausar || puede.reanudar}<form
							method="POST"
							action={puede.pausar ? '?/pausar' : '?/reanudar'}
							use:enhance={() => {
								cambiando = true;
								return async ({ update }) => {
									cambiando = false;
									await update();
								};
							}}
						>
							<button
								disabled={cambiando}
								style="cursor:pointer;padding:15px 24px;border-radius:999px;background:#3B2416;color:#FBF3E4;font:700 15px 'Instrument Sans';border:0"
								>{puede.pausar ? 'Pausar' : 'Reanudar'}</button
							>
						</form>{/if}
					{#if puede.cancelar}<form
							method="POST"
							action="?/cancelar"
							use:enhance={() => {
								cambiando = true;
								return async ({ update }) => {
									cambiando = false;
									seguroBaja = false;
									await update();
								};
							}}
						>
							{#if seguroBaja}<button
									disabled={cambiando}
									style="cursor:pointer;padding:15px 24px;border-radius:999px;background:#3B2416;color:#FBF3E4;font:700 15px 'Instrument Sans';border:0"
									>Sí, darla de baja</button
								>{:else}<button
									type="button"
									onclick={() => (seguroBaja = true)}
									style="cursor:pointer;padding:15px 24px;border-radius:999px;border:1.5px solid rgba(59,36,22,.12);background:transparent;color:#3B2416;font:700 15px 'Instrument Sans'"
									>Dar de baja</button
								>{/if}
						</form>{:else}<a
							href="{base}/suscripciones/{suscripcion.planId}"
							style="padding:15px 24px;border-radius:999px;background:#3B2416;color:#FBF3E4;font:700 15px 'Instrument Sans';text-decoration:none;border:0"
							>Volver a suscribirme</a
						>{/if}
					<a
						href={base || '/'}
						style="padding:15px 24px;border-radius:999px;border:1.5px solid rgba(59,36,22,.12);background:transparent;color:#3B2416;font:700 15px 'Instrument Sans';text-decoration:none"
						>Volver a la tienda</a
					>
				</div>
				{#if whatsapp}<p style="margin:16px 0 0;font-size:16px;line-height:1.55;color:#5E4332">
						¿Dudas? Escribinos&nbsp;<a
							href={wa(`¡Hola! Tengo una consulta sobre mi suscripción a ${suscripcion.plan}.`)}
							target="_blank"
							rel="noopener"
							style="color:#3B2416">por WhatsApp</a
						>.
					</p>{/if}
			</section>
		{/if}

		<!-- ============ NOSOTROS ============ -->
		<!-- ============ NO EXISTE (producto o página que no está) ============ -->
		{#if pagina === 'no-existe'}
			<section style="padding:clamp(40px,7cqw,96px) clamp(16px,4cqw,48px)">
				<VacioTienda {tienda} {tema} motivo="no-existe" />
			</section>
		{/if}

		{#if pagina === 'nosotros'}
			<section
				style="padding:clamp(18px,3cqw,32px) clamp(14px,3.5cqw,40px) 0;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,360px),1fr));gap:clamp(14px,2cqw,24px)"
			>
				<div
					style="padding:clamp(26px,4.5cqw,56px);border-radius:28px;background:var(--acento, #C9821A);color:#2A170C;display:flex;flex-direction:column;justify-content:flex-end;min-height:min(520px,110cqw);min-width:0"
				>
					{#if tienda.ciudad}<div
							style="font-size:13px;font-weight:700;letter-spacing:.08em;text-transform:uppercase"
						>
							{tienda.ciudad}
						</div>{/if}
					<h1
						style="margin:12px 0 0;{B};font-weight:800;font-size:clamp(46px,6.6cqw,92px);line-height:.88;letter-spacing:-.05em;overflow-wrap:break-word"
					>
						{tienda.portada.titulo}
					</h1>
				</div>
				{#if tienda.portada.foto || tienda.fotos[0]}
					<div
						style="border-radius:28px;overflow:hidden;background:#D9B98C;min-height:min(520px,110cqw);position:relative"
					>
						<img
							src={img(tienda.portada.foto ?? tienda.fotos[0], 1000)}
							alt=""
							style="position:absolute;inset:0;width:100%;height:100%;object-fit:cover"
						/>
					</div>
				{/if}
			</section>
			<section
				style="padding:clamp(14px,2cqw,24px) clamp(14px,3.5cqw,40px) 0;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,200px),1fr));gap:clamp(10px,1.4cqw,16px)"
			>
				{#if tienda.productos.length}<div style="padding:22px;border-radius:24px;background:#fff">
						<div style="{B};font-weight:800;font-size:48px;line-height:1;letter-spacing:-.04em">
							{tienda.productos.length}
						</div>
						<div style="font-size:14px;color:#7A5A44;margin-top:6px">productos en el menú</div>
					</div>{/if}
				{#if tienda.cobros.cuotas > 1}<div style="padding:22px;border-radius:24px;background:#fff">
						<div style="{B};font-weight:800;font-size:48px;line-height:1;letter-spacing:-.04em">
							{tienda.cobros.cuotas}
						</div>
						<div style="font-size:14px;color:#7A5A44;margin-top:6px">cuotas sin interés</div>
					</div>{/if}
				{#if tienda.promedio}<div style="padding:22px;border-radius:24px;background:#fff">
						<div style="{B};font-weight:800;font-size:48px;line-height:1;letter-spacing:-.04em">
							{tienda.promedio.toLocaleString('es-AR')}
						</div>
						<div style="font-size:14px;color:#7A5A44;margin-top:6px">de promedio en reseñas</div>
					</div>{/if}
			</section>
			<section
				style="padding:clamp(48px,7cqw,90px) clamp(14px,3.5cqw,40px);display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,340px),1fr));gap:clamp(20px,3cqw,40px)"
			>
				<div style="min-width:0">
					<h2
						style="margin:0;{B};font-weight:800;font-size:clamp(36px,5cqw,60px);letter-spacing:-.045em"
					>
						{tienda.envios.retiro ? 'Vení a vernos' : 'Hablemos'}
					</h2>
					<p style="margin:14px 0 0;font-size:16px;line-height:1.55;color:#5E4332">
						{tienda.nombre}{tienda.ciudad ? `, desde ${tienda.ciudad}` : ''}. {comoLlega(
							tienda.envios
						).replace('Comprás', 'Pedís')}
					</p>
					<div style="margin-top:16px;display:flex;flex-direction:column;gap:10px;font-size:15px">
						{#if tienda.direccion}<span style="display:flex;gap:10px"
								><Icon
									icon={I.phMapPinFill}
									size={19}
									style="color:#C9821A"
								/>{tienda.direccion}</span
							>{/if}{#if tienda.whatsapp}<span style="display:flex;gap:10px"
								><Icon
									icon={I.phWhatsappLogoFill}
									size={19}
									style="color:#25A85A"
								/>{tienda.whatsapp}</span
							>{/if}
					</div>
					{#if whatsapp}<a
							href={wa('¡Hola!')}
							target="_blank"
							rel="noopener"
							style="margin-top:18px;display:inline-flex;align-items:center;gap:8px;padding:14px 22px;border-radius:999px;background:#E8F7EC;color:#1F6F3A;font-size:14.5px;font-weight:700;text-decoration:none"
							><Icon icon={I.phWhatsappLogoFill} size={18} />Escribinos por WhatsApp</a
						>{/if}
				</div>
				{#if tienda.direccion}
					<a
						href="https://www.google.com/maps/search/?api=1&query={encodeURIComponent(
							[tienda.direccion, tienda.ciudad].filter(Boolean).join(', ')
						)}"
						target="_blank"
						rel="noopener"
						style="position:relative;display:block;min-height:340px;border-radius:28px;overflow:hidden;background:#EADBC2;text-decoration:none"
						><span
							style="position:absolute;inset:0;background-image:linear-gradient(rgba(59,36,22,.07) 1px,transparent 1px),linear-gradient(90deg,rgba(59,36,22,.07) 1px,transparent 1px);background-size:32px 32px"
						></span><span
							style="position:absolute;left:-10%;right:-10%;top:52%;height:16px;background:#fff;transform:rotate(-12deg)"
						></span><span
							style="position:absolute;top:-10%;bottom:-10%;left:40%;width:14px;background:#fff;transform:rotate(8deg)"
						></span><span
							style="position:absolute;left:44%;top:40%;max-width:52%;display:flex;align-items:center;gap:8px;padding:8px 12px 8px 8px;border-radius:999px;background:#3B2416;color:#FBF3E4;font-size:13px;font-weight:700;box-shadow:0 10px 20px -6px rgba(59,36,22,.6)"
							><span
								style="width:26px;height:26px;border-radius:50%;background:#F5C26B;color:#2A170C;display:grid;place-items:center;flex:none"
								>{inicial}</span
							><span style="overflow:hidden;text-overflow:ellipsis;white-space:nowrap"
								>{tienda.nombre}</span
							></span
						><span
							style="position:absolute;left:16px;bottom:14px;font-size:12px;font-weight:600;color:#7A5A44"
							>Ver en Google Maps</span
						></a
					>
				{/if}
			</section>
		{/if}

		<!-- footer -->
		<footer
			style="margin:0 clamp(14px,3.5cqw,40px) clamp(14px,3.5cqw,40px);padding:clamp(26px,4cqw,48px);border-radius:28px;background:#3B2416;color:#FBF3E4"
		>
			<div
				style="display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,190px),1fr));gap:28px"
			>
				<div style="min-width:0">
					<div style="{B};font-weight:800;font-size:32px;letter-spacing:-.04em">
						{tienda.nombre}
					</div>
					<p style="margin:10px 0 0;font-size:14px;color:#E6D3BC;line-height:1.5">
						{#if tienda.cobros.cuotas > 1}{tienda.cobros.cuotas} cuotas sin interés{#if tienda.cobros.transferencia > 0}&nbsp;y
								{tienda.cobros.transferencia}% off por transferencia{/if}, en toda la&nbsp;tienda.{:else if tienda.cobros.transferencia > 0}{tienda
								.cobros.transferencia}% off por transferencia, en toda la&nbsp;tienda.{:else}Compra
							protegida, con 10 días para&nbsp;arrepentirte.{/if}
					</p>
					<!-- Newsletter (§50): el bloque del pie del diseño; queda pendiente hasta confirmar por mail. -->
					<form onsubmit={nl.enviar} novalidate style="margin-top:12px;display:flex;gap:6px">
						<input
							class="nl"
							type="email"
							name="email"
							required
							maxlength="120"
							autocomplete="email"
							placeholder="Tu email"
							aria-label="Tu correo para las novedades"
							bind:value={nl.email}
							style="--ph:#C9B29A;flex:1;min-width:0;padding:12px 14px;border:0;border-radius:999px;background:rgba(251,243,228,.1);font:inherit;font-size:14px;color:#FBF3E4;outline:none"
						/>
						<button
							type="submit"
							aria-label="Sumarme"
							disabled={nl.enviando}
							style="flex:none;width:44px;height:44px;padding:0;border:0;border-radius:50%;background:#F5C26B;color:#2A170C;display:grid;place-items:center;cursor:pointer"
							><Icon icon={I.phArrowRightBold} size={16} /></button
						>
					</form>
					{#if nl.aviso}<p
							role="status"
							style="margin:8px 0 0;font-size:13px;line-height:1.45;color:{nl.aviso.ok
								? '#E6D3BC'
								: '#F5B0A0'}"
						>
							{nl.aviso.texto}{#if nl.aviso.link}&nbsp;<a
									href={nl.aviso.link}
									style="color:inherit;text-decoration:underline;text-underline-offset:3px;font-weight:600"
									>Confirmar acá</a
								>{/if}
						</p>{/if}
				</div>
				{#each foot as f (f.t)}<div
						style="display:flex;flex-direction:column;gap:8px;font-size:14px"
					>
						<b style="color:#F5C26B;margin-bottom:4px">{f.t}</b>{#each f.l as l (l.t)}<a
								href={l.href}
								style="color:#E6D3BC;text-decoration:none">{l.t}</a
							>{/each}
					</div>{/each}
			</div>
			<div
				style="margin-top:32px;padding-top:24px;border-top:1px solid color-mix(in srgb,currentColor 16%,transparent);display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,250px),1fr));gap:22px"
			>
				{#if tienda.cobros.mercadoPago || tienda.cobros.datosTransferencia}
					<div style="min-width:0">
						<div
							style="font-size:11.5px;font-weight:700;letter-spacing:.12em;text-transform:uppercase;opacity:.7;margin-bottom:10px"
						>
							Medios de pago
						</div>
						<div class="sellos">
							{#if tienda.cobros.mercadoPago}<span><Icon icon={I.logosVisa} size={34} /></span><span
									><Icon icon={I.logosMastercard} size={22} /></span
								><span><Icon icon={I.logosAmex} size={30} /></span><span
									><Icon icon={I.siMercadopago} size={17} style="color:#00B1EA" /><span
										style="color:#0A2D6E">Mercado Pago</span
									></span
								>{/if}{#if tienda.cobros.datosTransferencia}<span
									><Icon icon={I.phBankFill} size={14} style="color:#2F7A43" />Transferencia</span
								>{/if}
						</div>
					</div>
				{/if}
				<div style="min-width:0">
					<div
						style="font-size:11.5px;font-weight:700;letter-spacing:.12em;text-transform:uppercase;opacity:.7;margin-bottom:10px"
					>
						Medios de envío
					</div>
					<div class="sellos">
						{#if tienda.envios.correo}<span
								><span
									style="width:14px;height:14px;border-radius:3px;background:#FFD100;display:inline-block"
								></span><span style="color:#00205B">Correo Argentino</span></span
							>{/if}{#if tienda.envios.moto}<span
								><Icon icon={I.phMopedFill} size={15} style="color:#1A1A1A" />Envío en moto</span
							>{/if}{#if tienda.envios.retiro}<span
								><Icon icon={I.phStorefrontFill} size={14} style="color:#1A1A1A" />Retiro en el
								local</span
							>{/if}
					</div>
				</div>
				<div style="min-width:0">
					<div
						style="font-size:11.5px;font-weight:700;letter-spacing:.12em;text-transform:uppercase;opacity:.7;margin-bottom:10px"
					>
						Compra segura
					</div>
					<div class="sellos">
						<span
							><Icon icon={I.phLockSimpleFill} size={14} style="color:#2F7A43" />Sitio seguro SSL</span
						><span
							><Icon icon={I.phShieldCheckFill} size={14} style="color:#1A1A1A" />Compra protegida</span
						>
					</div>
				</div>
			</div>
			<div
				style="margin-top:28px;padding-top:16px;border-top:1px solid rgba(251,243,228,.12);display:flex;justify-content:space-between;gap:12px;flex-wrap:wrap;font-size:12.5px;color:#C9B29A"
			>
				<span
					>© {new Date().getFullYear()}
					{tienda.nombre}{tienda.ciudad ? ` · ${tienda.ciudad}` : ''}</span
				>
			</div>
			<div
				style="margin-top:16px;padding-top:16px;border-top:1px solid color-mix(in srgb,currentColor 16%,transparent);display:flex;justify-content:space-between;align-items:center;gap:12px 20px;flex-wrap:wrap;font-size:12.5px"
			>
				<span style="display:flex;gap:6px 16px;flex-wrap:wrap"
					><a
						href="/t/{tienda.slug}/arrepentimiento"
						style="color:inherit;text-decoration:underline;text-underline-offset:3px;font-weight:700;display:inline-flex;align-items:center;gap:5px"
						><Icon icon={I.phArrowCounterClockwiseBold} size={13} />Botón de arrepentimiento</a
					><a
						href="/legales/defensa-del-consumidor"
						style="color:inherit;opacity:0.75;text-decoration:underline;text-underline-offset:3px;font-weight:500"
						>Defensa de las y los consumidores</a
					><a
						href="/legales/terminos"
						style="color:inherit;opacity:0.75;text-decoration:underline;text-underline-offset:3px;font-weight:500"
						>Términos y condiciones</a
					><a
						href="/legales/privacidad"
						style="color:inherit;opacity:0.75;text-decoration:underline;text-underline-offset:3px;font-weight:500"
						>Políticas de privacidad</a
					></span
				>
				<a
					href="/"
					style="display:inline-flex;align-items:center;gap:8px;padding:5px 13px 5px 5px;border-radius:999px;background:#fff;color:#2A140C;text-decoration:none;font:500 12.5px 'Instrument Sans',sans-serif;box-shadow:0 1px 3px rgba(0,0,0,.12);white-space:nowrap"
					><span
						style="width:24px;height:24px;border-radius:50%;background:#FCEDE5;display:grid;place-items:center"
						><img src="/mark.svg" alt="" style="width:13px;height:auto;display:block" /></span
					>Tienda creada con <b style="font-weight:800;letter-spacing:-.02em">Tu Local Web</b></a
				>
			</div>
		</footer>

		{#if whatsapp && pagina !== 'checkout'}<a
				href={wa('¡Hola!')}
				target="_blank"
				rel="noopener"
				aria-label="Escribinos por WhatsApp"
				style="position:fixed;right:18px;bottom:18px;z-index:30;width:54px;height:54px;border-radius:50%;background:#25D366;color:#fff;display:grid;place-items:center;box-shadow:0 12px 26px -8px rgba(0,0,0,.4)"
				><Icon icon={I.phWhatsappLogoFill} size={28} /></a
			>{/if}
	</div>
</div>

<!-- eslint-enable svelte/no-navigation-without-resolve -->

<style>
	.zoom :global(img) {
		transition: transform 0.8s cubic-bezier(0.2, 0.7, 0.1, 1);
	}
	.zoom:hover :global(img) {
		transform: scale(1.04);
	}
	.dos {
		display: -webkit-box;
		-webkit-line-clamp: 2;
		line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}
	.sellos {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
	}
	.sellos > span {
		height: 28px;
		display: inline-flex;
		align-items: center;
		gap: 5px;
		padding: 0 9px;
		border-radius: 6px;
		background: #fff;
		color: #1a1a1a;
		box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.06);
		font:
			700 11.5px 'Instrument Sans',
			sans-serif;
		white-space: nowrap;
	}
	button:disabled {
		opacity: 0.45;
		cursor: not-allowed !important;
	}
	.nl::placeholder {
		color: var(--ph);
		opacity: 1;
	}
</style>
