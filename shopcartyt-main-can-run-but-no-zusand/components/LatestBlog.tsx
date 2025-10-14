import React from "react";
import Title from "./Title";

import Image from "next/image";

import Link from "next/link";
import { Calendar } from "lucide-react";
import dayjs from "dayjs";

const LatestBlog = async () => {

  return (
    <div className="mb-10 lg:mb-20">
      <Title>Latest Blog</Title>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-5 mt-5">

      </div>
    </div>
  );
};

export default LatestBlog;
