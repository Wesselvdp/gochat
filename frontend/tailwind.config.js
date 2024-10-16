/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./src/**/*.{html,js}", "../views/**/*.{html,templ,go}"],
  theme: {
    colors: {
      'main': 'var(--main-bg-color)',
      'text': 'var(--text-color)'
    },
    extend: {},
  },
  plugins: [],
}

