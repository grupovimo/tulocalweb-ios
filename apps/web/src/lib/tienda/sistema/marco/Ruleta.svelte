<script lang="ts">
	/**
	 * La ruleta de una ventana: de 4 a 8 porciones con los colores de la tienda (acento y
	 * superficie, alternados) y la flecha arriba. No sortea nada: cuando el servidor devuelve la
	 * porción (`porcion`), gira unas vueltas y cae en el medio de esa porción; al terminar avisa con
	 * `onlisto`. Con «Reducir movimiento», cae sin girar. Es decorativa para los lectores de pantalla:
	 * el resultado se anuncia en texto (marco/Ventana.svelte).
	 */
	let {
		porciones,
		porcion = null,
		onlisto
	}: { porciones: string[]; porcion?: number | null; onlisto?: () => void } = $props();

	const n = $derived(Math.max(1, porciones.length));
	const paso = $derived(360 / n);
	const R = 100;
	const punto = (a: number, r = R) => {
		const rad = ((a - 90) * Math.PI) / 180;
		return [R + r * Math.cos(rad), R + r * Math.sin(rad)];
	};
	const sector = (i: number) => {
		const [x1, y1] = punto(i * paso);
		const [x2, y2] = punto((i + 1) * paso);
		return `M${R},${R} L${x1},${y1} A${R},${R} 0 ${paso > 180 ? 1 : 0} 1 ${x2},${y2} Z`;
	};
	let giro = $state(0);
	let quieta = $state(false);
	$effect(() => {
		if (porcion === null || porcion < 0) return;
		quieta = matchMedia('(prefers-reduced-motion: reduce)').matches;
		// La porción i queda arriba (en la flecha) girando la rueda hasta su centro.
		giro = (quieta ? 0 : 360 * 6) + (360 - (porcion + 0.5) * paso);
		if (quieta) onlisto?.();
	});
	const corto = (t: string) => (t.length > 14 ? `${t.slice(0, 13)}…` : t);
</script>

<div class="tv-ruleta" aria-hidden="true">
	<svg
		viewBox="0 0 200 200"
		class="tv-ruleta-rueda"
		style:transform="rotate({giro}deg)"
		class:tv-ruleta-quieta={quieta}
		ontransitionend={() => porcion !== null && onlisto?.()}
	>
		{#each porciones as p, i (i)}
			{@const [tx, ty] = punto((i + 0.5) * paso, 58)}
			<path d={sector(i)} class={i % 2 ? 'tv-ruleta-b' : 'tv-ruleta-a'} />
			<text
				x={tx}
				y={ty}
				class={i % 2 ? 'tv-ruleta-tb' : 'tv-ruleta-ta'}
				transform="rotate({(i + 0.5) * paso - 90} {tx} {ty})"
				text-anchor="middle"
				dominant-baseline="middle">{corto(p)}</text
			>
		{/each}
		<circle cx={R} cy={R} r="14" class="tv-ruleta-centro" />
	</svg>
	<span class="tv-ruleta-flecha"></span>
</div>
