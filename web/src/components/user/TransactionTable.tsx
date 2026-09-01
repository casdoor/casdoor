import i18next from "i18next";
import {Link} from "react-router-dom";
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from "@/components/ui/table";
import * as Setting from "@/lib/setting";

/**
 * The user's transactions, read-only. Ported from web/src/table/TransactionTable.js
 * with `hideTag` and no user/action columns, which is how the user page uses it.
 */
export function TransactionTable({transactions}: {transactions: any[]}) {
  const rows = transactions ?? [];

  return (
    <div className="max-h-[420px] overflow-auto rounded-lg border">
      <Table>
        <TableHeader>
          <TableRow className="hover:bg-transparent">
            <TableHead className="w-[220px]">{i18next.t("general:Name")}</TableHead>
            <TableHead className="w-[160px]">{i18next.t("general:Created time")}</TableHead>
            <TableHead className="w-[150px]">{i18next.t("general:Application")}</TableHead>
            <TableHead className="w-[120px]">{i18next.t("general:Category")}</TableHead>
            <TableHead className="w-[140px]">{i18next.t("general:Type")}</TableHead>
            <TableHead className="w-[150px]">{i18next.t("general:Provider")}</TableHead>
            <TableHead className="w-[120px]">{i18next.t("general:Payment")}</TableHead>
            <TableHead className="w-[110px]">{i18next.t("general:State")}</TableHead>
            <TableHead className="w-[140px]">{i18next.t("product:Amount")}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.length === 0 ? (
            <TableRow className="hover:bg-transparent">
              <TableCell colSpan={9} className="h-20 text-center text-muted-foreground">
                {i18next.t("general:No data")}
              </TableCell>
            </TableRow>
          ) : (
            rows.map((record: any) => (
              <TableRow key={`${record.owner}/${record.name}`}>
                <TableCell>
                  <Link to={`/transactions/${record.owner}/${record.name}`} className="underline-offset-4 hover:underline">
                    {record.name}
                  </Link>
                </TableCell>
                <TableCell>{Setting.getFormattedDate(record.createdTime)}</TableCell>
                <TableCell>
                  {record.application ? (
                    <Link to={`/applications/${record.owner}/${record.application}`} className="underline-offset-4 hover:underline">
                      {record.application}
                    </Link>
                  ) : null}
                </TableCell>
                <TableCell>{record.category}</TableCell>
                <TableCell>{record.type}</TableCell>
                <TableCell>
                  {record.provider ? (
                    <Link to={`/providers/${record.owner}/${record.provider}`} className="underline-offset-4 hover:underline">
                      {record.provider}
                    </Link>
                  ) : null}
                </TableCell>
                <TableCell>
                  {record.payment ? (
                    <Link to={`/payments/${record.owner}/${record.payment}`} className="underline-offset-4 hover:underline">
                      {record.payment}
                    </Link>
                  ) : null}
                </TableCell>
                <TableCell>{record.state}</TableCell>
                <TableCell>{Setting.getPriceDisplay(record.amount, record.currency)}</TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </div>
  );
}
