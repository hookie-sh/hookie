import { Theme } from "@clerk/types";

function clerkElements(overrides: Partial<Record<string, string>> = {}) {
  return Object.assign(
    {},
    {
      card: "bg-background",
      navbarButton:
        'data-[color="primary"]:text-primary data-[color="neutral"]:text-foreground',
      navbarButtonText: "text-foreground",
      badge: "!text-primary-foreground !bg-primary",
      menuButton: "",
      formButtonPrimary:
        "!bg-primary !text-primary-foreground !shadow-none hover:!bg-primary/90 !py-2",
      formFieldHintText: "text-muted-foreground",
      formFieldInput:
        "!py-2 !outline-none !border-border focus:ring-ring focus-visible:ring-ring/50 focus-visible:ring-[3px]",
      dividerLine: "bg-border",
      alternativeMethodsBlockButtonText: "text-foreground",
      socialButtonsBlockButton:
        "text-foreground !py-2 border-border hover:bg-accent hover:text-accent-foreground",
      socialButtonsBlockButtonText: "!text-foreground",
      formFieldInputShowPasswordButton: "",
      otpCodeFieldInput: "bg-card text-card-foreground border-border",
      userButtonPopoverRootBox: "bg-popover",
      userButtonBox: "text-popover-foreground hover:opacity-80",
      userButtonTrigger: "focus:shadow-none focus-visible:ring-ring/50",
      userPreviewSecondaryIdentifier: "text-muted-foreground!",
      userButtonPopoverMain: "bg-popover",
      userButtonPopoverCard: "border-border",
      userPreviewTextContainer: "text-popover-foreground",
      userButtonPopoverActions: "bg-background text-foreground",
      userButtonPopoverActionButton:
        "text-foreground! hover:bg-accent hover:text-primary-foreground!",
      userButtonPopoverActionButtonText: "text-popover-foreground",
      userButtonPopoverFooter: "border-border",
      userButtonOuterIdentifier: "whitespace-nowrap hidden md:block",
      organizationSwitcherTrigger: "text-foreground [&>svg]:text-foreground",
      userPreview: "bg-background",
      organizationSwitcherPopoverActions: "bg-background text-foreground",
      organizationSwitcherPopoverActionButton:
        "hover:bg-accent hover:text-primary-foreground! text-foreground!",
      organizationSwitcherPopoverActionButtonText: "text-foreground",
      organizationSwitcherPopoverFooter: "bg-background text-foreground",
    },
    overrides
  );
}

export default function clerkTheme(
  elementsOverride: Partial<Record<string, string>> = {}
): Theme {
  return {
    elements: clerkElements(elementsOverride),
  };
}
