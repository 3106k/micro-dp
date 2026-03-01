import "./globals.css";
import { ToastProvider } from "@/components/ui/toast-provider";

export const metadata = {
  title: "micro-dp",
  description: "Data pipeline management",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="ja">
      <body className="min-h-screen bg-background text-foreground antialiased">
        <ToastProvider>{children}</ToastProvider>
      </body>
    </html>
  );
}
