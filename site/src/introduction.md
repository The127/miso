# Introduction

miso builds operating system images from a build file. The build file is
called an Imagefile and looks like a Dockerfile. It has `FROM`, `COPY` and
`RUN` steps, and every step becomes a cached layer.

The result comes out as standard systemd artifacts:

- disk images that boot under UEFI
- ISO images
- root file systems, with the kernel and initrd to boot them
- portable services, system extensions and configuration extensions
- the files `systemd-sysupdate` consumes

A `CHECK` step boots the result and runs a command in it. If the command
fails, the build fails and no output is written.

## What miso does not do

miso builds images. It does not apply updates, configure machines or run
anything on the target after installation. It writes standard systemd
formats and leaves the rest to whoever runs the machine.
