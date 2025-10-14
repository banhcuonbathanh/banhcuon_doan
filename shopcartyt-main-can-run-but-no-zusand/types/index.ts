// Category type
export type Category = {
  _id: string;
  title: string;
  slug: {
    current: string;
  };
  image?: string;
  productCount: number;
};

// Product type
export type Product = {
  _id: string;
  name: string;
  slug: {
    current: string;
  };
  images?: string[];
  price: number;
  discount?: number;
  stock: number;
  status?: "sale" | "normal";
  categories?: string[];
  description?: string;
  variant?: string; // featured, bestseller, latest, etc.
};