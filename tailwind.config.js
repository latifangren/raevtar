/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./internal/view/**/*.templ",
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
          cream: '#FDFCF5',
          paper: '#FFF8E8',
          ink: '#2D3748',
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
