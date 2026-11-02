<script lang="ts">
	import { invalidate, invalidateAll } from '$app/navigation';
	import { resolve } from '$app/paths';
	import Icon from '#lib/components/Icon.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { unsplash } from '#lib/landing/formato.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { pesos } from '#lib/panel/formato.ts';
	import { getUi } from '#lib/ui.svelte.ts';
	import { escucharTiempoReal } from '#lib/tiempo-real/panel.svelte.ts';
	import Modal from '#lib/components/Modal.svelte';
	import Segmentado from '#lib/components/Segmentado.svelte';
	import { subir as subirArchivo } from '#lib/panel/fotos.ts';
	import { PLANTILLAS, plantilla } from '#lib/plantillas.ts';
	import { acentos } from '#lib/tienda/acentos.ts';
	import { LARGO_DESCRIPCION, LARGO_TITULO, textosPorDefecto } from '#lib/tienda/compartir.ts';

	/** Mi tienda (design/PanelOperacion.dc.html, sStore): secciones, colores y vista previa en vivo. */
	let { data } = $props();
	const ui = getUi();
	type Seccion = (typeof data.diseno.secciones)[number];

	const SECCIONES = {
		hero: [I.phImageFill, 'Portada'],
		cats: [I.phTagFill, 'Categorías'],
		prods: [I.phStorefrontFill, 'Productos destacados'],
		revs: [I.phStarFill, 'Reseñas'],
		ig: [I.logosInstagramIcon, 'Instagram'],
		wa: [I.phWhatsappLogoFill, 'Botón de WhatsApp']
	} as const;
	// Lo que se edita es local y se guarda en cada cambio; al recargar vuelve lo guardado.
	let secciones = $derived<Seccion[]>(data.diseno.secciones.map((s) => ({ ...s })));
	let paleta = $derived(data.diseno.paleta);
	let plantillaId = $derived(data.plantillaId);
	let titulo = $derived(data.diseno.titulo ?? '');
	let lema = $derived(data.diseno.lema ?? '');
	let portada = $derived(data.diseno.portada ?? '');
	let logo = $derived(data.diseno.logo ?? '');
	// Al compartir: vacíos, lo sugerido y la imagen armada con la plantilla.
	let compartirTitulo = $derived(data.diseno.compartirTitulo ?? '');
	let compartirDescripcion = $derived(data.diseno.compartirDescripcion ?? '');
	let compartirFoto = $derived(data.diseno.compartirFoto ?? '');
	let imagenPropia = $derived(!!data.diseno.compartirFoto);
	// Lo sugerido sigue a lo que se edita (título de la portada y lema) con los datos de la tienda.
	const sugeridos = $derived(
		textosPorDefecto({
			...data.hechos,
			titulo: titulo.trim() || (data.diseno.titulo ? data.vista.tienda : data.hechos.titulo),
			lema: lema.trim()
		})
	);
	const host = $derived(data.direccion.split('/')[0]);
	// La vista previa de la imagen se vuelve a pedir cada vez que se guarda un cambio.
	let vuelta = $state(0);
	let cargandoImagen = $state(true);
	let sinPublicar = $derived(data.sinPublicar);
	const elegida = $derived(plantilla(plantillaId));
	// Color de acento: el de la plantilla o una de tres alternativas (lib/tienda/acentos.ts).
	const opciones = $derived(acentos(plantillaId));
	const pal = $derived({
		bg: elegida?.bg ?? '#F3E9DC',
		ink: elegida?.tx ?? '#3B2416',
		acc: (opciones[paleta] ?? opciones[0]).color
	});
	// La tienda pública, para abrirla (la dirección viene sin protocolo).
	const enlace = $derived(data.direccion.replace(/^[^/]+/, ''));

	// El logo que muestra la barra del panel: al guardarse otro, se recarga el armazón.
	// svelte-ignore state_referenced_locally -- el de la carga; después lo lleva el guardado
	let logoDelPanel = data.diseno.logo ?? '';
	// Los guardados van en fila: si dos salen juntos, el último en llegar tiene que ser el último hecho.
	let fila = Promise.resolve();
	function guardar() {
		sinPublicar = true;
		const cuerpo = JSON.stringify({
			secciones,
			paleta,
			plantilla: plantillaId,
			titulo,
			lema,
			portada,
			logo,
			compartirTitulo,
			compartirDescripcion,
			compartirFoto
		});
		fila = fila.then(async () => {
			const r = await fetch('/panel/mi-tienda', {
				method: 'PUT',
				headers: { 'content-type': 'application/json' },
				body: cuerpo
			});
			if (r.ok) {
				vuelta++;
				// El logo nuevo se ve enseguida en la barra del panel (en vez de las iniciales).
				if (logo !== logoDelPanel) {
					logoDelPanel = logo;
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

	let eligiendo = $state(false);
	function elegir(id: string) {
		eligiendo = false;
		if (id === plantillaId) return;
		plantillaId = id;
		void guardar();
	}

	// Logo y foto de portada: se suben como las fotos de los productos.
	let subiendo = $state<'' | 'logo' | 'portada' | 'compartir'>('');
	async function subirFoto(e: Event, cual: 'logo' | 'portada' | 'compartir') {
		const input = e.currentTarget as HTMLInputElement;
		const archivo = input.files?.[0];
		input.value = '';
		if (!archivo) return;
		subiendo = cual;
		try {
			const url = await subirArchivo(archivo);
			if (cual === 'logo') logo = url;
			else if (cual === 'compartir') compartirFoto = url;
			else portada = url;
			await guardar();
		} catch (err) {
			ui.toast({
				k: 'err',
				title: 'No pudimos subir la foto',
				desc: err instanceof Error ? err.message : 'Probá de nuevo en un rato.'
			});
		} finally {
			subiendo = '';
		}
	}
	function quitar(cual: 'logo' | 'portada' | 'compartir') {
		if (cual === 'logo') logo = '';
		else if (cual === 'compartir') compartirFoto = '';
		else portada = '';
		void guardar();
	}
	const fotoPortada = $derived(portada || data.vista.portada.foto);
	function subir(i: number) {
		if (!i) return;
		const a = [...secciones];
		[a[i - 1], a[i]] = [a[i], a[i - 1]];
		secciones = a;
		void guardar();
	}
	function alternar(i: number) {
		secciones = secciones.map((s, k) => (k === i ? { ...s, visible: !s.visible } : s));
		void guardar();
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
		if (!sinPublicar) return;
		await fila;
		publiqueEn = Date.now();
		await fetch('/panel/mi-tienda/publicar', { method: 'POST' });
		await invalidateAll();
		ui.toast({
			k: 'ok',
			title: 'Cambios publicados',
			desc: `${data.direccion} ya muestra la nueva versión`
		});
	}
	const visibles = $derived(secciones.filter((s) => s.visible));
	function elegirImagen(v: 'armada' | 'propia') {
		imagenPropia = v === 'propia';
		// Volver a la armada deja de usar la foto propia (si había una).
		if (!imagenPropia && compartirFoto) {
			compartirFoto = '';
			void guardar();
		}
	}
	const imagenPrevia = $derived(compartirFoto || `/panel/mi-tienda/compartir?v=${vuelta}`);
	$effect(() => {
		void imagenPrevia;
		cargandoImagen = true;
	});
</script>

<Seo title="Mi tienda · Panel" description="Editor de la tienda" path="/panel/mi-tienda" noindex />

<div class="pagina">
	<div class="cab" use:anim={{ t: 'rise' }}>
		<div class="c-txt">
			<h1>Mi tienda</h1>
			<div class="c-s">
				Plantilla {elegida?.nombre ?? data.plantilla} · <span class="dir">{data.direccion}</span>
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
			><Icon icon={I.phSwapBold} size={14} />Cambiar plantilla</button
		>
		<button type="button" class="publicar" class:listo={!sinPublicar} onclick={publicar}
			>{sinPublicar ? 'Publicar cambios' : 'Todo publicado'}</button
		>
	</div>

	<div class="editor" use:anim={{ t: 'rise', delay: 60 }}>
		<div class="panelito">
			<div class="p-t">Marca y portada</div>
			{#each [['logo', 'Logo', logo], ['portada', 'Foto de portada', portada]] as const as [cual, t, url] (cual)}
				<div class="foto-fila">
					<div class="miniatura" class:logo={cual === 'logo'}>
						{#if url}<img src={url} alt="" />{:else}<Icon
								icon={cual === 'logo' ? I.phStorefrontFill : I.phImageFill}
								size={18}
							/>{/if}
					</div>
					<div class="f-txt">
						<b>{t}</b>
						<span
							>{url
								? 'Subido'
								: cual === 'logo'
									? 'Sin logo: va el nombre en letras'
									: 'Sin foto: va la de un producto'}</span
						>
					</div>
					<label class="mini-btn" class:ocupado={subiendo === cual}>
						{subiendo === cual ? 'Subiendo…' : url ? 'Cambiar' : 'Subir'}
						<input
							type="file"
							accept="image/jpeg,image/png,image/webp"
							class="sr-only"
							disabled={!!subiendo}
							onchange={(e) => subirFoto(e, cual)}
						/>
					</label>
					{#if url}<button
							type="button"
							class="mini-btn icono"
							aria-label="Quitar {t.toLowerCase()}"
							onclick={() => quitar(cual)}><Icon icon={I.phTrash} size={15} /></button
						>{/if}
				</div>
			{/each}
			<label class="campo"
				><span>Título de la portada</span><input
					class="input"
					maxlength="80"
					placeholder={data.vista.tienda}
					bind:value={titulo}
					onchange={guardar}
				/></label
			>
			<label class="campo"
				><span>Lema</span><input
					class="input"
					maxlength="120"
					placeholder="Ej.: Hecho a mano en Rosario"
					bind:value={lema}
					onchange={guardar}
				/></label
			>
			<div class="p-t colores">Secciones</div>
			{#each secciones as s, i (s.id)}
				{@const [icono, t] = SECCIONES[s.id]}
				<div class="sec">
					<Icon icon={I.phDotsSixVerticalBold} size={15} style="color:var(--fa)" />
					<Icon icon={icono} size={16} style="color:#b0441d" />
					<span class="s-t" class:oculta={!s.visible}>{t}</span>
					<button type="button" aria-label="Subir {t}" disabled={i === 0} onclick={() => subir(i)}
						><Icon icon={I.phCaretUpBold} size={12} /></button
					>
					<button
						type="button"
						aria-label="{s.visible ? 'Ocultar' : 'Mostrar'} {t}"
						onclick={() => alternar(i)}
						><Icon icon={s.visible ? I.phEye : I.phEyeSlash} size={14} /></button
					>
				</div>
			{/each}
			<div class="p-t colores">Color de acento</div>
			<div class="paletas" role="group" aria-label="Color de acento">
				{#each opciones as o, i (o.color)}
					<button
						type="button"
						aria-label={i ? 'Color ' + (i + 1) : 'El de la plantilla'}
						title={i ? undefined : 'El de la plantilla'}
						aria-pressed={paleta === i}
						style:background="linear-gradient(135deg,{pal.bg} 50%,{o.color} 50%)"
						onclick={() => {
							paleta = i;
							void guardar();
						}}
					></button>
				{/each}
			</div>
		</div>

		<div class="derecha">
			<div class="navegador">
				<div class="barra">
					<span style:background="#E0A595"></span><span style:background="#E7CB8F"></span><span
						style:background="#A9CBA8"
					></span>
					<span class="url">{data.direccion}</span>
				</div>
				<div class="sitio" style:background={pal.bg} style:color={pal.ink}>
					<div class="s-cab">
						{#if logo}<img class="s-logo" src={logo} alt={data.vista.tienda} />{:else}<b
								>{data.vista.tienda}</b
							>{/if}<span class="esp"></span><span class="s-cats"
							>{data.vista.categorias.slice(0, 3).join(' · ')}</span
						><Icon icon={I.phHandbag} size={17} />
					</div>
					{#each visibles as s (s.id)}
						{#if s.id === 'hero'}
							<div class="hero">
								{#if fotoPortada}<img
										src={fotoPortada.startsWith('http') ? fotoPortada : unsplash(fotoPortada, 1000)}
										alt=""
									/>{/if}
								<div class="sombra"></div>
								<div class="lema">
									{titulo || data.vista.tienda}
									{#if lema}<small>{lema}</small>{/if}
								</div>
							</div>
						{:else if s.id === 'prods'}
							<div class="prods">
								{#each data.vista.productos as p, i (i)}
									<div>
										<div class="foto">
											{#if p.foto}<img src={unsplash(p.foto, 300)} alt="" />{/if}
										</div>
										<div class="pn">{p.nombre}</div>
										<div class="pp" style:color={pal.acc}>{pesos(p.precio)}</div>
									</div>
								{/each}
							</div>
						{:else if s.id === 'cats'}
							<div class="cats">
								{#each data.vista.categorias as c (c)}<span
										style:background={pal.acc}
										style:color={opciones[0].sobre}>{c}</span
									>{/each}
							</div>
						{:else if s.id === 'revs'}
							{#if data.vista.resena}
								<div class="revs">
									<span class="estrellas">★★★★★</span>
									{data.vista.resena.promedio.toLocaleString('es-AR')} · “{data.vista.resena.texto}”
								</div>
							{/if}
						{:else if s.id === 'ig'}
							<div class="ig">
								{#each data.vista.fotos as f, i (i)}<img src={unsplash(f, 200)} alt="" />{/each}
							</div>
						{:else}
							<div class="wa">
								<Icon icon={I.phWhatsappLogoFill} size={16} />¿Dudas? Escribinos por WhatsApp
							</div>
						{/if}
					{/each}
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
								bind:value={compartirTitulo}
								onchange={guardar}
							/></label
						>
						<label class="campo"
							><span>Descripción</span><textarea
								class="textarea"
								rows="3"
								maxlength={LARGO_DESCRIPCION}
								placeholder={sugeridos.descripcion}
								bind:value={compartirDescripcion}
								onchange={guardar}></textarea></label
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
									{#if compartirFoto}<img src={compartirFoto} alt="" />{:else}<Icon
											icon={I.phImageFill}
											size={18}
										/>{/if}
								</div>
								<div class="f-txt">
									<b>Foto para compartir</b>
									<span>{compartirFoto ? 'Subida' : 'Se recorta a 1200 × 630'}</span>
								</div>
								<label class="mini-btn" class:ocupado={subiendo === 'compartir'}>
									{subiendo === 'compartir' ? 'Subiendo…' : compartirFoto ? 'Cambiar' : 'Subir'}
									<input
										type="file"
										accept="image/jpeg,image/png,image/webp"
										class="sr-only"
										disabled={!!subiendo}
										onchange={(e) => subirFoto(e, 'compartir')}
									/>
								</label>
								{#if compartirFoto}<button
										type="button"
										class="mini-btn icono"
										aria-label="Quitar la foto para compartir"
										onclick={() => quitar('compartir')}><Icon icon={I.phTrash} size={15} /></button
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
									<b>{compartirTitulo.trim() || sugeridos.titulo}</b>
									<span class="t-desc">{compartirDescripcion.trim() || sugeridos.descripcion}</span>
									<span class="t-host">{host}</span>
								</div>
							</div>
							<span class="b-url">https://{data.direccion}</span>
						</div>
					</div>
				</div>
			</section>
		</div>
	</div>
</div>

<Modal bind:open={eligiendo} titulo="Elegí la plantilla" ancho={860}>
	<div class="m-cab">
		<h2>Elegí la plantilla</h2>
		<p>Se aplica a tu tienda cuando tocás «Publicar cambios». Tus productos no cambian.</p>
	</div>
	<div class="plantillas">
		{#each PLANTILLAS as p (p.id)}
			<button
				type="button"
				class="tpl"
				aria-pressed={p.id === plantillaId}
				onclick={() => elegir(p.id)}
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
		border-color: #d2582f;
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
	.miniatura.logo img {
		object-fit: contain;
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
	.mini-btn.ocupado {
		opacity: 0.6;
	}
	.campo {
		display: flex;
		flex-direction: column;
		gap: 4px;
		padding: 4px 4px 0;
		font-size: 13px;
		font-weight: 600;
	}
	.s-logo {
		height: 26px;
		width: auto;
		max-width: 140px;
		object-fit: contain;
	}
	.lema small {
		display: block;
		margin-top: 6px;
		font-size: 13px;
		font-weight: 500;
		opacity: 0.9;
	}
	@media (max-width: 760px) {
		.sec button,
		.mini-btn {
			min-width: 44px;
			min-height: 44px;
		}
		.paletas button {
			min-width: 44px;
			min-height: 44px;
		}
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
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: var(--sf);
		color: var(--tx);
		font-size: 13.5px;
		font-weight: 600;
		box-shadow: var(--sh);
		white-space: nowrap;
		text-decoration: none;
	}
	.publicar {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 11px 20px;
		border-radius: 999px;
		border: 0;
		background: linear-gradient(180deg, #c24e25, #b0441d);
		color: #fff;
		font:
			600 14px 'Instrument Sans',
			sans-serif;
		white-space: nowrap;
	}
	.publicar.listo {
		background: #c9b6a6;
	}
	.editor {
		display: grid;
		grid-template-columns: 300px minmax(0, 1fr);
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
	}
	.p-t {
		font-size: 14px;
		font-weight: 600;
		padding: 2px 4px 6px;
	}
	.p-t.colores {
		margin-top: 8px;
		padding-bottom: 2px;
	}
	.sec {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 9px 10px;
		border-radius: 14px;
		background: var(--sf);
		border: 1px solid var(--ln);
	}
	.s-t {
		flex: 1;
		font-size: 13.5px;
		font-weight: 600;
	}
	.s-t.oculta {
		opacity: 0.45;
	}
	.sec button {
		width: 26px;
		height: 26px;
		border: 0;
		border-radius: 8px;
		background: transparent;
		color: var(--mu);
		display: grid;
		place-items: center;
	}
	.sec button:disabled {
		opacity: 0.3;
	}
	.paletas {
		display: flex;
		gap: 8px;
		padding: 0 4px;
	}
	.paletas button {
		width: 36px;
		height: 36px;
		border-radius: 11px;
		border: 2px solid transparent;
		padding: 0;
	}
	.paletas button[aria-pressed='true'] {
		border-color: #d2582f;
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
	.sitio {
		transition:
			background 0.4s,
			color 0.4s;
		padding-bottom: 14px;
	}
	.s-cab {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 12px 16px;
	}
	.s-cab b {
		font-size: 15px;
	}
	.esp {
		flex: 1;
	}
	.s-cats {
		font-size: 12px;
		opacity: 0.7;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		min-width: 0;
	}
	.hero {
		position: relative;
		margin: 0 12px 12px;
		height: 170px;
		border-radius: 14px;
		overflow: hidden;
	}
	.hero img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	.sombra {
		position: absolute;
		inset: 0;
		background: linear-gradient(90deg, rgba(0, 0, 0, 0.6), transparent 70%);
	}
	.lema {
		position: absolute;
		left: 18px;
		top: 50%;
		transform: translateY(-50%);
		color: #fff;
		font-size: 22px;
		font-weight: 700;
		line-height: 1.05;
		max-width: 220px;
	}
	.prods {
		padding: 0 12px 12px;
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 8px;
	}
	.foto {
		aspect-ratio: 1;
		border-radius: 10px;
		overflow: hidden;
		background: rgba(0, 0, 0, 0.05);
	}
	.foto img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	.pn {
		margin-top: 4px;
		font-size: 11px;
		font-weight: 600;
	}
	.pp {
		font-size: 11px;
		font-weight: 700;
	}
	.cats {
		padding: 0 12px 12px;
		display: flex;
		gap: 6px;
		flex-wrap: wrap;
	}
	.cats span {
		font-size: 12px;
		font-weight: 600;
		padding: 7px 12px;
		border-radius: 999px;
		color: #fff;
	}
	.revs {
		margin: 0 12px 12px;
		padding: 12px;
		border-radius: 12px;
		background: rgba(0, 0, 0, 0.05);
		font-size: 12.5px;
		line-height: 1.4;
	}
	.estrellas {
		color: #e3a93b;
	}
	.ig {
		padding: 0 12px 12px;
		display: grid;
		grid-template-columns: repeat(6, 1fr);
		gap: 4px;
	}
	.ig img {
		width: 100%;
		aspect-ratio: 1;
		object-fit: cover;
		border-radius: 6px;
	}
	.wa {
		margin: 0 12px;
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 10px 12px;
		border-radius: 12px;
		background: #25d366;
		color: #fff;
		font-size: 12.5px;
		font-weight: 600;
	}
	.derecha {
		display: flex;
		flex-direction: column;
		gap: 16px;
		min-width: 0;
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
	.c-nota {
		margin: 0;
		padding: 0 4px;
		font-size: 12px;
		color: var(--mu);
		text-wrap: pretty;
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
	@media (max-width: 760px) {
		.c-grid {
			grid-template-columns: minmax(0, 1fr);
		}
		.burbuja {
			max-width: 100%;
		}
	}
	@media (max-width: 999px) {
		.editor {
			grid-template-columns: minmax(0, 1fr);
		}
	}
	@media (max-width: 560px) {
		.prods {
			grid-template-columns: repeat(2, 1fr);
		}
		.ig {
			grid-template-columns: repeat(3, 1fr);
		}
	}
</style>
