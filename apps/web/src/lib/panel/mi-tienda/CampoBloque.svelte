<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import Interruptor from '#lib/components/Interruptor.svelte';
	import * as I from '#lib/icons.ts';
	import Self from './CampoBloque.svelte';
	import CampoDescarga from './CampoDescarga.svelte';
	import CampoFotos from './CampoFotos.svelte';
	import CampoLink from './CampoLink.svelte';
	import CampoTabla from './CampoTabla.svelte';
	import ElegirProductos from './ElegirProductos.svelte';
	import { fechaLocal, isoDeLocal, URL_VIDEO, type Ayuda } from './campos.ts';
	import { FONDOS, FUENTES_PRODUCTOS, ICONOS, LIMITES, type CampoConNombres } from './catalogo.ts';
	import { ICONO_BENEFICIO } from './iconos.ts';
	import { TIPOS_PREMIO, VIDEO_CORTO } from '#lib/tienda/sistema/catalogo.ts';

	const NOMBRE_PREMIO: Record<string, string> = {
		nada: 'Sólo suscribirse',
		cuponFijo: 'Un cupón tuyo',
		cuponUnico: 'Un código propio',
		descarga: 'Una descarga'
	};

	/**
	 * Un campo del panel del bloque, armado desde su tipo en bloques.json (tiposCampo). `onvalor`
	 * con `confirmar` false sólo redibuja la vista previa (mientras se escribe); con true entra al
	 * historial y se guarda.
	 */
	let {
		campo,
		valor,
		ayuda,
		onvalor
	}: {
		campo: CampoConNombres;
		valor: unknown;
		ayuda: Ayuda;
		onvalor: (v: unknown, confirmar: boolean) => void;
	} = $props();

	const uid = $props.id();
	const str = $derived(typeof valor === 'string' ? valor : '');
	const placeholder = $derived(
		campo.sugerido === 'nombre'
			? ayuda.nombre
			: campo.sugerido === 'lema'
				? ayuda.lema || 'Ej.: Hecho a mano en Rosario'
				: typeof campo.porDefecto === 'string'
					? campo.porDefecto
					: ''
	);
	const lista = $derived(Array.isArray(valor) ? (valor as Record<string, unknown>[]) : []);
	const obj = $derived(
		valor && typeof valor === 'object' && !Array.isArray(valor)
			? (valor as Record<string, unknown>)
			: {}
	);
	const max = $derived(campo.max ?? LIMITES.itemsPorLista);
	let errorVideo = $state('');

	// Productos: la fuente, cuántos y (según la fuente) cuáles o de qué categoría.
	type Sel = { fuente: string; cantidad: number; ids?: string[]; categoria?: string };
	const sel = $derived<Sel>({ ...(campo.porDefecto as Sel), ...(obj as Partial<Sel>) });
	function ponerSel(cambios: Partial<Sel>, confirmar = true) {
		const s: Sel = { ...sel, ...cambios };
		if (s.fuente !== 'elegidos') delete s.ids;
		if (s.fuente !== 'categoria') delete s.categoria;
		onvalor(s, confirmar);
	}

	function itemCon(i: number, k: string, v: unknown) {
		return lista.map((x, j) => {
			if (j !== i) return x;
			const n = { ...x };
			if (v === '' || v === undefined) delete n[k];
			else n[k] = v;
			return n;
		});
	}
	function moverItem(i: number) {
		const r = [...lista];
		[r[i - 1], r[i]] = [r[i], r[i - 1]];
		onvalor(r, true);
	}
	/** El renglón que identifica un ítem en la lista: su primer texto. */
	function resumen(x: Record<string, unknown>, i: number) {
		for (const c of campo.item ?? []) {
			const v = x[c.id];
			if (typeof v === 'string' && v.trim() && c.tipo !== 'foto') return v.trim();
		}
		return `${campo.nombre} ${i + 1}`;
	}
	let abierto = $state<number | null>(null);
</script>

{#snippet elegirCupon(
	nombre: string,
	actual: string,
	alElegir: (c: string) => void,
	requerido?: boolean
)}
	<label class="campo">
		<span class="rotulo"
			>{nombre}{#if requerido}<span class="req" aria-hidden="true">&#32;·&nbsp;obligatorio</span
				>{/if}</span
		>
		{#if ayuda.cupones?.length}
			<select class="input" value={actual} onchange={(e) => alElegir(e.currentTarget.value)}>
				<option value="">Elegí un cupón</option>
				{#each ayuda.cupones as c (c.codigo)}<option value={c.codigo}
						>{c.codigo}{c.descripcion ? ` · ${c.descripcion}` : ''}</option
					>{/each}
			</select>
		{:else}
			<span class="ayuda">Todavía no tenés cupones activos: creá uno en Cupones.</span>
		{/if}
	</label>
{/snippet}

{#snippet rotulo()}
	<span class="rotulo"
		>{campo.nombre}{#if campo.requerido}<span class="req" aria-hidden="true"
				>&#32;·&nbsp;obligatorio</span
			>{/if}</span
	>
{/snippet}

{#if campo.tipo === 'texto' || campo.tipo === 'textoLargo'}
	<label class="campo">
		{@render rotulo()}
		{#if campo.tipo === 'texto'}
			<input
				class="input"
				maxlength={campo.max}
				{placeholder}
				required={campo.requerido}
				value={str}
				oninput={(e) => onvalor(e.currentTarget.value, false)}
				onchange={(e) => onvalor(e.currentTarget.value.trim(), true)}
			/>
		{:else}
			<textarea
				class="textarea"
				rows="4"
				maxlength={campo.max}
				{placeholder}
				value={str}
				oninput={(e) => onvalor(e.currentTarget.value, false)}
				onchange={(e) => onvalor(e.currentTarget.value.trim(), true)}></textarea>
		{/if}
		{#if campo.max && str.length > campo.max * 0.8}<span class="cuenta"
				>{str.length} / {campo.max}</span
			>{/if}
	</label>
{:else if campo.tipo === 'foto' || campo.tipo === 'fotos'}
	<div class="campo">
		{@render rotulo()}
		<CampoFotos
			fotos={campo.tipo === 'foto' ? (str ? [str] : []) : (lista as unknown as string[])}
			max={campo.tipo === 'foto' ? 1 : max}
			etiqueta={campo.nombre}
			subir={ayuda.subir}
			onfotos={(f) => onvalor(campo.tipo === 'foto' ? (f[0] ?? '') : f, true)}
		/>
	</div>
{:else if campo.tipo === 'link'}
	<div class="campo">
		{@render rotulo()}
		<CampoLink valor={str} etiqueta={campo.nombre} {ayuda} onvalor={(v, c) => onvalor(v, c)} />
	</div>
{:else if campo.tipo === 'boton'}
	{@const b = obj as { t?: string; a?: string }}
	<fieldset class="campo grupo-campo">
		<legend class="rotulo">{campo.nombre}</legend>
		<input
			class="input"
			maxlength="30"
			placeholder="Texto del botón (vacío: sin botón)"
			aria-label="{campo.nombre}: texto"
			value={b.t ?? ''}
			oninput={(e) => onvalor({ ...b, t: e.currentTarget.value }, false)}
			onchange={(e) => {
				const t = e.currentTarget.value.trim();
				onvalor(t ? { a: b.a || 'catalogo', t } : undefined, true);
			}}
		/>
		{#if b.t}
			<CampoLink
				valor={b.a ?? 'catalogo'}
				etiqueta="{campo.nombre}: adónde lleva"
				{ayuda}
				onvalor={(a, c) => onvalor(a ? { ...b, a } : { t: b.t, a: 'catalogo' }, c)}
			/>
		{/if}
	</fieldset>
{:else if campo.tipo === 'fondo'}
	{@const actual = str || (campo.porDefecto as string) || 'normal'}
	<div class="campo">
		{@render rotulo()}
		<div class="fondos" role="radiogroup" aria-label={campo.nombre}>
			{#each FONDOS.opciones as f, i (f)}
				<button
					type="button"
					role="radio"
					aria-checked={actual === f}
					class="fondo"
					onclick={() => onvalor(f === campo.porDefecto ? '' : f, true)}
					><span class="muestra {f}"></span>{FONDOS.nombres[i]}</button
				>
			{/each}
		</div>
	</div>
{:else if campo.tipo === 'fecha'}
	<label class="campo">
		{@render rotulo()}
		<input
			class="input"
			type="datetime-local"
			value={fechaLocal(valor)}
			onchange={(e) => onvalor(isoDeLocal(e.currentTarget.value), true)}
		/>
		<span class="ayuda">Hora de Argentina.</span>
	</label>
{:else if campo.tipo === 'numero'}
	{@const n = typeof valor === 'number' ? valor : ((campo.porDefecto as number) ?? campo.min ?? 0)}
	<div class="campo">
		<span class="rotulo" id="n-{uid}">{campo.nombre}</span>
		<div class="numero" role="group" aria-labelledby="n-{uid}">
			<button
				type="button"
				class="mini-btn icono"
				aria-label="Menos"
				disabled={n <= (campo.min ?? 0)}
				onclick={() => onvalor(n - 1, true)}><Icon icon={I.phMinusBold} size={13} /></button
			>
			<input
				class="input"
				type="number"
				inputmode="numeric"
				aria-labelledby="n-{uid}"
				min={campo.min}
				max={campo.max}
				value={n}
				onchange={(e) => {
					const x = Math.round(+e.currentTarget.value);
					const v = Math.min(campo.max ?? x, Math.max(campo.min ?? 0, x));
					e.currentTarget.value = String(v);
					onvalor(v, true);
				}}
			/>
			<button
				type="button"
				class="mini-btn icono"
				aria-label="Más"
				disabled={campo.max !== undefined && n >= campo.max}
				onclick={() => onvalor(n + 1, true)}><Icon icon={I.phPlusBold} size={13} /></button
			>
		</div>
	</div>
{:else if campo.tipo === 'si'}
	{@const on = typeof valor === 'boolean' ? valor : !!campo.porDefecto}
	<div class="campo fila-si">
		<span class="rotulo">{campo.nombre}</span>
		<Interruptor checked={on} label={campo.nombre} onchange={(v) => onvalor(v, true)} />
	</div>
{:else if campo.tipo === 'opcion'}
	{@const actual = str || (campo.porDefecto as string) || campo.opciones?.[0]}
	<div class="campo">
		{@render rotulo()}
		<div class="opciones" role="radiogroup" aria-label={campo.nombre}>
			{#each campo.opciones ?? [] as o, i (o)}
				<button
					type="button"
					role="radio"
					aria-checked={actual === o}
					class="opcion"
					onclick={() => onvalor(o, true)}>{campo.nombres?.[i] ?? o}</button
				>
			{/each}
		</div>
	</div>
{:else if campo.tipo === 'icono'}
	<div class="campo">
		{@render rotulo()}
		<div class="iconos" role="radiogroup" aria-label={campo.nombre}>
			{#each ICONOS as ic (ic)}
				{@const [svg, nombre] = ICONO_BENEFICIO[ic]}
				<button
					type="button"
					role="radio"
					aria-checked={str === ic}
					aria-label={nombre}
					title={nombre}
					class="icono-op"
					onclick={() => onvalor(ic, true)}><Icon icon={svg} size={18} /></button
				>
			{/each}
		</div>
	</div>
{:else if campo.tipo === 'video'}
	<label class="campo">
		{@render rotulo()}
		<input
			class="input"
			type="url"
			inputmode="url"
			maxlength="300"
			placeholder="https://www.youtube.com/watch?v=…"
			aria-invalid={!!errorVideo}
			value={str}
			onchange={(e) => {
				const u = e.currentTarget.value.trim();
				errorVideo = !u || URL_VIDEO.test(u) ? '' : 'Pegá el link de un video de YouTube o Vimeo.';
				if (!errorVideo) onvalor(u, true);
			}}
		/>
		{#if errorVideo}<span class="error">{errorVideo}</span>{/if}
	</label>
{:else if campo.tipo === 'videoCorto'}
	<label class="campo">
		{@render rotulo()}
		<input
			class="input"
			type="url"
			inputmode="url"
			maxlength="300"
			placeholder="https://www.instagram.com/reel/…"
			aria-invalid={!!errorVideo}
			value={str}
			onchange={(e) => {
				const u = e.currentTarget.value.trim();
				errorVideo =
					!u || VIDEO_CORTO.test(u)
						? ''
						: 'Pegá el link de un reel de Instagram, un TikTok, un Short de YouTube o un archivo .mp4.';
				if (!errorVideo) onvalor(u, true);
			}}
		/>
		{#if errorVideo}<span class="error">{errorVideo}</span>{:else}<span class="ayuda"
				>Un reel, un TikTok, un Short o un .mp4. Sumale una portada vertical.</span
			>{/if}
	</label>
{:else if campo.tipo === 'producto' || campo.tipo === 'elegidos'}
	<div class="campo">
		{@render rotulo()}
		<ElegirProductos
			ids={campo.tipo === 'producto' ? (str ? [str] : []) : (lista as unknown as string[])}
			productos={ayuda.cx.productos}
			etiqueta={campo.nombre}
			max={campo.tipo === 'producto' ? 1 : (campo.max ?? 24)}
			onids={(ids) => onvalor(campo.tipo === 'producto' ? (ids[0] ?? '') : ids, true)}
		/>
	</div>
{:else if campo.tipo === 'productos'}
	<fieldset class="campo grupo-campo">
		<legend class="rotulo">{campo.nombre}</legend>
		<select
			class="input"
			aria-label="{campo.nombre}: de dónde salen"
			value={sel.fuente}
			onchange={(e) => ponerSel({ fuente: e.currentTarget.value })}
		>
			{#each FUENTES_PRODUCTOS.opciones as f, i (f)}<option value={f}
					>{FUENTES_PRODUCTOS.nombres[i]}</option
				>{/each}
		</select>
		{#if sel.fuente === 'elegidos'}
			<ElegirProductos
				ids={sel.ids ?? []}
				productos={ayuda.cx.productos}
				etiqueta={campo.nombre}
				max={24}
				onids={(ids) => ponerSel({ ids, cantidad: Math.max(1, ids.length) })}
			/>
		{:else}
			{#if sel.fuente === 'categoria'}
				<select
					class="input"
					aria-label="{campo.nombre}: categoría"
					value={sel.categoria ?? ''}
					onchange={(e) => ponerSel({ categoria: e.currentTarget.value || undefined })}
				>
					<option value="">Elegí la categoría</option>
					{#each ayuda.cx.categorias as c (c)}<option value={c}>{c}</option>{/each}
				</select>
			{/if}
			<Self
				campo={{ id: 'cantidad', tipo: 'numero', nombre: 'Cuántos', min: 1, max: 24 }}
				valor={sel.cantidad}
				{ayuda}
				onvalor={(v) => ponerSel({ cantidad: v as number })}
			/>
		{/if}
	</fieldset>
{:else if campo.tipo === 'categoria'}
	<label class="campo">
		{@render rotulo()}
		<select class="input" value={str} onchange={(e) => onvalor(e.currentTarget.value, true)}>
			<option value="">Elegí la categoría</option>
			{#each ayuda.cx.categorias as c (c)}<option value={c}>{c}</option>{/each}
		</select>
	</label>
{:else if campo.tipo === 'categorias'}
	{@const elegidas = lista as unknown as string[]}
	<fieldset class="campo grupo-campo">
		<legend class="rotulo">{campo.nombre}</legend>
		{#each ayuda.cx.categorias as c (c)}
			<label class="check">
				<input
					type="checkbox"
					checked={elegidas.includes(c)}
					onchange={(e) =>
						onvalor(
							e.currentTarget.checked
								? ayuda.cx.categorias.filter((x) => x === c || elegidas.includes(x))
								: elegidas.filter((x) => x !== c),
							true
						)}
				/>{c}
			</label>
		{:else}
			<p class="ayuda">Tus productos todavía no tienen categorías.</p>
		{/each}
	</fieldset>
{:else if campo.tipo === 'grupo'}
	<details class="campo sub">
		<summary class="rotulo">{campo.nombre}</summary>
		<div class="hijos">
			{#each campo.item ?? [] as c (c.id)}
				<Self
					campo={c}
					valor={obj[c.id]}
					{ayuda}
					onvalor={(v, conf) => {
						const n = { ...obj };
						if (v === '' || v === undefined) delete n[c.id];
						else n[c.id] = v;
						onvalor(Object.keys(n).length ? n : undefined, conf);
					}}
				/>
			{/each}
		</div>
	</details>
{:else if campo.tipo === 'cupon'}
	{@render elegirCupon(campo.nombre, str, (c) => onvalor(c, true), campo.requerido)}
{:else if campo.tipo === 'premio'}
	{@const tipos = campo.opciones?.length ? campo.opciones : TIPOS_PREMIO}
	{@const pr = { tipo: 'nada', ...obj } as Record<string, unknown>}
	{@const poner = (cambios: Record<string, unknown>) => {
		const n: Record<string, unknown> = { ...pr, ...cambios };
		for (const k of Object.keys(n)) if (n[k] === '' || n[k] === undefined) delete n[k];
		onvalor(n, true);
	}}
	<fieldset class="campo grupo-campo">
		<legend class="rotulo">{campo.nombre}</legend>
		<div class="opciones" role="radiogroup" aria-label="Tipo de premio">
			{#each tipos as t (t)}
				<button
					type="button"
					role="radio"
					aria-checked={pr.tipo === t}
					class="opcion"
					onclick={() => onvalor({ tipo: t }, true)}>{NOMBRE_PREMIO[t] ?? t}</button
				>
			{/each}
		</div>
		{#if pr.tipo === 'cuponFijo'}
			{@render elegirCupon('Cupón', (pr.cupon as string) ?? '', (c) => poner({ cupon: c }), true)}
		{:else if pr.tipo === 'cuponUnico'}
			<div class="opciones" role="radiogroup" aria-label="Descuento">
				{#each [['porcentaje', 'Porcentaje'], ['monto', 'Monto fijo'], ['envioGratis', 'Envío gratis']] as [d, n] (d)}
					<button
						type="button"
						role="radio"
						aria-checked={pr.descuento === d}
						class="opcion"
						onclick={() =>
							poner({
								descuento: d,
								valor: d === 'envioGratis' ? undefined : d === 'monto' ? 5000 : 10
							})}>{n}</button
					>
				{/each}
			</div>
			<div class="dos">
				{#if pr.descuento && pr.descuento !== 'envioGratis'}
					<label class="campo"
						><span class="rotulo">{pr.descuento === 'monto' ? 'Pesos' : 'Porcentaje'}</span><input
							class="input"
							type="number"
							inputmode="numeric"
							min={pr.descuento === 'monto' ? 100 : 1}
							max={pr.descuento === 'monto' ? 100000 : 90}
							value={pr.valor ?? ''}
							onchange={(e) => {
								const lo = pr.descuento === 'monto' ? 100 : 1;
								const hi = pr.descuento === 'monto' ? 100000 : 90;
								const v = Math.min(hi, Math.max(lo, Math.round(+e.currentTarget.value || lo)));
								e.currentTarget.value = String(v);
								poner({ valor: v });
							}}
						/></label
					>
				{/if}
				<label class="campo"
					><span class="rotulo">Vence a los (días)</span><input
						class="input"
						type="number"
						inputmode="numeric"
						min="1"
						max="90"
						value={pr.dias ?? 7}
						onchange={(e) => {
							const v = Math.min(90, Math.max(1, Math.round(+e.currentTarget.value || 7)));
							e.currentTarget.value = String(v);
							poner({ dias: v });
						}}
					/></label
				>
			</div>
			<span class="ayuda">Cada persona recibe un código propio, de un solo uso.</span>
		{:else if pr.tipo === 'descarga'}
			<CampoDescarga
				archivo={(pr.archivo as string) ?? ''}
				nombreArchivo={(pr.nombreArchivo as string) ?? ''}
				onvalor={(v) => poner(v)}
			/>
			<span class="ayuda">Se muestra al terminar y se manda por mail.</span>
		{/if}
	</fieldset>
{:else if campo.tipo === 'opciones'}
	{@const elegidas = Array.isArray(valor) ? (valor as string[]) : []}
	<fieldset class="campo grupo-campo">
		<legend class="rotulo">{campo.nombre}</legend>
		<div class="casillas">
			{#each campo.opciones ?? [] as o, i (o)}
				<label class="casilla"
					><input
						type="checkbox"
						checked={elegidas.includes(o)}
						onchange={(e) =>
							onvalor(
								e.currentTarget.checked ? [...elegidas, o] : elegidas.filter((x) => x !== o),
								true
							)}
					/>{campo.nombres?.[i] ?? o}</label
				>
			{/each}
		</div>
		{#if campo.descripcion}<span class="ayuda">{campo.descripcion}</span>{/if}
	</fieldset>
{:else if campo.tipo === 'tabla'}
	<CampoTabla nombre={campo.nombre} {valor} onvalor={(v, c) => onvalor(v, c)} />
{:else if campo.tipo === 'lista'}
	<fieldset class="campo grupo-campo">
		<legend class="rotulo">{campo.nombre}</legend>
		{#each lista as x, i (i)}
			<div class="item" class:abierto={abierto === i}>
				<div class="item-cab">
					<button
						type="button"
						class="item-t"
						aria-expanded={abierto === i}
						onclick={() => (abierto = abierto === i ? null : i)}
						><Icon icon={abierto === i ? I.phCaretDownBold : I.phCaretRightBold} size={12} /><span
							>{resumen(x, i)}</span
						></button
					>
					<button
						type="button"
						class="mini-btn icono"
						aria-label="Subir {resumen(x, i)}"
						disabled={i === 0}
						onclick={() => moverItem(i)}><Icon icon={I.phCaretUpBold} size={12} /></button
					>
					<button
						type="button"
						class="mini-btn icono"
						aria-label="Quitar {resumen(x, i)}"
						onclick={() => {
							abierto = null;
							onvalor(
								lista.filter((_, k) => k !== i),
								true
							);
						}}><Icon icon={I.phTrash} size={13} /></button
					>
				</div>
				{#if abierto === i}
					<div class="hijos">
						{#each campo.item ?? [] as c (c.id)}
							<Self
								campo={c}
								valor={x[c.id]}
								{ayuda}
								onvalor={(v, conf) => onvalor(itemCon(i, c.id, v), conf)}
							/>
						{/each}
					</div>
				{/if}
			</div>
		{/each}
		{#if lista.length < max}
			<button
				type="button"
				class="agregar"
				onclick={() => {
					// El primer campo obligatorio con algo para empezar (un ícono, una categoría).
					const nuevo: Record<string, unknown> = {};
					for (const c of campo.item ?? []) {
						if (c.tipo === 'icono') nuevo[c.id] = ICONOS[lista.length % ICONOS.length];
						if (c.tipo === 'categoria' && ayuda.cx.categorias.length)
							nuevo[c.id] = ayuda.cx.categorias[lista.length % ayuda.cx.categorias.length];
					}
					abierto = lista.length;
					onvalor([...lista, nuevo], false);
				}}><Icon icon={I.phPlusBold} size={13} />Agregar</button
			>
		{:else}
			<p class="ayuda">Van hasta {max}.</p>
		{/if}
	</fieldset>
{/if}

<style>
	.campo {
		display: flex;
		flex-direction: column;
		gap: 6px;
		min-width: 0;
		margin: 0;
		padding: 0;
		border: 0;
	}
	.rotulo {
		font-size: 13px;
		font-weight: 600;
		padding: 0;
	}
	.req {
		font-weight: 500;
		color: var(--fa);
	}
	.cuenta,
	.ayuda {
		font-size: 12px;
		color: var(--mu);
		margin: 0;
	}
	.cuenta {
		align-self: flex-end;
	}
	.error {
		font-size: 12.5px;
		font-weight: 500;
		color: var(--error);
	}
	textarea.textarea {
		resize: vertical;
		font: inherit;
		font-size: 15px;
	}
	select.input {
		appearance: auto;
	}
	.grupo-campo {
		gap: 8px;
	}
	.fondos,
	.opciones {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 6px;
	}
	.fondo,
	.opcion {
		display: flex;
		align-items: center;
		gap: 8px;
		min-height: 44px;
		padding: 6px 10px;
		border: 1px solid var(--ln2);
		border-radius: 12px;
		background: var(--sf);
		color: var(--tx);
		font: inherit;
		font-size: 13px;
		font-weight: 600;
		text-align: left;
		cursor: pointer;
	}
	.fondo[aria-checked='true'],
	.opcion[aria-checked='true'],
	.icono-op[aria-checked='true'] {
		border-color: var(--terracota);
		box-shadow: 0 0 0 1px var(--terracota);
	}
	.muestra {
		width: 18px;
		height: 18px;
		flex: none;
		border-radius: 6px;
		box-shadow: inset 0 0 0 1px rgba(0, 0, 0, 0.15);
	}
	.muestra.normal {
		background: var(--bg);
	}
	.muestra.superficie {
		background: var(--sf2);
	}
	.muestra.acento {
		background: var(--terracota);
	}
	.muestra.oscuro {
		background: #2a140c;
	}
	.numero {
		display: flex;
		align-items: center;
		gap: 6px;
	}
	.numero .input {
		width: 72px;
		text-align: center;
	}
	.fila-si {
		flex-direction: row;
		align-items: center;
		justify-content: space-between;
		min-height: 44px;
	}
	.iconos {
		display: grid;
		grid-template-columns: repeat(5, minmax(0, 1fr));
		gap: 6px;
	}
	.icono-op {
		display: grid;
		place-items: center;
		min-height: 44px;
		border: 1px solid var(--ln2);
		border-radius: 12px;
		background: var(--sf);
		color: var(--tx);
		cursor: pointer;
	}
	.check {
		display: flex;
		align-items: center;
		gap: 10px;
		min-height: 40px;
		font-size: 14px;
	}
	.check input {
		width: 18px;
		height: 18px;
		accent-color: var(--terracota);
	}
	.sub > summary {
		cursor: pointer;
		min-height: 36px;
		display: flex;
		align-items: center;
	}
	.hijos {
		display: flex;
		flex-direction: column;
		gap: 12px;
		padding: 10px 0 4px;
	}
	.item {
		border: 1px solid var(--ln);
		border-radius: 14px;
		padding: 4px 6px;
	}
	.item.abierto {
		padding-bottom: 10px;
	}
	.item .hijos {
		padding: 6px 4px 0;
	}
	.item-cab {
		display: flex;
		align-items: center;
		gap: 4px;
	}
	.item-t {
		flex: 1;
		min-width: 0;
		display: flex;
		align-items: center;
		gap: 8px;
		min-height: 40px;
		padding: 0 4px;
		border: 0;
		background: none;
		color: inherit;
		font: inherit;
		font-size: 13.5px;
		font-weight: 600;
		text-align: left;
		cursor: pointer;
	}
	.item-t span {
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.agregar {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 6px;
		min-height: 40px;
		border: 1px dashed var(--ln2);
		border-radius: 12px;
		background: none;
		color: var(--tx);
		font: inherit;
		font-size: 13px;
		font-weight: 600;
		cursor: pointer;
	}
	.mini-btn {
		display: inline-grid;
		place-items: center;
		width: 36px;
		min-height: 36px;
		flex: none;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: var(--sf);
		color: var(--tx);
		cursor: pointer;
	}
	.mini-btn:disabled {
		opacity: 0.35;
	}
	@media (max-width: 760px) {
		.mini-btn {
			width: 44px;
			min-height: 44px;
		}
	}
	.dos {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
		gap: 8px;
	}
	.casillas {
		display: flex;
		flex-wrap: wrap;
		gap: 4px 14px;
	}
	.casilla {
		display: flex;
		align-items: center;
		gap: 8px;
		min-height: 36px;
		font-size: 13.5px;
		font-weight: 500;
		cursor: pointer;
	}
	.casilla input {
		width: 18px;
		height: 18px;
		margin: 0;
		accent-color: var(--terracota);
	}
	.subir-archivo {
		align-self: flex-start;
		cursor: pointer;
	}
</style>
