import { describe, expect, it } from "vitest";
import { bundledTmuxDaemonEnv, resolveBundledTmuxPath } from "./bundled-tmux";

describe("resolveBundledTmuxPath", () => {
	it("returns null in dev and on Windows", () => {
		expect(resolveBundledTmuxPath(false, "/resources", "darwin")).toBeNull();
		expect(resolveBundledTmuxPath(true, "/resources", "win32")).toBeNull();
	});

	it("points at the packaged resources dir on macOS/Linux", () => {
		expect(resolveBundledTmuxPath(true, "/Applications/Agent Orchestrator.app/Contents/Resources", "darwin")).toBe(
			"/Applications/Agent Orchestrator.app/Contents/Resources/tmux/tmux",
		);
		expect(resolveBundledTmuxPath(true, "/opt/Agent Orchestrator/resources", "linux")).toBe(
			"/opt/Agent Orchestrator/resources/tmux/tmux",
		);
	});
});

describe("bundledTmuxDaemonEnv", () => {
	it("sets AO_BUNDLED_TMUX_BIN only when a bundled binary exists", () => {
		expect(bundledTmuxDaemonEnv(false, "/resources", "darwin")).toEqual({});
		expect(
			bundledTmuxDaemonEnv(true, "/Applications/Agent Orchestrator.app/Contents/Resources", "darwin"),
		).toEqual({
			AO_BUNDLED_TMUX_BIN: "/Applications/Agent Orchestrator.app/Contents/Resources/tmux/tmux",
		});
	});
});
