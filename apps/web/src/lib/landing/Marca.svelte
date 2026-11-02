<script lang="ts">
	/**
	 * Isotipo de Tu Local Web: la casita con la bolsa, calcada del logo plano del tablero de identidad
	 * (static/marca/logo/). A color por defecto; con `color`, de un solo tono (para fondos oscuros o
	 * de color): la misma silueta pintada con ese color.
	 */
	let { color, alto = 20 }: { color?: string; alto?: number } = $props();
</script>

{#if color}
	<span
		class="mono"
		style:width="{alto}px"
		style:height="{alto}px"
		style:background-color={color}
		aria-hidden="true"
	></span>
{:else}
	<img src="/logo.svg" alt="" width={alto} height={alto} aria-hidden="true" />
{/if}

<style>
	img,
	span {
		display: block;
		flex: none;
	}
	.mono {
		mask: url(/marca/logo/logo-blanco.svg) center / contain no-repeat;
		-webkit-mask: url(/marca/logo/logo-blanco.svg) center / contain no-repeat;
	}
</style>
