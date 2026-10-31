<!-- SPDX-License-Identifier: Apache-2.0 -->
# Instructions

Read `README.md` first.

The coding standard is the `ai-coding-sop` repository at
`https://github.com/filmil/ai-coding-sop`.
Its `AGENTS.md`, its `prose-readability` skill, its `git-commit-rules`
skill and its `hermetic`, `bazel` and `bazel-bzl-files` fragments apply
here in full.
Where this file and ai-coding-sop differ, ai-coding-sop wins; ask the
operator about a conflict that matters.

# Agent operating rules

A change is complete only when `bazel test //...` passes, the integration
module passes (`cd integration && bazel test //...`), and new tests cover
the change.

1. Edit text sources only: SKiDL circuit scripts (`.py`), `BUILD.bazel`,
   `MODULE.bazel`, `.bzl`, Go sources, the locked requirements file. Do
   not commit generated outputs, except the docs that
   `bazel run //docs:update_docs` writes.
2. Keep every action hermetic. No host tools, no host files outside the
   workspace, no host environment variables, no network in actions. Add
   each external file with `http_archive` or `http_file` and a pinned
   `sha256`. `--action_env=PATH`, `use_default_shell_env = True` and the
   tags `local` and `no-sandbox` are forbidden. If a tool cannot run
   hermetically, stop and ask; do not make it work with a host install.
3. Find targets with `bazel query` and `bazel cquery`. Run
   `bazel run //tools:buildifier` on every change.
4. Once there are golden files, update them only with
   `bazel run //:update_goldens`, and give the reason in the commit
   message.
5. A change to a circuit lists the nets it adds, removes or changes in
   the pull request. A circuit written as code has no schematic, so it has
   no schematic renders to attach (issue #12). A change to a board in this
   repository (a `.kicad_pcb` fixture or an integration board) attaches
   the before and after `kicad_render` PNGs of hdlf_rules_kicad to the pull
   request, as hdlf_rules_kicad's own rule 5 asks; put them on a branch
   that is never merged and link them, and delete the branch after the
   merge.
6. One logical change per pull request: one rule, one feature, or one
   design change.
7. Stop and ask the operator before you:
   * delete a test, or make a tolerance larger;
   * change a toolchain version or a library version;
   * change a net or a part that has the `critical` tag;
   * run a release or publish workflow, upload a mirror release, or set
     a secret.
8. When a test fails and the cause is not clear after three attempts, stop.
   Write the failure and your hypotheses in the pull request.

# Rules particular to this repository

* **Tools come from toolchains only.** Every action runs its tool through
  `run_tool` in `//netlist/private` (issue #4), with the tool from
  `ctx.toolchains`.
  No private attribute names an external tool.
* **Public packages name no dev dependency.** A user of the module loads
  `//netlist` and `//`, where `@stardoc`, `@bazel_lib`, `@buildifier_prebuilt`
  and `@hermetic_cc_toolchain` are not visible. Documentation targets live in
  `//docs` and development tools in `//tools`.
* **Documentation is generated.** After changing a `.bzl` file, run
  `bazel run //docs:update_docs`; `bazel test //...` fails until then.
* **Linux x86_64 only** in v1.
* **Python only where the CAD library needs it.** Circuit scripts and the
  SKiDL runner are Python; every checker and helper is Go.
* **No scripts that nothing builds.** A helper is a Bazel tool, Go with a
  `go_test`, under `tools/` (as `tools/nooverrides` and `tools/mirror`), or
  a documented command; `.github/workflows/release_prep.sh` is the
  template's one exception.
* **Wheels are pinned.** Every Python package comes from
  `requirements_lock.txt`, by hash, binary wheels only. A package that
  PyPI has only as an sdist is built once, reproducibly, and mirrored
  (`tools/mirror/README.md`); the mirror test keeps the manifest and the
  lock equal. A new package or version is rule 7.
* **Check by exit code.** A local check passes when its command exits 0:
  run `bazel run //tools:buildifier_check` and `bazel run //tools/lint`
  and look at the status, not at the words they print.
* **Commits for a BCR pull request carry no `Co-Authored-By` trailer**
  (the `bazel-bcr-publish` skill says why).

`CLAUDE.md` and `GEMINI.md` are symlinks to this file.
