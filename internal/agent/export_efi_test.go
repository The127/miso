package agent

// EFI is where an image brings systemd-boot and the UKI stub, and the
// fallback loader of its ESP.
type EFI = efi

// EFIOf is the EFI of an architecture.
var EFIOf = efiOf
