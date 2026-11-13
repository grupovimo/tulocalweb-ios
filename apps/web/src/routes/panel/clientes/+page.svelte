<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import Icon from '#lib/components/Icon.svelte';
	import Modal from '#lib/components/Modal.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import FichaCliente from '#lib/panel/listas/FichaCliente.svelte';
	import MotorLista from '#lib/panel/listas/MotorLista.svelte';
	import { configClientes, escribir, type Cliente } from '#lib/panel/listas/clientes.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/** Clientes (PanelListas, pantalla Clientes) con su ficha, el alta y el etiquetado. */
	let { data } = $props();
	const ui = getUi();

	let alta = $state(false);
	let etiquetar = $state<string[]>([]);
	let etiquetando = $state(false);
	let ficha = $state(false);
	let elegido = $state<Cliente | null>(null);
	const config = configClientes(
		() => (alta = true),
		(ids) => {
			etiquetar = ids;
			etiquetando = true;
		}
	);

	async function accion(ids: string[], accion: string, valor?: unknown) {
		const r = await fetch('/panel/listas/clientes', {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ ids, accion, valor })
		});
		await invalidateAll();
		return r.ok ? ((await r.json()) as { hechos: number }).hechos : 0;
	}

	let nombre = $state('');
	let telefono = $state('');
	let email = $state('');
	let ciudad = $state('');
	let errorAlta = $state<string | null>(null);
	async function crear(e: SubmitEvent) {
		e.preventDefault();
		if (nombre.trim().length < 2) {
			errorAlta = 'Poné el nombre del cliente.';
			return;
		}
		if (telefono && telefono.replace(/\D/g, '').length < 8) {
			errorAlta = 'El WhatsApp tiene que tener la característica: 11 5555-1234.';
			return;
		}
		const n = await accion([], 'crear', { nombre, telefono, email, ciudad });
		if (!n) {
			errorAlta = 'No pudimos guardarlo. Revisá los datos.';
			return;
		}
		ui.toast({ k: 'ok', title: `${nombre.trim()} ya es cliente`, desc: 'Lo encontrás en Nuevos' });
		nombre = telefono = email = ciudad = '';
		errorAlta = null;
		alta = false;
	}

	// ---- Corregir sus datos y el pedido de supresión ----
	let corrigiendo = $state(false);
	let borrando = $state(false);
	let cNombre = $state('');
	let cEmail = $state('');
	let cTel = $state('');
	let errorDatos = $state<string | null>(null);
	let enviando = $state(false);
	function abrirCorregir(c: Cliente) {
		elegido = c;
		cNombre = c.nombre;
		cEmail = c.email ?? '';
		cTel = c.telefono ?? '';
		errorDatos = null;
		corrigiendo = true;
	}
	function abrirBorrar(c: Cliente) {
		elegido = c;
		errorDatos = null;
		borrando = true;
	}
	/** La acción sobre un cliente, con el mensaje de la API si no se pudo. */
	async function sobreCliente(accion: 'editar' | 'suprimir', valor?: unknown) {
		if (!elegido || enviando) return false;
		enviando = true;
		try {
			const r = await fetch('/panel/listas/clientes', {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify({ ids: [elegido.id], accion, valor })
			});
			if (!r.ok) {
				const p = (await r.json().catch(() => null)) as { message?: string } | null;
				errorDatos = p?.message ?? 'No pudimos guardarlo. Probá de nuevo.';
				return false;
			}
			await invalidateAll();
			return true;
		} finally {
			enviando = false;
		}
	}
	async function corregir(e: SubmitEvent) {
		e.preventDefault();
		if (cNombre.trim().length < 2) {
			errorDatos = 'Poné el nombre del cliente.';
			return;
		}
		if (cTel && cTel.replace(/\D/g, '').length < 8) {
			errorDatos = 'El WhatsApp tiene que tener la característica: 11 5555-1234.';
			return;
		}
		if (!(await sobreCliente('editar', { nombre: cNombre, email: cEmail, telefono: cTel }))) return;
		ui.toast({ k: 'ok', title: 'Datos corregidos', desc: 'En su ficha y en sus pedidos' });
		corrigiendo = false;
		ficha = false;
	}
	async function suprimir() {
		if (!(await sobreCliente('suprimir'))) return;
		ui.toast({
			k: 'ok',
			title: 'Datos borrados',
			desc: 'Sus ventas quedan como «Consumidor final»'
		});
		borrando = false;
		ficha = false;
	}

	let etiqueta = $state('');
	async function aplicarEtiqueta(e: SubmitEvent) {
		e.preventDefault();
		if (!etiquetar.length || !etiqueta.trim()) return;
		const n = await accion(etiquetar, 'etiquetar', etiqueta);
		ui.toast({
			k: 'ok',
			title: `Etiqueta «${etiqueta.trim()}»`,
			desc: `${n} ${n === 1 ? 'cliente' : 'clientes'}`
		});
		etiqueta = '';
		etiquetando = false;
	}
</script>

<Seo title="Clientes · Panel" description="Clientes de tu tienda" path="/panel/clientes" noindex />

<MotorLista
	{config}
	lista={data.lista}
	abrir={(c: Cliente) => {
		elegido = c;
		ficha = true;
	}}
/>

<FichaCliente
	bind:open={ficha}
	cliente={elegido}
	escribir={(c) => {
		if (!escribir(c))
			ui.toast({ k: 'warn', title: 'Sin WhatsApp cargado', desc: 'Sumalo desde su ficha' });
	}}
	cupon={async (c) => {
		await accion([c.id], 'cupon');
		elegido = { ...c, cupon: true };
		ui.toast({
			k: 'ok',
			title: `Cupón del 10% para ${c.nombre.split(' ')[0]}`,
			desc: 'Enviado por WhatsApp · vence en 7 días'
		});
	}}
	editar={abrirCorregir}
	suprimir={abrirBorrar}
/>

<Modal bind:open={corrigiendo} titulo="Corregir datos" ancho={460}>
	<form onsubmit={corregir}>
		<div class="m-ico"><Icon icon={I.phPencilSimpleFill} size={24} /></div>
		<div class="m-t">Corregir datos</div>
		<div class="m-d">
			Se corrigen en su ficha y en sus pedidos. Las facturas ya emitidas no cambian.
		</div>
		<div class="m-campos">
			<label class="field"
				>Nombre y apellido<input
					class="input"
					bind:value={cNombre}
					maxlength="80"
					autocomplete="off"
				/></label
			>
			<label class="field"
				>WhatsApp<input
					class="input"
					bind:value={cTel}
					inputmode="tel"
					placeholder="11 5555-1234"
					maxlength="30"
				/></label
			>
			<label class="field"
				>Email<input class="input" bind:value={cEmail} type="email" maxlength="120" /></label
			>
		</div>
		{#if errorDatos}<div class="m-error" role="alert">{errorDatos}</div>{/if}
		<div class="m-acc">
			<button type="button" class="btn btn-secondary" onclick={() => (corrigiendo = false)}
				>Cancelar</button
			>
			<button class="btn btn-primary" disabled={enviando}>Guardar</button>
		</div>
	</form>
</Modal>

<Modal bind:open={borrando} titulo="Borrar sus datos" ancho={460}>
	<div class="m-ico rojo"><Icon icon={I.phTrashFill} size={24} /></div>
	<div class="m-t">¿Borrar los datos de {elegido?.nombre ?? 'este cliente'}?</div>
	<div class="m-d">
		Es para cuando te lo pide. Borramos su ficha y su contacto, y sus pedidos quedan como venta a
		«Consumidor&nbsp;final». Las facturas emitidas se guardan como salieron, porque la ley lo pide.
		No se puede deshacer.
	</div>
	{#if errorDatos}<div class="m-error" role="alert">{errorDatos}</div>{/if}
	<div class="m-acc">
		<button type="button" class="btn btn-secondary" onclick={() => (borrando = false)}
			>Cancelar</button
		>
		<button type="button" class="btn btn-danger" disabled={enviando} onclick={suprimir}
			>Borrar sus datos</button
		>
	</div>
</Modal>

<Modal bind:open={alta} titulo="Nuevo cliente" ancho={460}>
	<form onsubmit={crear}>
		<div class="m-ico"><Icon icon={I.phUserPlusFill} size={24} /></div>
		<div class="m-t">Nuevo cliente</div>
		<div class="m-d">Para cargar a alguien que compró en el local o por teléfono.</div>
		<div class="m-campos">
			<label class="field"
				>Nombre y apellido<input
					class="input"
					bind:value={nombre}
					maxlength="80"
					autocomplete="off"
				/></label
			>
			<label class="field"
				>WhatsApp<input
					class="input"
					bind:value={telefono}
					inputmode="tel"
					placeholder="11 5555-1234"
					maxlength="30"
				/></label
			>
			<label class="field"
				>Email<input class="input" bind:value={email} type="email" maxlength="120" /></label
			>
			<label class="field">Ciudad<input class="input" bind:value={ciudad} maxlength="60" /></label>
		</div>
		{#if errorAlta}<div class="m-error" role="alert">{errorAlta}</div>{/if}
		<div class="m-acc">
			<button type="button" class="btn btn-secondary" onclick={() => (alta = false)}
				>Cancelar</button
			>
			<button class="btn btn-primary">Guardar cliente</button>
		</div>
	</form>
</Modal>

<Modal bind:open={etiquetando} titulo="Etiquetar clientes" ancho={420}>
	<form onsubmit={aplicarEtiqueta}>
		<div class="m-ico violeta"><Icon icon={I.phTagFill} size={24} /></div>
		<div class="m-t">
			Etiquetar {etiquetar.length}
			{etiquetar.length === 1 ? 'cliente' : 'clientes'}
		</div>
		<div class="m-d">
			Sirve para buscarlos y para mandarles campañas: «Mayoristas zona sur», «Feria de mayo».
		</div>
		<div class="m-campos">
			<label class="field"
				>Etiqueta<input
					class="input"
					bind:value={etiqueta}
					maxlength="30"
					autocomplete="off"
				/></label
			>
		</div>
		<div class="m-acc">
			<button type="button" class="btn btn-secondary" onclick={() => (etiquetando = false)}
				>Cancelar</button
			>
			<button class="btn btn-primary" disabled={!etiqueta.trim()}>Etiquetar</button>
		</div>
	</form>
</Modal>

<style>
	.m-ico {
		width: 48px;
		height: 48px;
		display: grid;
		place-items: center;
		border-radius: 15px;
		background: rgba(210, 88, 47, 0.12);
		color: #b0441d;
	}
	.m-ico.rojo {
		background: rgba(184, 58, 46, 0.12);
		color: #b83a2e;
	}
	.m-ico.violeta {
		background: rgba(106, 79, 160, 0.12);
		color: #6a4fa0;
	}
	.m-t {
		margin-top: 16px;
		font-size: 21px;
		font-weight: 600;
		letter-spacing: -0.02em;
	}
	.m-d {
		margin-top: 6px;
		font-size: 14.5px;
		line-height: 1.5;
		color: var(--mu);
	}
	.m-campos {
		margin-top: 16px;
		display: grid;
		gap: 12px;
	}
	.m-error {
		margin-top: 10px;
		font-size: 13px;
		font-weight: 600;
		color: #b83a2e;
	}
	.m-acc {
		margin-top: 20px;
		display: flex;
		gap: 8px;
		justify-content: flex-end;
	}
</style>
