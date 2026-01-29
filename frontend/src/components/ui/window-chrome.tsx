import React from "react";
import { cn } from "@/lib/utils";

export const WindowChrome = ({
  children,
  title,
  className,
}: {
  children: React.ReactNode;
  title?: string;
  className?: string;
}) => (
  <div
    className={cn(
      "bg-card border border-border rounded-md shadow-sm overflow-hidden",
      "hover:shadow-md transition-shadow duration-200",
      className,
    )}
  >
    {title && (
      <div className="bg-gradient-to-b from-secondary/80 to-secondary/40 border-b border-border px-3 py-1.5 flex items-center gap-2">
        <div className="flex gap-1.5">
          <div className="w-3 h-3 rounded-full bg-destructive/90" />
          <div className="w-3 h-3 rounded-full bg-warning/90" />
          <div className="w-3 h-3 rounded-full bg-success/90" />
        </div>
        <span className="flex-1 text-center text-xs font-chicago text-muted-foreground uppercase tracking-wider mr-8">
          {title}
        </span>
      </div>
    )}
    <div className="p-4 sm:p-6">{children}</div>
  </div>
);
