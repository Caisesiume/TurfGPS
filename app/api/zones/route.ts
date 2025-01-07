import { NextRequest, NextResponse } from 'next/server';
import { getHandler } from './GET';
import { ApiRequestError } from '@/lib/errors';
import dbConnectionMiddleware from '@/lib/middleware/dbConnectMiddleware';
// Import custom middleware functions
// import { authenticate } from '@/lib/middleware/authenticate';

export async function GET(req: NextRequest) {
	return handler(req, getHandler);
}

async function handler(
	req: NextRequest,
	methodHandler: (req: NextRequest) => Promise<NextResponse>,
) {
	try {
		// Step 1: Run Authentication Middleware
		// await authenticate(req);

		// Step 2: Run Database Connection Middleware
		await dbConnectionMiddleware(req, async () => {
			// Return a dummy response to satisfy the type requirement
			return NextResponse.next();
		});

		// Delegate to the method-specific handler
		return await methodHandler(req);
	} catch (error) {
		if (error instanceof ApiRequestError) {
			return NextResponse.json({ error: error.message }, { status: error.statusCode });
		}

		console.error('Unexpected error:', error);
		return NextResponse.json({ error: 'Internal Server Error' }, { status: 500 });
	}
}
