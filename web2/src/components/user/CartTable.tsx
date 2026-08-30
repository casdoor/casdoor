import i18next from "i18next";
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from "@/components/ui/table";
import * as Setting from "@/lib/setting";

/** The products a user has in their cart, ported from web/src/table/CartTable.js. */
export function CartTable({cart}: {cart: any[]}) {
  const rows = cart ?? [];

  return (
    <div className="overflow-x-auto rounded-lg border">
      <Table>
        <TableHeader>
          <TableRow className="hover:bg-transparent">
            <TableHead className="w-[200px]">{i18next.t("general:Name")}</TableHead>
            <TableHead className="w-[80px]">{i18next.t("product:Image")}</TableHead>
            <TableHead className="w-[120px]">{i18next.t("order:Price")}</TableHead>
            <TableHead className="w-[100px]">{i18next.t("product:Quantity")}</TableHead>
            <TableHead>{i18next.t("general:Detail")}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.length === 0 ? (
            <TableRow className="hover:bg-transparent">
              <TableCell colSpan={5} className="h-20 text-center text-muted-foreground">
                {i18next.t("general:No data")}
              </TableCell>
            </TableRow>
          ) : (
            rows.map((item: any, index: number) => (
              <TableRow key={`${item.owner}/${item.name}/${index}`}>
                <TableCell>{item.displayName}</TableCell>
                <TableCell>
                  {item.image ? (
                    <a target="_blank" rel="noreferrer" href={item.image}>
                      <img src={item.image} alt={item.displayName} className="h-10 w-10 object-contain" />
                    </a>
                  ) : null}
                </TableCell>
                <TableCell>{`${Setting.getCurrencySymbol(item.currency)}${item.price ?? ""}`}</TableCell>
                <TableCell>{item.quantity}</TableCell>
                <TableCell className="text-muted-foreground">{item.detail}</TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </div>
  );
}
