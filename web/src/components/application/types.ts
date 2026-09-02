/** every tab of the application edit page writes back through the same setter */
export interface ApplicationTabProps {
  application: any;
  updateField: (field: string, value: any) => void;
}

/** how the edit page lays its tabs out: across the top, or down the left side */
export type MenuMode = "horizontal" | "vertical";
