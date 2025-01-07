import { IUser } from "@/lib/types";
import { ObjectId } from "mongoose";

interface ServiceConfig {
    lastZoneFetch?: Date | string;
    totalZones?: number;
    fetchedBy?: IUser;
    _id?: ObjectId;
}