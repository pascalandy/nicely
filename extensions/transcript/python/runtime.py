"""The run deadline, retries, and the child processes the download starts."""

import logging
import os
import shlex
import signal
import subprocess
import time
from collections.abc import Callable, Sequence
from contextlib import suppress
from dataclasses import dataclass

import httpx

# The skill runner stops the command after 600 seconds. Keep one shared deadline
# with enough margin for Python cleanup and process teardown.
WORKFLOW_TOTAL_TIMEOUT = 570
# A child gets SIGTERM first, then SIGKILL this many seconds later
GRACE = 10.0

log = logging.getLogger("transcript")


class WorkflowTimeoutError(TimeoutError):
    """The bounded workflow has no time left for another external operation."""


@dataclass(frozen=True)
class RunBudget:
    """One monotonic deadline shared by every external operation in a run."""

    deadline: float

    @classmethod
    def start(cls, timeout: float = WORKFLOW_TOTAL_TIMEOUT) -> "RunBudget":
        """Start a deadline from the current monotonic time."""
        return cls(time.monotonic() + timeout)

    def remaining(self, operation: str, maximum: float | None = None) -> float:
        """Return a positive timeout or fail before starting an operation."""
        remaining = self.deadline - time.monotonic()
        if remaining <= 0:
            raise WorkflowTimeoutError(
                f"Workflow time budget exhausted before {operation}"
            )
        return min(remaining, maximum) if maximum is not None else remaining


def run_child(
    command: Sequence[str], *, timeout: float | None = None
) -> subprocess.CompletedProcess[str]:
    """subprocess.run that stops its child gently.

    The child leads its own process group, so what it starts, such as the
    ffmpeg that yt-dlp runs, stops with it. On a timeout or an interrupt, the
    child gets SIGTERM, so it can clean up, and SIGKILL only GRACE seconds
    later.
    """
    log.debug("run %s", _describe(command))
    started = time.monotonic()
    with subprocess.Popen(
        list(command),
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
        process_group=0,
    ) as process:
        try:
            stdout, stderr = process.communicate(timeout=timeout)
        except BaseException:
            stop_child(process)
            raise
    log.debug(
        "%s exited %s after %.1fs",
        command[0],
        process.returncode,
        time.monotonic() - started,
    )
    return subprocess.CompletedProcess(process.args, process.returncode, stdout, stderr)


def stop_child(process: subprocess.Popen[str]) -> None:
    """SIGTERM a child and the process group it leads, then SIGKILL after
    GRACE seconds. A descendant may still hold the pipes, so stop reading
    them GRACE seconds later."""
    send_signal(process, signal.SIGTERM)
    try:
        process.communicate(timeout=GRACE)
        return
    except subprocess.TimeoutExpired:
        send_signal(process, signal.SIGKILL)
    with suppress(subprocess.TimeoutExpired):
        process.communicate(timeout=GRACE)


def send_signal(process: subprocess.Popen[str], number: int) -> None:
    """Send `number` to the process group the child leads."""
    with suppress(ProcessLookupError):
        os.killpg(process.pid, number)


def _describe(command: Sequence[str]) -> str:
    """A command for debug output, with long arguments shortened."""
    return shlex.join(
        argument if len(argument) <= 80 else f"{argument[:77]}..."
        for argument in command
    )


def retry_request[T](
    func: Callable[[], T],
    max_attempts: int = 3,
    initial_delay: float = 1.0,
    deadline: float | None = None,
) -> T:
    """Retry only known transient failures within an optional monotonic deadline.

    Args:
        func: Zero-argument callable to execute.
        max_attempts: Maximum number of tries (must be >= 1).
        initial_delay: Seconds to wait before the first retry.

    Returns:
        The return value of *func* on the first successful call.

    Raises:
        Exception: Re-raises the last exception after all attempts are exhausted.
    """
    delay = initial_delay

    for attempt in range(1, max_attempts + 1):
        try:
            if attempt > 1:
                log.info(f"Retry {attempt}/{max_attempts}...")
            return func()
        except Exception as error:
            if not is_transient_error(error):
                raise
            if attempt == max_attempts:
                log.info(f"Failed after {max_attempts} attempts")
                raise
            if deadline is not None and time.monotonic() + delay >= deadline:
                raise TimeoutError("retry deadline exhausted") from error
            log.info(f"Failed with {type(error).__name__}, retrying in {delay}s...")
            time.sleep(delay)
            delay *= 2  # actual exponential backoff

    # Unreachable when max_attempts >= 1, but satisfies the type checker
    raise RuntimeError("retry_request called with max_attempts < 1")


def is_transient_error(error: Exception) -> bool:
    """Return whether repeating an idempotent request may reasonably succeed."""
    if isinstance(error, (httpx.TransportError, subprocess.TimeoutExpired)):
        return True
    if isinstance(error, httpx.HTTPStatusError):
        return error.response.status_code in {408, 429, 500, 502, 503, 504}
    return False
