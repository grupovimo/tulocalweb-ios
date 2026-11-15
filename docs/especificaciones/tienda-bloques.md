# Tienda online en bloques (diseño v2)

Reemplaza a las 17 plantillas monolíticas (`lib/tienda/plantillas/*.svelte`) y a `docs/especificaciones/plantillas-propias.md`.
Diseño: `docs/handoff/tienda-bloques/` (las tres direcciones, los bloques y las reglas). Catálogo:
`packages/contrato/bloques.json`. Plan aprobado el 14/11/2026.

## Por qué

- Las plantillas eran 11 archivos de unas 3.000 líneas, más un motor con 6 temas. Cada uno dibujaba todas las páginas, con cerca del 40 % de CSS en línea.
- Arreglar algo implicaba arreglarlo 12 veces.
- Personalizar se limitaba a recolorear.
- Había sólo 6 secciones fijas.
- Con pocos datos, la tienda se veía rota.

## Modelo

`tiendas.diseno_borrador` y `diseno_publicado` (jsonb) siguen siendo los mismos. Lo que es nuevo en v2 es lo marcado:

```ts
type Diseno = {
  version: 2;                        // nuevo. Sin version (o 1): diseño viejo, se migra al leer
  plantilla?: string;                // igual que hoy (sólo mientras el borrador la cambia)
  paleta: 0 | 1 | 2 | 3;             // igual (el acento rápido del onboarding)
  titulo?, lema?, portada?, logo?;   // igual: datos de la marca que usan otras pantallas
  compartirTitulo?, compartirDescripcion?, compartirFoto?;
  estilo?: EstiloTienda;             // lo de hoy más los campos nuevos (abajo)
  inicio: Bloque[];                  // nuevo: la portada, en orden
  paginas?: {                        // nuevo
    catalogo?: { variante: 'grilla' | 'lista'; filtros: 'costado' | 'arriba'; columnas: 2..5 };
    producto?: { variante: 'grilla' | 'miniaturas' | 'carrusel'; relacionados: boolean; calcularEnvio: boolean; bloques: Bloque[] };
    carrito?:  { variante: 'pagina' | 'panel'; envioGratisDesde: boolean };
    nosotros?: { bloques: Bloque[] };
  };
  secciones?: Seccion[];             // sólo en v1; en v2 no va
};

type Bloque = {
  id: string;          // [a-z0-9]{6,16}, único en el diseño (lo genera el editor)
  tipo: string;        // uno de bloques.json › bloques[].tipo
  variante: string;    // una de sus variantes
  oculto?: boolean;
  datos: Record<string, unknown>; // los campos del tipo; lo que falta toma porDefecto o lo sugerido
};

// Campos nuevos de EstiloTienda (todos opcionales; vacío = lo de la dirección):
direccion?: 'vidriera' | 'mostrador' | 'pedido';
forma?: 'recta' | 'redondeada' | 'pildora';
densidad?: 'aireada' | 'normal' | 'compacta';
ancho?: 'normal' | 'amplio';
fotoProducto?: 'cuadrada' | 'vertical' | 'apaisada';
encabezado?: 'izquierda' | 'centrado' | 'buscador';
pie?: 'simple' | 'columnas' | 'grande';
barra?: { variante: 'fija' | 'rotativa'; mensajes: { texto: string; link?: string }[]; fondo?: Fondo };
// anuncio/anuncioLink (v1) pasan a barra.mensajes[0]; textos (v1) pasan a los títulos de cada bloque.
```

Los límites, las opciones y los campos de cada bloque están en `bloques.json`. Go y la web validan contra ese
mismo archivo:
- Go lo embebe en `mitienda/bloques.go`.
- La web lo importa en `lib/tienda/sistema/catalogo.ts`.

El tope del PUT sube de 4 KB a `limites.bytesDiseno` (64 KB).

## Migración de v1 a v2

La migración es una función pura en los dos lados: `migrar.ts` en la web y `mitienda.Migrar` en Go. Las dos
corren los mismos casos de `packages/contrato/casos/diseno-v2.json`. Funciona así:

1. **Plantilla → preset.** La plantilla de la tienda define el preset. El preset trae la dirección, las variantes y el orden de los bloques; ver la tabla de abajo.
2. **Secciones → bloques**, en el mismo orden. Las ocultas quedan con `oculto: true`.

   | Sección | Bloque | Datos |
   |---|---|---|
   | `hero` | `portada` | `titulo`, `lema` y `portada` del diseño |
   | `cats` | `categorias` | |
   | `prods` | `productos` | fuente «vendidos» |
   | `revs` | `resenas` | |
   | `ig` | `galeria` | |
   | `wa` | `whatsapp` | |

   El título de cada bloque sale de `estilo.textos[seccion]`.
3. **Anuncio → barra.** `estilo.anuncio` y `estilo.anuncioLink` pasan a `estilo.barra` con un mensaje.
4. **Lo demás del estilo queda igual**: colores, fuentes, logos, redes, menú y leyenda del pie.

Mientras convivan, la tienda migra al vuelo cualquier diseño v1 que lea. La migración Go en la base es
`00xxx_diseno_v2.sql` más un comando de `mostradorctl`, con copia previa. Cuando termina, no queda ningún v1 y el
código v1 se borra.

## Presets (las 17 plantillas)

El preset de cada plantilla es `lib/tienda/sistema/presets.ts`: dirección, estilo base y bloques iniciales con
textos sugeridos. Ese archivo se exporta a `presets.json` para Go y para las apps.

| Dirección | Plantillas |
|---|---|
| Vidriera | ceibo, rosedal, quebracho, malbec, taller, torno |
| Mostrador | pampa, yerbal, andes, patio, hora, litoral, lente, tijera |
| Pedido | mostaza, miga, miga-clasica |

Cada preset conserva su paleta, su tipografía y sus funciones propias:
- comparador (Andes);
- aviso de +18 (Malbec);
- turnos (Tijera);
- drops y reserva (Pampa);
- pedido desde el menú (Mostaza y Miga).

Cambiar de preset con la tienda armada conserva los datos de los bloques del mismo tipo y se puede deshacer.

## Web

`apps/web/src/lib/tienda/sistema/`:

- `catalogo.ts`: tipos y validación (valibot) generados desde `bloques.json`, con los valores por defecto.
- `migrar.ts`, `presets.ts` y `tokens.ts`. `tokens.ts` arma las variables CSS por dirección y estilo, con el contraste AA de `acentos.ts`. Hay una sola hoja de estilos por dirección (`direcciones/*.css`), no CSS en línea.
- `formato.ts`: precio, promedio con un decimal, cuotas y precio con transferencia.
- `datos.ts`: resuelve los datos de cada bloque contra la tienda (productos por fuente, categorías con foto, beneficios automáticos) y decide si el bloque se dibuja (`seDibuja`).
- `primitivas/`: `Boton`, `Precio`, `Estrellas`, `Foto` (con placeholder de marca), `Seccion`, `Grilla` (columnas según la cantidad) y `TarjetaProducto` (variantes por dirección).
- `marco/`: `Encabezado`, `Pie`, `Barra`, `BotonWhatsApp` y `MenuMovil`.
- `bloques/<Tipo>.svelte`: un archivo por tipo, con las variantes por prop. `Bloques.svelte` recorre una lista y marca cada bloque con `data-bloque={id}`.
- `paginas/`:
  - `Inicio`, `Catalogo`, `Ficha`, `Carrito`, `Checkout`, `Listo`, `Nosotros`;
  - `Turnos` y `Turno`, `Suscripcion` y `MiSuscripcion`;
  - `NoExiste`, `Newsletter` y `Arrepentimiento`.

  Reutilizan toda la lógica que ya está compartida: carrito, cupones, envío, variantes, mayorista, turnos, suscripciones, reseñas, preguntas y medición.

`lib/tienda/Plantilla.svelte` dibuja con el sistema. Las 17 plantillas, `estilo/mapas.ts` y
`scripts/estilo/generar.ts` se borran cuando todo migró.

## Editor (web, iOS y Android iguales)

- **Páginas**: Inicio, Catálogo, Ficha, Carrito y Nosotros.
- **Lista de bloques**: arrastrar para ordenar, agregar entre medio (biblioteca con miniaturas por grupo), duplicar, ocultar y borrar, con deshacer y rehacer.
- **Panel del bloque**: elegir la variante con miniaturas. Los campos se generan desde `bloques.json`. Las fotos se suben con `POST /panel/fotos`.
- **Pestaña Estilo**: dirección, forma, densidad, ancho, fotos, encabezado, pie, colores, tipografías, barra, menú y redes.
- **Vista previa**: `/panel/mi-tienda/vista?pagina=...` recibe el borrador por `postMessage` (en las apps, por `evaluateJavaScript`) y lo dibuja sin recargar.
  - Al tocar un bloque, avisa `{tipo:'bloque', id}`: la web por `postMessage` al padre, iOS por `webkit.messageHandlers.tienda` y Android por `TiendaPuente.bloque`.
  - El editor abre el panel de ese bloque.
- Se mantienen el borrador y publicar.
