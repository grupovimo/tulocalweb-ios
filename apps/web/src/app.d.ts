declare global {
	namespace App {
		interface Locals {
			/** Sandbox de la demo del visitante (sólo en /prueba y /panel, para no leer D1 en la landing). */
			sandbox: import('#lib/server/sandbox/sandbox.ts').Sandbox | null;
			/** Token de la sesión de la API Go (DATA_SOURCE=api), leído de la cookie httpOnly. */
			token: string | null;
			/** Token de la sesión del equipo (/staff, DATA_SOURCE=api), leído de su cookie. */
			equipo: string | null;
			/** El panel abierto por alguien del equipo en sólo lectura ("ver la tienda"). */
			mirada: boolean;
			/** En el dominio propio de una tienda (mitienda.com.ar): el slug de esa tienda; si no, null. */
			dominio: string | null;
		}
		interface Error {
			message: string;
			code?: string;
		}
		interface Platform {
			/** El contexto del Worker (adapter-cloudflare): waitUntil para terminar después de responder. */
			ctx?: ExecutionContext;
		}
	}

	namespace Cloudflare {
		interface Env {
			PUBLIC_SITE_URL: string;
			/** Píxel de Meta de Tu Local Web (los anuncios, /l/). No es secreto. Vacío: no se mide. */
			META_PIXEL_ID?: string;
			/** Token de la API de conversiones de ese píxel (secreto): el alta se manda desde acá. */
			META_CAPI_TOKEN?: string;
			/** Código de «Probar eventos» (TEST…): mientras esté, los eventos van a la prueba. */
			META_CODIGO_PRUEBA?: string;
			/** sandbox: demo con datos por visitante (D1). api: la API Go del VPS. */
			DATA_SOURCE: 'sandbox' | 'api';
			/** URL de la API Go (DATA_SOURCE=api). */
			API_URL: string;
			/**
			 * Origen del WebSocket de tiempo real si no es el mismo sitio (sólo pruebas y desarrollo,
			 * p. ej. ws://localhost:4000). Vacío: el mismo sitio, que Caddy manda a la API.
			 */
			TIEMPO_REAL_ORIGEN?: string;
			/** Clave AES-256 en base64url (secreto). Cifra credenciales de integraciones. */
			ENCRYPTION_KEY: string;
			/** Clave que comparte el BFF con la API Go (X-Mostrador-BFF). Secreto. */
			CLAVE_BFF?: string;
			/** Secreto del timer tulocalweb-cron del VPS para /api/internal/*. */
			INTERNAL_SECRET: string;
			/** Workers AI (Tincho). */
			AI?: Ai;
			/** workers-ai en producción; simulado en desarrollo y pruebas (no llama al modelo). */
			HORNERO?: 'workers-ai' | 'simulado';
			/**
			 * Clave publicable de Clerk (pk_live_… / pk_test_…): de ella sale la Frontend API para
			 * seguir con Google o Apple. No es secreta. Vacía: no se muestran esos botones.
			 */
			CLERK_PUBLISHABLE_KEY?: string;
			/** App de socio de Tiendanube (importar con OAuth en la demo). Secretos de wrangler. */
			TIENDANUBE_CLIENT_ID?: string;
			TIENDANUBE_SECRETO?: string;
			/** Sólo en el preview de las pruebas: un Tiendanube falso (http://127.0.0.1:puerto). */
			TIENDANUBE_LOCAL?: string;
			/** Bot de Telegram de la plataforma (@mostradorhornero_bot). Secreto. */
			TELEGRAM_BOT_TOKEN?: string;
			/** Lo manda Telegram en X-Telegram-Bot-Api-Secret-Token. Secreto (lo generamos nosotros). */
			TELEGRAM_WEBHOOK_SECRET?: string;
		}
	}
}

export {};
