<script lang="ts">
	import MenuTienda from '../../MenuTienda.svelte';
	import { usarTienda } from '../contexto.ts';

	/**
	 * El menú del celular: el de todas las plantillas (MenuTienda: panel con la navegación, Inicio,
	 * el carrito y WhatsApp), con las pestañas del sistema. Toma el fondo y la letra del encabezado.
	 * Sólo se ve en el celular (base.css › .tv-solo-movil).
	 */
	const ctx = usarTienda();
</script>

<span class="tv-solo-movil tv-menu-movil"
	><MenuTienda
		tienda={ctx.tienda}
		base={ctx.base}
		nav={ctx.nav}
		estilo="display:grid;place-items:center;width:44px;height:44px"
		tam={22}
	/></span
>
