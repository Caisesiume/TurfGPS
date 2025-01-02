interface BoundingBox {
  northEast: {
    lat: number;
    lng: number;
  };
  southWest: {
    lat: number;
    lng: number;
  };
}

export async function fetchZonesInBounds(bounds: BoundingBox) {
  try {
    console.log({ bounds });
    
    const urlWithParams = `
        /api/zones?northEastLat=${bounds.northEast.lat}
            &northEastLng=${bounds.northEast.lng}
            &southWestLat=${bounds.southWest.lat}
            &southWestLng=${bounds.southWest.lng}
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
