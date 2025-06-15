'use client';

import { MapPin } from 'lucide-react';
import React from 'react';

interface WavyRouteProps {
    className?: string;
}

const WavyRoute: React.FC<WavyRouteProps> = ({ className }) => {
    return (
        <div className={`relative w-full h-48 ${className || ''}`}>
            {/* Straight Line with Circles */}
            <svg
                className='absolute top-0 left-0 w-full h-full'
                xmlns='http://www.w3.org/2000/svg'
                viewBox='-2 0 104 10' // Adjusted viewBox to add padding for circles
            >
                {/* Wavy Line */}
                <path
                    d='M0 10 Q25 6, 50 8 T100 10' /* Straight line */
                    className='stroke-green-500 fill-none transition-all duration-1000 ease-in-out [stroke-width:2] animate-wave'
                />
                {/* Circle at the start of the line */}
                <circle
                    cx='0'
                    cy='10'
                    r='2'
                    className='fill-green-500'
                />
                {/* MapPin Component at the end of the line */}
                <foreignObject x='97' y='2' width='6' height='6'>
                    <div className='w-full h-full flex items-center text-green-800'>
                        <MapPin />
                    </div>
                </foreignObject>
                {/* Circle at the end of the line */}
                <circle
                    cx='100'
                    cy='10'
                    r='2'
                    className='fill-green-500'
                />
                {/* Zone at the max value of the curve */}
                <circle
                    cx='27'
                    cy='5'
                    r='3'
                    className='fill-red-500 opacity-75'
                />
                {/* Zone at the min value of the curve */}
                <circle
                    cx='70'
                    cy='17'
                    r='3'
                    className='fill-red-500 opacity-75'
                />
            </svg>
        </div>
    );
};

export default WavyRoute;