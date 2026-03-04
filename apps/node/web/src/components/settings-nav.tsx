"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

const items = [
  { href: "/connections", label: "Connections" },
  { href: "/uploads", label: "Uploads" },
  { href: "/integrations", label: "Integrations" },
  { href: "/members", label: "Members" },
  { href: "/billing", label: "Billing" },
];

export function SettingsNav() {
  const pathname = usePathname();

  return (
    <nav className="border-b">
      <div className="container flex items-center gap-2 text-sm">
        {items.map((item) => {
          const active =
            pathname === item.href || pathname.startsWith(item.href + "/");
          return (
            <Link
              key={item.href}
              href={item.href}
              className={
                active
                  ? "border-b-2 border-foreground px-3 py-2 font-medium"
                  : "border-b-2 border-transparent px-3 py-2 text-muted-foreground hover:text-foreground"
              }
            >
              {item.label}
            </Link>
          );
        })}
      </div>
    </nav>
  );
}
