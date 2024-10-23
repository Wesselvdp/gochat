/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./src/**/*.{html,js}", "../views/**/*.{html,templ,go}", "./views/**/*.{html,templ,go}"],
  safelist: [
    'bg-main',
    'text-main'
  ],
  theme: {
    extend: {
      colors: {
        'main': 'var(--main-bg-color)',
        'text': 'var(--text-color)'
      },
    },
  },
  plugins: [],
}

