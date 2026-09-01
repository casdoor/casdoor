import {Toaster as SonnerToaster} from "sonner";
import {useTheme} from "@/hooks/use-theme";

export function Toaster() {
  const {resolvedTheme} = useTheme();
  return (
    <SonnerToaster
      theme={resolvedTheme as "light" | "dark"}
      position="top-center"
      richColors
      closeButton
      toastOptions={{
        classNames: {
          toast: "rounded-lg border shadow-lg",
        },
      }}
    />
  );
}
