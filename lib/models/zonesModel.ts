import { Schema } from 'mongoose';

const ZoneSchema: Schema = new Schema({
	currentOwner: {
		id: { type: Number, required: false },
		name: { type: String, required: false },
	},
	dateCreated: { type: String, required: true },
	dateLastTaken: { type: String, required: false },
	zoneId: { type: Number, required: true },
	isActive: { type: Boolean, required: false },
	location: {
		coordinates: { type: [Number], required: true },
		type: { type: String, enum: ['Point'], required: true },
	},
	name: { type: String, required: true },
	owner: { type: String, required: false },
	pointsPerHour: { type: Number, required: true },
	region: {
		area: {
			name: { type: String, required: false },
			id: { type: Number, required: false },
		},
		country: { type: String, required: false },
		name: { type: String, required: true },
		id: { type: Number, required: true },
	},
	takeoverPoints: { type: Number, required: true },
	totalTakeovers: { type: Number, required: true },
	type: {
		id: { type: Number, required: false },
		name: { type: String, required: false },
	},
});

ZoneSchema.index({ location: '2dsphere' });

export default ZoneSchema;
