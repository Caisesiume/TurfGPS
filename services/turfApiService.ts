import { BoundBox } from "@/lib/types";

export async function fetchZonesInBounds(bounds: BoundBox) {
  try {
    console.log({ bounds });
    
    const urlWithParams = `
        /api/zones?northEastLat=${bounds.northEast.latitude}
            &northEastLng=${bounds.northEast.longitude}
            &southWestLat=${bounds.southWest.latitude}
            &southWestLng=${bounds.southWest.longitude}
    `.replace(/\s+/g, "");

    const response = await fetch(urlWithParams);

    if (!response.ok) {
      console.log({ response });
      throw new Error("Failed to fetch zones");
    }

    const data = await response.json();
    return data;
  } catch (error) {
    throw error;
  }
}
