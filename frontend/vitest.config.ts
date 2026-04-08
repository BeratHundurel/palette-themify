import viteConfig from './vite.config';
import { defineConfig, mergeConfig } from 'vitest/config';

export default mergeConfig(
	viteConfig,
	defineConfig({
		resolve: {
			conditions: ['browser']
		},
		test: {
			environment: 'jsdom',
			globals: true,
			setupFiles: ['./src/lib/test/setup.ts'],
			include: ['src/**/*.{test,spec}.{ts,js}'],
			coverage: {
				provider: 'v8',
				reporter: ['text', 'json-summary', 'lcov'],
				reportsDirectory: './artifacts/test-coverage'
			}
		}
	})
);
