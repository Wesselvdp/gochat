/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./src/**/*.{html,js}", "../views/**/*.{html,templ,go}", "./views/**/*.{html,templ,go}"],
  safelist: [
    'bg-main',
    'text-main',
    'bg-tertiary',
    'bg-background-bubble'
  ],
  theme: {
    extend: {
      colors: {
        'main': 'var(--main-bg-color)',
        'brand-primary': 'var(--primary-brand-color)',
        'brand-secondary': 'var(--secondary-brand-color)',
        // background
        'background-primary': 'var(--primary-background-color)',
        'background-secondary': 'var(--secondary-background-color)',
        'background-tertiary': 'var(--tertiary-background-color)',

        'text': 'var(--text-color)',
        'heading': 'var(--heading-color)',
        'background-bubble': 'var(--chat-bubble)'
      },
    },
  },
  plugins: [],
}

