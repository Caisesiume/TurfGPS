import mongoose, { Schema } from 'mongoose';

const ServiceConfigSchema: Schema = new Schema({
  lastZoneFetch: { type: Date, required: true, default: Date.now },
  totalZones: { type: Number, required: true, default: 0 },
  fetchedBy: { type: mongoose.Types.ObjectId, ref: "Users", required: false, default: null },
});

export default ServiceConfigSchema;