<script lang="ts">
	import { onMount } from 'svelte';
	import CookieBanner from '#lib/components/CookieBanner.svelte';
	import Grano from '#lib/components/Grano.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import Header from '#lib/landing/Header.svelte';
	import CtaFinal from '#lib/landing/CtaFinal.svelte';
	import SitioFooter from '#lib/sitio/SitioFooter.svelte';
	import Migrar from '#lib/landing/Migrar.svelte';
	import PlantillaModal from '#lib/landing/PlantillaModal.svelte';
	import Precios from '#lib/landing/Precios.svelte';
	import Producto from '#lib/landing/Producto.svelte';
	import WidgetHornero from '#lib/landing/WidgetHornero.svelte';
	import Hero from '#lib/landing/Hero.svelte';
	import HorneroSeccion from '#lib/landing/HorneroSeccion.svelte';
	import Lenis from '#lib/landing/Lenis.svelte';
	import { CHAT } from '#lib/landing/datos.ts';
	import { Landing, setLanding } from '#lib/landing/estado.svelte.ts';

	/* Landing de Tu Local Web: recreación de la vista "Landing" de design/Mostrador v4.dc.html. */
	const l = setLanding(new Landing());
	onMount(() => {
		l.iniciar(CHAT.length);
		// La sección de plantillas ya no está: #plantillas (menú, footer, panel) abre la galería.
		const plantillas = () => {
			if (location.hash === '#plantillas') l.tplI = 0;
		};
		plantillas();
		addEventListener('hashchange', plantillas);
		return () => {
			l.detener();
			removeEventListener('hashchange', plantillas);
		};
	});
</script>

<Seo
	title="Tu Local Web · Vendé en todos lados, pagá menos"
	description="Tienda online, WhatsApp, Mercado Libre, local y factura de ARCA en un solo panel. Más barato que Tiendanube en cada plan."
	path="/"
	image="home"
	jsonLd={{
		'@context': 'https://schema.org',
		'@type': 'Organization',
		name: 'Tu Local Web',
		url: 'https://tulocalweb.com.ar',
		logo: 'https://tulocalweb.com.ar/logo.png'
	}}
/>

<Grano opacidad={0.3} />
<Lenis />

<div class="landing">
	<Header />
	<main>
		<Hero />
		<HorneroSeccion />
		<Producto />
		<Precios />
		<Migrar />
		<CtaFinal />
	</main>
	<SitioFooter cta={false} />
</div>

<PlantillaModal />
<WidgetHornero />

<CookieBanner />

<style>
	.landing {
		position: relative;
		overflow-x: clip;
		background: #fff;
		color: #2a140c;
	}
</style>
