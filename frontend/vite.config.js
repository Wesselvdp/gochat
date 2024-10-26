import { defineConfig } from "vite";
import path from "path";

export default defineConfig({
  build: {
    target: "modules",
    outDir: "dist",
    sourcemap: "inline",
    lib: {
      entry: path.resolve(__dirname, "src/index.ts"),
      name: "gochat",
      formats: ["iife"], // This ensures a single bundle
      fileName: () => "bundle.js", // This names your output file
    },
    rollupOptions: {
      output: {
        // Ensure no chunk splitting
        manualChunks: undefined,
      },
    },
    emptyOutDir: true,
    css: {
      preprocessorOptions: {
        scss: {
          sourcemap: true,
          // If you have any SCSS options
        }
      }
    }
  },
  resolve: {
    alias: [
      {
        // this is required for the SCSS modules
        find: /^~(.*)$/,
        replacement: '$1',
      },
    ],
  },
});
