/* eslint-disable @typescript-eslint/ban-ts-comment */
/* eslint-disable @typescript-eslint/no-explicit-any */
import { NextRequest, NextResponse } from 'next/server';
import * as models from '@/lib/dbModels';

const cache = new Map<string, { zones: any[]; timestamp: number }>();
const CACHE_EXPIRY_MS = 3600000; // 1 hour

async function fetchZonesFromDatabase(
	minLng: number,
	maxLng: number,
	minLat: number,
	maxLat: number,
) {
	console.log(' ---- Fetching zones from database ---- ');
	const zoneModel = models.default.zoneModel;
	const zones = await zoneModel
		.find({
			'location.coordinates': {
				$geoWithin: {
					$box: [
						[minLng, minLat],
						[maxLng, maxLat],
					],
				},
			},
		})
		.select({
			zoneId: 1,
			location: 1,
			_id: 0,
		})
		.lean();
	return zones;
}

export async function getHandler(req: NextRequest) {
	const { searchParams } = req.nextUrl;
	
	// Get parameters as strings first
	const minLngStr = searchParams.get('minLng');
	const maxLngStr = searchParams.get('maxLng');
	const minLatStr = searchParams.get('minLat');
	const maxLatStr = searchParams.get('maxLat');

	// Check if all required parameters are present
	if (!minLngStr || !maxLngStr || !minLatStr || !maxLatStr) {
		return NextResponse.json({ error: 'Missing required coordinates | gz100' }, { status: 400 });
	}

	// Parse coordinates
	const minLng = parseFloat(minLngStr);
	const maxLng = parseFloat(maxLngStr);
	const minLat = parseFloat(minLatStr);
	const maxLat = parseFloat(maxLatStr);

	// Validate parsed coordinates
	if (isNaN(minLng) || isNaN(maxLng) || isNaN(minLat) || isNaN(maxLat)) {
		return NextResponse.json({ error: 'Invalid coordinate format | gz101' }, { status: 400 });
	}

	// Validate coordinate bounds
	if (minLng < -180 || maxLng > 180 || minLat < -90 || maxLat > 90) {
		return NextResponse.json({ error: 'Coordinates out of bounds | gz102' }, { status: 400 });
	}

	// Validate coordinate ranges
	if (minLng >= maxLng || minLat >= maxLat) {
		return NextResponse.json({ error: 'Invalid coordinate range | gz103' }, { status: 400 });
	}

	// Build cache key based on the bounding box (optionally include zoom if needed)
	const cacheKey = `tile:${maxLat},${maxLng}:${minLat},${minLng}`;

	try {
		// Check in-memory cache
		const cachedData = cache.get(cacheKey);
		if (cachedData && Date.now() - cachedData.timestamp < CACHE_EXPIRY_MS) {
			return NextResponse.json(cachedData.zones);
		}

		// Fetch zones from the database
		const zones = await fetchZonesFromDatabase(minLng, maxLng, minLat, maxLat);
		console.log(' ---- Fetched zones from database ---- ');

		// Update cache with raw zones (optional, if needed elsewhere)
		cache.set(cacheKey, { zones, timestamp: Date.now() });
		console.log(' ---- Updated in-memory cache ---- ');

		// Get zoom level (default to 11 if not provided)
		const zoomParam = searchParams.get('zoom');
		const zoomLevel = zoomParam ? parseInt(zoomParam) : 11;

		// Dynamically import Supercluster
		const Supercluster = (await import('supercluster')).default;

		// Convert zones to GeoJSON features
		const features = zones.map((zone: any) => ({
			type: "Feature" as const,
			properties: { zoneId: zone.zoneId },
			geometry: zone.location, // expecting GeoJSON format (i.e., { type: 'Point', coordinates: [lng, lat] })
		}));

		// Calculate dynamic radius based on zoom level
		const dynamicRadius = Math.max(40, 100 - zoomLevel * 5); // Adjust the formula as needed

		// Initialize Supercluster with desired options
		const superclusterInstance = new Supercluster({
			radius: dynamicRadius, // Cluster radius in pixels (adjust as needed)
			maxZoom: 10, // Maximum zoom level at which to create clusters
		});
		superclusterInstance.load(features);

		// Get clusters within the bounding box at the given zoom level
		const clusters = superclusterInstance.getClusters(
			[minLng, minLat, maxLng, maxLat],
			zoomLevel
		);

		// Update cache with clusters instead of raw zones
		cache.set(cacheKey, { zones: clusters, timestamp: Date.now() });
		return NextResponse.json(clusters);
	} catch (error) {
		console.error('Error fetching zones:', error);
		return NextResponse.json({ error: 'Internal server error | gz200' }, { status: 500 });
	}
}
