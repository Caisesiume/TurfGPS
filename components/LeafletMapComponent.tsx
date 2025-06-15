import React, { useEffect, useState, useRef, useCallback } from 'react';
import { MapContainer, TileLayer, CircleMarker, useMap, Popup } from 'react-leaflet';
import 'leaflet/dist/leaflet.css';
import L from 'leaflet';
import debounce from 'lodash/debounce';
import { computeRadius, getMoveDistanceThreshold } from '@/lib/mapUtils';
import { useRequestManager } from '@/lib/requestUtils';
import MapErrorBoundary from './MapErrorBoundary';

type ClusterFeature = {
	type: 'Feature';
	properties: {
		cluster?: boolean;
		cluster_id?: number;
		point_count?: number;
		zoneId?: number;
	};
	geometry: {
		type: 'Point';
		coordinates: [number, number]; // [lng, lat]
	};
};

const LeafletMapComponent: React.FC = () => {
	const [zones, setZones] = useState<ClusterFeature[]>([]);
	const [currentZoom, setCurrentZoom] = useState<number>(11);
	const [isLoading, setIsLoading] = useState<boolean>(false);
	const [error, setError] = useState<string | null>(null);

	// Use refs to store last fetched center and zoom to avoid unnecessary re-renders.
	const lastFetchedZoomRef = useRef<number>(11);
	const lastFetchedCenterRef = useRef<L.LatLng | null>(null);
	const { createRequest } = useRequestManager();

	// Helper function to compare zones deeply
	const compareZones = useCallback((newZones: ClusterFeature[], oldZones: ClusterFeature[]) => {
		if (newZones.length !== oldZones.length) return false;
		
		// Deep comparison for zones - could be optimized based on specific needs
		return JSON.stringify(newZones) === JSON.stringify(oldZones);
	}, []);	// Updated fetchZones using async/await for clarity and error handling.
	const fetchZones = useCallback(async (bounds: L.LatLngBounds, zoomLevel: number) => {
		const minLat = bounds.getSouth();
		const maxLat = bounds.getNorth();
		const minLng = bounds.getWest();
		const maxLng = bounds.getEast();
		
		// Generate unique request ID to prevent race conditions
		const request = createRequest();
		
		// Only update if this is still the latest request
		if (request.isLatest()) {
			setIsLoading(true);
			setError(null);
		}
		
		try {
			// Pass current zoom to the API so that clustering is performed at that level.
			const response = await fetch(
				`/api/zones?minLat=${minLat}&maxLat=${maxLat}&minLng=${minLng}&maxLng=${maxLng}&zoom=${zoomLevel}`,
			);
			
			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}
			
			const data: ClusterFeature[] = await response.json();
			
			// Only update if this is still the latest request
			if (request.isLatest()) {
				// Better comparison using deep comparison
				if (!compareZones(data, zones)) {
					setZones(data);
				}
				setIsLoading(false);
			}
		} catch (error) {
			console.error('Error fetching zones:', error);
			// Only update error state if this was the latest request
			if (request.isLatest()) {
				setError(error instanceof Error ? error.message : 'Failed to fetch zones');
				setIsLoading(false);
			}
		}
	}, [zones, compareZones, createRequest]);

	// Component to handle map events with restricted fetch conditions.
	const MapEventHandler: React.FC = () => {
		const map = useMap();

		useEffect(() => {
			// On initial load, set the last fetched center and fetch zones.
			if (!lastFetchedCenterRef.current) {
				const initialCenter = map.getCenter();
				lastFetchedCenterRef.current = initialCenter;
				const currentZoomLevel = map.getZoom();
				setCurrentZoom(currentZoomLevel);
				fetchZones(map.getBounds(), currentZoomLevel);
			}

			// Define the handler for moveend/zoomend events.
			const handleMoveEnd = () => {
				const currentZoomLevel = map.getZoom();
				setCurrentZoom(currentZoomLevel);

				const currentCenter = map.getCenter();
				const movedDistance = lastFetchedCenterRef.current
					? currentCenter.distanceTo(lastFetchedCenterRef.current)
					: 0;

				const moveDistanceThreshold = getMoveDistanceThreshold(currentZoomLevel);

				if (
					(currentZoomLevel > 8 && currentZoomLevel !== lastFetchedZoomRef.current) ||
					movedDistance > moveDistanceThreshold
				) {
					lastFetchedZoomRef.current = currentZoomLevel;
					lastFetchedCenterRef.current = currentCenter;
					fetchZones(map.getBounds(), currentZoomLevel);
				}
			};

			// Debounce the move/zoom event handler to avoid excessive API calls.
			const debouncedHandleMoveEnd = debounce(handleMoveEnd, 300);

			map.on('moveend', debouncedHandleMoveEnd);
			map.on('zoomend', debouncedHandleMoveEnd);

			// Trigger initial update.
			debouncedHandleMoveEnd();			return () => {
				debouncedHandleMoveEnd.cancel();
				map.off('moveend', debouncedHandleMoveEnd);
				map.off('zoomend', debouncedHandleMoveEnd);
			};		}, [map, fetchZones]); // eslint-disable-line react-hooks/exhaustive-deps

		return null;
	};	const center = new L.LatLng(60.115, 16.187);
	const radius = computeRadius(currentZoom);
	console.log({ zoom: currentZoom });
	
	return (
		<MapErrorBoundary>
			<div className='relative'>
				{isLoading && (
					<div className='absolute top-4 left-4 z-[1000] bg-white px-3 py-2 rounded shadow-lg'>
						<div className='flex items-center space-x-2'>
							<div className='animate-spin rounded-full h-4 w-4 border-b-2 border-blue-500'></div>
							<span className='text-sm'>Loading zones...</span>
						</div>
					</div>
				)}
				{error && (
					<div className='absolute top-4 left-4 z-[1000] bg-red-100 border border-red-400 text-red-700 px-3 py-2 rounded'>
						<div className='flex items-center space-x-2'>
							<span className='text-sm'>⚠️ {error}</span>
							<button 
								onClick={() => setError(null)}
								className='ml-2 text-red-500 hover:text-red-700'
							>
								×
							</button>
						</div>
					</div>
				)}
				<MapContainer
					center={center}
					zoom={11}
					minZoom={8}
					style={{ height: '100vh', width: '100%' }}
				>
			<TileLayer
				url='https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png'
				attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
			/>
			<MapEventHandler />
			{zones.length > 0 &&
				zones.map((zone) => {
					if (!zone.geometry || !zone.geometry.coordinates) {
						console.error('Invalid zone:', zone);
						return null;
					}
					// Coordinates are in [lng, lat] order.
					const [lng, lat] = zone.geometry.coordinates;
					const position = new L.LatLng(lat, lng);

					// Render cluster markers differently from individual zones.
					if (zone.properties.cluster) {
						const clusterCount = zone.properties.point_count || 0;
						// Adjust cluster radius depending on the number of points.
						const clusterRadius = radius + Math.log(clusterCount + 1) * 3;

						return (
							<CircleMarker
								key={`cluster-${zone.properties.cluster_id}`}
								center={position}
								radius={clusterRadius} // Increase size for clusters.
								color='blue'
								fillOpacity={1}
							>
								<Popup>{clusterCount} zones</Popup>
							</CircleMarker>
						);
					} else {
						return (
							<CircleMarker
								key={zone.properties.zoneId}
								center={position}
								radius={radius}
								color='red'
							>
								<Popup>{zone.properties.zoneId}</Popup>
							</CircleMarker>
						);
					}				})}
		</MapContainer>
			</div>
		</MapErrorBoundary>
	);
};

export default LeafletMapComponent;
