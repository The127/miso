# Caching

miso keeps the result of the steps of a build and reuses it. A build that
changes only the end of an Imagefile runs only the steps from the change on.

## Cache keys

Every step has a cache key. `miso plan` lists the key of each step, and says
whether its layer is cached:

```
   2  3e624b25c9ad  cached  RUN apt-get update && apt-get install -y ...
   4  71d3dcea2f22  run     RUN : > /etc/fstab
```

`run` is a step that a build would run. After a change, the changed step and
every step after it say `run`.

The key of a step is made from:

- the key of the step before it
- the instruction as written
- what the instruction reads. For a stage's first step this is the digest of
  the base image. For `COPY` it is the content, the mode and the link target
  of every file it copies. For `COPY --from` it is the key of the stage it
  copies from.
- the version of miso

A step whose key is in the cache is not run again. When a step changes, its
key changes, and so do the keys of all the steps after it. The steps before it
keep their keys and their layers.

Some details:

- Comments, blank lines, line numbers, the case of the keyword and the names
  of stages are not part of a key.
- `RUN --network=none` and `RUN` have different keys.
- An upgrade of miso changes every key, so the first build with a new version
  builds everything again.

## Layers

`RUN` and `COPY` each produce a layer, which is the set of files the step
changed. `ENV` produces none. The layers are kept on the cache disk of the
builder VM.

`PARTITION`, `CMDLINE`, `OUTPUT` and `CHECK` do not change the root file
system. Changing them does not build the `RUN` and `COPY` steps before them
again.

## A RUN is cached by its text

miso does not know what a `RUN` downloads. The key depends on the text of the
command, not on the result. A step such as

```
RUN apt-get update && apt-get install -y htop
```

keeps its layer when the Debian archive has new packages. To build it again,
change the text of the step, or let `miso prune` remove the layer.

## Space

The layers are kept on the cache disk, and downloaded files, such as base
images, are kept in the cache directory. Both are described under
[the cache](cli.md#the-cache).

- `miso prune` removes what was not used for a while.
- `miso resize` makes the cache disk larger. It can grow up to 50 GiB until
  you resize it.

`miso shell` does not change the cache.
