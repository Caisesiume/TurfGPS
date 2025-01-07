/* eslint-disable @typescript-eslint/no-explicit-any */

export class ApiRequestError extends Error {
	public statusCode: number;
	public meta: any;
	constructor(message: string, statusCode: number, meta?: any) {
		super(message);
		this.name = 'ApiRequestError';
		this.statusCode = statusCode;
		this.meta = meta || {};
	}
}
