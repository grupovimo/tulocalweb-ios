/**
 * El bloque de newsletter del pie de la tienda (§50), sin marcado: cada plantilla lo dibuja con las
 * piezas de su diseño y usa este estado. Queda pendiente hasta confirmar por mail; en la demo no sale
 * el mail y el aviso trae el link de confirmar.
 */
export class AltaNewsletter {
	email = $state('');
	enviando = $state(false);
	/** Lo que se le dice a la persona después de enviar (o el error). */
	aviso = $state<{ ok: boolean; texto: string; link?: string } | null>(null);

	/** slug: un getter (la tienda es una prop de la plantilla). */
	constructor(private slug: () => string) {}

	enviar = async (e: SubmitEvent) => {
		e.preventDefault();
		if (this.enviando) return;
		const email = this.email.trim();
		if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
			this.aviso = { ok: false, texto: 'Revisá el email.' };
			return;
		}
		this.enviando = true;
		try {
			const r = await fetch(`/t/${this.slug()}/newsletter`, {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify({ email })
			});
			const j = (await r.json().catch(() => null)) as {
				error?: string;
				demo?: { link: string | null } | null;
			} | null;
			if (!r.ok) {
				this.aviso = { ok: false, texto: j?.error ?? 'No pudimos sumarte. Probá de nuevo.' };
				return;
			}
			this.email = '';
			if (j?.demo)
				this.aviso = j.demo.link
					? {
							ok: true,
							texto: 'En la demo no mandamos el mail de confirmación:',
							link: j.demo.link
						}
					: { ok: true, texto: 'Ya estabas suscripto.' };
			else
				this.aviso = {
					ok: true,
					texto: 'Te mandamos un mail: tocá el link para confirmar tu suscripción.'
				};
		} catch {
			this.aviso = { ok: false, texto: 'No pudimos sumarte. Probá de nuevo.' };
		} finally {
			this.enviando = false;
		}
	};
}
