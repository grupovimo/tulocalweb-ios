<script lang="ts">
	import { destinos, esPropio, URL_HTTPS, type Ayuda } from './campos.ts';

	/** Adónde lleva un link o un botón: una página, una categoría, un producto o un link propio. */
	let {
		valor,
		etiqueta,
		ayuda,
		onvalor
	}: {
		valor: string;
		etiqueta: string;
		ayuda: Ayuda;
		onvalor: (v: string, confirmar: boolean) => void;
	} = $props();

	const d = $derived(destinos(ayuda.cx));
	// Hasta que el link propio es válido no se guarda: el diseño nunca lleva uno a medio escribir.
	let escribiendo = $state(false);
	const propio = $derived(escribiendo || esPropio(valor));
	let error = $state('');
</script>

<select
	class="input"
	aria-label={etiqueta}
	value={propio ? 'propio' : valor}
	onchange={(e) => {
		const v = e.currentTarget.value;
		error = '';
		escribiendo = v === 'propio';
		if (!escribiendo) onvalor(v, true);
	}}
>
	<option value="">Sin link</option>
	<optgroup label="Páginas">
		{#each d.paginas as p (p.valor)}<option value={p.valor}>{p.nombre}</option>{/each}
	</optgroup>
	{#if d.categorias.length}
		<optgroup label="Categorías">
			{#each d.categorias as c (c.valor)}<option value={c.valor}>{c.nombre}</option>{/each}
		</optgroup>
	{/if}
	{#if d.productos.length}
		<optgroup label="Productos">
			{#each d.productos as p (p.valor)}<option value={p.valor}>{p.nombre}</option>{/each}
		</optgroup>
	{/if}
	<option value="propio">Un link propio…</option>
</select>
{#if propio}
	<input
		class="input"
		type="url"
		inputmode="url"
		maxlength="300"
		placeholder="https://"
		aria-label="{etiqueta}: link propio"
		aria-invalid={!!error}
		value={esPropio(valor) ? valor : ''}
		oninput={(e) => {
			const u = e.currentTarget.value.trim();
			if (URL_HTTPS.test(u)) onvalor(u, false);
		}}
		onchange={(e) => {
			const u = e.currentTarget.value.trim();
			error = URL_HTTPS.test(u) ? '' : 'Tiene que empezar con https://';
			if (!error) onvalor(u, true);
		}}
	/>
	{#if error}<span class="error">{error}</span>{/if}
{/if}

<style>
	select.input {
		appearance: auto;
	}
	.error {
		font-size: 12.5px;
		font-weight: 500;
		color: var(--error);
	}
</style>
