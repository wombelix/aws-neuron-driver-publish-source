## v0.1.0 (2025-02-05)

### Feat

- slog logging with INFO and DEBUG flag
- process rpm and update driver src
- git commit changed archive files
- cli args for folder variables
- **rpm**: Skip existing files
- process, download and verify rpms
- Process, Parse and Filter Primary
- process release-notes / changelog
- repomd handling and parsing added

### Fix

- Align spacing in commit msg

### Refactor

- error handling, vars renamed
- Handle each rpm individually, sort version slice
- Checksum verify and write logic to common.go
- Move shared funcs to common.go
