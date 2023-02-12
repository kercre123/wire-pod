package ble

// Connect initiates the sign-on process
func (v *VectorBLE) Connect(id int) error {
	v.ble.Reset()

	if err := v.ble.Connect(id); err != nil {
		return err
	}

	_, err := v.watch()
	return err
}
