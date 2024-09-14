import { ColumnDef } from "@tanstack/react-table";
import JsonCell from "./JSONCell";
import DateCell from "./DateCell";

export type TaskInfo = {
  task_id: string;
  command: string;
  status: "COMPLETED" | "RUNNING" | "CREATED";
  created_at: string;
  updated_at: string;
  completed_at?: string;
};

export type TaskEventInfo = {
  task_id: string;
  data: Record<string, string>;
  status: "SUCCESS" | "RUNNING" | "FAILED";
  created_at: string;
  updated_at: string;
  worker_id: string;
};

export type Healthz = {
  data: {
    db_status: "OK" | "NO";
    workers_status?: Record<string, string>;
    error?: string;
  };
  status: "OK" | "NO";
};

export const TaskEventInfoColumns: ColumnDef<TaskEventInfo>[] = [
  {
    accessorKey: "task_id",
    header: "ID",
  },
  {
    accessorKey: "worker_id",
    header: "Worker",
  },
  {
    accessorKey: "status",
    header: "Status",
  },
  {
    accessorKey: "data",
    header: "Data",
    cell: ({ row }) => {
      return (
        <JsonCell row={row} />
      );
    },
  },
  {
    accessorKey: "created_at",
    header: "Time",
    cell: ({ row }) => (<DateCell row={row} value="created_at" />),
  },
];

export const TaskInfoTableColumns: ColumnDef<TaskInfo>[] = [
  {
    accessorKey: "task_id",
    header: "ID",
  },
  {
    accessorKey: "command",
    header: "Command",
  },
  {
    accessorKey: "status",
    header: "Status",
  },
  {
    accessorKey: "created_at",
    header: "Created",
    cell: ({ row }) => (<DateCell row={row} value="created_at" />),
  },
  {
    accessorKey: "completed_at",
    header: "Completed",
    cell: ({ row }) => (<DateCell row={row} value="completed_at" />),
  },
];



