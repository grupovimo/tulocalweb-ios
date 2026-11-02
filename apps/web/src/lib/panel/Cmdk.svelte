<script lang="ts">
	import { Dialog } from 'bits-ui';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { getUi } from '#lib/ui.svelte.ts';
	import { compartirTienda } from './compartir.ts';
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import type { IconData } from '#lib/icons.ts';
	import { AREAS, rutaDe, visible } from './navegacion.ts';
	import { getShell } from './shell.svelte.ts';

	/**
	 * ⌘K (handoff): preguntarle a Tincho, las acciones (crear pedido, link de pago, cargar con foto,
	 * abrir caja) y saltar a cualquier sección. La opción 0 es Tincho; el resto va de la 1 en adelante.
	 */
	const shell = getShell();
	const ui = getUi();
	let q = $state('');
	type Item = {
		t: string;
		s: string;
		icon: IconData;
		id: string;
		grupo: string;
		go: () => void;
		/** La sección que tiene que poder ver la persona (las acciones también siguen su rol). */
		sec?: string;
	};
	const ACCIONES: Item[] = [
		{
			t: 'Crear pedido',
			s: 'Manual, por WhatsApp o en el local',
			icon: I.phPlusBold,
			id: 'a-pedido',
			grupo: 'Acciones',
			sec: 'Pedidos',
			go: () => ir('/panel/pedidos/nuevo')
		},
		{
			t: 'Crear link de pago',
			s: 'Cobrá por WhatsApp en segundos',
			icon: I.phLinkBold,
			id: 'a-link',
			grupo: 'Acciones',
			sec: 'Pedidos',
			go: () => shell.abrirLink()
		},
		{
			t: 'Cargar producto con foto',
			s: 'Subí las fotos y completá nombre y precio',
			icon: I.phCameraBold,
			id: 'a-foto',
			grupo: 'Acciones',
			sec: 'Productos',
			go: () => ir(`${rutaDe('Productos')}?nuevo=1`)
		},
		{
			t: 'Cobrar en el local',
			s: 'Punto de venta',
			icon: I.phCashRegisterBold,
			id: 'a-caja',
			grupo: 'Acciones',
			sec: 'Ventas en local',
			go: () => ir(rutaDe('Ventas en local'))
		},
		{
			t: 'Compartir mi tienda',
			s: 'El link, por WhatsApp o redes',
			icon: I.phStorefrontBold,
			id: 'a-compartir',
			grupo: 'Acciones',
			go: () => {
				shell.cmdk = false;
				const t = page.data.tienda as { nombre: string; slug: string } | undefined;
				if (t) void compartirTienda(ui, t);
			}
		}
	];
	const items = $derived(
		[
			...ACCIONES.filter((a) => !a.sec || visible(a.sec, shell.turnos, shell.permisos)),
			...AREAS.flatMap((a) =>
				a.items
					.filter((i) => visible(i.label, shell.turnos, shell.permisos))
					.map((i): Item => {
						const href = rutaDe(i.label);
						return {
							t: i.label,
							s: a.label,
							icon: i.icon,
							id: href,
							grupo: 'Ir a',
							go: () => ir(href)
						};
					})
			)
		].filter((i) => !q.trim() || `${i.t} ${i.s}`.toLowerCase().includes(q.trim().toLowerCase()))
	);
	// Si lo escrito coincide con una sección, Enter va ahí; si no, se lo pregunta a Tincho. Las
	// flechas la mueven (derived escribible: vuelve a calcularse al cambiar lo escrito).
	let sel = $derived(q.trim() && items.length ? 1 : 0);
	$effect(() => {
		if (!shell.cmdk) q = '';
	});

	const EJEMPLO = '¿Cuánto vendí ayer por Mercado Libre?';
	function preguntar() {
		shell.hornero(q.trim() || EJEMPLO);
	}
	function ir(href: string) {
		shell.cmdk = false;
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- rutas del panel armadas con rutaDe
		goto(href);
	}
	function teclas(e: KeyboardEvent) {
		if (e.key === 'ArrowDown') {
			e.preventDefault();
			sel = Math.min(items.length, sel + 1);
			document.getElementById(`cmdk-${sel}`)?.scrollIntoView({ block: 'nearest' });
		} else if (e.key === 'ArrowUp') {
			e.preventDefault();
			sel = Math.max(0, sel - 1);
			document.getElementById(`cmdk-${sel}`)?.scrollIntoView({ block: 'nearest' });
		} else if (e.key === 'Enter') {
			e.preventDefault();
			if (sel === 0) preguntar();
			else items[sel - 1]?.go();
		}
	}
</script>

<svelte:window
	onkeydown={(e) => {
		if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
			e.preventDefault();
			shell.cmdk = !shell.cmdk;
		}
	}}
/>

<Dialog.Root bind:open={shell.cmdk}>
	<Dialog.Portal>
		<Dialog.Overlay class="cmdk-fondo" />
		<Dialog.Content class="cmdk" aria-describedby={undefined}>
			<Dialog.Title class="sr-only">Buscar</Dialog.Title>
			<div class="cab">
				<Icon icon={I.phMagnifyingGlass} size={20} style="color:var(--mu)" />
				<input
					bind:value={q}
					onkeydown={teclas}
					placeholder="Buscá una sección o preguntale a Tincho…"
					aria-label="Buscar"
					aria-controls="cmdk-lista"
					aria-activedescendant="cmdk-{sel}"
				/>
				<!-- En la compu, la pista de la tecla; en el celular, un botón para cerrar. -->
				<button
					type="button"
					class="cerrar"
					aria-label="Cerrar"
					onclick={() => (shell.cmdk = false)}
					><kbd>esc</kbd><span class="c-x"><Icon icon={I.phXBold} size={14} /></span></button
				>
			</div>
			<div class="lista" id="cmdk-lista" role="listbox" aria-label="Resultados">
				<button
					type="button"
					id="cmdk-0"
					role="option"
					aria-selected={sel === 0}
					class="hornero"
					class:sel={sel === 0}
					onmouseenter={() => (sel = 0)}
					onclick={preguntar}
				>
					<span class="h-ico"><Icon icon={I.phSparkleFill} size={18} /></span>
					<span class="txt"
						><span class="t">Preguntale a Tincho</span><span class="s h-q"
							>“{q.trim() || EJEMPLO}”</span
						></span
					>
					<kbd class="h-kbd">↵</kbd>
				</button>
				{#each items as it, i (it.id)}
					{#if it.grupo !== items[i - 1]?.grupo}<div class="g">{it.grupo}</div>{/if}
					<button
						type="button"
						id="cmdk-{i + 1}"
						role="option"
						aria-selected={i + 1 === sel}
						class:sel={i + 1 === sel}
						onmouseenter={() => (sel = i + 1)}
						onclick={it.go}
					>
						<span class="ico"><Icon icon={it.icon} size={17} /></span>
						<span class="txt"><span class="t">{it.t}</span><span class="s">{it.s}</span></span>
						<kbd>↵</kbd>
					</button>
				{:else}
					<div class="nada">Ninguna sección se llama «{q}».</div>
				{/each}
			</div>
			<div class="pie">
				<span>↑↓ navegar</span><span>↵ abrir</span><span>esc cerrar</span><span class="esp"
				></span><span>⌘K desde cualquier pantalla</span>
			</div>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>

<style>
	:global(.cmdk-fondo) {
		position: fixed;
		inset: 0;
		z-index: 70;
		background: rgba(42, 20, 12, 0.34);
		backdrop-filter: blur(3px);
	}
	:global(.cmdk) {
		position: fixed;
		left: 0;
		right: 0;
		margin: 0 auto;
		top: 11vh;
		z-index: 71;
		width: min(640px, calc(100vw - 24px));
		border-radius: 24px;
		background: var(--sf);
		color: var(--tx);
		box-shadow:
			0 40px 100px -20px rgba(42, 20, 12, 0.6),
			0 0 0 1px var(--ln);
		overflow: hidden;
		outline: none;
		/* La lista toma lo que queda de la pantalla visible (con o sin teclado) y scrollea ella. */
		display: flex;
		flex-direction: column;
		max-height: calc(100dvh - 11vh - 24px);
	}
	:global(.cmdk[data-state='open']) {
		animation: cmdk-entra 0.28s cubic-bezier(0.2, 0.7, 0.1, 1) both;
	}
	@keyframes cmdk-entra {
		from {
			opacity: 0;
			transform: translateY(8px) scale(0.98);
		}
	}
	.cab {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 16px 18px;
		border-bottom: 1px solid var(--ln);
	}
	.cab input {
		flex: 1;
		min-width: 0;
		border: 0;
		outline: 0;
		background: transparent;
		font-size: 17px;
		font-weight: 500;
		color: var(--tx);
	}
	kbd {
		font-family: inherit;
		font-size: 12px;
		font-weight: 600;
		padding: 3px 8px;
		border-radius: 6px;
		background: var(--chip);
		color: var(--mu);
		white-space: nowrap;
	}
	.lista {
		flex: 1 1 auto;
		min-height: 0;
		max-height: 440px;
		overflow-y: auto;
		/* En el iPhone, con el fondo bloqueado, el gesto tiene que quedar en la lista. */
		overscroll-behavior: contain;
		-webkit-overflow-scrolling: touch;
		touch-action: pan-y;
		padding: 8px;
	}
	.cab,
	.pie {
		flex: none;
	}
	.g {
		padding: 12px 12px 4px;
		font-size: 12px;
		font-weight: 600;
		color: var(--fa);
	}
	.lista button {
		width: 100%;
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 10px 12px;
		border-radius: 12px;
		border: 0;
		background: transparent;
		color: var(--tx);
		text-align: left;
	}
	.lista button.sel {
		background: var(--chip);
	}
	.lista .hornero {
		padding: 12px;
		border-radius: 14px;
		background: rgba(210, 88, 47, 0.09);
	}
	.lista .hornero.sel {
		background: rgba(210, 88, 47, 0.16);
	}
	.h-ico {
		width: 36px;
		height: 36px;
		flex: none;
		border-radius: 11px;
		background: linear-gradient(140deg, #e8875f, #c24e25);
		color: #fff;
		display: grid;
		place-items: center;
	}
	.h-q {
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		font-size: 13px !important;
	}
	.h-kbd {
		background: var(--sf);
	}
	.ico {
		width: 32px;
		height: 32px;
		flex: none;
		border-radius: 10px;
		display: grid;
		place-items: center;
		background: rgba(210, 88, 47, 0.1);
		color: #b0441d;
	}
	.txt {
		flex: 1;
		min-width: 0;
	}
	.t {
		display: block;
		font-size: 14px;
		font-weight: 600;
	}
	.s {
		display: block;
		font-size: 12.5px;
		color: var(--mu);
	}
	.nada {
		padding: 28px;
		text-align: center;
		font-size: 14px;
		color: var(--mu);
	}
	.pie {
		display: flex;
		gap: 16px;
		padding: 10px 18px;
		border-top: 1px solid var(--ln);
		font-size: 12px;
		color: var(--fa);
	}
	.esp {
		flex: 1;
	}
	.cerrar {
		flex: none;
		border: 0;
		padding: 0;
		background: transparent;
		color: var(--mu);
		cursor: pointer;
	}
	.c-x {
		display: none;
	}
	/* Pantallas táctiles: sin pistas de teclado; el cerrar, de 44 px. */
	@media (pointer: coarse) {
		kbd {
			display: none;
		}
		.c-x {
			display: grid;
			place-items: center;
			width: 44px;
			height: 44px;
			margin: -8px -10px -8px 0;
			border-radius: 50%;
		}
	}
	@media (max-width: 759px) {
		:global(.cmdk) {
			top: 12px;
			max-height: calc(100dvh - 24px);
		}
		.lista {
			max-height: none;
		}
		.pie {
			display: none;
		}
	}
</style>
