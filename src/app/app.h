#ifndef APP_H
#define APP_H

#include "task_6_3/task_6_3.h"
#include "task_6_4/task_6_4.h"
#include "task_6_5/task_6_5.h"

typedef struct
{
    task_6_3_t task_6_3;
    task_6_4_t task_6_4;
    task_6_5_t task_6_5;
} app_t;

int app_init(app_t *app);
int app_run(app_t *app);

#endif
