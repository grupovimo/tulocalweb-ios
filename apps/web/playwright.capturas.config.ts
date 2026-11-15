import { defineConfig } from '@playwright/test';

// Capturas de las vitrinas (workflow «Capturas»): el mismo preview que los e2e, un solo navegador.
// El ancho lo pone cada captura; el proyecto arranca en escritorio.
export default defineConfig({
	testDir: 'e2e',
	testMatch: '**/*.cap.ts',
	timeout: 600_000,
	workers: 4,
	retries: 0,
	reporter: 'list',
	use: { baseURL: process.env.E2E_BASE ?? 'http://localhost:4173' },
	webServer: process.env.E2E_BASE
		? undefined
		: [{ command: 'bun run preview', url: 'http://localhost:4173', reuseExistingServer: true }],
	projects: [{ name: 'capturas', use: { viewport: { width: 1440, height: 900 } } }]
});
