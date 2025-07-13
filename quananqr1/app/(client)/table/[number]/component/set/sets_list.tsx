"use client";

import { SetInterface } from "@/schemaValidations/interface/types_set";
import SetCard from "./set";
import React from "react";

interface SetCardListProps {
  sets: SetInterface[];
}

export default function SetCardList({ sets }: SetCardListProps) {
  const [isMounted, setIsMounted] = React.useState(false);

  React.useEffect(() => {
    setIsMounted(true);
  }, []);

  if (!isMounted) {
    return null; // or a loading skeleton
  }

  return (
    <div className="space-y-4">
      {sets.map((set) => (
        <SetCard key={set.id} set={set} />
      ))}
    </div>
  );
}
