# Go Test Project

## Windows 11: clone, install and run (PowerShell)

This is a **Go** project. You do not need Python, pip, venv, or npm. Install Git and Go (the repository currently declares Go 1.23.4 as its minimum in the root module; a current Go toolchain is recommended). Run the following commands **one at a time**:

```powershell
go version
```

```powershell
git clone https://github.com/Tyler-Keesee/GO-Project-Starter.git
```

```powershell
Set-Location .\GO-Project-Starter
```

```powershell
go mod download
```

```powershell
go run hello.go
```

The program runs in the terminal. Stop it with Ctrl+C if it keeps running.

**Verification:** `go version` and `go mod download` only verify tools/dependencies. A successful `go run` is the local smoke test. If you see compile errors, include the exact Go version, current folder, and first compiler error. Do not run `go mod tidy` blindly just to make the clone start, since it rewrites module files.

---


Test

Created by: Tyler
Version: 0.1.0

## Safety of setup scripts

`Setup.ps1` is deliberately **non-destructive**: it checks Go and downloads only normal module dependencies. It does not install Go, edit PATH/GOPATH, or reinitialize `go.mod`. The old script did all three. `RunProject.bat` now runs the checker and launches `hello.go` only after prerequisites pass. Install Go separately with `winget install --id GoLang.Go --exact`, and reopen PowerShell before running the checker.
