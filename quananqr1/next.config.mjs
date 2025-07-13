// /** @type {import('next').NextConfig} */
// const nextConfig = {};

// export default nextConfig;

/** @type {import('next').NextConfig} */
const nextConfig = {
  images: {
    remotePatterns: [
      { hostname: "images.pexels.com" },
      { hostname: "images.unsplash.com" },
      { hostname: "unsplash.com" },
      { hostname: "example.com" },
      { hostname: "127.0.0.1" },
      { hostname: "localhost" },
      { hostname: "shop-golang" }
    ]
  }
};

export default nextConfig;
