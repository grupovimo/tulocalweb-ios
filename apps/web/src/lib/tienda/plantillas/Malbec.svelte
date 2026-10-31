<script lang="ts">
	import { camposFactura, OPCIONALES } from '#lib/tienda/factura.ts';
	import { envia } from '../envio.ts';
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
	import { comoLlega, cotizar, cotizarYa, cpValido, textoEnvios } from '../envio.svelte.ts';
	import { Seleccion, tarjetaDe } from '../seleccion.svelte.ts';
	import { partirTitulo } from '../titulo.ts';
	import type { OpcionEnvio, ProductoPublico, TiendaPublica } from '../tipos.ts';

	/**
	 * Plantilla Malbec (design/Plantilla Malbec.dc.html), con los datos de la tienda en lugar de los de
	 * ejemplo. El aviso de +18 se recuerda en el navegador. Lo que el diseño muestra sin datos atrás
	 * (club de vinos, puntajes, notas y ficha de cata, presentaciones por caja, maridajes, equipo,
	 * caja de regalo, cuenta regresiva) no se inventa: esas partes no se dibujan. El diseño no tiene
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

	// Aviso de +18 (Ley 24.788): se pregunta una vez y se recuerda en este navegador.
	let mayor = $state<boolean | null>(null);
	let menor = $state(false);
	const clave18 = $derived(`mayor18:${tienda.slug}`);
	$effect(() => {
		try {
			mayor = localStorage.getItem(clave18) === '1';
		} catch {
			mayor = false;
		}
	});
	function soyMayor() {
		mayor = true;
		try {
			localStorage.setItem(clave18, '1');
		} catch {
			// Sin almacenamiento (modo privado): se vuelve a preguntar en la próxima visita.
		}
	}

	const fmt = (n: number) => '$ ' + Math.round(n).toLocaleString('es-AR');
	const img = (f: string | null, w: number, h?: number) =>
		!f
			? null
			: /^https?:\/\//.test(f)
				? f
				: `https://images.unsplash.com/photo-${f}?auto=format&fit=crop&w=${w}${h ? '&h=' + h : ''}&q=72`;
	const porId = $derived(new Map(tienda.productos.map((p) => [p.id, p])));
	const rank = $derived(new Map(tienda.masVendidos.map((id, i) => [id, i])));
	// Con variantes de precios distintos, "Desde"; "Sin stock" sólo si no queda ninguna.
	const tarjeta = (p: ProductoPublico) => {
		const t = tarjetaDe(p);
		const rebaja = p.anterior && !t.desde ? p.anterior : null;
		return {
			id: p.id,
			n: p.nombre,
			cat: p.categoria,
			price: (t.desde ? 'Desde ' : '') + fmt(t.precio),
			old: rebaja ? fmt(rebaja) : null,
			tr: fmt(t.precio * (1 - tienda.cobros.transferencia / 100)),
			img: img(p.foto, 600, 800),
			tag: t.agotado
				? 'Sin stock'
				: rebaja
					? `−${Math.round((1 - p.precio / rebaja) * 100)}%`
					: p.etiqueta,
			href: `${base}/producto/${p.slug ?? p.id}`
		};
	};
	const agotado = (p: ProductoPublico) => tarjetaDe(p).agotado;
	const ver = (s: string) => tienda.secciones.includes(s as TiendaPublica['secciones'][number]);
	const whatsapp = $derived(tienda.whatsapp.replace(/\D/g, ''));
	const wa = (texto: string) =>
		`https://wa.me/${whatsapp.startsWith('54') ? whatsapp : '54' + whatsapp}?text=${encodeURIComponent(texto)}`;

	// Título de la portada: la segunda parte va en itálica dorada, como en el diseño.
	const titulo = $derived(partirTitulo(tienda.portada.titulo));
	// Nombre del producto: la última palabra en itálica ("Malbec Valle de Uco <i>2021</i>").
	const nombreFicha = $derived.by(() => {
		const p = (producto?.nombre ?? '').trim().split(/\s+/);
		return p.length > 1 ? [p.slice(0, -1).join(' '), p[p.length - 1]] : [p[0] ?? '', ''];
	});

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
	const best = $derived(
		tienda.masVendidos.slice(0, 6).flatMap((id) => (porId.get(id) ? [tarjeta(porId.get(id)!)] : []))
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
				].map(([t, d]) => ({ t, d, open: abierto === t, i: abierto === t ? I.phMinus : I.phPlus }))
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
	let cp = $state('');
	let cotizado = $state<OpcionEnvio[] | null>(null);
	async function calcular() {
		if (cpValido(cp)) cotizado = await cotizarYa(cp, sel?.precio ?? subtotal, tienda);
	}
	// Añada, tamaño, presentación…: lo elegido, la variante, su precio, stock y foto (seleccion.svelte.ts).
	const sel = $derived(producto ? new Seleccion(producto) : null);
	// El tachado no va con "Desde": no se sabe a qué variante se compara.
	const tachado = $derived(sel && !sel.desde ? sel.anterior : null);
	function agregar() {
		if (!sel || tienda.vitrina || !sel.agregar(carrito)) return;
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- ruta de la tienda con su slug
		goto(`${base}/carrito`);
	}
	/** Sumar desde "Para sumar": con opciones, primero se elige en la ficha. */
	function sumar(p: ProductoPublico) {
		if (tienda.vitrina) return;
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- ruta de la tienda con su slug
		if (!carrito.sumarProducto(p)) goto(`${base}/producto/${p.slug ?? p.id}`);
	}

	// Carrito.
	const lineas = $derived(carrito.detalle(tienda.productos));
	const subtotal = $derived(lineas.reduce((s, l) => s + l.subtotal, 0));
	const falta = $derived(
		tienda.envios.gratisDesde ? Math.max(0, tienda.envios.gratisDesde - subtotal) : 0
	);
	// Las seis casillas del diseño muestran cuánto falta para el envío gratis.
	const casillas = $derived(
		Array.from({ length: 6 }, (_, i) =>
			tienda.envios.gratisDesde && subtotal >= ((i + 1) / 6) * tienda.envios.gratisDesde
				? 'var(--acento, #E8B86A)'
				: 'transparent'
		)
	);
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
	const pagos = $derived([
		...(tienda.cobros.datosTransferencia
			? [
					{
						id: 'transferencia' as const,
						t: 'Transferencia',
						d: tienda.cobros.transferencia
							? `${tienda.cobros.transferencia}% off · alias al confirmar`
							: 'Alias al confirmar',
						i: I.phBank,
						c: '#9FD3A8'
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
	const navL = $derived([
		...(tienda.turnos
			? [
					{
						t: 'Reservar turno',
						href: `${base}/turnos`,
						on: pagina === 'turnos' || pagina === 'turno'
					}
				]
			: []),
		...tienda.categorias.slice(0, 2).map((c) => ({
			t: c.nombre,
			href: enlaceCat(c.nombre),
			on: pagina === 'catalogo' && categoria === c.nombre
		})),
		{ t: 'Nosotros', href: `${base}/nosotros`, on: pagina === 'nosotros' }
	]);
	// Reseñas y preguntas con las piezas de Malbec: fondo oscuro, dorado y Cormorant en los títulos.
	const tema: TemaOpiniones = {
		texto: '#F4E9E1',
		suave: '#BFA9A3',
		campo: '#1E0E12',
		tarjeta: '#24121A',
		borde: 'rgba(244,233,225,.14)',
		acento: 'var(--acento, #E8B86A)',
		sobreAcento: '#1E0E12',
		estrella: 'var(--acento, #E8B86A)',
		radio: '0px',
		fuente: "'Instrument Sans', sans-serif",
		titulos: "'Cormorant Garamond', serif",
		pesoTitulos: 500,
		error: '#FF9C8A'
	};
	const estrellas = (n: number) =>
		Array.from({ length: 5 }, (_, i) =>
			n >= i + 1 ? I.phStarFill : n >= i + 0.5 ? I.phStarHalfFill : I.phStar
		);
	const estilo = (on: boolean) => ({
		bd: on ? 'var(--acento, #E8B86A)' : 'rgba(244,233,225,.14)',
		bg: on ? 'rgba(232,184,106,.08)' : 'transparent'
	});
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
	<link
		href="https://fonts.googleapis.com/css2?family=Cormorant+Garamond:ital,wght@0,500;0,600;1,500;1,600&display=swap"
		rel="stylesheet"
	/>
</svelte:head>

<!-- eslint-disable svelte/no-navigation-without-resolve -- las rutas de la tienda se arman con su slug -->
<div
	class="malbec"
	style="min-height:100vh;font-family:'Instrument Sans',system-ui,sans-serif;color:#F4E9E1;background:#1E0E12"
>
	<div
		style="container-type:inline-size;width:100%;background:#1E0E12;position:relative;overflow-x:clip"
	>
		{#if mayor === false}
			<div
				role="dialog"
				aria-modal="true"
				aria-labelledby="aviso-18"
				style="position:fixed;inset:0;z-index:60;display:grid;place-items:center;padding:20px;background:rgba(18,7,9,.86);backdrop-filter:blur(10px);min-height:100vh"
			>
				<div
					style="width:min(460px,100%);padding:clamp(26px,5cqw,44px);border:1px solid rgba(232,184,106,.3);background:#1E0E12;text-align:center"
				>
					<div
						style="font-family:'Cormorant Garamond',serif;font-size:15px;letter-spacing:.3em;text-transform:uppercase;color:var(--acento, #E8B86A)"
					>
						{tienda.nombre}
					</div>
					<h2
						id="aviso-18"
						style="margin:16px 0 0;font-family:'Cormorant Garamond',serif;font-weight:500;font-size:40px;line-height:1.05"
					>
						¿Sos mayor de 18 años?
					</h2>
					{#if menor}
						<p style="margin:12px 0 0;font-size:14px;line-height:1.55;color:#BFA9A3">
							No podemos venderte alcohol. Te esperamos cuando seas mayor de&nbsp;edad.
						</p>
					{:else}
						<p style="margin:12px 0 0;font-size:14px;line-height:1.55;color:#BFA9A3">
							La venta de alcohol a menores está prohibida (Ley&nbsp;24.788).
						</p>
						<div style="margin-top:22px;display:flex;gap:8px">
							<button
								type="button"
								onclick={soyMayor}
								style="flex:1;padding:15px 10px;border:0;background:var(--acento, #E8B86A);color:#1E0E12;font:700 14px 'Instrument Sans';letter-spacing:.06em;cursor:pointer;white-space:nowrap"
								>Sí, soy mayor</button
							><button
								type="button"
								onclick={() => (menor = true)}
								style="flex:1;padding:15px;border:1px solid rgba(244,233,225,.3);background:transparent;color:#F4E9E1;font:600 14px 'Instrument Sans';cursor:pointer"
								>No</button
							>
						</div>
					{/if}
				</div>
			</div>
		{/if}

		<!-- barra promo -->
		<div
			style="display:flex;justify-content:center;align-items:center;gap:14px;flex-wrap:wrap;padding:9px 16px;background:#9E2F40;color:#F4E9E1;font-size:13px;text-align:center"
		>
			{#if conteo.actual}<span>{conteo.texto}</span><span
					style="display:flex;gap:4px;font-variant-numeric:tabular-nums"
					>{#each conteo.partes(4) as c (c.u)}<span
							style="padding:2px 7px;background:rgba(0,0,0,.25)">{c.v}{c.u}</span
						>{/each}</span
				>{:else}<span
					>{tienda.envios.gratisDesde
						? `Envío gratis desde ${fmt(tienda.envios.gratisDesde)}`
						: tienda.cobros.cuotas > 1
							? `${tienda.cobros.cuotas} cuotas sin interés`
							: tienda.cobros.transferencia > 0
								? ''
								: 'Compra protegida'}{#if tienda.cobros.transferencia > 0}{#if tienda.envios.gratisDesde || tienda.cobros.cuotas > 1}
							·
						{/if}{tienda.cobros.transferencia}% off por transferencia{/if}</span
				>{/if}
		</div>

		<!-- header -->
		<header
			style="position:sticky;top:0;z-index:20;display:grid;grid-template-columns:1fr auto 1fr;align-items:center;gap:16px;padding:16px clamp(16px,4cqw,48px);background:rgba(30,14,18,.92);backdrop-filter:blur(10px);border-bottom:1px solid rgba(244,233,225,.1)"
		>
			<div style="display:flex;gap:24px;align-items:center;min-width:0">
				{#if nar}<a href="{base}/catalogo" aria-label="Catálogo" style="display:grid;color:#F4E9E1"
						><Icon icon={I.phList} size={22} /></a
					>{/if}{#if nar && tienda.turnos}<a
						href="{base}/turnos"
						aria-label="Reservar turno"
						style="display:grid;flex:none;color:#F4E9E1"
						><Icon icon={I.phCalendarPlusBold} size={22} /></a
					>{/if}{#if wide}{#each navL as n (n.t)}<a
							href={n.href}
							style="text-decoration:none;padding:4px 0;font:500 13.5px 'Instrument Sans';letter-spacing:.08em;text-transform:uppercase;color:#F4E9E1;border-bottom:1px solid {n.on
								? 'var(--acento, #E8B86A)'
								: 'transparent'};white-space:nowrap">{n.t}</a
						>{/each}{/if}
			</div>
			<a
				href={base || '/'}
				style="text-decoration:none;color:#F4E9E1;text-align:center;line-height:1;min-width:0"
				><span
					style="display:block;font-family:'Cormorant Garamond',serif;font-weight:600;font-size:clamp(24px,2.6cqw,32px);letter-spacing:.04em"
					><LogoTienda {tienda} /></span
				>{#if tienda.ciudad}<span
						style="display:block;margin-top:3px;font-size:10px;letter-spacing:.34em;text-transform:uppercase;color:var(--acento, #E8B86A)"
						>{tienda.ciudad}</span
					>{/if}</a
			>
			<div style="display:flex;gap:18px;align-items:center;justify-content:flex-end">
				<a
					href="{base}/carrito"
					aria-label="Carrito"
					style="position:relative;color:#F4E9E1;display:grid"
					><Icon icon={I.phWine} size={22} />{#if carrito.unidades}<span
							style="position:absolute;top:-6px;right:-8px;min-width:17px;height:17px;border-radius:999px;background:var(--acento, #E8B86A);color:#1E0E12;font-size:10.5px;font-weight:700;display:grid;place-items:center"
							>{carrito.unidades}</span
						>{/if}</a
				>
			</div>
		</header>

		{#if tienda.vitrina}
			<div
				style="padding:10px 16px;background:#24121A;border-bottom:1px solid rgba(232,184,106,.3);font-size:13px;text-align:center;color:#DCCBC4"
			>
				Es una tienda de ejemplo de la plantilla Malbec: podés recorrerla, pero no se compra de
				verdad.
			</div>
		{/if}

		<!-- ============ INICIO ============ -->
		{#if pagina === 'inicio'}
			<!-- Las secciones en el orden que eligió la tienda en Mi tienda. -->
			{#each enOrden( tienda.secciones, ['hero', 'prods', 'cats', 'revs', 'ig'] ) as seccion (seccion)}
				{#if seccion === 'hero'}
					{#if ver('hero')}
						<section
							style="position:relative;min-height:clamp(560px,64cqw,860px);overflow:hidden;background:#2A161B"
						>
							{#if tienda.portada.foto}<img
									src={img(tienda.portada.foto, 1800)}
									alt=""
									style="position:absolute;inset:0;width:100%;height:100%;object-fit:cover;opacity:.55"
								/>{/if}
							<div
								style="position:absolute;inset:0;background:radial-gradient(120% 90% at 30% 70%,rgba(30,14,18,.2),rgba(30,14,18,.92) 75%)"
							></div>
							<div
								style="position:relative;padding:clamp(60px,10cqw,140px) clamp(16px,5cqw,72px) clamp(40px,6cqw,80px);max-width:860px"
							>
								{#if tienda.ciudad}<div
										style="font-size:12.5px;letter-spacing:.32em;text-transform:uppercase;color:var(--acento, #E8B86A)"
									>
										{tienda.ciudad}
									</div>{/if}
								<h1
									style="margin:18px 0 0;font-family:'Cormorant Garamond',serif;font-weight:500;font-size:clamp(54px,9.4cqw,140px);line-height:.9;letter-spacing:-.015em;overflow-wrap:anywhere"
								>
									{titulo[0]}{#if titulo[1]}<br /><i style="color:var(--acento, #E8B86A)"
											>{titulo[1]}</i
										>{/if}
								</h1>
								{#if tienda.portada.lema}<p
										style="margin:22px 0 0;font-size:17px;line-height:1.55;color:#DCCBC4;max-width:460px"
									>
										{tienda.portada.lema}
									</p>{/if}
								<div style="margin-top:28px;display:flex;gap:10px;flex-wrap:wrap">
									<a
										href="{base}/catalogo"
										style="padding:16px 28px;border:0;background:var(--acento, #E8B86A);color:#1E0E12;font:700 14px 'Instrument Sans';letter-spacing:.08em;text-transform:uppercase;text-decoration:none"
										>Ver la selección</a
									>{#if whatsapp}<a
											href={wa('¡Hola! Quiero una recomendación.')}
											target="_blank"
											rel="noopener"
											style="padding:16px 24px;border:1px solid rgba(244,233,225,.45);background:transparent;color:#F4E9E1;font:600 14px 'Instrument Sans';letter-spacing:.08em;text-transform:uppercase;text-decoration:none"
											>Pedí una recomendación</a
										>{/if}
								</div>
							</div>
						</section>
					{/if}

					<div
						style="display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));border-bottom:1px solid rgba(244,233,225,.1)"
					>
						{#each perks as p (p.t)}<div
								style="display:flex;align-items:center;gap:12px;padding:18px clamp(16px,3cqw,36px);border-right:1px solid rgba(244,233,225,.08)"
							>
								<Icon icon={p.i} size={22} style="color:var(--acento, #E8B86A)" />
								<div>
									<div style="font-size:13.5px;font-weight:600">{p.t}</div>
									<div style="font-size:12.5px;color:#BFA9A3">{p.d}</div>
								</div>
							</div>{/each}
					</div>
				{:else if seccion === 'prods'}
					{#if ver('prods') && best.length}
						<section style="padding:clamp(56px,8cqw,110px) clamp(16px,4cqw,48px) 0">
							<div
								style="display:flex;justify-content:space-between;align-items:end;gap:16px;flex-wrap:wrap"
							>
								<div>
									<div
										style="font-size:12.5px;letter-spacing:.3em;text-transform:uppercase;color:var(--acento, #E8B86A)"
									>
										Los más vendidos
									</div>
									<h2
										style="margin:10px 0 0;font-family:'Cormorant Garamond',serif;font-weight:500;font-size:clamp(40px,5.4cqw,70px);line-height:1"
									>
										Los más elegidos
									</h2>
								</div>
								<a
									href="{base}/catalogo"
									style="text-decoration:none;padding:0;font:600 13px 'Instrument Sans';letter-spacing:.1em;text-transform:uppercase;color:#F4E9E1;border-bottom:1px solid var(--acento, #E8B86A)"
									>Ver {tienda.productos.length === 1
										? 'el producto'
										: `los ${tienda.productos.length} productos`}</a
								>
							</div>
							<div
								style="margin-top:30px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(46%,220px),1fr));gap:clamp(12px,1.8cqw,22px)"
							>
								{#each best as p (p.id)}
									<a
										href={p.href}
										class="caja"
										style="border:1px solid rgba(244,233,225,.1);padding:0;background:#24121A;text-align:left;color:#F4E9E1;font-family:inherit;display:flex;flex-direction:column;transition:border-color .3s;text-decoration:none"
									>
										<span
											style="position:relative;display:block;aspect-ratio:3/4;overflow:hidden;background:linear-gradient(180deg,#3A2026,#24121A)"
											>{#if p.img}<img
													src={p.img}
													alt=""
													style="width:100%;height:100%;object-fit:cover;display:block"
												/>{/if}{#if p.tag}<span
													style="position:absolute;top:10px;left:10px;padding:4px 9px;background:#9E2F40;color:#F4E9E1;font-size:11px;font-weight:700;letter-spacing:.08em;text-transform:uppercase"
													>{p.tag}</span
												>{/if}</span
										>
										<span style="display:block;padding:14px 14px 16px"
											>{#if p.cat}<span
													style="display:block;font-size:11.5px;letter-spacing:.14em;text-transform:uppercase;color:#BFA9A3"
													>{p.cat}</span
												>{/if}<span
												style="display:block;margin-top:6px;font-family:'Cormorant Garamond',serif;font-size:23px;font-weight:600;line-height:1.05"
												>{p.n}</span
											><span
												style="display:flex;justify-content:space-between;align-items:baseline;gap:8px;flex-wrap:wrap;margin-top:12px"
												><b style="font-size:16px">{p.price}</b
												>{#if tienda.cobros.transferencia > 0}<span
														style="font-size:12px;color:#9FD3A8;font-weight:600"
														>{p.tr} transf.</span
													>{/if}</span
											></span
										>
									</a>
								{/each}
							</div>
						</section>
					{/if}
				{:else if seccion === 'cats'}
					{#if ver('cats') && tienda.categorias.length}
						<section style="padding:clamp(56px,8cqw,110px) clamp(16px,4cqw,48px) 0">
							<div style="text-align:center">
								<div
									style="font-size:12.5px;letter-spacing:.3em;text-transform:uppercase;color:var(--acento, #E8B86A)"
								>
									Categorías
								</div>
								<h2
									style="margin:10px 0 0;font-family:'Cormorant Garamond',serif;font-weight:500;font-size:clamp(40px,5.4cqw,70px);line-height:1"
								>
									Elegí por dónde empezar
								</h2>
							</div>
							<div
								style="margin-top:28px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(46%,180px),1fr));gap:10px"
							>
								{#each tienda.categorias.slice(0, 5) as c (c.nombre)}<a
										href={enlaceCat(c.nombre)}
										class="zoom"
										style="position:relative;display:block;aspect-ratio:1;border:0;padding:0;overflow:hidden;background:#3A2026;text-align:left;text-decoration:none"
										>{#if c.foto}<img
												src={img(c.foto, 500, 500)}
												alt=""
												style="position:absolute;inset:0;width:100%;height:100%;object-fit:cover;opacity:.75"
											/>{/if}<span
											style="position:absolute;inset:auto 0 0 0;padding:16px;background:linear-gradient(0deg,rgba(30,14,18,.9),transparent);color:#F4E9E1"
											><span
												style="display:block;font-family:'Cormorant Garamond',serif;font-size:26px;font-weight:600;line-height:1"
												>{c.nombre}</span
											><span
												style="display:block;margin-top:4px;font-size:12px;color:var(--acento, #E8B86A);letter-spacing:.06em"
												>{c.cantidad} {c.cantidad === 1 ? 'producto' : 'productos'}</span
											></span
										></a
									>{/each}
							</div>
						</section>
					{/if}
				{:else if seccion === 'revs'}
					{#if ver('revs') && tienda.resenas.length}
						<section style="padding:clamp(56px,8cqw,110px) clamp(16px,4cqw,48px) 0">
							<div
								style="display:flex;justify-content:space-between;align-items:end;gap:16px;flex-wrap:wrap"
							>
								<h2
									style="margin:0;font-family:'Cormorant Garamond',serif;font-weight:500;font-size:clamp(40px,5.4cqw,70px);line-height:1"
								>
									Lo que dicen
								</h2>
								{#if tienda.promedio}<span
										style="display:flex;align-items:center;gap:8px;font-size:14px"
										><span style="display:flex;color:var(--acento, #E8B86A)"
											>{#each estrellas(tienda.promedio) as e, i (i)}<Icon
													icon={e}
													size={15}
												/>{/each}</span
										><b>{tienda.promedio.toLocaleString('es-AR')}</b><span style="color:#BFA9A3"
											>· {tienda.resenas.length}
											{tienda.resenas.length === 1 ? 'reseña' : 'reseñas'}</span
										></span
									>{/if}
							</div>
							<div
								style="margin-top:26px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,290px),1fr));gap:14px"
							>
								{#each tienda.resenas.slice(0, 3) as r, i (i)}
									{@const pr = tienda.productos.find((p) => p.nombre === r.producto)}
									<div
										style="display:flex;gap:14px;padding:16px;border:1px solid rgba(244,233,225,.1);background:#24121A"
									>
										{#if pr?.foto}<img
												src={img(pr.foto, 200, 260)}
												alt=""
												style="width:88px;height:116px;object-fit:cover;flex:none;background:#3A2026"
											/>{/if}
										<div style="min-width:0">
											<p
												style="margin:0;font-family:'Cormorant Garamond',serif;font-size:21px;line-height:1.25;font-style:italic"
											>
												“{r.texto}”
											</p>
											<div style="margin-top:10px;font-size:12.5px;color:#BFA9A3">
												{r.autor}{#if r.producto}
													· {r.conCompra ? 'compró' : 'sobre'} {r.producto}{/if}
											</div>
											{#if r.conCompra}<div
													style="margin-top:6px;display:flex;align-items:center;gap:4px;font-size:12px;font-weight:600;color:#9FD3A8"
												>
													<Icon icon={I.phSealCheckFill} size={13} />Compra verificada
												</div>{/if}
											{#if r.respuesta}<div
													style="margin-top:10px;padding-left:10px;border-left:1px solid var(--acento, #E8B86A);font-size:13px;line-height:1.5;color:#BFA9A3"
												>
													<b style="color:var(--acento, #E8B86A)">Respuesta de {tienda.nombre}</b
													><br />{r.respuesta}
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
								style="margin:0;font-family:'Cormorant Garamond',serif;font-weight:500;font-size:clamp(40px,5.4cqw,70px);line-height:1"
							>
								En la mesa de todos
							</h2>
							<div
								style="margin-top:22px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(30%,160px),1fr));gap:6px"
							>
								{#each tienda.fotos as g (g)}<div
										style="aspect-ratio:1;overflow:hidden;background:#3A2026"
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
			<section style="padding:28px clamp(16px,4cqw,48px) clamp(56px,8cqw,100px)">
				<div style="font-size:12.5px;color:#BFA9A3">
					<a href={base || '/'} style="color:inherit;text-decoration:none">Inicio</a> /
					<span style="color:#F4E9E1">{categoria ?? 'Todo'}</span>
				</div>
				<div
					style="margin-top:12px;display:flex;justify-content:space-between;align-items:end;gap:16px;flex-wrap:wrap"
				>
					<h1
						style="margin:0;font-family:'Cormorant Garamond',serif;font-weight:500;font-size:clamp(50px,7.6cqw,104px);line-height:.92;overflow-wrap:anywhere"
					>
						{categoria ?? 'Todo'}
					</h1>
					<span style="font-size:13.5px;color:#BFA9A3"
						>{listado.length} {listado.length === 1 ? 'producto' : 'productos'}</span
					>
				</div>
				<div style="margin-top:20px;display:flex;gap:8px;overflow-x:auto;scrollbar-width:none">
					{#each [null, ...tienda.categorias.map((c) => c.nombre)] as c (c ?? '')}{@const on =
							c === categoria}<a
							href={enlaceCat(c)}
							style="flex:none;padding:9px 15px;border:1px solid {on
								? 'var(--acento, #E8B86A)'
								: 'rgba(244,233,225,.2)'};background:{on
								? 'var(--acento, #E8B86A)'
								: 'transparent'};color:{on
								? '#1E0E12'
								: '#F4E9E1'};font-size:13px;font-weight:600;white-space:nowrap;text-decoration:none"
							>{c ?? 'Todos'}</a
						>{/each}
				</div>
				<div style="margin-top:22px;display:flex;gap:clamp(20px,3cqw,40px);align-items:flex-start">
					{#if wide}
						<aside
							style="width:240px;flex:none;position:sticky;top:96px;display:flex;flex-direction:column;gap:26px;font-size:13.5px"
						>
							<div>
								<div
									style="font-size:11.5px;font-weight:700;letter-spacing:.16em;text-transform:uppercase;color:var(--acento, #E8B86A);margin-bottom:12px"
								>
									Ordenar
								</div>
								<div style="display:flex;flex-direction:column;gap:10px">
									{#each ORDENES as [k, t] (k)}{@const on = orden === k}<a
											href={enlaceCat(categoria, k)}
											aria-current={on ? 'true' : undefined}
											style="display:flex;align-items:center;gap:10px;color:#F4E9E1;text-decoration:none"
											><span
												style="width:16px;height:16px;border:1px solid {on
													? 'var(--acento, #E8B86A)'
													: 'rgba(244,233,225,.35)'};background:{on
													? 'var(--acento, #E8B86A)'
													: 'transparent'};display:grid;place-items:center;color:#1E0E12;flex:none"
												>{#if on}<Icon icon={I.phCheckBold} size={10} />{/if}</span
											>{t}</a
										>{/each}
								</div>
							</div>
							{#if tienda.categorias.length}<div>
									<div
										style="font-size:11.5px;font-weight:700;letter-spacing:.16em;text-transform:uppercase;color:var(--acento, #E8B86A);margin-bottom:12px"
									>
										Categoría
									</div>
									<div style="display:flex;flex-direction:column;gap:10px">
										{#each tienda.categorias as c (c.nombre)}{@const on = categoria === c.nombre}<a
												href={enlaceCat(on ? null : c.nombre)}
												aria-current={on ? 'true' : undefined}
												style="display:flex;align-items:center;gap:10px;color:#F4E9E1;text-decoration:none"
												><span
													style="width:16px;height:16px;border:1px solid {on
														? 'var(--acento, #E8B86A)'
														: 'rgba(244,233,225,.35)'};background:{on
														? 'var(--acento, #E8B86A)'
														: 'transparent'};display:grid;place-items:center;color:#1E0E12;flex:none"
													>{#if on}<Icon icon={I.phCheckBold} size={10} />{/if}</span
												>{c.nombre}<span style="margin-left:auto;color:#8F7A75">{c.cantidad}</span
												></a
											>{/each}
									</div>
								</div>{/if}
						</aside>
					{/if}
					<div style="flex:1;min-width:0">
						<div
							style="display:flex;justify-content:space-between;align-items:center;gap:10px;flex-wrap:wrap;padding-bottom:14px;border-bottom:1px solid rgba(244,233,225,.1);font-size:13.5px"
						>
							<span style="display:flex;gap:6px;flex-wrap:wrap"
								>{#if categoria}<a
										href={enlaceCat(null)}
										aria-label="Quitar el filtro {categoria}"
										style="display:flex;align-items:center;gap:6px;padding:7px 11px;background:#3A2026;color:#F4E9E1;text-decoration:none"
										>{categoria}<Icon icon={I.phX} size={11} /></a
									>{/if}</span
							>{#if nar}<label
									style="display:flex;align-items:center;gap:6px;font-weight:600;position:relative"
									><span class="oculto">Ordenar</span><select
										value={orden}
										onchange={(e) =>
											goto(enlaceCat(categoria, e.currentTarget.value), { replace: true })}
										style="appearance:none;border:0;background:transparent;padding:0 20px 0 0;font:600 13.5px 'Instrument Sans';color:#F4E9E1;cursor:pointer"
										>{#each ORDENES as [k, t] (k)}<option value={k} style="color:#1E0E12"
												>{t}</option
											>{/each}</select
									><Icon
										icon={I.phCaretDown}
										size={13}
										style="position:absolute;right:0;pointer-events:none"
									/></label
								>{/if}
						</div>
						{#if catalog.length}
							<div
								style="margin-top:22px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(46%,210px),1fr));gap:clamp(12px,1.8cqw,20px)"
							>
								{#each catalog as p (p.id)}
									<a
										href={p.href}
										class="caja"
										style="display:block;border:1px solid rgba(244,233,225,.1);padding:0;background:#24121A;text-align:left;color:#F4E9E1;font-family:inherit;text-decoration:none;transition:border-color .3s"
									>
										<span
											style="position:relative;display:block;aspect-ratio:3/4;overflow:hidden;background:#3A2026"
											>{#if p.img}<img
													src={p.img}
													alt=""
													style="width:100%;height:100%;object-fit:cover;display:block"
												/>{/if}{#if p.tag}<span
													style="position:absolute;top:10px;right:10px;padding:4px 8px;background:rgba(30,14,18,.85);border:1px solid var(--acento, #E8B86A);color:var(--acento, #E8B86A);font-size:12px;font-weight:700"
													>{p.tag}</span
												>{/if}</span
										>
										<span style="display:block;padding:12px 12px 14px"
											>{#if p.cat}<span
													style="display:block;font-size:11px;letter-spacing:.14em;text-transform:uppercase;color:#BFA9A3"
													>{p.cat}</span
												>{/if}<span
												style="display:block;margin-top:4px;font-family:'Cormorant Garamond',serif;font-size:21px;font-weight:600;line-height:1.05"
												>{p.n}</span
											><span
												style="display:flex;justify-content:space-between;align-items:baseline;gap:8px;margin-top:10px"
												><b style="font-size:15px">{p.price}</b>{#if p.old}<s
														style="font-size:11.5px;color:#BFA9A3">{p.old}</s
													>{/if}</span
											></span
										>
									</a>
								{/each}
							</div>
							{#if listado.length > mostrar}
								<div style="margin-top:40px;display:flex;justify-content:center">
									<button
										type="button"
										onclick={() => (mostrar += 24)}
										style="padding:15px 26px;border:1px solid rgba(244,233,225,.45);background:transparent;color:#F4E9E1;font:600 13px 'Instrument Sans';letter-spacing:.1em;text-transform:uppercase;cursor:pointer"
										>Ver más</button
									>
								</div>
							{/if}
						{:else}
							<p style="margin:40px 0;font-size:15px;color:#BFA9A3">
								Todavía no hay productos en esta categoría.
							</p>
						{/if}
					</div>
				</div>
			</section>
		{/if}

		<!-- ============ PRODUCTO ============ -->
		{#if pagina === 'producto' && producto && sel}
			<section style="padding:24px clamp(16px,4cqw,48px) 0">
				<div style="font-size:12.5px;color:#BFA9A3">
					<a href={base || '/'} style="color:inherit;text-decoration:none">Inicio</a> / {#if producto.categoria}<a
							href={enlaceCat(producto.categoria)}
							style="color:inherit;text-decoration:none">{producto.categoria}</a
						> /
					{/if}<span style="color:#F4E9E1">{producto.nombre}</span>
				</div>
				<div
					style="margin-top:18px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,400px),1fr));gap:clamp(24px,4.4cqw,64px);align-items:start"
				>
					<div style="display:flex;flex-direction:column;gap:10px;min-width:0">
						<div
							style="position:relative;aspect-ratio:4/5;overflow:hidden;background:radial-gradient(70% 60% at 50% 40%,#4A2630,#24121A)"
						>
							{#if sel.foto}<img
									src={img(sel.foto, 1100)}
									alt={producto.nombre}
									style="width:100%;height:100%;object-fit:cover"
								/>{/if}{#if tachado}<span
									style="position:absolute;top:16px;right:16px;width:72px;height:72px;border-radius:50%;border:1px solid var(--acento, #E8B86A);background:rgba(30,14,18,.82);display:grid;place-items:center;text-align:center;line-height:1"
									><span
										><b
											style="display:block;font-family:'Cormorant Garamond',serif;font-size:30px;color:var(--acento, #E8B86A)"
											>{Math.round((1 - sel.precio / tachado) * 100)}%</b
										><span style="font-size:9px;letter-spacing:.14em">OFF</span></span
									></span
								>{:else if producto.etiqueta}<span
									style="position:absolute;top:16px;left:16px;padding:4px 9px;background:#9E2F40;color:#F4E9E1;font-size:11px;font-weight:700;letter-spacing:.08em;text-transform:uppercase"
									>{producto.etiqueta}</span
								>{/if}
						</div>
						{#if sel.fotos.length > 1}<div
								style="display:grid;grid-template-columns:repeat(5,minmax(0,1fr));gap:8px"
							>
								{#each sel.fotos as f, i (f)}<button
										type="button"
										aria-label="Ver la foto {i + 1}"
										aria-pressed={i === sel.fotoIdx}
										onclick={() => sel.verFoto(i)}
										style="aspect-ratio:4/5;padding:0;border:1px solid {i === sel.fotoIdx
											? 'var(--acento, #E8B86A)'
											: 'rgba(244,233,225,.12)'};background:#3A2026;overflow:hidden;cursor:pointer"
										><img
											src={img(f, 200, 250)}
											alt=""
											style="display:block;width:100%;height:100%;object-fit:cover"
										/></button
									>{/each}
							</div>{/if}
					</div>
					<div style="position:sticky;top:96px;min-width:0">
						{#if producto.categoria}<div
								style="font-size:12.5px;letter-spacing:.24em;text-transform:uppercase;color:var(--acento, #E8B86A)"
							>
								{producto.categoria}
							</div>{/if}
						<h1
							style="margin:12px 0 0;font-family:'Cormorant Garamond',serif;font-weight:500;font-size:clamp(42px,5cqw,66px);line-height:.98;overflow-wrap:anywhere"
						>
							{nombreFicha[0]}{nombreFicha[1] ? ' ' : ''}{#if nombreFicha[1]}<i>{nombreFicha[1]}</i
								>{/if}
						</h1>
						<div style="margin-top:18px;display:flex;align-items:baseline;gap:12px">
							<b style="font-size:28px">{sel.desde ? 'Desde ' : ''}{fmt(sel.precio)}</b
							>{#if tachado}<s style="font-size:14px;color:#BFA9A3">{fmt(tachado)}</s>{/if}
						</div>
						<div style="margin-top:6px;display:flex;flex-direction:column;gap:3px;font-size:14px">
							{#if tienda.cobros.cuotas > 1}<span
									>{tienda.cobros.cuotas} cuotas sin interés de
									<b>{fmt(sel.precio / tienda.cobros.cuotas)}</b></span
								>{/if}{#if tienda.cobros.transferencia > 0}<span
									style="color:#9FD3A8;font-weight:600"
									>{fmt(sel.precio * (1 - tienda.cobros.transferencia / 100))} por transferencia ({tienda
										.cobros.transferencia}% off)</span
								>{/if}
						</div>
						{#each sel.opciones as o (o.nombre)}{@const swatch = o.valores.every(
								(x) => x.color || x.foto
							)}
							<div style="margin-top:20px">
								<div
									style="font-size:12px;font-weight:700;letter-spacing:.16em;text-transform:uppercase;margin-bottom:10px;color:{sel.aviso &&
									!o.elegido
										? 'var(--acento, #E8B86A)'
										: '#BFA9A3'}"
								>
									{o.nombre}{#if o.elegido}<span
											style="font-weight:400;letter-spacing:.04em;text-transform:none;color:#F4E9E1"
											>&nbsp;· {o.elegido}</span
										>{/if}
								</div>
								{#if swatch}<div style="display:flex;gap:10px;flex-wrap:wrap">
										{#each o.valores as x (x.valor)}<button
												type="button"
												aria-label="{o.nombre} {x.valor}"
												aria-pressed={x.elegido}
												disabled={!x.disponible && !x.elegido}
												onclick={() => sel.elegir(o.nombre, x.valor)}
												style="position:relative;width:38px;height:38px;padding:0;border:0;border-radius:50%;overflow:hidden;background:#3A2026;cursor:pointer;opacity:{x.disponible
													? 1
													: 0.4};box-shadow:inset 0 0 0 1px rgba(244,233,225,.2),0 0 0 {x.elegido
													? '2px'
													: '0px'} #1E0E12,0 0 0 {x.elegido ? '3px' : '0px'} var(--acento, #E8B86A)"
												>{#if x.color}<span
														style="display:block;width:100%;height:100%;background:{x.color}"
													></span>{:else}<img
														src={img(x.foto, 80, 80)}
														alt=""
														style="display:block;width:100%;height:100%;object-fit:cover"
													/>{/if}{#if !x.disponible}<span
														style="position:absolute;left:50%;top:-4px;bottom:-4px;width:1px;background:#F4E9E1;transform:rotate(45deg)"
													></span>{/if}</button
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
												style="min-width:54px;min-height:46px;padding:10px 16px;border:1px solid {x.elegido
													? 'var(--acento, #E8B86A)'
													: 'rgba(244,233,225,.2)'};background:{x.elegido
													? 'rgba(232,184,106,.1)'
													: 'transparent'};font:600 14px 'Instrument Sans';color:{x.elegido
													? 'var(--acento, #E8B86A)'
													: x.disponible
														? '#F4E9E1'
														: '#BFA9A3'};text-align:left;text-decoration:{x.disponible
													? 'none'
													: 'line-through'};cursor:{x.disponible ? 'pointer' : 'default'}"
												>{x.valor}{#if pv}<span
														style="display:block;font-size:12px;font-weight:400;color:#BFA9A3"
														>{fmt(pv.precio ?? producto.precio)}</span
													>{/if}</button
											>{/each}
									</div>{/if}
							</div>{/each}
						{#if sel.stock > 0 && sel.stock <= 3 && (!sel.conOpciones || sel.variante)}<div
								style="margin-top:14px;font-size:12.5px;letter-spacing:.06em;color:var(--acento, #E8B86A)"
							>
								{sel.stock === 1 ? 'Queda 1' : `Quedan ${sel.stock}`}{#if sel.variante}
									en {sel.texto.toLowerCase()}{/if}
							</div>{/if}
						<div style="margin-top:20px;display:flex;flex-direction:column;gap:8px">
							<button
								type="button"
								onclick={agregar}
								disabled={sel.stock === 0 || tienda.vitrina}
								style="padding:17px;border:0;background:var(--acento, #E8B86A);color:#1E0E12;font:700 14.5px 'Instrument Sans';letter-spacing:.08em;text-transform:uppercase;cursor:pointer"
								>{sel.boton('Agregar al carrito')}</button
							>{#if whatsapp}<a
									href={wa(
										`¡Hola! Quiero comprar ${producto.nombre}${sel.variante ? ` (${sel.texto})` : ''} (${fmt(sel.precio)}).`
									)}
									target="_blank"
									rel="noopener"
									style="display:flex;justify-content:center;align-items:center;gap:8px;padding:15px;border:1px solid rgba(244,233,225,.4);font-size:14px;font-weight:600;color:#F4E9E1;text-decoration:none"
									><Icon icon={I.phWhatsappLogo} size={18} />Consultanos por WhatsApp</a
								>{/if}
							<LanzamientoFicha {tienda} {producto} {sel} {tema} />
							<CompartirProducto {tienda} {producto} {tema} />
						</div>
						<div
							style="margin-top:22px;padding:4px 18px;border:1px solid rgba(244,233,225,.12);background:#24121A"
						>
							{#each acc as a, k (a.t)}<div
									style="padding:14px 0;border-top:{k ? '1px solid rgba(244,233,225,.1)' : '0'}"
								>
									<button
										type="button"
										onclick={() => (abierto = a.open ? '' : a.t)}
										aria-expanded={a.open}
										style="width:100%;display:flex;justify-content:space-between;align-items:center;padding:0;border:0;background:transparent;font:700 11.5px 'Instrument Sans';letter-spacing:.16em;text-transform:uppercase;color:var(--acento, #E8B86A);cursor:pointer;text-align:left"
										>{a.t}<Icon icon={a.i} size={13} /></button
									>{#if a.open}<p
											style="margin:10px 0 0;font-size:14px;line-height:1.55;color:#DCCBC4"
										>
											{a.d}
										</p>{/if}
								</div>{/each}
						</div>
						{#if envia(tienda.envios)}<div style="margin-top:12px;padding:16px;background:#24121A">
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
										style="flex:1;min-width:0;padding:11px 12px;background:#1E0E12;font:400 14px 'Instrument Sans';color:#F4E9E1;border:1px solid rgba(244,233,225,.12);border-radius:0"
										onkeydown={(e) => e.key === 'Enter' && calcular()}
									/><button
										type="button"
										onclick={calcular}
										style="padding:11px 16px;border:0;background:#F4E9E1;color:#1E0E12;font:700 13px 'Instrument Sans';cursor:pointer"
										>Calcular</button
									>
								</div>
								{#if cotizado}<div
										style="margin-top:10px;display:flex;flex-direction:column;gap:6px;font-size:13.5px"
									>
										{#each cotizado as o (o.id)}<span
												style="display:flex;justify-content:space-between;gap:10px"
												><span>{o.nombre} · {o.detalle}</span><b
													style="white-space:nowrap;color:{o.precio ? '#F4E9E1' : '#9FD3A8'}"
													>{o.precio ? fmt(o.precio) : 'Gratis'}</b
												></span
											>{/each}
									</div>{/if}
							</div>{/if}
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
						style="margin:0;font-family:'Cormorant Garamond',serif;font-weight:500;font-size:clamp(36px,4.6cqw,58px);line-height:1"
					>
						Si te gusta este, probá
					</h2>
					<div
						style="margin-top:24px;display:grid;grid-template-columns:repeat(auto-fill,minmax(min(46%,210px),1fr));gap:14px"
					>
						{#each related as p (p.id)}<a
								href={p.href}
								style="display:block;border:0;padding:0;background:transparent;text-align:left;color:#F4E9E1;font-family:inherit;text-decoration:none"
								><span style="display:block;aspect-ratio:3/4;overflow:hidden;background:#3A2026"
									>{#if p.img}<img
											src={p.img}
											alt=""
											style="width:100%;height:100%;object-fit:cover"
										/>{/if}</span
								><span
									style="display:block;margin-top:10px;font-family:'Cormorant Garamond',serif;font-size:20px;font-weight:600"
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
			<section style="padding:32px clamp(16px,4cqw,48px) clamp(56px,8cqw,100px)">
				<h1
					style="margin:0;font-family:'Cormorant Garamond',serif;font-weight:500;font-size:clamp(46px,6.4cqw,86px);line-height:1"
				>
					Tu selección
				</h1>
				{#if !lineas.length}
					<p style="margin:20px 0 0;font-size:15px;color:#BFA9A3">Todavía no sumaste nada.</p>
					<a
						href="{base}/catalogo"
						style="display:inline-block;margin-top:18px;padding:16px 28px;background:var(--acento, #E8B86A);color:#1E0E12;font:700 14px 'Instrument Sans';letter-spacing:.08em;text-transform:uppercase;text-decoration:none"
						>Ver la selección</a
					>
				{:else}
					{#if tienda.envios.gratisDesde}
						<div
							style="margin-top:16px;padding:14px 16px;background:#24121A;border:1px solid rgba(232,184,106,.3);display:flex;align-items:center;gap:12px;flex-wrap:wrap;font-size:14px"
						>
							<Icon icon={I.phPackage} size={22} style="color:var(--acento, #E8B86A)" /><span
								style="flex:1;min-width:200px"
								>{#if falta}Sumá <b>{fmt(falta)}</b> más y tenés
									<b style="color:var(--acento, #E8B86A)">envío gratis</b>{:else}<b
										style="color:var(--acento, #E8B86A)">Tenés envío gratis</b
									>{/if}</span
							><span style="display:flex;gap:4px"
								>{#each casillas as s, i (i)}<span
										style="width:14px;height:30px;border:1px solid var(--acento, #E8B86A);background:{s}"
									></span>{/each}</span
							>
						</div>
					{/if}
					<div
						style="margin-top:22px;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,360px),1fr));gap:clamp(22px,4cqw,56px);align-items:start"
					>
						<div style="min-width:0">
							{#each lineas as c (c.clave)}<div
									style="display:flex;gap:16px;padding:16px 0;border-bottom:1px solid rgba(244,233,225,.1)"
								>
									{#if c.producto.foto}<img
											src={img(c.producto.foto, 160, 224)}
											alt=""
											style="width:80px;height:112px;object-fit:cover;flex:none;background:#3A2026"
										/>{/if}
									<div style="flex:1;min-width:0;display:flex;flex-direction:column">
										<div style="display:flex;justify-content:space-between;gap:10px">
											<a
												href="{base}/producto/{c.producto.slug ?? c.id}"
												style="font-family:'Cormorant Garamond',serif;font-size:21px;font-weight:600;line-height:1.05;color:#F4E9E1;text-decoration:none"
												>{c.producto.nombre}</a
											><b style="white-space:nowrap">{fmt(c.subtotal)}</b>
										</div>
										<span style="font-size:12.5px;color:#BFA9A3;margin-top:4px"
											>{c.variante ? `${c.variante} · ` : ''}{fmt(c.producto.precio)} c/u</span
										>
										<span style="flex:1;min-height:12px"></span>
										<div style="display:flex;justify-content:space-between;align-items:center">
											<span
												style="display:flex;align-items:center;border:1px solid rgba(244,233,225,.2)"
												><button
													type="button"
													aria-label="Uno menos"
													onclick={() => carrito.fijar(c.clave, c.cantidad - 1)}
													style="width:34px;height:32px;display:grid;place-items:center;border:0;background:transparent;cursor:pointer;color:#F4E9E1"
													><Icon icon={I.phMinus} size={12} /></button
												><b style="width:24px;text-align:center;font-size:14px">{c.cantidad}</b
												><button
													type="button"
													aria-label="Uno más"
													disabled={c.cantidad >= c.producto.stock}
													onclick={() => carrito.fijar(c.clave, c.cantidad + 1)}
													style="width:34px;height:32px;display:grid;place-items:center;border:0;background:transparent;cursor:pointer;color:#F4E9E1"
													><Icon icon={I.phPlus} size={12} /></button
												></span
											><button
												type="button"
												onclick={() => carrito.fijar(c.clave, 0)}
												style="border:0;background:transparent;padding:0;font:400 12.5px 'Instrument Sans';color:#BFA9A3;text-decoration:underline;cursor:pointer"
												>Quitar</button
											>
										</div>
									</div>
								</div>{/each}
							{#if upsell.length}
								<div
									style="margin-top:24px;font-size:12px;font-weight:700;letter-spacing:.16em;text-transform:uppercase;color:var(--acento, #E8B86A)"
								>
									Para sumar
								</div>
								<div
									style="margin-top:12px;display:flex;gap:10px;overflow-x:auto;scrollbar-width:none"
								>
									{#each upsell as p (p.id)}<div style="flex:none;width:150px">
											<a
												href="{base}/producto/{p.slug ?? p.id}"
												style="display:block;aspect-ratio:3/4;overflow:hidden;background:#3A2026"
												>{#if p.foto}<img
														src={img(p.foto, 300, 400)}
														alt=""
														style="width:100%;height:100%;object-fit:cover"
													/>{/if}</a
											>
											<div
												style="margin-top:8px;font-family:'Cormorant Garamond',serif;font-size:17px;font-weight:600;line-height:1.1"
											>
												{p.nombre}
											</div>
											<div
												style="display:flex;justify-content:space-between;align-items:center;margin-top:4px"
											>
												<b style="font-size:14px">{tarjeta(p).price}</b><button
													type="button"
													aria-label="Sumar {p.nombre}"
													onclick={() => sumar(p)}
													style="width:28px;height:28px;border:1px solid var(--acento, #E8B86A);background:transparent;color:var(--acento, #E8B86A);display:grid;place-items:center;cursor:pointer"
													><Icon icon={I.phPlus} size={13} /></button
												>
											</div>
										</div>{/each}
								</div>
							{/if}
						</div>
						<div
							style="position:sticky;top:96px;padding:24px;background:#F4E9E1;color:#1E0E12;display:flex;flex-direction:column;gap:11px;font-size:14px"
						>
							<div style="font-family:'Cormorant Garamond',serif;font-size:32px;font-weight:600">
								Resumen
							</div>
							<span style="display:flex;justify-content:space-between"
								><span>{carrito.unidades} {carrito.unidades === 1 ? 'producto' : 'productos'}</span
								><span>{fmt(subtotal)}</span></span
							>
							<span style="display:flex;justify-content:space-between;gap:10px;color:#5A3A40"
								><span>Envío</span><span style="text-align:right"
									>Se calcula en el próximo paso</span
								></span
							>
							<div style="height:1px;background:rgba(30,14,18,.15)"></div>
							<span style="display:flex;justify-content:space-between;align-items:baseline"
								><b>Total</b><b style="font-family:'Cormorant Garamond',serif;font-size:34px"
									>{fmt(subtotal)}</b
								></span
							>
							{#if tienda.cobros.cuotas > 1}<span style="color:#5A3A40"
									>{tienda.cobros.cuotas} cuotas sin interés de {fmt(
										subtotal / tienda.cobros.cuotas
									)}</span
								>{/if}
							{#if tienda.cobros.transferencia > 0}
								<span style="color:#2F7D4F;font-weight:600"
									>{fmt(subtotal * (1 - tienda.cobros.transferencia / 100))} por transferencia</span
								>
							{/if}
							<a
								href="{base}/checkout"
								style="margin-top:4px;padding:16px;border:0;background:#1E0E12;color:#F4E9E1;font:700 14px 'Instrument Sans';letter-spacing:.08em;text-transform:uppercase;text-align:center;text-decoration:none"
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
				<section style="padding:32px clamp(16px,4cqw,48px) clamp(56px,8cqw,100px)">
					<p style="font-size:15px;color:#BFA9A3">Tu carrito está vacío.</p>
					<a href="{base}/catalogo" style="color:var(--acento, #E8B86A)">Ver la selección</a>
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
					style="padding:32px clamp(16px,4cqw,48px) clamp(56px,8cqw,100px);display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,360px),1fr));gap:clamp(22px,4cqw,56px);align-items:start"
				>
					<input type="hidden" name="carrito" value={JSON.stringify(carrito.lineas)} />
					<input type="hidden" name="envio" value={envio ?? ''} />
					<input type="hidden" name="pago" value={pago} />
					<div style="display:flex;flex-direction:column;gap:26px;min-width:0">
						<div
							style="display:flex;gap:8px;font-size:12px;letter-spacing:.14em;text-transform:uppercase"
						>
							<span style="color:var(--acento, #E8B86A);font-weight:700">1 Datos</span><span
								style="color:#8F7A75">— 2 Entrega — 3 Pago</span
							>
						</div>
						<div>
							<div style="font-family:'Cormorant Garamond',serif;font-size:34px;font-weight:600">
								Tus datos
							</div>
							<div
								style="margin-top:12px;display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));gap:8px"
							>
								{#each [['email', 'Email', 'email', 'email'], ['nombre', 'Nombre y apellido', 'text', 'name'], ['telefono', 'WhatsApp', 'tel', 'tel'], ['dni', 'DNI', 'text', 'off'], ...camposFactura(tienda)] as [n, l, tipo, ac] (n)}
									<label
										style="display:block;padding:11px 13px;border:1px solid rgba(244,233,225,.15);background:#24121A"
										><span style="display:block;font-size:11.5px;color:#BFA9A3">{l}</span><input
											name={n}
											type={tipo}
											autocomplete={ac as 'email'}
											required={!OPCIONALES.has(n)}
											maxlength="120"
											style="width:100%;margin-top:2px;padding:0;border:0;background:transparent;font:400 14.5px 'Instrument Sans';color:#F4E9E1;outline:none"
										/></label
									>
								{/each}
							</div>
							<!-- Sin bloque de newsletter en el diseño de esta plantilla: la casilla del checkout (§50). -->
							<label
								style="margin-top:10px;display:flex;align-items:center;gap:10px;min-height:44px;font:400 14.5px 'Instrument Sans';color:#F4E9E1;cursor:pointer"
								><input
									type="checkbox"
									name="novedades"
									style="flex:none;width:18px;height:18px;margin:0;accent-color:var(--acento, #E8B86A)"
								/>Quiero recibir novedades</label
							>
						</div>
						<div>
							<div style="font-family:'Cormorant Garamond',serif;font-size:34px;font-weight:600">
								Entrega
							</div>
							<div
								style="margin-top:12px;display:grid;grid-template-columns:minmax(0,140px) minmax(0,1fr);gap:8px"
							>
								<label
									style="display:block;padding:11px 13px;border:1px solid rgba(244,233,225,.15);background:#24121A"
									><span style="display:block;font-size:11.5px;color:#BFA9A3">Código postal</span
									><input
										name="cp"
										bind:value={cp}
										inputmode="numeric"
										maxlength="4"
										style="width:100%;margin-top:2px;padding:0;border:0;background:transparent;font:400 14.5px 'Instrument Sans';color:#F4E9E1;outline:none"
									/></label
								>
								<label
									style="display:block;padding:11px 13px;border:1px solid rgba(244,233,225,.15);background:#24121A"
									><span style="display:block;font-size:11.5px;color:#BFA9A3"
										>Dirección (calle, número, piso)</span
									><input
										name="direccion"
										maxlength="160"
										autocomplete="street-address"
										required={envio !== 'retiro'}
										style="width:100%;margin-top:2px;padding:0;border:0;background:transparent;font:400 14.5px 'Instrument Sans';color:#F4E9E1;outline:none"
									/></label
								>
							</div>
							<div
								style="margin-top:8px;display:flex;flex-direction:column;gap:8px"
								role="radiogroup"
								aria-label="Forma de entrega"
							>
								{#each opciones as s (s.id)}{@const on = envio === s.id}{@const c =
										estilo(on)}<button
										type="button"
										role="radio"
										aria-checked={on}
										onclick={() => (envio = s.id)}
										style="display:flex;align-items:center;gap:12px;padding:15px;border:1px solid {c.bd};background:{c.bg};text-align:left;cursor:pointer;font-family:inherit;color:#F4E9E1"
										><span
											style="width:18px;height:18px;border-radius:50%;border:1px solid var(--acento, #E8B86A);display:grid;place-items:center;flex:none"
											>{#if on}<span
													style="width:9px;height:9px;border-radius:50%;background:var(--acento, #E8B86A)"
												></span>{/if}</span
										><span style="flex:1;min-width:0"
											><span style="display:block;font-size:14.5px;font-weight:600">{s.nombre}</span
											><span style="display:block;font-size:12.5px;color:#BFA9A3">{s.detalle}</span
											></span
										><b style="color:{s.precio ? '#F4E9E1' : '#9FD3A8'};white-space:nowrap"
											>{s.precio ? fmt(s.precio) : 'Gratis'}</b
										></button
									>{:else}<span style="font-size:13.5px;color:#BFA9A3"
										>Poné tu código postal para ver cómo te llega.</span
									>{/each}
							</div>
							<EnvioCorreo {tienda} {cp} {envio} {opciones} />
						</div>
						<div>
							<div style="font-family:'Cormorant Garamond',serif;font-size:34px;font-weight:600">
								Pago
							</div>
							<div
								style="margin-top:12px;display:flex;flex-direction:column;gap:8px"
								role="radiogroup"
								aria-label="Forma de pago"
							>
								{#each pagos as s (s.id)}{@const on = pago === s.id}{@const c = estilo(on)}<button
										type="button"
										role="radio"
										aria-checked={on}
										onclick={() => (pago = s.id)}
										style="display:flex;align-items:center;gap:12px;padding:15px;border:1px solid {c.bd};background:{c.bg};text-align:left;cursor:pointer;font-family:inherit;color:#F4E9E1"
										><span
											style="width:18px;height:18px;border-radius:50%;border:1px solid var(--acento, #E8B86A);display:grid;place-items:center;flex:none"
											>{#if on}<span
													style="width:9px;height:9px;border-radius:50%;background:var(--acento, #E8B86A)"
												></span>{/if}</span
										><Icon icon={s.i} size={22} style="color:{s.c}" /><span
											style="flex:1;min-width:0"
											><span style="display:block;font-size:14.5px;font-weight:600">{s.t}</span
											><span style="display:block;font-size:12.5px;color:#BFA9A3">{s.d}</span></span
										></button
									>{/each}
							</div>
						</div>
						{#if error}<div
								role="alert"
								style="padding:12px 14px;border:1px solid #9E2F40;background:rgba(158,47,64,.18);color:#F4E9E1;font-size:14px"
							>
								{error}
							</div>{/if}
						<button
							disabled={!envio || !pagos.length || enviando || tienda.vitrina}
							style="padding:18px;border:0;background:var(--acento, #E8B86A);color:#1E0E12;font:700 15px 'Instrument Sans';letter-spacing:.08em;text-transform:uppercase;cursor:pointer;display:flex;justify-content:center;align-items:center;gap:8px"
							><Icon icon={I.phLockSimple} size={16} />{tienda.vitrina
								? 'Es una tienda de ejemplo'
								: !pagos.length
									? 'Esta tienda todavía no tiene medios de pago'
									: envio
										? `Pagar ${fmt(total)}`
										: 'Elegí cómo te llega'}</button
						>
					</div>
					<div
						style="position:sticky;top:96px;padding:24px;background:#24121A;border:1px solid rgba(244,233,225,.1);min-width:0"
					>
						<div style="font-family:'Cormorant Garamond',serif;font-size:30px;font-weight:600">
							Tu pedido
						</div>
						<div style="margin-top:14px;display:flex;flex-direction:column;gap:12px">
							{#each lineas as c (c.clave)}<div style="display:flex;gap:12px;align-items:center">
									<span style="position:relative;flex:none"
										>{#if c.producto.foto}<img
												src={img(c.producto.foto, 96, 132)}
												alt=""
												style="width:48px;height:66px;object-fit:cover;background:#3A2026;display:block"
											/>{:else}<span style="display:block;width:48px;height:66px;background:#3A2026"
											></span>{/if}<span
											style="position:absolute;top:-6px;right:-6px;width:20px;height:20px;border-radius:50%;background:var(--acento, #E8B86A);color:#1E0E12;font-size:11px;font-weight:700;display:grid;place-items:center"
											>{c.cantidad}</span
										></span
									><span
										style="flex:1;min-width:0;font-family:'Cormorant Garamond',serif;font-size:17px;font-weight:600;line-height:1.1"
										>{c.producto.nombre}{#if c.variante}<span
												style="display:block;margin-top:2px;font-family:'Instrument Sans',sans-serif;font-size:12.5px;font-weight:400;color:#BFA9A3"
												>{c.variante}</span
											>{/if}</span
									><b style="font-size:14px;white-space:nowrap">{fmt(c.subtotal)}</b>
								</div>{/each}
						</div>
						<div
							style="margin-top:16px;padding-top:14px;border-top:1px solid rgba(244,233,225,.1);display:flex;flex-direction:column;gap:8px;font-size:14px"
						>
							<span style="display:flex;justify-content:space-between"
								><span>Subtotal</span><span>{fmt(subtotal)}</span></span
							><span style="display:flex;justify-content:space-between"
								><span>Envío</span><span
									>{elegida ? (cuenta.envio ? fmt(cuenta.envio) : 'Gratis') : '—'}</span
								></span
							>{#if cuenta.descCupon}<span
									style="display:flex;justify-content:space-between;color:#9FD3A8;font-weight:600"
									><span>Cupón {cupon.aplicado}</span><span>− {fmt(cuenta.descCupon)}</span></span
								>{/if}{#if descuento}<span
									style="display:flex;justify-content:space-between;color:#9FD3A8;font-weight:600"
									><span>Transferencia −{tienda.cobros.transferencia}%</span><span
										>− {fmt(descuento)}</span
									></span
								>{/if}
							<div style="margin:6px 0 2px">
								<CampoCupon
									{cupon}
									lineas={carrito.lineas}
									tinta="#F4E9E1"
									suave="rgba(244,233,225,.55)"
									borde="rgba(244,233,225,.15)"
									fondo="#1E0E12"
									acento="var(--acento, #E8B86A)"
									sobreAcento="#1E0E12"
									radio="0px"
									fuente="'Instrument Sans',sans-serif"
									error="#FFB4AB"
								/>
							</div>
							<span
								style="display:flex;justify-content:space-between;align-items:baseline;margin-top:6px"
								><b>Total</b><b style="font-family:'Cormorant Garamond',serif;font-size:32px"
									>{fmt(total)}</b
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
				style="padding:clamp(56px,8cqw,110px) clamp(16px,4cqw,48px);max-width:720px;margin:0 auto;text-align:center"
			>
				<span
					style="display:inline-grid;place-items:center;width:64px;height:64px;border-radius:50%;border:1px solid var(--acento, #E8B86A);color:var(--acento, #E8B86A)"
					><Icon icon={I.phCheckBold} size={28} /></span
				>
				<h1
					style="margin:20px 0 0;font-family:'Cormorant Garamond',serif;font-weight:500;font-size:clamp(46px,6.4cqw,86px);line-height:.95;overflow-wrap:anywhere"
				>
					¡Gracias, <i style="color:var(--acento, #E8B86A)">{pedido.nombre}!</i>
				</h1>
				<p style="margin:16px 0 0;font-size:16px;line-height:1.55;color:#DCCBC4">
					Recibimos tu pedido <b>#{pedido.numero}</b> por <b>{fmt(pedido.total)}</b>. Te escribimos
					por mail con cada&nbsp;novedad.
				</p>
				<SenaPedido {tema} />
				{#if pedido.pago === 'transferencia'}
					<div
						style="margin-top:24px;padding:20px;background:#24121A;border:1px solid rgba(232,184,106,.3);text-align:left;font-size:14.5px;line-height:1.6"
					>
						<b>Transferí {fmt(pedido.total)}</b>
						{#if pedido.alias}al {pedido.esCbu ? 'CBU/CVU' : 'alias'}
							<b style="color:var(--acento, #E8B86A)">{pedido.alias}</b>{:else}con los datos que te
							mandamos por mail{/if} y mandanos el comprobante{#if whatsapp}&nbsp;<a
								href={wa(`¡Hola! Te mando el comprobante del pedido #${pedido.numero}.`)}
								target="_blank"
								rel="noopener"
								style="color:var(--acento, #E8B86A)">por WhatsApp</a
							>{:else}&nbsp;respondiendo el mail del pedido{/if}. Lo preparamos apenas
						se&nbsp;acredite.
					</div>
				{/if}
				<a
					href={base || '/'}
					style="display:inline-block;margin-top:28px;padding:16px 28px;background:var(--acento, #E8B86A);color:#1E0E12;font:700 14px 'Instrument Sans';letter-spacing:.08em;text-transform:uppercase;text-decoration:none"
					>Volver a la tienda</a
				>
				{#if pedido.estado === 'Entregado'}
					<div style="margin-top:clamp(40px,6cqw,64px)">
						<FormResena slug={tienda.slug} {tema} codigo={pedido.codigo} nombre={pedido.nombre} />
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
				style="display:grid;padding:32px clamp(16px,4cqw,48px) clamp(56px,8cqw,100px);grid-template-columns:repeat(auto-fit,minmax(min(100%,360px),1fr));gap:clamp(22px,4cqw,56px);align-items:start"
			>
				<input type="hidden" name="servicioId" value={reserva.servicioId ?? ''} />
				<input type="hidden" name="profesionalId" value={reserva.profesionalId ?? ''} />
				<input type="hidden" name="inicio" value={reserva.inicio ?? ''} />
				<div style="display:flex;flex-direction:column;gap:26px;min-width:0">
					<h1
						style="margin:0;font-family:'Cormorant Garamond',serif;font-weight:500;font-size:clamp(46px,6.4cqw,86px);line-height:.95;overflow-wrap:anywhere"
					>
						Reservá tu turno
					</h1>
					<div>
						<div style="font-family:'Cormorant Garamond',serif;font-size:34px;font-weight:600">
							1 · Servicio
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
										style="display:flex;align-items:center;gap:12px;padding:15px;text-align:left;border:1px solid {on
											? 'var(--acento, #E8B86A)'
											: 'rgba(244,233,225,.14)'};background:{on
											? 'rgba(232,184,106,.08)'
											: 'transparent'};cursor:pointer;font-family:inherit;color:#F4E9E1"
										><span
											style="width:18px;height:18px;border-radius:50%;border:1px solid var(--acento, #E8B86A);display:grid;place-items:center;flex:none"
											>{#if on}<span
													style="width:9px;height:9px;border-radius:50%;background:var(--acento, #E8B86A)"
												></span>{/if}</span
										><span style="flex:1;min-width:0"
											><span style="display:block;font-size:14.5px;font-weight:600">{s.nombre}</span
											><span style="display:block;font-size:12.5px;color:#BFA9A3"
												>{duracion(s.duracion)}</span
											></span
										><b style="font-size:14.5px;white-space:nowrap">{fmt(s.precio)}</b></button
									>{/each}
							</div>
						</div>
					</div>
					<div>
						<div style="font-family:'Cormorant Garamond',serif;font-size:34px;font-weight:600">
							2 · Profesional
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
										style="flex:none;width:112px;padding:10px;text-align:center;border:1px solid {on
											? 'var(--acento, #E8B86A)'
											: 'rgba(244,233,225,.14)'};background:{on
											? 'rgba(232,184,106,.08)'
											: 'transparent'};cursor:pointer;font-family:inherit;color:#F4E9E1"
										>{#if p?.foto}<img
												src={img(p.foto, 112, 112)}
												alt=""
												style="width:52px;height:52px;border-radius:50%;object-fit:cover;display:block;margin:0 auto"
											/>{:else}<span
												style="width:52px;height:52px;border-radius:50%;background:rgba(232,184,106,.08);color:var(--acento, #E8B86A);display:grid;place-items:center;margin:0 auto"
												><Icon icon={p ? I.phUserFill : I.phShuffleBold} size={22} /></span
											>{/if}<span
											style="display:block;margin-top:8px;font-size:13.5px;font-weight:700;overflow:hidden;text-overflow:ellipsis;white-space:nowrap"
											>{p?.nombre ?? 'Cualquiera'}</span
										><span
											style="display:block;font-size:11.5px;color:#BFA9A3;overflow:hidden;text-overflow:ellipsis;white-space:nowrap"
											>{p ? p.rol.split(' · ')[0] : 'El primero libre'}</span
										></button
									>{/each}
							</div>
						</div>
					</div>
					<div>
						<div style="font-family:'Cormorant Garamond',serif;font-size:34px;font-weight:600">
							3 · Día y horario
						</div>
						<div style="margin-top:12px">
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
											: 0.35};border:1px solid {sel
											? 'var(--acento, #E8B86A)'
											: 'rgba(244,233,225,.14)'};background:{sel
											? 'rgba(232,184,106,.08)'
											: 'transparent'};cursor:pointer;font-family:inherit;color:#F4E9E1"
										><span
											style="display:block;font-size:11px;font-weight:600;text-transform:uppercase"
											>{d.w}</span
										><span style="display:block;font-size:17px;font-weight:700;margin-top:1px"
											>{d.n}</span
										></button
									>{/each}
							</div>
							<div style="margin-top:16px;display:flex;flex-direction:column;gap:12px">
								{#if !reserva.servicioId}<p style="margin:0;font-size:14px;color:#BFA9A3">
										Elegí un servicio para ver los horarios.
									</p>{:else if reserva.horarios && !reserva.grupos.length}<p
										style="margin:0;font-size:14px;color:#BFA9A3"
									>
										No quedan horarios este día. Probá con otro día u otro&nbsp;profesional.
									</p>{/if}
								{#each reserva.grupos as g (g.k)}<div>
										<div
											style="font-size:12.5px;font-weight:700;color:#BFA9A3;margin-bottom:7px;display:flex;align-items:center;gap:6px"
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
														: 'line-through'};opacity:{h.libre ? 1 : 0.4};border:1px solid {on
														? 'var(--acento, #E8B86A)'
														: 'rgba(244,233,225,.14)'};background:{on
														? 'rgba(232,184,106,.08)'
														: 'transparent'};cursor:pointer;font-family:inherit;color:#F4E9E1"
													>{h.hora}</button
												>{/each}
										</div>
									</div>{/each}
							</div>
						</div>
					</div>
					{#if reserva.listo}<div>
							<div style="font-family:'Cormorant Garamond',serif;font-size:34px;font-weight:600">
								4 · Tus datos
							</div>
							<div style="margin-top:12px">
								<div
									style="display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));gap:8px"
								>
									{#each [['nombre', 'Nombre', 'text', 'name'], ['telefono', 'WhatsApp', 'tel', 'tel'], ['email', 'Email', 'email', 'email']] as [n, l, tipo, ac] (n)}
										<label
											style="display:block;padding:11px 13px;border:1px solid rgba(244,233,225,.15);background:#24121A"
											><span style="display:block;font-size:11.5px;color:#BFA9A3">{l}</span><input
												name={n}
												type={tipo}
												autocomplete={ac as 'email'}
												required
												maxlength="120"
												style="width:100%;margin-top:2px;padding:0;border:0;background:transparent;font:400 14.5px 'Instrument Sans';color:#F4E9E1;outline:none"
											/></label
										>
									{/each}
								</div>
							</div>
						</div>
						<div>
							<div style="font-family:'Cormorant Garamond',serif;font-size:34px;font-weight:600">
								Pago
							</div>
							<div style="margin-top:12px">
								<div
									style="display:flex;align-items:center;gap:12px;padding:15px;text-align:left;border:1px solid var(--acento, #E8B86A);background:rgba(232,184,106,.08);cursor:pointer;font-family:inherit;color:#F4E9E1"
								>
									<span
										style="width:18px;height:18px;border-radius:50%;border:1px solid var(--acento, #E8B86A);display:grid;place-items:center;flex:none"
										><span
											style="width:9px;height:9px;border-radius:50%;background:var(--acento, #E8B86A)"
										></span></span
									>{#if reserva.sena}<Icon
											icon={I.siMercadopago}
											size={22}
											style="color:#00B1EA;flex:none"
										/><span style="flex:1;min-width:0"
											><span style="display:block;font-size:14.5px;font-weight:600"
												>Mercado Pago</span
											><span style="display:block;font-size:12.5px;color:#BFA9A3"
												>Seña de {fmt(reserva.sena)} para confirmar · el resto, en el local</span
											></span
										>{:else}<Icon
											icon={I.phStorefrontFill}
											size={22}
											style="color:var(--acento, #E8B86A);flex:none"
										/><span style="flex:1;min-width:0"
											><span style="display:block;font-size:14.5px;font-weight:600"
												>Pagás en el local</span
											><span style="display:block;font-size:12.5px;color:#BFA9A3"
												>No hace falta seña: el turno queda confirmado</span
											></span
										>{/if}
								</div>
							</div>
						</div>{/if}
				</div>
				<div
					style="position:sticky;top:96px;padding:24px;background:#24121A;border:1px solid rgba(244,233,225,.1);min-width:0"
				>
					<div style="font-family:'Cormorant Garamond',serif;font-size:30px;font-weight:600">
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
							<div style="font-size:13px;color:#BFA9A3">
								{reserva.cuando ?? 'Elegí el día y el horario'}
							</div>
						</div>
					</div>
					{#if reserva.servicio}<div
							style="margin-top:16px;padding-top:14px;border-top:1px solid rgba(244,233,225,.1);display:flex;flex-direction:column;gap:8px;font-size:14px"
						>
							<span style="display:flex;justify-content:space-between;gap:10px"
								><span>{reserva.servicio.nombre}</span><span style="white-space:nowrap"
									>{fmt(reserva.servicio.precio)}</span
								></span
							><span style="display:flex;justify-content:space-between;color:#BFA9A3"
								><span>Duración total</span><span>{reserva.duracion}</span></span
							><span
								style="display:flex;justify-content:space-between;align-items:baseline;margin-top:6px"
								><b>Total</b><b style="font-family:'Cormorant Garamond',serif;font-size:32px"
									>{fmt(reserva.servicio.precio)}</b
								></span
							>
						</div>{/if}
					<div
						style="margin-top:14px;display:flex;gap:8px;font-size:13px;line-height:1.45;color:#BFA9A3"
					>
						<Icon icon={I.phShieldCheckFill} size={16} style="flex:none;margin-top:1px" /><span
							>{#if reserva.sena}Seña de {fmt(reserva.sena)} para confirmar. Podés cancelar desde el link
								del turno hasta {anticipacion} antes; la seña te la devuelve la&nbsp;tienda.{:else}No
								pagás nada ahora. Podés cancelar desde el link del turno hasta {anticipacion}&nbsp;antes.{/if}</span
						>
					</div>
					{#if error}<div
							role="alert"
							style="margin-top:12px;padding:12px 14px;border:1px solid #9E2F40;background:rgba(158,47,64,.18);color:#F4E9E1;font-size:14px"
						>
							{error}
						</div>{/if}
					<button
						disabled={!reserva.listo || reservando || tienda.vitrina}
						style="margin-top:16px;width:100%;padding:18px;border:0;background:var(--acento, #E8B86A);color:#1E0E12;font:700 15px 'Instrument Sans';letter-spacing:.08em;text-transform:uppercase;cursor:pointer"
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
				style="padding:clamp(56px,8cqw,110px) clamp(16px,4cqw,48px);max-width:720px;margin:0 auto;text-align:center"
			>
				<span
					style="display:inline-grid;place-items:center;width:64px;height:64px;border-radius:50%;border:1px solid var(--acento, #E8B86A);color:var(--acento, #E8B86A)"
					><Icon icon={cancelado ? I.phCalendarX : I.phCheckBold} size={28} /></span
				>
				<h1
					style="margin:20px 0 0;font-family:'Cormorant Garamond',serif;font-weight:500;font-size:clamp(46px,6.4cqw,86px);line-height:.95;overflow-wrap:anywhere"
				>
					{cancelado
						? 'Turno cancelado'
						: turno.estado === 'reservado'
							? `Falta la seña, ${turno.nombre}`
							: `¡Listo, ${turno.nombre}!`}
				</h1>
				<p style="margin:16px 0 0;font-size:16px;line-height:1.55;color:#DCCBC4">
					{#if cancelado}Liberamos el horario{turno.senaPagada
							? '. La seña te la devuelve la tienda por Mercado Pago'
							: ''}. Si querés, reservá&nbsp;otro.{:else if turno.estado === 'reservado'}Te
						guardamos el horario unos minutos mientras pagás la seña con Mercado&nbsp;Pago.{:else}Tu
						turno está confirmado. Te esperamos en&nbsp;{turno.tienda}.{/if}
				</p>
				{#if aviso === 'pendiente' || aviso === 'error'}<div
						role="status"
						style="margin-top:16px;padding:12px 14px;border:1px solid #9E2F40;background:rgba(158,47,64,.18);color:#F4E9E1;font-size:14px"
					>
						{aviso === 'pendiente'
							? 'Mercado Pago todavía no acreditó la seña: te avisamos cuando entre.'
							: 'El pago de la seña no salió. Podés reservar de nuevo.'}
					</div>{/if}
				<div
					style="margin-top:24px;padding:20px;background:#24121A;border:1px solid rgba(232,184,106,.3);text-align:left;font-size:14.5px;line-height:1.6"
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
								style="color:var(--acento, #E8B86A);flex:none"
							/>{turno.direccion}
						</div>{/if}
				</div>
				{#if error}<div
						role="alert"
						style="margin-top:16px;padding:12px 14px;border:1px solid #9E2F40;background:rgba(158,47,64,.18);color:#F4E9E1;font-size:14px"
					>
						{error}
					</div>{/if}
				<div style="margin-top:28px;display:flex;justify-content:center;gap:10px;flex-wrap:wrap">
					{#if !cancelado}<a
							href="{base}/turno/{turno.codigo}/turno.ics"
							download="turno.ics"
							data-sveltekit-reload
							style="display:flex;align-items:center;gap:8px;padding:16px 28px;background:var(--acento, #E8B86A);color:#1E0E12;font:700 14px 'Instrument Sans';letter-spacing:.08em;text-transform:uppercase;text-decoration:none;border:0"
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
									style="cursor:pointer;padding:16px 28px;background:var(--acento, #E8B86A);color:#1E0E12;font:700 14px 'Instrument Sans';letter-spacing:.08em;text-transform:uppercase;text-decoration:none;border:0"
									>Sí, cancelar el turno</button
								>{:else}<button
									type="button"
									onclick={() => (seguro = true)}
									style="cursor:pointer;padding:16px 24px;border:1px solid rgba(244,233,225,.45);background:transparent;color:#F4E9E1;font:600 14px 'Instrument Sans';letter-spacing:.08em;text-transform:uppercase;text-decoration:none"
									>Cancelar turno</button
								>{/if}
						</form>{:else if cancelado}<a
							href="{base}/turnos"
							style="padding:16px 28px;background:var(--acento, #E8B86A);color:#1E0E12;font:700 14px 'Instrument Sans';letter-spacing:.08em;text-transform:uppercase;text-decoration:none;border:0"
							>Reservar otro</a
						>{/if}
					<a
						href={base || '/'}
						style="padding:16px 24px;border:1px solid rgba(244,233,225,.45);background:transparent;color:#F4E9E1;font:600 14px 'Instrument Sans';letter-spacing:.08em;text-transform:uppercase;text-decoration:none"
						>Volver a la tienda</a
					>
				</div>
				{#if !cancelado && !turno.cancelable}<p
						style="margin:16px 0 0;font-size:16px;line-height:1.55;color:#DCCBC4"
					>
						Para cambiarlo o cancelarlo, escribile a la&nbsp;tienda{#if whatsapp}&nbsp;<a
								href={wa(`¡Hola! Tengo un turno el ${cuandoLargo(turno.inicio)}.`)}
								target="_blank"
								rel="noopener"
								style="color:var(--acento, #E8B86A)">por WhatsApp</a
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
			<section style="padding:clamp(56px,8cqw,110px) clamp(16px,4cqw,48px) 0;text-align:center">
				<div
					style="font-size:12.5px;letter-spacing:.3em;text-transform:uppercase;color:var(--acento, #E8B86A)"
				>
					{tienda.nombre}{tienda.ciudad ? ` · ${tienda.ciudad}` : ''}
				</div>
				<h1
					style="margin:18px auto 0;max-width:900px;font-family:'Cormorant Garamond',serif;font-weight:500;font-size:clamp(46px,7.4cqw,108px);line-height:.92;overflow-wrap:anywhere"
				>
					{titulo[0]}{titulo[1] ? ' ' : ''}{#if titulo[1]}<i style="color:var(--acento, #E8B86A)"
							>{titulo[1]}</i
						>{/if}
				</h1>
				{#if tienda.portada.lema}<p
						style="margin:18px auto 0;max-width:520px;font-size:16px;line-height:1.55;color:#DCCBC4"
					>
						{tienda.portada.lema}
					</p>{/if}
			</section>
			{#if tienda.fotos.length >= 3}
				<section
					style="padding:clamp(40px,6cqw,80px) clamp(16px,4cqw,48px) 0;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,280px),1fr));gap:14px"
				>
					{#each tienda.fotos.slice(0, 3) as g (g)}<div
							style="aspect-ratio:4/5;overflow:hidden;background:#3A2026"
						>
							<img src={img(g, 600, 750)} alt="" style="width:100%;height:100%;object-fit:cover" />
						</div>{/each}
				</section>
			{:else if tienda.portada.foto}
				<section style="padding:clamp(40px,6cqw,80px) clamp(16px,4cqw,48px) 0">
					<div style="aspect-ratio:16/7;overflow:hidden;background:#3A2026">
						<img
							src={img(tienda.portada.foto, 1600)}
							alt=""
							style="width:100%;height:100%;object-fit:cover"
						/>
					</div>
				</section>
			{/if}
			<section
				style="margin-top:clamp(56px,8cqw,100px);padding:clamp(40px,6cqw,80px) clamp(16px,4cqw,48px);background:#F4E9E1;color:#1E0E12;display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,320px),1fr));gap:clamp(24px,4cqw,56px)"
			>
				<div>
					<h2
						style="margin:0;font-family:'Cormorant Garamond',serif;font-weight:500;font-size:clamp(40px,5cqw,64px);line-height:1"
					>
						{tienda.envios.retiro ? 'Vení al local' : 'Hablemos'}
					</h2>
					<p style="margin:12px 0 0;font-size:15px;line-height:1.55;color:#5A3A40;max-width:420px">
						{comoLlega(tienda.envios)}
					</p>
					<div style="margin-top:20px;display:flex;flex-direction:column;gap:12px;font-size:15px">
						{#if tienda.direccion}<span style="display:flex;gap:12px"
								><Icon icon={I.phMapPin} size={19} style="flex:none" />{tienda.direccion}</span
							>{/if}{#if tienda.whatsapp}<span style="display:flex;gap:12px"
								><Icon icon={I.phWhatsappLogo} size={19} style="flex:none" />{tienda.whatsapp}</span
							>{/if}
					</div>
				</div>
				{#if whatsapp}
					<div style="display:flex;flex-direction:column;gap:10px;justify-content:center">
						<div style="font-family:'Cormorant Garamond',serif;font-size:30px;font-weight:600">
							Escribinos
						</div>
						<a
							href={wa('¡Hola!')}
							target="_blank"
							rel="noopener"
							style="display:flex;justify-content:center;align-items:center;gap:8px;padding:15px;background:#1E0E12;color:#F4E9E1;font-size:14px;font-weight:700;letter-spacing:.06em;text-transform:uppercase;text-decoration:none"
							><Icon icon={I.phWhatsappLogo} size={18} />Por WhatsApp</a
						>
					</div>
				{/if}
			</section>
		{/if}

		<!-- footer -->
		<footer
			style="padding:clamp(48px,6cqw,80px) clamp(16px,4cqw,48px) 26px;border-top:1px solid rgba(244,233,225,.1)"
		>
			<div
				style="display:grid;grid-template-columns:repeat(auto-fit,minmax(min(100%,190px),1fr));gap:30px"
			>
				<div style="min-width:0;max-width:380px">
					<div style="font-family:'Cormorant Garamond',serif;font-size:36px;font-weight:600">
						{tienda.nombre}
					</div>
					<p style="margin:10px 0 0;font-size:14px;color:#BFA9A3;line-height:1.55">
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
							style="font-size:11.5px;letter-spacing:.16em;text-transform:uppercase;color:var(--acento, #E8B86A);margin-bottom:4px"
							>{f.t}</b
						>{#each f.l as l (l.t)}<a href={l.href} style="color:#DCCBC4;text-decoration:none"
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
				style="margin-top:36px;padding-top:16px;border-top:1px solid rgba(244,233,225,.1);display:flex;justify-content:space-between;gap:14px;flex-wrap:wrap;font-size:12px;color:#8F7A75"
			>
				<span>Beber con moderación. Prohibida su venta a menores de 18&nbsp;años.</span><span
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
				style="position:fixed;right:18px;bottom:18px;z-index:30;width:54px;height:54px;border-radius:50%;background:#25D366;color:#fff;display:grid;place-items:center;box-shadow:0 12px 26px -8px rgba(0,0,0,.5)"
				><Icon icon={I.phWhatsappLogoFill} size={28} /></a
			>{/if}
	</div>
</div>

<!-- eslint-enable svelte/no-navigation-without-resolve -->

<style>
	.malbec :global(a:hover) {
		color: var(--acento, #e8b86a);
	}
	.caja:hover {
		border-color: rgba(232, 184, 106, 0.5) !important;
	}
	.caja :global(img),
	.zoom :global(img) {
		transition: transform 0.8s cubic-bezier(0.2, 0.7, 0.1, 1);
	}
	.caja:hover :global(img) {
		transform: scale(1.04);
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
</style>
