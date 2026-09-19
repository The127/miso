package boot

// Boot readies a VM whose init is the calling program.
func Boot() error {
	if err := switchRoot(); err != nil {
		return err
	}

	if err := mountAll(); err != nil {
		return err
	}

	return loadModules()
}
