import type { NextConfig } from "next";

// The browser talks to Next on the same origin; Next forwards /api/* to the Go
// server. That sidesteps CORS entirely — no browser preflight, no API changes.
const apiOrigin = process.env.API_ORIGIN ?? "http://localhost:8080";

const nextConfig: NextConfig = {
  async rewrites() {
    return [{ source: "/api/:path*", destination: `${apiOrigin}/api/:path*` }];
  },
};

export default nextConfig;
