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
	import { cuando, proximamente } from '../lanzamiento.ts';
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
		resumenMoto,
		textoEnvios
	} from '../envio.svelte.ts';
	import { Seleccion, tarjetaDe } from '../seleccion.svelte.ts';
	import { partirTitulo } from '../titulo.ts';
	import { Reserva } from '../reserva.svelte.ts';
	import { cuandoLargo, duracion } from '#lib/turnos/formato.ts';
	import type { OpcionEnvio, ProductoPublico, TiendaPublica } from '../tipos.ts';

	/**
	 * Plantilla Pampa (design/Plantilla Pampa.dc.html), con los datos de la tienda en lugar de los de
	 * ejemplo. Lo que el diseño muestra sin datos atrás (talles, colores, combo "fit completo", cuenta
	 * regresiva del drop, reserva de stock, aviso del próximo drop, cupón, seguidores) no se inventa:
	 * esas partes no se dibujan.
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
	// Drops y reserva (§52): el lanzamiento (o la oferta con fin) que se cuenta en la portada, los
	// productos que todavía no salieron y «Avisame» del próximo.
	const drop = $derived(conteo.actual);
	const proximos = $derived(
		tienda.productos
			.filter((p) => proximamente(p, conteo.ahora))
			.sort((a, b) => Date.parse(a.lanzamiento!) - Date.parse(b.lanzamiento!))
	);
	let avisoEmail = $state('');
	let avisoListo = $state<string | null>(null);
	let avisoError = $state<string | null>(null);
	async function avisarDrop(e: SubmitEvent, productoId: string) {
		e.preventDefault();
		avisoError = null;
		const email = avisoEmail.trim();
		if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) return (avisoError = 'Revisá el email.');
		const r = await fetch(`/t/${tienda.slug}/aviso`, {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ productoId, email })
		}).catch(() => null);
		if (r?.ok) avisoListo = email;
		else
			avisoError =
				((await r?.json().catch(() => null)) as { error?: string } | null)?.error ??
				'No se pudo. Probá de nuevo.';
	}
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
	// Con variantes de precios distintos, "Desde"; "Agotado" sólo si no queda ninguna.
	const tarjeta = (p: ProductoPublico) => {
		const t = tarjetaDe(p);
		const out = t.agotado;
		const sale = !!p.anterior && !t.desde;
		return {
			id: p.id,
			n: p.nombre,
			price: (t.desde ? 'Desde ' : '') + fmt(t.precio),
			img: img(p.foto, 640, 800),
			tag: out ? 'Agotado' : sale ? `−${off(p)}%` : p.etiqueta,
			out,
			op: out ? 0.4 : 1,
			flt: out ? 'grayscale(1)' : 'none',
			tbg: out
				? '#2A2A2A'
				: sale
					? '#FF4D2E'
					: p.etiqueta === 'Últimos'
						? '#F2F2F2'
						: 'var(--acento, #E6FF4D)',
			tfg: out ? '#8A8A8A' : '#111',
			pc: sale ? '#FF4D2E' : '#F2F2F2',
			sub: out
				? 'Sin stock por ahora'
				: sale
					? tienda.cobros.cuotas > 1
						? `Antes ${fmt(p.anterior!)} · ${tienda.cobros.cuotas} cuotas`
						: `Antes ${fmt(p.anterior!)}`
					: tienda.cobros.cuotas > 1
						? `${tienda.cobros.cuotas} × ${fmt(t.precio / tienda.cobros.cuotas)} sin interés`
						: '',
			href: `${base}/producto/${p.slug ?? p.id}`
		};
	};
	type Tarjeta = ReturnType<typeof tarjeta>;
	// Los títulos gigantes del diseño son cortos ("Drop 07"): con textos largos, el cuerpo se achica
	// para que la palabra más larga entre en su columna (ancho en cqw) sin cortarse por la mitad.
	const cabe = (texto: string, cqw: number) =>
		`${(cqw / (Math.max(1, ...texto.split(/\s+/).map((w) => w.length)) * 0.74)).toFixed(2)}cqw`;
	const ver = (s: string) => tienda.secciones.includes(s as TiendaPublica['secciones'][number]);
	const whatsapp = $derived(tienda.whatsapp.replace(/\D/g, ''));
	const wa = (texto: string) =>
		`https://wa.me/${whatsapp.startsWith('54') ? whatsapp : '54' + whatsapp}?text=${encodeURIComponent(texto)}`;

	// Título de la portada: el segundo tramo va en amarillo, como en el diseño ("Drop / 07").
	const titulo = $derived(partirTitulo(tienda.portada.titulo));

	// Cinta que corre arriba de todo, con lo que la tienda ofrece de verdad.
	const cinta = $derived([
		...(tienda.envios.gratisDesde ? [`Envío gratis desde ${fmt(tienda.envios.gratisDesde)}`] : []),
		...(tienda.envios.moto ? ['Envío en moto'] : []),
		tienda.cobros.cuotas > 1 ? `${tienda.cobros.cuotas} cuotas sin interés` : 'Compra protegida',
		...(tienda.cobros.transferencia > 0
			? [`${tienda.cobros.transferencia}% off por transferencia`]
			: []),
		...(tienda.envios.retiro ? ['Retiro en el local'] : [])
	]);
	const perks = $derived([
		tienda.envios.moto
			? { i: I.phMopedBold, t: 'Envío en moto', d: resumenMoto(tienda.envios) }
			: tienda.envios.correo || !tienda.envios.retiro
				? { i: I.phTruckBold, t: 'Envíos', d: entregaCorta(tienda.envios) }
				: { i: I.phStorefrontBold, t: 'Retiro', d: 'En el local, sin costo' },
		...(tienda.cobros.cuotas > 1
			? [{ i: I.phCreditCardBold, t: `${tienda.cobros.cuotas} cuotas`, d: 'Sin interés' }]
			: []),
		...(tienda.cobros.transferencia > 0
			? [{ i: I.phBankBold, t: `${tienda.cobros.transferencia}% off`, d: 'Por transferencia' }]
			: []),
		tienda.envios.gratisDesde
			? { i: I.phTruckBold, t: 'Envío gratis', d: `Desde ${fmt(tienda.envios.gratisDesde)}` }
			: { i: I.phArrowsClockwiseBold, t: 'Arrepentimiento', d: '10 días para devolverlo' }
	]);
	// Las cifras de Nosotros: sólo las que hay (nunca «0 productos»).
	const cifras = $derived([
		...(tienda.productos.length ? [{ v: String(tienda.productos.length), t: 'Productos' }] : []),
		...(tienda.cobros.cuotas > 1
			? [{ v: String(tienda.cobros.cuotas), t: 'Cuotas sin interés' }]
			: []),
		...(tienda.promedio
			? [{ v: tienda.promedio.toLocaleString('es-AR'), t: 'En reseñas' }]
			: tienda.cobros.transferencia > 0
				? [{ v: `${tienda.cobros.transferencia}%`, t: 'Off por transferencia' }]
				: [])
	]);
	const best = $derived(
		tienda.masVendidos.slice(0, 8).flatMap((id) => (porId.get(id) ? [tarjeta(porId.get(id)!)] : []))
	);
	const destacado = $derived(
		tienda.masVendidos.map((id) => porId.get(id)).find((p) => p && !tarjetaDe(p).agotado) ?? null
	);
	const destacadoPrecio = $derived(destacado ? tarjetaDe(destacado) : null);
	const maxOff = $derived(Math.max(0, ...tienda.productos.map(off)));

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
	const tabs = $derived([
		{ c: null, t: 'Todo', n: tienda.productos.length },
		...tienda.categorias.map((c) => ({ c: c.nombre, t: c.nombre, n: c.cantidad }))
	]);
	const tituloCat = $derived(categoria ? partirTitulo(categoria) : ['Todo el', 'catálogo']);

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
	// Talle, color, presentación…: lo elegido, la variante, su precio, stock y foto (seleccion.svelte.ts).
	const sel = $derived(producto ? new Seleccion(producto) : null);
	// El tachado no va con "Desde": no se sabe contra qué variante se compara.
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
	/** Sumar desde "Sumale": con opciones, primero se elige en la ficha. */
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
			.slice(0, 4)
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
							? `${tienda.cobros.transferencia}% off · alias al confirmar`
							: 'Alias al confirmar',
						i: I.phBankBold
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
						i: I.siMercadopago
					}
				]
			: [])
	]);
	// Sin datos de transferencia cargados, el medio de pago es el primero disponible.
	$effect(() => {
		if (pagos.length && !pagos.some((x) => x.id === pago)) pago = pagos[0].id;
	});
	const pasos = $derived([
		{ t: 'Datos', on: true },
		{ t: 'Entrega', on: !!envio },
		{ t: 'Pago', on: !!envio }
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
				{ t: 'Cambios', href: `/t/${tienda.slug}/arrepentimiento` },
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
		...tienda.categorias.slice(0, 3).map((c) => ({
			t: c.nombre,
			href: enlaceCat(c.nombre),
			on: pagina === 'catalogo' && categoria === c.nombre
		})),
		{ t: 'Nosotros', href: `${base}/nosotros`, on: pagina === 'nosotros' }
	]);
	// Reseñas y preguntas con las piezas de Pampa: afiche oscuro, esquinas rectas, verde lima.
	const tema: TemaOpiniones = {
		texto: '#F2F2F2',
		suave: '#8A8A8A',
		campo: '#111',
		tarjeta: '#161616',
		borde: '#2A2A2A',
		acento: 'var(--acento, #E6FF4D)',
		sobreAcento: '#111',
		radio: '0px',
		fuente: "'Archivo', system-ui, sans-serif",
		titulos: "'Archivo Black', sans-serif",
		pesoTitulos: 400,
		mayusculas: true
	};
	const estrellas = (n: number) =>
		Array.from({ length: 5 }, (_, i) =>
			n >= i + 1 ? I.phStarFill : n >= i + 0.5 ? I.phStarHalfFill : I.phStar
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
	// El nombre gigante del pie: que la palabra más larga entre entera en el renglón (antes se cortaba
	// a mitad de palabra, «ALMAC / EN»). En cqw, con el ancho de una letra de esta tipografía.
	const pieNombre = $derived(
		Math.min(22, Math.floor(120 / Math.max(4, ...tienda.nombre.split(/\s+/).map((w) => w.length))))
	);
</script>

<svelte:window bind:innerWidth={ancho} />

<svelte:head>
	<link
		href="https://fonts.googleapis.com/css2?family=Archivo+Black&family=Archivo:wght@400;500;600;700;800&display=swap"
		rel="stylesheet"
	/>
</svelte:head>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
{#snippet card(p: Tarjeta, cuadrada: boolean, sub: boolean)}
	<a
		href={p.href}
		class="celda"
		style="position:relative;display:block;border-right:2px solid #2A2A2A;border-bottom:2px solid #2A2A2A;padding:0;text-align:left;color:#F2F2F2;text-decoration:none;min-width:0"
	>
		<span
			class="zoom"
			style="position:relative;display:block;aspect-ratio:{cuadrada
				? '1'
				: '4/5'};overflow:hidden;background:#1E1E1E"
			>{#if p.img}<img
					src={p.img}
					alt={p.n}
					style="width:100%;height:100%;object-fit:cover;display:block;opacity:{p.op};filter:{p.flt}"
				/>{/if}{#if p.tag && !cuadrada}<span
					style="position:absolute;top:10px;left:10px;padding:5px 8px;background:{p.tbg};color:{p.tfg};font:800 11px 'Archivo';letter-spacing:.1em;text-transform:uppercase"
					>{p.tag}</span
				>{/if}{#if p.out && !cuadrada}<span
					style="position:absolute;inset:0;display:grid;place-items:center"
					><span
						style="padding:10px 16px;border:2px solid #F2F2F2;background:rgba(17,17,17,.7);font:800 13px 'Archivo';letter-spacing:.14em;text-transform:uppercase;transform:rotate(-6deg)"
						>Agotado</span
					></span
				>{/if}</span
		>
		<span style="display:block;padding:12px 14px {sub ? '14px' : '12px'}"
			><span style="display:flex;justify-content:space-between;gap:10px;align-items:baseline"
				><span
					style="font-size:{sub
						? '14px'
						: '13.5px'};font-weight:700;text-transform:uppercase;letter-spacing:.02em;min-width:0"
					>{p.n}</span
				><b style="font-size:14.5px;white-space:nowrap;color:{p.pc}">{p.price}</b></span
			>{#if sub && p.sub}<span style="display:block;margin-top:8px;font-size:12px;color:#8A8A8A"
					>{p.sub}</span
				>{/if}</span
		>
	</a>
{/snippet}

<div
	class="pampa"
	style="min-height:100vh;font-family:'Archivo',system-ui,sans-serif;color:#F2F2F2;background:#0B0B0B"
>
	<div
		style="container-type:inline-size;width:100%;overflow-x:clip;background:#111;position:relative"
	>
		<!-- cinta -->
		<div
			style="overflow:hidden;background:var(--acento, #E6FF4D);color:#111;border-bottom:2px solid #111"
		>
			<div class="mq" style="display:flex;width:max-content">
				{#each [0, 1] as k (k)}<span style="display:flex;flex:none" aria-hidden={k === 1}
						>{#each cinta as t (t)}<span
								style="padding:9px 22px;font:800 12.5px 'Archivo';letter-spacing:.14em;text-transform:uppercase;white-space:nowrap"
								>{t}&nbsp;&nbsp;✦</span
							>{/each}</span
					>{/each}
			</div>
		</div>

		<!-- header -->
		<header
			style="position:sticky;top:0;z-index:20;display:flex;align-items:stretch;background:#111;border-bottom:2px solid #2A2A2A;height:58px"
		>
			<a
				href={base || '/'}
				style="display:flex;align-items:center;min-width:0;border-right:2px solid #2A2A2A;padding:0 clamp(14px,2.4cqw,28px);color:#F2F2F2;text-decoration:none;font-family:'Archivo Black',sans-serif;font-size:{nar
					? 19
					: 24}px;letter-spacing:-.02em;text-transform:uppercase"
				><span style="overflow:hidden;text-overflow:ellipsis;white-space:nowrap"
					><LogoTienda {tienda} /></span
				></a
			>
			{#if wide}<nav style="display:flex">
					{#each nav as n (n.t)}<a
							href={n.href}
							style="display:flex;align-items:center;border-right:2px solid #2A2A2A;padding:0 20px;background:{n.on
								? '#F2F2F2'
								: 'transparent'};color:{n.on
								? '#111'
								: '#F2F2F2'};text-decoration:none;font:800 12.5px 'Archivo';letter-spacing:.12em;text-transform:uppercase;white-space:nowrap"
							>{n.t}</a
						>{/each}
				</nav>{/if}
			<span style="flex:1"></span>
			<a
				href="{base}/carrito"
				style="display:flex;align-items:center;gap:10px;flex:none;border-left:2px solid #2A2A2A;background:var(--acento, #E6FF4D);color:#111;padding:0 clamp(14px,2cqw,22px);text-decoration:none;font:800 13px 'Archivo';letter-spacing:.1em;text-transform:uppercase;white-space:nowrap"
				><Icon
					icon={I.phBagSimpleFill}
					size={18}
				/>Bolsa{#if carrito.unidades}&nbsp;({carrito.unidades}){/if}</a
			>
			{#if nar}<a
					href="{base}/catalogo"
					aria-label="Catálogo"
					style="display:grid;place-items:center;flex:none;width:58px;border-left:2px solid #2A2A2A;color:#F2F2F2"
					><Icon icon={I.phListBold} size={22} /></a
				>{/if}
			{#if nar && tienda.turnos}<a
					href="{base}/turnos"
					aria-label="Reservar turno"
					style="display:grid;place-items:center;flex:none;width:58px;border-left:2px solid #2A2A2A;background:{pagina ===
					'turnos'
						? '#F2F2F2'
						: 'transparent'};color:{pagina === 'turnos' ? '#111' : '#F2F2F2'}"
					><Icon icon={I.phCalendarPlusBold} size={22} /></a
				>{/if}
		</header>

		{#if tienda.vitrina}
			<div
				style="padding:10px 16px;background:#1A1A1A;border-bottom:2px solid #2A2A2A;color:var(--acento, #E6FF4D);font:800 12px 'Archivo';letter-spacing:.1em;text-transform:uppercase;text-align:center"
			>
				Es una tienda de ejemplo de la plantilla Pampa: podés recorrerla, pero no se compra
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
							style="position:relative;border-bottom:2px solid #2A2A2A;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,420px),1fr))"
						>
							<div
								style="padding:clamp(22px,4cqw,56px);display:flex;flex-direction:column;justify-content:space-between;gap:28px;min-width:0;border-right:2px solid #2A2A2A"
							>
								{#if (drop && drop.producto.stock > 0) || tienda.portada.lema}<div
										style="display:flex;align-items:center;gap:10px;font-size:12.5px;font-weight:800;letter-spacing:.14em;text-transform:uppercase"
									>
										<span
											style="width:10px;height:10px;background:#FF4D2E;border-radius:50%;flex:none"
										></span>{#if drop && drop.producto.stock > 0}{drop.tipo === 'lanzamiento'
												? 'Próximo drop'
												: 'En vivo'} · {drop.tipo === 'lanzamiento' ? '' : 'quedan '}{drop.producto
												.stock}
											{drop.producto.stock === 1 ? 'unidad' : 'unidades'}{:else}{tienda.portada
												.lema}{/if}
									</div>{/if}
								<h1
									style="margin:0;font-family:'Archivo Black',sans-serif;font-weight:400;font-size:min(clamp(68px,13cqw,190px),{cabe(
										tienda.portada.titulo,
										wide ? 42 : 88
									)});line-height:.82;letter-spacing:-.045em;text-transform:uppercase;overflow-wrap:anywhere"
								>
									{titulo[0]}{#if titulo[1]}<br /><span style="color:var(--acento, #E6FF4D)"
											>{titulo[1]}</span
										>{/if}
								</h1>
								<div>
									{#if tienda.ciudad || tienda.cobros.cuotas > 1 || tienda.cobros.transferencia > 0}<p
											style="margin:0 0 22px;font-size:17px;line-height:1.5;color:#BDBDBD;max-width:420px"
										>
											{tienda.nombre}{tienda.ciudad
												? `, desde ${tienda.ciudad}`
												: ''}.{#if tienda.cobros.cuotas > 1}&nbsp;{tienda.cobros.cuotas} cuotas sin interés{#if tienda.cobros.transferencia > 0}&nbsp;o
													{tienda.cobros.transferencia}% off pagando por&nbsp;transferencia{/if}.{:else if tienda.cobros.transferencia > 0}&nbsp;{tienda
													.cobros.transferencia}% off pagando por&nbsp;transferencia.{/if}
										</p>{/if}
									{#if drop}
										<div
											style="display:grid;grid-template-columns:repeat(4,minmax(0,1fr));border:2px solid #2A2A2A;max-width:440px"
										>
											{#each conteo.partes(4, ['DÍAS', 'HS', 'MIN', 'SEG']) as c (c.u)}<div
													style="padding:10px 0;text-align:center;border-right:2px solid #2A2A2A"
												>
													<div
														style="font-family:'Archivo Black',sans-serif;font-size:clamp(26px,3.4cqw,38px);line-height:1;font-variant-numeric:tabular-nums"
													>
														{c.v}
													</div>
													<div
														style="margin-top:4px;font-size:10.5px;font-weight:800;letter-spacing:.14em;color:#8A8A8A"
													>
														{c.u}
													</div>
												</div>{/each}
										</div>
										<div style="margin-top:12px;font-size:12.5px;color:#8A8A8A">
											{#if drop.tipo === 'lanzamiento'}{drop.producto.nombre} sale el {cuando(
													drop.hasta
												)}{#if drop.producto.reserva}&nbsp;· reservalo con {drop.producto.reserva}%
													de&nbsp;seña{/if}{:else}{drop.producto.nombre}: cierra el {cuando(
													drop.hasta
												)} · después vuelve a precio&nbsp;full{/if}
										</div>
									{/if}
									<div style="display:flex;gap:0;flex-wrap:wrap;margin-top:{drop ? '22px' : '0'}">
										<a
											href={drop
												? `${base}/producto/${drop.producto.slug ?? drop.producto.id}`
												: `${base}/catalogo`}
											style="padding:18px 28px;border:2px solid var(--acento, #E6FF4D);background:var(--acento, #E6FF4D);color:#111;font:800 14px 'Archivo';letter-spacing:.12em;text-transform:uppercase;text-decoration:none"
											>{drop?.tipo === 'lanzamiento'
												? 'Ver el drop →'
												: drop
													? 'Comprar el drop →'
													: 'Ver todo →'}</a
										><a
											href="{base}/nosotros"
											style="padding:18px 24px;border:2px solid #F2F2F2;border-left-width:0;background:transparent;color:#F2F2F2;font:800 14px 'Archivo';letter-spacing:.12em;text-transform:uppercase;text-decoration:none"
											>Nosotros</a
										>
									</div>
								</div>
							</div>
							{#if tienda.portada.foto || destacado || maxOff}<div
									style="position:relative;min-height:min(640px,120cqw);background:#1E1E1E"
								>
									{#if tienda.portada.foto}<img
											src={img(tienda.portada.foto, 1200)}
											alt=""
											style="position:absolute;inset:0;width:100%;height:100%;object-fit:cover;filter:grayscale(.25) contrast(1.05)"
										/>{/if}
									{#if maxOff}<span
											style="position:absolute;top:16px;right:16px;padding:8px 12px;background:#FF4D2E;color:#111;font:800 12px 'Archivo';letter-spacing:.12em;text-transform:uppercase;transform:rotate(3deg)"
											>Hasta −{maxOff}%</span
										>{/if}
									{#if destacado}<div
											style="position:absolute;left:0;right:0;bottom:0;display:flex;align-items:center;gap:12px;padding:14px 16px;background:#111;border-top:2px solid #2A2A2A"
										>
											{#if destacado.foto}<img
													src={img(destacado.foto, 120, 120)}
													alt=""
													style="width:46px;height:46px;object-fit:cover;flex:none"
												/>{/if}<span style="flex:1;min-width:0"
												><span style="display:block;font-size:14px;font-weight:700"
													>{destacado.nombre}</span
												><span style="display:block;font-size:12.5px;color:#8A8A8A"
													>{destacado.stock <= 3
														? destacado.stock === 1
															? 'Queda 1'
															: `Quedan ${destacado.stock}`
														: 'El más vendido'}</span
												></span
											><a
												href="{base}/producto/{destacado.slug ?? destacado.id}"
												style="padding:10px 14px;border:0;background:#F2F2F2;color:#111;font:800 12px 'Archivo';letter-spacing:.1em;text-transform:uppercase;text-decoration:none;white-space:nowrap"
												>{destacadoPrecio?.desde ? 'Desde ' : ''}{fmt(
													destacadoPrecio?.precio ?? destacado.precio
												)}</a
											>
										</div>{/if}
								</div>{/if}
						</section>
					{/if}

					<div
						style="display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));border-bottom:2px solid #2A2A2A"
					>
						{#each perks as p (p.t)}<div
								style="display:flex;align-items:center;gap:12px;padding:16px clamp(14px,2.4cqw,28px);border-right:2px solid #2A2A2A"
							>
								<Icon icon={p.i} size={22} style="color:var(--acento, #E6FF4D);flex:none" />
								<div>
									<div
										style="font-size:12.5px;font-weight:800;letter-spacing:.1em;text-transform:uppercase"
									>
										{p.t}
									</div>
									<div style="font-size:12.5px;color:#8A8A8A;margin-top:2px">{p.d}</div>
								</div>
							</div>{/each}
					</div>
				{:else if seccion === 'cats'}
					{#if ver('cats') && tienda.categorias.length}
						<div
							style="display:flex;overflow-x:auto;scrollbar-width:none;border-bottom:2px solid #2A2A2A"
						>
							{#each tabs as t (t.c ?? '')}<a
									href={enlaceCat(t.c)}
									style="flex:none;padding:14px 18px;border-right:2px solid #2A2A2A;color:#F2F2F2;text-decoration:none;font:800 12.5px 'Archivo';letter-spacing:.12em;text-transform:uppercase;white-space:nowrap"
									>{t.t} <span style="opacity:.55">{t.n}</span></a
								>{/each}
						</div>
					{/if}
				{:else if seccion === 'prods'}
					{#if proximos.length}
						<section style="padding:clamp(40px,6cqw,80px) 0 0">
							<div
								style="display:flex;justify-content:space-between;align-items:end;gap:16px;flex-wrap:wrap;padding:0 clamp(14px,2.4cqw,28px) 18px"
							>
								<h2
									style="margin:0;font-family:'Archivo Black',sans-serif;font-weight:400;font-size:clamp(40px,6.4cqw,88px);line-height:.85;letter-spacing:-.04em;text-transform:uppercase"
								>
									El drop
								</h2>
								<span
									style="font-size:12.5px;font-weight:800;letter-spacing:.12em;text-transform:uppercase;color:#8A8A8A"
									>{proximos.length}
									{proximos.length === 1
										? 'pieza'
										: 'piezas'}{#if proximos.some((p) => p.stock > 0)}&nbsp;· {proximos.reduce(
											(s, p) => s + p.stock,
											0
										)}
										unidades{/if}</span
								>
							</div>
							<div
								style="display:grid;grid-template-columns:repeat(auto-fill,minmax(min(50%,260px),1fr));border-top:2px solid #2A2A2A"
							>
								{#each proximos as p (p.id)}{@render card(tarjeta(p), false, true)}{/each}
							</div>
							<div
								style="margin:2px clamp(14px,2.4cqw,28px) 0;display:flex;flex-wrap:wrap;align-items:center;gap:14px;padding:clamp(20px,3cqw,32px);background:#1A1A1A;border:2px solid #2A2A2A"
							>
								<div style="flex:1 1 280px">
									<div
										style="font-family:'Archivo Black',sans-serif;font-size:clamp(24px,3cqw,34px);text-transform:uppercase;letter-spacing:-.02em;overflow-wrap:anywhere"
									>
										Avisame {proximos.length === 1 ? 'del drop' : `de ${proximos[0].nombre}`}
									</div>
									<div style="font-size:14px;color:#8A8A8A;margin-top:4px">
										{#if avisoListo}Listo: te escribimos a {avisoListo} cuando salga.{:else}Te
											escribimos por mail cuando salga. Nada de spam.{/if}
									</div>
									{#if avisoError}<div
											style="margin-top:6px;font-size:13px;color:#FF4D2E"
											role="alert"
										>
											{avisoError}
										</div>{/if}
								</div>
								{#if !avisoListo}<form
										onsubmit={(e) => avisarDrop(e, proximos[0].id)}
										novalidate
										style="flex:1 1 320px;display:flex;border:2px solid #F2F2F2;min-width:0"
									>
										<label for="aviso-drop" class="oculto">Tu email para el aviso</label><input
											id="aviso-drop"
											type="email"
											autocomplete="email"
											maxlength="200"
											placeholder="Tu email"
											bind:value={avisoEmail}
											disabled={tienda.vitrina}
											style="flex:1;min-width:0;padding:14px;border:0;background:transparent;font:400 14.5px 'Archivo';color:#F2F2F2;outline:none"
										/><button
											type="submit"
											disabled={tienda.vitrina}
											style="display:flex;align-items:center;padding:0 18px;border:0;background:var(--acento, #E6FF4D);color:#111;font:800 12.5px 'Archivo';letter-spacing:.1em;text-transform:uppercase;gap:7px;cursor:pointer;min-height:48px"
											><Icon icon={I.phBellRingingBold} size={17} />Avisame</button
										>
									</form>{/if}
							</div>
						</section>
					{/if}
					{#if ver('prods') && best.length}
						<section style="padding:clamp(40px,6cqw,80px) 0 0">
							<div
								style="display:flex;justify-content:space-between;align-items:end;gap:16px;flex-wrap:wrap;padding:0 clamp(14px,2.4cqw,28px) 18px"
							>
								<h2
									style="margin:0;font-family:'Archivo Black',sans-serif;font-weight:400;font-size:clamp(40px,6.4cqw,88px);line-height:.85;letter-spacing:-.04em;text-transform:uppercase"
								>
									Lo más pedido
								</h2>
								<a
									href="{base}/catalogo"
									style="font-size:12.5px;font-weight:800;letter-spacing:.12em;text-transform:uppercase;color:#8A8A8A;text-decoration:none"
									>{tienda.productos.length}
									{tienda.productos.length === 1 ? 'producto' : 'productos'} →</a
								>
							</div>
							<div
								style="display:grid;grid-template-columns:repeat(auto-fill,minmax(min(50%,260px),1fr));border-top:2px solid #2A2A2A"
							>
								{#each best as p (p.id)}{@render card(p, false, true)}{/each}
							</div>
						</section>
					{/if}
				{:else if seccion === 'revs'}
					{#if ver('revs') && tienda.resenas.length}
						<section style="padding:clamp(40px,6cqw,80px) clamp(14px,2.4cqw,28px) 0">
							<div
								style="display:flex;justify-content:space-between;align-items:end;gap:16px;flex-wrap:wrap"
							>
								<h2
									style="margin:0;font-family:'Archivo Black',sans-serif;font-weight:400;font-size:clamp(40px,6.4cqw,88px);line-height:.85;letter-spacing:-.04em;text-transform:uppercase"
								>
									Lo que dicen
								</h2>
								{#if tienda.promedio}<span
										style="display:flex;align-items:center;gap:8px;font-size:14px;font-weight:700"
										><span style="display:flex;color:var(--acento, #E6FF4D)"
											>{#each estrellas(tienda.promedio) as e, i (i)}<Icon
													icon={e}
													size={15}
												/>{/each}</span
										>{tienda.promedio.toLocaleString('es-AR')} · {tienda.resenas.length}
										{tienda.resenas.length === 1 ? 'reseña' : 'reseñas'}</span
									>{/if}
							</div>
							<div
								style="margin-top:22px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,300px),1fr));border:2px solid #2A2A2A"
							>
								{#each tienda.resenas.slice(0, 3) as r, i (i)}
									{@const pr = tienda.productos.find((p) => p.nombre === r.producto)}
									<div style="display:flex;gap:14px;padding:16px;border-right:2px solid #2A2A2A">
										{#if pr?.foto}<img
												src={img(pr.foto, 200, 250)}
												alt=""
												style="width:92px;height:116px;object-fit:cover;flex:none;background:#1E1E1E"
											/>{/if}
										<div style="min-width:0">
											{#if r.producto}<div
													style="margin-bottom:8px;font-size:12px;font-weight:800;letter-spacing:.1em;text-transform:uppercase;color:var(--acento, #E6FF4D)"
												>
													{r.producto}
												</div>{/if}
											<p style="margin:0;font-size:15px;line-height:1.4;font-weight:500">
												“{r.texto}”
											</p>
											<div style="margin-top:8px;font-size:12.5px;color:#8A8A8A">{r.autor}</div>
											{#if r.conCompra}<div
													style="margin-top:8px;display:flex;align-items:center;gap:5px;font-size:11px;font-weight:800;letter-spacing:.12em;text-transform:uppercase;color:#B9F27C"
												>
													<Icon icon={I.phSealCheckFill} size={13} />Compra verificada
												</div>{/if}
											{#if r.respuesta}<div
													style="margin-top:10px;padding:10px 12px;border-left:2px solid var(--acento, #E6FF4D);background:#1E1E1E;font-size:13px;line-height:1.5;color:#BDBDBD"
												>
													<b
														style="display:block;font-size:11px;letter-spacing:.12em;text-transform:uppercase;color:#F2F2F2"
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
						<section style="padding:clamp(40px,6cqw,80px) clamp(14px,2.4cqw,28px)">
							<h2
								style="margin:0;font-family:'Archivo Black',sans-serif;font-weight:400;font-size:clamp(40px,6.4cqw,88px);line-height:.85;letter-spacing:-.04em;text-transform:uppercase"
							>
								En la calle
							</h2>
							<div
								style="margin-top:20px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(30%,170px),1fr));gap:2px;background:#2A2A2A;border:2px solid #2A2A2A"
							>
								{#each tienda.fotos as g (g)}<div
										style="aspect-ratio:1;overflow:hidden;background:#1E1E1E"
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
				<div style="height:clamp(40px,6cqw,80px)"></div>
			{/if}
		{/if}

		<!-- ============ CATÁLOGO ============ -->
		{#if pagina === 'catalogo' && !tienda.productos.length}
			<section style="padding:clamp(40px,7cqw,96px) clamp(16px,4cqw,48px)">
				<VacioTienda {tienda} {tema} motivo="sin-productos" />
			</section>
		{:else if pagina === 'catalogo'}
			<section
				style="padding:clamp(26px,4cqw,48px) clamp(14px,2.4cqw,28px) 0;border-bottom:2px solid #2A2A2A"
			>
				<h1
					style="margin:0;font-family:'Archivo Black',sans-serif;font-weight:400;font-size:min(clamp(56px,10cqw,150px),{cabe(
						tituloCat.join(' '),
						92
					)});line-height:.82;letter-spacing:-.045em;text-transform:uppercase;overflow-wrap:anywhere"
				>
					{tituloCat[0]}{#if tituloCat[1]}<br /><span style="color:var(--acento, #E6FF4D)"
							>{tituloCat[1]}</span
						>{/if}
				</h1>
				<div
					style="margin-top:22px;display:flex;overflow-x:auto;scrollbar-width:none;border-top:2px solid #2A2A2A;margin-left:calc(-1 * clamp(14px,2.4cqw,28px));margin-right:calc(-1 * clamp(14px,2.4cqw,28px))"
				>
					{#each tabs as t (t.c ?? '')}{@const on = t.c === categoria}<a
							href={enlaceCat(t.c)}
							style="flex:none;padding:14px 18px;border-right:2px solid #2A2A2A;background:{on
								? '#F2F2F2'
								: 'transparent'};color:{on
								? '#111'
								: '#F2F2F2'};text-decoration:none;font:800 12.5px 'Archivo';letter-spacing:.12em;text-transform:uppercase;white-space:nowrap"
							>{t.t} <span style="opacity:.55">{t.n}</span></a
						>{/each}
				</div>
			</section>
			<div style="display:flex;flex-wrap:wrap;align-items:stretch;border-bottom:2px solid #2A2A2A">
				<span
					style="display:flex;align-items:center;gap:8px;padding:12px 16px;border-right:2px solid #2A2A2A;font:800 12px 'Archivo';letter-spacing:.12em;text-transform:uppercase"
					>{listado.length} {listado.length === 1 ? 'producto' : 'productos'}</span
				>
				<span style="flex:1"></span>
				<label
					style="position:relative;display:flex;align-items:center;gap:6px;padding:12px 16px;border-left:2px solid #2A2A2A;font-size:13px;font-weight:700"
					>Ordenar:<select
						value={orden}
						onchange={(e) => goto(enlaceCat(categoria, e.currentTarget.value), { replace: true })}
						style="appearance:none;border:0;background:transparent;padding:0 18px 0 0;font:700 13px 'Archivo';color:#F2F2F2;cursor:pointer"
						>{#each ORDENES as [k, t] (k)}<option value={k} style="background:#111">{t}</option
							>{/each}</select
					><Icon
						icon={I.phCaretDownBold}
						size={12}
						style="position:absolute;right:16px;pointer-events:none"
					/></label
				>
			</div>
			{#if catalog.length}
				<div
					style="display:grid;grid-template-columns:repeat(auto-fill,minmax(min(50%,240px),1fr))"
				>
					{#each catalog as p (p.id)}{@render card(p, false, false)}{/each}
				</div>
				{#if listado.length > mostrar}
					<div style="display:flex;justify-content:center;padding:28px">
						<button
							type="button"
							onclick={() => (mostrar += 24)}
							style="padding:16px 26px;border:2px solid #F2F2F2;background:transparent;color:#F2F2F2;font:800 13px 'Archivo';letter-spacing:.12em;text-transform:uppercase;cursor:pointer"
							>Cargar 24 más</button
						>
					</div>
				{:else}
					<div style="height:28px"></div>
				{/if}
			{:else}
				<p style="margin:0;padding:40px clamp(14px,2.4cqw,28px);font-size:15px;color:#8A8A8A">
					Todavía no hay productos en esta categoría.
				</p>
			{/if}
		{/if}

		<!-- ============ PRODUCTO ============ -->
		{#if pagina === 'producto' && producto && sel}
			<section
				style="display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,420px),1fr));border-bottom:2px solid #2A2A2A"
			>
				<div
					style="background:#2A2A2A;border-right:2px solid #2A2A2A;display:flex;flex-direction:column;gap:2px"
				>
					<div style="aspect-ratio:4/5;background:#1E1E1E;overflow:hidden">
						{#if sel.foto}<img
								src={img(sel.foto, 1100)}
								alt={producto.nombre}
								style="width:100%;height:100%;object-fit:cover"
							/>{/if}
					</div>
					{#if sel.fotos.length > 1}<div
							style="display:grid;grid-template-columns:repeat(auto-fill,minmax(72px,1fr));gap:2px"
						>
							{#each sel.fotos as f, i (f)}<button
									type="button"
									aria-label="Ver la foto {i + 1}"
									aria-pressed={i === sel.fotoIdx}
									onclick={() => sel.verFoto(i)}
									style="position:relative;aspect-ratio:1;padding:0;border:0;background:#1E1E1E;overflow:hidden;cursor:pointer"
									><img
										src={img(f, 200, 200)}
										alt=""
										style="width:100%;height:100%;object-fit:cover;display:block;opacity:{i ===
										sel.fotoIdx
											? 1
											: 0.55}"
									/>{#if i === sel.fotoIdx}<span
											style="position:absolute;inset:0;box-shadow:inset 0 0 0 3px var(--acento, #E6FF4D)"
										></span>{/if}</button
								>{/each}
						</div>{/if}
				</div>
				<div
					style="padding:clamp(20px,3.4cqw,44px);position:sticky;top:58px;align-self:start;min-width:0"
				>
					<div style="display:flex;gap:6px;flex-wrap:wrap">
						{#if producto.categoria}<a
								href={enlaceCat(producto.categoria)}
								style="padding:5px 8px;background:var(--acento, #E6FF4D);color:#111;font:800 11px 'Archivo';letter-spacing:.1em;text-transform:uppercase;text-decoration:none"
								>{producto.categoria}</a
							>{/if}{#if tachado}<span
								style="padding:5px 8px;background:#FF4D2E;color:#111;font:800 11px 'Archivo';letter-spacing:.1em;text-transform:uppercase"
								>−{Math.round((1 - sel.precio / tachado) * 100)}%</span
							>{:else if producto.etiqueta}<span
								style="padding:5px 8px;background:#F2F2F2;color:#111;font:800 11px 'Archivo';letter-spacing:.1em;text-transform:uppercase"
								>{producto.etiqueta}</span
							>{/if}{#if sel.stock > 0 && sel.stock <= 3 && (!sel.conOpciones || sel.variante)}<span
								style="padding:5px 8px;border:1.5px solid #FF4D2E;color:#FF4D2E;font:800 11px 'Archivo';letter-spacing:.1em;text-transform:uppercase"
								>{sel.stock === 1 ? 'Queda 1' : `Quedan ${sel.stock}`}{#if sel.variante}&nbsp;en {sel.texto.toLowerCase()}{/if}</span
							>{/if}
					</div>
					<h1
						style="margin:14px 0 0;font-family:'Archivo Black',sans-serif;font-weight:400;font-size:min(clamp(40px,5.4cqw,72px),{cabe(
							producto.nombre,
							wide ? 40 : 86
						)});line-height:.86;letter-spacing:-.04em;text-transform:uppercase;overflow-wrap:anywhere"
					>
						{producto.nombre}
					</h1>
					<div style="margin-top:16px;display:flex;align-items:baseline;gap:12px;flex-wrap:wrap">
						<b style="font-family:'Archivo Black',sans-serif;font-size:32px;font-weight:400"
							>{sel.desde ? 'Desde ' : ''}{fmt(sel.precio)}</b
						>{#if tachado}<s style="font-size:17px;color:#8A8A8A">{fmt(tachado)}</s>{/if}
					</div>
					{#if tienda.cobros.cuotas > 1 || tienda.cobros.transferencia > 0}<div
							style="margin-top:6px;font-size:14px;color:#BDBDBD"
						>
							{#if tienda.cobros.cuotas > 1}{tienda.cobros.cuotas} × {fmt(
									sel.precio / tienda.cobros.cuotas
								)} sin interés{/if}{#if tienda.cobros.transferencia > 0}
								{#if tienda.cobros.cuotas > 1}·{/if}
								<b style="color:#B9F27C"
									>{fmt(sel.precio * (1 - tienda.cobros.transferencia / 100))} por&nbsp;transferencia</b
								>{/if}
						</div>{/if}
					{#each sel.opciones as o (o.nombre)}{@const swatch = o.valores.every(
							(x) => x.color || x.foto
						)}
						<div style="margin-top:20px">
							<div
								style="margin-bottom:10px;font:800 12px 'Archivo';letter-spacing:.12em;text-transform:uppercase;color:{sel.aviso &&
								!o.elegido
									? '#FF4D2E'
									: '#8A8A8A'}"
							>
								{o.nombre}{#if o.elegido}: <span style="color:#F2F2F2">{o.elegido}</span>{/if}
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
												: 0.3};box-shadow:inset 0 0 0 1px rgba(242,242,242,.25),0 0 0 {x.elegido
												? '3px'
												: '0px'} #111,0 0 0 {x.elegido ? '5px' : '0px'} var(--acento, #E6FF4D)"
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
											style="padding:13px 4px;text-align:center;border:2px solid {x.elegido
												? 'var(--acento, #E6FF4D)'
												: '#2A2A2A'};background:{x.elegido
												? 'var(--acento, #E6FF4D)'
												: 'transparent'};color:{x.elegido
												? '#111'
												: x.disponible
													? '#F2F2F2'
													: '#8A8A8A'};font:800 13px 'Archivo';letter-spacing:.08em;text-transform:uppercase;text-decoration:{x.disponible
												? 'none'
												: 'line-through'};cursor:{x.disponible ? 'pointer' : 'default'}"
											>{x.valor}</button
										>{/each}
								</div>{/if}
						</div>{/each}
					<button
						type="button"
						onclick={agregar}
						disabled={sel.stock === 0 || tienda.vitrina}
						style="margin-top:20px;width:100%;padding:19px;border:0;background:var(--acento, #E6FF4D);color:#111;font:800 14.5px 'Archivo';letter-spacing:.12em;text-transform:uppercase;cursor:pointer"
						>{sel.boton('Sumar a la bolsa', 'Agotado')}</button
					>
					{#if whatsapp}<a
							href={wa(
								`¡Hola! Quiero comprar ${producto.nombre}${sel.variante ? ` (${sel.texto})` : ''} (${fmt(sel.precio)}).`
							)}
							target="_blank"
							rel="noopener"
							style="margin-top:8px;display:flex;justify-content:center;align-items:center;gap:8px;padding:16px;border:2px solid #F2F2F2;color:#F2F2F2;text-decoration:none;font:800 13px 'Archivo';letter-spacing:.1em;text-transform:uppercase"
							><Icon icon={I.phWhatsappLogoBold} size={18} />Comprar por WhatsApp</a
						>{/if}
					<LanzamientoFicha {tienda} {producto} {sel} {tema} />
					<CompartirProducto {tienda} {producto} {tema} />
					<div style="margin-top:18px;border:2px solid #2A2A2A">
						{#if envia(tienda.envios)}<div style="display:flex">
								<label for="cp-ficha" class="oculto">Código postal</label><input
									id="cp-ficha"
									bind:value={cp}
									inputmode="numeric"
									maxlength="4"
									placeholder="Código postal"
									onkeydown={(e) => e.key === 'Enter' && calcular()}
									style="flex:1;min-width:0;padding:13px 14px;border:0;background:transparent;font:400 14px 'Archivo';color:#F2F2F2;outline:none"
								/><button
									type="button"
									onclick={calcular}
									style="padding:13px 16px;border:0;background:#F2F2F2;color:#111;font:800 12px 'Archivo';letter-spacing:.1em;text-transform:uppercase;cursor:pointer"
									>Calcular</button
								>
							</div>{/if}
						{#if cotizado}<div
								style="border-top:2px solid #2A2A2A;padding:12px 14px;display:flex;flex-direction:column;gap:8px;font-size:13.5px"
							>
								{#each cotizado as o (o.id)}<span
										style="display:flex;justify-content:space-between;gap:10px"
										><span>{o.nombre} · {o.detalle}</span><b
											style="white-space:nowrap;color:{o.precio ? '#F2F2F2' : '#B9F27C'}"
											>{o.precio ? fmt(o.precio) : 'Gratis'}</b
										></span
									>{/each}
							</div>{/if}
					</div>
					<div style="margin-top:18px;border:2px solid #2A2A2A">
						{#each acc as a, i (a.t)}<div
								style="padding:13px 14px;{i ? 'border-top:2px solid #2A2A2A' : ''}"
							>
								<button
									type="button"
									onclick={() => (abierto = a.open ? '' : a.t)}
									aria-expanded={a.open}
									style="width:100%;display:flex;justify-content:space-between;align-items:center;gap:10px;padding:0;border:0;background:transparent;font:800 12.5px 'Archivo';letter-spacing:.12em;text-transform:uppercase;color:#F2F2F2;cursor:pointer;text-align:left"
									>{a.t}<Icon icon={a.i} size={13} /></button
								>{#if a.open}<p
										style="margin:10px 0 0;font-size:14px;line-height:1.6;color:#BDBDBD"
									>
										{a.d}
									</p>{/if}
							</div>{/each}
					</div>
					<div style="margin-top:18px;padding:16px 14px;border:2px solid #2A2A2A">
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
				<section style="padding:clamp(36px,5cqw,64px) 0 0">
					<h2
						style="margin:0 clamp(14px,2.4cqw,28px) 18px;font-family:'Archivo Black',sans-serif;font-weight:400;font-size:clamp(34px,5cqw,64px);line-height:.85;letter-spacing:-.04em;text-transform:uppercase"
					>
						Va con
					</h2>
					<div
						style="display:grid;grid-template-columns:repeat(auto-fill,minmax(min(50%,240px),1fr));border-top:2px solid #2A2A2A"
					>
						{#each related as p (p.id)}{@render card(p, true, false)}{/each}
					</div>
				</section>
			{/if}
			<div style="height:clamp(36px,5cqw,64px)"></div>
		{/if}

		<!-- ============ CARRITO ============ -->
		{#if pagina === 'carrito'}
			<section
				style="display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,380px),1fr));border-bottom:2px solid #2A2A2A"
			>
				<div style="padding:clamp(22px,3.4cqw,44px);border-right:2px solid #2A2A2A;min-width:0">
					<h1
						style="margin:0;font-family:'Archivo Black',sans-serif;font-weight:400;font-size:clamp(48px,7cqw,100px);line-height:.85;letter-spacing:-.045em;text-transform:uppercase"
					>
						Bolsa ({carrito.unidades})
					</h1>
					{#if !lineas.length}
						<p style="margin:18px 0 0;font-size:15px;color:#8A8A8A">Todavía no sumaste nada.</p>
						<a
							href="{base}/catalogo"
							style="display:inline-block;margin-top:18px;padding:18px 28px;background:var(--acento, #E6FF4D);color:#111;font:800 14px 'Archivo';letter-spacing:.12em;text-transform:uppercase;text-decoration:none"
							>Ver todo →</a
						>
					{:else}
						{#if tienda.envios.gratisDesde}
							<div
								style="margin-top:16px;padding:12px 14px;border:2px solid #2A2A2A;font-size:13.5px"
							>
								<div style="display:flex;justify-content:space-between;gap:10px">
									<span
										>{#if falta}Te faltan <b>{fmt(falta)}</b> para envío gratis{:else}<b
												>Tenés envío gratis</b
											>{/if}</span
									><span style="color:#8A8A8A;white-space:nowrap"
										>{fmt(tienda.envios.gratisDesde)}</span
									>
								</div>
								<div style="margin-top:8px;height:6px;background:#2A2A2A">
									<div
										style="width:{Math.min(
											100,
											(subtotal / tienda.envios.gratisDesde) * 100
										)}%;height:100%;background:var(--acento, #E6FF4D)"
									></div>
								</div>
							</div>
						{/if}
						<div style="margin-top:14px;border-top:2px solid #2A2A2A">
							{#each lineas as c (c.clave)}<div
									style="display:flex;gap:14px;padding:14px 0;border-bottom:2px solid #2A2A2A"
								>
									{#if c.producto.foto}<img
											src={img(c.producto.foto, 200, 240)}
											alt=""
											style="width:92px;height:110px;object-fit:cover;flex:none;background:#1E1E1E"
										/>{/if}
									<div style="flex:1;min-width:0;display:flex;flex-direction:column">
										<div style="display:flex;justify-content:space-between;gap:10px">
											<a
												href="{base}/producto/{c.producto.slug ?? c.id}"
												style="font-size:14px;font-weight:800;text-transform:uppercase;color:#F2F2F2;text-decoration:none"
												>{c.producto.nombre}</a
											><b style="white-space:nowrap">{fmt(c.subtotal)}</b>
										</div>
										{#if c.variante}<span
												style="margin-top:4px;font-size:12px;font-weight:800;letter-spacing:.1em;text-transform:uppercase;color:#F2F2F2"
												>{c.variante}</span
											>{/if}
										<span style="font-size:13px;color:#8A8A8A;margin-top:4px"
											>{fmt(c.producto.precio)} c/u</span
										>
										<span style="flex:1;min-height:10px"></span>
										<div style="display:flex;justify-content:space-between;align-items:center">
											<span style="display:flex;border:2px solid #2A2A2A"
												><button
													type="button"
													aria-label="Uno menos"
													onclick={() => carrito.fijar(c.clave, c.cantidad - 1)}
													style="width:34px;height:32px;display:grid;place-items:center;border:0;background:transparent;color:#F2F2F2;cursor:pointer"
													><Icon icon={I.phMinusBold} size={12} /></button
												><b style="width:26px;display:grid;place-items:center;font-size:14px"
													>{c.cantidad}</b
												><button
													type="button"
													aria-label="Uno más"
													disabled={c.cantidad >= c.producto.stock}
													onclick={() => carrito.fijar(c.clave, c.cantidad + 1)}
													style="width:34px;height:32px;display:grid;place-items:center;border:0;background:transparent;color:#F2F2F2;cursor:pointer"
													><Icon icon={I.phPlusBold} size={12} /></button
												></span
											><button
												type="button"
												onclick={() => carrito.fijar(c.clave, 0)}
												style="padding:0;border:0;background:transparent;font:800 12px 'Archivo';letter-spacing:.1em;text-transform:uppercase;color:#8A8A8A;cursor:pointer"
												>Quitar</button
											>
										</div>
									</div>
								</div>{/each}
						</div>
					{/if}
				</div>
				{#if lineas.length}
					<div
						style="padding:clamp(22px,3.4cqw,44px);background:#161616;display:flex;flex-direction:column;gap:12px;font-size:14px;min-width:0"
					>
						<div
							style="font-family:'Archivo Black',sans-serif;font-size:28px;text-transform:uppercase"
						>
							Resumen
						</div>
						<span style="display:flex;justify-content:space-between"
							><span>Subtotal</span><span>{fmt(subtotal)}</span></span
						>
						<span style="display:flex;justify-content:space-between;gap:10px;color:#8A8A8A"
							><span>Envío</span><span>Se calcula en el próximo paso</span></span
						>
						<div style="height:2px;background:#2A2A2A"></div>
						<span style="display:flex;justify-content:space-between;align-items:baseline"
							><b style="text-transform:uppercase;letter-spacing:.08em">Total</b><b
								style="font-family:'Archivo Black',sans-serif;font-size:30px;font-weight:400"
								>{fmt(subtotal)}</b
							></span
						>
						{#if tienda.cobros.cuotas > 1}<span style="color:#BDBDBD"
								>{tienda.cobros.cuotas} × {fmt(subtotal / tienda.cobros.cuotas)} sin interés</span
							>{/if}
						{#if tienda.cobros.transferencia > 0}<span style="color:#B9F27C;font-weight:700"
								>{fmt(subtotal * (1 - tienda.cobros.transferencia / 100))} por transferencia</span
							>{/if}
						<a
							href="{base}/checkout"
							style="margin-top:6px;padding:19px;background:var(--acento, #E6FF4D);color:#111;font:800 14px 'Archivo';letter-spacing:.12em;text-transform:uppercase;text-align:center;text-decoration:none"
							>Pagar →</a
						>
						{#if upsell.length}
							<div
								style="margin-top:12px;font-size:12.5px;font-weight:800;letter-spacing:.12em;text-transform:uppercase;color:#8A8A8A"
							>
								Sumale
							</div>
							<div
								style="display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:2px;background:#2A2A2A;border:2px solid #2A2A2A"
							>
								{#each upsell as p (p.id)}<button
										type="button"
										aria-label="Sumar {p.nombre} a la bolsa"
										onclick={() => sumarRapido(p)}
										style="display:block;border:0;background:#111;padding:8px;text-align:left;color:#F2F2F2;cursor:pointer;font-family:inherit;min-width:0"
										><span
											style="position:relative;display:block;aspect-ratio:1;overflow:hidden;background:#1E1E1E"
											>{#if p.foto}<img
													src={img(p.foto, 300, 300)}
													alt=""
													style="width:100%;height:100%;object-fit:cover"
												/>{/if}<span
												style="position:absolute;right:6px;bottom:6px;width:28px;height:28px;display:grid;place-items:center;background:var(--acento, #E6FF4D);color:#111"
												><Icon icon={I.phPlusBold} size={13} /></span
											></span
										><span
											style="margin-top:8px;display:flex;justify-content:space-between;gap:6px;font-size:12.5px;font-weight:700;text-transform:uppercase"
											><span style="min-width:0">{p.nombre}</span><span style="white-space:nowrap"
												>{tarjetaDe(p).desde ? 'Desde ' : ''}{fmt(tarjetaDe(p).precio)}</span
											></span
										></button
									>{/each}
							</div>
						{/if}
					</div>
				{/if}
			</section>
		{/if}

		<!-- ============ CHECKOUT ============ -->
		{#if pagina === 'checkout'}
			{#if !lineas.length}
				<section style="padding:clamp(22px,3.4cqw,44px);border-bottom:2px solid #2A2A2A">
					<p style="margin:0;font-size:15px;color:#8A8A8A">Tu bolsa está vacía.</p>
					<a
						href="{base}/catalogo"
						style="display:inline-block;margin-top:16px;color:var(--acento, #E6FF4D);font-weight:700"
						>Ver todo</a
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
					style="display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,380px),1fr));border-bottom:2px solid #2A2A2A"
				>
					<input type="hidden" name="carrito" value={JSON.stringify(carrito.lineas)} />
					<input type="hidden" name="envio" value={envio ?? ''} />
					<input type="hidden" name="pago" value={pago} />
					<div
						style="padding:clamp(22px,3.4cqw,44px);border-right:2px solid #2A2A2A;display:flex;flex-direction:column;gap:24px;min-width:0"
					>
						<div style="display:flex;border:2px solid #2A2A2A">
							{#each pasos as s, i (s.t)}<span
									style="flex:1;padding:11px 8px;text-align:center;{i < pasos.length - 1
										? 'border-right:2px solid #2A2A2A;'
										: ''}background:{s.on ? '#F2F2F2' : 'transparent'};color:{s.on
										? '#111'
										: '#8A8A8A'};font:800 11.5px 'Archivo';letter-spacing:.12em;text-transform:uppercase"
									>{s.t}</span
								>{/each}
						</div>
						<div>
							<div
								style="font-family:'Archivo Black',sans-serif;font-size:28px;text-transform:uppercase"
							>
								Tus datos
							</div>
							<div
								style="margin-top:12px;display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));gap:2px;background:#2A2A2A;border:2px solid #2A2A2A"
							>
								{#each [['email', 'Email', 'email', 'email'], ['nombre', 'Nombre y apellido', 'text', 'name'], ['telefono', 'Teléfono', 'tel', 'tel'], ['dni', 'DNI', 'text', 'off'], ...camposFactura(tienda)] as [n, l, tipo, ac] (n)}
									<label style="display:block;padding:11px 13px;background:#111"
										><span
											style="display:block;font-size:11px;font-weight:800;letter-spacing:.12em;text-transform:uppercase;color:#8A8A8A"
											>{l}</span
										><input
											name={n}
											type={tipo}
											autocomplete={ac as 'email'}
											required={!OPCIONALES.has(n)}
											maxlength="120"
											style="width:100%;margin-top:3px;padding:0;border:0;background:transparent;font:400 15px 'Archivo';color:#F2F2F2;outline:none"
										/></label
									>
								{/each}
							</div>
							<!-- Sin bloque de newsletter en el diseño de esta plantilla: la casilla del checkout (§50). -->
							<label
								style="margin-top:10px;display:flex;align-items:center;gap:10px;min-height:44px;font:400 15px 'Archivo';color:#F2F2F2;cursor:pointer"
								><input
									type="checkbox"
									name="novedades"
									style="flex:none;width:18px;height:18px;margin:0;accent-color:var(--acento, #E6FF4D)"
								/>Quiero recibir novedades</label
							>
						</div>
						<div>
							<div
								style="font-family:'Archivo Black',sans-serif;font-size:28px;text-transform:uppercase"
							>
								Entrega
							</div>
							<div
								style="margin-top:12px;display:grid;grid-template-columns:minmax(0,130px) minmax(0,1fr);gap:2px;background:#2A2A2A;border:2px solid #2A2A2A"
							>
								<label style="display:block;padding:11px 13px;background:#111"
									><span
										style="display:block;font-size:11px;font-weight:800;letter-spacing:.12em;text-transform:uppercase;color:#8A8A8A"
										>Código postal</span
									><input
										name="cp"
										bind:value={cp}
										inputmode="numeric"
										maxlength="4"
										style="width:100%;margin-top:3px;padding:0;border:0;background:transparent;font:400 15px 'Archivo';color:#F2F2F2;outline:none"
									/></label
								>
								<label style="display:block;padding:11px 13px;background:#111"
									><span
										style="display:block;font-size:11px;font-weight:800;letter-spacing:.12em;text-transform:uppercase;color:#8A8A8A"
										>Dirección (calle, número, piso)</span
									><input
										name="direccion"
										maxlength="160"
										autocomplete="street-address"
										required={envio !== 'retiro'}
										style="width:100%;margin-top:3px;padding:0;border:0;background:transparent;font:400 15px 'Archivo';color:#F2F2F2;outline:none"
									/></label
								>
							</div>
							<div
								style="margin-top:8px;display:flex;flex-direction:column;border:2px solid #2A2A2A;border-bottom:0"
								role="radiogroup"
								aria-label="Forma de entrega"
							>
								{#each opciones as s (s.id)}{@const on = envio === s.id}<button
										type="button"
										role="radio"
										aria-checked={on}
										onclick={() => (envio = s.id)}
										style="display:flex;align-items:center;gap:12px;padding:15px;border:0;border-bottom:2px solid #2A2A2A;background:{on
											? 'var(--acento, #E6FF4D)'
											: '#111'};color:{on
											? '#111'
											: '#F2F2F2'};text-align:left;cursor:pointer;font-family:inherit"
										><span
											style="width:18px;height:18px;border:2px solid currentColor;display:grid;place-items:center;flex:none"
											>{#if on}<span style="width:8px;height:8px;background:currentColor"
												></span>{/if}</span
										><span style="flex:1;min-width:0"
											><span style="display:block;font-size:14.5px;font-weight:700">{s.nombre}</span
											><span style="display:block;font-size:12.5px;opacity:.7">{s.detalle}</span
											></span
										><b style="white-space:nowrap">{s.precio ? fmt(s.precio) : 'Gratis'}</b></button
									>{:else}<span
										style="padding:15px;border-bottom:2px solid #2A2A2A;font-size:13.5px;color:#8A8A8A"
										>Poné tu código postal para ver cómo te llega.</span
									>{/each}
							</div>
							<EnvioCorreo {tienda} {cp} {envio} {opciones} />
						</div>
						<div>
							<div
								style="font-family:'Archivo Black',sans-serif;font-size:28px;text-transform:uppercase"
							>
								Pago
							</div>
							<div
								style="margin-top:12px;display:flex;flex-direction:column;border:2px solid #2A2A2A;border-bottom:0"
								role="radiogroup"
								aria-label="Forma de pago"
							>
								{#each pagos as s (s.id)}{@const on = pago === s.id}<button
										type="button"
										role="radio"
										aria-checked={on}
										onclick={() => (pago = s.id)}
										style="display:flex;align-items:center;gap:12px;padding:15px;border:0;border-bottom:2px solid #2A2A2A;background:{on
											? 'var(--acento, #E6FF4D)'
											: '#111'};color:{on
											? '#111'
											: '#F2F2F2'};text-align:left;cursor:pointer;font-family:inherit"
										><span
											style="width:18px;height:18px;border:2px solid currentColor;display:grid;place-items:center;flex:none"
											>{#if on}<span style="width:8px;height:8px;background:currentColor"
												></span>{/if}</span
										><Icon icon={s.i} size={22} style="flex:none" /><span style="flex:1;min-width:0"
											><span style="display:block;font-size:14.5px;font-weight:700">{s.t}</span
											><span style="display:block;font-size:12.5px;opacity:.7">{s.d}</span></span
										></button
									>{/each}
							</div>
						</div>
						{#if error}<div
								role="alert"
								style="padding:12px 14px;border:2px solid #FF4D2E;color:#FF4D2E;font-size:14px;font-weight:600"
							>
								{error}
							</div>{/if}
						<button
							disabled={!envio || !pagos.length || enviando || tienda.vitrina}
							style="padding:20px;border:0;background:var(--acento, #E6FF4D);color:#111;font:800 15px 'Archivo';letter-spacing:.12em;text-transform:uppercase;cursor:pointer;display:flex;justify-content:center;align-items:center;gap:8px"
							><Icon icon={I.phLockSimpleFill} size={16} />{tienda.vitrina
								? 'Es una tienda de ejemplo'
								: !pagos.length
									? 'Esta tienda todavía no tiene medios de pago'
									: envio
										? `Pagar ${fmt(total)}`
										: 'Elegí cómo te llega'}</button
						>
					</div>
					<div style="padding:clamp(22px,3.4cqw,44px);background:#161616;min-width:0">
						<div
							style="font-family:'Archivo Black',sans-serif;font-size:26px;text-transform:uppercase"
						>
							Tu pedido
						</div>
						<div style="margin-top:14px;display:flex;flex-direction:column;gap:12px">
							{#each lineas as c (c.clave)}<div style="display:flex;gap:12px;align-items:center">
									{#if c.producto.foto}<img
											src={img(c.producto.foto, 120, 140)}
											alt=""
											style="width:56px;height:66px;object-fit:cover;flex:none;background:#1E1E1E"
										/>{/if}<span
										style="flex:1;min-width:0;font-size:13.5px;font-weight:700;text-transform:uppercase"
										>{c.producto.nombre}{#if c.variante}<span
												style="display:block;margin-top:2px;font-size:11.5px;font-weight:800;letter-spacing:.1em;color:#BDBDBD"
												>{c.variante}</span
											>{:else}<br />{/if}<span
											style="font-weight:500;text-transform:none;color:#8A8A8A"
											>{c.cantidad} × {fmt(c.producto.precio)}</span
										></span
									><b style="font-size:14px;white-space:nowrap">{fmt(c.subtotal)}</b>
								</div>{/each}
						</div>
						<div
							style="margin-top:16px;padding-top:14px;border-top:2px solid #2A2A2A;display:flex;flex-direction:column;gap:8px;font-size:14px"
						>
							<span style="display:flex;justify-content:space-between"
								><span>Subtotal</span><span>{fmt(subtotal)}</span></span
							><span style="display:flex;justify-content:space-between"
								><span>{elegida ? elegida.nombre : 'Envío'}</span><span
									>{elegida ? (cuenta.envio ? fmt(cuenta.envio) : 'Gratis') : '—'}</span
								></span
							>{#if cuenta.descCupon}<span
									style="display:flex;justify-content:space-between;color:#B9F27C;font-weight:700"
									><span>Cupón {cupon.aplicado}</span><span>− {fmt(cuenta.descCupon)}</span></span
								>{/if}{#if descuento}<span
									style="display:flex;justify-content:space-between;color:#B9F27C;font-weight:700"
									><span>Transferencia −{tienda.cobros.transferencia}%</span><span
										>− {fmt(descuento)}</span
									></span
								>{/if}
							<div style="margin:6px 0 2px">
								<CampoCupon
									{cupon}
									lineas={carrito.lineas}
									tinta="#F2F2F2"
									suave="#8A8A8A"
									borde="#3A3A3A"
									fondo="#0B0B0B"
									acento="var(--acento, #E6FF4D)"
									sobreAcento="#111"
									radio="0px"
									fuente="'Archivo',sans-serif"
									error="#FF8A75"
								/>
							</div>
							<span
								style="display:flex;justify-content:space-between;align-items:baseline;margin-top:6px"
								><b style="text-transform:uppercase">Total</b><b
									style="font-family:'Archivo Black',sans-serif;font-size:28px;font-weight:400"
									>{fmt(total)}</b
								></span
							>
						</div>
						<div
							style="margin-top:16px;display:flex;align-items:center;gap:8px;font-size:12.5px;color:#8A8A8A"
						>
							<Icon icon={I.phShieldCheckBold} size={16} />Compra protegida · 10 días para
							arrepentirte
						</div>
					</div>
				</form>
			{/if}
		{/if}

		<!-- ============ LISTO (no está en el diseño: con las piezas del checkout) ============ -->
		{#if pagina === 'listo' && pedido}
			<section
				style="display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,380px),1fr));border-bottom:2px solid #2A2A2A"
			>
				<div
					style="padding:clamp(22px,3.4cqw,44px);border-right:2px solid #2A2A2A;display:flex;flex-direction:column;gap:20px;min-width:0"
				>
					<div
						style="display:flex;align-items:center;gap:10px;font-size:12.5px;font-weight:800;letter-spacing:.14em;text-transform:uppercase;color:var(--acento, #E6FF4D)"
					>
						<Icon icon={I.phCheckBold} size={16} />Pedido #{pedido.numero}
					</div>
					<h1
						style="margin:0;font-family:'Archivo Black',sans-serif;font-weight:400;font-size:min(clamp(48px,7cqw,100px),{cabe(
							`Gracias, ${pedido.nombre}!`,
							wide ? 42 : 86
						)});line-height:.85;letter-spacing:-.045em;text-transform:uppercase;overflow-wrap:anywhere"
					>
						¡Gracias,<br /><span style="color:var(--acento, #E6FF4D)">{pedido.nombre}!</span>
					</h1>
					<p style="margin:0;font-size:17px;line-height:1.5;color:#BDBDBD;max-width:460px">
						Recibimos tu pedido por <b style="color:#F2F2F2">{fmt(pedido.total)}</b>. Te escribimos
						por mail con cada&nbsp;novedad.
					</p>
					<div>
						<a
							href={base || '/'}
							style="display:inline-block;padding:18px 28px;background:var(--acento, #E6FF4D);color:#111;font:800 14px 'Archivo';letter-spacing:.12em;text-transform:uppercase;text-decoration:none"
							>Volver a la tienda</a
						>
					</div>
				</div>
				<SenaPedido {tema} />
				{#if pedido.pago === 'transferencia'}
					<div
						style="padding:clamp(22px,3.4cqw,44px);background:#161616;font-size:15px;line-height:1.6;min-width:0"
					>
						<div
							style="font-family:'Archivo Black',sans-serif;font-size:26px;text-transform:uppercase"
						>
							Transferí {fmt(pedido.total)}
						</div>
						<p style="margin:12px 0 0;color:#BDBDBD">
							{#if pedido.alias}Al {pedido.esCbu ? 'CBU/CVU' : 'alias'}
								<b style="color:var(--acento, #E6FF4D)">{pedido.alias}</b>{:else}Con los datos que
								te mandamos por mail{/if} y mandanos el comprobante{#if whatsapp}&nbsp;<a
									href={wa(`¡Hola! Te mando el comprobante del pedido #${pedido.numero}.`)}
									target="_blank"
									rel="noopener"
									style="color:var(--acento, #E6FF4D)">por WhatsApp</a
								>{:else}&nbsp;respondiendo el mail del pedido{/if}. Lo preparamos apenas se
							acredite.
						</p>
					</div>
				{/if}
			</section>
			{#if pedido.estado === 'Entregado'}
				<section
					style="padding:clamp(28px,4cqw,56px) clamp(14px,2.4cqw,28px);border-bottom:2px solid #2A2A2A"
				>
					<FormResena slug={tienda.slug} {tema} codigo={pedido.codigo} nombre={pedido.nombre} />
				</section>
			{/if}
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
				style="display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,380px),1fr));border-bottom:2px solid #2A2A2A"
			>
				<input type="hidden" name="servicioId" value={reserva.servicioId ?? ''} />
				<input type="hidden" name="profesionalId" value={reserva.profesionalId ?? ''} />
				<input type="hidden" name="inicio" value={reserva.inicio ?? ''} />
				<div
					style="padding:clamp(22px,3.4cqw,44px);border-right:2px solid #2A2A2A;display:flex;flex-direction:column;gap:24px;min-width:0"
				>
					<div style="display:flex;border:2px solid #2A2A2A">
						{#each PASOS_TURNO as t, i (t)}<span
								style="flex:1;min-width:0;padding:11px 4px;text-align:center;{i <
								PASOS_TURNO.length - 1
									? 'border-right:2px solid #2A2A2A;'
									: ''}background:{i <= reserva.paso ? '#F2F2F2' : 'transparent'};color:{i <=
								reserva.paso
									? '#111'
									: '#8A8A8A'};font:800 10.5px 'Archivo';letter-spacing:.08em;text-transform:uppercase;overflow:hidden;text-overflow:ellipsis"
								>{t}</span
							>{/each}
					</div>
					<h1
						style="margin:0;font-family:'Archivo Black',sans-serif;font-weight:400;font-size:min(clamp(48px,7cqw,100px),{cabe(
							'Reservá tu turno',
							wide ? 42 : 86
						)});line-height:.85;letter-spacing:-.045em;text-transform:uppercase;overflow-wrap:anywhere"
					>
						Reservá<br /><span style="color:var(--acento, #E6FF4D)">tu turno</span>
					</h1>
					<div>
						<div
							style="font-family:'Archivo Black',sans-serif;font-size:28px;text-transform:uppercase"
						>
							1 · Servicio
						</div>
						<div
							style="margin-top:12px;display:flex;flex-direction:column;border:2px solid #2A2A2A;border-bottom:0"
							role="radiogroup"
							aria-label="Servicio"
						>
							{#each tienda.turnos.servicios as s (s.id)}{@const on =
									reserva.servicioId === s.id}<button
									type="button"
									role="radio"
									aria-checked={on}
									onclick={() => reserva.elegirServicio(s.id)}
									style="display:flex;align-items:center;gap:12px;width:100%;padding:15px;border:0;border-bottom:2px solid #2A2A2A;background:{on
										? 'var(--acento, #E6FF4D)'
										: '#111'};color:{on
										? '#111'
										: '#F2F2F2'};text-align:left;cursor:pointer;font-family:inherit"
									><span
										style="width:18px;height:18px;border:2px solid currentColor;display:grid;place-items:center;flex:none"
										>{#if on}<span style="width:8px;height:8px;background:currentColor"
											></span>{/if}</span
									><span style="flex:1;min-width:0"
										><span style="display:block;font-size:14.5px;font-weight:700">{s.nombre}</span
										><span style="display:block;font-size:12.5px;opacity:.7;"
											>{duracion(s.duracion)}</span
										></span
									><b style="white-space:nowrap">{fmt(s.precio)}</b></button
								>{/each}
						</div>
					</div>
					<div>
						<div
							style="font-family:'Archivo Black',sans-serif;font-size:28px;text-transform:uppercase"
						>
							2 · Profesional
						</div>
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
									style="flex:none;width:110px;padding:10px;border:2px solid {on
										? 'var(--acento, #E6FF4D)'
										: '#2A2A2A'};background:{on ? 'var(--acento, #E6FF4D)' : '#111'};color:{on
										? '#111'
										: '#F2F2F2'};text-align:center;cursor:pointer;font-family:inherit"
									>{#if p?.foto}<img
											src={img(p.foto, 120, 120)}
											alt=""
											style="width:56px;height:56px;object-fit:cover;display:block;margin:0 auto"
										/>{:else}<span
											style="width:56px;height:56px;background:#1A1A1A;color:var(--acento, #E6FF4D);display:grid;place-items:center;margin:0 auto"
											><Icon icon={p ? I.phUserFill : I.phShuffleBold} size={22} /></span
										>{/if}<span
										style="display:block;margin-top:8px;font-size:13.5px;font-weight:700;overflow:hidden;text-overflow:ellipsis;white-space:nowrap"
										>{p?.nombre ?? 'Cualquiera'}</span
									><span
										style="display:block;font-size:11.5px;opacity:.7;overflow:hidden;text-overflow:ellipsis;white-space:nowrap"
										>{p ? p.rol.split(' · ')[0] : 'El primero libre'}</span
									></button
								>{/each}
						</div>
					</div>
					<div>
						<div style="display:flex;justify-content:space-between;align-items:center;gap:10px">
							<div
								style="font-family:'Archivo Black',sans-serif;font-size:28px;text-transform:uppercase"
							>
								3 · Día y horario
							</div>
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
									style="padding:9px 0;text-align:center;cursor:pointer;font-family:inherit;min-width:0;border:2px solid {on
										? 'var(--acento, #E6FF4D)'
										: '#2A2A2A'};background:{on ? 'var(--acento, #E6FF4D)' : '#111'};color:{on
										? '#111'
										: '#F2F2F2'}"
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
								<p style="margin:0;font-size:14px;opacity:.7;">
									Elegí un servicio para ver los horarios.
								</p>
							{:else if reserva.horarios && !reserva.grupos.length}
								<p style="margin:0;font-size:14px;opacity:.7;">
									No quedan horarios este día. Probá con otro día u otro&nbsp;profesional.
								</p>
							{/if}
							{#each reserva.grupos as g (g.k)}<div>
									<div
										style="font-size:12px;font-weight:700;letter-spacing:.08em;text-transform:uppercase;margin-bottom:7px;opacity:.7;"
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
													: 'line-through'};border:2px solid {on
													? 'var(--acento, #E6FF4D)'
													: '#2A2A2A'};background:{on ? 'var(--acento, #E6FF4D)' : '#111'};color:{on
													? '#111'
													: '#F2F2F2'}">{h.hora}</button
											>{/each}
									</div>
								</div>{/each}
						</div>
					</div>
					{#if reserva.listo}
						<div>
							<div
								style="font-family:'Archivo Black',sans-serif;font-size:28px;text-transform:uppercase"
							>
								4 · Tus datos
							</div>
							<div
								style="margin-top:12px;display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));gap:2px;background:#2A2A2A;border:2px solid #2A2A2A"
							>
								{#each [['nombre', 'Nombre', 'text', 'name'], ['telefono', 'WhatsApp', 'tel', 'tel'], ['email', 'Email', 'email', 'email']] as [n, l, tipo, ac] (n)}
									<label style="display:block;padding:11px 13px;background:#111"
										><span
											style="display:block;font-size:11px;font-weight:800;letter-spacing:.12em;text-transform:uppercase;color:#8A8A8A"
											>{l}</span
										><input
											name={n}
											type={tipo}
											autocomplete={ac as 'email'}
											required
											maxlength="120"
											style="width:100%;margin-top:3px;padding:0;border:0;background:transparent;font:400 15px 'Archivo';color:#F2F2F2;outline:none"
										/></label
									>
								{/each}
							</div>
						</div>
						<div>
							<div
								style="font-family:'Archivo Black',sans-serif;font-size:28px;text-transform:uppercase"
							>
								Pago
							</div>
							<div
								style="margin-top:12px;display:flex;flex-direction:column;border:2px solid #2A2A2A;border-bottom:0"
							>
								<div
									style="display:flex;align-items:center;gap:12px;width:100%;padding:15px;border:0;border-bottom:2px solid #2A2A2A;background:var(--acento, #E6FF4D);color:#111;text-align:left;cursor:pointer;font-family:inherit"
								>
									<span
										style="width:18px;height:18px;border:2px solid currentColor;display:grid;place-items:center;flex:none"
										><span style="width:8px;height:8px;background:currentColor"></span></span
									>{#if reserva.sena}<Icon
											icon={I.siMercadopago}
											size={22}
											style="color:#00B1EA;flex:none"
										/><span style="flex:1;min-width:0"
											><span style="display:block;font-size:14.5px;font-weight:700"
												>Mercado Pago</span
											><span style="display:block;font-size:12.5px;opacity:.7;"
												>Seña de {fmt(reserva.sena)} para confirmar · el resto, en el local</span
											></span
										>{:else}<Icon icon={I.phStorefrontFill} size={22} style="flex:none" /><span
											style="flex:1;min-width:0"
											><span style="display:block;font-size:14.5px;font-weight:700"
												>Pagás en el local</span
											><span style="display:block;font-size:12.5px;opacity:.7;"
												>No hace falta seña: el turno queda confirmado</span
											></span
										>{/if}
								</div>
							</div>
						</div>
					{/if}
				</div>
				<div style="padding:clamp(22px,3.4cqw,44px);background:#161616;min-width:0">
					<div
						style="font-family:'Archivo Black',sans-serif;font-size:26px;text-transform:uppercase"
					>
						Tu reserva
					</div>
					<div style="margin-top:14px;display:flex;align-items:center;gap:12px">
						{#if reserva.profesional?.foto}<img
								src={img(reserva.profesional.foto, 120, 120)}
								alt=""
								style="width:52px;height:52px;object-fit:cover;flex:none"
							/>{/if}
						<div style="min-width:0">
							<div style="font-size:15px;font-weight:700">
								{reserva.conQuien ? `Con ${reserva.conQuien}` : 'Con el primero libre'}
							</div>
							<div style="font-size:13px;color:#8A8A8A;">
								{reserva.cuando ?? 'Elegí el día y el horario'}
							</div>
						</div>
					</div>
					{#if reserva.servicio}
						<div
							style="margin-top:16px;padding-top:14px;border-top:2px solid #2A2A2A;display:flex;flex-direction:column;gap:8px;font-size:14px"
						>
							<span style="display:flex;justify-content:space-between;gap:10px"
								><span>{reserva.servicio.nombre}</span><span style="white-space:nowrap"
									>{fmt(reserva.servicio.precio)}</span
								></span
							><span style="display:flex;justify-content:space-between;color:#8A8A8A;"
								><span>Duración total</span><span>{reserva.duracion}</span></span
							><span
								style="display:flex;justify-content:space-between;align-items:baseline;margin-top:6px"
								><b>Total</b><b
									style="font-family:'Archivo Black',sans-serif;font-size:28px;font-weight:400"
									>{fmt(reserva.servicio.precio)}</b
								></span
							>
						</div>
					{/if}
					<p style="margin:14px 0 0;font-size:13px;line-height:1.5;color:#8A8A8A;">
						{#if reserva.sena}Seña de {fmt(reserva.sena)} para confirmar. Podés cancelar desde el link
							del turno hasta {anticipacion} antes; la seña te la devuelve la&nbsp;tienda.{:else}No
							pagás nada ahora. Podés cancelar desde el link del turno hasta {anticipacion}&nbsp;antes.{/if}
					</p>
					{#if error}<div
							role="alert"
							style="margin-top:12px;padding:12px 14px;border:2px solid #FF4D2E;color:#FF4D2E;font-size:14px;font-weight:600"
						>
							{error}
						</div>{/if}
					<button
						disabled={!reserva.listo || reservando || tienda.vitrina}
						style="margin-top:16px;width:100%;padding:20px;border:0;background:var(--acento, #E6FF4D);color:#111;font:800 15px 'Archivo';letter-spacing:.12em;text-transform:uppercase;cursor:pointer"
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
			<section style="padding:clamp(22px,3.4cqw,44px);border-bottom:2px solid #2A2A2A">
				<div
					style="display:flex;align-items:center;gap:10px;font-size:12.5px;font-weight:800;letter-spacing:.14em;text-transform:uppercase;color:var(--acento, #E6FF4D)"
				>
					<Icon icon={cancelado ? I.phCalendarX : I.phCheckBold} size={16} />Turno
				</div>
				<h1
					style="margin:20px 0 0;font-family:'Archivo Black',sans-serif;font-weight:400;font-size:clamp(40px,6cqw,84px);line-height:.9;letter-spacing:-.045em;text-transform:uppercase;overflow-wrap:anywhere"
				>
					{cancelado
						? 'Turno cancelado'
						: turno.estado === 'reservado'
							? `Falta la seña, ${turno.nombre}`
							: `¡Listo, ${turno.nombre}!`}
				</h1>
				<p style="margin:20px 0 0;font-size:17px;line-height:1.5;color:#BDBDBD;max-width:460px">
					{#if cancelado}Liberamos el horario{turno.senaPagada
							? '. La seña te la devuelve la tienda por Mercado Pago'
							: ''}. Si querés, reservá&nbsp;otro.{:else if turno.estado === 'reservado'}Te
						guardamos el horario unos minutos mientras pagás la seña con Mercado&nbsp;Pago.{:else}Tu
						turno está confirmado. Te esperamos en {turno.tienda}.{/if}
				</p>
				{#if aviso === 'pendiente' || aviso === 'error'}<div
						role="status"
						style="margin-top:14px;padding:12px 14px;border:2px solid #FF4D2E;color:#FF4D2E;font-size:14px;font-weight:600"
					>
						{aviso === 'pendiente'
							? 'Mercado Pago todavía no acreditó la seña: te avisamos cuando entre.'
							: 'El pago de la seña no salió. Podés reservar de nuevo.'}
					</div>{/if}
				<div
					style="margin-top:22px;padding:clamp(18px,3cqw,28px);background:#161616;border:2px solid #2A2A2A;max-width:620px"
				>
					<div
						style="font-size:16px;font-weight:700;text-decoration:{cancelado
							? 'line-through'
							: 'none'}"
					>
						{cuandoLargo(turno.inicio)} con {turno.profesional}
					</div>
					<div style="margin-top:4px;font-size:14px;color:#8A8A8A;">
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
						style="margin-top:12px;padding:12px 14px;border:2px solid #FF4D2E;color:#FF4D2E;font-size:14px;font-weight:600"
					>
						{error}
					</div>{/if}
				<div style="margin-top:22px;display:flex;gap:10px;flex-wrap:wrap;">
					{#if !cancelado}<a
							href="{base}/turno/{turno.codigo}/turno.ics"
							download="turno.ics"
							data-sveltekit-reload
							style="display:flex;align-items:center;gap:8px;padding:18px 28px;border:0;background:var(--acento, #E6FF4D);color:#111;font:800 14px 'Archivo';letter-spacing:.12em;text-transform:uppercase;text-decoration:none;cursor:pointer"
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
									style="display:inline-block;padding:16px 26px;border:2px solid #2A2A2A;background:transparent;color:#F2F2F2;font:800 14px 'Archivo';letter-spacing:.12em;text-transform:uppercase;text-decoration:none;cursor:pointer"
									>Sí, cancelar el turno</button
								>{:else}<button
									type="button"
									onclick={() => (seguro = true)}
									style="display:inline-block;padding:16px 26px;border:2px solid #2A2A2A;background:transparent;color:#F2F2F2;font:800 14px 'Archivo';letter-spacing:.12em;text-transform:uppercase;text-decoration:none;cursor:pointer"
									>Cancelar turno</button
								>{/if}
						</form>
					{:else if cancelado}<a
							href="{base}/turnos"
							style="padding:18px 28px;border:0;background:var(--acento, #E6FF4D);color:#111;font:800 14px 'Archivo';letter-spacing:.12em;text-transform:uppercase;text-decoration:none;cursor:pointer"
							>Reservar otro</a
						>{/if}
					<a
						href={base || '/'}
						style="display:inline-block;padding:16px 26px;border:2px solid #2A2A2A;background:transparent;color:#F2F2F2;font:800 14px 'Archivo';letter-spacing:.12em;text-transform:uppercase;text-decoration:none;cursor:pointer"
						>Volver a la tienda</a
					>
				</div>
				{#if !cancelado && !turno.cancelable}<p
						style="margin:14px 0 0;font-size:13.5px;opacity:.7;"
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
				style="position:relative;min-height:{tienda.portada.foto
					? 'clamp(420px,56cqw,720px)'
					: 'clamp(220px,26cqw,340px)'};border-bottom:2px solid #2A2A2A;overflow:hidden;background:#1E1E1E"
			>
				{#if tienda.portada.foto}<img
						src={img(tienda.portada.foto, 1800)}
						alt=""
						style="position:absolute;inset:0;width:100%;height:100%;object-fit:cover;filter:grayscale(1) contrast(1.1);opacity:.6"
					/>{/if}
				<div
					style="position:absolute;left:clamp(14px,2.4cqw,28px);right:clamp(14px,2.4cqw,28px);bottom:clamp(20px,3cqw,40px)"
				>
					{#if tienda.ciudad}<div
							style="font-size:12.5px;font-weight:800;letter-spacing:.14em;text-transform:uppercase;color:var(--acento, #E6FF4D)"
						>
							{tienda.ciudad}
						</div>{/if}
					<h1
						style="margin:10px 0 0;font-family:'Archivo Black',sans-serif;font-weight:400;font-size:min(clamp(50px,9cqw,136px),{cabe(
							tienda.portada.titulo,
							92
						)});line-height:.84;letter-spacing:-.045em;text-transform:uppercase;overflow-wrap:anywhere"
					>
						{titulo[0]}{#if titulo[1]}<br />{titulo[1]}{/if}
					</h1>
				</div>
			</section>
			<section
				style="display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,320px),1fr));border-bottom:2px solid #2A2A2A"
			>
				<div style="padding:clamp(22px,3.4cqw,44px);border-right:2px solid #2A2A2A">
					<p style="margin:0;font-size:clamp(19px,2.2cqw,26px);line-height:1.4;font-weight:500">
						{tienda.nombre}{tienda.ciudad ? `, desde ${tienda.ciudad}` : ''}. {comoLlega(
							tienda.envios
						)}
					</p>
				</div>
				{#if cifras.length}<div
						style="display:grid;grid-template-columns:repeat({cifras.length},minmax(0,1fr))"
					>
						{#each cifras as s (s.t)}<div
								style="padding:clamp(18px,2.6cqw,32px);border-right:2px solid #2A2A2A;display:flex;flex-direction:column;justify-content:flex-end;min-width:0"
							>
								<div
									style="font-family:'Archivo Black',sans-serif;font-size:clamp(34px,4.4cqw,58px);line-height:.9;color:var(--acento, #E6FF4D)"
								>
									{s.v}
								</div>
								<div
									style="margin-top:6px;font-size:12px;font-weight:800;letter-spacing:.1em;text-transform:uppercase;color:#8A8A8A"
								>
									{s.t}
								</div>
							</div>{/each}
					</div>{/if}
			</section>
			{#if tienda.direccion || whatsapp}
				<section
					style="display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,320px),1fr));border-bottom:2px solid #2A2A2A"
				>
					{#if tienda.direccion}
						<div
							style="padding:clamp(22px,3.4cqw,44px);border-right:2px solid #2A2A2A;display:flex;flex-direction:column;gap:14px"
						>
							<div
								style="font-family:'Archivo Black',sans-serif;font-size:32px;text-transform:uppercase"
							>
								{tienda.envios.retiro ? 'Showroom' : 'Dónde estamos'}
							</div>
							<div style="display:flex;flex-direction:column;gap:12px;font-size:15px">
								<span style="display:flex;gap:12px"
									><Icon
										icon={I.phMapPinBold}
										size={19}
										style="color:var(--acento, #E6FF4D);flex:none"
									/>{tienda.direccion}</span
								>{#if tienda.envios.retiro}<span style="display:flex;gap:12px"
										><Icon
											icon={I.phStorefrontBold}
											size={19}
											style="color:var(--acento, #E6FF4D);flex:none"
										/>Retiro sin costo</span
									>{/if}
							</div>
						</div>
					{/if}
					{#if whatsapp}
						<div style="padding:clamp(22px,3.4cqw,44px);display:flex;flex-direction:column;gap:2px">
							<div
								style="font-family:'Archivo Black',sans-serif;font-size:32px;text-transform:uppercase;margin-bottom:12px"
							>
								Hablemos
							</div>
							<a
								href={wa('¡Hola!')}
								target="_blank"
								rel="noopener"
								style="display:flex;align-items:center;gap:12px;padding:16px;background:#25D366;color:#111;text-decoration:none;font:800 13px 'Archivo';letter-spacing:.1em;text-transform:uppercase"
								><Icon icon={I.phWhatsappLogoFill} size={22} />{tienda.whatsapp}</a
							>
						</div>
					{/if}
				</section>
			{/if}
		{/if}

		<!-- footer -->
		<footer style="padding:clamp(28px,4cqw,48px) clamp(14px,2.4cqw,28px) 22px">
			<div
				aria-hidden="true"
				style="font-family:'Archivo Black',sans-serif;font-size:clamp(36px,{pieNombre}cqw,320px);line-height:.78;letter-spacing:-.06em;color:#1C1C1C;user-select:none;text-transform:uppercase;overflow-wrap:break-word"
			>
				{tienda.nombre}
			</div>
			<div
				style="margin-top:20px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,180px),1fr));gap:24px"
			>
				{#each foot as f (f.t)}<div
						style="display:flex;flex-direction:column;gap:8px;font-size:14px"
					>
						<b
							style="font-size:12px;font-weight:800;letter-spacing:.14em;text-transform:uppercase;color:var(--acento, #E6FF4D);margin-bottom:4px"
							>{f.t}</b
						>{#each f.l as l (l.t)}<a href={l.href} style="color:#BDBDBD;text-decoration:none"
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
				style="margin-top:30px;padding-top:14px;border-top:2px solid #2A2A2A;display:flex;justify-content:space-between;gap:14px;flex-wrap:wrap;font-size:12px;font-weight:700;letter-spacing:.08em;text-transform:uppercase;color:#6E6E6E"
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
				style="position:fixed;right:18px;bottom:18px;z-index:30;width:56px;height:56px;background:#25D366;color:#111;display:grid;place-items:center;border:2px solid #111;box-shadow:4px 4px 0 var(--acento, #E6FF4D)"
				><Icon icon={I.phWhatsappLogoFill} size={28} /></a
			>{/if}
	</div>
</div>

<!-- eslint-enable svelte/no-navigation-without-resolve -->

<style>
	.mq {
		animation: mq 28s linear infinite;
	}
	@keyframes mq {
		to {
			transform: translateX(-50%);
		}
	}
	@media (prefers-reduced-motion: reduce) {
		.mq {
			animation: none;
		}
	}
	.celda {
		background: #111;
	}
	.celda:hover {
		background: #161616;
	}
	.zoom :global(img) {
		transition: transform 0.7s cubic-bezier(0.2, 0.7, 0.1, 1);
	}
	.celda:hover .zoom :global(img) {
		transform: scale(1.05);
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
