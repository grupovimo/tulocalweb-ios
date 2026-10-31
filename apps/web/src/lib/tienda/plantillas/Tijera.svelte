<script lang="ts">
	import { camposFactura, OPCIONALES } from '#lib/tienda/factura.ts';
	import { enhance } from '$app/forms';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { untrack } from 'svelte';
	import Icon from '#lib/components/Icon.svelte';
	import LogoTienda from '../LogoTienda.svelte';
	import VacioTienda from '../VacioTienda.svelte';
	import { enOrden } from '../secciones.ts';
	import * as I from '#lib/icons.ts';
	import FormResena from '../FormResena.svelte';
	import CompartirProducto from '../CompartirProducto.svelte';
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
		entregaCorta,
		envia,
		textoEnvios
	} from '../envio.svelte.ts';
	import { Seleccion, tarjetaDe } from '../seleccion.svelte.ts';
	import { partirTitulo } from '../titulo.ts';
	import { Reserva } from '../reserva.svelte.ts';
	import {
		cuandoCorto,
		cuandoLargo as cuandoLargoT,
		duracion,
		horarioPorDias
	} from '#lib/turnos/formato.ts';
	import type {
		OpcionEnvio,
		ProductoPublico,
		ProximosTurnos,
		TiendaPublica,
		TurnoPublico
	} from '../tipos.ts';

	/**
	 * Plantilla Tijera (design/Plantilla Tijera.dc.html), con los datos de la tienda en lugar de los
	 * de ejemplo. Con turnos activos dibuja la reserva (pTurnos), los servicios, el equipo, los
	 * horarios y los libres de hoy. Lo que el diseño muestra sin datos atrás (pase mensual, cupón,
	 * cuenta regresiva, seguidores, puntaje de cada profesional) no se inventa: no se dibuja.
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
	let {
		tienda,
		pagina,
		producto = null,
		categoria = null,
		orden = 'vendidos',
		pedido = null,
		turno = null,
		proximos = null,
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
		turno?: TurnoPublico | null;
		/** Los libres de hoy y el próximo de cada profesional (portada). */
		proximos?: ProximosTurnos | null;
		/** Vuelta de Mercado Pago al pagar la seña: ok, pendiente o error. */
		aviso?: string | null;
		error?: string | null;
	} = $props();

	// En el dominio propio de la tienda los links no llevan /t/<slug> (tienda.base = '').
	const base = $derived(tienda.base ?? `/t/${tienda.slug}`);
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
	// Con variantes de precios distintos, "Desde"; "Sin stock" sólo si no queda ninguna.
	const tarjeta = (p: ProductoPublico) => {
		const t = tarjetaDe(p);
		return {
			id: p.id,
			b: p.categoria,
			n: p.nombre,
			price: (t.desde ? 'Desde ' : '') + fmt(t.precio),
			old: p.anterior && !t.desde ? fmt(p.anterior) : null,
			img: img(p.foto, 600, 600),
			tag: t.agotado ? 'Sin stock' : p.etiqueta,
			transf: fmt(t.precio * (1 - tienda.cobros.transferencia / 100)),
			agotado: t.agotado,
			href: `${base}/producto/${p.slug ?? p.id}`
		};
	};
	type Tarjeta = ReturnType<typeof tarjeta>;
	// Sin logo, la inicial en el cuadradito de la marca.
	const inicial = $derived(tienda.nombre.trim().charAt(0).toUpperCase());
	const ver = (s: string) => tienda.secciones.includes(s as TiendaPublica['secciones'][number]);
	const whatsapp = $derived(tienda.whatsapp.replace(/\D/g, ''));
	const wa = (texto: string) =>
		`https://wa.me/${whatsapp.startsWith('54') ? whatsapp : '54' + whatsapp}?text=${encodeURIComponent(texto)}`;

	// Título de la portada: el segundo tramo va en violeta, como en el diseño ("Tu turno, / en 30 segundos.").
	const titulo = $derived(partirTitulo(tienda.portada.titulo));

	const cats = $derived(tienda.categorias.slice(0, 6));
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
	const chipsCat = $derived([
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
	const resenasProd = $derived(
		producto ? tienda.resenas.filter((r) => r.producto === producto!.nombre) : []
	);
	const promedioProd = $derived(
		resenasProd.length
			? Math.round((resenasProd.reduce((s, r) => s + r.estrellas, 0) / resenasProd.length) * 10) /
					10
			: 0
	);
	let cantidad = $state(1);
	$effect(() => {
		if (producto) cantidad = 1;
	});
	// Tamaño, color, presentación…: lo elegido, la variante, su precio, stock y foto (seleccion.svelte.ts).
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
						i: I.phBankFill,
						c: '#2F7D4F'
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
		...(tienda.turnos
			? [
					{
						t: 'Turnos',
						l: [
							{ t: 'Reservar', href: `${base}/turnos` },
							{ t: 'Horarios', href: `${base}/nosotros` }
						]
					}
				]
			: []),
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
			? [{ t: 'Turnos', href: `${base}/turnos`, on: pagina === 'turnos' || pagina === 'turno' }]
			: []),
		{ t: 'Todo', href: `${base}/catalogo`, on: pagina === 'catalogo' && !categoria },
		...tienda.categorias.slice(0, 3).map((c) => ({
			t: c.nombre,
			href: enlaceCat(c.nombre),
			on: pagina === 'catalogo' && categoria === c.nombre
		})),
		{ t: 'Nosotros', href: `${base}/nosotros`, on: pagina === 'nosotros' }
	]);
	// Reseñas y preguntas con las piezas de Tijera: Syne en los títulos, violeta, botones de píldora.
	const tema: TemaOpiniones = {
		texto: '#1E1B2E',
		suave: '#5A566B',
		campo: '#F6F4F1',
		tarjeta: '#fff',
		borde: 'rgba(30,27,46,.12)',
		acento: '#1E1B2E',
		sobreAcento: '#fff',
		estrella: '#E3A93B',
		radio: '12px',
		radioBoton: '999px',
		fuente: "'Instrument Sans', system-ui, sans-serif",
		titulos: "'Syne', sans-serif",
		pesoTitulos: 800
	};
	const estrellas = (n: number) =>
		Array.from({ length: 5 }, (_, i) =>
			n >= i + 1 ? I.phStarFill : n >= i + 0.5 ? I.phStarHalfFill : I.phStar
		);
	/** Sumar desde una tarjeta o "Te puede gustar": con opciones, primero se elige en la ficha. */
	const sumar = (id: string) => {
		const p = porId.get(id);
		if (!p || tienda.vitrina) return;
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- ruta de la tienda con su slug
		if (!carrito.sumarProducto(p)) goto(`${base}/producto/${p.slug ?? p.id}`);
	};

	// ---- Turnos (pTurnos del diseño) ----
	const T = $derived(tienda.turnos);
	const cara = (f: string | null, w: number, h: number) =>
		!f
			? null
			: f.startsWith('https://')
				? f
				: `https://images.unsplash.com/photo-${f}?auto=format&fit=crop&w=${w}&h=${h}&q=70&crop=faces`;
	const reservaUrl = (q: Record<string, string>) =>
		`${base}/turnos?${new URLSearchParams(q).toString()}`;
	// La portada: los libres de hoy (heroSlots) y el próximo de cada uno (team.next).
	const heroSlots = $derived(
		(proximos?.hoy ?? []).slice(0, 6).map((inicio, i) => ({
			inicio,
			t: inicio.slice(11, 16),
			on: i === 0,
			href: reservaUrl({ servicio: T?.servicios[0]?.id ?? '', dia: inicio.slice(0, 10), inicio })
		}))
	);
	const horas = $derived(T ? horarioPorDias(T.horario) : []);
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
	const PASOS = ['Servicio', 'Profesional', 'Horario', 'Pago'];
	const pasos = $derived(
		PASOS.map((t, i) => {
			const p = reserva?.paso ?? 0;
			const k = i < p ? 'done' : i === p ? 'on' : '';
			return {
				t,
				n: k === 'done' ? '✓' : String(i + 1),
				bg: k === 'on' ? '#1E1B2E' : k === 'done' ? '#EEE9FF' : '#fff',
				fg: k === 'on' ? '#fff' : k === 'done' ? 'var(--acento, #5B3FD6)' : '#8B879A',
				nbg: k === 'on' || k === 'done' ? 'var(--acento, #5B3FD6)' : '#EAE6F2',
				nfg: k === '' ? '#8B879A' : '#fff'
			};
		})
	);
	const ICONO_FRANJA = { manana: I.phSunHorizon, tarde: I.phSun, noche: I.phMoon } as const;
	const slot = (k: 'on' | 'off' | '') => ({
		bg: k === 'on' ? 'var(--acento, #5B3FD6)' : k === 'off' ? 'transparent' : '#fff',
		fg: k === 'on' ? '#fff' : k === 'off' ? '#B7B3C4' : '#1E1B2E',
		bd:
			k === 'on'
				? 'var(--acento, #5B3FD6)'
				: k === 'off'
					? 'rgba(30,27,46,.06)'
					: 'rgba(30,27,46,.12)',
		td: k === 'off' ? 'line-through' : 'none'
	});
	const anticipacion = $derived.by(() => {
		const m = T?.anticipacionMin ?? 0;
		return m >= 120 && m % 60 === 0 ? `${m / 60} h` : `${m} min`;
	});
	let reservando = $state(false);
	let cancelando = $state(false);
	let seguro = $state(false);
</script>

<svelte:window bind:innerWidth={ancho} />

<svelte:head>
	<link
		href="https://fonts.googleapis.com/css2?family=Syne:wght@600;700;800&display=swap"
		rel="stylesheet"
	/>
</svelte:head>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
{#snippet card(p: Tarjeta)}
	<div
		style="padding:10px;border-radius:24px;background:#fff;text-align:left;color:#1E1B2E;box-shadow:0 1px 2px rgba(30,27,46,.06);min-width:0"
	>
		<a
			href={p.href}
			style="position:relative;display:block;aspect-ratio:1;border-radius:18px;overflow:hidden;background:#EAE6EF"
			>{#if p.img}<img
					src={p.img}
					alt={p.n}
					style="width:100%;height:100%;object-fit:cover;display:block"
				/>{/if}{#if p.tag}<span
					style="position:absolute;top:8px;left:8px;padding:4px 9px;border-radius:999px;background:var(--acento, #5B3FD6);color:#fff;font-size:11px;font-weight:700"
					>{p.tag}</span
				>{/if}</a
		>
		<span style="display:block;padding:12px 6px 4px"
			>{#if p.b}<span style="display:block;font-size:12px;color:#8B879A;font-weight:600">{p.b}</span
				>{/if}<a
				href={p.href}
				style="display:block;font-size:15px;font-weight:600;margin-top:2px;color:#1E1B2E;text-decoration:none"
				>{p.n}</a
			><span
				style="display:flex;justify-content:space-between;align-items:center;gap:8px;margin-top:8px"
				><span style="min-width:0"
					><b style="font-size:16px">{p.price}</b>{#if p.old}<s
							style="margin-left:6px;font-size:13px;color:#8B879A">{p.old}</s
						>{/if}{#if tienda.cobros.transferencia > 0}<span
							style="display:block;font-size:12px;color:#2F7D4F;font-weight:600"
							>{p.transf} transf.</span
						>{/if}</span
				><button
					type="button"
					aria-label="Sumar {p.n} al carrito"
					disabled={p.agotado || tienda.vitrina}
					onclick={() => sumar(p.id)}
					style="flex:none;width:36px;height:36px;border:0;border-radius:50%;background:#1E1B2E;color:#fff;display:grid;place-items:center;cursor:pointer"
					><Icon icon={I.phPlusBold} size={15} /></button
				></span
			></span
		>
	</div>
{/snippet}

<div
	class="tijera"
	style="min-height:100vh;font-family:'Instrument Sans',system-ui,sans-serif;color:#1E1B2E"
>
	<div
		style="container-type:inline-size;width:100%;background:#F6F4F1;position:relative;overflow-x:clip"
	>
		<!-- barra promo -->
		<div
			style="display:flex;justify-content:center;align-items:center;gap:12px;flex-wrap:wrap;padding:9px 16px;background:#1E1B2E;color:#F6F4F1;font-size:13px;text-align:center"
		>
			<span
				>{tienda.envios.gratisDesde
					? `Envío gratis desde ${fmt(tienda.envios.gratisDesde)}`
					: tienda.cobros.cuotas > 1
						? `${tienda.cobros.cuotas} cuotas sin interés`
						: 'Compra protegida'}</span
			>{#if wide && (tienda.cobros.cuotas > 1 || tienda.cobros.transferencia > 0)}<span
					style="display:flex;gap:4px"
					>{#if tienda.cobros.cuotas > 1}<span
							style="padding:2px 7px;border-radius:6px;background:rgba(246,244,241,.14)"
							>{tienda.cobros.cuotas} cuotas sin interés</span
						>{/if}{#if tienda.cobros.transferencia > 0}<span
							style="padding:2px 7px;border-radius:6px;background:rgba(246,244,241,.14)"
							>{tienda.cobros.transferencia}% off por transferencia</span
						>{/if}</span
				>{/if}
		</div>

		<!-- header -->
		<header
			style="position:sticky;top:0;z-index:20;display:flex;align-items:center;gap:{nar && T
				? 10
				: 16}px;padding:14px clamp(14px,3.5cqw,44px);background:rgba(246,244,241,.92);backdrop-filter:blur(10px);border-bottom:1px solid rgba(30,27,46,.08)"
		>
			<a
				href={base || '/'}
				style="display:flex;align-items:center;gap:10px;min-width:0;text-decoration:none;color:#1E1B2E"
				>{#if !tienda.logo}<span
						style="width:36px;height:36px;border-radius:11px;background:var(--acento, #5B3FD6);color:#fff;display:grid;place-items:center;flex:none;font-family:'Syne',sans-serif;font-weight:800;font-size:19px"
						>{inicial}</span
					>{/if}<span
					style="font-family:'Syne',sans-serif;font-weight:800;font-size:22px;letter-spacing:-.02em;overflow:hidden;text-overflow:ellipsis;white-space:nowrap"
					><LogoTienda {tienda} /></span
				></a
			>
			{#if wide}<nav style="display:flex;gap:4px;margin-left:12px">
					{#each nav as n (n.t)}<a
							href={n.href}
							style="text-decoration:none;font:600 14px 'Instrument Sans';padding:9px 14px;border-radius:999px;background:{n.on
								? '#EAE6F2'
								: 'transparent'};color:#1E1B2E;white-space:nowrap">{n.t}</a
						>{/each}
				</nav>{/if}
			<span style="flex:1"></span>
			<a
				href="{base}/carrito"
				aria-label="Carrito"
				style="position:relative;color:#1E1B2E;display:grid;flex:none"
				><Icon icon={I.phShoppingBag} size={22} />{#if carrito.unidades}<span
						style="position:absolute;top:-6px;right:-8px;min-width:17px;height:17px;border-radius:999px;background:var(--acento, #5B3FD6);color:#fff;font-size:10.5px;font-weight:700;display:grid;place-items:center"
						>{carrito.unidades}</span
					>{/if}</a
			>
			{#if nar && !T}<a
					href="{base}/catalogo"
					aria-label="Productos"
					style="display:grid;place-items:center;flex:none;width:40px;height:40px;border-radius:999px;background:#1E1B2E;color:#F6F4F1"
					><Icon icon={I.phList} size={16} /></a
				>{/if}
			{#if T}<a
					href="{base}/turnos"
					style="display:flex;align-items:center;gap:8px;flex:none;padding:11px {nar
						? 13
						: 18}px;border-radius:999px;background:#1E1B2E;color:#F6F4F1;font:700 14px 'Instrument Sans';white-space:nowrap;text-decoration:none"
					><Icon icon={I.phCalendarPlusBold} size={16} />Reservar</a
				>{/if}
		</header>

		{#if tienda.vitrina}
			<div
				style="padding:10px 16px;background:#EEE9FF;color:var(--acento, #5B3FD6);font-size:13px;font-weight:600;text-align:center"
			>
				Es una tienda de ejemplo de la plantilla Tijera: podés recorrerla, pero no se compra
				de&nbsp;verdad.
			</div>
		{/if}

		<!-- ============ INICIO ============ -->
		{#if pagina === 'inicio'}
			<!-- Las secciones en el orden que eligió la tienda en Mi tienda. -->
			{#each enOrden( tienda.secciones, ['hero', 'cats', 'prods', 'revs', 'ig'] ) as seccion (seccion)}
				{#if seccion === 'hero'}
					{#if ver('hero')}
						<section
							style="padding:clamp(18px,3cqw,40px) clamp(14px,3.5cqw,44px) 0;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,400px),1fr));gap:clamp(14px,2cqw,24px)"
						>
							<div
								style="display:flex;flex-direction:column;justify-content:center;gap:22px;padding:clamp(8px,2cqw,24px) 0;min-width:0"
							>
								{#if tienda.promedio && tienda.resenas.length}<span
										style="align-self:flex-start;display:flex;align-items:center;gap:8px;padding:7px 12px 7px 10px;border-radius:999px;background:#fff;box-shadow:0 1px 2px rgba(30,27,46,.08);font-size:13px;font-weight:600"
										><Icon
											icon={I.phStarFill}
											size={15}
											style="color:#E3A93B"
										/>{tienda.promedio.toLocaleString('es-AR')} · {tienda.resenas.length}
										{tienda.resenas.length === 1 ? 'reseña' : 'reseñas'}</span
									>{/if}
								<h1
									style="margin:0;font-family:'Syne',sans-serif;font-weight:800;font-size:clamp(42px,5.6cqw,92px);line-height:.92;letter-spacing:-.04em;overflow-wrap:anywhere"
								>
									{titulo[0]}{#if titulo[1]}<br /><span style="color:var(--acento, #5B3FD6)"
											>{titulo[1]}</span
										>{/if}
								</h1>
								{#if tienda.portada.lema}<p
										style="margin:0;font-size:17px;line-height:1.5;color:#5A566B;max-width:460px"
									>
										{tienda.portada.lema}
									</p>{/if}
								<div style="display:flex;gap:10px;flex-wrap:wrap">
									{#if T}<a
											href="{base}/turnos"
											style="display:flex;align-items:center;gap:8px;padding:16px 26px;border:0;border-radius:999px;background:var(--acento, #5B3FD6);color:#fff;font:700 15px 'Instrument Sans';text-decoration:none;box-shadow:0 14px 30px -12px rgba(91,63,214,.7)"
											><Icon icon={I.phCalendarCheckBold} size={17} />Reservar turno</a
										><a
											href="{base}/catalogo"
											style="padding:16px 22px;border:1.5px solid rgba(30,27,46,.18);border-radius:999px;background:transparent;color:#1E1B2E;font:700 15px 'Instrument Sans';text-decoration:none"
											>Ver productos</a
										>{:else}<a
											href="{base}/catalogo"
											style="display:flex;align-items:center;gap:8px;padding:16px 26px;border:0;border-radius:999px;background:var(--acento, #5B3FD6);color:#fff;font:700 15px 'Instrument Sans';text-decoration:none;box-shadow:0 14px 30px -12px rgba(91,63,214,.7)"
											><Icon icon={I.phShoppingBagFill} size={17} />Ver productos</a
										>{/if}{#if whatsapp && !T}<a
											href={wa('¡Hola!')}
											target="_blank"
											rel="noopener"
											style="padding:16px 22px;border:1.5px solid rgba(30,27,46,.18);border-radius:999px;background:transparent;color:#1E1B2E;font:700 15px 'Instrument Sans';text-decoration:none"
											>Escribinos</a
										>{/if}
								</div>
							</div>
							{#if tienda.portada.foto || heroSlots.length}<div
									style="position:relative;min-height:min(560px,120cqw);border-radius:32px;overflow:hidden;background:#CFC8DA"
								>
									{#if tienda.portada.foto}<img
											src={img(tienda.portada.foto, 1100)}
											alt=""
											style="position:absolute;inset:0;width:100%;height:100%;object-fit:cover"
										/>{/if}
									{#if heroSlots.length}
										<div
											style="position:absolute;left:16px;right:16px;bottom:16px;padding:16px;border-radius:22px;background:rgba(246,244,241,.95);backdrop-filter:blur(8px);box-shadow:0 20px 40px -16px rgba(30,27,46,.4)"
										>
											<div
												style="display:flex;justify-content:space-between;align-items:center;gap:10px;flex-wrap:wrap"
											>
												<b style="font-size:14px">Próximos turnos libres hoy</b><span
													style="font-size:12.5px;color:#2F7D4F;font-weight:700"
													>● {proximos?.hoy.length}
													{proximos?.hoy.length === 1 ? 'disponible' : 'disponibles'}</span
												>
											</div>
											<div
												style="margin-top:10px;display:flex;gap:6px;overflow-x:auto;scrollbar-width:none"
											>
												{#each heroSlots as h (h.inicio)}{@const st = slot(h.on ? 'on' : '')}<a
														href={h.href}
														style="flex:none;padding:9px 13px;border-radius:12px;border:1.5px solid {st.bd};background:{st.bg};color:{st.fg};font:700 13.5px 'Instrument Sans';text-decoration:none"
														>{h.t}</a
													>{/each}
											</div>
										</div>
									{/if}
								</div>{/if}
						</section>
					{/if}

					{#if T && T.servicios.length}
						<section style="padding:clamp(48px,7cqw,96px) clamp(14px,3.5cqw,44px) 0">
							<div
								style="display:flex;justify-content:space-between;align-items:end;gap:16px;flex-wrap:wrap"
							>
								<h2
									style="margin:0;font-family:'Syne',sans-serif;font-weight:800;font-size:clamp(34px,4.6cqw,58px);line-height:.95;letter-spacing:-.035em"
								>
									Servicios
								</h2>
								<span style="font-size:14px;color:#5A566B"
									>{T.sena
										? 'Precios finales · pagás en el local o seña online'
										: 'Precios finales · pagás en el local'}</span
								>
							</div>
							<div
								style="margin-top:22px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(100%,300px),1fr));gap:12px"
							>
								{#each T.servicios as s (s.id)}<a
										href={reservaUrl({ servicio: s.id })}
										class="sube"
										style="display:flex;gap:14px;align-items:center;padding:14px;border-radius:22px;background:#fff;text-align:left;text-decoration:none;color:#1E1B2E;box-shadow:0 1px 2px rgba(30,27,46,.06);transition:transform .2s"
										>{#if s.foto}<img
												src={img(s.foto, 200, 200)}
												alt=""
												style="width:76px;height:76px;border-radius:16px;object-fit:cover;flex:none;background:#E4E1EA"
											/>{:else}<span
												style="width:76px;height:76px;border-radius:16px;flex:none;background:#E4E1EA"
											></span>{/if}<span style="flex:1;min-width:0"
											><span style="display:flex;align-items:center;gap:7px;flex-wrap:wrap"
												><span style="font-size:16px;font-weight:700">{s.nombre}</span
												>{#if s.etiqueta}<span
														style="font-size:11px;font-weight:700;padding:3px 8px;border-radius:999px;background:#EEE9FF;color:var(--acento, #5B3FD6)"
														>{s.etiqueta}</span
													>{/if}</span
											>{#if s.descripcion}<span
													style="display:block;font-size:13px;color:#5A566B;margin-top:3px"
													>{s.descripcion}</span
												>{/if}<span style="display:flex;gap:10px;margin-top:8px;font-size:13px"
												><b>{fmt(s.precio)}</b><span
													style="color:#8B879A;display:flex;align-items:center;gap:4px"
													><Icon icon={I.phClock} size={13} />{duracion(s.duracion)}</span
												></span
											></span
										><Icon
											icon={I.phPlusCircleFill}
											size={28}
											style="color:var(--acento, #5B3FD6);flex:none"
										/></a
									>{/each}
							</div>
						</section>
					{/if}

					{#if T && T.profesionales.length}
						<section style="padding:clamp(48px,7cqw,96px) clamp(14px,3.5cqw,44px) 0">
							<h2
								style="margin:0;font-family:'Syne',sans-serif;font-weight:800;font-size:clamp(34px,4.6cqw,58px);line-height:.95;letter-spacing:-.035em"
							>
								El equipo
							</h2>
							<div
								style="margin-top:22px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(46%,220px),1fr));gap:12px"
							>
								{#each T.profesionales as p (p.id)}
									{@const prox = proximos?.porProfesional[p.id]}
									<a
										href={reservaUrl({ profesional: p.id })}
										style="display:block;border-radius:24px;overflow:hidden;background:#fff;box-shadow:0 1px 2px rgba(30,27,46,.06);text-decoration:none;color:#1E1B2E;min-width:0"
									>
										<div style="aspect-ratio:4/5;background:#E4E1EA">
											{#if p.foto}<img
													src={cara(p.foto, 500, 620)}
													alt=""
													style="width:100%;height:100%;object-fit:cover;display:block"
												/>{/if}
										</div>
										<div style="padding:14px">
											<div style="font-size:16px;font-weight:700">{p.nombre}</div>
											{#if p.rol}<div style="font-size:13px;color:#5A566B;margin-top:2px">
													{p.rol}
												</div>{/if}
											{#if prox}<div
													style="margin-top:10px;display:flex;justify-content:flex-end;align-items:center;font-size:12.5px"
												>
													<span style="color:#2F7D4F;font-weight:700">{cuandoCorto(prox)}</span>
												</div>{/if}
										</div>
									</a>
								{/each}
							</div>
						</section>
					{/if}
				{:else if seccion === 'cats'}
					{#if ver('cats') && cats.length}
						<section style="padding:clamp(48px,7cqw,96px) clamp(14px,3.5cqw,44px) 0">
							<div
								style="display:flex;justify-content:space-between;align-items:end;gap:16px;flex-wrap:wrap"
							>
								<h2
									style="margin:0;font-family:'Syne',sans-serif;font-weight:800;font-size:clamp(34px,4.6cqw,58px);line-height:.95;letter-spacing:-.035em"
								>
									Categorías
								</h2>
								{#if tienda.cobros.cuotas > 1}<span style="font-size:14px;color:#5A566B"
										>{tienda.cobros.cuotas} cuotas sin interés{#if tienda.cobros.transferencia > 0}&nbsp;·
											{tienda.cobros.transferencia}% off por transferencia{/if}</span
									>{:else if tienda.cobros.transferencia > 0}<span
										style="font-size:14px;color:#5A566B"
										>{tienda.cobros.transferencia}% off por transferencia</span
									>{/if}
							</div>
							<div
								style="margin-top:22px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(100%,300px),1fr));gap:12px"
							>
								{#each cats as c (c.nombre)}<a
										href={enlaceCat(c.nombre)}
										class="sube"
										style="display:flex;gap:14px;align-items:center;padding:14px;border-radius:22px;background:#fff;text-align:left;text-decoration:none;color:#1E1B2E;box-shadow:0 1px 2px rgba(30,27,46,.06);transition:transform .2s"
										>{#if c.foto}<img
												src={img(c.foto, 200, 200)}
												alt=""
												style="width:76px;height:76px;border-radius:16px;object-fit:cover;flex:none;background:#E4E1EA"
											/>{:else}<span
												style="width:76px;height:76px;border-radius:16px;flex:none;background:#E4E1EA"
											></span>{/if}<span style="flex:1;min-width:0"
											><span style="display:block;font-size:16px;font-weight:700">{c.nombre}</span
											><span style="display:block;font-size:13px;color:#5A566B;margin-top:3px"
												>{c.cantidad} {c.cantidad === 1 ? 'producto' : 'productos'}</span
											></span
										><Icon
											icon={I.phPlusCircleFill}
											size={28}
											style="color:var(--acento, #5B3FD6);flex:none"
										/></a
									>{/each}
							</div>
						</section>
					{/if}
				{:else if seccion === 'prods'}
					{#if ver('prods') && best.length}
						<section style="padding:clamp(48px,7cqw,96px) clamp(14px,3.5cqw,44px) 0">
							<div
								style="display:flex;justify-content:space-between;align-items:end;gap:16px;flex-wrap:wrap"
							>
								<h2
									style="margin:0;font-family:'Syne',sans-serif;font-weight:800;font-size:clamp(34px,4.6cqw,58px);line-height:.95;letter-spacing:-.035em"
								>
									Lo más elegido
								</h2>
								<a
									href="{base}/catalogo"
									style="display:flex;align-items:center;gap:8px;padding:11px 16px;border-radius:999px;background:#fff;font-size:13.5px;font-weight:700;color:#1E1B2E;text-decoration:none;box-shadow:0 1px 2px rgba(30,27,46,.08)"
									>Ver todo</a
								>
							</div>
							<div
								style="margin-top:22px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(46%,220px),1fr));gap:12px"
							>
								{#each best as p (p.id)}{@render card(p)}{/each}
							</div>
						</section>
					{/if}
				{:else if seccion === 'revs'}
					{#if ver('revs') && tienda.resenas.length}
						<section style="padding:clamp(48px,7cqw,96px) clamp(14px,3.5cqw,44px) 0">
							<div
								style="display:flex;justify-content:space-between;align-items:end;gap:16px;flex-wrap:wrap"
							>
								<h2
									style="margin:0;font-family:'Syne',sans-serif;font-weight:800;font-size:clamp(34px,4.6cqw,58px);line-height:.95;letter-spacing:-.035em"
								>
									Lo que dicen
								</h2>
								{#if tienda.promedio}<span
										style="display:flex;align-items:center;gap:6px;font-size:14px"
										><Icon icon={I.phStarFill} size={16} style="color:#E3A93B" /><b
											>{tienda.promedio.toLocaleString('es-AR')}</b
										><span style="color:#5A566B"
											>· {tienda.resenas.length}
											{tienda.resenas.length === 1 ? 'reseña' : 'reseñas'}</span
										></span
									>{/if}
							</div>
							<div
								style="margin-top:22px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,290px),1fr));gap:12px"
							>
								{#each tienda.resenas.slice(0, 3) as r, i (i)}
									{@const pr = tienda.productos.find((p) => p.nombre === r.producto)}
									<div
										style="padding:12px;border-radius:24px;background:#fff;box-shadow:0 1px 2px rgba(30,27,46,.06)"
									>
										{#if pr?.foto}<div
												style="aspect-ratio:4/3;border-radius:16px;overflow:hidden;background:#E4E1EA"
											>
												<img
													src={img(pr.foto, 600, 450)}
													alt=""
													style="width:100%;height:100%;object-fit:cover"
												/>
											</div>{/if}
										<div style="padding:12px 4px 4px">
											<div style="display:flex;gap:1px;color:#E3A93B">
												{#each estrellas(r.estrellas) as e, j (j)}<Icon icon={e} size={14} />{/each}
											</div>
											<p style="margin:8px 0 0;font-size:15px;line-height:1.45">“{r.texto}”</p>
											<div style="margin-top:8px;font-size:12.5px;color:#5A566B">
												{r.autor}{#if r.producto}
													· {r.conCompra ? 'compró' : 'sobre'} {r.producto}{/if}
											</div>
											{#if r.conCompra}<span
													style="margin-top:8px;display:inline-flex;align-items:center;gap:4px;padding:4px 10px;border-radius:999px;background:#EEEAFE;color:var(--acento, #5B3FD6);font-size:11.5px;font-weight:700"
													><Icon icon={I.phSealCheckFill} size={13} />Compra verificada</span
												>{/if}
											{#if r.respuesta}<div
													style="margin-top:10px;padding:12px 14px;border-radius:16px;background:#F6F4F1;font-size:13px;line-height:1.5;color:#4A4659"
												>
													<b style="display:block;color:#1E1B2E">Respuesta de {tienda.nombre}</b
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
						<section style="padding:clamp(48px,7cqw,96px) clamp(14px,3.5cqw,44px)">
							<h2
								style="margin:0;font-family:'Syne',sans-serif;font-weight:800;font-size:clamp(34px,4.6cqw,58px);line-height:.95;letter-spacing:-.035em"
							>
								Así trabajamos
							</h2>
							<div
								style="margin-top:20px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(30%,160px),1fr));gap:8px"
							>
								{#each tienda.fotos as g (g)}<div
										style="aspect-ratio:1;border-radius:16px;overflow:hidden;background:#E4E1EA"
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
				<div style="height:clamp(48px,7cqw,96px)"></div>
			{/if}
		{/if}

		<!-- ============ TURNOS (pTurnos) ============ -->
		{#if pagina === 'turnos' && T && reserva}
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
				style="display:block;padding:24px clamp(14px,3.5cqw,44px) clamp(48px,7cqw,90px)"
			>
				<input type="hidden" name="servicioId" value={reserva.servicioId ?? ''} />
				<input type="hidden" name="profesionalId" value={reserva.profesionalId ?? ''} />
				<input type="hidden" name="inicio" value={reserva.inicio ?? ''} />
				<div
					style="display:flex;gap:6px;font-size:12.5px;font-weight:700;letter-spacing:.06em;text-transform:uppercase;flex-wrap:wrap"
				>
					{#each pasos as s (s.t)}<span
							style="display:flex;align-items:center;gap:7px;padding:7px 12px;border-radius:999px;background:{s.bg};color:{s.fg}"
							><span
								style="width:18px;height:18px;border-radius:50%;display:grid;place-items:center;background:{s.nbg};color:{s.nfg};font-size:10.5px"
								>{s.n}</span
							>{s.t}</span
						>{/each}
				</div>
				<h1
					style="margin:18px 0 0;font-family:'Syne',sans-serif;font-weight:800;font-size:clamp(36px,5cqw,68px);line-height:.95;letter-spacing:-.04em;overflow-wrap:anywhere"
				>
					Reservá tu turno
				</h1>
				<div
					style="margin-top:24px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,360px),1fr));gap:clamp(16px,3cqw,36px);align-items:start"
				>
					<div style="display:flex;flex-direction:column;gap:24px;min-width:0">
						<div>
							<div style="font-size:15px;font-weight:700;margin-bottom:10px">1 · Servicio</div>
							<div
								style="display:flex;flex-direction:column;gap:8px"
								role="radiogroup"
								aria-label="Servicio"
							>
								{#each T.servicios as s (s.id)}{@const on = reserva.servicioId === s.id}<button
										type="button"
										role="radio"
										aria-checked={on}
										onclick={() => reserva.elegirServicio(s.id)}
										style="display:flex;align-items:center;gap:12px;padding:14px;border-radius:18px;border:1.5px solid {on
											? 'var(--acento, #5B3FD6)'
											: 'rgba(30,27,46,.1)'};background:{on
											? '#F3F0FF'
											: '#fff'};text-align:left;cursor:pointer;font-family:inherit;color:#1E1B2E"
										><span
											style="width:22px;height:22px;border-radius:7px;border:1.5px solid {on
												? 'var(--acento, #5B3FD6)'
												: 'rgba(30,27,46,.25)'};background:{on
												? 'var(--acento, #5B3FD6)'
												: '#fff'};color:#fff;display:grid;place-items:center;flex:none"
											>{#if on}<Icon icon={I.phCheckBold} size={13} />{/if}</span
										><span style="flex:1;min-width:0"
											><span style="display:block;font-size:15px;font-weight:600">{s.nombre}</span
											><span style="display:block;font-size:12.5px;color:#5A566B"
												>{duracion(s.duracion)}</span
											></span
										><b style="font-size:15px;white-space:nowrap">{fmt(s.precio)}</b></button
									>{/each}
							</div>
						</div>
						<div>
							<div style="font-size:15px;font-weight:700;margin-bottom:10px">2 · Profesional</div>
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
										style="flex:none;width:110px;padding:10px;border-radius:18px;border:1.5px solid {on
											? 'var(--acento, #5B3FD6)'
											: 'rgba(30,27,46,.1)'};background:{on
											? '#F3F0FF'
											: '#fff'};text-align:center;cursor:pointer;font-family:inherit;color:#1E1B2E"
										>{#if p?.foto}<img
												src={cara(p.foto, 120, 120)}
												alt=""
												style="width:56px;height:56px;border-radius:50%;object-fit:cover;display:block;margin:0 auto"
											/>{:else}<span
												style="width:56px;height:56px;border-radius:50%;background:#EEE9FF;color:var(--acento, #5B3FD6);display:grid;place-items:center;margin:0 auto"
												><Icon icon={p ? I.phUserFill : I.phShuffleBold} size={22} /></span
											>{/if}<span
											style="display:block;margin-top:8px;font-size:13.5px;font-weight:700;overflow:hidden;text-overflow:ellipsis;white-space:nowrap"
											>{p?.nombre ?? 'Cualquiera'}</span
										><span
											style="display:block;font-size:11.5px;color:#5A566B;overflow:hidden;text-overflow:ellipsis;white-space:nowrap"
											>{p ? p.rol.split(' · ')[0] : 'El primero libre'}</span
										></button
									>{/each}
							</div>
						</div>
						<div>
							<div
								style="display:flex;justify-content:space-between;align-items:center;margin-bottom:10px"
							>
								<span style="font-size:15px;font-weight:700">3 · Día y horario</span><span
									style="display:flex;align-items:center;gap:10px;font-size:14px;font-weight:600"
									>{#if reserva.hayAntes || reserva.hayDespues}<button
											type="button"
											aria-label="Días anteriores"
											disabled={!reserva.hayAntes}
											onclick={() => reserva.mover(-1)}
											style="display:grid;place-items:center;min-width:44px;min-height:44px;padding:4px;border:0;background:transparent;color:#1E1B2E;cursor:pointer"
											><Icon icon={I.phCaretLeftBold} size={14} /></button
										>{/if}{reserva.mes}{#if reserva.hayAntes || reserva.hayDespues}<button
											type="button"
											aria-label="Días siguientes"
											disabled={!reserva.hayDespues}
											onclick={() => reserva.mover(1)}
											style="display:grid;place-items:center;min-width:44px;min-height:44px;padding:4px;border:0;background:transparent;color:#1E1B2E;cursor:pointer"
											><Icon icon={I.phCaretRightBold} size={14} /></button
										>{/if}</span
								>
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
										style="padding:9px 0;border-radius:14px;text-align:center;background:{sel
											? 'var(--acento, #5B3FD6)'
											: '#fff'};color:{sel ? '#fff' : '#1E1B2E'};border:1.5px solid {sel
											? 'var(--acento, #5B3FD6)'
											: 'rgba(30,27,46,.08)'};opacity:{d.abierto
											? 1
											: 0.35};cursor:pointer;font-family:inherit;min-width:0"
										><span
											style="display:block;font-size:11px;font-weight:600;text-transform:uppercase"
											>{d.w}</span
										><span style="display:block;font-size:17px;font-weight:700;margin-top:1px"
											>{d.n}</span
										><span
											style="display:block;width:5px;height:5px;border-radius:50%;margin:3px auto 0;background:{!d.abierto
												? 'transparent'
												: sel
													? '#fff'
													: '#3F9A50'}"
										></span></button
									>{/each}
							</div>
							<div style="margin-top:16px;display:flex;flex-direction:column;gap:12px">
								{#if !reserva.servicioId}
									<p style="margin:0;font-size:14px;color:#5A566B">
										Elegí un servicio para ver los horarios.
									</p>
								{:else if reserva.horarios && !reserva.grupos.length}
									<p style="margin:0;font-size:14px;color:#5A566B">
										No quedan horarios este día. Probá con otro día u otro&nbsp;profesional.
									</p>
								{/if}
								{#each reserva.grupos as g (g.k)}<div>
										<div
											style="font-size:12.5px;font-weight:700;color:#5A566B;margin-bottom:7px;display:flex;align-items:center;gap:6px"
										>
											<Icon icon={ICONO_FRANJA[g.k]} size={14} />{g.t}
										</div>
										<div
											style="display:grid;grid-template-columns:repeat(auto-fill,minmax(76px,1fr));gap:6px"
											role="radiogroup"
											aria-label="Horarios a la {g.t.toLowerCase()}"
										>
											{#each g.s as h (h.inicio)}{@const st = slot(
													reserva.inicio === h.inicio ? 'on' : h.libre ? '' : 'off'
												)}<button
													type="button"
													role="radio"
													aria-checked={reserva.inicio === h.inicio}
													aria-label="{h.hora}{h.libre ? '' : ', ocupado'}"
													disabled={!h.libre}
													onclick={() => reserva.elegirHorario(h)}
													style="padding:10px 0;text-align:center;border-radius:12px;border:1.5px solid {st.bd};background:{st.bg};color:{st.fg};font:700 14px 'Instrument Sans';text-decoration:{st.td};cursor:pointer"
													>{h.hora}</button
												>{/each}
										</div>
									</div>{/each}
							</div>
						</div>
						{#if reserva.listo}
							<div>
								<div style="font-size:15px;font-weight:700;margin-bottom:10px">4 · Tus datos</div>
								<div
									style="display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));gap:8px"
								>
									{#each [['nombre', 'Nombre', 'text', 'name'], ['telefono', 'WhatsApp', 'tel', 'tel'], ['email', 'Email', 'email', 'email']] as [n, l, tipo, ac] (n)}
										<label
											style="display:block;padding:11px 14px;border-radius:16px;background:#fff;box-shadow:inset 0 0 0 1.5px rgba(30,27,46,.1)"
											><span style="display:block;font-size:11.5px;color:#8B879A">{l}</span><input
												name={n}
												type={tipo}
												autocomplete={ac as 'email'}
												required
												maxlength="120"
												style="width:100%;margin-top:2px;padding:0;border:0;background:transparent;font:400 14.5px 'Instrument Sans';color:#1E1B2E;outline:none"
											/></label
										>
									{/each}
								</div>
							</div>
							<div>
								<div style="font-size:15px;font-weight:700;margin-bottom:10px">Pago</div>
								<div
									style="display:flex;align-items:center;gap:12px;padding:15px;border-radius:18px;border:1.5px solid var(--acento, #5B3FD6);background:#F3F0FF"
								>
									<span
										style="width:20px;height:20px;border-radius:50%;border:2px solid var(--acento, #5B3FD6);display:grid;place-items:center;flex:none"
										><span
											style="width:10px;height:10px;border-radius:50%;background:var(--acento, #5B3FD6)"
										></span></span
									>{#if reserva.sena}<Icon
											icon={I.siMercadopago}
											size={22}
											style="color:#00B1EA;flex:none"
										/><span style="flex:1;min-width:0"
											><span style="display:block;font-size:14.5px;font-weight:600"
												>Mercado Pago</span
											><span style="display:block;font-size:12.5px;color:#5A566B"
												>Seña de {fmt(reserva.sena)} para confirmar · el resto, en el local</span
											></span
										>{:else}<Icon
											icon={I.phStorefrontFill}
											size={22}
											style="color:#1E1B2E;flex:none"
										/><span style="flex:1;min-width:0"
											><span style="display:block;font-size:14.5px;font-weight:600"
												>Pagás en el local</span
											><span style="display:block;font-size:12.5px;color:#5A566B"
												>No hace falta seña: el turno queda confirmado</span
											></span
										>{/if}
								</div>
							</div>
						{/if}
					</div>
					<div
						style="position:sticky;top:90px;padding:22px;border-radius:28px;background:#1E1B2E;color:#F6F4F1;min-width:0"
					>
						<div style="font-family:'Syne',sans-serif;font-weight:700;font-size:24px">
							Tu reserva
						</div>
						<div style="margin-top:16px;display:flex;align-items:center;gap:12px">
							{#if reserva.profesional?.foto}<img
									src={cara(reserva.profesional.foto, 120, 120)}
									alt=""
									style="width:52px;height:52px;border-radius:50%;object-fit:cover;flex:none"
								/>{:else}<span
									style="width:52px;height:52px;border-radius:50%;background:rgba(246,244,241,.08);color:#A996FF;display:grid;place-items:center;flex:none"
									><Icon icon={I.phShuffleBold} size={22} /></span
								>{/if}
							<div style="min-width:0">
								<div style="font-size:15px;font-weight:700">
									{reserva.conQuien ? `Con ${reserva.conQuien}` : 'Con el primero libre'}
								</div>
								<div style="font-size:13px;color:#B3AEC6">
									{reserva.cuando ?? 'Elegí el día y el horario'}
								</div>
							</div>
						</div>
						{#if reserva.servicio}
							<div
								style="margin-top:16px;padding-top:14px;border-top:1px solid rgba(246,244,241,.12);display:flex;flex-direction:column;gap:9px;font-size:14px"
							>
								<span style="display:flex;justify-content:space-between;gap:10px"
									><span>{reserva.servicio.nombre}</span><span style="white-space:nowrap"
										>{fmt(reserva.servicio.precio)}</span
									></span
								><span style="display:flex;justify-content:space-between;color:#B3AEC6"
									><span>Duración total</span><span>{reserva.duracion}</span></span
								><span
									style="display:flex;justify-content:space-between;align-items:baseline;margin-top:6px"
									><b>Total</b><b style="font-size:24px">{fmt(reserva.servicio.precio)}</b></span
								>
							</div>
						{/if}
						<div
							style="margin-top:14px;padding:12px;border-radius:16px;background:rgba(246,244,241,.08);font-size:13px;line-height:1.45;display:flex;gap:10px"
						>
							<Icon icon={I.phShieldCheckFill} size={20} style="color:#A996FF;flex:none" />
							<span
								>{#if reserva.sena}Seña de {fmt(reserva.sena)} para confirmar. Podés cancelar desde el
									link del turno hasta {anticipacion} antes; la seña te la devuelve la&nbsp;tienda.{:else}No
									pagás nada ahora. Podés cancelar desde el link del turno hasta {anticipacion}
									antes.{/if}</span
							>
						</div>
						{#if error}<div
								role="alert"
								style="margin-top:12px;padding:12px 14px;border-radius:16px;background:#FDECEC;color:#B3261E;font-size:14px"
							>
								{error}
							</div>{/if}
						<button
							disabled={!reserva.listo || reservando || tienda.vitrina}
							style="margin-top:16px;width:100%;padding:16px;border:0;border-radius:999px;background:var(--acento, #5B3FD6);color:#fff;font:700 15px 'Instrument Sans';cursor:pointer"
							>{tienda.vitrina
								? 'Es una tienda de ejemplo'
								: !reserva.listo
									? 'Elegí servicio y horario'
									: reserva.sena
										? 'Confirmar y pagar seña'
										: 'Confirmar turno'}</button
						>
					</div>
				</div>
			</form>
		{/if}

		<!-- ============ TURNO (no está en el diseño: con las piezas de "Tu reserva" y la confirmación) ============ -->
		{#if pagina === 'turno' && turno}
			{@const cancelado = turno.estado === 'cancelado'}
			<section
				style="padding:clamp(40px,6cqw,80px) clamp(14px,3.5cqw,44px);max-width:720px;margin:0 auto;text-align:center"
			>
				<span
					style="display:inline-grid;place-items:center;width:64px;height:64px;border-radius:50%;background:{cancelado
						? '#1E1B2E'
						: 'var(--acento, #5B3FD6)'};color:#fff;box-shadow:0 14px 30px -12px rgba(91,63,214,.7)"
					><Icon icon={cancelado ? I.phCalendarX : I.phCheckBold} size={28} /></span
				>
				<h1
					style="margin:20px 0 0;font-family:'Syne',sans-serif;font-weight:800;font-size:clamp(36px,5cqw,60px);line-height:.95;letter-spacing:-.04em;overflow-wrap:anywhere"
				>
					{cancelado
						? 'Turno cancelado'
						: turno.estado === 'reservado'
							? `Falta la seña, ${turno.nombre}`
							: `¡Listo, ${turno.nombre}!`}
				</h1>
				<p style="margin:14px 0 0;font-size:16px;line-height:1.55;color:#5A566B">
					{#if cancelado}Liberamos el horario{turno.senaPagada
							? '. La seña te la devuelve la tienda por Mercado Pago'
							: ''}. Si querés, reservá&nbsp;otro.{:else if turno.estado === 'reservado'}Te
						guardamos el horario unos minutos mientras pagás la seña con Mercado&nbsp;Pago.{:else}Tu
						turno está confirmado. Te esperamos en
						{turno.tienda}.{/if}
				</p>
				{#if aviso === 'pendiente' || aviso === 'error'}<div
						role="status"
						style="margin-top:16px;padding:12px 14px;border-radius:16px;background:#FDECEC;color:#B3261E;font-size:14px"
					>
						{aviso === 'pendiente'
							? 'Mercado Pago todavía no acreditó la seña: te avisamos cuando entre.'
							: 'El pago de la seña no salió. Podés reservar de nuevo.'}
					</div>{/if}
				<div
					style="margin-top:24px;padding:22px;border-radius:28px;background:#1E1B2E;color:#F6F4F1;text-align:left"
				>
					<div
						style="display:flex;align-items:center;gap:8px;font-size:12.5px;font-weight:700;letter-spacing:.08em;text-transform:uppercase;color:#A996FF"
					>
						<Icon icon={I.phCalendarCheckFill} size={15} />Turno
					</div>
					<div
						style="margin-top:8px;font-size:17px;font-weight:700;text-decoration:{cancelado
							? 'line-through'
							: 'none'}"
					>
						{cuandoLargoT(turno.inicio)} con {turno.profesional}
					</div>
					<div style="font-size:13.5px;color:#B3AEC6;margin-top:2px">
						{turno.servicio} · {fmt(turno.precio)}{turno.sena
							? ` · seña ${fmt(turno.sena)}${turno.senaPagada ? ' pagada' : ''}`
							: ' · pagás en el local'}
					</div>
					{#if turno.direccion}<div
							style="margin-top:14px;padding-top:12px;border-top:1px solid rgba(246,244,241,.12);display:flex;gap:10px;font-size:14px"
						>
							<Icon
								icon={I.phMapPinFill}
								size={18}
								style="color:#A996FF;flex:none"
							/>{turno.direccion}
						</div>{/if}
				</div>
				{#if error}<div
						role="alert"
						style="margin-top:12px;padding:12px 14px;border-radius:16px;background:#FDECEC;color:#B3261E;font-size:14px"
					>
						{error}
					</div>{/if}
				<div style="margin-top:22px;display:flex;justify-content:center;gap:10px;flex-wrap:wrap">
					{#if !cancelado}<a
							href="{base}/turno/{turno.codigo}/turno.ics"
							download="turno.ics"
							data-sveltekit-reload
							style="display:flex;align-items:center;gap:8px;padding:16px 26px;border-radius:999px;background:var(--acento, #5B3FD6);color:#fff;font:700 15px 'Instrument Sans';text-decoration:none;box-shadow:0 14px 30px -12px rgba(91,63,214,.7)"
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
									style="padding:16px 22px;border:0;border-radius:999px;background:#1E1B2E;color:#fff;font:700 15px 'Instrument Sans';cursor:pointer"
									>Sí, cancelar el turno</button
								>{:else}<button
									type="button"
									onclick={() => (seguro = true)}
									style="padding:16px 22px;border:1.5px solid rgba(30,27,46,.18);border-radius:999px;background:transparent;color:#1E1B2E;font:700 15px 'Instrument Sans';cursor:pointer"
									>Cancelar turno</button
								>{/if}
						</form>
					{:else if cancelado}<a
							href="{base}/turnos"
							style="display:flex;align-items:center;gap:8px;padding:16px 26px;border-radius:999px;background:var(--acento, #5B3FD6);color:#fff;font:700 15px 'Instrument Sans';text-decoration:none"
							><Icon icon={I.phCalendarPlusBold} size={17} />Reservar otro</a
						>{/if}
					<a
						href={base || '/'}
						style="padding:16px 22px;border:1.5px solid rgba(30,27,46,.18);border-radius:999px;color:#1E1B2E;font:700 15px 'Instrument Sans';text-decoration:none"
						>Volver a la tienda</a
					>
				</div>
				{#if !cancelado && !turno.cancelable}<p
						style="margin:14px 0 0;font-size:13.5px;color:#5A566B"
					>
						Para cambiarlo o cancelarlo, escribile a la&nbsp;tienda{#if whatsapp}&nbsp;<a
								href={wa(`¡Hola! Tengo un turno el ${cuandoLargoT(turno.inicio)}.`)}
								target="_blank"
								rel="noopener"
								style="color:var(--acento, #5B3FD6)">por WhatsApp</a
							>{/if}.
					</p>{/if}
			</section>
		{/if}

		<!-- ============ CATÁLOGO ============ -->
		{#if pagina === 'catalogo' && !tienda.productos.length}
			<section style="padding:clamp(40px,7cqw,96px) clamp(16px,4cqw,48px)">
				<VacioTienda {tienda} {tema} motivo="sin-productos" />
			</section>
		{:else if pagina === 'catalogo'}
			<section style="padding:24px clamp(14px,3.5cqw,44px) clamp(48px,7cqw,90px)">
				<h1
					style="margin:0;font-family:'Syne',sans-serif;font-weight:800;font-size:clamp(36px,5cqw,68px);line-height:.95;letter-spacing:-.04em;overflow-wrap:anywhere"
				>
					{categoria ?? 'Todos los productos'}
				</h1>
				<p style="margin:10px 0 0;font-size:15.5px;color:#5A566B;max-width:560px">
					{listado.length}
					{listado.length === 1 ? 'producto' : 'productos'}{#if tienda.cobros.cuotas > 1}&nbsp;· {tienda
							.cobros.cuotas} cuotas sin interés{#if tienda.cobros.transferencia > 0}&nbsp;o {tienda
								.cobros.transferencia}% off por&nbsp;transferencia{/if}{:else if tienda.cobros.transferencia > 0}&nbsp;·
						{tienda.cobros.transferencia}% off por&nbsp;transferencia{/if}.
				</p>
				<div
					style="margin-top:20px;display:flex;gap:8px;overflow-x:auto;scrollbar-width:none;padding:2px"
				>
					{#each chipsCat as c (c.c ?? '')}{@const on = c.c === categoria}<a
							href={enlaceCat(c.c)}
							style="flex:none;text-decoration:none;padding:9px 15px;border-radius:999px;background:{on
								? '#1E1B2E'
								: '#fff'};color:{on
								? '#fff'
								: '#1E1B2E'};font-size:13.5px;font-weight:600;white-space:nowrap;box-shadow:0 1px 2px rgba(30,27,46,.06)"
							>{c.t}<span style="opacity:.55;margin-left:5px">{c.n}</span></a
						>{/each}
				</div>
				<div
					style="margin-top:14px;display:flex;justify-content:flex-end;align-items:center;gap:10px;flex-wrap:wrap;font-size:13.5px"
				>
					<label style="position:relative;display:flex;align-items:center;gap:6px;font-weight:700"
						><span class="oculto">Ordenar</span><select
							value={orden}
							onchange={(e) => goto(enlaceCat(categoria, e.currentTarget.value), { replace: true })}
							style="appearance:none;border:0;background:transparent;padding:0 18px 0 0;font:700 13.5px 'Instrument Sans';color:#1E1B2E;cursor:pointer"
							>{#each ORDENES as [k, t] (k)}<option value={k}>{t}</option>{/each}</select
						><Icon
							icon={I.phCaretDownBold}
							size={12}
							style="position:absolute;right:0;pointer-events:none"
						/></label
					>
				</div>
				{#if catalog.length}
					<div
						style="margin-top:20px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(46%,220px),1fr));gap:12px"
					>
						{#each catalog as p (p.id)}{@render card(p)}{/each}
					</div>
					{#if listado.length > mostrar}
						<div style="margin-top:32px;display:flex;justify-content:center">
							<button
								type="button"
								onclick={() => (mostrar += 24)}
								style="padding:15px 26px;border:1.5px solid rgba(30,27,46,.18);border-radius:999px;background:transparent;color:#1E1B2E;font:700 15px 'Instrument Sans';cursor:pointer"
								>Ver más</button
							>
						</div>
					{/if}
				{:else}
					<p style="margin:40px 0;font-size:15.5px;color:#5A566B">
						Todavía no hay productos en esta categoría.
					</p>
				{/if}
			</section>
		{/if}

		<!-- ============ PRODUCTO ============ -->
		{#if pagina === 'producto' && producto && sel}
			<section
				style="padding:24px clamp(14px,3.5cqw,44px) 0;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,380px),1fr));gap:clamp(18px,4cqw,56px);align-items:start"
			>
				<div style="display:flex;flex-direction:column;gap:8px;min-width:0">
					<div
						style="position:relative;aspect-ratio:1;border-radius:28px;overflow:hidden;background:#EAE6EF"
					>
						{#if sel.foto}<img
								src={img(sel.foto, 1000)}
								alt={producto.nombre}
								style="width:100%;height:100%;object-fit:cover"
							/>{/if}{#if tachado || producto.etiqueta}<span
								style="position:absolute;top:14px;left:14px;padding:4px 9px;border-radius:999px;background:var(--acento, #5B3FD6);color:#fff;font-size:11px;font-weight:700"
								>{tachado
									? `−${Math.round((1 - sel.precio / tachado) * 100)}%`
									: producto.etiqueta}</span
							>{/if}
					</div>
					{#if sel.fotos.length > 1}<div
							style="display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:8px"
						>
							{#each sel.fotos as f, i (f)}<button
									type="button"
									aria-label="Ver la foto {i + 1}"
									aria-pressed={i === sel.fotoIdx}
									onclick={() => sel.verFoto(i)}
									style="aspect-ratio:1;padding:0;border:0;border-radius:16px;overflow:hidden;background:#EAE6EF;cursor:pointer;box-shadow:0 0 0 {i ===
									sel.fotoIdx
										? '2.5px'
										: '0px'} var(--acento, #5B3FD6)"
									><img
										src={img(f, 200, 200)}
										alt=""
										style="display:block;width:100%;height:100%;object-fit:cover"
									/></button
								>{/each}
						</div>{/if}
				</div>
				<div style="position:sticky;top:90px;min-width:0">
					{#if producto.categoria}<a
							href={enlaceCat(producto.categoria)}
							style="font-size:13px;font-weight:700;color:var(--acento, #5B3FD6);text-decoration:none"
							>{producto.categoria}</a
						>{/if}
					<h1
						style="margin:6px 0 0;font-family:'Syne',sans-serif;font-weight:800;font-size:clamp(36px,4.4cqw,54px);line-height:.98;letter-spacing:-.035em;overflow-wrap:anywhere"
					>
						{producto.nombre}
					</h1>
					{#if resenasProd.length}<div
							style="margin-top:10px;display:flex;align-items:center;gap:8px;font-size:13.5px"
						>
							<span style="display:flex;color:#E3A93B"
								>{#each estrellas(promedioProd) as e, i (i)}<Icon icon={e} size={14} />{/each}</span
							><b>{promedioProd.toLocaleString('es-AR')}</b><span style="color:#5A566B"
								>{resenasProd.length} {resenasProd.length === 1 ? 'reseña' : 'reseñas'}</span
							>
						</div>{/if}
					<div style="margin-top:18px;display:flex;align-items:baseline;gap:10px">
						<b style="font-size:30px;letter-spacing:-.02em"
							>{sel.desde ? 'Desde ' : ''}{fmt(sel.precio)}</b
						>{#if tachado}<s style="font-size:16px;color:#8B879A">{fmt(tachado)}</s>{/if}
					</div>
					<div style="margin-top:6px;display:flex;flex-direction:column;gap:3px;font-size:14px">
						{#if tienda.cobros.cuotas > 1}<span
								>{tienda.cobros.cuotas} cuotas sin interés de
								<b>{fmt(sel.precio / tienda.cobros.cuotas)}</b></span
							>{/if}{#if tienda.cobros.transferencia > 0}<span style="color:#2F7D4F;font-weight:700"
								>{fmt(sel.precio * (1 - tienda.cobros.transferencia / 100))} por transferencia</span
							>{/if}
					</div>
					{#each sel.opciones as o (o.nombre)}{@const swatch = o.valores.every(
							(x) => x.color || x.foto
						)}
						<div style="margin-top:20px">
							<div
								style="font-size:13.5px;margin-bottom:8px;color:{sel.aviso && !o.elegido
									? 'var(--acento, #5B3FD6)'
									: '#1E1B2E'}"
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
											style="width:40px;height:40px;padding:0;border:0;border-radius:50%;overflow:hidden;background:#EAE6EF;cursor:pointer;opacity:{x.disponible
												? 1
												: 0.4};box-shadow:inset 0 0 0 1px rgba(30,27,46,.12),0 0 0 {x.elegido
												? '2.5px'
												: '0px'} #F6F4F1,0 0 0 {x.elegido ? '5px' : '0px'} var(--acento, #5B3FD6)"
											>{#if x.color}<span
													style="display:block;width:100%;height:100%;background:{x.color}"
												></span>{:else}<img
													src={img(x.foto, 80, 80)}
													alt=""
													style="display:block;width:100%;height:100%;object-fit:cover"
												/>{/if}</button
										>{/each}
								</div>{:else}<div style="display:flex;gap:8px;flex-wrap:wrap">
									{#each o.valores as x (x.valor)}{@const pv =
											sel.opciones.length === 1
												? sel.variantes.find((v) => v.valores[o.nombre] === x.valor)
												: null}<button
											type="button"
											aria-pressed={x.elegido}
											disabled={!x.disponible && !x.elegido}
											onclick={() => sel.elegir(o.nombre, x.valor)}
											style="min-height:48px;padding:11px 16px;border-radius:14px;border:1.5px solid {x.elegido
												? 'var(--acento, #5B3FD6)'
												: 'rgba(30,27,46,.12)'};background:{x.elegido
												? '#F3F0FF'
												: '#fff'};font:600 14px 'Instrument Sans';color:{x.disponible
												? '#1E1B2E'
												: '#8B879A'};text-align:left;text-decoration:{x.disponible
												? 'none'
												: 'line-through'};cursor:{x.disponible ? 'pointer' : 'default'}"
											>{x.valor}{#if pv}<span
													style="display:block;font-size:12px;color:#5A566B;font-weight:500"
													>{fmt(pv.precio ?? producto.precio)}</span
												>{/if}</button
										>{/each}
								</div>{/if}
						</div>{/each}
					{#if sel.stock > 0 && sel.stock <= 3 && (!sel.conOpciones || sel.variante)}<div
							style="margin-top:14px;font-size:13px;font-weight:700;color:var(--acento, #5B3FD6)"
						>
							{sel.stock === 1 ? 'Queda 1' : `Quedan ${sel.stock}`}{#if sel.variante}
								en {sel.texto.toLowerCase()}{/if}
						</div>{/if}
					<div style="margin-top:20px;display:flex;gap:8px;flex-wrap:wrap">
						{#if sel.stock > 0}<span
								style="display:flex;align-items:center;border-radius:999px;background:#fff;box-shadow:0 1px 2px rgba(30,27,46,.08)"
								><button
									type="button"
									aria-label="Uno menos"
									disabled={cantidad <= 1}
									onclick={() => (cantidad = Math.max(1, cantidad - 1))}
									style="width:44px;height:48px;display:grid;place-items:center;border:0;background:transparent;color:#1E1B2E;cursor:pointer"
									><Icon icon={I.phMinusBold} size={14} /></button
								><b style="width:22px;text-align:center">{cantidad}</b><button
									type="button"
									aria-label="Uno más"
									disabled={cantidad >= sel.stock}
									onclick={() => (cantidad = Math.min(sel.stock, cantidad + 1))}
									style="width:44px;height:48px;display:grid;place-items:center;border:0;background:transparent;color:#1E1B2E;cursor:pointer"
									><Icon icon={I.phPlusBold} size={14} /></button
								></span
							>{/if}<button
							type="button"
							onclick={agregar}
							disabled={sel.stock === 0 || tienda.vitrina}
							style="flex:1;min-width:180px;padding:15px;border:0;border-radius:999px;background:#1E1B2E;color:#fff;font:700 15px 'Instrument Sans';cursor:pointer"
							>{sel.boton('Agregar al carrito')}</button
						>
					</div>
					{#if whatsapp}<a
							href={wa(
								`¡Hola! Quiero comprar ${producto.nombre}${sel.variante ? ` (${sel.texto})` : ''} (${fmt(sel.precio)}).`
							)}
							target="_blank"
							rel="noopener"
							style="margin-top:8px;display:flex;justify-content:center;align-items:center;gap:8px;padding:14px;border-radius:999px;border:1.5px solid rgba(30,27,46,.14);font-size:14.5px;font-weight:700;color:#1E1B2E;text-decoration:none"
							><Icon icon={I.logosWhatsappIcon} size={17} />Consultar por WhatsApp</a
						>{/if}
					<CompartirProducto {tienda} {producto} {tema} />
					{#if tienda.envios.retiro}<div
							style="margin-top:16px;padding:14px;border-radius:20px;background:#EEE9FF;display:flex;gap:12px;align-items:center"
						>
							<Icon
								icon={I.phStorefrontFill}
								size={26}
								style="color:var(--acento, #5B3FD6);flex:none"
							/>
							<div style="flex:1;font-size:13.5px;line-height:1.4">
								<b>Retiralo en el local</b>
								{tienda.direccion ? `en ${tienda.direccion}` : ''} y no pagás&nbsp;envío.
							</div>
						</div>{/if}
					{#if envia(tienda.envios)}<div
							style="margin-top:12px;padding:14px;border-radius:20px;background:#fff;box-shadow:0 1px 2px rgba(30,27,46,.06)"
						>
							<label for="cp-ficha" style="display:block;font-size:13.5px;font-weight:700"
								>Envío a domicilio</label
							>
							<div style="margin-top:8px;display:flex;gap:6px">
								<input
									id="cp-ficha"
									bind:value={cp}
									inputmode="numeric"
									maxlength="4"
									placeholder="Código postal"
									onkeydown={(e) => e.key === 'Enter' && calcular()}
									style="flex:1;min-width:0;padding:10px 12px;border:0;border-radius:12px;background:#F6F4F1;font:400 14px 'Instrument Sans';color:#1E1B2E"
								/><button
									type="button"
									onclick={calcular}
									style="padding:10px 14px;border:0;border-radius:12px;background:#1E1B2E;color:#fff;font:700 13.5px 'Instrument Sans';cursor:pointer"
									>Calcular</button
								>
							</div>
							{#if cotizado}<div
									style="margin-top:10px;display:flex;flex-direction:column;gap:6px;font-size:13.5px"
								>
									{#each cotizado as o (o.id)}<span
											style="display:flex;justify-content:space-between;gap:10px"
											><span>{o.nombre} · {o.detalle}</span><b
												style="white-space:nowrap;color:{o.precio ? '#1E1B2E' : '#2F7D4F'}"
												>{o.precio ? fmt(o.precio) : 'Gratis'}</b
											></span
										>{/each}
								</div>{/if}
						</div>{/if}
					<div
						style="margin-top:12px;padding:4px 14px;border-radius:20px;background:#fff;box-shadow:0 1px 2px rgba(30,27,46,.06)"
					>
						{#each acc as a, i (a.t)}<div
								style="padding:12px 0;{i ? 'border-top:1px solid rgba(30,27,46,.06)' : ''}"
							>
								<button
									type="button"
									onclick={() => (abierto = a.open ? '' : a.t)}
									aria-expanded={a.open}
									style="width:100%;display:flex;justify-content:space-between;align-items:center;gap:10px;padding:0;border:0;background:transparent;font:700 13.5px 'Instrument Sans';color:#1E1B2E;cursor:pointer;text-align:left"
									>{a.t}<Icon icon={a.i} size={13} /></button
								>{#if a.open}<p style="margin:8px 0 0;font-size:14px;line-height:1.6;color:#4A4659">
										{a.d}
									</p>{/if}
							</div>{/each}
					</div>
					<div
						style="margin-top:12px;padding:18px 14px;border-radius:20px;background:#fff;box-shadow:0 1px 2px rgba(30,27,46,.06)"
					>
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
				<section style="padding:clamp(48px,7cqw,90px) clamp(14px,3.5cqw,44px)">
					<h2
						style="margin:0;font-family:'Syne',sans-serif;font-weight:800;font-size:clamp(34px,4.6cqw,58px);line-height:.95;letter-spacing:-.035em"
					>
						Combinalo con
					</h2>
					<div
						style="margin-top:22px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(46%,220px),1fr));gap:12px"
					>
						{#each related as p (p.id)}{@render card(p)}{/each}
					</div>
				</section>
			{:else}
				<div style="height:clamp(48px,7cqw,90px)"></div>
			{/if}
		{/if}

		<!-- ============ CARRITO ============ -->
		{#if pagina === 'carrito'}
			<section style="padding:24px clamp(14px,3.5cqw,44px) clamp(48px,7cqw,90px)">
				<h1
					style="margin:0;font-family:'Syne',sans-serif;font-weight:800;font-size:clamp(40px,6cqw,72px);line-height:.95;letter-spacing:-.04em"
				>
					Tu carrito
				</h1>
				{#if !lineas.length}
					<p style="margin:16px 0 0;font-size:15.5px;color:#5A566B">Todavía no sumaste nada.</p>
					<a
						href="{base}/catalogo"
						style="display:inline-flex;align-items:center;gap:8px;margin-top:18px;padding:16px 26px;border-radius:999px;background:var(--acento, #5B3FD6);color:#fff;font:700 15px 'Instrument Sans';text-decoration:none;box-shadow:0 14px 30px -12px rgba(91,63,214,.7)"
						><Icon icon={I.phShoppingBagFill} size={17} />Ver productos</a
					>
				{:else}
					<div
						style="margin-top:22px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,340px),1fr));gap:clamp(16px,3cqw,36px);align-items:start"
					>
						<div style="display:flex;flex-direction:column;gap:10px;min-width:0">
							{#each lineas as c (c.clave)}<div
									style="display:flex;gap:14px;padding:12px;border-radius:22px;background:#fff;box-shadow:0 1px 2px rgba(30,27,46,.06)"
								>
									{#if c.producto.foto}<img
											src={img(c.producto.foto, 200, 200)}
											alt=""
											style="width:88px;height:88px;border-radius:16px;object-fit:cover;flex:none;background:#EAE6EF"
										/>{/if}
									<div style="flex:1;min-width:0;display:flex;flex-direction:column">
										<div style="display:flex;justify-content:space-between;gap:10px">
											<a
												href="{base}/producto/{c.producto.slug ?? c.id}"
												style="font-size:15px;font-weight:600;color:#1E1B2E;text-decoration:none"
												>{c.producto.nombre}</a
											><b style="white-space:nowrap">{fmt(c.subtotal)}</b>
										</div>
										<span style="font-size:13px;color:#5A566B"
											>{c.variante ? `${c.variante} · ` : ''}{fmt(c.producto.precio)} c/u</span
										>
										<span style="flex:1;min-height:10px"></span>
										<div style="display:flex;justify-content:space-between;align-items:center">
											<span
												style="display:flex;align-items:center;gap:4px;padding:0 4px;border-radius:999px;background:#F6F4F1"
												><button
													type="button"
													aria-label="Uno menos"
													onclick={() => carrito.fijar(c.clave, c.cantidad - 1)}
													style="width:30px;height:30px;display:grid;place-items:center;border:0;background:transparent;color:#1E1B2E;cursor:pointer"
													><Icon icon={I.phMinusBold} size={12} /></button
												><b style="font-size:14px;min-width:16px;text-align:center">{c.cantidad}</b
												><button
													type="button"
													aria-label="Uno más"
													disabled={c.cantidad >= c.producto.stock}
													onclick={() => carrito.fijar(c.clave, c.cantidad + 1)}
													style="width:30px;height:30px;display:grid;place-items:center;border:0;background:transparent;color:#1E1B2E;cursor:pointer"
													><Icon icon={I.phPlusBold} size={12} /></button
												></span
											><button
												type="button"
												aria-label="Quitar {c.producto.nombre}"
												onclick={() => carrito.fijar(c.clave, 0)}
												style="display:grid;padding:4px;border:0;background:transparent;color:#8B879A;cursor:pointer"
												><Icon icon={I.phTrash} size={18} /></button
											>
										</div>
									</div>
								</div>{/each}
							{#if tienda.envios.gratisDesde}
								<div
									style="padding:16px;border-radius:22px;background:var(--acento, #5B3FD6);color:#fff;display:flex;gap:14px;align-items:center;flex-wrap:wrap"
								>
									<Icon icon={I.phTruck} size={30} />
									<div style="flex:1;min-width:200px">
										<div style="font-size:15px;font-weight:700">
											{#if falta}Te faltan {fmt(falta)} para el envío gratis{:else}Tenés envío
												gratis{/if}
										</div>
										<div
											style="margin-top:8px;height:6px;border-radius:999px;background:rgba(255,255,255,.25);overflow:hidden"
										>
											<div
												style="width:{Math.min(
													100,
													(subtotal / tienda.envios.gratisDesde) * 100
												)}%;height:100%;background:#fff"
											></div>
										</div>
									</div>
								</div>
							{/if}
							{#if upsell.length}
								<div style="margin-top:10px;font-size:15px;font-weight:700">Combinalo con</div>
								<div style="display:flex;gap:10px;overflow-x:auto;scrollbar-width:none">
									{#each upsell as p (p.id)}<div
											style="flex:none;width:160px;padding:8px;border-radius:20px;background:#fff;box-shadow:0 1px 2px rgba(30,27,46,.06)"
										>
											<a
												href="{base}/producto/{p.slug ?? p.id}"
												style="display:block;aspect-ratio:1;border-radius:14px;overflow:hidden;background:#EAE6EF"
												>{#if p.foto}<img
														src={img(p.foto, 300, 300)}
														alt=""
														style="width:100%;height:100%;object-fit:cover"
													/>{/if}</a
											>
											<div style="padding:8px 4px 2px;font-size:13.5px;font-weight:600">
												{p.nombre}
											</div>
											<div
												style="padding:0 4px;display:flex;justify-content:space-between;align-items:center"
											>
												<b style="font-size:14px">{fmt(p.precio)}</b><button
													type="button"
													aria-label="Sumar {p.nombre}"
													onclick={() => sumar(p.id)}
													style="display:grid;padding:0;border:0;background:transparent;color:#1E1B2E;cursor:pointer"
													><Icon icon={I.phPlusCircleFill} size={26} /></button
												>
											</div>
										</div>{/each}
								</div>
							{/if}
						</div>
						<div
							style="position:sticky;top:90px;padding:22px;border-radius:28px;background:#fff;box-shadow:0 20px 40px -24px rgba(30,27,46,.25);display:flex;flex-direction:column;gap:11px;font-size:14px"
						>
							<div style="font-family:'Syne',sans-serif;font-weight:700;font-size:24px">
								Resumen
							</div>
							<span style="display:flex;justify-content:space-between"
								><span>Subtotal</span><span>{fmt(subtotal)}</span></span
							>
							<span style="display:flex;justify-content:space-between;gap:10px;color:#5A566B"
								><span>Envío</span><span>Se calcula en el próximo paso</span></span
							>
							<div style="height:1px;background:rgba(30,27,46,.08)"></div>
							<span style="display:flex;justify-content:space-between;align-items:baseline"
								><b>Total</b><b style="font-size:24px">{fmt(subtotal)}</b></span
							>
							{#if tienda.cobros.cuotas > 1}<span style="color:#5A566B"
									>{tienda.cobros.cuotas} cuotas sin interés de {fmt(
										subtotal / tienda.cobros.cuotas
									)}</span
								>{/if}
							{#if tienda.cobros.transferencia > 0}<span style="color:#2F7D4F;font-weight:700"
									>{fmt(subtotal * (1 - tienda.cobros.transferencia / 100))} pagando por transferencia</span
								>{/if}
							<a
								href="{base}/checkout"
								style="margin-top:4px;padding:16px;border-radius:999px;background:#1E1B2E;color:#fff;font:700 15px 'Instrument Sans';text-align:center;text-decoration:none"
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
					<p style="margin:0;font-size:15.5px;color:#5A566B">Tu carrito está vacío.</p>
					<a
						href="{base}/catalogo"
						style="display:inline-block;margin-top:16px;color:var(--acento, #5B3FD6);font-weight:700"
						>Ver productos</a
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
					style="padding:24px clamp(14px,3.5cqw,44px) clamp(48px,7cqw,90px);display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,360px),1fr));gap:clamp(16px,3cqw,40px);align-items:start"
				>
					<input type="hidden" name="carrito" value={JSON.stringify(carrito.lineas)} />
					<input type="hidden" name="envio" value={envio ?? ''} />
					<input type="hidden" name="pago" value={pago} />
					<div style="display:flex;flex-direction:column;gap:22px;min-width:0">
						<h1
							style="margin:0;font-family:'Syne',sans-serif;font-weight:800;font-size:clamp(36px,5cqw,60px);line-height:.95;letter-spacing:-.04em"
						>
							Confirmá y pagá
						</h1>
						<div>
							<div style="font-size:15px;font-weight:700;margin-bottom:10px">Tus datos</div>
							<div
								style="display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));gap:8px"
							>
								{#each [['email', 'Email', 'email', 'email'], ['nombre', 'Nombre y apellido', 'text', 'name'], ['telefono', 'Teléfono', 'tel', 'tel'], ['dni', 'DNI', 'text', 'off'], ...camposFactura(tienda)] as [n, l, tipo, ac] (n)}
									<label
										style="display:block;padding:11px 14px;border-radius:16px;background:#fff;box-shadow:inset 0 0 0 1.5px rgba(30,27,46,.1)"
										><span style="display:block;font-size:11.5px;color:#8B879A">{l}</span><input
											name={n}
											type={tipo}
											autocomplete={ac as 'email'}
											required={!OPCIONALES.has(n)}
											maxlength="120"
											style="width:100%;margin-top:2px;padding:0;border:0;background:transparent;font:400 14.5px 'Instrument Sans';color:#1E1B2E;outline:none"
										/></label
									>
								{/each}
							</div>
							<!-- Sin bloque de newsletter en el diseño de esta plantilla: la casilla del checkout (§50). -->
							<label
								style="margin-top:10px;display:flex;align-items:center;gap:10px;min-height:44px;font:400 14.5px 'Instrument Sans';color:#1E1B2E;cursor:pointer"
								><input
									type="checkbox"
									name="novedades"
									style="flex:none;width:18px;height:18px;margin:0;accent-color:var(--acento, #5B3FD6)"
								/>Quiero recibir novedades</label
							>
						</div>
						<div>
							<div style="font-size:15px;font-weight:700;margin-bottom:10px">Entrega</div>
							<div style="display:grid;grid-template-columns:minmax(0,130px) minmax(0,1fr);gap:8px">
								<label
									style="display:block;padding:11px 14px;border-radius:16px;background:#fff;box-shadow:inset 0 0 0 1.5px rgba(30,27,46,.1)"
									><span style="display:block;font-size:11.5px;color:#8B879A">Código postal</span
									><input
										name="cp"
										bind:value={cp}
										inputmode="numeric"
										maxlength="4"
										style="width:100%;margin-top:2px;padding:0;border:0;background:transparent;font:400 14.5px 'Instrument Sans';color:#1E1B2E;outline:none"
									/></label
								>
								<label
									style="display:block;padding:11px 14px;border-radius:16px;background:#fff;box-shadow:inset 0 0 0 1.5px rgba(30,27,46,.1)"
									><span style="display:block;font-size:11.5px;color:#8B879A"
										>Dirección (calle, número, piso)</span
									><input
										name="direccion"
										maxlength="160"
										autocomplete="street-address"
										required={envio !== 'retiro'}
										style="width:100%;margin-top:2px;padding:0;border:0;background:transparent;font:400 14.5px 'Instrument Sans';color:#1E1B2E;outline:none"
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
										style="display:flex;align-items:center;gap:12px;padding:15px;border-radius:18px;border:1.5px solid {on
											? 'var(--acento, #5B3FD6)'
											: 'rgba(30,27,46,.1)'};background:{on
											? '#F3F0FF'
											: '#fff'};text-align:left;cursor:pointer;font-family:inherit;color:#1E1B2E"
										><span
											style="width:20px;height:20px;border-radius:50%;border:2px solid {on
												? 'var(--acento, #5B3FD6)'
												: 'rgba(30,27,46,.25)'};display:grid;place-items:center;flex:none"
											>{#if on}<span
													style="width:10px;height:10px;border-radius:50%;background:var(--acento, #5B3FD6)"
												></span>{/if}</span
										><span style="flex:1;min-width:0"
											><span style="display:block;font-size:14.5px;font-weight:600">{s.nombre}</span
											><span style="display:block;font-size:12.5px;color:#5A566B">{s.detalle}</span
											></span
										><b
											style="font-size:14px;white-space:nowrap;color:{s.precio
												? '#1E1B2E'
												: '#2F7D4F'}">{s.precio ? fmt(s.precio) : 'Gratis'}</b
										></button
									>{:else}<span style="font-size:13.5px;color:#5A566B"
										>Poné tu código postal para ver cómo te llega.</span
									>{/each}
							</div>
							<EnvioCorreo {tienda} {cp} {envio} {opciones} />
						</div>
						<div>
							<div style="font-size:15px;font-weight:700;margin-bottom:10px">Pago</div>
							<div
								style="display:flex;flex-direction:column;gap:8px"
								role="radiogroup"
								aria-label="Forma de pago"
							>
								{#each pagos as s (s.id)}{@const on = pago === s.id}<button
										type="button"
										role="radio"
										aria-checked={on}
										onclick={() => (pago = s.id)}
										style="display:flex;align-items:center;gap:12px;padding:15px;border-radius:18px;border:1.5px solid {on
											? 'var(--acento, #5B3FD6)'
											: 'rgba(30,27,46,.1)'};background:{on
											? '#F3F0FF'
											: '#fff'};text-align:left;cursor:pointer;font-family:inherit;color:#1E1B2E"
										><span
											style="width:20px;height:20px;border-radius:50%;border:2px solid {on
												? 'var(--acento, #5B3FD6)'
												: 'rgba(30,27,46,.25)'};display:grid;place-items:center;flex:none"
											>{#if on}<span
													style="width:10px;height:10px;border-radius:50%;background:var(--acento, #5B3FD6)"
												></span>{/if}</span
										><Icon icon={s.i} size={22} style="color:{s.c};flex:none" /><span
											style="flex:1;min-width:0"
											><span style="display:block;font-size:14.5px;font-weight:600">{s.t}</span
											><span style="display:block;font-size:12.5px;color:#5A566B">{s.d}</span></span
										></button
									>{/each}
							</div>
						</div>
						{#if error}<div
								role="alert"
								style="padding:12px 14px;border-radius:16px;background:#FDECEC;color:#B3261E;font-size:14px"
							>
								{error}
							</div>{/if}
						<button
							disabled={!envio || !pagos.length || enviando || tienda.vitrina}
							style="padding:17px;border:0;border-radius:999px;background:var(--acento, #5B3FD6);color:#fff;font:700 15.5px 'Instrument Sans';cursor:pointer;display:flex;justify-content:center;align-items:center;gap:8px;box-shadow:0 14px 30px -12px rgba(91,63,214,.7)"
							><Icon icon={I.phLockSimpleFill} size={16} />{tienda.vitrina
								? 'Es una tienda de ejemplo'
								: !pagos.length
									? 'Esta tienda todavía no tiene medios de pago'
									: envio
										? `Pagar ${fmt(total)}`
										: 'Elegí cómo te llega'}</button
						>
					</div>
					<div style="position:sticky;top:90px;display:flex;flex-direction:column;gap:10px">
						<div
							style="padding:18px;border-radius:24px;background:#fff;box-shadow:0 1px 2px rgba(30,27,46,.06)"
						>
							<div
								style="display:flex;align-items:center;gap:8px;font-size:12.5px;font-weight:700;letter-spacing:.08em;text-transform:uppercase;color:#5A566B"
							>
								<Icon icon={I.phShoppingBagFill} size={15} />Tu pedido
							</div>
							<div style="margin-top:10px;display:flex;flex-direction:column;gap:10px">
								{#each lineas as c (c.clave)}<div style="display:flex;gap:10px;align-items:center">
										{#if c.producto.foto}<img
												src={img(c.producto.foto, 120, 120)}
												alt=""
												style="width:46px;height:46px;border-radius:12px;object-fit:cover;flex:none;background:#EAE6EF"
											/>{/if}<span style="flex:1;min-width:0;font-size:14px"
											>{c.producto.nombre}{#if c.cantidad > 1}<span style="color:#8B879A"
													>&nbsp;× {c.cantidad}</span
												>{/if}{#if c.variante}<span
													style="display:block;font-size:12.5px;color:#8B879A">{c.variante}</span
												>{/if}</span
										><b style="font-size:14px;white-space:nowrap">{fmt(c.subtotal)}</b>
									</div>{/each}
							</div>
							<div
								style="margin-top:14px;padding-top:12px;border-top:1px solid rgba(30,27,46,.08);display:flex;flex-direction:column;gap:8px;font-size:14px"
							>
								<span style="display:flex;justify-content:space-between"
									><span>Subtotal</span><span>{fmt(subtotal)}</span></span
								><span style="display:flex;justify-content:space-between"
									><span>Envío</span><span
										>{elegida ? (cuenta.envio ? fmt(cuenta.envio) : 'Gratis') : '—'}</span
									></span
								>{#if cuenta.descCupon}<span
										style="display:flex;justify-content:space-between;color:#2F7D4F;font-weight:700"
										><span>Cupón {cupon.aplicado}</span><span>− {fmt(cuenta.descCupon)}</span></span
									>{/if}{#if descuento}<span
										style="display:flex;justify-content:space-between;color:#2F7D4F;font-weight:700"
										><span>Transferencia −{tienda.cobros.transferencia}%</span><span
											>− {fmt(descuento)}</span
										></span
									>{/if}
								<div style="margin:6px 0 2px">
									<CampoCupon
										{cupon}
										lineas={carrito.lineas}
										tinta="#1E1B2E"
										suave="#8B879A"
										borde="rgba(30,27,46,.14)"
										fondo="#F6F4F1"
										acento="var(--acento, #5B3FD6)"
										sobreAcento="#fff"
										radio="12px"
										fuente="'Instrument Sans',sans-serif"
									/>
								</div>
							</div>
						</div>
						<div
							style="display:flex;justify-content:space-between;align-items:baseline;padding:0 6px"
						>
							<b>Pagás hoy</b><b style="font-size:24px">{fmt(total)}</b>
						</div>
						<div
							style="padding:12px;border-radius:16px;background:#1E1B2E;color:#F6F4F1;font-size:13px;line-height:1.45;display:flex;gap:10px;align-items:center"
						>
							<Icon icon={I.phShieldCheckFill} size={20} style="color:#A996FF;flex:none" />Compra
							protegida · 10 días para arrepentirte
						</div>
					</div>
				</form>
			{/if}
		{/if}

		<!-- ============ LISTO (no está en el diseño: con las piezas de "Tu reserva" y el checkout) ============ -->
		{#if pagina === 'listo' && pedido}
			<section
				style="padding:clamp(40px,6cqw,80px) clamp(14px,3.5cqw,44px);max-width:720px;margin:0 auto;text-align:center"
			>
				<span
					style="display:inline-grid;place-items:center;width:64px;height:64px;border-radius:50%;background:var(--acento, #5B3FD6);color:#fff;box-shadow:0 14px 30px -12px rgba(91,63,214,.7)"
					><Icon icon={I.phCheckBold} size={28} /></span
				>
				<h1
					style="margin:20px 0 0;font-family:'Syne',sans-serif;font-weight:800;font-size:clamp(36px,5cqw,60px);line-height:.95;letter-spacing:-.04em;overflow-wrap:anywhere"
				>
					¡Gracias, {pedido.nombre}!
				</h1>
				<p style="margin:14px 0 0;font-size:16px;line-height:1.55;color:#5A566B">
					Recibimos tu pedido <b style="color:#1E1B2E">#{pedido.numero}</b> por
					<b style="color:#1E1B2E">{fmt(pedido.total)}</b>. Te escribimos por mail con
					cada&nbsp;novedad.
				</p>
				{#if pedido.pago === 'transferencia'}
					<div
						style="margin-top:24px;padding:22px;border-radius:28px;background:#1E1B2E;color:#F6F4F1;text-align:left;font-size:14.5px;line-height:1.6"
					>
						<b>Transferí {fmt(pedido.total)}</b>
						{#if pedido.alias}al {pedido.esCbu ? 'CBU/CVU' : 'alias'}
							<b style="color:#A996FF">{pedido.alias}</b>{:else}con los datos que te mandamos por
							mail{/if} y mandanos el comprobante{#if whatsapp}&nbsp;<a
								href={wa(`¡Hola! Te mando el comprobante del pedido #${pedido.numero}.`)}
								target="_blank"
								rel="noopener"
								style="color:#A996FF">por WhatsApp</a
							>{:else}&nbsp;respondiendo el mail del pedido{/if}. Lo preparamos apenas se acredite.
					</div>
				{/if}
				<a
					href={base || '/'}
					style="display:inline-block;margin-top:28px;padding:16px 26px;border-radius:999px;background:#1E1B2E;color:#fff;font:700 15px 'Instrument Sans';text-decoration:none"
					>Volver a la tienda</a
				>
				{#if pedido.estado === 'Entregado'}
					<div style="margin-top:clamp(36px,5cqw,56px)">
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
			<section
				style="padding:clamp(18px,3cqw,40px) clamp(14px,3.5cqw,44px) 0;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,360px),1fr));gap:clamp(14px,2cqw,24px);align-items:stretch"
			>
				{#if tienda.portada.foto}<div
						style="position:relative;min-height:420px;border-radius:32px;overflow:hidden;background:#CFC8DA"
					>
						<img
							src={img(tienda.portada.foto, 1100)}
							alt=""
							style="position:absolute;inset:0;width:100%;height:100%;object-fit:cover"
						/>
					</div>{/if}
				<div
					style="display:flex;flex-direction:column;justify-content:center;gap:18px;padding:clamp(8px,2cqw,24px) 0;min-width:0"
				>
					<h1
						style="margin:0;font-family:'Syne',sans-serif;font-weight:800;font-size:clamp(36px,4.6cqw,66px);line-height:.94;letter-spacing:-.04em;overflow-wrap:anywhere"
					>
						{titulo[0]}{#if titulo[1]}<span style="color:var(--acento, #5B3FD6)"
								>{` ${titulo[1]}`}</span
							>{/if}
					</h1>
					<p style="margin:0;font-size:16px;line-height:1.6;color:#5A566B;max-width:500px">
						{tienda.nombre}{tienda.ciudad ? `, desde ${tienda.ciudad}` : ''}. {comoLlega(
							tienda.envios
						)}
					</p>
					{#if tienda.productos.length || tienda.cobros.cuotas > 1 || tienda.promedio}<div
							style="display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:10px"
						>
							{#if tienda.productos.length}<div
									style="padding:14px;border-radius:20px;background:#fff;box-shadow:0 1px 2px rgba(30,27,46,.06)"
								>
									<div
										style="font-family:'Syne',sans-serif;font-weight:800;font-size:clamp(26px,3cqw,36px);letter-spacing:-.03em"
									>
										{tienda.productos.length}
									</div>
									<div style="font-size:12.5px;color:#5A566B;margin-top:2px">productos</div>
								</div>{/if}
							{#if tienda.cobros.cuotas > 1}<div
									style="padding:14px;border-radius:20px;background:#fff;box-shadow:0 1px 2px rgba(30,27,46,.06)"
								>
									<div
										style="font-family:'Syne',sans-serif;font-weight:800;font-size:clamp(26px,3cqw,36px);letter-spacing:-.03em"
									>
										{tienda.cobros.cuotas}
									</div>
									<div style="font-size:12.5px;color:#5A566B;margin-top:2px">
										cuotas sin interés
									</div>
								</div>{/if}
							{#if tienda.promedio}<div
									style="padding:14px;border-radius:20px;background:#fff;box-shadow:0 1px 2px rgba(30,27,46,.06)"
								>
									<div
										style="font-family:'Syne',sans-serif;font-weight:800;font-size:clamp(26px,3cqw,36px);letter-spacing:-.03em"
									>
										{tienda.promedio.toLocaleString('es-AR')}
									</div>
									<div style="font-size:12.5px;color:#5A566B;margin-top:2px">
										de promedio en reseñas
									</div>
								</div>{/if}
						</div>{/if}
				</div>
			</section>
			<section
				style="padding:clamp(40px,6cqw,80px) clamp(14px,3.5cqw,44px);display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,320px),1fr));gap:12px"
			>
				{#if tienda.direccion || tienda.ciudad}
					<div
						style="padding:22px;border-radius:28px;background:#fff;box-shadow:0 1px 2px rgba(30,27,46,.06);display:flex;flex-direction:column;gap:14px"
					>
						<div style="font-family:'Syne',sans-serif;font-weight:700;font-size:24px">
							Dónde estamos
						</div>
						<div style="display:flex;flex-direction:column;gap:10px;font-size:14.5px">
							<span style="display:flex;gap:10px"
								><Icon
									icon={I.phMapPinFill}
									size={18}
									style="color:var(--acento, #5B3FD6);flex:none"
								/>{tienda.direccion || tienda.ciudad}</span
							>{#if tienda.envios.retiro}<span style="display:flex;gap:10px"
									><Icon
										icon={I.phStorefrontFill}
										size={18}
										style="color:var(--acento, #5B3FD6);flex:none"
									/>Retiro en el local sin costo</span
								>{/if}
						</div>
					</div>
				{/if}
				{#if T}
					<div
						style="padding:22px;border-radius:28px;background:#fff;box-shadow:0 1px 2px rgba(30,27,46,.06)"
					>
						<div style="font-family:'Syne',sans-serif;font-weight:700;font-size:24px">Horarios</div>
						<div style="margin-top:12px;display:flex;flex-direction:column">
							{#each horas as h (h.dias)}<div
									style="display:flex;justify-content:space-between;gap:10px;padding:11px 0;border-bottom:1px solid rgba(30,27,46,.06);font-size:14.5px;font-weight:{h.hoy
										? 700
										: 500};color:{h.cerrado ? '#8B879A' : '#1E1B2E'}"
								>
									<span>{h.dias}</span><span style="text-align:right">{h.horas}</span>
								</div>{/each}
						</div>
					</div>
				{/if}
				<div
					style="padding:22px;border-radius:28px;background:#fff;box-shadow:0 1px 2px rgba(30,27,46,.06)"
				>
					<div style="font-family:'Syne',sans-serif;font-weight:700;font-size:24px">
						Envíos y pagos
					</div>
					<div style="margin-top:12px;display:flex;flex-direction:column">
						{#each [...(tienda.envios.gratisDesde ? [['Envío gratis', `Desde ${fmt(tienda.envios.gratisDesde)}`]] : []), envia(tienda.envios) || !tienda.envios.retiro ? ['Envíos', entregaCorta(tienda.envios)] : ['Retiro', 'En el local'], ...(tienda.cobros.cuotas > 1 ? [['Tarjetas', `${tienda.cobros.cuotas} cuotas sin interés`]] : []), ...(tienda.cobros.transferencia > 0 ? [['Transferencia', `${tienda.cobros.transferencia}% off`]] : [])] as [d, h] (d)}<div
								style="display:flex;justify-content:space-between;gap:10px;padding:11px 0;border-bottom:1px solid rgba(30,27,46,.06);font-size:14.5px;font-weight:500"
							>
								<span>{d}</span><span style="text-align:right">{h}</span>
							</div>{/each}
					</div>
				</div>
				{#if whatsapp}
					<div
						style="padding:22px;border-radius:28px;background:#1E1B2E;color:#F6F4F1;display:flex;flex-direction:column;gap:10px"
					>
						<div style="font-family:'Syne',sans-serif;font-weight:700;font-size:24px">
							Escribinos
						</div>
						<a
							href={wa('¡Hola!')}
							target="_blank"
							rel="noopener"
							style="display:flex;align-items:center;gap:10px;padding:14px;border-radius:16px;background:#25D366;color:#fff;font-weight:700;text-decoration:none"
							><Icon icon={I.phWhatsappLogoFill} size={22} />{tienda.whatsapp}</a
						>
					</div>
				{/if}
			</section>
		{/if}

		<!-- footer -->
		<footer
			style="padding:clamp(40px,6cqw,72px) clamp(14px,3.5cqw,44px) 26px;background:#1E1B2E;color:#F6F4F1"
		>
			<div
				style="display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,200px),1fr));gap:28px"
			>
				<div style="min-width:0">
					<div
						style="font-family:'Syne',sans-serif;font-weight:800;font-size:34px;letter-spacing:-.02em;overflow-wrap:anywhere"
					>
						{tienda.nombre}
					</div>
					<p style="margin:10px 0 0;font-size:14px;color:#B3AEC6;line-height:1.55;max-width:300px">
						{#if tienda.cobros.cuotas > 1}{tienda.cobros.cuotas} cuotas sin interés{#if tienda.cobros.transferencia > 0}&nbsp;y
								{tienda.cobros.transferencia}% off por transferencia{/if}{:else if tienda.cobros.transferencia > 0}{tienda
								.cobros.transferencia}% off por transferencia{:else}Compra protegida{/if}{tienda.direccion ||
						tienda.ciudad
							? `. ${tienda.direccion || tienda.ciudad}`
							: ''}.
					</p>
				</div>
				{#each foot as f (f.t)}<div
						style="display:flex;flex-direction:column;gap:9px;font-size:14px"
					>
						<b
							style="font-size:12px;letter-spacing:.1em;text-transform:uppercase;color:#A996FF;margin-bottom:4px"
							>{f.t}</b
						>{#each f.l as l (l.t)}<a href={l.href} style="color:#DAD6E6;text-decoration:none"
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
				style="margin-top:36px;padding-top:16px;border-top:1px solid rgba(246,244,241,.1);display:flex;justify-content:space-between;gap:14px;flex-wrap:wrap;font-size:12.5px;color:#8B879A"
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
	.sube:hover {
		transform: translateY(-2px);
	}
	.oculto {
		position: absolute;
		width: 1px;
		height: 1px;
		overflow: hidden;
		clip: rect(0 0 0 0);
		white-space: nowrap;
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
