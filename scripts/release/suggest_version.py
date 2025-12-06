#!/usr/bin/env python3
"""
Suggest the next version (minor vs patch) using OpenAI, defaulting to a minor bump unless the changes look like small fixes.

Usage:
  python scripts/release/suggest_version.py [--prev-tag TAG]

Requires OPENAI_API_KEY.
"""

import argparse
import json
import os
import re
import subprocess
import sys
import urllib.error
import urllib.request

MODEL = "gpt-5.1"


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


def parse_version(tag):
    m = re.match(r"v?(\d+)\.(\d+)\.(\d+)", tag or "")
    if not m:
        return (0, 0, 0)
    return tuple(int(x) for x in m.groups())


def bump(version, kind):
    major, minor, patch = version
    if kind == "minor":
        return (major, minor + 1, 0)
    return (major, minor, patch + 1)


def format_version(version):
    return f"v{version[0]}.{version[1]}.{version[2]}"


def collect_commits_stats(commit_range):
    commits = ""
    stats = ""
    try:
        commits = run_git(["log", "--pretty=format:%h %s", commit_range])
        stats = run_git(["diff", "--stat", commit_range])
    except RuntimeError:
        pass
    return commits, stats


def call_openai(prompt):
    api_key = os.environ.get("OPENAI_API_KEY")
    if not api_key:
        raise RuntimeError("OPENAI_API_KEY is not set")
    body = {
        "model": MODEL,
        "messages": [
            {"role": "system", "content": "You choose semver bump: minor vs patch. Default to minor unless changes are very small bugfixes or trivial docs."},
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

    return payload["choices"][0]["message"]["content"]


def pick_choice(text):
    lowered = text.lower()
    if "patch" in lowered and "minor" not in lowered:
        return "patch", text
    if "minor" in lowered:
        return "minor", text
    return "minor", text


def main():
    parser = argparse.ArgumentParser(description="Suggest next version (minor vs patch).")
    parser.add_argument("--prev-tag", default="", help="Previous tag (default: latest tag).")
    args = parser.parse_args()

    prev_tag = args.prev_tag or detect_prev_tag()
    prev_version = parse_version(prev_tag)
    commit_range = f"{prev_tag}..HEAD" if prev_tag else "HEAD"
    commits, stats = collect_commits_stats(commit_range)

    prompt = f"""Given these commits and diff stats, choose minor or patch bump. Default to minor (x.y+1.0) unless changes are only small bugfixes/docs/chore.
Return a short reason and the chosen word.

Commits:
{commits or '(none)'}

Stats:
{stats or '(none)'}
"""
    try:
        response = call_openai(prompt)
        choice, reason = pick_choice(response)
    except Exception as exc:
        choice, reason = "minor", f"LLM unavailable, defaulting to minor: {exc}"

    next_version = bump(prev_version, choice)
    print(json.dumps({"choice": choice, "reason": reason.strip(), "version": format_version(next_version)}))


if __name__ == "__main__":
    main()
