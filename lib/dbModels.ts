import mongoose from 'mongoose';
import serviceConfigSchema from './models/serviceConfigModel';
import zoneSchema from './models/zonesModel';
import userSchema from './models/userModel';
import type { IZone } from './types';
import type { IUser } from './types';

//#region Interfaces

interface IServiceConfig extends Document {
    lastZoneFetch?: Date;
    totalZones?: number;
    fetchedBy?: IUser;
    _id?: mongoose.Types.ObjectId;
}
interface IZoneModel extends IZone, Omit<Document, 'location'> {
    location: IZone['location'];
}
interface IUserModel extends IUser, Document { }

//#endregion

// REGULAR DATABASE
const db = mongoose.connection.useDb('TurfGPSData');
const serviceConfigModel = db.model<IServiceConfig>('ServiceConfig', serviceConfigSchema, 'ServiceConfig');
const zoneModel = db.model<IZoneModel>('Zone', zoneSchema, 'Zones');
const userModel = db.model<IUserModel>('User', userSchema, 'Users');

const models = {
    serviceConfigModel,
    zoneModel,
    userModel,
};
export default models;