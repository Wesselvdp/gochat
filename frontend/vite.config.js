import { defineConfig, loadEnv } from "vite";
import { sentryVitePlugin } from "@sentry/vite-plugin";
// import { viteStaticCopy } from 'vite-plugin-static-copy'
import path from "path";
import tailwindcss from 'tailwindcss'
import autoprefixer from "autoprefixer";
import fs from 'fs'
import * as sass from 'sass'
export default defineConfig(({  mode }) => {
    const env = loadEnv(mode, process.cwd(), '')

    return {
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
                // input: Object.fromEntries(
                //     glob.sync('./src/themes/**/*.scss').map(file => [
                //         file.replace('./src/themes/', '').replace('.scss', ''),
                //         resolve(__dirname, file)
                //     ])
                // ),
                output: {
                    manualChunks: undefined,

                },

                // Ensure no chunk splitting

            },
        },
        css: {
            preprocessorOptions: {
                scss: {
                    sourcemap: true,
                    plugins: [tailwindcss, autoprefixer],
                    // If you have any SCSS options
                }
            }


        },
        emptyOutDir: false,


    plugins: [

        // viteStaticCopy({
        //     targets: [
        //         {
        //             src: path.resolve(__dirname, './src/themes') + '/[!.]*', // 1️⃣
        //             dest: './themes/', // 2️⃣
        //         },
        //     ],
        // }),

        {
            name: 'process-themes-folder',
            apply: 'build',

            async generateBundle() {
                const themesDir = path.resolve(__dirname, 'src/themes');
                const outputDir = path.resolve(__dirname, 'dist/themes');

                try {
                    const themeFiles = await fs.promises.readdir(themesDir); // Read theme files
                    await fs.promises.mkdir(outputDir, { recursive: true }); // Ensure output directory exists

                    for (const file of themeFiles) {
                        if (file.endsWith('.scss')) {
                            const inputPath = path.join(themesDir, file);
                            const outputPath = path.join(outputDir, file.replace('.scss', '.css'));
                            // Compile SCSS to CSS
                            const result = sass.compile(inputPath);
                            await fs.promises.writeFile(outputPath, result.css); // Write the compiled CSS
                        }
                    }
                } catch (err) {
                    console.error(`[process-themes-folder] Error: ${err.message}`);
                }
            },
        },

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
