import { PublicHeader } from "@/components/public-header";

export default function PublicLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <div className="min-h-screen flex flex-col">
      <PublicHeader />
      {children}
    </div>
  );
}
