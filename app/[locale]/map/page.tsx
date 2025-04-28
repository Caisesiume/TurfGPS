'use client';

import { useMemo } from 'react';
import dynamic from 'next/dynamic';
import { Provider } from 'react-redux';
import { useTranslations } from 'next-intl';
import CheckpointForm from '@/components/CheckpointForm';
import LanguageSwitcher from '@/components/LanguageSwitcher';
import { store } from '@/store/store';

const LeafletMapComponent = dynamic(() => import('@/components/LeafletMapComponent'), {
	ssr: false,
});

export default function RoutePlanner() {
	const t = useTranslations();

	const LeafletMap = useMemo(() => <LeafletMapComponent />, []);
	console.log(" RENDERING ");

	return (
		<Provider store={store}>
			<div className='min-h-screen bg-green-900 text-gray-900'>
				<div className='container mx-auto p-4'>
					<div className='flex justify-between items-center pt-8 pb-16 mb-4'>
						<h1 className='text-4xl font-bold text-green-600'>{t('title')}</h1>
						<LanguageSwitcher />
					</div>
					<CheckpointForm onSubmit={(checkpoints) => console.log(checkpoints)} />
					<div className='h-[600px] w-full rounded-lg overflow-hidden shadow-lg'>
						{LeafletMap}
					</div>
				</div>
			</div>
		</Provider>
	);
}