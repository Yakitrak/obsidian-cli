#!/usr/bin/env python3
"""
Generate release notes and a changelog entry using the OpenAI API.

Usage:
  python scripts/release/generate_notes.py [--version vX.Y.Z] [--prev-tag TAG] [--release-out path] [--changelog-out path]

Requires OPENAI_API_KEY in the environment and git available in the current repo.
"""

import argparse
import datetime
import json
import os
import subprocess
import sys
import urllib.error
import urllib.request


MODEL = "gpt-5.1"
DEFAULT_EXCLUDES = [
    ":(exclude)vendor",
    ":(exclude)bin",
    ":(exclude)*.png",
    ":(exclude)*.jpg",
    ":(exclude)*.jpeg",
    ":(exclude)*.gif",
    ":(exclude)*.svg",
    ":(exclude)*.webp",
]


def run_git(args):
    result = subprocess.run(["git"] + args, capture_output=True, text=True)
    if result.returncode != 0:
        raise RuntimeError(result.stderr.strip() or f"git {' '.join(args)} failed")
    return result.stdout.strip()


def detect_prev_tag():
    try:
        return run_git(["describe", "--tags", "--abbrev=0"])
    except RuntimeError:
        return ""


def detect_root_commit():
    return run_git(["rev-list", "--max-parents=0", "HEAD"]).splitlines()[0]


def collect_commits(commit_range):
    try:
        return run_git(["log", "--pretty=format:%h %s", commit_range])
    except RuntimeError:
        return ""


def collect_diff(commit_range):
    try:
        args = ["diff", commit_range, "--", "."]
        args.extend(DEFAULT_EXCLUDES)
        return run_git(args)
    except RuntimeError:
        return ""


def call_openai(prompt):
    api_key = os.environ.get("OPENAI_API_KEY")
    if not api_key:
        raise RuntimeError("OPENAI_API_KEY is not set")

    body = {
        "model": MODEL,
        "messages": [
            {
                "role": "system",
                "content": "You are a release-notes generator. Write concise, user-facing notes.",
            },
            {"role": "user", "content": prompt},
        ],
        "temperature": 0.2,
    }
    data = json.dumps(body).encode("utf-8")
    req = urllib.request.Request(
        "https://api.openai.com/v1/chat/completions",
        data=data,
        headers={
            "Content-Type": "application/json",
            "Authorization": f"Bearer {api_key}",
        },
    )
    try:
        with urllib.request.urlopen(req) as resp:
            payload = json.loads(resp.read().decode("utf-8"))
    except urllib.error.HTTPError as e:
        detail = e.read().decode("utf-8")
        raise RuntimeError(f"OpenAI API error: {e.code} {detail}") from e
    except urllib.error.URLError as e:
        raise RuntimeError(f"OpenAI API request failed: {e}") from e

    try:
        return payload["choices"][0]["message"]["content"]
    except Exception:
        raise RuntimeError(f"Unexpected OpenAI response: {payload}")


def build_prompt(version, prev_tag, commits, diff, extra_hint):
    today = datetime.date.today().isoformat()
    base = prev_tag if prev_tag else "initial release"
    hint_block = f"\nUser feedback: {extra_hint}\n" if extra_hint else ""
    return f"""Generate release documentation for version {version} (compared to {base}).
Return exactly two sections marked with the headers:
===RELEASE_NOTES===
...markdown for GitHub release body. Keep to 5-7 concise bullets. Lead with user-facing benefits. Mention breaking changes/default flips. Skip internal-only details (caches, watcher internals, refactors) unless they directly improve user experience (faster runs, fewer rescans, reliability). Use short bullet phrases.
===CHANGELOG_ENTRY===
...markdown changelog entry suitable for CHANGELOG.md, starting with a heading like "## [{version}] - {today}", then concise bullets. Mirror the release notes focus (user-facing, brief).

Focus on user-facing behavior. Avoid over-promising.

{hint_block}Commits:
{commits or '(none)'}

Diff:
{diff or '(diff unavailable)'}
"""


def parse_output(text):
    release = ""
    changelog = ""
    if "===RELEASE_NOTES===" in text:
        after = text.split("===RELEASE_NOTES===", 1)[1]
        if "===CHANGELOG_ENTRY===" in after:
            release, changelog = after.split("===CHANGELOG_ENTRY===", 1)
        else:
            release = after
    else:
        release = text
    return release.strip(), changelog.strip()


def main():
    parser = argparse.ArgumentParser(description="Generate release notes with OpenAI.")
    parser.add_argument("--version", default="", help="Version label (e.g., v1.2.3).")
    parser.add_argument("--prev-tag", default="", help="Previous tag to diff against (default: latest tag).")
    parser.add_argument("--release-out", default="", help="Path to write release notes markdown.")
    parser.add_argument("--changelog-out", default="", help="Path to write changelog entry markdown.")
    parser.add_argument("--extra", default="", help="Optional extra guidance/feedback for the LLM.")
    args = parser.parse_args()

    prev_tag = args.prev_tag or detect_prev_tag()
    if prev_tag:
        commit_range = f"{prev_tag}..HEAD"
    else:
        root = detect_root_commit()
        commit_range = f"{root}^..HEAD"

    commits = collect_commits(commit_range)
    diff = collect_diff(commit_range)
    version = args.version or "Unreleased"

    prompt = build_prompt(version, prev_tag, commits, diff, args.extra)
    content = call_openai(prompt)
    release_notes, changelog = parse_output(content)

    if args.release_out:
        with open(args.release_out, "w", encoding="utf-8") as f:
            f.write(release_notes + "\n")
    if args.changelog_out:
        with open(args.changelog_out, "w", encoding="utf-8") as f:
            f.write(changelog + "\n")

    if not args.release_out and not args.changelog_out:
        print("===RELEASE_NOTES===")
        print(release_notes)
        print("\n===CHANGELOG_ENTRY===")
        print(changelog)


if __name__ == "__main__":
    try:
        main()
    except Exception as exc:
        sys.stderr.write(f"Error: {exc}\n")
        sys.exit(1)
