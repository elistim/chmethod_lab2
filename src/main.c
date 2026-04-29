#include "app/app.h"

int main(void) 
{
    int status = 0;
    app_t app;
    if (app_init(&app) == 0) 
    {
        status =  app_run(&app);
    }
   
    return status;
}
