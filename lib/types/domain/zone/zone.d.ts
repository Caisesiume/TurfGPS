import { LocationCoordinates } from "../../global/coordinates";

export interface IZone {
    area: number; // Kommun
    currentOwner: {
        id: number;
        name: string;
    };
    dateCreated: string;
    dateLastTaken: string;
    zoneId: number;
    isActive: boolean;
    location: LocationCoordinates;
    name: string;
    owner: string;
    pointsPerHour: number;
    region: {
        area: {
            name: string;
            id: number;
        };
        country: string;
        name: string;
        id: number;
    };
    takeoverPoints: number;
    totalTakeovers: number;
    type: {
        id: number;
        name: string;
    }
}