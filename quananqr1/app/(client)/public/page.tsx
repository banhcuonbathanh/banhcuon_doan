import dishApiRequest from "@/apiRequests/dish";
import { formatCurrency } from "@/lib/utils";

import { StaticImport } from "next/dist/shared/lib/get-img-props";
import Image from "next/image";
import {
  Key,
  ReactElement,
  JSXElementConstructor,
  ReactNode,
  AwaitedReactNode,
  ReactPortal
} from "react";



export default async function Home() {
  // let dishList: DishListResType["data"] = [];

  // try {
  //   dishList = await dishController.listDishes();  // Use the imported instance
  // } catch (error) {
  //   console.error('Error fetching dishes:', error);
  //   return <div>Something went wrong</div>;
  // }
  // PublicDialog
  return (
    <div className="w-full space-y-4">
      <section className="relative z-10">
        <span className="absolute top-0 left-0 w-full h-full bg-black opacity-50 z-10"></span>
        <Image
          src="http://localhost:8888/uploads/folder1/folder2/Screenshot%202024-02-20%20at%2014.37.22.png"
          width={400}
          height={200}
          quality={100}
          alt="Banner"
          className="absolute top-0 left-0 w-full h-full object-cover"
        />
        <div className="z-20 relative py-10 md:py-20 px-4 sm:px-10 md:px-20">
          <h1 className="text-center text-xl sm:text-2xl md:text-4xl lg:text-5xl font-bold">
            Nhà hàng Big Boy
          </h1>
          <p className="text-center text-sm sm:text-base mt-4">
            Vị ngon, trọn khoảnh khắc
          </p>
        </div>
      </section>
      <section className="space-y-10 py-16">
        <h2 className="text-center text-2xl font-bold">Đa dạng các món ăn</h2>
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-10">
          {/* {dishList.map(
            (dish: {
              id: Key | null | undefined;
              image: string | StaticImport;
              name:
                | string
                | number
                | bigint
                | boolean
                | ReactElement<any, string | JSXElementConstructor<any>>
                | Iterable<ReactNode>
                | Promise<AwaitedReactNode>
                | null
                | undefined;
              description:
                | string
                | number
                | bigint
                | boolean
                | ReactElement<any, string | JSXElementConstructor<any>>
                | Iterable<ReactNode>
                | ReactPortal
                | Promise<AwaitedReactNode>
                | null
                | undefined;
              price: number;
            }) => (
              <div className="flex gap-4 w" key={dish.id}>
                <div className="flex-shrink-0">
                  <Image
                    src={dish.image}
                    width={150}
                    height={150}
                    quality={100}
                    alt={""}
                    className="object-cover w-[150px] h-[150px] rounded-md"
                  />
                </div>
                <div className="space-y-1">
                  <h3 className="text-xl font-semibold">{dish.name}</h3>
                  <p className="">{dish.description}</p>
                  <p className="font-semibold">{formatCurrency(dish.price)}</p>
                </div>
              </div>
            )
          )} */}
        </div>
      </section>
    </div>
  );
}
