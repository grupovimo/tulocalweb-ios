# Handoff · Tienda online en bloques (15/11/2026)

Canvas: https://claude.ai/artifact/97YxFsdrr5x7YzosLwyEmz. Nahuel se quedó con **las tres direcciones**: cada tienda
elige una (Mi tienda › Estilo › Dirección) y combina con ella cualquier bloque y variante. Los `.dc.html` de esta
carpeta son la referencia visual; lo que manda es este README más `packages/contrato/bloques.json`.

| Archivo | Qué muestra |
|---|---|
| `A-*` | Vidriera: moda, café, catálogo de almacén, ficha, celular (portada, ficha, bolsa, pago), recién abierta |
| `B-*` | Mostrador: lo mismo |
| `C-*` | Pedido: lo mismo (el café es menú para pedir con el pedido al costado) |
| `Bloques.dc.html` | Los bloques y sus variantes |
| `Reglas.dc.html` | Tokens de cada dirección, reglas y opciones del dueño |

Las fotos (`fotos/`) son las de los catálogos de ejemplo; en el canvas se llaman `/_blob/<id>`. El nombre del archivo
dice qué producto es.

## Tokens por dirección

| Token | A · Vidriera | B · Mostrador | C · Pedido |
|---|---|---|---|
| Títulos | Fraunces 300–400, −0,02em | Bricolage Grotesque 800, −0,03em | Geist 700, −0,025em |
| Texto | Figtree 400/500/600 | Instrument Sans 400–700 | Geist 400–600 |
| Escala (px) | 84 · 52 · 30 · 19 · 16 · 13 | 72 · 40 · 28 · 17 · 15 · 13 | 44 · 22 · 17 · 15 · 13 |
| Antetítulo | 12 px, 600, mayúsculas, +0,14em, color acento oscuro | chip píldora 12 px 700 | 13 px 600 |
| Radio tarjetas / fotos | 0 / 0 | 16–28 / 12–24 | 12 / 10–12 |
| Botón | rectangular, alto 52, borde 1 px | píldora, alto 52, 700 | radio 10, alto 44–48 |
| Tarjeta | borde 1 px `borde`, sin sombra | blanca + sombra `0 1px 2px α.06, 0 8px 24px -12px α.18` | blanca + borde 1 px |
| Ritmo vertical entre bloques | 120 | 56–72 | 24–32 |
| Margen lateral (1440) | 56 | 40 | 32 |
| Encabezado | logo centrado, menú a la izquierda | logo + buscador píldora + carrito oscuro | compacto 64 px + buscador; categorías al costado |
| Precio | 600, cuotas abajo en 13 px | 700 grande, cuotas en verde ok, precio con transferencia | 600, todo en una línea |
| Carrito | página / panel lateral | página con barra fija | siempre visible al costado (escritorio) y barra fija abajo (celular) |

La forma (recta, redondeada o píldora), la densidad y la tipografía arrancan con los valores de la dirección y el
dueño puede cambiarlas. Los colores salen de los roles (`fondo`, `superficie`, `texto`, `textoSuave`, `borde`,
`acento`, `sobreAcento`, zonas de encabezado y pie), igual que hoy.

## Reglas (valen para las tres)

1. **Nada vacío.** Un bloque sin datos no se dibuja; el editor dice qué le falta. La tienda con un producto se ve armada (`*-vacia`).
2. **Grillas llenas.** La cantidad de columnas sale de cuántos elementos hay: 1 → destacado, 2 → mitades, 3 → tercios, 4+ → la grilla elegida, recortada a filas completas en galerías.
3. **Sin grises.** Un producto o una categoría sin foto lleva la inicial y una textura con el acento (`*-vacia`).
4. **Contraste AA.** Texto sobre foto siempre con velo; si un color no se lee, se recalcula el texto encima.
5. **Un botón principal por pantalla**, nunca gris si se puede tocar.
6. **Escala tipográfica fija** (tabla de arriba); cuerpo ≥ 15 px.
7. **Sin renglones viudos** (`lib/tienda/titulo.ts`, `&nbsp;` en la última palabra).
8. **Lo flotante no tapa**: WhatsApp y la barra del pedido reservan su espacio.
9. **Números argentinos**: `$ 186.000`, `4,8`, `3 cuotas de $ 62.000`; promedios con un decimal.
10. Táctil ≥ 44 px; foco visible; botones e inputs reales con su `label`.
