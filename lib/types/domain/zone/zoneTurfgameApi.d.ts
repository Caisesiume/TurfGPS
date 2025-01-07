// Type definitions for zone entity from Turfgame API

export interface IZoneTurfgameApi {
    dateCreated: string; // Value will be on the time element, fetched from /feed
    time?: string; // Value will be on the time element, fetched from /feed
    id: number;
    latitude: number;
    longitude: number;
    name: string;
    pointsPerHour: number;
    region: {
        area: {
            id: number;
            name: string;
        };
        country: string;
        id: number;
        name: string;
    };
    takeoverPoints: number;
    totalTakeovers: number;
}