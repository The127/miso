package agent

// outputFile is the name of the file the tools write into the layer of a
// disk or a file system, and that Fetch reads back. The tools write it by
// shell, so the scripts and Fetch must agree on it.
const outputFile = "disk.raw"

// outputPath is where the tools see the file they write.
const outputPath = "/run/miso/out/" + outputFile
