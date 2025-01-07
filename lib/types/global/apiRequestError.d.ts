/* eslint-disable @typescript-eslint/no-explicit-any */
/* eslint-disable @typescript-eslint/no-unused-vars */
import { StatusCode } from './statusCodes';

export interface IApiRequestError {
	message: string;
	statusCode: StatusCode;
	name: string;
	meta: any;
// eslint-disable-next-line semi
}