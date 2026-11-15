<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import CookieBanner from '#lib/components/CookieBanner.svelte';
	import Grano from '#lib/components/Grano.svelte';
	import Icon from '#lib/components/Icon.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import { anim } from '#lib/anim.ts';
	import * as I from '#lib/icons.ts';
	import type { Angulo } from './angulos.ts';
	import EscenaAvisos from './EscenaAvisos.svelte';
	import EscenaCaja from './EscenaCaja.svelte';
	import EscenaPasate from './EscenaPasate.svelte';
	import { evento, iniciarPixel } from './pixel.ts';

	/**
	 * La página de un anuncio (/l/<ángulo>): el título del anuncio, la escena, una sola acción (crear
	 * la cuenta), tres pruebas, tres preguntas y el cierre. Sin menú ni footer largo, liviana para el
	 * navegador de Instagram. Los lineamientos: tengolibre/docs/lineamientos-landing.md.
	 */
	let { a, pixel }: { a: Angulo; pixel: string | null } = $props();

	// El onboarding es el registro; el origen ya quedó guardado al llegar a esta página.
	const registro = resolve('/onboarding');
	let heroCta: HTMLElement | undefined = $state();
	let fija = $state(false);

	const lead = () => evento('Lead', { content_name: a.id });

	onMount(() => {
		const parar = iniciarPixel(pixel);
		// La barra con el botón aparece cuando el del hero sale de pantalla.
		const io = new IntersectionObserver(
			([e]) => (fija = !e.isIntersecting && e.boundingClientRect.top < 0)
		);
		if (heroCta) io.observe(heroCta);
		// ViewContent: leyó hasta la mitad.
		let visto = false;
		const scroll = () => {
			const h = document.documentElement;
			if (!visto && h.scrollTop + innerHeight >= h.scrollHeight / 2) {
				visto = true;
				evento('ViewContent', { content_name: a.id });
			}
		};
		addEventListener('scroll', scroll, { passive: true });
		return () => {
			parar();
			io.disconnect();
			removeEventListener('scroll', scroll);
		};
	});
</script>

<Seo title={a.titulo} description={a.descripcion} path="/l/{a.id}" image="home" noindex />
<Grano opacidad={0.22} />

<div class="pagina">
	<header>
		<a class="logo" href={resolve('/')} aria-label="Tu Local Web"
			><img src="/wordmark-brown.webp" alt="Tu Local Web" width="150" height="32" /></a
		>
		<a class="mini" href={registro} onclick={lead}>Empezar gratis</a>
	</header>

	<main>
		<section class="hero">
			<div class="texto">
				<span class="etiqueta" use:anim={'rise'}>{a.etiqueta}</span>
				<h1 use:anim={{ t: 'rise', delay: 80 }}>{a.h1[0]} <span>{a.h1[1]}</span></h1>
				<p class="bajada" use:anim={{ t: 'rise', delay: 160 }}>{a.bajada}</p>
				<div class="cta" bind:this={heroCta} use:anim={{ t: 'rise', delay: 240 }}>
					<a class="btn-inflado btn-grande" href={registro} onclick={lead}
						>Crear mi tienda gratis <Icon icon={I.phArrowRightBold} size={16} /></a
					>
					<p class="nota">{a.nota}</p>
				</div>
			</div>
			<div class="escena" use:anim={{ t: 'fade', delay: 120 }}>
				{#if a.escena === 'avisos'}
					<EscenaAvisos />
				{:else if a.escena === 'caja'}
					<EscenaCaja />
				{:else}
					<EscenaPasate />
				{/if}
			</div>
		</section>

		<section class="pruebas">
			{#each a.pruebas as p, i (p.t)}
				<div class="prueba" use:anim={{ t: 'reveal', delay: i * 90 }}>
					<span class="n">0{i + 1}</span>
					<h2>{p.t}</h2>
					<p>{p.d}</p>
				</div>
			{/each}
		</section>

		<section class="preguntas">
			{#each a.preguntas as q (q.p)}
				<details>
					<summary>{q.p}<Icon icon={I.phCaretDownBold} size={16} /></summary>
					<p>{q.r}</p>
				</details>
			{/each}
		</section>

		<section class="cierre">
			<h2 use:anim={'reveal'}>{a.cierre}</h2>
			<a class="btn-inflado btn-grande" href={registro} onclick={lead}
				>Crear mi tienda gratis <Icon icon={I.phArrowRightBold} size={16} /></a
			>
			<p>{a.nota}</p>
		</section>
	</main>

	<footer>
		<!-- eslint-disable svelte/no-navigation-without-resolve -- páginas fuera del router y un link externo -->
		<a href="/legales/terminos">Términos</a>
		<a href="/legales/privacidad">Privacidad</a>
		<a href="/legales/cookies">Cookies</a>
		<a href="https://wa.me/542235926772" rel="noopener" target="_blank">WhatsApp</a>
		<!-- eslint-enable svelte/no-navigation-without-resolve -->
	</footer>
</div>

<div class="barra" class:fija aria-hidden={!fija}>
	<a class="btn-inflado" href={registro} onclick={lead} tabindex={fija ? 0 : -1}
		>Crear mi tienda gratis <Icon icon={I.phArrowRightBold} size={16} /></a
	>
</div>

<CookieBanner />

<style>
	.pagina {
		background: #fbf7f2;
		color: #2a140c;
		min-height: 100vh;
		overflow-x: clip;
	}
	header {
		max-width: 1200px;
		margin: 0 auto;
		padding: 18px 24px;
		display: flex;
		align-items: center;
		justify-content: space-between;
	}
	.logo img {
		display: block;
		height: 22px;
		width: auto;
	}
	.mini {
		white-space: nowrap;
		font-size: 15px;
		font-weight: 600;
		color: #d2582f;
		padding: 12px 4px;
	}
	.hero {
		max-width: 1200px;
		margin: 0 auto;
		padding: 24px 24px 40px;
		display: grid;
		grid-template-columns: 1.1fr 0.9fr;
		align-items: center;
		gap: 48px;
	}
	.etiqueta {
		display: inline-block;
		padding: 6px 14px;
		border-radius: 999px;
		background: #fcede5;
		color: #a8401c;
		font-size: 13.5px;
		font-weight: 600;
	}
	h1 {
		text-wrap: balance;
		margin: 20px 0 0;
		font-size: clamp(40px, 5.4vw, 72px);
		line-height: 1;
		letter-spacing: -0.045em;
		font-weight: 600;
	}
	h1 span {
		display: block;
		color: #d2582f;
	}
	.bajada {
		text-wrap: pretty;
		margin: 22px 0 0;
		max-width: 520px;
		font-size: 19px;
		line-height: 1.5;
		color: #5c4536;
	}
	.cta {
		margin-top: 30px;
	}
	.nota {
		text-wrap: pretty;
		margin: 14px 0 0;
		max-width: 440px;
		font-size: 14px;
		color: #7a6252;
	}
	.escena {
		display: flex;
		justify-content: center;
	}
	.pruebas {
		max-width: 1200px;
		margin: 0 auto;
		padding: clamp(72px, 10vw, 140px) 24px clamp(48px, 7vw, 96px);
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 40px;
	}
	.n {
		font-size: 14px;
		font-weight: 600;
		color: #d2582f;
	}
	.prueba h2 {
		text-wrap: balance;
		margin: 10px 0 0;
		font-size: 26px;
		letter-spacing: -0.03em;
		font-weight: 600;
		line-height: 1.1;
	}
	.prueba p {
		text-wrap: pretty;
		margin: 10px 0 0;
		font-size: 16.5px;
		line-height: 1.5;
		color: #5c4536;
	}
	.preguntas {
		max-width: 760px;
		margin: 0 auto;
		padding: 0 24px clamp(72px, 10vw, 140px);
	}
	details {
		border-bottom: 1px solid rgba(58, 29, 18, 0.12);
	}
	summary {
		list-style: none;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 16px;
		min-height: 44px;
		padding: 18px 0;
		font-size: 18px;
		font-weight: 600;
	}
	summary::-webkit-details-marker {
		display: none;
	}
	details[open] summary :global(svg) {
		transform: rotate(180deg);
	}
	details p {
		text-wrap: pretty;
		margin: 0 0 20px;
		font-size: 16.5px;
		line-height: 1.5;
		color: #5c4536;
	}
	.cierre {
		margin: 0 16px;
		padding: clamp(64px, 9vw, 120px) 24px;
		border-radius: 32px;
		background: #2a140c;
		color: #fbf7f2;
		text-align: center;
	}
	.cierre h2 {
		text-wrap: balance;
		margin: 0 auto 30px;
		max-width: 720px;
		font-size: clamp(34px, 4.6vw, 60px);
		line-height: 1.02;
		letter-spacing: -0.04em;
		font-weight: 600;
	}
	.cierre p {
		text-wrap: pretty;
		margin: 16px auto 0;
		max-width: 420px;
		font-size: 14px;
		color: #d9c8ba;
	}
	footer {
		display: flex;
		flex-wrap: wrap;
		justify-content: center;
		gap: 4px 22px;
		padding: 32px 24px 110px;
		font-size: 14px;
	}
	footer a {
		color: #7a6252;
		padding: 12px 0;
	}
	.barra {
		position: fixed;
		z-index: 40;
		left: 0;
		right: 0;
		bottom: 0;
		padding: 12px 16px calc(12px + env(safe-area-inset-bottom));
		background: rgba(251, 247, 242, 0.92);
		backdrop-filter: blur(12px);
		box-shadow: 0 -10px 30px -20px rgba(42, 20, 12, 0.5);
		transform: translateY(110%);
		transition: transform 0.35s cubic-bezier(0.2, 0.7, 0.1, 1);
		display: none;
	}
	.barra a {
		width: 100%;
		justify-content: center;
	}
	.barra.fija {
		transform: none;
	}
	@media (max-width: 899px) {
		.hero {
			grid-template-columns: 1fr;
			gap: 40px;
			padding-top: 8px;
		}
		.pruebas {
			grid-template-columns: 1fr;
			gap: 36px;
		}
		.barra {
			display: block;
		}
		.cta :global(.btn-inflado) {
			width: 100%;
			justify-content: center;
		}
	}
	@media (prefers-reduced-motion: reduce) {
		.barra {
			transition: none;
		}
	}
</style>
