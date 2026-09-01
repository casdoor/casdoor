import * as React from "react";
import {cn} from "@/lib/utils";

interface DayCount {
  date: string;
  count: number;
}

/**
 * The sign-in calendar heatmap of the dashboard. antd draws this with ECharts'
 * calendar coordinate system; recharts has no equivalent, so it is laid out by
 * hand: one column per week, one row per weekday, exactly like a contribution
 * graph. Colours come from the theme so it works in both light and dark.
 */
export function ActivityHeatmap({
  data,
  maxCount,
  dateRange,
  className,
}: {
  data: DayCount[];
  maxCount: number;
  dateRange?: [string, string] | string[];
  className?: string;
}) {
  const {weeks, monthLabels} = React.useMemo(() => {
    const counts = new Map<string, number>();
    (data ?? []).forEach((item) => counts.set(item.date, item.count));

    const end = dateRange?.[1] ? new Date(`${dateRange[1]}T00:00:00`) : new Date();
    let start: Date;
    if (dateRange?.[0]) {
      start = new Date(`${dateRange[0]}T00:00:00`);
    } else {
      start = new Date(end);
      start.setFullYear(end.getFullYear() - 1);
    }

    // pad back to the Sunday on or before the first day, so rows line up
    const cursor = new Date(start);
    cursor.setDate(cursor.getDate() - cursor.getDay());

    const built: {date: string; count: number; inRange: boolean}[][] = [];
    const labels: {index: number; label: string}[] = [];
    let lastMonth = -1;

    while (cursor <= end) {
      const week: {date: string; count: number; inRange: boolean}[] = [];
      for (let day = 0; day < 7; day++) {
        const iso = `${cursor.getFullYear()}-${String(cursor.getMonth() + 1).padStart(2, "0")}-${String(
          cursor.getDate(),
        ).padStart(2, "0")}`;
        const inRange = cursor >= start && cursor <= end;
        week.push({date: iso, count: counts.get(iso) ?? 0, inRange});
        if (day === 0 && inRange && cursor.getMonth() !== lastMonth) {
          lastMonth = cursor.getMonth();
          labels.push({index: built.length, label: cursor.toLocaleString(undefined, {month: "short"})});
        }
        cursor.setDate(cursor.getDate() + 1);
      }
      built.push(week);
    }

    return {weeks: built, monthLabels: labels};
  }, [data, dateRange]);

  const max = Math.max(maxCount ?? 0, 1);
  // five buckets, matching the five-stop colour ramp the antd dashboard uses
  const level = (count: number) => (count === 0 ? 0 : Math.min(4, Math.ceil((count / max) * 4)));
  // the ramp rides --chart-1, not --primary: the chrome is deliberately neutral,
  // so a greyscale heatmap would read as disabled rather than as data
  const LEVEL_CLASS = [
    "bg-muted",
    "bg-chart-1/25",
    "bg-chart-1/45",
    "bg-chart-1/70",
    "bg-chart-1",
  ];

  return (
    <div className={cn("overflow-x-auto", className)}>
      <div className="inline-block min-w-full">
        <div className="flex gap-[3px] pl-1 text-[10px] text-muted-foreground">
          {weeks.map((_, index) => {
            const label = monthLabels.find((item) => item.index === index);
            return (
              <span key={index} className="w-[13px] shrink-0">
                {label ? label.label : ""}
              </span>
            );
          })}
        </div>
        <div className="flex gap-[3px] pl-1 pt-1">
          {weeks.map((week, weekIndex) => (
            <div key={weekIndex} className="flex flex-col gap-[3px]">
              {week.map((day) => (
                <div
                  key={day.date}
                  title={`${day.date}  ${day.count}`}
                  className={cn(
                    "h-[13px] w-[13px] rounded-[2px]",
                    day.inRange ? LEVEL_CLASS[level(day.count)] : "bg-transparent",
                  )}
                />
              ))}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
