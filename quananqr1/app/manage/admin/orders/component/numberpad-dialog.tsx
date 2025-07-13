"use client";
import React, { useState, useEffect } from "react";
import { Dialog, DialogContent, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

interface NumericKeypadInputProps {
  value: number;
  onChange: (value: number) => void;
  max: number;
  onSubmit: (value: number) => void;
  className?: string;
}

const NumericKeypadInput: React.FC<NumericKeypadInputProps> = ({
  value,
  onChange,
  max,
  onSubmit,
  className
}) => {
  const [isOpen, setIsOpen] = useState<boolean>(false);
  const [tempValue, setTempValue] = useState<string>("");
  const [currentValue, setCurrentValue] = useState<number>(value);

  // Initialize tempValue when opening the dialog
  useEffect(() => {
    if (isOpen) {
      setTempValue(currentValue.toString());
    }
  }, [isOpen, currentValue]);

  // Update currentValue when prop value changes
  useEffect(() => {
    setCurrentValue(value);
  }, [value]);

  const handleNumberClick = (num: number): void => {
    setTempValue((prev) => {
      if (prev === "0") {
        return num.toString();
      }
      const newVal = prev + num.toString();
      const numericValue = parseInt(newVal);
      return !isNaN(numericValue) && numericValue <= max ? newVal : prev;
    });
  };

  const handleBackspace = (): void => {
    setTempValue((prev) => {
      const newVal = prev.slice(0, -1);
      return newVal === "" ? "0" : newVal;
    });
  };

  const handleClear = (): void => {
    setTempValue("0");
  };

  const handleSubmit = (): void => {
    const numValue = parseInt(tempValue) || 0;
    if (numValue <= max) {
      setCurrentValue(numValue);
      onChange(numValue);
      onSubmit(numValue);
      setIsOpen(false);
    }
  };

  const handleCancel = (): void => {
    setTempValue(currentValue.toString());
    setIsOpen(false);
  };

  const handleDialogChange = (open: boolean) => {
    if (!open) {
      handleCancel();
    }
    setIsOpen(open);
  };

  return (
    <>
      <Input
        type="text"
        value={currentValue}
        readOnly
        onClick={() => setIsOpen(true)}
        className={`cursor-pointer ${className}`}
      />

      <Dialog open={isOpen} onOpenChange={handleDialogChange}>
        <DialogContent className="sm:max-w-[350px]">
          <DialogTitle>Enter Number</DialogTitle>
          <div className="space-y-4">
            <Input
              type="text"
              value={tempValue}
              readOnly
              className="text-right text-xl h-12"
            />

            <div className="grid grid-cols-3 gap-2">
              {[1, 2, 3, 4, 5, 6, 7, 8, 9].map((num) => (
                <Button
                  key={num}
                  onClick={() => handleNumberClick(num)}
                  variant="outline"
                  className="h-12 text-xl"
                >
                  {num}
                </Button>
              ))}
              <Button onClick={handleClear} variant="outline" className="h-12">
                C
              </Button>
              <Button
                onClick={() => handleNumberClick(0)}
                variant="outline"
                className="h-12 text-xl"
              >
                0
              </Button>
              <Button
                onClick={handleBackspace}
                variant="outline"
                className="h-12"
              >
                ←
              </Button>
            </div>

            <div className="flex gap-2">
              <Button
                onClick={handleCancel}
                variant="outline"
                className="flex-1"
              >
                Cancel
              </Button>
              <Button onClick={handleSubmit} className="flex-1">
                Enter
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
};

export default NumericKeypadInput;
