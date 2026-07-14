import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  images: {
    // Local public/ images don't need a remote pattern — this is already
    // handled by default. This config is here for future remote image domains.
    remotePatterns: [],
  },
};

export default nextConfig;
