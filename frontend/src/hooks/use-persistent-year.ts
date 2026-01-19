import { useEffect } from "react";
import { useRouter, usePathname, useSearchParams } from "next/navigation";

export function usePersistentYear(key: string) {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();

  useEffect(() => {
    // Only run on client side
    if (typeof window === "undefined") return;

    const yearParam = searchParams.get("ano");
    // Prefix key to avoid collisions
    const storageKey = `todeolho_year_${key}`;
    const storedYear = localStorage.getItem(storageKey);

    if (yearParam !== null) {
      // URL has priority. If it exists, update storage to match.
      if (yearParam !== storedYear) {
        localStorage.setItem(storageKey, yearParam);
      }
    } else {
      // URL has no year. Check if we have a stored preference.
      if (storedYear !== null) {
        // Restore from storage
        const params = new URLSearchParams(searchParams.toString());
        params.set("ano", storedYear);
        router.replace(`${pathname}?${params.toString()}`);
      }
    }
  }, [searchParams, key, pathname, router]);
}
