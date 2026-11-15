<script lang="ts">
	import type { Snippet } from 'svelte';

	/** Un teléfono dibujado (sin imagen): marco, isla y pantalla. Lo de adentro lo pone la escena. */
	let {
		children,
		fondo = '#2a140c',
		proporcion = '9 / 18.5'
	}: { children: Snippet; fondo?: string; proporcion?: string } = $props();
</script>

<div class="tel" style:--fondo={fondo} style:aspect-ratio={proporcion}>
	<div class="pantalla">
		<span class="isla"></span>
		{@render children()}
	</div>
</div>

<style>
	.tel {
		width: min(300px, 78vw);
		padding: 10px;
		border-radius: 46px;
		background: #1b0d07;
		box-shadow:
			0 50px 80px -40px rgba(42, 20, 12, 0.75),
			inset 0 0 0 2px rgba(255, 255, 255, 0.08);
	}
	.pantalla {
		position: relative;
		height: 100%;
		border-radius: 37px;
		overflow: hidden;
		background: var(--fondo);
	}
	.isla {
		position: absolute;
		z-index: 2;
		top: 10px;
		left: 50%;
		width: 86px;
		height: 24px;
		margin-left: -43px;
		border-radius: 20px;
		background: #000;
	}
</style>
