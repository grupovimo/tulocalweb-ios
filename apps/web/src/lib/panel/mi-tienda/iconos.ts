/** Los íconos del editor: el de cada bloque en la lista y los de beneficios (bloques.json › iconos). */
import * as I from '#lib/icons.ts';
import type { Icono, TipoBloque } from '#lib/tienda/sistema/catalogo.ts';

type Ic = I.IconData;

export const ICONO_BLOQUE: Record<TipoBloque, Ic> = {
	portada: I.phImageFill,
	banner: I.phMegaphoneFill,
	oferta: I.phTimerFill,
	textoImagen: I.phArticleFill,
	video: I.phVideoCameraFill,
	beneficios: I.phSealCheckFill,
	productos: I.phSquaresFourFill,
	destacado: I.phStarFill,
	combo: I.phStackFill,
	categorias: I.phTagFill,
	menu: I.phForkKnifeFill,
	local: I.phMapPinFill,
	resenas: I.phChatCenteredTextFill,
	preguntas: I.phQuestionFill,
	logos: I.phCertificateFill,
	galeria: I.phImagesFill,
	clientes: I.phUsersThreeFill,
	videos: I.phFilmStripFill,
	newsletter: I.phEnvelopeSimpleFill,
	whatsapp: I.phWhatsappLogoFill,
	turnos: I.phCalendarCheckFill
};

export const ICONO_BENEFICIO: Record<Icono, [Ic, string]> = {
	envio: [I.phTruckFill, 'Envío'],
	cuotas: [I.phCreditCardFill, 'Cuotas'],
	transferencia: [I.phBankFill, 'Transferencia'],
	cambios: [I.phArrowsClockwiseFill, 'Cambios'],
	retiro: [I.phStorefrontFill, 'Retiro'],
	seguro: [I.phShieldCheckFill, 'Compra segura'],
	regalo: [I.phGiftFill, 'Regalo'],
	'hecho-a-mano': [I.phHandHeartFill, 'Hecho a mano'],
	reloj: [I.phClockFill, 'Horario'],
	estrella: [I.phStarFill, 'Destacado']
};
