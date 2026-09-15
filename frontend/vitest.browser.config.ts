import { defineConfig } from 'vitest/config';
import { playwright } from '@vitest/browser-playwright';
import { svelte } from '@sveltejs/vite-plugin-svelte';

export default defineConfig({
	plugins: [svelte()],
	test: {
		include: ['src/**/*.browser.test.ts'],
		browser: {
			enabled: true,
			provider: playwright({ launch: { headless: true } }),
			instances: [{ browser: 'chromium' }]
		}
	}
});
