import type { ButtonProps } from "@/components/ui/button";
import { Button } from "@/components/ui/button";

type SubmitButtonProps = ButtonProps & {
  loading: boolean;
  loadingLabel: string;
};

export function SubmitButton({
  loading,
  loadingLabel,
  children,
  disabled,
  ...props
}: SubmitButtonProps) {
  return (
    <Button {...props} disabled={loading || disabled}>
      {loading ? loadingLabel : children}
    </Button>
  );
}
