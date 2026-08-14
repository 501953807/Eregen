/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./src/**/*.{html,js,ts,vue}",
    "./layouts/**/*.html",
    "./content/**/*.md",
  ],
  theme: {
    extend: {
      colors: {
        // Curative Brand
        'curative-blue': '#074ADF',
        'curative-blue-dark': '#1F40B8',
        'curative-blue-light': '#DCE5FF',
        'curative-sky': '#B6EEFF',

        // Warm neutrals
        'curative-warm': '#F6F3EB',
        'curative-warm-dark': '#EDE8DC',
        'curative-cream': '#FDFDFC',
        'curative-white': '#FDFDFD',

        // Text
        'curative-text': '#181B25',
        'curative-text-secondary': '#4D576F',
        'curative-text-muted': '#77778B',

        // Borders
        'curative-border': '#E9ECEF',
        'curative-border-strong': '#CED4DA',

        // Semantic
        'hope-success': '#198754',
        'hope-warning': '#D97706',
        'hope-error': '#DC3545',
        'hope-info': '#0DCAF0',
      },
      fontFamily: {
        sans: ['Inter', 'Helvetica', 'system-ui', 'sans-serif'],
        serif: ['Noto Serif SC', 'Georgia', 'serif'],
        mono: ['JetBrains Mono', 'Fira Code', 'monospace'],
      },
      fontSize: {
        'curative-h1': ['70px', { lineHeight: '77px', letterSpacing: '-0.02em' }],
        'curative-h2': ['50px', { lineHeight: '55px', letterSpacing: '-0.01em' }],
        'curative-h3': ['24px', { lineHeight: '26.4px', fontWeight: '600' }],
        'curative-h4': ['18px', { lineHeight: '24px', fontWeight: '600' }],
        'curative-body-lg': ['20px', { lineHeight: '30px' }],
        'curative-body': ['16px', { lineHeight: '24px' }],
        'curative-sm': ['14px', { lineHeight: '20px' }],
        'curative-xs': ['12px', { lineHeight: '16px' }],
      },
      spacing: {
        'curative-120': '120px',
        'curative-96': '96px',
        'curative-80': '80px',
        'curative-64': '64px',
      },
      borderRadius: {
        'curative-sm': '4px',
        'curative-md': '8px',
        'curative-lg': '12px',
        'curative-xl': '16px',
        'curative-2xl': '24px',
        'curative-3xl': '32px',
        'curative-full': '9999px',
      },
      boxShadow: {
        'curative-sm': '0 1px 2px rgba(0, 0, 0, 0.04)',
        'curative-md': '0 4px 12px rgba(0, 0, 0, 0.06)',
        'curative-lg': '0 8px 24px rgba(0, 0, 0, 0.08)',
        'curative-xl': '0 16px 48px rgba(0, 0, 0, 0.12)',
        'blue': '0 4px 16px rgba(7, 74, 223, 0.20)',
        'blue-lg': '0 8px 32px rgba(7, 74, 223, 0.25)',
      },
      maxWidth: {
        'curative-narrow': '800px',
        'curative-wide': '1200px',
      },
      animation: {
        'fade-in': 'fadeIn 0.6s ease-out forwards',
        'slide-up': 'slideUp 0.5s ease-out forwards',
        'float': 'float 4s ease-in-out infinite',
        'pulse-dot': 'pulseDot 2s ease-in-out infinite',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        slideUp: {
          '0%': { opacity: '0', transform: 'translateY(20px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
        float: {
          '0%, 100%': { transform: 'translateY(0)' },
          '50%': { transform: 'translateY(-10px)' },
        },
        pulseDot: {
          '0%, 100%': { opacity: '1', transform: 'scale(1)' },
          '50%': { opacity: '0.6', transform: 'scale(1.3)' },
        },
      },
      transitionDuration: {
        'curative-fast': '150ms',
        'curative-base': '200ms',
        'curative-slow': '300ms',
      },
    },
  },
  plugins: [],
}
