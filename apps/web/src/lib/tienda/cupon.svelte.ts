/**
 * El cupón en el checkout de cualquier plantilla: pide la vista previa al servidor (que lo vuelve a
 * evaluar al comprar) y calcula los totales igual que el servidor.
 */
import type { ResultadoCupon } from './cupon.ts';

export class CuponCheckout {
	codigo = $state('');
	/** El código aplicado (el que va en el formulario) y su resultado. */
	aplicado = $state<string | null>(null);
	resultado = $state<ResultadoCupon | null>(null);
	cargando = $state(false);

	constructor(private slug: () => string) {}

	get ok() {
		return this.resultado?.ok ? this.resultado : null;
	}
	get motivo() {
		return this.resultado && !this.resultado.ok ? this.resultado.motivo : null;
	}

	async aplicar(lineas: { id: string; cantidad: number }[], email = '') {
		const codigo = this.codigo.trim().toUpperCase();
		if (!codigo) return;
		this.cargando = true;
		try {
			const r = await fetch(`/t/${this.slug()}/cupon`, {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify({ codigo, email, carrito: lineas })
			});
			this.resultado = r.ok
				? ((await r.json()) as ResultadoCupon)
				: { ok: false, motivo: 'No pudimos revisar el cupón. Probá de nuevo.' };
			this.aplicado = this.resultado.ok ? codigo : null;
		} finally {
			this.cargando = false;
		}
	}

	/** El código que dejó el link de un cupón (/t/[tienda]/c/[codigo]), si hay. */
	delLink(): string | null {
		const m = document.cookie.match(new RegExp(`(?:^|; )cupon-${this.slug()}=([A-Z0-9-]{3,20})`));
		return m ? m[1] : null;
	}

	quitar() {
		this.aplicado = null;
		this.resultado = null;
		this.codigo = '';
		// Si vino de un link, que no vuelva a aparecer.
		document.cookie = `cupon-${this.slug()}=; path=/t/${this.slug()}; max-age=0; samesite=lax`;
	}
}

/** Los totales del checkout, con la misma regla que el servidor (checkout/+page.server.ts y Go). */
export function totales(p: {
	subtotal: number;
	envio: number;
	pago: 'transferencia' | 'mercadopago';
	transferencia: number;
	cupon: CuponCheckout;
	/** El descuento de los combos completos (combos.ts). */
	combo?: number;
}) {
	const c = p.cupon.ok;
	const descCombo = p.combo ?? 0;
	const descCupon = c ? c.descuento : 0;
	const envio = c?.envioGratis ? 0 : p.envio;
	const transf =
		p.pago === 'transferencia' && (!c || c.combinable)
			? Math.round(((p.subtotal - descCombo - descCupon) * p.transferencia) / 100)
			: 0;
	return {
		descCombo,
		descCupon,
		envio,
		transf,
		total: p.subtotal + envio - descCombo - descCupon - transf
	};
}
