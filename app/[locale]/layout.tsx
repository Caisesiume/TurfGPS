import type { Metadata } from 'next';
import { notFound } from 'next/navigation';
import { NextIntlClientProvider } from 'next-intl';
import { getMessages } from 'next-intl/server';
import { routing } from '@/i18n/routing';
import localFont from 'next/font/local';
import '../globals.css';

const geistSans = localFont({
	src: '../fonts/GeistVF.woff',
	variable: '--font-geist-sans',
	weight: '100 900',
});
const geistMono = localFont({
	src: '../fonts/GeistMonoVF.woff',
	variable: '--font-geist-mono',
	weight: '100 900',
});

export const metadata: Metadata = {
	title: 'Turf GPS',
	description: 'Plan your Turf routes in a few clicks',
	keywords: ['Turf', 'Turfgame', 'GPS', 'Route', 'Planner', 'Map', 'Zone', 'Zones'],
	authors: [{ name: 'Caisesiume', url: 'https://github.com/Caisesiume' }],
	creator: 'Caisesiume',
};

export function generateStaticParams() {
	return [{ locale: 'en' }, { locale: 'sv' }];
}

export default async function RootLayout({
	children,
	params,
}: Readonly<{
	children: React.ReactNode;
	params: { locale: string };
}>) {
	const { locale } = await params;

	if (!routing.locales.includes(locale as any)) {
		notFound();
	}

	// Providing all messages to the client
	// side is the easiest way to get started
	const messages = await getMessages();

	return (
		<html lang={locale}>
			<body className={`${geistSans.variable} ${geistMono.variable} antialiased`}>
				<NextIntlClientProvider messages={messages} locale={locale}>
					{children}
				</NextIntlClientProvider>
			</body>
		</html>
	);
}
