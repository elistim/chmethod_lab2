#ifndef CONSOLE_OUTPUT_H
#define CONSOLE_OUTPUT_H

#ifdef _WIN32

#include <windows.h>
#include <wchar.h>

static inline void console_print(const wchar_t *text)
{
    DWORD written = 0;

    WriteConsoleW(GetStdHandle(STD_OUTPUT_HANDLE),
                  text,
                  (DWORD)wcslen(text),
                  &written,
                  NULL);
}

#else

#include <stdio.h>
#include <wchar.h>

static inline void console_print(const wchar_t *text)
{
    wprintf(L"%ls", text);
}

#endif

#endif
