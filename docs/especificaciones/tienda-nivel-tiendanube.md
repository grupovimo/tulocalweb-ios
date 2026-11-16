# Tiendas al nivel de Tiendanube (o mejor)

16/11/2026. Nahuel: «aún siento que no estamos al nivel de Tiendanube a nivel tienda», mirando
bylula.com.ar y blasterargentina.com.ar (las dos en Tiendanube). Pidió «vamos a por todo».
Todo va al sistema de bloques: las 17 plantillas, web, iOS y Android (ver `tienda-bloques.md`).

## Lo que tienen ellas y nosotros no

1. Cuotas sin interés en todos lados (barra, cada tarjeta, ficha). Hoy `vidriera.go` manda `cuotas = 0`.
2. Precio sin impuestos nacionales (Ley 27.743, régimen de transparencia fiscal al consumidor).
3. Ficha con caja de confianza (plazo y medios de envío, garantía/cambios), fechas estimadas al
   cotizar, etiqueta «Envío gratis» sobre la foto, «agregá y tenés envío gratis».
4. Tarjetas que venden solas: cuotas, precio con transferencia, «Agregar» rápido, etiqueta de
   descuento por cantidad.
5. Prueba social con caras: galería de clientes y fila de videos verticales.
6. Muestras de color en las variantes, guía de talles a mano, páginas de información en el menú.

## Contrato (lo comparten API, tienda, panel y apps)

### Cobros de la vidriera (`GET /t/{tienda}` → `cobros`)
- `cuotas`: entero, las cuotas sin interés que ofrece la tienda: 0 (ninguna), 3, 6, 9 o 12.
  Sólo con Mercado Pago conectado; si no, 0.
- `cuotasDesde`: entero (pesos), monto mínimo para ofrecerlas. Por defecto 30.000.
- `impuestos`: `{ "mostrar": bool }`. Si es true, la tienda muestra «Precio sin impuestos nacionales».

### Producto de la vidriera
- `iva`: número, la alícuota de IVA del producto: 21 (por defecto), 10.5, 27 o 0.
  Precio sin impuestos nacionales = precio / (1 + iva/100), redondeado a centavos, con coma.

### Configuración (panel → API)
- Cobros: «Cuotas sin interés» (ninguna, 3, 6, 9, 12) y «desde $». Explica que el costo lo absorbe la
  tienda y que hay que activarlas también en Mercado Pago (link a la configuración de MP). La
  preferencia de Checkout Pro limita `installments` al máximo elegido (0 → sin límite propio).
- Impuestos: «Mostrar precio sin impuestos nacionales». Por defecto prendido para responsables
  inscriptos (`negocio.condicion`), apagado para monotributo.
- Producto: campo «IVA» con 21 / 10,5 / 27 / 0 (por defecto 21).

### Diseño (bloques.json)
- `paginas.producto.confianza`: `{ envio?: string, garantia?: string, cambios?: string }`, textos
  cortos (hasta 120). Vacío: el envío se arma solo de la configuración de envíos (moto, correo,
  retiro, envío gratis desde $X); garantía y cambios no salen si están vacíos.
- Bloque nuevo `clientes` (variantes `grilla`, `carrusel`): `titulo`, `texto`, `items: [{foto, nombre?, link?}]`
  (hasta 12), `cta?: {t, a}` («Subí tu foto y etiquetanos»).
- Bloque nuevo `videos` (variante `fila`): `titulo`, `items: [{video (https mp4/webm o link de
  Instagram/TikTok/YouTube Shorts), portada (foto), titulo?, producto?}]` (hasta 10). Los videos se
  ven en vertical 9:16; al tocar se abre un visor a pantalla completa con el producto para agregar.
- Menú: las páginas de información (`menu` del estilo) pueden apuntar a páginas propias de texto.

## Reglas
- Nada se promete si la tienda no lo configuró (sin MP no hay cuotas; sin envío gratis no hay etiqueta).
- Sin renglones viudos; texto sobre foto con velo; contraste AA.
- En local sólo `bun run check`, lint y unitarios de lo tocado; nada de builds. Todo a `main`.
