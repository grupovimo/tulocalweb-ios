# Cobros con Nave, GOcuotas y Talo

Pedido de Nahuel (2/11/2026): sumar Nave, GOcuotas y Talo como medios de cobro, con la misma lógica que
Mercado Pago (se conectan en Integraciones y aparecen solos en el checkout de la tienda). OAuth cuando lo
haya; si no, la persona pega sus credenciales, con la guía de dónde sacarlas y verificación al guardar.

## Nave (Galicia / Naranja X)

Sin documentación pública: los datos salen del plugin oficial `nave-for-woocommerce` 1.0.3 (wordpress.org).
Sólo para clientes Galicia o Naranja X. **Sin OAuth**: credenciales del comercio (portal Nave → Integraciones):
`client_id`, `client_secret` y `pos_id`.

- Token (OAuth2 M2M): `POST https://services.apinaranja.com/security-ms/api/security/auth0/b2b/m2msPrivate`
  (sandbox: `https://homoservices.apinaranja.com/security-ms/api/security/auth0/b2b/m2ms`), JSON
  `{client_id, client_secret, audience: "https://naranja.com/ranty/merchants/api"}` → `{access_token, expires_in}`.
- Crear el cobro: `POST https://api.ranty.io/api/payment_request/ecommerce` (sandbox `api-sandbox.ranty.io`),
  Bearer. Cuerpo: `external_payment_id`, `seller.pos_id`, `transactions[{amount{currency,value "123.45"},
  products[{name, description, quantity, unit_price{currency,value}}]}]`, `buyer{user_id, session_id, name,
  user_email, doc_type "DNI", doc_number, phone "+54…", billing_address{street_1, street_2, city, region,
  country, zipcode}}`, `additional_info{callback_url, notification_url}`, `platform{id, type: "mktplace",
  data{callback_url, notification_url}}`, `duration_time` (segundos, 900). Respuesta: `id`, `checkout_url`.
- Estado de la intención: `GET {base}/payment_requests/{id}` → `status.name`: PENDING, PROCESSED,
  SUCCESS_PROCESSED (pagado), FAILURE_PROCESSED (reintenta), EXPIRED, DISABLED, BLOCKED.
- Aviso: POST JSON a `notification_url` con `payment_id` y `payment_check_url` (sin firma: el plugin pone
  un secreto por pedido en la URL). Se confirma pidiendo `GET https://punku.ranty.io/payments-ms/payments/{id}/internal`
  → `status.name`: APPROVED, REJECTED, CANCELLED, REFUNDED, PARTIALLY_REFUNDED, PURCHASE_REVERSED, CHARGEBACK_REVIEW.

## GOcuotas (cuotas con débito)

Docs: https://gocuotas-api.stoplight.io/docs/gocuotas (API Redirect V1). **Sin OAuth**: email y contraseña de
la cuenta de comercio (la dan los administradores de GOcuotas, integraciones@gocuotas.com).
Base: `https://www.gocuotas.com/api_redirect/v1` (sandbox `https://sandbox.gocuotas.com/api_redirect/v1`).

- Token: `POST /authentication?email=…&password=…` → `{token}`; después `Authorization: Bearer`.
- Checkout: `POST /checkouts` con parámetros en la query: `amount_in_cents`, `email`, `order_reference_id`,
  `phone_number`, `url_success`, `url_failure`, `webhook_url` → `{url_init}` (se redirige ahí).
- Aviso: POST JSON a `webhook_url` con `status`: undefined | approved | denied (el de cancelación llega a los
  30 min). Sin firma: confirmar con `GET /orders/{id}` (o la lista `GET /orders`).
- Sólo cuotas sin interés con débito; mínimo ~$ 2.000; el comprador valida su teléfono por SMS.

## Talo (transferencias, cripto, Pix)

Docs: https://docs.talo.com.ar · SDK oficial: github.com/talo-pay/talo-sdk.
Base `https://api.talo.com.ar` (sandbox `https://sandbox-api.talo.com.ar`).

- **Hay OAuth de partners** (como Mercado Pago): `partners.getAuthorizationUrl(partner_id, {referredUserId})`
  → la persona autoriza en Talo → vuelve con `code` → `exchangeToken({code, client_id: PARTNER_ID,
  client_secret: PARTNER_SECRET})` → `{token, user_id}`. Requiere que Tu Local Web sea partner de Talo.
- Sin partner: credenciales del comercio (panel Talo → Conexiones): `client_id`, `client_secret` y `user_id`;
  token con `POST /users/{user_id}/tokens` `{client_id, client_secret}` → `data.token`.
- Crear el cobro: `POST /payments/` `{user_id, price{amount (pesos, no centavos), currency "ARS"},
  payment_options ["transfer"], external_id, webhook_url, redirect_url?, motive?, client_data?, partner_id?}`
  → `data.id`, `data.payment_url`, `data.payment_status`.
- Estados: PENDING, SUCCESS, OVERPAID, UNDERPAID, EXPIRED.
- Aviso: POST JSON `{message, paymentId, externalId}` (sin firma): confirmar con `GET /payments/{paymentId}`.

## Cómo encaja

Igual que Mercado Pago: credenciales cifradas por tienda (como Correo y ARCA), conectar y desconectar desde
Integraciones (web y app, con la guía y la verificación), el medio aparece en el checkout de la tienda cuando
está conectado, y cada aviso se verifica consultando el estado a la plataforma antes de marcar el pedido pagado.
