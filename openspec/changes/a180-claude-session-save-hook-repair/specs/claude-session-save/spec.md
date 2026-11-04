## ADDED Requirements

### Requirement: Claude hooks locate their scripts from any working directory
The repository's Claude hook configuration SHALL invoke `save-session.sh` and `tools/sdd-history/agent_save_hook.py` through a path rooted at `$CLAUDE_PROJECT_DIR`, and SHALL invoke `save-session.sh` through `bash` so that a checkout without the executable bit still runs it.

#### Scenario: The session working directory is not the repository root
- **WHEN** a hook fires while the working directory is a subdirectory or another path
- **THEN** the hook command resolves the script from `$CLAUDE_PROJECT_DIR` and does not depend on the working directory

#### Scenario: The checkout has no executable bit
- **WHEN** `save-session.sh` is checked out with mode `100644`
- **THEN** the hook still runs the script through `bash`

### Requirement: Tool-call saving does not block and is bounded
The `PostToolUse` session save hook SHALL run asynchronously with the `--hook` argument. With `--hook`, the saver SHALL exit without writing when another save holds the lock or when the existing summary is younger than 300 seconds.

#### Scenario: A fresh summary exists
- **WHEN** the saver runs with `--hook` and the summary is younger than 300 seconds
- **THEN** the saver exits successfully and the summary is unchanged

#### Scenario: Another save is running
- **WHEN** the saver runs with `--hook` while another process holds the lock
- **THEN** the saver exits successfully without waiting and the summary is unchanged

#### Scenario: The summary is stale
- **WHEN** the saver runs with `--hook` and the summary is older than 300 seconds
- **THEN** the saver writes a new summary and keeps the previous one as a backup

### Requirement: Turn-end and manual saves are not dropped silently
The repository SHALL configure an asynchronous `Stop` hook that runs the saver without `--hook`. A saver run without `--hook` SHALL ignore the interval limit, SHALL wait up to 120 seconds for a running save, and SHALL exit with a nonzero status and write no summary when it cannot take the lock.

#### Scenario: A manual run meets a running save
- **WHEN** the saver runs without arguments while another process holds the lock and that process releases it within the wait
- **THEN** the saver waits and then writes a new summary

#### Scenario: A manual run follows a recent save
- **WHEN** the saver runs without arguments and the summary is younger than 300 seconds
- **THEN** the saver writes a new summary

#### Scenario: The lock never becomes free
- **WHEN** the saver runs without arguments and cannot take the lock within the wait
- **THEN** the saver leaves the summary unchanged and exits with a nonzero status

### Requirement: Save failures are recorded
A lock timeout, a temporary file creation failure, a write failure, or a replacement failure SHALL be reported on standard error and, when the directory is writable, appended to `.ai-context/save-failures.log`, because the hook commands discard standard error.

#### Scenario: A lock timeout occurs
- **WHEN** the saver cannot take the lock within the wait
- **THEN** the failure appears on standard error and in the failure log

### Requirement: Saving does not take the git index lock
The saver SHALL NOT rewrite the git index. Its `git status` and `git diff` calls SHALL run with index refresh disabled, so that a save running in the background cannot make a concurrent `git add` or `git commit` fail on `index.lock`.

#### Scenario: A tracked file is stat-dirty
- **WHEN** a tracked file has a changed modification time and unchanged content, and the saver runs
- **THEN** the bytes of `.git/index` are unchanged

### Requirement: Summary replacement is atomic
The saver SHALL write the new summary to a temporary file in the same directory and SHALL replace the summary only after the write succeeds. On a write failure, the saver SHALL keep the existing summary, remove the temporary file, and exit with a nonzero status. The saver SHALL rotate backups only after the current backup succeeds, and the temporary file name SHALL NOT match the backup pattern.

#### Scenario: The write fails
- **WHEN** writing the new summary fails because of a full disk or a file size limit
- **THEN** the existing summary and its manual memo are unchanged and the saver exits with a nonzero status

#### Scenario: The directory is not writable
- **WHEN** the temporary file cannot be created
- **THEN** the existing summary is unchanged and the saver exits with a nonzero status

#### Scenario: The save succeeds
- **WHEN** the saver finishes successfully
- **THEN** no temporary file remains

### Requirement: The summary stays usable
The saver SHALL limit the `git status` section to 200 lines and SHALL state how many lines it omitted. The saver SHALL NOT walk `openspec/changes/archive` when it lists recently modified files.

#### Scenario: Untracked files exceed the limit
- **WHEN** `git status --short` prints more than 200 lines
- **THEN** the summary holds the first 200 lines and the count of omitted lines

#### Scenario: A file under the archive was modified recently
- **WHEN** a file under `openspec/changes/archive` is younger than 30 minutes
- **THEN** the recent-files section lists neither that file nor the archive directory

### Requirement: Manual notes survive saves
The saver SHALL copy the existing body of the manual memo section into each new summary, and SHALL write the guidance text only when that body is empty.

#### Scenario: A person wrote a note
- **WHEN** the manual memo section holds a note and the saver runs twice
- **THEN** the latest summary holds that note exactly once

#### Scenario: No note exists
- **WHEN** no summary exists or its manual memo section is empty
- **THEN** the new summary holds the guidance text in that section

#### Scenario: A note line resembles a section heading
- **WHEN** a note holds a line that begins like the next section's heading but is not identical to it
- **THEN** the note is kept whole
