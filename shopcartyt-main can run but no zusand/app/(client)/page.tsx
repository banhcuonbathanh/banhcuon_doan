import Container from "@/components/Container";
import HomeBanner from "@/components/HomeBanner";
import HomeCategories from "@/components/HomeCategories";
import LatestBlog from "@/components/LatestBlog";
import ProductGrid from "@/components/ProductGrid";
import ShopByBrands from "@/components/ShopByBrands";


import React from "react";

const Home = async () => {
const categories = [
  {
    _id: "1",
    title: "Electronics",
    slug: { current: "electronics" },
    image: "/images/electronics.jpg",
    productCount: 45
  },
  {
    _id: "2",
    title: "Clothing",
    slug: { current: "clothing" },
    image: "/images/clothing.jpg",
    productCount: 120
  },
  {
    _id: "3",
    title: "Books",
    slug: { current: "books" },
    image: "/images/books.jpg",
    productCount: 89
  },
  {
    _id: "4",
    title: "Home & Garden",
    slug: { current: "home-garden" },
    image: "/images/home.jpg",
    productCount: 67
  },
  {
    _id: "5",
    title: "Sports",
    slug: { current: "sports" },
    image: "/images/sports.jpg",
    productCount: 34
  },
  {
    _id: "6",
    title: "Toys",
    slug: { current: "toys" },
    image: "/images/toys.jpg",
    productCount: 52
  }
];

  return (
    <Container className="bg-shop-light-pink">
      <HomeBanner />
      <ProductGrid />
      <HomeCategories categories={categories} />
      <ShopByBrands />
      <LatestBlog />
    </Container>
  );
};

export default Home;
