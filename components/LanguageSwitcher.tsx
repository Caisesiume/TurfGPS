'use client';

import { useLocale } from 'next-intl';
import { useRouter } from 'next/navigation';
import { Languages } from 'lucide-react';

export default function LanguageSwitcher() {
	const locale = useLocale();
	const router = useRouter();

	const handleChange = (newLocale: string) => {
		const currentPath = window.location.pathname;
		console.log({ currentPath });
		const newPath = currentPath.replace(`/${locale}`, `/${newLocale}`);

		router.push(newPath);
	};

	return (
		<div className='flex items-center space-x-2'>
			<div className='relative'>
				<Languages className='absolute left-2 top-1/2 transform -translate-y-1/2 text-green-600 pointer-events-none' />
				<select
					value={locale}
					onChange={(e) => handleChange(e.target.value)}
					className='appearance-none bg-green-700 text-white p-2 pl-8 rounded border border-green-500 focus:outline-none focus:none'
				>
					<option value='en'>English</option>
					<option value='sv'>Svenska</option>
					<option value='de'>Deutsch</option>
				</select>
			</div>
		</div>
	);
}
