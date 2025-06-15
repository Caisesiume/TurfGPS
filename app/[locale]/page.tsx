'use client';

import Link from 'next/link';
import Image from 'next/image';
import { useTranslations, useLocale } from 'next-intl';
import { ArrowRight, MapPin, Github, Menu, Route } from 'lucide-react';
import { useEffect, useState } from 'react';
import WavyRoute from '@/components/WavyRoute';
import TurfGPSLogo from '@/components/TurfGPSLogo';
import TurfGPSHeroImage from '@/components/TurfGPSHeroImage';


export default function LandingPage() {
	const t = useTranslations('landing');
	const locale = useLocale();
	const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);


	const links = {
		map: `/${locale}/map`,
		aboutUs: `/${locale}#about`,
		faq: `/${locale}#faq`,
		features: `/${locale}#features`,
	};

	// Smooth scroll to sections
	const scrollToSection = (sectionId: string) => {
		const section = document.getElementById(sectionId);
		if (section) {
			// Account for fixed header height
			const headerHeight = 64; // 4rem = 64px
			const elementPosition = section.offsetTop - headerHeight;
			
			window.scrollTo({
				top: elementPosition,
				behavior: 'smooth',
			});
		}
	};

	useEffect(() => {
		// Close mobile menu on route change
		const handleRouteChange = () => {
			if (isMobileMenuOpen) {
				setIsMobileMenuOpen(false);
			}
		};
		window.addEventListener('popstate', handleRouteChange);
		return () => {
			window.removeEventListener('popstate', handleRouteChange);
		};
	}, [isMobileMenuOpen]);

	// Handle hash navigation on page load
	useEffect(() => {
		const hash = window.location.hash.substring(1); // Remove the '#'
		if (hash) {
			// Small delay to ensure page is fully loaded
			setTimeout(() => {
				scrollToSection(hash);
			}, 100);
		}
	}, []);

	return (
		<div className='min-h-screen bg-green-950'>
			{/* Header */}
			<header className='fixed top-0 left-0 right-0 z-40 bg-green-950'>
				<div className='flex h-16 items-center justify-between px-4'>
					<Link href='/' className='flex items-center gap-2 text-green-500 text-xl font-bold'>
						<TurfGPSLogo size={32} />
						TURF GPS
					</Link>

					{/* Desktop Navigation */}
					<nav className='hidden md:flex items-center gap-6'>
						<Link href={links.features} className='text-gray-300 hover:text-green-500'>
							{t('navigation.features')}
						</Link>
						<Link href={links.aboutUs} className='text-gray-300 hover:text-green-500' onClick={(e) => {
							e.preventDefault();
							scrollToSection('about');
						}}>
							{t('navigation.about')}
						</Link>
						<Link href={links.faq} className='text-gray-300 hover:text-green-500' onClick={(e) => {
							e.preventDefault();
							scrollToSection('faq');
						}}>
							{t('navigation.faq')}
						</Link>
						<Link href={links.map} className='text-gray-300 hover:text-green-500'>
							{t('navigation.routePlanner')}
						</Link>
						<Link
							href={links.map}
							className='px-4 py-2 bg-green-600 text-white rounded hover:bg-green-500'
						>
							{t('navigation.getStarted')}
						</Link>
					</nav>

					{/* Mobile Menu Button */}
					<button className='md:hidden text-green-500' onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}>
						<Menu className='h-6 w-6' />
					</button>

					{/* Mobile Navigation */}
					{isMobileMenuOpen && (
						<nav className='md:hidden absolute top-16 left-0 right-0 bg-green-950 z-30'>
							<div className='flex flex-col items-center gap-4 py-4'>
								<Link href={links.features} className='text-gray-300 hover:text-green-500' onClick={() => setIsMobileMenuOpen(false)}>
									{t('navigation.features')}
								</Link>
								<Link href={links.aboutUs} className='text-gray-300 hover:text-green-500' onClick={(e) => {
									e.preventDefault();
									scrollToSection('about');
									setIsMobileMenuOpen(false);
								}}>
									{t('navigation.about')}
								</Link>
								<Link href={links.faq} className='text-gray-300 hover:text-green-500' onClick={(e) => {
									e.preventDefault();
									scrollToSection('faq');
									setIsMobileMenuOpen(false);
								}}>
									{t('navigation.faq')}
								</Link>
								<Link href={links.map} className='text-gray-300 hover:text-green-500' onClick={() => setIsMobileMenuOpen(false)}>
									{t('navigation.routePlanner')}
								</Link>
								<Link href={links.map} className='px-4 py-2 bg-green-600 text-white rounded hover:bg-green-500' onClick={() => setIsMobileMenuOpen(false)}>
									{t('navigation.getStarted')}
								</Link>
							</div>
						</nav>
					)}
				</div>
			</header>

			<main className='pt-16'>
				{/* Hero Section */}
				<section className='py-24 md:py-24' style={{ height: 'calc(100vh - 4rem)', justifySelf: 'center', alignContent: 'center' }}>
					<div className='max-w-7xl mx-auto px-4 grid lg:grid-cols-2 gap-12 items-center flex-grow self-center'>
						<div className='space-y-6'>
							<h1 className='text-4xl md:text-5xl font-bold text-green-500'>
								{t('hero.title')}
							</h1>
							<p className='text-gray-300 text-lg'>{t('hero.description')}</p>
							<div className='flex flex-col sm:flex-row gap-4'>
								<Link
									href={links.map}
									className='inline-flex items-center justify-center px-6 py-3 bg-green-600 text-white rounded hover:bg-green-500'
								>
									{t('hero.tryButton')}
								</Link>
								<Link
									href={links.features}
									className='inline-flex items-center justify-center px-6 py-3 border border-green-600 text-green-500 rounded hover:bg-green-900'
									onClick={(e) => {
										e.preventDefault();
										scrollToSection('features');
									}}
								>
									{t('hero.learnButton')}
								</Link>
							</div>
							<p className='text-sm text-gray-400'>{t('hero.freeText')}</p>
						</div>
						<div className='hidden lg:block'>
							<TurfGPSHeroImage 
								width={600} 
								height={400} 
								className='w-full h-auto rounded-lg border border-green-800'
							/>
						</div>
					</div>
				</section>

				{/* Features Section */}
				<section id='features' className='py-24 bg-green-900/50'>
					<div className='max-w-7xl mx-auto px-4'>
						<div className='text-center space-y-4'>
							<div className='inline-block px-4 py-1 bg-green-600 rounded-full text-white text-sm'>
								{t('features.badge')}
							</div>
							<h2 className='text-3xl md:text-4xl font-bold text-green-500'>
								{t('features.title')}
							</h2>
							<p className='text-gray-300 max-w-2xl mx-auto'>
								{t('features.description')}
							</p>
						</div>

						<div className='mt-16 grid md:grid-cols-2 gap-8'>
							<div className='space-y-12'>

								{/* Route planning */}
								<div className='flex gap-4'>
									<Route className='h-6 w-6 text-green-500 flex-shrink-0' />
									<div>
										<h3 className='font-bold text-green-500 mb-2'>
											{t('features.routePlanning.title')}
										</h3>
										<p className='text-gray-300'>
											{t('features.routePlanning.description')}
										</p>
									</div>
								</div>

								{/* Search and add zones to your routes */}
								<div className='flex gap-4'>
									<MapPin className='h-6 w-6 text-green-500 flex-shrink-0' />
									<div>
										<h3 className='font-bold text-green-500 mb-2'>
											{t('features.zoneFinder.title')}
										</h3>
										<p className='text-gray-300'>
											{t('features.zoneFinder.description')}
										</p>
									</div>
								</div>

								{/* Diverte journey via zones */}
								<div className='flex gap-4'>
									<Route className='h-6 w-6 text-green-500 flex-shrink-0' />
									<div>
										<h3 className='font-bold text-green-500 mb-2'>
											{t('features.divertJourney.title')}
										</h3>
										<p className='text-gray-300'>
											{t('features.divertJourney.description')}
										</p>
									</div>
								</div>
								<WavyRoute className='hidden md:block' />
							</div>
							<div className='hidden md:block'>
								<Image
									src='/placeholder.svg?height=400&width=600'
									alt='Features'
									width={600}
									height={400}
									className='rounded-lg border border-green-800'
								/>
							</div>
						</div>
					</div>
				</section>

				{/* About Section */}
				{/* <section id='about' className='py-24 bg-green-950'>
					<div className='max-w-7xl mx-auto px-4'>
						<div className='grid gap-12 lg:grid-cols-2'>
							<div className='flex items-center justify-center'>
								<Image
									src='/placeholder.svg?height=550&width=550'
									alt='About TurfGPS'
									width={550}
									height={550}
									className='rounded-lg object-cover border border-green-800'
								/>
							</div>
							<div className='flex flex-col justify-center space-y-4'>
								<div className='space-y-2'>
									<div className='inline-block rounded-lg bg-green-700 px-3 py-1 text-sm text-white'>
										{t('about.badge')}
									</div>
									<h2 className='text-3xl font-bold tracking-tighter text-green-500 md:text-4xl/tight'>
										{t('about.title')}
									</h2>
									<p className='text-gray-300 md:text-xl/relaxed'>
										{t('about.description1')}
									</p>
									<p className='text-gray-300 md:text-xl/relaxed'>
										{t('about.description2')}
									</p>
								</div>
								<div className='flex flex-col gap-2 min-[400px]:flex-row'>
									<Link
										href='/route-planner'
										className='inline-flex h-10 items-center justify-center rounded-md bg-green-700 px-4 py-2 text-sm font-medium text-white shadow transition-colors hover:bg-green-600 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-green-500'
									>
										{t('about.tryButton')}{' '}
										<ArrowRight className='ml-1 h-4 w-4' />
									</Link>
									<Link
										href='https://github.com/Caisesiume/TurfGPS'
										target='_blank'
										rel='noopener noreferrer'
										className='inline-flex h-10 items-center justify-center rounded-md border border-green-700 bg-transparent px-4 py-2 text-sm font-medium text-green-500 shadow transition-colors hover:bg-green-900 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-green-500'
									>
										<Github className='mr-2 h-4 w-4' />{' '}
										{t('about.githubButton')}
									</Link>
								</div>
							</div>
						</div>
					</div>
				</section> */}

				{/* FAQ Section */}
				<section id='faq' className='py-24 bg-green-900/50'>
					<div className='max-w-7xl mx-auto px-4'>
						<div className='text-center space-y-4'>
							<div className='inline-block rounded-lg bg-green-700 px-3 py-1 text-sm text-white'>
								{t('faq.badge')}
							</div>
							<h2 className='text-3xl font-bold tracking-tighter text-green-500 md:text-4xl/tight'>
								{t('faq.title')}
							</h2>
							<p className='max-w-2xl mx-auto text-gray-300 md:text-xl/relaxed'>
								{t('faq.description')}
							</p>
						</div>
						<div className='mt-16 grid gap-6'>
							{[0, 1, 2, 3, 4].map((index) => (
								<div
									key={index}
									className='rounded-lg border border-green-800 bg-green-950 p-6'
								>
									<h3 className='text-lg font-bold text-green-500 mb-2'>
										{t(`faq.questions.${index}.question`)}
									</h3>
									<p className='text-gray-300'>
										{t(`faq.questions.${index}.answer`)}
									</p>
								</div>
							))}
						</div>
					</div>
				</section>

				{/* CTA Section */}
				<section id='cta' className='py-24 bg-green-900/50'>
					<div className='max-w-7xl items-center mx-auto px-4'>
						<div className='text-center space-y-4'>
							<h2 className='text-3xl font-bold tracking-tighter text-green-500 md:text-4xl/tight'>
								{t('cta.title')}
							</h2>
							<p className='max-w-2xl mx-auto text-gray-300 md:text-xl/relaxed'>
								{t('cta.description')}
							</p>
						</div>
						<div className='mt-8 flex flex-col gap-2 min-[400px]:flex-row justify-center'>
							<Link
								href='/route-planner'
								className='inline-flex h-10 items-center justify-center rounded-md bg-green-700 px-4 py-2 text-sm font-medium text-white shadow transition-colors hover:bg-green-600 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-green-500'
							>
								{t('cta.tryButton')} <ArrowRight className='ml-1 h-4 w-4' />
							</Link>
							<Link
								href='#features'
								className='inline-flex h-10 items-center justify-center rounded-md border border-green-700 bg-transparent px-4 py-2 text-sm font-medium text-green-500 shadow transition-colors hover:bg-green-900 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-green-500'
							>
								{t('cta.exploreButton')}
							</Link>
						</div>
						<p className='mt-4 text-xs text-gray-400 text-center'>{t('hero.freeText')}</p>
					</div>
				</section>
			</main>

			{/* Footer */}
			<footer className='w-full border-t border-green-900 bg-green-950 py-6 md:py-12'>
				<div className='max-w-7xl mx-auto px-4'>
					<div className='grid gap-8 md:grid-cols-2 lg:grid-cols-4'>
						{/* Company Info */}
						<div className='space-y-4'>
							<div className='flex items-center gap-2'>
								<TurfGPSLogo size={32} className='rounded bg-green-700' />
								<span className='text-xl font-bold text-green-500'>TURF GPS</span>
							</div>
							<p className='text-sm text-gray-400'>{t('footer.description')}</p>
							<div className='flex gap-4'>
								<Link
									href='https://github.com/caisesiume/turfgps'
									className='text-gray-400 hover:text-green-500'
								>
									<Github className='h-5 w-5' />
									<span className='sr-only'>GitHub</span>
								</Link>
							</div>
						</div>

						{/* Tools Section */}
						<div className='space-y-4'>
							<h4 className='text-sm font-bold text-green-500'>
								{t('footer.tools')}
							</h4>
							<ul className='space-y-2 text-sm'>
								<li>
									<Link
										href={`/${locale}/map`}
										className='text-gray-400 hover:text-green-500'
									>
										{t('footer.routePlanner')}
									</Link>
								</li>
								{/* <li>
									<Link href='#' className='text-gray-400 hover:text-green-500'>
										{t('footer.zoneFinder')}
									</Link>
								</li>
								<li>
									<Link href='#' className='text-gray-400 hover:text-green-500'>
										{t('footer.statistics')}
									</Link>
								</li> */}
							</ul>
						</div>

						{/* Links Section */}
						<div className='space-y-4'>
							<h4 className='text-sm font-bold text-green-500'>
								{t('footer.links')}
							</h4>
							<ul className='space-y-2 text-sm'>
								<li>
									<Link
										href='#about'
										className='text-gray-400 hover:text-green-500'
										onClick={(e) => {
											e.preventDefault();
											scrollToSection('about');
										}}
									>
										{t('footer.about')}
									</Link>
								</li>
								<li>
									<Link
										href='#faq'
										className='text-gray-400 hover:text-green-500'
										onClick={(e) => {
											e.preventDefault();
											scrollToSection('faq');
										}}
									>
										{t('footer.faq')}
									</Link>
								</li>
								<li>
									<Link
										href='https://github.com/caisesiume/turfgps'
										className='text-gray-400 hover:text-green-500'
										target='_blank'
										rel='noopener noreferrer'
									>
										{t('footer.github')}
									</Link>
								</li>
								<li>
									<Link
										href='https://turfgame.com'
										className='text-gray-400 hover:text-green-500'
										target='_blank'
										rel='noopener noreferrer'
									>
										{t('footer.officialTurfgame')}
									</Link>
								</li>
							</ul>
						</div>

						{/* Resources Section */}
						<div className='space-y-4'>
							<h4 className='text-sm font-bold text-green-500'>
								Resources
							</h4>
							<ul className='space-y-2 text-sm'>
								<li>
									<Link href='#' className='text-gray-400 hover:text-green-500'>
										{t('footer.privacyPolicy')}
									</Link>
								</li>
								<li>
									<Link href='#' className='text-gray-400 hover:text-green-500'>
										{t('footer.termsOfUse')}
									</Link>
								</li>
							</ul>
						</div>
					</div>

					{/* Copyright Section */}
					<div className='mt-8 pt-8 border-t border-green-900 flex flex-col md:flex-row justify-between items-center'>
						<p className='text-xs text-gray-400'>
							{t('footer.copyright', { year: new Date().getFullYear() })}
						</p>
					</div>
				</div>
			</footer>
		</div>
	);
}
