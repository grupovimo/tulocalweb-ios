<script lang="ts">
	import { browser } from '$app/env';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { getUi } from '#lib/ui.svelte.ts';
	import { usarTienda } from '../contexto.ts';
	import { anotarCupon } from '../promos-api.ts';

	/**
	 * `?cupon=CODIGO` en cualquier página de la tienda (el link del mail del premio de una ventana):
	 * el cupón queda anotado para el carrito y el checkout (se aplica solo, como el link /c/CODIGO),
	 * se avisa que quedó aplicado y el parámetro se saca de la dirección.
	 *
	 * Uso: lo pone marco/Tienda.svelte.
	 */
	const ctx = usarTienda();
	const ui = getUi();
	let hecho = '';

	$effect(() => {
		if (!browser) return;
		const crudo = page.url.searchParams.get('cupon');
		if (!crudo) return;
		const codigo = crudo.trim().toUpperCase();
		if (codigo === hecho) return;
		hecho = codigo;
		const u = new URL(location.href);
		u.searchParams.delete('cupon');
		try {
			// eslint-disable-next-line svelte/no-navigation-without-resolve -- la misma página, sin el parámetro
			void goto(u, { state: page.state, shallow: true, replace: true }).catch(() => {});
		} catch {
			// El router todavía no arrancó (la primera carga): igual se saca de la dirección.
			history.replaceState(history.state, '', u);
		}
		if (!/^[A-Z0-9-]{3,20}$/.test(codigo)) return;
		anotarCupon(ctx.tienda.slug, ctx.base, codigo);
		ui?.toast({
			k: 'ok',
			title: `Cupón ${codigo} aplicado`,
			desc: 'Se descuenta solo en el carrito al pagar.'
		});
	});
</script>
