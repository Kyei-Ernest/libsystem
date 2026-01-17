import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  basePath: '/libsystem',
  images: {
    remotePatterns: [
      {
        protocol: 'http',
        hostname: 'localhost',
        port: '8088',
        pathname: '/api/v1/documents/**',
      },
      {
        protocol: 'http',
        hostname: '127.0.0.1',
        port: '8088',
        pathname: '/api/v1/documents/**',
      },
    ],
  },
};

export default nextConfig;
