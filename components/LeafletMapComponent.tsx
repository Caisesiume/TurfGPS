import React, { useEffect, useState, useRef } from 'react';
import { MapContainer, TileLayer, CircleMarker, useMap, Popup } from 'react-leaflet';
import 'leaflet/dist/leaflet.css';
import L from 'leaflet';
import debounce from 'lodash/debounce';
import { computeRadius, getMoveDistanceThreshold } from '@/lib/mapUtils';

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

	// Use refs to store last fetched center and zoom to avoid unnecessary re-renders.
	const lastFetchedZoomRef = useRef<number>(11);
	const lastFetchedCenterRef = useRef<L.LatLng | null>(null);

	// Updated fetchZones using async/await for clarity and error handling.
	const fetchZones = async (bounds: L.LatLngBounds, zoomLevel: number) => {
		const minLat = bounds.getSouth();
		const maxLat = bounds.getNorth();
		const minLng = bounds.getWest();
		const maxLng = bounds.getEast();
		try {
			// Pass current zoom to the API so that clustering is performed at that level.
			const response = await fetch(
				`/api/zones?minLat=${minLat}&maxLat=${maxLat}&minLng=${minLng}&maxLng=${maxLng}&zoom=${zoomLevel}`,
			);
			const data: ClusterFeature[] = await response.json();
			// Compare lengths (consider a deeper comparison if necessary)
			if (data.length !== zones.length) {
				setZones(data);
			}
		} catch (error) {
			console.error('Error fetching zones:', error);
		}
	};

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
			debouncedHandleMoveEnd();

			return () => {
				debouncedHandleMoveEnd.cancel();
				map.off('moveend', debouncedHandleMoveEnd);
				map.off('zoomend', debouncedHandleMoveEnd);
			};
		}, [map]);

		return null;
	};

	const center = new L.LatLng(60.115, 16.187);
	const radius = computeRadius(currentZoom);
	console.log({ zoom: currentZoom });
	
	return (
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
					}
				})}
		</MapContainer>
	);
};

export default LeafletMapComponent;
