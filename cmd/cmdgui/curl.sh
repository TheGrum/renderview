#!/bin/sh

./cmdgui -defaultflags=false -extraflags="width,int,0,height,int,0,page,int,0,search,string,puppy" "curl" "http://loremflickr.com/{{$.width}}/{{$.height}}/{{$.search}}" 

