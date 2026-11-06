# Tu Local Web para Android: estado y cómo retomar

Objetivo (Nahuel, 5/11/2026): Android **IDÉNTICA** a la app de iPhone (`../ios`), con el mismo núcleo y patrones que
Tengo Libre Android (el molde: `~/Trabajo/vimo/tengolibre/apps/android`, método en su `GUIA-CLON-IOS.md`). Nada se
publica en las tiendas hasta que esté. Plan: `~/.claude/plans/sigamos-haciendo-las-apps-humble-stream.md`.

## Reglas
- **Fuente de verdad: el código iOS.** Cada `.swift` tiene su `.kt` con el mismo nombre en la carpeta equivalente
  (`TuLocalWeb/Pantallas/Mas/Plan.swift` → `app/src/main/java/ar/com/tulocalweb/app/pantallas/mas/Plan.kt`), con los
  mismos nombres de vistas, funciones, textos y comportamiento.
- **Mismos nombres de tema que iOS:** `C.fondo`, `Estilo.tituloGrande`, `BotonPrincipal`, `TarjetaSeccion`… en
  `tema/`. Sin Material3.
- **Íconos:** por nombre de SF Symbol (`Icono("cart.fill")`; `.fill` elige la variante rellena). El equivalente de
  Material Symbols Rounded está en `herramientas/simbolos.txt`; al sumar uno, agregarlo y correr
  `herramientas/simbolos.sh` (pide fonttools: `python3 -m venv <dir> && <dir>/bin/pip install fonttools` y ese `bin` en
  el PATH).
- **Estado observable:** `app/Estado.kt` con `mutableStateOf`; `LaunchedEffect(estado.version)` es `.task(id:)`.
- **Compilación:** sólo en GitHub. `~/Trabajo/vimo/tulocalweb-android/publicar.sh "<mensaje>"` copia esto al repo
  público `grupovimo/tulocalweb-android`, escribe `paridad.txt`, copia `../pruebas/respuestas.json` a los recursos de
  test, commitea en los dos y empuja.
- **Respuestas simuladas compartidas:** `../pruebas/respuestas.json`, la misma que leen `TuLocalWebTests/PantallasTests`
  (iOS, como recurso del bundle de tests) y `PantallasTest` (Android). Cada prueba pisa o quita rutas.
- Sin APK: se revisa con la galería de capturas del CI.

## Hecho (6/11)
- Las 40 pantallas de iOS portadas (lo comprueba `ParidadTest`), el núcleo, el tema, `compartido/` (Hoy y Actividad)
  y el widget «Ventas de hoy» (Glance, chico y mediano).
- **Plataforma:** FCM sin google-services.json (secretos `TLW_FIREBASE_*`); los avisos llegan sólo con datos y la app
  arma la notificación (categoría `venta`: Ver pedido y Marcar como preparado). Los pedidos por preparar (la Live
  Activity) son una notificación en curso que la app y el servidor (`tipo: actividad`) mueven. Ingreso con Clerk por
  Custom Tab (vuelta `ar.com.tulocalweb.app://callback`, la misma de iOS). Bloqueo con huella o cara
  (BiometricPrompt). Escáner con CameraX + ML Kit. Dictado con SpeechRecognizer.
- **API** (apps/api): `internal/plataforma/fcm` (copia de Tengo Libre), `plataforma` en `auth.dispositivos`
  (migración `00077_android.sql`), avisos y actividad por FCM, `MOSTRADOR_FCM_CUENTA` en el worker.
- **Llave de subida:** `~/Trabajo/vimo/credenciales-android/tulocalweb-subida.{jks,clave}` (SHA-256
  3B:6C:DA:18:8C:55:C4:C6:4B:0F:E7:35:4F:45:24:84:AB:2C:76:FF:DA:6C:0B:EA:AC:D2:9A:8E:F2:16:97:62).

## Bugs del iOS encontrados al clonar (corregidos en los dos lados)
1. El aviso de venta no llevaba `category: "venta"` en APNs: los botones Ver pedido y Marcar como preparado nunca
   aparecían. Arreglado en la API (`push.Payload`).
2. Escanear inventario desde Más empujaba el producto en la pestaña Productos (no se veía nada). Ahora abre en la
   pestaña actual (`Navegacion.abrir`).

## Diferencias de plataforma aceptadas
- Ingreso sólo con Google (sin «Seguir con Apple»).
- «Face ID» → «tu huella o tu cara»; «iPhone» → «teléfono» / «celular».
- La Live Activity y la Dynamic Island son una notificación en curso (Live Update en Android 16).
- El widget se actualiza cada 30 min (mínimo de Android) y sin Instrument Sans; no hay widgets de pantalla bloqueada.
- «Marcar como preparado» desde el aviso de venta abre la app y lo marca ahí (en iOS también abre la app).
- El ColorPicker de iOS es una grilla de colores con el código (`SelectorDeColor`); el Slider, `Deslizador`.

## Sin probar en un teléfono
Nunca corrió en un aparato real. Con la prueba interna de Play: el ingreso con Clerk, los avisos con sus botones y la
notificación en curso, el escáner y el dictado.

## Falta
1. Cuenta de servicio de Firebase en el VPS (`MOSTRADOR_FCM_CUENTA`) y reiniciar el worker.
2. App en Play y prueba interna («Run workflow» → artifact `play`).
