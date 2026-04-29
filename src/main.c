#include "app/app.h"

#include <locale.h>

#ifdef _WIN32
#include <windows.h>
#endif

int main(void) {
#ifdef _WIN32
    SetConsoleCP(CP_UTF8);
    SetConsoleOutputCP(CP_UTF8);
#endif

    setlocale(LC_ALL, "");

    int status = 0;
    app_t app;
    if (app_init(&app) == 0) {
        status =  app_run(&app);
    }
   
    return status;
}
