"""Write the goldens with the source program, never with the engine.

The source is `authoring/andy/transcript/scripts/` of pascalandy/skills at
0c9c4a2. Its Deepgram request reads the fixture audio, a mock transport
answers with the fixture response, and its parser and saver write the three
files beside this script. From the engine folder, run:

    uv export --quiet --no-hashes --script <scripts>/transcript.py > <file>
    uv run --no-project --with-requirements <file> \
        python tests/goldens/make_goldens.py <scripts>
"""

import importlib.util
import sys
from pathlib import Path

import httpx

GOLDENS = Path(__file__).parent
FIXTURES = GOLDENS.parent / "fixtures"


def main(scripts: Path) -> None:
    spec = importlib.util.spec_from_file_location(
        "source_transcript", scripts / "transcript.py"
    )
    assert spec is not None and spec.loader is not None
    source = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(source)

    body = (FIXTURES / "deepgram_response.json").read_bytes()
    client = httpx.Client(
        transport=httpx.MockTransport(
            lambda _request: httpx.Response(
                200, content=body, headers={"Content-Type": "application/json"}
            )
        )
    )
    source.httpx.post = client.post
    response = source.transcribe_audio(
        FIXTURES / "audio.webm", "fake-key", source.RunBudget.start()
    )
    plain, sentences, json_data = source.parse_transcript(response)
    source.save_outputs(GOLDENS, plain, sentences, json_data)


if __name__ == "__main__":
    main(Path(sys.argv[1]))
