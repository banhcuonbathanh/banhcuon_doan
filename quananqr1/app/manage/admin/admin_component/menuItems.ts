import {
  Home,
  LineChart,
  ShoppingCart,
  Users2,
  Salad,
  Table,
  Group
} from "lucide-react";

const menuItems = [
  {
    title: "Dashboard",
    Icon: Home,
    href: "/"
  },
  {
    title: "Đơn hàng",
    Icon: ShoppingCart,
    href: "manage/admin/order"
  },
  {
    title: "Bàn ăn",
    Icon: Table,
    href: "manage/admin/table"
  },
  {
    title: "Món ăn",
    Icon: Salad,
    href: "admin/dish"
  },
  {
    title: "Set",
    Icon: Group,
    href: "admin/set"
  },

  {
    title: "Order",
    Icon: Group,
    href: "admin/orders"
  },
  {
    title: "Phân tích",
    Icon: LineChart,
    href: "admin/analytics"
  },
  {
    title: "Nhân viên",
    Icon: Users2,
    href: "admin/accounts"
  }
];

export default menuItems;
