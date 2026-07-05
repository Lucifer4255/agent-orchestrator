import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { TmuxPreflightGate } from "./TmuxPreflightGate";

const { getMock } = vi.hoisted(() => ({
	getMock: vi.fn(),
}));

vi.mock("../lib/api-client", () => ({
	apiClient: { GET: getMock },
	apiErrorMessage: (error: unknown, fallback = "Request failed") =>
		typeof error === "object" && error !== null && "message" in error && typeof (error as { message?: unknown }).message === "string"
			? ((error as { message: string }).message ?? fallback)
			: fallback,
}));

function renderGate(daemonStatus: { state: "ready" | "stopped"; port?: number }) {
	const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
	return render(
		<QueryClientProvider client={queryClient}>
			<TmuxPreflightGate daemonStatus={daemonStatus} />
		</QueryClientProvider>,
	);
}

beforeEach(() => {
	getMock.mockReset();
});

describe("TmuxPreflightGate", () => {
	it("does not render when the daemon is not ready", () => {
		renderGate({ state: "stopped" });
		expect(screen.queryByText("Install tmux to run agent sessions")).not.toBeInTheDocument();
		expect(getMock).not.toHaveBeenCalled();
	});

	it("does not render when tmux is available", async () => {
		getMock.mockResolvedValue({
			data: { available: true, runtime: "tmux", platform: "darwin", path: "/opt/homebrew/bin/tmux" },
		});
		renderGate({ state: "ready", port: 3001 });
		await waitFor(() => expect(getMock).toHaveBeenCalled());
		expect(screen.queryByText("Install tmux to run agent sessions")).not.toBeInTheDocument();
	});

	it("blocks the shell and shows install guidance when tmux is missing", async () => {
		getMock.mockResolvedValue({
			data: {
				available: false,
				runtime: "tmux",
				platform: "darwin",
				message: "tmux not found in PATH; required on macOS/Linux to start agent sessions",
				installHint: "brew install tmux",
			},
		});
		renderGate({ state: "ready", port: 3001 });

		expect(await screen.findByText("Install tmux to run agent sessions")).toBeInTheDocument();
		expect(screen.getByText("brew install tmux")).toBeInTheDocument();
	});

	it("re-probes when Re-check is clicked", async () => {
		getMock
			.mockResolvedValueOnce({
				data: {
					available: false,
					runtime: "tmux",
					platform: "darwin",
					installHint: "brew install tmux",
				},
			})
			.mockResolvedValueOnce({
				data: {
					available: true,
					runtime: "tmux",
					platform: "darwin",
					path: "/opt/homebrew/bin/tmux",
				},
			});
		renderGate({ state: "ready", port: 3001 });
		await screen.findByText("Install tmux to run agent sessions");

		await userEvent.click(screen.getByRole("button", { name: "Re-check" }));

		await waitFor(() => expect(getMock).toHaveBeenCalledTimes(2));
		await waitFor(() => expect(screen.queryByText("Install tmux to run agent sessions")).not.toBeInTheDocument());
	});
});
