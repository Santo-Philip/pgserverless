/** @type {import('tailwindcss').Config} */
export default {
	content: ['./src/**/*.{html,js,svelte,ts}'],
	theme: {
		extend: {
			colors: {
				nexbic: {
					50: '#f0f7ff',
					100: '#e0effe',
					200: '#bae0fd',
					300: '#7cc8fb',
					400: '#36aaf5',
					500: '#0c8ee5',
					600: '#0071c3',
					700: '#015a9e',
					800: '#064d83',
					900: '#0b416d',
				}
			}
		}
	},
	plugins: []
};
