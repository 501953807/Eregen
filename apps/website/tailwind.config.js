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
        primary: "#4A90D9",
        secondary: "#6B9BD1",
        success: "#2E7D32",
        warning: "#F57C00",
        danger: "#D32F2F",
        info: "#1976D2",
        background: "#FAFAFA",
        foreground: "#212121",
        cardBg: "#FFFFFF",
      },
      fontFamily: {
        sans: ['Inter', 'sans-serif'],
        heading: ['Noto Serif SC', 'serif'],
      },
      spacing: {
        "128": "32rem",
      },
    },
  },
  plugins: [],
}
