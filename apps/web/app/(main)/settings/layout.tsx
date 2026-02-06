"use client";

import { Button } from "@hookie/ui/components/button";
import { Separator } from "@hookie/ui/components/separator";
import Link from "next/link";
import { usePathname } from "next/navigation";

interface SettingsLayoutProps {
  children: React.ReactNode;
}

interface NavCategory {
  label: string;
  items: Array<{
    label: string;
    href: string;
  }>;
}

export default function SettingsLayout({ children }: SettingsLayoutProps) {
  const pathname = usePathname();

  const navCategories: NavCategory[] = [
    {
      label: "Account",
      items: [
        {
          label: "Usage",
          href: "/settings/usage",
        },
        {
          label: "Connected Clients",
          href: "/settings/connected-clients",
        },
      ],
    },
    {
      label: "Billing & Subscription",
      items: [
        {
          label: "Billing",
          href: "/settings/billing",
        },
      ],
    },
  ];

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex flex-col md:flex-row gap-8">
        <aside className="w-full md:w-64 shrink-0">
          <nav className="space-y-6">
            {navCategories.map((category) => (
              <div key={category.label} className="space-y-2">
                <h3 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider px-2">
                  {category.label}
                </h3>
                <div className="space-y-1">
                  {category.items.map((item) => {
                    const isActive = pathname === item.href;
                    return (
                      <Link key={item.href} href={item.href} className="block">
                        <Button
                          variant={isActive ? "default" : "ghost"}
                          className="w-full justify-start"
                        >
                          {item.label}
                        </Button>
                      </Link>
                    );
                  })}
                </div>
              </div>
            ))}
          </nav>
        </aside>

        <Separator orientation="vertical" className="hidden md:block" />

        <main className="flex-1 min-w-0">{children}</main>
      </div>
    </div>
  );
}
