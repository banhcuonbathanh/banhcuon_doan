"use client";

import { Heading } from "@/components/ui/heading";
import { Separator } from "@/components/ui/separator";
import {
  OrderDetailedResponse,
  PaginationInfo
} from "@/schemaValidations/interface/type_order";

import { useEffect, useState, useCallback } from "react";
import { get_Orders } from "@/zusstand/server/order-controller";
import { Button } from "@/components/ui/button";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";

import { useWebSocketStore } from "@/zusstand/web-socket/websocketStore";
import { WebSocketMessage } from "@/schemaValidations/interface/type_websocker";
import { APP_CONSTANTS } from "@/config";


import RestaurantSummary2 from "../restaurant-summary/restaurant-summary/restaurant-summary";

interface OrderClientProps {
  initialData: OrderDetailedResponse[];
  initialPagination: PaginationInfo;
}

export const OrderClient: React.FC<OrderClientProps> = ({
  initialData,
  initialPagination
}) => {
  // console.log("quananqr1/app/manage/admin/orders/component/order-client.tsx ");
  const [currentPage, setCurrentPage] = useState(
    initialPagination.current_page
  );
  const [data, setData] = useState(initialData);
  const [pagination, setPagination] = useState(initialPagination);
  const [isLoading, setIsLoading] = useState(false);

  const { addMessageHandler } = useWebSocketStore();

  const handlePageChange = async (newPage: number) => {
    console.log("OrderClient: handlePageChange triggered", newPage);
    setIsLoading(true);
    try {
      const orders = await get_Orders({
        page: newPage,
        page_size: pagination.page_size
      });

      setData(orders.data);
      setPagination(orders.pagination);
      setCurrentPage(newPage);
    } catch (error) {
      console.error("Error fetching orders:", error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleWebSocketMessage = useCallback((message: WebSocketMessage) => {
    console.log("OrderClient: WebSocket message received:", message);

    // Check if the message is a new order
    if (message.type === "order" && message.action === "new_order") {
      console.log("OrderClient: New order received, refreshing page");
      // Refresh the first page when a new order is received
      handlePageChange(1);
    }
  }, []);

  useEffect(() => {
    // Add message handler specifically for new orders
    const removeHandler = addMessageHandler(handleWebSocketMessage);

    // Cleanup function
    return () => {
      removeHandler(); // Remove the message handler
    };
  }, [addMessageHandler, handleWebSocketMessage]);

  // auto fetch -------------------------
  const [isVisible, setIsVisible] = useState(true);

  useEffect(() => {
    // Function to check if the page is visible
    const handleVisibilityChange = () => {
      setIsVisible(!document.hidden);
    };

    // Add visibility change listener
    document.addEventListener("visibilitychange", handleVisibilityChange);

    // Function to fetch orders
    const fetchOrders = async () => {
      // Only fetch if the page is visible
      if (!document.hidden) {
        try {
          const response = await get_Orders({
            page: 1,
            page_size: 10
          });

          setData(response.data);
          setPagination(response.pagination);
        } catch (error) {
          console.error("Error fetching orders:", error);
        }
      }
    };

    // Only set up interval if the component is visible
    let intervalId: NodeJS.Timeout | null = null;
    if (isVisible) {
      intervalId = setInterval(
        fetchOrders,
        APP_CONSTANTS.Intervel_revalidata_Page_Order
      );
    }

    // Cleanup function
    return () => {
      document.removeEventListener("visibilitychange", handleVisibilityChange);
      if (intervalId) {
        clearInterval(intervalId);
      }
    };
  }, [isVisible]);

  //
  return (
    <div className="grid flex-1 items-start gap-4 p-4 sm:px-6 sm:py-0 md:gap-8">
      <div className="space-y-2">
        <Card>
          <CardHeader>
            <CardTitle>dat ban</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex flex-col">
              <div className="flex flex-row justify-between items-center">
                <Heading
                  title={`so luong don : ${pagination.total_items}`}
                  description={`Page ${currentPage} of ${pagination.total_pages}`}
                />
              </div>

              <Separator className="my-4" />

              <RestaurantSummary2 restaurantLayoutProps={data} />
              {/* <RestaurantSummary restaurantLayoutProps={data} /> */}
              {/* <RestaurantLayout restaurantLayoutProps={data} /> */}
              {/* <YourComponent1 initialData={initialData} /> */}

              <div className="flex items-center justify-between space-x-2 py-4">
                <div className="flex-1 text-sm text-muted-foreground">
                  Showing {data.length} of {pagination.total_items} orders
                </div>
                <div className="flex space-x-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => handlePageChange(currentPage - 1)}
                    disabled={currentPage === 1 || isLoading}
                  >
                    Previous
                  </Button>
                  <div className="flex items-center space-x-2">
                    {[...Array(Math.min(5, pagination.total_pages))].map(
                      (_, idx) => {
                        const pageNum = idx + 1;
                        return (
                          <Button
                            key={pageNum}
                            variant={
                              pageNum === currentPage ? "default" : "outline"
                            }
                            size="sm"
                            onClick={() => handlePageChange(pageNum)}
                            disabled={isLoading}
                          >
                            {pageNum}
                          </Button>
                        );
                      }
                    )}
                  </div>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => handlePageChange(currentPage + 1)}
                    disabled={
                      currentPage === pagination.total_pages || isLoading
                    }
                  >
                    Next
                  </Button>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
};
