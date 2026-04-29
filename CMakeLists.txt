cmake_minimum_required(VERSION 3.10)
project(my_program C)

set(CMAKE_C_STANDARD 99)
set(CMAKE_C_STANDARD_REQUIRED ON)

if (MSVC)
    add_compile_options(/W4)
else()
    add_compile_options(-Wall)
endif()

set(CMAKE_RUNTIME_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR})

add_executable(my_program
    src/main.c
    src/app/app.c
    src/app/task_6_3/task_6_3.c
    src/app/task_6_4/task_6_4.c
    src/app/task_6_5/task_6_5.c
)

target_include_directories(my_program PRIVATE
    ${CMAKE_SOURCE_DIR}/src
)
