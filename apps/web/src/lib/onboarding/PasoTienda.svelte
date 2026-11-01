<script lang="ts">
	import { page } from '$app/state';
	import Icon from '#lib/components/Icon.svelte';
	import Interruptor from '#lib/components/Interruptor.svelte';
	import TplMini from '#lib/components/TplMini.svelte';
	import * as I from '#lib/icons.ts';
	import RUBROS from '@mostrador/contrato/rubros.json';
	import { PLANTILLAS, plantilla } from '#lib/plantillas.ts';
	import { DE_INSTAGRAM, PARA_EMPEZAR, RECOMENDADAS, SECCIONES } from './datos.ts';
	import Etiqueta from './Etiqueta.svelte';
	import { getOnboarding } from './estado.svelte.ts';
	import Galeria from './Galeria.svelte';
	import Miniatura from './Miniatura.svelte';

	/** Paso 5, Tu tienda: plantilla (con la vista previa), colores de Instagram, secciones y dominio. */
	const o = getOnboarding();
	// Con la API, la tienda de verdad; en la demo, la del handoff.
	const real = $derived(page.data.tienda as { nombre: string; slug: string } | null);
	const base = $derived(plantilla(o.e.plantilla) ?? PLANTILLAS[0]);
	// Con "Usar mis colores y logo", la vista previa toma los colores, la foto y la frase de Instagram.
	const actual = $derived({
		...base,
		tienda: real?.nombre ?? 'Café La Esquina',
		...(o.e.instagram
			? {
					img: `https://images.unsplash.com/photo-${DE_INSTAGRAM.img}?auto=format&fit=crop&w=900&q=75`,
					lema: DE_INSTAGRAM.tag
				}
			: {})
	});
	const paleta = $derived(
		o.e.instagram ? { bg: DE_INSTAGRAM.bg, tx: DE_INSTAGRAM.ink, ac: DE_INSTAGRAM.acc } : null
	);
	// En la demo, las del rubro del handoff; en una tienda real todavía no sabemos el rubro (el registro
	// no lo pregunta): tres estilos bien distintos para empezar, sin decir que son de su rubro.
	// Con el rubro elegido en «Tu negocio», su plantilla primero y dos estilos distintos al lado.
	const delRubro = $derived(
		o.demo ? null : (RUBROS.find((r) => r.id === o.e.rubro)?.plantilla ?? null)
	);
	const recomendadas = $derived(
		(o.demo
			? [...RECOMENDADAS]
			: delRubro
				? [delRubro, ...PARA_EMPEZAR.filter((id) => id !== delRubro)].slice(0, 3)
				: [...PARA_EMPEZAR]
		)
			.map((id) => plantilla(id))
			.filter((t) => !!t)
	);
	const mob = $derived(o.dispositivo === 'mob');

	// Se anima cada vez que cambia (como en el handoff).
	let marco = $state<HTMLElement>();
	$effect(() => {
		void [o.e.plantilla, o.e.instagram];
		marco?.animate(
			[
				{ opacity: 0.3, transform: 'scale(.98)' },
				{ opacity: 1, transform: 'none' }
			],
			{ duration: 420, easing: 'cubic-bezier(.2,.7,.1,1)' }
		);
	});
</script>

<div class="cab">
	<div class="cab-t">
		<Etiqueta
			rotulo="Tu tienda"
			titulo={o.demo
				? 'Elegí cómo se ve. Ya tiene tus productos adentro.'
				: 'Elegí cómo se ve tu tienda.'}
		/>
	</div>
	<div class="disp" role="group" aria-label="Dispositivo">
		<button type="button" aria-pressed={!mob} onclick={() => (o.dispositivo = 'desk')}
			><Icon icon={I.phDesktop} size={15} />Escritorio</button
		>
		<button type="button" aria-pressed={mob} onclick={() => (o.dispositivo = 'mob')}
			><Icon icon={I.phDeviceMobile} size={15} />Celular</button
		>
	</div>
</div>
<div class="cuerpo">
	<div class="escenario">
		<div class="marco" class:mob bind:this={marco}>
			<div class="pantalla">
				{#if !mob}
					<div class="barra">
						<span style:background="#E0A595"></span><span style:background="#E7CB8F"></span><span
							style:background="#A9CBA8"
						></span>
						<!-- El dominio entero y el nombre recortado con "…" si no entra (no desborda). -->
						<span class="url" title={real ? `${real.slug}.tulocalweb.com.ar` : undefined}
							>{#if real}<span class="u-slug">{real.slug}</span><span class="u-dom"
									>.tulocalweb.com.ar</span
								>{:else}laesquina.com.ar{/if}</span
						>
					</div>
				{/if}
				<div class="zoom" class:mob><TplMini t={actual} {paleta} /></div>
			</div>
		</div>
	</div>
	<div class="lado">
		<div class="caja">
			<div class="c-cab">
				<Icon icon={I.phSparkleFill} size={15} style="color:#b0441d" /><span
					>{o.demo || delRubro ? 'Recomendadas para tu rubro' : 'Para empezar'}</span
				>
			</div>
			<div class="recos">
				{#each recomendadas as t (t.id)}
					<button
						type="button"
						onclick={() => ((o.e.plantilla = t.id), o.guardar())}
						aria-pressed={o.e.plantilla === t.id}
					>
						<div class="reco" class:sel={o.e.plantilla === t.id}>
							<div class="reco-in"><Miniatura {t} /></div>
						</div>
						<div class="reco-n">{t.nombre}</div>
					</button>
				{/each}
			</div>
			<button type="button" class="ver" onclick={() => (o.galeria = true)}
				><Icon icon={I.phSquaresFour} size={15} />Ver las {PLANTILLAS.length} plantillas</button
			>
		</div>
		{#if o.demo}
			<div class="caja">
				<div class="ig">
					<Icon icon={I.logosInstagramIcon} size={20} />
					<span class="ig-t"
						><span>Usar mis colores y logo</span><span class="ig-d">Sacados de @cafelaesquina</span
						></span
					>
					<Interruptor
						bind:checked={o.e.instagram}
						label="Usar mis colores y logo"
						onchange={() => o.guardar()}
					/>
				</div>
				<div class="muestras">
					{#each ['#F3E9DC', '#6B3A22', '#2F5D50', '#3B2416'] as c (c)}<span style:background={c}
						></span>{/each}
				</div>
			</div>
		{/if}
		<div class="caja">
			<div class="sec-t">Secciones que armamos</div>
			<div class="secs">
				{#each SECCIONES as [k, ico, t] (k)}
					<button
						type="button"
						aria-pressed={o.e.secciones[k]}
						onclick={() => ((o.e.secciones[k] = !o.e.secciones[k]), o.guardar())}
						><Icon icon={ico} size={13} />{t}</button
					>
				{/each}
			</div>
		</div>
		{#if o.demo}
			<div class="caja dominio">
				<Icon icon={I.phCheckCircleFill} size={18} style="color:#237550;flex:none" />
				<span
					><span class="d-t">laesquina.com.ar</span><span class="d-d"
						>Disponible · lo reservamos por vos</span
					></span
				>
			</div>
		{/if}
	</div>
</div>
<Galeria />

<style>
	.cab {
		display: flex;
		align-items: end;
		gap: 16px;
		flex-wrap: wrap;
	}
	.cab-t {
		flex: 1 1 420px;
	}
	.disp {
		display: flex;
		padding: 3px;
		border-radius: 999px;
		background: rgba(58, 29, 18, 0.07);
	}
	.disp button {
		display: flex;
		align-items: center;
		gap: 6px;
		border: 0;
		font-size: 13px;
		font-weight: 600;
		padding: 8px 14px;
		border-radius: 999px;
		background: transparent;
		color: #2a140c;
	}
	.disp button[aria-pressed='true'] {
		background: #fff;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.12);
	}
	.cuerpo {
		margin-top: 22px;
		display: flex;
		gap: 20px;
		flex-wrap: wrap;
		align-items: flex-start;
	}
	.escenario {
		flex: 1 1 480px;
		min-width: 0;
		min-height: 520px;
		border-radius: 24px;
		background: #efe6da;
		display: grid;
		place-items: center;
		padding: 24px;
	}
	.marco {
		width: 100%;
		height: 470px;
		border-radius: 14px;
		padding: 0;
		background: transparent;
		box-shadow: 0 30px 70px -30px rgba(42, 20, 12, 0.6);
		transition:
			width 0.45s cubic-bezier(0.2, 0.7, 0.1, 1),
			height 0.45s cubic-bezier(0.2, 0.7, 0.1, 1);
	}
	.marco.mob {
		width: 260px;
		height: 520px;
		border-radius: 38px;
		padding: 9px;
		background: #2a140c;
	}
	.pantalla {
		width: 100%;
		height: 100%;
		border-radius: 14px;
		overflow: hidden;
		background: #fff;
		display: flex;
		flex-direction: column;
	}
	.mob .pantalla {
		border-radius: 30px;
	}
	.barra {
		height: 24px;
		flex: none;
		display: flex;
		align-items: center;
		gap: 5px;
		padding: 0 10px;
		background: #f1e9df;
	}
	.barra > span:not(.url) {
		width: 7px;
		height: 7px;
		border-radius: 50%;
	}
	.url {
		flex: 1;
		min-width: 0;
		display: flex;
		justify-content: center;
		margin: 0 clamp(8px, 6%, 40px);
		padding: 2px 10px;
		border-radius: 999px;
		background: #fff;
		font-size: 10px;
		color: #5c4536;
		white-space: nowrap;
		overflow: hidden;
	}
	.u-dom {
		flex: none;
	}
	.u-slug {
		min-width: 0;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.zoom {
		flex: 1;
		min-height: 0;
		zoom: 1.55;
	}
	.zoom.mob {
		zoom: 1;
	}
	.lado {
		flex: 0 1 320px;
		min-width: 0;
		display: flex;
		flex-direction: column;
		gap: 14px;
	}
	.caja {
		padding: 16px;
		border-radius: 20px;
		background: #fff;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.06);
	}
	.c-cab {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 14px;
		font-weight: 600;
	}
	.recos {
		margin-top: 12px;
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 8px;
	}
	.recos button {
		padding: 0;
		border: 0;
		background: transparent;
		text-align: left;
		color: #2a140c;
		min-width: 0;
	}
	.reco {
		height: 112px;
		border-radius: 12px;
		overflow: hidden;
		box-shadow:
			0 0 0 2px transparent,
			0 8px 18px -10px rgba(58, 29, 18, 0.5);
	}
	.reco.sel {
		box-shadow:
			0 0 0 2px #d2582f,
			0 8px 18px -10px rgba(58, 29, 18, 0.5);
	}
	.reco-in {
		width: 200%;
		height: 224px;
		zoom: 0.5;
	}
	.reco-n {
		margin-top: 6px;
		font-size: 12.5px;
		font-weight: 600;
	}
	.ver {
		min-height: 44px;
		margin-top: 12px;
		width: 100%;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		padding: 11px;
		border-radius: 999px;
		border: 1px solid rgba(58, 29, 18, 0.16);
		background: #fff;
		font-size: 13.5px;
		font-weight: 600;
		color: #2a140c;
	}
	.ig {
		display: flex;
		align-items: center;
		gap: 12px;
	}
	.ig-t {
		flex: 1;
		display: flex;
		flex-direction: column;
		font-size: 14px;
		font-weight: 600;
	}
	.ig-d {
		font-size: 12.5px;
		font-weight: 400;
		color: #7a6252;
	}
	.muestras {
		margin-top: 12px;
		display: flex;
		gap: 6px;
	}
	.muestras span {
		width: 26px;
		height: 26px;
		border-radius: 8px;
		box-shadow: inset 0 0 0 1px rgba(0, 0, 0, 0.08);
	}
	.sec-t {
		font-size: 14px;
		font-weight: 600;
	}
	.secs {
		margin-top: 10px;
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
	}
	.secs button {
		display: flex;
		align-items: center;
		gap: 5px;
		padding: 6px 10px;
		border-radius: 999px;
		border: 1px solid rgba(58, 29, 18, 0.14);
		background: #fff;
		font-size: 12.5px;
		font-weight: 500;
		color: #2a140c;
	}
	.secs button[aria-pressed='true'] {
		background: #fcede5;
		border-color: rgba(210, 88, 47, 0.45);
	}
	.dominio {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 14px 16px;
	}
	.d-t {
		display: block;
		font-size: 14px;
		font-weight: 600;
	}
	.d-d {
		display: block;
		font-size: 12.5px;
		color: #7a6252;
	}
	/* Celular: el escenario se achica y en "Escritorio" la tienda se ve a escala. */
	@media (max-width: 640px) {
		.escenario {
			min-height: 0;
			padding: 14px;
			border-radius: 20px;
		}
		.marco {
			height: 380px;
		}
		.zoom {
			zoom: 1;
		}
		.lado {
			flex: 1 1 100%;
		}
	}
	@media (max-width: 559px) {
		.disp button,
		.secs button {
			min-height: 44px;
		}
	}
</style>
