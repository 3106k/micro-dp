import Link from "next/link";

import { Button } from "@/components/ui/button";

export default function Home() {
  return (
    <main className="container flex min-h-screen flex-col items-start justify-center gap-4 py-12">
      <div className="space-y-2">
        <h1 className="text-3xl font-semibold tracking-tight">micro-dp</h1>
        <p className="text-muted-foreground">
          Data pipeline management platform
        </p>
      </div>
      <div className="flex gap-2">
        <Link href="/signin">
          <Button>Sign In</Button>
        </Link>
        <Link href="/signup">
          <Button variant="outline">Sign Up</Button>
        </Link>
      </div>
    </main>
  );
}
