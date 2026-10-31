# Decisiones

Lo que el código da por sentado, y por qué. Se agrega una entrada por decisión, nunca se reescribe.

## 1. Front siempre en Workers, API en Go para el VPS (20/10/2026)
La demo pública corre entera en Cloudflare con un sandbox por visitante en D1. La app real usa una
API en Go con Postgres 17 en un VPS, por concurrencia: RLS para aislar tiendas, reserva de stock con
`UPDATE` condicional, outbox en la misma transacción. D1 no tiene transacciones interactivas ni RLS.
El navegador sólo habla con SvelteKit (BFF); SvelteKit habla con Go server-to-server.

## 2. Una sola app SvelteKit
Landing, panel, tiendas y conector MCP comparten despliegue. Las tiendas van por path (`/t/<tienda>`)
en `tulocalweb.com.ar`; con dominio propio, por subdominio, resuelto en el hook `reroute`.

## 3. La impronta de Caja
Mismo gestor (bun), mismas convenciones y optimizaciones, y el código de Caja se reutiliza cuando
sirve. Excepciones pedidas: Playwright E2E en cada PR e íconos Phosphor.

## 4. Contrato único entre la demo y la API
`packages/contrato/openapi.yaml` define cada operación; una suite de contrato corre los mismos casos
contra el sandbox y contra la API Go, para que la demo no muestre algo que la app real no hace.

## 5. El aislamiento entre tiendas lo garantiza Postgres (PR #4)
Cada tabla con `tienda_id` pasa por `app.aplicar_rls`: RLS habilitada y **forzada** (alcanza al
dueño), política por tienda y lectura de staff. La tienda vive en `set_config('app.tienda_id', …, true)`,
local a la transacción: nunca un `SET` de sesión, que con un pool compartido filtraría una tienda a
otra. Sin tienda fijada, `app.tienda_actual()` es NULL y no se ve nada.
Roles: `mostrador_migrador` (dueño del esquema), `mostrador_app` (sin bypass, sin ser dueña de nada),
`mostrador_worker` (sólo el outbox) y `mostrador_staff` (sólo lectura). Tres tests lo sostienen: el
estructural (rompe si una tabla queda sin RLS; se verificó agregando una a propósito), el de cruces
entre dos tiendas y el de arquitectura (pgxpool sólo en `internal/plataforma/db`).

## 6. Outbox y worker en Go
El evento se inserta en la misma transacción que el cambio. El worker toma lotes con
`FOR UPDATE SKIP LOCKED` (varios workers no se pisan: probado con 4 en paralelo sobre 300 eventos),
reintenta con espera creciente (2, 4, 8… s, techo 1 h) y aparta el evento después de 8 intentos con
su error. La entrega es al menos una vez: los manejadores tienen que ser idempotentes. La app puede
insertar en el outbox pero no marcarlo.

## 7. Idempotency-Key en las escrituras
La clave se reserva en su propia transacción antes de ejecutar: un segundo pedido igual recibe 409
"en curso" o, si ya terminó, la respuesta guardada (`Idempotent-Replayed: true`). La misma clave con
otro cuerpo es 422. Un 5xx libera la clave para poder reintentar.

## 8. Errores como problem+json
Todos los errores salen como `application/problem+json` con `code` estable y un `detail` que le
sirve a una persona. Un panic en un handler es un 500 con ese formato, no un proceso caído.

## 9. El sandbox de la demo es "copia al escribir" (PR #5)
El cupo gratis de D1 son 100.000 escrituras por día. Si cada visitante copiara su tienda de ejemplo
(el handoff muestra 2.400 pedidos), 40 visitas lo agotarían. Por eso los datos de ejemplo de cada
rubro viven en el código, iguales para todos, y D1 guarda sólo el sandbox (una fila) y lo que el
visitante cambia (`cambios`, por entidad y clave). Leer es: ejemplo + cambios del sandbox.
El sandbox vence a las 24 h; `mostrador-jobs` llama a `/api/internal/cron` una vez por hora y la
purga va de a tandas de 50. En desarrollo y en los e2e se usa la D1 local, nunca la remota.

## 10. Contrato compartido entre la demo y la API Go
`packages/contrato` tiene el OpenAPI (de donde salen los tipos del front), los rubros y los casos
(`casos/*.json`). El mismo caso corre en vitest contra el sandbox y en Go contra Postgres, así la
demo no puede mostrar algo que la API real no devuelve. Go no puede embeber archivos fuera de su
módulo: sus rubros son una tabla propia que un test compara contra `rubros.json`.

## 11. Las páginas no saben de dónde salen los datos
`datosDe(locals)` devuelve la implementación de `DATA_SOURCE`: el sandbox (D1) o la API Go (BFF:
el navegador nunca habla con Go). Los hooks sólo leen el sandbox en `/demo` y `/panel`, para que
la landing no gaste lecturas de D1.

## 12. La demo no tiene cuentas (PR #7)
En `DATA_SOURCE=sandbox`, `/ingresar` funciona entero (validación, errores, modos, medidor de fuerza,
recuperar) pero no crea cuentas: ingresar, registrarse o entrar con Google/WhatsApp llevan a la
demo, y "recuperar" muestra "Revisá tu email" sin mandar nada. Las cuentas reales viven en la API Go
(Postgres): hashear contraseñas en Workers Free pasaría el tope de 10 ms de CPU, y una demo pública
no necesita guardar datos personales. El destino de "recuperar" vuelve en la respuesta, no en la
URL, para que no quede en el historial ni en logs.

## 13. Cuentas y sesiones en la API Go (PR #8)
- Contraseñas con argon2id (parámetros de OWASP); el hash guarda sus parámetros para poder subirlos.
- Un email inexistente tarda lo mismo y responde igual que una contraseña incorrecta, y "recuperar"
  responde 202 siempre: la API no delata quién tiene cuenta.
- Las sesiones guardan el sha256 del token. El token se entrega una sola vez al BFF, que lo guarda en
  la cookie `__Host-mostrador_sesion` (httpOnly, Secure, SameSite=Lax); el navegador nunca habla con
  Go. "Visto" se actualiza como mucho una vez por hora.
- Cambiar la contraseña con el link (un solo uso, 30 min) cierra todas las sesiones de la cuenta.
- Límites de intentos con un upsert atómico (`auth.intentos`), como Caja: 5 ingresos por email y 30
  por IP cada 15 min; 3 links de recuperación por email por hora.
- La IP del visitante llega del BFF en `X-Cliente-IP` y Go sólo le cree si viene con la clave del
  BFF; si no, usa la IP de la conexión.
- Cuentas, membresías y sesiones viven en el esquema `auth` (una persona puede tener varias
  tiendas), fuera de RLS. Los slugs tomados se registran en `app.slugs`: con RLS forzada ni el dueño
  ve las tiendas ajenas para chequear si un slug está libre.
- En Workers, `fetch` guardado como propiedad y llamado como método da "Illegal invocation": los
  clientes HTTP lo envuelven en una función. Lo detectó el e2e contra la API real, no los unitarios.

## 14. Legales y botón de arrepentimiento (PR #12)

- Los 8 legales salen de `design/Legal *.dc.html` con `scripts/legales.ts` (genera `src/lib/legales.ts`):
  el texto no se copia a mano, y el HTML queda sólo con `<b>` y `<a>` (links reescritos a rutas).
- Las páginas públicas comparten el layout `(sitio)`: header y footer del handoff y el banner de cookies.
  En el celular la nav del header baja a una fila deslizable y el índice lateral de los legales
  (sticky en escritorio) pasa a chips + un desplegable: con poco ancho, el sticky tapaba el texto.
- Botón de arrepentimiento (Res. 424/2020): con la API, `POST /v1/arrepentimientos` lo registra en
  `soporte.arrepentimientos` con código `ARR-AAAA-NNNNNN`, límite por IP y por email, y mail de
  confirmación. La app sólo puede insertar y leer el código; los datos personales los ve staff.
  En la demo no hay a quién revocarle nada: se valida igual y se da un código correlativo (tabla
  `contadores` de D1), sin guardar ningún dato del pedido.

## 15. Formularios públicos (PR #14)

- Contacto, postulación de partners y novedades: dominio Go `formularios` (tablas en `soporte`),
  con límite por IP y por email, mail de confirmación a quien escribe y aviso a la bandeja del
  equipo (`MOSTRADOR_CORREO_EQUIPO`, opcional; también recibe los arrepentimientos).
- La app sólo inserta y lee el id: las consultas y postulaciones las lee staff. Por eso la
  postulación repetida se detecta por el error de unicidad y no con `ON CONFLICT` (que exigiría leer
  los emails).
- La baja de novedades es un POST con el token del mail: abrir el link sólo muestra el botón, así los
  clientes de correo que previsualizan links no desuscriben a nadie.
- En la demo los formularios validan igual que Go y muestran el éxito, sin guardar ni mandar nada.
- Los montos usan espacio duro entre "$" y el número (nunca se parten en dos renglones).

## 16. /estado con monitor propio (PR #16)

- El monitor corre en `mostrador-jobs` (el cron horario que ya existía): un chequeo por componente,
  unos 170 pedidos por día. Vive fuera del sistema que mide y no necesita UptimeRobot ni el plan
  Pro de Cloudflare.
- Guarda en su propia base D1 (`mostrador-estado`): resumen por día, último resultado e incidentes.
  El sitio sólo la lee, también en modo API, porque el estado tiene que verse aunque lo demás falle.
- Al propio sitio lo chequea por service binding (un worker no puede pedirle por URL pública a otro
  de la misma zona); a los proveedores, que respondan (sin credenciales, un 401 también dice "arriba").
  La API Go se chequea sólo si `API_URL` está configurada.
- Un incidente se abre con dos fallas seguidas y se cierra solo cuando vuelve; los días sin chequeos
  se ven grises ("Sin datos"), nunca verdes inventados.
- Avisos a suscriptos por mail con Resend si `mostrador-jobs` tiene `RESEND_API_KEY`; por WhatsApp,
  cuando esté WhatsApp Cloud. La baja es un POST con token, como en novedades.

## 17. Onboarding (PR #17)

- `/onboarding` es el registro (decisión §0): todos los "Crear mi tienda" de la landing y del sitio
  llevan ahí, y "Usar esta plantilla" lo arranca con esa plantilla (`ob-tpl`, como en el handoff).
- Lo elegido queda en el navegador (`ob-estado`) para que recargar no pierda nada. En la demo, al
  terminar se crea el sandbox con la plantilla elegida y lo demás se guarda como cambio del sandbox.
  Con la API, el paso 1 pasa por `/ingresar` (que al registrar vuelve a `?paso=1`) y cada paso se
  guarda con `PUT /v1/panel/onboarding` (columna `tiendas.onboarding`, jsonb hasta 16 KB).
- Los datos que muestran los pasos son los del negocio de ejemplo del handoff (Café La Esquina);
  traerlos de verdad de cada integración llega con el PR de esa integración.
- **Con la API, menos es más (31/10)**: la demo sigue con los 8 pasos del handoff; una cuenta real
  recorre 6 (`PASOS_REALES` en `lib/onboarding/datos.ts`): Cuenta, Tu negocio (nombre, dirección y
  rubro opcional), Importación, Tu tienda, Cobros y envíos (Mercado Pago, transferencia y entregas)
  y ¡Listo!. Salen Canales (Mercado Pago pasa a Cobros; lo demás, a Integraciones), Tincho (el tono
  no lo leía nada), los objetivos (tampoco), ARCA y la política de cambios (al panel), y en ¡Listo!
  el equipo (no mandaba invitaciones) y «te quedan 2 cosas». El avance es pasos hechos / pasos.
- Se retoma de la tienda, no del navegador: `GET /v1/panel/onboarding` devuelve lo guardado (con el
  paso, por id) y si terminó. Al llegar a ¡Listo! queda terminado y las secciones pasan a la tienda,
  así el link y el QR ya muestran lo elegido; después `/onboarding` lleva al panel (sólo ¡Listo! se
  puede recargar). Cada paso es una entrada del historial: el «atrás» del celular vuelve un paso.
- ¡Listo! termina en algo accionable: compartir (hoja del sistema en el celular, copiar y WhatsApp en
  la compu, `compartirTienda` del panel), descargar el QR, ver la tienda y «Cargar mi primer
  producto» (`/panel/productos?nuevo=1`), o «Traer mis productos» si eligió importar.

## 18. Ingreso con Google, Mercado Pago, Mercado Libre y WhatsApp (PR #18)

- OAuth 2 con PKCE (S256) en la API: Go guarda el state (hasheado, un uso, 10 min) y el
  verificador; el BFF sólo redirige y, en `/ingresar/<proveedor>/vuelta`, canjea y guarda la
  sesión. Una cuenta nueva sigue al onboarding (tienda provisoria "La tienda de <nombre>").
- Vinculación por email sólo si el proveedor lo verifica (Google). Mercado Pago y Mercado Libre no
  garantizan el email: si ya hay una cuenta con ese email no se vinculan solas (error que explica
  cómo hacerlo desde Mis datos). Así nadie entra a una cuenta ajena creando otra en un proveedor.
- WhatsApp: link para entrar sin contraseña (15 min, un uso) y link de contraseña nueva, al número
  de la cuenta. Sin pantalla de código: el handoff no la tiene y sí tiene "Revisá tu WhatsApp". El
  link pide un clic para entrar porque WhatsApp abre los links para armar la vista previa. Responde
  igual haya o no cuenta con ese número.
- Cada proveedor se prende al cargar sus credenciales; apagado, el botón lo explica. En desarrollo
  los mensajes de WhatsApp van al log.

## 19. Los datos de cada demo (PR #19)

- Cada demo es la tienda de su plantilla, con su catálogo del handoff: `scripts/catalogos.ts` genera
  `src/lib/catalogos.ts` desde `Plantillas Tienda` (12 tiendas) y las plantillas trabajadas (Ceibo,
  Mostaza, Yerbal, Tijera). Las categorías de cada producto se asignan en el generador (el handoff
  da las de la tienda, no en cuál va cada uno).
- Clientes y pedidos de 30 días se generan con semilla por plantilla (`sandbox/demo.ts`): siempre
  los mismos, fechados desde que se creó el sandbox. No se guardan: D1 sólo guarda lo que cambia el
  visitante. Canales, pagos, envíos y estados son los del panel del handoff.
- `GET /v1/panel/resumen` (insignias, notificaciones, canales y prueba) está en el contrato. En la
  demo sale de esos datos; en Go, la prueba es real y lo demás va en cero hasta que existan los
  dominios (pedidos, carritos, WhatsApp, tareas).

## 20. Dominio: mostradorapp.com.ar (PR #52)

- mostrador.com.ar está registrado por un tercero (vence a mediados de diciembre). La marca sigue
  siendo Mostrador y el dominio será **mostradorapp.com.ar**: los mails públicos (ayuda@,
  legales@, seguridad@…) ya lo usan. Hasta que esté comprado y con DNS, el sitio sigue en
  tulocalweb.com.ar (`PUBLIC_SITE_URL` / `MOSTRADOR_SITIO`).
- Al pasarlo: cambiar esas dos variables, la ruta de `wrangler.jsonc`, y en los portales las vueltas
  de OAuth (`/integraciones/mercadopago/callback`, `/integraciones/mercadolibre/callback`,
  `/ingresar/*/vuelta`) y las notificaciones (`api.mostradorapp.com.ar/v1/webhooks/*`). Mercado
  Libre pide validar con QR desde el celular de Nahuel para editar la app.

## 21. Manejá Mostrador desde tu IA: servidor MCP con OAuth propio (PR #54)

- Mismo patrón que Caja: Mostrador es su propio servidor de autorización (OAuth 2.1 para clientes
  MCP) y expone `/api/mcp` sin estado. Registro dinámico (clientes públicos, vueltas https o
  localhost, 20 por hora por IP), PKCE S256 obligatorio, código de un solo uso (5 min), access token
  de 1 h y refresh token que rota en cada uso; si llega uno ya rotado, se corta la conexión entera.
  Códigos y tokens se guardan hasheados (sha256).
- En la demo todo vive en la D1 del sandbox (migración 0005): la conexión es del sandbox del
  visitante (el de su cookie al dar permiso), vence con él y se borra con él (`ON DELETE CASCADE`).
  El cron limpia códigos vencidos, clientes de más de 30 días sin conexiones y los límites.
- Una fila por conexión con el access token, el refresh token y la ventana del límite (60 pedidos
  por minuto): autenticar y contar es una sola escritura (`UPDATE … RETURNING`).
- Permisos: `lectura` (siempre) y `acciones` (si la IA lo pide y la persona lo deja marcado). Las
  acciones además respetan los interruptores de Integraciones: ajustar productos pide "Cargar y
  editar productos" y crear cupones, "Reembolsos y descuentos".
- Las herramientas trabajan sobre `Datos`, sin métodos nuevos: cuando exista la API Go, la misma
  lista sirve contra `DatosApi`. Montos en pesos enteros, con el texto formateado al lado.
- Con `DATA_SOURCE=api` los endpoints responden 501 con un mensaje claro: la versión real (atada a la
  cuenta y la tienda, no al sandbox) la va a atender Go.
- `/api/mcp` queda fuera del chequeo de CSRF (no usa cookies, sólo el token Bearer), como los
  endpoints de OAuth. La pantalla de consentimiento sí lo pasa.

## 22. El MCP en producción: OAuth en Go y el token de la IA como sesión limitada (PR #58)

- Con `DATA_SOURCE=api`, clientes, códigos y conexiones viven en Postgres (`auth.ia_*`, sólo hashes).
  La conexión es de una persona en una tienda: si le sacan la membresía, el token deja de servir.
- El token de acceso entra por el mismo `Autenticar` que las sesiones (no hay otra puerta), pero
  `ia.Guardia` lo deja pasar sólo a las rutas que usan las herramientas (lista explícita), con GET
  si la conexión es de lectura, y controla en el servidor los interruptores del panel ("Cargar y
  editar productos", "Reembolsos y descuentos"): el BFF también los mira, pero no es el que manda.
  Cuenta, usuarios, plan, sesiones y cobros quedan afuera aunque quien autorizó sea la dueña.
- Sólo dueña o admin autorizan. Refresco rotativo: si llega uno ya usado, se borra la conexión.
- El BFF sigue siendo el que habla con las IA (nombres de OAuth, `/api/mcp` JSON-RPC); Go sólo ve
  su propio contrato (`/v1/ia/*`, `/v1/panel/ia*`).
- Goose corre con migraciones fuera de orden: varias ramas suman migraciones a la vez.

## 23. Turnos con seña (PR #59)

- Modelo: `turnos_ajustes` (una fila por tienda; sin fila, apagado con seña 30 %, 60 min de
  anticipación y 14 días de vista), `turnos_servicios`, `turnos_profesionales` y `turnos`, todas con
  `app.aplicar_rls`. Cada turno copia el nombre del servicio y del profesional (no son FK): borrar
  uno no deja la agenda en blanco. El turno tiene un `codigo_publico` aleatorio para su página
  (`/t/{slug}/turno/{codigo}`), como los pedidos de la tienda online.
- Disponibilidad (`turnos.Libres` en Go, `lib/turnos/libres.ts` en la web): grilla de 15 min desde
  el comienzo de cada franja, el servicio tiene que entrar entero en una franja, el profesional lo
  tiene que hacer (lista vacía = todos) y no tener otro turno que se superponga. Los días que se
  ofrecen son hoy y los `diasVista − 1` siguientes, y nada antes de `ahora + anticipacionMin`.
  "Cualquiera" sirve si algún profesional está libre; al reservar se asigna el primero libre por
  `orden`. Las dos implementaciones corren la misma tabla de casos
  (`apps/api/internal/dominio/turnos/testdata/libres.json`).
- Sin cron para liberar horarios: un turno `reservado` sin seña pagada deja de bloquear cuando pasa
  su `venceEn` (15 min). La regla está en la disponibilidad, no en un proceso que tenga que correr;
  la fila queda `reservado` y la API la muestra con `vencido: true`. La preferencia de Mercado Pago
  vence a la misma hora, así nadie paga una seña de un horario que ya se liberó.
- Dos reservas al mismo horario: la reserva bloquea los profesionales activos de la tienda
  (`SELECT … FOR UPDATE`, siempre por id para no trabarse) y recién ahí mira los turnos e inserta,
  todo en la misma transacción. Una restricción de exclusión no alcanza porque "vencido" depende de
  la hora actual. El test concurrente lo prueba (10 a la vez: gana uno; sin el bloqueo, fallaba).
- Seña: porcentaje del precio del servicio, redondeado al peso. Precio, duración y seña salen del
  servidor. Con seña > 0 y Mercado Pago conectado, el turno queda `reservado` y se devuelve el
  `init_point` (Checkout Pro con la cuenta de la tienda, `mp_cobros.tipo = 'turno'`, sin cuotas).
  El webhook aprobado lo pasa a `confirmado` con `senaPagada`. Sin Mercado Pago o con seña 0, queda
  `confirmado` y sin seña. Si Mercado Pago no responde al armar el pago, el turno se cancela en el
  momento para no tomar el horario.
- Cancelar (la persona desde su página, hasta `anticipacionMin` antes; la tienda desde el panel)
  nunca devuelve la seña sola: la tienda la devuelve con "Devolver la seña" (refund del pago en
  Mercado Pago, sólo dueña o admin), y el turno queda con `senaDevuelta`.
- El turno a mano del panel no tiene seña ni mira la grilla; si se superpone con otro, pide
  confirmar (`forzar`). Turno nuevo y cancelado van al outbox (`turno.nuevo`, `turno.cancelado`)
  para los avisos.
## 24. Cupones en la tienda online (PR #65)

- Una sola regla en TS (`lib/tienda/cupon.ts`) y en Go (`vidriera/cupon.go`), con los mismos casos
  en `vidriera/testdata/cupones.json`. El navegador la usa para mostrar el descuento; al comprar,
  el servidor la vuelve a correr y nunca se cree el total del navegador.
- El uso se cuenta en la misma transacción que crea el pedido (el pedido guarda `cupon` y el cupón
  suma usos, vendido y descontado, migración 00028), así el límite de usos no se pasa.
- "VIP" = quien tiene algún pedido VIP con el mismo email; "en riesgo" = su última compra fue hace
  más de 60 días. Los dos se calculan en el servidor a partir de los pedidos.
- El cupón por categoría guarda la categoría (antes se perdía); si no quedó guardada, vale para
  toda la tienda.
- La vista previa del cupón (`/t/[tienda]/cupon`) tiene un tope de 30 pruebas por IP cada 10 min,
  con el limitador del #54, para que no se adivinen códigos probando.
## 25. Medición de la tienda: consentimiento primero y sin API de conversiones todavía (PR #67)

- Google Shopping, Google Analytics y Meta Ads usan las cuentas de la tienda, no credenciales de
  Mostrador: la tienda pega sus IDs (GA4 y píxel) y carga el feed `/t/{slug}/feed.xml` en Merchant
  Center y en Commerce Manager. "Conectada" es el ID guardado o el "Ya lo cargué" del feed: no hay
  API que confirme que Google o Meta lo leyeron.
- gtag.js y fbevents.js se cargan sólo si la tienda tiene IDs y quien compra aceptó las cookies de
  medición en el aviso de esa tienda (se recuerda por tienda en el navegador). Sin elección no se
  pide nada a Google ni a Meta; la vitrina de una plantilla no mide nunca. Un solo módulo
  (`lib/tienda/medicion.ts`) manda los eventos desde el layout, las rutas y el carrito compartido:
  las plantillas no se tocan.
- La compra se mide en la página del pedido con el número de pedido como `transaction_id` (GA4) y
  `eventID` (Meta), para que deduplique cuando llegue la API de conversiones.
- La API de conversiones de Meta queda para otro PR: necesita un token de acceso de la tienda
  guardado cifrado (como Mercado Pago) y mandar las compras desde Go con los datos de quien compra
  hasheados. Hasta entonces, Configurar la nombra como "Próximamente" sin prometer fecha.
- La CSP se suma sólo en `/t/`: scripts y conexiones al propio sitio, a Google y a Meta. El resto
  del sitio sigue sin CSP hasta el PR de endurecimiento.
## 26. Facturación con ARCA (PR #71)

- **La clave nunca sale de Mostrador.** Go genera la clave RSA de 2048 bits, la guarda cifrada con
  la clave de la tienda (`plataforma/cifrado`, como los tokens de Mercado Pago) y le da a la tienda
  sólo el pedido de certificado (CSR con `C=AR`, `O=<razón social>`, `CN=mostrador`,
  `serialNumber=CUIT <cuit>`). El certificado que pega la tienda se revisa contra esa clave y ese
  CUIT. Generar un pedido nuevo reemplaza la clave: el certificado anterior deja de servir (el
  asistente lo avisa). Sólo dueña o admin.
- **Sin dependencias.** La firma CMS (PKCS#7 SignedData) se arma a mano con `encoding/asn1`
  (atributos firmados como SET OF ordenados por DER); los tests la verifican con
  `openssl cms -verify` y contra un ARCA falso que valida como el real (10016, 10048, IVA de la C,
  condición del receptor).
- **Cola en la misma transacción del cobro.** El trigger `app.encolar_factura` escribe el
  comprobante cuando el pedido pasa a cobrado, en la misma transacción: si el pago se guardó, la
  factura está encolada. Un drenador (worker, cada 30 s) emite de a uno y en orden por tienda, con
  un advisory lock para que la numeración correlativa no se pise.
- **"No facturado" es un estado, no un hueco.** Sin ARCA conectada, o con Facturar al cobrar
  apagado, la venta cobrada queda como comprobante `no_facturado` con su motivo (la base lo exige):
  toda venta cobrada tiene su fila. El panel lo muestra como "No facturada: <motivo>".
- **Recuperación con FECompConsultar.** Antes de pedir el CAE se guarda el número intentado. Si la
  respuesta se cortó y ARCA ya numeró ese comprobante, el reintento lo consulta y recupera el CAE en
  vez de emitir otro por la misma venta (sacar la recuperación hace fallar el test). Un rechazo
  guarda el código y el mensaje de ARCA; los errores de red reintentan con espera creciente hasta
  5 min y 8 intentos.
- **Redondeo al IVA.** Los importes se congelan al encolar, en centavos, con `neto + iva = total`:
  el neto se calcula dividiendo el precio final por 1,21 y el IVA es lo que falta, así la suma
  cierra con dos decimales. La C va sin IVA.
- **Letra según la condición.** Monotributo emite C (y nota de crédito C); responsable inscripto,
  A si el cliente es inscripto con CUIT válido y B si no. La nota de crédito anula entera la factura
  emitida, asociada a ella, una sola vez.
- **La demo no miente.** Arranca conectada en homologación (CUIT de la dueña de ejemplo) para que
  Facturación muestre comprobantes. El pedido de certificado es real, pero lo arma el navegador del
  visitante con WebCrypto (`lib/panel/csr.ts`, el mismo sujeto que Go): una clave RSA tarda de 20 a
  100 ms de CPU y Workers Free da 10 ms. La clave no se guarda; guardar un certificado y probar la
  conexión responden que es una tienda de ejemplo. Facturar al cobrar se guarda con su historial y
  las ventas cobradas mientras estuvo apagado aparecen como no facturadas; anular deja la factura
  anulada y suma su nota de crédito.
- **Sin verificar todavía:** que ARCA acepte este CMS. Los tests pasan contra `openssl` y el ARCA
  falso, pero el WSAA real sólo se prueba con un certificado de homologación de verdad (WSASS con la
  clave fiscal de Nahuel). Se cierra cargando ese certificado en una tienda de prueba y tocando
  "Probar conexión".
## 27. Reseñas y preguntas de la tienda online (PR #77)

- **Con compra se deriva del pedido**, no se marca a mano: una reseña es "con compra" si trae el
  código del pedido (el público en Go, el firmado de la confirmación en la demo) y ese pedido está
  cobrado (Aprobado, Acreditada o Cobrado). Una por pedido y producto (409 `ya_reseno`). Con un
  código que no sirve se guarda igual, sin compra. Las de ejemplo de la demo que vienen de la tienda
  online cuentan con compra.
- **Ocultar, no borrar:** la tienda decide. Una oculta sale de la tienda y del promedio, pero queda
  en el panel con "Mostrar" (y Deshacer). La de una estrella sin compra se destaca en el panel con
  un aviso; no se oculta sola.
- La respuesta de la tienda es pública y va debajo de la reseña; vacía, se saca. Las preguntas de la
  ficha llegan al panel con id `t-…`, junto a las de Mercado Libre, y se ven en la ficha recién
  cuando están respondidas.
- **Tope por IP:** 10 reseñas y 10 preguntas por hora. En Go (`limite.Permitido`, con la IP que
  manda el BFF en `X-Cliente-IP`); en la demo, `golpe` de `lib/server/limite.ts` sobre la D1. En la
  vitrina de una plantilla no se guarda nada (responde que es una tienda de ejemplo).

## 28. Variantes y galería de fotos

- Un producto tiene hasta 3 **opciones** (`productos.opciones`, jsonb `[{nombre, valores}]`, hasta 20
  valores cada una) y, con opciones, se vende por **variante**: una fila de `variantes` por
  combinación (`valores` = `{"Talle": "M", "Color": "Negro"}`), con PK/FK compuestas con productos
  (como `ml_publicaciones`), única por `(tienda_id, producto_id, valores)` y RLS por tienda. Tope de
  100 combinaciones por producto (lo mismo que admite Mercado Libre por publicación).
- **El stock vive en la variante.** Es lo que se vende y lo que se cuenta: si viviera en el producto
  habría que repartirlo, y dos ventas de talles distintos competirían por el mismo número. El
  `productos.stock` de un producto con opciones es la suma de sus variantes activas y lo mantiene un
  trigger (`app.variantes_stock`), así que ningún camino (venta, ajuste, conteo, edición) lo deja
  desparejo, y como toca `productos.stock`, el trigger de Mercado Libre se entera solo. Las acciones
  de stock sobre el producto entero no lo tocan si tiene opciones.
- Al guardar opciones se generan o ajustan las combinaciones: las que siguen conservan stock, precio,
  SKU y foto; las nuevas arrancan en 0 y activas; las que dejan de existir se borran (los pedidos ya
  guardaron el texto). El editor manda sólo las variantes que toca, identificadas por sus valores.
  `opciones` sin mandar deja las que había; `[]` vuelve a producto simple.
- Vender una variante es un `UPDATE … WHERE activa AND stock >= n` sobre esa fila, unida a su producto:
  dos cajas no sobrevenden y nunca se acepta una variante inactiva o de otro producto. Un producto con
  opciones exige `varianteId`. Las líneas se descuentan en orden (producto, variante) para que dos
  ventas con los mismos productos en otro orden no se traben. El ítem guarda `varianteId` y el texto
  ("Talle M · Color Negro", nombre y valor de cada opción en orden).
- Inventario: la fila del producto trae `variantes` y las acciones `ajustar`/`ingreso`/`conteo`
  aceptan ids `producto:variante` (el `id` de cada variante en la lista ya viene así).
- **Galería**: `productos.fotos` (hasta 8); `foto` sigue siendo la primera, para todo lo que ya la
  usa. Un editor que no manda `fotos` reemplaza la portada con `foto` y deja el resto. La migración
  arranca `fotos` con la foto que había.
- Mercado Libre: con variantes se publica con `variations`. Cada opción se mapea a un atributo de la
  categoría que admita variaciones (por nombre o alias: talle→SIZE, color→COLOR…); la que no mapea va
  con su nombre como atributo personalizado y, si Mercado Libre lo rechaza, no se publica y se avisa.
  También se avisa si la categoría no admite variaciones o si las variantes tienen precios distintos
  (Mercado Libre pide uno solo). Qué variación es cada variante queda en
  `ml_publicaciones.variaciones`; con eso se sincroniza el stock por variación y las ventas entran
  como su variante. Una variante creada después de publicar no se agrega sola: hay que volver a
  publicar.


## 24. Hornero en Telegram con la API real (PR #68)

- Vive en Go en `internal/integraciones/telegram` (migración 00036: 00027–00035 las toman otras
  ramas). `telegram_vinculos` (una fila por tienda, `chat` UNIQUE: un chat no puede estar en dos
  tiendas aunque RLS no deje ver la otra) y `telegram_codigos` (sha256 del código, 10 min, un solo
  uso), las dos con `app.aplicar_rls`. El bot sólo sabe el chat o el código: `app.tienda_de_telegram`
  y `app.tienda_de_codigo_telegram` devuelven el id de la tienda (como `app.tienda_de_ml`) y todo lo
  demás se hace con la tienda fijada.
- Generar un código borra el anterior de la tienda; un código vencido se borra al intentar usarlo.
  Canjear: 10 intentos por hora por chat. Vincular otro chat a la misma tienda reemplaza al anterior;
  un chat que ya está en otra tienda no se mueve solo (409 `chat_ocupado`: primero `/desvincular`).
  Mover el chat de una tienda a otra implicaría tocar una tienda ajena desde un pedido que no es suyo.
- Para responder, el BFF le pide a Go **el contexto** de la tienda del chat (`POST /v1/telegram/contexto`:
  tienda, tablero de hoy y de 7 días, `ContextoHornero` y los mensajes del día), no una sesión
  limitada. Es lo más simple y lo más seguro: no hay un token nuevo que guardar, vencer, limitar ni
  dejar pasar por la Guardia, y el chat no puede pedir nada más que eso. El contexto sale de las
  mismas funciones que las rutas del panel (`tiendas.Obtener`, `ObtenerTablero`, `ContextoHornero`),
  así que el panel y Telegram no se separan. Si un día Hornero en Telegram tiene que hacer acciones,
  ahí sí conviene la sesión limitada de `ia` (`SesionIA` + Guardia).
- Las rutas `/v1/telegram/*` exigen la clave del BFF (`cuentas.EsBFF`, tiempo constante; sin clave
  configurada, nunca pasan): el que habla con Telegram es el Worker, que ya verificó el secreto del
  webhook. Las del panel (`/v1/panel/telegram`) van con la sesión; generar y desvincular, sólo dueña
  o admin. Desvincular devuelve el chat que estaba para que el BFF le avise (el navegador no lo ve).
- Topes: Go cuenta los mensajes por tienda y día de Argentina (`uso_dia`, `uso_n`, un solo `UPDATE …
  RETURNING`) y el Worker aplica los mismos de `cupo.ts`: 15 por tienda por día y el tope diario de
  Workers AI de toda la cuenta (el contador de D1, que es de la cuenta y no de la demo).
- Aviso de venta: el trigger `app.avisar_venta_telegram` encola `telegram.venta` al insertar un
  pedido de una tienda con chat (por cualquier camino: tienda online, panel, local, Mercado Libre);
  sin chat no encola nada. El worker manda "🛒 Nueva venta {canal}: pedido #N por $ X." por la Bot API
  con `MOSTRADOR_TELEGRAM_BOT_TOKEN` (sin token, no hace nada) y lo marca en `pedidos.avisos.telegram`
  (una sola vez aunque el outbox entregue dos). Un 4xx de Telegram (bot bloqueado, chat borrado) no
  se reintenta; un 429 o 5xx, sí. Los turnos nuevos todavía no avisan por Telegram con la API.


## 26. Panel de partners (PR #75)

- Un partner es una cuenta de usuario (no una tienda): `partners.partners` (nivel, código del link,
  cupón, perfil público y datos de cobro), `partners.clics` (contador por mes) y `partners.pagos`
  (liquidación de cada mes con el detalle por tienda congelado y el link de la factura), migración
  00037. RLS forzada por persona: `db.EnUsuario` fija `app.usuario_id`, como `EnTienda` fija la tienda.
- Atribución: una tienda es del partner si su origen (`/empezar`, migración 00030) trae su `ref`
  (sin distinguir mayúsculas) o su cupón. La lee `partners.tiendas_referidas()` (security definer:
  abre sólo esas tiendas y sólo nombre, plan, ciclo y fechas). El origen no cambia después del alta,
  así que una comisión no cambia de dueño.
- Comisión del mes = % del nivel × lo que la tienda paga por mes (el anual, dividido 12), redondeado al
  peso, durante los meses calendario 0 a 11 desde el alta (en hora de Argentina), sólo si la tienda
  está activa. Precios: `mercadopago.PrecioDe` en Go y `panel/planes.ts` en la web; misma tabla de
  casos (`partners/testdata/comisiones.json`).
- Estados de una tienda referida: en prueba (primeros 14 días, aunque haya elegido plan), activa (plan
  pago), sin plan pago (terminó la prueba en Gratis) y de baja (canceló o cerró). "Perdidas" en Inicio
  = sin plan pago + de baja.
- Nivel efectivo = el mayor entre el que dio el equipo y el que alcanzan las tiendas activas (Plata
  desde 5, Oro desde 20, como dice la landing). Directorio: Plata en adelante (lo publicado en la
  landing manda sobre la especificación, que decía sólo Oro).
- Pagos: se liquida a fin de mes y se paga el 10 del mes siguiente, contra la factura del partner (un
  link https, no se suben archivos). Si pasó el 10 y no hay factura, se muestra "se paga cuando subas
  tu factura". Mientras está pendiente, la liquidación se puede rehacer; aprobada o pagada, no.
- Aprobar, liquidar y marcar pagos lo hará el backoffice: por ahora son rutas internas con la clave del
  BFF (`/v1/interno/partners*`). Para probar en local: `mostradorctl partner-demo <email>` (sólo
  desarrollo, como superusuario) siembra la misma demo que el sandbox.
- Clics: la primera visita a `/empezar?ref=…` (sin origen previo) suma uno, después de responder. En la
  demo los clics son de ejemplo.
- El link de referido usa `SITE` (la dirección que funciona hoy); los QR de Imprimibles con `?ref=` pasan
  por `/empezar` en `mostradorapp.com.ar`, como el resto de las piezas impresas.
- Alta de una tienda para una clienta: arma el mail (mailto) con el link o abre el link para crearla.
  Pendiente: transferir la titularidad de una tienda (hoy la dueña suma al partner al equipo) y entrar
  al panel de una tienda donde se colabora (no hay cambio de tienda en la sesión todavía).
- Sin diseño en el handoff: el panel usa las piezas del panel de la tienda (barra de arriba, pestañas y
  tarjetas de Configuración, KPIs de Cupones, filas de MotorLista, Hoja, Modal, Segmentado, Alerta,
  Vacío con Hornero) y la tarjeta del directorio de /partners para la vista previa del perfil.


## 27. Backoffice del equipo (/staff)

- **Roles y permisos:** Soporte (lee tiendas, notas, extiende pruebas, regala packs, responde consultas), Administración (además plan, crédito, pausar, partners, facturación y auditoría) y Dueño (todo, incluido cerrar). La matriz vive en Go (`equipo.Permisos`), en SQL (`equipo.permisos`) y en la web (`lib/staff/permisos.ts`); los tests comparan las tres contra `apps/api/internal/dominio/equipo/testdata/permisos.json`.
- **Base de datos:** la API entra al backoffice con el rol `mostrador_staff` (pool aparte, `MOSTRADOR_BASE_STAFF`): lee todas las tiendas por políticas propias y sólo escribe en el esquema `equipo` (sesiones, notas, auditoría, respuestas). Los cambios a una tienda pasan por `equipo.accion_tienda` (SECURITY DEFINER), que chequea el permiso y anota en las dos auditorías en la misma transacción. RLS no se saltea en ningún caso.
- **Sesiones separadas:** la del equipo tiene su cookie (`__Host-mostrador_equipo`) y su tabla; un token de tienda no entra a `/v1/equipo/*` ni al revés. Las altas las hace `mostradorctl equipo-alta <email> <nombre> <rol>` con el migrador.
- **Ver la tienda como la dueña:** crea una mirada de 30 minutos con un token de sólo lectura; el panel muestra un aviso fijo y Go rechaza cualquier escritura (`equipo.SoloLectura`). Entrar a ver queda en la auditoría.
- **Demo:** `/staff` muestra un backoffice de ejemplo con datos inventados (no toca los sandboxes de los visitantes), con selector de rol para ver cómo cambia la navegación; las acciones responden "en la demo no se aplica".
- **Partners:** aprobar o rechazar una postulación queda acá; el alta del partner con nivel, link y cupón la hace el panel de partners (PR #75) cuando entren los dos.


## 29. Migración desde otras plataformas: las bases (migración 00041, sandbox 0007)

- Cuatro tablas con `app.aplicar_rls`: `clientes` (la ficha propia, origen `panel` o
  `importado:<plataforma>`, única por tienda y `lower(email)`), `historial_pedidos` (pedidos viejos,
  únicos por tienda, plataforma y externo), `redirecciones` (PK tienda y `desde`) e `importaciones`
  (una sola activa por tienda, índice parcial). La credencial de la plataforma vieja va cifrada con
  `plataforma/cifrado` (contexto `importacion:<tienda>`) y se borra al terminar, fallar o cancelar.
- **El historial no es un pedido de Mostrador**: no entra en Pedidos ni en el Resumen, no toca stock,
  facturas ni numeración. Sí cuenta para la lista y la ficha del cliente.
- **Avance por el outbox, sin polling**: crear inserta la importación y el evento `importacion.paso`
  (`{id, paso}`) en la misma transacción. El paso lee una página de la `Fuente` fuera de la
  transacción (es red), y en otra transacción bloquea la fila, verifica que `pasos` siga igual y que
  siga activa, guarda la página, avanza cursor y etapa (productos → clientes → pedidos →
  redirecciones) y encola el paso siguiente. Un evento entregado dos veces o el de una importación
  cancelada no hace nada. Un error de la fuente vuelve al outbox (espera creciente); en el último
  intento, o si envuelve `ErrDefinitivo`, la importación queda fallida con el error. Tope de 20.000
  pasos por si una fuente repite el cursor.
- Cada ítem va en su savepoint: el que no pasa las validaciones queda en `errores` (los últimos 50) y
  el resto entra. Los productos se guardan con `listas.GuardarEnTx` (lo mismo que el editor, sin armar
  la ficha que lee todo el catálogo); su URL vieja queda redirigida a `/producto/<id>`. Repetidos:
  URL ya redirigida o SKU que ya existe (producto), mismo email o mismo externo sin email (cliente),
  mismo externo (pedido). El upsert de clientes sólo completa lo que falta, nunca pisa nombre ni email.
- **Clientes**: la lista une por email en minúsculas los derivados de pedidos, las fichas (aunque no
  tengan compras) y el historial (suma pedidos, gastado, última compra, meses y segmento). El id de
  la fila es el email en minúsculas; sin email, el nombre (pedidos) o el id de la ficha. El alta de
  Clientes del panel ahora guarda una ficha en Go (antes no hacía nada).
- **Redirecciones**: el path viejo se normaliza igual en Go y en TS (sin dominio, query ni barra
  final, decodificado, en minúsculas; `vidriera/testdata/rutas.json`). `hacia` es un path relativo a la
  tienda y nunca `//…` ni `/\…`. El hook de SvelteKit, sólo ante un 404 de un GET en `/t/<slug>/…`,
  pregunta a Go (`GET /v1/publico/tiendas/{slug}/redireccion?path=`, 301 con Location) o a la D1 del
  sandbox, y responde 301 a `/t/<slug><hacia>`. Las páginas que existen no consultan nada.
- **Demo**: mismas tablas en D1 con `sandbox_id` y `ON DELETE CASCADE` (la purga del cron se las lleva
  con el sandbox). Sin outbox, el navegador pide "avanzar" (`POST /panel/importaciones`) y cada pedido
  hace un paso; `pasos` hace de candado con un UPDATE condicional. Un error de la fuente la deja
  fallida (no hay reintentos con espera). La credencial va cifrada con `ENCRYPTION_KEY` (sandbox
  0008, `importaciones.credencial`) y se borra al terminar, fallar o cancelar, como en Go.
- **Tiendanube con OAuth real, también en la demo** (migración 00042): `GET /panel/importar/tiendanube`
  arma un state de un solo uso (10 min, sólo su hash: `importacion_estados` en Go, `cambios`
  `tn-estado` en el sandbox) y manda a autorizar la app de socio; la vuelta
  (`/panel/importar/tiendanube/vuelta`, la URL registrada en la app) consume el state, canjea el código
  y crea la importación con `{token, tienda}` cifrado. Con la API, la web sólo le pasa código y state
  a Go (`POST /v1/panel/importaciones/tiendanube/vuelta`): el token nunca pasa por el Worker. En la
  demo canjea el Worker con `TIENDANUBE_CLIENT_ID` y `TIENDANUBE_SECRETO`. Sin la app configurada,
  el botón dice que no está disponible.
- **Fuentes de la demo** (`importacion/fuentes.ts`): Tiendanube, Shopify y WooCommerce, a la medida de
  un pedido del Worker: productos de a 5 (`POR_PAGINA`, cada uno se guarda como en el editor),
  clientes y pedidos de a 50 (un `json_each` por página), una página a la red por paso. Topes de la
  demo: 500 productos y 1000 clientes y pedidos. Un 429 de Tiendanube espera hasta 3 s y reintenta dos
  veces; si sigue, el paso devuelve el mismo cursor y el próximo lo reintenta. El e2e usa un
  Tiendanube falso en localhost (`TIENDANUBE_LOCAL`, sólo se acepta `http://localhost|127.0.0.1`).
- Las fuentes reales se registran en `importacion.Fuentes()` (Go) y `FUENTES` (TS); mientras falte la
  de una plataforma, crear responde 409 `sin_fuente`.

## 30. Correo Argentino con la API MiCorreo (migración 00047)

- **Cuenta de la tienda, credenciales de la plataforma.** Como el plugin oficial de WooCommerce: la
  tienda conecta con el email y la contraseña de su MiCorreo (`/users/validate`) y se guarda sólo el
  `customerId`; la contraseña no. El JWT sale de las credenciales de API de Tu Local Web
  (`MOSTRADOR_CA_USUARIO`/`MOSTRADOR_CA_CLAVE`, por ambiente) o de las de la tienda, cifradas con
  `plataforma/cifrado` (contexto `correo:<tienda>`). Detalle de la API en
  `docs/integraciones/correo-argentino.md`.
- **Nunca un precio inventado.** La tabla por región que cotizaba "Correo a sucursal" se borró (Go y
  TS). Correo se ofrece sólo con la cuenta conectada (`envios.correo` = hay fila en `ca_conexiones`) y
  con la cotización de `/rates` para ese CP; mientras no llega, o si falla, no aparece. Al confirmar,
  Go vuelve a cotizar: el precio nunca viene del navegador, y si Correo no responde la compra falla
  con 409 `sin_cotizacion`. La demo no tiene cuenta de MiCorreo: Correo no se ofrece ni se conecta.
- **Dos opciones, un envío.** El checkout ofrece `correo-domicilio` y `correo-sucursal`; el pedido
  queda con `envio = 'correo'` y la modalidad, provincia, localidad y sucursal en `pedidos.entrega`.
  La sucursal se valida contra `/agencies` (activas y con retiro).
- **Peso por unidad.** Los productos no tienen peso: la tienda carga un paquete típico (gramos por
  unidad y la caja). Peso = unidades × peso, con el tope de 25 kg de MiCorreo.
- **Alta idempotente, seguimiento a mano.** "Generar envío" manda `/shipping/import` con
  `extOrderId = TLW-<tienda>-<pedido>`; si MiCorreo dice que ya fue importada, se toma como hecha
  (como la recuperación de ARCA). MiCorreo no expone el seguimiento por API: el número llega al pagar
  el rótulo allá y la tienda lo pega en el pedido (`ca_envios` y `pedidos.seguimiento`). El
  seguimiento por API existe en Paq.ar, que pide acuerdo comercial.
- **Sin verificar todavía contra la API real**: sólo contra el MiCorreo falso de los tests. Se cierra
  con credenciales de apitest y una cuenta de MiCorreo de prueba.

## 31. Google y Apple por Clerk, sumado al ingreso propio (migración 00048)

- **Aditivo**: email y contraseña siguen igual (Go). Google, Apple y los demás sociales pasan por
  Clerk, una sola aplicación para todo Grupo Vimo (también la app de iPhone). Mercado Pago y Mercado
  Libre siguen con el OAuth propio de §18; el OAuth propio de Google se sacó (sus identidades se
  borran: la cuenta se vuelve a vincular sola por el email verificado).
- **Sin cookies de Clerk en tulocalweb.com.ar**: la instancia vive en grupovimo.com.ar y compartir
  sesión con otro dominio pide dominios satélite (pagos). En cambio la web habla con la Frontend API en
  modo nativo, como la app de iPhone (`lib/clerk.ts`): el token del cliente va en `Authorization` y
  sólo dura el ida y vuelta en sessionStorage. Para eso la instancia tiene que tener la Native API
  prendida, `https://tulocalweb.com.ar` en `allowed_origins` y `/ingresar/clerk` en las redirect URLs
  (sin eso Clerk no devuelve el `rotating_token_nonce`). No hace falta CSP nueva: no se carga ningún
  script ni iframe de Clerk, y las páginas fuera de `/t/` no tienen CSP.
- **Go no le cree al navegador**: `POST /v1/auth/clerk` recibe el token de sesión de Clerk, verifica
  firma (JWKS bajado una vez con la clave secreta), emisor (sale de la clave publicable), vencimiento y
  `azp` (si viene, tiene que ser un origen permitido; la app nativa no lo manda) y lee email
  verificado, nombre y foto con la Backend API. De ahí sigue el camino de §18: identidad vinculada
  (`auth.identidades`, proveedor `clerk`), vincular por email o cuenta nueva con onboarding y mail de
  bienvenida. La sesión que vuelve es la propia (cookie `__Host-mostrador_sesion`); la de Clerk se
  cierra apenas se tiene el token.
- La foto del proveedor queda en `auth.usuarios.foto` (sólo https) y sale en `GET /v1/auth/sesion`.
- En la demo los botones llevan al recorrido de ejemplo, como antes; WhatsApp sigue sólo en la demo.

## 32. Envíos y retiro configurables (migración 00049)

- **Una fila por tienda en `entregas`**: retiro (sí/no, dirección y horario en texto) y envío propio
  en moto (sí/no, costo fijo en pesos, zona y demora en texto). Sin fila, retiro en el local y sin
  moto, como antes. Dirección y horario vacíos: los del negocio (se resuelve al leer, en Go y en el
  sandbox). Correo Argentino no va acá: se ofrece si hay `ca_conexiones`.
- **Al menos una forma de entrega**: retiro, moto o Correo conectado. La moto pide la zona y el
  retiro, alguna dirección. Mismos mensajes en Go (`configuracion/entregas.go`) y en la web
  (`lib/panel/entregas.ts`, que usa el sandbox).
- **La zona de la moto es texto**: no se cruza con el CP (no hay de dónde sacar barrios). Se ofrece
  con cualquier CP válido y quien compra ve la zona antes de elegirla. El envío gratis de la
  promoción automática vale también para la moto. Un pedido con moto sigue el flujo de envío.
- **Nada prometido que la tienda no configuró**: se sacaron «Moto en el día en CABA» y «antes de las
  16 h» de las plantillas; la portada muestra el costo y la zona reales y el pie, «Envío en moto».
- Pantalla en `/panel/envios/configurar` (desde Envíos, Configuración y el onboarding). En la demo,
  el envío gratis sale de la automática (40.000, como en Go) y la moto arranca apagada.

## 33. Sesiones: dónde está abierta la cuenta, aviso de ingreso e inactivas (migración 00051)

- **Lugar real o nada**: la IP de cada sesión se ubica con DB-IP City Lite (mmdb, CC BY 4.0;
  `MOSTRADOR_GEOIP`, abierta una vez al arrancar, `plataforma/geoip`). En Argentina «Ciudad,
  Provincia» (sin el barrio entre paréntesis, que no es confiable; CABA como «Ciudad de Buenos
  Aires»); afuera, el país en castellano. Sin base, o con una IP privada o de loopback, el lugar
  queda vacío y la pantalla no lo muestra.
- **La IP de lo que llega directo**: la app de iPhone le pega a `api.tulocalweb.com.ar` sin pasar por
  el BFF, y Go veía la conexión de Caddy (`::1`). Si la conexión viene de loopback o de una red
  privada (Caddy, Docker), Go toma `Cf-Connecting-Ip`, que Cloudflare pisa siempre (el VPS sólo
  acepta 80/443 desde Cloudflare). Lo del BFF sigue igual: `X-Cliente-IP` sólo con la clave.
- **Opciones de la persona, no de la tienda**: las sesiones son de la persona, así que la seguridad
  vive en `auth.usuarios.seguridad`: `avisoIngreso` (prendido por defecto) e `inactivas`. Se
  sacaron «doble verificación», «otro país» y «alertas de cambios sensibles»: no existían.
- **«Entraron a tu cuenta»**: todo ingreso a una cuenta que ya existe (clave, Google y Apple por
  Clerk, Mercado Pago, link por WhatsApp) pasa por `sesionDeUsuario`, que encola `cuenta.ingreso`
  en la misma transacción. El registro y la cuenta nueva por un proveedor no avisan (ya sale la
  bienvenida). El evento va sin tienda (ninguna tienda ve las IP de quien entra; la app sólo puede
  insertar `cuenta.*` sin tienda) y el worker manda el mail con la plantilla de cuentas: dispositivo,
  lugar e IP si se saben, hora de Argentina y link a Configuración › Seguridad.
- **Inactivas de verdad**: con la opción prendida, una sesión cuyo `visto_en` tiene más de 7 días
  no vale y se borra al usarla; al prender la opción se cierran en el momento las que ya cumplen.
  Importa porque «Recordarme» ahora dura un año.
- La lista muestra todas las sesiones vigentes (la actual primero), con «Cerrar» y «Cerrar todas
  las demás» (`DELETE /v1/panel/sesiones`).

## 34. La vista previa al compartir una tienda (sin migración)

- **Imagen armada por el servidor, una por tienda y por producto**: `/t/<slug>/compartir/<v>.jpg` y
  `/t/<slug>/producto/<slug>/compartir/<v>.jpg`, 1200 × 630 en JPEG de menos de 300 KB. satori
  arma un SVG con las letras como trazos (el contenedor no necesita fuentes instaladas) y vips, el
  mismo que achica las fotos, lo pasa a JPEG. Las fuentes (WOFF latinos, OFL) van en
  `lib/server/compartir/fuentes` y se leen con `read()` del build. Nada nuevo en la imagen Docker.
- **El estilo de la plantilla**: colores, tipografía, radios y el acento elegido; tres composiciones
  (editorial, grilla, catálogo) con el adorno de cada una. Sin fotos queda armada igual: el nombre
  grande, la inicial en el acento y, en las de catálogo, las categorías con cuántos productos tienen.
  Sólo datos de la tienda: nunca la imagen genérica de Tu Local Web ni algo inventado.
- **WhatsApp guarda por URL**: `<v>` es un hash de lo que se dibuja (`versionDe`); al editar la
  tienda cambia la URL. Con la versión vigente se cachea un año (`immutable`); con una vieja, cinco
  minutos. En el servidor, memoria más disco (la carpeta temporal del contenedor). og:image siempre
  en tulocalweb.com.ar, también para las tiendas con dominio propio.
- **Textos**: título «Nombre · lema» (o lo que vende), descripción con datos reales (cantidad de
  productos, categorías, entrega, Mercado Pago) que nunca repite el título (`lib/tienda/compartir.ts`).
- **Editable en Mi tienda › Al compartir el link**: título, descripción y la imagen (la de la
  plantilla o una foto propia, recortada a 1200 × 630). Va en el diseño (`compartirTitulo`,
  `compartirDescripcion`, `compartirFoto` en `diseno_borrador`/`diseno_publicado`): sin migración, y
  se ve al publicar. La vista previa del editor usa el borrador (`/panel/mi-tienda/compartir`).

## 35. «Nueva venta #N» por mail a la dueña (migración 00052)

- **Cuándo**: el trigger `app.avisar_venta_tienda` encola `tienda.venta` en la misma transacción que
  el pedido, sólo para la tienda online (`canal = 'web'`). Por transferencia, al crearse (el mail dice
  que falta el comprobante); por Mercado Pago, cuando `estado_pago` pasa a acreditado, nunca antes.
  Marcar pagada una transferencia no manda otro: lo hace la misma dueña.
- **Una sola vez**: `pedidos.avisos->'tienda_venta'`, como los mails a quien compra.
- **A quién**: el email de Configuración › Negocio o, si no hay, el de la cuenta dueña (lo mismo que el
  Reply-To de los mails a quien compra). La respuesta le llega a quien compró.
- **Qué**: lo vendido, subtotal/envío/descuentos/total, medio de pago con su estado, la entrega,
  nombre, email, teléfono y nota de quien compró, y «Ver el pedido» al panel. Plantilla de los mails
  de pedidos.
- **Interruptor**: Notificaciones › Venta nueva › Email (`preferencias.canales.matriz["venta:email"]`),
  prendido por defecto; el worker lo mira al mandar. Sin columna nueva.

## 36. Visitas propias de la tienda, sin cookies (migración 00055)

- **Quién cuenta**: el servidor de la web, sin scripts de terceros. La carga completa de una página
  de la tienda (inicio, catálogo, producto, carrito, checkout) la cuenta `hooks.server.ts` al
  responder 200 a un pedido de documento; la navegación dentro de la tienda (que sólo pide
  `__data.json`, igual que el preload por hover) la avisa el layout de `/t/[tienda]` con
  `sendBeacon` a `/t/<slug>/vista` desde `afterNavigate`, que no corre en los preload ni al
  revalidar. Un evento por página vista.
- **Visitante del día**: sha256(sal + IP + navegador + tienda), 16 bytes. La sal es al azar, vive en
  la memoria del proceso y rota con el día de Argentina: nunca se guarda, así que no hay forma de
  volver a la persona ni de unir días. Por eso «visitas» de 7 o 30 días es la suma de las personas
  distintas de cada día. Si el proceso se reinicia en el día, alguien puede contar dos veces: se
  acepta. No hay cookies, así que no va en el aviso de cookies.
- **No cuentan**: user agents de bots y de vistas previas (WhatsApp, Facebook, Telegram, Slack…),
  Lighthouse y navegadores sin cabeza, clientes HTTP de programa, prefetch/prerender
  (`Sec-Purpose`/`Purpose`), la vista del equipo (`mirada`) y la dueña o su equipo con la sesión
  del panel en su propia tienda (la web pregunta a Go de qué tienda es la sesión, en segundo plano y
  con caché de 10 min; si no lo sabe, no cuenta). En un dominio propio no hay cookie del panel: ahí
  la dueña sí cuenta.
- **Sin latencia ni caídas**: `Medidor` (`lib/server/visitas.ts`) junta en memoria y manda lotes
  cada 5 s a `POST /v1/interno/visitas` (sólo con la clave del BFF) con timeout de 3 s; si la API
  no responde, el lote se pierde. Topes: 120 vistas por visitante y 600 por IP por hora, cola de
  10.000 y mapas acotados.
- **Qué se guarda** (Go, paquete `visitas`, con RLS): `visitas` = una fila por tienda, día, página
  y producto con vistas y visitantes (`pagina = 'tienda'` es el total del día); `visitas_vistos`
  (UNLOGGED) = los hashes del día para contar únicos, que el worker borra cada hora si son de días
  anteriores (`app.purgar_visitas`), junto con las visitas de más de 400 días. Sólo cuentan tiendas
  publicadas y productos que existen (por slug).
- **Dónde se ve**: el Resumen (visitas de hoy contra ayer, 7 y 30 días contra los anteriores,
  páginas, más vistos) y la conversión real = pedidos web no cancelados / visitas, por días; sin
  visitas en el período anterior no se muestra variación. Informes suma la pestaña «Visitas» por mes
  desde que se mide. Tincho recibe las visitas en su contexto (`ContextoHornero.visitas`). Sin
  ninguna visita, el panel lo dice y la conversión queda en «—». La demo sigue con su «Ahora en tu
  tienda» de ejemplo.

## 37. El equipo de cada tienda: roles fijos aplicados en Go e invitaciones por mail (migración 00054)

- **Roles fijos en vez de la matriz editable**: la matriz de Usuarios y permisos se guardaba pero
  Go no la aplicaba (cualquiera del equipo veía y hacía todo). Se reemplazó por cuatro roles claros
  que Go aplica en cada ruta: **Dueña** (todo), **Administración** (todo menos el plan, la
  facturación de Tu Local Web y el equipo), **Ventas** (pedidos, clientes y productos, sin precios,
  costos ni finanzas) y **Local** (punto de venta, pedidos y ver productos). «Depósito» pasó a
  «Local»; la columna `tiendas.permisos` queda sin uso. La tabla rol → permisos es una sola:
  `roles.porRol` en Go y `lib/panel/permisos.ts` en la web, comparadas por test contra
  `apps/api/internal/dominio/roles/testdata/roles.json`.
- **Go decide, la web oculta**: `roles.Guardia` resuelve cada pedido de `/v1/panel/` con un
  `ServeMux` propio cargado con `roles.Rutas` (el mismo patrón con que se registra la ruta) y responde
  403 si el rol no tiene el permiso. Una ruta que no está en la tabla sólo la usa la dueña (cerrado por
  defecto) y un test recorre el código y falla si una ruta del panel no tiene permiso. Lo que depende
  del cuerpo va en el handler: Ventas no ve `costo`/`margen` en productos e inventario, y al editar un
  producto Go conserva precio, costo y precios de variantes; la acción «precio» de la lista pide
  Catálogo. `GET /v1/panel/configuracion` no le manda los datos del negocio (alias, CBU) a quien no
  tiene Tienda ni el equipo a quien no es dueña. La seguridad y las sesiones son de cada persona: las
  ve cualquiera. Las IA conectadas heredan el rol de quien las conectó.
- **El Inicio sin plata para Ventas y Local** (31/10, sin migración): quien no tiene Finanzas no
  recibe montos en `GET /v1/panel/tablero` (ventas, serie, ticket, canales, meta, caja,
  liquidaciones, montos de transferencias y carritos, total de los últimos pedidos ni lo cobrado por
  Mercado Pago) ni las ventas en pesos del contexto de Tincho: se omiten, no van en cero, y
  `montos: false` lo dice. Go lo decide con el rol del autor (`tiendas.ConMontos`); el sandbox hace
  lo mismo (`datos/montos.ts`). La web saca esas tarjetas y Tincho (motor, modelo y MCP) responde
  con pedidos y unidades.
- **Invitar sólo por email**: el link (`/invitacion?token=…`) es de un solo uso y vence a los 7 días;
  el sha256 del token vive en `auth.invitaciones` (fuera de RLS: quien abre el link todavía no es de
  la tienda; de ahí sale la tienda y recién entonces se lee la invitación con la tienda fijada). El
  mail sale por el outbox (`cuenta.invitacion`, sin tienda, plantilla de los mails de cuentas). Hasta
  30 mails de invitación por tienda por hora. Reenviar arma un link nuevo y anula el anterior;
  cancelar borra la invitación y su link. Las invitaciones viejas por WhatsApp quedan «sin mandar»:
  se cancelan e invitan por email.
- **Aceptar**: con sesión, «Sumarme» (si ya era del equipo, conserva su rol) y la sesión pasa a esa
  tienda. Sin sesión: si el email ya tiene cuenta, ingresa (clave, Google o Apple) y vuelve al link;
  si no, crea la cuenta ahí con el email de la invitación (queda verificado) o con Google/Apple, y
  **no se crea otra tienda**. Para Google y Apple el token viaja en la cookie corta `tlw_invitacion`
  hasta `/ingresar/clerk`, que lo manda a `POST /v1/auth/clerk` (`invitacion`).
- **Quitar y cambiar rol**: quitar borra la membresía y cierra las sesiones de esa persona en esa
  tienda (su cuenta y sus otras tiendas siguen). A la dueña no se la quita ni se la degrada, nadie
  pasa a dueña y nadie se cambia a sí misma.
- **Varias tiendas**: `GET /v1/auth/tiendas` da la tienda actual, el rol, sus permisos y todas las
  tiendas de la persona (`app.tiendas_de_usuario`, que fija cada tienda de sus membresías para leerla
  con RLS forzada); `PUT /v1/auth/tienda` cambia la de la sesión. En la barra lateral aparece un
  selector simple cuando hay más de una.

## 38. La dirección de la tienda se cambia y la vieja anda para siempre (migración 00056)

- **Dónde**: al registrarse con Google, Apple o Mercado Pago la tienda nace como «La tienda de
  <nombre>». El paso «Tu negocio» del onboarding (con la API) pregunta «¿Cómo se llama tu tienda?»
  con ese nombre como sugerencia que invita a cambiarlo, y debajo la dirección
  `tulocalweb.com.ar/t/<slug>`, que se arma sola con el nombre hasta que se la toca. Se guardan al
  seguir (primero la dirección, que es lo que puede fallar). Después, en Configuración › Negocio ›
  Dirección de tu tienda, con «Deshacer» en el aviso. `CampoDireccion.svelte` es el mismo campo en
  los dos lados y pregunta la disponibilidad mientras se escribe (350 ms de pausa).
- **Regla**: de 2 a 60 letras minúsculas, números o guiones, sin guion al principio ni al final (el
  CHECK de `tiendas.slug` pasó de 48 a 60). Los reservados (rutas, marcas y las vitrinas de las
  plantillas) están en `packages/contrato/direccion.json` con sus casos; Go (`tiendas.Reservados`) y
  la web (`lib/tienda/direccion.ts`) se comparan con ese archivo en sus tests.
- **Una dirección no vuelve a estar libre**: `app.slugs` ya registraba cada slug tomado; el viejo se
  queda ahí, así que ninguna otra tienda puede tomarlo. `app.slugs_anteriores` dice de qué tienda
  era. La misma tienda sí puede volver a una suya (aunque hoy esté reservada: las tiendas de la demo
  arrancan con la de una vitrina). Las tablas viven en `app`, sin permisos para la app; sólo las
  tocan `app.estado_slug`, `app.cambiar_slug` (decide la tienda con `app.tienda_actual()`, bloquea su
  fila y deja que `app.slugs` resuelva quién gana si dos piden la misma) y `app.slug_vigente`.
- **Tope**: 5 cambios por tienda en 24 h (`app.slugs_cambios`, que además deja el historial).
- **Redirección**: ante un 404 de un GET en `/t/<slug>/…`, el hook pregunta primero
  `GET /v1/publico/slugs/{slug}` y, si fue de una tienda, responde 301 a `/t/<nuevo>/…` con la misma
  consulta (las vueltas de Mercado Pago de pedidos ya hechos traen la suya). Cubre la portada,
  producto, catálogo, carrito, pedido, la imagen para compartir y el feed. Recién si no, mira las
  redirecciones de la tienda importada. Las páginas que existen no consultan nada.
- **Lo que depende del slug** se arma siempre con el vigente: sitemap, og:image, QR del onboarding y
  del panel, mails y back_urls de Mercado Pago nuevas. Lo ya mandado con el viejo cae en la
  redirección. Los dominios propios van por el id de la tienda; si un isolate todavía tenía guardado
  el slug viejo de un dominio (5 min), lo olvida y responde 307 a la misma URL.
- **Demo**: el sandbox guarda la dirección como cambio de la tienda (`tienda/direccion` con las
  anteriores y los cambios del día); la vieja redirige sólo para el visitante dueño de ese sandbox.

## 39. Actividad: quién del equipo hizo qué (migración 00057)

Con equipo en las tiendas (§37), la dueña necesita saber quién hizo cada cosa. Antes la tabla
`auditoria` sólo la escribían las acciones del soporte de Tu Local Web.

- **Se anota en Go, donde se hace la acción**, en la misma transacción
  (`auditoria.Registrar(ctx, tx, tienda, registros...)`): si no se puede anotar, el cambio tampoco
  queda. Nunca desde la web.
- **El autor sale del contexto**: después de resolver la sesión, `servidor.conActor` deja en el
  contexto la persona (id, nombre y rol de ese momento, IP y agente) o la IA conectada (su nombre y
  quién la conectó), sólo para `/v1/panel/`. Lo que corre sin autor (worker, webhooks, importación,
  la tienda online) no se anota. Una mirada del soporte es de sólo lectura y no lleva autor.
  `auditoria.Silenciar` evita doble registro cuando una acción reusa otra (duplicar no es «creó»).
- **Una tabla, la de siempre**: `auditoria` (no `equipo.auditoria`, que es la bitácora interna del
  backoffice). La 00057 suma `actor_nombre`, `actor_rol`, `actor_ia`, `texto` (la frase sin el
  sujeto, ya escrita en castellano) y `actor_tipo = 'ia'`. El tipo (filtro) es el prefijo de
  `accion`: producto, pedido, venta (punto de venta), promocion, cobros, negocio, envios, dominio,
  integracion, personas, plan, datos, sesion, mitienda; lo del soporte (`actor_tipo = 'staff'`) es
  «soporte» y se lee con frases fijas.
- **Qué se anota**: productos (crear, precio/costo/stock con antes → después, nombre, publicar,
  pausar, borrar, duplicar, aumentos), pedidos (estado, marcar pagado y deshacerlo, cancelar, cargar a
  mano), cobros del punto de venta, cupones y promociones automáticas, alias/CBU/titular, nombre y
  dirección de la tienda, envíos y retiro, dominios, Mercado Pago, Mercado Libre, Correo y ARCA,
  equipo (invitar, reenviar, cambiar rol, quitar, cancelar), plan y packs, exportaciones (también la
  de la Actividad), cierre de sesiones y publicar el diseño de Mi tienda.
- **Nada sensible entero**: ni claves ni tokens; del CBU sólo los últimos 4 (`auditoria.Final`).
- **La ven la dueña y administración** (`roles.Tienda` en `GET /v1/panel/actividad` y `/exportar`);
  ventas y local reciben 403 y no la ven en el menú. RLS por tienda como el resto.
- **Lectura**: de a 50, del más nuevo al más viejo, con cursor (creado_en, id); filtros por persona,
  tipo y fecha (días de Argentina, inclusive). CSV con los mismos filtros (hasta 20.000 filas, sin
  fórmulas: lo que empieza con `= + - @` va con comilla).
- **Retención: un año.** `app.purgar_auditoria()` (SECURITY DEFINER, como `app.purgar_visitas`) la
  corre el worker una vez por día. La app sigue sin poder borrar ni reescribir registros.
- Lo anterior a esta migración no tiene autor ni frase: la pantalla lo dice («Lo de antes no se
  guardó»).

## 40. Push en el iPhone: dispositivos por sesión y APNs desde el worker (migración 00058)

El lado servidor de las notificaciones de la app de iPhone (docs/handoff/app-ios). La app todavía no
existe; esto deja listo lo que va a usar.

- **Registro**: la app, con su token de sesión y directo contra `api.tulocalweb.com.ar`, manda
  `POST /v1/panel/dispositivos` (token de APNs en hexadecimal, `entorno` sandbox/produccion, modelo y
  versión) y `DELETE /v1/panel/dispositivos` al cerrar sesión. Las dos son `roles.Panel` (cualquiera
  del equipo) y sólo con la sesión de una persona: una IA conectada o el soporte mirando reciben 403.
- **Atado a la sesión y a la persona**: `auth.dispositivos` guarda el hash de la sesión con
  `ON DELETE CASCADE`, así que salir, «Cerrar», contraseña nueva, quitar del equipo o una inactiva
  se llevan el iPhone; una sesión vencida queda pero el worker no la elige. El token es único: si
  vuelve a llegar con otra sesión (otra cuenta en el mismo iPhone) se mueve. Vive en `auth` con RLS
  forzada por persona (`db.EnUsuario`), como las fichas de partner; la app no inserta directo: pasa por
  `app.registrar_dispositivo`, que exige que la sesión sea de esa persona y esté vigente.
- **Envío**: APNs por HTTP/2 con autenticación por token (`plataforma/apns`, sólo biblioteca
  estándar): JWT ES256 firmado con el .p8, reusado 40 minutos y firmado de nuevo si APNs responde
  `ExpiredProviderToken`. Config: `MOSTRADOR_APNS_KEY_ID`, `MOSTRADOR_APNS_TEAM_ID`,
  `MOSTRADOR_APNS_CLAVE` (ruta al .p8, nunca en el repo) y `MOSTRADOR_APNS_TOPIC`
  (`ar.com.tulocalweb.app` por defecto). Sin eso, el worker lo dice una vez en el log y los
  manejadores no mandan nada. El host sale del `entorno` de cada iPhone (sandbox para lo instalado
  desde Xcode).
- **Tokens muertos**: 410, `BadDeviceToken` o `Unregistered` borran el token (`app.olvidar_dispositivo`).
  429 y 5xx hacen reintentar el evento; `apns-collapse-id` evita que un reintento se vea dos veces.
- **Qué se manda**, por el outbox (un evento aparte por canal, para que cada uno se reintente solo):
  - `push.venta`: «Nueva venta #N», en el mismo momento que el mail `tienda.venta` (el trigger de la
    00052 ahora encola los dos). Una sola vez (`pedidos.avisos->'push_venta'`), a quien tiene Pedidos.
    El monto sólo a quien tiene Finanzas (§37); nunca el nombre de quien compró.
  - `push.pregunta_ml`: «Pregunta nueva en Mercado Libre», cuando llega una pregunta nueva sin
    responder; a quien tiene Clientes, con el producto pero sin el texto de la pregunta.
  - `cuenta.ingreso_push`: «Entraron a tu cuenta», con el mismo interruptor que el mail
    (`avisoIngreso`), sólo a los iPhone de esa persona y no al de la sesión nueva. Con el aparato, sin
    IP ni lugar (se ve en la pantalla bloqueada).
- **A quién**: a los iPhone de las personas del equipo de la tienda con sesión vigente
  (`app.dispositivos_de_tienda`), aunque esa sesión esté parada en otra de sus tiendas.
- **Payload**: `aps.alert` (title/body en castellano), `sound: default`, `thread-id` por tienda
  (`tienda-<id>`; `cuenta` para lo de la persona) y, afuera de `aps`, `ruta` (`pedido/<numero>`,
  `preguntas/ml/<id>`, `cuenta/seguridad`) y `tienda` para abrir la pantalla.
- **Preferencia**: Notificaciones › «Push en el iPhone» es `venta:celular` de la matriz (prendido por
  defecto); el worker la mira al mandar. Sin columna nueva.

## 41. «Te quedaste sin stock»: mail y push al llegar a 0 (migración 00059)

- **Cuándo**: una venta deja en 0 un producto publicado o una de sus variantes activas. Toda venta
  pasa por `pedidos.CrearDesdeProductos` (tienda online, punto de venta, Crear pedido y Mercado
  Libre), así que el aviso se encola ahí, en Go y en la misma transacción que descuenta: si la venta
  no se hace, tampoco el aviso. Bajar el stock a mano (Inventario, editor) no avisa: lo hizo la misma
  persona. Un producto pausado no avisa.
- **Uno por producto por día** (de Argentina): `productos.sin_stock_avisado` guarda el día del último
  aviso; el UPDATE condicionado corre con la fila ya bloqueada por el descuento, así que de dos ventas
  a la vez avisa una. Dos variantes del mismo producto el mismo día: un aviso. Sin tabla aparte, nada
  que purgar.
- **Dos eventos** (`tienda.sin_stock` y `push.sin_stock`, datos `{producto, variante}`), uno por
  canal para que cada uno se reintente solo. Al mandar se vuelve a leer: si ya repusieron o lo
  pausaron, no sale.
- **A quién**: la dueña y quien tiene Productos (Dueña, Administración, Ventas; Local no). El mail, a
  la dueña en el email del negocio (o el de su cuenta si no hay) y al resto en el de su cuenta, uno
  por persona; si falla alguno pero ya salió otro, no se reintenta (no se le manda dos veces a quien
  ya lo tiene). El push, a los iPhone del equipo con Productos (`ruta: producto/<id>`).
- **Qué dice**: «Te quedaste sin stock», el producto (con la variante) y que en la tienda online
  figura «Sin stock». Botón «Reponer stock» a `/panel/productos?editar=<id>`, que abre el editor.
  Plantilla de los mails de la tienda.
- **Preferencia**: Notificaciones › «Sin stock» (`sinstock:email` y `sinstock:celular` de la matriz,
  prendidos por defecto); el worker la mira al mandar.

## 42. Compartir un producto (sin migración)

- **Dónde**: «Compartir» en cada fila de Productos y en el pie del editor; en la ficha pública, un
  botón chico debajo de comprar (las doce plantillas, `CompartirProducto.svelte`).
- **Cómo** (`lib/panel/compartir.ts`): en el celular, la hoja de compartir del sistema con
  «Nombre · $ precio — link» y la imagen del producto (`/t/<slug>/producto/<id>/compartir/0.jpg`, la
  de §34) como archivo si el navegador lo acepta; la imagen se espera 1,2 s como mucho (Safari sólo
  abre la hoja poco después del toque) y si no llega se comparte sin ella. En la compu, el link queda
  copiado y el aviso ofrece «Enviar por WhatsApp» (wa.me con el mismo texto); en la ficha pública,
  «Link copiado».
- **Sólo publicados**: uno pausado avisa «Primero publicalo», con el botón para publicarlo.
- **El link** es el slug guardado del producto (fijo desde que se creó, migración 00043): la lista y
  la ficha del panel ahora lo leen de la base en vez de armarlo del nombre.

## 43. El panel instalable como app y el push web (migración 00060)

Mientras no exista la app nativa, el panel se instala en la pantalla de inicio y recibe los mismos
avisos que el iPhone de §40.

- **Instalable sólo el panel**: el `<link rel="manifest">` (`static/panel/manifest.webmanifest`,
  «Tu Local Web», `start_url` y `scope` `/panel`, `standalone`), el apple-touch-icon y las metas de iOS
  van en el `<svelte:head>` del armazón del panel. La landing y las tiendas no llevan manifest: una
  tienda nunca se instala como Tu Local Web. Íconos (192, 512 «any» y «maskable», 180 para iOS e
  insignia monocroma de 96 para Android) salen del isotipo del handoff.
- **Service worker** (`static/sw-panel.js`, alcance `/panel`): no guarda nada del panel (pedidos,
  ventas y clientes van siempre a la red). Sólo guarda `panel/sin-conexion.html` (Tincho
  desenchufado, «Reintentar», se recarga sola al volver la red) y la muestra cuando una navegación
  del panel falla sin conexión.
- **Invitación** (`InstalarApp.svelte`, en el Inicio): sólo en el celular. iPhone: los pasos de
  Safari (Compartir › Agregar a inicio) con los íconos que se ven ahí; Android: botón que abre el
  diálogo del sistema (`beforeinstallprompt`, guardado temprano desde `app.html`). «Ahora no» la
  cierra para siempre en ese navegador; abierta desde el ícono (`display-mode: standalone`), no sale.
- **Push web** (Web Push estándar, sin Firebase): `plataforma/webpush` cifra con RFC 8291
  (aes128gcm, ECDH P-256, HKDF; verificado con el vector del apéndice A) y firma VAPID ES256, sólo
  con la biblioteca estándar. Config `MOSTRADOR_VAPID_PUBLICA`, `MOSTRADOR_VAPID_PRIVADA` (base64url,
  `go run ./cmd/vapid`) y `MOSTRADOR_VAPID_SUJETO` (`mailto:`); sin las tres, apagado (la API da la
  clave vacía y el panel dice «Todavía no está disponible»). Cambiar las claves invalida todas las
  suscripciones.
- **Atada a la sesión**: `auth.suscripciones_web`, hermana de `auth.dispositivos` (RLS por persona,
  `ON DELETE CASCADE` con la sesión, el endpoint único se mueve de sesión). Sólo endpoints de servicios
  de push conocidos (FCM, Apple, Mozilla, Windows): el worker les hace un POST.
  `GET/POST/DELETE /v1/panel/push-web` y `POST /v1/panel/push-web/prueba` (`roles.Panel`, sólo con
  la sesión de una persona).
- **Envío**: los mismos manejadores de §40 y §41 (`push.venta`, `push.pregunta_ml`,
  `push.sin_stock`, `cuenta.ingreso_push`) mandan por APNs y por Web Push con las mismas reglas de
  rol y montos y la misma preferencia (`venta:celular`). 404/410 borran la suscripción; 429 y 5xx
  reintentan el evento (el `Topic` evita verlo dos veces). El aviso es JSON (`title`, `body`, `url`
  del panel, `tienda`, `tag`); al tocarlo, el service worker elige esa tienda y abre la página.
- **En Notificaciones**, «Push en este celular»: activar (pide el permiso al tocar), «Mandame un
  aviso de prueba» (`cuenta.push_prueba`, sólo a esa sesión) y desactivar. En el iPhone sólo anda con
  el panel instalado e iOS 16.4 o más nuevo, y la fila lo dice.

## 44. Tiempo real con WebSockets: Postgres avisa, Go reparte, la web recarga lo suyo (migración 00061)

- **El origen es la base**: triggers por sentencia (tablas de transición) en las tablas que importan
  hacen `pg_notify('tiempo_real', …)`, que Postgres entrega recién al confirmarse la transacción: lo
  revertido no avisa y no hay forma de olvidarse de avisar desde un camino (panel, worker, webhook,
  tienda online, importación). El aviso dice sólo **qué** cambió: `{t: tienda, k: tipo, id}` y, para
  rutear, `u` (persona) o `s` (el id público de una sesión). Nunca montos, nombres ni textos.
- **Ráfagas**: más de 20 filas de una tienda en una sentencia (importación, aumento masivo) salen como
  un aviso `lote`; Postgres junta los idénticos de una transacción; Go vuelve a juntar en ventanas de
  250 ms (más de 25 del mismo tipo y tienda → un lote) y espacia las familias ruidosas (visitas cada
  10 s, carritos y actividad cada 2 s, importación cada 1 s; lo del medio sale al final, no se pierde).
  Un UPDATE sólo avisa si cambió algo de lo que se mira (no el «visto» de la sesión, un token de
  Mercado Pago renovado ni el borrador del diseño).
- **Go**: `plataforma/tiemporeal`. Una conexión dedicada con `LISTEN` (fuera del pool), latido cada
  30 s y reconexión con espera creciente y azar; al volver, a todos «resincronizá». El hub reparte por
  tienda y filtra por rol con la misma tabla de `roles` (§37): pedidos, pagos, productos, envíos,
  turnos y devoluciones a quien tiene Pedidos; clientes, carritos, reseñas y preguntas a Clientes;
  cupones a Catálogo; comprobantes y cobros a Finanzas; integraciones a Tienda o Finanzas;
  importación, actividad, Mi tienda y dominios a Tienda; plan y equipo a Cuenta; la tienda, visitas y
  tareas a cualquiera del equipo. Un tipo sin destinatarios no le llega a nadie (cerrado por defecto,
  con test contra los triggers). Las sesiones son de la persona: sólo a sus pestañas.
- **El endpoint**: `GET /v1/tiempo-real` con la cookie `__Host-tlw_sesion` (o la «mirada» del
  soporte) o `Authorization: Bearer` (la app de iPhone). Con cookie, el `Origin` es obligatorio y
  tiene que ser `MOSTRADOR_SITIO` (más `MOSTRADOR_TIEMPO_REAL_ORIGENES`; en desarrollo, localhost):
  si no, 403 antes del upgrade. Las IA conectadas no entran. Ping de protocolo y un `latido` de texto
  cada 25 s (Cloudflare corta a los 100 s de silencio; el navegador no ve los pings), 512 bytes como
  tope de lo que manda el navegador (no manda nada), hasta 12 conexiones por sesión y 60 por IP, cola
  de 256 mensajes por conexión: si se llena, se corta (4408) y el navegador reconecta y recarga.
- **Cortes**: la sesión se vuelve a mirar cada 5 min y se corta al vencer; cerrarla (salir, «Cerrar»,
  contraseña nueva, quitar del equipo) corta en el acto por el aviso de la base (4401, no reintenta);
  cambiar de tienda o que te cambien el rol corta con 4000 («resincronizá»): el panel reconecta con
  lo nuevo y recarga todo. Al apagarse la API, 1012 y reconexión.
- **La tienda online** (sin sesión): `GET /v1/tiempo-real/publico?tienda=<slug>[&pedido=<código>]`.
  De la tienda sólo llegan `producto.*` (stock y ficha) y, con el código público, los cambios de ESE
  pedido sin el número. Acepta el sitio y el dominio propio verificado de esa misma tienda.
- **Ruteo**: el navegador se conecta a `wss://tulocalweb.com.ar/v1/tiempo-real` (mismo origen: viaja
  la cookie `__Host-`) y Caddy manda esa ruta a la API; en un dominio propio, la tienda se conecta a
  tulocalweb.com.ar (la CSP de `/t/` lo permite). En desarrollo, el proxy de Vite; en las pruebas,
  `TIEMPO_REAL_ORIGEN` (p. ej. `ws://localhost:4000`).
- **La web**: un cliente por pestaña (`lib/tiempo-real`), con reconexión exponencial con azar (1 s a
  30 s), que detecta una conexión muerta por silencio (65 s) y reconecta al volver la red o la pestaña.
  Cada aviso invalida sólo lo suyo (`depends('tienda:<cosa>')` en los load, juntando 300 ms); al
  reconectar, al pedirlo el servidor o al volver la pestaña después de 15 s, `invalidateAll`. Con la
  pestaña oculta no recarga nada hasta que vuelve. «Reconectando…» en la barra sólo si el corte dura
  más de 4 s. Sin WebSocket, todo anda como antes. En la demo (sandbox) no se conecta.
- **Avisos**: pedido nuevo con toast «Ver» (no el que cargó uno mismo en el punto de venta o Crear
  pedido) y sonido opcional en Notificaciones, por navegador y apagado por defecto. Mi tienda no se
  recarga sola si otra persona publica (no pisa lo que se está editando): avisa. En la tienda online,
  si se agota algo del carrito (o el producto que se mira) se avisa antes de pagar, con «Ajustar
  carrito»; en la página del pedido, «¡Recibimos tu pago!» al acreditarse.
- **Sin origen todavía**: la Bandeja (WhatsApp) no tiene tabla en la API (vacía hoy); cuando la
  tenga, se suma su trigger y su clave. Las notificaciones del panel ya tienen la suya (§45).

## 45. La campanita del panel: notificaciones por persona (migración 00062)

- **Una fila por evento y por persona**: `notificaciones` (tienda, persona, `clave` del hecho, ícono,
  grupo ventas/sistema, título, detalle, `ruta` del panel, leída). RLS por tienda como todas y,
  además, políticas RESTRICTIVE por persona: cada una ve y marca sólo las suyas. La app no inserta
  directo: el worker usa `app.notificar` (SECURITY DEFINER), que sólo escribe en la tienda fijada y
  para quien es del equipo de esa tienda. `UNIQUE (tienda, persona, clave)` con `ON CONFLICT DO
  NOTHING`: reprocesar el outbox no duplica.
- **Desde los mismos eventos que el mail y el push**: `notificaciones.Generador.Envolver` envuelve
  los manejadores del worker; primero escribe la notificación y después corre el manejador que ya
  estaba. Si falla la notificación, el evento se reintenta sin haber mandado nada.
  - `tienda.venta` → «Nueva venta #N» (transferencia por confirmar, transferencia acreditada o pago
    acreditado en Mercado Pago), a quien tiene Pedidos.
  - `tienda.sin_stock` → «Te quedaste sin stock», a Productos (no sale si ya repusieron).
  - `push.pregunta_ml` → «Pregunta nueva en Mercado Libre» (sin el texto), a Clientes.
  - `pago.aprobado` (ya existía, sin manejador) → links de pago (a Finanzas), señas de turnos y pagos
    de pedidos cargados en el panel (a Pedidos). Los de la tienda online ya avisan como venta y el QR
    lo ve quien cobra.
  - Eventos nuevos, encolados en la misma transacción que el cambio: `tienda.resena` (reseña de la
    tienda online, a Clientes, con las estrellas y sin el texto), `integracion.reautorizar` (Mercado
    Pago o Mercado Libre dejaron de aceptar la conexión, sólo la primera vez; a Tienda o Finanzas),
    `equipo.aceptada` (alguien se sumó con la invitación; a Cuenta, no a quien se sumó),
    `plan.activado` y `plan.pausado` (la suscripción del abono; a Cuenta).
  - El monto, sólo con Finanzas (§37): Ventas y Local leen «2 productos · Transferencia por
    confirmar». Nunca datos de quien compró.
  - El certificado de ARCA por vencer no tiene aviso en la API (no hay nada que lo detecte): no hay
    notificación hasta que exista.
- **API** (`roles.Panel`, sólo la sesión de una persona; una IA o el soporte mirando la reciben vacía):
  `GET /v1/panel/notificaciones` (de a 20, de la más nueva a la más vieja, cursor y `sinLeer`),
  `GET /v1/panel/notificaciones/sin-leer`, `POST /v1/panel/notificaciones/{id}/leida` y
  `POST /v1/panel/notificaciones/leidas`. `resumen.notificaciones` queda vacío con la API (la demo
  sigue armando las suyas).
- **Retención: 90 días.** `app.purgar_notificaciones()` la corre el worker una vez por día; la lista
  tampoco muestra lo más viejo aunque todavía no se haya purgado.
- **Tiempo real**: un trigger por sentencia avisa `notificacion.nueva` (alta) y `notificacion.leida`
  (marcada) con `u`, sin id ni texto; el hub lo entrega sólo a las pestañas de esa persona paradas
  en esa tienda. La web recarga `tienda:notificaciones` (el armazón del panel).
- **La web**: la insignia con el número (hasta «9+») sobre la campanita, la lista con Todo / Ventas /
  Sistema, «Ver anteriores» con el cursor, «Marcar todo como leído» sólo si hay sin leer y vacío
  honesto. Tocar una la marca leída y abre su página (sólo rutas de `/panel`).
- **De paso**: `/panel/pedidos/<n>` abierto en el navegador (el «Ver el pedido» de los mails y el
  push) devolvía el JSON de la hoja; ahora lleva a `/panel/pedidos?pedido=<n>`.

## 46. Factura A desde el checkout (migración 00063)

- **Cuándo se ofrece**: sólo si la tienda es responsable inscripta y conectó ARCA
  (`TiendaPublica.cobros.facturaA`). Monotributistas emiten C, que sirve para cualquiera: no se pide nada.
- **El checkout**: dos campos opcionales más en los datos del comprador de cada plantilla
  (`lib/tienda/factura.ts`): «CUIT (si necesitás factura A)» y «Razón social». La web valida el
  verificador; Go lo vuelve a validar (`arca.CUITValido`) y, si la tienda no la ofrece, lo ignora.
- **Dónde queda**: en `pedidos.comprador.factura` (`{cuit, razonSocial}`) y en la nota del pedido, así
  la tienda lo ve aunque facture a mano.
- **La cola**: `app.encolar_factura` emite A (tipo 1, DocTipo 80, condición 1) cuando el pedido trae
  CUIT y la tienda es RI; si no, B a consumidor final como antes. Sólo se pide como responsable
  inscripto: a monotributistas que piden A no se les pregunta la condición (quedaría mal informada).

## 47. API de conversiones de Meta (migración 00064)

- **Para qué**: la compra llega a Meta desde el servidor además del píxel, así las campañas la
  cuentan aunque el navegador bloquee el píxel. Mismo evento (`Purchase`) y mismo `event_id` (el
  número de pedido): Meta junta los dos y no la cuenta doble.
- **El token**: lo genera la tienda en el Administrador de eventos y lo pega en Integraciones › Meta
  Ads. Go lo comprueba (un GET al píxel, que no crea eventos) y lo guarda cifrado con la clave de la
  tienda (`meta_conversiones`). Nunca vuelve al navegador. Código de eventos de prueba opcional.
- **Consentimiento**: sólo con las cookies de medición aceptadas. El aviso de la tienda deja la
  cookie `tlw-medir-<slug>` y el checkout del BFF manda IP, navegador, `_fbp`/`_fbc` y la URL sólo
  con `si`. El email y el teléfono se hashean (SHA-256) antes de encolar.
- **Envío**: outbox `meta.compra` → worker (`integraciones/conversiones`). Meta caída o 429: se
  reintenta. Un rechazo (token vencido, píxel ajeno): no se reintenta y el panel muestra el motivo.
  Compras de más de 7 días se descartan (Meta no las acepta).
- **La demo** no la tiene: conectar responde que es una tienda de ejemplo.

## 48. Aviso de vencimiento del certificado de ARCA (migración 00065)

- **Por qué**: el certificado dura dos años; vencido, las ventas quedan «sin facturar» y antes sólo
  se veía abriendo Integraciones › ARCA.
- **Cuándo**: a 30 días, a 7 y al vencer, una vez cada umbral (`arca_config.aviso_vence`). Subir un
  certificado nuevo (o generar otro pedido) lo vuelve a cero. Lo revisa el worker en la vuelta del
  drenador de facturas (`app.tiendas_con_arca_por_vencer`, sólo ids).
- **A quién**: evento `arca.vence` → campanita y mail a quien tiene Finanzas (la dueña, al email
  del negocio si lo cargó). El mail no sale si en el medio ya subió uno nuevo. Los dos llevan a
  `/panel/integraciones?arca`, que abre el asistente.

## 49. Precios mayoristas (migración 00066)

- **Qué es**: cada producto (y cada variante) puede tener un «precio mayorista» opcional
  (`productos.precio_mayorista`, `variantes.precio_mayorista`; en la variante, vacío usa el del
  producto sólo si la variante va al precio del producto: una presentación con precio propio no
  hereda el mayorista de otra). La tienda pone en Configuración › Venta mayorista un mínimo del carrito:
  unidades, monto o los dos (al menos uno). Con los dos, el carrito tiene que cumplir los dos; el monto
  se cuenta con los precios minoristas (si no, al pasar a mayorista dejaría de alcanzar).
- **Un solo interruptor**: prendida o apagada es `tiendas.automaticas->>'mayorista'`, la promoción
  automática «Precio mayorista» que ya mostraba Cupones (ahora con el mínimo de verdad en su detalle).
  El mínimo vive en `tiendas.mayorista_unidades` (12 de entrada, como el handoff) y
  `tiendas.mayorista_monto`. La regla es una: `dominio/mayorista` en Go y `lib/tienda/mayorista.ts`
  en la web, con los mismos casos en los tests.
- **Al comprar**: cuando el carrito llega al mínimo, todos los productos con precio mayorista pasan a
  ese precio (nunca a uno mayor que el de lista). Go lo recalcula en `vidriera.Comprar` con los
  precios de la base (`preciosCarrito`), pasa los precios a `pedidos.CrearDesdeProductos` y marca el
  pedido `mayorista = true` (la marca que ya filtraba Pedidos › Mayoristas); si ningún producto tenía
  precio mayorista, no se marca. El cupón y el envío gratis se miden con lo que se cobra. La demo
  hace lo mismo en el checkout del BFF (`detalleLineas` con el mínimo) y en su `crearPedido`.
- **La tienda online**: `TiendaPublica.mayorista` (el mínimo, null si no vende por mayor) y
  `mayorista` en cada producto y variante (null con la venta apagada). El carrito (`detalleLineas`)
  pasa solo a los precios mayoristas y deja el minorista tachado con el `anterior` que cada plantilla
  ya dibuja. Yerbal lo dibuja como el handoff: la barra de arriba («Comprás por mayor desde…»), la
  tabla «Precio por mayor» de la ficha, el aviso amarillo del carrito y «¿Tenés un comercio?». Las
  demás plantillas, con piezas propias: «Por mayor: $ X desde N unidades» en la ficha y en el carrito
  cuánto falta (o «Precio mayorista aplicado») con lo que se ahorra.
- **El panel**: la caja «Mayorista» del precio por canal del editor (la del handoff, «Desde 12 u»),
  con su precio por variante; Ventas lo ve pero no lo cambia (Go lo conserva, como los demás
  precios). Configuración › Venta mayorista pide Catálogo. Subir o bajar precios un % desde la lista
  de productos mueve también los mayoristas; el aumento de Listas de precio no los toca (tampoco toca
  los de las variantes).
- **Lo que no cambia**: el feed de productos y el píxel (y la API de conversiones, que repite lo del
  píxel) siguen con el precio minorista; el total de la compra sí es el cobrado.

## 50. Newsletter de cada tienda (migración 00067)

- **Sumarse**: el bloque de newsletter del pie, donde el handoff lo dibuja (Ceibo, Mostaza, Rosedal,
  Yerbal y las doce del motor de `Plantillas Tienda`), con sus piezas; el texto de arriba sigue
  siendo el de cobros (no se promete un 10 % off que la tienda no cargó). Las plantillas sin bloque
  en el diseño (Andes, Malbec, Miga, Miga Clásica, Pampa, Quebracho y Tijera) suman en el checkout
  la casilla «Quiero recibir novedades», desmarcada, con los estilos de sus campos.
- **Doble confirmación**: el alta queda pendiente y el outbox (`newsletter.confirmar`) manda «Confirmá
  tu suscripción» con el link `/t/<slug>/newsletter/confirmar/<token>`. La página confirma con un
  botón, no al abrirla (los antivirus de correo abren los links). Pedirlo de nuevo sin confirmar
  reenvía el mismo link; si ya confirmó, no sale nada (y la respuesta es la misma: no se filtra quién
  está suscripto). Topes: 10 por hora por IP y tienda, 300 por hora por tienda y 3 por día por email
  (`plataforma/limite`); en la demo, 20 por hora por visitante.
- **Baja con un clic**: cada novedad trae el link y los encabezados `List-Unsubscribe` y
  `List-Unsubscribe-Post: List-Unsubscribe=One-Click` (RFC 8058) a `/t/<slug>/newsletter/baja/<token>`.
  Esa misma página recibe el POST del servidor de correo (sin Origin: exceptuada del CSRF, el token es
  la credencial) y da de baja sin más pasos. La baja borra la fila: el email no queda guardado.
- **Datos**: `suscriptores` (email, origen tienda/checkout, token, confirmado), `newsletter_envios`
  (asunto, texto, hasta 4 productos, destinatarios) y `newsletter_entregas` (una por envío y
  destinatario: pendiente, enviado, omitido o fallido). Todo con RLS por tienda.
- **Panel**: Crecer › Newsletter, junto a Marketing (permiso Tienda, el de marketing en `roles.go`).
  Números (confirmados, sin confirmar, mails de hoy contra el tope), «Mandar novedad» con asunto,
  texto y hasta 4 productos publicados (foto, precio y link) y la vista previa del mail al costado,
  confirmación con la cantidad de destinatarios, historial de envíos (destinatarios, enviados y
  estado: enviando, sigue mañana o enviada), la lista con fecha, origen y estado, borrar uno (con
  Deshacer: se borra cuando vence el aviso) y exportar CSV.
- **Envío**: «Mandar» fija los destinatarios (los confirmados de ese momento) y encola
  `newsletter.tanda`. El worker manda de a 25, cada uno con su fila bloqueada y marcada apenas sale
  (una tanda repetida no duplica), y encola la siguiente: ya mismo, a los 10 minutos si el correo
  rechazó alguno (tres rechazos: fallido) o mañana a las 9 si la tienda llegó a **500 mails por día**.
  Quien se dio de baja en el medio queda omitido. Remitente con el nombre de la tienda
  (`correo.Mensaje.DeNombre`, la dirección sigue siendo hola@tulocalweb.com.ar) y Reply-To al email del
  negocio (o el de la dueña). El correo suma `Encabezados` (SMTP y Resend).
- **La demo**: los suscriptos de ejemplo salen de los clientes de la tienda de ejemplo; todo funciona
  y «Mandar» guarda la novedad en el historial sin mandar ningún mail (lo dice la página). Sumarse
  desde la tienda de la demo no manda el mail: el aviso trae el link de confirmar.

## 51. Suscripciones de la tienda con débito automático de Mercado Pago (migración 00068)

- **Qué**: la tienda arma planes en Ventas › Suscripciones (nombre, descripción, foto, precio por
  período, cada 2 semanas / mensual / cada 2 meses, productos del catálogo y/o texto de qué incluye,
  envío y/o retiro, cupo opcional, a la venta o pausado). Quien compra se suscribe desde la página del
  plan (`/t/{slug}/suscripciones/{plan}`: email, nombre, teléfono y entrega) y autoriza el débito en
  Mercado Pago. El club de vinos de Malbec y la suscripción de café de Mostaza salen de estos planes
  (sin planes, esa parte no se dibuja); las demás plantillas suman «Suscripciones» con sus piezas.
- **Cobro**: suscripción de Mercado Pago sin plan (preapproval `pending` con `init_point`) creada con
  el token de la cuenta conectada de la tienda, nunca con el de la plataforma (eso es el abono de Tu
  Local Web). `external_reference = SUS-<id>` para no confundirse con el abono. Las tres
  frecuencias son las que la API acepta sin vueltas (14 días, 1 mes, 2 meses). Sin Mercado Pago
  conectado, la tienda no ofrece planes y el panel lleva a conectarlo.
- **La plata la pone el servidor**: el navegador manda el plan y los datos; el precio sale del plan y
  la suscripción lo copia (nombre, precio, frecuencia). Editar el plan no cambia lo que pagan quienes ya
  están: es lo que autorizaron en Mercado Pago.
- **Avisos**: `subscription_preapproval` (firmado, como todos) va al outbox como `mp.suscripcion`; la
  tienda sale del débito guardado (`app.tienda_de_preapproval`) o de la cuenta del aviso. Si no es de
  una tienda, se concilia como el abono. El worker trae el preapproval con el token de la tienda y aplica
  el estado (autorizada → activa y «alta»; cancelada → «baja»; una cancelada no vuelve).
  `subscription_authorized_payment` va como `mp.cobro_suscripcion`: el worker trae el cobro autorizado;
  aprobado, arma el pedido; rechazado, avisa una vez (`rechazo:<id>`). El mismo cobro también llega
  como `payment` con la referencia de la suscripción y pasa por el mismo camino.
- **Un cobro = un pedido**: `suscripcion_cobros` con clave `pago:<id de Mercado Pago>` (PK por tienda),
  insertada en la misma transacción que el pedido con `ON CONFLICT DO NOTHING`. El pedido es de la
  tienda online (`canal web`, `pago mp`, `Aprobado`, `Para preparar`, `pedidos.suscripcion_id`): una
  línea «Suscripción: <plan>» por lo cobrado y debajo los productos del plan en $ 0 para preparar. El
  stock se descuenta sin frenar (el cobro ya se hizo). Los triggers de siempre hacen el resto: «Nueva
  venta», confirmación a quien compró y la factura de ARCA si está conectada. Si lo cobrado no llega al
  precio acordado, no hay pedido.
- **Link privado sin cuenta** (`/t/{slug}/suscripcion/{codigo}`, código aleatorio como el de los
  pedidos; es la `back_url` de Mercado Pago): ver, pausar, reanudar y dar de baja. Primero se cambia en
  Mercado Pago y recién después en la base; si Mercado Pago no responde, no cambia nada. La tienda
  cancela desde el panel igual.
- **Mails y campanita**: `suscripcion.alta` (mail a quien se suscribió con el link privado y a la
  tienda), `suscripcion.baja` (a los dos) y `suscripcion.rechazo` (a la tienda). Cada mail sale una vez
  (`suscripciones.avisos`). La campanita va a quien tiene Pedidos, con el precio sólo para Finanzas.
- **Roles**: ver suscriptores y cancelar, Pedidos; crear, editar, pausar y borrar planes, Tienda. Un
  plan con suscriptores no se borra: se pausa (sigue siendo de quienes ya están).
- **RLS** por tienda en `suscripcion_planes`, `suscripciones` y `suscripcion_cobros`. Topes del alta: 10
  por IP y 5 por email por hora; el cupo se mira con el plan bloqueado (cuentan las activas, las pausadas
  y las altas de la última hora que se están autorizando).
- **La demo**: los planes y suscriptores se ven y se crean; suscribirse la deja activa sin cobrar y la
  página lo dice. Malbec y Mostaza traen planes y suscriptores de ejemplo.
- **Para producción**: en la aplicación de Mercado Pago (Tus integraciones › Webhooks) hay que sumar los
  eventos «Planes y suscripciones» (`subscription_preapproval` y `subscription_authorized_payment`) a
  la misma URL `https://<api>/v1/webhooks/mercadopago` con la misma clave secreta. Mercado Pago pide
  que el email de quien paga sea el de su cuenta: si autoriza con otra, el débito queda pendiente.

## 52. Lanzamientos con reserva, ofertas con fin y encargo con seña (migración 00069)

- **Modelo**: el producto suma `lanzamiento`, `reserva_sena` y `encargo_sena` (% de seña, NULL = no
  se ofrece), `oferta_precio` y `oferta_hasta` (el precio de lista sigue en `precio`) y
  `lanzamiento_avisado`. «Avisame cuando salga» va en `avisos_lanzamiento` (un email por producto,
  RLS por tienda). El pedido con seña suma `sena`, `sena_tipo` (reserva o encargo), `sena_pagada_en`,
  `sena_vence` y `saldado_en`. Se configura aparte del producto (`GET/PUT
  /v1/panel/productos/{id}/lanzamiento`, PUT con Catálogo porque lleva precio y señas), así el editor
  de productos no cambia de forma.
- **La tienda online**: antes de la hora, el producto se ve con «Próximamente», su fecha y su cuenta
  regresiva y no se compra (Go lo rechaza al comprar: `proximamente`); a la hora queda a la venta
  solo, porque la vidriera lo calcula con el reloj (no hay estado que cambiar). La oferta vigente
  pone `precio` = oferta y `anterior` = lista y se cobra así: la vidriera fija el precio de cada
  línea del pedido. Las variantes con precio propio no entran en la oferta. Vencida, vuelve el de
  lista solo.
- **Cuenta regresiva**: las plantillas la dibujan con las piezas del handoff (`cd` de la barra de
  cada una; el bloque grande del hero en Pampa y la «Oferta relámpago» en Yerbal) sólo si hay algo
  real que contar: el lanzamiento próximo o la oferta con fin más cercanos (`lib/tienda/lanzamiento.ts`).
  Sin eso no se dibuja nada. Los cupones con vencimiento no se cuentan: muchos son códigos privados
  (VIP, carrito) y mostrarlos los publicaría. Al llegar a cero, la tienda se vuelve a pedir.
- **Avisos**: tope de 10 por hora por IP y 20.000 por producto; repetido no suma ni lo dice. El
  worker mira cada minuto `app.lanzamientos_por_avisar` (sólo ids), marca el producto con un UPDATE
  condicionado (dos workers no avisan dos veces) y encola `lanzamiento.avisos`: cada evento manda
  50 mails, marca cada email apenas sale (un reintento no repite) y encola el siguiente si quedan.
  Reprogramar a una fecha futura deja el aviso pendiente otra vez.
- **Seña** (reserva antes del lanzamiento o encargo ya lanzado): `POST /v1/publico/tiendas/{slug}/senas`.
  Precio y seña (% redondeado al peso, como los turnos) los calcula Go; el pedido se crea «Sin pagar»
  con el stock descontado (el cupo de la reserva es el stock) y retiro en el local (sin retiro, la
  entrega se coordina: lo dice la nota). La preferencia de Checkout Pro va con la cuenta de la tienda,
  sin cuotas, como cobro `pedido` por el monto de la seña y vence con `sena_vence` (30 min); pasado eso
  el worker cancela el pedido (`app.senas_vencidas`) y el stock vuelve. Si Mercado Pago no responde al
  armarla, se cancela en el momento. El webhook (idempotente como todos: `mp_cobros` + estado previo)
  la aplica: «Para preparar», pago en «Seña», `tienda.venta` y el mail «Recibimos tu seña». Si llega
  tarde a un pedido ya cancelado, sólo queda anotada para devolverla desde el pedido.
- **Saldado**: «Seña pagada · falta $X» en la hoja del pedido y «Marcar como saldado» (con Deshacer):
  el pago pasa a «Cobrado» y recién ahí salen el mail de pago completo y la factura (la seña no
  factura sola).
- **Quebracho**: el «A medida» del handoff es el configurador: un producto con opciones se arma paso a
  paso (una opción por paso, con tarjetas si los valores tienen color o foto y chips si no) y la
  barra oscura muestra lo elegido, el precio y, con encargo, la seña y «Encargar». El modelo no tiene
  medidas: la vista de planta del diseño se reemplaza por la foto de la variante elegida, y los pasos
  son las opciones que tenga (hasta 3).
- **Pampa**: «Drops y reserva» con los lanzamientos reales: el hero cuenta el próximo (o la oferta que
  termina), «El drop» lista lo que todavía no salió y «Avisame» anota el email (el diseño pedía
  WhatsApp; el aviso es por mail).
- **La demo**: todo se ve y se usa; la seña se da por pagada y la confirmación dice que en la demo no
  se cobra. No manda mails.
