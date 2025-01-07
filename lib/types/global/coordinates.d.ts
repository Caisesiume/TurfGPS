export interface Coordinates {
	latitude: number;
	longitude: number;
}

export interface LocationCoordinates {
	coordinates: [
		latitude: number,
		longitude: number,
	];
	type: string;
}
