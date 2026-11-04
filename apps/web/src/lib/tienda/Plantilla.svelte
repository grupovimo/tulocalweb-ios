<script lang="ts">
	import type { Component, ComponentProps } from 'svelte';
	import Andes from './plantillas/Andes.svelte';
	import Ceibo from './plantillas/Ceibo.svelte';
	import Malbec from './plantillas/Malbec.svelte';
	import Miga from './plantillas/Miga.svelte';
	import MigaClasica from './plantillas/MigaClasica.svelte';
	import Motor from './plantillas/Motor.svelte';
	import Mostaza from './plantillas/Mostaza.svelte';
	import Pampa from './plantillas/Pampa.svelte';
	import Quebracho from './plantillas/Quebracho.svelte';
	import Rosedal from './plantillas/Rosedal.svelte';
	import Tijera from './plantillas/Tijera.svelte';
	import Yerbal from './plantillas/Yerbal.svelte';
	import { acentoElegido } from './acentos.ts';
	import { cssEstilo } from './estilo/css.ts';

	type Props = ComponentProps<typeof Ceibo>;

	/**
	 * Dibuja la tienda con su plantilla. Las que tienen archivo propio en el handoff (Plantilla X.dc.html)
	 * usan su diseño; las demás salen del motor de plantillas (Plantillas Tienda.dc.html).
	 */
	const PROPIAS: Record<string, Component<Props>> = {
		andes: Andes,
		ceibo: Ceibo,
		malbec: Malbec,
		miga: Miga,
		'miga-clasica': MigaClasica,
		mostaza: Mostaza,
		pampa: Pampa,
		quebracho: Quebracho,
		rosedal: Rosedal,
		tijera: Tijera,
		yerbal: Yerbal
	};

	let props: Props = $props();
	const Vista = $derived(PROPIAS[props.tienda.plantilla] ?? Motor);
	// El color de acento elegido en Mi tienda: las plantillas lo leen de var(--acento).
	const acento = $derived(acentoElegido(props.tienda.plantilla, props.tienda.paleta));
	// Mi tienda › Estilo: los colores por rol y las tipografías elegidas (lib/tienda/estilo). Las del
	// motor los toman de sus tokens (Motor.svelte); acá van las fuentes y las zonas.
	const estilo = $derived(cssEstilo(props.tienda.plantilla, props.tienda.estilo));
	const raiz = $derived(
		[acento ? `--acento:${acento}` : '', estilo.raiz].filter(Boolean).join(';')
	);
</script>

<svelte:head>
	{#if estilo.fuentes}<link rel="stylesheet" href={estilo.fuentes} />{/if}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -- colores #RRGGBB validados en estilo/css.ts -->
	{#if estilo.reglas}{@html `<style>${estilo.reglas}</style>`}{/if}
</svelte:head>

{#if raiz || estilo.reglas}
	<div class="t-estilo" style={raiz} style:display="contents"><Vista {...props} /></div>
{:else}
	<Vista {...props} />
{/if}
