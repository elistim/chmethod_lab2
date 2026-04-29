cmake_minimum_required(VERSION 3.10)
project(my_program C)

# Set C standard
set(CMAKE_C_STANDARD 99)
set(CMAKE_C_STANDARD_REQUIRED ON)

# Set compiler warnings for the active toolchain
if (MSVC)
    add_compile_options(/W4)
else()
    add_compile_options(-Wall)
endif()

# Set output directory for built executables
set(CMAKE_RUNTIME_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR})

# Include directories
include_directories(${CMAKE_SOURCE_DIR}/include)

# Source files
set(SOURCES
    src/main.c
)

# Create executable
add_executable(my_program ${SOURCES})
