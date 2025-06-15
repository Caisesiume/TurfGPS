import React from 'react';

interface TurfGPSLogoProps {
	className?: string;
	size?: number;
}

export default function TurfGPSLogo({ className, size = 32 }: TurfGPSLogoProps) {
	return (
		<svg 
			width={size} 
			height={size} 
			viewBox='0 0 32 32' 
			fill='none' 
			xmlns='http://www.w3.org/2000/svg'
			className={className}
		>
			{/* Background circle */}
			<circle cx='16' cy='16' r='15' fill='#15803d' stroke='#22c55e' strokeWidth='2'/>
			
			{/* GPS crosshairs */}
			<circle cx='16' cy='16' r='10' fill='none' stroke='#22c55e' strokeWidth='1.5' opacity='0.6'/>
			<circle cx='16' cy='16' r='6' fill='none' stroke='#22c55e' strokeWidth='1'/>
			
			{/* Center dot (GPS point) */}
			<circle cx='16' cy='16' r='2' fill='#22c55e'/>
			
			{/* GPS signal lines */}
			<path d='M16 2 L16 6' stroke='#22c55e' strokeWidth='2' strokeLinecap='round'/>
			<path d='M16 26 L16 30' stroke='#22c55e' strokeWidth='2' strokeLinecap='round'/>
			<path d='M2 16 L6 16' stroke='#22c55e' strokeWidth='2' strokeLinecap='round'/>
			<path d='M26 16 L30 16' stroke='#22c55e' strokeWidth='2' strokeLinecap='round'/>
			
			{/* Corner zone indicators */}
			<rect x='7' y='7' width='3' height='3' fill='#22c55e' opacity='0.8'/>
			<rect x='22' y='7' width='3' height='3' fill='#22c55e' opacity='0.8'/>
			<rect x='7' y='22' width='3' height='3' fill='#22c55e' opacity='0.8'/>
			<rect x='22' y='22' width='3' height='3' fill='#22c55e' opacity='0.8'/>
		</svg>
	);
}
