/**
 * A helper function to compute radius based on zoom level.
 * Adjust the formula as needed.
 *
 * @param zoom - The current zoom level of the map.
 * @returns The computed radius.
 */
export const computeRadius = (zoom: number): number => {
	// Define a mapping of zoom levels to radii
	const zoomRadiusMap: { [key: number]: number } = {
		1: 1,
		2: 1,
		3: 1,
		4: 1,
		5: 1,
		6: 2,
		7: 2,
		8: 2,
		9: 3,
		10: 3,
		11: 3,
		12: 4,
		13: 5,
		14: 5,
		15: 7,
		16: 15,
		17: 25,
		18: 45,
	};

	// Ensure the zoom level is within the defined range
	const clampedZoom = Math.max(1, Math.min(zoom, 18));

	// Get the radius from the map
	const radius = zoomRadiusMap[clampedZoom];
	return radius;
};

/**
 * A helper function to determine the move distance threshold based on the zoom level.
 * This is intended to be used as a threshold of when to fetch new zones.
 *
 * @param zoom - The current zoom level of the map.
 * @returns The move distance threshold in meters.
 */
export const getMoveDistanceThreshold = (zoom: number): number => {
    // Define a more dynamic function to determine the move distance threshold based on the zoom level
    if (zoom > 15) return 500; // 100 meters for high zoom levels
    if (zoom > 12) return 1500; // 500 meters for medium-high zoom levels
    if (zoom > 10) return 7000; // 1000 meters for medium zoom levels
    if (zoom > 8) return 20000; // 2000 meters for medium-low zoom levels
    return 50000; // 5000 meters for low zoom levels
};
