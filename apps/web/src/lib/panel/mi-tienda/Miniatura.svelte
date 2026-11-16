<script lang="ts" module>
	/** El dibujo de cada variante (docs/handoff/tienda-bloques/Bloques.dc.html). */
	type Forma =
		| 'foto'
		| 'partido'
		| 'partidoIzq'
		| 'collage'
		| 'texto'
		| 'carrusel'
		| 'banda'
		| 'contador'
		| 'barra'
		| 'dos'
		| 'iconos'
		| 'grilla'
		| 'look'
		| 'lista'
		| 'mas'
		| 'tarjetas'
		| 'logos'
		| 'input'
		| 'circulos'
		| 'video'
		| 'videoTexto'
		| 'mapa'
		| 'cita';

	const FORMAS: Record<string, Forma | [Forma, number, number]> = {
		'portada.foto': 'foto',
		'portada.partida': 'partido',
		'portada.collage': 'collage',
		'portada.texto': 'texto',
		'portada.carrusel': 'carrusel',
		'banner.banda': 'banda',
		'banner.foto': 'foto',
		'banner.doble': 'dos',
		'oferta.productos': 'contador',
		'oferta.banda': 'banda',
		'textoImagen.derecha': 'partido',
		'textoImagen.izquierda': 'partidoIzq',
		'textoImagen.fondo': 'foto',
		'textoImagen.texto': 'texto',
		'video.ancho': 'video',
		'video.conTexto': 'videoTexto',
		'beneficios.fila': 'iconos',
		'beneficios.tarjetas': 'tarjetas',
		'productos.grilla': ['grilla', 4, 4],
		'productos.carrusel': 'carrusel',
		'productos.lista': 'lista',
		'destacado.partido': 'partido',
		'destacado.foto': 'foto',
		'combo.tarjeta': 'look',
		'categorias.tarjetas': ['grilla', 3, 3],
		'categorias.circulos': 'circulos',
		'categorias.lista': 'lista',
		'menu.lista': 'mas',
		'menu.columnas': ['grilla', 2, 4],
		'local.mapa': 'mapa',
		'local.tarjeta': 'tarjetas',
		'resenas.tarjetas': 'tarjetas',
		'resenas.carrusel': 'carrusel',
		'resenas.cita': 'cita',
		'preguntas.acordeon': 'mas',
		'logos.fila': 'logos',
		'galeria.fila': ['grilla', 6, 6],
		'galeria.mosaico': ['grilla', 3, 6],
		'clientes.grilla': ['grilla', 4, 4],
		'clientes.carrusel': 'carrusel',
		'videos.fila': ['grilla', 5, 5],
		'newsletter.centrado': 'input',
		'newsletter.banda': 'banda',
		'whatsapp.banda': 'banda',
		'whatsapp.tarjeta': 'input',
		'turnos.lista': 'lista',
		'barra.fija': 'barra',
		'barra.rotativa': 'barra',
		// Las tres direcciones (Mi tienda › Estilo).
		'direccion.vidriera': 'partido',
		'direccion.mostrador': ['grilla', 4, 4],
		'direccion.pedido': 'mas'
	};
</script>

<script lang="ts">
	/**
	 * Miniatura de un bloque o de una variante, como las de la biblioteca del handoff: rayas para el
	 * texto, gris para las fotos y el color de la marca donde va un botón.
	 */
	let { tipo, variante, alto = 84 }: { tipo: string; variante: string; alto?: number } = $props();
	const f = $derived(FORMAS[`${tipo}.${variante}`] ?? 'texto');
	const forma = $derived(Array.isArray(f) ? f[0] : f);
	const cols = $derived(Array.isArray(f) ? f[1] : 0);
	const celdas = $derived(Array.isArray(f) ? f[2] : 0);
</script>

<div class="mi" style:height="{alto}px" aria-hidden="true">
	{#if forma === 'foto'}
		<div class="im lleno"></div>
		<div class="sobre">
			<div class="ln blanco gruesa"></div>
			<div class="ln blanco" style:width="70%"></div>
			<div class="bt blanco"></div>
		</div>
	{:else if forma === 'partido' || forma === 'partidoIzq'}
		{#if forma === 'partidoIzq'}<div class="im" style:flex="1"></div>{/if}
		<div class="col centro">
			<div class="ln gruesa"></div>
			<div class="ln" style:width="70%"></div>
			<div class="ls" style:width="85%"></div>
			<div class="bt"></div>
		</div>
		{#if forma === 'partido'}<div class="im" style:flex="1"></div>{/if}
	{:else if forma === 'collage'}
		<div class="oscuro col fin" style:flex="1.2">
			<div class="ln blanco"></div>
			<div class="bt acento"></div>
		</div>
		<div class="im" style:flex="1"></div>
		<div class="col" style:flex=".8">
			<div class="im" style:flex="1"></div>
			<div class="acento-bg" style:flex="1"></div>
		</div>
	{:else if forma === 'texto' || forma === 'cita'}
		<div class="col centro medio">
			{#if forma === 'cita'}<span class="comilla">“</span>{/if}
			<div class="ln gruesa" style:width="80%"></div>
			<div class="ln gruesa" style:width="55%"></div>
			<div class="ls" style:width="60%"></div>
			{#if forma === 'texto'}<div class="bt"></div>{/if}
		</div>
	{:else if forma === 'carrusel'}
		<div class="im" style:flex="1"></div>
		<span class="flecha izq">‹</span><span class="flecha der">›</span>
		<div class="puntos"><span class="activo"></span><span></span><span></span></div>
	{:else if forma === 'banda'}
		<div class="oscuro fila-banda">
			<div class="col" style:width="55%">
				<div class="ln blanco gruesa"></div>
				<div class="ls tenue"></div>
			</div>
			<div class="bt acento"></div>
		</div>
	{:else if forma === 'contador'}
		<div class="col" style:flex="1">
			<div class="fila">
				<div class="ln" style:width="40%"></div>
				<span class="dig"></span><span class="dig"></span><span class="dig"></span>
			</div>
			<div class="grilla" style:grid-template-columns="repeat(4,1fr)" style:flex="1">
				<div class="im"></div>
				<div class="im"></div>
				<div class="im"></div>
				<div class="im"></div>
			</div>
		</div>
	{:else if forma === 'barra'}
		<div class="col" style:flex="1">
			<div class="oscuro barrita"><div class="ls blanco" style:width="50%"></div></div>
			<div class="linea-cab"></div>
			<div class="im tenue-bg" style:flex="1"></div>
		</div>
	{:else if forma === 'dos'}
		<div class="im" style:flex="1"></div>
		<div class="oscuro" style:flex="1"></div>
	{:else if forma === 'iconos'}
		<div class="grilla medio" style:grid-template-columns="repeat(4,1fr)" style:flex="1">
			{#each [0, 1, 2, 3] as i (i)}
				<div class="col centro medio">
					<span class="ico"></span>
					<div class="ls" style:width="80%"></div>
				</div>
			{/each}
		</div>
	{:else if forma === 'grilla'}
		<div class="grilla" style:grid-template-columns="repeat({cols},1fr)" style:flex="1">
			{#each Array.from({ length: celdas }, (_, i) => i) as i (i)}
				<div class="col">
					<div class="im" style:flex="1"></div>
					<div class="ls"></div>
					{#if cols < 6}<div class="ln fina" style:width="50%"></div>{/if}
				</div>
			{/each}
		</div>
	{:else if forma === 'look'}
		<div class="grilla" style:grid-template-columns="repeat(4,1fr)" style:flex="1">
			<div class="im"></div>
			<div class="im"></div>
			<div class="im"></div>
			<div class="marco col fin">
				<div class="ln" style:width="70%"></div>
				<div class="bt lleno-ancho"></div>
			</div>
		</div>
	{:else if forma === 'lista' || forma === 'mas'}
		<div class="col" style:flex="1">
			{#each [55, 45, 65] as w (w)}
				<div class="renglon" class:con-linea={forma === 'mas'}>
					<div class="col" style:flex="1">
						<div class="ln" style:width="{w}%"></div>
						{#if forma === 'lista'}<div class="ls" style:width="{w + 20}%"></div>{/if}
					</div>
					{#if forma === 'lista'}<div class="im cuadradito"></div>{:else}<span class="mas">+</span
						>{/if}
				</div>
			{/each}
		</div>
	{:else if forma === 'tarjetas'}
		<div class="grilla" style:grid-template-columns="repeat(3,1fr)" style:flex="1">
			{#each [0, 1, 2] as i (i)}
				<div class="marco fino col">
					<div class="ls acento-bg" style:width="50%"></div>
					<div class="ln fina"></div>
					<div class="ls"></div>
					<div class="ls" style:width="60%"></div>
				</div>
			{/each}
		</div>
	{:else if forma === 'logos'}
		<div class="col centro medio" style:flex="1">
			<div class="ls" style:width="40%"></div>
			<div class="fila logos">
				<span style:width="22%"></span><span style:width="16%"></span><span style:width="24%"
				></span><span style:width="18%"></span>
			</div>
		</div>
	{:else if forma === 'input'}
		<div class="col centro medio" style:flex="1">
			<div class="ln gruesa" style:width="60%"></div>
			<div class="ls" style:width="45%"></div>
			<div class="fila" style:width="70%">
				<div class="campo"></div>
				<div class="bt corto"></div>
			</div>
		</div>
	{:else if forma === 'circulos'}
		<div class="fila repartida" style:flex="1">
			{#each [0, 1, 2, 3] as i (i)}
				<div class="col centro medio">
					<span class="im circulo"></span>
					<div class="ls" style:width="70%"></div>
				</div>
			{/each}
		</div>
	{:else if forma === 'video' || forma === 'videoTexto'}
		<div class="im play-fondo" style:flex="1"><span class="play">▶</span></div>
		{#if forma === 'videoTexto'}
			<div class="col centro">
				<div class="ln gruesa"></div>
				<div class="ls" style:width="85%"></div>
				<div class="ls" style:width="65%"></div>
			</div>
		{/if}
	{:else if forma === 'mapa'}
		<div class="col centro">
			<div class="ln gruesa"></div>
			<div class="ls" style:width="80%"></div>
			<div class="ls" style:width="60%"></div>
			<div class="bt"></div>
		</div>
		<div class="im mapa" style:flex="1.3"><span class="pin"></span></div>
	{/if}
</div>

<style>
	/* Los grises del handoff, derivados del texto del panel para que se lean en el tema oscuro. */
	.mi {
		--m-im: color-mix(in srgb, var(--tx) 17%, transparent);
		--m-ln: color-mix(in srgb, var(--tx) 78%, transparent);
		--m-ls: color-mix(in srgb, var(--tx) 30%, transparent);
		--osc: color-mix(in srgb, var(--tx) 80%, var(--sf));
		--ac: #e7b04a;
		width: 100%;
		border-radius: 8px;
		background: var(--sf);
		border: 1px solid var(--ln2, rgba(0, 0, 0, 0.1));
		padding: 8px;
		display: flex;
		gap: 5px;
		overflow: hidden;
		position: relative;
		box-sizing: border-box;
	}
	.mi * {
		box-sizing: border-box;
	}
	.im {
		background: var(--m-im);
		border-radius: 4px;
		min-height: 6px;
	}
	.lleno {
		position: absolute;
		inset: 8px;
	}
	.sobre {
		position: absolute;
		left: 16px;
		bottom: 16px;
		width: 50%;
		display: flex;
		flex-direction: column;
		gap: 4px;
	}
	.ln {
		height: 5px;
		border-radius: 3px;
		background: var(--m-ln);
		width: 100%;
	}
	.ln.gruesa {
		height: 7px;
	}
	.ln.fina {
		height: 4px;
	}
	.ls {
		height: 4px;
		border-radius: 3px;
		background: var(--m-ls);
		width: 100%;
	}
	.bt {
		height: 10px;
		width: 34px;
		border-radius: 3px;
		background: var(--m-ln);
		margin-top: 2px;
	}
	.bt.corto {
		width: 26px;
		margin: 0;
	}
	.bt.lleno-ancho {
		width: 100%;
		height: 7px;
	}
	.blanco {
		background: #fff !important;
	}
	.acento,
	.acento-bg {
		background: var(--ac) !important;
	}
	.acento-bg {
		border-radius: 4px;
	}
	.tenue {
		background: color-mix(in srgb, #fff 45%, transparent) !important;
	}
	.tenue-bg {
		opacity: 0.5;
	}
	.oscuro {
		background: var(--osc);
		border-radius: 5px;
		padding: 7px;
	}
	.col {
		display: flex;
		flex-direction: column;
		gap: 4px;
		min-width: 0;
	}
	.col.centro {
		flex: 1;
		justify-content: center;
	}
	.col.medio {
		align-items: center;
	}
	.col.fin {
		justify-content: flex-end;
	}
	.fila {
		display: flex;
		align-items: center;
		gap: 4px;
	}
	.fila.repartida {
		justify-content: space-around;
	}
	.fila-banda {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 10px;
	}
	.grilla {
		display: grid;
		gap: 4px;
	}
	.grilla.medio {
		align-items: center;
	}
	.dig {
		width: 12px;
		height: 11px;
		border-radius: 3px;
		background: var(--osc);
	}
	.dig:first-of-type {
		margin-left: auto;
	}
	.barrita {
		height: 15px;
		padding: 0;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 3px;
	}
	.linea-cab {
		height: 12px;
		border-bottom: 1px solid var(--m-ls);
	}
	.ico {
		width: 16px;
		height: 16px;
		border-radius: 5px;
		border: 2px solid var(--m-ln);
	}
	.marco {
		border: 2px solid var(--m-ln);
		border-radius: 4px;
		padding: 4px;
	}
	.marco.fino {
		border-width: 1px;
		border-color: var(--m-ls);
	}
	.renglon {
		display: flex;
		align-items: center;
		gap: 5px;
		flex: 1;
	}
	.renglon.con-linea {
		border-bottom: 1px solid var(--m-ls);
	}
	.cuadradito {
		width: 20px;
		height: 20px;
		flex: none;
	}
	.mas {
		font-size: 10px;
		color: var(--m-ln);
	}
	.logos span {
		height: 10px;
		border-radius: 2px;
		background: var(--m-ls);
	}
	.logos {
		width: 100%;
		justify-content: center;
		gap: 8px;
	}
	.campo {
		flex: 1;
		height: 11px;
		border: 1px solid var(--m-ls);
		border-radius: 3px;
	}
	.circulo {
		width: 30px;
		height: 30px;
		border-radius: 50%;
	}
	.flecha {
		position: absolute;
		top: 40%;
		font-size: 13px;
		color: var(--m-ln);
	}
	.flecha.izq {
		left: 12px;
	}
	.flecha.der {
		right: 12px;
	}
	.puntos {
		position: absolute;
		bottom: 13px;
		left: 0;
		right: 0;
		display: flex;
		justify-content: center;
		gap: 3px;
	}
	.puntos span {
		width: 4px;
		height: 4px;
		border-radius: 2px;
		background: var(--m-ls);
	}
	.puntos .activo {
		width: 12px;
		background: var(--m-ln);
	}
	.play-fondo {
		display: grid;
		place-items: center;
	}
	.play {
		width: 24px;
		height: 24px;
		border-radius: 50%;
		background: #fff;
		color: #333;
		display: grid;
		place-items: center;
		font-size: 10px;
	}
	.comilla {
		font:
			700 22px/0.6 Georgia,
			serif;
		color: var(--m-ln);
	}
	.mapa {
		position: relative;
		background-image:
			linear-gradient(var(--sf) 1px, transparent 1px),
			linear-gradient(90deg, var(--sf) 1px, transparent 1px);
		background-size: 12px 12px;
	}
	.pin {
		position: absolute;
		left: 50%;
		top: 40%;
		width: 10px;
		height: 10px;
		border-radius: 50% 50% 50% 0;
		background: var(--ac);
		transform: translate(-50%, -50%) rotate(-45deg);
	}
</style>
