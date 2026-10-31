/**
 * Datos de la tienda de ejemplo de la demo: clientes y pedidos de los últimos 60 días sobre el
 * catálogo real de la plantilla (src/lib/catalogos.ts). Salen siempre iguales para una plantilla
 * (generador con semilla) y se fechan respecto de cuándo se creó el sandbox, así la demo se ve viva
 * pero no cambia mientras se usa. Lo que el visitante modifica se guarda aparte (tabla `cambios`).
 *
 * Canales, medios de pago, envíos y estados son los del panel del handoff.
 */
import { catalogoDe, type Catalogo, type ProductoEjemplo } from '#lib/catalogos.ts';
import { textoVariante } from '#lib/tienda/variantes.ts';
import { variantesDemo, type VariantesDemo } from './variantes.ts';

export type Canal = 'web' | 'wa' | 'ml' | 'ig' | 'pos' | 'tg';
export type MedioPago = 'mp' | 'transf' | 'modo' | 'payway' | 'uala' | 'efectivo' | 'tarjeta';
export type Envio = 'andreani' | 'oca' | 'correo' | 'moto' | 'retiro' | 'mercadoenvios';
export type EstadoPedido =
	'Sin pagar' | 'Para preparar' | 'Listo para despachar' | 'En camino' | 'Entregado' | 'Cancelado';

export type Cliente = {
	id: string;
	nombre: string;
	email: string | null;
	telefono: string | null;
	ciudad: string;
	foto: string | null;
	mayorista: boolean;
	alta: number;
};

export type Pedido = {
	numero: number;
	clienteId: string;
	canal: Canal;
	pago: MedioPago;
	/** "Aprobado", "Acreditada", "Pendiente", "Cobrado". */
	estadoPago: string;
	envio: Envio;
	ciudadEnvio: string;
	/** Adónde se manda (los de la tienda online); los de ejemplo la arman con el cliente. */
	entrega?: { direccion: string; cp: string } | null;
	estado: EstadoPedido;
	facturado: boolean;
	items: ItemPedido[];
	total: number;
	/** Milisegundos (epoch). */
	creado: number;
	/** Reservado o encargado con seña (§52): en la demo la seña se da por pagada (no se cobra). */
	sena?: { tipo: 'reserva' | 'encargo'; monto: number; pagada: boolean; saldado: boolean };
	/** Cobrado con precios mayoristas en la tienda online (§49). */
	mayorista?: boolean;
};

export type ItemPedido = {
	productoId: string;
	nombre: string;
	cantidad: number;
	precio: number;
	/** La variante vendida y su texto al momento de la compra ("Talle M · Negro"). */
	varianteId?: string | null;
	variante?: string | null;
};

export type Carrito = {
	id: string;
	clienteId: string;
	items: { productoId: string; nombre: string; cantidad: number; precio: number }[];
	total: number;
	creado: number;
};

export type Resena = {
	pedido: number;
	clienteId: string;
	productoId: string;
	producto: string;
	canal: Canal;
	estrellas: number;
	texto: string;
	creada: number;
};

/** Conversación de WhatsApp abierta (la bandeja). */
export type Conversacion = {
	clienteId: string;
	mensaje: string;
	creado: number;
	atiende: 'hornero' | 'vos';
	sinLeer: number;
};

export type Demo = {
	catalogo: Catalogo;
	clientes: Cliente[];
	/** Del más nuevo al más viejo. */
	pedidos: Pedido[];
	/** Unidades por producto (con opciones, la suma de sus variantes). */
	stock: Record<string, number>;
	/** Opciones, galería y variantes de los productos que las tienen. */
	variantes: Map<string, VariantesDemo>;
	/** Productos publicados en Mercado Libre. */
	enMl: string[];
	carritos: Carrito[];
	resenas: Resena[];
	conversaciones: Conversacion[];
};

/** Generador con semilla (mulberry32). */
function azar(semilla: number) {
	let a = semilla >>> 0;
	return () => {
		a = (a + 0x6d2b79f5) >>> 0;
		let t = a;
		t = Math.imul(t ^ (t >>> 15), t | 1);
		t ^= t + Math.imul(t ^ (t >>> 7), t | 61);
		return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
	};
}
const hash = (s: string) =>
	[...s].reduce((h, c) => Math.imul(h ^ c.charCodeAt(0), 16777619), 2166136261) >>> 0;

const NOMBRES = [
	'Lucía',
	'Martín',
	'Norma',
	'Carla',
	'Julián',
	'Valeria',
	'Sofía',
	'Tomás',
	'Camila',
	'Federico',
	'Florencia',
	'Matías',
	'Agustina',
	'Nicolás',
	'Paula',
	'Diego',
	'Julieta',
	'Gonzalo',
	'Micaela',
	'Santiago',
	'Belén',
	'Lautaro',
	'Rocío',
	'Facundo',
	'Antonella',
	'Joaquín',
	'Milagros',
	'Ezequiel'
];
const APELLIDOS = [
	'Fernández',
	'Gómez',
	'Ruiz',
	'Pereyra',
	'Sosa',
	'Acosta',
	'Paz',
	'Luna',
	'Ríos',
	'Martínez',
	'López',
	'Díaz',
	'Romero',
	'Álvarez',
	'Torres',
	'Benítez',
	'Medina',
	'Herrera',
	'Castro',
	'Molina',
	'Rojas',
	'Ortiz',
	'Silva',
	'Núñez',
	'Ramos',
	'Vera',
	'Suárez',
	'Giménez',
	'Cabrera',
	'Ferreyra'
];
const CIUDADES = [
	'CABA',
	'Córdoba',
	'Rosario',
	'Mendoza',
	'La Plata',
	'Mar del Plata',
	'Tucumán',
	'Salta',
	'Neuquén',
	'Posadas',
	'Lanús',
	'Quilmes',
	'San Isidro',
	'Bahía Blanca'
];
/** Caras del handoff (Unsplash). */
const FOTOS = [
	'1531058240690-006c446962d8',
	'1687422811062-a966b55cb217',
	'1764173039506-c5cc7b8ec14b',
	'1520333789090-1afc82db536a',
	'1687422810663-c316494f725a',
	'1556745753-b2904692b3cd',
	'1545262722-9e0d80a0bc01',
	'1562322140-8baeececf3df'
];
const MAYORISTAS = [
	'Distribuidora Sur',
	'Almacén El Puente',
	'Mayorista Norte',
	'Comercial Andina'
];

function elegir<T>(r: () => number, pesos: [T, number][]): T {
	const total = pesos.reduce((s, [, p]) => s + p, 0);
	let x = r() * total;
	for (const [v, p] of pesos) if ((x -= p) < 0) return v;
	return pesos[pesos.length - 1][0];
}

const MIN = 60_000;
const DIA = 1440 * MIN;
const PRIMER_NUMERO = 1048;
/** Días de historia: 30 para mostrar y 30 más para comparar ("vs. mes anterior"). */
export const DIAS_DEMO = 60;

const RESENAS_5 = [
	'Llegó rapidísimo y súper bien embalado. Ya es mi tercera compra.',
	'Muy buena atención por WhatsApp, me ayudaron a elegir.',
	'Excelente calidad, tal cual las fotos.',
	'Todo perfecto, lo recomiendo.',
	'Me encantó. El envío llegó antes de lo que decía.',
	'Es un regalo y quedó hermoso. Gracias por la tarjetita.'
];
const RESENAS_4 = [
	'Muy lindo, aunque tardó un día más de lo previsto.',
	'Buena calidad. El color es un poco distinto al de la foto.'
];

function generar(plantilla: string, base: number): Demo | null {
	const catalogo = catalogoDe(plantilla);
	if (!catalogo) return null;
	const r = azar(hash(plantilla));
	const sinAcento = (s: string) => s.normalize('NFD').replace(/[̀-ͯ]/g, '').toLowerCase();
	const productos = catalogo.productos;

	const clientes: Cliente[] = [];
	const usados = new Set<string>();
	const nuevoCliente = (alta: number): Cliente => {
		let n: string;
		let a: string;
		do {
			n = NOMBRES[Math.floor(r() * NOMBRES.length)];
			a = APELLIDOS[Math.floor(r() * APELLIDOS.length)];
		} while (usados.has(n + a));
		usados.add(n + a);
		const c: Cliente = {
			id: `c${clientes.length + 1}`,
			nombre: `${n} ${a}`,
			email: r() < 0.85 ? `${sinAcento(n)}.${sinAcento(a)}@gmail.com` : null,
			telefono:
				r() < 0.8
					? `+54 9 11 ${String(4000 + Math.floor(r() * 5999))}-${String(1000 + Math.floor(r() * 8999))}`
					: null,
			ciudad: CIUDADES[Math.floor(r() * CIUDADES.length)],
			foto: null,
			mayorista: false,
			alta
		};
		clientes.push(c);
		return c;
	};
	// Los que ya compraban antes de la ventana de la demo.
	for (let i = 0; i < 90; i++) nuevoCliente(base - (DIAS_DEMO + 1 + Math.floor(r() * 400)) * DIA);
	const mayoristas = MAYORISTAS.map((m, i): Cliente => ({
		id: `m${i + 1}`,
		nombre: m,
		email: `compras@${sinAcento(m).replace(/[^a-z]/g, '')}.com.ar`,
		telefono: null,
		ciudad: CIUDADES[i * 3],
		foto: null,
		mayorista: true,
		alta: base - (200 + i * 40) * DIA
	}));

	// Cuándo entró cada pedido: hoy, cada 30 a 75 minutos; antes, unos 20 por día (y un poco menos
	// el mes anterior: la tienda crece).
	const momentos: number[] = [];
	for (let t = base - 4 * MIN; t > base - DIAS_DEMO * DIA;) {
		momentos.push(t);
		const edad = base - t;
		t -= edad < DIA ? (30 + r() * 45) * MIN : (40 + r() * 100) * MIN * (edad > 30 * DIA ? 1.1 : 1);
	}

	// Del más viejo al más nuevo, así los clientes nuevos aparecen con su primera compra.
	const unidad = (p: ProductoEjemplo) => (p.precio >= 100_000 ? 1 : p.precio >= 20_000 ? 2 : 4);
	const pedidos: Pedido[] = [];
	for (const t of momentos.toReversed()) {
		const edad = base - t;
		let canal = elegir<Canal>(r, [
			['web', 30],
			['wa', 26],
			['ml', 18],
			['ig', 12],
			['pos', 18],
			['tg', 4]
		]);
		// El local atiende de 9 a 21; fuera de hora, esa venta entra por la tienda online.
		const horaLocal = ((((t - 3 * 3600_000) % DIA) + DIA) % DIA) / 3600_000;
		if (canal === 'pos' && (horaLocal < 9 || horaLocal >= 21)) canal = 'web';
		const mayorista = canal === 'pos' && r() < 0.15;
		const cliente = mayorista
			? mayoristas[Math.floor(r() * mayoristas.length)]
			: r() < 0.24
				? nuevoCliente(t)
				: clientes[Math.floor(r() ** 1.4 * clientes.length)];
		const cuantos = mayorista
			? 2
			: elegir(r, [
					[1, 55],
					[2, 32],
					[3, 13]
				]);
		const items: Pedido['items'] = [];
		for (let k = 0; k < cuantos; k++) {
			const p = productos[Math.floor(r() * productos.length)];
			if (items.some((x) => x.productoId === p.id)) continue;
			const cantidad = mayorista
				? 6 * (1 + Math.floor(r() * 6))
				: 1 + Math.floor(r() * unidad(p) * 0.5);
			items.push({
				productoId: p.id,
				nombre: p.nombre,
				cantidad,
				precio: mayorista ? Math.round(p.precio * 0.82) : p.precio
			});
		}
		const pago: MedioPago = mayorista
			? 'transf'
			: canal === 'pos'
				? elegir(r, [
						['efectivo', 35],
						['tarjeta', 35],
						['mp', 20],
						['modo', 10]
					])
				: canal === 'ml'
					? 'mp'
					: elegir(r, [
							['mp', 55],
							['transf', 20],
							['modo', 10],
							['payway', 8],
							['uala', 7]
						]);
		const envio: Envio =
			canal === 'pos'
				? mayorista
					? 'moto'
					: 'retiro'
				: elegir(r, [
						['andreani', 45],
						['oca', 15],
						['correo', 12],
						['moto', 12],
						['retiro', 16]
					]);
		// Los recientes están en curso; los viejos, entregados (algunos cancelados).
		let estado: EstadoPedido;
		if (canal === 'pos' && !mayorista) estado = 'Entregado';
		else if (edad < 6 * 3600_000)
			estado = elegir(r, [
				['Para preparar', 50],
				['Listo para despachar', 22],
				['Sin pagar', 14],
				['En camino', 14]
			]);
		else if (edad < 3 * DIA)
			estado = elegir(r, [
				['En camino', 50],
				['Entregado', 35],
				['Para preparar', 10],
				['Cancelado', 5]
			]);
		else
			estado = elegir(r, [
				['Entregado', 94],
				['Cancelado', 6]
			]);
		if (pago === 'transf' && estado === 'Para preparar' && r() < 0.3) estado = 'Sin pagar';
		const estadoPago =
			estado === 'Sin pagar'
				? 'Pendiente'
				: pago === 'transf'
					? 'Acreditada'
					: pago === 'efectivo'
						? 'Cobrado'
						: 'Aprobado';
		pedidos.push({
			numero: 0,
			clienteId: cliente.id,
			canal,
			pago,
			estadoPago,
			envio,
			ciudadEnvio: envio === 'retiro' ? catalogo.ciudad : cliente.ciudad,
			estado,
			facturado: estado !== 'Sin pagar' && estado !== 'Cancelado' && r() < 0.9,
			items,
			total: items.reduce((s, i) => s + i.precio * i.cantidad, 0),
			creado: t
		});
	}
	pedidos.reverse();
	const ultimo = Math.max(PRIMER_NUMERO, pedidos.length);
	pedidos.forEach((p, i) => (p.numero = ultimo - i));
	clientes.push(...mayoristas);

	// Las caras del handoff, para quienes compraron último.
	const porId = new Map(clientes.map((c) => [c.id, c]));
	let caras = 0;
	for (const p of pedidos) {
		const c = porId.get(p.clienteId)!;
		if (caras >= FOTOS.length) break;
		if (!c.mayorista && !c.foto) c.foto = FOTOS[caras++];
	}

	// Stock: uno o dos sin stock y uno con stock bajo (el del aviso).
	const stock: Record<string, number> = {};
	for (const p of productos) stock[p.id] = 4 + Math.floor(r() * 56);
	const bajo = productos[2] ?? productos[0];
	stock[bajo.id] = 3;
	const agotados = 1 + Math.floor(r() * 2);
	for (let k = 0, i = productos.length - 1; k < agotados && i > 2; k++, i -= 2)
		stock[productos[i].id] = 0;
	const enMl = productos.filter(() => r() < 0.75).map((p) => p.id);

	// Variantes: su propio azar, para no correr la semilla de lo demás. El stock del producto es la
	// suma de sus variantes, y cada venta de un producto con opciones lleva su variante (una de las
	// que cuestan lo mismo que el producto: los totales de los pedidos no cambian).
	const variantes = variantesDemo(catalogo);
	for (const [id, v] of variantes) stock[id] = v.variantes.reduce((s, x) => s + x.stock, 0);
	const rv = azar(hash(plantilla + ':variantes'));
	for (const p of pedidos)
		for (const i of p.items) {
			const v = variantes.get(i.productoId);
			const mismas = v?.variantes.filter((x) => x.precio === null) ?? [];
			if (!v || !mismas.length) continue;
			const x = mismas[Math.floor(rv() * mismas.length)];
			i.varianteId = x.id;
			i.variante = textoVariante(v.opciones, x.valores);
		}

	// Carritos abandonados de los últimos 3 días, de gente con WhatsApp.
	const conTelefono = clientes.filter((c) => c.telefono && !c.mayorista);
	const carritos: Carrito[] = Array.from({ length: 12 }, (_, k) => {
		const cliente = conTelefono[Math.floor(r() * conTelefono.length)];
		const items: Carrito['items'] = [];
		for (let i = 0, n = 1 + Math.floor(r() * 2); i < n; i++) {
			const p = productos[Math.floor(r() * productos.length)];
			if (!items.some((x) => x.productoId === p.id))
				items.push({ productoId: p.id, nombre: p.nombre, cantidad: 1, precio: p.precio });
		}
		return {
			id: `k${k + 1}`,
			clienteId: cliente.id,
			items,
			total: items.reduce((s, i) => s + i.precio * i.cantidad, 0),
			creado: base - Math.floor(20 + r() * 3 * 1440) * MIN
		};
	}).sort((a, b) => b.creado - a.creado);

	// Reseñas de pedidos entregados (no en el local), unos días después de comprar.
	const resenas: Resena[] = [];
	for (const p of pedidos) {
		if (p.estado !== 'Entregado' || p.canal === 'pos' || r() > 0.18) continue;
		const creada = p.creado + (2 + r() * 5) * DIA;
		const cinco = r() < 0.86;
		const textos = cinco ? RESENAS_5 : RESENAS_4;
		if (creada >= base) continue;
		resenas.push({
			pedido: p.numero,
			clienteId: p.clienteId,
			productoId: p.items[0].productoId,
			producto: p.items[0].nombre,
			canal: p.canal,
			estrellas: cinco ? 5 : 4,
			texto: textos[Math.floor(r() * textos.length)],
			creada
		});
	}
	resenas.sort((a, b) => b.creada - a.creada);

	// La bandeja: tres conversaciones abiertas con quienes escribieron último por WhatsApp.
	const deWa = [
		...new Set(
			pedidos
				.filter((p) => p.canal === 'wa' && !porId.get(p.clienteId)!.mayorista)
				.map((p) => p.clienteId)
		)
	];
	const pregunta = pedidos.find((p) => p.clienteId === deWa[0])!.items[0].nombre;
	const conversaciones: Conversacion[] = [
		{
			clienteId: deWa[0],
			mensaje: `Hola, ¿tienen stock de ${pregunta}?`,
			creado: base - 4 * MIN,
			atiende: 'hornero',
			sinLeer: 0
		},
		{
			clienteId: deWa[1],
			mensaje: 'Quiero cambiar la dirección de envío',
			creado: base - 12 * MIN,
			atiende: 'vos',
			sinLeer: 2
		},
		{
			clienteId: deWa[2],
			mensaje: '¿Me pasás la lista de precios mayorista?',
			creado: base - 22 * MIN,
			atiende: 'vos',
			sinLeer: 1
		}
	];

	return { catalogo, clientes, pedidos, stock, variantes, enMl, carritos, resenas, conversaciones };
}

// Por isolate: generar es barato (unos 600 pedidos), pero se repite en cada pedido del panel.
const cache = new Map<string, Demo | null>();

/** Los datos de ejemplo de una plantilla, fechados respecto de `base` (creación del sandbox). */
export function demoDe(plantilla: string, base: number): Demo | null {
	const clave = `${plantilla}:${base}`;
	if (!cache.has(clave)) {
		if (cache.size > 50) cache.clear();
		cache.set(clave, generar(plantilla, base));
	}
	return cache.get(clave)!;
}
