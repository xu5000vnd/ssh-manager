# SSH Key Management TUI – Requirements and Build Plan

## 1. Tech stack

| Layer | Choice | Rationale |
|-------|--------|-----------|
| **Language** | **Go** | Single static binary for Windows and macOS; no runtime to install; strong stdlib and crypto; fast startup. |
| **TUI framework** | **Bubble Tea** (`github.com/charmbracelet/bubbletea`) | Elm-style model/update/view; works on Windows (including Windows Terminal) and macOS; ecosystem includes Lipgloss and Bubbles. |
| **CLI framework** | **Cobra** (`github.com/spf13/cobra`) | Subcommand routing for CLI mode; integrates cleanly with Bubble Tea for TUI mode. |
| **Encryption** | AES-256-GCM + `golang.org/x/crypto/scrypt` for key derivation | AEAD encryption with password-derived key; no external crypto dependency beyond Go's extended stdlib. |
| **Config format** | JSON (encrypted on disk) | Simple to serialize/deserialize; easy to version and migrate later. |

---

## 2. Data model

### Connection struct

```go
type Connection struct {
    ID              string `json:"id"`
    Name            string `json:"name"`
    Group           string `json:"group,omitempty"`
    Host            string `json:"host"`
    Port            int    `json:"port"`                   // default 22
    Username        string `json:"username"`
    KeyPath         string `json:"key_path"`               // path to private key file
    ProxyJump       string `json:"proxy_jump,omitempty"`   // jump host, e.g. "bastion_user@bastion_host" or connection ID
    ExtraArgs       string `json:"extra_args,omitempty"`   // additional ssh flags, e.g. "-L 8080:localhost:80"
    Tags            string `json:"tags,omitempty"`         // comma-separated tags
    Notes           string `json:"notes,omitempty"`        // freeform description
    Favorite        bool   `json:"favorite"`
    ConnectCount    int    `json:"connect_count"`
    LastConnectedAt string `json:"last_connected_at,omitempty"` // RFC3339
}
```

### Config (what gets encrypted/decrypted)

```go
type Config struct {
    Version     int          `json:"version"`     // schema version, start at 1
    Connections []Connection `json:"connections"`
}
```

### On-disk format

- Config path: `~/.config/ssh-manager/config.enc` (macOS/Linux), `%APPDATA%\ssh-manager\config.enc` (Windows).
- File contents: `salt (32 bytes) || nonce (12 bytes) || AES-256-GCM ciphertext`.
- Backup rotation: keep last 3 backups as `config.enc.1`, `config.enc.2`, `config.enc.3`; rotate on every save.
- Decrypted config is **never** written to disk; only held in process memory.

---

## 3. Modes of operation

### 3.1 TUI mode (default)

Run without subcommands to launch the interactive terminal UI:

```
ssh-manager
```

### 3.2 CLI mode (non-interactive)

Subcommands for scripting and quick access without entering the TUI:

```
ssh-manager connect <name>               # connect directly by name (prompts password, then exec ssh)
ssh-manager list                          # print all connections as a table to stdout
ssh-manager list --json                   # print as JSON to stdout
ssh-manager add --name x --host y ...     # add a connection non-interactively
ssh-manager delete <name>                 # delete by name (with confirmation)
ssh-manager export -o file.json           # export to JSON file
ssh-manager import file.json              # import from JSON file
ssh-manager import-ssh-config             # import from ~/.ssh/config
ssh-manager export-ssh-config -o file     # export as SSH config format
ssh-manager password                      # change master password
```

All CLI subcommands prompt for the master password (or accept it via `SSH_MANAGER_PASSWORD` env var for automation).

---

## 4. Screens and navigation (TUI)

### Screen flow

```
Unlock (password) --> Home (connection list)
Home --> Connection Detail
Home --> Add/Edit Connection
Home --> Import (JSON or SSH config)
Home --> Export (JSON or SSH config)
Home --> Change Password
Connection Detail --> Add/Edit Connection
Connection Detail --> Connect (exec ssh)
Add/Edit Connection --> Home
```

### Screen details

1. **Unlock screen** (app start)
   - Password input (masked).
   - Derive encryption key via scrypt, decrypt `config.enc`.
   - If file does not exist, treat as first run: create empty config, ask user to set a password (confirm twice).
   - On wrong password, show error and retry; `q` to quit.

2. **Home (connection list)**
   - Table columns: Fav, Name, Group, Host, Username, Port.
   - Default sort: favorites first, then by most recently connected.
   - Toggle grouped view with `g` (collapse/expand by Group field).
   - Keybindings:
     - `Up/Down` navigate, `Enter` detail, `a` add, `e` edit, `d` delete
     - `c` connect, `f` toggle favorite
     - `Space` multi-select, then `d` bulk delete / `E` export selected / `t` bulk tag
     - `/` fuzzy search (filters across name, host, username, group, tags)
     - `g` toggle grouped view
     - `E` export, `I` import, `S` import SSH config
     - `P` change master password
     - `T` test connection (TCP dial)
     - `y` copy SSH command to clipboard
     - `q` quit
   - Footer: connection count, active filter, keybindings help.

3. **Connection detail**
   - Show all fields including group, proxy jump, extra args, notes, last connected, connect count.
   - Keybindings: `e` edit, `c` connect, `d` delete, `f` toggle favorite, `y` copy SSH command, `T` test, `Esc` back.

4. **Add/Edit connection**
   - Form with fields: Name, Group, Host, Port (default 22), Username, Key Path, Proxy Jump, Extra Args, Tags, Notes.
   - Validate: Name and Host are required; Port must be numeric 1-65535; Key Path should exist on disk (warn if not).
   - Tab/Shift-Tab to move between fields; `Ctrl+S` to save, `Esc` to cancel.
   - On save: update in-memory config, auto-persist (encrypt and write to disk).

5. **Delete confirmation**
   - Modal/overlay: "Delete connection <name>? [y/N]" (or "Delete N selected connections? [y/N]" for bulk).
   - On confirm: remove from list, persist.

6. **Connect**
   - Build command: `ssh -J <proxy_jump> -i <key_path> <username>@<host> -p <port> <extra_args>`.
   - Omit `-J` if proxy_jump is empty; omit extra_args if empty.
   - Update `connect_count` and `last_connected_at` on the connection, persist.
   - Exit the TUI and `exec` the ssh command so it takes over the terminal.

7. **Change password screen**
   - Prompt: current password (verify), new password, confirm new password.
   - Decrypt with old, re-encrypt with new, write to disk.
   - Show success and return to Home.

### 4.1 Import and export

#### JSON import/export

- **Export (`E` from Home):**
  - Prompt for file path (default: `./ssh-connections.json`).
  - If multi-select is active, export only selected connections; otherwise export all.
  - Write as plain JSON: `{ "version": 1, "connections": [...] }`.
  - Show warning: "Exported file is NOT encrypted. Store securely."
  - Show success message with file path.

- **Import (`I` from Home):**
  - Prompt for file path.
  - Read and parse JSON; validate schema and version field.
  - Detect duplicates by name AND by (host, port, username) tuple.
  - Show summary: "Found N connections (M duplicates detected)."
  - Offer choice: **Merge** (append new, skip duplicates) | **Replace** (overwrite all) | **Cancel**.
  - On confirm: update in-memory config, persist encrypted.
  - Show result: "Imported N connections."

#### SSH config import/export

- **Import SSH config (`S` from Home):**
  - Default path: `~/.ssh/config`; allow override.
  - Parse Host blocks: extract Host, HostName, User, Port, IdentityFile, ProxyJump.
  - Map to Connection struct (Name = Host alias, Host = HostName, etc.).
  - Skip wildcard entries (`Host *`).
  - Same merge/replace flow as JSON import.

- **Export SSH config:**
  - Available via CLI: `ssh-manager export-ssh-config -o file`.
  - Generate standard `~/.ssh/config` format with Host blocks.
  - Warn that exported file is plaintext.

---

## 5. Data flow

1. **Startup:** Read `config.enc` from disk (if exists) -> prompt password -> derive key (scrypt) -> decrypt (AES-256-GCM) -> load `Config` into memory.
2. **Runtime:** All CRUD operations modify the in-memory `Config`.
3. **Persist:** After every write operation (add/edit/delete/import/connect), serialize Config to JSON -> encrypt -> atomic write to `config.enc` (write temp, rename). Rotate backups.
4. **Connect:** Read connection params from memory -> build ssh command -> update connect_count and last_connected_at -> persist -> exit TUI -> exec ssh.
5. **Export:** Serialize in-memory Config (or selected subset) to plain JSON or SSH config format -> write to user-specified path.
6. **Import:** Read plain JSON or SSH config from path -> validate and parse -> detect duplicates -> merge or replace in-memory Config -> persist encrypted.
7. **Auto-lock:** After N minutes of inactivity (configurable, default 5), clear decryption key from memory -> show Unlock screen. Timer resets on any keypress.

---

## 6. Security

- **Key derivation:** scrypt with N=32768, r=8, p=1; 32-byte derived key; random 32-byte salt per encryption.
- **Encryption:** AES-256-GCM; random 12-byte nonce per encryption; nonce stored alongside ciphertext.
- **Atomic writes:** Write encrypted config to a temp file first, then rename, to prevent corruption on crash.
- **Backup rotation:** Keep last 3 encrypted backups on every save.
- **Memory:** Do not log connection details or passwords. Clear password/key from memory after key derivation and on lock.
- **Auto-lock:** Clear decryption key after configurable idle timeout; require password re-entry.
- **Key paths:** Store as-is (e.g. `~/.ssh/id_ed25519`); resolve `~` at connect time.
- **Export warning:** Always warn that exported JSON/SSH config is plaintext.
- **CLI password:** Accept via env var `SSH_MANAGER_PASSWORD` for scripting; warn in docs that env vars may be visible in process lists.

---

## 7. Project structure

```
ssh-manager/
  main.go                       # Entry point; Cobra root command setup
  cmd/
    root.go                     # Root command: no subcommand -> launch TUI
    connect.go                  # "connect <name>" subcommand
    list.go                     # "list" subcommand
    add.go                      # "add" subcommand
    delete.go                   # "delete <name>" subcommand
    export.go                   # "export" subcommand (JSON)
    import.go                   # "import" subcommand (JSON)
    import_ssh.go               # "import-ssh-config" subcommand
    export_ssh.go               # "export-ssh-config" subcommand
    password.go                 # "password" subcommand (change master password)
  internal/
    app/
      model.go                  # Main Bubble Tea model (app state, current screen, idle timer)
      update.go                 # Update function: handle messages, key events, screen transitions
      view.go                   # View function: render current screen
      screens/
        unlock.go               # Unlock/password screen
        home.go                 # Connection list with fuzzy search, grouped view, multi-select
        detail.go               # Connection detail screen
        form.go                 # Add/Edit connection form
        import.go               # Import screen (path input, duplicate summary, merge/replace choice)
        export.go               # Export screen (path input, success/warning)
        password.go             # Change master password screen
    config/
      config.go                 # Config and Connection structs, serialize/deserialize
      storage.go                # Config file path resolution (per OS), atomic read/write, backup rotation
      import_export.go          # Export to plain JSON, import from plain JSON (validate, deduplicate, merge/replace)
      sshconfig.go              # Parse and generate ~/.ssh/config format
    crypto/
      encrypt.go                # Encrypt and decrypt byte slices with AES-256-GCM
      derive.go                 # Key derivation with scrypt
    clipboard/
      clipboard.go              # Copy to clipboard (pbcopy on macOS, clip.exe on Windows)
    sshcmd/
      builder.go                # Build ssh command string from Connection (handles proxy_jump, extra_args)
      exec.go                   # Exec ssh, replacing current process
  go.mod
  go.sum
```

---

## 8. Implementation order

Each step should be completed and tested before moving to the next.

### Phase 1: Core

- [ ] **1. Project init**
  - [ ] Run `go mod init ssh-manager`
  - [ ] Add all dependencies to `go.mod` (`bubbletea`, `bubbles`, `lipgloss`, `cobra`, `fuzzy`, `uuid`, `x/crypto`)
  - [ ] Create directory structure: `cmd/`, `internal/app/screens/`, `internal/config/`, `internal/crypto/`, `internal/clipboard/`, `internal/sshcmd/`

- [ ] **2. Crypto layer** (`internal/crypto/`)
  - [ ] Implement `derive.go`: `DeriveKey(password []byte, salt []byte) ([]byte, error)` using scrypt (N=32768, r=8, p=1, keyLen=32)
  - [ ] Implement `encrypt.go`: `Encrypt(plaintext []byte, password []byte) ([]byte, error)` — generate random salt (32B) + nonce (12B), derive key, AES-256-GCM seal, return `salt || nonce || ciphertext`
  - [ ] Implement `encrypt.go`: `Decrypt(data []byte, password []byte) ([]byte, error)` — split salt/nonce/ciphertext, derive key, AES-256-GCM open
  - [ ] Write unit tests: encrypt then decrypt roundtrip, wrong password fails, corrupted data fails, empty input

- [ ] **3. Config and storage** (`internal/config/`)
  - [ ] Define `Connection` struct with all JSON fields (ID, Name, Group, Host, Port, Username, KeyPath, ProxyJump, ExtraArgs, Tags, Notes, Favorite, ConnectCount, LastConnectedAt)
  - [ ] Define `Config` struct (Version int, Connections []Connection)
  - [ ] Implement `storage.go`: `ConfigDir() string` — use `os.UserConfigDir()`, append `/ssh-manager/`; create dir if missing
  - [ ] Implement `storage.go`: `ConfigPath() string` — returns full path to `config.enc`
  - [ ] Implement `config.go`: `Load(path string, password []byte) (*Config, error)` — read file, call `crypto.Decrypt`, unmarshal JSON; return empty Config if file not found
  - [ ] Implement `config.go`: `Save(cfg *Config, path string, password []byte) error` — marshal JSON, call `crypto.Encrypt`, atomic write (write temp file, rename)
  - [ ] Implement backup rotation in `Save`: before writing, rotate existing `config.enc` -> `config.enc.1` -> `.2` -> `.3`; delete `.3` if exists
  - [ ] Write unit tests: save then load roundtrip, first-run (no file), atomic write does not corrupt on partial failure, backup files are created

- [ ] **4. SSH command builder** (`internal/sshcmd/`)
  - [ ] Implement `builder.go`: `BuildCommand(conn Connection) []string` — returns args slice: `["ssh", "-i", keyPath, "-p", port, ...]`; include `-J proxy_jump` if set; append extra_args if set; resolve `~` in key path
  - [ ] Implement `exec.go`: `Exec(args []string) error` — use `syscall.Exec` (Unix) or `os/exec` (Windows) to replace process with ssh
  - [ ] Write unit tests for `BuildCommand`: basic connection, with proxy jump, with extra args, with default port, tilde expansion

### Phase 2: TUI basics

- [ ] **5. Cobra root command** (`main.go`, `cmd/root.go`)
  - [ ] Set up Cobra root command in `cmd/root.go`
  - [ ] `main.go`: call `cmd.Execute()`
  - [ ] Root command `Run`: launch Bubble Tea program (TUI mode when no subcommand)

- [ ] **6. Bubble Tea model and screen framework** (`internal/app/model.go`)
  - [ ] Define `Screen` enum: `ScreenUnlock`, `ScreenHome`, `ScreenDetail`, `ScreenForm`, `ScreenImport`, `ScreenExport`, `ScreenPassword`
  - [ ] Define main `Model` struct: current screen, config, decrypted key, selected connection index, form state, filter string, idle timer, multi-select set, status message
  - [ ] Implement `Init()`, `Update()`, `View()` that delegate to per-screen handlers
  - [ ] Set up idle timer: reset on every keypress; fire `autoLockMsg` after timeout

- [ ] **7. Unlock screen** (`internal/app/screens/unlock.go`)
  - [ ] Render: app title/banner, password input (masked with `*`), status message area
  - [ ] On Enter: call `config.Load`; on success -> transition to Home; on error -> show "Wrong password" and stay
  - [ ] First-run detection: if `config.enc` does not exist, switch to "Set password" mode (enter password, confirm, create empty config)
  - [ ] `q` to quit the app

- [ ] **8. Home list screen** (`internal/app/screens/home.go`)
  - [ ] Render connection table using `lipgloss`: columns Fav(*), Name, Group, Host, User, Port
  - [ ] Cursor navigation: Up/Down/j/k to move, highlight selected row
  - [ ] Fuzzy search: `/` enters filter mode; type to filter across name, host, username, group, tags; `Esc` clears filter
  - [ ] Grouped view: `g` toggles; when grouped, show group headers, collapse/expand
  - [ ] Sorting: favorites first, then by `last_connected_at` descending, then by name ascending
  - [ ] Multi-select: `Space` toggles selection on current row; show selection count in footer
  - [ ] Footer: connection count (filtered/total), active filter text, keybindings help line
  - [ ] Keybinding dispatch: `Enter` -> Detail, `a` -> Form (add), `e` -> Form (edit), `d` -> Delete, `c` -> Connect, `f` -> toggle favorite, `y` -> copy SSH command, `T` -> health check, `E` -> Export, `I` -> Import JSON, `S` -> Import SSH config, `P` -> Change password, `q` -> Quit

- [ ] **9. Connection detail screen** (`internal/app/screens/detail.go`)
  - [ ] Render all Connection fields as labeled rows (Name, Group, Host, Port, Username, Key Path, Proxy Jump, Extra Args, Tags, Notes, Favorite, Connect Count, Last Connected)
  - [ ] Keybindings: `e` edit, `c` connect, `d` delete, `f` toggle favorite, `y` copy SSH command, `T` health check, `Esc` back to Home

- [ ] **10. Add/Edit form screen** (`internal/app/screens/form.go`)
  - [ ] Create text inputs for each field: Name, Group, Host, Port, Username, Key Path, Proxy Jump, Extra Args, Tags, Notes
  - [ ] Tab/Shift-Tab to move between fields; active field highlighted
  - [ ] Pre-fill fields when editing an existing connection
  - [ ] Validation on save: Name required, Host required, Port 1-65535, warn if Key Path does not exist on disk
  - [ ] `Ctrl+S` to save: generate UUID for new connections; update in-memory config; call `config.Save`; transition to Home with success message
  - [ ] `Esc` to cancel: discard changes, return to Home

### Phase 3: Actions

- [ ] **11. Delete with confirmation** (overlay in `home.go` or `detail.go`)
  - [ ] Single delete: `d` on Home or Detail -> show "Delete <name>? [y/N]"
  - [ ] Bulk delete: `d` with multi-select active -> show "Delete N connections? [y/N]"
  - [ ] On `y`: remove from config, persist, refresh Home, show "Deleted" message
  - [ ] On `n`/`Esc`: cancel, return to previous state

- [ ] **12. Connect (exec ssh)**
  - [ ] `c` from Home or Detail: get selected Connection
  - [ ] Call `sshcmd.BuildCommand(conn)`
  - [ ] Update `conn.ConnectCount++` and `conn.LastConnectedAt = time.Now().Format(RFC3339)`
  - [ ] Persist config
  - [ ] Send `tea.Quit` message, then in `main.go` after `tea.Program` exits, call `sshcmd.Exec(args)`

- [ ] **13. Favorites and sorting**
  - [ ] `f` from Home or Detail: toggle `conn.Favorite`, persist
  - [ ] Home sort function: favorites first, then by `LastConnectedAt` descending (empty last), then by Name ascending
  - [ ] Show `*` or star indicator in the Fav column for favorited connections

### Phase 4: Import/Export

- [ ] **14. Export to JSON** (`internal/config/import_export.go`, `internal/app/screens/export.go`)
  - [ ] `ExportJSON(cfg *Config, path string) error` — marshal Config to indented JSON, write to file
  - [ ] `ExportSelectedJSON(connections []Connection, path string) error` — export subset
  - [ ] TUI screen: prompt for file path (default `./ssh-connections.json`), if multi-select active export selected only
  - [ ] Show warning: "Exported file is NOT encrypted"
  - [ ] Show success: "Exported N connections to <path>"

- [ ] **15. Import from JSON** (`internal/config/import_export.go`, `internal/app/screens/import.go`)
  - [ ] `ImportJSON(path string) (*Config, error)` — read file, unmarshal, validate version field
  - [ ] `DetectDuplicates(existing, incoming []Connection) (new, duplicates []Connection)` — match by Name and by (Host, Port, Username) tuple
  - [ ] TUI screen: prompt for file path, show "Found N connections (M duplicates)"
  - [ ] Choice: Merge (append non-duplicates) | Replace (overwrite all) | Cancel
  - [ ] On confirm: update in-memory config, persist, show "Imported N connections"

- [ ] **16. Import from SSH config** (`internal/config/sshconfig.go`)
  - [ ] `ParseSSHConfig(path string) ([]Connection, error)` — parse Host blocks; extract Host (alias -> Name), HostName (-> Host), User, Port, IdentityFile (-> KeyPath), ProxyJump; skip `Host *` and wildcard patterns
  - [ ] TUI screen (`S` from Home): prompt for path (default `~/.ssh/config`), show parsed count, same merge/replace flow
  - [ ] Handle missing fields gracefully (default port 22, empty user, etc.)

- [ ] **17. Export to SSH config format** (CLI only for v1)
  - [ ] `GenerateSSHConfig(connections []Connection) string` — generate `Host` blocks with HostName, User, Port, IdentityFile, ProxyJump
  - [ ] CLI: `ssh-manager export-ssh-config -o <path>` — write generated config to file, warn plaintext

### Phase 5: Extras

- [ ] **18. Change master password** (`internal/app/screens/password.go`, `cmd/password.go`)
  - [ ] TUI screen: input current password -> verify by decrypting -> input new password -> confirm new password -> re-encrypt config with new password -> save
  - [ ] Show error if current password wrong; show error if new passwords don't match
  - [ ] CLI: `ssh-manager password` — same flow in terminal prompts (no TUI)

- [ ] **19. Copy to clipboard** (`internal/clipboard/`)
  - [ ] Implement `Copy(text string) error` — detect OS: exec `pbcopy` on macOS, `clip.exe` on Windows, `xclip`/`xsel` on Linux
  - [ ] `y` from Home: build SSH command for selected connection, copy to clipboard, show "Copied to clipboard" message
  - [ ] `y` from Detail: same behavior
  - [ ] Handle errors gracefully (clipboard tool not found)

- [ ] **20. Connection health check**
  - [ ] Implement `CheckHealth(host string, port int, timeout time.Duration) (bool, error)` — `net.DialTimeout("tcp", host:port, timeout)`
  - [ ] `T` from Home: check selected connection, show green "Reachable" or red "Unreachable" inline
  - [ ] `T` from Detail: same, show result in status area
  - [ ] Use 5-second timeout; show "Checking..." while in progress (run as `tea.Cmd` so TUI stays responsive)

- [ ] **21. Auto-lock on idle**
  - [ ] Add `idleTimeout` to model (default 5 min)
  - [ ] Start/reset a `tea.Tick` on every keypress
  - [ ] On tick fire with no intervening keypress: clear encryption key from memory, transition to Unlock screen
  - [ ] Show "Session locked due to inactivity" on Unlock screen

### Phase 6: CLI subcommands

- [ ] **22. CLI: `connect <name>`** (`cmd/connect.go`)
  - [ ] Prompt for password (or read `SSH_MANAGER_PASSWORD` env var)
  - [ ] Load config, find connection by name (case-insensitive), update stats, persist, exec ssh
  - [ ] Error if name not found; suggest closest match

- [ ] **23. CLI: `list`** (`cmd/list.go`)
  - [ ] Prompt for password, load config
  - [ ] Print connections as formatted table to stdout (Name, Group, Host, User, Port)
  - [ ] `--json` flag: print as JSON array
  - [ ] `--group <name>` flag: filter by group

- [ ] **24. CLI: `add`** (`cmd/add.go`)
  - [ ] Flags: `--name`, `--host`, `--port`, `--user`, `--key`, `--group`, `--proxy-jump`, `--extra-args`, `--tags`, `--notes`
  - [ ] Prompt for password, load config, validate, add connection, persist
  - [ ] Print "Added connection <name>"

- [ ] **25. CLI: `delete <name>`** (`cmd/delete.go`)
  - [ ] Prompt for password, load config, find by name
  - [ ] Confirm: "Delete <name>? [y/N]"
  - [ ] Delete, persist, print "Deleted connection <name>"

- [ ] **26. CLI: `export` and `import`** (`cmd/export.go`, `cmd/import.go`)
  - [ ] `export -o <path>`: prompt password, load, serialize JSON, write, warn plaintext
  - [ ] `import <path>`: prompt password, load, parse JSON, show count, confirm merge/replace, persist

- [ ] **27. CLI: `import-ssh-config` and `export-ssh-config`** (`cmd/import_ssh.go`, `cmd/export_ssh.go`)
  - [ ] `import-ssh-config [path]`: default `~/.ssh/config`, parse, merge/replace flow
  - [ ] `export-ssh-config -o <path>`: generate SSH config format, write, warn plaintext

### Phase 7: Polish

- [ ] **28. Keybindings help bar**
  - [ ] Render context-aware help bar at the bottom of every screen (show only relevant keys for current screen/state)
  - [ ] Use `lipgloss` subtle styling so it doesn't dominate the view

- [ ] **29. Error handling and user feedback**
  - [ ] Wrong password: clear message with retry prompt
  - [ ] File not found (import): show path and suggestion
  - [ ] Invalid JSON (import): show parse error details
  - [ ] Key path not found (form): show warning but allow save
  - [ ] Network error (health check): show error message
  - [ ] Clipboard unavailable: show fallback message with the command text

- [ ] **30. Responsive layout**
  - [ ] Handle terminal resize (`tea.WindowSizeMsg`): adjust table columns, truncate long fields
  - [ ] Minimum terminal size check: if too small, show "Please resize your terminal" message

- [ ] **31. Cross-platform build and test**
  - [ ] Test on macOS (Terminal.app, iTerm2, Warp)
  - [ ] Test on Windows (Windows Terminal, PowerShell, cmd.exe)
  - [ ] Build script or Makefile: `GOOS=darwin GOARCH=amd64`, `GOOS=darwin GOARCH=arm64`, `GOOS=windows GOARCH=amd64`
  - [ ] Verify clipboard works on both platforms
  - [ ] Verify config path resolution on both platforms
  - [ ] Verify `exec ssh` works on both platforms (syscall.Exec on Unix, os/exec on Windows)

- [ ] **32. README and documentation**
  - [ ] Install instructions (download binary or `go install`)
  - [ ] Usage: TUI mode and CLI subcommands with examples
  - [ ] Screenshots / GIF of TUI in action
  - [ ] Security notes: encryption details, export warning, env var caveat
  - [ ] Build from source instructions

---

## 9. Dependencies (go.mod)

```
github.com/charmbracelet/bubbletea      # TUI framework
github.com/charmbracelet/bubbles        # TUI components (textinput, table, list, viewport, etc.)
github.com/charmbracelet/lipgloss       # TUI styling
github.com/sahilm/fuzzy                 # Fuzzy matching for search/filter
github.com/spf13/cobra                  # CLI subcommand framework
github.com/google/uuid                  # Generate connection IDs
golang.org/x/crypto                     # scrypt key derivation
```

---

## 10. Future enhancements (out of scope for v1)

- Cloud sync or remote config storage.
- SSH key generation (`ssh-keygen`) from within the app.
- Multiple config profiles / workspaces.
- SSH agent integration (add/remove keys from agent).
- Passphrase-protected key handling (unlock key before connecting).
- SFTP file browser integration.
- Connection sharing (encrypted export with recipient's public key).
