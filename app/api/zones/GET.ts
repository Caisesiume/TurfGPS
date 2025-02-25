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
	const minLng = parseFloat(searchParams.get('minLng')!);
	const maxLng = parseFloat(searchParams.get('maxLng')!);
	const minLat = parseFloat(searchParams.get('minLat')!);
	const maxLat = parseFloat(searchParams.get('maxLat')!);

	const cacheKey = `tile:${maxLat},${maxLng}:${minLat},${minLng}`;

	// Check in-memory cache
	const cachedData = cache.get(cacheKey);
	if (cachedData && Date.now() - cachedData.timestamp < CACHE_EXPIRY_MS) {
		return NextResponse.json(cachedData.zones);
	}

	const zones = await fetchZonesFromDatabase(
		minLng,
		maxLng,
		minLat,
		maxLat,
	);
	console.log(' ---- Fetched zones from database ---- ');
	
	
	// Update in-memory cache
	cache.set(cacheKey, { zones, timestamp: Date.now() });
	console.log(' ---- Updated in-memory cache ---- ');
	
	return NextResponse.json(zones);
}
