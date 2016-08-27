#!/bin/sh
./cmdgui -extraflags="func,string,sin(x);x" "./plot" "{{$.func}} {{$.left}} {{$.right}} {{$.bottom}} {{$.top}}"

