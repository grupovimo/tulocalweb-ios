<script lang="ts">
	import { camposFactura, OPCIONALES } from '#lib/tienda/factura.ts';
	import { envia } from '../envio.ts';
	import { enhance } from '$app/forms';
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
	import { comoLlega, cotizar, cotizarYa, cpValido, textoEnvios } from '../envio.svelte.ts';
	import { Seleccion, tarjetaDe } from '../seleccion.svelte.ts';
	import { partirTitulo } from '../titulo.ts';
	import { page } from '$app/state';
	import { untrack } from 'svelte';
	import { Reserva } from '../reserva.svelte.ts';
	import { cuandoLargo, duracion } from '#lib/turnos/formato.ts';
	import type { OpcionEnvio, ProductoPublico, TiendaPublica } from '../tipos.ts';
	import { AltaNewsletter } from '../newsletter.svelte.ts';

	/**
	 * Plantilla Ceibo (design/Plantilla Ceibo.dc.html), con los datos de la tienda en lugar de los de
	 * ejemplo. Lo que el diseño muestra sin datos atrás (talles, colores, "Armá tu look", contadores de
	 * Instagram) no se inventa: esas partes no se dibujan.
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
	const footSpan = $derived(nar ? 'auto' : 'span 2');

	const fmt = (n: number) => '$ ' + Math.round(n).toLocaleString('es-AR');
	const img = (f: string | null, w: number, h?: number) =>
		!f
			? null
			: /^https?:\/\//.test(f)
				? f
				: `https://images.unsplash.com/photo-${f}?auto=format&fit=crop&w=${w}${h ? '&h=' + h : ''}&q=72`;
	const porId = $derived(new Map(tienda.productos.map((p) => [p.id, p])));
	// Con variantes de precios distintos, "Desde"; "Sin stock" sólo si no queda ninguna.
	const tarjeta = (p: ProductoPublico) => {
		const t = tarjetaDe(p);
		return {
			id: p.id,
			n: p.nombre,
			price: (t.desde ? 'Desde ' : '') + fmt(t.precio),
			old: p.anterior && !t.desde ? fmt(p.anterior) : null,
			img: img(p.foto, 700, 875),
			tag: t.agotado ? 'Sin stock' : p.etiqueta,
			tagBg: p.anterior ? 'var(--acento, #9C4A2F)' : '#1C1A17',
			cuotas:
				tienda.cobros.cuotas > 1
					? `${tienda.cobros.cuotas} × ${fmt(t.precio / tienda.cobros.cuotas)} sin interés`
					: '',
			transf: fmt(t.precio * (1 - tienda.cobros.transferencia / 100)),
			href: `${base}/producto/${p.slug ?? p.id}`
		};
	};
	const ver = (s: string) => tienda.secciones.includes(s as TiendaPublica['secciones'][number]);
	const whatsapp = $derived(tienda.whatsapp.replace(/\D/g, ''));
	const wa = (texto: string) =>
		`https://wa.me/${whatsapp.startsWith('54') ? whatsapp : '54' + whatsapp}?text=${encodeURIComponent(texto)}`;

	// Título de la portada: el último tramo va en itálica, como en el diseño ("Abrigo / que dura.").
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
	const cats = $derived(tienda.categorias.slice(0, 3));
	const best = $derived(
		tienda.masVendidos.slice(0, 4).flatMap((id) => (porId.get(id) ? [tarjeta(porId.get(id)!)] : []))
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
						(p) => p.id !== producto!.id && p.categoria === producto!.categoria && p.stock > 0
					)
					.slice(0, 4)
					.map(tarjeta)
			: []
	);
	const resenasDe = $derived(
		producto ? tienda.resenas.filter((r) => r.producto === producto!.nombre).length : 0
	);
	// Talle, color, presentación…: lo elegido, la variante, su precio, stock y foto (seleccion.svelte.ts).
	const sel = $derived(producto ? new Seleccion(producto) : null);
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
	const lineas = $derived(carrito.detalle(tienda.productos));
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
						c: '#2F6B47'
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
		{ t: 'Todo', href: `${base}/catalogo`, on: pagina === 'catalogo' && !categoria },
		...tienda.categorias.slice(0, 4).map((c) => ({
			t: c.nombre,
			href: enlaceCat(c.nombre),
			on: pagina === 'catalogo' && categoria === c.nombre
		})),
		{ t: 'Nosotros', href: `${base}/nosotros`, on: pagina === 'nosotros' }
	]);
	// Reseñas y preguntas con las piezas de Ceibo: serif para los títulos, esquinas de 2 px.
	const tema: TemaOpiniones = {
		texto: '#1C1A17',
		suave: '#6E655C',
		campo: '#F4EFEA',
		tarjeta: '#EDE6DE',
		borde: 'rgba(28,26,23,.14)',
		acento: '#1C1A17',
		sobreAcento: '#F4EFEA',
		radio: '2px',
		fuente: "'Instrument Sans', sans-serif",
		titulos: "'Instrument Serif', serif",
		pesoTitulos: 400
	};
	const estrellas = (n: number) =>
		Array.from({ length: 5 }, (_, i) =>
			n >= i + 1 ? I.phStarFill : n >= i + 0.5 ? I.phStarHalfFill : I.phStar
		);
	// ---- Turnos (la reserva sale de lib/tienda/reserva.svelte.ts; acá sólo se dibuja) ----
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

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
<div
	class="ceibo"
	style="min-height:100vh;font-family:'Instrument Sans',system-ui,sans-serif;color:#1C1A17"
>
	<div style="container-type:inline-size;width:100%;background:#F4EFEA;position:relative">
		<!-- barra promo -->
		<div
			style="display:flex;justify-content:center;align-items:center;gap:14px;flex-wrap:wrap;padding:9px 16px;background:#1C1A17;color:#F4EFEA;font-size:12.5px;letter-spacing:.02em;text-align:center"
		>
			{#if conteo.actual}<span>{conteo.texto}</span><span
					style="display:flex;gap:4px;font-variant-numeric:tabular-nums"
					>{#each conteo.partes(3) as c (c.u)}<span
							style="padding:2px 6px;border-radius:2px;background:rgba(244,239,234,.14)"
							>{c.v}<span style="opacity:.6;margin-left:2px">{c.u}</span></span
						>{/each}</span
				>{:else}<span
					>{tienda.envios.gratisDesde
						? `Envío gratis desde ${fmt(tienda.envios.gratisDesde)}`
						: tienda.cobros.cuotas > 1
							? `${tienda.cobros.cuotas} cuotas sin interés`
							: 'Compra protegida'}</span
				>{/if}{#if wide && tienda.cobros.cuotas > 1}<span style="opacity:.7"
					>{tienda.cobros.cuotas} cuotas sin interés{#if tienda.cobros.transferencia > 0}&nbsp;· {tienda
							.cobros.transferencia}% off por transferencia{/if}</span
				>{/if}
		</div>

		<!-- header -->
		<header
			style="position:sticky;top:0;z-index:20;display:flex;align-items:center;gap:20px;padding:16px clamp(16px,4cqw,48px);background:rgba(244,239,234,.92);backdrop-filter:blur(10px);border-bottom:1px solid rgba(28,26,23,.1)"
		>
			{#if nar}<a href="{base}/catalogo" aria-label="Catálogo" style="display:grid"
					><Icon icon={I.phList} size={22} /></a
				>{/if}
			<a
				href={base || '/'}
				style="text-decoration:none;font-family:'Instrument Serif',serif;font-size:28px;letter-spacing:-.01em;color:#1C1A17;line-height:1"
				><LogoTienda {tienda} /></a
			>
			{#if wide}<nav style="display:flex;gap:26px;margin-left:24px;font-size:14px">
					{#each nav as n (n.t)}<a
							href={n.href}
							style="text-decoration:none;padding:4px 0;font:500 14px 'Instrument Sans';color:#1C1A17;border-bottom:1px solid {n.on
								? '#1C1A17'
								: 'transparent'}">{n.t}</a
						>{/each}
				</nav>{/if}
			<span style="flex:1"></span>
			{#if nar && tienda.turnos}<a
					href="{base}/turnos"
					aria-label="Reservar turno"
					style="display:grid;color:#1C1A17"><Icon icon={I.phCalendarPlusBold} size={22} /></a
				>{/if}
			<a
				href="{base}/carrito"
				aria-label="Carrito"
				style="position:relative;color:#1C1A17;display:grid"
				><Icon icon={I.phHandbag} size={22} />{#if carrito.unidades}<span
						style="position:absolute;top:-6px;right:-8px;min-width:17px;height:17px;border-radius:999px;background:#1C1A17;color:#F4EFEA;font-size:10.5px;font-weight:700;display:grid;place-items:center"
						>{carrito.unidades}</span
					>{/if}</a
			>
		</header>

		{#if tienda.vitrina}
			<div style="padding:10px 16px;background:#E5DDD3;font-size:13px;text-align:center">
				Es una tienda de ejemplo de la plantilla Ceibo: podés recorrerla, pero no se compra de
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
							style="position:relative;height:clamp(520px,62cqw,820px);overflow:hidden;background:#CFC4B8"
						>
							{#if tienda.portada.foto}<img
									src={img(tienda.portada.foto, 1800)}
									alt=""
									style="position:absolute;inset:0;width:100%;height:100%;object-fit:cover;object-position:50% 25%"
								/>{/if}
							<div
								style="position:absolute;inset:0;background:linear-gradient(90deg,rgba(28,26,23,.55),rgba(28,26,23,.05) 65%)"
							></div>
							<div
								style="position:absolute;left:clamp(20px,5cqw,72px);bottom:clamp(32px,6cqw,88px);right:20px;max-width:640px;color:#F4EFEA"
							>
								{#if tienda.portada.lema}<div
										style="font-size:12.5px;letter-spacing:.24em;text-transform:uppercase"
									>
										{tienda.portada.lema}
									</div>{/if}
								<h1
									style="margin:14px 0 0;font-family:'Instrument Serif',serif;font-weight:400;font-size:clamp(54px,10cqw,132px);line-height:.9;letter-spacing:-.02em"
								>
									{titulo[0]}{#if titulo[1]}<br /><i>{titulo[1]}</i>{/if}
								</h1>
								<div style="margin-top:24px;display:flex;gap:10px;flex-wrap:wrap">
									<a
										href="{base}/catalogo"
										style="text-decoration:none;padding:15px 28px;border-radius:2px;background:#F4EFEA;color:#1C1A17;font:600 14px 'Instrument Sans';letter-spacing:.04em"
										>Ver la colección</a
									>{#if tienda.productos.some((p) => p.anterior)}<a
											href="{base}/catalogo"
											style="text-decoration:none;padding:15px 22px;border:1px solid rgba(244,239,234,.6);border-radius:2px;color:#F4EFEA;font:600 14px 'Instrument Sans';letter-spacing:.04em"
											>Ofertas</a
										>{/if}
								</div>
							</div>
						</section>
					{/if}

					<div
						style="display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));border-bottom:1px solid rgba(28,26,23,.1)"
					>
						{#each perks as p (p.t)}<div
								style="display:flex;align-items:center;gap:12px;padding:18px clamp(16px,3cqw,32px);border-right:1px solid rgba(28,26,23,.08)"
							>
								<Icon icon={p.i} size={22} />
								<div>
									<div style="font-size:13.5px;font-weight:600">{p.t}</div>
									<div style="font-size:12.5px;color:#6E655C">{p.d}</div>
								</div>
							</div>{/each}
					</div>
				{:else if seccion === 'cats'}
					{#if ver('cats') && cats.length}
						<section style="padding:clamp(48px,7cqw,96px) clamp(16px,4cqw,48px) 0">
							<div
								style="display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,240px),1fr));gap:clamp(10px,1.4cqw,18px)"
							>
								{#each cats as c (c.nombre)}
									<a
										href={enlaceCat(c.nombre)}
										class="zoom"
										style="position:relative;display:block;aspect-ratio:3/4;border-radius:2px;overflow:hidden;background:#D8CEC3;text-align:left"
									>
										{#if c.foto}<img
												src={img(c.foto, 800, 1067)}
												alt=""
												style="position:absolute;inset:0;width:100%;height:100%;object-fit:cover"
											/>{/if}
										<span
											style="position:absolute;inset:auto 0 0 0;padding:22px;background:linear-gradient(0deg,rgba(28,26,23,.6),transparent);color:#F4EFEA;display:flex;justify-content:space-between;align-items:end;gap:10px"
											><span
												style="font-family:'Instrument Serif',serif;font-size:36px;line-height:1"
												>{c.nombre}</span
											><span
												style="font-size:12.5px;letter-spacing:.1em;text-transform:uppercase;white-space:nowrap"
												>{c.cantidad} {c.cantidad === 1 ? 'producto' : 'productos'}</span
											></span
										>
									</a>
								{/each}
							</div>
						</section>
					{/if}
				{:else if seccion === 'prods'}
					{#if ver('prods') && best.length}
						<section style="padding:clamp(56px,8cqw,110px) clamp(16px,4cqw,48px) 0">
							<div
								style="display:flex;justify-content:space-between;align-items:end;gap:20px;flex-wrap:wrap"
							>
								<h2
									style="margin:0;font-family:'Instrument Serif',serif;font-weight:400;font-size:clamp(38px,5cqw,64px);line-height:1;letter-spacing:-.015em"
								>
									Lo más elegido
								</h2>
								<a
									href="{base}/catalogo"
									style="text-decoration:none;font:600 13.5px 'Instrument Sans';letter-spacing:.06em;text-transform:uppercase;color:#1C1A17;border-bottom:1px solid #1C1A17"
									>Ver todo</a
								>
							</div>
							<div
								style="margin-top:28px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(46%,240px),1fr));gap:clamp(12px,1.6cqw,22px) clamp(10px,1.4cqw,18px)"
							>
								{#each best as p (p.id)}
									<a
										href={p.href}
										class="zoom"
										style="display:block;text-decoration:none;text-align:left;color:#1C1A17;font-family:inherit"
									>
										<span
											style="position:relative;display:block;aspect-ratio:4/5;border-radius:2px;overflow:hidden;background:#E2D9CF"
											>{#if p.img}<img
													src={p.img}
													alt=""
													style="width:100%;height:100%;object-fit:cover;display:block"
												/>{/if}{#if p.tag}<span
													style="position:absolute;top:10px;left:10px;padding:4px 9px;border-radius:2px;background:{p.tagBg};color:#F4EFEA;font-size:11px;font-weight:600;letter-spacing:.06em;text-transform:uppercase"
													>{p.tag}</span
												>{/if}</span
										>
										<span style="display:block;margin-top:12px;font-size:14.5px;font-weight:500"
											>{p.n}</span
										>
										<span style="display:flex;gap:8px;align-items:baseline;margin-top:3px"
											><b style="font-size:15px;font-weight:600">{p.price}</b>{#if p.old}<s
													style="font-size:13px;color:#8C8278">{p.old}</s
												>{/if}</span
										>
										{#if p.cuotas}<span
												style="display:block;margin-top:2px;font-size:12.5px;color:#6E655C"
												>{p.cuotas}</span
											>{/if}
										{#if tienda.cobros.transferencia > 0}
											<span style="display:block;font-size:12.5px;color:#2F6B47;font-weight:600"
												>{p.transf} por transferencia</span
											>
										{/if}
									</a>
								{/each}
							</div>
						</section>
					{/if}
				{:else if seccion === 'revs'}
					{#if ver('revs') && tienda.resenas.length}
						<section style="padding:clamp(56px,8cqw,110px) clamp(16px,4cqw,48px) 0">
							<div
								style="display:flex;justify-content:space-between;align-items:end;gap:20px;flex-wrap:wrap"
							>
								<h2
									style="margin:0;font-family:'Instrument Serif',serif;font-weight:400;font-size:clamp(38px,5cqw,64px);line-height:1"
								>
									Lo que dicen
								</h2>
								{#if tienda.promedio}<span
										style="display:flex;align-items:center;gap:8px;font-size:14px"
										><span style="display:flex;color:#1C1A17"
											>{#each estrellas(tienda.promedio) as e, i (i)}<Icon
													icon={e}
													size={15}
												/>{/each}</span
										><b>{tienda.promedio.toLocaleString('es-AR')}</b></span
									>{/if}
							</div>
							<div
								style="margin-top:28px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,280px),1fr));gap:16px"
							>
								{#each tienda.resenas.slice(0, 3) as r, i (i)}
									{@const pr = tienda.productos.find((p) => p.nombre === r.producto)}
									<div
										style="display:flex;gap:14px;padding:16px;background:#EDE6DE;border-radius:2px"
									>
										{#if pr?.foto}<img
												src={img(pr.foto, 200, 250)}
												alt=""
												style="width:96px;height:120px;object-fit:cover;border-radius:2px;flex:none;background:#D8CEC3"
											/>{/if}
										<div style="min-width:0">
											<div style="display:flex;gap:1px">
												{#each estrellas(r.estrellas) as e, j (j)}<Icon icon={e} size={13} />{/each}
											</div>
											<p
												style="margin:8px 0 0;font-family:'Instrument Serif',serif;font-size:19px;line-height:1.25"
											>
												“{r.texto}”
											</p>
											<div style="margin-top:8px;font-size:12.5px;color:#6E655C">
												{r.autor}{#if r.producto}
													· {r.conCompra ? 'compró' : 'sobre'} {r.producto}{/if}
											</div>
											{#if r.conCompra}<div
													style="margin-top:6px;display:flex;align-items:center;gap:4px;font-size:12px;font-weight:600;color:#2F6B47"
												>
													<Icon icon={I.phSealCheckFill} size={13} />Compra verificada
												</div>{/if}
											{#if r.respuesta}<div
													style="margin-top:10px;padding-left:10px;border-left:2px solid #1C1A17;font-size:13px;line-height:1.5;color:#4A433C"
												>
													<b style="color:#1C1A17">Respuesta de {tienda.nombre}</b><br
													/>{r.respuesta}
												</div>{/if}
										</div>
									</div>
								{/each}
							</div>
						</section>
					{/if}
				{:else if seccion === 'ig'}
					{#if ver('ig') && tienda.fotos.length >= 3}
						<section style="padding:clamp(56px,8cqw,110px) clamp(16px,4cqw,48px)">
							<h2
								style="margin:0;font-family:'Instrument Serif',serif;font-weight:400;font-size:clamp(38px,5cqw,64px);line-height:1"
							>
								Usado de verdad
							</h2>
							<div
								style="margin-top:24px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(30%,160px),1fr));gap:6px"
							>
								{#each tienda.fotos as g (g)}<div
										style="position:relative;aspect-ratio:1;overflow:hidden;background:#D8CEC3"
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
			<section style="padding:28px clamp(16px,4cqw,48px) clamp(56px,8cqw,96px)">
				<div style="font-size:12.5px;color:#6E655C">
					<a href={base || '/'} style="color:inherit;text-decoration:none">Inicio</a> /
					<span style="color:#1C1A17">{categoria ?? 'Todo'}</span>
				</div>
				<div
					style="margin-top:14px;display:flex;justify-content:space-between;align-items:end;gap:16px;flex-wrap:wrap"
				>
					<h1
						style="margin:0;font-family:'Instrument Serif',serif;font-weight:400;font-size:clamp(46px,7cqw,96px);line-height:.95"
					>
						{categoria ?? 'Todo'}
					</h1>
					<span style="font-size:13.5px;color:#6E655C"
						>{listado.length} {listado.length === 1 ? 'producto' : 'productos'}</span
					>
				</div>
				<div
					style="margin-top:22px;display:flex;gap:8px;overflow-x:auto;scrollbar-width:none;padding-bottom:4px"
				>
					{#each [null, ...tienda.categorias.map((c) => c.nombre)] as c (c ?? '')}{@const on =
							c === categoria}<a
							href={enlaceCat(c)}
							style="flex:none;text-decoration:none;padding:9px 16px;border:1px solid {on
								? '#1C1A17'
								: 'rgba(28,26,23,.2)'};background:{on ? '#1C1A17' : 'transparent'};color:{on
								? '#F4EFEA'
								: '#1C1A17'};border-radius:999px;font-size:13px;font-weight:500;white-space:nowrap"
							>{c ?? 'Todo'}</a
						>{/each}
				</div>
				<div style="margin-top:22px">
					<div
						style="display:flex;justify-content:flex-end;align-items:center;gap:10px;padding-bottom:14px;border-bottom:1px solid rgba(28,26,23,.1);font-size:13.5px"
					>
						<label style="display:flex;align-items:center;gap:6px;font-weight:600"
							>Ordenar<select
								value={orden}
								onchange={(e) =>
									goto(enlaceCat(categoria, e.currentTarget.value), { replace: true })}
								style="border:0;background:transparent;font:600 13.5px 'Instrument Sans';color:#1C1A17;cursor:pointer"
								>{#each ORDENES as [k, t] (k)}<option value={k}>{t}</option>{/each}</select
							></label
						>
					</div>
					{#if catalog.length}
						<div
							style="margin-top:22px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(46%,220px),1fr));gap:clamp(14px,2cqw,26px) clamp(10px,1.4cqw,18px)"
						>
							{#each catalog as p (p.id)}
								<a
									href={p.href}
									style="display:block;text-decoration:none;text-align:left;color:#1C1A17;font-family:inherit"
								>
									<span
										style="position:relative;display:block;aspect-ratio:4/5;border-radius:2px;overflow:hidden;background:#E2D9CF"
										>{#if p.img}<img
												src={p.img}
												alt=""
												style="width:100%;height:100%;object-fit:cover;display:block"
											/>{/if}{#if p.tag}<span
												style="position:absolute;top:10px;left:10px;padding:4px 9px;border-radius:2px;background:{p.tagBg};color:#F4EFEA;font-size:11px;font-weight:600;letter-spacing:.06em;text-transform:uppercase"
												>{p.tag}</span
											>{/if}</span
									>
									<span style="display:block;margin-top:12px;font-size:14.5px;font-weight:500"
										>{p.n}</span
									>
									<span style="display:flex;gap:8px;align-items:baseline;margin-top:3px"
										><b style="font-size:15px;font-weight:600">{p.price}</b>{#if p.old}<s
												style="font-size:13px;color:#8C8278">{p.old}</s
											>{/if}</span
									>
									{#if p.cuotas}<span
											style="display:block;margin-top:2px;font-size:12.5px;color:#6E655C"
											>{p.cuotas}</span
										>{/if}
								</a>
							{/each}
						</div>
						{#if listado.length > mostrar}
							<div style="margin-top:40px;display:flex;justify-content:center">
								<button
									type="button"
									onclick={() => (mostrar += 24)}
									style="padding:0 22px;height:42px;border:1px solid rgba(28,26,23,.2);border-radius:2px;background:transparent;font:600 14px 'Instrument Sans';color:#1C1A17;cursor:pointer"
									>Ver más</button
								>
							</div>
						{/if}
					{:else}
						<p style="margin:40px 0;font-size:15px;color:#6E655C">
							Todavía no hay productos en esta categoría.
						</p>
					{/if}
				</div>
			</section>
		{/if}

		<!-- ============ PRODUCTO ============ -->
		{#if pagina === 'producto' && producto && sel}
			<section style="padding:24px clamp(16px,4cqw,48px) 0">
				<div style="font-size:12.5px;color:#6E655C">
					<a href={base || '/'} style="color:inherit;text-decoration:none">Inicio</a> / {#if producto.categoria}<a
							href={enlaceCat(producto.categoria)}
							style="color:inherit;text-decoration:none">{producto.categoria}</a
						> /
					{/if}<span style="color:#1C1A17">{producto.nombre}</span>
				</div>
				<div
					style="margin-top:18px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,400px),1fr));gap:clamp(24px,4cqw,64px);align-items:start"
				>
					<div style="display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:8px">
						<div
							style="grid-column:1/-1;position:relative;aspect-ratio:4/5;background:#E2D9CF;border-radius:2px;overflow:hidden"
						>
							{#if sel.foto}<img
									src={img(sel.foto, 1100)}
									alt={producto.nombre}
									style="width:100%;height:100%;object-fit:cover"
								/>{/if}{#if sel.anterior || producto.etiqueta}<span
									style="position:absolute;top:14px;left:14px;padding:5px 10px;border-radius:2px;background:var(--acento, #9C4A2F);color:#F4EFEA;font-size:11.5px;font-weight:600;letter-spacing:.06em;text-transform:uppercase"
									>{sel.anterior
										? `−${Math.round((1 - sel.precio / sel.anterior) * 100)}%`
										: producto.etiqueta}</span
								>{/if}
						</div>
						{#if sel.fotos.length > 1}{#each sel.fotos as f, i (f)}{#if i !== sel.fotoIdx}<button
										type="button"
										aria-label="Ver la foto {i + 1}"
										onclick={() => sel.verFoto(i)}
										style="aspect-ratio:4/5;padding:0;border:0;background:#E2D9CF;border-radius:2px;overflow:hidden;cursor:pointer"
										><img
											src={img(f, 600)}
											alt=""
											style="width:100%;height:100%;object-fit:cover"
										/></button
									>{/if}{/each}{/if}
					</div>
					<div style="position:sticky;top:90px">
						{#if producto.categoria}<div
								style="font-size:12.5px;letter-spacing:.2em;text-transform:uppercase;color:#6E655C"
							>
								{producto.categoria}
							</div>{/if}
						<h1
							style="margin:10px 0 0;font-family:'Instrument Serif',serif;font-weight:400;font-size:clamp(40px,4.6cqw,60px);line-height:1"
						>
							{producto.nombre}
						</h1>
						{#if resenasDe}<div
								style="margin-top:10px;display:flex;align-items:center;gap:8px;font-size:13px"
							>
								<span style="display:flex"
									>{#each estrellas(5) as e, i (i)}<Icon icon={e} size={13} />{/each}</span
								><span>{resenasDe} {resenasDe === 1 ? 'reseña' : 'reseñas'}</span>
							</div>{/if}
						<div style="margin-top:20px;display:flex;align-items:baseline;gap:12px">
							<b style="font-size:28px;font-weight:600;letter-spacing:-.01em"
								>{sel.desde ? 'Desde ' : ''}{fmt(sel.precio)}</b
							>{#if sel.anterior}<s style="font-size:17px;color:#8C8278">{fmt(sel.anterior)}</s
								>{/if}
						</div>
						<div style="margin-top:8px;display:flex;flex-direction:column;gap:4px;font-size:14px">
							{#if tienda.cobros.cuotas > 1}<span style="display:flex;align-items:center;gap:8px"
									><Icon icon={I.phCreditCard} size={17} />{tienda.cobros.cuotas} cuotas sin interés de
									<b>{fmt(sel.precio / tienda.cobros.cuotas)}</b></span
								>{/if}{#if tienda.cobros.transferencia > 0}<span
									style="display:flex;align-items:center;gap:8px;color:#2F6B47;font-weight:600"
									><Icon icon={I.phBank} size={17} />{fmt(
										sel.precio * (1 - tienda.cobros.transferencia / 100)
									)} por transferencia ({tienda.cobros.transferencia}% off)</span
								>{/if}
						</div>
						{#each sel.opciones as o (o.nombre)}{@const swatch = o.valores.every(
								(x) => x.color || x.foto
							)}
							<div style="margin-top:22px">
								<div
									style="font-size:13px;margin-bottom:10px;color:{sel.aviso && !o.elegido
										? 'var(--acento, #9C4A2F)'
										: '#1C1A17'}"
								>
									{o.nombre}{#if o.elegido}: <b>{o.elegido}</b>{/if}
								</div>
								{#if swatch}<div style="display:flex;gap:10px;flex-wrap:wrap">
										{#each o.valores as x (x.valor)}<button
												type="button"
												aria-label="{o.nombre} {x.valor}"
												aria-pressed={x.elegido}
												disabled={!x.disponible && !x.elegido}
												onclick={() => sel.elegir(o.nombre, x.valor)}
												style="width:34px;height:34px;padding:0;border:0;border-radius:50%;overflow:hidden;cursor:pointer;opacity:{x.disponible
													? 1
													: 0.4};box-shadow:inset 0 0 0 1px rgba(28,26,23,.2),0 0 0 {x.elegido
													? '3px'
													: '0px'} #F4EFEA,0 0 0 {x.elegido ? '4.5px' : '0px'} #1C1A17"
												>{#if x.color}<span
														style="display:block;width:100%;height:100%;background:{x.color}"
													></span>{:else}<img
														src={img(x.foto, 80, 80)}
														alt=""
														style="width:100%;height:100%;object-fit:cover"
													/>{/if}</button
											>{/each}
									</div>{:else}<div
										style="display:grid;grid-template-columns:repeat(auto-fill,minmax(min(30%,64px),1fr));gap:6px"
									>
										{#each o.valores as x (x.valor)}<button
												type="button"
												aria-pressed={x.elegido}
												disabled={!x.disponible && !x.elegido}
												onclick={() => sel.elegir(o.nombre, x.valor)}
												style="padding:12px 4px;text-align:center;border:1px solid {x.elegido
													? '#1C1A17'
													: 'rgba(28,26,23,.2)'};background:{x.elegido
													? '#1C1A17'
													: 'transparent'};color:{x.elegido
													? '#F4EFEA'
													: x.disponible
														? '#1C1A17'
														: '#B3A99E'};border-radius:2px;font:400 14px 'Instrument Sans';text-decoration:{x.disponible
													? 'none'
													: 'line-through'};cursor:{x.disponible ? 'pointer' : 'default'}"
												>{x.valor}</button
											>{/each}
									</div>{/if}
							</div>{/each}
						{#if sel.stock > 0 && sel.stock <= 3 && (!sel.conOpciones || sel.variante)}<div
								style="margin-top:16px;font-size:12.5px;color:var(--acento, #9C4A2F)"
							>
								{sel.stock === 1 ? 'Queda 1' : `Quedan ${sel.stock}`}{#if sel.variante}
									en {sel.texto.toLowerCase()}{/if}
							</div>{/if}
						<div style="margin-top:22px;display:flex;flex-direction:column;gap:8px">
							<button
								type="button"
								onclick={agregar}
								disabled={sel.stock === 0 || tienda.vitrina}
								style="padding:17px;border:0;border-radius:2px;background:#1C1A17;color:#F4EFEA;font:600 14.5px 'Instrument Sans';letter-spacing:.04em;cursor:pointer"
								>{sel.boton('Agregar al carrito')}</button
							>{#if whatsapp}<a
									href={wa(
										`¡Hola! Quiero comprar ${producto.nombre}${sel.variante ? ` (${sel.texto})` : ''} (${fmt(sel.precio)}).`
									)}
									target="_blank"
									rel="noopener"
									style="display:flex;justify-content:center;align-items:center;gap:8px;padding:15px;border:1px solid #1C1A17;border-radius:2px;font-size:14.5px;font-weight:600;color:#1C1A17;text-decoration:none"
									><Icon icon={I.phWhatsappLogo} size={18} />Comprar por WhatsApp</a
								>{/if}
							<LanzamientoFicha {tienda} {producto} {sel} {tema} />
							<CompartirProducto {tienda} {producto} {tema} />
						</div>
						{#if envia(tienda.envios)}<div
								style="margin-top:20px;padding:16px;background:#EDE6DE;border-radius:2px"
							>
								<label for="cp-ficha" style="font-size:13.5px;font-weight:600"
									>Calculá el envío</label
								>
								<div style="margin-top:10px;display:flex;gap:6px">
									<input
										id="cp-ficha"
										bind:value={cp}
										inputmode="numeric"
										maxlength="4"
										placeholder="Código postal"
										style="flex:1;min-width:0;padding:11px 12px;background:#F4EFEA;border:0;border-radius:2px;font:400 14px 'Instrument Sans';color:#1C1A17"
										onkeydown={(e) => e.key === 'Enter' && calcular()}
									/><button
										type="button"
										onclick={calcular}
										style="padding:11px 16px;border:0;background:#1C1A17;color:#F4EFEA;border-radius:2px;font:600 13.5px 'Instrument Sans';cursor:pointer"
										>Calcular</button
									>
								</div>
								{#if cotizado}<div
										style="margin-top:12px;display:flex;flex-direction:column;gap:8px;font-size:13.5px"
									>
										{#each cotizado as o (o.id)}<span
												style="display:flex;justify-content:space-between;gap:10px"
												><span>{o.nombre} · {o.detalle}</span><b
													style="white-space:nowrap;color:{o.precio ? '#1C1A17' : '#2F6B47'}"
													>{o.precio ? fmt(o.precio) : 'Gratis'}</b
												></span
											>{/each}
									</div>{/if}
							</div>{/if}
						<div style="margin-top:14px;border-top:1px solid rgba(28,26,23,.14)">
							{#each acc as a (a.t)}<div
									style="padding:16px 0;border-bottom:1px solid rgba(28,26,23,.14)"
								>
									<button
										type="button"
										onclick={() => (abierto = a.open ? '' : a.t)}
										aria-expanded={a.open}
										style="width:100%;display:flex;justify-content:space-between;padding:0;border:0;background:transparent;font:600 14px 'Instrument Sans';color:#1C1A17;cursor:pointer;text-align:left"
										>{a.t}<Icon icon={a.i} size={15} /></button
									>{#if a.open}<p
											style="margin:10px 0 0;font-size:14px;line-height:1.6;color:#4A433C"
										>
											{a.d}
										</p>{/if}
								</div>{/each}
						</div>
					</div>
				</div>
			</section>
			<section style="padding:clamp(48px,7cqw,90px) clamp(16px,4cqw,48px) 0;max-width:820px">
				<PreguntasProducto
					slug={tienda.slug}
					productoId={producto.id}
					preguntas={tienda.preguntas}
					tema={{ ...tema, tarjeta: 'transparent' }}
					titulo="Preguntas sobre este producto"
				/>
			</section>
			{#if related.length}
				<section style="padding:clamp(56px,8cqw,100px) clamp(16px,4cqw,48px)">
					<h2
						style="margin:0;font-family:'Instrument Serif',serif;font-weight:400;font-size:clamp(34px,4.4cqw,54px);line-height:1"
					>
						También te puede gustar
					</h2>
					<div
						style="margin-top:24px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(46%,220px),1fr));gap:16px 14px"
					>
						{#each related as p (p.id)}<a
								href={p.href}
								style="display:block;text-decoration:none;text-align:left;color:#1C1A17;font-family:inherit"
								><span
									style="display:block;aspect-ratio:4/5;border-radius:2px;overflow:hidden;background:#E2D9CF"
									>{#if p.img}<img
											src={p.img}
											alt=""
											style="width:100%;height:100%;object-fit:cover"
										/>{/if}</span
								><span style="display:block;margin-top:10px;font-size:14px;font-weight:500"
									>{p.n}</span
								><b style="font-size:14.5px;font-weight:600">{p.price}</b></a
							>{/each}
					</div>
				</section>
			{:else}
				<div style="height:clamp(56px,8cqw,100px)"></div>
			{/if}
		{/if}

		<!-- ============ CARRITO ============ -->
		{#if pagina === 'carrito'}
			<section style="padding:32px clamp(16px,4cqw,48px) clamp(56px,8cqw,96px)">
				<h1
					style="margin:0;font-family:'Instrument Serif',serif;font-weight:400;font-size:clamp(46px,6cqw,80px);line-height:1"
				>
					Tu carrito <span style="font-size:.5em;color:#6E655C">({carrito.unidades})</span>
				</h1>
				{#if !lineas.length}
					<p style="margin:24px 0 0;font-size:15px;color:#6E655C">Todavía no sumaste nada.</p>
					<a
						href="{base}/catalogo"
						style="display:inline-block;margin-top:18px;text-decoration:none;padding:15px 26px;border-radius:2px;background:#1C1A17;color:#F4EFEA;font:600 14px 'Instrument Sans';letter-spacing:.04em"
						>Ver la colección</a
					>
				{:else}
					{#if tienda.envios.gratisDesde}
						<div
							style="margin-top:14px;padding:12px 16px;background:#E5DDD3;border-radius:2px;font-size:13.5px;display:flex;align-items:center;gap:12px;flex-wrap:wrap"
						>
							<span style="flex:1;min-width:200px"
								>{#if falta}Te faltan <b>{fmt(falta)}</b> para el envío gratis{:else}<b
										>Tenés envío gratis</b
									>{/if}</span
							><span
								style="flex:1;min-width:160px;height:4px;background:rgba(28,26,23,.15);border-radius:2px;overflow:hidden"
								><span
									style="display:block;width:{Math.min(
										100,
										(subtotal / tienda.envios.gratisDesde) * 100
									)}%;height:100%;background:#1C1A17"
								></span></span
							>
						</div>
					{/if}
					<div
						style="margin-top:24px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,360px),1fr));gap:clamp(24px,4cqw,56px);align-items:start"
					>
						<div style="display:flex;flex-direction:column">
							{#each lineas as c (c.clave)}
								<div
									style="display:flex;gap:16px;padding:18px 0;border-bottom:1px solid rgba(28,26,23,.12)"
								>
									{#if c.producto.foto}<img
											src={img(c.producto.foto, 240, 300)}
											alt=""
											style="width:96px;height:120px;object-fit:cover;border-radius:2px;flex:none;background:#E2D9CF"
										/>{/if}
									<div style="flex:1;min-width:0;display:flex;flex-direction:column">
										<div style="display:flex;justify-content:space-between;gap:10px">
											<a
												href="{base}/producto/{c.producto.slug ?? c.id}"
												style="font-size:15px;font-weight:500;color:#1C1A17;text-decoration:none"
												>{c.producto.nombre}</a
											><b style="font-size:15px;font-weight:600;white-space:nowrap"
												>{fmt(c.subtotal)}</b
											>
										</div>
										<span style="font-size:13px;color:#6E655C;margin-top:4px"
											>{#if c.variante}{c.variante} ·
											{/if}{fmt(c.producto.precio)} c/u</span
										><span style="flex:1;min-height:12px"></span>
										<div style="display:flex;justify-content:space-between;align-items:center">
											<span
												style="display:flex;align-items:center;border:1px solid rgba(28,26,23,.2);border-radius:2px"
												><button
													type="button"
													aria-label="Uno menos"
													onclick={() => carrito.fijar(c.clave, c.cantidad - 1)}
													style="width:34px;height:34px;display:grid;place-items:center;border:0;background:transparent;cursor:pointer;color:#1C1A17"
													><Icon icon={I.phMinus} size={13} /></button
												><span style="width:26px;text-align:center;font-size:14px"
													>{c.cantidad}</span
												><button
													type="button"
													aria-label="Uno más"
													disabled={c.cantidad >= c.producto.stock}
													onclick={() => carrito.fijar(c.clave, c.cantidad + 1)}
													style="width:34px;height:34px;display:grid;place-items:center;border:0;background:transparent;cursor:pointer;color:#1C1A17"
													><Icon icon={I.phPlus} size={13} /></button
												></span
											><button
												type="button"
												onclick={() => carrito.fijar(c.clave, 0)}
												style="border:0;background:transparent;padding:0;font:400 13px 'Instrument Sans';text-decoration:underline;color:#6E655C;cursor:pointer"
												>Quitar</button
											>
										</div>
									</div>
								</div>
							{/each}
							{#if upsell.length}
								<div style="margin-top:28px">
									<div
										style="font-size:13px;font-weight:700;letter-spacing:.14em;text-transform:uppercase"
									>
										Te puede gustar
									</div>
									<div
										style="margin-top:14px;display:flex;gap:10px;overflow-x:auto;scrollbar-width:none"
									>
										{#each upsell as p (p.id)}<div style="flex:none;width:150px">
												<a
													href="{base}/producto/{p.slug ?? p.id}"
													style="display:block;aspect-ratio:4/5;border-radius:2px;overflow:hidden;background:#E2D9CF"
													>{#if p.foto}<img
															src={img(p.foto, 300, 375)}
															alt=""
															style="width:100%;height:100%;object-fit:cover"
														/>{/if}</a
												>
												<div style="margin-top:8px;font-size:13.5px">{p.nombre}</div>
												<div
													style="display:flex;justify-content:space-between;align-items:center;margin-top:2px"
												>
													<b style="font-size:14px;font-weight:600">{fmt(p.precio)}</b><button
														type="button"
														aria-label="Sumar {p.nombre}"
														onclick={() => sumarRapido(p)}
														style="width:28px;height:28px;border-radius:50%;border:1px solid #1C1A17;background:transparent;display:grid;place-items:center;cursor:pointer;color:#1C1A17"
														><Icon icon={I.phPlus} size={13} /></button
													>
												</div>
											</div>{/each}
									</div>
								</div>
							{/if}
						</div>
						<div
							style="position:sticky;top:90px;padding:24px;background:#EDE6DE;border-radius:2px;display:flex;flex-direction:column;gap:12px;font-size:14px"
						>
							<div style="font-family:'Instrument Serif',serif;font-size:30px">Resumen</div>
							<span style="display:flex;justify-content:space-between"
								><span>Subtotal</span><span>{fmt(subtotal)}</span></span
							>
							<span style="display:flex;justify-content:space-between;color:#6E655C"
								><span>Envío</span><span>Se calcula en el próximo paso</span></span
							>
							<div style="height:1px;background:rgba(28,26,23,.14)"></div>
							<span style="display:flex;justify-content:space-between;align-items:baseline"
								><b style="font-size:16px">Total</b><b style="font-size:24px;font-weight:600"
									>{fmt(subtotal)}</b
								></span
							>
							{#if tienda.cobros.cuotas > 1}<span style="font-size:13px;color:#6E655C"
									>{tienda.cobros.cuotas} cuotas sin interés de {fmt(
										subtotal / tienda.cobros.cuotas
									)}</span
								>{/if}
							{#if tienda.cobros.transferencia > 0}
								<span style="font-size:13.5px;color:#2F6B47;font-weight:600"
									>{fmt(subtotal * (1 - tienda.cobros.transferencia / 100))} pagando por transferencia</span
								>
							{/if}
							<a
								href="{base}/checkout"
								style="margin-top:6px;padding:17px;border-radius:2px;background:#1C1A17;color:#F4EFEA;font:600 14.5px 'Instrument Sans';letter-spacing:.04em;text-align:center;text-decoration:none"
								>Finalizar compra</a
							>
							<span style="display:flex;justify-content:center;gap:10px;color:#6E655C"
								>{#if tienda.cobros.mercadoPago}<Icon icon={I.logosVisa} size={30} /><Icon
										icon={I.logosMastercard}
										size={22}
									/><Icon icon={I.siMercadopago} size={20} style="color:#00B1EA" />{/if}</span
							>
						</div>
					</div>
				{/if}
			</section>
		{/if}

		<!-- ============ CHECKOUT ============ -->
		{#if pagina === 'checkout'}
			{#if !lineas.length}
				<section style="padding:32px clamp(16px,4cqw,48px) clamp(56px,8cqw,96px)">
					<p style="font-size:15px;color:#6E655C">Tu carrito está vacío.</p>
					<a href="{base}/catalogo" style="color:#1C1A17">Ver la colección</a>
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
					style="padding:32px clamp(16px,4cqw,48px) clamp(56px,8cqw,96px);display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,360px),1fr));gap:clamp(24px,4cqw,56px);align-items:start"
				>
					<input type="hidden" name="carrito" value={JSON.stringify(carrito.lineas)} />
					<input type="hidden" name="envio" value={envio ?? ''} />
					<input type="hidden" name="pago" value={pago} />
					<div style="display:flex;flex-direction:column;gap:28px">
						<div
							style="display:flex;gap:8px;font-size:12.5px;letter-spacing:.1em;text-transform:uppercase"
						>
							<span style="font-weight:700">1 Datos</span><span style="color:#8C8278"
								>— 2 Entrega — 3 Pago</span
							>
						</div>
						<div>
							<div style="font-family:'Instrument Serif',serif;font-size:32px">Tus datos</div>
							<div
								style="margin-top:14px;display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));gap:10px"
							>
								{#each [['email', 'Email', 'email', 'email'], ['nombre', 'Nombre y apellido', 'text', 'name'], ['telefono', 'Teléfono', 'tel', 'tel'], ['dni', 'DNI', 'text', 'off'], ...camposFactura(tienda)] as [n, l, tipo, ac] (n)}
									<label
										style="display:block;padding:10px 12px;border:1px solid rgba(28,26,23,.22);border-radius:2px;background:#F8F4EF"
										><span style="display:block;font-size:11.5px;color:#6E655C">{l}</span><input
											name={n}
											type={tipo}
											autocomplete={ac as 'email'}
											required={!OPCIONALES.has(n)}
											maxlength="120"
											style="width:100%;margin-top:2px;padding:0;border:0;background:transparent;font:400 14.5px 'Instrument Sans';color:#1C1A17;outline:none"
										/></label
									>
								{/each}
							</div>
						</div>
						<div>
							<div style="font-family:'Instrument Serif',serif;font-size:32px">Entrega</div>
							<div
								style="margin-top:14px;display:grid;grid-template-columns:minmax(0,140px) minmax(0,1fr);gap:10px"
							>
								<label
									style="display:block;padding:10px 12px;border:1px solid rgba(28,26,23,.22);border-radius:2px;background:#F8F4EF"
									><span style="display:block;font-size:11.5px;color:#6E655C">Código postal</span
									><input
										name="cp"
										bind:value={cp}
										inputmode="numeric"
										maxlength="4"
										style="width:100%;margin-top:2px;padding:0;border:0;background:transparent;font:400 14.5px 'Instrument Sans';color:#1C1A17;outline:none"
									/></label
								>
								<label
									style="display:block;padding:10px 12px;border:1px solid rgba(28,26,23,.22);border-radius:2px;background:#F8F4EF"
									><span style="display:block;font-size:11.5px;color:#6E655C"
										>Dirección (calle, número, piso)</span
									><input
										name="direccion"
										maxlength="160"
										autocomplete="street-address"
										required={envio !== 'retiro'}
										style="width:100%;margin-top:2px;padding:0;border:0;background:transparent;font:400 14.5px 'Instrument Sans';color:#1C1A17;outline:none"
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
										style="display:flex;align-items:center;gap:12px;padding:15px;border:1.5px solid {on
											? '#1C1A17'
											: 'rgba(28,26,23,.18)'};border-radius:2px;background:{on
											? '#F8F4EF'
											: 'transparent'};text-align:left;cursor:pointer;font-family:inherit;color:#1C1A17"
										><span
											style="width:18px;height:18px;border-radius:50%;border:1.5px solid #1C1A17;display:grid;place-items:center;flex:none"
											>{#if on}<span
													style="width:9px;height:9px;border-radius:50%;background:#1C1A17"
												></span>{/if}</span
										><span style="flex:1"
											><span style="display:block;font-size:14.5px;font-weight:500">{s.nombre}</span
											><span style="display:block;font-size:12.5px;color:#6E655C">{s.detalle}</span
											></span
										><b
											style="font-size:14px;font-weight:600;color:{s.precio
												? '#1C1A17'
												: '#2F6B47'}">{s.precio ? fmt(s.precio) : 'Gratis'}</b
										></button
									>{:else}<span style="font-size:13.5px;color:#6E655C"
										>Poné tu código postal para ver cómo te llega.</span
									>{/each}
							</div>
							<EnvioCorreo {tienda} {cp} {envio} {opciones} />
						</div>
						<div>
							<div style="font-family:'Instrument Serif',serif;font-size:32px">Pago</div>
							<div
								style="margin-top:14px;display:flex;flex-direction:column;gap:8px"
								role="radiogroup"
								aria-label="Forma de pago"
							>
								{#each pagos as s (s.id)}{@const on = pago === s.id}<button
										type="button"
										role="radio"
										aria-checked={on}
										onclick={() => (pago = s.id)}
										style="display:flex;align-items:center;gap:12px;padding:15px;border:1.5px solid {on
											? '#1C1A17'
											: 'rgba(28,26,23,.18)'};border-radius:2px;background:{on
											? '#F8F4EF'
											: 'transparent'};text-align:left;cursor:pointer;font-family:inherit;color:#1C1A17"
										><span
											style="width:18px;height:18px;border-radius:50%;border:1.5px solid #1C1A17;display:grid;place-items:center;flex:none"
											>{#if on}<span
													style="width:9px;height:9px;border-radius:50%;background:#1C1A17"
												></span>{/if}</span
										><Icon icon={s.i} size={22} style="color:{s.c}" /><span style="flex:1"
											><span style="display:block;font-size:14.5px;font-weight:500">{s.t}</span
											><span style="display:block;font-size:12.5px;color:#6E655C">{s.d}</span></span
										></button
									>{/each}
							</div>
						</div>
						{#if error}<div
								role="alert"
								style="padding:12px 14px;border-radius:2px;background:rgba(156,74,47,.12);color:var(--acento, #9C4A2F);font-size:14px"
							>
								{error}
							</div>{/if}
						<button
							disabled={!envio || !pagos.length || enviando || tienda.vitrina}
							style="padding:18px;border:0;border-radius:2px;background:#1C1A17;color:#F4EFEA;font:600 15px 'Instrument Sans';letter-spacing:.04em;cursor:pointer;display:flex;justify-content:center;align-items:center;gap:8px"
							><Icon icon={I.phLockSimple} size={16} />{tienda.vitrina
								? 'Es una tienda de ejemplo'
								: !pagos.length
									? 'Esta tienda todavía no tiene medios de pago'
									: envio
										? `Pagar ${fmt(total)}`
										: 'Elegí cómo te llega'}</button
						>
					</div>
					<div style="position:sticky;top:90px;padding:24px;background:#EDE6DE;border-radius:2px">
						<div style="font-family:'Instrument Serif',serif;font-size:28px">Tu pedido</div>
						<div style="margin-top:14px;display:flex;flex-direction:column;gap:12px">
							{#each lineas as c (c.clave)}<div style="display:flex;gap:12px;align-items:center">
									<span style="position:relative;flex:none"
										>{#if c.producto.foto}<img
												src={img(c.producto.foto, 120, 150)}
												alt=""
												style="width:56px;height:70px;object-fit:cover;border-radius:2px;background:#E2D9CF"
											/>{/if}<span
											style="position:absolute;top:-6px;right:-6px;width:20px;height:20px;border-radius:50%;background:#1C1A17;color:#F4EFEA;font-size:11px;display:grid;place-items:center"
											>{c.cantidad}</span
										></span
									><span style="flex:1;font-size:14px"
										>{c.producto.nombre}{#if c.variante}<span
												style="display:block;font-size:12.5px;color:#6E655C">{c.variante}</span
											>{/if}</span
									><b style="font-size:14px;font-weight:600">{fmt(c.subtotal)}</b>
								</div>{/each}
						</div>
						<div style="margin-top:18px">
							<CampoCupon
								{cupon}
								lineas={carrito.lineas}
								tinta="#1C1A17"
								suave="#6E655C"
								borde="rgba(28,26,23,.18)"
								fondo="#fff"
								acento="#1C1A17"
								sobreAcento="#F4EFEA"
								radio="2px"
							/>
						</div>
						<div
							style="margin-top:18px;padding-top:14px;border-top:1px solid rgba(28,26,23,.14);display:flex;flex-direction:column;gap:8px;font-size:14px"
						>
							<span style="display:flex;justify-content:space-between"
								><span>Subtotal</span><span>{fmt(subtotal)}</span></span
							><span style="display:flex;justify-content:space-between"
								><span>Envío</span><span
									>{elegida ? (cuenta.envio ? fmt(cuenta.envio) : 'Gratis') : '—'}</span
								></span
							>{#if cuenta.descCupon}<span
									style="display:flex;justify-content:space-between;color:#2F6B47;font-weight:600"
									><span>Cupón {cupon.aplicado}</span><span>− {fmt(cuenta.descCupon)}</span></span
								>{/if}{#if descuento}<span
									style="display:flex;justify-content:space-between;color:#2F6B47;font-weight:600"
									><span>Transferencia −{tienda.cobros.transferencia}%</span><span
										>− {fmt(descuento)}</span
									></span
								>{/if}<span
								style="display:flex;justify-content:space-between;align-items:baseline;margin-top:6px"
								><b>Total</b><b style="font-size:24px;font-weight:600">{fmt(total)}</b></span
							>
						</div>
						<div
							style="margin-top:16px;display:flex;align-items:center;gap:8px;font-size:12.5px;color:#6E655C"
						>
							<Icon icon={I.phShieldCheck} size={16} />Compra protegida · 10 días para arrepentirte
						</div>
					</div>
				</form>
			{/if}
		{/if}

		<!-- ============ TURNOS (no está en el diseño de esta plantilla: con las piezas del checkout) ============ -->
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
				style="padding:32px clamp(16px,4cqw,48px) clamp(56px,8cqw,96px);display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,360px),1fr));gap:clamp(24px,4cqw,56px);align-items:start"
			>
				<input type="hidden" name="servicioId" value={reserva.servicioId ?? ''} />
				<input type="hidden" name="profesionalId" value={reserva.profesionalId ?? ''} />
				<input type="hidden" name="inicio" value={reserva.inicio ?? ''} />
				<div style="display:flex;flex-direction:column;gap:28px;min-width:0">
					<div
						style="display:flex;gap:8px;flex-wrap:wrap;font-size:12.5px;letter-spacing:.1em;text-transform:uppercase"
					>
						{#each PASOS_TURNO as p, i (p)}<span
								style="font-weight:{i === reserva.paso ? 700 : 400};color:{i === reserva.paso
									? '#1C1A17'
									: '#8C8278'}">{i ? '— ' : ''}{i + 1} {p}</span
							>{/each}
					</div>
					<h1
						style="margin:0;font-family:'Instrument Serif',serif;font-weight:400;font-size:clamp(42px,6cqw,72px);line-height:1"
					>
						Reservá tu turno
					</h1>
					<div>
						<div style="display:flex;justify-content:space-between;align-items:center;gap:10px">
							<div style="font-family:'Instrument Serif',serif;font-size:32px">Servicio</div>
						</div>
						<div style="margin-top:14px">
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
										style="display:flex;align-items:center;gap:12px;padding:15px;border:1.5px solid {on
											? '#1C1A17'
											: 'rgba(28,26,23,.18)'};border-radius:2px;background:{on
											? '#F8F4EF'
											: 'transparent'};text-align:left;cursor:pointer;font-family:inherit;color:#1C1A17"
										><span
											style="width:18px;height:18px;border-radius:50%;border:1.5px solid #1C1A17;display:grid;place-items:center;flex:none"
											>{#if on}<span
													style="width:9px;height:9px;border-radius:50%;background:#1C1A17"
												></span>{/if}</span
										><span style="flex:1;min-width:0"
											><span style="display:block;font-size:14.5px;font-weight:600">{s.nombre}</span
											><span style="display:block;font-size:12.5px;color:#6E655C"
												>{duracion(s.duracion)}</span
											></span
										><b style="font-size:14px;white-space:nowrap">{fmt(s.precio)}</b></button
									>{/each}
							</div>
						</div>
					</div>
					<div>
						<div style="display:flex;justify-content:space-between;align-items:center;gap:10px">
							<div style="font-family:'Instrument Serif',serif;font-size:32px">Profesional</div>
						</div>
						<div style="margin-top:14px">
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
										style="flex:none;width:110px;padding:10px;border:1.5px solid {on
											? '#1C1A17'
											: 'rgba(28,26,23,.18)'};border-radius:2px;background:{on
											? '#F8F4EF'
											: 'transparent'};text-align:center;cursor:pointer;font-family:inherit;color:#1C1A17"
										>{#if p?.foto}<img
												src={img(p.foto, 120, 120)}
												alt=""
												style="width:56px;height:56px;border-radius:2px;object-fit:cover;display:block;margin:0 auto;background:#E2D9CF"
											/>{:else}<span
												style="width:56px;height:56px;border-radius:2px;background:#E2D9CF;color:#1C1A17;display:grid;place-items:center;margin:0 auto"
												><Icon icon={p ? I.phUserFill : I.phShuffleBold} size={22} /></span
											>{/if}<span
											style="display:block;margin-top:8px;font-size:13.5px;font-weight:600;overflow:hidden;text-overflow:ellipsis;white-space:nowrap"
											>{p?.nombre ?? 'Cualquiera'}</span
										><span
											style="display:block;font-size:11.5px;color:#6E655C;overflow:hidden;text-overflow:ellipsis;white-space:nowrap"
											>{p ? p.rol.split(' · ')[0] : 'El primero libre'}</span
										></button
									>{/each}
							</div>
						</div>
					</div>
					<div>
						<div style="display:flex;justify-content:space-between;align-items:center;gap:10px">
							<div style="font-family:'Instrument Serif',serif;font-size:32px">Día y horario</div>
							<span
								style="display:flex;align-items:center;gap:10px;font-size:14px;font-weight:600;color:#1C1A17"
								>{#if reserva.hayAntes || reserva.hayDespues}<button
										type="button"
										aria-label="Días anteriores"
										disabled={!reserva.hayAntes}
										onclick={() => reserva.mover(-1)}
										style="display:grid;place-items:center;min-width:44px;min-height:44px;padding:4px;border:0;background:transparent;color:#1C1A17;cursor:pointer"
										><Icon icon={I.phCaretLeftBold} size={14} /></button
									>{/if}{reserva.mes}{#if reserva.hayAntes || reserva.hayDespues}<button
										type="button"
										aria-label="Días siguientes"
										disabled={!reserva.hayDespues}
										onclick={() => reserva.mover(1)}
										style="display:grid;place-items:center;min-width:44px;min-height:44px;padding:4px;border:0;background:transparent;color:#1C1A17;cursor:pointer"
										><Icon icon={I.phCaretRightBold} size={14} /></button
									>{/if}</span
							>
						</div>
						<div style="margin-top:14px">
							<div
								style="display:grid;grid-template-columns:repeat(7,minmax(0,1fr));gap:6px"
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
										style="padding:9px 0;border:1.5px solid {on
											? '#1C1A17'
											: 'rgba(28,26,23,.18)'};border-radius:2px;background:{on
											? '#F8F4EF'
											: 'transparent'};text-align:center;cursor:pointer;font-family:inherit;color:#1C1A17;min-width:0;opacity:{d.abierto
											? 1
											: 0.35}"
										><span
											style="display:block;font-size:11px;font-weight:600;text-transform:uppercase;color:#6E655C"
											>{d.w}</span
										><span style="display:block;font-size:17px;font-weight:700;margin-top:1px"
											>{d.n}</span
										></button
									>{/each}
							</div>
							<div style="margin-top:16px;display:flex;flex-direction:column;gap:12px">
								{#if !reserva.servicioId}
									<p style="margin:0;font-size:14px;color:#6E655C">
										Elegí un servicio para ver los horarios.
									</p>
								{:else if reserva.horarios && !reserva.grupos.length}
									<p style="margin:0;font-size:14px;color:#6E655C">
										No quedan horarios este día. Probá con otro día u otro&nbsp;profesional.
									</p>
								{/if}
								{#each reserva.grupos as g (g.k)}<div>
										<div style="font-size:12.5px;font-weight:600;color:#6E655C;margin-bottom:7px">
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
													style="padding:10px 0;border:1.5px solid {on
														? '#1C1A17'
														: 'rgba(28,26,23,.18)'};border-radius:2px;background:{on
														? '#F8F4EF'
														: 'transparent'};text-align:center;cursor:pointer;font:600 14px 'Instrument Sans';color:#1C1A17;text-decoration:{h.libre
														? 'none'
														: 'line-through'};opacity:{h.libre ? 1 : 0.45}">{h.hora}</button
												>{/each}
										</div>
									</div>{/each}
							</div>
						</div>
					</div>
					{#if reserva.listo}
						<div>
							<div style="display:flex;justify-content:space-between;align-items:center;gap:10px">
								<div style="font-family:'Instrument Serif',serif;font-size:32px">Tus datos</div>
							</div>
							<div style="margin-top:14px">
								<div
									style="display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,200px),1fr));gap:10px"
								>
									{#each [['nombre', 'Nombre', 'text', 'name'], ['telefono', 'WhatsApp', 'tel', 'tel'], ['email', 'Email', 'email', 'email']] as [n, l, tipo, ac] (n)}
										<label
											style="display:block;padding:10px 12px;border:1px solid rgba(28,26,23,.22);border-radius:2px;background:#F8F4EF"
											><span style="display:block;font-size:11.5px;color:#6E655C">{l}</span><input
												name={n}
												type={tipo}
												autocomplete={ac as 'email'}
												required
												maxlength="120"
												style="width:100%;margin-top:2px;padding:0;border:0;background:transparent;font:400 14.5px 'Instrument Sans';color:#1C1A17;outline:none"
											/></label
										>
									{/each}
								</div>
							</div>
						</div>
						<div>
							<div style="display:flex;justify-content:space-between;align-items:center;gap:10px">
								<div style="font-family:'Instrument Serif',serif;font-size:32px">Pago</div>
							</div>
							<div style="margin-top:14px">
								<div
									style="display:flex;align-items:center;gap:12px;padding:15px;border:1.5px solid #1C1A17;border-radius:2px;background:#F8F4EF;color:#1C1A17"
								>
									<span
										style="width:18px;height:18px;border-radius:50%;border:1.5px solid #1C1A17;display:grid;place-items:center;flex:none"
										><span style="width:9px;height:9px;border-radius:50%;background:#1C1A17"
										></span></span
									>{#if reserva.sena}<Icon
											icon={I.siMercadopago}
											size={22}
											style="color:#00B1EA;flex:none"
										/><span style="flex:1;min-width:0"
											><span style="display:block;font-size:14.5px;font-weight:600"
												>Mercado Pago</span
											><span style="display:block;font-size:12.5px;color:#6E655C"
												>Seña de {fmt(reserva.sena)} para confirmar · el resto, en el local</span
											></span
										>{:else}<Icon
											icon={I.phStorefrontFill}
											size={22}
											style="color:#1C1A17;flex:none"
										/><span style="flex:1;min-width:0"
											><span style="display:block;font-size:14.5px;font-weight:600"
												>Pagás en el local</span
											><span style="display:block;font-size:12.5px;color:#6E655C"
												>No hace falta seña: el turno queda confirmado</span
											></span
										>{/if}
								</div>
							</div>
						</div>
					{/if}

					{#if error}<div
							role="alert"
							style="padding:12px 14px;border-radius:2px;background:rgba(156,74,47,.12);color:var(--acento, #9C4A2F);font-size:14px"
						>
							{error}
						</div>{/if}
					<button
						disabled={!reserva.listo || reservando || tienda.vitrina}
						style="padding:18px;border:0;border-radius:2px;background:#1C1A17;color:#F4EFEA;font:600 15px 'Instrument Sans';letter-spacing:.04em;cursor:pointer;display:flex;justify-content:center;align-items:center;gap:8px"
						><Icon icon={I.phCalendarCheckBold} size={16} />{tienda.vitrina
							? 'Es una tienda de ejemplo'
							: !reserva.listo
								? 'Elegí servicio y horario'
								: reserva.sena
									? 'Confirmar y pagar seña'
									: 'Confirmar turno'}</button
					>
				</div>
				<div
					style="position:sticky;top:90px;display:flex;flex-direction:column;gap:12px;min-width:0"
				>
					<div style="padding:24px;background:#EDE6DE;border-radius:2px">
						<div style="font-family:'Instrument Serif',serif;font-size:28px">Tu reserva</div>
						<div style="margin-top:14px;display:flex;align-items:center;gap:12px">
							{#if reserva.profesional?.foto}<img
									src={img(reserva.profesional.foto, 120, 120)}
									alt=""
									style="width:52px;height:52px;border-radius:2px;object-fit:cover;flex:none"
								/>{/if}
							<div style="min-width:0">
								<div style="font-size:15px;font-weight:600">
									{reserva.conQuien ? `Con ${reserva.conQuien}` : 'Con el primero libre'}
								</div>
								<div style="font-size:13px;color:#6E655C">
									{reserva.cuando ?? 'Elegí el día y el horario'}
								</div>
							</div>
						</div>
						{#if reserva.servicio}
							<div
								style="margin-top:16px;padding-top:14px;border-top:1px solid rgba(28,26,23,.14);display:flex;flex-direction:column;gap:8px;font-size:14px"
							>
								<span style="display:flex;justify-content:space-between;gap:10px"
									><span>{reserva.servicio.nombre}</span><span style="white-space:nowrap"
										>{fmt(reserva.servicio.precio)}</span
									></span
								><span style="display:flex;justify-content:space-between;color:#6E655C"
									><span>Duración total</span><span>{reserva.duracion}</span></span
								><span
									style="display:flex;justify-content:space-between;align-items:baseline;margin-top:6px"
									><b>Total</b><b style="font-size:24px;font-weight:600"
										>{fmt(reserva.servicio.precio)}</b
									></span
								>
							</div>
						{/if}
						<div
							style="margin-top:14px;display:flex;align-items:flex-start;gap:8px;font-size:12.5px;line-height:1.45;color:#6E655C"
						>
							<Icon icon={I.phShieldCheck} size={16} style="flex:none" /><span
								>{#if reserva.sena}Seña de {fmt(reserva.sena)} para confirmar. Podés cancelar desde el
									link del turno hasta {anticipacion} antes; la seña te la devuelve la&nbsp;tienda.{:else}No
									pagás nada ahora. Podés cancelar desde el link del turno hasta {anticipacion}
									antes.{/if}</span
							>
						</div>
					</div>
				</div>
			</form>
		{/if}

		<!-- ============ TURNO (no está en el diseño: con las piezas de la confirmación) ============ -->
		{#if pagina === 'turno' && turno}
			{@const cancelado = turno.estado === 'cancelado'}
			<section
				style="padding:clamp(48px,7cqw,96px) clamp(16px,4cqw,48px);max-width:720px;margin:0 auto;text-align:center"
			>
				<span
					style="display:inline-grid;place-items:center;width:64px;height:64px;border-radius:50%;background:#1C1A17;color:#F4EFEA"
					><Icon icon={cancelado ? I.phCalendarX : I.phCheckBold} size={26} /></span
				>
				<h1
					style="margin:20px 0 0;font-family:'Instrument Serif',serif;font-weight:400;font-size:clamp(42px,6cqw,72px);line-height:1;overflow-wrap:anywhere"
				>
					{cancelado
						? 'Turno cancelado'
						: turno.estado === 'reservado'
							? `Falta la seña, ${turno.nombre}`
							: `¡Listo, ${turno.nombre}!`}
				</h1>
				<p style="margin:14px 0 0;font-size:16px;line-height:1.55;color:#4A433C">
					{#if cancelado}Liberamos el horario{turno.senaPagada
							? '. La seña te la devuelve la tienda por Mercado Pago'
							: ''}. Si querés, reservá&nbsp;otro.{:else if turno.estado === 'reservado'}Te
						guardamos el horario unos minutos mientras pagás la seña con Mercado&nbsp;Pago.{:else}Tu
						turno está confirmado. Te esperamos en
						{turno.tienda}.{/if}
				</p>
				{#if aviso === 'pendiente' || aviso === 'error'}<div
						role="status"
						style="margin-top:16px;padding:12px 14px;border-radius:2px;background:rgba(156,74,47,.12);color:var(--acento, #9C4A2F);font-size:14px"
					>
						{aviso === 'pendiente'
							? 'Mercado Pago todavía no acreditó la seña: te avisamos cuando entre.'
							: 'El pago de la seña no salió. Podés reservar de nuevo.'}
					</div>{/if}
				<div
					style="margin-top:24px;padding:20px;background:#EDE6DE;border-radius:2px;text-align:left;font-size:14.5px;line-height:1.6"
				>
					<div
						style="font-size:16px;font-weight:600;text-decoration:{cancelado
							? 'line-through'
							: 'none'}"
					>
						{cuandoLargo(turno.inicio)} con {turno.profesional}
					</div>
					<div style="font-size:13.5px;color:#6E655C;margin-top:2px">
						{turno.servicio} · {fmt(turno.precio)}{turno.sena
							? ` · seña ${fmt(turno.sena)}${turno.senaPagada ? ' pagada' : ''}`
							: ' · pagás en el local'}
					</div>
					{#if turno.direccion}<div style="margin-top:12px;display:flex;gap:8px;font-size:14px">
							<Icon icon={I.phMapPinFill} size={18} style="flex:none" />{turno.direccion}
						</div>{/if}
				</div>
				{#if error}<div
						role="alert"
						style="margin-top:12px;padding:12px 14px;border-radius:2px;background:rgba(156,74,47,.12);color:var(--acento, #9C4A2F);font-size:14px"
					>
						{error}
					</div>{/if}
				<div style="margin-top:22px;display:flex;justify-content:center;gap:10px;flex-wrap:wrap">
					{#if !cancelado}<a
							href="{base}/turno/{turno.codigo}/turno.ics"
							download="turno.ics"
							data-sveltekit-reload
							style="display:flex;align-items:center;gap:8px;text-decoration:none;padding:15px 26px;border:0;border-radius:2px;background:#1C1A17;color:#F4EFEA;font:600 14px 'Instrument Sans';letter-spacing:.04em"
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
									style="text-decoration:none;padding:15px 26px;border:0;border-radius:2px;background:#1C1A17;color:#F4EFEA;font:600 14px 'Instrument Sans';letter-spacing:.04em;cursor:pointer"
									>Sí, cancelar el turno</button
								>{:else}<button
									type="button"
									onclick={() => (seguro = true)}
									style="text-decoration:none;padding:15px 26px;border:1px solid #1C1A17;border-radius:2px;background:transparent;color:#1C1A17;font:600 14px 'Instrument Sans';letter-spacing:.04em;cursor:pointer"
									>Cancelar turno</button
								>{/if}
						</form>
					{:else if cancelado}<a
							href="{base}/turnos"
							style="display:flex;align-items:center;gap:8px;text-decoration:none;padding:15px 26px;border:0;border-radius:2px;background:#1C1A17;color:#F4EFEA;font:600 14px 'Instrument Sans';letter-spacing:.04em"
							><Icon icon={I.phCalendarPlusBold} size={17} />Reservar otro</a
						>{/if}
					<a
						href={base || '/'}
						style="text-decoration:none;padding:15px 26px;border:1px solid #1C1A17;border-radius:2px;background:transparent;color:#1C1A17;font:600 14px 'Instrument Sans';letter-spacing:.04em"
						>Volver a la tienda</a
					>
				</div>
				{#if !cancelado && !turno.cancelable}<p
						style="margin:14px 0 0;font-size:13.5px;color:#6E655C"
					>
						Para cambiarlo o cancelarlo, escribile a la&nbsp;tienda{#if whatsapp}&nbsp;<a
								href={wa(`¡Hola! Tengo un turno el ${cuandoLargo(turno.inicio)}.`)}
								target="_blank"
								rel="noopener"
								style="color:#1C1A17">por WhatsApp</a
							>{/if}.
					</p>{/if}
			</section>
		{/if}

		<!-- ============ LISTO (no está en el diseño: con las piezas del checkout) ============ -->
		{#if pagina === 'listo' && pedido}
			<section
				style="padding:clamp(48px,7cqw,96px) clamp(16px,4cqw,48px);max-width:720px;margin:0 auto;text-align:center"
			>
				<span
					style="display:inline-grid;place-items:center;width:64px;height:64px;border-radius:50%;background:#1C1A17;color:#F4EFEA"
					><Icon icon={I.phCheckBold} size={28} /></span
				>
				<h1
					style="margin:20px 0 0;font-family:'Instrument Serif',serif;font-weight:400;font-size:clamp(42px,6cqw,72px);line-height:1"
				>
					¡Gracias, {pedido.nombre}!
				</h1>
				<p style="margin:14px 0 0;font-size:16px;line-height:1.55;color:#4A433C">
					Recibimos tu pedido <b>#{pedido.numero}</b> por <b>{fmt(pedido.total)}</b>. Te escribimos
					por mail con cada novedad.
				</p>
				<SenaPedido {tema} />
				{#if pedido.pago === 'transferencia'}
					<div
						style="margin-top:24px;padding:20px;background:#EDE6DE;border-radius:2px;text-align:left;font-size:14.5px;line-height:1.6"
					>
						<b>Transferí {fmt(pedido.total)}</b>
						{#if pedido.alias}al {pedido.esCbu ? 'CBU/CVU' : 'alias'}
							<b>{pedido.alias}</b>{:else}con los datos que te mandamos por mail{/if} y mandanos el comprobante{#if whatsapp}&nbsp;<a
								href={wa(`¡Hola! Te mando el comprobante del pedido #${pedido.numero}.`)}
								target="_blank"
								rel="noopener"
								style="color:#1C1A17">por WhatsApp</a
							>{:else}&nbsp;respondiendo el mail del pedido{/if}. Lo preparamos apenas se acredite.
					</div>
				{/if}
				<a
					href={base || '/'}
					style="display:inline-block;margin-top:28px;text-decoration:none;padding:15px 26px;border-radius:2px;background:#1C1A17;color:#F4EFEA;font:600 14px 'Instrument Sans';letter-spacing:.04em"
					>Volver a la tienda</a
				>
				{#if pedido.estado === 'Entregado'}
					<div style="margin-top:clamp(40px,6cqw,64px)">
						<FormResena slug={tienda.slug} {tema} codigo={pedido.codigo} nombre={pedido.nombre} />
					</div>
				{/if}
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
			<section style="padding:clamp(48px,7cqw,96px) clamp(16px,4cqw,48px) 0;text-align:center">
				{#if tienda.ciudad}<div
						style="font-size:12.5px;letter-spacing:.24em;text-transform:uppercase;color:#6E655C"
					>
						{tienda.ciudad}
					</div>{/if}
				<h1
					style="margin:16px auto 0;max-width:900px;font-family:'Instrument Serif',serif;font-weight:400;font-size:clamp(46px,7cqw,104px);line-height:.95;letter-spacing:-.02em"
				>
					{titulo[1] ? `${titulo[0]} ` : titulo[0]}{#if titulo[1]}
						<i>{titulo[1]}</i>{/if}
				</h1>
			</section>
			<section
				style="padding:clamp(40px,6cqw,80px) clamp(16px,4cqw,48px) 0;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,320px),1fr));gap:clamp(16px,3cqw,40px);align-items:center"
			>
				{#if tienda.portada.foto}<div
						style="aspect-ratio:4/5;border-radius:2px;overflow:hidden;background:#D8CEC3"
					>
						<img
							src={img(tienda.portada.foto, 1000)}
							alt=""
							style="width:100%;height:100%;object-fit:cover"
						/>
					</div>{/if}
				<div style="max-width:520px">
					<p
						style="margin:0;font-family:'Instrument Serif',serif;font-size:clamp(24px,2.6cqw,32px);line-height:1.3"
					>
						{tienda.nombre}{tienda.ciudad ? `, desde ${tienda.ciudad}` : ''}. {comoLlega(
							tienda.envios
						)}
					</p>
					{#if tienda.productos.length || tienda.cobros.cuotas > 1 || tienda.promedio}<div
							style="margin-top:28px;display:grid;grid-template-columns:repeat(3,1fr);gap:16px;border-top:1px solid rgba(28,26,23,.14);padding-top:22px"
						>
							{#if tienda.productos.length}<div>
									<div style="font-family:'Instrument Serif',serif;font-size:44px;line-height:1">
										{tienda.productos.length}
									</div>
									<div style="font-size:13px;color:#6E655C;margin-top:4px">productos</div>
								</div>{/if}
							{#if tienda.cobros.cuotas > 1}<div>
									<div style="font-family:'Instrument Serif',serif;font-size:44px;line-height:1">
										{tienda.cobros.cuotas}
									</div>
									<div style="font-size:13px;color:#6E655C;margin-top:4px">cuotas sin interés</div>
								</div>{/if}
							{#if tienda.promedio}<div>
									<div style="font-family:'Instrument Serif',serif;font-size:44px;line-height:1">
										{tienda.promedio.toLocaleString('es-AR')}
									</div>
									<div style="font-size:13px;color:#6E655C;margin-top:4px">
										de promedio en reseñas
									</div>
								</div>{/if}
						</div>{/if}
				</div>
			</section>
			<section
				style="margin-top:clamp(56px,8cqw,110px);padding:clamp(40px,6cqw,80px) clamp(16px,4cqw,48px);background:#1C1A17;color:#F4EFEA;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,320px),1fr));gap:clamp(24px,4cqw,56px)"
			>
				<div>
					<h2
						style="margin:0;font-family:'Instrument Serif',serif;font-weight:400;font-size:clamp(40px,5cqw,64px);line-height:1"
					>
						{tienda.envios.retiro ? 'Vení al local' : 'Hablemos'}
					</h2>
					<div
						style="margin-top:22px;display:flex;flex-direction:column;gap:14px;font-size:15px;color:#DDD3C8"
					>
						{#if tienda.direccion}<span style="display:flex;gap:12px"
								><Icon icon={I.phMapPin} size={19} />{tienda.direccion}</span
							>{/if}{#if tienda.whatsapp}<span style="display:flex;gap:12px"
								><Icon icon={I.phWhatsappLogo} size={19} />{tienda.whatsapp}</span
							>{/if}
					</div>
				</div>
				{#if whatsapp}
					<div style="display:flex;flex-direction:column;gap:10px;justify-content:center">
						<div style="font-family:'Instrument Serif',serif;font-size:30px">Escribinos</div>
						<a
							href={wa('¡Hola!')}
							target="_blank"
							rel="noopener"
							style="display:flex;justify-content:center;align-items:center;gap:8px;padding:16px;border:1px solid rgba(244,239,234,.4);border-radius:2px;font-size:14.5px;font-weight:600;color:#F4EFEA;text-decoration:none"
							><Icon icon={I.phWhatsappLogo} size={18} />Por WhatsApp</a
						>
					</div>
				{/if}
			</section>
		{/if}

		<!-- footer -->
		<footer
			style="padding:clamp(48px,6cqw,80px) clamp(16px,4cqw,48px) 28px;border-top:1px solid rgba(28,26,23,.12)"
		>
			<div
				style="display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,200px),1fr));gap:32px"
			>
				<div style="grid-column:{footSpan};min-width:0;max-width:440px">
					<div style="font-family:'Instrument Serif',serif;font-size:40px;line-height:1">
						{tienda.nombre}
					</div>
					<p style="margin:12px 0 0;font-size:14px;color:#6E655C;line-height:1.55">
						{#if tienda.cobros.cuotas > 1}{tienda.cobros.cuotas} cuotas sin interés{#if tienda.cobros.transferencia > 0}&nbsp;y
								{tienda.cobros.transferencia}% off por transferencia{/if}, en toda la tienda.{:else if tienda.cobros.transferencia > 0}{tienda
								.cobros.transferencia}% off por transferencia, en toda la tienda.{:else}Compra
							protegida, con 10 días para arrepentirte.{/if}
					</p>
					<!-- Newsletter (§50): el bloque del pie del diseño; queda pendiente hasta confirmar por mail. -->
					<form
						onsubmit={nl.enviar}
						novalidate
						style="margin-top:14px;display:flex;border-bottom:1px solid #1C1A17"
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
							style="--ph:#8C8278;flex:1;min-width:0;padding:10px 0;border:0;background:transparent;font:inherit;font-size:14.5px;color:#1C1A17;outline:none"
						/>
						<button
							type="submit"
							aria-label="Sumarme"
							disabled={nl.enviando}
							style="flex:none;width:44px;min-height:44px;padding:0;border:0;background:transparent;color:inherit;display:grid;place-items:center end;cursor:pointer"
							><Icon icon={I.phArrowRight} size={18} /></button
						>
					</form>
					{#if nl.aviso}<p
							role="status"
							style="margin:8px 0 0;font-size:13px;line-height:1.45;color:{nl.aviso.ok
								? '#6E655C'
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
							style="font-size:11.5px;letter-spacing:.14em;text-transform:uppercase;margin-bottom:4px"
							>{f.t}</b
						>{#each f.l as l (l.t)}<a href={l.href} style="color:#4A433C;text-decoration:none"
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
				style="margin-top:40px;padding-top:18px;border-top:1px solid rgba(28,26,23,.1);display:flex;justify-content:space-between;gap:16px;flex-wrap:wrap;font-size:12.5px;color:#8C8278"
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
	.ceibo :global(a:hover) {
		color: var(--acento, #9c4a2f);
	}
	.zoom :global(img) {
		transition: transform 0.8s cubic-bezier(0.2, 0.7, 0.1, 1);
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
	.nl::placeholder {
		color: var(--ph);
		opacity: 1;
	}
</style>
