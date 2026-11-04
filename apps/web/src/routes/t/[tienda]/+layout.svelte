<script lang="ts">
	import { afterNavigate } from '$app/navigation';
	import AvisoCookies from '#lib/tienda/AvisoCookies.svelte';
	import { pagina, preparar } from '#lib/tienda/medicion.ts';
	import AnuncioTienda from '#lib/tienda/AnuncioTienda.svelte';

	let { data, children } = $props();

	// La medición de la tienda (GA4 y píxel de Meta, sólo con consentimiento): una página vista por
	// navegación. Las plantillas no saben nada de esto.
	afterNavigate(({ type, from, to }) => {
		preparar(data.tienda);
		pagina();
		// Visitas propias, sin cookies (decisiones §36): la carga completa ya la contó el servidor;
		// cada página nueva dentro de la tienda se avisa sin esperar respuesta.
		if (type === 'enter' || data.tienda.vitrina || !to?.route.id) return;
		if (from?.url.pathname === to.url.pathname) return;
		navigator.sendBeacon?.(
			`/t/${data.tienda.slug}/vista`,
			JSON.stringify({ r: to.route.id, p: to.params?.id ?? '' })
		);
	});

	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import { carritoDe } from '#lib/tienda/carrito.svelte.ts';
	import { faltantes, textoFaltantes, type Faltante } from '#lib/tienda/faltantes.ts';
	import { conectarVidriera } from '#lib/tiempo-real/publico.svelte.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	// Tiempo real (§44): si se agota algo que está en el carrito (o el producto que se está mirando),
	// se avisa antes de pagar, en vez de enterarse por el error del checkout.
	const ui = getUi();
	let avisado = '';
	function revisarCarrito() {
		const fs = faltantes(carritoDe(data.tienda.slug).lineas, data.tienda.productos);
		const clave = JSON.stringify(fs);
		if (!fs.length || clave === avisado) {
			if (!fs.length) avisado = '';
			return;
		}
		avisado = clave;
		const t = textoFaltantes(fs);
		ui.toast({
			k: 'warn',
			title: t.titulo,
			desc: t.detalle,
			action: 'Ajustar carrito',
			onAction: () => ajustar(fs)
		});
	}
	function ajustar(fs: Faltante[]) {
		const c = carritoDe(data.tienda.slug);
		for (const f of fs) c.fijar(f.clave, f.quedan);
		avisado = '';
	}
	/** El producto que se está mirando (por slug o id), para avisar si se agota mientras tanto. */
	const mirado = () =>
		page.params.id
			? data.tienda.productos.find((p) => p.slug === page.params.id || p.id === page.params.id)
			: undefined;
	onMount(() => {
		if (!data.tiempoReal) return;
		const stock = () => new Map(data.tienda.productos.map((p) => [p.id, p.stock]));
		let antes = stock();
		const c = conectarVidriera({
			slug: data.tienda.slug,
			origen: data.tiempoReal.origen,
			alRecargar: (e) => {
				if (!e.tipo.startsWith('producto.')) return;
				const p = mirado();
				const en = carritoDe(data.tienda.slug).lineas.some((l) => l.id === p?.id);
				if (p && (antes.get(p.id) ?? 0) > 0 && p.stock === 0 && !en)
					ui.toast({
						k: 'warn',
						title: `Se agotó ${p.nombre}`,
						desc: 'Mientras lo mirabas se vendió el último.'
					});
				antes = stock();
				revisarCarrito();
			}
		});
		return () => c?.detener();
	});
	// Al entrar al carrito o al checkout también se revisa (el carrito pudo quedar de otro día).
	afterNavigate(({ to }) => {
		if (/\/(carrito|checkout)$/.test(to?.url.pathname ?? '')) revisarCarrito();
	});

	// Carrito abandonado: en el checkout, cuando quien compra deja su email (y al cambiar el nombre),
	// se guarda el carrito para poder recordárselo. Escucha los campos `email` y `nombre` del form,
	// así ninguna plantilla tiene que hacer nada. Se manda sólo si cambió algo.
	// Mi tienda › Estilo: el ícono de la pestaña.
	const icono = $derived(data.tienda.estilo?.icono || '');

	let ultimo = '';
	function alSalir(e: FocusEvent) {
		const el = e.target;
		if (!(el instanceof HTMLInputElement) || !el.form || !['email', 'nombre'].includes(el.name))
			return;
		if (data.tienda.vitrina || !page.url.pathname.endsWith('/checkout')) return;
		const campo = (n: string) =>
			(el.form!.elements.namedItem(n) as HTMLInputElement | null)?.value.trim() ?? '';
		const email = campo('email');
		const carrito = carritoDe(data.tienda.slug).lineas;
		if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email) || email.length > 120 || !carrito.length) return;
		const cuerpo = JSON.stringify({ email, nombre: campo('nombre').slice(0, 80), carrito });
		if (cuerpo === ultimo) return;
		ultimo = cuerpo;
		fetch(`/t/${data.tienda.slug}/carrito/guardar`, {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: cuerpo,
			keepalive: true
		}).catch(() => {});
	}
</script>

<svelte:head>
	{#if icono}<link rel="icon" href={icono} />{/if}
	<link rel="preconnect" href="https://fonts.googleapis.com" />
	<link
		href="https://fonts.googleapis.com/css2?family=Instrument+Sans:wght@400;500;600;700&family=Instrument+Serif:ital@0;1&display=swap"
		rel="stylesheet"
	/>
</svelte:head>

<svelte:document onfocusout={alSalir} />

<AnuncioTienda tienda={data.tienda} />
{@render children()}
{#key data.tienda.slug}
	<AvisoCookies tienda={data.tienda} />
{/key}
