<script lang="ts">
	import { resolve } from '$app/paths';
	import { deslizable } from '#lib/panel/deslizable.ts';
	import { untrack } from 'svelte';
	import { SvelteMap, SvelteSet } from 'svelte/reactivity';
	import { goto, invalidateAll } from '$app/navigation';
	import { page } from '$app/state';
	import Icon from '#lib/components/Icon.svelte';
	import Interruptor from '#lib/components/Interruptor.svelte';
	import MarcaIcono from '#lib/components/MarcaIcono.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { cara, unsplash } from '#lib/landing/formato.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { compartirTienda } from '#lib/panel/compartir.ts';
	import { pesos } from '#lib/panel/formato.ts';
	import HojaPedido from '#lib/panel/pedidos/HojaPedido.svelte';
	import { accionRapida, avisoDe, whatsappDe } from '#lib/panel/pedidos/acciones.ts';
	import {
		escribirFiltros,
		leerFiltros,
		VISTAS,
		vistaDe,
		type EstadoPedido,
		type Filtros
	} from '#lib/panel/pedidos/filtros.ts';
	import {
		ETAPAS,
		esperaPago,
		etapaDePedido,
		ICONO_ENVIO,
		ICONO_PAGO,
		NOMBRE_CANAL,
		NOMBRE_ENVIO,
		NOMBRE_PAGO,
		proximoDe,
		textoEstado
	} from '#lib/panel/pedidos/nombres.ts';
	import { getUi } from '#lib/ui.svelte.ts';
	import { escucharTiempoReal } from '#lib/tiempo-real/panel.svelte.ts';
	import type { components } from '@mostrador/contrato/esquema';

	type Fila = components['schemas']['PedidoFila'];
	type Detalle = components['schemas']['PedidoDetalle'];

	/** Pedidos (design/PanelPedidos.dc.html): vistas, embudo, filtros en la URL, lista virtual y acciones. */
	let { data } = $props();
	const ui = getUi();
	const f = $derived(leerFiltros(page.url.searchParams));
	const lista = $derived(data.lista);
	const vista = $derived(vistaDe(f));

	// ---- Navegar con filtros (quedan en la URL) ----
	function ir(cambio: Partial<Filtros>) {
		const q = escribirFiltros({ ...f, ...cambio }).toString();
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- misma página, sólo cambia la query
		void goto(q ? `?${q}` : '?', { replace: true, reset: false });
		arriba();
	}
	const alternar = <T,>(xs: T[], x: T) => (xs.includes(x) ? xs.filter((y) => y !== x) : [...xs, x]);

	// Búsqueda con un respiro mientras se escribe.
	let q = $state(untrack(() => f.q));
	let espera: ReturnType<typeof setTimeout>;
	function buscar(v: string) {
		q = v;
		clearTimeout(espera);
		espera = setTimeout(() => ir({ q: v }), 250);
	}

	// ---- Lista virtual (handoff: filas de 76 px, ventanas que se piden al scrollear) ----
	const FILA = 76;
	const VENTANA = 60;
	let ancho = $state(1440);
	const amplio = $derived(ancho >= 900);
	const celu = $derived(ancho < 760);
	const alto = $derived(ancho < 760 ? 520 : 600);
	// Con pocos pedidos la caja no deja un hueco en blanco abajo; vacía, lo justo para el vacío.
	const altoCaja = $derived(lista.total ? Math.min(alto, lista.total * FILA + 10) : 380);
	let scroll = $state(0);
	let caja = $state<HTMLElement>();
	let filas = $state.raw<(Fila | undefined)[]>([]);
	const pedidas = new SvelteSet<number>();
	$effect(() => {
		// Con cada lista nueva (filtros o cambios) se arranca de cero con la primera ventana.
		const l = lista;
		untrack(() => {
			filas = Array.from({ length: l.total }, (_, i) => l.filas[i]);
			pedidas.clear();
			pedidas.add(0);
		});
	});
	const inicio = $derived(Math.max(0, Math.floor(scroll / FILA) - 4));
	const fin = $derived(Math.min(lista.total, Math.ceil((scroll + alto) / FILA) + 6));
	const visibles = $derived(
		Array.from({ length: Math.max(0, fin - inicio) }, (_, k) => inicio + k)
	);
	$effect(() => {
		for (let i = inicio; i < fin; i++)
			if (!filas[i]) void pedirVentana(Math.floor(i / VENTANA) * VENTANA);
	});
	async function pedirVentana(desde: number) {
		if (pedidas.has(desde)) return;
		pedidas.add(desde);
		const params = escribirFiltros(f);
		params.set('desde', String(desde));
		params.set('cuantos', String(VENTANA));
		const r = await fetch(`/panel/pedidos/filas?${params}`);
		if (!r.ok) return pedidas.delete(desde);
		const l = (await r.json()) as typeof lista;
		const copia = filas.slice();
		l.filas.forEach((x, k) => (copia[desde + k] = x));
		filas = copia;
	}
	function arriba() {
		caja?.scrollTo({ top: 0 });
		scroll = 0;
	}
	let pendiente = 0;
	function alScrollear(e: Event) {
		const t = (e.currentTarget as HTMLElement).scrollTop;
		if (pendiente) return;
		pendiente = requestAnimationFrame(() => {
			pendiente = 0;
			scroll = t;
		});
	}

	// ---- Selección y acciones ----
	const seleccion = new SvelteMap<number, number>();
	const sumaSel = $derived([...seleccion.values()].reduce((s, x) => s + x, 0));
	function tildar(p: Fila) {
		if (seleccion.has(p.numero)) seleccion.delete(p.numero);
		else seleccion.set(p.numero, p.total);
	}
	function tildarVisibles() {
		for (let i = inicio; i < fin; i++) {
			const p = filas[i];
			if (p) seleccion.set(p.numero, p.total);
		}
	}

	async function cambiar(numero: number, estado: EstadoPedido) {
		const r = await fetch(`/panel/pedidos/${numero}`, {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ estado })
		});
		if (!r.ok) {
			ui.toast({
				k: 'warn',
				title: 'No se pudo',
				desc: 'El pedido ya cambió de etapa. Actualizamos la lista.'
			});
			await invalidateAll();
			return false;
		}
		return true;
	}

	// ---- Acción de un toque y WhatsApp desde la fila (lib/panel/pedidos/acciones.ts) ----
	const contexto = $derived({ retiro: data.retiro, slug: page.data.tienda.slug as string });
	const enCurso = new SvelteSet<number>();
	async function accion(p: Fila) {
		const a = accionRapida(p);
		if (!a || enCurso.has(p.numero)) return;
		enCurso.add(p.numero);
		try {
			if (!(await cambiar(p.numero, a.a))) return;
			await invalidateAll();
		} finally {
			enCurso.delete(p.numero);
		}
		ui.toast({
			k: 'ok',
			...avisoDe(p),
			action: 'Deshacer',
			onAction: async () => {
				if (await cambiar(p.numero, p.estado)) await invalidateAll();
			}
		});
	}

	const LOTE = [
		{ accion: 'preparar', t: 'Preparar', i: I.phPackageFill },
		{ accion: 'etiquetas', t: 'Etiquetas', i: I.phPrinterFill },
		{ accion: 'tarjetas', t: 'Tarjetas', i: I.phEnvelopeSimpleFill },
		{ accion: 'facturar', t: 'Facturar', i: I.phReceiptFill },
		{ accion: 'avisar', t: 'Avisar', i: I.phWhatsappLogoFill }
	] as const;
	async function lote(l: (typeof LOTE)[number]) {
		const numeros = [...seleccion.keys()];
		// Las tarjetas A6 de los paquetes se imprimen en otra pestaña (4 por A4).
		// Las etiquetas se imprimen en otra pestaña y los pedidos quedan listos para despachar.
		if (l.accion === 'etiquetas')
			window.open(
				`${resolve('/panel/pedidos/imprimir')}?que=etiquetas&n=${numeros.slice(0, 40).join(',')}`,
				'_blank'
			);
		if (l.accion === 'tarjetas') {
			// La impresión toma hasta 40 pedidos por vez.
			if (numeros.length > 40)
				ui.toast({
					k: 'warn',
					title: 'Se imprimen las primeras 40 tarjetas',
					desc: 'Para el resto, seleccioná otra tanda'
				});
			return window.open(
				`${resolve('/panel/pedidos/imprimir')}?n=${numeros.slice(0, 40).join(',')}`,
				'_blank'
			);
		}
		const r = await fetch('/panel/pedidos/lote', {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ numeros, accion: l.accion })
		});
		if (!r.ok)
			return ui.toast({ k: 'err', title: 'No se pudo', desc: 'Probá de nuevo en un rato.' });
		const { hechos, cambios } = (await r.json()) as {
			hechos: number;
			cambios: { numero: number; antes: EstadoPedido }[];
		};
		seleccion.clear();
		await invalidateAll();
		const saltados = numeros.length - hechos;
		ui.toast({
			k: 'ok',
			title: `${l.t} · ${hechos} ${hechos === 1 ? 'pedido' : 'pedidos'}`,
			desc: saltados
				? `${saltados} no estaban en la etapa justa y quedaron como estaban.`
				: 'Listo. Te avisamos si algo falla.',
			...(hechos && (l.accion === 'preparar' || l.accion === 'etiquetas')
				? {
						action: 'Deshacer',
						onAction: async () => {
							for (const c of cambios) await cambiar(c.numero, c.antes);
							await invalidateAll();
						}
					}
				: {})
		});
	}

	// ---- Hoja del pedido ----
	let hoja = $state(false);
	let detalle = $state<Detalle | null>(null);
	async function abrir(numero: number) {
		detalle = null;
		hoja = true;
		const r = await fetch(`/panel/pedidos/${numero}`);
		if (r.ok) detalle = await r.json();
		else {
			hoja = false;
			ui.toast({ k: 'warn', title: 'Ese pedido no existe', desc: `No encontramos el #${numero}.` });
		}
	}
	// Tiempo real (§44): la hoja abierta se actualiza sola cuando alguien del equipo (o un pago que
	// se acredita, un envío, una factura) cambia ese pedido. La lista se recarga por su `depends`.
	async function refrescar(numero: number) {
		const r = await fetch(`/panel/pedidos/${numero}`).catch(() => null);
		if (r?.ok && hoja && detalle?.numero === numero) detalle = await r.json();
	}
	escucharTiempoReal((e) => {
		const d = detalle;
		if (!hoja || !d) return;
		const deEste = e.lote || e.id === String(d.numero);
		if (
			e.tipo === 'resincronizar' ||
			e.tipo.startsWith('comprobante.') ||
			(/^(pedido|pago|envio)\./.test(e.tipo) && deEste)
		)
			void refrescar(d.numero);
	});
	$effect(() => {
		// El Resumen manda a /panel/pedidos?pedido=1048.
		const n = Number(page.url.searchParams.get('pedido'));
		if (n) untrack(() => abrir(n));
	});
	async function avanzarDesdeHoja(p: Detalle) {
		const a = proximoDe(p);
		if (!a || !(await cambiar(p.numero, a))) return;
		await Promise.all([invalidateAll(), abrir(p.numero)]);
		ui.toast({
			k: 'ok',
			title: `#${p.numero}: ${textoEstado({ estado: a, envio: p.envio }).toLowerCase()}`,
			desc: `Le avisamos a ${p.comprador.nombre.split(' ')[0]}`,
			action: 'Deshacer',
			onAction: async () => {
				if (await cambiar(p.numero, p.estado))
					await Promise.all([invalidateAll(), abrir(p.numero)]);
			}
		});
	}
	/** La etiqueta de envío del pedido (10 × 15 cm), en otra pestaña para imprimir o guardar en PDF. */
	const imprimir = (p: Detalle) =>
		window.open(`${resolve('/panel/pedidos/imprimir')}?que=etiquetas&f=a6&n=${p.numero}`, '_blank');

	// ---- Filtros visibles ----
	let mas = $state(false);
	const ORDENES = [
		['nuevos', 'Más nuevos'],
		['urgentes', 'Urgentes primero'],
		['monto', 'Mayor monto'],
		['viejos', 'Más viejos']
	] as const;
	const ordenTxt = $derived(ORDENES.find((o) => o[0] === f.orden)![1]);
	function ciclarOrden() {
		const i = ORDENES.findIndex((o) => o[0] === f.orden);
		ir({ orden: ORDENES[(i + 1) % ORDENES.length][0] });
	}
	const FECHAS = [
		['hoy', 'Hoy'],
		['7d', '7 días'],
		['30d', '30 días'],
		['todo', 'Todo']
	] as const;
	const chips = $derived.by(() => {
		const c: [string, Partial<Filtros>][] = [];
		for (const e of f.estado) c.push([e, { estado: f.estado.filter((x) => x !== e) }]);
		for (const k of f.canal)
			c.push([NOMBRE_CANAL[k] ?? k, { canal: f.canal.filter((x) => x !== k) }]);
		for (const k of f.pago) c.push([NOMBRE_PAGO[k] ?? k, { pago: f.pago.filter((x) => x !== k) }]);
		for (const k of f.envio)
			c.push([NOMBRE_ENVIO[k] ?? k, { envio: f.envio.filter((x) => x !== k) }]);
		if (f.minimo) c.push([`Desde ${pesos(f.minimo)}`, { minimo: 0 }]);
		if (f.vip) c.push(['VIP', { vip: false }]);
		if (f.hornero) c.push(['Vendidos por Tincho', { hornero: false }]);
		if (f.mayorista) c.push(['Mayoristas', { mayorista: false }]);
		if (f.fecha !== 'todo') c.push([FECHAS.find((x) => x[0] === f.fecha)![1], { fecha: 'todo' }]);
		if (f.q) c.push([`«${f.q}»`, { q: '' }]);
		return c;
	});
	function limpiar() {
		q = '';
		ir({
			q: '',
			estado: [],
			canal: [],
			pago: [],
			envio: [],
			fecha: 'todo',
			minimo: 0,
			vip: false,
			hornero: false,
			mayorista: false
		});
	}
	// El deslizador se mueve libre y se aplica al soltarlo; vuelve al valor de la URL si cambia.
	let minimo = $derived(f.minimo);

	// ---- Textos ----
	const n = $derived(lista.total.toLocaleString('es-AR'));
	const rango = $derived(
		lista.total
			? `Mostrando ${(inicio + 1).toLocaleString('es-AR')}–${Math.min(lista.total, inicio + Math.ceil(alto / FILA)).toLocaleString('es-AR')} de ${n}`
			: chips.length
				? 'Sin resultados'
				: ''
	);
	const ahora = Date.now();
	function hace(iso: string) {
		const h = (ahora - Date.parse(iso)) / 3_600_000;
		return h < 1
			? `hace ${Math.max(1, Math.round(h * 60))} min`
			: h < 24
				? `hace ${Math.round(h)} h`
				: h < 48
					? 'ayer'
					: `hace ${Math.round(h / 24)} días`;
	}
	function urgencia(p: Fila) {
		const h = (ahora - Date.parse(p.creado)) / 3_600_000;
		if (p.estado === 'Para preparar' || p.estado === 'Listo para despachar')
			return h < 20
				? ['Despachá hoy', 'rgba(184,58,46,.1)', '#B83A2E']
				: ['Atrasado', 'rgba(184,58,46,.1)', '#B83A2E'];
		// Esperando el pago de Mercado Pago no apura: a las 24 h se cancela solo.
		return p.estado === 'Sin pagar' && !esperaPago(p) && h > 24
			? ['Sin pagar +24 h', 'rgba(227,169,59,.18)', '#9A6A10']
			: null;
	}
	const iniciales = (s: string) =>
		s
			.split(' ')
			.map((w) => w[0])
			.slice(0, 2)
			.join('');
</script>

<svelte:window bind:innerWidth={ancho} />
<!-- eslint-disable svelte/no-navigation-without-resolve -- la descarga del CSV no es una navegación -->
<Seo title="Pedidos · Panel" description="Pedidos de tu tienda" path="/panel/pedidos" noindex />

<div class="pedidos">
	<div class="cab" use:anim={{ t: 'rise' }}>
		<div class="c-txt">
			<h1>Pedidos</h1>
			<div class="c-s">
				<span class="nw"><b>{n}</b> {lista.total === 1 ? 'pedido' : 'pedidos'}</span> ·
				<span class="nw">{pesos(lista.suma)}</span>
				·
				<span class="nw">ticket promedio {pesos(lista.total ? lista.suma / lista.total : 0)}</span>
			</div>
		</div>
		<a
			class="exportar"
			href="/panel/pedidos/exportar?{escribirFiltros(f)}"
			download
			onclick={() =>
				ui.toast({
					k: 'ok',
					title: `Exportando ${n} ${lista.total === 1 ? 'pedido' : 'pedidos'}`,
					desc: 'Con los filtros aplicados · en CSV para Excel'
				})}><Icon icon={I.phDownloadSimpleBold} size={14} />Exportar {n}</a
		>
		<a class="nuevo" href="/panel/pedidos/nuevo"
			><Icon icon={I.phPlusBold} size={14} />Crear pedido</a
		>
	</div>

	<div
		class="vistas"
		use:anim={{ t: 'rise', delay: 50 }}
		role="group"
		aria-label="Vistas"
		use:deslizable
	>
		{#each VISTAS as v (v.id)}
			{@const ic = {
				todos: [I.phTrayFill, 'var(--mu)'],
				hoy: [I.phLightningFill, '#D2582F'],
				'sin-pagar': [I.phHourglassMediumFill, '#9A6A10'],
				vip: [I.phCrownSimpleFill, '#C98A12'],
				mayoristas: [I.phPackageFill, '#6A4FA0'],
				hornero: [I.phSparkleFill, '#C24E25']
			}[v.id] as [typeof I.phTrayFill, string]}
			<button
				type="button"
				class="vista"
				aria-pressed={vista === v.id}
				onclick={() => {
					q = v.filtros.q ?? '';
					ir({
						q: '',
						estado: [],
						canal: [],
						pago: [],
						envio: [],
						fecha: 'todo',
						minimo: 0,
						vip: false,
						hornero: false,
						mayorista: false,
						...v.filtros
					});
				}}><Icon icon={ic[0]} size={15} style="color:{ic[1]}" />{v.t}</button
			>
		{/each}
	</div>

	<div
		class="etapas"
		use:anim={{ t: 'rise', delay: 100 }}
		role="group"
		aria-label="Etapas"
		use:deslizable
	>
		{#each ETAPAS as e (e.estado)}
			{@const on = f.estado.includes(e.estado)}
			<button
				type="button"
				class="etapa"
				class:on
				aria-pressed={on}
				onclick={() => ir({ estado: alternar(f.estado, e.estado) })}
			>
				<span class="e-t"><span class="e-p" style:background={e.c}></span>{e.t}</span>
				<span class="e-n">{(lista.etapas[e.estado] ?? 0).toLocaleString('es-AR')}</span>
				<span class="e-b" style:background={e.c} style:opacity={on || !f.estado.length ? 1 : 0.25}
				></span>
			</button>
		{/each}
	</div>

	<div class="filtros" use:anim={{ t: 'rise', delay: 150 }}>
		<div class="f-fila">
			<label class="buscar">
				<Icon icon={I.phMagnifyingGlass} size={16} style="color:var(--mu)" />
				<input
					value={q}
					oninput={(e) => buscar(e.currentTarget.value)}
					placeholder="Cliente, N.º de pedido, producto o ciudad"
					aria-label="Buscar pedidos"
				/>
				{#if q}<button
						type="button"
						class="b-x"
						aria-label="Borrar búsqueda"
						onclick={() => buscar('')}><Icon icon={I.phXCircleFill} size={16} /></button
					>{/if}
			</label>
			<div class="fechas" role="group" aria-label="Fecha">
				{#each FECHAS as [id, t] (id)}
					<button type="button" aria-pressed={f.fecha === id} onclick={() => ir({ fecha: id })}
						>{t}</button
					>
				{/each}
			</div>
			<button
				type="button"
				class="pildora"
				class:on={mas}
				aria-expanded={mas}
				onclick={() => (mas = !mas)}
				><Icon icon={I.phSlidersHorizontalBold} size={15} />Filtros{#if chips.length}<span
						class="cuenta">{chips.length}</span
					>{/if}</button
			>
			<button type="button" class="pildora" onclick={ciclarOrden} aria-label="Orden: {ordenTxt}"
				><Icon icon={I.phSortAscendingBold} size={15} />{ordenTxt}</button
			>
		</div>
		{#if mas}
			<div class="mas" use:anim={{ t: 'rise' }}>
				<div>
					<div class="m-t">Canal</div>
					<div class="opciones">
						{#each Object.keys(NOMBRE_CANAL) as k (k)}
							{@const on = f.canal.includes(k)}
							<button
								type="button"
								class="op"
								class:on
								aria-pressed={on}
								onclick={() => ir({ canal: alternar(f.canal, k) })}
								><MarcaIcono {k} size={18} radius={5} />{NOMBRE_CANAL[k]}<span class="op-n"
									>{lista.facetas.canal[k] ?? 0}</span
								></button
							>
						{/each}
					</div>
				</div>
				<div>
					<div class="m-t">Pago</div>
					<div class="opciones">
						{#each Object.keys(NOMBRE_PAGO) as k (k)}
							{@const on = f.pago.includes(k)}
							<button
								type="button"
								class="op"
								class:on
								aria-pressed={on}
								onclick={() => ir({ pago: alternar(f.pago, k) })}
								><Icon
									icon={ICONO_PAGO[k][0]}
									size={14}
									style="color:{ICONO_PAGO[k][1]}"
								/>{NOMBRE_PAGO[k]}<span class="op-n">{lista.facetas.pago[k] ?? 0}</span></button
							>
						{/each}
					</div>
				</div>
				<div>
					<div class="m-t">Envío</div>
					<div class="opciones">
						{#each Object.keys(NOMBRE_ENVIO) as k (k)}
							{@const on = f.envio.includes(k)}
							<button
								type="button"
								class="op"
								class:on
								aria-pressed={on}
								onclick={() => ir({ envio: alternar(f.envio, k) })}
								><Icon icon={ICONO_ENVIO[k]} size={14} style="color:var(--mu)" />{NOMBRE_ENVIO[
									k
								]}<span class="op-n">{lista.facetas.envio[k] ?? 0}</span></button
							>
						{/each}
					</div>
				</div>
				<div>
					<label class="m-t m-rango" for="minimo"
						>Monto mínimo<b>{minimo ? pesos(minimo) : 'Cualquiera'}</b></label
					>
					<input
						id="minimo"
						type="range"
						min="0"
						max="200000"
						step="5000"
						bind:value={minimo}
						onchange={() => ir({ minimo })}
					/>
				</div>
				<div class="banderas">
					{#each [['vip', 'Solo clientes VIP'], ['hornero', 'Vendidos por Tincho'], ['mayorista', 'Mayoristas']] as const as [k, t] (k)}
						<div class="bandera">
							<span>{t}</span>
							<Interruptor label={t} checked={f[k]} onchange={(v: boolean) => ir({ [k]: v })} />
						</div>
					{/each}
				</div>
			</div>
		{/if}
		{#if chips.length}
			<div class="chips">
				{#each chips as [t, cambio] (t)}
					<button
						type="button"
						class="chip"
						onclick={() => {
							if ('q' in cambio) q = '';
							ir(cambio);
						}}>{t}<Icon icon={I.phXBold} size={11} /></button
					>
				{/each}
				<button type="button" class="limpiar" onclick={limpiar}>Limpiar todo</button>
			</div>
		{/if}
	</div>

	<div class="lista" use:anim={{ t: 'rise', delay: 200 }}>
		<div class="l-cab">
			<button
				type="button"
				class="tilde"
				aria-label="Seleccionar los visibles"
				onclick={tildarVisibles}
			></button>
			<span class="l-r">{rango}</span>
		</div>
		<div
			class="scroll"
			style:height="{altoCaja}px"
			bind:this={caja}
			onscroll={alScrollear}
			data-lenis-prevent
			role="list"
			aria-label="Pedidos"
		>
			<div style:height="{lista.total * FILA}px" class="alto">
				{#each visibles as i (i)}
					{@const p = filas[i]}
					<div class="fila" style:top="{i * FILA}px" role="listitem">
						{#if p}
							{@const e = etapaDePedido(p)}
							{@const paso = ETAPAS.findIndex((x) => x.estado === p.estado)}
							{@const sel = seleccion.has(p.numero)}
							{@const u = urgencia(p)}
							{@const a = accionRapida(p)}
							{@const wa = whatsappDe(p, contexto)}
							<div class="f-in" class:sel>
								<button
									type="button"
									class="tilde"
									class:on={sel}
									aria-label="Seleccionar #{p.numero}"
									aria-pressed={sel}
									onclick={() => tildar(p)}
									>{#if sel}<Icon icon={I.phCheckBold} size={12} />{/if}</button
								>
								<button type="button" class="abrir" onclick={() => abrir(p.numero)}>
									<span class="av">
										{#if p.foto}<img src={cara(p.foto)} alt="" />{:else}<span class="av-ini"
												>{iniciales(p.cliente)}</span
											>{/if}
										<span class="av-canal"><MarcaIcono k={p.canal} size={20} radius={7} /></span>
									</span>
									<span class="quien">
										<span class="q-n"
											><span class="q-c">{p.cliente}</span>{#if p.vip}<Icon
													icon={I.phCrownSimpleFill}
													size={13}
													style="color:#C98A12;flex:none"
												/>{/if}{#if p.hornero}<span class="q-ai" title="Lo vendió Tincho">✦</span
												>{/if}</span
										>
										<!-- En el celular el monto pasa a este renglón: el nombre y los botones de 44 px entran. -->
										<span class="q-s"
											>#{p.numero} · {#if celu}{pesos(p.total)} ·
											{/if}{hace(p.creado)} · {p.ciudad}</span
										>
									</span>
									{#if amplio}
										<span class="prods">
											<span class="miniaturas"
												>{#each p.productos as x, j (j)}{#if x.foto}<img
															src={unsplash(x.foto, 80, '&h=80')}
															alt=""
														/>{/if}{/each}</span
											>
											<span class="p-t"
												>{p.cantidadProductos === 1
													? p.productos[0].nombre
													: `${p.cantidadProductos} productos`}</span
											>
										</span>
										<span class="pasos">
											<span class="barras"
												>{#each ['Pagado', 'Preparado', 'Enviado', 'Entregado'] as t, j (t)}<span
														title={t}
														style:background={paso > 0 && j < paso ? e.c : 'var(--sf3)'}
													></span>{/each}</span
											>
											<span class="p-e" style:color={e.c}
												><Icon icon={e.i} size={13} />{textoEstado(p)}</span
											>
										</span>
									{/if}
									{#if u}<span class="urg" style:background={u[1]} style:color={u[2]}>{u[0]}</span
										>{/if}
									<span class="monto" class:oculto={celu}>
										<span class="m-v">{pesos(p.total)}</span>
										<span class="m-p"
											><Icon
												icon={ICONO_PAGO[p.pago]?.[0] ?? I.phMoneyFill}
												size={12}
												style="color:{ICONO_PAGO[p.pago]?.[1]}"
											/>{NOMBRE_PAGO[p.pago] ?? p.pago}</span
										>
									</span>
								</button>
								<!-- Ancho fijo: con o sin WhatsApp, las columnas de la lista quedan alineadas. -->
								<span class="acciones">
									{#if wa}
										<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- link externo a WhatsApp -->
										<a
											class="wa"
											href={wa}
											target="_blank"
											rel="noopener noreferrer"
											title="Escribirle por WhatsApp"
											aria-label="WhatsApp a {p.cliente} por el #{p.numero}"
											><Icon icon={I.phWhatsappLogoFill} size={17} /></a
										>
									{/if}
									{#if a}
										<button
											type="button"
											class="accion"
											style:background={a.fondo}
											style:color={a.color}
											aria-label="{a.texto} #{p.numero}"
											title={amplio ? undefined : a.texto}
											disabled={enCurso.has(p.numero)}
											onclick={() => accion(p)}
											><Icon
												icon={a.icono}
												size={celu ? 19 : 14}
											/>{#if amplio}{a.texto}{/if}</button
										>
									{/if}
								</span>
							</div>
						{:else}
							<div class="f-in esqueleto"><span></span><span></span></div>
						{/if}
					</div>
				{/each}
			</div>
			{#if !lista.total}
				<div class="vacio">
					{#if chips.length}
						<div>
							<img src="/hornero/lupa.webp" alt="" width="110" height="110" />
							<div class="v-t">Ningún pedido con esos filtros</div>
							<button type="button" class="v-b" onclick={limpiar}>Limpiar filtros</button>
						</div>
					{:else}
						<!-- Sin filtros: todavía no entró ninguno. Qué hacer para que entre el primero. -->
						<div>
							<img src="/hornero/horno-vacio.webp" alt="" width="110" height="110" />
							<div class="v-t">Todavía no entró ningún pedido</div>
							<p class="v-x">
								Los de la tienda online, Mercado Libre y el local aparecen acá. Compartí tu tienda o
								cargá una venta a mano.
							</p>
							<div class="v-bs">
								<a class="v-b" href={resolve('/panel/pedidos/nuevo')}>Cargar un pedido</a>
								<button
									type="button"
									class="v-b v-b2"
									onclick={() => compartirTienda(ui, page.data.tienda)}>Compartir mi tienda</button
								>
							</div>
						</div>
					{/if}
				</div>
			{/if}
		</div>
		{#if scroll > 1200}
			<button
				type="button"
				class="arriba"
				aria-label="Volver arriba"
				onclick={() => caja?.scrollTo({ top: 0, behavior: 'smooth' })}
				><Icon icon={I.phArrowUpBold} size={16} /></button
			>
		{/if}
	</div>
</div>

{#if seleccion.size}
	<div class="lote" use:anim={{ t: 'rise' }} use:deslizable>
		<span class="lo-t"
			>{seleccion.size}
			{seleccion.size === 1 ? 'seleccionado' : 'seleccionados'} · {pesos(sumaSel)}</span
		>
		{#each LOTE as l (l.accion)}
			<button type="button" onclick={() => lote(l)}><Icon icon={l.i} size={14} />{l.t}</button>
		{/each}
		<button
			type="button"
			class="lo-x"
			aria-label="Quitar la selección"
			onclick={() => seleccion.clear()}><Icon icon={I.phXBold} size={14} /></button
		>
	</div>
{/if}

<!-- eslint-enable svelte/no-navigation-without-resolve -->
<HojaPedido bind:open={hoja} pedido={detalle} avanzar={avanzarDesdeHoja} {imprimir} {contexto} />

<style>
	.pedidos {
		display: flex;
		flex-direction: column;
		gap: 14px;
	}
	.cab {
		display: flex;
		align-items: end;
		gap: 12px;
		flex-wrap: wrap;
	}
	.c-txt {
		flex: 1;
		min-width: 220px;
	}
	h1 {
		margin: 0;
		font-size: 30px;
		letter-spacing: -0.035em;
		font-weight: 600;
	}
	.c-s {
		margin-top: 2px;
		font-size: 15px;
		color: var(--mu);
	}
	.nw {
		white-space: nowrap;
	}
	.c-s b {
		color: var(--tx);
	}
	.exportar {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 10px 16px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: var(--sf);
		color: var(--tx);
		font-size: 13.5px;
		font-weight: 600;
		box-shadow: var(--sh);
		white-space: nowrap;
	}
	.nuevo {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 10px 16px;
		border-radius: 999px;
		background: var(--tx);
		color: var(--sf);
		font-size: 13.5px;
		font-weight: 600;
		white-space: nowrap;
		text-decoration: none;
	}
	.vistas {
		display: flex;
		gap: 6px;
		overflow-x: auto;
		scrollbar-width: none;
	}
	.vista {
		display: flex;
		align-items: center;
		gap: 7px;
		border: 1.5px solid transparent;
		white-space: nowrap;
		font-size: 13px;
		font-weight: 600;
		padding: 8px 13px;
		border-radius: 999px;
		background: var(--sf);
		color: var(--tx);
	}
	.vista[aria-pressed='true'] {
		border-color: #d2582f;
		background: rgba(210, 88, 47, 0.08);
	}
	.etapas {
		display: flex;
		border-radius: 18px;
		overflow: hidden;
		background: var(--sf);
		box-shadow: var(--shd);
	}
	.etapa {
		flex: 1;
		min-width: 0;
		padding: 12px 14px;
		border: 0;
		border-right: 1px solid var(--ln);
		background: transparent;
		color: var(--tx);
		text-align: left;
		transition: background 0.2s;
	}
	.etapa:last-child {
		border-right: 0;
	}
	.etapa.on {
		background: var(--chip);
	}
	.e-t {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 12.5px;
		color: var(--mu);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.e-p {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		flex: none;
	}
	.e-n {
		display: block;
		font-size: 22px;
		font-weight: 600;
		letter-spacing: -0.03em;
		margin-top: 2px;
	}
	.e-b {
		display: block;
		margin-top: 6px;
		height: 4px;
		border-radius: 3px;
	}
	.filtros {
		display: flex;
		flex-direction: column;
		gap: 10px;
		padding: 12px;
		border-radius: 20px;
		background: var(--sf);
		box-shadow: var(--shd);
	}
	.f-fila {
		display: flex;
		gap: 8px;
		flex-wrap: wrap;
		align-items: center;
	}
	.buscar {
		flex: 1 1 240px;
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 10px 14px;
		border-radius: 999px;
		background: var(--sf3);
	}
	.buscar input {
		flex: 1;
		min-width: 0;
		border: 0;
		outline: 0;
		background: transparent;
		font-size: 14px;
		font-weight: 500;
		color: var(--tx);
	}
	.buscar:focus-within {
		box-shadow: 0 0 0 2px rgba(210, 88, 47, 0.35);
	}
	.b-x {
		border: 0;
		background: transparent;
		color: var(--mu);
		display: grid;
		place-items: center;
		padding: 0;
	}
	.fechas {
		display: flex;
		padding: 3px;
		border-radius: 999px;
		background: var(--chip);
	}
	.fechas button {
		border: 0;
		font-size: 12.5px;
		font-weight: 600;
		padding: 7px 12px;
		border-radius: 999px;
		background: transparent;
		color: var(--tx);
		white-space: nowrap;
	}
	.fechas button[aria-pressed='true'] {
		background: var(--sf);
		box-shadow: var(--sh);
	}
	.pildora {
		display: flex;
		align-items: center;
		gap: 7px;
		padding: 9px 14px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: var(--sf);
		color: var(--tx);
		font-size: 13px;
		font-weight: 600;
		white-space: nowrap;
	}
	.pildora.on {
		border-color: #d2582f;
		background: rgba(210, 88, 47, 0.08);
	}
	.cuenta {
		min-width: 18px;
		height: 18px;
		padding: 0 5px;
		border-radius: 999px;
		background: #b0441d;
		color: #fff;
		font-size: 11px;
		font-weight: 700;
		display: grid;
		place-items: center;
	}
	.mas {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
		gap: 12px;
		padding: 6px 2px 2px;
	}
	.m-t {
		font-size: 12px;
		font-weight: 600;
		color: var(--mu);
		margin-bottom: 6px;
	}
	.m-rango {
		display: flex;
		justify-content: space-between;
	}
	.m-rango b {
		color: var(--tx);
	}
	.mas input[type='range'] {
		width: 100%;
		accent-color: #d2582f;
	}
	.opciones {
		display: flex;
		flex-wrap: wrap;
		gap: 5px;
	}
	.op {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 6px 10px;
		border-radius: 999px;
		border: 1.5px solid var(--ln2);
		background: transparent;
		color: var(--tx);
		font-size: 12.5px;
		font-weight: 600;
		white-space: nowrap;
	}
	.op.on {
		border-color: #d2582f;
		background: rgba(210, 88, 47, 0.08);
	}
	.op-n {
		color: var(--fa);
		font-weight: 500;
	}
	.banderas {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.bandera {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 4px 0;
		font-size: 13px;
		font-weight: 500;
	}
	.bandera span {
		flex: 1;
	}
	.chips {
		display: flex;
		gap: 6px;
		flex-wrap: wrap;
		align-items: center;
	}
	.chip {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 5px 8px 5px 11px;
		border-radius: 999px;
		border: 0;
		background: rgba(210, 88, 47, 0.12);
		color: #b0441d;
		font-size: 12.5px;
		font-weight: 600;
	}
	.limpiar {
		border: 0;
		background: transparent;
		font-size: 12.5px;
		font-weight: 600;
		color: var(--mu);
		text-decoration: underline;
	}
	.lista {
		position: relative;
		border-radius: 24px;
		background: var(--sf);
		box-shadow: var(--shd);
		overflow: hidden;
	}
	.l-cab {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 10px 16px;
		border-bottom: 1px solid var(--ln);
		font-size: 12.5px;
		color: var(--mu);
	}
	.l-r {
		flex: 1;
	}
	.tilde {
		width: 20px;
		height: 20px;
		flex: none;
		padding: 0;
		border-radius: 6px;
		border: 1.5px solid var(--ln2);
		background: var(--sf);
		color: #fff;
		display: grid;
		place-items: center;
	}
	.tilde.on {
		border-color: #d2582f;
		background: #d2582f;
	}
	.scroll {
		overflow-y: auto;
		overscroll-behavior: contain;
		position: relative;
	}
	.alto {
		position: relative;
	}
	.fila {
		position: absolute;
		left: 0;
		right: 0;
		height: 76px;
		padding: 5px 10px;
	}
	.f-in {
		height: 100%;
		display: flex;
		align-items: center;
		gap: 14px;
		padding: 0 12px;
		border-radius: 16px;
		border: 1px solid transparent;
		transition: background 0.15s;
	}
	.f-in:hover {
		background: var(--chip);
	}
	.f-in.sel {
		background: rgba(210, 88, 47, 0.06);
		border-color: rgba(210, 88, 47, 0.3);
	}
	.abrir {
		flex: 1;
		min-width: 0;
		height: 100%;
		display: flex;
		align-items: center;
		gap: 14px;
		border: 0;
		padding: 0;
		background: transparent;
		color: var(--tx);
		text-align: left;
	}
	.av {
		position: relative;
		flex: none;
	}
	.av img,
	.av-ini {
		width: 40px;
		height: 40px;
		border-radius: 50%;
		object-fit: cover;
		display: block;
		background: var(--sf3);
	}
	.av-ini {
		display: grid;
		place-items: center;
		font-size: 12px;
		font-weight: 600;
		color: var(--mu);
	}
	.av-canal {
		position: absolute;
		right: -4px;
		bottom: -4px;
		border-radius: 7px;
		box-shadow: 0 0 0 2px var(--sf);
		display: block;
	}
	.quien {
		flex: 1 1 150px;
		min-width: 0;
	}
	.q-n {
		display: flex;
		align-items: center;
		gap: 6px;
	}
	.q-c {
		font-size: 14px;
		font-weight: 600;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.q-ai {
		font-size: 10.5px;
		font-weight: 700;
		color: #b0441d;
	}
	.q-s {
		display: block;
		font-size: 12px;
		color: var(--fa);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.prods {
		display: flex;
		align-items: center;
		gap: 8px;
		width: 170px;
		flex: none;
	}
	.miniaturas {
		display: flex;
		padding-left: 6px;
	}
	.miniaturas img {
		width: 30px;
		height: 30px;
		border-radius: 8px;
		object-fit: cover;
		margin-left: -6px;
		box-shadow: 0 0 0 2px var(--sf);
	}
	.p-t {
		font-size: 12px;
		color: var(--mu);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.pasos {
		width: 150px;
		flex: none;
	}
	.barras {
		display: flex;
		gap: 3px;
	}
	.barras span {
		flex: 1;
		height: 5px;
		border-radius: 3px;
	}
	.p-e {
		display: flex;
		align-items: center;
		gap: 5px;
		margin-top: 5px;
		font-size: 12px;
		font-weight: 600;
		white-space: nowrap;
	}
	.urg {
		flex: none;
		font-size: 11.5px;
		font-weight: 700;
		padding: 4px 8px;
		border-radius: 999px;
		white-space: nowrap;
	}
	.monto {
		flex: none;
		min-width: 84px;
		text-align: right;
	}
	.m-v {
		display: block;
		font-size: 15px;
		font-weight: 600;
		letter-spacing: -0.01em;
		white-space: nowrap;
	}
	.m-p {
		display: flex;
		justify-content: flex-end;
		align-items: center;
		gap: 4px;
		font-size: 11.5px;
		color: var(--mu);
		white-space: nowrap;
	}
	.accion {
		flex: none;
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 8px 12px;
		border-radius: 999px;
		border: 0;
		font-size: 12.5px;
		font-weight: 600;
		white-space: nowrap;
	}
	.acciones {
		flex: none;
		width: 204px;
		display: flex;
		align-items: center;
		justify-content: flex-end;
		gap: 8px;
	}
	.accion:disabled {
		opacity: 0.6;
	}
	/* WhatsApp a quien compró, con el mensaje del estado (los colores de «Reclamar pago» del handoff). */
	.wa {
		flex: none;
		width: 34px;
		height: 34px;
		border-radius: 50%;
		display: grid;
		place-items: center;
		background: #e9f8ee;
		color: #1f6f4a;
	}
	.esqueleto span {
		height: 12px;
		border-radius: 6px;
		background: var(--sf3);
		flex: 1;
	}
	.esqueleto span:first-child {
		flex: none;
		width: 40px;
		height: 40px;
		border-radius: 50%;
	}
	.vacio {
		position: absolute;
		inset: 0;
		display: grid;
		place-items: center;
		text-align: center;
		padding: 30px;
	}
	.vacio img {
		object-fit: contain;
	}
	.v-t {
		font-size: 16px;
		font-weight: 600;
		margin-top: 6px;
	}
	.v-b {
		display: inline-block;
		text-decoration: none;
		margin-top: 10px;
		padding: 9px 16px;
		border-radius: 999px;
		border: 0;
		background: var(--tx);
		color: var(--sf);
		font-family: inherit;
		font-size: 13px;
		font-weight: 600;
		cursor: pointer;
	}
	.v-b2 {
		background: var(--sf3);
		color: var(--tx);
	}
	.v-bs {
		display: flex;
		gap: 8px;
		justify-content: center;
		flex-wrap: wrap;
	}
	.v-x {
		max-width: 340px;
		margin: 6px auto 0;
		font-size: 13px;
		line-height: 1.45;
		color: var(--fa);
		text-wrap: pretty;
	}
	.arriba {
		position: absolute;
		right: 16px;
		bottom: 16px;
		width: 40px;
		height: 40px;
		border-radius: 50%;
		border: 0;
		background: var(--tx);
		color: var(--sf);
		display: grid;
		place-items: center;
		box-shadow: 0 10px 24px -8px rgba(42, 20, 12, 0.5);
	}
	.lote {
		position: fixed;
		left: 50%;
		bottom: 84px;
		translate: -50% 0;
		z-index: 60;
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 8px 8px 8px 18px;
		border-radius: 999px;
		background: #2a140c;
		color: #fff;
		box-shadow: 0 24px 50px -16px rgba(42, 20, 12, 0.7);
		max-width: calc(100vw - 16px);
		overflow-x: auto;
		scrollbar-width: none;
	}
	.lo-t {
		font-size: 14px;
		font-weight: 600;
		white-space: nowrap;
		margin-right: 6px;
	}
	.lote button {
		display: flex;
		align-items: center;
		gap: 6px;
		border: 0;
		font-size: 13px;
		font-weight: 600;
		padding: 9px 13px;
		border-radius: 999px;
		background: rgba(255, 255, 255, 0.1);
		color: #fff;
		white-space: nowrap;
	}
	.lote .lo-x {
		width: 34px;
		height: 34px;
		flex: none;
		padding: 0;
		justify-content: center;
		background: transparent;
		color: #d9c8ba;
	}
	@media (max-width: 899px) {
		.f-in,
		.abrir {
			gap: 10px;
		}
		.acciones {
			width: auto;
		}
	}
	@media (max-width: 759px) {
		/* En el celular la fila deja lugar al nombre: el medio de pago y la urgencia se ven en el
		   detalle (y el ícono del canal ya está en la foto); el monto va debajo del nombre. */
		.m-p,
		.urg,
		.monto.oculto {
			display: none;
		}
		.fila {
			padding: 5px 6px;
		}
		.acciones {
			width: 96px;
		}
		.f-in {
			gap: 8px;
			padding: 0 6px 0 10px;
		}
		/* En el handoff las etapas se cortan: acá se desplazan. */
		.etapas {
			overflow-x: auto;
			scrollbar-width: none;
		}
		.etapa {
			flex: 1 0 120px;
		}
		.lote {
			bottom: 92px;
		}
		/* Las acciones de la fila, a un toque: botones visibles de 44 px. */
		.wa,
		.accion {
			width: 44px;
			height: 44px;
			padding: 0;
			justify-content: center;
		}
		.urg {
			display: none;
		}
	}

	/* Área de toque de 44 px en pantallas táctiles, sin cambiar el dibujo del handoff. */
	@media (pointer: coarse) {
		.tilde,
		.accion {
			position: relative;
		}
		.tilde::after {
			content: '';
			position: absolute;
			inset: -12px;
		}
		.accion::after {
			content: '';
			position: absolute;
			inset: -7px -3px;
		}
	}
</style>
