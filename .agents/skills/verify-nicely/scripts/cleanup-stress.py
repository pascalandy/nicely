#!/usr/bin/env python
"""Repeat the real public-host cancellation path with measured private staging."""

import json
import os
import re
import signal
import subprocess
import sys
import time
from pathlib import Path


def require(condition: object, problem: object) -> None:
    if not condition:
        raise RuntimeError(problem)


def running(pid: int) -> bool:
    result = subprocess.run(
        ["ps", "-p", str(pid), "-o", "stat="],
        text=True,
        capture_output=True,
        check=False,
    )
    status = result.stdout.strip()
    return bool(status) and not status.startswith("Z")


def parent(pid: int) -> int:
    return int(
        subprocess.check_output(["ps", "-p", str(pid), "-o", "ppid="], text=True)
    )


def drive(run: Path, repo: Path, folder: Path, sig: signal.Signals, files: int) -> dict:
    env = dict(os.environ, CLEANUP_WORKER="stubborn", CLEANUP_FILES=str(files))
    program = repo / ".agents/skills/verify-nicely/scripts/verify.sh"
    call = subprocess.Popen(
        [
            "bash",
            str(program),
            "step",
            str(run),
            "cleanup-stress",
            "hello",
            "staged",
            str(folder / "worker.sh"),
            "--json",
        ],
        cwd=folder,
        env=env,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        start_new_session=True,
    )
    pids: list[int] = []
    try:
        deadline = time.monotonic() + 15
        while not (folder / "worker.pid").exists():
            if call.poll() is not None or time.monotonic() >= deadline:
                raise RuntimeError("worker did not become ready")
            time.sleep(0.01)
        pids = [
            int((folder / name).read_text())
            for name in ["worker.pid", "grandchild.pid"]
        ]
        stages = list(folder.glob(".ncly-stage-*"))
        require(
            len(stages) == 1 and len(list(stages[0].iterdir())) == files + 1,
            "wrong staging population",
        )
        require(stages[0].stat().st_mode & 63 == 0, "staging is not private")
        host = parent(parent(pids[0]))
        command = subprocess.check_output(
            ["ps", "-ww", "-p", str(host), "-o", "command="], text=True
        )
        # ps keeps argv's spelling, while Path.resolve follows /var on macOS.
        # Try word boundaries so an executable path may also contain spaces.
        command = command.strip()
        boundaries = [i for i, char in enumerate(command) if char == " "] + [
            len(command)
        ]
        candidates = [Path(command[:end]) for end in boundaries]
        require(
            any(
                path.is_file() and path.samefile(run / "build/ncly")
                for path in candidates
            ),
            f"worker ancestry expected {run / 'build/ncly'}, got {command!r}",
        )
        started = time.monotonic()
        os.kill(host, sig)
        stdout, stderr = call.communicate(timeout=20)
        elapsed = round((time.monotonic() - started) * 1000)
        want_exit, want_code = (
            (130, "INTERRUPTED") if sig == signal.SIGINT else (143, "TERMINATED")
        )
        require(call.returncode == want_exit, (call.returncode, stdout, stderr))
        answers = [
            json.loads(line) for line in stdout.splitlines() if line.startswith("{")
        ]
        answer = answers[-1]
        require(
            not answer["ok"]
            and answer["errors"][0]["code"] == want_code
            and answer["stage"],
            "verification failed: not answer['ok'] and answer['errors'][0]['code'] == want_code and answer['stage']",
        )
        require(not list(folder.glob(".ncly-stage-*")), "private staging remains")
        require(
            (folder / "cleanup.done").read_text() == "done\n",
            "handler cleanup did not finish",
        )
        require(not any(running(pid) for pid in pids), "a worker survived")
        require(
            (folder / "input.txt").read_text() == "original input\n", "input changed"
        )
        require(
            (folder / "output.txt").read_text() == "previous output\n",
            "previous output changed",
        )
        stats = json.loads((folder / "cleanup.json").read_text())
        require(
            all(isinstance(stats[key], int) for key in ["delete_ms", "cleanup_ms"]),
            "verification failed: all((isinstance(stats[key], int) for key in ['delete_ms', 'cleanup_ms']))",
        )
        require(stats["cleanup_ms"] < 2000, "handler cleanup exhausted its reservation")
        return dict(
            stats,
            signal=sig.name,
            files=files,
            signal_ms=elapsed,
            margin_ms=2000 - stats["cleanup_ms"],
        )
    finally:
        # These IDs come only from this invocation's private fixture.
        # Refuse a surviving worker before cleaning it up, even if the host exited.
        for path in [folder / "worker.pid", folder / "grandchild.pid"]:
            if path.exists():
                pid = int(path.read_text())
                if pid > 1 and running(pid):
                    try:
                        os.kill(pid, signal.SIGKILL)
                    except ProcessLookupError:
                        pass
        if call.poll() is None:
            os.killpg(call.pid, signal.SIGKILL)
            call.communicate()


def main() -> None:
    run = Path(sys.argv[1]).resolve()
    count = int(sys.argv[2]) if len(sys.argv) > 2 else 10
    files = int(sys.argv[3]) if len(sys.argv) > 3 else 20000
    require(count > 0 and files >= 0, "verification failed: count > 0 and files >= 0")
    repo = Path(__file__).resolve().parents[4]
    head = subprocess.check_output(
        ["git", "-C", str(repo), "rev-parse", "HEAD"], text=True
    ).strip()
    recorded = (run / "evidence/commit.txt").read_text().splitlines()
    require(
        recorded == [head[:12]],
        "stress requires a clean committed build at the current head",
    )
    archive = (
        repo / "testdata/extensions/hello/testdata/script/staged_cleanup.txtar"
    ).read_text()
    parts = re.split(r"^-- (.+) --\n", archive, flags=re.MULTILINE)
    fixtures = dict(zip(parts[1::2], parts[2::2], strict=True))
    results = []
    for sig in [signal.SIGINT, signal.SIGTERM]:
        for index in range(1, count + 1):
            folder = run / "scratch" / f"stress-{sig.name}-{index:02d}"
            folder.mkdir()
            for name in ["worker.sh", "grandchild.sh", "input.txt", "output.txt"]:
                (folder / name).write_text(fixtures[name])
            result = dict(drive(run, repo, folder, sig, files), run=index)
            results.append(result)
            text = json.dumps(result, sort_keys=True)
            print(text, flush=True)
            with (run / "evidence/cleanup-stress.txt").open("a") as evidence:
                evidence.write(text + "\n")
    summary = {
        "head": head,
        "runs": len(results),
        "files": files,
        "max_delete_ms": max(row["delete_ms"] for row in results),
        "max_cleanup_ms": max(row["cleanup_ms"] for row in results),
        "min_margin_ms": min(row["margin_ms"] for row in results),
    }
    (run / "evidence/cleanup-stress.json").write_text(
        json.dumps(summary, indent=2) + "\n"
    )
    print(json.dumps(summary, sort_keys=True), flush=True)


if __name__ == "__main__":
    main()
