import { cuentasDe, datosDe, usaSandbox } from '#lib/server/datos/index.ts';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async (event) => {
	const datos = datosDe(event.locals);
	const [m, t] = await Promise.all([datos.misDatos(), datos.tienda()]);
	// El email confirma «Borrar mi cuenta». En la demo no hay cuenta: la opción no aparece.
	const token = event.locals.token;
	const email =
		usaSandbox() || !token
			? null
			: await cuentasDe(event)
					.sesion(token)
					.then((d) => d.usuario.email)
					.catch(() => null);
	return { ...m, tienda: t.nombre, email };
};
