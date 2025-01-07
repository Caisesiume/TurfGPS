/* eslint-disable @typescript-eslint/no-require-imports */
/* eslint-disable @typescript-eslint/ban-ts-comment */

// Built for Node.js runtime
const fs = require('fs');
const path = require('path');

export const FetchZoneMode = {
	All: 0,
	One: 1,
	Some: 2,
};

//@ts-ignore
export async function fetchZones(fetchType, zoneIds = []) {
    try {
        let zones = [];
        let response;

        if (fetchType === FetchZoneMode.Some || fetchType === FetchZoneMode.One) {
            let queryZonens = [];
            if (zoneIds.length === 0) {
                queryZonens = [{ name: 'plattan' }, { name: 'ponyzone' }, { name: 'GAZone' }];
            } else {
                queryZonens = zoneIds.map((zoneId) => ({ id: zoneId }));
            }
            response = await fetch('https://api.turfgame.com/unstable/zones', {
                headers: {
                    'Content-Type': 'application/json',
                    Accept: '*/*',
                    'Accept-Encoding': 'gzip',
                },
                method: 'POST',
                body: JSON.stringify(queryZonens),
            });
            
            if (!response.ok) {
                throw new Error('Failed to fetch zones');
            }
            
            return await response.json();
        }

        if (fetchType === FetchZoneMode.All) {
            response = await fetch('https://api.turfgame.com/unstable/zones/all', {
                headers: {
                    'Content-Type': 'application/json',
                    Accept: '*/*',
                    'Accept-Encoding': 'gzip',
                },
                method: 'GET',
            });
        }

        // @ts-ignore
        zones = await response.json();

        const filePath = path.join(__dirname, 'zones.json');
        fs.writeFileSync(filePath, JSON.stringify(zones, null, 2));

        console.log('Zones data has been written to zones.json');
    } catch (error) {
        console.error('Error fetching zones:', error);
    }
}

fetchZones(FetchZoneMode.One);
