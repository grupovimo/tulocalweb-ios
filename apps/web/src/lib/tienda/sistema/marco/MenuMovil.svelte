<script lang="ts">
	import MenuTienda from '../../MenuTienda.svelte';
	import { usarTienda } from '../contexto.ts';
	import { categoriaHref } from '../datos.ts';

	/**
	 * El menú del celular: el de todas las plantillas (MenuTienda: panel con la navegación, Inicio,
	 * el carrito y WhatsApp), con las pestañas del sistema. Toma el fondo y la letra del encabezado.
	 * Sólo se ve en el celular (base.css › .tv-solo-movil).
	 */
	const ctx = usarTienda();
	// Las categorías que no van como pestañas (más de 4), también en el celular.
	const nav = $derived.by(() => {
		if (ctx.nav.some((n) => n.href.includes('cat='))) return ctx.nav;
		const cats = ctx.tienda.categorias.filter((c) => c.cantidad > 0);
		if (cats.length < 2) return ctx.nav;
		const i = Math.max(0, ctx.nav.findIndex((n) => n.href.endsWith('/catalogo')) + 1);
		const extra = cats
			.slice(0, 12)
			.map((c) => ({ t: c.nombre, href: categoriaHref(ctx.tienda, c.nombre) }));
		return [...ctx.nav.slice(0, i), ...extra, ...ctx.nav.slice(i)];
	});
</script>

<span class="tv-solo-movil tv-menu-movil"
	><MenuTienda
		tienda={ctx.tienda}
		base={ctx.base}
		{nav}
		estilo="display:grid;place-items:center;width:44px;height:44px"
		tam={22}
	/></span
>
