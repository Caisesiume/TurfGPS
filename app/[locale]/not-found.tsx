'use client';

import { useTranslations, useLocale } from 'next-intl';
import Link from 'next/link';

export default function CatchAllPage() {
	const t = useTranslations();
	const locale = useLocale();

	return (
		<main className='bg-green-600 dark:bg-green-700 grid min-h-screen place-items-center gap-y-6 px-6 py-24 sm:py-32 lg:px-8'>
			<h1 className='text-5xl font-semibold text-gray-900 sm:text-7xl'>
				{t('notFound.title')}
			</h1>
			<p className='text-6xl font-semibold text-green-800 dark:text-green-600'>404</p>
			<p className='text-lg font-medium text-gray-900 sm:text-xl/8'>
				{t('notFound.description')}
			</p>
			<div className='flex flex-col gap-4 items-center justify-center gap-8'>
				<Link href={`/${locale}`}>
					<button className='rounded-full bg-green-800 dark:bg-green-600 px-5 py-3 text-base font-semibold text-white dark:text-gray-900 shadow-xs hover:bg-green-700 dark:hover:bg-green-500 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-green-600'>
						{t('notFound.goHome')}
					</button>
				</Link>
				<Link href={`/${locale}/map`}>
					<button className='rounded-full bg-green-800 dark:bg-green-600 px-5 py-3 text-base font-semibold text-white dark:text-gray-900 shadow-xs hover:bg-green-700 dark:hover:bg-green-500 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-green-600'>
						{t('notFound.toMap')} <span aria-hidden='true'>&rarr;</span>
					</button>
				</Link>
			</div>
		</main>
	);
}

// export default function Example() {
//     const t =
// 	return (
// 		<>
// 			{/*
//           This example requires updating your template:

//           ```
//           <html class="h-full">
//           <body class="h-full">
//           ```
//         */}
// 			<main className='grid min-h-full place-items-center bg-green-600 px-6 py-24 sm:py-32 lg:px-8'>
// 				<div className='text-center'>
// 					<p className='text-4xl font-semibold text-text-green-700'>404</p>
// 					<h1 className='mt-4 text-5xl font-semibold tracking-tight text-balance text-gray-900 sm:text-7xl'>
// 						Page not found
// 					</h1>
// 					<p className='mt-6 text-lg font-medium text-pretty text-gray-500 sm:text-xl/8'>
// 						Sorry, we couldn’t find the page you’re looking for.
// 					</p>
// 					<div className='mt-10 flex items-center justify-center gap-x-6'>
// 						<a
// 							href='#'
// 							className='rounded-md bg-indigo-600 px-3.5 py-2.5 text-sm font-semibold text-white shadow-xs hover:bg-indigo-500 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600'
// 						>
// 							Go back home
// 						</a>
// 						<a href='#' className='text-sm font-semibold text-gray-900'>
// 							Contact support <span aria-hidden='true'>&rarr;</span>
// 						</a>
// 					</div>
// 				</div>
// 			</main>
// 		</>
// 	);
// }
