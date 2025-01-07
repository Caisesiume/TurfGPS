import { Schema } from 'mongoose';


const UserSchema: Schema = new Schema({
    blocktime: { type: Number, required: true },
    country: { type: String, required: true },
    userId: { type: Number, required: true }, // renamed from id to userId
    medals: { type: [Number], required: true },
    name: { type: String, required: true },
    place: { type: Number, required: true },
    points: { type: Number, required: true },
    pointsPerHour: { type: Number, required: true },
    rank: { type: Number, required: true },
    region: {
        id: { type: Number, required: true },
        name: { type: String, required: true },
    },
    taken: { type: Number, required: true },
    totalPoints: { type: Number, required: true },
    uniqueZonesTaken: { type: Number, required: true },
    zones: { type: [Number], required: true },
});

export default UserSchema;