"use client";

import { useWebSocketStore } from "@/zusstand/web-socket/websocketStore";
import { useEffect } from "react";
import { decodeToken } from "@/lib/utils";
import { Role } from "@/constants/type";

interface WebSocketWrapperProps {
  accessToken: string;
}

export default function WebSocketWrapper({
  accessToken
}: WebSocketWrapperProps) {
  console.log("quananqr1/app/manage/component/WebSocketWrapper.tsx");
  const { connect } = useWebSocketStore();

  useEffect(() => {
    try {
      // Decode the access token
      const decoded = decodeToken(accessToken);

      // Use the decoded information
      const tableToken = "..."; // You might want to replace this with actual table token logic

      connect({
        userId: decoded.id.toString(),
        isGuest: false,
        userToken: accessToken,
        tableToken,
        role: decoded.role,
        email: decoded.email
      });

      // Optional: Disconnect on unmount
      return () => {
        useWebSocketStore.getState().disconnect();
      };
    } catch (error) {
      console.error("Failed to decode access token", error);
      // Optionally handle token decoding error (e.g., logout, redirect)
    }
  }, [accessToken, connect]);

  return null; // This component doesn't render anything
}
