<script lang="ts">
	import { haceCorto, plural } from '#lib/panel/formato.ts';
	import { deslizable } from '#lib/panel/deslizable.ts';
	import { invalidateAll, replaceState } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import Icon from '#lib/components/Icon.svelte';
	import Interruptor from '#lib/components/Interruptor.svelte';
	import Modal from '#lib/components/Modal.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { marca } from '#lib/marcas.ts';
	import { anim } from '#lib/panel/anim.ts';
	import {
		CATEGORIAS,
		DE_MEDICION,
		IAS,
		OTRA_IA,
		INTEGRACIONES,
		PERMISOS_IA,
		type AjusteIntegracion,
		type Integracion
	} from '#lib/panel/integraciones.ts';
	import { getUi } from '#lib/ui.svelte.ts';
	import { validarGa4, validarPixel } from '#lib/tienda/medicion.ts';
	import Arca from './Arca.svelte';
	import Correo from './Correo.svelte';

	/**
	 * Integraciones (design/PanelOperacion.dc.html, sInt). "Configurar" no está diseñado: se arma con
	 * el modal y los interruptores del handoff. Mercado Pago es real (OAuth de la tienda); las demás
	 * llegan una por PR y en la demo se simulan.
	 */
	let { data } = $props();
	const ui = getUi();

	let cat = $state<(typeof CATEGORIAS)[number][0]>('todas');
	const lista = $derived(INTEGRACIONES.filter((i) => cat === 'todas' || i.categoria === cat));
	const conectadas = $derived(Object.values(data.estado.conectadas).filter(Boolean).length);
	let abierta = $state<Integracion | null>(null);
	let modal = $state(false);

	// La vuelta del OAuth de Mercado Pago deja ?mercadopago=conectada|error.
	let avisado = false;
	$effect(() => {
		const mp = page.url.searchParams.get('mercadopago');
		const ml = page.url.searchParams.get('mercadolibre');
		if ((!mp && !ml) || avisado) return;
		avisado = true;
		const nombre = mp ? 'Mercado Pago' : 'Mercado Libre';
		ui.toast(
			(mp ?? ml) === 'conectada'
				? {
						k: 'ok',
						title: `${nombre} conectado`,
						desc: mp ? 'Ya podés cobrar con tu cuenta' : 'Publicá productos desde su ficha'
					}
				: { k: 'err', title: `No se conectó ${nombre}`, desc: 'Probá de nuevo en un rato' }
		);
		replaceState(resolve('/panel/integraciones'), {});
	});

	async function ajustar(a: AjusteIntegracion) {
		const r = await fetch('/panel/integraciones/ajustar', {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify(a)
		});
		if (!r.ok) {
			const m = ((await r.json().catch(() => null)) as { message?: string } | null)?.message;
			ui.toast({ k: 'warn', title: 'Todavía no', desc: m ?? 'Probá de nuevo.' });
			return false;
		}
		await invalidateAll();
		return true;
	}

	const RUTA = {
		mp: '/panel/integraciones/mercadopago',
		ml: '/panel/integraciones/mercadolibre'
	} as const;
	async function conectarMP(id: 'mp' | 'ml' = 'mp') {
		const r = await fetch(RUTA[id], { method: 'POST' });
		if (!r.ok) {
			const m = ((await r.json().catch(() => null)) as { message?: string } | null)?.message;
			return ui.toast({
				k: 'warn',
				title: id === 'mp' ? 'Mercado Pago' : 'Mercado Libre',
				desc: m ?? 'Probá de nuevo.'
			});
		}
		location.href = ((await r.json()) as { url: string }).url;
	}

	async function desconectarMP(id: 'mp' | 'ml' = 'mp') {
		const r = await fetch(RUTA[id], { method: 'DELETE' });
		if (!r.ok) return ui.toast({ k: 'err', title: 'No se pudo', desc: 'Probá de nuevo.' });
		modal = false;
		await invalidateAll();
		ui.toast(
			id === 'mp'
				? {
						k: 'info',
						title: 'Mercado Pago desconectado',
						desc: 'La tienda deja de cobrar con tarjeta hasta que lo conectes'
					}
				: {
						k: 'info',
						title: 'Mercado Libre desconectado',
						desc: 'Las publicaciones quedan, pero sin sincronizar stock ni precio'
					}
		);
	}

	// Telegram: el link al bot con el código; Telegram se abre aparte y el chat se vincula al tocar Iniciar.
	let esperandoTg = $state(false);
	async function vincularTelegram() {
		const r = await fetch('/panel/integraciones/telegram', { method: 'POST' });
		if (!r.ok) {
			const m = ((await r.json().catch(() => null)) as { message?: string } | null)?.message;
			return ui.toast({ k: 'info', title: 'Telegram', desc: m ?? 'Probá de nuevo.' });
		}
		window.open(((await r.json()) as { url: string }).url, '_blank', 'noopener');
		esperandoTg = true;
	}
	async function desvincularTelegram() {
		const r = await fetch('/panel/integraciones/telegram', { method: 'DELETE' });
		if (!r.ok) return ui.toast({ k: 'err', title: 'No se pudo', desc: 'Probá de nuevo.' });
		modal = false;
		await invalidateAll();
		ui.toast({ k: 'info', title: 'Telegram desvinculado', desc: 'Tincho ya no te escribe ahí' });
	}

	async function ajustarML(recargo: number, pausarSinStock: boolean) {
		const r = await fetch(RUTA.ml, {
			method: 'PUT',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ recargo, pausarSinStock })
		});
		if (!r.ok) return ui.toast({ k: 'err', title: 'No se pudo', desc: 'Probá de nuevo.' });
		await invalidateAll();
	}

	// Google Analytics, Google Shopping y Meta Ads: con las cuentas de la tienda (IDs y feed).
	type Medicion = typeof data.medicion;
	const esMedicion = (id: string) => (DE_MEDICION as readonly string[]).includes(id);
	let ga4 = $state('');
	let pixel = $state('');
	let errorId = $state<string | null>(null);

	function abrir(i: Integracion) {
		abierta = i;
		ga4 = data.medicion.ga4 ?? '';
		pixel = data.medicion.pixel ?? '';
		errorId = null;
		modal = true;
	}

	async function guardarMedicion(m: Medicion) {
		const r = await fetch('/panel/integraciones/medicion', {
			method: 'PUT',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify(m)
		});
		if (!r.ok) {
			const e = ((await r.json().catch(() => null)) as { message?: string } | null)?.message;
			errorId = e ?? 'No se pudo guardar. Probá de nuevo.';
			return false;
		}
		await invalidateAll();
		return true;
	}

	async function guardarId(id: 'ga' | 'meta') {
		const v = id === 'ga' ? validarGa4(ga4) : validarPixel(pixel);
		if (!v.ok) return (errorId = v.error);
		errorId = null;
		const antes = data.medicion;
		const ahora = id === 'ga' ? { ...antes, ga4: v.valor } : { ...antes, pixel: v.valor };
		if ((id === 'ga' ? antes.ga4 : antes.pixel) === v.valor) return (modal = false);
		if (!(await guardarMedicion(ahora))) return;
		modal = false;
		const deshacer = async () => {
			if (!(await guardarMedicion(antes)))
				ui.toast({ k: 'err', title: 'No se pudo deshacer', desc: errorId ?? 'Probá de nuevo.' });
		};
		if (!v.valor)
			return ui.undoable(
				id === 'ga' ? 'Google Analytics desconectado' : 'Píxel de Meta borrado',
				'Tu tienda deja de medir con ese ID',
				deshacer
			);
		ui.toast({
			k: 'ok',
			title: id === 'ga' ? 'Google Analytics conectado' : 'Píxel de Meta guardado',
			desc: 'Mide a quienes acepten las cookies de tu tienda'
		});
	}

	// API de conversiones de Meta (§47): el token se manda una vez y no vuelve (Go lo guarda cifrado).
	let capiToken = $state('');
	let capiCodigo = $derived(data.conversiones?.codigoPrueba ?? '');
	let capiError = $state<string | null>(null);
	let capiGuardando = $state(false);

	const cuando = (iso: string) => {
		const h = haceCorto(iso);
		return h === 'ayer' ? 'ayer' : `hace ${h}`;
	};

	async function pedirConversiones(metodo: 'PUT' | 'DELETE', cuerpo?: object) {
		capiGuardando = true;
		try {
			const r = await fetch('/panel/integraciones/medicion/conversiones', {
				method: metodo,
				headers: cuerpo ? { 'content-type': 'application/json' } : undefined,
				body: cuerpo ? JSON.stringify(cuerpo) : undefined
			});
			if (!r.ok) {
				capiError =
					((await r.json().catch(() => null)) as { message?: string } | null)?.message ??
					'No se pudo guardar. Probá de nuevo.';
				return false;
			}
			capiError = null;
			await invalidateAll();
			return true;
		} finally {
			capiGuardando = false;
		}
	}

	async function guardarConversiones() {
		const yaEstaba = !!data.conversiones?.conectada;
		if (!yaEstaba && !capiToken.trim()) return (capiError = 'Pegá el token de acceso.');
		if (!(await pedirConversiones('PUT', { token: capiToken, codigoPrueba: capiCodigo }))) return;
		capiToken = '';
		ui.toast({
			k: 'ok',
			title: yaEstaba ? 'API de conversiones actualizada' : 'API de conversiones conectada',
			desc: 'Las compras llegan a Meta aunque el navegador bloquee el píxel'
		});
	}

	async function desconectarConversiones() {
		if (await pedirConversiones('DELETE'))
			ui.toast({ k: 'ok', title: 'API de conversiones desconectada', desc: 'Queda sólo el píxel' });
	}

	async function yaLoCargue(id: 'gm' | 'meta') {
		const m = data.medicion;
		if (!(await guardarMedicion({ ...m, feeds: { ...m.feeds, [id]: true } }))) return;
		if (id === 'gm') modal = false;
		ui.undoable(
			id === 'gm' ? 'Google Shopping conectado' : 'Catálogo de Meta conectado',
			id === 'gm' ? 'Google revisa los productos en unos días' : 'Meta lee el feed una vez por día',
			async () => {
				if (!(await guardarMedicion(m)))
					ui.toast({ k: 'err', title: 'No se pudo deshacer', desc: errorId ?? 'Probá de nuevo.' });
			}
		);
	}

	async function copiarFeed() {
		try {
			await navigator.clipboard.writeText(data.feed);
			ui.toast({ k: 'ok', title: 'Link copiado', desc: 'Pegalo como URL del feed' });
		} catch {
			ui.toast({ k: 'warn', title: 'No se pudo copiar', desc: 'Seleccioná el link y copialo' });
		}
	}

	async function desconectarMedicion(i: Integracion) {
		const antes = data.medicion;
		const ahora: Medicion =
			i.id === 'ga'
				? { ...antes, ga4: null }
				: i.id === 'gm'
					? { ...antes, feeds: { ...antes.feeds, gm: false } }
					: { ...antes, pixel: null, feeds: { ...antes.feeds, meta: false } };
		if (!(await guardarMedicion(ahora))) return;
		modal = false;
		ui.undoable(`${i.nombre} desconectado`, 'Tu tienda deja de mandarle datos', async () => {
			await guardarMedicion(antes);
		});
	}
	// ARCA tiene su asistente (datos fiscales, pedido de certificado, certificado y prueba).
	// Facturación manda acá con ?arca para abrirlo.
	let arcaAbierto = $state(false);
	$effect(() => {
		if (page.url.searchParams.get('arca') === null) return;
		arcaAbierto = true;
		replaceState(resolve('/panel/integraciones'), {});
	});
	const venceArca = $derived(
		data.arca.certificadoVence
			? Math.ceil((Date.parse(data.arca.certificadoVence) - Date.now()) / 86_400_000)
			: null
	);

	let correoAbierto = $state(false);

	async function boton(i: Integracion) {
		if (i.id === 'arca') return void (arcaAbierto = true);
		if (i.id === 'correo') return void (correoAbierto = true);
		const on = data.estado.conectadas[i.id];
		if (on || esMedicion(i.id)) return abrir(i);
		if (i.id === 'mp' || i.id === 'ml') return conectarMP(i.id);
		if (i.id === 'tg') return vincularTelegram();
		if (i.pronto)
			return ui.toast({
				k: 'info',
				title: i.nombre,
				desc: 'Todavía no se puede conectar. Llega pronto.'
			});
		if (await ajustar({ tipo: 'conexion', id: i.id, valor: true }))
			ui.toast({
				k: 'ok',
				title: `${i.nombre} conectada`,
				desc: 'Sincronizando datos, tarda unos segundos'
			});
	}

	async function desconectar(i: Integracion) {
		if (esMedicion(i.id)) return desconectarMedicion(i);
		if (i.id === 'mp' || i.id === 'ml') return desconectarMP(i.id);
		if (i.id === 'tg') return desvincularTelegram();
		if (!(await ajustar({ tipo: 'conexion', id: i.id, valor: false }))) return;
		modal = false;
		ui.toast({ k: 'info', title: `${i.nombre} desconectada`, desc: 'Podés volver a conectarla' });
	}

	// Manejá Tu Local Web desde tu IA: el servidor MCP con OAuth. Cada IA se conecta desde su lado con
	// la dirección del servidor; el permiso se da en /oauth/autorizar y acá se listan y desconectan.
	let iaAbierta = $state<(typeof IAS)[number] | null>(null);
	let modalIa = $state(false);
	const conexionesDe = (id: string) => data.mcp.conexiones.filter((c) => c.ia === id);

	function abrirIa(p: (typeof IAS)[number]) {
		if (!data.mcp.disponible)
			return ui.toast({
				k: 'info',
				title: p.nombre,
				desc: 'Llega pronto.'
			});
		iaAbierta = p;
		modalIa = true;
	}

	async function copiarUrl() {
		try {
			await navigator.clipboard.writeText(data.mcp.url);
			ui.toast({ k: 'ok', title: 'Dirección copiada', desc: 'Pegala en tu IA' });
		} catch {
			ui.toast({
				k: 'warn',
				title: 'No se pudo copiar',
				desc: 'Seleccioná la dirección y copiala'
			});
		}
	}

	async function desconectarIa(id: string, nombre: string) {
		const r = await fetch('/panel/integraciones/ia', {
			method: 'DELETE',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ id })
		});
		if (!r.ok) return ui.toast({ k: 'err', title: 'No se pudo', desc: 'Probá de nuevo.' });
		await invalidateAll();
		ui.toast({
			k: 'info',
			title: `${nombre} desconectado`,
			desc: 'Ya no puede ver ni tocar tu negocio'
		});
	}

	const PASOS: Record<(typeof IAS)[number]['id'], string[]> = {
		claude: [
			'En Claude, abrí Configuración\u00a0›\u00a0Conectores.',
			'Tocá Agregar conector personalizado, ponele Tu Local Web de nombre y pegá la\u00a0dirección.',
			'Tocá Conectar: Claude te trae acá para que permitas el\u00a0acceso.'
		],
		chatgpt: [
			'En ChatGPT, abrí Configuración\u00a0› Aplicaciones y conectores\u00a0› Configuración avanzada y prendé el modo\u00a0desarrollador.',
			'Volvé a Aplicaciones y conectores, tocá Crear, ponele Tu Local Web de nombre, pegá la dirección y elegí\u00a0OAuth.',
			'Tocá Crear: ChatGPT te trae acá para que permitas el\u00a0acceso.'
		],
		mcp: [
			'En tu IA, buscá dónde se agregan servidores MCP remotos (a veces dice conectores o\u00a0integraciones).',
			'Pegá la dirección. Usa OAuth: la IA te trae acá para que permitas el\u00a0acceso.'
		]
	};

	const hora = (ms: number) =>
		new Date(ms).toLocaleTimeString('es-AR', {
			hour: '2-digit',
			minute: '2-digit',
			timeZone: 'America/Argentina/Buenos_Aires'
		});

	const fecha = (iso?: string) =>
		iso
			? new Date(iso).toLocaleDateString('es-AR', {
					day: 'numeric',
					month: 'long',
					timeZone: 'America/Argentina/Buenos_Aires'
				})
			: '';
</script>

<Seo
	title="Integraciones · Panel"
	description="Integraciones de tu tienda"
	path="/panel/integraciones"
	noindex
/>

<div class="pagina" data-mp-marketplace-page="checkout-pro">
	<div class="cab" use:anim={{ t: 'rise' }}>
		<div class="c-txt">
			<h1>Integraciones</h1>
			<div class="sub">{conectadas} conectadas · todas incluidas en tu&nbsp;plan</div>
		</div>
	</div>

	{#if esperandoTg && !data.estado.conectadas.tg}
		<div class="espera" role="status">
			<Icon icon={I.logosTelegram} size={20} />
			<span>Tocá <b>Iniciar</b> en Telegram y volvé acá.</span>
			<button type="button" onclick={() => invalidateAll()}>Ya lo hice</button>
		</div>
	{/if}

	<div
		class="cats"
		role="group"
		aria-label="Categoría"
		use:anim={{ t: 'rise', delay: 100 }}
		use:deslizable
	>
		{#each CATEGORIAS as [id, t] (id)}
			<button
				type="button"
				class:on={cat === id}
				aria-pressed={cat === id}
				onclick={() => (cat = id)}>{t}</button
			>
		{/each}
	</div>

	<div class="grilla" use:anim={{ t: 'rise', delay: 150 }}>
		{#each lista as i (i.id)}
			{@const on = data.estado.conectadas[i.id]}
			{@const m = marca(i.marca)}
			<div class="tarjeta" class:on>
				<div class="t-cab">
					<span class="logo" style:background={m.bg}>
						{#if m.icon}<Icon
								icon={m.icon}
								size={22}
								style="color:{m.col ?? 'inherit'}"
							/>{:else}<span class="sigla" style:color={m.col}>{m.sigla}</span>{/if}
					</span>
					<span class="t-txt"
						><span class="t-n">{i.nombre}</span><span class="t-d">{i.detalle}</span></span
					>
				</div>
				<div class="t-pie">
					{#if i.id === 'arca' && on && venceArca !== null && venceArca < 30}
						<span class="t-st vence"
							>{venceArca > 0
								? `Vence en ${plural(venceArca, 'día', 'días')}`
								: 'Certificado vencido'}</span
						>
					{:else}
						<span class="t-st" class:on
							>{on ? '● Conectada' : i.pronto ? 'Próximamente' : 'Disponible'}</span
						>
					{/if}
					<!-- Lo que todavía no se puede conectar no ofrece un «Conectar» que sólo avisa que no. -->
					<button
						type="button"
						class="t-btn"
						class:on
						disabled={!on && i.pronto}
						data-mp-marketplace-connect={i.id === 'mp' && !on ? 'oauth' : undefined}
						onclick={() => boton(i)}
					>
						{on ? 'Configurar' : i.pronto ? 'Pronto' : 'Conectar'}
					</button>
				</div>
			</div>
		{/each}
	</div>

	<!-- Cobrar, vender y enviar va primero: manejar la tienda desde la IA es un extra. -->
	<section class="ia" use:anim={{ t: 'rise', delay: 200 }}>
		<div class="ia-cab">
			<img src="/hornero/whatsapp.webp" alt="" width="52" height="52" />
			<div class="ia-txt">
				<div class="ia-t">Manejá Tu Local Web desde tu IA<span class="mcp">MCP</span></div>
				<div class="ia-d">
					Entrás con tu cuenta de la IA y listo. Preguntale por ventas, stock o pedidos, y pedile
					que ajuste un producto, cree un cupón o arme un link de&nbsp;pago.
				</div>
			</div>
		</div>
		<div class="provs">
			{#each IAS as p (p.id)}
				{@const on = data.estado.ia[p.id]}
				<button type="button" class="prov" class:on onclick={() => abrirIa(p)}>
					<span class="p-ico" style:color={p.col}><Icon icon={p.icon} size={19} /></span>
					<span class="p-txt">
						<span class="p-t">{p.nombre}</span>
						<span class="p-st">{on ? 'Conectado' : 'Conectar con tu cuenta'}</span>
					</span>
					<Icon icon={on ? I.phCheckCircleFill : I.phArrowRightBold} size={16} />
				</button>
			{/each}
		</div>
		<div class="perms">
			<!-- «Responder mensajes» no tiene ninguna herramienta detrás (no hay WhatsApp ni Instagram). -->
			{#each PERMISOS_IA.filter((g) => data.demo || g.id !== 'responder') as g (g.id)}
				{@const on = data.estado.permisosIa[g.id]}
				<button
					type="button"
					class="perm"
					role="switch"
					aria-checked={on}
					disabled={g.id === 'leer'}
					onclick={() => ajustar({ tipo: 'permisoIa', id: g.id, valor: !on })}
				>
					<Icon icon={g.icon} size={17} />
					<span class="g-txt"><span class="g-t">{g.t}</span><span class="g-d">{g.d}</span></span>
					<span class="tg" class:on><span></span></span>
				</button>
			{/each}
		</div>
		{#if data.mcp.conexiones.length}
			<div class="conexiones" aria-label="IAs conectadas">
				{#each data.mcp.conexiones as c (c.id)}
					{@const p = IAS.find((x) => x.id === c.ia) ?? OTRA_IA}
					<div class="cx">
						<span class="p-ico" style:color={p.col}><Icon icon={p.icon} size={17} /></span>
						<span class="g-txt"
							><span class="g-t">{c.nombre}</span><span class="g-d"
								>{c.acciones ? 'Lectura y acciones' : 'Sólo lectura'} · desde las {hora(
									c.desde
								)}{#if c.usada}&nbsp;· usada a las {hora(c.usada)}{/if}</span
							></span
						>
						<button
							type="button"
							class="t-btn on rojo"
							onclick={() => desconectarIa(c.id, c.nombre)}>Desconectar</button
						>
					</div>
				{/each}
			</div>
		{/if}
	</section>
</div>

<Modal
	bind:open={modal}
	titulo={abierta ? `Configurar ${abierta.nombre}` : 'Configurar'}
	ancho={480}
	padding={24}
>
	{#if abierta}
		{@const m = marca(abierta.marca)}
		{@const mp = abierta.id === 'mp'}
		{@const conectada = data.estado.conectadas[abierta.id]}
		<div class="m-cab">
			<span class="logo grande" style:background={m.bg}>
				{#if m.icon}<Icon icon={m.icon} size={26} style="color:{m.col ?? 'inherit'}" />{:else}<span
						class="sigla"
						style:color={m.col}>{m.sigla}</span
					>{/if}
			</span>
			<div>
				<div class="m-t">{abierta.nombre}</div>
				<div class="m-st" class:apagada={!conectada}>
					{#if conectada}● Conectada{#if mp && data.mp.desde}&nbsp;desde el {fecha(
								data.mp.desde
							)}{/if}{:else}Sin conectar{/if}
				</div>
			</div>
		</div>
		{#if mp && data.mp.reautorizar}
			<div class="m-aviso" role="alert">
				<Icon icon={I.phWarningCircleFill} size={18} />
				<span
					>Mercado Pago pidió que vuelvas a conectar la cuenta. Hasta entonces no se
					puede&nbsp;cobrar.</span
				>
			</div>
		{/if}
		<div class="m-sec">Qué hace</div>
		<ul class="m-hace">
			{#each abierta.hace as h (h)}
				<li><Icon icon={I.phCheckCircleFill} size={16} />{h}</li>
			{/each}
		</ul>
		{#if mp && data.mp.usuario}
			<div class="m-cuenta">
				Cuenta de Mercado Pago <b>#{data.mp.usuario}</b>. La plata de cada venta va directo ahí: Tu
				Local Web no cobra&nbsp;comisión.
			</div>
		{/if}
		{#if abierta.id === 'tg'}
			<div class="m-cuenta">
				Vinculado con <b>{data.telegram.nombre}</b> en
				<a href="https://t.me/{data.telegram.bot}" target="_blank" rel="noopener"
					>@{data.telegram.bot}</a
				>. Probá preguntarle "¿qué se vendió&nbsp;hoy?".
			</div>
		{:else if abierta.id === 'ml'}
			<div class="m-cuenta">
				Cuenta <b>{data.ml.apodo}</b> · {data.ml.publicaciones}
				{data.ml.publicaciones === 1 ? 'publicación' : 'publicaciones'} desde Tu Local Web.
				{#if data.ml.preguntas}
					<a href={resolve('/panel/resenas-y-preguntas')}
						>{data.ml.preguntas}
						{data.ml.preguntas === 1 ? 'pregunta' : 'preguntas'} sin&nbsp;responder</a
					>{/if}
			</div>
			<div class="m-sec">Precio en Mercado Libre</div>
			<div class="recargos" role="group" aria-label="Recargo por comisión">
				{#each [0, 5, 10, 15, 20] as r (r)}
					<button
						type="button"
						class:on={data.ml.recargo === r}
						aria-pressed={data.ml.recargo === r}
						onclick={() => ajustarML(r, data.ml.pausarSinStock)}>{r ? `+${r}%` : 'Igual'}</button
					>
				{/each}
			</div>
			<div class="m-ops">
				<div class="m-op">
					<span class="g-txt"
						><span class="g-t">Pausar sin stock</span><span class="g-d"
							>Las publicaciones se pausan solas y vuelven cuando hay</span
						></span
					>
					<Interruptor
						label="Pausar sin stock"
						checked={data.ml.pausarSinStock}
						onchange={(v) => void ajustarML(data.ml.recargo, v)}
					/>
				</div>
			</div>
		{:else if abierta.id === 'ga'}
			<form
				id="f-medicion"
				class="m-form"
				novalidate
				onsubmit={(e) => {
					e.preventDefault();
					void guardarId('ga');
				}}
			>
				<label class="field">
					ID de medición de GA4
					<input
						class="input"
						name="ga4"
						placeholder="G-XXXXXXXXXX"
						autocomplete="off"
						spellcheck="false"
						autocapitalize="characters"
						bind:value={ga4}
						aria-invalid={!!errorId}
					/>
				</label>
				<div class="m-ayuda">
					Está en Google Analytics: Administrar › Flujos de datos › Web, como “ID de&nbsp;medición”.
				</div>
				{#if errorId}<div class="m-error" role="alert">{errorId}</div>{/if}
			</form>
			<div class="m-cuenta">
				Se carga sólo si quien compra acepta las cookies de medición en el aviso de tu&nbsp;tienda.
			</div>
		{:else if abierta.id === 'gm' || abierta.id === 'meta'}
			{@const gm = abierta.id === 'gm'}
			{#if !gm}
				<div class="m-sec">Píxel</div>
				<form
					id="f-medicion"
					class="m-form"
					novalidate
					onsubmit={(e) => {
						e.preventDefault();
						void guardarId('meta');
					}}
				>
					<label class="field">
						ID del píxel de Meta
						<input
							class="input"
							name="pixel"
							placeholder="15 o 16 números"
							inputmode="numeric"
							autocomplete="off"
							bind:value={pixel}
							aria-invalid={!!errorId}
						/>
					</label>
					<div class="m-ayuda">
						Está en el Administrador de eventos: Orígenes de datos › tu&nbsp;píxel.
					</div>
					{#if errorId}<div class="m-error" role="alert">{errorId}</div>{/if}
				</form>
			{/if}
			{#if !gm && data.conversiones && data.medicion.pixel}
				{@const cv = data.conversiones}
				<div class="m-sec">API de conversiones</div>
				<form
					class="m-form"
					novalidate
					onsubmit={(e) => {
						e.preventDefault();
						void guardarConversiones();
					}}
				>
					{#if cv.conectada}
						<div class="m-hecho">
							<Icon icon={I.phCheckCircleFill} size={16} />{cv.ultimoEnvio
								? `Conectada · última compra enviada ${cuando(cv.ultimoEnvio)}`
								: 'Conectada · todavía no hubo compras'}
						</div>
						{#if cv.ultimoError}
							<div class="m-error" role="alert">
								Meta rechazó la última compra: {cv.ultimoError}
							</div>
						{/if}
					{:else}
						<div class="m-ayuda">
							Las compras llegan a Meta desde nuestro servidor, aunque el navegador bloquee el
							píxel: tus campañas cuentan mejor las&nbsp;ventas.
						</div>
					{/if}
					<label class="field">
						{cv.conectada ? 'Token nuevo (opcional)' : 'Token de acceso'}
						<input
							class="input"
							type="password"
							name="token"
							autocomplete="off"
							placeholder={cv.conectada ? 'Dejalo vacío para no cambiarlo' : 'EAAG…'}
							bind:value={capiToken}
						/>
					</label>
					<div class="m-ayuda">
						Administrador de eventos › tu píxel › Configuración › API de conversiones › Generar
						token de&nbsp;acceso.
					</div>
					<label class="field">
						Código de prueba (opcional)
						<input
							class="input"
							name="codigoPrueba"
							autocomplete="off"
							placeholder="TEST12345"
							bind:value={capiCodigo}
						/>
					</label>
					<div class="m-ayuda">
						Con un código, las compras se ven en «Eventos de prueba» y no cuentan para tus campañas.
						Borralo cuando veas que&nbsp;llegan.
					</div>
					{#if capiError}<div class="m-error" role="alert">{capiError}</div>{/if}
					<div class="capi-acc">
						<button type="submit" class="b-sec b-chico" disabled={capiGuardando}
							>{cv.conectada ? 'Guardar' : 'Conectar'}</button
						>
						{#if cv.conectada}
							<button
								type="button"
								class="b-sec b-chico rojo"
								disabled={capiGuardando}
								onclick={desconectarConversiones}>Desconectar</button
							>
						{/if}
					</div>
				</form>
				<div class="m-cuenta">
					Sólo se mandan las compras de quien aceptó las cookies de tu tienda, con su email y
					teléfono cifrados (SHA-256).
				</div>
			{/if}
			<div class="m-sec">{gm ? 'Feed de productos' : 'Catálogo'}</div>
			<div class="m-feed">
				<code>{data.feed}</code>
				<button type="button" class="b-copiar" onclick={copiarFeed}
					><Icon icon={I.phCopyBold} size={15} />Copiar</button
				>
			</div>
			<ol class="m-pasos">
				{#if gm}
					<li>
						Entrá a <a href="https://merchants.google.com" target="_blank" rel="noopener"
							>Merchant Center</a
						> con la cuenta de Google de tu&nbsp;tienda.
					</li>
					<li>Andá a <b>Productos › Fuentes de datos › Agregar</b>.</li>
					<li>
						Elegí <b>Archivo programado</b> (o URL), pegá el link y dejá la descarga todos los&nbsp;días.
					</li>
					<li>Guardá. En unos días tus productos aparecen gratis en&nbsp;Google.</li>
				{:else}
					<li>
						Entrá a <a href="https://business.facebook.com/commerce" target="_blank" rel="noopener"
							>Commerce Manager</a
						> y elegí tu catálogo (o creá uno de&nbsp;comercio).
					</li>
					<li>Andá a <b>Catálogo › Orígenes de datos › Agregar artículos</b>.</li>
					<li>
						Elegí <b>Feed programado</b>, pegá el link y que se actualice todos los&nbsp;días.
					</li>
				{/if}
			</ol>
			<div class="m-feed-st">
				Precio, oferta y stock salen solos de tu tienda: el feed se actualiza cada&nbsp;hora.
			</div>
			{#if !gm}
				{#if data.medicion.feeds.meta}
					<div class="m-hecho"><Icon icon={I.phCheckCircleFill} size={16} />Catálogo cargado</div>
				{:else}
					<button type="button" class="b-sec b-chico" onclick={() => yaLoCargue('meta')}
						>Ya cargué el catálogo</button
					>
				{/if}
				<div class="m-cuenta">
					El píxel se carga sólo si quien compra acepta las cookies de tu&nbsp;tienda.
				</div>
			{/if}
		{:else if abierta.opciones.length}
			<div class="m-sec">Opciones</div>
			<div class="m-ops">
				{#each abierta.opciones as o (o.id)}
					{@const id = abierta.id}
					<div class="m-op">
						<span class="g-txt"><span class="g-t">{o.t}</span><span class="g-d">{o.d}</span></span>
						<Interruptor
							label={o.t}
							checked={data.estado.opciones[id]?.[o.id] ?? o.def}
							onchange={(v) => void ajustar({ tipo: 'opcion', id, clave: o.id, valor: v })}
						/>
					</div>
				{/each}
			</div>
		{/if}
		<div class="m-acc">
			{#if conectada}
				<button type="button" class="b-sec rojo" onclick={() => abierta && desconectar(abierta)}
					>Desconectar</button
				>
			{/if}
			{#if abierta.id === 'ga' || abierta.id === 'meta'}
				<button type="submit" class="b-pri" form="f-medicion">Guardar</button>
			{:else if abierta.id === 'gm' && !conectada}
				<button type="button" class="b-pri" onclick={() => yaLoCargue('gm')}>Ya lo cargué</button>
			{:else if (mp && data.mp.reautorizar) || (abierta.id === 'ml' && data.ml.reautorizar)}
				<button type="button" class="b-pri" onclick={() => conectarMP(mp ? 'mp' : 'ml')}
					>Volver a conectar</button
				>
			{:else}
				<button type="button" class="b-pri" onclick={() => (modal = false)}>Listo</button>
			{/if}
		</div>
	{/if}
</Modal>

<Arca bind:open={arcaAbierto} arca={data.arca} demo={data.demo} />

<Correo bind:open={correoAbierto} correo={data.correo} demo={data.demo} />

<Modal
	bind:open={modalIa}
	titulo={iaAbierta ? `Conectar ${iaAbierta.nombre}` : 'Conectar tu IA'}
	ancho={480}
	padding={24}
>
	{#if iaAbierta}
		{@const p = iaAbierta}
		{@const activas = conexionesDe(p.id)}
		<div class="m-cab">
			<span class="logo grande" style:background="var(--sf3)" style:color={p.col}
				><Icon icon={p.icon} size={26} /></span
			>
			<div>
				<div class="m-t">{p.nombre}</div>
				<div class="m-st" class:apagado={!activas.length}>
					{activas.length ? '● Conectado' : 'Servidor MCP con tu\u00a0cuenta'}
				</div>
			</div>
		</div>
		<div class="m-sec">Dirección del servidor</div>
		<div class="m-url">
			<code>{data.mcp.url}</code>
			<button type="button" class="b-sec" onclick={copiarUrl}
				><Icon icon={I.phCopyBold} size={15} />Copiar</button
			>
		</div>
		<div class="m-sec">Cómo conectarla</div>
		<ol class="m-pasos">
			{#each PASOS[p.id] as paso, i (i)}
				<li><span class="n">{i + 1}</span>{paso}</li>
			{/each}
		</ol>
		<div class="m-cuenta">
			Al permitir el acceso elegís si sólo puede leer o también hacer&nbsp;cambios.
		</div>
		{#if activas.length}
			<div class="m-sec">Conectadas</div>
			<div class="m-ops">
				{#each activas as c (c.id)}
					<div class="m-op">
						<span class="g-txt"
							><span class="g-t">{c.nombre}</span><span class="g-d"
								>{c.acciones ? 'Lectura y acciones' : 'Sólo lectura'} · desde las {hora(
									c.desde
								)}</span
							></span
						>
						<button
							type="button"
							class="t-btn on rojo"
							onclick={() => desconectarIa(c.id, c.nombre)}>Desconectar</button
						>
					</div>
				{/each}
			</div>
		{/if}
		<div class="m-acc">
			<button type="button" class="b-pri" onclick={() => (modalIa = false)}>Listo</button>
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
	.sub {
		margin-top: 2px;
		font-size: 15px;
		color: var(--mu);
	}
	.ia {
		padding: 20px;
		border-radius: 24px;
		background: var(--sf);
		box-shadow: var(--shd);
		display: flex;
		flex-direction: column;
		gap: 14px;
	}
	.ia-cab {
		display: flex;
		align-items: center;
		gap: 14px;
		flex-wrap: wrap;
	}
	.ia-cab img {
		width: 52px;
		height: 52px;
		object-fit: contain;
		flex: none;
	}
	.ia-txt {
		flex: 1 1 280px;
		min-width: 0;
	}
	.ia-t {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 17px;
		font-weight: 600;
	}
	.mcp {
		font-size: 11px;
		font-weight: 700;
		padding: 3px 8px;
		border-radius: 999px;
		background: rgba(210, 88, 47, 0.12);
		color: #b0441d;
	}
	.ia-d {
		font-size: 13.5px;
		color: var(--mu);
		margin-top: 2px;
	}
	.provs {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 10px;
	}
	.prov {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 12px 14px;
		border-radius: 16px;
		border: 1.5px solid var(--ln2);
		background: transparent;
		color: var(--mu);
		cursor: pointer;
		font-family: inherit;
		text-align: left;
	}
	.prov.on {
		border-color: rgba(47, 143, 99, 0.5);
		background: rgba(47, 143, 99, 0.06);
		color: #237550;
	}
	.p-ico {
		width: 36px;
		height: 36px;
		flex: none;
		border-radius: 11px;
		display: grid;
		place-items: center;
		background: var(--sf3);
	}
	.p-txt {
		flex: 1;
		min-width: 0;
	}
	.p-t {
		display: block;
		font-size: 14px;
		font-weight: 600;
		color: var(--tx);
	}
	.p-st {
		display: block;
		font-size: 12px;
		font-weight: 600;
	}
	.perms {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
		gap: 4px 18px;
		padding-top: 4px;
		border-top: 1px solid var(--ln);
	}
	.perm {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 10px 0;
		border: 0;
		background: transparent;
		cursor: pointer;
		font-family: inherit;
		color: var(--mu);
		text-align: left;
	}
	.perm:disabled {
		cursor: default;
	}
	.g-txt {
		flex: 1;
		min-width: 0;
	}
	.g-t {
		display: block;
		font-size: 13.5px;
		font-weight: 600;
		color: var(--tx);
	}
	.g-d {
		display: block;
		font-size: 12px;
		color: var(--fa);
	}
	.tg {
		position: relative;
		width: 36px;
		height: 21px;
		flex: none;
		border-radius: 999px;
		background: #d9cbbe;
		transition: background 0.25s;
	}
	.tg.on {
		background: #d2582f;
	}
	.tg span {
		position: absolute;
		top: 2px;
		left: 2px;
		width: 17px;
		height: 17px;
		border-radius: 50%;
		background: #fff;
		transition: left 0.25s;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
	}
	.tg.on span {
		left: 17px;
	}
	.cats {
		display: flex;
		gap: 6px;
		overflow-x: auto;
		scrollbar-width: none;
	}
	.cats button {
		border: 0;
		cursor: pointer;
		white-space: nowrap;
		font:
			600 13px 'Instrument Sans',
			sans-serif;
		padding: 9px 14px;
		border-radius: 999px;
		background: var(--sf);
		color: var(--tx);
		box-shadow: var(--sh);
	}
	.cats button.on {
		background: var(--tx);
		color: var(--sf);
	}
	.grilla {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
		gap: 12px;
	}
	.tarjeta {
		padding: 16px;
		border-radius: 20px;
		background: var(--sf);
		box-shadow: var(--shd);
		display: flex;
		flex-direction: column;
		gap: 12px;
		border: 1.5px solid transparent;
		transition: border-color 0.3s;
	}
	.tarjeta.on {
		border-color: rgba(47, 143, 99, 0.35);
	}
	.t-cab {
		display: flex;
		align-items: center;
		gap: 10px;
	}
	.logo {
		width: 42px;
		height: 42px;
		flex: none;
		border-radius: 12px;
		display: grid;
		place-items: center;
	}
	.logo.grande {
		width: 52px;
		height: 52px;
		border-radius: 16px;
	}
	.sigla {
		font-size: 8.5px;
		font-weight: 800;
	}
	.t-txt {
		flex: 1;
		min-width: 0;
	}
	.t-n {
		display: block;
		font-size: 15px;
		font-weight: 600;
	}
	.t-d {
		display: block;
		font-size: 12.5px;
		color: var(--mu);
	}
	.t-pie {
		display: flex;
		align-items: center;
		gap: 8px;
	}
	.t-st {
		flex: 1;
		font-size: 12.5px;
		font-weight: 600;
		color: var(--fa);
	}
	.t-st.on {
		color: #237550;
	}
	.t-st.vence {
		color: #8a5f0e;
	}
	.t-btn {
		padding: 8px 14px;
		border-radius: 999px;
		border: 0;
		background: var(--tx);
		color: var(--sf);
		font:
			600 12.5px 'Instrument Sans',
			sans-serif;
		cursor: pointer;
		white-space: nowrap;
	}
	.t-btn:disabled {
		background: var(--chip);
		color: var(--mu);
		cursor: default;
	}
	.t-btn.on {
		border: 1px solid var(--ln2);
		background: transparent;
		color: var(--tx);
	}
	.m-cab {
		display: flex;
		align-items: center;
		gap: 12px;
	}
	.m-t {
		font-size: 19px;
		font-weight: 600;
	}
	.m-st {
		font-size: 13px;
		font-weight: 600;
		color: #237550;
	}
	.m-st.apagada {
		color: var(--fa);
	}
	.m-form {
		margin-top: 12px;
		display: grid;
		gap: 6px;
	}
	.m-form .input {
		font-family: inherit;
		letter-spacing: 0.02em;
	}
	.m-ayuda {
		font-size: 12.5px;
		line-height: 1.45;
		color: var(--fa);
	}
	.m-error {
		font-size: 13px;
		font-weight: 600;
		color: #b83a2e;
	}
	.m-feed {
		margin-top: 8px;
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 8px 8px 8px 14px;
		border-radius: 14px;
		background: var(--sf2);
	}
	.m-feed code {
		flex: 1;
		min-width: 0;
		font-size: 12.5px;
		line-height: 1.4;
		overflow-wrap: anywhere;
		color: var(--tx);
	}
	.b-copiar {
		flex: none;
		display: inline-flex;
		align-items: center;
		gap: 6px;
		min-height: 44px;
		padding: 8px 14px;
		border: 0;
		border-radius: 999px;
		background: var(--tx);
		color: var(--sf);
		font:
			600 12.5px 'Instrument Sans',
			sans-serif;
		cursor: pointer;
	}
	.m-pasos {
		margin: 12px 0 0;
		padding-left: 20px;
		display: flex;
		flex-direction: column;
		gap: 6px;
		font-size: 13.5px;
		line-height: 1.45;
	}
	.m-pasos a {
		color: #b0441d;
		font-weight: 600;
	}
	.m-feed-st {
		margin-top: 10px;
		font-size: 12.5px;
		color: var(--fa);
	}
	.m-hecho {
		margin-top: 12px;
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 13.5px;
		font-weight: 600;
		color: #237550;
	}
	.capi-acc {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
	}
	.b-chico {
		margin-top: 12px;
		padding: 10px 16px;
		min-height: 44px;
		border-radius: 999px;
		font:
			600 13px 'Instrument Sans',
			sans-serif;
		cursor: pointer;
	}
	.m-aviso {
		margin-top: 14px;
		display: flex;
		gap: 8px;
		align-items: flex-start;
		padding: 12px 14px;
		border-radius: 14px;
		background: rgba(227, 169, 59, 0.16);
		color: #8a5f0e;
		font-size: 13.5px;
		line-height: 1.45;
	}
	.m-sec {
		margin-top: 16px;
		font-size: 12px;
		font-weight: 700;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		color: var(--fa);
	}
	.m-hace {
		margin: 8px 0 0;
		padding: 0;
		list-style: none;
		display: flex;
		flex-direction: column;
		gap: 7px;
	}
	.m-hace li {
		display: flex;
		gap: 8px;
		align-items: flex-start;
		font-size: 14px;
		line-height: 1.4;
	}
	.m-hace li :global(svg) {
		flex: none;
		margin-top: 2px;
		color: #237550;
	}
	.m-cuenta {
		margin-top: 12px;
		padding: 12px 14px;
		border-radius: 14px;
		background: var(--sf2);
		font-size: 13px;
		line-height: 1.5;
		color: var(--mu);
	}
	.m-cuenta b {
		color: var(--tx);
	}
	.m-cuenta a {
		color: #b0441d;
		font-weight: 600;
	}
	.espera {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 12px 16px;
		border-radius: 16px;
		background: #e8f4fb;
		font-size: 14px;
		flex-wrap: wrap;
	}
	.espera span {
		flex: 1;
		min-width: 180px;
	}
	.espera button {
		border: 0;
		border-radius: 999px;
		padding: 8px 14px;
		background: var(--tx);
		color: var(--sf);
		font:
			600 12.5px 'Instrument Sans',
			sans-serif;
		cursor: pointer;
	}
	.recargos {
		margin-top: 8px;
		display: flex;
		gap: 6px;
		flex-wrap: wrap;
	}
	.recargos button {
		border: 1px solid var(--ln2);
		background: transparent;
		color: var(--tx);
		padding: 8px 14px;
		border-radius: 999px;
		font:
			600 13px 'Instrument Sans',
			sans-serif;
		cursor: pointer;
	}
	.recargos button.on {
		background: var(--tx);
		color: var(--sf);
		border-color: var(--tx);
	}
	.m-ops {
		margin-top: 4px;
		display: flex;
		flex-direction: column;
	}
	.m-op {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 10px 0;
		border-bottom: 1px solid var(--ln);
	}
	.m-op:last-child {
		border-bottom: 0;
	}
	.m-acc {
		margin-top: 18px;
		display: flex;
		gap: 8px;
	}
	.m-acc button {
		flex: 1;
		padding: 13px;
		border-radius: 999px;
		font:
			600 14px 'Instrument Sans',
			sans-serif;
		cursor: pointer;
	}
	.b-sec {
		border: 1px solid var(--ln2);
		background: transparent;
		color: var(--tx);
	}
	.b-sec.rojo {
		color: #b83a2e;
	}
	.b-pri {
		border: 0;
		background: var(--tx);
		color: var(--sf);
	}
	/* Servidor MCP: piezas del mismo modal (m-cuenta, b-sec, t-btn) y de la lista de permisos. */
	.conexiones {
		display: flex;
		flex-direction: column;
		border-top: 1px solid var(--ln);
	}
	.cx {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 10px 0;
		border-bottom: 1px solid var(--ln);
	}
	.cx:last-child {
		border-bottom: 0;
	}
	.t-btn.rojo {
		color: #b83a2e;
	}
	.m-st.apagado {
		color: var(--fa);
	}
	.m-url {
		margin-top: 8px;
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 8px 8px 8px 14px;
		border-radius: 14px;
		background: var(--sf2);
	}
	.m-url code {
		flex: 1;
		min-width: 0;
		font-size: 13px;
		overflow-wrap: anywhere;
		user-select: all;
	}
	.m-url .b-sec {
		display: flex;
		align-items: center;
		gap: 6px;
		flex: none;
		min-height: 44px;
		padding: 0 14px;
		border-radius: 999px;
		font:
			600 13px 'Instrument Sans',
			sans-serif;
		cursor: pointer;
	}
	.m-pasos {
		margin: 8px 0 0;
		padding: 0;
		list-style: none;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.m-pasos li {
		display: flex;
		gap: 10px;
		align-items: flex-start;
		font-size: 14px;
		line-height: 1.45;
	}
	.m-pasos .n {
		width: 22px;
		height: 22px;
		flex: none;
		border-radius: 50%;
		display: grid;
		place-items: center;
		background: var(--tx);
		color: var(--sf);
		font-size: 12px;
		font-weight: 700;
	}
</style>
