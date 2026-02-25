import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import tailwindcss from '@tailwindcss/vite'
import svg from '@poppanator/sveltekit-svg'
import { realpathSync } from 'fs'
import { resolve, dirname } from 'path'
import { fileURLToPath } from 'url'

const __dirname = dirname(fileURLToPath(import.meta.url))

function resolveReal(rel: string): string {
	try { return realpathSync(resolve(__dirname, rel)) } catch { return resolve(__dirname, rel) }
}

const flowbiteDist = resolveReal('node_modules/flowbite-svelte/dist')

const resolvePnpmSources = {
	name: 'resolve-pnpm-sources',
	enforce: 'pre' as const,
	transform(code: string, id: string) {
		if (!id.endsWith('app.css')) return
		return { code: code.replace(/@source\s+"[^"]*flowbite-svelte[^"]*"/g, `@source "${flowbiteDist}"`) }
	}
}

export default defineConfig({
	define: {
		VERSION: JSON.stringify(process.env.VERSION ?? 'dev'),
	},
	plugins: [
		resolvePnpmSources,
		tailwindcss(),
		sveltekit(),
		svg({
			includePaths: ['./src/icons/']
		})
	]
});
