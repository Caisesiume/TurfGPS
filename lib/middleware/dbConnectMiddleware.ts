/* eslint-disable @typescript-eslint/no-explicit-any */
import mongoose from 'mongoose';
import { NextRequest } from 'next/server';
import dbConnect from '../mongodb';

const dbConnectionMiddleware = async (req: NextRequest, next: () => Promise<Response>) => {
    if (mongoose.connection.readyState !== 1 
        && mongoose.connection.readyState !== 2
    ) {
        await dbConnect();
    }
    return next();
};

export default dbConnectionMiddleware;
