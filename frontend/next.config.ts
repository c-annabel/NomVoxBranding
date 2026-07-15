import type { NextConfig } from "next";
import path from "path";

const nextConfig: NextConfig = {
  images: {
    remotePatterns: [],
    // Allow quality=90 (used by the bg image)
    qualities: [75, 90],
  },
  turbopack: {
    // Pin the workspace root so Turbopack doesn't pick up the wrong lockfile
    root: path.resolve(__dirname),
  },
};

export default nextConfig;
