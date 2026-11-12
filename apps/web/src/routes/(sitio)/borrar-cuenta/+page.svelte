<script lang="ts">
	import Seo from '#lib/components/Seo.svelte';

	/**
	 * Borrar la cuenta (la pide Google Play como «URL de eliminación de cuenta»): cómo se pide, qué se
	 * borra, qué se guarda y por cuánto. La API: DELETE /v1/auth/cuenta (desactiva ya y borra los datos
	 * personales a los 30 días).
	 */
	const PASOS = [
		'En la app de Tu Local Web: Más › Cuenta › Borrar mi cuenta.',
		'En la web: entrá a tu panel, andá a Configuración › Mis datos y tocá «Borrar mi cuenta».',
		'Escribí tu email para confirmar. La cuenta se desactiva en el momento.',
		'Si no podés entrar, escribinos a hola@tulocalweb.com.ar desde el email de tu cuenta y la borramos nosotros.'
	];
</script>

<Seo
	title="Borrar tu cuenta · Tu Local Web"
	description="Cómo borrar tu cuenta de Tu Local Web, qué datos se borran y cuáles se guardan."
	path="/borrar-cuenta"
/>

<!-- eslint-disable svelte/no-navigation-without-resolve -- ruta pública fija -->
<div class="ancho">
	<article>
		<h1>Borrar tu cuenta de Tu Local Web</h1>
		<p class="bajada">
			Podés borrar tu cuenta cuando quieras, desde la app o desde la web. No hace falta escribirnos.
		</p>

		<h2>Cómo se pide</h2>
		<ol>
			{#each PASOS as p (p)}<li>{p}</li>{/each}
		</ol>

		<h2>Qué pasa después</h2>
		<ul>
			<li>
				<b>En el momento:</b> la cuenta se desactiva, se cierran tus sesiones y se cierran las tiendas
				donde sos la única dueña (dejan de estar online).
			</li>
			<li>
				<b>A los 30 días:</b> borramos tus datos personales: nombre, email, teléfono, foto, contraseña
				y los ingresos con Google o Apple.
			</li>
			<li>
				<b>Durante esos 30 días</b> la podés recuperar escribiendo a hola@tulocalweb.com.ar. Después no
				se puede deshacer.
			</li>
		</ul>

		<h2>Qué se guarda</h2>
		<p>
			Las ventas, los pedidos y las facturas de las tiendas quedan, sin tus datos personales, porque
			la ley obliga a conservar los registros contables y fiscales (en Argentina, hasta 10 años).
			Las copias de seguridad diarias se renuevan solas en 14 días.
		</p>
		<p>
			Si querés llevarte tus datos antes, exportalos desde Configuración › Mis datos. Más detalle en
			la
			<a href="/legales/privacidad">política de privacidad</a>.
		</p>
	</article>
</div>

<style>
	article {
		max-width: 720px;
		margin: 0 auto;
		padding: 56px 0 96px;
		line-height: 1.6;
	}
	h1 {
		font-size: clamp(30px, 5vw, 42px);
		line-height: 1.1;
		letter-spacing: -0.02em;
		margin: 0 0 12px;
		text-wrap: balance;
	}
	.bajada {
		font-size: 18px;
		opacity: 0.8;
		margin: 0 0 32px;
	}
	h2 {
		font-size: 22px;
		margin: 32px 0 10px;
	}
	ol,
	ul {
		padding-left: 22px;
		display: grid;
		gap: 8px;
	}
	p,
	li {
		text-wrap: pretty;
	}
	a {
		color: var(--terracota);
	}
</style>
