import dbmodels from './dbModels'; // Adjust the import path as necessary
import { BoundBox } from './types'; // Adjust the import path as necessary
import { IZone } from './types';

const getZonesWithinBound = async (bound: BoundBox): Promise<IZone[]> => {
	const { zoneModel } = dbmodels;
	return await zoneModel
		.find({
			'location.coordinates': {
				$geoWithin: {
					$box: [
						[bound.southWest.longitude, bound.southWest.latitude],
						[bound.northEast.longitude, bound.northEast.latitude],
					],
				},
			},
		})
		.select({
			zoneId: 1,
			name: 1,
			location: 1,
		})
		.lean();
};

export default getZonesWithinBound;
