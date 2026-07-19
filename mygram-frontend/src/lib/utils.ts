import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function ownerId(record: { UserID?: number; user_id?: number }) {
  return record.user_id ?? record.UserID ?? 0;
}
