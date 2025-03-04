import React, { useEffect, useState } from 'react';
import { MapContainer, TileLayer, Polygon, useMap, CircleMarker } from 'react-leaflet';
import 'leaflet/dist/leaflet.css';
import L from 'leaflet';
import { IZone } from '@/lib/types';

const LeafletMapComponent: React.FC = () => {
	const [zones, setZones] = useState<IZone[]>([]);

	const fetchZones = (bounds: L.LatLngBounds) => {
		const minLat = bounds.getSouth();
		const maxLat = bounds.getNorth();
		const minLng = bounds.getWest();
		const maxLng = bounds.getEast();

		fetch(`/api/zones?minLat=${minLat}&maxLat=${maxLat}&minLng=${minLng}&maxLng=${maxLng}`)
			.then(response => response.json())
			.then(data => {
				if (JSON.stringify(data) !== JSON.stringify(zones)) {
					setZones(data as IZone[]);
				}
			})
			.catch(error => console.error('Error fetching zones:', error));
	};

	const MapEventHandler: React.FC = () => {
		const map = useMap();

		useEffect(() => {
			const handleMoveEnd = () => {
				const bounds = map.getBounds();
				fetchZones(bounds);
			};

			map.on('moveend', handleMoveEnd);
			handleMoveEnd(); // Fetch zones on initial load

			return () => {
				map.off('moveend', handleMoveEnd);
			};
		}, [map]);

		return null;
	};
	const center = new L.LatLng(60.115, 16.187);
	
	return (
		<MapContainer center={center} zoom={11} style={{ height: '100vh', width: '100%' }}>
			<TileLayer
				key={Math.random()}
				url='https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png'
				attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
			/>
			<MapEventHandler />
			{zones.length > 0 && zones.map(zone => {
				const coordinates = zone.location.coordinates;
				const position = new L.LatLng(coordinates[0], coordinates[1]);
				console.log('Rendering zone:', zone.zoneId);
				
				return (
					<CircleMarker key={zone.zoneId} center={position} radius={100} color='red'>
						<p>{zone.zoneId}</p>
					</CircleMarker>
				);
			})}
		</MapContainer>
	);
};

export default LeafletMapComponent;