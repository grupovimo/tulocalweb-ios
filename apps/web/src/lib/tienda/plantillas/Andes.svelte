<script lang="ts">
	import { camposFactura, OPCIONALES } from '#lib/tienda/factura.ts';
	import { enhance } from '$app/forms';
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
	import { textoAviso, textoMinimo } from '../mayorista.ts';
	import { partirTitulo } from '../titulo.ts';
	import { untrack } from 'svelte';
	import { Reserva } from '../reserva.svelte.ts';
	import { accionesDe, AltaPlan, cajaDe, textosSuscripcion } from '../suscripcion.svelte.ts';
	import { CADA, diaCobro, entregasDe, porPeriodo, queIncluye } from '../suscripciones.ts';
	import { cuandoLargo, duracion } from '#lib/turnos/formato.ts';
	import type { OpcionEnvio, ProductoPublico, TiendaPublica } from '../tipos.ts';

	/**
	 * Plantilla Andes (design/Plantilla Andes.dc.html), con los datos de la tienda en lugar de los de
	 * ejemplo. El comparador funciona con los productos ya cargados (se recuerda en el navegador). Lo
	 * que el diseño muestra sin datos atrás (colores, fichas técnicas, combo, cuenta regresiva,
	 * filtros por marca, cupón, factura A, garantía extendida) no se dibuja. El diseño no tiene
	 * newsletter: el checkout ofrece la casilla «Quiero recibir novedades» (§50).
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
		const rebaja = p.anterior && !t.desde;
		return {
			id: p.id,
			n: p.nombre,
			cat: p.categoria,
			price: (t.desde ? 'Desde ' : '') + fmt(t.precio),
			old: rebaja ? fmt(p.anterior!) : null,
			img: img(p.foto, 600, 600),
			tag: t.agotado ? 'Sin stock' : rebaja ? `−${off(p)}%` : p.etiqueta,
			tbg: !t.agotado && rebaja ? 'var(--acento, #2F6BFF)' : '#0E1116',
			cuotas:
				tienda.cobros.cuotas > 1
					? `${tienda.cobros.cuotas} × ${fmt(t.precio / tienda.cobros.cuotas)} sin interés`
					: '',
			transf: fmt(t.precio * (1 - tienda.cobros.transferencia / 100)),
			href: `${base}/producto/${p.slug ?? p.id}`
		};
	};
	const agotado = (p: ProductoPublico) => tarjetaDe(p).agotado;
	const ver = (s: string) => tienda.secciones.includes(s as TiendaPublica['secciones'][number]);
	const whatsapp = $derived(tienda.whatsapp.replace(/\D/g, ''));
	const wa = (texto: string) =>
		`https://wa.me/${whatsapp.startsWith('54') ? whatsapp : '54' + whatsapp}?text=${encodeURIComponent(texto)}`;

	// Título de la portada: el último tramo va con el degradé azul, como en el diseño.
	const titulo = $derived(partirTitulo(tienda.portada.titulo));
	// El producto de la portada: el más vendido con stock.
	const destacado = $derived(
		tienda.masVendidos.map((id) => porId.get(id)).find((p) => p && !agotado(p)) ??
			tienda.productos.find((p) => !agotado(p)) ??
			null
	);

	const perks = $derived([
		...(tienda.envios.moto
			? [{ i: I.phMopedFill, t: 'Envío en moto', d: resumenMoto(tienda.envios) }]
			: tienda.envios.gratisDesde
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
			: { i: I.phArrowUUpLeft, t: 'Arrepentimiento', d: 'Tenés 10 días para devolverlo' }
	]);
	const USOS = [
		['#0E1116', '#fff'],
		['var(--acento, #2F6BFF)', '#fff'],
		['#fff', '#0E1116'],
		['#fff', '#0E1116'],
		['#E3E8EF', '#0E1116']
	];
	const cats = $derived(
		tienda.categorias.slice(0, 5).map((c, i) => ({
			...c,
			bg: USOS[i][0],
			fg: USOS[i][1]
		}))
	);
	const best = $derived(
		tienda.masVendidos.slice(0, 4).flatMap((id) => (porId.get(id) ? [tarjeta(porId.get(id)!)] : []))
	);

	// Comparador: hasta 3 productos de la tienda, recordados en el navegador.
	let comparar = $state<string[]>([]);
	const claveCmp = $derived(`andes-comparar:${tienda.slug}`);
	$effect(() => {
		try {
			const g = JSON.parse(localStorage.getItem(claveCmp) ?? '[]');
			comparar = Array.isArray(g) ? g.filter((id) => typeof id === 'string').slice(0, 3) : [];
		} catch {
			comparar = [];
		}
	});
	function guardarCmp(l: string[]) {
		comparar = l;
		try {
			localStorage.setItem(claveCmp, JSON.stringify(l));
		} catch {
			// Sin almacenamiento: la comparación dura lo que la página.
		}
	}
	const enCmp = (id: string) => comparar.includes(id);
	const alternar = (id: string) =>
		guardarCmp(enCmp(id) ? comparar.filter((x) => x !== id) : [...comparar, id].slice(-3));
	// El comparador se abre sobre la página actual; al navegar se cierra solo.
	const aqui = $derived(`${pagina}|${producto?.id ?? ''}|${categoria ?? ''}`);
	let abiertoEn = $state<string | null>(null);
	const comparando = $derived(abiertoEn === aqui);
	function abrirCmp(id?: string) {
		if (id && !enCmp(id)) alternar(id);
		abiertoEn = aqui;
		window.scrollTo(0, 0);
	}
	const cmp = $derived.by(() => {
		const ps = comparar.flatMap((id) => (porId.get(id) ? [porId.get(id)!] : []));
		const rank = new Map(tienda.masVendidos.map((id, i) => [id, i]));
		const masVendido =
			ps.length > 1
				? [...ps].sort((a, b) => (rank.get(a.id) ?? 999) - (rank.get(b.id) ?? 999))[0]
				: null;
		const resenas = (p: ProductoPublico) =>
			tienda.resenas.filter((r) => r.producto === p.nombre).length;
		const min = (f: (p: ProductoPublico) => number) => Math.min(...ps.map(f));
		const max = (f: (p: ProductoPublico) => number) => Math.max(...ps.map(f));
		const varios = ps.length > 1;
		const desde = (p: ProductoPublico) => tarjetaDe(p).precio;
		return ps.map((p) => {
			const pick = masVendido?.id === p.id && rank.has(p.id);
			const filas: [string, string, boolean][] = [
				['Precio', tarjeta(p).price, varios && desde(p) === min(desde)],
				...(tienda.cobros.transferencia > 0
					? ([
							[
								'Por transferencia',
								fmt(desde(p) * (1 - tienda.cobros.transferencia / 100)),
								varios && desde(p) === min(desde)
							]
						] as [string, string, boolean][])
					: []),
				['Descuento', off(p) ? `−${off(p)}%` : '—', varios && off(p) > 0 && off(p) === max(off)],
				['Categoría', p.categoria ?? '—', false],
				[
					'Stock',
					agotado(p) ? 'Sin stock' : 'Disponible',
					varios && !agotado(p) && ps.some(agotado)
				],
				['Reseñas', String(resenas(p)), varios && resenas(p) > 0 && resenas(p) === max(resenas)]
			];
			return {
				p,
				n: p.nombre,
				price: tarjeta(p).price,
				cuotas:
					tienda.cobros.cuotas > 1
						? `${tienda.cobros.cuotas} × ${fmt(desde(p) / tienda.cobros.cuotas)}`
						: '',
				img: img(p.foto, 500, 375),
				pick,
				bd: pick ? 'var(--acento, #2F6BFF)' : 'transparent',
				btnBg: pick ? 'var(--acento, #2F6BFF)' : '#0E1116',
				rows: filas.map(([t, v, b]) => ({
					t,
					v,
					best: b,
					c: b ? 'var(--acento, #2F6BFF)' : '#0E1116'
				}))
			};
		});
	});

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
	const quick = $derived(
		[null, ...tienda.categorias.map((c) => c.nombre)].map((c) => ({
			c,
			t: c ?? 'Todo',
			i: c ? I.phTag : I.phSquaresFourFill,
			on: c === categoria
		}))
	);

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
						(p) => p.id !== producto!.id && p.categoria === producto!.categoria && !agotado(p)
					)
					.slice(0, 4)
					.map(tarjeta)
			: []
	);
	const resenasDe = $derived(
		producto ? tienda.resenas.filter((r) => r.producto === producto!.nombre) : []
	);
	const promedioDe = $derived(
		resenasDe.length ? resenasDe.reduce((s, r) => s + r.estrellas, 0) / resenasDe.length : 0
	);
	const esMasVendido = $derived(
		producto ? tienda.masVendidos.slice(0, 4).includes(producto.id) : false
	);
	// Tamaño, color, presentación…: lo elegido, la variante, su precio, stock y foto (seleccion.svelte.ts).
	const sel = $derived(producto ? new Seleccion(producto) : null);
	// El tachado no va con "Desde": no se sabe a qué variante se compara.
	const tachado = $derived(sel && !sel.desde ? sel.anterior : null);
	const plans = $derived(
		sel
			? [
					[(sel.desde ? 'Desde ' : '') + fmt(sel.precio), '1 pago', true],
					...(tienda.cobros.cuotas > 1
						? [
								[
									`${tienda.cobros.cuotas} × ${fmt(sel.precio / tienda.cobros.cuotas)}`,
									'sin interés',
									false
								]
							]
						: []),
					...(tienda.cobros.transferencia > 0
						? [[fmt(sel.precio * (1 - tienda.cobros.transferencia / 100)), 'transferencia', false]]
						: []),
					// Venta por mayor (§49): una caja más, con el mínimo del carrito.
					...(sel.mayorista !== null && tienda.mayorista
						? [[fmt(sel.mayorista), `por mayor desde ${textoMinimo(tienda.mayorista, fmt)}`, false]]
						: [])
				].map(([v, t, on]) => ({
					v: String(v),
					t: String(t),
					bd: on ? 'var(--acento, #2F6BFF)' : '#E3E8EF',
					bg: on ? '#F3F6FF' : '#fff'
				}))
			: []
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
	/** Sumar desde el comparador o el carrito: con opciones, primero se elige en la ficha. */
	function sumar(p: ProductoPublico, alCarrito = false) {
		if (tienda.vitrina) return;
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- ruta de la tienda con su slug
		if (!carrito.sumarProducto(p)) goto(`${base}/producto/${p.slug ?? p.id}`);
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- ruta de la tienda con su slug
		else if (alCarrito) goto(`${base}/carrito`);
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

	const upsell = $derived(
		tienda.masVendidos
			.filter(
				(id) =>
					!carrito.lineas.some((l) => l.id === id) && porId.has(id) && !agotado(porId.get(id)!)
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
	const ch = (on: boolean) => ({
		bd: on ? 'var(--acento, #2F6BFF)' : 'rgba(14,17,22,.1)',
		bg: on ? '#F3F6FF' : '#fff',
		rc: on ? 'var(--acento, #2F6BFF)' : '#C5CCD6'
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
						c: '#14894B',
						p: fmt(subtotal * (1 - tienda.cobros.transferencia / 100))
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
						c: '#00B1EA',
						p:
							tienda.cobros.cuotas > 1
								? `${tienda.cobros.cuotas} × ${fmt(subtotal / tienda.cobros.cuotas)}`
								: fmt(subtotal)
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
			t: 'Productos',
			l: [
				{ t: 'Todo', href: `${base}/catalogo` },
				...tienda.categorias.slice(0, 4).map((c) => ({ t: c.nombre, href: enlaceCat(c.nombre) }))
			]
		},
		{
			t: 'Ayuda',
			l: [
				{ t: 'Envíos', href: `${base}/nosotros` },
				{ t: 'Devoluciones', href: `/t/${tienda.slug}/arrepentimiento` },
				{ t: 'Medios de pago', href: `${base}/nosotros` }
			]
		},
		{
			t: 'Empresa',
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
	// La foto de la portada (o la del destacado); sin ninguna, la portada es sólo texto.
	const fotoHero = $derived(tienda.portada.foto ?? destacado?.foto ?? null);
	const stats = $derived([
		...(tienda.productos.length
			? [{ v: String(tienda.productos.length), t: 'productos', bg: '#0E1116', fg: '#fff' }]
			: []),
		...(tienda.cobros.cuotas > 1
			? [
					{
						v: `${tienda.cobros.cuotas}`,
						t: 'cuotas sin interés',
						bg: 'var(--acento, #2F6BFF)',
						fg: '#fff'
					}
				]
			: []),
		...(tienda.cobros.transferencia > 0
			? [
					{
						v: `${tienda.cobros.transferencia}%`,
						t: 'off por transferencia',
						bg: '#fff',
						fg: '#0E1116'
					}
				]
			: []),
		...(tienda.promedio
			? [
					{
						v: tienda.promedio.toLocaleString('es-AR'),
						t: `en ${tienda.resenas.length} ${tienda.resenas.length === 1 ? 'reseña' : 'reseñas'}`,
						bg: '#fff',
						fg: '#0E1116'
					}
				]
			: [])
	]);
	// Reseñas y preguntas con las piezas de Andes: Sora en los títulos, tarjetas blancas redondeadas.
	const tema: TemaOpiniones = {
		texto: '#0E1116',
		suave: '#69717E',
		campo: '#F3F5F8',
		tarjeta: '#fff',
		borde: 'rgba(14,17,22,.12)',
		acento: '#0E1116',
		sobreAcento: '#fff',
		estrella: '#F5A524',
		radio: '14px',
		fuente: "'Instrument Sans', sans-serif",
		titulos: "'Sora', sans-serif",
		pesoTitulos: 700
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
	const anticipacion = $derived.by(() => {
		const m = tienda.turnos?.anticipacionMin ?? 0;
		return m >= 120 && m % 60 === 0 ? `${m / 60} h` : `${m} min`;
	});
	let reservando = $state(false);
	let cancelando = $state(false);
	let seguro = $state(false);

	// ---- Suscripciones (§51; no están en el diseño de esta plantilla: con sus piezas) ----
	const planes = $derived(tienda.suscripciones);
	const alta = untrack(() => (pagina === 'suscripcion' && plan ? new AltaPlan(plan) : null));
	const caja = $derived(plan ? cajaDe(plan, tienda) : null);
	let cambiando = $state(false);
	let seguroBaja = $state(false);
</script>

<svelte:window bind:innerWidth={ancho} />

<svelte:head>
	<link
		href="https://fonts.googleapis.com/css2?family=Sora:wght@500;600;700;800&display=swap"
		rel="stylesheet"
	/>
</svelte:head>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
<div
	class="andes"
	style="min-height:100vh;font-family:'Instrument Sans',system-ui,sans-serif;color:#0E1116"
>
	<div style="container-type:inline-size;width:100%;background:#F3F5F8;position:relative">
		<!-- barra promo -->
		<div
			style="display:flex;justify-content:center;align-items:center;gap:12px;flex-wrap:wrap;padding:8px 16px;background:var(--acento, #2F6BFF);color:#fff;font-size:13px;font-weight:600;text-align:center"
		>
			{#if conteo.actual}<span>{conteo.texto}</span><span
					style="display:flex;gap:4px;font-variant-numeric:tabular-nums"
					>{#each conteo.partes(4) as c (c.u)}<span
							style="padding:2px 7px;border-radius:6px;background:rgba(0,0,0,.22)">{c.v}{c.u}</span
						>{/each}</span
				>{:else}<span
					>{tienda.envios.gratisDesde
						? `Envío gratis desde ${fmt(tienda.envios.gratisDesde)}`
						: tienda.cobros.cuotas > 1
							? `${tienda.cobros.cuotas} cuotas sin interés`
							: 'Compra protegida'}</span
				>{/if}{#if wide && tienda.cobros.cuotas > 1}<span style="opacity:.85;font-weight:500"
					>{tienda.cobros.cuotas} cuotas sin interés{#if tienda.cobros.transferencia > 0}&nbsp;· {tienda
							.cobros.transferencia}% off por transferencia{/if}</span
				>{/if}
		</div>

		<!-- header -->
		<header
			style="position:sticky;top:0;z-index:20;display:flex;align-items:center;gap:16px;padding:12px clamp(14px,3cqw,36px);background:rgba(243,245,248,.88);backdrop-filter:blur(12px);border-bottom:1px solid rgba(14,17,22,.07)"
		>
			{#if nar}<a
					href="{base}/catalogo"
					aria-label="Catálogo"
					style="display:grid;flex:none;color:#0E1116"><Icon icon={I.phList} size={22} /></a
				>{/if}
			<a
				href={base || '/'}
				style="display:flex;align-items:center;gap:9px;white-space:nowrap;flex:0 1 auto;min-width:0;text-decoration:none;color:#0E1116"
				><span
					style="width:32px;height:32px;border-radius:10px;background:#0E1116;display:grid;place-items:center;flex:none"
					><span
						style="width:12px;height:12px;border-radius:50%;border:3px solid var(--acento, #2F6BFF)"
					></span></span
				><span
					style="font-family:'Sora',sans-serif;font-weight:700;font-size:19px;letter-spacing:-.04em;overflow:hidden;text-overflow:ellipsis"
					><LogoTienda {tienda} /></span
				></a
			>
			{#if wide}<nav style="display:flex;gap:2px;margin-left:10px;min-width:0;overflow:hidden">
					{#each nav as n (n.t)}<a
							href={n.href}
							style="text-decoration:none;font:600 13.5px 'Instrument Sans';padding:8px 12px;border-radius:10px;white-space:nowrap;background:{n.on
								? '#fff'
								: 'transparent'};color:#0E1116">{n.t}</a
						>{/each}
				</nav>{/if}
			<span style="flex:1"></span>
			{#if tienda.productos.length}<button
					type="button"
					onclick={() => abrirCmp()}
					aria-label="Comparar"
					style="position:relative;display:flex;align-items:center;gap:6px;white-space:nowrap;flex:none;border:0;cursor:pointer;padding:9px 12px;border-radius:12px;background:#fff;box-shadow:inset 0 0 0 1px rgba(14,17,22,.08);color:#0E1116;font:600 13px 'Instrument Sans'"
					><Icon
						icon={I.phColumnsBold}
						size={16}
					/>{#if wide}Comparar{/if}{#if comparar.length}<span
							style="min-width:18px;height:18px;padding:0 5px;border-radius:999px;background:var(--acento, #2F6BFF);color:#fff;font-size:11px;font-weight:700;display:grid;place-items:center"
							>{comparar.length}</span
						>{/if}</button
				>{/if}
			{#if nar && tienda.turnos}<a
					href="{base}/turnos"
					aria-label="Reservar turno"
					style="display:flex;align-items:center;flex:none;padding:9px 12px;border-radius:12px;background:#fff;box-shadow:inset 0 0 0 1px rgba(14,17,22,.08);color:#0E1116"
					><Icon icon={I.phCalendarPlusBold} size={16} /></a
				>{/if}
			<a
				href="{base}/carrito"
				aria-label="Carrito"
				style="display:flex;align-items:center;gap:7px;white-space:nowrap;flex:none;text-decoration:none;padding:9px 14px;border-radius:12px;background:#0E1116;color:#fff;font:700 13.5px 'Instrument Sans'"
				><Icon icon={I.phShoppingCartSimpleBold} size={16} />{carrito.unidades}</a
			>
		</header>

		{#if tienda.vitrina}
			<div
				style="padding:10px 16px;background:#EAF0FF;color:#0E1116;font-size:13px;text-align:center"
			>
				Es una tienda de ejemplo de la plantilla Andes: podés recorrerla, pero no se compra
				de&nbsp;verdad.
			</div>
		{/if}

		<!-- ============ COMPARAR (sobre cualquier página) ============ -->
		{#if comparando}
			<section style="padding:24px clamp(14px,3cqw,36px) clamp(48px,7cqw,90px)">
				<div
					style="display:flex;justify-content:space-between;align-items:end;gap:16px;flex-wrap:wrap"
				>
					<h1
						style="margin:0;font-family:'Sora',sans-serif;font-weight:700;font-size:clamp(36px,5.4cqw,68px);line-height:.98;letter-spacing:-.055em"
					>
						Compará lado a lado
					</h1>
					<button
						type="button"
						onclick={() => (abiertoEn = null)}
						style="border:0;cursor:pointer;padding:10px 16px;border-radius:12px;background:#fff;color:#0E1116;font:700 13.5px 'Instrument Sans'"
						>Volver</button
					>
				</div>
				{#if cmp.length}
					<p style="margin:8px 0 0;font-size:15.5px;color:#69717E">
						{cmp.length > 1
							? 'Marcamos en azul lo mejor de cada fila.'
							: 'Sumá otro producto para compararlo.'}
					</p>
					<div
						style="margin-top:20px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,240px),1fr));gap:10px"
					>
						{#each cmp as c (c.p.id)}
							<div
								style="position:relative;padding:14px;border-radius:24px;background:#fff;border:2px solid {c.bd};display:flex;flex-direction:column;gap:12px"
							>
								{#if c.pick}<span
										style="position:absolute;top:-11px;left:14px;padding:4px 10px;border-radius:999px;background:var(--acento, #2F6BFF);color:#fff;font-size:11.5px;font-weight:700"
										>El más vendido</span
									>{/if}
								<a
									href="{base}/producto/{c.p.slug ?? c.p.id}"
									style="display:block;aspect-ratio:4/3;border-radius:16px;overflow:hidden;background:#EEF1F5"
									>{#if c.img}<img
											src={c.img}
											alt=""
											style="width:100%;height:100%;object-fit:cover"
										/>{/if}</a
								>
								<div>
									<div style="font-size:15.5px;font-weight:700">{c.n}</div>
									<div
										style="display:flex;align-items:baseline;gap:8px;margin-top:4px;flex-wrap:wrap"
									>
										<b style="font-family:'Sora',sans-serif;font-size:24px;letter-spacing:-.04em"
											>{c.price}</b
										>{#if c.cuotas}<span style="font-size:12.5px;color:#69717E">{c.cuotas}</span
											>{/if}
									</div>
								</div>
								<div style="display:flex;flex-direction:column">
									{#each c.rows as r (r.t)}<div
											style="display:flex;justify-content:space-between;align-items:center;gap:10px;padding:9px 0;border-top:1px solid #EEF1F5;font-size:13.5px"
										>
											<span style="color:#69717E">{r.t}</span><span
												style="display:flex;align-items:center;gap:6px;font-weight:700;color:{r.c};text-align:right"
												>{#if r.best}<Icon icon={I.phTrophyFill} size={13} />{/if}{r.v}</span
											>
										</div>{/each}
								</div>
								<div style="margin-top:auto;display:flex;flex-direction:column;gap:6px">
									<button
										type="button"
										onclick={() => sumar(c.p, true)}
										disabled={agotado(c.p) || tienda.vitrina}
										style="padding:13px;border:0;border-radius:12px;background:{c.btnBg};color:#fff;font:700 14px 'Instrument Sans';cursor:pointer"
										>{agotado(c.p)
											? 'Sin stock'
											: c.p.opciones?.length
												? 'Elegir opciones'
												: 'Agregar al carrito'}</button
									><button
										type="button"
										onclick={() => alternar(c.p.id)}
										style="display:flex;justify-content:center;align-items:center;gap:6px;padding:7px 11px;border:0;border-radius:10px;background:#EAF0FF;color:var(--acento, #2F6BFF);font:600 13.5px 'Instrument Sans';cursor:pointer"
										>Quitar de la comparación<Icon icon={I.phXBold} size={11} /></button
									>
								</div>
							</div>
						{/each}
					</div>
				{:else}
					<p style="margin:8px 0 0;font-size:15.5px;color:#69717E">
						Tocá «Comparar» en un producto para sumarlo&nbsp;acá.
					</p>
				{/if}
				{#if whatsapp}
					<div
						style="margin-top:14px;padding:16px 18px;border-radius:20px;background:#0E1116;color:#fff;display:flex;gap:14px;align-items:center;flex-wrap:wrap"
					>
						<Icon icon={I.phChatsCircleFill} size={28} style="color:#9DB8FF" />
						<div style="flex:1;min-width:min(100%,220px)">
							<div style="font-weight:700">¿No te decidís?</div>
							<div style="font-size:13.5px;color:#B4BCCB">
								Contanos para qué lo vas a usar y te recomendamos uno por&nbsp;WhatsApp.
							</div>
						</div>
						<a
							href={wa(
								cmp.length
									? `¡Hola! Estoy comparando ${cmp.map((c) => c.n).join(', ')}. ¿Cuál me recomiendan?`
									: '¡Hola! ¿Me ayudan a elegir?'
							)}
							target="_blank"
							rel="noopener"
							style="display:flex;align-items:center;gap:8px;padding:11px 16px;border-radius:12px;background:#25D366;color:#fff;font-weight:700;font-size:14px;text-decoration:none"
							><Icon icon={I.phWhatsappLogoFill} size={18} />Preguntar</a
						>
					</div>
				{/if}
			</section>
		{:else}
			<!-- ============ INICIO ============ -->
			{#if pagina === 'inicio'}
				<!-- Las secciones en el orden que eligió la tienda en Mi tienda. -->
				{#each enOrden( tienda.secciones, ['hero', 'cats', 'prods', 'revs', 'ig'] ) as seccion (seccion)}
					{#if seccion === 'hero'}
						{#if ver('hero')}
							<section style="padding:clamp(12px,2cqw,24px) clamp(14px,3cqw,36px) 0">
								<div
									style="position:relative;border-radius:30px;overflow:hidden;background:radial-gradient(80% 90% at 70% 40%,#1D2A4A 0%,#0E1116 62%);color:#fff;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,380px),1fr));min-height:{fotoHero
										? 'min(640px,130cqw)'
										: 'min(440px,100cqw)'}"
								>
									<div
										style="position:relative;z-index:1;padding:clamp(24px,4.4cqw,60px);display:flex;flex-direction:column;justify-content:space-between;gap:24px;min-width:0"
									>
										{#if destacado}<div style="display:flex;gap:6px;flex-wrap:wrap">
												<span
													style="padding:6px 11px;border-radius:999px;background:rgba(47,107,255,.2);color:#9DB8FF;font-size:12.5px;font-weight:700;max-width:100%;overflow:hidden;text-overflow:ellipsis;white-space:nowrap"
													>{destacado.etiqueta ?? 'Más vendido'} · {destacado.nombre}</span
												>{#if tienda.envios.gratisDesde}<span
														style="padding:6px 11px;border-radius:999px;background:rgba(255,255,255,.08);font-size:12.5px;font-weight:600;white-space:nowrap"
														>Envío gratis desde {fmt(tienda.envios.gratisDesde)}</span
													>{/if}
											</div>{/if}
										<div style="min-width:0">
											<h1
												style="margin:0;font-family:'Sora',sans-serif;font-weight:700;font-size:clamp(48px,7.6cqw,110px);line-height:.92;letter-spacing:-.06em;overflow-wrap:anywhere"
											>
												{titulo[0]}{#if titulo[1]}<br /><span
														style="background:linear-gradient(90deg,#9DB8FF,var(--acento, #2F6BFF));-webkit-background-clip:text;background-clip:text;color:transparent"
														>{titulo[1]}</span
													>{/if}
											</h1>
											{#if tienda.portada.lema}<p
													style="margin:16px 0 0;font-size:16.5px;line-height:1.55;color:#B4BCCB;max-width:420px"
												>
													{tienda.portada.lema}
												</p>{/if}
										</div>
										{#if destacado}{@const dt = tarjetaDe(destacado)}
											<div>
												<div style="display:flex;align-items:baseline;gap:12px;flex-wrap:wrap">
													<b
														style="font-family:'Sora',sans-serif;font-size:clamp(30px,3.4cqw,42px);letter-spacing:-.04em"
														>{dt.desde ? 'Desde ' : ''}{fmt(dt.precio)}</b
													>{#if destacado.anterior && !dt.desde}<s style="color:#7C8494"
															>{fmt(destacado.anterior)}</s
														><span
															style="padding:4px 9px;border-radius:999px;background:var(--acento, #2F6BFF);font-size:12px;font-weight:700"
															>−{off(destacado)}%</span
														>{/if}
												</div>
												{#if tienda.cobros.cuotas > 1 || tienda.cobros.transferencia > 0}<div
														style="margin-top:4px;font-size:13.5px;color:#B4BCCB"
													>
														{#if tienda.cobros.cuotas > 1}{tienda.cobros.cuotas} cuotas sin interés de
															<b style="color:#fff">{fmt(dt.precio / tienda.cobros.cuotas)}</b
															>{/if}{#if tienda.cobros.transferencia > 0}
															{#if tienda.cobros.cuotas > 1}·{/if}
															<b style="color:#7EE2A8"
																>{fmt(dt.precio * (1 - tienda.cobros.transferencia / 100))} por transferencia</b
															>{/if}
													</div>{/if}
												<div style="margin-top:18px;display:flex;gap:8px;flex-wrap:wrap">
													<a
														href="{base}/producto/{destacado.slug ?? destacado.id}"
														style="text-decoration:none;padding:15px 24px;border:0;border-radius:14px;background:#fff;color:#0E1116;font:700 14.5px 'Instrument Sans';cursor:pointer"
														>Comprar</a
													><button
														type="button"
														onclick={() => abrirCmp(destacado!.id)}
														style="display:flex;align-items:center;gap:7px;padding:15px 20px;border:1px solid rgba(255,255,255,.22);border-radius:14px;background:transparent;color:#fff;font:700 14.5px 'Instrument Sans';cursor:pointer"
														><Icon icon={I.phColumnsBold} size={16} />Comparar con otros</button
													>
												</div>
											</div>
										{:else}
											<div style="display:flex;gap:8px;flex-wrap:wrap">
												<a
													href="{base}/catalogo"
													style="text-decoration:none;padding:15px 24px;border-radius:14px;background:#fff;color:#0E1116;font:700 14.5px 'Instrument Sans'"
													>Ver los productos</a
												>
											</div>
										{/if}
									</div>
									{#if fotoHero}<div style="position:relative;min-height:380px">
											<div
												style="position:absolute;left:50%;top:50%;width:78%;aspect-ratio:1;transform:translate(-50%,-50%);border-radius:50%;background:radial-gradient(closest-side,rgba(47,107,255,.45),rgba(47,107,255,0))"
											></div>
											<img
												src={img(fotoHero, 1100)}
												alt=""
												style="position:absolute;left:50%;top:50%;width:min(86%,560px);aspect-ratio:1;object-fit:cover;transform:translate(-50%,-50%);border-radius:40px;mix-blend-mode:screen;opacity:.92"
											/>
										</div>{/if}
								</div>
							</section>
						{/if}

						<section
							style="padding:12px clamp(14px,3cqw,36px) 0;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,220px),1fr));gap:10px"
						>
							{#each perks as p (p.t)}<div
									style="display:flex;align-items:center;gap:12px;padding:14px 16px;border-radius:18px;background:#fff"
								>
									<span
										style="width:40px;height:40px;border-radius:12px;background:#EAF0FF;display:grid;place-items:center;flex:none"
										><Icon icon={p.i} size={20} style="color:var(--acento, #2F6BFF)" /></span
									>
									<div>
										<div style="font-size:14px;font-weight:700">{p.t}</div>
										<div style="font-size:12.5px;color:#69717E">{p.d}</div>
									</div>
								</div>{/each}
						</section>
					{:else if seccion === 'cats'}
						{#if ver('cats') && cats.length}
							<section style="padding:clamp(44px,6cqw,84px) clamp(14px,3cqw,36px) 0">
								<h2
									style="margin:0;font-family:'Sora',sans-serif;font-weight:700;font-size:clamp(30px,4cqw,50px);line-height:1;letter-spacing:-.05em"
								>
									¿Qué estás buscando?
								</h2>
								<div
									style="margin-top:20px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(46%,200px),1fr));gap:10px"
								>
									{#each cats as u (u.nombre)}<a
											href={enlaceCat(u.nombre)}
											style="position:relative;border:0;padding:16px;border-radius:22px;background:{u.bg};color:{u.fg};cursor:pointer;text-align:left;font-family:inherit;min-height:200px;display:flex;flex-direction:column;justify-content:space-between;overflow:hidden;text-decoration:none"
											><Icon icon={I.phSquaresFourFill} size={28} /><span
												><span
													style="display:block;font-family:'Sora',sans-serif;font-size:21px;font-weight:700;letter-spacing:-.04em;overflow-wrap:anywhere"
													>{u.nombre}</span
												><span style="display:block;margin-top:3px;font-size:13px;opacity:.75"
													>{u.cantidad} {u.cantidad === 1 ? 'producto' : 'productos'}</span
												></span
											></a
										>{/each}
								</div>
							</section>
						{/if}
					{:else if seccion === 'prods'}
						{#if ver('prods') && best.length}
							<section style="padding:clamp(44px,6cqw,84px) clamp(14px,3cqw,36px) 0">
								<div
									style="display:flex;justify-content:space-between;align-items:end;gap:16px;flex-wrap:wrap"
								>
									<h2
										style="margin:0;font-family:'Sora',sans-serif;font-weight:700;font-size:clamp(30px,4cqw,50px);line-height:1;letter-spacing:-.05em"
									>
										Más vendidos
									</h2>
									<a
										href="{base}/catalogo"
										style="text-decoration:none;padding:10px 16px;border-radius:12px;background:#fff;color:#0E1116;font:700 13.5px 'Instrument Sans'"
										>Ver los {tienda.productos.length} productos</a
									>
								</div>
								<div
									style="margin-top:20px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(46%,230px),1fr));gap:12px"
								>
									{#each best as p (p.id)}<div
											class="tarj"
											style="position:relative;padding:10px;border-radius:22px;background:#fff;color:#0E1116;display:flex;flex-direction:column;transition:box-shadow .25s;min-width:0"
										>
											<a
												href={p.href}
												style="display:flex;flex-direction:column;text-decoration:none;color:#0E1116"
												><span
													style="position:relative;display:block;aspect-ratio:1;border-radius:16px;overflow:hidden;background:#EEF1F5"
													>{#if p.img}<img
															src={p.img}
															alt=""
															style="width:100%;height:100%;object-fit:cover"
														/>{/if}{#if p.tag}<span
															style="position:absolute;top:8px;left:8px;padding:4px 9px;border-radius:999px;background:{p.tbg};color:#fff;font-size:11px;font-weight:700"
															>{p.tag}</span
														>{/if}</span
												><span style="display:block;padding:12px 4px 4px"
													>{#if p.cat}<span
															style="display:block;font-size:12px;color:#69717E;font-weight:600"
															>{p.cat}</span
														>{/if}<span
														style="display:block;margin-top:2px;font-size:15px;font-weight:700;line-height:1.25"
														>{p.n}</span
													><span style="display:flex;gap:8px;align-items:baseline;margin-top:10px"
														><b style="font-size:17px;letter-spacing:-.02em">{p.price}</b
														>{#if p.old}<s style="font-size:13px;color:#8A919D">{p.old}</s
															>{/if}</span
													>{#if p.cuotas}<span style="display:block;font-size:12.5px;color:#69717E"
															>{p.cuotas}</span
														>{/if}{#if tienda.cobros.transferencia > 0}<span
															style="display:flex;align-items:center;gap:5px;margin-top:6px;font-size:12px;font-weight:700;color:#14894B"
															><Icon icon={I.phBank} size={13} />{p.transf} por transferencia</span
														>{/if}</span
												></a
											>
											<button
												type="button"
												onclick={() => alternar(p.id)}
												aria-pressed={enCmp(p.id)}
												aria-label="Comparar {p.n}"
												style="position:absolute;top:18px;right:18px;display:flex;align-items:center;gap:5px;padding:5px 9px;border:0;border-radius:999px;background:#fff;font:700 11.5px 'Instrument Sans';color:#0E1116;cursor:pointer"
												><span
													style="width:12px;height:12px;border-radius:4px;border:1.5px solid {enCmp(
														p.id
													)
														? 'var(--acento, #2F6BFF)'
														: '#C5CCD6'};background:{enCmp(p.id)
														? 'var(--acento, #2F6BFF)'
														: '#fff'}"
												></span>{#if wide}Comparar{/if}</button
											>
										</div>{/each}
								</div>
							</section>
						{/if}
						{#if tienda.suscripciones.length}
							<!-- Suscripciones (no están en el diseño: con las piezas de «Más vendidos») -->
							<section style="padding:clamp(44px,6cqw,84px) clamp(14px,3cqw,36px) 0">
								<div
									style="display:flex;justify-content:space-between;align-items:end;gap:16px;flex-wrap:wrap"
								>
									<h2
										style="margin:0;font-family:'Sora',sans-serif;font-weight:700;font-size:clamp(30px,4cqw,50px);line-height:1;letter-spacing:-.05em"
									>
										Suscripciones
									</h2>
								</div>
								<div
									style="margin-top:20px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(46%,230px),1fr));gap:12px"
								>
									{#each tienda.suscripciones as p (p.id)}{@const f =
											p.foto ?? p.productos.find((x) => x.foto)?.foto ?? null}<a
											href="{base}/suscripciones/{p.id}"
											class="tarj"
											style="position:relative;padding:10px;border-radius:22px;background:#fff;color:#0E1116;display:flex;flex-direction:column;transition:box-shadow .25s;min-width:0;text-decoration:none"
											><span
												style="position:relative;display:block;aspect-ratio:1;border-radius:16px;overflow:hidden;background:#EEF1F5"
												>{#if f}<img
														src={img(f, 500, 500)}
														alt=""
														style="width:100%;height:100%;object-fit:cover"
													/>{/if}{#if p.agotado}<span
														style="position:absolute;top:8px;left:8px;padding:4px 9px;border-radius:999px;background:#0E1116;color:#fff;font-size:11px;font-weight:700"
														>Sin lugar</span
													>{/if}</span
											><span style="display:block;padding:12px 4px 4px"
												><span style="display:block;font-size:12px;color:#69717E;font-weight:600"
													>{CADA[p.frecuencia]}</span
												><span
													style="display:block;margin-top:2px;font-size:15px;font-weight:700;line-height:1.25"
													>{p.nombre}</span
												><span style="display:flex;gap:8px;align-items:baseline;margin-top:10px"
													><b style="font-size:17px;letter-spacing:-.02em">{fmt(p.precio)}</b></span
												>{#if queIncluye(p)}<span
														style="display:block;font-size:12.5px;color:#69717E"
														>{queIncluye(p)}</span
													>{/if}</span
											></a
										>{/each}
								</div>
							</section>
						{/if}
					{:else if seccion === 'revs'}
						{#if ver('revs') && tienda.resenas.length}
							<section style="padding:clamp(44px,6cqw,84px) clamp(14px,3cqw,36px) 0">
								<div
									style="display:flex;justify-content:space-between;align-items:end;gap:16px;flex-wrap:wrap"
								>
									<h2
										style="margin:0;font-family:'Sora',sans-serif;font-weight:700;font-size:clamp(30px,4cqw,50px);line-height:1;letter-spacing:-.05em"
									>
										Reseñas
									</h2>
									{#if tienda.promedio}<span
											style="display:flex;align-items:center;gap:8px;font-size:14px"
											><b style="font-family:'Sora',sans-serif;font-size:26px"
												>{tienda.promedio.toLocaleString('es-AR')}</b
											><span style="display:flex;color:#F5A524"
												>{#each estrellas(tienda.promedio) as e, i (i)}<Icon
														icon={e}
														size={15}
													/>{/each}</span
											><span style="color:#69717E"
												>{tienda.resenas.length}
												{tienda.resenas.length === 1 ? 'reseña' : 'reseñas'}</span
											></span
										>{/if}
								</div>
								<div
									style="margin-top:20px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,290px),1fr));gap:12px"
								>
									{#each tienda.resenas.slice(0, 3) as r, i (i)}
										{@const pr = tienda.productos.find((p) => p.nombre === r.producto)}
										<div
											style="padding:16px;border-radius:22px;background:#fff;display:flex;flex-direction:column;gap:12px"
										>
											<div style="display:flex;align-items:center;gap:10px">
												<div style="flex:1;min-width:0">
													<div style="font-size:14px;font-weight:700">{r.autor}</div>
													{#if r.conCompra}<div
															style="font-size:12px;color:#14894B;font-weight:600;display:flex;align-items:center;gap:4px"
														>
															<Icon icon={I.phSealCheckFill} size={13} />Compra verificada{#if r.producto}
																· {r.producto}{/if}
														</div>{:else if r.producto}<div style="font-size:12px;color:#69717E">
															Sobre {r.producto}
														</div>{/if}
												</div>
												<span style="display:flex;color:#F5A524;flex:none"
													>{#each estrellas(r.estrellas) as e, j (j)}<Icon
															icon={e}
															size={13}
														/>{/each}</span
												>
											</div>
											<p style="margin:0;font-size:15px;line-height:1.5">{r.texto}</p>
											{#if r.respuesta}<div
													style="padding:12px;border-radius:14px;background:#F3F5F8;font-size:13.5px;line-height:1.5;color:#3B4250"
												>
													<b style="color:#0E1116">Respuesta de {tienda.nombre}</b><br
													/>{r.respuesta}
												</div>{/if}
											{#if pr?.foto}<div style="display:flex;gap:6px">
													<img
														src={img(pr.foto, 140, 140)}
														alt=""
														style="width:64px;height:64px;border-radius:12px;object-fit:cover;background:#EEF1F5"
													/>
												</div>{/if}
										</div>
									{/each}
								</div>
							</section>
						{/if}
					{:else if seccion === 'ig'}
						{#if ver('ig') && tienda.fotos.length >= 3}
							<section style="padding:clamp(44px,6cqw,84px) clamp(14px,3cqw,36px)">
								<h2
									style="margin:0;font-family:'Sora',sans-serif;font-weight:700;font-size:clamp(30px,4cqw,50px);line-height:1;letter-spacing:-.05em"
								>
									En uso
								</h2>
								<div
									style="margin-top:18px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(30%,160px),1fr));gap:8px"
								>
									{#each tienda.fotos as g (g)}<div
											style="aspect-ratio:1;border-radius:18px;overflow:hidden;background:#EEF1F5"
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
					<div style="height:clamp(44px,6cqw,84px)"></div>
				{/if}
			{/if}

			<!-- ============ CATÁLOGO ============ -->
			{#if pagina === 'catalogo' && !tienda.productos.length}
				<section style="padding:clamp(40px,7cqw,96px) clamp(16px,4cqw,48px)">
					<VacioTienda {tienda} {tema} motivo="sin-productos" />
				</section>
			{:else if pagina === 'catalogo'}
				<section style="padding:22px clamp(14px,3cqw,36px) clamp(48px,7cqw,90px)">
					<div style="font-size:12.5px;color:#69717E">
						<a href={base || '/'} style="color:inherit;text-decoration:none">Inicio</a> /
						<span style="color:#0E1116">{categoria ?? 'Todo'}</span>
					</div>
					<div
						style="margin-top:10px;display:flex;justify-content:space-between;align-items:end;gap:16px;flex-wrap:wrap"
					>
						<h1
							style="margin:0;font-family:'Sora',sans-serif;font-weight:700;font-size:clamp(36px,5.4cqw,68px);line-height:.98;letter-spacing:-.055em;overflow-wrap:anywhere"
						>
							{categoria ?? 'Todo'}
						</h1>
						<span style="font-size:13.5px;color:#69717E"
							>{listado.length} {listado.length === 1 ? 'producto' : 'productos'}</span
						>
					</div>
					<div style="margin-top:18px;display:flex;gap:8px;overflow-x:auto;scrollbar-width:none">
						{#each quick as q (q.t)}<a
								href={enlaceCat(q.c)}
								style="flex:none;display:flex;align-items:center;gap:7px;padding:9px 13px;border-radius:12px;background:{q.on
									? '#0E1116'
									: '#fff'};color:{q.on
									? '#fff'
									: '#0E1116'};font-size:13.5px;font-weight:600;white-space:nowrap;box-shadow:inset 0 0 0 1px rgba(14,17,22,.08);text-decoration:none"
								><Icon icon={q.i} size={15} />{q.t}</a
							>{/each}
					</div>
					<div
						style="margin-top:18px;display:flex;gap:clamp(16px,2.6cqw,32px);align-items:flex-start"
					>
						{#if wide && tienda.categorias.length}<aside
								style="width:240px;flex:none;position:sticky;top:80px;display:flex;flex-direction:column;gap:12px"
							>
								<div style="padding:14px;border-radius:18px;background:#fff">
									<div
										style="display:flex;justify-content:space-between;font-size:13.5px;font-weight:700;margin-bottom:10px"
									>
										Categoría
									</div>
									<div style="display:flex;flex-direction:column;gap:9px;font-size:13.5px">
										{#each tienda.categorias as o (o.nombre)}{@const on = o.nombre === categoria}<a
												href={enlaceCat(on ? null : o.nombre)}
												style="display:flex;align-items:center;gap:9px;color:#0E1116;text-decoration:none"
												><span
													style="width:17px;height:17px;border-radius:5px;border:1.5px solid {on
														? 'var(--acento, #2F6BFF)'
														: '#C5CCD6'};background:{on
														? 'var(--acento, #2F6BFF)'
														: '#fff'};color:#fff;display:grid;place-items:center;flex:none"
													>{#if on}<Icon icon={I.phCheckBold} size={11} />{/if}</span
												>{o.nombre}<span style="margin-left:auto;color:#8A919D">{o.cantidad}</span
												></a
											>{/each}
									</div>
								</div>
							</aside>{/if}
						<div style="flex:1;min-width:0">
							<div
								style="display:flex;justify-content:space-between;align-items:center;gap:10px;flex-wrap:wrap;font-size:13.5px"
							>
								<span style="display:flex;gap:6px;flex-wrap:wrap"
									>{#if categoria}<a
											href={enlaceCat(null)}
											style="display:flex;align-items:center;gap:6px;padding:7px 11px;border-radius:10px;background:#EAF0FF;color:var(--acento, #2F6BFF);font-weight:600;text-decoration:none"
											>{categoria}<Icon icon={I.phXBold} size={11} /></a
										>{/if}</span
								>
								<label
									style="font-weight:700;display:flex;align-items:center;gap:6px;position:relative"
									><select
										value={orden}
										onchange={(e) =>
											goto(enlaceCat(categoria, e.currentTarget.value), { replace: true })}
										aria-label="Ordenar"
										style="appearance:none;-webkit-appearance:none;border:0;background:transparent;padding:0 18px 0 0;font:700 13.5px 'Instrument Sans';color:#0E1116;cursor:pointer"
										>{#each ORDENES as [k, t] (k)}<option value={k}>{t}</option>{/each}</select
									><span style="position:absolute;right:0;pointer-events:none;display:grid"
										><Icon icon={I.phCaretDownBold} size={12} /></span
									></label
								>
							</div>
							{#if catalog.length}
								<div
									style="margin-top:16px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(46%,210px),1fr));gap:10px"
								>
									{#each catalog as p (p.id)}<a
											href={p.href}
											style="display:block;text-decoration:none;padding:10px;border-radius:20px;background:#fff;text-align:left;font-family:inherit;color:#0E1116;min-width:0"
											><span
												style="position:relative;display:block;aspect-ratio:1;border-radius:14px;overflow:hidden;background:#EEF1F5"
												>{#if p.img}<img
														src={p.img}
														alt=""
														style="width:100%;height:100%;object-fit:cover"
													/>{/if}{#if p.tag}<span
														style="position:absolute;top:8px;left:8px;padding:4px 9px;border-radius:999px;background:{p.tbg};color:#fff;font-size:11px;font-weight:700"
														>{p.tag}</span
													>{/if}</span
											><span style="display:block;padding:10px 4px 2px"
												><span
													style="display:block;font-size:14.5px;font-weight:700;line-height:1.25"
													>{p.n}</span
												><span style="display:flex;gap:8px;align-items:baseline;margin-top:8px"
													><b style="font-size:16px">{p.price}</b>{#if p.old}<s
															style="font-size:12.5px;color:#8A919D">{p.old}</s
														>{/if}</span
												>{#if p.cuotas}<span style="display:block;font-size:12px;color:#69717E"
														>{p.cuotas}</span
													>{/if}</span
											></a
										>{/each}
								</div>
								{#if listado.length > mostrar}
									<div style="margin-top:32px;display:flex;justify-content:center">
										<button
											type="button"
											onclick={() => (mostrar += 24)}
											style="border:0;cursor:pointer;padding:12px 22px;border-radius:12px;background:#fff;color:#0E1116;font:700 13.5px 'Instrument Sans'"
											>Ver más</button
										>
									</div>
								{/if}
							{:else}
								<p style="margin:40px 0;font-size:15px;color:#69717E">
									Todavía no hay productos en esta categoría.
								</p>
							{/if}
						</div>
					</div>
				</section>
			{/if}

			<!-- ============ PRODUCTO ============ -->
			{#if pagina === 'producto' && producto && sel}
				<section style="padding:22px clamp(14px,3cqw,36px) 0">
					<div style="font-size:12.5px;color:#69717E">
						<a href={base || '/'} style="color:inherit;text-decoration:none">Inicio</a> / {#if producto.categoria}<a
								href={enlaceCat(producto.categoria)}
								style="color:inherit;text-decoration:none">{producto.categoria}</a
							> /
						{/if}<span style="color:#0E1116">{producto.nombre}</span>
					</div>
				</section>
				<section
					style="padding:14px clamp(14px,3cqw,36px) 0;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,400px),1fr));gap:clamp(18px,3.6cqw,48px);align-items:start"
				>
					<div style="display:flex;flex-direction:column;gap:10px">
						<div
							style="position:relative;aspect-ratio:1;border-radius:28px;overflow:hidden;background:radial-gradient(70% 70% at 50% 45%,#1D2A4A,#0E1116)"
						>
							{#if sel.foto}<img
									src={img(sel.foto, 1000)}
									alt={producto.nombre}
									style="width:100%;height:100%;object-fit:cover;mix-blend-mode:screen;opacity:.95"
								/>{/if}
						</div>
						{#if sel.fotos.length > 1}<div
								style="display:grid;grid-template-columns:repeat(5,minmax(0,1fr));gap:8px"
							>
								{#each sel.fotos as f, i (f)}<button
										type="button"
										aria-label="Ver la foto {i + 1}"
										aria-pressed={i === sel.fotoIdx}
										onclick={() => sel.verFoto(i)}
										style="aspect-ratio:1;padding:0;border:0;border-radius:14px;overflow:hidden;background:radial-gradient(70% 70% at 50% 45%,#1D2A4A,#0E1116);cursor:pointer;box-shadow:0 0 0 {i ===
										sel.fotoIdx
											? '2px'
											: '0px'} var(--acento, #2F6BFF)"
										><img
											src={img(f, 200, 200)}
											alt=""
											style="display:block;width:100%;height:100%;object-fit:cover;mix-blend-mode:screen;opacity:.95"
										/></button
									>{/each}
							</div>{/if}
					</div>
					<div style="position:sticky;top:80px;min-width:0">
						<div style="display:flex;gap:6px;flex-wrap:wrap">
							{#if esMasVendido}<span
									style="padding:5px 10px;border-radius:999px;background:#EAF0FF;color:var(--acento, #2F6BFF);font-size:12px;font-weight:700"
									>Más vendido{producto.categoria ? ` en ${producto.categoria}` : ''}</span
								>{:else if producto.etiqueta}<span
									style="padding:5px 10px;border-radius:999px;background:#EAF0FF;color:var(--acento, #2F6BFF);font-size:12px;font-weight:700"
									>{producto.etiqueta}</span
								>{/if}{#if sel.stock > 0}<span
									style="padding:5px 10px;border-radius:999px;background:#E7F6EE;color:#14894B;font-size:12px;font-weight:700"
									>{sel.stock <= 3 && (!sel.conOpciones || sel.variante)
										? (sel.stock === 1 ? 'Queda 1' : `Quedan ${sel.stock}`) +
											(sel.variante ? ` en ${sel.texto.toLowerCase()}` : '')
										: 'En stock'}</span
								>{:else}<span
									style="padding:5px 10px;border-radius:999px;background:#F3F5F8;color:#69717E;font-size:12px;font-weight:700"
									>Sin stock</span
								>{/if}
						</div>
						<h1
							style="margin:12px 0 0;font-family:'Sora',sans-serif;font-weight:700;font-size:clamp(32px,4cqw,50px);line-height:1;letter-spacing:-.05em;overflow-wrap:anywhere"
						>
							{producto.nombre}
						</h1>
						{#if resenasDe.length}<div
								style="margin-top:8px;display:flex;align-items:center;gap:8px;font-size:13.5px"
							>
								<span style="display:flex;color:#F5A524"
									>{#each estrellas(promedioDe) as e, i (i)}<Icon icon={e} size={14} />{/each}</span
								><b>{promedioDe.toLocaleString('es-AR', { maximumFractionDigits: 1 })}</b><span
									style="color:#69717E"
									>{resenasDe.length} {resenasDe.length === 1 ? 'reseña' : 'reseñas'}</span
								>
							</div>{/if}
						<div style="margin-top:16px;padding:16px;border-radius:20px;background:#fff">
							<div style="display:flex;align-items:baseline;gap:10px;flex-wrap:wrap">
								<b style="font-family:'Sora',sans-serif;font-size:34px;letter-spacing:-.04em"
									>{sel.desde ? 'Desde ' : ''}{fmt(sel.precio)}</b
								>{#if tachado}<s style="color:#8A919D">{fmt(tachado)}</s><span
										style="padding:3px 8px;border-radius:999px;background:var(--acento, #2F6BFF);color:#fff;font-size:12px;font-weight:700"
										>−{Math.round((1 - sel.precio / tachado) * 100)}%</span
									>{/if}
							</div>
							<div
								style="margin-top:10px;display:grid;grid-template-columns:repeat(auto-fit,minmax(130px,1fr));gap:6px"
							>
								{#each plans as p (p.t)}<span
										style="padding:10px;border-radius:12px;border:1.5px solid {p.bd};background:{p.bg};font-size:13px"
										><b style="display:block;font-size:15px">{p.v}</b><span style="color:#69717E"
											>{p.t}</span
										></span
									>{/each}
							</div>
						</div>
						{#each sel.opciones as o (o.nombre)}{@const swatch = o.valores.every(
								(x) => x.color || x.foto
							)}
							<div style="margin-top:14px">
								<div
									style="font-size:13.5px;margin-bottom:8px;color:{sel.aviso && !o.elegido
										? '#B42318'
										: '#69717E'}"
								>
									{o.nombre}{#if o.elegido}: <b style="color:#0E1116">{o.elegido}</b>{/if}
								</div>
								{#if swatch}<div style="display:flex;gap:10px;flex-wrap:wrap">
										{#each o.valores as x (x.valor)}<button
												type="button"
												aria-label="{o.nombre} {x.valor}"
												aria-pressed={x.elegido}
												disabled={!x.disponible && !x.elegido}
												onclick={() => sel.elegir(o.nombre, x.valor)}
												style="position:relative;width:40px;height:40px;padding:0;border:0;border-radius:50%;overflow:hidden;background:#EEF1F5;cursor:pointer;opacity:{x.disponible
													? 1
													: 0.35};box-shadow:inset 0 0 0 1px rgba(14,17,22,.12),0 0 0 {x.elegido
													? '2px'
													: '0px'} #F3F5F8,0 0 0 {x.elegido ? '4px' : '0px'} var(--acento, #2F6BFF)"
												>{#if x.color}<span
														style="display:block;width:100%;height:100%;background:{x.color}"
													></span>{:else}<img
														src={img(x.foto, 80, 80)}
														alt=""
														style="display:block;width:100%;height:100%;object-fit:cover"
													/>{/if}{#if !x.disponible}<span
														style="position:absolute;left:50%;top:-4px;bottom:-4px;width:1.5px;background:#0E1116;transform:rotate(45deg)"
													></span>{/if}</button
											>{/each}
									</div>{:else}<div style="display:flex;gap:6px;flex-wrap:wrap">
										{#each o.valores as x (x.valor)}{@const pv =
												sel.opciones.length === 1
													? sel.variantes.find((v) => v.valores[o.nombre] === x.valor)
													: null}<button
												type="button"
												aria-pressed={x.elegido}
												disabled={!x.disponible && !x.elegido}
												onclick={() => sel.elegir(o.nombre, x.valor)}
												style="min-width:52px;min-height:46px;padding:10px 14px;border-radius:12px;border:1.5px solid {x.elegido
													? 'var(--acento, #2F6BFF)'
													: '#E3E8EF'};background:{x.elegido
													? '#F3F6FF'
													: '#fff'};font:700 14px 'Instrument Sans';color:{x.disponible
													? '#0E1116'
													: '#8A919D'};text-align:left;text-decoration:{x.disponible
													? 'none'
													: 'line-through'};cursor:{x.disponible ? 'pointer' : 'default'}"
												>{x.valor}{#if pv}<span
														style="display:block;font-size:12.5px;color:#69717E;font-weight:400"
														>{fmt(pv.precio ?? producto.precio)}</span
													>{/if}</button
											>{/each}
									</div>{/if}
							</div>{/each}
						<div style="margin-top:14px;display:flex;gap:8px;flex-wrap:wrap">
							<button
								type="button"
								onclick={agregar}
								disabled={sel.stock === 0 || tienda.vitrina}
								style="flex:1;min-width:180px;padding:16px;border:0;border-radius:14px;background:#0E1116;color:#fff;font:700 15px 'Instrument Sans';cursor:pointer"
								>{sel.boton('Agregar al carrito')}</button
							><button
								type="button"
								onclick={() => abrirCmp(producto!.id)}
								style="padding:16px;border:0;border-radius:14px;background:#fff;box-shadow:inset 0 0 0 1.5px rgba(14,17,22,.12);color:#0E1116;cursor:pointer;display:grid;place-items:center"
								aria-label="Comparar"><Icon icon={I.phColumnsBold} size={18} /></button
							>
						</div>
						{#if whatsapp}<a
								href={wa(
									`¡Hola! Quiero comprar ${producto.nombre}${sel.variante ? ` (${sel.texto})` : ''} (${fmt(sel.precio)}).`
								)}
								target="_blank"
								rel="noopener"
								style="margin-top:8px;display:flex;justify-content:center;align-items:center;gap:8px;padding:15px;border-radius:14px;background:#fff;box-shadow:inset 0 0 0 1.5px rgba(14,17,22,.12);color:#0E1116;font:700 14.5px 'Instrument Sans';text-decoration:none"
								><Icon icon={I.phWhatsappLogoFill} size={18} style="color:#25D366" />Comprar por
								WhatsApp</a
							>{/if}
						<LanzamientoFicha {tienda} {producto} {sel} {tema} />
						<CompartirProducto {tienda} {producto} {tema} />
						<div style="margin-top:10px;padding:14px;border-radius:16px;background:#fff">
							<div style="display:flex;gap:6px">
								<input
									bind:value={cp}
									inputmode="numeric"
									maxlength="4"
									placeholder="Código postal"
									aria-label="Código postal"
									style="flex:1;min-width:0;padding:10px 12px;border:0;border-radius:10px;background:#F3F5F8;font:400 14px 'Instrument Sans';color:#0E1116"
									onkeydown={(e) => e.key === 'Enter' && calcular()}
								/><button
									type="button"
									onclick={calcular}
									style="padding:10px 14px;border:0;border-radius:10px;background:#0E1116;color:#fff;font:700 13px 'Instrument Sans';cursor:pointer"
									>Calcular</button
								>
							</div>
							{#if cotizado}<div
									style="margin-top:10px;display:flex;flex-direction:column;gap:6px;font-size:13.5px"
								>
									{#each cotizado as o (o.id)}<span
											style="display:flex;justify-content:space-between;gap:10px"
											><span style="display:flex;align-items:center;gap:6px"
												>{#if o.id === 'moto'}<Icon
														icon={I.phLightningFill}
														size={14}
														style="color:#14894B"
													/>{/if}{o.nombre} · {o.detalle}</span
											><b style="white-space:nowrap;color:{o.precio ? '#0E1116' : '#14894B'}"
												>{o.precio ? fmt(o.precio) : 'Gratis'}</b
											></span
										>{/each}
								</div>{/if}
						</div>
						<div style="margin-top:14px;display:flex;flex-direction:column;gap:8px">
							{#each acc as a (a.t)}<div style="padding:14px;border-radius:16px;background:#fff">
									<button
										type="button"
										onclick={() => (abierto = a.open ? '' : a.t)}
										aria-expanded={a.open}
										style="width:100%;display:flex;justify-content:space-between;align-items:center;padding:0;border:0;background:transparent;font:700 13.5px 'Instrument Sans';color:#0E1116;cursor:pointer;text-align:left"
										>{a.t}<Icon icon={a.i} size={12} /></button
									>{#if a.open}<p
											style="margin:8px 0 0;font-size:14px;line-height:1.55;color:#3B4250"
										>
											{a.d}
										</p>{/if}
								</div>{/each}
						</div>
					</div>
				</section>
				<section style="padding:clamp(28px,4cqw,48px) clamp(14px,3cqw,36px) 0">
					<div
						style="max-width:820px;padding:clamp(18px,3cqw,28px);border-radius:24px;background:#fff"
					>
						<PreguntasProducto
							slug={tienda.slug}
							productoId={producto.id}
							preguntas={tienda.preguntas}
							tema={{ ...tema, tarjeta: 'transparent' }}
							titulo="Preguntas sobre este producto"
						/>
					</div>
				</section>
				{#if related.length}
					<section style="padding:clamp(48px,7cqw,90px) clamp(14px,3cqw,36px)">
						<h2
							style="margin:0;font-family:'Sora',sans-serif;font-weight:700;font-size:clamp(28px,3.6cqw,44px);line-height:1;letter-spacing:-.05em"
						>
							También te puede interesar
						</h2>
						<div
							style="margin-top:18px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(46%,200px),1fr));gap:10px"
						>
							{#each related as p (p.id)}<a
									href={p.href}
									style="display:block;text-decoration:none;padding:10px;border-radius:20px;background:#fff;text-align:left;font-family:inherit;color:#0E1116;min-width:0"
									><span
										style="display:block;aspect-ratio:1;border-radius:14px;overflow:hidden;background:#EEF1F5"
										>{#if p.img}<img
												src={p.img}
												alt=""
												style="width:100%;height:100%;object-fit:cover"
											/>{/if}</span
									><span style="display:block;padding:10px 4px 2px;font-size:14px;font-weight:700"
										>{p.n}</span
									><b style="display:block;padding:0 4px 2px;font-size:15px">{p.price}</b></a
								>{/each}
						</div>
					</section>
				{:else}
					<div style="height:clamp(48px,7cqw,90px)"></div>
				{/if}
			{/if}

			<!-- ============ CARRITO ============ -->
			{#if pagina === 'carrito'}
				<section style="padding:22px clamp(14px,3cqw,36px) clamp(48px,7cqw,90px)">
					<h1
						style="margin:0;font-family:'Sora',sans-serif;font-weight:700;font-size:clamp(36px,5.4cqw,64px);line-height:1;letter-spacing:-.055em"
					>
						Carrito
					</h1>
					{#if !lineas.length}
						<p style="margin:16px 0 0;font-size:15px;color:#69717E">Todavía no sumaste nada.</p>
						<a
							href="{base}/catalogo"
							style="display:inline-block;margin-top:16px;text-decoration:none;padding:15px 24px;border-radius:14px;background:#0E1116;color:#fff;font:700 14.5px 'Instrument Sans'"
							>Ver los productos</a
						>
					{:else}
						{#if tienda.envios.gratisDesde}
							<div
								style="margin-top:14px;padding:12px 14px;border-radius:16px;background:#EAF0FF;display:flex;gap:12px;align-items:center;flex-wrap:wrap;font-size:13.5px"
							>
								<span style="flex:1;min-width:200px;display:flex;align-items:center;gap:10px"
									><Icon
										icon={I.phTruck}
										size={20}
										style="color:var(--acento, #2F6BFF);flex:none"
									/>{#if falta}<span>Te faltan <b>{fmt(falta)}</b> para el envío gratis</span
										>{:else}<b>Tenés envío gratis</b>{/if}</span
								><span
									style="flex:1;min-width:160px;height:6px;border-radius:3px;background:#fff;overflow:hidden"
									><span
										style="display:block;height:100%;width:{Math.min(
											100,
											(subtotal / tienda.envios.gratisDesde) * 100
										)}%;background:var(--acento, #2F6BFF);border-radius:3px"
									></span></span
								>
							</div>
						{/if}
						{#if avisoMayor && porMayor}
							<div
								style="margin-top:14px;padding:12px 14px;border-radius:16px;background:#EAF0FF;display:flex;gap:12px;align-items:center;flex-wrap:wrap;font-size:13.5px"
							>
								<span style="flex:1;min-width:200px;display:flex;align-items:center;gap:10px"
									><Icon
										icon={I.phPackage}
										size={20}
										style="color:var(--acento, #2F6BFF);flex:none"
									/><span>{avisoMayor.pre}<b>{avisoMayor.b}</b>{avisoMayor.post}</span></span
								><span
									style="flex:1;min-width:160px;height:6px;border-radius:3px;background:#fff;overflow:hidden"
									><span
										style="display:block;height:100%;width:{porMayor.progreso}%;background:var(--acento, #2F6BFF);border-radius:3px"
									></span></span
								>
							</div>
						{/if}
						<div
							style="margin-top:20px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,340px),1fr));gap:clamp(14px,3cqw,32px);align-items:start"
						>
							<div style="display:flex;flex-direction:column;gap:10px;min-width:0">
								{#each lineas as c (c.clave)}<div
										style="display:flex;gap:14px;padding:12px;border-radius:20px;background:#fff"
									>
										{#if c.producto.foto}<img
												src={img(c.producto.foto, 200, 200)}
												alt=""
												style="width:92px;height:92px;border-radius:14px;object-fit:cover;flex:none;background:#EEF1F5"
											/>{/if}
										<div style="flex:1;min-width:0;display:flex;flex-direction:column">
											<div style="display:flex;justify-content:space-between;gap:10px">
												<a
													href="{base}/producto/{c.producto.slug ?? c.id}"
													style="font-size:15px;font-weight:700;color:#0E1116;text-decoration:none"
													>{c.producto.nombre}</a
												><b style="white-space:nowrap">{fmt(c.subtotal)}</b>
											</div>
											<span style="font-size:13px;color:#69717E"
												>{c.variante ? `${c.variante} · ` : ''}{fmt(c.producto.precio)} c/u</span
											>
											<span style="flex:1;min-height:10px"></span>
											<div style="display:flex;justify-content:space-between;align-items:center">
												<span
													style="display:flex;align-items:center;gap:4px;padding:2px 4px;border-radius:10px;background:#F3F5F8"
													><button
														type="button"
														aria-label="Uno menos"
														onclick={() => carrito.fijar(c.clave, c.cantidad - 1)}
														style="width:28px;height:28px;display:grid;place-items:center;border:0;background:transparent;cursor:pointer;color:#0E1116"
														><Icon icon={I.phMinusBold} size={12} /></button
													><b style="font-size:14px;min-width:16px;text-align:center"
														>{c.cantidad}</b
													><button
														type="button"
														aria-label="Uno más"
														disabled={c.cantidad >= c.producto.stock}
														onclick={() => carrito.fijar(c.clave, c.cantidad + 1)}
														style="width:28px;height:28px;display:grid;place-items:center;border:0;background:transparent;cursor:pointer;color:#0E1116"
														><Icon icon={I.phPlusBold} size={12} /></button
													></span
												><button
													type="button"
													aria-label="Quitar"
													onclick={() => carrito.fijar(c.clave, 0)}
													style="border:0;background:transparent;padding:4px;cursor:pointer;display:grid;color:#8A919D"
													><Icon icon={I.phTrash} size={18} /></button
												>
											</div>
										</div>
									</div>{/each}
								{#if upsell.length}
									<div style="padding:14px;border-radius:20px;background:#fff">
										<div style="font-size:14px;font-weight:700;margin-bottom:10px">
											Te puede servir
										</div>
										<div style="display:flex;flex-direction:column;gap:8px">
											{#each upsell as a (a.id)}<div
													style="display:flex;align-items:center;gap:12px"
												>
													{#if a.foto}<img
															src={img(a.foto, 120, 120)}
															alt=""
															style="width:46px;height:46px;border-radius:10px;object-fit:cover;background:#EEF1F5;flex:none"
														/>{/if}<a
														href="{base}/producto/{a.slug ?? a.id}"
														style="flex:1;min-width:0;font-size:13.5px;font-weight:600;color:#0E1116;text-decoration:none"
														>{a.nombre}</a
													><b style="font-size:13.5px;white-space:nowrap">{tarjeta(a).price}</b
													><button
														type="button"
														aria-label="Sumar {a.nombre}"
														onclick={() => sumar(a)}
														style="width:30px;height:30px;border:0;border-radius:9px;background:#EAF0FF;color:var(--acento, #2F6BFF);display:grid;place-items:center;cursor:pointer;flex:none"
														><Icon icon={I.phPlusBold} size={14} /></button
													>
												</div>{/each}
										</div>
									</div>
								{/if}
							</div>
							<div
								style="position:sticky;top:80px;padding:20px;border-radius:24px;background:#fff;display:flex;flex-direction:column;gap:11px;font-size:14px"
							>
								<div
									style="font-family:'Sora',sans-serif;font-weight:700;font-size:22px;letter-spacing:-.04em"
								>
									Resumen
								</div>
								<span style="display:flex;justify-content:space-between"
									><span>Subtotal</span><span>{fmt(subtotal)}</span></span
								>
								<span style="display:flex;justify-content:space-between;gap:10px"
									><span>Envío</span><span style="color:#69717E;text-align:right"
										>Se calcula en el próximo paso</span
									></span
								>
								<div style="height:1px;background:#EEF1F5"></div>
								<span style="display:flex;justify-content:space-between;align-items:baseline"
									><b>Total</b><b
										style="font-family:'Sora',sans-serif;font-size:28px;letter-spacing:-.04em"
										>{fmt(subtotal)}</b
									></span
								>
								{#if tienda.cobros.cuotas > 1}<span style="color:#69717E"
										>{tienda.cobros.cuotas} cuotas sin interés de {fmt(
											subtotal / tienda.cobros.cuotas
										)}</span
									>{/if}
								{#if tienda.cobros.transferencia > 0}
									<span style="color:#14894B;font-weight:700"
										>{fmt(subtotal * (1 - tienda.cobros.transferencia / 100))} por transferencia</span
									>
								{/if}
								<a
									href="{base}/checkout"
									style="margin-top:4px;padding:16px;border-radius:14px;background:#0E1116;color:#fff;font:700 15px 'Instrument Sans';text-align:center;text-decoration:none"
									>Continuar</a
								>
							</div>
						</div>
					{/if}
				</section>
			{/if}

			<!-- ============ CHECKOUT ============ -->
			{#if pagina === 'checkout'}
				{#if !lineas.length}
					<section style="padding:22px clamp(14px,3cqw,36px) clamp(48px,7cqw,90px)">
						<p style="margin:0;font-size:15px;color:#69717E">Tu carrito está vacío.</p>
						<a
							href="{base}/catalogo"
							style="display:inline-block;margin-top:10px;color:var(--acento, #2F6BFF)"
							>Ver los productos</a
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
						style="padding:22px clamp(14px,3cqw,36px) clamp(48px,7cqw,90px);display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,360px),1fr));gap:clamp(14px,3cqw,36px);align-items:start"
					>
						<input type="hidden" name="carrito" value={JSON.stringify(carrito.lineas)} />
						<input type="hidden" name="envio" value={envio ?? ''} />
						<input type="hidden" name="pago" value={pago} />
						<div style="display:flex;flex-direction:column;gap:14px;min-width:0">
							<!-- 1 Datos -->
							<div
								style="padding:16px;border-radius:20px;background:#fff;border:2px solid transparent"
							>
								<div style="display:flex;align-items:center;gap:10px">
									<span
										style="width:26px;height:26px;border-radius:50%;background:var(--acento, #2F6BFF);color:#fff;display:grid;place-items:center;font-size:12.5px;font-weight:700"
										>1</span
									><b style="flex:1;font-size:15.5px">Tus datos</b>
								</div>
								<div
									style="margin-top:12px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,200px),1fr));gap:8px"
								>
									{#each [['email', 'Email', 'email', 'email'], ['nombre', 'Nombre y apellido', 'text', 'name'], ['telefono', 'Teléfono', 'tel', 'tel'], ['dni', 'DNI', 'text', 'off'], ...camposFactura(tienda)] as [n, l, tipo, ac] (n)}
										<label style="display:block;min-width:0"
											><span style="display:block;font-size:12.5px;color:#69717E;margin-bottom:4px"
												>{l}</span
											><input
												name={n}
												type={tipo}
												autocomplete={ac as 'email'}
												required={!OPCIONALES.has(n)}
												maxlength="120"
												style="width:100%;padding:10px 12px;border:0;border-radius:10px;background:#F3F5F8;font:400 14px 'Instrument Sans';color:#0E1116"
											/></label
										>
									{/each}
								</div>
								<!-- Sin bloque de newsletter en el diseño de esta plantilla: la casilla del checkout (§50). -->
								<label
									style="margin-top:10px;display:flex;align-items:center;gap:10px;min-height:44px;font:400 14px 'Instrument Sans';color:#0E1116;cursor:pointer"
									><input
										type="checkbox"
										name="novedades"
										style="flex:none;width:18px;height:18px;margin:0;accent-color:var(--acento, #2F6BFF)"
									/>Quiero recibir novedades</label
								>
							</div>
							<!-- 2 Entrega -->
							<div
								style="padding:16px;border-radius:20px;background:#fff;border:2px solid {envio
									? 'transparent'
									: 'var(--acento, #2F6BFF)'}"
							>
								<div style="display:flex;align-items:center;gap:10px">
									<span
										style="width:26px;height:26px;border-radius:50%;background:{envio
											? '#E7F6EE'
											: 'var(--acento, #2F6BFF)'};color:{envio
											? '#14894B'
											: '#fff'};display:grid;place-items:center;font-size:12.5px;font-weight:700"
										>{envio ? '✓' : '2'}</span
									><b style="flex:1;font-size:15.5px">Entrega</b><span
										style="font-size:13px;color:#69717E"
										>{elegida
											? `${elegida.nombre} · ${elegida.precio ? fmt(elegida.precio) : 'gratis'}`
											: ''}</span
									>
								</div>
								<div
									style="margin-top:12px;display:grid;grid-template-columns:minmax(0,120px) minmax(0,1fr);gap:8px"
								>
									<label style="display:block;min-width:0"
										><span style="display:block;font-size:12.5px;color:#69717E;margin-bottom:4px"
											>Código postal</span
										><input
											name="cp"
											bind:value={cp}
											inputmode="numeric"
											maxlength="4"
											style="width:100%;padding:10px 12px;border:0;border-radius:10px;background:#F3F5F8;font:400 14px 'Instrument Sans';color:#0E1116"
										/></label
									>
									<label style="display:block;min-width:0"
										><span style="display:block;font-size:12.5px;color:#69717E;margin-bottom:4px"
											>Dirección (calle, número, piso)</span
										><input
											name="direccion"
											maxlength="160"
											autocomplete="street-address"
											required={envio !== 'retiro'}
											style="width:100%;padding:10px 12px;border:0;border-radius:10px;background:#F3F5F8;font:400 14px 'Instrument Sans';color:#0E1116"
										/></label
									>
								</div>
								<div
									style="margin-top:12px;display:flex;flex-direction:column;gap:8px"
									role="radiogroup"
									aria-label="Forma de entrega"
								>
									{#each opciones as s (s.id)}{@const k = ch(envio === s.id)}<button
											type="button"
											role="radio"
											aria-checked={envio === s.id}
											onclick={() => (envio = s.id)}
											style="display:flex;align-items:center;gap:12px;padding:13px;border-radius:14px;border:1.5px solid {k.bd};background:{k.bg};text-align:left;cursor:pointer;font-family:inherit;color:#0E1116"
											><span
												style="width:18px;height:18px;border-radius:50%;border:2px solid {k.rc};display:grid;place-items:center;flex:none"
												>{#if envio === s.id}<span
														style="width:8px;height:8px;border-radius:50%;background:var(--acento, #2F6BFF)"
													></span>{/if}</span
											>
											<div style="flex:1;min-width:0">
												<div style="font-size:14.5px;font-weight:600">{s.nombre}</div>
												<div style="font-size:12.5px;color:#69717E">{s.detalle}</div>
											</div>
											<b style="font-size:13.5px;color:{s.precio ? '#0E1116' : '#14894B'}"
												>{s.precio ? fmt(s.precio) : 'Gratis'}</b
											></button
										>{:else}<span style="font-size:13.5px;color:#69717E"
											>Poné tu código postal para ver cómo te llega.</span
										>{/each}
								</div>
								<EnvioCorreo {tienda} {cp} {envio} {opciones} />
							</div>
							<!-- 3 Pago -->
							<div
								style="padding:16px;border-radius:20px;background:#fff;border:2px solid {envio
									? 'var(--acento, #2F6BFF)'
									: 'transparent'}"
							>
								<div style="display:flex;align-items:center;gap:10px">
									<span
										style="width:26px;height:26px;border-radius:50%;background:var(--acento, #2F6BFF);color:#fff;display:grid;place-items:center;font-size:12.5px;font-weight:700"
										>3</span
									><b style="flex:1;font-size:15.5px">Pago</b>
								</div>
								<div
									style="margin-top:12px;display:flex;flex-direction:column;gap:8px"
									role="radiogroup"
									aria-label="Forma de pago"
								>
									{#each pagos as p (p.id)}{@const k = ch(pago === p.id)}<button
											type="button"
											role="radio"
											aria-checked={pago === p.id}
											onclick={() => (pago = p.id)}
											style="display:flex;align-items:center;gap:12px;padding:13px;border-radius:14px;border:1.5px solid {k.bd};background:{k.bg};text-align:left;cursor:pointer;font-family:inherit;color:#0E1116"
											><span
												style="width:18px;height:18px;border-radius:50%;border:2px solid {k.rc};display:grid;place-items:center;flex:none"
												>{#if pago === p.id}<span
														style="width:8px;height:8px;border-radius:50%;background:var(--acento, #2F6BFF)"
													></span>{/if}</span
											><Icon icon={p.i} size={22} style="color:{p.c};flex:none" />
											<div style="flex:1;min-width:0">
												<div style="font-size:14.5px;font-weight:600">{p.t}</div>
												<div style="font-size:12.5px;color:#69717E">{p.d}</div>
											</div>
											<b style="font-size:13.5px;white-space:nowrap">{p.p}</b></button
										>{/each}
								</div>
							</div>
							{#if error}<div
									role="alert"
									style="padding:12px 14px;border-radius:14px;background:#FDECEC;color:#B42318;font-size:14px"
								>
									{error}
								</div>{/if}
							<button
								disabled={!envio || !pagos.length || enviando || tienda.vitrina}
								style="padding:17px;border:0;border-radius:14px;background:var(--acento, #2F6BFF);color:#fff;font:700 15.5px 'Instrument Sans';cursor:pointer;display:flex;justify-content:center;align-items:center;gap:8px"
								><Icon icon={I.phLockSimpleFill} size={16} />{tienda.vitrina
									? 'Es una tienda de ejemplo'
									: !pagos.length
										? 'Esta tienda todavía no tiene medios de pago'
										: envio
											? `Pagar ${fmt(total)}`
											: 'Elegí cómo te llega'}</button
							>
							<div
								style="display:flex;justify-content:center;gap:12px;align-items:center;flex-wrap:wrap;color:#69717E;font-size:12.5px"
							>
								{#if tienda.cobros.mercadoPago}<Icon icon={I.logosVisa} size={30} /><Icon
										icon={I.logosMastercard}
										size={22}
									/><Icon icon={I.siMercadopago} size={20} style="color:#00B1EA" />{/if}<span
									>Pago seguro con cifrado</span
								>
							</div>
						</div>
						<div
							style="position:sticky;top:80px;padding:20px;border-radius:24px;background:#0E1116;color:#fff"
						>
							<div
								style="font-family:'Sora',sans-serif;font-weight:700;font-size:22px;letter-spacing:-.04em"
							>
								Tu pedido
							</div>
							<div style="margin-top:14px;display:flex;flex-direction:column;gap:12px">
								{#each lineas as c (c.clave)}<div style="display:flex;gap:12px;align-items:center">
										{#if c.producto.foto}<img
												src={img(c.producto.foto, 120, 120)}
												alt=""
												style="width:52px;height:52px;border-radius:12px;object-fit:cover;background:#1D2330;flex:none"
											/>{/if}<span style="flex:1;min-width:0;font-size:14px"
											>{c.producto.nombre}<br /><span style="font-size:12.5px;color:#8A919D"
												>{c.variante ? `${c.variante} · ` : ''}{c.cantidad} × {fmt(
													c.producto.precio
												)}</span
											></span
										><b style="font-size:14px;white-space:nowrap">{fmt(c.subtotal)}</b>
									</div>{/each}
							</div>
							<div
								style="margin-top:16px;padding-top:14px;border-top:1px solid rgba(255,255,255,.1);display:flex;flex-direction:column;gap:8px;font-size:14px"
							>
								<span style="display:flex;justify-content:space-between"
									><span>Subtotal</span><span>{fmt(subtotal)}</span></span
								><span style="display:flex;justify-content:space-between"
									><span>Envío</span><span style={elegida && !cuenta.envio ? 'color:#7EE2A8' : ''}
										>{elegida ? (cuenta.envio ? fmt(cuenta.envio) : 'Gratis') : '—'}</span
									></span
								>{#if cuenta.descCupon}<span
										style="display:flex;justify-content:space-between;color:#7EE2A8"
										><span>Cupón {cupon.aplicado}</span><span>− {fmt(cuenta.descCupon)}</span></span
									>{/if}{#if descuento}<span
										style="display:flex;justify-content:space-between;color:#7EE2A8"
										><span>Transferencia −{tienda.cobros.transferencia}%</span><span
											>− {fmt(descuento)}</span
										></span
									>{/if}
								<div style="margin:6px 0 2px">
									<CampoCupon
										{cupon}
										lineas={carrito.lineas}
										tinta="#fff"
										suave="#8A919D"
										borde="rgba(255,255,255,.14)"
										fondo="#1D2330"
										acento="var(--acento, #2F6BFF)"
										sobreAcento="#fff"
										radio="10px"
										fuente="'Instrument Sans',sans-serif"
										error="#FFB4AB"
									/>
								</div>
								<span
									style="display:flex;justify-content:space-between;align-items:baseline;margin-top:6px"
									><b>Total</b><b style="font-family:'Sora',sans-serif;font-size:26px"
										>{fmt(total)}</b
									></span
								>
							</div>
							<div
								style="margin-top:14px;display:flex;align-items:center;gap:8px;font-size:12.5px;color:#B4BCCB"
							>
								<Icon icon={I.phShieldCheckFill} size={16} style="color:#9DB8FF" />Compra protegida
								· 10 días para arrepentirte
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
					style="padding:22px clamp(14px,3cqw,36px) clamp(48px,7cqw,90px);display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,360px),1fr));gap:clamp(14px,3cqw,36px);align-items:start"
				>
					<input type="hidden" name="servicioId" value={reserva.servicioId ?? ''} />
					<input type="hidden" name="profesionalId" value={reserva.profesionalId ?? ''} />
					<input type="hidden" name="inicio" value={reserva.inicio ?? ''} />
					<div style="display:flex;flex-direction:column;gap:14px;min-width:0">
						<h1
							style="margin:0;font-family:'Sora',sans-serif;font-weight:700;font-size:clamp(32px,4.6cqw,52px);line-height:1;letter-spacing:-.05em;overflow-wrap:anywhere"
						>
							Reservá tu turno
						</h1>
						<div style="padding:16px;border-radius:20px;background:#fff">
							<div style="display:flex;align-items:center;gap:10px">
								<span
									style="width:26px;height:26px;border-radius:50%;background:var(--acento, #2F6BFF);color:#fff;display:grid;place-items:center;font-size:12.5px;font-weight:700"
									>1</span
								><b style="flex:1;font-size:15.5px">Servicio</b>
							</div>
							<div style="margin-top:12px">
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
											style="display:flex;align-items:center;gap:12px;padding:15px;border-radius:14px;border:1.5px solid {ch(
												on
											).bd};background:{ch(on)
												.bg};text-align:left;cursor:pointer;font-family:inherit;color:#0E1116"
											><span
												style="width:18px;height:18px;border-radius:50%;border:2px solid {ch(on)
													.rc};display:grid;place-items:center;flex:none"
												>{#if on}<span
														style="width:8px;height:8px;border-radius:50%;background:var(--acento, #2F6BFF)"
													></span>{/if}</span
											><span style="flex:1;min-width:0"
												><span style="display:block;font-size:14.5px;font-weight:600"
													>{s.nombre}</span
												><span style="display:block;font-size:12.5px;color:#69717E"
													>{duracion(s.duracion)}</span
												></span
											><b style="font-size:14px;white-space:nowrap">{fmt(s.precio)}</b></button
										>{/each}
								</div>
							</div>
						</div>
						<div style="padding:16px;border-radius:20px;background:#fff">
							<div style="display:flex;align-items:center;gap:10px">
								<span
									style="width:26px;height:26px;border-radius:50%;background:var(--acento, #2F6BFF);color:#fff;display:grid;place-items:center;font-size:12.5px;font-weight:700"
									>2</span
								><b style="flex:1;font-size:15.5px">Profesional</b>
							</div>
							<div style="margin-top:12px">
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
											style="flex:none;width:110px;padding:10px;border-radius:14px;border:1.5px solid {ch(
												on
											).bd};background:{ch(on)
												.bg};text-align:center;cursor:pointer;font-family:inherit;color:#0E1116"
											>{#if p?.foto}<img
													src={img(p.foto, 120, 120)}
													alt=""
													style="width:56px;height:56px;border-radius:12px;object-fit:cover;display:block;margin:0 auto;background:#F3F5F8"
												/>{:else}<span
													style="width:56px;height:56px;border-radius:12px;background:#F3F5F8;color:#0E1116;display:grid;place-items:center;margin:0 auto"
													><Icon icon={p ? I.phUserFill : I.phShuffleBold} size={22} /></span
												>{/if}<span
												style="display:block;margin-top:8px;font-size:13.5px;font-weight:600;overflow:hidden;text-overflow:ellipsis;white-space:nowrap"
												>{p?.nombre ?? 'Cualquiera'}</span
											><span
												style="display:block;font-size:11.5px;color:#69717E;overflow:hidden;text-overflow:ellipsis;white-space:nowrap"
												>{p ? p.rol.split(' · ')[0] : 'El primero libre'}</span
											></button
										>{/each}
								</div>
							</div>
						</div>
						<div style="padding:16px;border-radius:20px;background:#fff">
							<div style="display:flex;align-items:center;gap:10px">
								<span
									style="width:26px;height:26px;border-radius:50%;background:var(--acento, #2F6BFF);color:#fff;display:grid;place-items:center;font-size:12.5px;font-weight:700"
									>3</span
								><b style="flex:1;font-size:15.5px">Día y horario</b><span
									style="display:flex;align-items:center;gap:10px;font-size:14px;font-weight:600;color:#0E1116"
									>{#if reserva.hayAntes || reserva.hayDespues}<button
											type="button"
											aria-label="Días anteriores"
											disabled={!reserva.hayAntes}
											onclick={() => reserva.mover(-1)}
											style="display:grid;place-items:center;min-width:44px;min-height:44px;padding:4px;border:0;background:transparent;color:#0E1116;cursor:pointer"
											><Icon icon={I.phCaretLeftBold} size={14} /></button
										>{/if}{reserva.mes}{#if reserva.hayAntes || reserva.hayDespues}<button
											type="button"
											aria-label="Días siguientes"
											disabled={!reserva.hayDespues}
											onclick={() => reserva.mover(1)}
											style="display:grid;place-items:center;min-width:44px;min-height:44px;padding:4px;border:0;background:transparent;color:#0E1116;cursor:pointer"
											><Icon icon={I.phCaretRightBold} size={14} /></button
										>{/if}</span
								>
							</div>
							<div style="margin-top:12px">
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
											style="padding:9px 0;border-radius:14px;border:1.5px solid {ch(on)
												.bd};background:{ch(on)
												.bg};text-align:center;cursor:pointer;font-family:inherit;color:#0E1116;min-width:0;opacity:{d.abierto
												? 1
												: 0.35}"
											><span
												style="display:block;font-size:11px;font-weight:600;text-transform:uppercase;color:#69717E"
												>{d.w}</span
											><span style="display:block;font-size:17px;font-weight:700;margin-top:1px"
												>{d.n}</span
											></button
										>{/each}
								</div>
								<div style="margin-top:16px;display:flex;flex-direction:column;gap:12px">
									{#if !reserva.servicioId}
										<p style="margin:0;font-size:14px;color:#69717E">
											Elegí un servicio para ver los horarios.
										</p>
									{:else if reserva.horarios && !reserva.grupos.length}
										<p style="margin:0;font-size:14px;color:#69717E">
											No quedan horarios este día. Probá con otro día u otro&nbsp;profesional.
										</p>
									{/if}
									{#each reserva.grupos as g (g.k)}<div>
											<div style="font-size:12.5px;font-weight:600;color:#69717E;margin-bottom:7px">
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
														style="padding:10px 0;border-radius:14px;border:1.5px solid {ch(on)
															.bd};background:{ch(on)
															.bg};text-align:center;cursor:pointer;font:600 14px 'Instrument Sans';color:#0E1116;text-decoration:{h.libre
															? 'none'
															: 'line-through'};opacity:{h.libre ? 1 : 0.45}">{h.hora}</button
													>{/each}
											</div>
										</div>{/each}
								</div>
							</div>
						</div>
						{#if reserva.listo}
							<div style="padding:16px;border-radius:20px;background:#fff">
								<div style="display:flex;align-items:center;gap:10px">
									<span
										style="width:26px;height:26px;border-radius:50%;background:var(--acento, #2F6BFF);color:#fff;display:grid;place-items:center;font-size:12.5px;font-weight:700"
										>4</span
									><b style="flex:1;font-size:15.5px">Tus datos</b>
								</div>
								<div style="margin-top:12px">
									<div
										style="display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,200px),1fr));gap:8px"
									>
										{#each [['nombre', 'Nombre', 'text', 'name'], ['telefono', 'WhatsApp', 'tel', 'tel'], ['email', 'Email', 'email', 'email']] as [n, l, tipo, ac] (n)}
											<label style="display:block;min-width:0"
												><span
													style="display:block;font-size:12.5px;color:#69717E;margin-bottom:4px"
													>{l}</span
												><input
													name={n}
													type={tipo}
													autocomplete={ac as 'email'}
													required
													maxlength="120"
													style="width:100%;padding:10px 12px;border:0;border-radius:10px;background:#F3F5F8;font:400 14px 'Instrument Sans';color:#0E1116"
												/></label
											>
										{/each}
									</div>
								</div>
							</div>
							<div style="padding:16px;border-radius:20px;background:#fff">
								<div style="display:flex;align-items:center;gap:10px">
									<span
										style="width:26px;height:26px;border-radius:50%;background:var(--acento, #2F6BFF);color:#fff;display:grid;place-items:center;font-size:12.5px;font-weight:700"
										>5</span
									><b style="flex:1;font-size:15.5px">Pago</b>
								</div>
								<div style="margin-top:12px">
									<div
										style="display:flex;align-items:center;gap:12px;padding:15px;border-radius:14px;border:1.5px solid {ch(
											true
										).bd};background:{ch(true).bg};color:#0E1116"
									>
										<span
											style="width:18px;height:18px;border-radius:50%;border:2px solid {ch(true)
												.rc};display:grid;place-items:center;flex:none"
											><span
												style="width:8px;height:8px;border-radius:50%;background:var(--acento, #2F6BFF)"
											></span></span
										>{#if reserva.sena}<Icon
												icon={I.siMercadopago}
												size={22}
												style="color:#00B1EA;flex:none"
											/><span style="flex:1;min-width:0"
												><span style="display:block;font-size:14.5px;font-weight:600"
													>Mercado Pago</span
												><span style="display:block;font-size:12.5px;color:#69717E"
													>Seña de {fmt(reserva.sena)} para confirmar · el resto, en el local</span
												></span
											>{:else}<Icon
												icon={I.phStorefrontFill}
												size={22}
												style="color:#0E1116;flex:none"
											/><span style="flex:1;min-width:0"
												><span style="display:block;font-size:14.5px;font-weight:600"
													>Pagás en el local</span
												><span style="display:block;font-size:12.5px;color:#69717E"
													>No hace falta seña: el turno queda confirmado</span
												></span
											>{/if}
									</div>
								</div>
							</div>
						{/if}

						{#if error}<div
								role="alert"
								style="padding:12px 14px;border-radius:14px;background:#FDECEC;color:#B42318;font-size:14px"
							>
								{error}
							</div>{/if}
						<button
							disabled={!reserva.listo || reservando || tienda.vitrina}
							style="padding:17px;border:0;border-radius:14px;background:var(--acento, #2F6BFF);color:#fff;font:700 15.5px 'Instrument Sans';cursor:pointer;display:flex;justify-content:center;align-items:center;gap:8px"
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
						style="position:sticky;top:80px;display:flex;flex-direction:column;gap:12px;min-width:0"
					>
						<div style="padding:20px;border-radius:24px;background:#0E1116;color:#fff">
							<div
								style="font-family:'Sora',sans-serif;font-weight:700;font-size:22px;letter-spacing:-.04em"
							>
								Tu reserva
							</div>
							<div style="margin-top:14px;display:flex;align-items:center;gap:12px">
								{#if reserva.profesional?.foto}<img
										src={img(reserva.profesional.foto, 120, 120)}
										alt=""
										style="width:52px;height:52px;border-radius:12px;object-fit:cover;flex:none"
									/>{/if}
								<div style="min-width:0">
									<div style="font-size:15px;font-weight:600">
										{reserva.conQuien ? `Con ${reserva.conQuien}` : 'Con el primero libre'}
									</div>
									<div style="font-size:13px;color:#B4BCCB">
										{reserva.cuando ?? 'Elegí el día y el horario'}
									</div>
								</div>
							</div>
							{#if reserva.servicio}
								<div
									style="margin-top:16px;padding-top:14px;border-top:1px solid rgba(255,255,255,.1);display:flex;flex-direction:column;gap:8px;font-size:14px"
								>
									<span style="display:flex;justify-content:space-between;gap:10px"
										><span>{reserva.servicio.nombre}</span><span style="white-space:nowrap"
											>{fmt(reserva.servicio.precio)}</span
										></span
									><span style="display:flex;justify-content:space-between;color:#B4BCCB"
										><span>Duración total</span><span>{reserva.duracion}</span></span
									><span
										style="display:flex;justify-content:space-between;align-items:baseline;margin-top:6px"
										><b>Total</b><b style="font-family:'Sora',sans-serif;font-size:26px"
											>{fmt(reserva.servicio.precio)}</b
										></span
									>
								</div>
							{/if}
							<div
								style="margin-top:14px;display:flex;align-items:flex-start;gap:8px;font-size:12.5px;line-height:1.45;color:#B4BCCB"
							>
								<Icon icon={I.phShieldCheckFill} size={16} style="flex:none;color:#9DB8FF" /><span
									>{#if reserva.sena}Seña de {fmt(reserva.sena)} para confirmar. Podés cancelar desde
										el link del turno hasta {anticipacion} antes; la seña te la devuelve la&nbsp;tienda.{:else}No
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
				<section style="padding:clamp(28px,4cqw,56px) clamp(14px,3cqw,36px) clamp(48px,7cqw,90px)">
					<div
						style="max-width:640px;margin:0 auto;padding:clamp(20px,3cqw,32px);border-radius:24px;background:#fff;text-align:center"
					>
						<span
							style="display:inline-grid;place-items:center;width:56px;height:56px;border-radius:50%;background:#E7F6EE;color:#14894B"
							><Icon icon={cancelado ? I.phCalendarX : I.phCheckBold} size={26} /></span
						>
						<h1
							style="margin:16px 0 0;font-family:'Sora',sans-serif;font-weight:700;font-size:clamp(32px,4.6cqw,52px);line-height:1;letter-spacing:-.05em;overflow-wrap:anywhere"
						>
							{cancelado
								? 'Turno cancelado'
								: turno.estado === 'reservado'
									? `Falta la seña, ${turno.nombre}`
									: `¡Listo, ${turno.nombre}!`}
						</h1>
						<p style="margin:12px 0 0;font-size:15.5px;line-height:1.55;color:#69717E">
							{#if cancelado}Liberamos el horario{turno.senaPagada
									? '. La seña te la devuelve la tienda por Mercado Pago'
									: ''}. Si querés, reservá&nbsp;otro.{:else if turno.estado === 'reservado'}Te
								guardamos el horario unos minutos mientras pagás la seña con Mercado&nbsp;Pago.{:else}Tu
								turno está confirmado. Te esperamos en
								{turno.tienda}.{/if}
						</p>
						{#if aviso === 'pendiente' || aviso === 'error'}<div
								role="status"
								style="margin-top:16px;padding:12px 14px;border-radius:14px;background:#FDECEC;color:#B42318;font-size:14px"
							>
								{aviso === 'pendiente'
									? 'Mercado Pago todavía no acreditó la seña: te avisamos cuando entre.'
									: 'El pago de la seña no salió. Podés reservar de nuevo.'}
							</div>{/if}
						<div
							style="margin-top:18px;padding:14px 16px;border-radius:16px;background:#EAF0FF;text-align:left;font-size:14.5px;line-height:1.6"
						>
							<div
								style="font-size:16px;font-weight:600;text-decoration:{cancelado
									? 'line-through'
									: 'none'}"
							>
								{cuandoLargo(turno.inicio)} con {turno.profesional}
							</div>
							<div style="font-size:13.5px;color:#69717E;margin-top:2px">
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
								style="margin-top:12px;padding:12px 14px;border-radius:14px;background:#FDECEC;color:#B42318;font-size:14px"
							>
								{error}
							</div>{/if}
						<div
							style="margin-top:22px;display:flex;justify-content:center;gap:10px;flex-wrap:wrap"
						>
							{#if !cancelado}<a
									href="{base}/turno/{turno.codigo}/turno.ics"
									download="turno.ics"
									data-sveltekit-reload
									style="display:flex;align-items:center;gap:8px;text-decoration:none;padding:15px 24px;border:0;border-radius:14px;background:#0E1116;color:#fff;font:700 14.5px 'Instrument Sans'"
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
											style="text-decoration:none;padding:15px 24px;border:0;border-radius:14px;background:#0E1116;color:#fff;font:700 14.5px 'Instrument Sans';cursor:pointer"
											>Sí, cancelar el turno</button
										>{:else}<button
											type="button"
											onclick={() => (seguro = true)}
											style="text-decoration:none;padding:15px 24px;border:0;border-radius:14px;background:#fff;box-shadow:inset 0 0 0 1px rgba(14,17,22,.08);color:#0E1116;font:700 14.5px 'Instrument Sans';cursor:pointer"
											>Cancelar turno</button
										>{/if}
								</form>
							{:else if cancelado}<a
									href="{base}/turnos"
									style="display:flex;align-items:center;gap:8px;text-decoration:none;padding:15px 24px;border:0;border-radius:14px;background:#0E1116;color:#fff;font:700 14.5px 'Instrument Sans'"
									><Icon icon={I.phCalendarPlusBold} size={17} />Reservar otro</a
								>{/if}
							<a
								href={base || '/'}
								style="text-decoration:none;padding:15px 24px;border:0;border-radius:14px;background:#fff;box-shadow:inset 0 0 0 1px rgba(14,17,22,.08);color:#0E1116;font:700 14.5px 'Instrument Sans'"
								>Volver a la tienda</a
							>
						</div>
						{#if !cancelado && !turno.cancelable}<p
								style="margin:14px 0 0;font-size:13.5px;color:#69717E"
							>
								Para cambiarlo o cancelarlo, escribile a la&nbsp;tienda{#if whatsapp}&nbsp;<a
										href={wa(`¡Hola! Tengo un turno el ${cuandoLargo(turno.inicio)}.`)}
										target="_blank"
										rel="noopener"
										style="color:#0E1116">por WhatsApp</a
									>{/if}.
							</p>{/if}
					</div>
				</section>
			{/if}

			<!-- ============ LISTO (no está en el diseño: con las piezas del checkout) ============ -->
			{#if pagina === 'listo' && pedido}
				<section style="padding:clamp(28px,4cqw,56px) clamp(14px,3cqw,36px) clamp(48px,7cqw,90px)">
					<div
						style="max-width:640px;margin:0 auto;padding:clamp(20px,3cqw,32px);border-radius:24px;background:#fff;text-align:center"
					>
						<span
							style="display:inline-grid;place-items:center;width:56px;height:56px;border-radius:50%;background:#E7F6EE;color:#14894B"
							><Icon icon={I.phCheckBold} size={26} /></span
						>
						<h1
							style="margin:16px 0 0;font-family:'Sora',sans-serif;font-weight:700;font-size:clamp(32px,4.6cqw,52px);line-height:1;letter-spacing:-.05em;overflow-wrap:anywhere"
						>
							¡Gracias, {pedido.nombre}!
						</h1>
						<p style="margin:12px 0 0;font-size:15.5px;line-height:1.55;color:#69717E">
							Recibimos tu pedido <b style="color:#0E1116">#{pedido.numero}</b> por
							<b style="color:#0E1116">{fmt(pedido.total)}</b>. Te escribimos por mail con cada
							novedad.
						</p>
						<SenaPedido {tema} />
						{#if pedido.pago === 'transferencia'}
							<div
								style="margin-top:18px;padding:14px 16px;border-radius:16px;background:#EAF0FF;text-align:left;font-size:14.5px;line-height:1.6;display:flex;gap:10px"
							>
								<Icon
									icon={I.phBankFill}
									size={22}
									style="color:var(--acento, #2F6BFF);flex:none"
								/>
								<span
									><b>Transferí {fmt(pedido.total)}</b>
									{#if pedido.alias}al {pedido.esCbu ? 'CBU/CVU' : 'alias'}
										<b>{pedido.alias}</b>{:else}con los datos que te mandamos por mail{/if} y mandanos
									el comprobante{#if whatsapp}&nbsp;<a
											href={wa(`¡Hola! Te mando el comprobante del pedido #${pedido.numero}.`)}
											target="_blank"
											rel="noopener"
											style="color:var(--acento, #2F6BFF)">por WhatsApp</a
										>{:else}&nbsp;respondiendo el mail del pedido{/if}. Lo preparamos apenas
									se&nbsp;acredite.</span
								>
							</div>
						{/if}
						<a
							href={base || '/'}
							style="display:inline-block;margin-top:22px;text-decoration:none;padding:15px 24px;border-radius:14px;background:#0E1116;color:#fff;font:700 14.5px 'Instrument Sans'"
							>Volver a la tienda</a
						>
					</div>
					{#if pedido.estado === 'Entregado'}
						<div style="margin-top:16px">
							<FormResena slug={tienda.slug} {tema} codigo={pedido.codigo} nombre={pedido.nombre} />
						</div>
					{/if}
				</section>
			{/if}

			<!-- ============ SUSCRIPCIÓN (un plan y su alta; no está en el diseño: con las piezas del checkout) ============ -->
			{#if pagina === 'suscripcion' && plan && alta && caja}
				<form
					method="POST"
					use:alta.form
					style="padding:22px clamp(14px,3cqw,36px) clamp(48px,7cqw,90px);display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,360px),1fr));gap:clamp(14px,3cqw,36px);align-items:start"
				>
					<input type="hidden" name="entrega" value={alta.entrega} />
					<div style="display:flex;flex-direction:column;gap:14px;min-width:0">
						<div>
							<div style="font-size:12.5px;font-weight:700;color:var(--acento, #2F6BFF)">
								Suscripción · {CADA[plan.frecuencia]}
							</div>
							<h1
								style="margin:6px 0 0;font-family:'Sora',sans-serif;font-weight:700;font-size:clamp(32px,4.6cqw,52px);line-height:1;letter-spacing:-.05em;overflow-wrap:anywhere"
							>
								{plan.nombre}
							</h1>
							{#if plan.descripcion}<p
									style="margin:10px 0 0;font-size:15.5px;line-height:1.55;color:#69717E;text-wrap:pretty"
								>
									{plan.descripcion}
								</p>{/if}
						</div>
						{#if planes.length > 1}<div
								style="padding:16px;border-radius:20px;background:#fff;border:2px solid transparent"
							>
								<div style="display:flex;align-items:center;gap:10px">
									<span
										style="width:26px;height:26px;border-radius:50%;background:var(--acento, #2F6BFF);color:#fff;display:grid;place-items:center;font-size:12.5px;font-weight:700"
										>1</span
									><b style="flex:1;font-size:15.5px">Elegí tu plan</b>
								</div>
								<div style="margin-top:12px;display:flex;flex-direction:column;gap:8px">
									{#each planes as p (p.id)}{@const on = p.id === plan.id}<a
											href="{base}/suscripciones/{p.id}"
											aria-current={on ? 'page' : undefined}
											style="display:flex;align-items:center;gap:12px;padding:13px;border-radius:14px;border:1.5px solid {ch(
												on
											).bd};background:{ch(on)
												.bg};text-align:left;cursor:pointer;font-family:inherit;color:#0E1116;text-decoration:none"
											><span
												style="width:18px;height:18px;border-radius:50%;border:2px solid {ch(on)
													.rc};display:grid;place-items:center;flex:none"
												>{#if on}<span
														style="width:8px;height:8px;border-radius:50%;background:var(--acento, #2F6BFF)"
													></span>{/if}</span
											><span style="flex:1;min-width:0"
												><span style="display:block;font-size:14.5px;font-weight:600"
													>{p.nombre}</span
												><span style="display:block;font-size:12.5px;color:#69717E"
													>{CADA[p.frecuencia]}</span
												></span
											><b style="font-size:13.5px;white-space:nowrap">{fmt(p.precio)}</b></a
										>{/each}
								</div>
							</div>{/if}
						<div
							style="padding:16px;border-radius:20px;background:#fff;border:2px solid transparent"
						>
							<div style="display:flex;align-items:center;gap:10px">
								<span
									style="width:26px;height:26px;border-radius:50%;background:var(--acento, #2F6BFF);color:#fff;display:grid;place-items:center;font-size:12.5px;font-weight:700"
									>{planes.length > 1 ? '2' : '1'}</span
								><b style="flex:1;font-size:15.5px">Tus datos</b>
							</div>
							<div
								style="margin-top:12px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,200px),1fr));gap:8px"
							>
								<label style="display:block;min-width:0"
									><span style="display:block;font-size:12.5px;color:#69717E;margin-bottom:4px"
										>Email</span
									><input
										name="email"
										type="email"
										autocomplete="email"
										required
										maxlength="120"
										style="width:100%;padding:10px 12px;border:0;border-radius:10px;background:#F3F5F8;font:400 14px 'Instrument Sans';color:#0E1116"
									/></label
								><label style="display:block;min-width:0"
									><span style="display:block;font-size:12.5px;color:#69717E;margin-bottom:4px"
										>Nombre y apellido</span
									><input
										name="nombre"
										type="text"
										autocomplete="name"
										required
										maxlength="80"
										style="width:100%;padding:10px 12px;border:0;border-radius:10px;background:#F3F5F8;font:400 14px 'Instrument Sans';color:#0E1116"
									/></label
								><label style="display:block;min-width:0"
									><span style="display:block;font-size:12.5px;color:#69717E;margin-bottom:4px"
										>WhatsApp</span
									><input
										name="telefono"
										type="tel"
										autocomplete="tel"
										required
										maxlength="20"
										style="width:100%;padding:10px 12px;border:0;border-radius:10px;background:#F3F5F8;font:400 14px 'Instrument Sans';color:#0E1116"
									/></label
								>
							</div>
						</div>
						<div
							style="padding:16px;border-radius:20px;background:#fff;border:2px solid transparent"
						>
							<div style="display:flex;align-items:center;gap:10px">
								<span
									style="width:26px;height:26px;border-radius:50%;background:var(--acento, #2F6BFF);color:#fff;display:grid;place-items:center;font-size:12.5px;font-weight:700"
									>{planes.length > 1 ? '3' : '2'}</span
								><b style="flex:1;font-size:15.5px">Cómo te llega</b>
							</div>
							<div
								style="margin-top:12px;display:flex;flex-direction:column;gap:8px"
								role="radiogroup"
								aria-label="Cómo te llega"
							>
								{#each [['envio', 'Envío', 'A tu casa, incluido en el precio', plan.envio], ['retiro', 'Retiro', tienda.envios.retiroDireccion || 'En el local', plan.retiro]] as const as [id, t, d, hay] (id)}{#if hay}{@const on =
											alta.entrega === id}<button
											type="button"
											role="radio"
											aria-checked={on}
											onclick={() => (alta.entrega = id)}
											style="display:flex;align-items:center;gap:12px;padding:13px;border-radius:14px;border:1.5px solid {ch(
												on
											).bd};background:{ch(on)
												.bg};text-align:left;cursor:pointer;font-family:inherit;color:#0E1116;text-decoration:none"
											><span
												style="width:18px;height:18px;border-radius:50%;border:2px solid {ch(on)
													.rc};display:grid;place-items:center;flex:none"
												>{#if on}<span
														style="width:8px;height:8px;border-radius:50%;background:var(--acento, #2F6BFF)"
													></span>{/if}</span
											><span style="flex:1;min-width:0"
												><span style="display:block;font-size:14.5px;font-weight:600">{t}</span
												><span style="display:block;font-size:12.5px;color:#69717E">{d}</span></span
											></button
										>{/if}{/each}
							</div>
							{#if alta.entrega === 'envio'}<div
									style="margin-top:10px;display:grid;grid-template-columns:minmax(0,140px) minmax(0,1fr);gap:8px"
								>
									<label style="display:block;min-width:0"
										><span style="display:block;font-size:12.5px;color:#69717E;margin-bottom:4px"
											>Código postal</span
										><input
											name="cp"
											type="text"
											autocomplete="postal-code"
											inputmode="numeric"
											maxlength="8"
											style="width:100%;padding:10px 12px;border:0;border-radius:10px;background:#F3F5F8;font:400 14px 'Instrument Sans';color:#0E1116"
										/></label
									><label style="display:block;min-width:0"
										><span style="display:block;font-size:12.5px;color:#69717E;margin-bottom:4px"
											>Dirección (calle, número, piso)</span
										><input
											name="direccion"
											type="text"
											autocomplete="street-address"
											required
											maxlength="160"
											style="width:100%;padding:10px 12px;border:0;border-radius:10px;background:#F3F5F8;font:400 14px 'Instrument Sans';color:#0E1116"
										/></label
									>
								</div>{/if}
						</div>
						{#if error}<div
								role="alert"
								style="padding:12px 14px;border-radius:14px;background:#FDECEC;color:#B42318;font-size:14px"
							>
								{error}
							</div>{/if}
					</div>
					<div
						style="position:sticky;top:80px;padding:20px;border-radius:24px;background:#0E1116;color:#fff;min-width:0"
					>
						<div
							style="font-family:'Sora',sans-serif;font-weight:700;font-size:22px;letter-spacing:-.04em"
						>
							{plan.nombre}
						</div>
						{#if queIncluye(plan)}<div style="margin-top:6px;font-size:13.5px;color:#8A919D">
								{queIncluye(plan)}
							</div>{/if}
						{#if caja.fotos.length}<div style="margin-top:14px;display:flex;gap:8px;flex-wrap:wrap">
								{#each caja.fotos.slice(0, 6) as f, i (i)}<img
										src={img(f, 120, 140)}
										alt=""
										style="width:52px;height:52px;border-radius:12px;object-fit:cover;background:#1D2330"
									/>{/each}
							</div>{/if}
						<div
							style="margin-top:16px;padding-top:14px;border-top:1px solid rgba(255,255,255,.1);display:flex;flex-direction:column;gap:8px;font-size:14px"
						>
							{#if caja.ahorro}<span style="display:flex;justify-content:space-between;gap:8px"
									><span
										>{caja.unidades} {caja.unidades === 1 ? 'producto' : 'productos'} sueltos</span
									><s style="color:#8A919D">{fmt(caja.sueltos)}</s></span
								><span
									style="display:flex;justify-content:space-between;gap:8px;color:#7EE2A8;font-weight:600"
									><span>Suscripto ahorrás {caja.pct}%</span><span>− {fmt(caja.ahorro)}</span></span
								>{/if}<span style="display:flex;justify-content:space-between;gap:8px"
								><span>{alta.entrega === 'envio' ? 'Envío' : 'Retiro'}</span><span
									style="color:#7EE2A8;font-weight:600"
									>{alta.entrega === 'envio' ? 'Incluido' : 'En el local'}</span
								></span
							>
							<span
								style="display:flex;justify-content:space-between;align-items:baseline;gap:8px;margin-top:6px"
								><b style="">{porPeriodo(plan.frecuencia)}</b><b
									style="font-family:'Sora',sans-serif;font-size:26px">{fmt(plan.precio)}</b
								></span
							>
						</div>
						<button
							disabled={alta.enviando || tienda.vitrina || plan.agotado}
							style="margin-top:16px;width:100%;padding:17px;border:0;border-radius:14px;background:var(--acento, #2F6BFF);color:#fff;font:700 15.5px 'Instrument Sans';cursor:pointer"
							>{tienda.vitrina
								? 'Es una tienda de ejemplo'
								: plan.agotado
									? 'Sin lugar por ahora'
									: 'Suscribirme'}</button
						>
						<div
							style="margin-top:14px;display:flex;flex-direction:column;gap:6px;font-size:12.5px;color:#B4BCCB"
						>
							<span style="display:flex;gap:8px"
								><Icon icon={I.phCheck} size={16} style="flex:none" />{demo
									? 'Es la demo: no se cobra nada'
									: 'Débito automático con Mercado Pago'}</span
							>
							<span style="display:flex;gap:8px"
								><Icon icon={I.phCheck} size={16} style="flex:none" />Pausás o cancelás
								cuando&nbsp;quieras</span
							>
							<span style="display:flex;gap:8px"
								><Icon icon={I.phCheck} size={16} style="flex:none" />{entregasDe(plan)}</span
							>
						</div>
					</div>
				</form>
			{/if}

			<!-- ============ TU SUSCRIPCIÓN (el link privado; no está en el diseño: con las piezas de la confirmación) ============ -->
			{#if pagina === 'mi-suscripcion' && suscripcion}
				{@const tx = textosSuscripcion(suscripcion, aviso === 'alta', demo)}
				{@const puede = accionesDe(suscripcion)}
				{@const icono =
					suscripcion.estado === 'cancelada'
						? I.phXBold
						: suscripcion.estado === 'pausada'
							? I.phPauseFill
							: I.phCheckBold}
				<section style="padding:clamp(28px,4cqw,56px) clamp(14px,3cqw,36px) clamp(48px,7cqw,90px)">
					<div
						style="max-width:640px;margin:0 auto;padding:clamp(20px,3cqw,32px);border-radius:24px;background:#fff;text-align:center"
					>
						<span
							style="display:inline-grid;place-items:center;width:56px;height:56px;border-radius:50%;background:#E7F6EE;color:#14894B"
							><Icon icon={icono} size={26} /></span
						>
						<h1
							style="margin:16px 0 0;font-family:'Sora',sans-serif;font-weight:700;font-size:clamp(32px,4.6cqw,52px);line-height:1;letter-spacing:-.05em;overflow-wrap:anywhere"
						>
							{tx.titulo}
						</h1>
						<p
							style="margin:12px 0 0;font-size:15.5px;line-height:1.55;color:#69717E;text-wrap:pretty"
						>
							{tx.texto}
						</p>
						<div
							style="margin-top:18px;padding:14px 16px;border-radius:16px;background:#EAF0FF;text-align:left;font-size:14.5px;line-height:1.6"
						>
							<div style="font-size:16px;font-weight:600">{suscripcion.plan}</div>
							<div style="font-size:13.5px;color:#69717E;margin-top:2px">
								{fmt(suscripcion.precio)}
								{CADA[suscripcion.frecuencia]}{suscripcion.cobros
									? ` · ${suscripcion.cobros} ${suscripcion.cobros === 1 ? 'cobro' : 'cobros'}`
									: ''}
							</div>
							{#if suscripcion.incluye}<div style="font-size:13.5px;color:#69717E;margin-top:2px">
									{suscripcion.incluye}
								</div>{/if}
							{#if suscripcion.estado === 'activa' && suscripcion.proximoCobro}<div
									style="margin-top:12px;display:flex;gap:8px;font-size:14px"
								>
									<Icon icon={I.phCalendarPlusBold} size={17} style="flex:none" />Próximo cobro: {diaCobro(
										suscripcion.proximoCobro
									)}
								</div>{/if}
							<div style="margin-top:12px;display:flex;gap:8px;font-size:14px">
								<Icon icon={I.phMapPinFill} size={17} style="flex:none" />{suscripcion.entrega ===
								'envio'
									? `Te llega a ${suscripcion.direccion}`
									: `Retirás en ${suscripcion.tienda}`}
							</div>
						</div>
						{#if error}<div
								role="alert"
								style="margin-top:12px;padding:12px 14px;border-radius:14px;background:#FDECEC;color:#B42318;font-size:14px"
							>
								{error}
							</div>{/if}
						<div
							style="margin-top:22px;display:flex;justify-content:center;gap:10px;flex-wrap:wrap"
						>
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
										style="display:inline-block;padding:15px 24px;border:0;border-radius:14px;background:#0E1116;color:#fff;font:700 14.5px 'Instrument Sans';cursor:pointer"
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
											style="display:inline-block;padding:15px 24px;border:0;border-radius:14px;background:#0E1116;color:#fff;font:700 14.5px 'Instrument Sans';cursor:pointer"
											>Sí, darla de baja</button
										>{:else}<button
											type="button"
											onclick={() => (seguroBaja = true)}
											style="display:inline-block;padding:15px 24px;border:0;border-radius:14px;background:#fff;box-shadow:inset 0 0 0 1px rgba(14,17,22,.08);color:#0E1116;font:700 14.5px 'Instrument Sans';cursor:pointer"
											>Dar de baja</button
										>{/if}
								</form>{:else}<a
									href="{base}/suscripciones/{suscripcion.planId}"
									style="display:inline-block;padding:15px 24px;border:0;border-radius:14px;background:#0E1116;color:#fff;font:700 14.5px 'Instrument Sans';text-decoration:none"
									>Volver a suscribirme</a
								>{/if}
							<a
								href={base || '/'}
								style="display:inline-block;padding:15px 24px;border:0;border-radius:14px;background:#fff;box-shadow:inset 0 0 0 1px rgba(14,17,22,.08);color:#0E1116;font:700 14.5px 'Instrument Sans';text-decoration:none"
								>Volver a la tienda</a
							>
						</div>
						{#if whatsapp}<p style="margin:14px 0 0;font-size:13.5px;color:#69717E">
								¿Dudas? Escribinos&nbsp;<a
									href={wa(`¡Hola! Tengo una consulta sobre mi suscripción a ${suscripcion.plan}.`)}
									target="_blank"
									rel="noopener"
									style="color:#0E1116">por WhatsApp</a
								>.
							</p>{/if}
					</div>
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
				<section style="padding:clamp(28px,4cqw,56px) clamp(14px,3cqw,36px) 0">
					{#if tienda.ciudad}<div
							style="font-size:12.5px;font-weight:700;color:var(--acento, #2F6BFF)"
						>
							{tienda.ciudad}
						</div>{/if}
					<h1
						style="margin:8px 0 0;max-width:900px;font-family:'Sora',sans-serif;font-weight:700;font-size:clamp(40px,6.4cqw,88px);line-height:.95;letter-spacing:-.06em;overflow-wrap:anywhere"
					>
						{tienda.portada.titulo}
					</h1>
					{#if stats.length}<div
							style="margin-top:24px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,200px),1fr));gap:10px"
						>
							{#each stats as s (s.t)}<div
									style="padding:18px;border-radius:20px;background:{s.bg};color:{s.fg}"
								>
									<div
										style="font-family:'Sora',sans-serif;font-weight:700;font-size:clamp(30px,3.6cqw,44px);letter-spacing:-.05em"
									>
										{s.v}
									</div>
									<div style="font-size:13.5px;margin-top:4px;opacity:.75">{s.t}</div>
								</div>{/each}
						</div>{/if}
				</section>
				<section
					style="padding:12px clamp(14px,3cqw,36px) 0;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,340px),1fr));gap:10px"
				>
					{#if tienda.portada.foto}<div
							style="position:relative;min-height:360px;border-radius:24px;overflow:hidden;background:#EEF1F5"
						>
							<img
								src={img(tienda.portada.foto, 1000)}
								alt=""
								style="position:absolute;inset:0;width:100%;height:100%;object-fit:cover"
							/>
						</div>{/if}
					<div
						style="padding:clamp(20px,3cqw,32px);border-radius:24px;background:#fff;display:flex;flex-direction:column;gap:14px"
					>
						<p style="margin:0;font-size:17px;line-height:1.6">
							{tienda.nombre}{tienda.ciudad ? `, desde ${tienda.ciudad}` : ''}. {comoLlega(
								tienda.envios
							)}
						</p>
						<div style="display:flex;flex-direction:column;gap:10px;font-size:14.5px">
							{#if tienda.direccion}<span style="display:flex;gap:10px"
									><Icon
										icon={I.phMapPinFill}
										size={18}
										style="color:var(--acento, #2F6BFF);flex:none"
									/>{tienda.direccion}</span
								>{/if}{#if tienda.cobros.cuotas > 1}<span style="display:flex;gap:10px"
									><Icon
										icon={I.phCreditCard}
										size={18}
										style="color:var(--acento, #2F6BFF);flex:none"
									/>{tienda.cobros.cuotas} cuotas sin interés{#if tienda.cobros.transferencia > 0}&nbsp;·
										{tienda.cobros.transferencia}% off por transferencia{/if}</span
								>{:else if tienda.cobros.transferencia > 0}<span style="display:flex;gap:10px"
									><Icon
										icon={I.phBank}
										size={18}
										style="color:var(--acento, #2F6BFF);flex:none"
									/>{tienda.cobros.transferencia}% off por transferencia</span
								>{/if}{#if tienda.envios.gratisDesde}<span style="display:flex;gap:10px"
									><Icon
										icon={I.phTruck}
										size={18}
										style="color:var(--acento, #2F6BFF);flex:none"
									/>Envío gratis desde {fmt(tienda.envios.gratisDesde)}</span
								>{/if}
						</div>
					</div>
				</section>
				{#if whatsapp}
					<section
						style="padding:12px clamp(14px,3cqw,36px) clamp(48px,7cqw,90px);display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,260px),1fr));gap:10px"
					>
						<a
							href={wa('¡Hola!')}
							target="_blank"
							rel="noopener"
							style="padding:18px;border-radius:20px;background:#25D366;color:#fff;display:flex;gap:12px;align-items:center;text-decoration:none"
							><Icon icon={I.phWhatsappLogoFill} size={26} />
							<div>
								<div style="font-weight:700">WhatsApp</div>
								<div style="font-size:13.5px;opacity:.8">{tienda.whatsapp}</div>
							</div></a
						>
					</section>
				{:else}
					<div style="height:clamp(48px,7cqw,90px)"></div>
				{/if}
			{/if}
		{/if}

		<!-- footer -->
		<footer
			style="padding:clamp(36px,5cqw,64px) clamp(14px,3cqw,36px) 24px;background:#0E1116;color:#fff"
		>
			<div
				style="display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,180px),1fr));gap:26px"
			>
				<div style="min-width:0;max-width:340px">
					<div
						style="font-family:'Sora',sans-serif;font-weight:700;font-size:24px;letter-spacing:-.04em;overflow-wrap:anywhere"
					>
						{tienda.nombre}
					</div>
					<p style="margin:10px 0 0;font-size:14px;color:#8A919D;line-height:1.55">
						{#if tienda.cobros.cuotas > 1}{tienda.cobros.cuotas} cuotas sin interés{#if tienda.cobros.transferencia > 0}&nbsp;y
								{tienda.cobros.transferencia}% off por transferencia{/if}, en toda la&nbsp;tienda.{:else if tienda.cobros.transferencia > 0}{tienda
								.cobros.transferencia}% off por transferencia, en toda la&nbsp;tienda.{:else}Compra
							protegida, con 10 días para&nbsp;arrepentirte.{/if}
					</p>
				</div>
				{#each foot as f (f.t)}<div
						style="display:flex;flex-direction:column;gap:9px;font-size:14px"
					>
						<b style="font-size:12.5px;color:#9DB8FF;margin-bottom:4px">{f.t}</b
						>{#each f.l as l (l.t)}<a href={l.href} style="color:#B4BCCB;text-decoration:none"
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
				style="margin-top:32px;padding-top:14px;border-top:1px solid rgba(255,255,255,.08);display:flex;justify-content:space-between;gap:14px;flex-wrap:wrap;font-size:12.5px;color:#69717E"
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
				style="position:fixed;right:18px;bottom:18px;z-index:30;width:54px;height:54px;border-radius:18px;background:#25D366;color:#fff;display:grid;place-items:center;box-shadow:0 12px 26px -8px rgba(0,0,0,.4)"
				><Icon icon={I.phWhatsappLogoFill} size={28} /></a
			>{/if}
	</div>
</div>

<!-- eslint-enable svelte/no-navigation-without-resolve -->

<style>
	.andes footer :global(a:hover) {
		color: #fff;
	}
	.tarj:hover {
		box-shadow: 0 18px 40px -22px rgba(14, 17, 22, 0.4);
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
