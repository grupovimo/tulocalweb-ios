<script lang="ts" module>
	export type EstadoDireccion = 'actual' | 'revisando' | 'ok' | 'mal';
</script>

<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import {
		HOST,
		MENSAJES,
		motivoDeFormato,
		normalizarSlug,
		type Disponibilidad
	} from '#lib/tienda/direccion.ts';

	/**
	 * La dirección de la tienda (<slug>.tulocalweb.com.ar) editable, con la disponibilidad en vivo
	 * (decisiones §38): el formato y los reservados se ven al escribir; si está libre lo dice Go
	 * (GET /panel/configuracion/direccion), con una pausa para no preguntar en cada tecla.
	 */
	let {
		slug = $bindable(),
		actual,
		estado = $bindable('actual'),
		id = 'direccion',
		etiqueta = 'Dirección de tu tienda',
		oninput
	}: {
		slug: string;
		/** La que tiene hoy (no se pregunta: es suya). */
		actual: string;
		estado?: EstadoDireccion;
		id?: string;
		etiqueta?: string;
		oninput?: () => void;
	} = $props();

	let mensaje = $state('');

	$effect(() => {
		const s = slug.replace(/-+$/, '');
		if (s === actual) {
			estado = 'actual';
			mensaje = '';
			return;
		}
		const m = motivoDeFormato(s);
		if (m) {
			estado = 'mal';
			mensaje = MENSAJES[m];
			return;
		}
		estado = 'revisando';
		mensaje = 'Fijándonos si está libre…';
		const ctl = new AbortController();
		const t = setTimeout(async () => {
			try {
				const r = await fetch(`/panel/configuracion/direccion?slug=${encodeURIComponent(s)}`, {
					signal: ctl.signal
				});
				if (!r.ok) throw new Error(String(r.status));
				const d = (await r.json()) as Disponibilidad;
				estado = d.motivo === 'actual' ? 'actual' : d.disponible ? 'ok' : 'mal';
				mensaje = d.motivo === 'libre' ? '¡Está libre!' : d.mensaje;
			} catch {
				if (ctl.signal.aborted) return;
				// Sin respuesta no se bloquea: al guardar lo vuelve a mirar Go.
				estado = 'ok';
				mensaje = 'No pudimos fijarnos ahora; lo revisamos al guardar.';
			}
		}, 350);
		return () => {
			clearTimeout(t);
			ctl.abort();
		};
	});
</script>

<div class="cd">
	<label class="cd-l" for={id}>{etiqueta}</label>
	<div class="cd-caja" class:mal={estado === 'mal'} class:bien={estado === 'ok'}>
		<input
			{id}
			value={slug}
			oninput={(e) => {
				const v = normalizarSlug(e.currentTarget.value);
				e.currentTarget.value = v;
				slug = v;
				oninput?.();
			}}
			onblur={() => (slug = slug.replace(/-+$/, ''))}
			maxlength="60"
			autocomplete="off"
			autocapitalize="off"
			spellcheck="false"
			inputmode="url"
			aria-describedby="{id}-estado"
			aria-invalid={estado === 'mal'}
		/>
		<span class="cd-pre" aria-hidden="true">.{HOST}</span>
	</div>
	<div
		class="cd-estado"
		class:mal={estado === 'mal'}
		class:bien={estado === 'ok'}
		id="{id}-estado"
		aria-live="polite"
	>
		{#if estado === 'revisando'}
			<Icon icon={I.phCircleNotchBold} size={14} class="gira" />
		{:else if estado === 'ok'}
			<Icon icon={I.phCheckCircleFill} size={14} />
		{:else if estado === 'mal'}
			<Icon icon={I.phWarningCircleFill} size={14} />
		{/if}
		<span>{estado === 'actual' ? 'Es la dirección que tenés ahora.' : mensaje}</span>
	</div>
</div>

<style>
	.cd {
		display: flex;
		flex-direction: column;
		gap: 6px;
		min-width: 0;
	}
	.cd-l {
		font-size: 12.5px;
		font-weight: 600;
		color: var(--mu, #7a6252);
	}
	.cd-caja {
		display: flex;
		align-items: center;
		min-width: 0;
		border-radius: 12px;
		border: 1px solid var(--ln2, rgba(58, 29, 18, 0.16));
		background: var(--bg, #fbf7f2);
		padding: 0 12px;
		transition: border-color 0.2s;
	}
	.cd-caja:focus-within {
		border-color: var(--terracota, #d2582f);
	}
	.cd-caja.mal {
		border-color: var(--error, #b83a2e);
	}
	.cd-pre {
		flex: none;
		font-size: 13.5px;
		color: var(--mu, #7a6252);
		white-space: nowrap;
	}
	.cd-caja input {
		flex: 1;
		min-width: 0;
		min-height: 44px;
		padding: 0;
		border: 0;
		background: transparent;
		font:
			600 14.5px 'Instrument Sans',
			sans-serif;
		color: var(--tx, #2a140c);
		outline: none;
	}
	.cd-estado {
		display: flex;
		align-items: flex-start;
		gap: 6px;
		min-height: 18px;
		font-size: 12.5px;
		line-height: 1.4;
		color: var(--mu, #7a6252);
	}
	.cd-estado :global(svg) {
		flex: none;
		margin-top: 2px;
	}
	.cd-estado.bien {
		color: var(--ok, #2f8f63);
	}
	.cd-estado.mal {
		color: var(--error, #b83a2e);
	}
	.cd-estado :global(.gira) {
		animation: gira 0.9s linear infinite;
	}
	@keyframes gira {
		to {
			transform: rotate(360deg);
		}
	}
	@media (prefers-reduced-motion: reduce) {
		.cd-estado :global(.gira) {
			animation: none;
		}
	}
</style>
