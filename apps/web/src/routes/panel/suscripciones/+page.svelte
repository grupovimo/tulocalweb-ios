<script lang="ts">
	import { goto, invalidateAll } from '$app/navigation';
	import Alerta from '#lib/components/Alerta.svelte';
	import ChipEstado from '#lib/components/ChipEstado.svelte';
	import Icon from '#lib/components/Icon.svelte';
	import Interruptor from '#lib/components/Interruptor.svelte';
	import Modal from '#lib/components/Modal.svelte';
	import Segmentado from '#lib/components/Segmentado.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import Vacio from '#lib/components/Vacio.svelte';
	import * as I from '#lib/icons.ts';
	import { unsplash } from '#lib/landing/formato.ts';
	import type { Estado } from '#lib/marcas.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { pesos } from '#lib/panel/formato.ts';
	import { getShell } from '#lib/panel/shell.svelte.ts';
	import {
		CADA,
		ESTADO_SUSCRIPCION,
		FRECUENCIAS,
		diaCobro,
		type EstadoSuscripcion
	} from '#lib/tienda/suscripciones.ts';
	import { getUi } from '#lib/ui.svelte.ts';
	import type { components } from '@mostrador/contrato/esquema';

	type Plan = components['schemas']['PlanSuscripcion'];
	type Editable = components['schemas']['PlanSuscripcionEditable'];
	type Suscriptor = components['schemas']['Suscriptor'];

	/**
	 * Suscripciones (§51): los planes con débito automático de Mercado Pago y quienes se suscribieron.
	 * No hay diseño en el handoff: se arma con las piezas de Ajustes de turnos y Pedidos (cabecera,
	 * tarjetas, filas, chips de estado, Segmentado, Interruptor y Modal).
	 */
	let { data } = $props();
	const ui = getUi();
	const shell = getShell();
	/** Crear y editar planes pide Tienda (Go lo exige igual); ver y cancelar, Pedidos. */
	const editaPlanes = $derived(shell.permisos.includes('tienda'));
	const veMontos = $derived(shell.permisos.includes('finanzas'));

	async function pedir(url: string, method: string, cuerpo?: unknown) {
		const r = await fetch(url, {
			method,
			headers: cuerpo === undefined ? {} : { 'content-type': 'application/json' },
			body: cuerpo === undefined ? undefined : JSON.stringify(cuerpo)
		});
		if (!r.ok) {
			const m = ((await r.json().catch(() => null)) as { message?: string } | null)?.message;
			ui.toast({ k: 'err', title: 'No se pudo', desc: m ?? 'Probá de nuevo.' });
			return null;
		}
		return r;
	}

	// ---- Resumen ----
	const activas = $derived(data.suscriptores.filter((s) => s.estado === 'activa'));
	const pausadas = $derived(data.suscriptores.filter((s) => s.estado === 'pausada'));
	const porMes = $derived(
		activas.reduce(
			(t, s) =>
				t +
				(s.frecuencia === 'quincenal'
					? s.precio * 2
					: s.frecuencia === 'bimestral'
						? s.precio / 2
						: s.precio),
			0
		)
	);

	// ---- Planes ----
	type Form = Editable & { id: string | null };
	let plan = $state<Form | null>(null);
	let conCupo = $state(false);
	const nuevoPlan = () => {
		conCupo = false;
		plan = {
			id: null,
			nombre: '',
			descripcion: '',
			foto: null,
			precio: 0,
			frecuencia: 'mensual',
			productos: [],
			incluye: '',
			envio: true,
			retiro: true,
			cupo: null,
			activo: true
		};
	};
	const editar = (p: Plan) => {
		conCupo = p.cupo !== null;
		plan = {
			id: p.id,
			nombre: p.nombre,
			descripcion: p.descripcion,
			foto: p.foto,
			precio: p.precio,
			frecuencia: p.frecuencia,
			productos: p.productos.map((x) => ({ ...x })),
			incluye: p.incluye,
			envio: p.envio,
			retiro: p.retiro,
			cupo: p.cupo,
			activo: p.activo
		};
	};
	const sinId = (p: Form): Editable => {
		// eslint-disable-next-line @typescript-eslint/no-unused-vars -- se descarta el id
		const { id, ...cuerpo } = p;
		return { ...cuerpo, cupo: conCupo ? cuerpo.cupo : null };
	};
	const listo = $derived(
		!!plan &&
			plan.nombre.trim().length >= 2 &&
			plan.precio >= 100 &&
			(plan.envio || plan.retiro) &&
			(plan.productos.length > 0 || plan.incluye.trim().length > 0) &&
			(!conCupo || (plan.cupo ?? 0) >= 1)
	);
	async function guardar(e: SubmitEvent) {
		e.preventDefault();
		if (!plan || !listo) return;
		const id = plan.id;
		const cuerpo = sinId(plan);
		const r = await pedir(
			id ? `/panel/suscripciones/planes/${id}` : '/panel/suscripciones/planes',
			id ? 'PUT' : 'POST',
			cuerpo
		);
		if (!r) return;
		plan = null;
		await invalidateAll();
		ui.toast({
			k: 'ok',
			title: id ? 'Plan guardado' : 'Plan creado',
			desc: cuerpo.activo
				? `${cuerpo.nombre} · ya se ve en tu tienda`
				: `${cuerpo.nombre} · pausado`
		});
	}
	async function pausar(p: Plan) {
		const antes = { ...sinIdPlan(p) };
		if (
			!(await pedir(`/panel/suscripciones/planes/${p.id}`, 'PUT', { ...antes, activo: !p.activo }))
		)
			return;
		await invalidateAll();
		ui.toast({
			k: 'ok',
			title: p.activo ? 'Plan pausado' : 'Plan activo',
			desc: p.activo
				? `${p.nombre} · no se ve en la tienda; quienes ya están siguen igual`
				: `${p.nombre} · ya se ve en tu tienda`,
			action: 'Deshacer',
			onAction: async () => {
				await pedir(`/panel/suscripciones/planes/${p.id}`, 'PUT', antes);
				await invalidateAll();
			}
		});
	}
	const sinIdPlan = (p: Plan): Editable => ({
		nombre: p.nombre,
		descripcion: p.descripcion,
		foto: p.foto,
		precio: p.precio,
		frecuencia: p.frecuencia,
		productos: p.productos,
		incluye: p.incluye,
		envio: p.envio,
		retiro: p.retiro,
		cupo: p.cupo,
		activo: p.activo
	});
	async function borrar(p: Plan) {
		if (!(await pedir(`/panel/suscripciones/planes/${p.id}`, 'DELETE'))) return;
		await invalidateAll();
		ui.toast({
			k: 'ok',
			title: 'Plan borrado',
			desc: p.nombre,
			action: 'Deshacer',
			onAction: async () => {
				await pedir('/panel/suscripciones/planes', 'POST', sinIdPlan(p));
				await invalidateAll();
			}
		});
	}

	// Productos del plan.
	let elegir = $state('');
	const opciones = $derived(
		data.productos.flatMap((p) =>
			p.variantes.length
				? p.variantes.map((v) => ({ clave: `${p.id}:${v.id}`, texto: `${p.nombre} · ${v.texto}` }))
				: [{ clave: `${p.id}:`, texto: p.nombre }]
		)
	);
	const nombreDe = (productoId: string, varianteId: string) => {
		const p = data.productos.find((x) => x.id === productoId);
		if (!p) return 'Producto que ya no está';
		const v = p.variantes.find((x) => x.id === varianteId);
		return v ? `${p.nombre} · ${v.texto}` : p.nombre;
	};
	function sumarProducto() {
		if (!plan || !elegir) return;
		const [productoId, varianteId] = elegir.split(':');
		const ya = plan.productos.find(
			(x) => x.productoId === productoId && x.varianteId === varianteId
		);
		if (ya) ya.cantidad = Math.min(99, ya.cantidad + 1);
		else plan.productos = [...plan.productos, { productoId, varianteId, cantidad: 1 }];
		elegir = '';
	}

	// ---- Suscriptores ----
	let filtro = $state<'vigentes' | 'todas'>('vigentes');
	const lista = $derived(
		filtro === 'vigentes'
			? data.suscriptores.filter((s) => s.estado !== 'cancelada')
			: data.suscriptores
	);
	const CHIP: Record<EstadoSuscripcion, Estado> = {
		activa: 'Pagado',
		pausada: 'Sin pagar',
		pendiente: 'Pago pendiente',
		cancelada: 'Cancelado'
	};
	let cancelar = $state<Suscriptor | null>(null);
	async function confirmarCancelar() {
		const s = cancelar;
		if (!s) return;
		cancelar = null;
		if (!(await pedir(`/panel/suscripciones/${s.id}/cancelar`, 'POST'))) return;
		await invalidateAll();
		ui.toast({
			k: 'ok',
			title: 'Suscripción cancelada',
			desc: `${s.nombre} · no se le cobra más y le avisamos por mail`
		});
	}
	const foto = (f: string | null, w: number) => (f ? unsplash(f, w, `&h=${w}`) : null);
	const plural = (n: number, uno: string, varios: string) => `${n} ${n === 1 ? uno : varios}`;
</script>

<Seo
	title="Suscripciones · Panel"
	description="Suscripciones"
	path="/panel/suscripciones"
	noindex
/>

<!-- eslint-disable svelte/no-navigation-without-resolve -- rutas del panel -->
<div class="pagina">
	<div class="cab" use:anim={{ t: 'rise' }}>
		<div>
			<h1>Suscripciones</h1>
			<div class="c-s">
				Planes con débito automático de Mercado Pago: cada cobro te crea el pedido solo
			</div>
		</div>
		{#if editaPlanes && data.mercadoPago}
			<button type="button" class="btn btn-primary" onclick={nuevoPlan}
				><Icon icon={I.phPlusBold} size={14} />Nuevo plan</button
			>
		{/if}
	</div>

	{#if !data.mercadoPago}
		<Alerta
			tipo="warn"
			titulo="Conectá Mercado Pago para ofrecer suscripciones"
			texto="Los planes se cobran con débito automático desde tu cuenta de Mercado Pago. Hasta que la conectes, la tienda no los muestra."
			accion="Conectar Mercado Pago"
			onaccion={() => goto('/panel/integraciones')}
		/>
	{/if}

	{#if data.planes.length || data.suscriptores.length}
		<div class="numeros" use:anim={{ t: 'rise', delay: 40 }}>
			<div class="num">
				<span class="n-v">{activas.length}</span><span class="n-l">Activas</span>
			</div>
			<div class="num">
				<span class="n-v">{pausadas.length}</span><span class="n-l">Pausadas</span>
			</div>
			<div class="num">
				<span class="n-v">{data.planes.filter((p) => p.activo).length}</span><span class="n-l"
					>Planes a la venta</span
				>
			</div>
			{#if veMontos}
				<div class="num">
					<span class="n-v">{pesos(porMes)}</span><span class="n-l">Por mes, con las activas</span>
				</div>
			{/if}
		</div>
	{/if}

	<div class="tarjeta" use:anim={{ t: 'rise', delay: 80 }}>
		<div class="t-cab"><span>Planes · {data.planes.length}</span></div>
		<div class="filas">
			{#each data.planes as p (p.id)}
				<div class="fila" class:pausado={!p.activo}>
					{#if foto(p.foto, 88)}<img src={foto(p.foto, 88)} alt="" />{:else}<span class="av"
							><Icon icon={I.phArrowsClockwise} size={20} /></span
						>{/if}
					<span class="m-txt"
						><span class="m-n"
							>{p.nombre}{#if !p.activo}<span class="tag gris">Pausado</span
								>{/if}{#if p.cupo !== null && p.suscriptores >= p.cupo}<span class="tag"
									>Sin cupo</span
								>{/if}</span
						><span class="m-e"
							>{pesos(p.precio)}
							{CADA[p.frecuencia]} · {plural(
								p.suscriptores,
								'suscripción',
								'suscripciones'
							)}{p.cupo !== null ? ` de ${p.cupo}` : ''}</span
						></span
					>
					{#if editaPlanes}
						<span class="acc">
							<button
								type="button"
								class="icon-btn"
								aria-label="{p.activo ? 'Pausar' : 'Activar'} {p.nombre}"
								onclick={() => pausar(p)}
								><Icon icon={p.activo ? I.phPauseCircle : I.phArrowsClockwise} size={16} /></button
							>
							<button
								type="button"
								class="icon-btn"
								aria-label="Editar {p.nombre}"
								onclick={() => editar(p)}><Icon icon={I.phPencilSimple} size={15} /></button
							>
							{#if !p.suscriptores}
								<button
									type="button"
									class="icon-btn"
									aria-label="Borrar {p.nombre}"
									onclick={() => borrar(p)}><Icon icon={I.phTrash} size={15} /></button
								>
							{/if}
						</span>
					{/if}
				</div>
			{:else}
				<Vacio
					hornero="caja"
					titulo="Todavía no tenés planes"
					texto="Armá una caja por mes, un café cada dos semanas o lo que vendas seguido: quien se suscribe paga solo con Mercado Pago."
				>
					{#if editaPlanes && data.mercadoPago}
						<button type="button" class="btn btn-primary" onclick={nuevoPlan}
							>Crear el primero</button
						>
					{/if}
				</Vacio>
			{/each}
		</div>
	</div>

	<div class="tarjeta" use:anim={{ t: 'rise', delay: 120 }}>
		<div class="t-cab">
			<span>Suscriptores · {lista.length}</span>
			<Segmentado
				bind:value={filtro}
				label="Qué suscripciones ver"
				opciones={[
					['vigentes', 'Vigentes'],
					['todas', 'Todas']
				]}
			/>
		</div>
		<div class="filas">
			{#each lista as s (s.id)}
				<div class="fila sus">
					<span class="m-txt"
						><span class="m-n">{s.nombre}</span><span class="m-e"
							>{s.plan} · {s.email}{s.telefono ? ` · ${s.telefono}` : ''}</span
						><span class="m-e"
							>{s.entrega === 'envio' ? `Envío a ${s.direccion}` : 'Retira en el local'}</span
						></span
					>
					<span class="dato"
						><span class="d-l">Próximo cobro</span><span class="d-v"
							>{s.estado === 'activa' ? (diaCobro(s.proximoCobro) ?? 'A confirmar') : '—'}</span
						></span
					>
					<span class="dato"
						><span class="d-l">Cobros</span><span class="d-v"
							>{s.cobros}{veMontos ? ` · ${pesos(s.precio)}` : ''}</span
						></span
					>
					<ChipEstado estado={CHIP[s.estado]} texto={ESTADO_SUSCRIPCION[s.estado]} />
					<span class="acc">
						{#if s.estado !== 'cancelada'}
							<button
								type="button"
								class="icon-btn"
								aria-label="Cancelar la suscripción de {s.nombre}"
								onclick={() => (cancelar = s)}><Icon icon={I.phXCircle} size={17} /></button
							>
						{/if}
					</span>
				</div>
			{:else}
				<div class="vacio">
					{data.planes.length
						? 'Cuando alguien se suscriba desde tu tienda, aparece acá.'
						: 'Creá un plan y compartí tu tienda: los suscriptores aparecen acá.'}
				</div>
			{/each}
		</div>
	</div>
</div>

<Modal
	bind:open={
		() => !!plan,
		(v) => {
			if (!v) plan = null;
		}
	}
	titulo={plan?.id ? 'Editar plan' : 'Nuevo plan'}
	ancho={560}
>
	{#if plan}
		<form onsubmit={guardar}>
			<div class="m-t">{plan.id ? 'Editar plan' : 'Nuevo plan'}</div>
			<div class="campos">
				<label class="field ancho"
					>Nombre<input
						class="input"
						bind:value={plan.nombre}
						maxlength="60"
						placeholder="Club de vinos, Café cada 2 semanas…"
						required
					/></label
				>
				<label class="field ancho"
					>Descripción<textarea
						class="input"
						rows="2"
						bind:value={plan.descripcion}
						maxlength="400"
						placeholder="Opcional: qué tiene de especial"></textarea></label
				>
				<label class="field"
					>Precio por período<input
						class="input"
						type="number"
						min="100"
						inputmode="numeric"
						bind:value={plan.precio}
						required
					/></label
				>
				<label class="field"
					>Se cobra<select class="select" bind:value={plan.frecuencia}
						>{#each FRECUENCIAS as [v, t] (v)}<option value={v}>{t}</option>{/each}</select
					></label
				>
				<label class="field ancho"
					>Foto (link)<input
						class="input"
						type="url"
						bind:value={plan.foto}
						maxlength="2048"
						placeholder="https://…"
					/></label
				>
			</div>

			<div class="m-sub">Qué incluye</div>
			{#if plan.productos.length}
				<div class="incluidos">
					{#each plan.productos as x, i (x.productoId + x.varianteId)}
						<div class="inc">
							<span class="inc-n">{nombreDe(x.productoId, x.varianteId)}</span>
							<input
								class="input chico"
								type="number"
								min="1"
								max="99"
								inputmode="numeric"
								aria-label="Cantidad de {nombreDe(x.productoId, x.varianteId)}"
								bind:value={x.cantidad}
							/>
							<button
								type="button"
								class="icon-btn"
								aria-label="Sacar {nombreDe(x.productoId, x.varianteId)}"
								onclick={() => plan && (plan.productos = plan.productos.filter((_, j) => j !== i))}
								><Icon icon={I.phTrash} size={15} /></button
							>
						</div>
					{/each}
				</div>
			{/if}
			<div class="sumar-prod">
				<select class="select" bind:value={elegir} aria-label="Producto del catálogo">
					<option value="">Sumar un producto del catálogo…</option>
					{#each opciones as o (o.clave)}<option value={o.clave}>{o.texto}</option>{/each}
				</select>
				<button type="button" class="btn btn-secondary" disabled={!elegir} onclick={sumarProducto}
					>Sumar</button
				>
			</div>
			<label class="field ancho texto"
				>Contalo con tus palabras<input
					class="input"
					bind:value={plan.incluye}
					maxlength="400"
					placeholder="6 botellas elegidas por la sommelier"
				/></label
			>

			<div class="opcion">
				<span class="o-txt"
					><span class="o-t">Con envío</span><span class="o-d">El precio ya incluye el envío</span
					></span
				>
				<Interruptor bind:checked={plan.envio} label="Con envío" />
			</div>
			<div class="opcion">
				<span class="o-txt"
					><span class="o-t">Retiro en el local</span><span class="o-d"
						>Quien se suscribe elige cómo lo recibe</span
					></span
				>
				<Interruptor bind:checked={plan.retiro} label="Retiro en el local" />
			</div>
			<div class="opcion">
				<span class="o-txt"
					><span class="o-t">Cupo</span><span class="o-d">Hasta cuántas suscripciones tomás</span
					></span
				>
				{#if conCupo}<input
						class="input chico"
						type="number"
						min="1"
						inputmode="numeric"
						aria-label="Cupo"
						bind:value={plan.cupo}
					/>{/if}
				<Interruptor bind:checked={conCupo} label="Cupo" />
			</div>
			<div class="opcion sin-borde">
				<span class="o-txt"
					><span class="o-t">A la venta</span><span class="o-d"
						>Pausado, no se ve en la tienda; quienes ya están siguen igual</span
					></span
				>
				<Interruptor bind:checked={plan.activo} label="A la venta" />
			</div>
			{#if plan.id}
				<p class="nota">
					Cambiar el precio vale para quienes se suscriban desde ahora: a quienes ya están se les
					sigue cobrando lo que autorizaron en Mercado Pago.
				</p>
			{/if}
			<div class="m-acc">
				<button type="button" class="btn btn-secondary" onclick={() => (plan = null)}
					>Cancelar</button
				>
				<button class="btn btn-primary" disabled={!listo}>Guardar</button>
			</div>
		</form>
	{/if}
</Modal>

<Modal
	bind:open={
		() => !!cancelar,
		(v) => {
			if (!v) cancelar = null;
		}
	}
	titulo="Cancelar la suscripción"
	ancho={440}
>
	{#if cancelar}
		<div class="m-t">¿Cancelar la suscripción de {cancelar.nombre}?</div>
		<p class="nota">
			Mercado Pago deja de cobrarle {cancelar.plan} y le avisamos por mail. No se puede deshacer: para
			volver, se suscribe de nuevo desde la tienda.
		</p>
		<div class="m-acc">
			<button type="button" class="btn btn-secondary" onclick={() => (cancelar = null)}
				>Dejarla como está</button
			>
			<button type="button" class="btn btn-danger" onclick={confirmarCancelar}
				>Cancelar suscripción</button
			>
		</div>
	{/if}
</Modal>

<style>
	.pagina {
		display: flex;
		flex-direction: column;
		gap: 18px;
	}
	.cab {
		display: flex;
		align-items: flex-end;
		justify-content: space-between;
		gap: 12px;
		flex-wrap: wrap;
	}
	h1 {
		margin: 4px 0 0;
		font-size: 30px;
		letter-spacing: -0.035em;
		font-weight: 600;
	}
	.c-s {
		margin-top: 2px;
		font-size: 15px;
		color: var(--mu);
		text-wrap: pretty;
	}
	.numeros {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
		gap: 10px;
	}
	.num {
		padding: 14px 16px;
		border-radius: 18px;
		background: var(--sf);
		box-shadow: var(--shd);
		display: flex;
		flex-direction: column;
		gap: 2px;
	}
	.n-v {
		font-size: 24px;
		font-weight: 600;
		letter-spacing: -0.03em;
	}
	.n-l {
		font-size: 12.5px;
		color: var(--mu);
	}
	.tarjeta {
		padding: 18px;
		border-radius: 22px;
		background: var(--sf);
		box-shadow: var(--shd);
	}
	.t-cab {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 10px;
		flex-wrap: wrap;
		font-size: 16px;
		font-weight: 600;
	}
	.filas {
		margin-top: 10px;
		display: flex;
		flex-direction: column;
	}
	.fila {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 11px 0;
		border-bottom: 1px solid var(--ln);
		flex-wrap: wrap;
	}
	.fila:last-child {
		border-bottom: 0;
	}
	.fila.pausado {
		opacity: 0.65;
	}
	.fila img,
	.av {
		width: 44px;
		height: 44px;
		flex: none;
		border-radius: 12px;
		object-fit: cover;
	}
	.av {
		display: grid;
		place-items: center;
		background: var(--chip);
		color: var(--mu);
	}
	.m-txt {
		flex: 1 1 200px;
		min-width: 0;
	}
	.m-n {
		display: flex;
		align-items: center;
		gap: 6px;
		flex-wrap: wrap;
		font-size: 14px;
		font-weight: 600;
	}
	.m-e {
		display: block;
		font-size: 12.5px;
		color: var(--mu);
		overflow-wrap: anywhere;
	}
	.dato {
		display: flex;
		flex-direction: column;
		min-width: 104px;
	}
	.d-l {
		font-size: 11.5px;
		color: var(--mu);
	}
	.d-v {
		font-size: 13.5px;
		font-weight: 600;
	}
	.tag {
		font-size: 11px;
		font-weight: 600;
		padding: 2px 8px;
		border-radius: 999px;
		background: rgba(210, 88, 47, 0.12);
		color: #b0441d;
	}
	.tag.gris {
		background: var(--chip);
		color: var(--mu);
	}
	.acc {
		display: flex;
		gap: 4px;
		margin-left: auto;
		min-width: 44px;
		justify-content: flex-end;
	}
	.vacio {
		padding: 20px 0 8px;
		text-align: center;
		color: var(--mu);
		font-size: 14px;
		text-wrap: pretty;
	}
	.m-t {
		font-size: 21px;
		font-weight: 600;
		letter-spacing: -0.02em;
		text-wrap: balance;
	}
	.campos {
		margin-top: 16px;
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 10px;
	}
	.campos .ancho {
		grid-column: 1 / -1;
	}
	textarea.input {
		resize: vertical;
		min-height: 64px;
		font: inherit;
	}
	.m-sub {
		margin-top: 16px;
		font-size: 13px;
		font-weight: 600;
		color: var(--mu);
	}
	.incluidos {
		margin-top: 8px;
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.inc {
		display: flex;
		align-items: center;
		gap: 8px;
	}
	.inc-n {
		flex: 1;
		min-width: 0;
		font-size: 14px;
		overflow-wrap: anywhere;
	}
	.input.chico {
		width: 78px;
		flex: none;
	}
	.sumar-prod {
		margin-top: 8px;
		display: flex;
		gap: 8px;
	}
	.sumar-prod .select {
		flex: 1;
		min-width: 0;
	}
	.field.texto {
		margin-top: 10px;
		display: flex;
		flex-direction: column;
	}
	.opcion {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 12px 0;
		border-bottom: 1px solid var(--ln);
	}
	.opcion.sin-borde {
		border-bottom: 0;
	}
	.o-txt {
		flex: 1;
		min-width: 0;
	}
	.o-t {
		display: block;
		font-size: 14.5px;
		font-weight: 600;
	}
	.o-d {
		display: block;
		font-size: 12.5px;
		color: var(--mu);
		text-wrap: pretty;
	}
	.nota {
		margin: 12px 0 0;
		font-size: 13.5px;
		line-height: 1.5;
		color: var(--mu);
		text-wrap: pretty;
	}
	.m-acc {
		margin-top: 20px;
		display: flex;
		gap: 8px;
		justify-content: flex-end;
		flex-wrap: wrap;
	}
	@media (max-width: 759px) {
		.campos {
			grid-template-columns: minmax(0, 1fr);
		}
		.fila.sus .dato {
			min-width: 0;
			flex: 1 1 40%;
		}
	}
</style>
