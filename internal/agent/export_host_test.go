package agent

// EFI is where an image brings systemd-boot and the UKI stub, and the
// fallback loader of its ESP.
type EFI = efi

// EFIOf is the EFI of an architecture.
var EFIOf = efiOf

// KeepStream writes the compressed kernel of a bzImage of the image into a
// directory of its own.
var KeepStream = keepStream
