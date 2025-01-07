export interface IUser {
    blocktime: number;
    country: string;
    userId: number;
    medals: number[];
    name: string;
    place: number;
    points: number;
    pointsPerHour: number;
    rank: number;
    region: {
        id: number;
        name: string;
    };
    taken: number;
    totalPoints: number;
    uniqueZonesTaken: number;
    zones: number[];
}