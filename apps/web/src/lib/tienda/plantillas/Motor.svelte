<script lang="ts">
	import { camposFactura, OPCIONALES } from '#lib/tienda/factura.ts';
	import { envia } from '../envio.ts';
	import { enhance } from '$app/forms';
	import { goto } from '$app/navigation';
	import Icon from '#lib/components/Icon.svelte';
	import LogoTienda from '../LogoTienda.svelte';
	import VacioTienda from '../VacioTienda.svelte';
	import { enOrden } from '../secciones.ts';
	import { acentoElegido } from '../acentos.ts';
	import * as I from '#lib/icons.ts';
	import FormResena from '../FormResena.svelte';
	import CompartirProducto from '../CompartirProducto.svelte';
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
	 * Motor de plantillas (design/Plantillas Tienda.dc.html): un solo marcado con temas para doce
	 * plantillas en tres familias (editorial, grilla y catálogo), con los datos de la tienda en lugar
	 * de los de ejemplo. Lo que el diseño muestra sin datos atrás (talles, combos, contadores,
	 * cuenta regresiva, buscador) no se inventa: esas partes no se dibujan. El newsletter del pie
	 * anda de verdad (§50).
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

	// Tema: familia, tipografía, peso, espaciado, mayúsculas, itálica, fondo, tinta, acento, tinta
	// sobre acento, acento de texto, superficie, superficie 2, apagado, línea, ok, radio, radio
	// grande, radio de botón, panel, tinta del panel, acento del panel, estrellas.
	type Familia = 'ed' | 'grid' | 'cat';
	type Tema = [
		Familia,
		string,
		number,
		string,
		string,
		string,
		string,
		string,
		string,
		string,
		string,
		string,
		string,
		string,
		string,
		string,
		string,
		string,
		string,
		string,
		string,
		string,
		string
	];
	// prettier-ignore
	const TH = {
		pampa: ['grid', "'Archivo Black',sans-serif", 400, '-.01em', 'uppercase', 'normal', '#141414', '#FFFFFF', '#E6FF4D', '#141414', '#E6FF4D', '#1E1E1E', '#2A2A2A', '#A3A3A3', 'rgba(255,255,255,.12)', '#B9F27C', '4px', '8px', '4px', '#232323', '#FFFFFF', '#E6FF4D', '#E6FF4D'],
		rosedal: ['ed', "'Cormorant Garamond',serif", 600, '-.01em', 'none', 'italic', '#F8EDEB', '#3A1F25', '#C26A7A', '#FFFFFF', '#A84F60', '#FFF7F5', '#F0DCD8', '#7E5F65', 'rgba(58,31,37,.12)', '#3E7A55', '14px', '22px', '999px', '#3A1F25', '#F8EDEB', '#F0A7B4', '#C26A7A'],
		quebracho: ['ed', "'DM Serif Display',serif", 400, '-.01em', 'none', 'italic', '#EFE9E1', '#2B2520', '#8A5A3E', '#FFFFFF', '#8A5A3E', '#F7F3EE', '#E2D8CC', '#6F6358', 'rgba(43,37,32,.12)', '#3E6E4E', '3px', '4px', '3px', '#2B2520', '#EFE9E1', '#D9A57E', '#2B2520'],
		andes: ['grid', "'Space Grotesk',sans-serif", 700, '-.045em', 'none', 'normal', '#F2F4F7', '#0F1720', '#2563EB', '#FFFFFF', '#1D4ED8', '#FFFFFF', '#E3E8EF', '#55606F', 'rgba(15,23,32,.1)', '#15803D', '12px', '20px', '12px', '#0F1720', '#F2F4F7', '#7AA7FF', '#F59E0B'],
		malbec: ['ed', "'Cormorant Garamond',serif", 500, '-.01em', 'none', 'italic', '#1E0E12', '#F4E9E1', '#C24A5A', '#FFFFFF', '#E8909B', '#2A161B', '#3A2026', '#BFA9A3', 'rgba(244,233,225,.12)', '#9FD3A8', '2px', '2px', '2px', '#F4E9E1', '#1E0E12', '#9E2F40', '#E8B86A'],
		patio: ['grid', "'Fredoka',sans-serif", 700, '-.02em', 'none', 'normal', '#FFF7E8', '#2A1E10', '#F08A24', '#2A1E10', '#B5560A', '#FFFFFF', '#FBE6C2', '#75644F', 'rgba(42,30,16,.1)', '#2F7A43', '18px', '28px', '999px', '#2A1E10', '#FFF7E8', '#FFB566', '#F08A24'],
		miga: ['cat', "'DM Serif Display',serif", 400, '-.01em', 'none', 'italic', '#FBF1E6', '#3A2414', '#B5651D', '#FFFFFF', '#9A5216', '#FFFFFF', '#F1DFC9', '#7A5F4C', 'rgba(58,36,20,.1)', '#3F7A4A', '12px', '18px', '999px', '#3A2414', '#FBF1E6', '#F2B36E', '#D98A2B'],
		taller: ['ed', "'Bricolage Grotesque',sans-serif", 800, '-.045em', 'none', 'normal', '#F3EEE6', '#2A231C', '#6B4A2F', '#FFFFFF', '#6B4A2F', '#FAF7F2', '#E5DCCF', '#71665A', 'rgba(42,35,28,.12)', '#3E6E4E', '8px', '10px', '8px', '#2A231C', '#F3EEE6', '#D7B48C', '#2A231C'],
		hora: ['grid', "'Bodoni Moda',serif", 500, '-.02em', 'none', 'italic', '#F5F5F2', '#161616', '#B08D57', '#161616', '#8A6A36', '#FFFFFF', '#E7E6E1', '#66665F', 'rgba(22,22,22,.12)', '#2F6B47', '0px', '0px', '0px', '#161616', '#F5F5F2', '#D4B27A', '#B08D57'],
		litoral: ['cat', "'Bricolage Grotesque',sans-serif", 800, '-.04em', 'none', 'normal', '#F3E9DC', '#3B2416', '#6B3A22', '#FFFFFF', '#6B3A22', '#FFF9F1', '#E7D6BF', '#78604C', 'rgba(59,36,22,.12)', '#3F7A4A', '12px', '18px', '999px', '#3B2416', '#F3E9DC', '#E9B36F', '#D98A2B'],
		torno: ['ed', "'Instrument Serif',serif", 400, '-.015em', 'none', 'italic', '#EDE6DD', '#2C2621', '#A4623D', '#FFFFFF', '#A4623D', '#F6F1EB', '#DED3C6', '#73685E', 'rgba(44,38,33,.12)', '#3E6E4E', '18px', '28px', '999px', '#2C2621', '#EDE6DD', '#E3A27E', '#2C2621'],
		lente: ['grid', "'Unbounded',sans-serif", 700, '-.04em', 'none', 'normal', '#FFF8EC', '#1F1A12', '#E0A31B', '#1F1A12', '#946400', '#FFFFFF', '#FAE6BC', '#736650', 'rgba(31,26,18,.1)', '#2F7A43', '12px', '22px', '999px', '#1F1A12', '#FFF8EC', '#F5C451', '#E0A31B']
	} satisfies Record<string, Tema>;
	type Id = keyof typeof TH;

	// Sólo la tipografía del tema elegido (Instrument Sans e Instrument Serif ya las carga el layout).
	const FUENTES: Record<Id, string | null> = {
		pampa: 'Archivo+Black',
		rosedal: 'Cormorant+Garamond:ital,wght@0,400;0,500;0,600;1,400;1,500;1,600',
		quebracho: 'DM+Serif+Display:ital@0;1',
		andes: 'Space+Grotesk:wght@400;700',
		malbec: 'Cormorant+Garamond:ital,wght@0,400;0,500;0,600;1,400;1,500;1,600',
		patio: 'Fredoka:wght@400;700',
		miga: 'DM+Serif+Display:ital@0;1',
		taller: 'Bricolage+Grotesque:wght@400;800',
		hora: 'Bodoni+Moda:ital,wght@0,400;0,500;1,400;1,500',
		litoral: 'Bricolage+Grotesque:wght@400;800',
		torno: null,
		lente: 'Unbounded:wght@400;700'
	};

	const tid = $derived.by((): Id => {
		const id = tienda.plantilla.toLowerCase();
		return id in TH ? (id as Id) : 'litoral';
	});
	const fuente = $derived(FUENTES[tid]);
	const nombrePlantilla = $derived(tid.charAt(0).toUpperCase() + tid.slice(1));
	const T = $derived.by(() => {
		const [
			,
			f,
			fw,
			ls,
			tt,
			fi,
			bg,
			ink,
			acc,
			ai,
			at,
			sf,
			sf2,
			mu,
			ln,
			ok,
			r,
			rL,
			rb,
			panel,
			panelInk,
			panelAcc,
			star
		] = TH[tid] as Tema;
		const fam = TH[tid][0] as Familia;
		// El acento elegido en Mi tienda; como texto sólo si va con texto claro encima (si no, no se lee).
		const elegido = acentoElegido(tid, tienda.paleta);
		const dark = tid === 'pampa' || tid === 'malbec';
		return {
			f,
			fw,
			ls,
			tt,
			fi,
			bg,
			ink,
			acc: elegido ?? acc,
			ai,
			at: elegido && ai.toUpperCase() === '#FFFFFF' ? elegido : at,
			sf,
			sf2,
			mu,
			ln,
			ok,
			r,
			rL,
			rb,
			panel,
			panelInk,
			panelAcc,
			star,
			f2: fam === 'ed' ? f : "'Instrument Sans',sans-serif",
			logoSz: tt === 'uppercase' ? '22px' : '26px',
			hb: dark ? 'rgba(20,14,14,.88)' : 'rgba(255,255,255,.72)'
		};
	});
	// Reseñas y preguntas con los tokens del tema elegido.
	const tema = $derived<TemaOpiniones>({
		texto: T.ink,
		suave: T.mu,
		campo: T.bg,
		tarjeta: T.sf,
		borde: T.ln,
		acento: T.ink,
		sobreAcento: T.bg,
		estrella: T.star,
		radio: T.r,
		radioBoton: T.rb,
		fuente: "'Instrument Sans', sans-serif",
		titulos: T.f,
		pesoTitulos: Number(T.fw) || 600,
		mayusculas: T.tt === 'uppercase',
		error: tid === 'pampa' || tid === 'malbec' ? '#FF9C8A' : undefined
	});
	const fam = $derived(TH[tid][0] as Familia);
	const isEd = $derived(fam === 'ed');
	const isGrid = $derived(fam === 'grid');
	const isCat = $derived(fam === 'cat');

	// En el dominio propio de la tienda los links no llevan /t/<slug> (tienda.base = '').
	const base = $derived(tienda.base ?? `/t/${tienda.slug}`);
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
	const off = (p: ProductoPublico) =>
		p.anterior && p.anterior > p.precio ? Math.round((1 - p.precio / p.anterior) * 100) : 0;
	// Con variantes de precios distintos, "Desde" (sin tachado ni descuento); "Sin stock" sólo si no
	// queda ninguna variante.
	const tarjeta = (p: ProductoPublico) => {
		const t = tarjetaDe(p);
		const o = t.desde ? 0 : off(p);
		const tag = t.agotado ? 'Sin stock' : o ? `−${o}%` : p.etiqueta;
		return {
			id: p.id,
			n: p.nombre,
			price: (t.desde ? 'Desde ' : '') + fmt(t.precio),
			old: p.anterior && !t.desde ? fmt(p.anterior) : null,
			img: img(p.foto, 640, isEd ? 800 : 640),
			tag,
			tagBg: o ? T.acc : T.ink,
			tagFg: o ? T.ai : T.bg,
			cuotas:
				tienda.cobros.cuotas > 1
					? `${tienda.cobros.cuotas} × ${fmt(t.precio / tienda.cobros.cuotas)} sin interés`
					: '',
			transf: fmt(t.precio * (1 - tienda.cobros.transferencia / 100)),
			agotado: t.agotado,
			href: `${base}/producto/${p.slug ?? p.id}`
		};
	};
	type Tarjeta = ReturnType<typeof tarjeta>;
	const ver = (s: string) => tienda.secciones.includes(s as TiendaPublica['secciones'][number]);
	const whatsapp = $derived(tienda.whatsapp.replace(/\D/g, ''));
	const wa = (texto: string) =>
		`https://wa.me/${whatsapp.startsWith('54') ? whatsapp : '54' + whatsapp}?text=${encodeURIComponent(texto)}`;

	// Título de la portada partido en dos, como en el diseño ("Todo para / tu compañero.").
	const titulo = $derived(partirTitulo(tienda.portada.titulo));

	// Medidas que cambian con la familia (renderVals del diseño).
	const catSpan = $derived(nar ? 'auto' : 'span 2');
	const perkMt = $derived(isCat ? '18px' : '0px');
	const catAr = $derived(isEd ? '3/4' : isGrid ? '4/3' : '16/10');
	const cardMin = $derived(isCat ? '200px' : '240px');
	const cardMinCat = $derived(isCat ? '180px' : '220px');
	const cardPad = $derived(isEd ? '0px' : '10px');
	const cardBg = $derived(isEd ? 'transparent' : T.sf);
	const cardAr = $derived(isEd ? '4/5' : '1/1');
	const cardTxtPad = $derived(isEd ? '12px 0 0' : '12px 6px 4px');
	const cardBtn = $derived(isCat);
	const cardBtnM = '4px 4px 2px';
	const sideF = $derived(!nar && !isGrid);

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
	const rail = $derived(tienda.categorias);
	const best = $derived(
		tienda.masVendidos.slice(0, 4).flatMap((id) => (porId.get(id) ? [tarjeta(porId.get(id)!)] : []))
	);
	const heroImg2 = $derived(
		img(
			tienda.fotos[0] ??
				porId.get(tienda.masVendidos[1] ?? '')?.foto ??
				tienda.categorias[0]?.foto ??
				null,
			600,
			600
		)
	);
	// Oferta del día (familia catálogo): la de mayor descuento con stock.
	const deal = $derived.by(() => {
		const conOferta = tienda.productos.filter((p) => off(p) > 0 && !tarjetaDe(p).agotado);
		if (!conOferta.length) return null;
		return tarjeta(conOferta.reduce((a, b) => (off(b) > off(a) ? b : a)));
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
	const precios = $derived(listado.map((p) => p.precio));
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
						'Tenés 10 días desde que lo recibís para arrepentirte de la compra, sin dar explicaciones (botón de arrepentimiento, abajo de\u00a0todo).'
					]
				].map(([t, d]) => ({ t, d, open: abierto === t, i: abierto === t ? I.phMinus : I.phPlus }))
			: []
	);
	const pd = $derived(producto ? tarjeta(producto) : null);
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
	// Talle, color, presentación…: lo elegido, la variante, su precio, stock y foto (seleccion.svelte.ts).
	const sel = $derived(producto ? new Seleccion(producto) : null);
	// El tachado no va con "Desde": no se sabe con qué variante se compara.
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
	/** Sumar desde una tarjeta o "Te puede gustar": con opciones, primero se elige en la ficha. */
	const sumar = (id: string) => {
		const p = porId.get(id);
		if (!p || tienda.vitrina) return;
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- ruta de la tienda con su slug
		if (!carrito.sumarProducto(p)) goto(`${base}/producto/${p.slug ?? p.id}`);
	};

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
	// Envío estimado en el resumen del carrito: el más barato a domicilio para ese código postal.
	let cpCarrito = $state('');
	const envioCarrito = $derived.by(() => {
		if (!cpValido(cpCarrito)) return null;
		const o = cotizar(cpCarrito, subtotal, tienda).filter((x) => x.id !== 'retiro');
		return o.length ? Math.min(...o.map((x) => x.precio)) : -1;
	});

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
						c: T.ok
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
	const CAMPOS = [
		['email', 'Email', 'email', 'email'],
		['nombre', 'Nombre y apellido', 'text', 'name'],
		['telefono', 'Teléfono', 'tel', 'tel'],
		['dni', 'DNI', 'text', 'off']
	] as const;
	const campos = $derived([...CAMPOS, ...camposFactura(tienda)]);

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

<svelte:head>
	{#if fuente}<link
			href="https://fonts.googleapis.com/css2?family={fuente}&display=swap"
			rel="stylesheet"
		/>{/if}
</svelte:head>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
{#snippet card(p: Tarjeta, zoom: boolean, transf: boolean)}
	<div
		style="padding:{cardPad};border-radius:{T.rL};background:{cardBg};text-align:left;color:{T.ink};display:flex;flex-direction:column"
	>
		<a
			href={p.href}
			class={zoom ? 'zoom' : ''}
			style="flex:1;display:flex;flex-direction:column;text-decoration:none;color:{T.ink};font-family:inherit"
		>
			<span
				style="position:relative;display:block;aspect-ratio:{cardAr};border-radius:{T.r};overflow:hidden;background:{T.sf2}"
				>{#if p.img}<img
						src={p.img}
						alt=""
						style="width:100%;height:100%;object-fit:cover;display:block"
					/>{/if}{#if p.tag}<span
						style="position:absolute;top:10px;left:10px;padding:4px 9px;border-radius:{T.rb};background:{p.tagBg};color:{p.tagFg};font-size:11px;font-weight:700;letter-spacing:.04em"
						>{p.tag}</span
					>{/if}</span
			>
			<span style="display:block;padding:{cardTxtPad}"
				><span style="display:block;font-size:14.5px;font-weight:500;line-height:1.3">{p.n}</span
				><span style="display:flex;gap:8px;align-items:baseline;margin-top:5px"
					><b style="font-size:16px;font-weight:700">{p.price}</b>{#if p.old}<s
							style="font-size:13px;color:{T.mu}">{p.old}</s
						>{/if}</span
				>{#if p.cuotas}<span style="display:block;margin-top:2px;font-size:12.5px;color:{T.mu}"
						>{p.cuotas}</span
					>{/if}{#if transf && tienda.cobros.transferencia > 0}<span
						style="display:block;font-size:12.5px;color:{T.ok};font-weight:600"
						>{p.transf} por transferencia</span
					>{/if}</span
			>
		</a>
		{#if cardBtn}<button
				type="button"
				onclick={() => sumar(p.id)}
				disabled={p.agotado || tienda.vitrina}
				aria-label="Agregar {p.n} al carrito"
				style="margin:{cardBtnM};display:flex;justify-content:center;align-items:center;gap:7px;padding:10px;border:0;border-radius:{T.rb};background:{T.ink};color:{T.bg};font:700 13.5px 'Instrument Sans';cursor:pointer"
				><Icon icon={I.phPlusBold} size={14} />Agregar</button
			>{/if}
	</div>
{/snippet}

<div
	class="motor"
	style="min-height:100vh;font-family:'Instrument Sans',system-ui,sans-serif;color:{T.ink};background:{T.bg}"
>
	<div
		style="container-type:inline-size;width:100%;background:{T.bg};color:{T.ink};position:relative"
	>
		<!-- barra promo -->
		<div
			style="display:flex;justify-content:center;align-items:center;gap:12px;flex-wrap:wrap;padding:9px 16px;background:{T.acc};color:{T.ai};font-size:13px;font-weight:600;text-align:center"
		>
			<span
				>{tienda.envios.gratisDesde
					? `Envío gratis desde ${fmt(tienda.envios.gratisDesde)}`
					: tienda.cobros.cuotas > 1
						? `${tienda.cobros.cuotas} cuotas sin interés`
						: 'Compra protegida'}</span
			>{#if wide && tienda.cobros.cuotas > 1}<span style="opacity:.85;font-weight:500"
					>{tienda.cobros.cuotas} cuotas sin interés{#if tienda.cobros.transferencia > 0}&nbsp;· {tienda
							.cobros.transferencia}% off por transferencia{/if}</span
				>{/if}
		</div>

		<!-- header -->
		<header
			style="position:sticky;top:0;z-index:20;display:flex;align-items:center;gap:18px;padding:14px clamp(14px,3.5cqw,44px);background:{T.hb};backdrop-filter:blur(10px);border-bottom:1px solid {T.ln}"
		>
			{#if nar}<a href="{base}/catalogo" aria-label="Catálogo" style="display:grid;color:{T.ink}"
					><Icon icon={I.phList} size={22} /></a
				>{/if}
			<a
				href={base || '/'}
				style="text-decoration:none;color:{T.ink};font-family:{T.f};font-weight:{T.fw};font-size:{T.logoSz};letter-spacing:{T.ls};text-transform:{T.tt};line-height:1;white-space:nowrap;min-width:0;overflow:hidden;text-overflow:ellipsis"
				><LogoTienda {tienda} /></a
			>
			{#if wide}<nav style="display:flex;gap:22px;margin-left:18px">
					{#each nav as n (n.href)}<a
							href={n.href}
							style="text-decoration:none;padding:4px 0;font:500 14px 'Instrument Sans';color:{T.ink};border-bottom:1.5px solid {n.on
								? T.ink
								: 'transparent'}">{n.t}</a
						>{/each}
				</nav>{/if}
			<span style="flex:1"></span>
			{#if nar && tienda.turnos}<a
					href="{base}/turnos"
					aria-label="Reservar turno"
					style="display:grid;color:{T.ink}"><Icon icon={I.phCalendarPlusBold} size={22} /></a
				>{/if}
			<a
				href="{base}/carrito"
				aria-label="Carrito"
				style="position:relative;color:{T.ink};display:grid"
				><Icon icon={I.phHandbag} size={22} />{#if carrito.unidades}<span
						style="position:absolute;top:-6px;right:-8px;min-width:17px;height:17px;border-radius:999px;background:{T.acc};color:{T.ai};font-size:10.5px;font-weight:700;display:grid;place-items:center"
						>{carrito.unidades}</span
					>{/if}</a
			>
		</header>

		{#if tienda.vitrina}
			<div style="padding:10px 16px;background:{T.sf2};font-size:13px;text-align:center">
				Es una tienda de ejemplo de la plantilla {nombrePlantilla}: podés recorrerla, pero no se
				compra de&nbsp;verdad.
			</div>
		{/if}

		<!-- ============ INICIO ============ -->
		{#if pagina === 'inicio'}
			<!-- Las secciones en el orden que eligió la tienda en Mi tienda. -->
			{#each enOrden( tienda.secciones, ['hero', 'cats', 'prods', 'revs', 'ig'] ) as seccion (seccion)}
				{#if seccion === 'hero'}
					{#if ver('hero')}
						{#if isEd}
							<section
								style="position:relative;height:clamp(500px,60cqw,800px);overflow:hidden;background:{T.sf2}"
							>
								{#if tienda.portada.foto}<img
										src={img(tienda.portada.foto, 1600)}
										alt=""
										style="position:absolute;inset:0;width:100%;height:100%;object-fit:cover"
									/>{/if}
								<div
									style="position:absolute;inset:0;background:linear-gradient(90deg,rgba(0,0,0,.55),rgba(0,0,0,.05) 65%)"
								></div>
								<div
									style="position:absolute;left:clamp(20px,5cqw,72px);bottom:clamp(30px,6cqw,88px);right:20px;max-width:680px;color:#fff"
								>
									{#if tienda.portada.lema}<div
											style="font-size:12.5px;letter-spacing:.24em;text-transform:uppercase"
										>
											{tienda.portada.lema}
										</div>{/if}
									<h1
										style="margin:14px 0 0;font-family:{T.f};font-weight:{T.fw};font-size:clamp(50px,9cqw,124px);line-height:.92;letter-spacing:{T.ls};text-transform:{T.tt};overflow-wrap:anywhere"
									>
										{titulo[0]}{#if titulo[1]}<br /><i style="font-style:{T.fi}">{titulo[1]}</i
											>{/if}
									</h1>
									<div style="margin-top:24px;display:flex;gap:10px;flex-wrap:wrap">
										<a
											href="{base}/catalogo"
											style="text-decoration:none;padding:15px 28px;border-radius:{T.rb};background:#fff;color:#141414;font:600 14.5px 'Instrument Sans'"
											>Ver el catálogo</a
										>{#if tienda.categorias[0]}<a
												href={enlaceCat(tienda.categorias[0].nombre)}
												style="text-decoration:none;padding:15px 22px;border:1px solid rgba(255,255,255,.6);border-radius:{T.rb};background:transparent;color:#fff;font:600 14.5px 'Instrument Sans'"
												>{tienda.categorias[0].nombre}</a
											>{/if}
									</div>
								</div>
							</section>
						{/if}
						{#if isGrid}
							<section
								style="padding:clamp(14px,2.4cqw,32px) clamp(14px,3.5cqw,44px) 0;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,380px),1fr));gap:clamp(10px,1.6cqw,20px)"
							>
								<div
									style="padding:clamp(26px,4.4cqw,60px);border-radius:{T.rL};background:{T.panel};color:{T.panelInk};display:flex;flex-direction:column;justify-content:space-between;gap:28px;min-height:min(560px,120cqw);min-width:0"
								>
									{#if tienda.portada.lema}<span
											style="align-self:flex-start;padding:7px 12px;border-radius:{T.rb};background:{T.acc};color:{T.ai};font-size:12.5px;font-weight:700;letter-spacing:.04em"
											>{tienda.portada.lema}</span
										>{:else}<span></span>{/if}
									<div style="min-width:0">
										<h1
											style="margin:0;font-family:{T.f};font-weight:{T.fw};font-size:clamp(46px,7cqw,100px);line-height:.9;letter-spacing:{T.ls};text-transform:{T.tt};overflow-wrap:anywhere"
										>
											{titulo[0]}{#if titulo[1]}<br /><span
													style="color:{T.panelAcc};font-style:{T.fi}">{titulo[1]}</span
												>{/if}
										</h1>
									</div>
									<div style="display:flex;gap:10px;flex-wrap:wrap">
										<a
											href="{base}/catalogo"
											style="text-decoration:none;padding:15px 24px;border-radius:{T.rb};background:{T.acc};color:{T.ai};font:700 15px 'Instrument Sans'"
											>Ver el catálogo</a
										>{#if tienda.categorias[0]}<a
												href={enlaceCat(tienda.categorias[0].nombre)}
												style="text-decoration:none;padding:15px 22px;border:1.5px solid currentColor;border-radius:{T.rb};background:transparent;color:inherit;font:700 15px 'Instrument Sans';opacity:.85"
												>{tienda.categorias[0].nombre}</a
											>{/if}
									</div>
								</div>
								<!-- El mosaico, sólo con algo adentro: sin fotos ni reseñas serían recuadros vacíos. -->
								{#if tienda.portada.foto || heroImg2 || best[0] || tienda.promedio}<div
										style="display:grid;grid-template-columns:1fr 1fr;grid-template-rows:1fr 1fr;gap:clamp(10px,1.6cqw,20px);min-height:min(560px,110cqw)"
									>
										<div
											style="grid-row:span 2;position:relative;border-radius:{T.rL};overflow:hidden;background:{T.sf2}"
										>
											{#if tienda.portada.foto}<img
													src={img(tienda.portada.foto, 1600)}
													alt=""
													style="position:absolute;inset:0;width:100%;height:100%;object-fit:cover"
												/>{/if}{#if best[0]}<a
													href={best[0].href}
													style="position:absolute;left:12px;bottom:12px;right:12px;width:fit-content;padding:8px 12px;border-radius:{T.r};background:{T.sf};color:{T.ink};font-size:13px;font-weight:700;text-decoration:none"
													>{best[0].n} · {best[0].price}</a
												>{/if}
										</div>
										<div
											style="position:relative;border-radius:{T.rL};overflow:hidden;background:{T.sf2};grid-row:span {tienda.promedio
												? 1
												: 2}"
										>
											{#if heroImg2}<img
													src={heroImg2}
													alt=""
													style="position:absolute;inset:0;width:100%;height:100%;object-fit:cover"
												/>{/if}
										</div>
										{#if tienda.promedio}<div
												style="border-radius:{T.rL};background:{T.acc};color:{T.ai};padding:18px;display:flex;flex-direction:column;justify-content:space-between"
											>
												<Icon icon={I.phStarFill} size={22} />
												<div>
													<div
														style="font-family:{T.f};font-weight:{T.fw};font-size:clamp(32px,4cqw,46px);line-height:1"
													>
														{tienda.promedio.toLocaleString('es-AR')}
													</div>
													<div style="font-size:13px;font-weight:600;margin-top:4px">
														{tienda.resenas.length}
														{tienda.resenas.length === 1 ? 'reseña' : 'reseñas'}
													</div>
												</div>
											</div>{/if}
									</div>{/if}
							</section>
						{/if}
						{#if isCat}
							<section
								style="padding:clamp(12px,2cqw,24px) clamp(12px,3cqw,36px) 0;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,340px),1fr));gap:10px"
							>
								<div
									style="grid-column:{deal || whatsapp
										? catSpan
										: 'auto'};position:relative;min-height:340px;border-radius:{T.rL};overflow:hidden;background:{T.sf2}"
								>
									{#if tienda.portada.foto}<img
											src={img(tienda.portada.foto, 1600)}
											alt=""
											style="position:absolute;inset:0;width:100%;height:100%;object-fit:cover"
										/>{/if}
									<div
										style="position:absolute;inset:0;background:linear-gradient(90deg,rgba(0,0,0,.6),rgba(0,0,0,0) 70%)"
									></div>
									<div
										style="position:absolute;left:clamp(18px,3.4cqw,44px);top:50%;transform:translateY(-50%);right:20px;max-width:520px;color:#fff"
									>
										{#if tienda.portada.lema}<span
												style="display:inline-block;padding:6px 11px;border-radius:{T.rb};background:{T.acc};color:{T.ai};font-size:12.5px;font-weight:700"
												>{tienda.portada.lema}</span
											>{/if}
										<h1
											style="margin:12px 0 0;font-family:{T.f};font-weight:{T.fw};font-size:clamp(38px,5.4cqw,72px);line-height:.95;letter-spacing:{T.ls};overflow-wrap:anywhere"
										>
											{titulo[1] ? `${titulo[0]} ` : titulo[0]}{#if titulo[1]}
												<span style="font-style:{T.fi}">{titulo[1]}</span>{/if}
										</h1>
										<a
											href="{base}/catalogo"
											style="display:inline-block;margin-top:18px;padding:13px 22px;border-radius:{T.rb};background:#fff;color:#141414;font:700 14.5px 'Instrument Sans';text-decoration:none"
											>Ver el catálogo</a
										>
									</div>
								</div>
								{#if deal || whatsapp}
									<div style="display:flex;flex-direction:column;gap:10px">
										{#if deal}<a
												href={deal.href}
												style="flex:1;padding:18px;border-radius:{T.rL};background:{T.panel};color:{T.panelInk};display:flex;flex-direction:column;justify-content:space-between;gap:10px;text-decoration:none"
											>
												<span style="display:flex;justify-content:space-between;align-items:center"
													><b style="font-size:15px">Oferta del día</b><Icon
														icon={I.phLightningFill}
														size={20}
														style="color:{T.panelAcc}"
													/></span
												>
												<span style="display:flex;gap:12px;align-items:center"
													>{#if deal.img}<img
															src={deal.img}
															alt=""
															style="width:76px;height:76px;border-radius:{T.r};object-fit:cover;background:{T.sf2}"
														/>{/if}<span
														><span style="display:block;font-size:14.5px;font-weight:600"
															>{deal.n}</span
														><span style="display:flex;gap:8px;align-items:baseline;margin-top:3px"
															><b style="font-size:20px">{deal.price}</b>{#if deal.old}<s
																	style="font-size:13px;opacity:.6">{deal.old}</s
																>{/if}</span
														></span
													></span
												>
											</a>{/if}
										{#if whatsapp}<a
												href={wa('¡Hola! Quiero hacer un pedido.')}
												target="_blank"
												rel="noopener"
												style="padding:16px 18px;border-radius:{T.rL};background:{T.sf};display:flex;gap:12px;align-items:center;color:{T.ink};text-decoration:none"
												><Icon icon={I.logosWhatsappIcon} size={28} /><span style="flex:1"
													><span style="display:block;font-size:14.5px;font-weight:700"
														>Pedí por WhatsApp</span
													><span style="display:block;font-size:12.5px;color:{T.mu}"
														>Mandanos la lista y te lo&nbsp;armamos</span
													></span
												><Icon icon={I.phArrowRightBold} size={16} /></a
											>{/if}
									</div>
								{/if}
							</section>
						{/if}
					{/if}
					{#if isCat && ver('cats') && rail.length}
						<section style="padding:14px clamp(12px,3cqw,36px) 0">
							<div style="display:flex;gap:8px;overflow-x:auto;scrollbar-width:none">
								{#each rail as c (c.nombre)}<a
										href={enlaceCat(c.nombre)}
										style="flex:none;display:flex;flex-direction:column;align-items:center;gap:7px;width:104px;padding:12px 6px;border-radius:{T.r};background:{T.sf};color:{T.ink};font:600 12.5px 'Instrument Sans';text-align:center;text-decoration:none"
										><span
											style="display:block;width:56px;height:56px;border-radius:50%;overflow:hidden;background:{T.sf2}"
											>{#if c.foto}<img
													src={img(c.foto, 140, 140)}
													alt=""
													style="width:100%;height:100%;object-fit:cover"
												/>{/if}</span
										>{c.nombre}</a
									>{/each}
							</div>
						</section>
					{/if}

					<div
						style="margin-top:{perkMt};display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));border-top:1px solid {T.ln};border-bottom:1px solid {T.ln}"
					>
						{#each perks as p (p.t)}<div
								style="display:flex;align-items:center;gap:12px;padding:16px clamp(14px,3cqw,32px)"
							>
								<Icon icon={p.i} size={22} style="color:{T.at}" />
								<div>
									<div style="font-size:13.5px;font-weight:600">{p.t}</div>
									<div style="font-size:12.5px;color:{T.mu}">{p.d}</div>
								</div>
							</div>{/each}
					</div>
				{:else if seccion === 'cats'}
					{#if ver('cats') && cats.length}
						<section style="padding:clamp(44px,6.5cqw,90px) clamp(14px,3.5cqw,44px) 0">
							<div
								style="display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,240px),1fr));gap:clamp(10px,1.4cqw,18px)"
							>
								{#each cats as c (c.nombre)}<a
										href={enlaceCat(c.nombre)}
										class="zoom"
										style="position:relative;display:block;aspect-ratio:{catAr};border-radius:{T.rL};overflow:hidden;background:{T.sf2};text-align:left"
										>{#if c.foto}<img
												src={img(c.foto, 800, isEd ? 1067 : 600)}
												alt=""
												style="position:absolute;inset:0;width:100%;height:100%;object-fit:cover"
											/>{/if}<span
											style="position:absolute;inset:auto 0 0 0;padding:20px;background:linear-gradient(0deg,rgba(0,0,0,.6),transparent);color:#fff;display:flex;justify-content:space-between;align-items:end;gap:10px"
											><span
												style="font-family:{T.f};font-weight:{T.fw};font-size:clamp(26px,3cqw,36px);line-height:1;letter-spacing:{T.ls};text-transform:{T.tt}"
												>{c.nombre}</span
											><span style="font-size:12.5px;white-space:nowrap"
												>{c.cantidad} {c.cantidad === 1 ? 'producto' : 'productos'}</span
											></span
										></a
									>{/each}
							</div>
						</section>
					{/if}
				{:else if seccion === 'prods'}
					{#if ver('prods') && best.length}
						<section style="padding:clamp(48px,7cqw,100px) clamp(14px,3.5cqw,44px) 0">
							<div
								style="display:flex;justify-content:space-between;align-items:end;gap:16px;flex-wrap:wrap"
							>
								<h2
									style="margin:0;font-family:{T.f};font-weight:{T.fw};font-size:clamp(34px,4.8cqw,62px);line-height:1;letter-spacing:{T.ls};text-transform:{T.tt}"
								>
									Lo más elegido
								</h2>
								<a
									href="{base}/catalogo"
									style="text-decoration:none;padding:0;font:600 14px 'Instrument Sans';color:{T.ink};border-bottom:1.5px solid {T.ink}"
									>Ver todo</a
								>
							</div>
							<div
								style="margin-top:24px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(46%,{cardMin}),1fr));gap:clamp(12px,1.6cqw,20px) clamp(10px,1.4cqw,16px)"
							>
								{#each best as p (p.id)}{@render card(p, true, true)}{/each}
							</div>
						</section>
					{/if}
				{:else if seccion === 'revs'}
					{#if ver('revs') && tienda.resenas.length}
						<section style="padding:clamp(48px,7cqw,100px) clamp(14px,3.5cqw,44px) 0">
							<div
								style="display:flex;justify-content:space-between;align-items:end;gap:16px;flex-wrap:wrap"
							>
								<h2
									style="margin:0;font-family:{T.f};font-weight:{T.fw};font-size:clamp(34px,4.8cqw,62px);line-height:1;letter-spacing:{T.ls};text-transform:{T.tt}"
								>
									Lo que dicen
								</h2>
								{#if tienda.promedio}<span
										style="display:flex;align-items:center;gap:8px;font-size:14px"
										><span style="display:flex;color:{T.star}"
											>{#each estrellas(tienda.promedio) as e, i (i)}<Icon
													icon={e}
													size={15}
												/>{/each}</span
										><b>{tienda.promedio.toLocaleString('es-AR')}</b><span style="color:{T.mu}"
											>· {tienda.resenas.length}
											{tienda.resenas.length === 1 ? 'reseña' : 'reseñas'}</span
										></span
									>{/if}
							</div>
							<div
								style="margin-top:24px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,290px),1fr));gap:14px"
							>
								{#each tienda.resenas.slice(0, 3) as r, i (i)}
									{@const pr = tienda.productos.find((p) => p.nombre === r.producto)}
									<div
										style="display:flex;gap:14px;padding:14px;background:{T.sf};border-radius:{T.rL}"
									>
										{#if pr?.foto}<img
												src={img(pr.foto, 200, 250)}
												alt=""
												style="width:92px;height:116px;object-fit:cover;border-radius:{T.r};flex:none;background:{T.sf2}"
											/>{/if}
										<div style="min-width:0">
											<div style="display:flex;gap:1px;color:{T.star}">
												{#each estrellas(r.estrellas) as e, j (j)}<Icon icon={e} size={13} />{/each}
											</div>
											<p style="margin:8px 0 0;font-size:15px;line-height:1.45">“{r.texto}”</p>
											<div style="margin-top:8px;font-size:12.5px;color:{T.mu}">
												{r.autor}{#if r.producto}
													· {r.conCompra ? 'compró' : 'sobre'} {r.producto.toLowerCase()}{/if}
											</div>
											{#if r.conCompra}<div
													style="margin-top:6px;display:flex;align-items:center;gap:4px;font-size:12px;font-weight:700;color:{T.ok}"
												>
													<Icon icon={I.phSealCheckFill} size={13} />Compra verificada
												</div>{/if}
											{#if r.respuesta}<div
													style="margin-top:10px;padding:10px 12px;border-radius:{T.r};background:{T.sf2};font-size:13px;line-height:1.5;color:{T.mu}"
												>
													<b style="color:{T.ink}">Respuesta de {tienda.nombre}</b><br
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
						<section style="padding:clamp(48px,7cqw,100px) clamp(14px,3.5cqw,44px)">
							<div>
								<div
									style="font-size:12.5px;letter-spacing:.2em;text-transform:uppercase;color:{T.mu}"
								>
									{tienda.nombre}
								</div>
								<h2
									style="margin:8px 0 0;font-family:{T.f};font-weight:{T.fw};font-size:clamp(34px,4.8cqw,62px);line-height:1;letter-spacing:{T.ls};text-transform:{T.tt}"
								>
									En fotos
								</h2>
							</div>
							<div
								style="margin-top:22px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(30%,160px),1fr));gap:6px"
							>
								{#each tienda.fotos as g (g)}<div
										style="aspect-ratio:1;overflow:hidden;border-radius:{T.r};background:{T.sf2}"
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
				<div style="height:clamp(48px,7cqw,100px)"></div>
			{/if}
		{/if}

		<!-- ============ CATÁLOGO ============ -->
		{#if pagina === 'catalogo' && !tienda.productos.length}
			<section style="padding:clamp(40px,7cqw,96px) clamp(16px,4cqw,48px)">
				<VacioTienda {tienda} {tema} motivo="sin-productos" />
			</section>
		{:else if pagina === 'catalogo'}
			<section style="padding:24px clamp(14px,3.5cqw,44px) clamp(48px,7cqw,90px)">
				<div style="font-size:12.5px;color:{T.mu}">
					<a href={base || '/'} style="color:inherit;text-decoration:none">Inicio</a> /
					<span style="color:{T.ink}">{categoria ?? 'Todo'}</span>
				</div>
				<div
					style="margin-top:12px;display:flex;justify-content:space-between;align-items:end;gap:16px;flex-wrap:wrap"
				>
					<h1
						style="margin:0;font-family:{T.f};font-weight:{T.fw};font-size:clamp(42px,6.4cqw,88px);line-height:.95;letter-spacing:{T.ls};text-transform:{T.tt};overflow-wrap:anywhere"
					>
						{categoria ?? 'Todo'}
					</h1>
					<span style="font-size:13.5px;color:{T.mu}"
						>{listado.length} {listado.length === 1 ? 'producto' : 'productos'}</span
					>
				</div>
				<div
					style="margin-top:20px;display:flex;gap:8px;overflow-x:auto;scrollbar-width:none;padding-bottom:2px"
				>
					{#each [null, ...tienda.categorias.map((c) => c.nombre)] as c (c ?? '')}{@const on =
							c === categoria}<a
							href={enlaceCat(c)}
							style="flex:none;text-decoration:none;padding:9px 15px;border-radius:{T.rb};border:1.5px solid {on
								? T.ink
								: T.ln};background:{on ? T.ink : 'transparent'};color:{on
								? T.bg
								: T.ink};font-size:13.5px;font-weight:600;white-space:nowrap">{c ?? 'Todo'}</a
						>{/each}
				</div>
				<div style="margin-top:20px;display:flex;gap:clamp(18px,3cqw,40px);align-items:flex-start">
					{#if sideF}
						<aside
							style="width:230px;flex:none;position:sticky;top:90px;display:flex;flex-direction:column;gap:24px;font-size:13.5px"
						>
							{#if precios.length > 1}<div>
									<div
										style="font-size:12px;font-weight:700;letter-spacing:.12em;text-transform:uppercase;margin-bottom:11px"
									>
										Precio
									</div>
									<div style="position:relative;height:3px;background:{T.ln};margin:10px 4px">
										<span style="position:absolute;left:0;right:0;top:0;bottom:0;background:{T.ink}"
										></span><span
											style="position:absolute;left:0;top:-6px;width:15px;height:15px;margin-left:-7px;border-radius:50%;background:{T.bg};border:2px solid {T.ink}"
										></span><span
											style="position:absolute;right:0;top:-6px;width:15px;height:15px;margin-right:-7px;border-radius:50%;background:{T.bg};border:2px solid {T.ink}"
										></span>
									</div>
									<div
										style="display:flex;justify-content:space-between;color:{T.mu};margin-top:12px"
									>
										<span>{fmt(Math.min(...precios))}</span><span>{fmt(Math.max(...precios))}</span>
									</div>
								</div>{/if}
							<div>
								<div
									style="font-size:12px;font-weight:700;letter-spacing:.12em;text-transform:uppercase;margin-bottom:11px"
								>
									Filtrar por
								</div>
								<div style="display:flex;flex-direction:column;gap:10px">
									{#each tienda.categorias as m (m.nombre)}{@const on = m.nombre === categoria}<a
											href={enlaceCat(on ? null : m.nombre)}
											style="display:flex;align-items:center;gap:10px;color:{T.ink};text-decoration:none"
											><span
												style="width:17px;height:17px;border-radius:4px;border:1.5px solid {T.ink};background:{on
													? T.ink
													: 'transparent'};color:{T.bg};display:grid;place-items:center"
												>{#if on}<Icon icon={I.phCheckBold} size={11} />{/if}</span
											>{m.nombre}<span style="margin-left:auto;color:{T.mu}">{m.cantidad}</span></a
										>{/each}
								</div>
							</div>
						</aside>
					{/if}
					<div style="flex:1;min-width:0">
						<div
							style="display:flex;justify-content:space-between;align-items:center;gap:10px;flex-wrap:wrap;padding-bottom:14px;border-bottom:1px solid {T.ln};font-size:13.5px"
						>
							<span style="display:flex;gap:6px;flex-wrap:wrap"
								>{#if categoria}<a
										href={enlaceCat(null)}
										aria-label="Quitar el filtro {categoria}"
										style="display:flex;align-items:center;gap:6px;padding:7px 11px;border-radius:{T.rb};background:{T.sf2};color:{T.ink};text-decoration:none"
										>{categoria}<Icon icon={I.phX} size={11} /></a
									>{/if}</span
							>
							<label
								style="position:relative;display:flex;align-items:center;gap:6px;font-weight:600"
								><span class="oculto">Ordenar</span><select
									value={orden}
									onchange={(e) =>
										goto(enlaceCat(categoria, e.currentTarget.value), { replace: true })}
									style="appearance:none;-webkit-appearance:none;border:0;padding:0 18px 0 0;background:transparent;font:600 13.5px 'Instrument Sans';color:{T.ink};cursor:pointer"
									>{#each ORDENES as [k, t] (k)}<option value={k}>{t}</option>{/each}</select
								><Icon
									icon={I.phCaretDown}
									size={13}
									style="position:absolute;right:0;pointer-events:none"
								/></label
							>
						</div>
						{#if catalog.length}
							<div
								style="margin-top:20px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(46%,{cardMinCat}),1fr));gap:clamp(12px,1.8cqw,22px) clamp(10px,1.4cqw,16px)"
							>
								{#each catalog as p (p.id)}{@render card(p, false, false)}{/each}
							</div>
							{#if listado.length > mostrar}
								<div style="margin-top:36px;display:flex;justify-content:center">
									<button
										type="button"
										onclick={() => (mostrar += 24)}
										style="padding:0 16px;height:38px;display:grid;place-items:center;border-radius:{T.rb};border:1px solid {T.ln};background:transparent;font:600 14px 'Instrument Sans';color:{T.ink};cursor:pointer"
										>Ver más</button
									>
								</div>
							{/if}
						{:else}
							<p style="margin:40px 0;font-size:15px;color:{T.mu}">
								Todavía no hay productos en esta&nbsp;categoría.
							</p>
						{/if}
					</div>
				</div>
			</section>
		{/if}

		<!-- ============ PRODUCTO ============ -->
		{#if pagina === 'producto' && producto && pd && sel}
			<section style="padding:22px clamp(14px,3.5cqw,44px) 0">
				<div style="font-size:12.5px;color:{T.mu}">
					<a href={base || '/'} style="color:inherit;text-decoration:none">Inicio</a> / {#if producto.categoria}<a
							href={enlaceCat(producto.categoria)}
							style="color:inherit;text-decoration:none">{producto.categoria}</a
						> /
					{/if}<span style="color:{T.ink}">{producto.nombre}</span>
				</div>
				<div
					style="margin-top:16px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,400px),1fr));gap:clamp(22px,4cqw,60px);align-items:start"
				>
					<div style="display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:8px">
						<div
							style="grid-column:1/-1;position:relative;aspect-ratio:{cardAr};border-radius:{T.rL};overflow:hidden;background:{T.sf2}"
						>
							{#if sel.foto}<img
									src={img(sel.foto, 1100)}
									alt={producto.nombre}
									style="width:100%;height:100%;object-fit:cover"
								/>{/if}{#if pd.tag}<span
									style="position:absolute;top:14px;left:14px;padding:5px 10px;border-radius:{T.rb};background:{pd.tagBg};color:{pd.tagFg};font-size:12px;font-weight:700"
									>{pd.tag}</span
								>{/if}
						</div>
						{#if sel.fotos.length > 1}<div
								style="grid-column:1/-1;display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:8px"
							>
								{#each sel.fotos as f, i (f)}<button
										type="button"
										aria-label="Ver la foto {i + 1}"
										aria-pressed={i === sel.fotoIdx}
										onclick={() => sel.verFoto(i)}
										style="aspect-ratio:{cardAr};padding:0;border:0;border-radius:{T.r};overflow:hidden;background:{T.sf2};cursor:pointer;opacity:{i ===
										sel.fotoIdx
											? 1
											: 0.7};box-shadow:inset 0 0 0 {i === sel.fotoIdx ? '2px' : '0px'} {T.ink}"
										><img
											src={img(f, 240, 240)}
											alt=""
											style="display:block;width:100%;height:100%;object-fit:cover"
										/></button
									>{/each}
							</div>{/if}
					</div>
					<div style="position:sticky;top:90px">
						{#if producto.categoria}<div
								style="font-size:12.5px;letter-spacing:.16em;text-transform:uppercase;color:{T.at};font-weight:600"
							>
								{producto.categoria}
							</div>{/if}
						<h1
							style="margin:10px 0 0;font-family:{T.f};font-weight:{T.fw};font-size:clamp(36px,4.4cqw,58px);line-height:1;letter-spacing:{T.ls};text-transform:{T.tt};overflow-wrap:anywhere"
						>
							{producto.nombre}
						</h1>
						{#if resenasDe}<div
								style="margin-top:10px;display:flex;align-items:center;gap:8px;font-size:13px"
							>
								<span style="display:flex;color:{T.star}"
									>{#each estrellas(5) as e, i (i)}<Icon icon={e} size={13} />{/each}</span
								><span>{resenasDe} {resenasDe === 1 ? 'reseña' : 'reseñas'}</span>
							</div>{/if}
						<div style="margin-top:18px;display:flex;align-items:baseline;gap:12px">
							<b style="font-size:30px;font-weight:700;letter-spacing:-.01em"
								>{sel.desde ? 'Desde ' : ''}{fmt(sel.precio)}</b
							>{#if tachado}<s style="font-size:17px;color:{T.mu}">{fmt(tachado)}</s>{/if}
						</div>
						<div style="margin-top:8px;display:flex;flex-direction:column;gap:4px;font-size:14px">
							{#if tienda.cobros.cuotas > 1}<span style="display:flex;align-items:center;gap:8px"
									><Icon icon={I.phCreditCard} size={17} />{tienda.cobros.cuotas} × {fmt(
										sel.precio / tienda.cobros.cuotas
									)} sin interés</span
								>{/if}{#if tienda.cobros.transferencia > 0}<span
									style="display:flex;align-items:center;gap:8px;color:{T.ok};font-weight:600"
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
									style="margin-bottom:10px;font-size:12.5px;letter-spacing:.12em;text-transform:uppercase;font-weight:700;color:{sel.aviso &&
									!o.elegido
										? T.at
										: T.ink}"
								>
									{o.nombre}{#if o.elegido}<span style="font-weight:400;color:{T.mu}"
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
												style="width:40px;height:40px;padding:0;border:0;border-radius:50%;overflow:hidden;background:{T.sf2};cursor:pointer;opacity:{x.disponible
													? 1
													: 0.4};box-shadow:inset 0 0 0 1px {T.ln},0 0 0 {x.elegido
													? '2.5px'
													: '0px'} {T.bg},0 0 0 {x.elegido ? '4.5px' : '0px'} {T.ink}"
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
												style="min-width:52px;min-height:46px;padding:10px 16px;border-radius:{T.rb};border:1.5px solid {x.elegido
													? T.ink
													: T.ln};background:{x.elegido ? T.ink : 'transparent'};color:{x.elegido
													? T.bg
													: x.disponible
														? T.ink
														: T.mu};font:600 14px 'Instrument Sans';text-decoration:{x.disponible
													? 'none'
													: 'line-through'};opacity:{x.disponible ? 1 : 0.6};cursor:{x.disponible
													? 'pointer'
													: 'default'}">{x.valor}</button
											>{/each}
									</div>{/if}
							</div>{/each}
						{#if sel.stock > 0 && sel.stock <= 3 && (!sel.conOpciones || sel.variante)}<div
								style="margin-top:16px;font-size:12.5px;color:{T.at};font-weight:600"
							>
								{sel.stock === 1 ? 'Queda 1' : `Quedan ${sel.stock}`}{#if sel.variante}
									en {sel.texto.toLowerCase()}{:else}
									en stock{/if}
							</div>{/if}
						<div style="margin-top:20px;display:flex;flex-direction:column;gap:8px">
							<button
								type="button"
								onclick={agregar}
								disabled={sel.stock === 0 || tienda.vitrina}
								style="padding:17px;border:0;border-radius:{T.rb};background:{T.ink};color:{T.bg};font:700 15px 'Instrument Sans';cursor:pointer"
								>{sel.boton('Agregar al carrito')}</button
							>{#if whatsapp}<a
									href={wa(
										`¡Hola! Quiero comprar ${producto.nombre}${sel.variante ? ` (${sel.texto})` : ''} (${fmt(sel.precio)}).`
									)}
									target="_blank"
									rel="noopener"
									style="display:flex;justify-content:center;align-items:center;gap:8px;padding:15px;border:1.5px solid {T.ink};border-radius:{T.rb};font-size:14.5px;font-weight:700;color:{T.ink};text-decoration:none"
									><Icon icon={I.phWhatsappLogo} size={18} />Comprar por WhatsApp</a
								>{/if}
							<CompartirProducto {tienda} {producto} {tema} />
						</div>
						{#if envia(tienda.envios)}<div
								style="margin-top:18px;padding:16px;background:{T.sf};border-radius:{T.rL}"
							>
								<label for="cp-ficha" style="font-size:13.5px;font-weight:700"
									>Calculá el envío</label
								>
								<div style="margin-top:10px;display:flex;gap:6px">
									<input
										id="cp-ficha"
										bind:value={cp}
										inputmode="numeric"
										maxlength="4"
										placeholder="Código postal"
										style="flex:1;min-width:0;padding:11px 12px;background:{T.bg};border:0;border-radius:{T.r};font:400 14px 'Instrument Sans';color:{T.ink}"
										onkeydown={(e) => e.key === 'Enter' && calcular()}
									/><button
										type="button"
										onclick={calcular}
										style="padding:11px 16px;border:0;background:{T.ink};color:{T.bg};border-radius:{T.r};font:700 13.5px 'Instrument Sans';cursor:pointer"
										>Calcular</button
									>
								</div>
								{#if cotizado}<div
										style="margin-top:12px;display:flex;flex-direction:column;gap:8px;font-size:13.5px"
									>
										{#each cotizado as o (o.id)}<span
												style="display:flex;justify-content:space-between;gap:10px"
												><span>{o.nombre} · {o.detalle}</span><b
													style="color:{o.precio ? T.ink : T.ok};white-space:nowrap"
													>{o.precio ? fmt(o.precio) : 'Gratis'}</b
												></span
											>{:else}<span style="color:{T.mu}"
												>No llegamos a ese código postal por&nbsp;ahora.</span
											>{/each}
									</div>{/if}
							</div>{/if}
						<div style="margin-top:14px;border-top:1px solid {T.ln}">
							{#each acc as a (a.t)}<div style="padding:16px 0;border-bottom:1px solid {T.ln}">
									<button
										type="button"
										onclick={() => (abierto = a.open ? '' : a.t)}
										aria-expanded={a.open}
										style="width:100%;display:flex;justify-content:space-between;padding:0;border:0;background:transparent;font:700 14px 'Instrument Sans';color:{T.ink};cursor:pointer;text-align:left"
										>{a.t}<Icon icon={a.i} size={15} /></button
									>{#if a.open}<p
											style="margin:10px 0 0;font-size:14.5px;line-height:1.6;color:{T.mu}"
										>
											{a.d}
										</p>{/if}
								</div>{/each}
						</div>
					</div>
				</div>
			</section>
			<section style="padding:clamp(44px,6cqw,84px) clamp(14px,3.5cqw,44px) 0;max-width:820px">
				<PreguntasProducto
					slug={tienda.slug}
					productoId={producto.id}
					preguntas={tienda.preguntas}
					tema={{ ...tema, tarjeta: 'transparent' }}
					titulo="Preguntas sobre este producto"
				/>
			</section>
			{#if related.length}
				<section style="padding:clamp(52px,7cqw,96px) clamp(14px,3.5cqw,44px)">
					<h2
						style="margin:0;font-family:{T.f};font-weight:{T.fw};font-size:clamp(32px,4.2cqw,52px);line-height:1;letter-spacing:{T.ls};text-transform:{T.tt}"
					>
						También te puede gustar
					</h2>
					<div
						style="margin-top:22px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(46%,210px),1fr));gap:14px"
					>
						{#each related as p (p.id)}<a
								href={p.href}
								style="display:block;text-decoration:none;text-align:left;color:{T.ink};font-family:inherit"
								><span
									style="display:block;aspect-ratio:{cardAr};border-radius:{T.r};overflow:hidden;background:{T.sf2}"
									>{#if p.img}<img
											src={p.img}
											alt=""
											style="width:100%;height:100%;object-fit:cover"
										/>{/if}</span
								><span style="display:block;margin-top:10px;font-size:14px;font-weight:500"
									>{p.n}</span
								><b style="font-size:15px">{p.price}</b></a
							>{/each}
					</div>
				</section>
			{:else}
				<div style="height:clamp(52px,7cqw,96px)"></div>
			{/if}
		{/if}

		<!-- ============ CARRITO ============ -->
		{#if pagina === 'carrito'}
			<section style="padding:28px clamp(14px,3.5cqw,44px) clamp(48px,7cqw,90px)">
				<h1
					style="margin:0;font-family:{T.f};font-weight:{T.fw};font-size:clamp(42px,6cqw,78px);line-height:1;letter-spacing:{T.ls};text-transform:{T.tt}"
				>
					Tu carrito <span style="font-size:.5em;color:{T.mu}">({carrito.unidades})</span>
				</h1>
				{#if !lineas.length}
					<p style="margin:24px 0 0;font-size:15px;color:{T.mu}">Todavía no sumaste nada.</p>
					<a
						href="{base}/catalogo"
						style="display:inline-block;margin-top:18px;text-decoration:none;padding:15px 26px;border-radius:{T.rb};background:{T.ink};color:{T.bg};font:700 15px 'Instrument Sans'"
						>Ver el catálogo</a
					>
				{:else}
					{#if tienda.envios.gratisDesde}
						<div
							style="margin-top:14px;padding:12px 16px;background:{T.sf};border-radius:{T.r};font-size:13.5px;display:flex;align-items:center;gap:12px;flex-wrap:wrap"
						>
							<span style="flex:1;min-width:200px"
								>{#if falta}Te faltan <b>{fmt(falta)}</b> para el envío gratis{:else}<b
										>Tenés envío gratis</b
									>{/if}</span
							><span
								style="flex:1;min-width:160px;height:5px;background:{T.ln};border-radius:3px;overflow:hidden"
								><span
									style="display:block;width:{Math.min(
										100,
										(subtotal / tienda.envios.gratisDesde) * 100
									)}%;height:100%;background:{T.acc}"
								></span></span
							>
						</div>
					{/if}
					<div
						style="margin-top:22px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,360px),1fr));gap:clamp(22px,4cqw,56px);align-items:start"
					>
						<div style="display:flex;flex-direction:column">
							{#each lineas as c (c.clave)}
								<div style="display:flex;gap:16px;padding:16px 0;border-bottom:1px solid {T.ln}">
									{#if c.producto.foto}<img
											src={img(c.producto.foto, 200, 220)}
											alt=""
											style="width:96px;height:110px;object-fit:cover;border-radius:{T.r};flex:none;background:{T.sf2}"
										/>{/if}
									<div style="flex:1;min-width:0;display:flex;flex-direction:column">
										<div style="display:flex;justify-content:space-between;gap:10px">
											<a
												href="{base}/producto/{c.producto.slug ?? c.id}"
												style="font-size:15px;font-weight:500;color:{T.ink};text-decoration:none"
												>{c.producto.nombre}</a
											><b style="white-space:nowrap">{fmt(c.subtotal)}</b>
										</div>
										<span style="font-size:13px;color:{T.mu};margin-top:4px"
											>{c.variante ? `${c.variante} · ` : ''}{fmt(c.producto.precio)} c/u</span
										><span style="flex:1;min-height:12px"></span>
										<div style="display:flex;justify-content:space-between;align-items:center">
											<span
												style="display:flex;align-items:center;border:1px solid {T.ln};border-radius:{T.rb}"
												><button
													type="button"
													aria-label="Uno menos"
													onclick={() => carrito.fijar(c.clave, c.cantidad - 1)}
													style="width:34px;height:34px;display:grid;place-items:center;border:0;background:transparent;cursor:pointer;color:{T.ink}"
													><Icon icon={I.phMinus} size={13} /></button
												><span style="width:24px;text-align:center;font-size:14px"
													>{c.cantidad}</span
												><button
													type="button"
													aria-label="Uno más"
													disabled={c.cantidad >= c.producto.stock}
													onclick={() => carrito.fijar(c.clave, c.cantidad + 1)}
													style="width:34px;height:34px;display:grid;place-items:center;border:0;background:transparent;cursor:pointer;color:{T.ink}"
													><Icon icon={I.phPlus} size={13} /></button
												></span
											><button
												type="button"
												onclick={() => carrito.fijar(c.clave, 0)}
												style="border:0;background:transparent;padding:0;font:400 13px 'Instrument Sans';text-decoration:underline;color:{T.mu};cursor:pointer"
												>Quitar</button
											>
										</div>
									</div>
								</div>
							{/each}
							{#if upsell.length}
								<div
									style="margin-top:26px;font-size:13px;font-weight:700;letter-spacing:.12em;text-transform:uppercase"
								>
									Te puede gustar
								</div>
								<div
									style="margin-top:12px;display:flex;gap:10px;overflow-x:auto;scrollbar-width:none"
								>
									{#each upsell as p (p.id)}<div style="flex:none;width:150px">
											<a
												href="{base}/producto/{p.slug ?? p.id}"
												style="display:block;aspect-ratio:1;border-radius:{T.r};overflow:hidden;background:{T.sf2}"
												>{#if p.foto}<img
														src={img(p.foto, 300, 300)}
														alt=""
														style="width:100%;height:100%;object-fit:cover"
													/>{/if}</a
											>
											<div style="margin-top:8px;font-size:13.5px">{p.nombre}</div>
											<div
												style="display:flex;justify-content:space-between;align-items:center;margin-top:2px"
											>
												<b style="font-size:14px">{tarjeta(p).price}</b><button
													type="button"
													aria-label="Sumar {p.nombre}"
													disabled={tienda.vitrina}
													onclick={() => sumar(p.id)}
													style="width:28px;height:28px;border-radius:50%;border:1.5px solid {T.ink};background:transparent;display:grid;place-items:center;cursor:pointer;color:{T.ink}"
													><Icon icon={I.phPlus} size={13} /></button
												>
											</div>
										</div>{/each}
								</div>
							{/if}
						</div>
						<div
							style="position:sticky;top:90px;padding:24px;background:{T.sf};border-radius:{T.rL};display:flex;flex-direction:column;gap:12px;font-size:14px"
						>
							<div
								style="font-family:{T.f};font-weight:{T.fw};font-size:28px;text-transform:{T.tt}"
							>
								Resumen
							</div>
							<span style="display:flex;justify-content:space-between"
								><span>Subtotal</span><span>{fmt(subtotal)}</span></span
							>
							<div style="display:flex;gap:6px">
								<input
									bind:value={cp}
									inputmode="numeric"
									maxlength="4"
									placeholder="Código postal"
									aria-label="Código postal"
									onkeydown={(e) => e.key === 'Enter' && (cpCarrito = cp)}
									style="flex:1;min-width:0;padding:11px 12px;background:{T.bg};border:0;border-radius:{T.r};font:400 14px 'Instrument Sans';color:{T.ink}"
								/><button
									type="button"
									onclick={() => (cpCarrito = cp)}
									style="padding:11px 14px;border:1.5px solid {T.ink};border-radius:{T.r};background:transparent;font:600 14px 'Instrument Sans';color:{T.ink};cursor:pointer"
									>Calcular</button
								>
							</div>
							<span style="display:flex;justify-content:space-between;gap:10px"
								><span>Envío</span><span
									style="text-align:right;color:{envioCarrito === null ? T.mu : T.ink}"
									>{envioCarrito === null
										? 'Se calcula en el próximo paso'
										: envioCarrito < 0
											? 'No llegamos a ese código'
											: envioCarrito === 0
												? 'Gratis'
												: `Desde ${fmt(envioCarrito)}`}</span
								></span
							>
							<div style="height:1px;background:{T.ln}"></div>
							<span style="display:flex;justify-content:space-between;align-items:baseline"
								><b style="font-size:16px">Total</b><b style="font-size:24px">{fmt(subtotal)}</b
								></span
							>
							{#if tienda.cobros.cuotas > 1}<span style="font-size:13px;color:{T.mu}"
									>{tienda.cobros.cuotas} cuotas sin interés de {fmt(
										subtotal / tienda.cobros.cuotas
									)}</span
								>{/if}
							{#if tienda.cobros.transferencia > 0}<span
									style="font-size:13.5px;color:{T.ok};font-weight:600"
									>{fmt(subtotal * (1 - tienda.cobros.transferencia / 100))} pagando por transferencia</span
								>{/if}
							<a
								href="{base}/checkout"
								style="margin-top:6px;padding:17px;border-radius:{T.rb};background:{T.ink};color:{T.bg};font:700 15px 'Instrument Sans';text-align:center;text-decoration:none"
								>Finalizar compra</a
							>
							<span style="display:flex;justify-content:center;gap:10px"
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
				<section style="padding:28px clamp(14px,3.5cqw,44px) clamp(48px,7cqw,90px)">
					<p style="font-size:15px;color:{T.mu}">Tu carrito está vacío.</p>
					<a href="{base}/catalogo" style="color:{T.ink}">Ver el catálogo</a>
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
					style="padding:28px clamp(14px,3.5cqw,44px) clamp(48px,7cqw,90px);display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,360px),1fr));gap:clamp(22px,4cqw,56px);align-items:start"
				>
					<input type="hidden" name="carrito" value={JSON.stringify(carrito.lineas)} />
					<input type="hidden" name="envio" value={envio ?? ''} />
					<input type="hidden" name="pago" value={pago} />
					<div style="display:flex;flex-direction:column;gap:26px">
						<div
							style="display:flex;gap:8px;font-size:12.5px;letter-spacing:.1em;text-transform:uppercase"
						>
							<span style="font-weight:700">1 Datos</span><span style="color:{T.mu}"
								>— 2 Entrega — 3 Pago</span
							>
						</div>
						<div>
							<div
								style="font-family:{T.f};font-weight:{T.fw};font-size:30px;text-transform:{T.tt}"
							>
								Tus datos
							</div>
							<div
								style="margin-top:12px;display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));gap:10px"
							>
								{#each campos as [n, l, tipo, ac] (n)}
									<label
										style="display:block;padding:10px 12px;border:1.5px solid {T.ln};border-radius:{T.r};background:{T.sf}"
										><span style="display:block;font-size:11.5px;color:{T.mu}">{l}</span><input
											name={n}
											type={tipo}
											autocomplete={ac}
											required={!OPCIONALES.has(n)}
											maxlength="120"
											style="width:100%;margin-top:2px;padding:0;border:0;background:transparent;font:400 14.5px 'Instrument Sans';color:{T.ink};outline:none"
										/></label
									>
								{/each}
							</div>
						</div>
						<div>
							<div
								style="font-family:{T.f};font-weight:{T.fw};font-size:30px;text-transform:{T.tt}"
							>
								Entrega
							</div>
							<div
								style="margin-top:12px;display:grid;grid-template-columns:minmax(0,140px) minmax(0,1fr);gap:10px"
							>
								<label
									style="display:block;padding:10px 12px;border:1.5px solid {T.ln};border-radius:{T.r};background:{T.sf}"
									><span style="display:block;font-size:11.5px;color:{T.mu}">Código postal</span
									><input
										name="cp"
										bind:value={cp}
										inputmode="numeric"
										maxlength="4"
										style="width:100%;margin-top:2px;padding:0;border:0;background:transparent;font:400 14.5px 'Instrument Sans';color:{T.ink};outline:none"
									/></label
								>
								<label
									style="display:block;padding:10px 12px;border:1.5px solid {T.ln};border-radius:{T.r};background:{T.sf}"
									><span style="display:block;font-size:11.5px;color:{T.mu}"
										>Dirección (calle, número, piso)</span
									><input
										name="direccion"
										maxlength="160"
										autocomplete="street-address"
										required={envio !== 'retiro'}
										style="width:100%;margin-top:2px;padding:0;border:0;background:transparent;font:400 14.5px 'Instrument Sans';color:{T.ink};outline:none"
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
											? T.ink
											: T.ln};border-radius:{T.r};background:{on
											? T.sf
											: 'transparent'};text-align:left;cursor:pointer;font-family:inherit;color:{T.ink}"
										><span
											style="width:18px;height:18px;border-radius:50%;border:2px solid {T.ink};display:grid;place-items:center;flex:none"
											>{#if on}<span
													style="width:8px;height:8px;border-radius:50%;background:{T.ink}"
												></span>{/if}</span
										><span style="flex:1"
											><span style="display:block;font-size:14.5px;font-weight:600">{s.nombre}</span
											><span style="display:block;font-size:12.5px;color:{T.mu}">{s.detalle}</span
											></span
										><b style="font-size:14px;color:{s.precio ? T.ink : T.ok}"
											>{s.precio ? fmt(s.precio) : 'Gratis'}</b
										></button
									>{:else}<span style="font-size:13.5px;color:{T.mu}"
										>Poné tu código postal para ver cómo te&nbsp;llega.</span
									>{/each}
							</div>
							<EnvioCorreo {tienda} {cp} {envio} {opciones} />
						</div>
						<div>
							<div
								style="font-family:{T.f};font-weight:{T.fw};font-size:30px;text-transform:{T.tt}"
							>
								Pago
							</div>
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
											? T.ink
											: T.ln};border-radius:{T.r};background:{on
											? T.sf
											: 'transparent'};text-align:left;cursor:pointer;font-family:inherit;color:{T.ink}"
										><span
											style="width:18px;height:18px;border-radius:50%;border:2px solid {T.ink};display:grid;place-items:center;flex:none"
											>{#if on}<span
													style="width:8px;height:8px;border-radius:50%;background:{T.ink}"
												></span>{/if}</span
										><Icon icon={s.i} size={22} style="color:{s.c}" /><span style="flex:1"
											><span style="display:block;font-size:14.5px;font-weight:600">{s.t}</span
											><span style="display:block;font-size:12.5px;color:{T.mu}">{s.d}</span></span
										></button
									>{/each}
							</div>
						</div>
						{#if error}<div
								role="alert"
								style="padding:12px 14px;border-radius:{T.r};background:color-mix(in srgb,#B42318 12%,transparent);color:#B42318;font-size:14px"
							>
								{error}
							</div>{/if}
						<button
							disabled={!envio || !pagos.length || enviando || tienda.vitrina}
							style="padding:18px;border:0;border-radius:{T.rb};background:{T.acc};color:{T.ai};font:700 15.5px 'Instrument Sans';cursor:pointer;display:flex;justify-content:center;align-items:center;gap:8px"
							><Icon icon={I.phLockSimpleFill} size={16} />{tienda.vitrina
								? 'Es una tienda de ejemplo'
								: !pagos.length
									? 'Esta tienda todavía no tiene medios de pago'
									: envio
										? `Pagar ${fmt(total)}`
										: 'Elegí cómo te llega'}</button
						>
					</div>
					<div style="position:sticky;top:90px;padding:24px;background:{T.sf};border-radius:{T.rL}">
						<div style="font-family:{T.f};font-weight:{T.fw};font-size:26px;text-transform:{T.tt}">
							Tu pedido
						</div>
						<div style="margin-top:14px;display:flex;flex-direction:column;gap:12px">
							{#each lineas as c (c.clave)}<div style="display:flex;gap:12px;align-items:center">
									{#if c.producto.foto}<img
											src={img(c.producto.foto, 120, 130)}
											alt=""
											style="width:56px;height:60px;object-fit:cover;border-radius:{T.r};background:{T.sf2}"
										/>{/if}<span style="flex:1;font-size:14px"
										>{c.producto.nombre}<br /><span style="font-size:12.5px;color:{T.mu}"
											>{c.variante ? `${c.variante} · ` : ''}x{c.cantidad}</span
										></span
									><b style="font-size:14px">{fmt(c.subtotal)}</b>
								</div>{/each}
						</div>
						<div style="margin-top:16px">
							<CampoCupon
								{cupon}
								lineas={carrito.lineas}
								tinta={T.ink}
								suave={T.mu}
								borde={T.ln}
								fondo={T.bg}
								acento={T.ink}
								sobreAcento={T.bg}
								radio={T.r}
								fuente="'Instrument Sans',sans-serif"
								error={tid === 'pampa' || tid === 'malbec' ? '#FFB4AB' : '#b3261e'}
							/>
						</div>
						<div
							style="margin-top:16px;padding-top:14px;border-top:1px solid {T.ln};display:flex;flex-direction:column;gap:8px;font-size:14px"
						>
							<span style="display:flex;justify-content:space-between"
								><span>Subtotal</span><span>{fmt(subtotal)}</span></span
							><span style="display:flex;justify-content:space-between"
								><span>Envío</span><span
									>{elegida ? (cuenta.envio ? fmt(cuenta.envio) : 'Gratis') : '—'}</span
								></span
							>{#if cuenta.descCupon}<span
									style="display:flex;justify-content:space-between;color:{T.ok};font-weight:600"
									><span>Cupón {cupon.aplicado}</span><span>− {fmt(cuenta.descCupon)}</span></span
								>{/if}{#if descuento}<span
									style="display:flex;justify-content:space-between;color:{T.ok};font-weight:600"
									><span>Transferencia −{tienda.cobros.transferencia}%</span><span
										>− {fmt(descuento)}</span
									></span
								>{/if}<span
								style="display:flex;justify-content:space-between;align-items:baseline;margin-top:6px"
								><b>Total</b><b style="font-size:24px">{fmt(total)}</b></span
							>
						</div>
						<div
							style="margin-top:14px;display:flex;align-items:center;gap:8px;font-size:12.5px;color:{T.mu}"
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
				style="padding:28px clamp(14px,3.5cqw,44px) clamp(48px,7cqw,90px);display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,360px),1fr));gap:clamp(22px,4cqw,56px);align-items:start"
			>
				<input type="hidden" name="servicioId" value={reserva.servicioId ?? ''} />
				<input type="hidden" name="profesionalId" value={reserva.profesionalId ?? ''} />
				<input type="hidden" name="inicio" value={reserva.inicio ?? ''} />
				<div style="display:flex;flex-direction:column;gap:26px;min-width:0">
					<div
						style="display:flex;gap:8px;flex-wrap:wrap;font-size:12.5px;letter-spacing:.1em;text-transform:uppercase"
					>
						{#each PASOS_TURNO as p, i (p)}<span
								style="font-weight:{i === reserva.paso ? 700 : 400};color:{i === reserva.paso
									? T.ink
									: T.mu}">{i ? '— ' : ''}{i + 1} {p}</span
							>{/each}
					</div>
					<h1
						style="margin:0;font-family:{T.f};font-weight:{T.fw};font-size:clamp(40px,6cqw,72px);line-height:1;letter-spacing:{T.ls};text-transform:{T.tt};overflow-wrap:anywhere"
					>
						Reservá tu turno
					</h1>
					<div>
						<div style="display:flex;justify-content:space-between;align-items:center;gap:10px">
							<div
								style="font-family:{T.f};font-weight:{T.fw};font-size:30px;text-transform:{T.tt}"
							>
								Servicio
							</div>
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
										style="display:flex;align-items:center;gap:12px;padding:15px;border:1.5px solid {on
											? T.ink
											: T.ln};border-radius:{T.r};background:{on
											? T.sf
											: 'transparent'};text-align:left;cursor:pointer;font-family:inherit;color:{T.ink}"
										><span
											style="width:18px;height:18px;border-radius:50%;border:2px solid {T.ink};display:grid;place-items:center;flex:none"
											>{#if on}<span
													style="width:8px;height:8px;border-radius:50%;background:{T.ink}"
												></span>{/if}</span
										><span style="flex:1;min-width:0"
											><span style="display:block;font-size:14.5px;font-weight:600">{s.nombre}</span
											><span style="display:block;font-size:12.5px;color:{T.mu}"
												>{duracion(s.duracion)}</span
											></span
										><b style="font-size:14px;white-space:nowrap">{fmt(s.precio)}</b></button
									>{/each}
							</div>
						</div>
					</div>
					<div>
						<div style="display:flex;justify-content:space-between;align-items:center;gap:10px">
							<div
								style="font-family:{T.f};font-weight:{T.fw};font-size:30px;text-transform:{T.tt}"
							>
								Profesional
							</div>
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
										style="flex:none;width:110px;padding:10px;border:1.5px solid {on
											? T.ink
											: T.ln};border-radius:{T.r};background:{on
											? T.sf
											: 'transparent'};text-align:center;cursor:pointer;font-family:inherit;color:{T.ink}"
										>{#if p?.foto}<img
												src={img(p.foto, 120, 120)}
												alt=""
												style="width:56px;height:56px;border-radius:{T.r};object-fit:cover;display:block;margin:0 auto;background:{T.sf2}"
											/>{:else}<span
												style="width:56px;height:56px;border-radius:{T.r};background:{T.sf2};color:{T.ink};display:grid;place-items:center;margin:0 auto"
												><Icon icon={p ? I.phUserFill : I.phShuffleBold} size={22} /></span
											>{/if}<span
											style="display:block;margin-top:8px;font-size:13.5px;font-weight:600;overflow:hidden;text-overflow:ellipsis;white-space:nowrap"
											>{p?.nombre ?? 'Cualquiera'}</span
										><span
											style="display:block;font-size:11.5px;color:{T.mu};overflow:hidden;text-overflow:ellipsis;white-space:nowrap"
											>{p ? p.rol.split(' · ')[0] : 'El primero libre'}</span
										></button
									>{/each}
							</div>
						</div>
					</div>
					<div>
						<div style="display:flex;justify-content:space-between;align-items:center;gap:10px">
							<div
								style="font-family:{T.f};font-weight:{T.fw};font-size:30px;text-transform:{T.tt}"
							>
								Día y horario
							</div>
							<span
								style="display:flex;align-items:center;gap:10px;font-size:14px;font-weight:600;color:{T.ink}"
								>{#if reserva.hayAntes || reserva.hayDespues}<button
										type="button"
										aria-label="Días anteriores"
										disabled={!reserva.hayAntes}
										onclick={() => reserva.mover(-1)}
										style="display:grid;place-items:center;min-width:44px;min-height:44px;padding:4px;border:0;background:transparent;color:{T.ink};cursor:pointer"
										><Icon icon={I.phCaretLeftBold} size={14} /></button
									>{/if}{reserva.mes}{#if reserva.hayAntes || reserva.hayDespues}<button
										type="button"
										aria-label="Días siguientes"
										disabled={!reserva.hayDespues}
										onclick={() => reserva.mover(1)}
										style="display:grid;place-items:center;min-width:44px;min-height:44px;padding:4px;border:0;background:transparent;color:{T.ink};cursor:pointer"
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
										style="padding:9px 0;border:1.5px solid {on
											? T.ink
											: T.ln};border-radius:{T.r};background:{on
											? T.sf
											: 'transparent'};text-align:center;cursor:pointer;font-family:inherit;color:{T.ink};min-width:0;opacity:{d.abierto
											? 1
											: 0.35}"
										><span
											style="display:block;font-size:11px;font-weight:600;text-transform:uppercase;color:{T.mu}"
											>{d.w}</span
										><span style="display:block;font-size:17px;font-weight:700;margin-top:1px"
											>{d.n}</span
										></button
									>{/each}
							</div>
							<div style="margin-top:16px;display:flex;flex-direction:column;gap:12px">
								{#if !reserva.servicioId}
									<p style="margin:0;font-size:14px;color:{T.mu}">
										Elegí un servicio para ver los horarios.
									</p>
								{:else if reserva.horarios && !reserva.grupos.length}
									<p style="margin:0;font-size:14px;color:{T.mu}">
										No quedan horarios este día. Probá con otro día u otro&nbsp;profesional.
									</p>
								{/if}
								{#each reserva.grupos as g (g.k)}<div>
										<div style="font-size:12.5px;font-weight:600;color:{T.mu};margin-bottom:7px">
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
														? T.ink
														: T.ln};border-radius:{T.r};background:{on
														? T.sf
														: 'transparent'};text-align:center;cursor:pointer;font:600 14px 'Instrument Sans';color:{T.ink};text-decoration:{h.libre
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
								<div
									style="font-family:{T.f};font-weight:{T.fw};font-size:30px;text-transform:{T.tt}"
								>
									Tus datos
								</div>
							</div>
							<div style="margin-top:12px">
								<div
									style="display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,200px),1fr));gap:10px"
								>
									{#each [['nombre', 'Nombre', 'text', 'name'], ['telefono', 'WhatsApp', 'tel', 'tel'], ['email', 'Email', 'email', 'email']] as [n, l, tipo, ac] (n)}
										<label
											style="display:block;padding:10px 12px;border:1.5px solid {T.ln};border-radius:{T.r};background:{T.sf}"
											><span style="display:block;font-size:11.5px;color:{T.mu}">{l}</span><input
												name={n}
												type={tipo}
												autocomplete={ac as 'email'}
												required
												maxlength="120"
												style="width:100%;margin-top:2px;padding:0;border:0;background:transparent;font:400 14.5px 'Instrument Sans';color:{T.ink};outline:none"
											/></label
										>
									{/each}
								</div>
							</div>
						</div>
						<div>
							<div style="display:flex;justify-content:space-between;align-items:center;gap:10px">
								<div
									style="font-family:{T.f};font-weight:{T.fw};font-size:30px;text-transform:{T.tt}"
								>
									Pago
								</div>
							</div>
							<div style="margin-top:12px">
								<div
									style="display:flex;align-items:center;gap:12px;padding:15px;border:1.5px solid {T.ink};border-radius:{T.r};background:{T.sf};color:{T.ink}"
								>
									<span
										style="width:18px;height:18px;border-radius:50%;border:2px solid {T.ink};display:grid;place-items:center;flex:none"
										><span style="width:8px;height:8px;border-radius:50%;background:{T.ink}"
										></span></span
									>{#if reserva.sena}<Icon
											icon={I.siMercadopago}
											size={22}
											style="color:#00B1EA;flex:none"
										/><span style="flex:1;min-width:0"
											><span style="display:block;font-size:14.5px;font-weight:600"
												>Mercado Pago</span
											><span style="display:block;font-size:12.5px;color:{T.mu}"
												>Seña de {fmt(reserva.sena)} para confirmar · el resto, en el local</span
											></span
										>{:else}<Icon
											icon={I.phStorefrontFill}
											size={22}
											style="color:{T.ink};flex:none"
										/><span style="flex:1;min-width:0"
											><span style="display:block;font-size:14.5px;font-weight:600"
												>Pagás en el local</span
											><span style="display:block;font-size:12.5px;color:{T.mu}"
												>No hace falta seña: el turno queda confirmado</span
											></span
										>{/if}
								</div>
							</div>
						</div>
					{/if}

					{#if error}<div
							role="alert"
							style="padding:12px 14px;border-radius:{T.r};background:color-mix(in srgb,#B42318 12%,transparent);color:#B42318;font-size:14px"
						>
							{error}
						</div>{/if}
					<button
						disabled={!reserva.listo || reservando || tienda.vitrina}
						style="padding:18px;border:0;border-radius:{T.rb};background:{T.acc};color:{T.ai};font:700 15.5px 'Instrument Sans';cursor:pointer;display:flex;justify-content:center;align-items:center;gap:8px"
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
					<div style="padding:24px;background:{T.sf};border-radius:{T.rL}">
						<div style="font-family:{T.f};font-weight:{T.fw};font-size:26px;text-transform:{T.tt}">
							Tu reserva
						</div>
						<div style="margin-top:14px;display:flex;align-items:center;gap:12px">
							{#if reserva.profesional?.foto}<img
									src={img(reserva.profesional.foto, 120, 120)}
									alt=""
									style="width:52px;height:52px;border-radius:{T.r};object-fit:cover;flex:none"
								/>{/if}
							<div style="min-width:0">
								<div style="font-size:15px;font-weight:600">
									{reserva.conQuien ? `Con ${reserva.conQuien}` : 'Con el primero libre'}
								</div>
								<div style="font-size:13px;color:{T.mu}">
									{reserva.cuando ?? 'Elegí el día y el horario'}
								</div>
							</div>
						</div>
						{#if reserva.servicio}
							<div
								style="margin-top:16px;padding-top:14px;border-top:1px solid {T.ln};display:flex;flex-direction:column;gap:8px;font-size:14px"
							>
								<span style="display:flex;justify-content:space-between;gap:10px"
									><span>{reserva.servicio.nombre}</span><span style="white-space:nowrap"
										>{fmt(reserva.servicio.precio)}</span
									></span
								><span style="display:flex;justify-content:space-between;color:{T.mu}"
									><span>Duración total</span><span>{reserva.duracion}</span></span
								><span
									style="display:flex;justify-content:space-between;align-items:baseline;margin-top:6px"
									><b>Total</b><b style="font-size:24px">{fmt(reserva.servicio.precio)}</b></span
								>
							</div>
						{/if}
						<div
							style="margin-top:14px;display:flex;align-items:flex-start;gap:8px;font-size:12.5px;line-height:1.45;color:{T.mu}"
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
				style="padding:clamp(44px,6.5cqw,90px) clamp(14px,3.5cqw,44px);max-width:720px;margin:0 auto;text-align:center"
			>
				<span
					style="display:inline-grid;place-items:center;width:64px;height:64px;border-radius:50%;background:{T.acc};color:{T.ai}"
					><Icon icon={cancelado ? I.phCalendarX : I.phCheckBold} size={26} /></span
				>
				<h1
					style="margin:20px 0 0;font-family:{T.f};font-weight:{T.fw};font-size:clamp(40px,6cqw,72px);line-height:1;letter-spacing:{T.ls};text-transform:{T.tt};overflow-wrap:anywhere"
				>
					{cancelado
						? 'Turno cancelado'
						: turno.estado === 'reservado'
							? `Falta la seña, ${turno.nombre}`
							: `¡Listo, ${turno.nombre}!`}
				</h1>
				<p style="margin:14px 0 0;font-size:16px;line-height:1.55;color:{T.mu}">
					{#if cancelado}Liberamos el horario{turno.senaPagada
							? '. La seña te la devuelve la tienda por Mercado Pago'
							: ''}. Si querés, reservá&nbsp;otro.{:else if turno.estado === 'reservado'}Te
						guardamos el horario unos minutos mientras pagás la seña con Mercado&nbsp;Pago.{:else}Tu
						turno está confirmado. Te esperamos en
						{turno.tienda}.{/if}
				</p>
				{#if aviso === 'pendiente' || aviso === 'error'}<div
						role="status"
						style="margin-top:16px;padding:12px 14px;border-radius:{T.r};background:color-mix(in srgb,#B42318 12%,transparent);color:#B42318;font-size:14px"
					>
						{aviso === 'pendiente'
							? 'Mercado Pago todavía no acreditó la seña: te avisamos cuando entre.'
							: 'El pago de la seña no salió. Podés reservar de nuevo.'}
					</div>{/if}
				<div
					style="margin-top:24px;padding:20px;background:{T.sf};border-radius:{T.rL};text-align:left;font-size:14.5px;line-height:1.6"
				>
					<div
						style="font-size:16px;font-weight:600;text-decoration:{cancelado
							? 'line-through'
							: 'none'}"
					>
						{cuandoLargo(turno.inicio)} con {turno.profesional}
					</div>
					<div style="font-size:13.5px;color:{T.mu};margin-top:2px">
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
						style="margin-top:12px;padding:12px 14px;border-radius:{T.r};background:color-mix(in srgb,#B42318 12%,transparent);color:#B42318;font-size:14px"
					>
						{error}
					</div>{/if}
				<div style="margin-top:22px;display:flex;justify-content:center;gap:10px;flex-wrap:wrap">
					{#if !cancelado}<a
							href="{base}/turno/{turno.codigo}/turno.ics"
							download="turno.ics"
							data-sveltekit-reload
							style="display:flex;align-items:center;gap:8px;text-decoration:none;padding:15px 26px;border:0;border-radius:{T.rb};background:{T.ink};color:{T.bg};font:700 15px 'Instrument Sans'"
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
									style="text-decoration:none;padding:15px 26px;border:0;border-radius:{T.rb};background:{T.ink};color:{T.bg};font:700 15px 'Instrument Sans';cursor:pointer"
									>Sí, cancelar el turno</button
								>{:else}<button
									type="button"
									onclick={() => (seguro = true)}
									style="text-decoration:none;padding:15px 26px;border:1.5px solid {T.ink};border-radius:{T.rb};background:transparent;color:{T.ink};font:700 15px 'Instrument Sans';cursor:pointer"
									>Cancelar turno</button
								>{/if}
						</form>
					{:else if cancelado}<a
							href="{base}/turnos"
							style="display:flex;align-items:center;gap:8px;text-decoration:none;padding:15px 26px;border:0;border-radius:{T.rb};background:{T.ink};color:{T.bg};font:700 15px 'Instrument Sans'"
							><Icon icon={I.phCalendarPlusBold} size={17} />Reservar otro</a
						>{/if}
					<a
						href={base || '/'}
						style="text-decoration:none;padding:15px 26px;border:1.5px solid {T.ink};border-radius:{T.rb};background:transparent;color:{T.ink};font:700 15px 'Instrument Sans'"
						>Volver a la tienda</a
					>
				</div>
				{#if !cancelado && !turno.cancelable}<p
						style="margin:14px 0 0;font-size:13.5px;color:{T.mu}"
					>
						Para cambiarlo o cancelarlo, escribile a la&nbsp;tienda{#if whatsapp}&nbsp;<a
								href={wa(`¡Hola! Tengo un turno el ${cuandoLargo(turno.inicio)}.`)}
								target="_blank"
								rel="noopener"
								style="color:{T.ink}">por WhatsApp</a
							>{/if}.
					</p>{/if}
			</section>
		{/if}

		<!-- ============ LISTO (no está en el diseño: con las piezas del checkout) ============ -->
		{#if pagina === 'listo' && pedido}
			<section
				style="padding:clamp(44px,6.5cqw,90px) clamp(14px,3.5cqw,44px);max-width:720px;margin:0 auto;text-align:center"
			>
				<span
					style="display:inline-grid;place-items:center;width:64px;height:64px;border-radius:50%;background:{T.acc};color:{T.ai}"
					><Icon icon={I.phCheckBold} size={28} /></span
				>
				<h1
					style="margin:20px 0 0;font-family:{T.f};font-weight:{T.fw};font-size:clamp(40px,6cqw,72px);line-height:1;letter-spacing:{T.ls};text-transform:{T.tt};overflow-wrap:anywhere"
				>
					¡Gracias, {pedido.nombre}!
				</h1>
				<p style="margin:14px 0 0;font-size:16px;line-height:1.55;color:{T.mu}">
					Recibimos tu pedido <b style="color:{T.ink}">#{pedido.numero}</b> por
					<b style="color:{T.ink}">{fmt(pedido.total)}</b>. Te escribimos por mail con
					cada&nbsp;novedad.
				</p>
				{#if pedido.pago === 'transferencia'}
					<div
						style="margin-top:24px;padding:20px;background:{T.sf};border-radius:{T.rL};text-align:left;font-size:14.5px;line-height:1.6"
					>
						<b>Transferí {fmt(pedido.total)}</b>
						{#if pedido.alias}al {pedido.esCbu ? 'CBU/CVU' : 'alias'}
							<b>{pedido.alias}</b>{:else}con los datos que te mandamos por mail{/if} y mandanos el comprobante{#if whatsapp}&nbsp;<a
								href={wa(`¡Hola! Te mando el comprobante del pedido #${pedido.numero}.`)}
								target="_blank"
								rel="noopener"
								style="color:{T.ink}">por WhatsApp</a
							>{:else}&nbsp;respondiendo el mail del pedido{/if}. Lo preparamos apenas
						se&nbsp;acredite.
					</div>
				{/if}
				<a
					href={base || '/'}
					style="display:inline-block;margin-top:28px;text-decoration:none;padding:15px 26px;border-radius:{T.rb};background:{T.ink};color:{T.bg};font:700 15px 'Instrument Sans'"
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
			<section style="padding:clamp(44px,6.5cqw,90px) clamp(14px,3.5cqw,44px) 0;text-align:center">
				{#if tienda.ciudad}<div
						style="font-size:12.5px;letter-spacing:.22em;text-transform:uppercase;color:{T.mu}"
					>
						{tienda.ciudad}
					</div>{/if}
				<h1
					style="margin:16px auto 0;max-width:920px;font-family:{T.f};font-weight:{T.fw};font-size:clamp(42px,6.6cqw,96px);line-height:.96;letter-spacing:{T.ls};text-transform:{T.tt};overflow-wrap:anywhere"
				>
					{titulo[1] ? `${titulo[0]} ` : titulo[0]}{#if titulo[1]}
						<i style="font-style:{T.fi}">{titulo[1]}</i>{/if}
				</h1>
			</section>
			<section
				style="padding:clamp(36px,5.5cqw,76px) clamp(14px,3.5cqw,44px) 0;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,320px),1fr));gap:clamp(16px,3cqw,40px);align-items:center"
			>
				{#if tienda.portada.foto}<div
						style="aspect-ratio:4/5;border-radius:{T.rL};overflow:hidden;background:{T.sf2}"
					>
						<img
							src={img(tienda.portada.foto, 1000, 1250)}
							alt=""
							style="width:100%;height:100%;object-fit:cover"
						/>
					</div>{/if}
				<div style="max-width:540px">
					<p style="margin:0;font-family:{T.f2};font-size:clamp(21px,2.4cqw,29px);line-height:1.35">
						{tienda.nombre}{tienda.ciudad ? `, desde ${tienda.ciudad}` : ''}. {comoLlega(
							tienda.envios
						)}
					</p>
					{#if tienda.productos.length || tienda.cobros.cuotas > 1 || tienda.promedio}<div
							style="margin-top:26px;display:grid;grid-template-columns:repeat(3,1fr);gap:14px;border-top:1px solid {T.ln};padding-top:20px"
						>
							{#if tienda.productos.length}<div>
									<div
										style="font-family:{T.f};font-weight:{T.fw};font-size:clamp(30px,3.6cqw,44px);line-height:1"
									>
										{tienda.productos.length}
									</div>
									<div style="font-size:13px;color:{T.mu};margin-top:4px">productos</div>
								</div>{/if}
							{#if tienda.cobros.cuotas > 1}<div>
									<div
										style="font-family:{T.f};font-weight:{T.fw};font-size:clamp(30px,3.6cqw,44px);line-height:1"
									>
										{tienda.cobros.cuotas}
									</div>
									<div style="font-size:13px;color:{T.mu};margin-top:4px">cuotas sin interés</div>
								</div>{/if}
							{#if tienda.promedio}<div>
									<div
										style="font-family:{T.f};font-weight:{T.fw};font-size:clamp(30px,3.6cqw,44px);line-height:1"
									>
										{tienda.promedio.toLocaleString('es-AR')}
									</div>
									<div style="font-size:13px;color:{T.mu};margin-top:4px">
										de promedio en&nbsp;reseñas
									</div>
								</div>{/if}
						</div>{/if}
				</div>
			</section>
			<section
				style="margin-top:clamp(52px,7cqw,100px);padding:clamp(36px,5.5cqw,76px) clamp(14px,3.5cqw,44px);background:{T.panel};color:{T.panelInk};display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,320px),1fr));gap:clamp(22px,4cqw,56px)"
			>
				<div>
					<h2
						style="margin:0;font-family:{T.f};font-weight:{T.fw};font-size:clamp(36px,4.6cqw,60px);line-height:1;text-transform:{T.tt}"
					>
						{tienda.envios.retiro ? 'Visitanos' : 'Hablemos'}
					</h2>
					<div
						style="margin-top:20px;display:flex;flex-direction:column;gap:13px;font-size:15px;opacity:.9"
					>
						{#if tienda.direccion}<span style="display:flex;gap:12px"
								><Icon icon={I.phMapPin} size={19} />{tienda.direccion}{tienda.ciudad
									? `, ${tienda.ciudad}`
									: ''}</span
							>{/if}{#if tienda.whatsapp}<span style="display:flex;gap:12px"
								><Icon icon={I.phWhatsappLogo} size={19} />{tienda.whatsapp}</span
							>{/if}
					</div>
				</div>
				{#if whatsapp}
					<div style="display:flex;flex-direction:column;gap:10px;justify-content:center">
						<div style="font-family:{T.f};font-weight:{T.fw};font-size:28px;text-transform:{T.tt}">
							Escribinos
						</div>
						<a
							href={wa('¡Hola!')}
							target="_blank"
							rel="noopener"
							style="display:flex;justify-content:center;align-items:center;gap:8px;padding:15px;border-radius:{T.rb};background:{T.acc};color:{T.ai};font-size:14.5px;font-weight:700;text-decoration:none"
							><Icon icon={I.phWhatsappLogo} size={18} />Por WhatsApp</a
						>
					</div>
				{/if}
			</section>
		{/if}

		<!-- footer -->
		<footer
			style="padding:clamp(44px,6cqw,76px) clamp(14px,3.5cqw,44px) 26px;border-top:1px solid {T.ln}"
		>
			<div
				style="display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,190px),1fr));gap:30px"
			>
				<div style="min-width:0;max-width:400px">
					<div
						style="font-family:{T.f};font-weight:{T.fw};font-size:34px;line-height:1;letter-spacing:{T.ls};text-transform:{T.tt}"
					>
						{tienda.nombre}
					</div>
					<p style="margin:12px 0 0;font-size:14px;color:{T.mu};line-height:1.55">
						{#if tienda.cobros.cuotas > 1}{tienda.cobros.cuotas} cuotas sin interés{#if tienda.cobros.transferencia > 0}&nbsp;y
								{tienda.cobros.transferencia}% off por transferencia{/if}, en toda la&nbsp;tienda.{:else if tienda.cobros.transferencia > 0}{tienda
								.cobros.transferencia}% off por transferencia, en toda la&nbsp;tienda.{:else}Compra
							protegida, con 10 días para&nbsp;arrepentirte.{/if}
					</p>
					<!-- Newsletter (§50): el bloque del pie del diseño; queda pendiente hasta confirmar por mail. -->
					<form
						onsubmit={nl.enviar}
						novalidate
						style="margin-top:12px;display:flex;border-bottom:1.5px solid {T.ink}"
					>
						<input
							class="nl"
							type="email"
							name="email"
							required
							maxlength="120"
							autocomplete="email"
							placeholder="Tu email"
							aria-label="Tu email"
							bind:value={nl.email}
							style="--ph:{T.mu};flex:1;min-width:0;padding:10px 0;border:0;background:transparent;font:inherit;font-size:14.5px;color:inherit;outline:none"
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
								? T.mu
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
							style="font-size:12px;letter-spacing:.12em;text-transform:uppercase;margin-bottom:4px"
							>{f.t}</b
						>{#each f.l as l (l.t)}<a href={l.href} style="color:{T.mu};text-decoration:none"
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
				style="margin-top:36px;padding-top:16px;border-top:1px solid {T.ln};display:flex;justify-content:space-between;gap:14px;flex-wrap:wrap;font-size:12.5px;color:{T.mu}"
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
