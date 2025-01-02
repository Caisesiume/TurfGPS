export interface Zone {
    area: number; // Kommun
    currentOwner: {
        id: number;
        name: string;
    };
    dateCreated: string;
    dateLastTaken: string;
    id: number;
    isActive: boolean;
    latitude: number;
    longitude: number;
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
}