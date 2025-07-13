import React, { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

interface NumericKeypadProps {
  value: number;
  onChange: (value: number) => void;
  onSubmit: () => void;
  min?: number;
  max?: number;
  className?: string;
  disabled?: boolean;
}

const NumericKeypad: React.FC<NumericKeypadProps> = ({
  value,
  onChange,
  onSubmit,
  min = 0,
  max = 999,
  className = "",
  disabled = false
}) => {
  const [inputValue, setInputValue] = useState(String(value));

  const handleNumberClick = (num: number) => {
    if (disabled) return;
    // Only replace the value if it's "0", otherwise concatenate
    const newValue = Number(inputValue) === 0 ? String(num) : inputValue + num;
    const numericValue = Number(newValue);
    
    if (numericValue <= max) {
      setInputValue(newValue);
      onChange(numericValue);
    }
  };

  const handleBackspace = () => {
    if (disabled) return;
    const newValue = inputValue.slice(0, -1) || "0";
    setInputValue(newValue);
    onChange(Number(newValue));
  };

  const handleClear = () => {
    if (disabled) return;
    setInputValue("0");
    onChange(0);
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (disabled) return;
    const newValue = e.target.value.replace(/[^0-9]/g, "");
    if (newValue === "") {
      setInputValue("0");
      onChange(0);
      return;
    }
    const numValue = Number(newValue);
    if (numValue >= min && numValue <= max) {
      setInputValue(newValue);
      onChange(numValue);
    }
  };

  return (
    <div className={`w-full max-w-xs mx-auto space-y-4 ${className} bg-white p-4 rounded-lg shadow-lg border border-gray-200`}>
      {/* Display */}
      <div className="relative">
        <Input
          type="text"
          value={inputValue}
          onChange={handleInputChange}
          className="text-2xl text-center h-14 bg-white border-2 border-gray-300"
          disabled={disabled}
        />
        <div className="absolute right-2 top-1/2 -translate-y-1/2 text-sm text-gray-600 font-medium">
          max: {max}
        </div>
      </div>

      {/* Keypad */}
      <div className="grid grid-cols-3 gap-2">
        {/* Numbers 1-9 */}
        {[1, 2, 3, 4, 5, 6, 7, 8, 9].map((num) => (
          <Button
            key={num}
            variant="outline"
            onClick={() => handleNumberClick(num)}
            className="h-14 text-xl bg-white hover:bg-gray-100 border-2"
            disabled={disabled}
          >
            {num}
          </Button>
        ))}

        {/* Clear, 0, Backspace */}
        <Button
          variant="outline"
          onClick={handleClear}
          className="h-14 text-sm bg-white hover:bg-gray-100 border-2"
          disabled={disabled}
        >
          Clear
        </Button>
        <Button
          variant="outline"
          onClick={() => handleNumberClick(0)}
          className="h-14 text-xl bg-white hover:bg-gray-100 border-2"
          disabled={disabled}
        >
          0
        </Button>
        <Button
          variant="outline"
          onClick={handleBackspace}
          className="h-14 bg-white hover:bg-gray-100 border-2"
          disabled={disabled}
        >
          ←
        </Button>

        {/* Submit button - spans full width */}
        <Button
          onClick={onSubmit}
          className="h-14 col-span-3 text-lg bg-blue-600 hover:bg-blue-700 text-white"
          disabled={disabled || Number(inputValue) < min || Number(inputValue) > max}
        >
          Submit
        </Button>
      </div>

      {/* Min-Max indicator */}
      <div className="text-center text-sm text-gray-700 font-medium">
        Enter a number between {min} and {max}
      </div>
    </div>
  );
};

export default NumericKeypad;