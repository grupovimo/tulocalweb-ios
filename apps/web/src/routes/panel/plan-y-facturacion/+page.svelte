<script lang="ts">
	import { invalidateAll, replaceState } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import Icon from '#lib/components/Icon.svelte';
	import MarcaIcono from '#lib/components/MarcaIcono.svelte';
	import Modal from '#lib/components/Modal.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import type { IconData } from '#lib/icons.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { pesos } from '#lib/panel/formato.ts';
	import { MOTIVOS } from '#lib/panel/permisos.ts';
	import {
		BENEFICIOS,
		conDescuento,
		PLANES,
		planVigente,
		precioAnual,
		type IdPlan
	} from '#lib/panel/planes.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/**
	 * Plan y facturación (design/PanelCuenta.dc.html, sPlan) con el abono de §56: elegir un plan no cobra
	 * nada; el resumen (el «pedido» de design/Mostrador v4.dc.html) muestra qué se paga hoy, línea por
	 * línea, desde cuándo rige y cuánto se cobra después, y recién «Confirmar» pasa por Mercado Pago.
	 * Las cuentas vienen hechas de la API (abono.Cotizar), las mismas con las que se cobra.
	 */
	let { data } = $props();
	const ui = getUi();
	type Ciclo = 'mes' | 'anio';
	type Plan = (typeof PLANES)[number];

	// Sin cortes raros: «14 de noviembre» va siempre en un renglón.
	const fecha = (iso: string) =>
		new Date(iso)
			.toLocaleDateString('es-AR', {
				day: 'numeric',
				month: 'long',
				timeZone: 'America/Argentina/Buenos_Aires'
			})
			.replaceAll(' ', '\u00a0');
	const indice = (id: string) => PLANES.findIndex((p) => p.id === id);
	const nombre = (id: string) => PLANES[indice(id)]?.nombre ?? id;
	const actual = $derived(PLANES[indice(data.actual)]);
	// Durante la prueba gratis, una tienda en Gratis usa Pro: los límites son los de Pro.
	const probando = $derived(data.enPrueba && data.actual === 'gratis');
	const vigente = $derived(PLANES[indice(planVigente(data.actual, data.enPrueba))]);
	const programado = $derived(data.abono.programado);

	// ---- Lo elegido (no cobra nada hasta confirmar) ----
	// Se reasignan al elegir; con datos nuevos (después de confirmar) vuelven al plan vigente.
	let ciclo = $derived<Ciclo>(data.ciclo);
	let elegido = $derived<IdPlan>(data.actual);
	const anual = $derived(ciclo === 'anio');
	const clave = (p: string, c: Ciclo) => `${p}/${c}`;
	const cot = $derived(
		data.cotizaciones[clave(elegido, elegido === 'gratis' ? data.ciclo : ciclo)]
	);
	// Sin cambio: el plan vigente (o el ya programado) con su ciclo.
	const sinCambio = $derived(!cot || cot.tipo === 'actual' || cot.tipo === 'programado');
	const permitido = (p: Plan) =>
		!!data.cotizaciones[clave(p.id, p.id === 'gratis' ? data.ciclo : ciclo)];
	function elegir(p: Plan) {
		if (!permitido(p)) {
			ui.toast({
				k: 'info',
				title: 'Tenés un cambio programado',
				desc: 'Deshacelo para elegir otro plan.'
			});
			return;
		}
		elegido = p.id;
		if (innerWidth < 1280)
			document.getElementById('resumen')?.scrollIntoView({ behavior: 'smooth', block: 'start' });
	}

	// ---- El resumen ----
	const porCiclo = (n: number, c: Ciclo) => `${pesos(n)} ${c === 'anio' ? 'por año' : 'por mes'}`;
	const NADA_HOY = $derived.by(() => {
		if (!cot) return '';
		switch (cot.tipo) {
			case 'prueba':
				return `Estás en la prueba gratis: no pagás nada hasta el ${fecha(data.finPrueba)}.`;
			case 'baja':
			case 'ciclo':
				return `Ya pagaste hasta el ${fecha(cot.hasta!)}: seguís con ${actual.nombre} hasta ese día.`;
			case 'gratis':
				return cot.hasta
					? `Ya pagaste hasta el ${fecha(cot.hasta)}: seguís con ${actual.nombre} hasta ese día.`
					: 'El plan Gratis no tiene cobros.';
			case 'mantener':
				return `Deshacés el cambio programado y seguís con ${actual.nombre}.`;
			default:
				return '';
		}
	});
	const RIGE = $derived(
		!cot
			? ''
			: new Date(cot.rige) <= new Date(Date.now() + 60_000)
				? 'hoy, al confirmar'
				: `el ${fecha(cot.rige)}`
	);
	const COMO = $derived.by(() => {
		switch (cot?.pago) {
			case 'suscripcion':
				return 'Cargás la tarjeta en Mercado Pago y queda el débito automático. Lo das de baja cuando quieras.';
			case 'diferencia':
				return 'Pagás la diferencia en Mercado Pago (tarjeta, dinero en cuenta o efectivo). Tu débito se actualiza solo.';
			case 'actualizar':
				return 'Actualizamos el monto de tu débito de Mercado Pago. No tenés que cargar nada.';
			case 'cancelar':
				return 'Damos de baja tu débito de Mercado Pago: no te cobramos más.';
			default:
				return '';
		}
	});
	const BOTON = $derived.by(() => {
		if (!cot || sinCambio) return cot?.tipo === 'programado' ? 'Ya está programado' : 'Es tu plan';
		if (cot.pago === 'diferencia' || (cot.pago === 'suscripcion' && cot.hoy > 0))
			return `Pagar ${pesos(cot.hoy)} con Mercado Pago`;
		if (cot.pago === 'suscripcion') return 'Seguir a Mercado Pago';
		if (cot.tipo === 'mantener') return 'Deshacer el cambio';
		return elegido === 'gratis' ? 'Pasar a Gratis' : 'Confirmar el cambio';
	});
	const NUEVO = $derived(PLANES[indice(elegido)]);
	const BOTON_CORTO = $derived(
		BOTON.replace(/^Pagar .* con Mercado Pago$/, 'Pagar con Mercado Pago')
	);

	// ---- Confirmar ----
	let enviando = $state(false);
	async function pedir(url: string, method: string, cuerpo?: unknown) {
		const r = await fetch(url, {
			method,
			headers: cuerpo === undefined ? {} : { 'content-type': 'application/json' },
			body: cuerpo === undefined ? undefined : JSON.stringify(cuerpo)
		});
		if (!r.ok) {
			const p = (await r.json().catch(() => null)) as { message?: string } | null;
			ui.toast({ k: 'err', title: 'No se pudo', desc: p?.message ?? 'Probá de nuevo.' });
		}
		return r.ok ? r : null;
	}
	async function confirmarCambio() {
		if (!cot || sinCambio || enviando) return;
		const q = cot;
		const antes = { plan: data.actual, ciclo: data.ciclo };
		enviando = true;
		const r = await pedir('/panel/plan', 'PUT', { plan: q.plan, ciclo: q.ciclo });
		if (!r) return void (enviando = false);
		// Hay que pasar por Mercado Pago: cargar la tarjeta del débito o pagar la diferencia.
		if (r.status === 200) {
			location.href = ((await r.json()) as { url: string }).url;
			return;
		}
		await invalidateAll();
		enviando = false;
		const n = nombre(q.plan);
		if (q.tipo === 'baja')
			ui.toast({
				k: 'ok',
				title: `Pasás a ${n} el ${fecha(q.rige)}`,
				desc: `Hasta ahí seguís con ${nombre(antes.plan)}. Desde ese día pagás ${porCiclo(q.porCiclo, q.ciclo)}.`,
				action: 'Deshacer',
				onAction: async () => {
					await pedir('/panel/plan', 'PUT', antes);
					await invalidateAll();
				}
			});
		else if (q.tipo === 'gratis')
			ui.toast({
				k: 'info',
				title: q.hasta ? `Pasás a Gratis el ${fecha(q.hasta)}` : 'Pasaste al plan Gratis',
				desc: 'Conservás productos, clientes y pedidos. No te cobramos más.'
			});
		else if (q.tipo === 'mantener')
			ui.toast({ k: 'ok', title: `Seguís con ${n}`, desc: 'Deshiciste el cambio programado.' });
		else
			ui.toast({
				k: 'ok',
				title: `Ya tenés el plan ${n}`,
				desc: q.cobro
					? `El próximo cobro es el ${fecha(q.cobro.fecha)}: ${pesos(q.cobro.importe)}.`
					: ''
			});
	}

	// La vuelta de Mercado Pago: el servidor ya confirmó el pago (o queda el aviso del webhook).
	onMount(() => {
		const q = page.url.searchParams;
		// Desde Configuración (o un link) llega el plan ya elegido.
		const pre = PLANES.find((p) => p.id === q.get('plan'));
		if (pre && pre.id !== data.actual) {
			elegir(pre);
			replaceState(resolve('/panel/plan-y-facturacion'), page.state);
			return;
		}
		const pago = q.get('pago');
		const vuelta = q.get('suscripcion') === 'vuelta';
		if (!pago && !vuelta) return;
		if (pago === 'error')
			ui.toast({
				k: 'err',
				title: 'El pago no se completó',
				desc: 'No te cobramos nada. Podés probar de nuevo.'
			});
		else if (pago === 'pendiente')
			ui.toast({
				k: 'info',
				title: 'Tu pago está pendiente',
				desc: 'Cuando Mercado Pago lo acredite, tu plan cambia solo.'
			});
		else
			ui.toast({
				k: 'ok',
				title:
					data.actual === 'gratis'
						? 'Estamos confirmando tu pago'
						: `¡Listo! Estás en ${actual.nombre}`,
				desc:
					data.actual === 'gratis'
						? 'En unos segundos se activa tu plan.'
						: 'Gracias por elegir Tu Local Web.'
			});
		replaceState(resolve('/panel/plan-y-facturacion'), page.state);
	});

	// ---- Uso ----
	const tope = $derived(vigente.hornero + data.uso.packs * 2000);
	/**
	 * Con la API real las respuestas de Tincho todavía no se cuentan (Go devuelve 0) y el pack de
	 * respuestas no se cobra en ninguna factura: no se muestra un medidor en cero ni se vende algo
	 * que no se cobra. En la demo siguen como en el handoff.
	 */
	const demo = $derived(!!page.data.demo);
	const USO = $derived(
		demo
			? ([
					['Productos', data.uso.productos, null],
					['Personas en el equipo', data.uso.personas, vigente.personas],
					['Respuestas de Tincho con IA', data.uso.hornero, tope],
					['Respuestas automáticas sin IA', data.uso.sinIa, null]
				] as const)
			: ([
					['Productos', data.uso.productos, null],
					['Personas en el equipo', data.uso.personas, vigente.personas]
				] as const)
	);
	async function pack() {
		if (!(await pedir('/panel/cuenta/pack', 'POST'))) return;
		await invalidateAll();
		ui.toast({
			k: 'ai',
			title: '2.000 respuestas más para Tincho',
			desc: 'Se suman a este mes · $ 4.900 en tu próxima factura',
			action: 'Deshacer',
			onAction: async () => {
				await pedir('/panel/cuenta/pack', 'DELETE');
				await invalidateAll();
			}
		});
	}

	// ---- Frenar y cancelar: pasan por el mismo resumen (lo pagado no se pierde) ----
	function aGratis() {
		elegido = 'gratis';
		document.getElementById('resumen')?.scrollIntoView({ behavior: 'smooth', block: 'start' });
	}
	function reactivar() {
		elegido = data.actual;
		ciclo = data.ciclo;
		document.getElementById('resumen')?.scrollIntoView({ behavior: 'smooth', block: 'start' });
	}
	const hastaGratis = $derived(data.cotizaciones[clave('gratis', data.ciclo)]?.hasta ?? null);
	let cancelando = $state(false);
	let paso = $state(0);
	let motivo = $state(0);
	// La oferta para quedarse por motivo (§57): el descuento se cobra de verdad (baja el débito de Mercado
	// Pago) y va una vez por tienda; si ya se usó, o el motivo es otro, alguien del equipo le escribe.
	const debito = $derived(
		data.cotizaciones[clave(data.actual, data.ciclo)]?.porCiclo ?? actual.precio
	);
	const cadaCobro = $derived(data.ciclo === 'anio' ? 'por año' : 'por mes');
	type Oferta = readonly [string, string, IconData, string, string, string];
	const CONTACTO: Oferta = [
		'Contanos qué te falta',
		'Muchas funciones se pueden activar o están por salir. Una persona del equipo te escribe hoy.',
		I.phChatsCircleBold,
		'Una persona del equipo te escribe',
		'Hoy, por WhatsApp o por correo',
		'Que me contacten'
	];
	const OFERTAS = $derived<Oferta[]>([
		data.beneficioUsado
			? CONTACTO
			: [
					'Te dejamos 3 cobros a mitad de precio',
					`Sabemos que no siempre es fácil. Seguí con ${actual.nombre} por ${pesos(conDescuento(debito, BENEFICIOS[0]))} ${cadaCobro} en tus próximos 3 cobros. Después vuelve a ${pesos(debito)}.`,
					I.phPercentBold,
					'50 % off en 3 cobros',
					'Baja el débito desde el próximo cobro',
					'Quiero el descuento'
				],
		CONTACTO,
		[
			'Mejor pausala',
			'Tu tienda queda guardada y la reactivás cuando quieras, sin cargar nada de nuevo.',
			I.phPauseCircleBold,
			'Pausa hasta 3 meses',
			'Gratis: pasás al plan Gratis y conservás todo',
			'Pausar en vez de cancelar'
		],
		[
			'Antes de irte, hablemos',
			'Contanos qué te ofrecen en la otra plataforma: vemos cómo ayudarte a quedarte o a mudarte sin perder nada.',
			I.phScalesBold,
			'Una persona del equipo te escribe',
			'Hoy, por WhatsApp o por correo',
			'Que me contacten'
		],
		data.beneficioUsado
			? CONTACTO
			: [
					'Tu próximo cobro a mitad de precio',
					`Si algo no anduvo bien, queremos saberlo. Mientras tanto, tu próximo cobro es de ${pesos(conDescuento(debito, BENEFICIOS[4]))} en vez de ${pesos(debito)}.`,
					I.phHandHeartBold,
					'50 % off en el próximo cobro',
					'Sin compromiso',
					'Me quedo con el descuento'
				]
	]);
	const oferta = $derived(OFERTAS[motivo]);
	function abrirCancelar() {
		paso = 0;
		motivo = 0;
		cancelando = true;
	}
	async function aceptarOferta() {
		if (motivo === 2) {
			cancelando = false;
			return aGratis();
		}
		const r = await pedir('/panel/cuenta/cancelacion', 'POST', { motivo, accion: 'oferta' });
		if (!r) return;
		cancelando = false;
		const b = (
			(await r.json().catch(() => ({}))) as { beneficio?: { pct: number; cobros: number } }
		).beneficio;
		if (b) await invalidateAll();
		ui.toast({
			k: 'ok',
			title: b ? '¡Gracias por quedarte!' : 'Listo, te escribimos hoy',
			desc: b
				? `${b.pct} % off en ${b.cobros === 1 ? 'tu próximo cobro' : `tus próximos ${b.cobros} cobros`}: ${pesos(conDescuento(debito, b))} en vez de ${pesos(debito)}.`
				: 'Una persona del equipo te contacta por WhatsApp o por correo.'
		});
	}
	async function confirmar() {
		// El motivo queda registrado; la baja es pasar a Gratis (el débito se cancela, lo pagado sigue).
		if (!(await pedir('/panel/cuenta/cancelacion', 'POST', { motivo, accion: 'cancelar' }))) return;
		const r = await pedir('/panel/plan', 'PUT', { plan: 'gratis', ciclo: data.ciclo });
		cancelando = false;
		if (!r) return;
		await invalidateAll();
		ui.toast({
			k: 'warn',
			title: 'Plan cancelado',
			desc: hastaGratis
				? `Tu tienda sigue con todo hasta el ${fecha(hastaGratis)}. No te cobramos más.`
				: 'Pasaste al plan Gratis. No te cobramos más.'
		});
	}
	const CONSECUENCIAS = $derived([
		[
			I.phCalendarCheck,
			hastaGratis
				? `Seguís usando ${actual.nombre} hasta el ${fecha(hastaGratis)}`
				: 'Pasás al plan Gratis ahora',
			'#2F8F63'
		],
		[I.phStorefront, 'Tu tienda sigue online con el plan Gratis', 'var(--mu)'],
		[I.phDatabase, 'Conservás productos, clientes y pedidos', 'var(--mu)'],
		[I.phXCircle, 'No te volvemos a cobrar', '#2F8F63']
	] as const);
</script>

<Seo
	title="Plan y facturación · Panel"
	description="Tu plan de Tu Local Web"
	path="/panel/plan-y-facturacion"
	noindex
/>

<div class="pagina">
	<div class="cab" use:anim={{ t: 'rise' }}>
		<div class="c-txt">
			<h1>Plan y facturación</h1>
			<div class="c-s">
				{#if probando}
					Estás probando <b>el plan Pro gratis</b> hasta el {fecha(data.finPrueba)}. Después seguís
					en Gratis, salvo que elijas&nbsp;otro.
				{:else}
					Estás en <b
						>plan {actual.nombre}{data.actual === 'gratis'
							? ''
							: data.ciclo === 'anio'
								? ' anual'
								: ' mensual'}</b
					>{data.enPrueba ? ' (en prueba)' : ''} ·
					{data.cancelado
						? `se da de baja el ${fecha(data.cancelado.hasta)}`
						: data.actual === 'gratis'
							? 'sin cobros'
							: data.proximoCobro
								? `próximo cobro el ${fecha(data.proximoCobro)}`
								: 'sin débito cargado'}
				{/if}
			</div>
		</div>
		<div class="ciclos" role="group" aria-label="Ciclo">
			<button type="button" aria-pressed={!anual} onclick={() => (ciclo = 'mes')}>Mensual</button>
			<button type="button" aria-pressed={anual} onclick={() => (ciclo = 'anio')}
				>Anual<span>−25 %</span></button
			>
		</div>
	</div>

	{#if programado}
		<div class="programado" use:anim={{ t: 'rise' }}>
			<Icon icon={I.phCalendarCheckFill} size={22} style="color:#b0441d;flex:none" />
			<div>
				<b
					>{programado.plan === 'gratis'
						? `Pasás al plan Gratis el ${fecha(programado.desde)}`
						: `Pasás a ${nombre(programado.plan)} ${programado.ciclo === 'anio' ? 'anual' : 'mensual'} el ${fecha(programado.desde)}`}</b
				>
				<span>Hasta ese día seguís con {actual.nombre}, que ya está pago.</span>
			</div>
			<button type="button" onclick={reactivar}>Deshacer</button>
		</div>
	{/if}

	{#if data.beneficio}
		<div class="programado beneficio" use:anim={{ t: 'rise' }}>
			<Icon icon={I.phPercentBold} size={22} style="color:#1f8a5b;flex:none" />
			<div>
				<b
					>{data.beneficio.pct} % off en {data.beneficio.cobros === 1
						? 'tu próximo cobro'
						: `tus próximos ${data.beneficio.cobros} cobros`}</b
				>
				<span
					>Pagás {pesos(conDescuento(debito, data.beneficio))} en vez de {pesos(debito)}
					{cadaCobro}. Después vuelve solo al precio del plan.</span
				>
			</div>
		</div>
	{/if}

	<div class="compra">
		<div class="planes" use:anim={{ t: 'rise', delay: 50 }} role="radiogroup" aria-label="Planes">
			{#each PLANES as p (p.id)}
				{@const cur = p.id === data.actual}
				{@const sel = p.id === elegido}
				{@const oscuro = p.id === 'pro'}
				{@const mensual = anual ? precioAnual(p.precio) : p.precio}
				{@const bloqueado = !permitido(p)}
				<button
					type="button"
					role="radio"
					aria-checked={sel}
					class="plan"
					class:oscuro
					class:cur
					class:sel
					class:bloqueado
					onclick={() => elegir(p)}
				>
					{#if cur}<span class="tuyo">Tu plan</span>{:else if p.id === 'pro'}<span
							class="tuyo popular">El más elegido</span
						>{/if}
					<span class="p-cab">
						<span>
							<span class="p-n">{p.nombre}</span>
							<span class="p-d">{p.lema}</span>
						</span>
						<span class="marca" aria-hidden="true"
							>{#if sel}<Icon icon={I.phCheckBold} size={14} />{/if}</span
						>
					</span>
					<span class="p-precio">
						<b>{p.precio ? pesos(mensual) : 'Gratis'}</b>{#if p.precio}<span>/ mes</span>{/if}
					</span>
					<span class="p-anual"
						>{#if !p.precio}Para siempre, sin tarjeta{:else if anual}{pesos(mensual * 12)} por año · ahorrás
							{pesos((p.precio - mensual) * 12)}{:else}IVA incluido · sin permanencia{/if}</span
					>
					<span class="p-incluye">
						{#each p.incluye as f (f)}<span><Icon icon={I.phCheckBold} size={14} />{f}</span>{/each}
					</span>
					<span class="p-btn" class:actual={cur && !sel}
						>{sel
							? cur
								? 'Tu plan actual'
								: 'Elegido'
							: cur
								? 'Volver a tu plan'
								: indice(p.id) > indice(data.actual)
									? `Elegir ${p.nombre}`
									: `Bajar a ${p.nombre}`}</span
					>
				</button>
			{/each}
		</div>

		<aside class="resumen" id="resumen" use:anim={{ t: 'rise', delay: 80 }} aria-live="polite">
			<div class="r-cab">
				<Icon icon={I.phShoppingCartSimpleBold} size={18} />
				<span>Resumen del cambio</span>
			</div>
			{#if cot && !sinCambio}
				<div class="r-de-a">
					<span class="chip-plan">{actual.nombre}</span>
					<Icon icon={I.phArrowRightBold} size={14} style="color:var(--fa)" />
					<span class="chip-plan nuevo"
						>{NUEVO.nombre}{elegido === 'gratis'
							? ''
							: cot.ciclo === 'anio'
								? ' anual'
								: ' mensual'}</span
					>
				</div>

				<div class="r-bloque">
					<div class="r-t">Hoy pagás</div>
					{#each cot.lineas as l (l.texto)}
						<div class="r-linea" class:credito={l.importe < 0}>
							<span>{l.texto}</span><span
								>{l.importe < 0 ? '−' : ''}{pesos(Math.abs(l.importe))}</span
							>
						</div>
					{:else}
						<p class="r-nada">{NADA_HOY}</p>
					{/each}
					<div class="r-total">
						<span>Total hoy</span>
						<b>{pesos(cot.hoy)}</b>
					</div>
					<div class="r-iva">
						IVA incluido{cot.tipo === 'sube' ? ' · sólo por los días que quedan' : ''}
					</div>
				</div>

				<div class="r-bloque">
					<div class="r-t">Después</div>
					{#if cot.cobro}
						<div class="r-dato">
							<Icon icon={I.phCalendarCheck} size={17} />
							<span
								>{cot.tipo === 'sube' ||
								cot.tipo === 'baja' ||
								cot.tipo === 'ciclo' ||
								cot.tipo === 'mantener'
									? 'Desde el próximo cobro'
									: 'Primer cobro'}, el <b>{fecha(cot.cobro.fecha)}</b>:
								<b>{porCiclo(cot.porCiclo, cot.ciclo)}</b>{#if cot.ciclo === 'anio'}
									<span class="r-eq"
										>(equivale a {pesos(Math.round(cot.porCiclo / 12))} por mes)</span
									>{/if}</span
							>
						</div>
						<div class="r-dato">
							<Icon icon={I.phArrowsClockwise} size={17} />
							<span
								>Débito automático {cot.ciclo === 'anio' ? 'cada año' : 'cada mes'}. Sin
								permanencia: lo das de baja cuando quieras.</span
							>
						</div>
					{:else}
						<div class="r-dato">
							<Icon icon={I.phCalendarCheck} size={17} />
							<span
								>{cot.hasta ? `Desde el ${fecha(cot.hasta)}` : 'Desde hoy'}:
								<b>plan Gratis, {pesos(0)}</b>. Sin cobros.</span
							>
						</div>
					{/if}
					<div class="r-dato">
						<Icon icon={I.phClock} size={17} />
						<span
							>{elegido === 'gratis' ? 'Pasás a Gratis' : `Tenés ${NUEVO.nombre}`}:
							<b>{RIGE}</b></span
						>
					</div>
				</div>

				{#if elegido !== 'gratis'}
					<div class="r-bloque">
						<div class="r-t">Con {NUEVO.nombre} tenés</div>
						<ul class="r-incluye">
							{#each NUEVO.incluye as f (f)}<li>
									<Icon icon={I.phCheckBold} size={13} />{f}
								</li>{/each}
						</ul>
					</div>
				{/if}

				{#if COMO}
					<div class="r-como">
						{#if cot.pago === 'suscripcion' || cot.pago === 'diferencia'}<MarcaIcono
								k="mp"
								size={30}
								radius={9}
							/>{:else}<span class="r-ico"><Icon icon={I.phShieldCheckBold} size={16} /></span>{/if}
						<span>{COMO}</span>
					</div>
				{/if}

				<button
					type="button"
					class="r-btn"
					class:roja={elegido === 'gratis'}
					disabled={enviando}
					onclick={confirmarCambio}
				>
					{#if cot.pago === 'suscripcion' || cot.pago === 'diferencia'}<Icon
							icon={I.phLockSimpleFill}
							size={15}
						/>{/if}
					{enviando ? 'Un momento…' : BOTON}
				</button>
				<button
					type="button"
					class="r-volver"
					onclick={() => ((elegido = data.actual), (ciclo = data.ciclo))}
					>Seguir con {actual.nombre}</button
				>
			{:else}
				<div class="r-vacio">
					<img src="/hornero/portapapeles.webp" alt="" width="64" height="64" />
					<b
						>{cot?.tipo === 'programado'
							? 'Ese cambio ya está programado'
							: 'Elegí un plan para ver el detalle'}</b
					>
					<span
						>{cot?.tipo === 'programado'
							? `Rige el ${fecha(cot.rige)}.`
							: 'Te mostramos cuánto pagás hoy, desde cuándo y cuánto después. Nada se cobra hasta que confirmes.'}</span
					>
				</div>
				{#if data.actual !== 'gratis' && data.proximoCobro && !programado}
					<div class="r-actual">
						<span>Tu próximo cobro</span>
						<b>{fecha(data.proximoCobro)}</b>
						<span
							>{pesos(conDescuento(debito, data.beneficio))}{data.ciclo === 'anio'
								? ' por año'
								: ' por mes'}{data.beneficio ? ` (con ${data.beneficio.pct} % off)` : ''}</span
						>
					</div>
				{/if}
			{/if}
			<div class="r-pie">
				<Icon icon={I.phShieldCheck} size={14} />Precios en pesos argentinos, IVA incluido. Tu Local
				Web nunca ve tu tarjeta.
			</div>
		</aside>
	</div>

	{#if cot && !sinCambio}
		<div class="barra-movil">
			<span><small>Hoy pagás</small><b>{pesos(cot.hoy)}</b></span>
			<button
				type="button"
				class:roja={elegido === 'gratis'}
				disabled={enviando}
				onclick={confirmarCambio}>{enviando ? 'Un momento…' : BOTON_CORTO}</button
			>
		</div>
	{/if}

	<div class="dos" use:anim={{ t: 'rise', delay: 100 }}>
		<div class="tarjeta uso">
			<div class="u-cab">
				<span>Uso de este mes</span>
				{#if demo}
					<button type="button" onclick={pack}
						><Icon icon={I.phSparkleFill} size={13} style="color:#b0441d" />+2.000 respuestas · $
						4.900</button
					>
				{/if}
			</div>
			{#each USO as [t, v, max] (t)}
				{@const w = max ? Math.min(100, (v / max) * 100) : v ? 18 : 0}
				<div>
					<div class="u-f">
						<span>{t}</span><span
							><b>{v.toLocaleString('es-AR')}</b> de {max
								? max.toLocaleString('es-AR')
								: 'ilimitado'}</span
						>
					</div>
					<div class="u-b">
						<div
							style:width="{w}%"
							style:background={w >= 100 ? '#B83A2E' : w > 80 ? '#E3A93B' : '#D2582F'}
						></div>
					</div>
				</div>
			{/each}
		</div>
		<div class="tarjeta medio">
			<div class="u-cab">
				<span>Medio de pago</span>
			</div>
			{#if data.medio}
				<div class="tarjeta-credito">
					<span class="visa"><Icon icon={I.logosVisa} size={30} /></span>
					<div>
						<div class="t-n">•••• {data.medio.ultimos}</div>
						<div class="t-v">Vence {data.medio.vence} · débito automático</div>
					</div>
				</div>
			{:else}
				<!-- El «Cambiar» de antes sólo decía que no se podía: sin medio, se explica cuándo se pide. -->
				<p class="sin-dato">
					Todavía no cargaste ninguno. Te lo pedimos con Mercado Pago cuando elijas un plan pago.
				</p>
			{/if}
			<div class="fact">
				<span>Datos de facturación</span>
				{#if data.facturacion.cuit}
					<span
						><b>{data.facturacion.razonSocial}</b><br /><span
							>CUIT {data.facturacion.cuit} · {data.facturacion.tipo}</span
						></span
					>
				{:else}
					<span
						>Sin CUIT cargado. <a href={resolve('/panel/configuracion')}>Cargalo en Configuración</a
						></span
					>
				{/if}
			</div>
		</div>
	</div>

	<div class="tarjeta facturas" use:anim={{ t: 'rise', delay: 150 }}>
		<div class="f-cab"><span>Tus facturas de Tu Local Web</span></div>
		{#each data.facturas as f (f.numero)}
			<div class="factura">
				<span class="f-ico"><Icon icon={I.phReceipt} size={16} /></span>
				<span class="f-txt"
					><span class="f-n">{f.detalle}</span><span class="f-d"
						>{f.fecha ? fecha(f.fecha) : f.numero}</span
					></span
				>
				<span class="pagada">Pagada</span>
				<b>{pesos(f.total)}</b>
			</div>
		{:else}
			<div class="vacio">
				{data.enPrueba
					? `Estás en la prueba: si elegís un plan, el primer cobro es el ${fecha(data.finPrueba)}.`
					: 'Todavía no hay facturas.'}
			</div>
		{/each}
	</div>

	<div class="frenar" use:anim={{ t: 'rise', delay: 200 }}>
		<Icon icon={I.phPauseCircle} size={26} style="color:var(--mu)" />
		<div class="fr-txt">
			<div class="fr-t">{data.cancelado ? 'Tu plan está cancelado' : '¿Necesitás frenar?'}</div>
			<div class="fr-d">
				{data.cancelado
					? `Seguís con ${actual.nombre} hasta el ${fecha(data.cancelado.hasta)}. Si cambiás de idea, seguís con un toque.`
					: 'Podés pasar al plan Gratis sin perder lo que ya pagaste. Tus datos quedan guardados.'}
			</div>
		</div>
		{#if data.cancelado}
			<button type="button" class="fr-a" onclick={reactivar}>Seguir con {actual.nombre}</button>
		{:else}
			{#if data.actual !== 'gratis'}<button type="button" class="fr-a" onclick={aGratis}
					>Pasar a Gratis</button
				>
				<button type="button" class="fr-c" onclick={abrirCancelar}>Cancelar plan</button>{/if}
		{/if}
	</div>
</div>

<Modal bind:open={cancelando} titulo="Cancelar el plan" ancho={520} padding={24}>
	<div class="pasos">
		{#each [0, 1, 2] as i (i)}<span class:on={i <= paso}></span>{/each}
	</div>
	{#if paso === 0}
		<div class="m-cab">
			<img src="/hornero/preocupado.webp" alt="" width="64" height="64" />
			<div>
				<b>¿Por qué te vas?</b>
				<div class="m-s">Nos ayuda a mejorar. Es opcional.</div>
			</div>
		</div>
		<div class="motivos" role="radiogroup" aria-label="Motivo">
			{#each MOTIVOS as m, i (m)}
				<button type="button" role="radio" aria-checked={motivo === i} onclick={() => (motivo = i)}
					><span class="radio"></span>{m}</button
				>
			{/each}
		</div>
		<div class="m-acc">
			<button type="button" class="b-sec" onclick={() => (cancelando = false)}>Me quedo</button>
			<button type="button" class="b-osc" onclick={() => (paso = 1)}>Continuar</button>
		</div>
	{:else if paso === 1}
		<b class="m-t">{oferta[0]}</b>
		<div class="m-d">{oferta[1]}</div>
		<div class="oferta">
			<Icon icon={oferta[2]} size={28} style="color:#b0441d" />
			<div>
				<b>{oferta[3]}</b>
				<div class="m-s">{oferta[4]}</div>
			</div>
		</div>
		<div class="m-acc envolver">
			<button type="button" class="b-pri" onclick={aceptarOferta}>{oferta[5]}</button>
			<button type="button" class="b-sec" onclick={() => (paso = 2)}>Cancelar igual</button>
		</div>
	{:else}
		<b class="m-t">Confirmá la cancelación</b>
		<div class="consecuencias">
			{#each CONSECUENCIAS as [icono, t, c] (t)}
				<div><Icon icon={icono} size={17} style="color:{c};flex:none;margin-top:1px" />{t}</div>
			{/each}
		</div>
		<div class="m-acc">
			<button type="button" class="b-sec" onclick={() => (cancelando = false)}>Volver</button>
			<button type="button" class="b-roja" onclick={confirmar}>Cancelar mi plan</button>
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
		align-items: end;
		gap: 12px;
		flex-wrap: wrap;
	}
	.c-txt {
		flex: 1;
		min-width: 220px;
	}
	h1 {
		margin: 0;
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
	.c-s b {
		color: var(--tx);
	}
	.ciclos {
		display: flex;
		padding: 4px;
		border-radius: 999px;
		background: var(--chip);
	}
	.ciclos button {
		border: 0;
		font:
			600 13px 'Instrument Sans',
			sans-serif;
		padding: 8px 14px;
		border-radius: 999px;
		background: transparent;
		color: var(--tx);
		white-space: nowrap;
	}
	.ciclos button[aria-pressed='true'] {
		background: var(--sf);
		box-shadow: var(--sh);
	}
	.ciclos span {
		margin-left: 6px;
		font-size: 11px;
		font-weight: 700;
		color: #237550;
	}
	/* La compra: los planes y, al costado (desde 1280 px), el resumen fijo; si no, abajo. */
	.compra {
		display: grid;
		grid-template-columns: minmax(0, 1fr) 360px;
		gap: 16px;
		align-items: start;
	}
	@media (max-width: 1279px) {
		.compra {
			grid-template-columns: minmax(0, 1fr);
		}
	}
	/* Cuatro planes: 2 × 2 o 1, nunca 3 + 1. */
	.planes {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 14px;
		padding-top: 10px;
	}
	@media (max-width: 599px) {
		.planes {
			grid-template-columns: 1fr;
		}
	}
	.plan {
		position: relative;
		padding: 20px;
		border-radius: 24px;
		background: var(--sf);
		color: var(--tx);
		box-shadow: var(--shd);
		border: 2px solid transparent;
		display: flex;
		flex-direction: column;
		gap: 12px;
		font: inherit;
		text-align: left;
		cursor: pointer;
		transition:
			border-color 0.15s,
			transform 0.15s,
			box-shadow 0.15s;
	}
	.plan:hover {
		transform: translateY(-2px);
	}
	.plan:focus-visible {
		outline: 3px solid rgba(210, 88, 47, 0.35);
		outline-offset: 2px;
	}
	.plan.sel {
		border-color: #d2582f;
		box-shadow:
			0 0 0 4px rgba(210, 88, 47, 0.14),
			var(--shd);
	}
	.plan.bloqueado {
		opacity: 0.5;
	}
	.p-cab {
		display: flex;
		justify-content: space-between;
		gap: 10px;
	}
	.p-cab > span:first-child {
		display: flex;
		flex-direction: column;
	}
	.marca {
		width: 24px;
		height: 24px;
		flex: none;
		border-radius: 50%;
		border: 2px solid var(--ln2);
		display: grid;
		place-items: center;
		color: #fff;
	}
	.oscuro .marca {
		border-color: rgba(251, 247, 242, 0.35);
	}
	.sel .marca {
		border-color: #d2582f;
		background: #d2582f;
	}
	.tuyo.popular {
		background: #2a140c;
		color: #f0a58a;
		box-shadow: 0 0 0 1px rgba(240, 165, 138, 0.4);
	}
	.oscuro .tuyo.popular {
		background: #d2582f;
		color: #fff;
	}
	.p-incluye span {
		display: flex;
		gap: 8px;
		font-size: 13.5px;
		line-height: 1.4;
	}
	.p-btn {
		text-align: center;
	}
	.plan:not(.sel) .p-btn:not(.actual) {
		background: var(--chip);
		color: var(--tx);
	}
	.oscuro:not(.sel) .p-btn:not(.actual) {
		background: rgba(251, 247, 242, 0.12);
		color: #fbf7f2;
	}
	.sel .p-btn:not(.actual) {
		background: #d2582f;
		color: #fff;
	}
	.programado {
		display: flex;
		align-items: center;
		gap: 12px;
		flex-wrap: wrap;
		padding: 14px 16px;
		border-radius: 18px;
		background: linear-gradient(135deg, rgba(210, 88, 47, 0.12), rgba(210, 88, 47, 0.04));
		border: 1px solid rgba(210, 88, 47, 0.25);
	}
	.beneficio {
		background: linear-gradient(135deg, rgba(31, 138, 91, 0.12), rgba(31, 138, 91, 0.04));
		border-color: rgba(31, 138, 91, 0.25);
	}
	.programado > div {
		flex: 1;
		min-width: 200px;
		display: flex;
		flex-direction: column;
		font-size: 14px;
	}
	.programado span {
		color: var(--mu);
		font-size: 13px;
	}
	.programado button {
		min-height: 44px;
		padding: 0 16px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: var(--sf);
		color: var(--tx);
		font:
			600 13.5px 'Instrument Sans',
			sans-serif;
	}
	/* El resumen: el «pedido» del handoff (Mostrador v4, cajón del pedido). */
	.resumen {
		position: sticky;
		top: 16px;
		padding: 18px;
		border-radius: 24px;
		background: var(--sf);
		box-shadow: var(--shd);
		display: flex;
		flex-direction: column;
		gap: 14px;
		scroll-margin-top: 16px;
	}
	@media (max-width: 1279px) {
		.resumen {
			position: static;
		}
	}
	.r-cab {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 16px;
		font-weight: 600;
	}
	.r-de-a {
		display: flex;
		align-items: center;
		gap: 8px;
		flex-wrap: wrap;
	}
	.chip-plan {
		padding: 5px 11px;
		border-radius: 999px;
		background: var(--chip);
		font-size: 13px;
		font-weight: 600;
	}
	.chip-plan.nuevo {
		background: rgba(210, 88, 47, 0.12);
		color: #b0441d;
	}
	.r-bloque {
		display: flex;
		flex-direction: column;
		gap: 8px;
		padding-top: 14px;
		border-top: 1px solid var(--ln);
	}
	.r-t {
		font-size: 12px;
		font-weight: 700;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		color: var(--fa);
	}
	.r-linea {
		display: flex;
		justify-content: space-between;
		gap: 12px;
		font-size: 14px;
	}
	.r-linea span:last-child {
		white-space: nowrap;
		font-variant-numeric: tabular-nums;
	}
	.r-linea.credito {
		color: #237550;
	}
	.r-nada {
		margin: 0;
		font-size: 13.5px;
		line-height: 1.45;
		color: var(--mu);
		text-wrap: pretty;
	}
	.r-total {
		display: flex;
		justify-content: space-between;
		align-items: baseline;
		gap: 12px;
		margin-top: 4px;
		padding-top: 10px;
		border-top: 1px dashed var(--ln2);
		font-size: 15px;
		font-weight: 600;
	}
	.r-total b {
		font-size: 30px;
		letter-spacing: -0.04em;
		font-variant-numeric: tabular-nums;
	}
	.r-iva {
		margin-top: -4px;
		text-align: right;
		font-size: 12px;
		color: var(--fa);
	}
	.r-dato {
		display: flex;
		gap: 10px;
		font-size: 13.5px;
		line-height: 1.45;
		color: var(--mu);
	}
	.r-dato :global(svg) {
		flex: none;
		margin-top: 1px;
		color: #b0441d;
	}
	.r-dato b {
		color: var(--tx);
	}
	.r-eq {
		color: var(--fa);
	}
	.r-incluye {
		margin: 0;
		padding: 0;
		list-style: none;
		display: flex;
		flex-direction: column;
		gap: 6px;
		font-size: 13.5px;
	}
	.r-incluye li {
		display: flex;
		gap: 8px;
	}
	.r-incluye :global(svg) {
		flex: none;
		margin-top: 3px;
		color: #237550;
	}
	.r-como {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 12px;
		border-radius: 16px;
		background: var(--chip);
		font-size: 13px;
		line-height: 1.45;
		color: var(--mu);
	}
	.r-ico {
		width: 30px;
		height: 30px;
		flex: none;
		border-radius: 9px;
		display: grid;
		place-items: center;
		background: rgba(47, 143, 99, 0.14);
		color: #237550;
	}
	.r-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 8px;
		min-height: 50px;
		border: 0;
		border-radius: 999px;
		background: linear-gradient(180deg, #c24e25, #b0441d);
		color: #fff;
		font:
			600 15px 'Instrument Sans',
			sans-serif;
		cursor: pointer;
	}
	.r-btn.roja {
		background: var(--tx);
		color: var(--sf);
	}
	.r-btn:disabled {
		opacity: 0.6;
	}
	.r-volver {
		margin-top: -6px;
		min-height: 44px;
		border: 0;
		background: transparent;
		color: var(--mu);
		font:
			600 13.5px 'Instrument Sans',
			sans-serif;
		cursor: pointer;
	}
	.r-vacio {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 6px;
		padding: 10px 6px 4px;
		text-align: center;
	}
	.r-vacio img {
		object-fit: contain;
	}
	.r-vacio b {
		font-size: 15px;
	}
	.r-vacio span {
		font-size: 13.5px;
		line-height: 1.45;
		color: var(--mu);
		text-wrap: pretty;
	}
	.r-actual {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 2px;
		padding: 12px;
		border-radius: 16px;
		background: var(--chip);
		font-size: 13px;
		color: var(--mu);
	}
	.r-actual b {
		font-size: 16px;
		color: var(--tx);
	}
	.r-pie {
		display: flex;
		gap: 6px;
		font-size: 12px;
		line-height: 1.4;
		color: var(--fa);
	}
	.r-pie :global(svg) {
		flex: none;
		margin-top: 1px;
	}
	/* En el celular, el total y el botón siempre a mano. */
	.barra-movil {
		display: none;
	}
	@media (max-width: 759px) {
		.barra-movil {
			position: sticky;
			/* Arriba de la barra de abajo del panel (76 px + el área segura) y del botón de Hornero. */
			bottom: calc(96px + env(safe-area-inset-bottom));
			z-index: 5;
			display: flex;
			align-items: center;
			justify-content: space-between;
			gap: 12px;
			padding: 10px 10px 10px 18px;
			border-radius: 999px;
			background: #2a140c;
			color: #fbf7f2;
			box-shadow: 0 12px 30px -10px rgba(42, 20, 12, 0.6);
		}
		.barra-movil span {
			display: flex;
			flex-direction: column;
			line-height: 1.15;
		}
		.barra-movil small {
			font-size: 11.5px;
			opacity: 0.75;
		}
		.barra-movil b {
			font-size: 19px;
			letter-spacing: -0.03em;
		}
		.barra-movil button {
			min-height: 44px;
			padding: 0 22px;
			border: 0;
			border-radius: 999px;
			background: #d2582f;
			color: #fff;
			font:
				600 14.5px 'Instrument Sans',
				sans-serif;
		}
		.barra-movil button.roja {
			background: #fbf7f2;
			color: #2a140c;
		}
	}
	.plan.oscuro {
		background: #2a140c;
		color: #fbf7f2;
	}
	.plan.cur {
		border-color: #d2582f;
	}
	.tuyo {
		position: absolute;
		top: -11px;
		left: 18px;
		padding: 4px 10px;
		border-radius: 999px;
		background: #b0441d;
		color: #fff;
		font-size: 11.5px;
		font-weight: 700;
	}
	.p-n {
		font-size: 18px;
		font-weight: 600;
	}
	.p-d {
		font-size: 13.5px;
		opacity: 0.75;
		margin-top: 2px;
	}
	.p-precio {
		display: flex;
		align-items: baseline;
		gap: 6px;
	}
	.p-precio b {
		font-size: 32px;
		letter-spacing: -0.04em;
	}
	.p-precio span,
	.p-anual {
		font-size: 13.5px;
		opacity: 0.75;
	}
	.p-anual {
		font-size: 12.5px;
		margin-top: -8px;
		opacity: 0.8;
	}
	.p-n,
	.p-d,
	.p-precio,
	.p-anual,
	.p-incluye {
		display: block;
	}
	.p-precio {
		display: flex;
	}
	.p-incluye {
		display: flex;
	}
	.p-incluye {
		display: flex;
		flex-direction: column;
		gap: 7px;
	}
	.p-incluye :global(svg) {
		flex: none;
		margin-top: 2px;
		color: #b0441d;
	}
	.oscuro .p-incluye :global(svg) {
		color: #f0a58a;
	}
	.p-btn {
		margin-top: auto;
		padding: 12px;
		border-radius: 999px;
		border: 0;
		background: var(--tx);
		color: var(--sf);
		font:
			600 14px 'Instrument Sans',
			sans-serif;
	}
	.oscuro .p-btn {
		background: #b0441d;
		color: #fff;
	}
	.p-btn.actual {
		background: transparent;
		color: var(--mu);
		border: 1.5px solid var(--ln2);
	}
	.oscuro .p-btn.actual {
		color: #fbf7f2;
		border-color: rgba(251, 247, 242, 0.3);
	}
	.dos {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 320px), 1fr));
		gap: 12px;
		align-items: start;
	}
	.tarjeta {
		padding: 18px;
		border-radius: 24px;
		background: var(--sf);
		box-shadow: var(--shd);
	}
	.uso,
	.medio {
		display: flex;
		flex-direction: column;
		gap: 14px;
	}
	.u-cab {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 10px;
		font-size: 16px;
		font-weight: 600;
	}
	.u-cab button {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 7px 12px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: transparent;
		color: var(--tx);
		font:
			600 12.5px 'Instrument Sans',
			sans-serif;
		white-space: nowrap;
	}
	.u-f {
		display: flex;
		justify-content: space-between;
		gap: 8px;
		font-size: 13.5px;
	}
	.u-f span:last-child {
		color: var(--mu);
		white-space: nowrap;
	}
	.u-f b {
		color: var(--tx);
	}
	.u-b {
		margin-top: 6px;
		height: 7px;
		border-radius: 4px;
		background: var(--chip);
		overflow: hidden;
	}
	.u-b div {
		height: 100%;
		border-radius: 4px;
	}
	.tarjeta-credito {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 14px;
		border-radius: 16px;
		background: linear-gradient(135deg, #2a140c, #5a3424);
		color: #fff;
	}
	.visa {
		display: grid;
		place-items: center;
		background: #fff;
		border-radius: 5px;
		padding: 3px 5px;
	}
	.t-n {
		font-size: 15px;
		font-weight: 600;
		letter-spacing: 0.06em;
	}
	.t-v {
		font-size: 12.5px;
		opacity: 0.75;
	}
	.sin-dato {
		margin: 0;
		font-size: 13px;
		line-height: 1.45;
		color: var(--fa);
	}
	.fact a {
		color: var(--terracota);
		font-weight: 600;
	}
	.fact {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 10px;
		font-size: 13.5px;
	}
	.fact > span:first-child,
	.fact span span {
		color: var(--mu);
	}
	.fact > span:last-child {
		text-align: right;
	}
	.facturas {
		padding: 0;
		overflow: hidden;
	}
	.f-cab {
		padding: 14px 18px;
		border-bottom: 1px solid var(--ln);
		font-size: 16px;
		font-weight: 600;
	}
	.factura {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 12px 18px;
		border-bottom: 1px solid var(--ln);
		flex-wrap: wrap;
	}
	.f-ico {
		width: 34px;
		height: 34px;
		border-radius: 10px;
		background: var(--chip);
		display: grid;
		place-items: center;
	}
	.f-txt {
		flex: 1;
		min-width: 160px;
	}
	.f-n {
		display: block;
		font-size: 14px;
		font-weight: 600;
	}
	.f-d {
		display: block;
		font-size: 12.5px;
		color: var(--mu);
	}
	.pagada {
		font-size: 12px;
		font-weight: 700;
		padding: 3px 9px;
		border-radius: 999px;
		background: rgba(47, 143, 99, 0.12);
		color: #237550;
	}
	.vacio {
		padding: 18px;
		font-size: 14px;
		color: var(--mu);
	}
	.frenar {
		display: flex;
		align-items: center;
		gap: 14px;
		flex-wrap: wrap;
		padding: 16px 18px;
		border-radius: 22px;
		border: 1px dashed var(--ln2);
	}
	.fr-txt {
		flex: 1;
		min-width: 220px;
	}
	.fr-t {
		font-size: 15px;
		font-weight: 600;
	}
	.fr-d {
		font-size: 13.5px;
		color: var(--mu);
	}
	.fr-a {
		padding: 10px 16px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: var(--sf);
		color: var(--tx);
		font:
			600 13.5px 'Instrument Sans',
			sans-serif;
	}
	.fr-c {
		padding: 10px 16px;
		border-radius: 999px;
		border: 0;
		background: transparent;
		color: #b83a2e;
		font:
			600 13.5px 'Instrument Sans',
			sans-serif;
	}
	.pasos {
		display: flex;
		gap: 6px;
		margin-bottom: 16px;
	}
	.pasos span {
		flex: 1;
		height: 4px;
		border-radius: 2px;
		background: var(--ln2);
	}
	.pasos span.on {
		background: #d2582f;
	}
	.m-cab {
		display: flex;
		gap: 12px;
		align-items: center;
	}
	.m-cab img {
		object-fit: contain;
	}
	.m-cab b,
	.m-t {
		font-size: 19px;
	}
	.m-s {
		font-size: 13.5px;
		color: var(--mu);
	}
	.motivos {
		margin-top: 14px;
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.motivos button {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 12px 14px;
		border-radius: 14px;
		border: 1.5px solid var(--ln2);
		background: transparent;
		color: var(--tx);
		font:
			500 14px 'Instrument Sans',
			sans-serif;
		text-align: left;
	}
	.radio {
		width: 18px;
		height: 18px;
		border-radius: 50%;
		border: 2px solid var(--ln2);
		display: grid;
		place-items: center;
		flex: none;
	}
	.motivos button[aria-checked='true'] {
		border-color: #d2582f;
		background: rgba(210, 88, 47, 0.06);
	}
	.motivos button[aria-checked='true'] .radio {
		border-color: #d2582f;
	}
	.motivos button[aria-checked='true'] .radio::after {
		content: '';
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: #d2582f;
	}
	.m-d {
		margin-top: 6px;
		font-size: 14px;
		line-height: 1.5;
		color: var(--mu);
	}
	.oferta {
		margin-top: 14px;
		padding: 16px;
		border-radius: 18px;
		background: linear-gradient(135deg, rgba(210, 88, 47, 0.12), rgba(210, 88, 47, 0.04));
		border: 1px solid rgba(210, 88, 47, 0.25);
		display: flex;
		align-items: center;
		gap: 12px;
	}
	.oferta b {
		font-size: 15px;
	}
	.consecuencias {
		margin-top: 10px;
		display: flex;
		flex-direction: column;
		gap: 8px;
		font-size: 14px;
	}
	.consecuencias div {
		display: flex;
		gap: 10px;
		align-items: flex-start;
	}
	.m-acc {
		margin-top: 16px;
		display: flex;
		gap: 8px;
	}
	.m-acc.envolver {
		flex-wrap: wrap;
	}
	.m-acc button {
		flex: 1;
		padding: 13px;
		border-radius: 999px;
		font:
			600 14px 'Instrument Sans',
			sans-serif;
	}
	.envolver .b-pri {
		flex: 1 1 180px;
	}
	.envolver .b-sec {
		flex: 1 1 140px;
	}
	.b-sec {
		border: 1px solid var(--ln2);
		background: transparent;
		color: var(--tx);
	}
	.b-osc {
		border: 0;
		background: var(--tx);
		color: var(--sf);
	}
	.b-pri {
		border: 0;
		background: linear-gradient(180deg, #c24e25, #b0441d);
		color: #fff;
	}
	.b-roja {
		border: 0;
		background: #b83a2e;
		color: #fff;
	}
</style>
