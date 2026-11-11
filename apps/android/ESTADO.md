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
- **Play (11/11/2026):** app 4976235202658506239, firma de Play (SHA-256
  49:4F:40:03:3E:72:3B:23:38:C4:0F:1B:8E:6F:4C:D8:EF:A6:64:39:CB:D1:BC:52:60:BD:43:34:88:F7:11:70). Versión 1 (1.0.0)
  en prueba interna para la lista «Equipo»; link https://play.google.com/apps/internaltest/4700866843622052380. Nada
  en producción. El `.aab` sale de `gh workflow run android.yml` en grupovimo/tulocalweb-android (artifact `play`).

## Bugs del iOS encontrados al clonar (corregidos en los dos lados)
1. El aviso de venta no llevaba `category: "venta"` en APNs: los botones Ver pedido y Marcar como preparado nunca
   aparecían. Arreglado en la API (`push.Payload`).
2. Escanear inventario desde Más empujaba el producto en la pestaña Productos (no se veía nada). Ahora abre en la
   pestaña actual (`Navegacion.abrir`).
3. Crear tienda se tragaba los errores de Clerk (tocabas Google, fallaba y no pasaba nada). Ahora los muestra como
   Ingresar.
4. Pedido manual: el contador podía leer una fila ya borrada al bajar a 0 (cierre de la app).
5. Cuenta › Avisos: si todavía no se había pedido el permiso, la fila mandaba a Ajustes en vez de pedirlo.
6. Usuarios y permisos: faltaba «(vos)» cuando la persona no tiene nombre (paréntesis mal puesto).

## Dudas de producto (resueltas el 7/11, «lo que sea mejor para el usuario»; iguales en los dos)
- Editar la meta: el chip que proyecta el ritmo del mes dice «A este ritmo» (antes «Igual que hoy»).
- Pedidos › Más › «Cancelar pedido» abre la hoja de cancelar ahí mismo (antes abría el detalle).
- Cancelar el plan, último paso: «Volver» vuelve a la oferta; cerrar sin cancelar es «Me quedo».

## Diferencias de plataforma aceptadas
- Ingreso sólo con Google (sin «Seguir con Apple»).
- «Face ID» → «tu huella o tu cara»; «iPhone» → «teléfono» / «celular».
- La Live Activity y la Dynamic Island son una notificación en curso (Live Update en Android 16).
- El widget se actualiza cada 30 min (mínimo de Android) y sin Instrument Sans; no hay widgets de pantalla bloqueada.
- «Marcar como preparado» desde el aviso de venta abre la app y lo marca ahí (en iOS también abre la app).
- El ColorPicker de iOS es una grilla de colores con el código (`SelectorDeColor`); el Slider, `Deslizador`.

## Repasadas (6/11)
- Cuatro agentes en paralelo, el `.swift` entero contra el `.kt` entero (app y núcleo; inicio, pedidos, productos y
  venta; Tincho, clientes y la primera mitad de Más; la otra mitad de Más y el tema). Unas 20 correcciones.
- La mía: textos de cada `.swift` que no aparecen en Android (sólo quedan las diferencias aceptadas) y endpoints
  (idénticos en los dos lados). Piezas sumadas: `Confirmacion(verTitulo)`, `Texto(monoespaciado)`,
  `TextoRico(escalaMinima)`, `Hoja(estirable)` y la hoja que vuelve a su lugar si no se puede cerrar.
- CI verde: Android 51 pantallas en claro y oscuro; iOS con `respuestas.json`. Galería:
  https://claude.ai/artifact/9wkueUx5Wp6QJA3sr6H4Up
- API desplegada el 6/11 con la migración 00077 (copia previa `/var/backups/tulocalweb/antes-202611060653.sql.gz`);
  el worker lee `/claves/fcm.json` (`MOSTRADOR_FCM_CUENTA`, volumen sumado en `/opt/tulocalweb/compose.yml`).

## Sin probar en un teléfono
Nunca corrió en un aparato real. Con la prueba interna de Play: el ingreso con Clerk, los avisos con sus botones y la
notificación en curso, el escáner y el dictado.

## Repaso visual (7/11), a revisar por Nahuel
- Galería: https://claude.ai/artifact/9wkueUx5Wp6QJA3sr6H4Up
- Arreglado: el modo oscuro que salía crema, Pedidos sin resumen, textos con «…», alturas desparejas, renglones viudos y 12 pantallas sin datos.
- También en el iPhone (4bc06ad: FilasParejas, Segmentado, BotonPrincipal, «Hola» sin nombre, «Copiar link», margen del plan); TestFlight subido.
- Para retomar: Nahuel mira la galería y hace un repaso profundo de todo. Lo que marque se arregla en Android y,
  si también pasa en el iPhone, en los dos lados, con TestFlight después.

## Falta
1. App en Play y prueba interna («Run workflow» → artifact `play`).
2. Probar en un teléfono el envío por FCM (todavía no hay ningún Android registrado).
