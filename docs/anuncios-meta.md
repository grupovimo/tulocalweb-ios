# Anuncios de Meta · páginas de aterrizaje

Una página por ángulo de anuncio, en `tulocalweb.com.ar/l/<ángulo>`. Cada una sigue lo que dijo el
anuncio y tiene una sola acción, crear la cuenta (va a `/onboarding`). No se indexan.

| Página | Ángulo |
|---|---|
| `/l/avisos` | Te enterás de cada venta |
| `/l/una-caja` | El local y la web, con un solo stock (plan Pro) |
| `/l/pasate` | Te traemos tu tienda gratis |

Los textos están en `apps/web/src/lib/anuncios/angulos.ts` y la plantilla en `Aterrizaje.svelte`.
Para sumar un ángulo, agregás una entrada y, si hace falta, una escena. Las promesas tienen que
salir de lo que el producto ya hace (`~/Trabajo/vimo/marketing/tulocalweb/competencia.md`). La
factura de ARCA no se anuncia hasta que ARCA acepte un certificado real.

## La URL del anuncio

```
https://tulocalweb.com.ar/l/avisos?utm_source={{site_source_name}}&utm_medium=paid&utm_campaign=avisos&utm_content={{ad.id}}
```

Al llegar, la página guarda el origen (`mostrador-origen`). Gana el primero, como en `/empezar`. Al
crear la cuenta, el origen queda en `tiendas.origen`, con `contenido` = id del anuncio. Si se entra
sin utm, queda `medio=aterrizaje` y `campana=<ángulo>`.

## Medición

- **Píxel** (`lib/anuncios/pixel.ts`): sólo en `/l/` y sólo con las cookies de marketing aceptadas.
  Manda `PageView`, `ViewContent` (a la mitad de la página) y `Lead` (al tocar el botón).
- **API de conversiones** (`lib/server/meta.ts`): `CompleteRegistration` desde el servidor, cuando
  el onboarding fija el origen de una cuenta nueva y la cookie `tlw-marketing` es `si`. Lleva el
  email hasheado, la IP, el navegador y `_fbp`/`_fbc`, y usa el `event_id` `alta-<tienda>`.
- Las políticas de cookies (v1.6) y de privacidad (v2.7) lo explican desde el 15/11/2026.

## Lo que hace falta de Meta (todavía no está)

1. Nahuel crea el Business Manager «Grupo Vimo», la cuenta publicitaria en ARS y el medio de pago.
2. En el Administrador de eventos, el píxel «Tu Local Web» y su token de la API de conversiones.
3. Se cargan en `/opt/tulocalweb/web.conf` (y en `CREDENCIALES.md`), y se recrea el contenedor web:
   `META_PIXEL_ID`, `META_CAPI_TOKEN` y, mientras se prueba, `META_CODIGO_PRUEBA=TEST…`.
4. Se verifica el dominio `tulocalweb.com.ar` con el TXT que da Meta, en el DNS de Cloudflare.
5. En la medición de eventos agregados, `CompleteRegistration` va primero.
6. AAIP: falta la modificación de la base de Tu Local Web en TAD (Meta, Estados Unidos, sólo con
   consentimiento). Es una declaración jurada y la firma Nahuel.

Sin el píxel cargado las páginas funcionan igual: sólo no miden en Meta, y el origen se guarda
siempre.
