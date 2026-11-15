import { rolesDe } from './roles.ts';
import type { Rol } from './tipos.ts';
import type { EstiloTienda } from './estilo.ts';

/**
 * Los colores de Mi tienda › Estilo que validan y se usan fuera del sistema de la tienda (el editor,
 * el logo). El CSS de la tienda lo arma lib/tienda/sistema/tokens.ts.
 */

const HEX = /^#[0-9A-F]{6}$/;

/** Un color elegido válido (#RRGGBB) o null. Lo que no valida no llega nunca al CSS. */
export function colorValido(c: string | undefined | null): string | null {
	const v = (c ?? '').trim().toUpperCase();
	return HEX.test(v) ? v : null;
}

/** El color que se ve de un rol (el elegido o el del punto de partida), para el editor y el logo. */
export function colorDe(
	plantilla: string,
	estilo: EstiloTienda | null | undefined,
	rol: Rol
): string | null {
	return colorValido(estilo?.colores?.[rol]) ?? rolesDe(plantilla)[rol] ?? null;
}

/** ¿Es oscuro? (para elegir el logo para fondos oscuros). */
export function esOscuro(hex: string | null): boolean {
	if (!hex) return false;
	const [r, g, b] = [1, 3, 5].map((i) => parseInt(hex.slice(i, i + 2), 16) / 255);
	return 0.2126 * r + 0.7152 * g + 0.0722 * b < 0.4;
}
