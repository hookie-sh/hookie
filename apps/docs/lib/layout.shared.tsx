import { LogoWordmark } from "@hookie/ui/components/logo-wordmark";

import type { BaseLayoutProps } from "fumadocs-ui/layouts/shared";

export function baseOptions(): BaseLayoutProps {
  return {
    nav: {
      title: <LogoWordmark className="h-6 text-foreground" />,
    },
  };
}
