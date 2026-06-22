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
        sans: ['"Space Grotesk"', 'Inter', 'system-ui', 'sans-serif'],
        mono: ['"JetBrains Mono"', '"Fira Code"', 'monospace'],
        head: ['"Archivo Black"', '"Inter"', 'system-ui', 'sans-serif'],
      },
      colors: {
        retro: {
          cream: '#F5F2ED',
          paper: '#FFFFFF',
          ink: '#000000',
          yellow: '#FACC15',
          sage: '#6A9B7D',
          sageLight: '#DDEBDD',
          wheat: '#EFE2B8',
          blush: '#E9B7A5',
          muted: '#6B7280',
        },
        // Semantic tokens mapped to retro palette (RETROUI-compatible)
        primary: {
          DEFAULT: '#000000',
          hover: '#1F2937',
          foreground: '#F5F2ED',
        },
        secondary: {
          DEFAULT: '#EFE2B8',
          hover: '#E5D4A8',
          foreground: '#000000',
        },
        accent: {
          DEFAULT: '#FACC15',
          hover: '#E6B800',
          foreground: '#000000',
        },
        muted: {
          DEFAULT: '#6B7280',
          foreground: '#FFFFFF',
        },
        destructive: {
          DEFAULT: '#E9B7A5',
          hover: '#DDA08B',
          foreground: '#000000',
        },
        success: {
          DEFAULT: '#6A9B7D',
          hover: '#527C62',
          foreground: '#FFFFFF',
        },
        card: {
          DEFAULT: '#FFFFFF',
          foreground: '#000000',
        },
        background: '#F5F2ED',
        foreground: '#000000',
        border: '#000000',
        ring: '#000000',
      },
      borderRadius: {
        sm: '2px',
        DEFAULT: '4px',
        md: '6px',
        lg: '8px',
      },
    },
  },
  plugins: [],
}
