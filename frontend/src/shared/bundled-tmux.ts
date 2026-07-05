function joinPath(...segments: string[]): string {
	return segments.map((segment) => segment.replace(/[/\\]+$/, "")).join("/");
}

export function bundledTmuxBinaryName(platform: NodeJS.Platform): string {
	return platform === "win32" ? "" : "tmux";
}

/** Absolute path to the tmux binary shipped in the desktop app bundle, if any. */
export function resolveBundledTmuxPath(
	isPackaged: boolean,
	resourcesPath: string,
	platform: NodeJS.Platform,
): string | null {
	if (!isPackaged || platform === "win32") return null;
	const name = bundledTmuxBinaryName(platform);
	if (!name) return null;
	return joinPath(resourcesPath, "tmux", name);
}

/** Env overrides passed to the daemon for tmux resolution (see tmux.ResolveBinary). */
export function bundledTmuxDaemonEnv(
	isPackaged: boolean,
	resourcesPath: string,
	platform: NodeJS.Platform,
): Record<string, string> {
	const bundled = resolveBundledTmuxPath(isPackaged, resourcesPath, platform);
	return bundled ? { AO_BUNDLED_TMUX_BIN: bundled } : {};
}
