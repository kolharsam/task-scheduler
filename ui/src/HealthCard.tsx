import { cn } from "@/lib/utils";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Healthz } from "./types";

type CardProps = React.ComponentProps<typeof Card>;
interface HealthCardProps extends CardProps {
  data: Healthz;
}

export function HealthCard({ data, className, ...cardProps }: HealthCardProps) {
  return (
    <Card className={cn("w-[380px]", className)} {...cardProps}>
      <CardHeader>
        <CardTitle>Services Health</CardTitle>
      </CardHeader>
      <CardContent className="grid gap-4">
        <div>
          <div className="mb-4 grid grid-cols-[25px_1fr] items-start pb-4 last:mb-0 last:pb-0">
            <span
              className={cn(
                "flex h-2 w-2 translate-y-1 rounded-full",
                data.status === "OK" ? "bg-lime-400" : "bg-red-600"
              )}
            />
            <div className="space-y-1">
              <p className="text-sm font-medium leading-none">
                Overall System Status
              </p>
            </div>
          </div>
          <div className="mb-4 grid grid-cols-[25px_1fr] items-start pb-4 last:mb-0 last:pb-0">
            <span
              className={cn(
                "flex h-2 w-2 translate-y-1 rounded-full",
                data.data.db_status === "OK" ? "bg-lime-400" : "bg-red-600"
              )}
            />
            <div className="space-y-1">
              <p className="text-sm font-medium leading-none">
                Database Status
              </p>
            </div>
          </div>
          <div className="mb-4 grid grid-cols-[25px_1fr] items-start pb-4 last:mb-0 last:pb-0">
            <span
              className={cn(
                "flex h-2 w-2 translate-y-1 rounded-full",
                data.status === "OK" ? "bg-lime-400" : "bg-red-600"
              )}
            />
            <div className="space-y-1">
              <p className="text-sm font-medium leading-none">Workers Status</p>
              <p className="text-sm text-muted-foreground">
                {data.data.workers_status &&
                  Object.keys(data.data.workers_status)}{" "}
                workers are connected
              </p>
            </div>
          </div>
          {data.data.error && (
            <div className="mb-4 grid grid-cols-[25px_1fr] items-start pb-4 last:mb-0 last:pb-0">
              Error:
              <details>{data.data.error}</details>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
