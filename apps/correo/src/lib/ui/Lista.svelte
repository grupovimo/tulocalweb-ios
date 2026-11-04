<script lang="ts">
	import type { Casilla, Fila } from '#lib/server/bandeja.ts';
	import { grupo, hora, iniciales } from '#lib/formato.ts';
	import { familia, nombre } from '#lib/etiquetas.ts';

	let {
		filas,
		pestana,
		contadores,
		activa,
		href,
		hrefPestana,
		onrefrescar,
		vacio,
		casillas = [],
		casilla = null,
		viendo = '',
		papelera = false,
		onvaciar
	}: {
		filas: Fila[];
		pestana: string;
		contadores: { sinleer: number };
		activa: string | null;
		href: (f: Fila) => string;
		hrefPestana: (p: string) => string;
		onrefrescar: () => void;
		vacio: string;
		casillas?: Casilla[];
		casilla?: string | null;
		viendo?: string;
		papelera?: boolean;
		onvaciar?: () => void;
	} = $props();

	// Con varias casillas juntas, cada fila dice de cuál es (las compartidas se mezclaban), con un
	// color fijo por casilla.
	const conEtiqueta = $derived(!casilla && casillas.length > 1);
	const COLORES = [
		'#2F6BFF',
		'#C2410C',
		'#15803D',
		'#7C3AED',
		'#B4475B',
		'#0E7490',
		'#A16207',
		'#4D7C0F'
	];
	const casillaDe = $derived(
		new Map(
			casillas.map((c, i) => [
				c.cuenta,
				{ nombre: c.personal ? 'Vos' : c.nombre, color: COLORES[i % COLORES.length] }
			])
		)
	);

	const PEST = $derived([
		{ id: 'todos', t: 'Todos', n: 0 },
		{ id: 'sinleer', t: 'Sin leer', n: contadores.sinleer }
	]);
	const conGrupo = $derived(
		filas.map((f, i) => {
			const g = grupo(f.fecha);
			return { f, g: i === 0 || grupo(filas[i - 1].fecha) !== g ? g : '' };
		})
	);
</script>

<section class="lista">
	{#if viendo || papelera}
		<div class="viendo">
			<i class="icon-inbox"></i><span class="v-t">{viendo}</span>
			{#if papelera && onvaciar}<button class="vaciar" onclick={onvaciar}
					><i class="icon-trash-2"></i>Vaciar papelera</button
				>{/if}
		</div>
	{/if}
	<div class="pest" role="tablist">
		{#each PEST as p (p.id)}
			<a
				class="p"
				class:act={pestana === p.id}
				role="tab"
				aria-selected={pestana === p.id}
				href={hrefPestana(p.id)}
				data-sveltekit-reset="false"
				>{p.t}{#if p.n}<span class="n">{p.n}</span>{/if}</a
			>
		{/each}
		<button class="refrescar" aria-label="Actualizar" onclick={onrefrescar}
			><i class="icon-rotate-cw"></i></button
		>
	</div>
	<div class="filas">
		{#each conGrupo as { f, g } (f.id)}
			{#if g}<div class="grupo">{g}</div>{/if}
			<a
				class="fila"
				class:act={activa === f.id}
				class:nueva={!f.leido}
				href={href(f)}
				data-sveltekit-reset="false"
			>
				<span class="av">{iniciales(f.nombre)}</span>
				<span class="cu">
					<span class="l1">
						<span class="nom">{f.nombre}</span>
						{#if f.enHilo > 1}<span class="cant">{f.enHilo}</span>{/if}
						{#if conEtiqueta && casillaDe.get(f.cuenta)}{@const c = casillaDe.get(f.cuenta)!}<span
								class="cas"
								style:--cas={c.color}
								title={f.casilla}>{c.nombre}</span
							>{/if}
						<span class="esp"></span>
						{#if f.adjuntos}<i class="icon-paperclip clip" aria-label="Con adjuntos"></i>{/if}
						{#if f.destacado}<i class="icon-star clip" aria-label="Destacado"></i>{/if}
						<span class="hora" class:sale={!!f.envio}
							>{f.envio ? `Sale ${hora(f.fecha)}` : hora(f.fecha)}</span
						>
					</span>
					<span class="asunto">{f.asunto}</span>
					<span class="l3"
						><span class="prev">{f.vista}</span>{#if f.etiquetas[0]}<span
								class="tag"
								style:background="var(--{familia(f.etiquetas[0])}Bg)"
								style:color="var(--{familia(f.etiquetas[0])}T)">{nombre(f.etiquetas[0])}</span
							>{/if}</span
					>
				</span>
			</a>
		{:else}
			<div class="vacio">
				<i class="icon-inbox"></i>
				<p>{vacio}</p>
			</div>
		{/each}
	</div>
</section>

<style>
	.lista {
		display: flex;
		flex-direction: column;
		min-height: 0;
		border-right: 1px solid var(--ls);
	}
	.pest {
		display: flex;
		align-items: center;
		gap: 4px;
		height: 56px;
		flex: none;
		padding: 0 12px 0 16px;
		border-bottom: 1px solid var(--ls);
		overflow-x: auto;
		scrollbar-width: none;
	}
	.p {
		display: inline-flex;
		align-items: center;
		gap: 7px;
		height: 56px;
		padding: 0 12px;
		white-space: nowrap;
		font-size: 16px;
		font-weight: 500;
		color: var(--mut);
		text-decoration: none;
	}
	.p.act {
		color: var(--accT);
		font-weight: 600;
		box-shadow: inset 0 -3px 0 var(--acc);
	}
	.n {
		min-width: 24px;
		height: 24px;
		padding: 0 7px;
		border-radius: 999px;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		font-size: 14px;
		font-weight: 600;
		background: var(--mu);
		color: var(--fg);
	}
	.p.act .n {
		background: var(--acc);
		color: #fff;
	}
	.refrescar {
		margin-left: auto;
		padding: 8px;
		color: var(--mut);
		font-size: 19px;
	}
	.filas {
		flex: 1;
		min-height: 0;
		overflow-y: auto;
	}
	.grupo {
		padding: 14px 20px 6px;
		font-size: 14px;
		font-weight: 600;
		color: var(--mut);
	}
	.fila {
		display: flex;
		gap: 14px;
		padding: var(--py) 18px var(--py) 16px;
		color: var(--fg);
		text-decoration: none;
	}
	.fila:hover {
		background: color-mix(in oklch, var(--sel) 60%, transparent);
	}
	.fila.act {
		background: var(--sel);
		box-shadow: inset 4px 0 0 var(--acc);
	}
	.av {
		width: 42px;
		height: 42px;
		flex: none;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 16px;
		font-weight: 600;
		background: var(--mu);
		color: var(--mut);
	}
	.cu {
		min-width: 0;
		flex: 1;
		line-height: 1.4;
		display: flex;
		flex-direction: column;
	}
	.l1 {
		display: flex;
		align-items: center;
		gap: 8px;
	}
	.nom {
		min-width: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		font-size: 17px;
		font-weight: 400;
	}
	.cant {
		font-size: 15px;
		color: var(--mut);
	}
	.esp {
		flex: 1;
	}
	.clip {
		font-size: 16px;
		color: var(--mut);
	}
	.hora {
		flex: none;
		font-size: 15px;
		color: var(--mut);
	}
	.asunto {
		font-size: 16px;
		font-weight: 400;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.l3 {
		display: flex;
		align-items: center;
		gap: 8px;
	}
	.prev {
		flex: 1;
		min-width: 0;
		font-size: 15px;
		color: var(--mut);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.tag {
		flex: none;
		white-space: nowrap;
		padding: 0 9px;
		border-radius: 999px;
		font-size: 14px;
		line-height: 24px;
		font-weight: 500;
	}
	.hora.sale {
		color: var(--espT);
		font-weight: 600;
	}
	.nueva .nom,
	.nueva .hora {
		font-weight: 700;
	}
	.nueva .asunto {
		font-weight: 600;
	}
	.nueva .hora {
		color: var(--accT);
	}
	.vacio {
		padding: 64px 24px;
		text-align: center;
		color: var(--mut);
	}
	.vacio i {
		font-size: 34px;
		color: var(--t3);
	}
	.vacio p {
		margin: 10px 0 0;
		font-size: 16px;
	}
	.viendo {
		display: flex;
		align-items: center;
		gap: 8px;
		min-height: 40px;
		padding: 6px 12px 6px 18px;
		border-bottom: 1px solid var(--ls);
		font-size: 13.5px;
		color: var(--mut);
	}
	.v-t {
		flex: 1;
		min-width: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		font-weight: 600;
	}
	.vaciar {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		height: 30px;
		padding: 0 10px;
		border: 1px solid var(--ls);
		border-radius: 999px;
		background: none;
		color: inherit;
		font: inherit;
		font-weight: 600;
		cursor: pointer;
	}
	.cas {
		flex: none;
		max-width: 40%;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		padding: 1px 7px;
		border-radius: 999px;
		font-size: 11.5px;
		font-weight: 700;
		color: var(--cas);
		background: color-mix(in srgb, var(--cas) 12%, transparent);
	}
</style>
