import { DishInterface } from "@/schemaValidations/interface/type_dish";
import {
  SetInterface,
  SetProtoDish
} from "@/schemaValidations/interface/types_set";
import { create } from "zustand";
import { persist } from "zustand/middleware";
const INITIAL_STATE = {
  setItems: [],
  dishItems: [],
  tableNumber: null,
  tabletoken: null,
  canhKhongRau: 0,
  canhCoRau: 0,
  smallBowl: 0,
  wantChili: false,
  selectedFilling: {
    mocNhi: false,
    thit: false,
    thitMocNhi: false
  }
};

// Add new interfaces for bowl options
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

// Extend existing interfaces
export interface DishOrderItemustand extends DishInterface {
  quantity: number;
}

export interface SetOrderItemustand extends SetInterface {
  quantity: number;
}

interface FormattedSetItem {
  id: number | null;
  name: string;
  displayString: string;
  itemsString: string;
  totalPrice: number;
  formattedTotalPrice: string;
}

interface FormattedDishItem {
  id: number;
  name: string;
  displayString: string;
  totalPrice: number;
  formattedTotalPrice: string;
}

interface OrderState extends BowlOptions {
  dishItems: DishOrderItemustand[];
  setItems: SetOrderItemustand[];
  isLoading: boolean;
  error: string | null;
  tableNumber: number | null;
  tabletoken: string | null;
  clearDataExecuted: boolean; // New flag to track clear execution
  // Add new bowl-related actions
  updateCanhKhongRau: (count: number) => void;
  updateCanhCoRau: (count: number) => void;
  updateSmallBowl: (count: number) => void;
  updateWantChili: (value: boolean) => void;
  updateSelectedFilling: (type: "mocNhi" | "thit" | "thitMocNhi") => void;

  // Existing actions
  addTableToken: (tableToken: string) => void;
  addTableNumber: (tableNumber: number) => void;
  addDishItem: (dish: DishInterface, quantity: number) => void;
  removeDishItem: (id: number) => void;
  updateDishQuantity: (id: number, quantity: number) => void;
  addSetItem: (set: SetInterface, quantity: number) => void;
  removeSetItem: (id: number) => void;
  updateSetQuantity: (id: number, quantity: number) => void;
  updateSetDishes: (setId: number, modifiedDishes: SetProtoDish[]) => void;
  clearOrder: () => void;
  setLoading: (isLoading: boolean) => void;
  setError: (error: string | null) => void;
  getFormattedSets: () => FormattedSetItem[];
  getFormattedDishes: () => FormattedDishItem[];
  getFormattedTotals: () => {
    totalItems: number;
    formattedTotalPrice: string;
  };
  getOrderSummary: () => {
    totalItems: number;
    totalPrice: number;
    dishes: DishOrderItemustand[];
    sets: SetOrderItemustand[];
  };
  findDishOrderItem: (id: number) => DishOrderItemustand | undefined;
  findSetOrderItem: (id: number) => SetOrderItemustand | undefined;

  // clearAllOrderData: () => void;
  // resetClearDataFlag: () => void;
}

const useOrderStore = create<OrderState>()(
  persist(
    (set, get) => ({
      ...INITIAL_STATE,
      clearDataExecuted: false,
      // New bowl-related state
      canhKhongRau: 0,
      canhCoRau: 0,
      smallBowl: 0,
      wantChili: false,
      selectedFilling: {
        mocNhi: false,
        thit: false,
        thitMocNhi: false
      },

      // New bowl-related actions
      updateCanhKhongRau: (count) => set({ canhKhongRau: count }),
      updateCanhCoRau: (count) => set({ canhCoRau: count }),
      updateSmallBowl: (count) => set({ smallBowl: count }),
      updateWantChili: (value) => set({ wantChili: value }),
      updateSelectedFilling: (type) =>
        set((state) => ({
          selectedFilling: {
            mocNhi: type === "mocNhi",
            thit: type === "thit",
            thitMocNhi: type === "thitMocNhi"
          }
        })),

      // Existing state
      dishItems: [],
      setItems: [],
      isLoading: false,
      error: null,
      tableNumber: 0,
      tabletoken: "",

      // Existing actions
      addTableToken: (token) => set({ tabletoken: token }),
      addTableNumber: (tableNumber) => set({ tableNumber }),
      addDishItem: (dish, quantity) =>
        set((state) => {
          const existingItem = state.dishItems.find((i) => i.id === dish.id);
          if (existingItem) {
            return {
              dishItems: state.dishItems.map((i) =>
                i.id === dish.id ? { ...i, quantity: i.quantity + quantity } : i
              )
            };
          } else {
            const newItem: DishOrderItemustand = { ...dish, quantity };
            return { dishItems: [...state.dishItems, newItem] };
          }
        }),
      removeDishItem: (id) =>
        set((state) => ({
          dishItems: state.dishItems.filter((i) => i.id !== id)
        })),
      updateDishQuantity: (id, quantity) =>
        set((state) => ({
          dishItems: state.dishItems.map((i) =>
            i.id === id ? { ...i, quantity } : i
          )
        })),
      addSetItem: (setItem, quantity) =>
        set((state) => {
          const existingItem = state.setItems.find(
            (item) => item.id === setItem.id
          );
          if (existingItem) {
            return {
              setItems: state.setItems.map((item) =>
                item.id === setItem.id
                  ? { ...item, quantity: item.quantity + quantity }
                  : item
              )
            };
          } else {
            const newItem: SetOrderItemustand = { ...setItem, quantity };
            return { setItems: [...state.setItems, newItem] };
          }
        }),
      updateSetQuantity: (id, quantity) =>
        set((state) => {
          return {
            setItems: state.setItems.map((i) =>
              i.id === id
                ? { ...i, quantity: Math.max(0, quantity) } // Ensure quantity never goes below 0
                : i
            )
          };
        }),

      removeSetItem: (id) =>
        set((state) => {
          const updatedSetItems = state.setItems.filter((i) => i.id !== id);

          return {
            setItems: updatedSetItems
          };
        }),

      updateSetDishes: (setId, modifiedDishes) =>
        set((state) => {
          const updatedSetItems = state.setItems.map((i) =>
            i.id === setId ? { ...i, dishes: modifiedDishes } : i
          );

          return {
            setItems: updatedSetItems
          };
        }),
      clearOrder: () =>
        set({
          setItems: [],
          dishItems: [],
          tableNumber: null,
          tabletoken: null,
          canhKhongRau: 0,
          canhCoRau: 0,
          smallBowl: 0,
          wantChili: false,
          selectedFilling: {
            mocNhi: false,
            thit: false,
            thitMocNhi: false
          }
        }),
      setLoading: (isLoading) => set({ isLoading }),
      setError: (error) => set({ error }),
      getOrderSummary: () => {
        const { dishItems, setItems } = get();
        const totalItems =
          dishItems.reduce((acc, item) => acc + item.quantity, 0) +
          setItems.reduce((acc, item) => acc + item.quantity, 0);
        const dishesPrice = dishItems.reduce(
          (acc, item) => acc + item.price * item.quantity,
          0
        );
        const setsPrice = setItems.reduce((acc, item) => {
          const setPrice = calculateSetPrice(item.dishes);
          return acc + setPrice * item.quantity;
        }, 0);
        const totalPrice = dishesPrice + setsPrice;
        return { totalItems, totalPrice, dishes: dishItems, sets: setItems };
      },
      findDishOrderItem: (id) => get().dishItems.find((item) => item.id === id),
      findSetOrderItem: (id) => get().setItems.find((item) => item.id === id),
      getFormattedSets: () => {
        const { setItems } = get();
        return setItems.map((set) => {
          const basePrice = calculateSetPrice(set.dishes);
          const totalPrice = basePrice * set.quantity;
          const displayString = `${set.name} - ${formatCurrency(basePrice)} x ${
            set.quantity
          }`;
          const itemsString = set.dishes
            .map((dish) => `${dish.name} x ${dish.quantity}`)
            .join(", ");
          return {
            id: set.id,
            name: set.name,
            displayString,
            itemsString,
            totalPrice,
            formattedTotalPrice: formatCurrency(totalPrice)
          };
        });
      },
      getFormattedDishes: () => {
        const { dishItems } = get();
        return dishItems.map((dish) => {
          const totalPrice = dish.price * dish.quantity;
          return {
            id: dish.id,
            name: dish.name,
            displayString: `${dish.name} - ${formatCurrency(dish.price)} x ${
              dish.quantity
            }`,
            totalPrice,
            formattedTotalPrice: formatCurrency(totalPrice)
          };
        });
      },
      getFormattedTotals: () => {
        const { dishItems, setItems } = get();
        const totalItems =
          dishItems.reduce((acc, item) => acc + item.quantity, 0) +
          setItems.reduce((acc, item) => acc + item.quantity, 0);
        const totalPrice =
          dishItems.reduce((acc, item) => acc + item.price * item.quantity, 0) +
          setItems.reduce(
            (acc, item) => acc + calculateSetPrice(item.dishes) * item.quantity,
            0
          );
        return {
          totalItems,
          formattedTotalPrice: formatCurrency(totalPrice)
        };
      },
      // clearAllOrderData: () => {
      //   const { clearDataExecuted } = get();
      //   console.log(
      //     "quananqr1/zusstand/order/order_zustand.ts clearAllOrderData",
      //     clearDataExecuted
      //   );
      //   // Only execute clear if it hasn't been done before
      //   if (!clearDataExecuted) {
      //     set({
      //       ...INITIAL_STATE,
      //       clearDataExecuted: true // Set flag to prevent multiple clears
      //     });
      //   }
      // },
      // resetClearDataFlag: () => {
      //   console.log(
      //     "quananqr1/zusstand/order/order_zustand.ts resetClearDataFlag"
      //   );
      //   set({ clearDataExecuted: false });
      // }
    }),
    {
      name: "order-storage", // unique name for localStorage
      partialize: (state) => ({
        // Only persist these fields
        dishItems: state.dishItems,
        setItems: state.setItems,
        tableNumber: state.tableNumber,
        tabletoken: state.tabletoken,
        canhKhongRau: state.canhKhongRau,
        canhCoRau: state.canhCoRau,
        smallBowl: state.smallBowl,
        wantChili: state.wantChili,
        selectedFilling: state.selectedFilling
      })
    }
  )
);

function calculateSetPrice(dishes: SetProtoDish[]): number {
  if (!dishes || dishes.length === 0) return 0;
  return dishes.reduce((acc, dish) => acc + dish.price * dish.quantity, 0);
}

function formatCurrency(amount: number): string {
  return `${amount}`;
}

export default useOrderStore;
