import { createMDX } from "fumadocs-mdx/next";

const withMDX = createMDX();

/** @type {import('next').NextConfig} */
const config = {
  reactStrictMode: true,
  // redirects: async () => {
  //   return [
  //     {
  //       source: "/",
  //       destination: "/platform",
  //       permanent: true,
  //     },
  //   ];
  // },
  async rewrites() {
    return [
      {
        source: "/:path*.mdx",
        destination: "/llms.mdx/:path*",
      },
    ];
  },
};

export default withMDX(config);
