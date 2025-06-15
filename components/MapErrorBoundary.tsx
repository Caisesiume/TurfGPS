import React, { Component, ErrorInfo, ReactNode } from 'react';

interface Props {
  children: ReactNode;
}

interface State {
  hasError: boolean;
  error?: Error;
}

class MapErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error): State {
    // Update state so the next render will show the fallback UI
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('Map error caught by boundary:', error, errorInfo);
    
    // You could also log the error to an error reporting service here
    // For example: logErrorToService(error, errorInfo);
  }
  render() {
    if (this.state.hasError) {
      return (
        <div className='flex items-center justify-center h-full bg-gray-100 border-2 border-dashed border-gray-300 rounded-lg'>
          <div className='text-center p-6'>
            <div className='text-red-500 text-6xl mb-4'>⚠️</div>
            <h2 className='text-xl font-semibold text-gray-800 mb-2'>
              Map failed to load
            </h2>
            <p className='text-gray-600 mb-4'>
              There was an error loading the map. Please try refreshing the page.
            </p>
            <button
              onClick={() => window.location.reload()}
              className='px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 transition-colors'
            >
              Refresh Page
            </button>
            {process.env.NODE_ENV === 'development' && this.state.error && (
              <details className='mt-4 text-left'>
                <summary className='cursor-pointer text-sm text-gray-500'>
                  Error Details (Development Only)
                </summary>
                <pre className='mt-2 text-xs text-red-600 bg-red-50 p-2 rounded'>
                  {this.state.error.toString()}
                </pre>
              </details>
            )}
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

export default MapErrorBoundary;
