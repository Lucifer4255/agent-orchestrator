import { existsSync } from "node:fs";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { spawnSync } from "node:child_process";
import { mkdirSync, writeFileSync } from "node:fs";

const scriptsDir = dirname(fileURLToPath(import.meta.url));
const repoRoot = resolve(scriptsDir, "..", "..");
const outDir = join(repoRoot, "frontend", "tmux");
const buildScript = join(repoRoot, "scripts", "build-tmux.sh");

mkdirSync(outDir, { recursive: true });

if (process.platform === "win32") {
	writeFileSync(
		join(outDir, "NOTICE.md"),
		"tmux is not bundled on Windows; the desktop app uses ConPTY instead.\n",
	);
	process.exit(0);
}

if (!existsSync(buildScript)) {
	console.error(`missing build script: ${buildScript}`);
	process.exit(1);
}

const result = spawnSync("bash", [buildScript, outDir], {
	cwd: repoRoot,
	stdio: "inherit",
});

if (result.error) {
	console.error(`failed to start build-tmux.sh: ${result.error.message}`);
	process.exit(1);
}

if (result.status !== 0) {
	process.exit(result.status ?? 1);
}

if (!existsSync(join(outDir, "tmux"))) {
	console.error(`build-tmux.sh did not produce ${join(outDir, "tmux")}`);
	process.exit(1);
}
