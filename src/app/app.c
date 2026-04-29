#include "app.h"
#include "console_output.h"

#include <stdio.h>

int app_init(app_t *app)
{
    (void)app;
    return 0;
}

int app_run(app_t *app)
{
    int choice = 0;

    console_print(L"Выберите задание:\n");
    console_print(L"1 - 6.3 Приближение функции\n");
    console_print(L"2 - 6.4 Численное дифференцирование\n");
    console_print(L"3 - 6.5 Численное интегрирование\n");
    console_print(L"Ваш выбор: ");

    if (scanf("%d", &choice) != 1) {
        console_print(L"Ошибка ввода.\n");
        return 1;
    }

    switch (choice) {
    case 1:
        return task_6_3_main(&app->task_6_3);
    case 2:
        return task_6_4_main(&app->task_6_4);
    case 3:
        return task_6_5_main(&app->task_6_5);
    default:
        console_print(L"Нет такого пункта меню.\n");
        return 1;
    }
}
