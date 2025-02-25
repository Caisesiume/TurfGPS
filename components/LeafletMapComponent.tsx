/* eslint-disable @typescript-eslint/ban-ts-comment */
/* eslint-disable @typescript-eslint/no-unused-vars */
import React, { useEffect, useState, useCallback } from 'react';
import { MapContainer, TileLayer, CircleMarker, Popup, useMap } from 'react-leaflet';
import L from 'leaflet';
import 'leaflet/dist/leaflet.css';
import { BoundBox, Coordinates, IZone, LocationCoordinates } from '@/lib/types';
import { debounce } from 'lodash';

interface Cluster {
	location: LocationCoordinates;
	id: number;
	coordinates: [number, number];
	cluster: boolean;
}

const TILE_SIZE = 256;

const getTileKey = (zoom: number, x: number, y: number) => `${zoom}/${x}/${y}`;

const calculateTileBounds = (bounds: L.LatLngBounds, zoom: number) => {
	const tileXStart = Math.floor(((bounds.getWest() + 180) / 360) * (1 << zoom));
	const tileXEnd = Math.floor(((bounds.getEast() + 180) / 360) * (1 << zoom));
	const tileYStart = Math.floor(
		((1 -
			Math.log(
				Math.tan((bounds.getNorth() * Math.PI) / 180) +
					1 / Math.cos((bounds.getNorth() * Math.PI) / 180),
			) /
				Math.PI) /
			2) *
			(1 << zoom),
	);
	const tileYEnd = Math.floor(
		((1 -
			Math.log(
				Math.tan((bounds.getSouth() * Math.PI) / 180) +
					1 / Math.cos((bounds.getSouth() * Math.PI) / 180),
			) /
				Math.PI) /
			2) *
			(1 << zoom),
	);
	//console.log({ tileXStart, tileXEnd, tileYStart, tileYEnd });

	return { tileXStart, tileXEnd, tileYStart, tileYEnd };
};

const LeafletMapComponent: React.FC = () => {
	const [clusters, setClusters] = useState<IZone[]>([]);
	const [tileCache, setTileCache] = useState<Record<string, Cluster[]>>({});
	const [center, setCenter] = useState<[number, number]>([60.156, 16.207]);
	const [zoom, setZoom] = useState<number>(11);
	const [loading, setLoading] = useState<boolean>(true);
	const [fetchQueue, setFetchQueue] = useState<(() => Promise<void>)[]>([]);
	const [isFetching, setIsFetching] = useState<boolean>(false);
	const [hasFetched, setHasFetched] = useState<boolean>(false);

	useEffect(() => {
		navigator.geolocation.getCurrentPosition(
			(position) => {
				setCenter([position.coords.latitude, position.coords.longitude]);
				setZoom(11);
				setLoading(false);
			},
			() => {
				setCenter([60.156, 16.207]); // Avesta, Dalarna, Sweden
				setZoom(11);
				setLoading(false);
			},
		);
	}, []);

	const fetchTileData = async (zoom: number, x: number, y: number) => {
		const bounds = getTileBounds(x, y, zoom);
		return bounds;
	};

	const updateClusters = async (bounds: L.LatLngBounds, zoom: number) => {
		const { tileXStart, tileXEnd, tileYStart, tileYEnd } = calculateTileBounds(bounds, zoom);
		setClusters(await fetchAllZonesInTile(zoom, tileXStart, tileXEnd, tileYStart, tileYEnd));
	};

	async function fetchAllZonesInTile(
		zoom: number,
		tileXStart: number,
		tileXEnd: number,
		tileYStart: number,
		tileYEnd: number,
	) {
		const bounds = [];
		for (let x = tileXStart; x <= tileXEnd; x++) {
			for (let y = tileYStart; y <= tileYEnd; y++) {
				bounds.push(getTileBounds(x, y, zoom));
			}
		}
	
		const flatTiles = bounds.flat();
		console.log({ flatTiles });
	
		// Determine the northEast and southWest-most bounds
		const northEast = {
			lat: Math.max(...flatTiles.map((bound) => bound.northEast.latitude)),
			lng: Math.max(...flatTiles.map((bound) => bound.northEast.longitude)),
		};
		const southWest = {
			lat: Math.min(...flatTiles.map((bound) => bound.southWest.latitude)),
			lng: Math.min(...flatTiles.map((bound) => bound.southWest.longitude)),
		};
	
		// Fetch the entire tile from the API using the calculated bounds
		const response = await fetch(
			`/api/zones?minLng=${southWest.lng}
			&maxLng=${northEast.lng}
			&minLat=${southWest.lat}
			&maxLat=${northEast.lat}`.replace(/\s+/g, ''),
		);
		const data = await response.json();
		console.log({ data });
		return data;
	}

	function getTileBounds(x: number, y: number, z: number): BoundBox {
		const n = Math.pow(2, z);
		const lonDeg = (x / n) * 360.0 - 180.0;
		const latRad = Math.atan(Math.sinh(Math.PI * (1 - (2 * y) / n)));
		const latDeg = (latRad * 180.0) / Math.PI;

		const tileNorthEast: Coordinates = {
			latitude: latDeg + 360.0 / n,
			longitude: lonDeg + 360.0 / n,
		};
		const tileSouthWest: Coordinates = {
			latitude: latDeg,
			longitude: lonDeg,
		};
		return {
			northEast: tileNorthEast,
			southWest: tileSouthWest,
		};
	}

	const enqueueFetch = useCallback((fetchFn: () => Promise<void>) => {
		setFetchQueue((prev) => [...prev, fetchFn]);
	}, []);

	const MapController = () => {
		const map = useMap();
		// @ts-ignore
		useEffect(() => {
			const onMoveEnd = debounce(() => {
				if (!hasFetched) {
					const bounds = map.getBounds();
					const zoom = map.getZoom();
					console.log({ bounds });
			
					updateClusters(bounds, zoom);
					setHasFetched(true);
				}
			}, 3000);

			map.on('dragend', onMoveEnd);
			onMoveEnd(); // Trigger on initial load

			return () => map.off('dragend', onMoveEnd);
		}, [map]);

		return null;
	};

	if (loading) {
		return <div>Loading...</div>;
	}

	return (
		<MapContainer center={center} zoom={zoom} style={{ height: '100%', width: '100%' }}>
			<TileLayer
				url='https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png'
				attribution='&copy; OpenStreetMap contributors'
				tileSize={TILE_SIZE}
			/>
			<MapController />
			{clusters.map((zone) =>
				(
					<CircleMarker
						key={`cluster-${zone.zoneId}`}
						center={[zone.location.coordinates[1], zone.location.coordinates[0]]}
						radius={2}
						pathOptions={{ color: '#ff0000' }}
					>
						<Popup>
							<div>{zone.name} zone</div>
						</Popup>
					</CircleMarker>
				)
			)}
		</MapContainer>
	);
};

export default LeafletMapComponent;
