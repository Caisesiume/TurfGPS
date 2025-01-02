import { NextRequest, NextResponse } from 'next/server';
import { gridBoxesCalculator } from '@/lib/gridBoxesCalculator';
import { BoundBox } from '@/lib/types/bound';

const lastRequestTime = new Map<string, number>();
const requestQueue: { req: NextRequest, resolve: (value: NextResponse | PromiseLike<NextResponse>) => void }[] = [];
let isProcessingQueue = false;

async function processQueue() {
	if (isProcessingQueue) return;
	isProcessingQueue = true;

	while (requestQueue.length > 0) {
		const { req, resolve } = requestQueue.shift()!;
		const response = await handleRequest(req);
		resolve(response);
		await new Promise(resolve => setTimeout(resolve, 50)); // Wait for 50ms between each queue item
	}

	isProcessingQueue = false;
}

async function handleRequest(req: NextRequest): Promise<NextResponse> {
	const clientIp = (req.headers.get('x-forwarded-for')?.split(',')[0].trim())
		|| req.headers.get('cf-connecting-ip')
		// eslint-disable-next-line @typescript-eslint/ban-ts-comment
		// @ts-ignore
		|| req.ip
		|| 'unknown';
	const now = Date.now();
	const lastTime = lastRequestTime.get(clientIp) || 0;

	if (now - lastTime < 1000) { // 60000 ms = 1 minute
		return NextResponse.json({ error: 'Rate limit exceeded. Try again later.' }, { status: 429 });
	}

	lastRequestTime.set(clientIp, now);

	try {
		const coordinatesMap = new Map(req.nextUrl.searchParams);
		console.log({ coordinatesMap });

		// Define bounds
		const northEast = {
			latitude: parseFloat(coordinatesMap.get('northEastLat')!),
			longitude: parseFloat(coordinatesMap.get('northEastLng')!),
		};

		const southWest = {
			latitude: parseFloat(coordinatesMap.get('southWestLat')!),
			longitude: parseFloat(coordinatesMap.get('southWestLng')!),
		};


		const boxSize = {
			lat: 0.025,
			lng: 0.05,
		};

		const [adjacentBoxes, operationData] = gridBoxesCalculator(northEast, southWest, boxSize);
		const gridBoxes: BoundBox[] = [...adjacentBoxes];

		console.log({ gridBoxes: JSON.stringify(gridBoxes), operationData });

		if (operationData.success) {
			console.log("---- CALL TO API FIRES ----");

			// Iterate over the gridBoxes and fetch the zones
			const zones = [];
			for (const box of gridBoxes) {
				const res = await fetch('https://api.example.com/zones', {
					method: 'POST',
					headers: {
						'Content-Type': 'application/json',
					},
					body: JSON.stringify(box),
				});

				if (!res.ok) {
					throw new Error(`${res.statusText}`);
				}

				const data = await res.json();
				zones.push(data);

				// Wait for 50ms between each call
				await new Promise(resolve => setTimeout(resolve, 500));
			}

			console.log({ zones });
			

			return NextResponse.json({ zones });
		}
		return NextResponse.json({ error: operationData.message }, { status: 500 });
	// eslint-disable-next-line @typescript-eslint/no-explicit-any
	} catch (error: unknown | any) {
		return NextResponse.json({ error: error.message }, { status: 500 });
	}
}

export async function GET(req: NextRequest) {
	return new Promise<NextResponse>((resolve) => {
		requestQueue.push({ req, resolve });
		processQueue();
	});
}
