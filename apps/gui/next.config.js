/** @type {import('next').NextConfig} */
const nextConfig = {
  transpilePackages: ["@hookie/ui"],
  experimental: {
    optimizePackageImports: ["@hookie/ui"],
  },
};

export default nextConfig;
