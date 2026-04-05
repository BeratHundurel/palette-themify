import adapter from '@sveltejs/adapter-node';
import adapterStatic from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

const isDesktopBuild = process.env.SVELTEKIT_DESKTOP === '1';
const desktopDistDir = process.env.DESKTOP_DIST_DIR;

const selectedAdapter = isDesktopBuild
	? adapterStatic({
			pages: desktopDistDir || 'dist',
			assets: desktopDistDir || 'dist',
			fallback: 'index.html',
			precompress: false,
			strict: false
		})
	: adapter();

const config = {
	preprocess: vitePreprocess(),
	kit: { adapter: selectedAdapter }
};

export default config;
