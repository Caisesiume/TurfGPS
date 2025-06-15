import React from 'react';

interface TurfGPSHeroImageProps {
	className?: string;
	width?: number;
	height?: number;
}

export default function TurfGPSHeroImage({ className, width = 600, height = 400 }: TurfGPSHeroImageProps) {
	return (
		<svg 
			width={width} 
			height={height} 
			viewBox='0 0 600 400' 
			fill='none' 
			xmlns='http://www.w3.org/2000/svg'
			className={className}
		>
			{/* Gradients */}
			<defs>
				<radialGradient id='bgGradient' cx='50%' cy='50%' r='50%'>
					<stop offset='0%' style={{ stopColor:'#1f2937', stopOpacity:1 }} />
					<stop offset='100%' style={{ stopColor:'#0f172a', stopOpacity:1 }} />
				</radialGradient>
				<linearGradient id='greenGradient' x1='0%' y1='0%' x2='100%' y2='100%'>
					<stop offset='0%' style={{ stopColor:'#22c55e', stopOpacity:1 }} />
					<stop offset='100%' style={{ stopColor:'#15803d', stopOpacity:1 }} />
				</linearGradient>
			</defs>

			{/* Background */}
			<rect width='600' height='400' fill='url(#bgGradient)' rx='12'/>
			
			{/* Road network background */}
			<g opacity='0.15'>
				{/* Main curved roads */}
				<path d='M0 120 Q200 80 400 120 Q500 140 600 130' stroke='#374151' strokeWidth='4' fill='none'/>
				<path d='M100 0 Q150 150 200 250 Q250 350 300 400' stroke='#374151' strokeWidth='4' fill='none'/>
				<path d='M0 280 Q250 320 450 260 Q550 240 600 250' stroke='#374151' strokeWidth='4' fill='none'/>
				
				{/* Secondary connecting roads */}
				<path d='M400 0 Q350 100 300 180 Q200 260 100 320' stroke='#374151' strokeWidth='3' fill='none'/>
				<path d='M150 400 Q250 300 350 200 Q450 100 550 50' stroke='#374151' strokeWidth='3' fill='none'/>
				
				{/* Road intersections */}
				<circle cx='200' cy='120' r='3' fill='#374151'/>
				<circle cx='300' cy='180' r='3' fill='#374151'/>
				<circle cx='250' cy='260' r='3' fill='#374151'/>
				<circle cx='350' cy='200' r='2' fill='#374151'/>
			</g>

			{/* Round green zone under the man */}
			<circle cx='300' cy='198' r='8' fill='url(#greenGradient)'/>

			{/* Turf GPS Man */}
			<image
				href='/TurfZoneTakenIcon.png'
				x='282'
				y='160'
				width='36'
				height='36'
			/>

			{/* Zone markers scattered around */}
			<rect x='150' y='120' width='12' height='12' fill='#ef4444' opacity='0.8' rx='2'/>
			<rect x='420' y='150' width='12' height='12' fill='#ef4444' opacity='0.8' rx='2'/>
			<rect x='180' y='280' width='12' height='12' fill='#ef4444' opacity='0.8' rx='2'/>
			<rect x='450' y='260' width='12' height='12' fill='#ef4444' opacity='0.8' rx='2'/>
			<rect x='120' y='250' width='12' height='12' fill='#ef4444' opacity='0.8' rx='2'/>
			<rect x='380' y='110' width='12' height='12' fill='#ef4444' opacity='0.8' rx='2'/>
			
			{/* Route path */}
			<path 
				d='M100 300 Q200 250 300 200 T500 120' 
				stroke='#22c55e' 
				strokeWidth='3' 
				fill='none' 
				strokeDasharray='8,4'
				opacity='0.6'
			/>
			
			{/* Navigation compass in corner */}
			<g transform='translate(500, 10)'>
				<circle cx='40' cy='40' r='35' fill='#0f172a' stroke='#22c55e' strokeWidth='2'/>
				<path d='M40 15 L50 35 L40 30 L30 35 Z' fill='#22c55e'/>
				<text x='40' y='60' textAnchor='middle' fill='#22c55e' fontSize='8' fontWeight='bold'>N</text>
			</g>
			
		</svg>
	);
}
