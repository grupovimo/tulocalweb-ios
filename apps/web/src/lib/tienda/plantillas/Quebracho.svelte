<script lang="ts">
	import { camposFactura, OPCIONALES } from '#lib/tienda/factura.ts';
	import { envia } from '../envio.ts';
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
	import { proximamente, senaDe } from '../lanzamiento.ts';
	import PreguntasProducto from '../PreguntasProducto.svelte';
	import type { TemaOpiniones } from '../opiniones.ts';
	import { carritoDe } from '../carrito.svelte.ts';
	import CampoCupon from '../CampoCupon.svelte';
	import { CuponCheckout, totales } from '../cupon.svelte.ts';
	import EnvioCorreo from '../EnvioCorreo.svelte';
	import { comoLlega, cotizar, cotizarYa, cpValido, textoEnvios } from '../envio.svelte.ts';
	import { Seleccion, tarjetaDe } from '../seleccion.svelte.ts';
	import { partirTitulo } from '../titulo.ts';
	import { Reserva } from '../reserva.svelte.ts';
	import { cuandoLargo, duracion } from '#lib/turnos/formato.ts';
	import type { OpcionEnvio, ProductoPublico, TiendaPublica } from '../tipos.ts';

	/**
	 * Plantilla Quebracho (design/Plantilla Quebracho.dc.html), con los datos de la tienda en lugar de
	 * los de ejemplo. Lo que el diseño muestra sin datos atrás (puntos sobre la foto del ambiente,
	 * diseño a medida, maderas y telas, franjas de entrega, seña, favoritos, contadores) no se dibuja.
	 * El diseño no tiene newsletter: el checkout ofrece la casilla «Quiero recibir novedades» (§50).
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
	const off = (p: ProductoPublico) =>
		p.anterior ? Math.round((1 - p.precio / p.anterior) * 100) : 0;
	// Con variantes de precios distintos, "Desde"; "Sin stock" sólo si no queda ninguna.
	const tarjeta = (p: ProductoPublico) => {
		const t = tarjetaDe(p);
		const sale = !!p.anterior && !t.desde;
		return {
			id: p.id,
			n: p.nombre,
			price: (t.desde ? 'Desde ' : '') + fmt(t.precio),
			old: sale ? fmt(p.anterior!) : null,
			img: img(p.foto, 640, 800),
			w: p.categoria,
			cuota: tienda.cobros.cuotas > 1 ? fmt(t.precio / tienda.cobros.cuotas) : '',
			tag: t.agotado ? 'Sin stock' : sale ? `−${off(p)}%` : p.etiqueta,
			href: `${base}/producto/${p.slug ?? p.id}`
		};
	};
	const ver = (s: string) => tienda.secciones.includes(s as TiendaPublica['secciones'][number]);
	const whatsapp = $derived(tienda.whatsapp.replace(/\D/g, ''));
	const wa = (texto: string) =>
		`https://wa.me/${whatsapp.startsWith('54') ? whatsapp : '54' + whatsapp}?text=${encodeURIComponent(texto)}`;

	// Título de la portada: el último tramo va en itálica verde, como en el diseño.
	const titulo = $derived(partirTitulo(tienda.portada.titulo));

	const lista = $derived(
		tienda.masVendidos.slice(0, 4).flatMap((id) => (porId.get(id) ? [porId.get(id)!] : []))
	);
	// La franja oscura del diseño ("día por día"), con lo que la tienda ofrece de verdad.
	const process = $derived([
		...(tienda.cobros.cuotas > 1
			? [
					{
						d: String(tienda.cobros.cuotas),
						i: I.phCreditCard,
						t: 'Cuotas sin interés',
						x: 'Con todas las tarjetas.'
					}
				]
			: []),
		...(tienda.cobros.transferencia > 0
			? [
					{
						d: `${tienda.cobros.transferencia}%`,
						i: I.phBank,
						t: 'Off por transferencia',
						x: 'Te mandamos el alias al confirmar.'
					}
				]
			: []),
		...(tienda.envios.gratisDesde
			? [
					{
						d: 'Gratis',
						i: I.phTruck,
						t: 'Envío gratis',
						x: `En compras desde ${fmt(tienda.envios.gratisDesde)}.`
					}
				]
			: []),
		...(tienda.envios.retiro
			? [
					{
						d: '$ 0',
						i: I.phStorefrontFill,
						t: 'Retiro en el local',
						x: tienda.direccion || 'Sin costo.'
					}
				]
			: []),
		{ d: '10', i: I.phArrowUUpLeft, t: 'Días para arrepentirte', x: 'Sin dar explicaciones.' }
	]);
	const feat = $derived(
		tienda.masVendidos
			.slice(0, 3)
			.flatMap((id) => (porId.get(id) ? [tarjeta(porId.get(id)!)] : []))
			.map((p, k) => ({
				...p,
				ar: k === 1 ? '3/4' : '4/5',
				mt: k === 1 && !nar ? '60px' : '0px'
			}))
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
	const catalog = $derived(listado.slice(0, mostrar).map(tarjeta));
	const enlaceCat = (c: string | null, o = orden) => {
		// eslint-disable-next-line svelte/prefer-svelte-reactivity -- local, sólo arma el link
		const q = new URLSearchParams();
		if (c) q.set('cat', c);
		if (o !== 'vendidos') q.set('orden', o);
		return `${base}/catalogo${q.size ? '?' + q : ''}`;
	};
	const rooms = $derived([
		{ c: null, t: 'Todo', n: tienda.productos.length },
		...tienda.categorias.map((c) => ({ c: c.nombre, t: c.nombre, n: c.cantidad }))
	]);

	// Ficha.
	let abierto = $state<string>('Descripción');
	const acc = $derived(
		producto
			? [
					['Descripción', producto.descripcion || `${producto.nombre}, de ${tienda.nombre}.`],
					['Envíos', textoEnvios(tienda.envios)],
					[
						'Cambios y devoluciones',
						'Tenés 10 días desde que lo recibís para arrepentirte de la compra, sin dar explicaciones (botón de arrepentimiento, abajo de todo).'
					]
				].map(([t, d]) => ({ t, d, open: abierto === t, i: abierto === t ? I.phMinus : I.phPlus }))
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
	const facts = $derived([
		...(tienda.cobros.cuotas > 1
			? [{ v: String(tienda.cobros.cuotas), t: 'cuotas sin interés' }]
			: []),
		...(tienda.cobros.transferencia > 0
			? [{ v: `${tienda.cobros.transferencia}%`, t: 'off por transferencia' }]
			: []),
		{ v: '10 días', t: 'para arrepentirte' }
	]);
	// Medida, madera, terminación…: lo elegido, la variante, su precio, stock y foto (seleccion.svelte.ts).
	const sel = $derived(producto ? new Seleccion(producto) : null);
	// Configurador (el «A medida» del handoff, §52): un producto con opciones se arma paso a paso, y
	// si la tienda le prendió el encargo con seña, se encarga desde la barra del configurador.
	const armar = $derived(!!producto?.opciones.length && !proximamente(producto));
	let encargoAbierto = $state(false);
	const senaArmado = $derived(
		producto?.encargo && sel && !sel.desde ? senaDe(sel.precio, producto.encargo) : null
	);
	function encargar() {
		if (!sel) return;
		if (sel.falta) {
			sel.aviso = true;
			return;
		}
		encargoAbierto = true;
	}
	// El tachado no va con "Desde": no se sabe contra qué variante se compara.
	const tachado = $derived(sel && !sel.desde ? sel.anterior : null);
	const stockFicha = $derived(
		!sel
			? ''
			: sel.stock === 0
				? 'Sin stock'
				: sel.stock <= 3 && (!sel.conOpciones || sel.variante)
					? (sel.stock === 1 ? 'Queda 1' : `Quedan ${sel.stock}`) +
						(sel.variante ? `\u00a0en ${sel.texto.toLowerCase()}` : '')
					: 'En stock'
	);
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
	/** Sumar desde "Va bien con": con opciones, primero se elige en la ficha. */
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
	const ICONO_ENVIO = {
		andreani: I.phTruck,
		'correo-domicilio': I.phPackage,
		'correo-sucursal': I.phPackage,
		moto: I.phMopedFill,
		retiro: I.phStorefrontFill
	};
	const pagos = $derived([
		...(tienda.cobros.datosTransferencia
			? [
					{
						id: 'transferencia' as const,
						t: 'Transferencia',
						d: tienda.cobros.transferencia
							? `${tienda.cobros.transferencia}% off · te mandamos el alias al confirmar`
							: 'Te mandamos el alias al confirmar',
						i: I.phBank,
						c: 'var(--acento, #4F5B3A)'
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
			t: 'Piezas',
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
		{ t: 'Todo', href: `${base}/catalogo`, on: pagina === 'catalogo' && !categoria },
		...tienda.categorias.slice(0, 4).map((c) => ({
			t: c.nombre,
			href: enlaceCat(c.nombre),
			on: pagina === 'catalogo' && categoria === c.nombre
		})),
		{ t: 'Nosotros', href: `${base}/nosotros`, on: pagina === 'nosotros' }
	]);
	// Reseñas y preguntas con las piezas de Quebracho: Newsreader en los títulos, esquinas rectas.
	const tema: TemaOpiniones = {
		texto: '#2B2520',
		suave: '#6F6358',
		campo: '#F7F3EE',
		tarjeta: '#E6DED2',
		borde: 'rgba(43,37,32,.14)',
		acento: '#2B2520',
		sobreAcento: '#EFE9E1',
		estrella: 'var(--acento, #4F5B3A)',
		radio: '0px',
		fuente: "'Familjen Grotesk', sans-serif",
		titulos: "'Newsreader', serif",
		pesoTitulos: 400
	};
	const estrellas = (n: number) =>
		Array.from({ length: 5 }, (_, i) =>
			n >= i + 1 ? I.phStarFill : n >= i + 0.5 ? I.phStarHalfFill : I.phStar
		);
	// Segunda foto de Nosotros: una del feed o la del más vendido.
	const fotoNos = $derived(tienda.fotos[0] ?? lista[0]?.foto ?? null);
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
	// El nombre gigante del pie: que la palabra más larga entre entera en el renglón (antes se cortaba
	// a mitad de palabra, «ALMAC / EN»). En cqw, con el ancho de una letra de esta tipografía.
	const pieNombre = $derived(
		Math.min(14, Math.floor(190 / Math.max(4, ...tienda.nombre.split(/\s+/).map((w) => w.length))))
	);
</script>

<svelte:window bind:innerWidth={ancho} />

<svelte:head>
	<link
		href="https://fonts.googleapis.com/css2?family=Newsreader:ital,opsz,wght@0,6..72,400;0,6..72,500;1,6..72,400;1,6..72,500&family=Familjen+Grotesk:wght@400;500;600;700&display=swap"
		rel="stylesheet"
	/>
</svelte:head>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
<div
	class="quebracho"
	style="min-height:100vh;font-family:'Familjen Grotesk',system-ui,sans-serif;color:#2B2520"
>
	<div style="container-type:inline-size;width:100%;background:#EFE9E1;position:relative">
		<!-- barra promo -->
		<div
			style="display:flex;justify-content:center;align-items:center;gap:12px;flex-wrap:wrap;padding:9px 16px;background:var(--acento, #4F5B3A);color:#EFE9E1;font-size:13px;text-align:center"
		>
			{#if conteo.actual}<span>{conteo.texto}</span><span
					style="display:flex;gap:4px;font-variant-numeric:tabular-nums"
					>{#each conteo.partes(4) as c (c.u)}<span
							style="padding:2px 7px;background:rgba(0,0,0,.2)">{c.v}{c.u}</span
						>{/each}</span
				>{:else}<span
					>{tienda.envios.gratisDesde
						? `Envío gratis desde ${fmt(tienda.envios.gratisDesde)}`
						: tienda.cobros.cuotas > 1
							? `${tienda.cobros.cuotas} cuotas sin interés`
							: 'Compra protegida'}</span
				>{/if}{#if wide && tienda.cobros.cuotas > 1}<span style="opacity:.8"
					>{tienda.cobros.cuotas} cuotas sin interés{#if tienda.cobros.transferencia > 0}&nbsp;· {tienda
							.cobros.transferencia}% off por transferencia{/if}</span
				>{/if}
		</div>

		<!-- header -->
		<header
			style="position:sticky;top:0;z-index:20;display:flex;align-items:center;gap:22px;padding:16px clamp(16px,4cqw,52px);background:rgba(239,233,225,.92);backdrop-filter:blur(10px);border-bottom:1px solid rgba(43,37,32,.12)"
		>
			{#if nar}<a
					href="{base}/catalogo"
					aria-label="Catálogo"
					style="display:grid;flex:none;color:#2B2520"><Icon icon={I.phList} size={22} /></a
				>{/if}
			<a
				href={base || '/'}
				style="text-decoration:none;color:#2B2520;font-family:'Newsreader',serif;font-size:28px;font-style:italic;letter-spacing:-.02em;line-height:1;white-space:nowrap;flex:0 1 auto;min-width:0;overflow:hidden;text-overflow:ellipsis"
				><LogoTienda {tienda} /></a
			>
			{#if wide}<nav
					style="display:flex;gap:24px;margin-left:10px;font-size:14px;min-width:0;overflow:hidden"
				>
					{#each nav as n (n.t)}<a
							href={n.href}
							style="text-decoration:none;padding:4px 0;font:500 14px 'Familjen Grotesk';color:#2B2520;border-bottom:1px solid {n.on
								? '#2B2520'
								: 'transparent'};white-space:nowrap">{n.t}</a
						>{/each}
				</nav>{/if}
			<span style="flex:1"></span>
			{#if nar && tienda.turnos}<a
					href="{base}/turnos"
					aria-label="Reservar turno"
					style="display:grid;flex:none;color:#2B2520"
					><Icon icon={I.phCalendarCheck} size={23} /></a
				>{/if}
			<a
				href="{base}/carrito"
				aria-label="Carrito"
				style="position:relative;color:#2B2520;display:grid;flex:none"
				><Icon icon={I.phArmchair} size={23} />{#if carrito.unidades}<span
						style="position:absolute;top:-6px;right:-8px;min-width:17px;height:17px;border-radius:999px;background:var(--acento, #4F5B3A);color:#fff;font-size:10.5px;font-weight:700;display:grid;place-items:center"
						>{carrito.unidades}</span
					>{/if}</a
			>
		</header>

		{#if tienda.vitrina}
			<div style="padding:10px 16px;background:#E6DED2;font-size:13px;text-align:center">
				Es una tienda de ejemplo de la plantilla Quebracho: podés recorrerla, pero no se compra
				de&nbsp;verdad.
			</div>
		{/if}

		<!-- ============ INICIO ============ -->
		{#if pagina === 'inicio'}
			<!-- Las secciones en el orden que eligió la tienda en Mi tienda. -->
			{#each enOrden(tienda.secciones, ['hero', 'prods', 'revs', 'ig']) as seccion (seccion)}
				{#if seccion === 'hero'}
					{#if ver('hero')}
						<section style="padding:clamp(22px,3.4cqw,44px) clamp(16px,4cqw,52px) 0">
							{#if tienda.portada.lema || tienda.ciudad}<div
									style="display:flex;justify-content:space-between;align-items:flex-end;gap:20px;flex-wrap:wrap;padding-bottom:clamp(16px,2.4cqw,28px);border-bottom:1px solid rgba(43,37,32,.2)"
								>
									<div
										style="font-size:12.5px;letter-spacing:.18em;text-transform:uppercase;color:#6F6358"
									>
										{tienda.portada.lema}
									</div>
									{#if tienda.ciudad && !tienda.portada.lema.includes(tienda.ciudad)}<div
											style="font-size:12.5px;letter-spacing:.18em;text-transform:uppercase;color:#6F6358"
										>
											{tienda.ciudad}
										</div>{/if}
								</div>{/if}
							<h1
								style="margin:clamp(16px,2.4cqw,28px) 0 clamp(18px,2.6cqw,30px);font-family:'Newsreader',serif;font-weight:400;font-size:clamp(52px,9.6cqw,150px);line-height:.86;letter-spacing:-.045em;overflow-wrap:anywhere"
							>
								{titulo[1] ? `${titulo[0]} ` : titulo[0]}{#if titulo[1]}
									<i style="color:var(--acento, #4F5B3A)">{titulo[1]}</i>{/if}
							</h1>
							{#if tienda.portada.foto}<div
									style="position:relative;aspect-ratio:{nar
										? '4/5'
										: '21/9'};overflow:hidden;background:#D8CCBC"
								>
									<img
										src={img(tienda.portada.foto, 1900)}
										alt=""
										style="position:absolute;inset:0;width:100%;height:100%;object-fit:cover"
									/>
								</div>{/if}
							{#if ver('prods') && lista.length}<div
									style="display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,240px),1fr));border-bottom:1px solid rgba(43,37,32,.2)"
								>
									{#each lista as p, k (p.id)}<a
											href="{base}/producto/{p.slug ?? p.id}"
											style="display:flex;align-items:center;gap:12px;padding:16px 4px;border-right:1px solid rgba(43,37,32,.1);text-decoration:none;text-align:left;color:#2B2520"
											><span
												style="width:26px;height:26px;border-radius:50%;background:{k === 0
													? '#2B2520'
													: 'transparent'};color:{k === 0
													? '#EFE9E1'
													: '#2B2520'};box-shadow:inset 0 0 0 1px #2B2520;display:grid;place-items:center;font-size:12px;font-weight:700;flex:none"
												>{k + 1}</span
											><span style="flex:1;min-width:0;font-size:14.5px">{p.nombre}</span><b
												style="font-size:14px;white-space:nowrap"
												>{tarjetaDe(p).desde ? 'Desde ' : ''}{fmt(tarjetaDe(p).precio)}</b
											></a
										>{/each}
								</div>{/if}
							<div
								style="margin-top:16px;display:flex;justify-content:space-between;align-items:center;gap:14px;flex-wrap:wrap"
							>
								{#if tienda.cobros.cuotas > 1 || tienda.cobros.transferencia > 0}<span
										style="font-size:15px"
										>{#if tienda.cobros.cuotas > 1}{tienda.cobros.cuotas} cuotas sin interés{/if}{#if tienda.cobros.transferencia > 0}
											{#if tienda.cobros.cuotas > 1}·{/if}
											<span style="color:var(--acento, #4F5B3A);font-weight:600"
												>{tienda.cobros.transferencia}% off por transferencia</span
											>{/if}</span
									>{/if}<a
									href="{base}/catalogo"
									style="text-decoration:none;padding:15px 24px;border:0;background:#2B2520;color:#EFE9E1;font:600 14px 'Familjen Grotesk';letter-spacing:.04em"
									>Ver todas las piezas</a
								>
							</div>
						</section>
					{/if}

					<section
						style="margin-top:clamp(56px,8cqw,110px);padding:clamp(36px,5cqw,64px) clamp(16px,4cqw,52px);background:#2B2520;color:#EFE9E1"
					>
						<div
							style="display:flex;justify-content:space-between;align-items:end;gap:16px;flex-wrap:wrap"
						>
							<h2
								style="margin:0;font-family:'Newsreader',serif;font-weight:400;font-size:clamp(38px,5.4cqw,72px);line-height:.95;letter-spacing:-.03em"
							>
								Comprar, <i style="color:#C9B98F">sin vueltas.</i>
							</h2>
							<span style="font-size:14px;color:#BDB2A6;max-width:340px"
								>{comoLlega(tienda.envios)}</span
							>
						</div>
						<div
							style="margin-top:28px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,180px),1fr));gap:1px;background:rgba(239,233,225,.14)"
						>
							{#each process as s (s.t)}<div style="padding:18px 16px 20px;background:#2B2520">
									<div style="display:flex;justify-content:space-between;align-items:baseline">
										<span
											style="font-family:'Newsreader',serif;font-size:40px;line-height:1;color:#C9B98F"
											>{s.d}</span
										><Icon icon={s.i} size={22} style="color:#BDB2A6" />
									</div>
									<div style="margin-top:14px;font-size:15.5px;font-weight:600">{s.t}</div>
									<div style="margin-top:4px;font-size:13.5px;line-height:1.45;color:#BDB2A6">
										{s.x}
									</div>
								</div>{/each}
						</div>
					</section>
				{:else if seccion === 'prods'}
					{#if ver('prods') && feat.length}
						<section style="padding:clamp(56px,8cqw,110px) clamp(16px,4cqw,52px) 0">
							<div
								style="display:flex;justify-content:space-between;align-items:end;gap:16px;flex-wrap:wrap"
							>
								<h2
									style="margin:0;font-family:'Newsreader',serif;font-weight:400;font-size:clamp(38px,5.4cqw,72px);line-height:.95;letter-spacing:-.03em"
								>
									Las más elegidas
								</h2>
								<a
									href="{base}/catalogo"
									style="text-decoration:none;padding:0;font:600 14px 'Familjen Grotesk';color:#2B2520;border-bottom:1px solid #2B2520"
									>Ver las {tienda.productos.length} piezas</a
								>
							</div>
							<div
								style="margin-top:28px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,260px),1fr));gap:clamp(14px,2cqw,26px)"
							>
								{#each feat as p (p.id)}<a
										href={p.href}
										class="zoom"
										style="display:block;text-decoration:none;text-align:left;color:#2B2520;font-family:inherit;margin-top:{p.mt}"
										><span
											style="position:relative;display:block;aspect-ratio:{p.ar};overflow:hidden;background:#D8CCBC"
											>{#if p.img}<img
													src={p.img}
													alt=""
													style="width:100%;height:100%;object-fit:cover"
												/>{/if}{#if p.tag}<span
													style="position:absolute;top:10px;left:10px;padding:4px 9px;background:#EFE9E1;font-size:11.5px;font-weight:600"
													>{p.tag}</span
												>{/if}</span
										><span
											style="display:flex;justify-content:space-between;align-items:baseline;gap:10px;margin-top:14px"
											><span style="font-family:'Newsreader',serif;font-size:24px;line-height:1.05"
												>{p.n}</span
											><b style="font-size:15px;white-space:nowrap">{p.price}</b></span
										>{#if p.w || p.cuota || p.old}<span
												style="display:flex;gap:10px;flex-wrap:wrap;margin-top:6px;font-size:13px;color:#6F6358"
												>{#if p.w}<span>{p.w}</span>{#if p.cuota}<span>·</span
														>{/if}{/if}{#if p.cuota}<span
														>{tienda.cobros.cuotas} × {p.cuota} sin interés</span
													>{/if}{#if p.old}{#if p.w || p.cuota}<span>·</span>{/if}<s>{p.old}</s
													>{/if}</span
											>{/if}</a
									>{/each}
							</div>
						</section>
					{/if}
				{:else if seccion === 'revs'}
					{#if ver('revs') && tienda.resenas.length}
						<section style="padding:clamp(56px,8cqw,110px) clamp(16px,4cqw,52px) 0">
							<div
								style="display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,340px),1fr));gap:clamp(18px,3cqw,40px);align-items:start"
							>
								<div>
									<h2
										style="margin:0;font-family:'Newsreader',serif;font-weight:400;font-size:clamp(38px,5cqw,64px);line-height:.95;letter-spacing:-.03em"
									>
										Lo que dicen <i>quienes compraron</i>
									</h2>
									{#if tienda.promedio}<div
											style="margin-top:14px;display:flex;align-items:center;gap:8px;font-size:14px"
										>
											<span style="display:flex;color:#2B2520"
												>{#each estrellas(tienda.promedio) as e, i (i)}<Icon
														icon={e}
														size={15}
													/>{/each}</span
											><b>{tienda.promedio.toLocaleString('es-AR')}</b><span style="color:#6F6358"
												>· {tienda.resenas.length}
												{tienda.resenas.length === 1 ? 'reseña' : 'reseñas'}</span
											>
										</div>{/if}
								</div>
								<div style="display:flex;flex-direction:column;gap:12px">
									{#each tienda.resenas.slice(0, 3) as r, i (i)}
										{@const pr = tienda.productos.find((p) => p.nombre === r.producto)}
										<div
											style="display:grid;grid-template-columns:{pr?.foto
												? '120px 1fr'
												: '1fr'};gap:16px;padding-bottom:12px;border-bottom:1px solid rgba(43,37,32,.14)"
										>
											{#if pr?.foto}<div
													style="aspect-ratio:4/5;overflow:hidden;background:#D8CCBC"
												>
													<img
														src={img(pr.foto, 240, 300)}
														alt=""
														style="width:100%;height:100%;object-fit:cover"
													/>
												</div>{/if}
											<div style="min-width:0">
												<p
													style="margin:0;font-family:'Newsreader',serif;font-size:20px;line-height:1.3;font-style:italic"
												>
													“{r.texto}”
												</p>
												<div style="margin-top:10px;font-size:13px;color:#6F6358">
													{r.autor}{#if r.producto}
														· {r.conCompra ? 'compró' : 'sobre'} {r.producto}{/if}
												</div>
												{#if r.conCompra}<div
														style="margin-top:6px;display:flex;align-items:center;gap:4px;font-size:12px;font-weight:600;letter-spacing:.04em;color:var(--acento, #4F5B3A)"
													>
														<Icon icon={I.phSealCheckFill} size={13} />Compra verificada
													</div>{/if}
												{#if r.respuesta}<div
														style="margin-top:10px;padding:12px 14px;background:#E6DED2;font-size:13.5px;line-height:1.55;color:#6F6358"
													>
														<b style="color:#2B2520">Respuesta de {tienda.nombre}</b><br
														/>{r.respuesta}
													</div>{/if}
											</div>
										</div>
									{/each}
								</div>
							</div>
						</section>
					{/if}
				{:else if seccion === 'ig'}
					{#if ver('ig') && tienda.fotos.length >= 3}
						<section style="padding:clamp(56px,8cqw,110px) clamp(16px,4cqw,52px)">
							<h2
								style="margin:0;font-family:'Newsreader',serif;font-weight:400;font-size:clamp(38px,5cqw,64px);line-height:.95"
							>
								En casa
							</h2>
							<div
								style="margin-top:20px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(30%,170px),1fr));gap:6px"
							>
								{#each tienda.fotos as g (g)}<div
										style="aspect-ratio:1;overflow:hidden;background:#D8CCBC"
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
				<div style="height:clamp(56px,8cqw,110px)"></div>
			{/if}
		{/if}

		<!-- ============ CATÁLOGO ============ -->
		{#if pagina === 'catalogo' && !tienda.productos.length}
			<section style="padding:clamp(40px,7cqw,96px) clamp(16px,4cqw,48px)">
				<VacioTienda {tienda} {tema} motivo="sin-productos" />
			</section>
		{:else if pagina === 'catalogo'}
			<section style="padding:28px clamp(16px,4cqw,52px) clamp(56px,8cqw,100px)">
				<div
					style="display:flex;justify-content:space-between;align-items:end;gap:16px;flex-wrap:wrap;padding-bottom:18px;border-bottom:1px solid rgba(43,37,32,.2)"
				>
					<h1
						style="margin:0;font-family:'Newsreader',serif;font-weight:400;font-size:clamp(50px,7.6cqw,110px);line-height:.88;letter-spacing:-.045em;overflow-wrap:anywhere"
					>
						{categoria ?? 'Todo'}
					</h1>
					<span style="font-size:13.5px;color:#6F6358"
						>{listado.length} {listado.length === 1 ? 'pieza' : 'piezas'}</span
					>
				</div>
				<div
					style="margin-top:16px;display:flex;gap:18px;overflow-x:auto;scrollbar-width:none;font-size:14.5px"
				>
					{#each rooms as r (r.t)}{@const on = r.c === categoria}<a
							href={enlaceCat(r.c)}
							style="flex:none;padding-bottom:6px;border-bottom:2px solid {on
								? '#2B2520'
								: 'transparent'};font-weight:{on
								? 600
								: 400};white-space:nowrap;text-decoration:none;color:#2B2520"
							>{r.t} <span style="color:#8C8278;font-weight:400">{r.n}</span></a
						>{/each}
				</div>
				<div
					style="margin-top:16px;display:flex;justify-content:flex-end;align-items:center;gap:10px;flex-wrap:wrap;font-size:13.5px"
				>
					<label style="display:flex;align-items:center;gap:6px;font-weight:600;position:relative"
						><select
							value={orden}
							onchange={(e) => goto(enlaceCat(categoria, e.currentTarget.value), { replace: true })}
							aria-label="Ordenar"
							style="appearance:none;-webkit-appearance:none;border:0;background:transparent;padding:0 20px 0 0;font:600 13.5px 'Familjen Grotesk';color:#2B2520;cursor:pointer"
							>{#each ORDENES as [k, t] (k)}<option value={k}>{t}</option>{/each}</select
						><span style="position:absolute;right:0;pointer-events:none;display:grid"
							><Icon icon={I.phCaretDown} size={13} /></span
						></label
					>
				</div>
				{#if catalog.length}
					<div
						style="margin-top:24px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(46%,250px),1fr));gap:clamp(18px,2.6cqw,32px) clamp(12px,1.8cqw,22px)"
					>
						{#each catalog as p (p.id)}<a
								href={p.href}
								class="zoom"
								style="display:block;text-decoration:none;text-align:left;color:#2B2520;font-family:inherit;min-width:0"
								><span
									style="position:relative;display:block;aspect-ratio:4/5;overflow:hidden;background:#D8CCBC"
									>{#if p.img}<img
											src={p.img}
											alt=""
											style="width:100%;height:100%;object-fit:cover"
										/>{/if}{#if p.tag}<span
											style="position:absolute;top:10px;left:10px;padding:4px 9px;background:#EFE9E1;font-size:11.5px;font-weight:600"
											>{p.tag}</span
										>{/if}</span
								><span
									style="display:flex;justify-content:space-between;align-items:baseline;gap:4px 10px;flex-wrap:wrap;margin-top:12px"
									><span style="font-family:'Newsreader',serif;font-size:21px;line-height:1.1"
										>{p.n}</span
									><b style="font-size:14.5px;white-space:nowrap">{p.price}</b></span
								>{#if p.w || p.cuota}<span
										style="display:block;margin-top:4px;font-size:12.5px;color:#6F6358"
										>{p.w ? (p.cuota ? `${p.w} · ` : p.w) : ''}{#if p.cuota}{tienda.cobros.cuotas} ×
											{p.cuota}{/if}</span
									>{/if}</a
							>{/each}
					</div>
					{#if listado.length > mostrar}
						<div style="margin-top:40px;display:flex;justify-content:center">
							<button
								type="button"
								onclick={() => (mostrar += 24)}
								style="padding:11px 22px;border:1px solid #2B2520;background:transparent;font:600 14px 'Familjen Grotesk';color:#2B2520;cursor:pointer"
								>Ver más</button
							>
						</div>
					{/if}
				{:else}
					<p style="margin:40px 0;font-size:15px;color:#6F6358">
						Todavía no hay piezas en esta categoría.
					</p>
				{/if}
			</section>
		{/if}

		<!-- ============ PRODUCTO ============ -->
		{#if pagina === 'producto' && producto && sel}
			{#if armar}
				<section style="padding:clamp(28px,4cqw,56px) clamp(16px,4cqw,52px) clamp(28px,4cqw,56px)">
					<div style="font-size:12.5px;letter-spacing:.18em;text-transform:uppercase;color:#6F6358">
						<a href={base || '/'} style="color:inherit;text-decoration:none">Inicio</a
						>{#if producto.categoria}&nbsp;·&nbsp;<a
								href={enlaceCat(producto.categoria)}
								style="color:inherit;text-decoration:none">{producto.categoria}</a
							>{/if}
						· Paso a paso
					</div>
					<h1
						style="margin:12px 0 0;font-family:'Newsreader',serif;font-weight:400;font-size:clamp(46px,7.4cqw,108px);line-height:.9;letter-spacing:-.04em;overflow-wrap:anywhere"
					>
						Armá tu <i style="color:var(--acento, #4F5B3A)">{producto.nombre}.</i>
					</h1>
					<div
						style="margin-top:28px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,380px),1fr));gap:clamp(18px,3.4cqw,48px);align-items:start"
					>
						<div
							style="position:{nar
								? 'static'
								: 'sticky'};top:90px;display:flex;flex-direction:column;gap:12px;min-width:0"
						>
							<div
								style="position:relative;aspect-ratio:4/3;background:#E6DED2;overflow:hidden;display:grid;place-items:center"
							>
								{#if sel.foto}<img
										src={img(sel.foto, 1200)}
										alt={producto.nombre}
										style="position:absolute;inset:0;width:100%;height:100%;object-fit:cover"
									/>{/if}
							</div>
							<div
								style="padding:16px;background:#2B2520;color:#EFE9E1;display:flex;justify-content:space-between;align-items:center;gap:12px;flex-wrap:wrap"
							>
								<div style="min-width:0">
									<div style="font-size:12.5px;color:#BDB2A6">
										Tu {producto.nombre.toLowerCase()} · {sel.texto || 'elegí cada paso'}
									</div>
									<div style="font-family:'Newsreader',serif;font-size:32px;line-height:1.1">
										{sel.desde ? 'Desde ' : ''}{fmt(sel.precio)}
									</div>
									<div style="font-size:12.5px;color:#BDB2A6">
										{#if senaArmado}Seña de {fmt(senaArmado)} ahora · el resto al&nbsp;retirarlo{:else if producto.encargo}Seña
											del {producto.encargo}% · el resto al&nbsp;retirarlo{:else if tienda.cobros.cuotas > 1}{tienda
												.cobros.cuotas} × {fmt(sel.precio / tienda.cobros.cuotas)} sin&nbsp;interés{:else}{stockFicha}{/if}
									</div>
								</div>
								<button
									type="button"
									onclick={producto.encargo ? encargar : agregar}
									disabled={sel.stock === 0 || tienda.vitrina}
									style="padding:14px 20px;min-height:48px;border:0;background:#C9B98F;color:#2B2520;font:700 14px 'Familjen Grotesk';cursor:pointer"
									>{producto.encargo && !sel.falta && sel.stock > 0
										? 'Encargar'
										: sel.boton('Agregar al pedido')}</button
								>
							</div>
							<LanzamientoFicha {tienda} {producto} {sel} {tema} bind:abierto={encargoAbierto} />
						</div>
						<div style="display:flex;flex-direction:column;gap:26px;min-width:0">
							{#each sel.opciones as o, i (o.nombre)}{@const swatch = o.valores.every(
									(x) => x.color || x.foto
								)}
								<div>
									<div
										style="display:flex;justify-content:space-between;align-items:baseline;gap:12px;font-size:15px;font-weight:600;color:{sel.aviso &&
										!o.elegido
											? '#8A3A26'
											: '#2B2520'}"
									>
										<span>{i + 1} · {o.nombre}</span>{#if o.elegido}<span
												style="font-family:'Newsreader',serif;font-size:22px;font-weight:400"
												>{o.elegido}</span
											>{/if}
									</div>
									{#if swatch}<div
											style="margin-top:10px;display:grid;grid-template-columns:repeat(auto-fit,minmax(130px,1fr));gap:8px"
										>
											{#each o.valores as x (x.valor)}<button
													type="button"
													aria-pressed={x.elegido}
													disabled={!x.disponible && !x.elegido}
													onclick={() => sel.elegir(o.nombre, x.valor)}
													style="padding:10px;border:1px solid {x.elegido
														? '#2B2520'
														: 'rgba(43,37,32,.18)'};background:{x.elegido
														? '#F7F3EE'
														: 'transparent'};cursor:{x.disponible
														? 'pointer'
														: 'default'};text-align:left;font-family:inherit;color:#2B2520;opacity:{x.disponible
														? 1
														: 0.45}"
													>{#if x.color}<span
															style="display:block;height:44px;background:{x.color};background-image:repeating-linear-gradient(90deg,rgba(0,0,0,.06) 0 2px,transparent 2px 9px)"
														></span>{:else}<img
															src={img(x.foto, 300, 120)}
															alt=""
															style="display:block;width:100%;height:44px;object-fit:cover"
														/>{/if}<span
														style="display:block;margin-top:8px;font-size:14px;font-weight:600"
														>{x.valor}</span
													>{#if !x.disponible}<span
															style="display:block;font-size:12.5px;color:#6F6358">Sin stock</span
														>{/if}</button
												>{/each}
										</div>{:else}<div style="margin-top:10px;display:flex;gap:8px;flex-wrap:wrap">
											{#each o.valores as x (x.valor)}<button
													type="button"
													aria-pressed={x.elegido}
													disabled={!x.disponible && !x.elegido}
													onclick={() => sel.elegir(o.nombre, x.valor)}
													style="padding:10px 14px;min-height:44px;border:1px solid {x.elegido
														? '#2B2520'
														: 'rgba(43,37,32,.18)'};background:{x.elegido
														? '#F7F3EE'
														: 'transparent'};font:400 14px 'Familjen Grotesk';color:{x.disponible
														? '#2B2520'
														: '#6F6358'};text-decoration:{x.disponible
														? 'none'
														: 'line-through'};white-space:nowrap;cursor:{x.disponible
														? 'pointer'
														: 'default'}">{x.valor}</button
												>{/each}
										</div>{/if}
								</div>
							{/each}
							{#if whatsapp}<div
									style="padding:16px;border:1px dashed rgba(43,37,32,.35);display:flex;gap:12px;align-items:center;flex-wrap:wrap"
								>
									<Icon
										icon={I.phPencilRuler}
										size={26}
										style="color:var(--acento, #4F5B3A);flex:none"
									/>
									<div style="flex:1;min-width:200px">
										<div style="font-size:15px;font-weight:600">¿Otra idea?</div>
										<div style="font-size:13.5px;color:#6F6358">
											Mandanos un boceto o una foto por WhatsApp y te lo&nbsp;cotizamos.
										</div>
									</div>
									<a
										href={wa(`¡Hola! Quiero cotizar ${producto.nombre} a medida.`)}
										target="_blank"
										rel="noopener"
										style="display:flex;align-items:center;gap:7px;padding:11px 14px;background:#25D366;color:#fff;font-weight:600;font-size:14px;text-decoration:none"
										><Icon icon={I.phWhatsappLogoFill} size={17} />Enviar boceto</a
									>
								</div>{/if}
						</div>
					</div>
				</section>
			{:else}
				<section
					style="padding:24px clamp(16px,4cqw,52px) 0;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,420px),1fr));gap:clamp(22px,4cqw,64px);align-items:start"
				>
					<div style="display:grid;grid-template-columns:1fr 1fr;gap:8px">
						<div
							style="position:relative;grid-column:1/-1;aspect-ratio:5/4;overflow:hidden;background:#D8CCBC"
						>
							{#if sel.foto}<img
									src={img(sel.foto, 1200)}
									alt={producto.nombre}
									style="width:100%;height:100%;object-fit:cover"
								/>{/if}{#if producto.etiqueta}<span
									style="position:absolute;top:10px;left:10px;padding:4px 9px;background:#EFE9E1;font-size:11.5px;font-weight:600"
									>{producto.etiqueta}</span
								>{/if}
						</div>
						{#if sel.fotos.length > 1}{#each sel.fotos as f, i (f)}{#if i !== sel.fotoIdx}<button
										type="button"
										aria-label="Ver la foto {i + 1}"
										onclick={() => sel.verFoto(i)}
										style="aspect-ratio:5/4;padding:0;border:0;background:#D8CCBC;overflow:hidden;cursor:pointer"
										><img
											src={img(f, 600)}
											alt=""
											style="width:100%;height:100%;object-fit:cover"
										/></button
									>{/if}{/each}{/if}
					</div>
					<div style="position:sticky;top:90px;min-width:0">
						<div
							style="font-size:12.5px;letter-spacing:.16em;text-transform:uppercase;color:#6F6358"
						>
							<a href={base || '/'} style="color:inherit;text-decoration:none">Inicio</a
							>{#if producto.categoria}&nbsp;·&nbsp;<a
									href={enlaceCat(producto.categoria)}
									style="color:inherit;text-decoration:none">{producto.categoria}</a
								>{/if}
							· {stockFicha}
						</div>
						<h1
							style="margin:10px 0 0;font-family:'Newsreader',serif;font-weight:400;font-size:clamp(44px,5.4cqw,72px);line-height:.92;letter-spacing:-.035em;overflow-wrap:anywhere"
						>
							{producto.nombre}
						</h1>
						{#if resenasDe}<div
								style="margin-top:12px;display:flex;align-items:center;gap:8px;font-size:13.5px"
							>
								<span style="display:flex"
									>{#each estrellas(5) as e, i (i)}<Icon icon={e} size={14} />{/each}</span
								><span style="color:#6F6358"
									>{resenasDe} {resenasDe === 1 ? 'reseña' : 'reseñas'}</span
								>
							</div>{/if}
						<div style="margin-top:16px;display:flex;align-items:baseline;gap:12px;flex-wrap:wrap">
							<b style="font-size:28px;font-weight:600"
								>{sel.desde ? 'Desde ' : ''}{fmt(sel.precio)}</b
							>{#if tachado}<s style="color:#8C8278">{fmt(tachado)}</s><span
									style="font-size:13px;color:var(--acento, #4F5B3A);font-weight:600"
									>−{Math.round((1 - sel.precio / tachado) * 100)}%</span
								>{/if}
						</div>
						{#if tienda.cobros.cuotas > 1 || tienda.cobros.transferencia > 0}<div
								style="margin-top:6px;font-size:14px"
							>
								{#if tienda.cobros.cuotas > 1}{tienda.cobros.cuotas} cuotas sin interés de
									<b>{fmt(sel.precio / tienda.cobros.cuotas)}</b
									>{/if}{#if tienda.cobros.transferencia > 0}
									{#if tienda.cobros.cuotas > 1}·{/if}
									<b style="color:var(--acento, #4F5B3A)"
										>{fmt(sel.precio * (1 - tienda.cobros.transferencia / 100))} por&nbsp;transferencia</b
									>{/if}
							</div>{/if}
						{#each sel.opciones as o (o.nombre)}{@const swatch = o.valores.every(
								(x) => x.color || x.foto
							)}
							<div style="margin-top:20px">
								<div
									style="margin-bottom:10px;font-size:14px;color:{sel.aviso && !o.elegido
										? '#8A3A26'
										: '#2B2520'}"
								>
									{o.nombre}{#if o.elegido}: <b style="font-weight:600">{o.elegido}</b>{/if}
								</div>
								{#if swatch}<div style="display:flex;gap:12px;flex-wrap:wrap">
										{#each o.valores as x (x.valor)}<button
												type="button"
												aria-label="{o.nombre} {x.valor}"
												aria-pressed={x.elegido}
												disabled={!x.disponible && !x.elegido}
												onclick={() => sel.elegir(o.nombre, x.valor)}
												style="width:36px;height:36px;padding:0;border:0;border-radius:50%;overflow:hidden;cursor:pointer;opacity:{x.disponible
													? 1
													: 0.4};box-shadow:inset 0 0 0 1px rgba(43,37,32,.2),0 0 0 {x.elegido
													? '3px'
													: '0px'} #EFE9E1,0 0 0 {x.elegido ? '4px' : '0px'} #2B2520"
												>{#if x.color}<span
														style="display:block;width:100%;height:100%;background:{x.color}"
													></span>{:else}<img
														src={img(x.foto, 80, 80)}
														alt=""
														style="width:100%;height:100%;object-fit:cover"
													/>{/if}</button
											>{/each}
									</div>{:else}<div
										style="display:grid;grid-template-columns:repeat(auto-fill,minmax(min(30%,72px),1fr));gap:6px"
									>
										{#each o.valores as x (x.valor)}<button
												type="button"
												aria-pressed={x.elegido}
												disabled={!x.disponible && !x.elegido}
												onclick={() => sel.elegir(o.nombre, x.valor)}
												style="padding:12px 6px;text-align:center;border:1px solid {x.elegido
													? '#2B2520'
													: 'rgba(43,37,32,.2)'};background:{x.elegido
													? '#2B2520'
													: 'transparent'};color:{x.elegido
													? '#EFE9E1'
													: x.disponible
														? '#2B2520'
														: '#6F6358'};font:400 14px 'Familjen Grotesk';text-decoration:{x.disponible
													? 'none'
													: 'line-through'};cursor:{x.disponible ? 'pointer' : 'default'}"
												>{x.valor}</button
											>{/each}
									</div>{/if}
							</div>{/each}
						{#if tienda.envios.retiro && tienda.direccion}<div
								style="margin-top:20px;padding:14px 16px;background:#E6DED2;display:flex;gap:12px;align-items:center"
							>
								<Icon
									icon={I.phCalendarCheck}
									size={24}
									style="color:var(--acento, #4F5B3A);flex:none"
								/>
								<div style="font-size:14px;line-height:1.4">
									También podés <b>retirarlo sin costo</b> en {tienda.direccion}.
								</div>
							</div>{/if}
						<div style="margin-top:16px;display:flex;flex-direction:column;gap:8px">
							<button
								type="button"
								onclick={agregar}
								disabled={sel.stock === 0 || tienda.vitrina}
								style="padding:17px;border:0;background:#2B2520;color:#EFE9E1;font:600 15px 'Familjen Grotesk';letter-spacing:.03em;cursor:pointer"
								>{sel.boton('Agregar al pedido')}</button
							>{#if whatsapp}<a
									href={wa(
										`¡Hola! Quiero comprar ${producto.nombre}${sel.variante ? ` (${sel.texto})` : ''} (${fmt(sel.precio)}).`
									)}
									target="_blank"
									rel="noopener"
									style="display:flex;justify-content:center;align-items:center;gap:8px;padding:15px;border:1px solid #2B2520;font-size:14.5px;font-weight:600;color:#2B2520;text-decoration:none"
									><Icon icon={I.phWhatsappLogo} size={18} />Comprar por WhatsApp</a
								>{/if}
							<LanzamientoFicha {tienda} {producto} {sel} {tema} />
							<CompartirProducto {tienda} {producto} {tema} />
						</div>
						{#if envia(tienda.envios)}<div
								style="margin-top:16px;padding:16px;border:1px solid rgba(43,37,32,.16)"
							>
								<label for="cp-ficha" style="font-size:14px;font-weight:600">Calculá el envío</label
								>
								<div style="margin-top:10px;display:flex;gap:6px">
									<input
										id="cp-ficha"
										bind:value={cp}
										inputmode="numeric"
										maxlength="4"
										placeholder="Código postal"
										style="flex:1;min-width:0;padding:11px 12px;border:0;background:#F7F3EE;font:400 14px 'Familjen Grotesk';color:#2B2520"
										onkeydown={(e) => e.key === 'Enter' && calcular()}
									/><button
										type="button"
										onclick={calcular}
										style="padding:11px 16px;border:0;background:#2B2520;color:#EFE9E1;font:600 13.5px 'Familjen Grotesk';cursor:pointer"
										>Calcular</button
									>
								</div>
								{#if cotizado}<div
										style="margin-top:10px;display:flex;flex-direction:column;gap:6px;font-size:13.5px"
									>
										{#each cotizado as o (o.id)}<span
												style="display:flex;justify-content:space-between;gap:10px"
												><span>{o.nombre} · {o.detalle}</span><b
													style="white-space:nowrap;color:{o.precio
														? '#2B2520'
														: 'var(--acento, #4F5B3A)'}">{o.precio ? fmt(o.precio) : 'Gratis'}</b
												></span
											>{/each}
									</div>{/if}
							</div>{/if}
						<div
							style="margin-top:14px;display:grid;grid-template-columns:repeat(3,1fr);border-top:1px solid rgba(43,37,32,.16)"
						>
							{#each facts as f (f.t)}<div style="padding:14px 4px 0">
									<div style="font-family:'Newsreader',serif;font-size:24px">{f.v}</div>
									<div style="font-size:12.5px;color:#6F6358">{f.t}</div>
								</div>{/each}
						</div>
						<div style="margin-top:18px;border-top:1px solid rgba(43,37,32,.16)">
							{#each acc as a (a.t)}<div
									style="padding:14px 0;border-bottom:1px solid rgba(43,37,32,.16)"
								>
									<button
										type="button"
										onclick={() => (abierto = a.open ? '' : a.t)}
										aria-expanded={a.open}
										style="width:100%;display:flex;justify-content:space-between;align-items:center;padding:0;border:0;background:transparent;font:600 14px 'Familjen Grotesk';color:#2B2520;cursor:pointer;text-align:left"
										>{a.t}<Icon icon={a.i} size={14} /></button
									>{#if a.open}<p
											style="margin:10px 0 0;font-size:14px;line-height:1.6;color:#6F6358"
										>
											{a.d}
										</p>{/if}
								</div>{/each}
						</div>
					</div>
				</section>
			{/if}
			<section style="padding:clamp(48px,7cqw,90px) clamp(16px,4cqw,52px) 0;max-width:860px">
				<PreguntasProducto
					slug={tienda.slug}
					productoId={producto.id}
					preguntas={tienda.preguntas}
					tema={{ ...tema, tarjeta: 'transparent' }}
					titulo="Preguntas sobre este producto"
				/>
			</section>
			{#if related.length}
				<section style="padding:clamp(56px,8cqw,100px) clamp(16px,4cqw,52px)">
					<h2
						style="margin:0;font-family:'Newsreader',serif;font-weight:400;font-size:clamp(34px,4.4cqw,56px);line-height:1"
					>
						Va bien con
					</h2>
					<div
						style="margin-top:22px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(46%,220px),1fr));gap:16px"
					>
						{#each related as p (p.id)}<a
								href={p.href}
								style="display:block;text-decoration:none;text-align:left;color:#2B2520;font-family:inherit;min-width:0"
								><span style="display:block;aspect-ratio:1;overflow:hidden;background:#D8CCBC"
									>{#if p.img}<img
											src={p.img}
											alt=""
											style="width:100%;height:100%;object-fit:cover"
										/>{/if}</span
								><span
									style="display:block;margin-top:10px;font-family:'Newsreader',serif;font-size:19px"
									>{p.n}</span
								><b style="font-size:14.5px">{p.price}</b></a
							>{/each}
					</div>
				</section>
			{:else}
				<div style="height:clamp(56px,8cqw,100px)"></div>
			{/if}
		{/if}

		<!-- ============ CARRITO ============ -->
		{#if pagina === 'carrito'}
			<section style="padding:32px clamp(16px,4cqw,52px) clamp(56px,8cqw,100px)">
				<h1
					style="margin:0;font-family:'Newsreader',serif;font-weight:400;font-size:clamp(46px,6.6cqw,92px);line-height:.92;letter-spacing:-.04em"
				>
					Tu pedido
				</h1>
				{#if !lineas.length}
					<p style="margin:24px 0 0;font-size:15px;color:#6F6358">Todavía no sumaste nada.</p>
					<a
						href="{base}/catalogo"
						style="display:inline-block;margin-top:18px;text-decoration:none;padding:15px 24px;background:#2B2520;color:#EFE9E1;font:600 14px 'Familjen Grotesk';letter-spacing:.04em"
						>Ver todas las piezas</a
					>
				{:else}
					{#if tienda.envios.gratisDesde}
						<div
							style="margin-top:18px;padding:14px 16px;background:#E6DED2;display:flex;gap:12px;align-items:center;flex-wrap:wrap;font-size:14px"
						>
							<span style="flex:1;min-width:200px;display:flex;align-items:center;gap:10px"
								><Icon
									icon={I.phTruck}
									size={22}
									style="color:var(--acento, #4F5B3A);flex:none"
								/>{#if falta}<span>Te faltan <b>{fmt(falta)}</b> para el envío gratis</span
									>{:else}<b>Tenés envío gratis</b>{/if}</span
							><span
								style="flex:1;min-width:160px;height:3px;background:rgba(43,37,32,.15);overflow:hidden"
								><span
									style="display:block;height:100%;width:{Math.min(
										100,
										(subtotal / tienda.envios.gratisDesde) * 100
									)}%;background:var(--acento, #4F5B3A)"
								></span></span
							>
						</div>
					{/if}
					<div
						style="margin-top:24px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,360px),1fr));gap:clamp(22px,4cqw,56px);align-items:start"
					>
						<div style="min-width:0">
							{#each lineas as c (c.clave)}<div
									style="display:flex;gap:16px;padding:18px 0;border-bottom:1px solid rgba(43,37,32,.14)"
								>
									{#if c.producto.foto}<img
											src={img(c.producto.foto, 220, 260)}
											alt=""
											style="width:110px;height:130px;object-fit:cover;flex:none;background:#D8CCBC"
										/>{/if}
									<div style="flex:1;min-width:0;display:flex;flex-direction:column">
										<div style="display:flex;justify-content:space-between;gap:10px">
											<a
												href="{base}/producto/{c.producto.slug ?? c.id}"
												style="font-family:'Newsreader',serif;font-size:23px;line-height:1.05;color:#2B2520;text-decoration:none"
												>{c.producto.nombre}</a
											><b style="white-space:nowrap">{fmt(c.subtotal)}</b>
										</div>
										<span style="font-size:13px;color:#6F6358;margin-top:4px"
											>{#if c.variante}{c.variante} ·
											{/if}{fmt(c.producto.precio)} c/u</span
										>
										<span style="flex:1;min-height:12px"></span>
										<div
											style="display:flex;justify-content:space-between;align-items:center;gap:12px"
										>
											<span
												style="display:flex;align-items:center;border:1px solid rgba(43,37,32,.2)"
												><button
													type="button"
													aria-label="Uno menos"
													onclick={() => carrito.fijar(c.clave, c.cantidad - 1)}
													style="width:34px;height:34px;display:grid;place-items:center;border:0;background:transparent;cursor:pointer;color:#2B2520"
													><Icon icon={I.phMinus} size={13} /></button
												><span style="width:26px;text-align:center;font-size:14px"
													>{c.cantidad}</span
												><button
													type="button"
													aria-label="Uno más"
													disabled={c.cantidad >= c.producto.stock}
													onclick={() => carrito.fijar(c.clave, c.cantidad + 1)}
													style="width:34px;height:34px;display:grid;place-items:center;border:0;background:transparent;cursor:pointer;color:#2B2520"
													><Icon icon={I.phPlus} size={13} /></button
												></span
											><button
												type="button"
												onclick={() => carrito.fijar(c.clave, 0)}
												style="border:0;background:transparent;padding:0;font:400 13px 'Familjen Grotesk';text-decoration:underline;color:#6F6358;cursor:pointer"
												>Quitar</button
											>
										</div>
									</div>
								</div>{/each}
							{#if upsell.length}
								<div style="margin-top:22px;padding:18px;background:#E6DED2">
									<div style="font-size:15px;font-weight:600">Va bien con</div>
									<div style="margin-top:12px;display:flex;flex-direction:column">
										{#each upsell as p (p.id)}<div
												style="display:flex;align-items:center;gap:12px;padding:8px 0"
											>
												{#if p.foto}<img
														src={img(p.foto, 120, 140)}
														alt=""
														style="width:48px;height:56px;object-fit:cover;background:#D8CCBC;flex:none"
													/>{/if}<a
													href="{base}/producto/{p.slug ?? p.id}"
													style="flex:1;min-width:0;font-size:14.5px;color:#2B2520;text-decoration:none"
													>{p.nombre}</a
												><b style="font-size:14px;white-space:nowrap"
													>{tarjetaDe(p).desde ? 'Desde ' : ''}{fmt(tarjetaDe(p).precio)}</b
												><button
													type="button"
													aria-label="Sumar {p.nombre}"
													onclick={() => sumarRapido(p)}
													style="width:30px;height:30px;flex:none;border:1px solid #2B2520;background:transparent;display:grid;place-items:center;cursor:pointer;color:#2B2520"
													><Icon icon={I.phPlus} size={13} /></button
												>
											</div>{/each}
									</div>
								</div>
							{/if}
						</div>
						<div
							style="position:sticky;top:90px;padding:24px;background:#F7F3EE;border:1px solid rgba(43,37,32,.12);display:flex;flex-direction:column;gap:11px;font-size:14px"
						>
							<div style="font-family:'Newsreader',serif;font-size:30px">Resumen</div>
							<span style="display:flex;justify-content:space-between"
								><span>Subtotal</span><span>{fmt(subtotal)}</span></span
							>
							<span style="display:flex;justify-content:space-between;gap:10px"
								><span>Envío</span><span style="color:#6F6358;text-align:right"
									>Se calcula en el próximo paso</span
								></span
							>
							<div style="height:1px;background:rgba(43,37,32,.14)"></div>
							<span style="display:flex;justify-content:space-between;align-items:baseline"
								><b>Total</b><b
									style="font-family:'Newsreader',serif;font-size:32px;font-weight:500"
									>{fmt(subtotal)}</b
								></span
							>
							{#if tienda.cobros.cuotas > 1 || tienda.cobros.transferencia > 0}<div
									style="display:grid;grid-template-columns:{tienda.cobros.cuotas > 1 &&
									tienda.cobros.transferencia > 0
										? '1fr 1fr'
										: '1fr'};gap:6px"
								>
									{#if tienda.cobros.cuotas > 1}<span
											style="padding:10px;background:#fff;border:1px solid #2B2520;text-align:center"
											><b style="display:block"
												>{tienda.cobros.cuotas} × {fmt(subtotal / tienda.cobros.cuotas)}</b
											><span style="font-size:12px;color:#6F6358">sin interés</span></span
										>{/if}{#if tienda.cobros.transferencia > 0}<span
											style="padding:10px;background:#fff;border:1px solid rgba(43,37,32,.16);text-align:center"
											><b style="display:block;color:var(--acento, #4F5B3A)"
												>{fmt(subtotal * (1 - tienda.cobros.transferencia / 100))}</b
											><span style="font-size:12px;color:#6F6358">por transferencia</span></span
										>{/if}
								</div>{/if}
							<a
								href="{base}/checkout"
								style="margin-top:4px;padding:16px;background:#2B2520;color:#EFE9E1;font:600 15px 'Familjen Grotesk';text-align:center;text-decoration:none"
								>Finalizar compra</a
							>
						</div>
					</div>
				{/if}
			</section>
		{/if}

		<!-- ============ CHECKOUT ============ -->
		{#if pagina === 'checkout'}
			{#if !lineas.length}
				<section style="padding:32px clamp(16px,4cqw,52px) clamp(56px,8cqw,100px)">
					<p style="margin:0;font-size:15px;color:#6F6358">Tu pedido está vacío.</p>
					<a href="{base}/catalogo" style="display:inline-block;margin-top:10px;color:#2B2520"
						>Ver todas las piezas</a
					>
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
					style="padding:32px clamp(16px,4cqw,52px) clamp(56px,8cqw,100px);display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,360px),1fr));gap:clamp(22px,4cqw,56px);align-items:start"
				>
					<input type="hidden" name="carrito" value={JSON.stringify(carrito.lineas)} />
					<input type="hidden" name="envio" value={envio ?? ''} />
					<input type="hidden" name="pago" value={pago} />
					<div style="display:flex;flex-direction:column;gap:26px;min-width:0">
						<div
							style="display:flex;gap:8px;font-size:12.5px;letter-spacing:.12em;text-transform:uppercase"
						>
							<b>1 Datos</b><span style="color:#8C8278">— 2 Entrega — 3 Pago</span>
						</div>
						<div>
							<div style="font-family:'Newsreader',serif;font-size:32px">Tus datos</div>
							<div
								style="margin-top:12px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,200px),1fr));gap:8px"
							>
								{#each [['email', 'Email', 'email', 'email'], ['nombre', 'Nombre y apellido', 'text', 'name'], ['telefono', 'Teléfono', 'tel', 'tel'], ['dni', 'DNI', 'text', 'off'], ...camposFactura(tienda)] as [n, l, tipo, ac] (n)}
									<label
										style="display:block;padding:11px 13px;border:1px solid rgba(43,37,32,.2);background:#F7F3EE"
										><span style="display:block;font-size:11.5px;color:#6F6358">{l}</span><input
											name={n}
											type={tipo}
											autocomplete={ac as 'email'}
											required={!OPCIONALES.has(n)}
											maxlength="120"
											style="width:100%;margin-top:2px;padding:0;border:0;background:transparent;font:400 14.5px 'Familjen Grotesk';color:#2B2520;outline:none"
										/></label
									>
								{/each}
							</div>
							<!-- Sin bloque de newsletter en el diseño de esta plantilla: la casilla del checkout (§50). -->
							<label
								style="margin-top:10px;display:flex;align-items:center;gap:10px;min-height:44px;font:400 14.5px 'Familjen Grotesk';color:#2B2520;cursor:pointer"
								><input
									type="checkbox"
									name="novedades"
									style="flex:none;width:18px;height:18px;margin:0;accent-color:var(--acento, #4F5B3A)"
								/>Quiero recibir novedades</label
							>
						</div>
						<div>
							<div style="font-family:'Newsreader',serif;font-size:32px">Entrega</div>
							<div
								style="margin-top:12px;display:grid;grid-template-columns:minmax(0,140px) minmax(0,1fr);gap:8px"
							>
								<label
									style="display:block;padding:11px 13px;border:1px solid rgba(43,37,32,.2);background:#F7F3EE"
									><span style="display:block;font-size:11.5px;color:#6F6358">Código postal</span
									><input
										name="cp"
										bind:value={cp}
										inputmode="numeric"
										maxlength="4"
										style="width:100%;margin-top:2px;padding:0;border:0;background:transparent;font:400 14.5px 'Familjen Grotesk';color:#2B2520;outline:none"
									/></label
								>
								<label
									style="display:block;padding:11px 13px;border:1px solid rgba(43,37,32,.2);background:#F7F3EE"
									><span style="display:block;font-size:11.5px;color:#6F6358"
										>Dirección (calle, número, piso)</span
									><input
										name="direccion"
										maxlength="160"
										autocomplete="street-address"
										required={envio !== 'retiro'}
										style="width:100%;margin-top:2px;padding:0;border:0;background:transparent;font:400 14.5px 'Familjen Grotesk';color:#2B2520;outline:none"
									/></label
								>
							</div>
							<div
								style="margin-top:8px;display:flex;flex-direction:column;gap:8px"
								role="radiogroup"
								aria-label="Forma de entrega"
							>
								{#each opciones as s (s.id)}{@const on = envio === s.id}<button
										type="button"
										role="radio"
										aria-checked={on}
										onclick={() => (envio = s.id)}
										style="display:flex;gap:12px;align-items:center;padding:15px;border:1.5px solid {on
											? '#2B2520'
											: 'rgba(43,37,32,.18)'};background:{on
											? '#F7F3EE'
											: 'transparent'};text-align:left;cursor:pointer;font-family:inherit;color:#2B2520"
										><Icon icon={ICONO_ENVIO[s.id]} size={24} style="flex:none" /><span
											style="flex:1;min-width:0"
											><span style="display:block;font-size:14.5px;font-weight:600">{s.nombre}</span
											><span style="display:block;font-size:12.5px;color:#6F6358">{s.detalle}</span
											></span
										><b
											style="white-space:nowrap;color:{s.precio
												? '#2B2520'
												: 'var(--acento, #4F5B3A)'}">{s.precio ? fmt(s.precio) : 'Gratis'}</b
										></button
									>{:else}<span style="font-size:13.5px;color:#6F6358"
										>Poné tu código postal para ver cómo te llega.</span
									>{/each}
							</div>
							<EnvioCorreo {tienda} {cp} {envio} {opciones} />
						</div>
						<div>
							<div style="font-family:'Newsreader',serif;font-size:32px">Pago</div>
							<div
								style="margin-top:12px;display:flex;flex-direction:column;gap:8px"
								role="radiogroup"
								aria-label="Forma de pago"
							>
								{#each pagos as s (s.id)}{@const on = pago === s.id}<button
										type="button"
										role="radio"
										aria-checked={on}
										onclick={() => (pago = s.id)}
										style="display:flex;align-items:center;gap:12px;padding:15px;border:1.5px solid {on
											? '#2B2520'
											: 'rgba(43,37,32,.18)'};background:{on
											? '#F7F3EE'
											: 'transparent'};text-align:left;cursor:pointer;font-family:inherit;color:#2B2520"
										><span
											style="width:18px;height:18px;border-radius:50%;border:1.5px solid #2B2520;display:grid;place-items:center;flex:none"
											>{#if on}<span
													style="width:9px;height:9px;border-radius:50%;background:#2B2520"
												></span>{/if}</span
										><Icon icon={s.i} size={22} style="color:{s.c};flex:none" /><span
											style="flex:1;min-width:0"
											><span style="display:block;font-size:14.5px;font-weight:600">{s.t}</span
											><span style="display:block;font-size:12.5px;color:#6F6358">{s.d}</span></span
										></button
									>{/each}
							</div>
						</div>
						{#if error}<div
								role="alert"
								style="padding:12px 14px;background:rgba(138,58,38,.12);color:#8A3A26;font-size:14px"
							>
								{error}
							</div>{/if}
						<button
							disabled={!envio || !pagos.length || enviando || tienda.vitrina}
							style="padding:18px;border:0;background:var(--acento, #4F5B3A);color:#fff;font:600 15.5px 'Familjen Grotesk';cursor:pointer;display:flex;justify-content:center;align-items:center;gap:8px"
							><Icon icon={I.phLockSimple} size={16} />{tienda.vitrina
								? 'Es una tienda de ejemplo'
								: !pagos.length
									? 'Esta tienda todavía no tiene medios de pago'
									: envio
										? `Pagar ${fmt(total)}`
										: 'Elegí cómo te llega'}</button
						>
					</div>
					<div style="position:sticky;top:90px;padding:24px;background:#2B2520;color:#EFE9E1">
						<div style="font-family:'Newsreader',serif;font-size:28px">Tu pedido</div>
						<div style="margin-top:14px;display:flex;flex-direction:column;gap:12px">
							{#each lineas as c (c.clave)}<div style="display:flex;gap:12px;align-items:center">
									{#if c.producto.foto}<img
											src={img(c.producto.foto, 120, 140)}
											alt=""
											style="width:56px;height:66px;object-fit:cover;background:#4A3F36;flex:none"
										/>{/if}<span style="flex:1;min-width:0;font-size:14px"
										>{c.producto.nombre}<br /><span style="font-size:12.5px;color:#BDB2A6"
											>{#if c.variante}{c.variante} ·
											{/if}{c.cantidad} × {fmt(c.producto.precio)}</span
										></span
									><b style="font-size:14px;white-space:nowrap">{fmt(c.subtotal)}</b>
								</div>{/each}
						</div>
						<div
							style="margin-top:16px;padding-top:14px;border-top:1px solid rgba(239,233,225,.14);display:flex;flex-direction:column;gap:8px;font-size:14px"
						>
							<span style="display:flex;justify-content:space-between"
								><span>Subtotal</span><span>{fmt(subtotal)}</span></span
							><span style="display:flex;justify-content:space-between"
								><span>Envío</span><span style={elegida && !cuenta.envio ? 'color:#C9D8A8' : ''}
									>{elegida ? (cuenta.envio ? fmt(cuenta.envio) : 'Gratis') : '—'}</span
								></span
							>{#if cuenta.descCupon}<span
									style="display:flex;justify-content:space-between;color:#C9D8A8"
									><span>Cupón {cupon.aplicado}</span><span>− {fmt(cuenta.descCupon)}</span></span
								>{/if}{#if descuento}<span
									style="display:flex;justify-content:space-between;color:#C9D8A8"
									><span>Transferencia −{tienda.cobros.transferencia}%</span><span
										>− {fmt(descuento)}</span
									></span
								>{/if}
							<div style="margin:6px 0 2px">
								<CampoCupon
									{cupon}
									lineas={carrito.lineas}
									tinta="#EFE9E1"
									suave="#BDB2A6"
									borde="rgba(239,233,225,.2)"
									fondo="rgba(239,233,225,.06)"
									acento="#C9D8A8"
									sobreAcento="#2B2520"
									radio="0px"
									fuente="'Familjen Grotesk',sans-serif"
									error="#FFB4AB"
								/>
							</div>
							<span
								style="display:flex;justify-content:space-between;align-items:baseline;margin-top:6px"
								><b>Total</b><b
									style="font-family:'Newsreader',serif;font-size:30px;font-weight:500"
									>{fmt(total)}</b
								></span
							>
						</div>
						<div
							style="margin-top:14px;display:flex;flex-direction:column;gap:8px;font-size:13px;color:#BDB2A6"
						>
							<span style="display:flex;gap:8px"
								><Icon icon={I.phShieldCheck} size={16} style="flex:none" />Compra protegida · 10
								días para arrepentirte</span
							>
						</div>
					</div>
				</form>
			{/if}
		{/if}

		<!-- ============ LISTO (no está en el diseño: con las piezas del checkout) ============ -->
		{#if pagina === 'listo' && pedido}
			<section
				style="padding:clamp(48px,7cqw,96px) clamp(16px,4cqw,52px);max-width:760px;margin:0 auto"
			>
				<div style="font-size:12.5px;letter-spacing:.18em;text-transform:uppercase;color:#6F6358">
					Pedido #{pedido.numero}
				</div>
				<h1
					style="margin:12px 0 0;font-family:'Newsreader',serif;font-weight:400;font-size:clamp(46px,6.6cqw,92px);line-height:.92;letter-spacing:-.04em;overflow-wrap:anywhere"
				>
					¡Gracias, <i style="color:var(--acento, #4F5B3A)">{pedido.nombre}!</i>
				</h1>
				<p style="margin:18px 0 0;font-size:16px;line-height:1.6;color:#6F6358">
					Recibimos tu pedido <b style="color:#2B2520">#{pedido.numero}</b> por
					<b style="color:#2B2520">{fmt(pedido.total)}</b>. Te escribimos por mail con cada novedad.
				</p>
				<SenaPedido {tema} />
				{#if pedido.pago === 'transferencia'}
					<div
						style="margin-top:22px;padding:14px 16px;background:#E6DED2;display:flex;gap:12px;align-items:center;font-size:14.5px;line-height:1.55"
					>
						<Icon icon={I.phBank} size={24} style="color:var(--acento, #4F5B3A);flex:none" />
						<span
							><b>Transferí {fmt(pedido.total)}</b>
							{#if pedido.alias}al {pedido.esCbu ? 'CBU/CVU' : 'alias'}
								<b>{pedido.alias}</b>{:else}con los datos que te mandamos por mail{/if} y mandanos el
							comprobante{#if whatsapp}&nbsp;<a
									href={wa(`¡Hola! Te mando el comprobante del pedido #${pedido.numero}.`)}
									target="_blank"
									rel="noopener"
									style="color:#2B2520">por WhatsApp</a
								>{:else}&nbsp;respondiendo el mail del pedido{/if}. Lo preparamos apenas
							se&nbsp;acredite.</span
						>
					</div>
				{/if}
				<a
					href={base || '/'}
					style="display:inline-block;margin-top:26px;text-decoration:none;padding:15px 24px;background:#2B2520;color:#EFE9E1;font:600 14px 'Familjen Grotesk';letter-spacing:.04em"
					>Volver a la tienda</a
				>
				{#if pedido.estado === 'Entregado'}
					<div style="margin-top:clamp(40px,6cqw,64px)">
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
				style="padding:32px clamp(16px,4cqw,52px) clamp(56px,8cqw,100px);display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,360px),1fr));gap:clamp(22px,4cqw,56px);align-items:start"
			>
				<input type="hidden" name="servicioId" value={reserva.servicioId ?? ''} />
				<input type="hidden" name="profesionalId" value={reserva.profesionalId ?? ''} />
				<input type="hidden" name="inicio" value={reserva.inicio ?? ''} />
				<div style="display:flex;flex-direction:column;gap:26px;min-width:0">
					<div
						style="display:flex;gap:8px;flex-wrap:wrap;font-size:12.5px;letter-spacing:.12em;text-transform:uppercase"
					>
						{#each PASOS_TURNO as t, i (t)}{#if i <= reserva.paso}<b>{i + 1} {t}</b>{:else}<span
									style="color:#8C8278">— {i + 1} {t}</span
								>{/if}{/each}
					</div>
					<h1
						style="margin:0;font-family:'Newsreader',serif;font-weight:400;font-size:clamp(46px,6.6cqw,92px);line-height:.92;letter-spacing:-.04em;overflow-wrap:anywhere"
					>
						Reservá <i style="color:var(--acento, #4F5B3A)">tu turno</i>
					</h1>
					<div>
						<div style="font-family:'Newsreader',serif;font-size:32px">1 · Servicio</div>
						<div
							style="margin-top:12px;display:flex;flex-direction:column;gap:8px"
							role="radiogroup"
							aria-label="Servicio"
						>
							{#each tienda.turnos.servicios as s (s.id)}{@const on =
									reserva.servicioId === s.id}<button
									type="button"
									role="radio"
									aria-checked={on}
									onclick={() => reserva.elegirServicio(s.id)}
									style="display:flex;align-items:center;gap:12px;width:100%;padding:15px;border:1.5px solid {on
										? '#2B2520'
										: 'rgba(43,37,32,.18)'};background:{on
										? '#F7F3EE'
										: 'transparent'};text-align:left;cursor:pointer;font-family:inherit;color:#2B2520"
									><span
										style="width:18px;height:18px;border-radius:50%;border:1.5px solid #2B2520;display:grid;place-items:center;flex:none"
										>{#if on}<span style="width:9px;height:9px;border-radius:50%;background:#2B2520"
											></span>{/if}</span
									><span style="flex:1;min-width:0"
										><span style="display:block;font-size:14.5px;font-weight:600">{s.nombre}</span
										><span style="display:block;font-size:12.5px;color:#6F6358;"
											>{duracion(s.duracion)}</span
										></span
									><b style="white-space:nowrap">{fmt(s.precio)}</b></button
								>{/each}
						</div>
					</div>
					<div>
						<div style="font-family:'Newsreader',serif;font-size:32px">2 · Profesional</div>
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
									style="flex:none;width:110px;padding:10px;border:1.5px solid {on
										? '#2B2520'
										: 'rgba(43,37,32,.18)'};background:{on
										? '#F7F3EE'
										: 'transparent'};text-align:center;cursor:pointer;font-family:inherit;color:#2B2520"
									>{#if p?.foto}<img
											src={img(p.foto, 120, 120)}
											alt=""
											style="width:56px;height:56px;border-radius:50%;object-fit:cover;display:block;margin:0 auto"
										/>{:else}<span
											style="width:56px;height:56px;border-radius:50%;background:#E6DED2;color:var(--acento, #4F5B3A);display:grid;place-items:center;margin:0 auto"
											><Icon icon={p ? I.phUserFill : I.phShuffleBold} size={22} /></span
										>{/if}<span
										style="display:block;margin-top:8px;font-size:13.5px;font-weight:600;overflow:hidden;text-overflow:ellipsis;white-space:nowrap"
										>{p?.nombre ?? 'Cualquiera'}</span
									><span
										style="display:block;font-size:11.5px;color:#6F6358;overflow:hidden;text-overflow:ellipsis;white-space:nowrap"
										>{p ? p.rol.split(' · ')[0] : 'El primero libre'}</span
									></button
								>{/each}
						</div>
					</div>
					<div>
						<div style="display:flex;justify-content:space-between;align-items:center;gap:10px">
							<div style="font-family:'Newsreader',serif;font-size:32px">3 · Día y horario</div>
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
									style="padding:9px 0;text-align:center;cursor:pointer;font-family:inherit;min-width:0;border:1.5px solid {on
										? 'var(--acento, #4F5B3A)'
										: 'rgba(43,37,32,.18)'};background:{on
										? 'var(--acento, #4F5B3A)'
										: 'transparent'};color:{on ? '#fff' : '#2B2520'}"
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
								<p style="margin:0;font-size:14px;color:#6F6358;">
									Elegí un servicio para ver los horarios.
								</p>
							{:else if reserva.horarios && !reserva.grupos.length}
								<p style="margin:0;font-size:14px;color:#6F6358;">
									No quedan horarios este día. Probá con otro día u otro&nbsp;profesional.
								</p>
							{/if}
							{#each reserva.grupos as g (g.k)}<div>
									<div
										style="font-size:12px;font-weight:700;letter-spacing:.08em;text-transform:uppercase;margin-bottom:7px;color:#6F6358;"
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
													: 'line-through'};border:1.5px solid {on
													? 'var(--acento, #4F5B3A)'
													: 'rgba(43,37,32,.18)'};background:{on
													? 'var(--acento, #4F5B3A)'
													: 'transparent'};color:{on ? '#fff' : '#2B2520'}">{h.hora}</button
											>{/each}
									</div>
								</div>{/each}
						</div>
					</div>
					{#if reserva.listo}
						<div>
							<div style="font-family:'Newsreader',serif;font-size:32px">4 · Tus datos</div>
							<div
								style="margin-top:12px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,200px),1fr));gap:8px"
							>
								{#each [['nombre', 'Nombre', 'text', 'name'], ['telefono', 'WhatsApp', 'tel', 'tel'], ['email', 'Email', 'email', 'email']] as [n, l, tipo, ac] (n)}
									<label
										style="display:block;padding:11px 13px;border:1px solid rgba(43,37,32,.2);background:#F7F3EE"
										><span style="display:block;font-size:11.5px;color:#6F6358">{l}</span><input
											name={n}
											type={tipo}
											autocomplete={ac as 'email'}
											required
											maxlength="120"
											style="width:100%;margin-top:2px;padding:0;border:0;background:transparent;font:400 14.5px 'Familjen Grotesk';color:#2B2520;outline:none"
										/></label
									>
								{/each}
							</div>
						</div>
						<div>
							<div style="font-family:'Newsreader',serif;font-size:32px">Pago</div>
							<div style="margin-top:12px;display:flex;flex-direction:column;gap:8px">
								<div
									style="display:flex;align-items:center;gap:12px;width:100%;padding:15px;border:1.5px solid #2B2520;background:#F7F3EE;text-align:left;cursor:pointer;font-family:inherit;color:#2B2520"
								>
									<span
										style="width:18px;height:18px;border-radius:50%;border:1.5px solid #2B2520;display:grid;place-items:center;flex:none"
										><span style="width:9px;height:9px;border-radius:50%;background:#2B2520"
										></span></span
									>{#if reserva.sena}<Icon
											icon={I.siMercadopago}
											size={22}
											style="color:#00B1EA;flex:none"
										/><span style="flex:1;min-width:0"
											><span style="display:block;font-size:14.5px;font-weight:600"
												>Mercado Pago</span
											><span style="display:block;font-size:12.5px;color:#6F6358;"
												>Seña de {fmt(reserva.sena)} para confirmar · el resto, en el local</span
											></span
										>{:else}<Icon icon={I.phStorefrontFill} size={22} style="flex:none" /><span
											style="flex:1;min-width:0"
											><span style="display:block;font-size:14.5px;font-weight:600"
												>Pagás en el local</span
											><span style="display:block;font-size:12.5px;color:#6F6358;"
												>No hace falta seña: el turno queda confirmado</span
											></span
										>{/if}
								</div>
							</div>
						</div>
					{/if}
				</div>
				<div
					style="position:sticky;top:90px;padding:24px;background:#2B2520;color:#EFE9E1;min-width:0"
				>
					<div style="font-family:'Newsreader',serif;font-size:28px">Tu reserva</div>
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
							<div style="font-size:13px;color:#BDB2A6;">
								{reserva.cuando ?? 'Elegí el día y el horario'}
							</div>
						</div>
					</div>
					{#if reserva.servicio}
						<div
							style="margin-top:16px;padding-top:14px;border-top:1px solid rgba(239,233,225,.14);display:flex;flex-direction:column;gap:8px;font-size:14px"
						>
							<span style="display:flex;justify-content:space-between;gap:10px"
								><span>{reserva.servicio.nombre}</span><span style="white-space:nowrap"
									>{fmt(reserva.servicio.precio)}</span
								></span
							><span style="display:flex;justify-content:space-between;color:#BDB2A6;"
								><span>Duración total</span><span>{reserva.duracion}</span></span
							><span
								style="display:flex;justify-content:space-between;align-items:baseline;margin-top:6px"
								><b>Total</b><b
									style="font-family:'Newsreader',serif;font-size:30px;font-weight:500"
									>{fmt(reserva.servicio.precio)}</b
								></span
							>
						</div>
					{/if}
					<p style="margin:14px 0 0;font-size:13px;line-height:1.5;color:#BDB2A6;">
						{#if reserva.sena}Seña de {fmt(reserva.sena)} para confirmar. Podés cancelar desde el link
							del turno hasta {anticipacion} antes; la seña te la devuelve la&nbsp;tienda.{:else}No
							pagás nada ahora. Podés cancelar desde el link del turno hasta {anticipacion}&nbsp;antes.{/if}
					</p>
					{#if error}<div
							role="alert"
							style="margin-top:12px;padding:12px 14px;background:rgba(138,58,38,.32);color:#F6D5CB;font-size:14px"
						>
							{error}
						</div>{/if}
					<button
						disabled={!reserva.listo || reservando || tienda.vitrina}
						style="margin-top:16px;width:100%;padding:18px;border:0;background:var(--acento, #4F5B3A);color:#fff;font:600 15.5px 'Familjen Grotesk';cursor:pointer"
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
				style="padding:clamp(48px,7cqw,96px) clamp(16px,4cqw,52px);max-width:760px;margin:0 auto"
			>
				<div style="font-size:12.5px;letter-spacing:.18em;text-transform:uppercase;color:#6F6358">
					Turno
				</div>
				<h1
					style="margin:12px 0 0;font-family:'Newsreader',serif;font-weight:400;font-size:clamp(46px,6.6cqw,92px);line-height:.92;letter-spacing:-.04em;overflow-wrap:anywhere"
				>
					{cancelado
						? 'Turno cancelado'
						: turno.estado === 'reservado'
							? `Falta la seña, ${turno.nombre}`
							: `¡Listo, ${turno.nombre}!`}
				</h1>
				<p style="margin:18px 0 0;font-size:16px;line-height:1.6;color:#6F6358">
					{#if cancelado}Liberamos el horario{turno.senaPagada
							? '. La seña te la devuelve la tienda por Mercado Pago'
							: ''}. Si querés, reservá&nbsp;otro.{:else if turno.estado === 'reservado'}Te
						guardamos el horario unos minutos mientras pagás la seña con Mercado&nbsp;Pago.{:else}Tu
						turno está confirmado. Te esperamos en {turno.tienda}.{/if}
				</p>
				{#if aviso === 'pendiente' || aviso === 'error'}<div
						role="status"
						style="margin-top:14px;padding:12px 14px;background:rgba(138,58,38,.12);color:#8A3A26;font-size:14px"
					>
						{aviso === 'pendiente'
							? 'Mercado Pago todavía no acreditó la seña: te avisamos cuando entre.'
							: 'El pago de la seña no salió. Podés reservar de nuevo.'}
					</div>{/if}
				<div
					style="margin-top:22px;padding:14px 16px;background:#E6DED2;font-size:14.5px;line-height:1.55"
				>
					<div
						style="font-size:16px;font-weight:700;text-decoration:{cancelado
							? 'line-through'
							: 'none'}"
					>
						{cuandoLargo(turno.inicio)} con {turno.profesional}
					</div>
					<div style="margin-top:4px;font-size:14px;color:#6F6358;">
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
						style="margin-top:12px;padding:12px 14px;background:rgba(138,58,38,.12);color:#8A3A26;font-size:14px"
					>
						{error}
					</div>{/if}
				<div style="margin-top:22px;display:flex;gap:10px;flex-wrap:wrap;">
					{#if !cancelado}<a
							href="{base}/turno/{turno.codigo}/turno.ics"
							download="turno.ics"
							data-sveltekit-reload
							style="display:flex;align-items:center;gap:8px;padding:15px 24px;border:0;background:var(--acento, #4F5B3A);color:#fff;font:600 14px 'Familjen Grotesk';letter-spacing:.04em;text-decoration:none;cursor:pointer"
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
									style="display:inline-block;padding:15px 24px;border:0;background:#2B2520;color:#EFE9E1;font:600 14px 'Familjen Grotesk';letter-spacing:.04em;text-decoration:none;cursor:pointer"
									>Sí, cancelar el turno</button
								>{:else}<button
									type="button"
									onclick={() => (seguro = true)}
									style="display:inline-block;padding:15px 24px;border:0;background:#2B2520;color:#EFE9E1;font:600 14px 'Familjen Grotesk';letter-spacing:.04em;text-decoration:none;cursor:pointer"
									>Cancelar turno</button
								>{/if}
						</form>
					{:else if cancelado}<a
							href="{base}/turnos"
							style="padding:15px 24px;border:0;background:var(--acento, #4F5B3A);color:#fff;font:600 14px 'Familjen Grotesk';letter-spacing:.04em;text-decoration:none;cursor:pointer"
							>Reservar otro</a
						>{/if}
					<a
						href={base || '/'}
						style="display:inline-block;padding:15px 24px;border:0;background:#2B2520;color:#EFE9E1;font:600 14px 'Familjen Grotesk';letter-spacing:.04em;text-decoration:none;cursor:pointer"
						>Volver a la tienda</a
					>
				</div>
				{#if !cancelado && !turno.cancelable}<p
						style="margin:14px 0 0;font-size:13.5px;color:#6F6358;"
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
			<section
				style="padding:clamp(36px,5cqw,72px) clamp(16px,4cqw,52px) 0;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,360px),1fr));gap:clamp(20px,4cqw,56px);align-items:end"
			>
				<h1
					style="margin:0;font-family:'Newsreader',serif;font-weight:400;font-size:clamp(48px,7.4cqw,110px);line-height:.88;letter-spacing:-.045em;overflow-wrap:anywhere"
				>
					{titulo[1] ? `${titulo[0]} ` : titulo[0]}{#if titulo[1]}
						<i style="color:var(--acento, #4F5B3A)">{titulo[1]}</i>{/if}
				</h1>
				<p style="margin:0;font-size:17px;line-height:1.6;color:#6F6358;max-width:460px">
					{tienda.nombre}{tienda.ciudad ? `, desde ${tienda.ciudad}` : ''}. {comoLlega(
						tienda.envios
					)}
				</p>
			</section>
			{#if tienda.portada.foto}
				<section
					style="padding:clamp(28px,4cqw,56px) clamp(16px,4cqw,52px) 0;display:grid;grid-template-columns:2fr 1fr;gap:8px"
				>
					<div style="aspect-ratio:16/10;overflow:hidden;background:#D8CCBC">
						<img
							src={img(tienda.portada.foto, 1400)}
							alt=""
							style="width:100%;height:100%;object-fit:cover"
						/>
					</div>
					<div style="display:grid;grid-template-rows:1fr 1fr;gap:8px;min-width:0">
						<div style="position:relative;overflow:hidden;background:#D8CCBC">
							{#if fotoNos}<img
									src={img(fotoNos, 700)}
									alt=""
									style="position:absolute;inset:0;width:100%;height:100%;object-fit:cover"
								/>{/if}
						</div>
						<div
							style="background:var(--acento, #4F5B3A);color:#EFE9E1;padding:18px;display:flex;flex-direction:column;justify-content:flex-end;min-width:0"
						>
							<div
								style="font-family:'Newsreader',serif;font-size:clamp(34px,4.4cqw,56px);line-height:1"
							>
								{tienda.productos.length}
							</div>
							<div style="font-size:13.5px;opacity:.85">
								{tienda.productos.length === 1 ? 'pieza en la tienda' : 'piezas en la tienda'}
							</div>
						</div>
					</div>
				</section>
			{/if}
			<section
				style="padding:clamp(48px,7cqw,90px) clamp(16px,4cqw,52px);display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,300px),1fr));gap:clamp(18px,3cqw,40px)"
			>
				<div>
					<div style="font-family:'Newsreader',serif;font-size:32px">
						{tienda.envios.retiro ? 'Vení al local' : 'Cómo comprar'}
					</div>
					<div style="margin-top:14px;display:flex;flex-direction:column;gap:12px;font-size:15px">
						{#if tienda.direccion}<span style="display:flex;gap:12px"
								><Icon icon={I.phMapPin} size={19} style="flex:none" />{tienda.direccion}</span
							>{/if}{#if tienda.cobros.cuotas > 1}<span style="display:flex;gap:12px"
								><Icon icon={I.phCreditCard} size={19} style="flex:none" />{tienda.cobros.cuotas} cuotas
								sin interés{#if tienda.cobros.transferencia > 0}&nbsp;· {tienda.cobros
										.transferencia}% off por transferencia{/if}</span
							>{:else if tienda.cobros.transferencia > 0}<span style="display:flex;gap:12px"
								><Icon icon={I.phBank} size={19} style="flex:none" />{tienda.cobros.transferencia}%
								off por transferencia</span
							>{/if}{#if tienda.envios.gratisDesde}<span style="display:flex;gap:12px"
								><Icon icon={I.phTruck} size={19} style="flex:none" />Envío gratis desde {fmt(
									tienda.envios.gratisDesde
								)}</span
							>{/if}
					</div>
				</div>
				{#if whatsapp}
					<div style="display:flex;flex-direction:column;gap:8px">
						<div style="font-family:'Newsreader',serif;font-size:32px">Escribinos</div>
						<a
							href={wa('¡Hola!')}
							target="_blank"
							rel="noopener"
							style="display:flex;align-items:center;gap:10px;padding:15px;background:#25D366;color:#fff;font-weight:600;text-decoration:none"
							><Icon icon={I.phWhatsappLogoFill} size={22} />{tienda.whatsapp}</a
						>
					</div>
				{/if}
			</section>
		{/if}

		<!-- footer -->
		<footer
			style="padding:clamp(44px,6cqw,80px) clamp(16px,4cqw,52px) 26px;background:#2B2520;color:#EFE9E1"
		>
			<div
				style="font-family:'Newsreader',serif;font-style:italic;font-size:clamp(36px,{pieNombre}cqw,200px);line-height:.8;letter-spacing:-.05em;color:#3E352E;overflow-wrap:break-word"
			>
				{tienda.nombre}
			</div>
			<div
				style="margin-top:28px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,180px),1fr));gap:26px"
			>
				<div style="min-width:0;max-width:340px">
					<p style="margin:0;font-size:14px;color:#BDB2A6;line-height:1.55">
						{#if tienda.cobros.cuotas > 1}{tienda.cobros.cuotas} cuotas sin interés{#if tienda.cobros.transferencia > 0}&nbsp;y
								{tienda.cobros.transferencia}% off por transferencia{/if}, en toda la&nbsp;tienda.{:else if tienda.cobros.transferencia > 0}{tienda
								.cobros.transferencia}% off por transferencia, en toda la&nbsp;tienda.{:else}Compra
							protegida, con 10 días para&nbsp;arrepentirte.{/if}
					</p>
				</div>
				{#each foot as f (f.t)}<div
						style="display:flex;flex-direction:column;gap:9px;font-size:14px"
					>
						<b
							style="font-size:12px;letter-spacing:.14em;text-transform:uppercase;color:#C9B98F;margin-bottom:4px"
							>{f.t}</b
						>{#each f.l as l (l.t)}<a href={l.href} style="color:#DCD3C8;text-decoration:none"
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
				style="margin-top:32px;padding-top:14px;border-top:1px solid rgba(239,233,225,.1);display:flex;justify-content:space-between;gap:14px;flex-wrap:wrap;font-size:12.5px;color:#8C8278"
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
	.quebracho section :global(a:not([style*='background']):hover) {
		color: var(--acento, #4f5b3a);
	}
	.zoom :global(img) {
		transition: transform 0.9s cubic-bezier(0.2, 0.7, 0.1, 1);
	}
	.zoom:hover :global(img) {
		transform: scale(1.04);
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
</style>
