"use client";

import React, { useState } from "react";
import { ChevronDown, ChevronRight } from "lucide-react";

interface CollapsibleSectionProps {
  title: string;
  children: React.ReactNode;
}

export const CollapsibleSection: React.FC<CollapsibleSectionProps> = ({
  title,
  children
}) => {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <div className="mt-2">
      <div
        className="flex items-center cursor-pointer select-none p-2 rounded"
        onClick={() => setIsOpen(!isOpen)}
      >
        <h3 className="text-md font-semibold">{title}</h3>
        {isOpen ? (
          <ChevronDown className="ml-2 h-4 w-4" />
        ) : (
          <ChevronRight className="ml-2 h-4 w-4" />
        )}
      </div>
      {isOpen && <div className="p-2 border-l ">{children}</div>}
    </div>
  );
};

export default CollapsibleSection;
