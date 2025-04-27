/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./src/**/*.{html,js,ts,scss}", // Make sure this path correctly points to your Lit components
    "../views/**/*.{html,templ,go}",
    // Add any other paths where you use Tailwind classes
  ],
  // Reduced safelist to only contain classes that aren't covered by the theme mapping
  safelist: ["bg-background-bubble"],
  theme: {
    extend: {
      colors: {
        // --- Shadcn Core Colors ---
        border: "var(--border)",
        input: "var(--input)",
        ring: "var(--ring)",

        background: "var(--background)",
        foreground: "var(--foreground)",

        primary: {
          DEFAULT: "var(--primary)",
          foreground: "var(--primary-foreground)",
        },
        secondary: {
          DEFAULT: "var(--secondary)",
          foreground: "var(--secondary-foreground)",
        },
        destructive: {
          DEFAULT: "var(--destructive)",
          foreground: "var(--destructive-foreground)",
        },
        muted: {
          DEFAULT: "var(--muted)",
          foreground: "var(--muted-foreground)",
        },
        accent: {
          DEFAULT: "var(--accent)",
          foreground: "var(--accent-foreground)",
        },
        popover: {
          DEFAULT: "var(--popover)",
          foreground: "var(--popover-foreground)",
        },
        card: {
          DEFAULT: "var(--card)",
          foreground: "var(--card-foreground)",
        },

        // --- Background Levels for Depth ---
        "level-1": "var(--background-level-1)",
        "level-2": "var(--background-level-2)",
        "level-3": "var(--background-level-3)",
        "level-4": "var(--background-level-4)",

        // --- Legacy Support ---
        "background-bubble": "var(--background-level-3)", // Map to appropriate level (formerly --chat-bubble)
      },
      borderRadius: {
        lg: "var(--border-radius)",
        md: "calc(var(--border-radius) - 0.2rem)",
        sm: "calc(var(--border-radius) - 0.4rem)",
      },
      // You can add other theme extensions like keyframes, animations here if needed
      // keyframes: { ... },
      // animation: { ... },
    },
  },
  plugins: [],
  darkMode: "class", // Enable class-based dark mode for shadcn
};
