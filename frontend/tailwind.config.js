/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,jsx}'],
  darkMode: 'class',
  theme: {
    fontFamily: {
      sans: ['Roboto', 'ui-sans-serif', 'system-ui', 'sans-serif'],
    },
    extend: {
      colors: {
        'wg-red': '#E81410',
        'wg-red-hover': '#B32317',
        'wg-gray-light': '#EAEAEA',
        'wg-body': '#666666',
        'wg-headline': '#2D3237',
        'wg-blue': '#035996',
        'wg-blue-hover': '#002663',
        'wg-black': '#000000',
      },
      fontSize: {
        body: ['14px', '24px'],
      },
      animation: {
        'fade-in': 'fadeIn 0.4s ease-out',
        'slide-up': 'slideUp 0.5s ease-out',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        slideUp: {
          '0%': { opacity: '0', transform: 'translateY(16px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
      },
    },
  },
  plugins: [],
}
