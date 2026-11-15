# Promociones y conversión en la tienda

15/11/2026. Pedido de Nahuel: «promos, banners, avisos, modales que se abren si te querés ir, modales
personalizables para promociones, lead magnets… tenemos que ser como Tiendanube y Shopify, o mejores».

Se apoya en el sistema de bloques (`tienda-bloques.md`): mismo catálogo, mismas direcciones, mismo
protocolo de vista previa. Nada se inventa: cada pieza visual sale de las primitivas del handoff
`docs/handoff/tienda-bloques/` (botones, tarjetas, fondos, tipografía de cada dirección).

## Qué ya existe (no se rehace, se conecta)
- Barra de anuncios (`estilo.barra`), bloques `banner`, `oferta`, `newsletter`.
- Cupones (`dominio/cupones`), newsletter con suscriptores (`dominio/newsletter`).
- Carrito abandonado, envío gratis desde un monto, aviso de stock («avisame»): **auditar** y completar
  lo que falte para que se vean en la tienda nueva.

## 1. Ventanas (popups)

`diseno.ventanas: Ventana[]` (hasta `limites.ventanas` = 8). Viven en el diseño: borrador y publicar
como todo lo demás. El catálogo está en `packages/contrato/bloques.json › ventanas`.

```ts
type Ventana = {
  id: string;            // a-z0-9, único
  nombre: string;        // interno, para el panel (hasta 40)
  activa: boolean;
  variante: 'centrada' | 'lateral' | 'completa' | 'inferior';
  contenido: 'promo' | 'cupon' | 'captacion' | 'ruleta';
  datos: Record<string, unknown>;   // campos según contenido (catálogo)
  disparo: { cuando: 'entrar' | 'segundos' | 'scroll' | 'salida' | 'carrito'; valor?: number };
  frecuencia: { cada: 'sesion' | 'dias' | 'siempre' | 'unaVez'; dias?: number };
  publico: 'todos' | 'nuevos' | 'recurrentes';
  paginas: ('inicio' | 'catalogo' | 'producto' | 'carrito' | 'nosotros')[];  // vacío = todas
  dispositivos: 'todos' | 'celular' | 'compu';
  desde?: string; hasta?: string;   // ISO; fuera de rango no sale
};
```

**Contenidos**
- `promo`: foto opcional, etiqueta, título, texto, botón (link), cuenta regresiva opcional (`hasta`).
- `cupon`: título, texto, código de un cupón existente (se muestra con «Copiar» y se aplica solo al
  carrito con un toque), condiciones cortas, cuenta regresiva opcional.
- `captacion` (lead magnet): título, texto, foto opcional, pide `email` (obligatorio) y opcionalmente
  `nombre` y `whatsapp`; casilla de consentimiento con link a privacidad. Premio, uno de:
  - `cuponFijo`: un cupón existente;
  - `cuponUnico`: el servidor genera un código propio de un uso (`porcentaje` | `monto` |
    `envioGratis`, `valor`, vence en `dias`);
  - `descarga`: un archivo o link (guía, recetario, catálogo PDF) que se muestra y se manda por mail;
    el archivo se sube con `POST /panel/archivos` (PDF, EPUB o ZIP hasta 20 MB, tipo por la firma,
    60 MB por tienda) y queda en `/f/archivos/<tienda>/<uuid>.<ext>`, junto a las fotos;
  - `nada`: sólo suscripción.
  El contacto entra a la newsletter con `origen = "ventana:<id>"` (sin duplicar por email).
- `ruleta`: 4 a 8 porciones `{texto, premio: cuponUnico-like | nada, peso}`; pide email antes de
  girar; **el servidor elige** la porción (pesos), una vez por email por ventana; la animación cae
  en la porción que devolvió el servidor.

**Disparadores**
- `entrar` (después de 1,5 s), `segundos` (valor 3–120), `scroll` (valor 10–90 %),
- `salida`: en compu, el mouse sale por arriba; en celular, subida rápida después de haber bajado más
  del 40 % o volver a la pestaña después de 20 s afuera; nunca antes de 8 s en la página,
- `carrito`: al agregar el primer producto.

**Reglas (mejores que la competencia, y las de Google contra intersticiales intrusivos)**
1. Una sola ventana por visita a la página; si varias califican, gana la primera de la lista.
2. Nunca en el checkout, la página de listo, legales ni arrepentimiento.
3. En celular, `completa` y `centrada` con disparo `entrar` se dibujan como `inferior` (hoja que
   sube y deja ver la tienda).
4. Siempre se cierra: X visible de 44 px, Esc, tocar afuera; `role="dialog"`, `aria-modal`, foco
   atrapado y devuelto, respeta «Reducir movimiento».
5. Frecuencia y «ya se suscribió» en `localStorage` (con try/catch); quien completó una captación
   no la vuelve a ver.
6. Texto sobre foto con velo (regla del handoff); sin renglones viudos (`titulo.ts`).
7. No tapa el botón de WhatsApp ni la barra de pedido; la ventana `lateral` va del lado contrario.

**Métricas**: por ventana y día — `vistas`, `cierres`, `envios` (captación/ruleta), `clics`
(promo/cupón), `ventas` y `monto` (pedidos que usaron el cupón de la ventana). Se mandan desde la
tienda en lote (`navigator.sendBeacon`) a `POST /t/{tienda}/promos/eventos`; la API agrega por día.

**Plantillas para crear rápido** (el editor ofrece estas y «En blanco»):
Bienvenida 10 % por mail · Antes de irte (cupón) · Ruleta de premios · Envío gratis desde $X ·
Lanzamiento con cuenta regresiva · Descargá la guía (lead magnet) · Sumate por WhatsApp.

## 2. Más herramientas de conversión (`estilo.conversion`)

```ts
conversion: {
  envioGratis: boolean;        // barra «Te faltan $X para el envío gratis» en bolsa y carrito (usa el monto ya configurado)
  pruebaSocial: { activa: boolean; cada: number /*s, 15–120*/; paginas: string[] };
       // avisos chicos «Ana de Córdoba compró Mate Imperial · hace 2 h»: SOLO pedidos reales pagados
       // de los últimos 7 días, nombre de pila + provincia, nunca inventados; sin pedidos, no sale
  etiquetas: { nuevo: boolean; oferta: boolean; ultimas: boolean; masVendido: boolean };
       // insignias automáticas en tarjetas y ficha: Nuevo (≤ 14 días), −X %, Últimas N, Más vendido (top 3 de 30 días)
  stockBajo: number;           // 0 = no; si quedan ≤ N: «Quedan 3» en la ficha
  compraFija: boolean;         // en celular, botón de agregar fijo abajo en la ficha
  sumaEnCarrito: boolean;      // en el carrito, «Sumá también» con combos/relacionados (hasta 3)
  favoritos: boolean;          // corazón en tarjetas y ficha, página /favoritos (en el dispositivo)
}
```

La barra de anuncios suma `hasta` por mensaje: cuenta regresiva y se esconde al vencer.

## 3. API (Go)
- `diseno.ventanas` y `estilo.conversion` validados desde el catálogo (como los bloques).
- `POST /t/{tienda}/promos/captar` `{ventana, email, nombre?, whatsapp?, consentimiento}` →
  `{premio: {tipo, codigo?, descarga?, vence?}}`; idempotente por (ventana, email); límite por IP.
- `POST /t/{tienda}/promos/ruleta` igual, devuelve `{porcion, premio}`.
- `POST /t/{tienda}/promos/eventos` (lote, sin respuesta útil, con tope por IP).
- `GET /t/{tienda}/promos/recientes` para la prueba social (cacheado 5 min).
- `GET /panel/promos/metricas?desde&hasta` por ventana. Ventas atribuidas por cupón.
- Cupones únicos: se crean en `cupones` con `origen = ventana:<id>`, un uso, vencimiento.
- Mail al captar: el premio y la descarga, con la marca de la tienda.
- `openapi.yaml` y el sandbox de la demo entienden todo.

## 4. Editor (web, iOS y Android idénticos)
Mi tienda suma la pestaña **Promociones**:
- lista de ventanas (activar, ordenar, duplicar, borrar), crear desde plantillas;
- panel de cada ventana: contenido, diseño (variante en miniaturas), cuándo sale, a quién, dónde,
  fechas; campos generados desde el catálogo;
- «Ver en la vista previa» abre la ventana en la vista previa (`{tipo:'ventana', id}` por el puente);
- métricas por ventana (vistas, % de envío, ventas);
- herramientas de conversión (sección 2) con interruptores y su explicación corta.
- Contactos captados: se ven en Newsletter con su origen.

## 5. Demos
Las demos estrenan: Lumbre con Bienvenida 10 %, Bruma con ruleta, Almacén Verde con envío gratis y
prueba social, para que se vea en vivo.
