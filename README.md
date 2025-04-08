# TurfGPS

TurfGPS is a web application designed to provide an interactive map interface for visualizing and managing routes between turf zones.

## Features

- **✅ Interactive Map**: Powered by Leaflet, the map supports zooming, panning, and dynamic updates.
- **✅ Zone Clustering**: Zones are clustered dynamically based on the current zoom level for better visualization.
- **✅ Responsive Design**: The application is fully responsive and works seamlessly across devices.
- **✅ Internationalization (i18n)**: Support for multiple languages.
- **✅ Route Planning Algorithm**: The application includes algorithm to optimize the path between zones.

## Tech Stack

- **Frontend**: React, Next.js, TypeScript
- **Mapping**: Leaflet with React-Leaflet
- **Styling**: Tailwind CSS
- **State Management**: Redux Toolkit as centralized state management
- **Utilities**: Lodash, custom helper functions in `lib/`
- **Testing**: Jest for unit testing
- **Build Tools**: PostCSS, Tailwind CSS

### Key Directories

- **`app/`**: Contains global assets like `favicon.ico`, global styles, and locale-specific directories.
- **`components/`**: Reusable React components, including the `LeafletMapComponent` for map rendering.
- **`i18n/`**: Internationalization files for multi-language support.
- **`lib/`**: Utility functions and helpers, such as `computeRadius` and `getMoveDistanceThreshold`.
- **`services/`**: Service layer for API calls and business logic.
- **`store/`**: State management logic.

## Getting Started

### Prerequisites

- Node.js (v20 or higher)
- npm

## Development

### Linting and Formatting

- Run ESLint:
  ```bash
  npm run lint
  ```

- Format code with Prettier:
  ```bash
  npm run format
  ```

### Testing

Run unit tests with Jest:
```bash
npm test
```

## Key Components

### LeafletMapComponent

The `LeafletMapComponent` is the core map interface, located in [`components/LeafletMapComponent.tsx`](components/LeafletMapComponent.tsx). It integrates with Leaflet and React-Leaflet to display zones dynamically based on the map's bounds and zoom level. Key features include:

- Fetching zones from the API based on map bounds.
- Dynamic clustering of zones.
- Event handling for `moveend` and `zoomend` to update the map.

### Utility Functions

Located in [`lib/`](lib/), utility functions like `computeRadius` and `getMoveDistanceThreshold` are used to calculate cluster sizes and movement thresholds dynamically.

## Contributing

Contributions are welcome! Please contact @Caisesiume in Turf discord if interested.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Leaflet](https://leafletjs.com/) for the mapping library.
- [React-Leaflet](https://react-leaflet.js.org/) for seamless React integration.
- [OpenStreetMap](https://www.openstreetmap.org/) for map tiles.

## Contact

For questions or support, please open an issue or contact the maintainers.