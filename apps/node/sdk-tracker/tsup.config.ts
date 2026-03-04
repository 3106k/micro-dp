import { defineConfig } from "tsup";

export default defineConfig([
  {
    entry: ["src/index.ts"],
    format: ["esm", "cjs"],
    dts: true,
    sourcemap: true,
    clean: true,
    target: "es2020"
  },
  {
    entry: { "micro-dp": "src/cdn.ts" },
    format: ["iife"],
    globalName: "MicroDP",
    sourcemap: true,
    target: "es2020",
    minify: true
  }
]);
