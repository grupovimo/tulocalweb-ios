<script lang="ts">
	import { camposFactura, OPCIONALES } from '#lib/tienda/factura.ts';
	import { enhance } from '$app/forms';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { untrack } from 'svelte';
	import { Reserva } from '../reserva.svelte.ts';
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
		envia,
		resumenMoto,
		textoEnvios
	} from '../envio.svelte.ts';
	import { Seleccion, tarjetaDe } from '../seleccion.svelte.ts';
	import { partirTitulo } from '../titulo.ts';
	import type { OpcionEnvio, ProductoPublico, TiendaPublica } from '../tipos.ts';

	/**
	 * Plantilla Miga (design/Plantilla Miga.dc.html), con los datos de la tienda en lugar de los de
	 * ejemplo. Lo que el diseño muestra sin datos atrás (cronograma del horno, cuenta regresiva,
	 * turnos de retiro, combos, suscripción, variantes, cupón, horarios, seguidores) no se inventa:
	 * esas partes no se dibujan, y sus cajas se reusan para lo que la tienda sí tiene.
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
	let ancho = $state(page.data.movil ? 390 : 1200);
	const nar = $derived(ancho < 820);
	const wide = $derived(!nar);

	const fmt = (n: number) => '$\u00a0' + Math.round(n).toLocaleString('es-AR');
	const img = (f: string | null, w: number, h?: number) =>
		!f
			? null
			: /^https?:\/\//.test(f)
				? f
				: `https://images.unsplash.com/photo-${f}?auto=format&fit=crop&w=${w}${h ? '&h=' + h : ''}&q=72`;
	const porId = $derived(new Map(tienda.productos.map((p) => [p.id, p])));
	// La pastilla de la foto: sin stock, oferta o la etiqueta que puso la tienda (nada de horarios).
	// Con variantes: "Sin stock" sólo si no queda ninguna, y sin descuento si el precio es "Desde".
	const pastilla = (p: ProductoPublico) =>
		tarjetaDe(p).agotado
			? { s: 'Sin stock', si: I.phClockFill, sc: '#7A5F4C' }
			: p.anterior && !tarjetaDe(p).desde
				? {
						s: `−${Math.round((1 - p.precio / p.anterior) * 100)}%`,
						si: I.phFireFill,
						sc: 'var(--acento, #D9452B)'
					}
				: p.etiqueta
					? { s: p.etiqueta, si: I.phSparkleFill, sc: '#9A2E18' }
					: null;
	// Con variantes de precios distintos, "Desde" (y sin tachado: no se sabe contra cuál).
	const tarjeta = (p: ProductoPublico) => {
		const t = tarjetaDe(p);
		return {
			id: p.id,
			n: p.nombre,
			d: p.descripcion || p.categoria || '',
			price: (t.desde ? 'Desde ' : '') + fmt(t.precio),
			old: p.anterior && !t.desde ? fmt(p.anterior) : null,
			img: img(p.foto, 500, 500),
			tag: pastilla(p),
			href: `${base}/producto/${p.slug ?? p.id}`
		};
	};
	const ver = (s: string) => tienda.secciones.includes(s as TiendaPublica['secciones'][number]);
	const whatsapp = $derived(tienda.whatsapp.replace(/\D/g, ''));
	const wa = (texto: string) =>
		`https://wa.me/${whatsapp.startsWith('54') ? whatsapp : '54' + whatsapp}?text=${encodeURIComponent(texto)}`;

	// Título de la portada: el último tramo va en rojo, como en el diseño ("pan de verdad, / todos los días.").
	const titulo = $derived(partirTitulo(tienda.portada.titulo));
	const inicial = $derived(tienda.nombre.trim().charAt(0).toUpperCase());

	// Las tarjetas grandes de color del diseño (eran combos): lo que la tienda ofrece de verdad.
	const perks = $derived([
		{
			k: envia(tienda.envios) ? 'Envíos' : 'Retiro',
			t:
				tienda.envios.gratisDesde && envia(tienda.envios)
					? `Envío gratis desde ${fmt(tienda.envios.gratisDesde)}`
					: tienda.envios.correo
						? 'Enviamos a todo el país'
						: tienda.envios.moto
							? 'Te lo llevamos en moto'
							: 'Retiralo en el local',
			// Sólo retiro: dónde (sin prometer envíos que la tienda no hace).
			d: !envia(tienda.envios)
				? tienda.direccion
					? `Sin costo, en ${tienda.direccion}.`
					: 'Sin costo. Te avisamos cuando esté listo.'
				: tienda.envios.retiro
					? `O lo retirás en el local${tienda.direccion ? `, ${tienda.direccion}` : ''}.`
					: 'Calculá el costo con tu código postal en cada producto.',
			bg: '#F6D776',
			fg: '#2A1C12',
			bbg: '#2A1C12',
			bfg: '#F4EADB'
		},
		...(tienda.cobros.cuotas > 1
			? [
					{
						k: 'Cuotas',
						t: `${tienda.cobros.cuotas} cuotas sin interés`,
						d: tienda.cobros.mercadoPago
							? 'Con tarjeta, a través de Mercado Pago.'
							: 'En toda la tienda.',
						bg: '#fff',
						fg: '#2A1C12',
						bbg: 'var(--acento, #D9452B)',
						bfg: '#F4EADB'
					}
				]
			: []),
		...(tienda.cobros.transferencia > 0
			? [
					{
						k: 'Transferencia',
						t: `${tienda.cobros.transferencia}% off`,
						d: 'Pagando por transferencia, te mandamos el alias al confirmar el pedido.',
						bg: 'var(--acento, #D9452B)',
						fg: '#F4EADB',
						bbg: '#F4EADB',
						bfg: 'var(--acento, #D9452B)'
					}
				]
			: [])
	]);
	const cats = $derived(tienda.categorias.slice(0, 6));
	const best = $derived(
		tienda.masVendidos.slice(0, 4).flatMap((id) => (porId.get(id) ? [tarjeta(porId.get(id)!)] : []))
	);

	// Catálogo: categoría, orden y de a 24.
	const ORDENES = [
		['vendidos', 'Más pedidos'],
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
	const rail = $derived([
		{
			t: 'Todo',
			c: null,
			img: img(tienda.portada.foto ?? tienda.productos.find((p) => p.foto)?.foto ?? null, 160, 160)
		},
		...tienda.categorias.map((c) => ({ t: c.nombre, c: c.nombre, img: img(c.foto, 160, 160) }))
	]);

	// Ficha.
	let abierto = $state<string>('');
	const acc = $derived(
		producto
			? [
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
	const resenasFicha = $derived(
		producto ? tienda.resenas.filter((r) => r.producto === producto!.nombre) : []
	);
	const promedioFicha = $derived(
		resenasFicha.length
			? Math.round((resenasFicha.reduce((s, r) => s + r.estrellas, 0) / resenasFicha.length) * 10) /
					10
			: 0
	);
	let cant = $state(1);
	$effect(() => {
		void producto?.id;
		cant = 1;
	});
	// Talle, color, presentación…: lo elegido, la variante, su precio, stock y foto (seleccion.svelte.ts).
	const sel = $derived(producto ? new Seleccion(producto) : null);
	// El tachado no va con "Desde": no se sabe contra qué presentación se compara.
	const tachado = $derived(sel && !sel.desde ? sel.anterior : null);
	let cp = $state('');
	let cotizado = $state<OpcionEnvio[] | null>(null);
	async function calcular() {
		if (cpValido(cp)) cotizado = await cotizarYa(cp, (sel?.precio ?? 0) * cant || subtotal, tienda);
	}
	function agregar() {
		if (!sel || tienda.vitrina || !sel.agregar(carrito, cant)) return;
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- ruta de la tienda con su slug
		goto(`${base}/carrito`);
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
			.slice(0, 2)
			.map((id) => porId.get(id)!)
	);
	/** Sumar desde "¿Le sumás…?": con opciones, primero se elige en la ficha. */
	const sumar = (p: ProductoPublico) => {
		if (tienda.vitrina) return;
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- ruta de la tienda con su slug
		if (!carrito.sumarProducto(p)) goto(`${base}/producto/${p.slug ?? p.id}`);
	};

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
	const iconoEnvio = (id: OpcionEnvio['id']) =>
		id === 'retiro' ? I.phStorefrontFill : id === 'moto' ? I.phMopedFill : I.phTruck;
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
						c: '#2F7A43'
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
								? `Dinero en cuenta, débito o crédito en hasta ${tienda.cobros.cuotas} cuotas sin interés`
								: 'Dinero en cuenta, débito o crédito',
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

	// Nosotros: cómo llega el pedido (la tarjeta de "Horarios" del diseño, con los envíos de la tienda).
	const llegadas = $derived([
		...(tienda.envios.correo ? [{ d: 'Correo Argentino', h: 'A sucursal' }] : []),
		...(tienda.envios.moto ? [{ d: 'Envío en moto', h: resumenMoto(tienda.envios) }] : []),
		...(tienda.envios.retiro ? [{ d: 'Retiro en el local', h: 'Sin costo' }] : []),
		...(tienda.envios.gratisDesde
			? [{ d: 'Envío gratis', h: `Desde ${fmt(tienda.envios.gratisDesde)}` }]
			: [])
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
		...(tienda.turnos
			? [
					{
						t: 'Reservar turno',
						href: `${base}/turnos`,
						on: pagina === 'turnos' || pagina === 'turno'
					}
				]
			: []),
		{ t: 'Catálogo', href: `${base}/catalogo`, on: pagina === 'catalogo' && !categoria },
		...tienda.categorias.slice(0, 3).map((c) => ({
			t: c.nombre,
			href: enlaceCat(c.nombre),
			on: pagina === 'catalogo' && categoria === c.nombre
		})),
		{ t: 'Nosotros', href: `${base}/nosotros`, on: pagina === 'nosotros' }
	]);
	// Reseñas y preguntas con las piezas de Miga: Schibsted Grotesk pesada, tarjetas blancas y botones de píldora.
	const tema: TemaOpiniones = {
		texto: '#2A1C12',
		suave: '#7A5F4C',
		campo: '#F4EADB',
		tarjeta: '#fff',
		borde: 'rgba(58,36,20,.14)',
		acento: 'var(--acento, #D9452B)',
		sobreAcento: '#F4EADB',
		estrella: '#D98A2B',
		radio: '12px',
		radioBoton: '999px',
		fuente: "'Instrument Sans', sans-serif",
		titulos: "'Schibsted Grotesk', sans-serif",
		pesoTitulos: 800
	};
	const estrellas = (n: number) =>
		Array.from({ length: 5 }, (_, i) =>
			n >= i + 1 ? I.phStarFill : n >= i + 0.5 ? I.phStarHalfFill : I.phStar
		);
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
</script>

<svelte:window bind:innerWidth={ancho} />

<svelte:head>
	<link rel="preconnect" href="https://fonts.googleapis.com" />
	<link
		href="https://fonts.googleapis.com/css2?family=Schibsted+Grotesk:wght@500;700;800;900&family=Caveat:wght@600;700&display=swap"
		rel="stylesheet"
	/>
</svelte:head>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
<div
	class="miga"
	style="min-height:100vh;font-family:'Instrument Sans',system-ui,sans-serif;font-variant-numeric:normal;color:#2A1C12;background:#F4EADB"
>
	<div style="container-type:inline-size;width:100%;background:#F4EADB;position:relative">
		<!-- barra -->
		<div
			style="display:flex;justify-content:center;align-items:center;gap:12px;flex-wrap:wrap;padding:9px 16px;background:#2A1C12;color:#F4EADB;font-size:13px;text-align:center"
		>
			{#if conteo.actual}<span>{conteo.texto}</span><span
					style="display:flex;gap:4px;font-variant-numeric:tabular-nums"
					>{#each conteo.partes(3) as c (c.u)}<span
							style="padding:2px 7px;border-radius:6px;background:rgba(251,241,230,.14)"
							>{c.v}{c.u}</span
						>{/each}</span
				>{:else}<span style="display:flex;align-items:center;gap:7px"
					>{tienda.envios.gratisDesde
						? `Envío gratis desde ${fmt(tienda.envios.gratisDesde)}`
						: tienda.cobros.cuotas > 1
							? `${tienda.cobros.cuotas} cuotas sin interés`
							: 'Compra protegida'}</span
				>{/if}{#if wide && tienda.cobros.transferencia > 0}<span
					style="display:flex;gap:4px;padding:2px 7px;border-radius:6px;background:rgba(251,241,230,.14)"
					>{tienda.cobros.transferencia}% off por transferencia</span
				>{/if}
		</div>

		<!-- header -->
		<header
			style="position:sticky;top:0;z-index:20;display:flex;align-items:center;gap:18px;padding:14px clamp(14px,3.5cqw,44px);background:rgba(251,241,230,.92);backdrop-filter:blur(10px);border-bottom:1px solid rgba(58,36,20,.1)"
		>
			<a
				href={base || '/'}
				style="display:flex;align-items:center;gap:10px;min-width:0;text-decoration:none;color:#2A1C12"
				>{#if !tienda.logo}<span
						style="width:40px;height:40px;flex:none;border-radius:50%;background:var(--acento, #D9452B);color:#F4EADB;display:grid;place-items:center;font-family:'Schibsted Grotesk',sans-serif;font-weight:800;letter-spacing:-.03em;font-size:20px"
						>{inicial}</span
					>{/if}<span
					style="font-family:'Schibsted Grotesk',sans-serif;font-weight:800;letter-spacing:-.03em;font-size:23px;line-height:1;overflow-wrap:anywhere"
					><LogoTienda {tienda} /></span
				></a
			>
			{#if wide}<nav style="display:flex;gap:4px;margin-left:12px">
					{#each nav as n (n.t)}<a
							href={n.href}
							style="text-decoration:none;font:600 14px 'Instrument Sans';padding:9px 14px;border-radius:999px;background:{n.on
								? '#F7E3A6'
								: 'transparent'};color:#2A1C12;white-space:nowrap">{n.t}</a
						>{/each}
				</nav>{/if}
			<span style="flex:1"></span>
			{#if wide && tienda.direccion}<span
					style="display:flex;align-items:center;gap:7px;font-size:13px;color:#7A5F4C"
					><Icon
						icon={I.phMapPinFill}
						size={15}
						style="color:var(--acento, #D9452B)"
					/>{tienda.direccion}</span
				>{/if}
			<a
				href="{base}/carrito"
				aria-label="Tu canasta"
				style="display:flex;align-items:center;gap:8px;flex:none;text-decoration:none;padding:10px 16px;border-radius:999px;background:#2A1C12;color:#F4EADB;font:700 14px 'Instrument Sans'"
				><Icon icon={I.phBasketFill} size={17} />{carrito.unidades}{#if wide && subtotal}<span
						style="opacity:.7;font-weight:500">· {fmt(subtotal)}</span
					>{/if}</a
			>
			{#if nar}<a
					href="{base}/catalogo"
					aria-label="Catálogo"
					style="display:grid;flex:none;color:#2A1C12"><Icon icon={I.phListBold} size={24} /></a
				>{/if}{#if nar && tienda.turnos}<a
					href="{base}/turnos"
					aria-label="Reservar turno"
					style="display:grid;flex:none;color:#2A1C12"
					><Icon icon={I.phCalendarPlusBold} size={24} /></a
				>{/if}
		</header>

		{#if tienda.vitrina}
			<div
				style="padding:10px 16px;background:#F7E3A6;color:#2A1C12;font-size:13px;text-align:center"
			>
				Es una tienda de ejemplo de la plantilla Miga: podés recorrerla, pero no se compra
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
							style="position:relative;padding:clamp(28px,5cqw,64px) clamp(14px,3.5cqw,44px) 0"
						>
							<div
								style="display:flex;justify-content:space-between;align-items:flex-end;gap:clamp(18px,3cqw,40px);flex-wrap:wrap"
							>
								<h1
									style="margin:0;flex:1 1 520px;min-width:0;font-family:'Schibsted Grotesk',sans-serif;font-weight:900;font-size:clamp(58px,10.4cqw,168px);line-height:.84;letter-spacing:-.065em;overflow-wrap:anywhere"
								>
									{titulo[0]}{#if titulo[1]}<br /><span style="color:var(--acento, #D9452B)"
											>{titulo[1]}</span
										>{/if}
								</h1>
								<div
									style="flex:0 1 340px;min-width:0;display:flex;flex-direction:column;gap:14px;padding-bottom:6px"
								>
									{#if tienda.ciudad && !tienda.portada.lema
											.toLowerCase()
											.includes(tienda.ciudad.toLowerCase())}<span
											style="font-family:'Caveat',cursive;font-size:26px;line-height:1;color:var(--acento, #D9452B);transform:rotate(-3deg);align-self:flex-start"
											>desde {tienda.ciudad} ↓</span
										>{/if}
									{#if tienda.portada.lema}<p
											style="margin:0;font-size:16.5px;line-height:1.5;color:#5E4A3A"
										>
											{tienda.portada.lema}
										</p>{/if}
									<div style="display:flex;gap:8px;flex-wrap:wrap">
										<a
											href="{base}/catalogo"
											style="display:flex;align-items:center;gap:8px;padding:15px 22px;border:0;border-radius:14px;background:#2A1C12;color:#F4EADB;font:700 15px 'Instrument Sans';text-decoration:none"
											><Icon icon={I.phBasketBold} size={17} />Hacer mi pedido</a
										><a
											href="{base}/nosotros"
											style="padding:15px 20px;border:2px solid #2A1C12;border-radius:14px;background:transparent;color:#2A1C12;font:700 15px 'Instrument Sans';text-decoration:none"
											>Conocenos</a
										>
									</div>
									{#if tienda.promedio && tienda.resenas.length}<div
											style="display:flex;align-items:center;gap:8px;font-size:13px;color:#5E4A3A"
										>
											<span style="display:flex;color:var(--acento, #D9452B)"
												>{#each estrellas(tienda.promedio) as e, i (i)}<Icon
														icon={e}
														size={14}
													/>{/each}</span
											><b style="color:#2A1C12">{tienda.promedio.toLocaleString('es-AR')}</b> ·
											{tienda.resenas.length}
											{tienda.resenas.length === 1 ? 'reseña' : 'reseñas'}
										</div>{/if}
								</div>
							</div>
							{#if ver('cats') && cats.length}
								<!-- "La vitrina de hoy" del diseño, con las categorías de la tienda -->
								<div
									style="margin-top:clamp(26px,4cqw,46px);border-radius:26px;background:#2A1C12;color:#F4EADB;padding:clamp(14px,2cqw,22px);position:relative"
								>
									<div
										style="display:flex;justify-content:space-between;align-items:center;gap:12px;flex-wrap:wrap;padding:2px 4px 14px"
									>
										<span
											style="font-family:'Schibsted Grotesk',sans-serif;font-weight:800;font-size:clamp(20px,2.2cqw,26px);letter-spacing:-.03em"
											>La vitrina</span
										><a
											href="{base}/catalogo"
											style="display:flex;align-items:center;gap:8px;padding:7px 12px;border-radius:999px;background:rgba(246,215,118,.14);font-size:13px;font-weight:600;color:#F4EADB;text-decoration:none"
											><span
												style="width:8px;height:8px;border-radius:50%;background:#F6D776;box-shadow:0 0 0 4px rgba(246,215,118,.25)"
											></span>Ver todo</a
										>
									</div>
									<div
										style="display:grid;grid-auto-flow:column;grid-auto-columns:minmax(150px,1fr);gap:10px;overflow-x:auto;scrollbar-width:none"
									>
										{#each cats as c (c.nombre)}<a
												href={enlaceCat(c.nombre)}
												style="display:block;border-radius:18px;padding:8px;background:rgba(244,234,219,.06);color:#F4EADB;text-decoration:none"
												><div
													style="position:relative;aspect-ratio:4/3;border-radius:12px;overflow:hidden;background:#4A3626"
												>
													{#if c.foto}<img
															src={img(c.foto, 400, 300)}
															alt=""
															style="width:100%;height:100%;object-fit:cover"
														/>{/if}
												</div>
												<div
													style="display:flex;justify-content:space-between;align-items:baseline;flex-wrap:wrap;gap:6px;padding:10px 4px 4px"
												>
													<b
														style="min-width:0;font-family:'Schibsted Grotesk',sans-serif;font-weight:900;font-size:24px;letter-spacing:-.04em"
														>{c.nombre}</b
													><span
														style="font-size:11px;font-weight:700;padding:3px 8px;border-radius:999px;background:rgba(244,234,219,.1);color:#D8C8B6;white-space:nowrap"
														>{c.cantidad}</span
													>
												</div></a
											>{/each}
									</div>
								</div>
							{/if}
						</section>
					{/if}
				{:else if seccion === 'prods'}
					{#if ver('prods') && best.length}
						<section style="padding:clamp(48px,7cqw,96px) clamp(14px,3.5cqw,44px) 0">
							<div
								style="display:flex;justify-content:space-between;align-items:end;gap:16px;flex-wrap:wrap"
							>
								<h2
									style="margin:0;font-family:'Schibsted Grotesk',sans-serif;font-weight:800;font-size:clamp(36px,4.8cqw,62px);line-height:1;letter-spacing:-.05em"
								>
									Lo más pedido
								</h2>
								<a
									href="{base}/catalogo"
									style="text-decoration:none;padding:11px 18px;border-radius:999px;background:#F7E3A6;color:#2A1C12;font:700 14px 'Instrument Sans'"
									>Ver todo</a
								>
							</div>
							<div
								style="margin-top:24px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(46%,230px),1fr));gap:clamp(10px,1.6cqw,18px)"
							>
								{#each best as p (p.id)}
									{@render tarjetaProducto(p, 19)}
								{/each}
							</div>
						</section>
					{/if}

					<section
						style="padding:clamp(48px,7cqw,96px) clamp(14px,3.5cqw,44px) 0;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,300px),1fr));gap:12px"
					>
						{#each perks as c (c.k)}<div
								style="position:relative;border-radius:28px;overflow:hidden;background:{c.bg};color:{c.fg};padding:22px;display:flex;flex-direction:column;gap:12px;min-height:300px"
							>
								<div
									style="font-size:12px;font-weight:700;letter-spacing:.12em;text-transform:uppercase;opacity:.75"
								>
									{c.k}
								</div>
								<div
									style="font-family:'Schibsted Grotesk',sans-serif;font-weight:800;letter-spacing:-.03em;font-size:clamp(28px,3cqw,36px);line-height:1.05"
								>
									{c.t}
								</div>
								<div style="font-size:14px;line-height:1.5;opacity:.85">{c.d}</div>
								<span style="flex:1"></span>
								<div style="display:flex;justify-content:flex-end;align-items:center;gap:10px">
									<a
										href="{base}/catalogo"
										style="padding:11px 16px;border:0;border-radius:999px;background:{c.bbg};color:{c.bfg};font:700 13.5px 'Instrument Sans';text-decoration:none"
										>Hacer mi pedido</a
									>
								</div>
							</div>{/each}
					</section>
				{:else if seccion === 'revs'}
					{#if ver('revs') && tienda.resenas.length}
						<section style="padding:clamp(48px,7cqw,96px) clamp(14px,3.5cqw,44px) 0">
							<h2
								style="margin:0;font-family:'Schibsted Grotesk',sans-serif;font-weight:800;font-size:clamp(36px,4.8cqw,62px);line-height:1;letter-spacing:-.05em"
							>
								Lo que dicen
							</h2>
							<div
								style="margin-top:24px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,290px),1fr));gap:12px"
							>
								{#each tienda.resenas.slice(0, 3) as r, i (i)}
									{@const pr = tienda.productos.find((p) => p.nombre === r.producto)}
									<div
										style="padding:12px;border-radius:18px;background:#fff;box-shadow:0 1px 2px rgba(58,36,20,.06)"
									>
										{#if pr?.foto}<div
												style="aspect-ratio:4/3;border-radius:18px;overflow:hidden;background:#F1DFC9"
											>
												<img
													src={img(pr.foto, 600, 450)}
													alt=""
													style="width:100%;height:100%;object-fit:cover"
												/>
											</div>{/if}
										<div style="padding:12px 6px 6px">
											<div style="display:flex;color:#D98A2B">
												{#each estrellas(r.estrellas) as e, j (j)}<Icon icon={e} size={14} />{/each}
											</div>
											<p style="margin:8px 0 0;font-size:15px;line-height:1.45">“{r.texto}”</p>
											<div style="margin-top:8px;font-size:12.5px;color:#7A5F4C">
												{r.autor}{#if r.producto}
													· {r.conCompra ? 'compró' : 'sobre'} {r.producto}{/if}
											</div>
											{#if r.conCompra}<span
													style="margin-top:8px;display:inline-flex;align-items:center;gap:5px;padding:4px 10px;border-radius:999px;background:#E6F0E2;font-size:12px;font-weight:700;color:#2F7A43"
													><Icon icon={I.phSealCheckFill} size={13} />Compra verificada</span
												>{/if}
											{#if r.respuesta}<div
													style="margin-top:10px;padding:10px 12px;border-radius:14px;background:#F4EADB;font-size:13.5px;line-height:1.5;color:#6A4F3C"
												>
													<b style="color:#2A1C12">Respuesta de {tienda.nombre}</b><br
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
						<section style="padding:clamp(48px,7cqw,96px) clamp(14px,3.5cqw,44px)">
							<h2
								style="margin:0;font-family:'Schibsted Grotesk',sans-serif;font-weight:800;font-size:clamp(36px,4.8cqw,62px);line-height:1"
							>
								Así sale
							</h2>
							<div
								style="margin-top:20px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(30%,160px),1fr));gap:8px"
							>
								{#each tienda.fotos as g (g)}<div
										style="aspect-ratio:1;border-radius:20px;overflow:hidden;background:#F1DFC9"
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

		<!-- ============ CATÁLOGO ============ -->
		{#if pagina === 'catalogo' && !tienda.productos.length}
			<section style="padding:clamp(40px,7cqw,96px) clamp(16px,4cqw,48px)">
				<VacioTienda {tienda} {tema} motivo="sin-productos" />
			</section>
		{:else if pagina === 'catalogo'}
			<section style="padding:24px clamp(14px,3.5cqw,44px) clamp(48px,7cqw,90px)">
				<div
					style="display:flex;justify-content:space-between;align-items:end;gap:16px;flex-wrap:wrap"
				>
					<h1
						style="margin:0;min-width:0;font-family:'Schibsted Grotesk',sans-serif;font-weight:800;font-size:clamp(42px,6cqw,80px);line-height:.98;letter-spacing:-.05em;overflow-wrap:anywhere"
					>
						{categoria ?? 'Todo'}
					</h1>
					<span style="font-size:13.5px;color:#7A5F4C"
						>{listado.length} {listado.length === 1 ? 'producto' : 'productos'}</span
					>
				</div>
				<div
					style="margin-top:20px;display:flex;gap:10px;overflow-x:auto;scrollbar-width:none;padding:4px 0"
				>
					{#each rail as c (c.c ?? '')}{@const on = c.c === categoria}<a
							href={enlaceCat(c.c)}
							style="flex:none;display:flex;flex-direction:column;align-items:center;gap:7px;width:96px;text-decoration:none;color:#2A1C12"
							><span
								style="width:76px;height:76px;border-radius:50%;overflow:hidden;background:#F1DFC9;box-shadow:0 0 0 {on
									? '3px'
									: '0px'} var(--acento, #D9452B)"
								>{#if c.img}<img
										src={c.img}
										alt=""
										style="width:100%;height:100%;object-fit:cover"
									/>{/if}</span
							><span
								style="font-size:13px;font-weight:{on
									? 700
									: 500};text-align:center;max-width:96px;overflow-wrap:anywhere">{c.t}</span
							></a
						>{/each}
				</div>
				<div
					style="margin-top:16px;display:flex;justify-content:flex-end;align-items:center;gap:10px;flex-wrap:wrap;font-size:13.5px"
				>
					<label
						style="position:relative;display:flex;align-items:center;gap:6px;font-weight:700;cursor:pointer"
						><span class="oculto">Ordenar</span><select
							value={orden}
							onchange={(e) => goto(enlaceCat(categoria, e.currentTarget.value), { replace: true })}
							style="appearance:none;-webkit-appearance:none;border:0;background:transparent;padding:0 18px 0 0;font:700 13.5px 'Instrument Sans';color:#2A1C12;cursor:pointer"
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
						style="margin-top:20px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(46%,210px),1fr));gap:12px"
					>
						{#each catalog as p (p.id)}
							{@render tarjetaProducto(p, 18)}
						{/each}
					</div>
					{#if listado.length > mostrar}
						<div style="margin-top:32px;display:flex;justify-content:center">
							<button
								type="button"
								onclick={() => (mostrar += 24)}
								style="padding:15px 22px;border:2px solid #2A1C12;border-radius:14px;background:transparent;color:#2A1C12;font:700 15px 'Instrument Sans';cursor:pointer"
								>Ver más</button
							>
						</div>
					{/if}
				{:else}
					<p style="margin:40px 0;font-size:15.5px;color:#6A4F3C">
						Todavía no hay productos en esta categoría.
					</p>
				{/if}
			</section>
		{/if}

		<!-- ============ PRODUCTO ============ -->
		{#if pagina === 'producto' && producto && sel}
			<section
				style="padding:24px clamp(14px,3.5cqw,44px) 0;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,400px),1fr));gap:clamp(20px,4cqw,56px);align-items:start"
			>
				<div style="display:flex;flex-direction:column;gap:8px">
					<div
						style="position:relative;aspect-ratio:1;border-radius:32px;overflow:hidden;background:#F1DFC9"
					>
						{#if sel.foto}<img
								src={img(sel.foto, 1100)}
								alt={producto.nombre}
								style="width:100%;height:100%;object-fit:cover"
							/>{/if}
						{#if pastilla(producto)}{@const t = pastilla(producto)!}<span
								style="position:absolute;top:14px;left:14px;display:flex;align-items:center;gap:6px;padding:7px 12px;border-radius:999px;background:#F4EADB;font-size:12.5px;font-weight:700;color:{t.sc}"
								><Icon icon={t.si} size={14} />{t.s}</span
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
									style="aspect-ratio:1;padding:0;border:0;border-radius:18px;overflow:hidden;background:#F1DFC9;cursor:pointer;box-shadow:0 0 0 {i ===
									sel.fotoIdx
										? '2.5px'
										: '0px'} #2A1C12"
									><img
										src={img(f, 200, 200)}
										alt=""
										style="display:block;width:100%;height:100%;object-fit:cover"
									/></button
								>{/each}
						</div>{/if}
				</div>
				<div style="position:sticky;top:90px;min-width:0">
					<div style="font-size:13px;font-weight:700;color:var(--acento, #D9452B)">
						<a href={base || '/'} style="color:inherit;text-decoration:none">Inicio</a
						>{#if producto.categoria}&nbsp;·&nbsp;<a
								href={enlaceCat(producto.categoria)}
								style="color:inherit;text-decoration:none">{producto.categoria}</a
							>{/if}
					</div>
					<h1
						style="margin:6px 0 0;font-family:'Schibsted Grotesk',sans-serif;font-weight:800;font-size:clamp(38px,4.6cqw,58px);line-height:1;letter-spacing:-.05em;overflow-wrap:anywhere"
					>
						{producto.nombre}
					</h1>
					{#if resenasFicha.length}<div
							style="margin-top:10px;display:flex;align-items:center;gap:8px;font-size:13.5px"
						>
							<span style="display:flex;color:#D98A2B"
								>{#each estrellas(promedioFicha) as e, i (i)}<Icon
										icon={e}
										size={14}
									/>{/each}</span
							><b>{promedioFicha.toLocaleString('es-AR')}</b><span style="color:#7A5F4C"
								>{resenasFicha.length} {resenasFicha.length === 1 ? 'reseña' : 'reseñas'}</span
							>
						</div>{/if}
					<div style="margin-top:18px;display:flex;align-items:baseline;gap:10px">
						<b style="font-size:30px">{sel.desde ? 'Desde ' : ''}{fmt(sel.precio)}</b
						>{#if tachado}<s style="color:#7A5F4C">{fmt(tachado)}</s>{/if}
					</div>
					{#if tienda.cobros.transferencia > 0}
						<div style="margin-top:4px;font-size:14px;color:#2F7A43;font-weight:700">
							{fmt(sel.precio * (1 - tienda.cobros.transferencia / 100))} por transferencia
						</div>
					{/if}
					{#if tienda.cobros.cuotas > 1}<div style="margin-top:2px;font-size:13.5px;color:#7A5F4C">
							{tienda.cobros.cuotas} cuotas sin interés de {fmt(sel.precio / tienda.cobros.cuotas)}
						</div>{/if}
					{#each sel.opciones as o (o.nombre)}{@const swatch = o.valores.every(
							(x) => x.color || x.foto
						)}
						<div style="margin-top:20px">
							<div
								style="margin-bottom:8px;font-size:13.5px;font-weight:700;color:{sel.aviso &&
								!o.elegido
									? 'var(--acento, #D9452B)'
									: '#2A1C12'}"
							>
								{o.nombre}{#if o.elegido}: <span style="font-weight:400">{o.elegido}</span>{/if}
							</div>
							{#if swatch}<div style="display:flex;gap:10px;flex-wrap:wrap">
									{#each o.valores as x (x.valor)}<button
											type="button"
											aria-label="{o.nombre} {x.valor}"
											aria-pressed={x.elegido}
											disabled={!x.disponible && !x.elegido}
											onclick={() => sel.elegir(o.nombre, x.valor)}
											style="width:40px;height:40px;padding:0;border:0;border-radius:50%;overflow:hidden;background:#F1DFC9;cursor:pointer;opacity:{x.disponible
												? 1
												: 0.4};box-shadow:inset 0 0 0 1px rgba(58,36,20,.14),0 0 0 {x.elegido
												? '2.5px'
												: '0px'} #F4EADB,0 0 0 {x.elegido ? '5px' : '0px'} #2A1C12"
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
											style="min-height:48px;padding:10px 16px;border-radius:14px;border:2px solid {x.elegido
												? '#2A1C12'
												: 'rgba(58,36,20,.2)'};background:{x.elegido
												? '#2A1C12'
												: 'transparent'};font:700 14px 'Instrument Sans';color:{x.elegido
												? '#F4EADB'
												: x.disponible
													? '#2A1C12'
													: '#7A5F4C'};text-align:left;text-decoration:{x.disponible
												? 'none'
												: 'line-through'};cursor:{x.disponible ? 'pointer' : 'default'}"
											>{x.valor}{#if pv}<span
													style="display:block;font-size:12px;font-weight:400;opacity:.75"
													>{fmt(pv.precio ?? producto.precio)}</span
												>{/if}</button
										>{/each}
								</div>{/if}
						</div>{/each}
					{#if sel.stock > 0 && sel.stock <= 3 && (!sel.conOpciones || sel.variante)}<div
							style="margin-top:12px;font-size:13px;font-weight:700;color:#9A2E18"
						>
							{sel.stock === 1 ? 'Queda 1' : `Quedan ${sel.stock}`}{#if sel.variante}
								en {sel.texto.toLowerCase()}{/if}
						</div>{/if}
					<div style="margin-top:20px;display:flex;gap:8px;flex-wrap:wrap">
						{#if sel.stock > 0}<span
								style="display:flex;align-items:center;border-radius:999px;background:#fff"
								><button
									type="button"
									aria-label="Uno menos"
									disabled={cant <= 1}
									onclick={() => (cant = Math.max(1, cant - 1))}
									style="width:44px;height:50px;display:grid;place-items:center;border:0;background:transparent;cursor:pointer;color:#2A1C12"
									><Icon icon={I.phMinusBold} size={14} /></button
								><b style="width:22px;text-align:center">{cant}</b><button
									type="button"
									aria-label="Uno más"
									disabled={cant >= sel.stock}
									onclick={() => (cant = Math.min(sel.stock, cant + 1))}
									style="width:44px;height:50px;display:grid;place-items:center;border:0;background:transparent;cursor:pointer;color:#2A1C12"
									><Icon icon={I.phPlusBold} size={14} /></button
								></span
							>{/if}<button
							type="button"
							onclick={agregar}
							disabled={sel.stock === 0 || tienda.vitrina}
							style="flex:1;min-width:180px;padding:16px;border:0;border-radius:999px;background:var(--acento, #D9452B);color:#F4EADB;font:700 15px 'Instrument Sans';cursor:pointer"
							>{sel.boton('Sumar al pedido')}</button
						>
					</div>
					{#if whatsapp}<a
							href={wa(
								`¡Hola! Quiero pedir ${producto.nombre}${sel.variante ? ` (${sel.texto})` : ''} (${fmt(sel.precio)}).`
							)}
							target="_blank"
							rel="noopener"
							style="margin-top:8px;display:flex;justify-content:center;align-items:center;gap:8px;padding:13px;border-radius:999px;border:1.5px solid rgba(58,36,20,.2);font-weight:700;color:#2A1C12;text-decoration:none"
							><Icon icon={I.logosWhatsappIcon} size={16} />Pedirlo por WhatsApp</a
						>{/if}
					<LanzamientoFicha {tienda} {producto} {sel} {tema} />
					<CompartirProducto {tienda} {producto} {tema} />
					{#if producto.descripcion}<p
							style="margin:14px 0 0;font-size:14.5px;line-height:1.6;color:#6A4F3C"
						>
							{producto.descripcion}
						</p>{/if}
					{#if envia(tienda.envios)}<div
							style="margin-top:14px;padding:12px;border-radius:18px;background:#fff"
						>
							<label for="cp-ficha" style="font-size:13.5px;font-weight:700">Calculá el envío</label
							>
							<div style="margin-top:8px;display:flex;gap:6px">
								<input
									id="cp-ficha"
									bind:value={cp}
									inputmode="numeric"
									maxlength="4"
									placeholder="Código postal"
									onkeydown={(e) => e.key === 'Enter' && calcular()}
									style="flex:1;min-width:0;padding:11px 12px;border:0;border-radius:12px;background:#F4EADB;font:400 14px 'Instrument Sans';color:#2A1C12"
								/><button
									type="button"
									onclick={calcular}
									style="padding:11px 14px;border-radius:12px;border:1.5px solid rgba(58,36,20,.2);background:transparent;font:700 14px 'Instrument Sans';color:#2A1C12;cursor:pointer"
									>Calcular</button
								>
							</div>
							{#if cotizado}<div
									style="margin-top:12px;display:flex;flex-direction:column;gap:8px;font-size:13.5px"
								>
									{#each cotizado as o (o.id)}<span
											style="display:flex;justify-content:space-between;gap:10px"
											><span>{o.nombre} · <span style="color:#7A5F4C">{o.detalle}</span></span><b
												style="white-space:nowrap;color:{o.precio ? '#2A1C12' : '#2F7A43'}"
												>{o.precio ? fmt(o.precio) : 'Gratis'}</b
											></span
										>{/each}
								</div>{/if}
						</div>{/if}
					<div style="margin-top:8px;padding:6px 16px;border-radius:18px;background:#fff">
						{#each acc as a, i (a.t)}<div
								style="padding:12px 0;border-bottom:{i < acc.length - 1
									? '1px solid rgba(58,36,20,.08)'
									: '0'}"
							>
								<button
									type="button"
									onclick={() => (abierto = a.open ? '' : a.t)}
									aria-expanded={a.open}
									style="width:100%;display:flex;justify-content:space-between;align-items:center;padding:0;border:0;background:transparent;font:700 14.5px 'Instrument Sans';color:#2A1C12;cursor:pointer;text-align:left"
									>{a.t}<Icon icon={a.i} size={14} /></button
								>{#if a.open}<p style="margin:8px 0 0;font-size:14px;line-height:1.6;color:#6A4F3C">
										{a.d}
									</p>{/if}
							</div>{/each}
					</div>
				</div>
			</section>
			<section style="padding:clamp(40px,6cqw,72px) clamp(14px,3.5cqw,44px) 0">
				<div
					style="max-width:820px;padding:clamp(18px,3cqw,28px);border-radius:28px;background:#fff"
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
				<section style="padding:clamp(52px,7cqw,96px) clamp(14px,3.5cqw,44px)">
					<h2
						style="margin:0;font-family:'Schibsted Grotesk',sans-serif;font-weight:800;font-size:clamp(32px,4.2cqw,52px);line-height:1"
					>
						Para acompañar
					</h2>
					<div
						style="margin-top:22px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(46%,200px),1fr));gap:12px"
					>
						{#each related as p (p.id)}<a
								href={p.href}
								style="display:block;padding:8px;border-radius:22px;background:#fff;text-decoration:none;color:#2A1C12"
								><span
									style="display:block;aspect-ratio:1;border-radius:16px;overflow:hidden;background:#F1DFC9"
									>{#if p.img}<img
											src={p.img}
											alt=""
											style="width:100%;height:100%;object-fit:cover"
										/>{/if}</span
								><span
									style="display:block;padding:10px 4px 2px;font-family:'Schibsted Grotesk',sans-serif;font-weight:800;letter-spacing:-.03em;font-size:17px"
									>{p.n}</span
								><b style="display:block;padding:0 4px 4px;font-size:14.5px">{p.price}</b></a
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
					style="margin:0;font-family:'Schibsted Grotesk',sans-serif;font-weight:800;font-size:clamp(42px,6cqw,76px);line-height:1;letter-spacing:-.05em"
				>
					Tu canasta
				</h1>
				{#if !lineas.length}
					<p style="margin:14px 0 0;font-size:15.5px;color:#6A4F3C">Todavía no sumaste nada.</p>
					<a
						href="{base}/catalogo"
						style="display:inline-flex;align-items:center;gap:8px;margin-top:18px;padding:15px 22px;border-radius:14px;background:#2A1C12;color:#F4EADB;font:700 15px 'Instrument Sans';text-decoration:none"
						><Icon icon={I.phBasketBold} size={17} />Hacer mi pedido</a
					>
				{:else}
					<div
						style="margin-top:22px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,340px),1fr));gap:clamp(16px,3cqw,36px);align-items:start"
					>
						<div style="display:flex;flex-direction:column;gap:10px;min-width:0">
							{#if tienda.envios.gratisDesde}
								<div
									style="padding:16px;border-radius:24px;background:#F7E3A6;display:flex;flex-direction:column;gap:10px"
								>
									<div style="display:flex;gap:12px;align-items:center">
										<Icon icon={I.phSparkleFill} size={22} style="color:#9A2E18;flex:none" />
										<span style="font-size:14px"
											>{#if falta}Te faltan <b>{fmt(falta)}</b> para el envío gratis{:else}<b
													>Tenés envío gratis</b
												>{/if}</span
										>
									</div>
									<div
										style="position:relative;height:4px;border-radius:2px;background:rgba(58,36,20,.14)"
									>
										<div
											style="position:absolute;left:0;top:0;bottom:0;width:{Math.min(
												100,
												(subtotal / tienda.envios.gratisDesde) * 100
											)}%;border-radius:2px;background:var(--acento, #D9452B)"
										></div>
									</div>
								</div>
							{/if}
							{#each lineas as c (c.clave)}<div
									style="display:flex;gap:14px;padding:12px;border-radius:18px;background:#fff"
								>
									{#if c.producto.foto}<img
											src={img(c.producto.foto, 200, 200)}
											alt=""
											style="width:84px;height:84px;border-radius:18px;object-fit:cover;flex:none;background:#F1DFC9"
										/>{/if}
									<div style="flex:1;min-width:0;display:flex;flex-direction:column">
										<div style="display:flex;justify-content:space-between;gap:10px">
											<a
												href="{base}/producto/{c.producto.slug ?? c.id}"
												style="min-width:0;font-family:'Schibsted Grotesk',sans-serif;font-weight:800;letter-spacing:-.03em;font-size:18px;color:#2A1C12;text-decoration:none;overflow-wrap:anywhere"
												>{c.producto.nombre}</a
											><b style="white-space:nowrap">{fmt(c.subtotal)}</b>
										</div>
										<span style="font-size:13px;color:#7A5F4C"
											>{c.variante ? `${c.variante} · ` : ''}{fmt(c.producto.precio)} c/u</span
										>
										<span style="flex:1;min-height:8px"></span>
										<div style="display:flex;justify-content:space-between;align-items:center">
											<span
												style="display:flex;align-items:center;gap:2px;padding:0 2px;border-radius:999px;background:#F4EADB"
												><button
													type="button"
													aria-label="Uno menos"
													onclick={() => carrito.fijar(c.clave, c.cantidad - 1)}
													style="width:30px;height:30px;display:grid;place-items:center;border:0;background:transparent;cursor:pointer;color:#2A1C12"
													><Icon icon={I.phMinusBold} size={12} /></button
												><b style="font-size:14px;min-width:14px;text-align:center">{c.cantidad}</b
												><button
													type="button"
													aria-label="Uno más"
													disabled={c.cantidad >= c.producto.stock}
													onclick={() => carrito.fijar(c.clave, c.cantidad + 1)}
													style="width:30px;height:30px;display:grid;place-items:center;border:0;background:transparent;cursor:pointer;color:#2A1C12"
													><Icon icon={I.phPlusBold} size={12} /></button
												></span
											><button
												type="button"
												aria-label="Quitar {c.producto.nombre}"
												onclick={() => carrito.fijar(c.clave, 0)}
												style="display:grid;padding:4px;border:0;background:transparent;cursor:pointer;color:#9A8270"
												><Icon icon={I.phTrash} size={18} /></button
											>
										</div>
									</div>
								</div>{/each}
							{#each upsell as p (p.id)}<div
									style="padding:16px;border-radius:24px;background:#F7E3A6;display:flex;gap:12px;align-items:center;flex-wrap:wrap"
								>
									{#if p.foto}<img
											src={img(p.foto, 120, 120)}
											alt=""
											style="width:44px;height:44px;border-radius:14px;object-fit:cover;flex:none;background:#F1DFC9"
										/>{:else}<Icon icon={I.phBasketFill} size={28} style="color:#9A2E18" />{/if}
									<div style="flex:1;min-width:200px">
										<div style="font-weight:700">¿Le sumás {p.nombre}?</div>
										{#if p.categoria}<div style="font-size:13px;color:#6A4F3C">
												{p.categoria}
											</div>{/if}
									</div>
									<button
										type="button"
										onclick={() => sumar(p)}
										style="padding:10px 14px;border:0;border-radius:999px;background:#2A1C12;color:#F4EADB;font:700 13.5px 'Instrument Sans';cursor:pointer"
										>+ {fmt(p.precio)}</button
									>
								</div>{/each}
						</div>
						<div
							style="position:sticky;top:90px;padding:22px;border-radius:28px;background:#fff;box-shadow:0 20px 40px -24px rgba(58,36,20,.3);display:flex;flex-direction:column;gap:11px;font-size:14px"
						>
							<div
								style="font-family:'Schibsted Grotesk',sans-serif;font-weight:800;letter-spacing:-.03em;font-size:26px"
							>
								Resumen
							</div>
							<span style="display:flex;justify-content:space-between"
								><span>Subtotal</span><span>{fmt(subtotal)}</span></span
							>
							<span style="display:flex;justify-content:space-between;gap:10px;color:#7A5F4C"
								><span>Envío</span><span style="text-align:right"
									>Se calcula en el próximo paso</span
								></span
							>
							<div style="height:1px;background:rgba(58,36,20,.1)"></div>
							<span style="display:flex;justify-content:space-between;align-items:baseline"
								><b>Total</b><b
									style="font-family:'Schibsted Grotesk',sans-serif;font-weight:800;letter-spacing:-.03em;font-size:28px"
									>{fmt(subtotal)}</b
								></span
							>
							{#if tienda.cobros.transferencia > 0}
								<span style="color:#2F7A43;font-weight:700"
									>{fmt(subtotal * (1 - tienda.cobros.transferencia / 100))} pagando por transferencia</span
								>
							{/if}
							{#if tienda.cobros.cuotas > 1}<span style="font-size:13px;color:#7A5F4C"
									>{tienda.cobros.cuotas} cuotas sin interés de {fmt(
										subtotal / tienda.cobros.cuotas
									)}</span
								>{/if}
							<a
								href="{base}/checkout"
								style="margin-top:4px;padding:16px;border-radius:999px;background:var(--acento, #D9452B);color:#F4EADB;font:700 15px 'Instrument Sans';text-align:center;text-decoration:none"
								>Elegir envío y pagar</a
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
					<p style="margin:0;font-size:15.5px;color:#6A4F3C">Tu canasta está vacía.</p>
					<a href="{base}/catalogo" style="color:#2A1C12;font-weight:700">Hacer mi pedido</a>
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
							style="margin:0;font-family:'Schibsted Grotesk',sans-serif;font-weight:800;font-size:clamp(38px,5cqw,62px);line-height:1;letter-spacing:-.05em"
						>
							Casi listo
						</h1>
						<div>
							<div style="font-size:15px;font-weight:700;margin-bottom:10px">Tus datos</div>
							<div
								style="display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));gap:8px"
							>
								{#each [['email', 'Email', 'email', 'email'], ['nombre', 'Nombre y apellido', 'text', 'name'], ['telefono', 'WhatsApp', 'tel', 'tel'], ['dni', 'DNI', 'text', 'off'], ...camposFactura(tienda)] as [n, l, tipo, ac] (n)}
									<label
										style="display:block;padding:11px 14px;border-radius:16px;background:#fff;box-shadow:inset 0 0 0 1.5px rgba(58,36,20,.1)"
										><span style="display:block;font-size:11.5px;color:#9A8270">{l}</span><input
											name={n}
											type={tipo}
											autocomplete={ac as 'email'}
											required={!OPCIONALES.has(n)}
											maxlength="120"
											style="width:100%;margin-top:2px;padding:0;border:0;background:transparent;font:400 14.5px 'Instrument Sans';color:#2A1C12;outline:none"
										/></label
									>
								{/each}
							</div>
							<!-- Sin bloque de newsletter en el diseño de esta plantilla: la casilla del checkout (§50). -->
							<label
								style="margin-top:10px;display:flex;align-items:center;gap:10px;min-height:44px;font:400 14.5px 'Instrument Sans';color:#2A1C12;cursor:pointer"
								><input
									type="checkbox"
									name="novedades"
									style="flex:none;width:18px;height:18px;margin:0;accent-color:var(--acento, #D9452B)"
								/>Quiero recibir novedades</label
							>
						</div>
						<div>
							<div style="font-size:15px;font-weight:700;margin-bottom:10px">Entrega</div>
							<div style="display:grid;grid-template-columns:minmax(0,140px) minmax(0,1fr);gap:8px">
								<label
									style="display:block;padding:11px 14px;border-radius:16px;background:#fff;box-shadow:inset 0 0 0 1.5px rgba(58,36,20,.1)"
									><span style="display:block;font-size:11.5px;color:#9A8270">Código postal</span
									><input
										name="cp"
										bind:value={cp}
										inputmode="numeric"
										maxlength="4"
										style="width:100%;margin-top:2px;padding:0;border:0;background:transparent;font:400 14.5px 'Instrument Sans';color:#2A1C12;outline:none"
									/></label
								>
								<label
									style="display:block;padding:11px 14px;border-radius:16px;background:#fff;box-shadow:inset 0 0 0 1.5px rgba(58,36,20,.1)"
									><span style="display:block;font-size:11.5px;color:#9A8270"
										>Dirección (calle, número, piso)</span
									><input
										name="direccion"
										maxlength="160"
										autocomplete="street-address"
										required={envio !== 'retiro'}
										style="width:100%;margin-top:2px;padding:0;border:0;background:transparent;font:400 14.5px 'Instrument Sans';color:#2A1C12;outline:none"
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
											? 'var(--acento, #D9452B)'
											: 'rgba(58,36,20,.1)'};background:{on
											? '#FFF6E0'
											: '#fff'};text-align:left;cursor:pointer;font-family:inherit;color:#2A1C12"
										><span
											style="width:20px;height:20px;border-radius:50%;border:2px solid {on
												? 'var(--acento, #D9452B)'
												: 'rgba(58,36,20,.25)'};display:grid;place-items:center;flex:none"
											>{#if on}<span
													style="width:10px;height:10px;border-radius:50%;background:var(--acento, #D9452B)"
												></span>{/if}</span
										><Icon
											icon={iconoEnvio(s.id)}
											size={22}
											style="color:var(--acento, #D9452B);flex:none"
										/><span style="flex:1;min-width:0"
											><span style="display:block;font-size:14.5px;font-weight:700">{s.nombre}</span
											><span style="display:block;font-size:12.5px;color:#7A5F4C">{s.detalle}</span
											></span
										><b style="font-size:14px;color:{s.precio ? '#2A1C12' : '#2F7A43'}"
											>{s.precio ? fmt(s.precio) : 'Gratis'}</b
										></button
									>{:else}<span style="font-size:13.5px;color:#7A5F4C"
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
											? 'var(--acento, #D9452B)'
											: 'rgba(58,36,20,.1)'};background:{on
											? '#FFF6E0'
											: '#fff'};text-align:left;cursor:pointer;font-family:inherit;color:#2A1C12"
										><span
											style="width:20px;height:20px;border-radius:50%;border:2px solid {on
												? 'var(--acento, #D9452B)'
												: 'rgba(58,36,20,.25)'};display:grid;place-items:center;flex:none"
											>{#if on}<span
													style="width:10px;height:10px;border-radius:50%;background:var(--acento, #D9452B)"
												></span>{/if}</span
										><Icon icon={s.i} size={22} style="color:{s.c};flex:none" /><span
											style="flex:1;min-width:0"
											><span style="display:block;font-size:14.5px;font-weight:600">{s.t}</span
											><span style="display:block;font-size:12.5px;color:#7A5F4C">{s.d}</span></span
										></button
									>{/each}
							</div>
						</div>
						{#if error}<div
								role="alert"
								style="padding:12px 14px;border-radius:16px;background:rgba(217,69,43,.12);color:#9A2E18;font-size:14px"
							>
								{error}
							</div>{/if}
						<button
							disabled={!envio || !pagos.length || enviando || tienda.vitrina}
							style="padding:17px;border:0;border-radius:999px;background:var(--acento, #D9452B);color:#F4EADB;font:700 15.5px 'Instrument Sans';cursor:pointer;display:flex;justify-content:center;align-items:center;gap:8px"
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
						style="position:sticky;top:90px;padding:22px;border-radius:28px;background:#2A1C12;color:#F4EADB"
					>
						<div
							style="font-family:'Schibsted Grotesk',sans-serif;font-weight:800;letter-spacing:-.03em;font-size:26px"
						>
							Tu pedido
						</div>
						<div style="margin-top:14px;display:flex;flex-direction:column;gap:12px">
							{#each lineas as c (c.clave)}<div style="display:flex;gap:12px;align-items:center">
									<span style="position:relative;flex:none"
										>{#if c.producto.foto}<img
												src={img(c.producto.foto, 120, 120)}
												alt=""
												style="display:block;width:50px;height:50px;border-radius:14px;object-fit:cover;background:#5A3A24"
											/>{:else}<span
												style="display:block;width:50px;height:50px;border-radius:14px;background:#5A3A24"
											></span>{/if}<span
											style="position:absolute;top:-6px;right:-6px;width:20px;height:20px;border-radius:50%;background:#F6D776;color:#2A1C12;font-size:11px;font-weight:700;display:grid;place-items:center"
											>{c.cantidad}</span
										></span
									><span style="flex:1;min-width:0;font-size:14px"
										>{c.producto.nombre}{#if c.variante}<span
												style="display:block;font-size:12.5px;color:#BFA48C">{c.variante}</span
											>{/if}</span
									><b style="font-size:14px;white-space:nowrap">{fmt(c.subtotal)}</b>
								</div>{/each}
						</div>
						<div
							style="margin-top:16px;padding-top:14px;border-top:1px solid rgba(251,241,230,.14);display:flex;flex-direction:column;gap:8px;font-size:14px"
						>
							<span style="display:flex;justify-content:space-between"
								><span>Subtotal</span><span>{fmt(subtotal)}</span></span
							><span style="display:flex;justify-content:space-between"
								><span>Envío</span><span
									>{elegida ? (cuenta.envio ? fmt(cuenta.envio) : 'Gratis') : '—'}</span
								></span
							>{#if cuenta.descCupon}<span
									style="display:flex;justify-content:space-between;color:#9FD9A6;font-weight:700"
									><span>Cupón {cupon.aplicado}</span><span>− {fmt(cuenta.descCupon)}</span></span
								>{/if}{#if descuento}<span
									style="display:flex;justify-content:space-between;color:#9FD9A6;font-weight:700"
									><span>Transferencia −{tienda.cobros.transferencia}%</span><span
										>− {fmt(descuento)}</span
									></span
								>{/if}
							<div style="margin:6px 0 2px">
								<CampoCupon
									{cupon}
									lineas={carrito.lineas}
									tinta="#F4EADB"
									suave="#E2CDB6"
									borde="rgba(244,234,219,.2)"
									fondo="rgba(244,234,219,.08)"
									acento="#F6D776"
									sobreAcento="#2A1C12"
									radio="12px"
									fuente="'Instrument Sans',sans-serif"
									error="#FFB4AB"
								/>
							</div>
							<span
								style="display:flex;justify-content:space-between;align-items:baseline;margin-top:6px"
								><b>Total</b><b
									style="font-family:'Schibsted Grotesk',sans-serif;font-weight:800;letter-spacing:-.03em;font-size:28px"
									>{fmt(total)}</b
								></span
							>
						</div>
						<div
							style="margin-top:16px;display:flex;align-items:center;gap:8px;font-size:12.5px;color:#E2CDB6"
						>
							<Icon icon={I.phShieldCheckFill} size={16} />Compra protegida · 10 días para
							arrepentirte
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
					style="display:inline-grid;place-items:center;width:64px;height:64px;border-radius:50%;background:var(--acento, #D9452B);color:#F4EADB"
					><Icon icon={I.phCheckBold} size={28} /></span
				>
				<h1
					style="margin:20px 0 0;font-family:'Schibsted Grotesk',sans-serif;font-weight:800;font-size:clamp(38px,5cqw,62px);line-height:1;letter-spacing:-.05em;overflow-wrap:anywhere"
				>
					¡Gracias, {pedido.nombre}!
				</h1>
				<p style="margin:14px 0 0;font-size:16px;line-height:1.6;color:#6A4F3C">
					Recibimos tu pedido <b>#{pedido.numero}</b> por <b>{fmt(pedido.total)}</b>. Te escribimos
					por mail con cada&nbsp;novedad.
				</p>
				<SenaPedido {tema} />
				{#if pedido.pago === 'transferencia'}
					<div
						style="margin-top:24px;padding:22px;border-radius:28px;background:#2A1C12;color:#F4EADB;text-align:left;font-size:14.5px;line-height:1.6"
					>
						<b>Transferí {fmt(pedido.total)}</b>
						{#if pedido.alias}al {pedido.esCbu ? 'CBU/CVU' : 'alias'}
							<b style="color:#F6D776">{pedido.alias}</b>{:else}con los datos que te mandamos por
							mail{/if} y mandanos el comprobante{#if whatsapp}&nbsp;<a
								href={wa(`¡Hola! Te mando el comprobante del pedido #${pedido.numero}.`)}
								target="_blank"
								rel="noopener"
								style="color:#F6D776">por WhatsApp</a
							>{:else}&nbsp;respondiendo el mail del pedido{/if}. Lo preparamos apenas
						se&nbsp;acredite.
					</div>
				{/if}
				<a
					href={base || '/'}
					style="display:inline-block;margin-top:28px;padding:15px 22px;border-radius:14px;background:#2A1C12;color:#F4EADB;font:700 15px 'Instrument Sans';text-decoration:none"
					>Volver a la tienda</a
				>
				{#if pedido.estado === 'Entregado'}
					<div style="margin-top:clamp(40px,6cqw,64px)">
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
				style="display:grid;padding:24px clamp(14px,3.5cqw,44px) clamp(48px,7cqw,90px);grid-template-columns:repeat(auto-fit,minmax(min(100%,360px),1fr));gap:clamp(16px,3cqw,40px);align-items:start"
			>
				<input type="hidden" name="servicioId" value={reserva.servicioId ?? ''} />
				<input type="hidden" name="profesionalId" value={reserva.profesionalId ?? ''} />
				<input type="hidden" name="inicio" value={reserva.inicio ?? ''} />
				<div style="display:flex;flex-direction:column;gap:22px;min-width:0">
					<h1
						style="margin:0;font-family:'Schibsted Grotesk',sans-serif;font-weight:800;font-size:clamp(38px,5cqw,62px);line-height:1;letter-spacing:-.05em"
					>
						Reservá tu turno
					</h1>
					<div>
						<div style="font-size:15px;font-weight:700;margin-bottom:10px">1 · Servicio</div>
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
									style="display:flex;align-items:center;gap:12px;padding:15px;text-align:left;border-radius:18px;border:1.5px solid {on
										? 'var(--acento, #D9452B)'
										: 'rgba(58,36,20,.1)'};background:{on
										? '#FFF6E0'
										: '#fff'};cursor:pointer;font-family:inherit;color:#2A1C12"
									><span
										style="width:20px;height:20px;border-radius:50%;border:2px solid {on
											? 'var(--acento, #D9452B)'
											: 'rgba(58,36,20,.25)'};display:grid;place-items:center;flex:none"
										>{#if on}<span
												style="width:10px;height:10px;border-radius:50%;background:var(--acento, #D9452B)"
											></span>{/if}</span
									><span style="flex:1;min-width:0"
										><span style="display:block;font-size:14.5px;font-weight:600">{s.nombre}</span
										><span style="display:block;font-size:12.5px;color:#7A5F4C"
											>{duracion(s.duracion)}</span
										></span
									><b style="font-size:14.5px;white-space:nowrap">{fmt(s.precio)}</b></button
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
									style="flex:none;width:112px;padding:10px;text-align:center;border-radius:18px;border:1.5px solid {on
										? 'var(--acento, #D9452B)'
										: 'rgba(58,36,20,.1)'};background:{on
										? '#FFF6E0'
										: '#fff'};cursor:pointer;font-family:inherit;color:#2A1C12"
									>{#if p?.foto}<img
											src={img(p.foto, 112, 112)}
											alt=""
											style="width:52px;height:52px;border-radius:50%;object-fit:cover;display:block;margin:0 auto"
										/>{:else}<span
											style="width:52px;height:52px;border-radius:50%;background:#FFF6E0;color:var(--acento, #D9452B);display:grid;place-items:center;margin:0 auto"
											><Icon icon={p ? I.phUserFill : I.phShuffleBold} size={22} /></span
										>{/if}<span
										style="display:block;margin-top:8px;font-size:13.5px;font-weight:700;overflow:hidden;text-overflow:ellipsis;white-space:nowrap"
										>{p?.nombre ?? 'Cualquiera'}</span
									><span
										style="display:block;font-size:11.5px;color:#7A5F4C;overflow:hidden;text-overflow:ellipsis;white-space:nowrap"
										>{p ? p.rol.split(' · ')[0] : 'El primero libre'}</span
									></button
								>{/each}
						</div>
					</div>
					<div>
						<div style="font-size:15px;font-weight:700;margin-bottom:10px">3 · Día y horario</div>
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
										: 0.35};border-radius:18px;border:1.5px solid {sel
										? 'var(--acento, #D9452B)'
										: 'rgba(58,36,20,.1)'};background:{sel
										? '#FFF6E0'
										: '#fff'};cursor:pointer;font-family:inherit;color:#2A1C12"
									><span
										style="display:block;font-size:11px;font-weight:600;text-transform:uppercase"
										>{d.w}</span
									><span style="display:block;font-size:17px;font-weight:700;margin-top:1px"
										>{d.n}</span
									></button
								>{/each}
						</div>
						<div style="margin-top:16px;display:flex;flex-direction:column;gap:12px">
							{#if !reserva.servicioId}<p style="margin:0;font-size:14px;color:#7A5F4C">
									Elegí un servicio para ver los horarios.
								</p>{:else if reserva.horarios && !reserva.grupos.length}<p
									style="margin:0;font-size:14px;color:#7A5F4C"
								>
									No quedan horarios este día. Probá con otro día u otro&nbsp;profesional.
								</p>{/if}
							{#each reserva.grupos as g (g.k)}<div>
									<div
										style="font-size:12.5px;font-weight:700;color:#7A5F4C;margin-bottom:7px;display:flex;align-items:center;gap:6px"
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
													: 0.4};border-radius:18px;border:1.5px solid {on
													? 'var(--acento, #D9452B)'
													: 'rgba(58,36,20,.1)'};background:{on
													? '#FFF6E0'
													: '#fff'};cursor:pointer;font-family:inherit;color:#2A1C12"
												>{h.hora}</button
											>{/each}
									</div>
								</div>{/each}
						</div>
					</div>
					{#if reserva.listo}<div>
							<div style="font-size:15px;font-weight:700;margin-bottom:10px">4 · Tus datos</div>
							<div
								style="display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));gap:8px"
							>
								{#each [['nombre', 'Nombre', 'text', 'name'], ['telefono', 'WhatsApp', 'tel', 'tel'], ['email', 'Email', 'email', 'email']] as [n, l, tipo, ac] (n)}
									<label
										style="display:block;padding:11px 14px;border-radius:16px;background:#fff;box-shadow:inset 0 0 0 1.5px rgba(58,36,20,.1)"
										><span style="display:block;font-size:11.5px;color:#9A8270">{l}</span><input
											name={n}
											type={tipo}
											autocomplete={ac as 'email'}
											required
											maxlength="120"
											style="width:100%;margin-top:2px;padding:0;border:0;background:transparent;font:400 14.5px 'Instrument Sans';color:#2A1C12;outline:none"
										/></label
									>
								{/each}
							</div>
						</div>
						<div>
							<div style="font-size:15px;font-weight:700;margin-bottom:10px">Pago</div>
							<div
								style="display:flex;align-items:center;gap:12px;padding:15px;text-align:left;border-radius:18px;border:1.5px solid var(--acento, #D9452B);background:#FFF6E0;cursor:pointer;font-family:inherit;color:#2A1C12"
							>
								<span
									style="width:20px;height:20px;border-radius:50%;border:2px solid var(--acento, #D9452B);display:grid;place-items:center;flex:none"
									><span
										style="width:10px;height:10px;border-radius:50%;background:var(--acento, #D9452B)"
									></span></span
								>{#if reserva.sena}<Icon
										icon={I.siMercadopago}
										size={22}
										style="color:#00B1EA;flex:none"
									/><span style="flex:1;min-width:0"
										><span style="display:block;font-size:14.5px;font-weight:600">Mercado Pago</span
										><span style="display:block;font-size:12.5px;color:#7A5F4C"
											>Seña de {fmt(reserva.sena)} para confirmar · el resto, en el local</span
										></span
									>{:else}<Icon
										icon={I.phStorefrontFill}
										size={22}
										style="color:var(--acento, #D9452B);flex:none"
									/><span style="flex:1;min-width:0"
										><span style="display:block;font-size:14.5px;font-weight:600"
											>Pagás en el local</span
										><span style="display:block;font-size:12.5px;color:#7A5F4C"
											>No hace falta seña: el turno queda confirmado</span
										></span
									>{/if}
							</div>
						</div>{/if}
				</div>
				<div
					style="position:sticky;top:90px;padding:22px;border-radius:28px;background:#2A1C12;color:#F4EADB;min-width:0"
				>
					<div
						style="font-family:'Schibsted Grotesk',sans-serif;font-weight:800;letter-spacing:-.03em;font-size:26px"
					>
						Tu reserva
					</div>
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
							<div style="font-size:13px;color:#E2CDB6">
								{reserva.cuando ?? 'Elegí el día y el horario'}
							</div>
						</div>
					</div>
					{#if reserva.servicio}<div
							style="margin-top:16px;padding-top:14px;border-top:1px solid rgba(251,241,230,.14);display:flex;flex-direction:column;gap:8px;font-size:14px"
						>
							<span style="display:flex;justify-content:space-between;gap:10px"
								><span>{reserva.servicio.nombre}</span><span style="white-space:nowrap"
									>{fmt(reserva.servicio.precio)}</span
								></span
							><span style="display:flex;justify-content:space-between;color:#E2CDB6"
								><span>Duración total</span><span>{reserva.duracion}</span></span
							><span
								style="display:flex;justify-content:space-between;align-items:baseline;margin-top:6px"
								><b>Total</b><b
									style="font-family:'Schibsted Grotesk',sans-serif;font-weight:800;letter-spacing:-.03em;font-size:28px"
									>{fmt(reserva.servicio.precio)}</b
								></span
							>
						</div>{/if}
					<div
						style="margin-top:14px;display:flex;gap:8px;font-size:13px;line-height:1.45;color:#E2CDB6"
					>
						<Icon icon={I.phShieldCheckFill} size={16} style="flex:none;margin-top:1px" /><span
							>{#if reserva.sena}Seña de {fmt(reserva.sena)} para confirmar. Podés cancelar desde el link
								del turno hasta {anticipacion} antes; la seña te la devuelve la&nbsp;tienda.{:else}No
								pagás nada ahora. Podés cancelar desde el link del turno hasta {anticipacion}&nbsp;antes.{/if}</span
						>
					</div>
					{#if error}<div
							role="alert"
							style="margin-top:12px;padding:12px 14px;border-radius:16px;background:rgba(217,69,43,.28);color:#FFD7CF;font-size:14px"
						>
							{error}
						</div>{/if}
					<button
						disabled={!reserva.listo || reservando || tienda.vitrina}
						style="margin-top:16px;width:100%;padding:17px;border:0;border-radius:999px;background:var(--acento, #D9452B);color:#F4EADB;font:700 15.5px 'Instrument Sans';cursor:pointer"
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

		<!-- ============ TURNO (con las piezas de la confirmación del pedido) ============ -->
		{#if pagina === 'turno' && turno}
			{@const cancelado = turno.estado === 'cancelado'}
			<section
				style="padding:clamp(40px,6cqw,80px) clamp(14px,3.5cqw,44px);max-width:720px;margin:0 auto;text-align:center"
			>
				<span
					style="display:inline-grid;place-items:center;width:64px;height:64px;border-radius:50%;background:var(--acento, #D9452B);color:#F4EADB"
					><Icon icon={cancelado ? I.phCalendarX : I.phCheckBold} size={28} /></span
				>
				<h1
					style="margin:20px 0 0;font-family:'Schibsted Grotesk',sans-serif;font-weight:800;font-size:clamp(38px,5cqw,62px);line-height:1;letter-spacing:-.05em;overflow-wrap:anywhere"
				>
					{cancelado
						? 'Turno cancelado'
						: turno.estado === 'reservado'
							? `Falta la seña, ${turno.nombre}`
							: `¡Listo, ${turno.nombre}!`}
				</h1>
				<p style="margin:14px 0 0;font-size:16px;line-height:1.6;color:#6A4F3C">
					{#if cancelado}Liberamos el horario{turno.senaPagada
							? '. La seña te la devuelve la tienda por Mercado Pago'
							: ''}. Si querés, reservá&nbsp;otro.{:else if turno.estado === 'reservado'}Te
						guardamos el horario unos minutos mientras pagás la seña con Mercado&nbsp;Pago.{:else}Tu
						turno está confirmado. Te esperamos en&nbsp;{turno.tienda}.{/if}
				</p>
				{#if aviso === 'pendiente' || aviso === 'error'}<div
						role="status"
						style="margin-top:16px;padding:12px 14px;border-radius:16px;background:rgba(217,69,43,.12);color:#9A2E18;font-size:14px"
					>
						{aviso === 'pendiente'
							? 'Mercado Pago todavía no acreditó la seña: te avisamos cuando entre.'
							: 'El pago de la seña no salió. Podés reservar de nuevo.'}
					</div>{/if}
				<div
					style="margin-top:24px;padding:22px;border-radius:28px;background:#2A1C12;color:#F4EADB;text-align:left;font-size:14.5px;line-height:1.6"
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
								style="color:var(--acento, #D9452B);flex:none"
							/>{turno.direccion}
						</div>{/if}
				</div>
				{#if error}<div
						role="alert"
						style="margin-top:16px;padding:12px 14px;border-radius:16px;background:rgba(217,69,43,.12);color:#9A2E18;font-size:14px"
					>
						{error}
					</div>{/if}
				<div style="margin-top:28px;display:flex;justify-content:center;gap:10px;flex-wrap:wrap">
					{#if !cancelado}<a
							href="{base}/turno/{turno.codigo}/turno.ics"
							download="turno.ics"
							data-sveltekit-reload
							style="display:flex;align-items:center;gap:8px;padding:15px 22px;border-radius:14px;background:#2A1C12;color:#F4EADB;font:700 15px 'Instrument Sans';text-decoration:none;border:0"
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
									style="cursor:pointer;padding:15px 22px;border-radius:14px;background:#2A1C12;color:#F4EADB;font:700 15px 'Instrument Sans';text-decoration:none;border:0"
									>Sí, cancelar el turno</button
								>{:else}<button
									type="button"
									onclick={() => (seguro = true)}
									style="cursor:pointer;padding:15px 20px;border:2px solid #2A1C12;border-radius:14px;background:transparent;color:#2A1C12;font:700 15px 'Instrument Sans';text-decoration:none"
									>Cancelar turno</button
								>{/if}
						</form>{:else if cancelado}<a
							href="{base}/turnos"
							style="padding:15px 22px;border-radius:14px;background:#2A1C12;color:#F4EADB;font:700 15px 'Instrument Sans';text-decoration:none;border:0"
							>Reservar otro</a
						>{/if}
					<a
						href={base || '/'}
						style="padding:15px 20px;border:2px solid #2A1C12;border-radius:14px;background:transparent;color:#2A1C12;font:700 15px 'Instrument Sans';text-decoration:none"
						>Volver a la tienda</a
					>
				</div>
				{#if !cancelado && !turno.cancelable}<p
						style="margin:14px 0 0;font-size:16px;line-height:1.6;color:#6A4F3C"
					>
						Para cambiarlo o cancelarlo, escribile a la&nbsp;tienda{#if whatsapp}&nbsp;<a
								href={wa(`¡Hola! Tengo un turno el ${cuandoLargo(turno.inicio)}.`)}
								target="_blank"
								rel="noopener"
								style="color:var(--acento, #D9452B)">por WhatsApp</a
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
				style="padding:clamp(18px,3cqw,40px) clamp(14px,3.5cqw,44px) 0;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,360px),1fr));gap:clamp(16px,3cqw,40px);align-items:center"
			>
				{#if tienda.portada.foto}<div
						style="position:relative;min-height:440px;border-radius:26px;transform:rotate(-1.5deg);overflow:hidden;background:#E9D3B5"
					>
						<img
							src={img(tienda.portada.foto, 1000)}
							alt=""
							style="position:absolute;inset:0;width:100%;height:100%;object-fit:cover"
						/>
					</div>{/if}
				<div style="min-width:0;display:flex;flex-direction:column;gap:18px">
					{#if tienda.ciudad}<div
							style="font-size:13px;font-weight:700;letter-spacing:.12em;text-transform:uppercase;color:var(--acento, #D9452B)"
						>
							{tienda.ciudad}
						</div>{/if}
					<h1
						style="margin:0;font-family:'Schibsted Grotesk',sans-serif;font-weight:800;font-size:clamp(42px,6cqw,80px);line-height:.98;letter-spacing:-.05em;overflow-wrap:anywhere"
					>
						{tienda.portada.titulo}
					</h1>
					<p style="margin:0;font-size:16px;line-height:1.6;color:#6A4F3C;max-width:500px">
						{tienda.nombre}{tienda.ciudad ? `, desde ${tienda.ciudad}` : ''}. {comoLlega(
							tienda.envios
						).replace('Comprás', 'Pedís')}
					</p>
					{#if tienda.productos.length || tienda.cobros.cuotas > 1 || tienda.promedio}<div
							style="display:grid;grid-template-columns:repeat(3,1fr);gap:10px"
						>
							{#if tienda.productos.length}<div
									style="padding:14px;border-radius:20px;background:#fff"
								>
									<div
										style="font-family:'Schibsted Grotesk',sans-serif;font-weight:800;letter-spacing:-.03em;font-size:clamp(24px,2.8cqw,34px)"
									>
										{tienda.productos.length}
									</div>
									<div style="font-size:12.5px;color:#7A5F4C;margin-top:2px">productos</div>
								</div>{/if}
							{#if tienda.cobros.cuotas > 1}<div
									style="padding:14px;border-radius:20px;background:#fff"
								>
									<div
										style="font-family:'Schibsted Grotesk',sans-serif;font-weight:800;letter-spacing:-.03em;font-size:clamp(24px,2.8cqw,34px)"
									>
										{tienda.cobros.cuotas}
									</div>
									<div style="font-size:12.5px;color:#7A5F4C;margin-top:2px">
										cuotas sin interés
									</div>
								</div>{/if}
							{#if tienda.promedio}<div style="padding:14px;border-radius:20px;background:#fff">
									<div
										style="font-family:'Schibsted Grotesk',sans-serif;font-weight:800;letter-spacing:-.03em;font-size:clamp(24px,2.8cqw,34px)"
									>
										{tienda.promedio.toLocaleString('es-AR')}
									</div>
									<div style="font-size:12.5px;color:#7A5F4C;margin-top:2px">
										de promedio en reseñas
									</div>
								</div>{/if}
						</div>{/if}
				</div>
			</section>
			<section
				style="padding:clamp(40px,6cqw,80px) clamp(14px,3.5cqw,44px);display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,300px),1fr));gap:12px"
			>
				<div
					style="padding:22px;border-radius:28px;background:#fff;display:flex;flex-direction:column;gap:12px"
				>
					<div
						style="font-family:'Schibsted Grotesk',sans-serif;font-weight:800;letter-spacing:-.03em;font-size:24px"
					>
						Cómo te llega
					</div>
					{#each llegadas as h (h.d)}<div
							style="display:flex;justify-content:space-between;gap:10px;padding-bottom:10px;border-bottom:1px solid rgba(58,36,20,.08);font-size:14.5px;color:#2A1C12"
						>
							<span>{h.d}</span><b style="font-weight:600;text-align:right">{h.h}</b>
						</div>{/each}
				</div>
				{#if tienda.direccion}<div
						style="padding:22px;border-radius:28px;background:#fff;display:flex;flex-direction:column;gap:12px"
					>
						<div
							style="font-family:'Schibsted Grotesk',sans-serif;font-weight:800;letter-spacing:-.03em;font-size:24px"
						>
							Dónde estamos
						</div>
						<span style="display:flex;gap:10px;font-size:14.5px"
							><Icon
								icon={I.phMapPinFill}
								size={18}
								style="color:var(--acento, #D9452B);flex:none"
							/>{tienda.direccion}</span
						>
					</div>{/if}
				{#if whatsapp}<div
						style="padding:22px;border-radius:28px;background:#2A1C12;color:#F4EADB;display:flex;flex-direction:column;gap:10px"
					>
						<div
							style="font-family:'Schibsted Grotesk',sans-serif;font-weight:800;letter-spacing:-.03em;font-size:24px"
						>
							Escribinos
						</div>
						<span style="font-size:14px;line-height:1.5;color:#E2CDB6"
							>Consultas y pedidos especiales, por&nbsp;WhatsApp.</span
						><span style="flex:1"></span><a
							href={wa('¡Hola!')}
							target="_blank"
							rel="noopener"
							style="display:flex;align-items:center;gap:10px;padding:14px;border-radius:16px;background:#25D366;color:#fff;font-weight:700;text-decoration:none"
							><Icon icon={I.phWhatsappLogoFill} size={22} />{tienda.whatsapp}</a
						>
					</div>{/if}
			</section>
		{/if}

		<!-- footer -->
		<footer
			style="padding:clamp(40px,6cqw,72px) clamp(14px,3.5cqw,44px) 26px;background:#2A1C12;color:#F4EADB"
		>
			<div
				style="display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,190px),1fr));gap:28px"
			>
				<div style="min-width:0;max-width:340px">
					<div
						style="font-family:'Schibsted Grotesk',sans-serif;font-weight:800;letter-spacing:-.03em;font-size:32px;overflow-wrap:anywhere"
					>
						{tienda.nombre}
					</div>
					<p style="margin:10px 0 0;font-size:14px;color:#E2CDB6;line-height:1.55">
						{#if tienda.portada.lema}{tienda.portada.lema}<br />{/if}{tienda.direccion}
					</p>
				</div>
				{#each foot as f (f.t)}<div
						style="display:flex;flex-direction:column;gap:9px;font-size:14px"
					>
						<b
							style="font-size:12px;letter-spacing:.12em;text-transform:uppercase;color:#F6D776;margin-bottom:4px"
							>{f.t}</b
						>{#each f.l as l (l.t)}<a href={l.href} style="color:#E8D8C6;text-decoration:none"
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
				style="margin-top:36px;padding-top:16px;border-top:1px solid rgba(251,241,230,.12);display:flex;justify-content:space-between;gap:14px;flex-wrap:wrap;font-size:12.5px;color:#BFA48C"
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

{#snippet tarjetaProducto(p: ReturnType<typeof tarjeta>, tam: number)}
	<a
		href={p.href}
		class="sube"
		style="display:block;padding:10px;border-radius:18px;background:#fff;text-decoration:none;text-align:left;font-family:inherit;color:#2A1C12;box-shadow:0 1px 2px rgba(58,36,20,.06)"
		><span
			style="position:relative;display:block;aspect-ratio:1;border-radius:20px;overflow:hidden;background:#F1DFC9"
			>{#if p.img}<img
					src={p.img}
					alt=""
					style="width:100%;height:100%;object-fit:cover"
				/>{/if}{#if p.tag}<span
					style="position:absolute;top:8px;left:8px;display:flex;align-items:center;gap:5px;padding:5px 9px;border-radius:999px;background:#F4EADB;font-size:11.5px;font-weight:700;color:{p
						.tag.sc}"><Icon icon={p.tag.si} size={13} />{p.tag.s}</span
				>{/if}</span
		><span style="display:block;padding:12px 6px 4px"
			><span
				style="display:block;font-family:'Schibsted Grotesk',sans-serif;font-weight:800;letter-spacing:-.03em;font-size:{tam}px;line-height:1.15"
				>{p.n}</span
			>{#if p.d}<span
					style="display:block;font-size:12.5px;color:#7A5F4C;margin-top:3px;white-space:nowrap;overflow:hidden;text-overflow:ellipsis"
					>{p.d}</span
				>{/if}<span
				style="display:flex;justify-content:space-between;align-items:center;gap:6px;margin-top:10px"
				><span style="min-width:0"
					><b style="font-size:16px">{p.price}</b>{#if p.old}<s
							style="margin-left:6px;opacity:.6;font-size:13px">{p.old}</s
						>{/if}</span
				><span
					style="width:36px;height:36px;flex:none;border-radius:50%;background:#2A1C12;color:#F4EADB;display:grid;place-items:center"
					><Icon icon={I.phPlusBold} size={15} /></span
				></span
			></span
		></a
	>
{/snippet}

<!-- eslint-enable svelte/no-navigation-without-resolve -->

<style>
	.miga :global(a:hover) {
		color: var(--acento, #d9452b);
	}
	.sube {
		transition: transform 0.25s;
	}
	.sube:hover {
		transform: translateY(-3px);
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
