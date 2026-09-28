// Package check boots an image miso made and runs its CHECKs in it. Nothing
// of it goes into the image: systemd credentials hand the image a shell on
// vsock for the length of the boot only.
package check
