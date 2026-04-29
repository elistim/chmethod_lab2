# chmethod_lab2
Лабораторная работа №2 по Численным методам

# CMake Build Notes

The project uses CMake presets for configuration and build.

## Requirements

You need:

- `cmake`
- `ninja`
- one C compiler:
  - MSVC from Visual Studio Build Tools
  - or GCC from MinGW/MSYS2
  - or Clang/LLVM

## Presets

- `default`: Release build with `Ninja`
- `debug`: Debug build with `Ninja`
- `mingw`: Release build with `MinGW Makefiles`

## Build

If the compiler is already available in your shell:

```powershell
cmake --preset default
cmake --build --preset default
```

## Windows Examples

### MSVC

Run the commands from `x64 Native Tools Command Prompt for VS` or from a PowerShell session where the MSVC environment is loaded.

```powershell
cmake --preset default
cmake --build --preset default
```

### MSYS2 / MinGW

```powershell
$env:PATH = "C:\msys64\mingw64\bin;" + $env:PATH
cmake --preset mingw
cmake --build --preset mingw
```
