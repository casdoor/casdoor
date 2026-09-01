import * as React from "react";
import {Link} from "react-router-dom";
import {Card, CardContent} from "@/components/ui/card";
import * as Setting from "@/lib/setting";

export interface GridCardItem {
  link: string;
  name: string;
  description?: string;
  logo?: string;
  createdTime?: string;
  isExternal?: boolean;
}

/** Card grid used by the Home > Apps and Home > Shortcuts pages. */
export function GridCards({items}: {items: GridCardItem[]}) {
  if (!items || items.length === 0) {
    return null;
  }

  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      {items.map((item) => {
        const body = (
          <Card className="h-full transition-colors hover:border-foreground/20">
            <CardContent className="flex items-start gap-3 p-4">
              {item.logo ? (
                <img src={item.logo} alt={item.name} className="h-10 w-10 shrink-0 rounded object-contain" />
              ) : null}
              <div className="min-w-0">
                <div className="truncate font-medium">{item.name}</div>
                {item.description ? (
                  <p className="mt-0.5 line-clamp-2 text-sm text-muted-foreground">{item.description}</p>
                ) : null}
                {item.createdTime ? (
                  <p className="mt-1 text-xs text-muted-foreground">{Setting.getFormattedDate(item.createdTime)}</p>
                ) : null}
              </div>
            </CardContent>
          </Card>
        );

        return item.isExternal ? (
          <a key={item.link + item.name} href={item.link} target="_blank" rel="noreferrer">
            {body}
          </a>
        ) : (
          <Link key={item.link + item.name} to={item.link}>
            {body}
          </Link>
        );
      })}
    </div>
  );
}
