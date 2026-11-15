<script lang="ts">
	import { tick } from 'svelte';
	import { invalidate, invalidateAll } from '$app/navigation';
	import { resolve } from '$app/paths';
	import Icon from '#lib/components/Icon.svelte';
	import Modal from '#lib/components/Modal.svelte';
	import Segmentado from '#lib/components/Segmentado.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { subir as subirArchivo } from '#lib/panel/fotos.ts';
	import AjustesPagina from '#lib/panel/mi-tienda/AjustesPagina.svelte';
	import Biblioteca from '#lib/panel/mi-tienda/Biblioteca.svelte';
	import Estilo from '#lib/panel/mi-tienda/Estilo.svelte';
	import ListaBloques from '#lib/panel/mi-tienda/ListaBloques.svelte';
	import PanelBloque from '#lib/panel/mi-tienda/PanelBloque.svelte';
	import * as L from '#lib/panel/mi-tienda/bloques.ts';
	import type { Ayuda } from '#lib/panel/mi-tienda/campos.ts';
	import {
		conBloques,
		defDe,
		ORDEN_PAGINAS,
		PAGINAS,
		type PaginaId
	} from '#lib/panel/mi-tienda/catalogo.ts';
	import type { DisenoV2 } from '#lib/panel/mi-tienda/diseno.ts';
	import { paraGuardar } from '#lib/panel/mi-tienda/guardar.ts';
	import * as H from '#lib/panel/mi-tienda/historial.ts';
	import { aplicarPreset } from '#lib/panel/mi-tienda/partida.ts';
	import type { AlaVista } from '#lib/panel/mi-tienda/vista.ts';
	import { PLANTILLAS, plantilla as plantillaDe } from '#lib/plantillas.ts';
	import { escucharTiempoReal } from '#lib/tiempo-real/panel.svelte.ts';
	import { LARGO_DESCRIPCION, LARGO_TITULO, textosPorDefecto } from '#lib/tienda/compartir.ts';
	import type { TipoBloque } from '#lib/tienda/sistema/catalogo.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/**
	 * Mi tienda, el editor en bloques (docs/especificaciones/tienda-bloques.md › Editor): elegir la
	 * página, armar su lista de bloques, el panel de cada uno, el estilo y el punto de partida, con
	 * deshacer y la vista previa en vivo. Cada cambio confirmado se guarda en el borrador; publicar lo
	 * pasa a la tienda.
	 */
	let { data } = $props();
	const ui = getUi();

	// ── El diseño y su historial ──
	// Todo cambio arma un diseño nuevo (sin mutar): el anterior queda para deshacer.
	// svelte-ignore state_referenced_locally
	let diseno = $state.raw<DisenoV2>(data.diseno);
	/** El último diseño confirmado (lo que entra al historial y se guarda). */
	// svelte-ignore state_referenced_locally
	let confirmado = data.diseno;
	let hist = $state.raw(H.vacio<DisenoV2>());
	// svelte-ignore state_referenced_locally
	let sinPublicar = $state(data.sinPublicar);
	const plantillaId = $derived(diseno.plantilla || data.plantillaId);
	const elegida = $derived(plantillaDe(plantillaId));

	/** Cambiar el diseño. Sin confirmar sólo se ve en la vista previa (mientras se escribe). */
	function cambiar(d: DisenoV2, confirmar = true) {
		diseno = d;
		if (confirmar) confirmarCambios();
	}
	function confirmarCambios() {
		if (diseno === confirmado || JSON.stringify(diseno) === JSON.stringify(confirmado)) return;
		hist = H.registrar(hist, confirmado);
		confirmado = diseno;
		void guardar();
	}
	function deshacer() {
		confirmarCambios();
		const r = H.deshacer(hist, confirmado);
		if (!r) return;
		hist = r.h;
		diseno = confirmado = r.estado;
		if (abierto && !lista.some((b) => b.id === abierto)) abierto = null;
		void guardar();
	}
	function rehacer() {
		const r = H.rehacer(hist, confirmado);
		if (!r) return;
		hist = r.h;
		diseno = confirmado = r.estado;
		if (abierto && !lista.some((b) => b.id === abierto)) abierto = null;
		void guardar();
	}
	function teclas(e: KeyboardEvent) {
		if (!(e.metaKey || e.ctrlKey)) return;
		// En un campo de texto, deshacer es el del navegador (letra por letra).
		const t = e.target as HTMLElement;
		if (t.closest('input,textarea,select,[contenteditable]')) return;
		const k = e.key.toLowerCase();
		if (k === 'z' && !e.shiftKey) {
			e.preventDefault();
			deshacer();
		} else if ((k === 'z' && e.shiftKey) || k === 'y') {
			e.preventDefault();
			rehacer();
		}
	}

	// ── Guardar (en fila: el último en llegar tiene que ser el último hecho) ──
	let fila = Promise.resolve();
	let vuelta = $state(0);
	// svelte-ignore state_referenced_locally
	let logoDelPanel = data.diseno.logo ?? '';
	function guardar() {
		sinPublicar = true;
		const d = paraGuardar(diseno);
		const cuerpo = JSON.stringify(d);
		fila = fila.then(async () => {
			const r = await fetch('/panel/mi-tienda', {
				method: 'PUT',
				headers: { 'content-type': 'application/json' },
				body: cuerpo
			});
			if (r.ok) {
				vuelta++;
				// El logo nuevo se ve enseguida en la barra del panel (en vez de las iniciales).
				if ((d.logo ?? '') !== logoDelPanel) {
					logoDelPanel = d.logo ?? '';
					void invalidate('tienda:armazon');
				}
			} else
				ui.toast({
					k: 'err',
					title: 'No se guardó el cambio',
					desc:
						((await r.json().catch(() => null)) as { message?: string } | null)?.message ??
						'Probá de nuevo en un rato.'
				});
		});
		return fila;
	}

	// Si otra persona del equipo publica mientras esta pantalla está abierta (§44), se avisa: no se
	// recarga sola para no pisar lo que se está editando acá.
	let publiqueEn = 0;
	escucharTiempoReal((e) => {
		if (e.tipo !== 'mitienda.publicada' || Date.now() - publiqueEn < 10_000) return;
		ui.toast({
			k: 'info',
			title: 'Otra persona publicó la tienda',
			desc: 'Alguien del equipo publicó cambios del diseño.',
			action: 'Ver lo publicado',
			onAction: () => invalidateAll()
		});
	});
	async function publicar() {
		confirmarCambios();
		if (!sinPublicar) return;
		await fila;
		publiqueEn = Date.now();
		await fetch('/panel/mi-tienda/publicar', { method: 'POST' });
		sinPublicar = false;
		await invalidateAll();
		ui.toast({
			k: 'ok',
			title: 'Cambios publicados',
			desc: `${data.direccion} ya muestra la nueva versión`
		});
	}

	// ── Páginas, bloques y paneles ──
	let pestana = $state<'pagina' | 'estilo'>('pagina');
	let pagina = $state<PaginaId>('inicio');
	let abierto = $state<string | null>(null);
	const lista = $derived(L.bloquesDe(diseno, pagina));
	const bloqueAbierto = $derived(lista.find((b) => b.id === abierto) ?? null);
	const ayuda: Ayuda = $derived({
		cx: data.contexto,
		nombre: data.vista.tienda,
		lema: diseno.lema ?? '',
		subir: subirArchivo
	});
	const ponerLista = (l: L.Bloque[], confirmar = true) =>
		cambiar(L.conBloquesEn(diseno, pagina, l), confirmar);
	const nombreDe = (b: L.Bloque) => defDe(b.tipo)?.nombre ?? b.tipo;

	let columna: HTMLElement | undefined = $state();
	async function abrir(id: string | null) {
		abierto = id;
		pestana = 'pagina';
		modo = 'editar';
		enviar({ tipo: 'seleccionar', id });
		await tick();
		if (id && matchMedia('(max-width: 999px)').matches)
			columna?.scrollIntoView({ behavior: 'smooth', block: 'start' });
	}
	function elegirPagina(p: PaginaId) {
		pagina = p;
		abierto = null;
		enviar({ tipo: 'diseno', diseno, pagina: p });
	}

	let biblioteca = $state(false);
	let agregarEn = $state(0);
	const donde = $derived(
		agregarEn >= lista.length
			? 'al final de la página'
			: agregarEn === 0
				? 'arriba de todo'
				: `después de ${nombreDe(lista[agregarEn - 1])}`
	);
	function pedirAgregar(en: number) {
		agregarEn = en;
		biblioteca = true;
	}
	function agregar(tipo: TipoBloque) {
		const b = L.crear(tipo, L.idsDe(diseno));
		ponerLista(L.agregar(lista, b, agregarEn));
		void abrir(b.id);
	}
	function mover(de: number, a: number) {
		ponerLista(L.mover(lista, de, a));
	}
	function alternar(i: number) {
		ponerLista(L.alternar(lista, i));
	}
	function duplicar(i: number) {
		const l = L.duplicar(lista, i, L.idsDe(diseno));
		ponerLista(l);
		void abrir(l[i + 1].id);
	}
	function borrar(i: number) {
		const b = lista[i];
		ponerLista(L.borrar(lista, i));
		abierto = null;
		ui.undoable(`Borraste ${nombreDe(b).toLowerCase()}`, undefined, deshacer);
	}

	// ── Punto de partida ──
	let eligiendo = $state(false);
	function partida(id: string) {
		eligiendo = false;
		if (id === plantillaId) return;
		cambiar(aplicarPreset(diseno, id));
		abierto = null;
		ui.undoable(
			`Punto de partida: ${plantillaDe(id)?.nombre ?? id}`,
			'Conservamos lo que cargaste en los bloques.',
			deshacer
		);
	}

	// ── Vista previa en vivo ──
	let dispositivo = $state<'celular' | 'compu'>('celular');
	let modo = $state<'editar' | 'vista'>('editar');
	let marco: HTMLIFrameElement | undefined = $state();
	let vistaLista = $state(false);
	function enviar(m: AlaVista) {
		if (vistaLista) marco?.contentWindow?.postMessage(m, location.origin);
	}
	$effect(() => {
		// Cada cambio (también mientras se escribe) se redibuja sin recargar.
		if (vistaLista) enviar({ tipo: 'diseno', diseno, pagina });
	});
	function deLaVista(e: MessageEvent) {
		if (e.origin !== location.origin || !marco || e.source !== marco.contentWindow) return;
		const m = e.data as { tipo?: string; id?: unknown };
		if (m?.tipo === 'lista') {
			vistaLista = true;
			enviar({ tipo: 'diseno', diseno, pagina });
			if (abierto) enviar({ tipo: 'seleccionar', id: abierto });
		} else if (m?.tipo === 'bloque' && typeof m.id === 'string') {
			if (lista.some((b) => b.id === m.id)) void abrir(m.id as string);
		}
	}

	// ── Al compartir el link ──
	const sugeridos = $derived(
		textosPorDefecto({
			...data.hechos,
			titulo: data.hechos.titulo,
			lema: (diseno.lema ?? '').trim()
		})
	);
	const host = $derived(data.direccion.split('/')[0]);
	// svelte-ignore state_referenced_locally
	let imagenPropia = $state(!!data.diseno.compartirFoto);
	let cargandoImagen = $state(true);
	let subiendoCompartir = $state(false);
	const imagenPrevia = $derived(diseno.compartirFoto || `/panel/mi-tienda/compartir?v=${vuelta}`);
	$effect(() => {
		void imagenPrevia;
		cargandoImagen = true;
	});
	function elegirImagen(v: 'armada' | 'propia') {
		imagenPropia = v === 'propia';
		if (!imagenPropia && diseno.compartirFoto) cambiar({ ...diseno, compartirFoto: undefined });
	}
	async function subirCompartir(e: Event) {
		const input = e.currentTarget as HTMLInputElement;
		const archivo = input.files?.[0];
		input.value = '';
		if (!archivo) return;
		subiendoCompartir = true;
		try {
			cambiar({ ...diseno, compartirFoto: await subirArchivo(archivo) });
		} catch (err) {
			ui.toast({
				k: 'err',
				title: 'No pudimos subir la foto',
				desc: err instanceof Error ? err.message : 'Probá de nuevo en un rato.'
			});
		} finally {
			subiendoCompartir = false;
		}
	}
	const enlace = $derived(data.direccion.replace(/^[^/]+/, ''));
</script>

<svelte:window onkeydown={teclas} onmessage={deLaVista} />

<Seo title="Mi tienda · Panel" description="Editor de la tienda" path="/panel/mi-tienda" noindex />

<div class="pagina" class:en-vista={modo === 'vista'}>
	<div class="cab" use:anim={{ t: 'rise' }}>
		<div class="c-txt">
			<h1>Mi tienda</h1>
			<div class="c-s">
				Punto de partida {elegida?.nombre ?? data.plantilla} ·
				<span class="dir">{data.direccion}</span>
			</div>
		</div>
		<a class="cambiar" href={resolve('/panel/tarjeta-para-paquetes')}
			><Icon icon={I.phEnvelopeSimpleFill} size={14} />Tarjeta para paquetes</a
		>
		<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- la tienda pública, en otra pestaña -->
		<a class="cambiar" href={enlace} target="_blank" rel="noopener"
			><Icon icon={I.phStorefrontFill} size={14} />Ver mi tienda</a
		>
		<button type="button" class="cambiar" onclick={() => (eligiendo = true)}
			><Icon icon={I.phSwapBold} size={14} />Punto de partida</button
		>
		<button type="button" class="publicar" class:listo={!sinPublicar} onclick={publicar}
			>{sinPublicar ? 'Publicar cambios' : 'Todo publicado'}</button
		>
	</div>

	<div class="modo">
		<Segmentado
			label="Qué ver"
			opciones={[
				['editar', 'Editar'],
				['vista', 'Vista previa']
			]}
			bind:value={modo}
		/>
	</div>

	<div class="editor" use:anim={{ t: 'rise', delay: 60 }}>
		<div class="panelito columna" bind:this={columna}>
			<div class="herramientas">
				<Segmentado
					label="Qué editar"
					opciones={[
						['pagina', 'Páginas'],
						['estilo', 'Estilo']
					]}
					bind:value={pestana}
				/>
				<div class="deshacer">
					<button
						type="button"
						class="mini-btn icono"
						aria-label="Deshacer"
						title="Deshacer (⌘Z)"
						aria-keyshortcuts="Meta+Z Control+Z"
						disabled={!hist.pasado.length}
						onclick={deshacer}><Icon icon={I.phArrowUUpLeftBold} size={15} /></button
					>
					<button
						type="button"
						class="mini-btn icono"
						aria-label="Rehacer"
						title="Rehacer (⇧⌘Z)"
						aria-keyshortcuts="Meta+Shift+Z Control+Shift+Z"
						disabled={!hist.futuro.length}
						onclick={rehacer}><Icon icon={I.phArrowUUpRightBold} size={15} /></button
					>
				</div>
			</div>

			{#if pestana === 'pagina'}
				{#if bloqueAbierto}
					{@const i = lista.indexOf(bloqueAbierto)}
					<PanelBloque
						bloque={bloqueAbierto}
						{ayuda}
						puedeDuplicar={!L.porQueNo(lista, bloqueAbierto.tipo)}
						onbloque={(b, c) =>
							ponerLista(
								lista.map((x) => (x.id === b.id ? b : x)),
								c
							)}
						onvolver={() => abrir(null)}
						onduplicar={() => duplicar(i)}
						onalternar={() => alternar(i)}
						onborrar={() => borrar(i)}
					/>
				{:else}
					<div class="paginas" role="group" aria-label="Página">
						{#each ORDEN_PAGINAS as p (p)}
							<button
								type="button"
								aria-pressed={pagina === p}
								class="pag"
								onclick={() => elegirPagina(p)}
								>{p === 'producto' ? 'Ficha' : PAGINAS[p].nombre}</button
							>
						{/each}
					</div>
					{#if pagina === 'catalogo' || pagina === 'producto' || pagina === 'carrito'}
						<div class="p-t">Ajustes de {PAGINAS[pagina].nombre.toLowerCase()}</div>
						<AjustesPagina {pagina} {diseno} {ayuda} oncambiar={cambiar} />
					{/if}
					{#if conBloques(pagina)}
						<div class="p-t">
							{pagina === 'inicio'
								? 'Bloques de la portada'
								: pagina === 'producto'
									? 'Bloques abajo de la ficha'
									: 'Bloques'}
						</div>
						{#if !lista.length}
							<p class="c-nota">
								{pagina === 'nosotros'
									? 'Contá quién está detrás: sumá un texto con imagen, el local o las preguntas frecuentes.'
									: 'Sumá bloques para contar más debajo de cada producto.'}
							</p>
						{/if}
						<ListaBloques
							{lista}
							seleccionado={abierto}
							cx={data.contexto}
							onabrir={abrir}
							onmover={mover}
							onalternar={alternar}
							onagregar={pedirAgregar}
						/>
					{/if}
				{/if}
			{:else}
				<Estilo {diseno} plantilla={plantillaId} {ayuda} oncambiar={cambiar} />
			{/if}
		</div>

		<div class="derecha">
			<div class="navegador">
				<div class="barra">
					<span style:background="#E0A595"></span><span style:background="#E7CB8F"></span><span
						style:background="#A9CBA8"
					></span>
					<span class="url">{data.direccion}</span>
				</div>
				<div class="vista" class:compu={dispositivo === 'compu'}>
					<iframe
						bind:this={marco}
						title="Vista previa de tu tienda"
						src="/panel/mi-tienda/vista?pagina=inicio&marcos=1"
					></iframe>
				</div>
			</div>
			<div class="dispositivo">
				<Segmentado
					label="Ver como"
					opciones={[
						['celular', 'Celular'],
						['compu', 'Compu']
					]}
					value={dispositivo}
					onchange={(v) => (dispositivo = v)}
				/>
			</div>
			<p class="c-nota centro">Tocá un bloque en la vista para editarlo.</p>
		</div>
	</div>

	<section class="panelito compartir" aria-labelledby="t-compartir">
		<div class="p-t" id="t-compartir">Al compartir el link</div>
		<div class="c-grid">
			<div class="c-campos">
				<label class="campo"
					><span>Título</span><input
						class="input"
						maxlength={LARGO_TITULO}
						placeholder={sugeridos.titulo}
						value={diseno.compartirTitulo ?? ''}
						oninput={(e) => cambiar({ ...diseno, compartirTitulo: e.currentTarget.value }, false)}
						onchange={confirmarCambios}
					/></label
				>
				<label class="campo"
					><span>Descripción</span><textarea
						class="textarea"
						rows="3"
						maxlength={LARGO_DESCRIPCION}
						placeholder={sugeridos.descripcion}
						value={diseno.compartirDescripcion ?? ''}
						oninput={(e) =>
							cambiar({ ...diseno, compartirDescripcion: e.currentTarget.value }, false)}
						onchange={confirmarCambios}></textarea></label
				>
				<div class="campo">
					<span>Imagen</span>
					<Segmentado
						label="Imagen al compartir"
						opciones={[
							['armada', 'La de tu plantilla'],
							['propia', 'Una foto tuya']
						]}
						value={imagenPropia ? 'propia' : 'armada'}
						onchange={elegirImagen}
					/>
				</div>
				{#if imagenPropia}
					<div class="foto-fila">
						<div class="miniatura">
							{#if diseno.compartirFoto}<img src={diseno.compartirFoto} alt="" />{:else}<Icon
									icon={I.phImageFill}
									size={18}
								/>{/if}
						</div>
						<div class="f-txt">
							<b>Foto para compartir</b>
							<span>{diseno.compartirFoto ? 'Subida' : 'Se recorta a 1200 × 630'}</span>
						</div>
						<label class="mini-btn" class:ocupado={subiendoCompartir}>
							{subiendoCompartir ? 'Subiendo…' : diseno.compartirFoto ? 'Cambiar' : 'Subir'}
							<input
								type="file"
								accept="image/jpeg,image/png,image/webp"
								class="sr-only"
								disabled={subiendoCompartir}
								onchange={subirCompartir}
							/>
						</label>
						{#if diseno.compartirFoto}<button
								type="button"
								class="mini-btn icono"
								aria-label="Quitar la foto para compartir"
								onclick={() => cambiar({ ...diseno, compartirFoto: undefined })}
								><Icon icon={I.phTrash} size={15} /></button
							>{/if}
					</div>
				{/if}
				<p class="c-nota">
					Vacíos, van los sugeridos. Se aplica al publicar; en un chat donde ya estaba el link,
					WhatsApp puede seguir mostrando la vista&nbsp;anterior.
				</p>
			</div>
			<div class="chat-wa">
				<div class="wa-t">Así se ve en WhatsApp</div>
				<div class="burbuja">
					<div class="tarjeta-wa">
						<div class="t-img" class:cargando={cargandoImagen}>
							<img
								src={imagenPrevia}
								alt="Imagen al compartir {data.vista.tienda}"
								onload={() => (cargandoImagen = false)}
								onerror={() => (cargandoImagen = false)}
							/>
						</div>
						<div class="t-txt">
							<b>{diseno.compartirTitulo?.trim() || sugeridos.titulo}</b>
							<span class="t-desc"
								>{diseno.compartirDescripcion?.trim() || sugeridos.descripcion}</span
							>
							<span class="t-host">{host}</span>
						</div>
					</div>
					<span class="b-url">https://{data.direccion}</span>
				</div>
			</div>
		</div>
	</section>
</div>

<Biblioteca bind:open={biblioteca} {lista} {donde} onelegir={agregar} />

<Modal bind:open={eligiendo} titulo="Elegí el punto de partida" ancho={860}>
	<div class="m-cab">
		<h2>Elegí el punto de partida</h2>
		<p>
			Cambia la portada, la dirección y los colores. Conservamos lo que cargaste en los bloques y lo
			podés deshacer. Se ve en tu tienda cuando tocás «Publicar&nbsp;cambios».
		</p>
	</div>
	<div class="plantillas">
		{#each PLANTILLAS as p (p.id)}
			<button
				type="button"
				class="tpl"
				aria-pressed={p.id === plantillaId}
				onclick={() => partida(p.id)}
			>
				<span class="tpl-foto" style:background={p.bg}
					><img src={p.img} alt="" loading="lazy" /></span
				>
				<span class="tpl-n">{p.nombre}</span>
				<span class="tpl-r">{p.rubro}</span>
			</button>
		{/each}
	</div>
</Modal>

<style>
	.m-cab h2 {
		margin: 0;
		font-size: 22px;
		letter-spacing: -0.02em;
	}
	.m-cab p {
		margin: 4px 0 16px;
		color: var(--mu);
		font-size: 14px;
		text-wrap: pretty;
	}
	.plantillas {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
		gap: 12px;
	}
	.tpl {
		display: flex;
		flex-direction: column;
		gap: 2px;
		padding: 8px;
		border-radius: 16px;
		border: 2px solid var(--ln);
		background: var(--sf);
		color: var(--tx);
		text-align: left;
		font: inherit;
		cursor: pointer;
	}
	.tpl[aria-pressed='true'] {
		border-color: var(--terracota);
	}
	.tpl-foto {
		display: block;
		aspect-ratio: 4 / 3;
		border-radius: 10px;
		overflow: hidden;
		margin-bottom: 6px;
	}
	.tpl-foto img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	.tpl-n {
		font-weight: 600;
		font-size: 14px;
	}
	.tpl-r {
		font-size: 12.5px;
		color: var(--mu);
	}
	.pagina {
		display: flex;
		flex-direction: column;
		gap: 18px;
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
	/* La dirección baja entera al renglón siguiente; sólo se corta si no entra ni sola. */
	.dir {
		display: inline-block;
		max-width: 100%;
		overflow-wrap: anywhere;
	}
	.cambiar {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 10px 16px;
		min-height: 40px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: var(--sf);
		color: var(--tx);
		font: inherit;
		font-size: 13.5px;
		font-weight: 600;
		box-shadow: var(--sh);
		white-space: nowrap;
		text-decoration: none;
		cursor: pointer;
	}
	.publicar {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 11px 20px;
		min-height: 44px;
		border-radius: 999px;
		border: 0;
		background: linear-gradient(180deg, #c24e25, #b0441d);
		color: #fff;
		font:
			600 14px 'Instrument Sans',
			sans-serif;
		white-space: nowrap;
		cursor: pointer;
	}
	.publicar.listo {
		background: #c9b6a6;
	}
	.modo {
		display: none;
	}
	.editor {
		display: grid;
		grid-template-columns: 340px minmax(0, 1fr);
		gap: 16px;
		align-items: start;
	}
	.panelito {
		padding: 14px;
		border-radius: 22px;
		background: var(--sf);
		box-shadow: var(--shd);
		display: flex;
		flex-direction: column;
		gap: 6px;
		min-width: 0;
	}
	.columna {
		scroll-margin-top: 80px;
	}
	.herramientas {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 8px;
		padding-bottom: 8px;
		margin-bottom: 2px;
		border-bottom: 1px solid var(--ln);
	}
	.deshacer {
		display: flex;
		gap: 4px;
	}
	.paginas {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
		padding: 2px 0 6px;
	}
	.pag {
		min-height: 36px;
		padding: 6px 12px;
		border: 1px solid var(--ln2);
		border-radius: 999px;
		background: var(--sf);
		color: var(--tx);
		font: inherit;
		font-size: 13px;
		font-weight: 600;
		cursor: pointer;
	}
	.pag[aria-pressed='true'] {
		background: var(--tx);
		border-color: var(--tx);
		color: var(--sf);
	}
	.p-t {
		font-size: 14px;
		font-weight: 600;
		padding: 8px 4px 4px;
	}
	.c-nota {
		margin: 0;
		padding: 0 4px;
		font-size: 12px;
		color: var(--mu);
		text-wrap: pretty;
	}
	.c-nota.centro {
		text-align: center;
		margin-top: -8px;
	}
	.mini-btn {
		display: inline-grid;
		place-items: center;
		min-height: 36px;
		padding: 0 12px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: var(--sf);
		color: var(--tx);
		font:
			600 13px 'Instrument Sans',
			sans-serif;
		cursor: pointer;
		white-space: nowrap;
	}
	.mini-btn.icono {
		width: 36px;
		padding: 0;
	}
	.mini-btn:disabled {
		opacity: 0.35;
		cursor: default;
	}
	.mini-btn.ocupado {
		opacity: 0.6;
	}
	.navegador {
		border-radius: 20px;
		overflow: hidden;
		background: #fff;
		box-shadow:
			0 30px 60px -30px rgba(42, 20, 12, 0.55),
			0 0 0 1px var(--ln);
	}
	.barra {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 9px 12px;
		background: #f1e9df;
	}
	.barra > span:not(.url) {
		width: 9px;
		height: 9px;
		border-radius: 50%;
	}
	.url {
		flex: 1;
		min-width: 0;
		margin: 0 30px;
		padding: 3px 10px;
		border-radius: 999px;
		background: #fff;
		font-size: 11.5px;
		color: #5c4536;
		text-align: center;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.derecha {
		display: flex;
		flex-direction: column;
		gap: 16px;
		min-width: 0;
		position: sticky;
		top: 16px;
	}
	/* La vista previa: la tienda de verdad con el borrador. */
	.vista {
		height: min(720px, 72vh);
		display: flex;
		justify-content: center;
		background: var(--sf2, #f4efe9);
	}
	.vista iframe {
		width: 390px;
		max-width: 100%;
		height: 100%;
		border: 0;
		background: #fff;
	}
	.vista.compu iframe {
		width: 100%;
	}
	.dispositivo {
		display: flex;
		justify-content: center;
		margin-top: -4px;
	}
	.foto-fila {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 8px 10px;
		border-radius: 14px;
		border: 1px solid var(--ln);
	}
	.miniatura {
		width: 44px;
		height: 44px;
		flex: none;
		border-radius: 10px;
		overflow: hidden;
		display: grid;
		place-items: center;
		background: var(--sf2, rgba(0, 0, 0, 0.05));
		color: var(--mu);
	}
	.miniatura img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	.f-txt {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
		font-size: 13.5px;
	}
	.f-txt span {
		font-size: 12px;
		color: var(--mu);
	}
	.campo {
		display: flex;
		flex-direction: column;
		gap: 4px;
		padding: 4px 4px 0;
		font-size: 13px;
		font-weight: 600;
	}
	.c-grid {
		display: grid;
		grid-template-columns: minmax(0, 1fr) minmax(0, 340px);
		gap: 16px;
		align-items: start;
	}
	.c-campos {
		display: flex;
		flex-direction: column;
		gap: 10px;
		min-width: 0;
	}
	.c-campos .textarea {
		resize: vertical;
		font: inherit;
		font-size: 15px;
	}
	.chat-wa {
		min-width: 0;
		padding: 14px;
		border-radius: 18px;
		background: #efe7dc;
	}
	.wa-t {
		font-size: 12px;
		font-weight: 600;
		color: #7a6252;
		margin-bottom: 8px;
	}
	.burbuja {
		margin-left: auto;
		max-width: min(330px, 100%);
		padding: 4px 4px 8px;
		border-radius: 14px 14px 4px 14px;
		background: #d8f3df;
		color: #1f2a20;
		font-size: 13.5px;
		line-height: 1.45;
	}
	.b-url {
		display: block;
		padding: 6px 8px 0;
		color: #1f6f4a;
		font-weight: 600;
		overflow-wrap: anywhere;
	}
	.tarjeta-wa {
		border-radius: 10px;
		overflow: hidden;
		background: rgba(0, 0, 0, 0.05);
	}
	.t-img {
		aspect-ratio: 1200 / 630;
		background: rgba(0, 0, 0, 0.06);
		transition: opacity 0.2s;
	}
	.t-img.cargando {
		opacity: 0.55;
	}
	.t-img img {
		width: 100%;
		height: 100%;
		object-fit: cover;
		display: block;
	}
	.t-txt {
		display: flex;
		flex-direction: column;
		gap: 2px;
		padding: 8px 10px 9px;
	}
	.t-txt b {
		font-size: 13.5px;
		line-height: 1.3;
		overflow-wrap: anywhere;
	}
	.t-desc {
		font-size: 12.5px;
		line-height: 1.35;
		color: #4b5a4e;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}
	.t-host {
		font-size: 12px;
		color: #6b786d;
	}
	@media (max-width: 999px) {
		/* En el celular y la tablet: editar o mirar la vista, una cosa por vez. */
		.modo {
			display: flex;
			justify-content: center;
			position: sticky;
			top: 8px;
			z-index: 5;
		}
		.editor {
			grid-template-columns: minmax(0, 1fr);
		}
		.derecha {
			position: static;
		}
		.pagina:not(.en-vista) .navegador,
		.pagina:not(.en-vista) .dispositivo,
		.pagina:not(.en-vista) .c-nota.centro {
			display: none;
		}
		.pagina.en-vista .columna,
		.pagina.en-vista .compartir {
			display: none;
		}
	}
	@media (max-width: 760px) {
		.c-grid {
			grid-template-columns: minmax(0, 1fr);
		}
		.burbuja {
			max-width: 100%;
		}
		.mini-btn,
		.pag,
		.cambiar {
			min-width: 44px;
			min-height: 44px;
		}
		.cambiar {
			flex: 1 1 auto;
			justify-content: center;
		}
		.vista {
			height: calc(100dvh - 220px);
		}
	}
</style>
