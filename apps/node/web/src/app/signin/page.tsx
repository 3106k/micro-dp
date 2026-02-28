import { Suspense } from "react";

import { SigninForm } from "./signin-form";

export default function SigninPage() {
  return (
    <main className="flex min-h-screen items-center justify-center p-4">
      <Suspense>
        <SigninForm />
      </Suspense>
    </main>
  );
}
