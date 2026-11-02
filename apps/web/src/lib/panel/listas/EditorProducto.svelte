<script lang="ts">
	import Hoja from '#lib/components/Hoja.svelte';
	import Icon from '#lib/components/Icon.svelte';
	import Interruptor from '#lib/components/Interruptor.svelte';
	import * as I from '#lib/icons.ts';
	import { unsplash } from '#lib/landing/formato.ts';
	import { getUi } from '#lib/ui.svelte.ts';
	import { page } from '$app/state';
	import {
		colorDe,
		errorOpciones,
		generarVariantes,
		MAX_FOTOS,
		MAX_OPCIONES,
		MAX_VALORES,
		normal,
		type Variante
	} from '#lib/tienda/variantes.ts';
	import type { components } from '@mostrador/contrato/esquema';
	import type { VentaMayorista } from '#lib/tienda/mayorista.ts';
	import {
		adivinarTipo,
		CAMPOS,
		tipoDe,
		TIPOS,
		type Atributo,
		type TipoProducto
	} from '#lib/tienda/tipo-producto.ts';
	import { compartirProducto } from '../compartir.ts';
	import { subir } from '../fotos.ts';
	import { pesos } from '../formato.ts';
	import LanzamientoProducto from './LanzamientoProducto.svelte';

	type Producto = components['schemas']['Producto'];

	/**
	 * Editor de producto (design/PanelCatalogo.dc.html, `edOpen`): fotos, nombre, descripción (con
	 * "Mejorar con Tincho"), precio por canal, stock, dónde se publica y cómo aparece en Google.
	 * Con `producto` null es el alta.
	 */
	let {
		open = $bindable(false),
		producto,
		tienda,
		dominio,
		recargoML = null,
		mayorista = null,
		guardado,
		borrado
	}: {
		open?: boolean;
		producto: Producto | null;
		tienda: string;
		dominio: string;
		/** El recargo de la conexión con Mercado Libre; null si no está conectado. */
		recargoML?: number | null;
		/** Configuración › Venta mayorista (§49): el mínimo, para la nota del precio mayorista. */
		mayorista?: VentaMayorista | null;
		guardado: (p: Producto, nuevo: boolean) => void;
		borrado: (nombre: string) => void;
	} = $props();
	const ui = getUi();

	let nombre = $state('');
	let descripcion = $state('');
	let precio = $state(0);
	// Precio mayorista (§49): 0 es «no tiene» (así lo saca Go al guardar).
	let precioMayor = $state(0);
	let costo = $state(0);
	// Ventas edita el producto pero no los precios ni los costos (docs/decisiones.md §37): se ven sin
	// poder cambiarlos y el costo no se muestra. Go igual conserva los que había.
	const conPrecios = $derived((page.data.acceso?.permisos ?? ['catalogo']).includes('catalogo'));
	let stock = $state(0);
	let categoria = $state('');
	let canales = $state<Record<string, boolean>>({ web: true, ml: false, ig: false, google: false });
	// Talle, color, presentación…: con opciones, el stock va por combinación.
	let opciones = $state<{ nombre: string; valores: string[] }[]>([]);
	let variantes = $state<Variante[]>([]);
	let nuevoValor = $state<string[]>([]);
	// Qué es (ropa, comida, un mueble…): según eso se piden sus datos y se proponen las opciones. En
	// el alta se pregunta primero; los productos de antes lo adivinan por sus opciones.
	let tipo = $state<TipoProducto | null>(null);
	let atributos = $state<Record<string, string>>({});
	let eligiendoTipo = $state(false);
	const defTipo = $derived(tipoDe(tipo));
	const ICONOS_TIPO: Record<TipoProducto, I.IconData> = {
		ropa: I.phTShirtFill,
		calzado: I.phSneakerFill,
		alimentos: I.phBowlFoodFill,
		belleza: I.phSparkleFill,
		hogar: I.phCouchFill,
		tecnologia: I.phDeviceMobileFill,
		otro: I.phDotsThreeOutlineFill
	};
	function elegirTipo(t: TipoProducto) {
		tipo = t;
		eligiendoTipo = false;
	}
	// La galería: la primera es la portada.
	let fotos = $state<string[]>([]);
	let subiendo = $state(0);
	let elegirFoto = $state<HTMLInputElement>();
	let error = $state<string | null>(null);
	let guardando = $state(false);
	let lanzamiento = $state<ReturnType<typeof LanzamientoProducto>>();
	// Borrar pide un segundo toque: «¿Lo borro?».
	let confirmando = $state(false);
	$effect(() => {
		if (!open) confirmando = false;
	});

	const mensaje = async (r: Response, si: string) =>
		((await r.json().catch(() => null)) as { message?: string } | null)?.message ?? si;
	async function duplicar() {
		if (!producto) return;
		guardando = true;
		const r = await fetch(`/panel/productos/${encodeURIComponent(producto.id)}/duplicar`, {
			method: 'POST'
		});
		guardando = false;
		if (!r.ok) {
			error = await mensaje(r, 'No pudimos duplicarlo.');
			return;
		}
		open = false;
		guardado((await r.json()) as Producto, true);
	}
	async function borrar() {
		if (!producto) return;
		if (!confirmando) {
			confirmando = true;
			return;
		}
		guardando = true;
		const r = await fetch(`/panel/productos/${encodeURIComponent(producto.id)}`, {
			method: 'DELETE'
		});
		guardando = false;
		confirmando = false;
		if (!r.ok) {
			error = await mensaje(r, 'No pudimos borrarlo.');
			return;
		}
		open = false;
		borrado(producto.nombre);
	}
	let mejorando = $state(false);

	$effect(() => {
		if (!open) return;
		const p = producto;
		nombre = p?.nombre ?? '';
		descripcion = p?.descripcion ?? '';
		precio = p?.precio ?? 0;
		precioMayor = p?.precioMayorista ?? 0;
		costo = p?.costo ?? 0;
		stock = p?.stock ?? 0;
		categoria = p?.categoria ?? '';
		canales = { web: true, ml: false, ig: false, google: false, ...(p?.canales ?? {}) };
		opciones = (p?.opciones ?? []).map((o) => ({ nombre: o.nombre, valores: [...o.valores] }));
		// Producto = ProductoEditable & {variantes: Variante[]}: el map de la intersección toma la editable.
		variantes = ((p?.variantes ?? []) as Variante[]).map((v) => ({
			...v,
			valores: { ...v.valores }
		}));
		nuevoValor = (p?.opciones ?? []).map(() => '');
		tipo = (p?.tipo as TipoProducto | null | undefined) ?? adivinarTipo(p?.opciones ?? []);
		atributos = { ...(p?.atributos ?? {}) };
		eligiendoTipo = !p;
		fotos = p?.fotos?.length ? [...p.fotos] : p?.foto ? [p.foto] : [];
		error = null;
	});

	// Sin costo cargado el margen no se sabe (no es 100 %).
	const margen = $derived(
		precio && costo > 0 ? Math.round(((precio - costo) / precio) * 100) : null
	);
	// El link no cambia al renombrar (docs: migración 00043): el de un producto ya creado es el suyo.
	const slug = $derived(
		producto?.slug ||
			nombre
				.normalize('NFD')
				.replace(/[̀-ͯ]/g, '')
				.toLowerCase()
				.replace(/[^a-z0-9]+/g, '-')
				.replace(/^-|-$/g, '') ||
			'producto'
	);
	/**
	 * El precio de la tienda y, con Mercado Libre conectado, el que se publica allá: con el recargo
	 * real de la conexión y el mismo redondeo que la API (precioML). Antes había un «Mayorista · Desde
	 * 10 u» y un «+10 %» fijos que no existían.
	 */
	const PRECIOS = $derived<(readonly [string, number, string])[]>(
		recargoML === null
			? [['Tienda y local', precio, 'Precio base']]
			: [
					['Tienda y local', precio, 'Precio base'],
					[
						'Mercado Libre',
						Math.floor((precio * (100 + recargoML) + 50) / 100),
						recargoML ? `+${recargoML} % por comisión` : 'Sin recargo'
					]
				]
	);
	/**
	 * La caja «Mayorista» del diseño (precio por canal, «Desde 10 u»), con el mínimo de verdad de
	 * Configuración › Venta mayorista. Con la venta apagada sólo aparece si el producto ya tenía uno.
	 */
	const notaMayorista = $derived(
		!mayorista?.activa
			? 'Venta por mayor apagada'
			: [
					mayorista.unidades !== null ? `Desde ${mayorista.unidades} u` : '',
					mayorista.monto !== null ? `desde ${pesos(mayorista.monto)}` : ''
				]
					.filter(Boolean)
					.join(' y ')
					.replace(/^d/, 'D')
	);
	const verMayorista = $derived(!!mayorista?.activa || precioMayor > 0);
	const mayorMal = $derived(precioMayor > 0 && precioMayor >= precio);
	const CANALES = $derived(
		(
			[
				['web', I.phGlobeSimpleFill, 'Tienda online', '#D2582F'],
				['ml', I.phHandshakeFill, 'Mercado Libre', '#2D3277'],
				['google', I.logosGoogleIcon, 'Google Shopping', 'inherit']
			] as const
		).filter(([k]) => k !== 'ml' || recargoML !== null)
	);
	const numero = (s: string) => Math.max(0, Number(s.replace(/\D/g, '')) || 0);

	// Las opciones completas (con nombre y algún valor) arman las combinaciones.
	const listas = $derived(opciones.filter((o) => o.nombre.trim() && o.valores.length));
	const total = $derived(variantes.reduce((s, v) => s + (v.activa ? v.stock : 0), 0));
	/** Rearma las combinaciones conservando lo que ya tenían (stock, precio, SKU, foto). */
	function recombinar() {
		variantes = generarVariantes(
			listas.map((o) => ({ nombre: o.nombre.trim(), valores: o.valores })),
			variantes
		);
	}
	function sumarOpcion() {
		if (opciones.length >= MAX_OPCIONES) return;
		const usadas = new Set(opciones.map((o) => normal(o.nombre)));
		const sugeridas = defTipo?.opciones.map((o) => o.nombre) ?? ['Talle', 'Color', 'Presentación'];
		const nombre = sugeridas.find((n) => !usadas.has(normal(n))) ?? '';
		opciones = [...opciones, { nombre, valores: [] }];
		nuevoValor = [...nuevoValor, ''];
	}
	function quitarOpcion(i: number) {
		opciones = opciones.filter((_, j) => j !== i);
		nuevoValor = nuevoValor.filter((_, j) => j !== i);
		recombinar();
	}
	/** Al renombrar una opción, las variantes siguen con sus valores bajo el nombre nuevo. */
	function renombrar(i: number, nombre: string) {
		const antes = opciones[i].nombre.trim();
		opciones[i].nombre = nombre;
		const ahora = nombre.trim();
		if (!antes || !ahora || antes === ahora) return recombinar();
		variantes = variantes.map((v) => {
			const { [antes]: x, ...resto } = v.valores;
			return x === undefined ? v : { ...v, valores: { ...resto, [ahora]: x } };
		});
		recombinar();
	}
	/** Suma los valores escritos (uno o varios separados por coma). */
	function sumarValores(i: number) {
		const o = opciones[i];
		// eslint-disable-next-line svelte/prefer-svelte-reactivity -- local, sólo evita repetidos
		const vistos = new Set(o.valores.map(normal));
		for (const x of nuevoValor[i].split(',').map((v) => v.trim())) {
			if (!x || x.length > 20 || vistos.has(normal(x)) || o.valores.length >= MAX_VALORES) continue;
			vistos.add(normal(x));
			o.valores.push(x);
		}
		nuevoValor[i] = '';
		recombinar();
	}
	function quitarValor(i: number, valor: string) {
		opciones[i].valores = opciones[i].valores.filter((v) => v !== valor);
		recombinar();
	}
	const camposTipo = $derived<Atributo[]>(
		(defTipo?.campos ?? []).flatMap((c) => (c === 'medidas' ? ['alto', 'ancho', 'largo'] : [c]))
	);
	const textoSumar = $derived.by(() => {
		if (opciones.length) return 'Sumar otra opción';
		if (!defTipo) return 'Sumar talles, colores u otra opción';
		const ns = defTipo.opciones.map((o) => o.nombre.toLowerCase());
		return `Sumar ${ns.join(', ')} u otra opción`;
	});
	const ejemploOpcion = (nombre: string) =>
		defTipo?.opciones.find((o) => normal(o.nombre) === normal(nombre))?.ejemplo ?? 'S, M, L…';
	const nombreVariante = (v: Variante) => listas.map((o) => v.valores[o.nombre.trim()]).join(' · ');
	const colorVariante = (v: Variante) =>
		listas.map((o) => colorDe(o.nombre, v.valores[o.nombre.trim()])).find(Boolean) ?? null;

	async function mejorar() {
		if (nombre.trim().length < 2) {
			error = 'Poné el nombre así Tincho sabe de qué se trata.';
			return;
		}
		mejorando = true;
		const r = await fetch('/panel/hornero/mejorar', {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({ nombre, descripcion })
		});
		mejorando = false;
		if (r.ok) {
			const anterior = descripcion;
			descripcion = ((await r.json()) as { texto: string }).texto;
			ui.toast({
				k: 'ai',
				title: 'Descripción mejorada',
				desc: 'Revisala antes de guardar',
				action: 'Deshacer',
				onAction: () => {
					descripcion = anterior;
				}
			});
		} else
			error =
				((await r.json().catch(() => null)) as { error?: string } | null)?.error ??
				'Tincho no pudo escribir ahora. Probá de nuevo en un rato.';
	}

	async function guardar() {
		error = null;
		if (nombre.trim().length < 2) {
			error = 'Poné un nombre de al menos 2 letras.';
			return;
		}
		const ops = listas.map((o) => ({ nombre: o.nombre.trim(), valores: o.valores }));
		if (opciones.length !== listas.length) {
			error = 'Cada opción necesita un nombre y al menos un valor.';
			return;
		}
		if (mayorMal) {
			error = 'El precio mayorista tiene que ser menor que el precio.';
			return;
		}
		const mal = errorOpciones(ops) ?? lanzamiento?.revisar() ?? null;
		if (mal) {
			error = mal;
			return;
		}
		guardando = true;
		const r = await fetch(`/panel/productos/${producto?.id ?? 'nuevo'}`, {
			method: 'PUT',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({
				nombre,
				descripcion,
				categoria: categoria.trim() || null,
				precio,
				precioMayorista: precioMayor,
				costo,
				stock,
				foto: fotos[0] ?? null,
				fotos,
				opciones: ops,
				variantes: variantes.map((v) => ({
					valores: v.valores,
					precio: v.precio,
					precioMayorista: v.precioMayorista ?? null,
					stock: v.stock,
					activa: v.activa,
					// Si se sacó de la galería la foto de una variante, la variante queda sin foto.
					...(v.foto && !fotos.includes(v.foto) ? { foto: null } : {})
				})),
				canales,
				...(tipo ? { tipo } : {}),
				// Sólo los datos de este tipo: lo que se cargó con otro tipo se va al cambiarlo.
				atributos: Object.fromEntries(
					Object.entries(atributos).filter(
						([k, x]) => x.trim() && camposTipo.includes(k as Atributo)
					)
				)
			})
		});
		if (!r.ok) {
			guardando = false;
			error =
				((await r.json().catch(() => null)) as { message?: string } | null)?.message ??
				'No pudimos guardarlo.';
			return;
		}
		const g = (await r.json()) as Producto & {
			mercadoLibre?: { publicado?: string; error?: string };
		};
		// Lanzamiento, señas y oferta (§52): van aparte, con el id ya guardado.
		const errorLanzamiento = (await lanzamiento?.guardar(g.id)) ?? null;
		guardando = false;
		if (errorLanzamiento) {
			// El producto quedó guardado: el editor sigue abierto para corregir el lanzamiento.
			error = errorLanzamiento;
			guardado(g, !producto);
			return;
		}
		open = false;
		guardado(g, !producto);
		// Mercado Libre contesta al publicar: si rechazó, se dice por qué (y el canal queda apagado).
		if (g.mercadoLibre?.error)
			ui.toast({ k: 'warn', title: 'No se publicó en Mercado Libre', desc: g.mercadoLibre.error });
		else if (g.mercadoLibre?.publicado)
			ui.toast({
				k: 'ok',
				title: 'Publicado en Mercado Libre',
				desc: 'El stock y el precio se sincronizan solos'
			});
	}

	/** Sube las fotos elegidas (de a una, hasta completar la galería). */
	async function sumarFotos(archivos: FileList | null) {
		const lista = [...(archivos ?? [])].slice(0, MAX_FOTOS - fotos.length);
		if (elegirFoto) elegirFoto.value = '';
		for (const a of lista) {
			subiendo++;
			try {
				fotos = [...fotos, await subir(a)];
			} catch (e) {
				ui.toast({ k: 'warn', title: 'No se subió la foto', desc: (e as Error).message });
			} finally {
				subiendo--;
			}
		}
	}
	const sacarFoto = (f: string) => (fotos = fotos.filter((x) => x !== f));
	const portada = (f: string) => (fotos = [f, ...fotos.filter((x) => x !== f)]);
</script>

<Hoja bind:open titulo={producto ? 'Editar producto' : 'Nuevo producto'} ancho={560}>
	{#snippet cabecera()}
		<div class="cab">
			<span class="c-t">{producto ? 'Editar producto' : 'Nuevo producto'}</span>
			<button type="button" class="redondo" aria-label="Cerrar" onclick={() => (open = false)}
				><Icon icon={I.phXBold} size={14} /></button
			>
		</div>
	{/snippet}

	<div class="cuerpo">
		{#if eligiendoTipo}
			<div class="tarjeta">
				<div class="t">¿Qué vas a vender?</div>
				<p class="t-sub">Así te pedimos sólo los datos que van con tu producto.</p>
				<div class="tipos">
					{#each TIPOS as t (t.id)}
						<button
							type="button"
							class="tipo"
							class:elegido={tipo === t.id}
							aria-pressed={tipo === t.id}
							onclick={() => elegirTipo(t.id)}
							><Icon icon={ICONOS_TIPO[t.id]} size={22} /><span class="tp-n">{t.nombre}</span><span
								class="tp-e">{t.ejemplo}</span
							></button
						>
					{/each}
				</div>
			</div>
		{:else}
			<input
				bind:this={elegirFoto}
				type="file"
				accept="image/jpeg,image/png,image/webp"
				multiple
				hidden
				aria-label="Elegir fotos"
				onchange={(e) => sumarFotos(e.currentTarget.files)}
			/>
			<div class="fotos">
				{#each fotos as f, i (f)}
					<div class="foto" class:f-1={i === 0}>
						<img src={unsplash(f, i === 0 ? 500 : 200, i === 0 ? '&h=500' : '&h=200')} alt="" />
						<button
							type="button"
							class="f-x"
							aria-label="Sacar la foto {i + 1}"
							onclick={() => sacarFoto(f)}><Icon icon={I.phXBold} size={11} /></button
						>{#if i > 0}<button
								type="button"
								class="f-portada"
								aria-label="Usar la foto {i + 1} de portada"
								onclick={() => portada(f)}><Icon icon={I.phStarFill} size={11} /></button
							>{/if}
					</div>
				{:else}
					<button
						type="button"
						class="f-1 f-vacia"
						onclick={() => elegirFoto?.click()}
						disabled={subiendo > 0}
						><Icon icon={I.phCameraPlusFill} size={28} /><span
							>{subiendo ? 'Subiendo…' : 'Sumá la primera foto'}</span
						></button
					>
				{/each}
				{#if fotos.length && fotos.length < MAX_FOTOS}<button
						type="button"
						class="f-mas"
						aria-label={subiendo ? 'Subiendo la foto' : 'Sumar una foto'}
						onclick={() => elegirFoto?.click()}
						disabled={subiendo > 0}
						><Icon icon={subiendo ? I.phSpinnerGap : I.phPlusBold} size={18} /></button
					>{/if}
			</div>

			<div class="tarjeta campos">
				<div class="tipo-actual">
					{#if defTipo}<Icon icon={ICONOS_TIPO[defTipo.id]} size={18} /><span
							><b>{defTipo.nombre}</b></span
						>{:else}<span>¿Qué tipo de producto es?</span>{/if}
					<button type="button" class="hornero" onclick={() => (eligiendoTipo = true)}
						>{defTipo ? 'Cambiar' : 'Elegir'}</button
					>
				</div>
				<label class="campo"
					>Nombre<input
						bind:value={nombre}
						maxlength="120"
						aria-invalid={error && nombre.trim().length < 2 ? 'true' : undefined}
					/></label
				>
				<div class="campo">
					<span class="c-fila"
						><label for="desc-prod">Descripción</label><button
							type="button"
							class="hornero"
							onclick={mejorar}
							disabled={mejorando}
							>{mejorando ? 'Tincho está escribiendo…' : '✦ Mejorar con Tincho'}</button
						></span
					>
					<textarea id="desc-prod" bind:value={descripcion} maxlength="4000"></textarea>
				</div>
				<label class="campo"
					>Categoría<input
						bind:value={categoria}
						maxlength="60"
						placeholder="Por ejemplo: Mates"
					/></label
				>
			</div>

			<div class="tarjeta">
				<div class="t">{recargoML === null ? 'Precio' : 'Precio por canal'}</div>
				<div class="precios">
					{#each PRECIOS as [t, v, nota], i (t)}
						<div class="precio">
							<div class="p-t">{t}</div>
							{#if i === 0}
								<input
									class="p-in"
									inputmode="numeric"
									aria-label="Precio base"
									readonly={!conPrecios}
									placeholder="$ 0"
									value={precio ? pesos(precio) : ''}
									onchange={(e) => (precio = numero(e.currentTarget.value))}
								/>
							{:else}
								<div class="p-v">{pesos(v)}</div>
							{/if}
							<div class="p-n">{nota}</div>
						</div>
					{/each}
					{#if verMayorista}
						<div class="precio">
							<div class="p-t">Mayorista</div>
							<input
								class="p-in"
								inputmode="numeric"
								aria-label="Precio mayorista"
								readonly={!conPrecios}
								placeholder="Sin precio"
								value={precioMayor ? pesos(precioMayor) : ''}
								onchange={(e) => (precioMayor = numero(e.currentTarget.value))}
							/>
							<div class="p-n" class:mal={mayorMal}>
								{mayorMal ? 'Tiene que ser menor que el precio' : notaMayorista}
							</div>
						</div>
					{/if}
				</div>
				{#if conPrecios}
					<div class="margen">
						Margen {#if margen === null}<b>sin dato</b>{:else}<b class:bajo={margen < 25}
								>{margen} %</b
							>{/if} · costo
						<input
							class="costo"
							inputmode="numeric"
							aria-label="Costo"
							placeholder="$ 0"
							value={costo ? pesos(costo) : ''}
							onchange={(e) => (costo = numero(e.currentTarget.value))}
						/>
					</div>
				{:else}
					<div class="margen">Los precios los cambia la dueña o administración.</div>
				{/if}
			</div>

			{#if camposTipo.length}
				<div class="tarjeta campos">
					<div class="t">Datos del producto</div>
					<div class="datos">
						{#each defTipo?.campos ?? [] as c (c)}
							{#if c === 'medidas'}
								<div class="campo ancho">
									Medidas en cm (alto × ancho × largo)
									<div class="medidas">
										{#each ['alto', 'ancho', 'largo'] as const as m (m)}
											<input
												inputmode="numeric"
												aria-label="{CAMPOS[m].nombre} en centímetros"
												placeholder={CAMPOS[m].nombre}
												value={atributos[m] ?? ''}
												oninput={(e) =>
													(atributos[m] = e.currentTarget.value.replace(/\D/g, '').slice(0, 8))}
											/>
										{/each}
									</div>
								</div>
							{:else}
								{@const f = CAMPOS[c]}
								<label class="campo"
									>{f.nombre}{#if f.numero}<span class="con-unidad"
											><input
												inputmode="numeric"
												placeholder={f.placeholder}
												value={atributos[c] ?? ''}
												oninput={(e) =>
													(atributos[c] = e.currentTarget.value.replace(/\D/g, '').slice(0, 8))}
											/><i>{f.unidad}</i></span
										>{:else}<input
											maxlength="60"
											placeholder={f.placeholder}
											bind:value={atributos[c]}
										/>{/if}</label
								>
							{/if}
						{/each}
					</div>
					{#if camposTipo.includes('peso')}<p class="t-sub">
							El peso y las medidas sirven para cotizar el envío.
						</p>{/if}
				</div>
			{/if}

			<div class="tarjeta">
				<div class="t">Variantes y stock</div>
				{#each opciones as o, i (i)}
					<div class="opcion">
						<span class="o-fila"
							><input
								class="o-n"
								aria-label="Opción {i + 1}"
								placeholder="Talle, color…"
								maxlength="20"
								value={o.nombre}
								onchange={(e) => renombrar(i, e.currentTarget.value)}
							/><button
								type="button"
								class="o-quitar"
								aria-label="Quitar la opción {o.nombre}"
								onclick={() => quitarOpcion(i)}><Icon icon={I.phXBold} size={12} /></button
							></span
						>
						<div class="valores">
							{#each o.valores as x (x)}<span class="valor"
									>{#if colorDe(o.nombre, x)}<span
											class="v-color chico"
											style="background:{colorDe(o.nombre, x)}"
										></span>{/if}{x}<button
										type="button"
										aria-label="Quitar {x}"
										onclick={() => quitarValor(i, x)}><Icon icon={I.phXBold} size={10} /></button
									></span
								>{/each}<input
								class="v-nuevo"
								aria-label="Sumar valores a {o.nombre || 'la opción'}"
								placeholder={o.valores.length ? 'Otro valor' : ejemploOpcion(o.nombre)}
								bind:value={nuevoValor[i]}
								onkeydown={(e) => {
									if (e.key === 'Enter' || e.key === ',') {
										e.preventDefault();
										sumarValores(i);
									}
								}}
								onblur={() => nuevoValor[i]?.trim() && sumarValores(i)}
							/>
						</div>
					</div>
				{/each}
				<div class="variantes">
					{#if listas.length}
						{#each variantes as v (nombreVariante(v))}
							<div class="variante" class:apagada={!v.activa} class:con-mayor={verMayorista}>
								<span class="v-color" style:background={colorVariante(v)}></span><span class="v-n"
									>{nombreVariante(v)}</span
								>
								<input
									class="v-p"
									inputmode="numeric"
									aria-label="Precio de {nombreVariante(v)}"
									readonly={!conPrecios}
									placeholder={pesos(precio)}
									value={v.precio === null ? '' : pesos(v.precio)}
									onchange={(e) =>
										(v.precio = e.currentTarget.value.trim()
											? numero(e.currentTarget.value)
											: null)}
								/>
								{#if verMayorista}
									<!-- Vacío: el mayorista del producto, si la variante va al precio del producto. -->
									<input
										class="v-p"
										inputmode="numeric"
										aria-label="Precio mayorista de {nombreVariante(v)}"
										readonly={!conPrecios}
										placeholder={precioMayor && v.precio === null
											? pesos(precioMayor)
											: 'Mayorista'}
										value={v.precioMayorista ? pesos(v.precioMayorista) : ''}
										onchange={(e) =>
											(v.precioMayorista = e.currentTarget.value.trim()
												? numero(e.currentTarget.value) || null
												: null)}
									/>
								{/if}
								<input
									class="v-s"
									inputmode="numeric"
									aria-label="Stock de {nombreVariante(v)}"
									value={v.stock}
									onchange={(e) => (v.stock = numero(e.currentTarget.value))}
								/><span class="v-u">u</span>
								<Interruptor label="Vender {nombreVariante(v)}" bind:checked={v.activa} />
							</div>
						{/each}
						<div class="v-total">
							{variantes.length}
							{variantes.length === 1 ? 'combinación' : 'combinaciones'} · {total} u en total
						</div>
					{:else}
						<div class="variante">
							<span class="v-color"></span><span class="v-n">Única</span>
							<input
								class="v-s"
								inputmode="numeric"
								aria-label="Stock"
								value={stock}
								onchange={(e) => (stock = numero(e.currentTarget.value))}
							/><span class="v-u">u</span>
						</div>
					{/if}
				</div>
				{#if opciones.length < MAX_OPCIONES}<button
						type="button"
						class="o-sumar"
						onclick={sumarOpcion}><Icon icon={I.phPlusBold} size={14} />{textoSumar}</button
					>{/if}
			</div>

			<LanzamientoProducto
				bind:this={lanzamiento}
				id={producto?.id ?? null}
				abierto={open}
				{precio}
				{conPrecios}
			/>

			<div class="tarjeta canales">
				<div class="t">Dónde se publica</div>
				{#each CANALES as [k, icon, t, col] (k)}
					<div class="canal">
						<Icon {icon} size={18} style="color:{col}" />
						<span>{t}</span>
						<Interruptor label={t} bind:checked={canales[k]} />
					</div>
				{/each}
			</div>

			<div class="tarjeta google">
				<div class="g-t">Así aparece en Google</div>
				<div class="g-u">{dominio} › {slug}</div>
				<div class="g-n">{nombre || 'Nombre del producto'} | {tienda}</div>
				<div class="g-d">{descripcion || 'La descripción aparece acá.'}</div>
			</div>
		{/if}
		{#if error}<div class="error" role="alert">{error}</div>{/if}
	</div>

	{#snippet pie()}
		<div class="pie" class:editando={!!producto}>
			{#if producto}
				<button
					type="button"
					class="icono"
					class:peligro={confirmando}
					onclick={borrar}
					disabled={guardando}
					aria-label={confirmando ? 'Sí, borrar el producto' : 'Borrar el producto'}
					title="Borrar"
					><Icon icon={I.phTrash} size={18} />{#if confirmando}<span>¿Lo borro?</span>{/if}</button
				>
				{#if !confirmando}
					<button
						type="button"
						class="icono"
						onclick={duplicar}
						disabled={guardando}
						aria-label="Duplicar el producto"
						title="Duplicar"><Icon icon={I.phCopyBold} size={18} /></button
					>
					<button
						type="button"
						class="icono"
						onclick={() =>
							producto && compartirProducto(ui, page.data.tienda?.slug ?? '', producto, guardar)}
						disabled={guardando}
						aria-label="Compartir el producto"
						title="Compartir"><Icon icon={I.phShareNetworkBold} size={18} /></button
					>
				{/if}
			{/if}
			<button type="button" class="cancelar" onclick={() => (open = false)}>Cancelar</button>
			{#if !eligiendoTipo}<button
					type="button"
					class="guardar"
					onclick={guardar}
					disabled={guardando}>Guardar y publicar</button
				>{/if}
		</div>
	{/snippet}
</Hoja>

<style>
	.cab {
		display: flex;
		align-items: center;
		gap: 12px;
	}
	.c-t {
		flex: 1;
		font-size: 20px;
		font-weight: 600;
		letter-spacing: -0.02em;
	}
	.redondo {
		width: 36px;
		height: 36px;
		border: 0;
		border-radius: 50%;
		background: var(--chip);
		color: var(--tx);
		display: grid;
		place-items: center;
	}
	.cuerpo {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}
	.fotos {
		display: grid;
		grid-template-columns: 2fr 1fr 1fr;
		gap: 8px;
	}
	.foto {
		position: relative;
	}
	.fotos img {
		display: block;
		width: 100%;
		height: 100%;
		aspect-ratio: 1;
		border-radius: 12px;
		object-fit: cover;
	}
	.fotos .f-1 {
		grid-row: span 2;
	}
	.fotos .f-1 img {
		border-radius: 16px;
	}
	.f-x,
	.f-portada {
		position: absolute;
		top: 6px;
		width: 24px;
		height: 24px;
		border: 0;
		border-radius: 50%;
		background: rgba(42, 20, 12, 0.6);
		color: #fff;
		display: grid;
		place-items: center;
	}
	.f-x {
		right: 6px;
	}
	.f-portada {
		left: 6px;
	}
	.f-vacia {
		aspect-ratio: auto;
		width: 100%;
		height: 100%;
		border: 1.5px dashed var(--ln2);
		background: var(--sf);
		color: var(--mu);
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 6px;
		font-size: 13px;
		font-weight: 600;
	}
	.f-mas {
		aspect-ratio: 1;
		border-radius: 12px;
		border: 1.5px dashed var(--ln2);
		background: transparent;
		display: grid;
		place-items: center;
		color: var(--mu);
	}
	.tarjeta {
		padding: 14px;
		border-radius: 18px;
		background: var(--sf);
		box-shadow: var(--sh);
	}
	.campos {
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.campo {
		display: flex;
		flex-direction: column;
		gap: 4px;
		font-size: 12.5px;
		font-weight: 600;
		color: var(--mu);
	}
	.campo input,
	.campo textarea {
		padding: 11px 12px;
		border-radius: 12px;
		border: 1px solid var(--ln2);
		background: var(--bg);
		font-size: 15px;
		font-weight: 500;
		color: var(--tx);
		outline: none;
	}
	.campo textarea {
		min-height: 72px;
		font-size: 14px;
		line-height: 1.5;
		resize: vertical;
	}
	.campo input:focus,
	.campo textarea:focus {
		border-color: #d2582f;
	}
	.c-fila {
		display: flex;
		justify-content: space-between;
	}
	.hornero {
		border: 0;
		background: transparent;
		padding: 0;
		font-size: 12.5px;
		font-weight: 600;
		color: #b0441d;
	}
	.t {
		font-size: 14px;
		font-weight: 600;
	}
	.t-sub {
		margin: 4px 0 0;
		font-size: 12.5px;
		color: var(--mu);
	}
	.tipos {
		margin-top: 12px;
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(124px, 1fr));
		gap: 8px;
	}
	.tipo {
		padding: 12px;
		border-radius: 14px;
		border: 1.5px solid var(--ln2);
		background: var(--bg);
		color: var(--tx);
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		gap: 4px;
		text-align: left;
	}
	.tipo :global(svg) {
		color: #d2582f;
	}
	.tipo.elegido {
		border-color: #d2582f;
		background: var(--sf3);
	}
	.tp-n {
		font-size: 14px;
		font-weight: 600;
	}
	.tp-e {
		font-size: 12px;
		color: var(--mu);
	}
	.tipo-actual {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 13.5px;
		color: var(--mu);
	}
	.tipo-actual :global(svg) {
		color: #d2582f;
	}
	.tipo-actual b {
		color: var(--tx);
		font-weight: 600;
	}
	.tipo-actual span {
		flex: 1;
	}
	.datos {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
		gap: 10px;
	}
	.campo.ancho {
		grid-column: 1 / -1;
	}
	.medidas input {
		min-width: 0;
	}
	.medidas {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 8px;
	}
	.con-unidad {
		position: relative;
		display: flex;
	}
	.con-unidad input {
		width: 100%;
		min-width: 0;
		padding-right: 34px;
	}
	.con-unidad i {
		position: absolute;
		right: 12px;
		top: 50%;
		transform: translateY(-50%);
		font-style: normal;
		font-size: 13px;
		color: var(--mu);
		pointer-events: none;
	}
	.precios {
		margin-top: 10px;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
		gap: 8px;
	}
	.precio {
		padding: 10px 12px;
		border-radius: 12px;
		background: var(--sf3);
	}
	.p-t {
		font-size: 12px;
		color: var(--mu);
	}
	.p-v,
	.p-in {
		font-size: 18px;
		font-weight: 600;
	}
	.p-in {
		width: 100%;
		border: 0;
		border-bottom: 1.5px dashed var(--ln2);
		background: transparent;
		color: var(--tx);
		padding: 0;
		outline: none;
	}
	.p-in:focus {
		border-bottom-color: #d2582f;
	}
	.p-n {
		font-size: 11.5px;
		color: var(--fa);
	}
	.p-n.mal {
		color: #b0441d;
	}
	.margen {
		margin-top: 10px;
		font-size: 12.5px;
		color: var(--mu);
		display: flex;
		align-items: center;
		gap: 4px;
		flex-wrap: wrap;
	}
	.margen b {
		color: #237550;
	}
	.margen b.bajo {
		color: #b0441d;
	}
	.costo {
		width: 110px;
		border: 0;
		border-bottom: 1.5px dashed var(--ln2);
		background: transparent;
		color: var(--tx);
		font-size: 12.5px;
		font-weight: 600;
		outline: none;
	}
	.variantes {
		margin-top: 8px;
	}
	.variante {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 9px 0;
		border-bottom: 1px solid var(--ln);
	}
	.v-color {
		width: 18px;
		height: 18px;
		border-radius: 50%;
		background: var(--sf3);
		box-shadow: inset 0 0 0 1px rgba(0, 0, 0, 0.12);
	}
	.v-color.chico {
		width: 12px;
		height: 12px;
	}
	.v-n {
		flex: 1;
		min-width: 0;
		font-size: 14px;
	}
	/* Con el precio mayorista, a 375 px los números bajan a su renglón (el nombre arriba). */
	@media (max-width: 760px) {
		.variante.con-mayor {
			flex-wrap: wrap;
		}
		.variante.con-mayor .v-n {
			flex-basis: calc(100% - 28px);
		}
	}
	.variante.apagada .v-n {
		color: var(--mu);
		text-decoration: line-through;
	}
	.v-p {
		width: 92px;
		text-align: right;
		border: 0;
		border-bottom: 1.5px dashed var(--ln2);
		background: transparent;
		color: var(--tx);
		font-size: 13px;
		font-weight: 600;
		outline: none;
	}
	.v-p::placeholder {
		color: var(--mu);
		font-weight: 500;
	}
	.v-total {
		padding-top: 9px;
		font-size: 12.5px;
		color: var(--mu);
	}
	.opcion {
		margin-top: 10px;
		padding: 10px 12px;
		border-radius: 12px;
		border: 1px solid var(--ln2);
		background: var(--bg);
	}
	.o-fila {
		display: flex;
		align-items: center;
		gap: 8px;
	}
	.o-n {
		flex: 1;
		min-width: 0;
		border: 0;
		background: transparent;
		color: var(--tx);
		font-size: 14px;
		font-weight: 600;
		outline: none;
	}
	.o-quitar {
		width: 26px;
		height: 26px;
		border: 0;
		border-radius: 50%;
		background: var(--chip);
		color: var(--mu);
		display: grid;
		place-items: center;
	}
	.valores {
		margin-top: 8px;
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 6px;
	}
	.valor {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		padding: 5px 6px 5px 10px;
		border-radius: 999px;
		background: var(--chip);
		font-size: 13px;
		font-weight: 600;
	}
	.valor button {
		width: 18px;
		height: 18px;
		border: 0;
		border-radius: 50%;
		background: transparent;
		color: var(--mu);
		display: grid;
		place-items: center;
	}
	.v-nuevo {
		flex: 1;
		min-width: 90px;
		padding: 5px 2px;
		border: 0;
		border-bottom: 1.5px dashed var(--ln2);
		background: transparent;
		color: var(--tx);
		font-size: 13px;
		outline: none;
	}
	.o-sumar {
		margin-top: 10px;
		width: 100%;
		padding: 10px;
		border-radius: 12px;
		border: 1.5px dashed var(--ln2);
		background: transparent;
		color: var(--mu);
		font-size: 13px;
		font-weight: 600;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 6px;
	}
	.v-s {
		width: 64px;
		text-align: right;
		border: 0;
		border-bottom: 1.5px dashed var(--ln2);
		background: transparent;
		color: var(--tx);
		font-size: 13px;
		font-weight: 600;
		outline: none;
	}
	.v-u {
		font-size: 13px;
		font-weight: 600;
		color: var(--mu);
	}
	.canales {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.canales .t {
		margin-bottom: 4px;
	}
	.canal {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 7px 0;
	}
	.canal span {
		flex: 1;
		font-size: 14px;
	}
	.g-t {
		font-size: 12px;
		color: var(--mu);
	}
	.g-u {
		margin-top: 6px;
		font-size: 12.5px;
		color: #237550;
		overflow-wrap: anywhere;
	}
	.g-n {
		font-size: 16px;
		color: #3e6bfd;
		font-weight: 500;
	}
	.g-d {
		font-size: 13px;
		color: var(--mu);
		line-height: 1.4;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}
	.error {
		font-size: 13px;
		font-weight: 600;
		color: #b83a2e;
	}
	.pie {
		width: 100%;
		display: flex;
		gap: 8px;
	}
	.icono {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		padding: 0 14px;
		min-height: 46px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: transparent;
		color: var(--mu);
		font-size: 14px;
		font-weight: 600;
	}
	.icono.peligro {
		border-color: #c0392b;
		background: #c0392b;
		color: #fff;
	}
	.cancelar {
		padding: 13px 18px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: transparent;
		color: var(--tx);
		font-size: 14px;
		font-weight: 600;
	}
	/* En el celular, editando, la cruz de arriba ya cierra: «Cancelar» deja lugar a borrar, duplicar
	   y compartir sin partir «Guardar y publicar». */
	@media (max-width: 760px) {
		.pie.editando .cancelar {
			display: none;
		}
	}
	.guardar {
		flex: 1;
		padding: 13px;
		border-radius: 999px;
		border: 0;
		background: linear-gradient(180deg, #c24e25, #b0441d);
		color: #fff;
		font-size: 14.5px;
		font-weight: 600;
	}
</style>
