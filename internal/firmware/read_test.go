package firmware_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/deb/debtest"
	"github.com/The127/miso/internal/firmware"
)

func TestTheFirmwareOfAPackageIsItsCodeAndVarsWithoutSecureBoot(t *testing.T) {
	// arrange
	pkg := debtest.Package(t, map[string]string{
		"./usr/share/OVMF/OVMF_CODE_4M.fd":         "the code",
		"./usr/share/OVMF/OVMF_CODE_4M.secboot.fd": "the secure boot code",
		"./usr/share/OVMF/OVMF_VARS_4M.fd":         "the vars",
		"./usr/share/OVMF/OVMF_VARS_4M.ms.fd":      "the vars with microsoft keys",
		"./usr/share/ovmf/OVMF.fd":                 "the old single file",
	})

	// act
	found, err := firmware.Read(bytes.NewReader(pkg))

	// assert
	require.NoError(t, err)
	assert.Equal(t, firmware.Firmware{Code: []byte("the code"), Vars: []byte("the vars")}, found)
}

func TestAPackageWithoutTheVarsIsRefusedNamingThem(t *testing.T) {
	// arrange
	pkg := debtest.Package(t, map[string]string{
		"./usr/share/OVMF/OVMF_CODE_4M.fd":    "the code",
		"./usr/share/OVMF/OVMF_VARS_4M.ms.fd": "the vars with microsoft keys",
	})

	// act
	_, err := firmware.Read(bytes.NewReader(pkg))

	// assert
	assert.ErrorContains(t, err, "usr/share/OVMF/OVMF_VARS_4M.fd")
}

func TestAPackageWithoutTheCodeIsRefusedNamingIt(t *testing.T) {
	// arrange
	pkg := debtest.Package(t, map[string]string{
		"./usr/share/OVMF/OVMF_CODE_4M.secboot.fd": "the secure boot code",
		"./usr/share/OVMF/OVMF_VARS_4M.fd":         "the vars",
	})

	// act
	_, err := firmware.Read(bytes.NewReader(pkg))

	// assert
	assert.ErrorContains(t, err, "usr/share/OVMF/OVMF_CODE_4M.fd")
}
