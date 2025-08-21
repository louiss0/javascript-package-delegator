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
			// SEO and social sharing configuration
			logo: {
				src: './src/assets/jpd-logo.svg',
				replacesTitle: false,
			},
			favicon: '/favicon.svg',
			head: [
				// Open Graph / Facebook
				{
					tag: 'meta',
					attrs: {
						property: 'og:type',
						content: 'website',
					},
				},
				{
					tag: 'meta',
					attrs: {
						property: 'og:site_name',
						content: 'JavaScript Package Delegator (JPD)',
					},
				},
				{
					tag: 'meta',
					attrs: {
						property: 'og:image',
						content: 'https://jpd-docs.onrender.com/og-image.svg',
					},
				},
				{
					tag: 'meta',
					attrs: {
						property: 'og:image:width',
						content: '1200',
					},
				},
				{
					tag: 'meta',
					attrs: {
						property: 'og:image:height',
						content: '630',
					},
				},
				// Twitter Card
				{
					tag: 'meta',
					attrs: {
						name: 'twitter:card',
						content: 'summary_large_image',
					},
				},
				{
					tag: 'meta',
					attrs: {
						name: 'twitter:site',
						content: '@louiss0',
					},
				},
				{
					tag: 'meta',
					attrs: {
						name: 'twitter:creator',
						content: '@louiss0',
					},
				},
				{
					tag: 'meta',
					attrs: {
						name: 'twitter:image',
						content: 'https://jpd-docs.onrender.com/og-image.svg',
					},
				},
				// Additional meta tags
				{
					tag: 'meta',
					attrs: {
						name: 'keywords',
						content: 'javascript, package manager, cli, go, npm, yarn, pnpm, bun, deno, universal, delegator',
					},
				},
				{
					tag: 'meta',
					attrs: {
						name: 'author',
						content: 'Shelton Louis',
					},
				},
				{
					tag: 'link',
					attrs: {
						rel: 'canonical',
						href: 'https://jpd-docs.onrender.com',
					},
				},
			],
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
