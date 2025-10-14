"use client";

import React, { useEffect, useState } from "react";
import ProductCard from "./ProductCard";
import { motion, AnimatePresence } from "motion/react";
import NoProductAvailable from "./NoProductAvailable";
import { Loader2 } from "lucide-react";
import Container from "./Container";
import HomeTabbar from "./HomeTabbar";
import { productType } from "@/constants/data";
import { Product } from "@/types";

// Sample products data - replace with your actual data source
const allProducts: Product[] = [
  {
    _id: "1",
    name: "Wireless Headphones",
    slug: { current: "wireless-headphones" },
    images: ["https://placehold.co/500x500/png?text=Headphones"],
    price: 79.99,
    discount: 10,
    stock: 25,
    status: "sale",
    categories: ["Electronics", "Audio"],
    description: "High-quality wireless headphones",
    variant: "featured"
  },
  {
    _id: "2",
    name: "Smart Watch",
    slug: { current: "smart-watch" },
    images: ["https://placehold.co/500x500/png?text=Watch"],
    price: 199.99,
    stock: 15,
    status: "normal",
    categories: ["Electronics", "Wearables"],
    description: "Feature-packed smartwatch",
    variant: "bestseller"
  },
  {
    _id: "3",
    name: "Running Shoes",
    slug: { current: "running-shoes" },
    images: ["https://placehold.co/500x500/png?text=Shoes"],
    price: 89.99,
    discount: 15,
    stock: 50,
    status: "sale",
    categories: ["Sports", "Footwear"],
    description: "Comfortable running shoes",
    variant: "featured"
  },
  {
    _id: "4",
    name: "Coffee Maker",
    slug: { current: "coffee-maker" },
    images: ["https://placehold.co/500x500/png?text=Coffee"],
    price: 129.99,
    stock: 8,
    status: "normal",
    categories: ["Home", "Kitchen"],
    description: "Programmable coffee maker",
    variant: "latest"
  },
  {
    _id: "5",
    name: "Yoga Mat",
    slug: { current: "yoga-mat" },
    images: ["https://placehold.co/500x500/png?text=YogaMat"],
    price: 29.99,
    stock: 100,
    status: "normal",
    categories: ["Sports", "Fitness"],
    description: "Non-slip yoga mat",
    variant: "bestseller"
  },
  {
    _id: "6",
    name: "Laptop Stand",
    slug: { current: "laptop-stand" },
    images: ["https://placehold.co/500x500/png?text=LaptopStand"],
    price: 39.99,
    discount: 5,
    stock: 30,
    status: "sale",
    categories: ["Electronics", "Accessories"],
    description: "Adjustable laptop stand",
    variant: "latest"
  }
];

const ProductGrid = () => {
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedTab, setSelectedTab] = useState(productType[0]?.title || "");

  useEffect(() => {
    const fetchData = async () => {
      setLoading(true);
      try {
        // Simulate API delay
        await new Promise(resolve => setTimeout(resolve, 500));
        
        // Filter products by variant
        const variant = selectedTab.toLowerCase();
        const filteredProducts = allProducts.filter(
          product => product.variant === variant
        ).sort((a, b) => a.name.localeCompare(b.name));
        
        setProducts(filteredProducts);
      } catch (error) {
        console.log("Product fetching Error", error);
      } finally {
        setLoading(false);
      }
    };
    fetchData();
  }, [selectedTab]);

  return (
    <Container className="flex flex-col lg:px-0 my-10">
      <HomeTabbar selectedTab={selectedTab} onTabSelect={setSelectedTab} />
      {loading ? (
        <div className="flex flex-col items-center justify-center py-10 min-h-80 space-y-4 text-center bg-gray-100 rounded-lg w-full mt-10">
          <motion.div className="flex items-center space-x-2 text-blue-600">
            <Loader2 className="w-5 h-5 animate-spin" />
            <span>Product is loading...</span>
          </motion.div>
        </div>
      ) : products?.length ? (
        <div className="grid grid-cols-2 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-2.5 mt-10">
          <>
            {products?.map((product) => (
              <AnimatePresence key={product?._id}>
                <motion.div
                  layout
                  initial={{ opacity: 0.2 }}
                  animate={{ opacity: 1 }}
                  exit={{ opacity: 0 }}
                >
                  <ProductCard key={product?._id} product={product} />
                </motion.div>
              </AnimatePresence>
            ))}
          </>
        </div>
      ) : (
        <NoProductAvailable selectedTab={selectedTab} />
      )}
    </Container>
  );
};

export default ProductGrid;