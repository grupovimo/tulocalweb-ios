<script lang="ts">
	import { camposFactura, OPCIONALES } from '#lib/tienda/factura.ts';
	import { enhance } from '$app/forms';
	import { page } from '$app/state';
	import { untrack } from 'svelte';
	import { goto } from '$app/navigation';
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
		cotizar,
		cotizarYa,
		cpValido,
		entregaCorta,
		envia,
		textoEnvios
	} from '../envio.svelte.ts';
	import { Seleccion, tarjetaDe } from '../seleccion.svelte.ts';
	import { avisoMayorista } from '../lineas.ts';
	import { lineaPorMayor, textoAviso } from '../mayorista.ts';
	import { partirTitulo } from '../titulo.ts';
	import { Reserva } from '../reserva.svelte.ts';
	import { cuandoLargo, duracion } from '#lib/turnos/formato.ts';
	import type { OpcionEnvio, ProductoPublico, TiendaPublica } from '../tipos.ts';
	import { AltaNewsletter } from '../newsletter.svelte.ts';

	/**
	 * Plantilla Rosedal (design/Plantilla Rosedal.dc.html), con los datos de la tienda en lugar de los
	 * de ejemplo. Lo que el diseño muestra sin datos atrás (test de piel, rutina −20 %, reposición
	 * automática, ingredientes, antes y después, muestra gratis, cupón, cuenta regresiva) no se
	 * inventa: el selector de piel pasa a elegir categoría y lo demás no se dibuja. El newsletter del
	 * pie anda de verdad (§50).
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
		| 'turno';
	// eslint-disable-next-line svelte/no-unused-props -- `proximos` (libres de hoy) sólo lo dibuja Tijera
	let {
		tienda,
		pagina,
		producto = null,
		categoria = null,
		orden = 'vendidos',
		pedido = null,
		turno = null,
		aviso = null,
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
	const abSpan = $derived(nar ? 'auto' : '2');

	const BG = ['#F3D9DE', '#F6E6D8', '#E8E3F1', '#E3EEE6'];
	const fmt = (n: number) => '$ ' + Math.round(n).toLocaleString('es-AR');
	const img = (f: string | null, w: number, h?: number) =>
		!f
			? null
			: /^https?:\/\//.test(f)
				? f
				: `https://images.unsplash.com/photo-${f}?auto=format&fit=crop&w=${w}${h ? '&h=' + h : ''}&q=72`;
	const porId = $derived(new Map(tienda.productos.map((p) => [p.id, p])));
	const rank = $derived(new Map(tienda.masVendidos.map((id, i) => [id, i])));
	// Puntaje de cada producto con sus reseñas (las reseñas nombran el producto).
	const puntaje = $derived.by(() => {
		// eslint-disable-next-line svelte/prefer-svelte-reactivity -- local, se arma una vez
		const m = new Map<string, { suma: number; n: number }>();
		for (const r of tienda.resenas) {
			const x = m.get(r.producto) ?? { suma: 0, n: 0 };
			m.set(r.producto, { suma: x.suma + r.estrellas, n: x.n + 1 });
		}
		return m;
	});
	const notaDe = (nombre: string) => {
		const x = puntaje.get(nombre);
		return x ? { r: (Math.round((x.suma / x.n) * 10) / 10).toLocaleString('es-AR'), n: x.n } : null;
	};
	// Con variantes de precios distintos, "Desde"; "Sin stock" sólo si no queda ninguna.
	const tarjeta = (p: ProductoPublico, i: number) => {
		const t = tarjetaDe(p);
		const old = p.anterior && !t.desde ? p.anterior : null;
		return {
			id: p.id,
			n: p.nombre,
			price: (t.desde ? 'Desde ' : '') + fmt(t.precio),
			old: old ? fmt(old) : null,
			tr: fmt(t.precio * (1 - tienda.cobros.transferencia / 100)),
			img: img(p.foto, 600, 750),
			tag: t.agotado
				? 'Sin stock'
				: old
					? `−${Math.round((1 - t.precio / old) * 100)}%`
					: p.etiqueta,
			for: p.categoria ? [p.categoria] : [],
			nota: notaDe(p.nombre),
			bg: BG[i % 4],
			href: `${base}/producto/${p.slug ?? p.id}`
		};
	};
	const ver = (s: string) => tienda.secciones.includes(s as TiendaPublica['secciones'][number]);
	const whatsapp = $derived(tienda.whatsapp.replace(/\D/g, ''));
	const wa = (texto: string) =>
		`https://wa.me/${whatsapp.startsWith('54') ? whatsapp : '54' + whatsapp}?text=${encodeURIComponent(texto)}`;

	// El logo del diseño es "rosa" en serif y "mosqueta" en versalitas: primera palabra y el resto.
	const logo = $derived.by(() => {
		const [a, ...b] = tienda.nombre.trim().split(/\s+/);
		return [a, b.join(' ')] as const;
	});
	// Título de la portada: el diseño lo corta en dos renglones.
	const titulo = $derived(partirTitulo(tienda.portada.titulo));

	const perks = $derived([
		...(tienda.envios.gratisDesde
			? [{ i: I.phTruck, t: 'Envío gratis', d: `Desde ${fmt(tienda.envios.gratisDesde)}` }]
			: []),
		...(tienda.cobros.cuotas > 1
			? [
					{
						i: I.phCreditCard,
						t: `${tienda.cobros.cuotas} cuotas sin interés`,
						d: 'Con todas las tarjetas'
					}
				]
			: []),
		...(tienda.cobros.transferencia > 0
			? [{ i: I.phBank, t: `${tienda.cobros.transferencia}% off`, d: 'Pagando por transferencia' }]
			: []),
		tienda.envios.retiro
			? { i: I.phStorefrontFill, t: 'Retiro en el local', d: tienda.ciudad || 'Sin costo' }
			: { i: I.phArrowsClockwise, t: 'Arrepentimiento', d: 'Tenés 10 días para devolverlo' }
	]);
	const ordenados = $derived(
		[...tienda.productos].sort((a, b) => (rank.get(a.id) ?? 999) - (rank.get(b.id) ?? 999))
	);
	const best = $derived(
		tienda.masVendidos
			.slice(0, 4)
			.flatMap((id, k) => (porId.get(id) ? [tarjeta(porId.get(id)!, k)] : []))
	);

	// Portada: el selector "Mi piel es" del diseño elige categoría y muestra sus tres más vendidos.
	const catsCon = $derived(tienda.categorias.filter((c) => c.cantidad > 0).map((c) => c.nombre));
	let elegida = $state<string | null>(null);
	const activa = $derived(elegida && catsCon.includes(elegida) ? elegida : (catsCon[0] ?? null));
	const rutina = $derived(
		ordenados
			.filter((p) => !activa || p.categoria === activa)
			.slice(0, 3)
			.map((p, k) => ({ ...tarjeta(p, k), num: '0' + (k + 1) }))
	);
	const cantActiva = $derived(
		activa ? tienda.productos.filter((p) => p.categoria === activa).length : tienda.productos.length
	);

	// Catálogo: categoría, orden y de a 24.
	const ORDENES = [
		['vendidos', 'Más vendidos'],
		['menor', 'Menor precio'],
		['mayor', 'Mayor precio'],
		['az', 'De la A a la Z']
	] as const;
	const listado = $derived.by(() => {
		const l = tienda.productos.filter((p) => !categoria || p.categoria === categoria);
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
	const catalog = $derived(listado.slice(0, mostrar).map(tarjeta));
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
						'Tenés 10 días desde que lo recibís para arrepentirte de la compra, sin dar explicaciones (botón de arrepentimiento, abajo de todo).'
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
	const notaFicha = $derived(producto ? notaDe(producto.nombre) : null);
	// Tono, tamaño, presentación…: lo elegido, la variante, su precio, stock y foto (seleccion.svelte.ts).
	const sel = $derived(producto ? new Seleccion(producto) : null);
	// El tachado no va con "Desde": no se sabe a qué presentación se compara.
	const tachado = $derived(sel && !sel.desde ? sel.anterior : null);
	let cp = $state('');
	let cotizado = $state<OpcionEnvio[] | null>(null);
	async function calcular() {
		if (cpValido(cp)) cotizado = await cotizarYa(cp, sel?.precio ?? subtotal, tienda);
	}
	function agregar() {
		if (!sel || tienda.vitrina || !sel.agregar(carrito)) return;
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- ruta de la tienda con su slug
		goto(`${base}/carrito`);
	}
	/** Sumar desde "Te puede gustar": con opciones, primero se elige en la ficha. */
	function sumarRapido(p: ProductoPublico) {
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- ruta de la tienda con su slug
		if (!carrito.sumarProducto(p)) goto(`${base}/producto/${p.slug ?? p.id}`);
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
	const elegidaEnvio = $derived(opciones.find((o) => o.id === envio) ?? null);
	const costoEnvio = $derived(elegidaEnvio?.precio ?? 0);
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
	const ch = (on: boolean) => ({
		bd: on ? 'var(--acento, #B4475B)' : '#EAD3CF',
		bg: on ? '#FFF6F7' : '#fff',
		rc: on ? 'var(--acento, #B4475B)' : '#D9BEC3'
	});
	const pagos = $derived([
		...(tienda.cobros.datosTransferencia
			? [
					{
						id: 'transferencia' as const,
						t: 'Transferencia',
						d: tienda.cobros.transferencia
							? `${tienda.cobros.transferencia}% off · alias al confirmar`
							: 'Alias al confirmar',
						i: I.phBankFill,
						c: '#2F7D55'
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
	const pasos = $derived([
		'var(--acento, #B4475B)',
		envio ? 'var(--acento, #B4475B)' : '#EAD3CF',
		'#EAD3CF'
	]);

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
		...tienda.categorias.slice(0, 3).map((c) => ({
			t: c.nombre,
			href: enlaceCat(c.nombre),
			on: pagina === 'catalogo' && categoria === c.nombre
		})),
		{ t: 'Todo', href: `${base}/catalogo`, on: pagina === 'catalogo' && !categoria },
		...(tienda.turnos
			? [
					{
						t: 'Reservar turno',
						href: `${base}/turnos`,
						on: pagina === 'turnos' || pagina === 'turno'
					}
				]
			: []),
		{ t: 'Nosotros', href: `${base}/nosotros`, on: pagina === 'nosotros' }
	]);
	// Reseñas y preguntas con las piezas de Rosedal: tarjetas blancas, botones de píldora, rosa.
	const tema: TemaOpiniones = {
		texto: '#3A1F25',
		suave: '#8C6A70',
		campo: '#FBF4F1',
		tarjeta: '#fff',
		borde: '#EAD3CF',
		acento: '#3A1F25',
		sobreAcento: '#fff',
		estrella: 'var(--acento, #B4475B)',
		radio: '14px',
		radioBoton: '999px',
		fuente: "'Hanken Grotesk', system-ui, sans-serif",
		titulos: "'Hanken Grotesk', system-ui, sans-serif",
		pesoTitulos: 800
	};
	const estrellas = (n: number) =>
		Array.from({ length: 5 }, (_, i) =>
			n >= i + 1 ? I.phStarFill : n >= i + 0.5 ? I.phStarHalfFill : I.phStar
		);
	const stats = $derived(
		[
			...(tienda.productos.length ? [{ v: String(tienda.productos.length), t: 'productos' }] : []),
			...(tienda.categorias.length
				? [
						{
							v: String(tienda.categorias.length),
							t: tienda.categorias.length === 1 ? 'categoría' : 'categorías'
						}
					]
				: []),
			...(tienda.cobros.cuotas > 1
				? [{ v: String(tienda.cobros.cuotas), t: 'cuotas sin interés' }]
				: []),
			...(tienda.promedio
				? [
						{
							v: tienda.promedio.toLocaleString('es-AR'),
							t: `en ${tienda.resenas.length} ${tienda.resenas.length === 1 ? 'reseña' : 'reseñas'}`
						}
					]
				: [])
		].map((s, k) => ({ ...s, bg: BG[k % 4] }))
	);
	const fotoNos = $derived(tienda.portada.foto ?? tienda.fotos[0] ?? null);
	const fotoCats = $derived(
		tienda.portada.foto ?? tienda.categorias.find((c) => c.foto)?.foto ?? null
	);
	// ---- Turnos (reserva y confirmación, con las piezas del checkout y de 'listo') ----
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
	const PASOS_TURNO = ['Servicio', 'Profesional', 'Horario', 'Pago'];
	const anticipacion = $derived.by(() => {
		const m = tienda.turnos?.anticipacionMin ?? 0;
		return m >= 120 && m % 60 === 0 ? `${m / 60} h` : `${m} min`;
	});
	let reservando = $state(false);
	let cancelando = $state(false);
	let seguro = $state(false);
</script>

<svelte:window bind:innerWidth={ancho} />

<svelte:head>
	<link
		href="https://fonts.googleapis.com/css2?family=Gloock&family=Hanken+Grotesk:wght@400;500;600;700;800&display=swap"
		rel="stylesheet"
	/>
</svelte:head>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
<div
	class="rosedal"
	style="min-height:100vh;font-family:'Hanken Grotesk',system-ui,sans-serif;color:#3A1F25"
>
	<div
		style="container-type:inline-size;width:100%;background:#FBF4F1;position:relative;overflow-x:clip"
	>
		<!-- barra promo -->
		<div
			style="display:flex;justify-content:center;align-items:center;gap:12px;flex-wrap:wrap;padding:9px 16px;background:#F3D9DE;color:#3A1F25;font-size:13px;font-weight:600;text-align:center"
		>
			{#if conteo.actual}<span>{conteo.texto}</span><span
					style="display:flex;gap:4px;font-variant-numeric:tabular-nums"
					>{#each conteo.partes(3) as c (c.u)}<span
							style="padding:2px 7px;border-radius:999px;background:#fff">{c.v}{c.u}</span
						>{/each}</span
				>{:else}<span
					>{#if tienda.envios.gratisDesde || tienda.cobros.cuotas > 1}{tienda.envios.gratisDesde
							? `Envío gratis desde ${fmt(tienda.envios.gratisDesde)}`
							: `${tienda.cobros.cuotas} cuotas sin interés`}{#if tienda.cobros.transferencia > 0}&nbsp;·
							{tienda.cobros.transferencia}% off por transferencia{/if}{:else if tienda.cobros.transferencia > 0}{tienda
							.cobros.transferencia}% off por transferencia{:else}Compra protegida{/if}</span
				>{/if}
		</div>

		<!-- header -->
		<header
			style="position:sticky;top:0;z-index:20;display:flex;align-items:center;gap:18px;padding:14px clamp(14px,3.5cqw,44px);background:rgba(251,244,241,.9);backdrop-filter:blur(12px)"
		>
			{#if nar}<a href="{base}/catalogo" aria-label="Catálogo" style="display:grid;color:#3A1F25"
					><Icon icon={I.phList} size={22} /></a
				>{/if}
			<a
				href={base || '/'}
				style="text-decoration:none;color:#3A1F25;display:flex;align-items:baseline;gap:6px;min-width:0"
				><LogoTienda {tienda}
					><span style="font-family:'Gloock',serif;font-size:25px;line-height:1">{logo[0]}</span
					>{#if logo[1]}<span
							style="font-size:11px;font-weight:800;letter-spacing:.3em;text-transform:uppercase;color:var(--acento, #B4475B)"
							>{logo[1]}</span
						>{/if}</LogoTienda
				></a
			>
			{#if wide}<nav style="display:flex;gap:4px;margin-left:14px">
					{#each nav as n (n.t)}<a
							href={n.href}
							style="text-decoration:none;font:600 14px 'Hanken Grotesk';padding:9px 14px;border-radius:999px;background:{n.on
								? '#F3D9DE'
								: 'transparent'};color:#3A1F25">{n.t}</a
						>{/each}
				</nav>{/if}
			<span style="flex:1"></span>
			{#if nar && tienda.turnos}<a
					href="{base}/turnos"
					aria-label="Reservar turno"
					style="display:grid;color:#3A1F25"><Icon icon={I.phCalendarPlusBold} size={22} /></a
				>{/if}
			<a
				href="{base}/carrito"
				aria-label="Carrito"
				style="position:relative;color:#3A1F25;display:grid"
				><Icon icon={I.phToteSimple} size={22} />{#if carrito.unidades}<span
						style="position:absolute;top:-6px;right:-8px;min-width:17px;height:17px;border-radius:999px;background:var(--acento, #B4475B);color:#fff;font-size:10.5px;font-weight:700;display:grid;place-items:center"
						>{carrito.unidades}</span
					>{/if}</a
			>
		</header>

		{#if tienda.vitrina}
			<div
				style="padding:9px 16px;background:#fff;color:#3A1F25;font-size:13px;font-weight:600;text-align:center"
			>
				Es una tienda de ejemplo de la plantilla Rosedal: podés recorrerla, pero no se compra de
				verdad.
			</div>
		{/if}

		<!-- ============ INICIO ============ -->
		{#if pagina === 'inicio'}
			<!-- Las secciones en el orden que eligió la tienda en Mi tienda. -->
			{#each enOrden( tienda.secciones, ['hero', 'cats', 'prods', 'revs', 'ig'] ) as seccion (seccion)}
				{#if seccion === 'hero'}
					{#if ver('hero')}
						<section
							style="position:relative;padding:clamp(30px,5cqw,64px) clamp(14px,3.5cqw,44px) 0;text-align:center;overflow:hidden"
						>
							<div
								style="position:absolute;left:50%;top:-30%;width:min(1100px,140%);aspect-ratio:1;transform:translateX(-50%);border-radius:50%;background:radial-gradient(closest-side,#F6D5DB,rgba(246,213,219,0))"
							></div>
							<div style="position:relative">
								{#if tienda.ciudad}<div
										style="font-size:12.5px;font-weight:800;letter-spacing:.24em;text-transform:uppercase;color:var(--acento, #B4475B)"
									>
										{tienda.ciudad}
									</div>{/if}
								<h1
									style="margin:16px auto 0;max-width:980px;font-family:'Gloock',serif;font-weight:400;font-size:clamp(48px,8cqw,118px);line-height:.94;letter-spacing:-.02em;overflow-wrap:anywhere"
								>
									{titulo[0]}{#if titulo[1]}<br />{titulo[1]}{/if}
								</h1>
								{#if tienda.portada.lema}<p
										style="margin:18px auto 0;max-width:540px;font-size:17px;line-height:1.55;color:#6E4F55"
									>
										{tienda.portada.lema}
									</p>{/if}
								{#if catsCon.length > 1 && rutina.length}
									<div
										style="margin:26px auto 0;max-width:720px;padding:8px;border-radius:999px;background:#fff;box-shadow:0 20px 50px -30px rgba(58,31,37,.5);display:flex;gap:4px;overflow-x:auto;scrollbar-width:none"
									>
										<span
											style="flex:none;display:flex;align-items:center;padding:0 14px;font-size:13.5px;font-weight:700;white-space:nowrap"
											>Buscás</span
										>
										{#each catsCon as c (c)}{@const on = c === activa}<button
												type="button"
												onclick={() => (elegida = c)}
												aria-pressed={on}
												style="flex:1 0 auto;border:0;cursor:pointer;padding:12px 16px;border-radius:999px;background:{on
													? '#3A1F25'
													: 'transparent'};color:{on
													? '#fff'
													: '#3A1F25'};font:700 14px 'Hanken Grotesk';white-space:nowrap;transition:background .25s"
												>{c}</button
											>{/each}
									</div>
								{/if}
							</div>
							{#if rutina.length}
								<div
									style="position:relative;margin-top:clamp(28px,4cqw,48px);display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,230px),1fr));gap:12px;text-align:left"
								>
									{#each rutina as r (r.id)}
										<a
											href={r.href}
											style="position:relative;border:0;padding:0;border-radius:28px;overflow:hidden;background:{r.bg};text-align:left;font-family:inherit;color:#3A1F25;min-height:380px;display:flex;flex-direction:column;text-decoration:none"
										>
											<span
												style="display:flex;justify-content:space-between;align-items:center;gap:8px;padding:16px 18px 0"
												><span
													style="font-family:'Gloock',serif;font-size:44px;line-height:1;color:var(--acento, #B4475B)"
													>{r.num}</span
												>{#if r.tag}<span
														style="padding:6px 11px;border-radius:999px;background:rgba(255,255,255,.7);font-size:12px;font-weight:800;letter-spacing:.08em;text-transform:uppercase"
														>{r.tag}</span
													>{/if}</span
											>
											<span
												style="flex:1;position:relative;margin:10px 18px 0;border-radius:20px;overflow:hidden;background:rgba(255,255,255,.4);min-height:200px"
												>{#if r.img}<img
														src={r.img}
														alt=""
														style="position:absolute;inset:0;width:100%;height:100%;object-fit:cover"
													/>{/if}</span
											>
											<span
												style="display:flex;justify-content:space-between;align-items:flex-end;gap:10px;padding:14px 18px 18px"
												><span style="min-width:0"
													><span style="display:block;font-size:16px;font-weight:700">{r.n}</span
													>{#if tienda.cobros.transferencia > 0}<span
															style="display:block;font-size:13px;color:#6E4F55;margin-top:2px"
															>{r.tr} por transferencia</span
														>{/if}</span
												><b style="font-size:16px;white-space:nowrap">{r.price}</b></span
											>
										</a>
									{/each}
								</div>
								<div
									style="position:relative;margin-top:14px;display:flex;justify-content:center;align-items:center;gap:14px;flex-wrap:wrap;padding:14px 18px;border-radius:999px;background:#3A1F25;color:#FBF4F1;font-size:14px"
								>
									<span
										>{#if activa}<b>{activa}</b> ·
										{/if}{cantActiva}
										{cantActiva === 1 ? 'producto' : 'productos'}</span
									><a
										href={activa ? enlaceCat(activa) : `${base}/catalogo`}
										style="padding:10px 18px;border:0;border-radius:999px;background:#F3D9DE;color:#3A1F25;font:800 13.5px 'Hanken Grotesk';text-decoration:none"
										>Ver {activa ? activa.toLowerCase() : 'todo'}</a
									>
								</div>
							{/if}
						</section>
					{/if}

					<div
						style="margin-top:clamp(40px,6cqw,72px);display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));gap:1px;background:#EAD3CF;border-top:1px solid #EAD3CF;border-bottom:1px solid #EAD3CF"
					>
						{#each perks as p (p.t)}<div
								style="display:flex;align-items:center;gap:12px;padding:18px clamp(14px,3cqw,32px);background:#FBF4F1"
							>
								<Icon icon={p.i} size={22} style="color:var(--acento, #B4475B)" />
								<div>
									<div style="font-size:14px;font-weight:700">{p.t}</div>
									<div style="font-size:12.5px;color:#8C6A70">{p.d}</div>
								</div>
							</div>{/each}
					</div>
				{:else if seccion === 'cats'}
					{#if ver('cats') && tienda.categorias.length}
						<section
							style="padding:clamp(56px,8cqw,104px) clamp(14px,3.5cqw,44px) 0;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,380px),1fr));gap:clamp(20px,4cqw,56px);align-items:center"
						>
							<div
								style="position:relative;aspect-ratio:1;border-radius:50%;overflow:hidden;background:#F3D9DE;max-width:520px;width:100%;justify-self:center"
							>
								{#if fotoCats}<img
										src={img(fotoCats, 1000, 1000)}
										alt=""
										style="width:100%;height:100%;object-fit:cover"
									/>{/if}
							</div>
							<div>
								<div
									style="font-size:12.5px;font-weight:800;letter-spacing:.24em;text-transform:uppercase;color:var(--acento, #B4475B)"
								>
									Categorías
								</div>
								<h2
									style="margin:12px 0 0;font-size:clamp(34px,4.6cqw,58px);line-height:1;letter-spacing:-.04em;font-weight:800"
								>
									Elegí por dónde empezar.
								</h2>
								<div style="margin-top:22px;display:flex;flex-direction:column;gap:10px">
									{#each tienda.categorias.slice(0, 6) as c, k (c.nombre)}<a
											href={enlaceCat(c.nombre)}
											style="display:flex;align-items:center;gap:14px;padding:14px 16px;border-radius:20px;background:#fff;color:#3A1F25;text-decoration:none"
											><span
												style="width:44px;height:44px;border-radius:50%;background:{BG[
													k % 4
												]};flex:none;overflow:hidden"
												>{#if c.foto}<img
														src={img(c.foto, 120, 120)}
														alt=""
														style="width:100%;height:100%;object-fit:cover"
													/>{/if}</span
											><span style="flex:1;min-width:0"
												><span style="display:block;font-size:15px;font-weight:700">{c.nombre}</span
												><span style="display:block;font-size:13px;color:#8C6A70"
													>{c.cantidad} {c.cantidad === 1 ? 'producto' : 'productos'}</span
												></span
											><span style="font-size:13px;font-weight:800;color:var(--acento, #B4475B)"
												>Ver</span
											></a
										>{/each}
								</div>
							</div>
						</section>
					{/if}
				{:else if seccion === 'prods'}
					{#if ver('prods') && best.length}
						<section style="padding:clamp(56px,8cqw,104px) clamp(14px,3.5cqw,44px) 0">
							<div
								style="display:flex;justify-content:space-between;align-items:end;gap:16px;flex-wrap:wrap"
							>
								<h2
									style="margin:0;font-size:clamp(34px,4.6cqw,58px);line-height:1;letter-spacing:-.04em;font-weight:800"
								>
									Los favoritos
								</h2>
								<a
									href="{base}/catalogo"
									style="text-decoration:none;padding:11px 18px;border-radius:999px;background:#fff;color:#3A1F25;font:700 14px 'Hanken Grotesk'"
									>Ver {tienda.productos.length === 1
										? 'el producto'
										: `los ${tienda.productos.length} productos`}</a
								>
							</div>
							<div
								style="margin-top:24px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(46%,230px),1fr));gap:14px"
							>
								{#each best as p (p.id)}<a
										href={p.href}
										class="zoom"
										style="display:block;text-decoration:none;text-align:left;font-family:inherit;color:#3A1F25"
										><span
											style="position:relative;display:block;aspect-ratio:4/5;border-radius:24px;overflow:hidden;background:{p.bg}"
											>{#if p.img}<img
													src={p.img}
													alt=""
													style="width:100%;height:100%;object-fit:cover;display:block"
												/>{/if}{#if p.tag}<span
													style="position:absolute;top:10px;left:10px;padding:5px 10px;border-radius:999px;background:#fff;font-size:11.5px;font-weight:800"
													>{p.tag}</span
												>{/if}<span
												style="position:absolute;right:10px;bottom:10px;width:40px;height:40px;border-radius:50%;background:#3A1F25;color:#fff;display:grid;place-items:center"
												><Icon icon={I.phPlusBold} size={16} /></span
											></span
										>{#if p.for.length}<span
												style="display:flex;gap:4px;margin-top:12px;flex-wrap:wrap"
												>{#each p.for as f (f)}<span
														style="padding:3px 8px;border-radius:999px;background:#F3D9DE;font-size:11px;font-weight:700"
														>{f}</span
													>{/each}</span
											>{/if}<span
											style="display:block;margin-top:8px;font-size:15.5px;font-weight:700"
											>{p.n}</span
										>{#if p.nota}<span
												style="display:flex;align-items:center;gap:8px;margin-top:4px;font-size:13px;color:#8C6A70"
												><span
													style="display:flex;align-items:center;gap:3px;color:#3A1F25;font-weight:700"
													><Icon
														icon={I.phStarFill}
														size={13}
														style="color:var(--acento, #B4475B)"
													/>{p.nota.r}</span
												>{p.nota.n}
												{p.nota.n === 1 ? 'reseña' : 'reseñas'}</span
											>{/if}<span
											style="display:flex;gap:8px;align-items:baseline;margin-top:6px;flex-wrap:wrap"
											><b style="font-size:16px">{p.price}</b
											>{#if tienda.cobros.transferencia > 0}<span
													style="font-size:12.5px;color:#2F7D55;font-weight:700"
													>{p.tr} transf.</span
												>{/if}</span
										></a
									>{/each}
							</div>
						</section>
					{/if}
				{:else if seccion === 'revs'}
					{#if ver('revs') && tienda.resenas.length}
						<section style="padding:clamp(56px,8cqw,104px) clamp(14px,3.5cqw,44px) 0">
							<div
								style="display:flex;justify-content:space-between;align-items:end;gap:16px;flex-wrap:wrap"
							>
								<h2
									style="margin:0;font-size:clamp(34px,4.6cqw,58px);line-height:1;letter-spacing:-.04em;font-weight:800"
								>
									Lo que dicen
								</h2>
								{#if tienda.promedio}<span
										style="display:flex;align-items:center;gap:8px;padding:11px 16px;border-radius:999px;background:#fff;font-size:13.5px;font-weight:700"
										><span style="display:flex;color:var(--acento, #B4475B)"
											>{#each estrellas(tienda.promedio) as e, i (i)}<Icon
													icon={e}
													size={15}
												/>{/each}</span
										>{tienda.promedio.toLocaleString('es-AR')}</span
									>{/if}
							</div>
							<div
								style="margin-top:24px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,300px),1fr));gap:14px"
							>
								{#each tienda.resenas.slice(0, 3) as r, i (i)}
									{@const pr = tienda.productos.find((p) => p.nombre === r.producto)}
									<div style="padding:12px;border-radius:28px;background:#fff">
										{#if pr?.foto}<div
												style="position:relative;aspect-ratio:3/2;border-radius:20px;overflow:hidden;background:#F3D9DE"
											>
												<img
													src={img(pr.foto, 600, 400)}
													alt=""
													style="width:100%;height:100%;object-fit:cover"
												/><span
													style="position:absolute;left:8px;top:8px;display:flex;gap:1px;padding:4px 8px;border-radius:999px;background:rgba(255,255,255,.85);color:var(--acento, #B4475B)"
													>{#each estrellas(r.estrellas) as e, j (j)}<Icon
															icon={e}
															size={12}
														/>{/each}</span
												>
											</div>{:else}<span
												style="display:flex;gap:1px;padding:6px 6px 0;color:var(--acento, #B4475B)"
												>{#each estrellas(r.estrellas) as e, j (j)}<Icon
														icon={e}
														size={14}
													/>{/each}</span
											>{/if}
										<div style="padding:14px 6px 6px">
											<p style="margin:0;font-family:'Gloock',serif;font-size:19px;line-height:1.3">
												“{r.texto}”
											</p>
											<div style="margin-top:10px;font-size:12.5px;color:#8C6A70">
												{r.autor}{#if r.producto}
													· {r.conCompra ? 'compró' : 'sobre'} {r.producto}{/if}
											</div>
											{#if r.conCompra}<span
													style="margin-top:8px;display:inline-flex;align-items:center;gap:4px;padding:4px 10px;border-radius:999px;background:#F3D9DE;color:var(--acento, #B4475B);font-size:11.5px;font-weight:800"
													><Icon icon={I.phSealCheckFill} size={13} />Compra verificada</span
												>{/if}
											{#if r.respuesta}<div
													style="margin-top:10px;padding:12px 14px;border-radius:18px;background:#FBF4F1;font-size:13px;line-height:1.5;color:#6E4F55"
												>
													<b
														style="display:block;font-size:11px;letter-spacing:.14em;text-transform:uppercase;color:var(--acento, #B4475B)"
														>Respuesta de {tienda.nombre}</b
													>{r.respuesta}
												</div>{/if}
										</div>
									</div>
								{/each}
							</div>
						</section>
					{/if}
				{:else if seccion === 'ig'}
					{#if ver('ig') && tienda.fotos.length >= 3}
						<section style="padding:clamp(56px,8cqw,104px) clamp(14px,3.5cqw,44px)">
							<h2
								style="margin:0;font-size:clamp(34px,4.6cqw,58px);line-height:1;letter-spacing:-.04em;font-weight:800"
							>
								Así se usa
							</h2>
							<div
								style="margin-top:20px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(30%,160px),1fr));gap:8px"
							>
								{#each tienda.fotos as g (g)}<div
										style="aspect-ratio:1;border-radius:22px;overflow:hidden;background:#F3D9DE"
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
				<div style="height:clamp(56px,8cqw,104px)"></div>
			{/if}
		{/if}

		<!-- ============ CATÁLOGO ============ -->
		{#if pagina === 'catalogo' && !tienda.productos.length}
			<section style="padding:clamp(40px,7cqw,96px) clamp(16px,4cqw,48px)">
				<VacioTienda {tienda} {tema} motivo="sin-productos" />
			</section>
		{:else if pagina === 'catalogo'}
			<section style="padding:24px clamp(14px,3.5cqw,44px) clamp(48px,7cqw,90px)">
				<h1
					style="margin:0;font-family:'Gloock',serif;font-weight:400;font-size:clamp(46px,6.6cqw,92px);line-height:.95;overflow-wrap:anywhere"
				>
					{categoria ?? 'Todo'}
				</h1>
				<div style="margin-top:18px;display:flex;gap:8px;overflow-x:auto;scrollbar-width:none">
					{#each [null, ...tienda.categorias.map((c) => c.nombre)] as c (c ?? '')}{@const on =
							c === categoria}<a
							href={enlaceCat(c)}
							style="flex:none;display:flex;align-items:center;gap:7px;padding:10px 15px;border-radius:999px;background:{on
								? '#3A1F25'
								: '#fff'};color:{on
								? '#fff'
								: '#3A1F25'};font-size:13.5px;font-weight:700;white-space:nowrap;text-decoration:none"
							>{#if c === null}<Icon icon={I.phSparkle} size={15} />{/if}{c ?? 'Todo'}</a
						>{/each}
				</div>
				<div style="margin-top:22px;display:flex;gap:clamp(18px,3cqw,40px);align-items:flex-start">
					{#if wide}<aside
							style="width:230px;flex:none;position:sticky;top:90px;display:flex;flex-direction:column;gap:24px;font-size:14px"
						>
							<div>
								<div
									style="font-size:12px;font-weight:800;letter-spacing:.16em;text-transform:uppercase;color:var(--acento, #B4475B);margin-bottom:12px"
								>
									Ordenar
								</div>
								<div style="display:flex;flex-wrap:wrap;gap:6px">
									{#each ORDENES as [k, t] (k)}{@const on = orden === k}<a
											href={enlaceCat(categoria, k)}
											style="padding:8px 12px;border-radius:999px;border:1.5px solid {on
												? 'var(--acento, #B4475B)'
												: '#EAD3CF'};background:{on
												? '#F3D9DE'
												: 'transparent'};font-size:13px;font-weight:600;color:#3A1F25;text-decoration:none"
											>{t}</a
										>{/each}
								</div>
							</div>
						</aside>{/if}
					<div style="flex:1;min-width:0">
						<div
							style="display:flex;justify-content:space-between;align-items:center;gap:10px;flex-wrap:wrap;padding-bottom:14px;border-bottom:1px solid #EAD3CF;font-size:13.5px"
						>
							<span style="display:flex;gap:6px;flex-wrap:wrap;align-items:center"
								>{#if categoria}<a
										href={enlaceCat(null)}
										aria-label="Quitar el filtro {categoria}"
										style="display:flex;align-items:center;gap:6px;padding:7px 12px;border-radius:999px;background:#F3D9DE;font-weight:600;color:#3A1F25;text-decoration:none"
										>{categoria}<Icon icon={I.phXBold} size={11} /></a
									>{/if}<span style="color:#8C6A70"
									>{listado.length} {listado.length === 1 ? 'producto' : 'productos'}</span
								></span
							>{#if nar}<label
									style="display:flex;align-items:center;gap:6px;font-weight:700;position:relative"
									><span class="oculto">Ordenar</span><select
										value={orden}
										onchange={(e) =>
											goto(enlaceCat(categoria, e.currentTarget.value), { replace: true })}
										style="appearance:none;border:0;background:transparent;padding:0 18px 0 0;font:700 13.5px 'Hanken Grotesk';color:#3A1F25;cursor:pointer"
										>{#each ORDENES as [k, t] (k)}<option value={k}>{t}</option>{/each}</select
									><Icon
										icon={I.phCaretDownBold}
										size={12}
										style="position:absolute;right:0;pointer-events:none"
									/></label
								>{/if}
						</div>
						{#if catalog.length}
							<div
								style="margin-top:20px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(46%,210px),1fr));gap:14px"
							>
								{#each catalog as p (p.id)}<a
										href={p.href}
										style="display:block;text-decoration:none;text-align:left;font-family:inherit;color:#3A1F25"
										><span
											style="position:relative;display:block;aspect-ratio:4/5;border-radius:22px;overflow:hidden;background:{p.bg}"
											>{#if p.img}<img
													src={p.img}
													alt=""
													style="width:100%;height:100%;object-fit:cover;display:block"
												/>{/if}{#if p.tag}<span
													style="position:absolute;top:10px;left:10px;padding:5px 10px;border-radius:999px;background:#fff;font-size:11.5px;font-weight:800"
													>{p.tag}</span
												>{/if}</span
										><span style="display:block;margin-top:10px;font-size:15px;font-weight:700"
											>{p.n}</span
										>{#if p.nota}<span
												style="display:flex;align-items:center;gap:8px;margin-top:3px;font-size:13px;color:#8C6A70"
												><span
													style="display:flex;align-items:center;gap:3px;color:#3A1F25;font-weight:700"
													><Icon
														icon={I.phStarFill}
														size={13}
														style="color:var(--acento, #B4475B)"
													/>{p.nota.r}</span
												>{p.nota.n}
												{p.nota.n === 1 ? 'reseña' : 'reseñas'}</span
											>{/if}<span
											style="display:flex;gap:8px;align-items:baseline;margin-top:5px;flex-wrap:wrap"
											><b style="font-size:15.5px">{p.price}</b>{#if p.old}<s
													style="font-size:13px;color:#A88A8F">{p.old}</s
												>{/if}</span
										></a
									>{/each}
							</div>
							{#if listado.length > mostrar}
								<div style="margin-top:36px;display:flex;justify-content:center">
									<button
										type="button"
										onclick={() => (mostrar += 24)}
										style="border:0;cursor:pointer;padding:11px 18px;border-radius:999px;background:#fff;color:#3A1F25;font:700 14px 'Hanken Grotesk'"
										>Ver más</button
									>
								</div>
							{/if}
						{:else}
							<p style="margin:40px 0;font-size:15px;color:#8C6A70">
								Todavía no hay productos en esta categoría.
							</p>
						{/if}
					</div>
				</div>
			</section>
		{/if}

		<!-- ============ PRODUCTO ============ -->
		{#if pagina === 'producto' && producto && sel}
			<section
				style="padding:24px clamp(14px,3.5cqw,44px) 0;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,400px),1fr));gap:clamp(20px,4cqw,60px);align-items:start"
			>
				<div
					style="display:grid;grid-template-columns:{sel.fotos.length > 1
						? '72px minmax(0,1fr)'
						: 'minmax(0,1fr)'};gap:10px"
				>
					{#if sel.fotos.length > 1}<div style="display:flex;flex-direction:column;gap:8px">
							{#each sel.fotos as f, i (f)}<button
									type="button"
									aria-label="Ver la foto {i + 1}"
									aria-pressed={i === sel.fotoIdx}
									onclick={() => sel.verFoto(i)}
									style="aspect-ratio:1;padding:0;border:0;border-radius:14px;overflow:hidden;background:#F3D9DE;cursor:pointer;box-shadow:0 0 0 {i ===
									sel.fotoIdx
										? '2.5px'
										: '0px'} var(--acento, #B4475B)"
									><img
										src={img(f, 160, 160)}
										alt=""
										style="display:block;width:100%;height:100%;object-fit:cover"
									/></button
								>{/each}
						</div>{/if}
					<div style="aspect-ratio:4/5;border-radius:30px;overflow:hidden;background:#F3D9DE">
						{#if sel.foto}<img
								src={img(sel.foto, 1000)}
								alt={producto.nombre}
								style="width:100%;height:100%;object-fit:cover"
							/>{/if}
					</div>
				</div>
				<div style="position:sticky;top:90px;min-width:0">
					<div style="display:flex;gap:6px;flex-wrap:wrap">
						{#if producto.categoria}<a
								href={enlaceCat(producto.categoria)}
								style="padding:5px 10px;border-radius:999px;background:#F3D9DE;font-size:12px;font-weight:800;color:#3A1F25;text-decoration:none"
								>{producto.categoria}</a
							>{/if}{#if tachado}<span
								style="padding:5px 10px;border-radius:999px;background:#F3D9DE;font-size:12px;font-weight:800"
								>−{Math.round((1 - sel.precio / tachado) * 100)}%</span
							>{:else if producto.etiqueta}<span
								style="padding:5px 10px;border-radius:999px;background:#F3D9DE;font-size:12px;font-weight:800"
								>{producto.etiqueta}</span
							>{/if}{#if (rank.get(producto.id) ?? 99) < 3}<span
								style="padding:5px 10px;border-radius:999px;background:#3A1F25;color:#fff;font-size:12px;font-weight:800"
								>Más vendido</span
							>{/if}
					</div>
					<h1
						style="margin:14px 0 0;font-family:'Gloock',serif;font-weight:400;font-size:clamp(40px,4.8cqw,62px);line-height:.98;overflow-wrap:anywhere"
					>
						{producto.nombre}
					</h1>
					{#if notaFicha}<div
							style="margin-top:10px;display:flex;align-items:center;gap:8px;font-size:14px"
						>
							<span style="display:flex;color:var(--acento, #B4475B)"
								>{#each estrellas(Number(notaFicha.r.replace(',', '.'))) as e, i (i)}<Icon
										icon={e}
										size={15}
									/>{/each}</span
							><b>{notaFicha.r}</b><span style="color:#8C6A70"
								>{notaFicha.n} {notaFicha.n === 1 ? 'reseña' : 'reseñas'}</span
							>
						</div>{/if}
					<div style="margin-top:18px;display:flex;align-items:baseline;gap:10px">
						<b style="font-size:30px">{sel.desde ? 'Desde ' : ''}{fmt(sel.precio)}</b
						>{#if tachado}<s style="color:#8C6A70">{fmt(tachado)}</s>{/if}{#if sel.variante}<span
								style="color:#8C6A70">{sel.opciones.map((o) => o.elegido).join(' · ')}</span
							>{/if}
					</div>
					{#if tienda.cobros.cuotas > 1 || tienda.cobros.transferencia > 0}<div
							style="margin-top:4px;font-size:14px"
						>
							{#if tienda.cobros.cuotas > 1}{tienda.cobros.cuotas} cuotas de
								<b>{fmt(sel.precio / tienda.cobros.cuotas)}</b
								>{/if}{#if tienda.cobros.transferencia > 0}{#if tienda.cobros.cuotas > 1}
									·
								{/if}<b style="color:#2F7D55"
									>{fmt(sel.precio * (1 - tienda.cobros.transferencia / 100))} por transferencia</b
								>{/if}
						</div>{/if}
					<!-- Venta por mayor (§49): con el estilo del renglón de las cuotas. -->
					{#if porMayorFicha}<div style="margin-top:4px;font-size:14px">{porMayorFicha}</div>{/if}
					{#each sel.opciones as o (o.nombre)}{@const swatch = o.valores.every(
							(x) => x.color || x.foto
						)}
						<div style="margin-top:20px">
							<div
								style="font-size:14px;color:{sel.aviso && !o.elegido
									? 'var(--acento, #B4475B)'
									: '#3A1F25'}"
							>
								{o.nombre}{#if o.elegido}: <b>{o.elegido}</b>{/if}
							</div>
							{#if swatch}<div style="margin-top:10px;display:flex;gap:10px;flex-wrap:wrap">
									{#each o.valores as x (x.valor)}<button
											type="button"
											aria-label="{o.nombre} {x.valor}"
											aria-pressed={x.elegido}
											disabled={!x.disponible && !x.elegido}
											onclick={() => sel.elegir(o.nombre, x.valor)}
											style="width:40px;height:40px;padding:0;border:0;border-radius:50%;overflow:hidden;background:#F3D9DE;cursor:pointer;opacity:{x.disponible
												? 1
												: 0.4};box-shadow:0 0 0 {x.elegido
												? '2.5px'
												: '0px'} #FBF4F1,0 0 0 {x.elegido ? '5px' : '0px'} var(--acento, #B4475B)"
											>{#if x.color}<span
													style="display:block;width:100%;height:100%;background:{x.color}"
												></span>{:else}<img
													src={img(x.foto, 80, 80)}
													alt=""
													style="display:block;width:100%;height:100%;object-fit:cover"
												/>{/if}</button
										>{/each}
								</div>{:else}<div
									style="margin-top:10px;display:flex;flex-direction:column;gap:8px"
								>
									{#each o.valores as x (x.valor)}{@const pv =
											sel.opciones.length === 1
												? sel.variantes.find((v) => v.valores[o.nombre] === x.valor)
												: null}<button
											type="button"
											aria-pressed={x.elegido}
											disabled={!x.disponible && !x.elegido}
											onclick={() => sel.elegir(o.nombre, x.valor)}
											style="display:flex;align-items:center;gap:12px;padding:15px;border-radius:20px;border:1.5px solid {x.elegido
												? 'var(--acento, #B4475B)'
												: '#EAD3CF'};background:{x.elegido
												? '#FFF6F7'
												: '#fff'};font:400 14px 'Hanken Grotesk';color:{x.disponible
												? '#3A1F25'
												: '#8C6A70'};text-align:left;cursor:{x.disponible ? 'pointer' : 'default'}"
											><span
												style="width:20px;height:20px;border-radius:50%;border:2px solid {x.elegido
													? 'var(--acento, #B4475B)'
													: '#D9BEC3'};display:grid;place-items:center;flex:none"
												>{#if x.elegido}<span
														style="width:10px;height:10px;border-radius:50%;background:var(--acento, #B4475B)"
													></span>{/if}</span
											><span
												style="flex:1;font-size:15px;font-weight:700;text-decoration:{x.disponible
													? 'none'
													: 'line-through'}">{x.valor}</span
											>{#if pv}<b style="text-decoration:{x.disponible ? 'none' : 'line-through'}"
													>{fmt(pv.precio ?? producto.precio)}</b
												>{/if}</button
										>{/each}
								</div>{/if}
						</div>{/each}
					{#if sel.stock > 0 && sel.stock <= 3 && (!sel.conOpciones || sel.variante)}<div
							style="margin-top:12px;font-size:13px;font-weight:700;color:var(--acento, #B4475B)"
						>
							{sel.stock === 1 ? 'Queda 1' : `Quedan ${sel.stock}`}{#if sel.variante}
								en {sel.texto.toLowerCase()}{/if}
						</div>{/if}
					<button
						type="button"
						onclick={agregar}
						disabled={sel.stock === 0 || tienda.vitrina}
						style="margin-top:20px;width:100%;padding:17px;border:0;border-radius:999px;background:#3A1F25;color:#fff;font:800 15px 'Hanken Grotesk';cursor:pointer"
						>{sel.boton(`Agregar · ${fmt(sel.precio)}`)}</button
					>
					{#if whatsapp}<a
							href={wa(
								`¡Hola! Quiero comprar ${producto.nombre}${sel.variante ? ` (${sel.texto})` : ''} (${fmt(sel.precio)}).`
							)}
							target="_blank"
							rel="noopener"
							style="margin-top:8px;display:flex;justify-content:center;align-items:center;gap:8px;padding:14px;border-radius:999px;border:1.5px solid #E2C3C8;font-size:14.5px;font-weight:700;color:#3A1F25;text-decoration:none"
							><Icon icon={I.logosWhatsappIcon} size={17} />Consultanos por WhatsApp</a
						>{/if}
					<LanzamientoFicha {tienda} {producto} {sel} {tema} />
					<CompartirProducto {tienda} {producto} {tema} />
					{#if envia(tienda.envios)}<div
							style="margin-top:18px;padding:16px;border-radius:22px;background:#fff"
						>
							<label for="cp-ficha" style="font-size:14px;font-weight:700">Envío</label>
							<div style="margin-top:10px;display:flex;gap:6px">
								<input
									id="cp-ficha"
									bind:value={cp}
									inputmode="numeric"
									maxlength="4"
									placeholder="Código postal"
									style="flex:1;min-width:0;padding:11px 12px;border:0;border-radius:14px;background:#FBF4F1;font:400 14px 'Hanken Grotesk';color:#3A1F25"
									onkeydown={(e) => e.key === 'Enter' && calcular()}
								/><button
									type="button"
									onclick={calcular}
									style="padding:11px 16px;border:0;border-radius:14px;background:#3A1F25;color:#fff;font:700 13.5px 'Hanken Grotesk';cursor:pointer"
									>Calcular</button
								>
							</div>
							{#if cotizado || tienda.envios.gratisDesde}<div
									style="margin-top:10px;display:flex;flex-direction:column;gap:6px;font-size:13.5px"
								>
									{#each cotizado ?? [] as o (o.id)}<span
											style="display:flex;justify-content:space-between;gap:10px"
											><span>{o.nombre} · {o.detalle}</span><b
												style="white-space:nowrap;color:{o.precio ? '#3A1F25' : '#2F7D55'}"
												>{o.precio ? fmt(o.precio) : 'Gratis'}</b
											></span
										>{/each}{#if tienda.envios.gratisDesde}<span
											style="display:flex;justify-content:space-between;gap:10px"
											><span>Gratis desde {fmt(tienda.envios.gratisDesde)}</span><b
												style="color:#2F7D55;white-space:nowrap"
												>{sel.precio >= tienda.envios.gratisDesde
													? 'Te llega gratis'
													: `Te faltan ${fmt(tienda.envios.gratisDesde - sel.precio)}`}</b
											></span
										>{/if}
								</div>{/if}
						</div>{/if}
					<div style="margin-top:12px;padding:4px 16px;border-radius:22px;background:#fff">
						{#each acc as a, k (a.t)}<div
								style="padding:14px 0;border-top:{k ? '1px solid #EAD3CF' : '0'}"
							>
								<button
									type="button"
									onclick={() => (abierto = a.open ? '' : a.t)}
									aria-expanded={a.open}
									style="width:100%;display:flex;justify-content:space-between;align-items:center;padding:0;border:0;background:transparent;font:800 12px 'Hanken Grotesk';letter-spacing:.14em;text-transform:uppercase;color:var(--acento, #B4475B);cursor:pointer;text-align:left"
									>{a.t}<Icon icon={a.i} size={12} /></button
								>{#if a.open}<p
										style="margin:10px 0 0;font-size:14px;line-height:1.6;color:#6E4F55"
									>
										{a.d}
									</p>{/if}
							</div>{/each}
					</div>
					<div style="margin-top:12px;padding:18px 16px;border-radius:22px;background:#fff">
						<PreguntasProducto
							slug={tienda.slug}
							productoId={producto.id}
							preguntas={tienda.preguntas}
							tema={{ ...tema, tarjeta: 'transparent' }}
							titulo="Preguntas"
						/>
					</div>
				</div>
			</section>
			{#if related.length}
				<section style="padding:clamp(52px,7cqw,96px) clamp(14px,3.5cqw,44px)">
					<h2
						style="margin:0;font-size:clamp(30px,4cqw,48px);line-height:1;letter-spacing:-.04em;font-weight:800"
					>
						También te puede gustar
					</h2>
					<div
						style="margin-top:22px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(46%,200px),1fr));gap:12px"
					>
						{#each related as p (p.id)}<a
								href={p.href}
								style="display:block;text-decoration:none;text-align:left;font-family:inherit;color:#3A1F25"
								><span
									style="display:block;aspect-ratio:1;border-radius:22px;overflow:hidden;background:{p.bg}"
									>{#if p.img}<img
											src={p.img}
											alt=""
											style="width:100%;height:100%;object-fit:cover"
										/>{/if}</span
								><span style="display:block;margin-top:10px;font-size:14.5px;font-weight:700"
									>{p.n}</span
								><b style="font-size:14.5px">{p.price}</b></a
							>{/each}
					</div>
				</section>
			{:else}
				<div style="height:clamp(52px,7cqw,96px)"></div>
			{/if}
		{/if}

		<!-- ============ CARRITO ============ -->
		{#if pagina === 'carrito'}
			<section style="padding:24px clamp(14px,3.5cqw,44px) clamp(48px,7cqw,90px)">
				<h1
					style="margin:0;font-family:'Gloock',serif;font-weight:400;font-size:clamp(44px,6cqw,80px);line-height:1"
				>
					Tu bolsa
				</h1>
				{#if !lineas.length}
					<p style="margin:20px 0 0;font-size:15px;color:#8C6A70">Todavía no sumaste nada.</p>
					<a
						href="{base}/catalogo"
						style="display:inline-block;margin-top:16px;padding:16px 26px;border-radius:999px;background:#3A1F25;color:#fff;font:800 15px 'Hanken Grotesk';text-decoration:none"
						>Ver los productos</a
					>
				{:else}
					<div
						style="margin-top:22px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,340px),1fr));gap:clamp(16px,3cqw,36px);align-items:start"
					>
						<div style="display:flex;flex-direction:column;gap:10px">
							{#each lineas as c (c.clave)}<div
									style="display:flex;gap:14px;padding:12px;border-radius:24px;background:#fff"
								>
									{#if c.producto.foto}<img
											src={img(c.producto.foto, 200, 240)}
											alt=""
											style="width:86px;height:100px;border-radius:18px;object-fit:cover;flex:none;background:#F3D9DE"
										/>{/if}
									<div style="flex:1;min-width:0;display:flex;flex-direction:column">
										<div style="display:flex;justify-content:space-between;gap:10px">
											<a
												href="{base}/producto/{c.producto.slug ?? c.id}"
												style="font-size:15px;font-weight:700;color:#3A1F25;text-decoration:none"
												>{c.producto.nombre}</a
											><b style="white-space:nowrap">{fmt(c.subtotal)}</b>
										</div>
										<span style="font-size:13px;color:#8C6A70"
											>{c.variante ? `${c.variante} · ` : ''}{fmt(c.producto.precio)} c/u</span
										>
										<span style="flex:1;min-height:10px"></span>
										<div style="display:flex;justify-content:space-between;align-items:center">
											<span
												style="display:flex;align-items:center;gap:4px;padding:2px 4px;border-radius:999px;background:#FBF4F1"
												><button
													type="button"
													aria-label="Uno menos"
													onclick={() => carrito.fijar(c.clave, c.cantidad - 1)}
													style="width:28px;height:28px;display:grid;place-items:center;border:0;background:transparent;cursor:pointer;color:#3A1F25"
													><Icon icon={I.phMinusBold} size={12} /></button
												><b style="font-size:14px;min-width:16px;text-align:center">{c.cantidad}</b
												><button
													type="button"
													aria-label="Uno más"
													disabled={c.cantidad >= c.producto.stock}
													onclick={() => carrito.fijar(c.clave, c.cantidad + 1)}
													style="width:28px;height:28px;display:grid;place-items:center;border:0;background:transparent;cursor:pointer;color:#3A1F25"
													><Icon icon={I.phPlusBold} size={12} /></button
												></span
											><button
												type="button"
												aria-label="Quitar {c.producto.nombre}"
												onclick={() => carrito.fijar(c.clave, 0)}
												style="border:0;background:transparent;padding:4px;cursor:pointer;color:#A88A8F;display:grid"
												><Icon icon={I.phTrash} size={18} /></button
											>
										</div>
									</div>
								</div>{/each}
							{#if upsell.length}
								<div style="padding:16px;border-radius:24px;background:#F3D9DE">
									<div style="font-size:15px;font-weight:800">Te puede gustar</div>
									<div style="font-size:13px;color:#6E4F55;margin-top:2px">
										Los más elegidos de la tienda
									</div>
									<div
										style="margin-top:12px;display:grid;grid-template-columns:repeat(3,1fr);gap:8px"
									>
										{#each upsell as p (p.id)}<div
												style="padding:8px;border-radius:18px;background:#fff;text-align:center;min-width:0"
											>
												<a
													href="{base}/producto/{p.slug ?? p.id}"
													style="display:block;aspect-ratio:1;border-radius:12px;overflow:hidden;background:#FBF4F1"
													>{#if p.foto}<img
															src={img(p.foto, 160, 160)}
															alt=""
															style="width:100%;height:100%;object-fit:cover"
														/>{/if}</a
												>
												<div
													style="margin-top:6px;font-size:12.5px;font-weight:700;line-height:1.2;overflow-wrap:anywhere"
												>
													{p.nombre}
												</div>
												<div
													style="margin-top:6px;display:flex;justify-content:center;align-items:center;gap:6px;flex-wrap:wrap"
												>
													<b style="font-size:12.5px;white-space:nowrap">{fmt(p.precio)}</b><button
														type="button"
														aria-label="Sumar {p.nombre}"
														onclick={() => sumarRapido(p)}
														style="width:28px;height:28px;border-radius:50%;border:0;background:#3A1F25;color:#fff;display:grid;place-items:center;cursor:pointer;flex:none"
														><Icon icon={I.phPlusBold} size={12} /></button
													>
												</div>
											</div>{/each}
									</div>
								</div>
							{/if}
						</div>
						<div
							style="position:sticky;top:90px;padding:22px;border-radius:28px;background:#fff;box-shadow:0 24px 50px -30px rgba(58,31,37,.4);display:flex;flex-direction:column;gap:11px;font-size:14px"
						>
							<div style="font-family:'Gloock',serif;font-size:28px">Resumen</div>
							{#if tienda.envios.gratisDesde}<div>
									<div style="display:flex;justify-content:space-between;font-size:13px">
										<span
											>{#if falta}Te faltan <b>{fmt(falta)}</b> para el envío gratis{:else}<b
													>Tenés envío gratis</b
												>{/if}</span
										>
									</div>
									<div
										style="margin-top:8px;height:6px;border-radius:3px;background:#F3D9DE;overflow:hidden"
									>
										<div
											style="width:{Math.min(
												100,
												(subtotal / tienda.envios.gratisDesde) * 100
											)}%;height:100%;background:var(--acento, #B4475B)"
										></div>
									</div>
								</div>{/if}
							{#if avisoMayor && porMayor}<div>
									<div style="display:flex;justify-content:space-between;font-size:13px">
										<span>{avisoMayor.pre}<b>{avisoMayor.b}</b>{avisoMayor.post}</span>
									</div>
									<div
										style="margin-top:8px;height:6px;border-radius:3px;background:#F3D9DE;overflow:hidden"
									>
										<div
											style="width:{porMayor.progreso}%;height:100%;background:var(--acento, #B4475B)"
										></div>
									</div>
								</div>{/if}
							<span style="display:flex;justify-content:space-between"
								><span>Subtotal</span><span>{fmt(subtotal)}</span></span
							>
							<span style="display:flex;justify-content:space-between;gap:10px;color:#8C6A70"
								><span>Envío</span><span style="text-align:right"
									>Se calcula en el próximo paso</span
								></span
							>
							<div style="height:1px;background:#EAD3CF"></div>
							<span style="display:flex;justify-content:space-between;align-items:baseline"
								><b>Total</b><b style="font-family:'Gloock',serif;font-size:30px">{fmt(subtotal)}</b
								></span
							>
							{#if tienda.cobros.cuotas > 1}<span style="color:#8C6A70"
									>{tienda.cobros.cuotas} cuotas sin interés de {fmt(
										subtotal / tienda.cobros.cuotas
									)}</span
								>{/if}
							{#if tienda.cobros.transferencia > 0}<span style="color:#2F7D55;font-weight:700"
									>{fmt(subtotal * (1 - tienda.cobros.transferencia / 100))} pagando por transferencia</span
								>{/if}
							<a
								href="{base}/checkout"
								style="margin-top:4px;padding:16px;border:0;border-radius:999px;background:#3A1F25;color:#fff;font:800 15px 'Hanken Grotesk';text-align:center;text-decoration:none"
								>Ir a pagar</a
							>
						</div>
					</div>
				{/if}
			</section>
		{/if}

		<!-- ============ CHECKOUT ============ -->
		{#if pagina === 'checkout'}
			{#if !lineas.length}
				<section style="padding:24px clamp(14px,3.5cqw,44px) clamp(48px,7cqw,90px)">
					<p style="font-size:15px;color:#8C6A70">Tu bolsa está vacía.</p>
					<a href="{base}/catalogo" style="color:#3A1F25;font-weight:700">Ver los productos</a>
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
					style="padding:24px clamp(14px,3.5cqw,44px) clamp(48px,7cqw,90px);display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,360px),1fr));gap:clamp(16px,3cqw,40px);align-items:start"
				>
					<input type="hidden" name="carrito" value={JSON.stringify(carrito.lineas)} />
					<input type="hidden" name="envio" value={envio ?? ''} />
					<input type="hidden" name="pago" value={pago} />
					<div style="display:flex;flex-direction:column;gap:22px;min-width:0">
						<div style="display:flex;gap:6px">
							{#each pasos as s, i (i)}<span
									style="flex:1;height:5px;border-radius:3px;background:{s}"
								></span>{/each}
						</div>
						<h1
							style="margin:0;font-family:'Gloock',serif;font-weight:400;font-size:clamp(38px,5cqw,62px);line-height:1"
						>
							Ya casi
						</h1>
						<div>
							<div style="font-size:15px;font-weight:800;margin-bottom:10px">Tus datos</div>
							<div
								style="display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));gap:8px"
							>
								{#each [['email', 'Email', 'email', 'email'], ['nombre', 'Nombre y apellido', 'text', 'name'], ['telefono', 'WhatsApp', 'tel', 'tel'], ['dni', 'DNI', 'text', 'off'], ...camposFactura(tienda)] as [n, l, tipo, ac] (n)}
									<label
										style="display:block;padding:11px 14px;border-radius:16px;background:#fff;box-shadow:inset 0 0 0 1.5px #EAD3CF"
										><span style="display:block;font-size:11.5px;color:#A88A8F">{l}</span><input
											name={n}
											type={tipo}
											autocomplete={ac as 'email'}
											required={!OPCIONALES.has(n)}
											maxlength="120"
											style="width:100%;margin-top:2px;padding:0;border:0;background:transparent;font:400 14.5px 'Hanken Grotesk';color:#3A1F25;outline:none"
										/></label
									>
								{/each}
							</div>
						</div>
						<div>
							<div style="font-size:15px;font-weight:800;margin-bottom:10px">Entrega</div>
							<div style="display:grid;grid-template-columns:minmax(0,140px) minmax(0,1fr);gap:8px">
								<label
									style="display:block;padding:11px 14px;border-radius:16px;background:#fff;box-shadow:inset 0 0 0 1.5px #EAD3CF"
									><span style="display:block;font-size:11.5px;color:#A88A8F">Código postal</span
									><input
										name="cp"
										bind:value={cp}
										inputmode="numeric"
										maxlength="4"
										style="width:100%;margin-top:2px;padding:0;border:0;background:transparent;font:400 14.5px 'Hanken Grotesk';color:#3A1F25;outline:none"
									/></label
								>
								<label
									style="display:block;padding:11px 14px;border-radius:16px;background:#fff;box-shadow:inset 0 0 0 1.5px #EAD3CF"
									><span style="display:block;font-size:11.5px;color:#A88A8F"
										>Dirección (calle, número, piso)</span
									><input
										name="direccion"
										maxlength="160"
										autocomplete="street-address"
										required={envio !== 'retiro'}
										style="width:100%;margin-top:2px;padding:0;border:0;background:transparent;font:400 14.5px 'Hanken Grotesk';color:#3A1F25;outline:none"
									/></label
								>
							</div>
							<div
								style="margin-top:8px;display:flex;flex-direction:column;gap:8px"
								role="radiogroup"
								aria-label="Forma de entrega"
							>
								{#each opciones as s (s.id)}{@const on = envio === s.id}{@const c = ch(on)}<button
										type="button"
										role="radio"
										aria-checked={on}
										onclick={() => (envio = s.id)}
										style="display:flex;align-items:center;gap:12px;padding:15px;border-radius:18px;border:1.5px solid {c.bd};background:{c.bg};text-align:left;cursor:pointer;font-family:inherit;color:#3A1F25"
										><span
											style="width:20px;height:20px;border-radius:50%;border:2px solid {c.rc};display:grid;place-items:center;flex:none"
											>{#if on}<span
													style="width:10px;height:10px;border-radius:50%;background:var(--acento, #B4475B)"
												></span>{/if}</span
										><span style="flex:1;min-width:0"
											><span style="display:block;font-size:14.5px;font-weight:700">{s.nombre}</span
											><span style="display:block;font-size:12.5px;color:#8C6A70">{s.detalle}</span
											></span
										><b style="color:{s.precio ? '#3A1F25' : '#2F7D55'};white-space:nowrap"
											>{s.precio ? fmt(s.precio) : 'Gratis'}</b
										></button
									>{:else}<span style="font-size:13.5px;color:#8C6A70"
										>Poné tu código postal para ver cómo te llega.</span
									>{/each}
							</div>
							<EnvioCorreo {tienda} {cp} {envio} {opciones} />
						</div>
						<div>
							<div style="font-size:15px;font-weight:800;margin-bottom:10px">Pago</div>
							<div
								style="display:flex;flex-direction:column;gap:8px"
								role="radiogroup"
								aria-label="Forma de pago"
							>
								{#each pagos as s (s.id)}{@const on = pago === s.id}{@const c = ch(on)}<button
										type="button"
										role="radio"
										aria-checked={on}
										onclick={() => (pago = s.id)}
										style="display:flex;align-items:center;gap:12px;padding:15px;border-radius:18px;border:1.5px solid {c.bd};background:{c.bg};text-align:left;cursor:pointer;font-family:inherit;color:#3A1F25"
										><span
											style="width:20px;height:20px;border-radius:50%;border:2px solid {c.rc};display:grid;place-items:center;flex:none"
											>{#if on}<span
													style="width:10px;height:10px;border-radius:50%;background:var(--acento, #B4475B)"
												></span>{/if}</span
										><Icon icon={s.i} size={22} style="color:{s.c}" /><span
											style="flex:1;min-width:0"
											><span style="display:block;font-size:14.5px;font-weight:700">{s.t}</span
											><span style="display:block;font-size:12.5px;color:#8C6A70">{s.d}</span></span
										></button
									>{/each}
							</div>
						</div>
						{#if error}<div
								role="alert"
								style="padding:12px 14px;border-radius:16px;background:#FFF6F7;box-shadow:inset 0 0 0 1.5px var(--acento, #B4475B);color:var(--acento, #B4475B);font-size:14px;font-weight:600"
							>
								{error}
							</div>{/if}
						<button
							disabled={!envio || !pagos.length || enviando || tienda.vitrina}
							style="padding:17px;border:0;border-radius:999px;background:var(--acento, #B4475B);color:#fff;font:800 15.5px 'Hanken Grotesk';cursor:pointer;display:flex;justify-content:center;align-items:center;gap:8px"
							><Icon icon={I.phLockSimpleFill} size={16} />{tienda.vitrina
								? 'Es una tienda de ejemplo'
								: !pagos.length
									? 'Esta tienda todavía no tiene medios de pago'
									: envio
										? `Pagar ${fmt(total)}`
										: 'Elegí cómo te llega'}</button
						>
					</div>
					<div
						style="position:sticky;top:90px;padding:22px;border-radius:28px;background:#3A1F25;color:#FBF4F1;min-width:0"
					>
						<div style="font-family:'Gloock',serif;font-size:26px">Tu pedido</div>
						<div style="margin-top:14px;display:flex;flex-direction:column;gap:12px">
							{#each lineas as c (c.clave)}<div style="display:flex;gap:12px;align-items:center">
									<span style="position:relative;flex:none"
										>{#if c.producto.foto}<img
												src={img(c.producto.foto, 100, 116)}
												alt=""
												style="width:50px;height:58px;border-radius:14px;object-fit:cover;background:#5A3640;display:block"
											/>{:else}<span
												style="display:block;width:50px;height:58px;border-radius:14px;background:#5A3640"
											></span>{/if}<span
											style="position:absolute;top:-6px;right:-6px;width:20px;height:20px;border-radius:50%;background:#F3D9DE;color:#3A1F25;font-size:11px;font-weight:800;display:grid;place-items:center"
											>{c.cantidad}</span
										></span
									><span style="flex:1;font-size:14px;min-width:0"
										>{c.producto.nombre}{#if c.variante}<span
												style="display:block;font-size:12.5px;color:#F3D9DE">{c.variante}</span
											>{/if}</span
									><b style="font-size:14px;white-space:nowrap">{fmt(c.subtotal)}</b>
								</div>{/each}
						</div>
						<div
							style="margin-top:16px;padding-top:14px;border-top:1px solid rgba(251,244,241,.14);display:flex;flex-direction:column;gap:8px;font-size:14px"
						>
							<span style="display:flex;justify-content:space-between"
								><span>Subtotal</span><span>{fmt(subtotal)}</span></span
							><span style="display:flex;justify-content:space-between"
								><span>Envío</span><span
									>{elegidaEnvio ? (cuenta.envio ? fmt(cuenta.envio) : 'Gratis') : '—'}</span
								></span
							>{#if cuenta.descCupon}<span
									style="display:flex;justify-content:space-between;color:#9FE0B8;font-weight:700"
									><span>Cupón {cupon.aplicado}</span><span>− {fmt(cuenta.descCupon)}</span></span
								>{/if}{#if descuento}<span
									style="display:flex;justify-content:space-between;color:#9FE0B8;font-weight:700"
									><span>Transferencia −{tienda.cobros.transferencia}%</span><span
										>− {fmt(descuento)}</span
									></span
								>{/if}
							<div style="margin:6px 0 2px">
								<CampoCupon
									{cupon}
									lineas={carrito.lineas}
									tinta="#FBF4F1"
									suave="#C9A9AF"
									borde="rgba(251,244,241,.2)"
									fondo="rgba(251,244,241,.08)"
									acento="var(--acento, #B4475B)"
									sobreAcento="#fff"
									radio="14px"
									fuente="'Hanken Grotesk',sans-serif"
									error="#FFB4AB"
								/>
							</div>
							<span
								style="display:flex;justify-content:space-between;align-items:baseline;margin-top:6px"
								><b>Total</b><b style="font-family:'Gloock',serif;font-size:28px">{fmt(total)}</b
								></span
							>
						</div>
					</div>
				</form>
			{/if}
		{/if}

		<!-- ============ LISTO (no está en el diseño: con las piezas del checkout) ============ -->
		{#if pagina === 'listo' && pedido}
			<section
				style="padding:clamp(40px,6cqw,80px) clamp(14px,3.5cqw,44px);max-width:720px;margin:0 auto;text-align:center"
			>
				<span
					style="display:inline-grid;place-items:center;width:64px;height:64px;border-radius:50%;background:var(--acento, #B4475B);color:#fff"
					><Icon icon={I.phCheckBold} size={28} /></span
				>
				<h1
					style="margin:18px 0 0;font-family:'Gloock',serif;font-weight:400;font-size:clamp(38px,5cqw,62px);line-height:1;overflow-wrap:anywhere"
				>
					¡Gracias, {pedido.nombre}!
				</h1>
				<p style="margin:14px 0 0;font-size:16px;line-height:1.55;color:#6E4F55">
					Recibimos tu pedido <b>#{pedido.numero}</b> por <b>{fmt(pedido.total)}</b>. Te escribimos
					por mail con cada&nbsp;novedad.
				</p>
				<SenaPedido {tema} />
				{#if pedido.pago === 'transferencia'}
					<div
						style="margin-top:22px;padding:20px;border-radius:22px;background:#fff;text-align:left;font-size:14.5px;line-height:1.6"
					>
						<b>Transferí {fmt(pedido.total)}</b>
						{#if pedido.alias}al {pedido.esCbu ? 'CBU/CVU' : 'alias'}
							<b>{pedido.alias}</b>{:else}con los datos que te mandamos por mail{/if} y mandanos el comprobante{#if whatsapp}&nbsp;<a
								href={wa(`¡Hola! Te mando el comprobante del pedido #${pedido.numero}.`)}
								target="_blank"
								rel="noopener"
								style="color:var(--acento, #B4475B);font-weight:700">por WhatsApp</a
							>{:else}&nbsp;respondiendo el mail del pedido{/if}. Lo preparamos apenas
						se&nbsp;acredite.
					</div>
				{/if}
				<a
					href={base || '/'}
					style="display:inline-block;margin-top:26px;padding:16px 26px;border-radius:999px;background:#3A1F25;color:#fff;font:800 15px 'Hanken Grotesk';text-decoration:none"
					>Volver a la tienda</a
				>
				{#if pedido.estado === 'Entregado'}
					<div style="margin-top:clamp(36px,5cqw,56px)">
						<FormResena slug={tienda.slug} {tema} codigo={pedido.codigo} nombre={pedido.nombre} />
					</div>
				{/if}
			</section>
		{/if}

		<!-- ============ TURNOS (con las piezas del checkout) ============ -->
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
				style="padding:24px clamp(14px,3.5cqw,44px) clamp(48px,7cqw,90px);display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,360px),1fr));gap:clamp(16px,3cqw,40px);align-items:start"
			>
				<input type="hidden" name="servicioId" value={reserva.servicioId ?? ''} />
				<input type="hidden" name="profesionalId" value={reserva.profesionalId ?? ''} />
				<input type="hidden" name="inicio" value={reserva.inicio ?? ''} />
				<div style="display:flex;flex-direction:column;gap:22px;min-width:0">
					<div style="display:flex;gap:6px">
						{#each PASOS_TURNO as t, i (t)}<span
								title={t}
								style="flex:1;height:5px;border-radius:3px;background:{i <= reserva.paso
									? 'var(--acento, #B4475B)'
									: '#EAD3CF'}"
							></span>{/each}
					</div>
					<h1
						style="margin:0;font-family:'Gloock',serif;font-weight:400;font-size:clamp(38px,5cqw,62px);line-height:1"
					>
						Reservá tu turno
					</h1>
					<div>
						<div style="font-size:15px;font-weight:800">1 · Servicio</div>
						<div
							style="margin-top:10px;display:flex;flex-direction:column;gap:8px"
							role="radiogroup"
							aria-label="Servicio"
						>
							{#each tienda.turnos.servicios as s (s.id)}{@const on =
									reserva.servicioId === s.id}<button
									type="button"
									role="radio"
									aria-checked={on}
									onclick={() => reserva.elegirServicio(s.id)}
									style="display:flex;align-items:center;gap:12px;width:100%;padding:15px;border-radius:18px;border:1.5px solid {on
										? 'var(--acento, #B4475B)'
										: '#EAD3CF'};background:{on
										? '#FFF6F7'
										: '#fff'};text-align:left;cursor:pointer;font-family:inherit;color:#3A1F25"
									><span
										style="width:20px;height:20px;border-radius:50%;border:2px solid {on
											? 'var(--acento, #B4475B)'
											: '#D9BEC3'};display:grid;place-items:center;flex:none"
										>{#if on}<span
												style="width:10px;height:10px;border-radius:50%;background:var(--acento, #B4475B)"
											></span>{/if}</span
									><span style="flex:1;min-width:0"
										><span style="display:block;font-size:14.5px;font-weight:700">{s.nombre}</span
										><span style="display:block;font-size:12.5px;color:#8C6A70;"
											>{duracion(s.duracion)}</span
										></span
									><b style="white-space:nowrap">{fmt(s.precio)}</b></button
								>{/each}
						</div>
					</div>
					<div>
						<div style="font-size:15px;font-weight:800">2 · Profesional</div>
						<div
							style="margin-top:12px;display:flex;gap:8px;overflow-x:auto;scrollbar-width:none;padding-bottom:2px"
							role="radiogroup"
							aria-label="Profesional"
						>
							{#each [null, ...reserva.profesionales] as p (p?.id ?? 'cualquiera')}{@const on =
									reserva.profesionalId === (p?.id ?? null)}<button
									type="button"
									role="radio"
									aria-checked={on}
									onclick={() => reserva.elegirProfesional(p?.id ?? null)}
									style="flex:none;width:110px;padding:10px;border-radius:18px;border:1.5px solid {on
										? 'var(--acento, #B4475B)'
										: '#EAD3CF'};background:{on
										? '#FFF6F7'
										: '#fff'};text-align:center;cursor:pointer;font-family:inherit;color:#3A1F25"
									>{#if p?.foto}<img
											src={img(p.foto, 120, 120)}
											alt=""
											style="width:56px;height:56px;border-radius:50%;object-fit:cover;display:block;margin:0 auto"
										/>{:else}<span
											style="width:56px;height:56px;border-radius:50%;background:#F3D9DE;color:var(--acento, #B4475B);display:grid;place-items:center;margin:0 auto"
											><Icon icon={p ? I.phUserFill : I.phShuffleBold} size={22} /></span
										>{/if}<span
										style="display:block;margin-top:8px;font-size:13.5px;font-weight:700;overflow:hidden;text-overflow:ellipsis;white-space:nowrap"
										>{p?.nombre ?? 'Cualquiera'}</span
									><span
										style="display:block;font-size:11.5px;color:#8C6A70;overflow:hidden;text-overflow:ellipsis;white-space:nowrap"
										>{p ? p.rol.split(' · ')[0] : 'El primero libre'}</span
									></button
								>{/each}
						</div>
					</div>
					<div>
						<div style="display:flex;justify-content:space-between;align-items:center;gap:10px">
							<div style="font-size:15px;font-weight:800">3 · Día y horario</div>
							<span style="display:flex;align-items:center;gap:8px;font-size:14px;font-weight:600"
								>{#if reserva.hayAntes || reserva.hayDespues}<button
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
									>{/if}</span
							>
						</div>
						<div
							style="margin-top:10px;display:grid;grid-template-columns:repeat(7,minmax(0,1fr));gap:6px"
							role="radiogroup"
							aria-label="Día"
						>
							{#each reserva.dias as d (d.dia)}{@const on = reserva.dia === d.dia}<button
									type="button"
									role="radio"
									aria-checked={on}
									aria-label="{d.w} {d.n}{d.abierto ? '' : ', cerrado'}"
									disabled={!d.abierto}
									onclick={() => reserva.elegirDia(d.dia)}
									style="padding:9px 0;text-align:center;cursor:pointer;font-family:inherit;min-width:0;border-radius:14px;border:1.5px solid {on
										? 'var(--acento, #B4475B)'
										: '#EAD3CF'};background:{on ? 'var(--acento, #B4475B)' : '#fff'};color:{on
										? '#fff'
										: '#3A1F25'}"
									><span
										style="display:block;font-size:11px;font-weight:700;letter-spacing:.06em;text-transform:uppercase"
										>{d.w}</span
									><span style="display:block;font-size:17px;font-weight:700;margin-top:1px"
										>{d.n}</span
									></button
								>{/each}
						</div>
						<div style="margin-top:14px;display:flex;flex-direction:column;gap:12px">
							{#if !reserva.servicioId}
								<p style="margin:0;font-size:14px;color:#8C6A70;">
									Elegí un servicio para ver los horarios.
								</p>
							{:else if reserva.horarios && !reserva.grupos.length}
								<p style="margin:0;font-size:14px;color:#8C6A70;">
									No quedan horarios este día. Probá con otro día u otro&nbsp;profesional.
								</p>
							{/if}
							{#each reserva.grupos as g (g.k)}<div>
									<div
										style="font-size:12px;font-weight:700;letter-spacing:.08em;text-transform:uppercase;margin-bottom:7px;color:#8C6A70;"
									>
										{g.t}
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
												style="padding:10px 0;text-align:center;font-family:inherit;font-size:14px;font-weight:700;cursor:pointer;text-decoration:{h.libre
													? 'none'
													: 'line-through'};border-radius:14px;border:1.5px solid {on
													? 'var(--acento, #B4475B)'
													: '#EAD3CF'};background:{on ? 'var(--acento, #B4475B)' : '#fff'};color:{on
													? '#fff'
													: '#3A1F25'}">{h.hora}</button
											>{/each}
									</div>
								</div>{/each}
						</div>
					</div>
					{#if reserva.listo}
						<div>
							<div style="font-size:15px;font-weight:800">4 · Tus datos</div>
							<div
								style="margin-top:10px;display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));gap:8px"
							>
								{#each [['nombre', 'Nombre', 'text', 'name'], ['telefono', 'WhatsApp', 'tel', 'tel'], ['email', 'Email', 'email', 'email']] as [n, l, tipo, ac] (n)}
									<label
										style="display:block;padding:11px 14px;border-radius:16px;background:#fff;box-shadow:inset 0 0 0 1.5px #EAD3CF"
										><span style="display:block;font-size:11.5px;color:#A88A8F">{l}</span><input
											name={n}
											type={tipo}
											autocomplete={ac as 'email'}
											required
											maxlength="120"
											style="width:100%;margin-top:2px;padding:0;border:0;background:transparent;font:400 14.5px 'Hanken Grotesk';color:#3A1F25;outline:none"
										/></label
									>
								{/each}
							</div>
						</div>
						<div>
							<div style="font-size:15px;font-weight:800">Pago</div>
							<div style="margin-top:10px;display:flex;flex-direction:column;gap:8px">
								<div
									style="display:flex;align-items:center;gap:12px;width:100%;padding:15px;border-radius:18px;border:1.5px solid var(--acento, #B4475B);background:#FFF6F7;text-align:left;cursor:pointer;font-family:inherit;color:#3A1F25"
								>
									<span
										style="width:20px;height:20px;border-radius:50%;border:2px solid var(--acento, #B4475B);display:grid;place-items:center;flex:none"
										><span
											style="width:10px;height:10px;border-radius:50%;background:var(--acento, #B4475B)"
										></span></span
									>{#if reserva.sena}<Icon
											icon={I.siMercadopago}
											size={22}
											style="color:#00B1EA;flex:none"
										/><span style="flex:1;min-width:0"
											><span style="display:block;font-size:14.5px;font-weight:700"
												>Mercado Pago</span
											><span style="display:block;font-size:12.5px;color:#8C6A70;"
												>Seña de {fmt(reserva.sena)} para confirmar · el resto, en el local</span
											></span
										>{:else}<Icon icon={I.phStorefrontFill} size={22} style="flex:none" /><span
											style="flex:1;min-width:0"
											><span style="display:block;font-size:14.5px;font-weight:700"
												>Pagás en el local</span
											><span style="display:block;font-size:12.5px;color:#8C6A70;"
												>No hace falta seña: el turno queda confirmado</span
											></span
										>{/if}
								</div>
							</div>
						</div>
					{/if}
				</div>
				<div
					style="position:sticky;top:90px;padding:22px;border-radius:28px;background:#3A1F25;color:#FBF4F1;min-width:0"
				>
					<div style="font-family:'Gloock',serif;font-size:26px">Tu reserva</div>
					<div style="margin-top:14px;display:flex;align-items:center;gap:12px">
						{#if reserva.profesional?.foto}<img
								src={img(reserva.profesional.foto, 120, 120)}
								alt=""
								style="width:52px;height:52px;border-radius:50%;object-fit:cover;flex:none"
							/>{/if}
						<div style="min-width:0">
							<div style="font-size:15px;font-weight:700">
								{reserva.conQuien ? `Con ${reserva.conQuien}` : 'Con el primero libre'}
							</div>
							<div style="font-size:13px;color:#D9BEC3;">
								{reserva.cuando ?? 'Elegí el día y el horario'}
							</div>
						</div>
					</div>
					{#if reserva.servicio}
						<div
							style="margin-top:16px;padding-top:14px;border-top:1px solid rgba(251,244,241,.14);display:flex;flex-direction:column;gap:8px;font-size:14px"
						>
							<span style="display:flex;justify-content:space-between;gap:10px"
								><span>{reserva.servicio.nombre}</span><span style="white-space:nowrap"
									>{fmt(reserva.servicio.precio)}</span
								></span
							><span style="display:flex;justify-content:space-between;color:#D9BEC3;"
								><span>Duración total</span><span>{reserva.duracion}</span></span
							><span
								style="display:flex;justify-content:space-between;align-items:baseline;margin-top:6px"
								><b>Total</b><b style="font-family:'Gloock',serif;font-size:28px;font-weight:400"
									>{fmt(reserva.servicio.precio)}</b
								></span
							>
						</div>
					{/if}
					<p style="margin:14px 0 0;font-size:13px;line-height:1.5;color:#D9BEC3;">
						{#if reserva.sena}Seña de {fmt(reserva.sena)} para confirmar. Podés cancelar desde el link
							del turno hasta {anticipacion} antes; la seña te la devuelve la&nbsp;tienda.{:else}No
							pagás nada ahora. Podés cancelar desde el link del turno hasta {anticipacion}&nbsp;antes.{/if}
					</p>
					{#if error}<div
							role="alert"
							style="margin-top:12px;padding:12px 14px;border-radius:16px;background:#FFF6F7;box-shadow:inset 0 0 0 1.5px var(--acento, #B4475B);color:var(--acento, #B4475B);font-size:14px;font-weight:600"
						>
							{error}
						</div>{/if}
					<button
						disabled={!reserva.listo || reservando || tienda.vitrina}
						style="margin-top:16px;width:100%;padding:17px;border:0;border-radius:999px;background:var(--acento, #B4475B);color:#fff;font:800 15.5px 'Hanken Grotesk';cursor:pointer"
						>{tienda.vitrina
							? 'Es una tienda de ejemplo'
							: !reserva.listo
								? 'Elegí servicio y horario'
								: reserva.sena
									? 'Confirmar y pagar seña'
									: 'Confirmar turno'}</button
					>
				</div>
			</form>
		{/if}

		<!-- ============ TURNO (no está en el diseño: con las piezas de 'listo') ============ -->
		{#if pagina === 'turno' && turno}
			{@const cancelado = turno.estado === 'cancelado'}
			<section
				style="padding:clamp(40px,6cqw,80px) clamp(14px,3.5cqw,44px);max-width:720px;margin:0 auto;text-align:center"
			>
				<span
					style="display:inline-grid;place-items:center;width:64px;height:64px;border-radius:50%;background:{cancelado
						? '#3A1F25'
						: 'var(--acento, #B4475B)'};color:#fff"
					><Icon icon={cancelado ? I.phCalendarX : I.phCheckBold} size={28} /></span
				>
				<h1
					style="margin:18px 0 0;font-family:'Gloock',serif;font-weight:400;font-size:clamp(38px,5cqw,62px);line-height:1;overflow-wrap:anywhere"
				>
					{cancelado
						? 'Turno cancelado'
						: turno.estado === 'reservado'
							? `Falta la seña, ${turno.nombre}`
							: `¡Listo, ${turno.nombre}!`}
				</h1>
				<p style="margin:14px 0 0;font-size:16px;line-height:1.55;color:#6E4F55">
					{#if cancelado}Liberamos el horario{turno.senaPagada
							? '. La seña te la devuelve la tienda por Mercado Pago'
							: ''}. Si querés, reservá&nbsp;otro.{:else if turno.estado === 'reservado'}Te
						guardamos el horario unos minutos mientras pagás la seña con Mercado&nbsp;Pago.{:else}Tu
						turno está confirmado. Te esperamos en {turno.tienda}.{/if}
				</p>
				{#if aviso === 'pendiente' || aviso === 'error'}<div
						role="status"
						style="margin-top:14px;padding:12px 14px;border-radius:16px;background:#FFF6F7;box-shadow:inset 0 0 0 1.5px var(--acento, #B4475B);color:var(--acento, #B4475B);font-size:14px;font-weight:600"
					>
						{aviso === 'pendiente'
							? 'Mercado Pago todavía no acreditó la seña: te avisamos cuando entre.'
							: 'El pago de la seña no salió. Podés reservar de nuevo.'}
					</div>{/if}
				<div
					style="margin-top:22px;padding:20px;border-radius:22px;background:#fff;text-align:left;font-size:14.5px;line-height:1.6"
				>
					<div
						style="font-size:16px;font-weight:700;text-decoration:{cancelado
							? 'line-through'
							: 'none'}"
					>
						{cuandoLargo(turno.inicio)} con {turno.profesional}
					</div>
					<div style="margin-top:4px;font-size:14px;color:#8C6A70;">
						{turno.servicio} · {fmt(turno.precio)}{turno.sena
							? ` · seña ${fmt(turno.sena)}${turno.senaPagada ? ' pagada' : ''}`
							: ' · pagás en el local'}
					</div>
					{#if turno.direccion}<div style="margin-top:8px;display:flex;gap:8px;font-size:14px">
							<Icon icon={I.phMapPinFill} size={17} style="flex:none" />{turno.direccion}
						</div>{/if}
				</div>
				{#if error}<div
						role="alert"
						style="margin-top:12px;padding:12px 14px;border-radius:16px;background:#FFF6F7;box-shadow:inset 0 0 0 1.5px var(--acento, #B4475B);color:var(--acento, #B4475B);font-size:14px;font-weight:600"
					>
						{error}
					</div>{/if}
				<div style="margin-top:22px;display:flex;gap:10px;flex-wrap:wrap;justify-content:center;">
					{#if !cancelado}<a
							href="{base}/turno/{turno.codigo}/turno.ics"
							download="turno.ics"
							data-sveltekit-reload
							style="display:flex;align-items:center;gap:8px;padding:16px 26px;border:0;border-radius:999px;background:var(--acento, #B4475B);color:#fff;font:800 15px 'Hanken Grotesk';text-decoration:none;cursor:pointer"
							><Icon icon={I.phCalendarPlusBold} size={17} />Agregar al calendario</a
						>{/if}
					{#if turno.cancelable}
						<form
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
									style="display:inline-block;padding:16px 26px;border:0;border-radius:999px;background:#3A1F25;color:#fff;font:800 15px 'Hanken Grotesk';text-decoration:none;cursor:pointer"
									>Sí, cancelar el turno</button
								>{:else}<button
									type="button"
									onclick={() => (seguro = true)}
									style="display:inline-block;padding:16px 26px;border:0;border-radius:999px;background:#3A1F25;color:#fff;font:800 15px 'Hanken Grotesk';text-decoration:none;cursor:pointer"
									>Cancelar turno</button
								>{/if}
						</form>
					{:else if cancelado}<a
							href="{base}/turnos"
							style="padding:16px 26px;border:0;border-radius:999px;background:var(--acento, #B4475B);color:#fff;font:800 15px 'Hanken Grotesk';text-decoration:none;cursor:pointer"
							>Reservar otro</a
						>{/if}
					<a
						href={base || '/'}
						style="display:inline-block;padding:16px 26px;border:0;border-radius:999px;background:#3A1F25;color:#fff;font:800 15px 'Hanken Grotesk';text-decoration:none;cursor:pointer"
						>Volver a la tienda</a
					>
				</div>
				{#if !cancelado && !turno.cancelable}<p
						style="margin:14px 0 0;font-size:13.5px;color:#8C6A70;"
					>
						Para cambiarlo o cancelarlo, escribile a la&nbsp;tienda{#if whatsapp}&nbsp;<a
								href={wa(`¡Hola! Tengo un turno el ${cuandoLargo(turno.inicio)}.`)}
								target="_blank"
								rel="noopener"
								style="color:inherit;font-weight:700">por WhatsApp</a
							>{/if}.
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
			<section style="padding:clamp(30px,5cqw,64px) clamp(14px,3.5cqw,44px) 0">
				<div
					style="display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,300px),1fr));gap:12px"
				>
					<div
						style="grid-column:{fotoNos
							? `span ${abSpan}`
							: '1 / -1'};padding:clamp(24px,4cqw,52px);border-radius:32px;background:#3A1F25;color:#FBF4F1;display:flex;flex-direction:column;justify-content:flex-end;min-height:360px;min-width:0"
					>
						<div
							style="font-size:12.5px;font-weight:800;letter-spacing:.24em;text-transform:uppercase;color:#F3D9DE"
						>
							{tienda.nombre}{tienda.ciudad ? ` · ${tienda.ciudad}` : ''}
						</div>
						<h1
							style="margin:14px 0 0;font-family:'Gloock',serif;font-weight:400;font-size:clamp(42px,6cqw,84px);line-height:.95;overflow-wrap:anywhere"
						>
							{tienda.portada.titulo}
						</h1>
						{#if tienda.portada.lema}<p
								style="margin:16px 0 0;max-width:560px;font-size:16px;line-height:1.55;color:#F3D9DE"
							>
								{tienda.portada.lema}
							</p>{/if}
					</div>
					{#if fotoNos}<div
							style="border-radius:32px;overflow:hidden;background:#F3D9DE;min-height:360px;position:relative"
						>
							<img
								src={img(fotoNos, 800)}
								alt=""
								style="position:absolute;inset:0;width:100%;height:100%;object-fit:cover"
							/>
						</div>{/if}
				</div>
				{#if stats.length}<div
						style="margin-top:12px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,220px),1fr));gap:12px"
					>
						{#each stats as s (s.t)}<div style="padding:22px;border-radius:26px;background:{s.bg}">
								<div
									style="font-family:'Gloock',serif;font-size:clamp(34px,4cqw,48px);line-height:1"
								>
									{s.v}
								</div>
								<div style="margin-top:6px;font-size:14px;color:#6E4F55">{s.t}</div>
							</div>{/each}
					</div>{/if}
			</section>
			<section
				style="padding:clamp(40px,6cqw,80px) clamp(14px,3.5cqw,44px);display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,320px),1fr));gap:12px"
			>
				<div
					style="padding:24px;border-radius:28px;background:#fff;display:flex;flex-direction:column;gap:12px"
				>
					<div style="font-family:'Gloock',serif;font-size:28px">
						{tienda.envios.retiro ? 'El local' : 'Envíos'}
					</div>
					{#if tienda.direccion}<span style="display:flex;gap:10px;font-size:15px"
							><Icon
								icon={I.phMapPinFill}
								size={18}
								style="color:var(--acento, #B4475B);flex:none"
							/>{tienda.direccion}</span
						>{/if}
					{#if envia(tienda.envios)}<span style="display:flex;gap:10px;font-size:15px"
							><Icon
								icon={I.phTruck}
								size={18}
								style="color:var(--acento, #B4475B);flex:none"
							/>{tienda.envios.correo
								? 'Enviamos a todo el país'
								: entregaCorta(tienda.envios)}{tienda.envios.gratisDesde
								? `, gratis desde ${fmt(tienda.envios.gratisDesde)}`
								: ''}</span
						>{/if}
					{#if tienda.envios.retiro}<span style="display:flex;gap:10px;font-size:15px"
							><Icon
								icon={I.phStorefrontFill}
								size={18}
								style="color:var(--acento, #B4475B);flex:none"
							/>Retiro sin costo en el local</span
						>{/if}
					{#if tienda.cobros.cuotas > 1}<span style="display:flex;gap:10px;font-size:15px"
							><Icon
								icon={I.phCreditCard}
								size={18}
								style="color:var(--acento, #B4475B);flex:none"
							/>{tienda.cobros.cuotas} cuotas sin interés{#if tienda.cobros.transferencia > 0}&nbsp;·
								{tienda.cobros.transferencia}% off por transferencia{/if}</span
						>{:else if tienda.cobros.transferencia > 0}<span
							style="display:flex;gap:10px;font-size:15px"
							><Icon
								icon={I.phBank}
								size={18}
								style="color:var(--acento, #B4475B);flex:none"
							/>{tienda.cobros.transferencia}% off por transferencia</span
						>{/if}
				</div>
				{#if whatsapp}
					<div
						style="padding:24px;border-radius:28px;background:#F3D9DE;display:flex;flex-direction:column;gap:10px"
					>
						<div style="font-family:'Gloock',serif;font-size:28px">Hablá con nosotros</div>
						<span style="font-size:14.5px;line-height:1.5;color:#6E4F55"
							>Escribinos por WhatsApp y te ayudamos a elegir o a seguir tu&nbsp;pedido.</span
						>
						<span style="flex:1"></span>
						<a
							href={wa('¡Hola!')}
							target="_blank"
							rel="noopener"
							style="display:flex;align-items:center;gap:10px;padding:14px;border-radius:16px;background:#25D366;color:#fff;font-weight:800;text-decoration:none"
							><Icon icon={I.phWhatsappLogoFill} size={22} />{tienda.whatsapp}</a
						>
					</div>
				{/if}
			</section>
		{/if}

		<!-- footer -->
		<footer style="padding:clamp(40px,6cqw,72px) clamp(14px,3.5cqw,44px) 26px;background:#F3D9DE">
			<div
				style="display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,190px),1fr));gap:28px"
			>
				<div style="min-width:0;max-width:360px">
					<div style="display:flex;align-items:baseline;gap:6px;flex-wrap:wrap">
						<span style="font-family:'Gloock',serif;font-size:32px">{logo[0]}</span
						>{#if logo[1]}<span
								style="font-size:11px;font-weight:800;letter-spacing:.3em;text-transform:uppercase;color:var(--acento, #B4475B)"
								>{logo[1]}</span
							>{/if}
					</div>
					<p style="margin:10px 0 0;font-size:14px;color:#6E4F55;line-height:1.55">
						{#if tienda.cobros.cuotas > 1}{tienda.cobros.cuotas} cuotas sin interés{#if tienda.cobros.transferencia > 0}&nbsp;y
								{tienda.cobros.transferencia}% off por transferencia{/if}, en toda la&nbsp;tienda.{:else if tienda.cobros.transferencia > 0}{tienda
								.cobros.transferencia}% off por transferencia, en toda la&nbsp;tienda.{:else}Compra
							protegida, con 10 días para&nbsp;arrepentirte.{/if}
					</p>
					<!-- Newsletter (§50): el bloque del pie del diseño; queda pendiente hasta confirmar por mail. -->
					<form
						onsubmit={nl.enviar}
						novalidate
						style="margin-top:12px;display:flex;padding:5px;border-radius:999px;background:#fff"
					>
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
							style="--ph:#A88A8F;flex:1;min-width:0;padding:9px 12px;border:0;background:transparent;font:inherit;font-size:14px;color:#3A1F25;outline:none"
						/>
						<button
							type="submit"
							disabled={nl.enviando}
							style="flex:none;padding:9px 16px;border:0;border-radius:999px;background:#3A1F25;color:#fff;font:inherit;font-size:13.5px;font-weight:700;cursor:pointer"
							>Sumarme</button
						>
					</form>
					{#if nl.aviso}<p
							role="status"
							style="margin:8px 0 0;font-size:13px;line-height:1.45;color:{nl.aviso.ok
								? '#6E4F55'
								: '#A3361F'}"
						>
							{nl.aviso.texto}{#if nl.aviso.link}&nbsp;<a
									href={nl.aviso.link}
									style="color:inherit;text-decoration:underline;text-underline-offset:3px;font-weight:600"
									>Confirmar acá</a
								>{/if}
						</p>{/if}
				</div>
				{#each foot as f (f.t)}<div
						style="display:flex;flex-direction:column;gap:9px;font-size:14px"
					>
						<b
							style="font-size:12px;letter-spacing:.14em;text-transform:uppercase;margin-bottom:4px"
							>{f.t}</b
						>{#each f.l as l (l.t)}<a href={l.href} style="color:#6E4F55;text-decoration:none"
								>{l.t}</a
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
				style="margin-top:36px;padding-top:16px;border-top:1px solid rgba(58,31,37,.12);display:flex;justify-content:space-between;gap:14px;flex-wrap:wrap;font-size:12.5px;color:#8C6A70"
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
						style="color:inherit;opacity:1;text-decoration:underline;text-underline-offset:3px;font-weight:700;display:inline-flex;align-items:center;gap:5px"
						><Icon icon={I.phArrowCounterClockwiseBold} size={13} />Botón de arrepentimiento</a
					><a
						href="/legales/defensa-del-consumidor"
						style="color:inherit;opacity:0.75;text-decoration:underline;text-underline-offset:3px;font-weight:500;display:inline-flex;align-items:center;gap:5px"
						>Defensa de las y los consumidores</a
					><a
						href="/legales/terminos"
						style="color:inherit;opacity:0.75;text-decoration:underline;text-underline-offset:3px;font-weight:500;display:inline-flex;align-items:center;gap:5px"
						>Términos y condiciones</a
					><a
						href="/legales/privacidad"
						style="color:inherit;opacity:0.75;text-decoration:underline;text-underline-offset:3px;font-weight:500;display:inline-flex;align-items:center;gap:5px"
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
				style="position:fixed;right:18px;bottom:18px;z-index:30;width:54px;height:54px;border-radius:50%;background:#25D366;color:#fff;display:grid;place-items:center;box-shadow:0 12px 26px -8px rgba(0,0,0,.35)"
				><Icon icon={I.phWhatsappLogoFill} size={28} /></a
			>{/if}
	</div>
</div>

<!-- eslint-enable svelte/no-navigation-without-resolve -->

<style>
	.rosedal :global(a:hover) {
		color: var(--acento, #b4475b);
	}
	.zoom :global(img) {
		transition: transform 0.8s cubic-bezier(0.2, 0.7, 0.1, 1);
	}
	.zoom:hover :global(img) {
		transform: scale(1.05);
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
	.oculto {
		position: absolute;
		width: 1px;
		height: 1px;
		overflow: hidden;
		clip: rect(0 0 0 0);
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
