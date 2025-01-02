import { useEffect, useState } from 'react';
import { MapContainer, TileLayer, Marker, useMap, Circle, Rectangle } from 'react-leaflet';
import { useDispatch, useSelector } from 'react-redux';
import L from 'leaflet';
import 'leaflet/dist/leaflet.css';
import 'leaflet-routing-machine';
import { fetchZonesInBounds } from '../services/turfApi';
import { setZonesLoading, setZonesSuccess, setZonesError } from '../store/zonesSlice';
import type { RootState } from '../store/store';
import type { Zone } from '../lib/types/zone';

interface LeafletMapComponentProps {
	checkpoints: string[];
}

function MapController({ checkpoints, setBounds }: { checkpoints: string[], setBounds: (bounds: L.LatLngBounds) => void }) {
	const dispatch = useDispatch();
	console.log('checkpoints', checkpoints);

	const map = useMap();
	const [lastBounds, setLastBounds] = useState<L.LatLngBounds | null>(null);

	useEffect(() => {
		const handleMoveEnd = async () => {
			const bounds = map.getBounds();

			try {
				dispatch(setZonesLoading());
				const zones: [Zone] = await fetchZonesInBounds({
					northEast: {
						lat: bounds.getNorthEast().lat,
						lng: bounds.getNorthEast().lng,
					},
					southWest: {
						lat: bounds.getSouthWest().lat,
						lng: bounds.getSouthWest().lng,
					},
				});
				dispatch(setZonesSuccess(zones));
				setLastBounds(bounds);
				setBounds(bounds); // Set the bounds to be used in the LeafletMapComponent
			} catch (error) {
				dispatch(
					setZonesError(error instanceof Error ? error.message : 'Failed to fetch zones'),
				);
			}
		};

		map.on('moveend', handleMoveEnd);
		handleMoveEnd(); // Initial fetch

		return () => {
			map.off('moveend', handleMoveEnd);
		};
	}, [map, dispatch, lastBounds, setBounds]);

	return null;
}

function ZoneMarkers() {
	const zones = useSelector((state: RootState) => state.zones.items);

	return (
		<>
			{Object.values(zones).map((zone: Zone) => (
				<Circle
					key={zone.id}
					center={[zone.latitude, zone.longitude]}
					radius={50}
					pathOptions={{
						color: '#7ec992',
						fillColor: '#0f2419',
						fillOpacity: 0.6,
					}}
				></Circle>
			))}
		</>
	);
}

export default function LeafletMapComponent({ checkpoints }: LeafletMapComponentProps) {
	const [bounds, setBounds] = useState<L.LatLngBounds | null>(null);

	useEffect(() => {
		delete L.Icon.Default.prototype.options.iconUrl;
		L.Icon.Default.mergeOptions({
			iconRetinaUrl: '/leaflet/marker-icon-2x.png',
			iconUrl: '/leaflet/marker-icon.png',
			shadowUrl: '/leaflet/marker-shadow.png',
		});
	}, []);

	const center: [number, number] = [0, 0];
	const zoom = 2;

	const generateSquares = (bounds: L.LatLngBounds | null) => {
		if (!bounds) return [];
		const squares = [];
		const SIZE_LAT = 0.05;
		const SIZE_LNG = 0.1;
		const northEast = bounds.getNorthEast();
		const southWest = bounds.getSouthWest();

		const latDiff = northEast.lat - southWest.lat;
		const lngDiff = northEast.lng - southWest.lng;

		const numLatSquares = Math.ceil(latDiff / SIZE_LAT) + 1;
		const numLngSquares = Math.ceil(lngDiff / SIZE_LNG) + 1;

		console.log({ numLatSquares, numLngSquares });

		const startLat = Math.floor(southWest.lat / SIZE_LAT) * SIZE_LAT;
		const startLng = Math.floor(southWest.lng / SIZE_LNG) * SIZE_LNG;

		for (let i = 0; i < numLatSquares; i++) {
			for (let j = 0; j < numLngSquares; j++) {
				const lat = startLat + i * SIZE_LAT;
				const lng = startLng + j * SIZE_LNG;
				const squareBounds = L.latLngBounds(
					[L.latLng(lat, lng), L.latLng(lat + SIZE_LAT, lng + SIZE_LNG)]
				);
				squares.push(squareBounds);
			}
		}
		return squares;
	};

	const squares = generateSquares(bounds);

	return (
		<MapContainer center={center} zoom={zoom} style={{ height: '100%', width: '100%' }}>
			<TileLayer
				url='https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png'
				attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors &copy; <a href="https://carto.com/attributions">CARTO</a>'
				maxZoom={19}
			/>
			<MapController checkpoints={checkpoints} setBounds={setBounds} />
			<ZoneMarkers />
			{squares.map((square, index) => (
				<Rectangle
					key={index}
					bounds={square}
					pathOptions={{ color: 'yellow', weight: 2 }}
				/>
			))}
			{checkpoints.map((checkpoint, index) => (
				<Marker
					key={index}
					position={[0, 0]}
					icon={L.divIcon({
						className: 'custom-icon',
						html: `
			  <div style="
				background-color: #7ec992; 
				color: #0f2419; 
				width: 24px; 
				height: 24px; 
				border-radius: 50%; 
				display: flex; 
				justify-content: center; 
				align-items: center; 
				font-weight: bold;
				border: 2px solid #0f2419;
			  ">
				${index + 1}
			  </div>
			`,
					})}
				/>
			))}
		</MapContainer>
	);
}
