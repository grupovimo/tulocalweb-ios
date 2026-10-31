import type * as v from 'valibot';
import { ClienteBff } from '../bff.ts';
import type { NuevaPregunta, NuevaResena, NuevaSena } from './esquemas.ts';

/** Las operaciones públicas en la API Go (con la IP del visitante: Go pone el tope). */
export class VidrieraApi extends ClienteBff {
	resena(slug: string, r: v.InferOutput<typeof NuevaResena>) {
		return this.pedir<void>('POST', `/v1/publico/tiendas/${slug}/resenas`, r);
	}
	pregunta(slug: string, q: v.InferOutput<typeof NuevaPregunta>) {
		return this.pedir<void>('POST', `/v1/publico/tiendas/${slug}/preguntas`, q);
	}
	/** «Avisame cuando salga» (§52). */
	aviso(slug: string, productoId: string, email: string) {
		return this.pedir<void>(
			'POST',
			`/v1/publico/tiendas/${slug}/productos/${encodeURIComponent(productoId)}/aviso`,
			{ email }
		);
	}
	/** Reservar o encargar con seña (§52): el pedido y adónde pagar la seña. */
	sena(slug: string, s: v.InferOutput<typeof NuevaSena>) {
		return this.pedir<{ numero: number; codigo: string; total: number; sena: number; url: string }>(
			'POST',
			`/v1/publico/tiendas/${slug}/senas`,
			{ ...s, varianteId: s.varianteId ?? '' }
		);
	}
}
