import type { Coordinates } from './types/coordinates';
import type { BoundBox } from './types/bound';

export type OperationData = {
	success: boolean;
	message?: string;
	sizing?: {
		latDiff: number;
		lngDiff: number;
	};
};

/**
 * Break down the bounds into grid boxes of the given size
 * @param northEast - The north east corner of the bounds
 * @param southWest - The south west corner of the bounds
 * @param boxSize - The size of the grid boxes
 * @returns The grid boxes
 */
export function gridBoxesCalculator(
	northEast: Coordinates,
	southWest: Coordinates,
	boxSize: { lat: number; lng: number } = { lat: 0.025, lng: 0.05 },
): [BoundBox[], OperationData] {
	const operationData = {
		success: true,
		message: '',
		sizing: {
			latDiff: 0,
			lngDiff: 0,
		},
	};
	// If the bound is larger than 1.0 in either direction, return an empty array
	if (
		Math.abs(northEast.latitude - southWest.latitude) > 0.9 ||
		Math.abs(northEast.longitude - southWest.longitude) > 0.85
	) {
		operationData.success = false;
		operationData.message = 'Bound is too large';
		operationData.sizing = {
			latDiff: Math.abs(northEast.latitude - southWest.latitude),
			lngDiff: Math.abs(northEast.longitude - southWest.longitude),
		};
		return [[], operationData];
	}

	const bounds: BoundBox = {
		northEast,
		southWest,
	};

	if (!bounds) {
		console.log({ boxSize });
		
		operationData.success = false;
		operationData.message = 'Bounds are not provided';
		return [[], operationData];
	};

	const squares = [] as BoundBox[];
	const sizeLat = 0.05;
	const sizeLng = 0.1;
	
	const NElat = bounds.northEast.latitude;
	const NElng = bounds.northEast.longitude;

	const SWlat = bounds.southWest.latitude;
	const SWlng = bounds.southWest.longitude;

	const latDiff = NElat - SWlat;
	const lngDiff = NElng - SWlng;

	const numLatSquares = Math.ceil(latDiff / sizeLat) + 1;
	const numLngSquares = Math.ceil(lngDiff / sizeLng) + 1;

	console.log({ numLatSquares, numLngSquares });

	const startLat = Math.floor(SWlat / sizeLat) * sizeLat;
	const startLng = Math.floor(SWlng / sizeLng) * sizeLng;

	for (let i = 0; i < numLatSquares; i++) {
		for (let j = 0; j < numLngSquares; j++) {
			const lat = startLat + i * sizeLat;
			const lng = startLng + j * sizeLng;
			const square: BoundBox = {
				northEast: {
					latitude: lat + sizeLat,
					longitude: lng + sizeLng,
				},
				southWest: {
					latitude: lat,
					longitude: lng,
				},
			};
			//console.log({ square });


			squares.push(square);
		}
	}

	return [squares, operationData];
}
