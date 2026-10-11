package agent

// EFI is where an image brings systemd-boot and the UKI stub, and the
// fallback loader of its ESP.
type EFI = efi

// EFIOf is the EFI of an architecture.
var EFIOf = efiOf

// MakeESP makes the ESP layer of an image of an architecture.
var MakeESP = makeESP

// CopyParts copies what a UKI of an image of an architecture is built from.
var CopyParts = copyParts
