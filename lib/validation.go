package lib


type Validation func() error

// Execute the validation chain and stop at the first error.
func Validate(validations ...Validation) error {
	for _, validation := range validations {
		err := validation()
		if err != nil {
			return err
		}
	}
	return nil
}
