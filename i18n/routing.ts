import { defineRouting } from 'next-intl/routing';
import { createNavigation } from 'next-intl/navigation';

export const routing = defineRouting({
	locales: ['en', 'sv'],
	defaultLocale: 'en',
	localePrefix: {
		mode: 'always',
		prefixes: {
			en: '/en',
			sv: '/se',
		},
	},
	pathnames: {
		'/': '/',
		'/organization': {
			en: '/organisation',
			sv: '/organisation',
		},
	},
});

export const { Link, redirect, usePathname, useRouter, getPathname } = createNavigation(routing);
