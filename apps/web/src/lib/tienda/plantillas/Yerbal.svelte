<script lang="ts">
	import { camposFactura, OPCIONALES } from '#lib/tienda/factura.ts';
	import { enhance } from '$app/forms';
	import { untrack } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
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
	import { comoLlega, cotizar, cotizarYa, cpValido, envia, textoEnvios } from '../envio.svelte.ts';
	import { Seleccion, tarjetaDe } from '../seleccion.svelte.ts';
	import { avisoMayorista } from '../lineas.ts';
	import { textoAviso, textoMinimo } from '../mayorista.ts';
	import { partirTitulo } from '../titulo.ts';
	import { Reserva } from '../reserva.svelte.ts';
	import { cuandoLargo, duracion } from '#lib/turnos/formato.ts';
	import type { OpcionEnvio, ProductoPublico, TiendaPublica } from '../tipos.ts';
	import { AltaNewsletter } from '../newsletter.svelte.ts';

	/**
	 * Plantilla Yerbal (design/Plantilla Yerbal.dc.html), con los datos de la tienda en lugar de los
	 * de ejemplo. El buscador, los filtros "en oferta / con stock" y la vista lista o grilla se
	 * resuelven en el navegador con los productos cargados. Lo que el diseño muestra sin datos atrás
	 * (presentaciones, orgánico y sin TACC, marcas, tabla nutricional, combos, turnos de entrega,
	 * cupones, factura A, cuenta regresiva) no se dibuja. El newsletter del pie anda de verdad (§50) y
	 * la venta por mayor también (§49): la barra de arriba, la tabla «Precio por mayor» de la ficha, el
	 * aviso del carrito y «¿Tenés un comercio?», con el mínimo y los precios mayoristas de la tienda.
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
		const old = p.anterior && !t.desde ? p.anterior : null;
		return {
			id: p.id,
			n: p.nombre,
			img: img(p.foto, 300, 300),
			price: (t.desde ? 'Desde ' : '') + fmt(t.precio),
			old: old ? fmt(old) : null,
			off: old ? `−${Math.round((1 - t.precio / old) * 100)}%` : null,
			tag: t.agotado ? 'Sin stock' : p.etiqueta,
			transf: fmt(t.precio * (1 - tienda.cobros.transferencia / 100)),
			stock: p.stock,
			agotado: t.agotado,
			href: `${base}/producto/${p.slug ?? p.id}`
		};
	};
	type Tarjeta = ReturnType<typeof tarjeta>;
	const whatsapp = $derived(tienda.whatsapp.replace(/\D/g, ''));
	const wa = (texto: string) =>
		`https://wa.me/${whatsapp.startsWith('54') ? whatsapp : '54' + whatsapp}?text=${encodeURIComponent(texto)}`;
	// "En oferta" va en la URL (?ofertas=1), así los botones de "Ver ofertas" llegan filtrados.
	const soloOferta = $derived(page.url.searchParams.has('ofertas'));
	// Sin logo, la inicial en el cuadradito de la marca.
	const inicial = $derived(tienda.nombre.trim().charAt(0).toUpperCase());
	const ver = (s: string) => tienda.secciones.includes(s as TiendaPublica['secciones'][number]);
	// Las líneas con variante no se tocan desde la tarjeta: ahí se elige en la ficha.
	const enCarrito = (id: string) =>
		carrito.lineas.find((l) => l.id === id && !l.varianteId)?.cantidad ?? 0;
	/** Sumar desde una tarjeta: con opciones, primero se elige en la ficha. */
	const sumar = (id: string) => {
		const p = porId.get(id);
		if (!p || tienda.vitrina) return;
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- ruta de la tienda con su slug
		if (!carrito.sumarProducto(p)) goto(`${base}/producto/${p.slug ?? p.id}`);
	};

	// Título de la portada, en dos renglones como en el diseño.
	const titulo = $derived(partirTitulo(tienda.portada.titulo));

	// Buscador: filtra los productos ya cargados (no hay búsqueda en el servidor).
	let busca = $state('');
	const sinTildes = (s: string) => s.normalize('NFD').replace(/[̀-ͯ]/g, '').toLowerCase();
	const encontrados = $derived.by(() => {
		const q = sinTildes(busca.trim());
		if (q.length < 2) return [];
		return tienda.productos
			.filter((p) => sinTildes(`${p.nombre} ${p.categoria ?? ''} ${p.descripcion}`).includes(q))
			.slice(0, 6);
	});

	// Inicio.
	const ofertas = $derived(
		tienda.productos
			.filter((p) => p.anterior && !tarjetaDe(p).agotado)
			.sort((a, b) => a.precio / a.anterior! - b.precio / b.anterior!)
	);
	// Con una oferta que termina (§52), la tarjeta es la «Oferta relámpago» del handoff, con su cuenta.
	const relampago = $derived(conteo.actual?.tipo === 'oferta' ? conteo.actual.producto : null);
	const oferta = $derived(relampago ?? ofertas[0] ?? null);
	// Lo que va al lado del banner de la portada: la oferta y el descuento por transferencia.
	const ladoHero = $derived(!!oferta || tienda.cobros.transferencia > 0);
	const maxOff = $derived(oferta ? Math.round((1 - oferta.precio / oferta.anterior!) * 100) : 0);
	const best = $derived(
		tienda.masVendidos.slice(0, 6).flatMap((id) => (porId.get(id) ? [tarjeta(porId.get(id)!)] : []))
	);
	const TILE_BG = [
		'#E8EEDF',
		'#DDE8DF',
		'#F3EBD3',
		'#EFE3D6',
		'#E3EDE0',
		'#F2EED2',
		'#F3E2D8',
		'#E3E9F2'
	];
	const nav = $derived([
		...(tienda.turnos
			? [{ t: 'Reservar turno', i: I.phCalendarCheckBold, href: `${base}/turnos` }]
			: []),
		...(ofertas.length
			? [{ t: 'Ofertas', i: I.phTagBold, href: enlaceCat(null, 'vendidos', true) }]
			: []),
		...tienda.categorias.slice(0, 6).map((c) => ({
			t: c.nombre,
			i: I.phPackageBold,
			href: enlaceCat(c.nombre)
		})),
		{ t: 'Nosotros', i: I.phStorefrontBold, href: `${base}/nosotros` }
	]);
	let cpInicio = $state('');
	let cotizadoInicio = $state<OpcionEnvio[] | null>(null);
	const iconoEnvio = (id: OpcionEnvio['id']) =>
		id === 'retiro' ? I.phStorefrontBold : id === 'moto' ? I.phMopedBold : I.phTruckBold;

	// Catálogo: categoría, orden, filtros del navegador y de a 24.
	const ORDENES = [
		['vendidos', 'Más vendidos'],
		['menor', 'Menor precio'],
		['mayor', 'Mayor precio'],
		['az', 'De la A a la Z']
	] as const;
	let soloStock = $state(false);
	let filtrosAbiertos = $state(false);
	let vista = $state<'lista' | 'grilla'>('lista');
	const listado = $derived.by(() => {
		const l = tienda.productos.filter(
			(p) =>
				(!categoria || p.categoria === categoria) &&
				(!soloOferta || p.anterior) &&
				(!soloStock || !tarjetaDe(p).agotado)
		);
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
	const filtros = $derived([
		{
			t: 'En oferta',
			n: tienda.productos.filter((p) => (!categoria || p.categoria === categoria) && p.anterior)
				.length,
			on: soloOferta,
			cambiar: () =>
				// eslint-disable-next-line svelte/no-navigation-without-resolve -- ruta de la tienda con su slug
				goto(enlaceCat(categoria, orden, !soloOferta), { replace: true })
		},
		{
			t: 'Con stock',
			n: tienda.productos.filter(
				(p) => (!categoria || p.categoria === categoria) && !tarjetaDe(p).agotado
			).length,
			on: soloStock,
			cambiar: () => (soloStock = !soloStock)
		}
	]);
	function enlaceCat(c: string | null, o = orden, of = false) {
		// eslint-disable-next-line svelte/prefer-svelte-reactivity -- local, sólo arma el link
		const q = new URLSearchParams();
		if (c) q.set('cat', c);
		if (o !== 'vendidos') q.set('orden', o);
		if (of) q.set('ofertas', '1');
		return `${base}/catalogo${q.size ? '?' + q : ''}`;
	}

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
			: []
	);
	// "Combinalo con": el producto y dos de su categoría, para sumarlos juntos (los otros, sin
	// opciones: no hay dónde elegirlas).
	const juntos = $derived.by(() => {
		const otros = related.filter((p) => !p.opciones?.length);
		return producto && !tarjetaDe(producto).agotado && otros.length
			? [producto, ...otros.slice(0, 2)]
			: [];
	});
	const resenasProd = $derived(
		producto ? tienda.resenas.filter((r) => r.producto === producto!.nombre) : []
	);
	const promedioProd = $derived(
		resenasProd.length
			? Math.round((resenasProd.reduce((s, r) => s + r.estrellas, 0) / resenasProd.length) * 10) /
					10
			: 0
	);
	// Presentación, molienda…: lo elegido, la variante, su precio, stock y foto (seleccion.svelte.ts).
	const sel = $derived(producto ? new Seleccion(producto) : null);
	// El tachado no va con "Desde": no se sabe a qué presentación se compara.
	const tachado = $derived(sel && !sel.desde ? sel.anterior : null);
	let cantidad = $state(1);
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
	function sumarJuntos() {
		if (!sel || tienda.vitrina) return;
		// Con opciones, el de la ficha va con lo elegido (si falta, se marca y no se suma nada).
		if ((sel.conOpciones || !enCarrito(sel.producto.id)) && !sel.agregar(carrito)) return;
		for (const p of juntos.slice(1)) if (!enCarrito(p.id)) carrito.sumarProducto(p);
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- ruta de la tienda con su slug
		goto(`${base}/carrito`);
	}

	// Carrito.
	const lineas = $derived(carrito.detalle(tienda.productos, tienda.mayorista ?? null));
	const subtotal = $derived(lineas.reduce((s, l) => s + l.subtotal, 0));
	const ahorro = $derived(
		lineas.reduce(
			(s, l) =>
				s + (l.producto.anterior ? (l.producto.anterior - l.producto.precio) * l.cantidad : 0),
			0
		)
	);
	const falta = $derived(
		tienda.envios.gratisDesde ? Math.max(0, tienda.envios.gratisDesde - subtotal) : 0
	);
	// Venta por mayor (§49): cuánto falta para el precio mayorista, o lo que ya se ahorra.
	const porMayor = $derived(avisoMayorista(lineas, tienda.mayorista ?? null));
	const avisoMayor = $derived(porMayor && textoAviso(porMayor, fmt));
	const minimoMayor = $derived(tienda.mayorista ? textoMinimo(tienda.mayorista, fmt) : '');
	// La tabla «Precio por mayor» de la ficha: el precio de lista y el mayorista con su mínimo.
	const escalas = $derived.by(() => {
		const m = tienda.mayorista;
		const mayor = sel?.mayorista ?? null;
		if (!m || !sel || mayor === null) return [];
		const desde =
			m.unidades !== null && m.monto !== null
				? `${m.unidades} u. y ${fmt(m.monto)}`
				: m.unidades !== null
					? `${m.unidades} u. o más`
					: `Desde ${fmt(m.monto ?? 0)}`;
		const hasta =
			m.unidades !== null && m.monto === null
				? m.unidades > 2
					? `1 a ${m.unidades - 1} u.`
					: '1 u.'
				: m.unidades === null
					? `Menos de ${fmt(m.monto ?? 0)}`
					: 'Por unidad';
		return [
			{ q: hasta, p: fmt(sel.precio), s: '', on: !porMayor?.aplicado },
			{
				q: desde,
				p: fmt(mayor),
				s: `−${Math.round((1 - mayor / sel.precio) * 100)}%`,
				on: !!porMayor?.aplicado
			}
		];
	});
	const upsell = $derived(
		tienda.masVendidos
			.filter(
				(id) =>
					!carrito.lineas.some((l) => l.id === id) &&
					porId.has(id) &&
					!tarjetaDe(porId.get(id)!).agotado
			)
			.slice(0, 1)
			.map((id) => porId.get(id)!)
	);
	const textoPedido = $derived(
		`¡Hola! Quiero pedir:\n${lineas.map((l) => `${l.cantidad} × ${l.producto.nombre}${l.variante ? ` ${l.variante}` : ''} (${fmt(l.subtotal)})`).join('\n')}\nTotal: ${fmt(subtotal)}`
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
						i: I.phBankBold,
						c: 'var(--acento, #2F5D50)'
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
			t: 'Comprar',
			l: [
				{ t: 'Todo', href: `${base}/catalogo` },
				...tienda.categorias.slice(0, 4).map((c) => ({ t: c.nombre, href: enlaceCat(c.nombre) }))
			]
		},
		{
			t: 'Ayuda',
			l: [
				{ t: 'Envíos', href: `${base}/nosotros` },
				{ t: 'Medios de pago', href: `${base}/nosotros` },
				{ t: 'Cambios', href: `/t/${tienda.slug}/arrepentimiento` }
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
	// Reseñas y preguntas con las piezas de Yerbal: Schibsted Grotesk, tarjetas blancas, esquinas de 10 px.
	const tema: TemaOpiniones = {
		texto: '#1E2B22',
		suave: '#6B7A6E',
		campo: '#EEF1E8',
		tarjeta: '#fff',
		borde: 'rgba(30,43,34,.14)',
		acento: '#1E2B22',
		sobreAcento: '#EEF1E8',
		estrella: '#D9A400',
		radio: '10px',
		fuente: "'Schibsted Grotesk', sans-serif",
		titulos: "'Schibsted Grotesk', sans-serif",
		pesoTitulos: 800
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
</script>

<svelte:window bind:innerWidth={ancho} />

<svelte:head>
	<link
		href="https://fonts.googleapis.com/css2?family=Schibsted+Grotesk:wght@400;500;600;700;800&display=swap"
		rel="stylesheet"
	/>
</svelte:head>

{#snippet buscador()}
	<label
		style="flex:1;min-width:0;display:flex;align-items:center;gap:8px;padding:10px 14px;border-radius:10px;background:#fff;border:1.5px solid rgba(30,43,34,.12);font-size:14px;color:#6B7A6E"
		><Icon icon={I.phMagnifyingGlassBold} size={16} style="flex:none" /><input
			bind:value={busca}
			type="search"
			aria-label="Buscar productos"
			placeholder="Buscá entre {tienda.productos.length} productos"
			style="flex:1;min-width:0;padding:0;border:0;background:transparent;font:400 14px 'Schibsted Grotesk';color:#1E2B22;outline:none"
		/></label
	>
{/snippet}

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
{#snippet paso(p: Tarjeta, alto: number)}
	{@const q = enCarrito(p.id)}
	{#if q}<span
			style="display:flex;align-items:center;justify-content:space-between;padding:3px;border-radius:9px;background:var(--acento, #2F5D50);color:#fff;flex:none"
			><button
				type="button"
				aria-label="Uno menos"
				onclick={() => carrito.fijar(p.id, q - 1)}
				style="width:{alto}px;height:{alto}px;display:grid;place-items:center;border:0;padding:0;background:transparent;color:#fff;cursor:pointer"
				><Icon icon={I.phMinusBold} size={13} /></button
			><b style="min-width:16px;text-align:center;font-size:14px">{q}</b><button
				type="button"
				aria-label="Uno más"
				disabled={q >= p.stock}
				onclick={() => sumar(p.id)}
				style="width:{alto}px;height:{alto}px;display:grid;place-items:center;border:0;padding:0;background:transparent;color:#fff;cursor:pointer"
				><Icon icon={I.phPlusBold} size={13} /></button
			></span
		>{:else}<button
			type="button"
			onclick={() => sumar(p.id)}
			disabled={p.agotado || tienda.vitrina}
			aria-label="Agregar {p.n}"
			style="display:flex;justify-content:center;align-items:center;gap:6px;padding:9px;border-radius:9px;border:1.5px solid var(--acento, #2F5D50);background:transparent;color:var(--acento, #2F5D50);font:700 13.5px 'Schibsted Grotesk';cursor:pointer"
			><Icon icon={I.phPlusBold} size={13} />{p.agotado ? 'Sin stock' : 'Agregar'}</button
		>{/if}
{/snippet}

{#snippet tarjetaGrilla(p: Tarjeta)}
	<div
		style="padding:10px;border-radius:14px;background:#fff;display:flex;flex-direction:column;gap:6px;min-width:0"
	>
		<a
			href={p.href}
			class="zoom"
			style="position:relative;display:block;aspect-ratio:1;border-radius:10px;overflow:hidden;background:#F2F4EE"
			>{#if p.img}<img
					src={p.img}
					alt=""
					style="width:100%;height:100%;object-fit:cover"
				/>{/if}{#if p.off}<span
					style="position:absolute;top:6px;left:6px;padding:3px 7px;border-radius:5px;background:#C2412D;color:#fff;font-size:11.5px;font-weight:800"
					>{p.off}</span
				>{/if}</a
		>
		<a
			href={p.href}
			style="font-size:13.5px;font-weight:600;line-height:1.3;min-height:35px;color:#1E2B22;text-decoration:none"
			>{p.n}</a
		>
		{#if p.tag}<span style="font-size:12px;color:#6B7A6E">{p.tag}</span>{/if}
		<span style="display:flex;align-items:baseline;gap:6px;flex-wrap:wrap"
			><b style="font-size:17px;letter-spacing:-.02em">{p.price}</b>{#if p.old}<s
					style="font-size:12px;color:#8A968C">{p.old}</s
				>{/if}</span
		>
		{#if tienda.cobros.transferencia > 0}<span
				style="font-size:11.5px;color:var(--acento, #2F5D50);font-weight:600"
				>{p.transf} por transferencia</span
			>{/if}
		<span style="flex:1"></span>
		{@render paso(p, 30)}
	</div>
{/snippet}

<div
	class="yerbal"
	style="min-height:100vh;font-family:'Schibsted Grotesk',system-ui,sans-serif;font-variant-numeric:normal;color:#1E2B22;background:#EEF1E8"
>
	<div
		style="container-type:inline-size;width:100%;background:#EEF1E8;position:relative;overflow-x:clip"
	>
		<!-- barra promo -->
		<div
			style="display:flex;justify-content:center;align-items:center;gap:10px;flex-wrap:wrap;padding:8px 14px;background:var(--acento, #2F5D50);color:#EEF1E8;font-size:13px;font-weight:600;text-align:center"
		>
			<Icon icon={I.phTruckBold} size={15} /><span
				>{tienda.envios.gratisDesde
					? `Envío gratis desde ${fmt(tienda.envios.gratisDesde)}`
					: tienda.cobros.cuotas > 1
						? `${tienda.cobros.cuotas} cuotas sin interés`
						: 'Compra protegida'}</span
			>{#if wide && tienda.mayorista}<span style="opacity:.6">·</span><span
					>Comprás por mayor desde {minimoMayor}</span
				>{:else if wide && tienda.cobros.transferencia > 0}<span style="opacity:.6">·</span><span
					>{tienda.cobros.transferencia}% off pagando por transferencia</span
				>{/if}
		</div>

		<!-- header -->
		<header
			style="position:sticky;top:0;z-index:20;background:#EEF1E8;border-bottom:1px solid rgba(30,43,34,.1)"
		>
			<div style="display:flex;align-items:center;gap:14px;padding:12px clamp(12px,3cqw,32px)">
				<a
					href={base || '/'}
					style="display:flex;align-items:center;gap:8px;min-width:0;flex:{nar
						? '0 1 auto'
						: 'none'};color:#1E2B22;text-decoration:none"
					>{#if !tienda.logo}<span
							style="width:36px;height:36px;border-radius:10px;background:var(--acento, #2F5D50);color:#E8C547;display:grid;place-items:center;flex:none;font-weight:800;font-size:19px;letter-spacing:-.03em"
							>{inicial}</span
						>{/if}<span
						style="font-weight:800;font-size:21px;letter-spacing:-.03em;white-space:nowrap;overflow:hidden;text-overflow:ellipsis"
						><LogoTienda {tienda} /></span
					></a
				>
				{#if wide && tienda.productos.length}{@render buscador()}{:else}<span style="flex:1"
					></span>{/if}
				{#if nar && tienda.turnos}<a
						href="{base}/turnos"
						aria-label="Reservar turno"
						style="display:grid;place-items:center;flex:none;width:40px;height:40px;border-radius:10px;background:var(--acento, #2F5D50);color:#E8C547"
						><Icon icon={I.phCalendarCheckBold} size={18} /></a
					>{/if}
				<a
					href="{base}/carrito"
					aria-label="Carrito"
					style="display:flex;align-items:center;gap:8px;flex:none;padding:10px 14px;border-radius:10px;background:#1E2B22;color:#EEF1E8;font:700 14px 'Schibsted Grotesk';white-space:nowrap;text-decoration:none"
					><Icon
						icon={I.phShoppingCartSimpleBold}
						size={17}
					/>{carrito.unidades}{#if wide && subtotal}<span style="opacity:.7;font-weight:500"
							>· {fmt(subtotal)}</span
						>{/if}</a
				>
			</div>
			{#if nar && tienda.productos.length}<div
					style="padding:0 clamp(12px,3cqw,32px) 10px;display:flex"
				>
					{@render buscador()}
				</div>{/if}
			{#if encontrados.length || busca.trim().length >= 2}
				<div
					style="position:absolute;top:100%;left:clamp(12px,3cqw,32px);right:clamp(12px,3cqw,32px);z-index:30;padding:6px;border-radius:14px;background:#fff;box-shadow:0 18px 40px -16px rgba(30,43,34,.45);display:flex;flex-direction:column;gap:2px"
				>
					{#each encontrados as p (p.id)}<a
							href="{base}/producto/{p.slug ?? p.id}"
							onclick={() => (busca = '')}
							style="display:flex;align-items:center;gap:12px;padding:8px;border-radius:10px;color:#1E2B22;text-decoration:none"
							>{#if p.foto}<img
									src={img(p.foto, 120, 120)}
									alt=""
									style="width:44px;height:44px;border-radius:9px;object-fit:cover;flex:none"
								/>{/if}<span
								style="flex:1;min-width:0;font-size:14px;font-weight:600;white-space:nowrap;overflow:hidden;text-overflow:ellipsis"
								>{p.nombre}</span
							><b style="font-size:14.5px;white-space:nowrap">{fmt(p.precio)}</b></a
						>{:else}<span style="padding:10px;font-size:13.5px;color:#6B7A6E"
							>No encontramos productos con «{busca.trim()}».</span
						>{/each}
				</div>
			{/if}
			{#if wide}<nav
					style="display:flex;gap:2px;padding:0 clamp(12px,3cqw,32px) 8px;overflow-x:auto;scrollbar-width:none"
				>
					{#each nav as n (n.t)}<a
							href={n.href}
							style="flex:none;display:flex;align-items:center;gap:6px;font:600 13.5px 'Schibsted Grotesk';padding:7px 11px;border-radius:8px;background:{(
								n.href === `${base}/turnos`
									? pagina === 'turnos' || pagina === 'turno'
									: pagina === 'catalogo' && categoria === n.t
							)
								? '#DDE8DF'
								: 'transparent'};color:#1E2B22;white-space:nowrap;text-decoration:none"
							><Icon icon={n.i} size={15} style="color:var(--acento, #2F5D50)" />{n.t}</a
						>{/each}
				</nav>{/if}
		</header>

		{#if tienda.vitrina}
			<div
				style="padding:10px 16px;background:#E8C547;color:#1E2B22;font-size:13px;font-weight:600;text-align:center"
			>
				Es una tienda de ejemplo de la plantilla Yerbal: podés recorrerla, pero no se compra
				de&nbsp;verdad.
			</div>
		{/if}

		<!-- ============ INICIO ============ -->
		{#if pagina === 'inicio'}
			<!-- Las secciones en el orden que eligió la tienda en Mi tienda. -->
			{#each enOrden(tienda.secciones, ['hero', 'cats', 'prods', 'revs']) as seccion (seccion)}
				{#if seccion === 'hero'}
					{#if ver('hero')}
						<section
							style="padding:clamp(12px,2cqw,24px) clamp(12px,3cqw,32px) 0;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,320px),1fr));gap:10px"
						>
							<div
								style="grid-column:{nar
									? 'span 1'
									: ladoHero
										? 'span 2'
										: '1 / -1'};position:relative;min-height:min(440px,100cqw);border-radius:16px;overflow:hidden;background:var(--acento, #2F5D50);color:#EEF1E8;padding:clamp(22px,4cqw,48px);display:flex;flex-direction:column;justify-content:flex-end"
							>
								{#if tienda.portada.foto}<img
										src={img(tienda.portada.foto, 1400)}
										alt=""
										style="position:absolute;inset:0;width:100%;height:100%;object-fit:cover;opacity:.55"
									/>{/if}
								<div
									style="position:absolute;inset:0;background:linear-gradient(0deg,rgba(20,40,32,.9),rgba(20,40,32,.1) 70%)"
								></div>
								<div style="position:relative;max-width:560px;min-width:0">
									{#if maxOff}<span
											style="display:inline-flex;padding:5px 10px;border-radius:6px;background:#E8C547;color:#1E2B22;font-size:12.5px;font-weight:800"
											>OFERTAS · HASTA {maxOff}% OFF</span
										>{/if}
									<h1
										style="margin:14px 0 0;font-weight:800;font-size:clamp(40px,6.4cqw,80px);line-height:.92;letter-spacing:-.045em;overflow-wrap:break-word"
									>
										{titulo[0]}{#if titulo[1]}<br />{titulo[1]}{/if}
									</h1>
									{#if tienda.portada.lema}<p
											style="margin:14px 0 0;font-size:16px;line-height:1.5;color:#D6E0D4"
										>
											{tienda.portada.lema}
										</p>{/if}
									<div style="margin-top:20px;display:flex;gap:8px;flex-wrap:wrap">
										<a
											href={ofertas.length ? enlaceCat(null, 'vendidos', true) : `${base}/catalogo`}
											style="padding:14px 22px;border-radius:10px;background:#E8C547;color:#1E2B22;font:800 15px 'Schibsted Grotesk';text-decoration:none"
											>{ofertas.length ? 'Ver ofertas' : 'Ver productos'}</a
										>{#if tienda.categorias[0]}<a
												href={enlaceCat(tienda.categorias[0].nombre)}
												style="padding:14px 20px;border:1.5px solid rgba(238,241,232,.4);border-radius:10px;color:#EEF1E8;font:700 15px 'Schibsted Grotesk';text-decoration:none"
												>{tienda.categorias[0].nombre}</a
											>{/if}
									</div>
								</div>
							</div>
							{#if ladoHero}<div
									style="display:grid;grid-template-rows:{oferta
										? '1fr 1fr'
										: '1fr'};gap:10px;min-width:0"
								>
									{#if oferta}
										<a
											href="{base}/producto/{oferta.slug ?? oferta.id}"
											style="padding:20px;border-radius:16px;background:#E8C547;color:#1E2B22;text-decoration:none;display:flex;flex-direction:column;justify-content:space-between;gap:10px"
											><span style="display:flex;justify-content:space-between;align-items:start"
												><b style="font-size:13px;letter-spacing:.04em"
													>{relampago ? 'OFERTA RELÁMPAGO' : 'OFERTA'}</b
												>{#if relampago}<span
														style="display:flex;gap:3px;font-variant-numeric:tabular-nums"
														>{#each conteo.partes(3) as c (c.u)}<span
																style="padding:3px 6px;border-radius:5px;background:#1E2B22;color:#E8C547;font-size:12.5px;font-weight:700"
																>{c.v}</span
															>{/each}</span
													>{:else}<span
														style="padding:3px 6px;border-radius:5px;background:#1E2B22;color:#E8C547;font-size:12.5px;font-weight:700"
														>−{maxOff}%</span
													>{/if}</span
											><span style="display:flex;gap:12px;align-items:center;min-width:0"
												>{#if oferta.foto}<img
														src={img(oferta.foto, 200, 200)}
														alt=""
														style="width:76px;height:76px;border-radius:12px;object-fit:cover;flex:none"
													/>{/if}<span style="min-width:0"
													><b style="display:block;font-size:17px;letter-spacing:-.02em"
														>{oferta.nombre}</b
													><span style="margin-top:4px;display:flex;align-items:baseline;gap:8px"
														><b style="font-size:24px;letter-spacing:-.02em">{fmt(oferta.precio)}</b
														><s style="font-size:13px;opacity:.6">{fmt(oferta.anterior!)}</s></span
													></span
												></span
											></a
										>
									{/if}
									{#if tienda.cobros.transferencia > 0}<div
											style="padding:20px;border-radius:16px;background:#fff;display:flex;align-items:center;gap:14px;flex-wrap:wrap"
										>
											<Icon
												icon={I.phBankBold}
												size={34}
												style="color:var(--acento, #2F5D50);flex:none"
											/>
											<div style="flex:1;min-width:160px">
												<b style="font-size:17px;letter-spacing:-.02em"
													>{tienda.cobros.transferencia}% off por transferencia</b
												>
												{#if tienda.cobros.cuotas > 1}<div
														style="font-size:13.5px;color:#56645A;margin-top:2px"
													>
														O hasta {tienda.cobros.cuotas} cuotas sin interés con&nbsp;tarjeta
													</div>{/if}
											</div>
											<a
												href="{base}/catalogo"
												style="padding:11px 14px;border-radius:10px;background:#1E2B22;color:#EEF1E8;font:700 13.5px 'Schibsted Grotesk';white-space:nowrap;text-decoration:none"
												>Comprar</a
											>
										</div>{/if}
								</div>{/if}
						</section>
					{/if}
				{:else if seccion === 'cats'}
					{#if ver('cats') && tienda.categorias.length}
						<section
							style="padding:12px clamp(12px,3cqw,32px) 0;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(30%,120px),1fr));gap:8px"
						>
							{#each tienda.categorias as c, i (c.nombre)}<a
									href={enlaceCat(c.nombre)}
									style="padding:14px 8px;border-radius:14px;background:#fff;color:#1E2B22;text-decoration:none;display:flex;flex-direction:column;align-items:center;gap:8px;min-width:0"
									><span
										style="width:48px;height:48px;border-radius:50%;overflow:hidden;background:{TILE_BG[
											i % TILE_BG.length
										]};display:grid;place-items:center"
										>{#if c.foto}<img
												src={img(c.foto, 120, 120)}
												alt=""
												style="width:100%;height:100%;object-fit:cover"
											/>{:else}<Icon
												icon={I.phPackageFill}
												size={24}
												style="color:var(--acento, #2F5D50)"
											/>{/if}</span
									><span
										style="font-size:13px;font-weight:700;text-align:center;line-height:1.2;overflow-wrap:anywhere"
										>{c.nombre}</span
									></a
								>{/each}
						</section>
					{/if}
				{:else if seccion === 'prods'}
					{#if ver('prods') && best.length}
						<section style="padding:clamp(32px,5cqw,56px) clamp(12px,3cqw,32px) 0">
							<div
								style="display:flex;justify-content:space-between;align-items:end;gap:12px;flex-wrap:wrap"
							>
								<h2
									style="margin:0;font-weight:800;font-size:clamp(28px,3.6cqw,42px);letter-spacing:-.04em"
								>
									Los más vendidos
								</h2>
								<a
									href="{base}/catalogo"
									style="font:700 14px 'Schibsted Grotesk';color:var(--acento, #2F5D50);display:flex;align-items:center;gap:6px;text-decoration:none"
									>Ver todos<Icon icon={I.phArrowRightBold} size={14} /></a
								>
							</div>
							<div
								style="margin-top:16px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(46%,190px),1fr));gap:8px"
							>
								{#each best as p (p.id)}{@render tarjetaGrilla(p)}{/each}
							</div>
						</section>
					{/if}

					{#if envia(tienda.envios)}<section
							style="padding:clamp(32px,5cqw,56px) clamp(12px,3cqw,32px) 0"
						>
							<div
								style="padding:clamp(18px,3cqw,32px);border-radius:16px;background:#fff;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,300px),1fr));gap:20px;align-items:center"
							>
								<div>
									<h2
										style="margin:0;font-weight:800;font-size:clamp(26px,3.2cqw,36px);letter-spacing:-.04em"
									>
										¿Cuándo te llega?
									</h2>
									<p style="margin:6px 0 0;font-size:14.5px;color:#56645A">
										Poné tu código postal y te decimos cuándo y&nbsp;cuánto.
									</p>
									<div style="margin-top:14px;display:flex;gap:6px;max-width:380px">
										<input
											bind:value={cpInicio}
											inputmode="numeric"
											maxlength="4"
											placeholder="Código postal"
											aria-label="Código postal"
											onkeydown={(e) =>
												e.key === 'Enter' &&
												cpValido(cpInicio) &&
												cotizarYa(cpInicio, subtotal, tienda).then((o) => (cotizadoInicio = o))}
											style="flex:1;min-width:0;padding:12px 14px;border-radius:10px;border:1.5px solid #1E2B22;background:transparent;font:600 15px 'Schibsted Grotesk';color:#1E2B22"
										/><button
											type="button"
											onclick={() =>
												cpValido(cpInicio) &&
												cotizarYa(cpInicio, subtotal, tienda).then((o) => (cotizadoInicio = o))}
											style="padding:12px 18px;border:0;border-radius:10px;background:#1E2B22;color:#EEF1E8;font:700 14px 'Schibsted Grotesk';cursor:pointer"
											>Calcular</button
										>
									</div>
								</div>
								{#if cotizadoInicio}
									<div style="display:flex;flex-direction:column;gap:6px">
										{#each cotizadoInicio as s (s.id)}<div
												style="display:flex;align-items:center;gap:12px;padding:12px 14px;border-radius:12px;background:#F2F4EE"
											>
												<Icon
													icon={iconoEnvio(s.id)}
													size={20}
													style="color:var(--acento, #2F5D50);flex:none"
												/><span style="flex:1;min-width:0;font-size:14px"
													><b>{s.nombre}</b> · {s.detalle}</span
												><b
													style="font-size:14px;white-space:nowrap;color:{s.precio
														? '#1E2B22'
														: 'var(--acento, #2F5D50)'}">{s.precio ? fmt(s.precio) : 'Gratis'}</b
												>
											</div>{:else}<span style="font-size:14px;color:#56645A"
												>No llegamos a ese código postal&nbsp;todavía.</span
											>{/each}
									</div>
								{/if}
							</div>
						</section>{/if}
				{:else if seccion === 'revs'}
					{#if ver('revs') && tienda.resenas.length}
						<section style="padding:clamp(32px,5cqw,56px) clamp(12px,3cqw,32px) 0">
							<h2
								style="margin:0;font-weight:800;font-size:clamp(28px,3.6cqw,42px);letter-spacing:-.04em"
							>
								{tienda.promedio
									? `${tienda.promedio.toLocaleString('es-AR')} de promedio en reseñas`
									: 'Lo que dicen'}
							</h2>
							<div
								style="margin-top:16px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,260px),1fr));gap:8px"
							>
								{#each tienda.resenas.slice(0, 4) as r, i (i)}
									{@const pr = tienda.productos.find((p) => p.nombre === r.producto)}
									<div
										style="display:flex;gap:12px;padding:12px;border-radius:14px;background:#fff"
									>
										{#if pr?.foto}<img
												src={img(pr.foto, 160, 160)}
												alt=""
												style="width:76px;height:76px;border-radius:10px;object-fit:cover;flex:none"
											/>{/if}
										<div style="min-width:0">
											<div style="display:flex;gap:1px;color:#D9A400">
												{#each estrellas(r.estrellas) as e, j (j)}<Icon icon={e} size={13} />{/each}
											</div>
											<p style="margin:5px 0 0;font-size:13.5px;line-height:1.4">{r.texto}</p>
											<div style="margin-top:4px;font-size:12px;color:#6B7A6E">
												{r.autor}{#if r.producto}
													· {r.conCompra ? 'compró' : 'sobre'} {r.producto}{/if}
											</div>
											{#if r.conCompra}<div
													style="margin-top:5px;display:flex;align-items:center;gap:4px;font-size:12px;font-weight:700;color:var(--acento, #2F5D50)"
												>
													<Icon icon={I.phSealCheckFill} size={13} />Compra verificada
												</div>{/if}
											{#if r.respuesta}<div
													style="margin-top:8px;padding:8px 10px;border-radius:10px;background:#EEF1E8;font-size:12.5px;line-height:1.45;color:#56645A"
												>
													<b style="color:#1E2B22">Respuesta de {tienda.nombre}</b><br
													/>{r.respuesta}
												</div>{/if}
										</div>
									</div>
								{/each}
							</div>
						</section>
					{/if}
				{/if}
			{/each}
			<div style="height:clamp(32px,5cqw,56px)"></div>
		{/if}

		<!-- ============ CATÁLOGO ============ -->
		{#if pagina === 'catalogo' && !tienda.productos.length}
			<section style="padding:clamp(40px,7cqw,96px) clamp(16px,4cqw,48px)">
				<VacioTienda {tienda} {tema} motivo="sin-productos" />
			</section>
		{:else if pagina === 'catalogo'}
			<section
				style="padding:16px clamp(12px,3cqw,32px) clamp(40px,6cqw,72px);display:flex;gap:clamp(14px,2.4cqw,28px);align-items:flex-start"
			>
				{#if wide}
					<aside
						style="width:240px;flex:none;position:sticky;top:120px;display:flex;flex-direction:column;gap:18px;font-size:13.5px"
					>
						<div
							style="padding:12px;border-radius:14px;background:#fff;display:flex;flex-direction:column;gap:2px"
						>
							{#each [{ t: 'Todo', c: null, n: tienda.productos.length }, ...tienda.categorias.map( (c) => ({ t: c.nombre, c: c.nombre, n: c.cantidad }) )] as c (c.t)}{@const on =
									c.c === categoria}<a
									href={enlaceCat(c.c, orden, soloOferta)}
									style="display:flex;justify-content:space-between;gap:8px;padding:8px 10px;border-radius:8px;background:{on
										? 'var(--acento, #2F5D50)'
										: 'transparent'};color:{on ? '#fff' : '#1E2B22'};font-weight:{on
										? 700
										: 500};text-decoration:none">{c.t}<span style="opacity:.6">{c.n}</span></a
								>{/each}
						</div>
						<div
							style="padding:14px;border-radius:14px;background:#fff;display:flex;flex-direction:column;gap:10px"
						>
							<b style="font-size:12px;letter-spacing:.08em">FILTRAR</b>
							{#each filtros as f (f.t)}<button
									type="button"
									role="checkbox"
									aria-checked={f.on}
									onclick={f.cambiar}
									style="display:flex;align-items:center;gap:9px;padding:0;border:0;background:transparent;font:400 13.5px 'Schibsted Grotesk';color:#1E2B22;cursor:pointer;text-align:left"
									><span
										style="width:18px;height:18px;border-radius:5px;border:1.5px solid var(--acento, #2F5D50);background:{f.on
											? 'var(--acento, #2F5D50)'
											: 'transparent'};display:grid;place-items:center;color:#fff;flex:none"
										>{#if f.on}<Icon icon={I.phCheckBold} size={11} />{/if}</span
									><span style="flex:1">{f.t}</span><span style="color:#8A968C">{f.n}</span></button
								>{/each}
						</div>
					</aside>
				{/if}
				<div style="flex:1;min-width:0">
					<div style="font-size:12.5px;color:#6B7A6E">
						<a href={base || '/'} style="color:inherit;text-decoration:none">Inicio</a> /
						<span style="color:#1E2B22">{categoria ?? 'Todo'}</span>
					</div>
					<div
						style="margin-top:6px;display:flex;justify-content:space-between;align-items:end;gap:12px;flex-wrap:wrap"
					>
						<h1
							style="margin:0;font-weight:800;font-size:clamp(32px,4.4cqw,52px);letter-spacing:-.045em;overflow-wrap:break-word;min-width:0"
						>
							{categoria ?? 'Todo'}
							<span style="font-size:.5em;color:#6B7A6E;font-weight:600">{listado.length}</span>
						</h1>
						<div style="display:flex;gap:6px;flex-wrap:wrap">
							{#if nar}<button
									type="button"
									aria-expanded={filtrosAbiertos}
									onclick={() => (filtrosAbiertos = !filtrosAbiertos)}
									style="display:flex;align-items:center;gap:6px;padding:9px 12px;border:0;border-radius:10px;background:#fff;font:700 13.5px 'Schibsted Grotesk';color:#1E2B22;cursor:pointer"
									><Icon icon={I.phFunnelBold} size={15} />Filtros{#if soloOferta || soloStock}
										· {Number(soloOferta) + Number(soloStock)}{/if}</button
								>{/if}<label
								style="display:flex;align-items:center;gap:6px;padding:9px 12px;border-radius:10px;background:#fff;font-size:13.5px;font-weight:700"
								><span
									style="position:absolute;width:1px;height:1px;overflow:hidden;clip:rect(0 0 0 0)"
									>Ordenar</span
								><select
									value={orden}
									onchange={(e) =>
										goto(enlaceCat(categoria, e.currentTarget.value, soloOferta), {
											replace: true
										})}
									style="border:0;padding:0;background:transparent;font:700 13.5px 'Schibsted Grotesk';color:#1E2B22;cursor:pointer"
									>{#each ORDENES as [k, t] (k)}<option value={k}>{t}</option>{/each}</select
								></label
							><span
								style="display:flex;padding:3px;border-radius:10px;background:#fff"
								role="radiogroup"
								aria-label="Vista"
								><button
									type="button"
									role="radio"
									aria-checked={vista === 'grilla'}
									aria-label="Grilla"
									onclick={() => (vista = 'grilla')}
									style="width:32px;height:32px;border:0;padding:0;border-radius:8px;background:{vista ===
									'grilla'
										? '#DDE8DF'
										: 'transparent'};display:grid;place-items:center;color:#1E2B22;cursor:pointer"
									><Icon icon={I.phSquaresFourBold} size={16} /></button
								><button
									type="button"
									role="radio"
									aria-checked={vista === 'lista'}
									aria-label="Lista"
									onclick={() => (vista = 'lista')}
									style="width:32px;height:32px;border:0;padding:0;border-radius:8px;background:{vista ===
									'lista'
										? '#DDE8DF'
										: 'transparent'};display:grid;place-items:center;color:#1E2B22;cursor:pointer"
									><Icon icon={I.phListBulletsBold} size={16} /></button
								></span
							>
						</div>
					</div>
					{#if nar && filtrosAbiertos}
						<div
							style="margin-top:12px;padding:14px;border-radius:14px;background:#fff;display:flex;flex-direction:column;gap:10px;font-size:13.5px"
						>
							{#each filtros as f (f.t)}<button
									type="button"
									role="checkbox"
									aria-checked={f.on}
									onclick={f.cambiar}
									style="display:flex;align-items:center;gap:9px;padding:0;border:0;background:transparent;font:400 13.5px 'Schibsted Grotesk';color:#1E2B22;cursor:pointer;text-align:left"
									><span
										style="width:18px;height:18px;border-radius:5px;border:1.5px solid var(--acento, #2F5D50);background:{f.on
											? 'var(--acento, #2F5D50)'
											: 'transparent'};display:grid;place-items:center;color:#fff;flex:none"
										>{#if f.on}<Icon icon={I.phCheckBold} size={11} />{/if}</span
									><span style="flex:1">{f.t}</span><span style="color:#8A968C">{f.n}</span></button
								>{/each}
						</div>
					{/if}
					<div style="margin-top:12px;display:flex;gap:6px;overflow-x:auto;scrollbar-width:none">
						{#each [null, ...tienda.categorias.map((c) => c.nombre)] as c (c ?? '')}{@const on =
								c === categoria}<a
								href={enlaceCat(c, orden, soloOferta)}
								style="flex:none;padding:8px 13px;border-radius:999px;background:{on
									? '#1E2B22'
									: '#fff'};color:{on
									? '#EEF1E8'
									: '#1E2B22'};font-size:13px;font-weight:700;white-space:nowrap;text-decoration:none"
								>{c ?? 'Todos'}</a
							>{/each}
					</div>
					{#if !catalog.length}
						<p style="margin:28px 0;font-size:15px;color:#6B7A6E">
							No hay productos con estos&nbsp;filtros.
						</p>
					{:else if vista === 'grilla'}
						<div
							style="margin-top:14px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(46%,190px),1fr));gap:8px"
						>
							{#each catalog as p (p.id)}{@render tarjetaGrilla(p)}{/each}
						</div>
					{:else}
						<div style="margin-top:14px;display:flex;flex-direction:column;gap:6px">
							{#each catalog as p (p.id)}
								{@const q = enCarrito(p.id)}
								<div
									style="display:flex;align-items:center;gap:12px;padding:8px;border-radius:12px;background:#fff;min-width:0"
								>
									<a
										href={p.href}
										aria-label={p.n}
										style="position:relative;display:block;width:64px;height:64px;flex:none;border-radius:9px;overflow:hidden;background:#F2F4EE"
										>{#if p.img}<img
												src={p.img}
												alt=""
												style="width:100%;height:100%;object-fit:cover"
											/>{/if}{#if p.off}<span
												style="position:absolute;left:3px;top:3px;padding:1px 5px;border-radius:4px;background:#C2412D;color:#fff;font-size:10.5px;font-weight:800"
												>{p.off}</span
											>{/if}</a
									>
									<div style="flex:1;min-width:0">
										<div
											style="display:flex;align-items:center;gap:2px 6px;min-width:0;flex-wrap:{nar
												? 'wrap'
												: 'nowrap'}"
										>
											<a
												href={p.href}
												class:dos={nar}
												style="font-size:14.5px;font-weight:600;line-height:1.25;overflow:hidden;color:#1E2B22;text-decoration:none;{nar
													? ''
													: 'white-space:nowrap;text-overflow:ellipsis'}">{p.n}</a
											>{#if p.tag}<span
													style="flex:none;padding:1px 5px;border-radius:4px;background:#E7EDF5;color:#2E4C7A;font-size:10px;font-weight:800;text-transform:uppercase"
													>{p.tag}</span
												>{/if}
										</div>
										<div
											style="font-size:12px;color:#6B7A6E;white-space:nowrap;overflow:hidden;text-overflow:ellipsis"
										>
											{#if tienda.cobros.transferencia > 0}<b style="color:var(--acento, #2F5D50)"
													>{p.transf} por transferencia</b
												>{/if}
										</div>
									</div>
									<div style="text-align:right;flex:none;min-width:{nar ? 0 : 74}px">
										<b style="display:block;font-size:16px;letter-spacing:-.02em">{p.price}</b
										>{#if p.old}<s style="font-size:11.5px;color:#8A968C">{p.old}</s>{/if}
									</div>
									{#if q}<span
											style="flex:none;display:flex;align-items:center;gap:2px;padding:3px;border-radius:9px;background:var(--acento, #2F5D50);color:#fff"
											><button
												type="button"
												aria-label="Uno menos"
												onclick={() => carrito.fijar(p.id, q - 1)}
												style="width:28px;height:28px;display:grid;place-items:center;border:0;padding:0;background:transparent;color:#fff;cursor:pointer"
												><Icon icon={I.phMinusBold} size={12} /></button
											><b style="min-width:16px;text-align:center;font-size:13.5px">{q}</b><button
												type="button"
												aria-label="Uno más"
												disabled={q >= p.stock}
												onclick={() => sumar(p.id)}
												style="width:28px;height:28px;display:grid;place-items:center;border:0;padding:0;background:transparent;color:#fff;cursor:pointer"
												><Icon icon={I.phPlusBold} size={12} /></button
											></span
										>{:else}<button
											type="button"
											aria-label="Agregar {p.n}"
											disabled={p.agotado || tienda.vitrina}
											onclick={() => sumar(p.id)}
											style="flex:none;width:40px;height:40px;padding:0;border-radius:9px;border:1.5px solid var(--acento, #2F5D50);background:transparent;color:var(--acento, #2F5D50);display:grid;place-items:center;cursor:pointer"
											><Icon icon={I.phPlusBold} size={15} /></button
										>{/if}
								</div>
							{/each}
						</div>
					{/if}
					{#if listado.length > mostrar}
						<button
							type="button"
							onclick={() => (mostrar += 24)}
							style="margin-top:16px;width:100%;padding:12px;border:0;text-align:center;border-radius:12px;background:#fff;font:700 14px 'Schibsted Grotesk';color:var(--acento, #2F5D50);cursor:pointer"
							>Ver más · ves {catalog.length} de {listado.length}</button
						>
					{/if}
				</div>
			</section>
		{/if}

		<!-- ============ PRODUCTO ============ -->
		{#if pagina === 'producto' && producto && sel}
			<section style="padding:16px clamp(12px,3cqw,32px) 0">
				<div style="font-size:12.5px;color:#6B7A6E">
					<a href={base || '/'} style="color:inherit;text-decoration:none">Inicio</a> / {#if producto.categoria}<a
							href={enlaceCat(producto.categoria)}
							style="color:inherit;text-decoration:none">{producto.categoria}</a
						> /
					{/if}<span style="color:#1E2B22">&#32;{producto.nombre}</span>
				</div>
				<div
					style="margin-top:12px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,360px),1fr));gap:clamp(16px,3cqw,40px);align-items:start"
				>
					<div style="display:flex;flex-direction:column;gap:8px;min-width:0">
						<div
							style="position:relative;aspect-ratio:1;border-radius:16px;overflow:hidden;background:#fff"
						>
							{#if sel.foto}<img
									src={img(sel.foto, 1000, 1000)}
									alt={producto.nombre}
									style="width:100%;height:100%;object-fit:cover"
								/>{/if}{#if tachado}<span
									style="position:absolute;top:12px;left:12px;padding:5px 10px;border-radius:6px;background:#C2412D;color:#fff;font-size:13px;font-weight:800"
									>−{Math.round((1 - sel.precio / tachado) * 100)}%</span
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
										style="aspect-ratio:1;padding:0;border:0;border-radius:10px;overflow:hidden;background:#fff;cursor:pointer;box-shadow:0 0 0 {i ===
										sel.fotoIdx
											? '2px'
											: '0px'} var(--acento, #2F5D50)"
										><img
											src={img(f, 200, 200)}
											alt=""
											style="display:block;width:100%;height:100%;object-fit:cover"
										/></button
									>{/each}
							</div>{/if}
					</div>
					<div
						style="position:sticky;top:120px;display:flex;flex-direction:column;gap:14px;min-width:0"
					>
						<div>
							{#if producto.categoria || producto.etiqueta}<div
									style="display:flex;gap:6px;flex-wrap:wrap"
								>
									{#if producto.categoria}<a
											href={enlaceCat(producto.categoria)}
											style="display:flex;align-items:center;gap:4px;padding:4px 9px;border-radius:6px;background:#DDE8DF;color:var(--acento, #2F5D50);font-size:12px;font-weight:800;text-transform:uppercase;text-decoration:none"
											>{producto.categoria}</a
										>{/if}{#if producto.etiqueta}<span
											style="padding:4px 9px;border-radius:6px;background:#F2F4EE;font-size:12px;font-weight:700"
											>{producto.etiqueta}</span
										>{/if}
								</div>{/if}
							<h1
								style="margin:10px 0 0;font-weight:800;font-size:clamp(32px,4cqw,48px);line-height:1;letter-spacing:-.045em;overflow-wrap:break-word"
							>
								{producto.nombre}
							</h1>
							{#if resenasProd.length}<div
									style="margin-top:8px;display:flex;align-items:center;gap:8px;font-size:13.5px"
								>
									<span style="display:flex;color:#D9A400"
										>{#each estrellas(promedioProd) as e, i (i)}<Icon
												icon={e}
												size={14}
											/>{/each}</span
									><b>{promedioProd.toLocaleString('es-AR')}</b><span style="color:#6B7A6E"
										>{resenasProd.length}
										{resenasProd.length === 1 ? 'reseña' : 'reseñas'}</span
									>
								</div>{/if}
						</div>
						{#each sel.opciones as o (o.nombre)}{@const swatch = o.valores.every(
								(x) => x.color || x.foto
							)}
							<div>
								<div
									style="font-size:13.5px;color:{sel.aviso && !o.elegido ? '#C2412D' : '#1E2B22'}"
								>
									<b>{o.nombre}</b>{#if o.elegido}: {o.elegido}{/if}
								</div>
								{#if swatch}<div style="margin-top:8px;display:flex;gap:10px;flex-wrap:wrap">
										{#each o.valores as x (x.valor)}<button
												type="button"
												aria-label="{o.nombre} {x.valor}"
												aria-pressed={x.elegido}
												disabled={!x.disponible && !x.elegido}
												onclick={() => sel.elegir(o.nombre, x.valor)}
												style="width:40px;height:40px;padding:0;border:0;border-radius:50%;overflow:hidden;background:#fff;cursor:pointer;opacity:{x.disponible
													? 1
													: 0.4};box-shadow:inset 0 0 0 1px rgba(30,43,34,.12),0 0 0 {x.elegido
													? '2px'
													: '0px'} #EEF1E8,0 0 0 {x.elegido ? '4px' : '0px'} var(--acento, #2F5D50)"
												>{#if x.color}<span
														style="display:block;width:100%;height:100%;background:{x.color}"
													></span>{:else}<img
														src={img(x.foto, 80, 80)}
														alt=""
														style="display:block;width:100%;height:100%;object-fit:cover"
													/>{/if}</button
											>{/each}
									</div>{:else}<div
										style="margin-top:8px;display:grid;grid-template-columns:repeat({o.valores.every(
											(x) => x.valor.length <= 7
										)
											? 4
											: 2},minmax(0,1fr));gap:6px"
									>
										{#each o.valores as x (x.valor)}{@const pv =
												sel.opciones.length === 1
													? sel.variantes.find((v) => v.valores[o.nombre] === x.valor)
													: null}<button
												type="button"
												aria-pressed={x.elegido}
												disabled={!x.disponible && !x.elegido}
												onclick={() => sel.elegir(o.nombre, x.valor)}
												style="min-height:48px;padding:10px 6px;border-radius:10px;border:1.5px solid {x.elegido
													? 'var(--acento, #2F5D50)'
													: 'rgba(30,43,34,.12)'};background:{x.elegido
													? '#DDE8DF'
													: '#fff'};text-align:center;font:400 13.5px 'Schibsted Grotesk';color:{x.disponible
													? '#1E2B22'
													: '#8A968C'};text-decoration:{x.disponible
													? 'none'
													: 'line-through'};cursor:{x.disponible ? 'pointer' : 'default'}"
												><b style="display:block;font-size:14.5px">{x.valor}</b>{#if pv}<span
														style="font-size:11.5px;color:#56645A"
														>{fmt(pv.precio ?? producto.precio)}</span
													>{/if}</button
											>{/each}
									</div>{/if}
							</div>{/each}
						<div style="padding:16px;border-radius:14px;background:#fff">
							<div style="display:flex;align-items:baseline;gap:10px;flex-wrap:wrap">
								<b style="font-size:34px;letter-spacing:-.03em"
									>{sel.desde ? 'Desde ' : ''}{fmt(sel.precio)}</b
								>{#if tachado}<s style="color:#8A968C">{fmt(tachado)}</s>{/if}
							</div>
							{#if tienda.cobros.cuotas > 1 || tienda.cobros.transferencia > 0}<div
									style="margin-top:4px;font-size:13.5px"
								>
									{#if tienda.cobros.cuotas > 1}{tienda.cobros.cuotas} cuotas sin interés de {fmt(
											sel.precio / tienda.cobros.cuotas
										)}{/if}{#if tienda.cobros.transferencia > 0}{#if tienda.cobros.cuotas > 1}
											·
										{/if}<b style="color:var(--acento, #2F5D50)"
											>{fmt(sel.precio * (1 - tienda.cobros.transferencia / 100))} por transferencia</b
										>{/if}
								</div>{/if}
							{#if escalas.length}<div
									style="margin-top:12px;border-radius:10px;overflow:hidden;border:1.5px solid rgba(30,43,34,.1)"
								>
									<div
										style="padding:8px 12px;background:#F2F4EE;font-size:12px;font-weight:800;letter-spacing:.04em"
									>
										PRECIO POR MAYOR
									</div>
									{#each escalas as t (t.q)}<div
											style="display:flex;justify-content:space-between;gap:8px;padding:9px 12px;border-top:1px solid rgba(30,43,34,.08);font-size:13.5px;background:{t.on
												? '#F7F9F4'
												: 'transparent'}"
										>
											<span>{t.q}</span><b>{t.p}</b><span
												style="color:var(--acento, #2F5D50);font-weight:700">{t.s}</span
											>
										</div>{/each}
								</div>{/if}
							{#if sel.stock > 0 && sel.stock <= 3 && (!sel.conOpciones || sel.variante)}<div
									style="margin-top:8px;font-size:13px;font-weight:700;color:#C2412D"
								>
									{sel.stock === 1 ? 'Queda 1' : `Quedan ${sel.stock}`}{#if sel.variante}
										en {sel.texto.toLowerCase()}{/if}
								</div>{/if}
							<div style="margin-top:12px;display:flex;gap:8px">
								{#if sel.stock > 0}<span
										style="display:flex;align-items:center;gap:2px;padding:3px;border-radius:10px;border:1.5px solid rgba(30,43,34,.15);flex:none"
										><button
											type="button"
											aria-label="Uno menos"
											disabled={cantidad <= 1}
											onclick={() => (cantidad = Math.max(1, cantidad - 1))}
											style="width:38px;height:38px;display:grid;place-items:center;border:0;padding:0;background:transparent;color:#1E2B22;cursor:pointer"
											><Icon icon={I.phMinusBold} size={14} /></button
										><b style="width:22px;text-align:center">{cantidad}</b><button
											type="button"
											aria-label="Uno más"
											disabled={cantidad >= sel.stock}
											onclick={() => (cantidad = Math.min(sel.stock, cantidad + 1))}
											style="width:38px;height:38px;display:grid;place-items:center;border:0;padding:0;background:transparent;color:#1E2B22;cursor:pointer"
											><Icon icon={I.phPlusBold} size={14} /></button
										></span
									>{/if}<button
									type="button"
									onclick={agregar}
									disabled={sel.stock === 0 || tienda.vitrina}
									style="flex:1;padding:14px;border:0;border-radius:10px;background:var(--acento, #2F5D50);color:#fff;font:800 15px 'Schibsted Grotesk';cursor:pointer"
									>{sel.boton(`Agregar · ${fmt(sel.precio * cantidad)}`)}</button
								>
							</div>
							{#if whatsapp}<a
									href={wa(
										`¡Hola! Quiero pedir ${producto.nombre}${sel.variante ? ` (${sel.texto})` : ''} (${fmt(sel.precio)}).`
									)}
									target="_blank"
									rel="noopener"
									style="margin-top:8px;display:flex;justify-content:center;align-items:center;gap:8px;padding:12px;border-radius:10px;background:#E6F5EA;color:#1F6F3A;font-size:14px;font-weight:700;text-decoration:none"
									><Icon icon={I.phWhatsappLogoFill} size={18} />Pedir por WhatsApp</a
								>{/if}
							<LanzamientoFicha {tienda} {producto} {sel} {tema} />
							<CompartirProducto {tienda} {producto} {tema} />
						</div>
						{#if envia(tienda.envios)}<div
								style="padding:14px;border-radius:14px;background:#fff;display:flex;flex-direction:column;gap:6px;font-size:13.5px"
							>
								<label
									for="cp-ficha"
									style="display:flex;align-items:center;gap:8px;font-weight:700"
									><Icon icon={I.phMapPinBold} size={16} />Calculá el envío</label
								>
								<div style="display:flex;gap:6px;padding-top:4px">
									<input
										id="cp-ficha"
										bind:value={cp}
										inputmode="numeric"
										maxlength="4"
										placeholder="Código postal"
										onkeydown={(e) => e.key === 'Enter' && calcular()}
										style="flex:1;min-width:0;padding:10px 12px;border-radius:9px;border:1.5px solid rgba(30,43,34,.14);background:transparent;font:400 14px 'Schibsted Grotesk';color:#1E2B22"
									/><button
										type="button"
										onclick={calcular}
										style="padding:10px 14px;border:0;border-radius:9px;background:#1E2B22;color:#EEF1E8;font:700 13.5px 'Schibsted Grotesk';cursor:pointer"
										>Calcular</button
									>
								</div>
								{#if cotizado}{#each cotizado as o (o.id)}<span
											style="display:flex;justify-content:space-between;gap:10px;padding:8px 0 0"
											><span>{o.nombre} · {o.detalle}</span><b
												style="white-space:nowrap;color:{o.precio
													? '#1E2B22'
													: 'var(--acento, #2F5D50)'}">{o.precio ? fmt(o.precio) : 'Gratis'}</b
											></span
										>{/each}{/if}
							</div>{/if}
						<div style="display:flex;flex-direction:column;gap:6px">
							{#each acc as a (a.t)}<div
									style="padding:12px 14px;border-radius:14px;background:#fff;font-size:14px"
								>
									<button
										type="button"
										onclick={() => (abierto = a.open ? '' : a.t)}
										aria-expanded={a.open}
										style="width:100%;display:flex;justify-content:space-between;align-items:center;padding:0;border:0;background:transparent;font:700 14px 'Schibsted Grotesk';color:#1E2B22;cursor:pointer;text-align:left"
										>{a.t}<Icon icon={a.i} size={14} /></button
									>{#if a.open}<p
											style="margin:8px 0 0;font-size:13.5px;line-height:1.55;color:#56645A"
										>
											{a.d}
										</p>{/if}
								</div>{/each}
						</div>
					</div>
				</div>
			</section>
			<section style="padding:clamp(24px,4cqw,40px) clamp(12px,3cqw,32px) 0">
				<div
					style="max-width:820px;padding:clamp(16px,2.6cqw,24px);border-radius:16px;background:#fff"
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
			{#if juntos.length}
				<section style="padding:clamp(32px,5cqw,56px) clamp(12px,3cqw,32px)">
					<h2
						style="margin:0;font-weight:800;font-size:clamp(24px,3cqw,34px);letter-spacing:-.04em"
					>
						Combinalo con
					</h2>
					<div
						style="margin-top:14px;padding:14px;border-radius:14px;background:#fff;display:flex;align-items:center;gap:10px;flex-wrap:wrap"
					>
						{#each juntos as t, i (t.id)}<a
								href="{base}/producto/{t.slug ?? t.id}"
								style="display:flex;align-items:center;gap:10px;min-width:0;color:#1E2B22;text-decoration:none"
								>{#if t.foto}<img
										src={img(t.foto, 140, 140)}
										alt=""
										style="width:64px;height:64px;border-radius:10px;object-fit:cover;flex:none"
									/>{/if}<span style="min-width:0"
									><span style="display:block;font-size:13.5px;font-weight:600"
										>{t.nombre}{#if i === 0 && sel.variante}&nbsp;· {sel.opciones
												.map((o) => o.elegido)
												.join(' · ')}{/if}</span
									><b style="font-size:14px">{fmt(i === 0 ? sel.precio : t.precio)}</b></span
								>{#if i < juntos.length - 1}<Icon
										icon={I.phPlusBold}
										size={16}
										style="color:#8A968C;margin:0 4px;flex:none"
									/>{/if}</a
							>{/each}<span style="flex:1"></span>
						<div style="text-align:right">
							<div style="font-size:12.5px;color:#56645A">Los {juntos.length} juntos</div>
							<b style="font-size:20px"
								>{fmt(juntos.reduce((s, p, i) => s + (i === 0 ? sel.precio : p.precio), 0))}</b
							>
						</div>
						<button
							type="button"
							onclick={sumarJuntos}
							disabled={tienda.vitrina}
							style="padding:12px 16px;border:0;border-radius:10px;background:#1E2B22;color:#EEF1E8;font:700 14px 'Schibsted Grotesk';cursor:pointer"
							>Sumar los {juntos.length}</button
						>
					</div>
					{#if related.length > 2}
						<div
							style="margin-top:10px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(46%,190px),1fr));gap:8px"
						>
							{#each related.slice(2).map(tarjeta) as p (p.id)}{@render tarjetaGrilla(p)}{/each}
						</div>
					{/if}
				</section>
			{:else}
				<div style="height:clamp(32px,5cqw,56px)"></div>
			{/if}
		{/if}

		<!-- ============ CARRITO ============ -->
		{#if pagina === 'carrito'}
			<section
				style="padding:16px clamp(12px,3cqw,32px) clamp(40px,6cqw,72px);display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,340px),1fr));gap:clamp(14px,3cqw,32px);align-items:start"
			>
				<div style="display:flex;flex-direction:column;gap:10px;min-width:0">
					<h1
						style="margin:0;font-weight:800;font-size:clamp(32px,4.4cqw,52px);letter-spacing:-.045em"
					>
						Tu carrito <span style="font-size:.5em;color:#6B7A6E;font-weight:600"
							>{carrito.unidades} {carrito.unidades === 1 ? 'producto' : 'productos'}</span
						>
					</h1>
					{#if !lineas.length}
						<p style="margin:6px 0 0;font-size:15px;color:#56645A">Todavía no sumaste nada.</p>
						<a
							href="{base}/catalogo"
							style="align-self:flex-start;padding:14px 22px;border-radius:10px;background:var(--acento, #2F5D50);color:#fff;font:800 15px 'Schibsted Grotesk';text-decoration:none"
							>Ver productos</a
						>
					{:else}
						{#if tienda.envios.gratisDesde}
							<div
								style="padding:12px 14px;border-radius:12px;background:#fff;display:flex;flex-direction:column;gap:8px"
							>
								<span style="font-size:13.5px"
									>{#if falta}Sumá <b>{fmt(falta)}</b> y el envío es gratis{:else}<b
											>Tenés envío gratis</b
										>{/if}</span
								><span
									style="display:block;height:6px;border-radius:3px;background:#DDE3D7;overflow:hidden"
									><span
										style="display:block;width:{Math.min(
											100,
											(subtotal / tienda.envios.gratisDesde) * 100
										)}%;height:100%;background:var(--acento, #2F5D50)"
									></span></span
								>
							</div>
						{/if}
						<div style="border-radius:14px;background:#fff;overflow:hidden">
							{#each lineas as c, i (c.clave)}<div
									style="display:flex;align-items:center;gap:12px;padding:10px 12px;border-top:{i
										? '1px solid rgba(30,43,34,.07)'
										: '0'}"
								>
									{#if c.producto.foto}<img
											src={img(c.producto.foto, 120, 120)}
											alt=""
											style="width:56px;height:56px;border-radius:9px;object-fit:cover;flex:none"
										/>{/if}
									<div style="flex:1;min-width:0">
										<a
											href="{base}/producto/{c.producto.slug ?? c.id}"
											style="display:block;font-size:14px;font-weight:600;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;color:#1E2B22;text-decoration:none"
											>{c.producto.nombre}</a
										>
										<div style="font-size:12px;color:#6B7A6E">
											{c.variante ? `${c.variante} · ` : ''}{fmt(c.producto.precio)} c/u ·
											<button
												type="button"
												onclick={() => carrito.fijar(c.clave, 0)}
												style="border:0;padding:0;background:transparent;font:600 12px 'Schibsted Grotesk';color:#6B7A6E;text-decoration:underline;cursor:pointer"
												>Quitar</button
											>
										</div>
									</div>
									<span
										style="display:flex;align-items:center;gap:2px;padding:2px;border-radius:8px;border:1.5px solid rgba(30,43,34,.14);flex:none"
										><button
											type="button"
											aria-label="Uno menos"
											onclick={() => carrito.fijar(c.clave, c.cantidad - 1)}
											style="width:28px;height:28px;display:grid;place-items:center;border:0;padding:0;background:transparent;color:#1E2B22;cursor:pointer"
											><Icon icon={I.phMinusBold} size={12} /></button
										><b style="min-width:16px;text-align:center;font-size:13.5px">{c.cantidad}</b
										><button
											type="button"
											aria-label="Uno más"
											disabled={c.cantidad >= c.producto.stock}
											onclick={() => carrito.fijar(c.clave, c.cantidad + 1)}
											style="width:28px;height:28px;display:grid;place-items:center;border:0;padding:0;background:transparent;color:#1E2B22;cursor:pointer"
											><Icon icon={I.phPlusBold} size={12} /></button
										></span
									>{#if wide}<b style="min-width:70px;text-align:right;font-size:14.5px"
											>{fmt(c.subtotal)}</b
										>{/if}
								</div>{/each}
						</div>
						{#if avisoMayor && porMayor}<div
								style="padding:12px 14px;border-radius:12px;background:#E8C547;display:flex;align-items:center;gap:10px;font-size:13.5px"
							>
								<Icon icon={I.phLightbulbBold} size={18} style="flex:none" /><span
									style="flex:1;min-width:0"
									>{avisoMayor.pre}<b>{avisoMayor.b}</b>{avisoMayor.post}</span
								>{#if !porMayor.aplicado && porMayor.falta.unidades}<b style="white-space:nowrap"
										>Sumar {porMayor.falta.unidades}</b
									>{/if}
							</div>{/if}
						{#each upsell as u (u.id)}<div
								style="padding:12px 14px;border-radius:12px;background:#E8C547;display:flex;align-items:center;gap:10px;font-size:13.5px"
							>
								<Icon icon={I.phLightbulbBold} size={18} style="flex:none" /><span
									style="flex:1;min-width:0">¿Te falta <b>{u.nombre}</b>? {fmt(u.precio)}</span
								><button
									type="button"
									disabled={tienda.vitrina}
									onclick={() => sumar(u.id)}
									style="border:0;padding:0;background:transparent;font:700 13.5px 'Schibsted Grotesk';color:#1E2B22;white-space:nowrap;cursor:pointer"
									>Sumar</button
								>
							</div>{/each}
					{/if}
				</div>
				{#if lineas.length}
					<div
						style="position:sticky;top:120px;padding:18px;border-radius:14px;background:#fff;display:flex;flex-direction:column;gap:10px;font-size:14px"
					>
						<b style="font-size:18px">Resumen</b>
						<span style="display:flex;justify-content:space-between;gap:10px"
							><span>Productos ({carrito.unidades})</span><span>{fmt(subtotal + ahorro)}</span
							></span
						>
						{#if ahorro}<span
								style="display:flex;justify-content:space-between;gap:10px;color:#C2412D"
								><span>Ofertas</span><span>− {fmt(ahorro)}</span></span
							>{/if}
						<span style="display:flex;justify-content:space-between;gap:10px"
							><span>Envío</span><span style="color:#56645A">Se calcula al pagar</span></span
						>
						<div style="height:1px;background:rgba(30,43,34,.1)"></div>
						<span style="display:flex;justify-content:space-between;align-items:baseline"
							><b>Total</b><b style="font-size:26px;letter-spacing:-.02em">{fmt(subtotal)}</b></span
						>
						{#if tienda.cobros.cuotas > 1}<span style="font-size:13px;color:#56645A"
								>{tienda.cobros.cuotas} cuotas sin interés de {fmt(
									subtotal / tienda.cobros.cuotas
								)}</span
							>{/if}
						{#if tienda.cobros.transferencia > 0}<span
								style="font-size:13.5px;color:var(--acento, #2F5D50);font-weight:700"
								>{fmt(subtotal * (1 - tienda.cobros.transferencia / 100))} pagando por transferencia</span
							>{/if}
						<a
							href="{base}/checkout"
							style="padding:15px;border-radius:10px;background:var(--acento, #2F5D50);color:#fff;font:800 15px 'Schibsted Grotesk';text-align:center;text-decoration:none"
							>Ir a pagar</a
						>{#if whatsapp}<a
								href={wa(textoPedido)}
								target="_blank"
								rel="noopener"
								style="display:flex;justify-content:center;align-items:center;gap:8px;padding:12px;border-radius:10px;background:#E6F5EA;color:#1F6F3A;font-size:14px;font-weight:700;text-decoration:none"
								><Icon icon={I.phWhatsappLogoFill} size={18} />Mandar el pedido por WhatsApp</a
							>{/if}
					</div>
				{/if}
			</section>
		{/if}

		<!-- ============ CHECKOUT ============ -->
		{#if pagina === 'checkout'}
			{#if !lineas.length}
				<section style="padding:16px clamp(12px,3cqw,32px) clamp(40px,6cqw,72px)">
					<p style="font-size:15px;color:#56645A">Tu carrito está vacío.</p>
					<a href="{base}/catalogo" style="color:var(--acento, #2F5D50);font-weight:700"
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
					style="padding:16px clamp(12px,3cqw,32px) clamp(40px,6cqw,72px);display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,340px),1fr));gap:clamp(14px,3cqw,32px);align-items:start"
				>
					<input type="hidden" name="carrito" value={JSON.stringify(carrito.lineas)} />
					<input type="hidden" name="envio" value={envio ?? ''} />
					<input type="hidden" name="pago" value={pago} />
					<div style="display:flex;flex-direction:column;gap:10px;min-width:0">
						<div style="display:flex;gap:6px">
							{#each ['1 · Datos', '2 · Entrega', '3 · Pago'] as t, i (t)}<span
									style="flex:1;padding:9px;border-radius:9px;background:{i === 0
										? 'var(--acento, #2F5D50)'
										: '#fff'};color:{i === 0
										? '#fff'
										: '#6B7A6E'};font-size:13px;font-weight:700;text-align:center">{t}</span
								>{/each}
						</div>
						<div
							style="padding:16px;border-radius:14px;background:#fff;display:grid;grid-template-columns:repeat(auto-fit,minmax(180px,1fr));gap:8px"
						>
							<b style="grid-column:1/-1;font-size:16px">Tus datos</b>
							{#each [['email', 'Email', 'email', 'email'], ['nombre', 'Nombre y apellido', 'text', 'name'], ['telefono', 'Teléfono', 'tel', 'tel'], ['dni', 'DNI (opcional)', 'text', 'off'], ...camposFactura(tienda)] as [n, l, tipo, ac] (n)}
								<label
									style="display:block;padding:9px 12px;border-radius:9px;border:1.5px solid rgba(30,43,34,.14)"
									><span style="display:block;font-size:11.5px;color:#6B7A6E">{l}</span><input
										name={n}
										type={tipo}
										autocomplete={ac as 'email'}
										required={!OPCIONALES.has(n)}
										maxlength="120"
										style="width:100%;margin-top:1px;padding:0;border:0;background:transparent;font:400 14.5px 'Schibsted Grotesk';color:#1E2B22;outline:none"
									/></label
								>
							{/each}
						</div>
						<div
							style="padding:16px;border-radius:14px;background:#fff;display:grid;grid-template-columns:minmax(0,130px) minmax(0,1fr);gap:8px"
						>
							<b style="grid-column:1/-1;font-size:16px">¿Dónde lo recibís?</b>
							<label
								style="display:block;padding:9px 12px;border-radius:9px;border:1.5px solid rgba(30,43,34,.14)"
								><span style="display:block;font-size:11.5px;color:#6B7A6E">Código postal</span
								><input
									name="cp"
									bind:value={cp}
									inputmode="numeric"
									maxlength="4"
									style="width:100%;margin-top:1px;padding:0;border:0;background:transparent;font:400 14.5px 'Schibsted Grotesk';color:#1E2B22;outline:none"
								/></label
							>
							<label
								style="display:block;padding:9px 12px;border-radius:9px;border:1.5px solid rgba(30,43,34,.14)"
								><span style="display:block;font-size:11.5px;color:#6B7A6E"
									>Calle, número y piso</span
								><input
									name="direccion"
									maxlength="160"
									autocomplete="street-address"
									required={envio !== 'retiro'}
									style="width:100%;margin-top:1px;padding:0;border:0;background:transparent;font:400 14.5px 'Schibsted Grotesk';color:#1E2B22;outline:none"
								/></label
							>
							<div
								style="grid-column:1/-1;display:flex;flex-direction:column;gap:6px"
								role="radiogroup"
								aria-label="Forma de entrega"
							>
								{#each opciones as s (s.id)}{@const on = envio === s.id}<button
										type="button"
										role="radio"
										aria-checked={on}
										onclick={() => (envio = s.id)}
										style="display:flex;align-items:center;gap:12px;padding:12px;border-radius:10px;border:1.5px solid {on
											? 'var(--acento, #2F5D50)'
											: 'rgba(30,43,34,.12)'};background:{on
											? '#F7F9F4'
											: 'transparent'};text-align:left;cursor:pointer;font-family:inherit;color:#1E2B22"
										><span
											style="width:18px;height:18px;border-radius:50%;border:2px solid var(--acento, #2F5D50);display:grid;place-items:center;flex:none"
											>{#if on}<span
													style="width:8px;height:8px;border-radius:50%;background:var(--acento, #2F5D50)"
												></span>{/if}</span
										><Icon
											icon={iconoEnvio(s.id)}
											size={20}
											style="color:var(--acento, #2F5D50);flex:none"
										/><span style="flex:1;min-width:0"
											><b style="font-size:14px">{s.nombre}</b><span
												style="display:block;font-size:12px;color:#6B7A6E">{s.detalle}</span
											></span
										><b
											style="font-size:14px;white-space:nowrap;color:{s.precio
												? '#1E2B22'
												: 'var(--acento, #2F5D50)'}">{s.precio ? fmt(s.precio) : 'Gratis'}</b
										></button
									>{:else}<span style="font-size:13.5px;color:#6B7A6E"
										>Poné tu código postal para ver cómo te&nbsp;llega.</span
									>{/each}
							</div>
							<EnvioCorreo {tienda} {cp} {envio} {opciones} />
						</div>
						<div
							style="padding:16px;border-radius:14px;background:#fff;display:flex;flex-direction:column;gap:6px"
							role="radiogroup"
							aria-label="Forma de pago"
						>
							<b style="font-size:16px">Pago</b>
							{#each pagos as p (p.id)}{@const on = pago === p.id}<button
									type="button"
									role="radio"
									aria-checked={on}
									onclick={() => (pago = p.id)}
									style="display:flex;align-items:center;gap:12px;padding:12px;border-radius:10px;border:1.5px solid {on
										? 'var(--acento, #2F5D50)'
										: 'rgba(30,43,34,.12)'};background:{on
										? '#F7F9F4'
										: 'transparent'};text-align:left;cursor:pointer;font-family:inherit;color:#1E2B22"
									><span
										style="width:18px;height:18px;border-radius:50%;border:2px solid var(--acento, #2F5D50);display:grid;place-items:center;flex:none"
										>{#if on}<span
												style="width:8px;height:8px;border-radius:50%;background:var(--acento, #2F5D50)"
											></span>{/if}</span
									><Icon icon={p.i} size={20} style="color:{p.c};flex:none" /><span
										style="flex:1;min-width:0"
										><b style="font-size:14px">{p.t}</b><span
											style="display:block;font-size:12px;color:#6B7A6E">{p.d}</span
										></span
									></button
								>{/each}
						</div>
					</div>
					<div
						style="position:sticky;top:120px;padding:18px;border-radius:14px;background:#1E2B22;color:#EEF1E8;display:flex;flex-direction:column;gap:10px;font-size:14px"
					>
						<b style="font-size:18px">Tu pedido</b>
						<div style="display:flex;flex-wrap:wrap;row-gap:6px">
							{#each lineas.slice(0, 8) as c (c.clave)}{#if c.producto.foto}<img
										src={img(c.producto.foto, 120, 120)}
										alt={c.producto.nombre}
										title="{c.cantidad} × {c.producto.nombre}{c.variante ? ` (${c.variante})` : ''}"
										style="width:44px;height:44px;border-radius:50%;object-fit:cover;margin-right:-8px;box-shadow:0 0 0 2px #1E2B22"
									/>{/if}{/each}
						</div>
						<span style="display:flex;justify-content:space-between;gap:10px"
							><span style="color:#B9C6BB">Productos ({carrito.unidades})</span><span
								>{fmt(subtotal)}</span
							></span
						>
						<span style="display:flex;justify-content:space-between;gap:10px"
							><span style="color:#B9C6BB">Envío{elegida ? ` · ${elegida.nombre}` : ''}</span><span
								>{elegida ? (cuenta.envio ? fmt(cuenta.envio) : 'Gratis') : '—'}</span
							></span
						>
						{#if cuenta.descCupon}<span
								style="display:flex;justify-content:space-between;gap:10px;color:#E8C547;font-weight:700"
								><span>Cupón {cupon.aplicado}</span><span>− {fmt(cuenta.descCupon)}</span></span
							>{/if}{#if descuento}<span
								style="display:flex;justify-content:space-between;gap:10px;color:#E8C547;font-weight:700"
								><span>Transferencia −{tienda.cobros.transferencia}%</span><span
									>− {fmt(descuento)}</span
								></span
							>{/if}
						<div style="margin:6px 0 2px">
							<CampoCupon
								{cupon}
								lineas={carrito.lineas}
								tinta="#EEF1E8"
								suave="#B9C6BB"
								borde="rgba(238,241,232,.25)"
								fondo="rgba(238,241,232,.06)"
								acento="#E8C547"
								sobreAcento="#1E2B22"
								radio="9px"
								fuente="'Schibsted Grotesk',sans-serif"
								error="#FFD7CF"
							/>
						</div>
						<span
							style="display:flex;justify-content:space-between;align-items:baseline;padding-top:8px;border-top:1px solid rgba(238,241,232,.14)"
							><b>Total</b><b style="font-size:26px">{fmt(total)}</b></span
						>
						{#if error}<div
								role="alert"
								style="padding:10px 12px;border-radius:10px;background:rgba(194,65,45,.18);color:#FFD7CF;font-size:13.5px;font-weight:600"
							>
								{error}
							</div>{/if}
						<button
							disabled={!envio || !pagos.length || enviando || tienda.vitrina}
							style="padding:15px;border:0;border-radius:10px;background:#E8C547;color:#1E2B22;font:800 15px 'Schibsted Grotesk';cursor:pointer"
							>{tienda.vitrina
								? 'Es una tienda de ejemplo'
								: !pagos.length
									? 'Esta tienda todavía no tiene medios de pago'
									: envio
										? `Confirmar pedido · ${fmt(total)}`
										: 'Elegí cómo te llega'}</button
						>
						<span style="display:flex;align-items:center;gap:8px;font-size:12.5px;color:#B9C6BB"
							><Icon icon={I.phShieldCheckFill} size={15} style="flex:none" />Compra protegida · 10
							días para&nbsp;arrepentirte</span
						>
					</div>
				</form>
			{/if}
		{/if}

		<!-- ============ LISTO (no está en el diseño: con las piezas del checkout) ============ -->
		{#if pagina === 'listo' && pedido}
			<section
				style="padding:clamp(24px,4cqw,48px) clamp(12px,3cqw,32px) clamp(40px,6cqw,72px);max-width:640px;margin:0 auto;display:flex;flex-direction:column;gap:10px"
			>
				<div
					style="padding:clamp(20px,3cqw,32px);border-radius:16px;background:var(--acento, #2F5D50);color:#EEF1E8;display:flex;flex-direction:column;gap:12px"
				>
					<span
						style="width:52px;height:52px;border-radius:50%;background:#E8C547;color:#1E2B22;display:grid;place-items:center"
						><Icon icon={I.phCheckBold} size={24} /></span
					>
					<h1
						style="margin:0;font-weight:800;font-size:clamp(32px,4.4cqw,52px);line-height:1;letter-spacing:-.045em;overflow-wrap:break-word"
					>
						¡Gracias, {pedido.nombre}!
					</h1>
					<p style="margin:0;font-size:15.5px;line-height:1.55;color:#D6E0D4">
						Recibimos tu pedido <b style="color:#fff">#{pedido.numero}</b> por
						<b style="color:#fff">{fmt(pedido.total)}</b>. Te escribimos por mail con
						cada&nbsp;novedad.
					</p>
				</div>
				<SenaPedido {tema} />
				{#if pedido.pago === 'transferencia'}
					<div
						style="padding:16px;border-radius:14px;background:#fff;font-size:14.5px;line-height:1.6"
					>
						<b>Transferí {fmt(pedido.total)}</b>
						{#if pedido.alias}al {pedido.esCbu ? 'CBU/CVU' : 'alias'}
							<b>{pedido.alias}</b>{:else}con los datos que te mandamos por mail{/if} y mandanos el comprobante{#if whatsapp}&nbsp;<a
								href={wa(`¡Hola! Te mando el comprobante del pedido #${pedido.numero}.`)}
								target="_blank"
								rel="noopener"
								style="color:var(--acento, #2F5D50);font-weight:700">por WhatsApp</a
							>{:else}&nbsp;respondiendo el mail del pedido{/if}. Lo preparamos apenas
						se&nbsp;acredite.
					</div>
				{/if}
				<a
					href={base || '/'}
					style="padding:15px;border-radius:10px;background:#1E2B22;color:#EEF1E8;font:800 15px 'Schibsted Grotesk';text-align:center;text-decoration:none"
					>Volver a la tienda</a
				>
				{#if pedido.estado === 'Entregado'}
					<div style="margin-top:clamp(16px,3cqw,32px)">
						<FormResena
							slug={tienda.slug}
							tema={{ ...tema, borde: 'transparent', radio: '14px', radioBoton: '10px' }}
							codigo={pedido.codigo}
							nombre={pedido.nombre}
						/>
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
				style="padding:16px clamp(12px,3cqw,32px) clamp(40px,6cqw,72px);display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,340px),1fr));gap:clamp(14px,3cqw,32px);align-items:start"
			>
				<input type="hidden" name="servicioId" value={reserva.servicioId ?? ''} />
				<input type="hidden" name="profesionalId" value={reserva.profesionalId ?? ''} />
				<input type="hidden" name="inicio" value={reserva.inicio ?? ''} />
				<div style="display:flex;flex-direction:column;gap:10px;min-width:0">
					<div style="display:flex;gap:6px">
						{#each PASOS_TURNO as t, i (t)}<span
								style="flex:1;min-width:0;padding:9px 4px;border-radius:9px;background:{i <=
								reserva.paso
									? 'var(--acento, #2F5D50)'
									: '#fff'};color:{i <= reserva.paso
									? '#fff'
									: '#6B7A6E'};font-size:12px;font-weight:700;text-align:center;overflow:hidden;text-overflow:ellipsis;white-space:nowrap"
								>{t}</span
							>{/each}
					</div>
					<h1
						style="margin:6px 0;font-weight:800;font-size:clamp(32px,4.4cqw,52px);line-height:1;letter-spacing:-.045em;overflow-wrap:break-word"
					>
						Reservá tu turno
					</h1>
					<div style="padding:16px;border-radius:14px;background:#fff;min-width:0">
						<b style="font-size:16px">1 · Servicio</b>
						<div
							style="margin-top:10px;display:flex;flex-direction:column;gap:6px"
							role="radiogroup"
							aria-label="Servicio"
						>
							{#each tienda.turnos.servicios as s (s.id)}{@const on =
									reserva.servicioId === s.id}<button
									type="button"
									role="radio"
									aria-checked={on}
									onclick={() => reserva.elegirServicio(s.id)}
									style="display:flex;align-items:center;gap:12px;width:100%;padding:12px;border-radius:10px;border:1.5px solid {on
										? 'var(--acento, #2F5D50)'
										: 'rgba(30,43,34,.12)'};background:{on
										? '#F7F9F4'
										: 'transparent'};text-align:left;cursor:pointer;font-family:inherit;color:#1E2B22"
									><span
										style="width:18px;height:18px;border-radius:50%;border:2px solid var(--acento, #2F5D50);display:grid;place-items:center;flex:none"
										>{#if on}<span
												style="width:8px;height:8px;border-radius:50%;background:var(--acento, #2F5D50)"
											></span>{/if}</span
									><span style="flex:1;min-width:0"
										><span style="display:block;font-size:14.5px;font-weight:700">{s.nombre}</span
										><span style="display:block;font-size:12.5px;color:#6B7A6E;"
											>{duracion(s.duracion)}</span
										></span
									><b style="white-space:nowrap">{fmt(s.precio)}</b></button
								>{/each}
						</div>
					</div>
					<div style="padding:16px;border-radius:14px;background:#fff;min-width:0">
						<b style="font-size:16px">2 · Profesional</b>
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
									style="flex:none;width:110px;padding:10px;border-radius:10px;border:1.5px solid {on
										? 'var(--acento, #2F5D50)'
										: 'rgba(30,43,34,.12)'};background:{on
										? '#F7F9F4'
										: 'transparent'};text-align:center;cursor:pointer;font-family:inherit;color:#1E2B22"
									>{#if p?.foto}<img
											src={img(p.foto, 120, 120)}
											alt=""
											style="width:56px;height:56px;border-radius:50%;object-fit:cover;display:block;margin:0 auto"
										/>{:else}<span
											style="width:56px;height:56px;border-radius:50%;background:#DDE8DF;color:var(--acento, #2F5D50);display:grid;place-items:center;margin:0 auto"
											><Icon icon={p ? I.phUserFill : I.phShuffleBold} size={22} /></span
										>{/if}<span
										style="display:block;margin-top:8px;font-size:13.5px;font-weight:700;overflow:hidden;text-overflow:ellipsis;white-space:nowrap"
										>{p?.nombre ?? 'Cualquiera'}</span
									><span
										style="display:block;font-size:11.5px;color:#6B7A6E;overflow:hidden;text-overflow:ellipsis;white-space:nowrap"
										>{p ? p.rol.split(' · ')[0] : 'El primero libre'}</span
									></button
								>{/each}
						</div>
					</div>
					<div style="padding:16px;border-radius:14px;background:#fff;min-width:0">
						<div style="display:flex;justify-content:space-between;align-items:center;gap:10px">
							<b style="font-size:16px">3 · Día y horario</b>
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
									style="padding:9px 0;text-align:center;cursor:pointer;font-family:inherit;min-width:0;border-radius:9px;border:1.5px solid {on
										? 'var(--acento, #2F5D50)'
										: 'rgba(30,43,34,.12)'};background:{on
										? 'var(--acento, #2F5D50)'
										: '#fff'};color:{on ? '#fff' : '#1E2B22'}"
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
								<p style="margin:0;font-size:14px;color:#6B7A6E;">
									Elegí un servicio para ver los horarios.
								</p>
							{:else if reserva.horarios && !reserva.grupos.length}
								<p style="margin:0;font-size:14px;color:#6B7A6E;">
									No quedan horarios este día. Probá con otro día u otro&nbsp;profesional.
								</p>
							{/if}
							{#each reserva.grupos as g (g.k)}<div>
									<div
										style="font-size:12px;font-weight:700;letter-spacing:.08em;text-transform:uppercase;margin-bottom:7px;color:#6B7A6E;"
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
													: 'line-through'};border-radius:9px;border:1.5px solid {on
													? 'var(--acento, #2F5D50)'
													: 'rgba(30,43,34,.12)'};background:{on
													? 'var(--acento, #2F5D50)'
													: '#fff'};color:{on ? '#fff' : '#1E2B22'}">{h.hora}</button
											>{/each}
									</div>
								</div>{/each}
						</div>
					</div>
					{#if reserva.listo}
						<div style="padding:16px;border-radius:14px;background:#fff;min-width:0">
							<b style="font-size:16px">4 · Tus datos</b>
							<div
								style="margin-top:10px;display:grid;grid-template-columns:repeat(auto-fit,minmax(180px,1fr));gap:8px"
							>
								{#each [['nombre', 'Nombre', 'text', 'name'], ['telefono', 'WhatsApp', 'tel', 'tel'], ['email', 'Email', 'email', 'email']] as [n, l, tipo, ac] (n)}
									<label
										style="display:block;padding:9px 12px;border-radius:9px;border:1.5px solid rgba(30,43,34,.14)"
										><span style="display:block;font-size:11.5px;color:#6B7A6E">{l}</span><input
											name={n}
											type={tipo}
											autocomplete={ac as 'email'}
											required
											maxlength="120"
											style="width:100%;margin-top:1px;padding:0;border:0;background:transparent;font:400 14.5px 'Schibsted Grotesk';color:#1E2B22;outline:none"
										/></label
									>
								{/each}
							</div>
						</div>
						<div style="padding:16px;border-radius:14px;background:#fff;min-width:0">
							<b style="font-size:16px">Pago</b>
							<div style="margin-top:10px;display:flex;flex-direction:column;gap:6px">
								<div
									style="display:flex;align-items:center;gap:12px;width:100%;padding:12px;border-radius:10px;border:1.5px solid var(--acento, #2F5D50);background:#F7F9F4;text-align:left;cursor:pointer;font-family:inherit;color:#1E2B22"
								>
									<span
										style="width:18px;height:18px;border-radius:50%;border:2px solid var(--acento, #2F5D50);display:grid;place-items:center;flex:none"
										><span
											style="width:8px;height:8px;border-radius:50%;background:var(--acento, #2F5D50)"
										></span></span
									>{#if reserva.sena}<Icon
											icon={I.siMercadopago}
											size={22}
											style="color:#00B1EA;flex:none"
										/><span style="flex:1;min-width:0"
											><span style="display:block;font-size:14.5px;font-weight:700"
												>Mercado Pago</span
											><span style="display:block;font-size:12.5px;color:#6B7A6E;"
												>Seña de {fmt(reserva.sena)} para confirmar · el resto, en el local</span
											></span
										>{:else}<Icon icon={I.phStorefrontFill} size={22} style="flex:none" /><span
											style="flex:1;min-width:0"
											><span style="display:block;font-size:14.5px;font-weight:700"
												>Pagás en el local</span
											><span style="display:block;font-size:12.5px;color:#6B7A6E;"
												>No hace falta seña: el turno queda confirmado</span
											></span
										>{/if}
								</div>
							</div>
						</div>
					{/if}
				</div>
				<div
					style="position:sticky;top:120px;padding:18px;border-radius:14px;background:#1E2B22;color:#EEF1E8;font-size:14px;min-width:0"
				>
					<b style="font-size:18px">Tu reserva</b>
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
							<div style="font-size:13px;color:#B9C6BB;">
								{reserva.cuando ?? 'Elegí el día y el horario'}
							</div>
						</div>
					</div>
					{#if reserva.servicio}
						<div
							style="margin-top:16px;padding-top:14px;border-top:1px solid rgba(238,241,232,.14);display:flex;flex-direction:column;gap:8px;font-size:14px"
						>
							<span style="display:flex;justify-content:space-between;gap:10px"
								><span>{reserva.servicio.nombre}</span><span style="white-space:nowrap"
									>{fmt(reserva.servicio.precio)}</span
								></span
							><span style="display:flex;justify-content:space-between;color:#B9C6BB;"
								><span>Duración total</span><span>{reserva.duracion}</span></span
							><span
								style="display:flex;justify-content:space-between;align-items:baseline;margin-top:6px"
								><b>Total</b><b style="font-size:26px">{fmt(reserva.servicio.precio)}</b></span
							>
						</div>
					{/if}
					<p style="margin:14px 0 0;font-size:13px;line-height:1.5;color:#B9C6BB;">
						{#if reserva.sena}Seña de {fmt(reserva.sena)} para confirmar. Podés cancelar desde el link
							del turno hasta {anticipacion} antes; la seña te la devuelve la&nbsp;tienda.{:else}No
							pagás nada ahora. Podés cancelar desde el link del turno hasta {anticipacion}&nbsp;antes.{/if}
					</p>
					{#if error}<div
							role="alert"
							style="margin-top:12px;padding:10px 12px;border-radius:10px;background:rgba(194,65,45,.18);color:#FFD7CF;font-size:13.5px;font-weight:600"
						>
							{error}
						</div>{/if}
					<button
						disabled={!reserva.listo || reservando || tienda.vitrina}
						style="margin-top:16px;width:100%;padding:15px;border:0;border-radius:10px;background:#E8C547;color:#1E2B22;font:800 15px 'Schibsted Grotesk';cursor:pointer"
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
				style="padding:clamp(24px,4cqw,48px) clamp(12px,3cqw,32px) clamp(40px,6cqw,72px);max-width:640px;margin:0 auto"
			>
				<span
					style="width:52px;height:52px;border-radius:50%;background:{cancelado
						? '#1E2B22'
						: 'var(--acento, #2F5D50)'};color:{cancelado
						? '#EEF1E8'
						: '#E8C547'};display:grid;place-items:center"
					><Icon icon={cancelado ? I.phCalendarX : I.phCheckBold} size={24} /></span
				>
				<h1
					style="margin:14px 0 0;font-weight:800;font-size:clamp(32px,4.4cqw,52px);line-height:1;letter-spacing:-.045em;overflow-wrap:break-word"
				>
					{cancelado
						? 'Turno cancelado'
						: turno.estado === 'reservado'
							? `Falta la seña, ${turno.nombre}`
							: `¡Listo, ${turno.nombre}!`}
				</h1>
				<p style="margin:12px 0 0;font-size:15.5px;line-height:1.55;color:#56645A">
					{#if cancelado}Liberamos el horario{turno.senaPagada
							? '. La seña te la devuelve la tienda por Mercado Pago'
							: ''}. Si querés, reservá&nbsp;otro.{:else if turno.estado === 'reservado'}Te
						guardamos el horario unos minutos mientras pagás la seña con Mercado&nbsp;Pago.{:else}Tu
						turno está confirmado. Te esperamos en {turno.tienda}.{/if}
				</p>
				{#if aviso === 'pendiente' || aviso === 'error'}<div
						role="status"
						style="margin-top:14px;padding:10px 12px;border-radius:10px;background:rgba(194,65,45,.18);color:#C2412D;font-size:13.5px;font-weight:600"
					>
						{aviso === 'pendiente'
							? 'Mercado Pago todavía no acreditó la seña: te avisamos cuando entre.'
							: 'El pago de la seña no salió. Podés reservar de nuevo.'}
					</div>{/if}
				<div
					style="margin-top:16px;padding:16px;border-radius:14px;background:#fff;font-size:14.5px;line-height:1.6"
				>
					<div
						style="font-size:16px;font-weight:700;text-decoration:{cancelado
							? 'line-through'
							: 'none'}"
					>
						{cuandoLargo(turno.inicio)} con {turno.profesional}
					</div>
					<div style="margin-top:4px;font-size:14px;color:#6B7A6E;">
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
						style="margin-top:12px;padding:10px 12px;border-radius:10px;background:rgba(194,65,45,.18);color:#C2412D;font-size:13.5px;font-weight:600"
					>
						{error}
					</div>{/if}
				<div style="margin-top:22px;display:flex;gap:10px;flex-wrap:wrap;">
					{#if !cancelado}<a
							href="{base}/turno/{turno.codigo}/turno.ics"
							download="turno.ics"
							data-sveltekit-reload
							style="display:flex;align-items:center;gap:8px;padding:15px 20px;border:0;border-radius:10px;background:var(--acento, #2F5D50);color:#fff;font:800 15px 'Schibsted Grotesk';text-align:center;text-decoration:none;cursor:pointer"
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
									style="display:inline-block;padding:15px 20px;border:0;border-radius:10px;background:#1E2B22;color:#EEF1E8;font:800 15px 'Schibsted Grotesk';text-align:center;text-decoration:none;cursor:pointer"
									>Sí, cancelar el turno</button
								>{:else}<button
									type="button"
									onclick={() => (seguro = true)}
									style="display:inline-block;padding:15px 20px;border:0;border-radius:10px;background:#1E2B22;color:#EEF1E8;font:800 15px 'Schibsted Grotesk';text-align:center;text-decoration:none;cursor:pointer"
									>Cancelar turno</button
								>{/if}
						</form>
					{:else if cancelado}<a
							href="{base}/turnos"
							style="padding:15px 20px;border:0;border-radius:10px;background:var(--acento, #2F5D50);color:#fff;font:800 15px 'Schibsted Grotesk';text-align:center;text-decoration:none;cursor:pointer"
							>Reservar otro</a
						>{/if}
					<a
						href={base || '/'}
						style="display:inline-block;padding:15px 20px;border:0;border-radius:10px;background:#1E2B22;color:#EEF1E8;font:800 15px 'Schibsted Grotesk';text-align:center;text-decoration:none;cursor:pointer"
						>Volver a la tienda</a
					>
				</div>
				{#if !cancelado && !turno.cancelable}<p
						style="margin:14px 0 0;font-size:13.5px;color:#6B7A6E;"
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
				style="padding:clamp(14px,2.4cqw,28px) clamp(12px,3cqw,32px) 0;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,340px),1fr));gap:10px"
			>
				<div
					style="padding:clamp(22px,4cqw,48px);border-radius:16px;background:var(--acento, #2F5D50);color:#EEF1E8;display:flex;flex-direction:column;justify-content:flex-end;min-height:min(440px,100cqw);min-width:0"
				>
					{#if tienda.ciudad}<span
							style="font-size:13px;font-weight:700;letter-spacing:.06em;color:#E8C547;text-transform:uppercase"
							>{tienda.ciudad}</span
						>{/if}
					<h1
						style="margin:12px 0 0;font-weight:800;font-size:clamp(38px,5.4cqw,68px);line-height:.95;letter-spacing:-.045em;overflow-wrap:break-word"
					>
						{tienda.portada.titulo}
					</h1>
					<p
						style="margin:14px 0 0;font-size:15.5px;line-height:1.55;color:#D6E0D4;max-width:480px"
					>
						{tienda.nombre}{tienda.ciudad ? `, desde ${tienda.ciudad}` : ''}. {comoLlega(
							tienda.envios
						)}
					</p>
				</div>
				{#if tienda.portada.foto || tienda.fotos[0]}
					<div
						style="border-radius:16px;overflow:hidden;background:#C9D2C2;min-height:min(440px,100cqw);position:relative"
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
				style="padding:10px clamp(12px,3cqw,32px) 0;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(46%,180px),1fr));gap:10px"
			>
				{#if tienda.productos.length}<div style="padding:18px;border-radius:14px;background:#fff">
						<b style="display:block;font-size:34px;letter-spacing:-.04em"
							>{tienda.productos.length}</b
						><span style="font-size:13.5px;color:#56645A">productos</span>
					</div>{/if}
				{#if tienda.categorias.length}<div style="padding:18px;border-radius:14px;background:#fff">
						<b style="display:block;font-size:34px;letter-spacing:-.04em"
							>{tienda.categorias.length}</b
						><span style="font-size:13.5px;color:#56645A"
							>{tienda.categorias.length === 1 ? 'categoría' : 'categorías'}</span
						>
					</div>{/if}
				{#if tienda.cobros.cuotas > 1}<div style="padding:18px;border-radius:14px;background:#fff">
						<b style="display:block;font-size:34px;letter-spacing:-.04em">{tienda.cobros.cuotas}</b
						><span style="font-size:13.5px;color:#56645A">cuotas sin interés</span>
					</div>{/if}
				{#if tienda.promedio}<div style="padding:18px;border-radius:14px;background:#fff">
						<b style="display:block;font-size:34px;letter-spacing:-.04em"
							>{tienda.promedio.toLocaleString('es-AR')}</b
						><span style="font-size:13.5px;color:#56645A">de promedio en reseñas</span>
					</div>{/if}
			</section>
			{#if tienda.direccion || whatsapp || tienda.mayorista}
				<section
					style="padding:clamp(32px,5cqw,56px) clamp(12px,3cqw,32px);display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,300px),1fr));gap:10px"
				>
					{#if tienda.direccion || whatsapp}<div
							style="padding:18px;border-radius:14px;background:#fff;display:flex;flex-direction:column;gap:8px"
						>
							<b style="font-size:18px">{tienda.envios.retiro ? 'Vení al local' : 'Hablemos'}</b>
							{#if tienda.direccion}<a
									href="https://www.google.com/maps/search/?api=1&query={encodeURIComponent(
										[tienda.direccion, tienda.ciudad].filter(Boolean).join(', ')
									)}"
									target="_blank"
									rel="noopener"
									style="display:flex;gap:8px;font-size:14px;color:#1E2B22;text-decoration:none"
									><Icon
										icon={I.phMapPinBold}
										size={17}
										style="color:var(--acento, #2F5D50);flex:none"
									/>{tienda.direccion}</a
								>{/if}
							{#if whatsapp}<a
									href={wa('¡Hola!')}
									target="_blank"
									rel="noopener"
									style="margin-top:4px;display:flex;justify-content:center;align-items:center;gap:8px;padding:11px;border-radius:10px;background:#E6F5EA;color:#1F6F3A;font-size:14px;font-weight:700;text-decoration:none"
									><Icon icon={I.phWhatsappLogoFill} size={17} />{tienda.whatsapp}</a
								>{/if}
						</div>{/if}
					{#if tienda.mayorista}<div
							style="padding:18px;border-radius:14px;background:#E8C547;display:flex;flex-direction:column;gap:8px"
						>
							<b style="font-size:18px">¿Tenés un comercio?</b><span
								style="font-size:14px;line-height:1.5"
								>Precios mayoristas desde {minimoMayor}: se aplican solos en el&nbsp;carrito.</span
							><a
								href={whatsapp ? wa('¡Hola! Quiero comprar por mayor.') : `${base}/catalogo`}
								target={whatsapp ? '_blank' : undefined}
								rel={whatsapp ? 'noopener' : undefined}
								style="margin-top:auto;padding:11px;border-radius:10px;background:#1E2B22;color:#EEF1E8;font-size:14px;font-weight:700;text-align:center;text-decoration:none"
								>{whatsapp ? 'Pedir lista mayorista' : 'Ver productos'}</a
							>
						</div>{/if}
				</section>
			{:else}
				<div style="height:clamp(32px,5cqw,56px)"></div>
			{/if}
		{/if}

		<!-- footer -->
		<footer
			style="padding:clamp(28px,4cqw,48px) clamp(12px,3cqw,32px) 22px;background:#1E2B22;color:#EEF1E8"
		>
			<div
				style="display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,170px),1fr));gap:24px"
			>
				<div style="min-width:0">
					<div style="font-weight:800;font-size:24px;letter-spacing:-.03em">{tienda.nombre}</div>
					<p style="margin:8px 0 0;font-size:13.5px;color:#B9C6BB;line-height:1.5">
						{#if tienda.cobros.cuotas > 1}{tienda.cobros.cuotas} cuotas sin interés{#if tienda.cobros.transferencia > 0}&nbsp;y
								{tienda.cobros.transferencia}% off por transferencia{/if}, en toda la&nbsp;tienda.{:else if tienda.cobros.transferencia > 0}{tienda
								.cobros.transferencia}% off por transferencia, en toda la&nbsp;tienda.{:else}Compra
							protegida, con 10 días para&nbsp;arrepentirte.{/if}
					</p>
					<!-- Newsletter (§50): el bloque del pie del diseño; queda pendiente hasta confirmar por mail. -->
					<form onsubmit={nl.enviar} novalidate style="margin-top:10px;display:flex;gap:6px">
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
							style="--ph:#8FA093;flex:1;min-width:0;padding:11px 12px;border:0;border-radius:9px;background:rgba(238,241,232,.08);font:inherit;font-size:13.5px;color:#EEF1E8;outline:none"
						/>
						<button
							type="submit"
							disabled={nl.enviando}
							style="flex:none;padding:11px 14px;border:0;border-radius:9px;background:#E8C547;color:#1E2B22;font:inherit;font-size:13.5px;font-weight:800;cursor:pointer"
							>OK</button
						>
					</form>
					{#if nl.aviso}<p
							role="status"
							style="margin:8px 0 0;font-size:13px;line-height:1.45;color:{nl.aviso.ok
								? '#B9C6BB'
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
						style="display:flex;flex-direction:column;gap:7px;font-size:13.5px"
					>
						<b style="color:#E8C547;margin-bottom:3px">{f.t}</b>{#each f.l as l (l.t)}<a
								href={l.href}
								style="color:#C9D4CB;text-decoration:none">{l.t}</a
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
				style="margin-top:24px;padding-top:14px;border-top:1px solid rgba(238,241,232,.1);display:flex;justify-content:space-between;gap:12px;flex-wrap:wrap;font-size:12.5px;color:#8FA093"
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
