import { Button } from "@/components/ui/button";

export default function Home() {
  return (
    <main className="container flex min-h-screen flex-col items-start justify-center gap-4 py-12">
      <div className="space-y-2">
        <h1 className="text-3xl font-semibold tracking-tight">micro-dp</h1>
        <p className="text-muted-foreground">
          Data pipeline management UI foundation with shadcn/ui
        </p>
      </div>
      <div className="flex gap-2">
        <Button>Primary Button</Button>
        <Button variant="outline">Outline Button</Button>
      </div>
    </main>
  );
}
