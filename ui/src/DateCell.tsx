import { Row } from "@tanstack/react-table";
import { TaskEventInfo, TaskInfo } from "./types";
import { useMemo } from "react";

const DateCell = ({ row, value }: { row: (Row<TaskInfo> | Row<TaskEventInfo>), value: string }) => {
  const formattedDate = useMemo(() => {
    const date = new Date(row.getValue(value));
    return new Intl.DateTimeFormat('en-US', {
      year: 'numeric',
      month: 'short',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      timeZone: Intl.DateTimeFormat().resolvedOptions().timeZone,
    }).format(date);
  }, [row, value]);

  return <span title={formattedDate}>{formattedDate}</span>;
};

export default DateCell;
