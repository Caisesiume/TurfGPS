import { useState } from 'react';
import { useTranslations } from 'next-intl';

interface CheckpointFormProps {
	onSubmit: (checkpoints: string[]) => void;
}

export default function CheckpointForm({ onSubmit }: CheckpointFormProps) {
	const t = useTranslations('form');
	const [checkpoints, setCheckpoints] = useState<string[]>(['']);

	const handleAddCheckpoint = () => {
		setCheckpoints([...checkpoints, '']);
	};

	const handleRemoveCheckpoint = (index: number) => {
		const newCheckpoints = checkpoints.filter((_, i) => i !== index);
		setCheckpoints(newCheckpoints);
	};

	const handleCheckpointChange = (index: number, value: string) => {
		const newCheckpoints = [...checkpoints];
		newCheckpoints[index] = value;
		setCheckpoints(newCheckpoints);
	};

	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();
		onSubmit(checkpoints.filter((cp) => cp.trim() !== ''));
	};

	return (
		<form onSubmit={handleSubmit} className='mb-4'>
			{checkpoints.map((checkpoint, index) => (
				<div key={index} className='flex mb-2'>
					<input
						type='text'
						value={checkpoint}
						onChange={(e) => handleCheckpointChange(index, e.target.value)}
						placeholder={`Checkpoint ${index + 1}`}
						className='flex-grow bg-white text-gray-900 p-2 rounded mr-2 border border-gray-300'
					/>
					<button
						type='button'
						onClick={() => handleRemoveCheckpoint(index)}
						className='bg-red-600 text-white p-2 rounded'
					>
						{t('removeCheckpoint')}
					</button>
				</div>
			))}
			<button
				type='button'
				onClick={handleAddCheckpoint}
				className='bg-green-600 text-white p-2 rounded mr-2'
			>
				{t('addCheckpoint')}
			</button>
			<button type='submit' className='bg-blue-600 text-white p-2 rounded'>
				{t('planRoute')}
			</button>
		</form>
	);
}
