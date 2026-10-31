<script lang="ts">
	import { invalidateAll, replaceState } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import Icon from '#lib/components/Icon.svelte';
	import Modal from '#lib/components/Modal.svelte';
	import Segmentado from '#lib/components/Segmentado.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import EditorProducto from '#lib/panel/listas/EditorProducto.svelte';
	import MotorLista from '#lib/panel/listas/MotorLista.svelte';
	import { compartirProducto } from '#lib/panel/compartir.ts';
	import { configProductos, type ProductoFila } from '#lib/panel/listas/productos.ts';
	import { SITE } from '#lib/site.ts';
	import { getUi } from '#lib/ui.svelte.ts';
	import { escucharTiempoReal } from '#lib/tiempo-real/panel.svelte.ts';
	import type { components } from '@mostrador/contrato/esquema';

	type Producto = components['schemas']['Producto'];
	type Importacion = components['schemas']['Importacion'];

	/**
	 * Productos: la lista de PanelListas más lo de PanelCatalogo (cargar con foto, importar Excel y
	 * el editor lateral).
	 */
	let { data } = $props();
	const publicado = (n: number) =>
		n === 0 ? 'sin publicar' : n === 1 ? 'publicado en 1 canal' : `publicado en ${n} canales`;
	const ui = getUi();
	const tienda = $derived(page.data.tienda as { nombre: string; slug: string });
	const dominio = $derived(`${SITE.replace(/^https?:\/\//, '')}/t/${tienda.slug}`);

	let editor = $state(false);
	let producto = $state<Producto | null>(null);
	let porcentaje = $state<string[]>([]);
	let cambiandoPrecio = $state(false);
	const config = configProductos(
		() => {
			producto = null;
			editor = true;
		},
		(ids) => {
			porcentaje = ids;
			cambiandoPrecio = true;
		},
		(p, ctx) =>
			compartirProducto(ui, tienda.slug, p, async () => {
				await ctx.accion([p.id], 'publicar', true);
				ui.toast({
					k: 'ok',
					title: `${p.nombre} publicado`,
					desc: 'Ya lo podés compartir',
					action: 'Compartir',
					onAction: () => compartirProducto(ui, tienda.slug, { ...p, publicado: true })
				});
			})
	);

	async function abrir(p: ProductoFila) {
		const r = await fetch(`/panel/productos/${encodeURIComponent(p.id)}`);
		if (!r.ok)
			return ui.toast({
				k: 'warn',
				title: 'Ese producto no existe',
				desc: 'Actualizamos la lista.'
			});
		producto = await r.json();
		editor = true;
	}

	let pct = $state(8);
	async function aplicarPrecio(e: SubmitEvent) {
		e.preventDefault();
		const ids = porcentaje;
		const r = await fetch('/panel/listas/productos', {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ ids, accion: 'precio', valor: pct })
		});
		cambiandoPrecio = false;
		await invalidateAll();
		if (!r.ok)
			return ui.toast({ k: 'err', title: 'No se pudo', desc: 'Probá de nuevo en un rato.' });
		const { hechos } = (await r.json()) as { hechos: number };
		const vuelta = Math.round((100 / (1 + pct / 100) - 100) * 100) / 100;
		ui.toast({
			k: 'ok',
			title: `Precio ${pct > 0 ? '+' : ''}${pct}% · ${hechos} ${hechos === 1 ? 'producto' : 'productos'}`,
			desc: 'Redondeado a $ 10',
			action: 'Deshacer',
			onAction: async () => {
				await fetch('/panel/listas/productos', {
					method: 'POST',
					headers: { 'content-type': 'application/json' },
					body: JSON.stringify({ ids, accion: 'precio', valor: vuelta })
				});
				await invalidateAll();
			}
		});
	}

	// Importar tienda: todo va por la cola de importación (productos y, en Tiendanube, clientes y
	// pedidos). Tiendanube se conecta por OAuth; Shopify, WooCommerce y Empretienda, con el link de la
	// tienda pública. En la demo cada "avanzar" hace un paso; con la API avanza sola y "avanzar" sólo
	// lee, así que si nada cambió se espera un poco antes de volver a preguntar.
	type Fuente = 'shopify' | 'woocommerce' | 'empretienda';
	const FUENTES = [
		['shopify', 'Shopify'],
		['woocommerce', 'WooCommerce'],
		['empretienda', 'Empretienda']
	] as const;
	const NOMBRE: Record<string, string> = {
		...Object.fromEntries(FUENTES),
		tiendanube: 'Tiendanube',
		mercadolibre: 'Mercado Libre'
	};
	const EJEMPLO: Record<Fuente, string> = {
		shopify: 'mitienda.myshopify.com',
		woocommerce: 'mitienda.com.ar',
		empretienda: 'mitienda.empretienda.com.ar'
	};
	let importando = $state(false);
	let fuente = $state<Fuente>('shopify');
	let tiendaImp = $state('');
	let buscando = $state(false);
	let errorImp = $state('');
	const post = (url: string, cuerpo: unknown) =>
		fetch(url, {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify(cuerpo)
		});
	const mensaje = async (r: Response, otro: string) =>
		((await r.json().catch(() => null)) as { message?: string } | null)?.message ?? otro;
	function abrirImportar() {
		errorImp = '';
		importando = true;
		void leerCola();
	}

	let cola = $state<Importacion | null>(null);
	let tiendanube = $state<boolean | null>(null);
	let avisoTN = $state('');
	const activa = (i: Importacion | null) => i?.estado === 'pendiente' || i?.estado === 'corriendo';
	const AVISOS: Record<string, string> = {
		'no-disponible': 'La conexión con Tiendanube todavía no está disponible.',
		cancelada: 'No se conectó la tienda: Tiendanube no dio el permiso.',
		estado: 'La conexión venció. Probá de nuevo.',
		canje: 'Tiendanube no aceptó la conexión. Probá de nuevo.',
		activa: 'Ya hay una importación en curso: esperá que termine.',
		'sin-cifrado': 'No podemos guardar la clave de tu tienda. Probá más tarde.',
		'sin-tiendanube': 'La conexión con Tiendanube todavía no está disponible.'
	};
	async function leerCola() {
		const r = await fetch('/panel/importaciones').catch(() => null);
		if (!r?.ok) return;
		const d = (await r.json()) as { importacion: Importacion | null; tiendanube: boolean };
		tiendanube = d.tiendanube;
		// La que está en curso (o la que se venía mirando) se sigue; una vieja terminada no se muestra.
		if (d.importacion && (activa(d.importacion) || cola)) {
			cola = d.importacion;
			if (activa(cola)) void seguir();
		}
	}
	// Tiempo real (§44): el avance de una importación que corre en otra pestaña, en otra persona del
	// equipo o en el worker se ve sin tocar nada (sólo se muestra: no se empuja desde acá).
	escucharTiempoReal(async (e) => {
		if (!e.tipo.startsWith('importacion.') || !importando || siguiendo) return;
		const r = await fetch('/panel/importaciones').catch(() => null);
		if (!r?.ok) return;
		const d = (await r.json()) as { importacion: Importacion | null };
		if (d.importacion && (activa(d.importacion) || cola)) cola = d.importacion;
	});
	let siguiendo = false;
	async function seguir() {
		if (siguiendo) return;
		siguiendo = true;
		try {
			while (importando && activa(cola)) {
				const antes = JSON.stringify(cola);
				const r = await post('/panel/importaciones', { accion: 'avanzar' }).catch(() => null);
				if (!r?.ok) break;
				cola = (await r.json()) as Importacion | null;
				if (JSON.stringify(cola) === antes) await new Promise((ok) => setTimeout(ok, 1500));
			}
		} finally {
			siguiendo = false;
		}
		if (cola && !activa(cola)) await invalidateAll();
	}
	async function cancelarCola() {
		if (!cola) return;
		const r = await post('/panel/importaciones', { accion: 'cancelar', id: cola.id });
		if (r.ok) cola = (await r.json()) as Importacion;
	}
	function cerrarCola() {
		cola = null;
		importando = false;
	}
	// La vuelta del OAuth trae ?tiendanube=importando (o el motivo del error), y el onboarding
	// ?importando=1 (arrancó) o ?importar=1 (no pudo leer la tienda): se abre el modal.
	$effect(() => {
		const tn = page.url.searchParams.get('tiendanube');
		const onb = page.url.searchParams.get('importando');
		const reintentar = page.url.searchParams.get('importar');
		if (!tn && !onb && !reintentar) return;
		replaceState(resolve('/panel/productos'), {});
		errorImp = reintentar
			? 'No pudimos leer tu tienda con ese link. Revisalo y probá de nuevo.'
			: '';
		avisoTN = !tn || tn === 'importando' ? '' : (AVISOS[tn] ?? 'No pudimos conectar tu tienda.');
		importando = true;
		if (tn === 'importando' || onb) cola = { estado: 'pendiente' } as Importacion;
		void leerCola();
	});
	// ?nuevo=1 (desde el Inicio o ⌘K): abre el editor de un producto nuevo de una.
	// Una sola vez por llegada: al recargar los datos después de guardar no se vuelve a abrir.
	let nuevoAbierto = false;
	$effect(() => {
		if (nuevoAbierto || !page.url.searchParams.has('nuevo')) return;
		nuevoAbierto = true;
		replaceState(resolve('/panel/productos'), {});
		producto = null;
		editor = true;
	});
	// ?editar=<id> (el botón «Reponer stock» del aviso de sin stock): abre ese producto.
	$effect(() => {
		const id = page.url.searchParams.get('editar');
		if (!id) return;
		replaceState(resolve('/panel/productos'), {});
		void abrir({ id } as ProductoFila);
	});
	/** Las etapas de cada plataforma: de un link, sin cuenta, sólo salen productos. */
	const etapasCola = $derived(
		cola?.plataforma === 'tiendanube'
			? ([
					['productos', 'Productos'],
					['clientes', 'Clientes'],
					['pedidos', 'Pedidos']
				] as const)
			: ([['productos', 'Productos']] as const)
	);
	/**
	 * Cuánto va. Tiendanube, por etapa (no dice el total hasta leer cada una); un link, por productos
	 * leídos, sin llegar al final hasta que termina.
	 */
	const avanceCola = $derived(
		!cola || !cola.etapa
			? 4
			: cola.estado === 'lista'
				? 100
				: cola.plataforma === 'tiendanube'
					? Math.max(
							4,
							(['productos', 'clientes', 'pedidos', 'redirecciones'].indexOf(cola.etapa) + 0.5) * 25
						)
					: Math.min(90, 4 + (cola.progreso?.productos ?? 0) / 3)
	);
	/** Shopify y WooCommerce no publican el stock de verdad: se avisa al terminar. */
	const sinStock = $derived(
		cola?.estado === 'lista' && (cola.plataforma === 'shopify' || cola.plataforma === 'woocommerce')
	);
	/** Con el link: primero de qué plataforma es (la que delata el dominio o la que responda). */
	async function importarTienda() {
		errorImp = '';
		buscando = true;
		try {
			const r = await post('/panel/productos/tienda', { fuente, tienda: tiendaImp });
			if (!r.ok) return void (errorImp = await mensaje(r, 'No pudimos leer la tienda.'));
			const d = (await r.json()) as { fuente: Fuente; origen: string };
			fuente = d.fuente;
			const c = await post('/panel/importaciones', {
				accion: 'crear',
				plataforma: d.fuente,
				url: d.origen
			});
			if (!c.ok) return void (errorImp = await mensaje(c, 'No pudimos empezar la importación.'));
			cola = (await c.json()) as Importacion;
			void seguir();
		} finally {
			buscando = false;
		}
	}

	let archivo = $state<HTMLInputElement>();
	async function importar(e: Event) {
		// currentTarget deja de existir después del primer await: se guarda antes.
		const input = e.currentTarget as HTMLInputElement;
		const f = input.files?.[0];
		if (!f) return;
		if (!/\.(csv|txt)$/i.test(f.name)) {
			ui.toast({
				k: 'warn',
				title: 'Guardalo como CSV',
				desc: 'En Excel: Archivo › Guardar como › CSV.'
			});
			return;
		}
		const r = await fetch('/panel/productos/importar', { method: 'POST', body: await f.text() });
		input.value = '';
		if (!r.ok) {
			const m = ((await r.json().catch(() => null)) as { message?: string } | null)?.message;
			return ui.toast({
				k: 'err',
				title: 'No pudimos importarlo',
				desc: m ?? 'Revisá el archivo.'
			});
		}
		const { creados, errores } = (await r.json()) as { creados: number; errores: string[] };
		await invalidateAll();
		ui.toast({
			k: 'ok',
			title: `${creados} ${creados === 1 ? 'producto importado' : 'productos importados'}`,
			desc: errores.length
				? `${errores.length === 1 ? 'Una fila sin nombre quedó' : `${errores.length} filas sin nombre quedaron`} afuera.`
				: 'Ya están publicados en tu tienda.'
		});
	}
</script>

<Seo
	title="Productos · Panel"
	description="Productos de tu tienda"
	path="/panel/productos"
	noindex
/>

<input bind:this={archivo} type="file" accept=".csv,text/csv" hidden onchange={importar} />

<Modal bind:open={importando} titulo="Importar tienda" ancho={460} padding={24}>
	<span class="m-ico"><Icon icon={I.phStorefrontFill} size={26} style="color:#2E2A6B" /></span>
	{#if cola}
		<div class="m-t" role="status">
			{cola.estado === 'lista'
				? `Tu tienda de ${NOMBRE[cola.plataforma] ?? 'tu plataforma'} ya está\u00a0acá`
				: cola.estado === 'fallida'
					? 'La importación se\u00a0cortó'
					: cola.estado === 'cancelada'
						? 'Frenaste la\u00a0importación'
						: `Importando desde\u00a0${NOMBRE[cola.plataforma] ?? 'tu tienda'}`}
		</div>
		<div class="m-d">
			{activa(cola)
				? 'Dejá esta ventana abierta mientras traemos\u00a0todo.'
				: cola.estado === 'lista'
					? 'Revisá precios y stock antes de\u00a0vender.'
					: 'Lo que alcanzó a entrar ya está en tu\u00a0tienda.'}
		</div>
		<div
			class="m-barra"
			role="progressbar"
			aria-label="Avance de la importación"
			aria-valuenow={avanceCola}
			aria-valuemax={100}
		>
			<span style="width:{avanceCola}%"></span>
		</div>
		<ul class="m-etapas">
			{#each etapasCola as [k, nombre] (k)}
				<li
					class="chip-estado"
					style:--chip-bg={cola.etapa === k && activa(cola) ? 'rgba(210,88,47,.13)' : 'var(--chip)'}
					style:--chip-fg={cola.etapa === k && activa(cola) ? '#C24E25' : 'var(--mu)'}
				>
					{nombre}
					{cola.progreso?.[k] ?? 0}
				</li>
			{/each}
		</ul>
		{#if cola.progreso?.repetidos || cola.progreso?.recortados}
			<div class="m-d">
				{[
					cola.progreso.repetidos && `${cola.progreso.repetidos} ya estaban`,
					cola.progreso.recortados && `${cola.progreso.recortados} recortados`
				]
					.filter(Boolean)
					.join(' · ')}
			</div>
		{/if}
		{#if sinStock}
			<div class="m-d">
				{NOMBRE[cola.plataforma]} no publica el stock: revisalo antes de&nbsp;vender.
			</div>
		{/if}
		{#if cola.errores?.length}
			<ul class="m-notas m-err" role="alert">
				{#each cola.errores.slice(-3) as e, i (i)}<li>{e.detalle}</li>{/each}
			</ul>
		{/if}
		<div class="m-acc">
			{#if activa(cola)}
				<button type="button" class="btn btn-secondary" onclick={cancelarCola}>Frenar</button>
			{:else}
				<button type="button" class="btn btn-dark" onclick={cerrarCola}>Listo</button>
			{/if}
		</div>
	{:else}
		<div class="m-t">Traé tus productos de otra&nbsp;tienda</div>
		<div class="m-d">
			Pegá el link y traemos cada producto con fotos, precio, variantes, categoría y descripción. Si
			importás dos veces, no se&nbsp;duplican.
		</div>
		{#if buscando}
			<div class="m-c" role="status">Buscando tu tienda…</div>
		{:else}
			{#if tiendanube === false}
				<button type="button" class="btn btn-secondary btn-block m-tn" disabled
					>Tiendanube no está disponible</button
				>
			{:else}
				<a
					class="btn btn-secondary btn-block m-tn"
					href={resolve('/panel/importar/tiendanube')}
					data-sveltekit-reload>Conectar Tiendanube</a
				>
			{/if}
			{#if avisoTN}<div class="m-err" role="alert">{avisoTN}</div>{/if}
			<div class="m-d m-tn-d">Con Tiendanube traemos también tus clientes y&nbsp;pedidos.</div>
			<div class="m-c">O pegá el link de otra plataforma</div>
			<div class="m-seg">
				<Segmentado opciones={FUENTES} bind:value={fuente} label="Plataforma de tu tienda" />
			</div>
			<label class="m-c" for="tienda-imp">Link de tu tienda</label>
			<input
				id="tienda-imp"
				class="m-in"
				bind:value={tiendaImp}
				placeholder={EJEMPLO[fuente]}
				autocomplete="url"
				onkeydown={(e) => e.key === 'Enter' && tiendaImp.trim() && importarTienda()}
			/>
			{#if errorImp}<div class="m-err" role="alert">{errorImp}</div>{/if}
			<div class="m-acc">
				<button type="button" class="b-sec" onclick={() => (importando = false)}>Cancelar</button>
				<button
					type="button"
					class="b-pri"
					disabled={!tiendaImp.trim() || buscando}
					onclick={importarTienda}>Importar</button
				>
			</div>
		{/if}
	{/if}
</Modal>

<MotorLista {config} lista={data.lista} {abrir}>
	{#snippet extra()}
		<button type="button" class="secundario" onclick={() => archivo?.click()}
			><Icon icon={I.phFileXlsBold} size={15} />Importar Excel</button
		>
		<button type="button" class="secundario" onclick={abrirImportar}
			><Icon icon={I.phLinkBold} size={15} />Importar tienda</button
		>
	{/snippet}
	{#snippet antes()}
		<!-- La invitación a cargar con foto es para arrancar: con productos, «Nuevo producto» está arriba. -->
		{#if !data.lista.total}
			<div class="foto">
				<img src="/hornero/foto.webp" alt="" width="64" height="64" />
				<div class="f-txt">
					<div class="f-t">Cargá productos con sus fotos</div>
					<!-- Tincho no completa la ficha desde la foto (no hay nada que lo haga): se dice lo que hay. -->
					<div class="f-d">
						Subí las fotos desde el celular y completá nombre, precio y&nbsp;stock.
					</div>
				</div>
				<button
					type="button"
					onclick={() => {
						producto = null;
						editor = true;
					}}><Icon icon={I.phCameraFill} size={16} style="color:#b0441d" />Cargar producto</button
				>
			</div>
		{/if}
	{/snippet}
</MotorLista>

<EditorProducto
	bind:open={editor}
	{producto}
	tienda={tienda.nombre}
	{dominio}
	recargoML={data.recargoML}
	mayorista={data.mayorista}
	guardado={async (p, nuevo) => {
		await invalidateAll();
		ui.toast({
			k: 'ok',
			title: nuevo ? 'Producto creado' : 'Producto guardado',
			desc: `${p.nombre} · ${publicado(Object.values(p.canales ?? {}).filter(Boolean).length)}`
		});
	}}
	borrado={async (nombre) => {
		await invalidateAll();
		ui.toast({ k: 'ok', title: 'Producto borrado', desc: nombre });
	}}
/>

<Modal bind:open={cambiandoPrecio} titulo="Cambiar precio" ancho={420}>
	<form onsubmit={aplicarPrecio}>
		<div class="m-ico"><Icon icon={I.phPercentBold} size={24} /></div>
		<div class="m-t">
			Cambiar precio de {porcentaje.length}
			{porcentaje.length === 1 ? 'producto' : 'productos'}
		</div>
		<div class="m-d">Subí o bajá todo de una. Redondeamos a $ 10.</div>
		<div class="m-rango">
			<input type="range" min="-30" max="50" step="1" bind:value={pct} aria-label="Porcentaje" />
			<b>{pct > 0 ? '+' : ''}{pct}%</b>
		</div>
		<div class="m-acc">
			<button type="button" class="btn btn-secondary" onclick={() => (cambiandoPrecio = false)}
				>Cancelar</button
			>
			<button class="btn btn-primary" disabled={!pct}>Aplicar</button>
		</div>
	</form>
</Modal>

<style>
	.foto {
		position: relative;
		overflow: hidden;
		display: flex;
		align-items: center;
		gap: 16px;
		padding: 16px 18px;
		border-radius: 22px;
		background: linear-gradient(120deg, #2a140c, #4a2416);
		color: #fbf7f2;
		flex-wrap: wrap;
	}
	.foto img {
		position: relative;
		object-fit: contain;
	}
	.f-txt {
		position: relative;
		flex: 1 1 260px;
	}
	.f-t {
		font-size: 16px;
		font-weight: 600;
	}
	.f-d {
		font-size: 13.5px;
		color: #d9c8ba;
	}
	.foto button {
		position: relative;
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 11px 18px;
		border-radius: 999px;
		border: 0;
		background: #fbf7f2;
		color: #2a140c;
		font-size: 13.5px;
		font-weight: 600;
		white-space: nowrap;
	}
	.secundario {
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
	.m-ico {
		width: 48px;
		height: 48px;
		display: grid;
		place-items: center;
		border-radius: 15px;
		background: rgba(210, 88, 47, 0.12);
		color: #b0441d;
	}
	.m-t {
		margin-top: 16px;
		font-size: 21px;
		font-weight: 600;
		letter-spacing: -0.02em;
	}
	.m-d {
		margin-top: 6px;
		font-size: 14.5px;
		line-height: 1.5;
		color: var(--mu);
	}
	.m-rango {
		margin-top: 18px;
		display: flex;
		align-items: center;
		gap: 12px;
	}
	.m-rango input {
		flex: 1;
		accent-color: #d2582f;
	}
	.m-rango b {
		min-width: 52px;
		text-align: right;
		font-size: 18px;
	}
	.m-acc {
		margin-top: 20px;
		display: flex;
		gap: 8px;
		justify-content: flex-end;
	}
	.m-ico {
		width: 48px;
		height: 48px;
		border-radius: 16px;
		background: rgba(46, 42, 107, 0.1);
		display: grid;
		place-items: center;
	}
	.m-t {
		margin-top: 12px;
		font-size: 19px;
		font-weight: 600;
	}
	.m-d {
		margin-top: 6px;
		font-size: 14px;
		line-height: 1.55;
		color: var(--mu);
	}
	.m-c {
		display: block;
		margin-top: 14px;
		font-size: 13px;
		font-weight: 600;
	}
	.m-in {
		margin-top: 6px;
		width: 100%;
		padding: 12px 14px;
		border-radius: 14px;
		border: 1.5px solid var(--ln2);
		background: var(--sf2);
		font:
			500 15px 'Instrument Sans',
			sans-serif;
		color: var(--tx);
		outline: none;
	}
	.m-in:focus {
		border-color: var(--terracota);
	}
	.m-err {
		margin-top: 8px;
		font-size: 13px;
		color: #b83a2e;
	}
	.m-seg {
		margin-top: 6px;
	}
	.m-tn {
		margin-top: 16px;
	}
	.m-tn-d {
		font-size: 13px;
	}
	.m-etapas {
		margin: 12px 0 0;
		padding: 0;
		list-style: none;
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
	}
	.m-seg :global(.seg) {
		display: flex;
	}
	/* Las tres plataformas entran a 375 px: menos aire a los costados de cada una. */
	.m-seg :global(.seg button) {
		flex: auto;
		padding-inline: 6px;
	}
	.m-notas {
		margin: 10px 0 0;
		padding-left: 18px;
		font-size: 14px;
		line-height: 1.55;
		color: var(--mu);
	}
	.m-barra {
		margin-top: 8px;
		height: 8px;
		border-radius: 999px;
		background: var(--sf3);
		overflow: hidden;
	}
	.m-barra span {
		display: block;
		height: 100%;
		border-radius: inherit;
		background: var(--terracota);
		transition: width 0.4s;
	}
	.m-acc {
		margin-top: 16px;
		display: flex;
		gap: 8px;
	}
	.m-acc button {
		flex: 1;
		padding: 13px;
		border-radius: 999px;
		font:
			600 14px 'Instrument Sans',
			sans-serif;
		cursor: pointer;
	}
	.b-sec {
		border: 1px solid var(--ln2);
		background: transparent;
		color: var(--tx);
	}
	.b-pri {
		border: 0;
		background: var(--tx);
		color: var(--sf);
	}
	.b-pri:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}
</style>
