<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import {
		CONTENIDO,
		DISPARO,
		FRECUENCIA,
		queFaltaVentana,
		VENTANAS,
		type ContenidoVentana,
		type Disparo,
		type Frecuencia,
		type Ventana
	} from '#lib/tienda/sistema/promos.ts';
	import CampoBloque from './CampoBloque.svelte';
	import type { Ayuda } from './campos.ts';
	import type { CampoConNombres } from './catalogo.ts';
	import { conContenido } from './ventanas.ts';

	/**
	 * El panel de una ventana (Mi tienda › Promociones): el contenido y sus campos (del catálogo), el
	 * diseño con miniaturas, cuándo sale, a quién, dónde y las fechas; lo que le falta para salir y
	 * sus números. «Ver en la vista previa» la abre con el borrador.
	 */
	let {
		ventana,
		ayuda,
		numeros = null,
		onventana,
		onvolver,
		onver,
		onduplicar,
		onborrar,
		puedeDuplicar
	}: {
		ventana: Ventana;
		ayuda: Ayuda;
		numeros?: {
			vistas: number;
			envios: number;
			clics: number;
			ventas: number;
			monto: number;
		} | null;
		onventana: (w: Ventana, confirmar: boolean) => void;
		onvolver: () => void;
		onver: () => void;
		onduplicar: () => void;
		onborrar: () => void;
		puedeDuplicar: boolean;
	} = $props();

	const w = $derived(ventana);
	const def = $derived(CONTENIDO[w.contenido]);
	const falta = $derived(queFaltaVentana(w));
	const poner = (cambios: Partial<Ventana>, confirmar = true) =>
		onventana({ ...w, ...cambios }, confirmar);
	const ponerDato = (id: string, v: unknown, confirmar: boolean) => {
		const datos = { ...w.datos };
		if (v === '' || v === undefined || v === null) delete datos[id];
		else datos[id] = v;
		poner({ datos }, confirmar);
	};

	// Campos armados con las opciones del catálogo, para los mismos controles que los bloques.
	const C = (c: CampoConNombres) => c;
	const campoNombre = C({
		id: 'nombre',
		tipo: 'texto',
		nombre: 'Nombre (sólo lo ves vos)',
		max: 40
	});
	const campoPaginas = C({
		id: 'paginas',
		tipo: 'opciones',
		nombre: 'En qué páginas',
		opciones: VENTANAS.paginas.opciones,
		nombres: VENTANAS.paginas.nombres,
		descripcion: VENTANAS.paginas.descripcion
	});
	const campoPublico = C({
		id: 'publico',
		tipo: 'opcion',
		nombre: 'A quién',
		opciones: VENTANAS.publicos.opciones,
		nombres: VENTANAS.publicos.nombres
	});
	const campoDispositivos = C({
		id: 'dispositivos',
		tipo: 'opcion',
		nombre: 'En qué dispositivos',
		opciones: VENTANAS.dispositivos.opciones,
		nombres: VENTANAS.dispositivos.nombres
	});
	const dd = $derived(DISPARO[w.disparo.cuando]);
	const ff = $derived(FRECUENCIA[w.frecuencia.cada]);
	const pct = (a: number, b: number) => (b ? `${Math.round((a / b) * 100)}\u00a0%` : '—');
</script>

<div class="panel-ventana">
	<div class="pb-cab">
		<button type="button" class="volver" onclick={onvolver}
			><Icon icon={I.phCaretLeftBold} size={13} />Ventanas</button
		>
		<div class="pb-acciones">
			<button
				type="button"
				class="mini-btn icono"
				aria-label="Duplicar {w.nombre}"
				title="Duplicar"
				disabled={!puedeDuplicar}
				onclick={onduplicar}><Icon icon={I.phCopy} size={15} /></button
			>
			<button
				type="button"
				class="mini-btn icono"
				aria-label="Borrar {w.nombre}"
				title="Borrar"
				onclick={onborrar}><Icon icon={I.phTrash} size={15} /></button
			>
		</div>
	</div>
	<h2 class="pb-t">{w.nombre}</h2>
	<button type="button" class="ver" onclick={onver}
		><Icon icon={I.phEye} size={15} />Ver en la vista previa</button
	>
	{#if !w.activa}
		<p class="aviso info"><Icon icon={I.phEyeSlash} size={15} />Está pausada: no sale.</p>
	{:else if falta}
		<p class="aviso" role="status">
			<Icon icon={I.phWarningFill} size={15} />No sale todavía. {falta}
		</p>
	{/if}
	{#if numeros && numeros.vistas}
		<dl class="numeros">
			<div>
				<dt>Vistas</dt>
				<dd>{numeros.vistas.toLocaleString('es-AR')}</dd>
			</div>
			{#if w.contenido === 'captacion' || w.contenido === 'ruleta'}
				<div>
					<dt>Dejaron el mail</dt>
					<dd>{pct(numeros.envios, numeros.vistas)}</dd>
				</div>
			{:else}
				<div>
					<dt>Tocaron</dt>
					<dd>{pct(numeros.clics, numeros.vistas)}</dd>
				</div>
			{/if}
			<div>
				<dt>Ventas</dt>
				<dd>{numeros.ventas.toLocaleString('es-AR')}</dd>
			</div>
		</dl>
	{/if}

	<div class="campos">
		<CampoBloque
			campo={campoNombre}
			valor={w.nombre}
			{ayuda}
			onvalor={(v, c) => poner({ nombre: (v as string) ?? '' }, c)}
		/>
	</div>

	<div class="sub">Qué muestra</div>
	<div class="contenidos" role="radiogroup" aria-label="Qué muestra">
		{#each VENTANAS.contenidos as c (c.tipo)}
			<button
				type="button"
				role="radio"
				aria-checked={w.contenido === c.tipo}
				class="contenido"
				onclick={() => onventana(conContenido(w, c.tipo as ContenidoVentana), true)}
				><b>{c.nombre}</b><span>{c.descripcion}</span></button
			>
		{/each}
	</div>
	<div class="campos">
		{#each def.campos as c (c.id)}
			<CampoBloque
				campo={c}
				valor={w.datos[c.id]}
				{ayuda}
				onvalor={(v, conf) => ponerDato(c.id, v, conf)}
			/>
		{/each}
	</div>

	<div class="sub">Diseño</div>
	<div class="variantes" role="radiogroup" aria-label="Diseño de la ventana">
		{#each VENTANAS.variantes as v, i (v)}
			<button
				type="button"
				role="radio"
				aria-checked={w.variante === v}
				class="variante"
				onclick={() => poner({ variante: v })}
			>
				<span class="mini mini--{v}" aria-hidden="true"><span></span></span>
				<b>{VENTANAS.nombresVariantes[i]}</b>
				<small>{VENTANAS.descripcionesVariantes[i]}</small>
			</button>
		{/each}
	</div>

	<div class="sub">Cuándo sale</div>
	<div class="campos">
		<label class="campo">
			<span class="rotulo">Momento</span>
			<select
				class="input"
				value={w.disparo.cuando}
				onchange={(e) => {
					const d = DISPARO[e.currentTarget.value as Disparo];
					poner({
						disparo: { cuando: d.cuando, ...(d.valor ? { valor: d.valor.porDefecto } : {}) }
					});
				}}
			>
				{#each VENTANAS.disparos as d (d.cuando)}<option value={d.cuando}>{d.nombre}</option>{/each}
			</select>
			<span class="ayuda">{dd.descripcion}</span>
		</label>
		{#if dd.valor}
			<CampoBloque
				campo={{
					id: 'valor',
					tipo: 'numero',
					nombre: dd.valor.unidad === '%' ? 'Porcentaje de la página' : 'Segundos',
					min: dd.valor.min,
					max: dd.valor.max,
					porDefecto: dd.valor.porDefecto
				}}
				valor={w.disparo.valor}
				{ayuda}
				onvalor={(v, c) => poner({ disparo: { ...w.disparo, valor: v as number } }, c)}
			/>
		{/if}
		<label class="campo">
			<span class="rotulo">Cada cuánto</span>
			<select
				class="input"
				value={w.frecuencia.cada}
				onchange={(e) => {
					const f = FRECUENCIA[e.currentTarget.value as Frecuencia];
					poner({ frecuencia: { cada: f.cada, ...(f.dias ? { dias: f.dias.porDefecto } : {}) } });
				}}
			>
				{#each VENTANAS.frecuencias as f (f.cada)}<option value={f.cada}>{f.nombre}</option>{/each}
			</select>
			<span class="ayuda">{ff.descripcion}</span>
		</label>
		{#if ff.dias}
			<CampoBloque
				campo={{
					id: 'dias',
					tipo: 'numero',
					nombre: 'Días',
					min: ff.dias.min,
					max: ff.dias.max,
					porDefecto: ff.dias.porDefecto
				}}
				valor={w.frecuencia.dias}
				{ayuda}
				onvalor={(v, c) => poner({ frecuencia: { ...w.frecuencia, dias: v as number } }, c)}
			/>
		{/if}
	</div>

	<div class="sub">A quién y dónde</div>
	<div class="campos">
		<CampoBloque
			campo={campoPublico}
			valor={w.publico}
			{ayuda}
			onvalor={(v) => poner({ publico: v as Ventana['publico'] })}
		/>
		<CampoBloque
			campo={campoPaginas}
			valor={w.paginas}
			{ayuda}
			onvalor={(v) => poner({ paginas: v as Ventana['paginas'] })}
		/>
		<CampoBloque
			campo={campoDispositivos}
			valor={w.dispositivos}
			{ayuda}
			onvalor={(v) => poner({ dispositivos: v as Ventana['dispositivos'] })}
		/>
	</div>

	<div class="sub">Fechas</div>
	<div class="campos">
		<CampoBloque
			campo={{ id: 'desde', tipo: 'fecha', nombre: 'Desde (opcional)' }}
			valor={w.desde}
			{ayuda}
			onvalor={(v) => poner({ desde: (v as string) || undefined })}
		/>
		<CampoBloque
			campo={{ id: 'hasta', tipo: 'fecha', nombre: 'Hasta (opcional)' }}
			valor={w.hasta}
			{ayuda}
			onvalor={(v) => poner({ hasta: (v as string) || undefined })}
		/>
	</div>
</div>

<style>
	.panel-ventana {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.pb-cab {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 8px;
	}
	.volver {
		display: flex;
		align-items: center;
		gap: 6px;
		min-height: 40px;
		padding: 0 10px 0 4px;
		border: 0;
		background: none;
		color: var(--mu);
		font: inherit;
		font-size: 13.5px;
		font-weight: 600;
		cursor: pointer;
	}
	.pb-acciones {
		display: flex;
		gap: 4px;
	}
	.pb-t {
		margin: 4px 4px 0;
		font-size: 19px;
		font-weight: 600;
		letter-spacing: -0.02em;
		overflow-wrap: anywhere;
	}
	.ver {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		min-height: 40px;
		margin: 4px 0;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: var(--sf);
		color: var(--tx);
		font: inherit;
		font-size: 13.5px;
		font-weight: 600;
		cursor: pointer;
	}
	.aviso {
		display: flex;
		gap: 8px;
		align-items: flex-start;
		margin: 4px 0;
		padding: 10px 12px;
		border-radius: 12px;
		background: color-mix(in srgb, var(--alerta) 16%, transparent);
		color: var(--tx);
		font-size: 13px;
		font-weight: 500;
		text-wrap: pretty;
	}
	.aviso :global(svg) {
		flex: none;
		margin-top: 1px;
		color: #b7791f;
	}
	.aviso.info {
		background: var(--chip);
	}
	.aviso.info :global(svg) {
		color: var(--mu);
	}
	.numeros {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 6px;
		margin: 2px 0;
	}
	.numeros div {
		padding: 8px 10px;
		border-radius: 12px;
		background: var(--chip);
	}
	.numeros dt {
		font-size: 11.5px;
		color: var(--mu);
	}
	.numeros dd {
		margin: 2px 0 0;
		font-size: 16px;
		font-weight: 600;
		font-variant-numeric: tabular-nums;
	}
	.sub {
		font-size: 12.5px;
		font-weight: 600;
		color: var(--mu);
		padding: 12px 4px 2px;
	}
	.contenidos {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 6px;
	}
	.contenido {
		display: flex;
		flex-direction: column;
		gap: 2px;
		padding: 10px;
		border: 2px solid var(--ln);
		border-radius: 12px;
		background: var(--sf);
		color: var(--tx);
		font: inherit;
		text-align: left;
		cursor: pointer;
	}
	.contenido b {
		font-size: 13.5px;
	}
	.contenido span {
		font-size: 12px;
		color: var(--mu);
		text-wrap: pretty;
	}
	.contenido[aria-checked='true'] {
		border-color: var(--terracota);
	}
	.variantes {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 8px;
	}
	.variante {
		display: flex;
		flex-direction: column;
		gap: 4px;
		padding: 6px;
		border: 2px solid transparent;
		border-radius: 12px;
		background: none;
		color: var(--tx);
		font: inherit;
		text-align: left;
		cursor: pointer;
	}
	.variante b {
		font-size: 12.5px;
	}
	.variante small {
		font-size: 11.5px;
		color: var(--mu);
		text-wrap: pretty;
	}
	.variante[aria-checked='true'] {
		border-color: var(--terracota);
	}
	/* Miniaturas: la tienda en gris claro y la ventana en su lugar. */
	.mini {
		position: relative;
		display: block;
		height: 64px;
		border-radius: 8px;
		background:
			linear-gradient(var(--ln) 0 0) 8px 8px / 40% 6px no-repeat,
			linear-gradient(var(--ln) 0 0) 8px 20px / 70% 18px no-repeat,
			var(--sf2, #f4efe9);
		overflow: hidden;
	}
	.mini span {
		position: absolute;
		background: var(--sf);
		border: 1.5px solid var(--terracota);
		border-radius: 4px;
		box-shadow: 0 4px 10px -6px rgb(0 0 0 / 0.4);
	}
	.mini--centrada {
		box-shadow: inset 0 0 0 100px rgb(0 0 0 / 0.12);
	}
	.mini--centrada span {
		inset: 14px 28%;
	}
	.mini--lateral span {
		left: 6px;
		bottom: 6px;
		width: 38%;
		height: 26px;
	}
	.mini--completa span {
		inset: 4px;
	}
	.mini--inferior {
		box-shadow: inset 0 0 0 100px rgb(0 0 0 / 0.12);
	}
	.mini--inferior span {
		left: 10%;
		right: 10%;
		bottom: -2px;
		height: 30px;
		border-radius: 6px 6px 0 0;
	}
	.campos {
		display: flex;
		flex-direction: column;
		gap: 14px;
		padding: 6px 4px 4px;
	}
	.campo {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.rotulo {
		font-size: 13px;
		font-weight: 600;
	}
	.ayuda {
		font-size: 12px;
		color: var(--mu);
		text-wrap: pretty;
	}
	.mini-btn {
		display: inline-grid;
		place-items: center;
		width: 36px;
		min-height: 36px;
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
		.ver {
			min-height: 44px;
		}
	}
</style>
