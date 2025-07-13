// config/loggerConfig.ts

export interface LogPath {
  path: string;
  enabled: boolean;
  description: string;
  enabledLogIds: number[];
  logDescriptions: {
    [key: number]: {
      description: string;
      location: string;
      status: "enabled" | "disabled";
    };
  };
}

export const loggerPaths: LogPath[] = [
  {
    path: "quananqr1/app/manage/admin/orders/restaurant-summary/restaurant-summary.tsx",
    enabled: false,
    description: "Restaurant Summary Component Logs",
    enabledLogIds: [1, 2, 3],
    logDescriptions: {
      1: {
        description: "Log initial dish aggregation state",
        location: "aggregateDishes function - initialization",
        status: "enabled"
      },
      2: {
        description: "Log aggregated dishes for order groups",
        location: "RestaurantSummary component - groupedOrders processing",
        status: "enabled"
      },
      3: {
        description: "Log aggregation completion state",
        location: "RestaurantSummary component - final state",
        status: "enabled"
      }
    }
  },
  {
    path: "/manage/admin/table",
    enabled: false,
    description: "Admin Table Management Logs",
    enabledLogIds: [1, 2, 3],
    logDescriptions: {
      1: {
        description: "Table component initialization state",
        location: "Table component - mount phase",
        status: "enabled"
      },
      2: {
        description: "Table data state changes",
        location: "Table component - data updates",
        status: "enabled"
      },
      3: {
        description: "Table component error handling",
        location: "Table component - error boundaries",
        status: "enabled"
      }
    }
  },
  {
    path: "quananqr1/app/manage/admin/orders/restaurant-summary/dishes-summary.tsx",
    enabled: true,
    description: "Dishes Summary Component Logs",
    enabledLogIds: [4],
    // enabledLogIds: [1, 2, 3, 4, 5, 6, 7, 8],
    logDescriptions: {
      1: {
        description: "Delivery button click tracking",
        location: "handleDeliveryClick function",
        status: "enabled"
      },
      2: {
        description: "Keypad submission attempt",
        location: "handleKeypadSubmit function - initial validation",
        status: "enabled"
      },
      3: {
        description: "Delivery quantity validation error",
        location: "handleKeypadSubmit function - quantity validation",
        status: "enabled"
      },
      4: {
        description: "Delivery creation success",
        location: "handleKeypadSubmit function - success path",
        status: "enabled"
      },
      5: {
        description: "Delivery creation error",
        location: "handleKeypadSubmit function - error handling",
        status: "enabled"
      },
      6: {
        description: "Reference order validation",
        location: "handleKeypadSubmit function - order validation",
        status: "enabled"
      },
      7: {
        description: "Guest order processing",
        location: "handleKeypadSubmit function - guest handling",
        status: "enabled"
      },
      8: {
        description: "Delivery details compilation",
        location: "handleKeypadSubmit function - delivery preparation",
        status: "enabled"
      }
    }
  },
  {
    path: "quananqr1/zusstand/delivery/delivery_zustand.ts",
    enabled: true,
    description: "Delivery Store State Management Logs",
    enabledLogIds: [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12],
    logDescriptions: {
      1: {
        description: "Delivery creation - Initial data preparation",
        location: "createDelivery function - initialization",
        status: "enabled"
      },
      2: {
        description: "Delivery API request monitoring",
        location: "createDelivery function - API communication",
        status: "enabled"
      },
      3: {
        description: "Dish items state changes",
        location: "Dish manipulation functions",
        status: "enabled"
      },
      4: {
        description: "Price calculation events",
        location: "Price calculation functions",
        status: "enabled"
      },
      5: {
        description: "Delivery status change events",
        location: "Status management functions",
        status: "enabled"
      },
      6: {
        description: "State persistence events",
        location: "Persistence middleware",
        status: "enabled"
      },
      7: {
        description: "Delivery creation error handling",
        location: "Error handling middleware",
        status: "enabled"
      },
      8: {
        description: "General delivery state updates",
        location: "State update functions",
        status: "enabled"
      },
      9: {
        description: "Order summary state changes",
        location: "Order summary processing",
        status: "enabled"
      },
      10: {
        description: "Authentication state validation",
        location: "Auth validation middleware",
        status: "enabled"
      },
      11: {
        description: "Delivery data transformation",
        location: "createDelivery function - data transformation",
        status: "enabled"
      },
      12: {
        description: "Delivery request validation",
        location: "createDelivery function - request validation",
        status: "enabled"
      }
    }
  }
];

const isDevelopment = process.env.NODE_ENV !== "production";

type LogLevel = "debug" | "info" | "warn" | "error";

function validateLogId(pathConfig: LogPath, logId: number): void {
  if (!pathConfig.logDescriptions[logId]) {
    throw new Error(`
      Invalid log ID: ${logId}
      Available log IDs for ${pathConfig.path}:
      ${Object.keys(pathConfig.logDescriptions)
        .map(
          (id) =>
            `\n- Log #${id}: ${
              pathConfig.logDescriptions[Number(id)].description
            }`
        )
        .join("")}
    `);
  }
}

function validatePath(path: string): LogPath {
  const pathConfig = loggerPaths.find((p) => path.startsWith(p.path));
  if (!pathConfig) {
    throw new Error(`
      Invalid path: ${path}
      Available paths:
      ${loggerPaths.map((p) => `\n- ${p.path}`).join("")}
    `);
  }
  return pathConfig;
}

export function getLogStatus(path: string): void {
  const pathConfig = validatePath(path);
  console.log(`
Log Status for: ${pathConfig.path}
Description: ${pathConfig.description}
Enabled: ${pathConfig.enabled}

Available Logs:
${Object.entries(pathConfig.logDescriptions)
  .map(
    ([id, log]) => `
Log #${id}:
- Description: ${log.description}
- Location: ${log.location}
- Status: ${log.status}
- Enabled: ${pathConfig.enabledLogIds.includes(Number(id)) ? "Yes" : "No"}
`
  )
  .join("")}
  `);
}

export function logWithLevel(
  message: Record<string, unknown>,
  path: string,
  level: LogLevel,
  logId: number
): void {
  // Validate message
  if (!message || typeof message !== "object") {
    throw new Error(`
      message: Must be an object (e.g., { dishMap }, { users })
      Received: ${JSON.stringify(message)}
    `);
  }

  // Validate path and get config
  const pathConfig = validatePath(path);

  // Validate log ID
  validateLogId(pathConfig, logId);

  // Check if logging is enabled for this path and log ID
  if (!pathConfig.enabled || !pathConfig.enabledLogIds.includes(logId)) {
    return;
  }

  // Get log description
  const logInfo = pathConfig.logDescriptions[logId];

  // Format the log prefix
  const logIdPrefix = `[Log #${logId}]`;
  const locationPrefix = `[${logInfo.location}]`;

  // Log based on level
  switch (level) {
    case "debug":
      isDevelopment &&
        console.log(
          `🔍 DEBUG ${logIdPrefix} ${locationPrefix} [${path}]:`,
          message
        );
      break;
    case "info":
      isDevelopment &&
        console.log(
          `ℹ️ INFO ${logIdPrefix} ${locationPrefix} [${path}]:`,
          message
        );
      break;
    case "warn":
      console.log(
        `⚠️ WARN ${logIdPrefix} ${locationPrefix} [${path}]:`,
        message
      );
      break;
    case "error":
      console.log(
        `❌ ERROR ${logIdPrefix} ${locationPrefix} [${path}]:`,
        message
      );
      break;
  }
}
