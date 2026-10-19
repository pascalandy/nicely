# /// script
# requires-python = ">=3.12"
# dependencies = [
#     "httpx",
#     "yt-dlp==2026.7.4",
#     "pycryptodomex",
# ]
# ///
"""Transcribe the audio of one YouTube video with Deepgram.

uv runs this script with the dependencies of its header, which hold every
runtime pin of the engine. Its sibling modules download the audio with
yt-dlp, send it to Deepgram, and save the transcript files.
"""
