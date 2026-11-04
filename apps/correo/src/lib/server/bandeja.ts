/**
 * La bandeja: carpetas, lista y lectura, juntando las cuentas de todas las casillas de la persona.
 * Los ids que salen de acá llevan la cuenta adelante ("<cuenta>~<email>"), porque un correo de una
 * casilla compartida vive en otra cuenta JMAP.
 */
import { Jmap, MAX_LLAMADAS, QUOTA, SUBMISSION, type Cuenta } from './jmap.ts';
import { sanear } from './html.ts';
import { esEtiqueta } from '../etiquetas.ts';

export const CARPETAS = [
	{ id: 'entrada', l: 'Entrada', rol: 'inbox' },
	{ id: 'destacados', l: 'Destacados', rol: null },
	{ id: 'enviados', l: 'Enviados', rol: 'sent' },
	{ id: 'borradores', l: 'Borradores', rol: 'drafts' },
	{ id: 'programados', l: 'Programados', rol: null },
	{ id: 'archivo', l: 'Archivo', rol: 'archive' },
	{ id: 'spam', l: 'Spam', rol: 'junk' },
	{ id: 'papelera', l: 'Papelera', rol: 'trash' }
] as const;
export type CarpetaId = (typeof CARPETAS)[number]['id'];
export type Pestana = 'todos' | 'sinleer';

type Buzon = {
	id: string;
	name: string;
	role: string | null;
	totalEmails: number;
	unreadEmails: number;
};
type Direccion = { name: string | null; email: string };
type EmailJmap = {
	id: string;
	threadId: string;
	mailboxIds: Record<string, boolean>;
	keywords: Record<string, boolean>;
	from: Direccion[] | null;
	to: Direccion[] | null;
	cc: Direccion[] | null;
	subject: string | null;
	preview: string;
	receivedAt: string;
	sentAt?: string | null;
	hasAttachment: boolean;
	size: number;
	messageId?: string[] | null;
	references?: string[] | null;
};

export type Casilla = {
	cuenta: string;
	email: string;
	nombre: string;
	personal: boolean;
	sinLeer: number;
};

export type Fila = {
	id: string;
	cuenta: string;
	casilla: string;
	nombre: string;
	email: string;
	asunto: string;
	vista: string;
	fecha: string;
	leido: boolean;
	destacado: boolean;
	adjuntos: boolean;
	enHilo: number;
	etiquetas: string[];
	/** Programado: el envío pendiente (para cancelarlo) y cuándo sale. */
	envio?: { id: string; sale: string };
};

export const SEP = '~';
export const idDe = (cuenta: string, email: string) => `${cuenta}${SEP}${email}`;
export function partir(id: string): [string, string] {
	const i = id.indexOf(SEP);
	if (i < 1) throw new Error('id inválido');
	return [id.slice(0, i), id.slice(i + 1)];
}

/** "ventas@grupovimo.com.ar" → "Ventas"; la propia, el nombre de la persona con "(vos)". */
function nombreCasilla(email: string, personal: boolean): string {
	const local = email.split('@')[0];
	const n = local.charAt(0).toUpperCase() + local.slice(1);
	return personal ? `${n} (vos)` : n;
}

async function buzones(j: Jmap, cuentas: Cuenta[]) {
	const tandas: Cuenta[][] = [];
	for (let i = 0; i < cuentas.length; i += MAX_LLAMADAS)
		tandas.push(cuentas.slice(i, i + MAX_LLAMADAS));
	const out = new Map<string, Buzon[]>();
	await Promise.all(
		tandas.map(async (t) => {
			const r = await j.llamar(
				t.map((c, i) => [
					'Mailbox/get',
					{ accountId: c.id, properties: ['name', 'role', 'totalEmails', 'unreadEmails'] },
					`b${i}`
				])
			);
			t.forEach((c, i) => out.set(c.id, (r.get(`b${i}`)?.list ?? []) as Buzon[]));
		})
	);
	return out;
}

const deRol = (bs: Buzon[], rol: string) => bs.find((b) => b.role === rol);

/** Stalwart no deja ver los envíos de una casilla personal compartida: se toma como "sin envíos". */
function sinEnvios(e: unknown): null {
	if (e instanceof Error && e.message.includes('accountNotSupportedByMethod')) return null;
	throw e;
}

/** El filtro de una carpeta en una cuenta (null: esa cuenta no tiene esa carpeta). */
function filtroDe(carpeta: CarpetaId, bs: Buzon[], pestana: Pestana) {
	const conds: Record<string, unknown>[] = [];
	const c = CARPETAS.find((x) => x.id === carpeta)!;
	if (c.rol) {
		const b = deRol(bs, c.rol);
		if (!b) return null;
		conds.push({ inMailbox: b.id });
	} else if (carpeta === 'destacados') {
		conds.push({ hasKeyword: '$flagged' });
		const fuera = ['trash', 'junk'].map((r) => deRol(bs, r)?.id).filter(Boolean);
		if (fuera.length) conds.push({ inMailboxOtherThan: fuera });
	} else return null;
	if (pestana === 'sinleer') conds.push({ notKeyword: '$seen' });
	return conds.length === 1 ? conds[0] : { operator: 'AND', conditions: conds };
}

export type Bandeja = {
	yo: { email: string; nombre: string };
	casillas: Casilla[];
	carpetas: { id: CarpetaId; l: string; n: number }[];
	filas: Fila[];
	contadores: { sinleer: number };
	espacio: { usado: number; total: number } | null;
};

export async function bandeja(
	j: Jmap,
	o: { carpeta: CarpetaId; casilla: string | null; pestana: Pestana; buscar: string }
): Promise<Bandeja> {
	const todas = await j.cuentas();
	const bs = await buzones(j, todas);
	const casillas: Casilla[] = todas.map((c) => ({
		cuenta: c.id,
		email: c.email,
		nombre: nombreCasilla(c.email, c.personal),
		personal: c.personal,
		sinLeer: deRol(bs.get(c.id)!, 'inbox')?.unreadEmails ?? 0
	}));
	const elegidas = todas.filter((c) => !o.casilla || c.id === o.casilla);
	const suma = (rol: string, k: 'unreadEmails' | 'totalEmails') =>
		elegidas.reduce((n, c) => n + (deRol(bs.get(c.id)!, rol)?.[k] ?? 0), 0);

	// Un pedido por cuenta (la consulta, sus hilos y los contadores de pestaña), todos a la vez:
	// Stalwart limita las llamadas por pedido.
	const texto = o.buscar.trim();
	const porCuenta = await Promise.all(
		elegidas.map(async (c) => {
			const b = bs.get(c.id)!;
			// Programados: los envíos que esperan su hora (EmailSubmission pendientes) y sus correos.
			if (o.carpeta === 'programados') {
				const r = await j
					.llamar(
						[
							[
								'EmailSubmission/query',
								{ accountId: c.id, filter: { undoStatus: 'pending' } },
								'sq'
							],
							[
								'EmailSubmission/get',
								{
									accountId: c.id,
									'#ids': { resultOf: 'sq', name: 'EmailSubmission/query', path: '/ids' }
								},
								'sg'
							],
							[
								'Email/get',
								{
									accountId: c.id,
									'#ids': { resultOf: 'sg', name: 'EmailSubmission/get', path: '/list/*/emailId' },
									properties: [
										'threadId',
										'keywords',
										'from',
										'to',
										'subject',
										'preview',
										'receivedAt',
										'hasAttachment'
									]
								},
								'g'
							]
						],
						[SUBMISSION]
					)
					.catch(sinEnvios);
				if (!r) return null;
				return { c, r, programados: true };
			}
			const f = filtroDe(o.carpeta, b, o.pestana);
			if (!f) return null;
			const filtro = texto ? { operator: 'AND', conditions: [f, { text: texto }] } : f;
			const llamadas: [string, Record<string, unknown>, string][] = [
				[
					'Email/query',
					{
						accountId: c.id,
						filter: filtro,
						sort: [{ property: 'receivedAt', isAscending: false }],
						collapseThreads: true,
						limit: 60
					},
					'q'
				],
				[
					'Email/get',
					{
						accountId: c.id,
						'#ids': { resultOf: 'q', name: 'Email/query', path: '/ids' },
						properties: [
							'threadId',
							'keywords',
							'from',
							'to',
							'subject',
							'preview',
							'receivedAt',
							'hasAttachment'
						]
					},
					'g'
				],
				[
					'Thread/get',
					{
						accountId: c.id,
						'#ids': { resultOf: 'g', name: 'Email/get', path: '/list/*/threadId' }
					},
					't'
				]
			];
			for (const [k, p] of [['s', 'sinleer']] as const) {
				const fp = filtroDe(o.carpeta, b, p);
				if (fp)
					llamadas.push([
						'Email/query',
						{ accountId: c.id, filter: fp, collapseThreads: true, calculateTotal: true, limit: 0 },
						k
					]);
			}
			// La cuota (el espacio usado) va en el pedido de la casilla propia: Stalwart acepta hasta
			// cuatro pedidos a la vez.
			const propia = c.id === (todas.find((x) => x.personal) ?? todas[0]).id;
			if (propia) llamadas.push(['Quota/get', { accountId: c.id, ids: null }, 'cuota']);
			// Cuántos envíos esperan su hora (el número de la carpeta Programados).
			llamadas.push([
				'EmailSubmission/query',
				{ accountId: c.id, filter: { undoStatus: 'pending' }, calculateTotal: true, limit: 0 },
				'pend'
			]);
			const extra = propia ? [QUOTA, SUBMISSION] : [SUBMISSION];
			// Una casilla personal compartida no deja ver sus envíos: va sin el contador.
			const r = await j.llamar(llamadas, extra).catch((e) =>
				sinEnvios(e) === null
					? j.llamar(
							llamadas.filter((x) => x[2] !== 'pend'),
							extra
						)
					: null
			);
			return {
				c,
				r: r!,
				programados: false
			};
		})
	);

	const filas: Fila[] = [];
	const contadores = { sinleer: 0 };
	let pendientes = 0;
	for (const x of porCuenta) {
		if (!x) continue;
		const { c, r } = x;
		pendientes += Number(r.get('pend')?.total ?? 0);
		if (x.programados) {
			const envios = (r.get('sg')?.list ?? []) as { id: string; emailId: string; sendAt: string }[];
			pendientes += envios.length;
			const correos = new Map(((r.get('g')?.list ?? []) as EmailJmap[]).map((e) => [e.id, e]));
			for (const s of envios) {
				const e = correos.get(s.emailId);
				if (!e) continue;
				const para = e.to?.[0] ?? { name: null, email: '' };
				filas.push({
					id: idDe(c.id, e.id),
					cuenta: c.id,
					casilla: c.email,
					nombre: para.name || para.email || '(sin destinatario)',
					email: para.email,
					asunto: e.subject || '(sin asunto)',
					vista: e.preview,
					fecha: s.sendAt,
					leido: true,
					destacado: false,
					adjuntos: e.hasAttachment,
					enHilo: 1,
					etiquetas: [],
					envio: { id: s.id, sale: s.sendAt }
				});
			}
			continue;
		}
		const hilos = new Map(
			((r.get('t')?.list ?? []) as { id: string; emailIds: string[] }[]).map((t) => [
				t.id,
				t.emailIds.length
			])
		);
		for (const e of (r.get('g')?.list ?? []) as EmailJmap[]) {
			const de = (o.carpeta === 'enviados' || o.carpeta === 'borradores'
				? e.to?.[0]
				: e.from?.[0]) ?? { name: null, email: '' };
			filas.push({
				id: idDe(c.id, e.id),
				cuenta: c.id,
				casilla: c.email,
				nombre: de.name || de.email.split('@')[0] || '(sin remitente)',
				email: de.email,
				asunto: e.subject || '(sin asunto)',
				vista: e.preview,
				fecha: e.receivedAt,
				leido: !!e.keywords.$seen,
				destacado: !!e.keywords.$flagged,
				adjuntos: e.hasAttachment,
				enHilo: hilos.get(e.threadId) ?? 1,
				etiquetas: Object.keys(e.keywords).filter(esEtiqueta)
			});
		}
		contadores.sinleer += Number(r.get('s')?.total ?? 0);
	}
	// Programados: el próximo en salir arriba; el resto, lo último que llegó arriba.
	filas.sort((a, b) =>
		o.carpeta === 'programados' ? a.fecha.localeCompare(b.fecha) : b.fecha.localeCompare(a.fecha)
	);

	const q = (
		(porCuenta.find((x) => x?.r.has('cuota'))?.r.get('cuota')?.list ?? []) as {
			resourceType: string;
			used: number;
			hardLimit: number;
		}[]
	).find((x) => x.resourceType === 'octets');
	return {
		yo: { email: j.cred.email, nombre: nombreCasilla(j.cred.email, false) },
		casillas,
		carpetas: CARPETAS.map((c) => ({
			id: c.id,
			l: c.l,
			n:
				c.id === 'entrada'
					? suma('inbox', 'unreadEmails')
					: c.id === 'borradores'
						? suma('drafts', 'totalEmails')
						: c.id === 'spam'
							? suma('junk', 'unreadEmails')
							: c.id === 'programados'
								? pendientes
								: 0
		})),
		filas: filas.slice(0, 80),
		contadores,
		espacio: q ? { usado: q.used, total: q.hardLimit } : null
	};
}

export type Adjunto = {
	parte: string;
	blob: string;
	nombre: string;
	tipo: string;
	tamano: number;
	cid: string | null;
};
export type Mensaje = {
	id: string;
	cuenta: string;
	casilla: string;
	asunto: string;
	de: Direccion;
	para: Direccion[];
	cc: Direccion[];
	fecha: string;
	destacado: boolean;
	html: string | null;
	texto: string;
	remotas: boolean;
	adjuntos: Adjunto[];
	anteriores: { id: string; de: Direccion; fecha: string; vista: string }[];
	messageId: string | null;
	referencias: string[];
	etiquetas: string[];
	/** Las carpetas de su casilla, para «Mover a». */
	carpetas: { id: string; nombre: string; rol: string | null }[];
};

type Parte = {
	partId: string;
	blobId: string;
	name: string | null;
	type: string;
	size: number;
	cid: string | null;
	disposition: string | null;
};

export async function mensaje(j: Jmap, id: string, conImagenes: boolean): Promise<Mensaje | null> {
	const [cuenta, eid] = partir(id);
	const cuentas = await j.cuentas();
	const c = cuentas.find((x) => x.id === cuenta);
	if (!c) return null;
	const r = await j.llamar([
		[
			'Email/get',
			{
				accountId: cuenta,
				ids: [eid],
				properties: [
					'threadId',
					'keywords',
					'from',
					'to',
					'cc',
					'subject',
					'receivedAt',
					'htmlBody',
					'textBody',
					'attachments',
					'bodyValues',
					'messageId',
					'references'
				],
				fetchHTMLBodyValues: true,
				fetchTextBodyValues: true,
				bodyProperties: ['partId', 'blobId', 'name', 'type', 'size', 'cid', 'disposition'],
				maxBodyValueBytes: 2_000_000
			},
			'm'
		],
		['Mailbox/get', { accountId: cuenta, properties: ['name', 'role'] }, 'c'],
		[
			'Thread/get',
			{ accountId: cuenta, '#ids': { resultOf: 'm', name: 'Email/get', path: '/list/*/threadId' } },
			't'
		],
		[
			'Email/get',
			{
				accountId: cuenta,
				'#ids': { resultOf: 't', name: 'Thread/get', path: '/list/*/emailIds' },
				properties: ['from', 'receivedAt', 'preview']
			},
			'h'
		]
	]);
	const e = (
		(r.get('m')?.list ?? []) as (EmailJmap & {
			htmlBody: Parte[];
			textBody: Parte[];
			attachments: Parte[];
			bodyValues: Record<string, { value: string }>;
		})[]
	)[0];
	if (!e) return null;
	// Que quede leído al abrirlo.
	if (!e.keywords.$seen)
		await j.llamar([
			['Email/set', { accountId: cuenta, update: { [eid]: { 'keywords/$seen': true } } }, 's']
		]);

	const valor = (ps: Parte[]) => ps.map((p) => e.bodyValues[p.partId]?.value ?? '').join('\n');
	const esHtml = e.htmlBody.some((p) => p.type === 'text/html');
	const adjuntos: Adjunto[] = e.attachments.map((p) => ({
		parte: p.partId,
		blob: p.blobId,
		nombre: p.name || 'adjunto',
		tipo: p.type,
		tamano: p.size,
		cid: p.cid
	}));
	const urlAdj = (a: Adjunto) =>
		`/adjunto/${encodeURIComponent(cuenta)}/${encodeURIComponent(a.blob)}?n=${encodeURIComponent(a.nombre)}&t=${encodeURIComponent(a.tipo)}`;
	const cids = new Map(
		adjuntos.filter((a) => a.cid).map((a) => [a.cid!.replace(/^<|>$/g, ''), urlAdj(a)])
	);
	const s = esHtml ? sanear(valor(e.htmlBody), { cids, remotas: conImagenes }) : null;
	const hilo = ((r.get('h')?.list ?? []) as EmailJmap[]).filter((x) => x.id !== eid);
	return {
		id,
		cuenta,
		casilla: c.email,
		asunto: e.subject || '(sin asunto)',
		de: e.from?.[0] ?? { name: null, email: '' },
		para: e.to ?? [],
		cc: e.cc ?? [],
		fecha: e.receivedAt,
		destacado: !!e.keywords.$flagged,
		html: s?.html ?? null,
		texto: valor(e.textBody),
		remotas: s?.remotas ?? false,
		// Las imágenes de adentro del correo (cid) no se listan como adjuntos.
		adjuntos: adjuntos.filter((a) => !(a.cid && s?.usados.has(a.cid.replace(/^<|>$/g, '')))),
		anteriores: hilo
			.filter((x) => x.receivedAt < e.receivedAt)
			.map((x) => ({
				id: idDe(cuenta, x.id),
				de: x.from?.[0] ?? { name: null, email: '' },
				fecha: x.receivedAt,
				vista: x.preview
			})),
		messageId: e.messageId?.[0] ?? null,
		referencias: e.references ?? [],
		etiquetas: Object.keys(e.keywords).filter(esEtiqueta),
		carpetas: ((r.get('c')?.list ?? []) as { id: string; name: string; role: string | null }[])
			.filter((b) => b.role !== 'drafts')
			.map((b) => ({ id: b.id, nombre: b.name, rol: b.role }))
	};
}

/** Mover a una carpeta (archivar, borrar, spam) o marcar. Archivo se crea si la cuenta no lo tiene. */
export async function accion(
	j: Jmap,
	ids: string[],
	que:
		| 'archivar'
		| 'borrar'
		| 'destacar'
		| 'quitar-destacado'
		| 'no-leido'
		| 'leido'
		| 'spam'
		| 'entrada'
		| 'etiquetar'
		| 'desetiquetar'
		| 'mover'
		| 'cancelar-envio',
	/** La etiqueta (keyword) o la carpeta (id) de etiquetar y mover. */
	valor = ''
) {
	const porCuenta = new Map<string, string[]>();
	for (const id of ids) {
		const [c, e] = partir(id);
		porCuenta.set(c, [...(porCuenta.get(c) ?? []), e]);
	}
	const cuentas = [...porCuenta.keys()];
	const bs = await buzones(
		j,
		(await j.cuentas()).filter((c) => cuentas.includes(c.id))
	);
	for (const [cuenta, emails] of porCuenta) {
		const b = bs.get(cuenta) ?? [];
		let cambio: Record<string, unknown>;
		if (que === 'cancelar-envio') {
			// El envío programado se cancela y el correo vuelve a Borradores, para editarlo o mandarlo.
			const borradores = deRol(b, 'drafts')?.id;
			const r = await j.llamar(
				[
					[
						'EmailSubmission/set',
						{ accountId: cuenta, update: { [valor]: { undoStatus: 'canceled' } } },
						's'
					],
					[
						'Email/set',
						{
							accountId: cuenta,
							update: Object.fromEntries(
								emails.map((e) => [
									e,
									{
										mailboxIds: borradores ? { [borradores]: true } : undefined,
										'keywords/$draft': true
									}
								])
							)
						},
						'e'
					]
				],
				[SUBMISSION]
			);
			if (r.get('s')?.notUpdated) throw new Error('Ya salió o no se puede cancelar.');
			continue;
		}
		if (que === 'etiquetar' || que === 'desetiquetar') {
			if (!esEtiqueta(valor)) throw new Error('Etiqueta inválida.');
			cambio = { [`keywords/${valor}`]: que === 'etiquetar' ? true : null };
		} else if (que === 'mover') {
			if (!b.some((x) => x.id === valor)) throw new Error('Esa carpeta no es de esta casilla.');
			cambio = { mailboxIds: { [valor]: true } };
		} else if (que === 'borrar' && deRol(b, 'trash')) {
			// Lo que ya está en la papelera se elimina de verdad (antes volvía a la papelera y no pasaba nada).
			const papelera = deRol(b, 'trash')!.id;
			const g = await j.llamar([
				['Email/get', { accountId: cuenta, ids: emails, properties: ['mailboxIds'] }, 'g']
			]);
			const lista = (g.get('g')?.list ?? []) as {
				id: string;
				mailboxIds: Record<string, boolean>;
			}[];
			const yaEn = lista.filter((x) => x.mailboxIds[papelera]).map((x) => x.id);
			const mover = emails.filter((e) => !yaEn.includes(e));
			await j.llamar([
				[
					'Email/set',
					{
						accountId: cuenta,
						destroy: yaEn,
						update: Object.fromEntries(mover.map((e) => [e, { mailboxIds: { [papelera]: true } }]))
					},
					'u'
				]
			]);
			continue;
		} else if (que === 'destacar' || que === 'quitar-destacado')
			cambio = { 'keywords/$flagged': que === 'destacar' ? true : null };
		else if (que === 'leido' || que === 'no-leido')
			cambio = { 'keywords/$seen': que === 'leido' ? true : null };
		else {
			const rol = { archivar: 'archive', borrar: 'trash', spam: 'junk', entrada: 'inbox' }[que];
			let destino = deRol(b, rol)?.id;
			if (!destino && rol === 'archive') {
				const r = await j.llamar([
					[
						'Mailbox/set',
						{ accountId: cuenta, create: { a: { name: 'Archivo', role: 'archive' } } },
						'c'
					]
				]);
				destino = ((r.get('c')?.created as Record<string, { id: string }>) ?? {}).a?.id;
			}
			if (!destino) throw new Error(`Esta casilla no tiene ${rol}`);
			cambio = { mailboxIds: { [destino]: true } };
		}
		await j.llamar([
			[
				'Email/set',
				{ accountId: cuenta, update: Object.fromEntries(emails.map((e) => [e, cambio])) },
				'u'
			]
		]);
	}
}

/**
 * Vacía la papelera de las casillas (o de una): elimina de verdad, de a 500 por vuelta. Devuelve
 * cuántos correos se borraron.
 */
export async function vaciarPapelera(j: Jmap, casilla: string | null): Promise<number> {
	const cuentas = (await j.cuentas()).filter((c) => !casilla || c.id === casilla);
	const bs = await buzones(j, cuentas);
	let total = 0;
	for (const c of cuentas) {
		const papelera = deRol(bs.get(c.id) ?? [], 'trash')?.id;
		if (!papelera) continue;
		for (let vuelta = 0; vuelta < 40; vuelta++) {
			const q = await j.llamar([
				['Email/query', { accountId: c.id, filter: { inMailbox: papelera }, limit: 500 }, 'q']
			]);
			const ids = (q.get('q')?.ids ?? []) as string[];
			if (!ids.length) break;
			await j.llamar([['Email/set', { accountId: c.id, destroy: ids }, 'd']]);
			total += ids.length;
		}
	}
	return total;
}
