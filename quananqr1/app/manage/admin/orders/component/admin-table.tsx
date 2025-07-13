import { OrderDetailedResponse, OrderTable } from "./new-order-column";

interface AdmiCnablelientProps {
  initialData: OrderDetailedResponse[];
}

// export default async function YourComponent() {
export const YourComponent1: React.FC<AdmiCnablelientProps> = ({
  initialData
}) => {
  return <OrderTable data={initialData} />;
};
