import React from "react";

interface BowlOptions {
  canhKhongRau: number;
  canhCoRau: number;
  smallBowl: number;
  wantChili: boolean;
  selectedFilling: {
    mocNhi: boolean;
    thit: boolean;
    thitMocNhi: boolean;
  };
}

interface ToppingSummaryProps {
  set: BowlOptions;
}

const ToppingSummary: React.FC<ToppingSummaryProps> = ({ set }) => {
  const { canhKhongRau, canhCoRau, smallBowl, selectedFilling, wantChili } =
    set;

  const toppingTotal = `Canh không rau: ${canhKhongRau} - Canh rau: ${canhCoRau} - Bát bé: ${smallBowl} - Nhân mọc nhĩ: ${
    selectedFilling.mocNhi ? "true" : "false"
  } - Nhân thịt: ${
    selectedFilling.thit ? "true" : "false"
  } - Nhân thịt và mọc nhĩ: ${
    selectedFilling.thitMocNhi ? "true" : "false"
  } - Có ớt: ${wantChili ? "true" : "false"}`;

  const parseOrderString = (str: string) => {
    const items = str.split(" - ");
    const orderData: { [key: string]: number | boolean } = {};

    items.forEach((item) => {
      const [key, value] = item.split(": ");
      if (value === "true" || value === "false") {
        orderData[key] = value === "true";
      } else {
        orderData[key] = parseInt(value) || 0;
      }
    });

    return orderData;
  };

  const orderData = parseOrderString(toppingTotal);
  //   const totalOrders = orderData['Canh không rau'] as number + orderData['Canh rau'] as number;

  return (
    <div className="mt-4 p-4  shadow-lg rounded-lg">
      <div className="space-y-2">
        <h2 className="text-xl font-bold text-gray-800">Order Summary</h2>
        <div className="flex items-center space-x-2">
          <span className="text-gray-700 font-medium">Total Orders:</span>
          {/* <span className="text-2xl font-bold text-blue-600">{totalOrders}</span> */}
        </div>

        <div className="mt-4 space-y-2">
          <div className="grid grid-cols-2 gap-2">
            {Object.entries(orderData).map(([key, value]) => (
              <div key={key} className="p-2 rounded">
                <span className="font-medium">{key}:</span>{" "}
                <span className="text-blue-600">
                  {typeof value === "boolean" ? (value ? "Yes" : "No") : value}
                </span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
};

export default ToppingSummary;
