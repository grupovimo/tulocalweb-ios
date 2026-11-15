<script lang="ts">
	import { browser } from '$app/env';
	import { page } from '$app/state';
	import { untrack } from 'svelte';
	import { usarTienda } from '../contexto.ts';
	import { disenoDeTienda } from '../migrar.ts';
	import { validarVentanas, type Ventana as TVentana } from '../promos.ts';
	import { anotarVista, contarVisita, elegir, sesionDe, varianteEn } from '../ventanas.ts';
	import { vistaVentanas } from '../vista-ventanas.svelte.ts';
	import Ventana from './Ventana.svelte';

	/**
	 * Las ventanas de la tienda (diseno.ventanas): en cada página elige la primera que califica
	 * (ventanas.ts: activa y completa, fechas, páginas, dispositivo, público y frecuencia; nunca en
	 * el checkout ni en los legales), arma su disparador y la abre una sola vez por página:
	 * - `entrar`: a los 1,5 s; `segundos`: a los que eligió; `scroll`: al bajar ese porcentaje;
	 * - `salida`: en la compu, el mouse sale por arriba; en el celular, una subida rápida después
	 *   de haber bajado más del 40 % o volver a la pestaña después de 20 s afuera; nunca antes de 8 s;
	 * - `carrito`: al agregar el primer producto.
	 * En la vista previa del editor no sale sola: la abre el editor (vista-ventanas.svelte.ts).
	 *
	 * Uso: la pone marco/Tienda.svelte.
	 */
	const ctx = usarTienda();
	const ventanas = $derived(validarVentanas(disenoDeTienda(ctx.tienda).ventanas));
	let abierta = $state<TVentana | null>(null);
	let movil = $state(false);
	let abajo = $state(0);

	function abrir(w: TVentana) {
		if (abierta) return;
		movil = matchMedia('(max-width: 759px)').matches;
		// La barra fija de abajo (ficha, carrito, pedido) queda a la vista: la hoja sube encima.
		const barra = document.querySelector<HTMLElement>('.tv-reserva-barra');
		abajo = movil && barra && barra.offsetParent !== null ? barra.offsetHeight : 0;
		abierta = w;
		if (!vistaVentanas.vista) anotarVista(ctx.tienda.slug, w.id);
	}

	// La ventana pedida por el editor (vista previa): con lo último del borrador.
	$effect(() => {
		void vistaVentanas.vez;
		const id = vistaVentanas.forzada;
		if (!browser || !vistaVentanas.vista) return;
		abierta = null;
		const w = id ? untrack(() => ventanas.find((x) => x.id === id)) : null;
		if (w) queueMicrotask(() => abrir(w));
	});
	const mostrada = $derived(
		abierta && vistaVentanas.vista
			? (ventanas.find((x) => x.id === abierta!.id) ?? abierta)
			: abierta
	);

	// Las unidades del carrito, para el disparo `carrito` (de 0 a algo).
	const unidades = $derived(ctx.carrito.unidades);

	// Una vez por página: se rearma al cambiar de camino.
	$effect(() => {
		void page.url.pathname;
		if (!browser || vistaVentanas.vista) return;
		const slug = ctx.tienda.slug;
		const memoria = contarVisita(slug);
		const esMovil = matchMedia('(max-width: 759px)').matches;
		const w = elegir(ventanas, {
			pagina: ctx.pagina,
			movil: esMovil,
			ahora: Date.now(),
			memoria,
			sesion: sesionDe(slug)
		});
		if (!w) return;
		return armar(w, esMovil);
	});

	let carritoAntes = -1;
	let alCarrito: (() => void) | null = null;
	$effect(() => {
		const u = unidades;
		if (carritoAntes === 0 && u > 0) alCarrito?.();
		carritoAntes = u;
	});

	function armar(w: TVentana, esMovil: boolean): () => void {
		const limpiar: (() => void)[] = [];
		let hecho = false;
		const salir = () => {
			if (hecho) return;
			hecho = true;
			for (const f of limpiar) f();
			abrir(w);
		};
		const despues = (ms: number, f: () => void) => {
			const t = setTimeout(f, ms);
			limpiar.push(() => clearTimeout(t));
		};
		const escuchar = <K extends keyof DocumentEventMap>(
			el: Document | Window,
			ev: K | string,
			f: (e: Event) => void
		) => {
			el.addEventListener(ev, f, { passive: true });
			limpiar.push(() => el.removeEventListener(ev, f));
		};
		const bajado = () => {
			const alto = document.documentElement.scrollHeight - innerHeight;
			return alto > 0 ? scrollY / alto : 0;
		};
		const inicio = Date.now();
		switch (w.disparo.cuando) {
			case 'entrar':
				despues(1500, salir);
				break;
			case 'segundos':
				despues((w.disparo.valor ?? 10) * 1000, salir);
				break;
			case 'scroll': {
				const meta = (w.disparo.valor ?? 50) / 100;
				escuchar(window, 'scroll', () => bajado() >= meta && salir());
				break;
			}
			case 'salida': {
				const listo = () => Date.now() - inicio >= 8000;
				if (!esMovil) {
					escuchar(document, 'mouseout', (e) => {
						const m = e as MouseEvent;
						if (!m.relatedTarget && m.clientY <= 0 && listo()) salir();
					});
				} else {
					let maximo = 0;
					let antes = { y: scrollY, t: Date.now() };
					escuchar(window, 'scroll', () => {
						maximo = Math.max(maximo, bajado());
						const ahora = { y: scrollY, t: Date.now() };
						const sube = antes.y - ahora.y;
						const dt = Math.max(1, ahora.t - antes.t);
						// Una subida rápida (más de 1,5 px por ms) después de haber bajado el 40 %.
						if (maximo > 0.4 && sube > 60 && sube / dt > 1.5 && listo()) salir();
						if (dt > 120) antes = ahora;
					});
				}
				let fuera = 0;
				escuchar(document, 'visibilitychange', () => {
					if (document.visibilityState === 'hidden') fuera = Date.now();
					else if (fuera && Date.now() - fuera > 20_000 && listo()) salir();
				});
				break;
			}
			case 'carrito':
				alCarrito = salir;
				limpiar.push(() => (alCarrito = null));
				break;
		}
		return () => {
			hecho = true;
			for (const f of limpiar) f();
		};
	}
</script>

{#if mostrada}
	{#key `${mostrada.id}:${vistaVentanas.vez}`}
		<Ventana
			ventana={mostrada}
			variante={varianteEn(mostrada, movil)}
			medir={!vistaVentanas.vista}
			{abajo}
			alCerrar={() => {
				abierta = null;
				if (vistaVentanas.vista) vistaVentanas.forzada = null;
			}}
		/>
	{/key}
{/if}
