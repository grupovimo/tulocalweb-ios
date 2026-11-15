<script lang="ts">
	import { browser } from '$app/env';
	import { faltan, textoFaltan, textoFaltanLargo } from '../ventanas.ts';

	/**
	 * La cuenta regresiva hasta `hasta` (ISO): «2 d 04:10:05» con el número tabular, y para los
	 * lectores de pantalla «Faltan 2 días y 4 horas» (sin anunciar cada segundo). Al vencer avisa
	 * con `alVencer` y no dibuja nada. `caja`: los números en cajitas (las ventanas); si no, en línea
	 * (la barra de anuncios).
	 */
	let {
		hasta,
		caja = false,
		alVencer
	}: { hasta: string; caja?: boolean; alVencer?: () => void } = $props();

	let ahora = $state(Date.now());
	$effect(() => {
		if (!browser) return;
		const t = setInterval(() => (ahora = Date.now()), 1000);
		return () => clearInterval(t);
	});
	const f = $derived(faltan(hasta, ahora));
	$effect(() => {
		if (browser && !f) alVencer?.();
	});
	const partes = $derived(
		f
			? [
					...(f.d ? [{ n: f.d, u: f.d === 1 ? 'día' : 'días' }] : []),
					{ n: f.h, u: 'h' },
					{ n: f.m, u: 'min' },
					...(f.d ? [] : [{ n: f.s, u: 's' }])
				]
			: []
	);
</script>

{#if f}
	{#if caja}
		<span class="tv-regresiva tv-regresiva--caja">
			<span class="tv-oculto">{textoFaltanLargo(f)}</span>
			{#each partes as p (p.u)}<span class="tv-regresiva-celda" aria-hidden="true"
					><b>{String(p.n).padStart(2, '0')}</b><small>{p.u}</small></span
				>{/each}
		</span>
	{:else}
		<span class="tv-regresiva"
			><span class="tv-oculto">{textoFaltanLargo(f)}</span><span aria-hidden="true"
				>{textoFaltan(f)}</span
			></span
		>
	{/if}
{/if}
