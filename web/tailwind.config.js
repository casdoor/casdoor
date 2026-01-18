/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./src/**/*.{js,jsx,ts,tsx}",
    "./public/index.html",
  ],
  theme: {
    extend: {
      colors: {
        primary: "rgb(89, 54, 213)",
      },
      borderRadius: {
        DEFAULT: "5px",
      },
    },
  },
  plugins: [],
  // Prevent Tailwind from conflicting with Ant Design during migration
  corePlugins: {
    preflight: false,
  },
}
