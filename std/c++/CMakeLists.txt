cmake_minimum_required(VERSION 3.16)

project(benc VERSION 2.0.0)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)

add_library(benc_cpp INTERFACE)
target_include_directories(benc_cpp INTERFACE
    $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}>
    $<INSTALL_INTERFACE:include/benc>)
add_executable(benc-test test.cpp)
target_link_libraries(benc-test PRIVATE benc_cpp)

include(CTest)
add_test(NAME cpp-runtime COMMAND benc-test)

include(GNUInstallDirs)
install(FILES std.hpp DESTINATION ${CMAKE_INSTALL_INCLUDEDIR}/benc)
install(TARGETS benc_cpp EXPORT bencCppTargets)
install(EXPORT bencCppTargets DESTINATION ${CMAKE_INSTALL_LIBDIR}/cmake/benc)
