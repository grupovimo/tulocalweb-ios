<script lang="ts">
	import { invalidate, invalidateAll } from '$app/navigation';
	import { resolve } from '$app/paths';
	import Icon from '#lib/components/Icon.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { getUi } from '#lib/ui.svelte.ts';
	import { escucharTiempoReal } from '#lib/tiempo-real/panel.svelte.ts';
	import Modal from '#lib/components/Modal.svelte';
	import Segmentado from '#lib/components/Segmentado.svelte';
	import { subir as subirArchivo } from '#lib/panel/fotos.ts';
	import { PLANTILLAS, plantilla } from '#lib/plantillas.ts';
	import { acentos, contraste } from '#lib/tienda/acentos.ts';
	import { colorValido } from '#lib/tienda/estilo/css.ts';
	import { FUENTES } from '#lib/tienda/estilo/fuentes.ts';
	import { COMBINACIONES, rolesDe } from '#lib/tienda/estilo/roles.ts';
	import { MAX_MENU, NOMBRE_PAGINA, PAGINAS, type ItemMenu } from '#lib/tienda/estilo/menu.ts';
	import type { EstiloTienda, Redes, SeccionConTitulo, Zona } from '#lib/tienda/estilo/estilo.ts';
	import type { Rol } from '#lib/tienda/estilo/tipos.ts';
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
	// Estilo: colores por rol, tipografías, logos, anuncio, textos y redes (lib/tienda/estilo).
	let estilo = $derived<EstiloTienda>(structuredClone(data.diseno.estilo ?? {}));
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
			compartirFoto,
			estilo
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
	type Foto = 'logo' | 'portada' | 'compartir' | 'logoOscuro' | 'icono';
	let subiendo = $state<'' | Foto>('');
	async function subirFoto(e: Event, cual: Foto) {
		const input = e.currentTarget as HTMLInputElement;
		const archivo = input.files?.[0];
		input.value = '';
		if (!archivo) return;
		subiendo = cual;
		try {
			const url = await subirArchivo(archivo);
			if (cual === 'logo') logo = url;
			else if (cual === 'logoOscuro' || cual === 'icono') estilo = { ...estilo, [cual]: url };
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
	function quitar(cual: Foto) {
		if (cual === 'logo') logo = '';
		else if (cual === 'logoOscuro' || cual === 'icono') estilo = { ...estilo, [cual]: '' };
		else if (cual === 'compartir') compartirFoto = '';
		else portada = '';
		void guardar();
	}
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
	// ── Estilo ──
	const original = $derived(rolesDe(plantillaId));
	const ROLES_COLOR: [Rol | Zona, string][] = [
		['fondo', 'Fondo'],
		['superficie', 'Tarjetas'],
		['texto', 'Texto'],
		['textoSuave', 'Texto suave'],
		['borde', 'Bordes'],
		['acento', 'Botones y precios'],
		['sobreAcento', 'Texto de los botones'],
		['encabezado', 'Fondo del encabezado'],
		['encabezadoTexto', 'Texto del encabezado'],
		['pie', 'Fondo del pie'],
		['pieTexto', 'Texto del pie']
	];
	/** El color que se ve: el elegido, o el acento rápido, o el de la plantilla. */
	function colorVisto(rol: Rol | Zona): string {
		const c = colorValido(estilo.colores?.[rol]);
		if (c) return c;
		if (rol === 'acento') return (opciones[paleta] ?? opciones[0]).color;
		// Sin color propio, el encabezado y el pie siguen al rol del que salen (fondo o texto).
		if (rol === 'encabezado') return colorVisto('fondo');
		if (rol === 'encabezadoTexto') return colorVisto('texto');
		if (rol === 'pie' || rol === 'pieTexto') {
			const o = original[rol];
			if (o === original.texto) return colorVisto('texto');
			if (o === original.fondo) return colorVisto('fondo');
		}
		return original[rol];
	}
	function ponerColores(c: Partial<Record<Rol | Zona, string>>, guardarYa = true) {
		const colores = { ...estilo.colores, ...c };
		for (const k of Object.keys(colores) as (Rol | Zona)[]) if (!colores[k]) delete colores[k];
		estilo = { ...estilo, colores };
		if (guardarYa) void guardar();
	}
	const personalizados = $derived(Object.keys(estilo.colores ?? {}).length);
	// Avisos de contraste (AA: 4,5 para texto; 3 para el suave).
	const avisos = $derived(
		[
			contraste(colorVisto('texto'), colorVisto('fondo')) < 4.5 &&
				'El texto casi no se lee sobre el fondo.',
			contraste(colorVisto('textoSuave'), colorVisto('fondo')) < 3 &&
				'El texto suave casi no se lee sobre el fondo.',
			contraste(colorVisto('sobreAcento'), colorVisto('acento')) < 4.5 &&
				'El texto de los botones casi no se lee sobre su color.',
			contraste(colorVisto('pieTexto'), colorVisto('pie')) < 4.5 &&
				'El texto del pie casi no se lee.'
		].filter((x): x is string => !!x)
	);
	function combinacion(c: Partial<Record<Rol, string>>) {
		ponerColores({ ...c, encabezado: '', encabezadoTexto: '', pie: '', pieTexto: '' });
	}
	function colorDeLaPlantilla() {
		estilo = { ...estilo, colores: undefined };
		paleta = 0;
		void guardar();
	}
	const OPCIONES_FUENTE = Object.entries(FUENTES).sort((a, b) =>
		a[1].nombre.localeCompare(b[1].nombre)
	);
	function ponerEstilo<K extends keyof EstiloTienda>(k: K, v: EstiloTienda[K]) {
		estilo = { ...estilo, [k]: v };
		void guardar();
	}
	const TITULOS: [SeccionConTitulo, string][] = [
		['cats', 'Categorías'],
		['prods', 'Productos destacados'],
		['revs', 'Reseñas'],
		['ig', 'Fotos']
	];
	function ponerTitulo(id: SeccionConTitulo, v: string) {
		estilo = { ...estilo, textos: { ...estilo.textos, [id]: v } };
		void guardar();
	}
	const REDES: [keyof Redes, string, string][] = [
		['instagram', 'Instagram', '@tutienda'],
		['tiktok', 'TikTok', '@tutienda'],
		['facebook', 'Facebook', 'facebook.com/tutienda'],
		['youtube', 'YouTube', '@tucanal'],
		['email', 'Email', 'hola@tutienda.com'],
		['telefono', 'Teléfono', '11 5555-5555']
	];
	function ponerRed(k: keyof Redes, v: string) {
		estilo = { ...estilo, redes: { ...estilo.redes, [k]: v } };
		void guardar();
	}
	// Menú: las pestañas propias (vacío: las de la plantilla).
	const menu = $derived<ItemMenu[]>(estilo.menu ?? []);
	const DESTINOS = $derived([
		...PAGINAS.map((p) => [p, NOMBRE_PAGINA[p]] as const),
		...data.vista.categorias.map((c) => [`cat:${c}`, `Categoría: ${c}`] as const),
		['link', 'Un link propio…'] as const
	]);
	const destinoDe = (a: string) => (a.startsWith('https://') ? 'link' : a);
	function ponerMenu(m: ItemMenu[], guardarYa = true) {
		estilo = { ...estilo, menu: m };
		if (guardarYa) void guardar();
	}
	function agregarPestana() {
		const usadas = new Set(menu.map((i) => i.a));
		const p = PAGINAS.find((x) => !usadas.has(x)) ?? 'catalogo';
		ponerMenu([...menu, { t: NOMBRE_PAGINA[p], a: p }]);
	}
	function cambiarDestino(i: number, d: string) {
		const nombreViejo = DESTINOS.find(([k]) => k === menu[i].a)?.[1];
		const m = menu.map((x) => ({ ...x }));
		// Si el texto era el nombre del destino, sigue al nuevo.
		const nuevo = DESTINOS.find(([k]) => k === d)?.[1] ?? '';
		if (d === 'link') m[i].a = 'https://';
		else m[i].a = d;
		if (
			d !== 'link' &&
			(!m[i].t || m[i].t === nombreViejo || m[i].t === nombreViejo?.replace('Categoría: ', ''))
		)
			m[i].t = nuevo.replace('Categoría: ', '');
		ponerMenu(m, d !== 'link');
	}
	function moverPestana(i: number) {
		if (!i) return;
		const m = [...menu];
		[m[i - 1], m[i]] = [m[i], m[i - 1]];
		ponerMenu(m);
	}
	// La vista previa: la tienda de verdad con el borrador (se recarga en cada guardado).
	let dispositivo = $state<'celular' | 'compu'>('celular');
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
			{#each [['logo', 'Logo', logo], ['logoOscuro', 'Logo para fondos oscuros', estilo.logoOscuro ?? ''], ['icono', 'Ícono de la pestaña', estilo.icono ?? ''], ['portada', 'Foto de portada', portada]] as const as [cual, t, url] (cual)}
				<div class="foto-fila">
					<div
						class="miniatura"
						class:logo={cual !== 'portada'}
						class:oscura={cual === 'logoOscuro'}
					>
						{#if url}<img src={url} alt="" />{:else}<Icon
								icon={cual === 'portada' ? I.phImageFill : I.phStorefrontFill}
								size={18}
							/>{/if}
					</div>
					<div class="f-txt">
						<b>{t}</b>
						<span
							>{url
								? 'Subido'
								: {
										logo: 'Sin logo: va el nombre en letras',
										logoOscuro: 'Para el encabezado o el pie oscuros',
										icono: 'Cuadrado, se ve en la pestaña',
										portada: 'Sin foto: va la de un producto'
									}[cual]}</span
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
			{#if logo}
				<label class="campo"
					><span
						>Tamaño del logo · {estilo.logoAlto || 'el de la plantilla'}{estilo.logoAlto
							? ' px'
							: ''}</span
					><input
						type="range"
						min="24"
						max="72"
						step="2"
						value={estilo.logoAlto || 36}
						oninput={(e) => (estilo = { ...estilo, logoAlto: +e.currentTarget.value })}
						onchange={guardar}
					/></label
				>
			{/if}
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
			<details class="grupo" open>
				<summary class="p-t colores">Colores</summary>
				<div class="sub">Acento rápido</div>
				<div class="paletas" role="group" aria-label="Color de acento">
					{#each opciones as o, i (o.color)}
						<button
							type="button"
							aria-label={i ? 'Color ' + (i + 1) : 'El de la plantilla'}
							title={i ? undefined : 'El de la plantilla'}
							aria-pressed={paleta === i && !estilo.colores?.acento}
							style:background="linear-gradient(135deg,{pal.bg} 50%,{o.color} 50%)"
							onclick={() => {
								paleta = i;
								ponerColores({ acento: '', sobreAcento: '' });
							}}
						></button>
					{/each}
				</div>
				<div class="sub">Combinaciones</div>
				<div class="combos">
					{#each COMBINACIONES as c (c.id)}
						<button type="button" class="combo" onclick={() => combinacion(c.colores)}>
							<span class="muestra" style:background={c.colores.fondo}
								><i style:background={c.colores.superficie}></i><i
									style:background={c.colores.texto}
								></i><i style:background={colorVisto('acento')}></i></span
							>{c.nombre}
						</button>
					{/each}
				</div>
				<div class="sub">Todos los colores</div>
				<div class="lista-colores">
					{#each ROLES_COLOR as [rol, t] (rol)}
						<label class="color-fila">
							<input
								type="color"
								value={colorVisto(rol).toLowerCase()}
								oninput={(e) => ponerColores({ [rol]: e.currentTarget.value.toUpperCase() }, false)}
								onchange={() => void guardar()}
							/>
							<span class="c-n">{t}</span>
							<span class="c-hex">{colorVisto(rol)}</span>
							{#if estilo.colores?.[rol]}<button
									type="button"
									class="mini-btn icono"
									aria-label="Volver {t.toLowerCase()} al de la plantilla"
									onclick={(e) => {
										e.preventDefault();
										ponerColores({ [rol]: '' });
									}}><Icon icon={I.phArrowCounterClockwiseBold} size={14} /></button
								>{/if}
						</label>
					{/each}
				</div>
				{#each avisos as a (a)}<p class="aviso">{a}</p>{/each}
				{#if personalizados || paleta}<button
						type="button"
						class="enlace"
						onclick={colorDeLaPlantilla}>Volver a los colores de la plantilla</button
					>{/if}
			</details>

			<details class="grupo">
				<summary class="p-t colores">Tipografías</summary>
				{#each [['fuenteTitulos', 'Títulos'], ['fuenteTexto', 'Texto']] as const as [k, t] (k)}
					<label class="campo"
						><span>{t}</span><select
							class="input"
							value={estilo[k] ?? ''}
							onchange={(e) => ponerEstilo(k, e.currentTarget.value)}
						>
							<option value="">La de la plantilla</option>
							{#each OPCIONES_FUENTE as [id, f] (id)}<option value={id}>{f.nombre}</option>{/each}
						</select></label
					>
				{/each}
			</details>

			<details class="grupo">
				<summary class="p-t colores">Menú</summary>
				<p class="c-nota">
					{menu.length
						? 'Así se ve el menú de tu tienda, en la compu y en el celular.'
						: 'Va el menú de la plantilla. Armá el tuyo para sumar pestañas como Nosotros.'}
				</p>
				{#each menu as it, i (i)}
					<div class="pestana">
						<div class="p-fila">
							<input
								class="input"
								maxlength="30"
								aria-label="Texto de la pestaña {i + 1}"
								value={it.t}
								onchange={(e) =>
									ponerMenu(menu.map((x, k) => (k === i ? { ...x, t: e.currentTarget.value } : x)))}
							/>
							<button
								type="button"
								class="mini-btn icono"
								aria-label="Subir la pestaña {it.t}"
								disabled={i === 0}
								onclick={() => moverPestana(i)}><Icon icon={I.phCaretUpBold} size={12} /></button
							>
							<button
								type="button"
								class="mini-btn icono"
								aria-label="Quitar la pestaña {it.t}"
								onclick={() => ponerMenu(menu.filter((_, k) => k !== i))}
								><Icon icon={I.phTrash} size={14} /></button
							>
						</div>
						<select
							class="input"
							aria-label="Adónde lleva {it.t}"
							value={destinoDe(it.a)}
							onchange={(e) => cambiarDestino(i, e.currentTarget.value)}
						>
							{#each DESTINOS as [k, t] (k)}<option value={k}>{t}</option>{/each}
						</select>
						{#if destinoDe(it.a) === 'link'}
							<input
								class="input"
								type="url"
								maxlength="300"
								placeholder="https://"
								aria-label="Link de {it.t}"
								value={it.a}
								onchange={(e) => {
									const u = e.currentTarget.value.trim();
									if (/^https:\/\/[^\s"'<>]+$/.test(u))
										ponerMenu(menu.map((x, k) => (k === i ? { ...x, a: u } : x)));
									else
										ui.toast({
											k: 'err',
											title: 'Ese link no es válido',
											desc: 'Tiene que empezar con https://'
										});
								}}
							/>
						{/if}
					</div>
				{/each}
				{#if menu.length < MAX_MENU}<button
						type="button"
						class="combo agregar"
						onclick={agregarPestana}><Icon icon={I.phPlusBold} size={14} />Agregar pestaña</button
					>{/if}
				{#if menu.length}<button type="button" class="enlace" onclick={() => ponerMenu([])}
						>Volver al menú de la plantilla</button
					>{/if}
			</details>

			<details class="grupo">
				<summary class="p-t colores">Títulos de las secciones</summary>
				{#each TITULOS as [id, t] (id)}
					<label class="campo"
						><span>{t}</span><input
							class="input"
							maxlength="60"
							placeholder="El de la plantilla"
							value={estilo.textos?.[id] ?? ''}
							onchange={(e) => ponerTitulo(id, e.currentTarget.value)}
						/></label
					>
				{/each}
			</details>

			<details class="grupo">
				<summary class="p-t colores">Anuncio y pie</summary>
				<label class="campo"
					><span>Barra de anuncio (arriba de todo)</span><input
						class="input"
						maxlength="120"
						placeholder="Ej.: Envío gratis desde $ 50.000"
						value={estilo.anuncio ?? ''}
						onchange={(e) => ponerEstilo('anuncio', e.currentTarget.value)}
					/></label
				>
				<label class="campo"
					><span>Link del anuncio (opcional)</span><input
						class="input"
						type="url"
						maxlength="300"
						placeholder="https://"
						value={estilo.anuncioLink ?? ''}
						onchange={(e) => ponerEstilo('anuncioLink', e.currentTarget.value.trim())}
					/></label
				>
				<label class="campo"
					><span>Línea del pie</span><input
						class="input"
						maxlength="160"
						placeholder="Ej.: Atendemos de lunes a sábado de 9 a 19"
						value={estilo.leyendaPie ?? ''}
						onchange={(e) => ponerEstilo('leyendaPie', e.currentTarget.value)}
					/></label
				>
			</details>

			<details class="grupo">
				<summary class="p-t colores">Redes y contacto</summary>
				{#each REDES as [k, t, ph] (k)}
					<label class="campo"
						><span>{t}</span><input
							class="input"
							maxlength={k === 'email' ? 120 : k === 'telefono' ? 30 : 100}
							placeholder={ph}
							value={estilo.redes?.[k] ?? ''}
							onchange={(e) => ponerRed(k, e.currentTarget.value)}
						/></label
					>
				{/each}
				<p class="c-nota">Se muestran en el pie de tu tienda.</p>
			</details>
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
						title="Vista previa de tu tienda"
						src="/panel/mi-tienda/vista?v={vuelta}"
						loading="lazy"
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
	}
	/* Estilo: grupos plegables con los títulos de sección de siempre (.p-t). */
	.grupo {
		border-top: 1px solid var(--ln);
		padding-top: 4px;
	}
	.grupo > summary {
		list-style: none;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: space-between;
		min-height: 36px;
	}
	.grupo > summary::-webkit-details-marker {
		display: none;
	}
	.grupo > summary::after {
		content: '';
		width: 7px;
		height: 7px;
		border-right: 2px solid var(--mu);
		border-bottom: 2px solid var(--mu);
		transform: rotate(45deg);
		margin-right: 6px;
		transition: transform 0.15s;
	}
	.grupo[open] > summary::after {
		transform: rotate(-135deg);
	}
	.grupo[open] {
		padding-bottom: 8px;
	}
	.sub {
		font-size: 12.5px;
		font-weight: 600;
		color: var(--mu);
		padding: 8px 4px 6px;
	}
	.combos {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(118px, 1fr));
		gap: 6px;
	}
	.combo {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 6px 8px;
		border: 1px solid var(--ln);
		border-radius: 10px;
		background: none;
		font: inherit;
		font-size: 13px;
		font-weight: 600;
		color: inherit;
		cursor: pointer;
		text-align: left;
	}
	.combo:hover {
		border-color: var(--mu);
	}
	.muestra {
		display: flex;
		gap: 2px;
		padding: 4px;
		border-radius: 7px;
		box-shadow: inset 0 0 0 1px rgba(0, 0, 0, 0.12);
	}
	.muestra i {
		width: 7px;
		height: 14px;
		border-radius: 2px;
	}
	.lista-colores {
		display: flex;
		flex-direction: column;
	}
	.color-fila {
		display: flex;
		align-items: center;
		gap: 10px;
		min-height: 40px;
		padding: 2px 4px;
		border-radius: 8px;
		cursor: pointer;
	}
	.color-fila:hover {
		background: var(--sf2, rgba(0, 0, 0, 0.04));
	}
	.color-fila input[type='color'] {
		width: 30px;
		height: 30px;
		flex: none;
		padding: 0;
		border: 0;
		border-radius: 8px;
		background: none;
		cursor: pointer;
	}
	.color-fila input[type='color']::-webkit-color-swatch-wrapper {
		padding: 0;
	}
	.color-fila input[type='color']::-webkit-color-swatch {
		border: 1px solid rgba(0, 0, 0, 0.15);
		border-radius: 8px;
	}
	.c-n {
		flex: 1;
		font-size: 14px;
	}
	.c-hex {
		font:
			500 12px ui-monospace,
			monospace;
		color: var(--mu);
	}
	.aviso {
		margin: 8px 4px 0;
		font-size: 13px;
		color: #9a3412;
		text-wrap: pretty;
	}
	.enlace {
		margin: 8px 4px 0;
		padding: 0;
		border: 0;
		background: none;
		font: inherit;
		font-size: 13px;
		font-weight: 600;
		color: #b0441d;
		cursor: pointer;
	}
	.miniatura.oscura {
		background: #1c1a17;
	}
	select.input {
		appearance: auto;
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
	.pestana {
		display: flex;
		flex-direction: column;
		gap: 6px;
		padding: 10px 0;
		border-bottom: 1px solid var(--ln);
	}
	.p-fila {
		display: flex;
		align-items: center;
		gap: 6px;
	}
	.p-fila .input {
		flex: 1;
		min-width: 0;
	}
	.agregar {
		margin-top: 10px;
		justify-content: center;
		width: 100%;
	}
</style>
