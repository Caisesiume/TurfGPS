'use client';

import { useState, useEffect } from 'react';
import dynamic from 'next/dynamic';
import { Provider } from 'react-redux';
import { useTranslations } from 'next-intl';
import CheckpointForm from '../../components/CheckpointForm';
import LanguageSwitcher from '../../components/LanguageSwitcher';
import { store } from '../../store/store';

const LeafletMapComponent = dynamic(() => import('../../components/LeafletMapComponent'), {
	ssr: false,
});

import { fetchZonesInBounds } from '@/services/turfApiService';
import { IZone } from '@/lib/types/domain';

export default function RoutePlanner() {
	const t = useTranslations();

	return (
		<Provider store={store}>
			<div className='min-h-screen bg-green-900 text-gray-900'>
				<div className='container mx-auto p-4'>
					<h1 className='text-4xl font-bold mb-4 text-green-700'>{t('title')}</h1>
					<LanguageSwitcher />
					<CheckpointForm onSubmit={(checkpoints) => console.log(checkpoints)} />
					<div className='h-[600px] w-full rounded-lg overflow-hidden shadow-lg'>
						<LeafletMapComponent />
					</div>
				</div>
			</div>
		</Provider>
	);
}
