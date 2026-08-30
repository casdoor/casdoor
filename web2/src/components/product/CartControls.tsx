import * as React from "react";
import {Minus, Plus, ShoppingCart} from "lucide-react";
import {useNavigate} from "react-router-dom";
import {Button} from "@/components/ui/button";
import * as UserBackend from "@/backend/UserBackend";
import {cn} from "@/lib/utils";

interface QuantityStepperProps {
  value: number;
  min?: number;
  max?: number;
  disabled?: boolean;
  onIncrease: () => void;
  onDecrease: () => void;
  /** omit to make the number read-only, as the cart page does */
  onChange?: (value: number) => void;
  className?: string;
}

/** "- n +" control of the store and cart pages (web/src/common/product/CartControls.js). */
export function QuantityStepper({
  value,
  min = 1,
  max,
  disabled,
  onIncrease,
  onDecrease,
  onChange,
  className,
}: QuantityStepperProps) {
  const parsed = value === null || value === undefined ? NaN : Number(value);
  const normalized = Number.isFinite(parsed) ? parsed : min;

  return (
    <div className={cn("inline-flex h-9 items-center rounded-md border", className)}>
      <Button
        variant="ghost"
        size="iconSm"
        className="h-full rounded-r-none"
        aria-label="Decrease"
        disabled={disabled || normalized <= min}
        onClick={onDecrease}
      >
        <Minus />
      </Button>
      <input
        type="number"
        className="h-full w-12 border-0 bg-transparent text-center text-sm outline-none [appearance:textfield] [&::-webkit-inner-spin-button]:appearance-none"
        value={normalized}
        readOnly={!onChange}
        disabled={disabled}
        onChange={(e) => onChange?.(Number(e.target.value))}
      />
      <Button
        variant="ghost"
        size="iconSm"
        className="h-full rounded-l-none"
        aria-label="Increase"
        disabled={disabled || (max !== undefined && normalized >= max)}
        onClick={onIncrease}
      >
        <Plus />
      </Button>
    </div>
  );
}

/** Floating cart button with the item count badge. */
export function FloatingCartButton({itemCount}: {itemCount: number}) {
  const navigate = useNavigate();

  return (
    <button
      type="button"
      onClick={() => navigate("/cart")}
      className="fixed bottom-12 right-12 z-50 flex h-14 w-14 items-center justify-center rounded-full bg-primary text-primary-foreground shadow-lg transition-transform hover:scale-105"
      aria-label="Cart"
    >
      <ShoppingCart className="!h-6 !w-6" />
      {itemCount > 0 ? (
        <span className="absolute -right-1 -top-1 flex h-5 min-w-5 items-center justify-center rounded-full bg-destructive px-1 text-xs text-destructive-foreground">
          {itemCount}
        </span>
      ) : null}
    </button>
  );
}

/** Number of items in the signed-in user's cart, refreshed on demand. */
export function useCartItemCount(account: any): [number, (count: number) => void] {
  const [count, setCount] = React.useState(0);

  React.useEffect(() => {
    if (!account) {
      return;
    }
    UserBackend.getUser(account.owner, account.name).then((res: any) => {
      if (res.status === "ok" && res.data?.cart) {
        setCount(res.data.cart.length);
      }
    });
  }, [account]);

  return [count, setCount];
}
