<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import Interruptor from '#lib/components/Interruptor.svelte';
	import * as I from '#lib/icons.ts';
	import { getUi } from '#lib/ui.svelte.ts';
	import { plantilla as plantillaDe } from '#lib/plantillas.ts';
	import { acentos, contraste } from '#lib/tienda/acentos.ts';
	import { colorValido } from '#lib/tienda/estilo/css.ts';
	import { FUENTES } from '#lib/tienda/estilo/fuentes.ts';
	import { COMBINACIONES, rolesDe } from '#lib/tienda/estilo/roles.ts';
	import { MAX_MENU, NOMBRE_PAGINA, PAGINAS, type ItemMenu } from '#lib/tienda/estilo/menu.ts';
	import type { Redes, Zona } from '#lib/tienda/estilo/estilo.ts';
	import type { Rol } from '#lib/tienda/estilo/tipos.ts';
	import { estiloFinal } from '#lib/tienda/sistema/presets.ts';
	import CampoBloque from './CampoBloque.svelte';
	import Miniatura from './Miniatura.svelte';
	import type { Ayuda } from './campos.ts';
	import { BARRA, OPCIONES_ESTILO, type OpcionEstilo } from './catalogo.ts';
	import { queContradice } from './faltan.ts';
	import type { DisenoV2, EstiloV2 } from './diseno.ts';

	/**
	 * Mi tienda › Estilo: dirección, forma, densidad, ancho, fotos, encabezado y pie; marca,
	 * colores, tipografías, barra de anuncios, menú, redes y WhatsApp flotante. Vacío es lo del punto
	 * de partida; el diseño guarda sólo lo que la tienda cambió.
	 */
	let {
		diseno,
		plantilla,
		ayuda,
		oncambiar
	}: {
		diseno: DisenoV2;
		/** El punto de partida (la plantilla del borrador o la de la tienda). */
		plantilla: string;
		ayuda: Ayuda;
		oncambiar: (d: DisenoV2, confirmar: boolean) => void;
	} = $props();
	const ui = getUi();

	const estilo = $derived<EstiloV2>(diseno.estilo ?? {});
	const final = $derived(estiloFinal(plantilla, diseno.paleta, estilo));
	function ponerEstilo(cambios: Partial<EstiloV2>, confirmar = true) {
		const e: EstiloV2 = { ...estilo, ...cambios };
		for (const k of Object.keys(cambios) as (keyof EstiloV2)[])
			if (e[k] === undefined || e[k] === '') delete e[k];
		oncambiar({ ...diseno, estilo: e }, confirmar);
	}

	// ── Estructura ──
	const DIRECCIONES = [
		['vidriera', 'Vidriera', 'Mucho aire, fotos grandes y botones rectos.'],
		['mostrador', 'Mostrador', 'Tarjetas, botón píldora y las cuotas a la vista.'],
		['pedido', 'Pedido', 'Compacta, para agregar en un toque con el pedido a la vista.']
	] as const;
	const ESTRUCTURA: [Exclude<OpcionEstilo, 'direccion'>, string][] = [
		['forma', 'Forma de botones y tarjetas'],
		['densidad', 'Espacio entre bloques'],
		['ancho', 'Ancho de la página'],
		['fotoProducto', 'Fotos de producto'],
		['encabezado', 'Encabezado'],
		['pie', 'Pie']
	];

	// ── Marca ──
	type Foto = 'logo' | 'logoOscuro' | 'icono';
	let subiendo = $state<'' | Foto>('');
	const urlDe = (c: Foto) => (c === 'logo' ? (diseno.logo ?? '') : (estilo[c] ?? ''));
	function ponerFoto(c: Foto, url: string) {
		if (c === 'logo') oncambiar({ ...diseno, logo: url || undefined }, true);
		else ponerEstilo({ [c]: url });
	}
	async function subirFoto(e: Event, c: Foto) {
		const input = e.currentTarget as HTMLInputElement;
		const archivo = input.files?.[0];
		input.value = '';
		if (!archivo) return;
		subiendo = c;
		try {
			ponerFoto(c, await ayuda.subir(archivo));
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

	// ── Colores ──
	const opciones = $derived(acentos(plantilla));
	const elegida = $derived(plantillaDe(plantilla));
	const original = $derived(rolesDe(plantilla));
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
		if (rol === 'acento') return (opciones[diseno.paleta] ?? opciones[0]).color;
		if (rol === 'encabezado') return colorVisto('fondo');
		if (rol === 'encabezadoTexto') return colorVisto('texto');
		if (rol === 'pie' || rol === 'pieTexto') {
			const o = original[rol];
			if (o === original.texto) return colorVisto('texto');
			if (o === original.fondo) return colorVisto('fondo');
		}
		return original[rol];
	}
	function ponerColores(c: Partial<Record<Rol | Zona, string>>, confirmar = true) {
		const colores = { ...estilo.colores, ...c };
		for (const k of Object.keys(colores) as (Rol | Zona)[]) if (!colores[k]) delete colores[k];
		ponerEstilo({ colores: Object.keys(colores).length ? colores : undefined }, confirmar);
	}
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
	const personalizados = $derived(Object.keys(estilo.colores ?? {}).length);
	const OPCIONES_FUENTE = Object.entries(FUENTES).sort((a, b) =>
		a[1].nombre.localeCompare(b[1].nombre)
	);

	// ── Menú ──
	const menu = $derived<ItemMenu[]>(estilo.menu ?? []);
	const DESTINOS = $derived([
		...PAGINAS.map((p) => [p, NOMBRE_PAGINA[p]] as const),
		...ayuda.cx.categorias.map((c) => [`cat:${c}`, `Categoría: ${c}`] as const),
		['link', 'Un link propio…'] as const
	]);
	const destinoDe = (a: string) => (a.startsWith('https://') ? 'link' : a);
	const ponerMenu = (m: ItemMenu[]) => ponerEstilo({ menu: m.length ? m : undefined });
	function agregarPestana() {
		const usadas = new Set(menu.map((i) => i.a));
		const p = PAGINAS.find((x) => !usadas.has(x)) ?? 'catalogo';
		ponerMenu([...menu, { t: NOMBRE_PAGINA[p], a: p }]);
	}
	function cambiarDestino(i: number, d: string) {
		const nombreViejo = DESTINOS.find(([k]) => k === menu[i].a)?.[1];
		const m = menu.map((x) => ({ ...x }));
		const nuevo = DESTINOS.find(([k]) => k === d)?.[1] ?? '';
		// El link propio se guarda recién cuando es válido.
		if (d === 'link') {
			enLink = i;
			return;
		}
		m[i].a = d;
		if (!m[i].t || m[i].t === nombreViejo || m[i].t === nombreViejo?.replace('Categoría: ', ''))
			m[i].t = nuevo.replace('Categoría: ', '');
		ponerMenu(m);
	}
	let enLink = $state<number | null>(null);
	function moverPestana(i: number) {
		const m = [...menu];
		[m[i - 1], m[i]] = [m[i], m[i - 1]];
		ponerMenu(m);
	}

	// ── Redes ──
	const REDES: [keyof Redes, string, string][] = [
		['instagram', 'Instagram', '@tutienda'],
		['tiktok', 'TikTok', '@tutienda'],
		['facebook', 'Facebook', 'facebook.com/tutienda'],
		['youtube', 'YouTube', '@tucanal'],
		['email', 'Email', 'hola@tutienda.com'],
		['telefono', 'Teléfono', '11 5555-5555']
	];
	function ponerRed(k: keyof Redes, v: string) {
		const redes = { ...estilo.redes, [k]: v.trim() };
		if (!redes[k]) delete redes[k];
		ponerEstilo({ redes: Object.keys(redes).length ? redes : undefined });
	}

	// ── Barra de anuncios ──
	const barra = $derived(estilo.barra ?? { variante: 'fija' as const, mensajes: [] });
	// Un anuncio que promete lo que la tienda no ofrece (envío gratis sin monto, cuotas sin cuotas).
	const barraContradice = $derived(
		queContradice(
			barra.mensajes.map((m) => m.texto),
			ayuda.cx
		)
	);
	function ponerBarra(cambios: Partial<NonNullable<EstiloV2['barra']>>, confirmar = true) {
		const b = { ...barra, ...cambios };
		if (!b.fondo) delete b.fondo;
		if (b.automaticos !== false) delete b.automaticos;
		ponerEstilo({ barra: b.mensajes.length || b.automaticos === false ? b : undefined }, confirmar);
	}
</script>

<div class="estilo">
	<section class="grupo-fijo" aria-labelledby="t-direccion">
		<div class="p-t" id="t-direccion">Dirección</div>
		<div class="direcciones" role="radiogroup" aria-labelledby="t-direccion">
			{#each DIRECCIONES as [id, nombre, desc] (id)}
				<button
					type="button"
					role="radio"
					aria-checked={final.direccion === id}
					class="direccion"
					onclick={() => ponerEstilo({ direccion: id })}
				>
					<Miniatura tipo="direccion" variante={id} alto={56} />
					<span class="d-txt"><b>{nombre}</b><span>{desc}</span></span>
				</button>
			{/each}
		</div>
	</section>

	<details class="grupo" open>
		<summary class="p-t">Forma y medidas</summary>
		{#each ESTRUCTURA as [k, t] (k)}
			<div class="sub" id="t-{k}">{t}</div>
			<div class="chips" role="radiogroup" aria-labelledby="t-{k}">
				{#each OPCIONES_ESTILO[k].opciones as o, i (o)}
					<button
						type="button"
						role="radio"
						aria-checked={final[k] === o}
						class="chip"
						onclick={() => ponerEstilo({ [k]: o })}>{OPCIONES_ESTILO[k].nombres[i]}</button
					>
				{/each}
			</div>
		{/each}
	</details>

	<details class="grupo">
		<summary class="p-t">Marca</summary>
		{#each [['logo', 'Logo'], ['logoOscuro', 'Logo para fondos oscuros'], ['icono', 'Ícono de la pestaña']] as const as [c, t] (c)}
			{@const url = urlDe(c)}
			<div class="foto-fila">
				<div class="miniatura" class:oscura={c === 'logoOscuro'}>
					{#if url}<img src={url} alt="" />{:else}<Icon icon={I.phStorefrontFill} size={18} />{/if}
				</div>
				<div class="f-txt">
					<b>{t}</b>
					<span
						>{url
							? 'Subido'
							: {
									logo: 'Sin logo: va el nombre en letras',
									logoOscuro: 'Para el encabezado o el pie oscuros',
									icono: 'Cuadrado, se ve en la pestaña'
								}[c]}</span
					>
				</div>
				<label class="mini-btn" class:ocupado={subiendo === c}>
					{subiendo === c ? 'Subiendo…' : url ? 'Cambiar' : 'Subir'}
					<input
						type="file"
						accept="image/jpeg,image/png,image/webp"
						class="sr-only"
						aria-label="Subir {t.toLowerCase()}"
						disabled={!!subiendo}
						onchange={(e) => subirFoto(e, c)}
					/>
				</label>
				{#if url}<button
						type="button"
						class="mini-btn icono"
						aria-label="Quitar {t.toLowerCase()}"
						onclick={() => ponerFoto(c, '')}><Icon icon={I.phTrash} size={15} /></button
					>{/if}
			</div>
		{/each}
		{#if diseno.logo}
			<label class="campo"
				><span
					>Tamaño del logo · {estilo.logoAlto
						? `${estilo.logoAlto} px`
						: 'el de la dirección'}</span
				><input
					type="range"
					min="24"
					max="72"
					step="2"
					value={estilo.logoAlto || 36}
					oninput={(e) => ponerEstilo({ logoAlto: +e.currentTarget.value }, false)}
					onchange={(e) => ponerEstilo({ logoAlto: +e.currentTarget.value })}
				/></label
			>
		{/if}
		<label class="campo"
			><span>Lema</span><input
				class="input"
				maxlength="120"
				placeholder="Ej.: Hecho a mano en Rosario"
				value={diseno.lema ?? ''}
				oninput={(e) => oncambiar({ ...diseno, lema: e.currentTarget.value }, false)}
				onchange={(e) =>
					oncambiar({ ...diseno, lema: e.currentTarget.value.trim() || undefined }, true)}
			/></label
		>
		<p class="c-nota">La portada lo usa si no le escribís una frase propia.</p>
	</details>

	<details class="grupo">
		<summary class="p-t">Colores</summary>
		<div class="sub">Acento rápido</div>
		<div class="paletas" role="group" aria-label="Color de acento">
			{#each opciones as o, i (o.color)}
				<button
					type="button"
					aria-label={i ? 'Color ' + (i + 1) : 'El de la plantilla'}
					title={i ? undefined : 'El de la plantilla'}
					aria-pressed={diseno.paleta === i && !estilo.colores?.acento}
					style:background="linear-gradient(135deg,{elegida?.bg ?? '#F3E9DC'} 50%,{o.color} 50%)"
					onclick={() => {
						const colores = { ...estilo.colores };
						delete colores.acento;
						delete colores.sobreAcento;
						oncambiar(
							{
								...diseno,
								paleta: i,
								estilo: { ...estilo, colores: Object.keys(colores).length ? colores : undefined }
							},
							true
						);
					}}
				></button>
			{/each}
		</div>
		<div class="sub">Combinaciones</div>
		<div class="combos">
			{#each COMBINACIONES as c (c.id)}
				<button
					type="button"
					class="combo"
					onclick={() =>
						ponerColores({
							...c.colores,
							encabezado: '',
							encabezadoTexto: '',
							pie: '',
							pieTexto: ''
						})}
				>
					<span class="muestra" style:background={c.colores.fondo}
						><i style:background={c.colores.superficie}></i><i style:background={c.colores.texto}
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
						onchange={(e) => ponerColores({ [rol]: e.currentTarget.value.toUpperCase() })}
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
		{#if personalizados || diseno.paleta}<button
				type="button"
				class="enlace"
				onclick={() =>
					oncambiar({ ...diseno, paleta: 0, estilo: { ...estilo, colores: undefined } }, true)}
				>Volver a los colores de la plantilla</button
			>{/if}
	</details>

	<details class="grupo">
		<summary class="p-t">Tipografías</summary>
		{#each [['fuenteTitulos', 'Títulos'], ['fuenteTexto', 'Texto']] as const as [k, t] (k)}
			<label class="campo"
				><span>{t}</span><select
					class="input"
					value={estilo[k] ?? ''}
					onchange={(e) => ponerEstilo({ [k]: e.currentTarget.value })}
				>
					<option value="">La de la dirección ({FUENTES[final[k]]?.nombre ?? final[k]})</option>
					{#each OPCIONES_FUENTE as [id, f] (id)}<option value={id}>{f.nombre}</option>{/each}
				</select></label
			>
		{/each}
	</details>

	<details class="grupo">
		<summary class="p-t">{BARRA.nombre}</summary>
		<p class="c-nota">{BARRA.descripcion.split('.')[0]}.</p>
		<CampoBloque
			campo={{
				id: 'variante',
				tipo: 'opcion',
				nombre: 'Cómo se ve',
				opciones: BARRA.variantes,
				nombres: BARRA.nombresVariantes
			}}
			valor={barra.variante}
			{ayuda}
			onvalor={(v) => ponerBarra({ variante: v as 'fija' | 'rotativa' })}
		/>
		{#each BARRA.campos as c (c.id)}
			<CampoBloque
				campo={c}
				valor={barra[c.id as 'mensajes' | 'fondo' | 'automaticos']}
				{ayuda}
				onvalor={(v, conf) =>
					ponerBarra(
						c.id === 'mensajes'
							? { mensajes: (v as { texto: string; link?: string }[] | undefined) ?? [] }
							: c.id === 'automaticos'
								? { automaticos: v !== false }
								: { fondo: v as 'oscuro' },
						conf
					)}
			/>
		{/each}
		{#if barraContradice}<p class="aviso" role="status">{barraContradice}</p>{/if}
	</details>

	<details class="grupo">
		<summary class="p-t">Menú</summary>
		<p class="c-nota">
			{menu.length
				? 'Así se ve el menú de tu tienda, en la compu y en el celular.'
				: 'Va el menú de la dirección. Armá el tuyo para sumar pestañas como Nosotros.'}
		</p>
		{#each menu as it, i (i)}
			<div class="pestana">
				<div class="p-fila">
					<input
						class="input"
						maxlength="30"
						aria-label="Texto de la pestaña {i + 1}"
						value={it.t}
						onchange={(e) => {
							const t = e.currentTarget.value.trim();
							if (t) ponerMenu(menu.map((x, k) => (k === i ? { ...x, t } : x)));
						}}
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
					value={enLink === i ? 'link' : destinoDe(it.a)}
					onchange={(e) => cambiarDestino(i, e.currentTarget.value)}
				>
					{#each DESTINOS as [k, t] (k)}<option value={k}>{t}</option>{/each}
				</select>
				{#if enLink === i || destinoDe(it.a) === 'link'}
					<input
						class="input"
						type="url"
						maxlength="300"
						placeholder="https://"
						aria-label="Link de {it.t}"
						value={destinoDe(it.a) === 'link' ? it.a : ''}
						onchange={(e) => {
							const u = e.currentTarget.value.trim();
							if (/^https:\/\/[^\s"'<>]+$/.test(u)) {
								enLink = null;
								ponerMenu(menu.map((x, k) => (k === i ? { ...x, a: u } : x)));
							} else
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
		{#if menu.length < MAX_MENU}<button type="button" class="combo agregar" onclick={agregarPestana}
				><Icon icon={I.phPlusBold} size={14} />Agregar pestaña</button
			>{/if}
		{#if menu.length}<button type="button" class="enlace" onclick={() => ponerMenu([])}
				>Volver al menú de la dirección</button
			>{/if}
	</details>

	<details class="grupo">
		<summary class="p-t">Pie, redes y WhatsApp</summary>
		<label class="campo"
			><span>Línea del pie</span><input
				class="input"
				maxlength="160"
				placeholder="Ej.: Atendemos de lunes a sábado de 9 a 19"
				value={estilo.leyendaPie ?? ''}
				onchange={(e) => ponerEstilo({ leyendaPie: e.currentTarget.value.trim() })}
			/></label
		>
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
		<div class="fila-si">
			<span class="f-txt"
				><b>Botón flotante de WhatsApp</b><span
					>{ayuda.cx.whatsapp
						? 'Con el número de Configuración › Negocio.'
						: 'Cargá tu WhatsApp en Configuración › Negocio para que aparezca.'}</span
				></span
			>
			<Interruptor
				checked={final.whatsappFlotante}
				label="Botón flotante de WhatsApp"
				onchange={(v) => ponerEstilo({ whatsappFlotante: v ? undefined : false })}
			/>
		</div>
	</details>
</div>

<style>
	.estilo {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.p-t {
		font-size: 14px;
		font-weight: 600;
		padding: 2px 4px 6px;
	}
	.direcciones {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.direccion {
		display: grid;
		grid-template-columns: 92px minmax(0, 1fr);
		gap: 10px;
		align-items: center;
		padding: 6px;
		border: 1px solid var(--ln);
		border-radius: 14px;
		background: var(--sf);
		color: var(--tx);
		font: inherit;
		text-align: left;
		cursor: pointer;
	}
	.direccion[aria-checked='true'],
	.chip[aria-checked='true'] {
		border-color: var(--terracota);
		box-shadow: 0 0 0 1px var(--terracota);
	}
	.d-txt {
		display: flex;
		flex-direction: column;
		font-size: 12.5px;
		color: var(--mu);
		text-wrap: pretty;
	}
	.d-txt b {
		font-size: 14px;
		color: var(--tx);
	}
	.chips {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
		padding: 0 4px;
	}
	.chip {
		min-height: 40px;
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
	.grupo {
		border-top: 1px solid var(--ln);
		padding-top: 4px;
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.grupo > summary {
		list-style: none;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: space-between;
		min-height: 44px;
		padding-bottom: 2px;
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
		padding-bottom: 10px;
	}
	.sub {
		font-size: 12.5px;
		font-weight: 600;
		color: var(--mu);
		padding: 8px 4px 2px;
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
		background: var(--sf2);
		color: var(--mu);
	}
	.miniatura img {
		width: 100%;
		height: 100%;
		object-fit: contain;
	}
	.miniatura.oscura {
		background: #1c1a17;
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
		text-wrap: pretty;
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
	select.input {
		appearance: auto;
	}
	.c-nota {
		margin: 0;
		padding: 0 4px;
		font-size: 12px;
		color: var(--mu);
		text-wrap: pretty;
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
		cursor: pointer;
	}
	.paletas button[aria-pressed='true'] {
		border-color: var(--terracota);
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
		min-height: 40px;
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
		background: var(--sf2);
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
		color: var(--terracota);
		cursor: pointer;
		align-self: flex-start;
		min-height: 32px;
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
	.fila-si {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 10px 4px 0;
	}
	@media (max-width: 760px) {
		.mini-btn,
		.paletas button {
			min-width: 44px;
			min-height: 44px;
		}
		.chip {
			min-height: 44px;
		}
	}
</style>
