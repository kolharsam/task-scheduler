import {
  ActivityIcon,
  ListChecksIcon,
  ListTodo,
  LogsIcon,
  Menu,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Sheet, SheetContent, SheetTrigger } from "@/components/ui/sheet";
import { Link, Route, Routes, useLocation } from "react-router-dom";
import clsx from "clsx";
import { Input } from "./components/ui/input";
import { Label } from "./components/ui/label";
import { useCallback, useEffect, useState } from "react";
import toast from "react-hot-toast";
import axios from "axios";
import {
  Healthz,
  TaskEventInfo,
  TaskEventInfoColumns,
  TaskInfo,
  TaskInfoTableColumns,
} from "./types";
import { DataTable } from "./DataTable";
import { HealthCard } from "./HealthCard";

export default function App() {
  const { pathname } = useLocation();

  return (
    <div className="grid min-h-screen w-full md:grid-cols-[220px_1fr] lg:grid-cols-[280px_1fr]">
      <div className="hidden border-r bg-muted/40 md:block">
        <div className="flex h-full max-h-screen flex-col gap-2">
          <div className="flex h-14 items-center border-b px-4 lg:h-[60px] lg:px-6">
            <Link to="/run" className="flex items-center gap-2 font-semibold">
              <span className="text-2xl">task-scheduler</span>
            </Link>
          </div>
          <div className="flex-1">
            <nav className="grid items-start px-2 text-sm font-medium lg:px-4">
              <Link
                to="/run"
                className={clsx(
                  "flex items-center gap-3 rounded-lg px-3 py-2 transition-all hover:text-primary",
                  ["/run", "/"].includes(pathname)
                    ? "text-primary bg-muted"
                    : "text-muted-foreground"
                )}
              >
                <ListTodo className="h-4 w-4" />
                Run Task
              </Link>
              <Link
                to="/task"
                className={clsx(
                  "flex items-center gap-3 rounded-lg px-3 py-2 transition-all hover:text-primary",
                  ["/task"].includes(pathname)
                    ? "text-primary bg-muted"
                    : "text-muted-foreground"
                )}
              >
                <ListChecksIcon className="h-4 w-4" />
                Pending Tasks
              </Link>
              <Link
                to="/logs"
                className={clsx(
                  "flex items-center gap-3 rounded-lg px-3 py-2 transition-all hover:text-primary",
                  pathname === "/logs"
                    ? "text-primary bg-muted "
                    : "text-muted-foreground"
                )}
              >
                <LogsIcon className="h-4 w-4" />
                Event Logs
              </Link>
              <Link
                to="/health"
                className={clsx(
                  "flex items-center gap-3 rounded-lg px-3 py-2 transition-all hover:text-primary",
                  pathname === "/health"
                    ? "text-primary bg-muted"
                    : "text-muted-foreground"
                )}
              >
                <ActivityIcon className="h-4 w-4" />
                Check Health
              </Link>
            </nav>
          </div>
        </div>
      </div>
      <div className="flex flex-col">
        <header className="flex h-14 items-center gap-4 border-b bg-muted/40 px-4 lg:h-[60px] lg:px-6">
          <Sheet>
            <SheetTrigger asChild>
              <Button
                variant="outline"
                size="icon"
                className="shrink-0 md:hidden"
              >
                <Menu className="h-5 w-5" />
                <span className="sr-only">Toggle navigation menu</span>
              </Button>
            </SheetTrigger>
            <SheetContent side="left" className="flex flex-col">
              <nav className="grid gap-2 text-lg font-medium">
                <Link
                  to="/run"
                  className="flex items-center gap-3 rounded-lg px-3 py-2 text-muted-foreground transition-all hover:text-primary"
                >
                  <ListTodo className="h-4 w-4" />
                  Run Task
                </Link>
                <Link
                  to="/health"
                  className="flex items-center gap-3 rounded-lg px-3 py-2 text-muted-foreground transition-all hover:text-primary"
                >
                  <ActivityIcon className="h-4 w-4" />
                  Check Health
                </Link>
                <Link
                  to="/logs"
                  className="flex items-center gap-3 rounded-lg bg-muted px-3 py-2 text-primary transition-all hover:text-primary"
                >
                  <LogsIcon className="h-4 w-4" />
                  Event Logs
                </Link>
              </nav>
            </SheetContent>
          </Sheet>
          <div className="w-full flex-1" />
        </header>
        <main className="flex flex-1 flex-col gap-4 p-4 lg:gap-6 lg:p-6">
          {/* MAIN CONTENT GOES HERE */}
          <Routes>
            <Route path="/run" element={<Run />} />
            <Route path="/logs" element={<EventLogs />} />
            <Route path="/health" element={<Health />} />
            <Route path="/task" element={<Tasks />} />
            <Route path="*" element={<Run />} />
          </Routes>
        </main>
      </div>
    </div>
  );
}

type TaskResponse = {
  data: {
    task_id: string;
  };
};

const Run = () => {
  const [command, setCommand] = useState("");
  const submitJob = useCallback(async () => {
    if (!command) {
      return;
    }
    const reqHeaders = new Headers();
    reqHeaders.append("content-type", "application/json");

    const trimmedCmd = command.trim();
    const reqBody = JSON.stringify({ command: trimmedCmd });
    const reqConfig = {
      method: "post",
      url: "http://localhost:8081/v1/task",
      headers: {
        "Content-Type": "application/json",
      },
      data: reqBody,
    };
    try {
      const res = await axios.request<TaskResponse>(reqConfig);
      if (res.data.data.task_id) {
        toast.success("added new task to run");
        setCommand("");
      }
    } catch (e) {
      console.error(e);
    }
  }, [command]);

  return (
    <>
      <div className="grid w-full max-w-sm items-center gap-1.5">
        <Label htmlFor="command">Command</Label>
        <Input
          type="text"
          id="command"
          placeholder="ls, netstat..."
          value={command}
          onChange={(e) => setCommand(e.target.value)}
        />
        <Button className="mt-2" onClick={submitJob} disabled={!command.length}>
          run cmd
        </Button>
      </div>
    </>
  );
};

const Tasks = () => {
  const [tasks, setTasks] = useState<TaskInfo[]>([]);

  const fetchTasks = async () => {
    const reqConfig = {
      method: "get",
      url: "http://localhost:8081/v1/task",
    };
    try {
      const resp = await axios.get<{ data: TaskInfo[] }>(reqConfig.url);
      if (resp.status === 200) {
        setTasks(resp.data.data);
      } else {
        setTasks([]);
      }
    } catch (err) {
      toast.error("there was an issue while fetching the tasks");
      console.error(err);
      return [];
    }
  };

  useEffect(() => {
    fetchTasks();
  }, []);

  return (
    <>
      <DataTable columns={TaskInfoTableColumns} data={tasks} />
    </>
  );
};

const EventLogs = () => {
  const [taskEvents, setTaskEvents] = useState<TaskEventInfo[]>([]);

  const fetchTaskEvents = async () => {
    const reqConfig = {
      method: "get",
      url: "http://localhost:8081/v1/events",
    };
    try {
      const resp = await axios.get<{ data: TaskEventInfo[] }>(reqConfig.url);
      if (resp.status === 200) {
        setTaskEvents(resp.data.data);
      } else {
        setTaskEvents([]);
      }
    } catch (err) {
      toast.error("there was an issue while fetching the tasks");
      console.error(err);
      return [];
    }
  };

  useEffect(() => {
    fetchTaskEvents();
  }, []);

  return (
    <>
      <DataTable columns={TaskEventInfoColumns} data={taskEvents} />
    </>
  );
};

const Health = () => {
  const [healthzData, setHealthz] = useState<Healthz>();

  const fetchTaskEvents = async () => {
    const reqConfig = {
      method: "get",
      url: "http://localhost:8081/v1/healthz",
    };
    try {
      const resp = await axios.get<Healthz>(reqConfig.url);
      if (resp.status === 200) {
        setHealthz(resp.data);
      } else {
        setHealthz(undefined);
      }
    } catch (err) {
      toast.error("there was an issue while fetching the health data");
      console.error(err);
      return [];
    }
  };

  useEffect(() => {
    fetchTaskEvents();
  }, []);

  return (
    <div className="w-full flex items-center justify-center">
      {healthzData ? (
        <HealthCard data={healthzData} />
      ) : (
        <span>unable to load health data</span>
      )}
    </div>
  );
};
