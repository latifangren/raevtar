/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./internal/view/**/*.templ",
    "./internal/view/**/*.go",
    "./internal/handler/*.go",
  ],
  theme: {
    extend: {
      fontFamily: {
        sans: ['Inter', 'system-ui', '-apple-system', 'sans-serif'],
        mono: ['"JetBrains Mono"', '"Fira Code"', 'monospace'],
      },
      colors: {
        retro: {
          cream: '#F5F2ED', // Background color from RetroUI
          paper: '#FFFFFF',
          ink: '#000000',   // Pure black for borders and text
          yellow: '#FACC15', // Vibrant yellow from RetroUI
          sage: '#6A9B7D',
          sageLight: '#DDEBDD',
          wheat: '#EFE2B8',
          blush: '#E9B7A5',
          muted: '#6B7280',
        },
        raevtar: {
          50:  '#FDFCF5',
          100: '#FFF8E8',
          200: '#EFE2B8',
          300: '#DDEBDD',
          400: '#8FB89B',
          500: '#6A9B7D',
          600: '#527C62',
          700: '#3F624D',
          800: '#344D3D',
          900: '#2D3748',
          950: '#1F2937',
        },
      },
    },
  },
  plugins: [],
}
