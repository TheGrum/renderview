#  RenderView  #
================

A tool to quickly and easily wrap algorithms, image generation functions, web services, and the like with an interactive GUI interface offering panning, zooming, paging, and editable parameters.

This is *not* meant as a replacement for a general GUI toolkit. It is meant for programmers who lack the time to invest in learning yet another graphical toolkit and all of its quirks and foibles just to throw up a quick interactive interface to get a feel for the behavior of back-end code.

# Model-View-Controller #

Basically, this is a View+Controller combination that operates on a Model you write - except that for many tasks, the built-in models will suffice, and all you have to implement is an image generation function.

# RenderView #

The eponymous control, this comes in multiple flavors, and handles window creation, view rendering, and control/event-handling.

# Shiny #

Shiny is an experimental native cross-platform GUI package for Go. At the moment it is usable only as a framebuffer+event loop. If all you need is the output of your image generation function with panning and zooming, RenderView+Shiny supports that.

# go-gtk #

go-gtk is a functional CGo based GTK2 binding. RenderView on the go-gtk backend supports automatic parameter editing widget generation in addition to the interactive image.

# RenderModel #

You can implement this to customize what information gets collected from the GUI and passed to your visualization code, and what gets returned.

Basically, this boils down to a bag of parameters and a Render function that the view calls.

# BasicRenderModel #

In most cases, the BasicRenderModel will suffice. It provides a concrete implementation of the RenderModel interface and adds a throttling mechanism that ensures that requests for a new rendering from the view code only get passed through to your code when you do not already have a render in process, provided that you signal it by setting the Rendering bool appropriately at the start and end of your render.

It moves rendering to a separate Goroutine, preventing a long-running image generator from freezing the UI.

# RenderParameter #

Each RenderParameter carries a type string, allowing the renderView code to read and set the values without reflection. There is also a blank parameter that is automatically returned when a missing parameter is requested, allowing the RenderView code to behave as if the parameters it uses are always present. By either including or omitting the default parameters, you can control whether your code pays attention to certain controller behaviors.

Hints can be provided to indicate whether parameters are only for use in communicating with the View and Controller, or should be exposed to the user.

ZoomRenderParameter in the Mandelbrot example provides a demonstration of using a custom parameter to react immediately to changes in a value and, by setting other parameters in response, implement custom behavior.

# ChangeMonitor #

A means to observe a subset of RenderParameters and determine if they have changed since the value was last checked.

# cmdgui #

More than an example, this is a tool that applies the automatic GUI creation concept to command line applications or, through the medium of curl and other url-grabbing applications, to web services. It uses Go Templates to perform argument rewriting, and exports all parameters to the environment as well.

    #!/bin/sh
    ./cmdgui -extraflags="func,string,sin(x);x" "./plot" "{{$.func}} {{$.left}} {{$.right}} {{$.bottom}} {{$.top}}"

This one line example takes a python command line plot generator, and turns it into an interactive function plotter supporting changing the function, panning, zooming, and hand-entering of plot axis dimensions.
