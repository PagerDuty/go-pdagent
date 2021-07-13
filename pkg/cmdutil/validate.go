package cmdutil

func ValidateEnumField(inputVal string, allowedValues []string, err error) error {
	for _, value := range allowedValues {
		if value == inputVal {
			return nil
		}
	}
	return err
}
