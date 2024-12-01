import { defineConfig, loadEnv } from "vite";
import { sentryVitePlugin } from "@sentry/vite-plugin";
import { viteStaticCopy } from 'vite-plugin-static-copy'
import path from "path";
import tailwindcss from 'tailwindcss'
import autoprefixer from "autoprefixer";

export default defineConfig(({  mode }) => {
  const env = loadEnv(mode, process.cwd(), '')

  return {  build: {
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
            plugins: [tailwindcss, autoprefixer],
            // If you have any SCSS options
          }
        }
      }
    },

    plugins: [

        viteStaticCopy({
            targets: [
                {
                    src: path.resolve(__dirname, './src/themes') + '/[!.]*', // 1️⃣
                    dest: './themes/', // 2️⃣
                },
            ],
        }),

      mode === 'production' && sentryVitePlugin({
        org: "torgonio",
        project: "go",
        authToken: env.SENTRY_AUTH_TOKEN,
      }),
    ],
    resolve: {
      alias: [
        {
          // this is required for the SCSS modules
          find: /^~(.*)$/,
          replacement: '$1',
        },
      ],
    },
  }
});
