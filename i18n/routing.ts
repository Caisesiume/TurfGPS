import { defineRouting } from 'next-intl/routing';
import { createNavigation } from 'next-intl/navigation';

/* REMINDER:

 1. Update language here
 2. Update pathnames here
 3. Update the LanguageSwitcher component
 4. Update the middleware.ts regex matcher

*/

export const routing = defineRouting({
	locales: ['en', 'sv', 'de'],
	defaultLocale: 'en',
	localePrefix: {
		mode: 'always',
		prefixes: {
			en: '/en',
			sv: '/sv',
			de: '/de',
		},
	},
	pathnames: {
		'/': '/',
		'/organization': {
			en: '/organisation',
			sv: '/organisation',
			de: '/organisation',
		},
	},
});

export const { Link, redirect, usePathname, useRouter, getPathname } = createNavigation(routing);
