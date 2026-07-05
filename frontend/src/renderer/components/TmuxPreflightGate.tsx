import { useQuery } from "@tanstack/react-query";
import { Loader2, Terminal } from "lucide-react";
import { Button } from "./ui/button";
import { apiClient, apiErrorMessage } from "../lib/api-client";
import type { components } from "../../api/schema";
import type { DaemonStatus } from "../../shared/daemon-status";

export type RuntimeStatus = components["schemas"]["RuntimeStatusResponse"];

export const runtimeStatusQueryKey = ["runtime-status"] as const;

export async function fetchRuntimeStatus(): Promise<RuntimeStatus> {
	const { data, error } = await apiClient.GET("/api/v1/runtime/status");
	if (error) throw new Error(apiErrorMessage(error, "Could not check terminal runtime"));
	if (!data) throw new Error("Could not check terminal runtime");
	return data;
}

type TmuxPreflightGateProps = {
	daemonStatus: DaemonStatus;
};

// TmuxPreflightGate blocks the shell when the daemon is up but tmux is missing on
// macOS/Linux. Windows uses ConPTY and never hits this path.
export function TmuxPreflightGate({ daemonStatus }: TmuxPreflightGateProps) {
	const enabled = daemonStatus.state === "ready" && Boolean(daemonStatus.port);
	const query = useQuery({
		queryKey: runtimeStatusQueryKey,
		queryFn: fetchRuntimeStatus,
		enabled,
		staleTime: 0,
	});

	if (!enabled || query.isLoading || query.isError || query.data?.available !== false || query.data?.runtime !== "tmux") {
		return null;
	}

	const installHint = query.data.installHint?.trim() || "brew install tmux";

	return (
		<div className="fixed inset-0 z-[100] flex items-center justify-center bg-background/95 p-6 backdrop-blur-sm">
			<div className="w-full max-w-[520px] rounded-lg border border-border bg-surface p-6 shadow-lg">
				<div className="flex items-start gap-3">
					<div className="mt-0.5 flex h-9 w-9 shrink-0 items-center justify-center rounded-md border border-border bg-background">
						<Terminal className="h-4 w-4 text-muted-foreground" aria-hidden />
					</div>
					<div className="min-w-0 flex-1">
						<h1 className="text-sm font-medium text-foreground">Install tmux to run agent sessions</h1>
						<p className="mt-2 text-[13px] leading-[1.5] text-muted-foreground">
							AO runs every agent session inside tmux on macOS and Linux. tmux was not found in your PATH, so new
							sessions cannot start until it is installed.
						</p>
						<div className="mt-4 rounded-md border border-border bg-background px-3 py-2">
							<p className="font-mono text-[11px] uppercase tracking-[0.12em] text-muted-foreground">Install</p>
							<pre className="mt-2 whitespace-pre-wrap font-mono text-[12px] leading-[1.5] text-foreground">{installHint}</pre>
						</div>
						{query.data.message ? (
							<p className="mt-3 text-[12px] text-muted-foreground">{query.data.message}</p>
						) : null}
						{query.isFetching ? (
							<p className="mt-3 flex items-center gap-2 text-[12px] text-muted-foreground">
								<Loader2 className="h-3.5 w-3.5 animate-spin" aria-hidden />
								Re-checking tmux…
							</p>
						) : null}
					</div>
				</div>
				<div className="mt-5 flex justify-end">
					<Button disabled={query.isFetching} onClick={() => void query.refetch()} type="button" variant="primary">
						Re-check
					</Button>
				</div>
			</div>
		</div>
	);
}
