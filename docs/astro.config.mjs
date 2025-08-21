// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// https://astro.build/config
export default defineConfig({
	site: 'https://jpd-docs.onrender.com',
	integrations: [
		starlight({
			title: 'javascript-package-delegator (jpd)',
			description: 'A universal JavaScript package manager CLI written in Go that delegates to npm, Yarn, pnpm, Bun, and Deno.',
			sidebar: [
				{
					label: 'Introduction',
					items: [{ label: 'Overview', link: '/' }]
				},
				{
					label: 'Tutorial',
					items: [{ label: 'Getting started', link: '/tutorial/getting-started' }]
				},
				{
					label: 'How-to',
					items: [
						{ label: 'Run in a subproject with --cwd', link: '/how-to/run-in-subproject' },
						{ label: 'Interactive install & uninstall', link: '/how-to/interactive' },
						{ label: 'Override or choose the agent', link: '/how-to/agent' },
						{ label: 'Use Volta automatically', link: '/how-to/volta' },
						{ label: 'Enable shell completion', link: '/how-to/completion' }
					]
				},
				{
					label: 'Reference',
					items: [
						{ label: 'Commands (all-in-one)', link: '/reference/commands' }
					]
				},
				{
					label: 'Explanation',
					items: [
						{ label: 'Mental model (Why / What / How)', link: '/explanation/mental-model' }
					]
				}
			],
			editLink: {
				baseUrl: 'https://github.com/louiss0/javascript-package-delegator/edit/main/docs/src/content/docs/'
			},
			social: [
				{ icon: 'github', label: 'GitHub', href: 'https://github.com/louiss0/javascript-package-delegator' }
			]
		}),
	],
});
