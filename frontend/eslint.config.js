import globals from "globals";
import pluginJs from "@eslint/js";
import tseslint from "typescript-eslint";
import litPlugin from "eslint-plugin-lit"; // Import Lit plugin

/** @type {import('eslint').Linter.FlatConfig[]} */
export default [
  {
    files: ["**/*.{js,mjs,cjs,ts}"],
    languageOptions: {
      globals: globals.browser, // Include browser globals
    },
  },
  pluginJs.configs.recommended, // JavaScript recommended rules
  ...tseslint.configs.recommended, // TypeScript recommended rules
  {
    files: ["**/*.ts"], // Apply Lit rules only to TypeScript files (adjust as needed)
    plugins: {
      lit: litPlugin, // Register the Lit plugin
    },
    rules: {
      ...litPlugin.configs.recommended.rules, // Include Lit recommended rules
      // Add custom Lit rules or overrides here, if needed
    },
  },
];
