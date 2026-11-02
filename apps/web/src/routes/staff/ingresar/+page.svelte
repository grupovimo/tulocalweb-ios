<script lang="ts">
	import { enhance } from '$app/forms';
	import Seo from '#lib/components/Seo.svelte';

	/** Ingreso del equipo: email y contraseña (las altas las hace `mostradorctl equipo-alta`). */
	let { form } = $props();
	let enviando = $state(false);
</script>

<Seo
	title="Equipo · Tu Local Web"
	description="Ingreso del equipo"
	path="/staff/ingresar"
	noindex
/>

<main class="pagina">
	<form
		method="POST"
		class="caja"
		use:enhance={() => {
			enviando = true;
			return async ({ update }) => {
				enviando = false;
				await update();
			};
		}}
	>
		<span class="marca"><img src="/logo.svg" alt="" width="22" height="22" />Equipo</span>
		<h1>Ingresar al backoffice</h1>
		<p class="sub">Sólo para el equipo de Tu Local Web.</p>
		<label class="campo"
			>Email<input
				name="email"
				type="email"
				autocomplete="username"
				value={form?.email ?? ''}
				required
			/></label
		>
		<label class="campo"
			>Contraseña<input
				name="clave"
				type="password"
				autocomplete="current-password"
				required
			/></label
		>
		{#if form?.error}<p class="error" role="alert">{form.error}</p>{/if}
		<button type="submit" class="boton" disabled={enviando}
			>{enviando ? 'Ingresando…' : 'Ingresar'}</button
		>
	</form>
</main>

<style>
	.pagina {
		min-height: 100vh;
		display: grid;
		place-items: center;
		padding: 16px;
		background: #2a140c;
		font-family: 'Instrument Sans', system-ui, sans-serif;
	}
	.caja {
		width: min(100%, 400px);
		display: flex;
		flex-direction: column;
		gap: 12px;
		padding: 28px 24px;
		border-radius: 24px;
		background: #fbf7f2;
		color: #2a140c;
	}
	.marca {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 12.5px;
		font-weight: 700;
		letter-spacing: 0.04em;
		text-transform: uppercase;
	}
	h1 {
		margin: 4px 0 0;
		font-size: 26px;
		letter-spacing: -0.03em;
		font-weight: 600;
	}
	.sub {
		margin: 0;
		color: #6b5244;
		font-size: 14.5px;
	}
	.campo {
		display: flex;
		flex-direction: column;
		gap: 4px;
		font-size: 12.5px;
		font-weight: 600;
		color: #6b5244;
	}
	.campo input {
		padding: 12px;
		border-radius: 12px;
		border: 1px solid rgba(58, 29, 18, 0.16);
		background: #fff;
		font:
			500 15px 'Instrument Sans',
			sans-serif;
		color: #2a140c;
		outline: none;
	}
	.campo input:focus {
		border-color: #d2582f;
	}
	.error {
		margin: 0;
		padding: 10px 12px;
		border-radius: 12px;
		background: #fcede5;
		color: #a8401c;
		font-size: 13.5px;
	}
	.boton {
		margin-top: 4px;
		padding: 13px;
		border-radius: 999px;
		border: 0;
		background: linear-gradient(180deg, #c24e25, #b0441d);
		color: #fff;
		font:
			600 15px 'Instrument Sans',
			sans-serif;
	}
</style>
