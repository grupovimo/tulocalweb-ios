import { readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import { describe, expect, it } from 'vitest';
import {
	CLAVE,
	CLAVE_DOCUMENTO,
	LIMITES_DOCUMENTO,
	subirDocumento,
	tipoDe,
	tipoDocumentoDe
} from './archivos.ts';
import { d1Local } from './sandbox/d1-local.ts';
import { dbSandbox } from './sandbox/sandbox.ts';

const bytes = (...xs: (number | string)[]) =>
	new Uint8Array(
		xs.flatMap((x) => (typeof x === 'string' ? [...x].map((c) => c.charCodeAt(0)) : [x]))
	);

/** Un ZIP con su primer archivo `mimetype` (cabecera local de 30 bytes + nombre + contenido). */
const zip = (nombre: string, contenido: string) =>
	bytes(
		0x50,
		0x4b,
		0x03,
		0x04,
		...new Array(22).fill(0),
		nombre.length,
		0,
		0,
		0,
		nombre,
		contenido
	);

const dir = join(import.meta.dirname, '../../../migraciones-sandbox');
const migraciones = readdirSync(dir)
	.filter((f) => f.endsWith('.sql'))
	.sort()
	.map((f) => readFileSync(join(dir, f), 'utf8'));
const nueva = () => dbSandbox(d1Local(migraciones) as unknown as D1Database);
const TIENDA = '7d3e2f10-aaaa-4bbb-8ccc-0123456789ab';
/** Un R2 en memoria: guarda lo que llega con su metadata. */
function enMemoria() {
	const guardado = new Map<string, { datos: Uint8Array; tipo?: string }>();
	return {
		guardado,
		async put(clave: string, datos: unknown, o?: { httpMetadata?: { contentType?: string } }) {
			guardado.set(clave, { datos: datos as Uint8Array, tipo: o?.httpMetadata?.contentType });
			return null;
		}
	} as unknown as Pick<R2Bucket, 'put'> & { guardado: typeof guardado };
}
const PDF = bytes('%PDF-1.7\n', '1 0 obj << >> endobj\n%%EOF');

/** El status de un error() de SvelteKit. */
async function status(p: Promise<unknown>) {
	try {
		await p;
		return 0;
	} catch (e) {
		return (e as { status?: number }).status ?? -1;
	}
}

describe('archivos', () => {
	it('reconoce el formato por la firma, no por lo que dice el navegador', () => {
		expect(tipoDe(bytes('RIFF', 0, 0, 0, 0, 'WEBPVP8 '))).toBe('image/webp');
		expect(tipoDe(bytes(0xff, 0xd8, 0xff, 0xe0, 0))).toBe('image/jpeg');
		expect(tipoDe(bytes(0x89, 'PNG', 13, 10, 26, 10, 0))).toBe('image/png');
		expect(tipoDe(bytes('<svg xmlns="http://www.w3.org/2000/svg">'))).toBeNull();
		expect(tipoDe(bytes('GIF89a', 0, 0, 0, 0))).toBeNull();
	});

	it('sólo sirve claves de fotos, sin salir de la carpeta', () => {
		const ok =
			'fotos/7d3e2f10-aaaa-4bbb-8ccc-0123456789ab/0b7f3c1e-1a2b-4c3d-8e9f-001122334455.webp';
		expect(CLAVE.test(ok)).toBe(true);
		expect(CLAVE.test('fotos/../secreto.webp')).toBe(false);
		expect(CLAVE.test(ok.replace('fotos/', 'otra/'))).toBe(false);
		// El id del sandbox (la cookie, base64url) nunca puede ser parte de una clave.
		expect(CLAVE.test('fotos/Ab_cD-ef/0b7f3c1e-1a2b-4c3d-8e9f-001122334455.webp')).toBe(false);
		expect(CLAVE.test(ok.replace('.webp', '.svg'))).toBe(false);
	});
});

describe('archivos para descargar', () => {
	it('PDF, EPUB y ZIP por los primeros bytes; el resto no', () => {
		expect(tipoDocumentoDe(PDF)).toBe('application/pdf');
		expect(tipoDocumentoDe(zip('mimetype', 'application/epub+zip'))).toBe('application/epub+zip');
		expect(tipoDocumentoDe(zip('guia.txt', 'hola'))).toBe('application/zip');
		// Un HTML que dice ser PDF, un ejecutable, una foto: no.
		expect(tipoDocumentoDe(bytes('<html><body>%PDF-'))).toBeNull();
		expect(tipoDocumentoDe(bytes('MZ', 0x90, 0, 3, 0))).toBeNull();
		expect(tipoDocumentoDe(bytes(0xff, 0xd8, 0xff, 0xe0, 0))).toBeNull();
		expect(tipoDocumentoDe(bytes('%PD'))).toBeNull();
	});

	it('las claves de archivos no se mezclan con las de fotos', () => {
		const ok = `archivos/${TIENDA}/0b7f3c1e-1a2b-4c3d-8e9f-001122334455.pdf`;
		expect(CLAVE_DOCUMENTO.test(ok)).toBe(true);
		expect(CLAVE_DOCUMENTO.test(ok.replace('.pdf', '.epub'))).toBe(true);
		expect(CLAVE_DOCUMENTO.test(ok.replace('.pdf', '.zip'))).toBe(true);
		expect(CLAVE_DOCUMENTO.test(ok.replace('.pdf', '.html'))).toBe(false);
		expect(CLAVE_DOCUMENTO.test('archivos/../secreto.pdf')).toBe(false);
		expect(CLAVE.test(ok)).toBe(false);
	});

	it('sube un PDF tal cual y lo deja en archivos/<tienda>/ con su tipo', async () => {
		const db = nueva();
		const r2 = enMemoria();
		const clave = await subirDocumento(db, TIENDA, PDF, null, r2);
		expect(CLAVE_DOCUMENTO.test(clave)).toBe(true);
		expect(clave.startsWith(`archivos/${TIENDA}/`)).toBe(true);
		expect(r2.guardado.get(clave)).toEqual({ datos: PDF, tipo: 'application/pdf' });
		// Dos subidas, dos URL distintas: ninguna pisa a otra.
		expect(await subirDocumento(db, TIENDA, PDF, null, r2)).not.toBe(clave);
	});

	it('rechaza lo que no es PDF, EPUB ni ZIP y lo que pasa de 20 MB', async () => {
		const db = nueva();
		const r2 = enMemoria();
		const html = bytes('<script>alert(1)</script>');
		expect(await status(subirDocumento(db, TIENDA, html, null, r2))).toBe(415);
		const grande = new Uint8Array(LIMITES_DOCUMENTO.archivo + 1);
		grande.set(PDF);
		expect(await status(subirDocumento(db, TIENDA, grande, null, r2))).toBe(413);
		expect(r2.guardado.size).toBe(0);
	});

	it('el tope por tienda cuenta sólo los archivos, no las fotos', async () => {
		const d1 = d1Local(migraciones) as unknown as D1Database;
		const db = dbSandbox(d1);
		const ahora = Date.now();
		// Fotos que ya llenan su propio tope: no frenan un archivo.
		await d1
			.prepare('insert into archivos (clave, duenio, bytes, creado_en) values (?, ?, ?, ?)')
			.bind(`fotos/${TIENDA}/a.webp`, TIENDA, 15 * 1024 * 1024, ahora)
			.run();
		const r2 = enMemoria();
		expect(await status(subirDocumento(db, TIENDA, PDF, null, r2))).toBe(0);
		// Archivos que ya llenan los 60 MB: el siguiente no entra.
		await d1
			.prepare('insert into archivos (clave, duenio, bytes, creado_en) values (?, ?, ?, ?)')
			.bind(`archivos/${TIENDA}/b.pdf`, TIENDA, LIMITES_DOCUMENTO.porDuenio, ahora)
			.run();
		expect(await status(subirDocumento(db, TIENDA, PDF, null, r2))).toBe(413);
		// Otra tienda sí puede.
		const otra = '8d3e2f10-aaaa-4bbb-8ccc-0123456789ab';
		expect(await status(subirDocumento(db, otra, PDF, null, r2))).toBe(0);
	});
});
