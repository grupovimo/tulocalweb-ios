<script lang="ts">
	import { usarTienda } from '../contexto.ts';

	/**
	 * «¿Sos mayor de 18 años?» antes de mostrar la tienda (función propia de Malbec, Ley 24.788). Se
	 * pregunta una vez y se recuerda en este navegador. Lo pone marco/Tienda.svelte sólo si el preset
	 * lo trae.
	 */
	const ctx = usarTienda();
	const clave = $derived(`mayor18:${ctx.tienda.slug}`);
	let mayor = $state<boolean | null>(null);
	let menor = $state(false);
	$effect(() => {
		try {
			mayor = localStorage.getItem(clave) === '1';
		} catch {
			mayor = false;
		}
	});
	function soyMayor() {
		mayor = true;
		try {
			localStorage.setItem(clave, '1');
		} catch {
			// Sin almacenamiento (modo privado): se vuelve a preguntar en la próxima visita.
		}
	}
</script>

{#if mayor === false}
	<div class="tv-18" role="dialog" aria-modal="true" aria-labelledby="tv-18-titulo">
		<div class="tv-18-caja tv-tarjeta">
			<p class="tv-ante">{ctx.tienda.nombre}</p>
			<h2 id="tv-18-titulo" class="tv-h2">¿Sos mayor de 18&nbsp;años?</h2>
			{#if menor}
				<p class="tv-suave">
					No podemos venderte alcohol. Te esperamos cuando seas mayor de&nbsp;edad.
				</p>
			{:else}
				<p class="tv-suave">La venta de alcohol a menores está prohibida (Ley&nbsp;24.788).</p>
				<div class="tv-botones">
					<button type="button" class="tv-btn tv-btn--primario" onclick={soyMayor}
						>Sí, soy mayor</button
					>
					<button type="button" class="tv-btn tv-btn--secundario" onclick={() => (menor = true)}
						>No</button
					>
				</div>
			{/if}
		</div>
	</div>
{/if}
