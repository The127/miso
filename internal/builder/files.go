package builder

import "github.com/The127/miso/internal/protocol"

// Files are the files a copy carries from the build context, handed in
// because the builder reads no files of the host.
type Files func(protocol.Copy) protocol.Files
