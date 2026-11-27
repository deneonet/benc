cmake_minimum_required(VERSION 3.16)
project(benc_c VERSION 2.0.0 LANGUAGES C)

add_library(benc_c INTERFACE)
target_include_directories(benc_c INTERFACE
    $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}>
    $<INSTALL_INTERFACE:include/benc>)

set(CMAKE_C_STANDARD 17)
set(CMAKE_C_STANDARD_REQUIRED ON)
add_executable(benc-c-test test.c)
add_executable(benc-c-hardening hardening_test.c)
target_link_libraries(benc-c-test PRIVATE benc_c)
target_link_libraries(benc-c-hardening PRIVATE benc_c)

include(CTest)
add_test(NAME c-runtime COMMAND benc-c-test)
add_test(NAME c-hardening COMMAND benc-c-hardening)

include(GNUInstallDirs)
install(FILES benc.h DESTINATION ${CMAKE_INSTALL_INCLUDEDIR}/benc)
install(TARGETS benc_c EXPORT bencCTargets)
install(EXPORT bencCTargets DESTINATION ${CMAKE_INSTALL_LIBDIR}/cmake/benc)
