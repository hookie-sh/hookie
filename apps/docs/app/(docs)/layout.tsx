import { baseOptions } from "@/lib/layout.shared";
import { source } from "@/lib/source";
import { DocsLayout, type DocsLayoutProps } from "fumadocs-ui/layouts/docs";

function docsOptions(): DocsLayoutProps {
  return {
    ...baseOptions(),
    tree: source.getPageTree(),
    // links: [
    //   {
    //     type: "custom",
    //     children: <GithubInfo owner="hookie-sh" repo="hookie" />,
    //   },
    // ],
  };
}

export default function Layout({ children }: LayoutProps<"/">) {
  return <DocsLayout {...docsOptions()}>{children}</DocsLayout>;
}
